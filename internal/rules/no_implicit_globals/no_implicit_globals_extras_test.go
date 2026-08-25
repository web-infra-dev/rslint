package no_implicit_globals

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoImplicitGlobalsExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it
// covers, so future refactors can't silently regress them without breaking a
// named lock-in.
func TestNoImplicitGlobalsExtras(t *testing.T) {
	lexical := []any{map[string]any{"lexicalBindings": true}}
	es5 := rule.LanguageOptions{ECMAVersion: 5}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoImplicitGlobalsRule,
		withScriptDefaults([]rule_tester.ValidTestCase{
			// ---- Dimension 4: receiver wrappers, opaque TS wrappers ----

			// The single unwrap an AssignmentExpression's Left gets sheds an
			// `as`/`<T>`/`!` wrapper but never a `satisfies`, so a directly
			// wrapped target is no pattern at all (see findPureAssignmentRoot
			// doc comment). Inside a real pattern the same wrapper is
			// transparent — see the invalid cases.
			{Code: `(foo satisfies any) = 1;`},
			{Code: `(Array satisfies any) = 1;`},
			{Code: `(foo satisfies any as any) = 1;`},
			{Code: `((foo as any) satisfies any) = 1;`},

			// An AssignmentExpression's Left is unwrapped exactly once, so a
			// second wrapper leaves something that is no longer a pattern.
			{Code: `(foo as any)! = 1;`},
			{Code: `foo!! = 1;`},
			{Code: `((foo as any) as any) = 1;`},
			{Code: `(Array as any)! = 1;`},
			// A wrapper around the destructuring target fails the same test:
			// `[foo]` is no longer a destructuring pattern once an assertion
			// wraps it.
			{Code: `([foo] as any) = arr;`},

			// ---- Dimension 4: declaration vs expression forms ----

			// Class expressions never bind a name at the declaration site.
			{Code: `const C = class Foo {};`},
			{Code: `const C = class Array {};`},

			// ---- Dimension 4: nesting / traversal boundaries ----

			// Locks in: `let`/`const` in a for-loop head get their own
			// per-iteration scope, not the global scope.
			{Code: `for (let i = 0; i < 1; i++) {}`, Options: lexical},
			{Code: `for (const i of []) {}`, Options: lexical},

			// Locks in: double-nested block does not leak `let` to global scope.
			{Code: `{ { let foo = 1; } }`, Options: lexical},

			// ---- Dimension 4: graceful degradation ----

			// Empty destructuring patterns bind no names — nothing to report,
			// and CollectBindingNames must not panic on them.
			{Code: `const {} = {};`, Options: lexical},
			{Code: `const [] = [];`, Options: lexical},

			// ---- Dimension 4: `declare` / ambient forms have no runtime binding ----
			//
			// A documented divergence: upstream reports the bare `declare`
			// forms and each overload signature (see the rule doc).

			{Code: `declare var foo: number;`},
			{Code: `declare function foo(): void;`},
			{Code: `declare class Foo {}`, Options: lexical},
			{Code: `declare global { var foo: number; }`},
			{Code: `declare global { function foo(): void; }`},

			// ---- Branch lock-in: `using` has no ESLint analog ----
			//
			// (`await using` can't be tested at true global scope: top-level
			// await requires module context, which disables the whole
			// declaration-checking path via hasNonGlobalTopLevelScope — so
			// `using`'s shared code path is the only reachable case.)
			{Code: `using foo = getResource();`, Options: lexical},

			// ---- Real-user: readonly builtin shadowed by a for-in/for-of loop variable is a common false-positive concern ----
			{Code: `for (const Array of [[1], [2]]) { Array.length; }`, Options: lexical},

			// ---- Options contract: an empty options object fills the schema
			// default (lexicalBindings: false), same as omitting options entirely ----
			{Code: `const foo = 1;`, Options: []any{map[string]any{}}},

			// ---- Leak strictness follows the same module-ness the declaration
			// checks use, so every extension answers it the one way ----

			// A .js file is a module under ESLint's default language selection
			// even with no import/export syntax of its own, so its top level is
			// strict and cannot leak.
			{Code: `foo = 1;`, FileName: "js/default-module-leak.js", TSConfig: "tsconfig.allow-js.json"},
			// ESLint's default language selection picks CommonJS for .cjs alone,
			// so a .cts follows the module default and its top level is strict
			// whether or not it carries module syntax of its own.
			{Code: `foo = 1;`, FileName: "cts/default-module-leak.cts"},
			{Code: `export {}; foo = 1;`, FileName: "cts/module-leak.cts"},

			// ---- `/* exported */` exempts the read-only global assignment ----
			//
			// Upstream skips an exported variable before reaching the reference
			// loop that reports assignmentToReadonlyGlobal.
			{Code: "/* exported Array */\nArray = 1;"},
			// The directive is read from the shared inline-directive view, so it
			// still applies where this rule's declaration checks are switched
			// off — a module file keeps its readonly-global protection.
			{
				Code:            "/* exported Array */\nexport {};\nArray = 1;",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			},

			// ---- A global script's top-level declarations all define the
			// global-scope variable, type-only ones included ----
			//
			// The name then has a definition, so it is neither an implicit
			// variable nor a definition-less read-only global.
			{Code: `foo = 1; interface foo {}`},
			{Code: `foo = 1; type foo = number;`},
			{Code: `Array = 1; interface Array {}`},
			{Code: `Array = 1; type Array<T> = T;`},
			{Code: `foo = 1; namespace foo {}`},

			// ---- Strictness comes from the scope the write is evaluated in ----
			//
			// TS namespace and enum scopes are strict, so a write inside them
			// records no implicit global.
			{Code: `namespace N { foo = 1; }`},
			{Code: `namespace N { function f() { foo = 1; } }`},
			{Code: `enum E { A = (foo = 1) }`},
			{Code: `namespace N { enum E { A = (foo = 1) } }`},
			// Everything a class body evaluates is strict, decorators on its
			// own members included.
			{Code: `class C { p = (foo = 1); }`},
			{Code: `class C { [foo = 1]: number; }`},
			{Code: `class C { static { foo = 1; } }`},
			{Code: `class C extends (foo = 1) {}`},
			{Code: `class C { @dec(foo = 1) m() {} }`},

			// A member expression names the property being assigned; neither
			// identifier inside it is itself a write target.
			{Code: `[foo.bar] = arr;`},

			// PatternVisitor skips decorator subtrees even when the decorated
			// class appears inside an expression-shaped assignment pattern.
			{Code: `[@foo class {}] = arr;`},

			// Within an expression-shaped pattern element, the right side of a
			// compound assignment stays a read. A writable left side therefore
			// leaves no implicit write to report for `bar`.
			{Code: `[foo += bar] = arr;`, Globals: map[string]any{"foo": "writable"}},

			// Unlike plugin-kit's ordinary JavaScript object, the shared exported
			// directive view has no inherited __proto__ setter: the name is kept
			// and exempts its declaration like every other exported name.
			{Code: `/* exported __proto__ */ var __proto__;`},

			// ---- Block scopes exist from ES2015 on ----
			{Code: `{ function foo() {} }`},
			{Code: `if (true) { function foo() {} }`},
		}),
		withScriptDefaults([]rule_tester.InvalidTestCase{
			// ---- Dimension 4: receiver wrappers on the leak/readonly identifier ----

			// Locks in: parenthesized assignment target is still a pure write —
			// findPureAssignmentRoot passes transparently through parens.
			{
				Code:   `(foo) = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalVariableLeak"}},
			},
			{
				Code:   `((foo)) = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalVariableLeak"}},
			},
			// Non-null and type assertions are transparent too (unlike
			// `satisfies`).
			{
				Code:   `foo! = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalVariableLeak"}},
			},
			{
				Code:   `Array! = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "assignmentToReadonlyGlobal"}},
			},
			{
				Code:   `(foo as any) = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalVariableLeak", Line: 1, Column: 1, EndLine: 1, EndColumn: 17}},
			},
			{
				Code:   `(<any>foo) = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalVariableLeak", Line: 1, Column: 1, EndLine: 1, EndColumn: 15}},
			},
			{
				Code:   `(Array as any) = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "assignmentToReadonlyGlobal"}},
			},
			// Inside a pattern, and in a for-in/for-of head, the target is
			// visited directly, so a wrapper is transparent at any depth.
			{
				Code:   `[foo!!] = arr;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalVariableLeak"}},
			},
			{
				Code:   `[(foo as any)!] = arr;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalVariableLeak"}},
			},
			{
				Code:   `({a: (foo as any)!} = obj);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalVariableLeak"}},
			},
			{
				Code:   `for ((foo as any)! of arr) {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalVariableLeak"}},
			},
			{
				Code:   `for (foo! in obj) {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalVariableLeak"}},
			},

			// PatternVisitor traverses parser-accepted expression-shaped pattern
			// elements. It visits ordinary expression children, but treats call
			// arguments as reads and continues only through the callee.
			{
				Code: `[foo + bar] = arr;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalVariableLeak", Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
					{MessageId: "globalVariableLeak", Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:   `[fn(arg)] = arr;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalVariableLeak"}},
			},
			{
				Code: `[foo ? bar : baz] = arr;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalVariableLeak"},
					{MessageId: "globalVariableLeak"},
					{MessageId: "globalVariableLeak"},
				},
			},
			// PatternVisitor treats a compound assignment's left side as an
			// assignment pattern and records both the compound write and the
			// containing destructuring write. Its right side remains a read.
			{
				Code:    `[foo += bar] = arr;`,
				Globals: map[string]any{"bar": "writable"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalVariableLeak", Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
					{MessageId: "globalVariableLeak", Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
				},
			},
			// An equals expression nested in another expression-shaped pattern
			// element is another assignment-pattern write; both reports use the
			// containing destructuring assignment's range.
			{
				Code:    `[(foo = bar) + baz] = arr;`,
				Globals: map[string]any{"baz": "writable"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalVariableLeak", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
					{MessageId: "globalVariableLeak", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
				},
			},
			// Function and class names reached through PatternVisitor's synthetic
			// expression traversal are assignment targets, not ordinary lexical
			// references to their expression-local bindings.
			{
				Code:   `[function foo(){}] = arr;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalVariableLeak"}},
			},
			{
				Code:   `[class C {}] = arr;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalVariableLeak"}},
			},
			// The decorator's callee and arguments stay reads, while the class
			// name outside that subtree remains a synthetic pattern write.
			{
				Code:   `[@dec(foo) class C {}] = arr;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalVariableLeak"}},
			},
			// Every identifier PatternVisitor visits is resolved from the scope
			// containing the outer assignment. Arrow parameters therefore become
			// synthetic writes, while a real outer binding still suppresses one.
			{
				Code:   `[(foo => 0)] = arr;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalVariableLeak"}},
			},
			{
				Code: `var C; [class C {}] = arr;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalNonLexicalBinding"},
				},
			},

			// A default under object rest still contributes an extra write, and
			// both diagnostics use the range of the outer assignment.
			{
				Code: `[{...foo = bar}] = arr;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalVariableLeak", Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
					{MessageId: "globalVariableLeak", Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code: `[foo! = 1] = arr;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalVariableLeak"},
					{MessageId: "globalVariableLeak"},
				},
			},
			{
				Code:   `(Array) = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "assignmentToReadonlyGlobal"}},
			},

			// ---- Dimension 4: access/key forms in destructuring assignment ----

			// Numeric-literal key.
			{
				Code:   `({0: foo} = obj);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalVariableLeak"}},
			},
			// Computed key — the key expression itself is a read, not a write;
			// only the bound `foo` leaks.
			{
				Code:   `const k = "x"; ({[k]: foo} = obj);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalVariableLeak"}},
			},

			// ---- Dimension 4: nesting / traversal boundaries ----

			// Locks in: nested function declarations don't bleed to the
			// global-scope check — exactly one error, for `outer`; `inner`
			// must not also be reported.
			{
				Code:   `function outer() { function inner() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalNonLexicalBinding"}},
			},

			// An overload pair binds one name, so only the implementation is
			// reported. Upstream counts one def per signature and reports the
			// pair twice — a documented divergence (see the rule doc).
			{
				Code:   "function foo(a: string): void;\nfunction foo(a: any): void {}",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalNonLexicalBinding", Line: 2}},
			},

			// `var` hoists past a bare block to the true global scope — this
			// is untested upstream (their block-scoping tests only cover
			// `const`/`let`/`class`).
			{
				Code:   `{ var foo = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalNonLexicalBinding"}},
			},
			// `var` in a for-loop head also hoists to the global scope.
			{
				Code:   `for (var i = 0; i < 1; i++) {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalNonLexicalBinding"}},
			},
			{
				Code:   `for (var x in {}) {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalNonLexicalBinding"}},
			},

			// ---- Dimension 4: graceful degradation ----

			// Rest element in an object destructuring declaration: both `a`
			// and `rest` are separate bindings from the same declarator, and
			// (matching ESLint's def.node quirk) share the same reported
			// position.
			{
				Code:    `const {a, ...rest} = obj;`,
				Options: lexical,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalLexicalBinding"},
					{MessageId: "globalLexicalBinding"},
				},
			},
			// Object rest in a destructuring *assignment* (not a declaration)
			// is still a write target for leak purposes.
			{
				Code:   `({...foo} = obj);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalVariableLeak"}},
			},

			// ---- Branch lock-in: default value inside a destructuring assignment ----

			// Locks in findPureAssignmentRoot's IsDefaultValueInDestructuringAssignment
			// branch: `foo`'s own `= 1` is a pattern default, not a real
			// assignment, so the walk must continue past it to the enclosing
			// `[foo = 1] = arr` to find the true leak root. The default also
			// carries a scope write of its own, so upstream reports the root
			// once per default plus once for the assignment.
			{
				Code: `[foo = 1] = arr;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalVariableLeak", Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
					{MessageId: "globalVariableLeak", Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code: `({x: foo = 1} = obj);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalVariableLeak"},
					{MessageId: "globalVariableLeak"},
				},
			},
			// A shorthand property keeps its default on the property itself
			// rather than in a nested assignment, and still counts.
			{
				Code: `({foo = 1} = obj);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalVariableLeak"},
					{MessageId: "globalVariableLeak"},
				},
			},
			// Nested defaults stack: `foo` sits under both `foo = 1` and
			// `[foo = 1] = []`, so the write count reaches three.
			{
				Code: `[[foo = 1] = []] = arr;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalVariableLeak"},
					{MessageId: "globalVariableLeak"},
					{MessageId: "globalVariableLeak"},
				},
			},
			// A read-only global assigned through a default duplicates the
			// same way.
			{
				Code: `/*global foo:readonly*/ [foo = 1] = arr;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToReadonlyGlobal", Line: 1, Column: 25, EndLine: 1, EndColumn: 40},
					{MessageId: "assignmentToReadonlyGlobal", Line: 1, Column: 25, EndLine: 1, EndColumn: 40},
				},
			},
			// Only the pattern side counts: the two diagnostics both belong to
			// `bar`, since `foo` in the default's value is a read.
			{
				Code: `[bar = foo] = arr;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalVariableLeak"},
					{MessageId: "globalVariableLeak"},
				},
			},

			// ---- Diagnostic contract: full Line/Column/EndLine/EndColumn per container, including a multi-line case ----
			//
			// Upstream's own suite mostly asserts message text only; these
			// lock in exact ranges rslint reports, one per message category,
			// each over two declarations/writes on separate lines.
			{
				Code: "var foo = 1;\nvar bar = 2;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalNonLexicalBinding", Line: 1, Column: 5, EndLine: 1, EndColumn: 12},
					{MessageId: "globalNonLexicalBinding", Line: 2, Column: 5, EndLine: 2, EndColumn: 12},
				},
			},
			{
				Code: "function foo() {\n  return 1;\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalNonLexicalBinding", Line: 1, Column: 1, EndLine: 3, EndColumn: 2},
				},
			},
			{
				Code:    "class Foo {\n  method() {}\n}",
				Options: lexical,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalLexicalBinding", Line: 1, Column: 1, EndLine: 3, EndColumn: 2},
				},
			},
			{
				Code:    "const foo = 1;\nlet bar = 2;",
				Options: lexical,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalLexicalBinding", Line: 1, Column: 7, EndLine: 1, EndColumn: 14},
					{MessageId: "globalLexicalBinding", Line: 2, Column: 5, EndLine: 2, EndColumn: 12},
				},
			},
			{
				Code: "foo = 1;\nbar = 2;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalVariableLeak", Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
					{MessageId: "globalVariableLeak", Line: 2, Column: 1, EndLine: 2, EndColumn: 8},
				},
			},
			{
				Code: "Array = 1;\nObject = 2;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToReadonlyGlobal", Line: 1, Column: 1, EndLine: 1, EndColumn: 10},
					{MessageId: "assignmentToReadonlyGlobal", Line: 2, Column: 1, EndLine: 2, EndColumn: 11},
				},
			},
			{
				Code: "var Array = 1;\nvar Object = 2;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclarationOfReadonlyGlobal", Line: 1, Column: 5, EndLine: 1, EndColumn: 14},
					{MessageId: "redeclarationOfReadonlyGlobal", Line: 2, Column: 5, EndLine: 2, EndColumn: 15},
				},
			},

			// ---- Real-user: a locally-scoped-looking accumulator forgets `var`/`let` inside a function ----
			//
			// The classic bug this rule exists to catch: a script author means
			// `total` to be local to the IIFE, forgets the declaration, and
			// silently creates a global. Wrapped in an IIFE (a function
			// expression, not a declaration) so the only diagnostic is the leak.
			{
				Code:   `(function calc(items) { total = 0; for (const item of items) { total += item; } return total; })();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalVariableLeak"}},
			},

			// ---- Leak strictness: CommonJS keeps its own top-level scope
			// without becoming a module, so its top level stays sloppy ----

			{
				Code:     `foo = 1;`,
				FileName: "cjs/commonjs-leak.cjs",
				TSConfig: "tsconfig.allow-js.json",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "globalVariableLeak"}},
			},
			{
				Code:     `export {}; foo = 1;`,
				FileName: "cjs/commonjs-module-syntax-leak.cjs",
				TSConfig: "tsconfig.allow-js.json",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "globalVariableLeak"}},
			},

			// ---- Dimension 4: `satisfies` inside a real pattern ----
			//
			// A pattern element is visited directly, so the wrapper the
			// AssignmentExpression unwrap would have rejected is transparent
			// here.
			{
				Code:   `[foo satisfies any] = arr;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalVariableLeak", Line: 1, Column: 1, EndLine: 1, EndColumn: 26}},
			},
			{
				Code:   `[[foo satisfies any]] = arr;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalVariableLeak"}},
			},
			{
				Code:   `({a: foo satisfies any} = obj);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalVariableLeak"}},
			},
			{
				Code:   `[(foo satisfies any) as any] = arr;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalVariableLeak"}},
			},
			{
				Code:   `for ((foo satisfies any) of arr) {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalVariableLeak"}},
			},
			{
				Code:   `for ((Array satisfies any) in obj) {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "assignmentToReadonlyGlobal"}},
			},

			// ---- Only a global script's own top level defines the
			// global-scope variable ----
			//
			// An inner scope holds a separate variable, and a value reference
			// never resolves to a type-only one, so the write still reaches the
			// global scope.
			{
				Code: `function f() { foo = 1; interface foo {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalNonLexicalBinding"},
					{MessageId: "globalVariableLeak"},
				},
			},
			{
				Code:   `{ Array = 1; interface Array {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "assignmentToReadonlyGlobal"}},
			},

			// ---- Strictness: a class's own decorators are evaluated before
			// the class scope exists ----
			{
				Code:   `@dec(foo = 1) class C {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalVariableLeak", Line: 1, Column: 6, EndLine: 1, EndColumn: 13}},
			},
			{
				Code:   `const C = @dec(foo = 1) class {};`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalVariableLeak"}},
			},

			// ---- ES3/ES5 have no block scopes, so a block-level function
			// declaration binds in the global scope ----
			{
				Code:            `{ function foo() {} }`,
				LanguageOptions: es5,
				Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "globalNonLexicalBinding", Line: 1, Column: 3, EndLine: 1, EndColumn: 20}},
			},
			{
				Code:            `if (true) { function foo() {} }`,
				LanguageOptions: es5,
				Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "globalNonLexicalBinding"}},
			},
			// It binds in the enclosing variable scope, not always the global
			// one: only `outer` is a global declaration here.
			{
				Code:            `function outer() { { function foo() {} } }`,
				LanguageOptions: es5,
				Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "globalNonLexicalBinding"}},
			},

			// ---- `/* exported */` does not exempt a leak ----
			//
			// Upstream collects leaks from the global scope's implicit
			// variables, a separate pass the directive never reaches.
			{
				Code:   "/* exported foo */\nfoo = 1;",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalVariableLeak"}},
			},
		}),
	)
}

func withScriptDefaults[T rule_tester.ValidTestCase | rule_tester.InvalidTestCase](cases []T) []T {
	for i := range cases {
		switch testCase := any(&cases[i]).(type) {
		case *rule_tester.ValidTestCase:
			if testCase.FileName == "" && testCase.LanguageOptions.SourceType == "" {
				testCase.LanguageOptions.SourceType = "script"
			}
		case *rule_tester.InvalidTestCase:
			if testCase.FileName == "" && testCase.LanguageOptions.SourceType == "" {
				testCase.LanguageOptions.SourceType = "script"
			}
		}
	}
	return cases
}
