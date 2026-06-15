// TestNoMultipleResolvedExtras locks in branches and edge shapes that the upstream test suite
// doesn't exercise. Each case carries an inline comment pointing at the specific branch /
// Dimension 4 row / tsgo AST quirk it covers, so future refactors can't silently regress
// them without breaking a named lock-in.
package no_multiple_resolved_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/promise/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/no_multiple_resolved"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoMultipleResolvedExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_multiple_resolved.NoMultipleResolvedRule,
		[]rule_tester.ValidTestCase{

			// ---- Dimension 4: parenthesized Promise constructor ----
			// tsgo preserves paren nodes; SkipParentheses must unwrap before checking callee
			{Code: `new (Promise)((resolve, reject) => {
        if (error) { reject(error) } else { resolve(value) }
      })`},

			// ---- Dimension 4: parenthesized resolver call ----
			{Code: `new Promise((resolve, reject) => {
        if (error) { reject(error) } else { (resolve)(value) }
      })`},

			// ---- Dimension 4: TypeScript this parameter ignored ----
			// The TS 'this' param must be stripped so it isn't counted as a resolver param
			{Code: `new Promise(function(this: any, resolve: () => void, reject: (e: any) => void) {
        if (error) { reject(error) } else { resolve() }
      })`},

			// ---- Dimension 4: arrow function with expression body ----
			// Expression-body arrows don't have a Block; must handle them directly
			{Code: `new Promise((resolve) => resolve(42))`},

			// ---- Dimension 4: nested Promise shadows outer resolver names ----
			// Inner new Promise((resolve, reject) => ...) shadows outer names, so inner
			// calls are analyzed separately by the listener; outer sees no resolver calls
			{Code: `new Promise((resolve, reject) => {
        new Promise((resolve, reject) => {
          if (e) { reject(e) } else { resolve(v) } // valid inner promise
        })
        resolve(v) // outer: only one call
      })`},

			// ---- Dimension 4: executor with only one parameter, if/else both resolve ----
			// if/else branches are mutually exclusive; no single path resolves twice.
			{Code: `new Promise((resolve) => {
        if (cond) {
          resolve()
        } else {
          resolve()
        }
      })`},

			// ---- Dimension 4: no executor arguments ----
			// No resolver params → no tracking needed
			{Code: `new Promise(() => { someCall() })`},

			// ---- Dimension 4: non-Promise new expression ignored ----
			{Code: `new NotPromise((resolve, reject) => { reject(e); resolve(v) })`},

			// ---- Dimension 4: nested function shadows resolve ----
			// Inner function has its own 'resolve' param → shadowed, outer resolve not tracked inside
			{Code: `new Promise((resolve, reject) => {
        fn(function(resolve) {  // shadows outer resolve
          resolve()   // inner resolve, not outer
          resolve()   // inner resolve, not outer
        })
        resolve(v)  // outer: only one call
      })`},

			// ---- Dimension 2: deeply nested scopes — only inner callback analyzed ----
			{Code: `new Promise((resolve, reject) => {
        setTimeout(() => {
          if (error) {
            reject(error)
          } else {
            resolve(value)
          }
        }, 0)
      })`},

			// ---- Dimension 2: labeled statement passes through ----
			{Code: `new Promise((resolve, reject) => {
        outer: {
          if (error) { reject(error); break outer }
          resolve(value)
        }
      })`},

			// ---- Dimension 1: try-catch where resolver is last throwable ----
			// With non-async try block — resolver call IS the last throwable
			{Code: `new Promise((resolve, reject) => {
        try {
          resolve(mayThrow())
        } catch (e) {
          reject(e)
        }
      })`},

			// ---- Real-user: promise wrapping node-callback style ----
			// Common pattern: wrapping a node-style callback API
			{Code: `new Promise((resolve, reject) => {
        fs.readFile(path, (err, data) => {
          if (err) {
            reject(err)
            return
          }
          resolve(data)
        })
      })`},

			// ---- Real-user: chained callbacks with early returns ----
			{Code: `new Promise((resolve, reject) => {
        step1((err1, r1) => {
          if (err1) { reject(err1); return }
          step2((err2, r2) => {
            if (err2) { reject(err2); return }
            resolve(r2)
          })
        })
      })`},

			// N/A: Dimension 3 (autofix) — this rule has no autofix
			// N/A: optional chaining on resolver call (e.g. resolve?.()) — no test case in upstream

			// ---- Try-catch where await occurs after resolve ----
			// await after resolve should not trigger a false positive when catch handles reject
			{Code: `new Promise(async (resolve, reject) => {
        try {
          const r = await foo();
          resolve();
          await r;
        } catch (error) {
          reject(error);
        }
      })`},

			// ---- Labeled break from outer loop avoids false positive ----
			{Code: `new Promise((resolve, reject) => {
        outer: while (foo) {
          switch (x) {
            case 1: resolve(a); break outer
          }
          reject(b)
        }
      })`},

			// ---- Unlabeled break inside labeled block inside loop ----
			{Code: `new Promise((resolve, reject) => {
        while (foo) {
          bar: {
            resolve(1);
            break;
          }
          resolve(2); // unreachable, should not be reported
        }
      })`},
		},
		[]rule_tester.InvalidTestCase{

			// ---- Dimension 4: parenthesized executor call ----
			// (resolve) is an Identifier after skipping parens
			{
				Code: `new Promise(((resolve, reject) => {
        reject(e)
        resolve(v)
      }))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "Promise should not be resolved multiple times. Promise is already resolved on line 2.",
						Line:    3,
					},
				},
			},

			// ---- Dimension 1: do-while body always runs ----
			// do-while body runs at least once; resolver call within it counts
			{
				Code: `new Promise((resolve, reject) => {
        do {
          reject(e)
        } while (false)
        resolve(v)
      })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "Promise should not be resolved multiple times. Promise is already resolved on line 3.",
						Line:    5,
					},
				},
			},

			// ---- Dimension 1: for-loop body treated as potential ----
			{
				Code: `new Promise((resolve, reject) => {
        for (let i = 0; i < n; i++) {
          reject(e)
        }
        resolve(v)
      })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "Promise should not be resolved multiple times. Promise is potentially resolved on line 3.",
						Line:    5,
					},
				},
			},

			// ---- Dimension 1: for-in body treated as potential ----
			{
				Code: `new Promise((resolve, reject) => {
        for (const k in obj) {
          reject(k)
        }
        resolve(v)
      })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "Promise should not be resolved multiple times. Promise is potentially resolved on line 3.",
						Line:    5,
					},
				},
			},

			// ---- Dimension 1: for-of body treated as potential ----
			{
				Code: `new Promise((resolve, reject) => {
        for (const x of arr) {
          reject(x)
        }
        resolve(v)
      })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "Promise should not be resolved multiple times. Promise is potentially resolved on line 3.",
						Line:    5,
					},
				},
			},

			// ---- Dimension 1: switch with resolve in one case + after switch ----
			// reject(e) in case 1 (line 3) is the first settlement; outer resolve(v) conflicts.
			{
				Code: `new Promise((resolve, reject) => {
        switch (x) {
          case 1: reject(e); break
          case 2: resolve(v); break
        }
        resolve(v)
      })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "Promise should not be resolved multiple times. Promise is potentially resolved on line 3.",
						Line:    6,
					},
				},
			},

			// Locks in upstream computeCatchEntryState arm: try-catch where throwable is
			// NOT the resolver call → catch entry state inherits try's resolution
			{
				Code: `new Promise((resolve, reject) => {
        try {
          resolve(v)
          mayThrow()
        } catch (e) {
          reject(e)
        }
      })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "Promise should not be resolved multiple times. Promise is potentially resolved on line 3.",
						Line:    6,
					},
				},
			},

			// ---- Real-user: resolve then resolve (same name) ----
			{
				Code: `new Promise((resolve) => {
        resolve(1)
        resolve(2)
      })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "Promise should not be resolved multiple times. Promise is already resolved on line 2.",
						Line:    3,
					},
				},
			},

			// ---- Real-user: reject in nested callback, then resolve at executor level ----
			{
				Code: `new Promise((resolve, reject) => {
        fn((error, value) => {
          reject(error)
          resolve(value)
        })
      })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "Promise should not be resolved multiple times. Promise is already resolved on line 3.",
						Line:    4,
					},
				},
			},

			// ---- Fallthrough switch potential double resolution ----
			{
				Code: `new Promise((resolve, reject) => {
        switch (x) {
          case 1: reject(e)
          case 2: resolve(v); break
        }
      })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "Promise should not be resolved multiple times. Promise is potentially resolved on line 3.",
						Line:    4,
					},
				},
			},

			// ---- Function declaration inside executor ----
			{
				Code: `new Promise((resolve, reject) => {
        function onDone(e, v) {
          reject(e)
          resolve(v)
        }
        fn(onDone)
      })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "Promise should not be resolved multiple times. Promise is already resolved on line 3.",
						Line:    4,
					},
				},
			},

			// ---- Nested executor calling outer resolve multiple times ----
			{
				Code: `new Promise((resolve) => {
        new Promise((res) => {
          resolve(1)
          resolve(2)
        })
      })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "Promise should not be resolved multiple times. Promise is already resolved on line 3.",
						Line:    4,
					},
				},
			},

			// ---- Resolver call inside return statement expression ----
			{
				Code: `new Promise((resolve, reject) => {
        if (error) reject(error)
        return resolve(value)
      })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "Promise should not be resolved multiple times. Promise is potentially resolved on line 2.",
						Line:    3,
					},
				},
			},

			// ---- Resolver call inside conditional expression ----
			{
				Code: `new Promise((resolve, reject) => {
        error ? reject(error) : resolve(value)
        resolve(value)
      })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "Promise should not be resolved multiple times. Promise is already resolved on line 2.",
						Line:    3,
					},
				},
			},

			// ---- Resolver call inside binary logical expression ----
			{
				Code: `new Promise((resolve, reject) => {
        error && reject(error)
        resolve(value)
      })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "Promise should not be resolved multiple times. Promise is potentially resolved on line 2.",
						Line:    3,
					},
				},
			},

			// ---- Resolver call inside comma expression ----
			{
				Code: `new Promise((resolve, reject) => {
        (reject(e), resolve(v))
      })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "Promise should not be resolved multiple times. Promise is already resolved on line 2.",
						Line:    2,
					},
				},
			},

			// ---- Loop with conditional break rejoining after loop ----
			{
				Code: `new Promise((resolve, reject) => {
        while (foo) {
          if (error) { reject(error); break }
        }
        resolve(value)
      })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "Promise should not be resolved multiple times. Promise is potentially resolved on line 3.",
						Line:    5,
					},
				},
			},

			// ---- Loop with conditional continue rejoining after loop ----
			{
				Code: `new Promise((resolve, reject) => {
        while (foo) {
          if (error) { reject(error); continue }
        }
        resolve(value)
      })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "Promise should not be resolved multiple times. Promise is potentially resolved on line 3.",
						Line:    5,
					},
				},
			},

			// ---- Labeled block with break rejoining after block ----
			{
				Code: `new Promise((resolve, reject) => {
        outer: {
          if (error) { reject(error); break outer }
        }
        resolve(value)
      })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "Promise should not be resolved multiple times. Promise is potentially resolved on line 3.",
						Line:    5,
					},
				},
			},

			// ---- Statement-head condition resolve ----
			{
				Code: `new Promise((resolve) => {
        if (resolve(1)) { foo() }
        resolve(2)
      })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "Promise should not be resolved multiple times. Promise is already resolved on line 2.",
						Line:    3,
					},
				},
			},

			// ---- Statement-head while loop condition resolve ----
			{
				Code: `new Promise((resolve) => {
        while (resolve(1)) { foo() }
        resolve(2)
      })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "Promise should not be resolved multiple times. Promise is already resolved on line 2.",
						Line:    3,
					},
				},
			},

			// ---- Optional chaining resolve?.() followed by resolve ----
			{
				Code: `new Promise((resolve) => {
        resolve?.(1)
        resolve(2)
      })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "Promise should not be resolved multiple times. Promise is potentially resolved on line 2.",
						Line:    3,
					},
				},
			},

			// ---- Certain resolve followed by optional chaining resolve?.() ----
			{
				Code: `new Promise((resolve) => {
        resolve(1)
        resolve?.(2)
      })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "Promise should not be resolved multiple times. Promise is already resolved on line 2.",
						Line:    3,
					},
				},
			},
		},
	)
}
