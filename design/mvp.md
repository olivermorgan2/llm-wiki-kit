# llm-wiki-kit — MVP

**Last updated:** 2026-06-30

> Scoped from [`design/prd-normalized.md`](prd-normalized.md) **plus** the
> accepted PRD addenda [`design/prd-addenda/001`–`005`](prd-addenda/), which
> are authoritative alongside the normalized PRD. The 21 functional
> acceptance criteria are carried by
> [addendum 001](prd-addenda/001-mvp-acceptance-criteria.md); the slice order
> by [addendum 004](prd-addenda/004-mvp-slice-order-and-fixture-plan.md). This
> document does not rewrite the addenda — it sequences them.

## Product name

**llm-wiki-kit** (working name; CLI binary `llm-wiki`).

> **MVP assumption — name lock.** The repo/plugin working name is
> `llm-wiki-kit` and the CLI is `llm-wiki` for all MVP planning and
> implementation. The final public plugin name, command namespace, and
> marketplace listing remain open (Q1 / QB1) and Oliver can override before
> packaging/public release; an override is a bounded rename, not a scope
> change. See [`knowledge/open-questions.md`](../knowledge/open-questions.md).

## One-line description

A versioned Claude Code plugin paired with a bundled, self-contained Go
`llm-wiki` engine that lets a technical researcher or team create and
maintain portable, repository-native knowledge bundles in the Open Knowledge
Format (OKF v0.1) — structured, validated, and agent-readable without a
hosted wiki.

## Product goal

Ship a first release in which a proficient user can install the plugin into a
Git repository, initialize a conformant OKF wiki (core or academic-research
profile), author and enrich pages through a deterministic, preview-before-write
workflow, and rely on CI as the authoritative conformance gate — all without
any language runtime installed and without the wiki ever depending on the kit
to stay readable. Success is that the engine is the single deterministic
authority across skills, hooks, CI, and direct CLI, and that no managed
operation can write outside its configured boundaries.

## Target users

### Primary user

A technically proficient researcher or small technical team already using
Claude Code and Git who wants durable, inspectable, agent-readable knowledge
without a hosted wiki or proprietary database.

### Secondary user

The initial domain user is an academic / scientific-literature researcher
organizing sources, claims, methods, entities, questions, and syntheses with
explicit provenance. Also benefiting: AI engineers maintaining agent context,
and technical teams that want to define their own domain profile.

## Core problem

Teams that build agent systems scatter knowledge across docs, papers, code,
schemas, notes, and memory. Agents cannot consume it consistently because it
lacks stable structure, explicit relationships, and predictable provenance.
OKF supplies a minimal interchange format but no opinionated authoring
workflow, domain templates, deterministic validation beyond base conformance,
or repository-maintenance automation — so every team rebuilds those
conventions itself. The alternatives (hosted wikis, proprietary databases,
ad-hoc Markdown) either lock content behind a runtime, lose portability, or
provide no enforceable structure at all.

## Product principles

1. **Format, not platform.** Content stays plain Markdown + YAML and remains
   readable, editable, and portable after the plugin is removed. No
   proprietary runtime is ever required to read a wiki.
2. **The deterministic engine is the authority.** All parsing, validation,
   indexing, diff planning, and filesystem safety live in the Go engine. Model
   judgment orchestrates and drafts; it is never described as, or substituted
   for, deterministic validation.
3. **One implementation, four surfaces.** Skills, hooks, CI, and the direct
   CLI all call the same engine; skill scripts are thin adapters with no
   duplicated policy or mutation logic. Identical repository state yields
   identical findings everywhere.
4. **OKF baseline and profile conformance are reported separately.** A profile
   rule is never presented as a universal OKF rule.
5. **Review proportional to risk.** New pages are fast; existing-page edits are
   previewed via a staged plan/apply; renames, deletions, and migrations
   require explicit approval. Unknown frontmatter fields and unrelated human
   content are never silently overwritten.
6. **Provenance without theatre.** Sourced, model-generated claims must cite
   resolvable sources; human statements need not unless the profile says so.

## MVP scope

The PRD defines a deliberately wide MVP (PRD §18). Per
[addendum 004](prd-addenda/004-mvp-slice-order-and-fixture-plan.md), the full
surface stays in MVP — it is **sequenced**, not cut. The "must-pass spine"
(Slices 0–3) proves the architecture end to end; Slices 4–6 harden the rest of
the MVP surface and remain required for MVP completion.

### In scope

- **Engine + contract spine.** Self-contained Go `llm-wiki` CLI; OS/arch
  detection; checksum verification; versioned JSON skill↔engine contract;
  documented stable exit codes; core-profile `validate` with OKF-vs-profile
  reporting at three severities (error/warning/suggestion); safe filesystem
  layer (atomic writes, bounded write scope, symlink/path-traversal rejection).
- **Install / init.** Install into new and non-empty Git repos without data
  loss; `--dry-run`; refuse silent overwrites; record plugin/CLI/OKF/profile
  versions; checksum-verified correct binary per platform; init with core,
  academic-research, or a valid local custom profile.
- **Authoring + staged mutation.** `page inspect` / `page plan` / `page apply`
  with hash-bound, stale-rejecting plans; authoring skill that drafts,
  validates, and shows a diff; preserve unknown fields; resolvable provenance
  citations for sourced model claims; no duplicate page on repeated unchanged
  input.
- **Academic-research profile.** The profile and per-type acceptance fixtures
  from [addendum 003](prd-addenda/003-academic-research-profile-contract.md)
  (page types `source`, `claim`, `method`, `question`, `synthesis` over `core`,
  one inheritance level).
- **Enrichment + index maintenance.** Enrichment of existing pages through the
  same staged plan/apply path; deterministic index maintenance with no model
  calls, stable ordering, and dry-run/diff.
- **Hooks + CI.** Optional Claude Code and Git hooks invoking the same engine;
  one ready-to-use GitHub Actions workflow that fails on configured errors.
- **Custom profile (local-file only) + lifecycle.** `profile create` /
  `profile validate`; `init --profile <id-or-path>` from a local profile;
  upgrade and uninstall that preserve wiki content and local profile
  extensions; `doctor` diagnostics.
- **Cross-platform.** Versioned, checksum-verified release artifacts and a
  green automated test matrix on macOS arm64/x86-64, Linux arm64/x86-64,
  Windows x86-64.

### Out of scope

**Product-level non-goals (from the normalized PRD §8):**

- Hosted SaaS or proprietary wiki UI; real-time collaboration; enterprise
  permissions.
- Embeddings, vector storage, or semantic search infrastructure.
- External-system ingestion connectors.
- Automated schema evolution or migrations.
- Semantic contradiction and near-duplicate detection as deterministic checks.
- Multiple active bundles per repository; cross-bundle linking.
- Windows ARM64; static publishing; graph visualization.

**Deferred by this MVP scoping (would otherwise be assumed in scope):**

- A profile **registry** and a third-party-profile **trust model** (signing,
  provenance, allow-lists, remote fetch) — scoped to **Phase 3 (Ecosystem)**
  per [addendum 005](prd-addenda/005-custom-profile-boundary.md). MVP custom
  profiles are local-file only.
- CI templates beyond GitHub Actions (GitLab CI, CircleCI, …) — post-MVP per
  [addendum 002 A1](prd-addenda/002-mvp-planning-assumptions.md). The engine
  contract does not change, so each is a thin follow-up.
- A backward-compatibility shim for Claude Code versions below the MVP floor —
  post-MVP per [addendum 002 A2](prd-addenda/002-mvp-planning-assumptions.md).
  Behavior below the floor is "documented unsupported," not "gracefully
  degraded."
- Pre-release backward-compatibility of the JSON contract — the contract
  starts at v1 and may break during MVP development (skills and engine ship in
  lockstep); compatibility applies from first public release per
  [addendum 002 A4](prd-addenda/002-mvp-planning-assumptions.md).
- Conditional-section syntax and richer profile field sets beyond the
  addendum-003 minimum (Q4 assumption-locked to that minimum).

## Primary outputs

- A versioned Claude Code **plugin** bundling per-platform, checksum-verified
  `llm-wiki` binaries, skills, optional hooks, profiles (`core`,
  `academic-research`, `custom-template`), and a GitHub Actions CI workflow.
- The `llm-wiki` **CLI** itself — usable directly, identically to the
  skill/hook/CI paths.
- In a user's repo: an initialized OKF wiki bundle (config, profile,
  templates, instructions, deterministic indexes) of plain Markdown + YAML.

## Success criteria

The MVP succeeds if a user can:

1. Install the plugin into a real Git repository (new or non-empty) and run the
   correct checksum-verified `llm-wiki` binary with no language runtime
   installed, without losing any existing file.
2. Initialize a conformant wiki under the core, academic-research, or a local
   custom profile, and immediately author a valid page.
3. Author and enrich pages through the staged inspect/plan/apply workflow —
   seeing a diff before any write, with unknown fields and human content
   preserved, sourced claims carrying resolvable citations, and a stale plan
   refused rather than silently applied.
4. Get identical OKF-vs-profile validation findings whether the check runs from
   a skill, a hook, the CI workflow, or the direct CLI — with CI failing the
   build on configured errors.
5. Upgrade or uninstall the kit without losing wiki content or local profile
   extensions, and never observe a write outside the configured boundary.

## Deferred to later

- **Profile registry + third-party trust model** → Phase 3 (Ecosystem);
  reopening it reopens Slice 6 and requires a dedicated trust-model ADR before
  any remote-fetch issue ([addendum 005](prd-addenda/005-custom-profile-boundary.md)).
- **Additional CI targets** beyond GitHub Actions → post-MVP, one issue per
  target.
- **Claude Code backward-compatibility shim** → post-MVP, its own issue; does
  not gate MVP acceptance.
- **Richer academic-research templates / conditional-section syntax** → a
  future addendum supersedes the addendum-003 minimum and regenerates fixtures.

## Acceptance criteria for this document

This MVP statement is acceptable when it:

- names a clear product and user — **yes** (llm-wiki-kit; technical
  researcher/team on Claude Code + Git);
- lists what is in and out of scope without ambiguity — **yes** (every core
  capability classified; product non-goals separated from MVP-scoping
  deferrals by source);
- can drive the build-out plan, ADRs, and issue backlog without further
  interpretation — **yes** (see
  [`design/build-out-plan.md`](build-out-plan.md)).

**Binding ship gate.** The 21 functional acceptance criteria in
[addendum 001](prd-addenda/001-mvp-acceptance-criteria.md) are the binding
definition of done; the success criteria above are their user-facing summary.
The only hard *quantitative* release gate is **zero out-of-boundary writes**
(addendum 001); all other quantitative success signals are measurement-only
until a baseline exists.

**Open / assumption-locked items carried into planning** (none block scoping;
each is traceable and overridable):

- Final plugin name / namespace / license / marketplace (Q1, QB1) — working
  name locked for MVP.
- Go version and YAML library (Q2) — assumption-locked below and in the
  build-out plan's "Decisions needing ADRs"; resolved by the first engine ADR.
- Q3/Q6/Q7/Q8 — assumption-locked for MVP per
  [addendum 002](prd-addenda/002-mvp-planning-assumptions.md).
- Q5 (registry/trust) — scoped to Phase 3 per addendum 005.
