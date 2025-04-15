package cueparser

import (
	"fmt"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/encoding/yaml"
)

// ProcessData processes YAML data and extracts container images from PodSpec
func ProcessData(data []byte) ([]string, error) {
	// Create a CUE context
	ctx := cuecontext.New()

	// Load the PodSpec schema from the CUE Central Registry
	bis := load.Instances([]string{"cue.dev/x/k8s.io/api/core/v1"}, nil)
	if len(bis) == 0 {
		return nil, fmt.Errorf("failed to load PodSpec schema")
	}

	pkgV := ctx.BuildInstance(bis[0])
	if pkgV.Err() != nil {
		// If we can't load from the Central Registry, return an error
		// This will cause the code to fall back to the original implementation
		return nil, fmt.Errorf("failed to build PodSpec schema: %v", pkgV.Err())
	}

	podSpec := pkgV.LookupPath(cue.ParsePath("#PodSpec"))
	if podSpec.Err() != nil {
		return nil, fmt.Errorf("failed to lookup PodSpec: %v", podSpec.Err())
	}

	// Load the YAML data
	dataV, err := yaml.Extract("", data)
	if err != nil {
		return nil, fmt.Errorf("failed to extract YAML: %v", err)
	}

	dataValue := ctx.BuildFile(dataV)
	if dataValue.Err() != nil {
		return nil, fmt.Errorf("failed to build YAML data: %v", dataValue.Err())
	}

	// Unify the YAML value with the schema and validate
	combined := podSpec.Unify(dataValue)
	if err := combined.Validate(cue.Concrete(true)); err != nil {
		return nil, fmt.Errorf("validation error: %v", err)
	}

	// Extract container images
	var images []string

	// Try to get containers from the validated PodSpec
	containersValue := combined.LookupPath(cue.ParsePath("containers"))
	if containersValue.Exists() {
		// Iterate through containers
		iter, err := containersValue.List()
		if err != nil {
			return nil, fmt.Errorf("failed to iterate containers: %v", err)
		}

		for iter.Next() {
			container := iter.Value()
			imageValue := container.LookupPath(cue.ParsePath("image"))
			if imageValue.Exists() {
				image, err := imageValue.String()
				if err != nil {
					return nil, fmt.Errorf("failed to get image string: %v", err)
				}
				images = append(images, image)
			}
		}
	}

	// Try to get initContainers from the validated PodSpec
	initContainersValue := combined.LookupPath(cue.ParsePath("initContainers"))
	if initContainersValue.Exists() {
		// Iterate through initContainers
		iter, err := initContainersValue.List()
		if err != nil {
			return nil, fmt.Errorf("failed to iterate initContainers: %v", err)
		}

		for iter.Next() {
			container := iter.Value()
			imageValue := container.LookupPath(cue.ParsePath("image"))
			if imageValue.Exists() {
				image, err := imageValue.String()
				if err != nil {
					return nil, fmt.Errorf("failed to get image string: %v", err)
				}
				images = append(images, image)
			}
		}
	}

	return images, nil
}

// ProcessKubernetesYAML processes a Kubernetes YAML document and extracts container images
func ProcessKubernetesYAML(data []byte) ([]string, error) {
	// First, try to extract the PodSpec directly
	images, err := ProcessData(data)
	if err == nil && len(images) > 0 {
		return images, nil
	}

	// If that fails, try to extract the PodSpec from a Kubernetes resource
	// This is a simplified approach - in a real implementation, you would need to
	// handle different Kubernetes resource types (Deployment, StatefulSet, etc.)

	// For now, we'll just return the error from the first attempt
	return nil, fmt.Errorf("failed to extract PodSpec: %v", err)
}

// ProcessKubernetesListYAML processes a Kubernetes List YAML document and extracts container images
func ProcessKubernetesListYAML(data []byte) ([]string, error) {
	// Split the YAML document by "---" to handle multiple resources
	docs := strings.Split(string(data), "---")

	var allImages []string
	for _, doc := range docs {
		if strings.TrimSpace(doc) == "" {
			continue
		}

		images, err := ProcessKubernetesYAML([]byte(doc))
		if err != nil {
			// Log the error but continue processing other documents
			fmt.Printf("Error processing document: %v\n", err)
			continue
		}

		allImages = append(allImages, images...)
	}

	return allImages, nil
}
