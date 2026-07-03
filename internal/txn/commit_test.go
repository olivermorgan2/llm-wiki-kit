package txn

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// recordHook returns a step hook that appends every op to *seq and never crashes.
func recordHook(seq *[]stepOp) func(stepOp) error {
	return func(op stepOp) error {
		*seq = append(*seq, op)
		return nil
	}
}

// crashAt returns a hook that crashes (returns an error) the first time it is
// called with the op equal to want, and records every op it sees up to and
// including that point.
func crashAt(want stepOp) func(stepOp) error {
	return func(op stepOp) error {
		if op == want {
			return errors.New("injected crash")
		}
		return nil
	}
}

// --- Commit: happy path ---

func TestCommitHappyPath(t *testing.T) {
	root := t.TempDir()
	canon := canonRoot(t, root)
	seed(t, canon, "a.md", "old", 0o644)

	changes := []FileChange{
		{Target: "a.md", Data: []byte("new-a"), Mode: 0o600},
		{Target: "nested/dir/b.md", Data: []byte("created-b"), Mode: 0o640},
		{Target: "empty.md", Data: nil, Mode: 0o644},
	}
	tx, err := Begin(root, changes)
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	// Committed content and modes are exact, nested parents created.
	want := map[string]fileState{
		"a.md":            {data: []byte("new-a"), mode: 0o600},
		"nested/dir/b.md": {data: []byte("created-b"), mode: 0o640},
		"empty.md":        {data: []byte(""), mode: 0o644},
	}
	assertTreeEqual(t, want, snapshot(t, canon))

	// Staging is gone.
	if _, serr := os.Stat(filepath.Join(canon, fsafe_stagingRoot())); serr == nil {
		if entries, _ := os.ReadDir(filepath.Join(canon, fsafe_stagingRoot())); len(entries) != 0 {
			t.Fatalf("staging not cleaned after commit: %v", entries)
		}
	}

	// Second Commit is refused.
	if err := tx.Commit(); !errors.Is(err, ErrTxnDone) {
		t.Fatalf("second Commit err = %v, want ErrTxnDone", err)
	}
}

func fsafe_stagingRoot() string { return filepath.FromSlash(".llm-wiki/staging") }

// --- Commit: golden op sequence ---

func TestCommitGoldenSequence(t *testing.T) {
	root := t.TempDir()
	canon := canonRoot(t, root)
	seed(t, canon, "existing.md", "old", 0o644) // one regular base -> exercises a preimage blob

	changes := []FileChange{
		{Target: "existing.md", Data: []byte("x0"), Mode: 0o644},
		{Target: "b.md", Data: []byte("x1"), Mode: 0o644},
		{Target: "c.md", Data: []byte("x2"), Mode: 0o644},
	}
	tx, err := Begin(root, changes)
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}
	var seq []stepOp
	tx.stepHook = recordHook(&seq)
	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	want := []stepOp{
		{opCapturePreimage, 0}, {opCapturePreimage, 1}, {opCapturePreimage, 2},
		{opPreimagesDone, 0},
		{opIntent, 0},
		{opCommitStep, 0}, {opStepMarker, 0},
		{opCommitStep, 1}, {opStepMarker, 1},
		{opCommitStep, 2}, {opStepMarker, 2},
		{opCommitted, 0},
		{opCleanupManifest, 0},
		{opCleanupRemoveAll, 0},
	}
	if len(seq) != len(want) {
		t.Fatalf("op sequence length = %d, want %d\ngot: %v", len(seq), len(want), seq)
	}
	for i := range want {
		if seq[i] != want[i] {
			t.Fatalf("op[%d] = %+v, want %+v\nfull: %v", i, seq[i], want[i], seq)
		}
	}
}

// --- Commit: revalidation refuses a stale or non-regular target ---

func TestCommitStaleTargetRefused(t *testing.T) {
	root := t.TempDir()
	canon := canonRoot(t, root)
	seed(t, canon, "a.md", "old", 0o644)

	tx, err := Begin(root, []FileChange{{Target: "a.md", Data: []byte("new"), Mode: 0o644}})
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}

	// Mutate the live target after Begin.
	if err := os.WriteFile(filepath.Join(canon, "a.md"), []byte("changed-out-of-band"), 0o644); err != nil {
		t.Fatalf("mutate target: %v", err)
	}
	before := snapshot(t, canon)

	if err := tx.Commit(); !errors.Is(err, ErrStale) {
		t.Fatalf("Commit over stale target err = %v, want ErrStale", err)
	}
	// Zero tree mutation.
	assertTreeEqual(t, before, snapshot(t, canon))

	// Still Abort-able after a refused Commit.
	if err := tx.Abort(); err != nil {
		t.Fatalf("Abort after stale Commit: %v", err)
	}
}

func TestCommitNonRegularTargetRefused(t *testing.T) {
	root := t.TempDir()
	canon := canonRoot(t, root)
	seed(t, canon, "a.md", "old", 0o644)

	tx, err := Begin(root, []FileChange{{Target: "a.md", Data: []byte("new"), Mode: 0o644}})
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}

	// Replace the target with a symlink after Begin.
	if err := os.Remove(filepath.Join(canon, "a.md")); err != nil {
		t.Fatalf("rm target: %v", err)
	}
	if err := os.Symlink(filepath.Join(canon, "elsewhere"), filepath.Join(canon, "a.md")); err != nil {
		t.Fatalf("symlink target: %v", err)
	}

	if err := tx.Commit(); !errors.Is(err, ErrNonRegularTarget) {
		t.Fatalf("Commit over symlinked target err = %v, want ErrNonRegularTarget", err)
	}
	if err := tx.Abort(); err != nil {
		t.Fatalf("Abort: %v", err)
	}
}
