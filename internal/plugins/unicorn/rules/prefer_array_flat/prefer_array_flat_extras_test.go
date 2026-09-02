package prefer_array_flat_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_array_flat"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestPreferArrayFlatExtras locks in branches and edge shapes that the
// upstream test suite does not exercise. Each case carries an inline comment
// pointing at the specific branch, Dimension 4 row, or upstream issue it
// covers, so future refactors cannot silently regress it without breaking a
// named lock-in. The 1:1 upstream migration lives in the
// prefer_array_flat_upstream_*_test.go files.
func TestPreferArrayFlatExtras(t *testing.T) {
	var suite upstreamSuite

	// ---- Dimension 4: single- and multi-level parenthesized receivers/body ----
	suite.addFixed(
		`((array)).flatMap(x => (x))`,
		`((array)).flatMap(x => (x))`,
		`((array)).flat()`,
		`Array#flatMap()`,
		nil,
	)
	suite.addFixed(
		`((array)).reduce((a, b) => ((a.concat(b))), (([])))`,
		`((array)).reduce((a, b) => ((a.concat(b))), (([])))`,
		`((array)).flat()`,
		`Array#reduce()`,
		nil,
	)

	// ---- Dimension 4: TS receiver/expression wrappers ----
	suite.addFixed(
		`array!.flatMap(x => x)`,
		`array!.flatMap(x => x)`,
		`(array!).flat()`,
		`Array#flatMap()`,
		nil,
	)
	suite.addFixed(
		`(array as unknown[]).flatMap(x => x)`,
		`(array as unknown[]).flatMap(x => x)`,
		`(array as unknown[]).flat()`,
		`Array#flatMap()`,
		nil,
	)
	suite.addFixed(
		`(array satisfies unknown[]).flatMap(x => x)`,
		`(array satisfies unknown[]).flatMap(x => x)`,
		`(array satisfies unknown[]).flat()`,
		`Array#flatMap()`,
		nil,
	)
	suite.addValid(nil, `[].concat((maybeArray as unknown))`)
	suite.addFixed(
		`_.flatten(array!)`,
		`_.flatten(array!)`,
		`(array!).flat()`,
		`_.flatten()`,
		nil,
	)
	suite.addValid(nil,
		`array.flatMap(x => x!)`,
		`array.reduce((a, b) => (a.concat(b) as unknown[]), [])`,
	)

	// ---- Dimension 4: type annotations are transparent on plain parameters ----
	suite.addFixed(
		`array.reduce((a: unknown[], b: unknown[]) => (a.concat(b)), [])`,
		`array.reduce((a: unknown[], b: unknown[]) => (a.concat(b)), [])`,
		`array.flat()`,
		`Array#reduce()`,
		nil,
	)

	// ---- Dimension 4: optional call/member boundaries ----
	suite.addValid(nil,
		`array.flatMap?.(x => x)`,
		`(array?.flatMap)(x => x)`,
		`array.reduce((a, b) => a?.concat(b), [])`,
		`array.reduce((a, b) => a.concat?.(b), [])`,
		`_.flatten?.(array)`,
		`_?.flatten(array)`,
	)
	suite.addFixed(
		`array?.items.flatMap(x => x)`,
		`array?.items.flatMap(x => x)`,
		`array?.items.flat()`,
		`Array#flatMap()`,
		nil,
	)

	// ---- Dimension 4: computed/static/numeric/Symbol access stays unmatched ----
	suite.addValid(nil,
		`array["flatMap"](x => x)`,
		"array[`flatMap`](x => x)",
		`array[0](x => x)`,
		`array[Symbol.iterator](x => x)`,
		`array["reduce"]((a, b) => a.concat(b), [])`,
		`[]["concat"](array)`,
		`_["flatten"](array)`,
	)

	// ---- Dimension 4: TypeScript type arguments are removed with the legacy call ----
	suite.addFixed(
		`array.flatMap<unknown>(x => x)`,
		`array.flatMap<unknown>(x => x)`,
		`array.flat()`,
		`Array#flatMap()`,
		nil,
	)
	suite.addFixed(
		`_.flatten<Array<unknown>>(array)`,
		`_.flatten<Array<unknown>>(array)`,
		`array.flat()`,
		`_.flatten()`,
		nil,
	)

	// ---- Dimension 4: comments around parens suppress fixes like ESTree ----
	suite.addNoFix(
		`_.flatten((foo /* keep */))`,
		`_.flatten((foo /* keep */))`,
		`_.flatten()`,
		nil,
	)
	suite.addNoFix(
		`[].concat(...((foo /* keep */)))`,
		`[].concat(...((foo /* keep */)))`,
		`[].concat()`,
		nil,
	)
	suite.addFixed(
		`_.flatten((foo./* keep */bar))`,
		`_.flatten((foo./* keep */bar))`,
		`(foo./* keep */bar).flat()`,
		`_.flatten()`,
		nil,
	)

	// ---- Dimension 4: side-effecting arguments remain single-evaluation fixes ----
	suite.addFixed(
		`_.flatten(getArray())`,
		`_.flatten(getArray())`,
		`getArray().flat()`,
		`_.flatten()`,
		nil,
	)
	suite.addFixed(
		`[].concat(...getArray())`,
		`[].concat(...getArray())`,
		`getArray().flat()`,
		`[].concat()`,
		nil,
	)

	// ---- Dimension 4: ASI handling must not detach an embedded statement ----
	suite.addFixed(
		`if (condition)
			Array.prototype.concat.call([], ...value)`,
		`Array.prototype.concat.call([], ...value)`,
		`value.flat()`,
		`Array.prototype.concat()`,
		nil,
	)

	// ---- Dimension 4: multi-line report and fix range ----
	suite.addFixed(
		`const result = array.reduce(
			(a, b) => [...a, ...b],
			[],
		);`,
		`array.reduce(
			(a, b) => [...a, ...b],
			[],
		)`,
		`array.flat()`,
		`Array#reduce()`,
		nil,
	)

	// ---- Dimension 4: graceful degradation for empty/spread/holey shapes ----
	suite.addValid(nil,
		`array.flatMap()`,
		`array.flatMap(...callbacks)`,
		`array.reduce()`,
		`array.reduce((a, b) => [...a, ...b,,], [])`,
		`[].concat()`,
		`_.flatten()`,
		`_.flatten(...arrays)`,
	)

	// ---- Real-user: #2660 parenthesized non-array object receiver ----
	suite.addValid(nil, `
		const randomObject = {
			flatMap(function_: (value: unknown) => void) {
				function_(1);
			},
		};
		(randomObject).flatMap(value => value);
	`)

	// ---- Real-user: v74 leaves plain concat normalization to prefer-spread ----
	suite.addValid(nil,
		`declare const subAppLoads: SubAppLoad | SubAppLoad[];
		const loads: SubAppLoad[] = [].concat(subAppLoads);`,
		`const values = [].concat([1]).concat([2]);`,
	)

	// Locks in upstream arrayFlatMap.testFunction() arm 1: PascalCase unknowns
	// are treated as obvious non-array receivers.
	suite.addValid(nil, `function consume(Items: unknown) { return Items.flatMap(x => x); }`)

	// Locks in upstream arrayFlatMap.testFunction() arm 2: a PascalCase const
	// initialized with an array remains eligible.
	suite.addFixed(
		`const Items = new Array(); Items.flatMap(item => item);`,
		`Items.flatMap(item => item)`,
		`Items.flat()`,
		`Array#flatMap()`,
		nil,
	)

	// Locks in upstream arrayFlatMap.testFunction() arm 3: a lower-case const
	// initialized with a known non-array remains ignored.
	suite.addValid(nil, `const collection = class {}; collection.flatMap(x => x);`)

	// Locks in upstream arrayReduce.testFunction() arm 1: concat reducer.
	suite.addFixed(
		`items.reduce((left, right) => left.concat(right), [])`,
		`items.reduce((left, right) => left.concat(right), [])`,
		`items.flat()`,
		`Array#reduce()`,
		nil,
	)

	// Locks in upstream arrayReduce.testFunction() arm 2: spread reducer.
	suite.addFixed(
		`items.reduce((left, right) => [...left, ...right], [])`,
		`items.reduce((left, right) => [...left, ...right], [])`,
		`items.flat()`,
		`Array#reduce()`,
		nil,
	)

	// v74 skips only receivers proven non-array. An all-non-array union and an
	// asserted Set are skipped; unknown and mixed unions remain reportable.
	suite.addValid(nil,
		`function f(foo: Set<number[]> | Uint8Array) { foo.reduce((a, b) => a.concat(b), []); }`,
		`declare const foo: unknown; (foo as Set<number[]>).reduce((a, b) => a.concat(b), []);`,
		`function f(foo: Set<number[]>) { foo!.reduce((a, b) => [...a, ...b], []); }`,
	)
	suite.addFixed(
		`function f(foo: Set<number[]> | number[][]) { foo.reduce((a, b) => a.concat(b), []); }`,
		`foo.reduce((a, b) => a.concat(b), [])`,
		`foo.flat()`,
		`Array#reduce()`,
		nil,
	)
	suite.addFixed(
		`declare const foo: unknown; (foo satisfies Set<number[]>).reduce((a, b) => a.concat(b), []);`,
		`(foo satisfies Set<number[]>).reduce((a, b) => a.concat(b), [])`,
		`(foo satisfies Set<number[]>).flat()`,
		`Array#reduce()`,
		nil,
	)
	suite.valid = append(suite.valid, rule_tester.ValidTestCase{
		Code:     `const foo = {}; foo.reduce((a, b) => a.concat(b), []);`,
		FileName: "file.js",
	})
	const unknownReduce = `let foo; foo.reduce((a, b) => a.concat(b), []);`
	suite.invalid = append(suite.invalid, rule_tester.InvalidTestCase{
		Code:     unknownReduce,
		FileName: "file.js",
		Output:   []string{`let foo; foo.flat();`},
		Errors: []rule_tester.InvalidTestCaseError{
			upstreamError(
				unknownReduce,
				`foo.reduce((a, b) => a.concat(b), [])`,
				`Array#reduce()`,
				0,
			),
		},
	})

	// Locks in upstream arrayReduce.testFunction() arm 3: default/rest
	// parameters are not plain identifiers.
	suite.addValid(nil,
		`items.reduce((left = [], right) => left.concat(right), [])`,
		`items.reduce((...values) => values[0].concat(values[1]), [])`,
	)

	// Locks in upstream emptyArrayConcat: only a spread argument is a
	// flattening pattern; plain concat normalization remains valid.
	suite.addValid(nil, `[].concat(value)`)
	suite.addFixed(
		`[].concat(...values)`,
		`[].concat(...values)`,
		`values.flat()`,
		`[].concat()`,
		nil,
	)

	// Locks in upstream arrayPrototypeConcat.testFunction() arms: apply accepts
	// only a non-spread second argument, while call accepts only spread.
	suite.addValid(nil,
		`Array.prototype.concat.apply([], ...values)`,
		`Array.prototype.concat.call([], value)`,
	)
	suite.addFixed(
		`Array.prototype.concat.call([], ...values)`,
		`Array.prototype.concat.call([], ...values)`,
		`values.flat()`,
		`Array.prototype.concat()`,
		nil,
	)

	// Locks in upstream isNodeMatchesNameOrPath() roots: this, super, and meta
	// properties all participate in configured dotted paths.
	pathOptions := map[string]any{
		"functions": []any{
			"this.flatten",
			"super.flatten",
			"import.meta.flatten",
			"new.target.flatten",
		},
	}
	suite.addFixed(
		`class A { method() { return this.flatten(values); } }`,
		`this.flatten(values)`,
		`values.flat()`,
		`this.flatten()`,
		pathOptions,
	)
	suite.addFixed(
		`class A extends B { method() { return super.flatten(values); } }`,
		`super.flatten(values)`,
		`values.flat()`,
		`super.flatten()`,
		pathOptions,
	)
	suite.addFixed(
		`function flatten(array) { this(array) }`,
		`this(array)`,
		`array.flat()`,
		`this()`,
		map[string]any{"functions": []any{"this"}},
	)
	suite.addFixed(
		`class A extends B { constructor(array) { super(array) } }`,
		`super(array)`,
		`array.flat()`,
		`super()`,
		map[string]any{"functions": []any{"super"}},
	)
	suite.addFixed(
		`const values = import.meta.flatten(input);`,
		`import.meta.flatten(input)`,
		`input.flat()`,
		`import.meta.flatten()`,
		pathOptions,
	)
	suite.addFixed(
		`function flatten() { return new.target.flatten(input); }`,
		`new.target.flatten(input)`,
		`input.flat()`,
		`new.target.flatten()`,
		pathOptions,
	)

	// Locks in upstream create() comment branch: comments outside the selected
	// array suppress the fix, but the diagnostic remains.
	suite.addNoFix(
		`[] /* receiver */.concat(...value)`,
		`[] /* receiver */.concat(...value)`,
		`[].concat()`,
		nil,
	)

	// Locks in upstream fix() member-object branches for `new` with and without
	// constructor parentheses.
	suite.addFixed(
		`_.flatten(new Collection)`,
		`_.flatten(new Collection)`,
		`(new Collection).flat()`,
		`_.flatten()`,
		nil,
	)
	suite.addFixed(
		`_.flatten(new Collection())`,
		`_.flatten(new Collection())`,
		`new Collection().flat()`,
		`_.flatten()`,
		nil,
	)

	// Locks in upstream create() traversal: nested matching calls each report,
	// and overlapping fixes settle over two passes.
	nestedOptions := map[string]any{
		"functions": []any{"flat"},
	}
	nestedCode := `flat(_.flatten(array))`
	suite.invalid = append(suite.invalid, rule_tester.InvalidTestCase{
		Code: nestedCode,
		Output: []string{
			`_.flatten(array).flat()`,
			`array.flat().flat()`,
		},
		Options: nestedOptions,
		Errors: []rule_tester.InvalidTestCaseError{
			upstreamError(nestedCode, nestedCode, `flat()`, 0),
			upstreamError(nestedCode, `_.flatten(array)`, `_.flatten()`, 0),
		},
	})

	// Locks in option defaults: no options and an explicit empty object produce
	// the same default Lodash diagnostic and fix.
	suite.addFixed(
		`_.flatten(values)`,
		`_.flatten(values)`,
		`values.flat()`,
		`_.flatten()`,
		map[string]any{},
	)

	// N/A: string/numeric/private object or class property keys are not rule
	// inputs; only call callee member accesses are inspected.
	// N/A: class declaration/expression and function declaration/expression
	// variants are containers only; the rule independently visits CallExpression.
	// N/A: async/generator declaration variants do not change call matching.
	// N/A: ancestor scope walks, this-binding boundaries, and static blocks do
	// not apply; only const-variable receiver classification consults a symbol.
	// N/A: object SpreadAssignment, binding RestElement, empty destructuring,
	// overload signatures, abstract members, and declare members contain no
	// candidate call by themselves and are ignored by the CallExpression listener.

	suite.run(t)
}

func TestPreferArrayFlatEditDemand(t *testing.T) {
	t.Parallel()

	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(
		`array.flatMap(value => value);`,
		"edit-demand.ts",
		"tsconfig.json",
	)
	if err != nil {
		t.Fatal(err)
	}

	run := func(demand rule.EditDemand) rule.RuleDiagnostic {
		t.Helper()

		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program:     lintprogram.NewFromCompiler(program),
			File:        sourceFile.FileName(),
			HasTypeInfo: true,
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{
					Name:     prefer_array_flat.PreferArrayFlatRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return prefer_array_flat.PreferArrayFlatRule.Run(ctx, nil)
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
		if len(diagnostics) != 1 {
			t.Fatalf("demand %d: diagnostics = %d, want 1", demand, len(diagnostics))
		}
		return diagnostics[0]
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
	for demand, diagnostic := range map[rule.EditDemand]rule.RuleDiagnostic{
		rule.EditDemandNone:       diagnosticsOnly,
		rule.EditDemandAutofix:    autofixOnly,
		rule.EditDemandSuggestion: suggestionOnly,
	} {
		if got, want := withoutEdits(diagnostic), withoutEdits(allEdits); !reflect.DeepEqual(got, want) {
			t.Errorf("demand %d changed diagnostic identity:\ngot  %#v\nwant %#v", demand, got, want)
		}
	}
	if diagnosticsOnly.FixesPtr != nil || suggestionOnly.FixesPtr != nil {
		t.Fatal("non-autofix demand materialized fixes")
	}
	if autofixOnly.FixesPtr == nil || !reflect.DeepEqual(autofixOnly.FixesPtr, allEdits.FixesPtr) {
		t.Fatal("autofix and all-edits demands produced different fixes")
	}
	if fixes := *allEdits.FixesPtr; len(fixes) == 0 {
		t.Fatal("all-edits demand produced no fixes")
	}
	for _, diagnostic := range []rule.RuleDiagnostic{diagnosticsOnly, autofixOnly, suggestionOnly, allEdits} {
		if diagnostic.Suggestions != nil {
			t.Fatal("autofix-only rule materialized suggestions")
		}
	}
}
