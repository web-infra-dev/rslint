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

			// ---- Dimension 4: X!.y non-null assertion receiver ----
			{Code: `(long!).yy = 1;`},

			// ---- Dimension 4: (X as any).y / X satisfies T type-expression
			// receiver wrappers ----
			{Code: `(long as any).yy = 1;`},
			{Code: `(long satisfies unknown).yy = 1;`},

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
			// and abstract/declare members (body-absent forms) must not crash,
			// and their (short) names are still checked through the same
			// MethodDeclaration/PropertyDeclaration paths ----
			{Code: `declare function longName(xx: number): void;`},
			{Code: `abstract class Foo { abstract longName(): void; }`},
			{Code: `declare class Foo { longName: number; }`},

			// ---- Branch lock-in: BindingElement's isKeyAndValueSame==true
			// arm — the same isValidPatternPropertyValueNodes implementation
			// PropertyAssignment's pattern branch also uses — requires
			// `properties` enabled; only the key side is reported ----
			{Code: `var { a: a } = {};`, Options: map[string]any{"properties": "never"}},

			// ---- Differences from ESLint: a non-computed PropertyAssignment
			// key inside an object literal used as a destructuring-assignment
			// target is never visited by the linter's traversal unless the
			// key and value differ (see id_length.go's KindPropertyAssignment
			// case comment) — so unlike ESLint, `{ a: a }` here is not
			// flagged even with the default `properties: "always"`. The
			// equivalent real declaration (`var { a: a } = {}`, covered in
			// the upstream suite) is unaffected and still flags the key ----
			{Code: `({ a: a } = {});`},
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
			// ---- Dimension 4: X!.y non-null assertion receiver ----
			{
				Code: `(a!).b = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'a' is too short (< 2).", Line: 1, Column: 2},
					{MessageId: "tooShort", Message: "Identifier name 'b' is too short (< 2).", Line: 1, Column: 6},
				},
			},
			// ---- Dimension 4: (X as any).y type-assertion receiver ----
			{
				Code: `(a as any).b = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'a' is too short (< 2).", Line: 1, Column: 2},
					{MessageId: "tooShort", Message: "Identifier name 'b' is too short (< 2).", Line: 1, Column: 12},
				},
			},
			// ---- Dimension 4: X satisfies T type-expression receiver ----
			{
				Code: `(a satisfies unknown).b = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'a' is too short (< 2).", Line: 1, Column: 2},
					{MessageId: "tooShort", Message: "Identifier name 'b' is too short (< 2).", Line: 1, Column: 23},
				},
			},
			// ---- Dimension 4: (X as any).y wrapping the destructuring
			// value side of an ObjectPattern-equivalent member-expression
			// target ----
			{
				Code: `({ a: (obj.x as any) } = {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'x' is too short (< 2).", Line: 1, Column: 12},
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

			// ---- Dimension 4: graceful degradation — TS overload signature
			// and abstract/declare members (body-absent forms) are still
			// checked through the ordinary name paths ----
			{
				Code:   `declare function q(xx: number): void;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 1, Column: 18}},
			},
			{
				Code:   `abstract class Foo { abstract q(): void; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 1, Column: 31}},
			},
			{
				Code:   `declare class Foo { q: number; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "tooShort", Message: "Identifier name 'q' is too short (< 2).", Line: 1, Column: 21}},
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
			// ---- Real-user: event-handler callback parameter named `e`,
			// one of the most commonly reported id-length false positives ----
			{
				Code: `el.addEventListener('click', function (e) { use(e); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooShort", Message: "Identifier name 'e' is too short (< 2).", Line: 1, Column: 40},
				},
			},
		},
	)
}
