package validate

import (
	"testing"

	"github.com/olivermorgan2/llm-wiki-kit/internal/contract"
	"github.com/olivermorgan2/llm-wiki-kit/internal/yamladapter"
)

// firstWithCode returns the first finding with the given code, or a zero
// Finding and false.
func firstWithCode(fs []contract.Finding, code string) (contract.Finding, bool) {
	for _, f := range fs {
		if f.Code == code {
			return f, true
		}
	}
	return contract.Finding{}, false
}

func TestCoreRequiredTitleMissingIsProfileError(t *testing.T) {
	fs := evaluatePage(yamladapter.New(), "alpha.md", []byte("---\ntype: concept\ndescription: d\n---\n"))
	f, ok := firstWithCode(fs, "core-required-title")
	if !ok {
		t.Fatalf("missing title should raise core-required-title, got %+v", fs)
	}
	if f.Ruleset != contract.RulesetProfile || f.Severity != contract.SeverityError {
		t.Errorf("ruleset/severity = %q/%q, want profile/error", f.Ruleset, f.Severity)
	}
}

func TestCoreRequiredTitleEmptyIsProfileError(t *testing.T) {
	fs := evaluatePage(yamladapter.New(), "alpha.md", []byte("---\ntype: concept\ntitle: \"  \"\ndescription: d\n---\n"))
	if _, ok := firstWithCode(fs, "core-required-title"); !ok {
		t.Fatalf("empty title should raise core-required-title, got %+v", fs)
	}
}

func TestCoreRequiredDescriptionMissingIsProfileError(t *testing.T) {
	fs := evaluatePage(yamladapter.New(), "alpha.md", []byte("---\ntype: concept\ntitle: Alpha\n---\n"))
	f, ok := firstWithCode(fs, "core-required-description")
	if !ok {
		t.Fatalf("missing description should raise core-required-description, got %+v", fs)
	}
	if f.Ruleset != contract.RulesetProfile || f.Severity != contract.SeverityError {
		t.Errorf("ruleset/severity = %q/%q, want profile/error", f.Ruleset, f.Severity)
	}
}

func TestCoreFieldTypeWrongScalarIsProfileError(t *testing.T) {
	// title declared as a sequence where a scalar string is required.
	fs := evaluatePage(yamladapter.New(), "alpha.md", []byte("---\ntype: concept\ntitle:\n  - a\n  - b\ndescription: d\n---\n"))
	f, ok := firstWithCode(fs, "core-field-type")
	if !ok {
		t.Fatalf("sequence-valued title should raise core-field-type, got %+v", fs)
	}
	if f.Ruleset != contract.RulesetProfile || f.Severity != contract.SeverityError {
		t.Errorf("ruleset/severity = %q/%q, want profile/error", f.Ruleset, f.Severity)
	}
	// A wrong-typed title must not also raise the required-title rule (rulesets
	// stay non-overlapping; one problem, one finding).
	if _, ok := firstWithCode(fs, "core-required-title"); ok {
		t.Errorf("wrong-typed title should not also raise core-required-title: %+v", fs)
	}
}

func TestCoreFieldTypeWrongSequenceIsProfileError(t *testing.T) {
	// tags declared as a scalar where a sequence is required.
	fs := evaluatePage(yamladapter.New(), "alpha.md", []byte("---\ntype: concept\ntitle: Alpha\ndescription: d\ntags: notalist\n---\n"))
	if _, ok := firstWithCode(fs, "core-field-type"); !ok {
		t.Fatalf("scalar-valued tags should raise core-field-type, got %+v", fs)
	}
}

func TestCoreRecommendedMissingIsSuggestion(t *testing.T) {
	// All required fields present, but no recommended fields.
	fs := evaluatePage(yamladapter.New(), "alpha.md", []byte("---\ntype: concept\ntitle: Alpha\ndescription: d\n---\n"))
	f, ok := firstWithCode(fs, "core-recommended-missing")
	if !ok {
		t.Fatalf("absent recommended fields should raise a suggestion, got %+v", fs)
	}
	if f.Ruleset != contract.RulesetProfile || f.Severity != contract.SeveritySuggestion {
		t.Errorf("ruleset/severity = %q/%q, want profile/suggestion", f.Ruleset, f.Severity)
	}
	// Exactly one aggregate suggestion per page (unique fingerprint).
	if got := findingsWithCode(fs, "core-recommended-missing"); len(got) != 1 {
		t.Errorf("want a single aggregate recommended-missing finding, got %d", len(got))
	}
}

func TestCoreRecommendedPresentNoSuggestion(t *testing.T) {
	fs := evaluatePage(yamladapter.New(), "alpha.md", []byte("---\ntype: concept\ntitle: Alpha\ndescription: d\ntimestamp: 2026-01-01\ntags: [x]\naliases: [y]\nresource: r\n---\n"))
	if _, ok := firstWithCode(fs, "core-recommended-missing"); ok {
		t.Errorf("all recommended fields present should not raise a suggestion: %+v", fs)
	}
}

func TestCoreKebabFilenameNonKebabIsWarning(t *testing.T) {
	fs := evaluatePage(yamladapter.New(), "My_Page.md", []byte("---\ntype: concept\ntitle: Alpha\ndescription: d\ntags: [x]\n---\n"))
	f, ok := firstWithCode(fs, "core-kebab-filename")
	if !ok {
		t.Fatalf("non-kebab filename should raise a warning, got %+v", fs)
	}
	if f.Ruleset != contract.RulesetProfile || f.Severity != contract.SeverityWarning {
		t.Errorf("ruleset/severity = %q/%q, want profile/warning", f.Ruleset, f.Severity)
	}
}

func TestCoreKebabFilenameKebabOK(t *testing.T) {
	fs := evaluatePage(yamladapter.New(), "alpha-beta.md", []byte("---\ntype: concept\ntitle: Alpha\ndescription: d\ntags: [x]\n---\n"))
	if _, ok := firstWithCode(fs, "core-kebab-filename"); ok {
		t.Errorf("kebab-case filename should not warn: %+v", fs)
	}
}

// Criterion 5: OKF and profile findings are reported separately, each tagged
// with its ruleset. A page missing both `type` (OKF) and `title` (profile)
// yields one finding under each ruleset with the right code.
func TestOKFAndProfileFindingsTaggedDistinctly(t *testing.T) {
	fs := evaluatePage(yamladapter.New(), "alpha.md", []byte("---\ndescription: d\ntags: [x]\n---\n"))
	okf, ok := firstWithCode(fs, "okf-type-present")
	if !ok || okf.Ruleset != contract.RulesetOKF {
		t.Fatalf("expected an OKF-tagged okf-type-present finding, got %+v", fs)
	}
	prof, ok := firstWithCode(fs, "core-required-title")
	if !ok || prof.Ruleset != contract.RulesetProfile {
		t.Fatalf("expected a profile-tagged core-required-title finding, got %+v", fs)
	}
}
