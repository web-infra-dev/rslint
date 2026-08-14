package test_framework

import "github.com/microsoft/typescript-go/shim/ast"

// IsPromiseChainCall reports promise-chain method calls: the callee is a
// member access of then / catch / finally with a statically known name, and
// the argument count is within the method's arity — at least 1, at most 2 for
// then and at most 1 for catch / finally.
func IsPromiseChainCall(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	call := node.AsCallExpression()
	name := promiseChainMethodName(call.Expression)
	if name == "" {
		return false
	}

	argumentCount := 0
	if call.Arguments != nil {
		argumentCount = len(call.Arguments.Nodes)
	}
	if argumentCount == 0 {
		return false
	}

	switch name {
	case "then":
		return argumentCount < 3
	case "catch", "finally":
		return argumentCount < 2
	default:
		return false
	}
}

func promiseChainMethodName(callee *ast.Node) string {
	if callee == nil {
		return ""
	}
	callee = ast.SkipParentheses(callee)
	if callee == nil {
		return ""
	}

	switch callee.Kind {
	case ast.KindPropertyAccessExpression:
		name := callee.AsPropertyAccessExpression().Name()
		if name != nil && name.Kind == ast.KindIdentifier {
			return name.AsIdentifier().Text
		}
	case ast.KindElementAccessExpression:
		key := ast.SkipParentheses(callee.AsElementAccessExpression().ArgumentExpression)
		if key == nil {
			return ""
		}
		switch key.Kind {
		case ast.KindStringLiteral:
			return key.AsStringLiteral().Text
		case ast.KindNoSubstitutionTemplateLiteral:
			return key.AsNoSubstitutionTemplateLiteral().Text
		}
	}
	return ""
}

// AbruptCompletionPropagatesFailure reports whether a failure produced at node
// can escape boundary. An enclosing catch can suppress an awaited rejection or
// an explicit throw, but cannot intercept the later rejection of a returned
// promise. A finally block that may return, break, or continue suppresses any
// of those pending completions on that path.
func AbruptCompletionPropagatesFailure(node, boundary *ast.Node) bool {
	catchCanSuppress := node != nil &&
		(node.Kind == ast.KindAwaitExpression || node.Kind == ast.KindThrowStatement)
	current := node
	for current != nil && current.Parent != nil && current.Parent != boundary {
		parent := current.Parent
		if parent.Kind == ast.KindTryStatement {
			tryStatement := parent.AsTryStatement()
			if tryStatement != nil {
				if tryStatement.FinallyBlock != nil &&
					tryStatement.FinallyBlock != current &&
					statementMaySuppressFailure(tryStatement.FinallyBlock) {
					return false
				}
				if catchCanSuppress &&
					tryStatement.TryBlock == current && tryStatement.CatchClause != nil {
					catchClause := tryStatement.CatchClause.AsCatchClause()
					if catchClause == nil || !statementAlwaysThrows(catchClause.Block) {
						return false
					}
				}
			}
		}
		current = parent
	}
	return true
}

func statementAlwaysThrows(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindThrowStatement:
		return true
	case ast.KindBlock:
		for _, statement := range node.Statements() {
			if statementAlwaysThrows(statement) {
				return true
			}
			switch statement.Kind {
			case ast.KindReturnStatement, ast.KindBreakStatement, ast.KindContinueStatement:
				return false
			}
		}
		return false
	case ast.KindIfStatement:
		ifStatement := node.AsIfStatement()
		return ifStatement != nil &&
			ifStatement.ElseStatement != nil &&
			statementAlwaysThrows(ifStatement.ThenStatement) &&
			statementAlwaysThrows(ifStatement.ElseStatement)
	case ast.KindTryStatement:
		tryStatement := node.AsTryStatement()
		if tryStatement == nil {
			return false
		}
		if tryStatement.FinallyBlock != nil &&
			statementAlwaysThrows(tryStatement.FinallyBlock) {
			return true
		}
		return statementAlwaysThrows(tryStatement.TryBlock) &&
			(tryStatement.CatchClause == nil ||
				statementAlwaysThrows(tryStatement.CatchClause.AsCatchClause().Block))
	default:
		return false
	}
}

func statementMaySuppressFailure(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindReturnStatement, ast.KindBreakStatement, ast.KindContinueStatement:
		return true
	}
	if IsFunction(node) {
		return false
	}
	found := false
	node.ForEachChild(func(child *ast.Node) bool {
		if statementMaySuppressFailure(child) {
			found = true
			return true
		}
		return false
	})
	return found
}
