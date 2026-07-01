package validate

import (
	"testing"

	"github.com/olivermorgan2/llm-wiki-kit/internal/contract"
)

func sampleFindings() []contract.Finding {
	return []contract.Finding{
		{Ruleset: contract.RulesetProfile, Severity: contract.SeverityError, Code: codeCoreReqTitle, Path: "a.md"},
		{Ruleset: contract.RulesetProfile, Severity: contract.SeverityWarning, Code: codeCoreKebab, Path: "a.md"},
		{Ruleset: contract.RulesetProfile, Severity: contract.SeveritySuggestion, Code: codeCoreRecommended, Path: "a.md"},
	}
}

// Core-only runs pass nil overrides; Resolve is the identity in that case.
func TestResolveNilOverridesIsIdentity(t *testing.T) {
	in := sampleFindings()
	got := Resolve(in, nil)
	if len(got) != len(in) {
		t.Fatalf("Resolve changed finding count: %d != %d", len(got), len(in))
	}
	for i := range in {
		if got[i].Severity != in[i].Severity || got[i].Code != in[i].Code {
			t.Errorf("finding %d changed under nil overrides: %+v", i, got[i])
		}
	}
}

// A profile override may promote a warning to an error.
func TestResolvePromotesSeverity(t *testing.T) {
	got := Resolve(sampleFindings(), map[string]contract.Severity{codeCoreKebab: contract.SeverityError})
	f, ok := firstWithCode(got, codeCoreKebab)
	if !ok {
		t.Fatal("kebab finding missing after resolve")
	}
	if f.Severity != contract.SeverityError {
		t.Errorf("severity = %q, want error (promoted)", f.Severity)
	}
	// Non-overridden findings are untouched.
	if other, _ := firstWithCode(got, codeCoreReqTitle); other.Severity != contract.SeverityError {
		t.Errorf("unrelated finding changed: %+v", other)
	}
}

// A profile override may demote an error to a warning.
func TestResolveDemotesSeverity(t *testing.T) {
	got := Resolve(sampleFindings(), map[string]contract.Severity{codeCoreReqTitle: contract.SeverityWarning})
	f, _ := firstWithCode(got, codeCoreReqTitle)
	if f.Severity != contract.SeverityWarning {
		t.Errorf("severity = %q, want warning (demoted)", f.Severity)
	}
}

// Overrides change severity only; they never add or remove findings.
func TestResolveNeverChangesCount(t *testing.T) {
	got := Resolve(sampleFindings(), map[string]contract.Severity{
		codeCoreKebab:    contract.SeverityError,
		"unmatched-code": contract.SeverityError,
	})
	if len(got) != 3 {
		t.Errorf("count = %d, want 3 (overrides must not add/remove findings)", len(got))
	}
}
