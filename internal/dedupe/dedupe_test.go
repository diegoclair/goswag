package dedupe

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"
)

func fixture(t *testing.T, name string) []byte {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join("testdata", name))
	require.NoError(t, err)

	return raw
}

// decode walks the rewritten document the way a consumer would, rather than
// through the ordered model the rewrite itself uses.
func decode(t *testing.T, document []byte) map[string]any {
	t.Helper()

	var out map[string]any
	require.NoError(t, json.Unmarshal(document, &out))

	return out
}

func operation(t *testing.T, document map[string]any, path, method string) map[string]any {
	t.Helper()

	paths, ok := document["paths"].(map[string]any)
	require.True(t, ok, "document has no paths")

	item, ok := paths[path].(map[string]any)
	require.True(t, ok, "no path %q", path)

	op, ok := item[method].(map[string]any)
	require.True(t, ok, "no %s on %q", method, path)

	return op
}

func TestRoundTripKeepsTheDocumentByteForByte(t *testing.T) {
	t.Parallel()

	raw := fixture(t, "swagger.json")

	node, err := parseJSON(raw)
	require.NoError(t, err)

	assert.Equal(t, string(raw), string(node.marshalIndent()),
		"re-rendering an untouched document must not move a single byte, or every rewrite buries its own diff")
}

func TestDocumentFactorsTheRepeatedErrorResponses(t *testing.T) {
	t.Parallel()

	out, stats, err := Document(fixture(t, "swagger.json"))
	require.NoError(t, err)

	document := decode(t, out)
	responses, ok := document["responses"].(map[string]any)
	require.True(t, ok, "the rewrite added no reusable responses")

	for _, name := range []string{"BadRequest", "Unauthorized", "InternalServerError"} {
		assert.Contains(t, responses, name)
	}
	assert.Equal(t, len(responses), stats.Responses)

	op := operation(t, document, "/products/{uuid}", "get")
	got := op["responses"].(map[string]any)["400"]
	assert.Equal(t, map[string]any{"$ref": "#/responses/BadRequest"}, got,
		"an operation should point at the shared response instead of restating it")
}

func TestDocumentFactorsTheRepeatedParameters(t *testing.T) {
	t.Parallel()

	out, stats, err := Document(fixture(t, "swagger.json"))
	require.NoError(t, err)

	document := decode(t, out)
	parameters, ok := document["parameters"].(map[string]any)
	require.True(t, ok, "the rewrite added no reusable parameters")

	assert.Contains(t, parameters, "UserTokenHeader")
	assert.Equal(t, len(parameters), stats.Parameters)

	op := operation(t, document, "/products/", "get")
	assert.Contains(t, op["parameters"], map[string]any{"$ref": "#/parameters/UserTokenHeader"})
}

func TestDocumentLeavesRarelyRepeatedThingsInline(t *testing.T) {
	t.Parallel()

	out, _, err := Document(fixture(t, "swagger.json"))
	require.NoError(t, err)

	document := decode(t, out)
	if parameters, ok := document["parameters"].(map[string]any); ok {
		assert.NotContains(t, parameters, "PageQuery",
			"a parameter used once pays for a reusable object it never amortises")
	}

	op := operation(t, document, "/products/", "get")
	assert.Contains(t, op["parameters"], map[string]any{
		"type": "number", "description": "Page, 1-based", "name": "page", "in": "query",
	})
}

// TestDocumentWaitsForTheThirdUse pins the point where a reusable object starts
// paying for itself; the fixture alone cannot, having nothing used exactly twice.
func TestDocumentWaitsForTheThirdUse(t *testing.T) {
	t.Parallel()

	operations := func(count int) string {
		var paths []string
		for i := range count {
			paths = append(paths, `"/r`+string(rune('a'+i))+`/":{"get":{`+
				`"parameters":[{"type":"string","name":"trace","in":"header"}],`+
				`"responses":{"418":{"description":"Teapot"}}}}`)
		}
		return `{"swagger":"2.0","paths":{` + strings.Join(paths, ",") + `}}`
	}

	_, twice, err := Document([]byte(operations(2)))
	require.NoError(t, err)
	assert.True(t, twice.Empty(), "two copies cost about what the shared object would: %+v", twice)

	_, thrice, err := Document([]byte(operations(3)))
	require.NoError(t, err)
	assert.Equal(t, 1, thrice.Responses)
	assert.Equal(t, 1, thrice.Parameters)
}

func TestDocumentNamesCollidingResponsesByWhatTheyReturn(t *testing.T) {
	t.Parallel()

	out, _, err := Document(fixture(t, "swagger.json"))
	require.NoError(t, err)

	responses := decode(t, out)["responses"].(map[string]any)
	assert.Contains(t, responses, "OKProduct")
	assert.Contains(t, responses, "OKListing")

	for name := range responses {
		assert.NotEqual(t, "OK2", name, "a counter tells a reader nothing about what the response carries")
	}
}

func TestDocumentIsIdempotent(t *testing.T) {
	t.Parallel()

	once, first, err := Document(fixture(t, "swagger.json"))
	require.NoError(t, err)
	require.False(t, first.Empty())

	twice, second, err := Document(once)
	require.NoError(t, err)

	assert.Equal(t, string(once), string(twice), "a second run must not touch an already rewritten document")
	assert.True(t, second.Empty(), "a second run found candidates again: %+v", second)
}

func TestDocumentIsDeterministic(t *testing.T) {
	t.Parallel()

	raw := fixture(t, "swagger.json")

	first, _, err := Document(raw)
	require.NoError(t, err)

	for range 20 {
		again, _, err := Document(raw)
		require.NoError(t, err)
		require.Equal(t, string(first), string(again),
			"map iteration order leaked into the names, so two runs of the same spec would disagree")
	}
}

func TestDocumentKeepsEveryOperationIntact(t *testing.T) {
	t.Parallel()

	raw := fixture(t, "swagger.json")

	out, _, err := Document(raw)
	require.NoError(t, err)

	before, after := decode(t, raw), decode(t, out)
	resolved := resolveRefs(after, after)

	assert.Equal(t, before["paths"], resolved.(map[string]any)["paths"],
		"with the refs resolved the rewrite must describe exactly the API it was given")
}

// resolveRefs inlines what the rewrite lifted out, which is the only way to
// compare the two documents as the same API rather than as the same bytes.
func resolveRefs(value, root any) any {
	switch typed := value.(type) {
	case map[string]any:
		if ref, ok := typed["$ref"].(string); ok && len(typed) == 1 {
			if target, found := lookup(root, ref); found {
				return resolveRefs(target, root)
			}
		}

		out := make(map[string]any, len(typed))
		for key, item := range typed {
			out[key] = resolveRefs(item, root)
		}
		return out

	case []any:
		out := make([]any, len(typed))
		for i, item := range typed {
			out[i] = resolveRefs(item, root)
		}
		return out

	default:
		return value
	}
}

func lookup(root any, ref string) (any, bool) {
	// Only the containers this rewrite creates are inlined; definitions stay as
	// they were written, on both sides of the comparison.
	segments := strings.Split(strings.TrimPrefix(ref, "#/"), "/")
	if len(segments) != 2 || (segments[0] != "responses" && segments[0] != "parameters") {
		return nil, false
	}

	container, ok := root.(map[string]any)[segments[0]].(map[string]any)
	if !ok {
		return nil, false
	}

	target, ok := container[segments[1]]
	return target, ok
}

func TestDocumentPutsTheSharedObjectsAboveThePaths(t *testing.T) {
	t.Parallel()

	out, _, err := Document(fixture(t, "swagger.json"))
	require.NoError(t, err)

	root, err := parseJSON(out)
	require.NoError(t, err)

	var order []string
	for _, m := range root.members {
		order = append(order, m.key)
	}

	assert.Less(t, indexOf(order, "responses"), indexOf(order, "paths"),
		"the shared vocabulary belongs where a reader meets it first")
	assert.Less(t, indexOf(order, "parameters"), indexOf(order, "paths"))
}

func indexOf(values []string, want string) int {
	for i, v := range values {
		if v == want {
			return i
		}
	}
	return -1
}

func TestDocumentWithoutPathsIsLeftAlone(t *testing.T) {
	t.Parallel()

	raw := []byte(`{"swagger":"2.0"}`)

	out, stats, err := Document(raw)
	require.NoError(t, err)

	assert.Equal(t, string(raw), string(out))
	assert.True(t, stats.Empty())
}

func TestDocumentRejectsSomethingThatIsNotJSON(t *testing.T) {
	t.Parallel()

	_, _, err := Document([]byte("not a spec"))
	assert.Error(t, err)
}

func TestFilesRewritesEveryOutputSwagWrote(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	for _, name := range []string{"swagger.json", "swagger.yaml", "docs.go"} {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), fixture(t, name), 0o600))
	}

	stats, err := Files(dir)
	require.NoError(t, err)
	require.False(t, stats.Empty())

	rewritten, err := os.ReadFile(filepath.Join(dir, "swagger.json"))
	require.NoError(t, err)
	assert.Contains(t, string(rewritten), `"#/responses/BadRequest"`)

	t.Run("the yaml says the same thing as the json", func(t *testing.T) {
		gotYAML, err := os.ReadFile(filepath.Join(dir, "swagger.yaml"))
		require.NoError(t, err)

		wantYAML, err := yaml.JSONToYAML(rewritten)
		require.NoError(t, err)

		assert.Equal(t, string(wantYAML), string(gotYAML))
	})

	t.Run("the embedded document keeps its template actions", func(t *testing.T) {
		gotGo, err := os.ReadFile(filepath.Join(dir, "docs.go"))
		require.NoError(t, err)

		source := string(gotGo)
		assert.Contains(t, source, "{{ marshal .Schemes }}")
		assert.Contains(t, source, `"title": "{{.Title}}"`)
		assert.Contains(t, source, `"#/responses/BadRequest"`)
		assert.NotContains(t, source, "__goswag_action_")
	})
}

func TestFilesSaysWhatIsMissingWhenThereIsNoJSON(t *testing.T) {
	t.Parallel()

	_, err := Files(t.TempDir())
	assert.ErrorIs(t, err, ErrNoJSON)
}

func TestFilesWithoutTheOtherOutputsStillRewritesTheJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "swagger.json"), fixture(t, "swagger.json"), 0o600))

	stats, err := Files(dir)
	require.NoError(t, err)
	assert.False(t, stats.Empty())
}

func TestMaskActionsOnlyTouchesWhatBreaksTheParse(t *testing.T) {
	t.Parallel()

	template := `{"schemes": {{ marshal .Schemes }}, "title": "{{.Title}}", "note": "a } brace"}`

	masked, actions := maskActions(template)
	require.Len(t, actions, 1, "an action inside a string is already valid where it stands")

	var probe map[string]any
	require.NoError(t, json.Unmarshal([]byte(masked), &probe),
		"masking exists so the embedded document parses: %s", masked)

	assert.Equal(t, "{{.Title}}", probe["title"], "an action inside a string must survive untouched")

	restored := masked
	for placeholder, action := range actions {
		restored = strings.ReplaceAll(restored, `"`+placeholder+`"`, action)
	}
	assert.Equal(t, template, restored)
}

func TestPascalCase(t *testing.T) {
	t.Parallel()

	for input, want := range map[string]string{
		"Bad Request":           "BadRequest",
		"Internal Server Error": "InternalServerError",
		"user-token header":     "UserTokenHeader",
		"listing_uuid path":     "ListingUuidPath",
		"":                      "",
		"---":                   "",
	} {
		assert.Equal(t, want, pascalCase(input), "pascalCase(%q)", input)
	}
}

func TestSchemaNameReachesThroughAnArray(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		schema string
		want   string
	}{
		{"a referenced type", `{"$ref":"#/definitions/pkg.viewmodel.Product"}`, "Product"},
		{"a list of them", `{"type":"array","items":{"$ref":"#/definitions/pkg.Listing"}}`, "Listing list"},
		{"a plain type", `{"type":"string"}`, "string"},
		{"nothing at all", `{}`, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			node, err := parseJSON([]byte(tc.schema))
			require.NoError(t, err)

			assert.Equal(t, tc.want, schemaName(node))
		})
	}
}
