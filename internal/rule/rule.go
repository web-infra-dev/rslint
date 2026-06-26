package rule

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// DiagnosticSeverity represents the severity level of a diagnostic
type DiagnosticSeverity int

const (
	SeverityError DiagnosticSeverity = iota
	SeverityWarning
	SeverityOff
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
	// Schema declares the options validator for this rule. Use Tuple(...) for rules
	// with multiple positional arguments; use Object/Enum/etc. for a single option.
	Schema Schema
	// Run is the legacy, non-validated rule execution callback. Receives the raw configuration object.
	Run func(ctx RuleContext, options any) RuleListeners
	// RunWithOptions is the schema-driven rule execution callback. Receives parsed, validated,
	// and default-hydrated options as a []any slice matching the declared Tuple schema.
	RunWithOptions func(ctx RuleContext, options []any) RuleListeners
}

// ValidateAndHydrateOptions decodes raw config options against the rule's Schema,
// applying defaults and returning the validated options as a []any slice.
// The top-level Schema must be a Tuple or Array — both produce []any from Validate.
func ValidateAndHydrateOptions(schema Schema, ruleName string, raw any) ([]any, error) {
	if schema == nil {
		return nil, nil //nolint:nilnil // no Schema means no options to validate, not an error
	}
	val, err := schema.Validate(raw)
	if err != nil {
		return nil, fmt.Errorf("configuration validation failed for rule %q: %w", ruleName, err)
	}
	result, ok := val.([]any)
	if !ok {
		return nil, fmt.Errorf("rule %q: top-level Schema must be Tuple or Array (got %T); wrap the schema in rule.Tuple(...)", ruleName, val)
	}
	return result, nil
}

func CreateRule(r Rule) Rule {
	return Rule{
		Name:             "@typescript-eslint/" + r.Name,
		RequiresTypeInfo: r.RequiresTypeInfo,
		Schema:           r.Schema,
		Run:              r.Run,
		RunWithOptions:   r.RunWithOptions,
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
	SourceFile                 *ast.SourceFile
	Settings                   map[string]interface{}
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
