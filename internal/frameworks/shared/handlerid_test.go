package shared

import (
	"strings"
	"testing"
)

func TestUniqueIdentifier(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantPrefix string // funcName the result must start with
		wantSuffix bool   // whether a "_<hash>" suffix is expected
	}{
		{
			name:       "Receiver method on a package — short name plus stable hash suffix",
			input:      "github.com/diegoclair/goswag/internal/frameworks/echo.(*Echo).GET-fm",
			wantPrefix: "GET_",
			wantSuffix: true,
		},
		{
			name:       "Plain top-level function in main — gets hash from package name",
			input:      "main.handleLogin",
			wantPrefix: "handleLogin_",
			wantSuffix: true,
		},
		{
			name:       "Raw identifier with no qualifier — returned unchanged (defensive)",
			input:      "raw",
			wantPrefix: "raw",
			wantSuffix: false,
		},
		{
			name:       "-fm suffix is stripped before hashing",
			input:      "github.com/foo/bar.(*Handler).handleLogout-fm",
			wantPrefix: "handleLogout_",
			wantSuffix: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UniqueIdentifier(tt.input)
			if tt.wantSuffix {
				if !strings.HasPrefix(got, tt.wantPrefix) {
					t.Fatalf("UniqueIdentifier(%q) = %q; want prefix %q", tt.input, got, tt.wantPrefix)
				}
				suffix := got[len(tt.wantPrefix):]
				if len(suffix) != 8 {
					t.Fatalf("UniqueIdentifier(%q) suffix = %q; want 8-char hash", tt.input, suffix)
				}
			} else if got != tt.wantPrefix {
				t.Fatalf("UniqueIdentifier(%q) = %q; want %q", tt.input, got, tt.wantPrefix)
			}
		})
	}
}

// TestUniqueIdentifier_disambiguates_collisions guards the bug that motivated
// this helper: handlers with identical short names in different packages
// (typical of monoliths with bounded contexts — provider/authroute and
// nexus/authroute both expose handleLogin) used to produce duplicate
// function declarations in the generated goswag.go.
func TestUniqueIdentifier_disambiguates_collisions(t *testing.T) {
	a := UniqueIdentifier("github.com/app/internal/provider/authroute.(*Handler).handleLogin-fm")
	b := UniqueIdentifier("github.com/app/internal/nexus/authroute.(*Handler).handleLogin-fm")

	if a == b {
		t.Fatalf("expected different identifiers for handlers in different packages, got both %q", a)
	}
	if !strings.HasPrefix(a, "handleLogin_") || !strings.HasPrefix(b, "handleLogin_") {
		t.Fatalf("expected both names to start with handleLogin_, got %q and %q", a, b)
	}
}

// TestUniqueIdentifier_is_deterministic guards against hash drift: the same
// input must always produce the same identifier so the generated goswag.go
// is stable across runs (clean diffs in version control).
func TestUniqueIdentifier_is_deterministic(t *testing.T) {
	input := "github.com/app/internal/provider/authroute.(*Handler).handleLogin-fm"
	first := UniqueIdentifier(input)
	second := UniqueIdentifier(input)
	if first != second {
		t.Fatalf("non-deterministic output: %q vs %q", first, second)
	}
}
