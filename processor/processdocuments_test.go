package processor

import "testing"

// processDocuments backs both ProcessFile and ProcessStdin. Previously stdin
// only ever processed the first document; this verifies the shared path
// collects images from every document in a multi-document stream.
func TestProcessDocumentsMultiple(t *testing.T) {
	data := []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: first
spec:
  containers:
  - name: c
    image: image-one
---
apiVersion: v1
kind: Pod
metadata:
  name: second
spec:
  containers:
  - name: c
    image: image-two
`)

	images, err := processDocuments(data)
	if err != nil {
		t.Fatalf("processDocuments() error = %v", err)
	}

	expected := []string{"image-one", "image-two"}
	if len(images) != len(expected) {
		t.Fatalf("expected %d images, got %d: %v", len(expected), len(images), images)
	}
	for i, img := range images {
		if img != expected[i] {
			t.Errorf("expected image %q, got %q", expected[i], img)
		}
	}
}
