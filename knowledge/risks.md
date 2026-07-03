# Risk register — llm-wiki-kit

Live risks with mitigation and status. Seeded from PRD §16
([`design/prd.md`](../design/prd.md)) plus cross-cutting risks called out
elsewhere in the PRD. Status: `open` (mitigation planned, not yet built),
`mitigated-by-design` (addressed in PRD/architecture), or `closed`.

## Product / design risks (PRD §16)

| Risk | Mitigation | Status |
|---|---|---|
| Profile rules mistaken for OKF rules | Report OKF and profile conformance separately; never present a profile rule as universal OKF. Bounded by [ADR-004](../design/adr/adr-004-validation-and-severity-model.md) (accepted). | open |
| Schema becomes rigid too early | Keep `core` conservative; prefer recommendations before requirements. | open |
| Hooks treated as a security boundary | Position CI as authoritative; document hook limitations; post-write hooks are feedback, not prevention. | open |
| Model overwrites human work | Preview existing-page changes; preserve unknown/unrelated content. Validation applies to human edits too — bounded by [ADR-004](../design/adr/adr-004-validation-and-severity-model.md) (accepted); safe-write gate [ADR-005](../design/adr/adr-005-safe-filesystem-layer.md) (accepted). | open |
| Sourced claims lose provenance | Require resolvable citations; preserve existing citations. | open |
| Custom-profile complexity becomes a programming language | Declarative data only; one inheritance level; explicit validation. | open |
| Cross-platform binaries drift | Automate release builds, checksums, and platform integration tests. Bounded by [ADR-002](../design/adr/adr-002-platform-binary-selection.md) (accepted). | open |
| Compilation drops important research facts | Track evidence gaps; keep sources addressable; deeper evaluation later. | open |
| Upgrades overwrite customization | Track ownership; preview repository changes; preserve local profile extensions. Ownership model (plugin-owned vs repo-owned + version record) decided in [ADR-009](../design/adr/adr-009-install-upgrade-uninstall-ownership.md) (proposed); upgrade/uninstall preservation implemented Phase 7. | open |
| Skills bypass deterministic safety | Route managed mutations through staged engine plans; narrow tool access. Mandatory engine write-gate in [ADR-005](../design/adr/adr-005-safe-filesystem-layer.md) (accepted); staged inspect/plan/apply model in [ADR-006](../design/adr/adr-006-staged-mutation-transaction-model.md) (proposed). | open |
| Skill adapters duplicate core logic | All policy/mutation in shared engine; test adapters as integrations only. | open |
| A reviewed diff becomes stale before application | Bind plans to source/target hashes; reject stale plans. Referenced by [ADR-003](../design/adr/adr-003-json-contract-and-exit-codes.md) (accepted); hash-binding mechanism owned by [ADR-006](../design/adr/adr-006-staged-mutation-transaction-model.md) (proposed). | open |
| YAML round-trip drops unknown fields (comments best-effort) | Acceptance criterion 6 (binding) requires unknown **frontmatter field** preservation; comment preservation is a best-effort, non-gated design-quality bar. [ADR-001](../design/adr/adr-001-go-toolchain-and-yaml.md) (accepted) adopts a node-aware YAML library (`goccy/go-yaml`) over archived `yaml.v3`, preserving unknown fields (gated) and comments (best-effort); the unknown-field round-trip fixture gates Phase 3. | open |

## Cross-cutting risks (from elsewhere in the PRD)

| Risk | Mitigation | Status |
|---|---|---|
| Supply chain / tampered binary | Versioned, checksum-verified release artifacts; verify checksums before executing bundled CLI (PRD §9, §14 Security). Checksum-verify-before-exec is a decided chokepoint in [ADR-002](../design/adr/adr-002-platform-binary-selection.md) (accepted). **Residual risk:** the checksum manifest ships inside the same payload, so verification catches wrong-platform/mismatched/corrupt binaries but **not** a maliciously rebuilt payload (attacker can recompute the manifest). The trust root that would close this — release **signing / provenance attestation** — is explicitly **deferred** by ADR-002; [ADR-009](../design/adr/adr-009-install-upgrade-uninstall-ownership.md) (proposed) confirms signing/provenance is **out of its scope** and re-defers it to a dedicated supply-chain ADR; risk stays open. | open |
| Path traversal / symlink escape | Engine canonicalizes & validates paths; resolves symlinks; rejects escapes; bounded write scope (FR10, §14). Single mandatory write-gate in [ADR-005](../design/adr/adr-005-safe-filesystem-layer.md) (accepted) — criterion 17 hard gate. | open |
| Untrusted source material as instructions | Treat imported content as data, never as agent instructions; repo content cannot override plugin/user instructions (FR10, §14 Security). Inputs-as-data at the filesystem boundary **contributes to** this mitigation via the write-gate in [ADR-005](../design/adr/adr-005-safe-filesystem-layer.md) (accepted); the full instruction-injection boundary (skill prompting, adapter/model instruction handling) is not owned by ADR-005 alone. | open |
| Stale mutation plan applied after target changed | `apply` rechecks hashes and refuses stale/modified plans without mutating (FR7, FR11). Referenced by [ADR-003](../design/adr/adr-003-json-contract-and-exit-codes.md) (accepted); mechanism owned by [ADR-006](../design/adr/adr-006-staged-mutation-transaction-model.md) (proposed). | open |
| Partial multi-file writes on interruption | Atomic writes; no partial updates across multi-file ops; safe recovery from interruption (FR1, FR10, §14 Reliability). **Per-file** atomic write+rename and the `.llm-wiki/` staging area are provided by [ADR-005](../design/adr/adr-005-safe-filesystem-layer.md) (accepted); the **cross-file** all-or-nothing transaction model (staging manifest, commit ordering, recovery/rollback, partial-commit detection) that makes a multi-file batch atomic is owned by [ADR-006](../design/adr/adr-006-staged-mutation-transaction-model.md) (proposed). ADR-005 alone gives the necessary primitive, not full multi-file interruption safety. | open |
| Divergent findings across skills/hooks/CI/CLI | One deterministic implementation shared by all four; acceptance criterion requires identical findings for identical state (§17). Shared JSON envelope [ADR-003](../design/adr/adr-003-json-contract-and-exit-codes.md) + single validation engine [ADR-004](../design/adr/adr-004-validation-and-severity-model.md) (both proposed) — criterion 15. | open |

## Process risk (this collaboration)

| Risk | Mitigation | Status |
|---|---|---|
| Normalized PRD drifts from source intent | Adversarial Codex review of `design/prd-normalized.md`; findings captured as `design/prd-addenda/` + decisions log before MVP scoping. | mitigated — review ran 2026-06-30 (`NEEDS_REVISION`); findings landed as addenda 001–005. |

## PRD review findings (Codex 2026-06-30)

Risks surfaced by the adversarial review
([archive](reviews/2026-06-30-codex-prd-review.md)). Each is addressed by a
PRD addendum but stays `open` because the underlying engineering work is
not built yet — the addendum bounds the risk, it does not retire it.

| Risk | Mitigation | Status |
|---|---|---|
| Canonical normalized PRD omits acceptance criteria | `prd-normalizer`'s 11-field form dropped `design/prd.md` §17; `/prd-to-mvp` reading only the normalized file would miss/dilute the ship gate. Carried §17 forward into the canonical input via [addendum 001](../design/prd-addenda/001-mvp-acceptance-criteria.md); dogfood note filed for upstream. | open |
| Academic-research profile underspecified | Initial domain value proposition had prose only and open templates (Q4); unplannable/untestable. Minimum MVP contract (page types, fields, sections, per-type valid/invalid fixtures) fixed in [addendum 003](../design/prd-addenda/003-academic-research-profile-contract.md). | open |
| MVP surface too wide for first release | Wide MVP (packaging, 5 binaries, install/upgrade/uninstall, custom profiles, hooks, CI, authoring, enrichment, staged mutation, indexes) risks an unshippable flat backlog. Slice order separates the must-pass spine (Slices 0–3) from in-MVP hardening (4–6) in [addendum 004](../design/prd-addenda/004-mvp-slice-order-and-fixture-plan.md). | open |
| Planning assumptions silently treated as final decisions | Q3/Q6/Q7/Q8 assumption-locked, not decided. Each MVP issue built on an assumption must name it ([addendum 002](../design/prd-addenda/002-mvp-planning-assumptions.md)) so an override is traceable to the issues it touches. | open |
