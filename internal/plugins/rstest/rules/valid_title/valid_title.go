package valid_title

import (
	"regexp"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/valid_title"
)

// The accepted specifiers are Rstest's, not Jest's. formatRegExp is
// /%[sdjifoOc%]/ (packages/core/src/runtime/util.ts:159 at rstest c4b67c72) and
// %# / %$ are substituted ahead of it by formatName, so `%p` — valid to Jest —
// is left verbatim by Rstest, while `%O` and `%c` — reported by Jest — are
// expanded by formatTemplate.
// cspell:ignore sdjifo
var invalidEachSpecifier = regexp.MustCompile(`%[^sdjifoOc#$%]`)

// ValidTitleRule enforces rstest/valid-title.
var ValidTitleRule = shared.NewRule(shared.Config{
	Name:                 "rstest/valid-title",
	FunctionNameDataKey:  "rstestFunctionName",
	InvalidEachSpecifier: invalidEachSpecifier,
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		// import.meta.rstest.test(...) and @rstest/playwright are official
		// registration shapes, so title validation applies to them too.
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return shared.Runtime{
			Parse: func(node *ast.Node) *testFramework.ParsedCall {
				parsed := analysis.ParseFnCall(node)
				if parsed == nil {
					return nil
				}
				return &parsed.ParsedCall
			},
			// `.for` formats its title like `.each` (runtime.ts:466-576), and the
			// parser's conclusion survives alias resolution where the call
			// site's member names do not.
			IsParameterized: func(node *ast.Node, _ *testFramework.ParsedCall) bool {
				parsed := analysis.ParseFnCall(node)
				return parsed != nil && parsed.IsParameterized()
			},
		}
	},
})
