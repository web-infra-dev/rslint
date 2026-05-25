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
			// ====================================================================
			// No control flow in finally
			// ====================================================================
			{Code: `var foo = function() { try { return 1; } catch(err) { return 2; } finally { console.log("hola!") } }`},

			// ====================================================================
			// Function-like boundaries: all 7 kinds stop return/throw propagation
			// ====================================================================
			// FunctionDeclaration
			{Code: `var foo = function() { try {} finally { function a(x) { return x } } }`},
			// FunctionExpression
			{Code: `var foo = function() { try {} finally { var a = function(x) { return x } } }`},
			// ArrowFunction
			{Code: `var foo = function() { try {} finally { var a = (x) => { return x } } }`},
			// MethodDeclaration (object literal)
			{Code: `var foo = function() { try {} finally { var obj = { method() { return 1 } } } }`},
			// GetAccessor
			{Code: `var foo = function() { try {} finally { var obj = { get x() { return 1 } } } }`},
			// SetAccessor
			{Code: `var foo = function() { try {} finally { var obj = { set x(v) { return } } } }`},
			// Constructor
			{Code: `var foo = function() { try {} finally { class C { constructor() { return } } } }`},

			// Special function variants
			// async function
			{Code: `var foo = function() { try {} finally { async function a() { return 1 } } }`},
			// generator function
			{Code: `var foo = function() { try {} finally { function* gen() { return 1 } } }`},
			// async generator
			{Code: `var foo = function() { try {} finally { async function* gen() { return 1 } } }`},
			// async arrow
			{Code: `var foo = function() { try {} finally { var a = async () => { return 1 } } }`},

			// Throw inside nested function
			{Code: `var foo = function() { try {} finally { function a() { throw new Error() } } }`},
			// Complex control flow inside nested function (from ESLint original tests)
			{Code: `var foo = function() { try { return 1 } catch(err) { return 2 } finally { var a = function(x) { if(!x) { throw new Error() } } } }`},
			{Code: `var foo = function() { try { return 1 } catch(err) { return 2 } finally { var a = function(x) { while(true) { if(x) { break } else { continue } } } } }`},
			{Code: `var foo = function() { try { return 1 } catch(err) { return 2 } finally { var a = function(x) { label: while(true) { if(x) { break label; } else { continue } } } } }`},
			// Arrow function body expression (no control flow statement)
			{Code: `var foo = function() { try { return 1; } catch(err) { return 2 } finally { (x) => x } }`},

			// ====================================================================
			// Class-like boundaries: ClassDeclaration and ClassExpression
			// ====================================================================
			// ClassDeclaration — return in method
			{Code: `var foo = function() { try {} finally { class C { method() { return 1 } } } }`},
			// ClassExpression — return in method
			{Code: `var foo = function() { try {} finally { var C = class { method() { return 1 } } } }`},
			// ClassDeclaration — throw in static method
			{Code: `var foo = function() { try {} finally { class C { static fail() { throw new Error() } } } }`},
			// Multi-level: arrow function inside class method inside finally
			{Code: `var foo = function() { try {} finally { class C { method() { var fn = () => { return 1 } } } } }`},

			// ====================================================================
			// Loop (iteration) boundaries for unlabeled break: all 5 kinds
			// ====================================================================
			{Code: `var foo = function() { try {} finally { while (true) break; } }`},
			{Code: `var foo = function() { try {} finally { for (var i = 0; i < 10; i++) break; } }`},
			{Code: `var foo = function() { try {} finally { for (var x in obj) break; } }`},
			{Code: `var foo = function() { try {} finally { for (var x of arr) break; } }`},
			{Code: `var foo = function() { try {} finally { do { break; } while (true); } }`},

			// ====================================================================
			// Loop boundaries for unlabeled continue: all 5 kinds
			// ====================================================================
			{Code: `var foo = function() { try {} finally { while (true) continue; } }`},
			{Code: `var foo = function() { try {} finally { for (var i = 0; i < 10; i++) continue; } }`},
			{Code: `var foo = function() { try {} finally { for (var x in obj) continue; } }`},
			{Code: `var foo = function() { try {} finally { for (var x of arr) continue; } }`},
			{Code: `var foo = function() { try {} finally { do { continue; } while (true); } }`},

			// ====================================================================
			// Switch boundary for unlabeled break (NOT for continue)
			// ====================================================================
			{Code: `var foo = function() { try {} finally { switch (true) { case true: break; } } }`},

			// ====================================================================
			// Labeled break/continue with label INSIDE finally (safe)
			// ====================================================================
			// Label on loop
			{Code: `var foo = function() { try {} finally { label: while (true) { break label; } } }`},
			{Code: `var foo = function() { try {} finally { label: while (true) { continue label; } } }`},
			{Code: `var foo = function() { try {} finally { label: for (var i = 0; i < 10; i++) { break label; } } }`},
			{Code: `var foo = function() { try {} finally { label: for (var i = 0; i < 10; i++) { continue label; } } }`},
			{Code: `var foo = function() { try {} finally { label: for (var x in obj) { break label; } } }`},
			{Code: `var foo = function() { try {} finally { label: for (var x of arr) { continue label; } } }`},
			{Code: `var foo = function() { try {} finally { label: do { break label; } while (true); } }`},
			// Label on plain block (break only, continue on block is invalid syntax)
			{Code: `var foo = function() { try {} finally { label: { break label; } } }`},

			// ====================================================================
			// Labeled continue with intermediate loop in finally
			// (loop is sentinel per ESLint behavior — all 5 loop types)
			// ====================================================================
			{Code: `label: while (true) { try {} finally { while (true) { continue label; } } }`},
			{Code: `label: while (true) { try {} finally { for (var i = 0; i < 10; i++) { continue label; } } }`},
			{Code: `label: while (true) { try {} finally { for (var x in obj) { continue label; } } }`},
			{Code: `label: while (true) { try {} finally { for (var x of arr) { continue label; } } }`},
			{Code: `label: while (true) { try {} finally { do { continue label; } while (true); } }`},

			// ====================================================================
			// No finally block at all
			// ====================================================================
			{Code: `var foo = function() { try { return 1 } catch(err) { return 2 } }`},

			// Control flow in try/catch but not finally
			{Code: `var foo = function() { try { throw new Error() } catch(err) { return 2 } finally { console.log("done") } }`},
			{Code: `var foo = function() { try { return 1 } catch(err) { throw new Error() } finally { console.log("done") } }`},

			// ====================================================================
			// Deep nesting — safe due to function/class boundaries
			// ====================================================================
			// Return in arrow inside if inside finally
			{Code: `var foo = function() { try {} finally { if (true) { var fn = () => { return 1 }; } } }`},
			// Return in function inside loop inside finally
			{Code: `var foo = function() { try {} finally { while (true) { (function() { return 1 })(); break; } } }`},
			// Break in loop inside nested try (safe because loop is sentinel)
			{Code: `var foo = function() { try {} finally { try { while (true) { break; } } catch(e) {} } }`},
			// Deeply nested: function → class → arrow
			{Code: `var foo = function() { try {} finally { (function() { class C { method() { return (() => { return 1 })() } } })() } }`},

			// ====================================================================
			// Nested try-finally — only inner finally matters
			// ====================================================================
			// Inner try-finally with no control flow in either finally
			{Code: `var foo = function() { try {} finally { try { console.log(1) } finally { console.log(2) } } }`},
			// Return in inner try body (not in any finally)
			{Code: `var foo = function() { try {} finally { var fn = function() { try { return 1 } finally { console.log(2) } } } }`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// ====================================================================
			// Basic: return in finally
			// ====================================================================
			{
				Code: `var foo = function() { try { return 1; } catch(err) { return 2; } finally { return 3; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1, Column: 77},
				},
			},
			{
				Code: `var foo = function() { try { return 1; } finally { return 3; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1, Column: 52},
				},
			},

			// ====================================================================
			// Basic: throw in finally
			// ====================================================================
			{
				Code: `var foo = function() { try { return 1 } catch(err) { return 2 } finally { throw new Error() } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1, Column: 75},
				},
			},
			{
				Code: `var foo = function() { try {} finally { throw new Error() } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1, Column: 41},
				},
			},

			// ====================================================================
			// Unlabeled break/continue — targeting loop/switch OUTSIDE finally
			// ====================================================================
			{
				Code: `while (true) try {} finally { break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1, Column: 31},
				},
			},
			{
				Code: `while (true) try {} finally { continue; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1, Column: 31},
				},
			},
			{
				Code: `for (;;) try {} finally { break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1},
				},
			},
			{
				Code: `for (var x in obj) try {} finally { continue; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1},
				},
			},
			{
				Code: `for (var x of arr) try {} finally { break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1},
				},
			},
			{
				Code: `do { try {} finally { break; } } while (true)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1},
				},
			},

			// ====================================================================
			// Labeled break — label OUTSIDE finally
			// ====================================================================
			{
				Code: `label: try { return 0; } finally { break label; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1, Column: 36},
				},
			},
			{
				Code: `label: try {} finally { break label; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1, Column: 25},
				},
			},
			// Labeled break — inner label exists but targets outer
			{
				Code: `outer: { try {} finally { inner: { break outer; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1},
				},
			},

			// ====================================================================
			// Labeled continue — label OUTSIDE finally (no intermediate loop)
			// ====================================================================
			{
				Code: `label: while (true) try {} finally { continue label; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1, Column: 38},
				},
			},
			{
				Code: `label: for (;;) try {} finally { continue label; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1},
				},
			},

			// ====================================================================
			// Continue passes through switch (switch is NOT sentinel for continue)
			// ====================================================================
			{
				Code: `while (true) try {} finally { switch (true) { case true: continue; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1, Column: 58},
				},
			},

			// ====================================================================
			// Control flow inside non-boundary constructs in finally
			// (if/else, blocks, conditional — these are NOT sentinels)
			// ====================================================================
			// Return inside if in finally
			{
				Code: `var foo = function() { try {} finally { if (true) { return 1; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1},
				},
			},
			// Return inside if-else in finally
			{
				Code: `var foo = function() { try {} finally { if (true) { return 1; } else { return 2; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1},
					{MessageId: "unsafeUsage", Line: 1},
				},
			},
			// Throw inside conditional in finally
			{
				Code: `var foo = function() { try {} finally { if (cond) throw new Error() } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1},
				},
			},
			// Return inside nested blocks in finally
			{
				Code: `var foo = function() { try {} finally { { { return 1; } } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1},
				},
			},

			// ====================================================================
			// Control flow in try/catch body INSIDE outer finally
			// (the whole try-catch is inside the outer finally block)
			// ====================================================================
			// Return in inner try body
			{
				Code: `var foo = function() { try {} finally { try { return 1; } catch(e) {} } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1},
				},
			},
			// Return in inner catch body
			{
				Code: `var foo = function() { try {} finally { try { x } catch(e) { return 1; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1},
				},
			},
			// Throw in inner catch body inside outer finally
			{
				Code: `var foo = function() { try {} finally { try { x } catch(e) { throw e; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1},
				},
			},

			// ====================================================================
			// Nested try-finally — inner finally
			// ====================================================================
			{
				Code: `var foo = function() { try {} finally { try {} finally { return 1; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1, Column: 58},
				},
			},
			// Throw in nested inner finally
			{
				Code: `var foo = function() { try {} finally { try {} finally { throw new Error(); } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1},
				},
			},
			// Triple nesting — return in innermost finally
			{
				Code: `var foo = function() { try {} finally { try {} finally { try {} finally { return 1; } } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1},
				},
			},

			// ====================================================================
			// Multiple unsafe statements in one finally
			// ====================================================================
			{
				Code: `var foo = function() { try {} finally { return 1; return 2; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1, Column: 41},
					{MessageId: "unsafeUsage", Line: 1, Column: 51},
				},
			},
			// Mixed statement types in finally
			{
				Code: `var foo = function() { try {} finally { return 1; throw new Error(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1},
					{MessageId: "unsafeUsage", Line: 1},
				},
			},

			// ====================================================================
			// Complex nesting — unsafe despite intermediate non-boundary constructs
			// ====================================================================
			// Return in for-loop body inside finally (loop doesn't stop return)
			{
				Code: `var foo = function() { try {} finally { for (var i = 0; i < 1; i++) { return 1; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1},
				},
			},
			// Throw in switch case inside finally (switch doesn't stop throw)
			{
				Code: `var foo = function() { try {} finally { switch (x) { case 1: throw new Error(); } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1},
				},
			},
			// Return deep in if → for → switch inside finally (none are sentinels for return)
			{
				Code: `var foo = function() { try {} finally { if (true) { for (var i = 0; i < 1; i++) { switch (x) { default: return 1; } } } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1},
				},
			},

			// ====================================================================
			// From ESLint original tests — return value contains function/object
			// (the outer return is still unsafe even though inner returns are safe)
			// ====================================================================
			{
				Code: `var foo = function() { try { return 1 } catch(err) { return 2 } finally { return function(x) { return y } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1},
				},
			},
			{
				Code: `var foo = function() { try { return 1 } catch(err) { return 2 } finally { return { x: function(c) { return c } } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1},
				},
			},

			// ====================================================================
			// From ESLint original tests — break/continue across switch boundaries
			// ====================================================================
			// Unlabeled break in finally inside switch case (switch is OUTSIDE finally)
			{
				Code: `var foo = function() { switch (true) { case true: try {} finally { break; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1},
				},
			},
			// Labeled break from switch case inside finally to outer label
			{
				Code: `var foo = function() { a: while (true) try {} finally { switch (true) { case true: break a; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1},
				},
			},
			// Labeled break across nested switches (outer switch label)
			{
				Code: `var foo = function() { a: switch (true) { case true: try {} finally { switch (true) { case true: break a; } } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1},
				},
			},
		},
	)
}
