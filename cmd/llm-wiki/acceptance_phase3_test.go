package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/olivermorgan2/llm-wiki-kit/internal/contract"
)

// Phase 3 (Authoring + staged mutation) acceptance corpus — the named,
// criterion-traceable gate evidence for the Phase 3 exit gate
// (design/build-out-plan.md §"Phase 3"). It sits alongside the Phase 2 corpus in
// acceptance_test.go and carries the same stable TestAcceptance prefix so the CI
// step name-selects both phases in one cache-defeating run
// (`go test ./cmd/llm-wiki -run '^TestAcceptance' -count=1 -v`) and prints one
// legible PASS line per criterion per platform. Each test is doc-commented with
// the acceptance criterion and ADR it proves:
//
//	Criterion 6  — unknown frontmatter fields survive a full plan/apply
//	  round-trip into the committed live file (ADR-006 staged mutation over the
//	  OKF loss-free round-trip; the enrichment half is Phase 5).
//	Criterion 9  — every sourced claim carries a citation that resolves fully
//	  offline within the bundle's fixtures (ADR-008 citation resolver).
//	Criterion 10 — re-running an identical committed draft is a no-op: no
//	  duplicate page, empty transaction, byte-identical tree (ADR-006).
//	Criterion 11 — an edit to an existing page is previewed before any live
//	  write; the committed bytes equal exactly what was previewed (ADR-006).
//	Criterion 12 — managed writes reach the tree only through staged plan/apply;
//	  draft/inspect/plan touch only the .llm-wiki/ staging area (ADR-005/006).
//	Criterion 13 — a plan whose target changed out of band is rejected with zero
//	  mutation of the whole tree (ADR-006 base-state binding).
//
// The corpus deliberately overlaps some assertions with the per-command unit
// tests (page_test.go / pageplan_test.go / pageapply_test.go) and the skill-flow
// tests in authoring_flow_test.go — same rationale as the Phase 2 overlap note:
// those give per-command coverage, this gives criterion-mapped, name-selectable
// gate evidence. Each test is kept an end-to-end journey (init → plan → apply →
// verify), not a copy of a unit assertion. It reuses the existing in-package
// harness (exec, decodeEnvelope, initBundle, writeContentFile) and snapshot
// proofs (snapshotAll, changedPaths, allUnder, hashFile) plus the authoringDraft
// fixture from authoring_flow_test.go (kept byte-aligned with
// skills/wiki-authoring/SKILL.md) rather than a checked-in testdata/ fixture:
// scratch bundles are seeded programmatically with \n literals into t.TempDir(),
// which is platform-identical byte-for-byte.
//
// Portability: the Windows leg runs this corpus, so every envelope/diff
// assertion uses slash-form literals and every disk touch goes through
// filepath.FromSlash/Join; snapshot keys are already ToSlash'd (snapshotAll).

// unknownFieldsLivePage is an existing live page carrying two fields outside the
// core schema (`custom_field`, the hyphenated `x-tool-meta`), reusing the shapes
// from pageplan_test.go so the round-trip is checked against the same edit case
// the plan-layer unit test stages — but here through a full apply into the tree.
const unknownFieldsLivePage = "---\ntitle: Old\ntype: concept\ncustom_field: keep-me\nx-tool-meta: 7\n---\n\n# Old\n\nBody.\n"

// unknownFieldsEditDraft edits a known field (title) while retaining both unknown
// fields — the proposed content the author feeds to `page plan --content`.
const unknownFieldsEditDraft = "---\ntitle: New\ntype: concept\ncustom_field: keep-me\nx-tool-meta: 7\n---\n\n# Old\n\nBody.\n"

// unknownFieldsNewDraft is a brand-new page whose draft carries an unknown field,
// exercising the new-page (absent-base) direction of the round-trip.
const unknownFieldsNewDraft = "---\ntitle: Fresh\ntype: concept\ncustom_field: keep\n---\n\n# Fresh\n\nBody.\n"

// TestAcceptanceCriterion6UnknownFieldsRoundTripPlanApply — criterion 6: fields
// outside the core schema survive a full staged plan/apply into the committed
// live file (ADR-006 over the OKF loss-free round-trip). The delta over the
// plan-layer unit test (which inspects the staged postimage) is that this asserts
// the fields on the *committed* bytes after apply, in both the edit and new-page
// directions, and that those committed bytes equal exactly the previewed staged
// content. The enrichment half of criterion 6 is Phase 5 and out of scope here.
func TestAcceptanceCriterion6UnknownFieldsRoundTripPlanApply(t *testing.T) {
	// Edit direction: an existing page with unknown fields, edited on a known field.
	dir := initBundle(t)
	live := filepath.Join(dir, "wiki", "keep.md")
	if err := os.WriteFile(live, []byte(unknownFieldsLivePage), 0o644); err != nil {
		t.Fatalf("seed live page: %v", err)
	}
	cf := writeContentFile(t, unknownFieldsEditDraft)

	pout, _, pcode := exec(t, "page", "plan", "--root", dir, "wiki/keep.md", "--content", cf, "--json")
	if pcode != int(contract.ExitSuccess) {
		t.Fatalf("plan exit = %d, want 0\n%s", pcode, pout)
	}
	penv := decodeEnvelope(t, pout)
	if penv.Plan == nil || penv.Plan.NoOp {
		t.Fatalf("edit should be a real staged change, got %+v", penv.Plan)
	}
	if _, _, acode := exec(t, "page", "apply", "--root", dir, penv.Plan.Transaction, "--json"); acode != int(contract.ExitSuccess) {
		t.Fatalf("apply exit = %d, want 0", acode)
	}

	committed, err := os.ReadFile(live)
	if err != nil {
		t.Fatalf("read committed page: %v", err)
	}
	for _, field := range []string{"custom_field: keep-me", "x-tool-meta: 7", "title: New"} {
		if !strings.Contains(string(committed), field) {
			t.Errorf("committed live file dropped %q after plan/apply:\n%s", field, committed)
		}
	}
	// The committed bytes are exactly what the plan previewed.
	if got := hashFile(t, live); got != penv.Plan.StagedHash {
		t.Errorf("committed page hash = %q, want previewed staged hash %q", got, penv.Plan.StagedHash)
	}

	// New-page direction: a fresh page whose draft carries an unknown field.
	nf := writeContentFile(t, unknownFieldsNewDraft)
	np, _, ncode := exec(t, "page", "plan", "--root", dir, "wiki/fresh.md", "--content", nf, "--json")
	if ncode != int(contract.ExitSuccess) {
		t.Fatalf("new-page plan exit = %d, want 0\n%s", ncode, np)
	}
	nenv := decodeEnvelope(t, np)
	if nenv.Plan == nil || !nenv.Plan.BaseAbsent {
		t.Fatalf("new-page plan should report an absent base: %+v", nenv.Plan)
	}
	if _, _, acode := exec(t, "page", "apply", "--root", dir, nenv.Plan.Transaction, "--json"); acode != int(contract.ExitSuccess) {
		t.Fatalf("new-page apply exit = %d, want 0", acode)
	}
	freshLive := filepath.Join(dir, "wiki", "fresh.md")
	freshBytes, err := os.ReadFile(freshLive)
	if err != nil {
		t.Fatalf("read committed new page: %v", err)
	}
	if !strings.Contains(string(freshBytes), "custom_field: keep") {
		t.Errorf("committed new page dropped its unknown field:\n%s", freshBytes)
	}
	if got := hashFile(t, freshLive); got != nenv.Plan.StagedHash {
		t.Errorf("committed new page hash = %q, want previewed staged hash %q", got, nenv.Plan.StagedHash)
	}
}

// TestAcceptanceCriterion9SourcedClaimCitationsResolveOffline — criterion 9: the
// SKILL.md worked example (authoringDraft) carries one in-bundle citation
// (wiki/index.md) and one https citation; both classes resolve fully offline, so
// inspect-content and a post-apply validate report zero core-citation-* findings
// (ADR-008 resolver; release gate: 100% within fixtures). A negative control
// (an in-bundle citation to a missing target) proves the resolution is not
// vacuous by producing a core-citation-unresolved finding — a warning-severity
// result, so it is asserted on the findings, not the exit code.
func TestAcceptanceCriterion9SourcedClaimCitationsResolveOffline(t *testing.T) {
	t.Setenv("LLM_WIKI_EVIDENCE_SECTIONS", "Evidence")
	dir := initBundle(t)
	cf := writeContentFile(t, authoringDraft)

	// inspect-content: the sourced-claim draft resolves clean, offline.
	iout, _, icode := exec(t, "page", "inspect", "--root", dir, "wiki/photosynthesis.md", "--content", cf, "--json")
	if icode != int(contract.ExitSuccess) {
		t.Fatalf("inspect-content exit = %d, want 0\n%s", icode, iout)
	}
	ienv := decodeEnvelope(t, iout)
	for _, f := range ienv.Findings {
		if strings.HasPrefix(f.Code, "core-citation-") {
			t.Errorf("criterion 9: sourced-claim draft raised a citation finding: %+v", f)
		}
	}

	// plan → apply, then validate the committed bundle: still zero citation findings.
	pout, _, pcode := exec(t, "page", "plan", "--root", dir, "wiki/photosynthesis.md", "--content", cf, "--json")
	if pcode != int(contract.ExitSuccess) {
		t.Fatalf("plan exit = %d, want 0\n%s", pcode, pout)
	}
	penv := decodeEnvelope(t, pout)
	if _, _, acode := exec(t, "page", "apply", "--root", dir, penv.Plan.Transaction, "--json"); acode != int(contract.ExitSuccess) {
		t.Fatalf("apply exit = %d, want 0", acode)
	}
	vout, _, vcode := exec(t, "validate", dir, "--json")
	if vcode != int(contract.ExitSuccess) {
		t.Fatalf("validate after apply exit = %d, want 0\n%s", vcode, vout)
	}
	venv := decodeEnvelope(t, vout)
	for _, f := range venv.Findings {
		if strings.HasPrefix(f.Code, "core-citation-") {
			t.Errorf("criterion 9: committed bundle raised a citation finding: %+v", f)
		}
	}

	// Negative control: an in-bundle citation to a missing target does not resolve,
	// proving the offline resolver is doing real work (core-citation-unresolved,
	// warning severity — assert the finding, not the exit code).
	missing := strings.Replace(authoringDraft, "wiki/index.md", "wiki/nope.md", 1)
	if missing == authoringDraft {
		t.Fatal("negative-control fixture did not diverge from authoringDraft")
	}
	mf := writeContentFile(t, missing)
	mout, _, _ := exec(t, "page", "inspect", "--root", dir, "wiki/photosynthesis.md", "--content", mf, "--json")
	menv := decodeEnvelope(t, mout)
	sawUnresolved := false
	for _, f := range menv.Findings {
		if f.Code == "core-citation-unresolved" {
			sawUnresolved = true
		}
	}
	if !sawUnresolved {
		t.Errorf("negative control: a missing in-bundle citation target must raise core-citation-unresolved, got %+v", menv.Findings)
	}
}

// TestAcceptanceCriterion10RepeatedUnchangedInputIsNoOp — criterion 10: once a
// draft is committed, re-running the identical draft is a no-op — it stages no
// transaction, lists no affected paths, creates no duplicate page, and leaves the
// tree (staging included) byte-identical (ADR-006 fixed-point round-trip).
func TestAcceptanceCriterion10RepeatedUnchangedInputIsNoOp(t *testing.T) {
	t.Setenv("LLM_WIKI_EVIDENCE_SECTIONS", "Evidence")
	dir := initBundle(t)
	cf := writeContentFile(t, authoringDraft)

	// First pass commits the page.
	p1, _, _ := exec(t, "page", "plan", "--root", dir, "wiki/photosynthesis.md", "--content", cf, "--json")
	env1 := decodeEnvelope(t, p1)
	if env1.Plan == nil || env1.Plan.Transaction == "" {
		t.Fatalf("first plan produced no transaction: %+v", env1.Plan)
	}
	if _, _, code := exec(t, "page", "apply", "--root", dir, env1.Plan.Transaction, "--json"); code != int(contract.ExitSuccess) {
		t.Fatalf("first apply exit = %d, want 0", code)
	}
	afterCommit := snapshotAll(t, dir)

	// Second pass with the identical draft is a no-op.
	p2, _, pcode := exec(t, "page", "plan", "--root", dir, "wiki/photosynthesis.md", "--content", cf, "--json")
	if pcode != int(contract.ExitSuccess) {
		t.Fatalf("second plan exit = %d, want 0\n%s", pcode, p2)
	}
	env2 := decodeEnvelope(t, p2)
	if env2.Plan == nil || !env2.Plan.NoOp {
		t.Fatalf("criterion 10: identical re-draft should be a no-op, got %+v", env2.Plan)
	}
	if env2.Plan.Transaction != "" {
		t.Errorf("no-op plan should stage no transaction, got %q", env2.Plan.Transaction)
	}
	if len(env2.AffectedPaths) != 0 {
		t.Errorf("no-op plan should list no affected paths, got %v", env2.AffectedPaths)
	}

	// Exactly one page file, and the whole tree is unchanged by the no-op plan.
	matches, _ := filepath.Glob(filepath.Join(dir, "wiki", "photosynthesis*.md"))
	if len(matches) != 1 {
		t.Errorf("expected exactly one photosynthesis page, found %v", matches)
	}
	if changed := changedPaths(afterCommit, snapshotAll(t, dir)); len(changed) != 0 {
		t.Errorf("no-op plan changed the tree: %v", changed)
	}
}

// TestAcceptanceCriterion11EditPreviewPrecedesApply — criterion 11: an edit to an
// existing page is previewed by plan (a non-empty diff carrying the edited text)
// before any live write; declining to apply leaves the live file untouched and
// touches only staging; applying commits exactly the previewed bytes (ADR-006).
func TestAcceptanceCriterion11EditPreviewPrecedesApply(t *testing.T) {
	t.Setenv("LLM_WIKI_EVIDENCE_SECTIONS", "Evidence")
	dir := initBundle(t)
	live := filepath.Join(dir, "wiki", "photosynthesis.md")

	// Commit the base page.
	cf := writeContentFile(t, authoringDraft)
	p1, _, _ := exec(t, "page", "plan", "--root", dir, "wiki/photosynthesis.md", "--content", cf, "--json")
	env1 := decodeEnvelope(t, p1)
	if _, _, code := exec(t, "page", "apply", "--root", dir, env1.Plan.Transaction, "--json"); code != int(contract.ExitSuccess) {
		t.Fatalf("seed apply exit = %d, want 0", code)
	}
	committed := snapshotAll(t, dir)
	originalHash := hashFile(t, live)

	// An edit that keeps the citations but adds a paragraph.
	edited := strings.Replace(authoringDraft,
		"Plants convert light energy into chemical energy stored in sugars.\n",
		"Plants convert light energy into chemical energy stored in sugars.\n\nThis occurs in the chloroplast.\n", 1)
	ef := writeContentFile(t, edited)

	p2, _, pcode := exec(t, "page", "plan", "--root", dir, "wiki/photosynthesis.md", "--content", ef, "--json")
	if pcode != int(contract.ExitSuccess) {
		t.Fatalf("edit plan exit = %d, want 0\n%s", pcode, p2)
	}
	env2 := decodeEnvelope(t, p2)
	if env2.Plan == nil || env2.Plan.NoOp {
		t.Fatalf("edit should be a real change, got %+v", env2.Plan)
	}
	if !strings.Contains(env2.Plan.Diff, "chloroplast") {
		t.Errorf("criterion 11: the preview diff must show the edit; got:\n%s", env2.Plan.Diff)
	}

	// Decline path: the preview precedes the write, so the live page is unchanged
	// and only the staging area grew.
	if got := hashFile(t, live); got != originalHash {
		t.Errorf("declined edit changed the live page: %q != %q", got, originalHash)
	}
	if changed := changedPaths(committed, snapshotAll(t, dir)); !allUnder(changed, ".llm-wiki/") {
		t.Errorf("declined edit touched non-staging paths: %v", changed)
	}

	// Apply path: the committed bytes equal exactly what was previewed.
	if _, _, code := exec(t, "page", "apply", "--root", dir, env2.Plan.Transaction, "--json"); code != int(contract.ExitSuccess) {
		t.Fatalf("edit apply exit = %d, want 0", code)
	}
	if got := hashFile(t, live); got != env2.Plan.StagedHash {
		t.Errorf("committed edit hash = %q, want previewed staged hash %q", got, env2.Plan.StagedHash)
	}
}

// TestAcceptanceCriterion12WritesOnlyThroughStagedPlanApply — criterion 12:
// managed writes reach the repository only through staged plan/apply. After
// draft/inspect/plan only the .llm-wiki/ staging area has changed and the target
// page is still absent; only apply materializes the page, and its sole live-tree
// change is exactly the target (staging is cleaned afterward) (ADR-005/006).
func TestAcceptanceCriterion12WritesOnlyThroughStagedPlanApply(t *testing.T) {
	t.Setenv("LLM_WIKI_EVIDENCE_SECTIONS", "Evidence")
	dir := initBundle(t)
	cf := writeContentFile(t, authoringDraft)
	base := snapshotAll(t, dir)

	// inspect-content stages nothing; plan stages only under .llm-wiki/.
	if _, _, code := exec(t, "page", "inspect", "--root", dir, "wiki/photosynthesis.md", "--content", cf, "--json"); code != int(contract.ExitSuccess) {
		t.Fatalf("inspect-content exit = %d, want 0", code)
	}
	pout, _, pcode := exec(t, "page", "plan", "--root", dir, "wiki/photosynthesis.md", "--content", cf, "--json")
	if pcode != int(contract.ExitSuccess) {
		t.Fatalf("plan exit = %d, want 0\n%s", pcode, pout)
	}
	penv := decodeEnvelope(t, pout)

	afterPlan := changedPaths(base, snapshotAll(t, dir))
	if !allUnder(afterPlan, ".llm-wiki/") {
		t.Errorf("criterion 12: draft/inspect/plan mutated non-staging paths: %v", afterPlan)
	}
	if _, ok := snapshotAll(t, dir)["wiki/photosynthesis.md"]; ok {
		t.Error("criterion 12: the page was written before apply")
	}

	// After apply the only live-tree change (outside staging) is the target page.
	preApply := snapshotAll(t, dir)
	if _, _, code := exec(t, "page", "apply", "--root", dir, penv.Plan.Transaction, "--json"); code != int(contract.ExitSuccess) {
		t.Fatalf("apply exit = %d, want 0", code)
	}
	liveChanges := []string{}
	for _, p := range changedPaths(preApply, snapshotAll(t, dir)) {
		if !strings.HasPrefix(p, ".llm-wiki/") {
			liveChanges = append(liveChanges, p)
		}
	}
	if len(liveChanges) != 1 || liveChanges[0] != "wiki/photosynthesis.md" {
		t.Errorf("criterion 12: apply changed live paths %v, want only [wiki/photosynthesis.md]", liveChanges)
	}
	// Staging is cleaned after a committed apply.
	if entries, _ := os.ReadDir(filepath.Join(dir, ".llm-wiki", "staging")); len(entries) != 0 {
		t.Errorf("criterion 12: staging not cleaned after apply: %v", entries)
	}
}

// staleBasePage is a plain (uncited) existing page for the stale-plan journey;
// citations are irrelevant to base-state binding, so it avoids the evidence
// machinery and isolates the stale-rejection behavior.
const staleBasePage = "---\ntitle: Stale\ntype: concept\n---\n\n# Stale\n\nBody.\n"

// TestAcceptanceCriterion13StalePlanRejectedZeroMutation — criterion 13: a plan
// whose target is changed out of band between plan and apply is rejected
// (exit 4, invalid-invocation, the standard six-field envelope, no apply payload)
// and mutates nothing. The proof is a whole-tree snapshot taken around the
// rejected apply — stronger than the plan-layer unit test's single-file check
// (ADR-006 base-state binding).
func TestAcceptanceCriterion13StalePlanRejectedZeroMutation(t *testing.T) {
	dir := initBundle(t)
	live := filepath.Join(dir, "wiki", "keep.md")
	if err := os.WriteFile(live, []byte(staleBasePage), 0o644); err != nil {
		t.Fatalf("seed live page: %v", err)
	}

	// Plan an edit against the live page.
	edited := strings.Replace(staleBasePage, "Body.\n", "Body.\n\nAn added paragraph.\n", 1)
	ef := writeContentFile(t, edited)
	pout, _, pcode := exec(t, "page", "plan", "--root", dir, "wiki/keep.md", "--content", ef, "--json")
	if pcode != int(contract.ExitSuccess) {
		t.Fatalf("plan exit = %d, want 0\n%s", pcode, pout)
	}
	penv := decodeEnvelope(t, pout)
	if penv.Plan == nil || penv.Plan.Transaction == "" {
		t.Fatalf("plan produced no transaction: %+v", penv.Plan)
	}

	// Change the live target out of band, invalidating the plan's captured base.
	stale := strings.Replace(staleBasePage, "Body.\n", "Body.\n\nWritten out of band.\n", 1)
	if err := os.WriteFile(live, []byte(stale), 0o644); err != nil {
		t.Fatalf("out-of-band write: %v", err)
	}

	// Whole-tree snapshot with the stale content on disk, before the rejected apply.
	before := snapshotAll(t, dir)

	aout, _, acode := exec(t, "page", "apply", "--root", dir, penv.Plan.Transaction, "--json")
	if acode != int(contract.ExitInvalidInvocation) {
		t.Fatalf("stale apply exit = %d, want 4\n%s", acode, aout)
	}
	aenv := decodeEnvelope(t, aout)
	if aenv.Status != contract.StatusInvalidInvocation {
		t.Errorf("status = %q, want invalid-invocation", aenv.Status)
	}
	if aenv.Apply != nil {
		t.Errorf("stale rejection must not carry an apply payload: %+v", aenv.Apply)
	}
	var generic map[string]json.RawMessage
	if err := json.Unmarshal([]byte(aout), &generic); err != nil {
		t.Fatalf("stdout not JSON: %v", err)
	}
	if len(generic) != 6 {
		t.Errorf("stale rejection envelope must be exactly six fields, got %d: %s", len(generic), aout)
	}

	// Whole-tree zero-mutation proof: the rejected apply changed nothing.
	if changed := changedPaths(before, snapshotAll(t, dir)); len(changed) != 0 {
		t.Errorf("criterion 13: rejected apply mutated the tree: %v", changed)
	}
}
