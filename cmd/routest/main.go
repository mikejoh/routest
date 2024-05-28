package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mikejoh/routest/internal/buildinfo"
)

type routestOptions struct {
	version bool
}

func main() {
	var routestOpts routestOptions
	flag.BoolVar(&routestOpts.version, "version", false, "Print the version number.")
	flag.Parse()

	if routestOpts.version {
		fmt.Println(buildinfo.Get())
		os.Exit(0)
	}
}
