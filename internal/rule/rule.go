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
	Schema0            Schema
	Schema1            Schema
	Run                func(ctx RuleContext, options any) RuleListeners
	RunWithOptions   func(ctx RuleContext, options any) RuleListeners
}

// ValidateAndHydrateOptions decodes raw config options against the rule's Schema0 and Schema1,
// applying defaults and returning the validated typed options.
func ValidateAndHydrateOptions(schema0 Schema, schema1 Schema, ruleName string, raw any) (any, error) {
	if schema0 == nil {
		return nil, nil
	}

	// Normalize raw config into a slice of 2 option elements
	var raw0, raw1 any
	if raw != nil {
		if arr, ok := raw.([]any); ok {
			if len(arr) > 0 {
				raw0 = arr[0]
			}
			if len(arr) > 1 {
				raw1 = arr[1]
			}
		} else if arrInterface, ok := raw.([]interface{}); ok {
			if len(arrInterface) > 0 {
				raw0 = arrInterface[0]
			}
			if len(arrInterface) > 1 {
				raw1 = arrInterface[1]
			}
		} else {
			// Single option value passed directly
			raw0 = raw
		}
	}

	// Validate Option 0
	val0, err := schema0.Validate(raw0)
	if err != nil {
		return nil, fmt.Errorf("configuration validation failed for rule %q (option 0): %w", ruleName, err)
	}

	// If there's no Option 1 schema, just return the validated Option 0
	if schema1 == nil {
		return val0, nil
	}

	// Validate Option 1
	val1, err := schema1.Validate(raw1)
	if err != nil {
		return nil, fmt.Errorf("configuration validation failed for rule %q (option 1): %w", ruleName, err)
	}

	// Return as a slice containing both validated options
	return []any{val0, val1}, nil
}

func CreateRule(r Rule) Rule {
	return Rule{
		Name:             "@typescript-eslint/" + r.Name,
		RequiresTypeInfo: r.RequiresTypeInfo,
		Schema0:          r.Schema0,
		Schema1:          r.Schema1,
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
