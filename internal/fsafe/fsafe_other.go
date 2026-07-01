//go:build !unix

package fsafe

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// commit is the non-unix (windows) fallback per-file atomic write. Unlike the
// unix path (fsafe_unix.go), this platform lacks openat/renameat-against-fd
// semantics on Go 1.24, so it uses a path-based implementation: it materializes
// the staging tmp dir and destination parent as real (non-symlink) directory
// chains, revalidates the destination parent immediately before the rename, and
// commits with os.Rename.
//
// This still refuses pre-planted symlink components (mkdirRealChain) and a
// parent swapped to a symlink observed at revalidation (assertRealDirChain), so
// no bytes are written through a symlink escape under sequential use. It does
// not fully close the concurrent check/use window that renameat-against-fd
// closes on unix; that residual is accepted here because unix (darwin, linux)
// is the primary engine target and provides the closed guarantee.
func (g *gate) commit(safe string, data []byte, perm fs.FileMode) error {
	stagingTmp := filepath.Join(g.root, StagingDir, "tmp")
	if err := g.mkdirRealChain(stagingTmp); err != nil {
		return err
	}

	destParent := filepath.Dir(safe)
	if err := g.mkdirRealChain(destParent); err != nil {
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
	if err := g.assertRealDirChain(destParent); err != nil {
		return err
	}
	if err := os.Rename(tmpName, safe); err != nil {
		return err
	}
	committed = true

	if d, err := os.Open(filepath.Dir(safe)); err == nil {
		_ = d.Sync()
		_ = d.Close()
	}
	return nil
}

// mkdirRealChain creates dir by walking from the canonical root and ensuring
// every component is a real (non-symlink) directory, creating missing ones with
// a non-following os.Mkdir. It rejects with ErrSymlinkEscape the moment a
// component exists as a symlink, so no bytes are ever written through a symlink
// escape. This replaces os.MkdirAll, which silently follows symlink components.
func (g *gate) mkdirRealChain(dir string) error {
	comps, err := g.relComponents(dir)
	if err != nil {
		return err
	}
	// g.root is canonical (New EvalSymlinks-resolved it) and thus a real dir.
	cur := g.root
	for _, c := range comps {
		cur = filepath.Join(cur, c)
		if err := ensureRealDir(cur); err != nil {
			return err
		}
	}
	return nil
}

// assertRealDirChain verifies, without creating anything, that every component
// from the root down to and including dir is an existing real directory. It is
// the pre-commit revalidation that rejects a parent that became a symlink after
// mkdirRealChain with ErrSymlinkEscape before the rename can follow it.
func (g *gate) assertRealDirChain(dir string) error {
	comps, err := g.relComponents(dir)
	if err != nil {
		return err
	}
	cur := g.root
	for _, c := range comps {
		cur = filepath.Join(cur, c)
		info, err := os.Lstat(cur)
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return ErrSymlinkEscape
		}
		if !info.IsDir() {
			return fmt.Errorf("fsafe: %q is not a directory", cur)
		}
	}
	return nil
}

// ensureRealDir makes path a real directory: it accepts an existing real
// directory, creates a missing one with a non-following os.Mkdir, rejects an
// existing symlink with ErrSymlinkEscape, and rejects an existing non-directory.
// A lost create race (EEXIST) is re-checked via Lstat so a concurrently-planted
// symlink is still caught.
func ensureRealDir(path string) error {
	if info, err := os.Lstat(path); err == nil {
		return classifyRealDir(path, info)
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.Mkdir(path, 0o755); err != nil {
		// Someone may have created it concurrently; re-inspect rather than
		// trust the create. If it is now a real dir we proceed, otherwise the
		// symlink/non-dir classification (or the original error) wins.
		info, statErr := os.Lstat(path)
		if statErr != nil {
			return err
		}
		return classifyRealDir(path, info)
	}
	return nil
}

// classifyRealDir accepts a real directory and rejects symlinks/non-directories.
func classifyRealDir(path string, info os.FileInfo) error {
	if info.Mode()&os.ModeSymlink != 0 {
		return ErrSymlinkEscape
	}
	if !info.IsDir() {
		return fmt.Errorf("fsafe: %q exists and is not a directory", path)
	}
	return nil
}
