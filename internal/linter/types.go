package linter

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Compatibility alias keeps linter-focused callers source-compatible while
// ownership lives in internal/rule. New framework code should name
// rule.ConfiguredRule directly.
type ConfiguredRule = rule.ConfiguredRule

type RuleHandler = func(sourceFile *ast.SourceFile) []rule.ConfiguredRule
type DiagnosticHandler = func(diagnostic rule.RuleDiagnostic)

// FileScope describes user-supplied "lint targets" (CLI args).
//
// Both fields are independently nullable:
//   - nil slice  → that dimension does not constrain (e.g. Files=nil
//     means "no per-file restriction").
//   - empty slice (len 0, non-nil) → that dimension matches NOTHING.
//     This is how the CLI distinguishes "no files arg supplied" from
//     "files arg supplied but empty".
//   - both nil → all program files pass scope.
//   - both empty → no program files pass scope (lint phase is silent).
//
// FileScope only restricts the lint-rule phase. Type-check (Phase 2 of
// RunLinter) ignores FileScope and reports diagnostics for every file
// the TypeScript program loaded — see RunLinterOptions for details.
type FileScope struct {
	Files []string
	Dirs  []string
}

// FileFilter is a generic "should this file be processed" predicate.
// nil means "everything passes".
type FileFilter func(absPath string) bool

// LintResult holds the outcome of a RunLinter invocation.
type LintResult struct {
	LintedFileCount int32
	ExecutedRules   map[string]struct{}
}

// RunLinterOptions configures a multi-program lint (and optional type-check) pass.
//
// Zero-value semantics:
//   - SingleThreaded=false                → use the default parallel work group
//   - Programs entries                    → must be non-nil Programs created by
//     internal/program whenever either lint or type-check consumes them
//   - Scope.{Files,Dirs}=nil              → process all program files
//   - ExcludePaths=nil                    → fall back to the linter default
//     (substring match against utils.ExcludePaths). Pass an explicit empty
//     slice to disable the default.
//   - PerProgramFilter=nil                → no per-program ad-hoc filter
//     (for example config global ignores). Entries within the slice
//     may be nil individually.
//   - GetRulesForFile=nil                 → no lint rules executed
//   - PreparedPlan=nil                    → RunLinter collects targets and
//     resolves rules through GetRulesForFile during the lint phase. Callers
//     that need the same resolved targets before native execution may build a
//     plan with PrepareLintPlan and pass it here.
//   - Syntax and type capabilities         → derived from each Program. Syntax
//     errors suppress rules for that file; source-only Programs filter rules
//     requiring type information and never participate in type-check.
//   - TypeCheck=false                     → skip the type-check phase
//   - Consumer=zero                        → diagnostics are dropped and no
//     optional native edit artifacts are materialized
//   - Timing=nil                          → per-rule timing collection is off
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
	// Programs contains immutable rslint source universes. Their construction
	// strategy is encapsulated by Program and is not part of lint semantics.
	Programs       []*program.Program
	SingleThreaded bool
	// Cwd is the working directory of the linting run, forwarded verbatim to
	// every RuleContext. See RuleContext.ProcessCurrentDirectory for what rules
	// may assume of it.
	Cwd string

	Scope            FileScope
	ExcludePaths     []string
	PerProgramFilter []FileFilter
	// TargetFiles, when non-nil, enables an exact per-Program lint target plan.
	// Entries are parallel to Programs; a missing, nil, or empty entry means
	// that Program has no lint-rule targets. CLI/API use this after resolving
	// lint targets from config `files`/ignores independently from TypeScript
	// Program membership. nil preserves the legacy Program scan.
	TargetFiles [][]string

	GetRulesForFile RuleHandler
	// PreparedPlan, when non-nil, must have been built from these same Phase 1
	// options with PrepareLintPlan. RunLinter consumes its per-Program files and
	// resolved rules without collecting targets or calling GetRulesForFile again.
	// The callback remains required to distinguish a lint pass from
	// --type-check-only and to preserve the existing zero-value contract.
	PreparedPlan *LintPlan
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
	HasTypeInfo     bool
	GetRulesForFile RuleHandler
	ExcludePaths    []string
	// Cwd has the same meaning as RunLinterOptions.Cwd.
	Cwd string
	// Consumer has the same native-only semantics as RunLinterOptions.Consumer.
	Consumer rule.DiagnosticConsumer
}
