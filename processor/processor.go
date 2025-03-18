package processor

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/mpv/kir/yamlparser"
)

func ProcessStdin() ([]string, error) {
	reader := bufio.NewReader(os.Stdin)
	var data []byte
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("error reading stdin: %v", err)
		}
		data = append(data, line...)
		if err == io.EOF {
			break
		}
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
