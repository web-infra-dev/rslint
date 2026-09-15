package prefer_as_const

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferAsConstRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferAsConstRule, []rule_tester.ValidTestCase{
		{Code: "let foo = 'baz' as const;"},
		{Code: "let foo = 1 as const;"},
		{Code: "let foo = { bar: 'baz' as const };"},
		{Code: "let foo = { bar: 1 as const };"},
		{Code: "let foo = { bar: 'baz' };"},
		{Code: "let foo = { bar: 2 };"},
		{Code: "let foo = <bar>'bar';"},
		{Code: "let foo = <string>'bar';"},
		{Code: "let foo = 'bar' as string;"},
		{Code: "let foo = `bar` as `bar`;"},
		{Code: "let foo = `bar` as `foo`;"},
		{Code: "let foo = `bar` as 'bar';"},
		{Code: `let foo = "bar" as 'bar';`},
		{Code: `let foo = '\x62ar' as 'bar';`},
		{Code: `let foo = 1.0 as 1;`},
		{Code: `let foo = -1 as -1;`},
		{Code: "let foo: string = 'bar';"},
		{Code: "let foo: number = 1;"},
		{Code: "let foo: null = null;"},
		{Code: "let foo: 'bar' = baz;"},
		{Code: "let foo = 'bar';"},
		{Code: "let foo: 'bar';"},
		{Code: "let foo = { bar };"},
		{Code: "let foo: 'baz' = 'baz' as const;"},
		{Code: `
			class foo {
				bar = 'baz';
			}
		`},
		{Code: `
			class foo {
				bar: 'baz';
			}
		`},
		{Code: `
			class foo {
				bar;
			}
		`},
		{Code: `
			class foo {
				bar = <baz>'baz';
			}
		`},
		{Code: `
			class foo {
				bar: string = 'baz';
			}
		`},
		{Code: `
			class foo {
				bar: number = 1;
			}
		`},
		{Code: `
			class foo {
				bar = 'baz' as const;
			}
		`},
		{Code: `
			class foo {
				bar = 2 as const;
			}
		`},
		{Code: `
			class foo {
				get bar(): 'bar' {}
				set bar(bar: 'bar') {}
			}
		`},
		{Code: `
			class foo {
				bar = () => 'bar' as const;
			}
		`},
		{Code: `
			type BazFunction = () => 'baz';
			class foo {
				bar: BazFunction = () => 'bar';
			}
		`},
		{Code: `
			class foo {
				bar(): void {}
			}
		`},
	}, []rule_tester.InvalidTestCase{
		{
			Code: "let foo = { bar: 'baz' as 'baz' };",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferConstAssertion",
					Line:      1,
					Column:    27,
				},
			},
			Output: []string{"let foo = { bar: 'baz' as const };"},
		},
		{
			Code: "let foo = { bar: 1 as 1 };",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferConstAssertion",
					Line:      1,
					Column:    23,
				},
			},
			Output: []string{"let foo = { bar: 1 as const };"},
		},
		{
			Code: "let []: 'bar' = 'bar';",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "variableConstAssertion",
					Line:      1,
					Column:    9,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "variableSuggest",
							Output:    "let [] = 'bar' as const;",
						},
					},
				},
			},
		},
		{
			Code: "let foo: 'bar' = 'bar';",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "variableConstAssertion",
					Line:      1,
					Column:    10,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "variableSuggest",
							Output:    "let foo = 'bar' as const;",
						},
					},
				},
			},
		},
		{
			Code: "let foo: 2 = 2;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "variableConstAssertion",
					Line:      1,
					Column:    10,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "variableSuggest",
							Output:    "let foo = 2 as const;",
						},
					},
				},
			},
		},
		{
			Code: "let foo: 'bar' = 'bar' as 'bar';",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferConstAssertion",
					Line:      1,
					Column:    27,
				},
			},
			Output: []string{"let foo: 'bar' = 'bar' as const;"},
		},
		{
			Code: "let foo = <'bar'>'bar';",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferConstAssertion",
					Line:      1,
					Column:    12,
				},
			},
			Output: []string{"let foo = <const>'bar';"},
		},
		{
			Code: "let foo = <4>4;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferConstAssertion",
					Line:      1,
					Column:    12,
				},
			},
			Output: []string{"let foo = <const>4;"},
		},
		{
			Code: "let foo = 'bar' as 'bar';",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferConstAssertion",
					Line:      1,
					Column:    20,
				},
			},
			Output: []string{"let foo = 'bar' as const;"},
		},
		{
			Code: "let foo = 5 as 5;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferConstAssertion",
					Line:      1,
					Column:    16,
				},
			},
			Output: []string{"let foo = 5 as const;"},
		},
		{
			Code: `
class foo {
	bar: 'baz' = 'baz';
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "variableConstAssertion",
					Line:      3,
					Column:    7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "variableSuggest",
							Output: `
class foo {
	bar = 'baz' as const;
}
			`,
						},
					},
				},
			},
		},
		{
			Code: `
class foo {
	bar: 2 = 2;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "variableConstAssertion",
					Line:      3,
					Column:    7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "variableSuggest",
							Output: `
class foo {
	bar = 2 as const;
}
			`,
						},
					},
				},
			},
		},
		{
			Code: `
class foo {
	foo = <'bar'>'bar';
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferConstAssertion",
					Line:      3,
					Column:    9,
				},
			},
			Output: []string{`
class foo {
	foo = <const>'bar';
}
			`},
		},
		{
			Code: `
class foo {
	foo = 'bar' as 'bar';
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferConstAssertion",
					Line:      3,
					Column:    17,
				},
			},
			Output: []string{`
class foo {
	foo = 'bar' as const;
}
			`},
		},
		{
			Code: `
class foo {
	foo = 5 as 5;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferConstAssertion",
					Line:      3,
					Column:    13,
				},
			},
			Output: []string{`
class foo {
	foo = 5 as const;
}
			`},
		},
		{
			Code: `let truth: true = true;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "variableConstAssertion",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "variableSuggest",
							Output:    `let truth = true as const;`,
						},
					},
				},
			},
		},
		{
			Code: `let falsity = false as false;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferConstAssertion"},
			},
			Output: []string{`let falsity = false as const;`},
		},
		{
			Code: `let value /* before colon */ : /* after: colon */ '值' = /* before value */ '值';`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "variableConstAssertion",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "variableSuggest",
							Output:    `let value /* before colon */  = /* before value */ '值' as const;`,
						},
					},
				},
			},
		},
		{
			Code: `class C { readonly ['x:y'] /* before colon */ : /* after: colon */ '值' = /* before value */ '值'; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "variableConstAssertion",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "variableSuggest",
							Output:    `class C { readonly ['x:y'] /* before colon */  = /* before value */ '值' as const; }`,
						},
					},
				},
			},
		},
		{
			Code: `const value = /* before value */ '值' as /* before type */ '值';`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferConstAssertion"},
			},
			Output: []string{`const value = /* before value */ '值' as /* before type */ const;`},
		},
		{
			Code: `const value = < /* before type */ '值'>/* before value */ '值';`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferConstAssertion"},
			},
			Output: []string{`const value = < /* before type */ const>/* before value */ '值';`},
		},
	})
}

func TestPreferAsConstDefersEdits(t *testing.T) {
	const source = `let annotation /* before colon */ : /* after: colon */ '值' = /* before value */ '值';
const assertion = /* before value */ 'x' as /* before type */ 'x';`

	run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
			FileName: "/prefer-as-const.ts",
			Path:     "/prefer-as-const.ts",
		}, source, core.ScriptKindTS)
		comments := rule.NewCommentStore(sourceFile)
		var diagnostics []rule.RuleDiagnostic
		ctx := rule.RuleContext{
			SourceFile:     sourceFile,
			Comments:       comments,
			DisableManager: rule.NewDisableManager(sourceFile, comments),
		}.WithDiagnosticConsumer(
			PreferAsConstRule.Name,
			rule.SeverityError,
			rule.DiagnosticConsumer{
				Demand: demand,
				Report: func(diagnostic rule.RuleDiagnostic) {
					diagnostics = append(diagnostics, diagnostic)
				},
			},
		)

		listeners := PreferAsConstRule.Run(ctx, nil)
		var visit func(*ast.Node) bool
		visit = func(node *ast.Node) bool {
			if listener := listeners[node.Kind]; listener != nil {
				listener(node)
			}
			node.ForEachChild(visit)
			return false
		}
		visit(sourceFile.AsNode())
		return diagnostics
	}

	diagnosticsOnly := run(rule.EditDemandNone)
	if len(diagnosticsOnly) != 2 {
		t.Fatalf("diagnostics-only run produced %d diagnostics, want 2", len(diagnosticsOnly))
	}
	for _, diagnostic := range diagnosticsOnly {
		if diagnostic.FixesPtr != nil || diagnostic.Suggestions != nil {
			t.Fatalf("diagnostics-only %s report materialized edits", diagnostic.Message.Id)
		}
	}

	withEdits := run(rule.EditDemandAll)
	if len(withEdits) != 2 {
		t.Fatalf("edit-demand run produced %d diagnostics, want 2", len(withEdits))
	}
	for _, diagnostic := range withEdits {
		switch diagnostic.Message.Id {
		case "preferConstAssertion":
			fixes := diagnostic.Fixes()
			if len(fixes) != 1 {
				t.Fatalf("assertion report produced %d fixes, want 1", len(fixes))
			}
			fix := fixes[0]
			if got := source[fix.Range.Pos():fix.Range.End()]; got != "'x'" {
				t.Fatalf("assertion fix replaces %q, want 'x'", got)
			}
			if fix.Text != constAssertionText {
				t.Fatalf("assertion replacement = %q, want %q", fix.Text, constAssertionText)
			}
		case "variableConstAssertion":
			if diagnostic.Suggestions == nil || len(*diagnostic.Suggestions) != 1 {
				t.Fatal("annotation report did not produce exactly one suggestion")
			}
			fixes := (*diagnostic.Suggestions)[0].Fixes()
			if len(fixes) != 2 {
				t.Fatalf("annotation suggestion produced %d fixes, want 2", len(fixes))
			}
			if got := source[fixes[0].Range.Pos():fixes[0].Range.End()]; got != ": /* after: colon */ '值'" {
				t.Fatalf("annotation removal covers %q", got)
			}
			if fixes[1].Range.Pos() != fixes[1].Range.End() || fixes[1].Text != asConstSuffixText {
				t.Fatalf("annotation insertion = %#v", fixes[1])
			}
		default:
			t.Fatalf("unexpected diagnostic %q", diagnostic.Message.Id)
		}
	}
}
