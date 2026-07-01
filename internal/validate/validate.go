// Package validate is the single deterministic validation engine. It will emit
// ruleset-tagged, three-severity findings for both base OKF conformance and
// profile conformance from one pass (ADR-004). This skeleton wires the seam but
// implements no rules yet — that is later Phase 1 work.
package validate

import (
	"github.com/olivermorgan2/llm-wiki-kit/internal/contract"
	"github.com/olivermorgan2/llm-wiki-kit/internal/profile"
)

// Engine runs OKF and profile validation over a set of paths.
type Engine struct {
	// Profiles supplies the active profile and its rule data. It is nil in the
	// skeleton because no rules run yet.
	Profiles *profile.Loader
}

// New returns a validation Engine using the given profile loader.
func New(loader *profile.Loader) *Engine {
	return &Engine{Profiles: loader}
}

// Run validates the given repository paths and returns the findings. The
// skeleton implements no rules, so it always returns an empty, non-nil slice;
// real OKF/profile rules and severities are a later Phase 1 issue (ADR-004).
func (e *Engine) Run(paths []string) []contract.Finding {
	return []contract.Finding{}
}
