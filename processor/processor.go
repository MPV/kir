package processor

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/mpv/kir/yamlparser"
)

func ProcessStdin() ([]string, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("error reading stdin: %v", err)
	}
	return processDocuments(data)
}

func ProcessFile(filePath string) ([]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}
	return processDocuments(data)
}

// processDocuments splits a (possibly multi-document) YAML stream and collects
// the images from every document. Both the file and stdin paths go through it
// so they handle multi-document input identically.
func processDocuments(data []byte) ([]string, error) {
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
