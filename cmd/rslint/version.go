package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"

	"github.com/web-infra-dev/rslint/internal/buildinfo"
)

func runVersion(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("rslint --version", flag.ContinueOnError)
	flags.SetOutput(stderr)
	asJSON := flags.Bool("json", false, "print build metadata as JSON")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(stderr, "usage: rslint --version [--json]")
		return 2
	}
	info := buildinfo.Current()
	var err error
	if *asJSON {
		err = json.NewEncoder(stdout).Encode(info)
	} else {
		_, err = fmt.Fprintln(stdout, info)
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}
