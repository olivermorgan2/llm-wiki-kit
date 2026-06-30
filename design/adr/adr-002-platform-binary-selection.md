# ADR-002: Ship and select one checksum-verified platform binary

**Status:** proposed
**Date:** 2026-06-30

## Context

The plugin ships a self-contained `llm-wiki` CLI for five platforms
(macOS arm64/x86-64, Linux arm64/x86-64, Windows x86-64) with **no user
runtime**, and release artifacts must be **versioned and checksum-verified**
(PRD §9, §14 Security). Acceptance criteria 2 (correct checksum-verified CLI
runs on every supported platform without runtime setup) and 21 (automated tests
pass on all OS/arch), plus the continuous cross-platform gate, depend on how the
plugin ships and selects the right binary. MVP assumption-locked **one**
binary-selection mechanism (open-questions Q7; addendum 002 A3; build-out-plan
#2). The Codex PRD review flagged packaging as a blocking item to resolve or
lock before issue generation. Risks addressed: cross-platform binary drift and
supply-chain / tampered binary (`knowledge/risks.md`).

**Scope note.** This ADR owns *shipping + selecting + integrity-verifying* the
binary. The broader install / upgrade / uninstall **asset-ownership** question
(plugin-owned vs repo-owned, content preservation — criteria 1, 3, 20) is
ADR-009's domain and is explicitly **not** decided here.

## Options considered

### Option A: Per-platform binaries shipped in the plugin, selected by an OS+arch detector that verifies a checksum before exec

- Pros: fully offline / self-contained — no download at install or run, so
  criterion 2 holds with no network; one deterministic path detects OS+arch and
  resolves a single matching `bin/llm-wiki`; checksums shipped alongside are
  verified before execution (PRD §14); simplest to test on all five platforms
  in CI (criterion 21).
- Cons: larger plugin payload (five binaries bundled); every release rebuilds
  and re-checksums all five; the selection shim is code to keep correct.

### Option B: A thin launcher that downloads the correct binary on first use

- Pros: smaller initial payload; fetches only the needed platform binary.
- Cons: requires network at install/run, breaking the runtime-free,
  works-anywhere promise and complicating criterion 2 in sandboxed CI; adds a
  download-integrity attack surface; harder to guarantee identical offline
  behavior across platforms; contradicts "self-contained" (PRD §1, §9).

## Decision

Adopt **Option A** — ship per-platform binaries in the plugin and select via a
single OS+arch detection path that verifies the artifact checksum before
execution. It is the only option that keeps the engine self-contained and
offline-capable (criterion 2) and makes the five-platform test matrix
(criterion 21) deterministic, while satisfying the checksum-verification
security requirement (PRD §14). MVP commits to exactly this one mechanism, not a
pluggable set; the precise on-disk layout is finalized in the implementation
issue. This advances open-question Q7.

## Consequences

- Easier: criteria 2 and 21 are testable offline; integrity verification has one
  well-defined chokepoint; no runtime network dependency.
- Harder: release engineering must build, checksum, and bundle five binaries per
  release; plugin payload grows.
- Maintain: the OS+arch selection shim, the per-platform build/checksum
  pipeline, and platform integration tests (mitigates the cross-platform-drift
  and supply-chain risks in `knowledge/risks.md`).
- Deferred / validation implications: advances Q7 (record in `knowledge/log.md`).
  Install/upgrade/uninstall asset ownership (criteria 1, 3, 20) is out of scope
  and owned by ADR-009. Criteria 2 and 21 are the acceptance hooks; both are
  mechanism-agnostic (addendum 002 A3), so an override changes only this ADR and
  the packaging issue.
