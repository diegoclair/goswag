package generator

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/diegoclair/goswag/v2/models"
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
	Summary      string
	Description  string
	Tags         []string
	Accepts      []string
	Produces     []string
	Reads        any
	Returns      []models.ReturnType // example: map[statusCode]responseBody
	QueryParams  []Param
	HeaderParams []Param
	PathParams   []Param
}

type Group struct {
	GroupName string
	Routes    []Route
	Groups    []Group
}

func GenerateSwagger(routes []Route, groups []Group, defaultResponses []models.ReturnType) {
	var (
		packagesToImport = make(map[string]bool)
		fullFileContent  = &strings.Builder{}
	)

	log.Printf("Generating %s file...", fileName)

	routes, groups = addDefaultResponses(routes, groups, defaultResponses)

	ambiguous := ambiguousTypeNames(routes, groups)

	if routes != nil {
		writeRoutes("", routes, fullFileContent, packagesToImport, ambiguous)
	}

	if groups != nil {
		writeGroup(groups, fullFileContent, packagesToImport, ambiguous)
	}

	f, err := os.Create(fmt.Sprintf("./%s", fileName))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	writeFileContent(f, fullFileContent.String(), packagesToImport)

	log.Printf("%s file generated successfully!", fileName)
}

// addDefaultResponses adds the default responses to the routes and groups if it are not empty
func addDefaultResponses(routes []Route, groups []Group, defaultResponses []models.ReturnType) ([]Route, []Group) {
	if len(defaultResponses) == 0 {
		return routes, groups
	}

	for i := range routes {
		routes[i].Returns = append(routes[i].Returns, defaultResponses...)
	}

	for i := range groups {
		groups[i].Routes, groups[i].Groups = addDefaultResponses(groups[i].Routes, groups[i].Groups, defaultResponses)
	}

	return routes, groups
}

func writeFileContent(file io.Writer, content string, packagesToImport map[string]bool) {
	fmt.Fprintf(file, "package main\n\n")

	if len(packagesToImport) > 0 {
		fmt.Fprintf(file, "import (\n")

		for _, pkg := range sortedKeys(packagesToImport) {
			fmt.Fprintf(file, "\t_ \"%s\"\n", pkg)
		}

		fmt.Fprintf(file, ")\n\n")
	}

	fmt.Fprintf(file, "%s", content)
}

func writeRoutes(groupName string, routes []Route, s *strings.Builder, packagesToImport, ambiguous map[string]bool) {
	for _, r := range routes {
		addLineIfNotEmpty(s, r.Summary, "// @Summary %s\n")
		addTextIfNotEmptyOrDefault(s, r.Summary, "// @Description %s\n", r.Description)

		if len(r.Tags) > 0 {
			fmt.Fprintf(s, "// @Tags %s\n", strings.Join(r.Tags, ","))
		} else if groupName != "" {
			fmt.Fprintf(s, "// @Tags %s\n", groupName)
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
			fmt.Fprintf(s, "// @Param request body %s true \"Request\"\n", annotationTypeName(r.Reads, ambiguous))
			addBodyPackageToImport(r.Reads, packagesToImport)
		}

		for _, param := range r.PathParams {
			fmt.Fprintf(s, "// @Param %s path %s %t \"%s\"\n",
				param.Name, param.ParamType, param.Required, param.Description)
		}

		for _, param := range r.QueryParams {
			fmt.Fprintf(s, "// @Param %s query %s %t \"%s\"\n",
				param.Name, param.ParamType, param.Required, param.Description)
		}

		for _, param := range r.HeaderParams {
			fmt.Fprintf(s, "// @Param %s header %s %t \"%s\"\n",
				param.Name, param.ParamType, param.Required, param.Description)
		}

		if r.Returns != nil {
			writeReturns(r.Returns, s, packagesToImport, ambiguous)
		}

		if r.Path != "" {
			fmt.Fprintf(s, "// @Router %s [%s]\n", r.Path, strings.ToLower(r.Method))
		}

		if r.FuncName != "" {
			fmt.Fprintf(s, "func %s() {} //nolint:unused \n", r.FuncName)
		}

		s.WriteString("\n")
	}
}

func writeReturns(returns []models.ReturnType, s *strings.Builder, packagesToImport, ambiguous map[string]bool) {
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
			fmt.Fprintf(s, "// %s %d\n", respType, data.StatusCode)
			continue
		}

		fmt.Fprintf(s, "// %s %d {object} %s", respType, data.StatusCode, annotationTypeName(data.Body, ambiguous))

		addPackageToImport(data, packagesToImport)
		handleOverrideStructFields(s, data, ambiguous)

		s.WriteString("\n")
	}
}

func writeGroup(groups []Group, s *strings.Builder, packagesToImport, ambiguous map[string]bool) {
	for _, g := range groups {
		writeRoutes(g.GroupName, g.Routes, s, packagesToImport, ambiguous)

		if g.Groups != nil {
			writeGroup(g.Groups, s, packagesToImport, ambiguous)
		}
	}
}

// genericArgPkgRe matches a fully-qualified package path followed by a type name, e.g.
// "github.com/foo/bar.Type". reflect renders generic type arguments this way (full path)
// while the outer type keeps its short package name, which is how we tell them apart.
var genericArgPkgRe = regexp.MustCompile(`([\w.\-]+(?:/[\w.\-]+)+)\.(\w+)`)

// addBodyPackageToImport adds every package the body's type depends on to the import map:
// the type itself, its slice/array element, and any package named inside generic type
// arguments (reflect exposes no API to walk those, so we parse them out of t.String()).
func addBodyPackageToImport(body any, packagesToImport map[string]bool) {
	if body == nil {
		return
	}

	t := reflect.TypeOf(body)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	collectPackages(t, packagesToImport)

	for _, match := range genericArgPkgRe.FindAllStringSubmatch(t.String(), -1) {
		packagesToImport[match[1]] = true
	}
}

// collectPackages walks the containers a body can be wrapped in — []T, [n]T, *T,
// map[K]V — until it reaches named types, which are the ones needing an import.
// It stops at named types instead of descending into their fields: those resolve
// through their own package, and a self-referencing struct would not terminate.
func collectPackages(t reflect.Type, packagesToImport map[string]bool) {
	if t.PkgPath() != "" {
		packagesToImport[t.PkgPath()] = true
		return
	}

	switch t.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Array:
		collectPackages(t.Elem(), packagesToImport)
	case reflect.Map:
		collectPackages(t.Key(), packagesToImport)
		collectPackages(t.Elem(), packagesToImport)
	}
}

// addPackageToImport adds the package to import.
func addPackageToImport(data models.ReturnType, packagesToImport map[string]bool) {
	addBodyPackageToImport(data.Body, packagesToImport)
}

// ambiguousTypeNames returns the short names ("pkg.Type") that more than one package
// declares among the referenced bodies. swag resolves a short name by scanning the
// imports of the generated file, so those names would otherwise be answered by
// whichever package it happens to look at first.
func ambiguousTypeNames(routes []Route, groups []Group) map[string]bool {
	// One package per name is enough to decide: the second distinct one already
	// answers the question, and keeping the whole set would allocate a map per
	// type name on a route table that can be very large.
	firstPkg := map[string]string{}
	ambiguous := map[string]bool{}

	addBody := func(body any) {
		for shortName, pkgPath := range typeNamesOf(body) {
			switch seen, ok := firstPkg[shortName]; {
			case !ok:
				firstPkg[shortName] = pkgPath
			case seen != pkgPath:
				ambiguous[shortName] = true
			}
		}
	}

	var walk func(routes []Route, groups []Group)
	walk = func(routes []Route, groups []Group) {
		for _, r := range routes {
			addBody(r.Reads)

			for _, ret := range r.Returns {
				addBody(ret.Body)

				for _, override := range ret.OverrideStructFields {
					addBody(override)
				}
			}
		}

		for _, g := range groups {
			walk(g.Routes, g.Groups)
		}
	}
	walk(routes, groups)

	return ambiguous
}

// typeNamesOf maps every short name the annotation of body will contain — the body's
// own type and each of its generic arguments — to the package that declares it.
func typeNamesOf(body any) map[string]string {
	names := map[string]string{}
	if body == nil {
		return names
	}

	_, t := unwrapType(reflect.TypeOf(body))
	if t.PkgPath() == "" {
		return names
	}

	ident, args := splitGenericArgs(t.Name())
	names[pkgNameOf(t)+"."+ident] = t.PkgPath()

	for _, match := range genericArgPkgRe.FindAllStringSubmatch(args, -1) {
		names[path.Base(match[1])+"."+match[2]] = match[1]
	}

	return names
}

// annotationTypeName renders the type name swag reads from the annotation.
func annotationTypeName(body any, ambiguous map[string]bool) string {
	prefix, t := unwrapType(reflect.TypeOf(body))
	if t.PkgPath() == "" {
		return prefix + t.String()
	}

	ident, args := splitGenericArgs(t.Name())
	name := prefix + qualifyTypeName(pkgNameOf(t), ident, t.PkgPath(), ambiguous)

	if args == "" {
		return name
	}

	// reflect writes generic arguments with the full package path; swag expects the
	// same short form as the outer type.
	return name + "[" + genericArgPkgRe.ReplaceAllStringFunc(args, func(arg string) string {
		match := genericArgPkgRe.FindStringSubmatch(arg)
		return qualifyTypeName(path.Base(match[1]), match[2], match[1], ambiguous)
	})
}

// qualifyTypeName keeps the short name unless it is ambiguous, in which case it uses
// the name swag gives to a definition it found in more than one package: the package
// path with its separators replaced by underscores.
func qualifyTypeName(pkgName, ident, pkgPath string, ambiguous map[string]bool) string {
	shortName := pkgName + "." + ident
	if !ambiguous[shortName] {
		return shortName
	}

	return strings.NewReplacer("/", "_", ".", "_", `\`, "_").Replace(pkgPath) + "." + ident
}

// unwrapType strips the containers swag renders as a prefix, returning that prefix
// and the named type underneath it.
func unwrapType(t reflect.Type) (prefix string, elem reflect.Type) {
	for {
		switch t.Kind() {
		case reflect.Pointer:
			t = t.Elem()
		case reflect.Slice:
			prefix += "[]"
			t = t.Elem()
		default:
			return prefix, t
		}
	}
}

// splitGenericArgs splits "Generic[pkg/path.Arg]" into "Generic" and "pkg/path.Arg]",
// the trailing bracket included so the caller can write it back unchanged.
func splitGenericArgs(typeName string) (ident, args string) {
	ident, args, _ = strings.Cut(typeName, "[")
	return ident, args
}

func pkgNameOf(t reflect.Type) string {
	return strings.TrimSuffix(t.String(), "."+t.Name())
}

func handleOverrideStructFields(s *strings.Builder, data models.ReturnType, ambiguous map[string]bool) {
	if data.OverrideStructFields != nil {
		for i, key := range sortedKeys(data.OverrideStructFields) {
			if i == 0 {
				s.WriteString("{")
			}

			fmt.Fprintf(s, "%s=%s", key, annotationTypeName(data.OverrideStructFields[key], ambiguous))
			if i == len(data.OverrideStructFields)-1 {
				s.WriteString("}")
			} else {
				s.WriteString(",")
			}
		}
	}
}

// sortedKeys keeps the generated file stable across runs: map iteration order would
// otherwise reshuffle it on every generation.
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}

func addTextIfNotEmptyOrDefault(s *strings.Builder, defaultText, format string, text ...string) {
	if text != nil {
		if len(text) >= 1 && strings.TrimSpace(text[0]) != "" {
			fmt.Fprintf(s, format, strings.Join(text, ","))
			return
		}
	}

	if defaultText != "" {
		fmt.Fprintf(s, format, defaultText)
	}
}

func addLineIfNotEmpty(s *strings.Builder, data, format string) {
	if data != "" {
		fmt.Fprintf(s, format, data)
	}
}
