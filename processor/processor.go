package processor

import (
	"fmt"
	"io"
	"os"

	"github.com/mpv/kir/k8s"
	"github.com/mpv/kir/yamlparser"
)

// ProcessStdin reads a (possibly multi-document) manifest stream from r and
// returns its images. Taking an io.Reader keeps it testable and lets the CLI
// inject stdin.
func ProcessStdin(config *k8s.Config, r io.Reader) ([]string, error) {
	return yamlparser.ProcessReader(config, r)
}

func ProcessFile(config *k8s.Config, filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}
	defer file.Close()
	return yamlparser.ProcessReader(config, file)
}
