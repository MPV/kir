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
// stdin, or resolving the file arguments). Errors processing individual files
// are logged and the remaining files are still processed, but Execute then
// returns an error so the process exits non-zero: a tool that feeds an image
// scanner must not report success when some manifests could not be read.
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
	failures := 0
	for _, filePath := range files {
		images, err := processor.ProcessFile(filePath)
		if err != nil {
			log.Printf("error: %v", err)
			failures++
			continue
		}
		printImages(out, images)
	}
	if failures > 0 {
		return fmt.Errorf("%d of %d file(s) could not be processed", failures, len(files))
	}
	return nil
}

func printImages(out io.Writer, images []string) {
	for _, image := range images {
		fmt.Fprintln(out, image)
	}
}
