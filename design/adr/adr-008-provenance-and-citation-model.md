# ADR-008: Provenance and citation model (context-based, deterministic offline resolution)

**Status:** accepted
**Date:** 2026-07-03

## Context

Phase 3 (Authoring + staged mutation) must let the engine and its authoring
adapter make **sourced claims that carry resolvable citations**, and must
**preserve existing citations** through the staged edit lifecycle. PRD **FR5**
requires that a sourced claim cite a **resolvable URL, repository path, or OKF
document**, that existing citations be preserved, that **evidence be
distinguished from inference**, and that the engine **never invent a citation**;
it also fixes that the `resource` field is **asset identity, not a citation
substitute**. Acceptance **criterion 9** ("model-generated sourced claims
contain resolvable citations") is a **100%-within-fixtures** gate — the
fleet-wide rate is measurement-only (addendum 001), not a threshold. MVP
principle 6 ("provenance without theatre") forbids a citation mechanism that
looks rigorous but adds ceremony a human reader cannot follow.

What is already decided and constrains this ADR:

- **[ADR-004](adr-004-validation-and-severity-model.md)** (accepted) fixed the
  single OKF-vs-profile validation engine, the three severities (error /
  warning / suggestion), and the resolution precedence **core defaults →
  profile overrides → baseline suppression** with an inherited parse-failure
  gate. It explicitly deferred the citation rules to this ADR: "provenance/
  citation rules (criterion 9) interlock but are specified in ADR-008"
  ([adr-004](adr-004-validation-and-severity-model.md), Consequences). Citation
  rules must therefore land as **findings inside that existing engine**, not as
  a parallel checker.
- **[ADR-006](adr-006-staged-mutation-transaction-model.md)** (accepted) owns
  staged mutation mechanics (`plan`/`apply`), byte-level preservation, and
  **stale-plan rejection** — its only `apply` refusal; it has **no
  content-approval gate**. It stated "the provenance/citation-**preservation
  content rules** are out of scope here" — delegated to this ADR. The plan-time
  citation-loss **content gate is therefore ADR-008-owned**, not an ADR-006
  mechanism; and the block it triggers routes through the existing
  **[ADR-003](adr-003-json-contract-and-exit-codes.md) `approval` member**, not
  a new commit mechanism.
- **[ADR-003](adr-003-json-contract-and-exit-codes.md)** (accepted) froze the v1
  envelope `{contractVersion, operation, status, findings, affectedPaths,
  approval}`. ADR-008 may add **rule codes** inside `findings` and use the
  existing `approval` member; it introduces **no new public contract**.
- **[ADR-007](adr-007-profile-system-boundary.md)** (accepted) made profiles
  **declarative data** with one inheritance level. Per-type citation rules must
  be **expressible as profile data**, not engine code.
- **[ADR-005](adr-005-safe-filesystem-layer.md)** (accepted) fixed the mandatory
  write-gate and its canonicalize/resolve-symlink boundary.
  **[ADR-009](adr-009-install-upgrade-uninstall-ownership.md)** (accepted) fixed
  `.llm-wiki/` as the engine-metadata anchor. A citation resolver that touches
  the filesystem must stay inside those boundaries.
- **Shipped code sets the grammar.** `internal/validate/links.go` already parses
  **exactly inline Markdown links** `[label](target)` (reference-style links,
  autolinks, and raw HTML are deliberately deferred — `links.go:11-16`), skips
  **image** links `![alt](t)`, and resolves in-bundle targets through
  `resolveTarget` (strip `#fragment`/`?query`, strip a leading `/` as
  bundle-root-absolute, `path.Clean`, and treat any `../` escape as
  **unresolved**). It emits a single warning-severity `core-broken-link` finding
  per page. ADR-008 must **reuse this resolver**, not fork it.
- **[Addendum 003](../prd-addenda/003-academic-research-profile-contract.md)**
  fixes the minimum `academic-research` rules the mechanism must be able to
  instantiate in Phase 4: `assessment: supported` with zero citations in
  `## Evidence` = error; a `question` page is never a valid citation target =
  error; a `source` with neither `doi` nor `canonical_url` = suggestion.

This ADR settles: (1) how a citation is **represented**; (2) what "**resolvable**"
means per target class, **deterministically and offline**; (3) **who asserts**
that a claim is sourced; (4) the **core-vs-profile rule split, severities, and
new rule codes**; (5) how existing citations are **preserved** through
`plan`/`apply`. It does **not** cover supply-chain provenance (release signing /
attestation) — that is a **separate deferred ADR** (see Consequences), and the
two senses of "provenance" must not be conflated.

## Options considered

### Option A: Context-based citations — ordinary inline links, evidentiary status from profile-designated context, one shared deterministic offline resolver

- A citation is an **ordinary inline Markdown link** `[label](target)` in the
  page body — the exact grammar `links.go` already parses. Whether that link is
  a **citation** (evidence for a claim) versus a **navigational** link comes
  from **context**: a profile-designated **evidence context** (e.g. addendum
  003's `## Evidence` section under a supported claim), or the authoring
  adapter's placement. No new syntax, no frontmatter registry, no parser.
- "Resolvable" is a **three-class deterministic, offline** predicate sharing one
  resolution table with the existing broken-link resolver (below): http(s) URL,
  in-bundle OKF document, repo-relative path.
- Sourced-ness is asserted **structurally** by declarative profile rules (the
  engine enforces what the page claims about itself) plus **procedurally** by the
  adapter contract (proved on criterion-9 fixtures) — the engine never *infers*
  authorship.
- Preservation is ADR-001/ADR-006 byte mechanics plus **one** plan-time
  citation-loss finding routed through the **existing `approval` member**.
- Pros: zero new public surface — grammar, resolver, engine, envelope, and
  profile-data model are all already shipped or accepted; citations stay legible
  in plain Markdown and on GitHub (principle 6); claim-and-evidence adjacency
  that FR5 requires is preserved because the evidence lives next to the claim;
  section-scoped profile rules (addendum 003) are naturally expressible because
  status is contextual; the `resource` field stays asset identity, never a
  citation (FR5). Offline determinism satisfies criterion 9 without a network.
- Cons: "is this link a citation?" is a property of **context**, so the
  engine needs a declarative way for a profile to mark evidence contexts (new
  profile vocabulary, though only *data*); a bare inline link carries no
  machine-readable "this is a citation" tag, so enumerating a page's citations
  means knowing its profile's evidence contexts, not reading one frontmatter
  list.

### Option B: Frontmatter `citations:` registry

- Each page lists its citations in a machine-enumerable frontmatter block
  alongside `resource`.
- Pros: trivially enumerable; a citation is a first-class typed field.
- Cons: a flat frontmatter list **cannot express section-scoped rules** —
  addendum 003's "`assessment: supported` needs a citation **in `## Evidence`**"
  is inherently positional, and a page-level list loses the claim-to-evidence
  adjacency FR5 exists to protect; it **severs evidence from the sentence it
  supports**; and it adds a second citation-like schema **next to `resource`**,
  which FR5 explicitly keeps distinct. It also duplicates link data the body
  already contains, inviting drift between the list and the prose.

### Option C: Dedicated citation micro-syntax (footnotes / shortcodes)

- A bespoke footnote (`[^cite]`) or shortcode (`cite:...`) grammar marks
  citations unambiguously.
- Pros: a citation is lexically unmistakable; enumeration is a parse.
- Cons: needs a **new parser the MVP deferred** (`links.go` deliberately
  handles only inline links); degrades human and GitHub legibility, which is
  exactly the "citation theatre" MVP principle 6 rejects; and a private syntax
  is dead in every downstream Markdown viewer, contradicting the offline,
  plain-file ethos.

## Decision

Adopt **Option A**. A citation is an ordinary inline Markdown link; evidentiary
status comes from profile-designated context; resolution is a three-class
deterministic offline predicate shared with the broken-link resolver;
sourced-ness is asserted structurally (profile data) and procedurally (adapter
contract); severities live inside ADR-004's precedence; preservation is
ADR-001/ADR-006 mechanics plus a plan-time citation-loss finding routed through
the existing `approval` member; only new rule codes are added, with **no
contract bump**. Options B and C are rejected: B cannot express the
section-scoped, adjacency-preserving rules addendum 003 and FR5 require and
duplicates `resource`; C needs a deferred parser and violates the
"provenance without theatre" bar.

The following sub-decisions are load-bearing and are fixed here.

**1. Citation representation and what is *not* a citation.** A citation is the
inline-link grammar `links.go` already parses: `[label](target)`, with the
`target` capture trimmed. A link is a **citation** only when it appears in a
**profile-designated evidence context**; the same grammar elsewhere is a
**navigational** link, governed by the unchanged `core-broken-link` rule. The
following are **never** citations: **image** links `![alt](t)` (assets, skipped
exactly as today); the **`resource`** frontmatter field (asset identity per
FR5); and any link **outside** an evidence context. Reference-style links,
autolinks, and raw HTML remain out of scope for the MVP, inheriting
`links.go`'s existing deferral — a page that needs a citation writes it as an
inline link. The engine **never invents** a citation (FR5): absence of a
required citation is *reported*, never *filled*.

**2. "Resolvable" per target class — deterministic, offline, and total.** Inside
an evidence context every inline-link target is a citation target (sub-decision
1), so the classification must be **total over the shipped inline-link grammar**
and checked without any network I/O. A target is classified by the following
**ordered, deterministic** procedure; the first matching class wins, and the
final catch-all guarantees no target is left without a verdict:

1. **http(s) URL** (target begins with a case-insensitive `http:`/`https:`
   scheme, per the shipped `uriScheme` regex): **resolvable iff it is
   syntactically a valid absolute URL with a non-empty host.** An http(s) target
   that is **syntactically invalid** (e.g. `https://` with an empty host) is
   **malformed** (`core-citation-malformed`), *not* unresolved — it can never
   resolve. Liveness is **never fetched and never gated** — reachability is
   measurement-only (addendum 001), consistent with `links.go` never fetching
   external targets.
2. **In-bundle OKF document** (a scheme-less intra-wiki reference, per
   `isIntraWiki`, whose `resolveTarget` cleaned path stays **within the bundle
   root**): **resolvable iff that cleaned path is a member of the bundle file
   set** (`exists`), reusing the shipped semantics verbatim — `#fragment`/
   `?query` stripped, leading `/` treated as bundle-root-absolute, `path.Clean`
   applied. A cleaned path inside the bundle that is **absent** is
   `core-citation-unresolved`.
3. **Repo-relative path** (a scheme-less reference that, via the sub-decision-3
   `../` refinement, resolves **outside** the bundle but **inside** the
   repository root): **resolvable iff a read-only existence check (`stat`) inside
   the repo root succeeds**; a well-formed repo-path that is absent is
   `core-citation-unresolved`.
4. **Malformed catch-all** (`core-citation-malformed`) — every remaining target:
   **empty**; a **non-http(s) URI scheme** (e.g. `mailto:`, `ftp:`, a private
   `repo:`-style scheme); a **fragment-only** target (`#sec`) or a
   **protocol-relative** target (`//host/p`) — neither of which `isIntraWiki`
   accepts and neither of which names a resolvable citation; and any
   scheme-less path that **escapes the repo root** via `../`. Malformed targets
   can never resolve.

The **malformed-vs-unresolved** split is load-bearing: *malformed* (classes 1
invalid-host and 4) can never resolve regardless of repository state, while
*unresolved* (classes 2 and 3, well-formed but absent) may become resolvable
when the target is created. This is exactly what lets a profile promote
"malformed" to error while leaving "unresolved" a warning, and it makes the
partition total — no citation target falls between the two codes.

**3. Repo-path class mechanics, read scope, and the `../` refinement.** The
repository root is anchored at the **nearest ancestor directory containing
`.llm-wiki/`** (the ADR-009 engine-metadata marker); when no such ancestor
exists, the anchor **falls back to the bundle root** and the repo-path class is
**empty** (every scheme-less target is either in-bundle or an escape). Within an
anchored repo root, this ADR **refines** `links.go`'s current blanket rule that
*any* `../` is unresolved: a `../` sequence that **stays inside the repo root**
now resolves against the repo-path class, while a sequence that **escapes the
repo root remains unresolved/malformed**. This refinement is applied **uniformly
through one resolution table** — navigational links get the same widened
resolution, so there is exactly one resolver and one notion of "escape," not two.
The repo-path check is a **read-only `stat`**, never a read of contents, and is
**never performed above the anchor**. It introduces the first validation-time
filesystem read (`links.go` today takes membership through an injected `exists
func(string)` seam and touches no disk; `links.go:36,48`), so it must run behind
the **same injection seam** and through the **same canonicalize / resolve-symlink
primitives** the ADR-005 write-gate uses: a target that canonicalizes outside the
repo root is an **escape** (malformed), never followed. Note that ADR-005 is a
**write** gate — reads are outside its stated scope — so this read path is not
*governed* by ADR-005 but is **bounded by the same repo-root anchor and
symlink-resolution rules**, staying strictly inside the write boundary it can
never exceed. The rejected
alternative — a private `repo:` URI prefix to keep the classes lexically
disjoint without touching `../` — is declined because it invents a scheme that
is dead in every Markdown viewer (a principle-6 violation) and still needs the
same repo-root read check; refining `../` keeps citations as ordinary links.

**4. Who asserts a claim is "sourced" / "model-generated."** The engine **never
infers** authorship or intent. Sourced-ness is decided along two independent
seams:

- **Structural (enforced by the engine over declared data).** A profile
  declares, as data, which sections/fields are **evidence contexts** and which
  claims **oblige** a citation; the engine enforces **what the page states about
  itself**. Addendum 003's rule ("`assessment: supported` ⇒ a citation in
  `## Evidence`") is the archetype: the page *declares* the claim supported, so
  the obligation is decidable from the page alone. Human-authored prose that
  makes no such declaration carries **no** citation obligation — "human-authored
  may remain uncited unless the profile says so."
- **Procedural (bound by contract, proved on fixtures).** FR3/FR5 bind the
  **authoring adapter**: a model-generated sourced claim must arrive already
  cited. Criterion 9 is **fixture-scoped by design** (100% within the acceptance
  fixtures), so "model-generated" is decidable in exactly the place the gate
  measures it — the adapter's own output fixtures — without the engine having to
  classify arbitrary prose at large.

**5. Core-vs-profile rule split, severities, new codes, and the profile
vocabulary sketch.** All citation findings resolve severity inside ADR-004's
fixed precedence (**core defaults → profile overrides → baseline suppression**),
inherit the **parse-failure gate** (malformed YAML fails before any citation
finding is computed), and aggregate **one problem into one finding** keyed by
the `{ruleset, code, path}` baseline fingerprint.

**Ruleset tag (decided here, reconciled with shipped reality).** The contract
`Ruleset` enum has exactly two values — `okf` and `profile`
(`internal/contract/envelope.go:38-39`); **there is no "core" ruleset**. The
`core-` code prefix denotes an **engine-shipped default rule** (rule origin and
naming), **not** a ruleset tag. Every new `core-citation-*` and
`profile-citation-*` code is tagged **`ruleset: profile`**, matching the shipped
`core-broken-link` finding — which is itself emitted as `RulesetProfile`
(`internal/validate/links.go:61`) and which the citation rules share a resolver
with and **subsume** in evidence contexts (sub-decision 7). Tagging citations
`profile` keeps the criterion-5 OKF-vs-profile split coherent (citation
obligations arise from **profile-designated** evidence contexts, not universal
OKF structure) and guarantees the subsuming and subsumed findings sit in the
**same** ruleset, so a single target never yields one OKF and one profile
finding.

*Core mechanism (engine-shipped default rules; new codes, `ruleset: profile`):*

| Condition | Code | Default severity |
|---|---|---|
| Citation target malformed (invalid-host http(s) / empty / non-http(s) scheme / fragment-only / protocol-relative / repo-root escape) | `core-citation-malformed` | **warning** (profile-promotable to error) |
| Citation target well-formed but unresolved in an evidence context | `core-citation-unresolved` | **warning** (profile-promotable to error) |
| Same normalized target cited twice in one evidence context | `core-citation-duplicate` | **suggestion** |
| Net loss of an existing citation at plan time (sub-decision 6) | `core-citation-loss` | **warning** (apply-gated via `approval`, sub-decision 6) |
| Unresolved link in a **navigational** (non-evidence) context | `core-broken-link` (unchanged) | **warning** |

*Profile rules (Phase 4 data — vocabulary fixed now, keys not frozen;
`ruleset: profile`):*

| Condition | Code | Severity |
|---|---|---|
| A claim that obliges a citation has none | `profile-citation-required` | **error** |
| Citation points at a forbidden target type (e.g. a `question` page) | `profile-citation-target-type` | **error** |

A profile that requires a citation to *resolve* does **not** introduce a new
code: it is expressed as a **severity promotion of `core-citation-unresolved`**
(and, where wanted, `core-citation-malformed`) from warning to **error** within
the ADR-004 configurable set — same code, promoted severity — so an unresolved
required citation is **one** finding, never a `core-` warning and a separate
`profile-` error for the same target. This is the ADR-004 promotion mechanism
(a severity change on the same code), not a second finding.

The **minimum declarative vocabulary** a profile needs to express the Phase-4
rules — *sketched here, not frozen* — is: (a) which sections/fields are
**evidence contexts**; (b) a **require-citation-when** condition over declared
fields (addendum 003's `assessment: supported`); (c) a set of **forbidden target
types** (page types that may never be an evidence target); and (d) a **per-rule
severity** override within the ADR-004 set. Exact YAML keys are **deferred to
the Phase 4 profile issue**, mirroring ADR-003's deferral of the numeric
exit-code values: the mechanism is decided now; the surface syntax is fixed when
the `academic-research` profile lands, so addendum 003 can still be superseded by
Oliver without reopening this ADR.

**6. Preservation: byte mechanics plus an ADR-008-owned plan-time citation-loss
gate.** Byte-level preservation of frontmatter and body stays exactly where it
is — ADR-001's node-aware YAML round-trip and ADR-006's staged transaction
(unknown fields preserved, comments best-effort). ADR-006's only `apply` refusal
is **stale-plan rejection** (base hashes changed since plan;
[adr-006](adr-006-staged-mutation-transaction-model.md) Decision); it has **no**
content-approval gate. ADR-008 therefore **owns** a new apply-time content rule
and does not attribute it to ADR-006. At **plan time**, the engine computes, per
page and per evidence context, the set of **normalized citation targets** in the
source and in the staged result; a **net loss** (a normalized target present
before and absent after) emits a `core-citation-loss` finding. **The apply-time
block is carried by the ADR-003 `approval` member, not by the warning
severity:** a plan containing an un-approved `core-citation-loss` finding sets
the envelope's existing **`approval`** member, and `apply` returns
**approval-required** (`StatusApprovalRequired` / `ExitApprovalRequired`,
`internal/contract/{envelope,exitcode}.go`) and **refuses to commit** until the
loss is explicitly approved — so the `core-citation-loss` warning is *reported*
non-failingly per ADR-004, while the separate `approval` member is what gates the
commit (no ADR-004 contradiction, and **no envelope change** — the member
already exists per ADR-003). This makes FR4/FR5 "preserve existing citations"
*enforceable* rather than aspirational: a rewrite that silently drops a source is
surfaced and gated at `apply`, while a deliberate, approved removal still goes
through. **Normalization** for this equality (and
for `core-citation-duplicate`) is defined per class so it is deterministic: for
**in-bundle** and **repo-path** targets, the `resolveTarget`-cleaned path
(`#fragment`/`?query` stripped, leading `/` normalized, `path.Clean`); for
**http(s) URL** targets, the **trimmed target string compared verbatim** —
scheme detection is case-insensitive (the shipped `uriScheme` regex), but the
target is otherwise compared byte-for-byte with no network-aware canonicalization
(no host case-folding, no fragment stripping) — offline determinism over URL
cleverness.

**7. Finding overlap with `core-broken-link` (one resolver, one finding).**
There is **one** resolver shared by navigational links and citations. A target
is evaluated **once**; its **context** decides which rule owns it. In an
**evidence context**, the citation rule **subsumes** the broken-link finding for
that target — a single unresolved citation yields `core-citation-unresolved`,
**not** an additional `core-broken-link`, so one bad link is never two findings.
Outside evidence contexts, behavior is **unchanged**: `core-broken-link` fires
exactly as it does today. New codes follow the shipped `core-*` naming
convention and carry `ruleset: profile` (sub-decision 5) — the same ruleset as
the `core-broken-link` they subsume — so the subsuming and subsumed findings are
always in one ruleset and a single target is never double-counted across the
OKF/profile split.

## Consequences

- **Easier:** criterion 9 becomes a deterministic, offline, fixture-provable
  gate with no network dependency; citations stay legible plain-Markdown links
  (principle 6); the claim-and-evidence adjacency FR5 wants falls out of the
  representation; addendum 003's section-scoped rules are expressible as profile
  data with no engine fork; "preserve existing citations" (FR4/FR5) is
  enforceable through the **existing** approval path; and the whole model reuses
  the shipped resolver, the ADR-004 engine, the ADR-003 envelope, and the
  ADR-007 data model — **no new public contract**.
- **Harder:** the engine must carry a notion of **evidence context** and a
  shared three-class resolver with an anchored repo root and the read-only
  `../` refinement; the profile vocabulary (evidence contexts, require-when,
  forbidden types, per-rule severity) must be designed as data in Phase 4; the
  plan-time citation-loss diff and per-class normalization need tests, including
  the duplicate and loss edge cases and the `core-broken-link` subsumption; the
  `../` refinement is a **deliberate, uniformly-applied change to shipped
  `core-broken-link` behavior** and must be regression-tested so navigational
  resolution widens consistently, not divergently.
- **Maintain:** the three-class resolution table and its repo-root anchor
  (`.llm-wiki/` marker, bundle-root fallback); the new `core-citation-*` codes
  and the Phase-4 `profile-citation-*` codes; the per-class normalization rule;
  the evidence-context profile vocabulary once its keys are frozen; and the
  single-resolver invariant (one target, one finding).
- **Deferred / out of scope:** implementing `page inspect`/`plan`/`apply`, the
  authoring adapters, and the `academic-research` profile are **Phase 3/4**
  work — this ADR fixes the model they consume, not their code. Phase 3 consumes
  the **core mechanism** (representation, resolver, severities, citation-loss
  approval) and the **criterion-9 fixtures**; **Phase 4** consumes the **profile
  citation vocabulary** and instantiates addendum 003. The exact profile YAML
  keys are deferred to the Phase 4 issue (ADR-003-style). **Supply-chain
  provenance — release signing and provenance attestation — is explicitly NOT
  this ADR**: it remains re-deferred to a dedicated supply-chain ADR (per
  ADR-002/ADR-009), and the residual supply-chain risk in
  [`knowledge/risks.md`](../../knowledge/risks.md) stays `open` independent of
  this decision. The "sourced claims lose provenance" risk stays `open` — its
  **mechanism is decided here**, but it retires only when Phase 3/4 implements
  and fixture-proves it. Validation hook: **criterion 9**, interlocking with
  ADR-004 (severity engine) and ADR-006 (staged approval).
