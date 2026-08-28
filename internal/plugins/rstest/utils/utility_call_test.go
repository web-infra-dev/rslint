package utils_test

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// utilityCallProbe reports every call ParseRstestUtilityCall matches, naming
// the receiver and member it read, so a test case's code doubles as the
// expected parse.
var utilityCallProbe = rule.Rule{
	Name: "rstest/utility-call-probe",
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				parsed := rstestUtils.ParseRstestUtilityCall(node)
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

// ParseRstestUtilityCall matches the receiver on the name written at the call
// site, because Rstest's build does: where the binding came from never enters
// into it, and the wrappers TypeScript erases before the build runs are
// transparent.
func TestParseRstestUtilityCall(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &utilityCallProbe,
		[]rule_tester.ValidTestCase{
			// A different name never reaches the build's rewrite, however it
			// was bound.
			{Code: `import { rs as vi } from '@rstest/core'; vi.mock('./m');`},
			{Code: `import * as core from '@rstest/core'; core.rs.mock('./m');`},
			{Code: `mocker.mock('./m');`},
			// Computed members and optional chains are the shapes no API
			// accepts, so they are not parsed for any of them.
			{Code: `rs['mock']('./m');`},
			{Code: `rs?.mock('./m');`},
			{Code: `rs.mock?.('./m');`},
			{Code: `rs?.foo.mock('./m');`},
			// Not a member call at all.
			{Code: `rs('./m');`},
			{Code: `mock('./m');`},
		},
		[]rule_tester.InvalidTestCase{
			{Code: `rs.mock('./m');`, Errors: matched("rs.mock", 4, 8)},
			{Code: `rstest.hoisted(setup);`, Errors: matched("rstest.hoisted", 8, 15)},
			// The member is read as written, whether or not it is one Rstest
			// manages: naming the API is the caller's job.
			{Code: `rs.fn();`, Errors: matched("rs.fn", 4, 6)},
			{Code: `rs.somethingElse();`, Errors: matched("rs.somethingElse", 4, 17)},
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

// The plugin-managed list is what tells a rule which matching strategy an API
// needs, so it has to stay exactly the set Rstest's build rewrites.
func TestIsPluginManagedAPI(t *testing.T) {
	managed := []string{
		"mock", "mockRequire", "doMock", "doMockRequire",
		"unmock", "doUnmock", "unmockRequire", "doUnmockRequire",
		"importMock", "requireMock", "importActual", "requireActual",
		"resetModules", "hoisted",
	}
	for _, member := range managed {
		if !rstestUtils.IsPluginManagedAPI(member) {
			t.Errorf("IsPluginManagedAPI(%q) = false, want true", member)
		}
	}

	// Ordinary functions on the same object, and names that are not on it.
	ordinary := []string{
		"fn", "spyOn", "mocked", "isMockFunction", "mockObject",
		"clearAllMocks", "resetAllMocks", "restoreAllMocks",
		"stubEnv", "stubGlobal", "useFakeTimers", "waitFor", "waitUntil",
		"setConfig", "", "Mock", "test", "expect",
	}
	for _, member := range ordinary {
		if rstestUtils.IsPluginManagedAPI(member) {
			t.Errorf("IsPluginManagedAPI(%q) = true, want false", member)
		}
	}
}
