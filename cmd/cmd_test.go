package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	approvals "github.com/approvals/go-approval-tests"
)

const podManifest = `
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
  containers:
  - name: test-container
    image: test-image
`

// verifyExecute runs Execute and approves what it writes (stdout) and the
// error it returns (stderr) as separate golden files, so a change to either
// stream in a later PR is a distinct, reviewable diff.
func verifyExecute(t *testing.T, args []string) {
	t.Helper()

	var out bytes.Buffer
	err := Execute(args, &out)

	approvals.VerifyString(t, out.String(), approvals.Options().ForFile().WithAdditionalInformation("stdout"))

	stderr := ""
	if err != nil {
		stderr = err.Error()
	}
	approvals.VerifyString(t, stderr, approvals.Options().ForFile().WithAdditionalInformation("stderr"))
}

func TestExecuteFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "test.yaml"), podManifest)
	t.Chdir(dir)

	verifyExecute(t, []string{"test.yaml"})
}

func TestExecuteStdin(t *testing.T) {
	withStdin(t, podManifest)

	verifyExecute(t, []string{"-"})
}

// One good file alongside one that fails to parse. The good file's images are
// still written; whether the failure surfaces on stderr is what later PRs
// change.
func TestExecuteFileFailure(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "good.yaml"), podManifest)
	writeFile(t, filepath.Join(dir, "bad.yaml"), "this: is: not: valid: yaml:\n  - [unclosed\n")
	t.Chdir(dir)

	verifyExecute(t, []string{"good.yaml", "bad.yaml"})
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
}

// withStdin replaces os.Stdin with a pipe carrying input for the duration of
// the test, restoring it afterwards.
func withStdin(t *testing.T, input string) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("creating pipe: %v", err)
	}
	old := os.Stdin
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = old })

	if _, err := w.WriteString(input); err != nil {
		t.Fatalf("writing to stdin pipe: %v", err)
	}
	w.Close()
}
