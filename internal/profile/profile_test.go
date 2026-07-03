package profile

import (
	"errors"
	"testing"
)

// Resolve returns the shipped core profile for its id, carrying the pinned
// version the bundle config records.
func TestResolveCoreReturnsShippedProfile(t *testing.T) {
	p, err := Resolve(CoreID)
	if err != nil {
		t.Fatalf("Resolve(%q) returned error: %v", CoreID, err)
	}
	if p.ID != "core" {
		t.Errorf("ID = %q, want core", p.ID)
	}
	if p.Version != "0.1.0" {
		t.Errorf("Version = %q, want 0.1.0", p.Version)
	}
}

// Ids outside the shipped set — including the not-yet-shipped academic profile,
// the empty id, and an arbitrary unknown — all resolve to ErrUnknownProfile.
// Defaulting to core is the CLI's job, not Resolve's, so "" is unknown here.
func TestResolveUnknownProfileIsError(t *testing.T) {
	for _, id := range []string{"academic-research", "", "bogus"} {
		if _, err := Resolve(id); !errors.Is(err, ErrUnknownProfile) {
			t.Errorf("Resolve(%q) error = %v, want ErrUnknownProfile", id, err)
		}
	}
}

type stubAdapter struct{}

func (stubAdapter) Unmarshal([]byte, any) error { return nil }
func (stubAdapter) Marshal(any) ([]byte, error) { return nil, nil }

// ADR-001 requires all YAML access to route through the internal adapter
// interface, never goccy/go-yaml at call sites. The loader must therefore hold
// exactly the adapter it was given.
func TestNewLoaderRoutesYAMLThroughInjectedAdapter(t *testing.T) {
	a := stubAdapter{}

	l := NewLoader(a)
	if l.yaml != a {
		t.Error("NewLoader must store the injected YAML adapter so all YAML access is routed through it")
	}
}
