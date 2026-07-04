package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/olivermorgan2/llm-wiki-kit/internal/contract"
)

// withStdin swaps os.Stdin for a temp file holding content for the duration of
// fn, so a `--content -` invocation reads the piped draft. run() reads os.Stdin
// directly, so the swap is the seam.
func withStdin(t *testing.T, content string, fn func()) {
	t.Helper()
	p := filepath.Join(t.TempDir(), "stdin.md")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write stdin file: %v", err)
	}
	f, err := os.Open(p)
	if err != nil {
		t.Fatalf("open stdin file: %v", err)
	}
	defer f.Close()
	saved := os.Stdin
	os.Stdin = f
	defer func() { os.Stdin = saved }()
	fn()
}

// draftWithMalformedCitation cites a malformed target inside its evidence context.
const cliMalformedCitationDraft = "---\ntype: concept\ntitle: Bad Cite\ndescription: A draft.\n" +
	"timestamp: 2026-07-04\ntags: []\naliases: []\nresource: https://example.com\n---\n\n" +
	"# Bad Cite\n\n## Evidence\n\nSee [x](mailto:nobody@example.com).\n"

// page inspect --content <file> validates a new-page draft (absent live target):
// exit reflects the draft findings, and the page payload carries the draft's hash.
func TestPageInspectContentNewPageFromFile(t *testing.T) {
	t.Setenv("LLM_WIKI_EVIDENCE_SECTIONS", "Evidence")
	dir := initBundle(t)
	cf := writeContentFile(t, cliMalformedCitationDraft)

	stdout, _, code := exec(t, "page", "inspect", "--root", dir, "wiki/new.md", "--content", cf, "--json")
	// A malformed citation is a warning → success-with-warnings (exit 1).
	if code != int(contract.ExitSuccessWithWarnings) {
		t.Fatalf("page inspect --content exit = %d, want 1\n%s", code, stdout)
	}
	env := decodeEnvelope(t, stdout)
	if env.Operation != "page inspect" {
		t.Errorf("operation = %q, want \"page inspect\"", env.Operation)
	}
	if env.Page == nil || env.Page.Path != "wiki/new.md" {
		t.Fatalf("page payload = %+v, want path wiki/new.md", env.Page)
	}
	// The draft page is absent on disk; content-inspect must not have created it.
	if _, err := os.Stat(filepath.Join(dir, "wiki", "new.md")); !os.IsNotExist(err) {
		t.Errorf("content-inspect created the draft on disk: %v", err)
	}
	has := false
	for _, f := range env.Findings {
		if strings.HasPrefix(f.Code, "core-citation-") {
			has = true
		}
	}
	if !has {
		t.Errorf("expected a core-citation-* finding over the draft; got %+v", env.Findings)
	}
}

// page inspect --content - reads the draft from stdin.
func TestPageInspectContentFromStdin(t *testing.T) {
	dir := initBundle(t)
	draft := "---\ntype: concept\ntitle: Piped\ndescription: Via stdin.\n" +
		"timestamp: 2026-07-04\ntags: []\naliases: []\nresource: https://x\n---\n\n# Piped\n"

	var stdout string
	var code int
	withStdin(t, draft, func() {
		stdout, _, code = exec(t, "page", "inspect", "--root", dir, "wiki/piped.md", "--content", "-", "--json")
	})
	if code != int(contract.ExitSuccess) {
		t.Fatalf("stdin content-inspect exit = %d, want 0\n%s", code, stdout)
	}
	env := decodeEnvelope(t, stdout)
	if env.Page == nil || env.Page.Path != "wiki/piped.md" {
		t.Errorf("page payload = %+v, want path wiki/piped.md", env.Page)
	}
	if len(env.Findings) != 0 {
		t.Errorf("clean piped draft findings = %+v, want none", env.Findings)
	}
}

// A draft that introduces a broken link surfaces the finding even though no live
// page exists — proving the engine ran over the proposed content in bundle context.
func TestPageInspectContentSeesDraftBrokenLink(t *testing.T) {
	dir := initBundle(t)
	draft := "---\ntype: concept\ntitle: Linker\ndescription: Links out.\n" +
		"timestamp: 2026-07-04\ntags: []\naliases: []\nresource: https://x\n---\n\nSee [ghost](wiki/ghost.md).\n"
	cf := writeContentFile(t, draft)

	stdout, _, _ := exec(t, "page", "inspect", "--root", dir, "wiki/linker.md", "--content", cf, "--json")
	env := decodeEnvelope(t, stdout)
	found := false
	for _, f := range env.Findings {
		if f.Code == "core-broken-link" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a core-broken-link finding from the draft; got %+v", env.Findings)
	}
}

// Regression: inspect without --content still requires the live page — a missing
// page with no --content is invalid-invocation (exit 4), unchanged from today.
func TestPageInspectMissingPageStillExit4WithoutContent(t *testing.T) {
	dir := initBundle(t)
	_, _, code := exec(t, "page", "inspect", "--root", dir, "wiki/nope.md", "--json")
	if code != int(contract.ExitInvalidInvocation) {
		t.Errorf("missing page without --content exit = %d, want 4", code)
	}
}

// Regression: an unknown flag on page inspect is still invalid-invocation.
func TestPageInspectUnknownFlagRegression(t *testing.T) {
	dir := initBundle(t)
	_, _, code := exec(t, "page", "inspect", "--root", dir, "--bogus", "wiki/index.md", "--json")
	if code != int(contract.ExitInvalidInvocation) {
		t.Errorf("unknown flag exit = %d, want 4", code)
	}
}

// --content with no value is a bad invocation (exit 4).
func TestPageInspectContentRequiresValue(t *testing.T) {
	dir := initBundle(t)
	_, _, code := exec(t, "page", "inspect", "--root", dir, "wiki/x.md", "--content")
	if code != int(contract.ExitInvalidInvocation) {
		t.Errorf("--content with no value exit = %d, want 4", code)
	}
}
