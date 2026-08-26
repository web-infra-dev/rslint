package linter

import (
	"context"
	"runtime"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// LintPlan is the immutable Phase 1 input shared by native lint execution and
// host-side consumers such as third-party plugin dispatch. It preserves the
// selected files per Program, including syntax-error and zero-rule files needed
// for LintedFileCount, while resolving each eligible file's complete rule set
// exactly once.
type LintPlan struct {
	programs []programLintPlan
}

type programLintPlan struct {
	program *program.Program
	files   []lintFilePlan
}

// lintFilePlan freezes one AST generation, its resolved rules, shared rule
// environment, and checker policy as a coherent execution unit. Parallel
// slices would allow these decisions to drift by index across plan reuse.
type lintFilePlan struct {
	file           *ast.SourceFile
	rules          []rule.ConfiguredRule
	environment    *rule.RuleEnvironment
	hasTypeChecker bool
}

type lintPlanFileRef struct {
	programIndex int
	fileIndex    int
}

// LintTarget is one non-syntax-error file paired with its non-empty configured
// rule set. It is a projection of LintPlan for consumers that do not need
// zero-rule files or per-Program execution metadata.
type LintTarget struct {
	File  *ast.SourceFile
	Rules []rule.ConfiguredRule
}

// PrepareLintPlan collects the Phase 1 target files and resolves their rules.
// Rule resolution uses at most GOMAXPROCS workers unless SingleThreaded is set.
// GetRulesForFile must therefore support concurrent calls whenever the caller
// requests normal parallel execution, matching Consumer.Report's run-scoped
// concurrency requirement. Source-universe validation happens once in the
// internal/program constructor, before the Program can reach this planner.
func PrepareLintPlan(opts RunLinterOptions) (*LintPlan, error) {
	if opts.GetRulesForFile == nil {
		return &LintPlan{}, nil
	}
	// Validate the complete ordered input before target projection invokes a
	// caller-supplied FileFilter for any earlier Program.
	if err := validatePrograms(opts.Programs); err != nil {
		return nil, err
	}
	if opts.ExcludePaths == nil {
		opts.ExcludePaths = utils.ExcludePaths
	}

	plan := &LintPlan{programs: make([]programLintPlan, len(opts.Programs))}
	programOpts := make([]programPlanOptions, len(opts.Programs))
	totalFiles := 0
	for programIndex := range opts.Programs {
		programOpts[programIndex] = programPlanOptionsFor(opts, programIndex)
		programPlan, err := newProgramLintPlan(programOpts[programIndex])
		if err != nil {
			return nil, err
		}
		plan.programs[programIndex] = programPlan
		totalFiles += len(plan.programs[programIndex].files)
	}

	refs := make([]lintPlanFileRef, 0, totalFiles)
	for programIndex, programPlan := range plan.programs {
		for fileIndex := range programPlan.files {
			refs = append(refs, lintPlanFileRef{
				programIndex: programIndex,
				fileIndex:    fileIndex,
			})
		}
	}

	resolve := func(ref lintPlanFileRef, ctx context.Context) {
		programPlan := &plan.programs[ref.programIndex]
		resolveProgramLintPlanFile(programOpts[ref.programIndex], programPlan, ref.fileIndex, ctx)
	}

	workerCount := min(runtime.GOMAXPROCS(0), len(refs))
	if opts.SingleThreaded || workerCount < 2 {
		ctx := context.Background()
		for _, ref := range refs {
			resolve(ref, ctx)
		}
		return plan, nil
	}

	chunkSize := (len(refs) + workerCount - 1) / workerCount
	work := core.NewWorkGroup(false)
	for worker := range workerCount {
		start := worker * chunkSize
		end := min(start+chunkSize, len(refs))
		if start >= end {
			continue
		}
		chunk := refs[start:end]
		work.Queue(func() {
			ctx := context.Background()
			for _, ref := range chunk {
				resolve(ref, ctx)
			}
		})
	}
	work.RunAndWait()
	return plan, nil
}

func newProgramLintPlan(opts programPlanOptions) (programLintPlan, error) {
	if err := validateProgram(opts.Program); err != nil {
		return programLintPlan{}, err
	}
	files := collectFilesToLint(opts)
	filePlans := make([]lintFilePlan, len(files))
	for index, file := range files {
		filePlans[index].file = file
	}
	return programLintPlan{
		program: opts.Program,
		files:   filePlans,
	}, nil
}

func prepareProgramLintPlan(opts programPlanOptions) (programLintPlan, error) {
	plan, err := newProgramLintPlan(opts)
	if err != nil {
		return programLintPlan{}, err
	}
	ctx := context.Background()
	for fileIndex := range plan.files {
		resolveProgramLintPlanFile(opts, &plan, fileIndex, ctx)
	}
	return plan, nil
}

func resolveProgramLintPlanFile(opts programPlanOptions, plan *programLintPlan, fileIndex int, ctx context.Context) {
	filePlan := &plan.files[fileIndex]
	file := filePlan.file
	if shouldSkipRulesForSyntax(opts, file, ctx) {
		return
	}
	rules := opts.GetRulesForFile(file)
	// Program capability is the only checker gate at this boundary. Callers with
	// a narrower request policy, such as LSP HasTypeInfo=false, filter the
	// configured rule set before planning without inspecting Program adapters.
	filePlan.hasTypeChecker = opts.Program.CanProvideTypeChecker(file)
	if filePlan.hasTypeChecker {
		filePlan.rules = rules
	} else {
		filePlan.rules = rule.FilterNonTypeAwareRules(rules)
	}
	filePlan.environment = firstNativeRuleEnvironment(filePlan.rules)
}

func firstNativeRuleEnvironment(rules []rule.ConfiguredRule) *rule.RuleEnvironment {
	for _, configuredRule := range rules {
		if !configuredRule.IsEslintPluginRule && configuredRule.Environment != nil {
			return configuredRule.Environment
		}
	}
	return nil
}

// Targets returns the plan's plugin-facing projection in stable Program/file
// order. The rule slices are shared immutable plan state and must be read-only.
func (p *LintPlan) Targets() []LintTarget {
	if p == nil {
		return nil
	}
	var targets []LintTarget
	for _, programPlan := range p.programs {
		for _, filePlan := range programPlan.files {
			rules := filePlan.rules
			if len(rules) == 0 {
				continue
			}
			targets = append(targets, LintTarget{File: filePlan.file, Rules: rules})
		}
	}
	return targets
}

func programPlanOptionsFor(opts RunLinterOptions, programIndex int) programPlanOptions {
	programOpts := programPlanOptions{
		Program:         opts.Programs[programIndex],
		ExcludePaths:    opts.ExcludePaths,
		GetRulesForFile: opts.GetRulesForFile,
	}

	if programIndex < len(opts.PerProgramFilter) {
		programOpts.FileFilter = opts.PerProgramFilter[programIndex]
	}
	if opts.TargetFiles != nil {
		programOpts.HasTargetFiles = true
		if programIndex < len(opts.TargetFiles) {
			programOpts.TargetFiles = opts.TargetFiles[programIndex]
		}
	} else {
		programOpts.Scope = opts.Scope
		programOpts.FileFilter = composeOwnedFilter(
			programOpts.FileFilter,
			buildOwnedFileSet(programOpts.Program),
		)
	}
	return programOpts
}
