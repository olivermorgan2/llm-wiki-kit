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
	"github.com/olivermorgan2/llm-wiki-kit/internal/yamladapter"
)

// Engine runs OKF and core-profile validation over a filesystem of pages. All
// YAML access flows through the injected adapter (ADR-001); the engine never
// imports goccy directly.
type Engine struct {
	yaml yamladapter.Adapter
}

// New returns a validation Engine that decodes YAML through the given adapter.
func New(yaml yamladapter.Adapter) *Engine {
	return &Engine{yaml: yaml}
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

	exists := func(target string) bool { return files[target] }
	for _, pg := range pages {
		out = append(out, evaluatePage(e.yaml, pg.path, pg.content)...)
		// Link resolution needs the page body; a page whose frontmatter fails to
		// split already yields okf-yaml-parse and has no link findings.
		if _, body, err := splitFrontmatter(pg.content); err == nil {
			out = append(out, linkRules(pg.path, body, exists)...)
		}
	}
	return out
}
