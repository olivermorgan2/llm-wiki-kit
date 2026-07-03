//go:build unix

package fsafe

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

// commit performs the per-file atomic write entirely relative to verified
// directory file descriptors, which closes the TOCTOU check/use gaps that a
// path-based implementation cannot (ADR-005 "zero out-of-boundary writes").
//
// The chain from the canonical root down to the staging tmp dir and down to the
// destination parent is opened one component at a time with
// openat(O_NOFOLLOW|O_DIRECTORY), so any component that is (or becomes) a
// symlink is refused rather than followed. The payload temp file is created with
// openat(O_CREAT|O_EXCL|O_NOFOLLOW) inside the staging fd, and the final commit
// is a renameat between the staging fd and the destination-parent fd. renameat
// resolves both names against the open inodes, not against re-walked pathnames,
// so a concurrent attacker who swaps any parent to an external symlink after the
// chain is opened cannot redirect either the write or the rename outside the
// boundary. The verified parent used to open the fd is the exact parent used by
// the rename.
func (g *gate) commit(safe string, data []byte, perm fs.FileMode) error {
	// g.root is canonical (New EvalSymlinks-resolved it) so O_NOFOLLOW on the
	// root itself is safe; it anchors every subsequent openat.
	rootFD, err := unix.Openat(unix.AT_FDCWD, g.root,
		unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if err != nil {
		return err
	}
	defer unix.Close(rootFD)

	// Verified real staging tmp dir handle (creating .llm-wiki and
	// .llm-wiki/tmp as needed, refusing any symlink component).
	stagingFD, err := g.openRealDirChain(rootFD, []string{StagingDir, "tmp"}, true)
	if err != nil {
		return err
	}
	defer unix.Close(stagingFD)

	// Verified real destination parent handle.
	parentComps, err := g.relComponents(filepath.Dir(safe))
	if err != nil {
		return err
	}
	destFD, err := g.openRealDirChain(rootFD, parentComps, true)
	if err != nil {
		return err
	}
	defer unix.Close(destFD)

	// Create the payload temp file inside the verified staging inode. Because it
	// is created via the staging fd (not a pathname), no swap of .llm-wiki or
	// .llm-wiki/tmp to a symlink can place these bytes outside the boundary.
	tmpName, tmpFile, err := createTempAt(stagingFD, perm)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = unix.Unlinkat(stagingFD, tmpName, 0)
		}
	}()

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return err
	}
	// Set the requested mode before the file becomes visible via rename. fchmod
	// on the open fd sets the exact bits regardless of umask.
	if err := tmpFile.Chmod(perm); err != nil {
		_ = tmpFile.Close()
		return err
	}
	if err := tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}

	// Commit relative to the two verified directory descriptors. This is the
	// step that closes the destination-parent check/use gap: destFD refers to
	// the inode that was validated, so even if safe's parent path was swapped to
	// a symlink after the chain was opened, the rename targets the verified
	// directory, never the attacker's symlink.
	if err := unix.Renameat(stagingFD, tmpName, destFD, filepath.Base(safe)); err != nil {
		return err
	}
	committed = true

	// Best-effort durability: fsync the destination parent directory so the
	// rename survives power loss. Failure here does not undo the committed write.
	_ = unix.Fsync(destFD)
	return nil
}

// removeRegular unlinks base inside the verified parent directory safeParent.
// It opens the real (non-symlink) directory chain from the canonical root down
// to safeParent with openat(O_NOFOLLOW), so no swapped-in symlink parent can
// redirect the delete, then inspects base with fstatat(AT_SYMLINK_NOFOLLOW) and
// refuses anything that is not a regular file. The unlinkat uses flag 0 (never
// AT_REMOVEDIR and never following base as a symlink). A missing base surfaces
// ENOENT, which satisfies errors.Is(err, fs.ErrNotExist) for idempotent rollback.
func (g *gate) removeRegular(safeParent, base string) error {
	rootFD, err := unix.Openat(unix.AT_FDCWD, g.root,
		unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if err != nil {
		return err
	}
	defer unix.Close(rootFD)

	comps, err := g.relComponents(safeParent)
	if err != nil {
		return err
	}
	parentFD, err := g.openRealDirChain(rootFD, comps, false)
	if err != nil {
		return err
	}
	defer unix.Close(parentFD)

	var st unix.Stat_t
	if err := unix.Fstatat(parentFD, base, &st, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		return err // ENOENT here satisfies errors.Is(err, fs.ErrNotExist).
	}
	if st.Mode&unix.S_IFMT != unix.S_IFREG {
		return fmt.Errorf("fsafe: refusing to remove non-regular file %q", base)
	}
	return unix.Unlinkat(parentFD, base, 0)
}

// openRealDirChain returns a file descriptor for the directory reached by
// walking comps down from rootFD, guaranteeing every component is a real
// (non-symlink) directory reached without following a symlink. When create is
// true, missing components are created with a non-following mkdirat. The
// returned fd is a fresh descriptor the caller must close; it never aliases
// rootFD. A nil/empty comps yields a dup of rootFD (the directory is the root
// itself).
func (g *gate) openRealDirChain(rootFD int, comps []string, create bool) (int, error) {
	// Start from a dup so the walk can uniformly close each intermediate level
	// without ever closing the caller's rootFD.
	cur, err := unix.Dup(rootFD)
	if err != nil {
		return -1, err
	}
	for _, c := range comps {
		next, err := openRealDir(cur, c, create)
		unix.Close(cur)
		if err != nil {
			return -1, err
		}
		cur = next
	}
	return cur, nil
}

// openRealDir opens the child directory name of the directory referenced by
// parentFD. O_NOFOLLOW makes openat fail if the child is a symlink; O_DIRECTORY
// makes it fail if the child is not a directory. A symlink child is reported as
// ErrSymlinkEscape so the caller can distinguish a boundary escape from an
// ordinary I/O error. When create is true a missing child is created first with
// a non-following mkdirat; a lost create race is re-opened (and a
// concurrently-planted symlink still refused by O_NOFOLLOW and reclassified).
func openRealDir(parentFD int, name string, create bool) (int, error) {
	flags := unix.O_RDONLY | unix.O_DIRECTORY | unix.O_NOFOLLOW | unix.O_CLOEXEC
	fd, err := unix.Openat(parentFD, name, flags, 0)
	if err == nil {
		return fd, nil
	}
	if isSymlinkAt(parentFD, name) {
		return -1, ErrSymlinkEscape
	}
	if create && errors.Is(err, unix.ENOENT) {
		if mkErr := unix.Mkdirat(parentFD, name, 0o755); mkErr != nil && !errors.Is(mkErr, unix.EEXIST) {
			return -1, mkErr
		}
		fd, err = unix.Openat(parentFD, name, flags, 0)
		if err == nil {
			return fd, nil
		}
		if isSymlinkAt(parentFD, name) {
			return -1, ErrSymlinkEscape
		}
	}
	return -1, err
}

// isSymlinkAt reports whether name, relative to parentFD, exists as a symlink.
// It uses fstatat with AT_SYMLINK_NOFOLLOW so it inspects the link itself, and
// is only ever consulted after an openat failure to classify the cause.
func isSymlinkAt(parentFD int, name string) bool {
	var st unix.Stat_t
	if err := unix.Fstatat(parentFD, name, &st, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		return false
	}
	return st.Mode&unix.S_IFMT == unix.S_IFLNK
}

// createTempAt creates a uniquely-named temp file inside the directory
// referenced by dirFD using openat(O_CREAT|O_EXCL|O_NOFOLLOW), retrying on name
// collisions. The file is created with mode 0600; the caller fchmods it to the
// requested perm. It returns the base name (for renameat/unlinkat against
// dirFD) and the open *os.File.
func createTempAt(dirFD int, perm fs.FileMode) (string, *os.File, error) {
	flags := unix.O_RDWR | unix.O_CREAT | unix.O_EXCL | unix.O_NOFOLLOW | unix.O_CLOEXEC
	for attempt := 0; attempt < 10000; attempt++ {
		name := "w-" + randName()
		fd, err := unix.Openat(dirFD, name, flags, 0o600)
		if err == nil {
			return name, os.NewFile(uintptr(fd), name), nil
		}
		if errors.Is(err, unix.EEXIST) {
			continue
		}
		return "", nil, err
	}
	return "", nil, errors.New("fsafe: could not create a unique staging temp file")
}

// randName returns a random hex token for temp-file naming. crypto/rand makes
// names unpredictable so a pre-planted name at the temp path cannot be guessed.
func randName() string {
	var b [10]byte
	if _, err := rand.Read(b[:]); err != nil {
		// crypto/rand.Read effectively never fails; fall back to a fixed token
		// so O_EXCL still forces uniqueness via the retry loop.
		return "fallback"
	}
	return hex.EncodeToString(b[:])
}
