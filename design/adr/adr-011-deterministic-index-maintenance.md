# ADR-011: Deterministic index maintenance (fenced generated regions, staged writes)

**Status:** accepted
**Date:** 2026-07-08
**Accepted:** 2026-07-08
**Accepted by:** Oliver (owner)

## Context

Phase 5 (Enrichment + index maintenance) restarts from the accepted build-out
plan after the July 2026 Phase 5–7 rollback and governance reset. Its exit gate
(`design/build-out-plan.md` §Phase 5) is acceptance **criterion 11 (enrichment)**
**plus index reliability**. Enrichment itself reuses accepted mechanics —
[ADR-006](adr-006-staged-mutation-transaction-model.md) (staged plan/apply),
[ADR-008](adr-008-provenance-and-citation-model.md) (citation preservation),
[ADR-001](adr-001-go-toolchain-and-yaml.md) (round-trip) — and needs no new ADR.
**Index maintenance does**: the engine must keep a reserved index page
deterministically current, and no accepted ADR yet fixes how. This is Phase 5
issue I1 (**#76**), the gate-blocking prerequisite: implementation issues I2–I7
may not file until it is accepted.

**What the product requires.** PRD **FR9 (Index maintenance)** fixes the
contract: indexes are *deterministic derived views from page metadata*, and
index operations must avoid duplicate entries, preserve clearly marked
human-maintained sections, use title and description where available, produce
stable ordering, avoid model calls, and support dry-run and diff output. PRD
**§14** adds the reliability/performance envelope: repeating an unchanged
deterministic operation produces no diff; output is stable for identical inputs;
and index regeneration on the 5,000-doc reference corpus completes **without a
model call**. The OKF base conformance rules reserve **`index.md`** and
**`log.md`** at the bundle root ("follow OKF rules when present"). The PRD
authoring flow's penultimate step is "**update the relevant index through the
engine**", and FR8 already defaults a **stale index** to **warning** severity
"unless promoted by the profile".

**What is already owned elsewhere and must not be redefined here.**

- [ADR-003](adr-003-json-contract-and-exit-codes.md) owns the versioned JSON
  envelope and the fixed exit-code buckets.
- [ADR-004](adr-004-validation-and-severity-model.md) fixed the three-severity
  model, set the **stale-index default to warning** (profile-promotable), and
  **explicitly deferred "index-consistency checks"** to the index-maintenance
  ADR (its Consequences say the interlock is owned by "ADR-010" — the number
  then provisionally reserved for index maintenance; see the numbering
  correction below).
- [ADR-005](adr-005-safe-filesystem-layer.md) owns per-file atomicity and
  path/symlink safety; [ADR-006](adr-006-staged-mutation-transaction-model.md)
  owns the staged-mutation lifecycle (stage → validate → journaled atomic
  commit, hash-bound stale-plan rejection).
- [ADR-010](adr-010-profile-data-schema-and-rule-vocabulary.md) owns the
  **closed profile vocabulary** and its `severities` override map.
- [ADR-009](adr-009-install-upgrade-uninstall-ownership.md) classifies
  `index.md` ownership for install/upgrade/uninstall.

**Numbering correction (required by issue #76).** The build-out plan and the
ADR list at its end provisionally reserved **"ADR-010 — Index maintenance"** for
this decision. That number was assigned by `adr-alloc` (sequential, never
reused) to the Phase-4 **profile-data schema and rule/citation vocabulary** ADR,
which is now accepted and ratified. Index maintenance therefore takes the next
free number, **ADR-011**, per the numbering note recorded in ADR-010's
Consequences and in `knowledge/log.md`. As a consequence, ADR-004's Consequences
sentence "index-consistency checks interlock with ADR-010" now points at **this
ADR (ADR-011)**; the reference is not re-edited in ADR-004 (accepted ADRs are
not edited in place), it is reconciled here. Likewise the plan's provisional
"ADR-011 (hooks + CI)" and "ADR-012 (custom profile)" will each renumber to the
next free number when their phases (6 and 7) land, via `adr-alloc`.

**Not an input.** The discarded Phase 5–7 direct-to-main commits (which included
an unreviewed index implementation) are explicitly **not** a base for this
decision, per issue #76's out-of-scope list. This ADR is drafted fresh from the
accepted plan and the PRD.

This ADR settles, as numbered sub-decisions: the index inventory; derivation
from page metadata and missing-metadata behavior; the human-section preservation
convention; stable ordering, duplicate policy, and byte-idempotency; stale-index
validation (finding code + default severity); the dry-run/diff and JSON
contract surface; the write path; the runtime constraints; and the criterion-11
relationship.

## Options considered

### Option A: Engine-generated fenced regions inside the reserved `index.md`, mutated as a standalone staged operation

- The engine maintains exactly one generated index — the OKF-reserved bundle-root
  `index.md` — replacing only the bytes inside an HTML-comment fence pair, and
  routes every index mutation through the ADR-006 staged transaction as its own
  operation (invoked by skills right after a page apply).
- Pros: satisfies FR9's *preserve marked human sections* directly (everything
  outside the fence is byte-preserved); reuses the repo's own `adr-index` fence
  idiom; adds **no** new write path, **no** new envelope fields, **no** new
  profile keys, and **no** new dependency; byte-idempotency and no-model-call are
  expressible as test-gated invariants; keeps each transaction atomic and lets
  `validate` catch a missed refresh because staleness is only a warning.
- Cons: fence parsing has genuine edge cases (absent / duplicate / unterminated /
  nested fences) that must be specified; byte-idempotency becomes a strict test
  gate including newline discipline; the decoupled write op means the tree and
  the index can be momentarily out of sync between a page apply and the index
  refresh (bounded by a warning-severity `validate` finding, not an error).

### Option B: Fully engine-owned index files (no human-editable sections)

- The engine owns `index.md` end to end and overwrites the whole file on
  regeneration.
- Pros: simplest possible generation and idempotency story — no fence parsing.
- Cons: **violates FR9's** "preserve clearly marked human-maintained sections"
  requirement outright; an author's hand-written preamble or curation notes in
  `index.md` would be destroyed on the next regen. Rejected.

### Option C: Fold the index update into every page-mutation transaction

- Each page plan/apply also stages `index.md`, so the tree and its index commit
  atomically together.
- Pros: the index is never even momentarily stale; one transaction covers both.
- Cons: `index.md` becomes a shared write target of *every* page plan, so its
  ADR-006 base hash is contended by concurrent or successive plans — a plan
  staged against one `index.md` base is rejected as stale the moment another
  plan touches the index first, multiplying stale-plan rejections. It converts a
  deliberately **warning-severity** condition (a briefly stale index) into
  **apply-blocking** coupling on unrelated page edits, and makes every page plan's
  diff include index churn. Rejected, with the trade-off recorded; this is the
  one genuinely contestable call and the coupled alternative is presented
  honestly for adversarial review.

A fourth alternative — **model-assisted index curation** — is noted only to be
rejected: it fails FR9 and §14 outright (index maintenance must avoid model calls
and be a deterministic derived view), so it is not developed as a full option.

## Decision

Adopt **Option A**. Index maintenance is a **deterministic derived view** written
into **fenced generated regions** of the OKF-reserved `index.md`, mutated only
through the **ADR-006 staged transaction** as a standalone operation. The
following sub-decisions are load-bearing and fixed here; each maps to an issue
#76 scope item.

### 1. Index inventory (scope item 1)

The MVP maintains **exactly one** generated index: the OKF-reserved
bundle-root **`index.md`** (`wiki/index.md` in the reference layout), with
per-type sections *inside* a single generated region. The **exact markdown
format** of that region (section headers, list structure, per-entry layout) is
deliberately **not** fixed here — an ADR fixes decisions, not byte formats — and
is left for the implementation issue(s) to define and lock behind fixtures (see
Consequences). There are **no** separate per-type index files and **no**
profile-specific index files in the MVP. This is
deferred, not forgotten: each additional generated file multiplies the staleness
and human-preservation surface, and a *profile-configurable* index view would
extend ADR-010's **closed** profile vocabulary, which by ADR-010's own rule
requires a new ADR. No new generated-file naming or location rules are needed as
a result. The other OKF-reserved file, **`log.md`, is explicitly out of scope** —
it is an append log, not a derived index, and its maintenance is a separate
concern.

### 2. Derivation from page metadata (scope item 2)

Index entries are derived **solely** from page frontmatter already parsed by the
engine — `type`, `title`, `description` — plus the **bundle-relative path**
(taken relative to the same bundle boundary [ADR-005](adr-005-safe-filesystem-layer.md)
uses for path/symlink safety and [ADR-006](adr-006-staged-mutation-transaction-model.md)
uses for its staging root — this ADR defines no new boundary), read through the
existing ADR-001 `yamladapter`. There is **no new parser and no body scraping**. Missing-metadata behavior is fixed and total:

- **`title`** absent or empty → fall back to the **filename stem** (so every
  indexed page has a display label).
- **`description`** absent or empty → **omit** it (no placeholder text).
- A page whose YAML **fails to parse** is **excluded** from the generated region.
  Its parse failure is already an ADR-004 hard error surfaced by `validate`, and
  excluding it keeps regeneration total and deterministic rather than emitting a
  half-entry.

"Uses title and description where available" (FR9) is thereby satisfied exactly:
both are used when present, and each has a defined fallback when not.

### 3. Human-maintained section preservation (scope item 3)

Preservation uses an **HTML-comment fence** convention, deliberately the same
pattern as the repo's own `adr-index` table:

```
<!-- llm-wiki:index:start -->
... engine-generated entries (the only bytes the engine may rewrite) ...
<!-- llm-wiki:index:end -->
```

The engine may replace **only** the bytes strictly *inside* one fence pair.
Everything outside the fence — any human-authored preamble, notes, or curation —
is **byte-preserved**. Edge cases are fixed:

- **Fences absent:** the engine **never silently rewrites** the file. Dry-run
  shows an *append-a-fenced-region-at-end* proposal; it applies only through the
  staged path (sub-decision 7), never as an in-place overwrite. The
  dry-run/staged-plan preview shows the **full file diff** (not just the appended
  region), so if the end-of-file placement would land after human-authored
  trailing content the user does not want the region below, the user can **abort
  before apply** and reposition a fence pair by hand first.
- **Duplicate, nested, or unterminated fences:** the engine **refuses** to mutate
  and emits a finding (sub-decision 5); it never guesses which region is
  authoritative.
- **`init`** scaffolds `index.md` with the fence pair already present, so the
  common path starts in a managed state.

### 4. Stable ordering, duplicates, byte-idempotency (scope item 4)

- **Duplicates are structurally impossible:** entries are keyed on the
  **bundle-relative path** (the ADR-005/ADR-006 bundle boundary, as defined in
  sub-decision 2), one entry per page, so FR9's "avoid duplicate entries" holds
  by construction rather than by a de-dup pass.
- **Sort key:** `type` (bytewise ascending), then bundle-relative path (bytewise
  ascending). Titles are **display-only and never sort keys**, avoiding
  locale/case-fold nondeterminism. Ordering is thus stable and
  platform-independent.
- **Byte-idempotency (test-gated invariant):** regenerating twice over an
  unchanged tree produces a **byte-identical** file, and regenerating an
  already-current index produces an **empty diff** (PRD §14 "unchanged operation →
  no diff"). The generated region uses **LF line endings** regardless of host
  platform, so idempotency does not depend on the writer's OS.

### 5. Stale-index validation (scope item 5)

`validate` gains an engine-shipped finding **`core-index-stale`**, tagged
**`ruleset: profile`** (matching ADR-010's taxonomy for engine-shipped default
rules), default severity **warning** per FR8 / ADR-004, **promotable to error**
via ADR-010's existing `severities` profile map — **no new severity machinery**.
"Stale" is defined precisely: the **current fenced region's bytes differ from the
freshly recomputed region**. A companion finding **`core-index-unmanaged`**
covers an index file that cannot be maintained (missing / duplicate / nested /
unterminated fences, per sub-decision 3), same **warning** default. `validate`
**reports only**; it never mutates the index. This discharges ADR-004's deferred
"index-consistency checks" interlock.

### 6. Dry-run / diff and JSON contract (scope item 6)

The `llm-wiki index` command surface is fixed as:

- A **non-mutating** check/diff mode: `--dry-run` shows the unified diff of the
  fenced region between its current and recomputed bytes, and exits without
  writing.
- `--json` uses the **ADR-003 v1 envelope** and the existing exit-code buckets
  **unchanged**. The **only** contract delta is a new `operation` value
  identifying the index operation; this stays inside ADR-003's envelope shape —
  **no new fields, no version bump**.

### 7. Write path (scope item 7)

Index mutation is a **standalone staged operation** routed through the
**ADR-006** transaction (stage → validate → journaled atomic commit), invoked by
the enrichment/authoring skills immediately after a page apply — realizing the
PRD authoring flow's "update the relevant index through the engine" step —
**rather than folding `index.md` into every page-mutation plan**. Rationale
(the contestable call, stated for review): including `index.md` in every page
plan makes concurrent and successive plans contend on `index.md`'s base hash and
multiplies ADR-006 stale-plan rejections on otherwise-independent page edits,
while a briefly stale index is only a **warning**-severity condition. Decoupling
keeps each transaction atomic, keeps page diffs free of index churn, and relies
on `validate` (`core-index-stale`) to catch any missed refresh — self-healing via
a subsequent regen. The index operation itself is a normal ADR-006 transaction:
per-file atomic commit, hash-bound stale-plan rejection, journaled recovery — it
**composes** ADR-005/006 and adds nothing to them.

### 8. Runtime constraints (scope item 8)

- **No model calls anywhere** in index maintenance — a hard, test-visible
  invariant (FR9, §14).
- Implementation lives in a **new `internal/index` package** using the **standard
  library plus the existing `yamladapter` only** — **no new runtime dependency**
  (per the overlay rule: a new runtime dependency requires its own ADR).
- This meets PRD §14: full regeneration over the 5,000-doc reference corpus
  completes without a model call, with stable output for identical input.

### 9. Relationship to Phase 5 acceptance criterion 11 (scope item 9)

`index.md` **is an existing page**, so every index mutation is previewable before
apply, satisfying criterion 11 ("existing-page edits previewed before apply") for
the index exactly as for any enriched page: `--dry-run`/diff in direct CLI use,
and the ADR-006 staged-plan preview in skill flows. **Enrichment itself adds no
new decision** here — it reuses ADR-006 (staged path), ADR-008 (citation
preservation), and ADR-001 (round-trip) exactly as issue #76 states. ADR-011
records this explicitly so Phase 5 needs no second ADR. Should implementation
uncover a genuinely new enrichment decision boundary, the correct move is to
**stop and raise a new ADR**, not to widen this one.

## Consequences

- **Easier:** Phase 5 issues I2–I7 build against one settled surface — a single
  generated file, a fixed fence convention, a defined sort key, one finding code
  and its default, and clear no-model-call / no-new-dependency constraints;
  staleness is **observable** (`core-index-stale`), **promotable** (ADR-010
  `severities`), and **self-healing** (a subsequent regen). Criterion 11 is met
  for the index with no new mechanism.
- **Harder:** fence parsing must handle absent / duplicate / nested /
  unterminated cases explicitly; byte-idempotency is a strict test gate that
  includes LF newline discipline for the generated region; the decoupled write op
  means the tree and index can be momentarily out of sync between a page apply and
  the index refresh (bounded to a warning, caught by `validate`).
- **Maintain:** the fence names (`llm-wiki:index:start`/`:end`), the sort key,
  the `core-index-stale` and `core-index-unmanaged` codes and their warning
  defaults, and the no-model-call / stdlib-only / no-new-dependency constraints.
- **Implementation must define and fixture the generated-region format.** This
  ADR fixes *what* is derived and *how it is bounded and ordered*, but not the
  concrete markdown of the region (section headers, list/entry layout). The
  implementation issue(s) must **specify that format and lock it behind golden
  fixtures**, since byte-idempotency (sub-decision 4) and stale detection
  (sub-decision 5) both compare against exactly those bytes — an unpinned format
  would make both invariants untestable.
- **Deferred / out of scope:** per-type and profile-specific index files;
  a profile-configurable index vocabulary (would extend ADR-010's closed
  vocabulary → new ADR); sub-bundle indexes; and `log.md` maintenance (not an
  index).
- **Boundaries referenced, not redefined:** ADR-003 (envelope — one new
  `operation` value only), ADR-004 (severity model — this ADR supplies the code
  and default into it, discharging its deferred index-consistency interlock),
  ADR-005/006 (write path — composed, not extended), ADR-008 (citations — index
  links are navigational, never citation contexts), ADR-010 (closed profile
  vocabulary — no new keys; severity promotion uses the existing `severities`
  map), ADR-001/002/009 (toolchain / self-contained binary / `index.md` ownership).
- **Numbering-correction note:** provisional build-out-plan "ADR-010 — index
  maintenance" is **this** ADR, renumbered to **ADR-011** by `adr-alloc` after
  the number went to the Phase-4 schema ADR; ADR-004's "interlock with ADR-010"
  reference is reconciled to ADR-011 here (ADR-004 is not edited in place). The
  plan's provisional ADR-011 (hooks + CI) and ADR-012 (custom profile) likewise
  renumber at Phases 6 and 7 via `adr-alloc`.
- **Status / acceptance:** drafted **`proposed`**. Acceptance follows an
  independent cross-vendor adversarial review reaching `READY` (the reviewer for
  this ADR is **Qwen3.7 Max via OpenRouter**, substituted for OpenAI Codex per
  the overlay's equal-or-higher, non-Anthropic reviewer rule and recorded in
  `knowledge/log.md`) **plus Oliver's explicit acceptance**; only then does the
  status flip to `accepted`, in the same PR, before merge. This keeps
  ratification debt at 0 and unblocks I2–I7 at merge.
