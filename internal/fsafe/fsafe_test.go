package fsafe

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// canonRoot mirrors New's canonicalization so tests can compare returned paths
// against the same absolute, symlink-resolved form (macOS t.TempDir lives under
// /var -> /private/var).
func canonRoot(t *testing.T, dir string) string {
	t.Helper()
	abs, err := filepath.Abs(dir)
	if err != nil {
		t.Fatalf("Abs(%q): %v", dir, err)
	}
	real, err := filepath.EvalSymlinks(abs)
	if err != nil {
		t.Fatalf("EvalSymlinks(%q): %v", abs, err)
	}
	return real
}

func newGate(t *testing.T, root string) Gate {
	t.Helper()
	g, err := New(root)
	if err != nil {
		t.Fatalf("New(%q): %v", root, err)
	}
	return g
}

// --- New: canonicalization and errors ---

func TestNewCanonicalizesValidDir(t *testing.T) {
	root := t.TempDir()
	g, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if g == nil {
		t.Fatal("New returned nil gate")
	}
	// A trivially-in-boundary resolve should land under the canonical root.
	got, err := g.Resolve("child.txt")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	want := filepath.Join(canonRoot(t, root), "child.txt")
	if got != want {
		t.Fatalf("Resolve = %q, want %q", got, want)
	}
}

func TestNewRejectsMissingPath(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist")
	if _, err := New(missing); err == nil {
		t.Fatal("New(missing) = nil error, want error")
	}
}

func TestNewRejectsNonDir(t *testing.T) {
	file := filepath.Join(t.TempDir(), "a-file")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatalf("seed file: %v", err)
	}
	if _, err := New(file); err == nil {
		t.Fatal("New(file) = nil error, want error")
	}
}

// --- Resolve: happy path ---

func TestResolveInBoundaryRelative(t *testing.T) {
	root := t.TempDir()
	g := newGate(t, root)

	got, err := g.Resolve("docs/page.md")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	want := filepath.Join(canonRoot(t, root), "docs", "page.md")
	if got != want {
		t.Fatalf("Resolve = %q, want %q", got, want)
	}
}

func TestResolveInBoundaryAbsolute(t *testing.T) {
	root := t.TempDir()
	g := newGate(t, root)

	abs := filepath.Join(canonRoot(t, root), "sub", "file.txt")
	got, err := g.Resolve(abs)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if got != abs {
		t.Fatalf("Resolve = %q, want %q", got, abs)
	}
}

// --- Resolve: traversal rejection ---

func TestResolveRejectsTraversal(t *testing.T) {
	root := t.TempDir()
	g := newGate(t, root)

	cases := []string{
		"../outside.txt",
		"a/../../x.txt",
		"../../etc/passwd",
	}
	for _, in := range cases {
		if _, err := g.Resolve(in); !errors.Is(err, ErrOutsideBoundary) {
			t.Fatalf("Resolve(%q) err = %v, want ErrOutsideBoundary", in, err)
		}
	}
}

func TestResolveRejectsAbsoluteOutside(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	g := newGate(t, root)

	target := filepath.Join(outside, "loot.txt")
	if _, err := g.Resolve(target); !errors.Is(err, ErrOutsideBoundary) {
		t.Fatalf("Resolve(%q) err = %v, want ErrOutsideBoundary", target, err)
	}
}

// --- Resolve: symlink escape rejection ---

func TestResolveRejectsSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	g := newGate(t, root)

	link := filepath.Join(root, "link")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatalf("Symlink: %v", err)
	}

	// The lexical form (root/link/loot.txt) is inside root, but the symlink
	// resolves outside -> ErrSymlinkEscape.
	if _, err := g.Resolve("link/loot.txt"); !errors.Is(err, ErrSymlinkEscape) {
		t.Fatalf("Resolve through escaping symlink err = %v, want ErrSymlinkEscape", err)
	}
}

// --- Resolve: in-boundary symlink allowed ---

func TestResolveAllowsInBoundarySymlink(t *testing.T) {
	root := t.TempDir()
	g := newGate(t, root)

	canon := canonRoot(t, root)
	sub := filepath.Join(canon, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatalf("mkdir sub: %v", err)
	}
	link := filepath.Join(canon, "link")
	if err := os.Symlink(sub, link); err != nil {
		t.Fatalf("Symlink: %v", err)
	}

	got, err := g.Resolve("link/file.txt")
	if err != nil {
		t.Fatalf("Resolve in-boundary symlink: %v", err)
	}
	want := filepath.Join(sub, "file.txt")
	if got != want {
		t.Fatalf("Resolve = %q, want %q", got, want)
	}
}

// --- WriteFile: atomic success (bytes + perm) ---

func TestWriteFileAtomicSuccess(t *testing.T) {
	root := t.TempDir()
	g := newGate(t, root)

	data := []byte("hello atomic world")
	const perm os.FileMode = 0o640
	if err := g.WriteFile("nested/dir/out.txt", data, perm); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	dest := filepath.Join(canonRoot(t, root), "nested", "dir", "out.txt")
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("ReadFile dest: %v", err)
	}
	if string(got) != string(data) {
		t.Fatalf("content = %q, want %q", got, data)
	}
	info, err := os.Stat(dest)
	if err != nil {
		t.Fatalf("Stat dest: %v", err)
	}
	if info.Mode().Perm() != perm {
		t.Fatalf("perm = %v, want %v", info.Mode().Perm(), perm)
	}
}

// --- WriteFile: staging created lazily under .llm-wiki ---

func TestWriteFileCreatesStagingLazily(t *testing.T) {
	root := t.TempDir()
	g := newGate(t, root)

	staging := filepath.Join(canonRoot(t, root), StagingDir)
	if _, err := os.Stat(staging); !os.IsNotExist(err) {
		t.Fatalf("staging exists before first write: stat err = %v", err)
	}

	if err := g.WriteFile("a.txt", []byte("x"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if _, err := os.Stat(filepath.Join(staging, "tmp")); err != nil {
		t.Fatalf("staging tmp not created after write: %v", err)
	}
}

// --- WriteFile: failing write leaves destination intact, no temp leftover ---

func TestWriteFileFailureLeavesDestinationIntact(t *testing.T) {
	root := t.TempDir()
	g := newGate(t, root)

	canon := canonRoot(t, root)
	// Make the destination an existing directory so the final rename fails
	// after the temp file has been created and written.
	dest := filepath.Join(canon, "target")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatalf("mkdir dest dir: %v", err)
	}
	sentinel := filepath.Join(dest, "keep.txt")
	if err := os.WriteFile(sentinel, []byte("original"), 0o600); err != nil {
		t.Fatalf("seed sentinel: %v", err)
	}

	err := g.WriteFile("target", []byte("clobber"), 0o600)
	if err == nil {
		t.Fatal("WriteFile over existing dir = nil error, want error")
	}

	// Destination directory and its contents untouched.
	got, readErr := os.ReadFile(sentinel)
	if readErr != nil {
		t.Fatalf("sentinel disturbed: %v", readErr)
	}
	if string(got) != "original" {
		t.Fatalf("sentinel content = %q, want original", got)
	}

	// No leftover temp files in the staging tmp dir.
	tmpDir := filepath.Join(canon, StagingDir, "tmp")
	entries, rdErr := os.ReadDir(tmpDir)
	if rdErr != nil {
		if os.IsNotExist(rdErr) {
			return // staging tmp never materialized; also acceptable-clean.
		}
		t.Fatalf("ReadDir staging tmp: %v", rdErr)
	}
	if len(entries) != 0 {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Fatalf("leftover temp files in staging: %s", strings.Join(names, ", "))
	}
}

// --- WriteFile: refuses guard targets, writes zero bytes ---

func TestWriteFileRefusesTraversalTarget(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	g := newGate(t, root)

	target := filepath.Join(outside, "escape.txt")
	err := g.WriteFile(target, []byte("payload"), 0o600)
	if !errors.Is(err, ErrOutsideBoundary) {
		t.Fatalf("WriteFile(%q) err = %v, want ErrOutsideBoundary", target, err)
	}
	if _, statErr := os.Stat(target); !os.IsNotExist(statErr) {
		t.Fatalf("escape target was written: stat err = %v", statErr)
	}
	// A rejected guard target must not even create the staging area.
	if _, statErr := os.Stat(filepath.Join(canonRoot(t, root), StagingDir)); !os.IsNotExist(statErr) {
		t.Fatalf("staging created despite guard rejection: stat err = %v", statErr)
	}
}

func TestWriteFileRefusesSymlinkEscapeTarget(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	g := newGate(t, root)

	link := filepath.Join(root, "link")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatalf("Symlink: %v", err)
	}

	err := g.WriteFile("link/escape.txt", []byte("payload"), 0o600)
	if !errors.Is(err, ErrSymlinkEscape) {
		t.Fatalf("WriteFile through escaping symlink err = %v, want ErrSymlinkEscape", err)
	}
	if _, statErr := os.Stat(filepath.Join(outside, "escape.txt")); !os.IsNotExist(statErr) {
		t.Fatalf("escape target was written: stat err = %v", statErr)
	}
}
