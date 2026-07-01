// Package yamladapter defines the internal seam through which all engine YAML
// access flows.
//
// ADR-001 adopts github.com/goccy/go-yaml but confines it behind this
// interface: call sites in profile/ and validate/ depend on Adapter, never on
// goccy directly, so a future toolchain revision changes only the concrete
// implementation and go.mod. The concrete node-aware implementation — which
// must preserve unknown frontmatter fields on round-trip (criterion 6) — lands
// with the validation issue, not this skeleton.
package yamladapter

// Adapter is the engine's single point of YAML serialization.
type Adapter interface {
	// Unmarshal decodes YAML data into v.
	Unmarshal(data []byte, v any) error
	// Marshal encodes v to YAML. The concrete implementation preserves unknown
	// fields present in the source on round-trip (ADR-001, criterion 6).
	Marshal(v any) ([]byte, error)
}
