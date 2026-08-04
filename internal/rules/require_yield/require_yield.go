package require_yield

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var missingYieldMessage = rule.RuleMessage{
	Id:          "missingYield",
	Description: "This generator function does not have 'yield'.",
}

func isGenerator(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindMethodDeclaration:
		return node.BodyData().AsteriskToken != nil
	}
	return false
}

type stackFrame struct {
	node     *ast.Node
	hasYield bool
}

func startSkippingDecorators(sourceText string, node *ast.Node) int {
	pos := node.Pos()
	if modifiers := node.Modifiers(); modifiers != nil {
		for _, modifier := range modifiers.Nodes {
			if modifier != nil && modifier.Kind == ast.KindDecorator && modifier.End() > pos {
				pos = modifier.End()
			}
		}
	}
	return scanner.SkipTrivia(sourceText, pos)
}

func functionDeclarationHeadStart(sourceText string, node *ast.Node) int {
	pos := node.Pos()
	if modifiers := node.Modifiers(); modifiers != nil {
		for _, modifier := range modifiers.Nodes {
			if modifier == nil {
				continue
			}
			if modifier.Kind == ast.KindExportKeyword || modifier.Kind == ast.KindDefaultKeyword {
				pos = modifier.End()
				continue
			}
			return scanner.SkipTrivia(sourceText, modifier.Pos())
		}
	}
	return scanner.SkipTrivia(sourceText, pos)
}

func generatorHeadRange(sourceFile *ast.SourceFile, node *ast.Node) core.TextRange {
	parameters := node.ParameterList()
	if parameters == nil {
		return utils.GetFunctionHeadLoc(sourceFile, node)
	}

	// ParameterList.Pos is the full start of the first parameter (or the
	// closing paren), so the preceding byte is the opening paren consumed by
	// the parser. Validate the invariant and retain the generic scanner-based
	// helper as an error-recovery fallback.
	sourceText := sourceFile.Text()
	openParen := parameters.Pos() - 1
	if openParen < 0 || openParen >= len(sourceText) || sourceText[openParen] != '(' {
		return utils.GetFunctionHeadLoc(sourceFile, node)
	}

	var start int
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		start = functionDeclarationHeadStart(sourceText, node)
	case ast.KindMethodDeclaration:
		start = startSkippingDecorators(sourceText, node)
	case ast.KindFunctionExpression:
		parent := node.Parent
		if parent != nil && (parent.Kind == ast.KindPropertyAssignment || parent.Kind == ast.KindPropertyDeclaration) {
			start = startSkippingDecorators(sourceText, parent)
		} else {
			start = scanner.SkipTrivia(sourceText, node.Pos())
		}
	default:
		return utils.GetFunctionHeadLoc(sourceFile, node)
	}
	if start < 0 || start > openParen {
		return utils.GetFunctionHeadLoc(sourceFile, node)
	}
	return core.NewTextRange(start, openParen)
}

// https://eslint.org/docs/latest/rules/require-yield
var RequireYieldRule = rule.Rule{
	Name: "require-yield",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		if !strings.Contains(ctx.SourceFile.Text(), "*") {
			return nil
		}

		stack := make([]stackFrame, 0, 8)

		enter := func(node *ast.Node) {
			stack = append(stack, stackFrame{node: node})
		}

		exit := func(node *ast.Node) {
			n := len(stack)
			if n == 0 {
				return
			}
			top := stack[n-1]
			stack = stack[:n-1]
			if isGenerator(top.node) && !top.hasYield && utils.HasNonEmptyFunctionBody(top.node) {
				ctx.ReportRange(
					generatorHeadRange(ctx.SourceFile, top.node),
					missingYieldMessage,
				)
			}
		}

		countYield := func(node *ast.Node) {
			for i := len(stack) - 1; i >= 0; i-- {
				bp, be, ok := utils.BodyLikeRange(stack[i].node)
				if !ok {
					continue
				}
				if node.Pos() >= bp && node.End() <= be {
					stack[i].hasYield = true
					return
				}
			}
		}

		return rule.RuleListeners{
			ast.KindFunctionDeclaration:                              enter,
			rule.ListenerOnExit(ast.KindFunctionDeclaration):         exit,
			ast.KindFunctionExpression:                               enter,
			rule.ListenerOnExit(ast.KindFunctionExpression):          exit,
			ast.KindMethodDeclaration:                                enter,
			rule.ListenerOnExit(ast.KindMethodDeclaration):           exit,
			ast.KindArrowFunction:                                    enter,
			rule.ListenerOnExit(ast.KindArrowFunction):               exit,
			ast.KindGetAccessor:                                      enter,
			rule.ListenerOnExit(ast.KindGetAccessor):                 exit,
			ast.KindSetAccessor:                                      enter,
			rule.ListenerOnExit(ast.KindSetAccessor):                 exit,
			ast.KindConstructor:                                      enter,
			rule.ListenerOnExit(ast.KindConstructor):                 exit,
			ast.KindPropertyDeclaration:                              enter,
			rule.ListenerOnExit(ast.KindPropertyDeclaration):         exit,
			ast.KindClassStaticBlockDeclaration:                      enter,
			rule.ListenerOnExit(ast.KindClassStaticBlockDeclaration): exit,

			ast.KindYieldExpression: countYield,
		}
	},
}
