// Package profile holds the data-driven validation profiles (one inheritance
// level: a profile extends the core profile). This skeleton fixes only the
// package seam; profile parsing and rule data are later Phase 1 work
// (ADR-004, ADR-007).
package profile

import (
	"errors"
	"fmt"

	"github.com/olivermorgan2/llm-wiki-kit/internal/yamladapter"
)

// CoreID is the id of the shipped core profile — the default profile init
// materializes a reference to (ADR-007).
const CoreID = "core"

// ErrUnknownProfile is returned by Resolve for any id outside the shipped set.
var ErrUnknownProfile = errors.New("profile: unknown profile id")

// Profile is a data-driven validation profile. Version is the pinned profile
// version the bundle config records (ADR-007); the remaining fields are filled
// in when real profile loading lands.
type Profile struct {
	ID      string
	Version string
}

// Resolve returns the shipped profile registered under id, or ErrUnknownProfile
// for any unregistered id (including the not-yet-shipped academic-research
// profile and the empty id — defaulting to core is the caller's responsibility,
// not Resolve's). Resolution is by id against an in-code registry: init reads no
// rule data, so no embed.FS or profile YAML is consulted here. The Phase-7
// local-path profile resolution seam (ADR-007) attaches at this function — a
// filesystem-backed lookup would extend, not replace, this shipped-set switch.
func Resolve(id string) (Profile, error) {
	switch id {
	case CoreID:
		return Profile{ID: CoreID, Version: "0.1.0"}, nil
	default:
		return Profile{}, fmt.Errorf("%w: %q", ErrUnknownProfile, id)
	}
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
