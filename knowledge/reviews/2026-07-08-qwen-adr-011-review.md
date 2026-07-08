Subject: ADR-011 — Deterministic index maintenance
Reviewer: Qwen3.7 Max via OpenRouter (`qwen/qwen3.7-max`)
Date: 2026-07-08
Verdict: READY
Score: 4/5

Blocking findings: None

Non-blocking findings (clarity only, folded into the ADR draft):
1. Bundle-relative path definition reference — add explicit ADR-005/ADR-006 traceability for the entry/sort key.
2. Generated-region format deferred without explicit callout — Consequences should note the implementation issue must define and fixture-test the generated-region markdown format.
3. Fence-absent append edge case — clarify that the dry-run/staged preview shows the full file diff so the user can abort if append placement conflicts with existing content.

Gate status: ADR-011 is READY for Oliver's explicit acceptance. Status flips
`proposed` → `accepted` only after Oliver accepts, in the same PR before merge,
keeping ratification debt at 0. Phase 5 implementation issues I2–I7 remain
blocked until ADR-011 is accepted and merged.

---

## ADVERSARIAL REVIEW — ADR-011 DETERMINISTIC INDEX MAINTENANCE

**Reviewer:** Qwen3.7 Max / OpenRouter (`qwen/qwen3.7-max`)  
**Subject:** ADR-011 deterministic index maintenance  
**Date:** 2026-07-08  
**Verdict:** READY  
**Score:** 4/5

---

### BLOCKING FINDINGS

None.

---

### NON-BLOCKING FINDINGS

1. **Bundle-relative path definition reference.** Sub-decisions 2 and 4 use "bundle-relative path" as the entry key and sort key, but don't explicitly reference which accepted ADR defines the bundle boundary. The reader must infer this comes from ADR-005 (safe filesystem layer) and ADR-006 (staging area). Recommend adding a parenthetical reference like "(per ADR-005 boundary)" for traceability. This is a clarity improvement, not a correctness issue.

2. **Generated region format deferred without explicit callout.** Sub-decision 1 states "per-type sections *inside* a single generated region" but the exact markdown format of those sections (headers, list structure, entry layout) is left unspecified. This is appropriate for an ADR (which fixes decisions, not formats), but the Consequences section could explicitly note that the implementation issue must define and fixture this format. Currently the Consequences mention "the fence names" and "the sort key" as items to maintain, but not the generated-region format itself.

3. **Fence-absent append behavior edge case.** Sub-decision 3 states that when fences are absent, dry-run shows "an *append-a-fenced-region-at-end* proposal" which applies through the staged path. This is safe (preview + staged transaction), but the ADR doesn't explicitly address what happens if the file has trailing content that the user wants to preserve after the appended fence. The staged-plan preview will show the full diff, so the user can abort, but a one-sentence clarification ("the append proposal shows the full file diff, allowing the user to abort if the placement conflicts with existing content") would eliminate ambiguity.

---

### SCOPE COVERAGE (Issue #76, 9 items)

1. **Index inventory** — DECIDED. Exactly one generated index (`index.md`), no per-type files, no profile-specific files. Per-type/profile views explicitly deferred with rationale (multiplies staleness surface, would extend ADR-010 closed vocabulary → new ADR). `log.md` explicitly out of scope (append log, not derived index). ✓

2. **Derivation from metadata / missing metadata** — DECIDED. Entries derived solely from frontmatter (`type`, `title`, `description`) + bundle-relative path via ADR-001 `yamladapter`. Missing-metadata behavior is total and fixed: title absent → filename stem; description absent → omit; YAML parse failure → exclude (already an ADR-004 hard error). No new parser, no body scraping. ✓

3. **Human-maintained section preservation** — DECIDED. HTML-comment fence convention (`<!-- llm-wiki:index:start -->` / `<!-- llm-wiki:index:end -->`), matching repo's own `adr-index` idiom. Engine may replace only bytes strictly inside one fence pair; everything outside is byte-preserved. Edge cases fixed: fences absent → append proposal via staged path; duplicate/nested/unterminated → refuse and emit finding; `init` scaffolds fence pair. ✓

4. **Stable ordering / duplicates / idempotency** — DECIDED. Duplicates structurally impossible (keyed on bundle-relative path, one entry per page). Sort key: `type` (bytewise ascending), then path (bytewise ascending); titles display-only, never sort keys. Byte-idempotency is a test-gated invariant: regenerate twice over unchanged tree → byte-identical; regenerate current index → empty diff. LF line endings regardless of host platform. ✓

5. **Stale-index validation behavior** — DECIDED. New finding `core-index-stale` (ruleset: profile, warning default, promotable via ADR-010 `severities` map). Stale = current fenced region bytes differ from freshly recomputed region. Companion finding `core-index-unmanaged` for unmaintainable index (missing/duplicate/nested/unterminated fences), same warning default. `validate` reports only, never mutates. Discharges ADR-004's deferred "index-consistency checks" interlock. ✓

6. **Dry-run/diff + JSON contract** — DECIDED. `--dry-run` non-mutating, shows unified diff of fenced region (current vs recomputed). `--json` uses ADR-003 v1 envelope unchanged; only delta is one new `operation` value (no new fields, no version bump). ✓

7. **Write path through ADR-006** — DECIDED. Standalone staged operation routed through ADR-006 transaction (stage → validate → journaled atomic commit), invoked by skills after page apply. Composes ADR-005/006, adds nothing to them. Rejects Option C (fold into every page plan) on base-hash contention grounds. ✓

8. **Runtime constraints / no model calls / no dependency creep** — DECIDED. No model calls anywhere (hard, test-visible invariant per FR9/§14). New `internal/index` package using stdlib + existing `yamladapter` only; no new runtime dependency (overlay rule: new dependency requires new ADR). Meets §14: 5,000-doc corpus regeneration without model call, stable output for identical input. ✓

9. **Criterion 11 relationship** — DECIDED. `index.md` is an existing page, so every mutation is previewable: `--dry-run`/diff in direct CLI use, ADR-006 staged-plan preview in skill flows. Enrichment adds no new decision (reuses ADR-006/008/001). Should implementation uncover a new enrichment decision boundary, stop and raise a new ADR. ✓

All 9 scope items are decided or explicitly deferred with rationale.

---

### ADR BOUNDARY CHECK

**ADR-003 (JSON contract):** ADR-011 uses ADR-003 v1 envelope unchanged, adds only one new `operation` value. No new fields, no version bump. Consistent. ✓

**ADR-004 (validation/severity):** ADR-011 supplies `core-index-stale` and `core-index-unmanaged` findings at warning default, promotable via ADR-010's `severities` map. Discharges ADR-004's deferred "index-consistency checks" interlock. ADR-004's Consequences reference "ADR-010" (the provisional number), which ADR-011 reconciles to ADR-011 via the numbering correction (ADR-004 not edited in place, per overlay rule). Consistent. ✓

**ADR-005 (safe FS):** ADR-011 composes ADR-005's per-file atomicity and path/symlink safety. Adds no new write path. Index mutation routes through ADR-006, which routes through ADR-005. Consistent. ✓

**ADR-006 (staged mutation):** ADR-011 uses ADR-006's full transaction lifecycle (stage → validate → journaled atomic commit, hash-bound stale-plan rejection). The standalone-operation call (Option A vs Option C) is a composition choice, not an extension. Consistent. ✓

**ADR-008 (provenance/citation):** ADR-011 clarifies index links are "navigational, never citation contexts." ADR-008's citation mechanism does not apply to index entries. Consistent. ✓

**ADR-010 (profile schema):** ADR-011 uses ADR-010's existing `severities` override map for severity promotion. Adds no new profile keys. Finding codes `core-index-stale` (with `ruleset: profile`) match ADR-010's taxonomy for engine-shipped default rules that are profile-configurable (parallel to `core-broken-link`). Consistent. ✓

**PRD FR9 (Index maintenance):** ADR-011 addresses all FR9 requirements: deterministic derived view from page metadata, avoid duplicate entries (structurally impossible via path-keying), preserve clearly marked human-maintained sections (fence convention), use title and description where available (with defined fallbacks), stable ordering (type+path bytewise), avoid model calls (hard invariant), support dry-run and diff output (`--dry-run` unified diff). Consistent. ✓

**PRD §14 (reliability/performance):** ADR-011 addresses the reliability envelope: repeating unchanged operation produces no diff (byte-idempotency test-gated), output stable for identical inputs (sort key deterministic, LF-only), 5,000-doc corpus regeneration without model call (no-model-call invariant + stdlib-only constraint). Consistent. ✓

No contradictions found with accepted ADRs or PRD requirements.

---

### CONTESTABLE CALL ANALYSIS: Option A (standalone) vs Option C (folded)

**The call:** Index mutation is a standalone ADR-006 staged operation (Option A), not folded into every page-mutation plan (Option C).

**Option C's apparent advantage:** Tree and index commit atomically; index never momentarily stale.

**Option A's defense (stated in ADR-011):** If `index.md` is in every page plan, its ADR-006 base hash becomes a contention point. Plan A (editing page X) and plan B (editing page Y), staged in sequence or concurrently, both record index.md's base hash at their respective stage times. If plan A applies first and updates index.md, plan B's recorded base hash for index.md is now stale, and ADR-006's `apply` rejects plan B as a stale plan. This converts a warning-severity condition (briefly stale index) into apply-blocking coupling on unrelated page edits.

**Review assessment:** The defense is technically sound. ADR-006's hash-bound stale-plan rejection is strict: any target whose base hash no longer matches is rejected. Including index.md in every page plan would indeed multiply stale-plan rejections, because index.md changes on every page edit (its regenerated content reflects the new page). The user would need to re-plan (re-stage) plan B after plan A applies, even though plan B's page edit is unrelated to plan A's page edit. This is poor UX for a warning-severity condition.

**Is the trade-off honestly presented?** Yes. The ADR explicitly names this as "the one genuinely contestable call" and presents Option C's advantage ("the index is never even momentarily stale; one transaction covers both") before rejecting it. The Consequences section records the trade-off: "the decoupled write op means the tree and index can be momentarily out of sync between a page apply and the index refresh (bounded to a warning, caught by `validate`)."

**Does the self-healing claim hold?** Yes. ADR-011 states that a missed refresh is caught by `validate` (`core-index-stale` at warning severity) and self-healed by a subsequent regen. The user can run `llm-wiki index --dry-run` to see the diff, then `llm-wiki index` to apply. In skill flows, the skill invokes the index refresh after page apply, so the window of staleness is bounded by the skill's execution time. If a skill crashes between page apply and index refresh, `validate` catches it on the next run.

**Verdict on the contestable call:** The rejection of Option C is well-reasoned and the trade-off is honestly presented. The standalone-operation design is defensible.

---

### TESTABILITY CHECK

Each normative claim maps to a testable assertion:

- **Index inventory:** Test that exactly one `index.md` is generated per bundle; no per-type or profile-specific index files exist. ✓
- **Derivation:** Test entry content against frontmatter; test fallbacks (title absent → stem, description absent → omitted, YAML parse failure → excluded). ✓
- **Fence convention:** Test fence parsing edge cases (absent → append proposal; duplicate/nested/unterminated → refuse); test byte-preservation outside fence. ✓
- **Stable ordering:** Test sort order (type ascending, path ascending); test bytewise comparison across runs. ✓
- **Byte-idempotency:** Test that regenerating twice produces byte-identical output; test that regenerating current index produces empty diff. ✓
- **LF line endings:** Test that generated region uses LF regardless of host platform. ✓
- **core-index-stale:** Test that `validate` emits finding when fenced region differs from recomputed; test absence when current. ✓
- **core-index-unmanaged:** Test that `validate` emits finding for unmaintainable index (missing/duplicate/nested/unterminated fences). ✓
- **Severity promotion:** Test that ADR-010 `severities` map promotes warning to error. ✓
- **Dry-run:** Test that `--dry-run` produces diff without mutation. ✓
- **JSON contract:** Test that `--json` emits ADR-003 envelope with new `operation` value. ✓
- **Write path:** Test that mutation routes through ADR-006 transaction (staged, validated, journaled). ✓
- **No model calls:** Test that index regeneration does not invoke any model API (mock/instrument). ✓
- **No new dependency:** Verify `go.mod` unchanged; verify `internal/index` imports only stdlib + `yamladapter`. ✓
- **Criterion 11:** Test that `--dry-run` shows preview before apply; test ADR-006 staged-plan preview in skill flow. ✓

All normative claims are implementable with clear tests.

---

### EVIDENCE HONESTY AND WORKFLOW CORRECTNESS

- **Status:** `proposed` — correctly recorded. ✓
- **No false acceptance claims:** ADR states "**no review has run and no acceptance is claimed**." Knowledge layer (index.md, state.md, log.md) repeats this. ✓
- **Reviewer substitution recorded:** Issue #76 comment records Oliver's approval of Qwen3.7 Max substitution for Codex. Knowledge/log.md records the substitution. ADR-011 Consequences section names the reviewer. ✓
- **Numbering correction:** Provisional ADR-010 → actual ADR-011 explicitly recorded in Context, Consequences, and log.md. ADR-004's "interlock with ADR-010" reconciled to ADR-011 (ADR-004 not edited in place). ✓
- **Local validation:** `sync-adr-index --check` in sync (11 rows); `git diff --check` clean; placeholder scan clean; `go test ./...` green (docs-only change). ✓
- **No product/source code written:** Confirmed (docs-only change). ✓
- **No accepted ADR edited in place:** Confirmed (ADR-004 reference reconciled in ADR-011, not edited). ✓
- **Not pushed / no PR:** Confirmed (local branch `docs/adr-011-index-maintenance`, no PR yet). ✓

Evidence is honest. Workflow is correct.

---

### VALIDATION / GATE NOTE (for knowledge/log.md)

Qwen3.7 Max / OpenRouter adversarial review of ADR-011 (deterministic index maintenance) returned **READY** (score 4/5). All 9 issue-#76 scope items decided or explicitly deferred with rationale. No contradictions with accepted ADR-003/004/005/006/008/010 or PRD FR9/§14. The contestable call (standalone staged operation vs folded-into-every-page-plan) is well-argued and honestly presented. All normative claims are testable. Three non-blocking clarity improvements noted (bundle-relative path reference, generated-region format callout, fence-absent append edge case); none block acceptance. Reviewer substitution (Codex → Qwen3.7 Max) recorded on issue #76 and in log.md; overlay-legal (non-Anthropic, cross-vendor). Verbatim review artifact archived at `knowledge/reviews/2026-07-08-qwen-adr-011-review.md`. **Gate status:** ADR-011 is READY for Oliver's explicit acceptance. Status flips `proposed` → `accepted` only after Oliver accepts, in the same PR before merge, keeping ratification debt at 0. Phase 5 implementation issues I2–I7 remain blocked until ADR-011 is accepted and merged.
