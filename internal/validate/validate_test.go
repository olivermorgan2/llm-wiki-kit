package validate

import "testing"

// The skeleton wires the validation seam but implements no rules yet (rule
// content is later Phase 1 work, ADR-004). Run must therefore always return an
// empty, non-nil slice so callers can range over it uniformly.
func TestEngineRunReturnsNoFindingsYet(t *testing.T) {
	e := New(nil)

	got := e.Run([]string{"pages/example.md"})
	if got == nil {
		t.Fatal("Run must return a non-nil slice, even with no rules implemented")
	}
	if len(got) != 0 {
		t.Errorf("skeleton must produce no findings yet, got %d", len(got))
	}
}
