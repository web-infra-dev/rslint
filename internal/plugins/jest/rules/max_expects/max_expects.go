package max_expects

import (
	_ "embed"
	"fmt"
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed max_expects.schema.json
var schemaJSON []byte

const defaultMax = 5

type options struct {
	Max int
}

func parseOptions(rawOptions []any) options {
	opts := options{Max: defaultMax}
	if len(rawOptions) == 0 {
		return opts
	}

	optsMap, _ := rawOptions[0].(map[string]interface{})
	if maxVal, ok := internalUtils.CoerceIntegral(optsMap["max"]); ok && maxVal >= 1 {
		opts.Max = maxVal
	}

	return opts
}

func buildExceededMaxAssertionMessage(count, maxAllowed int) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "exceededMaxAssertion",
		Description: fmt.Sprintf("Too many assertion calls (%d) - maximum allowed is %d", count, maxAllowed),
		Data: map[string]string{
			"count": strconv.Itoa(count),
			"max":   strconv.Itoa(maxAllowed),
		},
	}
}

func isTestCallbackFunction(fn *ast.Node, analysis *utils.JestCallAnalysis) bool {
	parent := fn.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	if parent == nil || parent.Kind != ast.KindCallExpression {
		return true
	}
	return analysis.ParseTestCall(parent) != nil
}

func shouldCountExpectCall(jestFnCall *utils.ParsedJestFnCall) bool {
	if jestFnCall == nil || jestFnCall.Kind != utils.JestFnTypeExpect {
		return false
	}

	headNode := jestFnCall.Head.Local.Node
	if headNode != nil && utils.IsMemberAccessNode(headNode.Parent) {
		return false
	}

	return true
}

func maybeResetCountForFunctionLike(node *ast.Node, analysis *utils.JestCallAnalysis, count *int) {
	if node.Body() == nil {
		return
	}
	if isTestCallbackFunction(node, analysis) {
		*count = 0
	}
}

var MaxExpectsRule = rule.Rule{
	Name:   "jest/max-expects",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		analysis := utils.GetJestCallAnalysis(ctx)
		opts := parseOptions(options)
		count := 0

		maybeResetCount := func(node *ast.Node) {
			maybeResetCountForFunctionLike(node, analysis, &count)
		}

		return rule.RuleListeners{
			ast.KindFunctionExpression:                      maybeResetCount,
			rule.ListenerOnExit(ast.KindFunctionExpression): maybeResetCount,
			ast.KindArrowFunction:                           maybeResetCount,
			rule.ListenerOnExit(ast.KindArrowFunction):      maybeResetCount,
			ast.KindMethodDeclaration:                       maybeResetCount,
			rule.ListenerOnExit(ast.KindMethodDeclaration):  maybeResetCount,
			ast.KindGetAccessor:                             maybeResetCount,
			rule.ListenerOnExit(ast.KindGetAccessor):        maybeResetCount,
			ast.KindSetAccessor:                             maybeResetCount,
			rule.ListenerOnExit(ast.KindSetAccessor):        maybeResetCount,
			ast.KindConstructor:                             maybeResetCount,
			rule.ListenerOnExit(ast.KindConstructor):        maybeResetCount,
			ast.KindCallExpression: func(node *ast.Node) {
				jestFnCall := analysis.ParseFnCall(node)
				if jestFnCall == nil {
					return
				}

				if jestFnCall.Kind == utils.JestFnTypeTest {
					count = 0
					return
				}

				if !shouldCountExpectCall(jestFnCall) {
					return
				}

				count++
				if count > opts.Max {
					ctx.ReportNode(node, buildExceededMaxAssertionMessage(count, opts.Max))
				}
			},
		}
	},
}
