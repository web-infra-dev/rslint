package prefer_type_error

import (
	"regexp"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const messageID = "prefer-type-error"

var message = rule.RuleMessage{
	Id:          messageID,
	Description: "`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.",
}

var typeCheckIdentifiers = map[string]bool{
	"isArguments":       true,
	"isArray":           true,
	"isArrayBuffer":     true,
	"isArrayLike":       true,
	"isArrayLikeObject": true,
	"isBigInt":          true,
	"isBoolean":         true,
	"isBuffer":          true,
	"isDate":            true,
	"isElement":         true,
	"isError":           true,
	"isFinite":          true,
	"isFunction":        true,
	"isInteger":         true,
	"isLength":          true,
	"isMap":             true,
	"isNaN":             true,
	"isNative":          true,
	"isNil":             true,
	"isNull":            true,
	"isNumber":          true,
	"isObject":          true,
	"isObjectLike":      true,
	"isPlainObject":     true,
	"isPrototypeOf":     true,
	"isRegExp":          true,
	"isSafeInteger":     true,
	"isSet":             true,
	"isString":          true,
	"isSymbol":          true,
	"isTypedArray":      true,
	"isUndefined":       true,
	"isView":            true,
	"isWeakMap":         true,
	"isWeakSet":         true,
	"isWindow":          true,
	"isXMLDoc":          true,
}

var typeCheckGlobalIdentifiers = map[string]bool{
	"isNaN":    true,
	"isFinite": true,
}

// errorNameRegexp mirrors upstream's /^(?:[A-Z][\da-z]*)*Error$/.
//
// This is a repository-authored fixed pattern with the same meaning under RE2
// and JavaScript regexp semantics, so the standard regexp package is safe here.
var errorNameRegexp = regexp.MustCompile(`^(?:[A-Z][\da-z]*)*Error$`)

// PreferTypeErrorRule enforces TypeError for type-checking if branches.
//
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/rules/prefer-type-error.js
var PreferTypeErrorRule = rule.Rule{
	Name:   "unicorn/prefer-type-error",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindThrowStatement: func(node *ast.Node) {
				throwStatement := node.AsThrowStatement()
				errorConstructor := errorConstructorIdentifier(throwStatement.Expression)
				if errorConstructor == nil ||
					!ctx.Globals.Access("Error").IsDeclared() ||
					!unicornutil.IsGlobalReference(ctx, errorConstructor) {
					return
				}

				block := node.Parent
				if block == nil || block.Kind != ast.KindBlock ||
					len(block.AsBlock().Statements.Nodes) != 1 {
					return
				}

				ifStatementNode := block.Parent
				if ifStatementNode == nil || ifStatementNode.Kind != ast.KindIfStatement {
					return
				}
				ifStatement := ifStatementNode.AsIfStatement()
				if !isTypecheckingExpression(ctx, ifStatement.Expression, nil) {
					return
				}

				ctx.ReportNodeWithDeferredFixes(errorConstructor, message, func() []rule.RuleFix {
					if !ctx.Globals.Access("TypeError").IsDeclared() ||
						utils.IsShadowed(errorConstructor, "TypeError") {
						return nil
					}
					return []rule.RuleFix{
						rule.RuleFixReplace(ctx.SourceFile, errorConstructor, "TypeError"),
					}
				})
			},
		}
	},
}

func errorConstructorIdentifier(node *ast.Node) *ast.Node {
	node = ast.SkipParentheses(node)
	if node == nil || node.Kind != ast.KindNewExpression {
		return nil
	}
	expression := node.AsNewExpression()
	callee := ast.SkipParentheses(expression.Expression)
	if callee == nil || !ast.IsIdentifier(callee) || callee.Text() != "Error" {
		return nil
	}
	return callee
}

func isTypecheckingExpression(ctx rule.RuleContext, node *ast.Node, callExpression *ast.CallExpression) bool {
	node = ast.SkipParentheses(node)

	switch node.Kind {
	case ast.KindIdentifier:
		return isTypecheckingIdentifier(ctx, node, callExpression, false)

	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		return isTypecheckingMemberExpression(ctx, node, callExpression)

	case ast.KindCallExpression:
		call := node.AsCallExpression()
		return isTypecheckingExpression(ctx, call.Expression, call)

	case ast.KindTypeOfExpression:
		return true

	case ast.KindPrefixUnaryExpression:
		prefix := node.AsPrefixUnaryExpression()
		return prefix.Operator == ast.KindExclamationToken &&
			isTypecheckingExpression(ctx, prefix.Operand, nil)

	case ast.KindBinaryExpression:
		return isTypecheckingBinaryExpression(ctx, node.AsBinaryExpression(), callExpression)

	default:
		return false
	}
}

func isTypecheckingIdentifier(ctx rule.RuleContext, node *ast.Node, callExpression *ast.CallExpression, memberExpression bool) bool {
	if callExpression == nil || callExpression.Arguments == nil ||
		len(callExpression.Arguments.Nodes) == 0 ||
		node == nil || !ast.IsIdentifier(node) {
		return false
	}

	name := node.Text()
	if memberExpression {
		return typeCheckIdentifiers[name]
	}
	return typeCheckGlobalIdentifiers[name] &&
		ctx.Globals.Access(name).IsDeclared() &&
		unicornutil.IsGlobalReference(ctx, node)
}

func isTypecheckingMemberExpression(ctx rule.RuleContext, node *ast.Node, callExpression *ast.CallExpression) bool {
	property := accessExpressionProperty(node)
	if isTypecheckingIdentifier(ctx, property, callExpression, true) {
		return true
	}

	object := utils.AccessExpressionObject(node)
	if object != nil && ast.IsAccessExpression(object) {
		return isTypecheckingMemberExpression(ctx, object, callExpression)
	}
	return false
}

func accessExpressionProperty(node *ast.Node) *ast.Node {
	if node.Kind == ast.KindPropertyAccessExpression {
		return node.AsPropertyAccessExpression().Name()
	}
	return node.AsElementAccessExpression().ArgumentExpression
}

func isTypecheckingBinaryExpression(ctx rule.RuleContext, binary *ast.BinaryExpression, callExpression *ast.CallExpression) bool {
	operator := binary.OperatorToken.Kind
	switch operator {
	case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
		return isTypecheckingExpression(ctx, binary.Left, callExpression) &&
			isTypecheckingExpression(ctx, binary.Right, callExpression)

	case ast.KindInstanceOfKeyword:
		return !isErrorConstructor(binary.Right)
	}

	if isExistenceCheck(binary) {
		return false
	}
	return isTypecheckingExpression(ctx, binary.Left, callExpression) ||
		isTypecheckingExpression(ctx, binary.Right, callExpression)
}

func isErrorConstructor(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if ast.IsIdentifier(node) {
		return errorNameRegexp.MatchString(node.Text())
	}
	if node.Kind != ast.KindPropertyAccessExpression || ast.IsOptionalChain(node) {
		return false
	}
	property := node.AsPropertyAccessExpression()
	return ast.IsIdentifier(property.Name()) &&
		errorNameRegexp.MatchString(property.Name().Text())
}

func isExistenceCheck(binary *ast.BinaryExpression) bool {
	switch binary.OperatorToken.Kind {
	case ast.KindEqualsEqualsToken,
		ast.KindExclamationEqualsToken,
		ast.KindEqualsEqualsEqualsToken,
		ast.KindExclamationEqualsEqualsToken:
	default:
		return false
	}

	return (isTypeofExpression(binary.Left) && isUndefinedString(binary.Right)) ||
		(isTypeofExpression(binary.Right) && isUndefinedString(binary.Left))
}

func isTypeofExpression(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	return node != nil && node.Kind == ast.KindTypeOfExpression
}

func isUndefinedString(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	return node != nil && node.Kind == ast.KindStringLiteral && node.Text() == "undefined"
}
