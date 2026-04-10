package prefer_todo

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	rslintUtils "github.com/web-infra-dev/rslint/internal/utils"
)

// Message Builders

func buildEmptyTestErrorMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "emptyTest",
		Description: "Prefer todo test case over empty test case",
	}
}

func buildUnimplementedTestErrorMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unimplementedTest",
		Description: "Prefer todo test case over unimplemented test case",
	}
}

func isStringNode(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral, ast.KindTemplateExpression:
		return true
	default:
		return false
	}
}

func isEmptyFunction(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindArrowFunction:
		arr := node.AsArrowFunction()
		if arr.Body == nil {
			return false
		}
		if arr.Body.Kind == ast.KindBlock {
			block := arr.Body.AsBlock()
			return block != nil && (block.Statements == nil || len(block.Statements.Nodes) == 0)
		}
		return false
	case ast.KindFunctionExpression:
		fn := node.AsFunctionExpression()
		if fn.Body == nil || fn.Body.Kind != ast.KindBlock {
			return false
		}
		block := fn.Body.AsBlock()
		return block != nil && (block.Statements == nil || len(block.Statements.Nodes) == 0)
	default:
		return false
	}
}

func isTargetedTestCase(jestFn *utils.ParsedJestFnCall) bool {
	for _, m := range jestFn.Members {
		if m != "skip" {
			return false
		}
	}

	return !strings.HasPrefix(jestFn.Name, "x") && !strings.HasPrefix(jestFn.Name, "f")
}

func buildCalleeWithTodo(ctx rule.RuleContext, callExpr *ast.CallExpression, jestFn *utils.ParsedJestFnCall) string {
	src := ctx.SourceFile.Text()
	callee := ast.SkipParentheses(callExpr.Expression)
	if callee == nil {
		return ""
	}

	if len(jestFn.Members) == 0 {
		head := jestFn.Head.Local.Node
		if head == nil {
			return ""
		}
		headTrim := rslintUtils.TrimNodeTextRange(ctx.SourceFile, head)
		return src[headTrim.Pos():headTrim.End()] + ".todo"
	}
	entry := jestFn.MemberEntries[0]
	mem := entry.Node
	if mem == nil {
		return ""
	}

	calleeTrim := rslintUtils.TrimNodeTextRange(ctx.SourceFile, callee)
	prefix := src[calleeTrim.Pos():mem.Pos()]
	var mid string
	if mem.Kind == ast.KindIdentifier {
		mid = "todo"
	} else {
		mid = "'todo'"
	}
	suffix := src[mem.End():callee.End()]
	return prefix + mid + suffix
}

func buildTodoCallReplacementTitleOnly(ctx rule.RuleContext, callExpr *ast.CallExpression, jestFn *utils.ParsedJestFnCall, titleArg *ast.Node) string {
	src := ctx.SourceFile.Text()
	calleeTodo := buildCalleeWithTodo(ctx, callExpr, jestFn)
	titleText := src[titleArg.Pos():titleArg.End()]
	return fmt.Sprintf("%s(%s)", calleeTodo, titleText)
}

func buildTodoCallReplacementUnimplemented(ctx rule.RuleContext, callExpr *ast.CallExpression, jestFn *utils.ParsedJestFnCall) string {
	src := ctx.SourceFile.Text()
	args := callExpr.Arguments
	if args == nil {
		return ""
	}
	inner := src[args.Pos():args.End()]
	return fmt.Sprintf("%s(%s)", buildCalleeWithTodo(ctx, callExpr, jestFn), inner)
}

var PreferTodoRule = rule.Rule{
	Name: "jest/prefer-todo",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				jestFnCall := utils.ParseJestFnCall(node, ctx)
				if jestFnCall == nil || jestFnCall.Kind != utils.JestFnTypeTest {
					return
				}

				if !isTargetedTestCase(jestFnCall) {
					return
				}

				callExpr := node.AsCallExpression()
				if callExpr == nil || callExpr.Arguments == nil || len(callExpr.Arguments.Nodes) == 0 {
					return
				}

				args := callExpr.Arguments.Nodes
				if len(args) == 0 {
					return
				}

				title := args[0]
				if title == nil || !isStringNode(title) {
					return
				}

				if len(args) >= 2 && isEmptyFunction(args[1]) {
					newText := buildTodoCallReplacementTitleOnly(ctx, callExpr, jestFnCall, title)
					callRange := rslintUtils.TrimNodeTextRange(ctx.SourceFile, node)
					ctx.ReportNodeWithFixes(node, buildEmptyTestErrorMessage(), rule.RuleFixReplaceRange(callRange, newText))

				} else if len(args) == 1 {
					newText := buildTodoCallReplacementUnimplemented(ctx, callExpr, jestFnCall)
					callRange := rslintUtils.TrimNodeTextRange(ctx.SourceFile, node)
					ctx.ReportNodeWithFixes(node, buildUnimplementedTestErrorMessage(), rule.RuleFixReplaceRange(callRange, newText))
				}
			},
		}
	},
}
