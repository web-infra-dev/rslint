package no_useless_else

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Recognize a try block whose prefix consists of proven-safe operations and
// whose last statement exits. The shared graph may consider the prefix
// throwable even for a simple parameter assignment or a skipped access.
func (a *exitAnalysis) exitsAfterSafeOperations(node *ast.Node) bool {
	if node.Kind == ast.KindBlock {
		stmts := node.Statements()
		if len(stmts) != 1 {
			return false
		}
		node = stmts[0]
	}
	if node.Kind != ast.KindTryStatement {
		return false
	}
	stmt := node.AsTryStatement()
	if stmt.CatchClause == nil || stmt.FinallyBlock != nil {
		return false
	}
	stmts := stmt.TryBlock.Statements()
	if len(stmts) == 0 {
		return false
	}
	last := stmts[len(stmts)-1]
	switch last.Kind {
	case ast.KindReturnStatement:
		if last.Expression() != nil && !a.safePrimitiveRead(last.Expression()) {
			return false
		}
	case ast.KindExpressionStatement:
		if !a.simpleExitExpression(last.Expression()) {
			return false
		}
	default:
		return false
	}
	for _, prefix := range stmts[:len(stmts)-1] {
		if !a.safeExpressionStatement(prefix) {
			return false
		}
	}
	return true
}

func (a *exitAnalysis) simpleExitExpression(expression *ast.Node) bool {
	expression = utils.SkipAssertionsAndParens(expression)
	if expression == nil {
		return false
	}
	switch expression.Kind {
	case ast.KindCallExpression:
		if ast.IsOptionalChain(expression) || !a.isProcessExit(expression) {
			return false
		}
		for _, argument := range expression.Arguments() {
			if !a.safePrimitiveRead(argument) {
				return false
			}
		}
		return true
	case ast.KindTemplateExpression:
		for _, span := range expression.AsTemplateExpression().TemplateSpans.Nodes {
			if a.simpleExitExpression(span.Expression()) {
				return true
			}
			if !a.safeTemplateSubstitution(span.Expression()) {
				break
			}
		}
	}
	return false
}

func (a *exitAnalysis) safeExpressionStatement(node *ast.Node) bool {
	if node.Kind != ast.KindExpressionStatement {
		return false
	}
	expression := utils.SkipAssertionsAndParens(node.Expression())
	if expression == nil {
		return false
	}
	if ast.IsOptionalChain(expression) {
		return a.skipsOptionalChain(expression)
	}
	if !ast.IsAssignmentExpression(expression, true) {
		return false
	}
	assignment := expression.AsBinaryExpression()
	target := utils.SkipAssertionsAndParens(assignment.Left)
	value := utils.SkipAssertionsAndParens(assignment.Right)
	if target == nil || target.Kind != ast.KindIdentifier || target.Text() == "process" ||
		value == nil || !ast.IsPrimitiveLiteralValue(value, true) {
		return false
	}
	symbol := a.ctx.Refs.ResolveInFile(target)
	return symbol != nil && symbol.ValueDeclaration != nil &&
		symbol.ValueDeclaration.Kind == ast.KindParameter && parameterReadIsSafe(target, symbol.ValueDeclaration)
}

func (a *exitAnalysis) skipsOptionalChain(expression *ast.Node) bool {
	for ast.IsOptionalChain(expression) {
		receiver := expression.Expression()
		if ast.IsOptionalChainRoot(expression) && a.safePrimitiveRead(receiver) {
			if nullish, known := a.nullish(receiver); known && nullish {
				return true
			}
		}
		expression = receiver
	}
	return false
}

// Restrict the additional proof to literal values and initialized
// identifiers. A folded value alone does not prove its evaluation is safe.
func (a *exitAnalysis) safePrimitiveRead(expression *ast.Node) bool {
	expression = utils.SkipAssertionsAndParens(expression)
	if expression == nil {
		return false
	}
	switch expression.Kind {
	case ast.KindIdentifier:
		return a.safeRead(expression)
	case ast.KindNullKeyword:
		return true
	case ast.KindVoidExpression:
		return a.safePrimitiveRead(expression.Expression())
	default:
		return ast.IsPrimitiveLiteralValue(expression, true)
	}
}
