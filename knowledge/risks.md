# Risk register — llm-wiki-kit

Live risks with mitigation and status. Seeded from PRD §16
([`design/prd.md`](../design/prd.md)) plus cross-cutting risks called out
elsewhere in the PRD. Status: `open` (mitigation planned, not yet built),
`mitigated-by-design` (addressed in PRD/architecture), or `closed`.

## Product / design risks (PRD §16)

| Risk | Mitigation | Status |
|---|---|---|
| Profile rules mistaken for OKF rules | Report OKF and profile conformance separately; never present a profile rule as universal OKF. | open |
| Schema becomes rigid too early | Keep `core` conservative; prefer recommendations before requirements. | open |
| Hooks treated as a security boundary | Position CI as authoritative; document hook limitations; post-write hooks are feedback, not prevention. | open |
| Model overwrites human work | Preview existing-page changes; preserve unknown/unrelated content. | open |
| Sourced claims lose provenance | Require resolvable citations; preserve existing citations. | open |
| Custom-profile complexity becomes a programming language | Declarative data only; one inheritance level; explicit validation. | open |
| Cross-platform binaries drift | Automate release builds, checksums, and platform integration tests. | open |
| Compilation drops important research facts | Track evidence gaps; keep sources addressable; deeper evaluation later. | open |
| Upgrades overwrite customization | Track ownership; preview repository changes; preserve local profile extensions. | open |
| Skills bypass deterministic safety | Route managed mutations through staged engine plans; narrow tool access. | open |
| Skill adapters duplicate core logic | All policy/mutation in shared engine; test adapters as integrations only. | open |
| A reviewed diff becomes stale before application | Bind plans to source/target hashes; reject stale plans. | open |
| YAML round-trip drops unknown fields/comments | Acceptance criterion 6 requires unknown-field preservation; planning assumption locks a node-aware YAML library (`goccy/go-yaml`, ADR-001 candidate) over archived `yaml.v3`; round-trip fixture gates Phase 3. | open |

## Cross-cutting risks (from elsewhere in the PRD)

| Risk | Mitigation | Status |
|---|---|---|
| Supply chain / tampered binary | Versioned, checksum-verified release artifacts; verify checksums before executing bundled CLI (PRD §9, §14 Security). | open |
| Path traversal / symlink escape | Engine canonicalizes & validates paths; resolves symlinks; rejects escapes; bounded write scope (FR10, §14). | open |
| Untrusted source material as instructions | Treat imported content as data, never as agent instructions; repo content cannot override plugin/user instructions (FR10, §14 Security). | open |
| Stale mutation plan applied after target changed | `apply` rechecks hashes and refuses stale/modified plans without mutating (FR7, FR11). | open |
| Partial multi-file writes on interruption | Atomic writes; no partial updates across multi-file ops; safe recovery from interruption (FR1, FR10, §14 Reliability). | open |
| Divergent findings across skills/hooks/CI/CLI | One deterministic implementation shared by all four; acceptance criterion requires identical findings for identical state (§17). | open |

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
