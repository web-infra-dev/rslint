package main

import (
	"context"

	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/suppressions"
)

func applyLintSuppressions(
	ctx context.Context,
	baseline *suppressions.File,
	files []linter.LintedFile,
	targets []target.File,
	diagnostics []rule.RuleDiagnostic,
	plugin linter.PluginDispatchRecord,
	args lintArgs,
) ([]rule.RuleDiagnostic, string) {
	if plugin.DispatchError != nil && args.Suppressions.Mode != suppressions.Check {
		return diagnostics, "Cannot update suppressions because plugin lint dispatch failed."
	}
	fallback := args.ConfigCatalog.FilesWithConfigFallback(targets)
	incompleteFiles := make(map[string]bool)
	failedRules := make(map[string][]string)
	for _, notice := range plugin.Notices {
		if notice.Kind != linter.EslintPluginUnverifiedRule {
			incompleteFiles[notice.FilePath] = true
		}
		if notice.RuleName != "" {
			failedRules[notice.FilePath] = append(failedRules[notice.FilePath], notice.RuleName)
		}
	}
	// A file-wide update needs a usable file result. Recording named rules
	// can still use their individual counts, including completed native rules.
	wholeFileUpdate := args.Suppressions.Mode == suppressions.Prune ||
		(args.Suppressions.Mode == suppressions.Record && args.Suppressions.Rules == nil)
	checked := make([]suppressions.FileResult, 0, len(files))
	for _, file := range files {
		if file.SyntaxError || fallback[file.Path] || (wholeFileUpdate && incompleteFiles[file.Path]) {
			continue
		}
		skipped := file.SkippedErrorRules
		if failed := failedRules[file.Path]; len(failed) > 0 {
			skipped = append(failed, skipped...)
		}
		checked = append(checked, suppressions.FileResult{Path: file.Path, SkippedRules: skipped})
	}
	var failure string
	switch {
	case len(args.ConfigCatalog.Failures) > 0:
		failure = "A lint configuration failed to load. Files without their intended configuration keep their original suppressions."
	case len(incompleteFiles) > 0 || plugin.DispatchError != nil:
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
