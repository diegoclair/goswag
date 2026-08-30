package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsStdlibImport(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		// stdlib
		{"strings", true},
		{"net/http", true},
		{"go/parser", true},
		{"encoding/json", true},
		// third-party (first segment contains ".")
		{"github.com/foo/bar", false},
		{"golang.org/x/mod/modfile", false},
		{"gopkg.in/yaml.v3", false},
		{"example.com/internal/thing", false},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := isStdlibImport(tt.path); got != tt.want {
				t.Errorf("isStdlibImport(%q) = %v; want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestNormalizeOutputTypes(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    string
		wantErr bool
	}{
		{name: "Default value is accepted unchanged", value: "go,json,yaml", want: "go,json,yaml"},
		{name: "Spec-only selection drops the go file", value: "json,yaml", want: "json,yaml"},
		{name: "Single type", value: "json", want: "json"},
		{name: "yml alias is supported by swag", value: "go,yml", want: "go,yml"},
		{name: "Surrounding spaces are tolerated", value: " json , yaml ", want: "json,yaml"},
		{name: "Case is normalised", value: "JSON,YAML", want: "json,yaml"},
		{name: "Typo is refused instead of reaching swag", value: "json,yamll", wantErr: true},
		{name: "Unknown type is refused", value: "xml", wantErr: true},
		{name: "Empty value is refused", value: "", wantErr: true},
		{name: "Empty element is refused", value: "json,,yaml", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeOutputTypes(tt.value)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("normalizeOutputTypes(%q) = %q, nil; want error", tt.value, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("normalizeOutputTypes(%q): unexpected error: %v", tt.value, err)
			}
			if got != tt.want {
				t.Errorf("normalizeOutputTypes(%q) = %q; want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestBuildSwagArgs(t *testing.T) {
	tests := []struct {
		name string
		cfg  docsConfig
		pdl  int
		want []string
	}{
		{
			name: "Defaults keep swag's own output types and parseInternal",
			cfg:  docsConfig{output: "./docs", outputTypes: defaultOutputTypes, parseInternal: true},
			pdl:  1,
			want: []string{"init", "--pdl", "1", "-g", "./goswag/main.go", "-o", "./docs", "--ot", "go,json,yaml", "--parseInternal"},
		},
		{
			name: "Spec-only output types are forwarded as --ot",
			cfg:  docsConfig{output: "./docs", outputTypes: "json,yaml", parseInternal: true},
			pdl:  0,
			want: []string{"init", "--pdl", "0", "-g", "./goswag/main.go", "-o", "./docs", "--ot", "json,yaml", "--parseInternal"},
		},
		{
			name: "parseInternal disabled",
			cfg:  docsConfig{output: "./out", outputTypes: "json", parseInternal: false},
			pdl:  3,
			want: []string{"init", "--pdl", "3", "-g", "./goswag/main.go", "-o", "./out", "--ot", "json"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildSwagArgs(tt.cfg, tt.pdl, "./goswag/main.go")
			if len(got) != len(tt.want) {
				t.Fatalf("args = %q; want %q", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("args = %q; want %q", got, tt.want)
				}
			}
		})
	}
}

func TestDocsFlagDefaults(t *testing.T) {
	cfg := docsConfig{}
	fs := newDocsFlagSet(&cfg)
	if err := fs.Parse(nil); err != nil {
		t.Fatal(err)
	}
	if cfg.outputTypes != "go,json,yaml" {
		t.Errorf("default output-types = %q; want %q", cfg.outputTypes, "go,json,yaml")
	}

	cfg = docsConfig{}
	fs = newDocsFlagSet(&cfg)
	if err := fs.Parse([]string{"--output-types", "json,yaml"}); err != nil {
		t.Fatal(err)
	}
	if cfg.outputTypes != "json,yaml" {
		t.Errorf("output-types = %q; want %q", cfg.outputTypes, "json,yaml")
	}
}

func TestRunDocs_InvalidOutputTypesFailsBeforeAnyWork(t *testing.T) {
	// A missing input directory would also fail, so pointing at one proves the
	// value is rejected first.
	err := runDocs([]string{"--output-types", "jsonn", "--input", filepath.Join(t.TempDir(), "missing")})
	if err == nil {
		t.Fatal("expected an error for an invalid --output-types value")
	}
	if !containsSubstring(err.Error(), "invalid --output-types") {
		t.Errorf("error = %q; want it to mention the invalid --output-types value", err)
	}
}

func TestDetectPDL(t *testing.T) {
	// Each scenario lays out a temp project with its own go.mod and a
	// generated goswag.go containing a curated set of imports. We then run
	// detectPDL from that directory and assert the chosen level.
	tests := []struct {
		name       string
		modulePath string
		imports    []string
		wantPDL    int
		wantReason string // substring expected in the reason
	}{
		{
			name:       "Stub with only stdlib imports — no dep parsing needed",
			modulePath: "example.com/myapp",
			imports:    []string{`"context"`, `"net/http"`},
			wantPDL:    0,
			wantReason: "no external deps",
		},
		{
			name:       "Stub with imports inside the user module — still no deps to parse",
			modulePath: "example.com/myapp",
			imports: []string{
				`"context"`,
				`"example.com/myapp/internal/handlers"`,
				`"example.com/myapp/internal/viewmodel"`,
			},
			wantPDL:    0,
			wantReason: "no external deps",
		},
		{
			name:       "Stub references an external dep — must parse external models",
			modulePath: "example.com/myapp",
			imports: []string{
				`"context"`,
				`"example.com/myapp/internal/handlers"`,
				`"github.com/diegoclair/go_utils/resterrors"`,
			},
			wantPDL:    1,
			wantReason: "github.com/diegoclair/go_utils/resterrors",
		},
		{
			name:       "Module path exact match must not be classified as external",
			modulePath: "example.com/myapp",
			imports:    []string{`"example.com/myapp"`},
			wantPDL:    0,
			wantReason: "no external deps",
		},
		{
			name:       "Module path prefix collision must not match as internal",
			modulePath: "example.com/myapp",
			// "example.com/myapp-extra" shares a prefix but is a different module.
			imports:    []string{`"example.com/myapp-extra/pkg"`},
			wantPDL:    1,
			wantReason: "example.com/myapp-extra/pkg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectDir := t.TempDir()
			goswagDir := filepath.Join(projectDir, "goswag")
			if err := os.Mkdir(goswagDir, 0o755); err != nil {
				t.Fatal(err)
			}
			writeFile(t, filepath.Join(projectDir, "go.mod"),
				"module "+tt.modulePath+"\n\ngo 1.22\n")
			writeFile(t, filepath.Join(goswagDir, "goswag.go"),
				stubWithImports(tt.imports))

			// detectPDL reads go.mod via os.Getwd, so we chdir into the temp project.
			prevWD, _ := os.Getwd()
			t.Cleanup(func() { _ = os.Chdir(prevWD) })
			if err := os.Chdir(projectDir); err != nil {
				t.Fatal(err)
			}

			pdl, reason, err := detectPDL("./goswag")
			if err != nil {
				t.Fatalf("detectPDL: unexpected error: %v", err)
			}
			if pdl != tt.wantPDL {
				t.Errorf("pdl = %d; want %d (reason: %q)", pdl, tt.wantPDL, reason)
			}
			if !containsSubstring(reason, tt.wantReason) {
				t.Errorf("reason = %q; want substring %q", reason, tt.wantReason)
			}
		})
	}
}

func TestDetectPDL_MissingStub(t *testing.T) {
	projectDir := t.TempDir()
	writeFile(t, filepath.Join(projectDir, "go.mod"), "module example.com/myapp\n\ngo 1.22\n")
	if err := os.Mkdir(filepath.Join(projectDir, "goswag"), 0o755); err != nil {
		t.Fatal(err)
	}

	prevWD, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(prevWD) })
	if err := os.Chdir(projectDir); err != nil {
		t.Fatal(err)
	}

	if _, _, err := detectPDL("./goswag"); err == nil {
		t.Fatal("expected error when goswag.go is missing")
	}
}

func TestReadModulePath_WalksUpwards(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/walks/up\n")
	nested := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	prevWD, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(prevWD) })
	if err := os.Chdir(nested); err != nil {
		t.Fatal(err)
	}

	got, err := readModulePath()
	if err != nil {
		t.Fatal(err)
	}
	if got != "example.com/walks/up" {
		t.Errorf("got %q; want %q", got, "example.com/walks/up")
	}
}

// --- helpers ---

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func stubWithImports(imports []string) string {
	s := "package main\n\nimport (\n"
	for _, imp := range imports {
		s += "\t" + imp + "\n"
	}
	s += ")\n\nfunc main() {}\n"
	return s
}

func containsSubstring(haystack, needle string) bool {
	if needle == "" {
		return true
	}
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
