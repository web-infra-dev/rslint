package linter

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// sourceFilePlan freezes configuration before parsing. Only file-local rules
// enter this plan; rules needing other ASTs keep their complete root group.
type sourceFilePlan struct {
	name  string
	rules []rule.ConfiguredRule
	build func(context.Context, []string) (*program.Program, error)
}

// prepareNativeLintPlan owns source materialization policy. Consumers that
// retain or share ASTs request complete Programs. Otherwise, an entire root
// group must support independent execution before any file is streamed.
func prepareNativeLintPlan(ctx context.Context, native NativeGeneration, retainSources bool) (*LintPlan, error) {
	if native.RulesForPath == nil {
		if len(native.RootGroups) != 0 {
			return nil, errNilRuleHandler
		}
		return nil, nil //nolint:nilnil // A nil plan selects type-check-only or an empty generation.
	}
	programs := slices.Clone(native.Programs)
	targets := slices.Clone(native.TargetsByProgram)
	if len(programs) != len(targets) {
		return nil, errTargetsByProgramLength
	}
	var sources []sourceFilePlan
	resolved := make(map[string][]rule.ConfiguredRule)
	for _, group := range native.RootGroups {
		if group.Build == nil {
			return nil, errors.New("linter: source roots require a builder")
		}
		for _, name := range group.FileNames {
			if name == "" {
				return nil, errors.New("linter: source root path must not be empty")
			}
		}
		stream := !retainSources && !native.TypeCheck
		if stream {
			rules := make([][]rule.ConfiguredRule, len(group.FileNames))
			if err := runSourceTasks(ctx, len(rules), sourceWorkerCount(len(rules), native.SingleThreaded), func(ctx context.Context, index, _ int) error {
				rules[index] = native.RulesForPath(group.FileNames[index])
				return ctx.Err()
			}); err != nil {
				return nil, err
			}
			for index, name := range group.FileNames {
				if _, duplicate := resolved[name]; duplicate {
					return nil, fmt.Errorf("linter: duplicate source root %q", name)
				}
				resolved[name] = rules[index]
				stream = stream && canIsolateSourceFile(rules[index])
			}
		}
		if stream {
			for _, name := range group.FileNames {
				sources = append(sources, sourceFilePlan{
					name: name, rules: resolved[name], build: group.Build,
				})
			}
			continue
		}
		p, err := buildRootProgram(ctx, group.Build, group.FileNames)
		if err != nil {
			return nil, err
		}
		programs = append(programs, p)
		targets = append(targets, group.FileNames)
	}
	plan, err := PrepareLintPlanContext(ctx, PrepareLintPlanOptions{
		Programs: programs, TargetsByProgram: targets, SingleThreaded: native.SingleThreaded,
		GetRulesForFile: func(file *ast.SourceFile) []rule.ConfiguredRule {
			if rules, ok := resolved[file.FileName()]; ok {
				return rules
			}
			return native.RulesForPath(file.FileName())
		},
	})
	if err != nil {
		return nil, err
	}
	plan.sources = sources
	return plan, nil
}

func canIsolateSourceFile(rules []rule.ConfiguredRule) bool {
	for _, configured := range rules {
		if configured.IsEslintPluginRule || (!configured.RequiresTypeInfo && !configured.SupportsFileIsolation) {
			return false
		}
	}
	return true
}

func buildRootProgram(ctx context.Context, build func(context.Context, []string) (*program.Program, error), names []string) (*program.Program, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	p, err := build(ctx, names)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := validateProgram(p); err != nil {
		return nil, err
	}
	files := p.SourceFiles()
	if len(files) != len(names) {
		return nil, errors.New("linter: root builder changed the source universe")
	}
	for index, file := range files {
		if file == nil || file.FileName() != names[index] || p.CanProvideTypeChecker(file) {
			return nil, fmt.Errorf("linter: root builder must preserve source-only root %q", names[index])
		}
	}
	return p, nil
}
