// TestCommaStyleExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Each case carries an inline comment pointing
// at the specific branch / Dimension 4 row / tsgo AST quirk it covers, so
// future refactors can't silently regress them without breaking a named
// lock-in.
package comma_style_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/rules/comma_style"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestCommaStyleExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&comma_style.CommaStyleRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: receiver wrappers — N/A ----
			// N/A: comma-style validates the *position* of `,` tokens
			// between list elements, never the shape of any element itself.
			// Receiver wrappers around list elements (parens, non-null,
			// `as`, `satisfies`, optional-chain) don't change the
			// before-token / after-token text-position lookup.
			//
			// ---- Dimension 4: access / key forms — N/A ----
			// N/A: the rule never inspects member access or property keys.
			//
			// ---- Dimension 4: declaration / container forms ----
			// Lock the FunctionExpression / ArrowFunction / FunctionDeclaration
			// listener fan-out: each must validate its parameters list against
			// the configured style.
			{Code: "function f(a,\n b) {}"},
			{Code: "const f = function (a,\n b) {};"},
			{Code: "const f = (a,\n b) => {};"},
			{Code: "class A { m(a,\n b) {} }"},
			{Code: "class A { constructor(a,\n b) {} }"},
			{Code: "class A { get x() { return 0; } set x(v,\n) {} }"},

			// ---- Dimension 4: same-kind nesting — listener boundary ----
			// Nested classes / functions each have their own param lists.
			// The outer listener must NOT pick up commas inside the inner
			// container's lists (that's the inner listener's job, with the
			// "filter by inner item range" guarantee in `validateList`).
			{Code: "function outer(a, b) {\n  function inner(c, d) {}\n}"},
			{Code: "class A { m() { class B { n() {} } } }"},

			// ---- Dimension 4: graceful degradation — empty / one-element lists ----
			{Code: "function f() {}"},
			{Code: "f();"},
			{Code: "new Foo();"},
			{Code: "var foo = [];"},
			{Code: "var foo = {};"},
			{Code: "var foo = [a];"},
			{Code: "var foo = {a: 1};"},
			{Code: "type T = [];"},
			{Code: "type T = {};"},
			{Code: "type Foo<A> = A;"},

			// ---- tsgo AST quirk: class expression with multi-implements
			// inside a variable initializer ----
			// VariableDeclarationList scan reaches the implements types
			// `X, Y` at no enclosing-bracket depth; the "comma inside the
			// declaration's range" filter is what prevents the heritage
			// commas from being attributed to the var-decl level. Removed
			// the filter would attribute the implements commas as
			// var-decl separators and over-report.
			{Code: "interface X {} interface Y {}\nvar a = class implements X, Y {};"},
			{Code: "interface X {} interface Y {}\nfunction f(x = class implements X, Y {}) {}"},

			// ---- tsgo AST quirk: TypeAliasDeclaration RHS with type
			// arguments ----
			// `type Foo<A, B> = Bar<C, D>` — without the `list.End() + 1`
			// scanEnd bound, checkTypeParams would scan past Foo's `>`
			// into the Bar<C, D> region and double-report the inner commas.
			{Code: "type X = number; type Y = string;\ntype Foo<A, B> = Map<X, Y>;"},

			// ---- tsgo AST quirk: TypeElement separator absorbed ----
			// tsgo's `parseTypeMemberSemicolon` consumes the trailing `,`
			// (or `;`) AS PART OF each TypeElement's range, so member.End()
			// lands past the separator. `itemContentEnd` walks the item's
			// tokens to find the real content end so the separator-comma
			// filter doesn't classify the separator as "inside" the
			// member and skip it.
			{Code: "type T = { a: string, b: string };"},
			{Code: "interface I { a: string; b: string }"},

			// ---- tsgo AST quirk: nested sequence expression ----
			// `(a, (b, c))` — outer BinaryExpression-comma fires on outer
			// `,`; inner fires on its own. Both validate independently
			// since the inner's parent is `ParenthesizedExpression`, not a
			// comma-BinExpr.
			{Code: "var x = (a, (b, c));"},

			// ---- Locks in validateCommaItemSpacing arm 1: all on same line ----
			// Single-line lists hit arm 1 of the switch — every same-line
			// list shape, default or "first", should yield no diagnostic.
			{Code: "var a = [1, 2, 3];"},
			{Code: "var a = {x: 1, y: 2};"},
			{Code: "f(1, 2, 3);"},
			{Code: "var a = [1, 2, 3];", Options: optStr("first")},

			// ---- Locks in upstream "consecutive holes on separate lines"
			// skip-condition (validateOneComma skip 3) ----
			// `[a,\n,\nb]` has two commas vertically stacked. The first
			// gets skipped by the "tokenAfter is comma & not same-line"
			// check; the lone second comma is what reports.
			{Code: "var a = [\n , \n 1, \n 2 \n];"},

			// ---- Locks in array-hole skip (validateOneComma skips 1+2) ----
			// `[, a]` (single hole) — leading comma's `tokenBefore` is `[`,
			// upstream skips. Multi-hole `[, , a]` — second comma's
			// `tokenBefore` walks back through commas to `[`, also skipped.
			// Lock both shapes in.
			{Code: "var a = [, 1];"},
			{Code: "var a = [, , 1];"},
			{Code: "var a = [, , , 1];"},

			// ---- Locks in: TSX-tagged file with type-args / type-params ----
			// `<>` in `.tsx` is JSX. Inside type-args / type-params we
			// still need to recognize `<` / `>`. checkTypeArgs uses
			// `list.End() + 1` so the close `>` is in the scan token
			// slice for trailing-comma `tokenAfter` lookups.
			{Code: "function f<A, B>() {}", Tsx: true},
			{Code: "var x = foo<A,\n B>();", Tsx: true},

			// ---- Audit: default parameter with nested object literal ----
			// Inner object's commas must NOT be attributed to the parameter
			// list. Confirmed by the "comma inside item range" filter on
			// the FunctionDeclaration's Parameters list.
			{Code: "function f(a = { b: 1,\n c: 2 }, d) {}"},

			// ---- Audit: computed property name shape ----
			{Code: "var x = { [a]: 1 };"},

			// ---- Audit: JSX self-closing with type args (TSX) ----
			{Code: "var x = <Foo<A,\n B> />;", Tsx: true},

			// ---- Audit: as / satisfies wrappers ----
			{Code: "var x = [1,\n 2] as const;"},
			{Code: "var x = [1,\n 2] satisfies number[];"},

			// ---- Audit: decorator-call expression ----
			{Code: "function dec(): any { return () => {}; }\n@dec({a: 1,\n b: 2}) class A {}"},

			// ---- Audit: generator / async function params ----
			{Code: "function* g(a,\n b) { yield a; }"},
			{Code: "async function f(a,\n b) { return a; }"},

			// ---- Audit: generic constraint with nested object type ----
			// Object-type member commas live inside the TypeParameter's
			// range; the filter excludes them at the type-param-list
			// level.
			{Code: "function f<T extends { a: number,\n b: string }>(x: T) { return x; }"},

			// ---- Audit: nested generic type args ----
			// `Array<Map<string,\nnumber>>` — outer has 1 type-arg, no
			// comma; inner Map has 2 type-args separated by a violating
			// comma — but inner Map's `,` is on the same line as `number`,
			// so under "last" no diagnostic fires.
			{Code: "type T = Array<Map<string,\n number>>;"},

			// ---- Audit: abstract method (TSEmptyBodyFunctionExpression bucket) ----
			{
				Code:    "abstract class A {\n  abstract m(a,\n b): void;\n}",
				Options: optWithExc("last", map[string]any{"TSEmptyBodyFunctionExpression": true}),
			},

			// ---- Audit: overload signatures ----
			// First decl body-less (→ TSDeclareFunction bucket), second
			// has a body (→ FunctionDeclaration bucket).
			{
				Code:    "function f(a,\n b): number;\nfunction f(a, b) { return 0; }",
				Options: optWithExc("last", map[string]any{"TSDeclareFunction": true}),
			},

			// ---- Audit: for-init VariableDeclarationList ----
			// for-init wraps a VariableDeclarationList without a
			// surrounding VariableStatement; the same listener fires.
			{Code: "for (var i = 0, j = 1; i < 10; i++) {}"},
			{
				Code:    "for (var i = 0,\n j = 1; i < 10; i++) {}",
				Options: optWithExc("last", map[string]any{"VariableDeclaration": true}),
			},

			// ---- Audit: SequenceExpression inside `if` ----
			{Code: "if (a, b) {}", Options: optStr("first")},

			// ---- Audit: all-empty containers ----
			{Code: "function f() {}"},
			{Code: "f();"},
			{Code: "[];"},
			{Code: "({});"},
			{Code: "type T = [];"},
			{Code: "type T = {};"},

			// ---- Real-user: eslint/eslint#6006 — parenthesized array item ----
			// `[\n  ('',\n  ),\n  def,\n]` — the `,` after `)` is on same
			// line as `)`, so sameLineBefore=true, sameLineAfter=false →
			// no diagnostic under style=last. Historic false-positive.
			{
				Code: "const def: any = {};\nconst abc = [\n  (''\n  ),\n  def,\n];",
			},

			// ---- Real-user: eslint/eslint#10273 — sparse array w/ nested parens ----
			// Trailing-comma items + leading-hole separators across lines
			// should NOT report under style=last; each separator comma is
			// on the same line as its preceding `]`-terminated element.
			{
				Code: "class Thing {}\n" +
					"function f(_n: any) {}\n" +
					"const testArray = [\n" +
					"  [1],,\n" +
					"  [new Thing()],\n" +
					"  [new Thing(),],\n" +
					"  [new Thing()],\n" +
					"  [{an: 'object'}],,\n" +
					"  [[7, 8]],,\n" +
					"  [9]\n" +
					"];",
			},

			// ---- Real-user: eslint/eslint#12756 — consecutive holes on
			// separate lines should not error ----
			// `const [\n  ,,\n  a, b,\n] = arr` and the variant with each
			// hole on its own line. Skip path: tokenAfter is comma AND
			// not on same line as current comma → don't report.
			{
				Code: "const [\n  ,,\n  cp = 'config.json',\n  af = 'auth',\n] = ['1','2','3','4'];",
			},
			{
				Code: "const [\n  ,\n  ,\n  cp = 'config.json',\n  af = 'auth',\n] = ['1','2','3','4'];",
			},

		},
		[]rule_tester.InvalidTestCase{
			// ---- Locks in validateCommaItemSpacing arm 2 (lone comma) ----
			// `a,\n,\nb` — both before and after the second `,` are on
			// different lines; arm 2 fires with the 'between' fix shape
			// (strip first linebreak, prepend comma).
			{
				Code:   "var a = 1\n,\nb = 2;\n",
				Output: []string{"var a = 1,\nb = 2;\n"},
				Errors: errLone(2, 1),
			},

			// ---- Locks in arm-2 block-comment branch ----
			// When the first comment after the lone comma is a Block
			// comment on the same line, the fix uses the configured style
			// (here 'last') instead of 'between' so the block comment
			// keeps its line relative to the comma. Style=last leaves the
			// trivia after the comma untouched, so the `\n` between the
			// block comment and `b` is preserved verbatim.
			{
				Code:   "[a\n,/*block*/\nb];",
				Output: []string{"[a,\n/*block*/\nb];"},
				Errors: errLone(2, 1),
			},

			// ---- Locks in arm 3 (style=first && !sameLineAfter) ----
			{
				Code:    "var x = [1,\n2];",
				Output:  []string{"var x = [1\n,2];"},
				Options: optStr("first"),
				Errors:  errFirst(1, 11),
			},

			// ---- Locks in arm 4 (style=last && sameLineAfter) ----
			{
				Code:   "var x = [1\n, 2];",
				Output: []string{"var x = [1,\n 2];"},
				Errors: errLast(2, 1),
			},

			// ---- Locks in CRLF-aware linebreak removal ----
			// `removeFirstLinebreak` must treat CR LF as one sequence, not
			// strip only LF and leave the CR behind. Without this, the
			// 'between' fix would leave a stray `\r` in the output.
			{
				Code:   "var a = 1\r\n,\r\nb = 2;\r\n",
				Output: []string{"var a = 1,\r\nb = 2;\r\n"},
				Errors: errLone(2, 1),
			},

			// ---- Locks in sequence-comma chain validation ----
			// `(a, b, c)` — left-associative `BinaryExpression(BinaryExpression(a, b), c)`.
			// The OUTER comma-BinExpr fires once, flattens the chain into
			// `[a, b, c]`, and finds 2 separator commas. Inner level
			// doesn't re-fire because parent is comma-BinExpr.
			{
				Code:    "(a\n, b\n, c);",
				Output:  []string{"(a,\n b,\n c);"},
				Options: optStr("last"),
				Errors: errs(
					errSpec{"expectedCommaLast", 2, 1},
					errSpec{"expectedCommaLast", 3, 1},
				),
			},

			// ---- Locks in: trailing comma after lone item reports under "last" ----
			// `[a,\n]` — trailing `,` has tokenAfter = `]`. sameLineBefore
			// = true (a on same line as comma), sameLineAfter = false
			// (comma on prev line vs `]`). With style=last, the arm
			// !sameLineBefore && !sameLineAfter doesn't fire (only one
			// side differs), and arm 4 doesn't fire either (sameLineAfter
			// is false). So no report — confirm green.
			//
			// (Confirmed green: no entry needed here; included in valid
			// set above as `var foo = [1, ];` style cases.)
			//
			// What we *do* lock: trailing comma where comma is on its own
			// line, both before AND after are different lines → arm 2.
			{
				Code:   "var a = [1\n,\n];",
				Output: []string{"var a = [1,\n];"},
				Errors: errLone(2, 1),
			},

			// ---- tsgo quirk lock-in: heritage clause inside class
			// expression inside variable initializer ----
			// Regression test for the original implementation that scanned
			// VariableDeclarationList with depth tracking and incorrectly
			// reported the `implements X, Y` comma as a var-decl separator.
			// The AST-driven filter must keep this green.
			{
				Code:    "interface X {} interface Y {}\nvar a = class implements\nX\n, Y {};",
				Output:  []string{"interface X {} interface Y {}\nvar a = class implements\nX,\n Y {};"},
				Options: optStr("last"),
				Errors:  errLast(4, 1),
			},

			// ---- tsgo quirk lock-in: typeMember separator absorbed
			// (style=first reordering) ----
			// `type T = { a: T1, b: T2 }` style=first should report the
			// member separator `a: T1,` and move the `,` to start of
			// `b: T2`'s line. The `itemContentEnd` trim is what surfaces
			// this comma as a separator; without it the comma is hidden
			// inside the member's range.
			{
				Code: "type T = {\n" +
					"  a: string,\n" +
					"  b: number\n" +
					"};",
				Output: []string{
					"type T = {\n" +
						"  a: string\n" +
						"  ,b: number\n" +
						"};",
				},
				Options: optStr("first"),
				Errors:  errFirst(2, 12),
			},

			// ---- Audit: trailing comma in multi-line array under "first" ----
			// All three commas (between elements + trailing) violate; the
			// fix moves each to the start of the following item's line.
			{
				Code:    "var x = [1,\n2,\n3,\n];",
				Output:  []string{"var x = [1\n,2\n,3\n,];"},
				Options: optStr("first"),
				Errors: errs(
					errSpec{"expectedCommaFirst", 1, 11},
					errSpec{"expectedCommaFirst", 2, 2},
					errSpec{"expectedCommaFirst", 3, 2},
				),
			},

			// ---- Audit: nested ArrayLiteral inside ObjectLiteral ----
			// Only the OUTER object property comma is reported (the inner
			// `[1, 2]` is single-line and quiet).
			{
				Code:    "var x = {\n  a: [1, 2],\n  b: [3, 4]\n  , c: 5\n};",
				Output:  []string{"var x = {\n  a: [1, 2],\n  b: [3, 4],\n   c: 5\n};"},
				Options: optStr("last"),
				Errors:  errLast(4, 3),
			},

			// ---- Audit: class with extends + implements + type-args ----
			// Multi-line type-args under "last" report the leading comma.
			{
				Code:    "class A extends X<\n  P\n  , Q\n> implements Y {}\ninterface X<T1, T2> {}\ninterface Y {}",
				Output:  []string{"class A extends X<\n  P,\n   Q\n> implements Y {}\ninterface X<T1, T2> {}\ninterface Y {}"},
				Options: optStr("last"),
				Errors:  errLast(3, 3),
			},

			// ---- Real-user: eslint-stylistic#598 — trailing comma in
			// object should report (used to be missed pre-fix) ----
			// `{a, b\n  ,}` with style=last: trailing `,` is on a
			// different line from `b` but on the same line as `}` → arm 4
			// fires (expectedCommaLast).
			{
				Code:   "let a: any, b: any; const c = {a, b\n  ,};",
				Output: []string{"let a: any, b: any; const c = {a, b,\n  };"},
				Errors: errLast(2, 3),
			},

			// ---- Real-user: eslint-stylistic#599 — comma between default
			// import and named bindings should report ----
			{
				Code:   "import a\n  , {Foo} from 'module';",
				Output: []string{"import a,\n   {Foo} from 'module';"},
				Errors: errLast(2, 3),
			},
			{
				Code:    "import c,\n  {Bar} from 'module';",
				Output:  []string{"import c\n  ,{Bar} from 'module';"},
				Options: optStr("first"),
				Errors:  errFirst(1, 9),
			},

			// ---- Real-user: import/export with attributes (`with`)
			// — comma between attributes spanning lines ----
			{
				Code:   "import x from 'x' with {type: 'json'\n  ,foo: 'bar'};",
				Output: []string{"import x from 'x' with {type: 'json',\n  foo: 'bar'};"},
				Errors: errLast(2, 3),
			},

			// ---- Real-user: dynamic import with options arg ----
			{
				Code:   "import('x'\n  ,{with: {type: 'json'}});",
				Output: []string{"import('x',\n  {with: {type: 'json'}});"},
				Errors: errLast(2, 3),
			},

			// ---- Real-user: type alias with multi-line generics ----
			{
				Code:   "type G1<A\n  ,B> = Map<A, B>;",
				Output: []string{"type G1<A,\n  B> = Map<A, B>;"},
				Errors: errLast(2, 3),
			},

			// ---- Real-user: rsbuild bulk diff — function-like
			// parameter list with trailing comma under style=first ----
			// Regression: an earlier impl used `node.End()` as scanEnd
			// for function-like parameter lists, which extended past
			// the close `)` and into the body, mis-classifying any
			// body-level commas as parameter-list separators. Then
			// fixed to `list.End()+1`, but that was off-by-one due to
			// tsgo's `nodePos() == TokenFullStart`. Now uses
			// `validateBracketedListAtCloseChar` which SkipTrivia's to
			// the actual `)` byte. Lock both shapes in.
			{
				Code:    "const f = (\n  a: number,\n  b: number,\n) => {};",
				Output:  []string{"const f = (\n  a: number\n  ,b: number\n,) => {};"},
				Options: optStr("first"),
				Errors: errs(
					errSpec{"expectedCommaFirst", 2, 12},
					errSpec{"expectedCommaFirst", 3, 12},
				),
			},

			// ---- Real-user: rsbuild bulk diff — trailing comma in call
			// argument containing a template literal ----
			// Regression: the scanner doesn't carry template state across
			// `GetScannerForSourceFile(sf, pos)`, so a closing back-tick
			// inside a call argument was mis-tokenized as a fresh
			// `NoSubstitutionTemplateLiteral` swallowing every byte up
			// to EOF — including the call's trailing `,` and `)`. Now
			// `validateList` skips PAST each item's range by re-creating
			// the scanner at item.End(), AND inserts a pseudo-token
			// representing the item so adjacent-comma sameLine checks
			// still resolve correctly.
			{
				Code: "function bar(x: any): string { return ''; }\n" +
					"function foo() {\n" +
					"  return `foo ${bar(\n" +
					"    'x'\n" +
					"  )} baz`\n" +
					"  ,'y';\n" +
					"}",
				Output: []string{
					"function bar(x: any): string { return ''; }\n" +
						"function foo() {\n" +
						"  return `foo ${bar(\n" +
						"    'x'\n" +
						"  )} baz`,\n" +
						"  'y';\n" +
						"}",
				},
				Errors: errLast(6, 3),
			},

			// ---- Real-user: type literal with method signatures ----
			// Member separator commas + parameter commas — independently
			// reported. The reported position is for the method's
			// param-list comma (the TypeLiteral's member-comma reports
			// would only fire if a member's `,` were itself on a wrong
			// line).
			{
				Code: "type ML = {\n" +
					"  m1(a: number\n" +
					"  ,b: number): void;\n" +
					"};",
				Output: []string{
					"type ML = {\n" +
						"  m1(a: number,\n" +
						"  b: number): void;\n" +
						"};",
				},
				Errors: errLast(3, 3),
			},
		},
	)
}
