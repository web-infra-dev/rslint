package require_array_sort_compare

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed require_array_sort_compare.schema.json
var schemaJSON []byte

func buildRequireCompareMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "requireCompare",
		Description: "Require 'compare' argument.",
	}
}

type RequireArraySortCompareOptions struct {
	IgnoreStringArrays bool
}

func parseOptions(options []any) RequireArraySortCompareOptions {
	opts := RequireArraySortCompareOptions{
		IgnoreStringArrays: true,
	}
	if len(options) == 0 {
		return opts
	}
	optsMap, _ := options[0].(map[string]any)
	if value, ok := optsMap["ignoreStringArrays"].(bool); ok {
		opts.IgnoreStringArrays = value
	}
	return opts
}

var RequireArraySortCompareRule = rule.CreateRule(rule.Rule{
	Name:             "require-array-sort-compare",
	Schema:           rule.NewSchema(schemaJSON),
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				expr := node.AsCallExpression()
				if len(expr.Arguments.Nodes) != 0 {
					return
				}
				callee := expr.Expression

				if !ast.IsAccessExpression(callee) {
					return
				}

				if propertyName, found := checker.Checker_getAccessedPropertyName(ctx.TypeChecker, callee); !found || (propertyName != "sort" && propertyName != "toSorted") {
					return
				}

				calleeObjType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, callee.Expression())

				isArray := checker.Checker_isArrayType(ctx.TypeChecker, calleeObjType)
				if !isArray && !utils.Every(utils.UnionTypeParts(calleeObjType), func(t *checker.Type) bool {
					return checker.Checker_isArrayType(ctx.TypeChecker, t)
				}) {
					return
				}

				if opts.IgnoreStringArrays && isArray {
					if utils.Every(checker.Checker_getTypeArguments(ctx.TypeChecker, calleeObjType), func(t *checker.Type) bool {
						// Numeric types cannot be strings; avoid printing their type names.
						if utils.IsTypeFlagSet(t, checker.TypeFlagsNumberLike) {
							return false
						}
						return utils.GetTypeName(ctx.TypeChecker, t) == "string"
					}) {
						return
					}
				}

				ctx.ReportNode(node, buildRequireCompareMessage())
			},
		}
	},
})
