package contract

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// The exact numeric values were deferred by ADR-003 to this implementation
// issue. Once published they are a frozen public surface; this test is the
// freeze. Changing a value here is a breaking contract change.
func TestExitCodeValuesAreFrozen(t *testing.T) {
	cases := []struct {
		got  ExitCode
		want int
		name string
	}{
		{ExitSuccess, 0, "success"},
		{ExitSuccessWithWarnings, 1, "success-with-warnings"},
		{ExitValidationFailure, 2, "validation-failure"},
		{ExitApprovalRequired, 3, "approval-required"},
		{ExitInvalidInvocation, 4, "invalid-invocation"},
		{ExitSystemFailure, 5, "system-or-filesystem-failure"},
	}
	for _, c := range cases {
		if int(c.got) != c.want {
			t.Errorf("exit code %s = %d, want %d (frozen public surface)", c.name, int(c.got), c.want)
		}
	}
}

// Each of the six semantic buckets (ADR-003) maps 1:1 to its PRD §12 outcome
// via the envelope status.
func TestExitCodeForStatusMapsEveryBucket(t *testing.T) {
	cases := []struct {
		status Status
		want   ExitCode
	}{
		{StatusSuccess, ExitSuccess},
		{StatusSuccessWithWarnings, ExitSuccessWithWarnings},
		{StatusValidationFailure, ExitValidationFailure},
		{StatusApprovalRequired, ExitApprovalRequired},
		{StatusInvalidInvocation, ExitInvalidInvocation},
		{StatusSystemFailure, ExitSystemFailure},
	}
	for _, c := range cases {
		if got := ExitCodeForStatus(c.status); got != c.want {
			t.Errorf("ExitCodeForStatus(%q) = %d, want %d", c.status, int(got), int(c.want))
		}
	}
}

// An unknown status is a defensive system failure, never a false success.
func TestExitCodeForUnknownStatusIsSystemFailure(t *testing.T) {
	if got := ExitCodeForStatus(Status("nonsense")); got != ExitSystemFailure {
		t.Errorf("ExitCodeForStatus(unknown) = %d, want %d", int(got), int(ExitSystemFailure))
	}
}

// The canonical table must cover all six buckets, with unique codes and
// statuses, and agree with ExitCodeForStatus.
func TestExitCodesTableIsCompleteAndConsistent(t *testing.T) {
	if len(ExitCodes) != 6 {
		t.Fatalf("ExitCodes table has %d entries, want 6", len(ExitCodes))
	}
	seenCode := map[ExitCode]bool{}
	seenStatus := map[Status]bool{}
	for _, info := range ExitCodes {
		if seenCode[info.Code] {
			t.Errorf("duplicate exit code %d in table", int(info.Code))
		}
		if seenStatus[info.Status] {
			t.Errorf("duplicate status %q in table", info.Status)
		}
		seenCode[info.Code] = true
		seenStatus[info.Status] = true

		if got := ExitCodeForStatus(info.Status); got != info.Code {
			t.Errorf("table says %q -> %d, but ExitCodeForStatus returns %d",
				info.Status, int(info.Code), int(got))
		}
		if strings.TrimSpace(info.Meaning) == "" {
			t.Errorf("table entry for %q has no meaning", info.Status)
		}
	}
}

// The published documentation table must match the constants, so the docs can
// never silently drift from the frozen surface.
func TestDocumentedExitCodeTableMatchesConstants(t *testing.T) {
	const docPath = "../../docs/contract/exit-codes.md"
	raw, err := os.ReadFile(docPath)
	if err != nil {
		t.Fatalf("reading %s: %v", docPath, err)
	}
	doc := string(raw)

	for _, info := range ExitCodes {
		codeCell := fmt.Sprintf("| %d |", int(info.Code))
		if !strings.Contains(doc, codeCell) {
			t.Errorf("doc is missing a table row for exit code %d", int(info.Code))
		}
		if !strings.Contains(doc, string(info.Status)) {
			t.Errorf("doc is missing the %q bucket name", info.Status)
		}
	}
}
