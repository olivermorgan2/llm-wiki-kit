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
func (e *Engine) Run(fsys fs.FS) []contract.Finding {
	out := []contract.Finding{}
	fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || path.Ext(p) != ".md" {
			return nil
		}
		content, readErr := fs.ReadFile(fsys, p)
		if readErr != nil {
			return nil
		}
		out = append(out, evaluatePage(e.yaml, p, content)...)
		return nil
	})
	return out
}
