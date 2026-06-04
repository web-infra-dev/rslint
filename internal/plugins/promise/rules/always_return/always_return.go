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

// stmtsTerminate walks stmts and returns (terminates, reportNode):
//
//	(true,  nil)   — every execution path through stmts terminates.
//	(false, node)  — paths do not all terminate; node is the first branching
//	                 statement (if/switch/try/loop) to blame, matching upstream's
//	                 code-path-segment reporting.
//	(false, nil)   — paths do not terminate, but no specific sub-node can be
//	                 blamed (flat body); the caller should fall back to cb.
//
// Nested function scopes are not entered.
func stmtsTerminate(stmts []*ast.Node) (bool, *ast.Node) {
	var candidate *ast.Node
	for _, s := range stmts {
		ok, node := stmtTerminates(s)
		if ok {
			// A terminating statement was reached — all paths from here are
			// covered, including any incomplete paths left open by earlier
			// branching statements (e.g. if-without-else followed by return).
			return true, nil
		}
		// Record the first specific blame node; a later terminator may still
		// override the whole thing with (true, nil).
		if candidate == nil {
			candidate = node // nil for flat stmts, non-nil for branching ones
		}
	}
	return false, candidate
}

// stmtTerminates returns (terminates, reportNode):
//
//	(true,  nil)   — statement terminates all paths.
//	(false, s)     — branching statement that does not terminate; s is the node
//	                 to report on (the statement itself).
//	(false, nil)   — non-branching statement that does not terminate; no specific
//	                 blame sub-node; the caller propagates nil upward.
func stmtTerminates(s *ast.Node) (bool, *ast.Node) {
	if s == nil {
		return false, nil
	}
	switch s.Kind {
	case ast.KindReturnStatement, ast.KindThrowStatement:
		return true, nil
	case ast.KindExpressionStatement:
		if isProcessExitOrAbort(s.AsExpressionStatement().Expression) {
			return true, nil
		}
		return false, nil
	case ast.KindBlock:
		return stmtsTerminate(s.Statements())
	case ast.KindIfStatement:
		is := s.AsIfStatement()
		if is.ElseStatement == nil {
			return false, s
		}
		tOk, _ := stmtTerminates(is.ThenStatement)
		eOk, _ := stmtTerminates(is.ElseStatement)
		if tOk && eOk {
			return true, nil
		}
		return false, s
	case ast.KindSwitchStatement:
		sw := s.AsSwitchStatement()
		if sw.CaseBlock == nil {
			return false, s
		}
		hasDefault := false
		for _, clause := range sw.CaseBlock.AsCaseBlock().Clauses.Nodes {
			if clause.Kind == ast.KindDefaultClause {
				hasDefault = true
			}
			stmts := clause.Statements()
			if len(stmts) > 0 {
				if ok, _ := stmtsTerminate(stmts); !ok {
					return false, s
				}
			}
		}
		if !hasDefault {
			return false, s
		}
		return true, nil
	case ast.KindTryStatement:
		ts := s.AsTryStatement()
		if ts.TryBlock == nil {
			return false, s
		}
		if ok, _ := stmtsTerminate(ts.TryBlock.Statements()); !ok {
			return false, s
		}
		if ts.CatchClause != nil {
			cc := ts.CatchClause.AsCatchClause()
			if cc.Block == nil {
				return false, s
			}
			if ok, _ := stmtsTerminate(cc.Block.Statements()); !ok {
				return false, s
			}
		}
		return true, nil
	case ast.KindWhileStatement:
		ws := s.AsWhileStatement()
		if ws.Statement == nil {
			return false, s
		}
		if ok, _ := stmtTerminates(ws.Statement); !ok {
			return false, s
		}
		return true, nil
	case ast.KindForStatement:
		fs := s.AsForStatement()
		if fs.Statement == nil {
			return false, s
		}
		if ok, _ := stmtTerminates(fs.Statement); !ok {
			return false, s
		}
		return true, nil
	case ast.KindDoStatement:
		ds := s.AsDoStatement()
		if ds.Statement == nil {
			return false, s
		}
		if ok, _ := stmtTerminates(ds.Statement); !ok {
			return false, s
		}
		return true, nil
	}
	return false, nil
}

// isProcessExitOrAbort reports whether node is a call to process.exit() or process.abort().
func isProcessExitOrAbort(node *ast.Node) bool {
	if node == nil || !ast.IsCallExpression(node) {
		return false
	}
	call := node.AsCallExpression()
	if call.QuestionDotToken != nil {
		return false
	}
	callee := ast.SkipOuterExpressions(call.Expression, skipTransparent)
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
				if ok, reportOn := stmtsTerminate(body.Statements()); !ok {
					if reportOn == nil {
						reportOn = cb
					}
					ctx.ReportNode(reportOn, buildThenShouldReturnOrThrowMessage())
				}
			},
		}
	},
}
