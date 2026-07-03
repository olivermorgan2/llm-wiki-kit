# Codex ADR-008 re-review — 2026-07-03

Scope: second adversarial ADR gate for `docs/adr-008-provenance-citation-model`
(PR TBD, issue #32) after the revision commit addressing the loop-1 blockers —
the Phase 3 provenance/citation model prerequisite.

Reviewer: Codex via Hermes `openai-codex` provider fallback (standalone codex
CLI unavailable). Re-run of the same independent adversarial ADR-review pass,
which re-read the revised file on disk and re-verified every fix against the
accepted ADRs and the shipped `internal/contract` + `internal/validate/links.go`
code.

---

1. Verdict: READY — Score 5/5

The revision is decision-complete. Each prior blocker is not merely reworded but
resolved at the mechanism level, grounded in the correct ADR/contract surface
with accurate file:line citations that were independently confirmed. No new
blocking findings were introduced by the revision.

2. Status of prior blockers

Prior blocker B1 (preservation/approval) — FIXED.
- Sub-decision 6 now explicitly states ADR-006's only `apply` refusal is
  stale-plan rejection and it has **no** content-approval gate, so ADR-008
  **owns** the new apply-time content rule and does not attribute it to ADR-006.
  Verified: ADR-006's Decision contains only stale-hash refusal (adr-006:102)
  and no approval/deletion language. The gate is grounded in the ADR-003
  `approval` member returning `StatusApprovalRequired`/`ExitApprovalRequired` —
  all confirmed to exist (`internal/contract/envelope.go:83` Approval field /
  `:28` status; `internal/contract/exitcode.go:18,39`). The block is carried by
  the approval member, **not** the warning severity ("the `core-citation-loss`
  warning is *reported* non-failingly per ADR-004, while the separate `approval`
  member is what gates the commit"), and the loss code is retabled as "warning
  (apply-gated via `approval`)". No envelope change. Sits inside ADR-006's own
  delegation of "provenance/citation-preservation content rules" (adr-006:174).

Prior blocker B2 (total partition) — FIXED.
- Sub-decision 2 is retitled "deterministic, offline, and **total**" and
  replaced with an ordered first-match-wins procedure plus a catch-all class 4.
  Invalid-host http(s) is now explicitly **malformed, not unresolved**. Class 4
  enumerates the previously-uncovered cases: empty, non-http(s) scheme,
  fragment-only (`#sec`), protocol-relative (`//host/p`), and repo-root escape.
  The malformed-vs-unresolved split is now a partition (malformed = class-1
  invalid-host + class-4; unresolved = well-formed-but-absent classes 2/3) with
  the explicit guarantee "no citation target falls between the two codes." The
  false "syntactically disjoint" claim is gone; the ordered classifier also
  folds the resolution-order non-blocking item. Total over the shipped grammar.

Prior blocker B3 (double-count) — FIXED.
- `profile-citation-unresolved` is removed from the profile table (only
  `profile-citation-required` and `profile-citation-target-type` remain). The
  required-must-resolve case is recast as a **severity promotion of
  `core-citation-unresolved`** (same code, promoted severity — one finding,
  never a `core-` warning plus a separate `profile-` error), explicitly invoking
  the ADR-004 promotion mechanism. Consistent with the one-finding invariant.

Prior blocker B4 (ruleset tag) — FIXED.
- Sub-decision 5 adds a dedicated "Ruleset tag (decided here)" block: the enum
  is exactly `okf`/`profile` with "no 'core' ruleset," verified against
  `internal/contract/envelope.go:38-39`; the `core-` prefix is a naming/origin
  marker, not a ruleset. Every new `core-citation-*` and `profile-citation-*`
  code is tagged **`ruleset: profile`**, reconciled with the shipped
  `core-broken-link` → `RulesetProfile` it subsumes (verified at
  `internal/validate/links.go:61`). The criterion-5 rationale is sound (citation
  obligations arise from profile-designated evidence contexts, not universal
  OKF) and guarantees the subsuming/subsumed findings share a ruleset.

3. Status of prior non-blocking findings

- Resolution order pinned — FOLDED (ordered first-match: in-bundle before
  repo-path before catch-all).
- Approval-vs-severity wording — FOLDED (sub-decision 6 makes the distinction
  explicit and cites the ADR-004 non-contradiction).
- Read-path FS injection seam + ADR-005-is-a-write-gate caveat — FOLDED
  (sub-decision 3: repo-path `stat` runs behind the same injected `exists` seam
  and canonicalize/symlink primitives; ADR-005 is a write gate, reads are
  outside its scope but bounded by the same repo-root anchor/symlink rules).
- URL verbatim / case-insensitive scheme — FOLDED (scheme detection
  case-insensitive per shipped `uriScheme`; target otherwise byte-for-byte).

4. New blocking findings

None.

5. Non-blocking findings to carry into implementation

- Gate class 3 on `isIntraWiki` explicitly. First-match ordering plus
  `isIntraWiki`'s false-on-`//`/`#` (links.go:74) already makes the verdict
  deterministic, but the implementer should wire class 3 to require
  `isIntraWiki` so a `//`/`#` target cannot be speculatively stat-tested against
  the repo root before falling to the catch-all. Wording-tightening, not a
  decision gap.
- Define whether a present-but-unresolved citation satisfies a require-citation
  obligation. `profile-citation-required` and a promoted `core-citation-unresolved`
  are distinct codes (both firing for one under-cited claim is two *different-code*
  findings, permitted by per-code aggregation). The Phase-4 require-citation-when
  vocabulary should state whether "has a citation" counts link-presence or
  resolvable-presence. Phase-4 profile-vocabulary detail, within the
  deferred-keys scope.

6. Validation assessment

`check-plan` / `go test` can now prove the mechanical surface (envelope
conformance, code serialization, resolver fixtures, criterion-9 fixture pass).
The four semantic gaps they could not see in loop 1 are now closed *in the
decision text itself*: the apply-time approval gate is owned and
contract-grounded (B1), the classifier is provably total with a catch-all (B2),
the promotion avoids double-counting by construction (B3), and the ruleset
placement is fixed and reconciled with shipped `RulesetProfile` (B4). The two
residual carry-ins are genuinely implementation/Phase-4 concerns, not
decision-level holes.

- check-plan ADR-C1–C5 pass; ADR-C6 semantic-conflict remains a standing WARN.
- `sync-adr-index --check` clean; placeholder scan clean.
- `GOTOOLCHAIN=local /Users/hermes/sdk/go1.24.13/bin/go build/vet/test ./...` —
  PASS across all 11 packages (docs-only change; suite unaffected).

7. Recommended gate action

Accept now. The ADR is decision-complete and internally consistent with ADR-003
(envelope/approval/exit codes), ADR-004 (severity precedence, promotion,
criterion-5 separation), ADR-005 (write-gate boundary correctly scoped), ADR-006
(owns the content-approval gate ADR-006 delegated, no false attribution),
ADR-007 (profile rules as deferred-key data), and ADR-009 (`.llm-wiki/`
repo-root anchor) — and with the shipped `links.go` resolver and `RulesetProfile`
tag it reuses. Move to Status: accepted and merge; the two non-blocking items
ride into the Phase 3/4 implementation issues as notes.

---

Disposition: status flipped `proposed` → `accepted` under the 2026-07-03
autonomous-phase mandate, **flagged for Oliver's async ratification** (matching
ADR-006/007/009); ADR index re-synced; the two non-blocking carry-ins recorded
for the Phase 3/4 implementation issues. Supply-chain signing/provenance stays
deferred to a dedicated supply-chain ADR — **not** covered by ADR-008 — so the
residual supply-chain risk and the "sourced claims lose provenance" risk stay
`open` (the latter now has its mechanism decided here). This record preserves the
READY / 5/5 verdict as issued.
