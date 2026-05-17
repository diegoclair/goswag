// Package shared contains helpers reused across framework adapters
// (echo, gin, and future ones).
package shared

import (
	"crypto/sha1"
	"encoding/hex"
	"strings"
)

// UniqueIdentifier returns a unique Go identifier derived from a fully
// qualified function name, used as the stub function name in the
// generated goswag.go file.
//
// Two handlers can legitimately share the same short name (e.g. handleLogin)
// when they live in different packages — typical of monoliths organised
// around bounded contexts with parallel folder structures. Using only the
// short name produces "redeclared in this block" compile errors in
// goswag.go. We disambiguate by appending a short hash of the full
// qualifier (package path + receiver) so identical short names in different
// packages produce different identifiers.
//
// The name only needs to be a valid, unique Go identifier — it never leaks
// into the OpenAPI spec, which is built from the @Router/@Summary/etc
// annotations attached to the stub.
//
// Examples:
//
//	"github.com/foo/nexus/authroute.(*Handler).handleLogin-fm"
//	  → "handleLogin_a3f2c9d1"
//	"github.com/foo/provider/authroute.(*Handler).handleLogin-fm"
//	  → "handleLogin_b71e04f8"
//	"main.handleLogin"
//	  → "handleLogin_<hash>"
//	"raw" (no qualifier — defensive fallback)
//	  → "raw"
func UniqueIdentifier(fullName string) string {
	fullName = strings.TrimSuffix(fullName, "-fm")
	parts := strings.Split(fullName, ".")
	funcName := parts[len(parts)-1]

	// No qualifier (e.g. bare "foo") — nothing to disambiguate.
	if len(parts) < 2 {
		return funcName
	}

	qualifier := fullName[:len(fullName)-len(funcName)-1]
	h := sha1.Sum([]byte(qualifier))
	return funcName + "_" + hex.EncodeToString(h[:4])
}
