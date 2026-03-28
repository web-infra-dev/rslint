package no_extra_bind

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// isThisInComputedPropertyName checks whether a ThisKeyword node is inside a
// ComputedPropertyName that belongs to a MethodDeclaration, GetAccessor,
// SetAccessor, or Constructor. In that case, `this` is evaluated in the outer
// scope, not the method's own `this` scope.
//
// Walk up the parent chain; stop at nodes that create their own `this` binding
// (functions, methods, etc.) but NOT at ArrowFunction (arrows inherit `this`).
func isThisInComputedPropertyName(node *ast.Node) bool {
	for current := node.Parent; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindComputedPropertyName:
			if p := current.Parent; p != nil {
				switch p.Kind {
				case ast.KindMethodDeclaration, ast.KindGetAccessor,
					ast.KindSetAccessor, ast.KindConstructor:
					return true
				}
			}
			return false
		case ast.KindFunctionExpression, ast.KindFunctionDeclaration,
			ast.KindMethodDeclaration, ast.KindGetAccessor,
			ast.KindSetAccessor, ast.KindConstructor:
			return false
		}
	}
	return false
}

// isSideEffectFree returns true if the node is guaranteed to have no side effects.
// Matches ESLint's SIDE_EFFECT_FREE_NODE_TYPES: Literal, Identifier, ThisExpression, FunctionExpression.
func isSideEffectFree(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral,
		ast.KindRegularExpressionLiteral, ast.KindNoSubstitutionTemplateLiteral,
		ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword:
		return true
	case ast.KindIdentifier, ast.KindThisKeyword, ast.KindFunctionExpression:
		return true
	}
	return false
}

// bindMatch holds all the information about a detected .bind() call needed for
// both reporting and autofix.
type bindMatch struct {
	callNode   *ast.Node // the outermost CallExpression
	memberNode *ast.Node // the PropertyAccessExpression or ElementAccessExpression
}

// https://eslint.org/docs/latest/rules/no-extra-bind
var NoExtraBindRule = rule.Rule{
	Name: "no-extra-bind",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		type scopeInfo struct {
			match     *bindMatch
			thisFound bool
			upper     *scopeInfo
		}

		var scope *scopeInfo

		// getBindMatch checks if funcNode is the callee of a .bind() call.
		// Returns a bindMatch if so, otherwise nil.
		getBindMatch := func(funcNode *ast.Node) *bindMatch {
			current := funcNode.Parent
			for current != nil && current.Kind == ast.KindParenthesizedExpression {
				current = current.Parent
			}
			if current == nil {
				return nil
			}

			var memberNode *ast.Node
			switch current.Kind {
			case ast.KindPropertyAccessExpression:
				propAccess := current.AsPropertyAccessExpression()
				if propAccess == nil || propAccess.Name() == nil {
					return nil
				}
				if propAccess.Name().Text() != "bind" {
					return nil
				}
				memberNode = current
			case ast.KindElementAccessExpression:
				elemAccess := current.AsElementAccessExpression()
				if elemAccess == nil || elemAccess.ArgumentExpression == nil {
					return nil
				}
				arg := elemAccess.ArgumentExpression
				if arg.Kind == ast.KindStringLiteral && arg.Text() == "bind" {
					memberNode = current
				} else if arg.Kind == ast.KindNoSubstitutionTemplateLiteral && arg.Text() == "bind" {
					memberNode = current
				} else {
					return nil
				}
			default:
				return nil
			}

			callParent := memberNode.Parent
			for callParent != nil && callParent.Kind == ast.KindParenthesizedExpression {
				callParent = callParent.Parent
			}
			if callParent == nil || callParent.Kind != ast.KindCallExpression {
				return nil
			}

			callExpr := callParent.AsCallExpression()
			if callExpr == nil {
				return nil
			}

			callee := callExpr.Expression
			for callee != nil && callee.Kind == ast.KindParenthesizedExpression {
				callee = callee.AsParenthesizedExpression().Expression
			}
			if callee != memberNode {
				return nil
			}

			argCount := 0
			if callExpr.Arguments != nil {
				argCount = len(callExpr.Arguments.Nodes)
			}
			if argCount != 1 {
				return nil
			}

			firstArg := callExpr.Arguments.Nodes[0]
			if firstArg.Kind == ast.KindSpreadElement {
				return nil
			}

			return &bindMatch{callNode: callParent, memberNode: memberNode}
		}

		msg := rule.RuleMessage{
			Id:          "unexpected",
			Description: "The function binding is unnecessary.",
		}

		// buildFixes produces the autofix for removing .bind(arg).
		// It removes two ranges:
		//   1. From the member-access token (. or [) to end of memberNode → removes ".bind" / "['bind']"
		//   2. From end of the call's callee expression to end of callNode → removes "(arg)"
		// Returns nil if the argument has side effects (matching ESLint behavior).
		buildFixes := func(m *bindMatch) []rule.RuleFix {
			callExpr := m.callNode.AsCallExpression()
			firstArg := callExpr.Arguments.Nodes[0]
			if !isSideEffectFree(firstArg) {
				return nil
			}

			// The expression to keep (left side of the member access)
			var keepExpr *ast.Node
			switch m.memberNode.Kind {
			case ast.KindPropertyAccessExpression:
				keepExpr = m.memberNode.AsPropertyAccessExpression().Expression
			case ast.KindElementAccessExpression:
				keepExpr = m.memberNode.AsElementAccessExpression().Expression
			}
			if keepExpr == nil {
				return nil
			}

			// Find the actual '.' or '[' token, skipping any comments between
			// the object expression and the member-access operator.
			sourceText := ctx.SourceFile.Text()
			tokenStart := scanner.SkipTrivia(sourceText, keepExpr.End())

			// Range 1: remove the member-access part (.bind / ['bind'] / [`bind`])
			fix1 := rule.RuleFixRemoveRange(core.NewTextRange(tokenStart, m.memberNode.End()))
			// Range 2: remove the call arguments part ((arg))
			fix2 := rule.RuleFixRemoveRange(core.NewTextRange(callExpr.Expression.End(), m.callNode.End()))
			return []rule.RuleFix{fix1, fix2}
		}

		report := func(m *bindMatch) {
			if fixes := buildFixes(m); len(fixes) > 0 {
				ctx.ReportNodeWithFixes(m.callNode, msg, fixes...)
			} else {
				ctx.ReportNode(m.callNode, msg)
			}
		}

		enterFunction := func(node *ast.Node) {
			scope = &scopeInfo{
				match:     getBindMatch(node),
				thisFound: false,
				upper:     scope,
			}
		}

		exitFunction := func(node *ast.Node) {
			if scope != nil {
				if scope.match != nil && !scope.thisFound {
					report(scope.match)
				}
				scope = scope.upper
			}
		}

		enterThisScope := func(node *ast.Node) {
			scope = &scopeInfo{
				match:     nil,
				thisFound: false,
				upper:     scope,
			}
		}

		exitThisScope := func(node *ast.Node) {
			if scope != nil {
				scope = scope.upper
			}
		}

		return rule.RuleListeners{
			ast.KindFunctionExpression:                       enterFunction,
			rule.ListenerOnExit(ast.KindFunctionExpression):  exitFunction,
			ast.KindFunctionDeclaration:                      enterFunction,
			rule.ListenerOnExit(ast.KindFunctionDeclaration): exitFunction,

			ast.KindMethodDeclaration:                       enterThisScope,
			rule.ListenerOnExit(ast.KindMethodDeclaration):  exitThisScope,
			ast.KindGetAccessor:                             enterThisScope,
			rule.ListenerOnExit(ast.KindGetAccessor):        exitThisScope,
			ast.KindSetAccessor:                             enterThisScope,
			rule.ListenerOnExit(ast.KindSetAccessor):        exitThisScope,
			ast.KindConstructor:                             enterThisScope,
			rule.ListenerOnExit(ast.KindConstructor):        exitThisScope,

			ast.KindArrowFunction: func(node *ast.Node) {
				if m := getBindMatch(node); m != nil {
					report(m)
				}
			},

			ast.KindThisKeyword: func(node *ast.Node) {
				if scope == nil {
					return
				}
				target := scope
				if isThisInComputedPropertyName(node) && scope.upper != nil {
					target = scope.upper
				}
				target.thisFound = true
			},
		}
	},
}
