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

func noObjectConstructorConfiguredRules(options []any) func(*ast.SourceFile) []linter.ConfiguredRule {
	return func(*ast.SourceFile) []linter.ConfiguredRule {
		return []linter.ConfiguredRule{{
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
