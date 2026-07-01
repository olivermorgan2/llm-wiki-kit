# Open questions — llm-wiki-kit

Unresolved product/technical decisions. Seeded from PRD §19 "Remaining
product decisions" ([`design/prd.md`](../design/prd.md)). When one is
resolved, record the resolution in [`log.md`](log.md) (and an
ADR in `design/adr/` if architectural), then mark it `closed` here.

Owner is **Oliver** unless noted.

| # | Question | Owner | Status |
|---|---|---|---|
| Q1 | Plugin name, command namespace, license, and marketplace location. (Repo/working name is `llm-wiki-kit`; CLI is `llm-wiki`. License chosen for the repo scaffold = MIT; "license" in PRD §19 also covers the distributed *plugin*'s license/marketplace listing — confirm they're the same.) | Oliver | open |
| Q2 | Exact Go YAML library and supported Go version for development. | Oliver | **closed** — human-accepted **[ADR-001](../design/adr/adr-001-go-toolchain-and-yaml.md) (accepted 2026-07-01)** → Go 1.24.x (conservative current-stable line) + `github.com/goccy/go-yaml` (node-aware, round-trip-preserving; `yaml.v3` is archived). ADR-001 **ratifies** the assumption-lock; exact Go patch pinned in `go.mod` at the first-engine-issue time. See [`log.md`](log.md) (2026-07-01 acceptance). |
| Q3 | Is GitHub Actions the only MVP CI template, or one of several? | Oliver | open · **assumption-locked for MVP** → GitHub Actions only ([addendum 002 A1](../design/prd-addenda/002-mvp-planning-assumptions.md)). Must resolve or keep the lock before `/prd-to-mvp`. |
| Q4 | Exact research-profile templates and conditional-section syntax. | Oliver | open · **assumption-locked for MVP** → minimum contract in [addendum 003](../design/prd-addenda/003-academic-research-profile-contract.md). Must resolve or keep the lock before `/prd-to-mvp`. |
| Q5 | Profile registry and trust model for third-party profiles. | Oliver | open · **scoped to Phase 3** ([addendum 005](../design/prd-addenda/005-custom-profile-boundary.md)); does not block MVP. |
| Q6 | Minimum supported Claude Code version. | Oliver | open · **assumption-locked for MVP** → single version floor, no compat shim ([addendum 002 A2](../design/prd-addenda/002-mvp-planning-assumptions.md)). Must resolve or keep the lock before `/prd-to-mvp`. |
| Q7 | Exact packaging mechanism for selecting the correct platform binary inside the plugin. | Oliver | **closed (binary-selection scope only)** — human-accepted **[ADR-002](../design/adr/adr-002-platform-binary-selection.md) (accepted 2026-07-01)** → one mechanism ([addendum 002 A3](../design/prd-addenda/002-mvp-planning-assumptions.md)): ship per-platform binaries in the plugin + OS/arch-select + checksum-verify before exec. Closes Q7 for **ship/select/verify** only. **Still deferred:** full install/upgrade/uninstall **asset ownership** (criteria 1, 3, 20) → **ADR-009**, and release **signing/provenance** (the trust root against a rebuilt payload) → ADR-009 or a dedicated supply-chain ADR; residual supply-chain risk stays `open` in [`risks.md`](risks.md). See [`log.md`](log.md) (2026-07-01 acceptance). |
| Q8 | JSON contract versioning and compatibility policy. | Oliver | **closed** — human-accepted **[ADR-003](../design/adr/adr-003-json-contract-and-exit-codes.md) (accepted 2026-07-01)** → one versioned JSON envelope (`--json` opt-in on every command); starts at v1, no backward compat until first public release ([addendum 002 A4](../design/prd-addenda/002-mvp-planning-assumptions.md)); six fixed exit-code semantic buckets. **Carry-forward (not a reopened question):** the exact **numeric** exit-code values are deferred to the Phase 1 implementation/spec issue and must be published as a stable code→meaning table **before Phase 1 closes**. See [`log.md`](log.md) (2026-07-01 acceptance). |

## Questions raised during bootstrap

| # | Question | Owner | Status |
|---|---|---|---|
| QB1 | Final product/plugin name vs working repo name `llm-wiki-kit` — lock before issue/PR phases to avoid a later rename. | Oliver | open · **working-name locked for MVP** → `llm-wiki-kit` repo/plugin, `llm-wiki` CLI ([`design/mvp.md`](../design/mvp.md) "Product name"). Override allowed before packaging/public release; it is a bounded rename, not a scope change. |
| QB2 | Codex adversarial PRD review of `design/prd-normalized.md`. | Oliver | **closed** — review ran 2026-06-30, verdict `NEEDS_REVISION` ([archive](reviews/2026-06-30-codex-prd-review.md)). All three blocking findings + non-blocking findings accepted and addressed via [`design/prd-addenda/001`–`005`](../design/prd-addenda/); see [`log.md`](log.md). Q3/Q4/Q6/Q7/Q8 are now assumption-locked (above); Q5 scoped to Phase 3. |
