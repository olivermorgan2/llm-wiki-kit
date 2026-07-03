// Package fsafe is the mandatory engine-level filesystem-safety gate that every
// write routes through (ADR-005). Its guarantees are: canonicalize +
// boundary-check + symlink-resolve before any write, and per-file atomic
// write+rename, plus an engine-managed .llm-wiki/ staging area.
//
// This package implements the per-file primitives: the Gate seam, the
// canonicalize/boundary/symlink guards that hold the "zero out-of-boundary
// writes" release gate (criterion 17), and the per-file atomic write. The
// cross-file transaction model (staging manifest, commit ordering, rollback)
// is out of scope here and owned by ADR-006.
package fsafe

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// StagingDir is the engine-managed staging area created under the boundary
// root. For this issue (ADR-005) its sole role is to hold temp files for
// per-file atomic writes, so the write-to-temp + rename happens on the same
// filesystem as the destination. The cross-file transaction model (plan
// manifest, commit ordering, rollback) that also lives under this directory is
// owned by ADR-006 and is deliberately out of scope here.
const StagingDir = ".llm-wiki"

// Sentinel guard errors returned by Resolve and WriteFile. Callers map every
// fsafe guard/I-O error to contract.StatusSystemFailure -> ExitSystemFailure
// (5) per ADR-003's system-or-filesystem-failure bucket. fsafe stays decoupled
// from internal/contract; performing that mapping is the caller's job.
var (
	// ErrOutsideBoundary means the path lexically resolves outside the
	// configured boundary root (plain "../" traversal or an absolute path that
	// is not contained).
	ErrOutsideBoundary = errors.New("fsafe: path resolves outside the configured boundary")
	// ErrSymlinkEscape means the path stays lexically inside the boundary but
	// traverses a symlink whose real target escapes the boundary root.
	ErrSymlinkEscape = errors.New("fsafe: path traverses a symlink escaping the boundary")
)

// Gate is the single chokepoint all engine writes route through.
type Gate interface {
	// Resolve canonicalises path, resolves symlinks, and confirms it stays
	// within the configured boundary, returning the safe absolute path or an
	// error if it would escape.
	Resolve(path string) (string, error)
	// WriteFile atomically writes data to a boundary-checked path using
	// write-to-temp + fsync + rename, so the file is never observed
	// half-written.
	WriteFile(path string, data []byte, perm fs.FileMode) error
	// Remove deletes a boundary-checked regular file, routing the deletion
	// through the same canonicalize/boundary/symlink guards as WriteFile so no
	// mutation escapes the boundary. It refuses a non-regular target (symlink,
	// directory, device, socket) and returns an error satisfying
	// errors.Is(err, fs.ErrNotExist) when the target is already absent, which
	// makes preimage-restore rollback idempotent (ADR-006). This is the delete
	// half of the ADR-005 gate chokepoint; it adds no new write path.
	Remove(path string) error
}

// gate is the concrete Gate confined to a single canonicalized boundary root.
type gate struct {
	root string // absolute, symlink-resolved boundary root
}

// New returns a Gate confined to boundary. boundary must be an existing
// directory; it is canonicalized once (filepath.Abs then EvalSymlinks) and used
// as the immutable containment root for every subsequent operation.
func New(boundary string) (Gate, error) {
	abs, err := filepath.Abs(boundary)
	if err != nil {
		return nil, fmt.Errorf("fsafe: canonicalize boundary: %w", err)
	}
	root, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return nil, fmt.Errorf("fsafe: resolve boundary: %w", err)
	}
	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("fsafe: stat boundary: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("fsafe: boundary %q is not a directory", root)
	}
	return &gate{root: root}, nil
}

// Resolve implements Gate. It rejects escapes before any bytes are written:
// relative paths are root-relative, absolute paths are accepted only if
// contained, a lexical "../" escape is ErrOutsideBoundary, and a symlink whose
// real target leaves the root is ErrSymlinkEscape. In-boundary symlinks are
// allowed and their resolved target is returned.
func (g *gate) Resolve(path string) (string, error) {
	cand := path
	if !filepath.IsAbs(cand) {
		cand = filepath.Join(g.root, cand)
	}
	cand = filepath.Clean(cand)

	// Fast lexical containment check: catches plain "../" traversal and
	// absolute paths outside the root without touching the filesystem.
	if !g.within(cand) {
		return "", ErrOutsideBoundary
	}

	// The final target usually does not exist yet, so resolve symlinks on the
	// longest existing prefix and re-append the missing tail.
	resolved, err := g.resolveExisting(cand)
	if err != nil {
		return "", err
	}
	if !g.within(resolved) {
		return "", ErrSymlinkEscape
	}
	return resolved, nil
}

// within reports whether p is lexically contained by the boundary root (the
// root itself counts as contained).
func (g *gate) within(p string) bool {
	rel, err := filepath.Rel(g.root, p)
	if err != nil {
		return false
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false
	}
	return true
}

// resolveExisting walks up from cand to the longest existing ancestor,
// resolves that ancestor's symlinks, then re-appends the non-existent tail
// components. This surfaces a symlink inside the boundary that points outside
// it, even though the final target does not exist yet.
func (g *gate) resolveExisting(cand string) (string, error) {
	var missing []string
	cur := cand
	for {
		if _, err := os.Lstat(cur); err == nil {
			real, err := filepath.EvalSymlinks(cur)
			if err != nil {
				return "", err
			}
			for i := len(missing) - 1; i >= 0; i-- {
				real = filepath.Join(real, missing[i])
			}
			return real, nil
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			// Reached the filesystem root with nothing existing; the cleaned
			// lexical candidate is the best available answer.
			return cand, nil
		}
		missing = append(missing, filepath.Base(cur))
		cur = parent
	}
}

// relComponents returns the path components of dir relative to the canonical
// boundary root. A nil slice means dir is the root itself. It returns
// ErrOutsideBoundary if dir is not contained by the root, so the commit helpers
// never create or inspect anything above the boundary.
func (g *gate) relComponents(dir string) ([]string, error) {
	rel, err := filepath.Rel(g.root, dir)
	if err != nil {
		return nil, err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return nil, ErrOutsideBoundary
	}
	if rel == "." {
		return nil, nil
	}
	return strings.Split(rel, string(filepath.Separator)), nil
}

// WriteFile implements Gate with a per-file atomic write: Resolve first (no
// bytes written on any guard error), then hand off to the platform commit,
// which materializes the bounded staging tmp dir and the destination parent,
// writes the payload to a staging temp file, fsyncs it, sets its mode, and
// atomically renames it into place. On any error after the temp is created the
// temp is removed and the destination is left untouched (old-or-new, never
// half-written).
//
// Symlink hardening (ADR-005 "zero out-of-boundary writes"): on unix (darwin,
// linux) the commit performs every step relative to verified directory file
// descriptors obtained with openat(O_NOFOLLOW|O_DIRECTORY) walking down from
// the canonical root, and commits with renameat against those descriptors.
// Because the temp create and the rename resolve their names against open
// inodes rather than re-walked pathnames, the check/use gaps are closed end to
// end: no concurrent swap of .llm-wiki, .llm-wiki/tmp, or any destination
// parent component to a symlink can redirect the payload bytes or the rename
// outside the boundary. See commit in fsafe_unix.go.
//
// On non-unix platforms (windows) the commit falls back to a path-based
// implementation that still refuses pre-planted symlink components; see
// fsafe_other.go.
//
// Caveat: the atomic rename assumes a single filesystem under the boundary
// root. If a boundary ever spans multiple mounts, cross-mount rename is not
// atomic; MVP assumes one filesystem and does not handle that case.
func (g *gate) WriteFile(path string, data []byte, perm fs.FileMode) error {
	safe, err := g.Resolve(path)
	if err != nil {
		return err
	}
	return g.commit(safe, data, perm)
}

// Remove deletes a boundary-checked regular file. It resolves and boundary-checks
// the parent chain exactly like WriteFile (in-boundary symlink parents are
// resolved; an escaping parent is ErrSymlinkEscape), but never follows a symlink
// for the final component: the target is inspected without following and refused
// if it is not a regular file. The actual unlink is performed by the
// platform-specific removeRegular against the verified parent so the delete
// cannot be redirected outside the boundary. See removeRegular in
// fsafe_unix.go / fsafe_other.go.
func (g *gate) Remove(path string) error {
	cand := path
	if !filepath.IsAbs(cand) {
		cand = filepath.Join(g.root, cand)
	}
	cand = filepath.Clean(cand)
	if !g.within(cand) {
		return ErrOutsideBoundary
	}
	if cand == g.root {
		return fmt.Errorf("fsafe: refusing to remove the boundary root %q", g.root)
	}
	// Resolve the parent (following in-boundary symlinks, rejecting escapes) but
	// keep the final component literal so a symlinked target is not followed.
	safeParent, err := g.Resolve(filepath.Dir(cand))
	if err != nil {
		return err
	}
	return g.removeRegular(safeParent, filepath.Base(cand))
}
