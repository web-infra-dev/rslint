package linter

import (
	"context"
	"runtime"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// LintPlan is the immutable Phase 1 input shared by native lint execution and
// host-side consumers such as third-party plugin dispatch. It preserves the
// selected files per Program, including syntax-error and zero-rule files needed
// for LintedFileCount, while resolving each eligible file's complete rule set
// exactly once.
type LintPlan struct {
	programs   []programLintPlan
	standalone []programLintPlan
}

type programLintPlan struct {
	program     *compiler.Program
	runtime     rule.SourceRuntime
	sourceFiles []*ast.SourceFile
	files       []*ast.SourceFile
	rules       [][]ConfiguredRule
}

type lintPlanFileRef struct {
	standalone   bool
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
	// concurrency requirement. It rejects a standalone source set that contains
	// different AST objects for one non-empty ts-go Path, because its Runtime can
	// resolve that identity to only one object.
func PrepareLintPlan(opts RunLinterOptions) (*LintPlan, error) {
	if opts.GetRulesForFile == nil {
		return &LintPlan{}, nil
	}
	if opts.ExcludePaths == nil {
		opts.ExcludePaths = utils.ExcludePaths
	}

	plan := &LintPlan{
		programs:   make([]programLintPlan, len(opts.Programs)),
		standalone: make([]programLintPlan, len(opts.Standalone)),
	}
	programOpts := make([]runProgramOptions, len(opts.Programs))
	standaloneOpts := make([]runProgramOptions, len(opts.Standalone))
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
	for sourceSetIndex := range opts.Standalone {
		standaloneOpts[sourceSetIndex] = runStandaloneOptionsFor(opts, sourceSetIndex, nil)
		sourceSetPlan, err := newProgramLintPlan(standaloneOpts[sourceSetIndex])
		if err != nil {
			return nil, err
		}
		plan.standalone[sourceSetIndex] = sourceSetPlan
		totalFiles += len(plan.standalone[sourceSetIndex].files)
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
	for sourceSetIndex, sourceSetPlan := range plan.standalone {
		for fileIndex := range sourceSetPlan.files {
			refs = append(refs, lintPlanFileRef{
				standalone:   true,
				programIndex: sourceSetIndex,
				fileIndex:    fileIndex,
			})
		}
	}

	resolve := func(ref lintPlanFileRef, ctx context.Context) {
		if ref.standalone {
			sourceSetPlan := &plan.standalone[ref.programIndex]
			resolveProgramLintPlanFile(standaloneOpts[ref.programIndex], sourceSetPlan, ref.fileIndex, ctx)
			return
		}
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
	var sourceFiles []*ast.SourceFile
	var files []*ast.SourceFile
	if opts.Standalone {
		if opts.SourceFilesValidated {
			sourceFiles = opts.SourceFiles
		} else {
			var err error
			sourceFiles, err = validateStandaloneSourceSet(opts.SourceFiles, opts.Runtime)
			if err != nil {
				return programLintPlan{}, err
			}
		}
		files = collectStandaloneFilesToLint(sourceFiles, opts.ExcludePaths)
	} else {
		files = collectFilesToLint(opts)
	}
	return programLintPlan{
		program:     opts.Program,
		runtime:     opts.Runtime,
		sourceFiles: sourceFiles,
		files:       files,
		rules:       make([][]ConfiguredRule, len(files)),
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
	// Standalone source sets never have a Program or TypeChecker. Enforce that
	// invariant here instead of relying on callers to encode it indirectly as a
	// non-nil, empty TypeInfoFiles set: nil is the public option's zero value and
	// otherwise means "do not filter" for Program-backed files.
	if opts.Standalone {
		plan.rules[fileIndex] = FilterNonTypeAwareRules(rules)
		return
	}
	plan.rules[fileIndex] = filterRulesForTypeInfo(
		rules,
		file.FileName(),
		opts.TypeInfoFiles,
	)
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
	for _, sourceSetPlan := range p.standalone {
		for fileIndex, file := range sourceSetPlan.files {
			rules := sourceSetPlan.rules[fileIndex]
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

func runStandaloneOptionsFor(opts RunLinterOptions, sourceSetIndex int, prepared *programLintPlan) runProgramOptions {
	sourceSet := opts.Standalone[sourceSetIndex]
	return runProgramOptions{
		Standalone:           true,
		SourceFiles:          sourceSet.Files,
		Runtime:              sourceSet.Runtime,
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
}
