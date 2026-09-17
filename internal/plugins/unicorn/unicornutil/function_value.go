package unicornutil

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// IsNodeValueNotFunction mirrors upstream's is-node-value-not-function helper.
// It rejects callback arguments that cannot be a callback (literals, objects,
// arrays, etc.), while treating `.bind()` calls as possible functions.
func IsNodeValueNotFunction(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindArrayLiteralExpression,
		ast.KindObjectLiteralExpression,
		ast.KindClassExpression,
		ast.KindTemplateExpression,
		ast.KindNoSubstitutionTemplateLiteral,
		// ESTree's UnaryExpression covers `!x` / `-x` as well as `typeof x`,
		// `void x` and `delete x`; tsgo splits the last three into their own
		// kinds, so all five have to be listed here.
		ast.KindPrefixUnaryExpression,
		ast.KindPostfixUnaryExpression,
		ast.KindTypeOfExpression,
		ast.KindVoidExpression,
		ast.KindDeleteExpression,
		// Literals (ESTree collapses these into one `Literal` node type).
		ast.KindStringLiteral,
		ast.KindNumericLiteral,
		ast.KindBigIntLiteral,
		ast.KindRegularExpressionLiteral,
		ast.KindTrueKeyword,
		ast.KindFalseKeyword,
		ast.KindNullKeyword,
		// mostLikelyNotNodeTypes
		ast.KindAwaitExpression,
		ast.KindNewExpression,
		ast.KindTaggedTemplateExpression,
		ast.KindThisKeyword:
		return true
	case ast.KindBinaryExpression:
		// ESTree splits BinaryExpression / LogicalExpression / AssignmentExpression;
		// upstream lists BinaryExpression and AssignmentExpression as impossible,
		// LogicalExpression is not. Match by operator.
		operator := node.AsBinaryExpression().OperatorToken.Kind
		return isImpossibleBinaryOperator(operator)
	case ast.KindCallExpression:
		// ESTree wraps optional calls in ChainExpression; the call heuristic
		// must not classify those unknown values as non-functions.
		if ast.IsOptionalChain(node) {
			return false
		}
		// For ordinary calls, upstream only accepts a `.bind()` result as a callback.
		_, isBind := MatchDotMethodCall(node, DotMethodCallOptions{Method: "bind"})
		return !isBind
	}
	return utils.IsUndefinedIdentifier(node)
}

// isImpossibleBinaryOperator returns true for arithmetic / comparison / bitwise
// operators (ESTree BinaryExpression) and assignment operators (ESTree
// AssignmentExpression), but false for `&&` / `||` / `??` (ESTree
// LogicalExpression), matching upstream's impossible / most-likely-not sets.
func isImpossibleBinaryOperator(operator ast.Kind) bool {
	switch operator {
	case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken,
		ast.KindCommaToken:
		return false
	}
	return true
}
