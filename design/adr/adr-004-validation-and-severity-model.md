# ADR-004: Separate OKF/profile reporting with a three-severity validation model

**Status:** accepted
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
skills/hooks/CI/CLI (criterion 15).

**Severity-resolution precedence (decided here).** Effective severity resolves
in a fixed order: **core defaults → profile overrides → baseline suppression**.
Core supplies the FR8 default severities; a profile may promote/demote within
the configurable set (e.g. broken links warning→error); baseline suppression is
applied **last and only as a differential filter**, never as a severity change.
Phase 1 issues must specify and test this precedence before implementation.

**Baseline scope and its hard boundary against structural errors.** Baseline
mode (FR8) supports adoption on a pre-existing wiki by suppressing **already
present** findings while still flagging new or worsened ones. It is strictly a
**differential filter over findings the validator successfully produced**, and
it is bounded as follows:

- **Malformed YAML is never baseline-suppressible.** Parsing must succeed
  before any finding — or any baseline comparison — can be computed, so a parse
  failure is a hard validation error that no baseline can hide (criterion 7
  holds unconditionally).
- **Missing profile-required fields** are hard errors (criterion 7). A baseline
  may record such a finding as pre-existing so an *adoption* run does not fail
  on legacy debt, but this suppression applies **only in the explicit adoption
  baseline mode** and is **not** honored by release-gate / CI fixture runs,
  which always evaluate errors at full severity.
- **Exit codes:** when baseline suppresses the last remaining error-severity
  finding, the run exits success (or success-with-warnings) per ADR-003; when
  any un-suppressed error remains — always including malformed YAML — the run
  exits with the validation-failure code. CI / release-gate runs do not load an
  adoption baseline, so criterion 7 always fails the gate there.

## Consequences

- Easier: separate OKF/profile reporting (criterion 5) falls out of finding
  tags; configurable severities and profile promotion are data, not code forks;
  baseline adoption is supported.
- Harder: the severity-resolution precedence decided above (core defaults →
  profile overrides → baseline suppression) must be precisely implemented and
  tested, including the baseline/structural-error boundary; every finding
  carries metadata.
- Maintain: the severity default table (core + addendum 003 profile additions),
  the OKF-vs-profile tagging, and baseline-diff logic.
- Deferred / validation implications: validation hooks are criteria 5, 7, 8
  (and 15 for cross-surface parity). Validator false-positive rate is
  measurement-only (addendum 001) — instrument, don't gate, for MVP.
  Provenance/citation rules (criterion 9) interlock but are specified in ADR-008;
  index-consistency checks interlock with ADR-010. No new public contract beyond
  ADR-003's envelope.
