package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
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

func TestRunVersion(t *testing.T) {
	for _, arg := range []string{"--version", "-v"} {
		t.Run(arg, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Run([]string{arg}, nil, &stdout, &stderr)

			if code != 0 {
				t.Errorf("exit code = %d, want 0", code)
			}
			if got := stdout.String(); !strings.HasPrefix(got, "kir ") {
				t.Errorf("stdout = %q, want prefix %q", got, "kir ")
			}
			if stderr.Len() != 0 {
				t.Errorf("stderr = %q, want empty", stderr.String())
			}
		})
	}
}

func TestRunFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")
	if err := os.WriteFile(path, []byte(podManifest), 0644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	code := Run([]string{path}, nil, &stdout, &stderr)

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if got, want := stdout.String(), "test-image\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunStdin(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"-"}, strings.NewReader(podManifest), &stdout, &stderr)

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if got, want := stdout.String(), "test-image\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}

// A file that cannot be processed makes Run exit non-zero, while the images
// from the files that succeeded are still written.
func TestRunFileFailure(t *testing.T) {
	dir := t.TempDir()
	good := filepath.Join(dir, "good.yaml")
	bad := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(good, []byte(podManifest), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bad, []byte("this: is: not: valid: yaml:\n  - [unclosed\n"), 0644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	code := Run([]string{good, bad}, nil, &stdout, &stderr)

	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if got, want := stdout.String(), "test-image\n"; got != want {
		t.Errorf("stdout = %q, want the good file's image %q", got, want)
	}
}

// Which values are unreportable is imageref's business; this pins the CLI
// contract around them. The hostile bytes are written here rather than in an
// approvals fixture, which tooling would normalise (see AGENTS.md).
const hostileImageManifest = `
apiVersion: v1
kind: Pod
metadata:
  name: hostile
spec:
  containers:
  - name: reportable
    image: registry.k8s.io/nginx-slim:0.8
  - name: forges-a-second-entry
    image: "evil\nsecond-line"
  - name: repaints-the-terminal
    image: "nginx:1.0\x1b[2K\rregistry.io/trusted:safe"
`

func TestRunRejectsUnreportableImages(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"-"}, strings.NewReader(hostileImageManifest), &stdout, &stderr)

	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if got, want := stdout.String(), "registry.k8s.io/nginx-slim:0.8\n"; got != want {
		t.Errorf("stdout = %q, want only the reportable image %q", got, want)
	}
	if got, want := strings.Count(stderr.String(), "error:"), 2; got != want {
		t.Errorf("stderr reported %d errors, want %d:\n%s", got, want, stderr.String())
	}
	if strings.ContainsRune(stderr.String(), '\x1b') {
		t.Errorf("stderr contains a raw escape byte, want it quoted:\n%q", stderr.String())
	}
}

func TestRunNoArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run(nil, nil, &stdout, &stderr)

	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "Usage") {
		t.Errorf("stderr = %q, want it to mention usage", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want empty", stdout.String())
	}
}
