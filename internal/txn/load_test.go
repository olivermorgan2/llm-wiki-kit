package txn

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// Load reconstructs the staged-but-not-committed transaction a prior Begin left
// on disk (the page plan / apply split): a fresh Load by id, then Commit, applies
// exactly the staged change set and cleans staging — as if a separate apply
// invocation picked up the plan.
func TestLoadThenCommitAppliesStagedSet(t *testing.T) {
	root := t.TempDir()
	canon := canonRoot(t, root)
	seed(t, canon, "a.md", "old-a", 0o644)

	changes := []FileChange{
		{Target: "a.md", Data: []byte("new-a"), Mode: 0o600},
		{Target: "nested/b.md", Data: []byte("created-b"), Mode: 0o644},
	}
	staged, err := Begin(root, changes)
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}
	id := staged.ID()

	// A separate apply reconstructs the transaction from its id alone.
	tx, err := Load(root, id)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	want := map[string]fileState{
		"a.md":        {data: []byte("new-a"), mode: 0o600},
		"nested/b.md": {data: []byte("created-b"), mode: 0o644},
	}
	assertTreeEqual(t, want, snapshot(t, canon))

	// Staging is cleaned after a committed apply.
	if entries, _ := os.ReadDir(filepath.Join(canon, fsafe_stagingRoot())); len(entries) != 0 {
		t.Fatalf("staging not cleaned after Load+Commit: %v", entries)
	}
}

// A Load then Commit over a target that changed out-of-band since Begin is a
// stale-plan rejection: ErrStale, zero tree mutation, still Abort-able.
func TestLoadThenCommitStaleTargetRejected(t *testing.T) {
	root := t.TempDir()
	canon := canonRoot(t, root)
	seed(t, canon, "a.md", "old", 0o644)

	staged, err := Begin(root, []FileChange{{Target: "a.md", Data: []byte("new"), Mode: 0o644}})
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}
	id := staged.ID()

	// Mutate the live target after the plan was staged.
	if err := os.WriteFile(filepath.Join(canon, "a.md"), []byte("changed-out-of-band"), 0o644); err != nil {
		t.Fatalf("mutate target: %v", err)
	}
	before := snapshot(t, canon)

	tx, err := Load(root, id)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if err := tx.Commit(); !errors.Is(err, ErrStale) {
		t.Fatalf("Commit over stale target err = %v, want ErrStale", err)
	}
	assertTreeEqual(t, before, snapshot(t, canon))

	// A stale plan is still Abort-able through the loaded transaction.
	if err := tx.Abort(); err != nil {
		t.Fatalf("Abort after stale Commit: %v", err)
	}
	if _, serr := os.Stat(filepath.Join(canon, filepath.FromSlash(tx.txnDirRel()))); !os.IsNotExist(serr) {
		t.Fatalf("staging dir survived Abort: stat err = %v", serr)
	}
}

// Load of an id with no staged transaction — never staged, or already committed
// and cleaned — is ErrTxnNotFound. An id that is not a valid txn token is the
// same not-found outcome, without touching the filesystem.
func TestLoadMissingOrInvalidIsNotFound(t *testing.T) {
	root := t.TempDir()

	if _, err := Load(root, "0123456789abcdef"); !errors.Is(err, ErrTxnNotFound) {
		t.Fatalf("Load missing id err = %v, want ErrTxnNotFound", err)
	}
	if _, err := Load(root, "not-a-hex-id"); !errors.Is(err, ErrTxnNotFound) {
		t.Fatalf("Load invalid id err = %v, want ErrTxnNotFound", err)
	}
}

// StagingDir locates a loaded transaction's staging directory (so a plan/apply
// layer can co-locate sidecars), and Targets reports the committed target paths
// in canonical commit order.
func TestLoadStagingDirAndTargets(t *testing.T) {
	root := t.TempDir()
	canon := canonRoot(t, root)

	staged, err := Begin(root, []FileChange{
		{Target: "z.md", Data: []byte("z"), Mode: 0o644},
		{Target: "a.md", Data: []byte("a"), Mode: 0o644},
	})
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}

	tx, err := Load(root, staged.ID())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	wantDir := filepath.Join(canon, filepath.FromSlash(tx.txnDirRel()))
	if tx.StagingDir() != wantDir {
		t.Errorf("StagingDir() = %q, want %q", tx.StagingDir(), wantDir)
	}
	if _, err := os.Stat(filepath.Join(tx.StagingDir(), "manifest.json")); err != nil {
		t.Errorf("manifest not under StagingDir(): %v", err)
	}

	got := tx.Targets()
	want := []string{"a.md", "z.md"}
	if len(got) != len(want) {
		t.Fatalf("Targets() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Targets()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
