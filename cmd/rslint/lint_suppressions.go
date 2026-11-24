package main

import (
	"context"
	"slices"

	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/suppressions"
)

func applyLintSuppressions(
	ctx context.Context,
	baseline *suppressions.File,
	observation linter.ObservationResult,
	targets target.Plan,
	diagnostics []rule.RuleDiagnostic,
	args lintArgs,
) ([]rule.RuleDiagnostic, string) {
	pluginFailures, dispatchErr := observation.PluginFailures()
	if dispatchErr != nil && args.Suppressions.Mode != suppressions.Check {
		return diagnostics, "Cannot update suppressions because plugin lint dispatch failed."
	}
	fallback, configFailed := args.ConfigCatalog.FilesWithConfigFallback(targets)
	pluginIncomplete := dispatchErr != nil
	files := observation.Native.Files
	results := make([]suppressions.FileResult, len(files))
	hasCheckedFile := false
	for index, file := range files {
		failure := pluginFailures[file.Path]
		pluginIncomplete = pluginIncomplete || failure.FileIncomplete
		result := &results[index]
		result.Path = file.Path
		result.Unchecked = file.SyntaxError || fallback[file.Path]
		if result.Unchecked {
			continue
		}
		hasCheckedFile = true
		result.Incomplete = failure.FileIncomplete
		result.SkippedRules = file.SkippedErrorRules
		if len(failure.Rules) > 0 {
			result.SkippedRules = slices.Concat(result.SkippedRules, failure.Rules)
		}
	}
	var failure string
	switch {
	case configFailed:
		failure = "A lint configuration failed to load. Files without their intended configuration keep their original suppressions."
	case pluginIncomplete:
		failure = "Plugin lint results are incomplete."
	}
	if failure != "" && !hasCheckedFile {
		return diagnostics, failure
	}
	visible, unused, err := baseline.Apply(ctx, results, diagnostics)
	if err != nil {
		return diagnostics, err.Error()
	}
	if failure != "" {
		return visible, failure
	}
	if unused && !args.PassOnUnpruned {
		return visible, "Unused suppressions remain. Run with --prune-suppressions to remove them, or --pass-on-unpruned-suppressions to allow them."
	}
	return visible, ""
}
