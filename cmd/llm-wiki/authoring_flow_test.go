package main

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/olivermorgan2/llm-wiki-kit/internal/contract"
)

// These tests execute the exact command sequence prescribed by the shipped
// wiki-authoring skill (skills/wiki-authoring/SKILL.md) against scratch bundles,
// mapped to issue #38's acceptance criteria 9–12. They leave the TestAcceptance*
// namespace to #39; the criterion-9 sourced-claim fixture below is the one #39
// builds its named corpus on, so it is kept aligned with the SKILL.md worked
// example byte for byte.
//
// authoringDraft is that worked example: a core-profile page whose every sourced
// claim carries a resolvable citation inside its `## Evidence` context — one
// in-bundle target and one https target. It must inspect and validate clean under
// LLM_WIKI_EVIDENCE_SECTIONS=Evidence.
const authoringDraft = "---\n" +
	"type: concept\n" +
	"title: Photosynthesis\n" +
	"description: How plants convert light into chemical energy.\n" +
	"timestamp: 2026-07-04\n" +
	"tags:\n" +
	"  - biology\n" +
	"aliases:\n" +
	"  - light-reactions\n" +
	"resource: https://example.com/photosynthesis\n" +
	"---\n\n" +
	"# Photosynthesis\n\n" +
	"Plants convert light energy into chemical energy stored in sugars.\n\n" +
	"## Evidence\n\n" +
	"An overview lives in the bundle index at [overview](wiki/index.md).\n" +
	"Further mechanism detail comes from an [external source](https://example.com/calvin-cycle).\n"

// hashFile returns the lowercase-hex SHA-256 of the file at path.
func hashFile(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

// snapshotAll walks root (including .llm-wiki) and returns a map of every
// bundle-relative slash path to its content hash — the basis for asserting which
// paths a step did or did not touch.
func snapshotAll(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		raw, rErr := os.ReadFile(p)
		if rErr != nil {
			return rErr
		}
		rel, rErr := filepath.Rel(root, p)
		if rErr != nil {
			return rErr
		}
		sum := sha256.Sum256(raw)
		out[filepath.ToSlash(rel)] = hex.EncodeToString(sum[:])
		return nil
	})
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	return out
}

// changedPaths returns the sorted set of paths that were added, removed, or whose
// hash changed between before and after.
func changedPaths(before, after map[string]string) []string {
	changed := map[string]bool{}
	for k, v := range after {
		if before[k] != v {
			changed[k] = true
		}
	}
	for k := range before {
		if _, ok := after[k]; !ok {
			changed[k] = true
		}
	}
	out := make([]string, 0, len(changed))
	for k := range changed {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// allUnder reports whether every path in paths is under the given slash prefix.
func allUnder(paths []string, prefix string) bool {
	for _, p := range paths {
		if !strings.HasPrefix(p, prefix) {
			return false
		}
	}
	return true
}

// TestAuthoringSkillFlowHappyPath walks the full skill sequence — inspect-content
// (validate), plan (diff preview), apply (commit), validate — and asserts it ends
// with a committed valid page whose bytes equal the previewed staged content.
// Criterion 9 (sourced claims carry resolvable citations) and criterion 11
// (preview precedes any write) are checked inline.
func TestAuthoringSkillFlowHappyPath(t *testing.T) {
	t.Setenv("LLM_WIKI_EVIDENCE_SECTIONS", "Evidence")
	dir := initBundle(t)
	cf := writeContentFile(t, authoringDraft)
	live := filepath.Join(dir, "wiki", "photosynthesis.md")

	// Step 3 — validate the draft: inspect-content is clean (criterion 9: no
	// core-citation-* findings for the sourced claims, no error findings).
	iout, _, icode := exec(t, "page", "inspect", "--root", dir, "wiki/photosynthesis.md", "--content", cf, "--json")
	if icode != int(contract.ExitSuccess) {
		t.Fatalf("inspect-content exit = %d, want 0\n%s", icode, iout)
	}
	ienv := decodeEnvelope(t, iout)
	for _, f := range ienv.Findings {
		if strings.HasPrefix(f.Code, "core-citation-") {
			t.Errorf("criterion 9: sourced-claim draft raised a citation finding: %+v", f)
		}
		if f.Severity == contract.SeverityError {
			t.Errorf("draft has an error-severity finding: %+v", f)
		}
	}

	// Step 4 — stage: plan previews a non-empty diff against an absent base and
	// touches nothing but the staging area (criterion 11: preview before write).
	beforePlan := snapshotAll(t, dir)
	pout, _, pcode := exec(t, "page", "plan", "--root", dir, "wiki/photosynthesis.md", "--content", cf, "--json")
	if pcode != int(contract.ExitSuccess) {
		t.Fatalf("plan exit = %d, want 0\n%s", pcode, pout)
	}
	penv := decodeEnvelope(t, pout)
	if penv.Plan == nil || penv.Plan.NoOp {
		t.Fatalf("plan should be a real staged change, got %+v", penv.Plan)
	}
	if !penv.Plan.BaseAbsent {
		t.Errorf("new page plan should report an absent base: %+v", penv.Plan)
	}
	if strings.TrimSpace(penv.Plan.Diff) == "" {
		t.Error("plan diff is empty; a preview must be shown before write")
	}
	if _, err := os.Stat(live); !os.IsNotExist(err) {
		t.Errorf("plan created the live page before apply: %v", err)
	}
	afterPlan := snapshotAll(t, dir)
	if changed := changedPaths(beforePlan, afterPlan); !allUnder(changed, ".llm-wiki/") {
		t.Errorf("plan touched paths outside the staging area: %v", changed)
	}

	// Step 5 — apply: commit the staged transaction.
	aout, _, acode := exec(t, "page", "apply", "--root", dir, penv.Plan.Transaction, "--json")
	if acode != int(contract.ExitSuccess) {
		t.Fatalf("apply exit = %d, want 0\n%s", acode, aout)
	}
	aenv := decodeEnvelope(t, aout)
	if aenv.Apply == nil || len(aenv.Apply.Committed) != 1 || aenv.Apply.Committed[0] != "wiki/photosynthesis.md" {
		t.Fatalf("apply payload = %+v, want committed [wiki/photosynthesis.md]", aenv.Apply)
	}
	// The committed page is byte-equal to the previewed staged content.
	if got := hashFile(t, live); got != penv.Plan.StagedHash {
		t.Errorf("committed page hash = %q, want staged hash %q", got, penv.Plan.StagedHash)
	}

	// Step 6 — confirm: validate is clean over the committed bundle.
	vout, _, vcode := exec(t, "validate", dir, "--json")
	if vcode != int(contract.ExitSuccess) {
		t.Fatalf("validate after apply exit = %d, want 0\n%s", vcode, vout)
	}
}

// TestAuthoringSkillFlowIdempotent covers criterion 10: re-running the identical
// draft after it is committed is a no-op — no duplicate page, nothing to apply,
// and the tree is byte-identical.
func TestAuthoringSkillFlowIdempotent(t *testing.T) {
	t.Setenv("LLM_WIKI_EVIDENCE_SECTIONS", "Evidence")
	dir := initBundle(t)
	cf := writeContentFile(t, authoringDraft)

	// First pass: plan + apply commits the page.
	p1, _, _ := exec(t, "page", "plan", "--root", dir, "wiki/photosynthesis.md", "--content", cf, "--json")
	env1 := decodeEnvelope(t, p1)
	if env1.Plan == nil || env1.Plan.Transaction == "" {
		t.Fatalf("first plan produced no transaction: %+v", env1.Plan)
	}
	if _, _, code := exec(t, "page", "apply", "--root", dir, env1.Plan.Transaction, "--json"); code != int(contract.ExitSuccess) {
		t.Fatalf("first apply exit = %d, want 0", code)
	}

	afterCommit := snapshotAll(t, dir)

	// Second pass with the identical draft: plan is a no-op with an empty txn.
	p2, _, pcode := exec(t, "page", "plan", "--root", dir, "wiki/photosynthesis.md", "--content", cf, "--json")
	if pcode != int(contract.ExitSuccess) {
		t.Fatalf("second plan exit = %d, want 0\n%s", pcode, p2)
	}
	env2 := decodeEnvelope(t, p2)
	if env2.Plan == nil || !env2.Plan.NoOp {
		t.Errorf("criterion 10: identical re-draft should be a no-op, got %+v", env2.Plan)
	}
	if env2.Plan != nil && env2.Plan.Transaction != "" {
		t.Errorf("no-op plan should stage no transaction, got %q", env2.Plan.Transaction)
	}
	if len(env2.AffectedPaths) != 0 {
		t.Errorf("no-op plan should list no affected paths, got %v", env2.AffectedPaths)
	}

	// Exactly one page file exists, and the tree is unchanged by the no-op plan.
	matches, _ := filepath.Glob(filepath.Join(dir, "wiki", "photosynthesis*.md"))
	if len(matches) != 1 {
		t.Errorf("expected exactly one photosynthesis page, found %v", matches)
	}
	if changed := changedPaths(afterCommit, snapshotAll(t, dir)); len(changed) != 0 {
		t.Errorf("no-op plan changed the tree: %v", changed)
	}
}

// TestAuthoringSkillFlowEditPreview covers criterion 11 for an existing page: the
// plan previews the edit; the decline path (no apply) leaves the tree unchanged;
// the apply path commits exactly the previewed bytes.
func TestAuthoringSkillFlowEditPreview(t *testing.T) {
	t.Setenv("LLM_WIKI_EVIDENCE_SECTIONS", "Evidence")
	dir := initBundle(t)
	live := filepath.Join(dir, "wiki", "photosynthesis.md")

	// Commit the initial page.
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

	// Plan previews the change.
	p2, _, pcode := exec(t, "page", "plan", "--root", dir, "wiki/photosynthesis.md", "--content", ef, "--json")
	if pcode != int(contract.ExitSuccess) {
		t.Fatalf("edit plan exit = %d, want 0\n%s", pcode, p2)
	}
	env2 := decodeEnvelope(t, p2)
	if env2.Plan == nil || env2.Plan.NoOp {
		t.Fatalf("edit should be a real change, got %+v", env2.Plan)
	}
	if !strings.Contains(env2.Plan.Diff, "chloroplast") {
		t.Errorf("diff should show the edit; got:\n%s", env2.Plan.Diff)
	}

	// Decline path: do not apply. The live page is unchanged; only staging grew.
	if got := hashFile(t, live); got != originalHash {
		t.Errorf("declined edit changed the live page: %q != %q", got, originalHash)
	}
	if changed := changedPaths(committed, snapshotAll(t, dir)); !allUnder(changed, ".llm-wiki/") {
		t.Errorf("declined edit touched non-staging paths: %v", changed)
	}

	// Apply path: commit exactly the previewed bytes.
	if _, _, code := exec(t, "page", "apply", "--root", dir, env2.Plan.Transaction, "--json"); code != int(contract.ExitSuccess) {
		t.Fatalf("edit apply exit = %d, want 0", code)
	}
	if got := hashFile(t, live); got != env2.Plan.StagedHash {
		t.Errorf("committed edit hash = %q, want previewed staged hash %q", got, env2.Plan.StagedHash)
	}
}

// TestAuthoringSkillFlowEngineMediatedMutation covers criterion 12: the flow
// itself performs no direct repository write. After draft+inspect+plan only the
// staging area has changed; after apply only the target page has changed.
func TestAuthoringSkillFlowEngineMediatedMutation(t *testing.T) {
	t.Setenv("LLM_WIKI_EVIDENCE_SECTIONS", "Evidence")
	dir := initBundle(t)
	cf := writeContentFile(t, authoringDraft)

	base := snapshotAll(t, dir)

	// Draft is a scratch file outside the bundle; inspect-content stages nothing.
	if _, _, code := exec(t, "page", "inspect", "--root", dir, "wiki/photosynthesis.md", "--content", cf, "--json"); code != int(contract.ExitSuccess) {
		t.Fatalf("inspect-content exit = %d, want 0", code)
	}
	pout, _, _ := exec(t, "page", "plan", "--root", dir, "wiki/photosynthesis.md", "--content", cf, "--json")
	penv := decodeEnvelope(t, pout)

	// After inspect+plan: only .llm-wiki/staging changed; no page was written.
	afterPlan := changedPaths(base, snapshotAll(t, dir))
	if !allUnder(afterPlan, ".llm-wiki/") {
		t.Errorf("criterion 12: draft/inspect/plan mutated non-staging paths: %v", afterPlan)
	}
	if _, ok := snapshotAll(t, dir)["wiki/photosynthesis.md"]; ok {
		t.Error("criterion 12: the page was written before apply")
	}

	// After apply: the only live-tree change (outside staging) is the target page.
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
}

// TestAuthoringSkillFlowCitationLossGuard is a regression over the citation-loss
// gate the skill relies on: an edit dropping a citation stages a plan whose
// apply refuses without --approve (exit 3, zero mutation) and commits with it.
func TestAuthoringSkillFlowCitationLossGuard(t *testing.T) {
	t.Setenv("LLM_WIKI_EVIDENCE_SECTIONS", "Evidence")
	dir := initBundle(t)
	live := filepath.Join(dir, "wiki", "photosynthesis.md")

	// Commit the cited page.
	cf := writeContentFile(t, authoringDraft)
	p1, _, _ := exec(t, "page", "plan", "--root", dir, "wiki/photosynthesis.md", "--content", cf, "--json")
	env1 := decodeEnvelope(t, p1)
	if _, _, code := exec(t, "page", "apply", "--root", dir, env1.Plan.Transaction, "--json"); code != int(contract.ExitSuccess) {
		t.Fatalf("seed apply exit = %d, want 0", code)
	}
	committedHash := hashFile(t, live)

	// Edit that removes the https citation from the evidence context.
	dropped := strings.Replace(authoringDraft,
		"Further mechanism detail comes from an [external source](https://example.com/calvin-cycle).\n",
		"Further mechanism detail is inferred.\n", 1)
	df := writeContentFile(t, dropped)

	p2, _, pcode := exec(t, "page", "plan", "--root", dir, "wiki/photosynthesis.md", "--content", df, "--json")
	if pcode != int(contract.ExitSuccess) {
		t.Fatalf("loss plan exit = %d, want 0 (plan itself succeeds)\n%s", pcode, p2)
	}
	env2 := decodeEnvelope(t, p2)
	if env2.Approval == nil || !env2.Approval.Required {
		t.Fatalf("dropping a citation must raise the approval gate: %+v", env2.Approval)
	}

	// Unapproved apply refuses: exit 3, zero mutation.
	uout, _, ucode := exec(t, "page", "apply", "--root", dir, env2.Plan.Transaction, "--json")
	if ucode != int(contract.ExitApprovalRequired) {
		t.Fatalf("unapproved apply exit = %d, want 3\n%s", ucode, uout)
	}
	if got := hashFile(t, live); got != committedHash {
		t.Errorf("refused apply mutated the live page: %q != %q", got, committedHash)
	}

	// Approved apply commits.
	if _, _, code := exec(t, "page", "apply", "--root", dir, env2.Plan.Transaction, "--approve", "--json"); code != int(contract.ExitSuccess) {
		t.Fatalf("approved apply exit = %d, want 0", code)
	}
	if got := hashFile(t, live); got == committedHash {
		t.Error("approved apply did not change the page")
	}
}
