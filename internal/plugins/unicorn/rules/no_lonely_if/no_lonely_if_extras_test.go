package no_lonely_if_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_lonely_if"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoLonelyIfLanguageForms(t *testing.T) {
	for _, test := range []struct{ condition, filename string }{
		{"value as boolean", "case.ts"},
		{"value satisfies boolean", "case.ts"},
		{"value!", "case.ts"},
		{"fn<string>", "case.ts"},
		{"<boolean>value", "case.ts"},
		{"<Component enabled={value} />", "case.tsx"},
		{"<>value</>", "case.tsx"},
	} {
		t.Run(test.condition, func(t *testing.T) {
			rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_lonely_if.NoLonelyIfRule, nil,
				[]rule_tester.InvalidTestCase{{
					Code: "if (a) { if (" + test.condition + ") foo(); }", FileName: test.filename,
					Output: []string{"if (a && " + test.condition + ") foo();"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-lonely-if", Message: lonelyIfMessage,
						Line: 1, Column: 10, EndLine: 1, EndColumn: 22 + len(test.condition)}},
				}})
		})
	}
}

func TestNoLonelyIfFixSafety(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_lonely_if.NoLonelyIfRule, nil,
		[]rule_tester.InvalidTestCase{
			// Upstream v76.0.0 misses these parentheses and changes the condition.
			{Code: "if (x => x) { if (b) foo(); }", Output: []string{"if ((x => x) && b) foo();"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-lonely-if", Line: 1, Column: 15, EndLine: 1, EndColumn: 28}}},
			{Code: "if (a) { if (x => x) foo(); }", Output: []string{"if (a && (x => x)) foo();"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-lonely-if", Line: 1, Column: 10, EndLine: 1, EndColumn: 28}}},
			// With no gap after the outer block, upstream omits the semicolon.
			{Code: "if (a) { if (b) foo() }[].map(bar)", Output: []string{"if (a && b) foo();[].map(bar)"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-lonely-if", Line: 1, Column: 10, EndLine: 1, EndColumn: 22}}},
		})
	for _, test := range []struct {
		code, output      string
		column, endColumn int
	}{
		{"if (a) { if (b) foo() }bar()", "if (a && b) foo();bar()", 10, 22},
		{"if (a) { if (b) foo() } bar()", "if (a && b) foo();bar()", 10, 22},
		{"if (a) { if (b) foo() } /* between */ bar()", "if (a && b) foo(); /* between */ bar()", 10, 22},
		{"if (a) { if (b) foo() }\r\nbar()", "if (a && b) foo()\r\nbar()", 10, 22},
		{"if (a) { if (b) foo() }\u2028bar()", "if (a && b) foo()\u2028bar()", 10, 22},
		{"if (a) { if (b) foo() } /*\u2029*/ bar()", "if (a && b) foo() /*\u2029*/ bar()", 10, 22},
		{"if (a) { if (b) foo() }\u2028[0].map(bar)", "if (a && b) foo();[0].map(bar)", 10, 22},
		{"if (a) { if (b) while (c) {} }bar()", "if (a && b) while (c) {}bar()", 10, 29},
		{"if (a) { if (b) debugger }foo()", "if (a && b) debugger;foo()", 10, 25},
		{"while (ready) { if (a) { if (b) break }foo() }", "while (ready) { if (a && b) break;foo() }", 26, 38},
		{"while (ready) { if (a) { if (b) continue }foo() }", "while (ready) { if (a && b) continue;foo() }", 26, 41},
		{"function f() { if (a) { if (b) return }foo() }", "function f() { if (a && b) return;foo() }", 25, 38},
		{"if (a) { if (b) foo() };bar()", "if (a && b) foo();bar()", 10, 22},
	} {
		t.Run(test.code, func(t *testing.T) {
			rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_lonely_if.NoLonelyIfRule, nil,
				[]rule_tester.InvalidTestCase{{Code: test.code, Output: []string{test.output},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-lonely-if", Message: lonelyIfMessage,
						Line: 1, Column: test.column, EndLine: 1, EndColumn: test.endColumn}},
				}})
		})
	}
}

// RuleTester rejects syntax errors; exercise recovery directly to ensure that
// missing braces never produce an invalid range or partial rewrite.
func TestNoLonelyIfUnterminatedBlocks(t *testing.T) {
	for _, code := range []string{"if (a) { if (b) {}", "if (a) { if (b) {"} {
		source := parser.ParseSourceFile(ast.SourceFileParseOptions{
			FileName: "/unterminated.ts", Path: "/unterminated.ts",
		}, code, core.ScriptKindTS)
		var diagnostics []rule.RuleDiagnostic
		ctx := rule.RuleContext{SourceFile: source}.WithDiagnosticConsumer(
			"unicorn/no-lonely-if", rule.SeverityError, rule.DiagnosticConsumer{
				Demand: rule.EditDemandAll,
				Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) },
			})
		listener := no_lonely_if.NoLonelyIfRule.Run(ctx, nil)[ast.KindIfStatement]
		outer := source.Statements.Nodes[0]
		inner := outer.AsIfStatement().ThenStatement.AsBlock().Statements.Nodes[0]
		listener(inner)
		if len(diagnostics) != 1 {
			t.Fatalf("%q: got %d diagnostics", code, len(diagnostics))
		}
		diagnostic := diagnostics[0]
		if diagnostic.Message.Id != "no-lonely-if" || diagnostic.Range.Pos() != 9 ||
			diagnostic.Range.End() != len(code) || len(diagnostic.Fixes()) != 0 {
			t.Errorf("%q: unexpected diagnostic: %+v", code, diagnostic)
		}
	}
}

func TestNoLonelyIfEditDemand(t *testing.T) {
	const code = "if (a) { if (b) foo(); }"
	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, source, err := helper.CreateTestProgram(code, "edit-demand.js", "tsconfig.json")
	if err != nil {
		t.Fatal(err)
	}
	var autofix *rule.RuleDiagnostic
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: lintprogram.NewFromCompiler(program), File: source.FileName(),
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{Name: "unicorn/no-lonely-if", Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners { return no_lonely_if.NoLonelyIfRule.Run(ctx, nil) },
				}}
			},
			Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }},
		})
		if len(diagnostics) != 1 {
			t.Fatalf("demand %d: got %d diagnostics", demand, len(diagnostics))
		}
		diagnostic := diagnostics[0]
		if diagnostic.Message.Id != "no-lonely-if" || diagnostic.Message.Description != lonelyIfMessage ||
			diagnostic.Range.Pos() != 9 || diagnostic.Range.End() != 22 ||
			diagnostic.RuleName != "unicorn/no-lonely-if" || diagnostic.Severity != rule.SeverityError {
			t.Fatalf("demand %d: unexpected diagnostic: %+v", demand, diagnostic)
		}
		wantFix := demand&rule.EditDemandAutofix != 0
		if (diagnostic.FixesPtr != nil) != wantFix || diagnostic.Suggestions != nil {
			t.Fatalf("demand %d: unexpected edit artifacts", demand)
		}
		if wantFix {
			if autofix != nil && !reflect.DeepEqual(autofix.FixesPtr, diagnostic.FixesPtr) {
				t.Fatalf("demand %d changed the fix", demand)
			}
			autofix = &diagnostic
			fixes := *diagnostic.FixesPtr
			if len(fixes) != 1 || fixes[0].Range.Pos() != 0 || fixes[0].Range.End() != len(code) {
				t.Fatalf("demand %d: expected one replacement covering the outer if", demand)
			}
		}
		output, _, fixed := linter.ApplyRuleFixes(code, diagnostics)
		wantOutput := code
		if wantFix {
			wantOutput = "if (a && b) foo();"
		}
		if output != wantOutput || fixed != wantFix {
			t.Fatalf("demand %d: got %q, fixed %t", demand, output, fixed)
		}
	}
}

func TestNoLonelyIfDirectives(t *testing.T) {
	for _, code := range []string{
		"/* eslint-disable unicorn/no-lonely-if */\nif (a) { if (b) foo(); }",
		"if (a) { if (b) foo(); } // eslint-disable-line unicorn/no-lonely-if",
		"if (a) {\n// eslint-disable-next-line unicorn/no-lonely-if\nif (b) foo();\n}",
	} {
		helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
		program, source, err := helper.CreateTestProgram(code, "directives.js", "tsconfig.json")
		if err != nil {
			t.Fatal(err)
		}
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: lintprogram.NewFromCompiler(program), File: source.FileName(),
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{Name: "unicorn/no-lonely-if", Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners { return no_lonely_if.NoLonelyIfRule.Run(ctx, nil) },
				}}
			},
			Consumer: rule.DiagnosticConsumer{Demand: rule.EditDemandAll, Report: func(d rule.RuleDiagnostic) {
				t.Errorf("disabled rule reported %+v for %q", d, code)
			}},
		})
	}
}

func TestNoLonelyIfExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_lonely_if.NoLonelyIfRule,
		[]rule_tester.ValidTestCase{
			{Code: "{ if (b) foo(); }"},
			{Code: "while (a) { if (b) foo(); }"},
			{Code: "if (a) { if (b) foo(); else bar(); }"},
			{Code: "if (a) if (b) foo(); else bar();"},
			{Code: "if (a) { ; if (b) foo(); }"},
			{Code: "if (a) { if (b) foo(); ; }"},
			{Code: "if (a) { { if (b) foo(); } }"},
			{Code: "if (a) { label: if (b) foo(); }"},
		},
		[]rule_tester.InvalidTestCase{
			// comments without blocks
			{
				Code:   "if (a) /* before */ if /* prefix */ ( b ) /* body */ foo();",
				Output: []string{"/* before */ if (a &&  /* prefix */ b ) /* body */ foo();"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 1, Column: 21, EndLine: 1, EndColumn: 60},
				},
			},
			// trailing comment into block
			{
				Code:   "if (a) { if (b) { foo(); } /* after */ }",
				Output: []string{"if (a && b) { foo();  /* after */ }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 1, Column: 10, EndLine: 1, EndColumn: 27},
				},
			},
			// unbraced line comment
			{
				Code:   "if (a) // before\nif (b) foo();",
				Output: []string{"// before\nif (a && b) foo();"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 2, Column: 1, EndLine: 2, EndColumn: 14},
				},
			},
			// condition line comment and block
			{
				Code:   "if (a) // condition\n{ /* leading */ if (b) /* body */ { foo(); } /* trailing */ }",
				Output: []string{"/* leading */ if (a && b) // condition\n /* body */ { foo();  /* trailing */ }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 2, Column: 17, EndLine: 2, EndColumn: 45},
				},
			},
			// JavaScript whitespace and UTF-16
			{
				Code:   "/* 😀 */ if (a) {\ufeff/* 😀 */ if\ufeff/* 😀 */ (  b ) foo(); }",
				Output: []string{"/* 😀 */ /* 😀 */ if (a && \ufeff/* 😀 */ b ) foo();"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 1, Column: 28, EndLine: 1, EndColumn: 53},
				},
			},
			// CRLF and preserved ASI gap
			{
				Code:   "if (a) {\r\n if (b) foo()\r\n}\r\n// between\r\n[0].map(bar)",
				Output: []string{"if (a && b) foo();\r\n// between\r\n[0].map(bar)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 2, Column: 2, EndLine: 2, EndColumn: 14},
				},
			},
			// bare carriage return ASI gap
			{
				Code:   "if (a) { if (b) foo() }\r[0].map(bar)",
				Output: []string{"if (a && b) foo();[0].map(bar)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 1, Column: 10, EndLine: 1, EndColumn: 22},
				},
			},
			// no semicolon at EOF
			{
				Code:   "if (a) { if (b) foo() }",
				Output: []string{"if (a && b) foo()"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 1, Column: 10, EndLine: 1, EndColumn: 22},
				},
			},
			// safe following statement
			{
				Code:   "if (a) { if (b) foo() }\nbar()",
				Output: []string{"if (a && b) foo()\nbar()"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 1, Column: 10, EndLine: 1, EndColumn: 22},
				},
			},
			// postfix expression ASI
			{
				Code:   "if (a) { if (b) value++ }\n[0].map(bar)",
				Output: []string{"if (a && b) value++;\n[0].map(bar)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 1, Column: 10, EndLine: 1, EndColumn: 24},
				},
			},
			// nested control flow ASI
			{
				Code:   "if (a) { if (b) for (;;) foo() }\n[0].map(bar)",
				Output: []string{"if (a && b) for (;;) foo();\n[0].map(bar)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 1, Column: 10, EndLine: 1, EndColumn: 31},
				},
			},
			// nested block control flow
			{
				Code:   "if (a) { if (b) while (c) {} }\n[0].map(bar)",
				Output: []string{"if (a && b) while (c) {}\n[0].map(bar)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 1, Column: 10, EndLine: 1, EndColumn: 29},
				},
			},
			// return ASI
			{
				Code:   "function f() { if (a) { if (b) return value }\n[0].map(bar) }",
				Output: []string{"function f() { if (a && b) return value;\n[0].map(bar) }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 1, Column: 25, EndLine: 1, EndColumn: 44},
				},
			},
			// bare return ASI
			{
				Code:   "function f() { if (a) { if (b) return }\n[0].map(bar) }",
				Output: []string{"function f() { if (a && b) return;\n[0].map(bar) }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 1, Column: 25, EndLine: 1, EndColumn: 38},
				},
			},
			// throw ASI
			{
				Code:   "function f() { if (a) { if (b) throw value }\n[0].map(bar) }",
				Output: []string{"function f() { if (a && b) throw value;\n[0].map(bar) }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 1, Column: 25, EndLine: 1, EndColumn: 43},
				},
			},
			// variable ASI
			{
				Code:   "if (a) { if (b) var value = {} }\n[0].map(bar)",
				Output: []string{"if (a && b) var value = {};\n[0].map(bar)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 1, Column: 10, EndLine: 1, EndColumn: 31},
				},
			},
			// labeled break ASI
			{
				Code:   "outer: while (ready) { if (a) { if (b) break outer }\n[0].map(bar) }",
				Output: []string{"outer: while (ready) { if (a && b) break outer;\n[0].map(bar) }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 1, Column: 33, EndLine: 1, EndColumn: 51},
				},
			},
			// bare break
			{
				Code:   "while (ready) { if (a) { if (b) break }\n[0].map(bar) }",
				Output: []string{"while (ready) { if (a && b) break\n[0].map(bar) }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 1, Column: 26, EndLine: 1, EndColumn: 38},
				},
			},
			// do while ASI
			{
				Code:   "if (a) { if (b) do {} while (c) }\n[0].map(bar)",
				Output: []string{"if (a && b) do {} while (c);\n[0].map(bar)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
				},
			},
			// inner consequent with else
			{
				Code:   "if (a) { if (b) if (c) foo(); else bar(); }",
				Output: []string{"if (a && b) if (c) foo(); else bar();"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 1, Column: 10, EndLine: 1, EndColumn: 42},
				},
			},
			// optional and computed access
			{
				Code:   "if (obj?.[key]) { if (fn?.()) foo(); }",
				Output: []string{"if (obj?.[key] && fn?.()) foo();"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 1, Column: 19, EndLine: 1, EndColumn: 37},
				},
			},
			// private fields
			{
				Code:   "class C { #ready; f() { if (this.#ready) { if (this[\"enabled\"]) foo(); } } }",
				Output: []string{"class C { #ready; f() { if (this.#ready && this[\"enabled\"]) foo(); } }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 1, Column: 44, EndLine: 1, EndColumn: 71},
				},
			},
			// JSDoc parentheses
			{
				Code:   "if (/** @type {boolean} */ (a || b)) { if ((c ?? d)) foo(); }",
				Output: []string{"if (/** @type {boolean} */ (a || b) && (c ?? d)) foo();"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 1, Column: 40, EndLine: 1, EndColumn: 60},
				},
			},
		},
	)
}
