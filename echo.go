package goswag

import (
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"
)

type Router interface {
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) Swagger
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) Swagger
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) Swagger
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) Swagger
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) Swagger
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) Swagger
}

type Group interface {
	Router
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

// Group automatically create tags for the swagger documentation.
//
// Group creates a new router group with prefix and optional group-level middleware.
func (s *echoSwagger) Group(prefix string, m ...echo.MiddlewareFunc) Group {
	g := &echoGroup{g: s.e.Group(prefix, m...), groupName: prefix}
	s.groups = append(s.groups, g)

	return g
}

// POST registers a new POST route for a path with matching handler in the
// router with optional route-level middleware.
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

// GET registers a new GET route for a path with matching handler in the router
// with optional route-level middleware.
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

// PUT registers a new PUT route for a path with matching handler in the
// router with optional route-level middleware.
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

// DELETE registers a new DELETE route for a path with matching handler in the router
// with optional route-level middleware.
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

// PATCH registers a new PATCH route for a path with matching handler in the
// router with optional route-level middleware.
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

// OPTIONS registers a new OPTIONS route for a path with matching handler in the
// router with optional route-level middleware.
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

// POST implements `Echo#POST()` for sub-routes within the Group.
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

// GET implements `Echo#GET()` for sub-routes within the Group.
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

// PUT implements `Echo#PUT()` for sub-routes within the Group.
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

// DELETE implements `Echo#DELETE()` for sub-routes within the Group.
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

// PATCH implements `Echo#PATCH()` for sub-routes within the Group.
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

// OPTIONS implements `Echo#OPTIONS()` for sub-routes within the Group.
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

// Summary is used to define the summary of the route.
func (r *echoRoute) Summary(value string) Swagger {
	r.summary = value
	return r
}

// The default value is the same as the summary.
func (r *echoRoute) Description(value string) Swagger {
	r.description = value
	return r
}

// The name of group will be used as default if it is not empty and the tags are not defined.
func (r *echoRoute) Tags(value ...string) Swagger {
	r.tags = value
	return r
}

// The default value is json.
// If you want to add a different value, check the swag documentation to know what are the possible values.
// swag docs: https://github.com/swaggo/swag#mime-types
func (r *echoRoute) Accept(value ...string) Swagger {
	r.accepts = value
	return r
}

// The default value is json.
// If you want to add a different value, check the swag documentation to know what are the possible values.
// swag docs: https://github.com/swaggo/swag#mime-types
func (r *echoRoute) Produce(value ...string) Swagger {
	r.produces = value
	return r
}

// Read is used to define the request body of the route.
func (r *echoRoute) Read(value interface{}) Swagger {
	r.reads = value
	return r
}

// Returns is used to define the return of the route.
// The first parameter is the status code.
// The second parameter is the body of the response.
// The third parameter is used to override the fields of the response body, it is is optional.
// Example:
// if you have a response body like this:
//
//	type ResponseBody struct {
//		ID   string `json:"id"`
//		Data interface `json:"data"`
//	}
//
// the swagger will be generated with the data field as a string field.
// if you want to override the data field and specify that it is a struct for example, you can do this:
// OverrideStructFields: map[string]interface{}{"data": SomeStruct{}}
// where the SomeStruct{} is the struct that you want to use to override the data field.
//
// It accepts generic structs as well, but only for the first struct, if you have more deep generic fields, it may not work.
//
// ATTENTION: The OverrideStructFields don't work with GenericStructs yet.
//
// Example using generic struct:
//
//	type ResponseBody[T any] struct {
//			Data T   `json:"data"`
//	}
//
// Then you will set the body like this:
//
//	ReturnType {
//		StatusCode: http.StatusOK,
//		Body: ResponseBody[SomeStruct]{},
//	}
func (r *echoRoute) Returns(returns []ReturnType) Swagger {
	r.returns = returns
	return r
}

// QueryParam is used to define the query parameters of the route and if it is required or not.
func (r *echoRoute) QueryParam(name, description, paramType string, required bool) Swagger {
	r.queryParams = append(r.queryParams, param{name, description, paramType, required})
	return r
}

// HeaderParam is used to define the header parameters of the route and if it is required or not.
func (r *echoRoute) HeaderParam(name, description, paramType string, required bool) Swagger {
	r.headerParams = append(r.headerParams, param{name, description, paramType, required})
	return r
}
