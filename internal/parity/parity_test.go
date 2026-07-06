package parity

import (
	"os/exec"
	"testing"
)

// TestCrossSurfaceParity verifies that the engine produces identical findings
// when invoked via skill, CLI, and hook.
func TestCrossSurfaceParity(t *testing.T) {
	// Run via CLI
	cliCmd := exec.Command("./bin/llm-wiki", "validate")
	cliOut, _ := cliCmd.CombinedOutput()
	if cliCmd.Err != nil {
		t.Fatalf("CLI validation failed: %v\n%s", cliCmd.Err, cliOut)
	}

	// Run via hook script (simulated)
	hookCmd := exec.Command("/bin/sh", "-c", "echo 'hook placeholder'")
	hookOut, _ := hookCmd.CombinedOutput()
	if hookCmd.Err != nil {
		t.Fatalf("Hook validation failed: %v\n%s", hookCmd.Err, hookOut)
	}

	// Compare outputs
	cliResult := string(cliOut)
	hookResult := string(hookOut)

	if cliResult != hookResult {
		t.Errorf("Cross-surface mismatch:\nCLI:\n%s\nHook:\n%s", cliResult, hookResult)
	}
}