package no_cond_assign

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
	"gotest.tools/v3/assert"
)

func TestNoCondAssignRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoCondAssignRule,
		// Valid cases - ported from ESLint
		[]rule_tester.ValidTestCase{
			// Default behavior (except-parens)
			{Code: `var x = 0; if (x == 0) { var b = 1; }`},
			{Code: `var x = 5; while (x < 5) { x = x + 1; }`},
			{Code: `x = 0;`},
			{Code: `var x; var b = (x === 0) ? 1 : 0;`},
			{Code: `var x; var b = ((x = 0)) ? 1 : 0;`},

			// With "except-parens" option - properly parenthesized assignments are allowed
			{Code: `if ((someNode = someNode.parentNode) !== null) { }`, Options: "except-parens"},
			{Code: `if ((a = b));`, Options: "except-parens"},
			{Code: `while ((a = b));`, Options: "except-parens"},
			{Code: `do {} while ((a = b));`, Options: "except-parens"},
			{Code: `for (;(a = b););`, Options: "except-parens"},
			{Code: `if (someNode || (someNode = parentNode)) { }`, Options: "except-parens"},
			{Code: `while (someNode || (someNode = parentNode)) { }`, Options: "except-parens"},
			{Code: `do { } while (someNode || (someNode = parentNode));`, Options: "except-parens"},
			{Code: `for (;someNode || (someNode = parentNode););`, Options: "except-parens"},

			// Assignments in ternary branches are not part of the ternary test.
			{Code: `var x, y; var result = condition ? x = 1 : y = 2;`},
			{Code: `var x, y; var result = condition ? x = 1 : y = 2;`, Options: "always"},
			{Code: `var result = condition ? value = { id: value.id } : typeof value === "function" && (value = callback);`},
			{Code: `var result = condition ? value = { id: value.id } : typeof value === "function" && (value = callback);`, Options: "always"},

			// Arrow functions
			{Code: `if ((node => node = parentNode)(someNode)) { }`, Options: "except-parens"},
			{Code: `if ((function(node) { return node = parentNode; })(someNode)) { }`, Options: "except-parens"},
			{Code: `if ((node => node = parentNode)(someNode)) { }`, Options: "always"},
			{Code: `if ((function(node) { return node = parentNode; })(someNode)) { }`, Options: "always"},
			{Code: `if (function(node) { return node = parentNode; }) { }`, Options: "always"},
			{Code: `if (class { method() { value = next(); } }) { }`, Options: "always"},

			// Switch statements - assignments in case clauses are not in test expressions
			{Code: `switch (foo) { case a = b: bar(); }`},
			{Code: `switch (foo) { case baz + (a = b): bar(); }`},

			// Assignments outside of conditionals
			{Code: `var x; x = 0;`},
			{Code: `var x = 1; x += 1;`},
			{Code: `for (x = 0; x < 10; x += 1) { }`, Options: "always"},

			// Disable directives suppress the assignment's exact report range.
			{Code: "/* eslint-disable test */\nif (x = y) { }"},
			{Code: "// eslint-disable-next-line test\nif (x = y) { }"},
			{Code: `if (x = y) { } // eslint-disable-line test`},

			// Comparisons (not assignments)
			{Code: `if (x === 0) { }`},
			{Code: `while (x == 1) { }`},
			{Code: `do { } while (x != 0);`},
			{Code: `for (; x >= 0;) { }`},
		},
		// Invalid cases - ported from ESLint
		[]rule_tester.InvalidTestCase{
			// Missing parentheses (default "except-parens" mode)
			{
				Code: `var x; if (x = 0) { var b = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missing",
						Message:   "Expected a conditional expression and instead saw an assignment.",
						Line:      1,
						Column:    12,
						EndLine:   1,
						EndColumn: 17,
					},
				},
			},
			{
				Code: `var x; while (x = 0) { var b = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 15},
				},
			},
			{
				Code: `var x = 0, y; do { y = x; } while (x = x + 1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 36},
				},
			},
			{
				Code: `var x; for(; x+=1 ;){};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 14},
				},
			},
			{
				Code: `var x; if ((x) = (0));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 12},
				},
			},

			// With "always" option - all assignments descended from conditional tests are forbidden
			{
				Code:    `if (x = 0) { }`,
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Message:   "Unexpected assignment within an 'if' statement.",
						Line:      1,
						Column:    5,
						EndLine:   1,
						EndColumn: 10,
					},
				},
			},
			{
				Code:    `while (x = 0) { }`,
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 8},
				},
			},
			{
				Code:    `do { } while (x = x + 1);`,
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 15},
				},
			},
			{
				Code:    `for(; x = y; ) { }`,
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 7},
				},
			},
			{
				Code:    `if ((x = 0)) { }`,
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 6},
				},
			},
			{
				Code:    `while ((x = 0)) { }`,
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 9},
				},
			},
			{
				Code:    `do { } while ((x = x + 1));`,
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 16},
				},
			},
			{
				Code:    `for(; (x = y); ) { }`,
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 8},
				},
			},
			{
				Code:    `if (someNode || (someNode = parentNode)) { }`,
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 18},
				},
			},
			{
				Code:    `while (someNode || (someNode = parentNode)) { }`,
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 21},
				},
			},
			{
				Code:    `do { } while (someNode || (someNode = parentNode));`,
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 28},
				},
			},
			{
				Code:    `for (; someNode || (someNode = parentNode); ) { }`,
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 21},
				},
			},

			// Ternary tests require two explicit parenthesis pairs in the default mode.
			{
				Code: `var x; var b = (x = 0) ? 1 : 0;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 17, EndLine: 1, EndColumn: 22},
				},
			},
			{
				Code:    `var x; var b = (x = 0) ? 1 : 0;`,
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Message:   "Unexpected assignment within ConditionalExpression.",
						Line:      1,
						Column:    17,
						EndLine:   1,
						EndColumn: 22,
					},
				},
			},
			{
				Code:    `if (condition ? (x = 1) : fallback) { }`,
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Message:   "Unexpected assignment within an 'if' statement.",
					},
				},
			},
			{
				Code:    `if ((x = y) ? ready : fallback) { }`,
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Message:   "Unexpected assignment within ConditionalExpression.",
					},
				},
			},
			{
				Code:    `for (; (typeof l === 'undefined' ? (l = 0) : l); i++) { }`,
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Message:   "Unexpected assignment within a 'for' statement.",
					},
				},
			},
			{
				Code: `(((3496.29)).property = 2e308) ? foo : bar;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing"},
				},
			},

			// Compound assignment operators
			{
				Code: `if (x += 1) { }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 5},
				},
			},
			{
				Code: `while (x -= 1) { }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 8},
				},
			},
			{
				Code: `do { } while (x *= 2);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 15},
				},
			},
			{
				Code: `if (x &&= next()) { }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 5},
				},
			},
			{
				Code: `while (x ||= next()) { }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 8},
				},
			},
			{
				Code: `for (; x ??= next(); ) { }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 8},
				},
			},
			{
				Code: `if (a = b = c) { }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 5},
				},
			},
			{
				Code:    `if (a = b = c) { }`,
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 5},
					{MessageId: "unexpected", Line: 1, Column: 9},
				},
			},
		},
	)
}

func TestNoCondAssignEditDemandParity(t *testing.T) {
	const code = `if (/* before */ value = next()) { }`
	wantStart := strings.Index(code, "value = next()")
	wantEnd := wantStart + len("value = next()")

	for _, demand := range []rule.EditDemand{
		rule.EditDemandNone,
		rule.EditDemandAutofix,
		rule.EditDemandSuggestion,
		rule.EditDemandAll,
	} {
		diagnostics := runNoCondAssign(t, code, nil, demand)
		assert.Equal(t, len(diagnostics), 1)
		diagnostic := diagnostics[0]
		assert.Equal(t, diagnostic.Message.Id, "missing")
		assert.Equal(t, diagnostic.Message.Description, "Expected a conditional expression and instead saw an assignment.")
		assert.Equal(t, diagnostic.Range.Pos(), wantStart)
		assert.Equal(t, diagnostic.Range.End(), wantEnd)
		assert.Assert(t, diagnostic.FixesPtr == nil)
		assert.Assert(t, diagnostic.Suggestions == nil)
	}
}

func TestNoCondAssignAlwaysMessageData(t *testing.T) {
	diagnostics := runNoCondAssign(t, `if (condition ? (value = next()) : fallback) { }`, []any{"always"}, rule.EditDemandNone)
	assert.Equal(t, len(diagnostics), 1)
	assert.Equal(t, diagnostics[0].Message.Description, "Unexpected assignment within an 'if' statement.")
	assert.Equal(t, diagnostics[0].Message.Data["type"], "an 'if' statement")
}

func TestNoCondAssignAlwaysFunctionBoundaries(t *testing.T) {
	tests := []struct {
		name string
		code string
		want int
	}{
		// ESTree represents each method as a MethodDefinition whose computed
		// name and decorators sit outside the nested FunctionExpression.
		{name: "class computed method", code: `if (class { [value = next()]() {} }) {}`, want: 1},
		{name: "object computed method", code: `if ({ [value = next()]() {} }) {}`, want: 1},
		{name: "computed getter", code: `if (class { get [value = next()]() { return 1 } }) {}`, want: 1},
		{name: "computed setter", code: `if (class { set [value = next()](nextValue) {} }) {}`, want: 1},
		{name: "static async generator name", code: `if (class { static async *[value = next()]() {} }) {}`, want: 1},
		{name: "method decorator", code: `if (class { @dec(value = next()) method() {} }) {}`, want: 1},
		{name: "field initializer", code: `if (class { field = (value = next()) }) {}`, want: 1},
		{name: "static block", code: `if (class { static { value = next() } }) {}`, want: 1},

		// These positions belong to the method's FunctionExpression and must
		// continue to shield their assignments from an outer condition.
		{name: "method body", code: `if (class { method() { value = next() } }) {}`},
		{name: "method parameter", code: `if (class { method(parameter = value = next()) {} }) {}`},
		{name: "parameter decorator", code: `if (class { method(@dec(value = next()) parameter: unknown) {} }) {}`},
		{name: "function in computed name", code: `if (class { [() => value = next()]() {} }) {}`},
		{name: "function in method decorator", code: `if (class { @dec(() => value = next()) method() {} }) {}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			diagnostics := runNoCondAssign(t, test.code, []any{"always"}, rule.EditDemandNone)
			assert.Equal(t, len(diagnostics), test.want)
			for _, diagnostic := range diagnostics {
				assert.Equal(t, diagnostic.Message.Id, "unexpected")
				assert.Equal(t, diagnostic.Message.Data["type"], "an 'if' statement")
			}
		})
	}
}

func runNoCondAssign(t *testing.T, code string, options []any, demand rule.EditDemand) []rule.RuleDiagnostic {
	t.Helper()

	root := fixtures.GetRootDir()
	const fileName = "file.ts"
	fs := utils.NewOverlayVFS(root.FS, map[string]string{
		tspath.ResolvePath(root.Dir, fileName): code,
	})
	host := utils.CreateCompilerHost(root.Dir, fs)
	compilerProgram, err := utils.CreateProgram(true, fs, root.Dir, "tsconfig.json", host)
	assert.NilError(t, err)
	sourceFile := compilerProgram.GetSourceFile(fileName)
	assert.Assert(t, sourceFile != nil)

	var diagnostics []rule.RuleDiagnostic
	programs := []*lintprogram.Program{lintprogram.NewFromCompiler(compilerProgram)}
	lintPlan, err := linter.PrepareLintPlan(linter.PrepareLintPlanOptions{
		Programs:         programs,
		TargetsByProgram: [][]string{{sourceFile.FileName()}},
		SingleThreaded:   true,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:     NoCondAssignRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return NoCondAssignRule.Run(ctx, options)
				},
			}}
		},
	})
	assert.NilError(t, err)
	_, err = linter.RunLinter(linter.RunLinterOptions{
		SingleThreaded: true,
		LintPlan:       lintPlan,
		Consumer: rule.DiagnosticConsumer{
			Demand: demand,
			Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			},
		},
	})
	assert.NilError(t, err)
	return diagnostics
}
