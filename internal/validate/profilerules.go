package validate

import (
	"fmt"
	"sort"
	"strings"

	"github.com/olivermorgan2/llm-wiki-kit/internal/contract"
	"github.com/olivermorgan2/llm-wiki-kit/internal/profile"
)

// profileTypeFindings evaluates the profile-data-driven, type-conditional
// structural rules for one page (ADR-010 sub-decision 3). It fires only when the
// page's frontmatter `type` is a profiled type in prof; any other type receives
// only the OKF/core engine rules (unknown types stay accepted, sub-decision 5),
// so the zero Profile (no Types) yields nothing and core-profile bundles validate
// exactly as before Phase 4.
//
// Each rule kind aggregates one problem into one finding per {ruleset, code, path}
// (ADR-004 FR8) and emits at its addendum-003 default severity; the caller applies
// profile severity overrides via Resolve. Rule kinds run in the ADR-010 table
// order (required field, enum, list-min, required section, recommended pair) and
// map iteration is sorted, so messages are deterministic.
func profileTypeFindings(prof profile.Profile, pagePath string, m map[string]any, body []byte) []contract.Finding {
	pageType, _ := m["type"].(string)
	rules, ok := prof.Types[pageType]
	if !ok {
		return nil
	}

	var out []contract.Finding

	// Required fields the profile ADDS beyond core. Fires only on an absent field
	// (a present-but-empty list is the list-min rule's concern, sub-decision 3).
	if missing := absentFields(m, rules.Required); len(missing) > 0 {
		out = append(out, profileFinding(contract.SeverityError, codeProfileRequiredField,
			fmt.Sprintf("type %q requires field(s): %s", pageType, strings.Join(missing, ", ")), pagePath))
	}

	// Enum-constrained fields whose present value is outside the allowed set.
	if bad := enumViolations(m, rules.Enums); len(bad) > 0 {
		out = append(out, profileFinding(contract.SeverityError, codeProfileFieldEnum,
			"field value(s) outside the allowed set: "+strings.Join(bad, ", "), pagePath))
	}

	// List-min fields that are present but not a list of at least N items.
	if short := listMinViolations(m, rules.ListMin); len(short) > 0 {
		out = append(out, profileFinding(contract.SeverityError, codeProfileListMin,
			"list field(s) below the required minimum length: "+strings.Join(short, ", "), pagePath))
	}

	// Required Markdown sections absent from the body (reuse parseATX).
	if missingSec := absentSections(body, rules.RequiredSections); len(missingSec) > 0 {
		out = append(out, profileFinding(contract.SeverityError, codeProfileRequiredSection,
			"missing required section(s): "+strings.Join(missingSec, ", "), pagePath))
	}

	// Recommended any-of groups with no member present (advisory suggestion).
	if groups := unsatisfiedAnyOf(m, rules.RecommendedAnyOf); len(groups) > 0 {
		out = append(out, profileFinding(contract.SeveritySuggestion, codeProfileRecommendedPair,
			"provide at least one of each recommended group: "+strings.Join(groups, ", "), pagePath))
	}

	return out
}

// absentFields returns, in declared order, the required fields that are absent
// from m. Presence is "key exists with a non-null, non-empty(-string) value"
// (ADR-010 sub-decision 3): a present-but-empty list counts as present here so it
// falls to the list-min rule, never trips required-field, and no field yields two
// findings.
func absentFields(m map[string]any, required []string) []string {
	var missing []string
	for _, f := range required {
		if !fieldPresent(m, f) {
			missing = append(missing, f)
		}
	}
	return missing
}

// enumViolations returns "field=value" entries (fields sorted) for present fields
// whose value lies outside the profile's allowed set. Presence is tested with the
// same fieldPresent predicate as the required-field rule (ADR-010 sub-decision 3):
// an absent field — including an empty/whitespace string, which is "absent" — is
// skipped so it can only be the required-field rule's concern, never both. A
// non-scalar present value cannot match a scalar set and is reported as a
// violation.
//
// enums and listMin are authored on disjoint field shapes (a scalar-enum field vs
// a list-min field), so no well-formed profile — including addendum 003's —
// constrains one field with both; the "never trips both" invariant holds by
// construction, and the engine does not police a self-contradictory profile that
// declares a single field as both a scalar enum and a list.
func enumViolations(m map[string]any, enums map[string][]string) []string {
	var bad []string
	for _, field := range sortedKeys(enums) {
		if !fieldPresent(m, field) {
			continue // absent (incl. ""/whitespace): required-field's concern, not enum's.
		}
		s, isStr := m[field].(string)
		if !isStr || !containsString(enums[field], s) {
			bad = append(bad, fmt.Sprintf("%s=%v", field, m[field]))
		}
	}
	return bad
}

// listMinViolations returns "field (min N)" entries (fields sorted) for present
// fields that are not a YAML sequence of at least N items. Presence uses the same
// fieldPresent predicate as required-field (ADR-010 sub-decision 3): an absent
// field — including an empty/whitespace string — is skipped, so `authors:`
// omitted or `authors: ""` is required-field only, while `authors: []` or
// `authors: "x"` (present, not a ≥N list) is list-min only. A field never trips
// both required-field and list-min.
func listMinViolations(m map[string]any, listMin map[string]int) []string {
	var short []string
	for _, field := range sortedIntMapKeys(listMin) {
		if !fieldPresent(m, field) {
			continue
		}
		seq, isSeq := m[field].([]any)
		if !isSeq || len(seq) < listMin[field] {
			short = append(short, fmt.Sprintf("%s (min %d)", field, listMin[field]))
		}
	}
	return short
}

// absentSections returns, in declared order, the required section titles whose
// ATX heading is absent from body. Section titles are compared exactly and
// case-sensitively (parseATX / the same rule as splitEvidenceContexts); the
// message renders them with a leading "## " for readability, but matching is on
// the title text at any heading level.
func absentSections(body []byte, required []string) []string {
	if len(required) == 0 {
		return nil
	}
	present := headingTitles(body)
	var missing []string
	for _, title := range required {
		if !present[title] {
			missing = append(missing, "## "+title)
		}
	}
	return missing
}

// unsatisfiedAnyOf returns, in declared order, a rendered "(a | b)" entry for each
// recommended any-of group with no present member (ADR-010 sub-decision 3). An
// empty group is treated as satisfied (there is nothing to require).
func unsatisfiedAnyOf(m map[string]any, groups [][]string) []string {
	var out []string
	for _, group := range groups {
		if len(group) == 0 {
			continue
		}
		satisfied := false
		for _, field := range group {
			if fieldPresent(m, field) {
				satisfied = true
				break
			}
		}
		if !satisfied {
			out = append(out, "("+strings.Join(group, " | ")+")")
		}
	}
	return out
}

// fieldPresent reports whether key holds a present value: the key exists, is not
// YAML null, and — for a string value — is not empty/whitespace. A present list
// (including an empty one), number, bool, or mapping counts as present; list
// emptiness is the list-min rule's concern, not presence.
func fieldPresent(m map[string]any, key string) bool {
	v, ok := m[key]
	if !ok || v == nil {
		return false
	}
	if s, isStr := v.(string); isStr {
		return strings.TrimSpace(s) != ""
	}
	return true
}

// headingTitles collects the set of ATX heading titles present in body, at any
// level, using the same parser as the evidence-context splitter.
func headingTitles(body []byte) map[string]bool {
	titles := map[string]bool{}
	for _, line := range strings.Split(string(body), "\n") {
		if _, title, ok := parseATX([]byte(line)); ok {
			titles[title] = true
		}
	}
	return titles
}

// containsString reports whether set contains s.
func containsString(set []string, s string) bool {
	for _, x := range set {
		if x == s {
			return true
		}
	}
	return false
}

// sortedKeys returns the sorted keys of a map[string][]string for deterministic
// iteration.
func sortedKeys(m map[string][]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// sortedIntMapKeys returns the sorted keys of a map[string]int.
func sortedIntMapKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
