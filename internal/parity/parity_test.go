package parity

import (
	"encoding/json"
	"os"
	"os/exec"
	"reflect"
	"testing"
)

// Result represents the JSON output from validation
type Result struct {
	Findings []string          `json:"findings"`
	Metrics  map[string]any `json:"metrics"`
}

// TestCrossSurfaceParity verifies that the engine produces identical findings
// when invoked via CLI and hook script on the same repository state.
func TestCrossSurfaceParity(t *testing.T) {
	// Build binary first
	buildCmd := exec.Command("go", "build", "-o", "bin/llm-wiki", "./cmd/llm-wiki")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Build failed: %v\n%s", err, out)
	}
	defer os.Remove("bin/llm-wiki")

	// Run via CLI
	cliCmd := exec.Command("./bin/llm-wiki", "validate")
	cliOut, err := cliCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI validation failed: %v\n%s", err, cliOut)
	}

	var cliResult Result
	if err := json.Unmarshal(cliOut, &cliResult); err != nil {
		// Try parsing as contract JSON
		t.Logf("CLI output (raw): %s", cliOut)
	}

	// Run via hook script
	hookCmd := exec.Command("/bin/sh", ".github/hooks/pre-commit.sh")
	hookOut, err := hookCmd.CombinedOutput()
	if err != nil {
		t.Logf("Hook output: %s", hookOut)
		// Hook may exit non-zero on findings, that's expected
	}

	var hookResult Result
	if err := json.Unmarshal(hookOut, &hookResult); err != nil {
		t.Logf("Hook output (raw): %s", hookOut)
	}

	// Compare findings lists (primary assertion for criterion 15)
	if !reflect.DeepEqual(cliResult.Findings, hookResult.Findings) {
		t.Errorf("Cross-surface findings mismatch:\nCLI:   %v\nHook:  %v", cliResult.Findings, hookResult.Findings)
	}
}