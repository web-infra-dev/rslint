package linter

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var (
	errNilRuleHandler         = errors.New("linter: GetRulesForFile must not be nil")
	errTargetNotInProgram     = errors.New("linter: target is not present in its bound Program")
	errTargetsByProgramLength = errors.New("linter: TargetsByProgram must have one entry per Program")
	errPlanWorkerAborted      = errors.New("linter: parallel plan worker aborted without a panic value")
)

// LintPlan is the immutable Phase 1 input shared by native lint execution and
// host-side consumers such as third-party plugin dispatch. It preserves the
// selected files per Program, including syntax-error and zero-rule files needed
// for LintedFileCount, while resolving each eligible file's complete rule set
// exactly once.
type LintPlan struct {
	programs                  []programLintPlan
	syntacticDiagnosticGroups []syntacticDiagnosticGroup
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

// syntacticDiagnosticGroup freezes the diagnostics for one malformed target.
// Keeping this sparse plan-level fact out of lintFilePlan preserves the hot
// execution unit's file/rule locality without querying a Program twice.
type syntacticDiagnosticGroup struct {
	programIndex int
	diagnostics  []rule.RuleDiagnostic
}

type syntacticDiagnosticKey struct {
	path     string
	ruleName string
	pos      int
	end      int
}

// programRulePlanOptions contains the rule-resolution policy for one already
// selected Program projection. Target discovery and ownership happen before
// this boundary.
type programRulePlanOptions struct {
	Program         *program.Program
	SkipSyntaxCheck bool
	GetRulesForFile RuleHandler
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
// concurrency requirement. Every target must resolve in its corresponding
// Program; the complete binding is validated before rule resolution begins.
func PrepareLintPlan(opts PrepareLintPlanOptions) (*LintPlan, error) {
	return PrepareLintPlanContext(context.Background(), opts)
}

// PrepareLintPlanContext is PrepareLintPlan with caller-owned cancellation.
// The returned plan is either a complete immutable generation or an error;
// partially resolved plans are never published. A parallel worker's abnormal
// exit is re-raised after every worker joins and takes precedence over caller
// cancellation.
func PrepareLintPlanContext(ctx context.Context, opts PrepareLintPlanOptions) (*LintPlan, error) {
	if ctx == nil {
		return nil, errors.New("linter: context must not be nil")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(opts.TargetsByProgram) != len(opts.Programs) {
		return nil, errTargetsByProgramLength
	}
	// Validate the complete ordered input before rule resolution can produce
	// side effects for an earlier Program.
	if err := validatePrograms(opts.Programs); err != nil {
		return nil, err
	}
	if opts.GetRulesForFile == nil {
		return nil, errNilRuleHandler
	}
	plan := &LintPlan{programs: make([]programLintPlan, len(opts.Programs))}
	totalFiles := 0
	for programIndex := range opts.Programs {
		files, err := resolveExactProgramFiles(opts.Programs[programIndex], opts.TargetsByProgram[programIndex])
		if err != nil {
			return nil, fmt.Errorf("linter: Program index %d: %w", programIndex, err)
		}
		plan.programs[programIndex] = programLintPlanFromFiles(opts.Programs[programIndex], files)
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

	resolve := func(resolveCtx context.Context, ref lintPlanFileRef) []rule.RuleDiagnostic {
		if resolveCtx.Err() != nil {
			return nil
		}
		programPlan := &plan.programs[ref.programIndex]
		return resolveProgramLintPlanFile(programRulePlanOptions{
			Program:         programPlan.program,
			GetRulesForFile: opts.GetRulesForFile,
		}, programPlan, ref.fileIndex, resolveCtx)
	}

	workerCount := min(runtime.GOMAXPROCS(0), len(refs))
	if opts.SingleThreaded || workerCount < 2 {
		for _, ref := range refs {
			if diagnostics := resolve(ctx, ref); len(diagnostics) > 0 {
				plan.syntacticDiagnosticGroups = append(
					plan.syntacticDiagnosticGroups,
					syntacticDiagnosticGroup{
						programIndex: ref.programIndex,
						diagnostics:  diagnostics,
					},
				)
			}
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		return plan, nil
	}

	workerCtx, cancelWorkers := context.WithCancel(ctx)
	defer cancelWorkers()
	chunkSize := (len(refs) + workerCount - 1) / workerCount
	groupsByChunk := make([][]syntacticDiagnosticGroup, workerCount)
	var (
		panicOnce  sync.Once
		panicked   bool
		panicValue any
	)
	work := core.NewWorkGroup(false)
	for worker := range workerCount {
		start := worker * chunkSize
		end := min(start+chunkSize, len(refs))
		if start >= end {
			continue
		}
		chunkIndex := worker
		chunk := refs[start:end]
		work.Queue(func() {
			completed := false
			defer func() {
				if completed {
					return
				}
				recovered := recover()
				if recovered == nil {
					recovered = errPlanWorkerAborted
				}
				panicOnce.Do(func() {
					panicked = true
					panicValue = recovered
					cancelWorkers()
				})
			}()
			var groups []syntacticDiagnosticGroup
			for _, ref := range chunk {
				diagnostics := resolve(workerCtx, ref)
				if len(diagnostics) > 0 {
					groups = append(groups, syntacticDiagnosticGroup{
						programIndex: ref.programIndex,
						diagnostics:  diagnostics,
					})
				}
			}
			groupsByChunk[chunkIndex] = groups
			completed = true
		})
	}
	work.RunAndWait()
	if panicked {
		panic(panicValue)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	groupCount := 0
	for _, groups := range groupsByChunk {
		groupCount += len(groups)
	}
	if groupCount > 0 {
		plan.syntacticDiagnosticGroups = make([]syntacticDiagnosticGroup, 0, groupCount)
		for _, groups := range groupsByChunk {
			plan.syntacticDiagnosticGroups = append(plan.syntacticDiagnosticGroups, groups...)
		}
	}
	return plan, nil
}

func newProgramLintPlanForFiles(sourceProgram *program.Program, files []*ast.SourceFile) (programLintPlan, error) {
	if err := validateProgram(sourceProgram); err != nil {
		return programLintPlan{}, err
	}
	return programLintPlanFromFiles(sourceProgram, files), nil
}

func programLintPlanFromFiles(sourceProgram *program.Program, files []*ast.SourceFile) programLintPlan {
	filePlans := make([]lintFilePlan, len(files))
	for index, file := range files {
		filePlans[index].file = file
	}
	return programLintPlan{
		program: sourceProgram,
		files:   filePlans,
	}
}

func prepareProgramLintPlanForFiles(opts programRulePlanOptions, files []*ast.SourceFile) (programLintPlan, error) {
	plan, err := newProgramLintPlanForFiles(opts.Program, files)
	if err != nil {
		return programLintPlan{}, err
	}
	ctx := context.Background()
	for fileIndex := range plan.files {
		// Compatibility callers consume only the execution projection. Syntax
		// diagnostics still gate rule resolution, but their presentation remains
		// owned by the caller.
		_ = resolveProgramLintPlanFile(opts, &plan, fileIndex, ctx)
	}
	return plan, nil
}

func resolveProgramLintPlanFile(
	opts programRulePlanOptions,
	plan *programLintPlan,
	fileIndex int,
	ctx context.Context,
) []rule.RuleDiagnostic {
	filePlan := &plan.files[fileIndex]
	file := filePlan.file
	if !opts.SkipSyntaxCheck {
		diagnostics := CollectFileSyntacticDiagnostics(ctx, opts.Program, file)
		if len(diagnostics) > 0 {
			return diagnostics
		}
	}
	if ctx.Err() != nil {
		return nil
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
	return nil
}

// SyntacticDiagnostics returns diagnostics frozen by this plan. Project-backed
// Programs may be omitted when the caller's program-diagnostics phase reports
// the same syntax failures.
func (p *LintPlan) SyntacticDiagnostics(programDiagnosticsIncluded bool) []rule.RuleDiagnostic {
	if p == nil || len(p.syntacticDiagnosticGroups) == 0 {
		return nil
	}
	seen := make(map[syntacticDiagnosticKey]struct{})
	var diagnostics []rule.RuleDiagnostic
	for _, group := range p.syntacticDiagnosticGroups {
		programPlan := p.programs[group.programIndex]
		if programDiagnosticsIncluded && programPlan.program.CanProvideProgramDiagnostics() {
			continue
		}
		for _, diagnostic := range group.diagnostics {
			key := syntacticDiagnosticKey{
				path:     diagnostic.FilePath,
				ruleName: diagnostic.RuleName,
				pos:      diagnostic.Range.Pos(),
				end:      diagnostic.Range.End(),
			}
			if _, duplicate := seen[key]; duplicate {
				continue
			}
			seen[key] = struct{}{}
			diagnostics = append(diagnostics, diagnostic)
		}
	}
	return diagnostics
}

// HasSyntacticDiagnostics reports whether any exact target has a syntax error,
// independently of which diagnostic producer will present it.
func (p *LintPlan) HasSyntacticDiagnostics() bool {
	return p != nil && len(p.syntacticDiagnosticGroups) > 0
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

func (p *LintPlan) fileCount() int {
	if p == nil {
		return 0
	}
	fileCount := 0
	for _, programPlan := range p.programs {
		fileCount += len(programPlan.files)
	}
	return fileCount
}

func (p *LintPlan) sourcePrograms() []*program.Program {
	programs := make([]*program.Program, len(p.programs))
	for index, programPlan := range p.programs {
		programs[index] = programPlan.program
	}
	return programs
}

func resolveExactProgramFiles(sourceProgram *program.Program, targets []string) ([]*ast.SourceFile, error) {
	// Exact target plans commonly select a Program's complete universe in the
	// same stable order. Preserve the Program-owned slice when selection makes
	// no change, avoiding a map and pointer-slice allocation without inspecting
	// how the Program was constructed.
	files := sourceProgram.SourceFiles()
	if len(targets) == len(files) {
		exact := true
		for fileIndex, target := range targets {
			file := sourceProgram.GetSourceFile(target)
			if file == nil {
				return nil, fmt.Errorf("%w: %q", errTargetNotInProgram, target)
			}
			if file != files[fileIndex] {
				exact = false
				break
			}
		}
		if exact {
			return files, nil
		}
	}

	var filesToLint []*ast.SourceFile
	seen := make(map[string]struct{}, len(targets))
	for _, target := range targets {
		file := sourceProgram.GetSourceFile(target)
		if file == nil {
			return nil, fmt.Errorf("%w: %q", errTargetNotInProgram, target)
		}
		fileName := file.FileName()
		if _, ok := seen[fileName]; ok {
			continue
		}
		seen[fileName] = struct{}{}
		filesToLint = append(filesToLint, file)
	}
	return filesToLint, nil
}
