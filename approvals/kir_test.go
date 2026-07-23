package processor

import (
	"strings"
	"testing"

	approvals "github.com/approvals/go-approval-tests"
	"github.com/mpv/kir/processor"
)

// verify processes file and approves what the tool would print to stdout (the
// images) and to stderr (the error message, if any) as two separate golden
// files. Expected behavior lives entirely in the goldens: a change to either
// stream — an image appearing or disappearing, or an error starting or
// stopping — surfaces as a reviewable diff.
func verify(t *testing.T, file string) {
	t.Helper()

	images, err := processor.ProcessFile(file)

	stdout := strings.Join(images, "\n")
	approvals.VerifyString(t, stdout, approvals.Options().ForFile().WithAdditionalInformation("stdout"))

	stderr := ""
	if err != nil {
		stderr = err.Error()
	}
	approvals.VerifyString(t, stderr, approvals.Options().ForFile().WithAdditionalInformation("stderr"))
}

func TestKind(t *testing.T) {
	kinds := []string{"Pod", "CronJob", "DaemonSet", "Deployment", "Job", "ReplicaSet", "StatefulSet"}

	for _, kind := range kinds {
		t.Run(kind, func(t *testing.T) {
			verify(t, "kir_test.TestKind."+kind+".input.yaml")
		})
	}
}

// Non-workload documents (e.g. a Service) are skipped rather than treated as
// errors. Through verify() this shows up as the stderr golden going from the
// old "unsupported kind Service" error to empty.
func TestSkipsNonWorkloads(t *testing.T) {
	t.Run("Service", func(t *testing.T) {
		verify(t, "kir_test.TestSkipsNonWorkloads.Service.input.yaml")
	})
}

// A file that mixes supported workloads with a non-workload document yields the
// images from the workloads; the non-workload document is skipped without
// discarding the rest.
func TestMixed(t *testing.T) {
	verify(t, "kir_test.TestMixed.WorkloadAndService.input.yaml")
}

func TestMultiple(t *testing.T) {
	verify(t, "kir_test.TestMultiple.input.yaml")
}
