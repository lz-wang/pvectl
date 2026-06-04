package main

import (
	"fmt"
	"os"

	"github.com/lz-wang/pvectl/cmd"
)

var version = "dev"

func main() {
	if err := cmd.Run(os.Args, version); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
