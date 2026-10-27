package main

import (
	"context"
	"os"

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
	selection, err := target.SelectInputs(ctx, args.FS, cwd, args.Positional)
	if err != nil {
		return err
	}
	args.AllowFiles = append(selection.Files, selection.Unmatched...)
	args.AllowDirs = selection.Directories
	args.GlobRoots = rslintconfig.ResolveDirectoryIdentities(selection.GlobRoots, args.FS)
	args.ScanDirectories = selection.ScanDirectories
	return nil
}
