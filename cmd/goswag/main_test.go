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
