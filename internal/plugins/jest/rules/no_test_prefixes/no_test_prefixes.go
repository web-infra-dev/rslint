package no_test_prefixes

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Message Builder

func buildErrorUsePreferredNameMessage(preferredName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "usePreferredName",
		Description: fmt.Sprintf("Use \"%s\" instead", preferredName),
	}
}

var NoTestPrefixesRule = rule.Rule{
	Name: "jest/no-test-prefixes",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				jestFnCall := utils.ParseJestFnCall(node, ctx)
				if jestFnCall == nil {
					return
				}

				if jestFnCall.Kind != utils.JestFnTypeDescribe && jestFnCall.Kind != utils.JestFnTypeTest {
					return
				}

				if !strings.HasPrefix(jestFnCall.Name, "f") && !strings.HasPrefix(jestFnCall.Name, "x") {
					return
				}

				callExpr := node.AsCallExpression()
				if callExpr == nil {
					return
				}

				modifier := ""
				if strings.HasPrefix(jestFnCall.Name, "f") {
					modifier = "only"
				} else if strings.HasPrefix(jestFnCall.Name, "x") {
					modifier = "skip"
				}

				parts := []string{jestFnCall.Name[1:], modifier}
				parts = append(parts, jestFnCall.Members...)
				preferredName := strings.Join(parts, ".")
				reportFix(callExpr.Expression, preferredName, ctx)
			},
		}
	},
}

func reportFix(node *ast.Node, preferredName string, ctx rule.RuleContext) {
	switch node.Kind {
	case ast.KindIdentifier, ast.KindPropertyAccessExpression,
		ast.KindElementAccessExpression:
		ctx.ReportNodeWithFixes(node, buildErrorUsePreferredNameMessage(preferredName),
			rule.RuleFixReplace(ctx.SourceFile, node, preferredName),
		)
	case ast.KindCallExpression:
		reportFix(node.AsCallExpression().Expression, preferredName, ctx)
	case ast.KindTaggedTemplateExpression:
		reportFix(node.AsTaggedTemplateExpression().Tag, preferredName, ctx)
	}
}
