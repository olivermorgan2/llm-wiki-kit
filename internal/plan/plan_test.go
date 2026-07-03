package plan

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/olivermorgan2/llm-wiki-kit/internal/contract"
	"github.com/olivermorgan2/llm-wiki-kit/internal/fsafe"
	"github.com/olivermorgan2/llm-wiki-kit/internal/validate"
	"github.com/olivermorgan2/llm-wiki-kit/internal/yamladapter"
)

// validPage is a minimal core-profile-clean page body: every required and
// recommended field present, kebab-case filename assumed by the caller.
const validPage = `---
type: concept
title: Sample Page
description: A sample page used in plan tests.
timestamp: 2026-07-03
tags:
  - sample
aliases:
  - sample-page
resource: https://example.com
---

# Sample Page

Body text.
`

// warnPage omits the recommended fields (suggestion) and uses a non-kebab
// filename when written as such; here it drops recommended fields so a warning
// arises from the filename the caller chooses.
const missingRecommendedPage = `---
type: concept
title: Sample Page
description: A sample page missing recommended fields.
---

# Sample Page
`

// writePage writes content to root/rel, creating parent dirs, and returns the
// absolute path.
func writePage(t *testing.T, root, rel, content string) string {
	t.Helper()
	abs := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
	return abs
}

func TestResolvePageRelative(t *testing.T) {
	root := t.TempDir()
	writePage(t, root, "wiki/index.md", validPage)

	ref, err := ResolvePage(root, "wiki/index.md")
	if err != nil {
		t.Fatalf("ResolvePage: %v", err)
	}
	if ref.Rel != "wiki/index.md" {
		t.Errorf("Rel = %q, want wiki/index.md", ref.Rel)
	}
}

func TestResolvePageContainedAbsolute(t *testing.T) {
	root := t.TempDir()
	writePage(t, root, "wiki/index.md", validPage)
	// Build the absolute path under the canonicalized root: fsafe's lexical
	// containment check runs before symlink resolution, so an absolute path
	// threaded through an uncanonicalized prefix (e.g. macOS /var → /private/var)
	// would be refused as an escape. A genuinely-contained absolute must be
	// expressed against the canonical root.
	canonRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatalf("EvalSymlinks: %v", err)
	}
	abs := filepath.Join(canonRoot, "wiki", "index.md")

	ref, err := ResolvePage(root, abs)
	if err != nil {
		t.Fatalf("ResolvePage(abs): %v", err)
	}
	if ref.Rel != "wiki/index.md" {
		t.Errorf("Rel = %q, want wiki/index.md", ref.Rel)
	}
}

func TestResolvePageOutsideBoundary(t *testing.T) {
	root := t.TempDir()
	_, err := ResolvePage(root, filepath.Join("..", "outside.md"))
	if !errors.Is(err, fsafe.ErrOutsideBoundary) {
		t.Errorf("err = %v, want fsafe.ErrOutsideBoundary", err)
	}
}

func TestResolvePageSymlinkEscape(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink semantics differ on windows")
	}
	root := t.TempDir()
	outside := t.TempDir()
	writePage(t, outside, "secret.md", validPage)
	// A symlinked directory inside the boundary that points outside it.
	link := filepath.Join(root, "escape")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	_, err := ResolvePage(root, "escape/secret.md")
	if !errors.Is(err, fsafe.ErrSymlinkEscape) {
		t.Errorf("err = %v, want fsafe.ErrSymlinkEscape", err)
	}
}

func TestResolvePageMissing(t *testing.T) {
	root := t.TempDir()
	_, err := ResolvePage(root, "wiki/nope.md")
	if !errors.Is(err, ErrPageNotFound) {
		t.Errorf("err = %v, want ErrPageNotFound", err)
	}
}

func TestResolvePageDirectory(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "wiki.md"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	_, err := ResolvePage(root, "wiki.md")
	if !errors.Is(err, ErrPageNotFound) {
		t.Errorf("err = %v, want ErrPageNotFound for a directory target", err)
	}
}

func TestResolvePageNotMarkdown(t *testing.T) {
	root := t.TempDir()
	writePage(t, root, "notes.txt", "hello")
	_, err := ResolvePage(root, "notes.txt")
	if !errors.Is(err, ErrNotMarkdown) {
		t.Errorf("err = %v, want ErrNotMarkdown", err)
	}
}

func TestInspectValidPage(t *testing.T) {
	root := t.TempDir()
	abs := writePage(t, root, "wiki/index.md", validPage)

	rep, err := Inspect(root, "wiki/index.md", yamladapter.New(), nil)
	if err != nil {
		t.Fatalf("Inspect: %v", err)
	}
	if !rep.Parsed {
		t.Error("Parsed = false, want true for a well-formed page")
	}
	if len(rep.Findings) != 0 {
		t.Errorf("Findings = %+v, want none", rep.Findings)
	}
	if rep.Status != contract.StatusSuccess {
		t.Errorf("Status = %q, want success", rep.Status)
	}
	// Content hash must be the SHA-256 of the exact bytes on disk.
	raw, err := os.ReadFile(abs)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	sum := sha256.Sum256(raw)
	if want := hex.EncodeToString(sum[:]); rep.ContentHash != want {
		t.Errorf("ContentHash = %q, want %q", rep.ContentHash, want)
	}
	if rep.Path != "wiki/index.md" {
		t.Errorf("Path = %q, want wiki/index.md", rep.Path)
	}
}

func TestInspectWarningsOnly(t *testing.T) {
	root := t.TempDir()
	// Missing recommended fields is a suggestion, not a warning; a non-kebab
	// filename is a warning. Use both to land on success-with-warnings.
	writePage(t, root, "wiki/Index_Page.md", missingRecommendedPage)

	rep, err := Inspect(root, "wiki/Index_Page.md", yamladapter.New(), nil)
	if err != nil {
		t.Fatalf("Inspect: %v", err)
	}
	if rep.Status != contract.StatusSuccessWithWarnings {
		t.Errorf("Status = %q, want success-with-warnings (findings: %+v)", rep.Status, rep.Findings)
	}
	if !rep.Parsed {
		t.Error("Parsed = false, want true")
	}
}

func TestInspectMalformedYAML(t *testing.T) {
	cases := map[string]string{
		"no-leading-marker": "type: concept\ntitle: X\n---\n\n# X\n",
		"unterminated":      "---\ntype: concept\ntitle: X\n\n# X\n",
		"unparseable":       "---\ntype: concept\ntitle: {broken\n---\n\n# X\n",
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			abs := writePage(t, root, "wiki/bad.md", body)

			rep, err := Inspect(root, "wiki/bad.md", yamladapter.New(), nil)
			if err != nil {
				t.Fatalf("Inspect: %v", err)
			}
			if rep.Parsed {
				t.Error("Parsed = true, want false for malformed YAML")
			}
			parseFindings := 0
			for _, f := range rep.Findings {
				if f.Code == validate.CodeYAMLParse {
					parseFindings++
				}
			}
			if parseFindings != 1 {
				t.Errorf("okf-yaml-parse findings = %d, want exactly 1 (%+v)", parseFindings, rep.Findings)
			}
			if rep.Status != contract.StatusValidationFailure {
				t.Errorf("Status = %q, want validation-failure", rep.Status)
			}
			// Malformed YAML still hashes: the bytes exist.
			raw, _ := os.ReadFile(abs)
			sum := sha256.Sum256(raw)
			if want := hex.EncodeToString(sum[:]); rep.ContentHash != want {
				t.Errorf("ContentHash = %q, want %q", rep.ContentHash, want)
			}
		})
	}
}

// Bundle-context parity: broken-link findings depend on the full bundle file
// set, so a link to an existing sibling produces no finding while a link to a
// missing target does — proving whole-bundle filtering, not single-file eval.
func TestInspectBrokenLinkNeedsBundleContext(t *testing.T) {
	linkPage := func(target string) string {
		return "---\ntype: concept\ntitle: Linker\ndescription: Links out.\n" +
			"timestamp: 2026-07-03\ntags: []\naliases: []\nresource: https://x\n---\n\n" +
			"See [other](" + target + ").\n"
	}

	t.Run("existing-sibling", func(t *testing.T) {
		root := t.TempDir()
		// Links resolve bundle-root-relative, so the target is the full
		// bundle path of the sibling, not a directory-relative name.
		writePage(t, root, "wiki/other.md", validPage)
		writePage(t, root, "wiki/linker.md", linkPage("wiki/other.md"))

		rep, err := Inspect(root, "wiki/linker.md", yamladapter.New(), nil)
		if err != nil {
			t.Fatalf("Inspect: %v", err)
		}
		for _, f := range rep.Findings {
			if f.Code == "core-broken-link" {
				t.Errorf("unexpected broken-link finding for existing sibling: %+v", f)
			}
		}
	})

	t.Run("missing-target", func(t *testing.T) {
		root := t.TempDir()
		writePage(t, root, "wiki/linker.md", linkPage("wiki/ghost.md"))

		rep, err := Inspect(root, "wiki/linker.md", yamladapter.New(), nil)
		if err != nil {
			t.Fatalf("Inspect: %v", err)
		}
		found := false
		for _, f := range rep.Findings {
			if f.Code == "core-broken-link" {
				found = true
			}
		}
		if !found {
			t.Errorf("expected core-broken-link finding for missing target; got %+v", rep.Findings)
		}
	})
}

// Overrides apply the ADR-004 profile-override layer: demoting the broken-link
// code changes the filtered finding's severity and the resulting status.
func TestInspectOverridesChangeSeverity(t *testing.T) {
	root := t.TempDir()
	linker := "---\ntype: concept\ntitle: Linker\ndescription: Links out.\n" +
		"timestamp: 2026-07-03\ntags: []\naliases: []\nresource: https://x\n---\n\n" +
		"See [other](ghost.md).\n"
	writePage(t, root, "wiki/linker.md", linker)

	// Default: broken-link is a warning → success-with-warnings.
	base, err := Inspect(root, "wiki/linker.md", yamladapter.New(), nil)
	if err != nil {
		t.Fatalf("Inspect: %v", err)
	}
	if base.Status != contract.StatusSuccessWithWarnings {
		t.Fatalf("baseline Status = %q, want success-with-warnings", base.Status)
	}

	// Promote broken-link to error → validation-failure.
	promoted, err := Inspect(root, "wiki/linker.md", yamladapter.New(),
		map[string]contract.Severity{"core-broken-link": contract.SeverityError})
	if err != nil {
		t.Fatalf("Inspect(override): %v", err)
	}
	if promoted.Status != contract.StatusValidationFailure {
		t.Errorf("overridden Status = %q, want validation-failure", promoted.Status)
	}
	for _, f := range promoted.Findings {
		if f.Code == "core-broken-link" && f.Severity != contract.SeverityError {
			t.Errorf("broken-link severity = %q, want error after override", f.Severity)
		}
	}
}
