// TestDotLocationExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Each case carries an inline comment pointing
// at the specific branch / Dimension 4 row / tsgo AST quirk it covers, so
// future refactors can't silently regress them without breaking a named
// lock-in.
package dot_location_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/rules/dot_location"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestDotLocationExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&dot_location.DotLocationRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: TS non-null assertion receiver ----
			{Code: "obj!.\nprop", Options: optsObject()},
			{Code: "obj!\n.prop", Options: optsProperty()},

			// ---- Dimension 4: TS as-expression receiver (lives inside parens) ----
			{Code: "(obj as Foo).\nprop", Options: optsObject()},
			{Code: "(obj as Foo)\n.prop", Options: optsProperty()},

			// ---- Dimension 4: TS satisfies-expression receiver (lives inside parens) ----
			{Code: "(obj satisfies Foo).\nprop", Options: optsObject()},
			{Code: "(obj satisfies Foo)\n.prop", Options: optsProperty()},

			// ---- Dimension 4: chained property access — each dot on object-side ----
			{Code: "a.\nb.\nc", Options: optsObject()},
			{Code: "a\n.b\n.c", Options: optsProperty()},

			// ---- Dimension 4: chained TS QualifiedName ----
			{Code: "type T = A.\nB.\nC", Options: optsObject()},
			{Code: "type T = A\n.B\n.C", Options: optsProperty()},

			// ---- Dimension 4: ImportType with chained qualifier ----
			{Code: "type T = import('m').\nA.\nB", Options: optsObject()},
			{Code: "type T = import('m')\n.A\n.B", Options: optsProperty()},

			// ---- Dimension 4: chained JSX tag name (<A.B.C />) ----
			{Code: "const _ = <A.B.\nC />", Options: optsObject(), Tsx: true},
			{Code: "const _ = <A\n.B\n.C />", Options: optsProperty(), Tsx: true},

			// ---- Branch lock-in: ElementAccessExpression skipped — arr[x] is never reported ----
			{Code: "arr\n[\n0\n]\n[\n1\n]", Options: optsObject()},
			{Code: "arr\n[\n0\n]\n[\n1\n]", Options: optsProperty()},

			// ---- Branch lock-in: ImportType with no qualifier → checkDotLocation returns early ----
			{Code: "type Foo = import('m')\n", Options: optsObject()},
			{Code: "type Foo = import('m')\n", Options: optsProperty()},

			// ---- Branch lock-in: ImportType with import attributes ----
			{Code: "type T = import('m', { with: { type: 'json' } }).\nA", Options: optsObject()},
			{Code: "type T = import('m', { with: { type: 'json' } })\n.A", Options: optsProperty()},

			// ---- Branch lock-in: typeof ImportType ----
			{Code: "type T = typeof import('m').\nA", Options: optsObject()},
			{Code: "type T = typeof import('m')\n.A", Options: optsProperty()},

			// ---- Real-user: Promise / fluent method chain ----
			{Code: "fetch(url).\nthen(r => r.json()).\nthen(data => use(data))", Options: optsObject()},
			{Code: "fetch(url)\n.then(r => r.json())\n.then(data => use(data))", Options: optsProperty()},

			// ---- Real-user: jQuery-style chain on string-method call ----
			{Code: "name.trim().\nsplit(' ').\nfilter(Boolean)", Options: optsObject()},
			{Code: "name.trim()\n.split(' ')\n.filter(Boolean)", Options: optsProperty()},

			// ---- Dimension 4: NewExpression as receiver ----
			// tsgo: outer PAE.Expression.Kind = KindNewExpression. Verifies
			// the scan skips through `new Foo(...)` (including its argument
			// list with internal tokens like `,`) and lands on `)` as prev.
			{Code: "new Foo().\nbar", Options: optsObject()},
			{Code: "new Foo()\n.bar", Options: optsProperty()},

			// ---- Dimension 4: receiver is a CallExpression with TS type arguments ----
			// `Array.from<T>(x)` — the `<T>` inside Expression must not trip
			// scanner up (TS angle brackets coexist with JSX angle brackets
			// in tsx mode). Validates Expression.End() lands on the `)` of
			// the call, not on the `>` of the type argument list.
			{Code: "Array.from<number>([1]).\nmap(fn)", Options: optsObject()},
			{Code: "Array.from<number>([1])\n.map(fn)", Options: optsProperty()},

			// ---- Dimension 4: receiver is paren-wrapped await / yield ----
			// tsgo: Expression.Kind = ParenthesizedExpression, prev-token = `)`.
			{Code: "async function f() { return (await foo).\nbar }", Options: optsObject()},
			{Code: "function* g() { yield (yield 1)\n.x }", Options: optsProperty()},

			// ---- Dimension 4: chained NonNullExpression `obj!!.prop` ----
			// tsgo allows redundant `!!`; the outer NonNullExpression wraps
			// the inner one. prev-token to dot is still `!`, on same line.
			{Code: "obj!!\n.prop", Options: optsProperty()},
			{Code: "obj!!.\nprop", Options: optsObject()},

			// ---- Dimension 4: multi-stage `as` chain `(a as B as C).prop` ----
			// Each `as` keeps the expression flat (no extra paren); the
			// paren wraps the outermost AsExpression. prev-token = `)`.
			{Code: "(a as B as C).\nprop", Options: optsObject()},
			{Code: "(a as B as C)\n.prop", Options: optsProperty()},

			// ---- Dimension 4: nested ImportType in type arguments ----
			// `import('m').A<import('n').B>` — outer ImportType qualifier=A
			// with TypeArguments containing an inner ImportType. Verifies
			// depth-counting buffer flushes BOTH dots in source order.
			{Code: "type T = import('m').\nA<import('n').\nB>", Options: optsObject()},
			{Code: "type T = import('m')\n.A<import('n')\n.B>", Options: optsProperty()},

			// ---- Dimension 4: ImportType inside TypeReference generic ----
			// `Array<import('m').A.B>` — outer TypeReference (not listened),
			// inner ImportType + QualifiedName both fire. depth counts only
			// our listened kinds, so flushing happens when ImportType exits.
			{Code: "type T = Array<import('m').\nA.\nB>", Options: optsObject()},

			// ---- Dimension 4: deep chain (5 segments) — depth+sort stress ----
			{Code: "a.\nb.\nc.\nd.\ne", Options: optsObject()},

			// ---- Dimension 4: mixed optional + regular chain ----
			// `a?.b.c?.d.e` — alternating optional/regular dots. All on
			// object side when option='object'.
			{Code: "a?.\nb.\nc?.\nd.\ne", Options: optsObject()},

			// ---- Branch lock-in: receiver with `delete` operator ----
			// `delete obj\n.prop` — the PAE is `obj\n.prop`, UnaryExpression
			// wraps it. dot still detected via normal path.
			{Code: "delete obj.\nprop", Options: optsObject()},

			// ---- Real-user: eslint#2504 historical regression ----
			// `(console)\n.log("hi")` in property mode — paren-wrapped
			// receiver MUST NOT be reported when dot is on same line as
			// property. Was a long-standing bug fixed in eslint/eslint#11933.
			{Code: "(console)\n.log('hi')", Options: optsProperty()},

			// ---- Real-user: React class instance state access ----
			// `this.state.foo.bar` — three-level PAE chain typical in class
			// components. No newlines, no reports.
			{Code: "class C { x = 1; m() { return this.state.foo.bar; } }", Options: optsObject()},

			// ---- Real-user: DOM query chain ----
			{Code: "document.querySelector('.x').\naddEventListener('click', fn)", Options: optsObject()},
			{Code: "document.querySelector('.x')\n.addEventListener('click', fn)", Options: optsProperty()},

			// ---- Real-user: styled-components style template tag ----
			// `styled.div` is PAE on identifier — tagged template doesn't
			// change the PAE shape. Dot single-line, no report.
			{Code: "const Btn = styled.div`color: red`", Options: optsObject()},

			// ---- Real-user: typeof import qualifier (Drizzle/Prisma pattern) ----
			{Code: "type Schema = typeof import('./schema').\nschema", Options: optsObject()},

			// ---- Real-user: lodash-style deep chain ----
			{Code: "_.chain(arr).map(f).filter(g).reduce(h, 0).value()", Options: optsObject()},
			{Code: "_.chain(arr)\n.map(f)\n.filter(g)\n.reduce(h, 0)\n.value()", Options: optsProperty()},

			// ---- Dimension 4: JSX tag name with `this` (TS-allowed) ----
			// tsgo: tag name = PAE(Expression=KindThisKeyword, name=Component).
			// `this` is a valid JsxTagNameExpression member.
			{Code: "const X = <this.Component />", Options: optsObject(), Tsx: true},
			{Code: "const X = <this.\nComponent />", Options: optsObject(), Tsx: true},
			{Code: "const X = <this\n.Component />", Options: optsProperty(), Tsx: true},

			// ---- Dimension 4: JSX element wrapped in paren as receiver ----
			// `(<Foo />).props` — PAE.Expression.Kind = ParenthesizedExpression
			// containing JsxSelfClosingElement. prev-token = `)`.
			{Code: "const x = (<Foo />).props", Options: optsObject(), Tsx: true},
			{Code: "const x = (<Foo />).\nprops", Options: optsObject(), Tsx: true},
			{Code: "const x = (<Foo />)\n.props", Options: optsProperty(), Tsx: true},

			// ---- Dimension 4: ClassExpression receiver ----
			// `(class {}).constructor` — paren-wrapped class expression.
			{Code: "const c = (class {}).\nconstructor", Options: optsObject()},
			{Code: "const c = (class {})\n.constructor", Options: optsProperty()},

			// ---- Dimension 4: TaggedTemplateExpression as PAE receiver ----
			// `styled.div\`...\`.attrs({})` — outer PAE.Expression is a
			// TaggedTemplateExpression spanning the tag + template literal.
			{Code: "const B = styled.div`x`.\nattrs({})", Options: optsObject()},
			{Code: "const B = styled.div`x`\n.attrs({})", Options: optsProperty()},

			// ---- Dimension 4: Unicode identifier (BMP) ----
			// Greek / CJK identifiers — verifies byte-offset position math
			// works on non-ASCII code points. tsgo source positions are
			// byte offsets; column reporting uses UTF-16 code units (see
			// rule_tester's GetECMALineAndUTF16CharacterOfPosition).
			{Code: "var α = {}; α.\nβ", Options: optsObject()},
			{Code: "var 变量 = {}; 变量.\n属性", Options: optsObject()},

			// ---- Dimension 4: Unicode identifier (SMP / surrogate pair) ----
			// `𝕒` is U+1D552, encoded as a UTF-16 surrogate pair (2 code
			// units, 4 UTF-8 bytes). Locks in that column counts go through
			// UTF-16 — otherwise positions on lines containing surrogate
			// pairs would be off by 1 per occurrence.
			{Code: "var 𝕒: any = {}; 𝕒.\n𝕓", Options: optsObject()},

			// ---- Dimension 4: MetaProperty nested as PAE.Expression ----
			// `import.meta.url` — outer PAE wraps inner MetaProperty.
			// Buffer+sort must emit MetaProperty's dot (if any) before
			// PAE's dot when both exist.
			{Code: "const u = import.meta.\nurl", Options: optsObject()},
			{Code: "const u = import.meta\n.url", Options: optsProperty()},

			// ---- Dimension 4: MetaProperty `new.target` nested ----
			{Code: "function f() { return new.target.\nname }", Options: optsObject()},

			// ---- Dimension 4: ElementAccessExpression as PAE.Expression ----
			// `obj[k].method` — outer PAE prev-token is `]`. We don't fire
			// on the inner ElementAccess; only the outer PAE.
			{Code: "obj[k].\nmethod()", Options: optsObject()},
			{Code: "obj[k]\n.method()", Options: optsProperty()},

			// ---- Dimension 4: ConditionalExpression in paren receiver ----
			{Code: "(cond ? a : b).\nmethod()", Options: optsObject()},
			{Code: "(cond ? a : b)\n.method()", Options: optsProperty()},

			// ---- Dimension 4: ArrayLiteralExpression as receiver ----
			// prev-token = `]`. Same shape as ElementAccess closing bracket.
			{Code: "[1, 2].\nlength", Options: optsObject()},
			{Code: "[1, 2]\n.length", Options: optsProperty()},

			// ---- Dimension 4: StringLiteral / RegExp literal as receiver ----
			{Code: "'x'.\nlength", Options: optsObject()},
			{Code: "/foo/.\ntest('x')", Options: optsObject()},
			{Code: "'x'\n.length", Options: optsProperty()},

			// ---- Dimension 4: ObjectLiteralExpression in paren ----
			{Code: "({a: 1}).\na", Options: optsObject()},
			{Code: "({a: 1})\n.a", Options: optsProperty()},

			// ---- Dimension 4: Block comment crossing newlines (valid in property mode) ----
			// `obj /* a\nb */.x` — block comment puts the dot on a different
			// SOURCE line than `obj`. In property mode, dot and property `x`
			// are still on the same line (line 2), so no report.
			{Code: "obj /* a\nb */.x", Options: optsProperty()},

			// ---- Dimension 4: Line comment between obj and dot (property mode) ----
			// `obj // c\n.x` — line comment forces `\n`. In property mode,
			// dot and prop are on the same line (line 2), so no report.
			{Code: "obj // c\n.x", Options: optsProperty()},

			// ---- Dimension 4: Line comment between dot and property (object mode) ----
			// `obj.// c\nx` — line comment forces `\n` after dot. In object
			// mode, dot is still adjacent to obj on line 1 → no report.
			{Code: "obj.// c\nx", Options: optsObject()},

			// ---- Dimension 4: Multiple comments around dot ----
			// Already partially covered by upstream `foo /* a */ . /* b */ \n /* c */ bar`.
			// Property-mode variant where dot and prop are forced apart by
			// a line comment.
			{Code: "obj /* a */ // b\n.x", Options: optsProperty()},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: TS non-null assertion — token before dot is `!` ----
			// Locks in that prev-token detection captures `!` (not the identifier
			// before it) when the receiver is `obj!`.
			{
				Code:    "obj!\n.prop",
				Output:  []string{"obj!.\nprop"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
				},
			},
			{
				Code:    "obj!.\nprop",
				Output:  []string{"obj!\n.prop"},
				Options: optsProperty(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotBeforeProperty", Line: 1, Column: 5},
				},
			},

			// ---- Dimension 4: TS as-expression — token before dot is `)` ----
			{
				Code:    "(obj as Foo)\n.prop",
				Output:  []string{"(obj as Foo).\nprop"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
				},
			},
			{
				Code:    "(obj as Foo).\nprop",
				Output:  []string{"(obj as Foo)\n.prop"},
				Options: optsProperty(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotBeforeProperty", Line: 1, Column: 13},
				},
			},

			// ---- Dimension 4: chained PropertyAccessExpression — independent dots ----
			// Verifies inner and outer dots each get their own diagnostic.
			{
				Code:    "a\n.b\n.c",
				Output:  []string{"a.\nb.\nc"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
					{MessageId: "expectedDotAfterObject", Line: 3, Column: 1},
				},
			},

			// ---- Dimension 4: chained TS QualifiedName — independent dots ----
			{
				Code:    "type T = A\n.B\n.C",
				Output:  []string{"type T = A.\nB.\nC"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
					{MessageId: "expectedDotAfterObject", Line: 3, Column: 1},
				},
			},

			// ---- Dimension 4: ImportType + nested QualifiedName ----
			// Locks in that the outer ImportType node and the inner
			// QualifiedName node each report independently.
			{
				Code:    "type T = import('m')\n.A\n.B",
				Output:  []string{"type T = import('m').\nA.\nB"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
					{MessageId: "expectedDotAfterObject", Line: 3, Column: 1},
				},
			},

			// ---- Dimension 4: JSX tag name chain ----
			{
				Code:    "const _ = <A.B\n.C />",
				Output:  []string{"const _ = <A.B.\nC />"},
				Options: optsObject(),
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
				},
			},

			// ---- Branch lock-in upstream getProperty arm C (PropertyAccess) ×
			//      fix branch I (non-decimal-integer numeric token) ----
			// BigInt literal `5n` — Kind is BigIntLiteral, NOT NumericLiteral,
			// so isDecimalIntegerNumericToken returns false. Fix collapses to
			// `5n.\ntoString()` with no space; `5n.foo` parses fine since
			// BigInt literal consumes the trailing `n`.
			{
				Code:    "5n\n.toString()",
				Output:  []string{"5n.\ntoString()"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
				},
			},

			// ---- Branch lock-in fix branch I: scientific notation ----
			// `5e10` has an exponent, fails the decimal-integer regex even
			// though it is a NumericLiteral. No leading space inserted.
			{
				Code:    "5e10\n.toString()",
				Output:  []string{"5e10.\ntoString()"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
				},
			},

			// ---- Branch lock-in fix branch I: hex literal ----
			{
				Code:    "0xff\n.toString()",
				Output:  []string{"0xff.\ntoString()"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
				},
			},

			// ---- Branch lock-in fix branch I: float literal (has decimal point) ----
			// `5.5` is NumericLiteral but the regex `^(?:0|0[0-7]*[89][0-9]*|[1-9](?:_?[0-9])*)$`
			// doesn't match (contains `.`). Resulting `5.5.toExp()` is a valid
			// property access on a float.
			{
				Code:    "5.5\n.toExp()",
				Output:  []string{"5.5.\ntoExp()"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
				},
			},

			// ---- Branch lock-in fix branch H ×  questionDot dotText ----
			// `10\n?.prop` exercises the path where the prev token IS a
			// decimal integer literal but dotText is `?.` (starts with `?`,
			// not `.`) — branch H's condition requires BOTH, so we take
			// branch I and emit no space. Locks in `10?.\nprop`, not
			// `10 ?.\nprop`.
			{
				Code:    "100\n?.prop",
				Output:  []string{"100?.\nprop"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1, EndLine: 2, EndColumn: 3},
				},
			},

			// ---- Branch lock-in: MetaProperty newline (import.meta) ----
			// Upstream tests MetaProperty only as a `valid` case (`import.meta`).
			// We lock in the report path for whitespace splits — `import\n.meta`
			// is syntactically legal even though linter-formatted code never
			// produces it.
			{
				Code:    "const m = import\n.meta",
				Output:  []string{"const m = import.\nmeta"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
				},
			},
			{
				Code:    "const m = import.\nmeta",
				Output:  []string{"const m = import\n.meta"},
				Options: optsProperty(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotBeforeProperty", Line: 1, Column: 17},
				},
			},

			// ---- Branch lock-in: MetaProperty newline (new.target) ----
			// `new.target` is only valid inside a function body in JS, hence
			// the wrapping `function f() { ... }`.
			{
				Code:    "function f() { return new\n.target }",
				Output:  []string{"function f() { return new.\ntarget }"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
				},
			},
			{
				Code:    "function f() { return new.\ntarget }",
				Output:  []string{"function f() { return new\n.target }"},
				Options: optsProperty(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotBeforeProperty", Line: 1, Column: 26},
				},
			},

			// ---- Branch lock-in: ImportType with ImportAttributes — token-before-dot is `)` ----
			// Locks in that scanning from node.Pos() correctly skips OVER the
			// attributes object literal `{ with: { type: 'json' } }` (containing
			// its own dots) and lands on the `)` token as prev-token. Without
			// the `minDotPos` filter we'd false-positive on the attribute's
			// internal `with:` colon-prefixed property — but those aren't dots
			// so this is mostly a defensive lock-in.
			{
				Code:    "type T = import('m', { with: { type: 'json' } })\n.A",
				Output:  []string{"type T = import('m', { with: { type: 'json' } }).\nA"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
				},
			},

			// ---- Branch lock-in: `typeof import(...)` — IsTypeOf=true ----
			// Verifies the `typeof` prefix doesn't break ImportType dot
			// resolution: prev-token is still `)`, dot is between `)` and
			// the qualifier.
			{
				Code:    "type T = typeof import('m')\n.A",
				Output:  []string{"type T = typeof import('m').\nA"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
				},
			},

			// ---- Branch lock-in: optional chain on chained PAE ----
			// `a?.b?.c?.d` — three optional-chain dots, all should report
			// independently when each lands on its own line.
			{
				Code:    "a\n?.b\n?.c\n?.d",
				Output:  []string{"a?.\nb?.\nc?.\nd"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1, EndLine: 2, EndColumn: 3},
					{MessageId: "expectedDotAfterObject", Line: 3, Column: 1, EndLine: 3, EndColumn: 3},
					{MessageId: "expectedDotAfterObject", Line: 4, Column: 1, EndLine: 4, EndColumn: 3},
				},
			},

			// ---- Real-user: long fluent chain in option=object mode ----
			// Mirrors a common Promise-chain shape that gets lint-cleaned-up
			// by the rule.
			{
				Code:    "fetch(url)\n.then(r => r.json())\n.then(data => use(data))",
				Output:  []string{"fetch(url).\nthen(r => r.json()).\nthen(data => use(data))"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
					{MessageId: "expectedDotAfterObject", Line: 3, Column: 1},
				},
			},
			{
				Code:    "name.trim().\nsplit(' ').\nfilter(Boolean)",
				Output:  []string{"name.trim()\n.split(' ')\n.filter(Boolean)"},
				Options: optsProperty(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotBeforeProperty", Line: 1, Column: 12},
					{MessageId: "expectedDotBeforeProperty", Line: 2, Column: 11},
				},
			},

			// ---- Dimension 4 invalid: NewExpression receiver — prev-token is `)` ----
			{
				Code:    "new Foo()\n.bar",
				Output:  []string{"new Foo().\nbar"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
				},
			},

			// ---- Dimension 4 invalid: CallExpression with type arguments ----
			// Locks in that the TS `<number>` angle brackets inside the
			// receiver expression don't confuse dot detection.
			{
				Code:    "Array.from<number>([1])\n.map(fn)",
				Output:  []string{"Array.from<number>([1]).\nmap(fn)"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
				},
			},

			// ---- Dimension 4 invalid: paren-wrapped await receiver ----
			{
				Code:    "async function f() { return (await foo)\n.bar }",
				Output:  []string{"async function f() { return (await foo).\nbar }"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
				},
			},

			// ---- Dimension 4 invalid: multi-stage `as` chain ----
			// `(a as B as C).prop` — prev-token to dot is `)`.
			{
				Code:    "(a as B as C)\n.prop",
				Output:  []string{"(a as B as C).\nprop"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
				},
			},

			// ---- Dimension 4 invalid: chained NonNullExpression `obj!!\n.prop` ----
			// Two `!` tokens — prev-token-of-dot is still `!`.
			{
				Code:    "obj!!\n.prop",
				Output:  []string{"obj!!.\nprop"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
				},
			},

			// ---- Dimension 4 invalid: nested ImportType in type arguments ----
			// `import('m')\n.A<import('n')\n.B>` — two ImportType nodes; the
			// inner one is INSIDE the outer's TypeArguments. Buffer+sort must
			// emit outer's dot (line 2 col 1) before inner's (line 3 col 1).
			{
				Code:    "type T = import('m')\n.A<import('n')\n.B>",
				Output:  []string{"type T = import('m').\nA<import('n').\nB>"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
					{MessageId: "expectedDotAfterObject", Line: 3, Column: 1},
				},
			},

			// ---- Dimension 4 invalid: ImportType inside Array<...> generic ----
			// Outer is non-listened TypeReference. Inner ImportType + QN
			// both fire — depth counting must flush together.
			{
				Code:    "type T = Array<import('m')\n.A\n.B>",
				Output:  []string{"type T = Array<import('m').\nA.\nB>"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
					{MessageId: "expectedDotAfterObject", Line: 3, Column: 1},
				},
			},

			// ---- Dimension 4 invalid: 5-segment deep chain ----
			// Stress-test depth+sort with all dots on separate lines.
			{
				Code:    "a\n.b\n.c\n.d\n.e",
				Output:  []string{"a.\nb.\nc.\nd.\ne"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
					{MessageId: "expectedDotAfterObject", Line: 3, Column: 1},
					{MessageId: "expectedDotAfterObject", Line: 4, Column: 1},
					{MessageId: "expectedDotAfterObject", Line: 5, Column: 1},
				},
			},

			// ---- Dimension 4 invalid: mixed optional + regular chain ----
			// `a\n?.b\n.c\n?.d\n.e` — verifies `?.` and `.` both detected,
			// AND their report ranges differ (2 chars for `?.`, 1 for `.`).
			{
				Code:    "a\n?.b\n.c\n?.d\n.e",
				Output:  []string{"a?.\nb.\nc?.\nd.\ne"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1, EndLine: 2, EndColumn: 3},
					{MessageId: "expectedDotAfterObject", Line: 3, Column: 1, EndLine: 3, EndColumn: 2},
					{MessageId: "expectedDotAfterObject", Line: 4, Column: 1, EndLine: 4, EndColumn: 3},
					{MessageId: "expectedDotAfterObject", Line: 5, Column: 1, EndLine: 5, EndColumn: 2},
				},
			},

			// ---- Real-user: React this.state chain across lines ----
			// `this.state.value\n.toString()` — common pattern in legacy
			// React class components. PAE chain with newline only on
			// the outermost dot.
			{
				Code:    "class C { x = 1; m() { return this.state.value\n.toString(); } }",
				Output:  []string{"class C { x = 1; m() { return this.state.value.\ntoString(); } }"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
				},
			},

			// ---- Real-user: DOM querySelector chain ----
			{
				Code:    "document.querySelector('.x')\n.addEventListener('click', fn)",
				Output:  []string{"document.querySelector('.x').\naddEventListener('click', fn)"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
				},
			},

			// ---- Real-user: typeof ImportType — Drizzle/Prisma schema pattern ----
			{
				Code:    "type S = typeof import('./schema')\n.schema",
				Output:  []string{"type S = typeof import('./schema').\nschema"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
				},
			},

			// ---- Real-user: Lodash chain (5+ links) ----
			{
				Code:    "_.chain(arr)\n.map(f)\n.filter(g)\n.reduce(h, 0)\n.value()",
				Output:  []string{"_.chain(arr).\nmap(f).\nfilter(g).\nreduce(h, 0).\nvalue()"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
					{MessageId: "expectedDotAfterObject", Line: 3, Column: 1},
					{MessageId: "expectedDotAfterObject", Line: 4, Column: 1},
					{MessageId: "expectedDotAfterObject", Line: 5, Column: 1},
				},
			},

			// ---- Real-user: styled-components-like template tag ----
			// `styled.div\`...\`` — dot in styled.div is on same line, no
			// report. But `styled\n.div\`...\`` would report.
			{
				Code:    "const Btn = styled\n.div`color: red`",
				Output:  []string{"const Btn = styled.\ndiv`color: red`"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
				},
			},

			// ---- Dimension 4 invalid: JSX `<this.X />` with newline ----
			{
				Code:    "const X = <this\n.Component />",
				Output:  []string{"const X = <this.\nComponent />"},
				Options: optsObject(),
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
				},
			},

			// ---- Dimension 4 invalid: paren-wrapped JSX as receiver ----
			{
				Code:    "const x = (<Foo />)\n.props",
				Output:  []string{"const x = (<Foo />).\nprops"},
				Options: optsObject(),
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
				},
			},

			// ---- Dimension 4 invalid: ClassExpression receiver ----
			{
				Code:    "const c = (class {})\n.constructor",
				Output:  []string{"const c = (class {}).\nconstructor"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
				},
			},

			// ---- Dimension 4 invalid: TaggedTemplateExpression as receiver ----
			// PAE.Expression spans the entire `tag\`...\``; prev-token is the
			// closing backtick of the template.
			{
				Code:    "const B = styled.div`x`\n.attrs({})",
				Output:  []string{"const B = styled.div`x`.\nattrs({})"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
				},
			},

			// ---- Dimension 4 invalid: Unicode identifier — column math ----
			// CJK identifiers are 1 UTF-16 code unit per char but multi-byte
			// in UTF-8. Locks in that the rule_tester's column assertion
			// (UTF-16) matches our diagnostic Range positions even across
			// multi-byte sources.
			{
				Code:    "var 变量 = {}; 变量\n.属性",
				Output:  []string{"var 变量 = {}; 变量.\n属性"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1, EndLine: 2, EndColumn: 2},
				},
			},

			// ---- Dimension 4 invalid: SMP surrogate pair on the dot line ----
			// `𝕒` is U+1D552 (4 UTF-8 bytes, 2 UTF-16 code units). The dot
			// follows on a new line, so column position is unambiguous.
			{
				Code:    "var 𝕒: any = {}; 𝕒\n.𝕓",
				Output:  []string{"var 𝕒: any = {}; 𝕒.\n𝕓"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1, EndLine: 2, EndColumn: 2},
				},
			},

			// ---- Dimension 4 invalid: MetaProperty nested as PAE.Expression ----
			// `import.meta\n.url` — only the outer PAE reports (its dot is
			// cross-line); the inner MetaProperty is single-line and stays
			// silent. depth=2 builds inside MetaProperty visit.
			{
				Code:    "const u = import.meta\n.url",
				Output:  []string{"const u = import.meta.\nurl"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
				},
			},
			// And the reverse: MetaProperty dot cross-line, outer PAE dot OK.
			// Locks in that only one diagnostic fires (no double-report).
			{
				Code:    "const u = import\n.meta.url",
				Output:  []string{"const u = import.\nmeta.url"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
				},
			},
			// Both dots cross-line — both report, sorted by source position.
			{
				Code:    "const u = import\n.meta\n.url",
				Output:  []string{"const u = import.\nmeta.\nurl"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
					{MessageId: "expectedDotAfterObject", Line: 3, Column: 1},
				},
			},

			// ---- Dimension 4 invalid: ElementAccess as PAE receiver ----
			{
				Code:    "obj[k]\n.method()",
				Output:  []string{"obj[k].\nmethod()"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
				},
			},

			// ---- Dimension 4 invalid: Conditional in paren ----
			{
				Code:    "(cond ? a : b)\n.method()",
				Output:  []string{"(cond ? a : b).\nmethod()"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
				},
			},

			// ---- Dimension 4 invalid: ArrayLiteralExpression receiver ----
			{
				Code:    "[1, 2]\n.length",
				Output:  []string{"[1, 2].\nlength"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
				},
			},

			// ---- Dimension 4 invalid: ObjectLiteralExpression in paren ----
			{
				Code:    "({a: 1})\n.a",
				Output:  []string{"({a: 1}).\na"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
				},
			},

			// ---- Dimension 4 invalid: RegExpLiteral receiver ----
			{
				Code:    "/foo/\n.test('x')",
				Output:  []string{"/foo/.\ntest('x')"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
				},
			},

			// ---- Dimension 4 invalid: block comment newline pushes dot off-line ----
			// `obj /* a\nb */.x` — the block comment internally crosses a
			// line, so the dot SOURCE line differs from `obj`'s line. ESLint
			// reports `expectedDotAfterObject`. Locks in that our
			// SameLineByPos check uses raw byte positions and correctly
			// detects the cross-line dot through trailing trivia.
			{
				Code:    "obj /* a\nb */.x",
				Output:  []string{"obj. /* a\nb */x"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 5},
				},
			},

			// ---- Dimension 4 invalid: line comment after object ----
			// `obj // c\n.x` — line comment ends at `\n`. dot is on line 2.
			// object mode: dot not on obj's line → report.
			{
				Code:    "obj // c\n.x",
				Output:  []string{"obj. // c\nx"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 2, Column: 1},
				},
			},

			// ---- Dimension 4 invalid: line comment after dot ----
			// `obj.// c\nx` — line comment forces `\n` AFTER dot. dot on
			// line 1, prop on line 2. property mode reports.
			{
				Code:    "obj.// c\nx",
				Output:  []string{"obj// c\n.x"},
				Options: optsProperty(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotBeforeProperty", Line: 1, Column: 4},
				},
			},

			// ---- Real-user regression lock-in: tsgo template-literal
			// scanner re-entry in chained calls ----
			//
			// rsbuild `packages/core/src/plugins/asset.ts:43-47` shape:
			// `rule\n  .oneOf(\`${assetType}-asset-url\`)\n  .type(...)
			//  \n  .resourceQuery(...)\n  .set(...)`. Earlier port
			// scanned from `pae.Expression.Pos()` for prev-token tracking,
			// which forced tsgo's raw scanner to re-tokenize the inner
			// `\`${x}\`` template substitution out-of-context — the
			// closing backtick was mis-classified as the START of a new
			// template literal and the scanner silently swallowed the
			// rest of the chain. Outcome: only the innermost dot got
			// reported, every dot AFTER the template was missed (3
			// silent false-negatives per chain link of this shape).
			// Fix: scan from `Expression.End()` instead, sidestepping
			// the receiver's template scanner state entirely.
			//
			// This lock-in fails on the buggy implementation with
			// 1/4 reports; passes on the fixed implementation with
			// the correct 4/4.
			{
				Code: "declare const rule: any;\n" +
					"declare const assetType: string;\n" +
					"declare const URL_QUERY_REGEX: RegExp;\n" +
					"declare const generatorOptions: any;\n" +
					"rule\n" +
					"  .oneOf(`${assetType}-asset-url`)\n" +
					"  .type('asset/resource')\n" +
					"  .resourceQuery(URL_QUERY_REGEX)\n" +
					"  .set('generator', generatorOptions);",
				Output: []string{"declare const rule: any;\n" +
					"declare const assetType: string;\n" +
					"declare const URL_QUERY_REGEX: RegExp;\n" +
					"declare const generatorOptions: any;\n" +
					"rule.\n" +
					"  oneOf(`${assetType}-asset-url`).\n" +
					"  type('asset/resource').\n" +
					"  resourceQuery(URL_QUERY_REGEX).\n" +
					"  set('generator', generatorOptions);"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 6, Column: 3},
					{MessageId: "expectedDotAfterObject", Line: 7, Column: 3},
					{MessageId: "expectedDotAfterObject", Line: 8, Column: 3},
					{MessageId: "expectedDotAfterObject", Line: 9, Column: 3},
				},
			},

			// ---- Real-user regression lock-in: short variant ----
			// Minimum case that triggers the same template-literal scanner
			// confusion — `r\n.a(\`${x}\`)\n.b()`. Reproduces with one
			// inner-substitution template and two chain links.
			{
				Code: "declare const r: any;\n" +
					"declare const x: string;\n" +
					"r\n" +
					".a(`${x}`)\n" +
					".b()",
				Output: []string{"declare const r: any;\n" +
					"declare const x: string;\n" +
					"r.\n" +
					"a(`${x}`).\n" +
					"b()"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 4, Column: 1},
					{MessageId: "expectedDotAfterObject", Line: 5, Column: 1},
				},
			},

			// ---- Real-user regression lock-in: multi-substitution
			// template + chain ----
			// `r\n.a(\`${x}-${y}\`)\n.b()` — two `${...}` placeholders.
			// Verifies that scanner state doesn't break across multiple
			// substitution boundaries either.
			{
				Code: "declare const r: any;\n" +
					"declare const x: string;\n" +
					"declare const y: string;\n" +
					"r\n" +
					".a(`${x}-${y}`)\n" +
					".b()",
				Output: []string{"declare const r: any;\n" +
					"declare const x: string;\n" +
					"declare const y: string;\n" +
					"r.\n" +
					"a(`${x}-${y}`).\n" +
					"b()"},
				Options: optsObject(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedDotAfterObject", Line: 5, Column: 1},
					{MessageId: "expectedDotAfterObject", Line: 6, Column: 1},
				},
			},
		},
	)
}
