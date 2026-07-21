package rule

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// DiagnosticSeverity represents the severity level of a diagnostic
type DiagnosticSeverity int

const (
	SeverityOff DiagnosticSeverity = iota
	SeverityWarning
	SeverityError
)

// String returns the string representation of the severity
func (s DiagnosticSeverity) String() string {
	switch s {
	case SeverityError:
		return "error"
	case SeverityWarning:
		return "warn"
	case SeverityOff:
		return "off"
	default:
		return "error"
	}
}

// String returns the string representation of the severity
func (s DiagnosticSeverity) Int() int {
	switch s {
	case SeverityError:
		return 1
	case SeverityWarning:
		return 2
	case SeverityOff:
		return 0
	default:
		return 0
	}
}

// ParseSeverity converts a string to DiagnosticSeverity
func ParseSeverity(level string) DiagnosticSeverity {
	switch level {
	case "error":
		return SeverityError
	case "warn", "warning":
		return SeverityWarning
	case "off":
		return SeverityOff
	default:
		return SeverityError // Default to error for unknown values
	}
}

// NormalizeOptions returns a rule's options in ESLint context.options form
// ([]any). config.parseArrayRuleConfig never collapses RuleConfig.Options, but
// LegacyUnwrapOptions (the compatibility shim for pre-migration `parseOptions
// any` bodies) does — a rule that round-trips through it and then wants the
// eslint-format array back (e.g. reads optArray[0]) would silently miss a
// single-option config otherwise. Re-wrapping a bare value lets every caller
// read options[0] uniformly, whether the option arrived wrapped (multi-option
// or an explicit array) or unwrapped (a single option).
//
// It returns an empty (non-nil) slice when no options were configured, so both
// native rules (which key on `len == 0 → defaults`) and the eslint-plugin host
// (which serializes context.options to JSON and needs `[]`, not `null`) share a
// single normalization path.
func NormalizeOptions(raw any) []any {
	if raw == nil {
		return []any{}
	}
	if arr, ok := raw.([]interface{}); ok {
		return arr
	}
	return []any{raw}
}

// LegacyUnwrapOptions is NormalizeOptions' inverse: it collapses a rule's options
// array (Run's []any parameter — ESLint's context.options, the config array
// after the severity level) back to the single bare value most existing rule
// implementations parse. Empty → nil; a single element → that element;
// otherwise the slice itself. This is the compatibility shim old
// `parseOptions(options any)` bodies call so they don't need to change beyond
// their Run signature.
func LegacyUnwrapOptions(options []any) any {
	switch len(options) {
	case 0:
		return nil
	case 1:
		return options[0]
	default:
		return options
	}
}

const (
	lastTokenKind                        ast.Kind = 1000
	lastOnExitTokenKind                  ast.Kind = 2000
	lastOnAllowPatternTokenKind          ast.Kind = 3000
	lastOnAllowPatternOnExitTokenKind    ast.Kind = 4000
	lastOnNotAllowPatternTokenKind       ast.Kind = 5000
	lastOnNotAllowPatternOnExitTokenKind ast.Kind = 6000
)

func ListenerOnExit(kind ast.Kind) ast.Kind {
	return kind + 1000
}

// TODO(port): better name
func ListenerOnAllowPattern(kind ast.Kind) ast.Kind {
	return kind + lastOnExitTokenKind
}
func ListenerOnNotAllowPattern(kind ast.Kind) ast.Kind {
	return kind + lastOnAllowPatternOnExitTokenKind
}

type RuleListeners map[ast.Kind](func(node *ast.Node))

type Rule struct {
	Name             string
	RequiresTypeInfo bool
	// IsEslintPluginRule marks a placeholder rule whose actual execution
	// happens in a Node worker — an ESLint-plugin rule mounted via the
	// config's object-form `plugins`. Its Run is a no-op in Go; the linter
	// splits these out and dispatches them to the plugin-lint host.
	IsEslintPluginRule bool
	// Schema validates the rule's resolved options array (ESLint's
	// context.options — the config array after the severity level) before
	// linting starts, filling schema-declared `default` values into the
	// options in place the way ajv's `useDefaults` does for ESLint (see
	// [Schema.Validate]). Rules that take no options should set it to the
	// shared [EmptyArraySchema]. nil means the rule has not declared a schema
	// yet (most rules, until migrated one-by-one): its options pass through
	// unvalidated, exactly as before the schema framework existed.
	Schema *Schema
	Run    func(ctx RuleContext, options []any) RuleListeners
}

func CreateRule(r Rule) Rule {
	return Rule{
		Name:             "@typescript-eslint/" + r.Name,
		RequiresTypeInfo: r.RequiresTypeInfo,
		Schema:           r.Schema,
		Run:              r.Run,
	}
}

type RuleMessage struct {
	Id          string
	Description string
	// Data exposes the placeholder values that were substituted into
	// Description (e.g. ESLint's `report({ data: { type } })`). Downstream
	// reporters / IDE clients can use this for structured access. Optional —
	// rules that don't carry placeholders may leave it nil.
	Data map[string]string
}

type RuleFix struct {
	Text  string
	Range core.TextRange
}

func RuleFixInsertBefore(file *ast.SourceFile, node *ast.Node, text string) RuleFix {
	trimmed := utils.TrimNodeTextRange(file, node)
	return RuleFix{
		Text:  text,
		Range: trimmed.WithEnd(trimmed.Pos()),
	}
}
func RuleFixInsertAfter(node *ast.Node, text string) RuleFix {
	return RuleFix{
		Text:  text,
		Range: node.Loc.WithPos(node.End()),
	}
}
func RuleFixReplace(file *ast.SourceFile, node *ast.Node, text string) RuleFix {
	return RuleFixReplaceRange(utils.TrimNodeTextRange(file, node), text)
}
func RuleFixReplaceRange(textRange core.TextRange, text string) RuleFix {
	return RuleFix{
		Text:  text,
		Range: textRange,
	}
}
func RuleFixRemove(file *ast.SourceFile, node *ast.Node) RuleFix {
	return RuleFixReplace(file, node, "")
}
func RuleFixRemoveRange(textRange core.TextRange) RuleFix {
	return RuleFixReplaceRange(textRange, "")
}

type RuleSuggestion struct {
	Message  RuleMessage
	FixesArr []RuleFix
}

func (s RuleSuggestion) Fixes() []RuleFix {
	return s.FixesArr
}

// DiagnosticOrigin identifies the subsystem that produced a diagnostic.
// Its zero value is a lint diagnostic so existing rule and plugin producers
// remain source-compatible; TypeScript producers must opt in explicitly.
type DiagnosticOrigin uint8

const (
	DiagnosticOriginLint DiagnosticOrigin = iota
	DiagnosticOriginTypeScript
	// DiagnosticOriginUnusedDisableDirective marks a synthetic diagnostic
	// produced by --report-unused-disable-directives, not by a rule.
	DiagnosticOriginUnusedDisableDirective
)

// UnusedDisableDirectiveRuleName is the sentinel RuleName used for
// diagnostics with DiagnosticOriginUnusedDisableDirective. It is not a real
// rule and cannot be configured via --rule or `rules`.
const UnusedDisableDirectiveRuleName = "unused-disable-directive"

type RuleDiagnostic struct {
	Range    core.TextRange
	RuleName string
	Message  RuleMessage
	// nil if no fixes were provided
	FixesPtr *[]RuleFix
	// nil if no suggestions were provided
	Suggestions *[]RuleSuggestion
	// SourceFile is the file this diagnostic anchors to. It is the
	// ast.SourceFileLike interface (Text + ECMALineMap) rather than a
	// concrete *ast.SourceFile so that ESLint-plugin diagnostics — which
	// are produced in a Node worker and have no ts-go AST — can supply a
	// lightweight text-only implementation (internal/linter.textSourceFile)
	// and still render line/column through the scanner.
	SourceFile ast.SourceFileLike
	// FilePath is the diagnostic's file name. Stored separately because
	// ast.SourceFileLike exposes no FileName(); native diagnostics set it
	// from SourceFile.FileName(), plugin diagnostics from the wire path.
	FilePath string
	Severity DiagnosticSeverity
	Origin   DiagnosticOrigin
	// PreFormatted indicates that Message.Description already contains
	// structured formatting (e.g. indented continuation lines from tsc diagnostics).
	// The renderer will use a simple 2-space indent instead of the │ border style.
	PreFormatted bool
}

func (d RuleDiagnostic) Fixes() []RuleFix {
	if d.FixesPtr == nil {
		return []RuleFix{}
	}
	return *d.FixesPtr
}

type RuleContext struct {
	SourceFile *ast.SourceFile
	Settings   map[string]interface{}
	// ConfigGlobals contains only globals from the effective
	// `languageOptions.globals` configuration, before inline comments are
	// applied. A false value is an explicit "off" setting.
	ConfigGlobals map[string]bool
	// InlineGlobals contains `/* global */` declaration metadata in first-name
	// source order. Rules can use its exact name ranges without scanning source
	// text again. Treat the slice and its nested ranges as read-only.
	InlineGlobals []InlineGlobal
	// Globals is the fully resolved set of declared global names for this
	// file — config `languageOptions.globals` merged with inline
	// `/* global */` comments, resolved once per file by the linter. Rules
	// should read this instead of parsing comments or config themselves. Nil
	// only when neither source mentions any globals; a name maps to false when
	// its final setting is "off".
	Globals map[string]bool
	// Comments lazily provides every comment in SourceFile, in source order.
	// Rules should call Comments.All instead of walking the token tree with
	// utils.ForEachComment. The first consumer computes the list and every
	// later consumer for this file reuses it.
	Comments                   *CommentStore
	Program                    *compiler.Program
	TypeChecker                *checker.Checker
	DisableManager             *DisableManager
	ReportRange                func(textRange core.TextRange, msg RuleMessage)
	ReportRangeWithFixes       func(textRange core.TextRange, msg RuleMessage, fixes ...RuleFix)
	ReportRangeWithSuggestions func(textRange core.TextRange, msg RuleMessage, suggestions ...RuleSuggestion)
	ReportNode                 func(node *ast.Node, msg RuleMessage)
	ReportNodeWithFixes        func(node *ast.Node, msg RuleMessage, fixes ...RuleFix)
	ReportNodeWithSuggestions  func(node *ast.Node, msg RuleMessage, suggestions ...RuleSuggestion)
	// ReportNodeWithFixesAndSuggestions emits a single diagnostic carrying
	// BOTH an autofix and one or more suggestions. Used by rules that follow
	// upstream's "promote first suggestion to fix while keeping the
	// suggestion" pattern (e.g., ESLint's
	// `enableDangerousAutofixThisMayCauseInfiniteLoops` in
	// react-hooks/exhaustive-deps).
	ReportNodeWithFixesAndSuggestions func(node *ast.Node, msg RuleMessage, fixes []RuleFix, suggestions []RuleSuggestion)
	// ReportRangeWithFixesAndSuggestions is the range-keyed twin of
	// ReportNodeWithFixesAndSuggestions. Same semantics, anchors the
	// diagnostic at an explicit TextRange instead of a node's trimmed
	// range.
	ReportRangeWithFixesAndSuggestions func(textRange core.TextRange, msg RuleMessage, fixes []RuleFix, suggestions []RuleSuggestion)
}

func ReportNodeWithFixesOrSuggestions(ctx RuleContext, node *ast.Node, fix bool, msg RuleMessage, suggestionMsg RuleMessage, fixes ...RuleFix) {
	if fix {
		ctx.ReportNodeWithFixes(node, msg, fixes...)
	} else {
		ctx.ReportNodeWithSuggestions(node, msg, RuleSuggestion{
			Message:  suggestionMsg,
			FixesArr: fixes,
		})
	}
}
