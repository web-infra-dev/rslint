package linter

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type ConfiguredRule struct {
	Name             string
	Settings         map[string]interface{}
	Severity         rule.DiagnosticSeverity
	RequiresTypeInfo bool
	Run              func(ctx rule.RuleContext) rule.RuleListeners
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
//     (multi-config ownership / config `ignores`). Entries within the slice
//     may be nil individually.
//   - GetRulesForFile=nil                 → no lint rules executed
//   - TypeInfoFiles=nil                   → no gap-file distinction
//     (all files may run type-aware rules)
//   - TypeCheck=false                     → skip the type-check phase
//   - SkipTypeCheckPrograms=nil           → every program participates in
//     type-check. When non-nil, must be parallel to Programs; entries set
//     to true mark the corresponding program to be skipped (typically the
//     gap-file fallback program with synthesized CompilerOptions).
//   - OnDiagnostic=nil                    → diagnostics are dropped
//
// Thread-safety: OnDiagnostic is invoked from multiple goroutines
// concurrently — Phase 1 fans out per program, Phase 2 (type-check) does
// the same. Callers MUST make their handler safe for concurrent calls
// (channel send, mutex-guarded slice append, sync.Map, etc.).
type RunLinterOptions struct {
	Programs       []*compiler.Program
	SingleThreaded bool

	Scope            FileScope
	ExcludePaths     []string
	PerProgramFilter []FileFilter

	GetRulesForFile RuleHandler
	TypeInfoFiles   map[string]struct{}

	TypeCheck             bool
	SkipTypeCheckPrograms []bool

	OnDiagnostic DiagnosticHandler
}

// LintSingleFileOptions configures a single-file, single-program lint pass.
// Designed for IDE/LSP per-keystroke usage. Does not run type-check.
type LintSingleFileOptions struct {
	Program         *compiler.Program
	File            string
	GetRulesForFile RuleHandler
	ExcludePaths    []string
	OnDiagnostic    DiagnosticHandler
}
