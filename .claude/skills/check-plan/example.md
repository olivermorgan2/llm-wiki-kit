# check-plan — worked example

Three short walk-throughs: a standalone pass, a standalone fail, and
a chained-mode fail-then-iterate.

---

## 1. Standalone pass

### Input

`/check-plan design/adr/adr-035-state-md-session-continuity.md`

### Skill behaviour

- Type detection: path matches `design/adr/adr-*.md` → `adr`.
- `criteria.md` loaded; ADR table selected.
- Each criterion runs:
  - `ADR-C1` — Context, Decision, Consequences all present and
    non-empty. PASS.
  - `ADR-C2` — Options A, B, C, D found under Options considered.
    PASS.
  - `ADR-C3` — Each option has both Pros and Cons. PASS.
  - `ADR-C4` — Decision body says "Adopt **Option A**", which
    matches `### Option A:`. PASS.
  - `ADR-C5` — `ADR-031`, `ADR-032`, `ADR-038` all referenced; all
    files exist. PASS (no warnings).
  - `ADR-C6` — Decision compared against accepted ADRs; no textual
    contradiction surfaced. PASS (no warnings).

### Output

```
✅ check-plan PASS — design/adr/adr-035-state-md-session-continuity.md
   No warnings.
```

The user proceeds confidently.

---

## 2. Standalone fail

### Input

A hand-crafted broken ADR at `/tmp/broken-adr.md`:

```markdown
# ADR-099: Use SQLite for everything

**Status:** proposed

## Context

We need a database.

## Options considered

### Option A: SQLite

- Pros: simple, embedded, ubiquitous.

### Option B: Postgres

- Pros: scalable, mature.
- Cons: heavier ops.

## Decision

Use SQLite.

(no Consequences section)
```

`/check-plan /tmp/broken-adr.md`

### Skill behaviour

- Type detection: body has `# ADR-099:` and `**Status:**`. → `adr`.
- ADR criteria run:
  - `ADR-C1` — Consequences section is missing. **FAIL** (deterministic).
  - `ADR-C2` — Two options found. PASS.
  - `ADR-C3` — Option A is missing a Cons line. **FAIL** (deterministic).
  - `ADR-C4` — Decision says "Use SQLite" — does not match `### Option A:` exactly. **FAIL** (deterministic).
  - `ADR-C5` — No `ADR-NNN` tokens in body. PASS.
  - `ADR-C6` — No conflict with accepted ADRs. PASS.

### Output

```
❌ check-plan FAIL — /tmp/broken-adr.md
- ADR-C1 (deterministic): Consequences section is missing. Add it after the Decision; the template orders Context → Options → Decision → Consequences.
- ADR-C3 (deterministic): Option A is missing a `Cons:` line. Add at least one Cons bullet to Option A.
- ADR-C4 (deterministic): Decision body does not name an option from Options considered. Either say "Adopt **Option A**" / "Adopt **Option B**" explicitly, or rename the chosen option to match.
```

The user revises and re-runs.

---

## 3. Chained mode — fail then iterate then pass

`prepare-issue` is filling a prompt for issue #99. After the user
confirms the rendered prompt, `prepare-issue` invokes `/check-plan`
in chained mode against the in-memory rendering.

### Round 1

Rendered prompt is missing the `Acceptance criteria` section and
still has one `<!-- TODO: fill in -->` marker.

`/check-plan` returns:

```
❌ check-plan FAIL — prompts/issue-099-some-feature.md (round 1/3)
- PROMPT-C1 (deterministic): Acceptance criteria section is empty. Add at least one observable end-state outcome.
- PROMPT-C3 (deterministic): Unresolved TODO at line 47 (`{{PROJECT_SPECIFIC_CONSTRAINT_OR_DELETE_THIS_LINE}}`). Fill it in or delete the line.
```

`prepare-issue` surfaces the failures to the user and asks how to
revise. The user supplies an Acceptance criteria block and confirms
the TODO line should be deleted. `prepare-issue` applies both edits
in memory.

### Round 2

`/check-plan` re-runs. All deterministic checks pass. `PROMPT-C5`
(single-PR scope heuristic) fires a warning — Requirements + Scope
combined are 31 bullets, slightly over the soft cap.

### Output (chained mode return value)

```
✅ check-plan PASS — prompts/issue-099-some-feature.md (round 2/3)
   Warnings:
   - PROMPT-C5 (warning): Requirements + Scope combined are 31 bullets (soft cap ~25). Consider splitting if these aren't tightly cohesive.
```

`prepare-issue` proceeds to step 11 (write the file). The warning is
surfaced to the user but does not block.

### What `--skip-check` would have done

If the user had invoked `prepare-issue --skip-check` from the start,
none of this would have happened. `prepare-issue` would have written
the prompt with the TODO and missing Acceptance criteria, leaving a
one-line breadcrumb in the prompt body:

```
<!-- /check-plan was skipped via --skip-check per ADR-034 -->
```

The user is the auditor in that mode.

---

## 4. Two surfaces, one source of truth (per ADR-043)

The skill's evaluation logic is implemented as `bin/check-plan` so
that skills with chain points (`adr-writer`, `prepare-issue`) can
invoke it deterministically as a subprocess rather than chaining
to the slash-command. Both surfaces produce identical pass/fail
output for the same input — the slash-command is a wrapper that
adds artefact-type detection, structured rendering for the
operator, and the iterative-with-user revision loop.

### Same input, both surfaces

```bash
bin/check-plan --criteria-set adr \
  --input design/adr/adr-041-auto-mode-permission-contract.md
```

Output:

```
PASS  ADR-C1  Context, Decision, Consequences sections present and non-empty
PASS  ADR-C2  Options considered has 4 Option blocks
PASS  ADR-C3  Every Option block has Pros: and Cons: lines
PASS  ADR-C4  Decision names Option A
PASS  ADR-C5  All ADR-NNN tokens resolve to files in design/adr/
WARN  ADR-C6  Semantic-conflict check deferred (best-effort substring per ADR-034 not implemented in v1)
```

Exit 0 (no deterministic failures; warnings surfaced).

The slash-command equivalent — `/check-plan
design/adr/adr-041-auto-mode-permission-contract.md` — runs the
same evaluation under the hood (it invokes `bin/check-plan`
internally) and renders the identical results, framed for the
operator with the file name and a short summary.

### JSON for skill consumption

```bash
bin/check-plan --criteria-set adr \
  --input design/adr/adr-041-auto-mode-permission-contract.md \
  --format json
```

Returns a JSON envelope:

```json
{
  "criteria-set": "adr",
  "result": "pass",
  "criteria": [
    {"id":"ADR-C1","severity":"deterministic","status":"pass","message":"Context, Decision, Consequences sections present and non-empty","remediation":""},
    {"id":"ADR-C2","severity":"deterministic","status":"pass","message":"Options considered has 4 Option blocks","remediation":""},
    {"id":"ADR-C3","severity":"deterministic","status":"pass","message":"Every Option block has Pros: and Cons: lines","remediation":""},
    {"id":"ADR-C4","severity":"deterministic","status":"pass","message":"Decision names Option A","remediation":""},
    {"id":"ADR-C5","severity":"deterministic","status":"pass","message":"All ADR-NNN tokens resolve to files in design/adr/","remediation":""},
    {"id":"ADR-C6","severity":"warning","status":"warn","message":"Semantic-conflict check deferred (best-effort substring per ADR-034 not implemented in v1)","remediation":"Reviewer should manually check the Decision against accepted ADRs in the same area; revisit when a structural rule lands."}
  ]
}
```

`adr-writer` and `prepare-issue` consume this JSON shape from
their chain points; the iterative-with-user revision loop sits in
the producer skill, not in `bin/check-plan` (which is non-
interactive by construction — see ADR-043 *Implementation shape*
4).

### Stdin support

`bin/check-plan --input -` reads the artefact from stdin, which
is how chained producers invoke it without writing a temp file:

```bash
producer_renders_adr_in_memory | \
  bin/check-plan --criteria-set adr --input - --format json
```

### Reserved criteria-set names

`changelog`, `milestone-summary`, and `pr-body` are reserved
criteria-set names but their evaluators are not yet implemented.
Invoking them returns exit 2 with a clear message; the
implementing follow-up issues are tracked separately.
