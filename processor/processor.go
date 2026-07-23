package processor

import (
	"fmt"
	"os"

	"github.com/mpv/kir/yamlparser"
)

func ProcessStdin() ([]string, error) {
	return yamlparser.ProcessReader(os.Stdin)
}

func ProcessFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}
	defer file.Close()
	return yamlparser.ProcessReader(file)
}
