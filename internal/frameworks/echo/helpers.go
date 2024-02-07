package echo

import (
	"regexp"
	"strings"

	"github.com/diegoclair/goswag/internal/generator"
)

func getFuncName(name string) string {
	// echo has a method to get the function name, but it returns the full path of the function
	// we need to remove the package path and the "-fm" suffix
	fullFuncName := strings.TrimSuffix(name, "-fm")
	funcNameSplit := strings.Split(fullFuncName, ".")
	funcName := funcNameSplit[len(funcNameSplit)-1]

	return funcName
}

// toGoSwagRoute converts a slice of echoRoute to a slice of generator.Route.
// It iterates over each echoRoute in the input slice and appends its Route field to the output slice.
// Returns the converted slice of generator.Route.
func toGoSwagRoute(from []*echoRoute) []generator.Route {
	var routes []generator.Route
	for _, r := range from {
		routes = append(routes, r.Route)
	}

	return routes
}

// toGoSwagGroup converts a slice of echoGroup objects to a slice of generator.Group.
// It iterates over each echoGroup and creates a generator.Group object with the corresponding properties.
// The converted generator.Group objects are then returned as a slice.
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

// getPathParams extracts path parameters from the given path.
// It uses the :param format used by the Echo framework to define path parameters.
// The function returns a slice of strings containing all the extracted parameters.
func getPathParams(path string) []string {
	matches := pathParamRE.FindAllStringSubmatch(path, -1)
	var params []string
	for _, match := range matches {
		params = append(params, match[1])
	}

	return params
}
