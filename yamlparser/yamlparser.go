package yamlparser

import (
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
)

var supportedKinds = []string{"Pod", "Deployment", "DaemonSet", "ReplicaSet", "StatefulSet", "Job", "CronJob"}

// ProcessData processes YAML data and extracts container images using Cue
func ProcessData(data []byte) ([]string, error) {
	ctx := cuecontext.New()

	// Load the PodSpec schema from the CUE Central Registry
	bis := load.Instances([]string{"cue.dev/x/k8s.io/api/core/v1"}, nil)
	if len(bis) == 0 {
		return nil, fmt.Errorf("failed to load schema from CUE Central Registry")
	}

	pkgV := ctx.BuildInstance(bis[0])
	if pkgV.Err() != nil {
		return nil, fmt.Errorf("failed to build schema instance: %v", pkgV.Err())
	}

	podSpec := pkgV.LookupPath(cue.ParsePath("#PodSpec"))
	if podSpec.Err() != nil {
		return nil, fmt.Errorf("failed to find PodSpec schema: %v", podSpec.Err())
	}

	// Load the YAML data into a Cue value
	cueValue := ctx.CompileBytes(data)
	if cueValue.Err() != nil {
		return nil, fmt.Errorf("failed to decode YAML: %v", cueValue.Err())
	}

	// Unify the YAML data with the schema
	unified := podSpec.Unify(cueValue)
	if err := unified.Validate(cue.Concrete(true)); err != nil {
		return nil, fmt.Errorf("validation error: %v", err)
	}

	// Extract container images
	containers := unified.LookupPath(cue.ParsePath("spec.containers"))
	if !containers.Exists() {
		return nil, fmt.Errorf("containers section not found")
	}

	var images []string
	iter, err := containers.List()
	if err != nil {
		return nil, fmt.Errorf("failed to iterate containers: %v", err)
	}

	for iter.Next() {
		image := iter.Value().LookupPath(cue.ParsePath("image"))
		if image.Exists() {
			imgStr, err := image.String()
			if err != nil {
				return nil, fmt.Errorf("failed to extract image string: %v", err)
			}
			images = append(images, imgStr)
		}
	}

	return images, nil
}
