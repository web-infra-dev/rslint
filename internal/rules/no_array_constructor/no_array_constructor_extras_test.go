package no_array_constructor

// TestNoArrayConstructorExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it
// covers, so future refactors can't silently regress them without breaking a
// named lock-in.
//
// N/A Dimension 4 rows:
//   - Access/key forms: the rule never inspects a property or key, only the
//     callee identifier, its type arguments, and the argument list.
//   - Declaration/container forms (class/function declaration vs expression):
//     the rule targets Call/New expressions, not declarations.
//   - SpreadAssignment in an object literal / RestElement in a binding
//     pattern: the rule never inspects object literals or binding patterns.
//   - Overload signatures / `abstract` / `declare` members, empty class/
//     function bodies: already exercised by the upstream-migrated
//     "no semicolon required after TypeScript syntax" block
//     (`declare function foo()`, `function foo(): []`, ...).

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

func TestNoArrayConstructorExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoArrayConstructorRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: TS non-null assertion callee ----
			// `Array!` is a TSNonNullExpression, never collapsed back to a
			// bare Identifier by ESTree either, so upstream's
			// `node.callee.type !== "Identifier"` check also stays silent.
			{Code: `Array!();`},
			{Code: `new Array!();`},

			// ---- Dimension 4: TS type-expression wrapper callee ----
			{Code: `(Array as any)();`},
			{Code: `(Array satisfies Function)();`},

			// ---- Dimension 2: scoping — shadowed by function declaration ----
			{Code: `function Array() {} new Array();`},
			// ---- Dimension 2: scoping — shadowed by class declaration ----
			{Code: `class Array {} new Array();`},
			// ---- Dimension 2: scoping — shadowed by nested block `let` ----
			{Code: `{ let Array = class {}; new Array(); }`},
			// ---- Dimension 2: scoping — shadowed by catch parameter ----
			{Code: `try {} catch (Array) { new Array(); }`},
			// ---- Dimension 2: scoping — shadowed by a default import ----
			{Code: `import Array from "my-array"; new Array();`},
			// ---- Dimension 2: scoping — outer parameter shadow reaches
			// deeply nested arrows ----
			{Code: `function f(Array) { return () => () => new Array(); }`},

			// ---- Dimension 2: scoping — TypeScript type-space declarations.
			// scope-manager gives a type alias, interface, type parameter, or
			// type-only import an `Array` variable with identifiers, so upstream
			// stays silent even though TypeScript itself still resolves the call
			// to the global. ----
			{Code: `type Array = {}; Array();`},
			{Code: `interface Array {} Array();`},
			{Code: `import type { Array } from "my-array"; Array();`},
			{Code: `function f<Array>() { Array(); }`},
			{Code: `class C<Array> { m() { Array(); } }`},
			// ---- Dimension 2: scoping — class type parameter inside a static
			// member, which scope-manager keeps in the lexical scope chain while
			// TypeScript's resolver hides it ----
			{Code: `class C<Array> { static m() { Array(); } }`},

			// ---- Dimension 2: scoping — scope-manager puts a function's
			// parameter initializers in the same scope as its body declarations,
			// so a body binding shadows the call in a default value ----
			{Code: `function f(x = Array()) { var Array; }`},
			{Code: `function f(x = Array()) { let Array; }`},
			{Code: `function f(x = Array()) { function Array() {} }`},
			{Code: `function f(x = Array()) { class Array {} }`},
			// A `var` reaches the function scope from any depth.
			{Code: `function f(x = Array()) { if (a) { var Array; } }`},
			// The shadow also reaches a call nested inside the default value.
			{Code: `function f(x = () => Array()) { var Array; }`},
			{Code: `const g = (x = Array()) => { var Array; };`},
			// A parameter decorator is not an initializer, but scope-manager
			// still resolves a reference directly inside one against the
			// decorated function's own scope.
			{Code: `class C { m(@dec(Array()) x: number) { var Array; } }`},
			{Code: `class C { m(@dec(Array()) x: number) { let Array; } }`},
			// The enclosing class stays in the chain, so its name and type
			// parameters shadow a call nested inside a parameter decorator.
			{Code: "class Array { m(@dec(() => Array()) x: number) { } }"},
			{Code: "class C<Array> { m(@dec(() => Array()) x: number) { } }"},

			// ---- Dimension 2: scoping — a TypeScript namespace body is a
			// scope of its own, so a declaration inside it shadows the global
			// constructor for the whole module block ----
			{Code: `namespace N { const Array = () => 123; Array(); }`},
			{Code: `namespace N { function Array() {} new Array(); }`},
			{Code: `namespace N { import Array = globalThis.Array; Array(); }`},
			// A `var` in a namespace hoists to the module block, not out of it.
			{Code: `namespace N { var Array = 1; { new Array(); } }`},
			// The shadow reaches into nested namespaces, which are ordinary
			// lexical children of the outer module block.
			{Code: `namespace N { const Array = () => 1; namespace M { Array(); } }`},

			// Real-user: eslint/eslint#12273 (closed as intentional — a
			// single non-spread argument is left alone even when it's
			// obviously not a length, since the rule has no type
			// information to tell `Array('3')` apart from `Array(3)`).
			{Code: `Array('3');`},

			// Real-user: eslint/eslint#19494 (open at the time of porting —
			// a known upstream false negative). `new (0, Array)()` calls the
			// global Array constructor via a comma-operator callee, but the
			// callee is a BinaryExpression, not an Identifier, so neither
			// upstream nor this port detects it. Locked in to match
			// upstream's current (imperfect) behavior rather than silently
			// diverging from it.
			{Code: `new (0, Array)();`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized callee, multi-level ----
			// Upstream's own test only covers single-level `(Array)()`.
			directFixCase(`((Array))();`, `((Array))()`, `[];`),

			// ---- Dimension 2 / nesting boundary: only the inner, zero-
			// argument call is reported; the outer call has exactly one
			// non-spread argument (the inner CallExpression), so upstream's
			// own single-argument exception naturally excludes it — this
			// also proves the CallExpression listener doesn't misfire on
			// the outer node just because the inner one matched. ----
			directFixCase(`Array(Array());`, `Array()`, `Array([]);`),
			// Same boundary, NewExpression form on both sides — proves the
			// NewExpression listener is independent of the CallExpression
			// one and doesn't cross-report.
			directFixCase(`new Array(new Array());`, `new Array()`, `new Array([]);`),

			// Locks in the `nonSpreadCount` reduce (not just a length check):
			// three arguments where two are non-spread is still safe to
			// autofix directly, matching upstream's `Array(5, 6, ...args)`
			// case but with the spread positioned first instead of last.
			directFixCase(`Array(1, ...a, 2, ...b);`, `Array(1, ...a, 2, ...b)`, `[1, ...a, 2, ...b];`),
			// Inverse: only one non-spread argument among three — stays
			// suggestion-only, matching upstream's `Array(5, ...args)` case
			// but with the non-spread argument last instead of first.
			suggestCase(`Array(...a, 1, ...b);`, `Array(...a, 1, ...b)`, "useLiteral", `[...a, 1, ...b];`),

			// ---- Contract: exact diagnostic message text ----
			// The rule only ever emits one top-level messageId
			// ("preferLiteral"); asserted with an exact string match here.
			{
				Code:   `new Array();`,
				Output: []string{`[];`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferLiteral",
					Message:   "The array literal notation [] is preferable.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 12,
				}},
			},

			// ---- Dimension 2: scoping — a `let` in a nested block stays in
			// that block's scope, so it never reaches the parameter initializer ----
			directFixCase(
				`function f(x = Array()) { { let Array; } }`,
				`Array()`,
				`function f(x = []) { { let Array; } }`,
			),
			// A body that declares nothing leaves the default value bound to
			// the global.
			directFixCase(
				`function f(x = Array()) { }`,
				`Array()`,
				`function f(x = []) { }`,
			),
			// ---- Dimension 2: scoping — scope-manager attaches a scope
			// created inside a parameter decorator to the enclosing class,
			// not to the decorated function, so the function's own body
			// bindings never reach the call ----
			directFixCase(
				"class C { m(@dec(() => Array()) x: number) { var Array; } }",
				`Array()`,
				"class C { m(@dec(() => []) x: number) { var Array; } }",
			),
			directFixCase(
				"class C { m(@dec(function () { return Array(); }) x: number) { var Array; } }",
				`Array()`,
				"class C { m(@dec(function () { return []; }) x: number) { var Array; } }",
			),
			directFixCase(
				"class C { m(@dec(class { p = Array(); }) x: number) { var Array; } }",
				`Array()`,
				"class C { m(@dec(class { p = []; }) x: number) { var Array; } }",
			),
			// The decorated function's parameters and type parameters are out
			// of the chain for the same reason.
			directFixCase(
				"class C { m(@dec(() => Array()) x: number, Array: any) { } }",
				`Array()`,
				"class C { m(@dec(() => []) x: number, Array: any) { } }",
			),
			directFixCase(
				"class C { m<Array>(@dec(() => Array()) x: number) { } }",
				`Array()`,
				"class C { m<Array>(@dec(() => []) x: number) { } }",
			),

			// A type parameter is only in scope inside its own declaration.
			directFixCase(
				`declare function f<Array>(): void; Array();`,
				`Array()`,
				`declare function f<Array>(): void; [];`,
			),

			// ---- Dimension 2: scoping — unshadowed reference reachable
			// through nested arrows still reports ----
			directFixCase(
				`function f() { return () => () => Array(); }`,
				`Array()`,
				`function f() { return () => () => []; }`,
			),
		},
	)
}

// TestNoArrayConstructorEditDemand verifies that the fix/suggestion builders
// don't change what the rule reports: diagnostic count, message, and range
// stay identical across every edit demand, and fixes/suggestions are
// materialized only when requested — for both branches of the rule's
// fix-vs-suggestion split.
func TestNoArrayConstructorEditDemand(t *testing.T) {
	t.Parallel()

	t.Run("direct fix", func(t *testing.T) {
		t.Parallel()
		testNoArrayConstructorEditDemand(t, "() => Array();\n", false)
	})
	t.Run("suggestion only", func(t *testing.T) {
		t.Parallel()
		testNoArrayConstructorEditDemand(t, "() => Array?.();\n", true)
	})
}

func testNoArrayConstructorEditDemand(t *testing.T, source string, wantSuggestion bool) {
	t.Helper()

	program, sourceFile := createNoArrayConstructorProgram(t, "edit-demand.ts", source)
	options := rule_tester.ResolveTestCaseOptions(t, &NoArrayConstructorRule, nil)

	diagnostics := make(map[rule.EditDemand]rule.RuleDiagnostic, 4)
	for _, demand := range []rule.EditDemand{
		rule.EditDemandNone,
		rule.EditDemandAutofix,
		rule.EditDemandSuggestion,
		rule.EditDemandAll,
	} {
		got := lintNoArrayConstructorWithDemand(program, sourceFile, options, demand)
		if len(got) != 1 {
			t.Fatalf("demand %d: diagnostics = %d, want 1", demand, len(got))
		}
		if got[0].Message.Id != "preferLiteral" {
			t.Errorf("demand %d: unexpected message id %q", demand, got[0].Message.Id)
		}
		if got[0].Message.Description != "The array literal notation [] is preferable." {
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

	// EditDemandNone never materializes fixes or suggestions regardless of
	// which branch the rule took.
	if d := diagnostics[rule.EditDemandNone]; d.FixesPtr != nil || d.Suggestions != nil {
		t.Errorf("EditDemandNone unexpectedly materialized edits: fixes=%#v suggestions=%#v", d.FixesPtr, d.Suggestions)
	}

	if wantSuggestion {
		for _, demand := range []rule.EditDemand{rule.EditDemandAutofix} {
			if d := diagnostics[demand]; d.FixesPtr != nil || d.Suggestions != nil {
				t.Errorf("demand %d unexpectedly materialized edits: fixes=%#v suggestions=%#v", demand, d.FixesPtr, d.Suggestions)
			}
		}
		for _, demand := range []rule.EditDemand{rule.EditDemandSuggestion, rule.EditDemandAll} {
			d := diagnostics[demand]
			if d.FixesPtr != nil {
				t.Errorf("demand %d unexpectedly materialized autofixes: %#v", demand, d.FixesPtr)
			}
			if d.Suggestions == nil || len(*d.Suggestions) != 1 {
				t.Fatalf("demand %d: suggestions = %#v, want exactly one", demand, d.Suggestions)
			}
			suggestion := (*d.Suggestions)[0]
			if suggestion.Message.Id != "useLiteral" {
				t.Errorf("demand %d: unexpected suggestion id %q", demand, suggestion.Message.Id)
			}
			fixes := suggestion.Fixes()
			if len(fixes) != 1 || fixes[0].Text != "[]" {
				t.Errorf("demand %d: unexpected suggestion fixes %#v", demand, fixes)
			}
		}
		return
	}

	for _, demand := range []rule.EditDemand{rule.EditDemandSuggestion} {
		if d := diagnostics[demand]; d.FixesPtr != nil || d.Suggestions != nil {
			t.Errorf("demand %d unexpectedly materialized edits: fixes=%#v suggestions=%#v", demand, d.FixesPtr, d.Suggestions)
		}
	}
	for _, demand := range []rule.EditDemand{rule.EditDemandAutofix, rule.EditDemandAll} {
		d := diagnostics[demand]
		if d.Suggestions != nil {
			t.Errorf("demand %d unexpectedly materialized suggestions: %#v", demand, d.Suggestions)
		}
		if d.FixesPtr == nil || len(*d.FixesPtr) != 1 || (*d.FixesPtr)[0].Text != "[]" {
			t.Errorf("demand %d: unexpected autofix %#v", demand, d.FixesPtr)
		}
	}
}

func lintNoArrayConstructorWithDemand(
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
		GetRulesForFile: noArrayConstructorConfiguredRules(options),
		ExcludePaths:    []string{},
		Consumer: rule.DiagnosticConsumer{
			Demand: demand,
			Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			},
		},
	})
	return diagnostics
}

func noArrayConstructorConfiguredRules(options []any) func(*ast.SourceFile) []linter.ConfiguredRule {
	return func(*ast.SourceFile) []linter.ConfiguredRule {
		return []linter.ConfiguredRule{{
			Name:     NoArrayConstructorRule.Name,
			Severity: rule.SeverityError,
			Run: func(ctx rule.RuleContext) rule.RuleListeners {
				return NoArrayConstructorRule.Run(ctx, options)
			},
		}}
	}
}

func createNoArrayConstructorProgram(t testing.TB, fileName string, code string) (*compiler.Program, *ast.SourceFile) {
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
