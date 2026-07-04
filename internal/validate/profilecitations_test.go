package validate

import (
	"testing"
	"testing/fstest"

	"github.com/olivermorgan2/llm-wiki-kit/internal/contract"
	"github.com/olivermorgan2/llm-wiki-kit/internal/profile"
	"github.com/olivermorgan2/llm-wiki-kit/internal/yamladapter"
)

// citationProfile is a synthetic profile whose `claim` type carries the
// addendum-003 citation obligations: require a citation in `## Evidence` when
// assessment is supported, and forbid a `question` page as an evidence target.
func citationProfile() profile.Profile {
	return profile.Profile{
		ID: "test", Version: "1.0", Extends: "core",
		Types: map[string]profile.TypeRules{
			"claim": {
				EvidenceSections: []string{"Evidence"},
				Citation: &profile.CitationRules{
					RequireWhen:          map[string]string{"assessment": "supported"},
					ForbiddenTargetTypes: []string{"question"},
				},
			},
		},
	}
}

func runCitation(fsys fstest.MapFS) []contract.Finding {
	return NewWithOptions(yamladapter.New(), Options{Profile: citationProfile()}).Run(fsys)
}

// A supported claim with no citation in ## Evidence is a profile-citation-required
// error (ADR-010 sub-decision 3).
func TestCitationRequiredFiresOnSupportedClaimWithNoCitation(t *testing.T) {
	body := "---\ntype: claim\ntitle: C\ndescription: d\ntags: [x]\nconfidence: high\nassessment: supported\n---\n# C\n\n## Evidence\n\nNo links here.\n"
	got := runCitation(fstest.MapFS{"c.md": {Data: []byte(body)}})
	fs := findingsWithCode(got, codeProfileCitationRequired)
	if len(fs) != 1 || fs[0].Severity != contract.SeverityError {
		t.Fatalf("want one profile-citation-required error, got %+v", got)
	}
}

// The obligation is about PRESENCE: a present-but-unresolved citation satisfies
// it (no profile-citation-required) but still raises core-citation-unresolved
// (ADR-008 carry-in №2). One target, two independent axes.
func TestCitationPresentButUnresolvedSatisfiesObligation(t *testing.T) {
	body := "---\ntype: claim\ntitle: C\ndescription: d\ntags: [x]\nconfidence: high\nassessment: supported\n---\n# C\n\n## Evidence\n\nSee [the source](missing-source.md).\n"
	got := runCitation(fstest.MapFS{"c.md": {Data: []byte(body)}})
	if n := len(findingsWithCode(got, codeProfileCitationRequired)); n != 0 {
		t.Errorf("present citation should satisfy the obligation, got %d required findings: %+v", n, got)
	}
	if n := len(findingsWithCode(got, codeCoreCitationUnresolved)); n != 1 {
		t.Errorf("unresolved citation should still warn, got %d core-citation-unresolved: %+v", n, got)
	}
}

// A resolvable citation satisfies the obligation and raises nothing.
func TestCitationResolvableIsClean(t *testing.T) {
	claim := "---\ntype: claim\ntitle: C\ndescription: d\ntags: [x]\nconfidence: high\nassessment: supported\n---\n# C\n\n## Evidence\n\nSee [the source](src.md).\n"
	src := "---\ntype: source\ntitle: S\ndescription: d\ntags: [x]\n---\n# S\n"
	got := runCitation(fstest.MapFS{"c.md": {Data: []byte(claim)}, "src.md": {Data: []byte(src)}})
	for _, code := range []string{codeProfileCitationRequired, codeCoreCitationUnresolved, codeProfileCitationTargetType} {
		if n := len(findingsWithCode(got, code)); n != 0 {
			t.Errorf("resolvable citation should raise no %s, got %d: %+v", code, n, got)
		}
	}
}

// requireWhen not met (assessment != supported) carries no obligation, even with
// an empty evidence section.
func TestCitationNoObligationWhenTriggerFalse(t *testing.T) {
	body := "---\ntype: claim\ntitle: C\ndescription: d\ntags: [x]\nconfidence: high\nassessment: open\n---\n# C\n\n## Evidence\n\nNothing.\n"
	got := runCitation(fstest.MapFS{"c.md": {Data: []byte(body)}})
	if n := len(findingsWithCode(got, codeProfileCitationRequired)); n != 0 {
		t.Errorf("an open claim carries no citation obligation, got %d: %+v", n, got)
	}
}

// A claim citing a `question` page as evidence is a profile-citation-target-type
// error — a question is never evidence (addendum 003).
func TestCitationForbiddenTargetTypeFiresOnQuestion(t *testing.T) {
	claim := "---\ntype: claim\ntitle: C\ndescription: d\ntags: [x]\nconfidence: high\nassessment: supported\n---\n# C\n\n## Evidence\n\nPer [the question](q.md).\n"
	q := "---\ntype: question\ntitle: Q\ndescription: d\ntags: [x]\nstatus: open\n---\n# Q\n\n## Evidence gap\n"
	got := runCitation(fstest.MapFS{"c.md": {Data: []byte(claim)}, "q.md": {Data: []byte(q)}})
	fs := findingsWithCode(got, codeProfileCitationTargetType)
	if len(fs) != 1 || fs[0].Severity != contract.SeverityError {
		t.Fatalf("want one profile-citation-target-type error, got %+v", got)
	}
	// Citing the question also satisfies the presence obligation (a link is present).
	if n := len(findingsWithCode(got, codeProfileCitationRequired)); n != 0 {
		t.Errorf("a present (if forbidden) citation still satisfies presence, got %d required: %+v", n, got)
	}
}

// A citation to a non-forbidden page type (a source) is fine.
func TestCitationAllowedTargetTypeIsClean(t *testing.T) {
	claim := "---\ntype: claim\ntitle: C\ndescription: d\ntags: [x]\nconfidence: high\nassessment: supported\n---\n# C\n\n## Evidence\n\nPer [the source](src.md).\n"
	src := "---\ntype: source\ntitle: S\ndescription: d\ntags: [x]\n---\n# S\n"
	got := runCitation(fstest.MapFS{"c.md": {Data: []byte(claim)}, "src.md": {Data: []byte(src)}})
	if n := len(findingsWithCode(got, codeProfileCitationTargetType)); n != 0 {
		t.Errorf("citing a source is allowed, got %d target-type findings: %+v", n, got)
	}
}

// Links OUTSIDE the evidence section are navigational, not citations: they do not
// satisfy the obligation and are not type-checked.
func TestCitationLinksOutsideEvidenceAreNotCitations(t *testing.T) {
	claim := "---\ntype: claim\ntitle: C\ndescription: d\ntags: [x]\nconfidence: high\nassessment: supported\n---\n# C\n\nSee [q](q.md) up here.\n\n## Evidence\n\nNothing cited.\n"
	q := "---\ntype: question\ntitle: Q\ndescription: d\ntags: [x]\nstatus: open\n---\n# Q\n\n## Evidence gap\n"
	got := runCitation(fstest.MapFS{"c.md": {Data: []byte(claim)}, "q.md": {Data: []byte(q)}})
	// The navigational link doesn't count as a citation → obligation unmet.
	if n := len(findingsWithCode(got, codeProfileCitationRequired)); n != 1 {
		t.Errorf("a link outside ## Evidence must not satisfy the obligation, got %d: %+v", n, got)
	}
	// ...and it isn't a forbidden-target citation either.
	if n := len(findingsWithCode(got, codeProfileCitationTargetType)); n != 0 {
		t.Errorf("a navigational link to a question is not a citation, got %d: %+v", n, got)
	}
}

// ADR-010 carry-in №1: the repo-path resolution class (class 3) is entered only
// for scheme-less intra-wiki `../` escapes; no non-intra-wiki shape (`//`, `#`,
// a non-http scheme, http(s), empty) ever reaches the repo stat.
func TestClassifyIsIntraWikiGatesRepoStat(t *testing.T) {
	var calls []string
	res := &resolver{
		exists:      func(string) bool { return false },
		bundleDir:   "wiki", // so a `../` escape can resolve inside the repo root
		repoResolve: func(rel string) RepoStatus { calls = append(calls, rel); return RepoAbsent },
	}
	for _, target := range []string{"//host/p", "#frag", "mailto:x@y.z", "http://h/p", "https://h", "ftp://h/p", "", "  "} {
		res.classify(target)
	}
	if len(calls) != 0 {
		t.Fatalf("non-intra-wiki targets reached the repo stat: %v", calls)
	}
	// A genuine scheme-less `../` escape that stays inside the repo DOES reach it.
	res.classify("../outside.md")
	if len(calls) != 1 || calls[0] != "outside.md" {
		t.Errorf("intra-wiki ../ escape should reach repoResolve once as outside.md, got %v", calls)
	}
}

// evidenceSectionsFor unions the global (env) override with the profile type's
// evidence sections, deduplicated and order-stable (global first).
func TestEvidenceSectionsForUnion(t *testing.T) {
	opts := Options{
		EvidenceSections: []string{"Sources", "Evidence"},
		Profile: profile.Profile{Types: map[string]profile.TypeRules{
			"claim": {EvidenceSections: []string{"Evidence", "Proof"}},
		}},
	}
	got := evidenceSectionsFor(opts, "claim")
	want := []string{"Sources", "Evidence", "Proof"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
	// A type with no profile evidence sections and no env override → nil.
	if s := evidenceSectionsFor(Options{}, "source"); s != nil {
		t.Errorf("zero options should yield nil evidence sections, got %v", s)
	}
}
