package yamlparser

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/mpv/kir/k8s"
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

	images, err := ProcessData(k8s.DefaultMatcher(), []byte(data))
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

	images, err := ProcessData(k8s.DefaultMatcher(), []byte(data))
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

// A document that cannot be processed must not discard the images found in the
// documents around it. Before this was fixed, one unparseable document anywhere
// in a stream returned no images at all — so a single bad object in a
// `kubectl get -A -o yaml` dump silently cost every image in it.
func TestProcessReaderKeepsImagesAroundABadDocument(t *testing.T) {
	stream := strings.Join([]string{
		"apiVersion: v1\nkind: Pod\nspec:\n  containers:\n  - {name: c, image: before-the-break}\n",
		"apiVersion: v1\nkind: Pod\nspec:\n  containers:\n  - {name: c, image: nginx, ports: [8080}\n",
		"apiVersion: v1\nkind: Pod\nspec:\n  containers:\n  - {name: c, image: after-the-break}\n",
	}, "---\n")

	images, err := ProcessReader(k8s.DefaultMatcher(), strings.NewReader(stream))

	if err == nil {
		t.Error("ProcessReader() error = nil, want the bad document reported")
	}
	want := []string{"before-the-break", "after-the-break"}
	if !slices.Equal(images, want) {
		t.Errorf("ProcessReader() images = %v, want %v", images, want)
	}
}

// Every failure in a stream is reported, not just the first, so the exit code
// and stderr account for all of them.
func TestProcessReaderReportsEveryBadDocument(t *testing.T) {
	bad := "apiVersion: v1\nkind: Pod\nspec:\n  containers:\n  - {name: c, image: nginx, ports: [8080}\n"
	stream := strings.Join([]string{bad, bad}, "---\n")

	images, err := ProcessReader(k8s.DefaultMatcher(), strings.NewReader(stream))

	if len(images) != 0 {
		t.Errorf("ProcessReader() images = %v, want none", images)
	}
	var joined interface{ Unwrap() []error }
	if !errors.As(err, &joined) {
		t.Fatalf("ProcessReader() error = %v, want a joined error covering both documents", err)
	}
	if got := len(joined.Unwrap()); got != 2 {
		t.Errorf("ProcessReader() reported %d failures, want 2", got)
	}
}
