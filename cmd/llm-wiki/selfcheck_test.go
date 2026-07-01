package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/olivermorgan2/llm-wiki-kit/internal/contract"
	"github.com/olivermorgan2/llm-wiki-kit/internal/platform"
)

// writeBundle builds a temp bundle root containing the host's artifact plus a
// matching bin/SHA256SUMS, then applies an optional corruption step. It returns
// the root and the slash-separated artifact path.
func writeBundle(t *testing.T, corrupt func(root, artifact string)) (root, artifact string) {
	t.Helper()
	root = t.TempDir()
	host, err := platform.Detect()
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	artifact = host.ArtifactPath()

	full := filepath.Join(root, filepath.FromSlash(artifact))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	const payload = "engine payload"
	if err := os.WriteFile(full, []byte(payload), 0o755); err != nil {
		t.Fatal(err)
	}

	sum, err := platform.Sum(strings.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := platform.WriteManifest(&buf, platform.Manifest{artifact: sum}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "bin", platform.ManifestName), buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}

	if corrupt != nil {
		corrupt(root, artifact)
	}
	return root, artifact
}

func TestSelfcheckSuccessJSON(t *testing.T) {
	root, _ := writeBundle(t, nil)
	stdout, _, code := exec(t, "selfcheck", "--root", root, "--json")

	env := decodeEnvelope(t, stdout)
	if env.Operation != "selfcheck" {
		t.Errorf("operation = %q, want selfcheck", env.Operation)
	}
	if env.Status != contract.StatusSuccess {
		t.Errorf("status = %q, want success", env.Status)
	}
	if len(env.Findings) != 0 {
		t.Errorf("selfcheck reports outcome via status, not findings; got %v", env.Findings)
	}
	if code != int(contract.ExitSuccess) {
		t.Errorf("exit code = %d, want %d", code, int(contract.ExitSuccess))
	}
}

func TestSelfcheckChecksumMismatchJSON(t *testing.T) {
	root, _ := writeBundle(t, func(root, artifact string) {
		full := filepath.Join(root, filepath.FromSlash(artifact))
		if err := os.WriteFile(full, []byte("tampered"), 0o755); err != nil {
			t.Fatal(err)
		}
	})
	stdout, _, code := exec(t, "selfcheck", "--root", root, "--json")

	env := decodeEnvelope(t, stdout)
	if env.Status != contract.StatusSystemFailure {
		t.Errorf("status = %q, want %q", env.Status, contract.StatusSystemFailure)
	}
	if code != int(contract.ExitSystemFailure) {
		t.Errorf("exit code = %d, want %d", code, int(contract.ExitSystemFailure))
	}
}

func TestSelfcheckMissingArtifactJSON(t *testing.T) {
	root, _ := writeBundle(t, func(root, artifact string) {
		if err := os.Remove(filepath.Join(root, filepath.FromSlash(artifact))); err != nil {
			t.Fatal(err)
		}
	})
	stdout, _, code := exec(t, "selfcheck", "--root", root, "--json")

	env := decodeEnvelope(t, stdout)
	if env.Status != contract.StatusSystemFailure {
		t.Errorf("status = %q, want %q", env.Status, contract.StatusSystemFailure)
	}
	if code != int(contract.ExitSystemFailure) {
		t.Errorf("exit code = %d, want %d", code, int(contract.ExitSystemFailure))
	}
}

func TestSelfcheckHumanSuccess(t *testing.T) {
	root, _ := writeBundle(t, nil)
	stdout, stderr, code := exec(t, "selfcheck", "--root", root)

	if code != int(contract.ExitSuccess) {
		t.Errorf("exit code = %d, want %d", code, int(contract.ExitSuccess))
	}
	if strings.Contains(stdout, "{") {
		t.Errorf("default output must be human-readable, not JSON: %s", stdout)
	}
	if !strings.Contains(stdout, "selfcheck") {
		t.Errorf("success output should mention selfcheck: %s", stdout)
	}
	if stderr != "" {
		t.Errorf("no stderr expected on success, got: %s", stderr)
	}
}

func TestSelfcheckHumanFailureWritesStderr(t *testing.T) {
	root, _ := writeBundle(t, func(root, artifact string) {
		if err := os.Remove(filepath.Join(root, filepath.FromSlash(artifact))); err != nil {
			t.Fatal(err)
		}
	})
	stdout, stderr, code := exec(t, "selfcheck", "--root", root)

	if code != int(contract.ExitSystemFailure) {
		t.Errorf("exit code = %d, want %d", code, int(contract.ExitSystemFailure))
	}
	if stderr == "" {
		t.Error("a failed selfcheck should explain the reason on stderr")
	}
	if strings.Contains(stdout, "{") {
		t.Errorf("human-mode failure must not print an envelope to stdout: %s", stdout)
	}
}

// The --root=<dir> form is honored just like the space-separated form.
func TestSelfcheckRootEqualsForm(t *testing.T) {
	root, _ := writeBundle(t, nil)
	_, _, code := exec(t, "selfcheck", "--root="+root, "--json")
	if code != int(contract.ExitSuccess) {
		t.Errorf("exit code = %d, want %d", code, int(contract.ExitSuccess))
	}
}
