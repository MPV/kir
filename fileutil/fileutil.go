package fileutil

import (
	"log"
	"os"
	"path/filepath"
)

func FindFiles(args []string) []string {
	var files []string
	for _, filePath := range args {
		// Handle glob patterns
		matchedFiles, err := filepath.Glob(filePath)
		if err != nil {
			log.Fatalf("error: %v", err)
		}

		for _, matchedFile := range matchedFiles {
			fileInfo, err := os.Stat(matchedFile)
			if err != nil {
				log.Fatalf("error: %v", err)
			}

			if fileInfo.IsDir() {
				err := filepath.Walk(matchedFile, func(path string, info os.FileInfo, err error) error {
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
				files = append(files, matchedFile)
			}
		}
	}
	return files
}
