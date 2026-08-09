package processor

import (
	"bytes"
	"io"
	"os"
	"strconv"
	"strings"
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

	approvals.VerifyString(t, newlineTerminated(stdout.String()), approvals.Options().ForFile().WithAdditionalInformation("stdout"))
	approvals.VerifyString(t, newlineTerminated(stderr.String()), approvals.Options().ForFile().WithAdditionalInformation("stderr"))
	approvals.VerifyString(t, newlineTerminated(strconv.Itoa(code)), approvals.Options().ForFile().WithAdditionalInformation("exitcode"))
}

// newlineTerminated ends a non-empty string with a newline, mirroring what
// go-approval-tests does to the received output from v1.6.0 on (empty output
// stays empty there too). Normalising here rather than relying on the library
// keeps the goldens byte-identical across that upgrade: every golden is a
// POSIX text file, and the suite passes on either side of it.
func newlineTerminated(s string) string {
	if s == "" || strings.HasSuffix(s, "\n") {
		return s
	}
	return s + "\n"
}

func TestKind(t *testing.T) {
	// PodTemplate and ReplicationController are built-in kinds carrying a
	// PodSpec that the previous fixed kind list omitted; they are entries in
	// the built-in configuration now.
	kinds := []string{"Pod", "CronJob", "DaemonSet", "Deployment", "Job", "PodTemplate", "ReplicaSet", "ReplicationController", "StatefulSet"}

	for _, kind := range kinds {
		t.Run(kind, func(t *testing.T) {
			verify(t, []string{"kir_test.TestKind." + kind + ".input.yaml"}, nil)
		})
	}
}

// A document kir has no configuration for is skipped: no images, no error,
// exit 0. Service is a built-in without a PodSpec; Lookalike is an undescribed
// custom resource, which is skipped for the same reason — being undescribed —
// whether or not it happens to have a field named containers.
func TestSkipsNonWorkloads(t *testing.T) {
	for _, name := range []string{"Service", "Lookalike"} {
		t.Run(name, func(t *testing.T) {
			verify(t, []string{"kir_test.TestSkipsNonWorkloads." + name + ".input.yaml"}, nil)
		})
	}
}

// A custom resource is invisible until described, and read like a built-in once
// it is. Both halves are pinned, because the first is the cost of this approach
// and the second is the benefit.
func TestCustomResource(t *testing.T) {
	input := "kir_test.TestCustomResource.input.yaml"

	t.Run("Undescribed", func(t *testing.T) {
		verify(t, []string{input}, nil)
	})

	t.Run("Configured", func(t *testing.T) {
		verify(t, []string{"--config", "kir_test.TestCustomResource.config.yaml", input}, nil)
	})

	// An Argo Workflow holds bare containers in a list of templates, each of
	// which has a container, a script, or neither. One JMESPath expression
	// covers it — select both shapes across the list, flatten, and templates
	// with neither drop out — which a plain field path could not express.
	t.Run("Workflow", func(t *testing.T) {
		verify(t, []string{
			"--config", "kir_test.TestCustomResource.Workflow.config.yaml",
			"kir_test.TestCustomResource.Workflow.input.yaml",
		}, nil)
	})
}

func TestMultiple(t *testing.T) {
	verify(t, []string{"kir_test.TestMultiple.input.yaml"}, nil)
}

// A file mixing supported workloads with a non-workload document yields the
// workloads' images; the non-workload is skipped without discarding the rest.
func TestMixed(t *testing.T) {
	verify(t, []string{"kir_test.TestMixed.input.yaml"}, nil)
}

// A List — the envelope `kubectl get ... -o yaml` wraps multiple objects in —
// is unwrapped and each item processed on its own. The fixture holds the same
// three objects as TestMixed, so the goldens pin that the List envelope and a
// multi-document file yield identical images: workload items contribute theirs,
// and a non-workload item is skipped without discarding the items around it.
// Nothing else in the suite exercises the List branch or its item handling.
func TestList(t *testing.T) {
	verify(t, []string{"kir_test.TestList.input.yaml"}, nil)
}

// A file that cannot be parsed is reported on stderr and makes kir exit
// non-zero (rather than reporting success), pinning that contract through the
// real CLI seam.
func TestFailure(t *testing.T) {
	t.Run("BadYAML", func(t *testing.T) {
		verify(t, []string{"kir_test.TestFailure.BadYAML.input.yaml"}, nil)
	})

	// One unparseable document does not cost the images in the documents around
	// it: the fixture's first and third workloads are still reported, the bad
	// middle one is named on stderr, and the exit code is non-zero (ADR 0008).
	// BadYAML above cannot pin this — its whole file is unparseable, so it
	// passes whether or not partial results survive.
	t.Run("PartialStream", func(t *testing.T) {
		verify(t, []string{"kir_test.TestFailure.PartialStream.input.yaml"}, nil)
	})
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

	// StdinMultiDoc pins the multi-document stdin contract: every document in a
	// piped stream is processed, not just the first. Without this the single-doc
	// Stdin case above passes regardless, so the regression would go unnoticed.
	t.Run("StdinMultiDoc", func(t *testing.T) {
		f, err := os.Open("kir_test.TestCLI.StdinMultiDoc.input.yaml")
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

	t.Run("Version", func(t *testing.T) {
		verify(t, []string{"--version"}, nil)
	})
}
