package no_array_constructor

// cspell:ignore rvice rray

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestSourceMayUseArrayConstructor(t *testing.T) {
	if !sourceMayUseArrayConstructor(nil) || !sourceMayUseArrayConstructor(&ast.SourceFile{}) {
		t.Fatal("missing parser identifier metadata must conservatively keep listeners")
	}

	for _, testCase := range []struct {
		name string
		code string
		want bool
	}{
		{name: "ordinary call", code: `service.method()`, want: false},
		{name: "string is conservative", code: `const value = "Array()"`, want: true},
		{name: "comment is conservative", code: `// Array\nservice.method()`, want: true},
		{name: "substring is conservative", code: `TypedArray()`, want: true},
		{name: "unrelated escape", code: `s\u0065rvice.method()`, want: false},
		{name: "escaped string", code: `const value = "\u0041rray()"`, want: false},
		{name: "computed property is conservative", code: `service["Array"]()`, want: true},
		{name: "array call", code: `Array()`, want: true},
		{name: "escaped array identifier", code: `Arr\u0061y()`, want: true},
		{name: "array property", code: `service.Array()`, want: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/source.ts",
				Path:     "/source.ts",
			}, testCase.code, core.ScriptKindTS)
			if got := sourceMayUseArrayConstructor(sourceFile); got != testCase.want {
				t.Fatalf("sourceMayUseArrayConstructor(%q) = %v, want %v", testCase.code, got, testCase.want)
			}
			listeners := NoArrayConstructorRule.Run(rule.RuleContext{SourceFile: sourceFile}, nil)
			if got := len(listeners) != 0; got != testCase.want {
				t.Fatalf("listener presence for %q = %v, want %v", testCase.code, got, testCase.want)
			}
		})
	}
}

func TestNoArrayConstructorExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoArrayConstructorRule,
		[]rule_tester.ValidTestCase{
			// Text filter hits still require an actual Array callee.
			{Code: `const value = "Array()"; service.method();`},
			{Code: `/* Array */ service.method(a, b); new Other();`},
			{Code: `TypedArray(); new ArrayLike(a, b);`},
			{Code: `const value = "\u0041rray()"; s\u0065rvice.method();`},
			{Code: `service["\u0041rray"]();`},

			// One SpreadElement is still exactly one argument upstream.
			{Code: `Array(...values);`},

			// ESTree preserves these TypeScript wrappers around the callee, so
			// unlike grouping parentheses they must not be unwrapped.
			{Code: `Array!();`},
			{Code: `(Array!)();`},
			{Code: `(Array as any)();`},
			{Code: `(Array satisfies ((...args: unknown[]) => unknown))();`},

			// Parenthesizing the identifier must not affect the one-argument
			// exception.
			{Code: `(Array)(value);`},
			{Code: `new (Array)(value);`},
			{Code: `new (Arr\u0061y)(...values);`},
			{Code: `(Arr\u0061y)?.<string>();`},
			{Code: `new (Arr\u0061y)<string>(a, b);`},
		},
		[]rule_tester.InvalidTestCase{
			// new Array (without parentheses)
			{
				Code: `new Array;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useLiteral",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 10,
					},
				},
				Output: []string{`[];`},
			},
			// Nested parentheses - bug test
			{
				Code: `Array((x), y);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useLiteral",
						Line:      1,
						Column:    1,
					},
				},
				Output: []string{`[(x), y];`},
			},
			{
				Code: `Array(foo(), bar());`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useLiteral",
						Line:      1,
						Column:    1,
					},
				},
				Output: []string{`[foo(), bar()];`},
			},
			// SourceFile.HasIdentifier and the AST both normalize identifier
			// escapes, so the file-level fast path must retain these reports.
			{
				Code: `Arr\u0061y(a, b);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useLiteral", Line: 1, Column: 1},
				},
				Output: []string{`[a, b];`},
			},
			{
				Code: `new Arr\u0061y();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useLiteral", Line: 1, Column: 1},
				},
				Output: []string{`[];`},
			},

			// This typescript-eslint wrapper is intentionally syntactic: a
			// local binding named Array still matches.
			{
				Code: `const Array = (...values: unknown[]) => values; Array(a, b);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useLiteral", Line: 1, Column: 49},
				},
				Output: []string{`const Array = (...values: unknown[]) => values; [a, b];`},
			},
			{
				Code: `function make(Array: any) { return new Array(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useLiteral", Line: 1, Column: 36},
				},
				Output: []string{`function make(Array: any) { return []; }`},
			},

			// @typescript-eslint/parser erases grouping parentheses around a
			// callee. tsgo retains them, so the Go rule unwraps only this shape.
			{
				Code: `(Array)();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useLiteral", Line: 1, Column: 1},
				},
				Output: []string{`[];`},
			},
			{
				Code: `((Array))(a, b);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useLiteral", Line: 1, Column: 1},
				},
				Output: []string{`[a, b];`},
			},
			{
				Code: `new (Array);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useLiteral", Line: 1, Column: 1},
				},
				Output: []string{`[];`},
			},
			{
				Code: `new ((Array))(a, b);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useLiteral", Line: 1, Column: 1},
				},
				Output: []string{`[a, b];`},
			},

			// Boundary-only fixes must preserve everything inside the argument
			// list exactly, including trivia that NodeList.End excludes.
			{
				Code: `Array(/* keep */);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useLiteral", Line: 1, Column: 1},
				},
				Output: []string{`[/* keep */];`},
			},
			{
				Code: `Array(a, b, /* keep before close */);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useLiteral", Line: 1, Column: 1},
				},
				Output: []string{`[a, b, /* keep before close */];`},
			},
			{
				Code: `Array(
  a,
  b,
  // keep the closing-line trivia
);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useLiteral", Line: 1, Column: 1},
				},
				Output: []string{`[
  a,
  b,
  // keep the closing-line trivia
];`},
			},
			{
				Code: `Array /* remove with callee */ (a, b);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useLiteral", Line: 1, Column: 1},
				},
				Output: []string{`[a, b];`},
			},
			{
				Code: "Array(/* fake delimiters: ( ) */ foo(\")\"), /[(]/, `()`);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useLiteral", Line: 1, Column: 1},
				},
				Output: []string{"[/* fake delimiters: ( ) */ foo(\")\"), /[(]/, `()`];"},
			},
			{
				Code: "#!/usr/bin/env node\r\nArray(\r\n  a,\r\n  b,\r\n);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useLiteral", Line: 2, Column: 1},
				},
				Output: []string{"#!/usr/bin/env node\r\n[\r\n  a,\r\n  b,\r\n];"},
			},
			{
				Code: `const π = 1; Array(π, 2);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useLiteral", Line: 1, Column: 14},
				},
				Output: []string{`const π = 1; [π, 2];`},
			},

			// Overlapping outer/inner diagnostics are applied atomically over
			// separate fixer passes.
			{
				Code: `Array(Array(a, b), c);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useLiteral", Line: 1, Column: 1},
					{MessageId: "useLiteral", Line: 1, Column: 7},
				},
				Output: []string{
					`[Array(a, b), c];`,
					`[[a, b], c];`,
				},
			},
		},
	)
}

func TestNoArrayConstructorEscapedNames(t *testing.T) {
	// Every character can be literal, a fixed-width escape, or a braced escape.
	// Exercise all 243 spellings without adding a separate decoder to the rule.
	var valid []rule_tester.ValidTestCase
	var invalid []rule_tester.InvalidTestCase
	for variant := range 243 {
		var name strings.Builder
		remaining := variant
		for _, char := range "Array" {
			switch remaining % 3 {
			case 0:
				name.WriteRune(char)
			case 1:
				fmt.Fprintf(&name, `\u%04x`, char)
			case 2:
				fmt.Fprintf(&name, `\u{%x}`, char)
			}
			remaining /= 3
		}
		identifier := name.String()
		valid = append(valid, rule_tester.ValidTestCase{
			Code: identifier + "(value); new " + identifier + "<string>();",
		})
		for _, code := range []string{
			"(" + identifier + ")?.(a, b);",
			"new ((" + identifier + "))(a, b);",
		} {
			invalid = append(invalid, rule_tester.InvalidTestCase{
				Code: code,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "useLiteral",
					Message:   "The array literal notation [] is preferable.",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: len(code),
				}},
				Output: []string{"[a, b];"},
			})
		}
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoArrayConstructorRule, valid, invalid)
}

func TestBuildArrayConstructorFixesRecoveryAST(t *testing.T) {
	const source = "Array(a, b"
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/no-array-constructor-recovery.ts",
		Path:     "/no-array-constructor-recovery.ts",
	}, source, core.ScriptKindTS)

	var call *ast.Node
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindCallExpression {
			call = node
			return true
		}
		return node.ForEachChild(visit)
	}
	sourceFile.AsNode().ForEachChild(visit)
	if call == nil {
		t.Fatal("recovery fixture has no call expression")
	}

	reportRange := core.NewTextRange(0, call.End())
	fixes := buildArrayConstructorFixes(source, call, call.AsCallExpression().Arguments, reportRange)
	if len(fixes) != 1 || fixes[0].Text != "[]" || fixes[0].Range != reportRange {
		t.Fatalf("recovery fixes = %#v, want one full-range replacement", fixes)
	}
}

func TestNoArrayConstructorEditDemand(t *testing.T) {
	t.Parallel()

	const suppressed = "// eslint-disable-next-line @typescript-eslint/no-array-constructor\n" +
		"Array();\n" +
		"new Array(a, b); // rslint-disable-line @typescript-eslint/no-array-constructor\n" +
		"/* eslint-disable @typescript-eslint/no-array-constructor */\n" +
		"Arr\\u0061y();\nnew Array;\n" +
		"/* eslint-enable @typescript-eslint/no-array-constructor */\n"
	const source = suppressed + "const first = Array(a, b);\nnew Array;"
	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(
		source,
		"no-array-constructor-edit-demand.ts",
		"tsconfig.json",
	)
	if err != nil {
		t.Fatal(err)
	}

	run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		t.Helper()

		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program:     lintprogram.NewFromCompiler(program),
			File:        sourceFile.FileName(),
			HasTypeInfo: true,
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{
					Name:     NoArrayConstructorRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return NoArrayConstructorRule.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{
				Demand: demand,
				Report: func(diagnostic rule.RuleDiagnostic) {
					diagnostics = append(diagnostics, diagnostic)
				},
			},
		})
		if len(diagnostics) != 2 {
			t.Fatalf("demand %d: diagnostics = %d, want 2", demand, len(diagnostics))
		}
		return diagnostics
	}

	diagnosticsOnly := run(rule.EditDemandNone)
	autofixOnly := run(rule.EditDemandAutofix)
	suggestionOnly := run(rule.EditDemandSuggestion)
	allEdits := run(rule.EditDemandAll)

	withoutEdits := func(diagnostic rule.RuleDiagnostic) rule.RuleDiagnostic {
		diagnostic.FixesPtr = nil
		diagnostic.Suggestions = nil
		return diagnostic
	}
	for index := range allEdits {
		wantIdentity := withoutEdits(allEdits[index])
		for demand, diagnostics := range map[rule.EditDemand][]rule.RuleDiagnostic{
			rule.EditDemandNone:       diagnosticsOnly,
			rule.EditDemandAutofix:    autofixOnly,
			rule.EditDemandSuggestion: suggestionOnly,
		} {
			if got := withoutEdits(diagnostics[index]); !reflect.DeepEqual(got, wantIdentity) {
				t.Errorf(
					"demand %d changed diagnostic %d:\ngot  %#v\nwant %#v",
					demand,
					index,
					got,
					wantIdentity,
				)
			}
		}

		if diagnosticsOnly[index].FixesPtr != nil || suggestionOnly[index].FixesPtr != nil {
			t.Fatalf("diagnostic %d: non-autofix demand materialized fixes", index)
		}
		if autofixOnly[index].FixesPtr == nil ||
			!reflect.DeepEqual(autofixOnly[index].FixesPtr, allEdits[index].FixesPtr) {
			t.Fatalf("diagnostic %d: autofix and all-edits demands produced different fixes", index)
		}
		for _, diagnostics := range [][]rule.RuleDiagnostic{
			diagnosticsOnly,
			autofixOnly,
			suggestionOnly,
			allEdits,
		} {
			if diagnostics[index].Suggestions != nil {
				t.Fatalf("diagnostic %d: autofix-only rule materialized suggestions", index)
			}
		}
	}

	callFixes := allEdits[0].Fixes()
	if len(callFixes) != 2 {
		t.Fatalf("call fixes = %#v, want two boundary replacements", callFixes)
	}
	if got := source[callFixes[0].Range.Pos():callFixes[0].Range.End()]; got != "Array(" {
		t.Fatalf("left boundary replaces %q, want %q", got, "Array(")
	}
	if got := source[callFixes[1].Range.Pos():callFixes[1].Range.End()]; got != ")" {
		t.Fatalf("right boundary replaces %q, want %q", got, ")")
	}

	noParensFixes := allEdits[1].Fixes()
	if len(noParensFixes) != 1 {
		t.Fatalf("no-parens fixes = %#v, want one fallback replacement", noParensFixes)
	}
	if got := source[noParensFixes[0].Range.Pos():noParensFixes[0].Range.End()]; got != "new Array" {
		t.Fatalf("fallback replaces %q, want %q", got, "new Array")
	}

	fixed, unapplied, changed := linter.ApplyRuleFixes(source, allEdits)
	if !changed || len(unapplied) != 0 {
		t.Fatalf("ApplyRuleFixes changed=%v unapplied=%d", changed, len(unapplied))
	}
	if want := suppressed + "const first = [a, b];\n[];"; fixed != want {
		t.Fatalf("fixed source = %q, want %q", fixed, want)
	}
}
