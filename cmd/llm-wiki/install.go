package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/olivermorgan2/llm-wiki-kit/internal/contract"
	"github.com/olivermorgan2/llm-wiki-kit/internal/manifest"
	"github.com/olivermorgan2/llm-wiki-kit/internal/platform"
	"github.com/olivermorgan2/llm-wiki-kit/internal/profile"
	"github.com/olivermorgan2/llm-wiki-kit/internal/scaffold"
)

// runInstall installs the kit's bundle into a new or non-empty repository per
// ADR-009. It plans the scaffold change set, appends the .llm-wiki/manifest.json
// version-record, and commits all of it through one ADR-006 transaction so a
// non-empty-repo install is all-or-nothing and loses no file. It shares init's
// flag style, exit-3 approval refusal, and exit-code mapping (cmdRefuse /
// cmdInvalid / cmdSystemFailure).
//
// It never silently overwrites: any pre-existing planned target — including an
// existing manifest, which means "already installed" — refuses with
// approval-required (exit 3) and --force is the grant. --dry-run computes the
// identical plan and refusal but mutates nothing, never calling txn.Begin (Begin
// stages under .llm-wiki/, which is itself a write). platform.Detect runs as a
// fail-closed ADR-002 precondition; an unsupported platform is a system failure
// (exit 5). The platform key is machine-specific and is deliberately not
// recorded in the manifest.
func runInstall(args []string, stdout, stderr io.Writer, jsonMode bool) int {
	profileID := profile.CoreID
	force := false
	dryRun := false
	target := "."
	haveTarget := false

	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--profile" || a == "-profile":
			if i+1 >= len(args) {
				return cmdInvalid("install", stdout, stderr, jsonMode, "--profile requires a value")
			}
			i++
			profileID = args[i]
		case strings.HasPrefix(a, "--profile="):
			profileID = strings.TrimPrefix(a, "--profile=")
		case strings.HasPrefix(a, "-profile="):
			profileID = strings.TrimPrefix(a, "-profile=")
		case a == "--force" || a == "-force":
			force = true
		case a == "--dry-run" || a == "-dry-run":
			dryRun = true
		case strings.HasPrefix(a, "-"):
			return cmdInvalid("install", stdout, stderr, jsonMode, fmt.Sprintf("unknown flag %q", a))
		default:
			if haveTarget {
				return cmdInvalid("install", stdout, stderr, jsonMode, fmt.Sprintf("unexpected argument %q", a))
			}
			target = a
			haveTarget = true
		}
	}

	p, err := profile.Resolve(profileID)
	if err != nil {
		return cmdInvalid("install", stdout, stderr, jsonMode, fmt.Sprintf("unknown profile %q", profileID))
	}

	// Install never creates the target dir: fsafe/txn require an existing
	// boundary (the repo the plugin installs into).
	info, err := os.Stat(target)
	if err != nil || !info.IsDir() {
		return cmdInvalid("install", stdout, stderr, jsonMode, fmt.Sprintf("target %q must be an existing directory", target))
	}

	// ADR-002 fail-closed precondition: install only proceeds on a supported
	// platform. Unsupported OS/arch is a system failure (exit 5), the same
	// bucket selfcheck uses for integrity failures.
	if _, err := platform.Detect(); err != nil {
		return cmdSystemFailure("install", stdout, stderr, jsonMode, err)
	}

	// The scaffold plus the manifest recording it, committed as one set. Hashes
	// in the manifest are computed from this exact in-memory slice, so they
	// match the committed bytes despite the date-stamped scaffold templates.
	changes := scaffold.Plan(p, time.Now().UTC())
	m := manifest.Build(manifest.Versions{
		Plugin:  version,
		CLI:     version,
		OKF:     scaffold.OKFVersion,
		Profile: manifest.ProfileRecord{ID: p.ID, Version: p.Version},
	}, changes)
	mc, err := m.Change()
	if err != nil {
		return cmdSystemFailure("install", stdout, stderr, jsonMode, err)
	}
	changes = append(changes, mc)
	sort.Slice(changes, func(i, j int) bool { return changes[i].Target < changes[j].Target })

	conflicts, err := scaffold.Conflicts(target, changes)
	if err != nil {
		return cmdSystemFailure("install", stdout, stderr, jsonMode, err)
	}
	if len(conflicts) > 0 && !force {
		return cmdRefuse("install", stdout, stderr, jsonMode, conflicts)
	}

	// --dry-run mirrors the real outcome (same plan, same refusal above) but
	// mutates nothing: it returns here, before txn.Begin would stage anything.
	if dryRun {
		env := contract.New("install", contract.StatusSuccess)
		env.AffectedPaths = planTargets(changes)
		if jsonMode {
			return emit(stdout, env)
		}
		fmt.Fprintf(stdout, "install (dry-run): would create %d file(s) under %s\n", len(changes), target)
		for _, c := range changes {
			fmt.Fprintf(stdout, "  %s\n", c.Target)
		}
		return int(contract.ExitCodeForStatus(env.Status))
	}

	if err := commitScaffold(target, changes); err != nil {
		return cmdSystemFailure("install", stdout, stderr, jsonMode, err)
	}

	env := contract.New("install", contract.StatusSuccess)
	env.AffectedPaths = planTargets(changes)
	if jsonMode {
		return emit(stdout, env)
	}
	fmt.Fprintf(stdout, "install: created %d file(s) under %s\n", len(changes), target)
	for _, c := range changes {
		fmt.Fprintf(stdout, "  %s\n", c.Target)
	}
	return int(contract.ExitCodeForStatus(env.Status))
}
