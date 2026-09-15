package linter

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type RuleHandler = func(sourceFile *ast.SourceFile) []rule.ConfiguredRule
type DiagnosticHandler = func(diagnostic rule.RuleDiagnostic)

// LintResult holds the outcome of a RunLinter invocation.
type LintResult struct {
	LintedFileCount int32
	ExecutedRules   map[string]struct{}
}

// PrepareLintPlanOptions configures exact lint-plan construction for an ordered
// Program sequence. TargetsByProgram must have one entry per Program; an empty
// entry means that Program has no lint targets. Planning never scans a Program
// to infer targets.
//
// GetRulesForFile is called once for every selected file without syntax errors.
// It may be called concurrently unless SingleThreaded is set.
type PrepareLintPlanOptions struct {
	Programs         []*program.Program
	TargetsByProgram [][]string
	SingleThreaded   bool
	GetRulesForFile  RuleHandler
}

// RunLinterOptions configures execution of a prepared lint plan and an optional
// program-wide type-check pass.
//
// Zero-value semantics:
//   - SingleThreaded=false     → use the default parallel work group
//   - LintPlan=nil             → skip the lint-rule phase
//   - TypeCheckOnlyPrograms    → used only when LintPlan is nil and
//     TypeCheck is true; entries must be non-nil Programs created by
//     internal/program
//   - Syntax/type capabilities → derived from each Program. Syntax
//     errors suppress rules for that file; source-only Programs filter rules
//     requiring type information and never participate in type-check.
//   - TypeCheck=false          → skip the type-check phase
//   - Consumer=zero            → diagnostics are dropped and no
//     optional native edit artifacts are materialized
//   - Timing=nil               → per-rule timing collection is off
//     and the lint hot path pays no instrumentation cost. When non-nil, every
//     rule Run call and listener invocation is timed and accumulated into the
//     collector, keyed by rule name. Callers may share one collector across
//     multiple RunLinter invocations (e.g. --fix re-lint passes) to aggregate.
//
// Thread-safety: Consumer.Report is invoked from multiple goroutines
// concurrently — Phase 1 fans out per program AND per file shard within
// each program (one worker per pool checker), Phase 2 (type-check) fans
// out per program. Callers MUST make their handler safe for concurrent
// calls (channel send, mutex-guarded slice append, sync.Map, etc.).
type RunLinterOptions struct {
	SingleThreaded bool
	// Cwd is the working directory of the linting run, forwarded verbatim to
	// every RuleContext. See RuleContext.ProcessCurrentDirectory for what rules
	// may assume of it.
	Cwd string

	// LintPlan owns the ordered Program sequence and its complete immutable
	// file/rule projection for native execution and optional third-party plugin
	// dispatch.
	LintPlan *LintPlan
	// TypeCheckOnlyPrograms supplies the complete Program sequence only for a
	// planless type-check-only pass. A non-nil LintPlan already owns the Program
	// sequence used by both lint and optional type-check phases, so providing
	// TypeCheckOnlyPrograms with a plan is invalid.
	TypeCheckOnlyPrograms []*program.Program
	// Consumer owns diagnostic delivery and the optional edit artifacts needed
	// by native Go consumers. It does not alter the separate eslint-plugin
	// reverse-dispatch request.
	Consumer rule.DiagnosticConsumer

	TypeCheck bool

	Timing *TimingCollector
}

// LintSingleFileOptions configures a single-file, single-program rule pass.
// The caller must handle syntactic diagnostics before invoking it.
type LintSingleFileOptions struct {
	// Program is the exact rslint source generation containing File.
	Program *program.Program
	// File is the exact source-file name exposed by Program.
	File string
	// HasTypeInfo controls whether rules marked RequiresTypeInfo are eligible.
	// Non-type-aware rules may still use the Program's checker for local analysis.
	HasTypeInfo bool
	// GetRulesForFile resolves the selected file's configured rules. nil keeps
	// the historical no-op behavior and does not validate Program or File.
	GetRulesForFile RuleHandler
	// Cwd has the same meaning as RunLinterOptions.Cwd.
	Cwd string
	// Consumer has the same native-only semantics as RunLinterOptions.Consumer.
	Consumer rule.DiagnosticConsumer
}
