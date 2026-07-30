package no_useless_assignment

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoUselessAssignmentExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it covers,
// so future refactors can't silently regress them without breaking a named
// lock-in. The upstream 1:1 migration lives in
// no_useless_assignment_upstream_test.go.
//
// N/A: Dimension 3 (autofix boundaries) — the rule reports only, so there is no
// fix or suggestion and therefore no TestNoUselessAssignmentEditDemand either.
func TestNoUselessAssignmentExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUselessAssignmentRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: parenthesized read of the assigned value ----
			{Code: `function f() { let v = 1; g((v)); v = 2; g(v); }`},
			{Code: `function f() { let v = 1; g(((v))); v = 2; g(v); }`},

			// ---- Dimension 4: TS type-expression wrappers around a read ----
			{Code: `function f() { let v = 1; g(v as unknown); v = 2; g(v); }`},
			{Code: `function f() { let v: number | null = 1; g(v!); v = 2; g(v); }`},
			// ---- Real-user: eslint/eslint#19002 (`satisfies` hides the read) ----
			{Code: `const foo = 'foo';
if (true) {}
([foo] satisfies Array<string>).forEach(str => { console.log(str); });`},

			// ---- Dimension 4: optional chain skips the assignment ----
			{Code: `function f() { let v = 1; a?.m(v = 2); g(v); }`},
			{Code: `function f() { let v = 1; a?.b[v = 2]; g(v); }`},
			{Code: `function f() { let v = 1; a?.b?.c(v = 2); g(v); }`},

			// ---- Dimension 4: computed destructuring key is read before the write ----
			{Code: `function f() { let v = 1, k = 'a'; g(v, k); ({ [v]: k } = o); g(k); }`},
			// ---- Dimension 4: string-literal key in a declaration pattern ----
			{Code: `function f() { let { 'k': v } = o; g(v); }`},

			// ---- Dimension 4: class expression field initializer is its own code path ----
			{Code: `let y = 1; const C = class { p = (y = 2); }; g(y);`},
			// ---- Dimension 4: class-field arrow only writes from another code path ----
			{Code: `let y = 1; g(y); class C { m = () => { y = 2; }; }`},

			// ---- Dimension 4: graceful degradation on body-absent members ----
			{Code: `abstract class A { abstract m(): void; }`},
			{Code: `declare const x: number;`},
			{Code: `function f(a: string): void;
function f(a: number): void;
function f(a: any): void { let v = 1; g(v); }`},
			{Code: `function f() {}`},
			{Code: `class A {}`},

			// ---- Dimension 4: empty destructuring pattern writes nothing ----
			{Code: `function f() { let v = 1; ({} = o); g(v); }`},
			{Code: `function f() { let v = 1; [] = arr; g(v); }`},

			// Locks in verifyAssignmentIsUsed() arm 3: a variable with no read
			// reference belongs to no-unused-vars, not to this rule.
			{Code: `function f() { let v = 1; v = 2; v = 3; }`},
			// Locks in the collection listener arm that skips a variable whose
			// declaration is outside the current code path.
			{Code: `function f() { let v = 1; g(v); return function () { v = 2; }; }`},
			// Locks in verifyAssignmentIsUsed() arm 5: a read from another code
			// path can happen at any time, so nothing can be proven unused.
			{Code: `function f() { let v = 1; v = 2; setTimeout(() => g(v)); }`},
			// Locks in the collection listener arm that skips an unresolved
			// global: nothing is known about where it is read.
			{Code: `undeclaredGlobal = 1; g(undeclaredGlobal); undeclaredGlobal = 2;`},
			// Locks in isIdentifierEvaluatedAfterAssignment() arm 2: an
			// identifier inside the assigned expression is read before the write.
			{Code: `function f() { let x = 1; x = x + 1; g(x); }`},
			// Locks in the module-export arms: a binding that leaves the file
			// can still be observed after the assignment.
			{Code: `export class foo {}
console.log(foo);
foo = 1;`},
			{Code: `export default function foo() {}
console.log(foo);
foo = 1;`},
			{Code: `let foo = 1;
console.log(foo);
foo = 2;
export { foo as bar };`},

			// ---- Dimension 4: a `typeof` type query reads the variable ----
			{Code: `function f() { let v = 1; g(v); v = 2; type T = typeof v; h<T>(); }`},

			// ---- Real-user: eslint/eslint#20947 (parameter decorator argument) ----
			// rslint treats the decorator's read as belonging to the method's own
			// code path, so the variable is left alone; ESLint reports it.
			{Code: `declare function Body(p: unknown): any;
const pipe = {};
class C { handler(@Body(pipe) _b: unknown) {} }`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized assignment target ----
			{
				Code: `function f() { let v = 1; g(v); (v) = 2; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 1, Column: 34, EndLine: 1, EndColumn: 35},
				},
			},
			{
				Code: `function f() { let v = 1; g(v); ((v)) = 2; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 1, Column: 35, EndLine: 1, EndColumn: 36},
				},
			},
			{
				Code: `function f() { let v = 1; g(v); (v)++; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 1, Column: 34, EndLine: 1, EndColumn: 35},
				},
			},

			// ---- Dimension 4: destructuring key forms ----
			{
				Code: `function f() { let v = 1; g(v); ({ v } = o); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 1, Column: 36, EndLine: 1, EndColumn: 37},
				},
			},
			{
				Code: `function f() { let v = 1; g(v); ({ 'k': v } = o); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 1, Column: 41, EndLine: 1, EndColumn: 42},
				},
			},
			{
				Code: `function f() { let v = 1; g(v); ({ 0: v } = o); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 1, Column: 39, EndLine: 1, EndColumn: 40},
				},
			},
			{
				Code: `function f() { let v = 1, k = 'a'; g(v, k); ({ [k]: v } = o); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 1, Column: 53, EndLine: 1, EndColumn: 54},
				},
			},

			// ---- Dimension 4: async / generator / accessor / constructor roots ----
			{
				Code: `async function f() { let v = 1; g(v); v = 2; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 1, Column: 39},
				},
			},
			{
				Code: `async function* f() { let v = 1; g(v); v = 2; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 1, Column: 40},
				},
			},
			{
				Code: `const o = { *m() { let v = 1; g(v); v = 2; } };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 1, Column: 37},
				},
			},
			{
				Code: `const o = { set p(x) { let v = 1; g(v); v = 2; } };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 1, Column: 41},
				},
			},
			{
				Code: `class C { constructor() { let v = 1; g(v); v = 2; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 1, Column: 44},
				},
			},

			// ---- Dimension 4: same-kind nesting must not bleed across roots ----
			{
				Code: `function f() { let v = 1; g(v); v = 2; function h() { let v = 1; g(v); v = 2; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 1, Column: 33},
					{MessageId: "unnecessaryAssignment", Line: 1, Column: 72},
				},
			},
			{
				Code: `class A { static { let v = 1; g(v); v = 2; class B { static { let v = 1; g(v); v = 2; } } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 1, Column: 37},
					{MessageId: "unnecessaryAssignment", Line: 1, Column: 80},
				},
			},
			{
				Code: `const o = { m() { let v = 1; g(v); v = 2; return () => { let v = 1; g(v); v = 2; }; } };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 1, Column: 36},
					{MessageId: "unnecessaryAssignment", Line: 1, Column: 75},
				},
			},
			// ---- Dimension 4: a shadowing block binding is a separate variable ----
			{
				Code: `function f() { let v = 1; { let v = 2; g(v); v = 3; } g(v); v = 4; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 1, Column: 46},
					{MessageId: "unnecessaryAssignment", Line: 1, Column: 61},
				},
			},

			// ---- Dimension 4: rest / spread targets in a pattern ----
			{
				Code: `function f() { let a = 1, r = 2; g(a, r); ({ a, ...r } = o); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 1, Column: 46},
					{MessageId: "unnecessaryAssignment", Line: 1, Column: 52},
				},
			},
			{
				Code: `function f() { let a = 1, r = 2; g(a, r); [a, ...r] = arr; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 1, Column: 44},
					{MessageId: "unnecessaryAssignment", Line: 1, Column: 50},
				},
			},
			// Locks in extractIdentifiersFromPattern() arm 3: an array hole is
			// skipped without shifting the elements after it.
			{
				Code: `function f() { let a = 1, b = 2; g(a, b); [a, , b] = arr; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 1, Column: 44},
					{MessageId: "unnecessaryAssignment", Line: 1, Column: 49},
				},
			},
			// Locks in extractIdentifiersFromPattern() arm 4: a rest element
			// nested inside another pattern still yields its target.
			{
				Code: `function f() { let a = 1; g(a); ({ x: [ ...a ] } = o); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 1, Column: 44},
				},
			},

			// Locks in isIdentifierUsedBetweenAssignedAndEqualSign() arm 1: an
			// update expression has no assigned expression at all.
			{
				Code: `function f() { let v = 1; g(v); v--; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 1, Column: 33},
				},
			},
			// Locks in the otherAssignmentAfterTargetAssignment arm that fires
			// when the target sits inside the other assignment's expression.
			{
				Code: `function f() { let x = 1; g(x); x = (x = 2); h(x); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 1, Column: 38},
				},
			},
			// Locks in the collection listener's unreachable-segment guard: the
			// dead store before `return` is reported, the one after it is not.
			{
				Code: `function f() { let v = 1; g(v); v = 2; return; v = 3; g(v); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 1, Column: 33},
				},
			},
			// Locks in verifyAssignmentIsUsed() arm 5 read-side detail: a write
			// from another code path is not a read, so the last store is unused.
			{
				Code: `function f() { let v = 1; g(v); setTimeout(() => { v = 2; }); v = 3; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 1, Column: 63},
				},
			},
			// Locks in the module-scope arms: an unexported binding in a module
			// is still reported.
			{
				Code: `import 'x';
let foo = 1;
console.log(foo);
foo = 2;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 4, Column: 1, EndLine: 4, EndColumn: 4},
				},
			},

			// ---- Dimension 4: decorators run where the member is declared ----
			{
				Code: `declare function Dec(p: unknown): any;
let v = 1;
class C { @Dec(v) m() {} }
v = 2;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 4, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			// ---- Dimension 4: an overload signature carries no body ----
			{
				Code: `function f(a: string): void;
function f(a: number): void;
function f(a: any): void { let v = 1; g(v); v = 2; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 3, Column: 45},
				},
			},
			// ---- Dimension 4: TS type-expression wrappers around a read ----
			{
				Code: `function f() { let v = 1; v = 2; g(v as any); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 1, Column: 20, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code: `function f() { let v: number | null = 1; v = 2; g(v!); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 1, Column: 20, EndLine: 1, EndColumn: 21},
				},
			},
			// ---- Real-user: eslint/eslint#20491 (early return in the else arm) ----
			{
				Code: `function props(parameters) {
	let content = '';

	if (typeof parameters === 'string') {
		content = parameters;
	} else {
		return {};
	}

	return { content };
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 2, Column: 6, EndLine: 2, EndColumn: 13},
				},
			},
			// ---- Real-user: eslint/eslint#19818 (update on a parameter) ----
			{
				Code: `[1, 2, 3].map(i => ++i);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
				},
			},

			// ---- Message text and multi-line position ----
			{
				Code: `function f() {
  let veryLongName = 1;
  g(veryLongName);
  veryLongName =
    2;
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryAssignment",
						Message:   "The value assigned to 'veryLongName' is not used in subsequent statements.",
						Line:      4, Column: 3, EndLine: 4, EndColumn: 15,
					},
				},
			},
			{
				Code: `function f() { let v = 1; g(v); v = 2; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryAssignment",
						Message:   "The value assigned to 'v' is not used in subsequent statements.",
						Line:      1, Column: 33, EndLine: 1, EndColumn: 34,
					},
				},
			},
		},
	)
}
