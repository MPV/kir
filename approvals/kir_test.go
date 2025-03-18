package processor

import (
	"strings"
	"testing"

	approvals "github.com/approvals/go-approval-tests"
	"github.com/mpv/kir/processor"
)

func TestKind(t *testing.T) {

	cases := []struct {
		name string
		file string
	}{
		{
			name: "Pod",
			file: "kir_test.TestKind.Pod.input.yaml",
		},
		{
			name: "CronJob",
			file: "kir_test.TestKind.CronJob.input.yaml",
		},
		{
			name: "DaemonSet",
			file: "kir_test.TestKind.DaemonSet.input.yaml",
		},
		{
			name: "Deployment",
			file: "kir_test.TestKind.Deployment.input.yaml",
		},
		{
			name: "Job",
			file: "kir_test.TestKind.Job.input.yaml",
		},
		{
			name: "ReplicaSet",
			file: "kir_test.TestKind.ReplicaSet.input.yaml",
		},
		{
			name: "StatefulSet",
			file: "kir_test.TestKind.StatefulSet.input.yaml",
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

func TestError(t *testing.T) {

	cases := []struct {
		name string
		file string
	}{
		{
			name: "Service",
			file: "kir_test.TestError.Service.input.yaml",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			images, err := processor.ProcessFile(tc.file)
			if err == nil {
				t.Fatalf("ProcessFile() error = %v", err)
			}
			imagesAsString := strings.Join(images, "\n")
			approvals.VerifyString(t, imagesAsString)
		})
	}
}
