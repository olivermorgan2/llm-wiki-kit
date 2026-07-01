//go:build unix

package fsafe

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSecurityCorpusSymlinkRejected materializes each symlink-escape case from
// the corpus (create the declared symlink pointing outside the boundary, then
// hand its in-boundary lexical path to the gate) and asserts both Resolve and
// WriteFile reject it with ErrSymlinkEscape while writing zero bytes outside.
// Symlink semantics are Unix-specific, so this test is gated //go:build unix,
// matching fsafe_unix_test.go; the traversal half runs everywhere.
func TestSecurityCorpusSymlinkRejected(t *testing.T) {
	corpus := loadCorpus(t)
	if len(corpus.Symlink) == 0 {
		t.Fatal("no symlink cases in corpus")
	}
	for _, tc := range corpus.Symlink {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			root := t.TempDir()
			outside := t.TempDir()
			g := newGate(t, root)

			target := strings.ReplaceAll(tc.Target, outsidePlaceholder, canonRoot(t, outside))
			link := filepath.Join(canonRoot(t, root), filepath.FromSlash(tc.Link))
			if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
				t.Fatalf("mkdir link parent: %v", err)
			}
			if err := os.Symlink(target, link); err != nil {
				t.Fatalf("Symlink(%q -> %q): %v", link, target, err)
			}

			want := sentinelFor(t, tc.Expect)
			if _, err := g.Resolve(tc.Input); !errors.Is(err, want) {
				t.Fatalf("Resolve(%q) err = %v, want %v", tc.Input, err, want)
			}
			if err := g.WriteFile(tc.Input, []byte("payload"), 0o600); !errors.Is(err, want) {
				t.Fatalf("WriteFile(%q) err = %v, want %v", tc.Input, err, want)
			}
			if n := countTree(t, outside); n != 0 {
				t.Fatalf("payload escaped through symlink: %d files under outside dir, want 0", n)
			}
		})
	}
}
