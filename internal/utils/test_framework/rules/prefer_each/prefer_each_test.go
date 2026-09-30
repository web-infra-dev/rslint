package prefer_each_test

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/prefer_each"
)

// The synthetic parser classifies calls by callee name alone so the frame
// logic is exercised without any framework's provenance rules: `it` and `test`
// are tests, `describe` a suite, `hook` a hook, everything else is not a
// registration.
func sharedRule() rule.Rule {
	return shared.NewRule(shared.Config{
		Name: "test/prefer-each",
		SingleTestFn: func(name string) string {
			if name == "it" {
				return "it"
			}
			return "test"
		},
		Prepare: func(rule.RuleContext) shared.Runtime {
			return shared.Runtime{Parse: func(node *ast.Node) *testFramework.ParsedCall {
				callee := node.AsCallExpression().Expression
				if callee.Kind != ast.KindIdentifier {
					return nil
				}
				name := callee.AsIdentifier().Text
				switch name {
				case "it", "test":
					return &testFramework.ParsedCall{Name: name, Kind: testFramework.FnKindTest}
				case "describe":
					return &testFramework.ParsedCall{Name: name, Kind: testFramework.FnKindDescribe}
				case "hook":
					return &testFramework.ParsedCall{Name: name, Kind: testFramework.FnKindHook}
				}
				return nil
			}}
		},
	})
}

func preferEach(fn string, line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "preferEach",
		Message:   "prefer using `" + fn + ".each` rather than a manual loop",
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}

func TestSharedPreferEachLoopFrames(t *testing.T) {
	r := sharedRule()
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&r,
		[]rule_tester.ValidTestCase{
			{Code: `for (const row of rows) { consume(row); }`},
			{Code: `while (rows.length) { it('a', () => {}); }`},
			{Code: `it('a', () => { for (const row of rows) { consume(row); } });`},
			{Code: `for (const row of getRows(it('once', () => {}))) {}`},
			{Code: `for (let i = 0; check(it('once', () => {})); i++) {}`},
			{Code: `for (const row of rows) { other(row); }`},
		},
		[]rule_tester.InvalidTestCase{
			// The single-test message follows the parsed API name.
			{Code: `for (const row of rows) { it(row, () => {}); }`, Errors: []rule_tester.InvalidTestCaseError{preferEach("it", 1, 1, 1, 47)}},
			{Code: `for (const row of rows) { test(row, () => {}); }`, Errors: []rule_tester.InvalidTestCaseError{preferEach("test", 1, 1, 1, 49)}},
			{Code: `for (const row of rows) { it(row, () => {}); test(row, () => {}); }`, Errors: []rule_tester.InvalidTestCaseError{preferEach("describe", 1, 1, 1, 68)}},
			{Code: `for (const row of rows) { hook(); }`, Errors: []rule_tester.InvalidTestCaseError{preferEach("describe", 1, 1, 1, 36)}},
			{Code: `for (const row in rows) { it(row, () => {}); }`, Errors: []rule_tester.InvalidTestCaseError{preferEach("it", 1, 1, 1, 47)}},
			{Code: `for (let i = 0; i < n; i++) { it(i, () => {}); }`, Errors: []rule_tester.InvalidTestCaseError{preferEach("it", 1, 1, 1, 49)}},
			// The outer loop registers on its own, so a nested business loop does
			// not cancel its report, in either statement order.
			{Code: `for (const a of as) { it(a, () => {}); for (const b of a.items) { setup(b); } }`, Errors: []rule_tester.InvalidTestCaseError{preferEach("it", 1, 1, 1, 80)}},
			{Code: `for (const a of as) { for (const b of a.items) { setup(b); } it(a, () => {}); }`, Errors: []rule_tester.InvalidTestCaseError{preferEach("it", 1, 1, 1, 80)}},
			// Nested registering loops report separately, inner first; each frame holds
			// only its own registrations, so the outer loop's single `it` stays `it`.
			{
				Code: `for (const a of as) { it(a, () => {}); for (const b of a.items) { it(b, () => {}); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					preferEach("it", 1, 40, 1, 85),
					preferEach("it", 1, 1, 1, 87),
				},
			},
			// A loop inside a test callback is judged by what it registers.
			{Code: `it('a', () => { for (const row of rows) { test(row, () => {}); } });`, Errors: []rule_tester.InvalidTestCaseError{preferEach("test", 1, 17, 1, 65)}},
		},
	)
}
