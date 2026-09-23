package dedupe

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// minUses is where a reusable object starts paying for itself: at two uses the
// hoisted copy costs about what the two inline copies did.
const minUses = 3

var httpMethods = map[string]bool{
	"get": true, "put": true, "post": true, "delete": true,
	"options": true, "head": true, "patch": true,
}

// Stats reports what a rewrite factored out, so the caller can tell the user
// whether the run changed anything.
type Stats struct {
	Responses  int
	Parameters int
}

func (s Stats) Empty() bool { return s.Responses == 0 && s.Parameters == 0 }

// Document keeps the API it was given and only changes where it is written down,
// which is what makes a second run over the result a no-op.
func Document(raw []byte) ([]byte, Stats, error) {
	root, err := parseJSON(raw)
	if err != nil {
		return nil, Stats{}, err
	}

	paths := root.get("paths")
	if paths == nil || paths.kind != kindObject {
		return raw, Stats{}, nil
	}

	responses := collect(paths, collectResponses)
	parameters := collect(paths, collectParameters)

	nameCandidates(responses, existingNames(root, "responses"))
	nameCandidates(parameters, existingNames(root, "parameters"))

	applyRefs(paths, responses, parameters)

	stats := Stats{Responses: len(responses), Parameters: len(parameters)}
	hoist(root, "responses", responses)
	hoist(root, "parameters", parameters)

	return root.marshalIndent(), stats, nil
}

type candidate struct {
	signature string
	hint      string
	// qualifier separates candidates that share a hint by what they carry,
	// so a reader meets OKProduct rather than OK4.
	qualifier string
	value     *node
	uses      int
	name      string
}

type collector func(operation *node, visit func(hint, qualifier string, value *node))

func collectResponses(operation *node, visit func(string, string, *node)) {
	responses := operation.get("responses")
	if responses == nil || responses.kind != kindObject {
		return
	}
	for _, m := range responses.members {
		visit(responseHint(m.key, m.value), schemaName(m.value.get("schema")), m.value)
	}
}

func collectParameters(operation *node, visit func(string, string, *node)) {
	parameters := operation.get("parameters")
	if parameters == nil || parameters.kind != kindArray {
		return
	}
	for _, item := range parameters.items {
		// A parameter is already identified by its name and location; nothing
		// else about it would tell two same-named ones apart any better.
		visit(parameterHint(item), "", item)
	}
}

func collect(paths *node, collect collector) map[string]*candidate {
	found := map[string]*candidate{}

	eachOperation(paths, func(operation *node) {
		collect(operation, func(hint, qualifier string, value *node) {
			if value.isRef() {
				return
			}
			signature := value.signature()
			if existing, ok := found[signature]; ok {
				existing.uses++
				return
			}
			found[signature] = &candidate{
				signature: signature, hint: hint, qualifier: qualifier,
				value: value.clone(), uses: 1,
			}
		})
	})

	for signature, c := range found {
		if c.uses < minUses {
			delete(found, signature)
		}
	}
	return found
}

func eachOperation(paths *node, visit func(*node)) {
	for _, path := range paths.members {
		if path.value.kind != kindObject {
			continue
		}
		for _, operation := range path.value.members {
			if httpMethods[strings.ToLower(operation.key)] && operation.value.kind == kindObject {
				visit(operation.value)
			}
		}
	}
}

func applyRefs(paths *node, responses, parameters map[string]*candidate) {
	eachOperation(paths, func(operation *node) {
		if node := operation.get("responses"); node != nil && node.kind == kindObject {
			for i, m := range node.members {
				if c, ok := responses[m.value.signature()]; ok {
					node.members[i].value = refNode("#/responses/" + c.name)
				}
			}
		}
		if node := operation.get("parameters"); node != nil && node.kind == kindArray {
			for i, item := range node.items {
				if c, ok := parameters[item.signature()]; ok {
					node.items[i] = refNode("#/parameters/" + c.name)
				}
			}
		}
	})
}

func hoist(root *node, key string, candidates map[string]*candidate) {
	if len(candidates) == 0 {
		return
	}

	target := root.get(key)
	if target == nil || target.kind != kindObject {
		target = objectNode()
		root.insertBefore("paths", key, target)
	}

	for _, c := range sorted(candidates) {
		target.set(c.name, c.value)
	}
	sort.Slice(target.members, func(i, j int) bool { return target.members[i].key < target.members[j].key })
}

func existingNames(root *node, key string) map[string]bool {
	taken := map[string]bool{}
	if node := root.get(key); node != nil && node.kind == kindObject {
		for _, m := range node.members {
			taken[m.key] = true
		}
	}
	return taken
}

// nameCandidates assigns names in a fixed order so the same document always
// produces the same spec, whatever order the map handed them over in.
func nameCandidates(candidates map[string]*candidate, taken map[string]bool) {
	ordered := sorted(candidates)

	shared := map[string]int{}
	for _, c := range ordered {
		shared[pascalCase(c.hint)]++
	}

	for _, c := range ordered {
		base := pascalCase(c.hint)
		if shared[base] > 1 {
			base += pascalCase(c.qualifier)
		}
		if base == "" {
			base = "Shared"
		}

		name := base
		for i := 2; taken[name]; i++ {
			name = base + strconv.Itoa(i)
		}
		taken[name] = true
		c.name = name
	}
}

func sorted(candidates map[string]*candidate) []*candidate {
	out := make([]*candidate, 0, len(candidates))
	for _, c := range candidates {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].hint != out[j].hint {
			return out[i].hint < out[j].hint
		}
		return out[i].signature < out[j].signature
	})
	return out
}

func responseHint(status string, value *node) string {
	if description := stringValue(value.get("description")); description != "" {
		return description
	}
	return "status " + status
}

func parameterHint(value *node) string {
	name := stringValue(value.get("name"))
	if name == "" {
		return ""
	}
	return name + " " + stringValue(value.get("in"))
}

// schemaName names a response by what it returns, reaching through an array to
// the element so a list and a single item stay apart.
func schemaName(schema *node) string {
	if schema == nil || schema.kind != kindObject {
		return ""
	}
	if ref := stringValue(schema.get("$ref")); ref != "" {
		return shortTypeName(ref)
	}
	if items := schema.get("items"); items != nil {
		if ref := stringValue(items.get("$ref")); ref != "" {
			return shortTypeName(ref) + " list"
		}
	}
	return stringValue(schema.get("type"))
}

func shortTypeName(ref string) string {
	if i := strings.LastIndexAny(ref, "/"); i >= 0 {
		ref = ref[i+1:]
	}
	if i := strings.LastIndex(ref, "."); i >= 0 {
		ref = ref[i+1:]
	}
	return ref
}

func stringValue(n *node) string {
	if n == nil || n.kind != kindScalar {
		return ""
	}
	raw := string(n.raw)
	if len(raw) < 2 || raw[0] != '"' {
		return ""
	}
	var out string
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return ""
	}
	return out
}

func pascalCase(text string) string {
	var out strings.Builder
	upper := true
	for _, r := range text {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			if upper {
				out.WriteRune(unicode.ToUpper(r))
				upper = false
				continue
			}
			out.WriteRune(r)
		default:
			upper = true
		}
	}
	return out.String()
}
