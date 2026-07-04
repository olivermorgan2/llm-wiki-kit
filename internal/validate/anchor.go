package validate

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/olivermorgan2/llm-wiki-kit/internal/fsafe"
)

// AnchorRepo builds the ADR-008 repo-path resolution Options for a bundle rooted
// at bundleRoot. It anchors the repository at the nearest ancestor directory
// (bundle root inclusive — `install` puts `.llm-wiki/` in the bundle root, the
// common case) that contains the `.llm-wiki/` engine-metadata marker (ADR-009).
// On a hit it returns Options carrying the bundle-root-relative BundleDir and a
// read-only RepoResolve backed by an fsafe.Gate confined to the repo root, and
// ok=true. When no ancestor carries the marker it returns the zero Options and
// ok=false: the repo-path class is empty and every bundle-escaping target is
// malformed (ADR-008 bundle-root fallback). The returned RepoResolve performs a
// read-only stat through the same canonicalize/resolve-symlink primitives as the
// ADR-005 write gate and never stats above the anchor; EvidenceSections is left
// empty (profile loading is Phase 4).
func AnchorRepo(bundleRoot string) (Options, bool) {
	abs, err := filepath.Abs(bundleRoot)
	if err != nil {
		return Options{}, false
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		resolved = abs // best effort when the bundle root cannot be canonicalized
	}

	dir := resolved
	for {
		marker := filepath.Join(dir, fsafe.StagingDir)
		if info, statErr := os.Stat(marker); statErr == nil && info.IsDir() {
			return anchorAt(dir, resolved)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return Options{}, false // reached the filesystem root with no marker
		}
		dir = parent
	}
}

// anchorAt builds the Options for a repo rooted at repoRoot with the bundle at
// bundleRoot (both canonical absolute paths).
func anchorAt(repoRoot, bundleRoot string) (Options, bool) {
	rel, err := filepath.Rel(repoRoot, bundleRoot)
	if err != nil {
		return Options{}, false
	}
	bundleDir := filepath.ToSlash(rel)
	if bundleDir == "." {
		bundleDir = ""
	}

	gate, err := fsafe.New(repoRoot)
	if err != nil {
		return Options{}, false
	}
	repoResolve := func(rel string) RepoStatus {
		safe, err := gate.Resolve(rel)
		if errors.Is(err, fsafe.ErrOutsideBoundary) || errors.Is(err, fsafe.ErrSymlinkEscape) {
			return RepoEscape
		}
		if err != nil {
			return RepoAbsent
		}
		if _, err := os.Stat(safe); err != nil {
			return RepoAbsent
		}
		return RepoFound
	}
	return Options{BundleDir: bundleDir, RepoResolve: repoResolve}, true
}
