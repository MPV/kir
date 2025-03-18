package processor

import (
	"strings"
	"testing"

	approvals "github.com/approvals/go-approval-tests"
	"github.com/mpv/kir/processor"
)

func TestKir(t *testing.T) {

	cases := []struct {
		name string
		file string
	}{
		{
			name: "CronJob",
			file: "kir_test.TestKir.CronJob.input.yaml",
		},
		{
			name: "DaemonSet",
			file: "kir_test.TestKir.DaemonSet.input.yaml",
		},
		{
			name: "Deployment",
			file: "kir_test.TestKir.Deployment.input.yaml",
		},
		{
			name: "Job",
			file: "kir_test.TestKir.Job.input.yaml",
		},
		{
			name: "ReplicaSet",
			file: "kir_test.TestKir.ReplicaSet.input.yaml",
		},
		{
			name: "StatefulSet",
			file: "kir_test.TestKir.StatefulSet.input.yaml",
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
