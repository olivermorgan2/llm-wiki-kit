package txn

import (
	"errors"
	"io/fs"
	"strings"
	"testing"
)

// TestManifestRoundTripSortsEntries confirms marshalManifest emits entries in
// lexicographic path order (the commit order) regardless of input order, and
// that decoding reproduces the same records.
func TestManifestRoundTripSortsEntries(t *testing.T) {
	m := manifest{
		Version: manifestVersion,
		Txn:     "0123456789abcdef",
		Entries: []manifestEntry{
			{Path: "docs/b.md", Mode: 0o644, Staged: hashBytes([]byte("b")), Base: absentBase()},
			{Path: ".llm-wiki/manifest.json", Mode: 0o644, Staged: hashBytes([]byte("m")),
				Base: regularBase(hashBytes([]byte("old")), 0o644)},
			{Path: "docs/a.md", Mode: 0o600, Staged: hashBytes([]byte("a")),
				Base: regularBase(hashBytes([]byte("")), 0o600)},
		},
	}

	data, err := marshalManifest(m)
	if err != nil {
		t.Fatalf("marshalManifest: %v", err)
	}

	got, err := decodeManifest(data)
	if err != nil {
		t.Fatalf("decodeManifest: %v", err)
	}

	wantOrder := []string{".llm-wiki/manifest.json", "docs/a.md", "docs/b.md"}
	if len(got.Entries) != len(wantOrder) {
		t.Fatalf("entries = %d, want %d", len(got.Entries), len(wantOrder))
	}
	for i, w := range wantOrder {
		if got.Entries[i].Path != w {
			t.Fatalf("entry[%d].Path = %q, want %q", i, got.Entries[i].Path, w)
		}
	}
	if got.Version != manifestVersion || got.Txn != m.Txn {
		t.Fatalf("header = {v:%d txn:%q}, want {v:%d txn:%q}", got.Version, got.Txn, manifestVersion, m.Txn)
	}
}

// TestManifestRejectsMalformedBase covers the sentinel XOR rule: a base record
// must be exactly one of absent or hashed, never both and never neither.
func TestManifestRejectsMalformedBase(t *testing.T) {
	cases := map[string]string{
		"both":    `{"version":1,"txn":"x","entries":[{"path":"a","mode":420,"staged":"h","base":{"absent":true,"sha256":"h"}}]}`,
		"neither": `{"version":1,"txn":"x","entries":[{"path":"a","mode":420,"staged":"h","base":{}}]}`,
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := decodeManifest([]byte(raw)); err == nil {
				t.Fatalf("decodeManifest(%s) = nil error, want validation error", name)
			}
		})
	}
}

// TestAbsentIsNotEmptyHash guards ADR-006's absent != empty distinction: the
// absent sentinel must never collide with the hash of empty content.
func TestAbsentIsNotEmptyHash(t *testing.T) {
	absent := absentBase()
	empty := regularBase(hashBytes([]byte("")), 0o644)

	if !absent.Absent {
		t.Fatal("absentBase().Absent = false, want true")
	}
	if absent.SHA256 != "" {
		t.Fatalf("absentBase().SHA256 = %q, want empty", absent.SHA256)
	}
	if empty.Absent {
		t.Fatal("empty-file base marked absent")
	}
	if empty.SHA256 == "" {
		t.Fatal("empty-file base has no hash")
	}
	if err := absent.validate(); err != nil {
		t.Fatalf("absent base invalid: %v", err)
	}
	if err := empty.validate(); err != nil {
		t.Fatalf("empty base invalid: %v", err)
	}
}

// TestDecodeManifestRejectsWrongVersion ensures an unknown schema version is
// refused rather than silently accepted.
func TestDecodeManifestRejectsWrongVersion(t *testing.T) {
	raw := `{"version":2,"txn":"x","entries":[]}`
	_, err := decodeManifest([]byte(raw))
	if err == nil || !strings.Contains(err.Error(), "version") {
		t.Fatalf("decodeManifest(v2) err = %v, want version error", err)
	}
}

// TestHashBytesIsLowercaseHexSHA256 pins the hashing contract the manifest and
// preimage blobs share.
func TestHashBytesIsLowercaseHexSHA256(t *testing.T) {
	// Known SHA-256 of the empty string.
	const emptySHA = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if got := hashBytes(nil); got != emptySHA {
		t.Fatalf("hashBytes(empty) = %q, want %q", got, emptySHA)
	}
	if got := hashBytes([]byte("abc")); got != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
		t.Fatalf("hashBytes(abc) = %q", got)
	}
}

// Ensure fs is referenced (mode type documentation) without an unused import if
// the file evolves; a no-op assertion keeps the intent explicit.
var _ fs.FileMode = manifestEntry{}.Mode

func TestBaseValidateExplicitErrors(t *testing.T) {
	both := baseRecord{Absent: true, SHA256: "h"}
	if err := both.validate(); err == nil || !errors.Is(err, errBadBaseRecord) {
		t.Fatalf("both.validate() = %v, want errBadBaseRecord", err)
	}
	neither := baseRecord{}
	if err := neither.validate(); err == nil || !errors.Is(err, errBadBaseRecord) {
		t.Fatalf("neither.validate() = %v, want errBadBaseRecord", err)
	}
}
