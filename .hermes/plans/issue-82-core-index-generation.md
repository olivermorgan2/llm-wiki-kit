# Implementation Plan — Issue #82: Core index generation and fenced-region parsing

**Issue:** #82 `feat(index): Core index generation and fenced-region parsing`
**Milestone:** Phase 5 — Enrichment + index maintenance (index track item I2)
**Governing ADR:** ADR-011 (deterministic index maintenance), sub-decisions **2, 3, 4, 8**; secondary ADR-001 (yamladapter seam).
**Branch:** `feat/82-core-index-generation` (current).
**Deliverable of this issue:** a new **pure** `internal/index` package (stdlib + `yamladapter` only). No CLI, no filesystem walk, no validate findings, no staging.

---

## Context

ADR-011 fixes *what* the index is (a deterministic derived view of page frontmatter, written into an HTML-comment fenced region of the OKF-reserved `index.md`, byte-idempotent, LF-only, no model calls) but deliberately leaves the *concrete markdown format* to the implementation issue, requiring it be **locked behind fixtures**. Issue #82 is the gate-unblocked first implementation step (I1/#76 ADR-011 already accepted). It builds only the deterministic core primitives that later issues compose: #83 (CLI `llm-wiki index` + directory walk + `--dry-run`/`--json` + ADR-006 staging), #85 (validate findings `core-index-stale` / `core-index-unmanaged`), #86 (`init` fence scaffolding), #84 (integration + cross-platform/perf tests). Keeping this issue pure (no I/O, no model, no CLI) makes the byte-idempotency and no-model-call invariants unit-testable now and lets the boundary-safety layer (fsafe/ADR-005) stay where it already lives.

---

## Files to inspect (builder pre-read)

| Path | Why |
|---|---|
| `design/adr/adr-011-deterministic-index-maintenance.md` | Authoritative contract — sub-decisions 2 (derivation/fallbacks), 3 (fence convention + edge cases), 4 (sort key, duplicates, LF, byte-idempotency), 8 (stdlib-only, no new dependency). |
| `internal/yamladapter/yamladapter.go`, `internal/yamladapter/goccy.go` | `Adapter` interface, `New()`, `Unmarshal(data, v)` (ignores unknown fields → decode straight into a small struct). |
| `internal/validate/frontmatter.go` | `splitFrontmatter(data) (yaml, body, err)` — the pattern to **mirror** (unexported; will be re-implemented locally). |
| `internal/validate/rules.go` | `evaluatePage`, `hasNonEmptyString`, modeled field set `{type,title,description}` — canonical field-reading pattern. |
| `internal/plan/plan.go` | `ResolvePage`, `PageRef{Root,Rel,Abs}`, `relForRoot` — bundle-relative resolution (context for #83; **not** used in #82). |
| `internal/validate/validate_test.go`, `internal/plan/plan_test.go`, `internal/yamladapter/goccy_test.go` | House test style: stdlib `testing` only, table-driven `[]struct{}`+`t.Run`, `fstest.MapFS`/`t.TempDir()`, inline expected consts (the repo's "golden" idiom — **no** `-update` machinery). |
| `.github/workflows/test.yml` | The full CI gate (build → vet → acceptance → test); there is **no linter**. |
| `phase-5-plan.md` | Cross-reference only — note the conflicts flagged below. |

---

## Files to create / change

**Create (no existing files change — stay in scope):**

- `internal/index/index.go` — package doc (cite ADR-011); `PageMetadata`; fence marker consts; sentinel errors; `GenerateIndex`; `ExtractFencedRegion`; `Reconstruct`.
- `internal/index/metadata.go` — `ParsePageMetadata` + local `splitFrontmatter` + fallback logic.
- `internal/index/index_test.go` — the full unit matrix below.

**Unchanged:** `go.mod`/`go.sum` (stdlib + existing `yamladapter` only — ADR-011 sub-decision 8 forbids a new runtime dependency without a superseding ADR). No edits to `validate`, `plan`, or `cmd`.

---

## API decisions

All names from the issue's acceptance criteria are provided. Two signatures are **intentional, owner-approved refinements** of the issue's illustrative signatures (see Conflicts §):

```go
package index

// PageMetadata is one indexable page, derived solely from frontmatter + path
// (ADR-011 sub-decision 2). No body scraping, no new parser.
type PageMetadata struct {
    Type        string // frontmatter `type`; empty → "unknown"
    Title       string // frontmatter `title`; empty → filename stem
    Description string // frontmatter `description`; may be empty (omitted at render)
    Path        string // bundle-relative, slash-form; the sort + dedup key
}

// GenerateIndex renders the deterministic INNER region body (the only bytes the
// engine may rewrite — no fence markers). Sorted, LF-only, byte-idempotent.
func GenerateIndex(pages []PageMetadata) string

// ParsePageMetadata parses one page's raw bytes into PageMetadata. PURE: no file
// I/O. relPath is recorded verbatim as Path and used for the title-stem fallback.
// Returns an error on missing/unterminated frontmatter or YAML parse failure;
// the #83 walk EXCLUDES such pages (ADR-011 sub-decision 2).
func ParsePageMetadata(content []byte, relPath string) (PageMetadata, error)

// ExtractFencedRegion splits content around exactly one fence pair.
//   before = bytes up to and including the start-marker line + its newline
//   fenced = the inner bytes (compared against GenerateIndex for staleness)
//   after  = the end-marker line and everything after it
// Reconstruct(before, GenerateIndex(pages), after) rebuilds the file with human
// content byte-preserved and markers immutable.
func ExtractFencedRegion(content []byte) (before, fenced, after []byte, err error)

// Reconstruct assembles before + body + after (covers phase-5-plan's "MergeIndex").
func Reconstruct(before, body, after []byte) []byte

const (
    FenceStart = "<!-- llm-wiki:index:start -->"
    FenceEnd   = "<!-- llm-wiki:index:end -->"
)

var (
    ErrNoFence       = errors.New("index: no fenced region found")     // → #83 append-at-EOF path
    ErrMalformedFence = errors.New("index: malformed fenced region")   // duplicate/nested/unterminated/orphan-end → #85 core-index-unmanaged
)
```

**Generated region format (LOCKED here, fixtured in tests — required by ADR-011 Consequences).** `GenerateIndex` builds the string as:

1. Group pages by `Type`. Ordered types = unique `Type` values sorted **bytewise ascending**.
2. Within each type, entries sorted by `Path` **bytewise ascending** (`Title` is display-only, never a sort key — ADR-011 sub-decision 4).
3. For each type section (a single blank line `\n` separates consecutive sections):
   - header line: `## ` + `Type` + `\n`
   - per entry: `- [` + `Title` + `](` + `Path` + `)` + (`Description != "" ? " — " + Description : ""`) + `\n`
4. Empty input → empty string `""`.

Concrete golden example (pin verbatim as an inline const in the test):

```markdown
## claim
- [Claim One](claims/claim-001.md) — First claim
- [claim-002](claims/claim-002.md)

## source
- [Source One](sources/source-001.md)
```

Rationale for including `[Title](Path) — Description`: ADR-011 sub-decision 2 mandates "uses title and description where available"; the issue's minimal `- claim-001.md` example is under-specified against the ADR, and the ADR wins. `Path` remains the link target, sort key, and structural dedup key (one entry per bundle-relative path → duplicates structurally impossible, ADR-011 sub-decision 4).

**Fence detection rules:** a marker matches only as a whole line (compare after `strings.TrimRight(line, "\r")`, mirroring `validate.splitFrontmatter` so CRLF files parse). Exactly one start + one matching end → success. Zero markers → `ErrNoFence`. Everything else (≥2 starts, ≥2 ends, start-without-end, end-without-start, nested start-before-end) → `ErrMalformedFence`. Two sentinels suffice because #85 maps *missing* to the append path and *all malformed variants* to the single `core-index-unmanaged` finding; sub-classification is deferred and unnecessary now.

**ParsePageMetadata internals:** local `splitFrontmatter` → `yamladapter.New().Unmarshal(fm, &struct{Type,Title,Description string with yaml tags})` → fallbacks (`Type==""`→`"unknown"`; `Title==""`→`path.Base` minus `path.Ext`; `Description` kept as-is). `Path = relPath`. Any split error or `Unmarshal` error is returned unwrapped-through (`fmt.Errorf("index: parse %s: %w", relPath, err)`); an *absent* `type` on a successful parse is **not** an error (→`"unknown"`), only a genuine YAML parse failure is (→ caller excludes).

---

## Edge cases

- **Fences absent** → `ErrNoFence` (never rewrite; #83 owns append-at-EOF).
- **Duplicate / nested / unterminated / orphan-end fences** → `ErrMalformedFence` (refuse; never guess authoritative region).
- **CRLF markers** → detected (trailing `\r` trimmed on the marker line only; inner bytes taken verbatim).
- **CRLF inside an existing inner region** → shows as "stale" vs the LF-only recompute and self-heals to LF on regen; acceptable and consistent with ADR-011 sub-decision 4 (cross-platform enforcement test deferred to #84).
- **Empty pages slice** → `GenerateIndex` returns `""`; `Reconstruct` yields `…start\n<endmarker>…` (end marker immediately after start's newline).
- **Missing `title`** → filename stem; **missing `description`** → entry has no `— …` suffix; **missing `type`** → `## unknown` section.
- **Unknown/extra frontmatter fields** → ignored by `Unmarshal` (round-trip preservation is not needed here — index is read-only over metadata).
- **Nested subdir paths** sort by full bundle-relative path, not basename (`a/z.md` before `b/a.md`).
- **Input slice order** must not affect output (sort is total on `(Type, Path)`).

---

## Unit test matrix (`internal/index/index_test.go`)

**GenerateIndex**
- Golden format: exact byte-equality against the pinned inline const (title link, em-dash description, omitted-when-empty, stem fallback rendered).
- Idempotency: `GenerateIndex(p) == GenerateIndex(p)` byte-for-byte.
- Input-order independence: shuffled slice → identical output.
- Sort order: types bytewise, then paths bytewise (incl. nested-path case `a/z.md` vs `b/a.md`).
- Grouping: multiple pages of one type under a single header.
- LF only: no `\r` anywhere in output.
- Empty input → `""`.

**ExtractFencedRegion / Reconstruct**
- Happy path: split correct; `Reconstruct(before, fenced, after) == original`.
- `ErrNoFence` on zero markers.
- `ErrMalformedFence` for: two starts, two ends, nested (start,start,end,end), unterminated (start, no end), orphan end (end, no start).
- CRLF markers still detected.
- Human preamble/trailing content byte-preserved through extract→reconstruct.
- Reconstruct with a fresh body: `before + GenerateIndex(pages) + after` places end marker correctly for both non-empty and empty bodies; running it twice is byte-identical (round-trip idempotency).
- Staleness invariant sanity: `bytes.Equal(fenced, []byte(GenerateIndex(pages)))` true when current, false when a page changes (documents the comparison #83/#85 rely on).

**ParsePageMetadata**
- Full frontmatter → all fields + `Path==relPath`.
- Title-stem fallback (missing/empty title).
- Description omitted (missing/empty).
- Type → "unknown" (missing/empty type; **not** an error).
- Missing frontmatter (no leading `---`) → error.
- Unterminated frontmatter → error.
- Malformed YAML → error (caller-excludes semantics).
- Unknown extra fields ignored.
- Subdir relPath: `Path` keeps slash-form; stem fallback uses basename.

**No-model-call invariant:** satisfied structurally — the package imports only stdlib + `yamladapter`. Add a package-doc note asserting no client/model import; the executable "mock panics if called" assertion is #85's (per phase-5-plan), not #82's.

---

## Validation commands

Run from repo root (mirrors `.github/workflows/test.yml`; there is no linter — `go vet` is the only static check):

```bash
gofmt -l internal/index                                   # expect empty output
go build ./...
go vet ./...
go test ./internal/index/... -count=1 -v                  # targeted
go test ./cmd/llm-wiki -run '^TestAcceptance' -count=1     # acceptance gate — confirm no regression
go test ./...                                              # full suite
```

Definition of done for this issue: all six commands clean; the golden-format test locks the region bytes; ADR-011 sub-decisions 2/3/4/8 each map to a passing test.

---

## Commit slices (all on `feat/82-core-index-generation`, one PR)

1. `feat(index): PageMetadata + pure ParsePageMetadata with frontmatter split and fallbacks (ADR-011, #82)` — `metadata.go` + its tests.
2. `feat(index): deterministic GenerateIndex with golden-fixtured region format (ADR-011, #82)` — `GenerateIndex` in `index.go` + golden/idempotency/sort/LF tests.
3. `feat(index): fenced-region extraction, reconstruct, and edge-case errors (ADR-011, #82)` — `ExtractFencedRegion`/`Reconstruct`/sentinels + edge-case tests.

Each commit compiles and its tests pass independently. PR body: `Closes #82`, references ADR-011, fills every template section, cites the exact `go test ./...` result. Merge only after fresh-context adversarial review reaches `READY` (overlay hard gate — no waiver) **and** green CI.

---

## Risks, conflicts & deferrals

**Conflicts flagged (ADR/architecture wins per CLAUDE.md):**
- **API naming, issue vs `phase-5-plan.md`.** The plan lists `ExtractExistingIndex(...)` + `MergeIndex(...)`; the issue lists `ExtractFencedRegion(...)` + `ParsePageMetadata(...)`. Follow the **issue** (authoritative for this issue); `Reconstruct` covers `MergeIndex`'s role. No functional loss.
- **Finding severity.** `phase-5-plan.md` (lines 176–177) says `core-index-stale`/`core-index-unmanaged` default **error**; ADR-011 sub-decision 5 says default **warning** (profile-promotable). **ADR-011 wins.** This is #85's scope, not #82 — flagged here so #85 does not inherit the plan's `error`.
- **Region format under-specification.** The issue's minimal example omits title/description; ADR-011 sub-decision 2 requires them. Richer `[Title](Path) — Description` format chosen and locked behind a fixture (ADR-011 Consequences).
- **Signature refinements (owner-approved in planning):** `GenerateIndex` returns the inner body only (not the marker-wrapped block) so "rewrite only inside the fence" is a clean invariant; `ParsePageMetadata(content, relPath)` is pure (no `path string` disk read) so path-safety stays in fsafe/ADR-005 and exclusion-on-parse-failure lives in the #83 walk.

**Risks:**
- *Byte-idempotency fragility* (trailing newline / blank-line placement) → mitigated by the exact inline golden const + reconstruct-twice test.
- *CRLF churn* in existing inner regions → documented self-heal to LF; cross-platform enforcement deferred to #84.
- *Third copy of `splitFrontmatter`* (validate + plan + index) → minor DRY debt; extraction to a shared helper is out of scope (would touch `validate`/`plan`), noted as a possible follow-up issue.
- *goccy `yaml:` struct tags* → low risk; the struct-decode path is covered by the ParsePageMetadata tests.

**Explicitly deferred (NOT in #82):**
- #83 — CLI `llm-wiki index`, directory walk (via fsafe/`plan.ResolvePage`), `--dry-run` diff, `--json` (ADR-003 envelope, `operation:"index"`), append-at-EOF for `ErrNoFence`, ADR-006 staged write.
- #84 — integration tests, cross-platform (Windows/CRLF) enforcement, 5,000-doc perf, mock-client no-model-call assertion.
- #85 — validate findings `core-index-stale` / `core-index-unmanaged` + severity wiring (default **warning** per ADR-011).
- #86 — `init` fence scaffolding.
- #80 / #81 — enrichment track.
