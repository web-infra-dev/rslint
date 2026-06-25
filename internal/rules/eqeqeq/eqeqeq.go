package eqeqeq

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// strictEquivalent maps a loose operator to its strict counterpart.
func strictEquivalent(kind ast.Kind) string {
	switch kind {
	case ast.KindEqualsEqualsToken:
		return "==="
	case ast.KindExclamationEqualsToken:
		return "!=="
	default:
		return ""
	}
}

// looseEquivalent maps a strict operator to its loose counterpart.
func looseEquivalent(kind ast.Kind) string {
	switch kind {
	case ast.KindEqualsEqualsEqualsToken:
		return "=="
	case ast.KindExclamationEqualsEqualsToken:
		return "!="
	default:
		return ""
	}
}

// isLooseEquality checks if the operator is == or !=.
func isLooseEquality(kind ast.Kind) bool {
	return kind == ast.KindEqualsEqualsToken || kind == ast.KindExclamationEqualsToken
}

// isStrictEquality checks if the operator is === or !==.
func isStrictEquality(kind ast.Kind) bool {
	return kind == ast.KindEqualsEqualsEqualsToken || kind == ast.KindExclamationEqualsEqualsToken
}

// isTypeOf checks if a node is a typeof expression, unwrapping parentheses.
func isTypeOf(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return ast.SkipParentheses(node).Kind == ast.KindTypeOfExpression
}

// literalKindGroup returns a group identifier for literal kinds that represent
// the same JS typeof type. Returns -1 for non-literal nodes.
// Unwraps parentheses before checking.
func literalKindGroup(node *ast.Node) int {
	if node == nil {
		return -1
	}
	switch ast.SkipParentheses(node).Kind {
	case ast.KindStringLiteral:
		return 0
	case ast.KindNumericLiteral:
		return 1
	case ast.KindTrueKeyword, ast.KindFalseKeyword:
		return 2
	case ast.KindNullKeyword:
		return 3
	case ast.KindRegularExpressionLiteral:
		return 4
	default:
		return -1
	}
}

// areLiteralsAndSameType checks if both operands are literals of the same JS type,
// unwrapping parentheses on both sides.
func areLiteralsAndSameType(left, right *ast.Node) bool {
	lg := literalKindGroup(left)
	rg := literalKindGroup(right)
	return lg >= 0 && lg == rg
}

// reportEqeqeq reports the eqeqeq violation with fix or suggestion.
// For typeof and same-type-literal comparisons, an autofix is provided.
// For all other cases, a suggestion is provided instead.
func reportEqeqeq(ctx rule.RuleContext, operatorToken *ast.Node, left, right *ast.Node, expectedOperator string) {
	actualOperator := scanner.TokenToString(operatorToken.Kind)
	msg := rule.RuleMessage{
		Id:          "unexpected",
		Description: "Expected '" + expectedOperator + "' and instead saw '" + actualOperator + "'.",
	}
	suggestionMsg := rule.RuleMessage{
		Id:          "replaceOperator",
		Description: "Use '" + expectedOperator + "' instead of '" + actualOperator + "'.",
	}
	opRange := core.NewTextRange(
		utils.TrimNodeTextRange(ctx.SourceFile, operatorToken).Pos(),
		operatorToken.End(),
	)
	fix := rule.RuleFixReplaceRange(opRange, expectedOperator)
	canAutofix := isTypeOf(left) || isTypeOf(right) || areLiteralsAndSameType(left, right)

	rule.ReportNodeWithFixesOrSuggestions(ctx, operatorToken, canAutofix, msg, suggestionMsg, fix)
}

// EqeqeqRule requires use of === and !==.
// https://eslint.org/docs/latest/rules/eqeqeq
var EqeqeqRule = rule.Rule{
	Name: "eqeqeq",
	Schema0: rule.Enum("always", "smart").Default("always"),
	Schema1: rule.Object(map[string]rule.Schema{
		"null": rule.Enum("always", "never", "ignore").Default("always"),
	}),
	RunWithOptions: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts, _ := options.([]any)
		mode, _ := opts[0].(string)
		optsMap, _ := opts[1].(map[string]any)
		nullOption, _ := optsMap["null"].(string)

		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				if binary == nil || binary.OperatorToken == nil {
					return
				}

				opKind := binary.OperatorToken.Kind
				left := binary.Left
				right := binary.Right

				if left == nil || right == nil {
					return
				}

				switch mode {
				case "always":
					handleAlwaysMode(ctx, binary, opKind, left, right, nullOption)
				case "smart":
					handleSmartMode(ctx, binary, opKind, left, right)
				}
			},
		}
	},
}

// handleAlwaysMode implements the "always" mode logic.
func handleAlwaysMode(ctx rule.RuleContext, binary *ast.BinaryExpression, opKind ast.Kind, left, right *ast.Node, nullOption string) {
	isNullCheck := utils.IsNullLiteral(left) || utils.IsNullLiteral(right)

	if isLooseEquality(opKind) {
		if nullOption == "always" || !isNullCheck {
			reportEqeqeq(ctx, binary.OperatorToken, left, right, strictEquivalent(opKind))
		}
	} else if isStrictEquality(opKind) && nullOption == "never" && isNullCheck {
		reportEqeqeq(ctx, binary.OperatorToken, left, right, looseEquivalent(opKind))
	}
}

// handleSmartMode implements the "smart" mode logic.
// In smart mode, == is allowed for:
// 1. Comparing typeof expressions
// 2. Comparing two literals of the same type
// 3. Comparing against null
func handleSmartMode(ctx rule.RuleContext, binary *ast.BinaryExpression, opKind ast.Kind, left, right *ast.Node) {
	if !isLooseEquality(opKind) {
		return
	}

	// Allow typeof comparisons
	if isTypeOf(left) || isTypeOf(right) {
		return
	}

	// Allow same-type literal comparisons
	if areLiteralsAndSameType(left, right) {
		return
	}

	// Allow null comparisons
	if utils.IsNullLiteral(left) || utils.IsNullLiteral(right) {
		return
	}

	// None of the smart exceptions apply; flag it
	reportEqeqeq(ctx, binary.OperatorToken, left, right, strictEquivalent(opKind))
}
