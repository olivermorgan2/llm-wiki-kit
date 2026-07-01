package platform

import (
	"errors"
	"fmt"
	"io/fs"
)

// ManifestName is the checksum file that ships in bin/ alongside the per-
// platform binaries. Its bundle-relative location is "bin/" + ManifestName.
const ManifestName = "SHA256SUMS"

// The three integrity-failure modes ADR-002 requires the gate to catch. All of
// them are terminal: the caller must refuse to execute the selected binary.
var (
	// ErrManifestEntryMissing: the selected platform's artifact has no entry in
	// the manifest, so its integrity cannot be established.
	ErrManifestEntryMissing = errors.New("artifact not listed in manifest")
	// ErrArtifactMissing: the selected binary is absent from the bundle.
	ErrArtifactMissing = errors.New("artifact file missing")
	// ErrChecksumMismatch: the selected binary's SHA-256 does not match the
	// manifest (corruption, wrong-platform packaging, or tampering).
	ErrChecksumMismatch = errors.New("artifact checksum mismatch")
)

// Verify checks that the binary selected for platform p exists in fsys and that
// its SHA-256 matches the digest the manifest records for it.
//
// It fails closed: a missing manifest entry, a missing artifact, or any digest
// mismatch returns a non-nil error and the caller must refuse to execute the
// binary. Verify performs no network access — everything it needs is fsys and
// the already-parsed manifest.
func Verify(fsys fs.FS, p Platform, m Manifest) error {
	path := p.ArtifactPath()

	want, ok := m[path]
	if !ok {
		return fmt.Errorf("%w: %s", ErrManifestEntryMissing, path)
	}

	f, err := fsys.Open(path)
	if err != nil {
		return fmt.Errorf("%w: %s: %v", ErrArtifactMissing, path, err)
	}
	defer f.Close()

	got, err := Sum(f)
	if err != nil {
		return fmt.Errorf("reading artifact %s: %w", path, err)
	}
	if got != want {
		return fmt.Errorf("%w: %s: have %s, want %s", ErrChecksumMismatch, path, got, want)
	}
	return nil
}

// VerifyBundle is the single call a launcher or the selfcheck command makes: it
// detects the running platform, loads bin/SHA256SUMS from fsys, and verifies
// this platform's artifact against it. On success it returns the verified
// Platform. Like Verify it fails closed and performs no network access.
func VerifyBundle(fsys fs.FS) (Platform, error) {
	p, err := Detect()
	if err != nil {
		return p, err
	}
	name := "bin/" + ManifestName
	f, err := fsys.Open(name)
	if err != nil {
		return p, fmt.Errorf("open manifest %s: %w", name, err)
	}
	defer f.Close()
	m, err := ParseManifest(f)
	if err != nil {
		return p, fmt.Errorf("parse manifest %s: %w", name, err)
	}
	return p, Verify(fsys, p, m)
}
