package reactutil

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// LastReturnedExpression mirrors eslint-plugin-react's ast.loopNodes helper:
// scan top-level statements backwards, and when a switch is encountered recurse
// only into its final case. A non-empty trailing switch with no return in that
// case deliberately prevents falling back to earlier statements.
func LastReturnedExpression(statements []*ast.Node) *ast.Node {
	for i := len(statements) - 1; i >= 0; i-- {
		statement := statements[i]
		if statement == nil {
			continue
		}
		if statement.Kind == ast.KindReturnStatement {
			return statement.AsReturnStatement().Expression
		}
		if statement.Kind != ast.KindSwitchStatement {
			continue
		}
		switchStatement := statement.AsSwitchStatement()
		if switchStatement == nil || switchStatement.CaseBlock == nil {
			continue
		}
		caseBlock := switchStatement.CaseBlock.AsCaseBlock()
		if caseBlock == nil || caseBlock.Clauses == nil || len(caseBlock.Clauses.Nodes) == 0 {
			continue
		}
		lastClause := caseBlock.Clauses.Nodes[len(caseBlock.Clauses.Nodes)-1].AsCaseOrDefaultClause()
		if lastClause == nil || lastClause.Statements == nil {
			return nil
		}
		return LastReturnedExpression(lastClause.Statements.Nodes)
	}
	return nil
}

// LastReturnedExpressionInBody applies LastReturnedExpression to a function
// body and safely handles bodyless declarations.
func LastReturnedExpressionInBody(body *ast.Node) *ast.Node {
	if body == nil || body.Kind != ast.KindBlock || body.AsBlock().Statements == nil {
		return nil
	}
	return LastReturnedExpression(body.AsBlock().Statements.Nodes)
}

// LastDirectReturnedExpression scans top-level statements backwards without
// descending into switch clauses. ESLint uses this narrower behavior for
// propTypes getters.
func LastDirectReturnedExpression(body *ast.Node) (*ast.Node, bool) {
	if body == nil || body.Kind != ast.KindBlock || body.AsBlock().Statements == nil {
		return nil, false
	}
	statements := body.AsBlock().Statements.Nodes
	for index := len(statements) - 1; index >= 0; index-- {
		statement := statements[index]
		if statement != nil && statement.Kind == ast.KindReturnStatement {
			return statement.AsReturnStatement().Expression, true
		}
	}
	return nil, false
}
