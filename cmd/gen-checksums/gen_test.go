package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/olivermorgan2/llm-wiki-kit/internal/platform"
)

func writeFile(t *testing.T, root, rel string, data []byte) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestBuildManifestCoversBinariesAndExcludesManifest(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "bin/linux_amd64/llm-wiki", []byte("linux payload"))
	writeFile(t, root, "bin/darwin_arm64/llm-wiki", []byte("darwin payload"))
	writeFile(t, root, "bin/SHA256SUMS", []byte("stale contents"))

	m, err := buildManifest(root)
	if err != nil {
		t.Fatalf("buildManifest: %v", err)
	}
	if _, ok := m["bin/SHA256SUMS"]; ok {
		t.Error("the manifest must not checksum itself")
	}
	if len(m) != 2 {
		t.Fatalf("len = %d, want 2 (%v)", len(m), m)
	}
	for _, want := range []string{"bin/linux_amd64/llm-wiki", "bin/darwin_arm64/llm-wiki"} {
		if _, ok := m[want]; !ok {
			t.Errorf("missing entry %q", want)
		}
	}
}

// The generated manifest, once serialized and reparsed, must satisfy the
// verification gate for the host's own artifact — closing the produce->consume
// loop end to end.
func TestBuildManifestRoundTripsThroughVerify(t *testing.T) {
	root := t.TempDir()
	host, err := platform.Detect()
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	writeFile(t, root, host.ArtifactPath(), []byte("host payload"))

	m, err := buildManifest(root)
	if err != nil {
		t.Fatalf("buildManifest: %v", err)
	}

	var buf bytes.Buffer
	if err := platform.WriteManifest(&buf, m); err != nil {
		t.Fatalf("WriteManifest: %v", err)
	}
	parsed, err := platform.ParseManifest(&buf)
	if err != nil {
		t.Fatalf("ParseManifest: %v", err)
	}
	if err := platform.Verify(os.DirFS(root), host, parsed); err != nil {
		t.Fatalf("Verify should accept the generated manifest: %v", err)
	}
}

// gen writes bin/SHA256SUMS under root and reports success.
func TestGenWritesManifestFile(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "bin/linux_amd64/llm-wiki", []byte("payload"))

	var out, errb bytes.Buffer
	if code := gen(root, &out, &errb); code != 0 {
		t.Fatalf("gen exit = %d, stderr = %s", code, errb.String())
	}
	data, err := os.ReadFile(filepath.Join(root, "bin", "SHA256SUMS"))
	if err != nil {
		t.Fatalf("read SHA256SUMS: %v", err)
	}
	m, err := platform.ParseManifest(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("parse written manifest: %v", err)
	}
	if _, ok := m["bin/linux_amd64/llm-wiki"]; !ok {
		t.Errorf("written manifest missing entry: %s", data)
	}
}

// With no bin/ tree, gen fails rather than writing an empty manifest.
func TestGenFailsWhenNoBinaries(t *testing.T) {
	root := t.TempDir()
	var out, errb bytes.Buffer
	if code := gen(root, &out, &errb); code == 0 {
		t.Errorf("gen should fail when there is nothing to checksum")
	}
}
