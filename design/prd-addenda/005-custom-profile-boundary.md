# Addendum 005 — Custom-profile boundary (local-only for MVP)

- **Source finding:** Codex PRD review (2026-06-30), non-blocking finding
  (Medium). Custom profiles are in MVP scope (`design/prd.md` §8, §11.4),
  while third-party profile registry/trust is an open question (Q5).
  Leaving the two adjacent creates avoidable scope ambiguity: a local
  custom-profile template can ship without solving registry trust.
- **Decision:** **Accept — draw the boundary explicitly.** MVP custom
  profiles are **local-file only**. Registry distribution and third-party
  trust are **Phase 3 (Ecosystem)** unless Oliver decides otherwise.
- **Knowledge links:** `knowledge/open-questions.md` Q5 (scoped to Phase 3,
  stays open); `knowledge/log.md` (verdict entry).

## In MVP scope (custom profiles)

- A documented `profiles/custom-template/` (profile.yaml, templates,
  `examples/valid/`, `examples/invalid/`, README) — per PRD §11.4.
- `llm-wiki profile create <id>` — scaffold a local profile from the
  template.
- `llm-wiki profile validate <path>` — validate a local profile file/dir.
- `llm-wiki init --profile <id-or-path>` — initialize with a **local**
  profile by id or filesystem path.
- One inheritance level: every custom profile `extends: core` (PRD §11).
- No CLI recompilation required to use a custom profile (PRD §11.4).

A custom profile is a plain-text, version-controlled artifact living **in
the user's own repository** (or a path they supply). Trust is implicit:
it's the user's own file, reviewed like any other file in their repo.

## Explicitly out of MVP scope (→ Phase 3)

- A profile **registry** or hosted index of shareable profiles.
- A **trust model** for third-party profiles (signing, provenance,
  allow-lists, sandboxing of fetched profiles).
- Fetching or installing a profile from a remote/third-party source.
- Multiple / chained custom-profile inheritance (already out per PRD §11).

## Why this is safe to ship without registry trust

Local custom profiles never cross a trust boundary the user doesn't
already control: the profile is a file they authored or pulled into their
own repo and can read. The deterministic engine validates the profile
(`profile validate`) and resolves it deterministically with actionable
errors on conflicts/invalid overrides (PRD §11). No remote fetch means no
supply-chain or untrusted-author surface in MVP — that surface only appears
with registry distribution, which is precisely what Phase 3 must design
the trust model for.

**Assumption lock (Q5):** Q5 (registry + third-party trust model) stays
**open** but is scoped to **Phase 3**; it does **not** block MVP planning.
If Oliver pulls registry/trust into MVP, that reopens Slice 6 (addendum
004) and requires its own trust-model ADR before any remote-fetch issue is
filed.
