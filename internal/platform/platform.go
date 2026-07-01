// Package platform implements the single OS+arch detection path and the
// integrity gate that ADR-002 requires: it maps the running platform to exactly
// one shipped llm-wiki binary and verifies that binary's bundled SHA-256
// checksum before it would be executed.
//
// Verification is an *integrity* check (accidental corruption / wrong-platform
// / mismatched packaging), not an *authenticity* check against a maliciously
// rebuilt payload. Cryptographic signing / provenance attestation is explicitly
// deferred (ADR-009 or a dedicated supply-chain ADR); the residual tampered-
// binary risk is tracked in knowledge/risks.md.
package platform

import (
	"errors"
	"fmt"
	"runtime"
)

// ErrUnsupportedPlatform is returned when the running OS/arch is not one of the
// five platforms ADR-002 ships a binary for.
var ErrUnsupportedPlatform = errors.New("unsupported platform")

// binaryName is the base name of the shipped engine binary. On Windows the
// artifact carries the .exe suffix (see Platform.ArtifactPath).
const binaryName = "llm-wiki"

// Platform identifies one of the five supported OS/arch targets.
type Platform struct {
	OS   string // GOOS, e.g. "darwin"
	Arch string // GOARCH, e.g. "arm64"
	Key  string // canonical "<os>_<arch>" selection key
}

// supported is the closed set of platforms ADR-002 ships. It is the single
// source of truth for the five-platform matrix; membership is what makes
// selection deterministic.
var supported = map[string]bool{
	"darwin_arm64":  true,
	"darwin_amd64":  true,
	"linux_arm64":   true,
	"linux_amd64":   true,
	"windows_amd64": true,
}

// Detect resolves the running platform via the Go runtime. It is the one
// detection path the engine uses; everything else flows from its result.
func Detect() (Platform, error) {
	return detect(runtime.GOOS, runtime.GOARCH)
}

// detect is the testable core of Detect: it maps an explicit (goos, goarch)
// pair to a supported Platform or returns ErrUnsupportedPlatform. Keeping the
// runtime lookup out of this function is what lets the full matrix be tested on
// any single host.
func detect(goos, goarch string) (Platform, error) {
	key := goos + "_" + goarch
	if !supported[key] {
		return Platform{}, fmt.Errorf("%w: %s", ErrUnsupportedPlatform, key)
	}
	return Platform{OS: goos, Arch: goarch, Key: key}, nil
}

// ArtifactPath is the slash-separated, bundle-relative location of this
// platform's shipped binary: bin/<key>/llm-wiki[.exe]. It is always slash-
// separated so it doubles as both a filesystem path (via io/fs) and a manifest
// key across operating systems.
func (p Platform) ArtifactPath() string {
	name := binaryName
	if p.OS == "windows" {
		name += ".exe"
	}
	return "bin/" + p.Key + "/" + name
}
