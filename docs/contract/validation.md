# llm-wiki validation: rulesets, severities, and the baseline boundary

This document describes what `llm-wiki validate` checks on the **core** profile
and how findings resolve to a status and exit code. It mirrors the engine in
`internal/validate` and the contract in `internal/contract`; the decision record
is [ADR-004](../../design/adr/adr-004-validation-and-severity-model.md) (with the
YAML toolchain in [ADR-001](../../design/adr/adr-001-go-toolchain-and-yaml.md) and
the envelope/exit codes in
[ADR-003](../../design/adr/adr-003-json-contract-and-exit-codes.md)).

`validate` runs a **single deterministic pass** over every `*.md` page in the
target directory (default: the current directory), recursing subdirectories in
lexical order. Each page's leading `---`-fenced YAML frontmatter is parsed and
checked. Findings are reported through the JSON envelope (`--json`), each tagged
with its **ruleset** (`okf` or `profile`) so base OKF conformance and profile
conformance are reported **separately**.

## Rulesets and rules

Every finding carries a `ruleset` tag, a `severity`, a stable `code`, a
`message`, and the `path` it applies to.

### OKF ruleset (`ruleset: okf`) — base conformance

| Code | Severity | Fires when |
| ---- | -------- | ---------- |
| `okf-yaml-parse` | error | The frontmatter block is missing, unterminated, or the YAML does not parse. This is the criterion-7 hard error and the **never-suppressible** boundary case. On a parse failure it is the only finding produced for the page. |
| `okf-type-present` | error | `type` is absent or an empty/whitespace string. Unknown `type` **values** are accepted — per-type rules are a later slice. |

### Core-profile ruleset (`ruleset: profile`)

| Code | Severity | Fires when |
| ---- | -------- | ---------- |
| `core-required-title` | error | `title` is absent or an empty/whitespace string. |
| `core-required-description` | error | `description` is absent or an empty/whitespace string. |
| `core-field-type` | error | A modeled field has the wrong YAML type: `title`/`description`/`type` must be scalar strings; `tags`/`aliases` must be sequences. Aggregated into one finding per page. |
| `core-recommended-missing` | suggestion | One or more recommended fields (`timestamp`, `tags`, `aliases`, `resource`) are absent. Aggregated into one finding per page. |
| `core-kebab-filename` | warning | The page filename is not kebab-case (`[a-z0-9]` words joined by single hyphens, ending in `.md`). |

`type` presence is an **OKF** rule; the core profile does not re-report a missing
`type`, keeping the two rulesets non-overlapping for clean separation. A field
that is present but wrong-typed raises `core-field-type` only — never also a
`core-required-*` finding.

## Severity resolution precedence

Effective severity resolves in a fixed order (ADR-004):

1. **Core defaults** — the severities in the tables above, emitted by the engine.
2. **Profile overrides** — a profile may promote or demote a rule within the
   three-severity set (e.g. a warning → error). Overrides change severity
   **only**; they never add or remove findings. No profile file ships in this
   issue; the core-only path applies no overrides.
3. **Baseline suppression** — applied **last**, as a differential filter only
   (see below). Baseline never changes a severity; it only drops findings.

## Baseline boundary

Baseline (adoption) mode lets a pre-existing wiki suppress **already-known**
findings while still surfacing new or worsened ones. A finding's identity for
baseline comparison is its **fingerprint**: the tuple `{ruleset, code, path}`
(severity and message are excluded; no line numbers in this issue). The filter
is bounded by a hard boundary:

- **`okf-yaml-parse` errors are never suppressed** — parsing must succeed before
  any finding or baseline comparison can exist, so criterion 7 holds
  unconditionally.
- **Other error-severity findings** (missing required field, wrong field type)
  are suppressed **only in explicit adoption mode**. Release-gate / CI runs
  always evaluate errors at full severity, so a baselined error is still
  reported there.
- **Warnings and suggestions** are a pure differential filter: suppressed when
  baselined, in either mode.

The **default CLI path loads no adoption baseline** — it runs release-gate
semantics, so structural errors always fail the gate. There is no `--baseline`
flag or on-disk baseline format in this issue; the baseline primitive and its
boundary are built and unit-tested, but the adoption-mode CLI surface is later
work.

## Status and exit codes

After resolution and baseline filtering, the remaining findings reduce to one
envelope `status` (and its [exit code](exit-codes.md)):

| Remaining findings | Status | Exit |
| ------------------ | ------ | ---- |
| any `error` | `validation-failure` | 2 |
| else any `warning` | `success-with-warnings` | 1 |
| else (clean, or `suggestion` only) | `success` | 0 |

Suggestions are advisory and never affect the exit code.
