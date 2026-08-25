package generator

import (
	"io"
	"strings"
	"testing"

	"github.com/diegoclair/goswag/v2/internal/generator/testutil"
	viewmodelA "github.com/diegoclair/goswag/v2/internal/generator/testutil/apia/viewmodel"
	viewmodelB "github.com/diegoclair/goswag/v2/internal/generator/testutil/apib/viewmodel"
	"github.com/diegoclair/goswag/v2/models"
	"github.com/stretchr/testify/assert"
)

const (
	mangledPkgA = "github_com_diegoclair_goswag_v2_internal_generator_testutil_apia_viewmodel"
	mangledPkgB = "github_com_diegoclair_goswag_v2_internal_generator_testutil_apib_viewmodel"
)

func TestAnnotationTypeName(t *testing.T) {
	ambiguous := map[string]bool{"viewmodel.LoginResponse": true}

	tests := []struct {
		name      string
		body      any
		ambiguous map[string]bool
		want      string
	}{
		{
			name: "Should return the struct name and package name",
			body: models.ReturnType{},
			want: "models.ReturnType",
		},
		{
			name: "Should not return * if the struct is a pointer",
			body: &models.ReturnType{},
			want: "models.ReturnType",
		},
		{
			name: "Should keep the slice prefix",
			body: []models.ReturnType{},
			want: "[]models.ReturnType",
		},
		{
			name: "Should shorten the package of a generic argument",
			body: testutil.StructGeneric[testutil.TestGeneric]{},
			want: "testutil.StructGeneric[testutil.TestGeneric]",
		},
		{
			name: "Should shorten the package of a sliced generic argument",
			body: testutil.StructGeneric[[]testutil.TestGeneric]{},
			want: "testutil.StructGeneric[[]testutil.TestGeneric]",
		},
		{
			name: "Should keep a primitive generic argument as is",
			body: testutil.StructGeneric[int]{},
			want: "testutil.StructGeneric[int]",
		},
		{
			name:      "Should use the full package path when the name is ambiguous",
			body:      viewmodelA.LoginResponse{},
			ambiguous: ambiguous,
			want:      mangledPkgA + ".LoginResponse",
		},
		{
			name:      "Should use the full package path of the other package too",
			body:      viewmodelB.LoginResponse{},
			ambiguous: ambiguous,
			want:      mangledPkgB + ".LoginResponse",
		},
		{
			name:      "Should keep an unambiguous name short in the ambiguous package",
			body:      viewmodelA.OnlyInA{},
			ambiguous: ambiguous,
			want:      "viewmodel.OnlyInA",
		},
		{
			name:      "Should use the full package path inside a generic argument",
			body:      testutil.StructGeneric[[]viewmodelB.LoginResponse]{},
			ambiguous: ambiguous,
			want:      "testutil.StructGeneric[[]" + mangledPkgB + ".LoginResponse]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := annotationTypeName(tt.body, tt.ambiguous)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAmbiguousTypeNames(t *testing.T) {
	tests := []struct {
		name   string
		routes []Route
		groups []Group
		want   map[string]bool
	}{
		{
			name: "Should find no ambiguity when every name has a single package",
			routes: []Route{
				{Reads: viewmodelA.LoginResponse{}, Returns: []models.ReturnType{{StatusCode: 200, Body: viewmodelA.OnlyInA{}}}},
			},
			want: map[string]bool{},
		},
		{
			name: "Should find the name declared by two packages",
			groups: []Group{
				{Routes: []Route{{Returns: []models.ReturnType{{StatusCode: 200, Body: viewmodelA.LoginResponse{}}}}}},
				{Groups: []Group{{Routes: []Route{{Reads: viewmodelB.LoginResponse{}}}}}},
			},
			want: map[string]bool{"viewmodel.LoginResponse": true},
		},
		{
			name: "Should find the name when it only shows up as a generic argument",
			routes: []Route{
				{Returns: []models.ReturnType{
					{StatusCode: 200, Body: testutil.StructGeneric[[]viewmodelA.LoginResponse]{}},
					{StatusCode: 201, Body: testutil.StructGeneric[viewmodelB.LoginResponse]{}},
				}},
			},
			want: map[string]bool{"viewmodel.LoginResponse": true},
		},
		{
			name: "Should find the name declared in an overridden field",
			routes: []Route{
				{Returns: []models.ReturnType{
					{StatusCode: 200, Body: viewmodelA.LoginResponse{}},
					{StatusCode: 201, Body: testutil.OverrideStruct{}, OverrideStructFields: map[string]any{"body": viewmodelB.LoginResponse{}}},
				}},
			},
			want: map[string]bool{"viewmodel.LoginResponse": true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ambiguousTypeNames(tt.routes, tt.groups)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Two bounded contexts declaring the same response type is the case that used to make
// swag answer both routes with whichever package its import scan reached first.
func TestWriteGroup_sameTypeNameInTwoPackages(t *testing.T) {
	groups := []Group{
		{
			GroupName: "a",
			Routes: []Route{{
				Path:    "/a/login",
				Method:  "POST",
				Reads:   viewmodelA.LoginResponse{},
				Returns: []models.ReturnType{{StatusCode: 200, Body: viewmodelA.LoginResponse{}}},
			}},
		},
		{
			GroupName: "b",
			Routes: []Route{{
				Path:    "/b/login",
				Method:  "POST",
				Reads:   viewmodelB.LoginResponse{},
				Returns: []models.ReturnType{{StatusCode: 200, Body: viewmodelB.LoginResponse{}}},
			}},
		},
	}

	var b strings.Builder
	writeGroup(groups, &b, map[string]bool{}, ambiguousTypeNames(nil, groups))

	assert.Equal(t,
		"// @Tags a\n// @Accept json\n// @Produce json\n"+
			"// @Param request body "+mangledPkgA+".LoginResponse true \"Request\"\n"+
			"// @Success 200 {object} "+mangledPkgA+".LoginResponse\n"+
			"// @Router /a/login [post]\n\n"+
			"// @Tags b\n// @Accept json\n// @Produce json\n"+
			"// @Param request body "+mangledPkgB+".LoginResponse true \"Request\"\n"+
			"// @Success 200 {object} "+mangledPkgB+".LoginResponse\n"+
			"// @Router /b/login [post]\n\n",
		b.String())
}

func TestAddLineIfNotEmpty(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		format   string
		expected string
	}{
		{
			name:     "Should return empty string",
			input:    "",
			format:   "",
			expected: "",
		},
		{
			name:     "Should return empty string even if we have format",
			input:    "",
			format:   "test %s",
			expected: "",
		},
		{
			name:     "Should return the input string",
			input:    "test",
			format:   "some %s",
			expected: "some test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var b strings.Builder
			addLineIfNotEmpty(&b, tt.input, tt.format)
			result := b.String()

			if result != tt.expected {
				t.Errorf("Expected %s, but got %s", tt.expected, result)
			}
		})
	}
}

func TestAddTextIfNotEmptyOrDefault_slice(t *testing.T) {
	var tests = []struct {
		name        string
		input       []string
		defaultText string
		format      string
		expected    string
	}{
		{
			name:        "Should return default text",
			input:       []string{},
			defaultText: "default",
			format:      "some %s",
			expected:    "some default",
		},
		{
			name:        "Should return the input string",
			input:       []string{"test"},
			defaultText: "default",
			format:      "some %s",
			expected:    "some test",
		},
		{
			name:        "Should return the multiple input string separated by comma",
			input:       []string{"test", "test2"},
			defaultText: "default",
			format:      "some %s",
			expected:    "some test,test2",
		},
		{
			name:        "Should add nothing if input and default text are empty",
			input:       []string{},
			defaultText: "",
			format:      "some %s",
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var b strings.Builder
			addTextIfNotEmptyOrDefault(&b, tt.defaultText, tt.format, tt.input...)
			result := b.String()

			if result != tt.expected {
				t.Errorf("Expected %s, but got %s", tt.expected, result)
			}
		})
	}
}

func TestAddTextIfNotEmptyOrDefault_string(t *testing.T) {
	var tests = []struct {
		name        string
		input       string
		defaultText string
		format      string
		expected    string
	}{
		{
			name:        "Should return default text",
			input:       "",
			defaultText: "default",
			format:      "some %s",
			expected:    "some default",
		},
		{
			name:        "Should return the input string",
			input:       "test",
			defaultText: "default",
			format:      "some %s",
			expected:    "some test",
		},
		{
			name:        "Should add nothing if input and default text are empty",
			input:       "",
			defaultText: "",
			format:      "some %s",
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var b strings.Builder
			addTextIfNotEmptyOrDefault(&b, tt.defaultText, tt.format, tt.input)
			result := b.String()

			if result != tt.expected {
				t.Errorf("Expected %s, but got %s", tt.expected, result)
			}
		})
	}
}

func TestWriteGroup(t *testing.T) {
	var tests = []struct {
		name                  string
		groups                []Group
		expectedStringBuilder string
	}{
		{
			name: "Should return string with the group name",
			groups: []Group{
				{
					GroupName: "test",
					Routes: []Route{
						{
							Description: "test group",
							Path:        "/test",
							Method:      "GET",
						},
					},
				},
			},
			expectedStringBuilder: "// @Description test group\n// @Tags test\n// @Router /test [get]\n\n",
		},
		{
			name: "Should recursively return string with the group name",
			groups: []Group{
				{
					GroupName: "test",
					Routes: []Route{
						{
							Path:        "/test",
							Description: "test group",
						},
					},
					Groups: []Group{
						{
							GroupName: "test2",
							Routes: []Route{
								{
									Path:        "/test2",
									Description: "test group 2",
								},
							},
						},
					},
				},
			},
			expectedStringBuilder: "// @Description test group\n// @Tags test\n// @Router /test []\n\n// @Description test group 2\n// @Tags test2\n// @Router /test2 []\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var b strings.Builder
			writeGroup(tt.groups, &b, map[string]bool{}, nil)

			assert.Equal(t, tt.expectedStringBuilder, b.String())
		})
	}
}

func TestWriteRoutes(t *testing.T) {
	var tests = []struct {
		name                  string
		groupName             string
		routes                []Route
		expectedStringBuilder string
	}{
		{
			name:      "Should group name as tag of route",
			groupName: "test",
			routes: []Route{
				{},
			},
			expectedStringBuilder: "// @Tags test\n\n",
		},
		{
			name:      "Should add summary and description if we have summary",
			groupName: "",
			routes: []Route{
				{
					Summary: "test",
				},
			},
			expectedStringBuilder: "// @Summary test\n// @Description test\n\n",
		},
		{
			name:      "Should add description if we have description",
			groupName: "",
			routes: []Route{
				{
					Description: "test",
				},
			},
			expectedStringBuilder: "// @Description test\n\n",
		},
		{
			name:      "Should add tags if we have tags",
			groupName: "",
			routes: []Route{
				{
					Tags: []string{"test"},
				},
			},
			expectedStringBuilder: "// @Tags test\n\n",
		},
		{
			name:      "Should add tags, instead of group if we have tags",
			groupName: "group_test",
			routes: []Route{
				{
					Tags: []string{"tag_test"},
				},
			},
			expectedStringBuilder: "// @Tags tag_test\n\n",
		},
		{
			name:      "Should add default accept json if we have post method",
			groupName: "",
			routes: []Route{
				{
					Method: "POST",
				},
			},
			expectedStringBuilder: "// @Accept json\n\n",
		},
		{
			name:      "Should add accept text instead of default json",
			groupName: "",
			routes: []Route{
				{
					Method:  "POST",
					Accepts: []string{"text"},
				},
			},
			expectedStringBuilder: "// @Accept text\n\n",
		},
		{
			name:      "Should add produces if we have return",
			groupName: "",
			routes: []Route{
				{
					Returns: []models.ReturnType{
						{},
					},
				},
			},
			expectedStringBuilder: "// @Produce json\n\n",
		},
		{
			name:      "Should add request body if we have reads",
			groupName: "",
			routes: []Route{
				{
					Reads: models.ReturnType{},
				},
			},
			expectedStringBuilder: "// @Param request body models.ReturnType true \"Request\"\n\n",
		},
		{
			name:      "Should add path params if we have path params",
			groupName: "",
			routes: []Route{
				{
					PathParams: []Param{
						{
							Name:        "test",
							Description: "someTest",
							ParamType:   "string",
							Required:    true,
						},
					},
				},
			},
			expectedStringBuilder: "// @Param test path string true \"someTest\"\n\n",
		},
		{
			name:      "Should add query params if we have query params",
			groupName: "",
			routes: []Route{
				{
					QueryParams: []Param{
						{
							Name:        "test",
							Description: "test",
							ParamType:   "string",
							Required:    true,
						},
					},
				},
			},
			expectedStringBuilder: "// @Param test query string true \"test\"\n\n",
		},
		{
			name:      "Should add header params if we have header params",
			groupName: "",
			routes: []Route{
				{
					HeaderParams: []Param{
						{
							Name:        "test",
							Description: "test",
							ParamType:   "string",
							Required:    true,
						},
					},
				},
			},
			expectedStringBuilder: "// @Param test header string true \"test\"\n\n",
		},
		{
			name:      "Should add router if we have path",
			groupName: "",
			routes: []Route{
				{
					Path:   "/test",
					Method: "GET",
				},
			},
			expectedStringBuilder: "// @Router /test [get]\n\n",
		},
		{
			name:      "Should add func name if we have func name",
			groupName: "",
			routes: []Route{
				{
					FuncName: "test",
				},
			},
			expectedStringBuilder: "func test() {} //nolint:unused \n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var b strings.Builder
			writeRoutes(tt.groupName, tt.routes, &b, map[string]bool{}, nil)

			assert.Equal(t, tt.expectedStringBuilder, b.String())
		})
	}
}

func TestWriteReturns(t *testing.T) {
	var tests = []struct {
		name                  string
		returns               []models.ReturnType
		expectedStringBuilder string
		expectedPackages      map[string]bool
	}{
		{
			name: "Should return the struct name and package name as success 200",
			returns: []models.ReturnType{
				{
					StatusCode: 200,
					Body:       models.ReturnType{},
				},
			},
			expectedStringBuilder: "// @Success 200 {object} models.ReturnType\n",
			expectedPackages:      map[string]bool{"github.com/diegoclair/goswag/v2/models": true},
		},
		{
			name: "Should do nothing if we do not have status code",
			returns: []models.ReturnType{
				{
					Body: models.ReturnType{},
				},
			},
			expectedStringBuilder: "",
			expectedPackages:      map[string]bool{},
		},
		{
			name: "Should return the struct name and package name as failure 400",
			returns: []models.ReturnType{
				{
					StatusCode: 400,
					Body:       models.ReturnType{},
				},
			},
			expectedStringBuilder: "// @Failure 400 {object} models.ReturnType\n",
			expectedPackages:      map[string]bool{"github.com/diegoclair/goswag/v2/models": true},
		},
		{
			name: "Should add only status code if we do not have body",
			returns: []models.ReturnType{
				{
					StatusCode: 400,
				},
			},
			expectedStringBuilder: "// @Failure 400\n",
			expectedPackages:      map[string]bool{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var (
				b    strings.Builder
				pkgs = make(map[string]bool)
			)

			writeReturns(tt.returns, &b, pkgs, nil)

			assert.Equal(t, tt.expectedStringBuilder, b.String())
			assert.Equal(t, tt.expectedPackages, pkgs)
		})
	}
}

func Test_handleOverrideStructFields(t *testing.T) {
	var tests = []struct {
		name                  string
		data                  models.ReturnType
		expectedStringBuilder string
	}{
		{
			name:                  "Should do nothing if we do not have override struct fields",
			data:                  models.ReturnType{},
			expectedStringBuilder: "",
		},
		{
			name: "Should add override struct fields",
			data: models.ReturnType{
				Body: testutil.OverrideStruct{},
				OverrideStructFields: map[string]any{
					"test": testutil.TestGeneric{},
				},
			},
			expectedStringBuilder: "{test=testutil.TestGeneric}",
		},
		{
			name: "Should add multiple override struct fields",
			data: models.ReturnType{
				Body: testutil.OverrideStruct{},
				OverrideStructFields: map[string]any{
					"test":  testutil.TestGeneric{},
					"test2": testutil.TestGeneric{},
				},
			},
			expectedStringBuilder: "{test=testutil.TestGeneric,test2=testutil.TestGeneric}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var b strings.Builder
			handleOverrideStructFields(&b, tt.data, nil)

			assert.Equal(t, tt.expectedStringBuilder, b.String())
		})
	}
}

func Test_writeFileContent(t *testing.T) {
	type args struct {
		file             io.Writer
		content          string
		packagesToImport map[string]bool
	}
	tests := []struct {
		name     string
		args     args
		expected string
	}{
		{
			name: "Should write the file content",
			args: args{
				file:             &strings.Builder{},
				content:          "test",
				packagesToImport: map[string]bool{"test": true},
			},
			expected: "package main\n\nimport (\n\t_ \"test\"\n)\n\ntest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writeFileContent(tt.args.file, tt.args.content, tt.args.packagesToImport)
		})
	}
}

func Test_addDefaultResponses(t *testing.T) {
	type args struct {
		routes           []Route
		groups           []Group
		defaultResponses []models.ReturnType
	}
	tests := []struct {
		name     string
		args     args
		expected []Route
	}{
		{
			name: "Should do nothing if we do not have default responses",
			args: args{
				routes: []Route{
					{},
				},
				groups: []Group{
					{
						Routes: []Route{
							{},
						},
					},
				},
				defaultResponses: []models.ReturnType{},
			},
			expected: []Route{
				{},
			},
		},
		{
			name: "Should add default responses to routes",
			args: args{
				routes: []Route{
					{},
				},
				groups: []Group{
					{
						Routes: []Route{
							{},
						},
					},
				},
				defaultResponses: []models.ReturnType{
					{
						StatusCode: 200,
					},
				},
			},
			expected: []Route{
				{
					Returns: []models.ReturnType{
						{
							StatusCode: 200,
						},
					},
				},
			},
		},
		{
			name: "Should add default responses to groups",
			args: args{
				routes: []Route{
					{},
				},
				groups: []Group{
					{
						Routes: []Route{
							{},
						},
					},
				},
				defaultResponses: []models.ReturnType{
					{
						StatusCode: 204,
					},
				},
			},
			expected: []Route{
				{
					Returns: []models.ReturnType{
						{
							StatusCode: 204,
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRoutes, gotGroups := addDefaultResponses(tt.args.routes, tt.args.groups, tt.args.defaultResponses)
			assert.Equal(t, tt.expected, gotRoutes)
			assert.Equal(t, tt.expected, gotGroups[0].Routes)
		})
	}
}

func Test_addPackageToImport(t *testing.T) {
	tests := []struct {
		name         string
		data         models.ReturnType
		initialPkgs  map[string]bool
		expectedPkgs map[string]bool
	}{
		{
			name: "Should add package for non-generic type",
			data: models.ReturnType{
				Body: models.ReturnType{},
			},
			initialPkgs: make(map[string]bool),
			expectedPkgs: map[string]bool{
				"github.com/diegoclair/goswag/v2/models": true,
			},
		},
		{
			name: "Should add package for generic type",
			data: models.ReturnType{
				Body: testutil.StructGeneric[testutil.TestGeneric]{},
			},
			initialPkgs: make(map[string]bool),
			expectedPkgs: map[string]bool{
				"github.com/diegoclair/goswag/v2/internal/generator/testutil": true,
			},
		},
		{
			name: "Should not add package for primitive type",
			data: models.ReturnType{
				Body: 42,
			},
			initialPkgs:  make(map[string]bool),
			expectedPkgs: map[string]bool{},
		},
		{
			name: "Should not add package for nil body",
			data: models.ReturnType{
				Body: nil,
			},
			initialPkgs:  make(map[string]bool),
			expectedPkgs: map[string]bool{},
		},
		{
			name: "Should not duplicate existing package",
			data: models.ReturnType{
				Body: models.ReturnType{},
			},
			initialPkgs: map[string]bool{
				"github.com/diegoclair/goswag/v2/models": true,
			},
			expectedPkgs: map[string]bool{
				"github.com/diegoclair/goswag/v2/models": true,
			},
		},
		{
			name: "Should add package for pointer to struct",
			data: models.ReturnType{
				Body: &models.ReturnType{},
			},
			initialPkgs: make(map[string]bool),
			expectedPkgs: map[string]bool{
				"github.com/diegoclair/goswag/v2/models": true,
			},
		},
		{
			name: "Should add package for a plain slice",
			data: models.ReturnType{
				Body: []testutil.TestGeneric{},
			},
			initialPkgs: make(map[string]bool),
			expectedPkgs: map[string]bool{
				"github.com/diegoclair/goswag/v2/internal/generator/testutil": true,
			},
		},
		{
			name: "Should add both packages for a generic type argument from another package",
			data: models.ReturnType{
				Body: testutil.StructGeneric[models.ReturnType]{},
			},
			initialPkgs: make(map[string]bool),
			expectedPkgs: map[string]bool{
				"github.com/diegoclair/goswag/v2/internal/generator/testutil": true,
				"github.com/diegoclair/goswag/v2/models":                      true,
			},
		},
		{
			name: "Should add both packages for a generic wrapping a slice from another package",
			data: models.ReturnType{
				Body: testutil.StructGeneric[[]models.ReturnType]{},
			},
			initialPkgs: make(map[string]bool),
			expectedPkgs: map[string]bool{
				"github.com/diegoclair/goswag/v2/internal/generator/testutil": true,
				"github.com/diegoclair/goswag/v2/models":                      true,
			},
		},
		{
			name: "Should add package for a slice of pointers",
			data: models.ReturnType{
				Body: []*models.ReturnType{},
			},
			initialPkgs: make(map[string]bool),
			expectedPkgs: map[string]bool{
				"github.com/diegoclair/goswag/v2/models": true,
			},
		},
		{
			name: "Should add packages from both sides of a map",
			data: models.ReturnType{
				Body: map[testutil.TestGeneric]models.ReturnType{},
			},
			initialPkgs: make(map[string]bool),
			expectedPkgs: map[string]bool{
				"github.com/diegoclair/goswag/v2/internal/generator/testutil": true,
				"github.com/diegoclair/goswag/v2/models":                      true,
			},
		},
		{
			name: "Should add every package in a nested generic",
			data: models.ReturnType{
				Body: testutil.StructGeneric[testutil.StructGeneric[[]models.ReturnType]]{},
			},
			initialPkgs: make(map[string]bool),
			expectedPkgs: map[string]bool{
				"github.com/diegoclair/goswag/v2/internal/generator/testutil": true,
				"github.com/diegoclair/goswag/v2/models":                      true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packagesToImport := tt.initialPkgs
			addPackageToImport(tt.data, packagesToImport)
			assert.Equal(t, tt.expectedPkgs, packagesToImport)
		})
	}
}
