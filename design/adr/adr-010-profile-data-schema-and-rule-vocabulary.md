# ADR-010: Declarative profile-data schema and profile rule / citation vocabulary

**Status:** accepted
**Date:** 2026-07-04

## Context

[ADR-007](adr-007-profile-system-boundary.md) (accepted) fixed that profiles are
**declarative data files** with **one inheritance level** (`academic-research`
extends `core`; `core` extends nothing), loaded through the ADR-001
`yamladapter.Adapter` via the `internal/profile` `Loader`, and consumed by the
ADR-004 validation engine so OKF and profile findings stay separable
(criterion 5). It deliberately left the **exact profile YAML schema** open, to be
fixed when the `academic-research` profile lands.
[ADR-008](adr-008-provenance-and-citation-model.md) (accepted) fixed the citation
**mechanism** (representation, three-class offline resolver, severities, the
`core-citation-*` codes, plan-time citation-loss gate) and **sketched** — but did
not freeze — the profile citation vocabulary, deferring the exact YAML keys and
two Codex carry-ins to "the Phase 4 profile issue":

- **Carry-in №1:** resolution class 3 (repo-path) should be explicitly gated on
  `isIntraWiki`.
- **Carry-in №2:** whether a **present-but-unresolved** citation satisfies a
  require-citation obligation.

[Addendum 003](../prd-addenda/003-academic-research-profile-contract.md) is the
**authoritative minimum contract** the schema must express and *no more*: five
added/tightened page types (`source`, `claim`, `method`, `question`,
`synthesis`), their required/recommended fields, enum value sets, list-min
constraints, required Markdown sections, the source `doi`/`canonical_url`
recommended pair, and the two citation obligations (supported `claim` needs a
cited `## Evidence`; a `question` is never a valid evidence target).

Today `internal/profile/profile.go` is a stub (`Profile{ID, Version}`, a
hardcoded `Resolve` switch, a `Loader` that loads nothing); no profile YAML
exists; and every core rule is hardcoded Go in `internal/validate/rules.go` with
**no per-page-type dispatch** and **no** enum / list-min / required-section /
recommended-pair rule kinds. The **biggest Phase-4 risk**
(`knowledge/risks.md`) is that the profile schema becomes a general programming
language, or that migrating core's rules into data silently changes findings.
This ADR settles: (a) the profile YAML schema; (b) the core-rule boundary and
its golden-parity constraint; (c–d) the two ADR-008 carry-ins; (e) unknown-type
handling; (f) the finding-code namespace for profile rules. It fixes the surface
syntax ADR-007/ADR-008 deferred, so addendum 003 can still be superseded by
Oliver without reopening those ADRs.

## Options considered

### Option A: Minimal per-type declarative schema; new rule *kinds* are engine capabilities driven by profile data; core rules stay engine-code

- The profile file declares, per profiled page type, a fixed, closed set of
  **rule-kind inputs** — added required/recommended fields, enum value sets,
  list-min counts, required section titles, a recommended any-of group, evidence
  sections, and citation obligations. The engine implements each **rule kind**
  once; the profile only supplies **data** for it. `core`'s existing OKF and
  core-profile rules stay engine-code (authoritative, byte-identical); the `core`
  profile file declares core's profile-layer vocabulary **descriptively** so
  `extends: core` resolves against a real parent and the merge is exercised,
  without re-deriving existing core findings from data in Phase 4.
- Pros: expresses addendum 003 exactly with a closed vocabulary — no conditional
  logic, no expressions, no computed rules, so the "profile becomes a programming
  language" risk stays bounded; **golden parity is structural**, not a
  reproduction effort, because the code that emits every existing finding is
  untouched; the new rule kinds are profile-agnostic engine capabilities any
  future profile can reuse; and it directly realizes ADR-007 (academic-research
  *is* fully data-driven for everything it adds) and instantiates ADR-008's
  sketched vocabulary.
- Cons: two rule substrates coexist during the MVP — engine-coded core rules and
  data-driven profile rule kinds — so "all rules are data" is not yet true for
  `core`; a fully-declarative `core` migration remains a separate, golden-guarded
  future issue.

### Option B: Full migration — every rule (OKF + core + profile) becomes profile data now

- Pros: one uniform substrate; "profiles are data" is literally true end-to-end.
- Cons: the engine must **reproduce byte-identical** messages, codes, and
  severities for every shipped `core-*`/`okf-*` finding from a data
  representation, which is exactly the migration-drift risk the register warns
  about — high blast radius across every existing test for a Phase whose value is
  the *academic-research* profile, not a core rewrite. The parse-failure gate,
  the field-type rule's fixed-order message, and the recommended-field aggregation
  are engine mechanics poorly served by a data schema. Over-reach for addendum
  003.

### Option C: Expression-based schema (predicates / conditional-section DSL)

- Pros: maximal flexibility — arbitrary conditions over frontmatter and body.
- Cons: this *is* the "custom-profile complexity becomes a programming language"
  risk realized; it needs a parser/evaluator the MVP has no call for; conditional
  logic beyond addendum 003 is explicitly deferred by Q4. Rejected on the risk
  register alone.

## Decision

Adopt **Option A**. The profile schema is a **closed, per-type, declarative
vocabulary**; the engine owns a fixed set of **rule kinds** and the profile
supplies their data; `core`'s existing rules remain engine-code and byte-identical.
The following sub-decisions are load-bearing and fixed here.

### 1. Profile YAML schema (the closed vocabulary)

A profile file is loaded through the ADR-001 adapter into this shape (shown for
`academic-research`; the full field/section/enum values are addendum 003's and
are frozen in the Phase-4 profile issue that ships the file):

```yaml
profile:
  id: academic-research
  version: "1.0"
  extends: core            # exactly one level; core extends nothing

types:                     # per-page-type rules, keyed on frontmatter `type`
  source:
    required: [authors, source_type]        # fields ADDED beyond core-required
    recommended: [publication_date, review_status]
    enums:
      source_type: [paper, preprint, book, dataset, webpage, report, other]
      review_status: [unreviewed, in-review, reviewed]
    listMin:
      authors: 1
    recommendedAnyOf:                        # >=1 present, else one suggestion
      - [doi, canonical_url]
  claim:
    required: [confidence, assessment]
    enums:
      confidence: [low, medium, high]
      assessment: [supported, contested, refuted, open]
    requiredSections: [Evidence, Counterevidence, Assessment]
    evidenceSections: [Evidence]             # citation contexts for this type
    citation:
      requireWhen: { assessment: supported } # obligation trigger (sub-decision 3)
      forbiddenTargetTypes: [question]       # sub-decision 3
  method:
    requiredSections: [Assumptions, Strengths, Limitations]
  question:
    required: [status]
    enums:
      status: [open, answered, abandoned]
    requiredSections: [Evidence gap]
  synthesis:
    requiredSections: [Scope, Findings, Agreement, Disagreement, "Evidence gaps"]
```

The vocabulary is **exactly** these keys — `required`, `recommended`, `enums`,
`listMin`, `recommendedAnyOf`, `requiredSections`, `evidenceSections`, and
`citation.{requireWhen, forbiddenTargetTypes}` — plus the `profile` header. There
is **no expression syntax, no conditional-section syntax, and no cross-field
logic** beyond `citation.requireWhen` (a single field-equals-value trigger). Any
richer need supersedes addendum 003 with a new addendum and, if it needs a new
rule *kind*, a new ADR — not free-form profile authoring. `requireWhen` is the
one deliberate concession, and it is intentionally the weakest possible
conditional (one field, one value) so it cannot grow into a predicate language.
Per-type `required`/`recommended` list **only the fields the profile adds** on
top of what the page inherits from core (which is why `title`/`description`/`type`
never appear there — core already requires them for every page).

### 2. Core-rule boundary and the golden-parity constraint (binding on I2)

`core`'s OKF rules (`okf-*`) and core-profile rules (`core-*`:
`core-required-title`, `core-required-description`, `core-field-type`,
`core-recommended-missing`, `core-kebab-filename`, `core-broken-link`, and the
`core-citation-*` family) **stay engine-code and are not re-derived from profile
data in Phase 4**. Every existing finding **code, severity, message, and
aggregation stays byte-identical**, and **every existing test passes unchanged** —
this is a hard acceptance gate on the loader issue (I2). The shipped
`profiles/core/profile.yaml` declares core's profile-layer vocabulary
**descriptively** (its required/recommended fields) so that (i) `extends: core`
has a real, loadable parent, (ii) the one-level merge is exercised and tested,
and (iii) the file documents the core contract; but the Phase-4 engine emits the
existing core findings from the **same Go code as today**, not from that file. A
fully-declarative `core` migration, if ever wanted, is a **separate future issue
guarded by golden tests**, explicitly out of Phase-4 scope. Rationale: the value
of Phase 4 is the `academic-research` profile; a core rewrite would spend the
Phase's risk budget on drift with no user-visible gain.

### 3. Profile rule kinds, the citation obligations, and both ADR-008 carry-ins

The engine gains a **per-type rule-dispatch layer** keyed on the page's
frontmatter `type`. For a page whose `type` is a **profiled** type in the
resolved profile, it evaluates the type's declared rules; for any other page it
runs **only** the existing engine rules (sub-decision 5). The rule kinds and
their default severities (matching addendum 003) are:

| Rule kind | Fires when | Code | Default severity |
|---|---|---|---|
| Required field | a declared `required` field is absent/empty | `profile-required-field` | error |
| Enum | a present field's value is outside its `enums` set | `profile-field-enum` | error |
| List-min | a `listMin` field is not a list of ≥ N items | `profile-list-min` | error |
| Required section | a `requiredSections` ATX heading is absent | `profile-required-section` | error |
| Recommended any-of | no member of a `recommendedAnyOf` group is present | `profile-recommended-pair` | suggestion |
| Citation required | `citation.requireWhen` holds and the type's evidence section has **no citation** | `profile-citation-required` | error |
| Forbidden target type | a citation resolves to a page whose `type` is in `forbiddenTargetTypes` | `profile-citation-target-type` | error |

Required Markdown sections reuse the shipped `parseATX` heading parser
(`internal/validate/citations.go`); section-title matching is exact and
case-sensitive, consistent with `splitEvidenceContexts`. Every finding aggregates
one problem into one finding per `{ruleset, code, path}` (ADR-004 FR8) and
resolves severity inside ADR-004's precedence (core defaults → profile overrides
→ baseline suppression), inheriting the parse-failure gate.

**Carry-in №2 (present-but-unresolved satisfies the obligation) — decided.**
`profile-citation-required` fires on the **absence of *any* inline-link citation**
in the designated evidence section. **Resolvability is a separate axis:** a
present-but-unresolved citation **satisfies** the require-citation obligation
(so `profile-citation-required` does **not** fire) and independently raises the
existing `core-citation-unresolved` warning (promotable to error by a severity
override, per ADR-008 sub-decision 5). One target therefore yields at most one
obligation verdict and, separately, its resolvability verdict — never a
`profile-` "required" error *and* a `core-` "unresolved" error for the same
missing-vs-present question. This is exactly ADR-008's model: "a profile that
requires a citation to *resolve* is a severity promotion of
`core-citation-unresolved`, not a new code," while `profile-citation-required` is
about **presence**, not resolvability.

**Carry-in №1 (`isIntraWiki` gate on class 3) — decided.** In
`internal/validate/links.go` `classify`, the repo-path class (class 3) is entered
**only** for targets that pass `isIntraWiki` (scheme-less, not protocol-relative
`//…`, not a bare `#fragment`, not empty). A target failing `isIntraWiki` can
never reach the repo-path `stat`: it is malformed by class 4. This makes the
"never stat above the anchor / never follow a non-intra-wiki shape" boundary an
explicit, test-guarded gate rather than an emergent property, and guarantees `//`
and `#` targets never trigger a filesystem read.

### 4. Evidence-context scoping and forbidden-target-type resolution

Evidence contexts are **profile-designated per page type** (`evidenceSections`),
so a `claim`'s `## Evidence` is a citation context while a `synthesis`'s
`## Findings` is not — only `claim` carries a citation obligation in addendum 003.
The engine selects a page's evidence sections by its frontmatter `type` when
splitting evidence contexts (ADR-008's `splitEvidenceContexts`); a page of a type
with no `evidenceSections` has no citation contexts and behaves exactly as today
(all links navigational). Forbidden-target-type checking resolves an intra-wiki
citation target to its bundle path (the shared resolver) and reads **that target
page's frontmatter `type`**; only in-bundle, resolvable targets are type-checked
(external/absent/malformed targets cannot be, and are governed by the
`core-citation-*` codes). This is a read of already-walked page frontmatter — no
new filesystem read beyond the ADR-008 repo-path `stat`.

### 5. Unknown page types stay OKF/core-accepted

A page whose `type` is **not** a profiled type in the resolved profile is **not**
an error: it receives the OKF and core-profile rules exactly as today and **no**
profile per-type rules. Addendum 003 does not restrict the page-type set, and the
core types `concept`/`entity` (and any author-invented type) must keep validating.
Profile per-type rules fire **only** for the types the profile declares.

### 6. Finding-code namespace

New profile-data-driven rules use the **generic `profile-*` prefix**, tagged
`ruleset: profile`, matching the `profile-citation-*` codes ADR-008 already
fixed. This yields a coherent three-part taxonomy: `okf-*` = OKF ruleset (engine);
`core-*` = engine-shipped **default** rules tagged `ruleset: profile`; `profile-*`
= **profile-data-driven** rules tagged `ruleset: profile`. The codes name the
**rule kind**, not the profile, because the rule kinds are engine capabilities any
profile reuses — a value the whole ADR-007 data model exists to provide. **This
supersedes the Phase-4 plan's provisional `academic-*` suggestion**, for
consistency with ADR-008's already-decided `profile-citation-*` naming and to keep
codes profile-agnostic; the divergence is recorded in `knowledge/log.md`.

## Consequences

- **Easier:** I2–I7 build against one settled schema and code set; golden parity
  for `core` is structural (the code is untouched), not a reproduction burden;
  addendum 003 is expressible with a closed vocabulary that cannot grow into a
  language; both ADR-008 carry-ins have concrete, testable semantics; and
  criterion 5's OKF-vs-profile split is preserved because every new code is
  `ruleset: profile`.
- **Harder:** two rule substrates coexist for the MVP (engine-coded core rules +
  data-driven profile rule kinds), so "everything is data" is not yet literally
  true for `core`; the engine must select evidence sections and per-type rules by
  frontmatter `type`, adding a type-keyed dispatch the pre-Phase-4 engine lacked;
  and forbidden-target-type checking needs the walked pages' frontmatter types
  available at evaluation time.
- **Maintain:** the closed profile vocabulary and its per-type rule kinds; the
  `profiles/core/profile.yaml` descriptive parent and the golden-parity invariant;
  the `profile-*` code family and its severities; the per-type evidence-section
  scoping; and the `isIntraWiki` class-3 gate.
- **Deferred / out of scope:** a fully-declarative `core` migration (separate
  golden-guarded issue); conditional-section syntax and richer field sets (Q4,
  still assumption-locked to addendum 003); deep profile hierarchies (ADR-007,
  one level only); local-path/custom profiles (Phase 7 seam). **ADR-numbering
  note:** the build-out plan provisionally reserved **ADR-010 for index
  maintenance** (Phase 5); `adr-alloc` assigned the next free number to *this*
  schema decision, so the index-maintenance ADR renumbers to the next free number
  when Phase 5 lands. **Ratification:** accepted under the 2026-07-03
  autonomous-phase mandate and **flagged for Oliver's async ratification**
  alongside ADR-006/007/008/009.
