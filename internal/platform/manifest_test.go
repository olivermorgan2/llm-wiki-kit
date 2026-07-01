package platform

import (
	"bytes"
	"strings"
	"testing"
)

// Sum is the shared SHA-256 helper; the empty-input digest is a fixed known
// value, which pins the algorithm and lowercase-hex encoding.
func TestSumEmptyInput(t *testing.T) {
	got, err := Sum(strings.NewReader(""))
	if err != nil {
		t.Fatalf("Sum: %v", err)
	}
	const want = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if got != want {
		t.Errorf("Sum(\"\") = %q, want %q", got, want)
	}
}

func TestParseManifestParsesEntries(t *testing.T) {
	h1 := strings.Repeat("a", 64)
	h2 := strings.Repeat("b", 64)
	doc := h1 + "  bin/darwin_arm64/llm-wiki\n" + h2 + "  bin/linux_amd64/llm-wiki\n"

	m, err := ParseManifest(strings.NewReader(doc))
	if err != nil {
		t.Fatalf("ParseManifest: %v", err)
	}
	if len(m) != 2 {
		t.Fatalf("len = %d, want 2", len(m))
	}
	if got := m["bin/darwin_arm64/llm-wiki"]; got != h1 {
		t.Errorf("darwin entry = %q, want %q", got, h1)
	}
	if got := m["bin/linux_amd64/llm-wiki"]; got != h2 {
		t.Errorf("linux entry = %q, want %q", got, h2)
	}
}

// GNU sha256sum binary mode ("HASH *path") and blank lines must both be
// tolerated, and hashes are normalised to lowercase.
func TestParseManifestAcceptsBinaryModeBlankLinesAndUppercase(t *testing.T) {
	h := strings.Repeat("C", 64)
	doc := "\n" + h + " *bin/windows_amd64/llm-wiki.exe\n\n"

	m, err := ParseManifest(strings.NewReader(doc))
	if err != nil {
		t.Fatalf("ParseManifest: %v", err)
	}
	if got := m["bin/windows_amd64/llm-wiki.exe"]; got != strings.ToLower(h) {
		t.Errorf("entry = %q, want %q", got, strings.ToLower(h))
	}
}

func TestParseManifestRejectsMalformed(t *testing.T) {
	cases := map[string]string{
		"short hash":   "abc  bin/x\n",
		"non-hex hash": strings.Repeat("z", 64) + "  bin/x\n",
		"missing path": strings.Repeat("a", 64) + "  \n",
		"no separator": strings.Repeat("a", 64) + "bin/x\n",
	}
	for name, doc := range cases {
		if _, err := ParseManifest(strings.NewReader(doc)); err == nil {
			t.Errorf("%s: expected error, got nil", name)
		}
	}
}

func TestWriteManifestRoundTripsAndIsSorted(t *testing.T) {
	m := Manifest{
		"bin/linux_amd64/llm-wiki":  strings.Repeat("b", 64),
		"bin/darwin_arm64/llm-wiki": strings.Repeat("a", 64),
	}

	var buf bytes.Buffer
	if err := WriteManifest(&buf, m); err != nil {
		t.Fatalf("WriteManifest: %v", err)
	}

	// Output must be deterministic: entries sorted by path.
	wantFirst := strings.Repeat("a", 64) + "  bin/darwin_arm64/llm-wiki\n"
	if !strings.HasPrefix(buf.String(), wantFirst) {
		t.Errorf("output not sorted by path:\n%s", buf.String())
	}

	got, err := ParseManifest(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	if len(got) != len(m) {
		t.Fatalf("round-trip len = %d, want %d", len(got), len(m))
	}
	for k, v := range m {
		if got[k] != v {
			t.Errorf("round-trip[%q] = %q, want %q", k, got[k], v)
		}
	}
}
