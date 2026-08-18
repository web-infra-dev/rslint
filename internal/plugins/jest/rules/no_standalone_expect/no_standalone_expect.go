package no_standalone_expect

import (
	_ "embed"

	"github.com/microsoft/typescript-go/shim/ast"
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
		// The shared traversal asks about the same node more than once per
		// visit, so keep the last parse around; a miss only costs a re-parse.
		var lastNode *ast.Node
		var lastParsed *utils.ParsedJestFnCall
		parseCall := func(node *ast.Node) *utils.ParsedJestFnCall {
			if node == lastNode {
				return lastParsed
			}
			parsed := utils.ParseJestFnCall(node, ctx)
			lastNode = node
			lastParsed = parsed
			return parsed
		}

		return shared.Runtime{
			IsTestCall: func(node *ast.Node) bool {
				parsed := parseCall(node)
				return parsed != nil && parsed.Kind == utils.JestFnTypeTest
			},
			IsDescribeCall: func(node *ast.Node) bool {
				parsed := parseCall(node)
				return parsed != nil && parsed.Kind == utils.JestFnTypeDescribe
			},
			ClassifyExpectCall: func(node *ast.Node) shared.ExpectCallKind {
				parsed := parseCall(node)
				if parsed == nil || parsed.Kind != utils.JestFnTypeExpect {
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
