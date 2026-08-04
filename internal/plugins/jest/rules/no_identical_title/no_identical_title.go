package no_identical_title

import (
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/no_identical_title"
)

func parseJestCall(node *ast.Node, ctx rule.RuleContext) *shared.ParsedCall {
	parsed := jestUtils.ParseJestFnCall(node, ctx)
	if parsed == nil {
		return nil
	}
	return &shared.ParsedCall{
		Call:          &parsed.ParsedCall,
		Parameterized: slices.Contains(parsed.Members, "each"),
	}
}

var NoIdenticalTitleRule = shared.NewRule(shared.Config{
	Name:  "jest/no-identical-title",
	Parse: parseJestCall,
})
