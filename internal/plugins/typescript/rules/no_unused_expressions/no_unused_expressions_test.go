package no_unused_expressions

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnusedExpressionsRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnusedExpressionsRule, []rule_tester.ValidTestCase{
		// ================================================================
		// Side-effect expressions (always valid)
		// ================================================================
		{Code: `function f(){}`},
		{Code: `a = b`},
		{Code: `new a`},
		{Code: `{}`},
		{Code: `f(); g()`},
		{Code: `i++`},
		{Code: `a()`},
		{Code: `delete foo.bar`},
		{Code: `void new C`},

		// ================================================================
		// Directives
		// ================================================================
		{Code: `"use strict";`},
		{Code: `"directive one"; "directive two"; f();`},
		{Code: `function foo() {"use strict"; return true; }`},
		{Code: `var foo = () => {"use strict"; return true; }`},
		{Code: `function foo() {"directive one"; "directive two"; f(); }`},

		// ================================================================
		// allowShortCircuit
		// ================================================================
		{Code: `a && a()`, Options: map[string]interface{}{"allowShortCircuit": true}},
		{Code: `a() || (b = c)`, Options: map[string]interface{}{"allowShortCircuit": true}},

		// ================================================================
		// allowTernary
		// ================================================================
		{Code: `a ? b() : c()`, Options: map[string]interface{}{"allowTernary": true}},
		{
			Code:    `a ? b() || (c = d) : e()`,
			Options: map[string]interface{}{"allowShortCircuit": true, "allowTernary": true},
		},

		// ================================================================
		// yield / await
		// ================================================================
		{Code: `function* foo(){ yield 0; }`},
		{Code: `async function foo() { await 5; }`},
		{Code: `async function foo() { await foo.bar; }`},
		{
			Code:    `async function foo() { bar && await baz; }`,
			Options: map[string]interface{}{"allowShortCircuit": true},
		},
		{
			Code:    `async function foo() { foo ? await bar : await baz; }`,
			Options: map[string]interface{}{"allowTernary": true},
		},

		// ================================================================
		// Tagged template literals
		// ================================================================
		{Code: "tag`tagged template literal`", Options: map[string]interface{}{"allowTaggedTemplates": true}},
		{
			Code:    `shouldNotBeAffectedByAllowTemplateTagsOption()`,
			Options: map[string]interface{}{"allowTaggedTemplates": true},
		},

		// ================================================================
		// Dynamic import
		// ================================================================
		{Code: `import('./foo')`},
		{Code: `import('./foo').then(() => {})`},

		// ================================================================
		// Optional chaining calls
		// ================================================================
		{Code: `func?.("foo")`},
		{Code: `obj?.foo("bar")`},
		{Code: `test.age?.toLocaleString();`},
		{Code: `one[2]?.[3][4]?.();`},
		{Code: `a?.['b']?.c();`},

		// ================================================================
		// Assignments (not unused)
		// ================================================================
		{Code: `let a = (a?.b).c;`},
		{Code: `let b = a?.['b'];`},
		{Code: `let c = one[2]?.[3][4];`},

		// ================================================================
		// TypeScript: module/namespace directives
		// ================================================================
		{Code: "module Foo {\n  'use strict';\n}"},
		{Code: "namespace Foo {\n  'use strict';\n\n  export class Foo {}\n  export class Bar {}\n}"},
		{Code: "function foo() {\n  'use strict';\n\n  return null;\n}"},

		// ================================================================
		// TypeScript: new with generics
		// ================================================================
		{Code: `class Foo<T> {} new Foo<string>();`},

		// ================================================================
		// allowShortCircuit with optional chaining
		// ================================================================
		{Code: `foo && foo?.();`, Options: map[string]interface{}{"allowShortCircuit": true}},
		{Code: `foo && import('./foo');`, Options: map[string]interface{}{"allowShortCircuit": true}},

		// ================================================================
		// allowTernary with imports
		// ================================================================
		{Code: `foo ? import('./foo') : import('./bar');`, Options: map[string]interface{}{"allowTernary": true}},

		// ================================================================
		// Update expressions (have side effects)
		// ================================================================
		{Code: `i--`},
		{Code: `--i`},
		{Code: `++i`},

		// ================================================================
		// Compound / logical assignments (have side effects)
		// ================================================================
		{Code: `a += 1`},
		{Code: `a &&= b`},
		{Code: `a ||= b`},
		{Code: `a ??= b`},

		// ================================================================
		// TypeScript assertions wrapping calls (inner has side effects)
		// ================================================================
		{Code: `foo() as any;`},
		{Code: `foo()!;`},
		{Code: `<any>foo();`},
		{Code: `foo() satisfies string;`},

		// ================================================================
		// TS non-null assertion wrapping call in short-circuit
		// ================================================================
		{Code: `foo && foo()!;`, Options: map[string]interface{}{"allowShortCircuit": true}},

		// ================================================================
		// Instantiation expression wrapping call (has side effects)
		// ================================================================
		{Code: `declare function getSet(): Set<unknown>; getSet()<string>();`},

		// ================================================================
		// Satisfies (not unwrapped, defaults to not disallowed)
		// ================================================================
		{Code: `declare const foo: string; foo satisfies string;`},
		{Code: `foo() satisfies string;`},

		// ================================================================
		// Combined allowShortCircuit + allowTernary
		// ================================================================
		{
			Code:    `a ? b && c() : d()`,
			Options: map[string]interface{}{"allowShortCircuit": true, "allowTernary": true},
		},
		{Code: `a ?? b()`, Options: map[string]interface{}{"allowShortCircuit": true}},
		{Code: `a || b()`, Options: map[string]interface{}{"allowShortCircuit": true}},

		// ================================================================
		// Yield without expression, generator
		// ================================================================
		{Code: `function* foo(){ yield; }`},

		// ================================================================
		// More directive edge cases
		// ================================================================
		{Code: `"use strict"; "use asm"; f();`},
		{Code: `var foo = () => {"use strict"; return true; }`},
		{Code: `function foo() { var foo = "use strict"; return true; }`},

		// ================================================================
		// Class static block — leading string treated as directive
		// ================================================================
		{Code: `class C { static { 'use strict'; } }`},

		// ================================================================
		// Deep nesting: TS wrappers + options combined
		// ================================================================
		// type assertion > parens > short-circuit > non-null > call
		{
			Code:    "declare const foo: Function | undefined;\n<any>(foo && foo()!)",
			Options: map[string]interface{}{"allowShortCircuit": true},
		},
		// instantiation > parens > short-circuit > call
		{
			Code:    `(Foo && Foo())<string, number>;`,
			Options: map[string]interface{}{"allowShortCircuit": true},
		},
		// deeply nested ternary + short-circuit both valid
		{
			Code:    `a ? (b && c()) : (d || e())`,
			Options: map[string]interface{}{"allowShortCircuit": true, "allowTernary": true},
		},
		// chained optional calls
		{Code: `a?.()?.b();`},
	}, []rule_tester.InvalidTestCase{
		// ================================================================
		// Basic unused expressions
		// ================================================================
		{
			Code:   `0`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		{
			Code:   `a`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		{
			Code:   `f(), 0`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		{
			Code:   `{0}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 2}},
		},
		{
			Code:   `[]`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		{
			Code:   `a && b();`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		{
			Code:   `a() || false`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		{
			Code:   `a || (b = c)`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},

		// ================================================================
		// Template literals
		// ================================================================
		{
			Code:   "`untagged template literal`",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		{
			Code:   "tag`tagged template literal`",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},

		// ================================================================
		// Options mismatches
		// ================================================================
		{
			Code:    `a && b()`,
			Options: map[string]interface{}{"allowTernary": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		{
			Code:    `a ? b() : c()`,
			Options: map[string]interface{}{"allowShortCircuit": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		{
			Code:    `a || b`,
			Options: map[string]interface{}{"allowShortCircuit": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		{
			Code:    `a() && b`,
			Options: map[string]interface{}{"allowShortCircuit": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		{
			Code:    `a ? b : 0`,
			Options: map[string]interface{}{"allowTernary": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		{
			Code:    `a ? b : c()`,
			Options: map[string]interface{}{"allowTernary": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},

		// ================================================================
		// Member access (no side effects)
		// ================================================================
		{
			Code:   `foo.bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},

		// ================================================================
		// Unary (no side effects: !, +)
		// ================================================================
		{
			Code:   `!a`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		{
			Code:   `+a`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},

		// ================================================================
		// Directives not at prologue
		// ================================================================
		{
			Code:   `"directive one"; f(); "directive two";`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 23}},
		},
		{
			Code:   `function foo() {"directive one"; f(); "directive two"; }`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 39}},
		},
		{
			Code:   `if (0) { "not a directive"; f(); }`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 10}},
		},
		{
			Code:   `function foo() { var foo = true; "use strict"; }`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 34}},
		},
		{
			Code:   `var foo = () => { var foo = true; "use strict"; }`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 35}},
		},

		// ================================================================
		// Untagged template literals with options
		// ================================================================
		{
			Code:    "`untagged template literal`",
			Options: map[string]interface{}{"allowTaggedTemplates": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		{
			Code:    "tag`tagged template literal`",
			Options: map[string]interface{}{"allowTaggedTemplates": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},

		// ================================================================
		// Optional chaining (member access without call)
		// ================================================================
		{
			Code:   `obj?.foo`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		{
			Code:   `obj?.foo.bar`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		{
			Code:   `obj?.foo().bar`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},

		// ================================================================
		// Class static block — leading string treated as directive (matches @typescript-eslint)
		// Non-leading string IS flagged.
		// ================================================================
		{
			Code:   `class C { static { const x = 1; 'use strict'; } }`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 33}},
		},

		// ================================================================
		// TypeScript-specific: optional chaining member access (from TS test file)
		// ================================================================
		{
			Code:   `if (0) 0;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 8}},
		},
		{
			Code:   "f(0), {};",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		{
			Code:   "a, b();",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		{
			Code: "a() &&\n  function namedFunctionInExpressionContext() {\n    f();\n  };",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		{
			Code:   `a?.b;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		{
			Code:   `(a?.b).c;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		{
			Code:   `a?.['b'];`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		{
			Code:   `(a?.['b']).c;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		{
			Code:   `a?.b()?.c;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		{
			Code:   `(a?.b()).c;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		{
			Code:   `one[2]?.[3][4];`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		{
			Code:   `one.two?.three.four;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},

		// ================================================================
		// TypeScript: module/namespace — directive not at prologue
		// ================================================================
		{
			Code:   "module Foo {\n  const foo = true;\n  'use strict';\n}",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 3, Column: 3}},
		},
		{
			Code:   "namespace Foo {\n  export class Foo {}\n  export class Bar {}\n\n  'use strict';\n}",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 5, Column: 3}},
		},
		{
			Code:   "function foo() {\n  const foo = true;\n\n  ('use strict');\n}",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 4, Column: 3}},
		},

		// ================================================================
		// Optional chaining with allowShortCircuit — member access still invalid
		// ================================================================
		{
			Code:    `foo && foo?.bar;`,
			Options: map[string]interface{}{"allowShortCircuit": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},

		// ================================================================
		// Ternary with optional member access
		// ================================================================
		{
			Code:    `foo ? foo?.bar : bar.baz;`,
			Options: map[string]interface{}{"allowTernary": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},

		// ================================================================
		// TypeScript: instantiation expression (unused)
		// ================================================================
		{
			Code:   "class Foo<T> {}\nFoo<string>;",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 2, Column: 1}},
		},
		{
			Code:   `Map<string, string>;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},

		// ================================================================
		// TypeScript: plain identifier
		// ================================================================
		{
			Code:   "declare const foo: number | undefined;\nfoo;",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 2, Column: 1}},
		},

		// ================================================================
		// TypeScript: assertions wrapping non-call expressions
		// ================================================================
		{
			Code:   "declare const foo: number | undefined;\nfoo as any;",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 2, Column: 1}},
		},
		{
			Code:   "declare const foo: number | undefined;\n<any>foo;",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 2, Column: 1}},
		},
		{
			Code:   "declare const foo: number | undefined;\nfoo!;",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 2, Column: 1}},
		},

		// ================================================================
		// Literals: boolean, null, bigint, regex
		// ================================================================
		{
			Code:   `true;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		{
			Code:   `false;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		{
			Code:   `null;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		{
			Code:   `1n;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		{
			Code:   `/regex/;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},

		// ================================================================
		// Unary without side effects: -, ~, typeof
		// ================================================================
		{
			Code:   `-a;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		{
			Code:   `~a;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		{
			Code:   `typeof foo;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},

		// ================================================================
		// this as standalone expression
		// ================================================================
		{
			Code:   `function foo() { this; }`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 18}},
		},

		// ================================================================
		// Nested parenthesized expressions
		// ================================================================
		{
			Code:   `((a));`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},

		// ================================================================
		// Arithmetic / comparison / bitwise binary expressions
		// ================================================================
		{
			Code:   `a + b;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		{
			Code:   `a === b;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		{
			Code:   `a & b;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},

		// ================================================================
		// TS satisfies — NOT unwrapped (matches @typescript-eslint behavior)
		// satisfies defaults to "not disallowed", so these are NOT flagged.
		// See valid cases above if needed.
		// ================================================================

		// ================================================================
		// Nested TS assertions wrapping non-call
		// ================================================================
		{
			Code:   `declare const foo: any; (foo as any)!;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 25}},
		},
		{
			Code:   `declare const foo: any; (foo as any) as number;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 25}},
		},

		// ================================================================
		// Comma expression (both sides are calls, but comma itself is disallowed)
		// ================================================================
		{
			Code:   `f(), g();`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},

		// ================================================================
		// Ternary: one branch valid, one invalid (without allowTernary, always flagged)
		// ================================================================
		{
			Code:   `a ? b() || (c = d) : e`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},

		// ================================================================
		// Class static block: multiple leading strings are all directives;
		// non-leading strings after non-directive are flagged
		// ================================================================
		{
			Code: "class C { static { const x = 1; 'foo'; 'bar'; } }",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedExpression", Line: 1, Column: 33},
				{MessageId: "unusedExpression", Line: 1, Column: 40},
			},
		},

		// ================================================================
		// allowShortCircuit: nested logical — right side must be valid
		// ================================================================
		{
			Code:    `a && (b ?? c);`,
			Options: map[string]interface{}{"allowShortCircuit": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},

		// ================================================================
		// allowTernary: both branches must be valid, one is identifier
		// ================================================================
		{
			Code:    `a ? b() : c;`,
			Options: map[string]interface{}{"allowTernary": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},

		// ================================================================
		// Element access (no side effects)
		// ================================================================
		{
			Code:   `a['b'];`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},

		// ================================================================
		// Template literal with interpolation (still no side effects)
		// ================================================================
		{
			Code:   "`hello ${world}`;",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},

		// ================================================================
		// Deep nesting invalid: TS wrappers around non-call
		// ================================================================
		// instantiation > identifier (no call)
		{
			Code:    `(Foo && Foo)<string, number>;`,
			Options: map[string]interface{}{"allowShortCircuit": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		// deeply nested ternary + short-circuit: one branch invalid
		{
			Code:    `a ? (b && c()) : (d || e)`,
			Options: map[string]interface{}{"allowShortCircuit": true, "allowTernary": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		// chained optional member access after call (result is member access)
		{
			Code:   `a?.()?.b;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		// arrow function expression as statement
		{
			Code:   `(() => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		// function expression as statement
		{
			Code:   `(function() {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		// class expression as statement
		{
			Code:   `(class {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
		// object literal as statement
		{
			Code:   `({a: 1});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedExpression", Line: 1, Column: 1}},
		},
	})
}
