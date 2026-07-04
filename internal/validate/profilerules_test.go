package validate

import (
	"testing"
	"testing/fstest"

	"github.com/olivermorgan2/llm-wiki-kit/internal/contract"
	"github.com/olivermorgan2/llm-wiki-kit/internal/profile"
	"github.com/olivermorgan2/llm-wiki-kit/internal/yamladapter"
)

// testProfile is a synthetic academic-research-shaped profile exercising every
// ADR-010 per-type rule kind. The real academic-research profile ships in #57
// (I5); the engine is tested against this synthetic profile so I3's rule kinds
// stand alone.
func testProfile() profile.Profile {
	return profile.Profile{
		ID: "test", Version: "1.0", Extends: "core",
		Types: map[string]profile.TypeRules{
			"source": {
				Required: []string{"authors", "source_type"},
				Enums: map[string][]string{
					"source_type": {"paper", "preprint", "book", "dataset", "webpage", "report", "other"},
				},
				ListMin:          map[string]int{"authors": 1},
				RecommendedAnyOf: [][]string{{"doi", "canonical_url"}},
			},
			"claim": {
				Required: []string{"confidence", "assessment"},
				Enums: map[string][]string{
					"confidence": {"low", "medium", "high"},
					"assessment": {"supported", "contested", "refuted", "open"},
				},
				RequiredSections: []string{"Evidence", "Counterevidence", "Assessment"},
			},
		},
	}
}

// run validates a single page under the test profile and returns its findings.
func runProfile(t *testing.T, name, content string) []contract.Finding {
	t.Helper()
	e := NewWithOptions(yamladapter.New(), Options{Profile: testProfile()})
	return e.Run(fstest.MapFS{name: {Data: []byte(content)}})
}

// A page whose type is not a profiled type receives no profile-* findings
// (ADR-010 sub-decision 5: unknown types stay accepted).
func TestProfileRulesSkipUnprofiledType(t *testing.T) {
	got := runProfile(t, "c.md", "---\ntype: concept\ntitle: C\ndescription: d\ntags: [x]\n---\n# C\n")
	for _, f := range got {
		if f.Ruleset == contract.RulesetProfile && isProfileDataCode(f.Code) {
			t.Errorf("unprofiled type produced profile-data finding %q", f.Code)
		}
	}
}

// A zero Profile (the pre-Phase-4 engine) runs no per-type rules even for a name
// that would otherwise be profiled.
func TestProfileRulesZeroProfileIsNoop(t *testing.T) {
	e := NewWithOptions(yamladapter.New(), Options{})
	got := e.Run(fstest.MapFS{"s.md": {Data: []byte("---\ntype: source\ntitle: S\ndescription: d\ntags: [x]\n---\n")}})
	for _, f := range got {
		if isProfileDataCode(f.Code) {
			t.Errorf("zero profile produced profile-data finding %q", f.Code)
		}
	}
}

// Required field absent → exactly one profile-required-field error; a
// present-but-empty list does NOT trip it (it falls to list-min).
func TestProfileRequiredField(t *testing.T) {
	// source_type absent (authors present, valid).
	got := runProfile(t, "s.md", "---\ntype: source\ntitle: S\ndescription: d\ntags: [x]\nauthors: [A]\n---\n")
	if fs := findingsWithCode(got, codeProfileRequiredField); len(fs) != 1 || fs[0].Severity != contract.SeverityError {
		t.Fatalf("want one required-field error, got %+v", got)
	}
	// authors present but empty list → required-field satisfied, list-min fires instead.
	got = runProfile(t, "s.md", "---\ntype: source\ntitle: S\ndescription: d\ntags: [x]\nauthors: []\nsource_type: paper\n---\n")
	if len(findingsWithCode(got, codeProfileRequiredField)) != 0 {
		t.Errorf("empty list should not trip required-field: %+v", got)
	}
	if len(findingsWithCode(got, codeProfileListMin)) != 1 {
		t.Errorf("empty list should trip list-min: %+v", got)
	}
}

// Enum: an out-of-set value trips profile-field-enum; a valid value does not; a
// non-scalar present value is also a violation.
func TestProfileEnum(t *testing.T) {
	bad := runProfile(t, "s.md", "---\ntype: source\ntitle: S\ndescription: d\ntags: [x]\nauthors: [A]\nsource_type: journal\n---\n")
	if fs := findingsWithCode(bad, codeProfileFieldEnum); len(fs) != 1 || fs[0].Severity != contract.SeverityError {
		t.Fatalf("want one enum error, got %+v", bad)
	}
	ok := runProfile(t, "s.md", "---\ntype: source\ntitle: S\ndescription: d\ntags: [x]\nauthors: [A]\nsource_type: paper\n---\n")
	if len(findingsWithCode(ok, codeProfileFieldEnum)) != 0 {
		t.Errorf("valid enum should not fire: %+v", ok)
	}
	nonScalar := runProfile(t, "s.md", "---\ntype: source\ntitle: S\ndescription: d\ntags: [x]\nauthors: [A]\nsource_type: [paper]\n---\n")
	if len(findingsWithCode(nonScalar, codeProfileFieldEnum)) != 1 {
		t.Errorf("non-scalar enum value should be a violation: %+v", nonScalar)
	}
}

// List-min: a present non-list and a too-short list both trip list-min; a list of
// >= N does not.
func TestProfileListMin(t *testing.T) {
	str := runProfile(t, "s.md", "---\ntype: source\ntitle: S\ndescription: d\ntags: [x]\nauthors: single\nsource_type: paper\n---\n")
	if len(findingsWithCode(str, codeProfileListMin)) != 1 {
		t.Errorf("non-list authors should trip list-min: %+v", str)
	}
	ok := runProfile(t, "s.md", "---\ntype: source\ntitle: S\ndescription: d\ntags: [x]\nauthors: [A, B]\nsource_type: paper\n---\n")
	if len(findingsWithCode(ok, codeProfileListMin)) != 0 {
		t.Errorf(">=1 authors should not trip list-min: %+v", ok)
	}
}

// Required sections: a missing section trips profile-required-section; all present
// does not. Matching is by heading title at any level.
func TestProfileRequiredSection(t *testing.T) {
	body := "---\ntype: claim\ntitle: C\ndescription: d\ntags: [x]\nconfidence: high\nassessment: open\n---\n" +
		"# C\n\n## Evidence\n\n## Assessment\n" // Counterevidence missing
	got := runProfile(t, "c.md", body)
	fs := findingsWithCode(got, codeProfileRequiredSection)
	if len(fs) != 1 || fs[0].Severity != contract.SeverityError {
		t.Fatalf("want one required-section error, got %+v", got)
	}

	full := "---\ntype: claim\ntitle: C\ndescription: d\ntags: [x]\nconfidence: high\nassessment: open\n---\n" +
		"# C\n\n## Evidence\n\n## Counterevidence\n\nNone found.\n\n## Assessment\n\nSettled.\n"
	if len(findingsWithCode(runProfile(t, "c.md", full), codeProfileRequiredSection)) != 0 {
		t.Errorf("all sections present should not fire: %+v", full)
	}
}

// Recommended any-of: neither member present → one suggestion; one present → none.
func TestProfileRecommendedAnyOf(t *testing.T) {
	none := runProfile(t, "s.md", "---\ntype: source\ntitle: S\ndescription: d\ntags: [x]\nauthors: [A]\nsource_type: paper\n---\n")
	fs := findingsWithCode(none, codeProfileRecommendedPair)
	if len(fs) != 1 || fs[0].Severity != contract.SeveritySuggestion {
		t.Fatalf("want one recommended-pair suggestion, got %+v", none)
	}
	withDOI := runProfile(t, "s.md", "---\ntype: source\ntitle: S\ndescription: d\ntags: [x]\nauthors: [A]\nsource_type: paper\ndoi: 10.1/x\n---\n")
	if len(findingsWithCode(withDOI, codeProfileRecommendedPair)) != 0 {
		t.Errorf("a present doi satisfies the any-of group: %+v", withDOI)
	}
}

// A page can trip several rule kinds at once, each aggregated into one finding.
func TestProfileMultipleViolationsOnePerCode(t *testing.T) {
	// source missing authors (required) + bad enum + no doi/url (suggestion).
	got := runProfile(t, "s.md", "---\ntype: source\ntitle: S\ndescription: d\ntags: [x]\nsource_type: journal\n---\n")
	if n := len(findingsWithCode(got, codeProfileRequiredField)); n != 1 {
		t.Errorf("required-field count = %d, want 1", n)
	}
	if n := len(findingsWithCode(got, codeProfileFieldEnum)); n != 1 {
		t.Errorf("enum count = %d, want 1", n)
	}
	if n := len(findingsWithCode(got, codeProfileRecommendedPair)); n != 1 {
		t.Errorf("recommended-pair count = %d, want 1", n)
	}
}

// Severity overrides still compose via Resolve (ADR-004 profile-override layer):
// a profile can promote profile-recommended-pair from suggestion to error.
func TestProfileSeverityComposesViaResolve(t *testing.T) {
	got := runProfile(t, "s.md", "---\ntype: source\ntitle: S\ndescription: d\ntags: [x]\nauthors: [A]\nsource_type: paper\n---\n")
	resolved := Resolve(got, map[string]contract.Severity{codeProfileRecommendedPair: contract.SeverityError})
	fs := findingsWithCode(resolved, codeProfileRecommendedPair)
	if len(fs) != 1 || fs[0].Severity != contract.SeverityError {
		t.Fatalf("override should promote recommended-pair to error, got %+v", resolved)
	}
}

// All profile-data findings are tagged ruleset:profile so OKF/profile reporting
// stays separate (criterion 5).
func TestProfileFindingsAreRulesetProfile(t *testing.T) {
	got := runProfile(t, "s.md", "---\ntype: source\ntitle: S\ndescription: d\ntags: [x]\nsource_type: journal\n---\n")
	for _, f := range got {
		if isProfileDataCode(f.Code) && f.Ruleset != contract.RulesetProfile {
			t.Errorf("finding %q ruleset = %v, want profile", f.Code, f.Ruleset)
		}
	}
}

// isProfileDataCode reports whether code is one of the ADR-010 profile-data rule
// codes (used by tests to isolate them from OKF/core findings).
func isProfileDataCode(code string) bool {
	switch code {
	case codeProfileRequiredField, codeProfileFieldEnum, codeProfileListMin,
		codeProfileRequiredSection, codeProfileRecommendedPair:
		return true
	}
	return false
}
