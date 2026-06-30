# claude-issue-executor — plan-mode and `--no-prompt` details

This file contains the harness-level rhythm rules, significance / trivial
checklists, auto-mode behaviour, alignment-review obligation, and
`--no-prompt` mode bodies referenced from [`SKILL.md`](SKILL.md).
Routing logic and the chat plan-gate (8-step rule) stay in SKILL.md.

## Plan-mode rhythm

The chat plan-gate in SKILL.md is **convention** — the assistant follows the
8-step rule because the skill says so. Claude Code also ships a
**harness-level** mechanism: plan mode (toggled with
`shift+tab shift+tab`) locks the assistant out of all mutating tools
until the user explicitly exits plan mode with approval. The two are
complementary: the chat plan-gate operates *inside* plan mode when the
user has entered it, so both run together rather than competing.

This section defines when the executor requests plan-mode entry and
how it routes sessions of different sizes. The rules below come from
ADR-039
and instance the **category-2** rule of the kit-wide auto-mode
permission contract — see
[`docs/workflow-guide.md` §7](../../docs/workflow-guide.md#7-auto-mode-permission-contract-adr-041)
for the canonical contract.

### Significant checklist

A session is **significant** if it meets any of:

- modifies 3+ files, OR
- edits any `skills/*/SKILL.md`, OR
- edits any `templates/*` file, OR
- edits `bin/*` (scripts, installer), OR
- modifies `.claude/settings*.json` or other harness config, OR
- otherwise carries blast radius beyond a single small fix.

### Trivial checklist

A session is **trivial** if it is one of:

- single typo,
- single-line doc tweak,
- status-line / config-default tweak,
- single-file rename within scratch space,
- `feature-ideas.md` status flip,
- ADR status flip (proposed → accepted),
- single-PR scope with no design decisions and no ADR linkage.

This list is the **single source of truth shared with ADR-038's
`--no-prompt` mode**. When either ADR's checklist evolves, both must
move together — ADR-038's mandatory content-boundary review is the
enforcement point.

### Hybrid path

When the executor parses the invocation and reads the prompt's Scope
section, it classifies the session against the two checklists and
routes to one of three branches:

- **Clearly-significant** → the assistant *requests* the user toggle
  plan mode and waits. The assistant cannot enter plan mode itself —
  only the user can press `shift+tab shift+tab`. After the user
  toggles, the assistant proposes the plan inside plan mode (the chat
  plan-gate's 8-step rule above runs here, with the harness-level
  lock providing belt-and-braces enforcement). The user approves and
  exits plan mode (`shift+tab` once); optionally enables auto-accept
  edits (`shift+tab` again) for the execution phase.
- **Clearly-trivial** → the executor proceeds with the chat plan-gate
  alone. No plan-mode request is made; the friction of toggling is
  not warranted for the work in question.
- **Borderline** → the executor asks once: *"Significant? yes / no /
  decide for me based on scope."* Proceed accordingly.

### Auto-mode behaviour

The executor is **category 2** in the kit-wide auto-mode permission
contract — *operator-acknowledged-bypass*. Auto-mode may proceed
through the plan-mode gate but the bypass must be explicit and
operator-acknowledged, never silent.

When a session starts under auto-mode (operator has pre-authorized
auto-mode for the session via an explicit toggle), the executor
**asks once** at session start, before classifying the session
against the significance checklist:

> *"Enter plan mode for this task? yes / no / decide-from-scope."*

- **`yes`** — request plan-mode entry as normal. The clearly-
  significant branch of the Hybrid path applies regardless of how
  the checklist would have classified the session.
- **`decide-from-scope`** — fall through to the Hybrid path's
  classifier. Clearly-significant requests plan mode; clearly-
  trivial skips it; borderline asks the per-Hybrid-path question.
- **`no`** — the executor writes a one-line acknowledgement in
  chat output, **before any mutating edit**:

  > *"Plan mode bypassed by operator (cat-2 operator-acknowledged
  > bypass per workflow-guide §7)."*

  The acknowledgement is mandatory. Without it the bypass is
  silent and the cat-2 contract is violated. The executor then
  proceeds with the chat plan-gate alone (the 8-step rule above),
  which still requires explicit user approval before any mutating
  edit.

The ask-once rule applies *only* under auto-mode. When the
operator is interacting normally (no auto-mode toggle), the
Hybrid path runs as before — there is no ask-once question to
answer because operator approval at the plan gate is already
explicit.

### Alignment-review obligation

When the trivial checklist above is amended, ADR-038's
`--no-prompt` criteria must be reviewed and aligned in the same change.
ADR-038 carries a mandatory content-boundary review obligation that
serves as the enforcement point for keeping the two checklists in
lockstep.

The cat-2 contract text in
[`docs/workflow-guide.md` §7](../../docs/workflow-guide.md#7-auto-mode-permission-contract-adr-041)
and the significance checklist above must also move together: when
either is amended, review the other in the same change. PR review
is the enforcement point until ADR-034's plan-checker grows a
structural rule for it (deferred to issue #72).

## `--no-prompt` mode

Per ADR-038, the
executor accepts a `--no-prompt` flag that **skips prompt generation
entirely** and runs from the issue body alone. The prompt artefact is
not written. A one-line breadcrumb — `issue executed without prompt
per ADR-038` — is appended to the first commit's message so the
audit trail survives.

### When `--no-prompt` is appropriate

The criteria for `--no-prompt` are exactly the **Trivial checklist**
above (lines beginning *"A session is **trivial** if it is one of:"*).
That checklist is the **single source of truth**; the `--no-prompt`
mode does not duplicate it. If the trivial checklist evolves, this
mode's criteria evolve in lockstep — see the **Alignment-review
obligation** section just above.

### Auto-detect (with confirmation)

When the user invokes `/claude-issue-executor <issue-number>` (no
explicit `--no-prompt`), the executor auto-detects candidates
*conservatively*:

- The issue has **zero `ADR-NNN` references** in its body, AND
- The issue has at least one of the labels `chore`, `docs`, or
  `bugfix-trivial`.

When both conditions hold, the executor *suggests* `--no-prompt`
mode and asks for confirmation: *"Issue looks trivial — skip prompt
generation? (yes / no)"*. On `yes`, proceed without prompt. On `no`,
fall back to the standard auto-chain path. The suggestion never
short-circuits silently.

### Override

Explicit `--no-prompt` overrides auto-detection — no confirmation is
asked. This is for the user who already knows the issue is trivial
and wants the lowest-ceremony path. The breadcrumb is still left in
the commit message.

### Interactions

- **Plan-mode rhythm** still applies. `--no-prompt`
  affects the *prompt* step, not the plan-mode gate. A trivial
  issue with `--no-prompt` typically also classifies as
  clearly-trivial against the significance checklist, so plan mode
  is not requested. But if the executor is ever invoked with
  `--no-prompt` on something that *is* significant (e.g. a multi-
  file refactor mislabelled `chore`), the significance gate still
  fires — the two are independent.
- **`/check-plan`** does not run when `--no-prompt` is
  set, because there is no rendered prompt to check. The skip is
  noted in the evaluation summary alongside the breadcrumb.
- **`design/state.md`** updates still happen. The
  `state:in-flight` zone records `Status: verified` (or `executing`
  mid-session) regardless of whether a prompt was generated.
