package echo

import (
	"testing"

	"github.com/diegoclair/goswag/internal/generator"
)

func Test_getFuncName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Should return the function name",
			args: args{name: "github.com/diegoclair/goswag/internal/frameworks/echo.(*Echo).GET-fm"},
			want: "GET",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getFuncName(tt.args.name); got != tt.want {
				t.Errorf("getFuncName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_toGoSwagRoute(t *testing.T) {
	type args struct {
		from []*echoRoute
	}
	tests := []struct {
		name string
		args args
		want []generator.Route
	}{
		{
			name: "Should return the generator.Route",
			args: args{from: []*echoRoute{
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
			args: args{from: []*echoRoute{
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
			if got := toGoSwagRoute(tt.args.from); len(got) != len(tt.want) {
				t.Errorf("toGoSwagRoute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_toGoSwagGroup(t *testing.T) {
	type args struct {
		from []*echoGroup
	}
	tests := []struct {
		name string
		args args
		want []generator.Group
	}{
		{
			name: "Should return the generator.Group",
			args: args{from: []*echoGroup{
				{
					groupName: "group",
					routes: []*echoRoute{
						{
							Route: generator.Route{
								Method: "GET",
							},
						},
					},
					groups: []*echoGroup{
						{
							groupName: "subgroup",
							routes: []*echoRoute{
								{
									Route: generator.Route{
										Method: "POST",
									},
								},
							},
						},
					},
				},
			}},
			want: []generator.Group{
				{
					GroupName: "group",
					Routes: []generator.Route{
						{
							Method: "GET",
						},
					},
					Groups: []generator.Group{
						{
							GroupName: "subgroup",
							Routes: []generator.Route{
								{
									Method: "POST",
								},
							},
							Groups: nil,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toGoSwagGroup(tt.args.from); len(got) != len(tt.want) {
				t.Errorf("toGoSwagGroup() = %v, want %v", got, tt.want)
			}
		})
	}
}
