package linter

import (
	"context"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type ConfiguredRule struct {
	Name             string
	Settings         map[string]interface{}
	Severity         rule.DiagnosticSeverity
	RequiresTypeInfo bool
	// IsEslintPluginRule indicates this rule's execution is delegated to the
	// Node Worker pool over IPC. When true, the linter must NOT call Run
	// during the normal listener traversal — it collects the rule into a
	// per-Program lintEslintPlugin batch instead.
	IsEslintPluginRule bool
	// ConfigKey is the directory of the rslint config that owns this rule
	// for this file (i.e. the `cfgDir` passed into GetEnabledRules). Used
	// by the compat dispatcher to tell the runner which config a file's
	// plugin lookups should resolve against — critical when monorepos
	// install multiple versions of the same plugin in different sub-
	// packages. Empty when no JS config is in play (JSON config path).
	ConfigKey string
	// Options is the user-supplied options block from `rules: { 'foo': ['error', {...}, ...] }`.
	// Native rules access it indirectly through the captured closure in Run; compat
	// rules need it explicitly so the linter can serialize it into the IPC batch
	// payload. Nil when the user wrote a bare severity (`'error'`) with no options.
	// Shape: typically a single object (most ESLint rules) or array (positional).
	Options interface{}
	// LanguageOptions is the per-file ESLint-compat language options
	// already flattened for IPC — a plain `map[string]any` carrying
	// whatever subtree the user wrote under `languageOptions` minus the
	// native-only fields (`parserOptions.project` / `projectService`,
	// which the worker has no business seeing).
	//
	// Opaque to Go: the worker decodes what it needs (globals → scope
	// manager, parserOptions.ecmaVersion → oxc-parser, etc.). New
	// ESLint flat-config fields land here without ANY Go change.
	//
	// Nil when the user didn't set any compat-relevant fields — the
	// worker falls back to its own defaults in that case.
	LanguageOptions map[string]any
	Run             func(ctx rule.RuleContext) rule.RuleListeners
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

// ─── eslint-plugin compat-rule plumbing ───────────────────────────────
//
// When the linter encounters ConfiguredRules with IsEslintPluginRule=true
// it does NOT execute them in the native pipeline (their Run is a
// placeholder no-op). Instead it groups them per Program into a
// CompatBatch and hands the batch to the caller-supplied
// CompatBatchHandler. CLI runs install a handler that reverse-RPCs into
// the Node parent's WorkerPool over stdio IPC. LSP runs install one
// that sends each batch back to the LSP client as a
// `rslint/lintCompatBatch` custom request; the client (the VS Code
// extension or any other Node-hosted LSP client) executes the rules in
// its own WorkerPool and replies with diagnostics. The handler returns
// per-file results that the linter converts into `rule.RuleDiagnostic`
// and streams through OnDiagnostic, indistinguishable from native
// diagnostics downstream.

// CompatLintFile is one file in a per-Program batch. File content is
// USUALLY not shipped on the wire — the runner re-reads from disk via
// `readFileSync(req.filePath, 'utf8')` inside the worker, keeping the
// CLI's IPC payload bounded (shipping source for hundreds of files adds
// latency the worker avoids with a direct stat+read). The exception is
// the LSP / `--api` path, where a file may be an UNSAVED editor buffer
// whose content differs from disk: there `Text` carries the overlay so
// the worker lints what the user sees, not the stale on-disk file (#3).
//
// JSON field names are lowercase to match the runner's TypeScript shape.
// `LanguageOptions` is an OPAQUE map (Go side never introspects it)
// flattened from the user's `languageOptions.*` declarations minus the
// native-only fields (`parserOptions.project`, `projectService`). The
// worker decodes whichever subset it consumes. Adding a new ESLint
// flat-config field requires zero Go changes — just teach the worker
// to read it and expose the field in the TS API.
type CompatLintFile struct {
	Path            string                 `json:"path"`
	LanguageOptions map[string]any         `json:"languageOptions,omitempty"`
	Settings        map[string]interface{} `json:"settings,omitempty"`
	// ConfigKey is the directory of the rslint config that owns this
	// file. The runner uses it to look up the plugin instance set the
	// file should be linted against (in monorepos, different sub-package
	// configs can install different plugin versions). Empty when no JS
	// config governs the file.
	ConfigKey string `json:"configKey,omitempty"`
	// Text, when non-nil, is the source the worker MUST lint instead of
	// re-reading disk — set on the LSP/--api path for (possibly unsaved)
	// editor buffers. A nil pointer omits the field (worker reads disk); a
	// non-nil pointer lets even an empty unsaved buffer ("") override the
	// on-disk content (#3).
	Text *string `json:"text,omitempty"`
}

// CompatRuleConfig carries the per-rule options the Worker needs.
// Severity is reattached Go-side from ConfiguredRule.Severity, NOT sent
// over the wire, so this struct holds only options.
//
// Wire shape: an array of positional values, mirroring ESLint's
// `context.options` exactly. The user's `'rule': ['error', { foo: 1 }]`
// becomes `Options: [{ foo: 1 }]` (severity stripped). Bare-severity
// configurations (`'rule': 'error'`) carry an empty array.
type CompatRuleConfig struct {
	Options []interface{} `json:"options"`
}

// NormalizeRuleOptions converts a user-supplied options value (whatever
// ESLint flat-config accepted) into the positional array shape expected
// by ESLint plugins' `context.options`.
//
// Inputs we tolerate:
//   - nil                      → []
//   - []interface{}            → as-is (already positional array)
//   - any single value         → [value] (most ESLint rules use one
//     positional object: `'rule': ['error', { ... }]` → Options: [{...}])
//
// This is the bridge between rslint's "options" field (which can be any
// JSON value because ESLint's API doesn't strictly type it) and ESLint's
// invariant that `context.options` is always an array.
func NormalizeRuleOptions(raw interface{}) []interface{} {
	if raw == nil {
		return []interface{}{}
	}
	if arr, ok := raw.([]interface{}); ok {
		return arr
	}
	return []interface{}{raw}
}

// CompatBatch is one per-Program request. The dispatcher returns one
// CompatFileResult per file in Files (in the same order).
//
// JSON field names match the runner's TypeScript shape exactly.
type CompatBatch struct {
	Files []CompatLintFile            `json:"files"`
	Rules map[string]CompatRuleConfig `json:"rules"`
	// CollectFixes asks the runner to materialize `descriptor.fix(fixer)`
	// output into per-diagnostic `Fixes`. ALWAYS purely about whether the
	// fix payload is built — NOT about whether it's applied to disk. The
	// consumer (CLI fix-loop, LSP code-action handler) is responsible for
	// applying or surfacing the resulting fixes; the linter never writes
	// files.
	CollectFixes    bool   `json:"collectFixes"`
	SuggestionsMode string `json:"suggestionsMode,omitempty"` // "off" | "eager"
}

// CompatFix mirrors rule.RuleFix wire-side (byte offsets).
type CompatFix struct {
	Range [2]int `json:"range"`
	Text  string `json:"text"`
}

// CompatSuggestion mirrors rule.RuleSuggestion wire-side. `Fixes` is nil
// in suggestionsMode='off'.
type CompatSuggestion struct {
	MessageId string      `json:"messageId,omitempty"`
	Desc      string      `json:"desc,omitempty"`
	Fixes     []CompatFix `json:"fixes,omitempty"`
}

// CompatDiagnostic is one diagnostic produced by a compat rule on one file.
type CompatDiagnostic struct {
	RuleName    string             `json:"ruleName"`
	MessageId   string             `json:"messageId,omitempty"`
	Message     string             `json:"message"`
	StartPos    int                `json:"startPos"`
	EndPos      int                `json:"endPos"`
	Fixes       []CompatFix        `json:"fixes,omitempty"`
	Suggestions []CompatSuggestion `json:"suggestions,omitempty"`
}

// CompatFileResult is the per-file result returned from the dispatcher.
// `ParseError` non-empty signals the file failed to parse — its
// Diagnostics is empty in that case. Other files in the same batch are
// unaffected.
type CompatFileResult struct {
	FilePath    string             `json:"filePath"`
	Diagnostics []CompatDiagnostic `json:"diagnostics"`
	ParseError  string             `json:"parseError,omitempty"`
	Cancelled   bool               `json:"cancelled,omitempty"`
	// RuleErrors carries per-rule failures the worker caught during
	// `rule.create(context)` or listener execution. The rule produced
	// no diagnostics for this file (it failed before it could), but
	// other rules continued. Surfaced to stderr by the linter so a
	// plugin bug is visible — silent rule-not-firing was previously
	// indistinguishable from "rule simply found no issues".
	RuleErrors []CompatRuleError `json:"ruleErrors,omitempty"`
}

// CompatRuleError is one rule's failure record for one file. Shape
// mirrors `ecma-language-plugin.ts`'s `ruleErrors` field exactly.
type CompatRuleError struct {
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

// CompatBatchHandler dispatches a batch to the Worker pool over IPC and
// returns the per-file results in the SAME order as batch.Files. nil
// handler = compat rules are silently skipped (used by `--api` mode and
// tests that don't need compat).
//
// ctx is the lint-level context propagated from RunLinterOptions.Ctx.
// Implementations MUST honor it: a cancelled ctx during the IPC await
// should fail the request promptly rather than running to the
// implementation's own internal timeout. This is the difference between
// "Ctrl-C returns in the time it takes the current file to finish" vs.
// "Ctrl-C returns when the current compat batch's 60s/30s timeout fires".
//
// Cardinality: a single program lint may call this handler MORE THAN
// ONCE — once per unique rule-options-and-severity bucket. The typical
// case (single config, no ESLint `overrides`) is one call per program;
// multi-config-shared-tsconfig setups or configs that vary rules
// per-file via `overrides` produce one call per signature. Each call's
// `batch.Rules` is homogeneous across the batch's files (every file
// shares the same options & severity for every rule in the batch).
// Implementations need not deduplicate work across calls.
type CompatBatchHandler func(ctx context.Context, batch CompatBatch) ([]CompatFileResult, error)

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
	// CompatDispatchErrors counts how many per-Program compat batches
	// failed (dispatcher returned non-nil error: peer crash, IPC timeout,
	// malformed response). Diagnostics from native rules already streamed
	// through OnDiagnostic; the count lets the caller surface the runner
	// failure (exit 2, distinct from the lint-error exit 1).
	CompatDispatchErrors int32
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

	// CompatRuleDispatcher is the hook that the CLI runtime (runCLI) and
	// the LSP server install. When non-nil, ConfiguredRules with
	// IsEslintPluginRule=true are batched per Program and dispatched
	// here; the returned diagnostics are converted to rule.RuleDiagnostic
	// and streamed through OnDiagnostic alongside native ones. Leave nil
	// in `--api` / tests that don't need compat — the linter then
	// silently skips compat rules.
	CompatRuleDispatcher CompatBatchHandler
	// SendCompatFileText makes the compat batch carry each file's source
	// text (CompatLintFile.Text) instead of letting the worker re-read
	// disk. Set by the LSP/--api path so UNSAVED editor buffers lint their
	// in-memory overlay; left false by the CLI to keep the IPC payload
	// bounded (the worker reads the on-disk file, authoritative there).
	// See #3.
	SendCompatFileText bool

	// CollectFixes / SuggestionsMode propagate to per-Program runs and are
	// shipped in the lintEslintPlugin batch payload so the Worker invokes
	// `descriptor.fix(fixer)` and `descriptor.suggest[i].fix(fixer)` only
	// when the consumer asked for them.
	//
	// CollectFixes is intentionally about PAYLOAD only — it does not
	// instruct the linter to write files. Applying fixes (CLI --fix
	// fix-loop, LSP code-action / fixAll) is the caller's responsibility.
	// CLI sets it to true whenever --fix is on so the fix-loop has
	// something to apply; LSP sets it to true unconditionally so Quick
	// Fix and source.fixAll for ESLint-plugin rules work the same way as
	// for native rules. Tests that don't exercise fixes leave it false to
	// avoid the small cost of materializing fix payloads.
	CollectFixes    bool
	SuggestionsMode string

	// Ctx is consulted at file-loop boundaries inside each Program. When
	// it fires (Done() returns), RunLinter returns the partial result
	// collected so far with a context.Canceled / DeadlineExceeded error.
	// Diagnostics already streamed via OnDiagnostic are kept — only the
	// not-yet-visited files are dropped. nil disables cancellation.
	//
	// Wiring is intentionally lightweight: each file boundary is a poll
	// point, so cancel latency is bounded by the time to finish the
	// current file's rule traversal. A truly stuck rule on a giant file
	// is NOT interruptible — but the user perception of "Ctrl-C works"
	// is satisfied for any realistic lint run.
	Ctx context.Context
}

// LintSingleFileOptions configures a single-file, single-program lint pass.
// Designed for IDE/LSP per-keystroke usage. Does not run type-check.
//
// CompatRuleDispatcher / CollectFixes / SuggestionsMode mirror the eslint-
// plugin compat fields on RunLinterOptions. The LSP path installs a
// dispatcher that sends each batch back to the client as an
// `rslint/lintCompatBatch` LSP request; the client (extension)
// executes the rules in its own WorkerPool. Leave nil/zero in callers
// that don't need plugin rule execution.
//
// CollectFixes: set to true so plugin-rule fixes ride along on each
// diagnostic. The LSP server reads `ruleDiag.Fixes()` when synthesising
// code actions, so without CollectFixes Quick Fix and source.fixAll
// would silently miss every ESLint-plugin rule.
type LintSingleFileOptions struct {
	Program         *compiler.Program
	File            string
	GetRulesForFile RuleHandler
	ExcludePaths    []string
	OnDiagnostic    DiagnosticHandler

	CompatRuleDispatcher CompatBatchHandler
	CollectFixes         bool
	SuggestionsMode      string

	// Ctx is forwarded to the CompatRuleDispatcher so a cancelled LSP
	// request (per-keystroke supersession, document close, server shutdown)
	// aborts the in-flight compat IPC instead of running to its per-batch
	// ceiling. Nil disables cancellation.
	Ctx context.Context
}
