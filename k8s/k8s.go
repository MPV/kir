// Package k8s finds container images in Kubernetes manifests that have been
// decoded into plain Go values. Where a kind keeps its PodSpec is configuration
// (resources.yaml), not Go code.
package k8s

import (
	_ "embed"
	"fmt"
	"strings"

	"sigs.k8s.io/yaml"
)

//go:embed resources.yaml
var defaultResources string

// containerFields are the PodSpec fields that carry images, listed in the order
// kir reports them.
var containerFields = []string{"containers", "initContainers", "ephemeralContainers"}

// Resource says where one kind keeps its PodSpecs.
type Resource struct {
	Kind string `json:"kind"`
	// PodSpecs are paths to PodSpec-shaped nodes.
	PodSpecs []string `json:"podSpecs,omitempty"`
	// Documents are paths to whole objects, each processed in its own right.
	// A List uses this to reach its items.
	Documents []string `json:"documents,omitempty"`
}

// Config maps kinds to the paths at which they hold PodSpecs.
type Config struct {
	Resources []Resource `json:"resources"`

	byKind map[string]Resource
}

// LoadConfig parses a resource configuration.
func LoadConfig(data []byte) (*Config, error) {
	var config Config
	if err := yaml.UnmarshalStrict(data, &config); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	config.index()
	return &config, nil
}

// DefaultConfig returns the built-in configuration.
func DefaultConfig() *Config {
	config, err := LoadConfig([]byte(defaultResources))
	if err != nil {
		// The embedded config is parsed in tests; a failure here is a bug.
		panic(fmt.Sprintf("embedded config does not parse: %v", err))
	}
	return config
}

// Merge returns a copy of c with other's entries applied over it. An entry for
// a kind already configured replaces it, so a user can correct a built-in as
// well as add a custom resource.
func (c *Config) Merge(other *Config) *Config {
	merged := &Config{Resources: append(append([]Resource{}, c.Resources...), other.Resources...)}
	merged.index()
	return merged
}

func (c *Config) index() {
	c.byKind = make(map[string]Resource, len(c.Resources))
	for _, resource := range c.Resources {
		c.byKind[resource.Kind] = resource
	}
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
	for _, path := range resource.PodSpecs {
		for _, node := range resolve(doc, path) {
			images = append(images, podSpecImages(node)...)
		}
	}
	for _, path := range resource.Documents {
		for _, node := range resolve(doc, path) {
			if nested, ok := node.(map[string]any); ok {
				images = append(images, c.FindImages(nested)...)
			}
		}
	}
	return images
}

// resolve follows a dot-separated path from node, expanding lists at any
// segment marked `[*]`. It returns every node the path reaches, which is none
// when the path does not apply to this document.
func resolve(node any, path string) []any {
	nodes := []any{node}

	for _, segment := range strings.Split(path, ".") {
		expand := strings.HasSuffix(segment, "[*]")
		segment = strings.TrimSuffix(segment, "[*]")

		var next []any
		for _, current := range nodes {
			object, ok := current.(map[string]any)
			if !ok {
				continue
			}
			value, ok := object[segment]
			if !ok {
				continue
			}
			if !expand {
				next = append(next, value)
				continue
			}
			if list, ok := value.([]any); ok {
				next = append(next, list...)
			}
		}
		nodes = next
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
		containers, ok := podSpec[field].([]any)
		if !ok {
			continue
		}
		for _, entry := range containers {
			container, ok := entry.(map[string]any)
			if !ok {
				continue
			}
			if image, ok := container["image"].(string); ok && image != "" {
				images = append(images, image)
			}
		}
	}
	return images
}
