// Package profiles embeds the shipped validation-profile data files so the
// standalone engine binary can resolve profiles offline (ADR-002 single
// self-contained binary; ADR-007 declarative data profiles). The profile YAML
// lives at the repository root under profiles/<id>/profile.yaml — the canonical
// location alongside each profile's examples/ fixtures — and `go:embed` cannot
// reach across `../`, so the embed directive lives here, in a package rooted at
// that directory, rather than in internal/profile.
//
// Only the profile.yaml rule files are embedded; the examples/ corpora are test
// fixtures read from disk (internal/validate/examples_fixtures_test.go), never
// shipped in the binary.
package profiles

import "embed"

// FS holds the shipped profile rule data. internal/profile reads
// "<id>/profile.yaml" from it; the id→path mapping is owned there.
//
//go:embed core/profile.yaml academic-research/profile.yaml
var FS embed.FS
