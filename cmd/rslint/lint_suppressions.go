package main

import (
	"context"
	"slices"

	configLint "github.com/web-infra-dev/rslint/internal/config/lint"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/suppressions"
)

// Lint runs once, but rules and their options can differ between the two names.
func mayRecordAlias(options suppressions.Options, resolver *configLint.Resolver, kept, discarded target.File) bool {
	if options.Mode != suppressions.Record || (options.Rules != nil && len(options.Rules) == 0) {
		return false
	}
	for _, file := range [...]target.File{kept, discarded} {
		resolved, ok := resolver.ResolveTarget(file)
		if !ok || slices.ContainsFunc(resolved.EnabledRules, func(r rule.ConfiguredRule) bool {
			return r.Severity == rule.SeverityError && options.SelectsRule(r.Name)
		}) {
			return true
		}
	}
	return false
}

func applyLintSuppressions(
	ctx context.Context,
	baseline *suppressions.File,
	observation linter.ObservationResult,
	diagnostics []rule.RuleDiagnostic,
	records []linter.PluginDispatchRecord,
	args lintArgs,
	commitSources func() error,
) ([]rule.RuleDiagnostic, string) {
	defer baseline.DiscardUpdate()
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
		// Independent native fixes can still be published after checking the
		// current alias counts. Incomplete plugin counts must stay on disk.
		if commitSources != nil {
			if err := baseline.BeginUpdate(); err != nil {
				return diagnostics, err.Error()
			}
			if err := commitSources(); err != nil {
				return diagnostics, err.Error()
			}
		}
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
		checked = append(checked, suppressions.FileResult{
			Path:         file.Path,
			SkippedRules: skipped,
		})
	}
	visible, unused, err := baseline.Prepare(ctx, checked, diagnostics)
	if err == nil && commitSources != nil {
		err = commitSources()
	}
	if err == nil {
		err = baseline.CommitUpdate()
	}
	if err != nil {
		return visible, err.Error()
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
