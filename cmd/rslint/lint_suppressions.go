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
	fallback := args.ConfigCatalog.FilesWithConfigFallback(targets)
	failedRules := make(map[string][]string)
	protocolFailure := plugin.DispatchError != nil
	for _, notice := range plugin.Notices {
		if notice.Kind != linter.EslintPluginUnverifiedRule {
			protocolFailure = true
		}
		if notice.RuleName != "" {
			failedRules[notice.FilePath] = append(failedRules[notice.FilePath], notice.RuleName)
		}
	}
	if protocolFailure && args.Suppressions.Mode != suppressions.Check {
		return diagnostics, "Cannot update suppressions because plugin lint results are incomplete."
	}
	checked := make([]suppressions.FileResult, 0, len(files))
	for _, file := range files {
		if file.SyntaxError || fallback[file.Path] {
			continue
		}
		skipped := file.SkippedErrorRules
		if failed := failedRules[file.Path]; len(failed) > 0 {
			skipped = append(failed, skipped...)
		}
		checked = append(checked, suppressions.FileResult{Path: file.Path, SkippedRules: skipped})
	}
	const configFailure = "A lint configuration failed to load. Files using a fallback configuration keep their original suppressions."
	if len(fallback) > 0 && len(checked) == 0 {
		return diagnostics, configFailure
	}
	visible, unused, err := baseline.Apply(ctx, checked, diagnostics)
	if err != nil {
		return diagnostics, err.Error()
	}
	if len(fallback) > 0 {
		return visible, configFailure
	}
	if protocolFailure {
		return visible, "Cannot verify suppressions because plugin lint results are incomplete."
	}
	if unused && !args.PassOnUnpruned {
		return visible, "Unused suppressions remain. Run with --prune-suppressions to remove them, or --pass-on-unpruned-suppressions to allow them."
	}
	return visible, ""
}
