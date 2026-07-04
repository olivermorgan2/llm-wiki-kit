package plan

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/olivermorgan2/llm-wiki-kit/internal/txn"
)

// approvalVersion is the on-disk schema version of the plan-scoped approval
// sidecar. A decode of any other version is refused so an older engine never
// misreads a newer record.
const approvalVersion = 1

// ApprovalFileName is the plan-scoped approval sidecar co-located with a staged
// transaction under .llm-wiki/staging/<txn-id>/. Its presence means apply must
// refuse until approval is explicitly granted (ADR-003 approval-required). Plan
// writes it (writeApprovalSidecar) when a staged edit drops an existing
// evidence-context citation (ADR-008 sub-decision 6, #37); the read/refuse/grant
// plumbing here is generic and indifferent to what raised the requirement.
const ApprovalFileName = "approval.json"

// ApprovalRecord is a plan's approval requirement, surfaced into the ADR-003
// envelope's approval field when apply refuses. Reason explains why approval is
// required; Paths are the affected targets. The record's mere presence in the
// staging dir is the requirement — a caller grants it out of band (the apply
// --approve flag).
type ApprovalRecord struct {
	Version int      `json:"version"`
	Reason  string   `json:"reason"`
	Paths   []string `json:"paths,omitempty"`
}

// ApplyResult is the outcome of an apply. On a committed apply, AppliedPaths
// lists the committed targets (commit order) and ApprovalRequired is nil. When a
// staged plan carries an un-granted approval requirement, ApprovalRequired
// carries it and AppliedPaths is empty — nothing was committed and the plan is
// retained on disk for a granted re-run.
type ApplyResult struct {
	TxnID            string
	AppliedPaths     []string
	ApprovalRequired *ApprovalRecord
}

// Apply commits the staged transaction identified by txnID — the id a prior page
// plan reported — into the live tree through the ADR-006 transaction layer
// (criterion 12). It re-verifies every recorded base hash against the live tree
// first; any drift is a stale-plan rejection (txn.ErrStale) that leaves the tree
// bit-identical and the plan on disk, so re-applying a stale plan rejects
// identically until it is re-planned (criterion 13). If the staged plan carries
// an un-granted approval requirement and approved is false, Apply commits nothing,
// returns the requirement, and retains the plan so a re-run with approval can
// proceed (ADR-003). txn.ErrTxnNotFound means no plan is staged under that id;
// boundary/fsafe errors pass through for the caller's system-failure mapping.
//
// A commit interrupted mid-flight by a real I/O error is rolled back to the
// pre-commit tree and its staging cleaned by the transaction layer itself before
// the error surfaces here; a clean apply cleans its own staging on success. A
// stale or approval rejection leaves the durable preview in place, exactly as
// page plan left it — the engine's recovery scan is what collects an abandoned
// preview.
func Apply(root, txnID string, approved bool) (*ApplyResult, error) {
	tx, err := txn.Load(root, txnID)
	if err != nil {
		return nil, err
	}

	if !approved {
		rec, err := readApprovalRecord(tx.StagingDir())
		if err != nil {
			return nil, err
		}
		if rec != nil {
			// Refuse: commit nothing, leave the plan staged for a granted re-run.
			return &ApplyResult{TxnID: txnID, ApprovalRequired: rec}, nil
		}
	}

	if err := tx.Commit(); err != nil {
		// Stale revalidation leaves the tree unmutated and the plan Abort-able; a
		// mid-flight failure already rolled back and cleaned its own staging. Either
		// way the tree is intact — surface the cause for the caller's mapping.
		return nil, err
	}

	return &ApplyResult{TxnID: txnID, AppliedPaths: tx.Targets()}, nil
}

// readApprovalRecord loads the approval sidecar from a transaction's staging dir,
// returning nil when no requirement is recorded (the common case). A present but
// malformed or wrong-version record is an error rather than a silent bypass, so a
// gate a plan raised is never lost to a decode failure.
func readApprovalRecord(stagingDir string) (*ApprovalRecord, error) {
	data, err := os.ReadFile(filepath.Join(stagingDir, ApprovalFileName))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("plan: read approval record: %w", err)
	}
	var rec ApprovalRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, fmt.Errorf("plan: decode approval record: %w", err)
	}
	if rec.Version != approvalVersion {
		return nil, fmt.Errorf("plan: unsupported approval record version %d (want %d)", rec.Version, approvalVersion)
	}
	return &rec, nil
}
