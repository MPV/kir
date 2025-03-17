package cmd

import (
	"bytes"
	"log"
	"os"
	"testing"
)

func TestExecuteFile(t *testing.T) {
	dir := setupTestDir(t)
	defer os.RemoveAll(dir)

	createTestFile(dir, "test.yaml", `
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
  containers:
  - name: test-container
    image: test-image
`)

	testCases := []struct {
		args           []string
		expectedOutput string
	}{
		{[]string{"test.yaml"}, "test-image\n"},
	}

	for _, tc := range testCases {
		t.Run(tc.args[0], func(t *testing.T) {
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			os.Args = append([]string{"cmd"}, tc.args...)
			Execute(tc.args)

			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			_, err := buf.ReadFrom(r)
			if err != nil {
				t.Fatalf("Error reading from buffer: %v", err)
			}
			output := buf.String()

			if output != tc.expectedOutput {
				t.Errorf("expected %q, got %q", tc.expectedOutput, output)
			}
		})
	}
}

func TestExecuteStdin(t *testing.T) {
	testCases := []struct {
		args           []string
		input          string
		expectedOutput string
	}{
		{[]string{"-"}, `
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
  containers:
  - name: test-container
    image: test-image
`, "test-image\n"},
	}

	for _, tc := range testCases {
		t.Run(tc.args[0], func(t *testing.T) {
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			oldStdin := os.Stdin
			stdinR, stdinW, _ := os.Pipe()
			os.Stdin = stdinR
			stdinW.Write([]byte(tc.input))
			stdinW.Close()

			os.Args = append([]string{"cmd"}, tc.args...)
			Execute(tc.args)

			w.Close()
			os.Stdout = oldStdout
			os.Stdin = oldStdin

			var buf bytes.Buffer
			_, err := buf.ReadFrom(r)
			if err != nil {
				t.Fatalf("Error reading from buffer: %v", err)
			}
			output := buf.String()

			if output != tc.expectedOutput {
				t.Errorf("expected %q, got %q", tc.expectedOutput, output)
			}
		})
	}
}

func setupTestDir(t *testing.T) string {
	dir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	err = os.Chdir(dir)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	return dir
}

func createTestFile(dir, name, content string) {
	err := os.WriteFile(dir+"/"+name, []byte(content), 0644)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
