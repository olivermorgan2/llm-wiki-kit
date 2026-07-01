// Package fsafe is the mandatory engine-level filesystem-safety gate that every
// write routes through (ADR-005). Its guarantees are: canonicalize +
// boundary-check + symlink-resolve before any write, and per-file atomic
// write+rename, plus an engine-managed .llm-wiki/ staging area.
//
// This skeleton fixes only the Gate seam. The concrete guards and the
// traversal/symlink fixtures that hold the "zero out-of-boundary writes"
// release gate (criterion 17) are a later Phase 1 issue. The cross-file
// transaction model (staging manifest, commit ordering, rollback) is out of
// scope here and owned by ADR-006.
package fsafe

import "io/fs"

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
