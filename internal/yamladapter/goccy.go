package yamladapter

import (
	goccy "github.com/goccy/go-yaml"
)

// goccyAdapter is the concrete Adapter backed by github.com/goccy/go-yaml
// (ADR-001). It is the single place in the engine that imports goccy; every
// other package depends only on the Adapter interface.
type goccyAdapter struct{}

// New returns the concrete goccy-backed Adapter.
func New() Adapter {
	return goccyAdapter{}
}

// OrderedMap is an order-preserving YAML mapping. Decoding YAML into an
// *OrderedMap and re-encoding it round-trips every field — including ones the
// engine does not model — with key order intact (ADR-001, criterion 6). It is
// the representation page plan/apply use so unknown frontmatter survives the
// staged-mutation cycle. The alias keeps goccy confined to this file: callers
// build on yamladapter.OrderedMap and never import goccy, so a toolchain swap
// stays localized here.
type OrderedMap = goccy.MapSlice

// Unmarshal decodes YAML data into v via goccy. Fields present in the source
// but not modeled by v are ignored rather than rejected: the engine inspects
// only the fields it models, and unknown-field round-trip preservation
// (criterion 6) is Slice 2 authoring, not decode-time validation.
func (goccyAdapter) Unmarshal(data []byte, v any) error {
	return goccy.Unmarshal(data, v)
}

// Marshal encodes v to YAML via goccy. Decoding frontmatter into an OrderedMap
// and marshaling it back preserves key order and every field present in the
// source, including ones the engine does not model — the round-trip guarantee
// page plan/apply rely on so unknown frontmatter survives the staged-mutation
// cycle (ADR-001, criterion 6).
func (goccyAdapter) Marshal(v any) ([]byte, error) {
	return goccy.Marshal(v)
}
