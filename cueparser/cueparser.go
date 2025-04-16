package cueparser

import (
	"fmt"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/encoding/yaml"
)

// ProcessData processes YAML data and extracts container images from PodSpec
func ProcessData(data []byte) ([]string, error) {
	// Create a CUE context
	ctx := cuecontext.New()

	// Load the PodSpec schema from the CUE Central Registry
	bis := load.Instances([]string{"cue.dev/x/k8s.io/api/core/v1"}, nil)
	if len(bis) == 0 {
		// Fall back to local schema if Central Registry is not available
		return processWithLocalSchema(data)
	}

	pkgV := ctx.BuildInstance(bis[0])
	if pkgV.Err() != nil {
		// Fall back to local schema if Central Registry is not available
		return processWithLocalSchema(data)
	}

	podSpec := pkgV.LookupPath(cue.ParsePath("#PodSpec"))
	if podSpec.Err() != nil {
		// Fall back to local schema if Central Registry is not available
		return processWithLocalSchema(data)
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

// processWithLocalSchema processes YAML data using a local schema definition
func processWithLocalSchema(data []byte) ([]string, error) {
	// Create a CUE context
	ctx := cuecontext.New()

	// Load the local schema with support for all Kubernetes resource types
	schema := `
	#Container: {
		name: string
		image: string
		command?: [...string]
		args?: [...string]
		ports?: [...{
			containerPort: int
			protocol?: string
			name?: string
			hostPort?: int
		}]
		env?: [...{
			name: string
			value?: string
			valueFrom?: {...}
		}]
		volumeMounts?: [...{
			name: string
			mountPath: string
			readOnly?: bool
		}]
		resources?: {
			limits?: {...}
			requests?: {...}
		}
		...
	}

	#PodSpec: {
		containers?: [...#Container]
		initContainers?: [...#Container]
		volumes?: [...{
			name: string
			...
		}]
		restartPolicy?: string
		...
	}

	#ObjectMeta: {
		name?: string
		namespace?: string
		labels?: [string]: string
		annotations?: [string]: string
		...
	}

	#PodTemplateSpec: {
		metadata?: #ObjectMeta
		spec: #PodSpec
	}

	#Pod: {
		apiVersion: string
		kind: "Pod"
		metadata?: #ObjectMeta
		spec: #PodSpec
	}

	#Deployment: {
		apiVersion: string
		kind: "Deployment"
		metadata?: #ObjectMeta
		spec: {
			replicas?: int
			selector?: {...}
			template: #PodTemplateSpec
			...
		}
	}

	#DaemonSet: {
		apiVersion: string
		kind: "DaemonSet"
		metadata?: #ObjectMeta
		spec: {
			selector?: {...}
			template: #PodTemplateSpec
			...
		}
	}

	#StatefulSet: {
		apiVersion: string
		kind: "StatefulSet"
		metadata?: #ObjectMeta
		spec: {
			replicas?: int
			selector?: {...}
			template: #PodTemplateSpec
			...
		}
	}

	#ReplicaSet: {
		apiVersion: string
		kind: "ReplicaSet"
		metadata?: #ObjectMeta
		spec: {
			replicas?: int
			selector?: {...}
			template: #PodTemplateSpec
			...
		}
	}

	#Job: {
		apiVersion: string
		kind: "Job"
		metadata?: #ObjectMeta
		spec: {
			template: #PodTemplateSpec
			...
		}
	}

	#CronJob: {
		apiVersion: string
		kind: "CronJob"
		metadata?: #ObjectMeta
		spec: {
			schedule: string
			jobTemplate: {
				metadata?: #ObjectMeta
				spec: {
					template: #PodTemplateSpec
					...
				}
			}
			...
		}
	}
	`
	schemaValue := ctx.CompileString(schema)
	if schemaValue.Err() != nil {
		return nil, fmt.Errorf("failed to compile schema: %v", schemaValue.Err())
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

	// Try to determine the resource kind
	kindValue := dataValue.LookupPath(cue.ParsePath("kind"))
	if !kindValue.Exists() {
		return nil, fmt.Errorf("kind field not found in resource")
	}

	kind, err := kindValue.String()
	if err != nil {
		return nil, fmt.Errorf("failed to get kind string: %v", err)
	}

	// Get the appropriate schema based on the resource kind
	var resourceSchema cue.Value
	switch kind {
	case "Pod":
		resourceSchema = schemaValue.LookupPath(cue.ParsePath("#Pod"))
	case "Deployment":
		resourceSchema = schemaValue.LookupPath(cue.ParsePath("#Deployment"))
	case "DaemonSet":
		resourceSchema = schemaValue.LookupPath(cue.ParsePath("#DaemonSet"))
	case "StatefulSet":
		resourceSchema = schemaValue.LookupPath(cue.ParsePath("#StatefulSet"))
	case "ReplicaSet":
		resourceSchema = schemaValue.LookupPath(cue.ParsePath("#ReplicaSet"))
	case "Job":
		resourceSchema = schemaValue.LookupPath(cue.ParsePath("#Job"))
	case "CronJob":
		resourceSchema = schemaValue.LookupPath(cue.ParsePath("#CronJob"))
	default:
		return nil, fmt.Errorf("unsupported resource kind: %s", kind)
	}

	// Unify the YAML value with the schema and validate
	combined := resourceSchema.Unify(dataValue)
	if err := combined.Validate(cue.Concrete(true)); err != nil {
		return nil, fmt.Errorf("validation error: %v", err)
	}

	// Extract container images based on the resource kind
	var podSpec cue.Value
	switch kind {
	case "Pod":
		podSpec = combined.LookupPath(cue.ParsePath("spec"))
	case "Deployment", "DaemonSet", "StatefulSet", "ReplicaSet", "Job":
		podSpec = combined.LookupPath(cue.ParsePath("spec.template.spec"))
	case "CronJob":
		podSpec = combined.LookupPath(cue.ParsePath("spec.jobTemplate.spec.template.spec"))
	}

	if podSpec.Err() != nil {
		return nil, fmt.Errorf("failed to get PodSpec: %v", podSpec.Err())
	}

	var images []string

	// Extract container images from the PodSpec
	containersValue := podSpec.LookupPath(cue.ParsePath("containers"))
	if containersValue.Exists() {
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

	// Extract init container images from the PodSpec
	initContainersValue := podSpec.LookupPath(cue.ParsePath("initContainers"))
	if initContainersValue.Exists() {
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
	return ProcessData(data)
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
