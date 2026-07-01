package validate

import (
	"strings"
	"testing"

	"github.com/olivermorgan2/llm-wiki-kit/internal/contract"
)

// existsIn builds an existence predicate from a fixed set of bundle paths.
func existsIn(paths ...string) func(string) bool {
	set := make(map[string]bool, len(paths))
	for _, p := range paths {
		set[p] = true
	}
	return func(t string) bool { return set[t] }
}

// (1) A link whose target is absent from the bundle is one core-broken-link finding.
func TestLinkRulesBrokenTargetIsFinding(t *testing.T) {
	got := linkRules("concepts/alpha.md", []byte("See [beta](claims/missing.md)."), existsIn("concepts/alpha.md"))
	f, ok := firstWithCode(got, codeCoreBrokenLink)
	if !ok {
		t.Fatalf("missing target should raise core-broken-link, got %+v", got)
	}
	if !strings.Contains(f.Message, "claims/missing.md") {
		t.Errorf("message should name the broken target, got %q", f.Message)
	}
}

// (2) A link whose target exists in the bundle raises nothing. (AC #3)
func TestLinkRulesValidTargetNoFinding(t *testing.T) {
	got := linkRules("concepts/alpha.md", []byte("See [real](claims/real.md)."),
		existsIn("concepts/alpha.md", "claims/real.md"))
	if len(got) != 0 {
		t.Errorf("valid link should yield no finding, got %+v", got)
	}
}

// (3) External schemes, protocol-relative, mailto, and pure fragments are skipped.
func TestLinkRulesExternalAndFragmentSkipped(t *testing.T) {
	body := []byte("[a](https://example.com/x) [b](http://x) [c](mailto:me@x.com) " +
		"[d](//cdn.example.com/x) [e](#heading) [f](ftp://host/x)")
	got := linkRules("page.md", body, existsIn("page.md"))
	if len(got) != 0 {
		t.Errorf("external/protocol-relative/mailto/fragment links should be skipped, got %+v", got)
	}
}

// (4) The finding carries the confirmed tag/severity/code. (AC #1)
func TestLinkRulesTagSeverityCode(t *testing.T) {
	got := linkRules("page.md", []byte("[x](missing.md)"), existsIn("page.md"))
	f, ok := firstWithCode(got, codeCoreBrokenLink)
	if !ok {
		t.Fatalf("expected a core-broken-link finding, got %+v", got)
	}
	if f.Ruleset != contract.RulesetProfile {
		t.Errorf("ruleset = %q, want profile", f.Ruleset)
	}
	if f.Severity != contract.SeverityWarning {
		t.Errorf("severity = %q, want warning (ADR-004 FR8 default)", f.Severity)
	}
	if f.Code != codeCoreBrokenLink {
		t.Errorf("code = %q, want %q", f.Code, codeCoreBrokenLink)
	}
	if f.Path != "page.md" {
		t.Errorf("path = %q, want page.md", f.Path)
	}
}

// (5) Two broken targets on one page aggregate into exactly one finding naming both.
func TestLinkRulesAggregatesPerPage(t *testing.T) {
	got := linkRules("page.md", []byte("[a](one.md) and [b](two.md)"), existsIn("page.md"))
	if n := len(findingsWithCode(got, codeCoreBrokenLink)); n != 1 {
		t.Fatalf("want exactly one aggregated finding, got %d (%+v)", n, got)
	}
	f, _ := firstWithCode(got, codeCoreBrokenLink)
	if !strings.Contains(f.Message, "one.md") || !strings.Contains(f.Message, "two.md") {
		t.Errorf("message should name both broken targets, got %q", f.Message)
	}
}

// Image links (![alt](t)) are assets, not pages: they must not be link-checked.
func TestLinkRulesImageLinksSkipped(t *testing.T) {
	got := linkRules("page.md", []byte("![diagram](img/missing.png)"), existsIn("page.md"))
	if len(got) != 0 {
		t.Errorf("image links should be skipped, got %+v", got)
	}
}

// A #fragment or ?query suffix is stripped before resolution; the base path resolves.
func TestLinkRulesFragmentAndQueryStripped(t *testing.T) {
	got := linkRules("page.md", []byte("[a](real.md#section) [b](real.md?v=1)"),
		existsIn("page.md", "real.md"))
	if len(got) != 0 {
		t.Errorf("fragment/query suffix should be stripped before resolution, got %+v", got)
	}
}

// A target escaping the bundle root via ../ is unresolved (broken), and the
// filesystem is never consulted for it (fs-safety enforcement is issue #5).
func TestLinkRulesEscapingRootIsBroken(t *testing.T) {
	got := linkRules("concepts/alpha.md", []byte("[up](../../etc/passwd)"), existsIn("concepts/alpha.md"))
	if _, ok := firstWithCode(got, codeCoreBrokenLink); !ok {
		t.Errorf("root-escaping target should be broken, got %+v", got)
	}
}

// Targets resolve bundle-root-relative regardless of the linking page's location.
func TestLinkRulesBundleRootRelative(t *testing.T) {
	// From a nested page, "claims/real.md" is relative to the bundle root, not the page.
	got := linkRules("concepts/deep/alpha.md", []byte("[r](claims/real.md)"),
		existsIn("concepts/deep/alpha.md", "claims/real.md"))
	if len(got) != 0 {
		t.Errorf("target should resolve against bundle root, got %+v", got)
	}
}
