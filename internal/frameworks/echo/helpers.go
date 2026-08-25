package echo

import (
	"github.com/diegoclair/goswag/v2/internal/frameworks/shared"
	"github.com/diegoclair/goswag/v2/internal/generator"
	"github.com/labstack/echo/v5"
)

// echo v5 defaults RouteInfo.Name to "method:path", so the generated stub name
// must come from the handler itself.
func newRoute(ri echo.RouteInfo, h echo.HandlerFunc) *echoRoute {
	return &echoRoute{
		Path:     ri.Path,
		Method:   ri.Method,
		FuncName: getFuncName(echo.HandlerName(h)),
	}
}

// getFuncName turns a fully qualified handler name into a unique Go identifier.
func getFuncName(name string) string {
	return shared.UniqueIdentifier(name)
}

// toGoSwagRoute converts a slice of echoRoute to a slice of generator.Route.
func toGoSwagRoute(from []*echoRoute) []generator.Route {
	var routes []generator.Route
	for _, r := range from {
		routes = append(routes, r.Route)
	}

	return routes
}

// toGoSwagGroup converts a slice of echoGroup to a slice of generator.Group,
// preserving the nesting.
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
