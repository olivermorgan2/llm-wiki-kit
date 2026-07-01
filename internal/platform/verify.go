package platform

import (
	"errors"
	"fmt"
	"io/fs"
)

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
