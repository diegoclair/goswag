package testutil

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"testing"
)

// swag parses a goswag.go that does not compile, so only a type-check catches it.
func TypeCheckGoFile(t *testing.T, path string) {
	t.Helper()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}

	conf := types.Config{Importer: importer.Default()}
	if _, err := conf.Check("main", fset, []*ast.File{file}, nil); err != nil {
		t.Fatalf("type-check %s: %v", path, err)
	}
}
