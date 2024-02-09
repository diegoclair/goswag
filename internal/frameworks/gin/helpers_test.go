package gin

import (
	"reflect"
	"testing"

	"github.com/diegoclair/goswag/internal/generator"
	"github.com/gin-gonic/gin"
)

func handler1(c *gin.Context) {}
func handler2(c *gin.Context) {}
func handler3(c *gin.Context) {}

func TestGetFuncName(t *testing.T) {
	type args struct {
		handlers []gin.HandlerFunc
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Should return the function name of the last handler",
			args: args{
				handlers: []gin.HandlerFunc{handler1, handler2, handler3},
			},
			want: "handler3",
		},
		{
			name: "Should return the function name of the last handler",
			args: args{
				handlers: []gin.HandlerFunc{handler1, handler2},
			},
			want: "handler2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getFuncName(tt.args.handlers...); got != tt.want {
				t.Errorf("getFuncName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPathParams(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Should return the path parameters",
			args: args{
				path: "/path/:id",
			},
			want: []string{"id"},
		},
		{
			name: "Should return the path parameters",
			args: args{
				path: "/path/:id/:name",
			},
			want: []string{"id", "name"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getPathParams(tt.args.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getPathParams() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToGoSwagRoute(t *testing.T) {
	type args struct {
		from []*ginRoute
	}
	tests := []struct {
		name string
		args args
		want []generator.Route
	}{
		{
			name: "Should return the generator.Route",
			args: args{from: []*ginRoute{
				{
					Route: generator.Route{
						Method: "GET",
					},
				},
			}},
			want: []generator.Route{
				{
					Method: "GET",
				},
			},
		},
		{
			name: "Should return the generator.Route for multiple routes",
			args: args{from: []*ginRoute{
				{
					Route: generator.Route{
						Method: "GET",
					},
				},
				{
					Route: generator.Route{
						Method: "POST",
					},
				},
			}},
			want: []generator.Route{
				{
					Method: "GET",
				},
				{
					Method: "POST",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toGoSwagRoute(tt.args.from); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toGoSwagRoute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToGoSwagGroup(t *testing.T) {
	type args struct {
		from []*ginGroup
	}
	tests := []struct {
		name string
		args args
		want []generator.Group
	}{
		{
			name: "Should return the generator.Group",
			args: args{from: []*ginGroup{
				{
					groupName: "group1",
					routes: []*ginRoute{
						{
							Route: generator.Route{
								Method: "GET",
							},
						},
					},
				},
			}},
			want: []generator.Group{
				{
					GroupName: "group1",
					Routes: []generator.Route{
						{
							Method: "GET",
						},
					},
				},
			},
		},
		{
			name: "Should return the generator.Group for multiple groups",
			args: args{from: []*ginGroup{
				{
					groupName: "group1",
					routes: []*ginRoute{
						{
							Route: generator.Route{
								Method: "GET",
							},
						},
					},
				},
				{
					groupName: "group3",
					routes: []*ginRoute{
						{
							Route: generator.Route{
								Method: "PUT",
							},
						},
					},
				},
			}},
			want: []generator.Group{
				{
					GroupName: "group1",
					Routes: []generator.Route{
						{
							Method: "GET",
						},
					},
				},
				{
					GroupName: "group3",
					Routes: []generator.Route{
						{
							Method: "PUT",
						},
					},
					Groups: nil,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toGoSwagGroup(tt.args.from); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toGoSwagGroup() = %v, want %v", got, tt.want)
			}
		})
	}
}
