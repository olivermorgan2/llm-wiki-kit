package fsafe

import (
	"errors"
	"io/fs"
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

// --- Remove: deletes a boundary-checked regular file ---

// TestRemoveDeletesRegularFile confirms the Remove primitive deletes a regular
// file inside the boundary (the rollback path relies on it to restore absence).
func TestRemoveDeletesRegularFile(t *testing.T) {
	root := t.TempDir()
	g := newGate(t, root)

	target := filepath.Join(canonRoot(t, root), "docs", "a.md")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatalf("mkdir parent: %v", err)
	}
	if err := os.WriteFile(target, []byte("bye"), 0o600); err != nil {
		t.Fatalf("seed target: %v", err)
	}

	if err := g.Remove("docs/a.md"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("target still present after Remove: stat err = %v", err)
	}
}

// TestRemoveMissingIsNotExist ensures removing an absent target yields an error
// that is fs.ErrNotExist, so resumable rollback can treat it as idempotent.
func TestRemoveMissingIsNotExist(t *testing.T) {
	root := t.TempDir()
	g := newGate(t, root)

	err := g.Remove("never-existed.txt")
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Remove(missing) err = %v, want fs.ErrNotExist", err)
	}
}

// TestRemoveRejectsTraversal ensures a "../" escape is refused, deleting nothing.
func TestRemoveRejectsTraversal(t *testing.T) {
	root := t.TempDir()
	g := newGate(t, root)

	if err := g.Remove("../outside.txt"); !errors.Is(err, ErrOutsideBoundary) {
		t.Fatalf("Remove(traversal) err = %v, want ErrOutsideBoundary", err)
	}
}

// TestRemoveRejectsSymlinkEscape ensures a symlink parent that escapes the
// boundary is refused and the external target is left intact.
func TestRemoveRejectsSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	g := newGate(t, root)

	loot := filepath.Join(outside, "loot.txt")
	if err := os.WriteFile(loot, []byte("keep"), 0o600); err != nil {
		t.Fatalf("seed loot: %v", err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "link")); err != nil {
		t.Fatalf("Symlink: %v", err)
	}

	if err := g.Remove("link/loot.txt"); !errors.Is(err, ErrSymlinkEscape) {
		t.Fatalf("Remove through escaping symlink err = %v, want ErrSymlinkEscape", err)
	}
	if _, err := os.Stat(loot); err != nil {
		t.Fatalf("external target disturbed by Remove: %v", err)
	}
}

// TestRemoveRejectsDirectory ensures a directory target is refused (transactions
// commit and roll back only regular files, per ADR-006).
func TestRemoveRejectsDirectory(t *testing.T) {
	root := t.TempDir()
	g := newGate(t, root)

	dir := filepath.Join(canonRoot(t, root), "d")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := g.Remove("d"); err == nil {
		t.Fatal("Remove(dir) = nil error, want error")
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("directory disturbed by refused Remove: %v", err)
	}
}

// TestRemoveRejectsSymlinkTarget ensures a symlink as the final component is
// refused as non-regular, and neither the link nor its target is removed.
func TestRemoveRejectsSymlinkTarget(t *testing.T) {
	root := t.TempDir()
	g := newGate(t, root)

	canon := canonRoot(t, root)
	real := filepath.Join(canon, "real.txt")
	if err := os.WriteFile(real, []byte("x"), 0o600); err != nil {
		t.Fatalf("seed real: %v", err)
	}
	link := filepath.Join(canon, "link")
	if err := os.Symlink(real, link); err != nil {
		t.Fatalf("Symlink: %v", err)
	}

	if err := g.Remove("link"); err == nil {
		t.Fatal("Remove(symlink target) = nil error, want error")
	}
	if _, err := os.Lstat(link); err != nil {
		t.Fatalf("symlink removed by refused Remove: %v", err)
	}
	if _, err := os.Stat(real); err != nil {
		t.Fatalf("symlink target removed by refused Remove: %v", err)
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

// --- WriteFile: staging symlink escapes write zero bytes outside (finding 1) ---

// countTree returns the number of regular files anywhere under dir. Used to
// assert that a poisoned staging symlink caused zero payload bytes to land
// outside the boundary.
func countTree(t *testing.T, dir string) int {
	t.Helper()
	n := 0
	err := filepath.Walk(dir, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			n++
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Walk(%q): %v", dir, err)
	}
	return n
}

// TestWriteFileRejectsStagingDirSymlink covers finding 1: a pre-planted
// .llm-wiki symlink pointing outside the boundary must not let the temp write
// (payload bytes) escape. os.MkdirAll/os.CreateTemp would follow it; the
// real-directory-chain guard refuses it before any bytes are written.
func TestWriteFileRejectsStagingDirSymlink(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()

	// root/.llm-wiki -> outside, created before the gate ever runs.
	if err := os.Symlink(outside, filepath.Join(root, StagingDir)); err != nil {
		t.Fatalf("Symlink staging dir: %v", err)
	}
	g := newGate(t, root)

	err := g.WriteFile("a.txt", []byte("payload"), 0o600)
	if !errors.Is(err, ErrSymlinkEscape) {
		t.Fatalf("WriteFile with symlinked staging err = %v, want ErrSymlinkEscape", err)
	}
	if n := countTree(t, outside); n != 0 {
		t.Fatalf("payload bytes escaped: %d files written under outside dir, want 0", n)
	}
	// The intended destination must also be untouched.
	if _, statErr := os.Stat(filepath.Join(canonRoot(t, root), "a.txt")); !os.IsNotExist(statErr) {
		t.Fatalf("destination written despite staging escape: stat err = %v", statErr)
	}
}

// TestWriteFileRejectsStagingTmpSymlink covers finding 1 one level deeper: a
// real .llm-wiki dir but a .llm-wiki/tmp symlink pointing outside. The tmp
// component must be rejected, again writing zero bytes outside.
func TestWriteFileRejectsStagingTmpSymlink(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()

	staging := filepath.Join(root, StagingDir)
	if err := os.MkdirAll(staging, 0o755); err != nil {
		t.Fatalf("mkdir staging: %v", err)
	}
	// root/.llm-wiki/tmp -> outside
	if err := os.Symlink(outside, filepath.Join(staging, "tmp")); err != nil {
		t.Fatalf("Symlink staging tmp: %v", err)
	}
	g := newGate(t, root)

	err := g.WriteFile("a.txt", []byte("payload"), 0o600)
	if !errors.Is(err, ErrSymlinkEscape) {
		t.Fatalf("WriteFile with symlinked staging tmp err = %v, want ErrSymlinkEscape", err)
	}
	if n := countTree(t, outside); n != 0 {
		t.Fatalf("payload bytes escaped: %d files written under outside dir, want 0", n)
	}
}
