package fileutil

import (
	"os"
	"path/filepath"
)

func FindFiles(args []string) ([]string, error) {
	var files []string
	for _, filePath := range args {
		// Handle glob patterns
		matchedFiles, err := filepath.Glob(filePath)
		if err != nil {
			return nil, err
		}

		for _, matchedFile := range matchedFiles {
			fileInfo, err := os.Stat(matchedFile)
			if err != nil {
				return nil, err
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
					return nil, err
				}
			} else {
				files = append(files, matchedFile)
			}
		}
	}
	return files, nil
}
