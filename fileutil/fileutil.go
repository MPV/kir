package fileutil

import (
	"fmt"
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

		// filepath.Glob returns no matches and no error both for a glob
		// pattern that matches nothing and for a literal path that does not
		// exist. Either way the argument named nothing, which is almost
		// always a typo; surface it instead of silently skipping it.
		if len(matchedFiles) == 0 {
			return nil, fmt.Errorf("no such file or match for %q", filePath)
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
