package main

import (
	"os"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

// TestMain registers the real kir program as a command testscript can invoke.
// testscript runs it as a subprocess, so the .txtar scripts under
// testdata/script exercise the true end-to-end CLI contract — argv and stdin
// in, and stdout, stderr, and the process exit code out — against the actual
// built binary rather than an in-process function call.
func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		// main() reads os.Args and calls os.Exit via log.Fatal on the failure
		// paths; reaching the return means it finished cleanly (exit 0).
		"kir": func() int { main(); return 0 },
	}))
}

func TestScripts(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata/script",
	})
}
