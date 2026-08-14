package no_standalone_expect

import (
	_ "embed"

	"github.com/microsoft/typescript-go/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/no_standalone_expect"
)

//go:embed no_standalone_expect.schema.json
var schemaJSON []byte

func isStaticExpectCall(parsed *rstestUtils.ParsedRstestExpectCall) bool {
	if rstestUtils.IsStaticRstestExpectCall(parsed) {
		return true
	}

	// The shared helper deliberately treats every two-member static chain as
	// reportable. Rstest's typed `expect.not.<asymmetric>()` form is still a
	// value constructor rather than an assertion, so allow its known built-ins
	// without widening broken chains such as `expect.resolves.toBe()`.
	return parsed != nil &&
		parsed.Entry == rstestUtils.RstestExpectEntryStatic &&
		len(parsed.Modifiers) == 1 &&
		parsed.Modifiers[0] == "not" &&
		rstestUtils.RSTEST_ASYMMETRIC_MATCHERS[parsed.Matcher]
}

var NoStandaloneExpectRule = shared.NewRule(shared.Config{
	Name:   "rstest/no-standalone-expect",
	Schema: rule.NewSchema(schemaJSON),
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return shared.Runtime{
			IsTestCall: func(node *ast.Node) bool {
				parsed := analysis.ParseFnCall(node)
				return parsed != nil && parsed.Kind == rstestUtils.RstestFnTypeTest
			},
			IsDescribeCall: func(node *ast.Node) bool {
				parsed := analysis.ParseFnCall(node)
				return parsed != nil && parsed.Kind == rstestUtils.RstestFnTypeDescribe
			},
			ClassifyExpectCall: func(node *ast.Node) shared.ExpectCallKind {
				parsed := analysis.ParseExpectCall(node)
				if parsed == nil {
					return shared.NotExpectCall
				}
				if isStaticExpectCall(parsed) {
					return shared.StaticExpectCall
				}
				return shared.AssertingExpectCall
			},
		}
	},
})
