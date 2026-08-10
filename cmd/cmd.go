package cmd

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/mpv/kir/fileutil"
	"github.com/mpv/kir/imageref"
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
			failures := logErrors(logger, err)
			failures += printImages(stdout, logger, "stdin", images)
			if failures > 0 {
				return 1
			}
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
		// Not `continue`: a file that failed on one document may still have
		// yielded images from the others, and dropping them would defeat the
		// point of reporting the failure.
		failures += logErrors(logger, err)
		failures += printImages(stdout, logger, filePath, images)
	}
	if failures > 0 {
		return 1
	}
	return 0
}

// logErrors writes one "error:" line per failure and returns how many it wrote.
//
// A single input can fail on more than one document, and ProcessReader packs
// those into one joined error. Unwrapping it here keeps stderr to one failure
// per line, which is what anything reading that stream expects.
func logErrors(logger *log.Logger, err error) int {
	if err == nil {
		return 0
	}
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		count := 0
		for _, e := range joined.Unwrap() {
			count += logErrors(logger, e)
		}
		return count
	}
	logger.Printf("error: %v", err)
	return 1
}

// printImages writes every reportable image to w, reports the rest to logger,
// and returns how many it rejected. An unreportable reference gets the same
// treatment as a malformed document (ADR 0008): named on stderr, counted
// against the exit code, and not allowed to discard the images beside it.
func printImages(w io.Writer, logger *log.Logger, source string, images []string) int {
	rejected := 0
	for _, image := range images {
		if err := imageref.Validate(image); err != nil {
			logger.Printf("error: %s: %v", source, err)
			rejected++
			continue
		}
		fmt.Fprintln(w, image)
	}
	return rejected
}
