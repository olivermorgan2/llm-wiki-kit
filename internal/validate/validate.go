// Package validate is the single deterministic validation engine. It emits
// ruleset-tagged (OKF vs profile), three-severity findings for both base OKF
// conformance and core-profile conformance from one pass (ADR-004). One engine
// keeps findings identical across skills/hooks/CI/CLI (criterion 15).
//
// The engine produces findings at their core-default severity. The remaining
// ADR-004 precedence layers are applied by the caller in fixed order: Resolve
// (profile overrides) then ApplyBaseline (differential filter), with StatusFor
// reducing the result to an envelope status.
package validate

import (
	"io/fs"
	"path"

	"github.com/olivermorgan2/llm-wiki-kit/internal/contract"
	"github.com/olivermorgan2/llm-wiki-kit/internal/profile"
	"github.com/olivermorgan2/llm-wiki-kit/internal/yamladapter"
)

// Engine runs OKF and core-profile validation over a filesystem of pages. All
// YAML access flows through the injected adapter (ADR-001); the engine never
// imports goccy directly.
type Engine struct {
	yaml yamladapter.Adapter
	opts Options
}

// Options configures ADR-008 citation resolution. The zero value reproduces the
// pre-citation engine exactly: no link is ever a citation (empty
// EvidenceSections), and the repo-path class is empty (nil RepoResolve), so only
// the navigational broken-link rule runs.
type Options struct {
	// EvidenceSections lists the ATX heading titles that open a profile-designated
	// evidence context. Empty means no link is ever classified as a citation.
	EvidenceSections []string
	// BundleDir is the bundle root relative to the repo root in slash form
	// ("" when equal); used to map bundle-escaping `../` targets into the repo.
	BundleDir string
	// RepoResolve is the read-only repo-path existence check (ADR-008
	// sub-decision 3). nil means no `.llm-wiki/` anchor: the repo-path class is
	// empty and every bundle-escaping target is malformed.
	RepoResolve func(string) RepoStatus
	// Profile is the resolved active profile (ADR-007/ADR-010). Its per-type rules
	// drive the type-conditional structural findings (profile-*). The zero value
	// (no Types) runs no profile rules, so a core-profile bundle validates exactly
	// as before Phase 4 (golden parity, ADR-010 sub-decision 2).
	Profile profile.Profile
}

// New returns a validation Engine that decodes YAML through the given adapter,
// with citation resolution disabled (the zero Options).
func New(yaml yamladapter.Adapter) *Engine {
	return &Engine{yaml: yaml}
}

// NewWithOptions returns a validation Engine wired with ADR-008 citation
// resolution options. Callers build opts via AnchorRepo (repo-path anchor) and,
// in Phase 4, profile-loaded EvidenceSections.
func NewWithOptions(yaml yamladapter.Adapter, opts Options) *Engine {
	return &Engine{yaml: yaml, opts: opts}
}

// Run walks fsys for markdown pages and returns every finding at its core-default
// severity. The walk is deterministic (fs.WalkDir visits entries in lexical
// order) and recurses into subdirectories; only files ending in `.md` are
// validated. The result is always non-nil. Filesystem read anomalies are out of
// scope here (the safe-filesystem layer is ADR-005 / issue #5): an unreadable
// entry is skipped rather than reported as a validation finding.
//
// The single walk also records every bundle file path (not just `.md`) so the
// intra-wiki broken-link rule can resolve link targets against the full bundle;
// links may point forward in the walk, so pages are evaluated after the walk
// completes, preserving the deterministic lexical order.
func (e *Engine) Run(fsys fs.FS) []contract.Finding {
	out := []contract.Finding{}
	files := map[string]bool{}
	type page struct {
		path    string
		content []byte
	}
	var pages []page
	fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		files[p] = true
		if path.Ext(p) != ".md" {
			return nil
		}
		content, readErr := fs.ReadFile(fsys, p)
		if readErr != nil {
			return nil
		}
		pages = append(pages, page{path: p, content: content})
		return nil
	})

	res := &resolver{
		exists:      func(target string) bool { return files[target] },
		bundleDir:   e.opts.BundleDir,
		repoResolve: e.opts.RepoResolve,
	}

	// Pre-pass: record every page's frontmatter `type` so a cross-page citation
	// rule (a claim citing a `question`, ADR-010 sub-decision 4) can read the
	// target page's type. Links may point forward in the walk, so all types must
	// be known before any page is evaluated. Unparseable frontmatter contributes
	// no type (its page gets okf-yaml-parse and no citation-target verdict).
	pageTypes := map[string]string{}
	for _, pg := range pages {
		if fm, _, err := splitFrontmatter(pg.content); err == nil {
			var mm map[string]any
			if e.yaml.Unmarshal(fm, &mm) == nil {
				if t, ok := mm["type"].(string); ok {
					pageTypes[pg.path] = t
				}
			}
		}
	}

	for _, pg := range pages {
		out = append(out, evaluatePage(e.yaml, pg.path, pg.content)...)
		// Link/citation resolution and the profile rules need the page body and
		// frontmatter; a page whose frontmatter fails to split already yields
		// okf-yaml-parse and gets none of them (the parse-failure gate is
		// inherited). splitFrontmatter is re-run here rather than threaded out of
		// evaluatePage so evaluatePage's signature — and its many unit-test call
		// sites — stay unchanged.
		if fm, body, err := splitFrontmatter(pg.content); err == nil {
			// Profile type/citation rules run only over frontmatter that also
			// parses as YAML (the same gate evaluatePage applies).
			var m map[string]any
			parsed := e.yaml.Unmarshal(fm, &m) == nil
			pageType := ""
			if parsed {
				pageType, _ = m["type"].(string)
			}
			// A page's evidence contexts are the global (env) designation unioned
			// with its profile type's evidence sections (ADR-010 sub-decision 4).
			// core-citation-* findings evaluate over this union so env-designated
			// sections keep working (Phase 3 parity); the citation OBLIGATION uses
			// the profile's per-type sections only (profileCitationFindings).
			evSections := evidenceSectionsFor(e.opts, pageType)
			out = append(out, linkRules(pg.path, body, res, evSections)...)
			if parsed {
				out = append(out, profileTypeFindings(e.opts.Profile, pg.path, m, body)...)
				out = append(out, profileCitationFindings(e.opts.Profile, pg.path, m, body, res, pageTypes)...)
			}
		}
	}
	return out
}

// evidenceSectionsFor returns the evidence-context section titles that apply to a
// page of pageType: the global Options.EvidenceSections (the demoted
// LLM_WIKI_EVIDENCE_SECTIONS override) unioned with the resolved profile type's
// EvidenceSections. The union is order-stable and deduplicated; an empty result
// is the zero-cost default where no link is ever a citation.
func evidenceSectionsFor(opts Options, pageType string) []string {
	profSections := opts.Profile.Types[pageType].EvidenceSections
	if len(opts.EvidenceSections) == 0 && len(profSections) == 0 {
		return nil
	}
	seen := map[string]bool{}
	var out []string
	for _, s := range opts.EvidenceSections {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	for _, s := range profSections {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}
