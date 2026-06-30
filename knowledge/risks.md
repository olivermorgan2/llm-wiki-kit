# Risk register — llm-wiki-kit

Live risks with mitigation and status. Seeded from PRD §16
([`design/prd.md`](../design/prd.md)) plus cross-cutting risks called out
elsewhere in the PRD. Status: `open` (mitigation planned, not yet built),
`mitigated-by-design` (addressed in PRD/architecture), or `closed`.

## Product / design risks (PRD §16)

| Risk | Mitigation | Status |
|---|---|---|
| Profile rules mistaken for OKF rules | Report OKF and profile conformance separately; never present a profile rule as universal OKF. | open |
| Schema becomes rigid too early | Keep `core` conservative; prefer recommendations before requirements. | open |
| Hooks treated as a security boundary | Position CI as authoritative; document hook limitations; post-write hooks are feedback, not prevention. | open |
| Model overwrites human work | Preview existing-page changes; preserve unknown/unrelated content. | open |
| Sourced claims lose provenance | Require resolvable citations; preserve existing citations. | open |
| Custom-profile complexity becomes a programming language | Declarative data only; one inheritance level; explicit validation. | open |
| Cross-platform binaries drift | Automate release builds, checksums, and platform integration tests. | open |
| Compilation drops important research facts | Track evidence gaps; keep sources addressable; deeper evaluation later. | open |
| Upgrades overwrite customization | Track ownership; preview repository changes; preserve local profile extensions. | open |
| Skills bypass deterministic safety | Route managed mutations through staged engine plans; narrow tool access. | open |
| Skill adapters duplicate core logic | All policy/mutation in shared engine; test adapters as integrations only. | open |
| A reviewed diff becomes stale before application | Bind plans to source/target hashes; reject stale plans. | open |

## Cross-cutting risks (from elsewhere in the PRD)

| Risk | Mitigation | Status |
|---|---|---|
| Supply chain / tampered binary | Versioned, checksum-verified release artifacts; verify checksums before executing bundled CLI (PRD §9, §14 Security). | open |
| Path traversal / symlink escape | Engine canonicalizes & validates paths; resolves symlinks; rejects escapes; bounded write scope (FR10, §14). | open |
| Untrusted source material as instructions | Treat imported content as data, never as agent instructions; repo content cannot override plugin/user instructions (FR10, §14 Security). | open |
| Stale mutation plan applied after target changed | `apply` rechecks hashes and refuses stale/modified plans without mutating (FR7, FR11). | open |
| Partial multi-file writes on interruption | Atomic writes; no partial updates across multi-file ops; safe recovery from interruption (FR1, FR10, §14 Reliability). | open |
| Divergent findings across skills/hooks/CI/CLI | One deterministic implementation shared by all four; acceptance criterion requires identical findings for identical state (§17). | open |

## Process risk (this collaboration)

| Risk | Mitigation | Status |
|---|---|---|
| Normalized PRD drifts from source intent | Adversarial Codex review of `design/prd-normalized.md`; findings captured as `design/prd-addenda/` + decisions log before MVP scoping. | open |
