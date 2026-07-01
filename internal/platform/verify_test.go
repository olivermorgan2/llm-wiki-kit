package platform

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"testing/fstest"
)

func TestVerifySuccess(t *testing.T) {
	p, _ := detect("linux", "amd64")
	content := []byte("fake linux binary")
	sum, err := Sum(bytes.NewReader(content))
	if err != nil {
		t.Fatalf("Sum: %v", err)
	}
	fsys := fstest.MapFS{p.ArtifactPath(): {Data: content}}
	m := Manifest{p.ArtifactPath(): sum}

	if err := Verify(fsys, p, m); err != nil {
		t.Fatalf("Verify should accept a matching artifact: %v", err)
	}
}

func TestVerifyChecksumMismatch(t *testing.T) {
	p, _ := detect("linux", "amd64")
	fsys := fstest.MapFS{p.ArtifactPath(): {Data: []byte("tampered payload")}}
	m := Manifest{p.ArtifactPath(): strings.Repeat("a", 64)} // deliberately wrong

	if err := Verify(fsys, p, m); !errors.Is(err, ErrChecksumMismatch) {
		t.Fatalf("err = %v, want ErrChecksumMismatch", err)
	}
}

func TestVerifyMissingArtifact(t *testing.T) {
	p, _ := detect("linux", "amd64")
	fsys := fstest.MapFS{} // artifact absent
	m := Manifest{p.ArtifactPath(): strings.Repeat("a", 64)}

	if err := Verify(fsys, p, m); !errors.Is(err, ErrArtifactMissing) {
		t.Fatalf("err = %v, want ErrArtifactMissing", err)
	}
}

func TestVerifyMissingManifestEntry(t *testing.T) {
	p, _ := detect("linux", "amd64")
	fsys := fstest.MapFS{p.ArtifactPath(): {Data: []byte("x")}}
	m := Manifest{} // no entry for this platform

	if err := Verify(fsys, p, m); !errors.Is(err, ErrManifestEntryMissing) {
		t.Fatalf("err = %v, want ErrManifestEntryMissing", err)
	}
}
