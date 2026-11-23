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
	files []linter.LintedFile,
	diagnostics []rule.RuleDiagnostic,
	plugin linter.PluginDispatchRecord,
	args lintArgs,
) ([]rule.RuleDiagnostic, string) {
	updating := args.Suppressions.Mode != suppressions.Check
	configFailed := len(args.ConfigCatalog.Failures) > 0
	if updating && configFailed {
		return diagnostics, "Cannot update suppressions because a lint configuration failed to load."
	}
	failedRules := make(map[string][]string)
	missingFile := false
	for _, notice := range plugin.Notices {
		switch notice.Kind {
		case linter.EslintPluginMissingFileResult:
			missingFile = true
		case linter.EslintPluginUnconfiguredDiagnostic, linter.EslintPluginUnverifiedRule:
			failedRules[notice.FilePath] = append(failedRules[notice.FilePath], notice.RuleName)
		}
	}
	pluginFailed := slices.ContainsFunc(diagnostics, func(d rule.RuleDiagnostic) bool {
		return d.Origin == rule.DiagnosticOriginPluginFailure
	})
	incomplete := pluginFailed || plugin.DispatchError != nil || len(plugin.Notices) != 0
	if updating && incomplete {
		return diagnostics, "Cannot update suppressions because plugin lint results are incomplete."
	}
	checked := make([]suppressions.FileResult, 0, len(files))
	for _, file := range files {
		if file.SyntaxError {
			continue
		}
		skipped := file.SkippedErrorRules
		if failed := failedRules[file.Path]; len(failed) > 0 {
			skipped = append(failed, skipped...)
		}
		checked = append(checked, suppressions.FileResult{Path: file.Path, SkippedRules: skipped})
	}
	visible, unused, err := baseline.Apply(ctx, checked, diagnostics)
	if err != nil {
		return diagnostics, err.Error()
	}
	if unused && !args.PassOnUnpruned {
		if configFailed {
			return visible, "A lint configuration failed to load. Fix the configuration and rerun rslint to check for unused suppressions."
		}
		if incomplete {
			return visible, "Unused suppressions remain. Fix the plugin failure before running --prune-suppressions."
		}
		return visible, "Unused suppressions remain. Run with --prune-suppressions to remove them, or --pass-on-unpruned-suppressions to allow them."
	}
	if missingFile || (incomplete && !pluginFailed) {
		return visible, "Cannot verify suppressions because plugin lint results are incomplete."
	}
	return visible, ""
}
