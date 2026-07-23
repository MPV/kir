package cmd

import (
	"fmt"
	"io"
	"log"

	"github.com/mpv/kir/fileutil"
	"github.com/mpv/kir/processor"
)

// Execute writes the container images found in args to out. A single "-"
// argument reads a manifest stream from stdin; otherwise args are treated as
// files, directories, or globs.
//
// It returns an error for failures that prevent producing any output (reading
// stdin, or resolving the file arguments). Errors processing an individual
// file are logged and the remaining files are still processed.
func Execute(args []string, out io.Writer) error {
	if len(args) == 1 && args[0] == "-" {
		images, err := processor.ProcessStdin()
		if err != nil {
			return err
		}
		printImages(out, images)
		return nil
	}

	files, err := fileutil.FindFiles(args)
	if err != nil {
		return err
	}
	for _, filePath := range files {
		images, err := processor.ProcessFile(filePath)
		if err != nil {
			log.Printf("error: %v", err)
		}
		printImages(out, images)
	}
	return nil
}

func printImages(out io.Writer, images []string) {
	for _, image := range images {
		fmt.Fprintln(out, image)
	}
}
