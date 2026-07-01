# Evaluation — Issue #2: OS/arch detection + release-artifact checksum verification

**Branch:** `issue-002-platform-binary-selection` (cut from synced `main`).
**ADR:** [ADR-002](../design/adr/adr-002-platform-binary-selection.md) (ship /
select / integrity-verify the platform binary). Envelope + exit-code buckets
reuse [ADR-003](../design/adr/adr-003-json-contract-and-exit-codes.md).
**Method:** strict TDD (red → green per unit), incremental commits.

## What changed

**Detection + integrity gate (`internal/platform/`)**
- `platform.go` — `Detect()` over a testable `detect(goos,goarch)` seam
  (`runtime.GOOS/GOARCH`); the closed five-target set
  (`darwin/linux` × `arm64/amd64`, `windows_amd64`) or `ErrUnsupportedPlatform`.
  `Platform.ArtifactPath()` → `bin/<key>/llm-wiki[.exe]` (slash-separated).
- `manifest.go` — `Sum` (lowercase-hex SHA-256), `ParseManifest` /
  `WriteManifest` for the GNU `sha256sum`-style `SHA256SUMS` format
  (text + binary mode, malformed lines hard-fail, deterministic sorted output).
- `verify.go` — `Verify(fs.FS, Platform, Manifest)` fails closed with sentinel
  errors (`ErrManifestEntryMissing`, `ErrArtifactMissing`, `ErrChecksumMismatch`);
  `VerifyBundle(fs.FS)` = detect → load `bin/SHA256SUMS` → verify. `ManifestName`.

**Generator (`cmd/gen-checksums/`)**
- `main.go` — walks `<root>/bin`, checksums each binary via `platform.Sum`,
  writes `bin/SHA256SUMS` via `platform.WriteManifest`. Excludes the manifest
  itself; fails rather than writing an empty manifest.

**CLI (`cmd/llm-wiki/`)**
- `cli.go` — new `selfcheck` subcommand + usage entry; `--root`/`--root=` bundle
  root override (default: the executable's directory). Any verification error →
  `system-or-filesystem-failure` (exit 5); success → exit 0.

**Docs**
- `docs/contract/platform-selection.md` — supported platforms, `bin/` layout,
  `SHA256SUMS` format, `selfcheck` usage, fail-closed semantics.

## Commits (in order)

- `ab38748` feat(platform): OS/arch detection + artifact path for five targets (ADR-002, #2)
- `f7d5831` feat(platform): SHA256SUMS manifest parse/format + shared Sum helper (ADR-002, #2)
- `5dad2e0` feat(platform): fail-closed checksum verification gate (ADR-002, #2)
- `7bf294e` feat(gen-checksums): local SHA256SUMS generator for the bin/ bundle (ADR-002, #2)
- `15470e8` feat(cli): add selfcheck command + VerifyBundle integrity gate (ADR-002, #2)

## Verification performed

- `gofmt -l` clean; `go vet ./...` OK; `go build ./...` OK (Go 1.24.13).
- `go test ./...` — all packages pass (issue-#2 adds 20 tests: 13 platform
  detection/manifest/verify/bundle + 5 CLI selfcheck + 2 generator).
- Manual smoke on a synthetic bundle: `gen-checksums` writes `SHA256SUMS`;
  `selfcheck` (human + `--json`) reports `success`/exit 0; after tampering,
  compiled binary returns `system-or-filesystem-failure`/exit 5; empty bundle
  (clean-checkout default) fails closed with exit 5.

## Acceptance criteria (issue #2)

- [x] Each supported OS/arch resolves to exactly one binary via one detection
      path (`detect` matrix test; unsupported rejected).
- [x] A tampered/corrupt/wrong-platform/missing artifact is rejected before
      execution with the `system-or-filesystem-failure` exit code (5).
- [x] Selection + verification covered by unit tests; no network access at
      selection time (`fs.FS` seam, `fstest.MapFS`).
- [x] On-disk layout + CI-generatable checksum manifest defined (`bin/<key>/`,
      `SHA256SUMS`, `gen-checksums`).

## Design note — outcome via status, not a finding

The plan sketched "one Finding on mismatch." Implementation instead carries the
selfcheck outcome on the envelope `status` with `findings: []`, matching the
issue-#1 convention (`runUnknown` emits no finding) and ADR-004's meaning of a
*finding* as an OKF/profile validation result — an integrity failure is a
system/filesystem condition, not a validation finding. This keeps the ADR-003
contract unchanged (no new `ruleset` value invented).

## Non-goals honored (ADR-002 vs ADR-009)

- No install / init / upgrade / uninstall / asset-ownership work (ADR-009).
- No exec/launch of the selected binary — `selfcheck` verifies and reports only.
- No signing / provenance / authenticity; residual tampered-binary risk stays
  open in `knowledge/risks.md` (already recorded — no edit required).
- No network at detect/verify time; the ADR-003 envelope/exit-code contract is
  unchanged.

## Follow-ups (out of scope for #2)

- **Release pipeline:** GitHub Actions workflow that builds the five binaries,
  runs `gen-checksums`, and bundles them (separate infra issue).
- **Launcher/exec** of the verified binary and full install/lifecycle → ADR-009.
- **Signing / provenance attestation** → ADR-009 or a dedicated supply-chain ADR.
- `design/architecture.md` does not exist yet; note the new `internal/platform`
  package for the next `/workflow-docs` run.

## Commands to reproduce

```bash
export PATH="$HOME/sdk/go1.24.13/bin:$PATH"   # or your own Go 1.24.13
cd /Users/hermes/llm-wiki-kit
gofmt -l internal cmd && go vet ./... && go build ./... && go test ./...

# end-to-end on a synthetic bundle
SC="$(mktemp -d)"; KEY="$(go env GOOS)_$(go env GOARCH)"
mkdir -p "$SC/bin/$KEY"; printf 'pretend binary' > "$SC/bin/$KEY/llm-wiki"
go run ./cmd/gen-checksums -root "$SC"
go run ./cmd/llm-wiki selfcheck --root "$SC" --json   # success, exit 0
printf 'tampered' > "$SC/bin/$KEY/llm-wiki"
go run ./cmd/llm-wiki selfcheck --root "$SC" --json   # system-or-filesystem-failure, exit 5
```
```

## Next step

`/pr-review-packager` (or `gh pr create`) to open a PR from this branch
referencing ADR-002 and `Closes #2`, then a Codex review loop before merge.
