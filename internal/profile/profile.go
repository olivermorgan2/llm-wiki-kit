// Package profile holds the data-driven validation profiles (one inheritance
// level: a profile extends the core profile). This skeleton fixes only the
// package seam; profile parsing and rule data are later Phase 1 work
// (ADR-004, ADR-007).
package profile

import "github.com/olivermorgan2/llm-wiki-kit/internal/yamladapter"

// Profile is a data-driven validation profile. Its fields are filled in when
// real profile loading lands.
type Profile struct {
	ID string
}

// Loader reads profiles from disk. All YAML access is routed through the
// injected yamladapter.Adapter (ADR-001), never goccy/go-yaml at call sites.
type Loader struct {
	yaml yamladapter.Adapter
}

// NewLoader returns a Loader that routes every YAML operation through the given
// adapter.
func NewLoader(a yamladapter.Adapter) *Loader {
	return &Loader{yaml: a}
}
