package gin

import (
	"reflect"
	"regexp"
	"runtime"
	"strings"

	"github.com/diegoclair/goswag/internal/generator"
	"github.com/gin-gonic/gin"
)

// getFuncName retrieves the name of the function associated with the last handler in the given list of gin.HandlerFunc.
// It uses the reflect package to obtain the function name from the pointer value of the last handler.
// The function name is extracted by splitting the full function name string using the dot separator and returning the last element.
// The retrieved function name is then returned as a string.
func getFuncName(handlers ...gin.HandlerFunc) string {
	lastHandler := handlers[len(handlers)-1]

	fullFuncName := runtime.FuncForPC(reflect.ValueOf(lastHandler).Pointer()).Name()
	funcNameSplit := strings.Split(fullFuncName, ".")
	funcName := funcNameSplit[len(funcNameSplit)-1]
	funcName = strings.TrimSuffix(funcName, "-fm")

	return funcName
}

var pathParamRE = regexp.MustCompile(`:(.*?)(/|$)`)

// getPathParams extracts path parameters from the given path.
// It uses the :param format used by the Gin framework to define path parameters.
// The function returns a slice of strings containing all the extracted parameters.
func getPathParams(path string) []string {
	matches := pathParamRE.FindAllStringSubmatch(path, -1)
	var params []string
	for _, match := range matches {
		params = append(params, match[1])
	}

	return params
}

// toGoSwagRoute converts a slice of ginRoute to a slice of generator.Route.
// It iterates over each ginRoute in the input slice and appends its Route field to the output slice.
// Returns the converted slice of generator.Route.
func toGoSwagRoute(from []*ginRoute) []generator.Route {
	var routes []generator.Route
	for _, r := range from {
		routes = append(routes, r.Route)
	}

	return routes
}

// toGoSwagGroup converts a slice of ginGroup objects to a slice of generator.Group.
// It iterates over each ginGroup and creates a generator.Group object with the corresponding properties.
// The converted generator.Group objects are then returned as a slice.
func toGoSwagGroup(from []*ginGroup) []generator.Group {
	var groups []generator.Group
	for _, g := range from {
		groups = append(groups, generator.Group{
			GroupName: g.groupName,
			Routes:    toGoSwagRoute(g.routes),
		})
	}

	return groups
}
