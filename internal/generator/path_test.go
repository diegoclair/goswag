package generator

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpenAPIPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		want string
	}{
		{"echo and gin spell a parameter with a colon", "/users/:id", "/users/{id}"},
		{"every segment of a nested path", "/a/:x/b/:y/c", "/a/{x}/b/{y}/c"},
		{"a trailing slash after the parameter", "/products/:uuid/", "/products/{uuid}/"},
		{"an underscore in the name", "/docs/:document_id/attempts", "/docs/{document_id}/attempts"},
		{"a digit in the name", "/v/:id2", "/v/{id2}"},
		{"gin's named catch-all", "/static/*filepath", "/static/{filepath}"},
		{"a path that is only a parameter", "/:id", "/{id}"},
		{"a path with nothing to translate", "/health", "/health"},
		{"a path already written for a spec", "/users/{id}", "/users/{id}"},
		{"an empty path", "", ""},
		{"echo's unnamed catch-all, which a spec cannot express", "/files/*", "/files/*"},
		{"a colon that is not starting a segment", "/ratio/1:2", "/ratio/1:2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, openAPIPath(tt.path))
		})
	}
}

func TestOpenAPIPathIsIdempotent(t *testing.T) {
	t.Parallel()

	for _, path := range []string{"/users/:id", "/a/:x/b/:y", "/static/*filepath", "/plain"} {
		once := openAPIPath(path)
		assert.Equal(t, once, openAPIPath(once), "translating %q twice must not change it again", path)
	}
}

func TestPathParamsOf(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		want []string
	}{
		{"in the order the path spells them", "/a/:x/b/:y", []string{"x", "y"}},
		{"a named catch-all counts", "/static/*filepath", []string{"filepath"}},
		{"nothing to find", "/health", nil},
		{"an unnamed catch-all names nothing", "/files/*", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, pathParamsOf(tt.path))
		})
	}
}

func TestWithRoutePathParams(t *testing.T) {
	t.Parallel()

	t.Run("describes what the route declares and the caller did not", func(t *testing.T) {
		t.Parallel()

		got := withRoutePathParams("/users/:id", nil)
		assert.Equal(t, []Param{{Name: "id", Description: "id", ParamType: "string", Required: true}}, got,
			"a spec is rejected outright when a path parameter goes undescribed")
	})

	t.Run("keeps what the caller said about it", func(t *testing.T) {
		t.Parallel()

		declared := []Param{{Name: "id", Description: "The user's id", ParamType: "string", Required: true}}

		assert.Equal(t, declared, withRoutePathParams("/users/:id", declared))
	})

	t.Run("fills only the gaps", func(t *testing.T) {
		t.Parallel()

		declared := []Param{{Name: "id", Description: "The user's id", ParamType: "string", Required: true}}

		got := withRoutePathParams("/users/:id/posts/:post_id", declared)
		assert.Equal(t, []Param{
			{Name: "id", Description: "The user's id", ParamType: "string", Required: true},
			{Name: "post_id", Description: "post_id", ParamType: "string", Required: true},
		}, got)
	})

	t.Run("leaves a path without parameters alone", func(t *testing.T) {
		t.Parallel()

		assert.Nil(t, withRoutePathParams("/health", nil))
	})

	t.Run("leaves an empty path alone", func(t *testing.T) {
		t.Parallel()

		assert.Nil(t, withRoutePathParams("", nil))
	})
}

func TestWriteRoutesDescribesAParameterisedRoute(t *testing.T) {
	t.Parallel()

	var got strings.Builder
	writeRoutes("", []Route{{
		Path:   "/users/:id/posts/:post_id",
		Method: "GET",
		PathParams: []Param{
			{Name: "id", Description: "The user's id", ParamType: "string", Required: true},
		},
	}}, &got, map[string]bool{}, map[string]bool{})

	annotations := got.String()
	assert.Contains(t, annotations, "// @Router /users/{id}/posts/{post_id} [get]",
		"the router path reaches the annotation in the shape a spec consumer reads")
	assert.Contains(t, annotations, `// @Param id path string true "The user's id"`)
	assert.Contains(t, annotations, `// @Param post_id path string true "post_id"`,
		"the parameter the route declares but the caller never described still has to be there, and swag refuses an empty comment")
}
