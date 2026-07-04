package profile

import (
	"errors"
	"testing"

	"github.com/olivermorgan2/llm-wiki-kit/internal/yamladapter"
)

// Resolve returns the shipped core profile for its id, carrying the pinned
// version the bundle config records. Core extends nothing.
func TestResolveCoreReturnsShippedProfile(t *testing.T) {
	p, err := Resolve(CoreID)
	if err != nil {
		t.Fatalf("Resolve(%q) returned error: %v", CoreID, err)
	}
	if p.ID != "core" {
		t.Errorf("ID = %q, want core", p.ID)
	}
	if p.Version != "0.1.0" {
		t.Errorf("Version = %q, want 0.1.0 (scaffold config parity)", p.Version)
	}
	if p.Extends != "" {
		t.Errorf("Extends = %q, want empty (core extends nothing)", p.Extends)
	}
}

// academic-research now resolves (it did not before I2): its merged view carries
// its own identity, extends core, and — being minimal at I2 — has no per-type
// rules yet (#57/I5 fills them).
func TestResolveAcademicResearchReturnsMergedProfile(t *testing.T) {
	p, err := Resolve("academic-research")
	if err != nil {
		t.Fatalf("Resolve(academic-research) returned error: %v", err)
	}
	if p.ID != "academic-research" {
		t.Errorf("ID = %q, want academic-research", p.ID)
	}
	if p.Version != "1.0" {
		t.Errorf("Version = %q, want 1.0", p.Version)
	}
	if p.Extends != "core" {
		t.Errorf("Extends = %q, want core", p.Extends)
	}
	// Merge always materializes non-nil maps for the resolved view.
	if p.Types == nil || p.Severities == nil {
		t.Errorf("merged profile should have non-nil Types/Severities maps, got Types=%v Severities=%v", p.Types, p.Severities)
	}
}

// Ids outside the shipped set — the empty id and an arbitrary unknown — resolve
// to ErrUnknownProfile. Defaulting to core is the CLI's job, not Resolve's, so
// "" is unknown here. academic-research is NO LONGER unknown (I2 ships it).
func TestResolveUnknownProfileIsError(t *testing.T) {
	for _, id := range []string{"", "bogus", "Core", "academic"} {
		if _, err := Resolve(id); !errors.Is(err, ErrUnknownProfile) {
			t.Errorf("Resolve(%q) error = %v, want ErrUnknownProfile", id, err)
		}
	}
}

// ShippedIDs is the single source of truth for the resolvable set and returns it
// sorted; each id must actually resolve.
func TestShippedIDsAllResolve(t *testing.T) {
	ids := ShippedIDs()
	if len(ids) != 2 || ids[0] != "academic-research" || ids[1] != "core" {
		t.Fatalf("ShippedIDs() = %v, want [academic-research core]", ids)
	}
	for _, id := range ids {
		if _, err := Resolve(id); err != nil {
			t.Errorf("shipped id %q does not resolve: %v", id, err)
		}
	}
}

// Load parses a well-formed profile file's bytes into a Profile, reading the full
// closed vocabulary (ADR-010 sub-decision 1) through the injected adapter.
func TestLoadParsesFullVocabulary(t *testing.T) {
	src := []byte(`
profile:
  id: demo
  version: "2.0"
  extends: core
types:
  claim:
    required: [confidence, assessment]
    recommended: [tags]
    enums:
      confidence: [low, medium, high]
    listMin:
      authors: 1
    recommendedAnyOf:
      - [doi, canonical_url]
    requiredSections: [Evidence, Assessment]
    evidenceSections: [Evidence]
    citation:
      requireWhen: { assessment: supported }
      forbiddenTargetTypes: [question]
severities:
  core-citation-unresolved: error
`)
	l := NewLoader(yamladapter.New())
	p, err := l.Load(src)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if p.ID != "demo" || p.Version != "2.0" || p.Extends != "core" {
		t.Fatalf("header mismatch: %+v", p)
	}
	claim, ok := p.Types["claim"]
	if !ok {
		t.Fatal("claim type not parsed")
	}
	if len(claim.Required) != 2 || claim.Required[0] != "confidence" {
		t.Errorf("Required = %v", claim.Required)
	}
	if claim.ListMin["authors"] != 1 {
		t.Errorf("ListMin = %v", claim.ListMin)
	}
	if len(claim.Enums["confidence"]) != 3 {
		t.Errorf("Enums = %v", claim.Enums)
	}
	if len(claim.RecommendedAnyOf) != 1 || len(claim.RecommendedAnyOf[0]) != 2 {
		t.Errorf("RecommendedAnyOf = %v", claim.RecommendedAnyOf)
	}
	if claim.Citation == nil || claim.Citation.RequireWhen["assessment"] != "supported" {
		t.Errorf("Citation = %+v", claim.Citation)
	}
	if len(claim.Citation.ForbiddenTargetTypes) != 1 || claim.Citation.ForbiddenTargetTypes[0] != "question" {
		t.Errorf("ForbiddenTargetTypes = %v", claim.Citation.ForbiddenTargetTypes)
	}
	if p.Severities["core-citation-unresolved"] != "error" {
		t.Errorf("Severities = %v", p.Severities)
	}
}

// Malformed inputs — unparseable YAML, a missing id, and an out-of-set severity
// override — all return ErrMalformedProfile from Load.
func TestLoadMalformedIsError(t *testing.T) {
	l := NewLoader(yamladapter.New())
	cases := map[string][]byte{
		"unparseable-yaml": []byte("profile: [this: is not a mapping"),
		"missing-id":       []byte("profile:\n  version: \"1.0\"\n"),
		"bad-severity":     []byte("profile:\n  id: x\nseverities:\n  some-code: fatal\n"),
	}
	for name, src := range cases {
		if _, err := l.Load(src); !errors.Is(err, ErrMalformedProfile) {
			t.Errorf("Load(%s) error = %v, want ErrMalformedProfile", name, err)
		}
	}
}

// Merge is additive/tightening: the child's rules union over the parent's, the
// child keeps its identity, list fields union order-stably, maps take the child
// on conflict, and the child's citation block replaces the parent's for a shared
// type. Merge must not mutate its inputs.
func TestMergeIsAdditiveAndNonMutating(t *testing.T) {
	base := Profile{
		ID: "core", Version: "0.1.0",
		Types: map[string]TypeRules{
			"claim":  {Required: []string{"confidence"}, ListMin: map[string]int{"authors": 1}},
			"shared": {Recommended: []string{"tags"}},
		},
		Severities: map[string]string{"core-broken-link": "warning"},
	}
	ext := Profile{
		ID: "child", Version: "1.0", Extends: "core",
		Types: map[string]TypeRules{
			"claim":  {Required: []string{"confidence", "assessment"}, ListMin: map[string]int{"authors": 2}},
			"source": {Required: []string{"authors"}},
		},
		Severities: map[string]string{"core-citation-unresolved": "error"},
	}
	got := Merge(base, ext)

	if got.ID != "child" || got.Version != "1.0" || got.Extends != "core" {
		t.Fatalf("identity should be the child's: %+v", got)
	}
	// claim: fields union (order-stable, no duplicate confidence); child tightens listMin.
	claim := got.Types["claim"]
	if len(claim.Required) != 2 || claim.Required[0] != "confidence" || claim.Required[1] != "assessment" {
		t.Errorf("claim.Required = %v, want [confidence assessment]", claim.Required)
	}
	if claim.ListMin["authors"] != 2 {
		t.Errorf("claim.ListMin[authors] = %d, want 2 (child tightens)", claim.ListMin["authors"])
	}
	// parent-only and child-only types both survive.
	if _, ok := got.Types["shared"]; !ok {
		t.Error("parent-only type `shared` was dropped")
	}
	if _, ok := got.Types["source"]; !ok {
		t.Error("child-only type `source` is missing")
	}
	// severities union.
	if got.Severities["core-broken-link"] != "warning" || got.Severities["core-citation-unresolved"] != "error" {
		t.Errorf("severities not unioned: %v", got.Severities)
	}
	// inputs untouched.
	if len(base.Types["claim"].Required) != 1 || base.Types["claim"].ListMin["authors"] != 1 {
		t.Error("Merge mutated its base input")
	}
	if _, ok := base.Severities["core-citation-unresolved"]; ok {
		t.Error("Merge mutated base.Severities")
	}
}

// A resolved academic-research merges core (empty types) with its own (currently
// empty) types, so the end-to-end resolve path exercises the merge, not just
// synthetic profiles.
func TestResolveMergesThroughCoreParent(t *testing.T) {
	p, err := Resolve("academic-research")
	if err != nil {
		t.Fatalf("resolve error: %v", err)
	}
	// Extends is preserved and the parent (core) is a root — resolution did not
	// error on the one-level check.
	if p.Extends != "core" {
		t.Fatalf("Extends = %q", p.Extends)
	}
}

type stubAdapter struct{}

func (stubAdapter) Unmarshal([]byte, any) error { return nil }
func (stubAdapter) Marshal(any) ([]byte, error) { return nil, nil }

// ADR-001 requires all YAML access to route through the internal adapter
// interface, never goccy/go-yaml at call sites. The loader must therefore hold
// exactly the adapter it was given.
func TestNewLoaderRoutesYAMLThroughInjectedAdapter(t *testing.T) {
	a := stubAdapter{}

	l := NewLoader(a)
	if l.yaml != a {
		t.Error("NewLoader must store the injected YAML adapter so all YAML access is routed through it")
	}
}
