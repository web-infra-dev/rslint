package no_unsafe_finally

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnsafeFinallyRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnsafeFinallyRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// No control flow in finally
			{Code: `var foo = function() { try { return 1; } catch(err) { return 2; } finally { console.log("hola!") } }`},

			// Return inside a nested function in finally is safe
			{Code: `var foo = function() { try { return 1 } catch(err) { return 2 } finally { function a(x) { return x } } }`},
			{Code: `var foo = function() { try { return 1 } catch(err) { return 2 } finally { var a = function(x) { return x } } }`},
			{Code: `var foo = function() { try { return 1 } catch(err) { return 2 } finally { var a = (x) => { return x } } }`},

			// Break inside a loop in finally is safe
			{Code: `var foo = function() { try {} finally { while (true) break; } }`},
			{Code: `var foo = function() { try {} finally { for (var i = 0; i < 10; i++) break; } }`},
			{Code: `var foo = function() { try {} finally { for (var x in obj) break; } }`},
			{Code: `var foo = function() { try {} finally { for (var x of arr) break; } }`},
			{Code: `var foo = function() { try {} finally { do { break; } while (true); } }`},

			// Continue inside a loop in finally is safe
			{Code: `var foo = function() { try {} finally { while (true) continue; } }`},
			{Code: `var foo = function() { try {} finally { for (var i = 0; i < 10; i++) continue; } }`},

			// Break inside switch in finally is safe
			{Code: `var foo = function() { try {} finally { switch (true) { case true: break; } } }`},

			// Labeled break/continue where the label is inside finally is safe
			{Code: `var foo = function() { try {} finally { label: while (true) { break label; } } }`},
			{Code: `var foo = function() { try {} finally { label: while (true) { continue label; } } }`},
			{Code: `var foo = function() { try {} finally { label: for (var i = 0; i < 10; i++) { break label; } } }`},

			// Throw inside nested function in finally is safe
			{Code: `var foo = function() { try {} finally { function a() { throw new Error() } } }`},

			// Class boundary stops propagation
			{Code: `var foo = function() { try {} finally { class C { method() { return 1 } } } }`},

			// No finally block at all
			{Code: `var foo = function() { try { return 1 } catch(err) { return 2 } }`},

			// Control flow in try/catch but not finally
			{Code: `var foo = function() { try { throw new Error() } catch(err) { return 2 } finally { console.log("done") } }`},

			// Return in getter/setter inside finally is safe
			{Code: `var foo = function() { try {} finally { var obj = { get x() { return 1 } } } }`},
			{Code: `var foo = function() { try {} finally { var obj = { set x(v) { return } } } }`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// Return in finally
			{
				Code: `var foo = function() { try { return 1; } catch(err) { return 2; } finally { return 3; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1, Column: 77},
				},
			},
			// Return in finally (no catch)
			{
				Code: `var foo = function() { try { return 1; } finally { return 3; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1, Column: 52},
				},
			},
			// Throw in finally
			{
				Code: `var foo = function() { try { return 1 } catch(err) { return 2 } finally { throw new Error() } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1, Column: 75},
				},
			},
			// Break with label targeting outside finally
			{
				Code: `label: try { return 0; } finally { break label; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1, Column: 36},
				},
			},
			// Unlabeled break exits loop outside finally
			{
				Code: `while (true) try {} finally { break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1, Column: 31},
				},
			},
			// Unlabeled continue exits loop outside finally
			{
				Code: `while (true) try {} finally { continue; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1, Column: 31},
				},
			},
			// Continue with label targeting outside finally
			{
				Code: `label: while (true) try {} finally { continue label; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1, Column: 38},
				},
			},
			// Return in nested try-finally (inner finally)
			{
				Code: `var foo = function() { try {} finally { try {} finally { return 1; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1, Column: 58},
				},
			},
			// Throw in finally (simple)
			{
				Code: `var foo = function() { try {} finally { throw new Error() } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1, Column: 41},
				},
			},
			// Break in finally (labeled, label outside)
			{
				Code: `label: try {} finally { break label; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1, Column: 25},
				},
			},
			// Continue in finally exits switch and loop outside
			{
				Code: `while (true) try {} finally { switch (true) { case true: continue; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1, Column: 58},
				},
			},
			// Multiple unsafe statements in finally
			{
				Code: `var foo = function() { try {} finally { return 1; return 2; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1, Column: 41},
					{MessageId: "unsafeUsage", Line: 1, Column: 51},
				},
			},
		},
	)
}
