package no_object_constructor

// TestNoObjectConstructorExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it
// covers, so future refactors can't silently regress them without breaking a
// named lock-in.
//
// N/A Dimension 4 rows:
//   - Access/key forms: the rule never inspects a property or key, only the
//     callee identifier and the argument count.
//   - Declaration/container forms (class/function declaration vs expression):
//     the rule targets Call/New expressions, not declarations.
//   - Overload signatures / `abstract` / `declare` members, empty class/
//     function bodies: none of these are inputs the rule inspects.

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// literalCase builds an InvalidTestCase for a call whose replacement needs no
// wrapping parentheses: it neither starts an expression statement nor opens an
// arrow function's concise body.
func literalCase(code, pattern string) rule_tester.InvalidTestCase {
	return suggestionCase(code, pattern, "{}", "useLiteral", false)
}

func TestNoObjectConstructorExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoObjectConstructorRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: TS non-null assertion callee ----
			// `Object!` is a TSNonNullExpression, never collapsed back to a
			// bare Identifier by ESTree either, so upstream's
			// `node.callee.type !== "Identifier"` check also stays silent.
			{Code: `Object!();`},
			{Code: `new Object!();`},

			// ---- Dimension 4: TS type-expression wrapper callee ----
			{Code: `(Object as any)();`},
			{Code: `new (Object as any)();`},

			// ---- Dimension 4: graceful degradation — spread argument ----
			// A spread counts as one argument, matching upstream's
			// `node.arguments.length` truthiness check.
			{Code: `Object(...args);`},
			{Code: `new Object(...args);`},

			// ---- Dimension 2: scoping — shadowed by function declaration ----
			{Code: `function Object() {} new Object();`},
			// ---- Dimension 2: scoping — shadowed by class declaration ----
			{Code: `class Object {} new Object();`},
			// ---- Dimension 2: scoping — shadowed by nested block `let` ----
			{Code: `{ let Object = class {}; new Object(); }`},
			// ---- Dimension 2: scoping — shadowed by catch parameter ----
			{Code: `try {} catch (Object) { new Object(); }`},
			// ---- Dimension 2: scoping — outer parameter shadow reaches deeply
			// nested arrows ----
			{Code: `function f(Object) { return () => () => new Object(); }`},

			// ---- Dimension 2: scoping — TypeScript type-space declarations.
			// scope-manager gives a type alias, interface, or type parameter an
			// `Object` variable with identifiers, so upstream stays silent even
			// though TypeScript itself still resolves the call to the global. ----
			{Code: `type Object = {}; Object();`},
			{Code: `interface Object {} Object();`},
			{Code: `{ type Object = {}; Object(); }`},
			{Code: `type Object = {}; function g() { Object(); }`},
			{Code: `function f<Object>() { Object(); }`},
			{Code: `class C<Object> { m() { Object(); } }`},
			// ---- Dimension 2: scoping — class type parameter inside a static
			// member, which scope-manager keeps in the lexical scope chain while
			// TypeScript's resolver hides it ----
			{Code: `class C<Object> { static m() { Object(); } }`},
			{Code: `class C<Object> { static p = Object(); }`},

			// ---- Dimension 2: scoping — scope-manager puts a function's
			// parameter initializers in the same scope as its body
			// declarations, so a body binding shadows the call in a default
			// value ----
			{Code: `function f(x = new Object()) { var Object; }`},
			{Code: `class C { m(x = new Object()) { var Object; } }`},
			// A parameter decorator is not an initializer, but scope-manager
			// still resolves a reference directly inside one against the
			// decorated function's own scope.
			{Code: `class C { m(@dec(new Object()) x: number) { var Object; } }`},
			{Code: `class C { m(@dec(new Object()) x: number, Object: any) { } }`},
			// The same function scope holds TypeScript type-space and
			// import-equals declarations.
			{Code: `function f(a = new Object()) { type Object = {}; }`},
			{Code: `function f(a = new Object()) { interface Object {} }`},
			{Code: `function f(a = new Object()) { import Object = require("x"); }`},
			// The enclosing class stays in the chain, so its name and type
			// parameters shadow a call nested inside a parameter decorator.
			{Code: `class Object { m(@dec(() => new Object()) x: number) { } }`},
			{Code: `class C<Object> { m(@dec(() => new Object()) x: number) { } }`},
			// A class's own decorator is evaluated in the scope holding the
			// class, but a reference sitting directly in one, with no scope in
			// between, still resolves against the class scope.
			{Code: `@dec(new Object()) class C<Object> { }`},
			{Code: `const C = @dec(new Object()) class Object { };`},
			// A class declaration's name is declared in the scope holding the
			// class, so it shadows the call wherever the decorator resolves.
			{Code: `@dec(() => new Object()) class Object { }`},

			// ---- Dimension 2: scoping — a member's decorators and computed
			// name are evaluated in the enclosing class or object literal, so
			// only that outer scope shadows a call there ----
			{Code: `class C<Object> { @dec(new Object()) m() { } }`},
			{Code: `class C<Object> { [Object()]() { } }`},
			{Code: `class Object { [new Object()]() { } }`},

			// ---- Dimension 2: scoping — a function declaration with no block
			// to hold it is defined in the innermost scope that already exists
			// at its position ----
			{Code: `function f(a = new Object()) { if (b) function Object() {} }`},
			{Code: `function f(a = new Object()) { lbl: function Object() {} }`},
			{Code: `function f(a = new Object()) { while (b) function Object() {} }`},
			{Code: `function f() { if (b) function Object() {} Object(); }`},
			{Code: `if (b) function Object() {} Object();`},
			{Code: `namespace N { if (b) function Object() {} Object(); }`},

			// ---- Dimension 2: scoping — declarations in a namespace body ----
			{Code: `namespace N { const Object = f; Object(); }`},
			{Code: `namespace N { function Object() {} Object(); }`},
			{Code: `namespace N { namespace Object {} Object(); }`},
			{Code: `namespace N { import Object = require("x"); Object(); }`},
			{Code: `namespace N.M { const Object = f; Object(); }`},
			// A dotted namespace name declares nothing itself, but the body
			// it opens is still a scope of its own.
			{Code: `namespace N.Object { const Object = f; Object(); }`},
			{Code: `namespace N { const Object = f; namespace M.Object { Object(); } }`},
			// An export specifier alongside a local declaration still leaves
			// that declaration in the namespace.
			{Code: `namespace N { const Object = f; export { Object }; Object(); }`},
			{Code: `namespace N { import Object = require("x"); export { Object }; Object(); }`},

			// Locks in upstream check() arm 1: callee is not an Identifier at
			// all (a member access), independent of the NewExpression form
			// upstream's own `new globalThis.Object` case already covers.
			{Code: `Object.assign({}, x);`},
			// Locks in upstream check() arm 2: callee is an Identifier whose
			// name isn't "Object".
			{Code: `Array();`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized callee, single and multi-level ----
			// ESTree flattens grouping parentheses, so `(Object)()` still has
			// an Identifier callee there; tsgo needs an explicit SkipParentheses.
			{
				Code: `(Object)();`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferLiteral", Line: 1, Column: 1, EndLine: 1, EndColumn: 11,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "useLiteral", Output: `({});`}},
				}},
			},
			{
				Code: `((Object))();`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferLiteral", Line: 1, Column: 1, EndLine: 1, EndColumn: 13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "useLiteral", Output: `({});`}},
				}},
			},
			{
				Code: `new (Object)();`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferLiteral", Line: 1, Column: 1, EndLine: 1, EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "useLiteral", Output: `({});`}},
				}},
			},

			// ---- Dimension 4: nesting boundary ----
			// Only the inner, zero-argument call is reported; the outer call
			// has one argument (the inner CallExpression) and stays silent.
			{
				Code: `Object(Object());`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferLiteral", Line: 1, Column: 8, EndLine: 1, EndColumn: 16,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "useLiteral", Output: `Object({});`}},
				}},
			},

			// Locks in upstream needsParentheses(): the call neither starts an
			// ExpressionStatement nor follows an `=>` (it sits inside the
			// parameter's default value), so the fix stays a bare `{}`.
			{
				Code: `const f = (a = Object()) => a;`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferLiteral", Line: 1, Column: 16, EndLine: 1, EndColumn: 24,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "useLiteral", Output: `const f = (a = {}) => a;`}},
				}},
			},

			// Locks in upstream needsParentheses(): the NewExpression form of
			// the arrow's own concise body also needs wrapping parens, not
			// just the CallExpression form upstream's own test covers.
			{
				Code: `const f = () => new Object();`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferLiteral", Line: 1, Column: 17, EndLine: 1, EndColumn: 29,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "useLiteral", Output: `const f = () => ({});`}},
				}},
			},

			// Locks in upstream needsParentheses(): the call only *starts* the
			// concise body, so a bare `{}` would still be parsed as the arrow's
			// block body — the token right before it is the `=>`, which is what
			// upstream tests, not whether the call is the whole body.
			{
				Code: `const f = () => Object().x;`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferLiteral", Line: 1, Column: 17, EndLine: 1, EndColumn: 25,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "useLiteral", Output: `const f = () => ({}).x;`}},
				}},
			},

			// The explicit parentheses already open the concise body, so the
			// token before the call is `(` and the fix stays a bare `{}`.
			{
				Code: `const f = () => (Object().x);`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferLiteral", Line: 1, Column: 18, EndLine: 1, EndColumn: 26,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "useLiteral", Output: `const f = () => ({}.x);`}},
				}},
			},

			// ---- Dimension 2: scoping — unshadowed reference reachable
			// through nested arrows still reports ----
			{
				Code: `function f() { return () => () => Object(); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferLiteral", Line: 1, Column: 35, EndLine: 1, EndColumn: 43,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "useLiteral", Output: `function f() { return () => () => ({}); }`}},
				}},
			},

			// ---- Dimension 2: scoping — a namespace body that declares nothing
			// leaves the call bound to the global; the fix needs wrapping parens
			// because the call starts an ExpressionStatement ----
			{
				Code: `namespace N { Object(); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferLiteral", Line: 1, Column: 15, EndLine: 1, EndColumn: 23,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "useLiteral", Output: `namespace N { ({}); }`}},
				}},
			},
			// A namespace-local declaration doesn't reach the outer scope.
			{
				Code: `namespace N { const Object = f; } Object();`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferLiteral", Line: 1, Column: 35, EndLine: 1, EndColumn: 43,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "useLiteral", Output: `namespace N { const Object = f; } ({});`}},
				}},
			},
			// ---- Dimension 2: scoping — scope-manager attaches a scope
			// created inside a parameter decorator to the enclosing class,
			// not to the decorated function, so neither the function's body
			// bindings nor its parameters and type parameters reach the
			// call ----
			asiCase(`class C { m(@dec(() => new Object()) x: number) { var Object; } }`, "new Object()", false, false),
			// A class property initializer is not an expression statement, so
			// the replacement needs no wrapping parentheses.
			{
				Code: `class C { m(@dec(class { p = new Object(); }) x: number) { var Object; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferLiteral", Line: 1, Column: 30, EndLine: 1, EndColumn: 42,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "useLiteral",
						Output:    `class C { m(@dec(class { p = {}; }) x: number) { var Object; } }`,
					}},
				}},
			},
			asiCase(`class C { m(@dec(() => new Object()) x: number, Object: any) { } }`, "new Object()", false, false),
			asiCase(`class C { m<Object>(@dec(() => new Object()) x: number) { } }`, "new Object()", false, false),

			// ---- Dimension 2: scoping — the same holds for a class's own
			// decorator: a scope created inside one is attached to the scope
			// holding the class, so neither the class's type parameters nor a
			// class expression's own name reach the call ----
			asiCase(`@dec(() => new Object()) class C<Object> { }`, "new Object()", false, false),
			literalCase(`@dec(function () { return new Object(); }) class C<Object> { }`, "new Object()"),
			literalCase(`@dec(class { p = new Object(); }) class C<Object> { }`, "new Object()"),
			asiCase(`const C = @dec(() => new Object()) class Object { };`, "new Object()", false, false),
			literalCase(`const C = @dec(class { p = new Object(); }) class Object { };`, "new Object()"),

			// ---- Dimension 2: scoping — a decorator or computed name belongs
			// to the enclosing class or object literal, not to the member that
			// carries it, so the member's own type parameters, parameters, and
			// body bindings never reach the call ----
			literalCase(`class C { @dec(new Object()) m() { var Object; } }`, "new Object()"),
			literalCase(`class C { @dec(new Object()) m(Object: any) { } }`, "new Object()"),
			literalCase(`class C { @dec(new Object()) m<Object>() { } }`, "new Object()"),
			literalCase(`class C { @dec(Object()) get m() { var Object; return 1; } }`, "Object()"),
			literalCase(`class C { [new Object()]() { var Object; } }`, "new Object()"),
			literalCase(`class C { [Object()](Object: any) { } }`, "Object()"),
			literalCase(`class C { [Object()]<Object>() { } }`, "Object()"),
			literalCase(`const o = { [Object()]() { var Object; } };`, "Object()"),

			// ---- Dimension 2: scoping — a block or a `let`-scoped loop does
			// hold the function declaration, keeping it out of the function
			// scope the parameter initializer resolves against ----
			literalCase(`function f(a = new Object()) { { function Object() {} } }`, "new Object()"),
			literalCase(`function f(a = new Object()) { for (let i; ;) function Object() {} }`, "new Object()"),
			{
				Code: `function f() { { function Object() {} } Object(); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferLiteral", Line: 1, Column: 41, EndLine: 1, EndColumn: 49,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "useLiteral",
						Output:    `function f() { { function Object() {} } ({}); }`,
					}},
				}},
			},

			// ---- Dimension 2: scoping — a namespace re-export names what the
			// namespace exports without declaring it inside, so scope-manager
			// creates no variable and the call still reaches the global ----
			asiCase(`namespace N { export { Object } from "x"; Object(); }`, "Object()", false, false),
			asiCase(`namespace N { export type { Object } from "x"; Object(); }`, "Object()", false, false),
			asiCase(`namespace N { export * as Object from "x"; Object(); }`, "Object()", false, false),
			asiCase(`namespace N { export { Other as Object } from "x"; Object(); }`, "Object()", false, false),
			asiCase(`namespace N { export { Object }; Object(); }`, "Object()", false, false),

			// ---- Dimension 2: scoping — only a namespace whose name is a
			// plain identifier declares a variable, so neither segment of a
			// dotted name shadows the call ----
			asiCase(`namespace N.Object { Object(); }`, "Object()", false, false),
			asiCase(`namespace Object.M { Object(); }`, "Object()", false, false),
			asiCase(`namespace N { namespace M.Object { } Object(); }`, "Object()", false, false),

			// A type parameter is only in scope inside its own declaration.
			{
				Code: `declare function f<Object>(): void; Object();`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferLiteral", Line: 1, Column: 37, EndLine: 1, EndColumn: 45,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "useLiteral", Output: `declare function f<Object>(): void; ({});`}},
				}},
			},

			// ---- Real-user: eslint/eslint#17672 (JSX before Object(), fixed
			// via a preceding-semicolon suggestion) — NewExpression variant,
			// which upstream's own regression test never exercises (it only
			// covers the CallExpression form for the JSX cases) ----
			asiCase("\n<foo />\nnew Object()\n", "new Object()", true, true),

			// ---- Real-user: eslint/eslint#17649 (`++index` before Object(),
			// autofix produced a syntax error before the semicolon fix) —
			// bare `new Object` (no parens) variant, which upstream's own
			// regression test never exercises (it only covers `Object()`) ----
			asiCase("\n++index\nnew Object\n", "new Object", true, false),

			// ---- Contract: documented divergence — utils.NeedsPrecedingSemicolon
			// doesn't model TypeScript-only node kinds (see its doc comment and
			// the NOTE in no_object_constructor.go), so it falls back to the
			// conservative "needs a semicolon" answer after a type alias. This
			// mirrors the port-rule/no_array_constructor lock-in for the same
			// shared-utility gap — upstream's own no-object-constructor test
			// suite has no TypeScript-parser block to catch it, but the rule
			// hits the identical code path since it shares the helper. ----
			{
				Code: "type T = Foo\nObject()",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferLiteral", Line: 2, Column: 1, EndLine: 2, EndColumn: 9,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "useLiteralAfterSemicolon",
						Output:    "type T = Foo\n;({})",
					}},
				}},
			},
		},
	)
}

// TestNoObjectConstructorEditDemand verifies that the suggestion builder does
// not change what the rule reports: diagnostic count, message, and range stay
// identical across every edit demand, and the suggestion is materialized only
// when it was requested.
func TestNoObjectConstructorEditDemand(t *testing.T) {
	t.Parallel()

	const source = "() => Object();\n"

	program, sourceFile := createNoObjectConstructorProgram(t, "edit-demand.ts", source)
	options := rule_tester.ResolveTestCaseOptions(t, &NoObjectConstructorRule, nil)

	diagnostics := make(map[rule.EditDemand]rule.RuleDiagnostic, 4)
	for _, demand := range []rule.EditDemand{
		rule.EditDemandNone,
		rule.EditDemandAutofix,
		rule.EditDemandSuggestion,
		rule.EditDemandAll,
	} {
		got := lintNoObjectConstructorWithDemand(program, sourceFile, options, demand)
		if len(got) != 1 {
			t.Fatalf("demand %d: diagnostics = %d, want 1", demand, len(got))
		}
		if got[0].Message.Id != "preferLiteral" {
			t.Errorf("demand %d: unexpected message id %q", demand, got[0].Message.Id)
		}
		if got[0].Message.Description != "The object literal notation {} is preferable." {
			t.Errorf("demand %d: unexpected message %q", demand, got[0].Message.Description)
		}
		diagnostics[demand] = got[0]
	}

	diagnosticsOnly := diagnostics[rule.EditDemandNone]
	for demand, diagnostic := range diagnostics {
		want, got := diagnosticsOnly, diagnostic
		want.FixesPtr, want.Suggestions = nil, nil
		got.FixesPtr, got.Suggestions = nil, nil
		if !reflect.DeepEqual(got, want) {
			t.Errorf("demand %d changed diagnostic metadata:\ngot:  %#v\nwant: %#v", demand, got, want)
		}
	}

	// The rule offers suggestions only, so neither the diagnostics-only nor
	// the autofix demand may materialize anything.
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix} {
		diagnostic := diagnostics[demand]
		if diagnostic.FixesPtr != nil || diagnostic.Suggestions != nil {
			t.Errorf(
				"demand %d unexpectedly materialized edits: fixes=%#v suggestions=%#v",
				demand,
				diagnostic.FixesPtr,
				diagnostic.Suggestions,
			)
		}
	}

	for _, demand := range []rule.EditDemand{rule.EditDemandSuggestion, rule.EditDemandAll} {
		diagnostic := diagnostics[demand]
		if diagnostic.FixesPtr != nil {
			t.Errorf("demand %d unexpectedly materialized autofixes: %#v", demand, diagnostic.FixesPtr)
		}
		if diagnostic.Suggestions == nil || len(*diagnostic.Suggestions) != 1 {
			t.Fatalf("demand %d: suggestions = %#v, want exactly one", demand, diagnostic.Suggestions)
		}
		suggestion := (*diagnostic.Suggestions)[0]
		if suggestion.Message.Id != "useLiteral" {
			t.Errorf("demand %d: unexpected suggestion id %q", demand, suggestion.Message.Id)
		}
		fixes := suggestion.Fixes()
		if len(fixes) != 1 || fixes[0].Text != "({})" {
			t.Errorf("demand %d: unexpected suggestion fixes %#v", demand, fixes)
		}
	}
}

func lintNoObjectConstructorWithDemand(
	program *compiler.Program,
	sourceFile *ast.SourceFile,
	options []any,
	demand rule.EditDemand,
) []rule.RuleDiagnostic {
	var diagnostics []rule.RuleDiagnostic
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program:         lintprogram.NewFromCompiler(program),
		File:            sourceFile.FileName(),
		HasTypeInfo:     true,
		GetRulesForFile: noObjectConstructorConfiguredRules(options),
		Consumer: rule.DiagnosticConsumer{
			Demand: demand,
			Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			},
		},
	})
	return diagnostics
}

func noObjectConstructorConfiguredRules(options []any) func(*ast.SourceFile) []rule.ConfiguredRule {
	return func(*ast.SourceFile) []rule.ConfiguredRule {
		return []rule.ConfiguredRule{{
			Name:     NoObjectConstructorRule.Name,
			Severity: rule.SeverityError,
			Run: func(ctx rule.RuleContext) rule.RuleListeners {
				return NoObjectConstructorRule.Run(ctx, options)
			},
		}}
	}
}

func createNoObjectConstructorProgram(t testing.TB, fileName string, code string) (*compiler.Program, *ast.SourceFile) {
	t.Helper()

	root := fixtures.GetRootDir()
	fs := utils.NewOverlayVFS(root.FS, map[string]string{tspath.ResolvePath(root.Dir, fileName): code})
	host := utils.CreateCompilerHost(root.Dir, fs)
	program, err := utils.CreateProgram(true, fs, root.Dir, "tsconfig.json", host)
	if err != nil {
		t.Fatalf("failed to create program: %v", err)
	}
	sourceFile := program.GetSourceFile(fileName)
	if sourceFile == nil {
		t.Fatalf("source file %q not found", fileName)
	}
	return program, sourceFile
}
