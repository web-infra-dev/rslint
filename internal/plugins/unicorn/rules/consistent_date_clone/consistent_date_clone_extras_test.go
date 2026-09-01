// TestConsistentDateCloneExtras locks in tsgo-specific AST shapes, every
// reachable upstream gate, and edit-demand behavior. The direct v74.0.0
// migration lives in the sibling consistent_date_clone_upstream_test.go file.
package consistent_date_clone_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	consistent_date_clone "github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/consistent_date_clone"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestConsistentDateCloneExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&consistent_date_clone.ConsistentDateCloneRule,
		[]rule_tester.ValidTestCase{
			// Locks in upstream isNewExpression() arm 1: the callee must be the
			// exact identifier Date rather than a same-named member.
			{Code: `new globalThis.Date(date.getTime())`, FileName: "file.js"},

			// Locks in upstream isNewExpression() arms 2 and 3: exactly one
			// non-spread argument is required.
			{Code: `new Date()`, FileName: "file.js"},
			{Code: `new Date(date.getTime(), other)`, FileName: "file.js"},
			{Code: `new Date(...[date.getTime()])`, FileName: "file.js"},

			// Locks in upstream isMethodCall() arms 1-4: the inner call must be a
			// non-optional, non-computed, zero-argument getTime method call.
			{Code: `new Date(date.getTime?.())`, FileName: "file.js"},
			{Code: `new Date((date?.getTime)())`, FileName: "file.js"},
			{Code: `new Date(date["getTime"]())`, FileName: "file.js"},
			{Code: `new Date(date.getTime(...[]))`, FileName: "file.js"},
			{Code: `new Date(date.getTime)`, FileName: "file.js"},

			// Dimension 4: wrappers around the argument or Date callee remain visible
			// to TSESTree and therefore prevent the direct shape from matching.
			{Code: `new Date((date.getTime() as number))`, FileName: "file.ts", Tsx: false},
			{Code: `new (Date as DateConstructor)(date.getTime())`, FileName: "file.ts", Tsx: false},

			// Dimension 4: a private method name is not an Identifier property.
			{
				Code:     `class Clock { #getTime() { return 0 } clone(date) { return new Date(date.#getTime()) } }`,
				FileName: "file.js",
			},

			// ---- Real-user: upstream #2437 component construction without milliseconds ----
			{Code: `new Date(date.getFullYear(), date.getMonth(), date.getDate(), date.getHours(), date.getMinutes(), date.getSeconds())`, FileName: "file.js"},
			// ---- Real-user: upstream #2437 component construction with milliseconds ----
			{Code: `new Date(date.getFullYear(), date.getMonth(), date.getDate(), date.getHours(), date.getMinutes(), date.getSeconds(), date.getMilliseconds())`, FileName: "file.js"},

			// N/A: JSX tag names and object-member key equivalence classes are not
			// expression shapes consumed by this rule.
			// N/A: the rule has an empty options schema, so defaults and option
			// combinations do not apply.
		},
		[]rule_tester.InvalidTestCase{
			// Dimension 4: parentheses around the Date callee and argument call are
			// transparent for detection but preserved by the fixer.
			{
				Code:     `new (Date)(date.getTime())`,
				FileName: "file.js",
				Output:   []string{`new (Date)(date)`},
				Errors:   []rule_tester.InvalidTestCaseError{expectedError(`new (Date)(date.getTime())`, `getTime()`, 0)},
			},
			{
				Code:     `new Date((date.getTime()))`,
				FileName: "file.js",
				Output:   []string{`new Date((date))`},
				Errors:   []rule_tester.InvalidTestCaseError{expectedError(`new Date((date.getTime()))`, `getTime()`, 0)},
			},
			{
				Code:     `new Date((date.getTime)())`,
				FileName: "file.js",
				Output:   []string{`new Date((date))`},
				Errors:   []rule_tester.InvalidTestCaseError{expectedError(`new Date((date.getTime)())`, `getTime)()`, 0)},
			},
			{
				Code:     `new Date(((date)).getTime /* call */ ())`,
				FileName: "file.js",
				Output:   []string{`new Date(((date)))`},
				Errors:   []rule_tester.InvalidTestCaseError{expectedError(`new Date(((date)).getTime /* call */ ())`, `getTime /* call */ ()`, 0)},
			},

			// The property diagnostic starts after receiver-side comments, while the
			// fix removes those comments together with the redundant method call.
			{
				Code:     `new Date(date /* receiver */ .getTime())`,
				FileName: "file.js",
				Output:   []string{`new Date(date)`},
				Errors:   []rule_tester.InvalidTestCaseError{expectedError(`new Date(date /* receiver */ .getTime())`, `getTime()`, 0)},
			},
			{
				Code: `new Date(
	date
		.getTime(
			/* comment */
		),
)`,
				FileName: "file.js",
				Output: []string{`new Date(
	date,
)`},
				Errors: []rule_tester.InvalidTestCaseError{expectedError(`new Date(
	date
		.getTime(
			/* comment */
		),
)`, `getTime(
			/* comment */
		)`, 0)},
			},

			// Dimension 4: TypeScript wrappers on the receiver are part of the value
			// being cloned and must survive the rewrite.
			{
				Code:     `new Date(date!.getTime())`,
				FileName: "file.ts",
				Tsx:      false,
				Output:   []string{`new Date(date!)`},
				Errors:   []rule_tester.InvalidTestCaseError{expectedError(`new Date(date!.getTime())`, `getTime()`, 0)},
			},
			{
				Code:     `new Date((date satisfies Date).getTime())`,
				FileName: "file.ts",
				Tsx:      false,
				Output:   []string{`new Date((date satisfies Date))`},
				Errors:   []rule_tester.InvalidTestCaseError{expectedError(`new Date((date satisfies Date).getTime())`, `getTime()`, 0)},
			},

			// Type arguments do not change either call-expression shape.
			{
				Code:     `new Date<string>(date.getTime<number>())`,
				FileName: "file.ts",
				Tsx:      false,
				Output:   []string{`new Date<string>(date)`},
				Errors:   []rule_tester.InvalidTestCaseError{expectedError(`new Date<string>(date.getTime<number>())`, `getTime<number>()`, 0)},
			},

			// The rule is syntax-based: a local declaration named Date still matches.
			{
				Code:     `class Date { constructor(value) {} } new Date(date.getTime())`,
				FileName: "file.js",
				Output:   []string{`class Date { constructor(value) {} } new Date(date)`},
				Errors:   []rule_tester.InvalidTestCaseError{expectedError(`class Date { constructor(value) {} } new Date(date.getTime())`, `getTime()`, 0)},
			},

			// Same-kind siblings report in source order and both fixes apply in one pass.
			{
				Code:     `const first = new Date(a.getTime()); const second = new Date(b.getTime());`,
				FileName: "file.js",
				Output:   []string{`const first = new Date(a); const second = new Date(b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					expectedError(`const first = new Date(a.getTime()); const second = new Date(b.getTime());`, `getTime()`, 0),
					expectedError(`const first = new Date(a.getTime()); const second = new Date(b.getTime());`, `getTime()`, 1),
				},
			},
		},
	)
}

func TestConsistentDateCloneEditDemand(t *testing.T) {
	t.Parallel()

	const fileName = "edit-demand.ts"
	const source = `const clone = new Date(date.getTime());`
	const fixedSource = `const clone = new Date(date);`

	compilerProgram, sourceFile := createConsistentDateCloneProgram(t, fileName, source)
	diagnostics := make(map[rule.EditDemand]rule.RuleDiagnostic, 4)
	for _, demand := range []rule.EditDemand{
		rule.EditDemandNone,
		rule.EditDemandAutofix,
		rule.EditDemandSuggestion,
		rule.EditDemandAll,
	} {
		got := lintConsistentDateCloneWithDemand(compilerProgram, sourceFile, demand)
		if len(got) != 1 {
			t.Fatalf("demand %d: diagnostics = %d, want 1", demand, len(got))
		}
		diagnostics[demand] = got[0]
	}

	base := diagnostics[rule.EditDemandNone]
	for demand, diagnostic := range diagnostics {
		want := base
		want.FixesPtr = nil
		want.Suggestions = nil
		got := diagnostic
		got.FixesPtr = nil
		got.Suggestions = nil
		if !reflect.DeepEqual(got, want) {
			t.Errorf("demand %d changed diagnostic metadata:\ngot:  %#v\nwant: %#v", demand, got, want)
		}
		if diagnostic.Suggestions != nil {
			t.Errorf("demand %d unexpectedly materialized suggestions", demand)
		}
	}

	if diagnostics[rule.EditDemandNone].FixesPtr != nil {
		t.Errorf("none demand unexpectedly materialized fixes")
	}
	if diagnostics[rule.EditDemandSuggestion].FixesPtr != nil {
		t.Errorf("suggestion-only demand unexpectedly materialized fixes")
	}
	autofixOnly := diagnostics[rule.EditDemandAutofix].FixesPtr
	allFixes := diagnostics[rule.EditDemandAll].FixesPtr
	if autofixOnly == nil || allFixes == nil || !reflect.DeepEqual(*autofixOnly, *allFixes) {
		t.Fatalf("autofix artifacts differ between autofix-only and all demand")
	}
	output, unapplied, fixed := linter.ApplyRuleFixes(source, []rule.RuleDiagnostic{diagnostics[rule.EditDemandAll]})
	if !fixed || len(unapplied) != 0 || output != fixedSource {
		t.Fatalf("autofix result = %q, unapplied = %d, fixed = %v; want %q", output, len(unapplied), fixed, fixedSource)
	}
}

func createConsistentDateCloneProgram(t testing.TB, fileName, source string) (*compiler.Program, *ast.SourceFile) {
	t.Helper()
	rootDir := fixtures.GetRootDir()
	fs := utils.NewOverlayVFS(rootDir.FS, map[string]string{tspath.ResolvePath(rootDir.Dir, fileName): source})
	host := utils.CreateCompilerHost(rootDir.Dir, fs)
	compilerProgram, err := utils.CreateProgram(true, fs, rootDir.Dir, "tsconfig.json", host)
	if err != nil {
		t.Fatalf("failed to create program: %v", err)
	}
	sourceFile := compilerProgram.GetSourceFile(fileName)
	if sourceFile == nil {
		t.Fatalf("source file %q not found", fileName)
	}
	return compilerProgram, sourceFile
}

func lintConsistentDateCloneWithDemand(
	compilerProgram *compiler.Program,
	sourceFile *ast.SourceFile,
	demand rule.EditDemand,
) []rule.RuleDiagnostic {
	var diagnostics []rule.RuleDiagnostic
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program:     program.NewFromCompiler(compilerProgram),
		File:        sourceFile.FileName(),
		HasTypeInfo: true,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:     consistent_date_clone.ConsistentDateCloneRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return consistent_date_clone.ConsistentDateCloneRule.Run(ctx, nil)
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
	return diagnostics
}
