# ADR-007: Data-driven profile system with one inheritance level

**Status:** proposed
**Date:** 2026-07-03

## Context

The wiki's rules are split into a universal **Objective Knowledge Format (OKF)**
layer and pluggable **domain profiles** that add or tighten rules on top of it
(PRD §5, §8). Acceptance criterion **4** requires `init`/`validate` to work
against the shipped **`core`** and **`academic-research`** profiles (custom
profiles are in scope but bounded — see below), and criterion **5** requires
OKF and profile conformance to be reported **separately** so a profile rule is
never presented as universal. [ADR-004](adr-004-validation-and-severity-model.md)
(accepted) already fixed the OKF-vs-profile separation and the three-severity
model; what it did **not** fix is the **profile system boundary** itself: what a
profile *is* on disk, how profiles compose, how one is resolved, and what
`init` materializes. Phase 2 needs exactly the `core` slice of this to stand up
`init`.

Constraints already set by accepted artifacts and addenda:

- The existing `internal/profile` seam (`internal/profile/profile.go`) fixes a
  `Profile` value and a `Loader` that routes **all** YAML through the injected
  `yamladapter.Adapter` ([ADR-001](adr-001-go-toolchain-and-yaml.md), accepted)
  — never `goccy/go-yaml` at call sites. This ADR ratifies that shape rather
  than inventing a new one.
- [Addendum 003](../prd-addenda/003-academic-research-profile-contract.md) fixes
  the **minimum** `academic-research` contract (page types, fields, sections,
  per-type valid/invalid fixtures) — the profile format must express exactly
  that, no more.
- [Addendum 005](../prd-addenda/005-custom-profile-boundary.md) and open
  question **Q5** scope third-party **profile registry + trust** to Phase 3;
  the MVP custom-profile surface is **local-file only**. Profiles must therefore
  be declarative **data**, not code, so that trusting one later is a data-review
  problem, not a code-execution problem.
- Risk (`knowledge/risks.md`): "custom-profile complexity becomes a programming
  language" and "schema becomes rigid too early" — both push toward minimal,
  declarative, shallow composition.

## Options considered

### Option A: Declarative data profiles with exactly one inheritance level (a profile extends `core`), resolved by profile id from the plugin-shipped set

- Pros: a profile is a YAML data file loaded through the ADR-001 adapter and
  consumed by the ADR-004 validation engine, so OKF-vs-profile separation
  (criterion 5) falls out of the data model — the engine knows which rules came
  from `core`/OKF and which from the domain profile. **One** inheritance level
  (`academic-research` extends `core`; `core` extends nothing) is enough to
  express addendum 003 while keeping resolution trivial and terminating, and it
  directly ratifies the `internal/profile` seam. Resolution is **by profile
  id** against the shipped set for the MVP, with a documented seam to add
  **local-path** resolution in Phase 7 without changing the format. Declarative
  data keeps custom profiles a review-the-data problem, honoring the Phase-3
  trust boundary (Q5, addendum 005).
- Cons: a data schema cannot express arbitrary conditional logic, so genuinely
  procedural rules would need a future engine-level rule type rather than
  profile authoring; one inheritance level forbids deep profile hierarchies, so
  a future family of near-identical profiles would repeat shared rules or force
  a later ADR to widen the model.

### Option B: Profiles as code / executable rule plugins

- Pros: maximal expressiveness — a profile could compute any rule it wants.
- Cons: executing third-party profile code is exactly the trust/registry
  problem addendum 005 defers to Phase 3, pulled forward into the MVP as a
  code-execution attack surface; it contradicts "declarative data only, one
  inheritance level" in the risk register; and it makes OKF-vs-profile
  separation (criterion 5) a runtime property of untrusted code rather than a
  property of the data model. Over-powerful for the addendum-003 contract.

### Option C: Flat, self-contained profiles with no inheritance

- Pros: no resolution/merge logic at all — each profile fully lists its rules.
- Cons: `academic-research` would have to restate every `core`/OKF rule, so a
  change to `core` silently diverges from every domain profile (drift the
  OKF/profile split exists to prevent); it also blurs which rules are universal
  vs domain-specific, working against criterion 5. The single level of
  inheritance in Option A removes this duplication at negligible complexity.

## Decision

Adopt **Option A** — profiles are **declarative data files** with **exactly one
inheritance level** (`academic-research` extends `core`; `core` extends
nothing), loaded through the ADR-001 `yamladapter.Adapter` via the existing
`internal/profile` `Loader`, and consumed by the ADR-004 validation engine so
OKF and profile findings stay separable (criterion 5). Profiles are **resolved
by id** against the plugin-shipped set for the MVP; the loader keeps a
documented seam for **local-path** resolution added in **Phase 7** (the custom
profile boundary, Q5/addendum 005) **without a format change**. `init`
materializes the selected profile into the bundle as a **reference to the
shipped profile** (plus the bundle config recording which profile is active),
not a copy of the rule data — Phase 2 exercises only the **`core`** path.
Option B is rejected because executing profile code pulls the Phase-3 trust
problem into the MVP; Option C is rejected because restating `core` in every
profile invites drift and blurs the OKF/profile line criterion 5 protects.

## Consequences

- Easier: `init --profile core` (Phase 2) has an unambiguous target — resolve
  `core` by id and materialize its reference into the bundle; `academic-research`
  (Phase 4) is expressible as a one-level extension per addendum 003; the
  OKF-vs-profile separation (criterion 5) is a property of the data model, and
  the `internal/profile` seam is ratified, not reworked.
- Harder: the profile schema must be expressive enough for addendum 003 yet
  stay declarative and shallow; the id-resolution path and the reserved
  local-path seam both need tests; keeping `core` conservative (per the risk
  register) is an ongoing editorial discipline, not a one-time choice.
- Maintain: the profile data schema, the `core` and `academic-research` shipped
  profiles, the id-based resolver plus the documented Phase-7 local-path seam,
  and the loader's exclusive use of the ADR-001 adapter.
- Deferred / validation implications: criterion 4 is proved incrementally —
  **core** in Phase 2 (`init` + `validate` clean), **academic-research** in
  Phase 4, **custom/local-path** in Phase 7; criterion 5 is proved by the
  ADR-004 engine over these profiles. Third-party profile **registry/trust**
  stays deferred to Phase 3 (Q5, addendum 005); the profile **format** is fixed
  here so that deferral changes only resolution, not the data model. The
  "schema too rigid too early" and "custom-profile becomes a language" risks in
  `knowledge/risks.md` stay `open`, now bounded by the declarative one-level
  model.
