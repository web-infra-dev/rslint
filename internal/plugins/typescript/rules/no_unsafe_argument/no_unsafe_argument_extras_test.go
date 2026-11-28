// TestNoUnsafeArgumentExtras locks in branches and edge shapes that the
// upstream suite in no_unsafe_argument_upstream_test.go does not exercise.
// Each case identifies the upstream branch, universal edge shape, or real-user
// regression it protects.
package no_unsafe_argument

import (
	"context"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func runNoUnsafeArgumentLenientProgram(
	t *testing.T,
	entryFileName string,
	files map[string]string,
) []rule.RuleDiagnostic {
	t.Helper()

	root := fixtures.GetRootDir()
	overlays := make(map[string]string, len(files))
	rootNames := make([]string, 0, len(files))
	for fileName, code := range files {
		filePath := tspath.ResolvePath(root.Dir, fileName)
		overlays[filePath] = code
		rootNames = append(rootNames, filePath)
	}
	fs := utils.NewOverlayVFS(root.FS, overlays)
	host := utils.CreateCompilerHost(root.Dir, fs)
	program, err := utils.CreateProgramFromOptionsLenient(
		false,
		&core.CompilerOptions{},
		rootNames,
		host,
	)
	if err != nil {
		t.Fatalf("create lenient program: %v", err)
	}
	entryPath := tspath.ResolvePath(root.Dir, entryFileName)
	sourceFile := program.GetSourceFile(entryPath)
	if sourceFile == nil {
		t.Fatalf("lenient program did not contain %s", entryFileName)
	}

	var diagnostics []rule.RuleDiagnostic
	programs := []*lintprogram.Program{lintprogram.NewFromCompiler(program)}
	lintPlan, err := linter.PrepareLintPlan(linter.PrepareLintPlanOptions{
		Programs:         programs,
		TargetsByProgram: [][]string{{sourceFile.FileName()}},
		SingleThreaded:   true,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:             NoUnsafeArgumentRule.Name,
				Severity:         rule.SeverityError,
				RequiresTypeInfo: true,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return NoUnsafeArgumentRule.Run(ctx, nil)
				},
			}}
		},
	})
	if err != nil {
		t.Fatalf("PrepareLintPlan: %v", err)
	}
	_, err = linter.RunLinter(linter.RunLinterOptions{
		SingleThreaded: true,
		LintPlan:       lintPlan,
		Consumer: rule.DiagnosticConsumer{
			Demand: rule.EditDemandNone,
			Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			},
		},
	})
	if err != nil {
		t.Fatalf("run linter: %v", err)
	}
	return diagnostics
}

func TestNoUnsafeArgumentTypeScriptVersionBoundary(t *testing.T) {
	t.Parallel()
	// ---- Real-user: rspack packages/rspack/src/config/defaults.ts ----
	// TypeScript 5.9 types tty as boolean, while TypeScript 6 and typescript-go
	// 7 type it as any. The rule follows its checker, matching the current
	// supported upstream combination by reporting the argument.

	const fileName = "file.ts"
	const code = `
import type { InfrastructureLogging } from './no_unsafe_argument_types';

declare const infrastructureLogging: InfrastructureLogging;
declare const process: { env: { TERM?: string } };

const tty =
  (infrastructureLogging as any).stream?.isTTY &&
  process.env.TERM !== 'dumb';

declare function setDefault<T, P extends keyof T>(
  object: T,
  property: P,
  value: T[P],
): void;
setDefault(infrastructureLogging, 'colors', tty);
`
	const typesCode = `
export type InfrastructureLogging = {
  colors?: boolean;
  stream?: { isTTY?: boolean };
};
`

	diagnostics := runNoUnsafeArgumentLenientProgram(t, fileName, map[string]string{
		fileName:                      code,
		"no_unsafe_argument_types.ts": typesCode,
	})
	if len(diagnostics) != 1 {
		t.Fatalf("expected one logical-AND diagnostic, got %#v", diagnostics)
	}
	diagnostic := diagnostics[0]
	if diagnostic.Message.Id != "unsafeArgument" ||
		diagnostic.Message.Description != "Unsafe argument of type `any` assigned to a parameter of type `boolean | undefined`." {
		t.Fatalf("unexpected logical-AND diagnostic: %#v", diagnostic)
	}
	if got := code[diagnostic.Range.Pos():diagnostic.Range.End()]; got != "tty" {
		t.Fatalf("logical-AND diagnostic range covers %q, want tty", got)
	}
}

func TestNoUnsafeArgumentES5ArrayConstraintSpread(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.es5.json",
		t,
		&NoUnsafeArgumentRule,
		[]rule_tester.ValidTestCase{
			{
				Code: `
declare function acceptStrings(...values: string[]): void;

function forward<T extends readonly any[]>(values: T): void {
  acceptStrings(...values);
}
`,
			},
			{
				// Upstream skips array unions even when only ES5 libs are loaded.
				Code: `
declare function acceptStrings(...values: string[]): void;
declare const values: string[] | any[];

acceptStrings(...values);
`,
			},
			{
				// A generic array-union constraint does not make the spread an any[].
				Code: `
declare function acceptStrings(...values: string[]): void;

function forward<T extends string[] | any[]>(values: T): void {
  acceptStrings(...values);
}
`,
			},
		},
		nil,
	)
}

func TestNoUnsafeArgumentLogicalExpressionBoundaries(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnsafeArgumentRule,
		[]rule_tester.ValidTestCase{
			{
				// Locks in the non-strict checker type for a simple unannotated const.
				TSConfig: "tsconfig.unstrict.json",
				Code: `
declare function acceptBoolean(value: boolean): void;
declare const anyValue: any;
const value = anyValue && true;
acceptBoolean(value);
`,
			},
			{
				// ---- Dimension 4: direct, nested, aliased, property, and return shapes ----
				TSConfig: "tsconfig.unstrict.json",
				Code: `
declare function acceptBoolean(value: boolean): void;
declare const anyValue: any;

acceptBoolean(anyValue && true);
acceptBoolean((anyValue && true));
acceptBoolean(anyValue && anyValue && true);
acceptBoolean((anyValue && true) || false);

let letValue = anyValue && true;
var varValue = anyValue && true;
const alias = letValue;
const objectValue = { value: anyValue && true };
function getValue() { return anyValue && true; }

acceptBoolean(letValue);
acceptBoolean(varValue);
acceptBoolean(alias);
acceptBoolean(objectValue.value);
acceptBoolean(getValue());
`,
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				// strictNullChecks keeps `any && boolean` unsafe upstream.
				Code: `
declare function acceptBoolean(value: boolean): void;
declare const anyValue: any;
const value = anyValue && true;
acceptBoolean(value);
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeArgument",
					Line:      5,
					Column:    15,
					EndColumn: 20,
				}},
			},
			{
				// An explicit any annotation remains authoritative.
				TSConfig: "tsconfig.unstrict.json",
				Code: `
declare function acceptBoolean(value: boolean): void;
declare const anyValue: any;
const value: any = anyValue && true;
acceptBoolean(value);
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeArgument",
					Line:      5,
					Column:    15,
					EndColumn: 20,
				}},
			},
			{
				// An any-typed right operand remains unsafe.
				TSConfig: "tsconfig.unstrict.json",
				Code: `
declare function acceptBoolean(value: boolean): void;
declare const anyValue: any;
const value = true && anyValue;
acceptBoolean(value);
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeArgument",
					Line:      5,
					Column:    15,
					EndColumn: 20,
				}},
			},
			{
				// An error-typed right operand retains upstream's error diagnostic.
				TSConfig: "tsconfig.unstrict.json",
				Code: `
declare function acceptBoolean(value: boolean): void;
declare const anyValue: any;
const value = anyValue && missingName;
acceptBoolean(value);
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeArgument",
					Message:   "Unsafe argument of type error typed assigned to a parameter of type `boolean`.",
					Line:      5,
					Column:    15,
					EndLine:   5,
					EndColumn: 20,
				}},
			},
		},
	)
}

func TestNoUnsafeArgumentLogicalAndMutableFlow(t *testing.T) {
	// A later write makes the flow type any. Run this in an isolated Program:
	// RuleTester batches virtual files, which changes this no-strict flow type.
	const fileName = "no_unsafe_argument_mutable_flow.ts"
	const code = `
declare function acceptBoolean(value: boolean): void;
declare const anyValue: any;
let value = anyValue && true;
value = anyValue;
acceptBoolean(value);
`
	diagnostics := runNoUnsafeArgumentLenientProgram(t, fileName, map[string]string{
		fileName: code,
	})
	if len(diagnostics) != 1 {
		t.Fatalf("expected one unsafe mutable-flow diagnostic, got %#v", diagnostics)
	}
	diagnostic := diagnostics[0]
	if diagnostic.Message.Id != "unsafeArgument" ||
		diagnostic.Message.Description != "Unsafe argument of type `any` assigned to a parameter of type `boolean`." {
		t.Fatalf("unexpected mutable-flow diagnostic: %#v", diagnostic)
	}
	if got := code[diagnostic.Range.Pos():diagnostic.Range.End()]; got != "value" {
		t.Fatalf("mutable-flow diagnostic range covers %q, want value", got)
	}
}

func TestNoUnsafeArgumentExtras(t *testing.T) {
	// N/A Dimension 4 access/key forms: the rule does not inspect member keys.
	// N/A Dimension 4 autofix boundaries: the rule has no fixes or suggestions.
	// N/A Dimension 4 scope boundaries: every listener is node-local.
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnsafeArgumentRule,
		[]rule_tester.ValidTestCase{
			{
				// Upstream skips non-tuple iterables, including Set<any>.
				Code: `
declare function acceptStrings(...values: string[]): void;
declare const values: Set<any>;
acceptStrings(...values);
`,
			},
			{
				// An Iterable<any> constraint is not itself an any-typed spread.
				Code: `
declare function acceptStrings(...values: string[]): void;
function forward<T extends Iterable<any>>(values: T): void {
  acceptStrings(...values);
}
`,
			},
			{
				// An array constraint is not itself an any[] spread.
				Code: `
declare function acceptStrings(...values: string[]): void;
function forward<T extends readonly any[]>(values: T): void {
  acceptStrings(...values);
}
`,
			},
			{
				// Upstream does not compare generic element types of non-tuple array spreads.
				Code: `
declare function acceptSets(...values: Set<string>[]): void;
declare const values: Set<any>[];
acceptSets(...values);
`,
			},
			{
				// Parameters<T> is a conditional type, not a concrete tuple or any[].
				// Checking its iterable yield type creates an extra diagnostic.
				Code: `
declare function acceptStrings(...values: string[]): void;
function forward<T extends (...values: any[]) => void>(...values: Parameters<T>) {
  acceptStrings(...values);
}
`,
			},
			{
				// Mirrors the generic method wrapper from sidebar-operate-model.ts.
				Code: `
class Model {
  add(value: string): void {}
  remove(value: number): void {}
  withLoading<T extends this['add'] | this['remove']>(operateFn: T) {
    return async (...args: Parameters<T>) => {
      await operateFn.call(this, ...args);
    };
  }
}
`,
			},
			{
				// Tuple unions and type parameters are not concrete tuple spreads.
				Code: `
declare function acceptStrings(...values: string[]): void;
declare const values: [any] | [string, any];
acceptStrings(...values);
function forward<T extends [any]>(values: T): void {
  acceptStrings(...values);
}
`,
			},
			{
				// NewExpression shares the same iterable-spread boundary as calls.
				Code: `
declare class Box { constructor(...values: string[]); }
declare const values: Set<any>;
new Box(...values);
function forward<T extends readonly any[]>(values: T): void {
  new Box(...values);
}
`,
			},
			{
				// ---- Real-user: typescript-eslint#10415 generic-constraint false negative ----
				// This is a known upstream false negative; parity means not reporting it.
				Code: `
declare const constrained: <T extends number>(value: T) => void;
constrained(1 as any);
`,
			},
			{
				// ---- Real-user: typescript-eslint#11801 nested object any enhancement ----
				// Upstream currently checks generic type arguments, not object properties.
				Code: `
declare const fromLibrary: { value: any };
declare function acceptObject(value: { value: number }): void;
acceptObject(fromLibrary);
`,
			},
			{
				// ---- Real-user: typescript-eslint#5014 recursive types must not recurse forever ----
				Code: `
type RecursiveArray = Array<RecursiveArray>;
declare const values: RecursiveArray;
declare function identity<T>(value: T): T;
identity(values);
`,
			},
			{
				// A safe non-tuple iterable is ignored, like any other iterable.
				Code: `
declare function acceptStrings(...values: string[]): void;
declare const values: Set<string>;
acceptStrings(...values);
`,
			},
			{
				// Generic iterable spreads are ignored without resolving their constraints.
				Code: `
declare function acceptStrings(...values: string[]): void;
function forward<T extends Iterable<string>>(values: T): void {
  acceptStrings(...values);
}
forward(new Set<string>());
`,
			},
			{
				// Ignoring an iterable does not depend on the receiver type.
				Code: `
declare function acceptAnything(...values: any[]): void;
declare const values: Set<any>;
acceptAnything(...values);
`,
			},
			{
				// String spreads are ignored along with other non-tuple iterables.
				Code: `
declare function acceptStrings(...values: string[]): void;
acceptStrings(...'safe');
`,
			},
			{
				// Locks in FunctionSignature's generic rest-type branch.
				Code: `
declare function genericRest<T extends string[]>(...values: T): void;
genericRest('safe', 1 as any);
`,
			},
			{
				// ---- Dimension 4: parenthesized new Map() keeps upstream's safe special case ----
				Code: `
declare function acceptsMap(value: Map<string, string>): void;
acceptsMap((new Map()));
acceptsMap(((new Map())));
`,
			},
			{
				// ---- Dimension 4: empty arguments degrade gracefully ----
				Code: `
declare function optional(value?: number): void;
optional();
`,
			},
			{
				// Lazy signature fast path: safe literals, safe TypeReferences,
				// fewer args than parameters, empty tuple rest, and any-typed callees.
				Code: `
declare function manyParams(
  a: number,
  b?: string,
  c?: boolean,
  d?: null,
  ...rest: Set<string>[]
): void;
manyParams(1);
manyParams(1, 'ok', true, null, new Set<string>());
manyParams(1, ` + "`plain`" + `, false, null);

declare function emptyTupleRest(...args: []): void;
emptyTupleRest(1 as any);

declare function fixedOnly(a: string, b: number): void;
declare const varTuple: [string, ...string[]];
fixedOnly(...varTuple, 1 as any);

declare const anyFn: any;
anyFn(1, 1 as any, new Set<any>(), ...(1 as any), ...([] as any[]), ...([1 as any] as const));

declare const logger: any;
logger({ a: 1, b: 'ok' });
logger('msg', { a: 1 });
logger(...[{ a: 1 }]);
`,
			},
			{
				// Deferred recursive conditional types passed to `unknown` or an identical
				// receiver must not eagerly expand type arguments.
				Code: `
interface Link {
  value: Link;
}
type Value<T> = T extends { value: infer V } ? Value<V> : T;
type Box<T> = [Value<T>];
interface Envelope<T> {
  payload: T;
}
declare const box: Box<Link>;
declare const packet: Envelope<Box<Link>>;
declare function collect(x: unknown): void;
declare function takeBox(x: Box<Link>): void;

collect(box);
collect(packet);
takeBox(box);
`,
			},
			{
				// Expansive deferred recursive generics must stay bounded without overflowing
				// recursion/instantiation limits across unknown, any[], any-callee, and matching params.
				Code: `
type Inf<T> = [Inf<[T]>];
type Tree<T> = [Tree<[T]>, Tree<[T]>];
declare const x: Inf<string>;
declare const tree: Tree<string>;
declare function sink(v: unknown): void;
declare function acceptAnyRest(...args: any[]): void;
declare const anyFn: any;
declare function takeInf(x: Inf<string>): void;
declare function takeTree(x: Tree<string>): void;

sink(x);
sink(tree);
console.log(x, tree);
acceptAnyRest(x, tree);
anyFn(x);
anyFn(tree);
takeInf(x);
takeTree(tree);
`,
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				// Skipping an iterable must not consume the parameter for a later argument.
				Code: `
declare function acceptValues(first: string, second: number): void;
declare const values: Set<any>;
acceptValues(...values, 1 as any);
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeArgument",
					Message:   "Unsafe argument of type `any` assigned to a parameter of type `string`.",
					Line:      4,
					Column:    25,
					EndLine:   4,
					EndColumn: 33,
				}},
			},
			{
				// Only concrete tuples advance the cursor when an iterable is between them.
				Code: `
declare function acceptValues(first: number, second: string, third: boolean): void;
declare const values: Set<any>;
acceptValues(...[1] as [number], ...values, ...[1 as any] as [any]);
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeTupleSpread",
					Message:   "Unsafe spread of a tuple type. The argument is of type `any` and is assigned to a parameter of type `string`.",
					Line:      4,
					Column:    45,
					EndLine:   4,
					EndColumn: 67,
				}},
			},
			{
				// A concrete readonly any[] must still take the unsafe-array branch.
				Code: `
declare function acceptStrings(...values: string[]): void;
declare const values: readonly any[];
acceptStrings(...values);
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeArraySpread",
					Message:   "Unsafe spread of an `readonly any[]` array type.",
					Line:      4,
					Column:    15,
					EndLine:   4,
					EndColumn: 24,
				}},
			},
			{
				// Locks in upstream FunctionSignature.create(): classify a generic rest parameter through its array constraint.
				Code: `
declare function acceptStrings<T extends string[]>(...values: T): void;
function forward<T extends string[]>(value: any): void {
  acceptStrings<T>(value);
}
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeArgument",
					Message:   "Unsafe argument of type `any` assigned to a parameter of type `string`.",
					Line:      4,
					Column:    20,
					EndLine:   4,
					EndColumn: 25,
				}},
			},
			{
				// A non-tuple spread does not consume the next fixed parameter for later arguments.
				Code: `
declare function acceptValues(first: string, second: number): void;
declare const values: string[];
acceptValues(...values, 1 as any);
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeArgument",
					Message:   "Unsafe argument of type `any` assigned to a parameter of type `string`.",
					Line:      4,
					Column:    25,
					EndLine:   4,
					EndColumn: 33,
				}},
			},
			{
				// A non-tuple spread preserves the upstream receiver for a later argument even when a rest parameter exists.
				Code: `
declare function acceptValues(first: string, ...rest: number[]): void;
declare const values: string[];
acceptValues(...values, 1 as any);
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeArgument",
					Message:   "Unsafe argument of type `any` assigned to a parameter of type `string`.",
					Line:      4,
					Column:    25,
					EndLine:   4,
					EndColumn: 33,
				}},
			},
			{
				// ---- Dimension 4: one parenthesized argument; ESTree excludes parens ----
				Code: `
declare function acceptNumber(value: number): void;
declare const anyValue: any;
acceptNumber((anyValue));
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeArgument",
					Message:   "Unsafe argument of type `any` assigned to a parameter of type `number`.",
					Line:      4,
					Column:    15,
					EndLine:   4,
					EndColumn: 23,
				}},
			},
			{
				// ---- Dimension 4: multiply parenthesized argument ----
				Code: `
declare function acceptNumber(value: number): void;
declare const anyValue: any;
acceptNumber(((anyValue)));
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeArgument",
					Line:      4,
					Column:    16,
					EndLine:   4,
					EndColumn: 24,
				}},
			},
			{
				// ---- Dimension 4: comment trivia inside parentheses ----
				Code: `
declare function acceptNumber(value: number): void;
declare const anyValue: any;
acceptNumber((/* leading */ anyValue));
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeArgument",
					Line:      4,
					Column:    29,
					EndLine:   4,
					EndColumn: 37,
				}},
			},
			{
				// ---- Dimension 4: TS non-null assertion wrapper ----
				Code: `
declare function acceptNumber(value: number): void;
declare const anyValue: any;
acceptNumber(anyValue!);
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeArgument",
					Line:      4,
					Column:    14,
					EndLine:   4,
					EndColumn: 23,
				}},
			},
			{
				// ---- Dimension 4: TS as-expression wrapper ----
				Code: `
declare function acceptNumber(value: number): void;
declare const anyValue: any;
acceptNumber(anyValue as any);
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeArgument",
					Line:      4,
					Column:    14,
					EndLine:   4,
					EndColumn: 29,
				}},
			},
			{
				// ---- Dimension 4: TS satisfies-expression wrapper ----
				Code: `
declare function acceptNumber(value: number): void;
declare const anyValue: any;
acceptNumber(anyValue satisfies unknown);
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeArgument",
					Line:      4,
					Column:    14,
					EndLine:   4,
					EndColumn: 40,
				}},
			},
			{
				// ---- Dimension 4: optional call ----
				Code: `
declare const anyValue: any;
declare const maybeFunction: ((value: number) => void) | undefined;
maybeFunction?.(anyValue);
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeArgument",
					Line:      4,
					Column:    17,
					EndLine:   4,
					EndColumn: 25,
				}},
			},
			{
				// Locks in the NewExpression listener's ordinary argument branch.
				Code: `
declare const anyValue: any;
declare class Box { constructor(value: number); }
new Box(anyValue);
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeArgument",
					Line:      4,
					Column:    9,
					EndLine:   4,
					EndColumn: 17,
				}},
			},
			{
				// Locks in NewExpression plus tuple-spread checking.
				Code: `
declare const anyValue: any;
declare class Box { constructor(value: number); }
new Box(...([anyValue] as const));
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeTupleSpread",
					Message:   "Unsafe spread of a tuple type. The argument is of type `any` and is assigned to a parameter of type `number`.",
					Line:      4,
					Column:    9,
					EndLine:   4,
					EndColumn: 33,
				}},
			},
			{
				// Locks in FunctionSignature.getNextParameterType()'s tuple fallback.
				Code: `
declare function tupleRest(...values: [number, string]): void;
declare const anyValue: any;
tupleRest(...([1, 'ok', anyValue] as [number, string, any]));
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeTupleSpread",
					Message:   "Unsafe spread of a tuple type. The argument is of type `any` and is assigned to a parameter of type `string`.",
					Line:      4,
					Column:    11,
					EndLine:   4,
					EndColumn: 60,
				}},
			},
			{
				// Safe literal/TypeReference prefix must preserve parameter cursor for later unsafe args,
				// including deeply nested and recursive TypeReferences containing `any`.
				Code: `
type RecWithAny = [RecWithAny, any];
declare const recWithAny: RecWithAny;
declare function takeLater(
  a: number,
  b: string,
  c: Set<string>,
  d: Set<Set<string>>,
  e: [RecWithAny, string],
): void;
takeLater(1, 'ok', new Set<string>(), new Set([new Set<any>()]), recWithAny);
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unsafeArgument",
						Message:   "Unsafe argument of type `Set<Set<any>>` assigned to a parameter of type `Set<Set<string>>`.",
						Line:      11,
						Column:    39,
						EndLine:   11,
						EndColumn: 64,
					},
					{
						MessageId: "unsafeArgument",
						Message:   "Unsafe argument of type `RecWithAny` assigned to a parameter of type `[RecWithAny, string]`.",
						Line:      11,
						Column:    66,
						EndLine:   11,
						EndColumn: 76,
					},
				},
			},
			{
				// Explicit `this` parameter is excluded from Signature_parameters and must not shift lazy cursor.
				Code: `
declare const obj: {
  m(this: { id: number }, a: number, b: string): void;
};
obj.m(1, 2 as any);
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeArgument",
					Message:   "Unsafe argument of type `any` assigned to a parameter of type `string`.",
					Line:      5,
					Column:    10,
					EndLine:   5,
					EndColumn: 18,
				}},
			},
			{
				// Multiple unsafe rest arguments reuse lazily resolved rest parameter type.
				Code: `
declare function withRest(head: number, ...tail: string[]): void;
withRest(1, 'ok', 2 as any, 'still-ok', 3 as any);
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unsafeArgument",
						Message:   "Unsafe argument of type `any` assigned to a parameter of type `string`.",
						Line:      3,
						Column:    19,
						EndLine:   3,
						EndColumn: 27,
					},
					{
						MessageId: "unsafeArgument",
						Message:   "Unsafe argument of type `any` assigned to a parameter of type `string`.",
						Line:      3,
						Column:    41,
						EndLine:   3,
						EndColumn: 49,
					},
				},
			},
			{
				// Tagged template expression with safe literal span before an unsafe span.
				Code: `
declare function tag(strings: TemplateStringsArray, a: number, b: string): void;
tag` + "`" + `${1}${2 as any}` + "`" + `;
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeArgument",
					Message:   "Unsafe argument of type `any` assigned to a parameter of type `string`.",
					Line:      3,
					Column:    11,
					EndLine:   3,
					EndColumn: 19,
				}},
			},
		},
	)
}

func TestNoUnsafeArgumentDeferredRecursiveTypeDoesNotPoisonTypeCheck(t *testing.T) {
	t.Parallel()

	type tsDiagSummary struct {
		code int32
		line int
		col  int
		text string
	}

	summarizeASTDiags := func(t *testing.T, sourceFile *ast.SourceFile, diags []*ast.Diagnostic) []tsDiagSummary {
		t.Helper()
		text := sourceFile.Text()
		out := make([]tsDiagSummary, 0, len(diags))
		for _, d := range diags {
			if d.File() == nil {
				t.Errorf("unexpected file-less diagnostic: TS%d %s", d.Code(), d.String())
				continue
			}
			line, col := scanner.GetECMALineAndUTF16CharacterOfPosition(d.File(), d.Pos())
			out = append(out, tsDiagSummary{
				code: d.Code(),
				line: line + 1,
				col:  int(col) + 1,
				text: text[d.Pos() : d.Pos()+d.Len()],
			})
		}
		return out
	}

	summarizeRuleTSDiags := func(t *testing.T, diags []rule.RuleDiagnostic, allowRuleDiag bool) []tsDiagSummary {
		t.Helper()
		out := make([]tsDiagSummary, 0, len(diags))
		for _, d := range diags {
			if d.RuleName == NoUnsafeArgumentRule.Name {
				if !allowRuleDiag {
					t.Errorf("unexpected %s diagnostic: %#v", NoUnsafeArgumentRule.Name, d)
				}
				continue
			}
			if !strings.HasPrefix(d.RuleName, "TypeScript(TS") || !strings.HasSuffix(d.RuleName, ")") {
				t.Errorf("unexpected non-TypeScript diagnostic: %#v", d)
				continue
			}
			codeStr := strings.TrimSuffix(strings.TrimPrefix(d.RuleName, "TypeScript(TS"), ")")
			codeNum, err := strconv.Atoi(codeStr)
			if err != nil {
				t.Fatalf("parse TS code from %q: %v", d.RuleName, err)
			}
			if d.SourceFile == nil {
				t.Errorf("TypeScript diagnostic missing source file: %#v", d)
				continue
			}
			line, col := scanner.GetECMALineAndUTF16CharacterOfPosition(d.SourceFile, d.Range.Pos())
			out = append(out, tsDiagSummary{
				code: int32(codeNum),
				line: line + 1,
				col:  int(col) + 1,
				text: d.SourceFile.Text()[d.Range.Pos():d.Range.End()],
			})
		}
		return out
	}

	cases := []struct {
		name              string
		body              string
		allowRuleDiag     bool
		want              []tsDiagSummary
		wantTypeCheckOnly []tsDiagSummary
	}{
		{
			name: "deferred recursive tuple and wrapper passed to unknown",
			body: `
declare const box: Box<Link>;
declare const packet: Envelope<Box<Link>>;
declare function collect(x: unknown): void;

collect(box);
collect(packet);
`,
			want: []tsDiagSummary{
				{code: 2322, line: 19, col: 7, text: "mustBeNumber"},
			},
		},
		{
			name: "recursive cast passed to unknown resolves signature before argument type",
			body: `
declare function marker(x: unknown): void;
marker(null as unknown as Value<Link>);
`,
			want: []tsDiagSummary{
				{code: 2589, line: 13, col: 27, text: "Value<Link>"},
				{code: 2322, line: 15, col: 7, text: "mustBeNumber"},
			},
		},
		{
			name: "recursive cast passed to concrete parameter resolves signature before argument type",
			body: `
declare function takeNum(x: number): void;
takeNum(null as unknown as Value<Link>);
`,
			allowRuleDiag: true,
			want: []tsDiagSummary{
				{code: 2589, line: 13, col: 28, text: "Value<Link>"},
				{code: 2322, line: 15, col: 7, text: "mustBeNumber"},
			},
		},
		{
			name: "recursive cast in spread argument resolves signature before spread type",
			body: `
declare function spreadInto(...xs: unknown[]): void;
spreadInto(...(null as unknown as [Value<Link>]));
`,
			want: []tsDiagSummary{
				{code: 2589, line: 13, col: 36, text: "Value<Link>"},
				{code: 2322, line: 15, col: 7, text: "mustBeNumber"},
			},
		},
		{
			name: "nested generic call with literal arg followed by recursive callee call",
			body: `
declare function register<T>(name: string): void;
declare function flush(): void;
flush(register<Value<Link>>("listener"));
declare function getCallback(): Value<Link>;
getCallback()(undefined);
`,
			// Visiting the nested register<Value<Link>>("listener") call resolves its
			// explicit type argument at line 14 (matching main), instead of letting
			// getCallback()(undefined) record a file-less global TS2589.
			want: []tsDiagSummary{
				{code: 2554, line: 14, col: 7, text: `register<Value<Link>>("listener")`},
				{code: 2589, line: 14, col: 16, text: "Value<Link>"},
				{code: 2322, line: 18, col: 7, text: "mustBeNumber"},
			},
			wantTypeCheckOnly: []tsDiagSummary{
				{code: 2554, line: 14, col: 7, text: `register<Value<Link>>("listener")`},
				{code: 2589, line: 15, col: 33, text: "Value<Link>"},
				{code: 2322, line: 18, col: 7, text: "mustBeNumber"},
			},
		},
		{
			name: "direct generic call with literal arg followed by recursive callee call",
			body: `
declare function register<T>(name: string): void;
register<Value<Link>>("listener");
declare function getCallback(): Value<Link>;
getCallback()(undefined);
`,
			want: []tsDiagSummary{
				{code: 2589, line: 13, col: 10, text: "Value<Link>"},
				{code: 2322, line: 17, col: 7, text: "mustBeNumber"},
			},
		},
		{
			name: "ts-expect-error on explicit type argument call is preserved before later callback lookup",
			body: `
declare function register<T>(name: string): void;
// @ts-expect-error
register<Value<Link>>("listener");
declare function getCallback(): Value<Link>;
getCallback()(undefined);
`,
			want: []tsDiagSummary{
				{code: 2322, line: 18, col: 7, text: "mustBeNumber"},
			},
		},
		{
			name: "new expression with explicit recursive type argument and literal arg",
			body: `
declare class Reg<T> { constructor(name: string); }
new Reg<Value<Link>>("listener");
declare function getCallback(): Value<Link>;
getCallback()(undefined);
`,
			want: []tsDiagSummary{
				{code: 2589, line: 13, col: 9, text: "Value<Link>"},
				{code: 2322, line: 17, col: 7, text: "mustBeNumber"},
			},
		},
		{
			name: "chained generic callee with literal arg before explicit-type-arg recursive callee",
			body: `
declare function api<T>(): { on(name: string): void };
api<Value<Link>>().on("ready");
declare function getCallback(): Value<Link>;
getCallback()<string>("event");
`,
			want: []tsDiagSummary{
				{code: 2589, line: 13, col: 5, text: "Value<Link>"},
				{code: 2322, line: 17, col: 7, text: "mustBeNumber"},
			},
		},
		{
			name: "assertion callee with literal arg preserves ts-expect-error before later explicit-type-arg call",
			body: `
declare function make<T>(): unknown;
// @ts-expect-error
(make<Value<Link>>() as (name: string) => void)("ready");
declare function register<T>(name: string): void;
register<Value<Link>>("later");
`,
			want: []tsDiagSummary{
				{code: 2322, line: 18, col: 7, text: "mustBeNumber"},
			},
		},
		{
			name: "assertion callee with literal arg before explicit-type-arg recursive callee does not poison type-check",
			body: `
declare function make<T>(): unknown;
(make<Value<Link>>() as (name: string) => void)("ready");
declare function getCallback(): Value<Link>;
getCallback()<string>("event");
`,
			want: []tsDiagSummary{
				{code: 2589, line: 13, col: 7, text: "Value<Link>"},
				{code: 2322, line: 17, col: 7, text: "mustBeNumber"},
			},
		},
		{
			name: "await-wrapped assertion callee with literal arg preserves ts-expect-error",
			body: `
declare function make<T>(): unknown;
async function run() {
  // @ts-expect-error
  (await (make<Value<Link>>() as Promise<(name: string) => void>))("ready");
}
declare function register<T>(name: string): void;
register<Value<Link>>("later");
`,
			want: []tsDiagSummary{
				{code: 2322, line: 20, col: 7, text: "mustBeNumber"},
			},
		},
		{
			name: "non-generic call callee with explicit type arguments and literal arg preserves ts-expect-error",
			body: `
declare function make(): (name: string) => void;
// @ts-expect-error
make<Value<Link>>()("ready");
declare function register<T>(name: string): void;
register<Value<Link>>("later");
`,
			want: []tsDiagSummary{
				{code: 2322, line: 18, col: 7, text: "mustBeNumber"},
			},
		},
		{
			name: "tagged template with explicit recursive type argument and literal span",
			body: `
declare function tag<T>(s: TemplateStringsArray, x: string): void;
tag<Value<Link>>` + "`${\"listener\"}`" + `;
declare function getCallback(): Value<Link>;
getCallback()(undefined);
`,
			want: []tsDiagSummary{
				{code: 2589, line: 13, col: 5, text: "Value<Link>"},
				{code: 2322, line: 17, col: 7, text: "mustBeNumber"},
			},
		},
		{
			name: "failed earlier argument before unknown-typed recursive payload does not poison type-check",
			body: `
declare function make<T>(): T;
declare function send(event: 'ready', payload: unknown): void;
send('typo', make<Value<Link>>());
declare function getCallback(): Value<Link>;
getCallback()("event");
`,
			want: []tsDiagSummary{
				{code: 2345, line: 14, col: 6, text: "'typo'"},
				{code: 2589, line: 14, col: 19, text: "Value<Link>"},
				{code: 2322, line: 18, col: 7, text: "mustBeNumber"},
			},
			wantTypeCheckOnly: []tsDiagSummary{
				{code: 2345, line: 14, col: 6, text: "'typo'"},
				{code: 2589, line: 15, col: 33, text: "Value<Link>"},
				{code: 2322, line: 18, col: 7, text: "mustBeNumber"},
			},
		},
		{
			name: "extra trailing argument after unknown-typed recursive payload does not poison type-check",
			body: `
declare function make<T>(): T;
declare function send(event: 'ready', payload: unknown): void;
send('ready', make<Value<Link>>(), 'extra');
declare function getCallback(): Value<Link>;
getCallback()("event");
`,
			want: []tsDiagSummary{
				{code: 2589, line: 14, col: 20, text: "Value<Link>"},
				{code: 2554, line: 14, col: 36, text: "'extra'"},
				{code: 2322, line: 18, col: 7, text: "mustBeNumber"},
			},
			wantTypeCheckOnly: []tsDiagSummary{
				{code: 2554, line: 14, col: 36, text: "'extra'"},
				{code: 2589, line: 15, col: 33, text: "Value<Link>"},
				{code: 2322, line: 18, col: 7, text: "mustBeNumber"},
			},
		},
		{
			name: "missing trailing argument after unknown-typed recursive payload does not poison type-check",
			body: `
declare function make<T>(): T;
declare function send3(event: 'ready', payload: unknown, tag: string): void;
send3('ready', make<Value<Link>>());
declare function getCallback(): Value<Link>;
getCallback()("event");
`,
			want: []tsDiagSummary{
				{code: 2554, line: 14, col: 1, text: "send3"},
				{code: 2589, line: 14, col: 21, text: "Value<Link>"},
				{code: 2322, line: 18, col: 7, text: "mustBeNumber"},
			},
			wantTypeCheckOnly: []tsDiagSummary{
				{code: 2554, line: 14, col: 1, text: "send3"},
				{code: 2589, line: 15, col: 33, text: "Value<Link>"},
				{code: 2322, line: 18, col: 7, text: "mustBeNumber"},
			},
		},
		{
			name: "excess quick-type recursive cast does not emit file-less global TS2589",
			body: `
declare function sendOne(event: string): void;
sendOne('ready', null as unknown as Value<Link>);
`,
			want: []tsDiagSummary{
				{code: 2554, line: 13, col: 18, text: "null as unknown as Value<Link>"},
				{code: 2589, line: 13, col: 37, text: "Value<Link>"},
				{code: 2322, line: 15, col: 7, text: "mustBeNumber"},
			},
		},
		{
			name: "ts-expect-error on any-asserted call does not suppress declaration TS2589",
			body: `
declare function getPayload(): Value<Link>;
declare const emit: unknown;
// @ts-expect-error
(emit as any)(getPayload(), missingName);
`,
			want: []tsDiagSummary{
				{code: 2589, line: 12, col: 32, text: "Value<Link>"},
				{code: 2322, line: 17, col: 7, text: "mustBeNumber"},
			},
		},
		{
			name: "ts-expect-error on any-returning factory call does not suppress declaration TS2589",
			body: `
declare function getPayload(): Value<Link>;
declare function getEmit(): any;
// @ts-expect-error
getEmit()(getPayload(), missingName);
`,
			want: []tsDiagSummary{
				{code: 2589, line: 12, col: 32, text: "Value<Link>"},
				{code: 2322, line: 17, col: 7, text: "mustBeNumber"},
			},
		},
		{
			name: "ts-expect-error on declaration stays active when later called inside any-asserted call",
			body: `
// @ts-expect-error
declare function getPayload(): Value<Link>;
declare const emit: unknown;
(emit as any)(getPayload());
`,
			want: []tsDiagSummary{
				{code: 2322, line: 17, col: 7, text: "mustBeNumber"},
			},
		},
		{
			name: "non-generic new-expression callee with explicit type arguments and literal arg preserves ts-expect-error",
			body: `
interface Factory {
  new (): (name: string) => void;
}
declare const Factory: Factory;
// @ts-expect-error
(new Factory<Value<Link>>())("ready");
declare function register<T>(name: string): void;
register<Value<Link>>("later");
`,
			want: []tsDiagSummary{
				{code: 2322, line: 21, col: 7, text: "mustBeNumber"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			const fileName = "deferred_recursive_typecheck.ts"
			code := `
export {};
interface Link {
  value: Link;
}
type Value<T> = T extends { value: infer V } ? Value<V> : T;
type Box<T> = [Value<T>];
interface Envelope<T> {
  payload: T;
}
` + tc.body + `
const mustBeNumber: number = "bad";
`

			root := fixtures.GetRootDir()
			entryPath := tspath.ResolvePath(root.Dir, fileName)
			createProgram := func() (*compiler.Program, *lintprogram.Program, *ast.SourceFile) {
				fs := utils.NewOverlayVFS(root.FS, map[string]string{entryPath: code})
				host := utils.CreateCompilerHost(root.Dir, fs)
				program, err := utils.CreateProgramFromOptionsLenient(
					false,
					&core.CompilerOptions{
						Target: core.ScriptTargetES2020,
						Strict: core.TSTrue,
					},
					[]string{entryPath},
					host,
				)
				if err != nil {
					t.Fatalf("create program: %v", err)
				}
				sourceFile := program.GetSourceFile(entryPath)
				if sourceFile == nil {
					t.Fatalf("program did not contain %s", fileName)
				}
				return program, lintprogram.NewFromCompiler(program), sourceFile
			}

			ctx := context.Background()

			// Baseline: --type-check-only without running any lint rule.
			wantBaseline := tc.wantTypeCheckOnly
			if wantBaseline == nil {
				wantBaseline = tc.want
			}
			baselineCompilerProg, _, baselineFile := createProgram()
			if globals := baselineCompilerProg.GetGlobalDiagnostics(ctx); len(globals) != 0 {
				t.Fatalf("baseline --type-check-only produced global diagnostics: %v", globals)
			}
			baselineDiags := summarizeASTDiags(t, baselineFile, baselineCompilerProg.GetSemanticDiagnostics(ctx, baselineFile))
			if !slices.Equal(baselineDiags, wantBaseline) {
				t.Fatalf("baseline --type-check-only diagnostics = %+v, want %+v", baselineDiags, wantBaseline)
			}

			program, lintProg, sourceFile := createProgram()
			lintPlan, err := linter.PrepareLintPlan(linter.PrepareLintPlanOptions{
				Programs:         []*lintprogram.Program{lintProg},
				TargetsByProgram: [][]string{{sourceFile.FileName()}},
				SingleThreaded:   true,
				GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
					return []rule.ConfiguredRule{{
						Name:             NoUnsafeArgumentRule.Name,
						Severity:         rule.SeverityError,
						RequiresTypeInfo: true,
						Run: func(ctx rule.RuleContext) rule.RuleListeners {
							return NoUnsafeArgumentRule.Run(ctx, nil)
						},
					}}
				},
			})
			if err != nil {
				t.Fatalf("PrepareLintPlan: %v", err)
			}

			var reported []rule.RuleDiagnostic
			_, err = linter.RunLinter(linter.RunLinterOptions{
				SingleThreaded: true,
				TypeCheck:      true,
				LintPlan:       lintPlan,
				Consumer: rule.DiagnosticConsumer{
					Demand: rule.EditDemandNone,
					Report: func(diagnostic rule.RuleDiagnostic) {
						reported = append(reported, diagnostic)
					},
				},
			})
			if err != nil {
				t.Fatalf("RunLinter: %v", err)
			}

			if globals := program.GetGlobalDiagnostics(ctx); len(globals) != 0 {
				var msgs []string
				for _, g := range globals {
					msgs = append(msgs, g.String())
				}
				t.Fatalf("expected no global diagnostics (no file-less TS2589), got %v", msgs)
			}

			gotReported := summarizeRuleTSDiags(t, reported, tc.allowRuleDiag)
			if !slices.Equal(gotReported, tc.want) {
				t.Fatalf("RunLinter --type-check diagnostics = %+v, want %+v", gotReported, tc.want)
			}

			gotNoEmit := summarizeASTDiags(t, sourceFile, lintProg.NoEmitDiagnostics(ctx))
			if !slices.Equal(gotNoEmit, tc.want) {
				t.Fatalf("NoEmitDiagnostics = %+v, want %+v", gotNoEmit, tc.want)
			}

			gotSemantic := summarizeASTDiags(t, sourceFile, program.GetSemanticDiagnostics(ctx, sourceFile))
			if !slices.Equal(gotSemantic, tc.want) {
				t.Fatalf("GetSemanticDiagnostics = %+v, want %+v", gotSemantic, tc.want)
			}
		})
	}
}
