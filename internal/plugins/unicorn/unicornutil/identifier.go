package unicornutil

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// PlainParameterIdentifier returns the identifier declared by a parameter
// whose ESTree shape is a plain Identifier. Type annotations are allowed;
// rest and default parameters are not.
func PlainParameterIdentifier(parameter *ast.Node) *ast.Node {
	if !ast.IsParameterDeclaration(parameter) {
		return nil
	}
	declaration := parameter.AsParameterDeclaration()
	if declaration == nil || declaration.DotDotDotToken != nil ||
		declaration.Initializer != nil {
		return nil
	}
	name := declaration.Name()
	if name == nil || !ast.IsIdentifier(name) {
		return nil
	}
	return name
}

// IsSameIdentifier reports whether two expressions are identifiers with the
// same name. Parentheses are transparent because ESTree does not preserve
// them; TypeScript assertion wrappers deliberately remain significant.
func IsSameIdentifier(left *ast.Node, right *ast.Node) bool {
	left = ast.SkipParentheses(left)
	right = ast.SkipParentheses(right)
	return left != nil && right != nil &&
		ast.IsIdentifier(left) && ast.IsIdentifier(right) &&
		left.AsIdentifier().Text == right.AsIdentifier().Text
}

// IsSameReference uses Unicorn's scope-free constant-key comparison for member
// paths. The core helper intentionally only recognizes literal property keys;
// Unicorn additionally equates e.g. object["a" + "b"] with object.ab.
// All non-member comparisons retain the shared helper's literal semantics.
func IsSameReference(left, right *ast.Node) bool {
	return isSameReference(left, right, false, nil)
}

// IsSameComparisonOperand additionally normalizes computed TypeScript keys and
// static string literals, matching Unicorn's comparison.isSame operation.
func IsSameComparisonOperand(left, right *ast.Node) bool {
	return isSameReference(left, right, true, nil)
}

// IsSameComparisonOperandWithKeyEvaluator lets a rule supply its own static
// evaluation for computed keys while retaining Unicorn's reference comparison.
func IsSameComparisonOperandWithKeyEvaluator(left, right *ast.Node, evaluateKey func(*ast.Node) (string, bool)) bool {
	return isSameReference(left, right, true, evaluateKey)
}

func isSameReference(left, right *ast.Node, comparison bool, evaluateKey func(*ast.Node) (string, bool)) bool {
	left, right = utils.SkipAssertionsAndParens(left), utils.SkipAssertionsAndParens(right)
	if left == nil || right == nil {
		return false
	}
	if comparison {
		left = ast.SkipOuterExpressions(left, ast.OEKParentheses|ast.OEKAssertions|ast.OEKExpressionsWithTypeArguments)
		right = ast.SkipOuterExpressions(right, ast.OEKParentheses|ast.OEKAssertions|ast.OEKExpressionsWithTypeArguments)
		if value, ok := utils.GetStaticStringLiteralValue(left); ok {
			other, known := utils.GetStaticStringLiteralValue(right)
			return known && value == other
		}
	}
	if !ast.IsAccessExpression(left) || !ast.IsAccessExpression(right) {
		return utils.IsSameReference(left, right, true)
	}
	if !isSameReference(left.Expression(), right.Expression(), comparison, evaluateKey) {
		return false
	}
	if name, known := referencePropertyName(left, comparison, evaluateKey); known {
		other, known := referencePropertyName(right, comparison, evaluateKey)
		return known && name == other
	}
	if left.Kind != right.Kind {
		return false
	}
	_, leftKey := utils.MemberExpressionParts(left)
	_, rightKey := utils.MemberExpressionParts(right)
	return isSameReference(leftKey, rightKey, comparison, evaluateKey)
}

func referencePropertyName(node *ast.Node, comparison bool, evaluateKey func(*ast.Node) (string, bool)) (string, bool) {
	if node.Kind == ast.KindPropertyAccessExpression {
		return utils.AccessExpressionStaticName(node)
	}
	key := utils.ESTreeRuntimeExpression(node.AsElementAccessExpression().ArgumentExpression)
	if comparison {
		key = ast.SkipOuterExpressions(key, ast.OEKParentheses|ast.OEKAssertions|ast.OEKExpressionsWithTypeArguments)
	} else if _, wrapped := utils.TransparentExpression(key); wrapped {
		return "", false
	}
	if evaluateKey != nil {
		return evaluateKey(key)
	}
	return utils.NewStaticStringEvaluatorWithoutScope().EvalToString(key)
}

// IsGlobalReference reports whether an identifier refers to an environment
// global rather than an authored declaration in the current file.
func IsGlobalReference(ctx rule.RuleContext, identifier *ast.Node) bool {
	if identifier == nil || !ast.IsIdentifier(identifier) {
		return false
	}
	if ctx.Refs != nil {
		return ctx.Refs.IsGlobalReference(identifier)
	}
	return !utils.IsShadowed(identifier, identifier.AsIdentifier().Text)
}
