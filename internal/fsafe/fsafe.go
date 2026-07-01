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

// WriteFile implements Gate with a per-file atomic write: Resolve first (no
// bytes written on any guard error), create the bounded parent directory and
// the lazily-created staging tmp dir, then write-to-temp + fsync + chmod +
// atomic rename into place, with a best-effort fsync of the destination parent
// directory for durability. On any error after the temp is created the temp is
// removed and the destination is left untouched (old-or-new, never
// half-written).
//
// Caveat: the atomic rename assumes a single filesystem under the boundary
// root. If a boundary ever spans multiple mounts, cross-mount rename is not
// atomic; MVP assumes one filesystem and does not handle that case.
func (g *gate) WriteFile(path string, data []byte, perm fs.FileMode) error {
	safe, err := g.Resolve(path)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(safe), 0o755); err != nil {
		return err
	}

	stagingTmp := filepath.Join(g.root, StagingDir, "tmp")
	if err := os.MkdirAll(stagingTmp, 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(stagingTmp, "w-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	committed := false
	defer func() {
		if !committed {
			_ = os.Remove(tmpName)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, perm); err != nil {
		return err
	}
	if err := os.Rename(tmpName, safe); err != nil {
		return err
	}
	committed = true

	// Best-effort fsync of the destination parent directory so the rename is
	// durable across power loss. Failure here does not undo the committed write.
	if d, err := os.Open(filepath.Dir(safe)); err == nil {
		_ = d.Sync()
		_ = d.Close()
	}
	return nil
}
