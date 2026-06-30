# PRD: Claude Code Kit for OKF Wikis

**Status:** MVP definition  
**Last updated:** 2026-06-20

## 1. Product definition

The Claude Code Kit for OKF Wikis is a versioned Claude Code plugin for creating and maintaining portable, repository-native knowledge bundles based on the Open Knowledge Format (OKF).

The plugin combines:

- Claude Code skills as the user-facing workflow and semantic reasoning layer.
- A bundled, self-contained Go engine, exposed through the `llm-wiki` CLI, for all deterministic scaffolding, validation, indexing, profile management, diff planning, and safe filesystem operations.
- Optional Claude Code and Git hooks for local guardrails.
- Ready-to-use CI configuration for authoritative repository enforcement.
- Versioned profiles that define domain-specific page types, templates, and validation policy.

Skills must delegate deterministic work to the shared engine rather than reimplementing it in prompts or skill-specific scripts. The same engine is used by skills, hooks, CI, and direct CLI workflows.

The wiki remains plain Markdown and YAML. It must remain readable, editable, and portable without Claude Code or the kit.

## 2. Problem

Teams that build agent systems often keep useful knowledge across documentation, papers, source code, schemas, notes, dashboards, and individual memory. These materials are difficult for agents to consume consistently because they lack stable structure, explicit relationships, and predictable provenance.

OKF supplies a deliberately minimal interchange format, but it does not provide an opinionated authoring workflow, domain templates, deterministic validation beyond base conformance, or repository maintenance automation. Teams must otherwise recreate these conventions for every wiki.

## 3. Product vision

The kit should let a user create a useful LLM wiki quickly, keep its structure trustworthy as it grows, and adapt it to a specific domain without sacrificing OKF portability.

Claude Code acts as an authoring and maintenance assistant. Skills orchestrate the user workflow and semantic work; the Go engine remains the deterministic authority for structure, validation, diffs, and writes. The repository remains the source of truth.

## 4. Target users

### Primary MVP user

A technically proficient researcher or technical team already using Claude Code and Git that wants durable, inspectable, agent-readable knowledge without adopting a hosted wiki or proprietary database.

### Initial domain user

An academic or scientific literature researcher who needs to organize sources, concepts, claims, methods, entities, and syntheses with explicit provenance.

### Secondary users

- AI engineers maintaining agent context.
- Software, data, analytics, and operations teams.
- Technical writers and open-source maintainers.
- Teams that want to define their own domain-specific profile.

## 5. Jobs to be done

- When starting a wiki, provide a valid repository structure, default profile, templates, instructions, and validation without runtime setup.
- When authoring knowledge, help create or enrich correctly structured pages while preserving existing human work.
- When using source material, preserve a traceable distinction between evidence, inference, and unsupported statements.
- When editing directly, provide the same validation contract used by model-assisted workflows.
- When a domain needs stronger conventions, allow a custom profile without changing or recompiling the CLI.
- When the kit evolves, upgrade it without silently overwriting repository customizations.

## 6. Design principles

### Format, not platform

Core content is plain Markdown with YAML frontmatter. No proprietary runtime is required to read it.

### OKF baseline plus explicit profiles

Base OKF conformance and stricter kit-profile conformance are separate results. The product must never present a profile rule as a universal OKF requirement.

### Deterministic enforcement

Filesystem safety, parsing, validation, indexing, and profile resolution belong in the Go CLI. Model judgment must not be described as deterministic validation.

### One deterministic implementation

Skills, hooks, CI, and direct CLI use must call the same Go engine. Skill-specific scripts may be thin adapters, but must not duplicate validation, indexing, profile resolution, approval policy, or filesystem mutation logic.

### Lightweight skill context

`SKILL.md` files should contain only workflow, judgment criteria, and resource-routing instructions. Templates, examples, and reference material should remain supporting files loaded on demand. Executable deterministic logic should not be embedded in prompts.

### Review proportional to risk

Routine creation should be fast. Updates to existing pages should be previewed. Renames, deletions, migrations, and broad edits require explicit approval.

### Human and machine legibility

Users may edit wiki files directly. Unknown valid fields and unrelated human-authored content must survive kit operations.

### Provenance without citation theatre

Model-generated claims derived from identifiable sources must cite those sources. Human-authored statements do not require citations unless the active profile says otherwise.

## 7. Standards and compatibility

### OKF compatibility

The MVP targets OKF v0.1 and must pin the target version in the bundle configuration.

Base conformance follows the [official OKF specification](https://github.com/GoogleCloudPlatform/knowledge-catalog/blob/main/okf/SPEC.md):

- Concept documents are UTF-8 Markdown with parseable YAML frontmatter.
- Every concept document has a non-empty `type`.
- Reserved `index.md` and `log.md` files follow OKF rules when present.
- Unknown types and additional fields are accepted.
- Broken links do not invalidate base OKF conformance.

The kit uses bundle-root-relative Markdown links as canonical output while accepting valid relative links.

### Claude Code compatibility

The product is distributed as a versioned plugin because plugins are Claude Code's reusable, updateable distribution mechanism for skills, hooks, and executables. See the [Claude Code plugin documentation](https://code.claude.com/docs/en/plugins).

Project instructions guide behavior but are not enforcement. Skills contain task-specific procedures and may execute supporting scripts without loading those scripts into model context. Plugin executables under `bin/` are available to skill workflows. Hooks and the shared Go engine provide deterministic controls. See the official documentation for [project instructions](https://code.claude.com/docs/en/memory), [skills](https://code.claude.com/docs/en/skills), and [hooks](https://code.claude.com/docs/en/hooks).

## 8. MVP scope

### In scope

- Versioned Claude Code plugin.
- Bundled Go CLI requiring no separate runtime.
- Structured skill-to-engine contracts and staged plan/apply workflows.
- Installation into new and existing Git repositories.
- One configurable wiki bundle per repository.
- Base OKF validation and stricter profile validation.
- Generic core profile.
- Academic and scientific research profile extending core.
- Custom-profile template and scaffolding workflow.
- Authoring and enrichment skills.
- Deterministic index maintenance.
- Provenance requirements for sourced model-generated claims.
- Risk-based preview and approval behavior.
- Optional local hooks and ready-to-use CI enforcement.
- Upgrade, dry-run, doctor, and uninstall workflows.

### Out of scope

- Hosted SaaS or proprietary wiki UI.
- Real-time collaboration or enterprise permissions.
- Embeddings, vector storage, or semantic search infrastructure.
- External-system ingestion connectors.
- Automated schema evolution or migrations.
- Semantic contradiction and near-duplicate detection as deterministic checks.
- Multiple active bundles in one repository.
- Cross-bundle linking.
- Windows ARM64.
- Static publishing and graph visualization.

## 9. Supported platforms

The plugin must ship a self-contained CLI for:

- macOS ARM64.
- macOS x86-64.
- Linux ARM64.
- Linux x86-64.
- Windows x86-64.

Users must not need to install Go, Python, Node.js, or third-party libraries. Release artifacts must be versioned and checksum-verified.

## 10. Repository architecture

### Plugin package

The plugin package should be equivalent to:

```text
plugin/
├── .claude-plugin/
│   └── plugin.json
├── bin/
│   └── llm-wiki                  # Platform-compatible Go executable
├── skills/
│   ├── init/
│   │   └── SKILL.md
│   ├── author/
│   │   ├── SKILL.md
│   │   ├── templates/
│   │   └── examples/
│   ├── enrich/
│   │   └── SKILL.md
│   └── profile/
│       ├── SKILL.md
│       └── references/
├── hooks/
│   └── hooks.json
└── profiles/
    ├── core/
    ├── academic-research/
    └── custom-template/
```

Skills are the primary interactive interface. They invoke `llm-wiki` subcommands rather than editing managed wiki files directly when the engine can perform the operation.

### Initialized repository

The installed repository structure should be equivalent to:

```text
.claude/
  settings.json                 # Optional project hook integration
wiki/
  index.md
  concepts/
  entities/
  sources/
  templates/
  schema/
    profile.yaml                # Resolved machine-readable profile
    current.md                  # Human-readable profile documentation
    changelog.md
.llm-wiki/
  config.yaml
  install-manifest.json         # Kit-owned files and installed versions
CLAUDE.md
README.md
```

Plugin skills, hooks, and executables remain plugin-owned rather than being copied into every repository unless required by Claude Code's project configuration model.

The bundle root defaults to `wiki/` but is configurable. MVP supports one active bundle per repository; internal APIs must not assume the literal path `wiki/` so multi-bundle support remains possible later.

## 11. Profile model

Profiles are versioned, data-driven policies layered on top of OKF. They define page types, directories, templates, fields, sections, citations, naming, and validation severities.

### Inheritance

MVP supports one inheritance level:

```yaml
profile:
  id: academic-research
  version: "1.0"
  extends: core
```

Every custom profile extends `core`. Multiple inheritance and chained custom-profile inheritance are out of scope for MVP.

Profile resolution must be deterministic. Conflicts, invalid overrides, and incompatible versions must fail with actionable messages.

### Core profile

The generic core profile is selected when the user does not choose a more specific profile. It defines a conservative common contract:

- Required `type`, `title`, and `description`.
- Recommended `timestamp`, `tags`, `aliases`, and `resource`.
- Kebab-case filenames.
- Bundle-root-relative Markdown links.
- Preservation of unknown fields.
- Common page types: `concept`, `entity`, `source`, and `synthesis`.
- Common sections are recommended rather than required unless needed for structural integrity.

### Academic research profile

The research profile extends core and adds:

- `claim`: a proposition with supporting evidence, counterevidence, confidence, and assessment.
- `method`: a research or analytical method with assumptions, strengths, and limitations.
- `question`: an open research question and evidence-gap record.
- Research-specific `source` fields such as authors, publication date, DOI or canonical URL, source type, and review status.
- Research-specific `synthesis` sections for scope, findings, agreement, disagreement, and evidence gaps.

The profile must distinguish evidence-bearing pages from workflow state and derived outputs. A question or generated output must not silently become evidence for another claim.

### Custom-profile template

The plugin ships a documented template containing:

```text
profiles/custom-template/
  profile.yaml
  templates/
  examples/valid/
  examples/invalid/
  README.md
```

The CLI supports:

```text
llm-wiki profile create <id>
llm-wiki profile validate <path>
llm-wiki init --profile <id-or-path>
```

Custom profiles must not require recompiling the CLI.

## 12. Skill-to-engine contract

The plugin must define a stable, documented contract between skills and the Go engine.

### Invocation

- Skills invoke the `llm-wiki` executable made available by the plugin.
- Skill-specific scripts are permitted only as thin adapters for assembling inputs or presenting outputs.
- Core behavior must remain callable directly from the CLI without Claude Code.
- Raw user arguments must never be interpolated into shell commands.

### Structured data

- Machine-oriented commands accept or emit versioned JSON where structured exchange is required.
- Human-readable output remains the default for direct terminal use.
- Every structured response identifies contract version, operation, status, findings, affected paths, and any required approval.
- File paths passed between skills and the engine are canonicalized and validated by the engine.

### Exit codes

The CLI must document stable exit codes for at least:

- Success.
- Success with warnings.
- Validation failure.
- Approval required.
- Invalid invocation or configuration.
- System or filesystem failure.

### Tool access

Skills should use the narrowest practical tool set. Where supported, mutating skills should prefer reads plus approved `llm-wiki` commands over unrestricted shell or direct writes to managed paths.

## 13. Functional requirements

### FR1. Installation and initialization

The plugin and CLI must:

- Detect the supported operating system and architecture.
- Verify the bundled CLI version and checksum.
- Install into new or non-empty Git repositories.
- Offer core, academic-research, or custom profile selection.
- Show all proposed filesystem changes.
- Support `--dry-run`.
- Refuse silent overwrites.
- Record plugin, CLI, OKF, and profile versions.
- Recover safely from interruption.

### FR2. Upgrade and uninstall

Upgrades must distinguish plugin-owned assets from repository-owned configuration and content.

The CLI must:

- Detect available structural upgrades.
- Preview changes before applying them.
- Preserve local profile extensions and wiki content.
- Require approval for migrations or incompatible profile changes.
- Record upgrade history.
- Support removal of kit-owned repository configuration without deleting wiki content.

### FR3. Authoring skill

The explicitly invocable authoring skill must:

- Read the resolved profile and selected type template.
- Search for exact and likely existing concepts before creating a page.
- Accept notes, repository files, URLs already supplied by the user, or direct instructions.
- Fill profile fields and sections without fabricating unavailable facts.
- Add useful internal links.
- Place model-drafted content in an engine-approved staging location or submit it through a structured engine input.
- Ask the engine to validate the draft, calculate affected files, and generate a reviewable plan.
- Update the relevant index through the engine.
- Add citations for model-generated sourced claims.
- Label inference and uncertainty.
- Preserve unknown fields and unrelated existing content.
- Validate and show the engine-generated diff.

New pages may be written before showing the diff. Changes to existing pages must be previewed before application.

The authoring skill must not apply a stale plan. Plans must be bound to hashes of the source and target state they were calculated from.

### FR4. Enrichment skill

The enrichment skill must:

- Preserve path-based concept identity.
- Preview changes to existing pages.
- Improve structure, descriptions, relationships, caveats, or provenance.
- Avoid optimizing for link count.
- Preserve citations and unknown fields.
- Update `timestamp` only after a meaningful content change.
- Never rename, delete, or merge pages without explicit approval.
- Use the same staged plan/apply workflow as authoring for every existing-page mutation.

### FR5. Provenance

When model-generated content derives from identifiable material, the workflow must:

- Cite a resolvable URL, repository path, or OKF document.
- Preserve existing citations.
- Distinguish evidence from inference.
- Never invent a citation or imply a source was read when it was not.
- Use `resource` for the underlying asset identity, not as a substitute for claim-supporting citations.

Human-authored knowledge may remain uncited unless the active profile requires otherwise.

### FR6. Validation

The Go CLI must support:

```text
llm-wiki validate
llm-wiki validate --okf-only
llm-wiki validate --changed
llm-wiki validate --baseline <file>
llm-wiki validate --format json
```

It must report base OKF conformance separately from profile conformance.

Deterministic checks include:

- UTF-8 and YAML parsing.
- Required field presence and configured field types.
- OKF reserved-file rules.
- Timestamp format.
- Profile type registration.
- H1/title consistency when configured.
- Required sections.
- Internal link resolution.
- Exact duplicate paths and titles.
- Index consistency.
- Profile and configuration validity.

Semantic similarity, factual correctness, contradiction analysis, and source-quality assessment are model-assisted operations, not validator errors.

### FR7. Page inspection and mutation planning

The shared engine must support the equivalent of:

```text
llm-wiki page inspect <target> --format json
llm-wiki page plan --draft <path> --target <path> --format json
llm-wiki page apply --plan <plan-id>
```

`inspect` returns resolved profile rules, current metadata, relevant hashes, and safe target information. `plan` validates staged content and returns a deterministic diff, affected-file list, findings, risk class, and approval requirement. `apply` rechecks hashes and refuses stale or modified plans before performing atomic writes.

### FR8. Severity and baseline policy

Validation applies to every file in the configured bundle, including direct human edits.

Findings have three configurable severities:

- **Error:** fails validation and CI.
- **Warning:** reported but does not fail by default.
- **Suggestion:** advisory quality improvement.

The default policy treats invalid YAML, missing profile-required fields, and invalid field types as errors. Broken links and stale indexes default to warnings unless promoted by the profile.

Baseline mode allows adoption in an existing repository without making unrelated edits fail because of pre-existing findings. New or worsened findings still apply.

### FR9. Index maintenance

Indexes are deterministic derived views from page metadata.

Index operations must:

- Avoid duplicate entries.
- Preserve clearly marked human-maintained sections.
- Use title and description where available.
- Produce stable ordering.
- Avoid model calls.
- Support dry-run and diff output.

### FR10. Safe filesystem behavior

All deterministic operations must:

- Restrict writes to the repository and configured bundle or metadata paths.
- Resolve and validate symlinks.
- Use atomic writes.
- Avoid partial updates across multi-file operations.
- Detect dirty-worktree conflicts relevant to changed files.
- Avoid rewriting unrelated formatting.
- Treat imported content as data, never as agent instructions.
- Produce machine-readable and human-readable failure output.
- Restrict staging files to an engine-managed location under `.llm-wiki/`.
- Bind mutation plans to current file hashes and expire or reject stale plans.
- Prevent skills from bypassing approval policy by calling lower-level mutating commands.

### FR11. Review policy

Default behavior is risk-based:

- New page: write, validate, and show diff.
- Existing page: preview diff before applying.
- Rename, delete, migration, or broad edit: explicit approval.
- All deterministic write commands: `--dry-run` support.
- Non-interactive use: requires an explicit review policy flag.
- Existing-page edits: stage, plan, review, then apply by plan ID.

### FR12. Hooks and CI

Local enforcement is optional:

- Claude Code `PreToolUse` hooks may block unsafe paths or protected-file mutations before execution.
- `PostToolUse` hooks may validate completed writes and return feedback, but must not claim to have prevented or reversed them.
- An optional Git pre-commit hook may validate changed wiki files.

Hooks must invoke the same `llm-wiki` engine used by skills. The plugin must ship a ready-to-use CI workflow that invokes that engine. CI is the authoritative team enforcement mechanism and runs full validation by default.

### FR13. Diagnostics and documentation

The CLI must provide a `doctor` command that checks platform support, versions, configuration, profile resolution, filesystem access, hooks, and CI setup.

Documentation must cover installation, initialization, authoring, direct editing, validation, profiles, custom-profile creation, hooks, CI, upgrades, recovery, and uninstall.

## 14. Non-functional requirements

### Usability

- A proficient user can install the plugin and create the first valid page within 15 minutes.
- Normal use requires no runtime installation.
- Errors identify the affected file, rule, severity, and likely fix.

### Reliability

- Repeating an unchanged deterministic operation produces no diff.
- Interrupted operations do not leave partially written files.
- Unknown valid fields survive round trips.
- Validation output is stable for identical inputs and configuration.
- Structured skill-to-engine responses remain backward compatible within a major contract version.
- Applying a stale mutation plan fails without changing repository files.

### Performance

On a documented reference machine with 5,000 concept documents:

- Full deterministic validation completes within five seconds.
- Changed-file validation completes within one second.
- Index regeneration completes without a model call.

### Portability

- Core wiki content remains plain Markdown and YAML.
- Wiki content remains usable after removing the plugin.
- Profiles and configuration are plain text and version-controlled.

### Maintainability

- Root `CLAUDE.md` should remain below 200 lines.
- Skills have one primary responsibility and concise trigger descriptions.
- Profile behavior is data-driven.
- CLI behavior has automated unit and integration tests.
- Deterministic behavior has one implementation shared by skills, hooks, CI, and direct use.
- Skill-specific scripts contain no duplicated policy or mutation logic.

### Security

- Source material is untrusted input.
- Repository content cannot override plugin or user instructions merely by containing imperative text.
- Path traversal and symlink escape attempts are rejected.
- Release checksums are verified before executing bundled artifacts.

## 15. Success metrics

### Activation

- Percentage of users who create a first conformant page.
- Median time from plugin installation to first conformant page.

### Quality

- Percentage of generated pages passing validation without manual structural repair.
- Percentage of authoring diffs accepted without substantive correction.
- Broken-link findings per 100 pages.
- Validator false-positive rate.
- Percentage of sourced model-generated claims with resolvable citations.

### Reliability

- Successful upgrade rate.
- Rate of interrupted operations recovered without corruption.
- Number of confirmed writes outside configured boundaries: target zero.

The product must not optimize for average link count or percentage of edits performed through kit workflows.

## 16. Risks and mitigations

| Risk | Mitigation |
|---|---|
| Profile rules are mistaken for OKF rules | Report OKF and profile conformance separately. |
| Schema becomes rigid too early | Keep core conservative; use recommendations before requirements. |
| Hooks are treated as a security boundary | Position CI as authoritative and document hook limitations. |
| Model overwrites human work | Preview existing-page changes and preserve unknown/unrelated content. |
| Sourced claims lose provenance | Require resolvable citations and preserve existing citations. |
| Custom-profile complexity becomes a programming language | Support declarative data, one inheritance level, and explicit validation. |
| Cross-platform binaries drift | Automate release builds, checksums, and platform integration tests. |
| Compilation drops important research facts | Track evidence gaps and keep sources addressable; add deeper evaluation later. |
| Upgrades overwrite customization | Track ownership and preview repository changes. |
| Skills bypass deterministic safety | Route managed mutations through staged engine plans and narrow tool access. |
| Skill adapters duplicate core logic | Keep all policy and mutations in the shared engine and test adapters as integrations only. |
| A reviewed diff becomes stale before application | Bind plans to hashes and reject stale plans. |

## 17. MVP acceptance criteria

The MVP is ready when:

- The plugin installs and updates through a documented Claude Code plugin distribution flow.
- The correct checksum-verified Go CLI runs on every supported platform without runtime setup.
- Installation succeeds in new and non-empty test repositories without losing existing files.
- Users can initialize with core, academic-research, or a valid custom profile.
- Generated documents report separate OKF and profile conformance.
- Unknown frontmatter fields survive authoring and enrichment.
- Malformed YAML and missing profile-required fields fail validation.
- Broken links are reported at the configured severity.
- Model-generated sourced claims contain resolvable citations in acceptance fixtures.
- Repeating unchanged authoring inputs does not create a duplicate page.
- Existing-page edits are previewed before application.
- Skills perform managed existing-page writes through the engine's staged plan/apply workflow.
- A changed target causes a previously generated plan to be rejected without mutation.
- Structured engine output conforms to its documented contract and exit codes.
- Skills, hooks, CI, and direct CLI validation produce the same findings for the same repository state.
- Renames, deletions, and migrations require explicit approval.
- Pre-write guards reject tested path traversal and symlink escapes.
- Post-write validation is described as feedback, not prevention.
- The CI workflow fails on configured errors.
- Upgrade and uninstall preserve wiki content and local profile extensions.
- Automated tests pass on all supported operating systems and architectures.

## 18. Release plan

### MVP

- Plugin packaging and marketplace-compatible distribution.
- Cross-platform Go CLI.
- Installation, upgrade, doctor, and uninstall.
- OKF and profile validation.
- Core and academic-research profiles.
- Custom-profile template.
- Authoring and enrichment skills.
- Versioned skill-to-engine JSON contract.
- Staged inspect/plan/apply mutation workflow.
- Deterministic indexes.
- Provenance policy.
- Optional hooks and CI workflow.

### Phase 2: Schema evolution

- Recurrence analysis for fields and headings.
- Evidence-backed profile-change proposals.
- Dry-run migrations with rollback guidance.
- Compatibility windows and deprecation policy.
- Additional domain profiles.

Frequency alone must never automatically promote a field or section into a schema rule.

### Phase 3: Ecosystem

- Optional import connectors.
- Additional agent integrations.
- Multiple bundles per repository.
- Static publishing or graph visualization.
- Semantic quality audits and contradiction analysis.

## 19. Remaining product decisions

- Plugin name, command namespace, license, and marketplace location.
- Exact Go YAML library and supported Go version for development.
- Whether GitHub Actions is the only MVP CI template or one of several.
- Exact research-profile templates and conditional-section syntax.
- Profile registry and trust model for third-party profiles.
- Minimum supported Claude Code version.
- Exact packaging mechanism for selecting the correct platform binary inside the plugin.
- JSON contract versioning and compatibility policy.

## 20. Verified references

- [Open Knowledge Format v0.1 specification](https://github.com/GoogleCloudPlatform/knowledge-catalog/blob/main/okf/SPEC.md)
- [Google Cloud introduction to OKF](https://cloud.google.com/blog/products/data-analytics/how-the-open-knowledge-format-can-improve-data-sharing)
- [Claude Code plugins](https://code.claude.com/docs/en/plugins)
- [Claude Code project instructions](https://code.claude.com/docs/en/memory)
- [Claude Code skills](https://code.claude.com/docs/en/skills)
- [Claude Code hooks](https://code.claude.com/docs/en/hooks)
