# Platform binary selection and checksum verification

This document describes the on-disk contract implemented for
[ADR-002](../../design/adr/adr-002-platform-binary-selection.md): how the plugin
ships per-platform `llm-wiki` binaries, how exactly one is selected, and how its
integrity is verified before execution.

> **Scope.** ADR-002 owns *shipping + selecting + integrity-verifying* the
> binary. Install / init / upgrade / uninstall / asset-ownership is **ADR-009**
> and is deliberately not implemented here. Verification is an **integrity**
> check (corruption / wrong-platform / mismatched packaging), **not** an
> authenticity check — signing / provenance attestation is deferred and the
> residual tampered-binary risk stays open in
> [`knowledge/risks.md`](../../knowledge/risks.md).

## Supported platforms

Selection covers exactly five targets, keyed `<goos>_<goarch>`:

| Key | OS / arch |
| --- | --------- |
| `darwin_arm64` | macOS, Apple Silicon |
| `darwin_amd64` | macOS, Intel |
| `linux_arm64` | Linux, arm64 |
| `linux_amd64` | Linux, x86-64 |
| `windows_amd64` | Windows, x86-64 |

Detection is a single path (`platform.Detect`, over `runtime.GOOS/GOARCH`). Any
other OS/arch is rejected as an unsupported platform; there is no fallback, which
is what keeps the five-platform matrix deterministic.

## On-disk layout

Binaries ship under `bin/`, one directory per platform key, alongside a single
checksum manifest:

```
bin/
  SHA256SUMS
  darwin_arm64/llm-wiki
  darwin_amd64/llm-wiki
  linux_arm64/llm-wiki
  linux_amd64/llm-wiki
  windows_amd64/llm-wiki.exe
```

The Windows artifact carries the `.exe` suffix; all others have none. Paths are
always slash-separated so a manifest generated on one OS verifies on another.

## Checksum manifest (`SHA256SUMS`)

The manifest is a GNU coreutils `sha256sum`-style text file: one entry per line,

```
<64-hex-sha256><space><type><bundle-relative-path>
```

where `<type>` is a space (text mode) or `*` (binary mode). Digests are
lowercase hex. Entries are written sorted by path so the file is deterministic
and diffable. It is compatible with `sha256sum -c` / `shasum -a 256 -c`.

Generate it locally with:

```bash
go run ./cmd/gen-checksums -root <bundle-root>   # writes <root>/bin/SHA256SUMS
```

The full multi-platform release pipeline that builds the five binaries and runs
the generator in CI is deferred to a later infra issue.

## Verifying: `llm-wiki selfcheck`

`selfcheck` detects the running platform, loads `bin/SHA256SUMS`, and verifies
this platform's shipped binary against it:

```bash
llm-wiki selfcheck                # verify the bundle next to the executable
llm-wiki selfcheck --root <dir>   # verify a bundle at an explicit root
llm-wiki selfcheck --json         # emit the versioned contract envelope
```

The bundle root defaults to the directory containing the running executable and
can be overridden with `--root <dir>` (or `--root=<dir>`).

It **fails closed**: a tampered, corrupt, wrong-platform, missing, or unlisted
artifact — or a missing/unparseable manifest — is rejected before execution with
the contract's `system-or-filesystem-failure` status and exit code `5` (see
[exit-codes.md](exit-codes.md)). On success it reports `success` / exit `0`.

The outcome is carried by the envelope `status`, not a validation `finding`:
findings are OKF/profile validation results (ADR-004), whereas an integrity
failure is a system/filesystem condition. Verification requires **no network
access**. In a source checkout with no built binaries under `bin/`, `selfcheck`
correctly reports `system-or-filesystem-failure` — that is expected fail-closed
behaviour, not a defect.
