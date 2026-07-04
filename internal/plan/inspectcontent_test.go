package plan

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/olivermorgan2/llm-wiki-kit/internal/contract"
	"github.com/olivermorgan2/llm-wiki-kit/internal/fsafe"
	"github.com/olivermorgan2/llm-wiki-kit/internal/validate"
	"github.com/olivermorgan2/llm-wiki-kit/internal/yamladapter"
)

// citedDraft has a valid frontmatter and a sourced claim citing an https target
// inside its `## Evidence` context — the shape the authoring skill produces.
const citedDraft = "---\ntype: concept\ntitle: Cited Draft\ndescription: A draft with a sourced claim.\n" +
	"timestamp: 2026-07-04\ntags: []\naliases: []\nresource: https://example.com\n---\n\n" +
	"# Cited Draft\n\n## Evidence\n\nA sourced claim [src](https://example.com/p).\n"

// malformedCitationDraft cites a malformed target inside its evidence context so
// the resolver fires a core-citation-* finding.
const malformedCitationDraft = "---\ntype: concept\ntitle: Bad Cite\ndescription: A draft with a malformed citation.\n" +
	"timestamp: 2026-07-04\ntags: []\naliases: []\nresource: https://example.com\n---\n\n" +
	"# Bad Cite\n\n## Evidence\n\nSee [x](mailto:nobody@example.com).\n"

// snapshotTree walks root and returns a map of bundle-relative slash paths to
// content hashes, for asserting a read-only operation left the tree untouched.
func snapshotTree(t *testing.T, root string) map[string]string {
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

// (a) A new-page draft (no live target on disk) is inspected in full bundle
// context: the report reflects the proposed bytes, the citation resolver sees
// its evidence context, and the content hash is the SHA-256 of the draft as read.
func TestInspectContentNewPageDraft(t *testing.T) {
	root := t.TempDir()
	// Seed a sibling so the bundle is not empty; the draft page itself is absent.
	writePage(t, root, "wiki/index.md", validPage)

	rep, err := InspectContent(root, "wiki/new.md", []byte(malformedCitationDraft),
		yamladapter.New(), []string{"Evidence"}, nil)
	if err != nil {
		t.Fatalf("InspectContent: %v", err)
	}
	if rep.Path != "wiki/new.md" {
		t.Errorf("Path = %q, want wiki/new.md", rep.Path)
	}
	if !rep.Parsed {
		t.Error("Parsed = false, want true for well-formed draft frontmatter")
	}
	// The citation resolver, wired with the Evidence section, must see the draft's
	// malformed citation target.
	hasCitation := false
	for _, f := range rep.Findings {
		if strings.HasPrefix(f.Code, "core-citation-") {
			hasCitation = true
		}
	}
	if !hasCitation {
		t.Errorf("expected a core-citation-* finding over the draft evidence context; got %+v", rep.Findings)
	}
	// ContentHash is the hash of the proposed bytes exactly as read.
	sum := sha256.Sum256([]byte(malformedCitationDraft))
	if want := hex.EncodeToString(sum[:]); rep.ContentHash != want {
		t.Errorf("ContentHash = %q, want %q (hash of the proposed bytes)", rep.ContentHash, want)
	}
	// No citation findings for the clean cited draft.
	clean, err := InspectContent(root, "wiki/new.md", []byte(citedDraft),
		yamladapter.New(), []string{"Evidence"}, nil)
	if err != nil {
		t.Fatalf("InspectContent(clean): %v", err)
	}
	for _, f := range clean.Findings {
		if strings.HasPrefix(f.Code, "core-citation-") {
			t.Errorf("clean cited draft raised a citation finding: %+v", f)
		}
	}
	if clean.Status != contract.StatusSuccess {
		t.Errorf("clean draft Status = %q, want success (%+v)", clean.Status, clean.Findings)
	}
}

// (b) A draft over an existing page replaces the live bytes in the run: a draft
// that introduces a broken link produces the finding even though the live page
// has none, proving the engine ran over the proposed content.
func TestInspectContentReplacesExistingPage(t *testing.T) {
	root := t.TempDir()
	writePage(t, root, "wiki/page.md", validPage) // clean live page, no broken link

	draft := "---\ntype: concept\ntitle: Edited\ndescription: Now links to a ghost.\n" +
		"timestamp: 2026-07-04\ntags: []\naliases: []\nresource: https://x\n---\n\n" +
		"See [ghost](wiki/ghost.md).\n"

	rep, err := InspectContent(root, "wiki/page.md", []byte(draft), yamladapter.New(), nil, nil)
	if err != nil {
		t.Fatalf("InspectContent: %v", err)
	}
	found := false
	for _, f := range rep.Findings {
		if f.Code == "core-broken-link" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a core-broken-link finding from the draft; got %+v", rep.Findings)
	}
	// The live page on disk is untouched and still has no broken link.
	live, err := Inspect(root, "wiki/page.md", yamladapter.New(), nil, nil)
	if err != nil {
		t.Fatalf("Inspect(live): %v", err)
	}
	for _, f := range live.Findings {
		if f.Code == "core-broken-link" {
			t.Errorf("live page unexpectedly has a broken link: %+v", f)
		}
	}
}

// (c) A draft in a not-yet-existing subdirectory is inspected in full bundle
// context (overlay directory synthesis): a link to an existing bundle page
// resolves (no broken-link), proving both the synthesized draft is walked and
// the real bundle files are visible.
func TestInspectContentDraftInNewSubdir(t *testing.T) {
	root := t.TempDir()
	writePage(t, root, "wiki/index.md", validPage)

	draft := "---\ntype: concept\ntitle: Nested\ndescription: In a new subdir.\n" +
		"timestamp: 2026-07-04\ntags: []\naliases: []\nresource: https://x\n---\n\n" +
		"Up to [index](wiki/index.md), down to [ghost](wiki/sub/ghost.md).\n"

	rep, err := InspectContent(root, "wiki/sub/deep/new.md", []byte(draft), yamladapter.New(), nil, nil)
	if err != nil {
		t.Fatalf("InspectContent: %v", err)
	}
	if rep.Path != "wiki/sub/deep/new.md" {
		t.Errorf("Path = %q, want wiki/sub/deep/new.md", rep.Path)
	}
	brokenTargets := []string{}
	for _, f := range rep.Findings {
		if f.Code == "core-broken-link" {
			brokenTargets = append(brokenTargets, f.Message)
		}
	}
	// The existing target must resolve; only the ghost is broken.
	if len(brokenTargets) != 1 {
		t.Errorf("want exactly one broken link (the ghost), got %d: %v\nfindings: %+v",
			len(brokenTargets), brokenTargets, rep.Findings)
	}
	// The draft file must NOT have been written to disk.
	if _, err := os.Stat(filepath.Join(root, "wiki", "sub", "deep", "new.md")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("draft in new subdir must not touch disk; stat err = %v", err)
	}
}

// (d) Content-inspect is read-only: the live tree is byte-identical before and
// after, and the draft target is never created.
func TestInspectContentLeavesTreeUntouched(t *testing.T) {
	root := t.TempDir()
	writePage(t, root, "wiki/index.md", validPage)

	before := snapshotTree(t, root)
	if _, err := InspectContent(root, "wiki/new.md", []byte(citedDraft),
		yamladapter.New(), []string{"Evidence"}, nil); err != nil {
		t.Fatalf("InspectContent: %v", err)
	}
	after := snapshotTree(t, root)

	if len(before) != len(after) {
		t.Fatalf("tree file count changed: %d → %d", len(before), len(after))
	}
	beforeKeys := make([]string, 0, len(before))
	for k := range before {
		beforeKeys = append(beforeKeys, k)
	}
	sort.Strings(beforeKeys)
	for _, k := range beforeKeys {
		if before[k] != after[k] {
			t.Errorf("file %s changed hash: %s → %s", k, before[k], after[k])
		}
	}
	if _, ok := after["wiki/new.md"]; ok {
		t.Error("content-inspect created the draft on disk")
	}
}

// (e) Boundary escape and non-.md paths are rejected exactly as the live inspect
// path rejects them — the resolution gate is unchanged for content-inspect.
func TestInspectContentRejectsBoundaryAndNonMarkdown(t *testing.T) {
	root := t.TempDir()

	_, err := InspectContent(root, filepath.Join("..", "outside.md"), []byte(citedDraft),
		yamladapter.New(), nil, nil)
	if !errors.Is(err, fsafe.ErrOutsideBoundary) {
		t.Errorf("boundary escape err = %v, want fsafe.ErrOutsideBoundary", err)
	}

	_, err = InspectContent(root, "notes.txt", []byte("hi"), yamladapter.New(), nil, nil)
	if !errors.Is(err, ErrNotMarkdown) {
		t.Errorf("non-md err = %v, want ErrNotMarkdown", err)
	}
}

// A draft whose target path is an existing directory is rejected (not a page).
func TestInspectContentRejectsDirectoryTarget(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "wiki.md"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	_, err := InspectContent(root, "wiki.md", []byte(citedDraft), yamladapter.New(), nil, nil)
	if !errors.Is(err, ErrPageNotFound) {
		t.Errorf("directory target err = %v, want ErrPageNotFound", err)
	}
}

// Malformed draft frontmatter still hashes and reports parse failure, mirroring
// live inspect — normalization is a later plan concern, not an inspect gate.
func TestInspectContentMalformedYAML(t *testing.T) {
	root := t.TempDir()
	bad := "---\ntype: concept\ntitle: {broken\n---\n\n# Bad\n"
	rep, err := InspectContent(root, "wiki/bad.md", []byte(bad), yamladapter.New(), nil, nil)
	if err != nil {
		t.Fatalf("InspectContent: %v", err)
	}
	if rep.Parsed {
		t.Error("Parsed = true, want false for malformed draft YAML")
	}
	parse := 0
	for _, f := range rep.Findings {
		if f.Code == validate.CodeYAMLParse {
			parse++
		}
	}
	if parse != 1 {
		t.Errorf("okf-yaml-parse findings = %d, want 1 (%+v)", parse, rep.Findings)
	}
}
