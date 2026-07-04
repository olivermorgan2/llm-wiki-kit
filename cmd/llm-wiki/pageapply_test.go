package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/olivermorgan2/llm-wiki-kit/internal/contract"
	"github.com/olivermorgan2/llm-wiki-kit/internal/plan"
)

// planPage runs `page plan` for a new page and returns the staging transaction
// id its apply consumes.
func planPage(t *testing.T, dir, rel, content string) string {
	t.Helper()
	cf := writeContentFile(t, content)
	stdout, _, code := exec(t, "page", "plan", "--root", dir, rel, "--content", cf, "--json")
	if code != int(contract.ExitSuccess) {
		t.Fatalf("page plan exit = %d, want 0\n%s", code, stdout)
	}
	env := decodeEnvelope(t, stdout)
	if env.Plan == nil || env.Plan.Transaction == "" {
		t.Fatalf("page plan produced no staging transaction: %+v", env.Plan)
	}
	return env.Plan.Transaction
}

const applyPage = "---\ntitle: Applied\ntype: concept\ncustom_field: keep\n---\n\n# Applied\n\nBody.\n"

// Happy path: apply a staged plan → exit 0, operation "page apply", an apply
// payload naming the transaction and committed file, the target in
// affectedPaths, the live file written, and staging cleaned (criterion 12).
func TestPageApplyCommitsPlan(t *testing.T) {
	dir := initBundle(t)
	txnID := planPage(t, dir, "wiki/applied.md", applyPage)

	stdout, _, code := exec(t, "page", "apply", "--root", dir, txnID, "--json")
	if code != int(contract.ExitSuccess) {
		t.Fatalf("page apply exit = %d, want 0\n%s", code, stdout)
	}
	env := decodeEnvelope(t, stdout)
	if env.Operation != "page apply" {
		t.Errorf("operation = %q, want \"page apply\"", env.Operation)
	}
	if env.Status != contract.StatusSuccess {
		t.Errorf("status = %q, want success", env.Status)
	}
	if env.Apply == nil {
		t.Fatal("apply payload missing")
	}
	if env.Apply.Transaction != txnID {
		t.Errorf("apply.transaction = %q, want %q", env.Apply.Transaction, txnID)
	}
	if len(env.Apply.Committed) != 1 || env.Apply.Committed[0] != "wiki/applied.md" {
		t.Errorf("apply.committed = %v, want [wiki/applied.md]", env.Apply.Committed)
	}
	if len(env.AffectedPaths) != 1 || env.AffectedPaths[0] != "wiki/applied.md" {
		t.Errorf("affectedPaths = %v, want [wiki/applied.md]", env.AffectedPaths)
	}

	// The live page now exists and staging is gone.
	if _, err := os.Stat(filepath.Join(dir, "wiki", "applied.md")); err != nil {
		t.Errorf("committed page missing: %v", err)
	}
	if entries, _ := os.ReadDir(filepath.Join(dir, ".llm-wiki", "staging")); len(entries) != 0 {
		t.Errorf("staging not cleaned after apply: %v", entries)
	}
}

// Stale-plan rejection (criterion 13): an out-of-band edit to the target between
// plan and apply causes exit 4 (invalid-invocation), the six-field envelope, and
// a bit-identical tree.
func TestPageApplyStalePlanRejected(t *testing.T) {
	dir := initBundle(t)
	live := filepath.Join(dir, "wiki", "keep.md")
	if err := os.WriteFile(live, []byte(applyPage), 0o644); err != nil {
		t.Fatalf("seed live page: %v", err)
	}
	txnID := planPage(t, dir, "wiki/keep.md", applyPage+"\nEdited.\n")

	// Change the live target out of band after planning.
	stale := applyPage + "\nOut-of-band.\n"
	if err := os.WriteFile(live, []byte(stale), 0o644); err != nil {
		t.Fatalf("out-of-band write: %v", err)
	}

	stdout, stderr, code := exec(t, "page", "apply", "--root", dir, txnID, "--json")
	if code != int(contract.ExitInvalidInvocation) {
		t.Fatalf("stale apply exit = %d, want 4\n%s", code, stdout)
	}
	env := decodeEnvelope(t, stdout)
	if env.Status != contract.StatusInvalidInvocation {
		t.Errorf("status = %q, want invalid-invocation", env.Status)
	}
	if env.Apply != nil {
		t.Errorf("stale rejection must not carry an apply payload: %+v", env.Apply)
	}
	var generic map[string]json.RawMessage
	if err := json.Unmarshal([]byte(stdout), &generic); err != nil {
		t.Fatalf("stdout not JSON: %v", err)
	}
	if len(generic) != 6 {
		t.Errorf("stale rejection envelope must be six fields, got %d: %s", len(generic), stdout)
	}

	// Zero mutation: the live page still holds the out-of-band bytes.
	got, err := os.ReadFile(live)
	if err != nil {
		t.Fatalf("read live: %v", err)
	}
	if string(got) != stale {
		t.Errorf("stale apply mutated the target:\n%s", got)
	}
	// The human path reports it as a stale plan, not a generic bad invocation.
	humanOut, humanErr, hcode := exec(t, "page", "apply", "--root", dir, txnID)
	if hcode != int(contract.ExitInvalidInvocation) {
		t.Errorf("human stale apply exit = %d, want 4", hcode)
	}
	if !strings.Contains(humanErr+humanOut, "stale") {
		t.Errorf("human output should explain the plan is stale; got:\n%s%s", humanOut, humanErr)
	}
	_ = stderr
}

// Applying an id with no staged plan is invalid-invocation (exit 4).
func TestPageApplyUnknownTransaction(t *testing.T) {
	dir := initBundle(t)
	_, _, code := exec(t, "page", "apply", "--root", dir, "0123456789abcdef")
	if code != int(contract.ExitInvalidInvocation) {
		t.Errorf("unknown apply exit = %d, want 4", code)
	}
}

// Missing <txn-id> is invalid-invocation (exit 4).
func TestPageApplyMissingArg(t *testing.T) {
	dir := initBundle(t)
	_, _, code := exec(t, "page", "apply", "--root", dir)
	if code != int(contract.ExitInvalidInvocation) {
		t.Errorf("missing-arg apply exit = %d, want 4", code)
	}
}

// Approval plumbing (ADR-003): an un-granted approval requirement recorded on a
// staged plan makes apply refuse with exit 3 and the approval envelope field,
// committing nothing; re-running with --approve grants it and commits.
func TestPageApplyApprovalRequiredThenApprove(t *testing.T) {
	dir := initBundle(t)
	txnID := planPage(t, dir, "wiki/gated.md", applyPage)

	// Plant the approval sidecar #37's citation-loss trigger will write.
	rec := plan.ApprovalRecord{Version: 1, Reason: "citation loss", Paths: []string{"wiki/gated.md"}}
	data, err := json.Marshal(rec)
	if err != nil {
		t.Fatalf("marshal approval: %v", err)
	}
	sidecar := filepath.Join(dir, ".llm-wiki", "staging", txnID, plan.ApprovalFileName)
	if err := os.WriteFile(sidecar, data, 0o644); err != nil {
		t.Fatalf("plant approval sidecar: %v", err)
	}

	// Un-granted apply refuses.
	stdout, _, code := exec(t, "page", "apply", "--root", dir, txnID, "--json")
	if code != int(contract.ExitApprovalRequired) {
		t.Fatalf("unapproved apply exit = %d, want 3\n%s", code, stdout)
	}
	env := decodeEnvelope(t, stdout)
	if env.Status != contract.StatusApprovalRequired {
		t.Errorf("status = %q, want approval-required", env.Status)
	}
	if env.Approval == nil || !env.Approval.Required || env.Approval.Reason != "citation loss" {
		t.Errorf("approval field not set as expected: %+v", env.Approval)
	}
	if env.Apply != nil {
		t.Errorf("refusal must not carry an apply payload: %+v", env.Apply)
	}
	if _, statErr := os.Stat(filepath.Join(dir, "wiki", "gated.md")); !os.IsNotExist(statErr) {
		t.Errorf("unapproved apply created the target; stat err = %v", statErr)
	}

	// Granting with --approve commits.
	grantOut, _, gcode := exec(t, "page", "apply", "--root", dir, txnID, "--approve", "--json")
	if gcode != int(contract.ExitSuccess) {
		t.Fatalf("approved apply exit = %d, want 0\n%s", gcode, grantOut)
	}
	if _, statErr := os.Stat(filepath.Join(dir, "wiki", "gated.md")); statErr != nil {
		t.Errorf("approved apply did not commit the page: %v", statErr)
	}
}

// Human output for a clean apply names the committed page.
func TestPageApplyHumanOutput(t *testing.T) {
	dir := initBundle(t)
	txnID := planPage(t, dir, "wiki/human.md", applyPage)

	stdout, _, code := exec(t, "page", "apply", "--root", dir, txnID)
	if code != int(contract.ExitSuccess) {
		t.Fatalf("apply exit = %d, want 0\n%s", code, stdout)
	}
	if !strings.Contains(stdout, "applied") || !strings.Contains(stdout, "wiki/human.md") {
		t.Errorf("human output should confirm the applied page:\n%s", stdout)
	}
}
