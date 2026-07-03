package txn

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// --- shared matrix fixtures ---

func matrixChanges() []FileChange {
	return []FileChange{
		{Target: "existing.md", Data: []byte("new-existing"), Mode: 0o644},
		{Target: "new1.md", Data: []byte("new1"), Mode: 0o644},
		{Target: "new2.md", Data: []byte("new2"), Mode: 0o600},
	}
}

// matrixCommitted is the tree after a successful commit of matrixChanges over a
// pre-commit tree holding only existing.md.
func matrixCommitted() map[string]fileState {
	return map[string]fileState{
		"existing.md": {data: []byte("new-existing"), mode: 0o644},
		"new1.md":     {data: []byte("new1"), mode: 0o644},
		"new2.md":     {data: []byte("new2"), mode: 0o600},
	}
}

func seedMatrixPre(t *testing.T, canon string) {
	t.Helper()
	seed(t, canon, "existing.md", "old", 0o644)
}

func stagingRootAbs(canon string) string {
	return filepath.Join(canon, filepath.FromSlash(".llm-wiki/staging"))
}

func assertStagingEmpty(t *testing.T, canon string) {
	t.Helper()
	entries, err := os.ReadDir(stagingRootAbs(canon))
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		t.Fatalf("ReadDir staging: %v", err)
	}
	if len(entries) != 0 {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Fatalf("staging not empty: %v", names)
	}
}

// --- Test 9: Recover no-op / idempotency ---

func TestRecoverNoStagingIsEmpty(t *testing.T) {
	root := t.TempDir()
	rep, err := Recover(root)
	if err != nil {
		t.Fatalf("Recover on empty boundary: %v", err)
	}
	if len(rep.Transactions) != 0 {
		t.Fatalf("Recover reported %d transactions, want 0", len(rep.Transactions))
	}
	// Idempotent second run.
	rep2, err := Recover(root)
	if err != nil || len(rep2.Transactions) != 0 {
		t.Fatalf("second Recover = (%+v, %v), want empty", rep2, err)
	}
}

func TestRecoverEmptyStagingIsEmpty(t *testing.T) {
	root := t.TempDir()
	canon := canonRoot(t, root)
	if err := os.MkdirAll(stagingRootAbs(canon), 0o755); err != nil {
		t.Fatalf("mkdir staging: %v", err)
	}
	rep, err := Recover(root)
	if err != nil {
		t.Fatalf("Recover: %v", err)
	}
	if len(rep.Transactions) != 0 {
		t.Fatalf("Recover reported %d transactions, want 0", len(rep.Transactions))
	}
}

// --- Test 10: interruption x recovery matrix (acceptance criteria 3, 20) ---

// probeCommitSequence records the full deterministic op sequence of a successful
// commit of matrixChanges.
func probeCommitSequence(t *testing.T) []stepOp {
	t.Helper()
	root := t.TempDir()
	canon := canonRoot(t, root)
	seedMatrixPre(t, canon)
	tx, err := Begin(root, matrixChanges())
	if err != nil {
		t.Fatalf("probe Begin: %v", err)
	}
	var seq []stepOp
	tx.stepHook = recordHook(&seq)
	if err := tx.Commit(); err != nil {
		t.Fatalf("probe Commit: %v", err)
	}
	return seq
}

// crashIsPreCommit reports whether a crash at op leaves the tree in its
// pre-commit state (the crash is before the intent marker is durably written).
func crashIsPreCommit(op stepOp) bool {
	switch op.kind {
	case opCapturePreimage, opPreimagesDone, opIntent:
		return true
	default:
		return false
	}
}

func TestInterruptionRecoveryMatrix(t *testing.T) {
	seq := probeCommitSequence(t)
	if len(seq) == 0 {
		t.Fatal("empty probe sequence")
	}
	for idx, op := range seq {
		op := op
		t.Run(fmt.Sprintf("op%02d_%d_i%d", idx, op.kind, op.index), func(t *testing.T) {
			root := t.TempDir()
			canon := canonRoot(t, root)
			seedMatrixPre(t, canon)
			pre := snapshot(t, canon)

			tx, err := Begin(root, matrixChanges())
			if err != nil {
				t.Fatalf("Begin: %v", err)
			}
			tx.stepHook = crashAt(op)
			if err := tx.Commit(); err == nil {
				t.Fatalf("Commit crashing at %+v returned nil, want injected crash", op)
			}

			if _, err := Recover(root); err != nil {
				t.Fatalf("Recover: %v", err)
			}

			want := matrixCommitted()
			if crashIsPreCommit(op) {
				want = pre
			}
			assertTreeEqual(t, want, snapshot(t, canon))
			assertStagingEmpty(t, canon)

			rep2, err := Recover(root)
			if err != nil {
				t.Fatalf("second Recover: %v", err)
			}
			if len(rep2.Transactions) != 0 {
				t.Fatalf("second Recover reported %d transactions, want 0", len(rep2.Transactions))
			}
		})
	}
}

// --- Test 11: rollback path (damaged postimage forces preimage restore) ---

func TestRecoverRollsBackOnDamagedPostimage(t *testing.T) {
	root := t.TempDir()
	canon := canonRoot(t, root)
	seed(t, canon, "e.md", "", 0o644) // empty existing -> restored empty, not absent
	seed(t, canon, "existing.md", "old", 0o644)
	pre := snapshot(t, canon)

	// Entries sort to: e.md(0), existing.md(1), new1.md(2).
	changes := []FileChange{
		{Target: "e.md", Data: []byte("filled"), Mode: 0o644},
		{Target: "existing.md", Data: []byte("new"), Mode: 0o644},
		{Target: "new1.md", Data: []byte("created"), Mode: 0o644},
	}
	tx, err := Begin(root, changes)
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}
	// Crash after committing steps 0 and 1 (e.md, existing.md), before step 2.
	tx.stepHook = crashAt(stepOp{kind: opCommitStep, index: 2})
	if err := tx.Commit(); err == nil {
		t.Fatal("Commit returned nil, want injected crash")
	}

	// Corrupt a staged postimage so roll-forward is impossible.
	if err := os.WriteFile(filepath.Join(canon, filepath.FromSlash(".llm-wiki/staging"), txDirName(t, canon), "files", "0000"), []byte("garbage"), 0o644); err != nil {
		t.Fatalf("corrupt staged postimage: %v", err)
	}

	rep, err := Recover(root)
	if err != nil {
		t.Fatalf("Recover: %v", err)
	}
	if len(rep.Transactions) != 1 || rep.Transactions[0].Outcome != RolledBack {
		t.Fatalf("Recover report = %+v, want single RolledBack", rep.Transactions)
	}
	// Tree is byte-and-mode identical to pre-commit (empty e.md restored empty).
	assertTreeEqual(t, pre, snapshot(t, canon))
	assertStagingEmpty(t, canon)
}

// txDirName returns the single transaction dir name under staging (the tests
// create exactly one at a time).
func txDirName(t *testing.T, canon string) string {
	t.Helper()
	entries, err := os.ReadDir(stagingRootAbs(canon))
	if err != nil {
		t.Fatalf("ReadDir staging: %v", err)
	}
	for _, e := range entries {
		if txnIDRe.MatchString(e.Name()) {
			return e.Name()
		}
	}
	t.Fatal("no transaction dir under staging")
	return ""
}

// --- Test 12: crash during rollback resumes and never flips forward ---

// setupDamagedPartial builds a post-intent partial commit whose staged postimage
// is damaged, so Recover must roll back. Returns the root, canonical root, and
// the pre-commit snapshot.
func setupDamagedPartial(t *testing.T) (root, canon string, pre map[string]fileState) {
	t.Helper()
	root = t.TempDir()
	canon = canonRoot(t, root)
	seed(t, canon, "existing.md", "old", 0o644)
	pre = snapshot(t, canon)

	// Entries sort to: existing.md(0), new1.md(1).
	changes := []FileChange{
		{Target: "existing.md", Data: []byte("new"), Mode: 0o644},
		{Target: "new1.md", Data: []byte("created"), Mode: 0o644},
	}
	tx, err := Begin(root, changes)
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}
	// Crash after writing existing.md's bytes but before its step marker.
	tx.stepHook = crashAt(stepOp{kind: opCommitStep, index: 1})
	if err := tx.Commit(); err == nil {
		t.Fatal("Commit returned nil, want injected crash")
	}
	// Damage new1.md's staged postimage so roll-forward is refused.
	if err := os.WriteFile(filepath.Join(stagingRootAbs(canon), txDirName(t, canon), "files", "0001"), []byte("garbage"), 0o644); err != nil {
		t.Fatalf("corrupt staged postimage: %v", err)
	}
	return root, canon, pre
}

func probeRollbackSequence(t *testing.T) []stepOp {
	t.Helper()
	root, _, _ := setupDamagedPartial(t)
	var seq []stepOp
	if _, err := recoverWithHook(root, recordHook(&seq)); err != nil {
		t.Fatalf("probe rollback Recover: %v", err)
	}
	return seq
}

func TestRecoverResumesRollbackAtEachStep(t *testing.T) {
	seq := probeRollbackSequence(t)
	// Only interrupt rollback-phase ops (rb-intent through rolled-back and
	// cleanup); forward-phase ops never appear once postimages are damaged.
	for idx, op := range seq {
		op := op
		t.Run(fmt.Sprintf("rb%02d_%d_i%d", idx, op.kind, op.index), func(t *testing.T) {
			root, canon, pre := setupDamagedPartial(t)

			// First recovery crashes partway through the rollback.
			if _, err := recoverWithHook(root, crashAt(op)); err == nil {
				t.Fatalf("recoverWithHook crashing at %+v returned nil", op)
			}
			// A fresh recovery resumes and completes it.
			if _, err := Recover(root); err != nil {
				t.Fatalf("resume Recover: %v", err)
			}
			assertTreeEqual(t, pre, snapshot(t, canon))
			assertStagingEmpty(t, canon)
		})
	}
}

func TestRecoverRollbackIsOneWayDoor(t *testing.T) {
	root, canon, pre := setupDamagedPartial(t)

	// Enter rollback and crash right after rb-intent is written (before the
	// first preimage is restored).
	if _, err := recoverWithHook(root, crashAt(stepOp{kind: opRBRestore, index: 1})); err == nil {
		t.Fatal("recoverWithHook returned nil, want crash")
	}

	// Repair the damaged staged postimage so postimages are now intact. Even so,
	// recovery must NOT flip to roll-forward once rb-intent exists.
	dir := txDirName(t, canon)
	repaired := []byte("created")
	if err := os.WriteFile(filepath.Join(stagingRootAbs(canon), dir, "files", "0001"), repaired, 0o644); err != nil {
		t.Fatalf("repair staged postimage: %v", err)
	}

	rep, err := Recover(root)
	if err != nil {
		t.Fatalf("Recover: %v", err)
	}
	if len(rep.Transactions) != 1 || rep.Transactions[0].Outcome != RolledBack {
		t.Fatalf("outcome = %+v, want RolledBack (one-way door)", rep.Transactions)
	}
	assertTreeEqual(t, pre, snapshot(t, canon))
	assertStagingEmpty(t, canon)
}

// --- Test 13: cleanup / garbage classification ---

func TestRecoverCleansManifestlessGarbage(t *testing.T) {
	root := t.TempDir()
	canon := canonRoot(t, root)
	garbage := filepath.Join(stagingRootAbs(canon), "0123456789abcdef")
	if err := os.MkdirAll(filepath.Join(garbage, "files"), 0o755); err != nil {
		t.Fatalf("mkdir garbage: %v", err)
	}
	if err := os.WriteFile(filepath.Join(garbage, "files", "0000"), []byte("orphan"), 0o644); err != nil {
		t.Fatalf("seed garbage: %v", err)
	}

	rep, err := Recover(root)
	if err != nil {
		t.Fatalf("Recover: %v", err)
	}
	if len(rep.Transactions) != 1 || rep.Transactions[0].Outcome != CleanedUp {
		t.Fatalf("report = %+v, want single CleanedUp", rep.Transactions)
	}
	if _, err := os.Stat(garbage); !os.IsNotExist(err) {
		t.Fatalf("garbage dir survived Recover: stat err = %v", err)
	}
}

func TestRecoverResumesCommittedCleanup(t *testing.T) {
	root := t.TempDir()
	canon := canonRoot(t, root)
	seedMatrixPre(t, canon)

	tx, err := Begin(root, matrixChanges())
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}
	// Crash right when cleanup is about to remove the manifest: the tree is fully
	// committed and the committed marker is present.
	tx.stepHook = crashAt(stepOp{kind: opCleanupManifest})
	if err := tx.Commit(); err == nil {
		t.Fatal("Commit returned nil, want crash")
	}

	rep, err := Recover(root)
	if err != nil {
		t.Fatalf("Recover: %v", err)
	}
	if len(rep.Transactions) != 1 || rep.Transactions[0].Outcome != Committed {
		t.Fatalf("report = %+v, want single Committed", rep.Transactions)
	}
	assertTreeEqual(t, matrixCommitted(), snapshot(t, canon))
	assertStagingEmpty(t, canon)
}

func TestRecoverHandlesMultipleTxnDirsIndependently(t *testing.T) {
	root := t.TempDir()
	canon := canonRoot(t, root)
	seedMatrixPre(t, canon)

	// One real committed-but-uncleaned transaction.
	tx, err := Begin(root, matrixChanges())
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}
	tx.stepHook = crashAt(stepOp{kind: opCleanupManifest})
	if err := tx.Commit(); err == nil {
		t.Fatal("Commit returned nil, want crash")
	}
	// One manifest-less garbage dir alongside it.
	garbage := filepath.Join(stagingRootAbs(canon), "abcdefabcdefabcd")
	if err := os.MkdirAll(garbage, 0o755); err != nil {
		t.Fatalf("mkdir garbage: %v", err)
	}

	rep, err := Recover(root)
	if err != nil {
		t.Fatalf("Recover: %v", err)
	}
	if len(rep.Transactions) != 2 {
		t.Fatalf("Recover handled %d dirs, want 2: %+v", len(rep.Transactions), rep.Transactions)
	}
	assertTreeEqual(t, matrixCommitted(), snapshot(t, canon))
	assertStagingEmpty(t, canon)
}

// --- Test 14: inline rollback on a real mid-commit error ---

func TestCommitInlineRollbackOnStepError(t *testing.T) {
	root := t.TempDir()
	canon := canonRoot(t, root)
	seed(t, canon, "existing.md", "old", 0o644)
	pre := snapshot(t, canon)

	// Entries sort to: existing.md(0), new1.md(1).
	changes := []FileChange{
		{Target: "existing.md", Data: []byte("new"), Mode: 0o644},
		{Target: "new1.md", Data: []byte("created"), Mode: 0o644},
	}
	tx, err := Begin(root, changes)
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}
	// Non-crashing hook: corrupt entry 1's staged postimage before step 0 runs.
	// Step 0 commits, step 1's commitStep then fails on the damaged postimage,
	// triggering the inline journaled rollback.
	dir := txDirName(t, canon)
	tx.stepHook = func(op stepOp) error {
		if op == (stepOp{kind: opCommitStep, index: 0}) {
			_ = os.WriteFile(filepath.Join(stagingRootAbs(canon), dir, "files", "0001"), []byte("garbage"), 0o644)
		}
		return nil
	}

	err = tx.Commit()
	if err == nil {
		t.Fatal("Commit returned nil, want rollback error")
	}
	if !errors.Is(err, errPostimageDamaged) {
		t.Fatalf("Commit err = %v, want errPostimageDamaged cause", err)
	}
	assertTreeEqual(t, pre, snapshot(t, canon))
	assertStagingEmpty(t, canon)

	// A refused-and-rolled-back commit is terminal.
	if err := tx.Commit(); !errors.Is(err, ErrTxnDone) {
		t.Fatalf("post-rollback Commit err = %v, want ErrTxnDone", err)
	}
}

// --- Test 15: all transaction writes stay inside the boundary ---

func TestCommitWritesStayInsideBoundary(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	canon := canonRoot(t, root)
	seed(t, canon, "existing.md", "old", 0o644)

	tx, err := Begin(root, matrixChanges())
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}
	if n := countRegularFiles(t, outside); n != 0 {
		t.Fatalf("%d files escaped the boundary, want 0", n)
	}
}

func TestBeginRefusesPoisonedStagingSymlink(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	// Pre-plant .llm-wiki as a symlink escaping the boundary.
	if err := os.Symlink(outside, filepath.Join(canonRoot(t, root), ".llm-wiki")); err != nil {
		t.Fatalf("symlink staging: %v", err)
	}
	_, err := Begin(root, matrixChanges())
	if err == nil {
		t.Fatal("Begin through poisoned staging symlink = nil error, want ErrSymlinkEscape")
	}
	if n := countRegularFiles(t, outside); n != 0 {
		t.Fatalf("%d files escaped through poisoned staging, want 0", n)
	}
}

// A manifest whose recorded txn id does not match its staging dir is not a
// trustworthy in-flight transaction: Recover must treat it as garbage and clean
// it up rather than drive recovery from its (untrusted) entries and journal.
func TestRecoverCleansManifestWithMismatchedTxnID(t *testing.T) {
	root := t.TempDir()
	canon := canonRoot(t, root)
	seed(t, canon, "victim.md", "original", 0o644)
	pre := snapshot(t, canon)

	id := "0123456789abcdef"
	dir := filepath.Join(stagingRootAbs(canon), id)
	if err := os.MkdirAll(filepath.Join(dir, "journal"), 0o755); err != nil {
		t.Fatalf("mkdir staging txn: %v", err)
	}
	// Manifest bound to a different txn id, with an entry that would mutate the
	// tree if recovery trusted it.
	m := manifest{Version: manifestVersion, Txn: "ffffffffffffffff", Entries: []manifestEntry{
		{Path: "victim.md", Mode: 0o644, Staged: hashBytes([]byte("HACKED")), Base: absentBase()},
	}}
	data, err := marshalManifest(m)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), data, 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	// An intent marker: without the id check, recovery would try to roll this
	// forward/back from the untrusted manifest.
	if err := os.WriteFile(filepath.Join(dir, "journal", jIntent), []byte("{}"), 0o644); err != nil {
		t.Fatalf("write intent marker: %v", err)
	}

	rep, err := Recover(root)
	if err != nil {
		t.Fatalf("Recover: %v", err)
	}
	if len(rep.Transactions) != 1 || rep.Transactions[0].Outcome != CleanedUp {
		t.Fatalf("report = %+v, want single CleanedUp", rep.Transactions)
	}
	assertTreeEqual(t, pre, snapshot(t, canon)) // victim.md untouched
	if _, serr := os.Stat(dir); !os.IsNotExist(serr) {
		t.Fatalf("mismatched-txn staging dir survived: %v", serr)
	}
}

// seedOutsideTxnDir plants a manifest-less 16-hex "transaction" dir with a file
// under outside, the shape Recover would garbage-collect if it followed a
// poisoned staging symlink out of the boundary. Returns the txn dir path.
func seedOutsideTxnDir(t *testing.T, dir string) string {
	t.Helper()
	txnDir := filepath.Join(dir, "0123456789abcdef")
	if err := os.MkdirAll(filepath.Join(txnDir, "files"), 0o755); err != nil {
		t.Fatalf("mkdir outside txn: %v", err)
	}
	if err := os.WriteFile(filepath.Join(txnDir, "files", "0000"), []byte("precious"), 0o644); err != nil {
		t.Fatalf("seed outside txn: %v", err)
	}
	return txnDir
}

func assertOutsideTxnIntact(t *testing.T, txnDir string) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(txnDir, "files", "0000"))
	if err != nil {
		t.Fatalf("outside content removed or unreadable: %v", err)
	}
	if string(b) != "precious" {
		t.Fatalf("outside content mutated: got %q, want %q", b, "precious")
	}
}

// Recover must not follow a poisoned .llm-wiki symlink out of the boundary and
// garbage-collect the external directory it points at.
func TestRecoverRefusesPoisonedLlmWikiSymlink(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	canon := canonRoot(t, root)

	// .llm-wiki -> outside, so .llm-wiki/staging/<id> resolves to
	// outside/staging/<id>, a manifest-less dir Recover would otherwise clean up.
	outsideTxn := seedOutsideTxnDir(t, filepath.Join(outside, "staging"))
	if err := os.Symlink(outside, filepath.Join(canon, ".llm-wiki")); err != nil {
		t.Fatalf("symlink .llm-wiki: %v", err)
	}

	if _, err := Recover(root); err == nil {
		t.Fatal("Recover through poisoned .llm-wiki symlink = nil error, want refusal")
	}
	assertOutsideTxnIntact(t, outsideTxn)
}

// Recover must not follow a poisoned .llm-wiki/staging symlink out of the
// boundary, even when .llm-wiki itself is a real directory.
func TestRecoverRefusesPoisonedStagingSymlink(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	canon := canonRoot(t, root)

	outsideTxn := seedOutsideTxnDir(t, outside)
	if err := os.MkdirAll(filepath.Join(canon, ".llm-wiki"), 0o755); err != nil {
		t.Fatalf("mkdir .llm-wiki: %v", err)
	}
	if err := os.Symlink(outside, filepath.Join(canon, ".llm-wiki", "staging")); err != nil {
		t.Fatalf("symlink staging: %v", err)
	}

	if _, err := Recover(root); err == nil {
		t.Fatal("Recover through poisoned staging symlink = nil error, want refusal")
	}
	assertOutsideTxnIntact(t, outsideTxn)
}

func countRegularFiles(t *testing.T, dir string) int {
	t.Helper()
	n := 0
	err := filepath.Walk(dir, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			n++
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Walk(%q): %v", dir, err)
	}
	return n
}
