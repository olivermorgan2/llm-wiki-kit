package profile

import "testing"

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
