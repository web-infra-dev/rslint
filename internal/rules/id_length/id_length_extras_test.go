// TestIdLengthExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Each case carries an inline comment pointing
// at the specific branch / Dimension 4 row / tsgo AST quirk it covers, so
// future refactors can't silently regress them without breaking a named
// lock-in.
//
// Dimension walk notes for id-length:
//   - Dimension 3 (autofix boundaries): N/A — the rule has no autofix or
//     suggestions.
//   - Dimension 4 access/key forms: element access (X['y'], X[`y`], X[0]) is
//     always ast.KindElementAccessExpression in tsgo, a distinct
//     kind from the KindPropertyAccessExpression this rule dispatches on, so
//     it never reaches isValidMemberExpressionTarget — locked in below.
//   - Dimension 4 declaration/container forms: class-field arrow
//     (`componentDidMount = () => {}`) is a PropertyDeclaration whose
//     Initializer is an ArrowFunction, not itself a name-bearing node — the
//     field's own name is what's checked, same as any other PropertyDeclaration.
package id_length

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestIdLengthExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&IdLengthRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: (X).y / ((X)).y single and multi-level
			// parenthesized receiver — the object side is still checked, but
			// here "long" passes the length check regardless ----
			{Code: `(long).yy = 1;`},
			{Code: `((long)).yy = 1;`},

			// ---- Dimension 4: the TS-only receiver wrappers X!, X as T and
			// X satisfies T survive into typescript-eslint's AST as
			// TSNonNullExpression / TSAsExpression / TSSatisfiesExpression,
			// none of them a supported parent, so the object identifier under
			// one is never checked no matter how short it is ----
			{Code: `(a!).yy = 1;`},
			{Code: `(a as any).yy = 1;`},
			{Code: `(a satisfies unknown).yy = 1;`},
			{Code: `[...a!] = b;`},

			// ---- A `!` around a whole destructuring target is opaque to the
			// pattern conversion too: typescript-eslint hands a
			// TSNonNullExpression's operand to the ordinary expression
			// conversion, so the literal stays an ArrayExpression /
			// ObjectExpression and none of the unconditional ArrayPattern /
			// AssignmentPattern / RestElement entries can match inside it ----
			{Code: `([x]! = source);`},
			{Code: `([x = 0]! = source);`},
			{Code: `({ key: x = 0 }! = source);`},
			{Code: `({ key: x }! = source);`},
			{Code: `([...x]! = source);`},
			{Code: `({ ...x }! = source);`},

			// ---- ...at whatever depth it sits, since the climb stops at the
			// first one it crosses ----
			{Code: `([[x]!] = source);`},
			{Code: `([([x]!)] = source);`},
			{Code: `(([x]!)! = source);`},
			{Code: `([...[x]]! = source);`},

			// ---- ...and the value side of a genuine pattern is just as
			// opaque: the wrapper, not the identifier inside it, is what the
			// Property/ArrayPattern entry would have to match ----
			{Code: `({ key: [x]! } = source);`},
			{Code: `({ key: x! } = source);`},

			// ---- A member expression under a rest element is neither a
			// direct assignment target nor an ObjectPattern property value ----
			{Code: `({ ...obj.x } = source);`},

			// ---- A string-literal key has no `name` for upstream's
			// non-pattern Property branch to compare the value against ----
			{Code: `const obj = { 'x': x };`},
			{Code: `const obj = { 1: x };`},

			// ---- Dimension 4: the same wrapper around the destructuring
			// value side of an ObjectPattern-equivalent member-expression
			// target ----
			{Code: `({ a: (obj.x as any) } = {});`},
			{Code: `class Foo { value = x as unknown }`},

			// ---- Dimension 4: element access X['y'] / X[`y`] / X[0] is
			// never a match — element access is a distinct AST kind from the
			// dotted access this rule inspects ----
			{Code: `a['b'] = 1;`},
			{Code: "a[`b`] = 1;"},
			{Code: `a[0] = 1;`},
			{Code: `({ a: obj['x'] } = {});`},

			// ---- Dimension 4: optional chain can never be an assignment
			// target (SyntaxError), so `a?.b` as a plain read is never
			// checked as a member-expression target regardless of length ----
			{Code: `a?.b;`},

			// ---- Dimension 4: class expression name is not checked
			// (upstream has no ClassExpression entry in SUPPORTED_EXPRESSIONS) ----
			{Code: `const longName = class q {};`},
			{Code: `const longName = class {};`},

			// ---- Dimension 4: class-field arrow function — the field name
			// is checked, the arrow's own (implicit, absent) name is not ----
			{Code: `class Foo { longName = () => {}; }`},

			// ---- Dimension 4: async / generator / async-generator method
			// and function variants all route through the same name check ----
			{Code: `async function longName() {}`},
			{Code: `function* longName() {}`},
			{Code: `async function* longName() {}`},
			{Code: `class Foo { async longName() {} }`},
			{Code: `class Foo { *longName() {} }`},
			{Code: `class Foo { async *longName() {} }`},

			// ---- Dimension 4: nesting boundary — inner class's own name is
			// independently checked, doesn't bleed from/into the outer class ----
			{Code: `class Outer { method() { class Inner {} return Inner; } }`},

			// ---- Dimension 4: nesting boundary — a short parameter name in
			// a nested function does not leak a false match from the outer
			// function's differently-shaped parameter list ----
			{Code: `function outer(longOne) { return function (longTwo) { return longOne + longTwo; }; }`},

			// ---- Dimension 4: graceful degradation — empty class body,
			// empty function body, empty destructuring pattern, empty
			// arguments list must not crash ----
			{Code: `class Empty {}`},
			{Code: `function empty() {}`},
			{Code: `var {} = {};`},
			{Code: `var [] = [];`},
			{Code: `empty();`},

			// ---- Dimension 4: graceful degradation — TS overload signature
			// and abstract/declare members (body-absent forms) must not
			// crash ----
			{Code: `declare function longName(xx: number): void;`},
			{Code: `abstract class Foo { abstract longName(): void; }`},
			{Code: `declare class Foo { longName: number; }`},

			// ---- ...and a body-less function declaration is a
			// TSDeclareFunction in ESTree rather than a FunctionDeclaration,
			// so its name is not checked however short it is ----
			{Code: `declare function q(xx: number): void;`},
			{Code: `declare namespace NN { function q(): void; }`},

			// ---- Branch lock-in: BindingElement's isKeyAndValueSame==true
			// arm — the same isValidPatternPropertyValueNodes implementation
			// PropertyAssignment's pattern branch also uses — requires
			// `properties` enabled; only the key side is reported ----
			{Code: `var { a: a } = {};`, Options: map[string]any{"properties": "never"}},

			// ---- ...and the same arm reached through an assignment
			// destructuring's PropertyAssignment rather than a binding ----
			{Code: `({ a: a } = {});`, Options: map[string]any{"properties": "never"}},
			{Code: `({ longName: longName } = {});`},

			// ---- Branch lock-in: computed key on a BindingElement (real
			// destructuring binding) is excluded, same as a plain object
			// literal's computed key ----
			{Code: `function ff({ [a]: prop }) {}`},

			// ---- Real-user: array-destructured (x, y) coordinate pair
			// pulled out of a tuple-returning helper, a common false-positive
			// report against id-length's default min:2 ----
			{Code: `const [longX, longY] = getCoordinates();`},

			// ---- Real-user: short loop counters exempted via the
			// documented `exceptions` escape hatch ----
			{Code: `for (let i = 0; i < items.length; i++) { use(items[i]); }`, Options: map[string]any{"exceptions": []any{"i"}}},

			// ---- exceptionPatterns is compiled as `new RegExp(pattern, "u")`,
			// so a pattern spelling syntax only JavaScript has still exempts
			// what it matches: lookahead, lookbehind and a `\p{...}` property
			// escape (none of which Go's RE2 can compile at all) ----
			{Code: `var x = 1;`, Options: map[string]any{"exceptionPatterns": []any{"(?=x)"}}},
			{Code: `var x = 1;`, Options: map[string]any{"exceptionPatterns": []any{"(?<=^)x"}}},
			{Code: `var e = 1;`, Options: map[string]any{"exceptionPatterns": []any{"^\\p{Ll}$"}}},

			// ---- A TS parameter property is a TSParameterProperty in
			// ESTree, which is not a supported parent, so the binding name
			// goes unchecked no matter which modifiers spell it ----
			{Code: `class Foo { constructor(private x: number) {} }`},
			{Code: `class Foo { constructor(public readonly x: number) {} }`},
			{Code: `class Foo { protected constructor(readonly x: number) {} }`},

			// ---- An object literal's method/accessor name is a Property
			// key, so `properties: never` exempts it (a class member's name
			// is a MethodDefinition and stays checked) ----
			{Code: `const obj = { x() {} };`, Options: map[string]any{"properties": "never"}},
			{Code: `const obj = { get x() { return 1; } };`, Options: map[string]any{"properties": "never"}},
			{Code: `const obj = { set x(v2) {} };`, Options: map[string]any{"properties": "never"}},
			{Code: `const obj = { async *x() {} };`, Options: map[string]any{"properties": "never"}},

			// ---- ...and a computed one is never checked at all, matching
			// Property's `!parent.computed` guard ----
			{Code: `const obj = { [x]() {} };`},
			{Code: `const obj = { get [x]() { return 1; } };`},

			// ---- A method key inside a dynamic import's options object is
			// still an import attribute key ----
			{Code: `import('foo.json', { with: { a() {} } });`},

			// ---- A plain parameter is a direct child of the function node
			// upstream, so it is only checked when that node is one of the
			// supported ones. Every TS signature-only function-like — function
			// and constructor types, call/construct/method/index signatures,
			// ambient and overload declarations, abstract methods — becomes a
			// TS-specific ESTree node that is not in SUPPORTED_EXPRESSIONS ----
			{Code: `type Fn = (x: number) => void;`},
			{Code: `type Ct = new (x: number) => void;`},
			{Code: `declare function fn(x: number): void;`},
			{Code: `interface II { new (x: number): II }`},
			{Code: `interface II { mm(x: number): void }`},
			{Code: `interface II { (x: number): void }`},
			{Code: `interface II { [x: string]: number }`},
			{Code: `abstract class Foo { abstract mm(x: number): void; }`},
			{Code: `declare class Foo { constructor(x: number); }`},

			// ---- An arrow's concise body only matches when it is a bare
			// identifier: a block body, a member expression and a TS wrapper
			// all put an unsupported parent in between ----
			{Code: `const fn = (arg) => { x };`},
			{Code: `const fn = (arg) => x.y;`},
			{Code: `const fn = (arg) => x as any;`},

			// ---- A plain assignment expression is not a supported parent;
			// only the AssignmentPattern a destructuring target turns it into
			// is ----
			{Code: `x = 0;`},

			// ---- A computed key in a genuine object literal exempts both
			// halves of the property: upstream's non-pattern Property branch
			// guards on `!parent.computed` ----
			{Code: `const obj = { [x]: x };`},

			// ---- ...while a computed key in a destructuring pattern is
			// compared against the value like any other key, so
			// `properties: never` exempts the matching pair entirely ----
			{Code: `const {[x]: x} = obj;`, Options: map[string]any{"properties": "never"}},

			// ---- `obj.x = 0` inside a destructuring target is an
			// AssignmentPattern holding a MemberExpression, not the
			// AssignmentExpression upstream's MemberExpression entry requires ----
			{Code: `({ key: obj.x = 0 } = source);`},
			{Code: `([obj.x = 0] = source);`},
			{Code: `for ({ key: obj.x = 1 } of source) {}`},

			// ---- A for-in/of target that is a member expression on its own
			// fails the same AssignmentExpression test... ----
			{Code: `for (obj.x of source) {}`},

			// ---- ...and an object pattern's member target is gated on
			// `properties` like every other MemberExpression match ----
			{Code: `for ({ key: obj.x } of source) {}`, Options: map[string]any{"properties": "never"}},

			// ---- Parenthesizing the object literal stops upstream's parser
			// from converting it to an ObjectPattern at all ----
			{Code: `(({ key: obj.x }) = source);`},

			// ---- Parentheses around a dynamic import's options object, or
			// around the `with` value, are absent from ESLint's AST, so the
			// attribute key stays exempt ----
			{Code: `import("m", ({ with: { x: "json" } }));`},
			{Code: `import("m", { with: ({ x: "json" }) });`},

			// ---- Interface members are TSMethodSignature-shaped in ESTree,
			// so neither a `constructor` member nor a short one is checked ----
			{Code: `interface II { constructor(): void }`, Options: map[string]any{"min": 0, "max": 5}},
			{Code: `interface II { x(): void }`},

			// ---- `constructor` is exempt like any other name ----
			{Code: `class Foo { constructor() {} }`, Options: map[string]any{"min": 0, "max": 5, "exceptions": []any{"constructor"}}},
			{Code: `class Foo { constructor() {} }`, Options: map[string]any{"min": 0, "max": 5, "exceptionPatterns": []any{"^construct"}}},

			// ---- An object literal's `constructor` method is a Property
			// key, so `properties: never` exempts it ----
			{Code: `const o2 = { constructor() {} };`, Options: map[string]any{"min": 0, "max": 5, "properties": "never"}},

			// ---- An `abstract` or `accessor` class member is a
			// TSAbstract*/AccessorProperty node in ESTree rather than a
			// MethodDefinition/PropertyDefinition, so its name is never
			// checked ----
			{Code: `abstract class Cc { abstract x(): void }`},
			{Code: `abstract class Cc { abstract x: number }`},
			{Code: `abstract class Cc { abstract get x(): number }`},
			{Code: `abstract class Cc { abstract [y]: number }`},
			{Code: `class Cc { accessor x = 1; }`},

			// ---- Only a class declaration's own `extends` clause is a
			// direct ClassDeclaration child upstream: `implements` becomes a
			// TSClassImplements, an interface's `extends` a TSInterfaceHeritage,
			// and a class expression's superclass hangs off the unsupported
			// ClassExpression ----
			{Code: `class Cc implements x {}`},
			{Code: `interface Ii extends x {}`},
			{Code: `const Cc = class extends x {};`},
			{Code: `const Cc = class Dd extends x {};`},

			// ---- ...and a superclass that is not a bare identifier puts an
			// unsupported node between the two either way ----
			{Code: `class Cc extends x! {}`},
			{Code: `class Cc extends (x as any) {}`},
			{Code: `class Cc extends obj.x {}`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: (X).y parenthesized receiver — both the
			// object and property identifiers of a direct assignment target
			// are checked, matching upstream's checker function taking only
			// the MemberExpression itself (see id_length.go doc comment) ----
			{
				Code: `(a).b = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'a' is too short (< 2).", Line: 1, Column: 2},
					{MessageId: "tooShort", Message: "Identifier name 'b' is too short (< 2).", Line: 1, Column: 5},
				},
			},
			// ---- Dimension 4: ((X)).y multi-level parenthesized receiver ----
			{
				Code: `((a)).b = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'a' is too short (< 2).", Line: 1, Column: 3},
					{MessageId: "tooShort", Message: "Identifier name 'b' is too short (< 2).", Line: 1, Column: 7},
				},
			},
			// ---- Dimension 4: X!.y non-null assertion receiver — the
			// property is still checked, but the object identifier now sits
			// under a TSNonNullExpression, which is not a supported parent ----
			{
				Code: `(a!).b = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'b' is too short (< 2).", Line: 1, Column: 6},
				},
			},
			// ---- Dimension 4: (X as any).y type-assertion receiver ----
			{
				Code: `(a as any).b = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'b' is too short (< 2).", Line: 1, Column: 12},
				},
			},
			// ---- Dimension 4: X satisfies T type-expression receiver ----
			{
				Code: `(a satisfies unknown).b = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'b' is too short (< 2).", Line: 1, Column: 23},
				},
			},
			// ---- Dimension 4: a parenthesized bare-identifier destructuring
			// value (`{ a: (b) }`) still matches PropertyAssignment's pattern
			// value comparison, which compares against the wrapper-climbed
			// node rather than the raw Initializer field ----
			{
				Code: `({ a: (b) } = {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'b' is too short (< 2).", Line: 1, Column: 8},
				},
			},
			// ---- Dimension 4: a parenthesized class-field initializer is
			// still matched by PropertyDeclaration's unconditional
			// Name()-or-Initializer check ----
			{
				Code: `class Foo { xx = (q); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 1, Column: 19},
				},
			},
			// ---- Dimension 4: a parenthesized rest target in assignment
			// destructuring is still recognized as an assignment target —
			// ast.IsAssignmentTarget must be invoked on the wrapper-climbed
			// node, not on the raw SpreadElement/SpreadAssignment itself
			// (GetAssignmentTarget never special-cases ObjectLiteralExpression
			// directly; it is only reachable by starting the climb from the
			// identifier through the Spread wrapper) ----
			{
				Code: `[...(q)] = arr;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 1, Column: 6},
				},
			},
			{
				Code: `({ ...(q) } = {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 1, Column: 8},
				},
			},

			// ---- Dimension 4: async/generator/async-generator variants all
			// still check the declared name ----
			{
				Code:   `async function q() {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 1, Column: 16}},
			},
			{
				Code:   `function* q() {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 1, Column: 11}},
			},
			{
				Code:   `async function* q() {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 1, Column: 17}},
			},
			{
				Code:   `class Foo { async q() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 1, Column: 19}},
			},
			{
				Code:   `class Foo { *q() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 1, Column: 14}},
			},

			// ---- Dimension 4: nesting boundary — a same-kind class nested
			// inside another class member reports its own short name
			// independently of the outer class's (longer) name ----
			{
				Code:   `class Outer { method() { class q {} return q; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 1, Column: 32}},
			},

			// ---- ...while the implementation that follows an overload
			// signature is a real FunctionDeclaration, so only its name is
			// reported ----
			{
				Code: `function q(xx: number): void;
function q(xx: any) { return xx; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 2, Column: 10},
				},
			},

			// ---- Dimension 4: graceful degradation — an ambient class
			// member is a body-absent form that keeps its
			// PropertyDefinition/MethodDefinition shape, so its name is
			// still checked ----
			{
				Code:   `declare class Foo { q: number; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 1, Column: 21}},
			},
			{
				Code:   `declare class Foo { q(): void; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 1, Column: 21}},
			},
			{
				Code: `class Foo { q(); q() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 1, Column: 13},
					{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 1, Column: 18},
				},
			},

			// ---- Branch lock-in: rest element in assignment-destructuring
			// (SpreadElement / SpreadAssignment write context via
			// ast.IsAssignmentTarget) is checked; the same shape in a plain
			// (read) spread position is not (see valid case for the object
			// spread contrast) ----
			{
				Code:   `[...q] = arr;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 1, Column: 5}},
			},
			{
				Code:   `({ ...q } = {});`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 1, Column: 7}},
			},

			// ---- Branch lock-in: import attribute key is exempt even when
			// it independently violates min/max — locks in isImportAttributeKey
			// for the static form, since upstream's own test only exercises
			// it within a min:1,max:3 config where "type" happens to fit ----
			{
				Code:    `import longName from 'foo.json' with { longAttributeName: 'json' };`,
				Options: map[string]any{"max": 5},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooLong", Message: "Identifier name 'longName' is too long (> 5)."},
				},
			},

			// ---- Real-user: array-destructured (x, y) coordinate pair
			// too short under a stricter min, a common false-positive
			// report shape against id-length ----
			{
				Code: `const [x, y] = getCoordinates();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 8},
					{MessageId: "tooShort", Message: "Identifier name 'y' is too short (< 2).", Line: 1, Column: 11},
				},
			},

			// ---- A backreference — also RE2-incompatible — exempts the
			// identifier it matches and leaves the rest reported ----
			{
				Code:    `var xx = 1; var ab = 2;`,
				Options: map[string]any{"min": 3, "exceptionPatterns": []any{`^(.)\1$`}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'ab' is too short (< 3).", Line: 1, Column: 17},
				},
			},

			// ---- Real-user: event-handler callback parameter named `e`,
			// one of the most commonly reported id-length false positives ----
			{
				Code: `el.addEventListener('click', function (e) { use(e); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'e' is too short (< 2).", Line: 1, Column: 40},
				},
			},

			// ---- A default value on a parameter property restores the
			// AssignmentPattern that ESTree checks unconditionally, so the
			// binding name is back in scope for the rule ----
			{
				Code: `class Foo { constructor(private x = 1) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 33},
				},
			},

			// ---- An object literal's method name is checked when
			// `properties` is on (the default) ----
			{
				Code: `const obj = { x() {} };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 15},
				},
			},

			// ---- ...while a class member's name ignores `properties`
			// entirely, computed or not ----
			{
				Code:    `class C { x() {} }`,
				Options: map[string]any{"properties": "never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'C' is too short (< 2).", Line: 1, Column: 7},
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 11},
				},
			},
			{
				Code:    `class C { get [x]() { return 1; } }`,
				Options: map[string]any{"properties": "never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'C' is too short (< 2).", Line: 1, Column: 7},
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 16},
				},
			},

			// ---- A class constructor is a MethodDefinition whose key is an
			// Identifier upstream, so its name counts against `max`; tsgo has
			// no name node there, so the report lands on the keyword ----
			{
				Code:    `class Foo { constructor() {} }`,
				Options: map[string]any{"min": 0, "max": 5},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooLong", Message: "Identifier name 'constructor' is too long (> 5).", Line: 1, Column: 13, EndColumn: 24},
				},
			},
			{
				Code:    `class Foo { public constructor() {} }`,
				Options: map[string]any{"min": 0, "max": 5},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooLong", Message: "Identifier name 'constructor' is too long (> 5).", Line: 1, Column: 20, EndColumn: 31},
				},
			},
			{
				Code:    `const Cls = class { constructor() {} };`,
				Options: map[string]any{"min": 0, "max": 5},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooLong", Message: "Identifier name 'constructor' is too long (> 5).", Line: 1, Column: 21, EndColumn: 32},
				},
			},

			// ---- ...including an overload signature, which is its own
			// MethodDefinition ----
			{
				Code:    `class C2 { constructor(); constructor(a?: number) {} }`,
				Options: map[string]any{"min": 0, "max": 5},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooLong", Message: "Identifier name 'constructor' is too long (> 5).", Line: 1, Column: 12},
					{MessageId: "tooLong", Message: "Identifier name 'constructor' is too long (> 5).", Line: 1, Column: 27},
				},
			},

			// ---- An object literal's `constructor` method is an ordinary
			// property key ----
			{
				Code:    `const o2 = { constructor() {} };`,
				Options: map[string]any{"min": 0, "max": 5},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooLong", Message: "Identifier name 'constructor' is too long (> 5).", Line: 1, Column: 14},
				},
			},

			// ---- A defaulted or rest binding is an AssignmentPattern /
			// RestElement upstream, not a Property, so `properties: never`
			// does not exempt it ----
			{
				Code:    `const { x = 1 } = obj;`,
				Options: map[string]any{"properties": "never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 9},
				},
			},
			{
				Code:    `const { ...x } = obj;`,
				Options: map[string]any{"properties": "never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 12},
				},
			},
			{
				Code:    `function ff({ x = 1 }) {}`,
				Options: map[string]any{"properties": "never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 15},
				},
			},

			// ---- ...while an ordinary (non-abstract) member of the same
			// abstract class stays checked ----
			{
				Code: `abstract class Cc { declare x: number; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 29},
				},
			},

			// ---- ...and a renamed-with-default binding still reports only
			// the local name, never the property name it came from ----
			{
				Code:    `const { longName: b = 1 } = obj;`,
				Options: map[string]any{"properties": "never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'b' is too short (< 2).", Line: 1, Column: 19},
				},
			},

			// ---- The same holds for a destructuring assignment, where the
			// default lives on a ShorthandPropertyAssignment ----
			{
				Code:    `({ x = 1 } = obj);`,
				Options: map[string]any{"properties": "never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 4},
				},
			},

			// ---- A rest parameter keeps its own RestElement wrapper, an
			// unconditional entry, so it is checked even in the TS
			// signature-only function-likes whose plain parameters are not ----
			{
				Code: `type Fn = (...x: number[]) => void;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 15},
				},
			},
			{
				Code: `declare function fn(...x: number[]): void;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 24},
				},
			},
			{
				Code: `interface II { mm(...x: number[]): void }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 22},
				},
			},

			// ---- An overload signature is body-less (TSEmptyBodyFunction-
			// Expression), so only the implementation's parameter is checked ----
			{
				Code: `class Foo { mm(x: number): void; mm(x: any) { return x; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 37},
				},
			},

			// ---- The other half of upstream's unconditional
			// ArrowFunctionExpression entry: a concise body that is a bare
			// identifier is a direct child of the arrow ----
			{
				Code: `const fn = (arg) => x;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 21},
				},
			},
			{
				Code: `const fn = (arg) => (x);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 22},
				},
			},

			// ---- An array literal that is a destructuring-assignment target
			// is an ArrayPattern upstream, whose entry is unconditional ----
			{
				Code: `([x] = source);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 3},
				},
			},
			{
				Code: `([x = 0] = source);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 3},
				},
			},
			{
				Code: `for ([x] of source) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 7},
				},
			},

			// ---- A default inside a destructuring target is an
			// AssignmentPattern, an unconditional entry, so `properties` does
			// not exempt it ----
			{
				Code: `({ key: x = 0 } = source);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 9},
				},
			},
			{
				Code:    `({ key: x = 0 } = source);`,
				Options: map[string]any{"properties": "never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 9},
				},
			},

			// ---- A computed key in a destructuring pattern still goes
			// through the ordinary key/value comparison, so a key matching the
			// value reports the key... ----
			{
				Code: `const {[x]: x} = obj;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 9},
				},
			},
			{
				Code: `({[x]: x} = obj);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 4},
				},
			},

			// ---- ...and a key that differs reports only the value, whether
			// or not that value carries a default ----
			{
				Code: `const {[x]: y} = obj;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'y' is too short (< 2).", Line: 1, Column: 13},
				},
			},
			{
				Code: `const {[x]: y = 1} = obj;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'y' is too short (< 2).", Line: 1, Column: 13},
				},
			},

			// ---- An object pattern's member target in a for-in/of head:
			// upstream reads `parent.parent.parent.parent.left`, and
			// ForInStatement/ForOfStatement carry a `left` just as
			// AssignmentExpression does ----
			{
				Code: `for ({ key: obj.x } of source) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 17},
				},
			},
			{
				Code: `for ({ key: obj.x } in source) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 17},
				},
			},

			// ---- A parenthesized property value in a genuine object literal
			// is compared by the identifier's own text, not the wrapper's, so
			// a value spelled like its key is reported alongside it ----
			{
				Code: `const obj = { x: (x) };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 15},
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 19},
				},
			},

			// ---- The other side of the same `!` coin: what the wrapper
			// leaves behind is a genuine object literal, so its property goes
			// through the non-pattern Property branch, which compares the
			// key's text against every identifier of the property ----
			{
				Code: `({ q: q }! = source);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 1, Column: 4},
					{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 1, Column: 7},
				},
			},

			// ---- ...and `obj.q = 0` under one is an AssignmentExpression
			// again rather than the AssignmentPattern a real pattern would
			// have made of it, which is exactly what upstream's
			// MemberExpression entry requires ----
			{
				Code: `({ key: obj.q = 0 }! = source);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 1, Column: 13},
				},
			},
			{
				Code: `([obj.q = 0]! = source);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 1, Column: 7},
				},
			},

			// ---- A shorthand property is checked either way — the
			// ObjectPattern and non-pattern branches converge on the same
			// `properties` gate for it ----
			{
				Code: `({ q }! = source);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 1, Column: 4},
				},
			},

			// ---- Only the target the `!` wraps is spared; a sibling
			// element of the same pattern still matches ----
			{
				Code: `([q, [q]!] = source);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 1, Column: 3},
				},
			},

			// ---- ...and everything below the wrapper reverts to ordinary
			// object-literal semantics, nested members included ----
			{
				Code: `({ key: { x } }! = source);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 11},
				},
			},

			// ---- A string-literal key in a destructuring pattern can never
			// equal the value's name, so the value is what gets reported ----
			{
				Code: `({ 'x': x } = source);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 9},
				},
			},
			{
				Code: `let { 'x': x } = source;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 12},
				},
			},

			// ---- A private member is checked at its use site too, since
			// upstream's MemberExpression entry takes PrivateIdentifier
			// properties like any other ----
			{
				Code: `class Cc { #x; mm() { this.#x = 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShortPrivate", Message: "Identifier name '#x' is too short (< 2).", Line: 1, Column: 12},
					{MessageId: "tooShortPrivate", Message: "Identifier name '#x' is too short (< 2).", Line: 1, Column: 28},
				},
			},

			// ---- `this` and `super` receivers are keywords rather than
			// identifiers, so only the property side is left to check ----
			{
				Code: `this.x = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 6},
				},
			},
			{
				Code: `class Cc { mm() { super.x = 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 25},
				},
			},

			// ---- An assignment destructuring's key is an ordinary
			// expression the linter visits like any other, so a key spelled
			// like its value is reported on the key side — the same arm
			// `var { a: a } = {}` takes ----
			{
				Code: `({ a: a } = {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'a' is too short (< 2).", Line: 1, Column: 4},
				},
			},
			{
				Code: `({ a: a, b: c } = {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'a' is too short (< 2).", Line: 1, Column: 4},
					{MessageId: "tooShort", Message: "Identifier name 'c' is too short (< 2).", Line: 1, Column: 13},
				},
			},
			{
				Code: `([{ a: a }] = {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'a' is too short (< 2).", Line: 1, Column: 5},
				},
			},
			{
				Code: `for ({ a: a } of s) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'a' is too short (< 2).", Line: 1, Column: 8},
				},
			},

			// ---- A logical or compound assignment is an AssignmentExpression
			// like any other, so its member target still matches ----
			{
				Code: `obj.x ??= 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 5},
				},
			},

			// ---- A class declaration's superclass expression is as direct a
			// ClassDeclaration child as its name upstream, and that entry is
			// unconditional, so a bare-identifier superclass is checked ----
			{
				Code: `class Cc extends x {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 18},
				},
			},

			// ---- ...through parentheses, past type arguments, and alongside
			// an `implements` clause that stays exempt ----
			{
				Code: `class Cc extends (x) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 19},
				},
			},
			{
				Code: `class Cc extends x<number> {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 18, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code: `class Cc extends x implements yy {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 18},
				},
			},

			// ---- ...and for every shape tsgo still parses as a class
			// declaration ----
			{
				Code: `abstract class Aa extends x {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 27},
				},
			},
			{
				Code: `declare class Dd extends x {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 26},
				},
			},
			{
				Code: `export default class extends x {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 30},
				},
			},

			// ---- A variable declarator's or a parameter's type annotation
			// is folded into the name Identifier's own range upstream, so the
			// report spans `q: number`, not just `q` ----
			{
				Code: `function fn(q: number) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 1, Column: 13, EndLine: 1, EndColumn: 22},
				},
			},
			{
				Code: `const fn = function (q: number) {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 1, Column: 22, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code: `const fn = (q: number) => {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 1, Column: 13, EndLine: 1, EndColumn: 22},
				},
			},
			{
				Code: `class Cc { mm(q: number) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 1, Column: 15, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code: `class Cc { constructor(q: number) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 1, Column: 24, EndLine: 1, EndColumn: 33},
				},
			},
			{
				Code: `const fn = { mm(q?: number) {} };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 1, Column: 17, EndLine: 1, EndColumn: 27},
				},
			},

			// ---- ...including a defaulted parameter, whose annotation stays
			// on the AssignmentPattern's left ----
			{
				Code: `function fn(q: number = 1) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 1, Column: 13, EndLine: 1, EndColumn: 22},
				},
			},

			// ---- An optional parameter absorbs its `?` the same way, with or
			// without an annotation to follow it ----
			{
				Code: `function fn(q?) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 1, Column: 13, EndLine: 1, EndColumn: 15},
				},
			},

			// ---- A declarator's annotation reaches past a definite
			// assignment `!`, the only place that token can appear ----
			{
				Code: `const q: number = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 1, Column: 7, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code: `let q!: number;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 1, Column: 5, EndLine: 1, EndColumn: 15},
				},
			},

			// ---- A catch binding is a declarator upstream too ----
			{
				Code: `try {} catch (q: unknown) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 1, Column: 15, EndLine: 1, EndColumn: 25},
				},
			},

			// ---- Trailing trivia is not part of the annotation's range, but
			// trivia inside it is ----
			{
				Code: `function fn(q: /*c*/ number) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 1, Column: 13, EndLine: 1, EndColumn: 28},
				},
			},

			// ---- ...while a rest parameter's annotation belongs to its
			// RestElement wrapper, a destructured parameter's to the pattern,
			// and a class field's to the PropertyDefinition, so none of those
			// widens the identifier ----
			{
				Code: `function fn(...q: number[]) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 1, Column: 16, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code: `function fn({ q }: { q: number }) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 1, Column: 15, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code: `class Cc { q: number = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: `class Cc { #q: number = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShortPrivate", Message: "Identifier name '#q' is too short (< 2).", Line: 1, Column: 12, EndLine: 1, EndColumn: 14},
				},
			},
		},
	)
}
