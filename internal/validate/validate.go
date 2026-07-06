package validate

import (
    "bytes"
    "encoding/json"
    "os/exec"
    "testing"
)

type Result struct {
    Findings []string               `json:"findings"`
    Metrics  map[string]interface{} `json:"metrics,omitempty"`
}

// execCommand runs a command and returns stdout as a string and any error.
func execCommand(name string, args ...string) (string, error) {
    cmd := exec.Command(name, args...)
    var out bytes.Buffer
    cmd.Stdout = &out
    cmd.Stderr = &out
    if err := cmd.Run(); err != nil {
        return out.String(), err
    }
    return out.String(), nil
}

// runSkill invokes the Hermes skill runner for the validation skill.
func runSkill() (*Result, error) {
    // Assuming the skill is named "llm-wiki-validate"
    out, err := execCommand("hermes", "skill-run", "llm-wiki-validate")
    if err != nil {
        return nil, err
    }
    var r Result
    if jsonErr := json.Unmarshal([]byte(out), &r); jsonErr != nil {
        // Fallback: treat raw output as a single finding
        r.Findings = []string{out}
    }
    return &r, nil
}

// runCLI invokes the compiled binary directly.
func runCLI() (*Result, error) {
    out, err := execCommand("./bin/llm-wiki", "validate")
    if err != nil {
        return nil, err
    }
    var r Result
    if jsonErr := json.Unmarshal([]byte(out), &r); jsonErr != nil {
        r.Findings = []string{out}
    }
    return &r, nil
}

// runHook simulates the pre‑commit hook script.
func runHook() (*Result, error) {
    // The hook script resides at .github/hooks/pre-commit.sh
    out, err := execCommand("/bin/sh", ".github/hooks/pre-commit.sh")
    if err != nil {
        return nil, err
    }
    var r Result
    if jsonErr := json.Unmarshal([]byte(out), &r); jsonErr != nil {
        r.Findings = []string{out}
    }
    return &r, nil
}

// TestCrossSurfaceParity validates that all three entry points produce identical findings.
func TestCrossSurfaceParity(t *testing.T) {
    modes := []struct {
        name string
        fn   func() (*Result, error)
    }{
        {"skill", runSkill},
        {"cli", runCLI},
        {"hook", runHook},
    }

    var reference *Result
    for _, m := range modes {
        res, err := m.fn()
        if err != nil {
            t.Fatalf("%s execution failed: %v", m.name, err)
        }
        if reference == nil {
            reference = res
            continue
        }
        // Compare findings slice
        if len(res.Findings) != len(reference.Findings) {
            t.Fatalf("%s findings count mismatch: %d vs %d", m.name, len(res.Findings), len(reference.Findings))
        }
        for i, f := range res.Findings {
            if f != reference.Findings[i] {
                t.Fatalf("%s finding %d differs. got %q expected %q", m.name, i, f, reference.Findings[i])
            }
        }
        // Compare metrics (non‑deterministic values such as timings are ignored)
        // Only compare keys that exist in both maps.
        for k, v := range res.Metrics {
            if refV, ok := reference.Metrics[k]; ok && v != refV {
                t.Fatalf("%s metric %s differs: %v vs %v", m.name, k, v, refV)
            }
        }
    }
}
