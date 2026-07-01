# llm-wiki exit codes and JSON contract envelope

This document publishes the stable, frozen surface decided in
[ADR-003](../../design/adr/adr-003-json-contract-and-exit-codes.md): the
versioned JSON envelope shared by every surface, and the numeric exit codes the
`llm-wiki` CLI returns.

> **Stability.** ADR-003 fixed the six semantic buckets and deferred the exact
> numeric exit-code values to the first implementation issue (#1). Those values
> are published here and are now **frozen** — callers (skills, hooks, CI) branch
> on them, so changing one is a breaking contract change. The canonical source
> of truth is `internal/contract` (the named constants and the `ExitCodes`
> table); this document mirrors it and a test enforces that they agree.

## Exit codes

Each code maps 1:1 to one PRD §12 outcome and to one envelope `status` value.

| Code | Bucket (`status`) | Meaning |
| ---- | ----------------- | ------- |
| 0 | `success` | Operation completed with no failing findings. |
| 1 | `success-with-warnings` | Operation completed; non-failing warnings were reported. |
| 2 | `validation-failure` | Validation failed on at least one error-severity finding. |
| 3 | `approval-required` | The operation requires approval before it can proceed. |
| 4 | `invalid-invocation` | The invocation or configuration was invalid. |
| 5 | `system-or-filesystem-failure` | An unexpected system or filesystem error occurred. |

## JSON contract envelope (v1)

Machine surfaces pass `--json` (or set `LLM_WIKI_JSON`) and receive one
envelope shape on every command. Human-readable text is the default for
interactive terminal use.

The envelope always carries exactly these six fields:

| Field | Type | Notes |
| ----- | ---- | ----- |
| `contractVersion` | string | Currently `v1`. |
| `operation` | string | The command that ran, e.g. `version`, `validate`. |
| `status` | string | One of the six bucket values in the table above. |
| `findings` | array | Validation findings; `[]` when none. Each carries `ruleset` (`okf`/`profile`), `severity` (`error`/`warning`/`suggestion`), `code`, `message`, and optional `path` (ADR-004). |
| `affectedPaths` | array | Repository-relative paths the operation touched; `[]` when none. |
| `approval` | object or null | An approval requirement, or `null` when none is involved. |

Example (`llm-wiki validate --json` on a clean no-op skeleton):

```json
{
  "contractVersion": "v1",
  "operation": "validate",
  "status": "success",
  "findings": [],
  "affectedPaths": [],
  "approval": null
}
```
