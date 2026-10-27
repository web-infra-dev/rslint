package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
)

// Prepare raw CLI patterns only. IPC-provided files are exact paths and never
// enter this function. Missing patterns retain the existing literal skip policy.
func prepareCLIInputs(ctx context.Context, args *lintArgs) error {
	if len(args.Positional) == 0 {
		return nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	args.AllowFiles, args.AllowDirs = nil, nil
	type cliInput struct {
		path      string
		directory bool
		glob      bool
	}
	inputs := make([]cliInput, len(args.Positional))
	var patterns []string
	for index, input := range args.Positional {
		absolute, err := filepath.Abs(input)
		if err != nil {
			return err
		}
		absolute = tspath.NormalizePath(absolute)
		inputs[index].path = absolute
		info, statErr := os.Stat(absolute)
		if statErr == nil && info.IsDir() {
			inputs[index].directory = true
			continue
		}
		if statErr != nil && strings.ContainsAny(input, "*?[{") {
			inputs[index].glob = true
			patterns = append(patterns, input)
		}
	}
	expansions, err := target.ExpandGlobs(ctx, args.FS, cwd, patterns)
	if err != nil {
		return err
	}
	literals := make(map[string]bool)
	globIndex := 0
	for _, input := range inputs {
		if input.directory {
			args.AllowDirs = append(args.AllowDirs, input.path)
			continue
		}
		if !input.glob {
			args.AllowFiles = append(args.AllowFiles, input.path)
			literals[input.path] = true
			continue
		}
		expansion := expansions[globIndex]
		globIndex++
		if len(expansion.Files) == 0 {
			args.AllowFiles = append(args.AllowFiles, input.path)
			literals[input.path] = true
			continue
		}
		args.GlobRoots = append(args.GlobRoots, rslintconfig.ResolveDirectoryIdentities([]string{expansion.Root}, args.FS)...)
		if args.ScanDirectories == nil {
			args.ScanDirectories = make(map[string][]string)
		}
		for _, file := range expansion.Files {
			args.AllowFiles = append(args.AllowFiles, file)
			args.ScanDirectories[file] = append(args.ScanDirectories[file], expansion.Root)
		}
	}
	for file := range literals {
		delete(args.ScanDirectories, file)
	}
	return nil
}
