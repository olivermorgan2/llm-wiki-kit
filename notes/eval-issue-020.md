# Evaluation — Issue #20: New-repo + non-empty-repo install/init acceptance corpus

**Branch:** `issue-020-acceptance-corpus` (cut from synced `main`).
**PR:** [#27](https://github.com/olivermorgan2/llm-wiki-kit/pull/27).
**ADRs:** [ADR-002](../design/adr/adr-002-platform-binary-selection.md) (five
supported platforms), [ADR-003](../design/adr/adr-003-json-contract-and-exit-codes.md)
(envelope + exit codes), [ADR-005](../design/adr/adr-005-safe-filesystem-layer.md)
(`.llm-wiki/` boundary), [ADR-006](../design/adr/adr-006-staged-mutation-transaction-model.md)
(staged mutation; refusal precedes `txn.Begin`),
[ADR-009](../design/adr/adr-009-install-upgrade-uninstall-ownership.md)
(version-record manifest).
**Method:** durable, criterion-traceable acceptance corpus over the existing
in-process CLI harness; named per-platform CI gate step.

Backs the **Phase 2 (Install / init)** exit gate
(`design/build-out-plan.md` §"Phase 2"), whose exit criteria are: acceptance
criteria **2**, **3**, **4 (core)** pass on all five platforms, plus criterion
**1** (install half of the documented distribution flow) demonstrated.

## What this delivers

The install/init behaviors already shipped in #24/#25 with scattered
per-command unit tests. What was missing is a **named, criterion-mapped
acceptance corpus** that runs as a legible per-platform gate step, plus this
gate-evidence artefact. #20 adds exactly that — no product behavior changes.

## What changed

- **`cmd/llm-wiki/acceptance_test.go`** *(new)* — six `TestAcceptance`-prefixed
  tests, each doc-commented with the criterion/ADR it proves, reusing the
  existing harness (`exec`, `decodeEnvelope`, `snapshotTree`, `sha256Hex`,
  `installTargets`, `initTargets`). New helpers `seedTree` (nested paths) and
  `seedNonEmptyRepo` (hermetic fake git repo seeded with `\n` literals into
  `t.TempDir()` — platform-identical bytes, no real `git init`).
- **`.github/workflows/test.yml`** — one named step in the `test` job,
  `go test ./cmd/llm-wiki -run '^TestAcceptance' -count=1 -v`, per matrix leg.
  `cross-compile-smoke` untouched.
- **`prompts/issue-020-acceptance-corpus.md`** *(new)* — the session prompt.
- **`notes/eval-issue-020.md`** *(new)* — this evidence artefact.

## Criteria → test mapping

| Criterion | Test | Journey | Proves |
|---|---|---|---|
| **2** — CLI runs on every platform | `TestAcceptanceCriterion2VersionRunsOnHostPlatform` | `version --json` | exit 0, six-field ADR-003 envelope. The "runs" half; the checksum/cross-compile half is the `cross-compile-smoke` job. |
| **3** — install into a new repo | `TestAcceptanceCriterion3InstallNewRepoThenValidateClean` | empty dir → `install --json` → `validate --json` | four `installTargets` written; ADR-009 manifest decoded (schema/plugin/cli/okf/profile `core`, exactly 3 plugin-owned assets, hashes re-derived from on-disk bytes, never self-listed); bundle validates clean. |
| **3** — non-empty repo, no file loss | `TestAcceptanceCriterion3InstallNonEmptyRepoNoFileLoss` | `seedNonEmptyRepo` → snapshot → `install --json` → snapshot | every seeded file (incl. `.git/*`) byte-identical; tree grows by exactly the 4 scaffold files; manifest catalogues only the 3 scaffold assets, no user files. |
| **3** — collision refusal is zero-mutation | `TestAcceptanceCriterion3CollisionRefusalIsZeroMutation` | seed + user `llm-wiki.yaml` & `wiki/index.md` → snapshot → `install --json` → snapshot | exit 3; approval lists exactly the 2 conflicts (sorted, slash-form); whole tree byte-identical; `.llm-wiki/` absent (refusal precedes `txn.Begin`, ADR-006). |
| **1** — install half (dry-run) | `TestAcceptanceCriterion1InstallDryRunFullPlanNoOp` | `seedNonEmptyRepo` → snapshot → `install --dry-run --json` → snapshot | exit 0; envelope lists the full 4-path plan; tree byte-identical; `.llm-wiki/` absent (dry-run returns before `txn.Begin`). |
| **4 (core)** — init with core profile | `TestAcceptanceCriterion4InitCoreThenValidateClean` | empty dir → `init --json` → `validate --json` | envelope `affectedPaths` = exactly the 3 `initTargets`, each present on disk; **no** ADR-009 version-record manifest (init ≠ install; the shared ADR-006 txn layer may still leave an empty `.llm-wiki/` working area, so the corpus asserts manifest-absence, not an exact whole-tree snapshot); bundle validates clean. Core profile only (academic-research = Phase 4, custom = Phase 7). |

## Verification

Evidence is separated into five categories. **This PR does not prove
"all five platforms green" or "full matrix green."** Read each category on
its own terms.

### 1. Local validation

Go 1.24.13, `GOTOOLCHAIN=local`, `/Users/hermes/sdk/go1.24.13/bin/go`, on a
Unix (developer macOS/arm64) host:

```
gofmt -l .                                                       # empty
go vet ./...                                                     # clean
go build ./...                                                   # clean
go test ./...                                                    # all packages pass on this host
go test ./cmd/llm-wiki -run '^TestAcceptance' -count=1 -v        # 6/6 PASS
```

Note `go test ./...` passes locally because the host is Unix; the
file-permission-mode assertions that fail on Windows (category 4) hold here.
Local validation is single-host and is **not** evidence about any other
platform.

### 2. Named acceptance-step evidence (CI)

The `Acceptance corpus (Phase 2 gate evidence)` CI step runs
`go test ./cmd/llm-wiki -run '^TestAcceptance' -count=1 -v` before the full
suite, per matrix leg. Latest observed evidence — run
[28648011244](https://github.com/olivermorgan2/llm-wiki-kit/actions/runs/28648011244)
(branch tip `2ac87a6`; each push re-runs the matrix, so consult the newest
run on PR #27 for the current tip):

| Platform (ADR-002) | Acceptance step | Runner |
|---|---|---|
| linux-amd64 | **6/6 PASS** | ran |
| linux-arm64 | **6/6 PASS** | ran |
| macos-arm64 | **6/6 PASS** | ran |
| windows-amd64 | **6/6 PASS** | ran |
| macos-amd64 | not run | runner unavailable (category 5) |

The acceptance corpus is green on **all four platforms where a runner executed
it (4 of 5), including Windows**. It has **not** been observed on macos-amd64.

### 3. Full CI matrix status — NOT green

| Platform (ADR-002) | `test` job (full `go test ./...`) |
|---|---|
| linux-amd64 | green |
| linux-arm64 | green |
| macos-arm64 | green |
| windows-amd64 | **RED** — pre-existing `internal/` tests (category 4) |
| macos-amd64 | did not run (category 5) |
| `cross-compile-smoke` (5 target binaries + checksum manifest) | green |

Three platforms are fully green. Windows' full job is red on unrelated
pre-existing tests. macos-amd64 produced no full-suite evidence. The full
matrix is therefore **not** green.

### 4. Known pre-existing failures (not introduced by #20)

The windows-amd64 full `go test ./...` is red **only** in `internal/fsafe`
and `internal/txn`, on Unix file-permission-mode assertions
(`mode = -rw-rw-rw-, want -rw-r-----`, etc.) that cannot hold on Windows.
These fail identically on `main`'s own CI run
([28646677901](https://github.com/olivermorgan2/llm-wiki-kit/actions/runs/28646677901),
headSha `fb65639`, the commit this branch was cut from) — they predate #20 and
live in `internal/`, which this issue's scope forbids touching (owned by the
txn/fsafe issues). Because the acceptance step targets only `./cmd/llm-wiki`
and is ordered **before** the full suite (a declared deviation — see below), it
still emits green Windows evidence (6/6 PASS, category 2) even though the job
ends red.

### 5. Unavailable-runner caveat (macos-amd64)

The macos-amd64 (`macos-13`) leg did not dispatch: the runner did not become
available in this environment. It stayed queued/never-ran here, was cancelled
by the concurrency guard on prior pushes, and is likewise queued on `main`'s
own run. This is runner-availability latency, not a code issue — but it means
macos-amd64 has produced **no** acceptance evidence and **no** full-suite
evidence. Its code path is identical to macos-arm64 (fully green, 6/6
acceptance) and `cross-compile-smoke` builds the `darwin/amd64` binary cleanly;
that is supporting **inference**, not observed per-platform CI.

Both the Windows redness (category 4) and the macos-amd64 gap (category 5) are
for Hermes to weigh at the Phase 2 gate; neither is a #20 regression.

## Deviation from the plan (declared)

The plan placed the named acceptance step **after** the existing `go test ./...`
step. On the Windows leg that ordering means the acceptance step never runs,
because the pre-existing `internal/` permission failures fail `go test ./...`
first and abort the job — leaving zero per-platform acceptance evidence for
Windows, which defeats a core purpose of the issue. The step is therefore run
**before** `go test ./...`. It targets only `./cmd/llm-wiki` (not the failing
`internal/` packages), so the acceptance evidence is produced on every leg
regardless of the unrelated suite's state. Behavior and assertions are
otherwise exactly as planned.

One assertion was also corrected against reality: the plan expected `init` to
create **no `.llm-wiki/`**. In fact the shared ADR-006 transaction layer leaves
an empty `.llm-wiki/{staging,tmp}` working area for both `init` and `install`,
so the corpus asserts the true ADR-009 distinction — `init` writes **no
version-record manifest** — instead of the (incorrect) absence of `.llm-wiki/`.

## Non-goals honored

- **No authoring / enrichment fixtures** — install/init lifecycle only.
- **No Phase 3+ inspect/plan/apply fixtures.**
- **No deep transaction-interruption re-testing** — the crash/recovery matrix is
  owned by `internal/txn` (#23); the corpus asserts refusal/dry-run
  zero-mutation at the CLI seam only.
- **Core profile only** — academic-research (Phase 4) and custom (Phase 7) are
  out of scope.
- **No release builds (#17), no upgrade/uninstall, no closeout.**
- **`design/state.md` untouched** — the state refresh is owned by **#21**.

## Pointer to closeout

Phase 2 closeout (milestone close + `design/state.md` refresh) is **#21**.
This corpus and evidence note are the gate inputs; #21 consumes them.

## Commands to reproduce

```bash
export GOTOOLCHAIN=local
GO=/Users/hermes/sdk/go1.24.13/bin/go
cd /Users/hermes/llm-wiki-kit
"$GO" version   # go1.24.13
gofmt -l . && "$GO" vet ./... && "$GO" build ./... && "$GO" test ./...
"$GO" test ./cmd/llm-wiki -run '^TestAcceptance' -count=1 -v
```
