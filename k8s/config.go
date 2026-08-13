package k8s

import (
	_ "embed"
	"fmt"

	"github.com/jmespath/go-jmespath"
	"sigs.k8s.io/yaml"
)

//go:embed resources.yaml
var defaultResources string

// Resource says where one kind keeps its images, as JMESPath expressions.
type Resource struct {
	Kind string `json:"kind"`
	// PodSpecs select PodSpec-shaped nodes, whose container fields are read.
	PodSpecs []string `json:"podSpecs,omitempty"`
	// Containers select containers directly, for resources that hold bare
	// containers rather than a PodSpec — which the walk cannot recognise,
	// there being no PodSpec shape to match.
	Containers []string `json:"containers,omitempty"`
	// Documents select whole objects, each processed in its own right.
	Documents []string `json:"documents,omitempty"`
}

// Config maps kinds to where they keep their images. A kind with no entry is
// not unsupported — it is inferred by the structural walk instead.
type Config struct {
	Resources []Resource `json:"resources"`

	byKind map[string]expressions
}

// expressions holds one resource's compiled queries.
type expressions struct {
	podSpecs   []*jmespath.JMESPath
	containers []*jmespath.JMESPath
	documents  []*jmespath.JMESPath
}

// LoadConfig parses a resource configuration and compiles its expressions. A
// malformed expression is an error here rather than one that silently matches
// nothing later.
func LoadConfig(data []byte) (*Config, error) {
	var config Config
	if err := yaml.UnmarshalStrict(data, &config); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	if err := config.compile(); err != nil {
		return nil, err
	}
	return &config, nil
}

// DefaultConfig returns the built-in configuration.
func DefaultConfig() *Config {
	config, err := LoadConfig([]byte(defaultResources))
	if err != nil {
		// The embedded config is loaded in tests; a failure here is a bug.
		panic(fmt.Sprintf("embedded config does not load: %v", err))
	}
	return config
}

// Merge returns a copy of c with other's entries applied over it. An entry for
// a kind already configured replaces it, so a user can correct a built-in as
// well as add a resource.
func (c *Config) Merge(other *Config) *Config {
	merged := &Config{Resources: append(append([]Resource{}, c.Resources...), other.Resources...)}
	// Both halves compiled when they were loaded, so this cannot fail.
	if err := merged.compile(); err != nil {
		panic(fmt.Sprintf("merging already-compiled configs: %v", err))
	}
	return merged
}

func (c *Config) compile() error {
	c.byKind = make(map[string]expressions, len(c.Resources))
	for _, resource := range c.Resources {
		var compiled expressions
		var err error
		if compiled.podSpecs, err = compileAll(resource.Kind, "podSpecs", resource.PodSpecs); err != nil {
			return err
		}
		if compiled.containers, err = compileAll(resource.Kind, "containers", resource.Containers); err != nil {
			return err
		}
		if compiled.documents, err = compileAll(resource.Kind, "documents", resource.Documents); err != nil {
			return err
		}
		c.byKind[resource.Kind] = compiled
	}
	return nil
}

func compileAll(kind, field string, exprs []string) ([]*jmespath.JMESPath, error) {
	compiled := make([]*jmespath.JMESPath, 0, len(exprs))
	for _, expr := range exprs {
		parsed, err := jmespath.Compile(expr)
		if err != nil {
			return nil, fmt.Errorf("%s.%s: %q: %w", kind, field, expr, err)
		}
		compiled = append(compiled, parsed)
	}
	return compiled, nil
}

// Kinds returns the configured kinds, for diagnostics.
func (c *Config) Kinds() []string {
	kinds := make([]string, 0, len(c.Resources))
	for _, resource := range c.Resources {
		kinds = append(kinds, resource.Kind)
	}
	return kinds
}

// FindImages returns the images of doc, a manifest decoded into plain Go
// values.
//
// An entry for the document's kind decides on its own: its expressions are
// followed and the walk is not consulted. Everything else is inferred
// structurally. So configuration is never needed for a document whose shape
// speaks for itself, and always available for one whose does not — a resource
// holding bare containers, or one the walk reads wrongly, which an entry with
// no expressions silences outright.
//
// The two never both contribute to the same document, so an image cannot be
// reported twice.
func (c *Config) FindImages(doc map[string]any) []string {
	if images, configured := c.lookup(doc); configured {
		return images
	}
	return infer(doc)
}

// lookup applies the configured expressions for doc's kind, reporting whether
// the kind was configured at all — which is what makes an entry with no
// expressions mean "this kind has no images", rather than "fall back".
func (c *Config) lookup(doc map[string]any) ([]string, bool) {
	kind, _ := doc["kind"].(string)
	resource, ok := c.byKind[kind]
	if !ok {
		return nil, false
	}

	var images []string
	for _, expr := range resource.podSpecs {
		for _, node := range search(expr, doc) {
			images = append(images, configuredPodSpecImages(node)...)
		}
	}
	for _, expr := range resource.containers {
		for _, node := range search(expr, doc) {
			images = append(images, containerImages(node)...)
		}
	}
	// A nested document goes back through FindImages, so an item of a
	// configured List is itself either configured or inferred.
	for _, expr := range resource.documents {
		for _, node := range search(expr, doc) {
			if nested, ok := node.(map[string]any); ok {
				images = append(images, c.FindImages(nested)...)
			}
		}
	}
	return images, true
}

// search runs one expression and returns the nodes it selected.
//
// A JMESPath projection yields one entry per input element, null where the
// field is absent — `spec.templates[*].container` over templates that hold no
// container, say — so nulls are dropped rather than counted as matches.
func search(expr *jmespath.JMESPath, doc any) []any {
	result, err := expr.Search(doc)
	if err != nil || result == nil {
		return nil
	}

	list, ok := result.([]any)
	if !ok {
		return []any{result}
	}

	nodes := make([]any, 0, len(list))
	for _, node := range list {
		if node != nil {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// configuredPodSpecImages reads the images out of a node the configuration has
// declared to be a PodSpec. Unlike the walk, nothing here checks that claim:
// an explicit entry is taken at its word.
func configuredPodSpecImages(node any) []string {
	podSpec, ok := node.(map[string]any)
	if !ok {
		return nil
	}

	var images []string
	for _, field := range containerFields {
		images = append(images, containerImages(podSpec[field])...)
	}
	return images
}

// containerImages reads the image of a container, or of every container in a
// list of them.
func containerImages(node any) []string {
	switch n := node.(type) {
	case []any:
		var images []string
		for _, item := range n {
			images = append(images, containerImages(item)...)
		}
		return images
	case map[string]any:
		if image, ok := n["image"].(string); ok && image != "" {
			return []string{image}
		}
	}
	return nil
}
