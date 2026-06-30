# Addendum 002 — MVP planning assumptions (packaging, CI, version floor, contract)

- **Source finding:** Codex PRD review (2026-06-30), Blocking finding #2
  (High). Open questions Q3, Q6, Q7, Q8 (and adjacent Q1, Q2) are not
  harmless implementation details — they set packaging shape, CI template
  scope, Claude Code version floor, and JSON-contract compatibility, all of
  which determine implementation shape and acceptance tests.
- **Decision:** **Assumption-lock for MVP planning.** Oliver has not made
  these decisions final. Rather than block `/prd-to-mvp`, this addendum
  records explicit *planning assumptions* that `/prd-to-mvp` and
  `/issue-planner` may build on. Each assumption is overridable by Oliver;
  if overridden, the affected MVP issues are revised, not silently kept.
  None of these is recorded in `knowledge/log.md` as a settled
  product decision — the open questions stay **open** with an
  "assumption-locked" marker.
- **Knowledge links:** `knowledge/open-questions.md` (Q3/Q4/Q6/Q7/Q8
  marked must-resolve-or-assumption-lock; QB2 = this review outcome).

## Assumptions (each labeled with the open question it covers)

### A1 — CI template scope (Q3)
**Assumption:** GitHub Actions is the **only** MVP CI template. The plugin
ships one ready-to-use GitHub Actions workflow that invokes the same
`llm-wiki` engine; other CI systems (GitLab CI, CircleCI, etc.) are
**post-MVP**. Acceptance criterion 19 ("CI workflow fails on configured
errors") is tested against this single GitHub Actions workflow.
**If overridden:** add one issue per additional CI target; the engine
contract does not change (CI is a thin invoker), so the blast radius is the
workflow files and their tests only.

### A2 — Minimum supported Claude Code version (Q6)
**Assumption:** MVP targets a **single minimum Claude Code version**,
pinned in `plugin.json` / docs at implementation time (the current stable
release at first-issue time). **No backward-compatibility shim** for older
Claude Code versions is in MVP scope. Behavior on versions below the floor
is "documented unsupported," not "gracefully degraded."
**If overridden:** a compatibility-shim capability becomes its own
post-MVP issue; it does not gate the MVP acceptance criteria.

### A3 — Platform-binary selection mechanism (Q7)
**Assumption:** the plugin uses **one** binary-selection mechanism to pick
the correct platform binary at install/run time (detect OS+arch, then
resolve a single matching checksum-verified `bin/llm-wiki`). MVP commits to
exactly one mechanism — not a pluggable set — and the five supported
platforms (macOS arm64/x86-64, Linux arm64/x86-64, Windows x86-64) are
covered by that one path. The exact mechanism (e.g. per-platform binaries
shipped in the plugin vs. a launcher that selects among them) is an
implementation detail to be fixed in the install/init ADR, but MVP ships
**one**.
**If overridden:** only the install/init mechanism issue and its ADR
change; acceptance criteria 2 and 21 (correct binary per platform, tests
green on all platforms) are mechanism-agnostic.

### A4 — JSON contract versioning & compatibility (Q8)
**Assumption:** the skill-to-engine JSON contract **starts at v1** with
**no backward-compatibility guarantee until the first public release**.
During MVP development, breaking the v1 contract is allowed (skills and
engine ship together in one versioned plugin, so they move in lockstep).
The "structured responses remain backward compatible within a major
contract version" reliability requirement (PRD §14) is interpreted as
**applying from the first released v1 onward**, not across pre-release
iterations. Every structured response still carries its contract version,
operation, status, findings, affected paths, and approval fields from day
one (acceptance criterion 14).
**If overridden:** a pre-release compatibility policy becomes an explicit
ADR; it does not change the acceptance criteria, only the freedom to break
the contract before release.

### A5 — Naming / license / toolchain (Q1, Q2) — non-blocking, recorded for completeness
**Assumption:** MVP planning proceeds under the **working** names
(`llm-wiki-kit` repo, `llm-wiki` CLI) and the repository-scaffold license
(MIT). The Go YAML library and supported Go version are deferred to the
first engine ADR; they do not block MVP *scoping* (issue generation), only
the first engine *implementation* issue. These are flagged so a late rename
or toolchain choice is a known, bounded follow-up (QB1).

## Net effect on `/prd-to-mvp`

`/prd-to-mvp` may scope the MVP as if A1–A5 hold. Any issue that depends on
an assumption must name the assumption (e.g. "assumes GitHub Actions only,
per addendum 002 A1") so that an override later is traceable to the issues
it touches.
