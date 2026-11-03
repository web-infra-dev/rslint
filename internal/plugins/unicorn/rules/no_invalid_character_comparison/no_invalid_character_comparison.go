// Ported from eslint-plugin-unicorn v79.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/no-invalid-character-comparison.js
package no_invalid_character_comparison

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/jsnum"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

var NoInvalidCharacterComparisonRule = rule.Rule{
	Name:   "unicorn/no-invalid-character-comparison",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		var evaluator *utils.StaticStringEvaluator
		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				operator := binary.OperatorToken.Kind
				if !ast.IsEqualityOperator(operator) {
					return
				}
				left, right := utils.ESTreeRuntimeExpression(binary.Left), utils.ESTreeRuntimeExpression(binary.Right)
				for _, pair := range [][2]*ast.Node{{left, right}, {right, left}} {
					value, ok := utils.GetStaticStringLiteralValue(pair[1])
					if !ok || ecmascript.StringCodeUnitCount(value) <= 1 {
						continue
					}
					receiver := singleCharacterReceiver(pair[0])
					if receiver == nil {
						continue
					}
					if evaluator == nil {
						evaluator = utils.NewStaticStringEvaluatorWithReferenceResolver(ctx.TypeChecker, ctx.SourceFile, ctx.Refs)
					}
					if !unicornutil.IsString(ctx, receiver, evaluator) {
						continue
					}
					result := "false"
					if operator == ast.KindExclamationEqualsToken || operator == ast.KindExclamationEqualsEqualsToken {
						result = "true"
					}
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "no-invalid-character-comparison",
						Description: "A single character can never equal the multi-character string `" + value + "`, so this comparison is always " + result + ".",
						Data:        map[string]string{"string": value, "result": result},
					})
					return
				}
			},
		}
	},
}

func singleCharacterReceiver(node *ast.Node) *ast.Node {
	// ESTree represents the entire optional chain as a ChainExpression.
	if node == nil || ast.IsOptionalChain(node) {
		return nil
	}
	maximumArguments := 1
	if call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
		Methods: []string{"charAt", "at"}, MaximumArguments: &maximumArguments, RejectSpreadElement: true,
	}); ok {
		return call.Object
	}
	if ast.IsElementAccessExpression(node) {
		access := node.AsElementAccessExpression()
		if isStaticStringIndex(utils.ESTreeRuntimeExpression(access.ArgumentExpression)) {
			return access.Expression
		}
	}
	return nil
}

func isStaticStringIndex(node *ast.Node) bool {
	if node != nil && node.Kind == ast.KindNumericLiteral {
		return true
	}
	value, ok := utils.GetStaticStringLiteralValue(node)
	if !ok {
		return false
	}
	number, ok := ecmascript.StringToNumber(value)
	index := jsnum.Number(number)
	return ok && index >= 0 && index <= jsnum.MaxSafeInteger && index.Floor() == index && index.String() == value
}
