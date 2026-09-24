// Command generate_config_schema updates the model-backed portions of
// rslint-schema.json from the Go configuration types. Human-authored schema
// prose and conditional shapes remain in the checked-in document.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/web-infra-dev/rslint/internal/config/schemagen"
)

func main() {
	filePath := flag.String("file", "rslint-schema.json", "path to the root configuration schema")
	check := flag.Bool("check", false, "fail when the schema is not up to date")
	flag.Parse()

	existing, err := os.ReadFile(*filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read %s: %v\n", *filePath, err)
		os.Exit(1)
	}
	generated, result, err := schemagen.Merge(existing)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate schema: %v\n", err)
		os.Exit(1)
	}

	equal, err := schemagen.CanonicalEqual(existing, generated)
	if err != nil {
		fmt.Fprintf(os.Stderr, "compare schema: %v\n", err)
		os.Exit(1)
	}
	if *check {
		if !equal {
			fmt.Fprintf(os.Stderr, "%s is out of date; run: pnpm generate:config-schema\n", *filePath)
			os.Exit(1)
		}
		fmt.Printf("%s is up to date\n", *filePath)
		return
	}
	if equal {
		fmt.Printf("%s is up to date\n", *filePath)
		return
	}
	if err := os.WriteFile(*filePath, generated, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write %s: %v\n", *filePath, err)
		os.Exit(1)
	}
	fmt.Printf("updated %s (added %d, removed %d, updated %d)\n", *filePath, len(result.Added), len(result.Removed), len(result.Updated))
}
