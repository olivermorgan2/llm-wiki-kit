package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/olivermorgan2/llm-wiki-kit/internal/contract"
)

// Phase 4 (Academic-research profile) acceptance corpus — the named,
// criterion-traceable gate evidence for the Phase 4 exit gate
// (design/build-out-plan.md §"Phase 4"; addendum 003). It sits alongside the
// Phase 2/3 corpora and carries the same stable TestAcceptance prefix so the CI
// step name-selects all phases in one cache-defeating run
// (`go test ./cmd/llm-wiki -run '^TestAcceptance' -count=1 -v`) and prints one
// legible PASS line per criterion per platform. It proves:
//
//	Criterion 4 (academic-research) — init with the academic-research profile,
//	  author each of the five profiled page types (source, claim, method,
//	  question, synthesis) through the staged page plan/apply workflow, and the
//	  bundle validates clean under that profile (ADR-006 staged mutation +
//	  ADR-007/ADR-010 profile rules).
//	Addendum-003 gate (non-vacuous) — negative controls: an invalid page of a
//	  profiled type trips exactly the addendum-003 rule for it, proving the rules
//	  actually fire end-to-end through the CLI.
//	Evidence-obligation journey — a supported claim with no citation is flagged
//	  (profile-citation-required); adding a resolvable citation clears it
//	  (ADR-008 carry-in №2 / addendum 003).
//
// Journeys go through the same in-package harness (exec, decodeEnvelope,
// writeContentFile) as the Phase 3 corpus; pages are authored under wiki/ with
// bundle-root-relative citations. Windows runs this leg too, so every path uses
// slash-form literals and disk touches go through filepath.FromSlash/Join.

// initAcademicBundle scaffolds a fresh academic-research bundle and returns its
// root. The scaffold itself validates clean (proved in scaffold + cli tests); the
// journeys author on top of it.
func initAcademicBundle(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if _, _, code := exec(t, "init", dir, "--profile", "academic-research", "--json"); code != int(contract.ExitSuccess) {
		t.Fatalf("init --profile academic-research exit = %d, want 0", code)
	}
	return dir
}

// authorPage drives one page through staged page plan → page apply and fails the
// test on any non-success. It is the managed-mutation path a researcher uses.
func authorPage(t *testing.T, dir, path, content string) {
	t.Helper()
	cf := writeContentFile(t, content)
	pout, _, pcode := exec(t, "page", "plan", "--root", dir, path, "--content", cf, "--json")
	if pcode != int(contract.ExitSuccess) {
		t.Fatalf("plan %q exit = %d, want 0\n%s", path, pcode, pout)
	}
	penv := decodeEnvelope(t, pout)
	if penv.Plan == nil || penv.Plan.NoOp {
		t.Fatalf("authoring %q should be a real staged change, got %+v", path, penv.Plan)
	}
	if _, _, acode := exec(t, "page", "apply", "--root", dir, penv.Plan.Transaction, "--json"); acode != int(contract.ExitSuccess) {
		t.Fatalf("apply %q exit = %d, want 0", path, acode)
	}
}

// hasFindingCode reports whether findings contains a finding with the given code.
func hasFindingCode(findings []contract.Finding, code string) bool {
	for _, f := range findings {
		if f.Code == code {
			return true
		}
	}
	return false
}

// Valid authored pages, one per profiled type. The claim cites the source via a
// bundle-root-relative path, so the source is authored first.
const (
	acSourceDraft = "---\n" +
		"type: source\ntitle: A Cited Source\ndescription: A source the claim cites.\n" +
		"timestamp: 2026-01-01\ntags: [nlp]\naliases: [cited-source]\nresource: https://example.com/s\n" +
		"authors: [Author One]\nsource_type: paper\ndoi: 10.1/x\n---\n# A Cited Source\n\nA summary.\n"

	acMethodDraft = "---\n" +
		"type: method\ntitle: An Authored Method\ndescription: A method with its trade-offs.\n" +
		"timestamp: 2026-01-01\ntags: [method]\naliases: [authored-method]\nresource: https://example.com/m\n---\n" +
		"# An Authored Method\n\n## Assumptions\n\nA.\n\n## Strengths\n\nB.\n\n## Limitations\n\nC.\n"

	acQuestionDraft = "---\n" +
		"type: question\ntitle: An Open Question\ndescription: Workflow state, not evidence.\n" +
		"timestamp: 2026-01-01\ntags: [open]\naliases: [open-question]\nresource: https://example.com/q\n" +
		"status: open\n---\n# An Open Question\n\n## Evidence gap\n\nNot yet studied.\n"

	acSynthesisDraft = "---\n" +
		"type: synthesis\ntitle: An Authored Synthesis\ndescription: A cross-source integration.\n" +
		"timestamp: 2026-01-01\ntags: [survey]\naliases: [authored-synthesis]\nresource: https://example.com/y\n---\n" +
		"# An Authored Synthesis\n\n## Scope\n\nS.\n\n## Findings\n\nF.\n\n## Agreement\n\nA.\n\n## Disagreement\n\nD.\n\n## Evidence gaps\n\nG.\n"

	acClaimDraft = "---\n" +
		"type: claim\ntitle: A Supported Cited Claim\ndescription: A claim with cited evidence.\n" +
		"timestamp: 2026-01-01\ntags: [nlp]\naliases: [supported-claim]\nresource: https://example.com/c\n" +
		"confidence: high\nassessment: supported\n---\n# A Supported Cited Claim\n\n" +
		"## Evidence\n\nSee [the source](wiki/source.md).\n\n## Counterevidence\n\nNone found.\n\n## Assessment\n\nSupported.\n"
)

// TestAcceptanceCriterion4AcademicResearchAuthorEachProfiledType — criterion 4
// (academic-research): a researcher inits with the profile and authors each of
// the five profiled types through staged plan/apply, and the whole bundle
// validates clean under the profile.
func TestAcceptanceCriterion4AcademicResearchAuthorEachProfiledType(t *testing.T) {
	dir := initAcademicBundle(t)

	// Author in dependency order: the source the claim cites comes first.
	pages := []struct{ path, content string }{
		{"wiki/source.md", acSourceDraft},
		{"wiki/method.md", acMethodDraft},
		{"wiki/question.md", acQuestionDraft},
		{"wiki/synthesis.md", acSynthesisDraft},
		{"wiki/claim.md", acClaimDraft},
	}
	for _, pg := range pages {
		authorPage(t, dir, pg.path, pg.content)
	}

	// Every authored page is committed on disk.
	for _, pg := range pages {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(pg.path))); err != nil {
			t.Errorf("authored page %q missing after apply: %v", pg.path, err)
		}
	}

	// The whole bundle (scaffold + authored pages) validates clean under the
	// academic-research profile (resolved from llm-wiki.yaml).
	vout, _, vcode := exec(t, "validate", dir, "--json")
	venv := decodeEnvelope(t, vout)
	if venv.Status != contract.StatusSuccess {
		t.Errorf("validate status = %q, want success (%+v)", venv.Status, venv.Findings)
	}
	if len(venv.Findings) != 0 {
		t.Errorf("authored academic bundle must validate clean, got %+v", venv.Findings)
	}
	if vcode != int(contract.ExitSuccess) {
		t.Errorf("validate exit = %d, want 0", vcode)
	}
}

// Invalid authored pages for the negative controls — each an otherwise-complete
// page of a profiled type with one deliberate defect.
const (
	acSourceMissingAuthors = "---\n" +
		"type: source\ntitle: A Source Missing Authors\ndescription: Missing the required authors.\n" +
		"timestamp: 2026-01-01\ntags: [nlp]\naliases: [bad-source]\nresource: https://example.com/x\n" +
		"source_type: paper\ndoi: 10.1/x\n---\n# A Source Missing Authors\n\nBody.\n"

	acMethodMissingSection = "---\n" +
		"type: method\ntitle: A Method Missing Limitations\ndescription: Missing a required section.\n" +
		"timestamp: 2026-01-01\ntags: [method]\naliases: [bad-method]\nresource: https://example.com/x\n---\n" +
		"# A Method Missing Limitations\n\n## Assumptions\n\nA.\n\n## Strengths\n\nB.\n"

	acQuestionBadStatus = "---\n" +
		"type: question\ntitle: A Question With a Bad Status\ndescription: status outside its enum.\n" +
		"timestamp: 2026-01-01\ntags: [open]\naliases: [bad-question]\nresource: https://example.com/x\n" +
		"status: maybe\n---\n# A Question With a Bad Status\n\n## Evidence gap\n\nGap.\n"
)

// TestAcceptanceCriterion4AcademicResearchNegativeControls — the addendum-003
// gate is non-vacuous: an invalid page of a profiled type, authored through the
// same plan/apply path, trips exactly its addendum-003 rule at validate. Each
// case uses a fresh bundle so the finding is unambiguous.
func TestAcceptanceCriterion4AcademicResearchNegativeControls(t *testing.T) {
	cases := []struct {
		name, path, content, wantCode string
	}{
		{"source-missing-authors", "wiki/bad-source.md", acSourceMissingAuthors, "profile-required-field"},
		{"method-missing-section", "wiki/bad-method.md", acMethodMissingSection, "profile-required-section"},
		{"question-bad-status", "wiki/bad-question.md", acQuestionBadStatus, "profile-field-enum"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			dir := initAcademicBundle(t)
			authorPage(t, dir, c.path, c.content)

			vout, _, _ := exec(t, "validate", dir, "--json")
			venv := decodeEnvelope(t, vout)
			if !hasFindingCode(venv.Findings, c.wantCode) {
				t.Fatalf("validate should flag %q for %s, got %+v", c.wantCode, c.name, venv.Findings)
			}
			if venv.Status != contract.StatusValidationFailure {
				t.Errorf("an error-severity profile finding must fail validation, status = %q", venv.Status)
			}
		})
	}
}

// A supported claim with no citation in its Evidence section, then the same claim
// edited to cite the resolvable source.
const (
	acClaimNoCitation = "---\n" +
		"type: claim\ntitle: A Supported Uncited Claim\ndescription: Supported but uncited.\n" +
		"timestamp: 2026-01-01\ntags: [nlp]\naliases: [uncited-claim]\nresource: https://example.com/c\n" +
		"confidence: high\nassessment: supported\n---\n# A Supported Uncited Claim\n\n" +
		"## Evidence\n\nNo citation here.\n\n## Counterevidence\n\nNone.\n\n## Assessment\n\nAsserted.\n"

	acClaimNowCited = "---\n" +
		"type: claim\ntitle: A Supported Uncited Claim\ndescription: Supported but uncited.\n" +
		"timestamp: 2026-01-01\ntags: [nlp]\naliases: [uncited-claim]\nresource: https://example.com/c\n" +
		"confidence: high\nassessment: supported\n---\n# A Supported Uncited Claim\n\n" +
		"## Evidence\n\nSee [the source](wiki/source.md).\n\n## Counterevidence\n\nNone.\n\n## Assessment\n\nAsserted.\n"
)

// TestAcceptanceCriterion4AcademicResearchEvidenceObligationJourney — the
// evidence-obligation journey for a claim (ADR-008 carry-in №2 / addendum 003):
// a supported claim with no citation is flagged profile-citation-required;
// editing it to cite a resolvable source clears the obligation.
func TestAcceptanceCriterion4AcademicResearchEvidenceObligationJourney(t *testing.T) {
	dir := initAcademicBundle(t)
	authorPage(t, dir, "wiki/source.md", acSourceDraft)

	// Supported, uncited → the obligation fires.
	authorPage(t, dir, "wiki/claim.md", acClaimNoCitation)
	vout, _, _ := exec(t, "validate", dir, "--json")
	venv := decodeEnvelope(t, vout)
	if !hasFindingCode(venv.Findings, "profile-citation-required") {
		t.Fatalf("a supported uncited claim must be flagged profile-citation-required, got %+v", venv.Findings)
	}

	// Edit the same claim to cite the resolvable source → obligation cleared.
	authorPage(t, dir, "wiki/claim.md", acClaimNowCited)
	vout, _, vcode := exec(t, "validate", dir, "--json")
	venv = decodeEnvelope(t, vout)
	if hasFindingCode(venv.Findings, "profile-citation-required") {
		t.Errorf("a cited supported claim must clear the obligation, got %+v", venv.Findings)
	}
	if venv.Status != contract.StatusSuccess || len(venv.Findings) != 0 {
		t.Errorf("the cited bundle must validate clean, status=%q findings=%+v", venv.Status, venv.Findings)
	}
	if vcode != int(contract.ExitSuccess) {
		t.Errorf("validate exit = %d, want 0", vcode)
	}
}
