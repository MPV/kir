package fileutil

import (
	"log"
	"os"
	"path/filepath"
)

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
