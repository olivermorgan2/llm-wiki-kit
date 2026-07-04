package yamladapter

import (
	"strings"
	"testing"
)

// New returns a concrete Adapter, so the seam declared in yamladapter.go has a
// real implementation callers can construct.
func TestNewReturnsAdapter(t *testing.T) {
	var _ Adapter = New()
}

func TestUnmarshalDecodesIntoMap(t *testing.T) {
	var m map[string]any
	if err := New().Unmarshal([]byte("type: concept\ntitle: Alpha\n"), &m); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if m["type"] != "concept" {
		t.Errorf("type = %v, want concept", m["type"])
	}
	if m["title"] != "Alpha" {
		t.Errorf("title = %v, want Alpha", m["title"])
	}
}

func TestUnmarshalDecodesIntoStruct(t *testing.T) {
	var page struct {
		Type  string `yaml:"type"`
		Title string `yaml:"title"`
	}
	if err := New().Unmarshal([]byte("type: concept\ntitle: Alpha\n"), &page); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if page.Type != "concept" || page.Title != "Alpha" {
		t.Errorf("decoded = %+v, want {concept Alpha}", page)
	}
}

// Unknown fields — ones the target struct does not model — must decode without
// error. The engine only inspects modeled fields; unknown-field preservation
// (criterion 6) is Slice 2 authoring, but decoding must not reject them.
func TestUnmarshalIgnoresUnknownFields(t *testing.T) {
	var page struct {
		Type string `yaml:"type"`
	}
	err := New().Unmarshal([]byte("type: concept\ncustom_field: kept\n"), &page)
	if err != nil {
		t.Fatalf("Unmarshal rejected unknown field: %v", err)
	}
	if page.Type != "concept" {
		t.Errorf("type = %q, want concept", page.Type)
	}
}

func TestUnmarshalMalformedYAMLReturnsError(t *testing.T) {
	var m map[string]any
	// An unterminated flow mapping is unambiguously malformed; goccy surfaces a
	// decode error. (Note: goccy treats an unterminated flow *sequence* like
	// "[broken" leniently as a plain scalar, so that is not a parse failure.)
	if err := New().Unmarshal([]byte("title: {broken\n"), &m); err == nil {
		t.Fatal("Unmarshal accepted malformed YAML, want error")
	}
}

// Marshal encodes a value back to YAML (criterion 6 groundwork for page plan).
func TestMarshalEncodesValue(t *testing.T) {
	out, err := New().Marshal(map[string]any{"type": "concept"})
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	if !strings.Contains(string(out), "type: concept") {
		t.Errorf("Marshal output = %q, want it to contain 'type: concept'", out)
	}
}

// An OrderedMap round-trips: decoding YAML into one and marshaling it back
// preserves key order and every field, including ones the engine does not model
// (unknown-field preservation, ADR-001 criterion 6). This is the property page
// plan relies on so unknown frontmatter survives the staged-mutation cycle.
func TestOrderedMapRoundTripPreservesUnknownFields(t *testing.T) {
	src := "title: Alpha\ntype: concept\ncustom_field: keep-me\nnested:\n  a: 1\n  b: two\n"
	a := New()

	var m OrderedMap
	if err := a.Unmarshal([]byte(src), &m); err != nil {
		t.Fatalf("Unmarshal into OrderedMap: %v", err)
	}
	out, err := a.Marshal(m)
	if err != nil {
		t.Fatalf("Marshal OrderedMap: %v", err)
	}

	got := string(out)
	for _, field := range []string{"title: Alpha", "type: concept", "custom_field: keep-me"} {
		if !strings.Contains(got, field) {
			t.Errorf("round-trip dropped %q:\n%s", field, got)
		}
	}
	// Key order is preserved: title precedes type precedes custom_field.
	if ti, tj := strings.Index(got, "title"), strings.Index(got, "type"); ti < 0 || tj < 0 || ti > tj {
		t.Errorf("key order not preserved (title should precede type):\n%s", got)
	}
	// Re-marshaling the round-tripped output is a fixed point (idempotent).
	var m2 OrderedMap
	if err := a.Unmarshal(out, &m2); err != nil {
		t.Fatalf("re-unmarshal: %v", err)
	}
	out2, err := a.Marshal(m2)
	if err != nil {
		t.Fatalf("re-marshal: %v", err)
	}
	if string(out2) != got {
		t.Errorf("round-trip not idempotent:\n first: %q\nsecond: %q", got, out2)
	}
}
