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
	const incomplete = "Cannot update suppressions because plugin lint results are incomplete."
	failedFiles := make(map[string]bool)
	missing := false
	for _, record := range records {
		if record.Observation != observation.Index {
			continue
		}
		if record.DispatchError != nil {
			if args.Suppressions.Mode != suppressions.Check {
				return diagnostics, incomplete
			}
			// A transport error can span more than the synthetic diagnostic's
			// anchor file. Report the original failure and available findings.
			return diagnostics, ""
		}
		for _, notice := range record.Notices {
			failedFiles[notice.FilePath] = true
			missing = missing || notice.Kind == linter.EslintPluginMissingFileResult
		}
	}
	for _, d := range diagnostics {
		if d.Origin == rule.DiagnosticOriginPluginFailure {
			failedFiles[d.FilePath] = true
		}
	}
	if len(failedFiles) > 0 && args.Suppressions.Mode != suppressions.Check {
		return diagnostics, incomplete
	}
	files := observation.Native.Files
	checked := make([]string, 0, len(files))
	for _, file := range files {
		if !file.SyntaxError && !failedFiles[file.Path] {
			checked = append(checked, file.Path)
		}
	}
	visible, unused, err := baseline.Apply(ctx, checked, diagnostics)
	if err != nil {
		return diagnostics, err.Error()
	}
	if missing {
		return visible, "Cannot verify suppressions because a plugin file result is missing."
	}
	if unused && !args.PassOnUnpruned {
		return visible, "Unused suppressions remain. Run with --prune-suppressions to remove them, or --pass-on-unpruned-suppressions to allow them."
	}
	return visible, ""
}
