package main

import (
	"context"
	"os"
	"syscall"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
)

// Prepare raw CLI patterns only. IPC-provided files are exact paths and never
// enter this function. Missing patterns retain the existing literal skip policy.
func prepareCLIInputs(ctx context.Context, args *lintArgs) error {
	if len(args.Positional) == 0 && (len(args.AllowFiles) > 0 || len(args.AllowDirs) > 0) {
		return nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	// Go honors a valid shell PWD on Unix. Keep it for target selection and
	// config ownership, but retain the operands if the OS names cwd differently.
	var onMatch func(target.InputMatch)
	if native, err := syscall.Getwd(); err == nil && native != cwd {
		args.inputOrigins.cwd = tspath.NormalizePath(native)
		onMatch = func(match target.InputMatch) {
			args.inputOrigins.matches = append(args.inputOrigins.matches, match)
		}
	}
	if len(args.Positional) == 0 {
		if onMatch != nil {
			onMatch(target.InputMatch{Operand: ".", Path: tspath.NormalizePath(cwd), Directory: true})
		}
		return nil
	}
	selection, err := target.SelectInputsWithMatches(ctx, args.FS, cwd, args.Positional, onMatch)
	if err != nil {
		return err
	}
	args.AllowFiles = append(selection.Files, selection.Unmatched...)
	args.AllowDirs = selection.Directories
	args.GlobRoots = rslintconfig.ResolveDirectoryIdentities(selection.GlobRoots, args.FS)
	args.ScanDirectories = selection.ScanDirectories
	return nil
}

type cliInputOrigins struct {
	cwd     string
	matches []target.InputMatch
}

type cliInputPaths struct {
	files       map[string]string
	directories map[string]string
}

// Only a baseline consumer needs to project the selected targets onto the
// spelling of the operands in the OS working directory.
func (origins cliInputOrigins) paths() cliInputPaths {
	if len(origins.matches) == 0 {
		return cliInputPaths{}
	}
	paths := cliInputPaths{files: make(map[string]string), directories: make(map[string]string)}
	// SelectInputs retains the first operand when several select the same spelling.
	for i := len(origins.matches) - 1; i >= 0; i-- {
		match := origins.matches[i]
		operand := tspath.ResolvePath(origins.cwd, match.Operand)
		switch {
		case match.Directory:
			paths.directories[match.Path] = operand
		case match.Files == nil:
			paths.files[match.Path] = operand
		default:
			for _, path := range match.Files {
				paths.files[path] = rebaseInputPath(path, match.Path, operand)
			}
		}
	}
	return paths
}

func (paths cliInputPaths) forFile(path string) string {
	if paths.files == nil {
		return path
	}
	if operand, ok := paths.files[path]; ok {
		return operand
	}
	for dir := tspath.GetDirectoryPath(path); ; {
		if operand, ok := paths.directories[dir]; ok {
			return rebaseInputPath(path, dir, operand)
		}
		parent := tspath.GetDirectoryPath(dir)
		if parent == dir {
			return path
		}
		dir = parent
	}
}

func rebaseInputPath(path, root, operand string) string {
	relative := tspath.GetRelativePathFromDirectory(root, path, tspath.ComparePathsOptions{UseCaseSensitiveFileNames: true})
	return tspath.CombinePaths(operand, relative)
}
