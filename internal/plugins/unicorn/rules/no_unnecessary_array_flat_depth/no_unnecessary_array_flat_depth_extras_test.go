// TestNoUnnecessaryArrayFlatDepthExtras locks in tsgo-specific AST shapes,
// real-user inputs, every reachable upstream gate, and autofix boundaries.
// The direct v74.0.0 migration lives in the sibling upstream test file.
package no_unnecessary_array_flat_depth_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	no_unnecessary_array_flat_depth "github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_unnecessary_array_flat_depth"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestNoUnnecessaryArrayFlatDepthExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_unnecessary_array_flat_depth.NoUnnecessaryArrayFlatDepthRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: access/key forms other than an identifier-named
			// dotted property do not match. ----
			// Locks in upstream create() arm 1: only a non-computed `.flat` call
			// with exactly one argument is eligible.
			{Code: `array["flat"](1)`, FileName: "file.js"},
			{Code: "array[`flat`](1)", FileName: "file.js"},
			{Code: `array[0](1)`, FileName: "file.js"},
			{Code: `array[Symbol.iterator](1)`, FileName: "file.js"},
			{Code: `array.flat?.(1)`, FileName: "file.js"},
			{Code: `(array?.flat)(1)`, FileName: "file.js"},

			// ---- Dimension 4: expression wrappers on the inspected argument do
			// not count as the literal itself. ----
			// Locks in upstream create() arm 2: the sole argument must itself be
			// the numeric literal value 1 rather than a wrapper or lookalike.
			{Code: `array.flat(...[1])`, FileName: "file.js"},
			{Code: `array.flat(1n)`, FileName: "file.js"},
			{Code: `array.flat("1")`, FileName: "file.js"},
			{Code: `array.flat(1 as number)`, FileName: "file.ts"},
			{Code: `array.flat(1 satisfies number)`, FileName: "file.ts"},
			{Code: `array.flat(1!)`, FileName: "file.ts"},

			// Locks in upstream create() arm 3: receivers known to be non-arrays
			// are skipped, including arrow and class expressions.
			{
				Code:     `class Collection { flat(depth: number) {} } function f(value: Collection) { value.flat(1) }`,
				FileName: "file.ts",
			},
			{Code: `(() => {}).flat(1)`, FileName: "file.js"},
			{Code: `(class {}).flat(1)`, FileName: "file.js"},

			// ---- Dimension 4: private identifier properties do not match the
			// identifier-named `.flat` method required by the rule. ----
			{
				Code:     `class Collection { #flat(depth: number) {} run() { this.#flat(1) } }`,
				FileName: "file.ts",
			},

			// ---- Dimension 4: graceful degradation for empty and spread argument
			// lists; neither shape is a numeric-literal argument. ----
			{Code: `array.flat()`, FileName: "file.js"},
			{Code: `array.flat(...[])`, FileName: "file.js"},

			// ---- Real-user: upstream #1629 uses a const binding for the default
			// depth. The rule deliberately remains literal-only. ----
			{Code: `const defaultDepth = 1; rows.flat(defaultDepth)`, FileName: "file.js"},
			// ---- Real-user: upstream #1629 uses a mutable depth that may change
			// before the call; it must not be treated as the literal value 1. ----
			{Code: `let levels = 1; levels = getDepth(); rows.flat(levels)`, FileName: "file.js"},

			// ---- Dimension 4 N/A: declaration/container variants, object-property
			// key classes, overload signatures, and body-absent declarations are not
			// inputs to this CallExpression rule. ----
			// ---- Dimension 4 N/A: the rule has an empty options schema, so defaults
			// and option combinations do not apply. ----
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: single- and multi-level receiver parentheses are
			// transparent for detection and remain intact after the fix. ----
			depthInvalid(`(array).flat(1)`, `1`, `(array).flat()`, "file.js"),
			depthInvalid(`((array)).flat(1)`, `1`, `((array)).flat()`, "file.js"),

			// ---- Dimension 4: TypeScript non-null, assertion, and satisfies
			// wrappers on the receiver do not hide an otherwise eligible call. ----
			depthInvalid(`array!.flat(1)`, `1`, `array!.flat()`, "file.ts"),
			depthInvalid(`(array as any).flat(1)`, `1`, `(array as any).flat()`, "file.ts"),
			depthInvalid(
				`(array satisfies unknown).flat(1)`,
				`1`,
				`(array satisfies unknown).flat()`,
				"file.ts",
			),
			depthInvalid(`array?.value.flat(1)`, `1`, `array?.value.flat()`, "file.js"),

			// ---- Dimension 4: parentheses around the inspected argument are
			// transparent, but the full wrapper is removed by the fix. ----
			depthInvalid(`array.flat((1))`, `1`, `array.flat()`, "file.js"),
			depthInvalid(`array.flat(((0x01)))`, `0x01`, `array.flat()`, "file.js"),

			// Locks in upstream create() report arm: directly visible mismatched
			// receivers and syntactic arrays still report rather than being skipped.
			depthInvalid(`"text".flat(1)`, `1`, `"text".flat()`, "file.js"),
			{
				Code:     `[[1]].flat(1)`,
				FileName: "file.js",
				Output:   []string{`[[1]].flat()`},
				Errors: []rule_tester.InvalidTestCaseError{
					expectedDepthError(`[[1]].flat(1)`, `1`, 1),
				},
			},

			// ---- Dimension 4: TypeScript type arguments do not change the call
			// shape and remain intact when the depth is removed. ----
			depthInvalid(`array.flat<number>(1)`, `1`, `array.flat<number>()`, "file.ts"),

			// ---- Dimension 3: the only argument's trailing comma is removed. ----
			depthInvalid(`array.flat(1,)`, `1`, `array.flat()`, "file.js"),

			// ---- Dimension 3: comments before and after the argument are
			// preserved with the same surrounding whitespace as upstream. ----
			depthInvalid(
				`array.flat(/* keep */ 1)`,
				`1`,
				`array.flat(/* keep */ )`,
				"file.js",
			),
			depthInvalid(
				`array.flat(1 /* keep */)`,
				`1`,
				`array.flat( /* keep */)`,
				"file.js",
			),
			depthInvalid(
				`array.flat(/* before */ (1) /* after */,)`,
				`1`,
				`array.flat(/* before */ )`,
				"file.js",
			),
			depthInvalid(
				"array.flat(\n  /* before */\n  1,\n)",
				`1`,
				"array.flat(\n  /* before */\n  \n)",
				"file.js",
			),

			// ---- Dimension 4: same-kind nesting and siblings are visited without
			// bleeding across call boundaries; every literal receives one report. ----
			{
				Code:     `array.flat(1).flat(1)`,
				FileName: "file.js",
				Output:   []string{`array.flat().flat()`},
				Errors: []rule_tester.InvalidTestCaseError{
					expectedDepthError(`array.flat(1).flat(1)`, `1`, 0),
					expectedDepthError(`array.flat(1).flat(1)`, `1`, 1),
				},
			},
			{
				Code:     `const first = a.flat(1); const second = b.flat(1);`,
				FileName: "file.js",
				Output:   []string{`const first = a.flat(); const second = b.flat();`},
				Errors: []rule_tester.InvalidTestCaseError{
					expectedDepthError(`const first = a.flat(1); const second = b.flat(1);`, `1`, 0),
					expectedDepthError(`const first = a.flat(1); const second = b.flat(1);`, `1`, 1),
				},
			},
		},
	)
}

func TestNoUnnecessaryArrayFlatDepthEditDemand(t *testing.T) {
	t.Parallel()

	const fileName = "edit-demand.ts"
	const source = `const flattened = rows.flat(1);`
	const fixedSource = `const flattened = rows.flat();`

	compilerProgram, sourceFile := createNoUnnecessaryArrayFlatDepthProgram(t, fileName, source)
	diagnostics := make(map[rule.EditDemand]rule.RuleDiagnostic, 4)
	for _, demand := range []rule.EditDemand{
		rule.EditDemandNone,
		rule.EditDemandAutofix,
		rule.EditDemandSuggestion,
		rule.EditDemandAll,
	} {
		got := lintNoUnnecessaryArrayFlatDepthWithDemand(compilerProgram, sourceFile, demand)
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

func createNoUnnecessaryArrayFlatDepthProgram(
	t testing.TB,
	fileName string,
	source string,
) (*compiler.Program, *ast.SourceFile) {
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

func lintNoUnnecessaryArrayFlatDepthWithDemand(
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
				Name:     no_unnecessary_array_flat_depth.NoUnnecessaryArrayFlatDepthRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return no_unnecessary_array_flat_depth.NoUnnecessaryArrayFlatDepthRule.Run(ctx, nil)
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
