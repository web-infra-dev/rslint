package no_test_return_statement

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/no_test_return_statement"
)

// Rstest awaits whatever a test callback returns and ignores a value that is
// not thenable, so a return statement never changes a test's outcome by
// itself. The message therefore states the convention without claiming the
// test is broken.
var NoTestReturnStatementRule = shared.NewRule(shared.Config{
	Name: "rstest/no-test-return-statement",
	Message: rule.RuleMessage{
		Id:          "noTestReturnStatement",
		Description: "Return statements are not allowed in tests",
	},
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return shared.Runtime{
			// The analysis resolves both `(name, fn, timeout)` and
			// `(name, options, fn)`, and follows a callback name to the local
			// function its binding always holds.
			TestCallback: func(node *ast.Node) *ast.Node {
				if node.Kind != ast.KindCallExpression {
					return nil
				}
				function, _ := analysis.TestCallback(node)
				return function
			},
		}
	},
})
