package consistent_test_it

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/consistent_test_it"
)

var ConsistentTestItRule = shared.NewRule(shared.Config{
	Name: "jest/consistent-test-it",
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
			// `xit` and `fit` spell `it`, and `xtest` spells `test`.
			Check: func(node *ast.Node, parsed *testFramework.ParsedCall, preferred string) (*ast.Node, string, bool) {
				if strings.HasSuffix(parsed.Name, preferred) {
					return nil, "", false
				}
				return utils.ESTreeRuntimeExpression(node.AsCallExpression().Expression), oppositeKeyword(preferred), true
			},
			Fix: func(node *ast.Node, parsed *testFramework.ParsedCall, preferred string) []rule.RuleFix {
				// Replace only the root identifier, whatever it is bound as. A
				// chain such as `it.only.each` keeps its modifiers; every Jest
				// chain with more than one member starts with `it` or `test`,
				// and each one has a counterpart under the other name.
				root := parsed.Head.Local.Node
				if root == nil || root.Kind != ast.KindIdentifier {
					return nil
				}
				return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, root, preferredName(parsed.Name, preferred))}
			},
		}
	},
})

func preferredName(name, preferred string) string {
	if name == "fit" {
		return "test.only"
	}
	if strings.HasPrefix(name, "f") || strings.HasPrefix(name, "x") {
		return name[:1] + preferred
	}
	return preferred
}

func oppositeKeyword(keyword string) string {
	if keyword == "test" {
		return "it"
	}
	return "test"
}
