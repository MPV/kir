package k8s

import (
	"slices"
	"strings"
	"testing"

	"sigs.k8s.io/yaml"
)

// decodeDoc turns a manifest into a whole document, which is what Config reads.
// The walk's own helper yields `any`, since it descends into fragments too.
func decodeDoc(t *testing.T, manifest string) map[string]any {
	t.Helper()
	var doc map[string]any
	if err := yaml.Unmarshal([]byte(manifest), &doc); err != nil {
		t.Fatalf("decoding fixture: %v", err)
	}
	return doc
}

func TestDefaultConfigLoads(t *testing.T) {
	if len(DefaultConfig().Kinds()) == 0 {
		t.Fatal("embedded config describes no kinds")
	}
}

// The built-in entries are an accelerator, not knowledge: they save the walk
// from re-deriving where a Deployment keeps its PodSpec, but they are not the
// reason kir understands one. Deleting them all must change no answer.
//
// This is what stops resources.yaml from quietly becoming the hardcoded kind
// list that #26 set out to remove.
func TestBuiltInConfigIsRedundant(t *testing.T) {
	manifests := map[string]string{
		"Pod": `
kind: Pod
spec:
  containers: [{name: app, image: app:1}]
`,
		"Deployment": `
kind: Deployment
spec:
  template:
    spec:
      containers: [{name: app, image: app:2}]
`,
		"CronJob": `
kind: CronJob
spec:
  jobTemplate:
    spec:
      template:
        spec:
          containers: [{name: app, image: app:3}]
`,
		"PodTemplate": `
kind: PodTemplate
template:
  spec:
    containers: [{name: app, image: app:4}]
`,
		"ReplicationController": `
kind: ReplicationController
spec:
  template:
    spec:
      containers: [{name: app, image: app:5}]
`,
		"List": `
kind: List
items:
- kind: Pod
  spec:
    containers: [{name: a, image: a:1}]
- kind: Deployment
  spec:
    template:
      spec:
        containers: [{name: b, image: b:1}]
`,
	}

	configured := DefaultConfig()
	empty, err := LoadConfig([]byte("resources: []\n"))
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	for kind, manifest := range manifests {
		t.Run(kind, func(t *testing.T) {
			doc := decodeDoc(t, manifest)
			viaConfig := configured.FindImages(doc)
			viaWalk := empty.FindImages(doc)
			if len(viaWalk) == 0 {
				t.Fatalf("walk alone found nothing for %s", kind)
			}
			if !slices.Equal(viaConfig, viaWalk) {
				t.Errorf("configured = %v, inferred = %v — they must agree", viaConfig, viaWalk)
			}
		})
	}
}

// A configured kind is looked up and never also inferred, so an image cannot be
// reported twice for describing something kir would have found anyway.
func TestConfiguredKindIsNotAlsoInferred(t *testing.T) {
	rollout := decodeDoc(t, `
kind: Rollout
spec:
  template:
    spec:
      containers: [{name: app, image: app:1.4.2}]
`)

	inferred := DefaultConfig().FindImages(rollout)
	if want := []string{"app:1.4.2"}; !slices.Equal(inferred, want) {
		t.Fatalf("inferred = %v, want %v", inferred, want)
	}

	extra, err := LoadConfig([]byte("resources:\n  - kind: Rollout\n    podSpecs: [spec.template.spec]\n"))
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if got := DefaultConfig().Merge(extra).FindImages(rollout); !slices.Equal(got, inferred) {
		t.Errorf("configured = %v, inferred = %v — describing a kind must not duplicate it", got, inferred)
	}
}

// An entry with no expressions means "this kind has no images", which is how a
// user overrules the walk. Neither mechanism can do this alone: the walk cannot
// be told to ignore something, and configuration alone has nothing to ignore.
func TestEmptyEntrySilencesAKind(t *testing.T) {
	rollout := decodeDoc(t, `
kind: Rollout
spec:
  template:
    spec:
      containers: [{name: app, image: app:1.4.2}]
`)

	silence, err := LoadConfig([]byte("resources:\n  - kind: Rollout\n    podSpecs: []\n"))
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if got := DefaultConfig().Merge(silence).FindImages(rollout); len(got) != 0 {
		t.Errorf("FindImages() = %v, want no images", got)
	}
}

// The reach configuration adds: a resource holding bare containers, which has
// no PodSpec shape for the walk to match.
func TestConfigReachesBareContainers(t *testing.T) {
	workflow := decodeDoc(t, `
kind: Workflow
spec:
  templates:
  - name: build
    container: {image: builder:1}
  - name: report
    script: {image: python:3.12}
  - name: fanout
    dag: {tasks: [{name: a}]}
`)

	if got := DefaultConfig().FindImages(workflow); len(got) != 0 {
		t.Fatalf("undescribed Workflow = %v, want no images — the walk cannot see bare containers", got)
	}

	extra, err := LoadConfig([]byte("resources:\n  - kind: Workflow\n    containers: [\"spec.templates[*].[container, script][]\"]\n"))
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	want := []string{"builder:1", "python:3.12"}
	if got := DefaultConfig().Merge(extra).FindImages(workflow); !slices.Equal(got, want) {
		t.Errorf("FindImages() = %v, want %v", got, want)
	}
}

// Merging replaces a kind's expressions, so a built-in entry can be corrected
// and not merely extended.
func TestMergeReplacesExpressions(t *testing.T) {
	pod := decodeDoc(t, `
kind: Pod
elsewhere:
  containers: [{name: app, image: app:1}]
`)

	extra, err := LoadConfig([]byte("resources:\n  - kind: Pod\n    podSpecs: [elsewhere]\n"))
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	want := []string{"app:1"}
	if got := DefaultConfig().Merge(extra).FindImages(pod); !slices.Equal(got, want) {
		t.Errorf("FindImages() = %v, want %v", got, want)
	}
}

func TestLoadConfigRejectsUnknownFields(t *testing.T) {
	if _, err := LoadConfig([]byte("resources:\n  - kind: Pod\n    podSpec: [spec]\n")); err == nil {
		t.Error("LoadConfig() accepted an unknown field, want an error")
	}
}

// Expressions are compiled when the config loads, so a typo is an error naming
// the offending kind and field rather than one that silently matches nothing.
func TestLoadConfigRejectsBadExpression(t *testing.T) {
	_, err := LoadConfig([]byte("resources:\n  - kind: Pod\n    podSpecs: [\"spec[\"]\n"))
	if err == nil {
		t.Fatal("LoadConfig() accepted a malformed expression, want an error")
	}
	if !strings.Contains(err.Error(), "Pod.podSpecs") {
		t.Errorf("error = %q, want it to name the offending kind and field", err)
	}
}
