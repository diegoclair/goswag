package echo

import (
	"github.com/diegoclair/goswag/internal/frameworks/shared"
	"github.com/diegoclair/goswag/internal/generator"
)

// getFuncName returns a unique Go identifier for the handler whose fully
// qualified name is the input string. See shared.UniqueIdentifier for the
// rationale (collision avoidance across packages with same short name).
func getFuncName(name string) string {
	return shared.UniqueIdentifier(name)
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
