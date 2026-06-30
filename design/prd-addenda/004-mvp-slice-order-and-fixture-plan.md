# Addendum 004 — MVP slice order and fixture plan

- **Source finding:** Codex PRD review (2026-06-30), non-blocking finding
  (Medium). The MVP surface is wide — plugin packaging, five platform
  binaries, install/upgrade/uninstall, custom profiles, hooks, CI,
  authoring, enrichment, staged mutation, and indexes (`design/prd.md`
  §18). It is coherent but too broad to treat as one undifferentiated first
  slice.
- **Decision:** **Accept — define a slice order inside MVP.** This does
  **not** cut anything from MVP scope (§8 is unchanged). It separates the
  **must-pass first-release path** (the thin vertical slice that proves the
  architecture end to end) from **follow-on hardening still inside MVP**,
  so `/prd-to-mvp` and `/issue-planner` can sequence issues instead of
  filing one flat backlog.
- **Knowledge links:** `knowledge/risks.md` ("MVP surface too wide for
  first release").

## Slice order (all slices are inside MVP; order ≠ scope cut)

### Slice 0 — Engine + contract spine (must-pass, blocks everything)
The deterministic authority that every other slice calls. Without this the
"one implementation shared by skills/hooks/CI/CLI" invariant cannot hold.
- `llm-wiki` CLI skeleton; OS/arch detection; checksum verification.
- Versioned JSON contract shell: contract version, operation, status,
  findings, affected paths, approval (acceptance criterion 14).
- Documented stable exit codes (success / warnings / validation failure /
  approval required / invalid invocation / system failure).
- `validate` (core profile) + separate OKF-vs-profile reporting
  (criteria 5, 7, 8).
- Safe filesystem layer: bounded writes, symlink/path-traversal rejection,
  atomic writes (criterion 17).
- **Exit gate:** criteria 5, 7, 8, 14, 17 pass on core profile.

### Slice 1 — Install / init (must-pass first-release path)
- Install into new and non-empty Git repos without data loss; `--dry-run`;
  refuse silent overwrites; record versions (criterion 3).
- Init with **core** profile; correct checksum-verified binary per platform
  (criteria 2, 4).
- **Exit gate:** criteria 2, 3, 4 (core) pass on all five platforms.

### Slice 2 — Authoring + staged mutation (must-pass first-release path)
This is the thinnest end-to-end "create real value" path.
- `page inspect` / `page plan` / `page apply` with hash-bound,
  stale-rejecting plans (criteria 12, 13).
- Authoring skill: new-page write, validate, show diff; preserve unknown
  fields; provenance citation for sourced claims in fixtures; no duplicate
  on repeated unchanged input (criteria 6, 9, 10, 11).
- **Exit gate:** criteria 6, 9, 10, 11, 12, 13 pass.

### Slice 3 — Academic-research profile (must-pass first-release path)
The initial domain value proposition. Depends on addendum 003.
- Ship `academic-research` profile + the per-type acceptance fixtures from
  addendum 003.
- Init with academic-research; author each profiled page type.
- **Exit gate:** criterion 4 (academic-research) + addendum-003 fixtures
  pass.

### Slice 4 — Enrichment + index maintenance (follow-on hardening, in MVP)
- Enrichment skill on existing pages via the same staged plan/apply path;
  preview before apply; preserve citations/unknown fields.
- Deterministic index maintenance (no model calls; stable ordering;
  dry-run/diff).
- **Exit gate:** criterion 11 (enrichment) + index reliability (§14).

### Slice 5 — Hooks + CI (follow-on hardening, in MVP)
- Optional Claude Code + Git hooks invoking the same engine.
- One GitHub Actions CI workflow (per addendum 002 A1) that fails on
  configured errors; same findings across skills/hooks/CI/CLI.
- **Exit gate:** criteria 15, 18, 19 pass.

### Slice 6 — Custom profile + upgrade/uninstall/doctor (follow-on, in MVP)
- Custom-profile template + `profile create`/`profile validate`; local-file
  only (per addendum 005).
- Upgrade/uninstall preserving wiki content + local profile extensions;
  `doctor` diagnostics.
- **Exit gate:** criteria 4 (custom), 16, 20 pass.

### Cross-cutting — Cross-platform test matrix (closes MVP)
- Automated tests green on macOS arm64/x86-64, Linux arm64/x86-64,
  Windows x86-64 (criterion 21). Runs continuously from Slice 0; **MVP is
  not done until this is green on all five platforms.**

## Must-pass first-release path (the spine)

**Slices 0 → 1 → 2 → 3.** This is the minimum that proves: deterministic
engine + contract, install into a real repo, author a real domain page
through the staged workflow, and demonstrate the academic-research value
proposition. Slices 4–6 harden the rest of the MVP surface and may be
sequenced after the spine is green, but remain required for MVP completion.

## Fixture plan (acceptance fixtures by criterion)

| Fixture set | Proves criteria | Location |
|---|---|---|
| OKF-valid + OKF-invalid core pages | 5, 7 | `profiles/core/examples/` |
| Broken-link page at configured severity | 8 | engine test data |
| Unknown-field round-trip page | 6 | authoring tests |
| Sourced-claim citation fixtures | 9 | `profiles/academic-research/examples/` (addendum 003) |
| Duplicate-input idempotency case | 10 | authoring tests |
| Stale-plan rejection case | 13 | engine test data |
| Path-traversal / symlink-escape attempts | 17 | engine security tests |
| Per-platform smoke repo | 2, 21 | CI matrix |

Any slice that bounds coverage (e.g. defers a CI target) must say so in its
issues rather than silently narrowing the matrix.
