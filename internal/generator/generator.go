package generator

import (
	"fmt"
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

	fmt.Fprintf(f, "package main\n\n")

	if len(packagesToImport) > 0 {
		fmt.Fprintf(f, "import (\n")

		for _, pkg := range packagesToImport {
			fmt.Fprintf(f, "\t_ \"%s\"\n", pkg)
		}

		fmt.Fprintf(f, ")\n\n")
	}

	fmt.Fprintf(f, "%s", fullFileContent.String())

	log.Printf("%s file generated successfully!", fileName)
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

		addTextIfNotEmptyOrDefault(s, "json", "// @Accept %s\n", r.Accepts...)
		addTextIfNotEmptyOrDefault(s, "json", "// @Produce %s\n", r.Produces...)

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

		s.WriteString(fmt.Sprintf("// @Router %s [%s]\n", r.Path, strings.ToLower(r.Method)))

		s.WriteString(fmt.Sprintf("func %s() {}\n\n", r.FuncName))
	}

	return packagesToImport
}

func writeReturns(returns []models.ReturnType, s *strings.Builder) (packagesToImport []string) {
	for _, data := range returns {
		respType := "@Success"
		firstDigit := data.StatusCode / 100

		if firstDigit != http.StatusOK/100 { // <> 2xx
			respType = "@Failure"
		}

		if data.Body == nil {
			s.WriteString(fmt.Sprintf("// %s %d\n", respType, data.StatusCode))
			continue
		}

		bodyName := getStructAndPackageName(data.Body)
		isGeneric := bodyName[len(bodyName)-1:] == "]"

		if isGeneric {
			split := strings.Split(bodyName, "]")
			insideGenericsFullName := split[len(split)-2]
			lastSlashIndex := strings.LastIndex(insideGenericsFullName, "/")
			beforePkg := insideGenericsFullName[:lastSlashIndex]

			correctlyResponseType := strings.Replace(bodyName, beforePkg+"/", "", -1) // remove full package from the struct name

			pkg := strings.Split(correctlyResponseType, ".")[0]
			fullPathPackage := beforePkg + "/" + pkg

			s.WriteString(fmt.Sprintf("// %s %d {object} %s", respType, data.StatusCode, correctlyResponseType))

			packagesToImport = append(packagesToImport, fullPathPackage)
		} else {
			s.WriteString(fmt.Sprintf("// %s %d {object} %s", respType, data.StatusCode, getStructAndPackageName(data.Body)))
		}

		if data.OverrideStructFields != nil {
			i := 0
			for key, object := range data.OverrideStructFields {
				if i == 0 {
					s.WriteString("{")
				}

				s.WriteString(fmt.Sprintf("%s=%s", key, getStructAndPackageName(object)))
				if i == len(data.OverrideStructFields)-1 {
					s.WriteString("}\n")
				} else {
					s.WriteString(",")
				}
				i++
			}
		} else {
			s.WriteString("\n")
		}

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

func getStructAndPackageName(body any) string {
	return reflect.TypeOf(body).String()
}

func addTextIfNotEmptyOrDefault(s *strings.Builder, defaultText, format string, text ...string) {
	if text != nil {
		if len(text) >= 1 && strings.TrimSpace(text[0]) != "" {
			s.WriteString(fmt.Sprintf(format, strings.Join(text, ",")))
			return
		}
	}

	s.WriteString(fmt.Sprintf(format, defaultText))
}

func addLineIfNotEmpty(s *strings.Builder, data, format string) {
	if data != "" {
		s.WriteString(fmt.Sprintf(format, data))
	}
}
