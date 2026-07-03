# ADR-009: Install / upgrade / uninstall asset-ownership model

**Status:** accepted
**Date:** 2026-07-03

## Context

The plugin installs into a user's repository, is later upgraded, and can be
uninstalled — and across all three it must **never lose or silently overwrite a
user's content or local extensions** (PRD §8 FR; §14 Reliability). Acceptance
criterion **1** covers the documented install/distribution flow, and criteria
**3** and **20** require install into a **non-empty** repo and later
upgrade/uninstall to **preserve** repository content and local customization.
[ADR-002](adr-002-platform-binary-selection.md) (accepted) explicitly **carved
this out** of its scope: it owns *ship/select/verify* the binary and named the
broader "install / upgrade / uninstall **asset ownership** (plugin-owned vs
repo-owned, content preservation — criteria 1, 3, 20)" as **ADR-009's domain**.
Open question **Q7** is closed only for ship/select/verify and points its
install-ownership remainder here.

Phase 2 needs the **install half**: install into new and non-empty repos with
`--dry-run` and silent-overwrite refusal, recording plugin/CLI/OKF/profile
versions. Upgrade and uninstall are **decided here** but **implemented in
Phase 7**. This ADR does not re-open how mutations are made safe — that is
[ADR-005](adr-005-safe-filesystem-layer.md)'s write-gate and
[ADR-006](adr-006-staged-mutation-transaction-model.md)'s cross-file
transaction — it decides **which assets the plugin owns** so upgrade/uninstall
know what they may touch. It also depends on
[ADR-007](adr-007-profile-system-boundary.md) for what a materialized profile
looks like on disk.

Constraints: the MVP excludes any profile **registry / third-party trust**
(addendum 005, Q5), so ownership is a two-class problem, not a marketplace one;
and, mirroring ADR-002, release **signing / provenance attestation** — the
trust root against a maliciously rebuilt payload — is **out of scope** here.
Risks addressed (`knowledge/risks.md`): "upgrades overwrite customization" and
"model overwrites human work" (the ownership record is the mechanism that lets
upgrade preserve repo-owned content and local profile extensions).

## Options considered

### Option A: Explicit ownership + version-record manifest — every installed asset is classified plugin-owned or repo-owned in a recorded manifest, and install/upgrade/uninstall act only within their class

- Pros: install writes **plugin-owned** assets through the ADR-006 transaction
  and records an **ownership + version manifest** (the plugin, CLI, OKF, and
  profile versions plus the path and hash of each plugin-owned asset). Because
  ownership is **explicit and recorded**, upgrade can replace plugin-owned
  assets while leaving every **repo-owned** file (user content, local profile
  extensions) untouched (criteria 3, 20), and uninstall can remove exactly the
  plugin-owned set and nothing else. Install **refuses to silently overwrite**
  any pre-existing file it did not itself record as plugin-owned, so a
  non-empty repo loses no file (criterion 3); `--dry-run` renders the full
  planned write set from the same ADR-006 plan with **zero mutation**
  (criterion 1). The manifest is declarative data, consistent with the ADR-007
  profile model and the offline/self-contained promise.
- Cons: the manifest is state the engine must write, version, and keep truthful
  across upgrades (a drifted manifest mis-reports ownership); classifying each
  asset (including ADR-007 materialized-profile references and bundle config)
  requires a deliberate ownership decision per asset type; and upgrade/uninstall
  must reconcile the recorded manifest against the live tree (detecting
  user-modified plugin-owned files) rather than blindly overwriting.

### Option B: Convention/path-prefix ownership — everything under a reserved directory is plugin-owned, everything else is the user's

- Pros: no manifest to maintain; ownership is "read the path."
- Cons: it cannot express plugin-owned assets that must live **outside** a
  reserved directory (e.g. a config or profile reference the user is expected to
  see at the repo root), nor record the **versions** criterion 1 requires; it
  gives upgrade no way to detect that a user edited a plugin-owned file, so
  preservation (criteria 3, 20) degrades to "hope the user stayed out of the
  reserved dir"; and a path convention silently reclassifies a file if the user
  moves it. Ownership-by-convention is a subset of what the manifest records,
  without the version record or modification detection.

### Option C: No ownership record — idempotent overwrite / diff-against-shipped

- Pros: simplest install — write the shipped set every time.
- Cons: with nothing recording what the plugin owns, upgrade cannot
  distinguish a user's edit from a stale plugin file, so it must either
  overwrite (losing customization — the exact criteria-3/20 failure) or never
  update (defeating upgrade); uninstall cannot know what to remove; and there is
  no place to record plugin/CLI/OKF/profile versions (criterion 1). It converts
  a preservation guarantee into a guess.

## Decision

Adopt **Option A** — an **explicit ownership + version-record manifest**. Every
**install-managed** asset is classified **plugin-owned** or **repo-owned** and
recorded, alongside the **plugin, CLI, OKF, and profile versions**, in a
version-record file written at install time. The manifest is a record of the
assets install *manages*, not a full inventory of the repository: **plugin-owned**
entries are the files the plugin writes and may later replace or remove;
**repo-owned** entries are recorded only where install must *reference* existing
repo content (bundle config, a local profile-extension reference) so upgrade
knows to preserve rather than touch it — arbitrary pre-existing repo files are
neither written nor catalogued. Install writes only plugin-owned assets, through
the [ADR-006](adr-006-staged-mutation-transaction-model.md) transaction so a
non-empty-repo install is all-or-nothing and loses no file; it **refuses to
silently overwrite** any pre-existing file not recorded as plugin-owned; and
`--dry-run` lists the complete planned write set with **zero mutation**
(criterion 1). **Upgrade** (Phase 7) replaces only plugin-owned assets and
preserves every repo-owned file and local profile extension, reconciling the
manifest against the live tree to detect user-modified plugin-owned files rather
than clobbering them; **uninstall** (Phase 7) removes exactly the recorded
plugin-owned set (criteria 3, 20). Option B is rejected because a path
convention cannot record versions or detect user edits; Option C is rejected
because without an ownership record preservation becomes a guess.

**Manifest specification (decision-level constraints).** So the manifest is
placeable, protectable, and unambiguous for the Phase 2 implementer:

- **Path & format.** A single JSON document at `.llm-wiki/manifest.json`, inside
  the same engine-managed `.llm-wiki/` tree ADR-005 reserves and ADR-006 stages
  under — never at the repo root and never inside user content.
- **Ownership of the manifest itself.** The manifest is **plugin-owned** engine
  metadata: it is written and rewritten only by the engine through the ADR-006
  transaction, is removed by uninstall, and is itself subject to the
  silent-overwrite refusal (a pre-existing non-manifest file at that path is a
  conflict, not a target).
- **Minimum fields.** A `schemaVersion`; the `plugin`, `cli`, `okf`, and
  `profile` versions (profile as id + version); and, per managed asset, a record
  of `{ path, class (plugin-owned | repo-owned), hash (current on-disk content
  hash at last write), lastInstalledHash (the hash the plugin last wrote) }`.
  `hash` and `lastInstalledHash` diverging is the user-modification signal.
- **First install (no manifest).** Absence of `.llm-wiki/manifest.json` means
  *not installed*: install proceeds against an empty ownership set, writes the
  manifest as one asset of the same ADR-006 transaction, and treats every
  pre-existing file as repo-owned and untouchable (silent-overwrite refusal).
- **Missing / unreadable / hash-inconsistent.** Recovery **fails closed** rather
  than guessing ownership. For **upgrade/uninstall**, a *missing* manifest is an
  error (nothing proves what the plugin owns — refuse and tell the user to
  re-install); an *unreadable/corrupt* manifest is a hard error with **zero
  mutation**; a plugin-owned asset whose live `hash` no longer matches its
  `lastInstalledHash` is **hash-inconsistent** (user-modified) and handled per
  the next point.
- **User-modified plugin-owned asset.** Never silently clobbered. During
  upgrade or uninstall a user-modified plugin-owned asset is a **conflict**:
  default behaviour is **skip and report** it (preserving the user's edit); it
  is overwritten or removed only under an explicit `--force`. This keeps
  criteria 3/20 preservation true even inside the plugin-owned class.

**Explicitly out of scope (deferred).** Release **signing / provenance
attestation** — the trust root against a maliciously rebuilt payload — is
**not** decided here; mirroring ADR-002's deferral, it belongs to a **dedicated
supply-chain / signing ADR**, and the residual "supply chain / tampered binary"
risk stays `open` in `knowledge/risks.md`. Third-party profile **registry /
trust** likewise stays deferred to Phase 3 (Q5, addendum 005): ownership here is
a two-class (plugin vs repo) model, not a marketplace one. Upgrade/uninstall are
**decided** here and **implemented** in Phase 7.

## Consequences

- Easier: install/upgrade/uninstall each act within a **recorded** ownership
  class, so preserving user content and local profile extensions (criteria 3,
  20) is a lookup, not a heuristic; the version record gives criterion 1 a
  concrete artifact; `--dry-run` reuses the ADR-006 plan for a zero-mutation
  preview; and building on ADR-006 means a non-empty-repo install is
  all-or-nothing.
- Harder: the engine must write, version, and keep the ownership/version
  manifest truthful across upgrades, classify every asset type (including
  ADR-007 profile references and bundle config), and reconcile the manifest
  against the live tree to detect user-modified plugin-owned files before
  upgrade/uninstall touch them.
- Maintain: the `.llm-wiki/manifest.json` schema (`schemaVersion`,
  plugin/CLI/OKF/profile versions, and per-asset `path`/`class`/`hash`/
  `lastInstalledHash`), the plugin-owned status of the manifest itself, the
  install/upgrade/uninstall preservation rules, the fail-closed
  missing/corrupt-manifest policy and the skip-and-report (`--force`-to-override)
  conflict rule for user-modified plugin-owned assets, the silent-overwrite
  refusal, and the `--dry-run` plan — all riding on the ADR-005 gate and ADR-006
  transaction (no new write path).
- Deferred / validation implications: criterion 1 (install half) and criterion
  3 (non-empty-repo no-file-loss) are proved in **Phase 2**; upgrade/uninstall
  preservation (criteria 3, 20 at lifecycle scope) is proved in **Phase 7**.
  Release **signing/provenance** is deferred to a dedicated supply-chain ADR and
  its risk stays `open`; profile **registry/trust** stays Phase 3. Q7's
  install-ownership remainder is decided here (record in `knowledge/log.md` and
  annotate `knowledge/open-questions.md`); the signing remainder is **not**.
