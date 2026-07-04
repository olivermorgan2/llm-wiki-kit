package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/olivermorgan2/llm-wiki-kit/internal/contract"
)

// canonicalScalarPage is a page whose frontmatter is already in the marshaler's
// canonical (fixed-point) form — simple scalars, one per line — so feeding its
// exact bytes back to page plan is a no-op.
const canonicalScalarPage = "---\ntitle: Same\ntype: concept\ncustom_field: keep\n---\n\n# Same\n\nBody.\n"

// writeContentFile writes proposed page content to a temp file and returns its
// path, for feeding to `page plan --content`.
func writeContentFile(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "proposed.md")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write content file: %v", err)
	}
	return p
}

// Happy path — new page: exit 0, operation "page plan", a plan payload with an
// absent base and a full-file diff, the target listed in affectedPaths, and no
// live file created.
func TestPagePlanNewPage(t *testing.T) {
	dir := initBundle(t)
	content := writeContentFile(t, "---\ntitle: Fresh\ntype: concept\ncustom_field: keep\n---\n\n# Fresh\n\nBody.\n")

	stdout, _, code := exec(t, "page", "plan", "--root", dir, "wiki/fresh.md", "--content", content, "--json")
	if code != int(contract.ExitSuccess) {
		t.Fatalf("page plan exit = %d, want 0\n%s", code, stdout)
	}
	env := decodeEnvelope(t, stdout)
	if env.Operation != "page plan" {
		t.Errorf("operation = %q, want \"page plan\"", env.Operation)
	}
	if env.Plan == nil {
		t.Fatal("plan payload missing")
	}
	if env.Plan.Path != "wiki/fresh.md" {
		t.Errorf("plan.path = %q, want wiki/fresh.md", env.Plan.Path)
	}
	if !env.Plan.BaseAbsent {
		t.Error("plan.baseAbsent = false, want true for a new page")
	}
	if env.Plan.NoOp {
		t.Error("plan.noOp = true, want false")
	}
	if env.Plan.Transaction == "" {
		t.Error("plan.transaction empty, want a staging txn id")
	}
	if !strings.Contains(env.Plan.Diff, "@@ -0,0 ") || !strings.Contains(env.Plan.Diff, "+# Fresh") {
		t.Errorf("plan.diff should be a full-file addition:\n%s", env.Plan.Diff)
	}
	if len(env.AffectedPaths) != 1 || env.AffectedPaths[0] != "wiki/fresh.md" {
		t.Errorf("affectedPaths = %v, want [wiki/fresh.md]", env.AffectedPaths)
	}

	// Planning never mutates live files: the target must not exist.
	if _, err := os.Stat(filepath.Join(dir, "wiki", "fresh.md")); !os.IsNotExist(err) {
		t.Errorf("live target should not exist after plan; err = %v", err)
	}
	// The staged change is on disk under the reported transaction.
	manifest := filepath.Join(dir, ".llm-wiki", "staging", env.Plan.Transaction, "manifest.json")
	if _, err := os.Stat(manifest); err != nil {
		t.Errorf("staging manifest missing: %v", err)
	}
}

// A whole-page edit over an existing page preserves unknown frontmatter fields
// in the staged postimage and never touches the live file.
func TestPagePlanExistingPreservesUnknownFields(t *testing.T) {
	dir := initBundle(t)
	live := filepath.Join(dir, "wiki", "keep.md")
	existing := "---\ntitle: Old\ntype: concept\ncustom_field: keep-me\nx-tool-meta: 7\n---\n\n# Old\n\nBody.\n"
	if err := os.WriteFile(live, []byte(existing), 0o644); err != nil {
		t.Fatalf("write existing: %v", err)
	}
	content := writeContentFile(t, "---\ntitle: New\ntype: concept\ncustom_field: keep-me\nx-tool-meta: 7\n---\n\n# Old\n\nBody.\n")

	stdout, _, code := exec(t, "page", "plan", "--root", dir, "wiki/keep.md", "--content", content, "--json")
	if code != int(contract.ExitSuccess) {
		t.Fatalf("page plan exit = %d, want 0\n%s", code, stdout)
	}
	env := decodeEnvelope(t, stdout)
	if env.Plan == nil || env.Plan.NoOp {
		t.Fatalf("expected a non-no-op plan payload: %+v", env.Plan)
	}

	// The staged postimage preserves the unknown fields.
	blob, err := os.ReadFile(filepath.Join(dir, ".llm-wiki", "staging", env.Plan.Transaction, "files", "0000"))
	if err != nil {
		t.Fatalf("read staged postimage: %v", err)
	}
	for _, field := range []string{"custom_field: keep-me", "x-tool-meta: 7", "title: New"} {
		if !strings.Contains(string(blob), field) {
			t.Errorf("staged postimage dropped %q:\n%s", field, blob)
		}
	}

	// Live file is unchanged.
	got, err := os.ReadFile(live)
	if err != nil {
		t.Fatalf("read live: %v", err)
	}
	if string(got) != existing {
		t.Errorf("live page mutated by plan:\n%s", got)
	}
}

// Repeated identical input is a no-op: exit 0, plan.noOp true, no transaction,
// and no staging dir is created.
func TestPagePlanNoOp(t *testing.T) {
	dir := initBundle(t)
	live := filepath.Join(dir, "wiki", "same.md")
	if err := os.WriteFile(live, []byte(canonicalScalarPage), 0o644); err != nil {
		t.Fatalf("write page: %v", err)
	}
	content := writeContentFile(t, canonicalScalarPage)

	stdout, _, code := exec(t, "page", "plan", "--root", dir, "wiki/same.md", "--content", content, "--json")
	if code != int(contract.ExitSuccess) {
		t.Fatalf("page plan exit = %d, want 0\n%s", code, stdout)
	}
	env := decodeEnvelope(t, stdout)
	if env.Plan == nil || !env.Plan.NoOp {
		t.Fatalf("expected a no-op plan payload: %+v", env.Plan)
	}
	if env.Plan.Transaction != "" {
		t.Errorf("plan.transaction = %q, want empty for a no-op", env.Plan.Transaction)
	}
	if len(env.AffectedPaths) != 0 {
		t.Errorf("affectedPaths = %v, want empty for a no-op", env.AffectedPaths)
	}
	// No staging dir was created.
	if entries, err := os.ReadDir(filepath.Join(dir, ".llm-wiki", "staging")); err == nil && len(entries) != 0 {
		t.Errorf("no-op created staging dirs: %v", entries)
	}
}

// Boundary refusal: a target outside the root is exit 5, system failure, the
// standard six-field envelope, and leaks no content.
func TestPagePlanBoundaryRefusal(t *testing.T) {
	dir := initBundle(t)
	content := writeContentFile(t, "---\ntitle: X\n---\n\n# X\n")

	stdout, stderr, code := exec(t, "page", "plan", "--root", dir, "../evil.md", "--content", content, "--json")
	if code != int(contract.ExitSystemFailure) {
		t.Fatalf("exit = %d, want 5\n%s", code, stdout)
	}
	env := decodeEnvelope(t, stdout)
	if env.Status != contract.StatusSystemFailure {
		t.Errorf("status = %q, want system-or-filesystem-failure", env.Status)
	}
	if env.Plan != nil {
		t.Errorf("refusal must not carry a plan payload: %+v", env.Plan)
	}
	var generic map[string]json.RawMessage
	if err := json.Unmarshal([]byte(stdout), &generic); err != nil {
		t.Fatalf("stdout not JSON: %v", err)
	}
	if len(generic) != 6 {
		t.Errorf("refusal envelope must be exactly six fields, got %d: %s", len(generic), stdout)
	}
	_ = stderr
}

// Missing <path> is invalid-invocation (exit 4).
func TestPagePlanMissingPath(t *testing.T) {
	dir := initBundle(t)
	_, _, code := exec(t, "page", "plan", "--root", dir)
	if code != int(contract.ExitInvalidInvocation) {
		t.Errorf("exit = %d, want 4", code)
	}
}

// Human output shows the diff and the staged transaction for a real change.
func TestPagePlanHumanOutput(t *testing.T) {
	dir := initBundle(t)
	content := writeContentFile(t, "---\ntitle: Human\ntype: concept\n---\n\n# Human\n")

	stdout, _, code := exec(t, "page", "plan", "--root", dir, "wiki/human.md", "--content", content)
	if code != int(contract.ExitSuccess) {
		t.Fatalf("exit = %d, want 0\n%s", code, stdout)
	}
	if !strings.Contains(stdout, "staged change") || !strings.Contains(stdout, "+# Human") {
		t.Errorf("human output missing plan summary/diff:\n%s", stdout)
	}
	if !strings.Contains(stdout, "base:   absent (new page)") {
		t.Errorf("human output should show an absent base:\n%s", stdout)
	}
}
