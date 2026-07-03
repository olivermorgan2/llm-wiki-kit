package txn

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// stepOpKind enumerates the mutating disk operations a transaction performs, in
// the order they occur during a commit and its recovery. The stepHook is invoked
// with a stepOp immediately before each one so tests can inject a crash at any
// point and assert the recovery endpoint.
type stepOpKind int

const (
	opCapturePreimage  stepOpKind = iota // about to capture the preimage for entry Index
	opPreimagesDone                      // about to write preimages.json (capture barrier)
	opIntent                             // about to write journal/intent
	opCommitStep                         // about to write target for entry Index
	opStepMarker                         // about to write journal/step-Index
	opCommitted                          // about to write journal/committed
	opCleanupManifest                    // about to remove the manifest (cleanup start)
	opCleanupRemoveAll                   // about to RemoveAll the staging dir
	opRBIntent                           // about to write journal/rb-intent
	opRBRestore                          // about to restore the preimage for entry Index
	opRBStepMarker                       // about to write journal/rb-step-Index
	opRolledBack                         // about to write journal/rolled-back
)

// stepOp identifies a single mutating operation for hook injection and the
// golden-sequence test.
type stepOp struct {
	kind  stepOpKind
	index int
}

// checkpoint invokes the step hook before a mutating op. A non-nil return
// simulates a crash and is propagated immediately with no compensation.
func (t *Txn) checkpoint(op stepOp) error {
	if t.stepHook != nil {
		return t.stepHook(op)
	}
	return nil
}

// Commit validates the staged set against the live tree, captures preimages,
// then journals an ordered sequence of per-file atomic writes from staging into
// the tree. On success the staging area is cleaned. A revalidation failure
// returns before any mutation (the transaction stays Abort-able). A real I/O
// error after the intent marker triggers the inline rollback driver, restoring
// the pre-commit tree, and Commit returns the original cause.
func (t *Txn) Commit() error {
	if t.done {
		return ErrTxnDone
	}
	if err := t.revalidate(); err != nil {
		return err
	}
	if err := t.capturePreimages(); err != nil {
		return err
	}

	if err := t.checkpoint(stepOp{kind: opIntent}); err != nil {
		return err
	}
	if err := t.writeMarker(jIntent, t.metaPayload()); err != nil {
		return err
	}

	entries := t.manifest.Entries
	for i := range entries {
		if err := t.checkpoint(stepOp{kind: opCommitStep, index: i}); err != nil {
			return err
		}
		if err := t.commitStep(i); err != nil {
			return t.failCommit(i, err)
		}
		if err := t.checkpoint(stepOp{kind: opStepMarker, index: i}); err != nil {
			return err
		}
		if err := t.writeMarker(stepMarker(i), stepPayload(i, entries[i].Path)); err != nil {
			return t.failCommit(i, err)
		}
	}

	if err := t.checkpoint(stepOp{kind: opCommitted}); err != nil {
		return err
	}
	if err := t.writeMarker(jCommitted, t.metaPayload()); err != nil {
		return t.failCommit(len(entries)-1, err)
	}

	if err := t.cleanup(); err != nil {
		return err
	}
	t.done = true
	return nil
}

// failCommit rolls back a partially-applied commit and returns the original
// cause. upper is the highest step index that may have been applied.
func (t *Txn) failCommit(upper int, cause error) error {
	if n := len(t.manifest.Entries) - 1; upper > n {
		upper = n
	}
	if rbErr := t.driveRollback(upper); rbErr != nil {
		return fmt.Errorf("txn: commit failed (%v); rollback also failed: %w", cause, rbErr)
	}
	t.done = true
	return fmt.Errorf("txn: commit rolled back to pre-commit state: %w", cause)
}

// revalidate confirms, without mutating anything, that every staged postimage
// still hashes to its record and every live target is still absent-or-regular and
// matches its recorded base hash and mode. This guarantees the captured preimage
// equals the recorded base. A mismatch is ErrStale (or ErrNonRegularTarget for a
// target that became non-regular), leaving the transaction Abort-able.
func (t *Txn) revalidate() error {
	for i, e := range t.manifest.Entries {
		sb, err := os.ReadFile(t.abs(t.stagedRel(i)))
		if err != nil {
			return fmt.Errorf("txn: read staged postimage %q: %w", e.Path, err)
		}
		if hashBytes(sb) != e.Staged {
			return fmt.Errorf("%w: staged postimage for %q changed", ErrStale, e.Path)
		}

		target := t.abs(filepath.FromSlash(e.Path))
		info, err := os.Lstat(target)
		if errors.Is(err, fs.ErrNotExist) {
			if !e.Base.Absent {
				return fmt.Errorf("%w: %q vanished since Begin", ErrStale, e.Path)
			}
			continue
		}
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("%w: %q", ErrNonRegularTarget, e.Path)
		}
		if e.Base.Absent {
			return fmt.Errorf("%w: %q appeared since Begin", ErrStale, e.Path)
		}
		cur, err := os.ReadFile(target)
		if err != nil {
			return err
		}
		if hashBytes(cur) != e.Base.SHA256 {
			return fmt.Errorf("%w: %q changed since Begin", ErrStale, e.Path)
		}
		if info.Mode().Perm() != e.Base.Mode {
			return fmt.Errorf("%w: mode of %q changed since Begin", ErrStale, e.Path)
		}
	}
	return nil
}

// capturePreimages durably backs up every existing target's pre-commit bytes as
// a content-addressed blob under preimages/, then writes preimages.json as the
// capture barrier. Absent targets contribute no blob (their base record's absent
// sentinel is the backup). This runs before the intent marker, so a crash here
// leaves the user tree untouched.
func (t *Txn) capturePreimages() error {
	var blobs []string
	for i, e := range t.manifest.Entries {
		if err := t.checkpoint(stepOp{kind: opCapturePreimage, index: i}); err != nil {
			return err
		}
		if e.Base.Absent {
			continue
		}
		bytes, err := os.ReadFile(t.abs(filepath.FromSlash(e.Path)))
		if err != nil {
			return fmt.Errorf("txn: capture preimage for %q: %w", e.Path, err)
		}
		if err := t.gate.WriteFile(t.preimageBlobRel(e.Base.SHA256), bytes, internalMode); err != nil {
			return err
		}
		blobs = append(blobs, e.Base.SHA256)
	}
	if err := t.checkpoint(stepOp{kind: opPreimagesDone}); err != nil {
		return err
	}
	data, err := marshalPreimageSet(preimageSet{Version: manifestVersion, Txn: t.id, Blobs: blobs})
	if err != nil {
		return err
	}
	return t.gate.WriteFile(t.preimageSetRel(), data, internalMode)
}

// commitStep writes entry i's staged postimage into the tree via the gate's
// per-file atomic write. It re-reads the staged bytes (the double-write) so the
// staged postimage stays intact for roll-forward and the step is idempotent. A
// staged postimage that no longer matches its hash yields errPostimageDamaged so
// the recovery driver rolls back instead of forward.
func (t *Txn) commitStep(i int) error {
	e := t.manifest.Entries[i]
	sb, err := os.ReadFile(t.abs(t.stagedRel(i)))
	if err != nil {
		return err
	}
	if hashBytes(sb) != e.Staged {
		return fmt.Errorf("%w: %q", errPostimageDamaged, e.Path)
	}
	return t.gate.WriteFile(filepath.FromSlash(e.Path), sb, e.Mode)
}
