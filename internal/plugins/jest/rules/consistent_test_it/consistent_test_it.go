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
				// Replace only the root identifier. A chain such as
				// `it.only.each` keeps its modifiers; every Jest chain with more
				// than one member starts with `it` or `test`, and each one has a
				// counterpart under the other name.
				root := parsed.Head.Local.Node
				if root == nil || root.Kind != ast.KindIdentifier {
					return nil
				}
				text := preferredName(parsed.Name, preferred)
				name, _, _ := strings.Cut(text, ".")
				return shared.RespellFixes(ctx, shared.Respell{
					Root: root, Symbol: ctx.Refs.Resolve(root), Global: parsed.Head.Type == jestUtils.JEST_GLOBAL_MODE,
					Text: text, Name: name, IsModule: isJestGlobalsModule,
					Respelled: respelled,
				})
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

func isJestGlobalsModule(module string) bool {
	return module == jestUtils.JestGlobalsModule
}

// respelled returns the name calls through specifier are fixed to: `it` and
// `fit` become `test`, `xit` becomes `xtest`, and the reverse.
func respelled(specifier *ast.Node) string {
	imported := testFramework.ImportedSpecifierName(specifier)
	var text string
	switch imported {
	case "it", "fit", "xit":
		text = preferredName(imported, "test")
	case "test", "xtest":
		text = preferredName(imported, "it")
	default:
		return ""
	}
	name, _, _ := strings.Cut(text, ".")
	return name
}
