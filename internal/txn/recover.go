package txn

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/olivermorgan2/llm-wiki-kit/internal/fsafe"
)

// Journal marker file names. Each is a create-only marker whose presence is the
// recorded fact; its small JSON body is for debugging only.
const (
	jIntent     = "intent"
	jCommitted  = "committed"
	jRBIntent   = "rb-intent"
	jRolledBack = "rolled-back"
)

func stepMarker(i int) string   { return fmt.Sprintf("step-%s", pad4(i)) }
func rbStepMarker(i int) string { return fmt.Sprintf("rb-step-%s", pad4(i)) }

// Outcome is the disposition Recover assigned to one transaction dir.
type Outcome string

const (
	// Committed: a committed transaction whose staging was resumed and cleaned.
	Committed Outcome = "committed"
	// RolledForward: a partial commit completed from intact staged postimages.
	RolledForward Outcome = "rolled-forward"
	// RolledBack: a partial commit undone to the pre-commit tree from preimages.
	RolledBack Outcome = "rolled-back"
	// AbortedClean: staged but never committed (no intent); staging removed.
	AbortedClean Outcome = "aborted-clean"
	// CleanedUp: terminal garbage (no readable manifest); dir removed.
	CleanedUp Outcome = "cleaned-up"
)

// TxnResult is the outcome for a single recovered transaction.
type TxnResult struct {
	ID      string
	Outcome Outcome
}

// Report is the result of a Recover scan.
type Report struct {
	Transactions []TxnResult
}

// Recover scans .llm-wiki/staging under boundary, resolving every interrupted
// transaction to a terminal state — rolling forward or back per its journal and
// cleaning staging — and returns a Report of what it did. A missing or empty
// staging area yields an empty Report with no writes. Recover is idempotent:
// running it again after a complete scan yields an empty Report.
func Recover(boundary string) (Report, error) {
	return recoverWithHook(boundary, nil)
}

// recoverWithHook is Recover with an injectable step hook (test-only) used to
// crash recovery mid-rollback or mid-roll-forward and prove resumability.
func recoverWithHook(boundary string, hook func(stepOp) error) (Report, error) {
	g, err := fsafe.New(boundary)
	if err != nil {
		return Report{}, err
	}
	root, err := g.Resolve(".")
	if err != nil {
		return Report{}, err
	}

	stagingAbs := filepath.Join(root, stagingRootRel)
	dirents, err := os.ReadDir(stagingAbs)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Report{}, nil
		}
		return Report{}, err
	}

	var rep Report
	for _, de := range dirents {
		id := de.Name()
		if !txnIDRe.MatchString(id) {
			continue // not an engine transaction dir; leave it be
		}
		t := &Txn{gate: g, root: root, id: id, stepHook: hook}
		outcome, err := t.recoverOne()
		if err != nil {
			return rep, err
		}
		rep.Transactions = append(rep.Transactions, TxnResult{ID: id, Outcome: outcome})
	}
	return rep, nil
}

// recoverOne classifies one transaction dir from its manifest and journal and
// drives it to a terminal state. See the ADR-006 recovery table.
func (t *Txn) recoverOne() (Outcome, error) {
	data, err := os.ReadFile(t.abs(t.manifestRel()))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return CleanedUp, t.removeTxnDir() // no manifest: terminal garbage
		}
		return "", err
	}
	m, err := decodeManifest(data)
	if err != nil {
		return CleanedUp, t.removeTxnDir() // unreadable manifest: terminal garbage
	}
	t.manifest = m

	switch {
	case t.markerExists(jCommitted):
		if err := t.cleanup(); err != nil {
			return "", err
		}
		return Committed, nil

	case t.markerExists(jRolledBack):
		if err := t.cleanup(); err != nil {
			return "", err
		}
		return RolledBack, nil

	case t.markerExists(jRBIntent):
		// Rollback already began: never flip to forward. Resume it.
		if err := t.driveRollback(len(m.Entries) - 1); err != nil {
			return "", err
		}
		return RolledBack, nil

	case t.markerExists(jIntent):
		lastDone := t.highestStepMarker()
		if t.postimagesIntact() {
			if err := t.driveRollForward(lastDone); err != nil {
				if errors.Is(err, errPostimageDamaged) {
					if rbErr := t.driveRollback(len(m.Entries) - 1); rbErr != nil {
						return "", rbErr
					}
					return RolledBack, nil
				}
				return "", err
			}
			return RolledForward, nil
		}
		if err := t.driveRollback(len(m.Entries) - 1); err != nil {
			return "", err
		}
		return RolledBack, nil

	default:
		// Manifest present, nothing mutated (no intent): staged but never
		// committed. Remove staging, leave the tree untouched.
		if err := t.cleanup(); err != nil {
			return "", err
		}
		return AbortedClean, nil
	}
}

// driveRollForward completes the commit from lastDone+1, re-writing any step
// whose target write may not have been durably recorded (idempotent double
// write), then writes the committed marker and cleans staging.
func (t *Txn) driveRollForward(lastDone int) error {
	entries := t.manifest.Entries
	for i := lastDone + 1; i < len(entries); i++ {
		if err := t.checkpoint(stepOp{kind: opCommitStep, index: i}); err != nil {
			return err
		}
		if err := t.commitStep(i); err != nil {
			return err
		}
		if err := t.checkpoint(stepOp{kind: opStepMarker, index: i}); err != nil {
			return err
		}
		if err := t.writeMarker(stepMarker(i), stepPayload(i, entries[i].Path)); err != nil {
			return err
		}
	}
	if err := t.checkpoint(stepOp{kind: opCommitted}); err != nil {
		return err
	}
	if err := t.writeMarker(jCommitted, t.metaPayload()); err != nil {
		return err
	}
	return t.cleanup()
}

// driveRollback restores the pre-commit tree from preimage records in reverse
// commit order (a one-way door once rb-intent is written), then cleans staging.
// upper is the highest entry index that may have been mutated; restoring an
// unmutated or already-restored step is idempotent, so a rollback interrupted by
// a second crash resumes correctly by re-running from upper.
func (t *Txn) driveRollback(upper int) error {
	if err := t.checkpoint(stepOp{kind: opRBIntent}); err != nil {
		return err
	}
	if err := t.writeMarker(jRBIntent, t.metaPayload()); err != nil {
		return err
	}
	if n := len(t.manifest.Entries) - 1; upper > n {
		upper = n
	}
	for i := upper; i >= 0; i-- {
		if err := t.checkpoint(stepOp{kind: opRBRestore, index: i}); err != nil {
			return err
		}
		if err := t.restorePreimage(i); err != nil {
			return err
		}
		if err := t.checkpoint(stepOp{kind: opRBStepMarker, index: i}); err != nil {
			return err
		}
		if err := t.writeMarker(rbStepMarker(i), stepPayload(i, t.manifest.Entries[i].Path)); err != nil {
			return err
		}
	}
	if err := t.checkpoint(stepOp{kind: opRolledBack}); err != nil {
		return err
	}
	if err := t.writeMarker(jRolledBack, t.metaPayload()); err != nil {
		return err
	}
	return t.cleanup()
}

// restorePreimage returns entry i's target to its pre-commit state: an absent
// base is restored by removing the committed file (a missing file is tolerated
// for idempotent resume); a regular base is restored by atomically writing the
// verified preimage bytes with the recorded mode.
func (t *Txn) restorePreimage(i int) error {
	e := t.manifest.Entries[i]
	target := filepath.FromSlash(e.Path)
	if e.Base.Absent {
		if err := t.gate.Remove(target); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		return nil
	}
	blob, err := os.ReadFile(t.abs(t.preimageBlobRel(e.Base.SHA256)))
	if err != nil {
		return fmt.Errorf("txn: read preimage for %q: %w", e.Path, err)
	}
	if hashBytes(blob) != e.Base.SHA256 {
		return fmt.Errorf("txn: preimage blob for %q is corrupt", e.Path)
	}
	return t.gate.WriteFile(target, blob, e.Base.Mode)
}

// postimagesIntact reports whether every staged postimage still hashes to its
// recorded value, i.e. whether roll-forward is possible.
func (t *Txn) postimagesIntact() bool {
	for i, e := range t.manifest.Entries {
		sb, err := os.ReadFile(t.abs(t.stagedRel(i)))
		if err != nil || hashBytes(sb) != e.Staged {
			return false
		}
	}
	return true
}

// highestStepMarker returns the greatest entry index whose commit step marker is
// present, or -1 if none. Markers are written in commit order.
func (t *Txn) highestStepMarker() int {
	hi := -1
	for i := range t.manifest.Entries {
		if t.markerExists(stepMarker(i)) {
			hi = i
		}
	}
	return hi
}

// markerExists reports whether a journal marker file is present.
func (t *Txn) markerExists(name string) bool {
	_, err := os.Lstat(t.abs(t.journalRel(name)))
	return err == nil
}

// writeMarker writes a journal marker through the gate (create-or-overwrite; the
// presence, not the body, is the fact).
func (t *Txn) writeMarker(name string, payload []byte) error {
	return t.gate.WriteFile(t.journalRel(name), payload, internalMode)
}

// metaPayload is the small JSON body shared by the phase markers.
func (t *Txn) metaPayload() []byte {
	return []byte(fmt.Sprintf("{%q:%q}", "txn", t.id))
}

// stepPayload is the small JSON body for a per-step marker.
func stepPayload(i int, target string) []byte {
	b, _ := json.Marshal(struct {
		Step   int    `json:"step"`
		Target string `json:"target"`
	}{Step: i, Target: target})
	return b
}
