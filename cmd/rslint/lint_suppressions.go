package main

import (
	"context"

	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/suppressions"
)

func applyLintSuppressions(
	ctx context.Context,
	baseline *suppressions.File,
	observation linter.ObservationResult,
	diagnostics []rule.RuleDiagnostic,
	records []linter.PluginDispatchRecord,
	args lintArgs,
) ([]rule.RuleDiagnostic, string) {
	failedRules := make(map[string][]string)
	var dispatchFailed, missingFile bool
	for _, record := range records {
		if record.Observation != observation.Index {
			continue
		}
		dispatchFailed = dispatchFailed || record.DispatchError != nil
		for _, notice := range record.Notices {
			switch notice.Kind {
			case linter.EslintPluginMissingFileResult:
				missingFile = true
			case linter.EslintPluginUnconfiguredDiagnostic:
				failedRules[notice.FilePath] = append(failedRules[notice.FilePath], notice.RuleName)
			}
		}
	}
	for _, d := range diagnostics {
		if d.Origin == rule.DiagnosticOriginPluginFailure {
			failedRules[d.FilePath] = append(failedRules[d.FilePath], d.RuleName)
		}
	}
	incomplete := dispatchFailed || missingFile || len(failedRules) > 0
	if incomplete && args.Suppressions.Mode != suppressions.Check {
		return diagnostics, "Cannot update suppressions because plugin lint results are incomplete."
	}
	files := observation.Native.Files
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
	if missingFile {
		return visible, "Cannot verify suppressions because a plugin file result is missing."
	}
	// Failed plugin runs already report their errors, and pruning has to wait
	// until all plugin results are available.
	if unused && !incomplete && !args.PassOnUnpruned {
		if len(args.ConfigCatalog.Failures) > 0 {
			return visible, "A lint configuration failed to load. Fix the configuration and rerun rslint to check for unused suppressions."
		}
		return visible, "Unused suppressions remain. Run with --prune-suppressions to remove them, or --pass-on-unpruned-suppressions to allow them."
	}
	return visible, ""
}
