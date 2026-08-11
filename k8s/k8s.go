// Package k8s finds container images in Kubernetes manifests that have been
// decoded into plain Go values.
//
// Images are found two ways. A structural walk infers them, matching nodes
// against the Kubernetes API types, which needs no configuration and reaches
// custom resources. A configuration (resources.yaml, plus anything the user
// supplies) states where a given kind keeps its images, which is exact and can
// reach what the walk cannot see. Config.FindImages puts them together: an
// entry wins for its kind, everything else is inferred.
package k8s

import (
	"encoding/json"
	"maps"
	"slices"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

// containerFields are the PodSpec fields that carry images, listed in the order
// kir reports them.
var containerFields = []string{"containers", "initContainers", "ephemeralContainers"}

// maxDepth bounds the walk. Real manifests nest a handful of levels; the limit
// only guards against absurd input.
const maxDepth = 100

// infer returns the images of every PodSpec-shaped node reachable from doc, a
// manifest decoded into plain Go values (maps, slices, scalars).
//
// Nothing here knows what a Deployment is. The walk descends until it meets a
// node shaped like a PodSpec, which is why a custom resource that embeds one is
// understood on the same footing as a built-in workload — and why a List needs
// no special case, its items being just more nodes.
func infer(doc any) []string {
	var images []string
	find(doc, &images, 0)
	return images
}

func find(node any, images *[]string, depth int) {
	if depth > maxDepth {
		return
	}

	switch n := node.(type) {
	case map[string]any:
		if found, ok := podSpecImages(n); ok {
			*images = append(*images, found...)
			return // a PodSpec does not contain another PodSpec
		}
		// Sorted, so output order depends on the manifest rather than on Go's
		// randomised map iteration.
		for _, key := range slices.Sorted(maps.Keys(n)) {
			find(n[key], images, depth+1)
		}
	case []any:
		for _, item := range n {
			find(item, images, depth+1)
		}
	}
}

// podSpecImages reports whether node is PodSpec-shaped, and if so its images.
//
// The test is not "has a field called containers" but "does that field decode
// into the real corev1 type, rejecting unknown fields". The Kubernetes Go types
// are the schema, so a custom resource embedding a genuine PodSpec matches,
// while a lookalike — a field named containers holding something else — does
// not.
func podSpecImages(node map[string]any) ([]string, bool) {
	var images []string
	matched := false

	for _, field := range containerFields {
		value, ok := node[field]
		if !ok {
			continue
		}
		containers, err := decodeContainers(field, value)
		if err != nil || len(containers) == 0 {
			continue
		}
		matched = true
		for _, container := range containers {
			if container.Image != "" {
				images = append(images, container.Image)
			}
		}
	}

	return images, matched
}

// decodeContainers strictly decodes one container list into its corev1 type.
// An error means "not that type", which is the signal the walk needs.
func decodeContainers(field string, value any) ([]corev1.Container, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	if field == "ephemeralContainers" {
		var ephemeral []corev1.EphemeralContainer
		if err := yaml.UnmarshalStrict(data, &ephemeral); err != nil {
			return nil, err
		}
		containers := make([]corev1.Container, 0, len(ephemeral))
		for _, ec := range ephemeral {
			containers = append(containers, corev1.Container(ec.EphemeralContainerCommon))
		}
		return containers, nil
	}

	var containers []corev1.Container
	if err := yaml.UnmarshalStrict(data, &containers); err != nil {
		return nil, err
	}
	return containers, nil
}
