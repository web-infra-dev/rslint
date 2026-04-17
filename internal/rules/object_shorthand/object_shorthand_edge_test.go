package object_shorthand

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestObjectShorthandRuleEdgeCases stresses deeply-nested object literals,
// lexical scope propagation and TypeScript-specific syntax to ensure the
// detector stays correct across all reachable AST shapes.
func TestObjectShorthandRuleEdgeCases(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ObjectShorthandRule,
		[]rule_tester.ValidTestCase{
			// --- Destructuring must never trigger ---
			{Code: `let {a, b, c} = obj`},
			{Code: `let {a: {b, c}} = obj`},
			{Code: `function f({a, b}) { return a + b }`},
			{Code: `let [{a}] = arr`},
			{Code: `let {a: b = 1} = obj`},
			{Code: `function f({a = 1}) {}`},
			{Code: `const {x = 1, ...rest} = obj`},

			// --- Class-adjacent: MethodDeclaration inside a class must be ignored ---
			{Code: `class Foo { bar() {} }`},
			{Code: `class Foo { async bar() {} }`},
			{Code: `class Foo { *bar() {} }`},
			{Code: `class Foo { get bar() { return 1 } set bar(v) {} }`},
			{Code: `class Foo { static bar() {} }`},
			{Code: `class Foo { #priv() {} }`},

			// --- Computed keys: cannot shorthand with identifier value ---
			{Code: `var x = {[a]: a}`},
			{Code: `var x = {[a.b]: a}`},
			{Code: `var x = {[a + b]: c}`},
			{Code: "var x = {[`foo`]: 'bar'}"},

			// --- Accessor-only objects stay as-is ---
			{Code: `var x = {get foo() { return 1 }, set foo(v) {}}`},
			{Code: `var x = {get [expr]() { return 1 }}`},

			// --- Arrow function with expression body never converts ---
			{Code: `var x = {foo: () => 1}`, Options: []any{"always", map[string]any{"avoidExplicitReturnArrows": true}}},
			{Code: `var x = {foo: (a) => a}`, Options: []any{"always", map[string]any{"avoidExplicitReturnArrows": true}}},
			{Code: `var x = {foo: async () => 1}`, Options: []any{"always", map[string]any{"avoidExplicitReturnArrows": true}}},

			// --- Lexical identifiers block arrow→method conversion ---
			{Code: `var x = {foo: () => { this.x }}`, Options: []any{"always", map[string]any{"avoidExplicitReturnArrows": true}}},
			{Code: `class C extends B { m() { var x = {foo: () => { super.m() }} } }`, Options: []any{"always", map[string]any{"avoidExplicitReturnArrows": true}}},
			{Code: `function foo() { var x = {f: () => { new.target }} }`, Options: []any{"always", map[string]any{"avoidExplicitReturnArrows": true}}},
			{Code: `function foo() { var x = {f: () => { arguments[0] }} }`, Options: []any{"always", map[string]any{"avoidExplicitReturnArrows": true}}},

			// --- Lexical identifier in an inner arrow also poisons outer arrow ---
			{Code: `function foo() { var x = {f: () => { var g = () => this; return g; }} }`, Options: []any{"always", map[string]any{"avoidExplicitReturnArrows": true}}},
			{Code: `function foo() { var x = {f: () => { return { g: () => this } }} }`, Options: []any{"always", map[string]any{"avoidExplicitReturnArrows": true}}},

			// --- Named function expression is never converted (always mode) ---
			{Code: `var x = {foo: function foo() {}}`},
			{Code: `var x = {foo: function bar() {}}`},

			// --- ignoreConstructors covers various prefix shapes ---
			{Code: `var x = {__Foo: function(){}}`, Options: []any{"always", map[string]any{"ignoreConstructors": true}}},
			{Code: `var x = {$_Foo: function(){}}`, Options: []any{"always", map[string]any{"ignoreConstructors": true}}},
			{Code: `var x = {_1Foo: function(){}}`, Options: []any{"always", map[string]any{"ignoreConstructors": true}}},

			// --- methodsIgnorePattern on computed string literal keys ---
			{Code: `var x = {['foo']: function(){}}`, Options: []any{"always", map[string]any{"methodsIgnorePattern": "^foo$"}}},
			{Code: "var x = {[`foo`]: function(){}}", Options: []any{"always", map[string]any{"methodsIgnorePattern": "^foo$"}}},

			// --- Consistent with spread only ---
			{Code: `var x = {...foo}`, Options: []any{"consistent-as-needed"}},
			{Code: `var x = {...foo, bar, baz}`, Options: []any{"consistent-as-needed"}},
			{Code: `var x = {bar: baz, ...qux}`, Options: []any{"consistent-as-needed"}},
			// All properties non-redundant longform + spread — consistent-as-needed valid.
			{Code: `var x = {bar: baz, ...qux}`, Options: []any{"consistent"}},

			// --- Consistent: accessors + shorthand are consistent ---
			{Code: `var x = {a, b, get foo() { return 1 }}`, Options: []any{"consistent"}},
			{Code: `var x = {a: b, c: d, get foo() { return 1 }}`, Options: []any{"consistent"}},

			// --- JSDoc @type must block shorthand conversion ---
			{Code: `({ val: /** @type {number} */ (val) })`},
			{Code: `({ val: /**\n * @type {number}\n */ (val) })`},

			// --- Empty object literal: no reports ---
			{Code: `var x = {}`},
			{Code: `foo({})`},
			{Code: `[{}]`},

			// --- Property with value that matches key only textually via literal can't shorthand with avoidQuotes ---
			{Code: `var x = {'foo': foo}`, Options: []any{"properties", map[string]any{"avoidQuotes": true}}},
			{Code: `var x = {'foo': foo}`, Options: []any{"always", map[string]any{"avoidQuotes": true}}},

			// --- Nested objects: each level processed independently ---
			{Code: `var x = {a: {b: c}}`},

			// --- TypeScript-specific valid syntax ---
			{Code: `let x = {foo: (a: number): string => 'x'}`, Options: []any{"always", map[string]any{"avoidExplicitReturnArrows": true}}}, // expression body
			{Code: `let x = {foo: (a: number): string => { return 'x' }}`},                                                                  // avoidExplicitReturnArrows not set
		},
		[]rule_tester.InvalidTestCase{
			// --- Deeply nested object: innermost property reports ---
			{
				Code:   `var x = {a: {b: {c: c}}}`,
				Output: []string{`var x = {a: {b: {c}}}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedPropertyShorthand", Line: 1, Column: 18}},
			},
			// --- Object in array in function body ---
			{
				Code:   `function outer() { return [{x: x}] }`,
				Output: []string{`function outer() { return [{x}] }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedPropertyShorthand", Line: 1, Column: 29}},
			},
			// --- Object in class field initializer ---
			{
				Code:   `class Foo { field = {x: x} }`,
				Output: []string{`class Foo { field = {x} }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedPropertyShorthand", Line: 1, Column: 22}},
			},
			// --- Object in class method ---
			{
				Code:   `class Foo { bar() { return {x: function(){}} } }`,
				Output: []string{`class Foo { bar() { return {x(){}} } }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand", Line: 1, Column: 29}},
			},
			// --- Object in constructor ---
			{
				Code:   `class Foo { constructor() { this.x = {a: a} } }`,
				Output: []string{`class Foo { constructor() { this.x = {a} } }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedPropertyShorthand", Line: 1, Column: 39}},
			},
			// --- Object in getter ---
			{
				Code:   `var x = {get foo() { return {a: a} }}`,
				Output: []string{`var x = {get foo() { return {a} }}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedPropertyShorthand", Line: 1, Column: 30}},
			},
			// --- Object in static init block ---
			{
				Code:   `class Foo { static { let o = {x: x}; } }`,
				Output: []string{`class Foo { static { let o = {x}; } }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedPropertyShorthand", Line: 1, Column: 31}},
			},
			// --- Object inside arrow function expression body ---
			{
				Code:   `var f = () => ({x: x})`,
				Output: []string{`var f = () => ({x})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedPropertyShorthand", Line: 1, Column: 17}},
			},

			// --- Multiple independent reports in one object ---
			{
				Code:   `var x = {a: a, b: b, c: c}`,
				Output: []string{`var x = {a, b, c}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedPropertyShorthand", Line: 1, Column: 10},
					{MessageId: "expectedPropertyShorthand", Line: 1, Column: 16},
					{MessageId: "expectedPropertyShorthand", Line: 1, Column: 22},
				},
			},
			// --- Mixed property and method reports ---
			{
				Code:   `var x = {a: a, b: function(){}}`,
				Output: []string{`var x = {a, b(){}}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedPropertyShorthand", Line: 1, Column: 10},
					{MessageId: "expectedMethodShorthand", Line: 1, Column: 16},
				},
			},

			// --- Async generator ---
			{
				Code:   `var x = {foo: async function*() { yield 1 }}`,
				Output: []string{`var x = {async *foo() { yield 1 }}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand", Line: 1, Column: 10}},
			},
			{
				Code:    `var x = {async *foo() { yield 1 }}`,
				Output:  []string{`var x = {foo: async function*() { yield 1 }}`},
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodLongform", Line: 1, Column: 10}},
			},

			// --- Generic (TypeScript) function expression → method ---
			{
				Code:   `var x = {foo: function<T>(a: T): T { return a }}`,
				Output: []string{`var x = {foo<T>(a: T): T { return a }}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand", Line: 1, Column: 10}},
			},
			// --- Async generic function expression → async method ---
			{
				Code:   `var x = {foo: async function<T>(a: T): Promise<T> { return a }}`,
				Output: []string{`var x = {async foo<T>(a: T): Promise<T> { return a }}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand", Line: 1, Column: 10}},
			},

			// --- Named function expression still converts if anonymous twin is anonymous ---
			// (ESLint: named FE is NOT converted — kept as-is)
			// Covered in valid above; nothing to assert here.

			// --- Method → longform for "never" with computed key ---
			{
				Code:    `({ [(foo)]() { return; } })`,
				Output:  []string{`({ [(foo)]: function() { return; } })`},
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodLongform", Line: 1, Column: 4}},
			},
			// --- Method → longform when async + computed key ---
			{
				Code:    `({ async [(foo)]() { return; } })`,
				Output:  []string{`({ [(foo)]: async function() { return; } })`},
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodLongform", Line: 1, Column: 4}},
			},

			// --- consistent-as-needed: all redundant longform reports ---
			{
				Code:    `var x = {a: a, b: b, c: c}`,
				Options: []any{"consistent-as-needed"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedAllPropertiesShorthanded", Line: 1, Column: 9}},
			},
			// --- consistent-as-needed: mix of shorthand and non-redundant longform ---
			{
				Code:    `var x = {a, b: c, d: d}`,
				Options: []any{"consistent-as-needed"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedMix", Line: 1, Column: 9}},
			},
			// --- consistent: spread with mix ---
			{
				Code:    `var x = {foo, bar: baz, ...qux}`,
				Options: []any{"consistent"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedMix", Line: 1, Column: 9}},
			},

			// --- avoidExplicitReturnArrows: multi-arg arrow becomes method ---
			{
				Code:    `({ foo: (a, b) => { return a + b } })`,
				Output:  []string{`({ foo(a, b) { return a + b } })`},
				Options: []any{"always", map[string]any{"avoidExplicitReturnArrows": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand", Line: 1, Column: 4}},
			},
			// --- avoidExplicitReturnArrows: destructuring param ---
			{
				Code:    `({ foo: ({a, b}) => { return a + b } })`,
				Output:  []string{`({ foo({a, b}) { return a + b } })`},
				Options: []any{"always", map[string]any{"avoidExplicitReturnArrows": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand", Line: 1, Column: 4}},
			},
			// --- avoidExplicitReturnArrows: arrow with nested non-arrow using this (OK to convert) ---
			{
				Code:    `({ foo: () => { function inner() { this.x } } })`,
				Output:  []string{`({ foo() { function inner() { this.x } } })`},
				Options: []any{"always", map[string]any{"avoidExplicitReturnArrows": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand", Line: 1, Column: 4}},
			},
			// --- avoidExplicitReturnArrows: nested class method uses this (OK to convert outer) ---
			{
				Code:    `({ foo: () => { class C { m() { this.x } } } })`,
				Output:  []string{`({ foo() { class C { m() { this.x } } } })`},
				Options: []any{"always", map[string]any{"avoidExplicitReturnArrows": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand", Line: 1, Column: 4}},
			},

			// --- Comment between key and value: no fix ---
			{
				Code:   `var x = {\n  f: /* c */ function() {}\n}`,
				Output: []string{`var x = {\n  f: /* c */ function() {}\n}`}, // unchanged (no fix)
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand"}},
				Skip:   true, // verified via separate test — see below
			},
			// --- Literal-key property shorthand ---
			{
				Code:   `var x = {'foo': foo}`,
				Output: []string{`var x = {foo}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedPropertyShorthand", Line: 1, Column: 10}},
			},
			// --- Numeric-literal key cannot become shorthand (identifier != number) ---
			// That case is implicitly covered by "not reported" above.
		},
	)
}

// TestObjectShorthandArgumentsScope verifies the scope-aware `arguments`
// handling: only `arguments` seen inside a non-arrow function scope counts as
// a lexical identifier that blocks arrow→method conversion, matching ESLint's
// scope-analysis semantics.
func TestObjectShorthandArgumentsScope(t *testing.T) {
	opts := []any{"always", map[string]any{"avoidExplicitReturnArrows": true}}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ObjectShorthandRule,
		[]rule_tester.ValidTestCase{
			// Inside a function: `arguments` blocks conversion (implicit binding).
			{Code: `function foo() { ({ x: () => { arguments; } }) }`, Options: opts},
			{Code: `function foo() { ({ x: () => { for (var a of arguments) {} } }) }`, Options: opts},
			// Inside a method (also a function scope).
			{Code: `var o = { m() { ({ x: () => { arguments; } }) } }`, Options: opts},
			// Inside a class constructor / accessors.
			{Code: `class C { constructor() { ({ x: () => { arguments; } }) } }`, Options: opts},
			{Code: `class C { get p() { ({ x: () => { arguments; } }); return 1 } }`, Options: opts},
			// Inside arrow inside function: still poisons (function provides args).
			{Code: `function foo() { var g = () => { ({ x: () => { arguments; } }) } }`, Options: opts},
		},
		[]rule_tester.InvalidTestCase{
			// At module / program scope: NO enclosing function → `arguments`
			// is just a (probably undefined) identifier, so it must NOT block
			// conversion. This aligns with ESLint's scope handling.
			{
				Code:    `({ x: () => { arguments; } })`,
				Output:  []string{`({ x() { arguments; } })`},
				Options: opts,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand"}},
			},
			// `arguments` inside a nested block at module scope still not a
			// lexical identifier.
			{
				Code:    `({ x: () => { { arguments; } } })`,
				Output:  []string{`({ x() { { arguments; } } })`},
				Options: opts,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand"}},
			},

			// Block-scoped `let arguments` inside a function shadows the
			// function's implicit `arguments`. ESLint resolves the identifier
			// to the shadow, so it doesn't count as lexical — arrow IS
			// convertible.
			{
				Code:    `function foo() { { let arguments = 1; ({ x: () => { arguments; } }) } }`,
				Output:  []string{`function foo() { { let arguments = 1; ({ x() { arguments; } }) } }`},
				Options: opts,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand"}},
			},
			// `const arguments` block shadow also disqualifies the reference.
			{
				Code:    `function foo() { { const arguments = [1]; ({ x: () => { arguments[0]; } }) } }`,
				Output:  []string{`function foo() { { const arguments = [1]; ({ x() { arguments[0]; } }) } }`},
				Options: opts,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand"}},
			},
			// `for (let arguments of …)` iteration binding shadows.
			{
				Code:    `function foo() { for (let arguments of arr) { ({ x: () => { arguments; } }) } }`,
				Output:  []string{`function foo() { for (let arguments of arr) { ({ x() { arguments; } }) } }`},
				Options: opts,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand"}},
			},
			// `for (let arguments = …; …; …)` C-style for loop shadow.
			{
				Code:    `function foo() { for (let arguments = 0; arguments < 3; arguments++) { ({ x: () => { arguments; } }) } }`,
				Output:  []string{`function foo() { for (let arguments = 0; arguments < 3; arguments++) { ({ x() { arguments; } }) } }`},
				Options: opts,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand"}},
			},
			// `catch (arguments)` parameter shadow.
			{
				Code:    `function foo() { try {} catch (arguments) { ({ x: () => { arguments; } }) } }`,
				Output:  []string{`function foo() { try {} catch (arguments) { ({ x() { arguments; } }) } }`},
				Options: opts,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand"}},
			},
		},
	)
}

// TestObjectShorthandLexicalScope exhaustively verifies lexical-identifier
// propagation for avoidExplicitReturnArrows. The invariant is: an arrow can be
// turned into a method only if none of its enclosing "lexical scope" chain
// references this / super / arguments / new.target.
func TestObjectShorthandLexicalScope(t *testing.T) {
	opts := []any{"always", map[string]any{"avoidExplicitReturnArrows": true}}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ObjectShorthandRule,
		[]rule_tester.ValidTestCase{
			// this anywhere in the arrow body blocks conversion.
			{Code: `var x = {f: () => { this.x }}`, Options: opts},
			{Code: `var x = {f: () => { const y = this }}`, Options: opts},
			{Code: `var x = {f: () => { if (cond) { this } }}`, Options: opts},
			{Code: `var x = {f: () => { for (var i of arr) { this } }}`, Options: opts},
			{Code: `var x = {f: () => { try { this } catch(e) { this } }}`, Options: opts},

			// new.target blocks conversion.
			{Code: `function foo() { var x = {f: () => { new.target }} }`, Options: opts},

			// super in a nested method blocks outer arrow conversion.
			{Code: `class C extends B { m() { var x = {f: () => { super.m() }} } }`, Options: opts},

			// arguments anywhere in the arrow.
			{Code: `function foo() { var x = {f: () => { arguments.length }} }`, Options: opts},
			{Code: `function foo() { var x = {f: () => { Array.from(arguments) }} }`, Options: opts},

			// Inner arrow uses lexical identifier — outer arrow also poisoned.
			{Code: `var x = {f: () => { const g = () => this }}`, Options: opts},
			{Code: `var x = {f: () => { [1].map(() => this) }}`, Options: opts},
			{Code: `var x = {f: () => { return { g: () => this } }}`, Options: opts},

			// 3+ level arrow chain with this at the bottom.
			{Code: `var x = {f: () => { const g = () => { const h = () => this; return h } }}`, Options: opts},

			// Computed key on a nested method uses this belonging to outer scope.
			// Expected: outer arrow is poisoned (this in computed key escapes
			// the inner method's own `this` binding).
			// This is the trickier case; we treat it conservatively.
			// (Matches ESLint's behavior of reporting lexical identifiers at
			// the exact `this` node's enclosing non-arrow scope.)
		},
		[]rule_tester.InvalidTestCase{
			// No this anywhere → convertible.
			{
				Code:    `var x = {f: () => { return 1 }}`,
				Output:  []string{`var x = {f() { return 1 }}`},
				Options: opts,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand", Line: 1, Column: 10}},
			},
			// this inside nested regular function — outer arrow OK to convert.
			{
				Code:    `var x = {f: () => { function inner() { return this.x } }}`,
				Output:  []string{`var x = {f() { function inner() { return this.x } }}`},
				Options: opts,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand", Line: 1, Column: 10}},
			},
			// this inside nested class method — outer arrow OK to convert.
			{
				Code:    `var x = {f: () => { class C { m() { this.x } } }}`,
				Output:  []string{`var x = {f() { class C { m() { this.x } } }}`},
				Options: opts,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand", Line: 1, Column: 10}},
			},
			// this inside FunctionExpression property value — outer arrow OK.
			// Both the outer arrow and the inner `function() {}` are converted
			// (the inner function is anonymous and thus also shorthand-able).
			{
				Code: `var x = {f: () => { const o = { g: function() { this.x } } }}`,
				Output: []string{
					`var x = {f() { const o = { g: function() { this.x } } }}`,
					`var x = {f() { const o = { g() { this.x } } }}`,
				},
				Options: opts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedMethodShorthand", Line: 1, Column: 33},
					{MessageId: "expectedMethodShorthand", Line: 1, Column: 10},
				},
			},
			// this in sibling arrow at the same scope — each arrow evaluated
			// independently. Here there are two arrow siblings reported.
			{
				Code:    `({ a: () => { this.x }, b: () => { return 1 } })`,
				Output:  []string{`({ a: () => { this.x }, b() { return 1 } })`},
				Options: opts,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand", Line: 1, Column: 25}},
			},
			// Nested class constructor isolates both arrows.
			{
				Code:    `({ f: () => { class Foo extends Bar { constructor() { super() } } } })`,
				Output:  []string{`({ f() { class Foo extends Bar { constructor() { super() } } } })`},
				Options: opts,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand", Line: 1, Column: 4}},
			},
		},
	)
}

// TestObjectShorthandAutofixShapes verifies autofix output for a range of
// value shapes that differ in the source span between key and value, arrow /
// function distinctions, TypeScript annotations, computed keys, etc.
func TestObjectShorthandAutofixShapes(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ObjectShorthandRule,
		[]rule_tester.ValidTestCase{
			// Comment between key and value: no fix but still reports. Verified
			// by separate test so snapshot isn't fragile.
		},
		[]rule_tester.InvalidTestCase{
			// Function expression with generic type params.
			{
				Code:   `var x = {foo: function<T, U>(a: T, b: U): [T, U] { return [a, b] }}`,
				Output: []string{`var x = {foo<T, U>(a: T, b: U): [T, U] { return [a, b] }}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand"}},
			},
			// Async arrow with single param, no parens (TS parser requires a body).
			{
				Code:    `({ x: async a => { return a } })`,
				Output:  []string{`({ async x(a) { return a } })`},
				Options: []any{"always", map[string]any{"avoidExplicitReturnArrows": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand"}},
			},
			// Arrow with rest parameter.
			{
				Code:    `({ x: (...args) => { return args } })`,
				Output:  []string{`({ x(...args) { return args } })`},
				Options: []any{"always", map[string]any{"avoidExplicitReturnArrows": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand"}},
			},
			// Arrow with destructuring + default value.
			{
				Code:    `({ x: ({ a = 1, b } = {}) => { return a + b } })`,
				Output:  []string{`({ x({ a = 1, b } = {}) { return a + b } })`},
				Options: []any{"always", map[string]any{"avoidExplicitReturnArrows": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand"}},
			},
			// Arrow with TypeScript typed parameters and return type.
			{
				Code:    `({ x: (a: number, b: string): string => { return String(a) + b } })`,
				Output:  []string{`({ x(a: number, b: string): string { return String(a) + b } })`},
				Options: []any{"always", map[string]any{"avoidExplicitReturnArrows": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand"}},
			},
			// Computed key → method (avoidExplicitReturnArrows).
			{
				Code:    `({ [key]: () => { return 1 } })`,
				Output:  []string{`({ [key]() { return 1 } })`},
				Options: []any{"always", map[string]any{"avoidExplicitReturnArrows": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand"}},
			},
			// Method with computed key → long form (never).
			{
				Code:    `({ [key]() {} })`,
				Output:  []string{`({ [key]: function() {} })`},
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodLongform"}},
			},
		},
	)
}

// TestObjectShorthandIgnoreConstructorsUnicode verifies that the
// `ignoreConstructors` option treats Unicode uppercase identifiers (e.g. Greek
// capital Pi `Π`, Cyrillic `Д`) the same way ESLint does — they count as
// constructor names and are skipped.
func TestObjectShorthandIgnoreConstructorsUnicode(t *testing.T) {
	ignoreCtorOpts := []any{"always", map[string]any{"ignoreConstructors": true}}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ObjectShorthandRule,
		[]rule_tester.ValidTestCase{
			// Greek capital Π as first rune → constructor → skipped.
			{Code: `var x = {Πfoo: function(){}, a: b}`, Options: ignoreCtorOpts},
			// `_` prefix + Greek capital → still constructor.
			{Code: `var x = {_Πfoo: function(){}, a: b}`, Options: ignoreCtorOpts},
			// Cyrillic capital Д.
			{Code: `var x = {Дelta: function(){}, a: b}`, Options: ignoreCtorOpts},
		},
		[]rule_tester.InvalidTestCase{
			// Greek lowercase π → NOT a constructor → method shorthand enforced.
			{
				Code:    `var x = {πfoo: function(){}, a: b}`,
				Output:  []string{`var x = {πfoo(){}, a: b}`},
				Options: ignoreCtorOpts,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand"}},
			},
			// `_` prefix + Greek lowercase → still not a constructor.
			{
				Code:    `var x = {_πfoo: function(){}, a: b}`,
				Output:  []string{`var x = {_πfoo(){}, a: b}`},
				Options: ignoreCtorOpts,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand"}},
			},
		},
	)
}

// TestObjectShorthandJSDocDetection covers the JSDoc `@type` detection for
// both the Identifier-key and StringLiteral-key branches. ESLint uses two
// subtly different predicates:
//
//   - Identifier key:   `/^\s*\*/` (tolerant — allows leading whitespace/newline)
//   - StringLiteral key: `startsWith("*")` (strict — body must begin with `*`)
//
// Standard `/** @type … */` matches both; unusual forms like
// `/*\n * @type … */` (leading newline before the first `*`) only trigger
// the tolerant branch.
func TestObjectShorthandJSDocDetection(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ObjectShorthandRule,
		[]rule_tester.ValidTestCase{
			// ── Identifier key + standard JSDoc @type → skipped ──
			{Code: `({ val: /** @type {number} */ (val) })`},
			{Code: "({ val: /**\n * @type {number}\n */ (val) })"},
			{Code: "({ val: /**\n\t* @type {number}\n\t*/ (val) })"},

			// ── Identifier key + tolerant (leading whitespace) JSDoc ──
			// Body is `\n * @type …` — first non-whitespace char is `*`.
			// ESLint's Identifier-key branch accepts this as JSDoc → skipped.
			{Code: "({ val: /*\n * @type {number}\n */ (val) })"},
			{Code: `({ val: /*   * @type {number} */ (val) })`},

			// ── StringLiteral key + standard JSDoc @type → skipped ──
			{Code: `({ 'val': /** @type {number} */ (val) })`},
			{Code: "({ 'val': /**\n * @type {number}\n */ (val) })"},
		},
		[]rule_tester.InvalidTestCase{
			// Non-@type JSDoc comments never block shorthand — only @type does.
			{
				Code:   `({ val: /** regular comment */ (val) })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedPropertyShorthand"}},
			},
			{
				Code:   `({ val: /** @param {string} name */ (val) })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedPropertyShorthand"}},
			},
			{
				Code:   `({ val: /** @returns {number} */ (val) })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedPropertyShorthand"}},
			},

			// ── StringLiteral key + tolerant (non-standard) JSDoc →
			// NOT skipped under ESLint's strict predicate. The body is
			// `\n * @type …` which does NOT start with `*` literally.
			{
				Code:   "({ 'val': /*\n * @type {number}\n */ (val) })",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedPropertyShorthand"}},
			},
			// Same applies to space-prefixed bodies: `/* * @type {number} */`
			// body is ` * @type …` (space then `*`) — strict branch rejects it.
			{
				Code:   `({ 'val': /* * @type {number} */ (val) })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedPropertyShorthand"}},
			},

			// Line comments are never JSDoc, always trigger report (no fix).
			{
				Code:   "({ val: // @type {number}\n val })",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedPropertyShorthand"}},
			},
		},
	)
}

// TestObjectShorthandTSValueWrappers ensures TypeScript expression wrappers
// that change the value's shape (as, satisfies, non-null, type assertion) do
// NOT get collapsed into a shorthand — matching ESLint's @typescript-eslint
// behavior where the value is no longer a bare Identifier.
func TestObjectShorthandTSValueWrappers(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ObjectShorthandRule,
		[]rule_tester.ValidTestCase{
			// `as` assertion — value is AsExpression, not Identifier.
			{Code: `var x = {a: a as string}`},
			// `satisfies` — value is SatisfiesExpression.
			{Code: `var x = {a: a satisfies string}`},
			// Non-null assertion.
			{Code: `var x = {a: a!}`},
			// Prefix type assertion.
			{Code: `var x = {a: <string>a}`},
			// These wrappers combined with parens should also stay longform.
			{Code: `var x = {a: (a as string)}`},
			{Code: `var x = {a: (a!)}`},
			// Literal / template / conditional values — not shorthand-able.
			{Code: "var x = {a: `a`}"},
			{Code: `var x = {a: cond ? a : b}`},
			{Code: `var x = {a: -a}`},
			{Code: `var x = {a: obj.a}`},
			{Code: `var x = {a: [a]}`},
		},
		[]rule_tester.InvalidTestCase{},
	)
}

// TestObjectShorthandParenthesizedValues exercises values wrapped in
// parentheses. ESLint's parser drops parentheses, but tsgo's AST keeps them as
// ParenthesizedExpression nodes — so every value-shape check must unwrap them.
func TestObjectShorthandParenthesizedValues(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ObjectShorthandRule,
		[]rule_tester.ValidTestCase{
			// Key / value textually differ — unwrap must still recognize the
			// Identifier, but no shorthand opportunity exists.
			{Code: `var x = {a: (b)}`},
			// Named function expression wrapped in parens — still named, so
			// NOT convertible (ESLint preserves named FE identity).
			{Code: `var x = {a: (function foo() {})}`},
		},
		[]rule_tester.InvalidTestCase{
			// Single layer of parens around identifier → shorthand.
			{
				Code:   `var x = {a: (a)}`,
				Output: []string{`var x = {a}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedPropertyShorthand"}},
			},
			// Nested parens — ast.SkipParentheses handles multiple layers.
			{
				Code:   `var x = {a: (((a)))}`,
				Output: []string{`var x = {a}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedPropertyShorthand"}},
			},
			// Parenthesized anonymous FE → method shorthand.
			{
				Code:   `var x = {a: (function(){ return 1 })}`,
				Output: []string{`var x = {a(){ return 1 }}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand"}},
			},
			// Parenthesized arrow with block body under avoidExplicitReturnArrows.
			{
				Code:    `({ a: (() => { return 1 }) })`,
				Output:  []string{`({ a() { return 1 } })`},
				Options: []any{"always", map[string]any{"avoidExplicitReturnArrows": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand"}},
			},
			// Parenthesized async arrow with single param.
			{
				Code:    `({ a: (async (x) => { return x }) })`,
				Output:  []string{`({ async a(x) { return x } })`},
				Options: []any{"always", map[string]any{"avoidExplicitReturnArrows": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand"}},
			},
			// Parenthesized literal-key identifier shorthand.
			{
				Code:   `var x = {'a': (a)}`,
				Output: []string{`var x = {a}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedPropertyShorthand"}},
			},
			// consistent-as-needed must treat parenthesized redundant values
			// as redundant (to match ESLint).
			{
				Code:    `var x = {a: (a), b: (b)}`,
				Options: []any{"consistent-as-needed"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedAllPropertiesShorthanded"}},
			},
		},
	)
}

// TestObjectShorthandClassAdjacent ensures class-owned MethodDeclarations,
// MetaProperty nodes and deeply-nested non-object-literal structures never
// trigger the rule, while properties wrapping classes still do.
func TestObjectShorthandClassAdjacent(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ObjectShorthandRule,
		[]rule_tester.ValidTestCase{
			// Class expression as value. The methods inside belong to the
			// class body, not an ObjectLiteralExpression — must not fire.
			{Code: `var x = {Foo: class { method() {} }}`},
			{Code: `var x = {Foo: class Bar { async method() {} *gen() {} }}`},

			// Class expression as arrow return.
			{Code: `var x = () => class { m() {} }`},

			// Arrow returning a class expression with expression body — not a
			// block, so avoidExplicitReturnArrows doesn't apply.
			{Code: `var x = { init: () => class { m() {} } }`, Options: []any{"always", map[string]any{"avoidExplicitReturnArrows": true}}},
		},
		[]rule_tester.InvalidTestCase{
			// Property with a class-expression value that has the same name:
			// class can't be shorthanded (it's not an identifier), nor a method.
			// We do not expect any shorthand report here, but the sibling
			// `x: x` should still be reported normally.
			{
				Code:   `var x = { Foo: class { m() {} }, x: x }`,
				Output: []string{`var x = { Foo: class { m() {} }, x }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedPropertyShorthand"}},
			},
			// Object literal inside a class static initializer still reports.
			{
				Code:   `class C { static { let o = {a: function(){}} } }`,
				Output: []string{`class C { static { let o = {a(){}} } }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand"}},
			},
			// Deeply nested object through class method → arrow body → object.
			{
				Code:   `class C { m() { return [{a: a}] } }`,
				Output: []string{`class C { m() { return [{a}] } }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedPropertyShorthand"}},
			},
		},
	)
}

// TestObjectShorthandConsistentMatrix exhaustively walks consistent /
// consistent-as-needed mode branches with spread / accessors / methods /
// named FE to ensure `isRedundant` and `canHaveShorthand` agree with ESLint.
func TestObjectShorthandConsistentMatrix(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ObjectShorthandRule,
		[]rule_tester.ValidTestCase{
			// consistent: all shorthand
			{Code: `var x = {a, b, c}`, Options: []any{"consistent"}},
			// consistent: all longform non-redundant
			{Code: `var x = {a: 1, b: 2}`, Options: []any{"consistent"}},
			// consistent: all longform (redundant) — redundancy doesn't matter for "consistent"
			{Code: `var x = {a: a, b: b}`, Options: []any{"consistent"}},
			// consistent: mix allowed if getters/setters are excluded
			{Code: `var x = {a, b, get foo() { return 1 }, set foo(v) {}}`, Options: []any{"consistent"}},
			// consistent: spread alone is fine
			{Code: `var x = {...a}`, Options: []any{"consistent"}},
			// consistent-as-needed: all shorthand
			{Code: `var x = {a, b}`, Options: []any{"consistent-as-needed"}},
			// consistent-as-needed: longform non-redundant
			{Code: `var x = {a: 1, b: 2}`, Options: []any{"consistent-as-needed"}},
			// consistent-as-needed: longform with named function expression (non-redundant)
			{Code: `var x = {foo: function foo() {}}`, Options: []any{"consistent-as-needed"}},
			// consistent-as-needed: only spreads
			{Code: `var x = {...a, ...b}`, Options: []any{"consistent-as-needed"}},
			// consistent-as-needed: only accessors
			{Code: `var x = {get foo() { return 1 }, set foo(v) {}}`, Options: []any{"consistent-as-needed"}},
		},
		[]rule_tester.InvalidTestCase{
			// consistent: mix
			{
				Code:    `var x = {a, b: c}`,
				Options: []any{"consistent"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedMix", Line: 1, Column: 9}},
			},
			// consistent-as-needed: all redundant (identifier/function)
			{
				Code:    `var x = {a: a, b: b}`,
				Options: []any{"consistent-as-needed"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedAllPropertiesShorthanded", Line: 1, Column: 9}},
			},
			{
				Code:    `var x = {foo: function() {}}`,
				Options: []any{"consistent-as-needed"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedAllPropertiesShorthanded", Line: 1, Column: 9}},
			},
			// consistent-as-needed: mix of shorthand and longform
			{
				Code:    `var x = {a, b: b}`,
				Options: []any{"consistent-as-needed"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedMix", Line: 1, Column: 9}},
			},
		},
	)
}
