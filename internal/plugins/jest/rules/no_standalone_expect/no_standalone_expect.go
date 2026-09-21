package no_standalone_expect

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/no_standalone_expect"
)

//go:embed no_standalone_expect.schema.json
var schemaJSON []byte

var NoStandaloneExpectRule = shared.NewRule(shared.Config{
	Name:   "jest/no-standalone-expect",
	Schema: rule.NewSchema(schemaJSON),
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := utils.GetJestCallAnalysis(ctx)
		return shared.Runtime{
			IsTestCall: func(node *ast.Node) bool {
				return analysis.ParseTestCall(node) != nil
			},
			IsDescribeCall: func(node *ast.Node) bool {
				parsed := analysis.ParseFnCall(node)
				return parsed != nil && parsed.Kind == utils.JestFnTypeDescribe
			},
			ClassifyExpectCall: func(node *ast.Node) shared.ExpectCallKind {
				parsed := analysis.ParseExpectCall(node)
				if parsed == nil {
					return shared.NotExpectCall
				}
				if utils.IsStaticExpectMatcher(parsed.Matcher, parsed.Head.Local.Node) {
					return shared.StaticExpectCall
				}
				return shared.AssertingExpectCall
			},
		}
	},
})
