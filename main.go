package main

import (
	"log"
	"os"

	"github.com/mpv/kir/cmd"
)

func init() {
	log.SetFlags(0) // Disable timestamps in log output
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: kir <file_path> [<file_path_2> ...] or kir -")
		return
	}

	cmd.Execute(os.Args[1:])
}
