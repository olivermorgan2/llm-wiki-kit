//go:build unix

package fsafe

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"golang.org/x/sys/unix"
)

// openRoot opens a no-follow directory fd for g.root, mirroring how commit
// anchors its walk. Test helper; the caller closes the fd.
func openRoot(t *testing.T, g *gate) int {
	t.Helper()
	fd, err := unix.Openat(unix.AT_FDCWD, g.root,
		unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if err != nil {
		t.Fatalf("openat root: %v", err)
	}
	return fd
}

// --- White-box: the directory-chain walk refuses symlink components ---

// TestOpenRealDirChainRejectsSymlinkComponent is the unix analogue of the
// non-unix mkdirRealChain guard test: a component that is a symlink escape must
// be rejected with ErrSymlinkEscape and nothing may be created through it.
func TestOpenRealDirChainRejectsSymlinkComponent(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	g := newGate(t, root).(*gate)

	// root/newdir -> outside, as if a parent became a symlink after Resolve.
	if err := os.Symlink(outside, filepath.Join(g.root, "newdir")); err != nil {
		t.Fatalf("Symlink parent: %v", err)
	}

	rootFD := openRoot(t, g)
	defer unix.Close(rootFD)

	fd, err := g.openRealDirChain(rootFD, []string{"newdir", "sub"}, true)
	if !errors.Is(err, ErrSymlinkEscape) {
		if err == nil {
			unix.Close(fd)
		}
		t.Fatalf("openRealDirChain through symlink component err = %v, want ErrSymlinkEscape", err)
	}
	// It must not have created anything through the escaping symlink.
	if _, statErr := os.Stat(filepath.Join(outside, "sub")); !os.IsNotExist(statErr) {
		t.Fatalf("directory created through symlink escape: stat err = %v", statErr)
	}
}

// --- White-box: commit binds to the verified inode, not the pathname ---

// TestCommitBindsToVerifiedDirFDNotPath is the deterministic proof that the
// destination-parent check/use gap is closed. It opens a verified fd for the
// destination parent, then makes that parent's *pathname* route through a
// symlink to an external directory, and finally commits via renameat against the
// fd. The rename must land in the originally-verified inode (inside the
// boundary), never at the path the symlink now points to.
func TestCommitBindsToVerifiedDirFDNotPath(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	g := newGate(t, root).(*gate)

	// Real chain root/p/d, then take a verified fd for p/d.
	realParent := filepath.Join(g.root, "p")
	realDir := filepath.Join(realParent, "d")
	if err := os.MkdirAll(realDir, 0o755); err != nil {
		t.Fatalf("mkdir real chain: %v", err)
	}

	rootFD := openRoot(t, g)
	defer unix.Close(rootFD)

	destFD, err := g.openRealDirChain(rootFD, []string{"p", "d"}, false)
	if err != nil {
		t.Fatalf("openRealDirChain(p/d): %v", err)
	}
	defer unix.Close(destFD)

	// Now swap the *path* p to a symlink escape while keeping the real inode
	// alive under a different name: rename p -> p_real (p/d keeps its inode,
	// which destFD still references), then root/p -> outside.
	if err := os.Rename(realParent, filepath.Join(g.root, "p_real")); err != nil {
		t.Fatalf("rename p aside: %v", err)
	}
	if err := os.Symlink(outside, realParent); err != nil {
		t.Fatalf("symlink p -> outside: %v", err)
	}

	// Stage a temp and commit relative to the two verified fds, exactly as
	// commit does.
	stagingFD, err := g.openRealDirChain(rootFD, []string{StagingDir, "tmp"}, true)
	if err != nil {
		t.Fatalf("openRealDirChain(staging): %v", err)
	}
	defer unix.Close(stagingFD)

	name, f, err := createTempAt(stagingFD, 0o600)
	if err != nil {
		t.Fatalf("createTempAt: %v", err)
	}
	if _, err := f.Write([]byte("payload")); err != nil {
		t.Fatalf("write temp: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close temp: %v", err)
	}

	if err := unix.Renameat(stagingFD, name, destFD, "f.txt"); err != nil {
		t.Fatalf("renameat into verified fd: %v", err)
	}

	// The file must live in the verified inode (now reachable as p_real/d),
	// not through the attacker's symlink into outside.
	verified := filepath.Join(g.root, "p_real", "d", "f.txt")
	if got, err := os.ReadFile(verified); err != nil || string(got) != "payload" {
		t.Fatalf("verified inode content = %q, err = %v; want payload committed to verified dir", got, err)
	}
	if n := countTree(t, outside); n != 0 {
		t.Fatalf("commit escaped through swapped symlink: %d files under outside, want 0", n)
	}
}

// --- Adversarial stress: concurrent parent swap never escapes ---

// TestWriteFileConcurrentDestParentSwapNoEscape hammers WriteFile while another
// goroutine repeatedly swaps the destination parent between a real directory and
// a symlink escaping the boundary. Which code path each write takes is timing
// dependent, but the safety invariant is not: because commit opens the parent
// with openat(O_NOFOLLOW) and commits with renameat against that fd, no payload
// bytes may ever land outside the boundary. The assertion (zero out-of-boundary
// files) therefore holds deterministically under any interleaving.
func TestWriteFileConcurrentDestParentSwapNoEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	g := newGate(t, root)
	canon := canonRoot(t, root)

	dPath := filepath.Join(canon, "d")
	if err := os.Mkdir(dPath, 0o755); err != nil {
		t.Fatalf("mkdir d: %v", err)
	}

	const iterations = 400
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			// Errors are expected while the parent is a symlink or missing; we
			// only care that nothing escapes.
			_ = g.WriteFile("d/f.txt", []byte("payload"), 0o600)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			// RemoveAll on a symlink removes only the link, never the target,
			// so the outside dir is never deleted out from under the assertion.
			_ = os.RemoveAll(dPath)
			_ = os.Symlink(outside, dPath)
			_ = os.Remove(dPath)
			_ = os.Mkdir(dPath, 0o755)
		}
	}()

	wg.Wait()

	if n := countTree(t, outside); n != 0 {
		t.Fatalf("concurrent parent swap escaped: %d files under outside, want 0", n)
	}
}

// TestWriteFileConcurrentStagingSwapNoEscape is the staging-area counterpart:
// it swaps .llm-wiki/tmp between a real directory and an external symlink while
// writing. The temp file is created via openat against the verified staging fd,
// so payload bytes can never be redirected outside the boundary regardless of
// interleaving.
func TestWriteFileConcurrentStagingSwapNoEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	g := newGate(t, root)
	canon := canonRoot(t, root)

	tmpPath := filepath.Join(canon, StagingDir, "tmp")
	if err := os.MkdirAll(tmpPath, 0o755); err != nil {
		t.Fatalf("mkdir staging tmp: %v", err)
	}

	const iterations = 400
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			_ = g.WriteFile("out.txt", []byte("payload"), 0o600)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			_ = os.RemoveAll(tmpPath)
			_ = os.Symlink(outside, tmpPath)
			_ = os.Remove(tmpPath)
			_ = os.MkdirAll(tmpPath, 0o755)
		}
	}()

	wg.Wait()

	if n := countTree(t, outside); n != 0 {
		t.Fatalf("concurrent staging swap escaped: %d files under outside, want 0", n)
	}
}
