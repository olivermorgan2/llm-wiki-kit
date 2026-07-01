// Package contract defines the versioned JSON envelope and the stable
// exit-code set that every llm-wiki surface (CLI, skills, hooks, CI) shares.
//
// The envelope is the single schema decided in ADR-003: emitting the same
// shape everywhere is what makes "the same repository state yields the same
// findings across surfaces" achievable. Human-readable output is the default
// for interactive use; this JSON envelope is opt-in via --json.
package contract

import (
	"encoding/json"
	"io"
)

// ContractVersion is the schema version carried by every envelope. Per ADR-003
// the contract starts at v1 with no backward-compatibility guarantee until the
// first public release; skills and engine ship in lockstep.
const ContractVersion = "v1"

// Status is the outcome an envelope reports. Each value is one of the six
// ADR-003 semantic buckets and maps 1:1 to an ExitCode (see exitcode.go).
type Status string

const (
	StatusSuccess             Status = "success"
	StatusSuccessWithWarnings Status = "success-with-warnings"
	StatusValidationFailure   Status = "validation-failure"
	StatusApprovalRequired    Status = "approval-required"
	StatusInvalidInvocation   Status = "invalid-invocation"
	StatusSystemFailure       Status = "system-or-filesystem-failure"
)

// Ruleset tags whether a finding comes from base OKF conformance or from a
// profile, so the two can be reported separately (ADR-004, criterion 5).
type Ruleset string

const (
	RulesetOKF     Ruleset = "okf"
	RulesetProfile Ruleset = "profile"
)

// Severity is one of the three configurable levels in ADR-004's validation
// model. Rule content is deferred to later Phase 1 issues; the skeleton only
// fixes the shape the envelope carries.
type Severity string

const (
	SeverityError      Severity = "error"
	SeverityWarning    Severity = "warning"
	SeveritySuggestion Severity = "suggestion"
)

// Finding is a single validation result carried by the envelope. The concrete
// rules that produce findings are later Phase 1 work (ADR-004); this type only
// nails down the wire shape.
type Finding struct {
	Ruleset  Ruleset  `json:"ruleset"`
	Severity Severity `json:"severity"`
	Code     string   `json:"code"`
	Message  string   `json:"message"`
	Path     string   `json:"path,omitempty"`
}

// Approval describes an approval requirement an operation raised. It is nil in
// the envelope when no approval is involved. The approval *policy* (what
// triggers it, how it is granted) is owned by later ADRs; this is only the
// contract shape.
type Approval struct {
	Required bool     `json:"required"`
	Reason   string   `json:"reason,omitempty"`
	Paths    []string `json:"paths,omitempty"`
}

// Envelope is the versioned response shape shared by every surface. It always
// carries exactly the six ADR-003 fields: contractVersion, operation, status,
// findings, affectedPaths, approval.
type Envelope struct {
	ContractVersion string    `json:"contractVersion"`
	Operation       string    `json:"operation"`
	Status          Status    `json:"status"`
	Findings        []Finding `json:"findings"`
	AffectedPaths   []string  `json:"affectedPaths"`
	Approval        *Approval `json:"approval"`
}

// New returns an envelope stamped with the current contract version for the
// given operation and status. Findings and AffectedPaths are initialised to
// non-nil empty slices so they serialize as [] rather than null; Approval is
// nil until an operation sets it.
func New(operation string, status Status) *Envelope {
	return &Envelope{
		ContractVersion: ContractVersion,
		Operation:       operation,
		Status:          status,
		Findings:        []Finding{},
		AffectedPaths:   []string{},
		Approval:        nil,
	}
}

// WriteJSON writes the envelope as indented JSON followed by a trailing
// newline. Output is deterministic: struct fields serialize in declaration
// order and the envelope holds no maps.
func (e *Envelope) WriteJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(e)
}
