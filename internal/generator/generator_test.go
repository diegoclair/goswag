package generator

import (
	"io"
	"strings"
	"testing"

	"github.com/diegoclair/goswag/internal/generator/testutil"
	"github.com/diegoclair/goswag/models"
	"github.com/stretchr/testify/assert"
)

func TestGetStructAndPackageName(t *testing.T) {
	type args struct {
		body interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Should return the struct name and package name",
			args: args{
				body: models.ReturnType{},
			},
			want: "models.ReturnType",
		},
		{
			name: "Should not return * if the struct is a pointer",
			args: args{
				body: &models.ReturnType{},
			},
			want: "models.ReturnType",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getStructAndPackageName(tt.args.body)
			assert.Equal(t, tt.want, got)
		})
	}
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
			writeGroup(tt.groups, &b)

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
					PathParams: []string{"test"},
				},
			},
			expectedStringBuilder: "// @Param test path string true \"test\" \n\n",
		},
		{
			name:      "Should add path params with camel case",
			groupName: "",
			routes: []Route{
				{
					PathParams: []string{"test_test"},
				},
			},
			expectedStringBuilder: "// @Param test_test path string true \"testTest\" \n\n",
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
			writeRoutes(tt.groupName, tt.routes, &b)

			assert.Equal(t, tt.expectedStringBuilder, b.String())
		})
	}
}

func TestWriteReturns(t *testing.T) {
	var tests = []struct {
		name                  string
		returns               []models.ReturnType
		expectedStringBuilder string
		expectedPackages      []string
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
		},
		{
			name: "Should do nothing if we do not have status code",
			returns: []models.ReturnType{
				{
					Body: models.ReturnType{},
				},
			},
			expectedStringBuilder: "",
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
		},
		{
			name: "Should add only status code if we do not have body",
			returns: []models.ReturnType{
				{
					StatusCode: 400,
				},
			},
			expectedStringBuilder: "// @Failure 400\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var b strings.Builder
			pkgs := writeReturns(tt.returns, &b)

			assert.Equal(t, tt.expectedStringBuilder, b.String())
			assert.Equal(t, tt.expectedPackages, pkgs)
		})
	}
}

func Test_writeIfIsGenericType(t *testing.T) {
	var tests = []struct {
		name                  string
		data                  models.ReturnType
		respType              string
		expectedIsGeneric     bool
		expectedStringBuilder string
		expectedPkg           []string
	}{
		{
			name: "Should return false if the body is not a generic type",
			data: models.ReturnType{
				Body: models.ReturnType{},
			},
			respType:              "@Success",
			expectedStringBuilder: "",
			expectedIsGeneric:     false,
		},
		{
			name: "Should return true if the body is a generic type",
			data: models.ReturnType{
				StatusCode: 200,
				Body:       testutil.StructGeneric[testutil.TestGeneric]{},
			},
			respType:              "@Success",
			expectedStringBuilder: "// @Success 200 {object} testutil.StructGeneric[testutil.TestGeneric]",
			expectedPkg:           []string{"github.com/diegoclair/goswag/internal/generator/testutil"},
			expectedIsGeneric:     true,
		},
		{
			name: "Should return true and correctly response when generic is type of slice",
			data: models.ReturnType{
				StatusCode: 200,
				Body:       testutil.StructGeneric[[]testutil.TestGeneric]{},
			},
			respType:              "@Success",
			expectedStringBuilder: "// @Success 200 {object} testutil.StructGeneric[[]testutil.TestGeneric]",
			expectedPkg:           []string{"github.com/diegoclair/goswag/internal/generator/testutil"},
			expectedIsGeneric:     true,
		},
		{
			name: "Should return true and no pkg when generic is a primitive type",
			data: models.ReturnType{
				StatusCode: 200,
				Body:       testutil.StructGeneric[int]{},
			},
			respType:              "@Success",
			expectedStringBuilder: "// @Success 200 {object} testutil.StructGeneric[int]",
			expectedIsGeneric:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var b strings.Builder
			result, pkgs := writeIfIsGenericType(&b, tt.data, tt.respType)

			assert.Equal(t, tt.expectedIsGeneric, result)
			assert.Equal(t, tt.expectedStringBuilder, b.String())
			assert.Equal(t, tt.expectedPkg, pkgs)
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
				OverrideStructFields: map[string]interface{}{
					"test": testutil.TestGeneric{},
				},
			},
			expectedStringBuilder: "{test=testutil.TestGeneric}",
		},
		{
			name: "Should add multiple override struct fields",
			data: models.ReturnType{
				Body: testutil.OverrideStruct{},
				OverrideStructFields: map[string]interface{}{
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
			handleOverrideStructFields(&b, tt.data)

			assert.Equal(t, tt.expectedStringBuilder, b.String())
		})
	}
}

func Test_writeFileContent(t *testing.T) {
	type args struct {
		file             io.Writer
		content          string
		packagesToImport []string
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
				packagesToImport: []string{"test"},
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
