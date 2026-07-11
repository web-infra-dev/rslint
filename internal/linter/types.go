package linter

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type ConfiguredRule struct {
	Name     string
	Settings map[string]interface{}
	// Globals is the config-declared `languageOptions.globals` for this file
	// (name → declared). The linter merges this with inline `/* global */`
	// comments (parsed once per file, same as DisableManager) before exposing
	// the combined result to rules as ctx.Globals — rules never parse either
	// source themselves. Nil when the config declares none.
	Globals          map[string]bool
	Severity         rule.DiagnosticSeverity
	RequiresTypeInfo bool
	// IsEslintPluginRule marks a rule that executes in the Node plugin-lint
	// worker (mounted via the config's object-form `plugins`) rather than natively
	// in Go. The linter splits these out and dispatches them; its Run is a
	// no-op placeholder.
	IsEslintPluginRule bool
	// Options is the raw user-configured rule options (ESLint's
	// post-severity args). Consumed when dispatching plugin rules to the
	// Node worker; native rules read options through Run's closure instead.
	Options []any
	Run     func(ctx rule.RuleContext) rule.RuleListeners
}

func FilterNonTypeAwareRules(rules []ConfiguredRule) []ConfiguredRule {
	filtered := make([]ConfiguredRule, 0, len(rules))
	for _, r := range rules {
		if !r.RequiresTypeInfo {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

type RuleHandler = func(sourceFile *ast.SourceFile) []ConfiguredRule
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
//   - Scope.{Files,Dirs}=nil              → process all program files
//   - ExcludePaths=nil                    → fall back to the linter default
//     (substring match against utils.ExcludePaths). Pass an explicit empty
//     slice to disable the default.
//   - PerProgramFilter=nil                → no per-program ad-hoc filter
//     (for example config global ignores). Entries within the slice
//     may be nil individually.
//   - GetRulesForFile=nil                 → no lint rules executed
//   - SyntaxErrorFiles=nil                → RunLinter checks each lint target
//     for syntax errors before resolving or running rules. A non-nil set means
//     the caller already performed that check and names the invalid files.
//   - TypeInfoFiles=nil                   → no gap-file distinction. A non-nil
//     set filters RequiresTypeInfo rules and withholds the TypeChecker for files
//     outside it. This field never restricts program-wide type-check.
//   - TypeCheck=false                     → skip the type-check phase
//   - SkipTypeCheckPrograms=nil           → every program participates in
//     type-check. When non-nil, must be parallel to Programs; entries set
//     to true mark the corresponding program to be skipped (typically the
//     non-project fallback Program with synthesized CompilerOptions).
//   - OnDiagnostic=nil                    → diagnostics are dropped
//
// Thread-safety: OnDiagnostic is invoked from multiple goroutines
// concurrently — Phase 1 fans out per program AND per file shard within
// each program (one worker per pool checker), Phase 2 (type-check) fans
// out per program. Callers MUST make their handler safe for concurrent
// calls (channel send, mutex-guarded slice append, sync.Map, etc.).
type RunLinterOptions struct {
	Programs       []*compiler.Program
	SingleThreaded bool

	Scope            FileScope
	ExcludePaths     []string
	PerProgramFilter []FileFilter
	// TargetFiles, when non-nil, enables an exact per-Program lint target plan.
	// Entries are parallel to Programs; a missing, nil, or empty entry means
	// that Program has no lint-rule targets. CLI/API use this after resolving
	// lint targets from config `files`/ignores independently from TypeScript
	// Program membership. nil preserves the legacy Program scan.
	TargetFiles [][]string

	GetRulesForFile  RuleHandler
	TypeInfoFiles    map[string]struct{}
	SyntaxErrorFiles map[string]struct{}

	TypeCheck             bool
	SkipTypeCheckPrograms []bool

	OnDiagnostic DiagnosticHandler
}

// LintSingleFileOptions configures a single-file, single-program rule pass.
// The caller must handle syntactic diagnostics before invoking it.
type LintSingleFileOptions struct {
	Program *compiler.Program
	File    string
	// HasTypeInfo controls whether rules marked RequiresTypeInfo are eligible.
	// Non-type-aware rules may still use the Program's checker for local analysis.
	HasTypeInfo     bool
	GetRulesForFile RuleHandler
	ExcludePaths    []string
	OnDiagnostic    DiagnosticHandler
}
