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
