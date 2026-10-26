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
	literals := make(map[string]bool)
	for _, input := range args.Positional {
		absolute, err := filepath.Abs(input)
		if err != nil {
			return err
		}
		absolute = tspath.NormalizePath(absolute)
		info, statErr := os.Stat(absolute)
		if statErr == nil && info.IsDir() {
			args.AllowDirs = append(args.AllowDirs, absolute)
			continue
		}
		if statErr == nil || !strings.ContainsAny(input, "*?[{") {
			args.AllowFiles = append(args.AllowFiles, absolute)
			literals[absolute] = true
			continue
		}
		root, matches, err := target.ExpandGlob(ctx, args.FS, cwd, input)
		if err != nil {
			return err
		}
		if len(matches) == 0 {
			args.AllowFiles = append(args.AllowFiles, absolute)
			literals[absolute] = true
			continue
		}
		args.GlobRoots = append(args.GlobRoots, rslintconfig.ResolveDirectoryIdentities([]string{root}, args.FS)...)
		if args.ScanDirectories == nil {
			args.ScanDirectories = make(map[string][]string)
		}
		for _, file := range matches {
			args.AllowFiles = append(args.AllowFiles, file)
			args.ScanDirectories[file] = append(args.ScanDirectories[file], root)
		}
	}
	for file := range literals {
		delete(args.ScanDirectories, file)
	}
	return nil
}
