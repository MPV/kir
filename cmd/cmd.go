package cmd

import (
	"fmt"
	"log"

	"github.com/mpv/kir/fileutil"
	"github.com/mpv/kir/processor"
)

func Execute(args []string) {
	if len(args) == 1 && args[0] == "-" {
		images, err := processor.ProcessStdin()
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		logImages(images)
		return
	}
	files := fileutil.FindFiles(args)
	for _, filePath := range files {
		images, err := processor.ProcessFile(filePath)
		if err != nil {
			log.Printf("error: %v", err)
		}
		logImages(images)
	}
}

func logImages(images []string) {
	for _, image := range images {
		fmt.Println(image)
	}
}
