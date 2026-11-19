package main

import (
	"context"
	"os"
	"slices"
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
	if len(args.inputOrigins.matches) > 1 && slices.ContainsFunc(args.inputOrigins.matches, func(match target.InputMatch) bool {
		return match.Directory || match.Files != nil
	}) {
		args.inputOrigins.configRoots = make(target.InputRoots)
	}
	args.AllowFiles = append(selection.Files, selection.Unmatched...)
	args.AllowDirs = selection.Directories
	args.GlobRoots = rslintconfig.ResolveDirectoryIdentities(selection.GlobRoots, args.FS)
	args.ScanDirectories = selection.ScanDirectories
	return nil
}

type cliInputOrigins struct {
	cwd         string
	matches     []target.InputMatch
	configRoots target.InputRoots
}

// Project only the inputs that reached each executed caller-visible path.
// Literal files may bypass selection that excludes the same path from a scan.
func (origins cliInputOrigins) paths(plan target.Plan, directoryRoots target.InputRoots) map[string][]string {
	if len(origins.matches) == 0 {
		return nil
	}
	files := append(slices.Clip(plan.Files), plan.SkippedAliases...)
	paths := make(map[string][]string)
	if len(origins.matches) == 1 {
		match := origins.matches[0]
		operand := tspath.ResolvePath(origins.cwd, match.Operand)
		// The sole operand is the source of every selected path.
		for _, file := range files {
			paths[file.Path] = []string{rebaseInputPath(file.Path, match.Path, operand)}
		}
		return paths
	}
	add := func(path, operand string) {
		if !slices.Contains(paths[path], operand) {
			paths[path] = append(paths[path], operand)
		}
	}
	selected := make(map[string]target.File, len(files))
	for _, file := range files {
		selected[file.Path] = file
	}
	directories := make(map[string][]string)
	for _, match := range origins.matches {
		operand := tspath.ResolvePath(origins.cwd, match.Operand)
		switch {
		case match.Directory:
			directories[match.Path] = append(directories[match.Path], operand)
		case match.Files == nil:
			add(match.Path, operand)
		default:
			for _, path := range match.Files {
				file, ok := selected[path]
				if ok && file.Match() != nil && file.Match().Selected() &&
					slices.Contains(origins.configRoots.For(file.ConfigDirectory, path), match.Path) {
					add(path, rebaseInputPath(path, match.Path, operand))
				}
			}
		}
	}
	for _, file := range files {
		for _, scopeRoot := range directoryRoots.For(file.ConfigDirectory, file.Path) {
			// A child config starts a new target walk. Config discovery retained
			// the original operands whose traversal reached that boundary.
			for _, root := range origins.configRoots.For(file.ConfigDirectory, scopeRoot) {
				for _, operand := range directories[root] {
					add(file.Path, rebaseInputPath(file.Path, root, operand))
				}
			}
		}
	}
	return paths
}

func rebaseInputPath(path, root, operand string) string {
	relative := tspath.GetRelativePathFromDirectory(root, path, tspath.ComparePathsOptions{UseCaseSensitiveFileNames: true})
	return tspath.CombinePaths(operand, relative)
}
