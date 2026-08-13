package k8s

import (
	"slices"
	"testing"

	"sigs.k8s.io/yaml"
)

// matcher compiles the embedded CUE schema once for the tests.
func matcher(t *testing.T) *Matcher {
	t.Helper()
	m, err := NewMatcher("")
	if err != nil {
		t.Fatalf("compiling embedded schema: %v", err)
	}
	return m
}

// decode turns a manifest fragment into the plain Go values FindImages walks.
func decode(t *testing.T, manifest string) any {
	t.Helper()
	var doc any
	if err := yaml.Unmarshal([]byte(manifest), &doc); err != nil {
		t.Fatalf("decoding fixture: %v", err)
	}
	return doc
}

// The kinds kir has always supported are found without being named: each is
// just a PodSpec at a different depth.
func TestFindImagesWorkloadKinds(t *testing.T) {
	tests := []struct {
		name     string
		manifest string
		want     []string
	}{
		{
			name: "Pod (PodSpec directly on spec)",
			manifest: `
kind: Pod
spec:
  containers:
  - name: app
    image: app:1
`,
			want: []string{"app:1"},
		},
		{
			name: "Deployment (PodSpec under a template)",
			manifest: `
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: app
        image: app:2
`,
			want: []string{"app:2"},
		},
		{
			name: "CronJob (PodSpec four levels down)",
			manifest: `
kind: CronJob
spec:
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: app
            image: app:3
`,
			want: []string{"app:3"},
		},
		{
			name: "List (every item contributes)",
			manifest: `
kind: List
items:
- kind: Pod
  spec:
    containers:
    - name: a
      image: a:1
- kind: Service
  spec:
    ports:
    - port: 80
- kind: Pod
  spec:
    containers:
    - name: b
      image: b:1
`,
			want: []string{"a:1", "b:1"},
		},
		{
			name: "all three container fields, in report order",
			manifest: `
kind: Pod
spec:
  containers:
  - name: app
    image: app:1
  initContainers:
  - name: init
    image: init:1
  ephemeralContainers:
  - name: debugger
    image: debug:1
    targetContainerName: app
`,
			want: []string{"app:1", "init:1", "debug:1"},
		},
		{
			name: "no PodSpec anywhere",
			manifest: `
kind: Service
spec:
  ports:
  - port: 80
`,
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matcher(t).FindImages(decode(t, tt.manifest))
			if !slices.Equal(got, tt.want) {
				t.Errorf("FindImages() = %v, want %v", got, tt.want)
			}
		})
	}
}

// The point of structural discovery: a custom resource the Kubernetes scheme
// cannot decode is understood, because its PodSpec is a PodSpec.
func TestFindImagesCustomResource(t *testing.T) {
	rollout := `
apiVersion: argoproj.io/v1alpha1
kind: Rollout
spec:
  strategy:
    canary:
      steps:
      - setWeight: 20
  template:
    spec:
      containers:
      - name: app
        image: app:1.4.2
`

	want := []string{"app:1.4.2"}
	if got := matcher(t).FindImages(decode(t, rollout)); !slices.Equal(got, want) {
		t.Errorf("FindImages() = %v, want %v", got, want)
	}
}

// The counterweight: matching on shape must not mean matching on a field name.
// Decoding into corev1.Container is what separates a PodSpec from a lookalike.
func TestFindImagesRejectsLookalikes(t *testing.T) {
	tests := []struct {
		name     string
		manifest string
	}{
		{
			name: "containers of another kind entirely",
			manifest: `
kind: ShippingManifest
spec:
  containers:
  - name: cargo-hold-1
    capacity: 40ft
    image: photo-of-container.jpg
`,
		},
		{
			name: "containers holding strings",
			manifest: `
kind: Warehouse
spec:
  containers:
  - CONTAINER-A
  - CONTAINER-B
`,
		},
		{
			name: "a manifest embedded as a string",
			manifest: `
kind: ConfigMap
data:
  pod.yaml: |
    kind: Pod
    spec:
      containers:
      - name: inner
        image: inner:1
`,
		},
		{
			name: "container status, which reports images but is not a PodSpec",
			manifest: `
kind: Pod
status:
  containerStatuses:
  - name: app
    image: app:1
    imageID: docker-pullable://app@sha256:abc
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matcher(t).FindImages(decode(t, tt.manifest)); len(got) != 0 {
				t.Errorf("FindImages() = %v, want no images", got)
			}
		})
	}
}

// Go randomises map iteration, so the walk sorts keys. Without that the golden
// files would flake whenever a document holds more than one PodSpec.
func TestFindImagesOrderIsStable(t *testing.T) {
	manifest := `
kind: List
items:
- kind: Pod
  spec:
    containers:
    - name: a
      image: a:1
- kind: Pod
  spec:
    containers:
    - name: b
      image: b:1
- kind: Pod
  spec:
    containers:
    - name: c
      image: c:1
`

	doc := decode(t, manifest)
	m := matcher(t)
	want := m.FindImages(doc)
	if len(want) != 3 {
		t.Fatalf("expected 3 images, got %d", len(want))
	}
	for range 50 {
		if got := m.FindImages(doc); !slices.Equal(got, want) {
			t.Fatalf("FindImages() = %v, want %v — order is not stable", got, want)
		}
	}
}

// Numbers must reach CUE as integers. A manifest is decoded YAML-to-JSON-to-Go,
// which leaves every number a float64; offering that to a schema saying int32
// rejects every container that declares a port — most real ones.
func TestFindImagesIntegerFields(t *testing.T) {
	manifest := `
kind: Pod
spec:
  containers:
  - name: app
    image: app:1
    ports:
    - containerPort: 80
      name: web
`

	want := []string{"app:1"}
	if got := matcher(t).FindImages(decode(t, manifest)); !slices.Equal(got, want) {
		t.Errorf("FindImages() = %v, want %v", got, want)
	}
}
