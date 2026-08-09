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
func ProcessStdin(matcher *k8s.Matcher, r io.Reader) ([]string, error) {
	return yamlparser.ProcessReader(matcher, r)
}

func ProcessFile(matcher *k8s.Matcher, filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}
	defer file.Close()
	return yamlparser.ProcessReader(matcher, file)
}
