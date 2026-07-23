package yamlparser

import (
	"testing"
)

func TestProcessData(t *testing.T) {
	data := `
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
  containers:
  - name: test-container
    image: test-image
`

	images, err := ProcessData([]byte(data))
	if err != nil {
		t.Fatalf("ProcessData() error = %v", err)
	}

	expected := []string{"test-image"}
	if len(images) != len(expected) {
		t.Fatalf("expected %d images, got %d", len(expected), len(images))
	}

	for i, img := range images {
		if img != expected[i] {
			t.Errorf("expected image %q, got %q", expected[i], img)
		}
	}
}

func TestProcessDataEphemeralContainers(t *testing.T) {
	data := `
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
  containers:
  - name: test-container
    image: test-image
  initContainers:
  - name: init-container
    image: init-image
  ephemeralContainers:
  - name: debugger
    image: ephemeral-image
    targetContainerName: test-container
`

	images, err := ProcessData([]byte(data))
	if err != nil {
		t.Fatalf("ProcessData() error = %v", err)
	}

	expected := []string{"test-image", "init-image", "ephemeral-image"}
	if len(images) != len(expected) {
		t.Fatalf("expected %d images, got %d: %v", len(expected), len(images), images)
	}

	for i, img := range images {
		if img != expected[i] {
			t.Errorf("expected image %q, got %q", expected[i], img)
		}
	}
}
