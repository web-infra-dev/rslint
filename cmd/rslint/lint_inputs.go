package main

import (
	"context"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/suppressions"
)

// Prepare raw CLI patterns only. IPC-provided files are exact paths and never
// enter this function. Missing patterns retain the existing literal skip policy.
func prepareCLIInputs(ctx context.Context, args *lintArgs, preflight lintPreflight) error {
	if len(args.Positional) == 0 && (len(args.AllowFiles) > 0 || len(args.AllowDirs) > 0) {
		return nil
	}
	cwd := preflight.cwd
	// Go honors a valid shell PWD on Unix. Keep it for target selection and
	// config ownership, but retain the operands if the OS names cwd differently.
	var onMatch func(target.InputMatch)
	if preflight.baseline != nil {
		if native := tspath.NormalizePath(preflight.baseline.NativeCWD()); native != cwd {
			args.inputOrigins.cwd = native
			onMatch = func(match target.InputMatch) {
				args.inputOrigins.matches = append(args.inputOrigins.matches, match)
			}
		}
	}
	if len(args.Positional) == 0 {
		if onMatch != nil {
			onMatch(target.InputMatch{Operand: ".", Path: cwd, Directory: true})
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

// Keep the requesting operand separate from the filename used to select and
// execute a target. Entries for selected but unexecuted names reserve their
// own baselines. A literal may bypass selection that excludes it from a scan.
func (origins cliInputOrigins) paths(plan target.Plan, directoryRoots target.InputRoots, explicitConfig, caseSensitive bool) map[string]suppressions.FileResult {
	if len(origins.matches) == 0 && len(plan.SkippedAliases) == 0 {
		return nil
	}
	paths := make(map[string]suppressions.FileResult)
	for _, alias := range plan.SkippedAliases {
		paths[alias.Path] = suppressions.FileResult{Path: alias.Path, CoalescedInto: alias.KeptPath}
	}
	if len(origins.matches) == 0 {
		return paths
	}
	selected := make(map[string]target.File, len(plan.Files)+len(plan.SkippedAliases))
	for _, file := range plan.Files {
		selected[file.Path] = file
	}
	for _, alias := range plan.SkippedAliases {
		selected[alias.Path] = alias.File
	}
	appendOperand := func(path, operand, coalescedInto string) {
		result := paths[path]
		result.Path = path
		result.CoalescedInto = coalescedInto
		if !slices.Contains(result.InputPaths, operand) {
			result.InputPaths = append(result.InputPaths, operand)
		}
		paths[path] = result
	}
	add := func(matched, caller, operand string) {
		if _, ok := selected[caller]; ok {
			matched = caller
		}
		if winner := paths[matched].CoalescedInto; winner != "" {
			matched = winner
		}
		// Case or Unicode spellings of the same entry can share a traversal.
		// Retain the operand on both names so the baseline knows whether the
		// skipped name was independently requested.
		if caller == matched {
			appendOperand(caller, operand, "")
			return
		}
		appendOperand(caller, operand, matched)
		if suppressions.SamePathSpelling(caller, matched) {
			appendOperand(matched, operand, "")
		}
	}
	if len(origins.matches) == 1 {
		match := origins.matches[0]
		operand := tspath.ResolvePath(origins.cwd, match.Operand)
		for path := range selected {
			add(path, path, rebaseInputPath(path, match.Path, operand))
		}
		return paths
	}
	type directory struct{ caller, operand string }
	rootID := func(path string) tspath.Path { return tspath.ToPath(path, "", caseSensitive) }
	directories := make(map[tspath.Path][]directory)
	for _, match := range origins.matches {
		operand := tspath.ResolvePath(origins.cwd, match.Operand)
		switch {
		case match.Directory:
			id := rootID(match.Path)
			directories[id] = append(directories[id], directory{match.Path, operand})
		case match.Files == nil:
			if _, ok := selected[match.Path]; ok {
				add(match.Path, match.Path, operand)
			}
		default:
			for _, path := range match.Files {
				file, ok := selected[path]
				if ok && file.Match() != nil && file.Match().Selected() &&
					slices.ContainsFunc(origins.configRoots.For(file.ConfigDirectory, path), func(root string) bool {
						return suppressions.SamePathSpelling(root, match.Path)
					}) {
					add(path, path, rebaseInputPath(path, match.Path, operand))
				}
			}
		}
	}
	for _, file := range selected {
		for _, scopeRoot := range directoryRoots.For(file.ConfigDirectory, file.Path) {
			// A child config starts a new target walk. Config discovery retained
			// the original operands whose traversal reached that boundary.
			roots := origins.configRoots.For(file.ConfigDirectory, scopeRoot)
			if explicitConfig {
				// An explicit config walks the requested directory directly.
				roots = []string{scopeRoot}
			}
			for _, root := range roots {
				for _, dir := range directories[rootID(root)] {
					if suppressions.SamePathSpelling(root, dir.caller) {
						caller := rebaseInputPath(file.Path, root, dir.caller)
						add(file.Path, caller, rebaseInputPath(file.Path, root, dir.operand))
					}
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
