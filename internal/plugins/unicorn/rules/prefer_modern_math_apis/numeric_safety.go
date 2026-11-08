package prefer_modern_math_apis

import (
	"math"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/jsnum"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

const numberWrappers = ast.OEKParentheses | ast.OEKAssertions | ast.OEKExpressionsWithTypeArguments

// Match the upstream rule's deliberately limited evaluation: one const
// initializer, builtin constants by name, and primitive literals. Folding more
// expressions would turn upstream suggestions into automatic changes.
func staticNumber(ctx rule.RuleContext, node *ast.Node) (float64, bool) {
	node = ast.SkipOuterExpressions(node, numberWrappers)
	// Use the helper's bound lexical lookup: the type checker can select a
	// builtin symbol instead of an authored const with the same name.
	if initializer := utils.GetConstVariableInitializer(node, nil); initializer != nil &&
		initializer.Parent != nil && ast.IsIdentifier(initializer.Parent.Name()) {
		node = ast.SkipOuterExpressions(initializer, numberWrappers)
	}
	argument := node
	sign := 1.0
	if node.Kind == ast.KindPrefixUnaryExpression {
		unary := node.AsPrefixUnaryExpression()
		if unary.Operator == ast.KindPlusToken || unary.Operator == ast.KindMinusToken {
			argument = ast.SkipOuterExpressions(unary.Operand, numberWrappers)
			if unary.Operator == ast.KindMinusToken {
				sign = -1
			}
		}
	}
	if ast.IsIdentifier(argument) {
		switch argument.Text() {
		case "Infinity", "NaN", "undefined":
			if !ctx.Globals.Access(argument.Text()).IsDeclared() || !unicornutil.IsGlobalReference(ctx, argument) {
				return 0, false
			}
			if argument.Text() == "Infinity" {
				return sign * math.Inf(1), true
			}
			return math.NaN(), true
		}
	}
	if ast.IsPropertyAccessExpression(argument) && !ast.IsOptionalChain(argument) {
		member := argument.AsPropertyAccessExpression()
		object := utils.ESTreeRuntimeExpression(member.Expression)
		if ast.IsIdentifier(object) && ast.IsIdentifier(member.Name()) {
			if value, ok := utils.BuiltinNumberConstant(object.Text(), member.Name().Text()); ok {
				if !ctx.Globals.Access(object.Text()).IsDeclared() || !unicornutil.IsGlobalReference(ctx, object) {
					return 0, false
				}
				return sign * value, true
			}
		}
	}
	if argument != node {
		if ast.IsNumericLiteral(argument) {
			return sign * float64(jsnum.FromString(argument.Text())), true
		}
		return 0, false
	}
	switch node.Kind {
	case ast.KindNumericLiteral:
		return float64(jsnum.FromString(node.Text())), true
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		value, ok := ecmascript.StringToNumber(node.Text())
		if !ok {
			value = math.NaN()
		}
		return value, true
	case ast.KindTrueKeyword:
		return 1, true
	case ast.KindFalseKeyword, ast.KindNullKeyword:
		return 0, true
	default:
		return 0, false
	}
}

func isFinite(value float64) bool { return !math.IsInf(value, 0) && !math.IsNaN(value) }

func canAutoFixSquares(ctx rule.RuleContext, node *ast.Node, squares []*ast.Node) bool {
	var singleValue float64
	hasInfinity, hasNaN, allZero := false, false, true
	for _, square := range squares {
		value, ok := staticNumber(ctx, square.AsBinaryExpression().Left)
		if !ok {
			return false
		}
		singleValue = value
		hasInfinity = hasInfinity || math.IsInf(value, 0)
		hasNaN = hasNaN || math.IsNaN(value)
		allZero = allZero && value == 0
	}
	if isNumericValueUnused(node) {
		return true
	}
	if len(squares) > 1 {
		if hasInfinity {
			return !hasNaN
		}
		return allZero || hasNaN
	}
	value := singleValue
	if squares[0].AsBinaryExpression().OperatorToken.Kind == ast.KindAsteriskAsteriskToken {
		return value == 0 || !isFinite(value)
	}
	// The product must round to a double before sqrt, as it does in JavaScript.
	square := float64(value * value)
	result, replacement := math.Sqrt(square), math.Abs(value)
	return result == replacement || math.IsNaN(result) && math.IsNaN(replacement)
}

func isNumericValueUnused(node *ast.Node) bool {
	for node != nil {
		expression := node
		for expression.Parent != nil && ast.IsOuterExpression(expression.Parent, numberWrappers) &&
			expression.Parent.Expression() == expression {
			expression = expression.Parent
		}
		node = expression.Parent
		if node == nil {
			return false
		}
		switch node.Kind {
		case ast.KindVoidExpression, ast.KindTypeOfExpression, ast.KindDeleteExpression:
			return true
		case ast.KindForStatement:
			loop := node.AsForStatement()
			return loop.Initializer == expression || loop.Incrementor == expression
		case ast.KindConditionalExpression:
			if node.AsConditionalExpression().Condition == expression {
				return false
			}
		case ast.KindBinaryExpression:
			binary := node.AsBinaryExpression()
			switch binary.OperatorToken.Kind {
			case ast.KindCommaToken:
				if binary.Left == expression {
					return true
				}
			case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
				if binary.Left == expression {
					return false
				}
			default:
				return false
			}
		default:
			return false
		}
	}
	return false
}
