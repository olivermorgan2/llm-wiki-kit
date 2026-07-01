package platform

import (
	"errors"
	"runtime"
	"testing"
)

func TestDetectSupportedPlatforms(t *testing.T) {
	cases := []struct {
		goos, goarch string
		wantKey      string
	}{
		{"darwin", "arm64", "darwin_arm64"},
		{"darwin", "amd64", "darwin_amd64"},
		{"linux", "arm64", "linux_arm64"},
		{"linux", "amd64", "linux_amd64"},
		{"windows", "amd64", "windows_amd64"},
	}
	for _, c := range cases {
		p, err := detect(c.goos, c.goarch)
		if err != nil {
			t.Fatalf("detect(%q,%q) unexpected error: %v", c.goos, c.goarch, err)
		}
		if p.Key != c.wantKey {
			t.Errorf("detect(%q,%q).Key = %q, want %q", c.goos, c.goarch, p.Key, c.wantKey)
		}
		if p.OS != c.goos || p.Arch != c.goarch {
			t.Errorf("detect(%q,%q) = {OS:%q Arch:%q}, want OS/arch preserved", c.goos, c.goarch, p.OS, p.Arch)
		}
	}
}

func TestDetectRejectsUnsupportedPlatforms(t *testing.T) {
	cases := []struct{ goos, goarch string }{
		{"linux", "386"},     // unsupported arch
		{"windows", "arm64"}, // unsupported os/arch combo
		{"plan9", "amd64"},   // unsupported os
		{"freebsd", "amd64"}, // unsupported os
		{"darwin", "ppc64"},  // unsupported arch
	}
	for _, c := range cases {
		if _, err := detect(c.goos, c.goarch); !errors.Is(err, ErrUnsupportedPlatform) {
			t.Errorf("detect(%q,%q) err = %v, want ErrUnsupportedPlatform", c.goos, c.goarch, err)
		}
	}
}

// Detect resolves the host it is compiled for; the test host is always one of
// the five supported targets, so it must succeed and echo the runtime values.
func TestDetectUsesRuntime(t *testing.T) {
	p, err := Detect()
	if err != nil {
		t.Fatalf("Detect() on the test host should succeed: %v", err)
	}
	if p.OS != runtime.GOOS || p.Arch != runtime.GOARCH {
		t.Errorf("Detect() = {OS:%q Arch:%q}, want runtime {%q,%q}", p.OS, p.Arch, runtime.GOOS, runtime.GOARCH)
	}
}

func TestArtifactPath(t *testing.T) {
	cases := []struct {
		goos, goarch string
		want         string
	}{
		{"darwin", "arm64", "bin/darwin_arm64/llm-wiki"},
		{"linux", "amd64", "bin/linux_amd64/llm-wiki"},
		{"windows", "amd64", "bin/windows_amd64/llm-wiki.exe"},
	}
	for _, c := range cases {
		p, err := detect(c.goos, c.goarch)
		if err != nil {
			t.Fatalf("detect(%q,%q): %v", c.goos, c.goarch, err)
		}
		if got := p.ArtifactPath(); got != c.want {
			t.Errorf("ArtifactPath for %s = %q, want %q", p.Key, got, c.want)
		}
	}
}
