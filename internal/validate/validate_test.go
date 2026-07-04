package validate

import (
	"testing"
	"testing/fstest"

	"github.com/olivermorgan2/llm-wiki-kit/internal/contract"
	"github.com/olivermorgan2/llm-wiki-kit/internal/yamladapter"
)

func engine() *Engine { return New(yamladapter.New()) }

func page(body string) *fstest.MapFile { return &fstest.MapFile{Data: []byte(body)} }

const validPage = "---\ntype: concept\ntitle: Alpha\ndescription: A page.\ntimestamp: 2026-01-01\ntags: [x]\naliases: [y]\nresource: r\n---\n# Alpha\n"

func TestRunReturnsNonNil(t *testing.T) {
	got := engine().Run(fstest.MapFS{})
	if got == nil {
		t.Fatal("Run must return a non-nil slice")
	}
}

func TestRunCompleteValidPageHasNoFindings(t *testing.T) {
	fsys := fstest.MapFS{"alpha-page.md": page(validPage)}
	if got := engine().Run(fsys); len(got) != 0 {
		t.Fatalf("complete valid page should yield no findings, got %+v", got)
	}
}

func TestRunReportsMissingRequiredField(t *testing.T) {
	fsys := fstest.MapFS{"alpha-page.md": page("---\ntype: concept\ndescription: d\ntags: [x]\n---\n")}
	got := engine().Run(fsys)
	if _, ok := firstWithCode(got, codeCoreReqTitle); !ok {
		t.Errorf("missing title should be reported, got %+v", got)
	}
}

// The walk is deterministic and recurses into subdirectories; only .md files
// are validated.
func TestRunWalksDeterministicallyAndFiltersMarkdown(t *testing.T) {
	fsys := fstest.MapFS{
		"b.md":         page("---\n---\n"),
		"a.md":         page("---\n---\n"),
		"sub/c.md":     page("---\n---\n"),
		"notes.txt":    page("not markdown"),
		"sub/data.yml": page("type: concept"),
	}
	got := engine().Run(fsys)

	// Three .md files, each missing type -> at least one okf-type-present each.
	var paths []string
	for _, f := range got {
		if f.Code == codeOKFTypePresent {
			paths = append(paths, f.Path)
		}
	}
	want := []string{"a.md", "b.md", "sub/c.md"}
	if len(paths) != len(want) {
		t.Fatalf("okf-type-present paths = %v, want %v", paths, want)
	}
	for i := range want {
		if paths[i] != want[i] {
			t.Errorf("path[%d] = %q, want %q (order must be deterministic)", i, paths[i], want[i])
		}
	}
	// The .txt and .yml files must not be validated.
	for _, f := range got {
		if f.Path == "notes.txt" || f.Path == "sub/data.yml" {
			t.Errorf("non-markdown file was validated: %+v", f)
		}
	}
}

// A broken intra-wiki link surfaces through Run at its page, resolved against
// the bundle path-set collected during the walk; a page whose links all resolve
// raises none. (issue #4)
func TestRunReportsBrokenIntraWikiLink(t *testing.T) {
	fsys := fstest.MapFS{
		"concepts/alpha.md": page(validPage + "\nSee [real](claims/real.md) and [gone](claims/missing.md).\n"),
		"claims/real.md":    page(validPage),
	}
	got := engine().Run(fsys)
	broken := findingsWithCode(got, codeCoreBrokenLink)
	if len(broken) != 1 {
		t.Fatalf("want exactly one core-broken-link finding, got %d (%+v)", len(broken), got)
	}
	if broken[0].Path != "concepts/alpha.md" {
		t.Errorf("finding path = %q, want concepts/alpha.md", broken[0].Path)
	}
}

func TestRunAllLinksValidNoBrokenLinkFinding(t *testing.T) {
	fsys := fstest.MapFS{
		"concepts/alpha.md": page(validPage + "\nSee [real](claims/real.md).\n"),
		"claims/real.md":    page(validPage),
	}
	if got := engine().Run(fsys); len(findingsWithCode(got, codeCoreBrokenLink)) != 0 {
		t.Errorf("all-valid links should yield no broken-link finding, got %+v", got)
	}
}

// NewWithOptions surfaces citation findings end-to-end when EvidenceSections is
// set, and a page whose frontmatter fails to split yields only okf-yaml-parse —
// the citation rules inherit the parse-failure gate.
func TestRunWithOptionsEmitsCitationFindings(t *testing.T) {
	opts := Options{EvidenceSections: []string{"Evidence"}}
	fsys := fstest.MapFS{
		"good.md": page(validPage + "\n## Evidence\nSee [x](claims/missing.md).\n"),
		"bad.md":  page("---\ntype: concept\ntitle: {broken\n## Evidence\n[x](gone.md)\n"),
	}
	got := NewWithOptions(yamladapter.New(), opts).Run(fsys)

	unres := findingsWithCode(got, codeCoreCitationUnresolved)
	if len(unres) != 1 || unres[0].Path != "good.md" {
		t.Fatalf("want one core-citation-unresolved on good.md, got %+v", got)
	}
	// bad.md fails to parse: only okf-yaml-parse, no citation finding.
	for _, f := range got {
		if f.Path == "bad.md" && f.Code != CodeYAMLParse {
			t.Errorf("parse-failed page must yield only okf-yaml-parse, got %+v", f)
		}
	}
}

// End-to-end through the precedence layers: a malformed page exits validation-
// failure; the parse error survives baseline in release-gate mode.
func TestRunThroughPrecedenceMalformedIsValidationFailure(t *testing.T) {
	fsys := fstest.MapFS{"bad.md": page("---\ntype: concept\ntitle: {broken\n---\n")}
	got := engine().Run(fsys)
	got = Resolve(got, nil)
	got = ApplyBaseline(got, nil, true)
	if StatusFor(got) != contract.StatusValidationFailure {
		t.Errorf("malformed page should be validation-failure, got %q (%+v)", StatusFor(got), got)
	}
}
