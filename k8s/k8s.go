// Package k8s finds container images in Kubernetes manifests that have been
// decoded into plain Go values. What counts as a PodSpec is defined by a CUE
// schema (podspec.cue) rather than by Go code.
package k8s

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"maps"
	"path"
	"slices"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	cuejson "cuelang.org/go/encoding/json"
)

// schemaFS holds the schema kir matches against: podspec.cue, and the CUE
// definitions generated from k8s.io/api by `cue get go` that it imports.
//
//go:embed all:cue.mod podspec.cue
var schemaFS embed.FS

// schemaRoot is where the embedded CUE module is mounted for the loader. The
// files never touch disk; the path only has to be absolute and consistent.
const schemaRoot = "/kir"

// schemaFile is the unit a user schema replaces.
const schemaFile = "podspec.cue"

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

// NewMatcher compiles a CUE schema. An empty schema uses the embedded
// podspec.cue; otherwise it stands in for that file, inside the same module, so
// a user schema can import the generated Kubernetes definitions too. The schema
// must define #Containers and #EphemeralContainers.
func NewMatcher(schema string) (*Matcher, error) {
	overlay, err := schemaOverlay(schema)
	if err != nil {
		return nil, err
	}

	instances := load.Instances([]string{"."}, &load.Config{Dir: schemaRoot, Overlay: overlay})
	if len(instances) != 1 {
		return nil, fmt.Errorf("loading schema: expected 1 instance, got %d", len(instances))
	}
	if err := instances[0].Err; err != nil {
		return nil, fmt.Errorf("loading schema: %w", err)
	}

	ctx := cuecontext.New()
	value := ctx.BuildInstance(instances[0])
	if err := value.Err(); err != nil {
		return nil, fmt.Errorf("compiling schema: %w", err)
	}

	definitions := map[string]cue.Value{}
	for _, name := range []string{"#Containers", "#EphemeralContainers"} {
		definition := value.LookupPath(cue.ParsePath(name))
		if err := definition.Err(); err != nil {
			return nil, fmt.Errorf("schema is missing %s: %w", name, err)
		}
		// Deliberately not .Eval()'d: pre-evaluating these is ~7x faster, but
		// it discards the closedness that makes a definition reject unknown
		// fields, and every lookalike would start matching.
		definitions[name] = definition
	}

	return &Matcher{ctx: ctx, definitions: definitions}, nil
}

// schemaOverlay mounts the embedded CUE module for the loader, substituting a
// caller-supplied schema for podspec.cue when one is given.
func schemaOverlay(schema string) (map[string]load.Source, error) {
	overlay := map[string]load.Source{}
	err := fs.WalkDir(schemaFS, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		data, err := schemaFS.ReadFile(name)
		if err != nil {
			return err
		}
		overlay[path.Join(schemaRoot, name)] = load.FromBytes(data)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("reading embedded schema: %w", err)
	}

	if schema != "" {
		overlay[path.Join(schemaRoot, schemaFile)] = load.FromString(schema)
	}
	return overlay, nil
}

// DefaultMatcher returns a Matcher over the embedded schema.
func DefaultMatcher() *Matcher {
	matcher, err := NewMatcher("")
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
//
// The candidate goes to CUE as JSON rather than through ctx.Encode. The
// manifest was decoded YAML-to-JSON-to-Go, which leaves every number a float64,
// and encoding that directly would offer CUE 8E+1 where the Kubernetes schema
// says int32 — every container with a port would fail to match. Round-tripping
// through JSON keeps whole numbers whole.
func (m *Matcher) unifies(definition string, value any) bool {
	data, err := json.Marshal(value)
	if err != nil {
		return false
	}
	expr, err := cuejson.Extract("", data)
	if err != nil {
		return false
	}
	encoded := m.ctx.BuildExpr(expr)
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
