package main

import (
	"os"

	"github.com/glycerine/tlago"
	"github.com/glycerine/tlago/tlc"
)

func main() {
	os.Exit(run())
}

func run() int {
	defer tlc.CleanupDistributedFiles()
	return tlago.RunCLI(os.Args[1:], os.Stdout, os.Stderr)
}
