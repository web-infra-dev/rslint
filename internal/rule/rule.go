package rule

import (
	"none.none/tsgolint/internal/utils"
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
)

func ListenerOnExit(kind ast.Kind) ast.Kind {
	return kind + ast.KindCount + 1
}

type RuleListeners map[ast.Kind](func (node *ast.Node))

type Rule struct {
	Name string
	Run  func(ctx RuleContext, options any) RuleListeners
}

type RuleMessage struct {
	Id          string
	Description string
}

type RuleFix struct {
	Text  string
	Range core.TextRange
}

func RuleFixInsertBefore(node *ast.Node, text string) RuleFix {
	trimmed := utils.TrimNodeTextRange(node)
	return RuleFix{
		Text: text,
		Range: trimmed.WithEnd(trimmed.Pos()),
	}
}
func RuleFixInsertAfter(node *ast.Node, text string) RuleFix {
	return RuleFix{
		Text: text,
		Range: node.Loc.WithPos(node.End()),
	}
}
func RuleFixReplace(node *ast.Node, text string) RuleFix {
	return RuleFixReplaceRange(utils.TrimNodeTextRange(node), text)
}
func RuleFixReplaceRange(textRange core.TextRange, text string) RuleFix {
	return RuleFix{
		Text: text,
		Range: textRange,
	}
}

type RuleSuggestion struct {
	Message  RuleMessage
	FixesArr []RuleFix
}

func (s RuleSuggestion) Fixes() []RuleFix {
	return s.FixesArr
}

type RuleDiagnostic struct {
	Range   core.TextRange
	RuleName string
	Message RuleMessage
	// nil if no fixes were provided
	FixesPtr *[]RuleFix
	// nil if no suggestions were provided
	Suggestions *[]RuleSuggestion
	SourceFile  *ast.SourceFile
}

func (d RuleDiagnostic) Fixes() []RuleFix {
	if d.FixesPtr == nil {
		return []RuleFix{}
	}
	return *d.FixesPtr
}

type RuleContext struct {
	SourceFile *ast.SourceFile
	Program                   *compiler.Program
	TypeChecker               *checker.Checker
	ReportRange                func(textRange core.TextRange, msg RuleMessage)
	ReportNode                func(node *ast.Node, msg RuleMessage)
	ReportNodeWithFixes       func(node *ast.Node, msg RuleMessage, fixes ...RuleFix)
	ReportNodeWithSuggestions func(node *ast.Node, msg RuleMessage, suggestions ...RuleSuggestion)
}
