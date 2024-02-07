package gin

import (
	"net/http"
	"path"

	"github.com/diegoclair/goswag/internal/generator"
	"github.com/diegoclair/goswag/models"
	"github.com/gin-gonic/gin"
)

type ginSwagger struct {
	g      *gin.Engine
	groups []*ginGroup
	routes []*ginRoute
}

func NewGin(g *gin.Engine) *ginSwagger {
	return &ginSwagger{
		g: g,
	}
}

func (s *ginSwagger) Gin() *gin.Engine {

	return s.g
}

func (s *ginSwagger) GenerateSwagger() {
	generator.GenerateSwagger(toGoSwagRoute(s.routes), toGoSwagGroup(s.groups))
}

func (s *ginSwagger) Group(relativePath string, handlers ...gin.HandlerFunc) models.GinRouter {
	g := &ginGroup{gg: s.g.Group(relativePath, handlers...), groupName: relativePath}
	s.groups = append(s.groups, g)

	return g
}

func (s *ginSwagger) Handle(httpMethod, relativePath string, handlers ...gin.HandlerFunc) models.Swagger {
	s.g.Handle(httpMethod, relativePath, handlers...)

	gr := &ginRoute{
		Route: generator.Route{
			Path:       relativePath,
			Method:     httpMethod,
			FuncName:   getFuncName(handlers...),
			PathParams: getPathParams(relativePath),
		},
	}

	s.routes = append(s.routes, gr)

	return gr
}

func (s *ginSwagger) POST(relativePath string, handlers ...gin.HandlerFunc) models.Swagger {
	s.g.POST(relativePath, handlers...)

	gr := &ginRoute{
		Route: generator.Route{
			Path:       relativePath,
			Method:     http.MethodPost,
			FuncName:   getFuncName(handlers...),
			PathParams: getPathParams(relativePath),
		},
	}

	s.routes = append(s.routes, gr)

	return gr
}

func (s *ginSwagger) GET(relativePath string, handlers ...gin.HandlerFunc) models.Swagger {
	s.g.GET(relativePath, handlers...)

	gr := &ginRoute{
		Route: generator.Route{
			Path:       relativePath,
			Method:     http.MethodGet,
			FuncName:   getFuncName(handlers...),
			PathParams: getPathParams(relativePath),
		},
	}

	s.routes = append(s.routes, gr)

	return gr
}

func (s *ginSwagger) PUT(relativePath string, handlers ...gin.HandlerFunc) models.Swagger {
	s.g.PUT(relativePath, handlers...)

	gr := &ginRoute{
		Route: generator.Route{
			Path:       relativePath,
			Method:     http.MethodPut,
			FuncName:   getFuncName(handlers...),
			PathParams: getPathParams(relativePath),
		},
	}

	s.routes = append(s.routes, gr)

	return gr
}

func (s *ginSwagger) DELETE(relativePath string, handlers ...gin.HandlerFunc) models.Swagger {
	s.g.DELETE(relativePath, handlers...)

	gr := &ginRoute{
		Route: generator.Route{
			Path:       relativePath,
			Method:     http.MethodDelete,
			FuncName:   getFuncName(handlers...),
			PathParams: getPathParams(relativePath),
		},
	}

	s.routes = append(s.routes, gr)

	return gr
}

func (s *ginSwagger) PATCH(relativePath string, handlers ...gin.HandlerFunc) models.Swagger {
	s.g.PATCH(relativePath, handlers...)

	gr := &ginRoute{
		Route: generator.Route{
			Path:       relativePath,
			Method:     http.MethodPatch,
			FuncName:   getFuncName(handlers...),
			PathParams: getPathParams(relativePath),
		},
	}

	s.routes = append(s.routes, gr)

	return gr
}

func (s *ginSwagger) OPTIONS(relativePath string, handlers ...gin.HandlerFunc) models.Swagger {
	s.g.OPTIONS(relativePath, handlers...)

	gr := &ginRoute{
		Route: generator.Route{
			Path:       relativePath,
			Method:     http.MethodOptions,
			FuncName:   getFuncName(handlers...),
			PathParams: getPathParams(relativePath),
		},
	}

	s.routes = append(s.routes, gr)

	return gr
}

func (s *ginSwagger) HEAD(relativePath string, handlers ...gin.HandlerFunc) models.Swagger {
	s.g.HEAD(relativePath, handlers...)

	gr := &ginRoute{
		Route: generator.Route{
			Path:       relativePath,
			Method:     http.MethodHead,
			FuncName:   getFuncName(handlers...),
			PathParams: getPathParams(relativePath),
		},
	}

	s.routes = append(s.routes, gr)

	return gr
}

type ginGroup struct {
	gg        *gin.RouterGroup
	groupName string
	routes    []*ginRoute
}

func (g *ginGroup) Handle(httpMethod, relativePath string, handlers ...gin.HandlerFunc) models.Swagger {
	g.gg.Handle(httpMethod, relativePath, handlers...)
	fullPath := path.Join(g.groupName, relativePath)

	gr := &ginRoute{
		Route: generator.Route{
			Path:       fullPath,
			Method:     httpMethod,
			FuncName:   getFuncName(handlers...),
			PathParams: getPathParams(fullPath),
		},
	}

	g.routes = append(g.routes, gr)

	return gr
}

func (g *ginGroup) POST(relativePath string, handlers ...gin.HandlerFunc) models.Swagger {
	g.gg.POST(relativePath, handlers...)
	fullPath := path.Join(g.groupName, relativePath)

	gr := &ginRoute{
		Route: generator.Route{
			Path:       fullPath,
			Method:     http.MethodPost,
			FuncName:   getFuncName(handlers...),
			PathParams: getPathParams(fullPath),
		},
	}

	g.routes = append(g.routes, gr)

	return gr
}

func (g *ginGroup) GET(relativePath string, handlers ...gin.HandlerFunc) models.Swagger {
	g.gg.GET(relativePath, handlers...)
	fullPath := path.Join(g.groupName, relativePath)

	gr := &ginRoute{
		Route: generator.Route{
			Path:       fullPath,
			Method:     http.MethodGet,
			FuncName:   getFuncName(handlers...),
			PathParams: getPathParams(fullPath),
		},
	}

	g.routes = append(g.routes, gr)

	return gr
}

func (g *ginGroup) PUT(relativePath string, handlers ...gin.HandlerFunc) models.Swagger {
	g.gg.PUT(relativePath, handlers...)
	fullPath := path.Join(g.groupName, relativePath)

	gr := &ginRoute{
		Route: generator.Route{
			Path:       fullPath,
			Method:     http.MethodPut,
			FuncName:   getFuncName(handlers...),
			PathParams: getPathParams(fullPath),
		},
	}

	g.routes = append(g.routes, gr)

	return gr
}

func (g *ginGroup) DELETE(relativePath string, handlers ...gin.HandlerFunc) models.Swagger {
	g.gg.DELETE(relativePath, handlers...)
	fullPath := path.Join(g.groupName, relativePath)

	gr := &ginRoute{
		Route: generator.Route{
			Path:       fullPath,
			Method:     http.MethodDelete,
			FuncName:   getFuncName(handlers...),
			PathParams: getPathParams(fullPath),
		},
	}

	g.routes = append(g.routes, gr)

	return gr
}

func (g *ginGroup) PATCH(relativePath string, handlers ...gin.HandlerFunc) models.Swagger {
	g.gg.PATCH(relativePath, handlers...)
	fullPath := path.Join(g.groupName, relativePath)

	gr := &ginRoute{
		Route: generator.Route{
			Path:       fullPath,
			Method:     http.MethodPatch,
			FuncName:   getFuncName(handlers...),
			PathParams: getPathParams(fullPath),
		},
	}

	g.routes = append(g.routes, gr)

	return gr
}

func (g *ginGroup) OPTIONS(relativePath string, handlers ...gin.HandlerFunc) models.Swagger {
	g.gg.OPTIONS(relativePath, handlers...)
	fullPath := path.Join(g.groupName, relativePath)

	gr := &ginRoute{
		Route: generator.Route{
			Path:       fullPath,
			Method:     http.MethodOptions,
			FuncName:   getFuncName(handlers...),
			PathParams: getPathParams(fullPath),
		},
	}

	g.routes = append(g.routes, gr)

	return gr
}

func (g *ginGroup) HEAD(relativePath string, handlers ...gin.HandlerFunc) models.Swagger {
	g.gg.HEAD(relativePath, handlers...)
	fullPath := path.Join(g.groupName, relativePath)

	gr := &ginRoute{
		Route: generator.Route{
			Path:       fullPath,
			Method:     http.MethodHead,
			FuncName:   getFuncName(handlers...),
			PathParams: getPathParams(fullPath),
		},
	}

	g.routes = append(g.routes, gr)

	return gr
}

type ginRoute struct {
	Route generator.Route
}

func (r *ginRoute) Summary(summary string) models.Swagger {
	r.Route.Summary = summary
	return r
}

func (r *ginRoute) Description(description string) models.Swagger {
	r.Route.Description = description
	return r
}

func (r *ginRoute) Tags(tags ...string) models.Swagger {
	r.Route.Tags = tags
	return r
}

func (r *ginRoute) Accepts(accepts ...string) models.Swagger {
	r.Route.Accepts = accepts
	return r
}

func (r *ginRoute) Produces(produces ...string) models.Swagger {
	r.Route.Produces = produces
	return r
}

func (r *ginRoute) Read(reads interface{}) models.Swagger {
	r.Route.Reads = reads
	return r
}

func (r *ginRoute) Returns(returns []models.ReturnType) models.Swagger {
	r.Route.Returns = returns
	return r
}

func (r *ginRoute) QueryParam(name, description, paramType string, required bool) models.Swagger {
	r.Route.QueryParams = append(r.Route.QueryParams, generator.Param{
		Name:        name,
		Description: description,
		ParamType:   paramType,
		Required:    required,
	})

	return r
}

func (r *ginRoute) HeaderParam(name, description, paramType string, required bool) models.Swagger {
	r.Route.HeaderParams = append(r.Route.HeaderParams, generator.Param{
		Name:        name,
		Description: description,
		ParamType:   paramType,
		Required:    required,
	})

	return r
}
