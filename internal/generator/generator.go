package generator

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/diegoclair/goswag/models"
	"github.com/ettle/strcase"
)

const fileName = "goswag.go"

type Param struct {
	Name        string
	Description string
	ParamType   string
	Required    bool
}

type Route struct {
	Path         string
	Method       string
	FuncName     string // it will be used to generate the function on the goswag.go file
	PathParams   []string
	Summary      string
	Description  string
	Tags         []string
	Accepts      []string
	Produces     []string
	Reads        interface{}
	Returns      []models.ReturnType // example: map[statusCode]responseBody
	QueryParams  []Param
	HeaderParams []Param
}

type Group struct {
	GroupName string
	Routes    []Route
	Groups    []Group
}

func GenerateSwagger(routes []Route, groups []Group) {
	var (
		packagesToImport []string
		fullFileContent  = &strings.Builder{}
	)

	log.Printf("Generating %s file...", fileName)

	if routes != nil {
		packagesToImport = append(packagesToImport, writeRoutes("", routes, fullFileContent)...)
	}

	if groups != nil {
		packagesToImport = append(packagesToImport, writeGroup(groups, fullFileContent)...)
	}

	f, err := os.Create(fmt.Sprintf("./%s", fileName))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	writeFileContent(f, fullFileContent.String(), packagesToImport)

	log.Printf("%s file generated successfully!", fileName)
}

func writeFileContent(file io.Writer, content string, packagesToImport []string) {
	fmt.Fprintf(file, "package main\n\n")

	if len(packagesToImport) > 0 {
		fmt.Fprintf(file, "import (\n")

		for _, pkg := range packagesToImport {
			fmt.Fprintf(file, "\t_ \"%s\"\n", pkg)
		}

		fmt.Fprintf(file, ")\n\n")
	}

	fmt.Fprintf(file, "%s", content)
}

func writeRoutes(groupName string, routes []Route, s *strings.Builder) (packagesToImport []string) {
	for _, r := range routes {
		addLineIfNotEmpty(s, r.Summary, "// @Summary %s\n")
		addTextIfNotEmptyOrDefault(s, r.Summary, "// @Description %s\n", r.Description)

		if len(r.Tags) > 0 {
			s.WriteString(fmt.Sprintf("// @Tags %s\n", strings.Join(r.Tags, ",")))
		} else if groupName != "" {
			s.WriteString(fmt.Sprintf("// @Tags %s\n", groupName))
		}

		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			// methods like get or delete do not have a request body
			addTextIfNotEmptyOrDefault(s, "json", "// @Accept %s\n", r.Accepts...)
		}

		if r.Returns != nil {
			// only add the produces if there is a return
			addTextIfNotEmptyOrDefault(s, "json", "// @Produce %s\n", r.Produces...)
		}

		if r.Reads != nil {
			s.WriteString(fmt.Sprintf("// @Param request body %s true \"Request\"\n", getStructAndPackageName(r.Reads)))
		}

		for _, param := range r.PathParams {
			s.WriteString(fmt.Sprintf("// @Param %s path string true \"%s\" \n", param, strcase.ToCamel(param)))
		}

		for _, param := range r.QueryParams {
			s.WriteString(fmt.Sprintf("// @Param %s query %s %t \"%s\"\n",
				param.Name, param.ParamType, param.Required, param.Description),
			)
		}

		for _, param := range r.HeaderParams {
			s.WriteString(fmt.Sprintf("// @Param %s header %s %t \"%s\"\n",
				param.Name, param.ParamType, param.Required, param.Description),
			)
		}

		if r.Returns != nil {
			packagesToImport = append(packagesToImport, writeReturns(r.Returns, s)...)
		}

		if r.Path != "" {
			s.WriteString(fmt.Sprintf("// @Router %s [%s]\n", r.Path, strings.ToLower(r.Method)))
		}

		if r.FuncName != "" {
			s.WriteString(fmt.Sprintf("func %s() {} //nolint:unused \n", r.FuncName))
		}

		s.WriteString("\n")
	}

	return packagesToImport
}

func writeReturns(returns []models.ReturnType, s *strings.Builder) (packagesToImport []string) {
	for _, data := range returns {
		if data.StatusCode == 0 {
			continue
		}

		respType := "@Success"
		firstDigit := data.StatusCode / 100

		if firstDigit != http.StatusOK/100 { // <> 2xx
			respType = "@Failure"
		}

		if data.Body == nil {
			s.WriteString(fmt.Sprintf("// %s %d\n", respType, data.StatusCode))
			continue
		}

		var isGeneric bool
		isGeneric, packagesToImport = writeIfIsGenericType(s, data, respType)

		if !isGeneric {
			// if it is not a generic type, we can write the response normally
			s.WriteString(fmt.Sprintf("// %s %d {object} %s", respType, data.StatusCode, getStructAndPackageName(data.Body)))
		}

		handleOverrideStructFields(s, data)

		s.WriteString("\n")

	}

	return packagesToImport
}

func writeGroup(groups []Group, s *strings.Builder) (packagesToImport []string) {
	for _, g := range groups {
		packagesToImport = append(packagesToImport, writeRoutes(g.GroupName, g.Routes, s)...)

		if g.Groups != nil {
			packagesToImport = append(packagesToImport, writeGroup(g.Groups, s)...)
		}
	}

	return packagesToImport
}

// writeIfIsGenericType writes the correctly response type if it is a generic type
// and returns the packages to import that need to be added to the goswag.go file to make it work
func writeIfIsGenericType(s *strings.Builder, data models.ReturnType, respType string) (isGeneric bool, packagesToImport []string) {
	bodyName := getStructAndPackageName(data.Body)

	// generic last character here will be ']'
	// testutil.StructGeneric[testutil.TestGeneric]
	isGeneric = bodyName[len(bodyName)-1:] == "]"
	if !isGeneric {
		return
	}

	isArray := strings.Contains(bodyName, "[[]")
	hasSlash := strings.Contains(bodyName, "/")

	if isArray && hasSlash {
		// example: testutil.StructGeneric[[]github.com/diegoclair/goswag/internal/generator/testutil.TestGeneric]

		bodyRemovedLastChar := bodyName[:len(bodyName)-1] // testutil.StructGeneric[[]github.com/diegoclair/goswag/internal/generator/testutil.TestGeneric

		// get the last text after '/'
		str := strings.Split(bodyRemovedLastChar, "/")
		insideGenericsFullName := str[len(str)-1] // testutil.TestGeneric

		pkg := strings.Split(insideGenericsFullName, ".")[0] // testutil

		insidePkg := strings.Split(bodyRemovedLastChar, "[[]")[1]                 // github.com/diegoclair/goswag/internal/generator/testutil.TestGeneric
		removedType := strings.Replace(insidePkg, insideGenericsFullName, "", -1) // github.com/diegoclair/goswag/internal/generator/
		fullInsidePkg := removedType + pkg                                        // github.com/diegoclair/goswag/internal/generator/testutil

		packagesToImport = append(packagesToImport, fullInsidePkg)

		correctlyResponseType := strings.Replace(bodyName, removedType, "", -1) // remove full package from the struct name

		s.WriteString(fmt.Sprintf("// %s %d {object} %s", respType, data.StatusCode, correctlyResponseType))

		return isGeneric, packagesToImport
	}

	if hasSlash {
		// example: testutil.StructGeneric[github.com/diegoclair/goswag/internal/generator/testutil.TestGeneric]

		bodyRemovedLastChar := bodyName[:len(bodyName)-1] // testutil.StructGeneric[github.com/diegoclair/goswag/internal/generator/testutil.TestGeneric

		// get the last text after '/'
		str := strings.Split(bodyRemovedLastChar, "/")
		insideGenericsFullName := str[len(str)-1] // testutil.TestGeneric

		pkg := strings.Split(insideGenericsFullName, ".")[0] // testutil

		insidePkg := strings.Split(bodyRemovedLastChar, "[")[1]                   // github.com/diegoclair/goswag/internal/generator/testutil.TestGeneric
		removedType := strings.Replace(insidePkg, insideGenericsFullName, "", -1) // github.com/diegoclair/goswag/internal/generator/
		fullInsidePkg := removedType + pkg                                        // github.com/diegoclair/goswag/internal/generator/testutil

		packagesToImport = append(packagesToImport, fullInsidePkg)

		correctlyResponseType := strings.Replace(bodyName, removedType, "", -1) // remove full package from the struct name

		s.WriteString(fmt.Sprintf("// %s %d {object} %s", respType, data.StatusCode, correctlyResponseType))

		return isGeneric, packagesToImport
	}

	// example: genericStruct[int] or genericStruct[string] or genericStruct[bool]
	// primitive types do not need to import packages

	s.WriteString(fmt.Sprintf("// %s %d {object} %s", respType, data.StatusCode, bodyName))

	return isGeneric, packagesToImport
}

func handleOverrideStructFields(s *strings.Builder, data models.ReturnType) {
	if data.OverrideStructFields != nil {
		i := 0
		for key, object := range data.OverrideStructFields {
			if i == 0 {
				s.WriteString("{")
			}

			s.WriteString(fmt.Sprintf("%s=%s", key, getStructAndPackageName(object)))
			if i == len(data.OverrideStructFields)-1 {
				s.WriteString("}")
			} else {
				s.WriteString(",")
			}
			i++
		}
	}
}

func getStructAndPackageName(body any) string {
	isPointer := reflect.TypeOf(body).Kind() == reflect.Ptr
	if isPointer {
		body = reflect.ValueOf(body).Elem().Interface()
	}

	return reflect.TypeOf(body).String()
}

func addTextIfNotEmptyOrDefault(s *strings.Builder, defaultText, format string, text ...string) {
	if text != nil {
		if len(text) >= 1 && strings.TrimSpace(text[0]) != "" {
			s.WriteString(fmt.Sprintf(format, strings.Join(text, ",")))
			return
		}
	}

	if defaultText != "" {
		s.WriteString(fmt.Sprintf(format, defaultText))
	}
}

func addLineIfNotEmpty(s *strings.Builder, data, format string) {
	if data != "" {
		s.WriteString(fmt.Sprintf(format, data))
	}
}
