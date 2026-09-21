package test_framework

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// StaticNumericLiteral reports whether node is a number written directly in
// source, looking through parentheses and the TypeScript wrappers that preserve
// the runtime value. zero and negativeZero describe that number, so callers
// that only need "is this provably a number" can ignore them.
//
// Matcher rules need this because a matcher argument that is not a literal may
// hold a string or a boolean at run time, and matchers do not agree on how they
// compare such values.
func StaticNumericLiteral(node *ast.Node) (zero, negativeZero, ok bool) {
	for node != nil {
		node = ast.SkipParentheses(node)
		switch node.Kind {
		case ast.KindAsExpression:
			node = node.AsAsExpression().Expression
		case ast.KindTypeAssertionExpression:
			node = node.AsTypeAssertion().Expression
		case ast.KindSatisfiesExpression:
			node = node.AsSatisfiesExpression().Expression
		case ast.KindNonNullExpression:
			node = node.AsNonNullExpression().Expression
		case ast.KindNumericLiteral:
			return utils.NormalizeNumericLiteral(node.AsNumericLiteral().Text) == "0", false, true
		case ast.KindPrefixUnaryExpression:
			unary := node.AsPrefixUnaryExpression()
			if unary == nil || (unary.Operator != ast.KindPlusToken && unary.Operator != ast.KindMinusToken) {
				return false, false, false
			}
			zero, negativeZero, ok := StaticNumericLiteral(unary.Operand)
			if !ok {
				return false, false, false
			}
			if zero && unary.Operator == ast.KindMinusToken {
				negativeZero = !negativeZero
			}
			return zero, negativeZero, true
		default:
			return false, false, false
		}
	}
	return false, false, false
}

// IsSideEffectFreeLiteral reports whether node is a literal whose evaluation
// cannot run user code, looking through the same wrappers as
// StaticNumericLiteral. Rules use it for arguments they keep in place while
// moving the evaluation point of the assertion's subject.
func IsSideEffectFreeLiteral(node *ast.Node) bool {
	for node != nil {
		node = ast.SkipParentheses(node)
		switch node.Kind {
		case ast.KindAsExpression:
			node = node.AsAsExpression().Expression
		case ast.KindTypeAssertionExpression:
			node = node.AsTypeAssertion().Expression
		case ast.KindSatisfiesExpression:
			node = node.AsSatisfiesExpression().Expression
		case ast.KindNonNullExpression:
			node = node.AsNonNullExpression().Expression
		case ast.KindStringLiteral,
			ast.KindNoSubstitutionTemplateLiteral,
			ast.KindNumericLiteral,
			ast.KindBigIntLiteral,
			ast.KindTrueKeyword,
			ast.KindFalseKeyword,
			ast.KindNullKeyword:
			return true
		default:
			return false
		}
	}
	return false
}
