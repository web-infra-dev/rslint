package consistent_indexed_object_style

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestConsistentIndexedObjectStyleRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ConsistentIndexedObjectStyleRule, []rule_tester.ValidTestCase{
		// Ordinary non-matches.
		{Code: "type Foo = Record<string, unknown>;"},
		{Code: "type Foo = { [key: string]: unknown; value: unknown };"},
		{Code: "interface Foo { [key: string]: unknown; value: unknown }"},
		{Code: "interface Foo { [key: string] }"},
		{Code: "interface Foo { [...key: string]: unknown }"},

		// Direct circular references through every upstream-recursed type shape.
		{Code: "type Foo = { [key: string]: Foo };"},
		{Code: "type Foo = { [key: string]: string | Foo };"},
		{Code: "interface Foo { [key: string]: Foo & {} }"},
		{Code: "interface Foo<T> { [key: string]: Foo<T> extends T ? string : number }"},
		{Code: "interface Foo { [key: string]: Foo[number] }"},
		{Code: "interface Foo { [key: string]: Array<Foo> }"},
		{Code: "type Foo = { [K in string]: Foo };"},
		// Upstream resolves a same-named declaration type parameter as the circular target.
		{Code: "type Foo<Foo> = { [key: string]: Foo };"},
		{Code: "interface Foo<Foo> { [key: string]: Foo }"},

		// Real-user circular graphs from upstream #7148/#7863 and the rspack audit.
		{Code: `
type Tree1 = { [key: string]: number | Tree2 };
type Tree2 = { [key: string]: number | Tree1 };
`},
		{Code: `
type JsonPrimitive = string | number | boolean | null;
type JsonArray = JsonValue[];
type JsonValue = JsonPrimitive | JsonObject | JsonArray;
type JsonObject = { [key: string]: JsonValue };
`},
		{Code: `
type Link0 = Link1;
type Link1 = Link2;
type Link2 = Link3;
type Link3 = Link4;
type Link4 = Link5;
type Link5 = Link6;
type Link6 = Link7;
type Link7 = Link8;
type Link8 = Foo;
type Foo = { [key: string]: Link0 };
`},
		{Code: `
type ExampleUnion = boolean | number;
type ExampleRoot = ExampleUnion | ExampleObject;
interface ExampleObject { [key: string]: ExampleRoot }
`},
		{Code: `
type JsonValueTypes = null | string | number | boolean | JsonObjectTypes | JsonValueTypes[];
type JsonObjectTypes = { [index: string]: JsonValueTypes } & {
  [index: string]: undefined | null | string | number | boolean | JsonObjectTypes | JsonValueTypes[];
};
`},
		{Code: `
type JsonPrimitive = string | number | boolean | null;
type JsonArray = JsonValue[];
type JsonValue = JsonPrimitive | JsonObject | JsonArray;
type JsonObject = { [Key in string]: JsonValue } & {
  [Key in string]?: JsonValue | undefined;
};
`},

		// Mapped-type exclusions.
		{Code: "type KeyValue = { [K in string]: K };"},
		{Code: `type EscapedKey = { [\u004b in string]: \u004b };`},
		{Code: "type KeyArray = { [K in string]: K[] };"},
		{Code: "type KeyRemap = { [K in string as K]: unknown };"},
		{Code: "type Keyof = { [K in keyof Base]: unknown };"},
		{Code: "type OuterCycle = { [K in string]: unknown } | OuterCycle;"},
		{Code: `
type MappedValue = string | MappedObject;
type MappedObject = { [K in string]: MappedValue };
`},
		// Index-signature mode requires an unqualified Record with exactly two arguments.
		{Code: "type Foo = Record;", Options: "index-signature"},
		{Code: "type Foo = Record<string>;", Options: "index-signature"},
		{Code: "type Foo = Record<string, unknown, never>;", Options: "index-signature"},
		{Code: "type Foo = Namespace.Record<string, unknown>;", Options: "index-signature"},
		{Code: "type Foo = { [key: string]: unknown };", Options: "index-signature"},
	}, []rule_tester.InvalidTestCase{
		{
			Code:   "interface Foo { [key: string]: unknown; }",
			Output: []string{"type Foo = Record<string, unknown>;"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferRecord",
				Message:   "A record is preferred over an index signature.",
				Line:      1,
				Column:    1,
			}},
		},
		{
			Code:   "interface Foo<A = unknown> { readonly [key: string]: A; }",
			Output: []string{"type Foo<A = unknown> = Readonly<Record<string, A>>;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRecord"}},
		},
		{
			// Heritage makes the conversion unsafe, but upstream still reports.
			Code:   "interface Foo extends Base { [key: string]: unknown }",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRecord"}},
		},
		{
			Code: "export default interface Foo { [key: string]: unknown }",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferRecord",
				Column:    16,
			}},
		},
		{
			Code:   "export interface Foo { [key: string]: unknown }",
			Output: []string{"export type Foo = Record<string, unknown>;"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferRecord",
				Column:    8,
			}},
		},
		{
			Code:   "type Foo = { [key: string]: unknown };",
			Output: []string{"type Foo = Record<string, unknown>;"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferRecord",
				Line:      1,
				Column:    12,
				EndLine:   1,
				EndColumn: 38,
			}},
		},

		// Upstream intentionally does not treat these wrappers as circular.
		{
			Code:   "interface Foo { [key: string]: Foo[] }",
			Output: []string{"type Foo = Record<string, Foo[]>;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRecord"}},
		},
		{
			Code:   "interface Foo { [key: string]: [Foo] }",
			Output: []string{"type Foo = Record<string, [Foo]>;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRecord"}},
		},
		{
			Code:   "interface Foo { [key: string]: () => Foo }",
			Output: []string{"type Foo = Record<string, () => Foo>;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRecord"}},
		},
		{
			Code:   "interface Foo { [key: string]: { value: Foo } }",
			Output: []string{"type Foo = Record<string, { value: Foo }>;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRecord"}},
		},
		{
			// The outer declaration is circular; only the nested type literal is reported.
			Code:   "interface Foo { [key: string]: { [inner: string]: Foo } }",
			Output: []string{"interface Foo { [key: string]: Record<string, Foo> }"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferRecord",
				Column:    32,
			}},
		},
		{
			Code:   "type Foo = { [key: string]: { [inner: string]: Foo } };",
			Output: []string{"type Foo = { [key: string]: Record<string, Foo> };"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferRecord",
				Column:    29,
			}},
		},
		{
			Code:   "type Foo = (value: { [key: string]: Foo }) => void;",
			Output: []string{"type Foo = (value: Record<string, Foo>) => void;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRecord"}},
		},
		{
			Code: `
type CycleA = CycleB;
type CycleB = CycleA;
type Foo = { [key: string]: CycleA };
`,
			Output: []string{`
type CycleA = CycleB;
type CycleB = CycleA;
type Foo = Record<string, CycleA>;
`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRecord"}},
		},
		{
			// A nested mapped key named like the alias shadows it and is not a circular reference.
			Code:   "type Foo = { [key: string]: { [Foo in string]: Foo } };",
			Output: []string{"type Foo = Record<string, { [Foo in string]: Foo }>;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRecord"}},
		},
		{
			// An inferred type parameter with the alias's name is likewise a distinct binding.
			Code:   "type Foo = { [key: string]: unknown extends infer Foo ? Foo : never };",
			Output: []string{"type Foo = Record<string, unknown extends infer Foo ? Foo : never>;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRecord"}},
		},
		{
			// More than eight declarations exercises the allocation-free visited set's overflow path.
			Code: `
type Cycle0 = Cycle1;
type Cycle1 = Cycle2;
type Cycle2 = Cycle3;
type Cycle3 = Cycle4;
type Cycle4 = Cycle5;
type Cycle5 = Cycle6;
type Cycle6 = Cycle7;
type Cycle7 = Cycle8;
type Cycle8 = Cycle0;
type Foo = { [key: string]: Cycle0 };
`,
			Output: []string{`
type Cycle0 = Cycle1;
type Cycle1 = Cycle2;
type Cycle2 = Cycle3;
type Cycle3 = Cycle4;
type Cycle4 = Cycle5;
type Cycle5 = Cycle6;
type Cycle6 = Cycle7;
type Cycle7 = Cycle8;
type Cycle8 = Cycle0;
type Foo = Record<string, Cycle0>;
`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRecord"}},
		},
		{
			Code:   "declare interface Foo { [key: string]: unknown }",
			Output: []string{"type Foo = Record<string, unknown>;"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferRecord",
				Column:    1,
			}},
		},
		{
			Code:   "export declare interface Foo { [key: string]: unknown }",
			Output: []string{"export type Foo = Record<string, unknown>;"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferRecord",
				Column:    8,
			}},
		},

		// Mapped types accept arbitrary constraints unless an upstream exclusion applies.
		{
			Code:   "type Foo = { [K in PropertyKey]: unknown };",
			Output: []string{"type Foo = Record<PropertyKey, unknown>;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRecord"}},
		},
		{
			Code:   "type Foo<Key extends PropertyKey> = { [K in Key]: unknown };",
			Output: []string{"type Foo<Key extends PropertyKey> = Record<Key, unknown>;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRecord"}},
		},
		{
			Code:   "type Foo = { [K in 'a' | 'b']: unknown };",
			Output: []string{"type Foo = Record<'a' | 'b', unknown>;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRecord"}},
		},
		{
			Code:   "type Foo = { [K in (keyof Base)]: unknown };",
			Output: []string{"type Foo = Record<keyof Base, unknown>;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRecord"}},
		},
		{
			Code:   "type Foo = { readonly [K in PropertyKey]-?: string };",
			Output: []string{"type Foo = Readonly<Required<Record<PropertyKey, string>>>;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRecord"}},
		},
		{
			Code:   "type Foo = { [K in string]?: unknown };",
			Output: []string{"type Foo = Partial<Record<string, unknown>>;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRecord"}},
		},
		{
			Code:   "type Foo = { -readonly [K in string]: unknown };",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRecord"}},
		},
		{
			// Upstream #11179: a missing mapped value converts to any.
			Code:   "type Foo = { [K in string] };",
			Output: []string{"type Foo = Record<string, any>;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRecord"}},
		},
		{
			// The inner K shadows the mapped key, so it does not block conversion.
			Code:   "type Foo = { [K in string]: <K>() => K };",
			Output: []string{"type Foo = Record<string, <K>() => K>;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRecord"}},
		},

		// Index-signature mode reports every unqualified two-argument Record.
		{
			Code:    "type Foo = Record<string, unknown>;",
			Options: "index-signature",
			Output:  []string{"type Foo = { [key: string]: unknown };"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferIndexSignature",
				Message:   "An index signature is preferred over a record.",
				Column:    12,
			}},
		},
		{
			Code:    "type Foo = Record<Key, unknown>;",
			Options: "index-signature",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferIndexSignature",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "preferIndexSignatureSuggestion",
					Output:    "type Foo = { [key: Key]: unknown };",
				}},
			}},
		},
		{
			Code:    "type Foo = Record<string | number, unknown>;",
			Options: "index-signature",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferIndexSignature",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "preferIndexSignatureSuggestion",
					Output:    "type Foo = { [key: string | number]: unknown };",
				}},
			}},
		},

		// Comments outside preserved key/value nodes downgrade fixes to suggestions.
		{
			Code: "type Foo = { /* keep */ [key: string]: unknown };",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferRecord",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "preferRecordSuggestion",
					Output:    "type Foo = Record<string, unknown>;",
				}},
			}},
		},
		{
			// A comment before the opening brace is outside the reported node and must survive the range fast path.
			Code:   "type Foo = /* keep */ { [key: string]: unknown };",
			Output: []string{"type Foo = /* keep */ Record<string, unknown>;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRecord"}},
		},
		{
			Code: "type Foo = { [K in /* keep */ 'a' | 'b']: unknown };",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferRecord",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "preferRecordSuggestion",
					Output:    "type Foo = Record<'a' | 'b', unknown>;",
				}},
			}},
		},
		{
			// A comment inside the preserved key type remains in an autofix.
			Code:   "type Foo = { [K in 'a' /* keep */ | 'b']: unknown };",
			Output: []string{"type Foo = Record<'a' /* keep */ | 'b', unknown>;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRecord"}},
		},
	})
}

func TestConsistentIndexedObjectStyleDisableDirectives(t *testing.T) {
	t.Parallel()

	const code = `// eslint-disable-next-line @typescript-eslint/consistent-indexed-object-style
type NextLine = { [key: string]: unknown };
/* eslint-disable @typescript-eslint/consistent-indexed-object-style */
type Scoped = { [key: string]: unknown };
/* eslint-enable @typescript-eslint/consistent-indexed-object-style */
type Reported = { [key: string]: unknown };`

	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(code, "disable-directives.ts", "tsconfig.json")
	if err != nil {
		t.Fatal(err)
	}
	sourceProgram := lintprogram.NewFromCompiler(program)

	var diagnostics []rule.RuleDiagnostic
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program:      sourceProgram,
		File:         sourceFile.FileName(),
		ExcludePaths: []string{},
		GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
			return []linter.ConfiguredRule{{
				Name:     "@typescript-eslint/consistent-indexed-object-style",
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return ConsistentIndexedObjectStyleRule.Run(ctx, nil)
				},
			}}
		},
		Consumer: rule.DiagnosticConsumer{
			Demand: rule.EditDemandNone,
			Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			},
		},
	})
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %d, want only the enabled type", len(diagnostics))
	}
	if got := code[diagnostics[0].Range.Pos():diagnostics[0].Range.End()]; got != "{ [key: string]: unknown }" {
		t.Fatalf("reported range text = %q", got)
	}
}

func TestConsistentIndexedObjectStyleEditDemand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		code        string
		options     []any
		disposition []editDemandExpectation
	}{
		{
			name: "record mode",
			code: `
type Fix = { [key: string]: unknown };
type Suggest = { /* keep */ [key: string]: unknown };
interface None extends Base { [key: string]: unknown }
`,
			disposition: []editDemandExpectation{expectAutofix, expectSuggestion, expectNoEdit},
		},
		{
			name: "index-signature mode",
			code: `
type Fix = Record<string, unknown>;
type Suggest = Record<Key, unknown>;
`,
			options:     []any{"index-signature"},
			disposition: []editDemandExpectation{expectAutofix, expectSuggestion},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
			program, sourceFile, err := helper.CreateTestProgram(test.code, "edit-demand.ts", "tsconfig.json")
			if err != nil {
				t.Fatal(err)
			}
			sourceProgram := lintprogram.NewFromCompiler(program)

			run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
				t.Helper()
				var diagnostics []rule.RuleDiagnostic
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program:      sourceProgram,
					File:         sourceFile.FileName(),
					ExcludePaths: []string{},
					GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
						return []linter.ConfiguredRule{{
							Name:     ConsistentIndexedObjectStyleRule.Name,
							Severity: rule.SeverityError,
							Run: func(ctx rule.RuleContext) rule.RuleListeners {
								return ConsistentIndexedObjectStyleRule.Run(ctx, test.options)
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
				if len(diagnostics) != len(test.disposition) {
					t.Fatalf("demand %d: diagnostics = %d, want %d", demand, len(diagnostics), len(test.disposition))
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

			for index, expectation := range test.disposition {
				identity := withoutEdits(diagnostics[rule.EditDemandAll][index])
				for demand, demandDiagnostics := range diagnostics {
					if got := withoutEdits(demandDiagnostics[index]); !reflect.DeepEqual(got, identity) {
						t.Errorf("diagnostic %d demand %d changed identity:\ngot  %#v\nwant %#v", index, demand, got, identity)
					}
				}

				none := diagnostics[rule.EditDemandNone][index]
				autofix := diagnostics[rule.EditDemandAutofix][index]
				suggestion := diagnostics[rule.EditDemandSuggestion][index]
				all := diagnostics[rule.EditDemandAll][index]
				if none.FixesPtr != nil || none.Suggestions != nil || autofix.Suggestions != nil || suggestion.FixesPtr != nil {
					t.Errorf("diagnostic %d leaked an edit into the wrong demand", index)
				}
				switch expectation {
				case expectAutofix:
					if autofix.FixesPtr == nil || all.FixesPtr == nil || suggestion.Suggestions != nil || all.Suggestions != nil {
						t.Errorf("diagnostic %d did not remain autofix-only", index)
					}
				case expectSuggestion:
					if suggestion.Suggestions == nil || all.Suggestions == nil || autofix.FixesPtr != nil || all.FixesPtr != nil {
						t.Errorf("diagnostic %d did not remain suggestion-only", index)
					}
				case expectNoEdit:
					if autofix.FixesPtr != nil || suggestion.Suggestions != nil || all.FixesPtr != nil || all.Suggestions != nil {
						t.Errorf("diagnostic %d unexpectedly materialized an edit", index)
					}
				}
			}
		})
	}
}

type editDemandExpectation uint8

const (
	expectAutofix editDemandExpectation = iota
	expectSuggestion
	expectNoEdit
)
