package k8s

import (
	"slices"
	"testing"

	"sigs.k8s.io/yaml"
)

// decode turns a manifest into the plain Go values FindImages reads.
func decode(t *testing.T, manifest string) map[string]any {
	t.Helper()
	var doc map[string]any
	if err := yaml.Unmarshal([]byte(manifest), &doc); err != nil {
		t.Fatalf("decoding fixture: %v", err)
	}
	return doc
}

func TestDefaultConfigParses(t *testing.T) {
	config := DefaultConfig()
	if len(config.Kinds()) == 0 {
		t.Fatal("embedded config describes no kinds")
	}
}

// Each configured kind is reached by following its declared path.
func TestFindImagesConfiguredKinds(t *testing.T) {
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
			name: "List (documents path expands items)",
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
			name: "a configured kind whose path does not apply",
			manifest: `
kind: Pod
status:
  containerStatuses:
  - name: app
    image: app:1
`,
			want: nil,
		},
		{
			name: "an unconfigured kind",
			manifest: `
kind: Service
spec:
  ports:
  - port: 80
`,
			want: nil,
		},
	}

	config := DefaultConfig()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := config.FindImages(decode(t, tt.manifest))
			if !slices.Equal(got, tt.want) {
				t.Errorf("FindImages() = %v, want %v", got, tt.want)
			}
		})
	}
}

// A custom resource is invisible until it is described, and then it is read
// exactly like a built-in. This is the whole trade of this approach.
func TestFindImagesCustomResourceNeedsConfig(t *testing.T) {
	rollout := decode(t, `
apiVersion: argoproj.io/v1alpha1
kind: Rollout
spec:
  template:
    spec:
      containers:
      - name: app
        image: app:1.4.2
`)

	if got := DefaultConfig().FindImages(rollout); len(got) != 0 {
		t.Errorf("undescribed Rollout = %v, want no images", got)
	}

	extra, err := LoadConfig([]byte(`
resources:
  - kind: Rollout
    podSpecs: [spec.template.spec]
`))
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	want := []string{"app:1.4.2"}
	if got := DefaultConfig().Merge(extra).FindImages(rollout); !slices.Equal(got, want) {
		t.Errorf("described Rollout = %v, want %v", got, want)
	}
}

// Merging replaces a kind's paths, so a built-in entry can be corrected and not
// merely extended.
func TestMergeReplacesPaths(t *testing.T) {
	pod := decode(t, `
kind: Pod
elsewhere:
  containers:
  - name: app
    image: app:1
`)

	extra, err := LoadConfig([]byte(`
resources:
  - kind: Pod
    podSpecs: [elsewhere]
`))
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	want := []string{"app:1"}
	if got := DefaultConfig().Merge(extra).FindImages(pod); !slices.Equal(got, want) {
		t.Errorf("FindImages() = %v, want %v", got, want)
	}
}

// A path that runs off the end of the document yields nothing rather than
// panicking, which is the normal case for an optional field.
func TestResolveMissingPath(t *testing.T) {
	doc := decode(t, `
kind: Deployment
spec: {}
`)
	if got := DefaultConfig().FindImages(doc); len(got) != 0 {
		t.Errorf("FindImages() = %v, want no images", got)
	}
}

func TestLoadConfigRejectsUnknownFields(t *testing.T) {
	if _, err := LoadConfig([]byte("resources:\n  - kind: Pod\n    podSpec: [spec]\n")); err == nil {
		t.Error("LoadConfig() accepted an unknown field, want an error")
	}
}
