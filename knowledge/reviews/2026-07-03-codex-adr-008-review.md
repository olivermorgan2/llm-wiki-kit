# Codex ADR-008 review — 2026-07-03

Scope: adversarial ADR gate for `docs/adr-008-provenance-citation-model`
(PR TBD, issue #32) — the Phase 3 provenance/citation model prerequisite.

Reviewer: Codex via Hermes `openai-codex` provider fallback (standalone codex
CLI unavailable — `codex-companion.mjs setup` reports `codex: not found`, same
condition as the PR #22 round). The gate was run as an independent adversarial
ADR-review pass with no stake in the draft, holding ADR-008 against the accepted
ADR-003/004/005/006/007/009 foundation, PRD FR5 / criterion 9, addendum 003, and
the shipped `internal/validate/links.go` + `internal/contract` code.

---

1. Verdict: NEEDS_REVISION — Score 3/5

The draft is strong on scope discipline (no new contract, profile rules as
ADR-007 data with keys deferred, supply-chain kept as a separate deferred ADR),
grammar reuse, repo-root anchoring, and the "who asserts sourced" seam. But four
decision-level gaps remain, two of which undercut load-bearing claims
(preservation enforceability and offline resolvability). Each is correctable
with a contained edit.

2. Blocking findings

Blocking finding B1 — Preservation enforcement is grounded in an ADR-006
discipline that does not exist (sub-decision 6, adr-008:268-270).
- The ADR stated the citation-loss gate reuses "the same deletions-require-
  approval discipline ADR-006 already applies to destructive commits." ADR-006
  applies no such discipline: its only `apply` refusal is for **stale** plans
  whose base hashes changed (adr-006:102, "`apply` refuses any plan whose
  recorded base hashes no longer match"). No accepted ADR defines that `apply`
  consults the envelope's `approval` member and refuses an *unapproved* plan for
  a content reason. ADR-003 supplies the `approval` member
  (`internal/contract/envelope.go:83`) and `ExitApprovalRequired`
  (`internal/contract/exitcode.go:18`) but not the apply-time enforcement.
- Required fix: ADR-008 must **own** the apply-time rule — a staged plan
  carrying an unapproved `core-citation-loss` sets the `approval` member and
  `apply` returns approval-required / refuses to commit — grounding it in
  ADR-003's approval member + `ExitApprovalRequired`, and stop attributing an
  approval-on-deletion discipline to ADR-006 (which gates only on stale hashes).

Blocking finding B2 — The three "syntactically disjoint" resolvable classes do
not partition the grammar's target space, so "resolvable" is undefined for some
evidence-context citations (sub-decision 2, adr-008:158-181).
- Sub-decision 1 makes *every* inline link inside an evidence context a
  citation, so the full `links.go` target space applies. But the classes cover
  only http(s), scheme-less in-bundle (`isIntraWiki`, which excludes `#…` and
  `//…` — links.go:73-78), and scheme-less repo-path. A **fragment-only** target
  (`[x](#sec)`) and a **protocol-relative** target (`[x](//host/p)`) match no
  class, and neither is in the malformed enumeration — their finding is
  undefined. Likewise a **syntactically invalid http(s) URL** (e.g. `https://`
  with empty host) is classified http(s), fails "valid absolute URL with a
  non-empty host," yet is neither a "non-http(s) scheme" malformed case nor a
  "well-formed but not present" unresolved case — it falls between the two codes
  whose split the ADR calls load-bearing (adr-008:178-181).
- Required fix: make the partition total — define a deterministic catch-all so
  every non-classifiable / syntactically-invalid citation target is
  `core-citation-malformed` (or explicitly declare such targets non-citations),
  and state that an invalid-host http(s) URL is malformed, not unresolved.

Blocking finding B3 — `profile-citation-unresolved` is a second code for a
condition ADR-004 models as a severity *promotion*, risking two findings for one
problem (sub-decision 5, adr-008:235 vs 245).
- The profile table listed `profile-citation-unresolved` (error) described as
  "profile promotion of the core warning," alongside the core
  `core-citation-unresolved` (warning). ADR-004 promotion is a severity change
  **on the same code** within the configurable set (adr-004:56-60), not a new
  code. As written, one unresolved required citation can match both codes —
  producing two findings for one target, contradicting the one-problem-one-
  finding invariant sub-decision 7 itself asserts (adr-008:284) and the
  `{ruleset, code, path}` aggregation (adr-008:228).
- Required fix: express the required-context escalation as a **profile promotion
  of `core-citation-unresolved`'s severity** (same code), and delete
  `profile-citation-unresolved` — or, if it is meant to be a genuinely distinct
  condition, define that condition so it cannot co-fire with the core code.

Blocking finding B4 — The ruleset tag for the new `core-citation-*` codes is
undecided and is actively contradicted by shipped reality (sub-decision 5/7,
adr-008:230-238, 287).
- The ADR labeled the citation codes "core mechanism (ships as engine defaults)"
  and said only that new codes are "ruleset-tagged per ADR-004." But the
  contract enum offers just `okf` and `profile` — there is no "core" ruleset
  (`internal/contract/envelope.go:38-39`), and the shipped `core-broken-link`,
  whose finding the citation rule *subsumes* in evidence contexts, is emitted as
  **`RulesetProfile`** (`internal/validate/links.go:61`). So the `core-` code
  prefix does **not** imply an OKF ruleset in this codebase. Criterion 5
  (separate OKF vs profile reporting) and the ADR's own `{ruleset, code, path}`
  fingerprint both require this decided per code.
- Required fix: state, per new `core-citation-*` code, whether it is `okf`- or
  `profile`-tagged, and reconcile explicitly with the shipped
  `core-broken-link → profile` precedent it subsumes.

3. Non-blocking findings (carry into implementation)

- "Syntactically disjoint" is a misnomer for in-bundle vs repo-path
  (adr-008:159, 165-173). These two scheme-less classes are distinguished by
  *where the target resolves*, not syntax. Pin the resolution order
  (bundle-membership first; then repo-path via the widened `../`) and state which
  class owns a scheme-less path that is bundle-relative but absent
  (in-bundle-unresolved) vs one that crosses the bundle boundary.
- `core-citation-loss` "warning + approval" (adr-008:237). ADR-004 defines a
  warning as non-failing; make explicit that the apply-time block comes from the
  **approval member**, not the warning severity, so a validate-time warning that
  nonetheless gates `apply` does not read as a contradiction of ADR-004.
- New read-path filesystem I/O (sub-decision 3, adr-008:194-198). `linkRules`
  today performs no direct FS access — membership is an injected `exists
  func(string)` (links.go:36,48). The repo-path `stat` introduces real FS reads
  into validation. Note that it must run through the same canonicalize/symlink-
  resolve primitives and the existing injection seam; and that ADR-005 is a
  **write** gate — reads are outside its stated scope, so "consistent with the
  ADR-005 write boundary" is intent-right but a slight category stretch.
- http(s) normalization asymmetry (sub-decision 6, adr-008:274-277). Fragment/
  query are stripped for in-bundle equality but URLs compare verbatim (fragment
  retained, no host case-folding). Deterministic and defensible, but flag it so
  `core-citation-duplicate`/citation-loss equality is unsurprising, and confirm
  scheme matching is case-insensitive (the shipped `uriScheme` regex is).

4. Validation assessment

`check-plan` / `go test` can prove: envelope-schema conformance, that the new
codes serialize and aggregate by fingerprint, resolver unit behavior on
fixtures, and the criterion-9 fixture pass. They **cannot** see any of the four
blockers: that the apply-time approval gate B1 relies on is actually
defined/wired (it is not — semantic); that the class partition B2 is total (it
is not — a fixture only exercises the cases someone thought to write); that the
ruleset tag B4 places citation findings on the correct side of the criterion-5
OKF/profile split; or that B3's promotion does not double-count. These are
exactly the decision-level gaps the tooling is blind to.

- check-plan ADR C1–C5 pass; ADR-C6 semantic-conflict remains a standing WARN.
- `sync-adr-index --check` clean.
- `go build ./...` PASS; docs-only change, full suite unaffected.
- Placeholder scan clean.

5. Recommendation

Revise and re-review. The model choice (Option A) is sound and most attack
surfaces pass, but land these edits before acceptance: (B1) ADR-008 owns the
apply-time citation-loss approval refusal, grounded in the ADR-003 approval
member, dropping the false ADR-006 attribution; (B2) make the three-class
partition total with a deterministic malformed catch-all covering fragment-only,
protocol-relative, and invalid-host http(s) targets; (B3) recast required-context
escalation as a severity promotion of `core-citation-unresolved` rather than a
second code; (B4) fix the ruleset tag per new code and reconcile with the shipped
`core-broken-link → profile` precedent it subsumes. All four are contained edits
— no re-architecting — so a single revision pass should reach READY.

---

Disposition: revisions applied on `docs/adr-008-provenance-citation-model` for
all four blockers and all four non-blocking findings; ADR-008 status remains
**proposed** pending a re-review to READY and subsequent human acceptance. This
record preserves the original NEEDS_REVISION verdict as issued.
