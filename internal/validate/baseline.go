package validate

import "github.com/olivermorgan2/llm-wiki-kit/internal/contract"

// Baseline is the set of finding fingerprints an adoption run treats as
// pre-existing. A fingerprint present here marks the finding as already-known so
// the differential filter can suppress it (subject to the hard boundary below).
type Baseline map[string]bool

// fingerprintSep separates fingerprint components. NUL cannot appear in a rule
// code or a filesystem path, so the joined string is collision-free.
const fingerprintSep = "\x00"

// Fingerprint is the stable identity of a finding for baseline comparison:
// {ruleset, code, path}. Severity and message are deliberately excluded so that
// a profile promoting a warning to an error, or a reworded message, does not
// change whether a finding matches the baseline. No line numbers in this issue.
func Fingerprint(f contract.Finding) string {
	return string(f.Ruleset) + fingerprintSep + f.Code + fingerprintSep + f.Path
}

// ApplyBaseline is the final layer of the ADR-004 precedence: a differential
// filter that drops findings recorded in base, bounded by a hard boundary that
// protects structural errors. It never changes a severity — it only drops
// findings. Ordering of the kept findings is preserved.
//
// Boundary (ADR-004):
//   - okf-yaml-parse errors are never suppressed — parsing must succeed before
//     any finding or baseline comparison exists (criterion 7, unconditional).
//   - Other error-severity findings (missing required field, wrong field type)
//     are suppressed only in adoption mode (releaseGate == false). Release-gate /
//     CI runs always evaluate errors at full severity, so a baselined error is
//     kept.
//   - Warnings and suggestions are a pure differential filter: suppressed when
//     baselined, regardless of mode.
func ApplyBaseline(findings []contract.Finding, base Baseline, releaseGate bool) []contract.Finding {
	out := make([]contract.Finding, 0, len(findings))
	for _, f := range findings {
		if keepUnderBaseline(f, base, releaseGate) {
			out = append(out, f)
		}
	}
	return out
}

// keepUnderBaseline reports whether a finding survives the baseline filter.
func keepUnderBaseline(f contract.Finding, base Baseline, releaseGate bool) bool {
	// Parse errors are the never-suppressible boundary case.
	if f.Code == codeOKFYAMLParse {
		return true
	}
	// A finding is a suppression candidate only if the baseline records it.
	if !base[Fingerprint(f)] {
		return true
	}
	// Errors are suppressible only in adoption mode; the release gate keeps them.
	if f.Severity == contract.SeverityError {
		return releaseGate
	}
	// Warnings and suggestions are differentially filtered in either mode.
	return false
}
