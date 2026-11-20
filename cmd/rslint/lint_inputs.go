package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
)

func cliWorkingDirectory() (string, error) {
	cwd, err := getCLIWorkingDirectory()
	if err != nil {
		return "", err
	}
	return target.NormalizeNativePath(cwd), nil
}

// IPC-supplied files are exact names. Only raw CLI arguments expand as globs.
func prepareCLIInputs(ctx context.Context, args *lintArgs, exactPaths []string) error {
	if len(exactPaths) == 0 && len(args.Positional) == 0 {
		return nil
	}
	cwd, err := cliWorkingDirectory()
	if err != nil {
		return err
	}
	if len(exactPaths) > 0 {
		args.AllowFiles, args.AllowDirs = classifyPaths(cwd, exactPaths)
		args.Positional = nil
		return nil
	}
	selection, err := target.SelectNativeInputs(ctx, args.FS, cwd, args.Positional)
	if err != nil {
		return err
	}
	args.AllowFiles = append(selection.Files, selection.Unmatched...)
	args.AllowDirs = selection.Directories
	args.GlobRoots = rslintconfig.ResolveDirectoryIdentities(selection.GlobRoots, args.FS)
	args.ScanDirectories = selection.ScanDirectories
	return nil
}

// Windows resolves drive-relative and current-drive-rooted names itself.
func classifyPaths(cwd string, paths []string) (files []string, dirs []string) {
	for _, p := range paths {
		if filepath.VolumeName(p) == "" && (p == "" || !os.IsPathSeparator(p[0])) {
			p = filepath.Join(cwd, p)
		}
		absPath, err := filepath.Abs(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "rslint: filepath.Abs(%q) failed: %v\n", p, err)
			continue
		}
		normalized := target.NormalizeNativePath(absPath)
		info, statErr := os.Stat(normalized)
		if statErr == nil && info.IsDir() {
			dirs = append(dirs, normalized)
		} else {
			files = append(files, normalized)
		}
	}
	return files, dirs
}
