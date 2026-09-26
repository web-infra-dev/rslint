package linter

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// deferredFilePlan retains construction inputs and resolved configuration, never
// an AST. Each execution owns the Program it builds, including derived caches.
type deferredFilePlan struct {
	sources *program.SourceSet
	index   int
	rules   []rule.ConfiguredRule
}

func (p deferredFilePlan) fileName() string { return p.sources.FileNames()[p.index] }

func prepareGenerationLintPlan(ctx context.Context, native NativeGeneration, allowDeferred bool) (*LintPlan, error) {
	opts := PrepareLintPlanOptions{
		Programs:         native.Programs,
		TargetsByProgram: native.TargetsByProgram,
		SingleThreaded:   native.SingleThreaded,
		GetRulesForFile:  native.RulesForFile,
	}
	if len(native.DeferredSources) == 0 {
		return PrepareLintPlanContext(ctx, opts)
	}
	// Validate every construction input before invoking configuration callbacks.
	if len(opts.TargetsByProgram) != len(opts.Programs) {
		return nil, errTargetsByProgramLength
	}
	if err := validatePrograms(opts.Programs); err != nil {
		return nil, err
	}
	if opts.GetRulesForFile == nil {
		return nil, errNilRuleHandler
	}
	for _, sources := range native.DeferredSources {
		if !sources.IsValid() {
			return nil, errors.New("linter: invalid deferred source set")
		}
	}
	opts.Programs = slices.Clone(opts.Programs)
	opts.TargetsByProgram = slices.Clone(opts.TargetsByProgram)
	var deferred []deferredFilePlan
	for _, sources := range native.DeferredSources {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		independent := allowDeferred && native.RulesForPath != nil
		var files []deferredFilePlan
		if independent {
			files = make([]deferredFilePlan, 0, len(sources.FileNames()))
			for index, path := range sources.FileNames() {
				if err := ctx.Err(); err != nil {
					return nil, err
				}
				rules := native.RulesForPath(path)
				for _, configured := range rules {
					if configured.IsEslintPluginRule || (!configured.RequiresTypeInfo && configured.RequiresProgram) {
						independent = false
						break
					}
				}
				if !independent {
					break
				}
				files = append(files, deferredFilePlan{sources: sources, index: index, rules: rules})
			}
		}
		if independent {
			deferred = append(deferred, files...)
			continue
		}
		// A cross-file consumer needs every root, including files on which that
		// rule is disabled. Do not split a source universe by configuration.
		sourceProgram, err := sources.Build()
		if err != nil {
			return nil, fmt.Errorf("linter: build source set: %w", err)
		}
		opts.Programs = append(opts.Programs, sourceProgram)
		opts.TargetsByProgram = append(opts.TargetsByProgram, sources.FileNames())
	}
	plan, err := PrepareLintPlanContext(ctx, opts)
	if err != nil {
		return nil, err
	}
	plan.deferredFiles = deferred
	return plan, nil
}
