package txn

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"

	"github.com/olivermorgan2/llm-wiki-kit/internal/fsafe"
)

// --- test helpers: canonical root + tree snapshots ---

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

type fileState struct {
	data []byte
	mode fs.FileMode
}

// snapshot records every regular file under dir (keyed by slash path relative to
// dir), skipping the engine staging area so a tree comparison reflects only the
// user-visible files a transaction is responsible for.
func snapshot(t *testing.T, dir string) map[string]fileState {
	t.Helper()
	out := map[string]fileState{}
	err := filepath.Walk(dir, func(p string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, rerr := filepath.Rel(dir, p)
		if rerr != nil {
			return rerr
		}
		rel = filepath.ToSlash(rel)
		if rel == fsafe.StagingDir || hasPrefixPath(rel, fsafe.StagingDir+"/") {
			return nil // ignore engine-managed staging/tmp
		}
		if info.Mode().IsRegular() {
			b, rerr := os.ReadFile(p)
			if rerr != nil {
				return rerr
			}
			out[rel] = fileState{data: b, mode: info.Mode().Perm()}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("snapshot(%q): %v", dir, err)
	}
	return out
}

func hasPrefixPath(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func assertTreeEqual(t *testing.T, want, got map[string]fileState) {
	t.Helper()
	if len(want) != len(got) {
		t.Fatalf("tree file count = %d, want %d\nwant=%v\ngot=%v", len(got), len(want), keys(want), keys(got))
	}
	for k, w := range want {
		g, ok := got[k]
		if !ok {
			t.Fatalf("missing file %q in tree", k)
		}
		if string(g.data) != string(w.data) {
			t.Fatalf("file %q content = %q, want %q", k, g.data, w.data)
		}
		// POSIX perm bits are not representable through package os on Windows
		// (writable files always stat 0666), so mode preservation is asserted
		// on unix only. Content/atomicity assertions above run everywhere.
		if runtime.GOOS != "windows" && g.mode != w.mode {
			t.Fatalf("file %q mode = %v, want %v", k, g.mode, w.mode)
		}
	}
}

func keys(m map[string]fileState) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

func seed(t *testing.T, root, rel, content string, mode fs.FileMode) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatalf("mkdir for %q: %v", rel, err)
	}
	if err := os.WriteFile(p, []byte(content), mode); err != nil {
		t.Fatalf("seed %q: %v", rel, err)
	}
	// os.WriteFile respects umask; force the exact bits so mode assertions hold.
	if err := os.Chmod(p, mode); err != nil {
		t.Fatalf("chmod %q: %v", rel, err)
	}
}

// --- Begin: staging layout and base records ---

func TestBeginStagesChangeSet(t *testing.T) {
	root := t.TempDir()
	canon := canonRoot(t, root)
	seed(t, canon, "docs/existing.md", "old", 0o644)
	seed(t, canon, "empty.md", "", 0o644)

	before := snapshot(t, canon)

	changes := []FileChange{
		{Target: "docs/existing.md", Data: []byte("new"), Mode: 0o644},
		{Target: "docs/new.md", Data: []byte("created"), Mode: 0o600},
		{Target: "empty.md", Data: []byte("filled"), Mode: 0o644},
	}
	tx, err := Begin(root, changes)
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}

	// Begin mutates nothing in the user tree.
	assertTreeEqual(t, before, snapshot(t, canon))

	// Manifest is present and decodes.
	mPath := filepath.Join(canon, filepath.FromSlash(tx.txnDirRel()), "manifest.json")
	data, err := os.ReadFile(mPath)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	m, err := decodeManifest(data)
	if err != nil {
		t.Fatalf("decode manifest: %v", err)
	}

	// Entries lexicographically sorted (== commit order).
	wantOrder := []string{"docs/existing.md", "docs/new.md", "empty.md"}
	if len(m.Entries) != len(wantOrder) {
		t.Fatalf("entries = %d, want %d", len(m.Entries), len(wantOrder))
	}
	for i, w := range wantOrder {
		if m.Entries[i].Path != w {
			t.Fatalf("entry[%d] = %q, want %q", i, m.Entries[i].Path, w)
		}
	}

	// Base records: existing regular, absent for new, empty != absent.
	byPath := map[string]manifestEntry{}
	for _, e := range m.Entries {
		byPath[e.Path] = e
	}
	if b := byPath["docs/existing.md"].Base; b.Absent || b.SHA256 != hashBytes([]byte("old")) {
		t.Fatalf("existing base = %+v, want regular hash(old)", b)
	}
	if b := byPath["docs/new.md"].Base; !b.Absent {
		t.Fatalf("new base = %+v, want absent", b)
	}
	if b := byPath["empty.md"].Base; b.Absent || b.SHA256 != hashBytes([]byte("")) {
		t.Fatalf("empty base = %+v, want regular hash(empty), not absent", b)
	}

	// Staged postimage files exist with the exact bytes, index == entry order.
	for i, e := range m.Entries {
		if e.Staged != hashBytes(changes[changeIndex(changes, e.Path)].Data) {
			t.Fatalf("entry %q staged hash mismatch", e.Path)
		}
		staged := filepath.Join(canon, filepath.FromSlash(tx.txnDirRel()), "files", pad4(i))
		got, rerr := os.ReadFile(staged)
		if rerr != nil {
			t.Fatalf("read staged files/%s: %v", pad4(i), rerr)
		}
		if hashBytes(got) != e.Staged {
			t.Fatalf("staged bytes for %q do not match recorded hash", e.Path)
		}
	}
}

// changeIndex finds the change whose cleaned target equals path.
func changeIndex(changes []FileChange, path string) int {
	for i, c := range changes {
		if cleanTarget(c.Target) == path {
			return i
		}
	}
	return -1
}

// --- Begin: rejection table (writes zero staging on any rejection) ---

func TestBeginRejections(t *testing.T) {
	type setup func(t *testing.T, root string) []FileChange
	cases := []struct {
		name string
		make setup
		want error
	}{
		{"empty set", func(t *testing.T, root string) []FileChange { return nil }, ErrEmptyChangeSet},
		{"traversal escape", func(t *testing.T, root string) []FileChange {
			return []FileChange{{Target: "../outside.txt", Data: []byte("x"), Mode: 0o644}}
		}, fsafe.ErrOutsideBoundary},
		{"symlink escape", func(t *testing.T, root string) []FileChange {
			outside := t.TempDir()
			if err := os.Symlink(outside, filepath.Join(canonRoot(t, root), "link")); err != nil {
				t.Fatalf("symlink: %v", err)
			}
			return []FileChange{{Target: "link/x.md", Data: []byte("x"), Mode: 0o644}}
		}, fsafe.ErrSymlinkEscape},
		{"in-boundary symlink component", func(t *testing.T, root string) []FileChange {
			canon := canonRoot(t, root)
			if err := os.MkdirAll(filepath.Join(canon, "sub"), 0o755); err != nil {
				t.Fatalf("mkdir: %v", err)
			}
			if err := os.Symlink(filepath.Join(canon, "sub"), filepath.Join(canon, "link")); err != nil {
				t.Fatalf("symlink: %v", err)
			}
			return []FileChange{{Target: "link/x.md", Data: []byte("x"), Mode: 0o644}}
		}, fsafe.ErrSymlinkEscape},
		{"non-regular dir target", func(t *testing.T, root string) []FileChange {
			if err := os.MkdirAll(filepath.Join(canonRoot(t, root), "d"), 0o755); err != nil {
				t.Fatalf("mkdir: %v", err)
			}
			return []FileChange{{Target: "d", Data: []byte("x"), Mode: 0o644}}
		}, ErrNonRegularTarget},
		{"non-regular symlink target", func(t *testing.T, root string) []FileChange {
			canon := canonRoot(t, root)
			seed(t, canon, "real.txt", "x", 0o644)
			if err := os.Symlink(filepath.Join(canon, "real.txt"), filepath.Join(canon, "link.txt")); err != nil {
				t.Fatalf("symlink: %v", err)
			}
			return []FileChange{{Target: "link.txt", Data: []byte("y"), Mode: 0o644}}
		}, ErrNonRegularTarget},
		{"duplicate target", func(t *testing.T, root string) []FileChange {
			return []FileChange{
				{Target: "docs/a.md", Data: []byte("1"), Mode: 0o644},
				{Target: "docs/./a.md", Data: []byte("2"), Mode: 0o644},
			}
		}, ErrDuplicateTarget},
		{"reserved staging target", func(t *testing.T, root string) []FileChange {
			return []FileChange{{Target: ".llm-wiki/staging/x", Data: []byte("x"), Mode: 0o644}}
		}, ErrReservedPath},
		{"reserved tmp target", func(t *testing.T, root string) []FileChange {
			return []FileChange{{Target: ".llm-wiki/tmp/x", Data: []byte("x"), Mode: 0o644}}
		}, ErrReservedPath},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			canon := canonRoot(t, root)
			changes := tc.make(t, root)
			_, err := Begin(root, changes)
			if !errors.Is(err, tc.want) {
				t.Fatalf("Begin err = %v, want %v", err, tc.want)
			}
			// A rejected Begin writes zero staging.
			if _, serr := os.Stat(filepath.Join(canon, fsafe.StagingDir, "staging")); !os.IsNotExist(serr) {
				t.Fatalf("staging created despite rejection: stat err = %v", serr)
			}
		})
	}
}

// TestBeginAcceptsDotLLMWikiTarget confirms a target elsewhere under .llm-wiki
// (e.g. the engine manifest ADR-009 writes through the transaction) is allowed;
// only the staging/ and tmp/ subtrees are reserved.
func TestBeginAcceptsDotLLMWikiTarget(t *testing.T) {
	root := t.TempDir()
	changes := []FileChange{{Target: ".llm-wiki/manifest.json", Data: []byte("{}"), Mode: 0o644}}
	tx, err := Begin(root, changes)
	if err != nil {
		t.Fatalf("Begin(.llm-wiki/manifest.json): %v", err)
	}
	if err := tx.Abort(); err != nil {
		t.Fatalf("Abort: %v", err)
	}
}

// --- Abort ---

func TestAbortRemovesStagingLeavesTreeUntouched(t *testing.T) {
	root := t.TempDir()
	canon := canonRoot(t, root)
	seed(t, canon, "keep.md", "keep", 0o644)
	before := snapshot(t, canon)

	tx, err := Begin(root, []FileChange{{Target: "new.md", Data: []byte("x"), Mode: 0o644}})
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}
	if err := tx.Abort(); err != nil {
		t.Fatalf("Abort: %v", err)
	}

	// Staging area for this txn is gone; user tree untouched.
	if _, serr := os.Stat(filepath.Join(canon, filepath.FromSlash(tx.txnDirRel()))); !os.IsNotExist(serr) {
		t.Fatalf("txn staging dir survived Abort: stat err = %v", serr)
	}
	assertTreeEqual(t, before, snapshot(t, canon))

	// Second Abort and any Commit are refused once the txn is done.
	if err := tx.Abort(); !errors.Is(err, ErrTxnDone) {
		t.Fatalf("second Abort err = %v, want ErrTxnDone", err)
	}
	if err := tx.Commit(); !errors.Is(err, ErrTxnDone) {
		t.Fatalf("Commit after Abort err = %v, want ErrTxnDone", err)
	}
}
