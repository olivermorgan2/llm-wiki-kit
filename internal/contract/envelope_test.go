package contract

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// The envelope must always carry exactly the six ADR-003 contract fields.
func TestNewEnvelopeCarriesAllSixContractFields(t *testing.T) {
	env := New("version", StatusSuccess)

	var buf bytes.Buffer
	if err := env.WriteJSON(&buf); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}

	var generic map[string]json.RawMessage
	if err := json.Unmarshal(buf.Bytes(), &generic); err != nil {
		t.Fatalf("emitted JSON does not parse: %v\n%s", err, buf.String())
	}

	for _, field := range []string{
		"contractVersion", "operation", "status",
		"findings", "affectedPaths", "approval",
	} {
		if _, ok := generic[field]; !ok {
			t.Errorf("envelope JSON is missing required field %q; got: %s", field, buf.String())
		}
	}
	if len(generic) != 6 {
		t.Errorf("envelope must carry exactly 6 fields, got %d: %s", len(generic), buf.String())
	}
}

// The optional page payload is omitted entirely when nil: a non-page envelope
// still serializes exactly the six ADR-003 fields, never a seventh "page" key.
func TestNonPageEnvelopeOmitsPageField(t *testing.T) {
	env := New("validate", StatusSuccess)

	var buf bytes.Buffer
	if err := env.WriteJSON(&buf); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}

	var generic map[string]json.RawMessage
	if err := json.Unmarshal(buf.Bytes(), &generic); err != nil {
		t.Fatalf("emitted JSON does not parse: %v\n%s", err, buf.String())
	}
	if _, ok := generic["page"]; ok {
		t.Errorf("non-page envelope must omit the page field; got: %s", buf.String())
	}
	if len(generic) != 6 {
		t.Errorf("non-page envelope must carry exactly 6 fields, got %d: %s", len(generic), buf.String())
	}
}

// A page-scoped envelope carries the six core fields plus exactly one optional
// seventh field, page, and nothing else.
func TestPageEnvelopeCarriesSevenFields(t *testing.T) {
	env := New("page inspect", StatusSuccess)
	env.Page = &PageReport{Path: "wiki/index.md", Parsed: true, ContentHash: "abc123"}

	var buf bytes.Buffer
	if err := env.WriteJSON(&buf); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}

	var generic map[string]json.RawMessage
	if err := json.Unmarshal(buf.Bytes(), &generic); err != nil {
		t.Fatalf("emitted JSON does not parse: %v\n%s", err, buf.String())
	}
	for _, field := range []string{
		"contractVersion", "operation", "status",
		"findings", "affectedPaths", "approval", "page",
	} {
		if _, ok := generic[field]; !ok {
			t.Errorf("page envelope JSON is missing field %q; got: %s", field, buf.String())
		}
	}
	if len(generic) != 7 {
		t.Errorf("page envelope must carry exactly 7 fields, got %d: %s", len(generic), buf.String())
	}
}

// A page plan envelope carries the six core fields plus exactly one optional
// seventh field, plan, and never the page field (the two payloads are mutually
// exclusive).
func TestPagePlanEnvelopeCarriesSevenFields(t *testing.T) {
	env := New("page plan", StatusSuccess)
	env.Plan = &PagePlan{Path: "wiki/index.md", BaseAbsent: true, StagedHash: "abc123", Diff: "--- /dev/null\n"}

	var buf bytes.Buffer
	if err := env.WriteJSON(&buf); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}

	var generic map[string]json.RawMessage
	if err := json.Unmarshal(buf.Bytes(), &generic); err != nil {
		t.Fatalf("emitted JSON does not parse: %v\n%s", err, buf.String())
	}
	for _, field := range []string{
		"contractVersion", "operation", "status",
		"findings", "affectedPaths", "approval", "plan",
	} {
		if _, ok := generic[field]; !ok {
			t.Errorf("page plan envelope JSON is missing field %q; got: %s", field, buf.String())
		}
	}
	if _, ok := generic["page"]; ok {
		t.Errorf("page plan envelope must omit the page field; got: %s", buf.String())
	}
	if len(generic) != 7 {
		t.Errorf("page plan envelope must carry exactly 7 fields, got %d: %s", len(generic), buf.String())
	}
}

func TestNewEnvelopeDefaults(t *testing.T) {
	env := New("validate", StatusSuccess)

	if env.ContractVersion != ContractVersion {
		t.Errorf("ContractVersion = %q, want %q", env.ContractVersion, ContractVersion)
	}
	if ContractVersion != "v1" {
		t.Errorf("ContractVersion constant = %q, want v1 (ADR-003 starts at v1)", ContractVersion)
	}
	if env.Operation != "validate" {
		t.Errorf("Operation = %q, want validate", env.Operation)
	}
	if env.Status != StatusSuccess {
		t.Errorf("Status = %q, want %q", env.Status, StatusSuccess)
	}
	if env.Findings == nil {
		t.Error("Findings must be a non-nil empty slice so it serializes as [], not null")
	}
	if env.AffectedPaths == nil {
		t.Error("AffectedPaths must be a non-nil empty slice so it serializes as [], not null")
	}
	if env.Approval != nil {
		t.Errorf("Approval must default to nil, got %+v", env.Approval)
	}
}

// Empty collections must serialize as [] rather than null so callers can parse
// them uniformly across every surface.
func TestEmptyCollectionsSerializeAsArrays(t *testing.T) {
	env := New("validate", StatusSuccess)

	var buf bytes.Buffer
	if err := env.WriteJSON(&buf); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	out := buf.String()

	if !strings.Contains(out, `"findings": []`) {
		t.Errorf("findings should serialize as []; got: %s", out)
	}
	if !strings.Contains(out, `"affectedPaths": []`) {
		t.Errorf("affectedPaths should serialize as []; got: %s", out)
	}
	if !strings.Contains(out, `"approval": null`) {
		t.Errorf("approval should serialize as null when absent; got: %s", out)
	}
	if !strings.HasSuffix(out, "\n") {
		t.Error("WriteJSON output should end with a trailing newline")
	}
}

// A fully-populated envelope must round-trip through JSON without loss.
func TestEnvelopeJSONRoundTrip(t *testing.T) {
	original := New("validate", StatusValidationFailure)
	original.Findings = []Finding{
		{
			Ruleset:  RulesetOKF,
			Severity: SeverityError,
			Code:     "okf.frontmatter.invalid-yaml",
			Message:  "frontmatter is not valid YAML",
			Path:     "pages/example.md",
		},
		{
			Ruleset:  RulesetProfile,
			Severity: SeverityWarning,
			Code:     "profile.link.broken",
			Message:  "link target does not exist",
			Path:     "pages/other.md",
		},
	}
	original.AffectedPaths = []string{"pages/example.md"}
	original.Approval = &Approval{
		Required: true,
		Reason:   "mutation touches tracked files",
		Paths:    []string{"pages/example.md"},
	}

	var buf bytes.Buffer
	if err := original.WriteJSON(&buf); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}

	var decoded Envelope
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("round-trip decode: %v", err)
	}

	roundTripped, err := json.Marshal(&decoded)
	if err != nil {
		t.Fatalf("re-marshal: %v", err)
	}
	originalCompact, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal original: %v", err)
	}
	if !bytes.Equal(roundTripped, originalCompact) {
		t.Errorf("round-trip mismatch:\n original: %s\n  decoded: %s", originalCompact, roundTripped)
	}
}
