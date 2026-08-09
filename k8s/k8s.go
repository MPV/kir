// Package k8s finds container images in Kubernetes manifests that have been
// decoded into plain Go values. What counts as a PodSpec is defined by a CUE
// schema (podspec.cue) rather than by Go code.
package k8s

import (
	_ "embed"
	"fmt"
	"maps"
	"slices"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
)

//go:embed podspec.cue
var defaultSchema string

// containerFields are the PodSpec fields that carry images, listed in the order
// kir reports them, paired with the schema definition each is matched against.
var containerFields = []struct{ field, definition string }{
	{"containers", "#Containers"},
	{"initContainers", "#Containers"},
	{"ephemeralContainers", "#EphemeralContainers"},
}

// maxDepth bounds the walk. Real manifests nest a handful of levels; the limit
// only guards against absurd input.
const maxDepth = 100

// Matcher recognises PodSpec-shaped nodes using a compiled CUE schema.
type Matcher struct {
	ctx *cue.Context
	// definitions holds the compiled schema for each container list, looked up
	// once so the walk never re-parses CUE.
	definitions map[string]cue.Value
}

// NewMatcher compiles a CUE schema. The schema must define #Containers and
// #EphemeralContainers; see podspec.cue.
func NewMatcher(schema string) (*Matcher, error) {
	ctx := cuecontext.New()

	value := ctx.CompileString(schema)
	if err := value.Err(); err != nil {
		return nil, fmt.Errorf("compiling schema: %w", err)
	}

	definitions := map[string]cue.Value{}
	for _, name := range []string{"#Containers", "#EphemeralContainers"} {
		definition := value.LookupPath(cue.ParsePath(name))
		if err := definition.Err(); err != nil {
			return nil, fmt.Errorf("schema is missing %s: %w", name, err)
		}
		definitions[name] = definition
	}

	return &Matcher{ctx: ctx, definitions: definitions}, nil
}

// DefaultMatcher returns a Matcher over the embedded schema.
func DefaultMatcher() *Matcher {
	matcher, err := NewMatcher(defaultSchema)
	if err != nil {
		// The embedded schema is compiled in tests; a failure here is a bug.
		panic(fmt.Sprintf("embedded schema does not compile: %v", err))
	}
	return matcher
}

// FindImages returns the images of every PodSpec-shaped node reachable from
// doc, a manifest decoded into plain Go values.
//
// Nothing here knows what a Deployment is. The walk descends until the schema
// recognises a node, which is why a custom resource embedding a PodSpec is
// understood on the same footing as a built-in workload — and why a List needs
// no special case, its items being just more nodes.
func (m *Matcher) FindImages(doc any) []string {
	var images []string
	m.find(doc, &images, 0)
	return images
}

func (m *Matcher) find(node any, images *[]string, depth int) {
	if depth > maxDepth {
		return
	}

	switch n := node.(type) {
	case map[string]any:
		if found, ok := m.podSpecImages(n); ok {
			*images = append(*images, found...)
			return // a PodSpec does not contain another PodSpec
		}
		// Sorted, so output order depends on the manifest rather than on Go's
		// randomised map iteration.
		for _, key := range slices.Sorted(maps.Keys(n)) {
			m.find(n[key], images, depth+1)
		}
	case []any:
		for _, item := range n {
			m.find(item, images, depth+1)
		}
	}
}

// podSpecImages reports whether node is PodSpec-shaped, and if so its images.
// A field named containers is not enough: its value has to unify with the
// schema, which is closed against unknown fields.
func (m *Matcher) podSpecImages(node map[string]any) ([]string, bool) {
	var images []string
	matched := false

	for _, cf := range containerFields {
		value, ok := node[cf.field]
		if !ok {
			continue
		}
		containers, ok := value.([]any)
		if !ok || len(containers) == 0 {
			continue
		}
		if !m.unifies(cf.definition, value) {
			continue
		}
		matched = true
		images = append(images, imagesOf(containers)...)
	}

	return images, matched
}

// unifies reports whether value satisfies the named schema definition.
func (m *Matcher) unifies(definition string, value any) bool {
	encoded := m.ctx.Encode(value)
	if encoded.Err() != nil {
		return false
	}
	return m.definitions[definition].Unify(encoded).Validate(cue.Concrete(false)) == nil
}

// imagesOf reads the image of each container in an already-validated list.
func imagesOf(containers []any) []string {
	var images []string
	for _, entry := range containers {
		container, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		if image, ok := container["image"].(string); ok && image != "" {
			images = append(images, image)
		}
	}
	return images
}
