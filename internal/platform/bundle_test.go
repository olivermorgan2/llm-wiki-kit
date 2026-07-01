package platform

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"testing/fstest"
)

func hostBundle(t *testing.T, content []byte, digest string) (Platform, fstest.MapFS) {
	t.Helper()
	host, err := Detect()
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	var buf bytes.Buffer
	if err := WriteManifest(&buf, Manifest{host.ArtifactPath(): digest}); err != nil {
		t.Fatalf("WriteManifest: %v", err)
	}
	fsys := fstest.MapFS{
		host.ArtifactPath():   {Data: content},
		"bin/" + ManifestName: {Data: buf.Bytes()},
	}
	return host, fsys
}

func TestVerifyBundleSuccess(t *testing.T) {
	content := []byte("engine payload")
	sum, err := Sum(bytes.NewReader(content))
	if err != nil {
		t.Fatalf("Sum: %v", err)
	}
	host, fsys := hostBundle(t, content, sum)

	p, err := VerifyBundle(fsys)
	if err != nil {
		t.Fatalf("VerifyBundle: %v", err)
	}
	if p.Key != host.Key {
		t.Errorf("returned platform %q, want %q", p.Key, host.Key)
	}
}

func TestVerifyBundleMissingManifest(t *testing.T) {
	host, err := Detect()
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	fsys := fstest.MapFS{host.ArtifactPath(): {Data: []byte("engine")}}

	if _, err := VerifyBundle(fsys); err == nil {
		t.Fatal("expected an error when bin/SHA256SUMS is absent")
	}
}

func TestVerifyBundlePropagatesMismatch(t *testing.T) {
	_, fsys := hostBundle(t, []byte("engine payload"), strings.Repeat("a", 64))

	if _, err := VerifyBundle(fsys); !errors.Is(err, ErrChecksumMismatch) {
		t.Fatalf("err = %v, want ErrChecksumMismatch", err)
	}
}
