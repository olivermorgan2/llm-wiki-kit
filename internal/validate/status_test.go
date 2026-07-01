package validate

import (
	"testing"

	"github.com/olivermorgan2/llm-wiki-kit/internal/contract"
)

func TestStatusForCleanIsSuccess(t *testing.T) {
	if got := StatusFor(nil); got != contract.StatusSuccess {
		t.Errorf("no findings -> %q, want success", got)
	}
}

func TestStatusForSuggestionsOnlyIsSuccess(t *testing.T) {
	fs := []contract.Finding{recSuggestion("a.md")}
	if got := StatusFor(fs); got != contract.StatusSuccess {
		t.Errorf("suggestions only -> %q, want success (advisory)", got)
	}
}

func TestStatusForWarningIsSuccessWithWarnings(t *testing.T) {
	fs := []contract.Finding{recSuggestion("a.md"), kebabWarn("A.md")}
	if got := StatusFor(fs); got != contract.StatusSuccessWithWarnings {
		t.Errorf("warning present -> %q, want success-with-warnings", got)
	}
}

func TestStatusForErrorIsValidationFailure(t *testing.T) {
	fs := []contract.Finding{kebabWarn("A.md"), missingTitle("a.md")}
	if got := StatusFor(fs); got != contract.StatusValidationFailure {
		t.Errorf("error present -> %q, want validation-failure", got)
	}
}

// The status maps to the exit-code bucket via the shared contract table.
func TestStatusForMapsToExitCodes(t *testing.T) {
	cases := []struct {
		findings []contract.Finding
		code     contract.ExitCode
	}{
		{nil, contract.ExitSuccess},
		{[]contract.Finding{kebabWarn("A.md")}, contract.ExitSuccessWithWarnings},
		{[]contract.Finding{missingTitle("a.md")}, contract.ExitValidationFailure},
	}
	for _, c := range cases {
		if got := contract.ExitCodeForStatus(StatusFor(c.findings)); got != c.code {
			t.Errorf("exit code = %d, want %d for %+v", got, c.code, c.findings)
		}
	}
}
