package non_nullable_type_assertion_style

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNonNullableTypeAssertionStyleRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NonNullableTypeAssertionStyleRule, []rule_tester.ValidTestCase{
		{Code: `
declare const original: number | string;
const cast = original as string;
    `},
		{Code: `
declare const original: number | undefined;
const cast = original as string | number | undefined;
    `},
		{Code: `
declare const original: number | any;
const cast = original as string | number | undefined;
    `},
		{Code: `
declare const original: number | undefined;
const cast = original as any;
    `},
		{Code: `
declare const original: string | undefined;
const cast = original as number;
    `},
		{Code: `
declare const original:
  | 'a'
  | 'b'
  | 'c'
  | 'd'
  | 'e'
  | 'f'
  | 'g'
  | 'h'
  | 'i'
  | undefined;
const cast = original as
  | 'a'
  | 'b'
  | 'c'
  | 'd'
  | 'e'
  | 'f'
  | 'g'
  | 'h'
  | 'j';
    `},
		{Code: `
declare const original: number | null | undefined;
const cast = original as number | null;
    `},
		{Code: `
type Type = { value: string };
declare const original: Type | number;
const cast = original as Type;
    `},
		{Code: `
type T = string;
declare const x: T | number;

const y = x as NonNullable<T>;
    `},
		{Code: `
type T = string | null;
declare const x: T | number;

const y = x as NonNullable<T>;
    `},
		{Code: `
const foo = [] as const;
    `},
		{Code: `
const x = 1 as 1;
    `},
		{Code: `
declare function foo<T = any>(): T;
const bar = foo() as number;
    `},
		{Code: `
function getValue<U, T extends U | string>(value: T | undefined): T {
  return value as T;
}
    `},
		{Code: `
function getValue<U, T extends U>(value: T | undefined): T {
  return value as T;
}
    `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
declare const maybe: string | undefined;
const bar = maybe as string;
      `,
			Output: []string{`
declare const maybe: string | undefined;
const bar = maybe!;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNonNullAssertion",
					Message:   "Use a ! assertion to more succinctly remove null and undefined from the type.",
					Line:      3,
					Column:    13,
					EndLine:   3,
					EndColumn: 28,
				},
			},
		},
		{
			Code: `
declare const maybe: string | undefined;
const bar = <string>maybe;
      `,
			Output: []string{`
declare const maybe: string | undefined;
const bar = maybe!;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNonNullAssertion",
					Line:      3,
					Column:    13,
					EndLine:   3,
					EndColumn: 26,
				},
			},
		},
		{
			Code: `
declare const maybe: string | undefined;
const bar = ((maybe)) as string;
      `,
			Output: []string{`
declare const maybe: string | undefined;
const bar = maybe!;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNonNullAssertion",
					Line:      3,
					Column:    13,
					EndLine:   3,
					EndColumn: 32,
				},
			},
		},
		{
			Code: `
declare const maybe: string | null;
const bar = maybe as string;
      `,
			Output: []string{`
declare const maybe: string | null;
const bar = maybe!;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNonNullAssertion",
					Line:      3,
					Column:    13,
				},
			},
		},
		{
			Code: `
declare const maybe: string | null | undefined;
const bar = maybe as string;
      `,
			Output: []string{`
declare const maybe: string | null | undefined;
const bar = maybe!;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNonNullAssertion",
					Line:      3,
					Column:    13,
				},
			},
		},
		{
			Code: `
type Type = { value: string };
declare const maybe: Type | undefined;
const bar = maybe as Type;
      `,
			Output: []string{`
type Type = { value: string };
declare const maybe: Type | undefined;
const bar = maybe!;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNonNullAssertion",
					Line:      4,
					Column:    13,
				},
			},
		},
		{
			Code: `
interface Interface {
  value: string;
}
declare const maybe: Interface | undefined;
const bar = maybe as Interface;
      `,
			Output: []string{`
interface Interface {
  value: string;
}
declare const maybe: Interface | undefined;
const bar = maybe!;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNonNullAssertion",
					Line:      6,
					Column:    13,
				},
			},
		},
		{
			Code: `
type T = string | null;
declare const x: T;

const y = x as NonNullable<T>;
      `,
			Output: []string{`
type T = string | null;
declare const x: T;

const y = x!;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNonNullAssertion",
					Line:      5,
					Column:    11,
				},
			},
		},
		{
			Code: `
type T = string | null | undefined;
declare const x: T;

const y = x as NonNullable<T>;
      `,
			Output: []string{`
type T = string | null | undefined;
declare const x: T;

const y = x!;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNonNullAssertion",
					Line:      5,
					Column:    11,
				},
			},
		},
		{
			Code: `
declare function nullablePromise(): Promise<string | null>;

async function fn(): Promise<string> {
  return (await nullablePromise()) as string;
}
      `,
			Output: []string{`
declare function nullablePromise(): Promise<string | null>;

async function fn(): Promise<string> {
  return (await nullablePromise())!;
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNonNullAssertion",
					Line:      5,
					Column:    10,
				},
			},
		},
		{
			Code: `
declare const a: string | null;

const b = (a || undefined) as string;
      `,
			Output: []string{`
declare const a: string | null;

const b = (a || undefined)!;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNonNullAssertion",
					Line:      4,
					Column:    11,
				},
			},
		},
		{
			Code: `
declare const value:
  | 'a'
  | 'b'
  | 'c'
  | 'd'
  | 'e'
  | 'f'
  | 'g'
  | 'h'
  | 'i'
  | undefined;
const cast = value as
  | 'a'
  | 'b'
  | 'c'
  | 'd'
  | 'e'
  | 'f'
  | 'g'
  | 'h'
  | 'i';
      `,
			Output: []string{`
declare const value:
  | 'a'
  | 'b'
  | 'c'
  | 'd'
  | 'e'
  | 'f'
  | 'g'
  | 'h'
  | 'i'
  | undefined;
const cast = value!;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferNonNullAssertion",
				Line:      13,
				Column:    14,
			}},
		},
	})
}

func TestNonNullableTypeAssertionStyleRule_noUncheckedIndexedAccess(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.noUncheckedIndexedAccess.json", t, &NonNullableTypeAssertionStyleRule, []rule_tester.ValidTestCase{
		{Code: `
function first<T>(array: ArrayLike<T>): T | null {
  return array.length > 0 ? (array[0] as T) : null;
}
      `},
		{Code: `
function first<T extends string | null>(array: ArrayLike<T>): T | null {
  return array.length > 0 ? (array[0] as T) : null;
}
      `},
		{Code: `
function first<T extends string | undefined>(array: ArrayLike<T>): T | null {
  return array.length > 0 ? (array[0] as T) : null;
}
      `},
		{Code: `
function first<T extends string | null | undefined>(
  array: ArrayLike<T>,
): T | null {
  return array.length > 0 ? (array[0] as T) : null;
}
      `},
		{Code: `
type A = 'a' | 'A';
type B = 'b' | 'B';
function first<T extends A | B | null>(array: ArrayLike<T>): T | null {
  return array.length > 0 ? (array[0] as T) : null;
}
      `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
function first<T extends string | number>(array: ArrayLike<T>): T | null {
  return array.length > 0 ? (array[0] as T) : null;
}
        `,
			Output: []string{`
function first<T extends string | number>(array: ArrayLike<T>): T | null {
  return array.length > 0 ? (array[0]!) : null;
}
        `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNonNullAssertion",
					Line:      3,
					Column:    30,
				},
			},
		},
	})
}

func TestNonNullableTypeAssertionStyleRule_withoutExplicitStrictNullChecks(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.no-explicit-strict.json", t, &NonNullableTypeAssertionStyleRule, nil, []rule_tester.InvalidTestCase{
		{
			Code: `
declare function normalize(value?: string): string | undefined;
const value = normalize() as string;
      `,
			Output: []string{`
declare function normalize(value?: string): string | undefined;
const value = normalize()!;
      `},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferNonNullAssertion",
				Line:      3,
				Column:    15,
			}},
		},
		{
			Code: `
declare const key: unique symbol;
function read(adm: { [key]?: string[] }): string[] {
  if (key in adm) {
    return adm[key] as string[];
  }
  return [];
}
      `,
			Output: []string{`
declare const key: unique symbol;
function read(adm: { [key]?: string[] }): string[] {
  if (key in adm) {
    return adm[key]!;
  }
  return [];
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferNonNullAssertion",
				Line:      5,
				Column:    12,
			}},
		},
		{
			Code: `
declare function getOptions(): { output: { hash?: string } };
if (getOptions().output.hash) {
  const hash = getOptions().output.hash as string;
}
      `,
			Output: []string{`
declare function getOptions(): { output: { hash?: string } };
if (getOptions().output.hash) {
  const hash = getOptions().output.hash!;
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferNonNullAssertion",
				Line:      4,
				Column:    16,
			}},
		},
		{
			Code: `
declare const values: string[];
if (values.length > 0) {
  const value = values.pop() as string;
}
      `,
			Output: []string{`
declare const values: string[];
if (values.length > 0) {
  const value = values.pop()!;
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferNonNullAssertion",
				Line:      4,
				Column:    17,
			}},
		},
	})
}

func TestNonNullableTypeAssertionStyleRule_strictNullChecksOverrides(t *testing.T) {
	t.Run("disabled explicitly", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.unstrict.json", t, &NonNullableTypeAssertionStyleRule, []rule_tester.ValidTestCase{
			{Code: `
declare const maybe: string | undefined;
const value = maybe as string;
        `},
		}, nil)
	})

	t.Run("enabled without strict", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.strict-null-checks-only.json", t, &NonNullableTypeAssertionStyleRule, nil, []rule_tester.InvalidTestCase{
			{
				Code: `
declare const maybe: string | undefined;
const value = maybe as string;
        `,
				Output: []string{`
declare const maybe: string | undefined;
const value = maybe!;
        `,
				},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferNonNullAssertion",
					Line:      3,
					Column:    15,
				}},
			},
		})
	})

	t.Run("disabled with strict", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.strict-with-null-checks-off.json", t, &NonNullableTypeAssertionStyleRule, []rule_tester.ValidTestCase{
			{Code: `
declare const maybe: string | undefined;
const value = maybe as string;
        `},
		}, nil)
	})
}

func TestNonNullableTypeAssertionStyleRule_editDemandAndSuppression(t *testing.T) {
	t.Parallel()

	const source = `declare const maybe: string | undefined;
const asAssertion = maybe as string;
const angleAssertion = <string>maybe;
const redundantParens = ((maybe)) as string;
// eslint-disable-next-line @typescript-eslint/non-nullable-type-assertion-style
const suppressed = maybe as string;
`
	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(
		source,
		"non-nullable-type-assertion-style-edit-demand.ts",
		"tsconfig.json",
	)
	if err != nil {
		t.Fatal(err)
	}

	run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		t.Helper()

		var diagnostics []rule.RuleDiagnostic
		_, err := linter.RunLinter(linter.RunLinterOptions{
			Programs:       []*compiler.Program{program},
			SingleThreaded: true,
			TargetFiles:    [][]string{{sourceFile.FileName()}},
			ExcludePaths:   []string{},
			GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
				return []linter.ConfiguredRule{{
					Name:             NonNullableTypeAssertionStyleRule.Name,
					Severity:         rule.SeverityError,
					RequiresTypeInfo: NonNullableTypeAssertionStyleRule.RequiresTypeInfo,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return NonNullableTypeAssertionStyleRule.Run(ctx, nil)
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
		if err != nil {
			t.Fatal(err)
		}
		if len(diagnostics) != 3 {
			t.Fatalf("demand %d: diagnostics = %d, want 3: %#v", demand, len(diagnostics), diagnostics)
		}
		return diagnostics
	}

	diagnostics := map[rule.EditDemand][]rule.RuleDiagnostic{
		rule.EditDemandNone:       run(rule.EditDemandNone),
		rule.EditDemandAutofix:    run(rule.EditDemandAutofix),
		rule.EditDemandSuggestion: run(rule.EditDemandSuggestion),
		rule.EditDemandAll:        run(rule.EditDemandAll),
	}
	withoutEdits := func(diagnostic rule.RuleDiagnostic) rule.RuleDiagnostic {
		diagnostic.FixesPtr = nil
		diagnostic.Suggestions = nil
		return diagnostic
	}
	expectedRanges := []string{
		"maybe as string",
		"<string>maybe",
		"((maybe)) as string",
	}

	for index, allEdits := range diagnostics[rule.EditDemandAll] {
		want := withoutEdits(allEdits)
		for demand, demandDiagnostics := range diagnostics {
			if got := withoutEdits(demandDiagnostics[index]); !reflect.DeepEqual(got, want) {
				t.Errorf("diagnostic %d changed for demand %d:\ngot:  %#v\nwant: %#v", index, demand, got, want)
			}
			if demandDiagnostics[index].Suggestions != nil {
				t.Errorf("diagnostic %d demand %d unexpectedly has suggestions", index, demand)
			}
		}

		if got := source[allEdits.Range.Pos():allEdits.Range.End()]; got != expectedRanges[index] {
			t.Errorf("diagnostic %d reports %q, want %q", index, got, expectedRanges[index])
		}
		autofix := diagnostics[rule.EditDemandAutofix][index].FixesPtr
		if autofix == nil || !reflect.DeepEqual(autofix, allEdits.FixesPtr) {
			t.Fatalf("diagnostic %d: autofix and all-edits demands produced different fixes", index)
		}
		if fixes := allEdits.Fixes(); len(fixes) != 1 || fixes[0].Text != "maybe!" || fixes[0].Range != allEdits.Range {
			t.Errorf("diagnostic %d fixes = %#v, want full-node replacement with maybe!", index, fixes)
		}
		if diagnostics[rule.EditDemandNone][index].FixesPtr != nil ||
			diagnostics[rule.EditDemandSuggestion][index].FixesPtr != nil {
			t.Errorf("diagnostic %d attached fixes without autofix demand", index)
		}
	}

	fixed, unapplied, changed := linter.ApplyRuleFixes(source, diagnostics[rule.EditDemandAll])
	if !changed || len(unapplied) != 0 {
		t.Fatalf("ApplyRuleFixes changed=%v unapplied=%d", changed, len(unapplied))
	}
	const expected = `declare const maybe: string | undefined;
const asAssertion = maybe!;
const angleAssertion = maybe!;
const redundantParens = maybe!;
// eslint-disable-next-line @typescript-eslint/non-nullable-type-assertion-style
const suppressed = maybe as string;
`
	if fixed != expected {
		t.Fatalf("fixed source = %q, want %q", fixed, expected)
	}
}
