package processor

import (
	"bufio"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/mpv/kir/yamlparser"
)

func ProcessStdin() ([]string, error) {
	reader := bufio.NewReader(os.Stdin)
	var data []byte
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			log.Fatalf("error reading stdin: %v", err)
		}
		data = append(data, line...)
		if err == io.EOF {
			break
		}
	}
	return yamlparser.ProcessData(data)
}

func ProcessFile(filePath string) ([]string, error) {
	// Read the YAML file
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("error reading file %s: %v", filePath, err)
		return nil, err
	}
	return yamlparser.ProcessData(data)
}

func FindFiles(args []string) []string {
	var files []string
	for _, filePath := range args {
		if filePath == "-" {
			files = append(files, filePath)
		} else {
			fileInfo, err := os.Stat(filePath)
			if err != nil {
				log.Fatalf("error: %v", err)
			}

			if fileInfo.IsDir() {
				err := filepath.Walk(filePath, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if !info.IsDir() {
						files = append(files, path)
					}
					return nil
				})
				if err != nil {
					log.Fatalf("error: %v", err)
				}
			} else {
				// Handle glob patterns
				matchedFiles, err := filepath.Glob(filePath)
				if err != nil {
					log.Fatalf("error: %v", err)
				}
				files = append(files, matchedFiles...)
			}
		}
	}
	return files
}
