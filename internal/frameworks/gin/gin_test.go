package gin

import (
	"testing"

	"github.com/diegoclair/goswag/internal/generator"
	"github.com/diegoclair/goswag/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewGin(t *testing.T) {
	t.Run("should return gin instances", func(t *testing.T) {
		g := gin.Default()
		got := NewGin(g)
		assert.NotNil(t, got)
		assert.NotNil(t, got.g)
		assert.Equal(t, g, got.g)
	})
}

func TestGinSwagger_Gin(t *testing.T) {
	t.Run("should return gin instance", func(t *testing.T) {
		g := gin.Default()
		got := NewGin(g)
		assert.NotNil(t, got.Gin())
		assert.Equal(t, g, got.Gin())
	})
}

func TestGinSwagger_Group(t *testing.T) {
	t.Run("should return gin group", func(t *testing.T) {
		g := gin.Default()
		got := NewGin(g)
		group := got.Group("/test")
		assert.NotNil(t, group)
		assert.Equal(t, "/test", group.(*ginGroup).groupName)
	})
}

func TestGinSwagger_Handle(t *testing.T) {
	type args struct {
		httpMethod   string
		relativePath string
		handlers     []gin.HandlerFunc
	}

	tests := []struct {
		name string
		args args
		want generator.Route
	}{
		{
			name: "should return gin route",
			args: args{
				httpMethod:   "GET",
				relativePath: "/test/:id/",
				handlers:     []gin.HandlerFunc{func(c *gin.Context) {}},
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
			g := gin.Default()
			got := NewGin(g)
			got.Handle(tt.args.httpMethod, tt.args.relativePath, tt.args.handlers...)
			assert.Equal(t, tt.want, got.routes[0].Route)
		})
	}
}

func TestGinSwagger_POST(t *testing.T) {
	type args struct {
		relativePath string
		handlers     []gin.HandlerFunc
	}

	tests := []struct {
		name string
		args args
		want generator.Route
	}{
		{
			name: "should return gin route",
			args: args{
				relativePath: "/test/:id/",
				handlers:     []gin.HandlerFunc{func(c *gin.Context) {}},
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
			g := gin.Default()
			got := NewGin(g)
			got.POST(tt.args.relativePath, tt.args.handlers...)
			assert.Equal(t, tt.want, got.routes[0].Route)
		})
	}
}

func TestGinSwagger_GET(t *testing.T) {
	type args struct {
		relativePath string
		handlers     []gin.HandlerFunc
	}

	tests := []struct {
		name string
		args args
		want generator.Route
	}{
		{
			name: "should return gin route",
			args: args{
				relativePath: "/test/:id/",
				handlers:     []gin.HandlerFunc{func(c *gin.Context) {}},
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
			g := gin.Default()
			got := NewGin(g)
			got.GET(tt.args.relativePath, tt.args.handlers...)
			assert.Equal(t, tt.want, got.routes[0].Route)
		})
	}
}

func TestGinSwagger_PUT(t *testing.T) {
	type args struct {
		relativePath string
		handlers     []gin.HandlerFunc
	}

	tests := []struct {
		name string
		args args
		want generator.Route
	}{
		{
			name: "should return gin route",
			args: args{
				relativePath: "/test/:id/",
				handlers:     []gin.HandlerFunc{func(c *gin.Context) {}},
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
			g := gin.Default()
			got := NewGin(g)
			got.PUT(tt.args.relativePath, tt.args.handlers...)
			assert.Equal(t, tt.want, got.routes[0].Route)
		})
	}
}

func TestGinSwagger_DELETE(t *testing.T) {
	type args struct {
		relativePath string
		handlers     []gin.HandlerFunc
	}

	tests := []struct {
		name string
		args args
		want generator.Route
	}{
		{
			name: "should return gin route",
			args: args{
				relativePath: "/test/:id/",
				handlers:     []gin.HandlerFunc{func(c *gin.Context) {}},
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
			g := gin.Default()
			got := NewGin(g)
			got.DELETE(tt.args.relativePath, tt.args.handlers...)
			assert.Equal(t, tt.want, got.routes[0].Route)
		})
	}
}

func TestGinSwagger_PATCH(t *testing.T) {
	type args struct {
		relativePath string
		handlers     []gin.HandlerFunc
	}

	tests := []struct {
		name string
		args args
		want generator.Route
	}{
		{
			name: "should return gin route",
			args: args{
				relativePath: "/test/:id/",
				handlers:     []gin.HandlerFunc{func(c *gin.Context) {}},
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
			g := gin.Default()
			got := NewGin(g)
			got.PATCH(tt.args.relativePath, tt.args.handlers...)
			assert.Equal(t, tt.want, got.routes[0].Route)
		})
	}
}

func TestGinSwagger_OPTIONS(t *testing.T) {
	type args struct {
		relativePath string
		handlers     []gin.HandlerFunc
	}

	tests := []struct {
		name string
		args args
		want generator.Route
	}{
		{
			name: "should return gin route",
			args: args{
				relativePath: "/test/:id/",
				handlers:     []gin.HandlerFunc{func(c *gin.Context) {}},
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
			g := gin.Default()
			got := NewGin(g)
			got.OPTIONS(tt.args.relativePath, tt.args.handlers...)
			assert.Equal(t, tt.want, got.routes[0].Route)
		})
	}
}

func TestGinSwagger_HEAD(t *testing.T) {
	type args struct {
		relativePath string
		handlers     []gin.HandlerFunc
	}

	tests := []struct {
		name string
		args args
		want generator.Route
	}{
		{
			name: "should return gin route",
			args: args{
				relativePath: "/test/:id/",
				handlers:     []gin.HandlerFunc{func(c *gin.Context) {}},
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
			g := gin.Default()
			got := NewGin(g)
			got.HEAD(tt.args.relativePath, tt.args.handlers...)
			assert.Equal(t, tt.want, got.routes[0].Route)
		})
	}
}

func TestGinGroup_Handle(t *testing.T) {
	type args struct {
		httpMethod   string
		relativePath string
		handlers     []gin.HandlerFunc
	}

	tests := []struct {
		name string
		args args
		want generator.Route
	}{
		{
			name: "should return gin route",
			args: args{
				httpMethod:   "GET",
				relativePath: "/test/:id/",
				handlers:     []gin.HandlerFunc{func(c *gin.Context) {}},
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
			g := &ginGroup{
				gg: gin.Default().Group(""),
			}
			group := g.Handle(tt.args.httpMethod, tt.args.relativePath, tt.args.handlers...)
			assert.Equal(t, tt.want, group.(*ginRoute).Route)
		})
	}
}

func TestGinGroup_POST(t *testing.T) {
	type args struct {
		relativePath string
		handlers     []gin.HandlerFunc
	}

	tests := []struct {
		name string
		args args
		want generator.Route
	}{
		{
			name: "should return gin route",
			args: args{
				relativePath: "/test/:id/",
				handlers:     []gin.HandlerFunc{func(c *gin.Context) {}},
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
			g := &ginGroup{
				gg: gin.Default().Group(""),
			}
			group := g.POST(tt.args.relativePath, tt.args.handlers...)
			assert.Equal(t, tt.want, group.(*ginRoute).Route)
		})
	}
}

func TestGinGroup_GET(t *testing.T) {
	type args struct {
		relativePath string
		handlers     []gin.HandlerFunc
	}

	tests := []struct {
		name string
		args args
		want generator.Route
	}{
		{
			name: "should return gin route",
			args: args{
				relativePath: "/test/:id/",
				handlers:     []gin.HandlerFunc{func(c *gin.Context) {}},
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
			g := &ginGroup{
				gg: gin.Default().Group(""),
			}
			group := g.GET(tt.args.relativePath, tt.args.handlers...)
			assert.Equal(t, tt.want, group.(*ginRoute).Route)
		})
	}
}

func TestGinGroup_PUT(t *testing.T) {
	type args struct {
		relativePath string
		handlers     []gin.HandlerFunc
	}

	tests := []struct {
		name string
		args args
		want generator.Route
	}{
		{
			name: "should return gin route",
			args: args{
				relativePath: "/test/:id/",
				handlers:     []gin.HandlerFunc{func(c *gin.Context) {}},
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
			g := &ginGroup{
				gg: gin.Default().Group(""),
			}
			group := g.PUT(tt.args.relativePath, tt.args.handlers...)
			assert.Equal(t, tt.want, group.(*ginRoute).Route)
		})
	}
}

func TestGinGroup_DELETE(t *testing.T) {
	type args struct {
		relativePath string
		handlers     []gin.HandlerFunc
	}

	tests := []struct {
		name string
		args args
		want generator.Route
	}{
		{
			name: "should return gin route",
			args: args{
				relativePath: "/test/:id/",
				handlers:     []gin.HandlerFunc{func(c *gin.Context) {}},
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
			g := &ginGroup{
				gg: gin.Default().Group(""),
			}
			group := g.DELETE(tt.args.relativePath, tt.args.handlers...)
			assert.Equal(t, tt.want, group.(*ginRoute).Route)
		})
	}
}

func TestGinGroup_PATCH(t *testing.T) {
	type args struct {
		relativePath string
		handlers     []gin.HandlerFunc
	}

	tests := []struct {
		name string
		args args
		want generator.Route
	}{
		{
			name: "should return gin route",
			args: args{
				relativePath: "/test/:id/",
				handlers:     []gin.HandlerFunc{func(c *gin.Context) {}},
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
			g := &ginGroup{
				gg: gin.Default().Group(""),
			}
			group := g.PATCH(tt.args.relativePath, tt.args.handlers...)
			assert.Equal(t, tt.want, group.(*ginRoute).Route)
		})
	}
}

func TestGinGroup_OPTIONS(t *testing.T) {
	type args struct {
		relativePath string
		handlers     []gin.HandlerFunc
	}

	tests := []struct {
		name string
		args args
		want generator.Route
	}{
		{
			name: "should return gin route",
			args: args{
				relativePath: "/test/:id/",
				handlers:     []gin.HandlerFunc{func(c *gin.Context) {}},
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
			g := &ginGroup{
				gg: gin.Default().Group(""),
			}
			group := g.OPTIONS(tt.args.relativePath, tt.args.handlers...)
			assert.Equal(t, tt.want, group.(*ginRoute).Route)
		})
	}
}

func TestGinGroup_HEAD(t *testing.T) {
	type args struct {
		relativePath string
		handlers     []gin.HandlerFunc
	}

	tests := []struct {
		name string
		args args
		want generator.Route
	}{
		{
			name: "should return gin route",
			args: args{
				relativePath: "/test/:id/",
				handlers:     []gin.HandlerFunc{func(c *gin.Context) {}},
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
			g := &ginGroup{
				gg: gin.Default().Group(""),
			}
			group := g.HEAD(tt.args.relativePath, tt.args.handlers...)
			assert.Equal(t, tt.want, group.(*ginRoute).Route)
		})
	}
}

func TestGinRoute_Summary(t *testing.T) {
	t.Run("should add summary", func(t *testing.T) {
		g := &ginRoute{}
		got := g.Summary("test")
		assert.NotNil(t, got)
		assert.Equal(t, "test", g.Route.Summary)
	})
}

func TestGinRoute_Description(t *testing.T) {
	t.Run("should add description", func(t *testing.T) {
		g := &ginRoute{}
		got := g.Description("test")
		assert.NotNil(t, got)
		assert.Equal(t, "test", g.Route.Description)
	})
}

func TestGinRoute_Tags(t *testing.T) {
	t.Run("should add tags", func(t *testing.T) {
		g := &ginRoute{}
		got := g.Tags("test")
		assert.NotNil(t, got)
		assert.Equal(t, []string{"test"}, g.Route.Tags)
	})
}

func TestGinRoute_Accepts(t *testing.T) {
	t.Run("should add accepts", func(t *testing.T) {
		g := &ginRoute{}
		got := g.Accepts("test")
		assert.NotNil(t, got)
		assert.Equal(t, []string{"test"}, g.Route.Accepts)
	})
}

func TestGinRoute_Produces(t *testing.T) {
	t.Run("should add produces", func(t *testing.T) {
		g := &ginRoute{}
		got := g.Produces("test")
		assert.NotNil(t, got)
		assert.Equal(t, []string{"test"}, g.Route.Produces)
	})
}

func TestGinRoute_Read(t *testing.T) {
	type testStruct struct {
		Name string
	}

	t.Run("should add read", func(t *testing.T) {
		g := &ginRoute{}
		got := g.Read(testStruct{})
		assert.NotNil(t, got)
		assert.Equal(t, testStruct{}, g.Route.Reads)
	})
}

func TestGinRoute_Returns(t *testing.T) {
	t.Run("should add returns", func(t *testing.T) {
		g := &ginRoute{}
		got := g.Returns([]models.ReturnType{})
		assert.NotNil(t, got)
		assert.Equal(t, []models.ReturnType{}, g.Route.Returns)
	})
}

func TestGinRoute_QueryParam(t *testing.T) {
	t.Run("should add query param", func(t *testing.T) {
		g := &ginRoute{}
		got := g.QueryParam("test", "test", "test", true)
		assert.NotNil(t, got)
		assert.Equal(t, []generator.Param{{Name: "test", Description: "test", ParamType: "test", Required: true}}, g.Route.QueryParams)
	})
}

func TestGinRoute_HeaderParam(t *testing.T) {
	t.Run("should add header param", func(t *testing.T) {
		g := &ginRoute{}
		got := g.HeaderParam("test", "test", "test", true)
		assert.NotNil(t, got)
		assert.Equal(t, []generator.Param{{Name: "test", Description: "test", ParamType: "test", Required: true}}, g.Route.HeaderParams)
	})
}

func TestGinRoute_PathParam(t *testing.T) {
	t.Run("should add path param", func(t *testing.T) {
		g := &ginRoute{}
		got := g.PathParam("test", "test", "test", true)
		assert.NotNil(t, got)
		assert.Equal(t, []generator.Param{{Name: "test", Description: "test", ParamType: "test", Required: true}}, g.Route.PathParams)
	})
}
