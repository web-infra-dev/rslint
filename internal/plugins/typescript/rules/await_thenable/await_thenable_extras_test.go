package await_thenable

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Expectations were checked against typescript-eslint v8.71.0 with TypeScript
// 6.0.2. The computed object key exception is documented in the rule's docs.
func TestAwaitThenablePromiseAggregatorsExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &AwaitThenableRule, []rule_tester.ValidTestCase{
		// all-promises
		{Code: `Promise.all([Promise.resolve(1), Promise.resolve(42)]);
export {};`},
		// allSettled-promises
		{Code: `Promise.allSettled([Promise.resolve(1), Promise.resolve(42)]);
export {};`},
		// any-promises
		{Code: `Promise.any([Promise.resolve(1), Promise.resolve(42)]);
export {};`},
		// race-promises
		{Code: `Promise.race([Promise.resolve(1), Promise.resolve(42)]);
export {};`},
		// subclass
		{Code: `class P extends Promise<unknown> {} P.all([1]);
export {};`},
		// subclass-alias
		{Code: `class P extends Promise<unknown> {} const Q = P; Q.any([1]);
export {};`},
		// shadow
		{Code: `const Promise = { all: (values: unknown[]) => values }; Promise.all([1]);
export {};`},
		// fake-constructor-name
		{Code: `interface PromiseConstructor { all(values: unknown[]): unknown; } declare const P: PromiseConstructor; P.all([1]);
export {};`},
		// object-method
		{Code: `({ all: (values: unknown[]) => values }).all([1]);
export {};`},
		// destructured
		{Code: `const { all } = Promise; all([1]);
export {};`},
		// bound-call
		{Code: `Promise.all.call(Promise, [1]);
export {};`},
		// thenable-constructor
		{Code: `declare const P: { all(values: unknown[]): Promise<unknown[]> }; P.all([1]);
export {};`},
		// non-aggregator
		{Code: `Promise.resolve([1]); Promise.reject([1]);
export {};`},
		// constructor-union
		{Code: `declare const P: PromiseConstructor | { all(values: unknown[]): unknown }; P.all([1]);
export {};`},
		// computed-type-only
		{Code: `declare const method: "all"; Promise[method]([1]);
export {};`},
		// computed-mutated
		{Code: `let method: "all" | "race" = "all"; method = "race"; Promise[method]([1]);
export {};`},
		// computed-side-effect
		{Code: `function method(): "all" { return "all"; } Promise[method()]([1]);
export {};`},
		// private-method
		{Code: `class P extends Promise<unknown> { static #all(values: unknown[]) { return values; } static run() { P.#all([1]); } }
export {};`},
		// literal-union
		{Code: `declare const value: number | Promise<number>; Promise.all([value]);
export {};`},
		// literal-any
		{Code: `declare const value: any; Promise.all([value]);
export {};`},
		// literal-unknown
		{Code: `declare const value: unknown; Promise.all([value]);
export {};`},
		// holes
		{Code: `Promise.all([, , Promise.resolve(1), ,]);
export {};`},
		// spread-promises
		{Code: `declare const values: Promise<number>[]; Promise.all([...values]);
export {};`},
		// spread-union
		{Code: `declare const values: (number | Promise<number>)[]; Promise.all([...values]);
export {};`},
		// spread-tuple
		{Code: `declare const values: [number, Promise<number>]; Promise.all([...values]);
export {};`},
		// spread-any
		{Code: `declare const values: any[]; Promise.all([...values]);
export {};`},
		// empty-array
		{Code: `Promise.all([]);
export {};`},
		// empty-tuple-variable
		{Code: `const values: [] = []; Promise.all(values);
export {};`},
		// generic-unconstrained
		{Code: `function f<T>(values: T[]) { Promise.all(values); }
export {};`},
		// generic-union
		{Code: `function f<T extends number | Promise<number>>(values: T[]) { Promise.all(values); }
export {};`},
		// generic-literal
		{Code: `function f<T extends number | Promise<number>>(value: T) { Promise.all([value]); }
export {};`},
		// readonly-interface
		{Code: `interface Values<A, B> extends ReadonlyArray<B> {} declare const values: Values<number, Promise<void>>; Promise.all(values);
export {};`},
		// set-promises
		{Code: `Promise.all(new Set([Promise.resolve(1)]));
export {};`},
		// map-promise-key
		{Code: `Promise.all(new Map([[Promise.resolve(1), 1]]));
export {};`},
		// generator-return
		{Code: `declare const values: Generator<Promise<number>, number, string>; Promise.all(values);
export {};`},
		// async-iterable
		{Code: `declare const values: AsyncIterable<number>; Promise.all(values);
export {};`},
		// structural-iterable
		{Code: `declare const values: { [Symbol.iterator](): Iterator<number> }; Promise.all(values);
export {};`},
		// Iterable without type parameters.
		{Code: `interface Values { [Symbol.iterator](): Iterator<number>; } declare const values: Values; Promise.all(values);
export {};`},
		// misleading-iterable-arg
		{Code: `interface Values<A, B> extends Iterable<B> {} declare const values: Values<Promise<number>, number>; Promise.all(values);
export {};`},
		// numeric-index-iterable
		{Code: `interface Values<A> extends Iterable<Promise<number>> { [index: number]: number; } declare const values: Values<Promise<number>>; Promise.all(values);
export {};`},
		// nullable-iterable
		{Code: `declare const values: number[] | null; Promise.all(values);
export {};`},
		// optional-iterable
		{Code: `declare const values: number[] | undefined; Promise.all(values);
export {};`},
		// Non-iterable input.
		{Code: `Promise.all(1);
export {};`},
		// string-iterable
		{Code: `Promise.all("abc");
export {};`},
		// string-object
		{Code: `Promise.all(new String("abc"));
export {};`},
		// any-argument
		{Code: `declare const values: any; Promise.all(values);
export {};`},
		// unknown-argument
		{Code: `declare const values: unknown; Promise.all(values);
export {};`},
		// never-argument
		{Code: `declare const values: never; Promise.all(values);
export {};`},
		// missing-argument
		{Code: `Promise.all();
export {};`},
		// extra-argument
		{Code: `Promise.all([Promise.resolve(1)], [1]);
export {};`},
		// thenable
		{Code: `declare const value: { then(onfulfilled: (value: number) => void): void }; Promise.all([value]);
export {};`},
		// optional-then
		{Code: `declare const value: { then?: (onfulfilled: (value: number) => void) => void }; Promise.all([value]);
export {};`},
	}, []rule_tester.InvalidTestCase{
		// all-literal
		{Code: `Promise.all([Promise.resolve(1), 42]);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 34, EndLine: 1, EndColumn: 36},
			},
		},
		// all-array-union
		{Code: `declare const values: (Promise<number> | number)[]; Promise.all(values);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 65, EndLine: 1, EndColumn: 71},
			},
		},
		// all-brackets
		{Code: `Promise['all']([1]);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
			},
		},
		// all-optional
		{Code: `Promise?.all?.([1]);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
			},
		},
		// allSettled-literal
		{Code: `Promise.allSettled([Promise.resolve(1), 42]);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 41, EndLine: 1, EndColumn: 43},
			},
		},
		// allSettled-array-union
		{Code: `declare const values: (Promise<number> | number)[]; Promise.allSettled(values);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 72, EndLine: 1, EndColumn: 78},
			},
		},
		// allSettled-brackets
		{Code: `Promise['allSettled']([1]);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 24, EndLine: 1, EndColumn: 25},
			},
		},
		// allSettled-optional
		{Code: `Promise?.allSettled?.([1]);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 24, EndLine: 1, EndColumn: 25},
			},
		},
		// any-literal
		{Code: `Promise.any([Promise.resolve(1), 42]);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 34, EndLine: 1, EndColumn: 36},
			},
		},
		// any-array-union
		{Code: `declare const values: (Promise<number> | number)[]; Promise.any(values);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 65, EndLine: 1, EndColumn: 71},
			},
		},
		// any-brackets
		{Code: `Promise['any']([1]);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
			},
		},
		// any-optional
		{Code: `Promise?.any?.([1]);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
			},
		},
		// race-literal
		{Code: `Promise.race([Promise.resolve(1), 42]);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 35, EndLine: 1, EndColumn: 37},
			},
		},
		// race-array-union
		{Code: `declare const values: (Promise<number> | number)[]; Promise.race(values);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 66, EndLine: 1, EndColumn: 72},
			},
		},
		// race-brackets
		{Code: `Promise['race']([1]);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 18, EndLine: 1, EndColumn: 19},
			},
		},
		// race-optional
		{Code: `Promise?.race?.([1]);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 18, EndLine: 1, EndColumn: 19},
			},
		},
		// handoff
		{Code: `export async function run() {
  await Promise.all([Promise.resolve(1), 42]);
}
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 2, Column: 42, EndLine: 2, EndColumn: 44},
			},
		},
		// alias
		{Code: `const P = Promise; P.all([1]);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 27, EndLine: 1, EndColumn: 28},
			},
		},
		// constructor-interface
		{Code: `interface P extends PromiseConstructor {} declare const p: P; p.all([1]);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 70, EndLine: 1, EndColumn: 71},
			},
		},
		// constructor-generic
		{Code: `function f<P extends PromiseConstructor>(p: P) { p.all([1]); }
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 57, EndLine: 1, EndColumn: 58},
			},
		},
		// constructor-intersection
		{Code: `declare const P: PromiseConstructor & { tag: true }; P.all([1]);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 61, EndLine: 1, EndColumn: 62},
			},
		},
		// computed-const
		{Code: `const method = 'all'; Promise[method]([1]);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 40, EndLine: 1, EndColumn: 41},
			},
		},
		// computed-folded
		{Code: `Promise['a' + 'll']([1]);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
			},
		},
		// computed-template
		{Code: "const end = \"ll\"; Promise[`a${end}`]([1]);\nexport {};",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 39, EndLine: 1, EndColumn: 40},
			},
		},
		// computed-frozen
		{Code: `const names = Object.freeze({aggregate: "all"}); Promise[names.aggregate]([1]);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 76, EndLine: 1, EndColumn: 77},
			},
		},
		// computed-string-call
		{Code: `Promise[String("all")]([1]);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 25, EndLine: 1, EndColumn: 26},
			},
		},
		// computed-assertion
		{Code: `Promise["all" as const]([1]);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 26, EndLine: 1, EndColumn: 27},
			},
		},
		// parentheses
		{Code: `(Promise.all)(([(1), Promise.resolve(2)]));
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 18, EndLine: 1, EndColumn: 19},
			},
		},
		// asserted-array
		{Code: `Promise.all([1, Promise.resolve(2)] as const);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 13, EndLine: 1, EndColumn: 45},
			},
		},
		// satisfies-array
		{Code: `Promise.all([1, Promise.resolve(2)] satisfies unknown[]);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 13, EndLine: 1, EndColumn: 56},
			},
		},
		// non-null-array
		{Code: `Promise.all([1, Promise.resolve(2)]!);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 13, EndLine: 1, EndColumn: 37},
			},
		},
		// literal-never
		{Code: `declare const value: never; Promise.all([value]);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 42, EndLine: 1, EndColumn: 47},
			},
		},
		// literal-nullish
		{Code: `Promise.all([null, undefined, true, "text", {}]);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 14, EndLine: 1, EndColumn: 18},
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 20, EndLine: 1, EndColumn: 29},
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 31, EndLine: 1, EndColumn: 35},
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 37, EndLine: 1, EndColumn: 43},
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 45, EndLine: 1, EndColumn: 47},
			},
		},
		// spread-numbers
		{Code: `declare const values: number[]; Promise.all([...values]);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 46, EndLine: 1, EndColumn: 55},
			},
		},
		// spread-empty
		{Code: `Promise.all([...[]]);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 14, EndLine: 1, EndColumn: 19},
			},
		},
		// call-spread
		{Code: `const args: [number[]] = [[1]]; Promise.all(...args);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 45, EndLine: 1, EndColumn: 52},
			},
		},
		// nested-array
		{Code: `Promise.all([[Promise.resolve(1)]]);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 14, EndLine: 1, EndColumn: 34},
			},
		},
		// never-array
		{Code: `declare const values: never[]; Promise.all(values);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 44, EndLine: 1, EndColumn: 50},
			},
		},
		// optional-tuple
		{Code: `declare const values: [Promise<number>?]; Promise.all(values);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 55, EndLine: 1, EndColumn: 61},
			},
		},
		// variadic-tuple
		{Code: `declare const values: [Promise<number>, ...number[]]; Promise.all(values);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 67, EndLine: 1, EndColumn: 73},
			},
		},
		// readonly-tuple
		{Code: `declare const values: readonly [Promise<number>, number]; Promise.all(values);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 71, EndLine: 1, EndColumn: 77},
			},
		},
		// array-intersection
		{Code: `declare const values: number[] & { tag: true }; Promise.all(values);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 61, EndLine: 1, EndColumn: 67},
			},
		},
		// generic-number
		{Code: `function f<T extends number>(values: T[]) { Promise.all(values); }
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 57, EndLine: 1, EndColumn: 63},
			},
		},
		// generic-array
		{Code: `function f<T extends number[]>(values: T) { Promise.all(values); }
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 57, EndLine: 1, EndColumn: 63},
			},
		},
		// array-subclass
		{Code: `class Values<A, B> extends Array<B> {} declare const values: Values<Promise<void>, number>; Promise.all(values);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 105, EndLine: 1, EndColumn: 111},
			},
		},
		// set-number
		{Code: `Promise.all(new Set([1]));
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 13, EndLine: 1, EndColumn: 25},
			},
		},
		// map-number
		{Code: `Promise.all(new Map([[1, Promise.resolve(1)]]));
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 13, EndLine: 1, EndColumn: 47},
			},
		},
		// generator-yield
		{Code: `declare const values: Generator<number, Promise<number>, unknown>; Promise.all(values);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 80, EndLine: 1, EndColumn: 86},
			},
		},
		// custom-iterable
		{Code: `interface Values<T> { [Symbol.iterator](): Iterator<T>; } declare const values: Values<number>; Promise.all(values);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 109, EndLine: 1, EndColumn: 115},
			},
		},
		// misleading-iterable-arg-reverse
		{Code: `interface Values<A, B> extends Iterable<B> {} declare const values: Values<number, Promise<number>>; Promise.all(values);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 114, EndLine: 1, EndColumn: 120},
			},
		},
		// mixed-iterable
		{Code: `declare const values: number[] | Iterable<Promise<number>>; Promise.all(values);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 73, EndLine: 1, EndColumn: 79},
			},
		},
		// Non-callable then property.
		{Code: `Promise.all([{ then: 1 }]);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 14, EndLine: 1, EndColumn: 25},
			},
		},
		// Then method without a callback.
		{Code: `Promise.all([{ then() {} }]);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 14, EndLine: 1, EndColumn: 27},
			},
		},
		// unicode-location
		{Code: `const text = "😀"; Promise.all([1, 2]);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 33, EndLine: 1, EndColumn: 34},
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 36, EndLine: 1, EndColumn: 37},
			},
		},
		// nested-calls
		{Code: `Promise.all([Promise.all([1]), 2]);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 32, EndLine: 1, EndColumn: 33},
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 27, EndLine: 1, EndColumn: 28},
			},
		},
		// comments-range
		{Code: `Promise.all([ /* before */ (42) /* after */ ]);
export {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPromiseAggregatorInput", Line: 1, Column: 29, EndLine: 1, EndColumn: 31},
			},
		},
	})
}

func runAwaitThenableWithDemand(t *testing.T, code string, demand rule.EditDemand) []rule.RuleDiagnostic {
	t.Helper()
	root := fixtures.GetRootDir()
	fileName := tspath.ResolvePath(root.Dir, "edit-demand.ts")
	fs := utils.NewOverlayVFS(root.FS, map[string]string{fileName: code})
	program, err := utils.CreateProgram(true, fs, root.Dir, "tsconfig.json", utils.CreateCompilerHost(root.Dir, fs))
	if err != nil {
		t.Fatal(err)
	}
	source := program.GetSourceFile(fileName)
	if source == nil {
		t.Fatal("missing edit-demand source")
	}
	typeChecker, release := program.GetTypeCheckerForFile(context.Background(), source)
	defer release()
	comments := rule.NewCommentStore(source)
	var diagnostics []rule.RuleDiagnostic
	ctx := (rule.RuleContext{
		SourceFile: source, TypeChecker: typeChecker, Comments: comments,
		DisableManager: rule.NewDisableManager(source, comments),
	}).WithProgram(lintprogram.NewFromCompiler(program)).WithDiagnosticConsumer(
		"@typescript-eslint/await-thenable", rule.SeverityError, rule.DiagnosticConsumer{
			Demand: demand,
			Report: func(diagnostic rule.RuleDiagnostic) { diagnostics = append(diagnostics, diagnostic) },
		},
	)
	listeners := AwaitThenableRule.Run(ctx, nil)
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if listener := listeners[node.Kind]; listener != nil {
			listener(node)
		}
		return node.ForEachChild(visit)
	}
	visit(source.AsNode())
	return diagnostics
}

func TestAwaitThenableEditDemand(t *testing.T) {
	for _, fixture := range []struct {
		name, code, reported, output string
		message, suggestion          rule.RuleMessage
	}{
		{
			name:     "await with trivia",
			code:     "async function run() { /* lead */ await /* keep */ 1; }",
			reported: "await /* keep */ 1",
			output:   "async function run() { /* lead */  /* keep */ 1; }",
			message:  buildAwaitMessage(), suggestion: buildRemoveAwaitMessage(),
		},
		{
			name:     "for await with trivia",
			code:     "async function run() { for /* lead */ await /* keep */ (const value of [1]) {} }",
			reported: "for /* lead */ await /* keep */ (const value of [1])",
			output:   "async function run() { for /* lead */  /* keep */ (const value of [1]) {} }",
			message:  buildForAwaitOfNonAsyncIterableMessage(), suggestion: buildConvertToOrdinaryForMessage(),
		},
		{
			name:     "await using with trivia",
			code:     "declare const resource: Disposable; async function run() { /* lead */ await /* keep */ using value = resource; }",
			reported: "resource;",
			output:   "declare const resource: Disposable; async function run() { /* lead */  /* keep */ using value = resource; }",
			message:  buildAwaitUsingOfNonAsyncDisposableMessage(), suggestion: buildRemoveAwaitMessage(),
		},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
				diagnostics := runAwaitThenableWithDemand(t, fixture.code, demand)
				if len(diagnostics) != 1 {
					t.Fatalf("demand %d: got %d diagnostics, want 1", demand, len(diagnostics))
				}
				diagnostic := diagnostics[0]
				start := strings.Index(fixture.code, fixture.reported)
				wantRange := core.NewTextRange(start, start+len(strings.TrimSuffix(fixture.reported, ";")))
				if diagnostic.Range != wantRange || !reflect.DeepEqual(diagnostic.Message, fixture.message) {
					t.Fatalf("demand %d: unexpected diagnostic range or message: %#v", demand, diagnostic)
				}
				if diagnostic.FixesPtr != nil {
					t.Fatalf("demand %d: unexpectedly produced an autofix", demand)
				}
				if demand&rule.EditDemandSuggestion == 0 {
					if diagnostic.Suggestions != nil {
						t.Fatalf("demand %d: unexpectedly produced suggestions", demand)
					}
					continue
				}
				if diagnostic.Suggestions == nil || len(*diagnostic.Suggestions) != 1 {
					t.Fatalf("demand %d: expected exactly one suggestion", demand)
				}
				suggestion := (*diagnostic.Suggestions)[0]
				fixes := suggestion.Fixes()
				if !reflect.DeepEqual(suggestion.Message, fixture.suggestion) || len(fixes) != 1 {
					t.Fatalf("demand %d: unexpected suggestion: %#v", demand, suggestion)
				}
				fix := fixes[0]
				if fixture.code[fix.Range.Pos():fix.Range.End()] != "await" || fix.Text != "" {
					t.Fatalf("demand %d: unexpected removal: %#v", demand, fix)
				}
				output := fixture.code[:fix.Range.Pos()] + fix.Text + fixture.code[fix.Range.End():]
				if output != fixture.output {
					t.Fatalf("demand %d: output = %q, want %q", demand, output, fixture.output)
				}
			}
		})
	}

	const multiple = `declare const resource: Disposable;
async function run() { await using first = resource, second = resource; }`
	const suppressed = `declare const resource: Disposable;
async function run() {
  // eslint-disable-next-line @typescript-eslint/await-thenable
  await 1;
  for await (const value of [1]) {} // eslint-disable-line @typescript-eslint/await-thenable
  /* eslint-disable @typescript-eslint/await-thenable */
  await using first = resource;
  /* eslint-enable @typescript-eslint/await-thenable */
}`
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
		diagnostics := runAwaitThenableWithDemand(t, multiple, demand)
		if len(diagnostics) != 2 {
			t.Fatalf("demand %d: multiple declarations produced %d diagnostics, want 2", demand, len(diagnostics))
		}
		for _, diagnostic := range diagnostics {
			if diagnostic.Message.Id != "awaitUsingOfNonAsyncDisposable" || diagnostic.FixesPtr != nil ||
				(diagnostic.Suggestions != nil && len(*diagnostic.Suggestions) != 0) {
				t.Fatalf("demand %d: unexpected edits for multiple declarations: %#v", demand, diagnostic)
			}
		}
		if diagnostics := runAwaitThenableWithDemand(t, suppressed, demand); len(diagnostics) != 0 {
			t.Fatalf("demand %d: suppressed reports escaped: %#v", demand, diagnostics)
		}
	}
}
