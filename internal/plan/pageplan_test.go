package plan

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/olivermorgan2/llm-wiki-kit/internal/fsafe"
	"github.com/olivermorgan2/llm-wiki-kit/internal/validate"
	"github.com/olivermorgan2/llm-wiki-kit/internal/yamladapter"
)

// canonicalPage returns raw with its frontmatter normalized to the fixed-point
// form Plan stages, so a page written with it and then fed back to Plan is a
// no-op. It exercises the same normalization the plan cycle uses.
func canonicalPage(raw string) string {
	return string(normalizePage([]byte(raw), yamladapter.New()))
}

// stagedManifest is the subset of the ADR-006 staging manifest the tests assert
// against: the base record (absent sentinel vs hashed) bound to each target.
type stagedManifest struct {
	Txn     string `json:"txn"`
	Entries []struct {
		Path   string `json:"path"`
		Staged string `json:"staged"`
		Base   struct {
			Absent bool   `json:"absent"`
			SHA256 string `json:"sha256"`
		} `json:"base"`
	} `json:"entries"`
}

// readManifest reads and decodes the staging manifest for txn id under root.
func readManifest(t *testing.T, root, id string) stagedManifest {
	t.Helper()
	p := filepath.Join(root, fsafe.StagingDir, "staging", id, "manifest.json")
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read manifest %s: %v", p, err)
	}
	var m stagedManifest
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	return m
}

// stagingDirCount returns the number of transaction dirs under .llm-wiki/staging
// (0 when the staging area does not exist).
func stagingDirCount(t *testing.T, root string) int {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(root, fsafe.StagingDir, "staging"))
	if errors.Is(err, os.ErrNotExist) {
		return 0
	}
	if err != nil {
		t.Fatalf("read staging dir: %v", err)
	}
	return len(entries)
}

// AC: a plan for a new page records an absent base state and shows a full-file
// diff, staging the change without creating the live file.
func TestPlanNewPageAbsentBaseFullFileDiff(t *testing.T) {
	root := t.TempDir()
	proposed := "---\ntitle: New Page\ntype: concept\ncustom_field: keep\n---\n\n# New Page\n\nBody line.\n"

	res, err := Plan(root, "wiki/new.md", []byte(proposed), yamladapter.New(), validate.Options{})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if !res.BaseAbsent {
		t.Error("BaseAbsent = false, want true for a new page")
	}
	if res.BaseHash != "" {
		t.Errorf("BaseHash = %q, want empty for an absent base", res.BaseHash)
	}
	if res.NoOp {
		t.Error("NoOp = true, want false for a new page")
	}
	if res.TxnID == "" {
		t.Error("TxnID empty, want a staged transaction id")
	}

	// Full-file diff: an absent old side (@@ -0,0 …) and every content line added.
	if !strings.Contains(res.Diff, "--- "+devNull) {
		t.Errorf("diff should mark the old side /dev/null:\n%s", res.Diff)
	}
	if !strings.Contains(res.Diff, "@@ -0,0 ") {
		t.Errorf("diff header should show an empty old side (-0,0):\n%s", res.Diff)
	}
	for _, line := range strings.Split(strings.TrimRight(res.Diff, "\n"), "\n") {
		if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "@@") {
			continue
		}
		if !strings.HasPrefix(line, "+") {
			t.Errorf("new-page diff line is not an addition: %q\n%s", line, res.Diff)
		}
	}
	if !strings.Contains(res.Diff, "+# New Page") {
		t.Errorf("diff should add the page body:\n%s", res.Diff)
	}

	// Planning never mutates live files: the target must not exist.
	if _, err := os.Stat(filepath.Join(root, "wiki", "new.md")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("live target should not exist after plan; stat err = %v", err)
	}

	// The staging manifest records the absent base sentinel for the target.
	m := readManifest(t, root, res.TxnID)
	if len(m.Entries) != 1 {
		t.Fatalf("manifest entries = %d, want 1", len(m.Entries))
	}
	e := m.Entries[0]
	if e.Path != "wiki/new.md" {
		t.Errorf("manifest path = %q, want wiki/new.md", e.Path)
	}
	if !e.Base.Absent {
		t.Error("manifest base.absent = false, want true (absent-target sentinel)")
	}
	if e.Staged != res.StagedHash {
		t.Errorf("manifest staged hash %q != result staged hash %q", e.Staged, res.StagedHash)
	}
}

// AC: a plan over an existing page preserves unknown frontmatter fields in the
// staged result and leaves the live file untouched.
func TestPlanExistingPagePreservesUnknownFields(t *testing.T) {
	root := t.TempDir()
	existing := canonicalPage("---\ntitle: Old\ntype: concept\ncustom_field: keep-me\nx-tool-meta: 42\n---\n\n# Old\n\nBody.\n")
	writePage(t, root, "wiki/p.md", existing)

	// A whole-page edit that changes the title but re-includes the unknown fields.
	proposed := "---\ntitle: New\ntype: concept\ncustom_field: keep-me\nx-tool-meta: 42\n---\n\n# Old\n\nBody.\n"
	res, err := Plan(root, "wiki/p.md", []byte(proposed), yamladapter.New(), validate.Options{})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if res.NoOp {
		t.Fatal("NoOp = true, want false for a title change")
	}
	if res.BaseAbsent {
		t.Error("BaseAbsent = true, want false over an existing page")
	}
	if res.BaseHash != hashBytes([]byte(existing)) {
		t.Errorf("BaseHash = %q, want hash of the live page", res.BaseHash)
	}

	staged := string(res.Staged)
	for _, field := range []string{"custom_field: keep-me", "x-tool-meta: 42", "title: New"} {
		if !strings.Contains(staged, field) {
			t.Errorf("staged result dropped %q:\n%s", field, staged)
		}
	}

	// The diff reflects the title change.
	if !strings.Contains(res.Diff, "-title: Old") || !strings.Contains(res.Diff, "+title: New") {
		t.Errorf("diff should show the title change:\n%s", res.Diff)
	}

	// The staged postimage on disk equals the reported staged bytes.
	blob, err := os.ReadFile(filepath.Join(root, fsafe.StagingDir, "staging", res.TxnID, "files", "0000"))
	if err != nil {
		t.Fatalf("read staged postimage: %v", err)
	}
	if string(blob) != staged {
		t.Errorf("staged postimage on disk differs from result.Staged")
	}

	// Planning never mutates live files: the page is byte-identical to before.
	live, err := os.ReadFile(filepath.Join(root, "wiki", "p.md"))
	if err != nil {
		t.Fatalf("read live page: %v", err)
	}
	if string(live) != existing {
		t.Errorf("live page was mutated by plan:\n got: %q\nwant: %q", live, existing)
	}
}

// AC: repeated identical input returns a no-op and creates no staging.
func TestPlanIdenticalInputIsNoOp(t *testing.T) {
	root := t.TempDir()
	existing := canonicalPage("---\ntitle: Same\ntype: concept\ncustom_field: keep\n---\n\n# Same\n\nBody.\n")
	writePage(t, root, "wiki/same.md", existing)

	if n := stagingDirCount(t, root); n != 0 {
		t.Fatalf("staging dir count = %d before plan, want 0", n)
	}

	res, err := Plan(root, "wiki/same.md", []byte(existing), yamladapter.New(), validate.Options{})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if !res.NoOp {
		t.Errorf("NoOp = false, want true for identical input")
	}
	if res.TxnID != "" {
		t.Errorf("TxnID = %q, want empty for a no-op", res.TxnID)
	}
	if res.Diff != "" {
		t.Errorf("Diff = %q, want empty for a no-op", res.Diff)
	}
	if n := stagingDirCount(t, root); n != 0 {
		t.Errorf("staging dir count = %d after no-op, want 0 (no duplicate staging)", n)
	}
}

// A boundary-escaping target is refused with the fsafe sentinel, unwrapped.
func TestPlanBoundaryEscape(t *testing.T) {
	root := t.TempDir()
	_, err := Plan(root, filepath.Join("..", "evil.md"), []byte("x"), yamladapter.New(), validate.Options{})
	if !errors.Is(err, fsafe.ErrOutsideBoundary) {
		t.Errorf("err = %v, want fsafe.ErrOutsideBoundary", err)
	}
}

// A non-markdown target is rejected before any filesystem work.
func TestPlanNotMarkdown(t *testing.T) {
	root := t.TempDir()
	_, err := Plan(root, "notes.txt", []byte("x"), yamladapter.New(), validate.Options{})
	if !errors.Is(err, ErrNotMarkdown) {
		t.Errorf("err = %v, want ErrNotMarkdown", err)
	}
}

// An existing non-regular target (a directory) is rejected, not planned.
func TestPlanNonRegularTarget(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "wiki", "dir.md"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	_, err := Plan(root, "wiki/dir.md", []byte("x"), yamladapter.New(), validate.Options{})
	if !errors.Is(err, ErrTargetNotRegular) {
		t.Errorf("err = %v, want ErrTargetNotRegular", err)
	}
}

// Content with unparseable frontmatter is staged verbatim (normalization is a
// canonicalization, not a validation gate) and still stages as a change.
func TestPlanVerbatimWhenFrontmatterUnparseable(t *testing.T) {
	root := t.TempDir()
	proposed := "---\ntitle: {broken\n---\n\n# Body\n"
	res, err := Plan(root, "wiki/bad.md", []byte(proposed), yamladapter.New(), validate.Options{})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if string(res.Staged) != proposed {
		t.Errorf("unparseable frontmatter should stage verbatim:\n got: %q\nwant: %q", res.Staged, proposed)
	}
}
