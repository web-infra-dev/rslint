package no_var

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/binder"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoVarComputedAssignmentKey locks in rule behavior inside an evaluated
// computed property name of a destructuring assignment target. The shared
// traversal contract is covered by internal/linter; this test protects the
// rule's listener-to-diagnostic integration.
func TestNoVarComputedAssignmentKey(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoVarRule,
		nil,
		[]rule_tester.InvalidTestCase{
			{
				// The self-reference deliberately makes var→let unsafe, so this
				// regression test asserts only diagnostics rather than edit behavior.
				Code: `({ [(() => { var value = value; })()]: target } = source);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar", Line: 1, Column: 14, EndLine: 1, EndColumn: 32},
				},
			},
		},
	)
}

func TestNoVarDeferredInitializerCalls(t *testing.T) {
	initializers := []struct {
		code string
		fix  bool
	}{
		// Stored bodies, including callbacks returned by an IIFE, do not run
		// while value is initialized. These fixes match ESLint 10.9.1.
		{`({ method() { return read(); } })`, true},
		{`({ method: function() { return read(); } })`, true},
		{`({ get value() { return read(); } })`, true},
		{`({ later: (arg = read()) => arg })`, true},
		{`(() => ({ method: function() { return read(); } }))()`, true},
		{`(() => { function later() { return read(); } return later; })()`, true},
		{`(() => { const later = () => read(); return later; })()`, true},
		{`class { method() { return read(); } }`, true},
		{`class { field = read(); }`, true},
		// Reduced from the ReadableStream constructor in streams-lib.js.
		{`function() { function Value() {} Value.prototype.tee = function() { return read(); }; return Value; }()`, true},

		// These calls actually read value before initialization completes.
		// Retain the existing safety protection even where upstream offers a fix.
		{`read()`, false},
		{`(() => read())()`, false},
		{`(function() { return read(); }).call(null)`, false},
		{`(function() { return read(); }).apply(null)`, false},
		{`(function() { return read(); }).bind(null)()`, false},
		{`(0, () => read())()`, false},
		{`(() => () => read())()()`, false},
		{`(function() { return function() { return read(); }; })()()`, false},
		{`(() => { function now() { return read(); } return now(); })()`, false},
		{`(() => { const now = () => read(); return now(); })()`, false},
		{`(() => { function first() { second(); } function second() { read(); first(); } first(); })()`, false},
		{`((callback) => callback())(() => read())`, false},
		{`[0].map(() => read())`, false},
		{`class { static field = read(); }`, false},
		{`class { static { read(); } }`, false},
		{`class { [read()]() {} }`, false},
		{`({ [read()]() {} })`, false},
		{`new class { field = read(); }`, false},
		{`new class { constructor() { read(); } }`, false},
		{`(() => { class C { field = read(); } return new C(); })()`, false},
		{`({ method() { return read(); } }).method()`, false},
		{`({ get value() { return read(); } }).value`, false},
		{`({...{get key() { return read(); }}})`, false},
		{`(() => { const object = { method() { return read(); } }; return object.method(); })()`, false},
	}
	var cases []rule_tester.InvalidTestCase
	for _, sourceType := range []string{"module", "script", "commonjs"} {
		for _, initializer := range initializers {
			declaration := "var value = " + initializer.code + ";"
			code := "function wrap() { " + declaration + " function read() { return value; } }"
			c := rule_tester.InvalidTestCase{
				Code:            code,
				FileName:        "deferred.js",
				TSConfig:        "tsconfig.allow-js.json",
				LanguageOptions: rule.LanguageOptions{SourceType: sourceType},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedVar",
					Message:   "Unexpected var, use let or const instead.",
					Line:      1, Column: 19, EndLine: 1, EndColumn: 19 + len(declaration),
				}},
			}
			if initializer.fix {
				c.Output = []string{strings.Replace(code, "var value", "let value", 1)}
			}
			cases = append(cases, c)
		}
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoVarRule, nil, cases)
}

func TestNoVarDeferredInitializerTypeScript(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoVarRule, nil, []rule_tester.InvalidTestCase{
		{
			Code:   `var value = { method: (() => read()) as () => unknown }; function read() { return value; }`,
			Output: []string{`let value = { method: (() => read()) as () => unknown }; function read() { return value; }`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedVar"}},
		},
		{
			Code:   `var value = class { method() { return (read satisfies Function)(); } }; function read() { return value; }`,
			Output: []string{`let value = class { method() { return (read satisfies Function)(); } }; function read() { return value; }`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedVar"}},
		},
		{
			Code:   `var value = ((() => (read as Function)()) as Function)(); function read() { return value; }`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedVar"}},
		},
		{
			Code:   `var value = ({ method: (() => read()) as Function }).method(); function read() { return value; }`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedVar"}},
		},
	})
}

func TestNoVarDeferredInitializerEditDemand(t *testing.T) {
	const body = "var value = { method() { return read(); } }; function read() { return value; }"
	for _, prefix := range []string{"/* comment */\n", "// eslint-disable-next-line no-var\n", "/* rslint-disable no-var */\n"} {
		for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
			sf := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/demand.js", Path: tspath.Path("/demand.js")}, prefix+body, core.ScriptKindJS)
			binder.BindSourceFile(sf)
			_, init, language := rule.ResolveLanguageDefaults(sf.FileName(), rule.LanguageOptions{})
			var diagnostics []rule.RuleDiagnostic
			ctx := rule.RuleContext{
				SourceFile: sf, LanguageOptions: language,
				Refs:           rule.NewRefStore(sf, &core.CompilerOptions{}, nil, init),
				DisableManager: rule.NewDisableManager(sf, rule.NewCommentStore(sf)),
			}.WithDiagnosticConsumer("no-var", rule.SeverityError, rule.DiagnosticConsumer{
				Demand: demand,
				Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) },
			})
			declaration := sf.Statements.Nodes[0].AsVariableStatement().DeclarationList
			NoVarRule.Run(ctx, nil)[ast.KindVariableDeclarationList](declaration)
			if strings.Contains(prefix, "disable") {
				if len(diagnostics) != 0 {
					t.Fatalf("suppressed diagnostic with demand %d: %v", demand, diagnostics)
				}
				continue
			}
			if len(diagnostics) != 1 {
				t.Fatalf("demand %d: got %d diagnostics", demand, len(diagnostics))
			}
			d := diagnostics[0]
			if d.Message.Id != "unexpectedVar" || d.Range.Pos() != len(prefix) || d.Range.End() != len(prefix)+44 || d.Suggestions != nil {
				t.Fatalf("demand %d: unexpected diagnostic: %v", demand, d)
			}
			if demand&rule.EditDemandAutofix != 0 {
				fixes := d.Fixes()
				if len(fixes) != 1 || fixes[0].Text != "let" || fixes[0].Range.Pos() != len(prefix) || fixes[0].Range.End() != len(prefix)+3 {
					t.Fatalf("demand %d: unexpected fixes: %v", demand, fixes)
				}
			} else if d.FixesPtr != nil {
				t.Fatalf("demand %d: unwanted fix artifact", demand)
			}
		}
	}
}
