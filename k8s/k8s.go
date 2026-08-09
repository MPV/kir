// Package k8s finds container images in Kubernetes manifests that have been
// decoded into plain Go values. Where a kind keeps its containers is
// configuration (resources.yaml), not Go code.
package k8s

import (
	_ "embed"
	"fmt"

	"github.com/jmespath/go-jmespath"
	"sigs.k8s.io/yaml"
)

//go:embed resources.yaml
var defaultResources string

// containerFields are the PodSpec fields that carry images, listed in the order
// kir reports them.
var containerFields = []string{"containers", "initContainers", "ephemeralContainers"}

// Resource says where one kind keeps its images, as JMESPath expressions.
type Resource struct {
	Kind string `json:"kind"`
	// PodSpecs select PodSpec-shaped nodes, whose container fields are read.
	PodSpecs []string `json:"podSpecs,omitempty"`
	// Containers select containers directly, for resources that hold bare
	// containers rather than a PodSpec.
	Containers []string `json:"containers,omitempty"`
	// Documents select whole objects, each processed in its own right. A List
	// uses this to reach its items.
	Documents []string `json:"documents,omitempty"`
}

// Config maps kinds to where they keep their images.
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
// malformed expression is an error here rather than a path that silently
// matches nothing later.
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
// well as add a custom resource.
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
// values. A kind with no entry in the configuration yields nothing.
func (c *Config) FindImages(doc map[string]any) []string {
	kind, _ := doc["kind"].(string)
	resource, ok := c.byKind[kind]
	if !ok {
		return nil
	}

	var images []string
	for _, expr := range resource.podSpecs {
		for _, node := range search(expr, doc) {
			images = append(images, podSpecImages(node)...)
		}
	}
	for _, expr := range resource.containers {
		for _, node := range search(expr, doc) {
			images = append(images, containerImages(node)...)
		}
	}
	for _, expr := range resource.documents {
		for _, node := range search(expr, doc) {
			if nested, ok := node.(map[string]any); ok {
				images = append(images, c.FindImages(nested)...)
			}
		}
	}
	return images
}

// search runs one expression and returns the nodes it selected.
//
// A JMESPath projection yields one entry per input element, null where the
// field is absent — `spec.templates[*].container` over templates that hold no
// container, say — so nulls are dropped rather than counted as matches. An
// expression selecting nothing returns no nodes, the normal case for a path
// that does not apply to this document.
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

// podSpecImages reads the images out of a node the configuration has declared
// to be a PodSpec. Nothing here checks that claim: precision comes from the
// configuration being right.
func podSpecImages(node any) []string {
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
