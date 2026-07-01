# llm-wiki-kit — Build-Out Plan

**Last updated:** 2026-06-30
**Granularity:** standard

> Sequences the MVP defined in [`design/mvp.md`](mvp.md). Reads the normalized
> PRD ([`design/prd-normalized.md`](prd-normalized.md)) **and** the accepted
> addenda [`001`–`005`](prd-addenda/), which are authoritative. The phase
> order is the slice order fixed in
> [addendum 004](prd-addenda/004-mvp-slice-order-and-fixture-plan.md); the
> binding ship gate is the 21 acceptance criteria in
> [addendum 001](prd-addenda/001-mvp-acceptance-criteria.md).

## Objective

Take the engine-centric OKF plugin from an empty repository to a complete,
cross-platform first release: a versioned Claude Code plugin bundling a
self-contained Go `llm-wiki` engine, skills, optional hooks, two shipped
profiles plus a local custom-profile scaffold, and a GitHub Actions
conformance workflow. By the end, all 21 functional acceptance criteria pass
and the automated test matrix is green on all five supported platforms. This
plan covers everything from repository/toolchain decisions through the MVP
ship gate; it does not implement product code (that is the executor's job).

## Build strategy

Seven phases map 1:1 to the addendum-004 slices. The "must-pass spine"
(Phases 1–4 / Slices 0–3) proves the architecture end to end before the
remaining MVP surface is hardened (Phases 5–7 / Slices 4–6). Across all
phases, the cross-platform test matrix runs continuously and is the final
gate that closes MVP.

1. **Settle load-bearing decisions** (toolchain, contract, packaging,
   validation, filesystem safety) as ADRs before code — the engine is the
   shared authority, so its shape must be decided first.
2. **Build the deterministic engine + contract spine** (Phase 1) so every
   other surface has one authority to call.
3. **Make it installable** (Phase 2) into real repos with the correct
   per-platform binary.
4. **Prove the value path end to end** (Phases 3–4): author and enrich a real
   domain page through the staged workflow under the academic-research
   profile.
5. **Harden the remaining surface** (Phases 5–7): enrichment + indexes, hooks
   + CI, custom profiles + lifecycle.
6. **Close on the cross-platform matrix**: tests green on all five platforms.

## Scope

- **In scope:** every in-scope capability listed in
  [`design/mvp.md`](mvp.md) — the full PRD §18 MVP surface, sequenced into
  Phases 1–7 plus the cross-platform gate.
- **Out of scope:** every item in the "Out of scope" / "Deferred to later"
  lists of [`design/mvp.md`](mvp.md) — registry/trust (→ Phase 3 Ecosystem),
  non-GitHub-Actions CI targets, Claude Code compatibility shims, pre-release
  contract compatibility, and richer profile templates beyond the
  addendum-003 minimum.
- **Assumptions** (each overridable; an override revises the affected issues,
  it does not silently persist):
  - **Toolchain — Go (MVP assumption, ADR-001 candidate, Q2).** Build the
    engine on **Go 1.24.x**, a conservative current-stable line at planning
    time with broad CI-runner support. The exact patch is pinned in ADR-001 at
    first-engine-issue time. Reversible: a later Go choice changes only the
    toolchain ADR and `go.mod`.
  - **Toolchain — YAML library (MVP assumption, ADR-001 candidate, Q2).** Use
    **`github.com/goccy/go-yaml`** for deterministic, node-aware YAML
    parsing/editing. Rationale: acceptance criterion 6 (binding) requires
    unknown-field **round-trip preservation**, which needs a node/AST-aware
    library; comment preservation is a best-effort, non-gated bonus the same
    library affords (see ADR-001). The formerly default `gopkg.in/yaml.v3` is
    archived /
    unmaintained as of 2025. Reversible: confined to the engine's YAML
    adapter behind an internal interface.
  - **Planning assumptions A1–A5** from
    [addendum 002](prd-addenda/002-mvp-planning-assumptions.md): GitHub
    Actions is the only MVP CI template (A1); a single Claude Code version
    floor, no shim (A2); one platform-binary-selection mechanism (A3); JSON
    contract starts at v1 with no pre-release back-compat (A4); working
    name/MIT license (A5). Every issue built on an assumption names it.
  - **Domain contract** is the addendum-003 minimum for academic-research;
    **custom profiles are local-file only** (addendum 005).
  - Oliver has tooling access (Go, `gh`, CI runners for five platforms) and at
    least one real research corpus to dry-run the academic-research profile.

## Success criteria

The plan is complete when a user can:

1. Install the plugin into a new or non-empty Git repo and run the correct
   checksum-verified `llm-wiki` binary with no runtime installed, losing no
   existing file.
2. Initialize and author under core, academic-research, or a local custom
   profile, with OKF and profile conformance reported separately.
3. Author and enrich pages through the staged inspect/plan/apply workflow —
   diff before write, unknown fields and human content preserved, sourced
   claims carrying resolvable citations, stale plans refused, no duplicate on
   unchanged input.
4. Get identical findings from skills, hooks, CI, and the direct CLI, with the
   GitHub Actions workflow failing on configured errors.
5. Upgrade and uninstall without losing wiki content or local profile
   extensions, with zero confirmed writes outside the configured boundary, on
   all five supported platforms.

## Repository structure

```text
llm-wiki-kit/                  # the plugin/engine source repo (what we build)
  cmd/llm-wiki/                # Go CLI entry point
  internal/                    # engine packages — the shared deterministic authority
    contract/                  #   versioned JSON envelope + stable exit codes
    validate/                  #   OKF-vs-profile validation, three severities
    profile/                   #   data-driven profiles, one inheritance level
    plan/                      #   inspect/plan/apply, hash-binding, stale rejection
    fsafe/                     #   atomic writes, path/symlink safety, bounded scope
    index/                     #   deterministic index maintenance (no model calls)
  bin/                         # per-platform checksum-verified llm-wiki binaries (shipped)
  .claude/                     # plugin manifest, skills (authoring/enrichment), hooks
  profiles/
    core/examples/{valid,invalid}/
    academic-research/examples/{valid,invalid}/   # addendum 003 fixtures
    custom-template/           # local custom-profile scaffold (addendum 005)
  .github/workflows/           # GitHub Actions conformance workflow (addendum 002 A1)
  testdata/                    # engine fixtures: broken-link, stale-plan, traversal/symlink
  design/                      # prd*, mvp, build-out-plan, adr/
  knowledge/                   # curated knowledge layer
```

## Phases

> Each phase is one delivery unit (one milestone for `issue-planner`). The
> **Acceptance criteria** line records which of the 21 addendum-001 criteria
> the phase's exit gate proves; the full mapping table follows the phases.

### Phase 1: Engine + contract spine (Slice 0 — must-pass, blocks everything)

- **Goal:** the deterministic authority every other surface calls — CLI
  skeleton, contract, core validation, filesystem safety.
- **Scope:**
  - `llm-wiki` CLI skeleton; OS/arch detection; checksum verification.
  - Versioned JSON contract shell (contract version, operation, status,
    findings, affected paths, approval) and documented stable exit codes.
  - `validate` on the core profile with OKF-vs-profile reporting at three
    severities; broken-link detection at configured severity.
  - Safe filesystem layer: bounded writes, atomic writes,
    symlink/path-traversal rejection.
- **ADR dependencies:** ADR-001 (Go + YAML), ADR-003 (JSON contract + exit
  codes), ADR-004 (validation/severity), ADR-005 (filesystem safety); ADR-002
  (binary selection) for OS/arch + checksum.
- **Deliverables:** runnable `llm-wiki validate`; contract schema doc + exit
  codes; `profiles/core/` with valid+invalid fixtures; security testdata
  (traversal/symlink); engine unit tests.
- **Exit criteria:** acceptance criteria **5, 7, 8, 14, 17** pass on the core
  profile.

### Phase 2: Install / init (Slice 1 — must-pass first-release path)

- **Goal:** install into a real repo and initialize a core-profile wiki with
  the correct per-platform binary.
- **Scope:**
  - Install into new and non-empty Git repos without data loss; `--dry-run`;
    refuse silent overwrites; record plugin/CLI/OKF/profile versions.
  - Init with the **core** profile; checksum-verified correct binary per
    platform (one selection mechanism, addendum 002 A3).
- **ADR dependencies:** ADR-002 (packaging + binary selection), ADR-009
  (install/lifecycle ownership model — install half), ADR-007 (profile system,
  core).
- **Deliverables:** install/init flow + `--dry-run`; version-record file;
  non-empty-repo and new-repo install fixtures.
- **Exit criteria:** acceptance criteria **2, 3, 4 (core)** pass on all five
  platforms; criterion **1** (install half of the documented distribution
  flow) demonstrated.

### Phase 3: Authoring + staged mutation (Slice 2 — must-pass first-release path)

- **Goal:** the thinnest end-to-end "create real value" path — author a new
  page through the staged, preview-before-write workflow.
- **Scope:**
  - `page inspect` / `page plan` / `page apply` with hash-bound,
    stale-rejecting plans.
  - Authoring skill (thin adapter): new-page write, validate, show diff;
    preserve unknown fields; provenance citations for sourced claims in
    fixtures; no duplicate page on repeated unchanged input.
- **ADR dependencies:** ADR-006 (staged mutation model), ADR-008 (provenance
  + citation resolution).
- **Deliverables:** `page` subcommands; authoring skill + adapter; unknown-
  field round-trip fixture; duplicate-input idempotency case; stale-plan
  rejection case.
- **Exit criteria:** acceptance criteria **6, 9, 10, 11, 12, 13** pass.

### Phase 4: Academic-research profile (Slice 3 — must-pass first-release path)

- **Goal:** the initial domain value proposition — author each profiled page
  type and validate it.
- **Scope:**
  - Ship `academic-research` (extends `core`, one level) with page types
    `source`, `claim`, `method`, `question`, `synthesis` per
    [addendum 003](prd-addenda/003-academic-research-profile-contract.md).
  - Per-type acceptance fixtures (≥1 valid + ≥1 invalid, each invalid
    targeting exactly one rule).
  - Init with academic-research; author each profiled page type.
- **ADR dependencies:** ADR-007 (profile system — academic-research
  realization).
- **Deliverables:** `profiles/academic-research/` profile + templates +
  `examples/{valid,invalid}/` covering the five added/tightened types.
- **Exit criteria:** acceptance criterion **4 (academic-research)** plus the
  addendum-003 fixtures pass. **Completes the must-pass spine.**

### Phase 5: Enrichment + index maintenance (Slice 4 — in-MVP hardening)

- **Goal:** extend the staged path to existing-page enrichment and keep
  indexes deterministically current.
- **Scope:**
  - Enrichment skill on existing pages via the same plan/apply path; preview
    before apply; preserve citations and unknown fields.
  - Deterministic index maintenance: no model calls, stable ordering,
    dry-run/diff.
- **ADR dependencies:** ADR-010 (index maintenance), ADR-006 (reused staged
  mutation).
- **Deliverables:** enrichment skill + adapter; index command with dry-run;
  enrichment preview fixtures; index-stability tests.
- **Exit criteria:** acceptance criterion **11 (enrichment)** plus index
  reliability (PRD §14: deterministic regeneration, no model call) hold.

### Phase 6: Hooks + CI (Slice 5 — in-MVP hardening)

- **Goal:** make the same engine the authority across hooks and CI, with CI as
  the team-authoritative gate.
- **Scope:**
  - Optional Claude Code + Git hooks invoking the same engine (post-write
    hooks documented as feedback, not prevention).
  - One GitHub Actions workflow (addendum 002 A1) that fails on configured
    errors; identical findings across skills/hooks/CI/CLI.
- **ADR dependencies:** ADR-011 (hooks + CI integration; Claude Code version
  floor, A2).
- **Deliverables:** optional hook scripts; `.github/workflows/` conformance
  workflow; cross-surface parity test (same state → same findings).
- **Exit criteria:** acceptance criteria **15, 18, 19** pass.

### Phase 7: Custom profile + upgrade/uninstall/doctor (Slice 6 — in-MVP hardening)

- **Goal:** local custom profiles and a safe kit lifecycle.
- **Scope:**
  - `profiles/custom-template/` + `profile create` / `profile validate`;
    `init --profile <id-or-path>` from a **local** profile (addendum 005).
  - Upgrade and uninstall that distinguish plugin-owned from repo-owned assets
    and preserve wiki content + local profile extensions; `doctor`
    diagnostics.
  - Renames / deletions / migrations require explicit approval.
- **ADR dependencies:** ADR-012 (custom-profile boundary), ADR-009
  (upgrade/uninstall ownership — lifecycle half).
- **Deliverables:** custom-template scaffold + `profile` subcommands;
  upgrade/uninstall flows with ownership tracking; `doctor`; approval-gated
  rename/delete fixtures.
- **Exit criteria:** acceptance criteria **4 (custom), 16, 20** pass; criterion
  **1** (update half of the documented distribution flow) demonstrated.

### Cross-cutting: Cross-platform test matrix (closes MVP)

- **Goal:** automated tests green on macOS arm64/x86-64, Linux arm64/x86-64,
  Windows x86-64.
- **Scope:** versioned, checksum-verified release artifacts; per-platform
  smoke repo; matrix runs continuously from Phase 1.
- **ADR dependencies:** ADR-002 (packaging/release), ADR-011 (CI matrix).
- **Deliverables:** release-build automation + checksums; CI matrix; per-
  platform smoke fixtures.
- **Exit criteria:** acceptance criteria **2 (per-platform), 21** green on all
  five platforms. **MVP is not done until this is green.**

## Acceptance-criteria → phase/milestone map

Every one of the 21 addendum-001 criteria maps to exactly one owning phase
(the phase whose exit gate proves it); criterion 1 and 4 span two phases as
noted. This table is the explicit milestone/task ↔ acceptance-criteria
mapping the ship gate requires.

| # | Acceptance criterion (abbrev.) | Owning phase | Milestone |
|---|---|---|---|
| 1 | Plugin installs **and updates** via documented flow | P2 (install) + P7 (update) | Foundation + Hardening |
| 2 | Correct checksum-verified CLI on every platform | P2 + Cross-cutting | Foundation + Gate |
| 3 | Install in new + non-empty repos, no file loss | P2 | Foundation |
| 4 | Init with core / academic-research / custom | P2 (core) + P4 (academic) + P7 (custom) | Foundation + Spine + Hardening |
| 5 | Separate OKF vs profile conformance | P1 | Foundation |
| 6 | Unknown frontmatter fields survive author/enrich | P3 | Spine |
| 7 | Malformed YAML / missing required field fails | P1 | Foundation |
| 8 | Broken links reported at configured severity | P1 | Foundation |
| 9 | Sourced model claims have resolvable citations | P3 | Spine |
| 10 | Repeated unchanged input → no duplicate page | P3 | Spine |
| 11 | Existing-page edits previewed before apply | P3 (author) + P5 (enrich) | Spine + Hardening |
| 12 | Managed writes go through staged plan/apply | P3 | Spine |
| 13 | Changed target → plan rejected, no mutation | P3 | Spine |
| 14 | Structured output conforms to contract + exit codes | P1 | Foundation |
| 15 | Same findings across skills/hooks/CI/CLI | P6 | Hardening |
| 16 | Renames/deletions/migrations need approval | P7 | Hardening |
| 17 | Pre-write guards reject traversal + symlink escape | P1 | Foundation |
| 18 | Post-write validation described as feedback | P6 | Hardening |
| 19 | CI workflow fails on configured errors | P6 | Hardening |
| 20 | Upgrade/uninstall preserve content + extensions | P7 | Hardening |
| 21 | Tests pass on all supported OS/arch | Cross-cutting | Gate |

## Milestone recommendation

Phases group into the project's macro-milestones (CLAUDE.md: Foundation, MVP,
Post-MVP). `issue-planner` may instead create one milestone per phase if finer
tracking is wanted; the grouping below is the recommended default.

| Milestone | Phases | Focus |
|---|---|---|
| Foundation | P1, P2 | Engine + contract spine; install/init with core profile |
| MVP-Spine | P3, P4 | Authoring + staged mutation; academic-research profile (must-pass path complete) |
| MVP-Hardening | P5, P6, P7 | Enrichment + indexes; hooks + CI; custom profile + lifecycle |
| Cross-platform gate | Cross-cutting | Green matrix on all five platforms — closes MVP |

## Initial issue backlog

> Issue-sized tasks. Each issue built on an assumption names it (e.g.
> "assumes GitHub Actions only — addendum 002 A1"). Turned into full issue
> bodies by `issue-planner` using `templates/issue-template.md`.

### Foundation — Phase 1 (Engine + contract spine)

- Scaffold `cmd/llm-wiki` + `internal/` engine packages on Go 1.24.x (ADR-001).
- Implement OS/arch detection + release-artifact checksum verification (ADR-002).
- Define the versioned JSON contract envelope + stable exit codes (ADR-003).
- Implement core-profile `validate` with separate OKF vs profile reporting and
  three severities (ADR-004).
- Implement broken-link detection at configured severity.
- Implement the safe filesystem layer: atomic writes, bounded scope,
  symlink/path-traversal rejection (ADR-005).
- Add `profiles/core/examples/{valid,invalid}` + security testdata.

### Foundation — Phase 2 (Install / init)

- Implement install into new and non-empty Git repos with `--dry-run` and
  silent-overwrite refusal (ADR-009).
- Implement per-platform binary selection (one mechanism — addendum 002 A3,
  ADR-002).
- Implement `init` with the core profile + version-record file (ADR-007).
- Add new-repo and non-empty-repo install fixtures.

### MVP-Spine — Phase 3 (Authoring + staged mutation)

- Implement `page inspect` / `page plan` / `page apply` with hash-binding and
  stale-plan rejection (ADR-006).
- Build the authoring skill adapter (draft → validate → diff) (ADR-006).
- Implement unknown-field round-trip preservation (ADR-001 YAML choice).
- Implement provenance citation requirement for sourced claims (ADR-008).
- Add idempotency (repeated-input) and stale-plan-rejection test cases.

### MVP-Spine — Phase 4 (Academic-research profile)

- Author the `academic-research` profile (extends core; 5 added/tightened
  types) (ADR-007, addendum 003).
- Write per-type valid+invalid fixtures (each invalid targets one rule).
- Wire `init --profile academic-research` and author each page type.

### MVP-Hardening — Phase 5 (Enrichment + index)

- Build the enrichment skill adapter on the staged plan/apply path (ADR-006).
- Implement deterministic index maintenance with dry-run/diff (ADR-010).
- Add enrichment-preview and index-stability tests.

### MVP-Hardening — Phase 6 (Hooks + CI)

- Add optional Claude Code + Git hooks invoking the engine (ADR-011).
- Add the GitHub Actions conformance workflow that fails on errors (assumes
  GitHub Actions only — addendum 002 A1; ADR-011).
- Add the cross-surface parity test (same state → same findings).

### MVP-Hardening — Phase 7 (Custom profile + lifecycle)

- Ship `profiles/custom-template/` + `profile create` / `profile validate`
  (local-only — addendum 005; ADR-012).
- Implement `init --profile <id-or-path>` for local profiles.
- Implement upgrade/uninstall with plugin-vs-repo ownership tracking (ADR-009).
- Implement approval-gated renames/deletions/migrations.
- Implement `doctor` diagnostics.

### Cross-platform gate (Cross-cutting)

- Automate release builds + checksums for all five platforms (ADR-002).
- Stand up the CI test matrix + per-platform smoke repo (assumes single
  Claude Code version floor — addendum 002 A2; ADR-011).

## Testing / validation strategy

- **Acceptance fixtures are the ship gate.** Per the addendum-004 fixture
  plan, each criterion has a concrete fixture: OKF valid/invalid core pages
  (5, 7), broken-link page (8), unknown-field round-trip (6), sourced-claim
  citations (9), duplicate-input idempotency (10), stale-plan rejection (13),
  path-traversal/symlink attempts (17), per-platform smoke repo (2, 21), plus
  the academic-research per-type fixtures (4, 5, 7, 8, 9).
- **Unit tests** for every engine package (validate, profile, plan, fsafe,
  contract, index) — happy path, edge cases, error handling.
- **Cross-surface parity test** (criterion 15): the same repository state run
  through skill, hook, CI, and direct CLI must yield identical findings.
- **Cross-platform matrix** (criterion 21): the full suite green on macOS
  arm64/x86-64, Linux arm64/x86-64, Windows x86-64 — runs continuously, gates
  MVP completion.
- **Security tests** (criterion 17): traversal/symlink-escape attempts are
  rejected pre-write; untrusted source material is treated as data, never as
  instructions.
- **The only hard quantitative gate** is zero out-of-boundary writes
  (addendum 001); other quantitative signals are instrumented but
  measurement-only for MVP.
- Any slice that bounds coverage (e.g. defers a CI target) must say so in its
  issues rather than silently narrowing the matrix.

## Risks and mitigations

### Risk 1 — One implementation drifts into per-surface logic

If a skill, hook, or CI step grows its own validation/mutation logic, the
"same findings everywhere" invariant (criterion 15) breaks. *Mitigation:* all
policy/mutation lives in the Go engine; adapters are tested as integrations
only; the parity test in Phase 6 is a hard gate.

### Risk 2 — Round-trip data loss on author/enrich

A YAML library that doesn't preserve unknown fields silently drops human
content (fails the binding criterion 6) and erodes trust; comment loss is a
lesser, non-gated quality concern. *Mitigation:* node-aware `goccy/go-yaml`
(ADR-001) preserves unknown fields (gated) and comments (best-effort);
unknown-field round-trip fixture in Phase 3; the staged plan always previews
before write.

### Risk 3 — Assumptions silently treated as final decisions

Q2/Q3/Q6/Q7/Q8 are assumption-locked, not decided; an unrecorded override
could strand built issues. *Mitigation:* every assumption-dependent issue
names its assumption (addendum 002); the first engine ADR (ADR-001) records
the Go/YAML lock as reversible.

### Risk 4 — Wide MVP becomes an unshippable flat backlog

Ten capability areas at once risk never converging. *Mitigation:* the
must-pass spine (Phases 1–4) is sequenced and gated before hardening (Phases
5–7); each phase has an observable exit gate tied to specific criteria.

### Risk 5 — Cross-platform binary drift

Five platforms can diverge late and block release. *Mitigation:* the test
matrix runs continuously from Phase 1, not at the end; release builds +
checksums are automated (ADR-002).

## Acceptance criteria for this document

This build-out plan is acceptable when it:

- matches the MVP statement — **yes** (phases cover every in-scope capability;
  out-of-scope items mirror `design/mvp.md`);
- sequences work in realistic phases — **yes** (7 phases, standard band, 1:1
  with the addendum-004 slices; spine vs hardening separated);
- identifies initial ADRs or decisions — **yes** (see below);
- produces a practical milestone and issue structure — **yes** (4 milestones,
  ~30 issue-sized tasks, full criteria-to-phase map).

> **Phase count justification.** 7 phases — within the `standard` band (5–8) —
> because the addendum-004 slice order already partitions the MVP into seven
> exit-gated delivery units (Slices 0–6) plus a continuous cross-platform
> gate; collapsing them would hide the must-pass-spine boundary the PRD review
> required, and splitting them finer would fragment single-engine concerns.

## Decisions needing ADRs

Direct input to `adr-writer` — one architectural question each. Provisional
numbers; `adr-writer` allocates final numbers. ADR-001–ADR-005 are
prerequisites for Phase 1 and should be drafted first.

1. **ADR-001 — Go toolchain & YAML library (Q2).** Go 1.24.x + `goccy/go-yaml`
   for node-aware, round-trip-preserving YAML. *Context:* criterion 6 needs
   unknown-field preservation; `yaml.v3` is archived. (MVP assumption above —
   ratify or revise.)
2. **ADR-002 — Plugin packaging & platform-binary selection (Q7 / A3).** One
   mechanism to ship + select the correct checksum-verified binary across five
   platforms. *Context:* criteria 2, 21; cross-platform gate.
3. **ADR-003 — JSON skill↔engine contract & exit codes (Q8 / A4).** Envelope
   shape, versioning (start at v1), and the stable exit-code set. *Context:*
   criterion 14; the contract every adapter depends on.
4. **ADR-004 — Validation architecture.** OKF-vs-profile separation and the
   three-severity model. *Context:* criteria 5, 7, 8.
5. **ADR-005 — Safe filesystem layer.** Atomic writes, path/symlink safety,
   bounded write scope. *Context:* criterion 17; the zero-out-of-boundary-write
   hard gate.
6. **ADR-006 — Staged mutation model.** inspect/plan/apply, hash-binding,
   stale-plan rejection. *Context:* criteria 11, 12, 13.
7. **ADR-007 — Profile system.** Data-driven profiles, one inheritance level;
   `core` + `academic-research`. *Context:* criteria 4, 5; addendum 003.
8. **ADR-008 — Provenance & citation model.** Resolvable citations for sourced
   model claims; citation preservation. *Context:* criterion 9.
9. **ADR-009 — Install/upgrade/uninstall ownership model.** Plugin-owned vs
   repo-owned assets; preserve content + local extensions. *Context:* criteria
   1, 3, 20.
10. **ADR-010 — Index maintenance.** Deterministic, no model calls, stable
    ordering. *Context:* criterion 11 support; PRD §14 reliability.
11. **ADR-011 — Hooks & CI integration.** Single-engine invocation; GitHub
    Actions workflow (A1); Claude Code version floor (A2); post-write hooks as
    feedback. *Context:* criteria 15, 18, 19.
12. **ADR-012 — Custom-profile boundary.** Local-file only for MVP; registry +
    third-party trust → Phase 3. *Context:* criterion 4 (custom); addendum 005;
    Q5.
