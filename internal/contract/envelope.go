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

// PageReport is the read-only page report an inspect-style page operation
// carries in the envelope. It records the page's bundle-relative path, whether
// its frontmatter parsed, and the content hash (lowercase-hex SHA-256, no
// prefix) — the base-hash input a later ADR-006 page plan/apply binds against.
// It is an operation-scoped payload: present only for page commands, nil (and
// omitted from JSON) everywhere else, so the six ADR-003 fields remain the
// invariant shape.
type PageReport struct {
	Path        string `json:"path"`
	Parsed      bool   `json:"parsed"`
	ContentHash string `json:"contentHash"`
}

// PagePlan is the staged-mutation preview a page plan operation carries in the
// envelope (ADR-006). It records the target's bundle-relative path, whether the
// plan is a no-op (proposed content already matches the live page), the staging
// transaction id binding the plan on disk (empty for a no-op), the captured
// base state (BaseAbsent marks a new page's absent-target sentinel; BaseHash is
// the live page's content hash otherwise), the staged-content hash, and a
// unified diff preview. Like Page it is operation-scoped: present only for page
// plan, nil (and omitted from JSON) everywhere else, so the six ADR-003 fields
// stay the invariant shape. A plan never mutates live files — the diff and the
// staged bytes live under .llm-wiki/staging/<transaction>/ until a later apply.
type PagePlan struct {
	Path        string `json:"path"`
	NoOp        bool   `json:"noOp"`
	Transaction string `json:"transaction,omitempty"`
	BaseAbsent  bool   `json:"baseAbsent"`
	BaseHash    string `json:"baseHash,omitempty"`
	StagedHash  string `json:"stagedHash"`
	Diff        string `json:"diff"`
}

// Envelope is the versioned response shape shared by every surface. It always
// carries exactly the six ADR-003 fields: contractVersion, operation, status,
// findings, affectedPaths, approval. Page-scoped operations additionally carry
// exactly one optional payload field — page (the read-only inspect report) or
// plan (the staged-mutation preview) — which is nil (and omitted from JSON) for
// every non-page operation, so the six-field shape is preserved everywhere else
// (ADR-006 anticipated the envelope carrying page-plan payloads; page inspect
// filled the first, read-only slot and page plan fills the second).
type Envelope struct {
	ContractVersion string      `json:"contractVersion"`
	Operation       string      `json:"operation"`
	Status          Status      `json:"status"`
	Findings        []Finding   `json:"findings"`
	AffectedPaths   []string    `json:"affectedPaths"`
	Approval        *Approval   `json:"approval"`
	Page            *PageReport `json:"page,omitempty"`
	Plan            *PagePlan   `json:"plan,omitempty"`
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
