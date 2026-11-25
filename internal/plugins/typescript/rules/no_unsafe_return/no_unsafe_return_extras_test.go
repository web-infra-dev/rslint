package no_unsafe_return

import (
	"context"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestNoUnsafeReturnFastPathBoundaries(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnsafeReturnRule, []rule_tester.ValidTestCase{
		{
			// Safe non-TypeReference values (objects, unions, functions, non-generic instances)
			// and safe TypeReferences (arrays, collections, nested/recursive generics) in sync/async functions.
			Code: `
interface Plain { value: number }
type RecursiveArray = Array<RecursiveArray>;
type BoxA = [BoxB];
type BoxB = [BoxA];
type SharedCycle = [CycleA, CycleB];
type CycleA = [CycleB];
type CycleB = [CycleC];
type CycleC = [CycleB];
declare const plain: Plain;
declare const unionVal: Set<any> | null;
declare const recArr: RecursiveArray;
declare const boxA: BoxA;
declare const sharedCycle: SharedCycle;

function syncLiterals() {
  if (Math.random() > 0.5) return 1;
  if (Math.random() > 0.5) return (1);
  if (Math.random() > 0.5) return 'ok';
  if (Math.random() > 0.5) return ` + "`plain`" + `;
  if (Math.random() > 0.5) return true;
  if (Math.random() > 0.5) return false;
  return null;
}

const arrowObj = (): { a: number } => ({ a: 1 });
const arrowPlain = (): Plain => plain;
const arrowUnion = (): Set<string> | null => unionVal;
const arrowFn = () => (x: number) => x + 1;
const arrowSafeGeneric = (): Set<Set<string>> => new Set([new Set(['a'])]);
const arrowRec = (): RecursiveArray => recArr;
const arrowMutualRec = (): BoxA => boxA;
const arrowSharedCycle = (): SharedCycle => sharedCycle;

async function asyncSafe() {
  if (Math.random() > 0.5) return 1;
  if (Math.random() > 0.5) return { a: 1 };
  if (Math.random() > 0.5) return new Set<string>();
  return recArr;
}
`,
		},
		{
			// Non-async functions returning Promise<any> or a thenable extending Promise<any>
			// remain suppressed even when the declared return type is Promise<string>.
			Code: `
declare const promiseAny: Promise<any>;
function syncPromiseAny(): Promise<string> {
  return promiseAny;
}
function syncPromiseAnyInferred() {
  return promiseAny;
}

interface Alias<T> extends Promise<any> {
  tag: T;
}
declare const aliasVal: Alias<number>;
function syncAlias(): Alias<number> {
  return aliasVal;
}

interface CustomThenable<T> extends Promise<T> {
  custom: true;
}
declare const customThenableAny: CustomThenable<any>;
function syncCustomThenable(): CustomThenable<string> {
  return customThenableAny;
}
`,
		},
		{
			// Deferred recursive conditional types returned to `unknown` / `Promise<unknown>`
			// must not eagerly expand type arguments.
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

function getPacket(): unknown {
  return packet;
}
function getBox(): unknown {
  return box;
}
async function getPacketAsync(): Promise<unknown> {
  return packet;
}
async function getBoxAsync(): Promise<unknown> {
  return box;
}
`,
		},
		{
			// Expansive deferred recursive generics must stay bounded without overflowing
			// recursion/instantiation limits.
			Code: `
type Inf<T> = [Inf<[T]>];
declare const x: Inf<string>;
function id(): Inf<string> {
  return x;
}
function idInferred() {
  return x;
}

type Tree<T> = [Tree<[T]>, Tree<[T]>];
declare const t: Tree<string>;
function id2(): Tree<string> {
  return t;
}
function id2Inferred() {
  return t;
}
`,
		},
	}, []rule_tester.InvalidTestCase{
		{
			// Deeply nested and recursive TypeReferences containing `any` must still report.
			Code: `
type RecWithAny = [RecWithAny, any];
declare const recWithAny: RecWithAny;
function returnRec(): [RecWithAny, string] {
  return recWithAny;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeReturnAssignment", Line: 5, Column: 3},
			},
		},
		{
			// Async function returning a non-TypeReference thenable (object literal) yielding `any`.
			Code: `
async function asyncObjThenable() {
  return {
    then(resolve: (value: any) => void) {
      resolve(1);
    },
  };
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeReturn", Line: 3, Column: 3},
			},
		},
		{
			// Async function returning a non-generic interface extending Promise<any>.
			Code: `
interface NonGenericThenable extends Promise<any> {
  tag: true;
}
declare const thenable: NonGenericThenable;
async function asyncNonGenericThenable() {
  return thenable;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeReturn", Line: 7, Column: 3},
			},
		},
		{
			// Async function returning Promise<Set<any>> assigned to Promise<Set<string>>.
			Code: `
async function asyncNestedUnsafeGeneric(): Promise<Set<string>> {
  return Promise.resolve(new Set<any>());
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeReturnAssignment", Line: 3, Column: 3},
			},
		},
		{
			// Literal inside a `with` statement is typed as error (`any`) by the checker
			// and must not be skipped by the literal fast path.
			TSConfig: "tsconfig.unstrict.json",
			Code: `
function withReturn(obj: any) {
  with (obj) {
    return 1;
  }
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeReturn", Line: 4, Column: 5},
			},
		},
		{
			// JSDoc `@type {any}` cast on a parenthesized literal must not be skipped
			// by the parenthesized literal fast path.
			FileName: "file.mjs",
			TSConfig: "tsconfig.allowJs.json",
			Code: `
function jsdocAnyReturn() {
  return /** @type {any} */ (1);
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeReturn", Line: 3, Column: 3},
			},
		},
		{
			// Non-async function returning Set<any> to Set<string> must still report unsafeReturnAssignment.
			Code: `
declare const setAny: Set<any>;
function syncSetAny(): Set<string> {
  return setAny;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeReturnAssignment", Line: 4, Column: 3},
			},
		},
		{
			// Non-async function returning Promise<Set<any>> to Promise<Set<string>> must still report
			// unsafeReturnAssignment (unlike Promise<any>, DiscriminateAnyType returns Safe).
			Code: `
declare const promiseSetAny: Promise<Set<any>>;
function syncPromiseSetAny(): Promise<Set<string>> {
  return promiseSetAny;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeReturnAssignment", Line: 4, Column: 3},
			},
		},
	})
}

func TestNoUnsafeReturnDeferredRecursiveTypeDoesNotPoisonTypeCheck(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                  string
		fn                    string
		allowSemanticTS2589   bool
		requireSemanticTS2589 bool
	}{
		{
			name: "sync packet to unknown",
			fn:   "function getPacket(): unknown { return packet; }",
		},
		{
			name: "sync box to unknown",
			fn:   "function getBox(): unknown { return box; }",
		},
		{
			name: "async packet to Promise<unknown>",
			fn:   "async function getPacketAsync(): Promise<unknown> { return packet; }",
		},
		{
			// TypeScript itself checks Awaited<Box<Link>> on `return box` inside an
			// async function and emits a file-attached semantic TS2589. The rule must
			// still avoid recording a file-less global TS2589 that would silence TS2322.
			name:                  "async box to Promise<unknown>",
			fn:                    "async function getBoxAsync(): Promise<unknown> { return box; }",
			allowSemanticTS2589:   true,
			requireSemanticTS2589: true,
		},
		{
			// Sync return of a Promise<any>-extending generic carrying a deferred
			// recursive type argument must hit the !isAsync PromiseAny exemption
			// before comparing generic type arguments in IsUnsafeAssignment.
			name:                  "sync Job<any, Box<Link>> to Job<unknown, [unknown]>",
			fn:                    "interface Job<R, M> extends Promise<R> { readonly metadata: M }\ndeclare const job: Job<any, Box<Link>>;\nfunction getJob(): Job<unknown, [unknown]> { return job; }",
			allowSemanticTS2589:   true,
			requireSemanticTS2589: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			const fileName = "deferred_recursive_return_typecheck.ts"
			code := `
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

` + tc.fn + `

const mustBeNumber: number = "bad";
`

			root := fixtures.GetRootDir()
			entryPath := tspath.ResolvePath(root.Dir, fileName)
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

			lintProg := lintprogram.NewFromCompiler(program)
			lintPlan, err := linter.PrepareLintPlan(linter.PrepareLintPlanOptions{
				Programs:         []*lintprogram.Program{lintProg},
				TargetsByProgram: [][]string{{sourceFile.FileName()}},
				SingleThreaded:   true,
				GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
					return []rule.ConfiguredRule{{
						Name:             NoUnsafeReturnRule.Name,
						Severity:         rule.SeverityError,
						RequiresTypeInfo: true,
						Run: func(ctx rule.RuleContext) rule.RuleListeners {
							return NoUnsafeReturnRule.Run(ctx, nil)
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

			ctx := context.Background()
			if globals := program.GetGlobalDiagnostics(ctx); len(globals) != 0 {
				var msgs []string
				for _, g := range globals {
					msgs = append(msgs, g.String())
				}
				t.Fatalf("expected no global diagnostics (no file-less TS2589), got %v", msgs)
			}

			hasTS2322 := false
			hasSemanticTS2589 := false
			for _, d := range reported {
				if d.RuleName == NoUnsafeReturnRule.Name {
					t.Errorf("unexpected %s diagnostic: %#v", NoUnsafeReturnRule.Name, d)
				}
				if d.RuleName == "TypeScript(TS2589)" {
					if !tc.allowSemanticTS2589 || d.SourceFile == nil {
						t.Errorf("unexpected TS2589 diagnostic from linter: %#v", d)
					} else {
						hasSemanticTS2589 = true
					}
				}
				if d.RuleName == "TypeScript(TS2322)" {
					hasTS2322 = true
				}
			}
			if !hasTS2322 {
				t.Fatalf("expected TypeScript(TS2322) to be reported under --type-check, got %#v", reported)
			}
			if tc.requireSemanticTS2589 && !hasSemanticTS2589 {
				t.Fatalf("expected file-attached TypeScript(TS2589) under --type-check, got %#v", reported)
			}

			noEmitHasTS2322 := false
			for _, d := range lintProg.NoEmitDiagnostics(ctx) {
				if d.Code() == 2589 && (!tc.allowSemanticTS2589 || d.File() == nil) {
					t.Errorf("unexpected TS2589 in NoEmitDiagnostics: %s", d.String())
				}
				if d.Code() == 2322 {
					noEmitHasTS2322 = true
				}
			}
			if !noEmitHasTS2322 {
				t.Fatalf("expected TS2322 in NoEmitDiagnostics")
			}
			semanticHasTS2322 := false
			for _, d := range program.GetSemanticDiagnostics(ctx, sourceFile) {
				if d.Code() == 2589 && (!tc.allowSemanticTS2589 || d.File() == nil) {
					t.Errorf("unexpected TS2589 in GetSemanticDiagnostics: %s", d.String())
				}
				if d.Code() == 2322 {
					semanticHasTS2322 = true
				}
			}
			if !semanticHasTS2322 {
				t.Fatalf("expected TS2322 in GetSemanticDiagnostics")
			}
		})
	}
}
