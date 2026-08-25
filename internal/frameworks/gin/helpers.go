package gin

import (
	"path"
	"reflect"
	"runtime"
	"strings"

	"github.com/diegoclair/goswag/v2/internal/frameworks/shared"
	"github.com/diegoclair/goswag/v2/internal/generator"
	"github.com/gin-gonic/gin"
)

// getFuncName resolves the route handler to a unique Go identifier; the last
// entry in the chain is the handler, the earlier ones are middlewares.
func getFuncName(handlers ...gin.HandlerFunc) string {
	lastHandler := handlers[len(handlers)-1]
	fullFuncName := runtime.FuncForPC(reflect.ValueOf(lastHandler).Pointer()).Name()
	return shared.UniqueIdentifier(fullFuncName)
}

// toGoSwagRoute converts a slice of ginRoute to a slice of generator.Route.
func toGoSwagRoute(from []*ginRoute) []generator.Route {
	var routes []generator.Route
	for _, r := range from {
		routes = append(routes, r.Route)
	}

	return routes
}

// toGoSwagGroup converts a slice of ginGroup to a slice of generator.Group.
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
