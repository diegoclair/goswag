package echo

import (
	"regexp"
	"strings"

	"github.com/diegoclair/goswag/internal/generator"
)

func getFuncName(name string) string {
	fullFuncName := strings.TrimSuffix(name, "-fm")
	funcNameSplit := strings.Split(fullFuncName, ".")
	funcName := funcNameSplit[len(funcNameSplit)-1]

	return funcName
}

func toGoSwagRoute(from []*echoRoute) []generator.Route {
	var routes []generator.Route
	for _, r := range from {
		routes = append(routes, r.Route)
	}

	return routes
}

func toGoSwagGroup(from []*echoGroup) []generator.Group {
	var groups []generator.Group
	for _, g := range from {
		groups = append(groups, generator.Group{
			GroupName: g.groupName,
			Routes:    toGoSwagRoute(g.routes),
			Groups:    toGoSwagGroup(g.groups)},
		)
	}

	return groups
}

var pathParamRE = regexp.MustCompile(`:(.*?)(/|$)`)

func getPathParams(path string) []string {
	matches := pathParamRE.FindAllStringSubmatch(path, -1)
	var params []string
	for _, match := range matches {
		params = append(params, match[1])
	}

	return params
}
