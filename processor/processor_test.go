package processor

import (
	"os"
	"testing"
)

func TestProcessFile(t *testing.T) {
	dir := setupTestDir(t)
	defer os.RemoveAll(dir)

	createTestFile(t, dir, "test.yaml", `
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
  containers:
  - name: test-container
    image: test-image
---
apiVersion: v1
kind: Pod
metadata:
  name: another-pod
spec:
  containers:
  - name: another-container
    image: another-image
`)

	images, err := ProcessFile("test.yaml")
	if err != nil {
		t.Fatalf("ProcessFile() error = %v", err)
	}

	expected := []string{"test-image", "another-image"}
	if len(images) != len(expected) {
		t.Fatalf("expected %d images, got %d", len(expected), len(images))
	}

	for i, img := range images {
		if img != expected[i] {
			t.Errorf("expected image %q, got %q", expected[i], img)
		}
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

func createTestFile(t *testing.T, dir, name, content string) {
	err := os.WriteFile(dir+"/"+name, []byte(content), 0644)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
}
