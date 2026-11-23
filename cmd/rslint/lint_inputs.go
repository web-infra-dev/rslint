package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"syscall"

	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
)

// Resolve the invocation root once. PWD can name a directory symlink, even
// when the Node host starts the process in the physical working directory.
func (args *lintArgs) invocationRoot() (target.InputRoot, error) {
	if args.inputRoot.Directory == "" {
		cwd, err := syscall.Getwd()
		root := target.InputRoot{Directory: cwd}
		if err == nil && runtime.GOOS == "windows" {
			// Windows Getwd can retain the name of a directory junction.
			root.WindowsVolumeDirectory = cwd
			root.Directory, err = filepath.EvalSymlinks(cwd)
		}
		if err != nil {
			return target.InputRoot{}, err
		}
		args.inputRoot = root
	}
	return args.inputRoot, nil
}

// Exact IPC targets and raw CLI patterns share the same invocation root.
func prepareCLIInputs(ctx context.Context, args *lintArgs, exactFiles []string) error {
	root, err := args.invocationRoot()
	if err != nil {
		return err
	}
	if len(exactFiles) > 0 {
		args.AllowFiles, args.AllowDirs = classifyPaths(root, exactFiles)
		return nil
	}
	if len(args.Positional) == 0 {
		return nil
	}
	selection, err := root.Select(ctx, args.FS, args.Positional)
	if err != nil {
		return err
	}
	args.AllowFiles = append(selection.Files, selection.Unmatched...)
	args.AllowDirs = selection.Directories
	args.GlobRoots = rslintconfig.ResolveDirectoryIdentities(selection.GlobRoots, args.FS)
	args.ScanDirectories = selection.ScanDirectories
	return nil
}

// Exact IPC targets bypass glob interpretation. A missing literal still
// reaches the CLI's file warning.
func classifyPaths(root target.InputRoot, paths []string) (files []string, dirs []string) {
	for _, p := range paths {
		path, err := root.ResolvePath(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "rslint: resolve input %q: %v\n", p, err)
			continue
		}
		info, err := os.Stat(path)
		if err == nil && info.IsDir() {
			dirs = append(dirs, path)
		} else {
			files = append(files, path)
		}
	}
	return files, dirs
}
