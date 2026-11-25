package main

import (
	"context"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/discovery"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/suppressions"
)

func applyLintSuppressions(
	ctx context.Context,
	baseline *suppressions.File,
	observation linter.ObservationResult,
	diagnostics []rule.RuleDiagnostic,
	targetPlan target.Plan,
	targetDirectories []rslintconfig.DirectoryIdentity,
	fs vfs.FS,
	args lintArgs,
) ([]rule.RuleDiagnostic, string) {
	pluginFailures, dispatchErr := observation.PluginFailures()
	if configFailureAffectsRun(args.ConfigCatalog, targetPlan, targetDirectories, fs) {
		if args.Suppressions.Mode == suppressions.Check {
			return diagnostics, "Cannot apply suppressions because a lint configuration failed to load."
		}
		return diagnostics, "Cannot update suppressions because a lint configuration failed to load."
	}
	if dispatchErr != nil && args.Suppressions.Mode != suppressions.Check {
		return diagnostics, "Cannot update suppressions because plugin lint dispatch failed."
	}
	pluginIncomplete := dispatchErr != nil
	files := observation.Native.Files
	results := make([]suppressions.FileResult, len(files))
	for index, file := range files {
		failure := pluginFailures[file.Path]
		pluginIncomplete = pluginIncomplete || failure.FileIncomplete
		result := &results[index]
		result.Path = file.Path
		result.Unchecked = file.SyntaxError
		if result.Unchecked {
			continue
		}
		result.Incomplete = failure.FileIncomplete
		result.SkippedRules = file.SkippedErrorRules
		if len(failure.Rules) > 0 {
			result.SkippedRules = slices.Concat(result.SkippedRules, failure.Rules)
		}
	}
	visible, unused, err := baseline.Apply(ctx, results, diagnostics)
	if err != nil {
		return diagnostics, err.Error()
	}
	if pluginIncomplete {
		return visible, "Plugin lint results are incomplete."
	}
	if unused && !args.PassOnUnpruned {
		return visible, "Unused suppressions remain. Run with --prune-suppressions to remove them, or --pass-on-unpruned-suppressions to allow them."
	}
	return visible, ""
}

// configFailureAffectsRun reports whether any failed config discovery candidate
// could govern a selected or requested target (including targets that fell back
// to an ancestor config or were ignored by that fallback). Ancestor failures
// shadowed by a closer valid config for every requested target do not affect
// the run.
func configFailureAffectsRun(
	catalog *discovery.ConfigCatalog,
	targetPlan target.Plan,
	targetDirectories []rslintconfig.DirectoryIdentity,
	fs vfs.FS,
) bool {
	if catalog == nil || catalog.Explicit || len(catalog.Failures) == 0 {
		return false
	}
	caseSensitive := fs == nil || fs.UseCaseSensitiveFileNames()
	opts := tspath.ComparePathsOptions{UseCaseSensitiveFileNames: caseSensitive}
	contains := func(parent, child string) bool {
		return parent != "" && child != "" && tspath.ContainsPath(parent, child, opts)
	}
	strictlyContains := func(parent, child string) bool {
		return contains(parent, child) && tspath.ComparePaths(parent, child, opts) != 0
	}
	canonicalDir := func(dir string) string {
		dir = tspath.NormalizePath(dir)
		if dir == "" || fs == nil {
			return dir
		}
		if realPath := fs.Realpath(dir); realPath != "" {
			return tspath.NormalizePath(realPath)
		}
		return dir
	}

	type loadedOwner struct {
		lexical      string
		canonical    string
		explicitOnly bool
	}
	owners := make([]loadedOwner, 0, len(catalog.Configs))
	for owner := range catalog.Configs {
		lexical := tspath.NormalizePath(owner)
		owners = append(owners, loadedOwner{
			lexical:      lexical,
			canonical:    canonicalDir(lexical),
			explicitOnly: catalog.Scopes[owner].ExplicitOnly,
		})
	}
	nearestOwner := func(path string, useCanonical, automaticOnly bool) string {
		best := ""
		for _, owner := range owners {
			if automaticOnly && owner.explicitOnly {
				continue
			}
			candidate := owner.lexical
			if useCanonical {
				candidate = owner.canonical
			}
			if !contains(candidate, path) {
				continue
			}
			if best == "" || strictlyContains(best, candidate) {
				best = candidate
			}
		}
		return best
	}
	fellBackAcross := func(owner, failDir string) bool {
		return owner == "" || strictlyContains(owner, failDir)
	}

	for _, failure := range catalog.Failures {
		failLex := tspath.NormalizePath(failure.Directory)
		if failLex == "" {
			return true
		}
		failReal := canonicalDir(failLex)

		for _, file := range targetPlan.Files {
			if contains(failLex, file.Path) && fellBackAcross(file.ConfigDirectory, failLex) {
				return true
			}
			if failReal != "" && contains(failReal, file.CanonicalPath) && fellBackAcross(canonicalDir(file.ConfigDirectory), failReal) {
				return true
			}
		}
		for _, file := range catalog.TargetFiles {
			if contains(failLex, file.Path) && fellBackAcross(nearestOwner(file.Path, false, false), failLex) {
				return true
			}
			if failReal != "" && contains(failReal, file.CanonicalPath) && fellBackAcross(nearestOwner(file.CanonicalPath, true, false), failReal) {
				return true
			}
		}
		for _, dir := range targetDirectories {
			if contains(dir.LexicalPath, failLex) ||
				(contains(failLex, dir.LexicalPath) && fellBackAcross(nearestOwner(dir.LexicalPath, false, true), failLex)) {
				return true
			}
			if failReal != "" && (contains(dir.CanonicalPath, failReal) ||
				(contains(failReal, dir.CanonicalPath) && fellBackAcross(nearestOwner(dir.CanonicalPath, true, true), failReal))) {
				return true
			}
		}
	}
	return false
}
