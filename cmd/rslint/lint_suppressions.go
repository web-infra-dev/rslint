package main

import (
	"context"
	"slices"

	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/suppressions"
)

type suppressionStatus uint8

const (
	suppressionOK suppressionStatus = iota
	suppressionUnused
	suppressionError
)

type suppressionOutcome struct {
	diagnostics []rule.RuleDiagnostic
	status      suppressionStatus
	reason      string
}

func (o suppressionOutcome) failed() bool {
	return o.status != suppressionOK
}

func (o suppressionOutcome) message() string {
	switch o.status {
	case suppressionError:
		return o.reason
	case suppressionUnused:
		return "Unused suppressions remain. Run with --prune-suppressions to remove them, or --pass-on-unpruned-suppressions to allow them."
	default:
		return ""
	}
}

func applyLintSuppressions(
	ctx context.Context,
	baseline *suppressions.File,
	observation linter.ObservationResult,
	diagnostics []rule.RuleDiagnostic,
	args lintArgs,
) suppressionOutcome {
	pluginFailures, dispatchErr := observation.PluginFailures()
	if args.ConfigCatalog.TargetConfigFailure {
		if args.Suppressions.Mode == suppressions.Check {
			return suppressionOutcome{
				diagnostics: diagnostics,
				status:      suppressionError,
				reason:      "Cannot apply suppressions because a lint configuration failed to load.",
			}
		}
		return suppressionOutcome{
			diagnostics: diagnostics,
			status:      suppressionError,
			reason:      "Cannot update suppressions because a lint configuration failed to load.",
		}
	}
	if dispatchErr != nil && args.Suppressions.Mode != suppressions.Check {
		return suppressionOutcome{
			diagnostics: diagnostics,
			status:      suppressionError,
			reason:      "Cannot update suppressions because plugin lint dispatch failed.",
		}
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
		return suppressionOutcome{
			diagnostics: diagnostics,
			status:      suppressionError,
			reason:      err.Error(),
		}
	}
	if pluginIncomplete {
		return suppressionOutcome{
			diagnostics: visible,
			status:      suppressionError,
			reason:      "Plugin lint results are incomplete.",
		}
	}
	if unused && !args.PassOnUnpruned {
		return suppressionOutcome{
			diagnostics: visible,
			status:      suppressionUnused,
		}
	}
	return suppressionOutcome{
		diagnostics: visible,
		status:      suppressionOK,
	}
}
