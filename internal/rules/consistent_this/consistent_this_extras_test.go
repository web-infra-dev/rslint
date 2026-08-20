// TestConsistentThisExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it
// covers, so future refactors can't silently regress them without breaking a
// named lock-in.
package consistent_this

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestConsistentThisExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ConsistentThisRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: parenthesized `this` receiver on the right-hand side ----
			{Code: "let that = (this);"},
			{Code: "let that = ((this));"},

			// ---- Dimension 4: parenthesized identifier on the assignment left-hand side ----
			{Code: "let that; (that) = this;"},

			// ---- Dimension 4: access/key forms — N/A: the rule never inspects a
			// property or method key (its only inputs are a declarator's plain
			// identifier name and an assignment's plain identifier target), so
			// key-form variation has no observable effect on it. ----

			// ---- Dimension 4: optional chain — N/A: `X?.y = this` is a parse
			// error (optional chaining cannot appear on an assignment's
			// left-hand side), and the rule only ever inspects the right-hand
			// side for a bare `this`, never through a member/chain expression. ----

			// ---- Dimension 4: declaration/container forms — class field with a
			// literal `this` initializer is never checked: PropertyDeclaration
			// is not one of the rule's listened node kinds ----
			{Code: "class Foo { that = this; }"},

			// ---- Dimension 4: declaration/container forms — class field whose
			// value is an arrow function is never checked: arrow functions are
			// excluded from the was-assigned scope walk ----
			{Code: "class Foo { bar = () => { var self; }; }", Options: []any{"self"}},

			// ---- Dimension 2: scoping & nesting — getter/setter/constructor/
			// static-method/computed-key/private/generator/async methods are all
			// checked the same as a plain function ----
			{Code: "class Foo { get bar() { var self; self = this; return self; } }", Options: []any{"self"}},
			{Code: "class Foo { set bar(v) { var self; self = this; } }", Options: []any{"self"}},

			// ---- Dimension 2: scoping & nesting — a class static block is
			// never checked: ClassStaticBlockDeclaration is not one of the
			// rule's listened node kinds ----
			{Code: "class Foo { static { var self; } }", Options: []any{"self"}},

			// ---- Dimension 2: scoping & nesting — a catch clause parameter is
			// never checked: CatchClause is not one of the rule's listened node
			// kinds ----
			{Code: "try {} catch (self) {}", Options: []any{"self"}},

			// ---- Dimension 2: scoping & nesting — same-kind nesting must not
			// bleed: each function's own alias is validated against its own
			// scope only ----
			{Code: "function outer() { var self = this; function inner() { var self = this; } }", Options: []any{"self"}},
			{Code: "function outer() { function inner() { var self; self = this; } }", Options: []any{"self"}},

			// ---- Dimension 4: graceful degradation — object rest / array
			// default / empty destructuring patterns are all skipped the same
			// way plain destructuring is, and must not crash ----
			{Code: "var {self, ...rest} = this;", Options: []any{"self"}},
			{Code: "var [self = 1] = this;", Options: []any{"self"}},
			{Code: "var {} = this;", Options: []any{"self"}},
			{Code: "var [] = this;", Options: []any{"self"}},

			// ---- Dimension 4: graceful degradation — a body-absent
			// declaration (ambient/abstract/overload signature) must not crash
			// and is never checked ----
			{Code: "declare function foo(): void;", Options: []any{"self"}},
			{Code: "abstract class Foo { abstract bar(): void; }", Options: []any{"self"}},
			{Code: "function foo(a: number): void;\nfunction foo(a: string): void;\nfunction foo(a: any) {}", Options: []any{"self"}},

			// ---- Locks in upstream checkWasAssigned() the-alias-was-never-declared
			// arm: scope.set.get(alias) is undefined, checkWasAssigned no-ops ----
			{Code: "function f() { var x = 1; }", Options: []any{"self"}},

			// ---- Locks in upstream checkWasAssigned() any-def-initialized arm:
			// variable.defs.some(...) short-circuits on the FIRST initialized
			// def, so an uninitialized sibling def of the same var is never
			// itself checked ----
			{Code: "var self; var self = this;", Options: []any{"self"}},
			{Code: "var self = this; var self;", Options: []any{"self"}},

			// ---- Locks in upstream checkAssignment(): a SequenceExpression
			// wrapping a valid `alias = this` assignment does not interfere
			// with the assignment's own operator/parent check ----
			{Code: "let that; (that = this, 1);"},

			// ---- An exported declaration is still a local binding of the
			// module scope, so the initialized / later-assigned arms of
			// checkWasAssigned() apply to it unchanged ----
			{Code: "export var that = this;"},
			{Code: "export var that;\nthat = this;"},

			// ---- eslint-scope opens a `for` scope only when the header
			// declares block-scoped bindings, so an assignment written in a
			// header that declares nothing (or declares with `var`) still
			// belongs to the enclosing scope ----
			{Code: "var self; for (self = this;;) { break; }", Options: []any{"self"}},
			{Code: "function f() { var self; for (self = this;;) { break; } }", Options: []any{"self"}},
			{Code: "function f() { var self; for (var i = 0; ;) self = this; }", Options: []any{"self"}},
			{Code: "function f() { var self; for (var x of (self = this, [])) {} }", Options: []any{"self"}},
			{Code: "function f() { var self; for (var k in (self = this, {})) {} }", Options: []any{"self"}},

			// ---- A computed key of an object-literal member is evaluated in
			// the scope enclosing the object literal, not in the member's own
			// scope, so an assignment written there counts as the enclosing
			// function's ----
			{Code: "function f() { var self; ({ [self = this]() {} }); }", Options: []any{"self"}},
			{Code: "function f() { var self; ({ get [self = this]() { return 1; } }); }", Options: []any{"self"}},
			{Code: "function f() { var self; ({ set [self = this](v) {} }); }", Options: []any{"self"}},
			{Code: "function f() { var self; ({ [self = this]: 1 }); }", Options: []any{"self"}},

			// ---- A `with` statement's object expression is evaluated in the
			// enclosing scope, so an assignment written there counts as the
			// enclosing function's ----
			{Code: "function f() { var self; with (self = this) { } }", Options: []any{"self"}},

			// ---- An enum body is a scope of its own, but the enum's own name
			// is bound in — and its members are irrelevant to — the enclosing
			// scope, so an assignment written outside the body still counts as
			// the enclosing function's ----
			{Code: "function f() { var self; enum E { A = 1 } self = this; }", Options: []any{"self"}},

			// ---- eslint-scope hands every name in a destructuring target the
			// whole assignment's right-hand side as its write expression, so a
			// pattern assigned `this` satisfies each alias it binds ----
			{Code: "var self; ({ self } = this)", Options: []any{"self"}},
			{Code: "var self; [self] = this", Options: []any{"self"}},
			{Code: "var self; ({ a: { self } } = this)", Options: []any{"self"}},
			{Code: "var self; ({ self = 1 } = this)", Options: []any{"self"}},
			{Code: "var self; ({ ...self } = this)", Options: []any{"self"}},
			{Code: "var self; [self = 1] = this", Options: []any{"self"}},
			{Code: "function f() { var self; ({ self } = this); }", Options: []any{"self"}},

			// ---- typescript-eslint unwraps one TypeScript wrapper off an
			// assignment's own left-hand side, so a name written under a
			// single `!`, `as` or `<T>` is still that assignment's write
			// target ----
			{Code: "var self; self! = this;", Options: []any{"self"}},
			{Code: "var self; (self as any) = this;", Options: []any{"self"}},
			{Code: "var self; (<any>self) = this;", Options: []any{"self"}},

			// ---- Inside a destructuring pattern the name is instead reached
			// by a visitor that descends through every node kind it has no
			// handler for, so no number of wrappers hides it there ----
			{Code: "var self; [self!!] = this;", Options: []any{"self"}},
			{Code: "var self; [self satisfies any] = this;", Options: []any{"self"}},
			{Code: "var self; ({ a: self! } = this);", Options: []any{"self"}},

			// ---- A shorthand property's initializer is the pattern's default
			// value, an expression of its own: the assignment written there is
			// a real assignment of `this`, not a target of the outer one ----
			{Code: "var obj: any; var x; var self; ({ x = self = this } = obj);", Options: []any{"self"}},
			{Code: "var obj: any; var x; var self; ({ x = self! = this } = obj);", Options: []any{"self"}},

			// ---- A signature with no body is a TSDeclareFunction or a
			// TSEmptyBodyFunctionExpression upstream, and neither of those is
			// a FunctionDeclaration or a FunctionExpression, so the exit
			// listeners never fire for one and its parameters go unchecked ----
			{Code: "declare function f(self: any): void", Options: []any{"self"}},
			{Code: "abstract class C { abstract m(self: any): void }", Options: []any{"self"}},
			{Code: "declare class C { m(self: any): void }", Options: []any{"self"}},
			{Code: "interface I { m(self: any): void }", Options: []any{"self"}},

			// ---- A class decorator is evaluated in the scope enclosing the
			// class rather than in the class scope, so an assignment written
			// there counts as the enclosing function's ----
			{Code: "function f() { var self; @((self = this, dec)) class C {} }", Options: []any{"self"}},

			// ---- A re-export binds no local name, so — unlike `export
			// import` below — it is no variable of the module scope ----
			{Code: "export { self } from './x'", Options: []any{"self"}},
			{Code: "export * as self from './x'", Options: []any{"self"}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Options contract: the schema declares no JSON Schema
			// "default" (upstream fills the default via meta.defaultOptions
			// instead), so parseOptions must supply ["that"] itself when no
			// options are configured at all. This must behave identically to
			// explicitly configuring ["that"] ----
			{
				Code: "var vm = this",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAlias", Line: 1, Column: 5, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code:    "var vm = this",
				Options: []any{"that"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAlias", Line: 1, Column: 5, EndLine: 1, EndColumn: 14},
				},
			},

			// ---- Dimension 4: TS non-null assertion / `as` / `satisfies` on
			// the right-hand side do NOT count as a bare `this` (unlike a
			// parenthesized `(this)`, these are distinct wrapper node kinds
			// that the rule intentionally does not unwrap) ----
			{
				Code: "let that: any; that = this!;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 5, EndLine: 1, EndColumn: 14},
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 16, EndLine: 1, EndColumn: 28},
				},
			},
			{
				Code: "let that: any; that = this as any;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 5, EndLine: 1, EndColumn: 14},
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 16, EndLine: 1, EndColumn: 34},
				},
			},
			{
				Code: "let that: any; that = this satisfies any;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 5, EndLine: 1, EndColumn: 14},
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 16, EndLine: 1, EndColumn: 41},
				},
			},

			// ---- Dimension 4: declaration/container forms — a class field
			// whose value is a plain (non-arrow) function expression IS
			// checked, unlike the literal-`this`-field and arrow-field cases
			// above ----
			{
				Code:    "class Foo { bar = function() { var self; } }",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 36, EndLine: 1, EndColumn: 40},
				},
			},

			// ---- Locks in upstream checkWasAssigned() report-every-def arm: a
			// twice-redeclared var reports each declarator individually, in
			// addition to the checkAssignment report on the bad assignment ----
			{
				Code:    "var self; var self; self = 42;",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 5, EndLine: 1, EndColumn: 9},
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 15, EndLine: 1, EndColumn: 19},
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 21, EndLine: 1, EndColumn: 30},
				},
			},

			// ---- Locks in upstream checkWasAssigned() alias-declared-via-
			// function/class-declaration arm: a function or class declaration
			// (not just var/let/const) can also shadow an alias name, and is
			// reported on its own whole declaration node ----
			{
				Code:    "function self() {}",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    "class self {}",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},

			// ---- Real-user: https://github.com/eslint/eslint/issues/1846 — a
			// designated alias used as a function parameter that is never
			// assigned `this` is still flagged, at the whole enclosing
			// function: eslint-scope makes a Parameter def's `.node` the
			// function rather than the parameter. ----
			{
				Code:    "function decorate(er, code, self) {}",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 1, EndLine: 1, EndColumn: 37},
				},
			},

			// ---- A class member's function is the FunctionExpression ESTree
			// hangs off the member, so a parameter of one is reported from the
			// signature on — the name and the modifiers belong to the member,
			// not to the function ----
			{
				Code:    "class C { public static async *m<T>(self) {} }",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 33, EndLine: 1, EndColumn: 45},
				},
			},

			// ---- A destructured parameter's names are parameter definitions
			// of the same shape, so they are reported at the function too ----
			{
				Code:    "function f({ a: [self] }) {}",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 1, EndLine: 1, EndColumn: 29},
				},
			},

			// ---- Locks in upstream checkWasAssigned(): a default parameter
			// value of `this` is not a "was assigned" write reference (default
			// values aren't AssignmentExpressions), so it is flagged the same
			// as an unassigned parameter above ----
			{
				Code:    "function f(self = this) {}",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 1, EndLine: 1, EndColumn: 27},
				},
			},

			// ---- Real-user: https://github.com/eslint/eslint/issues/19244 (fixed
			// by https://github.com/eslint/eslint/pull/19383) — an alias
			// declared at the top level of an ES module and assigned inside a
			// nested function is still flagged; the module's top-level scope
			// must be checked the same as a script's ----
			{
				Code:    "export {};\nvar that;\nfunction f() { that = this; }",
				Options: []any{"that"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 2, Column: 5, EndLine: 2, EndColumn: 9},
				},
			},

			// ---- Real-user: https://github.com/eslint/eslint/issues/2669 —
			// locks in upstream checkAssignment()'s non-alias branch: a
			// compound assignment (`+=`) of `this` to a name that is NOT a
			// designated alias is still reported, because that branch never
			// looks at the operator (unlike the alias branch, where the
			// operator matters) ----
			{
				Code:    "var sum = 0;\n$.each(total, function () {\n  sum += this;\n});",
				Options: []any{"that"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAlias", Line: 3, Column: 3, EndLine: 3, EndColumn: 14},
				},
			},

			// ---- Locks in upstream checkWasAssigned(): reference.from must be
			// exactly the checked scope. Rslint's Differences-from-ESLint entry:
			// a nested block (`if`, `for`, `try`, a bare `{}`, ...) is always a
			// distinct scope container in rslint, regardless of the configured
			// ECMAScript version — unlike ESLint, which under ecmaVersion 5
			// treats the whole function body as one flat scope ----
			{
				Code:    "function f() { var self; if (true) { self = this; } }",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 20, EndLine: 1, EndColumn: 24},
				},
			},

			// ---- A `for` header that declares block-scoped bindings IS its
			// own scope, so an assignment written there does not count as the
			// enclosing function's ----
			{
				Code:    "function f() { var self; for (let i = 0; ;) self = this; }",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 20, EndLine: 1, EndColumn: 24},
				},
			},

			// ---- A `with` statement body is a scope of its own, so an
			// assignment written inside it does not count as the enclosing
			// function's ----
			{
				Code:    "function f() { var self; with (obj) self = this; }",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 20, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    "function f() { var self; with (obj) { self = this; } }",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 20, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    "function f() { var self; with (obj) for (;;) { self = this; break; } }",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 20, EndLine: 1, EndColumn: 24},
				},
			},

			// ---- A computed key of a class member is evaluated in the class
			// scope, not the enclosing function's, so an assignment written
			// there does not count as the enclosing function's ----
			{
				Code:    "function f() { var self; class C { [self = this]() {} } }",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 20, EndLine: 1, EndColumn: 24},
				},
			},

			// ---- An enum body is a scope of its own, so an assignment written
			// in a member's initializer does not count as the enclosing scope's ----
			{
				Code:    "function f() { var self; enum E { A = (self = this, 1) } }",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 20, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    "var self; enum E { A = (self = this, 1) }",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 5, EndLine: 1, EndColumn: 9},
				},
			},

			// ---- Dimension 1: TS-specific syntax — a type alias, interface,
			// namespace or type parameter is a variable of its scope just like a
			// value declaration, and is reported when it shadows an alias name ----
			{
				Code:    "type self = string;",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    "interface self {}",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:    "namespace self {}",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:    "export namespace self {}",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 8, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:    "function f() { type self = string; }",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 16, EndLine: 1, EndColumn: 35},
				},
			},
			{
				Code:    "function f<self>() {}",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 12, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code:    "function f<self extends object>() {}",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 12, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    "class C { m<self>() {} }",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 13, EndLine: 1, EndColumn: 17},
				},
			},

			// ---- An alias declared with `export` or bound by an `import` is
			// a local of the module scope just like a plain declaration, and
			// is reported at the same position ESLint gives the declaration
			// its export wrapper holds ----
			{
				Code: "export var that;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 12, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code: "export let that;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 12, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code: "export function that() {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 8, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code: "export class that {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 8, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code: "export default function that() {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 16, EndLine: 1, EndColumn: 34},
				},
			},
			{
				Code: "import that from './x';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 8, EndLine: 1, EndColumn: 12},
				},
			},
			{
				Code: "import {a as that} from './x';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 9, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code: "import * as that from './x';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 8, EndLine: 1, EndColumn: 17},
				},
			},

			// ---- A destructuring assignment only satisfies the alias when
			// the whole pattern is assigned a plain `this`: a `this` buried in
			// the right-hand side, a compound operator, and a for-of header
			// are all still misses ----
			{
				Code:    "var self; ({ self } = { self: this })",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 5, EndLine: 1, EndColumn: 9},
				},
			},
			{
				Code:    "var self; ({ self } ??= this)",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 5, EndLine: 1, EndColumn: 9},
				},
			},
			{
				Code:    "var self; for ({ self } of [this]) {}",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 5, EndLine: 1, EndColumn: 9},
				},
			},

			// ---- An overload signature is not checked, but the
			// implementation it belongs to is ----
			{
				Code:    "function f(self: any): void; function f(self: any) { }",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 30, EndLine: 1, EndColumn: 55},
				},
			},

			// ---- Redeclaring a name with a conflicting meaning leaves the
			// binder holding only one of the declarations, while eslint-scope
			// keeps a single variable holding them all, so every one of them
			// is reported ----
			{
				Code:    "var self; function self() {}",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 5, EndLine: 1, EndColumn: 9},
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 11, EndLine: 1, EndColumn: 29},
				},
			},
			{
				Code:    "let self; var self;",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 5, EndLine: 1, EndColumn: 9},
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 15, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    "import self from './x'; var self;",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 8, EndLine: 1, EndColumn: 12},
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 29, EndLine: 1, EndColumn: 33},
				},
			},

			// ---- ...and one initialized declaration among them silences the
			// was-assigned check for the whole variable, leaving only the
			// checkAssignment report on the bad initializer ----
			{
				Code:    "var self = 1; function self() {}",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 5, EndLine: 1, EndColumn: 13},
				},
			},

			// ---- eslint-scope opens a class scope spanning the heritage
			// clauses, so a write there never satisfies a binding of the
			// enclosing function ----
			{
				Code:    "function f() { var self; class C extends (self = this, Object) {} }",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 20, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    "function f() { var self; (class extends (self = this, Object) {}); }",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 20, EndLine: 1, EndColumn: 24},
				},
			},

			// ---- `export import` binds a local name even though the binder
			// files it under the module's exports and leaves it out of the
			// module's locals ----
			{
				Code:    "export import self = require('./x')",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 8, EndLine: 1, EndColumn: 36},
				},
			},

			// ---- A default import is reported at its name: the `type`
			// keyword of a type-only import stays on the ImportDeclaration
			// upstream, outside the specifier ----
			{
				Code:    "import type self from './x'",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 13, EndLine: 1, EndColumn: 17},
				},
			},

			// ---- Only one wrapper is unwrapped off an assignment's own
			// left-hand side, and `satisfies` is not one of the kinds unwrapped
			// at all, so none of these writes the alias ----
			{
				Code:    "var self; self!! = this;",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 5, EndLine: 1, EndColumn: 9},
				},
			},
			{
				Code:    "var self; (self! as any) = this;",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 5, EndLine: 1, EndColumn: 9},
				},
			},
			{
				Code:    "var self; (self satisfies any) = this;",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 5, EndLine: 1, EndColumn: 9},
				},
			},

			// ---- A wrapped target is still only a write of what the
			// assignment's right-hand side holds, under the operator it holds
			// it with ----
			{
				Code:    "var self; self! = 1;",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 5, EndLine: 1, EndColumn: 9},
				},
			},
			{
				Code:    "var self; self! += this;",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 5, EndLine: 1, EndColumn: 9},
				},
			},

			// ---- The assignment in a shorthand property's default value is
			// checked as the plain assignment it is ----
			{
				Code: "var obj: any; var x; var self; ({ x = self = this } = obj);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAlias", Line: 1, Column: 39, EndLine: 1, EndColumn: 50},
				},
			},
		},
	)
}
