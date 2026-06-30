# Normalized PRD — llm-wiki-kit

> Canonical 11-field normalized PRD, produced by `prd-normalizer` from
> [`design/prd.md`](prd.md) ("PRD: Claude Code Kit for OKF Wikis",
> last updated 2026-06-20). The source PRD is left untouched. `[TBD]`
> marks fields the source does not resolve; nothing here is invented.
> Downstream skills (`prd-to-mvp`, `adr-writer`) read this file only.

## 1. Product name

llm-wiki-kit (working name) — the product is described in the source PRD
as the "Claude Code Kit for OKF Wikis". The final plugin name and command
namespace are unresolved (see Open questions and PRD §19).

## 2. One-line description

A versioned Claude Code plugin, paired with a bundled Go `llm-wiki`
engine, that lets technical researchers and teams create and maintain
portable, repository-native knowledge bundles in the Open Knowledge
Format (OKF) so their knowledge stays structured, validated, and
agent-readable without a hosted wiki.

## 3. Problem

Teams that build agent systems scatter useful knowledge across docs,
papers, code, schemas, notes, dashboards, and memory; agents cannot
consume it consistently because it lacks stable structure, explicit
relationships, and predictable provenance. OKF gives a minimal
interchange format but no opinionated authoring workflow, domain
templates, deterministic validation beyond base conformance, or
repository maintenance automation — so every team rebuilds those
conventions itself.

## 4. Target users

**Primary:** a technically proficient researcher or technical team
already using Claude Code and Git that wants durable, inspectable,
agent-readable knowledge without a hosted wiki or proprietary database.

**Initial domain user:** an academic or scientific literature researcher
organizing sources, concepts, claims, methods, entities, and syntheses
with explicit provenance.

**Secondary:** AI engineers maintaining agent context; software, data,
analytics, and operations teams; technical writers and open-source
maintainers; teams that want to define their own domain profile.

## 5. Goal

Let a user create a useful LLM wiki quickly, keep its structure
trustworthy as it grows, and adapt it to a specific domain without
sacrificing OKF portability. Claude Code acts as authoring/maintenance
assistant; the Go engine remains the deterministic authority for
structure, validation, diffs, and writes; the repository stays the
source of truth.

## 6. User stories / scenarios

- As a researcher starting a wiki, I run init so I get a valid repository
  structure, default profile, templates, instructions, and validation
  with no runtime setup, so I can author immediately.
- As an author, I invoke the authoring skill with notes/URLs/files so the
  kit drafts a correctly structured page, validates it, and shows a diff,
  so I add knowledge without breaking conformance.
- As a maintainer, I enrich an existing page and the kit previews the diff
  before applying via a staged plan/apply workflow, so my human edits and
  unknown fields are never silently overwritten.
- As a direct editor, I edit Markdown by hand and run `llm-wiki validate`,
  so I get the same validation contract that model-assisted workflows use.
- As a domain owner, I create a custom profile that extends `core` without
  recompiling the CLI, so my team's conventions are enforced.
- As a team, I rely on a ready-to-use CI workflow that runs the same
  engine, so CI is the authoritative enforcement of conformance.

## 7. Core capabilities

- Versioned Claude Code plugin with a bundled, self-contained Go `llm-wiki`
  CLI (no Go/Python/Node/runtime required); versioned, checksum-verified
  release artifacts across macOS arm64/x86-64, Linux arm64/x86-64, Windows
  x86-64.
- Install/initialize into new and non-empty Git repositories; `--dry-run`;
  refuse silent overwrites; record plugin/CLI/OKF/profile versions.
- Upgrade and uninstall that distinguish plugin-owned assets from
  repository-owned content and preserve local profile extensions and wiki
  content.
- Base OKF validation reported separately from stricter profile validation;
  three severities (error/warning/suggestion); baseline mode for adoption.
- Generic `core` profile and an `academic-research` profile (one
  inheritance level); custom-profile template and scaffolding workflow.
- Authoring and enrichment skills that delegate deterministic work to the
  engine; staged inspect/plan/apply mutation workflow with hash-bound,
  stale-rejecting plans.
- Provenance: resolvable citations for sourced model-generated claims;
  preservation of existing citations and unknown fields.
- Deterministic index maintenance (no model calls); safe filesystem
  behavior (atomic writes, symlink/path-traversal rejection, bounded write
  scope).
- Optional Claude Code and Git hooks; ready-to-use CI workflow; `doctor`
  diagnostics; risk-based review/approval policy.
- Versioned JSON skill-to-engine contract with documented stable exit
  codes.

## 8. Non-goals

- Hosted SaaS or proprietary wiki UI; real-time collaboration; enterprise
  permissions.
- Embeddings, vector storage, or semantic search infrastructure.
- External-system ingestion connectors.
- Automated schema evolution or migrations.
- Semantic contradiction and near-duplicate detection as deterministic
  checks.
- Multiple active bundles per repository; cross-bundle linking.
- Windows ARM64; static publishing; graph visualization.

## 9. Constraints and preferences

- Implementation language for the engine is Go; distribution is a Claude
  Code plugin (the platform's reusable, updateable mechanism).
- Core content stays plain Markdown + YAML frontmatter; wiki remains usable
  after the plugin is removed; profiles/config are plain, version-controlled
  text.
- Targets OKF v0.1, pinned in bundle config; uses bundle-root-relative
  Markdown links as canonical output.
- One deterministic implementation shared by skills, hooks, CI, and direct
  CLI; skill scripts may only be thin adapters with no duplicated policy or
  mutation logic.
- Root `CLAUDE.md` should stay under 200 lines; `SKILL.md` files hold only
  workflow/judgment/routing, not executable logic.
- Security: source material is untrusted input and cannot override
  instructions; release checksums verified before executing artifacts.
- Performance reference (5,000 docs): full validation < 5s, changed-file
  validation < 1s, index regeneration without a model call.
- Default branch `main`; repository `olivermorgan2/llm-wiki-kit`.
- License, plugin/command name, Go YAML library, supported Go version, CI
  template scope, and min Claude Code version are not yet decided
  (see Open questions).

## 10. Success signals

- Activation: share of users who create a first conformant page; median
  time from install to first conformant page (target: < 15 min for a
  proficient user).
- Quality: share of generated pages passing validation without manual
  structural repair; share of authoring diffs accepted without substantive
  correction; broken-link findings per 100 pages; validator
  false-positive rate; share of sourced model-generated claims with
  resolvable citations.
- Reliability: successful upgrade rate; rate of interrupted operations
  recovered without corruption; confirmed writes outside configured
  boundaries (target: zero).
- The product must NOT optimize for average link count or for the share of
  edits performed through kit workflows.

## 11. Open questions

(from PRD §19, plus carried decisions still pending)

- Plugin name, command namespace, license, and marketplace location.
- Exact Go YAML library and supported Go version for development.
- Whether GitHub Actions is the only MVP CI template or one of several.
- Exact research-profile templates and conditional-section syntax.
- Profile registry and trust model for third-party profiles.
- Minimum supported Claude Code version.
- Exact packaging mechanism for selecting the correct platform binary
  inside the plugin.
- JSON contract versioning and compatibility policy.
- [TBD] Pending adversarial Codex review of this normalized PRD — findings
  to be recorded under `design/prd-addenda/` and `knowledge/`.
