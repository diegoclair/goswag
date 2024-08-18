package echo

import (
	"testing"

	"github.com/diegoclair/goswag/internal/generator"
	"github.com/diegoclair/goswag/models"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestNewEcho(t *testing.T) {
	t.Run("should return echo instances", func(t *testing.T) {
		got := NewEcho()
		assert.NotNil(t, got)
		assert.NotNil(t, got.e)
	})
}

func TestEchoSwagger_Echo(t *testing.T) {
	t.Run("should return echo instance", func(t *testing.T) {
		s := &echoSwagger{
			e: echo.New(),
		}
		got := s.Echo()
		assert.NotNil(t, got)
		assert.NotNil(t, s.e)
		assert.Equal(t, s.e, got)
	})
}

func TestGroup(t *testing.T) {
	type args struct {
		prefix string
		m      []echo.MiddlewareFunc
	}
	tests := []struct {
		name string
		args args
		want echoGroup
	}{
		{
			name: "Test Group",
			args: args{
				prefix: "/test",
				m:      []echo.MiddlewareFunc{},
			},
			want: echoGroup{
				groupName: "/test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &echoSwagger{
				e: echo.New(),
			}
			got := s.Group(tt.args.prefix, tt.args.m...)
			assert.NotNil(t, got)

			assert.Equal(t, tt.want.groupName, s.groups[0].groupName)
		})
	}
}

func TestEchoSwagger_GET(t *testing.T) {
	type args struct {
		path string
		h    echo.HandlerFunc
		m    []echo.MiddlewareFunc
	}
	tests := []struct {
		name string
		args args
		want generator.Route
	}{
		{
			name: "Test GET",
			args: args{
				path: "/test/:id/",
				h:    func(c echo.Context) error { return nil },
				m:    []echo.MiddlewareFunc{},
			},
			want: generator.Route{
				Path:     "/test/:id/",
				Method:   "GET",
				FuncName: "func1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &echoSwagger{
				e: echo.New(),
			}
			got := s.GET(tt.args.path, tt.args.h, tt.args.m...)
			assert.NotNil(t, got)

			assert.Equal(t, tt.want, s.routes[0].Route)
		})
	}
}

func TestEchoSwagger_POST(t *testing.T) {
	type args struct {
		path string
		h    echo.HandlerFunc
		m    []echo.MiddlewareFunc
	}
	tests := []struct {
		name string
		args args
		want generator.Route
	}{
		{
			name: "Test POST",
			args: args{
				path: "/test/:id/",
				h:    func(c echo.Context) error { return nil },
				m:    []echo.MiddlewareFunc{},
			},
			want: generator.Route{
				Path:     "/test/:id/",
				Method:   "POST",
				FuncName: "func1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &echoSwagger{
				e: echo.New(),
			}
			got := s.POST(tt.args.path, tt.args.h, tt.args.m...)
			assert.NotNil(t, got)

			assert.Equal(t, tt.want, s.routes[0].Route)
		})
	}
}

func TestEchoSwagger_PUT(t *testing.T) {
	type args struct {
		path string
		h    echo.HandlerFunc
		m    []echo.MiddlewareFunc
	}
	tests := []struct {
		name string
		args args
		want generator.Route
	}{
		{
			name: "Test PUT",
			args: args{
				path: "/test/:id/",
				h:    func(c echo.Context) error { return nil },
				m:    []echo.MiddlewareFunc{},
			},
			want: generator.Route{
				Path:     "/test/:id/",
				Method:   "PUT",
				FuncName: "func1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &echoSwagger{
				e: echo.New(),
			}
			got := s.PUT(tt.args.path, tt.args.h, tt.args.m...)
			assert.NotNil(t, got)

			assert.Equal(t, tt.want, s.routes[0].Route)
		})
	}
}

func TestEchoSwagger_DELETE(t *testing.T) {
	type args struct {
		path string
		h    echo.HandlerFunc
		m    []echo.MiddlewareFunc
	}
	tests := []struct {
		name string
		args args
		want generator.Route
	}{
		{
			name: "Test DELETE",
			args: args{
				path: "/test/:id/",
				h:    func(c echo.Context) error { return nil },
				m:    []echo.MiddlewareFunc{},
			},
			want: generator.Route{
				Path:     "/test/:id/",
				Method:   "DELETE",
				FuncName: "func1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &echoSwagger{
				e: echo.New(),
			}
			got := s.DELETE(tt.args.path, tt.args.h, tt.args.m...)
			assert.NotNil(t, got)

			assert.Equal(t, tt.want, s.routes[0].Route)
		})
	}
}

func TestEchoSwagger_PATCH(t *testing.T) {
	type args struct {
		path string
		h    echo.HandlerFunc
		m    []echo.MiddlewareFunc
	}
	tests := []struct {
		name string
		args args
		want generator.Route
	}{
		{
			name: "Test PATCH",
			args: args{
				path: "/test/:id/",
				h:    func(c echo.Context) error { return nil },
				m:    []echo.MiddlewareFunc{},
			},
			want: generator.Route{
				Path:     "/test/:id/",
				Method:   "PATCH",
				FuncName: "func1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &echoSwagger{
				e: echo.New(),
			}
			got := s.PATCH(tt.args.path, tt.args.h, tt.args.m...)
			assert.NotNil(t, got)

			assert.Equal(t, tt.want, s.routes[0].Route)
		})
	}
}

func TestEchoSwagger_OPTIONS(t *testing.T) {
	type args struct {
		path string
		h    echo.HandlerFunc
		m    []echo.MiddlewareFunc
	}
	tests := []struct {
		name string
		args args
		want generator.Route
	}{
		{
			name: "Test OPTIONS",
			args: args{
				path: "/test/:id/",
				h:    func(c echo.Context) error { return nil },
				m:    []echo.MiddlewareFunc{},
			},
			want: generator.Route{
				Path:     "/test/:id/",
				Method:   "OPTIONS",
				FuncName: "func1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &echoSwagger{
				e: echo.New(),
			}
			got := s.OPTIONS(tt.args.path, tt.args.h, tt.args.m...)
			assert.NotNil(t, got)

			assert.Equal(t, tt.want, s.routes[0].Route)
		})
	}
}

func TestEchoSwagger_HEAD(t *testing.T) {
	type args struct {
		path string
		h    echo.HandlerFunc
		m    []echo.MiddlewareFunc
	}
	tests := []struct {
		name string
		args args
		want generator.Route
	}{
		{
			name: "Test HEAD",
			args: args{
				path: "/test/:id/",
				h:    func(c echo.Context) error { return nil },
				m:    []echo.MiddlewareFunc{},
			},
			want: generator.Route{
				Path:     "/test/:id/",
				Method:   "HEAD",
				FuncName: "func1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &echoSwagger{
				e: echo.New(),
			}
			got := s.HEAD(tt.args.path, tt.args.h, tt.args.m...)
			assert.NotNil(t, got)

			assert.Equal(t, tt.want, s.routes[0].Route)
		})
	}
}

func TestEchoGroup_Group(t *testing.T) {
	type args struct {
		prefix string
		m      []echo.MiddlewareFunc
	}
	tests := []struct {
		name string
		args args
		want echoGroup
	}{
		{
			name: "Test Group",
			args: args{
				prefix: "/test",
				m:      []echo.MiddlewareFunc{},
			},
			want: echoGroup{
				groupName: "/test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := echoGroup{
				g: echo.New().Group(""),
			}
			got := g.Group(tt.args.prefix, tt.args.m...)
			assert.NotNil(t, got)

			assert.Equal(t, tt.want.groupName, g.groups[0].groupName)
		})
	}
}

func TestEchoGroup_GET(t *testing.T) {
	type args struct {
		path string
		h    echo.HandlerFunc
		m    []echo.MiddlewareFunc
	}
	tests := []struct {
		name string
		args args
		want generator.Route
	}{
		{
			name: "Test GET",
			args: args{
				path: "/test/:id/",
				h:    func(c echo.Context) error { return nil },
				m:    []echo.MiddlewareFunc{},
			},
			want: generator.Route{
				Path:     "/test/:id/",
				Method:   "GET",
				FuncName: "func1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := echoGroup{
				g: echo.New().Group(""),
			}
			got := g.GET(tt.args.path, tt.args.h, tt.args.m...)
			assert.NotNil(t, got)

			assert.Equal(t, tt.want, g.routes[0].Route)
		})
	}
}

func TestEchoGroup_POST(t *testing.T) {
	type args struct {
		path string
		h    echo.HandlerFunc
		m    []echo.MiddlewareFunc
	}
	tests := []struct {
		name string
		args args
		want generator.Route
	}{
		{
			name: "Test POST",
			args: args{
				path: "/test/:id/",
				h:    func(c echo.Context) error { return nil },
				m:    []echo.MiddlewareFunc{},
			},
			want: generator.Route{
				Path:     "/test/:id/",
				Method:   "POST",
				FuncName: "func1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := echoGroup{
				g: echo.New().Group(""),
			}
			got := g.POST(tt.args.path, tt.args.h, tt.args.m...)
			assert.NotNil(t, got)

			assert.Equal(t, tt.want, g.routes[0].Route)
		})
	}
}

func TestEchoGroup_PUT(t *testing.T) {
	type args struct {
		path string
		h    echo.HandlerFunc
		m    []echo.MiddlewareFunc
	}
	tests := []struct {
		name string
		args args
		want generator.Route
	}{
		{
			name: "Test PUT",
			args: args{
				path: "/test/:id/",
				h:    func(c echo.Context) error { return nil },
				m:    []echo.MiddlewareFunc{},
			},
			want: generator.Route{
				Path:     "/test/:id/",
				Method:   "PUT",
				FuncName: "func1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := echoGroup{
				g: echo.New().Group(""),
			}
			got := g.PUT(tt.args.path, tt.args.h, tt.args.m...)
			assert.NotNil(t, got)

			assert.Equal(t, tt.want, g.routes[0].Route)
		})
	}
}

func TestEchoGroup_DELETE(t *testing.T) {
	type args struct {
		path string
		h    echo.HandlerFunc
		m    []echo.MiddlewareFunc
	}
	tests := []struct {
		name string
		args args
		want generator.Route
	}{
		{
			name: "Test DELETE",
			args: args{
				path: "/test/:id/",
				h:    func(c echo.Context) error { return nil },
				m:    []echo.MiddlewareFunc{},
			},
			want: generator.Route{
				Path:     "/test/:id/",
				Method:   "DELETE",
				FuncName: "func1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := echoGroup{
				g: echo.New().Group(""),
			}
			got := g.DELETE(tt.args.path, tt.args.h, tt.args.m...)
			assert.NotNil(t, got)

			assert.Equal(t, tt.want, g.routes[0].Route)
		})
	}
}

func TestEchoGroup_PATCH(t *testing.T) {
	type args struct {
		path string
		h    echo.HandlerFunc
		m    []echo.MiddlewareFunc
	}
	tests := []struct {
		name string
		args args
		want generator.Route
	}{
		{
			name: "Test PATCH",
			args: args{
				path: "/test/:id/",
				h:    func(c echo.Context) error { return nil },
				m:    []echo.MiddlewareFunc{},
			},
			want: generator.Route{
				Path:     "/test/:id/",
				Method:   "PATCH",
				FuncName: "func1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := echoGroup{
				g: echo.New().Group(""),
			}
			got := g.PATCH(tt.args.path, tt.args.h, tt.args.m...)
			assert.NotNil(t, got)

			assert.Equal(t, tt.want, g.routes[0].Route)
		})
	}
}

func TestEchoGroup_OPTIONS(t *testing.T) {
	type args struct {
		path string
		h    echo.HandlerFunc
		m    []echo.MiddlewareFunc
	}
	tests := []struct {
		name string
		args args
		want generator.Route
	}{
		{
			name: "Test OPTIONS",
			args: args{
				path: "/test/:id/",
				h:    func(c echo.Context) error { return nil },
				m:    []echo.MiddlewareFunc{},
			},
			want: generator.Route{
				Path:     "/test/:id/",
				Method:   "OPTIONS",
				FuncName: "func1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := echoGroup{
				g: echo.New().Group(""),
			}
			got := g.OPTIONS(tt.args.path, tt.args.h, tt.args.m...)
			assert.NotNil(t, got)

			assert.Equal(t, tt.want, g.routes[0].Route)
		})
	}
}

func TestEchoGroup_HEAD(t *testing.T) {
	type args struct {
		path string
		h    echo.HandlerFunc
		m    []echo.MiddlewareFunc
	}
	tests := []struct {
		name string
		args args
		want generator.Route
	}{
		{
			name: "Test HEAD",
			args: args{
				path: "/test/:id/",
				h:    func(c echo.Context) error { return nil },
				m:    []echo.MiddlewareFunc{},
			},
			want: generator.Route{
				Path:     "/test/:id/",
				Method:   "HEAD",
				FuncName: "func1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := echoGroup{
				g: echo.New().Group(""),
			}
			got := g.HEAD(tt.args.path, tt.args.h, tt.args.m...)
			assert.NotNil(t, got)

			assert.Equal(t, tt.want, g.routes[0].Route)
		})
	}
}

func TestEchoRoute_Summary(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name string
		args args
		want generator.Route
	}{
		{
			name: "Test Summary",
			args: args{
				value: "Test Summary",
			},
			want: generator.Route{
				Summary: "Test Summary",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &echoRoute{
				Route: generator.Route{},
			}
			got := r.Summary(tt.args.value)
			assert.NotNil(t, got)

			assert.Equal(t, tt.want.Summary, r.Route.Summary)
		})
	}
}

func TestEchoRoute_Description(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name string
		args args
		want generator.Route
	}{
		{
			name: "Test Description",
			args: args{
				value: "Test Description",
			},
			want: generator.Route{
				Description: "Test Description",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &echoRoute{
				Route: generator.Route{},
			}
			got := r.Description(tt.args.value)
			assert.NotNil(t, got)

			assert.Equal(t, tt.want.Description, r.Route.Description)
		})
	}
}

func TestEchoRoute_Tags(t *testing.T) {
	type args struct {
		value []string
	}
	tests := []struct {
		name string
		args args
		want generator.Route
	}{
		{
			name: "Test Tags",
			args: args{
				value: []string{"Test", "Tags"},
			},
			want: generator.Route{
				Tags: []string{"Test", "Tags"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &echoRoute{
				Route: generator.Route{},
			}
			got := r.Tags(tt.args.value...)
			assert.NotNil(t, got)

			assert.Equal(t, tt.want.Tags, r.Route.Tags)
		})
	}
}

func TestEchoRoute_Accepts(t *testing.T) {
	type args struct {
		value []string
	}
	tests := []struct {
		name string
		args args
		want generator.Route
	}{
		{
			name: "Test Accepts",
			args: args{
				value: []string{"Test", "Accepts"},
			},
			want: generator.Route{
				Accepts: []string{"Test", "Accepts"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &echoRoute{
				Route: generator.Route{},
			}
			got := r.Accepts(tt.args.value...)
			assert.NotNil(t, got)

			assert.Equal(t, tt.want.Accepts, r.Route.Accepts)
		})
	}
}

func TestEchoRoute_Produces(t *testing.T) {
	type args struct {
		value []string
	}
	tests := []struct {
		name string
		args args
		want generator.Route
	}{
		{
			name: "Test Produces",
			args: args{
				value: []string{"Test", "Produces"},
			},
			want: generator.Route{
				Produces: []string{"Test", "Produces"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &echoRoute{
				Route: generator.Route{},
			}
			got := r.Produces(tt.args.value...)
			assert.NotNil(t, got)

			assert.Equal(t, tt.want.Produces, r.Route.Produces)
		})
	}
}

func TestEchoRoute_Read(t *testing.T) {
	type args struct {
		value interface{}
	}
	tests := []struct {
		name string
		args args
		want generator.Route
	}{
		{
			name: "Test Read",
			args: args{
				value: "Test Read",
			},
			want: generator.Route{
				Reads: "Test Read",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &echoRoute{
				Route: generator.Route{},
			}
			got := r.Read(tt.args.value)
			assert.NotNil(t, got)

			assert.Equal(t, tt.want.Reads, r.Route.Reads)
		})
	}
}

func TestEchoRoute_Returns(t *testing.T) {
	type args struct {
		returns []models.ReturnType
	}
	tests := []struct {
		name string
		args args
		want generator.Route
	}{
		{
			name: "Test Returns",
			args: args{
				returns: []models.ReturnType{
					{
						StatusCode: 200,
						Body:       "Test",
					},
				},
			},
			want: generator.Route{
				Returns: []models.ReturnType{
					{
						StatusCode: 200,
						Body:       "Test",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &echoRoute{
				Route: generator.Route{},
			}
			got := r.Returns(tt.args.returns)
			assert.NotNil(t, got)

			assert.Equal(t, tt.want.Returns, r.Route.Returns)
		})
	}
}

func TestEchoRoute_QueryParam(t *testing.T) {
	type args struct {
		name        string
		description string
		paramType   string
		required    bool
	}
	tests := []struct {
		name string
		args args
		want generator.Route
	}{
		{
			name: "Test QueryParam",
			args: args{
				name:        "Test",
				description: "Test",
				paramType:   "Test",
				required:    true,
			},
			want: generator.Route{
				QueryParams: []generator.Param{
					{
						Name:        "Test",
						Description: "Test",
						ParamType:   "Test",
						Required:    true,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &echoRoute{
				Route: generator.Route{},
			}
			got := r.QueryParam(tt.args.name, tt.args.description, tt.args.paramType, tt.args.required)
			assert.NotNil(t, got)

			assert.Equal(t, tt.want.QueryParams, r.Route.QueryParams)
		})
	}
}

func TestEchoRoute_HeaderParam(t *testing.T) {
	type args struct {
		name        string
		description string
		paramType   string
		required    bool
	}
	tests := []struct {
		name string
		args args
		want generator.Route
	}{
		{
			name: "Test HeaderParam",
			args: args{
				name:        "Test",
				description: "Test",
				paramType:   "Test",
				required:    true,
			},
			want: generator.Route{
				HeaderParams: []generator.Param{
					{
						Name:        "Test",
						Description: "Test",
						ParamType:   "Test",
						Required:    true,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &echoRoute{
				Route: generator.Route{},
			}
			got := r.HeaderParam(tt.args.name, tt.args.description, tt.args.paramType, tt.args.required)
			assert.NotNil(t, got)

			assert.Equal(t, tt.want.HeaderParams, r.Route.HeaderParams)
		})
	}
}
