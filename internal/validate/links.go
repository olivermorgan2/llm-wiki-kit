package validate

import (
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/olivermorgan2/llm-wiki-kit/internal/contract"
)

// inlineLink matches an inline Markdown link `[text](target)` and captures a
// leading `!` (image marker) and the raw target. RE2 has no lookbehind, so the
// optional bang is captured explicitly and used to skip image links (assets, not
// pages). Reference-style links, autolinks, and raw HTML are out of scope for
// this issue (see notes/eval-issue-004.md follow-ups).
var inlineLink = regexp.MustCompile(`(!?)\[[^\]]*\]\(([^)]*)\)`)

// uriScheme matches a leading URI scheme like `http:`, `https:`, or `mailto:`.
var uriScheme = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]*:`)

// RepoStatus is the read-only verdict of the repo-path existence check
// (ADR-008 sub-decision 3). It is produced by a `stat` bounded by the repo-root
// anchor and the ADR-005 canonicalize/resolve-symlink primitives; the resolver
// never reads file contents and never stats above the anchor.
type RepoStatus int

const (
	// RepoAbsent means the target is well-formed and inside the repo root but no
	// file is there → unresolved (may become resolvable when the file is created).
	RepoAbsent RepoStatus = iota
	// RepoFound means a read-only stat inside the repo root succeeded → resolved.
	RepoFound
	// RepoEscape means the target canonicalizes outside the repo root → malformed
	// (an escape is never followed).
	RepoEscape
)

// resolver is ADR-008's single shared three-class citation/link resolver. One
// resolver classifies both navigational links (core-broken-link) and citation
// targets (core-citation-*) so there is exactly one notion of "resolvable" and
// one notion of "escape" (sub-decisions 2, 7).
type resolver struct {
	// exists reports in-bundle membership over ALL bundle files (the shipped
	// injected seam; links.go touches no disk for this class).
	exists func(string) bool
	// bundleDir is the bundle root relative to the repo root in slash form
	// ("" when they are the same directory); used to map a bundle-escaping `../`
	// target onto a repo-root-relative path.
	bundleDir string
	// repoResolve performs the ADR-008 sub-decision-3 read-only repo-path check.
	// nil means no `.llm-wiki/` anchor was found: the repo-path class is empty and
	// every bundle-escaping target is malformed (bundle-root fallback).
	repoResolve func(string) RepoStatus
}

// linkClass is the ADR-008 three-class partition plus the malformed catch-all.
type linkClass int

const (
	classExternal linkClass = iota // http(s) URL
	classInBundle                  // scheme-less reference inside the bundle root
	classRepo                      // bundle-escaping `../` resolving inside the repo root
	classMalformed                 // can never resolve
)

// resolution is the total verdict for one target. resolved is always true for a
// well-formed classExternal and always false for classMalformed. key is the
// class-qualified normalized dedupe key: bundle:/repo: targets normalize to the
// cleaned path, url: targets to the trimmed verbatim string, and malformed
// targets to their trimmed verbatim spelling. The class prefix is load-bearing:
// under a non-root bundle, bundle target `x.md` and repo target `../x.md` both
// clean to a path spelled `x.md` but name different files, so unprefixed keys
// would create false core-citation-duplicate findings (ambiguity #1).
type resolution struct {
	class    linkClass
	resolved bool
	key      string
}

// classify implements ADR-008 sub-decision 2's ordered, first-match-wins, total
// classification. The final catch-all guarantees no target is left without a
// verdict.
func (r *resolver) classify(target string) resolution {
	target = strings.TrimSpace(target)

	// 1. http(s) URL: resolvable iff a syntactically valid absolute URL with a
	// non-empty host. An invalid-host http(s) target is malformed, not unresolved
	// (it can never resolve). Liveness is never fetched. The key is the trimmed
	// target compared verbatim (sub-decision 6: no host case-folding or fragment
	// stripping — offline determinism over URL cleverness).
	if isHTTPScheme(target) {
		if u, err := url.Parse(target); err == nil && u.Host != "" {
			return resolution{class: classExternal, resolved: true, key: "url:" + target}
		}
		return resolution{class: classMalformed, resolved: false, key: "url:" + target}
	}

	// 2. Non-intra-wiki shapes (empty, `#frag`, `//host`, any non-http(s) URI
	// scheme) are malformed — none names a resolvable citation target.
	if !isIntraWiki(target) {
		return resolution{class: classMalformed, resolved: false, key: "malformed:" + target}
	}

	// 3/4. Scheme-less intra-wiki reference: apply the shipped resolveTarget
	// cleaning verbatim (strip `#fragment`/`?query`, strip a leading `/` as
	// bundle-root-absolute, path.Clean). Empty after strip → malformed.
	cleaned, ok := cleanTarget(target)
	if !ok {
		return resolution{class: classMalformed, resolved: false, key: "malformed:" + target}
	}
	if cleaned != ".." && !strings.HasPrefix(cleaned, "../") {
		// Inside the bundle root: in-bundle OKF document class.
		return resolution{class: classInBundle, resolved: r.exists(cleaned), key: "bundle:" + cleaned}
	}

	// Bundle-escaping `../`: the repo-path class (sub-decision 3's `../`
	// refinement). No anchor → malformed (bundle-root fallback).
	if r.repoResolve == nil {
		return resolution{class: classMalformed, resolved: false, key: "malformed:" + target}
	}
	repoRel := path.Clean(path.Join(r.bundleDir, cleaned))
	if repoRel == ".." || strings.HasPrefix(repoRel, "../") {
		return resolution{class: classMalformed, resolved: false, key: "malformed:" + target}
	}
	switch r.repoResolve(repoRel) {
	case RepoFound:
		return resolution{class: classRepo, resolved: true, key: "repo:" + repoRel}
	case RepoEscape:
		return resolution{class: classMalformed, resolved: false, key: "malformed:" + target}
	default: // RepoAbsent
		return resolution{class: classRepo, resolved: false, key: "repo:" + repoRel}
	}
}

// isHTTPScheme reports whether target begins with a case-insensitive
// `http:`/`https:` scheme.
func isHTTPScheme(target string) bool {
	lower := strings.ToLower(target)
	return strings.HasPrefix(lower, "http:") || strings.HasPrefix(lower, "https:")
}

// linkRules reports navigational broken links and citation findings for a page
// body. The body is split into evidence contexts (ADR-008 sub-decision 1): links
// inside a profile-designated evidence context are citation targets classified
// totally (core-citation-*), while links elsewhere keep the unchanged
// navigational broken-link behavior. There is one resolver and each target is
// classified once; its context decides which rule owns it (sub-decision 7), so a
// single bad link is never two findings.
//
// The one deliberate change to shipped navigational behavior is the ADR-008
// `../` widening: a bundle-escaping target that resolves inside the repo root
// (via an anchored repoResolve) no longer fires core-broken-link. With no anchor
// the repo-path class is empty and every `../` escape is still broken, exactly as
// before. An empty evidenceSections is the zero-cost default: the whole body is
// one navigational segment and no link is ever a citation.
func linkRules(pagePath string, body []byte, res *resolver, evidenceSections []string) []contract.Finding {
	segments := splitEvidenceContexts(body, evidenceSections)
	var out []contract.Finding
	out = append(out, res.brokenLinkFinding(pagePath, segments)...)
	out = append(out, res.citationFindings(pagePath, segments)...)
	return out
}

// brokenLinkFinding aggregates every unresolved navigational link across the
// non-evidence segments into a single warning-severity, profile-ruleset finding
// (ADR-004 FR8), matching the shipped per-page aggregation. A target that fails
// to resolve is broken whether the classifier calls it unresolved or malformed —
// e.g. `[x](?v=1)` empties on cleaning and stays broken. Only the `../` widening
// changes: a repo-resolving escape is no longer broken.
func (r *resolver) brokenLinkFinding(pagePath string, segments []segment) []contract.Finding {
	var broken []string
	seen := map[string]bool{}
	for _, seg := range segments {
		if seg.evidence {
			continue // evidence targets are citations, never broken-links (subsumption).
		}
		for _, m := range inlineLink.FindAllSubmatch(seg.text, -1) {
			if len(m[1]) > 0 {
				continue // image link `![alt](t)`: an asset, not a page.
			}
			target := strings.TrimSpace(string(m[2]))
			if !isIntraWiki(target) {
				continue // external/protocol-relative/fragment: skipped, never fetched.
			}
			if r.classify(target).resolved {
				continue
			}
			if !seen[target] {
				seen[target] = true
				broken = append(broken, target)
			}
		}
	}
	if len(broken) == 0 {
		return nil
	}
	return []contract.Finding{{
		Ruleset:  contract.RulesetProfile,
		Severity: contract.SeverityWarning,
		Code:     codeCoreBrokenLink,
		Message:  "broken intra-wiki link(s): " + strings.Join(broken, ", "),
		Path:     pagePath,
	}}
}

// isIntraWiki reports whether target is a relative reference into the bundle/repo
// (the only class the navigational rule evaluates). Targets with a URI scheme,
// protocol-relative targets (`//…`), pure fragments (`#…`), and empty targets are
// not intra-wiki and are skipped navigationally.
func isIntraWiki(target string) bool {
	if target == "" || strings.HasPrefix(target, "#") || strings.HasPrefix(target, "//") {
		return false
	}
	return !uriScheme.MatchString(target)
}

// cleanTarget strips any `#fragment`/`?query` suffix and cleans the target as a
// bundle-root-relative path (leading `/` treated as bundle-root-absolute). It
// returns the cleaned path and ok=true when there is something to resolve; ok is
// false only when the target empties (a bare fragment/query). Unlike the shipped
// resolveTarget, a `../` escape still returns ok=true with the cleaned path, so
// classify can route it to the repo-path class rather than treating every `../`
// as unresolved.
func cleanTarget(target string) (string, bool) {
	if i := strings.IndexAny(target, "#?"); i >= 0 {
		target = target[:i]
	}
	target = strings.TrimSpace(target)
	if target == "" {
		return "", false
	}
	target = strings.TrimPrefix(target, "/") // bundle-root-absolute → root-relative
	if target == "" {
		return "", false
	}
	return path.Clean(target), true
}
