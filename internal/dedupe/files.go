package dedupe

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sigs.k8s.io/yaml"
)

const jsonFile = "swagger.json"

var (
	yamlFiles = []string{"swagger.yaml", "swagger.yml"}
	docsFile  = "docs.go"
)

// ErrNoJSON says the step has nothing to read: the JSON document is the one swag
// writes in the shape this rewrite walks.
var ErrNoJSON = errors.New("swagger.json not found")

// Files re-derives the yaml and the embedded document from the rewritten JSON,
// so the three outputs never disagree about the same API.
func Files(dir string) (Stats, error) {
	raw, err := os.ReadFile(filepath.Join(dir, jsonFile))
	if err != nil {
		return Stats{}, fmt.Errorf("%w in %s", ErrNoJSON, dir)
	}

	rewritten, stats, err := Document(raw)
	if err != nil {
		return Stats{}, fmt.Errorf("rewriting %s: %w", jsonFile, err)
	}
	if stats.Empty() {
		return stats, nil
	}

	if err := os.WriteFile(filepath.Join(dir, jsonFile), rewritten, 0o600); err != nil {
		return stats, err
	}
	if err := writeYAML(dir, rewritten); err != nil {
		return stats, err
	}
	return stats, rewriteDocsGo(filepath.Join(dir, docsFile))
}

func writeYAML(dir string, document []byte) error {
	converted, err := yaml.JSONToYAML(document)
	if err != nil {
		return fmt.Errorf("converting the rewritten document to yaml: %w", err)
	}

	for _, name := range yamlFiles {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err != nil {
			continue
		}
		if err := os.WriteFile(path, converted, 0o600); err != nil {
			return err
		}
	}
	return nil
}

func rewriteDocsGo(path string) error {
	source, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	template, start, end, err := docTemplateBounds(string(source))
	if err != nil {
		return fmt.Errorf("reading the embedded document in %s: %w", docsFile, err)
	}

	masked, actions := maskActions(template)
	rewritten, _, err := Document([]byte(masked))
	if err != nil {
		return fmt.Errorf("rewriting the embedded document in %s: %w", docsFile, err)
	}

	restored := string(rewritten)
	for placeholder, action := range actions {
		restored = strings.ReplaceAll(restored, `"`+placeholder+`"`, action)
	}

	return os.WriteFile(path, []byte(string(source[:start])+restored+string(source[end:])), 0o600)
}

func docTemplateBounds(source string) (template string, start, end int, err error) {
	const marker = "docTemplate = `"

	i := strings.Index(source, marker)
	if i < 0 {
		return "", 0, 0, errors.New("no docTemplate literal")
	}

	start = i + len(marker)
	// The literal is a raw string, so it cannot itself contain a backtick.
	length := strings.IndexByte(source[start:], '`')
	if length < 0 {
		return "", 0, 0, errors.New("unterminated docTemplate literal")
	}
	return source[start : start+length], start, start + length, nil
}

// maskActions quotes the Go template actions that stand where a JSON value
// belongs, which is the only reason the embedded document does not parse.
func maskActions(template string) (string, map[string]string) {
	var out strings.Builder
	actions := map[string]string{}

	inString, escaped := false, false
	for i := 0; i < len(template); i++ {
		c := template[i]

		if inString {
			out.WriteByte(c)
			switch {
			case escaped:
				escaped = false
			case c == '\\':
				escaped = true
			case c == '"':
				inString = false
			}
			continue
		}

		if c == '"' {
			inString = true
			out.WriteByte(c)
			continue
		}

		if c == '{' && i+1 < len(template) && template[i+1] == '{' {
			closing := strings.Index(template[i:], "}}")
			if closing < 0 {
				out.WriteByte(c)
				continue
			}
			placeholder := fmt.Sprintf("__goswag_action_%d__", len(actions))
			actions[placeholder] = template[i : i+closing+2]
			out.WriteByte('"')
			out.WriteString(placeholder)
			out.WriteByte('"')
			i += closing + 1
			continue
		}

		out.WriteByte(c)
	}
	return out.String(), actions
}
