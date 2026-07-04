package plan

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/olivermorgan2/llm-wiki-kit/internal/txn"
	"github.com/olivermorgan2/llm-wiki-kit/internal/yamladapter"
)

// stageNewPagePlan plans a brand-new page under root and returns the staging
// transaction id its later apply consumes.
func stageNewPagePlan(t *testing.T, root, rel, content string) string {
	t.Helper()
	res, err := Plan(root, rel, []byte(content), yamladapter.New())
	if err != nil {
		t.Fatalf("Plan(%s): %v", rel, err)
	}
	if res.NoOp || res.TxnID == "" {
		t.Fatalf("expected a staged (non-no-op) plan for %s: %+v", rel, res)
	}
	return res.TxnID
}

// A clean apply commits exactly the previewed page and cleans staging
// (criterion 12).
func TestApplyCommitsPreviewedPage(t *testing.T) {
	root := t.TempDir()
	txnID := stageNewPagePlan(t, root, "wiki/fresh.md", validPage)

	res, err := Apply(root, txnID, false)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if res.ApprovalRequired != nil {
		t.Fatalf("clean apply must not require approval: %+v", res.ApprovalRequired)
	}
	if len(res.AppliedPaths) != 1 || res.AppliedPaths[0] != "wiki/fresh.md" {
		t.Fatalf("AppliedPaths = %v, want [wiki/fresh.md]", res.AppliedPaths)
	}

	// The committed page is the normalized staged content.
	got, err := os.ReadFile(filepath.Join(root, "wiki", "fresh.md"))
	if err != nil {
		t.Fatalf("read committed page: %v", err)
	}
	if want := normalizePage([]byte(validPage), yamladapter.New()); string(got) != string(want) {
		t.Errorf("committed page = %q, want %q", got, want)
	}

	// Staging is cleaned; a re-apply of the same id finds nothing.
	if entries, _ := os.ReadDir(filepath.Join(root, ".llm-wiki", "staging")); len(entries) != 0 {
		t.Errorf("staging not cleaned after apply: %v", entries)
	}
	if _, err := Apply(root, txnID, false); !errors.Is(err, txn.ErrTxnNotFound) {
		t.Errorf("re-apply err = %v, want ErrTxnNotFound", err)
	}
}

// Apply after an out-of-band edit to the target is a stale-plan rejection: the
// tree is left bit-identical and staging is discarded (criterion 13).
func TestApplyStalePlanRejectedZeroMutation(t *testing.T) {
	root := t.TempDir()
	live := writePage(t, root, "wiki/keep.md", validPage)

	// Plan an edit over the existing page.
	edited := validPage + "\nAdded paragraph.\n"
	res, err := Plan(root, "wiki/keep.md", []byte(edited), yamladapter.New())
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if res.NoOp {
		t.Fatalf("expected a real change, got a no-op")
	}

	// Change the live target out-of-band after the plan was staged.
	staleContent := validPage + "\nOut-of-band change.\n"
	if err := os.WriteFile(live, []byte(staleContent), 0o644); err != nil {
		t.Fatalf("out-of-band write: %v", err)
	}

	if _, err := Apply(root, res.TxnID, false); !errors.Is(err, txn.ErrStale) {
		t.Fatalf("Apply over stale target err = %v, want ErrStale", err)
	}

	// Tree bit-identical to the out-of-band content: zero mutation by apply.
	got, err := os.ReadFile(live)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(got) != staleContent {
		t.Errorf("apply mutated a stale target:\n got=%q\nwant=%q", got, staleContent)
	}
	// The plan is retained (a durable preview): re-applying a stale plan rejects
	// identically rather than silently vanishing.
	if _, err := Apply(root, res.TxnID, false); !errors.Is(err, txn.ErrStale) {
		t.Errorf("re-apply of a stale plan err = %v, want ErrStale (idempotent)", err)
	}
}

// A page plan that is staged but abandoned (never applied) is driven to a
// terminal state by the ADR-006 recovery scan: the preview is a staged-but-never-
// committed transaction, so Recover classifies it aborted-clean, removes staging,
// and leaves the tree untouched — after which its id no longer applies.
func TestApplyAbandonedPlanIsRecovered(t *testing.T) {
	root := t.TempDir()
	txnID := stageNewPagePlan(t, root, "wiki/abandoned.md", validPage)

	rep, err := txn.Recover(root)
	if err != nil {
		t.Fatalf("Recover: %v", err)
	}
	found := false
	for _, r := range rep.Transactions {
		if r.ID == txnID {
			found = true
			if r.Outcome != txn.AbortedClean {
				t.Errorf("recover outcome = %q, want aborted-clean", r.Outcome)
			}
		}
	}
	if !found {
		t.Fatalf("recover did not visit txn %s: %+v", txnID, rep.Transactions)
	}
	// Zero mutation: the planned page was never created and staging is cleaned.
	if _, statErr := os.Stat(filepath.Join(root, "wiki", "abandoned.md")); !os.IsNotExist(statErr) {
		t.Errorf("recover left the planned page behind; stat err = %v", statErr)
	}
	if entries, _ := os.ReadDir(filepath.Join(root, ".llm-wiki", "staging")); len(entries) != 0 {
		t.Errorf("recover left staging behind: %v", entries)
	}
	// The recovered plan's id no longer applies.
	if _, err := Apply(root, txnID, false); !errors.Is(err, txn.ErrTxnNotFound) {
		t.Errorf("apply after recover err = %v, want ErrTxnNotFound", err)
	}
}

// Apply of an id with no staged plan is ErrTxnNotFound.
func TestApplyUnknownTxnNotFound(t *testing.T) {
	root := t.TempDir()
	if _, err := Apply(root, "0123456789abcdef", false); !errors.Is(err, txn.ErrTxnNotFound) {
		t.Fatalf("Apply unknown id err = %v, want ErrTxnNotFound", err)
	}
}

// An un-granted approval requirement recorded against a staged plan makes apply
// refuse: nothing is committed and the plan is retained so a granted re-run can
// proceed. Approving commits it. This is the generic ADR-003 plumbing; the
// citation-loss trigger that writes the record is later Phase 3 work (#37).
func TestApplyApprovalRequiredRefusesThenGrants(t *testing.T) {
	root := t.TempDir()
	txnID := stageNewPagePlan(t, root, "wiki/gated.md", validPage)

	// Simulate a plan-time approval trigger by planting the sidecar #37 will write.
	rec := ApprovalRecord{Version: approvalVersion, Reason: "citation loss", Paths: []string{"wiki/gated.md"}}
	data, err := json.Marshal(rec)
	if err != nil {
		t.Fatalf("marshal approval: %v", err)
	}
	sidecar := filepath.Join(root, ".llm-wiki", "staging", txnID, ApprovalFileName)
	if err := os.WriteFile(sidecar, data, 0o644); err != nil {
		t.Fatalf("plant approval sidecar: %v", err)
	}

	// Un-granted apply refuses with zero mutation and retains the plan.
	res, err := Apply(root, txnID, false)
	if err != nil {
		t.Fatalf("Apply (unapproved): %v", err)
	}
	if res.ApprovalRequired == nil || res.ApprovalRequired.Reason != "citation loss" {
		t.Fatalf("expected an approval requirement, got %+v", res.ApprovalRequired)
	}
	if len(res.AppliedPaths) != 0 {
		t.Errorf("unapproved apply committed paths: %v", res.AppliedPaths)
	}
	if _, statErr := os.Stat(filepath.Join(root, "wiki", "gated.md")); !os.IsNotExist(statErr) {
		t.Errorf("unapproved apply created the target; stat err = %v", statErr)
	}
	if _, statErr := os.Stat(sidecar); statErr != nil {
		t.Errorf("unapproved apply must retain the staged plan: %v", statErr)
	}

	// Granting approval commits the page.
	granted, err := Apply(root, txnID, true)
	if err != nil {
		t.Fatalf("Apply (approved): %v", err)
	}
	if granted.ApprovalRequired != nil {
		t.Fatalf("granted apply still reported approval: %+v", granted.ApprovalRequired)
	}
	if _, statErr := os.Stat(filepath.Join(root, "wiki", "gated.md")); statErr != nil {
		t.Errorf("approved apply did not commit the page: %v", statErr)
	}
}
