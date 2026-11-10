package prefer_code_point

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var (
	messageErrorCharCodeAt = rule.RuleMessage{
		Id:          "error/charCodeAt",
		Description: "Prefer `String#codePointAt()` over `String#charCodeAt()`.",
	}
	messageErrorFromCharCode = rule.RuleMessage{
		Id:          "error/fromCharCode",
		Description: "Prefer `String.fromCodePoint()` over `String.fromCharCode()`.",
	}
	messageSuggestionCodePointAt = rule.RuleMessage{
		Id:          "suggestion/codePointAt",
		Description: "Use `String#codePointAt()`.",
	}
	messageSuggestionFromCodePoint = rule.RuleMessage{
		Id:          "suggestion/fromCodePoint",
		Description: "Use `String.fromCodePoint()`.",
	}
)

// PreferCodePointRule follows eslint-plugin-unicorn v79.0.0.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/prefer-code-point.js
var PreferCodePointRule = rule.Rule{
	Name:   "unicorn/prefer-code-point",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				// `isMethodCall(..., {method: 'charCodeAt', optionalCall: false})`:
				// an optional member (`string?.charCodeAt(...)`) still matches,
				// while an optional call (`string.charCodeAt?.(...)`) does not.
				call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
					Method:              "charCodeAt",
					AllowOptionalMember: true,
				})
				if !ok {
					return
				}

				target := call.Property
				if isResultUsedNumerically(node) {
					ctx.ReportNode(target, messageErrorCharCodeAt)
					return
				}

				ctx.ReportNodeWithDeferredSuggestions(target, messageErrorCharCodeAt, func() []rule.RuleSuggestion {
					return []rule.RuleSuggestion{{
						Message: messageSuggestionCodePointAt,
						FixesArr: []rule.RuleFix{
							rule.RuleFixReplace(ctx.SourceFile, target, "codePointAt"),
						},
					}}
				})
			},
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				target, ok := matchStringFromCharCode(node)
				if !ok {
					return
				}

				ctx.ReportNodeWithDeferredSuggestions(target, messageErrorFromCharCode, func() []rule.RuleSuggestion {
					return []rule.RuleSuggestion{{
						Message: messageSuggestionFromCodePoint,
						FixesArr: []rule.RuleFix{
							rule.RuleFixReplace(ctx.SourceFile, target, "fromCodePoint"),
						},
					}}
				})
			},
		}
	},
}

// matchStringFromCharCode mirrors `isMemberExpression(node, {object: 'String',
// property: 'fromCharCode', optional: false})`. Parentheses and JSDoc casts are
// transparent in ESTree, so they are skipped, but authored TypeScript
// assertions keep their wrapper node and therefore do not match the `String`
// identifier constraint.
func matchStringFromCharCode(node *ast.Node) (*ast.Node, bool) {
	propertyAccess := node.AsPropertyAccessExpression()
	if propertyAccess == nil || propertyAccess.QuestionDotToken != nil {
		return nil, false
	}

	property := propertyAccess.Name()
	if property == nil || !ast.IsIdentifier(property) ||
		property.AsIdentifier().Text != "fromCharCode" {
		return nil, false
	}

	object := utils.ESTreeRuntimeExpression(propertyAccess.Expression)
	if object == nil || !ast.IsIdentifier(object) ||
		object.AsIdentifier().Text != "String" {
		return nil, false
	}

	return property, true
}

// isResultUsedNumerically mirrors upstream's `isResultUsedNumerically`: the
// diagnostic is always reported, but replacing `charCodeAt` with `codePointAt`
// is only safe when the value is not consumed as a number. The walk skips
// expressions that do not change the value's role (chains, logical operands,
// conditional branches and TypeScript-only wrappers) and then checks the
// consuming parent expression.
func isResultUsedNumerically(node *ast.Node) bool {
	current := numericContextNode(node)
	parent := utils.ESTreeParent(current)
	if parent == nil {
		return false
	}

	switch parent.Kind {
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		if binary == nil || binary.OperatorToken == nil {
			return false
		}

		operator := binary.OperatorToken.Kind
		if isCompoundNumericAssignment(operator) {
			return utils.ESTreeRuntimeExpression(binary.Right) == current
		}

		return isNumericBinaryOperator(operator)
	case ast.KindPrefixUnaryExpression:
		unary := parent.AsPrefixUnaryExpression()
		switch unary.Operator {
		case ast.KindPlusToken, ast.KindMinusToken, ast.KindTildeToken:
			return true
		}
	}

	return false
}

// numericContextNode is the upward counterpart of the upstream walk over
// `ChainExpression`, `LogicalExpression`, conditional branches and TypeScript
// expression wrappers. ESTree omits tsgo's parenthesized-expression nodes, so
// they are skipped through ESTreeParent.
func numericContextNode(node *ast.Node) *ast.Node {
	current := node
	for {
		parent := utils.ESTreeParent(current)
		if parent == nil {
			return current
		}

		switch parent.Kind {
		case ast.KindAsExpression, ast.KindTypeAssertionExpression,
			ast.KindSatisfiesExpression, ast.KindNonNullExpression,
			ast.KindExpressionWithTypeArguments:
			// TypeScript-only wrappers have no runtime effect; the current
			// expression is always the wrapper's value.
			current = parent
		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			if binary == nil || binary.OperatorToken == nil ||
				!isLogicalOperator(binary.OperatorToken.Kind) {
				return current
			}
			current = parent
		case ast.KindConditionalExpression:
			conditional := parent.AsConditionalExpression()
			if conditional == nil ||
				(utils.ESTreeRuntimeExpression(conditional.WhenTrue) != current &&
					utils.ESTreeRuntimeExpression(conditional.WhenFalse) != current) {
				return current
			}
			current = parent
		default:
			return current
		}
	}
}

func isLogicalOperator(operator ast.Kind) bool {
	switch operator {
	case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
		return true
	}
	return false
}

func isNumericBinaryOperator(operator ast.Kind) bool {
	switch operator {
	case ast.KindPlusToken, ast.KindMinusToken, ast.KindAsteriskToken,
		ast.KindSlashToken, ast.KindPercentToken, ast.KindAsteriskAsteriskToken,
		ast.KindAmpersandToken, ast.KindBarToken, ast.KindCaretToken,
		ast.KindLessThanLessThanToken, ast.KindGreaterThanGreaterThanToken,
		ast.KindGreaterThanGreaterThanGreaterThanToken:
		return true
	}
	return false
}

func isCompoundNumericAssignment(operator ast.Kind) bool {
	switch operator {
	case ast.KindPlusEqualsToken, ast.KindMinusEqualsToken,
		ast.KindAsteriskEqualsToken, ast.KindSlashEqualsToken,
		ast.KindPercentEqualsToken, ast.KindAsteriskAsteriskEqualsToken,
		ast.KindAmpersandEqualsToken, ast.KindBarEqualsToken,
		ast.KindCaretEqualsToken, ast.KindLessThanLessThanEqualsToken,
		ast.KindGreaterThanGreaterThanEqualsToken,
		ast.KindGreaterThanGreaterThanGreaterThanEqualsToken:
		return true
	}
	return false
}
