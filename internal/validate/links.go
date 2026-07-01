package validate

import (
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

// linkRules reports broken intra-wiki links in a page body. Each inline
// `[text](target)` link whose target is a true intra-wiki reference is resolved
// bundle-root-relative and marked broken iff the cleaned path is not present in
// the bundle (exists reports membership over ALL bundle files, not just `.md`).
// External schemes, protocol-relative targets, and pure `#fragment` links are
// skipped, never fetched (external liveness is out of scope). A `#fragment` or
// `?query` suffix is stripped before resolution; anchors are not validated. A
// target escaping the bundle root via `../` counts as unresolved — the
// filesystem is never consulted for it (fs-safety enforcement is ADR-005 /
// issue #5).
//
// All broken targets on a page aggregate into a single warning-severity,
// profile-ruleset finding (ADR-004 FR8) so its {ruleset, code, path} fingerprint
// stays unique per page, matching the existing per-code aggregation pattern. A
// page with only valid or skipped links yields nothing.
func linkRules(pagePath string, body []byte, exists func(string) bool) []contract.Finding {
	var broken []string
	seen := map[string]bool{}

	for _, m := range inlineLink.FindAllSubmatch(body, -1) {
		if len(m[1]) > 0 {
			continue // image link `![alt](t)`: an asset, not a page.
		}
		target := strings.TrimSpace(string(m[2]))
		if !isIntraWiki(target) {
			continue
		}
		if resolved, ok := resolveTarget(target); ok && exists(resolved) {
			continue
		}
		if !seen[target] {
			seen[target] = true
			broken = append(broken, target)
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

// isIntraWiki reports whether target is a relative reference to another page in
// the bundle (the only class this rule evaluates). Targets with a URI scheme,
// protocol-relative targets (`//…`), pure fragments (`#…`), and empty targets are
// not intra-wiki and are skipped.
func isIntraWiki(target string) bool {
	if target == "" || strings.HasPrefix(target, "#") || strings.HasPrefix(target, "//") {
		return false
	}
	return !uriScheme.MatchString(target)
}

// resolveTarget strips any `#fragment`/`?query` suffix and cleans the target as a
// bundle-root-relative path. It returns the cleaned path and ok=true when the
// target names a resolvable in-bundle location; ok=false when there is nothing to
// resolve (a bare fragment/query) or the target escapes the bundle root via `../`
// (which counts as unresolved — broken).
func resolveTarget(target string) (string, bool) {
	if i := strings.IndexAny(target, "#?"); i >= 0 {
		target = target[:i]
	}
	target = strings.TrimSpace(target)
	if target == "" {
		return "", false
	}
	target = strings.TrimPrefix(target, "/") // bundle-root-absolute → root-relative
	cleaned := path.Clean(target)
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", false // escapes the bundle root
	}
	return cleaned, true
}
