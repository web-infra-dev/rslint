package no_ex_assign

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoExAssignExtras locks in the resolution-sensitive behaviors of the
// Refs-API implementation that the upstream ESLint suite doesn't exercise.
// Each case carries an inline comment pointing at the specific branch /
// Dimension row / real-user shape it covers, so future refactors can't
// silently regress them without breaking a named lock-in.
func TestNoExAssignExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoExAssignRule,
		[]rule_tester.ValidTestCase{
			// Branch lock-in: optional catch binding — no VariableDeclaration, rule bails.
			{Code: "try { } catch { e = 10; }"},
			// Branch lock-in: a nested function parameter shadows the catch binding,
			// so the scope walk resolves the write to the parameter symbol.
			{Code: "try { } catch (e) { const f = (e: unknown) => { e = 10; }; }"},
			// Branch lock-in: labels are not variable references (RefStore skips them).
			{Code: "try { } catch (e) { e: for (;;) { break e; } }"},
			// Branch lock-in: a type-query reference resolves to the catch binding
			// but is not a write.
			{Code: "try { } catch (e) { let x: typeof e; }"},
			// Dimension: rest element in a destructured catch binding, read only.
			{Code: "try { } catch ({...rest}) { log(rest); }"},
		},
		[]rule_tester.InvalidTestCase{
			// Real-user shape: assignment from a nested function still targets the
			// catch binding via the scope walk (upstream reports this too).
			{
				Code: "try { } catch (e) { function f() { e = 10; } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 36},
				},
			},
			// Branch lock-in: writes to a multi-name destructured binding are
			// reported in document order, not per-symbol order.
			{
				Code: "try { } catch ({a, b}) { b = 1; a = 2; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 26},
					{MessageId: "unexpected", Line: 1, Column: 33},
				},
			},
			// Dimension: rest element symbol comes from its BindingElement.
			{
				Code: "try { } catch ({...rest}) { rest = {}; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 29},
				},
			},
			// Dimension: name nested two pattern levels deep.
			{
				Code: "try { } catch ({a: {b}}) { b = 1; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 28},
				},
			},
			// Branch lock-in: parenthesized assignment target.
			{
				Code: "try { } catch (e) { (e) = 10; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 22},
				},
			},
		},
	)
}
