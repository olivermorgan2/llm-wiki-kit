// Package profile holds the data-driven validation profiles (ADR-007: one
// inheritance level — a domain profile extends the core profile). A profile is a
// declarative YAML rule file loaded through the injected yamladapter.Adapter
// (ADR-001), never goccy/go-yaml at call sites, and consumed by the ADR-004
// validation engine so OKF and profile findings stay separable (criterion 5).
// The schema is fixed by ADR-010: a closed, per-type vocabulary the engine owns
// rule kinds for, with the profile supplying only their data.
package profile

import (
	"errors"
	"fmt"
	"sort"

	"github.com/olivermorgan2/llm-wiki-kit/internal/yamladapter"
	"github.com/olivermorgan2/llm-wiki-kit/profiles"
)

// CoreID is the id of the shipped core profile — the default profile init
// materializes a reference to (ADR-007).
const CoreID = "core"

// ErrUnknownProfile is returned by Resolve for any id outside the shipped set.
var ErrUnknownProfile = errors.New("profile: unknown profile id")

// ErrMalformedProfile is returned when a profile file is present but its data is
// invalid: unparseable YAML, an out-of-set severity override, or an inheritance
// chain deeper than the single level ADR-007 permits.
var ErrMalformedProfile = errors.New("profile: malformed profile data")

// shippedProfiles maps a resolvable profile id to its embedded rule-file path
// (profiles.FS). It is the MVP shipped set; the Phase-7 local-path resolution
// seam (ADR-007) attaches at Resolve, extending — not replacing — this lookup.
var shippedProfiles = map[string]string{
	CoreID:              "core/profile.yaml",
	"academic-research": "academic-research/profile.yaml",
}

// validSeverities is the ADR-004 severity set a profile `severities` override may
// name. Kept as strings here so the profile package stays decoupled from
// internal/contract; the validation engine maps these to contract.Severity.
var validSeverities = map[string]bool{"error": true, "warning": true, "suggestion": true}

// Profile is a resolved, data-driven validation profile (ADR-007/ADR-010). After
// Resolve it is the fully merged view a domain profile presents to the engine:
// its own rules unioned over its `core` parent's. Version is the pinned profile
// version the bundle config records (ADR-007).
type Profile struct {
	ID      string
	Version string
	// Extends is the parent profile id ("" for core). Exactly one level is
	// permitted (ADR-007): a profile named here must not itself extend another.
	Extends string
	// Types holds the per-page-type rules keyed on frontmatter `type` (ADR-010
	// sub-decision 1). A page whose type is absent here receives only the
	// engine-owned OKF/core rules (unknown types stay accepted, sub-decision 5).
	Types map[string]TypeRules
	// Severities is the optional per-rule severity-override map (ADR-010; ADR-008
	// sub-decision 5(d)), code → one of error|warning|suggestion. It feeds the
	// profile-override layer of ADR-004's precedence via validate.Resolve; it only
	// restamps severity, never adds or removes findings.
	Severities map[string]string
}

// TypeRules is the closed per-type rule vocabulary (ADR-010 sub-decision 1).
// Every field is optional; the engine owns the rule kind and reads only the data
// present. Per-type Required/Recommended list only the fields the profile ADDS
// beyond what a page inherits from core.
type TypeRules struct {
	Required         []string            `yaml:"required"`
	Recommended      []string            `yaml:"recommended"`
	Enums            map[string][]string `yaml:"enums"`
	ListMin          map[string]int      `yaml:"listMin"`
	RecommendedAnyOf [][]string          `yaml:"recommendedAnyOf"`
	RequiredSections []string            `yaml:"requiredSections"`
	EvidenceSections []string            `yaml:"evidenceSections"`
	Citation         *CitationRules      `yaml:"citation"`
}

// CitationRules is a type's citation obligation vocabulary (ADR-008/ADR-010).
type CitationRules struct {
	// RequireWhen is the single, deliberately-weakest field-equals-value trigger
	// (ADR-010): when every listed field holds its listed value, the type's
	// evidence section must contain at least one citation.
	RequireWhen map[string]string `yaml:"requireWhen"`
	// ForbiddenTargetTypes is a denylist of page `type`s a citation may never
	// resolve to (e.g. a `question` page is never evidence).
	ForbiddenTargetTypes []string `yaml:"forbiddenTargetTypes"`
}

// profileFile mirrors the on-disk YAML shape: a `profile:` header plus top-level
// `types` and `severities` maps (ADR-010 sub-decision 1). It is unmarshaled
// through the adapter and flattened into Profile.
type profileFile struct {
	Profile struct {
		ID      string `yaml:"id"`
		Version string `yaml:"version"`
		Extends string `yaml:"extends"`
	} `yaml:"profile"`
	Types      map[string]TypeRules `yaml:"types"`
	Severities map[string]string    `yaml:"severities"`
}

// Loader reads and resolves profiles. All YAML access is routed through the
// injected yamladapter.Adapter (ADR-001), never goccy/go-yaml at call sites.
type Loader struct {
	yaml yamladapter.Adapter
}

// NewLoader returns a Loader that routes every YAML operation through the given
// adapter.
func NewLoader(a yamladapter.Adapter) *Loader {
	return &Loader{yaml: a}
}

// Load parses one profile file's bytes into a Profile and validates its
// self-contained invariants (ID present, severity overrides in-set). It does not
// resolve `extends`; use Resolve for the merged view. A parse failure or an
// invalid severity value returns ErrMalformedProfile.
func (l *Loader) Load(data []byte) (Profile, error) {
	var pf profileFile
	if err := l.yaml.Unmarshal(data, &pf); err != nil {
		return Profile{}, fmt.Errorf("%w: %v", ErrMalformedProfile, err)
	}
	if pf.Profile.ID == "" {
		return Profile{}, fmt.Errorf("%w: missing profile.id", ErrMalformedProfile)
	}
	for code, sev := range pf.Severities {
		if !validSeverities[sev] {
			return Profile{}, fmt.Errorf("%w: severity override %q=%q is not one of error|warning|suggestion",
				ErrMalformedProfile, code, sev)
		}
	}
	return Profile{
		ID:         pf.Profile.ID,
		Version:    pf.Profile.Version,
		Extends:    pf.Profile.Extends,
		Types:      pf.Types,
		Severities: pf.Severities,
	}, nil
}

// resolve loads the profile registered under id and, if it declares an `extends`
// parent, merges it over that parent (exactly one level). It is the engine of the
// package-level Resolve; the Loader carries the adapter so all YAML stays behind
// the ADR-001 seam.
func (l *Loader) resolve(id string) (Profile, error) {
	path, ok := shippedProfiles[id]
	if !ok {
		return Profile{}, fmt.Errorf("%w: %q", ErrUnknownProfile, id)
	}
	data, err := profiles.FS.ReadFile(path)
	if err != nil {
		// A shipped id whose file is missing is a build-integrity failure, not an
		// unknown id — surface it as malformed rather than masking it.
		return Profile{}, fmt.Errorf("%w: reading %q: %v", ErrMalformedProfile, path, err)
	}
	child, err := l.Load(data)
	if err != nil {
		return Profile{}, err
	}
	if child.Extends == "" {
		return child, nil
	}

	parentPath, ok := shippedProfiles[child.Extends]
	if !ok {
		return Profile{}, fmt.Errorf("%w: %q extends unknown profile %q",
			ErrMalformedProfile, id, child.Extends)
	}
	parentData, err := profiles.FS.ReadFile(parentPath)
	if err != nil {
		return Profile{}, fmt.Errorf("%w: reading parent %q: %v", ErrMalformedProfile, parentPath, err)
	}
	parent, err := l.Load(parentData)
	if err != nil {
		return Profile{}, err
	}
	if parent.Extends != "" {
		// ADR-007 fixes exactly one inheritance level: the parent must be a root.
		return Profile{}, fmt.Errorf("%w: %q extends %q, which itself extends %q (only one level permitted)",
			ErrMalformedProfile, id, child.Extends, parent.Extends)
	}
	return Merge(parent, child), nil
}

// Merge returns the one-level resolution of ext over its base parent (ADR-007).
// The child's identity (id, version, extends) is kept; rules are unioned so a
// page is validated by the parent's rules plus the child's. Per-type rules union
// by type key; within a shared type, list fields union (order-stable) and
// key-conflicting map entries and the citation block take the child's value.
// Severity overrides union with the child winning.
//
// The "never relaxes core" guarantee (ADR-007) holds **structurally**, not by an
// enforcement pass here: `core` — the only permitted parent (ADR-010
// sub-decision 2) — contributes **no** per-type rules or severity overrides, so
// there is nothing for a child to relax; the child-wins policy only ever ADDS.
// If a future `core` ever gained behavior-bearing per-type data, this policy
// would need a tightening check (max listMin, intersect enums, no severity
// demotion) — recorded here so the assumption is explicit, not silent.
//
// The returned Profile is a **deep copy**: it shares no slice, map, or Citation
// pointer with either input, so a caller may mutate the resolved profile without
// disturbing the sources (or the embedded parse).
func Merge(base, ext Profile) Profile {
	out := Profile{
		ID:         ext.ID,
		Version:    ext.Version,
		Extends:    ext.Extends,
		Types:      map[string]TypeRules{},
		Severities: map[string]string{},
	}
	for t, r := range base.Types {
		out.Types[t] = cloneTypeRules(r)
	}
	for t, r := range ext.Types {
		if existing, ok := out.Types[t]; ok {
			out.Types[t] = mergeTypeRules(existing, r)
		} else {
			out.Types[t] = cloneTypeRules(r)
		}
	}
	for code, sev := range base.Severities {
		out.Severities[code] = sev
	}
	for code, sev := range ext.Severities {
		out.Severities[code] = sev
	}
	return out
}

// mergeTypeRules unions two TypeRules for the same page type, child (ext) over
// parent (base), returning a value that aliases neither input. List fields union
// preserving parent-then-child first-seen order; key-conflicting map entries take
// the child; the child's citation block, if present, replaces the parent's.
func mergeTypeRules(base, ext TypeRules) TypeRules {
	out := TypeRules{
		Required:         unionStrings(base.Required, ext.Required),
		Recommended:      unionStrings(base.Recommended, ext.Recommended),
		Enums:            mergeStringSliceMap(base.Enums, ext.Enums),
		ListMin:          mergeIntMap(base.ListMin, ext.ListMin),
		RecommendedAnyOf: concatGroups(base.RecommendedAnyOf, ext.RecommendedAnyOf),
		RequiredSections: unionStrings(base.RequiredSections, ext.RequiredSections),
		EvidenceSections: unionStrings(base.EvidenceSections, ext.EvidenceSections),
		Citation:         cloneCitation(base.Citation),
	}
	if ext.Citation != nil {
		out.Citation = cloneCitation(ext.Citation)
	}
	return out
}

// cloneTypeRules returns a deep copy of r that shares no backing storage with it,
// so a parent-only or child-only type placed into a resolved profile cannot alias
// the source profile (or the embedded parse).
func cloneTypeRules(r TypeRules) TypeRules {
	return TypeRules{
		Required:         cloneStrings(r.Required),
		Recommended:      cloneStrings(r.Recommended),
		Enums:            mergeStringSliceMap(r.Enums, nil),
		ListMin:          mergeIntMap(r.ListMin, nil),
		RecommendedAnyOf: concatGroups(r.RecommendedAnyOf, nil),
		RequiredSections: cloneStrings(r.RequiredSections),
		EvidenceSections: cloneStrings(r.EvidenceSections),
		Citation:         cloneCitation(r.Citation),
	}
}

// cloneCitation returns a deep copy of c (nil for a nil input).
func cloneCitation(c *CitationRules) *CitationRules {
	if c == nil {
		return nil
	}
	out := &CitationRules{ForbiddenTargetTypes: cloneStrings(c.ForbiddenTargetTypes)}
	if c.RequireWhen != nil {
		out.RequireWhen = make(map[string]string, len(c.RequireWhen))
		for k, v := range c.RequireWhen {
			out.RequireWhen[k] = v
		}
	}
	return out
}

// cloneStrings copies a string slice, preserving nil-for-empty.
func cloneStrings(s []string) []string {
	if len(s) == 0 {
		return nil
	}
	return append([]string(nil), s...)
}

// unionStrings returns the order-stable union of a and b (a's order first, then
// b's new members) in a freshly-allocated slice, never mutating or aliasing
// either input. A nil result is returned only when both inputs are empty.
func unionStrings(a, b []string) []string {
	seen := map[string]bool{}
	var out []string
	add := func(src []string) {
		for _, s := range src {
			if !seen[s] {
				seen[s] = true
				out = append(out, s)
			}
		}
	}
	add(a)
	add(b)
	return out
}

// concatGroups appends b's groups after a's, deep-copying each inner slice so the
// result aliases neither input. Returns nil only when both inputs are empty.
func concatGroups(a, b [][]string) [][]string {
	if len(a) == 0 && len(b) == 0 {
		return nil
	}
	out := make([][]string, 0, len(a)+len(b))
	for _, g := range a {
		out = append(out, cloneStrings(g))
	}
	for _, g := range b {
		out = append(out, cloneStrings(g))
	}
	return out
}

// mergeStringSliceMap unions two map[string][]string with the child (b) winning
// on key conflict, deep-copying each value slice. Returns nil only when both
// inputs are empty.
func mergeStringSliceMap(a, b map[string][]string) map[string][]string {
	if len(a) == 0 && len(b) == 0 {
		return nil
	}
	out := map[string][]string{}
	for k, v := range a {
		out[k] = cloneStrings(v)
	}
	for k, v := range b {
		out[k] = cloneStrings(v)
	}
	return out
}

// mergeIntMap unions two map[string]int with the child (b) winning on key
// conflict. Returns nil only when both inputs are empty.
func mergeIntMap(a, b map[string]int) map[string]int {
	if len(a) == 0 && len(b) == 0 {
		return nil
	}
	out := map[string]int{}
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = v
	}
	return out
}

// Resolve returns the fully merged shipped profile registered under id, or
// ErrUnknownProfile for any unregistered id (including the empty id — defaulting
// to core is the caller's responsibility, not Resolve's) and ErrMalformedProfile
// for a shipped file that fails to parse or violates the one-level inheritance
// rule. Resolution is by id against the embedded shipped set (profiles.FS),
// routed through the ADR-001 adapter. The Phase-7 local-path profile resolution
// seam (ADR-007) attaches here: a filesystem-backed lookup would extend, not
// replace, the shipped-set path.
func Resolve(id string) (Profile, error) {
	return NewLoader(yamladapter.New()).resolve(id)
}

// ShippedIDs returns the resolvable shipped profile ids in sorted order. It backs
// the CLI's unknown-profile diagnostics and keeps the shipped set single-sourced.
func ShippedIDs() []string {
	ids := make([]string, 0, len(shippedProfiles))
	for id := range shippedProfiles {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
