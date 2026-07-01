package contract

// ExitCode is the process exit status the CLI returns. ADR-003 fixed the six
// semantic buckets and deferred the exact numeric values to this issue. The
// values below are now a frozen public surface: callers branch on them, so
// changing one is a breaking contract change.
type ExitCode int

const (
	// ExitSuccess: the operation completed with no findings above suggestion.
	ExitSuccess ExitCode = 0
	// ExitSuccessWithWarnings: completed, but non-failing warnings were reported.
	ExitSuccessWithWarnings ExitCode = 1
	// ExitValidationFailure: at least one error-severity finding (e.g. invalid
	// YAML or a missing profile-required field).
	ExitValidationFailure ExitCode = 2
	// ExitApprovalRequired: the operation needs approval before it can proceed.
	ExitApprovalRequired ExitCode = 3
	// ExitInvalidInvocation: the command line or configuration was invalid.
	ExitInvalidInvocation ExitCode = 4
	// ExitSystemFailure: an unexpected system or filesystem error occurred.
	ExitSystemFailure ExitCode = 5
)

// ExitCodeInfo is one row of the canonical code->meaning table. The published
// docs/contract/exit-codes.md mirrors this table; a test enforces they agree.
type ExitCodeInfo struct {
	Code    ExitCode
	Status  Status
	Meaning string
}

// ExitCodes is the canonical, ordered code->meaning table. It is the single
// source of truth the documentation mirrors.
var ExitCodes = []ExitCodeInfo{
	{ExitSuccess, StatusSuccess, "Operation completed with no failing findings."},
	{ExitSuccessWithWarnings, StatusSuccessWithWarnings, "Operation completed; non-failing warnings were reported."},
	{ExitValidationFailure, StatusValidationFailure, "Validation failed on at least one error-severity finding."},
	{ExitApprovalRequired, StatusApprovalRequired, "The operation requires approval before it can proceed."},
	{ExitInvalidInvocation, StatusInvalidInvocation, "The invocation or configuration was invalid."},
	{ExitSystemFailure, StatusSystemFailure, "An unexpected system or filesystem error occurred."},
}

// ExitCodeForStatus maps an envelope status to its process exit code. An
// unrecognised status is treated defensively as a system failure so a bug can
// never surface as a false success.
func ExitCodeForStatus(s Status) ExitCode {
	for _, info := range ExitCodes {
		if info.Status == s {
			return info.Code
		}
	}
	return ExitSystemFailure
}
