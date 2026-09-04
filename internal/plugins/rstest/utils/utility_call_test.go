package utils_test

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// utilityCallProbe reports every call ParseRstestPluginManagedCall matches,
// naming the receiver and member it read, so a test case's code doubles as the
// expected parse.
var utilityCallProbe = rule.Rule{
	Name: "rstest/utility-call-probe",
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				parsed := rstestUtils.ParseRstestPluginManagedCall(node)
				if parsed == nil {
					return
				}
				ctx.ReportNode(parsed.MemberNode, rule.RuleMessage{
					Id:          "matched",
					Description: parsed.Namespace + "." + parsed.Member,
				})
			},
		}
	},
}

func matched(description string, column, endColumn int) []rule_tester.InvalidTestCaseError {
	return []rule_tester.InvalidTestCaseError{{
		MessageId: "matched",
		Message:   description,
		Line:      1,
		Column:    column,
		EndLine:   1,
		EndColumn: endColumn,
	}}
}

// ParseRstestPluginManagedCall matches the receiver on the name written at the
// call site, because Rstest's build does: where the binding came from never
// enters into it, and the wrappers TypeScript erases before the build runs are
// transparent. Which shapes it accepts is per member, from the shared table.
func TestParseRstestPluginManagedCall(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &utilityCallProbe,
		[]rule_tester.ValidTestCase{
			// A different name never reaches the build's rewrite, however it
			// was bound.
			{Code: `import { rs as vi } from '@rstest/core'; vi.mock('./m');`},
			{Code: `import * as core from '@rstest/core'; core.rs.mock('./m');`},
			{Code: `mocker.mock('./m');`},
			// A member the build does not manage is not this parser's
			// business: `fn`, `spyOn`, `mocked` and the rest are ordinary
			// functions, matched by resolving the receiver instead.
			{Code: `rs.fn();`},
			{Code: `rs.spyOn(target, 'method');`},
			{Code: `rs.somethingElse();`},
			// The mock family is matched on the name as written and nothing
			// else: a bracketed key, an optional call and an optional receiver
			// all stay the throwing stub.
			{Code: `rs['mock']('./m');`},
			{Code: `rs.mock?.('./m');`},
			{Code: `rs?.mock('./m');`},
			{Code: `rs?.foo.mock('./m');`},
			// The two members that do read a bracketed key read a plain quoted
			// string only, and never an optional receiver.
			{Code: "rs[`importActual`]('./m');"},
			{Code: `rs[member]('./m');`},
			{Code: `rs?.importActual('./m');`},
			// Not a member call at all.
			{Code: `rs('./m');`},
			{Code: `mock('./m');`},
		},
		[]rule_tester.InvalidTestCase{
			{Code: `rs.mock('./m');`, Errors: matched("rs.mock", 4, 8)},
			{Code: `rstest.hoisted(setup);`, Errors: matched("rstest.hoisted", 8, 15)},
			// `importActual` and `requireActual` are the two members whose
			// rewrite reads a bracketed string key and an optional call.
			{Code: `rs['importActual']('./m');`, Errors: matched("rs.importActual", 4, 18)},
			{Code: `rs.requireActual?.('./m');`, Errors: matched("rs.requireActual", 4, 17)},
			{Code: `(rs as any)['importActual']('./m');`, Errors: matched("rs.importActual", 13, 27)},
			// A binding that has nothing to do with `@rstest/core` is still
			// matched, because the build matches it too.
			{
				Code:   `const rs = { mock() {} }; rs.mock('./m');`,
				Errors: matched("rs.mock", 30, 34),
			},
			{
				Code:   `import { rs } from './helpers'; rs.mock('./m');`,
				Errors: matched("rs.mock", 36, 40),
			},
			// Parentheses and TypeScript's type-only syntax are erased before
			// the build's rewrite runs, on either side of the callee.
			{Code: `(rs).mock('./m');`, Errors: matched("rs.mock", 6, 10)},
			{Code: `((rs)).mock('./m');`, Errors: matched("rs.mock", 8, 12)},
			{Code: `rs!.mock('./m');`, Errors: matched("rs.mock", 5, 9)},
			{Code: `(rs as any).mock('./m');`, Errors: matched("rs.mock", 13, 17)},
			{Code: `(rs satisfies object).mock('./m');`, Errors: matched("rs.mock", 23, 27)},
			{Code: `(rs.mock)('./m');`, Errors: matched("rs.mock", 5, 9)},
			{Code: `(rs.mock as any)('./m');`, Errors: matched("rs.mock", 5, 9)},
		},
	)
}
