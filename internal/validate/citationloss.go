package validate

import "strings"

// CitationLoss compares the normalized citation-target sets of sourceBody and
// stagedBody — the union across the opts.EvidenceSections evidence contexts of
// each body (ADR-008 sub-decision 6) — and returns the targets present in the
// source union but absent from the staged union. Lost targets are returned in
// source order carrying their first-seen trimmed spelling. An empty
// EvidenceSections (no context is designated) or no net loss returns nil.
//
// Targets are keyed by the shared resolver's class-qualified normalized key, so
// a spelling change that normalizes equal is not a loss, a case-differing URL is
// (sub-decision 6 forbids host folding), and set semantics mean dropping one of
// several duplicate occurrences is not a loss. Every non-image inline-link target
// in an evidence context counts, malformed ones included — dropping a malformed
// citation is still a gated removal. Resolved-ness never affects the key, so the
// diff never depends on whether a target currently exists.
func CitationLoss(sourceBody, stagedBody []byte, opts Options) []string {
	if len(opts.EvidenceSections) == 0 {
		return nil
	}
	// The single shared resolver, built from opts. exists is irrelevant to keys
	// (resolved-ness never changes a key), so it is stubbed false; bundleDir and
	// repoResolve do split the repo/bundle classes and so come from opts.
	res := &resolver{
		exists:      func(string) bool { return false },
		bundleDir:   opts.BundleDir,
		repoResolve: opts.RepoResolve,
	}

	srcOrder, srcSpelling := collectEvidenceTargets(sourceBody, opts.EvidenceSections, res)
	stagedOrder, _ := collectEvidenceTargets(stagedBody, opts.EvidenceSections, res)
	stagedSet := make(map[string]bool, len(stagedOrder))
	for _, k := range stagedOrder {
		stagedSet[k] = true
	}

	var lost []string
	for _, k := range srcOrder {
		if !stagedSet[k] {
			lost = append(lost, srcSpelling[k])
		}
	}
	return lost
}

// collectEvidenceTargets returns the distinct citation-target keys found across
// body's evidence contexts (first-seen order) and a map from each key to the
// first trimmed spelling that produced it. Image links are skipped (never a
// citation); every other inline-link target is classified once by the shared
// resolver. Links outside an evidence context are ignored — only evidence-context
// citations are subject to the preservation gate.
func collectEvidenceTargets(body []byte, sections []string, res *resolver) (order []string, spelling map[string]string) {
	spelling = map[string]string{}
	for _, seg := range splitEvidenceContexts(body, sections) {
		if !seg.evidence {
			continue
		}
		for _, m := range inlineLink.FindAllSubmatch(seg.text, -1) {
			if len(m[1]) > 0 {
				continue // image link: an asset, never a citation.
			}
			target := strings.TrimSpace(string(m[2]))
			key := res.classify(target).key
			if _, seen := spelling[key]; !seen {
				spelling[key] = target
				order = append(order, key)
			}
		}
	}
	return order, spelling
}
