package generator

import (
	"strings"
	"testing"

	"github.com/diegoclair/goswag/models"
	"github.com/stretchr/testify/assert"
)

func TestGetStructAndPackageName(t *testing.T) {
	body := models.ReturnType{}
	expected := "models.ReturnType"

	result := getStructAndPackageName(body)

	if result != expected {
		t.Errorf("Expected %s, but got %s", expected, result)
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
			expectedStringBuilder: "// @Description test group\n// @Tags test\n// @Accept json\n// @Produce json\n// @Router /test [get]\nfunc () {}\n\n",
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
			expectedStringBuilder: "// @Description test group\n// @Tags test\n// @Accept json\n// @Produce json\n// @Router /test []\nfunc () {}\n\n// @Description test group 2\n// @Tags test2\n// @Accept json\n// @Produce json\n// @Router /test2 []\nfunc () {}\n\n",
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
