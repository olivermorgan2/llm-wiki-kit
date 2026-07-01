//go:build !unix

package fsafe

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// --- Destination parent chain guards (non-unix path-based commit) ---

// TestMkdirRealChainRejectsSymlinkParent exercises the creation-time guard that
// closes the destination parent escape on the non-unix fallback. A real
// end-to-end WriteFile cannot reach mkdirRealChain with a symlink parent because
// Resolve already resolves and rejects any parent symlink that exists at
// resolution time; the residual gap is a parent that becomes a symlink *after*
// Resolve (a timing race). This white-box test plants that state directly and
// asserts mkdirRealChain — the code that runs post-Resolve — refuses to create
// through it.
func TestMkdirRealChainRejectsSymlinkParent(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	g := newGate(t, root).(*gate)

	// Simulate a parent component that appeared as a symlink escape after
	// resolution: root/newdir -> outside.
	if err := os.Symlink(outside, filepath.Join(g.root, "newdir")); err != nil {
		t.Fatalf("Symlink parent: %v", err)
	}

	err := g.mkdirRealChain(filepath.Join(g.root, "newdir", "sub"))
	if !errors.Is(err, ErrSymlinkEscape) {
		t.Fatalf("mkdirRealChain through symlink parent err = %v, want ErrSymlinkEscape", err)
	}
	// It must not have created anything through the escaping symlink.
	if _, statErr := os.Stat(filepath.Join(outside, "sub")); !os.IsNotExist(statErr) {
		t.Fatalf("directory created through symlink escape: stat err = %v", statErr)
	}
}

// TestAssertRealDirChainRejectsSymlinkParent exercises the pre-rename
// revalidation directly: even if the parent chain was a real directory when
// created, a component swapped to a symlink before commit must be caught so the
// rename cannot follow it outside the boundary.
func TestAssertRealDirChainRejectsSymlinkParent(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	g := newGate(t, root).(*gate)

	// Real parent chain first: root/a/b.
	parent := filepath.Join(g.root, "a", "b")
	if err := os.MkdirAll(parent, 0o755); err != nil {
		t.Fatalf("mkdir parent chain: %v", err)
	}
	if err := g.assertRealDirChain(parent); err != nil {
		t.Fatalf("assertRealDirChain on real chain: %v", err)
	}

	// Now swap the top component to a symlink escape (root/a -> outside) and
	// confirm revalidation rejects it before any rename could commit.
	if err := os.RemoveAll(filepath.Join(g.root, "a")); err != nil {
		t.Fatalf("remove real a: %v", err)
	}
	if err := os.Symlink(outside, filepath.Join(g.root, "a")); err != nil {
		t.Fatalf("Symlink a: %v", err)
	}
	if err := g.assertRealDirChain(parent); !errors.Is(err, ErrSymlinkEscape) {
		t.Fatalf("assertRealDirChain after symlink swap err = %v, want ErrSymlinkEscape", err)
	}
}
