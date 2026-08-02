package processor

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/mpv/kir/yamlparser"
)

// ProcessStdin reads a manifest stream from r and returns its images. Taking an
// io.Reader (rather than reading os.Stdin directly) keeps it testable and lets
// the CLI inject stdin.
func ProcessStdin(r io.Reader) ([]string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("error reading stdin: %v", err)
	}
	return yamlparser.ProcessData(data)
}

func ProcessFile(filePath string) ([]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	var images []string
	docs := bytes.Split(data, []byte("\n---\n"))
	for _, doc := range docs {
		imgs, err := yamlparser.ProcessData(doc)
		if err != nil {
			return nil, fmt.Errorf("error processing document: %v", err)
		}
		images = append(images, imgs...)
	}
	return images, nil
}
