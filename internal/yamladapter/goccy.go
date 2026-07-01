package yamladapter

import (
	"errors"

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

// Unmarshal decodes YAML data into v via goccy. Fields present in the source
// but not modeled by v are ignored rather than rejected: the engine inspects
// only the fields it models, and unknown-field round-trip preservation
// (criterion 6) is Slice 2 authoring, not decode-time validation.
func (goccyAdapter) Unmarshal(data []byte, v any) error {
	return goccy.Unmarshal(data, v)
}

// errMarshalNotImplemented is returned by the Marshal stub. Round-trip
// preservation of unknown frontmatter fields (criterion 6) is authored in
// Slice 2; until then Marshal fails loudly rather than emitting lossy YAML.
var errMarshalNotImplemented = errors.New("yamladapter: Marshal is not implemented (round-trip preservation lands in Slice 2, criterion 6)")

// Marshal is a documented not-implemented stub. See errMarshalNotImplemented.
func (goccyAdapter) Marshal(v any) ([]byte, error) {
	return nil, errMarshalNotImplemented
}
