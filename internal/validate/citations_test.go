package validate

import (
	"strings"
	"testing"

	"github.com/olivermorgan2/llm-wiki-kit/internal/contract"
)

// ----------------------------------------------------------------------------
// Phase A — classifier unit tests (ADR-008 sub-decision 2 resolution classes).
// ----------------------------------------------------------------------------

// (A1) A syntactically valid absolute http(s) URL is resolved, scheme match is
// case-insensitive.
func TestClassifyValidHTTPIsResolved(t *testing.T) {
	r := navResolver(existsIn())
	for _, target := range []string{"https://example.com/x", "HTTP://Example.com", "hTTps://a.b/c?q#f"} {
		got := r.classify(target)
		if got.class != classExternal || !got.resolved {
			t.Errorf("classify(%q) = %+v, want external+resolved", target, got)
		}
	}
}

// (A2) An http(s) URL with no host can never resolve: malformed, not unresolved.
func TestClassifyInvalidHostHTTPIsMalformed(t *testing.T) {
	r := navResolver(existsIn())
	for _, target := range []string{"https://", "http:opaque", "http://"} {
		got := r.classify(target)
		if got.class != classMalformed || got.resolved {
			t.Errorf("classify(%q) = %+v, want malformed", target, got)
		}
	}
}

// (A3) An in-bundle target present in the bundle is resolved; absent is
// unresolved (well-formed but not yet a member).
func TestClassifyInBundlePresentAndAbsent(t *testing.T) {
	r := navResolver(existsIn("claims/real.md"))
	if got := r.classify("claims/real.md"); got.class != classInBundle || !got.resolved {
		t.Errorf("present target = %+v, want in-bundle+resolved", got)
	}
	got := r.classify("claims/missing.md")
	if got.class != classInBundle || got.resolved {
		t.Errorf("absent target = %+v, want in-bundle+unresolved", got)
	}
}

// (A4) The malformed catch-all: empty, non-http(s) schemes, fragment-only,
// protocol-relative, and query-only targets.
func TestClassifyMalformedCatchAll(t *testing.T) {
	r := navResolver(existsIn())
	for _, target := range []string{"", "mailto:me@x.com", "ftp://h/x", "#frag", "//cdn/x", "?v=1"} {
		got := r.classify(target)
		if got.class != classMalformed || got.resolved {
			t.Errorf("classify(%q) = %+v, want malformed", target, got)
		}
	}
}

// (A5) With no repo anchor (nil repoResolve), a bundle-escaping ../ is malformed
// (bundle-root fallback: the repo-path class is empty).
func TestClassifyNoAnchorDotDotIsMalformed(t *testing.T) {
	r := navResolver(existsIn())
	got := r.classify("../shared/doc.md")
	if got.class != classMalformed || got.resolved {
		t.Errorf("no-anchor ../ = %+v, want malformed", got)
	}
}

// (A6) With an anchor, a ../ staying inside the repo root resolves via the
// repo-path class: RepoFound → resolved, RepoAbsent → unresolved.
func TestClassifyInRepoDotDotResolves(t *testing.T) {
	r := &resolver{
		exists:      existsIn(),
		bundleDir:   "wiki",
		repoResolve: repoStub(map[string]RepoStatus{"shared/doc.md": RepoFound}),
	}
	if got := r.classify("../shared/doc.md"); got.class != classRepo || !got.resolved {
		t.Errorf("in-repo ../ found = %+v, want repo+resolved", got)
	}
	if got := r.classify("../shared/gone.md"); got.class != classRepo || got.resolved {
		t.Errorf("in-repo ../ absent = %+v, want repo+unresolved", got)
	}
}

// (A7) A lexical escape past the repo root, and an injected RepoEscape verdict,
// are both malformed.
func TestClassifyRepoEscapeIsMalformed(t *testing.T) {
	lexical := &resolver{exists: existsIn(), bundleDir: "wiki", repoResolve: repoStub(nil)}
	if got := lexical.classify("../../etc/passwd"); got.class != classMalformed || got.resolved {
		t.Errorf("lexical repo escape = %+v, want malformed", got)
	}
	injected := &resolver{
		exists:      existsIn(),
		bundleDir:   "wiki",
		repoResolve: repoStub(map[string]RepoStatus{"shared/x.md": RepoEscape}),
	}
	if got := injected.classify("../shared/x.md"); got.class != classMalformed || got.resolved {
		t.Errorf("injected repo escape = %+v, want malformed", got)
	}
}

// (A8) The bundleDir join maps a bundle-escaping target onto its repo-root path:
// key is repo-class-qualified and normalized through the join.
func TestClassifyBundleDirJoin(t *testing.T) {
	r := &resolver{
		exists:      existsIn(),
		bundleDir:   "wiki",
		repoResolve: repoStub(map[string]RepoStatus{"docs/spec.md": RepoFound}),
	}
	got := r.classify("../docs/spec.md")
	if got.key != "repo:docs/spec.md" {
		t.Errorf("key = %q, want repo:docs/spec.md", got.key)
	}
}

// (A9) Normalization: fragment/query/`./` variants share one bundle key; bundle
// and repo keys never collide; url keys are trimmed-verbatim.
func TestClassifyNormalizationKeys(t *testing.T) {
	bundle := navResolver(existsIn("x.md"))
	for _, target := range []string{"x.md#s", "/x.md?q", "./x.md"} {
		if got := bundle.classify(target); got.key != "bundle:x.md" {
			t.Errorf("classify(%q).key = %q, want bundle:x.md", target, got.key)
		}
	}
	repo := &resolver{exists: existsIn(), bundleDir: "wiki", repoResolve: repoStub(map[string]RepoStatus{"x.md": RepoFound})}
	if got := repo.classify("../x.md"); got.key != "repo:x.md" {
		t.Errorf("repo ../x.md key = %q, want repo:x.md (must not collide with bundle:x.md)", got.key)
	}
	if got := navResolver(existsIn()).classify("  https://a.com/x  "); got.key != "url:https://a.com/x" {
		t.Errorf("url key = %q, want url:https://a.com/x (trimmed verbatim)", got.key)
	}
}

// ----------------------------------------------------------------------------
// Phase C — evidence-context splitting.
// ----------------------------------------------------------------------------

// evidenceTexts returns the concatenated text of the evidence segments as strings.
func evidenceTexts(segs []segment) []string {
	var out []string
	for _, s := range segs {
		if s.evidence {
			out = append(out, string(s.text))
		}
	}
	return out
}

// (C15) Heading match is exact and case-sensitive: `## Evidence` opens a context;
// `## evidence` and `## Evidence Notes` do not.
func TestSplitEvidenceContextsExactHeadingMatch(t *testing.T) {
	body := []byte("intro\n## Evidence\ncited\n## evidence\nnope\n## Evidence Notes\nnope2\n")
	segs := splitEvidenceContexts(body, []string{"Evidence"})
	ev := evidenceTexts(segs)
	if len(ev) != 1 {
		t.Fatalf("want exactly one evidence context, got %d (%q)", len(ev), ev)
	}
	if want := "## Evidence\ncited\n"; ev[0] != want {
		t.Errorf("evidence text = %q, want %q", ev[0], want)
	}
}

// (C16) Extent: a context ends at the next same-or-shallower heading; a deeper
// sub-heading stays inside; two matching headings open two contexts.
func TestSplitEvidenceContextsExtent(t *testing.T) {
	body := []byte("## Evidence\na\n### Sub\nb\n## Other\nc\n## Evidence\nd\n")
	segs := splitEvidenceContexts(body, []string{"Evidence"})
	ev := evidenceTexts(segs)
	if len(ev) != 2 {
		t.Fatalf("want two evidence contexts, got %d (%q)", len(ev), ev)
	}
	if want := "## Evidence\na\n### Sub\nb\n"; ev[0] != want {
		t.Errorf("first context = %q, want %q (sub-heading stays inside, ends at ## Other)", ev[0], want)
	}
	if want := "## Evidence\nd\n"; ev[1] != want {
		t.Errorf("second context = %q, want %q", ev[1], want)
	}
}

// (C17) Empty sections → one all-navigational segment (the zero-cost default).
func TestSplitEvidenceContextsEmptySectionsIsAllNavigational(t *testing.T) {
	body := []byte("## Evidence\nx\n")
	segs := splitEvidenceContexts(body, nil)
	if len(segs) != 1 || segs[0].evidence {
		t.Fatalf("empty sections should yield one navigational segment, got %+v", segs)
	}
	if string(segs[0].text) != string(body) {
		t.Errorf("segment text = %q, want whole body", segs[0].text)
	}
}

// ----------------------------------------------------------------------------
// Phase D — citation findings.
// ----------------------------------------------------------------------------

func citeRules(body string, res *resolver) []contract.Finding {
	return linkRules("page.md", []byte(body), res, []string{"Evidence"})
}

// (D18) A malformed citation target fires core-citation-malformed: warning,
// profile ruleset, page path, message naming the target.
func TestCitationMalformedFinding(t *testing.T) {
	got := citeRules("## Evidence\nSee [x](mailto:a@b.com).\n", navResolver(existsIn()))
	f, ok := firstWithCode(got, codeCoreCitationMalformed)
	if !ok {
		t.Fatalf("want core-citation-malformed, got %+v", got)
	}
	if f.Ruleset != contract.RulesetProfile || f.Severity != contract.SeverityWarning {
		t.Errorf("finding = %+v, want profile/warning", f)
	}
	if f.Path != "page.md" {
		t.Errorf("path = %q, want page.md", f.Path)
	}
	if !strings.Contains(f.Message, "mailto:a@b.com") {
		t.Errorf("message %q should name the target", f.Message)
	}
}

// (D19) An unresolved in-bundle citation subsumes the broken-link finding:
// exactly one core-citation-unresolved, zero core-broken-link.
func TestCitationUnresolvedSubsumesBrokenLink(t *testing.T) {
	got := citeRules("## Evidence\nSee [x](claims/missing.md).\n", navResolver(existsIn("page.md")))
	if n := len(findingsWithCode(got, codeCoreCitationUnresolved)); n != 1 {
		t.Fatalf("want exactly one core-citation-unresolved, got %d (%+v)", n, got)
	}
	if n := len(findingsWithCode(got, codeCoreBrokenLink)); n != 0 {
		t.Errorf("evidence target must not also fire core-broken-link, got %d (%+v)", n, got)
	}
}

// (D20) The same normalized target cited twice in one context → one
// core-citation-duplicate at suggestion severity.
func TestCitationDuplicateFinding(t *testing.T) {
	got := citeRules("## Evidence\n[a](x.md) and [b](./x.md#frag)\n", navResolver(existsIn("x.md")))
	dups := findingsWithCode(got, codeCoreCitationDuplicate)
	if len(dups) != 1 {
		t.Fatalf("want one core-citation-duplicate, got %d (%+v)", len(dups), got)
	}
	if dups[0].Severity != contract.SeveritySuggestion {
		t.Errorf("duplicate severity = %q, want suggestion", dups[0].Severity)
	}
}

// (D21) Duplicate scope is per-context: the same target in two evidence sections
// (once each) is not a duplicate.
func TestCitationDuplicateScopedPerContext(t *testing.T) {
	got := citeRules("## Evidence\n[a](x.md)\n## Evidence\n[b](x.md)\n", navResolver(existsIn("x.md")))
	if n := len(findingsWithCode(got, codeCoreCitationDuplicate)); n != 0 {
		t.Errorf("cross-context repetition should not be a duplicate, got %d (%+v)", n, got)
	}
}

// (D22) One evidence target yields one finding: a repeated malformed target
// yields only malformed; a repeated unresolved target yields only unresolved.
func TestCitationPrecedenceOneFindingPerTarget(t *testing.T) {
	mal := citeRules("## Evidence\n[a](mailto:x) [b](mailto:x)\n", navResolver(existsIn()))
	if n := len(findingsWithCode(mal, codeCoreCitationMalformed)); n != 1 {
		t.Errorf("want one malformed finding, got %d (%+v)", n, mal)
	}
	if n := len(findingsWithCode(mal, codeCoreCitationDuplicate)); n != 0 {
		t.Errorf("repeated malformed must not also be a duplicate, got %d (%+v)", n, mal)
	}
	unres := citeRules("## Evidence\n[a](gone.md) [b](gone.md)\n", navResolver(existsIn("page.md")))
	if n := len(findingsWithCode(unres, codeCoreCitationUnresolved)); n != 1 {
		t.Errorf("want one unresolved finding, got %d (%+v)", n, unres)
	}
	if n := len(findingsWithCode(unres, codeCoreCitationDuplicate)); n != 0 {
		t.Errorf("repeated unresolved must not also be a duplicate, got %d (%+v)", n, unres)
	}
}

// (D23) Resolved citation targets (present in-bundle, cited once) yield nothing.
func TestCitationResolvedTargetsYieldNothing(t *testing.T) {
	got := citeRules("## Evidence\n[a](x.md) and [b](https://example.com/p)\n", navResolver(existsIn("x.md")))
	if len(got) != 0 {
		t.Errorf("resolved citations should yield no findings, got %+v", got)
	}
}

// (D24) A broken link outside the evidence context still fires core-broken-link
// even when EvidenceSections is set.
func TestNavigationalUnaffectedByEvidenceConfig(t *testing.T) {
	got := citeRules("intro [nav](gone.md)\n## Evidence\n[cite](x.md)\n", navResolver(existsIn("page.md", "x.md")))
	if n := len(findingsWithCode(got, codeCoreBrokenLink)); n != 1 {
		t.Errorf("navigational broken link should still fire, got %d (%+v)", n, got)
	}
}

// (D25) Image links inside an evidence context are skipped (never citations).
func TestCitationImageLinksSkipped(t *testing.T) {
	got := citeRules("## Evidence\n![diagram](img/missing.png)\n", navResolver(existsIn()))
	if len(got) != 0 {
		t.Errorf("image links should be skipped in evidence contexts, got %+v", got)
	}
}
