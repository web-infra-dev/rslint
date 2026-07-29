package no_array_constructor

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoArrayConstructorRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoArrayConstructorRule, []rule_tester.ValidTestCase{
		// Single argument (creates array with size)
		{Code: `new Array(x);`},
		{Code: `Array(x);`},
		{Code: `new Array(9);`},
		{Code: `Array(9);`},

		// Namespaced (not global Array)
		{Code: `new foo.Array();`},
		{Code: `foo.Array();`},
		{Code: `new Array.foo();`},
		{Code: `Array.foo();`},

		// TypeScript with type arguments
		{Code: `new Array<Foo>(1, 2, 3);`},
		{Code: `new Array<Foo>();`},
		{Code: `Array<Foo>(1, 2, 3);`},
		{Code: `Array<Foo>();`},

		// Optional chaining with single argument
		{Code: `Array?.(x);`},
		{Code: `Array?.(9);`},
		{Code: `foo?.Array();`},
		{Code: `Array?.foo();`},
		{Code: `foo.Array?.();`},
		{Code: `Array.foo?.();`},
		{Code: `Array?.<Foo>(1, 2, 3);`},
		{Code: `Array?.<Foo>();`},
	}, []rule_tester.InvalidTestCase{
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
		// new Array()
		{
			Code: `new Array();`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useLiteral",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 12,
				},
			},
			Output: []string{`[];`},
		},
		// Array()
		{
			Code: `Array();`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useLiteral",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 8,
				},
			},
			Output: []string{`[];`},
		},
		// Optional chaining with no args
		{
			Code: `Array?.();`,
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
		// new Array with multiple args
		{
			Code: `new Array(x, y);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useLiteral",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 16,
				},
			},
			Output: []string{`[x, y];`},
		},
		// Array with multiple args
		{
			Code: `Array(x, y);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useLiteral",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 12,
				},
			},
			Output: []string{`[x, y];`},
		},
		// Optional chaining with multiple args
		{
			Code: `Array?.(x, y);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useLiteral",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 14,
				},
			},
			Output: []string{`[x, y];`},
		},
		// new Array with numeric args
		{
			Code: `new Array(0, 1, 2);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useLiteral",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 19,
				},
			},
			Output: []string{`[0, 1, 2];`},
		},
		// Array with numeric args
		{
			Code: `Array(0, 1, 2);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useLiteral",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 15,
				},
			},
			Output: []string{`[0, 1, 2];`},
		},
		// Optional chaining with numeric args
		{
			Code: `Array?.(0, 1, 2);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useLiteral",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 17,
				},
			},
			Output: []string{`[0, 1, 2];`},
		},
		// With comments (no args)
		{
			Code: `/* a */ /* b */ Array /* c */ /* d */ /* e */ /* f */?.(); /* g */ /* h */`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useLiteral",
					Line:      1,
					Column:    17,
					EndLine:   1,
					EndColumn: 58,
				},
			},
			Output: []string{`/* a */ /* b */ []; /* g */ /* h */`},
		},
		// With comments (with args)
		{
			Code: `/* a */ /* b */ Array /* c */ /* d */ /* e */ /* f */?.(x, y); /* g */ /* h */`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useLiteral",
					Line:      1,
					Column:    17,
					EndLine:   1,
					EndColumn: 62,
				},
			},
			Output: []string{`/* a */ /* b */ [x, y]; /* g */ /* h */`},
		},
		// Multi-line
		{
			Code: `
new Array(0, 1, 2);
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useLiteral",
					Line:      2,
					Column:    1,
					EndLine:   2,
					EndColumn: 19,
				},
			},
			Output: []string{`
[0, 1, 2];
`},
		},
		// Multi-line with comments
		{
			Code: `
/* a */ /* b */ Array /* c */ /* d */ /* e */ /* f */?.(
  0,
  1,
  2,
); /* g */ /* h */
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useLiteral",
					Line:      2,
					Column:    17,
				},
			},
			Output: []string{`
/* a */ /* b */ [
  0,
  1,
  2,
]; /* g */ /* h */
`},
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
	})
}

func TestNoArrayConstructorExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoArrayConstructorRule,
		[]rule_tester.ValidTestCase{
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
		},
		[]rule_tester.InvalidTestCase{
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
		return
	}

	reportRange := core.NewTextRange(0, call.End())
	fixes := buildArrayConstructorFixes(source, call, call.AsCallExpression().Arguments, reportRange)
	if len(fixes) != 1 || fixes[0].Text != "[]" || fixes[0].Range != reportRange {
		t.Fatalf("recovery fixes = %#v, want one full-range replacement", fixes)
	}
}

func TestNoArrayConstructorEditDemand(t *testing.T) {
	t.Parallel()

	const source = "const first = Array(a, b);\nnew Array;"
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
			Program:      program,
			File:         sourceFile.FileName(),
			HasTypeInfo:  true,
			ExcludePaths: []string{},
			GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
				return []linter.ConfiguredRule{{
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
	if want := "const first = [a, b];\n[];"; fixed != want {
		t.Fatalf("fixed source = %q, want %q", fixed, want)
	}
}
