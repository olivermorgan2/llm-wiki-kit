package fsafe

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Security acceptance corpus (issue #6, criterion 17). The gate's own tests
// already construct traversal/symlink attacks inline; this file adds a single
// centralized, documented corpus (testdata/security-cases.json) so criterion 17
// has one auditable list of what must be rejected pre-write. The symlink half of
// the corpus is consumed by the //go:build unix companion file.

// securityCorpus is the declarative attack corpus. Traversal inputs are literal
// (with an optional {{OUTSIDE}} placeholder for an absolute path outside the
// root); symlink cases are materialized at runtime rather than committed as live
// symlinks, so the corpus stays portable.
type securityCorpus struct {
	Traversal []traversalCase `json:"traversal"`
	Symlink   []symlinkCase   `json:"symlink"`
}

type traversalCase struct {
	Name      string `json:"name"`
	Input     string `json:"input"`
	Expect    string `json:"expect"`
	Criterion int    `json:"criterion"`
}

type symlinkCase struct {
	Name      string `json:"name"`
	Link      string `json:"link"`   // symlink path relative to the boundary root
	Target    string `json:"target"` // symlink target ({{OUTSIDE}} -> a temp dir outside root)
	Input     string `json:"input"`  // path handed to the gate (traverses Link)
	Expect    string `json:"expect"`
	Criterion int    `json:"criterion"`
}

// outsidePlaceholder is substituted with a per-case temp dir outside the
// boundary, so the corpus can describe absolute-outside and symlink-escape cases
// without hardcoding machine-specific paths.
const outsidePlaceholder = "{{OUTSIDE}}"

// loadCorpus reads and decodes the security corpus from the package testdata dir.
func loadCorpus(t *testing.T) securityCorpus {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "security-cases.json"))
	if err != nil {
		t.Fatalf("read corpus: %v", err)
	}
	var c securityCorpus
	if err := json.Unmarshal(raw, &c); err != nil {
		t.Fatalf("decode corpus: %v", err)
	}
	return c
}

// sentinelFor maps a corpus "expect" string to the concrete guard sentinel.
func sentinelFor(t *testing.T, expect string) error {
	t.Helper()
	switch expect {
	case "ErrOutsideBoundary":
		return ErrOutsideBoundary
	case "ErrSymlinkEscape":
		return ErrSymlinkEscape
	default:
		t.Fatalf("unknown expect sentinel %q", expect)
		return nil
	}
}

// TestSecurityCorpusTraversalRejected asserts every traversal case is rejected
// with its sentinel by both Resolve and WriteFile, and that WriteFile leaves no
// bytes outside the boundary. Traversal is cross-platform, so this test carries
// no build constraint.
func TestSecurityCorpusTraversalRejected(t *testing.T) {
	corpus := loadCorpus(t)
	if len(corpus.Traversal) == 0 {
		t.Fatal("no traversal cases in corpus")
	}
	for _, tc := range corpus.Traversal {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			root := t.TempDir()
			outside := t.TempDir()
			g := newGate(t, root)

			input := strings.ReplaceAll(tc.Input, outsidePlaceholder, canonRoot(t, outside))
			want := sentinelFor(t, tc.Expect)

			if _, err := g.Resolve(input); !errors.Is(err, want) {
				t.Fatalf("Resolve(%q) err = %v, want %v", input, err, want)
			}
			if err := g.WriteFile(input, []byte("payload"), 0o600); !errors.Is(err, want) {
				t.Fatalf("WriteFile(%q) err = %v, want %v", input, err, want)
			}
			// No payload bytes may have landed outside the boundary.
			if n := countTree(t, outside); n != 0 {
				t.Fatalf("payload escaped: %d files under outside dir, want 0", n)
			}
		})
	}
}
