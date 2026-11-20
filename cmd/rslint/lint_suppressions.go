package main

import (
	"cmp"
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
	diagnostics []rule.RuleDiagnostic,
	records []linter.PluginDispatchRecord,
	args lintArgs,
	targets []target.File,
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
			case linter.EslintPluginUnconfiguredDiagnostic, linter.EslintPluginUnverifiedRule:
				ids := failedRules[notice.FilePath]
				if !slices.Contains(ids, notice.RuleName) {
					failedRules[notice.FilePath] = append(ids, notice.RuleName)
				}
			}
		}
	}
	incomplete := dispatchFailed || missingFile || len(failedRules) > 0 || slices.ContainsFunc(diagnostics, func(d rule.RuleDiagnostic) bool {
		return d.Origin == rule.DiagnosticOriginPluginFailure
	})
	if incomplete && args.Suppressions.Mode != suppressions.Check {
		return diagnostics, "Cannot update suppressions because plugin lint results are incomplete."
	}
	// Use the identities that target selection used for deduplication.
	canonicalPaths := make(map[string]string)
	for _, target := range targets {
		if target.Path != target.CanonicalPath {
			canonicalPaths[target.Path] = target.CanonicalPath
		}
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
		checked = append(checked, suppressions.FileResult{
			Path:          file.Path,
			CanonicalPath: cmp.Or(canonicalPaths[file.Path], file.Path),
			SkippedRules:  skipped,
		})
	}
	visible, unused, err := baseline.Apply(ctx, checked, diagnostics)
	if err != nil {
		return diagnostics, err.Error()
	}
	if unused && !args.PassOnUnpruned {
		if len(args.ConfigCatalog.Failures) > 0 {
			return visible, "A lint configuration failed to load. Fix the configuration and rerun rslint to check for unused suppressions."
		}
		if incomplete {
			return visible, "Unused suppressions remain. Fix the plugin failure before running --prune-suppressions."
		}
		return visible, "Unused suppressions remain. Run with --prune-suppressions to remove them, or --pass-on-unpruned-suppressions to allow them."
	}
	if missingFile {
		return visible, "Cannot verify suppressions because a plugin file result is missing."
	}
	return visible, ""
}
