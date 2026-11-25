package main

import (
	"context"
	"slices"

	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/suppressions"
)

func applyLintSuppressions(
	ctx context.Context,
	baseline *suppressions.File,
	observation linter.ObservationResult,
	diagnostics []rule.RuleDiagnostic,
	args lintArgs,
) ([]rule.RuleDiagnostic, string) {
	pluginFailures, dispatchErr := observation.PluginFailures()
	if args.ConfigCatalog.TargetConfigFailure {
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
