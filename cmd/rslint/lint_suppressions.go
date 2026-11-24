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
	checked := make([]suppressions.FileResult, 0, len(files))
	for _, file := range files {
		failure := pluginFailures[file.Path]
		pluginIncomplete = pluginIncomplete || failure.FileIncomplete
		if file.SyntaxError || fallback[file.Path] {
			continue
		}
		skipped := file.SkippedErrorRules
		if len(failure.Rules) > 0 {
			skipped = slices.Concat(skipped, failure.Rules)
		}
		checked = append(checked, suppressions.FileResult{
			Path:         file.Path,
			SkippedRules: skipped,
			Incomplete:   failure.FileIncomplete,
		})
	}
	var failure string
	switch {
	case configFailed:
		failure = "A lint configuration failed to load. Files without their intended configuration keep their original suppressions."
	case pluginIncomplete:
		failure = "Plugin lint results are incomplete."
	}
	if failure != "" && len(checked) == 0 {
		return diagnostics, failure
	}
	visible, unused, err := baseline.Apply(ctx, checked, diagnostics)
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
