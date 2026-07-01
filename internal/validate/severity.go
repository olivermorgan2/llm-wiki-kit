package validate

import "github.com/olivermorgan2/llm-wiki-kit/internal/contract"

// Resolve applies the profile-override layer of the ADR-004 severity precedence
// (core defaults → profile overrides → baseline suppression). overrides maps a
// rule code to the severity a profile assigns it; a finding whose code is present
// is re-stamped with that severity, everything else is left at its core default.
// Overrides change severity only — they never add or remove findings. A nil (or
// empty) override map is the identity, which is what a core-only run passes.
//
// Resolve returns a new slice and does not mutate its input.
func Resolve(findings []contract.Finding, overrides map[string]contract.Severity) []contract.Finding {
	out := make([]contract.Finding, len(findings))
	copy(out, findings)
	if len(overrides) == 0 {
		return out
	}
	for i := range out {
		if sev, ok := overrides[out[i].Code]; ok {
			out[i].Severity = sev
		}
	}
	return out
}
