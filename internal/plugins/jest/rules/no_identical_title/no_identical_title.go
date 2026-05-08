package no_identical_title

import (
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Message Builder

func buildMultipleTestTitleMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "multipleTestTitle",
		Description: "Test title is used multiple times in the same describe block",
	}
}

func buildMultipleDescribeTitleMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "multipleDescribeTitle",
		Description: "Describe block title is used multiple times in the same describe block",
	}
}

type titleLayer struct {
	describeTitles map[string]struct{}
	testTitles     map[string]struct{}
}

func newTitleLayer() *titleLayer {
	return &titleLayer{
		describeTitles: make(map[string]struct{}),
		testTitles:     make(map[string]struct{}),
	}
}

func staticJestTitleValue(arg *ast.Node) (string, bool) {
	if arg == nil {
		return "", false
	}
	switch arg.Kind {
	case ast.KindStringLiteral:
		return arg.AsStringLiteral().Text, true
	case ast.KindNoSubstitutionTemplateLiteral:
		return arg.AsNoSubstitutionTemplateLiteral().Text, true
	case ast.KindTemplateExpression:
		te := arg.AsTemplateExpression()
		if te == nil || te.TemplateSpans == nil || len(te.TemplateSpans.Nodes) != 0 || te.Head == nil {
			return "", false
		}
		return te.Head.Text(), true
	default:
		return "", false
	}
}

var NoIdenticalTitleRule = rule.Rule{
	Name: "jest/no-identical-title",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		contexts := []*titleLayer{newTitleLayer()}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				jestFn := utils.ParseJestFnCall(node, ctx)
				if jestFn == nil {
					return
				}

				cur := contexts[len(contexts)-1]
				if jestFn.Kind == utils.JestFnTypeDescribe {
					contexts = append(contexts, newTitleLayer())
				}

				if slices.Contains(jestFn.Members, "each") {
					return
				}

				callExpr := node.AsCallExpression()
				if callExpr == nil || callExpr.Arguments == nil || len(callExpr.Arguments.Nodes) < 1 {
					return
				}
				arg0 := callExpr.Arguments.Nodes[0]
				title, ok := staticJestTitleValue(arg0)
				if !ok {
					return
				}

				if jestFn.Kind == utils.JestFnTypeTest {
					if _, ok := cur.testTitles[title]; ok {
						ctx.ReportNode(arg0, buildMultipleTestTitleMessage())
					}
					cur.testTitles[title] = struct{}{}
				}

				if jestFn.Kind != utils.JestFnTypeDescribe {
					return
				}
				if _, ok := cur.describeTitles[title]; ok {
					ctx.ReportNode(arg0, buildMultipleDescribeTitleMessage())
				}

				cur.describeTitles[title] = struct{}{}
			},
			rule.ListenerOnExit(ast.KindCallExpression): func(node *ast.Node) {
				if !utils.IsTypeOfJestFnCall(node, ctx, utils.JestFnTypeDescribe) {
					return
				}
				if len(contexts) > 1 {
					contexts = contexts[:len(contexts)-1]
				}
			},
		}
	},
}
