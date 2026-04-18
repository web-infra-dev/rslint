package no_lone_blocks

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoLoneBlocksRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoLoneBlocksRule,
		[]rule_tester.ValidTestCase{
			// ---- Blocks that belong to a containing statement (not lone) ----
			{Code: `if (foo) { if (bar) { baz(); } }`},
			{Code: `if (foo) { bar(); } else { baz(); }`},
			{Code: `if (foo) { bar(); } else if (baz) { qux(); }`},
			{Code: `do { bar(); } while (foo)`},
			{Code: `while (foo) { bar(); }`},
			{Code: `for (let i = 0; i < 10; i++) { bar(); }`},
			{Code: `for (const x of xs) { bar(); }`},
			{Code: `for (const x in obj) { bar(); }`},
			{Code: `async function f() { for await (const x of xs) { bar(); } }`},
			{Code: `try { foo(); } catch (e) { bar(); }`},
			{Code: `try { foo(); } catch { bar(); }`},
			{Code: `try { foo(); } finally { bar(); }`},
			{Code: `function foo() { while (bar) { baz(); } }`},
			{Code: `const f = () => { foo(); }`},
			{Code: `const f = async () => { await foo(); }`},
			{Code: `class C { method() { foo(); } }`},
			{Code: `class C { constructor() { foo(); } }`},
			{Code: `class C { get prop() { return 1; } }`},
			{Code: `class C { set prop(v) { this._v = v; } }`},
			{Code: `class C { static method() { foo(); } }`},
			{Code: `function* gen() { yield 1; }`},

			// ---- Block-level bindings justify a lone block ----
			{Code: `{ let x = 1; }`},
			{Code: `{ const x = 1; }`},
			{Code: `{ class Bar {} }`},
			{Code: `'use strict'; { function bar() {} }`},
			{Code: `export {}; { function bar() {} }`},
			{Code: `{ let x; var y; }`},
			{Code: `{ var x; let y; }`},
			{Code: `{ let x; const y = 1; class Z {} }`},

			// ---- Nested lone block, each with its own binding ----
			{Code: `{ {let y = 1;} let x = 1; }`},
			{Code: `{ let x = 1; { let y = 2; } }`},

			// ---- Different scopes, same name is fine ----
			{Code: `function foo() { { const x = 4 } const x = 3 }`},

			// ---- Switch: solo block per clause is allowed ----
			{Code: `
switch (foo) {
    case bar: {
        baz;
    }
}
`},
			{Code: `
switch (foo) {
    case bar: {
        baz;
    }
    case qux: {
        boop;
    }
}
`},
			{Code: `
switch (foo) {
    case bar:
    {
        baz;
    }
}
`},
			{Code: `
switch (foo) {
    default: {
        baz;
    }
}
`},
			{Code: `
switch (foo) {
    case bar: {
        a;
    }
    default: {
        b;
    }
}
`},

			// ---- Class static blocks ----
			{Code: `class C { static {} }`},
			{Code: `class C { static { foo; } }`},
			{Code: `class C { static { foo; bar; } }`},
			{Code: `class C { static { if (foo) { block; } } }`},
			{Code: `class C { static { lbl: { block; } } }`},
			{Code: `class C { static { { let block; } something; } }`},
			{Code: `class C { static { something; { const block = 1; } } }`},
			{Code: `class C { static { { function block(){} } something; } }`},
			{Code: `class C { static { something; { class block {} } } }`},

			// ---- Labeled block at program scope (parent is LabeledStatement, not "lone") ----
			{Code: `lbl: { foo(); }`},
			{Code: `lbl: { let x = 1; }`},

			// ---- TS namespace: a block inside a namespace body is not flagged ----
			// (ESLint's `isLoneBlock` only recognizes BlockStatement/StaticBlock/Program/SwitchCase,
			// and never flags blocks whose parent is a TSModuleBlock. rslint matches that.)
			{Code: `namespace N { { foo; } }`},
			{Code: `namespace N { { let x = 1; } }`},

			// ---- Explicit resource management (`using` / `await using`) ----
			{Code: `
{
    using x = makeDisposable();
}
`},
			{Code: `
async function f() {
    {
        await using x = makeDisposable();
    }
    bar();
}
`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Trivial program-scope lone blocks ----
			{
				Code: `{}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantBlock", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
				},
			},
			{
				Code: `{var x = 1;}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantBlock", Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: `foo(); {} bar();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantBlock", Line: 1, Column: 8, EndLine: 1, EndColumn: 10},
				},
			},

			// ---- Two sibling lone blocks at program scope ----
			{
				Code: `{} {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantBlock", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
					{MessageId: "redundantBlock", Line: 1, Column: 4, EndLine: 1, EndColumn: 6},
				},
			},

			// ---- Lone block nested inside a containing statement body ----
			{
				Code: `if (foo) { bar(); {} baz(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantNestedBlock", Line: 1, Column: 19, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code: `function foo() { bar(); {} baz(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantNestedBlock", Line: 1, Column: 25, EndLine: 1, EndColumn: 27},
				},
			},
			{
				Code: `while (foo) { {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantNestedBlock", Line: 1, Column: 15, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code: `for (;;) { {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantNestedBlock", Line: 1, Column: 12, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code: `do { {} } while (foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantNestedBlock", Line: 1, Column: 6, EndLine: 1, EndColumn: 8},
				},
			},
			{
				Code: `try { {} } catch {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantNestedBlock", Line: 1, Column: 7, EndLine: 1, EndColumn: 9},
				},
			},
			{
				Code: `try {} catch { {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantNestedBlock", Line: 1, Column: 16, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code: `try {} finally { {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantNestedBlock", Line: 1, Column: 18, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code: `class C { method() { {} } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantNestedBlock", Line: 1, Column: 22, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code: `class C { constructor() { {} } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantNestedBlock", Line: 1, Column: 27, EndLine: 1, EndColumn: 29},
				},
			},
			{
				Code: `const f = () => { {} };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantNestedBlock", Line: 1, Column: 19, EndLine: 1, EndColumn: 21},
				},
			},

			// ---- Multi-line lone block: full position assertion ----
			{
				Code: "{ \n{ } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantNestedBlock", Line: 2, Column: 1, EndLine: 2, EndColumn: 4},
					{MessageId: "redundantBlock", Line: 1, Column: 1, EndLine: 2, EndColumn: 6},
				},
			},
			{
				Code: "{\n    var x = 1;\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantBlock", Line: 1, Column: 1, EndLine: 3, EndColumn: 2},
				},
			},

			// ---- Triple nesting: every level reports (exit order: inner -> middle -> outer) ----
			{
				Code: `{ { {} } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantNestedBlock", Line: 1, Column: 5, EndLine: 1, EndColumn: 7},
					{MessageId: "redundantNestedBlock", Line: 1, Column: 3, EndLine: 1, EndColumn: 9},
					{MessageId: "redundantBlock", Line: 1, Column: 1, EndLine: 1, EndColumn: 11},
				},
			},

			// ---- Even in ES6+, a function declaration in a non-strict block does not
			// justify the block. ----
			{
				Code: `{ function bar() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantBlock", Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
				},
			},

			// ---- A block that only contains a non-binding (labeled statement / type-only /
			// var / declare) is still redundant. ----
			{
				Code: `{ lbl: foo(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantBlock", Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code: `{ type X = number; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantBlock", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code: `{ interface I {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantBlock", Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code: `{ declare var x: number; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantBlock", Line: 1, Column: 1, EndLine: 1, EndColumn: 27},
				},
			},
			{
				Code: `{ /* comment */ }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantBlock", Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
				},
			},

			// ---- Stack semantics: only bindings directly inside a block mark it. ----
			{
				// Outer has `let y` (marks outer). Inner has only `var` (does not mark inner).
				// Inner is reported as redundantNestedBlock; outer is marked and not reported.
				Code: "{ \n{var x = 1;}\n let y = 2; } {let z = 1;}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantNestedBlock", Line: 2, Column: 1},
				},
			},
			{
				// Inner has `let` (marks inner). Outer has only `var` (does not mark outer).
				// Inner is marked and skipped; outer is reported as redundantBlock.
				Code: "{ \n{let x = 1;}\n var y = 2; } {let z = 1;}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantBlock", Line: 1, Column: 1},
				},
			},
			{
				// Nothing is block-scoped anywhere. Every block reports.
				Code: "{ \n{var x = 1;}\n var y = 2; }\n {var z = 1;}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantNestedBlock", Line: 2, Column: 1},
					{MessageId: "redundantBlock", Line: 1, Column: 1},
					{MessageId: "redundantBlock", Line: 4, Column: 2},
				},
			},

			// ---- Switch: block that is not the sole statement of a case/default clause ----
			{
				Code: `
switch (foo) {
    case 1:
        foo();
        {
            bar;
        }
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantBlock", Line: 5, Column: 9},
				},
			},
			{
				Code: `
switch (foo) {
    case 1:
    {
        bar;
    }
    foo();
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantBlock", Line: 4, Column: 5},
				},
			},
			{
				Code: `
switch (foo) {
    default:
        foo();
        {
            bar;
        }
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantBlock", Line: 5, Column: 9},
				},
			},
			{
				Code: `
switch (foo) {
    default:
    {
        bar;
    }
    foo();
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantBlock", Line: 4, Column: 5},
				},
			},

			// ---- Function body containing a single lone block (else-if branch fires) ----
			{
				Code: `
function foo () {
    {
        const x = 4;
    }
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantNestedBlock", Line: 3, Column: 5},
				},
			},
			{
				Code: `
function foo () {
    {
        var x = 4;
    }
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantNestedBlock", Line: 3, Column: 5},
				},
			},

			// ---- Class static block cases ----
			{
				Code: `
class C {
    static {
        if (foo) {
            {
                let block;
            }
        }
    }
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantNestedBlock", Line: 5, Column: 13},
				},
			},
			{
				Code: `
class C {
    static {
        if (foo) {
            {
                block;
            }
            something;
        }
    }
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantNestedBlock", Line: 5, Column: 13},
				},
			},
			{
				Code: `
class C {
    static {
        {
            block;
        }
    }
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantNestedBlock", Line: 4, Column: 9},
				},
			},
			{
				Code: `
class C {
    static {
        {
            let block;
        }
    }
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantNestedBlock", Line: 4, Column: 9},
				},
			},
			{
				Code: `
class C {
    static {
        {
            const block = 1;
        }
    }
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantNestedBlock", Line: 4, Column: 9},
				},
			},
			{
				Code: `
class C {
    static {
        {
            function block() {}
        }
    }
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantNestedBlock", Line: 4, Column: 9},
				},
			},
			{
				Code: `
class C {
    static {
        {
            class block {}
        }
    }
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantNestedBlock", Line: 4, Column: 9},
				},
			},
			{
				Code: `
class C {
    static {
        {
            var block;
        }
        something;
    }
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantNestedBlock", Line: 4, Column: 9},
				},
			},
			{
				Code: `
class C {
    static {
        something;
        {
            var block;
        }
    }
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantNestedBlock", Line: 5, Column: 9},
				},
			},
			{
				Code: `
class C {
    static {
        {
            block;
        }
        something;
    }
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantNestedBlock", Line: 4, Column: 9},
				},
			},
			{
				Code: `
class C {
    static {
        something;
        {
            block;
        }
    }
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redundantNestedBlock", Line: 5, Column: 9},
				},
			},
		},
	)
}
