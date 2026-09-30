package valid_title

import (
	"regexp"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/valid_title"
)

// Jest formats `.each` titles with util.format: %p (pretty-format) is valid,
// while %O and %c are not expanded.
var invalidEachSpecifier = regexp.MustCompile(`%[^psdifjo#$%]`)

// ValidTitleRule enforces ESLint jest/valid-title.
var ValidTitleRule = shared.NewRule(shared.Config{
	Name:                 "jest/valid-title",
	FunctionNameDataKey:  "jestFunctionName",
	InvalidEachSpecifier: invalidEachSpecifier,
	LegacyAliases: map[string]string{
		"fdescribe": "describe",
		"xdescribe": "describe",
		"fit":       "it",
		"xit":       "it",
		"xtest":     "test",
	},
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := jestUtils.GetJestCallAnalysis(ctx)
		return shared.Runtime{
			Parse: func(node *ast.Node) *testFramework.ParsedCall {
				parsed := analysis.ParseFnCall(node)
				if parsed == nil {
					return nil
				}
				return &parsed.ParsedCall
			},
			IsParameterized: func(_ *ast.Node, parsed *testFramework.ParsedCall) bool {
				return slices.Contains(parsed.Members, "each")
			},
		}
	},
})
