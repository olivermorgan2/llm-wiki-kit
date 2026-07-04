package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/olivermorgan2/llm-wiki-kit/internal/contract"
	"github.com/olivermorgan2/llm-wiki-kit/internal/fsafe"
	"github.com/olivermorgan2/llm-wiki-kit/internal/plan"
	"github.com/olivermorgan2/llm-wiki-kit/internal/platform"
	"github.com/olivermorgan2/llm-wiki-kit/internal/profile"
	"github.com/olivermorgan2/llm-wiki-kit/internal/scaffold"
	"github.com/olivermorgan2/llm-wiki-kit/internal/txn"
	"github.com/olivermorgan2/llm-wiki-kit/internal/validate"
	"github.com/olivermorgan2/llm-wiki-kit/internal/yamladapter"
)

// version is the llm-wiki build version. It is a skeleton placeholder until the
// release pipeline stamps it.
const version = "0.1.0-dev"

const usage = `llm-wiki — deterministic engine for portable knowledge bundles

Usage:
  llm-wiki <command> [--json] [args]

Commands:
  version     Print the engine version.
  init        Scaffold a new wiki bundle in a target directory (default: the
              current directory, which must already exist). Writes a bundle
              config plus a minimal, immediately-valid core-profile wiki.
              Use --profile <id> to select a profile (default: core) and
              --force to overwrite an existing bundle. Without --force, init
              refuses when any target file already exists (exit 3).
  install     Install the kit's bundle into a new or non-empty repository
              (default target: the current directory, which must already
              exist), writing the scaffold plus a version-record manifest at
              .llm-wiki/manifest.json through one transaction (ADR-009). Use
              --profile <id> to select a profile (default: core), --dry-run to
              print the full planned write set without touching the tree, and
              --force to overwrite. Without --force, install refuses when any
              target already exists (exit 3), losing no user file.
  validate    Validate a wiki against the OKF base and core profile. Reports
              OKF and profile findings separately at three severities
              (error/warning/suggestion). Takes an optional target directory
              (default: the current directory). Set LLM_WIKI_SEVERITY to a
              comma-separated list of code=severity pairs (e.g.
              core-broken-link=error) to promote/demote a rule's severity. Set
              LLM_WIKI_EVIDENCE_SECTIONS to a comma-separated list of heading
              titles (e.g. Evidence) to designate evidence contexts; links inside
              them are classified as citations (core-citation-* findings) instead
              of navigational links.
  page inspect
              Report a single page without mutating it. Reads the page at
              <path> (relative to --root, default the current directory, or an
              absolute path contained by it) and reports its parse status,
              validation findings, and content hash (lowercase-hex SHA-256).
              Findings are identical to those validate reports for the same
              page; LLM_WIKI_SEVERITY overrides apply as with validate. The
              path must resolve to a regular .md file inside the bundle
              boundary; a path escaping the boundary is refused (exit 5). Pass
              --content <file> (or "-" for stdin) to inspect proposed content
              instead of the live page — the draft is validated in full bundle
              context without being staged, and its target may not yet exist (a
              new-page draft), which is the authoring flow's validate step.
  page plan   Preview a whole-page change without mutating the live tree
              (ADR-006). Reads the proposed page content from --content <file>
              (or stdin when the flag is omitted or set to "-"), normalizes its
              frontmatter so unknown fields survive, binds the source/target
              base hash (an absent sentinel for a new page), and prints a unified
              diff. The change set is staged under .llm-wiki/staging/<txn-id>/
              for a later apply; nothing is committed. When the proposed content
              already matches the live page the plan is a no-op and stages
              nothing. Takes <path> and --root like page inspect. When
              LLM_WIKI_EVIDENCE_SECTIONS designates evidence contexts and the
              edit drops an existing citation from them, the plan still succeeds
              (exit 0) but emits a core-citation-loss finding and records an
              approval requirement, so page apply refuses the removal (exit 3)
              until re-run with --approve (ADR-008).
  page apply  Commit a staged page plan into the live tree (ADR-006). Takes the
              <txn-id> a prior page plan reported and --root (default the current
              directory). Before mutating, apply re-verifies the plan's recorded
              base hashes against the live tree; if a source/target changed since
              the plan was made it is rejected with zero mutation (exit 4). A
              clean apply commits exactly the previewed files through the
              transaction layer and cleans staging. If the plan carries an
              approval requirement, apply refuses (exit 3) until re-run with
              --approve.
  selfcheck   Verify this platform's shipped binary against its bundled
              checksum (ADR-002). Use --root <dir> to point at the bundle;
              defaults to the directory containing the running executable.

Flags:
  --json      Emit the versioned JSON contract envelope instead of human text.
              Equivalent to setting LLM_WIKI_JSON=1.
`

// run is the testable CLI entry point. It parses args, dispatches to a command,
// and returns the process exit code. Output is written to stdout/stderr rather
// than the real streams so it can be captured in tests.
func run(args []string, stdout, stderr io.Writer) int {
	jsonMode := jsonEnabled()
	var positional []string
	for _, a := range args {
		switch a {
		case "--json", "-json":
			jsonMode = true
		case "-h", "--help", "help":
			fmt.Fprint(stdout, usage)
			return int(contract.ExitSuccess)
		default:
			positional = append(positional, a)
		}
	}

	if len(positional) == 0 {
		if jsonMode {
			env := contract.New("", contract.StatusInvalidInvocation)
			return emit(stdout, env)
		}
		fmt.Fprint(stderr, usage)
		return int(contract.ExitInvalidInvocation)
	}

	command, cmdArgs := positional[0], positional[1:]
	switch command {
	case "version":
		return runVersion(stdout, jsonMode)
	case "init":
		return runInit(cmdArgs, stdout, stderr, jsonMode)
	case "install":
		return runInstall(cmdArgs, stdout, stderr, jsonMode)
	case "validate":
		return runValidate(cmdArgs, stdout, jsonMode)
	case "page":
		return runPage(cmdArgs, stdout, stderr, jsonMode)
	case "selfcheck":
		return runSelfcheck(cmdArgs, stdout, stderr, jsonMode)
	default:
		return runUnknown(command, stdout, stderr, jsonMode)
	}
}

// jsonEnabled reports whether the LLM_WIKI_JSON env toggle requests JSON output.
func jsonEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("LLM_WIKI_JSON"))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

// severityOverrides parses the LLM_WIKI_SEVERITY configuration into the profile-
// override map consumed by validate.Resolve (ADR-004's profile-override layer).
// The value is a comma-separated list of `code=severity` pairs, e.g.
// `core-broken-link=error`; severity must be one of error/warning/suggestion.
// Blank entries and pairs with an unrecognized severity are ignored so a stray
// config value cannot crash validation. An empty or fully-invalid config yields
// nil — the identity that a core-only run passes to Resolve. Overrides are keyed
// by rule code; Resolve ignores any code no finding carries.
func severityOverrides(config string) map[string]contract.Severity {
	overrides := map[string]contract.Severity{}
	for _, pair := range strings.Split(config, ",") {
		code, sev, ok := strings.Cut(pair, "=")
		if !ok {
			continue
		}
		code = strings.TrimSpace(code)
		if code == "" {
			continue
		}
		switch contract.Severity(strings.TrimSpace(sev)) {
		case contract.SeverityError:
			overrides[code] = contract.SeverityError
		case contract.SeverityWarning:
			overrides[code] = contract.SeverityWarning
		case contract.SeveritySuggestion:
			overrides[code] = contract.SeveritySuggestion
		}
	}
	if len(overrides) == 0 {
		return nil
	}
	return overrides
}

// evidenceSections parses the LLM_WIKI_EVIDENCE_SECTIONS configuration into the
// list of ATX heading titles that open a profile-designated evidence context
// (ADR-008 sub-decision 1). The value is a comma-separated list of titles, each
// trimmed of surrounding whitespace with empty entries dropped, e.g.
// `LLM_WIKI_EVIDENCE_SECTIONS=Evidence,Sources`. An unset or empty variable
// yields nil — the zero-cost default where no link is ever a citation, so the
// engine behaves byte-identically to a run with citations disabled.
//
// This is the interim wiring channel for evidence-context designation, mirroring
// the LLM_WIKI_SEVERITY precedent; profile-loaded evidence vocabulary is Phase 4
// work that will supersede or augment it. It is engine configuration, not
// profile-specific citation vocabulary.
func evidenceSections() []string {
	var out []string
	for _, s := range strings.Split(os.Getenv("LLM_WIKI_EVIDENCE_SECTIONS"), ",") {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func runVersion(stdout io.Writer, jsonMode bool) int {
	env := contract.New("version", contract.StatusSuccess)
	if jsonMode {
		return emit(stdout, env)
	}
	fmt.Fprintf(stdout, "llm-wiki %s (contract %s)\n", version, contract.ContractVersion)
	return int(contract.ExitSuccess)
}

// runInit scaffolds a new wiki bundle in a target directory. It parses its own
// flags in the hand-rolled style of selfcheckRoot (the global loop only strips
// --json), resolves the requested profile, plans the change set, and writes it
// through one ADR-006 transaction so the bundle is created all-or-nothing.
//
// Refusal semantics: when any planned target already exists and --force was not
// given, init mutates nothing and reports StatusApprovalRequired (exit 3), not
// invalid-invocation. The invocation is well-formed — the operation simply
// awaits explicit permission to overwrite — and contract.Approval is shaped for
// exactly this. install (ADR-009, runInstall) reuses the same envelope through
// the shared cmdRefuse helper, so exit 3 keeps init and install consistent.
// --force is the approval grant.
func runInit(args []string, stdout, stderr io.Writer, jsonMode bool) int {
	profileID := profile.CoreID
	force := false
	target := "."
	haveTarget := false

	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--profile" || a == "-profile":
			if i+1 >= len(args) {
				return cmdInvalid("init", stdout, stderr, jsonMode, "--profile requires a value")
			}
			i++
			profileID = args[i]
		case strings.HasPrefix(a, "--profile="):
			profileID = strings.TrimPrefix(a, "--profile=")
		case strings.HasPrefix(a, "-profile="):
			profileID = strings.TrimPrefix(a, "-profile=")
		case a == "--force" || a == "-force":
			force = true
		case strings.HasPrefix(a, "-"):
			return cmdInvalid("init", stdout, stderr, jsonMode, fmt.Sprintf("unknown flag %q", a))
		default:
			if haveTarget {
				return cmdInvalid("init", stdout, stderr, jsonMode, fmt.Sprintf("unexpected argument %q", a))
			}
			target = a
			haveTarget = true
		}
	}

	p, err := profile.Resolve(profileID)
	if err != nil {
		return cmdInvalid("init", stdout, stderr, jsonMode, fmt.Sprintf("unknown profile %q", profileID))
	}

	// Init never creates the target dir: fsafe/txn require an existing boundary.
	info, err := os.Stat(target)
	if err != nil || !info.IsDir() {
		return cmdInvalid("init", stdout, stderr, jsonMode, fmt.Sprintf("target %q must be an existing directory", target))
	}

	changes := scaffold.Plan(p, time.Now().UTC())
	conflicts, err := scaffold.Conflicts(target, changes)
	if err != nil {
		return cmdSystemFailure("init", stdout, stderr, jsonMode, err)
	}
	if len(conflicts) > 0 && !force {
		return cmdRefuse("init", stdout, stderr, jsonMode, conflicts)
	}

	if err := commitScaffold(target, changes); err != nil {
		return cmdSystemFailure("init", stdout, stderr, jsonMode, err)
	}

	env := contract.New("init", contract.StatusSuccess)
	env.AffectedPaths = planTargets(changes)
	if jsonMode {
		return emit(stdout, env)
	}
	fmt.Fprintf(stdout, "init: created %d file(s) under %s\n", len(changes), target)
	for _, c := range changes {
		fmt.Fprintf(stdout, "  %s\n", c.Target)
	}
	return int(contract.ExitCodeForStatus(env.Status))
}

// commitScaffold writes the change set through one ADR-006 transaction. Any
// txn/fsafe error (including ErrStale from a TOCTOU race between the conflict
// check and commit — txn's internal rollback leaves the tree unmutated) is a
// system failure: the tree is left as it was, never re-rendered as a refusal.
// A best-effort Abort cleans up staging on the error path (ErrTxnDone, returned
// once a failed Commit has already finalized the transaction, is expected).
func commitScaffold(target string, changes []txn.FileChange) error {
	tx, err := txn.Begin(target, changes)
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		if abErr := tx.Abort(); abErr != nil && !errors.Is(abErr, txn.ErrTxnDone) {
			return fmt.Errorf("%w; abort also failed: %v", err, abErr)
		}
		return err
	}
	return nil
}

// planTargets returns the change set's targets in their existing sorted order —
// the affectedPaths the success envelope reports. It is taken from the plan,
// never from walking the tree.
func planTargets(changes []txn.FileChange) []string {
	paths := make([]string, len(changes))
	for i, c := range changes {
		paths[i] = c.Target
	}
	return paths
}

// cmdRefuse emits the approval-required refusal (exit 3) for operation op: the
// invocation is well-formed but the operation needs explicit permission to
// overwrite. init and install (ADR-009) share this envelope — a pre-existing
// planned target is a conflict, and --force is the approval grant.
func cmdRefuse(op string, stdout, stderr io.Writer, jsonMode bool, conflicts []string) int {
	env := contract.New(op, contract.StatusApprovalRequired)
	env.Approval = &contract.Approval{
		Required: true,
		Reason:   "target files already exist; re-run with --force to overwrite",
		Paths:    conflicts,
	}
	if jsonMode {
		return emit(stdout, env)
	}
	fmt.Fprintf(stderr, "llm-wiki: %s refused — %s\n", op, env.Approval.Reason)
	for _, p := range conflicts {
		fmt.Fprintf(stderr, "  %s\n", p)
	}
	return int(contract.ExitCodeForStatus(env.Status))
}

// cmdInvalid emits the invalid-invocation outcome (exit 4) for operation op:
// JSON carries the envelope; human mode explains the error and prints usage on
// stderr.
func cmdInvalid(op string, stdout, stderr io.Writer, jsonMode bool, msg string) int {
	if jsonMode {
		env := contract.New(op, contract.StatusInvalidInvocation)
		return emit(stdout, env)
	}
	fmt.Fprintf(stderr, "llm-wiki: %s\n\n%s", msg, usage)
	return int(contract.ExitInvalidInvocation)
}

// cmdSystemFailure maps a txn/fsafe error to the system-failure bucket (exit 5)
// for operation op.
func cmdSystemFailure(op string, stdout, stderr io.Writer, jsonMode bool, cause error) int {
	env := contract.New(op, contract.StatusSystemFailure)
	if jsonMode {
		return emit(stdout, env)
	}
	fmt.Fprintf(stderr, "llm-wiki: %s failed: %v\n", op, cause)
	return int(contract.ExitCodeForStatus(env.Status))
}

// runValidate walks a target directory (default ".") and reports OKF and
// core-profile findings through the contract envelope. It applies the ADR-004
// precedence in fixed order: engine core defaults → Resolve (configured profile
// overrides) → status. Configured severity overrides come from LLM_WIKI_SEVERITY;
// with none set the run keeps core-default severities. The default CLI path runs
// release-gate semantics and loads no adoption baseline, so structural errors
// always fail the gate (ADR-004).
func runValidate(args []string, stdout io.Writer, jsonMode bool) int {
	root := "."
	if len(args) > 0 {
		root = args[0]
	}

	// Anchor the repo-path resolution class at the nearest .llm-wiki/ marker
	// (ADR-008 sub-decision 3); ok=false falls back to the zero Options, which is
	// byte-identical to the pre-citation engine. Evidence contexts come from
	// LLM_WIKI_EVIDENCE_SECTIONS (the interim channel until profile loading lands
	// in Phase 4); page inspect and page plan wire this identically so all three
	// agree on which links are citations (criterion 15).
	opts, _ := validate.AnchorRepo(root)
	opts.EvidenceSections = evidenceSections()
	findings := validate.NewWithOptions(yamladapter.New(), opts).Run(os.DirFS(root))
	findings = validate.Resolve(findings, severityOverrides(os.Getenv("LLM_WIKI_SEVERITY")))
	status := validate.StatusFor(findings)

	env := contract.New("validate", status)
	env.Findings = findings

	if jsonMode {
		return emit(stdout, env)
	}
	fmt.Fprintf(stdout, "validate: %s — %d finding(s)\n", status, len(findings))
	for _, f := range findings {
		fmt.Fprintf(stdout, "  [%s/%s] %s: %s (%s)\n", f.Ruleset, f.Severity, f.Code, f.Message, f.Path)
	}
	return int(contract.ExitCodeForStatus(env.Status))
}

// runPage dispatches the page command group (ADR-006). Issue #42 lands the
// read-only inspect subcommand; page plan/apply arrive in later slices. An
// empty or unknown subcommand is invalid-invocation (exit 4), matching the
// unknown-profile precedent in runInit. The operation string is the
// space-separated invocation ("page inspect"), the convention every Phase 3
// page subcommand follows.
func runPage(args []string, stdout, stderr io.Writer, jsonMode bool) int {
	if len(args) == 0 {
		return cmdInvalid("page", stdout, stderr, jsonMode, "page requires a subcommand (inspect, plan, apply)")
	}
	sub, subArgs := args[0], args[1:]
	switch sub {
	case "inspect":
		return runPageInspect(subArgs, stdout, stderr, jsonMode)
	case "plan":
		return runPagePlan(subArgs, stdout, stderr, jsonMode)
	case "apply":
		return runPageApply(subArgs, stdout, stderr, jsonMode)
	default:
		return cmdInvalid("page", stdout, stderr, jsonMode, fmt.Sprintf("unknown page subcommand %q", sub))
	}
}

// runPageInspect reports a single page read-only. It parses its own flags in the
// hand-rolled selfcheckRoot/runInit style (the global loop only strips --json):
// --root/--root=<dir> (default "."), an optional --content/--content=<file> for
// inspecting proposed content (default: the live page; "-" reads stdin), and
// exactly one positional <path>. It maps the plan substrate's outcomes to the
// contract: a boundary escape is a system failure (exit 5, per fsafe's documented
// "every guard error → system-failure" rule), while a missing page, a non-.md
// path, or a bad invocation is invalid-invocation (exit 4). With --content the
// target may be absent (a new-page draft is validated in full bundle context
// without being staged); without it the live page must exist as before. On
// success it emits the standard envelope plus the optional page payload carrying
// path/parse-status/content-hash.
func runPageInspect(args []string, stdout, stderr io.Writer, jsonMode bool) int {
	root := "."
	content := ""
	haveContent := false
	var pagePath string
	havePath := false

	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--root" || a == "-root":
			if i+1 >= len(args) {
				return cmdInvalid("page inspect", stdout, stderr, jsonMode, "--root requires a value")
			}
			i++
			root = args[i]
		case strings.HasPrefix(a, "--root="):
			root = strings.TrimPrefix(a, "--root=")
		case strings.HasPrefix(a, "-root="):
			root = strings.TrimPrefix(a, "-root=")
		case a == "--content" || a == "-content":
			if i+1 >= len(args) {
				return cmdInvalid("page inspect", stdout, stderr, jsonMode, "--content requires a value")
			}
			i++
			content, haveContent = args[i], true
		case strings.HasPrefix(a, "--content="):
			content, haveContent = strings.TrimPrefix(a, "--content="), true
		case strings.HasPrefix(a, "-content="):
			content, haveContent = strings.TrimPrefix(a, "-content="), true
		case strings.HasPrefix(a, "-"):
			return cmdInvalid("page inspect", stdout, stderr, jsonMode, fmt.Sprintf("unknown flag %q", a))
		default:
			if havePath {
				return cmdInvalid("page inspect", stdout, stderr, jsonMode, fmt.Sprintf("unexpected argument %q", a))
			}
			pagePath = a
			havePath = true
		}
	}

	if !havePath {
		return cmdInvalid("page inspect", stdout, stderr, jsonMode, "page inspect requires a <path> argument")
	}

	// With --content the report comes from the proposed bytes over an overlay of
	// the live bundle (the authoring flow's validate step); without it, the live
	// page on disk is inspected as before.
	report, err := inspectReport(root, pagePath, content, haveContent)
	if err != nil {
		switch {
		case errors.Is(err, plan.ErrPageNotFound):
			return cmdInvalid("page inspect", stdout, stderr, jsonMode, fmt.Sprintf("no page at %q", pagePath))
		case errors.Is(err, plan.ErrNotMarkdown):
			return cmdInvalid("page inspect", stdout, stderr, jsonMode, fmt.Sprintf("%q is not a markdown page", pagePath))
		case errors.Is(err, errContentRead):
			return cmdInvalid("page inspect", stdout, stderr, jsonMode, err.Error())
		case errors.Is(err, fsafe.ErrOutsideBoundary), errors.Is(err, fsafe.ErrSymlinkEscape):
			// Boundary escape maps to system-failure per fsafe's documented rule.
			return cmdSystemFailure("page inspect", stdout, stderr, jsonMode, err)
		default:
			return cmdSystemFailure("page inspect", stdout, stderr, jsonMode, err)
		}
	}

	env := contract.New("page inspect", report.Status)
	env.Findings = report.Findings
	env.Page = &contract.PageReport{
		Path:        report.Path,
		Parsed:      report.Parsed,
		ContentHash: report.ContentHash,
	}

	if jsonMode {
		return emit(stdout, env)
	}
	parseState := "ok"
	if !report.Parsed {
		parseState = "failed"
	}
	fmt.Fprintf(stdout, "page inspect: %s — %s\n", report.Status, report.Path)
	fmt.Fprintf(stdout, "  parse: %s\n", parseState)
	fmt.Fprintf(stdout, "  hash:  sha256 %s\n", report.ContentHash)
	fmt.Fprintf(stdout, "  %d finding(s)\n", len(report.Findings))
	for _, f := range report.Findings {
		fmt.Fprintf(stdout, "  [%s/%s] %s: %s (%s)\n", f.Ruleset, f.Severity, f.Code, f.Message, f.Path)
	}
	return int(contract.ExitCodeForStatus(env.Status))
}

// inspectReport produces the page-inspect report for the live page or, when
// --content was given, for the proposed bytes. The evidence-section channel and
// severity overrides are wired identically for both paths so a draft is judged by
// the same rules as its committed form.
func inspectReport(root, pagePath, content string, haveContent bool) (*plan.Report, error) {
	overrides := severityOverrides(os.Getenv("LLM_WIKI_SEVERITY"))
	if !haveContent {
		return plan.Inspect(root, pagePath, yamladapter.New(), evidenceSections(), overrides)
	}
	proposed, err := readProposedContent(content, haveContent)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errContentRead, err)
	}
	return plan.InspectContent(root, pagePath, proposed, yamladapter.New(), evidenceSections(), overrides)
}

// errContentRead marks a failure to read the --content source (a bad file path or
// stdin error): a well-formed but unsatisfiable invocation, mapped to exit 4.
var errContentRead = errors.New("cannot read proposed content")

// runPagePlan previews a whole-page change without mutating the live tree
// (ADR-006). It parses its own flags in the hand-rolled runPageInspect style:
// --root/--root=<dir> (default "."), --content/--content=<file> for the proposed
// page (default stdin, or "-"), and exactly one positional <path>. It reads the
// proposed content, delegates to plan.Plan (which normalizes frontmatter through
// the yamladapter round-trip, binds the base hash, stages the change under
// .llm-wiki/staging/<txn-id>/, and renders a unified diff), and maps the
// substrate's outcomes to the contract: a boundary escape is a system failure
// (exit 5); a non-.md path, a non-regular target, or a bad invocation is
// invalid-invocation (exit 4). A successful plan — staged change or no-op —
// emits the standard envelope plus the plan payload.
func runPagePlan(args []string, stdout, stderr io.Writer, jsonMode bool) int {
	root := "."
	content := ""
	haveContent := false
	var pagePath string
	havePath := false

	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--root" || a == "-root":
			if i+1 >= len(args) {
				return cmdInvalid("page plan", stdout, stderr, jsonMode, "--root requires a value")
			}
			i++
			root = args[i]
		case strings.HasPrefix(a, "--root="):
			root = strings.TrimPrefix(a, "--root=")
		case strings.HasPrefix(a, "-root="):
			root = strings.TrimPrefix(a, "-root=")
		case a == "--content" || a == "-content":
			if i+1 >= len(args) {
				return cmdInvalid("page plan", stdout, stderr, jsonMode, "--content requires a value")
			}
			i++
			content, haveContent = args[i], true
		case strings.HasPrefix(a, "--content="):
			content, haveContent = strings.TrimPrefix(a, "--content="), true
		case strings.HasPrefix(a, "-content="):
			content, haveContent = strings.TrimPrefix(a, "-content="), true
		case strings.HasPrefix(a, "-"):
			return cmdInvalid("page plan", stdout, stderr, jsonMode, fmt.Sprintf("unknown flag %q", a))
		default:
			if havePath {
				return cmdInvalid("page plan", stdout, stderr, jsonMode, fmt.Sprintf("unexpected argument %q", a))
			}
			pagePath, havePath = a, true
		}
	}

	if !havePath {
		return cmdInvalid("page plan", stdout, stderr, jsonMode, "page plan requires a <path> argument")
	}

	proposed, err := readProposedContent(content, haveContent)
	if err != nil {
		return cmdInvalid("page plan", stdout, stderr, jsonMode, fmt.Sprintf("cannot read proposed content: %v", err))
	}

	// Wire the ADR-008 citation resolution the same way validate/inspect do so the
	// gate never fires on citations validate cannot see (criterion 15): the
	// repo-path anchor plus the interim evidence-section channel.
	opts, _ := validate.AnchorRepo(root)
	opts.EvidenceSections = evidenceSections()
	result, err := plan.Plan(root, pagePath, proposed, yamladapter.New(), opts)
	if err != nil {
		switch {
		case errors.Is(err, plan.ErrNotMarkdown):
			return cmdInvalid("page plan", stdout, stderr, jsonMode, fmt.Sprintf("%q is not a markdown page", pagePath))
		case errors.Is(err, plan.ErrTargetNotRegular), errors.Is(err, txn.ErrNonRegularTarget),
			errors.Is(err, txn.ErrReservedPath):
			return cmdInvalid("page plan", stdout, stderr, jsonMode, fmt.Sprintf("cannot plan over %q: %v", pagePath, err))
		case errors.Is(err, fsafe.ErrOutsideBoundary), errors.Is(err, fsafe.ErrSymlinkEscape):
			// Boundary escape maps to system-failure per fsafe's documented rule.
			return cmdSystemFailure("page plan", stdout, stderr, jsonMode, err)
		default:
			return cmdSystemFailure("page plan", stdout, stderr, jsonMode, err)
		}
	}

	env := contract.New("page plan", contract.StatusSuccess)
	if !result.NoOp {
		env.AffectedPaths = []string{result.Path}
	}
	env.Plan = &contract.PagePlan{
		Path:        result.Path,
		NoOp:        result.NoOp,
		Transaction: result.TxnID,
		BaseAbsent:  result.BaseAbsent,
		BaseHash:    result.BaseHash,
		StagedHash:  result.StagedHash,
		Diff:        result.Diff,
	}

	// Citation-loss gate (ADR-008 sub-decision 6): Plan wrote the approval sidecar
	// into the staging dir, so surface the loss as a warning finding (promotable
	// via LLM_WIKI_SEVERITY, though the gate is the approval member, not severity)
	// and set the approval requirement. The plan itself succeeded and is staged —
	// the status stays success/exit 0; only apply refuses until --approve.
	if result.LostCitations != nil {
		reason := plan.CitationLossReason(result.LostCitations)
		loss := validate.Resolve([]contract.Finding{{
			Ruleset:  contract.RulesetProfile,
			Severity: contract.SeverityWarning,
			Code:     validate.CodeCitationLoss,
			Message:  reason,
			Path:     result.Path,
		}}, severityOverrides(os.Getenv("LLM_WIKI_SEVERITY")))
		env.Findings = append(env.Findings, loss...)
		env.Approval = &contract.Approval{
			Required: true,
			Reason:   reason,
			Paths:    []string{result.Path},
		}
	}

	if jsonMode {
		return emit(stdout, env)
	}
	if result.NoOp {
		fmt.Fprintf(stdout, "page plan: no changes — %s\n", result.Path)
		return int(contract.ExitCodeForStatus(env.Status))
	}
	base := "sha256 " + result.BaseHash
	if result.BaseAbsent {
		base = "absent (new page)"
	}
	fmt.Fprintf(stdout, "page plan: staged change — %s\n", result.Path)
	fmt.Fprintf(stdout, "  txn:    %s\n", result.TxnID)
	fmt.Fprintf(stdout, "  base:   %s\n", base)
	fmt.Fprintf(stdout, "  staged: sha256 %s\n", result.StagedHash)
	if result.LostCitations != nil {
		fmt.Fprintf(stdout, "  approval required at apply: citation loss — %s (re-run apply with --approve)\n",
			strings.Join(result.LostCitations, ", "))
	}
	fmt.Fprint(stdout, result.Diff)
	return int(contract.ExitCodeForStatus(env.Status))
}

// readProposedContent loads the proposed page bytes: from the --content file
// when a real path was given, or from stdin when the flag was omitted or set to
// "-" (the pipe convention).
func readProposedContent(content string, haveContent bool) ([]byte, error) {
	if !haveContent || content == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(content)
}

// runPageApply commits a staged page plan into the live tree (ADR-006). It parses
// its own flags in the hand-rolled runPagePlan style: --root/--root=<dir>
// (default "."), --approve to grant a recorded approval requirement, and exactly
// one positional <txn-id> (the id a prior page plan reported). It delegates to
// plan.Apply, which re-verifies the recorded base hashes against the live tree
// before mutating, and maps the outcomes to the contract:
//   - a clean apply commits exactly the previewed change set → success with the
//     apply payload and affectedPaths;
//   - an un-granted approval requirement → approval-required (exit 3), the
//     approval envelope field, and zero mutation;
//   - a stale plan, a missing transaction, or a target that turned non-regular →
//     invalid-invocation (exit 4) with zero mutation — the invocation was
//     well-formed but the plan can no longer be applied (criterion 13);
//   - a boundary escape or other txn/fsafe error → system-failure (exit 5).
func runPageApply(args []string, stdout, stderr io.Writer, jsonMode bool) int {
	root := "."
	approved := false
	var txnID string
	haveTxn := false

	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--root" || a == "-root":
			if i+1 >= len(args) {
				return cmdInvalid("page apply", stdout, stderr, jsonMode, "--root requires a value")
			}
			i++
			root = args[i]
		case strings.HasPrefix(a, "--root="):
			root = strings.TrimPrefix(a, "--root=")
		case strings.HasPrefix(a, "-root="):
			root = strings.TrimPrefix(a, "-root=")
		case a == "--approve" || a == "-approve":
			approved = true
		case strings.HasPrefix(a, "-"):
			return cmdInvalid("page apply", stdout, stderr, jsonMode, fmt.Sprintf("unknown flag %q", a))
		default:
			if haveTxn {
				return cmdInvalid("page apply", stdout, stderr, jsonMode, fmt.Sprintf("unexpected argument %q", a))
			}
			txnID, haveTxn = a, true
		}
	}

	if !haveTxn {
		return cmdInvalid("page apply", stdout, stderr, jsonMode, "page apply requires a <txn-id> argument")
	}

	res, err := plan.Apply(root, txnID, approved)
	if err != nil {
		switch {
		case errors.Is(err, txn.ErrTxnNotFound):
			return cmdReject("page apply", stdout, stderr, jsonMode, fmt.Sprintf("no staged plan %q (never planned, or already applied)", txnID))
		case errors.Is(err, txn.ErrStale):
			return cmdReject("page apply", stdout, stderr, jsonMode, fmt.Sprintf("plan %q is stale — a target changed since it was planned; re-plan and re-apply", txnID))
		case errors.Is(err, txn.ErrNonRegularTarget):
			return cmdReject("page apply", stdout, stderr, jsonMode, fmt.Sprintf("plan %q target is no longer a regular file; re-plan", txnID))
		case errors.Is(err, fsafe.ErrOutsideBoundary), errors.Is(err, fsafe.ErrSymlinkEscape):
			// Boundary escape maps to system-failure per fsafe's documented rule.
			return cmdSystemFailure("page apply", stdout, stderr, jsonMode, err)
		default:
			return cmdSystemFailure("page apply", stdout, stderr, jsonMode, err)
		}
	}

	if res.ApprovalRequired != nil {
		env := contract.New("page apply", contract.StatusApprovalRequired)
		env.Approval = &contract.Approval{
			Required: true,
			Reason:   res.ApprovalRequired.Reason,
			Paths:    res.ApprovalRequired.Paths,
		}
		if jsonMode {
			return emit(stdout, env)
		}
		fmt.Fprintf(stderr, "llm-wiki: page apply refused — %s\n", res.ApprovalRequired.Reason)
		fmt.Fprintf(stderr, "  re-run with --approve to proceed\n")
		for _, p := range res.ApprovalRequired.Paths {
			fmt.Fprintf(stderr, "  %s\n", p)
		}
		return int(contract.ExitCodeForStatus(env.Status))
	}

	env := contract.New("page apply", contract.StatusSuccess)
	env.AffectedPaths = res.AppliedPaths
	env.Apply = &contract.PageApply{Transaction: res.TxnID, Committed: res.AppliedPaths}
	if jsonMode {
		return emit(stdout, env)
	}
	fmt.Fprintf(stdout, "page apply: applied %d file(s) — txn %s\n", len(res.AppliedPaths), res.TxnID)
	for _, p := range res.AppliedPaths {
		fmt.Fprintf(stdout, "  %s\n", p)
	}
	return int(contract.ExitCodeForStatus(env.Status))
}

// cmdReject emits an invalid-invocation outcome (exit 4) for a well-formed
// invocation the engine cannot carry out — a stale plan or a missing staging
// transaction. Unlike cmdInvalid it prints a concise reason without the full
// usage text, since the command line itself was valid; only the referenced plan
// is no longer applicable.
func cmdReject(op string, stdout, stderr io.Writer, jsonMode bool, msg string) int {
	if jsonMode {
		env := contract.New(op, contract.StatusInvalidInvocation)
		return emit(stdout, env)
	}
	fmt.Fprintf(stderr, "llm-wiki: %s: %s\n", op, msg)
	return int(contract.ExitInvalidInvocation)
}

// runSelfcheck detects the running platform and verifies its shipped binary
// against the bundled SHA256SUMS manifest (ADR-002). It fails closed: any
// verification error maps to the contract's system-or-filesystem-failure bucket
// (exit 5). Consistent with the other commands, the outcome is carried by the
// envelope status rather than a validation finding — findings are OKF/profile
// results (ADR-004), not integrity failures.
func runSelfcheck(args []string, stdout, stderr io.Writer, jsonMode bool) int {
	root, err := selfcheckRoot(args)
	if err == nil {
		var p platform.Platform
		p, err = platform.VerifyBundle(os.DirFS(root))
		if err == nil {
			return selfcheckOK(stdout, jsonMode, p.ArtifactPath())
		}
	}
	return selfcheckFail(stdout, stderr, jsonMode, err)
}

// selfcheckRoot resolves the bundle root: an explicit --root/--root=<dir> flag,
// otherwise the directory containing the running executable (where the plugin
// ships bin/ alongside the engine).
func selfcheckRoot(args []string) (string, error) {
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--root" || a == "-root":
			if i+1 >= len(args) {
				return "", fmt.Errorf("--root requires a directory argument")
			}
			return args[i+1], nil
		case strings.HasPrefix(a, "--root="):
			return strings.TrimPrefix(a, "--root="), nil
		case strings.HasPrefix(a, "-root="):
			return strings.TrimPrefix(a, "-root="), nil
		}
	}
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("locate executable: %w", err)
	}
	return filepath.Dir(exe), nil
}

func selfcheckOK(stdout io.Writer, jsonMode bool, artifact string) int {
	env := contract.New("selfcheck", contract.StatusSuccess)
	if jsonMode {
		return emit(stdout, env)
	}
	fmt.Fprintf(stdout, "selfcheck: OK — verified %s\n", artifact)
	return int(contract.ExitCodeForStatus(env.Status))
}

func selfcheckFail(stdout, stderr io.Writer, jsonMode bool, cause error) int {
	env := contract.New("selfcheck", contract.StatusSystemFailure)
	if jsonMode {
		return emit(stdout, env)
	}
	fmt.Fprintf(stderr, "llm-wiki: selfcheck failed: %v\n", cause)
	return int(contract.ExitCodeForStatus(env.Status))
}

func runUnknown(command string, stdout, stderr io.Writer, jsonMode bool) int {
	if jsonMode {
		env := contract.New(command, contract.StatusInvalidInvocation)
		return emit(stdout, env)
	}
	fmt.Fprintf(stderr, "llm-wiki: unknown command %q\n\n%s", command, usage)
	return int(contract.ExitInvalidInvocation)
}

// emit writes the envelope as JSON and returns the exit code its status maps to.
func emit(stdout io.Writer, env *contract.Envelope) int {
	if err := env.WriteJSON(stdout); err != nil {
		return int(contract.ExitSystemFailure)
	}
	return int(contract.ExitCodeForStatus(env.Status))
}
