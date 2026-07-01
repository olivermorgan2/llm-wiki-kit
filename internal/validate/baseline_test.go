package validate

import (
	"testing"

	"github.com/olivermorgan2/llm-wiki-kit/internal/contract"
)

func fp(f contract.Finding) string { return Fingerprint(f) }

func parseErr(path string) contract.Finding {
	return contract.Finding{Ruleset: contract.RulesetOKF, Severity: contract.SeverityError, Code: codeOKFYAMLParse, Path: path}
}
func missingTitle(path string) contract.Finding {
	return contract.Finding{Ruleset: contract.RulesetProfile, Severity: contract.SeverityError, Code: codeCoreReqTitle, Path: path}
}
func kebabWarn(path string) contract.Finding {
	return contract.Finding{Ruleset: contract.RulesetProfile, Severity: contract.SeverityWarning, Code: codeCoreKebab, Path: path}
}
func recSuggestion(path string) contract.Finding {
	return contract.Finding{Ruleset: contract.RulesetProfile, Severity: contract.SeveritySuggestion, Code: codeCoreRecommended, Path: path}
}

func TestFingerprintDependsOnRulesetCodePath(t *testing.T) {
	a := contract.Finding{Ruleset: contract.RulesetOKF, Code: "x", Path: "p.md"}
	b := contract.Finding{Ruleset: contract.RulesetProfile, Code: "x", Path: "p.md"}
	if Fingerprint(a) == Fingerprint(b) {
		t.Error("fingerprint must distinguish ruleset")
	}
	// Severity and message are not part of the fingerprint.
	c := a
	c.Severity = contract.SeverityError
	c.Message = "different"
	if Fingerprint(a) != Fingerprint(c) {
		t.Error("fingerprint must ignore severity and message")
	}
}

// A parse error is never baseline-suppressible, even in adoption mode with the
// fingerprint recorded (criterion 7 holds unconditionally).
func TestApplyBaselineNeverSuppressesParseError(t *testing.T) {
	f := parseErr("bad.md")
	base := Baseline{fp(f): true}
	got := ApplyBaseline([]contract.Finding{f}, base, false)
	if len(got) != 1 {
		t.Fatalf("parse error was suppressed in adoption mode: %+v", got)
	}
}

// Missing-required errors are suppressed only in adoption mode (releaseGate=false).
func TestApplyBaselineSuppressesMissingRequiredInAdoptionMode(t *testing.T) {
	f := missingTitle("a.md")
	base := Baseline{fp(f): true}
	got := ApplyBaseline([]contract.Finding{f}, base, false)
	if len(got) != 0 {
		t.Fatalf("adoption mode should suppress a baselined missing-required error: %+v", got)
	}
}

// Release-gate mode never suppresses errors, even baselined ones.
func TestApplyBaselineReleaseGateKeepsErrors(t *testing.T) {
	f := missingTitle("a.md")
	base := Baseline{fp(f): true}
	got := ApplyBaseline([]contract.Finding{f}, base, true)
	if len(got) != 1 {
		t.Fatalf("release gate must evaluate errors at full severity: %+v", got)
	}
}

// Warnings and suggestions are a pure differential filter: suppressed when
// baselined, regardless of mode.
func TestApplyBaselineSuppressesWarningsAndSuggestions(t *testing.T) {
	w, s := kebabWarn("A.md"), recSuggestion("a.md")
	base := Baseline{fp(w): true, fp(s): true}
	for _, gate := range []bool{true, false} {
		got := ApplyBaseline([]contract.Finding{w, s}, base, gate)
		if len(got) != 0 {
			t.Errorf("baselined warning/suggestion not suppressed (releaseGate=%v): %+v", gate, got)
		}
	}
}

// A finding not present in the baseline is always kept (new/worsened findings
// still surface).
func TestApplyBaselineKeepsUnbaselinedFindings(t *testing.T) {
	f := kebabWarn("A.md")
	got := ApplyBaseline([]contract.Finding{f}, Baseline{}, false)
	if len(got) != 1 {
		t.Fatalf("unbaselined finding should be kept: %+v", got)
	}
}

// A nil baseline suppresses nothing (the default CLI path).
func TestApplyBaselineNilBaselineIsIdentity(t *testing.T) {
	in := []contract.Finding{parseErr("b.md"), missingTitle("a.md"), kebabWarn("A.md")}
	got := ApplyBaseline(in, nil, false)
	if len(got) != len(in) {
		t.Fatalf("nil baseline changed count: %d != %d", len(got), len(in))
	}
}
