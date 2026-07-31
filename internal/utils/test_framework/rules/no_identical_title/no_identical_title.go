package no_identical_title

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

type ParsedCall struct {
	Call          *testFramework.ParsedCall
	Parameterized bool
}

type Config struct {
	Name  string
	Parse func(*ast.Node, rule.RuleContext) *ParsedCall
}

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

// NewRule creates a no-identical-title rule for a test framework.
func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name:   config.Name,
		Schema: rule.EmptyArraySchema,
		Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
			contexts := []*titleLayer{newTitleLayer()}

			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					parsed := config.Parse(node, ctx)
					if parsed == nil || parsed.Call == nil {
						return
					}

					current := contexts[len(contexts)-1]
					if parsed.Call.Kind == testFramework.FnKindDescribe {
						contexts = append(contexts, newTitleLayer())
					}

					if parsed.Parameterized {
						return
					}

					call := node.AsCallExpression()
					if call == nil || call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
						return
					}
					arg0 := call.Arguments.Nodes[0]
					title, ok := internalUtils.GetStaticStringLiteralValue(arg0)
					if !ok {
						return
					}

					if parsed.Call.Kind == testFramework.FnKindTest {
						if _, exists := current.testTitles[title]; exists {
							ctx.ReportNode(arg0, buildMultipleTestTitleMessage())
						}
						current.testTitles[title] = struct{}{}
						return
					}

					if parsed.Call.Kind != testFramework.FnKindDescribe {
						return
					}
					if _, exists := current.describeTitles[title]; exists {
						ctx.ReportNode(arg0, buildMultipleDescribeTitleMessage())
					}
					current.describeTitles[title] = struct{}{}
				},
				rule.ListenerOnExit(ast.KindCallExpression): func(node *ast.Node) {
					parsed := config.Parse(node, ctx)
					if parsed == nil || parsed.Call == nil ||
						parsed.Call.Kind != testFramework.FnKindDescribe {
						return
					}
					if len(contexts) > 1 {
						contexts = contexts[:len(contexts)-1]
					}
				},
			}
		},
	}
}
