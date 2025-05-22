package yamlparser

import (
	"github.com/mpv/kir/cueparser"
)

// ProcessData processes YAML data and extracts container images
func ProcessData(data []byte) ([]string, error) {
	// Use the CUE-based parser
	return cueparser.ProcessData(data)
}

// ProcessKubernetesListYAML processes a Kubernetes List YAML document and extracts container images
func ProcessKubernetesListYAML(data []byte) ([]string, error) {
	// Use the CUE-based parser
	return cueparser.ProcessKubernetesListYAML(data)
}
