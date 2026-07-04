package validate

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// (F28) AnchorRepo finds the nearest .llm-wiki/ marker, reports the bundle root
// relative to the repo root, and its RepoResolve stats inside the repo root:
// present → RepoFound, absent → RepoAbsent.
func TestAnchorRepoFindsNearestMarker(t *testing.T) {
	repo := t.TempDir()
	mustMkdir(t, filepath.Join(repo, ".llm-wiki"))
	mustMkdir(t, filepath.Join(repo, "wiki"))
	mustMkdir(t, filepath.Join(repo, "shared"))
	mustWrite(t, filepath.Join(repo, "shared", "doc.md"), "x")

	opts, ok := AnchorRepo(filepath.Join(repo, "wiki"))
	if !ok {
		t.Fatal("AnchorRepo should find the .llm-wiki marker")
	}
	if opts.BundleDir != "wiki" {
		t.Errorf("BundleDir = %q, want wiki", opts.BundleDir)
	}
	if opts.RepoResolve == nil {
		t.Fatal("RepoResolve must be non-nil on a hit")
	}
	if got := opts.RepoResolve("shared/doc.md"); got != RepoFound {
		t.Errorf("existing repo file = %v, want RepoFound", got)
	}
	if got := opts.RepoResolve("shared/gone.md"); got != RepoAbsent {
		t.Errorf("absent repo file = %v, want RepoAbsent", got)
	}
}

// (F29) With no .llm-wiki/ marker anywhere up the tree, AnchorRepo falls back:
// ok=false and the zero Options (repo-path class empty).
func TestAnchorRepoNoMarkerFallsBack(t *testing.T) {
	// A bare temp dir under the system temp root has no .llm-wiki ancestor.
	bundle := t.TempDir()
	if opts, ok := AnchorRepo(bundle); ok {
		t.Errorf("AnchorRepo without a marker should fall back, got ok=true %+v", opts)
	}
}

// (F30) A repo-relative target traversing an in-boundary symlink whose real
// target escapes the repo root is RepoEscape (unix only, mirroring fsafe).
func TestAnchorRepoSymlinkEscapeIsRepoEscape(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink-escape semantics are exercised on unix (see internal/fsafe)")
	}
	outside := t.TempDir()
	mustWrite(t, filepath.Join(outside, "secret.md"), "s")

	repo := t.TempDir()
	mustMkdir(t, filepath.Join(repo, ".llm-wiki"))
	mustMkdir(t, filepath.Join(repo, "wiki"))
	// A symlink inside the repo pointing outside it.
	if err := os.Symlink(outside, filepath.Join(repo, "escape")); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	opts, ok := AnchorRepo(filepath.Join(repo, "wiki"))
	if !ok {
		t.Fatal("AnchorRepo should find the marker")
	}
	if got := opts.RepoResolve("escape/secret.md"); got != RepoEscape {
		t.Errorf("symlink-escaping target = %v, want RepoEscape", got)
	}
}

func mustMkdir(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
