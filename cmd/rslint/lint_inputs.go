package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"golang.org/x/text/unicode/norm"
)

// Go honors the shell's PWD. On macOS this can differ from the on-disk cwd
// in case or Unicode spelling. Keep deliberate symlinks visible to rules.
func cliWorkingDirectory() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if runtime.GOOS == "darwin" {
		clean := filepath.Clean(cwd)
		native, err := syscall.Getwd()
		if err == nil && native != clean && strings.EqualFold(norm.NFC.String(native), norm.NFC.String(clean)) {
			if resolved, err := filepath.EvalSymlinks(cwd); err == nil && resolved == clean {
				cwd = native
			}
		}
	}
	return tspath.NormalizePath(cwd), nil
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
		normalized := tspath.NormalizePath(absPath)
		info, statErr := os.Stat(absPath)
		if statErr == nil && info.IsDir() {
			dirs = append(dirs, normalized)
		} else {
			files = append(files, normalized)
		}
	}
	return files, dirs
}
