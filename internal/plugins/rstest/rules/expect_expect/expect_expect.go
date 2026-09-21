package expect_expect

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/expect_expect"
)

// sourceMayContainRstestAssertion reports whether the file mentions `expect` or
// `assert` at all. Every form the analysis can resolve — `ctx.expect`,
// `rstest.expect`, `import { expect as check }`, `import.meta.rstest.expect`,
// and the same shapes for `assert` — spells one of those identifiers
// somewhere, so a file without either can skip the resolving hook and avoid
// forcing the analysis to collect test callbacks.
//
// Both names have to be checked. Gating on `expect` alone switched the hook off
// for a file that asserts only through `assert`, which left every resolved
// `assert` binding unrecognized however well the analysis could resolve it.
func sourceMayContainRstestAssertion(sourceFile *ast.SourceFile) bool {
	if sourceFile == nil || sourceFile.AsNode().Kind != ast.KindSourceFile {
		return true
	}
	return sourceFile.HasIdentifier("expect") || sourceFile.HasIdentifier("assert")
}

var ExpectExpectRule = shared.NewRule(shared.Config{
	Name: "rstest/expect-expect",
	// Rstest ships both `expect` and `assert` (chai) as globals
	// (packages/core/src/utils/constants.ts globalApiList), so the default
	// asserting functions match eslint-plugin-vitest rather than jest.
	DefaultAssertFunctionNames: []string{"expect", "assert"},
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		// Rstest reaches `expect` through the test context, namespace imports and
		// import aliases, and reaches Chai's `assert` through all but the test
		// context, none of which the callee-text patterns can match. Reuse the
		// same resolution rstest/no-conditional-expect consumes so both rules
		// agree on what an assertion is.
		var isAssertion func(node *ast.Node) bool
		if sourceMayContainRstestAssertion(ctx.SourceFile) {
			isAssertion = func(node *ast.Node) bool {
				return analysis.IsExpectCall(node) || analysis.IsAssertCall(node)
			}
		}
		return shared.Runtime{
			IsAssertion: isAssertion,
			ClassifyTest: func(node *ast.Node) shared.TestClassification {
				parsed := analysis.ParseTestCall(node)
				if parsed == nil {
					return shared.TestClassification{}
				}
				// Todo is a semantic field that survives const aliases; the
				// call-site Members can be empty for `const t = test.todo;
				// t('x')`, so never derive the exemption from Members here.
				return shared.TestClassification{IsTest: true, IsTodo: parsed.Todo}
			},
			ResolveNamedCallback: func(callNode *ast.Node) shared.NamedCallback {
				function, name := analysis.TestCallback(callNode)
				return shared.NamedCallback{
					DeclarationNode: function,
					Name:            name,
				}
			},
		}
	},
})
