package linter

import (
	"context"
	"runtime"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/program"
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
	program     *program.Program
	files       []*ast.SourceFile
	rules       [][]ConfiguredRule
	hasTypeInfo []bool
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
	Rules []ConfiguredRule
}

// PrepareLintPlan collects the Phase 1 target files and resolves their rules.
// Rule resolution uses at most GOMAXPROCS workers unless SingleThreaded is set.
// GetRulesForFile must therefore support concurrent calls whenever the caller
// requests normal parallel execution, matching Consumer.Report's run-scoped
// concurrency requirement. Standalone source validation happens once in
// program.NewStandalone, before the Program can reach this planner.
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
	programOpts := make([]runProgramOptions, len(opts.Programs))
	totalFiles := 0
	for programIndex := range opts.Programs {
		programOpts[programIndex] = runProgramOptionsFor(opts, programIndex, nil)
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

func newProgramLintPlan(opts runProgramOptions) (programLintPlan, error) {
	if err := validateProgram(opts.Program); err != nil {
		return programLintPlan{}, err
	}
	files := collectFilesToLint(opts)
	return programLintPlan{
		program:     opts.Program,
		files:       files,
		rules:       make([][]ConfiguredRule, len(files)),
		hasTypeInfo: make([]bool, len(files)),
	}, nil
}

func prepareProgramLintPlan(opts runProgramOptions) (programLintPlan, error) {
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

func resolveProgramLintPlanFile(opts runProgramOptions, plan *programLintPlan, fileIndex int, ctx context.Context) {
	file := plan.files[fileIndex]
	if shouldSkipRulesForSyntax(opts, file, ctx) {
		return
	}
	rules := opts.GetRulesForFile(file)
	if opts.Program.IsStandalone() {
		plan.rules[fileIndex] = FilterNonTypeAwareRules(rules)
		return
	}
	plan.hasTypeInfo[fileIndex] = fileHasTypeInfo(file.FileName(), opts.TypeInfoFiles)
	if plan.hasTypeInfo[fileIndex] {
		plan.rules[fileIndex] = rules
	} else {
		plan.rules[fileIndex] = FilterNonTypeAwareRules(rules)
	}
}

// Targets returns the plan's plugin-facing projection in stable Program/file
// order. The rule slices are shared immutable plan state and must be read-only.
func (p *LintPlan) Targets() []LintTarget {
	if p == nil {
		return nil
	}
	var targets []LintTarget
	for _, programPlan := range p.programs {
		for fileIndex, file := range programPlan.files {
			rules := programPlan.rules[fileIndex]
			if len(rules) == 0 {
				continue
			}
			targets = append(targets, LintTarget{File: file, Rules: rules})
		}
	}
	return targets
}

func runProgramOptionsFor(opts RunLinterOptions, programIndex int, prepared *programLintPlan) runProgramOptions {
	programOpts := runProgramOptions{
		Program:              opts.Programs[programIndex],
		Cwd:                  opts.Cwd,
		ExcludePaths:         opts.ExcludePaths,
		GetRulesForFile:      opts.GetRulesForFile,
		CollectExecutedRules: true,
		SyntaxErrorFiles:     opts.SyntaxErrorFiles,
		SingleThreaded:       opts.SingleThreaded,
		TypeInfoFiles:        opts.TypeInfoFiles,
		Timing:               opts.Timing,
		PreparedPlan:         prepared,
	}
	if prepared != nil {
		return programOpts
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
