package rule

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/utils"
)

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
)

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
