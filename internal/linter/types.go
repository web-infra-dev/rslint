package linter

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type ConfiguredRule struct {
	Name             string
	Environment      *RuleEnvironment
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

// RuleEnvironment is the effective file-level configuration shared by every
// configured rule in one resolved rule set. It is immutable after resolution;
// ConfiguredRule carries only a pointer so settings, language options, and
// globals are not cloned into every rule entry.
type RuleEnvironment struct {
	Settings map[string]interface{}
	// LanguageOptions is normalized once and used to construct ctx.Globals. Its
	// zero value selects latest.
	LanguageOptions rule.LanguageOptions
	// Globals contains config-declared language globals. Inline declarations
	// are merged once per source file during execution.
	Globals map[string]utils.GlobalAccess
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
//   - SyntaxErrorFiles=nil                → RunLinter checks each lint target
//     for syntax errors before resolving or running rules. A non-nil set means
//     the caller already performed that check and names the invalid files.
//   - TypeInfoFiles=nil                   → every file for which its Program
//     can supply a checker is eligible for type-aware rules. A non-nil set
//     further restricts that eligibility and checker delivery to named files.
//     Programs without checker capability remain syntax-only regardless of
//     this field. It never restricts program-wide type-check.
//   - TypeCheck=false                     → skip the type-check phase
//   - SkipTypeCheckPrograms=nil           → Phase 2 asks every Program for
//     program-wide diagnostics; Programs without that capability are no-ops.
//     When TypeCheck is true and this is non-nil, it must be parallel to
//     Programs; entries set
//     to true mark the corresponding program to be skipped (typically the
//     non-project fallback Program with synthesized CompilerOptions).
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
	PreparedPlan     *LintPlan
	TypeInfoFiles    map[string]struct{}
	SyntaxErrorFiles map[string]struct{}
	// Consumer owns diagnostic delivery and the optional edit artifacts needed
	// by native Go consumers. It does not alter the separate eslint-plugin
	// reverse-dispatch request.
	Consumer rule.DiagnosticConsumer

	TypeCheck             bool
	SkipTypeCheckPrograms []bool

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
	// CacheModuleSpecifiers lets Programs reusing exact SourceFile objects share
	// the module graph's syntax-only collection. An editor lints one file
	// against a new Program per keystroke, so this avoids re-reading every
	// unchanged file's imports while keeping resolution local to each Program.
	CacheModuleSpecifiers bool
}
