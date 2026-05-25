package default_case_last

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestDefaultCaseLastRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&DefaultCaseLastRule,
		[]rule_tester.ValidTestCase{
			// basic cases
			{Code: `switch (foo) {}`},
			{Code: `switch (foo) { case 1: bar(); break; }`},
			{Code: `switch (foo) { case 1: break; case 2: break; }`},
			{Code: `switch (foo) { default: bar(); break; }`},
			{Code: `switch (foo) { default: }`},
			{Code: `switch (foo) { case 1: break; default: break; }`},
			{Code: `switch (foo) { case 1: break; default: }`},
			{Code: `switch (foo) { case 1: default: break; }`},
			{Code: `switch (foo) { case 1: default: }`},
			{Code: `switch (foo) { case 1: break; case 2: break; default: break; }`},
			{Code: `switch (foo) { case 1: break; case 2: default: break; }`},
			{Code: `switch (foo) { case 1: case 2: default: }`},
			{Code: `switch (foo) { case 1: break; }`},
			{Code: `switch (foo) { case 1: case 2: break; }`},
			{Code: `switch (foo) { case 1: baz(); break; case 2: quux(); break; default: quuux(); break; }`},

			// nested switch — both default last
			{Code: `switch (a) { case 1: switch (b) { case 2: break; default: break; } break; default: break; }`},

			// switch inside default clause — both valid
			{Code: `switch (a) { case 1: break; default: switch (b) { case 2: break; default: break; } }`},

			// triple-nested — all default last
			{Code: `switch (a) { case 1: switch (b) { case 2: switch (c) { case 3: break; default: break; } break; default: break; } break; default: break; }`},

			// switch inside function/arrow/class method — default last
			{Code: `function f() { switch (a) { case 1: return 1; default: return 0; } }`},
			{Code: `const f = () => { switch (a) { case 1: return 1; default: return 0; } }`},
			{Code: `class C { m() { switch (a) { case 1: break; default: break; } } }`},

			// switch inside control structures — default last
			{Code: `if (x) { switch (a) { case 1: break; default: break; } }`},
			{Code: `for (let i = 0; i < 10; i++) { switch (a) { case 1: break; default: break; } }`},
			{Code: `while (x) { switch (a) { case 1: break; default: break; } }`},
			{Code: `try { switch (a) { case 1: break; default: break; } } catch (e) { switch (a) { case 1: break; default: break; } }`},

			// switch inside IIFE and labeled statement — default last
			{Code: `(function() { switch (a) { case 1: break; default: break; } })()`},
			{Code: `label: switch (a) { case 1: break; default: break; }`},

			// many fall-through cases then default
			{Code: `switch (a) { case 1: case 2: case 3: console.log('x'); break; default: break; }`},

			// do-while / for-in / for-of — default last
			{Code: `do { switch (a) { case 1: break; default: break; } } while (false)`},
			{Code: `for (const k in obj) { switch (k) { case 'a': break; default: break; } }`},
			{Code: `for (const v of arr) { switch (v) { case 1: break; default: break; } }`},

			// getter / setter / static block — default last
			{Code: `class G { get x() { switch (a) { case 1: return 1; default: return 0; } } }`},
			{Code: `class S { set x(v: any) { switch (v) { case 1: break; default: break; } } }`},
			{Code: `class SB { static { switch (a) { case 1: break; default: break; } } }`},

			// generator / async function — default last
			{Code: `function* gen() { switch (a) { case 1: yield 1; break; default: yield 0; } }`},
			{Code: `async function af() { switch (a) { case 1: break; default: break; } }`},

			// finally block — default last
			{Code: `try {} finally { switch (a) { case 1: break; default: break; } }`},

			// TypeScript namespace — default last
			{Code: `namespace NS { switch (a) { case 1: break; default: break; } }`},
		},
		[]rule_tester.InvalidTestCase{
			// basic cases
			{
				Code: `switch (foo) { default: bar(); break; case 1: baz(); break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 16},
				},
			},
			{
				Code: `switch (foo) { default: break; case 1: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 16},
				},
			},
			{
				Code: `switch (foo) { default: case 1: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 16},
				},
			},
			{
				Code: `switch (foo) { default: case 1: }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 16},
				},
			},
			{
				Code: `switch (foo) { case 1: break; default: break; case 2: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 31},
				},
			},
			{
				Code: `switch (foo) { case 1: default: break; case 2: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 24},
				},
			},
			{
				Code: `switch (foo) { case 1: default: case 2: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 24},
				},
			},
			{
				Code: `switch (foo) { case 1: default: case 2: }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 24},
				},
			},
			{
				Code: `switch (foo) { default: break; case 1: break; case 2: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 16},
				},
			},
			{
				Code: `switch (foo) { default: case 1: case 2: }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 16},
				},
			},

			// nested switch — outer default not last, inner valid (only outer reports)
			{
				Code: `switch (a) { default: switch (b) { case 2: break; default: break; } break; case 1: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 14},
				},
			},

			// nested switch — outer valid, inner default not last (only inner reports)
			{
				Code: `switch (a) { case 1: switch (b) { default: break; case 2: break; } break; default: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 35},
				},
			},

			// nested switch — both default not last (both report)
			{
				Code: `switch (a) { default: switch (b) { default: break; case 2: break; } break; case 1: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 14},
					{MessageId: "notLast", Line: 1, Column: 36},
				},
			},

			// triple-nested — only innermost invalid
			{
				Code: `switch (a) { case 1: switch (b) { case 2: switch (c) { default: break; case 3: break; } break; default: break; } break; default: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 56},
				},
			},

			// triple-nested — all three invalid
			{
				Code: `switch (a) { default: switch (b) { default: switch (c) { default: break; case 3: break; } break; case 2: break; } break; case 1: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 14},
					{MessageId: "notLast", Line: 1, Column: 36},
					{MessageId: "notLast", Line: 1, Column: 58},
				},
			},

			// switch inside function — invalid
			{
				Code: `function f() { switch (a) { default: return 0; case 1: return 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 29},
				},
			},

			// switch inside arrow function — invalid
			{
				Code: `const f = () => { switch (a) { default: return 0; case 1: return 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 32},
				},
			},

			// switch inside class method — invalid
			{
				Code: `class C { m() { switch (a) { default: break; case 1: break; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 30},
				},
			},

			// switch inside if — invalid
			{
				Code: `if (x) { switch (a) { default: break; case 1: break; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 23},
				},
			},

			// switch inside for loop — invalid
			{
				Code: `for (let i = 0; i < 10; i++) { switch (a) { default: break; case 1: break; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 45},
				},
			},

			// switch inside while — invalid
			{
				Code: `while (x) { switch (a) { default: break; case 1: break; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 26},
				},
			},

			// switch inside try/catch — both invalid
			{
				Code: `try { switch (a) { default: break; case 1: break; } } catch (e) { switch (a) { default: break; case 1: break; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 20},
					{MessageId: "notLast", Line: 1, Column: 80},
				},
			},

			// switch inside IIFE — invalid
			{
				Code: `(function() { switch (a) { default: break; case 1: break; } })()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 28},
				},
			},

			// labeled switch — invalid
			{
				Code: `label: switch (a) { default: break; case 1: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 21},
				},
			},

			// switch inside default clause of outer — inner invalid, outer valid
			{
				Code: `switch (a) { case 1: break; default: switch (b) { default: break; case 2: break; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 51},
				},
			},

			// multiple invalid switches in same block
			{
				Code: `switch (x) { default: break; case 1: break; } switch (y) { default: break; case 2: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 14},
					{MessageId: "notLast", Line: 1, Column: 60},
				},
			},

			// default with throw not last
			{
				Code: `function f() { switch (a) { default: throw new Error('no'); case 1: return 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 29},
				},
			},

			// empty default not last
			{
				Code: `switch (a) { default: case 1: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 14},
				},
			},

			// default with fall-through not last
			{
				Code: `switch (a) { case 1: default: case 2: case 3: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 22},
				},
			},

			// deeply nested in complex structure
			{
				Code: `function f() { for (let i = 0; i < 10; i++) { if (i > 5) { try { switch (a) { default: break; case 1: break; } } catch (e) {} } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 79},
				},
			},

			// do-while — invalid
			{
				Code: `do { switch (a) { default: break; case 1: break; } } while (false)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 19},
				},
			},

			// for-in — invalid
			{
				Code: `for (const k in obj) { switch (k) { default: break; case 'a': break; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 37},
				},
			},

			// for-of — invalid
			{
				Code: `for (const v of arr) { switch (v) { default: break; case 1: break; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 37},
				},
			},

			// getter — invalid
			{
				Code: `class G { get x() { switch (a) { default: return 1; case 1: return 0; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 34},
				},
			},

			// setter — invalid
			{
				Code: `class S { set x(v: any) { switch (v) { default: break; case 1: break; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 40},
				},
			},

			// static block — invalid
			{
				Code: `class SB { static { switch (a) { default: break; case 1: break; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 34},
				},
			},

			// generator function — invalid
			{
				Code: `function* gen() { switch (a) { default: yield 0; break; case 1: yield 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 32},
				},
			},

			// async function — invalid
			{
				Code: `async function af() { switch (a) { default: break; case 1: break; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 36},
				},
			},

			// finally block — invalid
			{
				Code: `try {} finally { switch (a) { default: break; case 1: break; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 31},
				},
			},

			// TypeScript namespace — invalid
			{
				Code: `namespace NS { switch (a) { default: break; case 1: break; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 29},
				},
			},
		},
	)
}
