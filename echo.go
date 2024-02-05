package goswag

import (
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"
)

type Router interface {
	// GET registers a new GET route for a path with matching handler in the router
	// with optional route-level middleware.
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) Swagger

	// POST registers a new POST route for a path with matching handler in the
	// router with optional route-level middleware.
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) Swagger

	// PUT registers a new PUT route for a path with matching handler in the
	// router with optional route-level middleware.
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) Swagger

	// DELETE registers a new DELETE route for a path with matching handler in the router
	// with optional route-level middleware.
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) Swagger

	// PATCH registers a new PATCH route for a path with matching handler in the
	// router with optional route-level middleware.
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) Swagger

	// OPTIONS registers a new OPTIONS route for a path with matching handler in the
	// router with optional route-level middleware.
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) Swagger
}

type Group interface {
	Router
	// Group automatically create tags for the swagger documentation.
	//
	// Group creates a new router group with prefix and optional group-level middleware.
	Group(prefix string, m ...echo.MiddlewareFunc) Group
}

type Echo interface {
	Group
	GenerateSwagger()
	Echo() *echo.Echo
}

type echoSwagger struct {
	e      *echo.Echo
	groups []*echoGroup
	routes []*echoRoute
}

func newSwaggerEcho() Echo {
	return &echoSwagger{
		e: echo.New(),
	}
}

func (s *echoSwagger) Echo() *echo.Echo {
	return s.e
}

func (s *echoSwagger) GenerateSwagger() {
	generateSwagger(toGoSwagRoute(s.routes), toGoSwagGroup(s.groups))
}

func getFuncName(name string) string {
	fullFuncName := strings.TrimSuffix(name, "-fm")
	funcNameSplit := strings.Split(fullFuncName, ".")
	funcName := funcNameSplit[len(funcNameSplit)-1]

	return funcName
}

func toGoSwagRoute(from []*echoRoute) []route {
	var routes []route
	for _, r := range from {
		routes = append(routes, r.route)
	}

	return routes
}

func toGoSwagGroup(from []*echoGroup) []group {
	var groups []group
	for _, g := range from {
		groups = append(groups, group{
			groupName: g.groupName,
			routes:    toGoSwagRoute(g.routes),
			groups:    toGoSwagGroup(g.groups)},
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

func (s *echoSwagger) Group(prefix string, m ...echo.MiddlewareFunc) Group {
	g := &echoGroup{g: s.e.Group(prefix, m...), groupName: prefix}
	s.groups = append(s.groups, g)

	return g
}

func (s *echoSwagger) POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) Swagger {
	ec := s.e.POST(path, h, m...)

	er := &echoRoute{
		EchoRoute: ec,
		route: route{
			path:       ec.Path,
			method:     ec.Method,
			funcName:   getFuncName(ec.Name),
			pathParams: getPathParams(ec.Path),
		},
	}

	s.routes = append(s.routes, er)

	return er
}

func (s *echoSwagger) GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) Swagger {
	ec := s.e.GET(path, h, m...)

	er := &echoRoute{
		EchoRoute: ec,
		route: route{
			path:       ec.Path,
			method:     ec.Method,
			funcName:   getFuncName(ec.Name),
			pathParams: getPathParams(ec.Path),
		},
	}

	s.routes = append(s.routes, er)

	return er
}

func (s *echoSwagger) PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) Swagger {
	ec := s.e.PUT(path, h, m...)

	er := &echoRoute{
		EchoRoute: ec,
		route: route{
			path:       ec.Path,
			method:     ec.Method,
			funcName:   getFuncName(ec.Name),
			pathParams: getPathParams(ec.Path),
		},
	}

	s.routes = append(s.routes, er)

	return er
}

func (s *echoSwagger) DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) Swagger {
	ec := s.e.DELETE(path, h, m...)

	er := &echoRoute{
		EchoRoute: ec,
		route: route{
			path:       ec.Path,
			method:     ec.Method,
			funcName:   getFuncName(ec.Name),
			pathParams: getPathParams(ec.Path),
		},
	}

	s.routes = append(s.routes, er)

	return er
}

func (s *echoSwagger) PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) Swagger {
	ec := s.e.PATCH(path, h, m...)

	er := &echoRoute{
		EchoRoute: ec,
		route: route{
			path:       ec.Path,
			method:     ec.Method,
			funcName:   getFuncName(ec.Name),
			pathParams: getPathParams(ec.Path),
		},
	}

	s.routes = append(s.routes, er)

	return er
}

func (s *echoSwagger) OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) Swagger {
	ec := s.e.OPTIONS(path, h, m...)

	er := &echoRoute{
		EchoRoute: ec,
		route: route{
			path:       ec.Path,
			method:     ec.Method,
			funcName:   getFuncName(ec.Name),
			pathParams: getPathParams(ec.Path),
		},
	}

	s.routes = append(s.routes, er)

	return er
}

type echoGroup struct {
	g         *echo.Group
	groupName string
	groups    []*echoGroup
	routes    []*echoRoute
}

// Group creates a new sub-group with prefix and optional sub-group-level middleware.
func (s *echoGroup) Group(prefix string, m ...echo.MiddlewareFunc) Group {
	ec := &echoGroup{g: s.g.Group(prefix, m...), groupName: prefix}
	s.groups = append(s.groups, ec)

	return ec
}

func (s *echoGroup) POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) Swagger {
	ec := s.g.POST(path, h, m...)

	er := &echoRoute{
		EchoRoute: ec,
		route: route{
			path:       ec.Path,
			method:     ec.Method,
			funcName:   getFuncName(ec.Name),
			pathParams: getPathParams(ec.Path),
		},
	}

	s.routes = append(s.routes, er)

	return er
}

func (s *echoGroup) GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) Swagger {
	ec := s.g.GET(path, h, m...)

	er := &echoRoute{
		EchoRoute: ec,
		route: route{
			path:       ec.Path,
			method:     ec.Method,
			funcName:   getFuncName(ec.Name),
			pathParams: getPathParams(ec.Path),
		},
	}

	s.routes = append(s.routes, er)

	return er
}

func (s *echoGroup) PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) Swagger {
	ec := s.g.PUT(path, h, m...)

	er := &echoRoute{
		EchoRoute: ec,
		route: route{
			path:       ec.Path,
			method:     ec.Method,
			funcName:   getFuncName(ec.Name),
			pathParams: getPathParams(ec.Path),
		},
	}

	s.routes = append(s.routes, er)

	return er
}

func (s *echoGroup) DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) Swagger {
	ec := s.g.DELETE(path, h, m...)

	er := &echoRoute{
		EchoRoute: ec,
		route: route{
			path:       ec.Path,
			method:     ec.Method,
			funcName:   getFuncName(ec.Name),
			pathParams: getPathParams(ec.Path),
		},
	}

	s.routes = append(s.routes, er)

	return er
}

func (s *echoGroup) PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) Swagger {
	ec := s.g.PATCH(path, h, m...)

	er := &echoRoute{
		EchoRoute: ec,
		route: route{
			path:       ec.Path,
			method:     ec.Method,
			funcName:   getFuncName(ec.Name),
			pathParams: getPathParams(ec.Path),
		},
	}

	s.routes = append(s.routes, er)

	return er
}

func (s *echoGroup) OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) Swagger {
	ec := s.g.OPTIONS(path, h, m...)

	er := &echoRoute{
		EchoRoute: ec,
		route: route{
			path:       ec.Path,
			method:     ec.Method,
			funcName:   getFuncName(ec.Name),
			pathParams: getPathParams(ec.Path),
		},
	}

	s.routes = append(s.routes, er)

	return er
}

type echoRoute struct {
	EchoRoute *echo.Route
	route
}

func (r *echoRoute) Summary(value string) Swagger {
	r.summary = value
	return r
}

func (r *echoRoute) Description(value string) Swagger {
	r.description = value
	return r
}

func (r *echoRoute) Tags(value ...string) Swagger {
	r.tags = value
	return r
}

func (r *echoRoute) Accept(value ...string) Swagger {
	r.accepts = value
	return r
}

func (r *echoRoute) Produce(value ...string) Swagger {
	r.produces = value
	return r
}

func (r *echoRoute) Read(value interface{}) Swagger {
	r.reads = value
	return r
}

func (r *echoRoute) Returns(returns []ReturnType) Swagger {
	r.returns = returns
	return r
}

func (r *echoRoute) QueryParam(name, description, paramType string, required bool) Swagger {
	r.queryParams = append(r.queryParams, param{name, description, paramType, required})
	return r
}

func (r *echoRoute) HeaderParam(name, description, paramType string, required bool) Swagger {
	r.headerParams = append(r.headerParams, param{name, description, paramType, required})
	return r
}
