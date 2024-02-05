package goswag

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/ettle/strcase"
)

type ReturnType struct {
	StatusCode int
	Body       interface{}
	// example: map[jsonFieldName]fieldType
	OverrideStructFields map[string]interface{}
}

type param struct {
	name        string
	description string
	paramType   string
	required    bool
}

type route struct {
	path         string
	method       string
	funcName     string // it will be used to generate the function on the goswag.go file
	pathParams   []string
	summary      string
	description  string
	tags         []string
	accepts      []string
	produces     []string
	reads        interface{}
	returns      []ReturnType // example: map[statusCode]responseBody
	queryParams  []param
	headerParams []param
}

type group struct {
	groupName string
	routes    []route
	groups    []group
}

func generateSwagger(routes []route, groups []group) {
	var (
		packagesToImport []string
		fullFileContent  = &strings.Builder{}
	)

	log.Printf("Generating goswag.go file...")

	if routes != nil {
		packagesToImport = append(packagesToImport, writeRoutes("", routes, fullFileContent)...)
	}

	if groups != nil {
		packagesToImport = append(packagesToImport, writeGroup(groups, fullFileContent)...)
	}

	f, err := os.Create("./goswag.go")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintf(f, "package main\n\n")

	if len(packagesToImport) > 0 {
		fmt.Fprintf(f, "import (\n")

		for _, pkg := range packagesToImport {
			fmt.Fprintf(f, "\t_ \"%s\"\n", pkg)
		}

		fmt.Fprintf(f, ")\n\n")
	}

	fmt.Fprintf(f, "%s", fullFileContent.String())

	defer f.Close()

	log.Printf("goswag.go file generated successfully!")
}

func writeRoutes(groupName string, routes []route, s *strings.Builder) (packagesToImport []string) {
	for _, r := range routes {
		addLineIfNotEmpty(s, r.summary, "// @Summary %s\n")
		addTextIfNotEmptyOrDefault(s, r.summary, "// @Description %s\n", r.description)

		if len(r.tags) > 0 {
			s.WriteString(fmt.Sprintf("// @Tags %s\n", strings.Join(r.tags, ",")))
		} else if groupName != "" {
			s.WriteString(fmt.Sprintf("// @Tags %s\n", groupName))
		}

		addTextIfNotEmptyOrDefault(s, "json", "// @Accept %s\n", r.accepts...)
		addTextIfNotEmptyOrDefault(s, "json", "// @Produce %s\n", r.produces...)

		if r.reads != nil {
			s.WriteString(fmt.Sprintf("// @Param request body %s true \"Request\"\n", getStructAndPackageName(r.reads)))
		}

		for _, param := range r.pathParams {
			s.WriteString(fmt.Sprintf("// @Param %s path string true \"%s\" \n", param, strcase.ToCamel(param)))
		}

		for _, param := range r.queryParams {
			s.WriteString(fmt.Sprintf("// @Param %s query %s %t \"%s\"\n",
				param.name, param.paramType, param.required, param.description),
			)
		}

		for _, param := range r.headerParams {
			s.WriteString(fmt.Sprintf("// @Param %s header %s %t \"%s\"\n",
				param.name, param.paramType, param.required, param.description),
			)
		}

		if r.returns != nil {
			packagesToImport = append(packagesToImport, writeReturns(r.returns, s)...)
		}

		s.WriteString(fmt.Sprintf("// @Router %s [%s]\n", r.path, strings.ToLower(r.method)))

		s.WriteString(fmt.Sprintf("func %s() {}\n\n", r.funcName))
	}

	return packagesToImport
}

func writeReturns(returns []ReturnType, s *strings.Builder) (packagesToImport []string) {
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

			s.WriteString(fmt.Sprintf("// %s %d {object} %s\n", respType, data.StatusCode, correctlyResponseType))

			return append(packagesToImport, fullPathPackage)
		}

		s.WriteString(fmt.Sprintf("// %s %d {object} %s", respType, data.StatusCode, getStructAndPackageName(data.Body)))

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

	return nil
}

func writeGroup(groups []group, s *strings.Builder) (packagesToImport []string) {
	for _, g := range groups {
		res := writeRoutes(g.groupName, g.routes, s)
		if res != nil {
			packagesToImport = append(packagesToImport, res...)
		}

		if g.groups != nil {
			res := writeGroup(g.groups, s)
			if res != nil {
				packagesToImport = append(packagesToImport, res...)
			}
		}
	}

	return packagesToImport
}

func getStructAndPackageName(body interface{}) string {
	return reflect.TypeOf(body).String()
}

func addTextIfNotEmptyOrDefault(s *strings.Builder, defaultText, format string, text ...string) {
	if text != nil {
		if len(text) == 1 && strings.TrimSpace(text[0]) != "" {
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
