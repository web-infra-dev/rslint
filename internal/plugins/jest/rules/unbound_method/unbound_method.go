package unbound_method

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	unboundMethod "github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/unbound_method"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// toThrowMatchers invoke the value passed to expect.
var toThrowMatchers = map[string]bool{
	"toThrow":                            true,
	"toThrowError":                       true,
	"toThrowErrorMatchingSnapshot":       true,
	"toThrowErrorMatchingInlineSnapshot": true,
}

var UnboundMethodRule = rule.Rule{
	Name:             "jest/unbound-method",
	Schema:           unboundMethod.UnboundMethodRule.Schema,
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		var analysis *utils.JestCallAnalysis
		return unboundMethod.CreateListeners(ctx, options, func(member *ast.Node) bool {
			call := parentCallOfArgument(member)
			if call == nil {
				return false
			}
			if isJestMockedCallee(ctx.SourceFile, call.Expression()) {
				return true
			}
			if analysis == nil {
				analysis = utils.GetJestCallAnalysis(ctx)
			}
			parsed := analysis.ParseFnCall(findTopMostCallExpression(call))
			if parsed == nil {
				return false
			}
			switch parsed.Kind {
			case utils.JestFnTypeJest:
				return len(parsed.MemberEntries) != 0 &&
					parsed.MemberEntries[0].Node.Kind == ast.KindIdentifier &&
					parsed.MemberEntries[0].Name == "mocked"
			case utils.JestFnTypeExpect:
				return !toThrowMatchers[parsed.Matcher]
			default:
				return false
			}
		})
	},
}

// parentCallOfArgument returns the call member is passed to, if the call is
// member's ESTree parent. An optional chain ends in a ChainExpression and a
// dynamic import is an ImportExpression, so neither has a CallExpression
// parent. A callee is never reported by the base rule, so it is skipped here.
func parentCallOfArgument(member *ast.Node) *ast.Node {
	if ast.IsOptionalChain(member) {
		return nil
	}
	outer := member
	for outer.Parent != nil && ast.IsParenthesizedExpression(outer.Parent) {
		outer = outer.Parent
	}
	call := outer.Parent
	if call == nil || !ast.IsCallExpression(call) || ast.IsImportCall(call) || call.Expression() == outer {
		return nil
	}
	return call
}

// findTopMostCallExpression walks up member/call chains like
// eslint-plugin-jest's findTopMostCallExpression. Unlike the shared helper, it
// stops at the ESTree nodes tsgo does not have: the ChainExpression that ends
// an optional chain, and the ImportExpression of a dynamic import.
func findTopMostCallExpression(node *ast.Node) *ast.Node {
	top := node
	for child, parent := node, node.Parent; parent != nil; child, parent = parent, parent.Parent {
		if ast.IsOptionalChain(child) && (!ast.IsOptionalChain(parent) || parent.Expression() != child) {
			return top
		}
		switch {
		case ast.IsParenthesizedExpression(parent), ast.IsAccessExpression(parent):
		case ast.IsCallExpression(parent) && !ast.IsImportCall(parent):
			top = parent
		default:
			return top
		}
	}
	return top
}

// isJestMockedCallee matches a `jest.mocked` member callee by name alone,
// without resolving `jest`.
func isJestMockedCallee(sourceFile *ast.SourceFile, callee *ast.Node) bool {
	inner := ast.SkipParentheses(callee)
	// A parenthesized optional chain is a ChainExpression, not a MemberExpression.
	if inner != callee && ast.IsOptionalChain(inner) {
		return false
	}
	switch inner.Kind {
	case ast.KindPropertyAccessExpression:
		name := inner.AsPropertyAccessExpression().Name()
		if !ast.IsIdentifier(name) || name.Text() != "mocked" {
			return false
		}
	case ast.KindElementAccessExpression:
		if !isSupportedAccessor(sourceFile, inner.AsElementAccessExpression().ArgumentExpression, "mocked") {
			return false
		}
	default:
		return false
	}
	return isSupportedAccessor(sourceFile, inner.Expression(), "jest")
}

// isSupportedAccessor mirrors eslint-plugin-jest's isSupportedAccessor: an
// identifier, a string literal, or a template literal whose raw text is value.
func isSupportedAccessor(sourceFile *ast.SourceFile, node *ast.Node, value string) bool {
	node = ast.SkipParentheses(node)
	switch node.Kind {
	case ast.KindIdentifier:
		return node.Text() == value
	case ast.KindStringLiteral:
		return node.AsStringLiteral().Text == value
	case ast.KindNoSubstitutionTemplateLiteral:
		return scanner.GetSourceTextOfNodeFromSourceFile(sourceFile, node, false) == "`"+value+"`"
	default:
		return false
	}
}
