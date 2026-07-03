# Installing the kit: distribution flow and version record

This document describes the on-disk contract implemented for
[ADR-009](../../design/adr/adr-009-install-upgrade-uninstall-ownership.md): how
the plugin installs its bundle into a repository without ever losing or silently
overwriting user content, and the version-record manifest install writes to
record what it owns.

> **Scope.** This is the **install half** of ADR-009. **Upgrade and uninstall**
> are decided in ADR-009 but implemented in **Phase 7** and are not built here.
> The multi-platform **release-build shim** that produces the shipped binaries
> (issue #17) is still **pending**, so today's install is exercised from a source
> checkout. Release **signing / provenance attestation** — the trust root against
> a maliciously rebuilt payload — is deferred to a dedicated supply-chain ADR
> (mirroring ADR-002); the residual tampered-binary risk stays open in
> [`knowledge/risks.md`](../../knowledge/risks.md).

## Distribution flow

The kit is distributed as a Claude Code plugin package. End to end:

1. **Package.** The plugin ships the per-platform engine binaries and a single
   checksum manifest, laid out exactly as
   [platform-selection.md](platform-selection.md) specifies:

   ```
   bin/
     SHA256SUMS
     darwin_arm64/llm-wiki
     darwin_amd64/llm-wiki
     linux_arm64/llm-wiki
     linux_amd64/llm-wiki
     windows_amd64/llm-wiki.exe
   ```

2. **Select + verify (ADR-002).** `platform.Detect` maps the running OS/arch to
   exactly one shipped binary, and `llm-wiki selfcheck` verifies that binary
   against `bin/SHA256SUMS` before it is trusted. Both are **fail-closed**: an
   unsupported platform or a failed integrity check stops the flow.

3. **Install (ADR-009).** `llm-wiki install` writes the wiki bundle **and** a
   version-record manifest into the target repository through one
   [ADR-006](../../design/adr/adr-006-staged-mutation-transaction-model.md)
   transaction. As a fail-closed precondition install re-runs `platform.Detect`,
   so an unsupported platform is rejected (`system-or-filesystem-failure`, exit
   `5`) before any planning.

## Installing: `llm-wiki install`

```bash
llm-wiki install                 # install into the current directory
llm-wiki install <dir>           # install into an explicit existing directory
llm-wiki install --profile core  # select a profile (default: core)
llm-wiki install --dry-run       # print the full planned write set; mutate nothing
llm-wiki install --force         # overwrite pre-existing planned targets
llm-wiki install --json          # emit the versioned contract envelope
```

The target defaults to the current directory and **must already exist** —
install never creates the boundary it writes into (the transaction and the
ADR-005 write-gate require an existing root). It installs into both **new** and
**non-empty** repositories.

The complete write set is four boundary-relative targets, committed in one
all-or-nothing transaction (sorted; the manifest sorts first on its leading dot):

```
.llm-wiki/manifest.json
llm-wiki.yaml
wiki/index.md
wiki/templates/page-template.md
```

Because the whole set commits through the ADR-006 transaction, a non-empty-repo
install either lands entirely or not at all, and no pre-existing repository file
is touched — satisfying acceptance criterion 3 (install into a non-empty repo
loses no file).

### Silent-overwrite refusal

Install **never silently overwrites**. Before committing it lstats every planned
target; any that already exists — a user's own `llm-wiki.yaml`, a prior bundle,
or an existing `.llm-wiki/manifest.json` (which means *already installed*) — is a
**conflict**. On a conflict without `--force`, install mutates nothing and
reports `approval-required` (exit `3`), listing the conflicting paths in the
envelope's `approval` block. `--force` is the approval grant: it overwrites the
conflicting targets and rewrites the manifest. This reuses `init`'s exit-3
approval envelope, so refusal semantics are identical across both commands.

### `--dry-run`

`--dry-run` computes the identical plan, manifest, and conflict set, then
**mutates nothing** — it returns before the transaction stages anything, so not
even `.llm-wiki/` is created. It mirrors the real outcome exactly: a conflict
without `--force` produces the same exit-3 refusal; otherwise it reports
`success` (exit `0`) with the full planned write set in `affectedPaths`. The JSON
envelope carries the same six ADR-003 fields as a real install — dry-run is
distinguished only by the human-readable wording (`would create …`), never by an
extra field.

### Exit codes

| Case | Status | Exit |
| --- | --- | --- |
| Bad flag, extra argument, unknown profile, or missing/non-directory target | `invalid-invocation` | `4` |
| Unsupported platform | `system-or-filesystem-failure` | `5` |
| Conflict without `--force` (real **or** `--dry-run`) | `approval-required` | `3` |
| Clean install, `--force`, or any `--dry-run` that would succeed | `success` | `0` |
| Transaction / filesystem error at commit (rolled back, tree unchanged) | `system-or-filesystem-failure` | `5` |

See [exit-codes.md](exit-codes.md) for the canonical table.

## Version record: `.llm-wiki/manifest.json`

Install writes a single JSON version-record inside the engine-managed
`.llm-wiki/` tree — never at the repo root and never inside user content. It is a
record of the assets install **manages**, plus the versions in effect, so Phase 7
upgrade/uninstall can act within a recorded ownership class instead of guessing.

### Schema

| Field | Meaning |
| --- | --- |
| `schemaVersion` | Manifest schema version (currently `"1"`), independent of the versions below. |
| `plugin` | Plugin version at install time. |
| `cli` | Engine/CLI version at install time. (One binary today; distinct fields for the future.) |
| `okf` | OKF target version, single-sourced from `scaffold.OKFVersion` — the same value written into `llm-wiki.yaml`'s `okfVersion`, so the two never drift. |
| `profile` | The ADR-007 profile reference: `{ id, version }`. |
| `assets[]` | One entry per install-managed asset: `{ path, class, hash, lastInstalledHash }`. |

Each asset records its ownership `class` — **`plugin-owned`** (written by the
plugin; may later be replaced or removed) or **`repo-owned`** (existing repo
content install only references). Install writes only **plugin-owned** assets;
`repo-owned` is defined now for Phase 7. `hash` is the current on-disk content
hash and `lastInstalledHash` is the hash the plugin last wrote; both are the
lowercase-hex SHA-256 of the committed bytes and are **equal at install time**.

Hashes are computed from the same in-memory change set the transaction commits,
so a recorded hash always matches the bytes on disk even though two scaffold
pages embed the install date.

### Self-exclusion

The manifest is itself **plugin-owned** engine metadata, but it **does not list
itself** as an asset: recording its own hash is a fixed point — writing the hash
into the manifest changes the manifest, invalidating the hash. It is still fully
protected: as a planned target it is covered by the silent-overwrite refusal (a
pre-existing non-manifest file at that path is a conflict), and it is plugin-owned
by construction. This exclusion is documented in the `internal/manifest` package
doc.

### The user-modification signal (Phase 7)

`hash` diverging from `lastInstalledHash` for a plugin-owned asset is the
**user-modification signal**: it means someone edited a file the plugin owns.
Phase 7 upgrade/uninstall use this to *skip and report* user-modified
plugin-owned files (preserving the edit) rather than clobbering them, overwriting
only under an explicit `--force`. Install writes the two hashes equal; nothing in
Phase 2 consumes their divergence yet.
