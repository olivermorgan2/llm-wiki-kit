# ADR-004: Separate OKF/profile reporting with a three-severity validation model

**Status:** proposed
**Date:** 2026-06-30

## Context

The engine's `validate` command must report **base OKF conformance separately
from profile conformance** (PRD §8 FR6) and apply to every file including direct
human edits (FR8). Findings carry three configurable severities — **Error**
(fails validation/CI), **Warning** (reported, non-failing by default),
**Suggestion** (advisory). Defaults (FR8): invalid YAML / missing
profile-required field / wrong field type = error; broken links / stale index =
warning unless promoted by the profile. Acceptance criteria: 5 (separate OKF and
profile conformance), 7 (malformed YAML and missing profile-required fields fail
validation), 8 (broken links at configured severity). Addendum 003 fixes
profile-specific defaults (missing required section = error; `source` with
neither `doi` nor `canonical_url` = suggestion; unsupported `claim` evidence =
error).

Risks addressed (`knowledge/risks.md`): profile rules mistaken for universal OKF
rules (report separately) and model overwrites human work (validation applies to
human edits too). The Codex review's success-signal finding bounds this:
validator false-positive rate has **no MVP threshold** — instrument, don't gate
(addendum 001).

## Options considered

### Option A: One validation engine emitting findings tagged by ruleset (OKF vs profile) and by one of three configurable severities, with profile-overridable promotion

- Pros: a single deterministic pass produces both OKF and profile results, tagged
  so they render separately (criterion 5) without two code paths; three
  severities map exactly to FR8 and addendum 003; configurability lets a profile
  promote a warning to error (e.g. broken links) without forking the engine; one
  engine ⇒ identical findings across surfaces (criterion 15).
- Cons: every finding must carry ruleset + severity metadata; severity-resolution
  precedence (core default vs profile override vs baseline) is extra logic to
  specify and test.

### Option B: Two independent validators (OKF validator + per-profile validator) with hard-coded severities

- Pros: conceptually clean separation; each validator is simple in isolation.
- Cons: duplicated traversal/parsing; risk of divergent findings between passes
  (undercuts criterion 15); hard-coded severities can't honor FR8's
  "configurable" requirement or profile promotion; two engines to keep in
  lockstep.

## Decision

Adopt **Option A** — a single validation engine emitting ruleset-tagged,
three-severity, profile-configurable findings. It satisfies separate reporting
(criterion 5) and the three-severity policy (FR8, addendum 003) from one
deterministic pass, which is also what keeps findings identical across
skills/hooks/CI/CLI (criterion 15). Baseline mode (FR8) suppresses pre-existing
findings on adoption while still flagging new or worsened ones.

## Consequences

- Easier: separate OKF/profile reporting (criterion 5) falls out of finding
  tags; configurable severities and profile promotion are data, not code forks;
  baseline adoption is supported.
- Harder: severity-resolution precedence (core defaults → profile overrides →
  baseline suppression) must be precisely specified and tested; every finding
  carries metadata.
- Maintain: the severity default table (core + addendum 003 profile additions),
  the OKF-vs-profile tagging, and baseline-diff logic.
- Deferred / validation implications: validation hooks are criteria 5, 7, 8
  (and 15 for cross-surface parity). Validator false-positive rate is
  measurement-only (addendum 001) — instrument, don't gate, for MVP.
  Provenance/citation rules (criterion 9) interlock but are specified in ADR-008;
  index-consistency checks interlock with ADR-010. No new public contract beyond
  ADR-003's envelope.
