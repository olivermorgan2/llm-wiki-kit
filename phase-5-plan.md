# Phase 5 Implementation Plan: Enrichment + Index Maintenance

## Overview

Phase 5 extends Phase 4's read-only authoring path to **enrichment** (updating existing pages with new claims/citations from model assistance) and introduces **deterministic index maintenance** (ADR-011). Both operations reuse the staged-mutation engine (ADR-006) and YAML round-trip (ADR-001).

## Acceptance Criteria

All Phase 5 issues must collectively satisfy:
- **Criterion 11 (enrichment)**: Model-assisted enrichment produces valid pages with preserved unknown fields and citation obligations (ADR-008)
- **PRD §14 / FR9 (index reliability)**: Index operations are deterministic, no model calls, byte-idempotent (no diff on no-change), preserve human-authored sections outside fenced regions, and complete on a 5,000-doc reference corpus without a page apply

## Issue Breakdown

### Issue 81: Enrichment Skill Adapter

**Goal**: Reuse `page inspect`/`page plan`/`page apply` for enrichment (not just authoring).

**Scope**:
- `llm-wiki-enrich` skill reads existing page via `page inspect --json`
- Skill proposes additions (new claims, citations) to the **existing YAML structure**
- Skill writes modified YAML back using the same round-trip path (ADR-001)
- `page plan` and `page apply` are **not modified** — they already handle existing pages
- New fixtures in `profiles/core/examples/enrichment/`
- New `Enrichment` test cases in `internal/plan/plan_test.go`

**Key Decisions**:
1. **Reuse, not extend**: The existing `page plan`/`page apply` already handle existing pages (they read the file, compute a diff). Enrichment just needs a different YAML input.
2. **Unknown field preservation**: The round-trip path (ADR-001) already preserves unknown fields. Enrichment must not break this.
3. **Citation obligations**: ADR-008 sub-decision 5 specifies `profile-citation-required` (rule code: `profile-citation-required`, ruleset: `profile`, default severity: `error`) fires when a profiled page type has no citation but has a `source: model` claim. This is an ADR-010 profile-data-driven rule, not a hardcoded core rule. Enrichment triggers validation via `page plan --validate` just like authoring.
4. **Stale plan rejection**: ADR-006's hash-binding already handles stale plans. If the user modifies the page between `llm-wiki-enrich` and `page plan`, the plan will be rejected.

**Fixtures**:
- `profiles/core/examples/enrichment/existing-page.md`: A valid page with unknown fields (`custom_field: "value"`)
- `profiles/core/examples/enrichment/enrichment-input.json`: Enrichment input (e.g., `{"new_citations": ["@source2"]}`)
- `profiles/core/examples/enrichment/expected-output.md`: Expected output (unknown field preserved, citations added, page valid)

**Dependencies**: ADR-001 (YAML round-trip), ADR-006 (staged mutations), ADR-008 (citations), ADR-010 (profile validation)

---

### Issue 82: Deterministic Index Maintenance Core

**Goal**: Implement ADR-011 sub-decision #2 (index generation logic) and sub-decision #3 (fence detection).

**Scope**:
- `internal/index/index.go`: Core index package
  - `GenerateIndex(pages []PageMetadata) string`: Deterministic generation, no model calls
  - `ExtractExistingIndex(indexMd []byte) (string, string)`: Extracts content before and after fenced region
  - `MergeIndex(existingBefore, generated, existingAfter string) string`: Combines sections
- Generated region format:
  ```
  <!-- llm-wiki:index:start -->
  ## Claims
  - claim-001.md
  - claim-002.md
  
  ## Sources
  - source-001.md
  <!-- llm-wiki:index:end -->
  ```
- Sort order: by page type, then by filename (deterministic)
- Unit tests for all three functions, including:
  - Idempotency: `GenerateIndex(pages) == GenerateIndex(pages)`
  - Fence detection: correctly identifies start/end markers
  - Edge cases: missing fences, multiple fences, nested fences

**Key Decisions**:
1. **Fence detection**: Uses HTML comment markers `<!-- llm-wiki:index:start -->` and `<!-- llm-wiki:index:end -->`
2. **Sort strategy**: `sort.Slice` with stable comparator (type first bytewise ascending, then bundle-relative path bytewise ascending — not filename)
3. **No model calls**: The `GenerateIndex` function takes `[]PageMetadata` derived from frontmatter via `yamladapter` (ADR-001) and produces output without any model invocation
4. **LF line endings**: Generated region **always uses LF** regardless of host OS (per ADR-011 sub-decision 4)
5. **PageMetadata structure**: Must include frontmatter type, title, description, and bundle-relative path (per ADR-011 sub-decision 2)

**Dependencies**: ADR-011 (index maintenance), ADR-001 (YAML metadata)

---

### Issue 83: Index CLI Command and Dry-Run

**Goal**: Implement `llm-wiki index` command with `--dry-run` flag (ADR-011 sub-decisions #1, #4, #5).

**Scope**:
- `cmd/llm-wiki/index.go`: New CLI command
  - `llm-wiki index`: Regenerates index via ADR-006 staged transaction (stage → validate → journaled commit)
  - `llm-wiki index --dry-run`: Computes new index, shows unified diff of fenced region, exits without staging
  - `llm-wiki index --json`: Emits ADR-003 envelope with `operation: "index"` and new stable finding codes `core-index-stale` and `core-index-unmanaged`
- Walks wiki directory, reads all page metadata via `yamladapter`, calls `GenerateIndex`
- Compares generated fenced region bytes to current fenced region bytes; emits diff if different
- If fences are absent (fences-absent case per ADR-011 sub-decision 3): proposes appending a fenced region at EOF through the ADR-006 staging path; `--dry-run` shows full file diff including the appended section so the user can abort if placement conflicts with existing human content; refuses if fences are duplicate / nested / unterminated and emits `core-index-unmanaged`
- All writes routed through the ADR-006 transaction layer (manifest + hash-binding + atomic commit + journal) — the index command must not bypass ADR-006 by writing directly to disk
- Integration tests in `cmd/llm-wiki/index_test.go`:
  - `--dry-run` stages plan but does not apply
  - `--dry-run` on unchanged index shows no diff (exit 0)
  - Fences missing → proposes append through staging; refuses if duplicate/nested/unterminated
  - Writes preserve bytes outside fenced region (byte-preservation test)
  - ADR-006 stale-plan rejection test when `index.md` base hash has moved

**Key Decisions**:
1. **CLI structure**: Uses `cobra` (like other commands)
2. **JSON output**: Uses ADR-003 envelope, `operation: "index"`
3. **Diff display**: Uses `diff` package (or simple line-by-line comparison)
4. **Idempotency test**: Runs `llm-wiki index` twice on unchanged wiki, asserts no diff

**Dependencies**: ADR-003 (JSON contract), ADR-011 (index maintenance), Issue 82 (core logic)

---

### Issue 84: Enrichment Preview Fixtures

**Goal**: Create realistic enrichment scenarios to validate the skill adapter (Issue 81).

**Scope**:
- `profiles/core/examples/enrichment/`: New directory
  - `existing-source.md`: A source page with unknown fields and existing citations
  - `existing-claim.md`: A claim page with unknown fields and one citation
  - `enrichment-input.json`: Input to `llm-wiki-enrich` (e.g., add a citation to claim-001)
  - `expected-output.md`: Expected enriched output (unknown fields preserved, new citation added, page valid)
- `profiles/academic-research/examples/enrichment/`: Academic profile variants
  - `existing-method.md`: Method page with unknown fields
  - `enrichment-input.json`: Add a new step to the method
  - `expected-output.md`: Expected output
- New `cmd/llm-wiki-enrich/enrichment_test.go`:
  - Loads existing page, runs enrichment, validates output matches expected
  - Checks unknown fields are preserved
  - Checks citations are added
  - Checks page remains valid (no warnings)

**Dependencies**: Issue 81 (skill adapter), ADR-008 (citations), ADR-010 (profile validation)

---

### Issue 85: Index Stability Tests

**Goal**: Validate ADR-011's stability requirements (sub-decisions #4, #5).

**Scope**:
- `internal/index/stability_test.go`: New test file
  - **Idempotency test**: 
    - Create a wiki with 50 pages (mixed types)
    - Run `GenerateIndex` twice on the same input
    - Assert byte-for-byte equality
  - **No-diff-on-no-change test**:
    - Create a wiki, generate index
    - Generate index again without modifying wiki
    - Assert no diff (empty output from `llm-wiki index --dry-run`)
  - **Human section preservation test**:
    - Create `wiki/index.md` with content before `<!-- llm-wiki:index:start -->` and after `<!-- llm-wiki:index:end -->`
    - Run `llm-wiki index`
    - Assert human sections are byte-for-byte preserved
  - **Sort order test**:
    - Create pages in non-alphabetical order (including subdirectories to test bundle-relative path)
    - Generate index
    - Assert sort order is deterministic (by type bytewise ascending, then bundle-relative path bytewise ascending — not filename)
- `cmd/llm-wiki/index_stability_test.go`: Integration tests
  - **Large wiki test**: Create 5000 pages, generate index, assert no model calls (measure execution time, should be <10s; per PRD §14 reference corpus)
  - **No model call assertion**: Use a mock client that panics if called; pass to index generation; assert no panic

**Key Decisions**:
1. **Idempotency**: Uses `bytes.Equal` for byte-for-byte comparison
2. **Human section preservation**: Reads index before and after, extracts non-fenced regions, asserts equality
3. **No model call**: Uses a `mock.Client` that panics on any method call
4. **Performance**: Asserts index generation for 5000 pages takes <10s (extrapolates from 500-page baseline; PRD §14 requires 5000-doc reference corpus)
5. **LF line endings**: Cross-platform test generates index on Windows (CRLF host) and asserts generated region uses LF only

**Dependencies**: ADR-011 (index maintenance), Issue 82 (core logic), Issue 83 (CLI command)

---

## Validation Finding Codes (ADR-004 Engine)

**Goal**: Implement `core-index-stale` and `core-index-unmanaged` findings as engine-side validation findings per ADR-011 sub-decision #5.

**Scope**:
- Add to `internal/validate/validate.go`:
  - **`core-index-stale`**: Fires when fences are correctly placed but generated content differs from current fenced region. Default severity: `error` (promotable via profile)
  - **`core-index-unmanaged`**: Fires when fences are malformed (duplicate start, no matching end, nested, unterminated). Default severity: `error` (promotable via profile)
- These findings discharge ADR-004's deferred "index-consistency checks"
- Validation runs during `page plan` and `page apply` when an existing index is detected
- New test cases in `internal/validate/validate_test.go`:
  - `core-index-stale` fires when index content is outdated
  - `core-index-stale` does not fire when index content matches generated
  - `core-index-unmanaged` fires for duplicate start fences
  - `core-index-unmanaged` fires for unterminated fences
  - `core-index-unmanaged` does not fire for correctly placed fences

**Key Decisions**:
1. **Finding codes**: Use `core-` prefix (engine-side, not profile-locked) per ADR-011 sub-decision 5
2. **Severity**: Both findings default to `error` (user can override via profile `min_severity` map)
3. **Validation timing**: Runs during `page plan` (preview) and `page apply` (enforcement) when `wiki/index.md` exists
4. **Scope**: Only validates index files that are engine-managed (have fences); completely unmanaged index files (no fences) are outside engine responsibility

**Dependencies**: ADR-004 (finding codes), ADR-011 (index maintenance), Issue 83 (CLI command)

---

## Post-Apply Index Refresh (Enrichment Skill Integration)

**Goal**: Ensure enrichment skill triggers index refresh after page apply per ADR-011 sub-decision #7 ("immediately after page apply").

**Scope**:
- Modify `llm-wiki-enrich` skill (Issue 81) to invoke `llm-wiki index` after successful `page apply`
- This is a separate staged transaction per ADR-006 (not bundled with page apply)
- New test in `cmd/llm-wiki-enrich/enrichment_test.go`:
  - After enrichment, assert `wiki/index.md` is updated
  - Assert `core-index-stale` does not fire on next `page plan`

**Key Decisions**:
1. **Separate transaction**: Index refresh is its own ADR-006 staged operation (not bundled with page write)
2. **Error handling**: If index refresh fails, enrichment still succeeds (but user sees `core-index-stale` on next validation)
3. **Timing**: Runs after `page apply`, before skill returns control to user

**Dependencies**: Issue 81 (enrichment skill), Issue 83 (index CLI command), ADR-006 (staged transactions)

---

## Init Fence Scaffolding (Phase 2 Backfill)

**Goal**: Ensure `wiki init` scaffolds `wiki/index.md` with correct fence markers per ADR-011 sub-decision #3.

**Scope**:
- Add to `internal/init/init.go` (Phase 2 work):
  - Generate `wiki/index.md` with correct fences and empty content
  - Template:
    ```markdown
    # Wiki Index
    
    <!-- llm-wiki:index:start -->
    <!-- llm-wiki:index:end -->
    ```
- New test in `internal/init/init_test.go`:
  - After `wiki init`, assert `wiki/index.md` exists with correct fences
  - Assert `page plan` does not emit `core-index-unmanaged`

**Key Decisions**:
1. **Fence format**: Use HTML comment markers `<!-- llm-wiki:index:start -->` and `<!-- llm-wiki:index:end -->`
2. **Empty content**: Fences are present but no entries (index generation will populate)
3. **File placement**: `wiki/index.md` (same directory as `wiki/`)

**Dependencies**: ADR-011 (index maintenance), Phase 2 (init system)

---

## Dependency Graph

```
Issue 81 (Enrichment Skill)
    ├── Issue 84 (Enrichment Fixtures) [depends on 81]
    └── Post-Apply Index Refresh [depends on 81, 83]

Issue 82 (Index Core)
    ├── Issue 83 (Index CLI) [depends on 82]
    ├── Issue 85 (Index Tests) [depends on 82, 83]
    ├── Validation Finding Codes [depends on 82, 83]
    └── (no other dependencies)

Phase 2 Backfill: Init Fence Scaffolding [standalone, can be done in parallel]
```

**Mapping to ADR-011's I2–I7 notation**: I2=Issue 82, I3=Issue 83, I4=Issue 85, I5=Validation Finding Codes, I6=Post-Apply Index Refresh, I7=Init Fence Scaffolding (Phase 2 backfill). Enrichment track (Issues 81, 84) is criterion 11 work outside the I2–I7 index numbering.

## Execution Order

**Parallel Track 1**: Enrichment
1. Issue 81: Implement enrichment skill adapter
2. Issue 84: Create enrichment fixtures and tests

**Parallel Track 2**: Index
1. Phase 2 backfill: Init fence scaffolding (can run concurrently)
2. Issue 82: Implement core index logic
3. Issue 83: Implement index CLI and dry-run (routed through ADR-006)
4. Validation finding codes: `core-index-stale` / `core-index-unmanaged`
5. Post-apply index refresh: enrichment skill triggers index after page apply
6. Issue 85: Implement stability tests (5000-page, LF enforcement, idempotency)

**Total**: 5 issues + 3 supplementary sections, 2 parallel tracks, ~10-12 days of work (estimated).

## Risk Mitigation

1. **Enrichment breaks unknown field preservation**: Tests in Issue 84 will catch this early
2. **Index generation is slow**: Performance test in Issue 85 will catch this
3. **Index modifies human sections**: Human section preservation test in Issue 85 will catch this
4. **Enrichment triggers false warnings**: Enrichment fixtures in Issue 84 include valid pages, so warnings indicate a bug

## Out of Scope

- **Per-profile index customization**: Not in Phase 5 (per ADR-011 sub-decision #2)
- **Log.md maintenance**: Not in Phase 5 (separate ADR candidate)
- **Index validation**: Not in Phase 5 (future ADR candidate)
- **Enrichment UI/UX**: Skill adapter only, no UI changes

## Success Criteria

All issues merged, all tests passing, all acceptance criteria satisfied:

- **Criterion 11**: Enrichment produces valid pages with preserved fields and citations ✓
- **PRD §14 / FR9**: Index operations are byte-idempotent, deterministic (no model calls), preserve human sections outside fenced regions, and complete on 5,000-doc reference corpus ✓
