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

// Marshal is a documented not-implemented stub in this issue; round-trip
// preservation (criterion 6) is Slice 2. It must return a clear error rather
// than silently producing lossy output.
func TestMarshalIsNotImplemented(t *testing.T) {
	_, err := New().Marshal(map[string]any{"type": "concept"})
	if err == nil {
		t.Fatal("Marshal stub returned nil error, want not-implemented error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "not implemented") {
		t.Errorf("Marshal error = %q, want it to mention 'not implemented'", err.Error())
	}
}
