package processor

import (
	"strings"
	"testing"

	approvals "github.com/approvals/go-approval-tests"
	"github.com/mpv/kir/processor"
)

func TestProcessor(t *testing.T) {

	cases := []struct {
		name string
		file string
	}{
		{
			name: "CronJob",
			file: "processor_test.TestProcessor.CronJob.input.yaml",
		},
		{
			name: "DaemonSet",
			file: "processor_test.TestProcessor.DaemonSet.input.yaml",
		},
		{
			name: "Deployment",
			file: "processor_test.TestProcessor.Deployment.input.yaml",
		},
		{
			name: "Job",
			file: "processor_test.TestProcessor.Job.input.yaml",
		},
		{
			name: "ReplicaSet",
			file: "processor_test.TestProcessor.ReplicaSet.input.yaml",
		},
		{
			name: "StatefulSet",
			file: "processor_test.TestProcessor.StatefulSet.input.yaml",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			images, err := processor.ProcessFile(tc.file)
			if err != nil {
				t.Fatalf("ProcessFile() error = %v", err)
			}
			imagesAsString := strings.Join(images, "\n")
			approvals.VerifyString(t, imagesAsString)
		})
	}
}
