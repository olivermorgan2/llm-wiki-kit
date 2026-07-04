package validate

import (
	"sort"
	"strings"

	"github.com/olivermorgan2/llm-wiki-kit/internal/contract"
	"github.com/olivermorgan2/llm-wiki-kit/internal/profile"
)

// profileCitationFindings evaluates a profiled page type's citation obligations
// (ADR-008 sub-decision 5; ADR-010 sub-decision 3/4): the require-citation
// obligation (profile-citation-required) and the forbidden-target-type check
// (profile-citation-target-type). It fires only for a page whose frontmatter
// `type` is a profiled type carrying a `citation` block; every other page returns
// nothing, so core bundles are unaffected.
//
// Citations are the inline links (non-image) inside the type's designated
// evidence sections — the same grammar and evidence-context model the shared
// resolver and core-citation-* rules use (ADR-008 sub-decision 1). The obligation
// is about PRESENCE: a present-but-unresolved citation satisfies it (carry-in №2)
// while its resolvability is reported separately by core-citation-unresolved.
func profileCitationFindings(prof profile.Profile, pagePath string, m map[string]any, body []byte, res *resolver, pageTypes map[string]string) []contract.Finding {
	pageType, _ := m["type"].(string)
	rules, ok := prof.Types[pageType]
	if !ok || rules.Citation == nil {
		return nil
	}
	citation := rules.Citation

	// Collect the inline-link targets that live inside this type's evidence
	// sections (in first-seen order). Images are never citations (sub-decision 1).
	var targets []string
	for _, seg := range splitEvidenceContexts(body, rules.EvidenceSections) {
		if !seg.evidence {
			continue
		}
		for _, mt := range inlineLink.FindAllSubmatch(seg.text, -1) {
			if len(mt[1]) > 0 {
				continue // image link `![alt](t)`
			}
			targets = append(targets, strings.TrimSpace(string(mt[2])))
		}
	}

	var out []contract.Finding

	// Require-citation obligation: when every requireWhen field holds its declared
	// value and the evidence section has NO citation at all, that is an error. A
	// present-but-unresolved citation counts as present and satisfies it.
	if requireWhenHolds(citation.RequireWhen, m) && len(targets) == 0 {
		out = append(out, profileFinding(contract.SeverityError, codeProfileCitationRequired,
			"type \""+pageType+"\" obliges a citation in its evidence section but none is present", pagePath))
	}

	// Forbidden-target-type: a citation that resolves to an in-bundle page whose
	// `type` is denied (e.g. a `question` is never evidence). Only resolvable
	// in-bundle targets have a type to check; external/absent/malformed targets are
	// governed by core-citation-* and cannot be type-checked.
	if len(citation.ForbiddenTargetTypes) > 0 {
		forbidden := map[string]bool{}
		for _, t := range citation.ForbiddenTargetTypes {
			forbidden[t] = true
		}
		var bad []string
		seen := map[string]bool{}
		for _, target := range targets {
			r := res.classify(target)
			if r.class != classInBundle || !r.resolved {
				continue
			}
			cleaned := strings.TrimPrefix(r.key, "bundle:")
			if forbidden[pageTypes[cleaned]] && !seen[target] {
				seen[target] = true
				bad = append(bad, target)
			}
		}
		if len(bad) > 0 {
			out = append(out, profileFinding(contract.SeverityError, codeProfileCitationTargetType,
				"citation(s) point at a forbidden target type ("+strings.Join(sortedForbidden(citation.ForbiddenTargetTypes), ", ")+"): "+strings.Join(bad, ", "), pagePath))
		}
	}

	return out
}

// requireWhenHolds reports whether every (field, value) pair in the requireWhen
// trigger matches the page's frontmatter (string comparison). An empty or nil
// trigger never holds — a type with no requireWhen carries no obligation.
func requireWhenHolds(requireWhen map[string]string, m map[string]any) bool {
	if len(requireWhen) == 0 {
		return false
	}
	for field, want := range requireWhen {
		got, isStr := m[field].(string)
		if !isStr || got != want {
			return false
		}
	}
	return true
}

// sortedForbidden returns a sorted copy of the forbidden-type list for a stable
// message.
func sortedForbidden(types []string) []string {
	out := append([]string(nil), types...)
	sort.Strings(out)
	return out
}
