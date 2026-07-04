package plan

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/olivermorgan2/llm-wiki-kit/internal/fsafe"
	"github.com/olivermorgan2/llm-wiki-kit/internal/validate"
	"github.com/olivermorgan2/llm-wiki-kit/internal/yamladapter"
)

// eviOpts designates the "Evidence" evidence context, the fixture the
// citation-loss plan tests edit.
var eviOpts = validate.Options{EvidenceSections: []string{"Evidence"}}

// citedPage is a canonical page citing one URL inside its evidence context, so
// re-planning its exact bytes is a no-op and an edit that drops the link is a
// real, gate-triggering change.
const citedPage = "---\ntitle: Cited\ntype: concept\ncustom_field: keep\n---\n\n## Evidence\n\nSee [src](https://example.com/p).\n"

// dropCitation is an edit of citedPage that removes the cited URL from evidence.
const dropCitation = "---\ntitle: Cited\ntype: concept\ncustom_field: keep\n---\n\n## Evidence\n\nCitation removed.\n"

func sidecarPath(root, txnID string) string {
	return filepath.Join(root, fsafe.StagingDir, "staging", txnID, ApprovalFileName)
}

// (13) A plan that drops an existing evidence citation records the loss and
// writes the approval sidecar (version 1, reason naming the target, path == rel)
// into the staged transaction.
func TestPlanCitationLossWritesApprovalSidecar(t *testing.T) {
	root := t.TempDir()
	writePage(t, root, "wiki/cited.md", canonicalPage(citedPage))

	res, err := Plan(root, "wiki/cited.md", []byte(dropCitation), yamladapter.New(), eviOpts)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if len(res.LostCitations) != 1 || res.LostCitations[0] != "https://example.com/p" {
		t.Fatalf("LostCitations = %v, want [https://example.com/p]", res.LostCitations)
	}

	data, err := os.ReadFile(sidecarPath(root, res.TxnID))
	if err != nil {
		t.Fatalf("read approval sidecar: %v", err)
	}
	var rec ApprovalRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		t.Fatalf("decode approval sidecar: %v", err)
	}
	if rec.Version != approvalVersion {
		t.Errorf("sidecar version = %d, want %d", rec.Version, approvalVersion)
	}
	if !strings.Contains(rec.Reason, "https://example.com/p") {
		t.Errorf("sidecar reason = %q, want it to name the lost target", rec.Reason)
	}
	if len(rec.Paths) != 1 || rec.Paths[0] != "wiki/cited.md" {
		t.Errorf("sidecar paths = %v, want [wiki/cited.md]", rec.Paths)
	}
}

// (14) End-to-end: the loss plan's own sidecar makes an unapproved apply refuse
// with the live file byte-unchanged and the plan retained; approving commits and
// cleans staging.
func TestPlanCitationLossApplyRefusesThenGrants(t *testing.T) {
	root := t.TempDir()
	existing := canonicalPage(citedPage)
	writePage(t, root, "wiki/cited.md", existing)

	res, err := Plan(root, "wiki/cited.md", []byte(dropCitation), yamladapter.New(), eviOpts)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if res.LostCitations == nil {
		t.Fatal("expected a citation loss to be recorded")
	}

	// Unapproved apply refuses: nothing committed, tree unchanged, plan retained.
	refused, err := Apply(root, res.TxnID, false)
	if err != nil {
		t.Fatalf("Apply(unapproved): %v", err)
	}
	if refused.ApprovalRequired == nil {
		t.Fatal("unapproved apply of a loss plan must refuse with an approval requirement")
	}
	live, err := os.ReadFile(filepath.Join(root, "wiki", "cited.md"))
	if err != nil {
		t.Fatalf("read live page: %v", err)
	}
	if string(live) != existing {
		t.Errorf("unapproved apply mutated the live page:\n%s", live)
	}
	if _, statErr := os.Stat(sidecarPath(root, res.TxnID)); statErr != nil {
		t.Errorf("unapproved apply must retain the staged plan: %v", statErr)
	}

	// Approved apply commits the edit and cleans staging.
	granted, err := Apply(root, res.TxnID, true)
	if err != nil {
		t.Fatalf("Apply(approved): %v", err)
	}
	if granted.ApprovalRequired != nil {
		t.Fatalf("approved apply still gated: %+v", granted.ApprovalRequired)
	}
	after, err := os.ReadFile(filepath.Join(root, "wiki", "cited.md"))
	if err != nil {
		t.Fatalf("read committed page: %v", err)
	}
	if string(after) != string(res.Staged) {
		t.Errorf("approved apply did not write the staged content:\n%s", after)
	}
	if n := stagingDirCount(t, root); n != 0 {
		t.Errorf("staging dir count = %d after approved apply, want 0", n)
	}
}

// (15) A plan that keeps the citation (edits only the title) raises no gate: no
// sidecar, nil LostCitations, and an unapproved apply commits.
func TestPlanNoCitationLossNoSidecar(t *testing.T) {
	root := t.TempDir()
	writePage(t, root, "wiki/cited.md", canonicalPage(citedPage))

	keep := "---\ntitle: Renamed\ntype: concept\ncustom_field: keep\n---\n\n## Evidence\n\nSee [src](https://example.com/p).\n"
	res, err := Plan(root, "wiki/cited.md", []byte(keep), yamladapter.New(), eviOpts)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if res.LostCitations != nil {
		t.Errorf("LostCitations = %v, want nil (citation kept)", res.LostCitations)
	}
	if _, statErr := os.Stat(sidecarPath(root, res.TxnID)); !os.IsNotExist(statErr) {
		t.Errorf("no-loss plan wrote a sidecar; stat err = %v", statErr)
	}
	if refused, err := Apply(root, res.TxnID, false); err != nil || refused.ApprovalRequired != nil {
		t.Errorf("unapproved apply of a no-loss plan should commit; err=%v approval=%+v", err, refused.ApprovalRequired)
	}
}

// (16) A new page has no existing citation to lose, and a no-op stages nothing —
// neither gates, even with evidence sections designated.
func TestPlanCitationLossNewPageAndNoOp(t *testing.T) {
	root := t.TempDir()

	// New page: BaseAbsent, so no existing citation can be lost.
	res, err := Plan(root, "wiki/fresh.md", []byte(citedPage), yamladapter.New(), eviOpts)
	if err != nil {
		t.Fatalf("Plan(new): %v", err)
	}
	if res.LostCitations != nil {
		t.Errorf("new-page LostCitations = %v, want nil", res.LostCitations)
	}
	if _, statErr := os.Stat(sidecarPath(root, res.TxnID)); !os.IsNotExist(statErr) {
		t.Errorf("new-page plan wrote a sidecar; stat err = %v", statErr)
	}

	// No-op: identical content stages nothing and cannot lose a citation.
	existing := canonicalPage(citedPage)
	writePage(t, root, "wiki/same.md", existing)
	noop, err := Plan(root, "wiki/same.md", []byte(existing), yamladapter.New(), eviOpts)
	if err != nil {
		t.Fatalf("Plan(no-op): %v", err)
	}
	if !noop.NoOp {
		t.Errorf("NoOp = false, want true for identical content")
	}
	if noop.LostCitations != nil {
		t.Errorf("no-op LostCitations = %v, want nil", noop.LostCitations)
	}
}

// (17) When either side's frontmatter fails to split, loss detection is skipped
// (the ADR-004 parse-failure gate is inherited): an existing page with
// unterminated frontmatter never gates even when its evidence citation vanishes.
func TestPlanCitationLossSkippedOnParseFailure(t *testing.T) {
	root := t.TempDir()
	// Frontmatter never closes → splitFrontmatter fails on the source side.
	broken := "---\ntitle: Broken\ntype: concept\n\n## Evidence\n\nSee [src](https://example.com/p).\n"
	writePage(t, root, "wiki/broken.md", broken)

	res, err := Plan(root, "wiki/broken.md", []byte(dropCitation), yamladapter.New(), eviOpts)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if res.LostCitations != nil {
		t.Errorf("parse-failure should skip loss detection, got %v", res.LostCitations)
	}
	if _, statErr := os.Stat(sidecarPath(root, res.TxnID)); !os.IsNotExist(statErr) {
		t.Errorf("parse-failure plan wrote a sidecar; stat err = %v", statErr)
	}
}
