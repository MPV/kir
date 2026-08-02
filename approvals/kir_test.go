package processor

import (
	"bytes"
	"io"
	"os"
	"strconv"
	"testing"

	approvals "github.com/approvals/go-approval-tests"
	"github.com/mpv/kir/cmd"
)

// verify runs the CLI in-process through cmd.Run — the real entry point — for
// the given args and stdin, and approves the (stdout, stderr, exit code) triple
// as three golden files. Every scenario goes through this one seam, so the
// goldens capture exactly what the tool emits. Granular, per-layer coverage
// lives in the k8s/yamlparser/processor/fileutil unit tests; this package is
// the behavioral golden layer.
func verify(t *testing.T, args []string, stdin io.Reader) {
	t.Helper()

	var stdout, stderr bytes.Buffer
	code := cmd.Run(args, stdin, &stdout, &stderr)

	approvals.VerifyString(t, stdout.String(), approvals.Options().ForFile().WithAdditionalInformation("stdout"))
	approvals.VerifyString(t, stderr.String(), approvals.Options().ForFile().WithAdditionalInformation("stderr"))
	approvals.VerifyString(t, strconv.Itoa(code), approvals.Options().ForFile().WithAdditionalInformation("exitcode"))
}

func TestKind(t *testing.T) {
	kinds := []string{"Pod", "CronJob", "DaemonSet", "Deployment", "Job", "ReplicaSet", "StatefulSet"}

	for _, kind := range kinds {
		t.Run(kind, func(t *testing.T) {
			verify(t, []string{"kir_test.TestKind." + kind + ".input.yaml"}, nil)
		})
	}
}

func TestError(t *testing.T) {
	t.Run("Service", func(t *testing.T) {
		verify(t, []string{"kir_test.TestError.Service.input.yaml"}, nil)
	})
}

func TestMultiple(t *testing.T) {
	verify(t, []string{"kir_test.TestMultiple.input.yaml"}, nil)
}

// TestCLI covers behaviour that only exists at the CLI boundary — stdin wiring,
// argument resolution, and no-args usage — which no file-argument scenario
// above reaches.
func TestCLI(t *testing.T) {
	t.Run("Stdin", func(t *testing.T) {
		f, err := os.Open("kir_test.TestCLI.Stdin.input.yaml")
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		verify(t, []string{"-"}, f)
	})

	t.Run("MissingFile", func(t *testing.T) {
		verify(t, []string{"does-not-exist.yaml"}, nil)
	})

	t.Run("Usage", func(t *testing.T) {
		verify(t, nil, nil)
	})
}
