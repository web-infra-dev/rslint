// TestCommaDangleExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Each case carries an inline comment pointing
// at the specific branch / Dimension 4 row / tsgo AST quirk it covers, so
// future refactors can't silently regress them without breaking a named
// lock-in.
package comma_dangle_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/rules/comma_dangle"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestCommaDangleExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&comma_dangle.CommaDangleRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: N/A — receiver / expression wrappers ----
			// N/A: this rule inspects token positions of *containers* (object
			// literal `{`/`}`, function param `(`/`)`, etc.), not the value
			// shape of receivers or inputs. Paren / non-null / as-cast /
			// satisfies / optional-chain wrappers around the *contents* of
			// a container don't change the close-bracket scan.
			//
			// ---- Dimension 4: N/A — access / key forms ----
			// N/A: the rule never inspects member-access shapes or computed-
			// vs-static keys. Dotted vs bracket vs computed access form is
			// invisible to it.
			//
			// ---- Dimension 4: array holes (KindOmittedExpression) ----
			// tsgo represents ESTree's `null` hole as a KindOmittedExpression
			// node. Upstream's `last(nodes)` short-circuits when the last
			// element is null/undefined; we mirror that by skipping when the
			// last node is OmittedExpression. Without this branch, `[a,,]` and
			// `[,,]` would incorrectly report or insert a trailing comma.
			{Code: `[a,,]`, Options: optStr("never")},
			{Code: `[a,,]`, Options: optStr("always")},
			{Code: `[a,b,,]`, Options: optStr("never")},
			{Code: `[,a,,]`, Options: optStr("always")},

			// ---- Dimension 4: parenthesized last item ----
			// tsgo preserves `(expr)` as KindParenthesizedExpression. We don't
			// unwrap before reading list.End() because the list's End() already
			// accounts for the paren's source range. Lock in multi-level parens.
			{Code: `[((1)),]`, Options: optStr("always")},
			{Code: `var x = { foo: ((1)),};`, Options: optStr("always")},
			{Code: `var x = [((1))]`, Options: optStr("never")},

			// ---- Dimension 4: spread in literal (NOT a rest binding) ----
			// SpreadElement in ArrayLiteralExpression / SpreadAssignment in
			// ObjectLiteralExpression are NOT rest bindings in upstream's model.
			// Trailing comma after them is allowed (and forced by `'always'`).
			{Code: `var a = [b, ...spread,];`, Options: optStr("always")},
			{Code: `var a = {...spread,};`, Options: optStr("always")},
			{Code: `var a = {a: 1, ...spread,};`, Options: optStr("always")},

			// ---- Dimension 4: spread in destructuring LHS (IS rest) ----
			// `[a, ...rest] = []` parses in tsgo as ArrayLiteralExpression with
			// SpreadElement (same node kind as the literal-context case). We use
			// `IsArrayLiteralOrObjectLiteralDestructuringPattern` to detect the
			// LHS context and switch SpreadElement to "rest" semantics — JS
			// forbids a trailing comma after a rest binding in destructuring.
			{Code: `[a, ...rest] = [];`, Options: optStr("always")},
			{Code: `({a, ...rest} = b);`, Options: optStr("always")},
			{Code: `for ([a, ...rest] of []);`, Options: optStr("always")},

			// ---- Dimension 4: class methods / constructors / accessors ----
			// tsgo's MethodDeclaration / Constructor / Get|SetAccessor are
			// distinct kinds; the `functions` slot must fire on every one to
			// match upstream's "all class-method param lists" coverage (where
			// upstream sees a FunctionExpression on every MethodDefinition).
			{Code: `class A { foo(a,) {} }`, Options: optStr("always")},
			{Code: `class A { foo(a) {} }`, Options: optStr("never")},
			{Code: `class A { constructor(a,) {} }`, Options: optStr("always")},
			{Code: `class A { set x(v,) {} }`, Options: optStr("always")},
			{Code: `class A { get x() { return 1; } }`, Options: optStr("always")},

			// ---- Dimension 4: nesting boundary ----
			// `class { foo() { class { bar() {} } } }` — inner class methods
			// must be visited by the inner-class listener, not bleed up to the
			// outer. Verifies one diagnostic per param list, not per ancestor.
			{Code: `class A { foo(a,) { class B { bar(c,) {} } } }`, Options: optStr("always")},
			{Code: `class A { foo(a) { class B { bar(c) {} } } }`, Options: optStr("never")},

			// ---- Branch lock-in: dynamic-import slot vs functions slot ----
			// `import(source)` is KindCallExpression with Expression.Kind ==
			// KindImportKeyword. We route its Arguments to `dynamicImports`,
			// not `functions`. Lock in the slot independence.
			{Code: `import(source)`, Options: optMap(map[string]any{"functions": "always", "dynamicImports": "never"})},
			{Code: `import(source,)`, Options: optMap(map[string]any{"functions": "never", "dynamicImports": "always"})},

			// ---- Branch lock-in: type arguments hard-coded to 'never' ----
			// Upstream uses `predicate.never` directly for TypeArgumentList,
			// ignoring the user's `generics` setting. Verify that even with
			// `generics: 'always'`, type-argument lists are still 'never'.
			{Code: `Bar<T>`, Options: optMap(map[string]any{"generics": "always"})},
			{Code: `foo<T>()`, Options: optMap(map[string]any{"generics": "always"})},
			{Code: `new Foo<T>()`, Options: optMap(map[string]any{"generics": "always"})},

			// ---- Branch lock-in: enum members ----
			// EnumDeclaration is a tsgo-only kind for the `enum` slot.
			{Code: `enum E { A = 1, B = 2 }`, Options: optStr("never")},
			{Code: `enum E { A = 1, B = 2, }`, Options: optStr("always")},

			// ---- Branch lock-in: tuple type ----
			// TupleType is a tsgo-only kind for the `tuples` slot.
			{Code: `type T = [string, number]`, Options: optStr("never")},
			{Code: `type T = [string, number,]`, Options: optStr("always")},

			// ---- Branch lock-in: function type ----
			// FunctionType is a tsgo-only kind for the `functions` slot.
			{Code: `type Fn = (a: number, b: number,) => void`, Options: optStr("always")},
			{Code: `type Fn = (a: number, b: number) => void`, Options: optStr("never")},

			// ---- Branch lock-in: interface generics ----
			// InterfaceDeclaration carries TypeParameters but no Parameters; the
			// `generics` slot fires while `functions` does not.
			{Code: `interface Foo<T,> { bar(): T }`, Options: optMap(map[string]any{"generics": "always"})},

			// ---- Branch lock-in: type alias generics ----
			{Code: `type Foo<T,> = T`, Options: optMap(map[string]any{"generics": "always"})},

			// ---- Branch lock-in: ImportAttributes on ExportAllDeclaration ----
			// `export * from 'foo' with {...}` has no ExportClause, only the
			// Attributes branch. Verifies the export-clause-nil arm.
			{Code: `export * from "foo" with {a: "b", c: "d"}`, Options: optMap(map[string]any{"importAttributes": "never"})},

			// ---- Branch lock-in: `ignore` slot ----
			// `ignore` should suppress all reports for that slot regardless of
			// trailing-comma state. Lock the option-pass-through.
			{Code: `var foo = {a: 1,}`, Options: optMap(map[string]any{"objects": "ignore"})},
			{Code: `var foo = {a: 1}`, Options: optMap(map[string]any{"objects": "ignore"})},

			// ---- Branch lock-in: empty container — no diagnostic ----
			// Empty list → no last item → no check. Lock in the early-exit.
			{Code: `var a = []`, Options: optStr("always")},
			{Code: `var a = {}`, Options: optStr("always")},
			{Code: `foo()`, Options: optStr("always")},
			{Code: `enum E {}`, Options: optStr("always")},

			// ---- Real-user: deeply nested literal (issue-tracker shape) ----
			// Production codebases produce deeply nested object/array literals;
			// nested-listener correctness is the most common drift surface.
			{Code: "var x = {\n  a: {\n    b: {\n      c: 1,\n    },\n  },\n};", Options: optStr("always-multiline")},

			// ---- Branch lock-in: JSX type-arguments (tsgo-specific carriers) ----
			// Before delegating to `node.TypeArgumentList()`, the listener fan-out
			// hand-rolled per-kind `As<X>().TypeArguments` paths and silently
			// dropped JsxOpeningElement / JsxSelfClosingElement. Lock the fix in.
			{Code: `const x = <Foo<T> />`, Tsx: true},
			{Code: `const x = <Foo<T>>x</Foo>`, Tsx: true},
			// Single-T TypeArguments has no TSX carve-out (carve-out is for
			// TypeParameter*Declaration*, not Instantiation), so `<Foo<T,>>` is
			// always unexpected — see invalid cases below.

			// ---- Branch lock-in: TaggedTemplateExpression type-arguments ----
			// Previously registered but had zero tests.
			{Code: "tag<T>`hello`"},

			// ---- Branch lock-in: TypeQuery type-arguments ----
			// `typeof Foo<T>`. Previously registered but had zero tests.
			{Code: `type X = typeof Foo<T>`},

			// ---- Branch lock-in: ImportType type-arguments ----
			// `import('foo').Bar<T>`. Previously registered but had zero tests.
			{Code: `type X = import("foo").Bar<T>`},

			// ---- Branch lock-in: ConstructorType NOT visited ----
			// `new (a: number) => Foo` is KindConstructorType. Upstream has no
			// `TSConstructorType` listener (only `TSFunctionType` for `(a) => T`),
			// so we mirror that and skip ConstructorType. All four combinations
			// below would emit a diagnostic if ConstructorType were listed —
			// they're locked in here as valid.
			{Code: `type C = new (a: number, b: number) => Foo`, Options: optStr("never")},
			{Code: `type C = new (a: number, b: number,) => Foo`, Options: optStr("always")},
			{Code: `type C = new (a: number, b: number,) => Foo`, Options: optStr("never")},
			{Code: `type C = new (a: number, b: number) => Foo`, Options: optStr("always")},

			// ---- Branch lock-in: async / generator / async-generator variants ----
			// All flow through FunctionLikeData → checkFunctionLike. Lock in
			// that none of the FunctionExpression / ArrowFunction sub-shapes
			// flip the rest-degrade or break the listener dispatch.
			{Code: `async function foo(a) {}`, Options: optMap(map[string]any{"functions": "never"})},
			{Code: `async function foo(a,) {}`, Options: optMap(map[string]any{"functions": "always"})},
			{Code: `function* gen(a) {}`, Options: optMap(map[string]any{"functions": "never"})},
			{Code: `function* gen(a,) {}`, Options: optMap(map[string]any{"functions": "always"})},
			{Code: `async function* g(a) {}`, Options: optMap(map[string]any{"functions": "never"})},
			{Code: `async function* g(a,) {}`, Options: optMap(map[string]any{"functions": "always"})},
			{Code: `const f = async (a) => 1;`, Options: optMap(map[string]any{"functions": "never"})},
			{Code: `const f = async (a,) => 1;`, Options: optMap(map[string]any{"functions": "always"})},
			{Code: `class C { async foo(a,) {} *gen(a,) {} async *both(a,) {} }`, Options: optMap(map[string]any{"functions": "always"})},

			// ---- Branch lock-in: body-less class members are NOT checked ----
			// `abstract method`, `declare class { method() }`, and overload
			// signatures all map to ESTree's `TSEmptyBodyFunctionExpression`,
			// which upstream's `FunctionExpression` listener does not match.
			// We mirror that by skipping body-less MethodDeclaration /
			// Constructor / Get|SetAccessor.
			//
			// All four cases below would emit a missing-comma diagnostic if the
			// body-less guard were absent. Lock them in as VALID.
			{Code: `abstract class A { abstract foo(a): void; }`, Options: optMap(map[string]any{"functions": "always"})},
			{Code: `declare class A { foo(a): void; }`, Options: optMap(map[string]any{"functions": "always"})},
			// (overload variant moved to invalid below: only the body-having
			// impl row reports; both overload-signature rows are skipped.)
			{Code: `abstract class A { abstract get x(): number; abstract set x(v): void; }`, Options: optMap(map[string]any{"functions": "always"})},
			{Code: `declare class A { constructor(a); }`, Options: optMap(map[string]any{"functions": "always"})},
			// Lock-in: trailing comma in abstract / declare methods is also NOT
			// flagged under 'never' (the body-less guard is unconditional).
			{Code: `abstract class A { abstract foo(a,): void; }`, Options: optStr("never")},
			{Code: `declare class A { foo(a,): void; }`, Options: optStr("never")},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized item with `never` ----
			{
				Code:    `[((1)),]`,
				Output:  []string{`[((1))]`},
				Options: optStr("never"),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 7}},
			},

			// ---- Dimension 4: spread in destructuring with `'always'` AND
			// existing trailing comma → forbid via degrade. ----
			{
				Code:    `[a, ...rest,] = [];`,
				Output:  []string{`[a, ...rest] = [];`},
				Options: optStr("always"),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 12}},
			},
			{
				Code:    `({a, ...rest,} = b);`,
				Output:  []string{`({a, ...rest} = b);`},
				Options: optStr("always"),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 13}},
			},

			// ---- Dimension 4: spread in array literal (NOT destructuring) is
			// not a rest binding — `'always'` forces a trailing comma. ----
			{
				Code:    `var a = [b, ...spread]`,
				Output:  []string{`var a = [b, ...spread,]`},
				Options: optStr("always"),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 22}},
			},

			// ---- Branch lock-in: class method param with `'never'` ----
			// MethodDeclaration listener must fire on class methods, not just
			// FunctionExpression-as-MethodDefinition.value (which is upstream's
			// shape). Lock the listener in.
			{
				Code:    `class A { foo(a,) {} }`,
				Output:  []string{`class A { foo(a) {} }`},
				Options: optStr("never"),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 16}},
			},
			{
				Code:    `class A { constructor(a,) {} }`,
				Output:  []string{`class A { constructor(a) {} }`},
				Options: optStr("never"),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 24}},
			},
			{
				Code:    `class A { set x(v,) {} }`,
				Output:  []string{`class A { set x(v) {} }`},
				Options: optStr("never"),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 18}},
			},

			// ---- Branch lock-in: dynamic-import slot independence ----
			{
				Code:    `import(source)`,
				Output:  []string{`import(source,)`},
				Options: optMap(map[string]any{"functions": "never", "dynamicImports": "always"}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 14}},
			},
			{
				Code:    `import(source,)`,
				Output:  []string{`import(source)`},
				Options: optMap(map[string]any{"functions": "always", "dynamicImports": "never"}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 14}},
			},

			// ---- Branch lock-in: type arguments are always 'never', regardless of `generics` ----
			{
				Code:    `Bar<T,>`,
				Output:  []string{`Bar<T>`},
				Options: optMap(map[string]any{"generics": "always"}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 6}},
			},
			{
				Code:    `foo<T,>()`,
				Output:  []string{`foo<T>()`},
				Options: optMap(map[string]any{"generics": "always"}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 6}},
			},
			{
				Code:    `new Foo<T,>()`,
				Output:  []string{`new Foo<T>()`},
				Options: optMap(map[string]any{"generics": "always"}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 10}},
			},

			// ---- Branch lock-in: enum members ----
			{
				Code:    `enum E { A = 1, B = 2 }`,
				Output:  []string{`enum E { A = 1, B = 2, }`},
				Options: optStr("always"),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 22}},
			},
			{
				Code:    `enum E { A = 1, B = 2, }`,
				Output:  []string{`enum E { A = 1, B = 2 }`},
				Options: optStr("never"),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 22}},
			},

			// ---- Branch lock-in: tuple type ----
			{
				Code:    `type T = [string, number]`,
				Output:  []string{`type T = [string, number,]`},
				Options: optStr("always"),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 25}},
			},
			{
				Code:    `type T = [string, number,]`,
				Output:  []string{`type T = [string, number]`},
				Options: optStr("never"),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 25}},
			},

			// ---- Branch lock-in: function type ----
			{
				Code:    `type Fn = (a: number, b: number) => void`,
				Output:  []string{`type Fn = (a: number, b: number,) => void`},
				Options: optStr("always"),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 32}},
			},

			// ---- Branch lock-in: interface / type-alias generics ----
			{
				Code:    `interface Foo<T> {}`,
				Output:  []string{`interface Foo<T,> {}`},
				Options: optMap(map[string]any{"generics": "always"}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 16}},
			},
			{
				Code:    `type Foo<T> = T`,
				Output:  []string{`type Foo<T,> = T`},
				Options: optMap(map[string]any{"generics": "always"}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 11}},
			},

			// ---- Branch lock-in: ImportAttributes on ExportAllDeclaration ----
			{
				Code:    `export * from "foo" with {a: "b", c: "d",}`,
				Output:  []string{`export * from "foo" with {a: "b", c: "d"}`},
				Options: optMap(map[string]any{"importAttributes": "never"}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 41}},
			},

			// ---- Branch lock-in: only-multiline → forbid on single-line ----
			// `only-multiline` allows trailing comma on multi-line and forbids
			// on single-line. Lock the single-line forbid branch.
			{
				Code:    `var foo = {a: 1,}`,
				Output:  []string{`var foo = {a: 1}`},
				Options: optStr("only-multiline"),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 16}},
			},

			// ---- Real-user: deeply nested multi-line ----
			// Three nested object literals, each missing a trailing comma. The
			// emit order follows listener-fire order (outermost ObjectLiteral
			// visited first, then descending into inner ones).
			{
				Code:    "var x = {\n  a: {\n    b: {\n      c: 1\n    }\n  }\n};",
				Output:  []string{"var x = {\n  a: {\n    b: {\n      c: 1,\n    },\n  },\n};"},
				Options: optStr("always-multiline"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 6, Column: 4},
					{MessageId: "missing", Line: 5, Column: 6},
					{MessageId: "missing", Line: 4, Column: 11},
				},
			},

			// ---- Branch lock-in: JSX TypeArguments (Bug fix from earlier
			// per-kind dispatch — JsxOpeningElement / JsxSelfClosingElement
			// were dropped). Type arguments are hard-coded 'never' regardless
			// of `generics` setting. ----
			{
				Code:   `const x = <Foo<T,> />`,
				Output: []string{`const x = <Foo<T> />`},
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 17}},
			},
			{
				Code:   `const x = <Foo<T,>>x</Foo>`,
				Output: []string{`const x = <Foo<T>>x</Foo>`},
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 17}},
			},
			{
				// Multi-arg JSX: no TSX `<T,>` carve-out applies (carve-out is
				// per-list-len, and applies only to TypeParameter*Declaration*,
				// not Instantiation).
				Code:   `const x = <Foo<T,U,> />`,
				Output: []string{`const x = <Foo<T,U> />`},
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 19}},
			},

			// ---- Branch lock-in: TaggedTemplate type-arguments ----
			{
				Code:   "tag<T,>`hello`",
				Output: []string{"tag<T>`hello`"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 6}},
			},

			// ---- Branch lock-in: TypeQuery type-arguments ----
			{
				Code:   `type X = typeof Foo<T,>`,
				Output: []string{`type X = typeof Foo<T>`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 22}},
			},

			// ---- Branch lock-in: ImportType type-arguments ----
			{
				Code:   `type X = import("foo").Bar<T,>`,
				Output: []string{`type X = import("foo").Bar<T>`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 29}},
			},

			// (removed) ConstructorType invalid asserts — upstream has no
			// TSConstructorType listener; ConstructorType is intentionally not
			// visited. See valid lock-in above for the 4-combination proof.

			// ---- Branch lock-in: async / generator with trailing comma + 'never' ----
			{
				Code:    `async function foo(a,) {}`,
				Output:  []string{`async function foo(a) {}`},
				Options: optMap(map[string]any{"functions": "never"}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 21}},
			},
			{
				Code:    `function* gen(a,) {}`,
				Output:  []string{`function* gen(a) {}`},
				Options: optMap(map[string]any{"functions": "never"}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 16}},
			},
			{
				Code:    `const f = async (a,) => 1`,
				Output:  []string{`const f = async (a) => 1`},
				Options: optMap(map[string]any{"functions": "never"}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 19}},
			},

			// (removed) `abstract class A { abstract foo(a): void; }` with
			// `{functions: 'always'}` — body-less class members are skipped
			// to match upstream's TSEmptyBodyFunctionExpression behavior.
			// See the valid lock-in above.

			// ---- Branch lock-in: overload signatures + body-having impl ----
			// Upstream behavior (verified via @stylistic/eslint-plugin v5):
			// body-less overload-signature rows are skipped; only the
			// body-having impl row reports under `{functions: 'always'}`.
			{
				Code:    "class A {\n  foo(a): string;\n  foo(a): number;\n  foo(a) { return a; }\n}",
				Output:  []string{"class A {\n  foo(a): string;\n  foo(a): number;\n  foo(a,) { return a; }\n}"},
				Options: optMap(map[string]any{"functions": "always"}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 4, Column: 8},
				},
			},

			// ---- Branch lock-in: rest-element + 'always-multiline' multi-line
			// + existing trailing comma → forbid via degrade. This is the
			// only invalid shape that exercises the `forceTrailingComma →
			// forbid (rest)` path inside `forceTrailingCommaIfMultiline`'s
			// multi-line branch. ----
			{
				Code:    "function f(\n  a,\n  ...rest,\n) {}",
				Output:  []string{"function f(\n  a,\n  ...rest\n) {}"},
				Options: optMap(map[string]any{"functions": "always-multiline"}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 3, Column: 10}},
			},

			// =====================================================================
			// Extra tsgo edge shapes + real-user scenarios beyond upstream coverage
			// =====================================================================

			// ---- Nested CallExpression: inner has trailing comma, outer doesn't ----
			// Both Call listeners fire independently; the inner's diagnostic should
			// not affect the outer's lastItem scan (since CallExpression.End() is
			// after `)`).
			{
				Code:    `outer(inner(a, b,), c,)`,
				Output:  []string{`outer(inner(a, b), c)`},
				Options: optStr("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 22},
					{MessageId: "unexpected", Line: 1, Column: 17},
				},
			},

			// ---- Method chaining: each Call independent ----
			{
				Code:    `foo(a,).bar(b,).baz(c,)`,
				Output:  []string{`foo(a).bar(b).baz(c)`},
				Options: optStr("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					// listener-fire order is preorder: outermost CallExpression
					// (= foo(a,).bar(b,).baz(c,)) visits first, then descending.
					{MessageId: "unexpected", Line: 1, Column: 22},
					{MessageId: "unexpected", Line: 1, Column: 14},
					{MessageId: "unexpected", Line: 1, Column: 6},
				},
			},

			// ---- ComputedPropertyName key ----
			// `{[x]: 1,}` — computed key in object literal. The OUTER list
			// (Properties) ends after `1`, then `,`. ComputedPropertyName itself
			// is part of the property, not a separate list.
			{
				Code:    `var x = {[k]: 1, [m]: 2,};`,
				Output:  []string{`var x = {[k]: 1, [m]: 2};`},
				Options: optStr("never"),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 24}},
			},

			// ---- Default parameter value ----
			// `function f(a = 1, b = 2,) {}` — default values shouldn't break
			// last-item end detection. Last param ends after `2`.
			{
				Code:    `function f(a = 1, b = 2) {}`,
				Output:  []string{`function f(a = 1, b = 2,) {}`},
				Options: optMap(map[string]any{"functions": "always"}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 24}},
			},

			// ---- Optional-chain call `foo?.(a,)` ----
			// In tsgo this is a CallExpression with QuestionDotToken; same kind,
			// same listener. The `?.` token doesn't affect end-of-list detection.
			{
				Code:    `foo?.(a,)`,
				Output:  []string{`foo?.(a)`},
				Options: optStr("never"),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 8}},
			},
			{
				Code:    `obj?.method(a,)`,
				Output:  []string{`obj?.method(a)`},
				Options: optStr("never"),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 14}},
			},

			// ---- Spread in MIDDLE of call args (NOT at end) ----
			// Last arg is `c`, not the spread — so `'always'` still forces a comma.
			{
				Code:    `foo(a, ...b, c)`,
				Output:  []string{`foo(a, ...b, c,)`},
				Options: optStr("always"),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 15}},
			},

			// ---- TypeScript: type-only import ----
			// `import type {Foo,} from 'bar'` — NamedImports still detected via
			// the same NamedBindings.Kind == KindNamedImports check.
			{
				Code:    `import type {Foo,} from 'bar';`,
				Output:  []string{`import type {Foo} from 'bar';`},
				Options: optStr("never"),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 17}},
			},
			// `import {type Foo,} from 'bar'` — inline type modifier per
			// TS 4.5+; ImportSpecifier carries an `IsTypeOnly` flag but the
			// listener doesn't care, comma detection is purely syntactic.
			{
				Code:    `import {type Foo,} from 'bar';`,
				Output:  []string{`import {type Foo} from 'bar';`},
				Options: optStr("never"),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 17}},
			},

			// ---- TypeScript: export type ----
			{
				Code:    `export type {Foo,} from 'bar';`,
				Output:  []string{`export type {Foo} from 'bar';`},
				Options: optStr("never"),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 17}},
			},

			// ---- Class fields: arrow function with trailing comma in params ----
			// `class A { foo = (a,) => 1; }` — the arrow is a class field
			// initializer. ArrowFunction listener handles it; class-field
			// nesting doesn't bleed.
			{
				Code:    `class A { foo = (a,) => 1; }`,
				Output:  []string{`class A { foo = (a) => 1; }`},
				Options: optStr("never"),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 19}},
			},

			// ---- IIFE: both function and outer call ----
			// `(function f(a,) {})(b,)` — both the FunctionExpression's params
			// and the outer call's args fire independently.
			{
				Code:    `(function f(a,) {})(b,)`,
				Output:  []string{`(function f(a) {})(b)`},
				Options: optStr("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 22},
					{MessageId: "unexpected", Line: 1, Column: 14},
				},
			},

			// ---- Object methods with getter/setter ----
			{
				Code:    `var o = { get x() { return 1 }, set x(v,) {}, };`,
				Output:  []string{`var o = { get x() { return 1 }, set x(v) {} };`},
				Options: optStr("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					// outer object trailing comma at col 45, setter param at col 40
					{MessageId: "unexpected", Line: 1, Column: 45},
					{MessageId: "unexpected", Line: 1, Column: 40},
				},
			},

			// ---- Decorator on class member ----
			// MethodDeclaration with leading decorator still processes Parameters.
			{
				Code:    `class A { @dec foo(a,) {} }`,
				Output:  []string{`class A { @dec foo(a) {} }`},
				Options: optMap(map[string]any{"functions": "never"}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 21}},
			},

			// ---- TS overload signatures: each FunctionDeclaration checked ----
			// Three overloads; the first two are body-less and the third has body.
			// All three are KindFunctionDeclaration → checkFunctionLike fires
			// for each. Each is a separate diagnostic.
			{
				Code:    "function foo(a: string,): string;\nfunction foo(a: number,): number;\nfunction foo(a: any,) { return a; }",
				Output:  []string{"function foo(a: string): string;\nfunction foo(a: number): number;\nfunction foo(a: any) { return a; }"},
				Options: optMap(map[string]any{"functions": "never"}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 23},
					{MessageId: "unexpected", Line: 2, Column: 23},
					{MessageId: "unexpected", Line: 3, Column: 20},
				},
			},

			// ---- Labelled tuple member ----
			// `[name: string, age: number,]` — TS labelled tuple. tsgo's
			// TupleTypeNode.Elements list still has 2 elements, last ends after
			// the type. Trailing comma applies the `tuples` slot.
			{
				Code:    `type T = [name: string, age: number,]`,
				Output:  []string{`type T = [name: string, age: number]`},
				Options: optMap(map[string]any{"tuples": "never"}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 36}},
			},

			// ---- Tuple with rest element ----
			// `[a: string, ...rest: number[]]` — rest in tuple is NamedTupleMember
			// with DotDotDotToken or RestType. In either case, last item is the
			// rest form. tsgo doesn't classify these as KindBindingElement, so my
			// isRestElement returns false → trailing comma is allowed.
			{
				Code:    `type T = [string, ...number[]]`,
				Output:  []string{`type T = [string, ...number[],]`},
				Options: optMap(map[string]any{"tuples": "always"}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 30}},
			},

			// ---- Deep type nesting ----
			// `Foo<Bar<Baz<T,>>>` — three nested TypeReferences with type args.
			// Each has type-args 'never' (hard-coded). Only the innermost has
			// a trailing comma.
			{
				Code:    `type X = Foo<Bar<Baz<T,>>>`,
				Output:  []string{`type X = Foo<Bar<Baz<T>>>`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 23}},
			},

			// ---- Mixed: Call type-args + call args + dynamicImports ----
			// `foo<T,>(a,)` with default 'never': both `T,` (type-args, hard-coded
			// never) and `a,` (functions slot, default never) fire.
			{
				Code:   `foo<T,>(a,)`,
				Output: []string{`foo<T>(a)`},
				Errors: []rule_tester.InvalidTestCaseError{
					// Call listener processes Arguments first, then TypeArgs:
					{MessageId: "unexpected", Line: 1, Column: 10},
					{MessageId: "unexpected", Line: 1, Column: 6},
				},
			},

			// ---- Real-user: React component with many props (multi-line) ----
			{
				Code: "const Component = ({\n  prop1,\n  prop2,\n  prop3\n}) => null;",
				Output: []string{
					"const Component = ({\n  prop1,\n  prop2,\n  prop3,\n}) => null;",
				},
				Options: optStr("always-multiline"),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 4, Column: 8}},
			},

			// ---- Real-user: large Redux action payload object ----
			{
				Code: "dispatch({\n  type: 'FETCH',\n  payload: {\n    id: 1,\n    items: [\n      { name: 'a' },\n      { name: 'b' }\n    ]\n  }\n});",
				Output: []string{
					"dispatch({\n  type: 'FETCH',\n  payload: {\n    id: 1,\n    items: [\n      { name: 'a' },\n      { name: 'b' },\n    ],\n  },\n});",
				},
				Options: optStr("always-multiline"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 9, Column: 4}, // outer object close `}` line
					{MessageId: "missing", Line: 8, Column: 6}, // items array close `]` line
					{MessageId: "missing", Line: 7, Column: 20}, // last item `{ name: 'b' }` ends at col 19, insert at col 20
				},
			},

			// ---- Real-user: long import list with mix of default + named ----
			{
				Code: "import React, {\n  useState,\n  useEffect,\n  useMemo,\n  useCallback\n} from 'react';",
				Output: []string{
					"import React, {\n  useState,\n  useEffect,\n  useMemo,\n  useCallback,\n} from 'react';",
				},
				Options: optMap(map[string]any{"imports": "always-multiline"}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 5, Column: 14}},
			},

			// ---- Real-user: TS function with generics + parameter + return type ----
			// `always` forces commas in:
			//   1. outer FunctionDecl Parameters `(arr, fn: (...))` — col 45
			//   2. outer FunctionDecl TypeParameters `<T, U>` — col 18
			//   3. inner FunctionType params `(x: T)` (the `fn` type) — col 39
			//   4. inner CallExpression args `arr.map(fn)` — col 71
			// Order is preorder DFS — but the exact relative order between (3)
			// inside a Parameter type and (4) deeper in the body depends on
			// listener-fire timing, so assert messageId only and let `Output`
			// pin the substantive fix.
			{
				Code:    "function map<T, U>(arr: T[], fn: (x: T) => U): U[] { return arr.map(fn); }",
				Output:  []string{"function map<T, U,>(arr: T[], fn: (x: T,) => U,): U[] { return arr.map(fn,); }"},
				Options: optStr("always"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing"},
					{MessageId: "missing"},
					{MessageId: "missing"},
					{MessageId: "missing"},
				},
			},
		},
	)
}
