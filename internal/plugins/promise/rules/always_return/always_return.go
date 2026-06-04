package always_return

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/promiseutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const skipTransparent = ast.OEKParentheses

type Options struct {
	IgnoreLastCallback       bool
	IgnoreAssignmentVariable []string
}

func parseOptions(options any) Options {
	opts := Options{
		IgnoreAssignmentVariable: []string{"globalThis"},
	}
	optsMap := utils.GetOptionsMap(options)
	if optsMap == nil {
		return opts
	}
	if v, ok := optsMap["ignoreLastCallback"].(bool); ok {
		opts.IgnoreLastCallback = v
	}
	if raw, ok := optsMap["ignoreAssignmentVariable"].([]interface{}); ok {
		vars := make([]string, 0, len(raw))
		for _, item := range raw {
			if s, ok := item.(string); ok {
				vars = append(vars, s)
			}
		}
		opts.IgnoreAssignmentVariable = vars
	}
	return opts
}

func buildThenShouldReturnOrThrowMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "thenShouldReturnOrThrow",
		Description: "Each then() should return a value or throw",
	}
}

// thenBlockCallback returns the callback node if node is a .then(cb, ...) call
// where cb is a FunctionExpression or ArrowFunction with a block body.
func thenBlockCallback(node *ast.Node) *ast.Node {
	if !promiseutil.IsMemberCall(node, "then") {
		return nil
	}
	args := node.Arguments()
	if len(args) == 0 {
		return nil
	}
	cb := ast.SkipOuterExpressions(args[0], skipTransparent)
	if cb == nil {
		return nil
	}
	if !isFunctionWithBlockBody(cb) {
		return nil
	}
	return cb
}

func isFunctionWithBlockBody(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionExpression:
		return true
	case ast.KindArrowFunction:
		body := node.Body()
		return body != nil && body.Kind == ast.KindBlock
	}
	return false
}

// allPathsTerminate reports whether every execution path through the given
// statement list ends with a return, throw, or process.exit/abort call.
// Nested function scopes are not entered.
func allPathsTerminate(stmts []*ast.Node) bool {
	for _, s := range stmts {
		if stmtTerminates(s) {
			return true
		}
	}
	return false
}

func stmtTerminates(s *ast.Node) bool {
	if s == nil {
		return false
	}
	switch s.Kind {
	case ast.KindReturnStatement, ast.KindThrowStatement:
		return true
	case ast.KindExpressionStatement:
		expr := s.AsExpressionStatement().Expression
		return isProcessExitOrAbort(expr)
	case ast.KindBlock:
		return allPathsTerminate(s.Statements())
	case ast.KindIfStatement:
		is := s.AsIfStatement()
		if is.ElseStatement == nil {
			return false
		}
		return stmtTerminates(is.ThenStatement) && stmtTerminates(is.ElseStatement)
	}
	return false
}

// isProcessExitOrAbort reports whether node is a call to process.exit() or process.abort().
func isProcessExitOrAbort(node *ast.Node) bool {
	if node == nil || !ast.IsCallExpression(node) {
		return false
	}
	callee := ast.SkipOuterExpressions(node.AsCallExpression().Expression, skipTransparent)
	if callee == nil || !ast.IsPropertyAccessExpression(callee) {
		return false
	}
	prop := callee.AsPropertyAccessExpression()
	if prop.QuestionDotToken != nil {
		return false
	}
	obj := ast.SkipOuterExpressions(prop.Expression, skipTransparent)
	if obj == nil || !ast.IsIdentifier(obj) || obj.AsIdentifier().Text != "process" {
		return false
	}
	name := prop.Name()
	if name == nil || !ast.IsIdentifier(name) {
		return false
	}
	n := name.AsIdentifier().Text
	return n == "exit" || n == "abort"
}

// isLastCallback mirrors eslint-plugin-promise's isLastCallback helper.
// It checks whether the .then(cb) call result is used as the terminal element
// of a chain (i.e., the call result is discarded or only wrapped in void/await).
func isLastCallback(thenCall *ast.Node) bool {
	target := thenCall
	for {
		parent := target.Parent
		// Skip parenthesized wrappers on the parent side.
		for parent != nil && ast.IsOuterExpression(parent, skipTransparent) {
			target = parent
			parent = target.Parent
		}
		if parent == nil {
			return false
		}

		switch parent.Kind {
		case ast.KindExpressionStatement:
			return true

		case ast.KindVoidExpression:
			return true

		case ast.KindAwaitExpression:
			target = parent
			continue

		case ast.KindBinaryExpression:
			bin := parent.AsBinaryExpression()
			if bin.OperatorToken.Kind == ast.KindCommaToken {
				// Sequence (comma) expression: if our target is NOT the right (last) operand,
				// the .then() result is discarded — this is the last callback.
				if bin.Right != target {
					return true
				}
				// We are the rightmost element; continue walking up.
				target = parent
				continue
			}
			return false

		case ast.KindPropertyAccessExpression:
			// Check for .catch() or .finally() chained after our .then() call.
			prop := parent.AsPropertyAccessExpression()
			if prop.Expression != target {
				return false
			}
			name := prop.Name()
			if name == nil || !ast.IsIdentifier(name) {
				return false
			}
			n := name.AsIdentifier().Text
			if n == "catch" || n == "finally" {
				// The property access must be the callee of a call expression.
				grandparent := parent.Parent
				for grandparent != nil && ast.IsOuterExpression(grandparent, skipTransparent) {
					grandparent = grandparent.Parent
				}
				if grandparent != nil && ast.IsCallExpression(grandparent) {
					callExpr := grandparent.AsCallExpression()
					callee := ast.SkipOuterExpressions(callExpr.Expression, skipTransparent)
					if callee == parent {
						target = grandparent
						continue
					}
				}
			}
			return false

		default:
			return false
		}
	}
}

// isIgnoredAssignment reports whether stmt is an assignment to a variable in ignoredVars.
// Mirrors upstream isIgnoredAssignment: checks ExpressionStatement > AssignmentExpression
// where the left-hand side root identifier is in the ignored list.
func isIgnoredAssignment(stmt *ast.Node, ignoredVars []string) bool {
	if stmt == nil || stmt.Kind != ast.KindExpressionStatement {
		return false
	}
	expr := stmt.AsExpressionStatement().Expression
	if expr == nil || !ast.IsBinaryExpression(expr) {
		return false
	}
	bin := expr.AsBinaryExpression()
	if !ast.IsAssignmentOperator(bin.OperatorToken.Kind) {
		return false
	}
	rootName := getRootObjectName(bin.Left)
	for _, v := range ignoredVars {
		if v == rootName {
			return true
		}
	}
	return false
}

// getRootObjectName returns the root identifier name of a member-access chain.
// Mirrors upstream getRootObjectName: Identifier → name, MemberExpression → recurse on object.
func getRootObjectName(node *ast.Node) string {
	if node == nil {
		return ""
	}
	node = ast.SkipOuterExpressions(node, skipTransparent)
	if node == nil {
		return ""
	}
	switch node.Kind {
	case ast.KindIdentifier:
		return node.AsIdentifier().Text
	case ast.KindPropertyAccessExpression:
		return getRootObjectName(node.AsPropertyAccessExpression().Expression)
	case ast.KindElementAccessExpression:
		return getRootObjectName(node.AsElementAccessExpression().Expression)
	}
	return ""
}

var AlwaysReturnRule = rule.Rule{
	Name: "promise/always-return",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				cb := thenBlockCallback(node)
				if cb == nil {
					return
				}

				if opts.IgnoreLastCallback && isLastCallback(node) {
					return
				}

				if len(opts.IgnoreAssignmentVariable) > 0 && isLastCallback(node) {
					body := cb.Body()
					if body != nil && body.Kind == ast.KindBlock {
						for _, stmt := range body.Statements() {
							if isIgnoredAssignment(stmt, opts.IgnoreAssignmentVariable) {
								return
							}
						}
					}
				}

				body := cb.Body()
				if body == nil {
					return
				}
				stmts := body.Statements()
				if !allPathsTerminate(stmts) {
					ctx.ReportNode(cb, buildThenShouldReturnOrThrowMessage())
				}
			},
		}
	},
}
