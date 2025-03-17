package cmd

import (
	"log"
	"os"
	"path/filepath"

	"github.com/mpv/kir/processor"
)

func Execute(args []string) {
	for i := 1; i < len(os.Args); i++ {
		filePath := os.Args[i]

		if filePath == "-" {
			images, err := processor.ProcessStdin()
			if err != nil {
				log.Fatalf("error: %v", err)
			}
			logImages(images)
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
						images, err := processor.ProcessFile(path)
						if err != nil {
							log.Printf("error: %v", err)
						}
						logImages(images)
					}
					return nil
				})
				if err != nil {
					log.Fatalf("error: %v", err)
				}
			} else {
				// Handle glob patterns
				files, err := filepath.Glob(filePath)
				if err != nil {
					log.Fatalf("error: %v", err)
				}

				for _, file := range files {
					images, err := processor.ProcessFile(file)
					if err != nil {
						log.Printf("error: %v", err)
					}
					logImages(images)
				}
			}
		}
	}
}

func logImages(images []string) {
	for _, image := range images {
		log.Println(image)
	}
}
