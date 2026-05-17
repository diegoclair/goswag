package gin

import (
	"path"
	"reflect"
	"runtime"
	"strings"

	"github.com/diegoclair/goswag/internal/frameworks/shared"
	"github.com/diegoclair/goswag/internal/generator"
	"github.com/gin-gonic/gin"
)

// getFuncName resolves the last handler in the chain to a unique Go
// identifier. The last handler is the one that defines the route (earlier
// entries are middlewares). See shared.UniqueIdentifier for the rationale
// behind the disambiguation suffix.
func getFuncName(handlers ...gin.HandlerFunc) string {
	lastHandler := handlers[len(handlers)-1]
	fullFuncName := runtime.FuncForPC(reflect.ValueOf(lastHandler).Pointer()).Name()
	return shared.UniqueIdentifier(fullFuncName)
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

func getFullPath(groupName, relativePath string) string {
	if groupName == "" {
		return relativePath
	}

	fullPath := path.Join(groupName, relativePath)

	if strings.HasSuffix(relativePath, "/") {
		fullPath += "/"
	}

	return fullPath
}
