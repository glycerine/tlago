package main

import (
	"os"

	"github.com/glycerine/tlago"
)

func main() {
	os.Exit(tlago.RunCLI(os.Args[1:], os.Stdout, os.Stderr))
}
