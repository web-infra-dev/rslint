package no_new_native_nonconstructor

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var nativeNonconstructorNames = map[string]struct{}{
	"Symbol": {},
	"BigInt": {},
}

// https://eslint.org/docs/latest/rules/no-new-native-nonconstructor
var NoNewNativeNonconstructorRule = rule.Rule{
	Name: "no-new-native-nonconstructor",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindNewExpression: func(node *ast.Node) {
				newExpr := node.AsNewExpression()
				if newExpr == nil || newExpr.Expression == nil {
					return
				}

				callee := utils.SkipAssertionsAndParens(newExpr.Expression)
				if callee == nil || callee.Kind != ast.KindIdentifier {
					return
				}

				name := callee.AsIdentifier().Text
				if _, ok := nativeNonconstructorNames[name]; !ok || utils.IsShadowed(callee, name) {
					return
				}

				ctx.ReportNode(callee, rule.RuleMessage{
					Id:          "noNewNonconstructor",
					Description: fmt.Sprintf("`%s` cannot be called as a constructor.", name),
				})
			},
		}
	},
}
