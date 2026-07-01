package validate

import (
	"testing"

	"github.com/olivermorgan2/llm-wiki-kit/internal/contract"
	"github.com/olivermorgan2/llm-wiki-kit/internal/yamladapter"
)

// findingsFor filters findings to those matching a code.
func findingsWithCode(fs []contract.Finding, code string) []contract.Finding {
	var out []contract.Finding
	for _, f := range fs {
		if f.Code == code {
			out = append(out, f)
		}
	}
	return out
}

func TestEvaluatePageValidHasNoOKFFindings(t *testing.T) {
	fs := evaluatePage(yamladapter.New(), "alpha.md", []byte("---\ntype: concept\ntitle: Alpha\ndescription: A page.\n---\n"))
	for _, f := range fs {
		if f.Ruleset == contract.RulesetOKF {
			t.Errorf("valid page produced an OKF finding: %+v", f)
		}
	}
}

func TestEvaluatePageMissingTypeIsOKFError(t *testing.T) {
	fs := evaluatePage(yamladapter.New(), "alpha.md", []byte("---\ntitle: Alpha\ndescription: A page.\n---\n"))
	got := findingsWithCode(fs, "okf-type-present")
	if len(got) != 1 {
		t.Fatalf("want 1 okf-type-present finding, got %d (%+v)", len(got), fs)
	}
	if got[0].Ruleset != contract.RulesetOKF {
		t.Errorf("ruleset = %q, want okf", got[0].Ruleset)
	}
	if got[0].Severity != contract.SeverityError {
		t.Errorf("severity = %q, want error", got[0].Severity)
	}
	if got[0].Path != "alpha.md" {
		t.Errorf("path = %q, want alpha.md", got[0].Path)
	}
}

func TestEvaluatePageEmptyTypeIsOKFError(t *testing.T) {
	fs := evaluatePage(yamladapter.New(), "alpha.md", []byte("---\ntype: \"\"\ntitle: Alpha\ndescription: d\n---\n"))
	if len(findingsWithCode(fs, "okf-type-present")) != 1 {
		t.Fatalf("empty type should raise okf-type-present, got %+v", fs)
	}
}

func TestEvaluatePageMalformedYAMLIsOKFParseError(t *testing.T) {
	fs := evaluatePage(yamladapter.New(), "bad.md", []byte("---\ntype: concept\ntitle: {broken\n---\n"))
	got := findingsWithCode(fs, "okf-yaml-parse")
	if len(got) != 1 {
		t.Fatalf("want 1 okf-yaml-parse finding, got %d (%+v)", len(got), fs)
	}
	if got[0].Ruleset != contract.RulesetOKF || got[0].Severity != contract.SeverityError {
		t.Errorf("parse finding ruleset/severity = %q/%q, want okf/error", got[0].Ruleset, got[0].Severity)
	}
	// On a parse failure, no other rules run for the page.
	if len(fs) != 1 {
		t.Errorf("parse failure should short-circuit to a single finding, got %+v", fs)
	}
}

func TestEvaluatePageUnterminatedFrontmatterIsOKFParseError(t *testing.T) {
	fs := evaluatePage(yamladapter.New(), "bad.md", []byte("---\ntype: concept\n"))
	if len(findingsWithCode(fs, "okf-yaml-parse")) != 1 {
		t.Fatalf("unterminated frontmatter should raise okf-yaml-parse, got %+v", fs)
	}
}
