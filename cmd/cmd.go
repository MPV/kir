package cmd

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/mpv/kir/fileutil"
	"github.com/mpv/kir/processor"
)

// Build metadata, injected at release time via -ldflags by GoReleaser.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// Run executes the kir CLI and returns the process exit code. A single "-"
// argument reads a manifest stream from stdin; otherwise the arguments are
// treated as files, directories, or globs. Images are written to stdout and
// errors to stderr.
//
// Run performs no os.Exit and touches neither os.Stdin/os.Stdout/os.Stderr nor
// the global logger, so the whole CLI contract — argv and stdin in, and stdout,
// stderr, and exit code out — is exercised by an in-process call.
//
// If any file fails to process, its error is logged and the remaining files are
// still processed, but Run returns a non-zero exit code: a tool that feeds an
// image scanner must not report success when some manifests could not be read.
func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	logger := log.New(stderr, "", 0)

	if len(args) == 0 {
		logger.Print("Usage: kir <file_path> [<file_path_2> ...] | kir - | kir --version")
		return 1
	}

	if len(args) == 1 {
		switch args[0] {
		case "-":
			if stdin == nil {
				stdin = strings.NewReader("")
			}
			images, err := processor.ProcessStdin(stdin)
			if err != nil {
				logger.Printf("error: %v", err)
				return 1
			}
			printImages(stdout, images)
			return 0
		case "--version", "-v":
			fmt.Fprintf(stdout, "kir %s (commit %s, built %s)\n", version, commit, date)
			return 0
		}
	}

	files, err := fileutil.FindFiles(args)
	if err != nil {
		logger.Printf("error: %v", err)
		return 1
	}
	failures := 0
	for _, filePath := range files {
		images, err := processor.ProcessFile(filePath)
		if err != nil {
			logger.Printf("error: %v", err)
			failures++
			continue
		}
		printImages(stdout, images)
	}
	if failures > 0 {
		return 1
	}
	return 0
}

func printImages(w io.Writer, images []string) {
	for _, image := range images {
		fmt.Fprintln(w, image)
	}
}
