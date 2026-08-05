package no_disabled_tests

import (
	"github.com/microsoft/typescript-go/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/no_disabled_tests"
)

func parseRstestCall(node *ast.Node, ctx rule.RuleContext) *shared.ParsedCall {
	parsed := rstestUtils.ParseRstestFnCall(node, ctx)
	if parsed == nil {
		return nil
	}
	return &shared.ParsedCall{
		Call: &parsed.ParsedCall,
		// Semantic fields rather than Members, so `const skipped = test.skip`
		// followed by `skipped('case', fn)` is still recognized.
		HasSkip: parsed.Skipped,
		HasTodo: parsed.Todo,
	}
}

var NoDisabledTestsRule = shared.NewRule(shared.Config{
	Name:  "rstest/no-disabled-tests",
	Parse: parseRstestCall,
	// `TestCall` accepts `(description, options, fn?)`.
	HasOptionsOverload: true,
})
