// TestNoUnsafeArgumentExtras locks in branches and edge shapes that the
// upstream suite in no_unsafe_argument_upstream_test.go does not exercise.
// Each case identifies the upstream branch, universal edge shape, or real-user
// regression it protects.
package no_unsafe_argument

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
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
		nil,
		[]rule_tester.InvalidTestCase{
			{
				Code: `
declare function acceptStrings(...values: string[]): void;

function forward<T extends readonly any[]>(values: T): void {
  acceptStrings(...values);
}
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeArgument",
					Message:   "Unsafe argument of type `any` assigned to a parameter of type `string`.",
					Line:      5,
					Column:    17,
					EndLine:   5,
					EndColumn: 26,
				}},
			},
			{
				// The ES5 fallback must retrieve an index type from each array
				// union rather than requiring the whole type to be an array.
				Code: `
declare function acceptStrings(...values: string[]): void;
declare const values: string[] | any[];

acceptStrings(...values);
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeArgument",
					Message:   "Unsafe argument of type `any` assigned to a parameter of type `string`.",
					Line:      5,
					Column:    15,
					EndLine:   5,
					EndColumn: 24,
				}},
			},
			{
				// Generic array-union constraints take the same fallback after the
				// checker resolves their base constraint.
				Code: `
declare function acceptStrings(...values: string[]): void;

function forward<T extends string[] | any[]>(values: T): void {
  acceptStrings(...values);
}
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeArgument",
					Message:   "Unsafe argument of type `any` assigned to a parameter of type `string`.",
					Line:      5,
					Column:    17,
					EndLine:   5,
					EndColumn: 26,
				}},
			},
		},
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
				// A safe iterable element type flows to the rest element type.
				Code: `
declare function acceptStrings(...values: string[]): void;
declare const values: Set<string>;
acceptStrings(...values);
`,
			},
			{
				// The generic constraint supplies the iterable element type.
				Code: `
declare function acceptStrings(...values: string[]): void;
function forward<T extends Iterable<string>>(values: T): void {
  acceptStrings(...values);
}
forward(new Set<string>());
`,
			},
			{
				// Any is safe when the iterable feeds a rest parameter that accepts any.
				Code: `
declare function acceptAnything(...values: any[]): void;
declare const values: Set<any>;
acceptAnything(...values);
`,
			},
			{
				// Strings use the same iterable yield path without introducing any.
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
		},
		[]rule_tester.InvalidTestCase{
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
				// A non-array iterable still spreads its element type into the rest parameter.
				Code: `
declare function acceptStrings(...values: string[]): void;
declare const values: Set<any>;
acceptStrings(...values);
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeArgument",
					Message:   "Unsafe argument of type `any` assigned to a parameter of type `string`.",
					Line:      4,
					Column:    15,
					EndLine:   4,
					EndColumn: 24,
				}},
			},
			{
				// A type parameter gets its iteration type through the generic constraint.
				Code: `
declare function acceptStrings(...values: string[]): void;
function forward<T extends Iterable<any>>(values: T): void {
  acceptStrings(...values);
}
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeArgument",
					Line:      4,
					Column:    17,
					EndLine:   4,
					EndColumn: 26,
				}},
			},
			{
				// Array-like generic constraints reach the same iterable path because T is not itself an array.
				Code: `
declare function acceptStrings(...values: string[]): void;
function forward<T extends readonly any[]>(values: T): void {
  acceptStrings(...values);
}
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeArgument",
					Line:      4,
					Column:    17,
					EndLine:   4,
					EndColumn: 26,
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
				// Ordinary array spreads participate in rslint's iterable element-type check.
				Code: `
declare function acceptSets(...values: Set<string>[]): void;
declare const values: Set<any>[];
acceptSets(...values);
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeArgument",
					Message:   "Unsafe argument of type `Set<any>` assigned to a parameter of type `Set<string>`.",
					Line:      4,
					Column:    12,
					EndLine:   4,
					EndColumn: 21,
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
		},
	)
}
