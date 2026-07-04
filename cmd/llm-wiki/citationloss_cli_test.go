package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/olivermorgan2/llm-wiki-kit/internal/contract"
	"github.com/olivermorgan2/llm-wiki-kit/internal/validate"
)

// citedLivePage cites one URL inside its evidence context; dropCitedPage is an
// edit that removes it. Both carry valid frontmatter so the plan captures a
// clean base and the loss diff runs.
const citedLivePage = "---\ntitle: Cited\ntype: concept\ncustom_field: keep\n---\n\n## Evidence\n\nSee [src](https://example.com/p).\n"
const dropCitedPage = "---\ntitle: Cited\ntype: concept\ncustom_field: keep\n---\n\n## Evidence\n\nCitation removed.\n"

// planDroppingCitation seeds a cited page, plans an edit that drops the citation
// with LLM_WIKI_EVIDENCE_SECTIONS=Evidence set, and returns the plan envelope and
// staging transaction id.
func planDroppingCitation(t *testing.T, dir string) (contract.Envelope, string) {
	t.Helper()
	writeFile(t, filepath.Join(dir, "wiki", "cited.md"), citedLivePage)
	cf := writeContentFile(t, dropCitedPage)
	stdout, _, code := exec(t, "page", "plan", "--root", dir, "wiki/cited.md", "--content", cf, "--json")
	if code != int(contract.ExitSuccess) {
		t.Fatalf("page plan exit = %d, want 0 (the plan itself succeeds)\n%s", code, stdout)
	}
	env := decodeEnvelope(t, stdout)
	if env.Plan == nil || env.Plan.Transaction == "" {
		t.Fatalf("loss plan produced no staging transaction: %+v", env.Plan)
	}
	return env, env.Plan.Transaction
}

// (19, criterion 1) A page plan that drops a cited source succeeds (exit 0) but
// emits a core-citation-loss finding (warning, ruleset profile) and sets the
// approval requirement naming the page.
func TestCLICitationLossPlanEmitsFindingAndApproval(t *testing.T) {
	t.Setenv("LLM_WIKI_EVIDENCE_SECTIONS", "Evidence")
	dir := initBundle(t)
	env, _ := planDroppingCitation(t, dir)

	var loss *contract.Finding
	for i := range env.Findings {
		if env.Findings[i].Code == validate.CodeCitationLoss {
			loss = &env.Findings[i]
		}
	}
	if loss == nil {
		t.Fatalf("no %s finding in plan envelope: %+v", validate.CodeCitationLoss, env.Findings)
	}
	if loss.Ruleset != contract.RulesetProfile || loss.Severity != contract.SeverityWarning {
		t.Errorf("loss finding ruleset/severity = %s/%s, want profile/warning", loss.Ruleset, loss.Severity)
	}
	if !strings.Contains(loss.Message, "https://example.com/p") {
		t.Errorf("loss finding should name the dropped target: %q", loss.Message)
	}
	if env.Approval == nil || !env.Approval.Required {
		t.Fatalf("plan must set an approval requirement: %+v", env.Approval)
	}
	if len(env.Approval.Paths) != 1 || env.Approval.Paths[0] != "wiki/cited.md" {
		t.Errorf("approval paths = %v, want [wiki/cited.md]", env.Approval.Paths)
	}
}

// (20-21, criteria 2-3) The staged loss plan makes an unapproved apply refuse
// with exit 3 and zero mutation; --approve commits it.
func TestCLICitationLossApplyRefusesThenApprove(t *testing.T) {
	t.Setenv("LLM_WIKI_EVIDENCE_SECTIONS", "Evidence")
	dir := initBundle(t)
	_, txnID := planDroppingCitation(t, dir)
	live := filepath.Join(dir, "wiki", "cited.md")

	// (20) Unapproved apply refuses: exit 3, approval-required, file unchanged.
	stdout, _, code := exec(t, "page", "apply", "--root", dir, txnID, "--json")
	if code != int(contract.ExitApprovalRequired) {
		t.Fatalf("unapproved apply exit = %d, want 3\n%s", code, stdout)
	}
	env := decodeEnvelope(t, stdout)
	if env.Status != contract.StatusApprovalRequired {
		t.Errorf("status = %q, want approval-required", env.Status)
	}
	if env.Approval == nil || !env.Approval.Required {
		t.Errorf("refusal must carry the approval field: %+v", env.Approval)
	}
	if got, _ := os.ReadFile(live); string(got) != citedLivePage {
		t.Errorf("unapproved apply mutated the live page:\n%s", got)
	}

	// (21) Approving commits: exit 0, apply payload, citation gone from disk.
	grantOut, _, gcode := exec(t, "page", "apply", "--root", dir, txnID, "--approve", "--json")
	if gcode != int(contract.ExitSuccess) {
		t.Fatalf("approved apply exit = %d, want 0\n%s", gcode, grantOut)
	}
	genv := decodeEnvelope(t, grantOut)
	if genv.Apply == nil || len(genv.Apply.Committed) != 1 || genv.Apply.Committed[0] != "wiki/cited.md" {
		t.Errorf("approved apply payload = %+v, want committed [wiki/cited.md]", genv.Apply)
	}
	got, err := os.ReadFile(live)
	if err != nil {
		t.Fatalf("read committed page: %v", err)
	}
	if strings.Contains(string(got), "https://example.com/p") {
		t.Errorf("approved apply did not drop the citation:\n%s", got)
	}
}

// (22, criterion 4) Moving a citation within the same evidence context is not a
// loss: no finding, no approval, and an unapproved apply commits.
func TestCLICitationMoveWithinContextNoGate(t *testing.T) {
	t.Setenv("LLM_WIKI_EVIDENCE_SECTIONS", "Evidence")
	dir := initBundle(t)
	writeFile(t, filepath.Join(dir, "wiki", "moved.md"), citedLivePage)

	// Same citation, different paragraph of the same evidence context.
	moved := "---\ntitle: Cited\ntype: concept\ncustom_field: keep\n---\n\n## Evidence\n\nMoved below.\n\nNow see [src](https://example.com/p).\n"
	cf := writeContentFile(t, moved)
	stdout, _, code := exec(t, "page", "plan", "--root", dir, "wiki/moved.md", "--content", cf, "--json")
	if code != int(contract.ExitSuccess) {
		t.Fatalf("page plan exit = %d, want 0\n%s", code, stdout)
	}
	env := decodeEnvelope(t, stdout)
	for _, f := range env.Findings {
		if f.Code == validate.CodeCitationLoss {
			t.Errorf("same-context move raised a citation-loss finding: %+v", f)
		}
	}
	if env.Approval != nil {
		t.Errorf("same-context move set an approval requirement: %+v", env.Approval)
	}
	// The unapproved apply commits (no gate).
	_, _, acode := exec(t, "page", "apply", "--root", dir, env.Plan.Transaction, "--json")
	if acode != int(contract.ExitSuccess) {
		t.Errorf("unapproved apply of a non-loss plan exit = %d, want 0", acode)
	}
}

// (23) Env-unset regression: the same citation-dropping edit raises no finding
// and no approval when LLM_WIKI_EVIDENCE_SECTIONS is unset (today's behavior).
func TestCLICitationLossDormantWhenEnvUnset(t *testing.T) {
	t.Setenv("LLM_WIKI_EVIDENCE_SECTIONS", "")
	dir := initBundle(t)
	writeFile(t, filepath.Join(dir, "wiki", "cited.md"), citedLivePage)
	cf := writeContentFile(t, dropCitedPage)

	stdout, _, code := exec(t, "page", "plan", "--root", dir, "wiki/cited.md", "--content", cf, "--json")
	if code != int(contract.ExitSuccess) {
		t.Fatalf("page plan exit = %d, want 0\n%s", code, stdout)
	}
	env := decodeEnvelope(t, stdout)
	for _, f := range env.Findings {
		if f.Code == validate.CodeCitationLoss {
			t.Errorf("no evidence sections set, but a loss finding fired: %+v", f)
		}
	}
	if env.Approval != nil {
		t.Errorf("no evidence sections set, but an approval requirement fired: %+v", env.Approval)
	}
	// The plan applies cleanly without approval.
	_, _, acode := exec(t, "page", "apply", "--root", dir, env.Plan.Transaction, "--json")
	if acode != int(contract.ExitSuccess) {
		t.Errorf("unapproved apply exit = %d, want 0 (no gate when env unset)", acode)
	}
}

// (24) Wiring smoke test: with the env set, validate and page inspect both
// surface the core-citation-* findings for a page whose evidence context holds a
// malformed citation target — the same evidence-section channel the plan uses.
func TestCLIEvidenceSectionsWiredIntoValidateAndInspect(t *testing.T) {
	t.Setenv("LLM_WIKI_EVIDENCE_SECTIONS", "Evidence")
	dir := initBundle(t)
	malformed := "---\ntitle: Bad\ntype: concept\ncustom_field: keep\n---\n\n## Evidence\n\nSee [x](mailto:nobody@example.com).\n"
	writeFile(t, filepath.Join(dir, "wiki", "bad.md"), malformed)

	hasCitationFinding := func(env contract.Envelope) bool {
		for _, f := range env.Findings {
			if strings.HasPrefix(f.Code, "core-citation-") {
				return true
			}
		}
		return false
	}

	vout, _, _ := exec(t, "validate", filepath.Join(dir, "wiki"), "--json")
	if !hasCitationFinding(decodeEnvelope(t, vout)) {
		t.Errorf("validate did not surface a core-citation-* finding with the env set:\n%s", vout)
	}

	iout, _, _ := exec(t, "page", "inspect", "--root", dir, "wiki/bad.md", "--json")
	if !hasCitationFinding(decodeEnvelope(t, iout)) {
		t.Errorf("page inspect did not surface a core-citation-* finding with the env set:\n%s", iout)
	}
}
