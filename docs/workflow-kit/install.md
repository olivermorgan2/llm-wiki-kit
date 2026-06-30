# Install Guide

How to install the Claude Code Workflow Kit into a **new** project. The
kit is project-agnostic (per ADR-028)
— software, research, content, curriculum, or any other structured
project that fits the kit's three assumptions: a git repo, GitHub, and
Claude Code.

> **Before you start:** the kit is for **new projects only**. It does
> not adapt or migrate existing repositories. If you want to add the
> workflow to an existing codebase, this kit is not the right fit —
> see [What this kit does not support](#what-this-kit-does-not-support) below and
> ADR-002.

The kit uses a **project-local installation model**: each new target
project gets its own copy of the skills under `.claude/skills/`. There is
no global install. See ADR-001
for the reasoning.

---

## 1. Prerequisites

Install these once per machine. The kit assumes macOS or Linux; commands
are cross-platform where possible.

| Tool | Purpose | Install |
|---|---|---|
| **Git** | Version control | Ships with macOS (Xcode CLT) or `brew install git` |
| **GitHub CLI (`gh`)** | Repo, issue, and PR operations from the terminal | `brew install gh` then `gh auth login` |
| **`jq`** | JSON parsing in shipped helper scripts such as `bin/changelog-collect`, `bin/release-suggest`, and `bin/pr-context` | `brew install jq` |
| **Claude Code** | AI-assisted development | `npm install -g @anthropic-ai/claude-code` |
| **SSH key** | Passwordless git over SSH | `ssh-keygen -t ed25519` then add the public key to GitHub |

### Verify the setup

```bash
git --version          # any recent 2.x is fine
gh --version           # any recent 2.x is fine
gh auth status         # should report "Logged in to github.com"
jq --version           # any 1.6+ is fine
claude --version       # confirms Claude Code is on your PATH
ssh -T git@github.com  # "Hi <user>! You've successfully authenticated"
```

If any of these fail, fix the failing tool before continuing — the rest of
the guide assumes they all pass.

---

## 2. Choose your target project

Pick **one** of these two paths. The rest of the guide assumes you end up
inside the target project directory.

**A. Create the repo on GitHub first** (recommended):

```bash
gh repo create my-project --public --clone   # use --private if appropriate
cd my-project
```

**B. Start locally, add the remote later:**

```bash
mkdir my-project && cd my-project
git init
```

---

## 3. Install the kit into the target project

You have three options, in order of preference:

- **Bootstrap install** (recommended for users) — one command pulls the
  kit at a pinned version from GitHub, runs the installer, and discards
  the temporary kit copy. See [3A](#3a-bootstrap-install-recommended)
  below.
- **Explicit fetch** — three explicit lines that do the same thing as
  the bootstrap script, more transparently. See
  [3B](#3b-explicit-fetch-alternative) below.
- **Manual install** — full do-it-yourself copy flow. See
  [3C](#3c-manual-install) below.

All three produce the same target-project layout described in
[`docs/repo-structure.md`](repo-structure.md). Per ADR-029, the kit is
designed for **per-project remote install** — there is no long-lived
local kit clone to maintain.

### 3A. Bootstrap install (recommended)

The bootstrap script (per ADR-029) ships as an asset on every tagged
release. It fetches the kit at the pinned version, runs the installer
against the current directory (or `--target=PATH`), and cleans up the
temporary kit copy on exit.

```bash
cd my-project                                              # target project root
bash <(curl -fsSL https://github.com/olivermorgan2/claude-workflow-kit/releases/download/v5.0.1/bootstrap-workflow-kit) \
  --project-name=my-project
```

Prefer to inspect the script before running, or pipe-to-bash makes
you nervous? Download it first:

```bash
gh release download v5.0.1 -p bootstrap-workflow-kit \
  -R olivermorgan2/claude-workflow-kit
chmod +x bootstrap-workflow-kit
./bootstrap-workflow-kit --project-name=my-project
```

What the bootstrap does, in order:

1. Resolves the target version (env var `WORKFLOW_KIT_VERSION`, then
   `gh release view`, then `git ls-remote --tags`).
2. Creates a temporary directory, traps cleanup on exit.
3. Fetches the kit (`gh repo clone --depth=1 --branch=vX.Y.Z` if `gh`
   is available; falls back to `git clone --depth=1` over HTTPS).
4. Verifies the fetched tree contains `bin/install-workflow-kit`.
5. Forwards all CLI args to that installer.

Environment variables:

| Variable | Default | Purpose |
|---|---|---|
| `WORKFLOW_KIT_VERSION` | latest tag | Pin to a specific tag, e.g. `v5.0.1`. |
| `WORKFLOW_KIT_REPO` | `olivermorgan2/claude-workflow-kit` | Override for forks. |

The installer's behaviour itself is unchanged from ADR-009:

1. Creates `design/adr/`, `prompts/`, `notes/`, and `.claude/skills/`
   in the target project.
2. Copies every non-optional skill from the kit's `skills/` into `.claude/skills/`.
   The `/review-pr` skill is optional and installs only with `--with-ai-review`.
3. Copies `bin/sync-adr-index` into `.claude/bin/` (ADR-023).
4. Seeds `prompts/_template.md` so new per-issue prompts stay
   consistent (ADR-008).
5. Renders `CLAUDE.md` from `templates/claude-md-template.md`
   (ADR-007). You are prompted only for the four day-one required
   fields — `PROJECT_NAME`, `GITHUB_OWNER`, `GITHUB_REPO`, and
   `DEFAULT_BRANCH` (owner/repo are pre-filled from the `origin`
   remote when present). Every other placeholder is filled with the
   `_TBD_` marker, so the rendered file never carries unresolved
   `{{...}}` syntax. Supply any optional value up front with
   `--set KEY=VALUE`, or fill it in later as the project settles.
6. `git init`s the target if it is not already a git repo, then
   creates an initial commit `chore: install workflow kit
   (project-local)`.

Useful installer flags (forwarded by the bootstrap):

| Flag | What it does |
|---|---|
| `--target=PATH` | Target project directory (default: current dir) |
| `--project-name=NAME` | Value for `{{PROJECT_NAME}}` in `CLAUDE.md` |
| `--set KEY=VALUE` | Provide a value for any other `{{UPPER_SNAKE}}` placeholder. Repeatable. |
| `--with-docs` | Also copy kit docs to `docs/workflow-kit/` in the target (ADR-010) |
| `--with-ai-review` | Install the optional AI PR review runtime and `/review-pr` skill under `.claude/` for this target. Default: off. |
| `--force` | Overwrite existing `CLAUDE.md` and re-copy skills/runtime files; preserves user-edited `.claude/ai-review/config.json` |
| `--no-commit` | Skip the initial commit |
| `--non-interactive` | Never prompt; fall back to defaults |
| `--license=ID` | Scaffold a starter `LICENSE` in the target. Supported: `mit`. Default: no `LICENSE` is written (license choice is the project author's call, per ADR-025). |
| `--license-holder=NAME` | Copyright holder for the rendered `LICENSE` (only used when `--license` is set). Falls back to `--project-name`, then to the target's basename. |
| `-h`, `--help` | Show full usage |

> **License scaffolding example.** Add an MIT LICENSE attributed to
> "Jane Doe" alongside the rest of the install:
>
> ```bash
> bash <(curl -fsSL https://github.com/olivermorgan2/claude-workflow-kit/releases/download/v5.0.1/bootstrap-workflow-kit) \
>   --project-name=my-project \
>   --license=mit \
>   --license-holder="Jane Doe"
> ```

The installer is idempotent: re-running on an already-installed target
skips files that already exist and makes no commit if nothing changed.
Pass `--force` to re-render `CLAUDE.md` or refresh installed kit files.
If AI review is installed, `--force` refreshes runtime scripts, prompt
packs, schemas, and `config.example.json`, but preserves any user-edited
`.claude/ai-review/config.json`.

### Optional AI PR review

AI PR review is an explicit per-target choice. To install it:

```bash
bash <(curl -fsSL https://github.com/olivermorgan2/claude-workflow-kit/releases/download/v5.0.1/bootstrap-workflow-kit) \
  --project-name=my-project \
  --with-ai-review
```

This installs the `/review-pr` skill plus the target-local runtime under
`.claude/bin/` and `.claude/ai-review/`. Provider config stays secret-free:

```bash
cp .claude/ai-review/config.example.json .claude/ai-review/config.json
export OPENROUTER_API_KEY=sk-or-...   # shell or secret manager only; never commit it
.claude/bin/review-pr --pr 123 --profile balanced --format md
```

Persistent secret options:

- shell profile export in `~/.zshrc` or `~/.bashrc`
- `direnv` / `.envrc` if `.envrc` is gitignored
- CI secret store for automation

Never put API keys in `.claude/ai-review/config.json`; `apiKeyEnv` names the
environment variable only. OpenRouter is just the documented default — see
[`docs/ai-review.md`](ai-review.md) for the review/publish flow and how to
point the review at OpenAI directly or any other OpenAI-compatible provider.

### 3B. Explicit fetch alternative

The bootstrap script is convenient, but if you want to see exactly
what is happening, this three-line form does the same thing — fetch
the kit at a pinned version into a temp dir, run the installer,
clean up:

```bash
TMPKIT="$(mktemp -d)" && \
  gh repo clone olivermorgan2/claude-workflow-kit "$TMPKIT" -- \
    --depth=1 --branch=v5.0.1 && \
  "$TMPKIT/bin/install-workflow-kit" --project-name=my-project && \
  rm -rf "$TMPKIT"
```

Replace the tag with whichever release you want to pin. Replace the
installer flags with whatever your project needs (see the table in
[3A](#3a-bootstrap-install-recommended)).

If `gh` is not installed, swap the clone line for plain git over
HTTPS:

```bash
git clone --depth=1 --branch=v5.0.1 \
  https://github.com/olivermorgan2/claude-workflow-kit.git "$TMPKIT"
```

### 3C. Manual install

The manual flow is the explicit, do-it-yourself version of what the
installer does. Use it when you want full visibility, or as a
diagnostic reference if the installer misbehaves.

#### 3C.1 Fetch the kit at a pinned version

```bash
TMPKIT="$(mktemp -d)"
gh repo clone olivermorgan2/claude-workflow-kit "$TMPKIT" -- --depth=1 --branch=v5.0.1
# or:  git clone --depth=1 --branch=v5.0.1 https://github.com/olivermorgan2/claude-workflow-kit.git "$TMPKIT"
```

#### 3C.2 From inside the target project, copy the skills

```bash
cd my-project                                   # target project root
mkdir -p .claude/skills .claude/bin design/adr prompts notes
cp -R "$TMPKIT/skills/"* .claude/skills/
cp "$TMPKIT/bin/sync-adr-index" .claude/bin/
chmod +x .claude/bin/sync-adr-index
cp "$TMPKIT/prompts/_template.md" prompts/_template.md
```

The `prompts/` folder holds one filled session brief per GitHub
issue (`issue-NNN-short-title.md`), copied from `_template.md`.
Freeform working notes stay in `notes/`. See
ADR-008.

This is the step that makes the install **project-local**: the
skills now live under the target project's own `.claude/skills/`,
tracked in the target project's git history. The temporary kit
clone is no longer required for those skills to work.

#### 3C.3 Render the starter `CLAUDE.md`

`CLAUDE.md` at the project root is Claude Code's primary rules file
for the project. Copy the template and fill in the project-specific
fields by hand:

```bash
cp "$TMPKIT/templates/claude-md-template.md" CLAUDE.md
```

The template uses `{{PLACEHOLDER}}` tokens (ADR-007). Open
`CLAUDE.md` and replace each one with real values, starting with
`{{PROJECT_NAME}}`. The automated installer above does this
substitution for you.

#### 3C.4 Commit the install and clean up

```bash
git add .claude CLAUDE.md design prompts notes
git commit -m "chore: install workflow kit (project-local)"
rm -rf "$TMPKIT"
```

---

## 4. Verify the install

After step 3 the target project should look like this (trimmed):

```
my-project/
  CLAUDE.md
  .claude/
    skills/
      idea-to-prd/          # plus the other skills shipped by the kit
  design/
    adr/
  prompts/
    _template.md            # blank template; copy to start an issue session
  notes/
```

A quick check:

```bash
ls .claude/skills          # should list the skill directories
ls prompts                 # should list _template.md
cat CLAUDE.md | head -20   # should show your project rules, not this kit's
```

For the full expected layout and which artifacts originate where, see
[`docs/repo-structure.md`](repo-structure.md).

---

## 5. Pick a starting path

The kit supports three ways to start planning, defined in
ADR-003. Pick the one that
matches what you have in hand:

| You have… | Use this skill | Status |
|---|---|---|
| A rough idea, no PRD | `idea-to-prd` | Issue #5 |
| A standard PRD | `prd-normalizer` → `prd-to-mvp` | Issue #6, #7 |
| Mixed notes / custom PRD format | `prd-normalizer` → `prd-to-mvp` | Issue #6, #7 |

Run the chosen skill inside the target project using Claude Code. For
how to invoke skills, use plan mode, and follow the approve-then-implement
loop, see [`claude-code-guide.md`](claude-code-guide.md). Each skill's
own `SKILL.md` remains the authoritative spec for its inputs and outputs.

## What this kit does not support

The kit intentionally does **not** include:

- **Retrofitting existing repositories.** The kit assumes a clean slate.
  Pointing it at an established codebase will not produce useful output.
- **A global install.** Every target project gets its own copy of the
  skills. This is a design choice, not a limitation (see ADR-001).
- **Team/multi-repo features.** The kit is scoped to a solo author or
  small team working on a single project at a time.
- **Non-GitHub providers.** The workflow is GitHub-first
  (ADR-004).
  GitLab, Bitbucket, and others are not supported.
- **Non-Claude AI tools.** The skills target Claude Code specifically
  (ADR-006).

If any of the above is a hard requirement for your project, this release
is not the right fit — revisit after Phase 2.

---

## Troubleshooting

**`gh: command not found`**
Install the GitHub CLI (`brew install gh` on macOS) and run `gh auth login`.

**`gh auth status` reports you are logged out**
Run `gh auth login`. Choose SSH as the git protocol to match the clone
command in step 3B.1.

**`cp: .../skills/*: No such file or directory`**
The kit was cloned to a different path than the guide assumes. Either
re-clone the kit or substitute your actual clone path in
step 3B.2 (or pass the correct path to `bin/install-workflow-kit`).

**`install-workflow-kit: error: kit skills/ not found`**
The installer resolves its sources relative to its own location. Run it
directly from the kit clone (e.g. `/path/to/kit-clone/bin/install-workflow-kit`),
not via a copy detached from the kit repo.

**Claude Code does not see the skills**
Confirm you are running Claude Code from the **target project's root**.
Project-local skills are discovered relative to the current working
directory. If you `cd` into a subdirectory, skills still resolve, but if
you run Claude Code from an unrelated directory they will not.

**I already have a `design/` or `.claude/` in the target project**
The copy commands in step 3 use `mkdir -p` and directory-level `cp -R`,
so they will not overwrite files that already exist with different
names. If you want to replace them wholesale, remove them first.

---

## What's next

- [`docs/workflow-guide.md`](workflow-guide.md) — end-to-end flow from idea to shipped release
- [`docs/repo-structure.md`](repo-structure.md) — how the kit is laid out and what it generates
- [`docs/github-setup.md`](github-setup.md) — GitHub labels, milestones, and branch conventions
