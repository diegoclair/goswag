package generator

import "strings"

// Echo and Gin both spell a path parameter ":name" and a catch-all "*name";
// OpenAPI knows neither, so a route written for the router has to be translated
// before it reaches an annotation.
const (
	paramLabel = ':'
	anyLabel   = '*'

	// The root package's StringType, repeated because importing it here would
	// close a cycle back through the framework adapters.
	stringParamType = "string"
)

// openAPIPath rewrites a router path into the only shape a spec consumer
// understands. A path already written that way is returned untouched.
func openAPIPath(path string) string {
	var out strings.Builder

	for i := 0; i < len(path); i++ {
		name, next, ok := paramAt(path, i)
		if !ok {
			out.WriteByte(path[i])
			continue
		}

		out.WriteByte('{')
		out.WriteString(name)
		out.WriteByte('}')
		i = next - 1
	}

	return out.String()
}

// pathParamsOf lists the parameters the route itself declares, in the order the
// path spells them.
func pathParamsOf(path string) []string {
	var names []string

	for i := 0; i < len(path); i++ {
		name, next, ok := paramAt(path, i)
		if !ok {
			continue
		}

		names = append(names, name)
		i = next - 1
	}

	return names
}

// paramAt reads a parameter that starts a segment. Anchoring on the segment
// keeps a colon that is part of a literal from being read as one.
func paramAt(path string, i int) (name string, next int, ok bool) {
	if path[i] != paramLabel && path[i] != anyLabel {
		return "", 0, false
	}
	if i > 0 && path[i-1] != '/' {
		return "", 0, false
	}

	end := i + 1
	for end < len(path) && isParamNameByte(path[end]) {
		end++
	}

	// Echo's catch-all may carry no name at all, and OpenAPI has nothing to say
	// about it, so it is left as the router wrote it.
	if end == i+1 {
		return "", 0, false
	}

	return path[i+1 : end], end, true
}

func isParamNameByte(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_'
}

// withRoutePathParams adds the parameters the path declares and the caller did
// not: OpenAPI rejects a path whose parameters are not all described, and a path
// parameter can only ever be a required string. The name is all goswag honestly
// knows to say about one nobody described, and swag refuses an empty comment.
func withRoutePathParams(path string, declared []Param) []Param {
	if path == "" {
		return declared
	}

	known := make(map[string]bool, len(declared))
	for _, param := range declared {
		known[param.Name] = true
	}

	out := declared
	for _, name := range pathParamsOf(path) {
		if known[name] {
			continue
		}
		known[name] = true
		out = append(out, Param{Name: name, Description: name, ParamType: stringParamType, Required: true})
	}

	return out
}
