// TestNoUnnecessaryTypeAssertionContextual covers contextual-assignability
// edge cases and false-positive guards beyond the original upstream migration.
package no_unnecessary_type_assertion

import (
	"context"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnnecessaryTypeAssertionReportingDemand(t *testing.T) {
	const code = "declare const value: number;\r\n" +
		"/* leading */\u00a0value /* keep */ as /* type */ number;\r\n" +
		"(<number>value);\r\nvalue!;\r\n" +
		"// eslint-disable-next-line @typescript-eslint/no-unnecessary-type-assertion\r\nvalue as number;"
	p, sourceFile, err := rule_tester.NewProgramHelper(fixtures.GetRootDir()).CreateTestProgram(code, "file.ts", "tsconfig.json")
	if err != nil || sourceFile == nil {
		t.Fatalf("create program: %v", err)
	}
	typeChecker, release := p.GetTypeCheckerForFile(context.Background(), sourceFile)
	defer release()
	program := lintprogram.NewFromCompiler(p)
	for _, test := range []struct {
		name   string
		demand rule.EditDemand
	}{
		{"diagnostics", rule.EditDemandNone},
		{"autofix", rule.EditDemandAutofix},
		{"suggestions", rule.EditDemandSuggestion},
		{"all", rule.EditDemandAll},
	} {
		t.Run(test.name, func(t *testing.T) {
			var diagnostics []rule.RuleDiagnostic
			ctx := rule.RuleContext{
				SourceFile:     sourceFile,
				TypeChecker:    typeChecker,
				DisableManager: rule.NewDisableManager(sourceFile, rule.NewCommentStore(sourceFile)),
			}.WithProgram(program).WithDiagnosticConsumer(NoUnnecessaryTypeAssertionRule.Name, rule.SeverityError, rule.DiagnosticConsumer{
				Demand: test.demand,
				Report: func(diagnostic rule.RuleDiagnostic) { diagnostics = append(diagnostics, diagnostic) },
			})
			listeners := NoUnnecessaryTypeAssertionRule.Run(ctx, nil)
			var visit func(*ast.Node) bool
			visit = func(node *ast.Node) bool {
				if listener := listeners[node.Kind]; listener != nil {
					listener(node)
				}
				node.ForEachChild(visit)
				return false
			}
			for range 2 {
				diagnostics = nil
				visit(sourceFile.AsNode())
				if len(diagnostics) != 3 {
					t.Fatalf("got %d diagnostics, want 3", len(diagnostics))
				}
				for index, want := range []struct{ diagnostic, fix string }{
					{"value /* keep */ as /* type */ number", " as /* type */ number"},
					{"<number>value", "<number>"},
					{"value!", "!"},
				} {
					diagnostic := diagnostics[index]
					if got := code[diagnostic.Range.Pos():diagnostic.Range.End()]; got != want.diagnostic || diagnostic.Message.Id != "unnecessaryAssertion" {
						t.Fatalf("diagnostic %d: range %q, message %q", index, got, diagnostic.Message.Id)
					}
					if diagnostic.Suggestions != nil {
						t.Fatalf("diagnostic %d unexpectedly has suggestions", index)
					}
					if test.demand&rule.EditDemandAutofix == 0 {
						if diagnostic.FixesPtr != nil {
							t.Fatalf("diagnostic %d unexpectedly has fixes", index)
						}
						continue
					}
					if diagnostic.FixesPtr == nil || len(*diagnostic.FixesPtr) != 1 {
						t.Fatalf("diagnostic %d should have one fix", index)
					}
					fix := (*diagnostic.FixesPtr)[0]
					if got := code[fix.Range.Pos():fix.Range.End()]; got != want.fix || fix.Text != "" {
						t.Fatalf("diagnostic %d: fix replaces %q with %q", index, got, fix.Text)
					}
				}
			}
		})
	}
}

func TestNoUnnecessaryTypeAssertionMethodBoundaries(t *testing.T) {
	const declarations = `
declare const value: string | Error;
declare const receiver: any;
declare function wrap<T>(value: T): T;
`
	var valid []rule_tester.ValidTestCase
	for _, code := range []string{
		`wrap({ [receiver(value as Error)]() {} });`,
		`wrap({ get [receiver(value as Error)]() { return 0; } });`,
		`wrap({ set [receiver(value as Error)](param: string) {} });`,
		`wrap(class { [receiver(value as Error)]() {} });`,
		`wrap(class { @receiver(value as Error) method() {} });`,
		`wrap(class { field = receiver(value as Error); });`,
		`wrap(class { static { receiver(value as Error); } });`,
		`wrap({ method() { wrap(value as Error); } });`,
		`wrap(() => receiver(value as Error));`,
		`wrap({ method() { const narrowed: Error = value as Error; } });`,
	} {
		valid = append(valid, rule_tester.ValidTestCase{Code: declarations + code})
	}
	valid = append(valid, rule_tester.ValidTestCase{
		Code:    declarations + `wrap({ method() { receiver(value as Error); } });`,
		Options: []any{map[string]any{"typesToIgnore": []any{"Error"}}},
	})

	var invalid []rule_tester.InvalidTestCase
	for _, code := range []string{
		`wrap({ method() { receiver(value as Error); } });`,
		`wrap({ async method() { receiver(value as Error); } });`,
		`wrap({ *method() { yield receiver(value as Error); } });`,
		`wrap({ get error() { return receiver(value as Error); } });`,
		`wrap({ get error(): unknown { return value as Error; } });`,
		`wrap({ set error(param: string) { receiver(value as Error); } });`,
		`wrap(class { method() { receiver(value as Error); } });`,
		`wrap(class { static method() { receiver(value as Error); } });`,
		`wrap(class { get error() { return receiver(value as Error); } });`,
		`wrap(class { set error(param: string) { receiver(value as Error); } });`,
		`wrap(class { constructor() { receiver(value as Error); } });`,
		`wrap({ method(param: unknown = value as Error) {} });`,
		`wrap(class { constructor(param: unknown = value as Error) {} });`,
		`wrap({ method({ [receiver(value as Error)]: param }: any) {} });`,
		`wrap({ method() { receiver((value as Error)); } });`,
		`wrap(() => ({ method() { receiver(value as Error); } }));`,
		`wrap({ method() { wrap({ method() { receiver(value as Error); } }); } });`,
		`wrap({ method: function() { receiver(value as Error); } });`,
		`wrap({ method: () => { receiver(value as Error); } });`,
		`declare function run<T>(input: T, options: { onError(param: { error: string | Error }): void }): void;
run({}, { onError(param) { receiver(param.error as Error); } });`,
	} {
		invalid = append(invalid, rule_tester.InvalidTestCase{
			Code:   declarations + code,
			Output: []string{declarations + strings.ReplaceAll(code, " as Error", "")},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "contextuallyUnnecessary"}},
		})
	}
	invalid = append(invalid, rule_tester.InvalidTestCase{
		Code:   declarations + `wrap({ method() { receiver(<Error>value); } });`,
		Output: []string{declarations + `wrap({ method() { receiver(value); } });`},
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "contextuallyUnnecessary",
			Line:      5, Column: 28, EndLine: 5, EndColumn: 40,
		}},
	})
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryTypeAssertionRule, valid, invalid)
}

func TestNoUnnecessaryTypeAssertionContextual(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnnecessaryTypeAssertionRule,
		[]rule_tester.ValidTestCase{
			{Code: `const value = 5 as any;`},
			{
				Code:    `declare const x: number; x as /* comment */ number;`,
				Options: []any{map[string]any{"typesToIgnore": []any{"number"}}},
			},
			{
				Code:    `declare const x: number; < /* comment */ number > x;`,
				Options: []any{map[string]any{"typesToIgnore": []any{"number"}}},
			},
			{
				Code:    `declare const x: number; x as (/* comment */ (number));`,
				Options: []any{map[string]any{"typesToIgnore": []any{"number"}}},
			},
			{
				Code:    `declare const x: number; <(number)>x;`,
				Options: []any{map[string]any{"typesToIgnore": []any{"number"}}},
			},
			{Code: `declare const value: unknown; const result: number = value as number;`},
			{Code: `declare function fn(x: string | undefined): void; fn(undefined as string | undefined);`},
			{Code: `declare function fn(x: number[]): void; fn([...(1 as any)]);`},
			{Code: `declare const source: unknown; const { value } = source as { value: string };`},
			{Code: `declare let x: number | undefined; x ??= 1 as number;`},
			{Code: `declare let x: number; const y = (x = 1 as number);`},
			{Code: `declare const x: object | string; let y = x as {} | string; y = 1;`},
			{Code: `declare const x: object | string; let y = x as {} | string; y = 1;`},
			{Code: `type T = [any]; declare const x: T; declare function f(x: object): void; f(x as object);`},
			{Code: `declare const value: 'a'; const result = true ? (value as string) : 'b';`},
			{Code: `declare const value: 'a'; const result = true ? false ? value as string : 'b' : 'c';`},
			{Code: `declare const value: 'a'; switch (value as string) { case 'a': break; }`},
			{Code: `
interface W { name?: string }
interface N { name: string }
declare const n: N;
declare function id<T>(value: T): T;
const result = (id({ value: n as W })) satisfies { value: W };
`},
			{Code: `
interface W { name?: string }
interface N { name: string }
declare const n: N;
declare function id<T>(value: T): T;
const result = (((id({ value: n as W })))) satisfies { value: W };
`},
			{Code: `type ValuePath = 'values' | ` + "`values.${string}`" + `; declare function apply(paths: ValuePath[]): void; declare const ids: string[]; apply(ids.map(id => ` + "`values.${id}`" + ` as ValuePath));`},
			{Code: `
declare function fn(x: string): void;
declare function fn(x: number): void;
declare const value: string | number;
fn(value as string);
`},
			{Code: `
declare function fn<T>(x: T): T;
declare const value: 'a';
fn(value as string);
`},
			{Code: `
type Infer<T> = T extends () => infer V ? V : never;
declare function fn<T>(o: { p: T }): { [K in keyof T]: Infer<T[K]> };
const result = fn({ p: { a: Object as () => string } });
result.a.toLowerCase();
`},
			{Code: `
type Test<T extends Record<string, unknown>> = {};
declare function inferred<T extends Test<never>[]>(input: { addons?: T }): void;
inferred({ addons: [{} as Test<{ parameters: { potato: boolean } }>] });
`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:    `declare const x: number; x as /* comment */ number;`,
				Output:  []string{`declare const x: number; x;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion"}},
				Options: []any{map[string]any{"typesToIgnore": []any{}}},
			},
			{
				Code:    `declare const x: number; x as /* comment */ number;`,
				Output:  []string{`declare const x: number; x;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion"}},
				Options: []any{map[string]any{"typesToIgnore": []any{"string"}}},
			},
			{
				Code:    `declare const x: number; x as (number);`,
				Output:  []string{`declare const x: number; x;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion"}},
				Options: []any{map[string]any{"typesToIgnore": []any{"(number)"}}},
			},
			{
				Code:   `declare const x: string | undefined; (x as string | undefined)?.trim();`,
				Output: []string{`declare const x: string | undefined; (x)?.trim();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion"}},
			},
			{
				Code:   `declare const x: string | undefined; (x as string | undefined)?.[0];`,
				Output: []string{`declare const x: string | undefined; (x)?.[0];`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion"}},
			},
			{
				Code:   `declare const x: (() => string) | undefined; (x as (() => string) | undefined)?.();`,
				Output: []string{`declare const x: (() => string) | undefined; (x)?.();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion"}},
			},
			{
				Code:   `declare const x: string | undefined; (x as string | undefined) || 'fallback';`,
				Output: []string{`declare const x: string | undefined; (x) || 'fallback';`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion"}},
			},
			{
				Code:   `declare const x: string | undefined; (x as string | undefined) && x.trim();`,
				Output: []string{`declare const x: string | undefined; (x) && x.trim();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion"}},
			},
			{
				Code:   `declare const x: string | undefined; (x as string | undefined) ?? 'fallback';`,
				Output: []string{`declare const x: string | undefined; (x) ?? 'fallback';`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion"}},
			},
			{
				Code:   `declare const value: boolean; !(value as boolean);`,
				Output: []string{`declare const value: boolean; !(value);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion"}},
			},
			{
				Code:   `declare const value: Error; throw value as Error;`,
				Output: []string{`declare const value: Error; throw value;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion"}},
			},
			{
				Code:   `declare let value: string; value = 'a' as string;`,
				Output: []string{`declare let value: string; value = 'a';`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "contextuallyUnnecessary"}},
			},
			{
				Code:   `declare const value: 'a'; declare function id<T>(value: T): T; id<string>(value as string);`,
				Output: []string{`declare const value: 'a'; declare function id<T>(value: T): T; id<string>(value);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "contextuallyUnnecessary"}},
			},
			{
				Code:   `declare const value: 'a'; function f(parameter: string = value as string): void {}`,
				Output: []string{`declare const value: 'a'; function f(parameter: string = value): void {}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "contextuallyUnnecessary"}},
			},
			{
				Code:   `const x = ((1 as 1));`,
				Output: []string{`const x = ((1));`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion"}},
			},
			{
				Code:   `class C { readonly x = (1 as 1); }`,
				Output: []string{`class C { readonly x = (1); }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion"}},
			},
			{
				Code:   `declare namespace JSX { interface IntrinsicElements { div: { p: string } } } declare const x: 'a'; <div p={x as string} />;`,
				Output: []string{`declare namespace JSX { interface IntrinsicElements { div: { p: string } } } declare const x: 'a'; <div p={x} />;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "contextuallyUnnecessary"}},
				Tsx:    true,
			},
			{
				Code:   `declare namespace JSX { interface IntrinsicElements { div: { p?: string | null } } } declare const x: 'a'; <div p={true ? x as string : null} />;`,
				Output: []string{`declare namespace JSX { interface IntrinsicElements { div: { p?: string | null } } } declare const x: 'a'; <div p={true ? x : null} />;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "contextuallyUnnecessary"}},
				Tsx:    true,
			},
			{
				Code:   `declare namespace JSX { interface IntrinsicElements { div: { p?: string | null } } } declare const x: 'a'; <div p={(x as string) || null} />;`,
				Output: []string{`declare namespace JSX { interface IntrinsicElements { div: { p?: string | null } } } declare const x: 'a'; <div p={(x) || null} />;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "contextuallyUnnecessary"}},
				Tsx:    true,
			},
			{
				Code:   `declare namespace JSX { interface IntrinsicElements { div: { p?: string | null } } } declare const x: string | null; <div p={x!} />;`,
				Output: []string{`declare namespace JSX { interface IntrinsicElements { div: { p?: string | null } } } declare const x: string | null; <div p={x} />;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "contextuallyUnnecessary"}},
				Tsx:    true,
			},
			{
				Code:   `declare namespace JSX { interface IntrinsicElements { div: { p?: string | null } } } declare const x: 'a'; declare function f(value: string): string; <div p={f(x as string)} />;`,
				Output: []string{`declare namespace JSX { interface IntrinsicElements { div: { p?: string | null } } } declare const x: 'a'; declare function f(value: string): string; <div p={f(x)} />;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "contextuallyUnnecessary"}},
				Tsx:    true,
			},
			{
				Code:   `declare const x: number; x  as number;`,
				Output: []string{`declare const x: number; x;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion"}},
			},
			{
				Code:   "const v: number = 5;\nconst x = <number>(<any>v);",
				Output: []string{"const v: number = 5;\nconst x = v;"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion"}},
			},
			{
				Code:   "declare function fn(param: number): void;\nfn(42 as unknown as number);",
				Output: []string{"declare function fn(param: number): void;\nfn(42);"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "contextuallyUnnecessary",
					Line:      2,
					Column:    4,
					EndLine:   2,
					EndColumn: 27,
				}},
			},
			{
				Code:   `const fn = (): {} => <any>{};`,
				Output: []string{`const fn = (): {} => ({});`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "contextuallyUnnecessary",
				}},
			},
			{
				Code:   "function fn(x: number): void {}\nfn(5 as any);",
				Output: []string{"function fn(x: number): void {}\nfn(5);"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "contextuallyUnnecessary",
					Message:   "This assertion is unnecessary since the receiver accepts the original type of the expression.",
					Line:      2,
					Column:    4,
					EndLine:   2,
					EndColumn: 12,
				}},
			},
			{
				Code:   `type AB = 'a' | 'b'; declare const a: 'a'; declare function fn(x: AB): void; fn(a as AB);`,
				Output: []string{`type AB = 'a' | 'b'; declare const a: 'a'; declare function fn(x: AB): void; fn(a);`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "contextuallyUnnecessary",
					Line:      1,
					Column:    81,
					EndLine:   1,
					EndColumn: 88,
				}},
			},
			{
				Code:   `declare function fn(x: { value: number }): void; fn({ value: 42 as number });`,
				Output: []string{`declare function fn(x: { value: number }): void; fn({ value: 42 });`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "contextuallyUnnecessary",
					Line:      1,
					Column:    62,
					EndLine:   1,
					EndColumn: 74,
				}},
			},
			{
				Code:   `declare const a: 'a'; const value: string = a as string;`,
				Output: []string{`declare const a: 'a'; const value: string = a;`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "contextuallyUnnecessary",
					Line:      1,
					Column:    45,
					EndLine:   1,
					EndColumn: 56,
				}},
			},
			{
				Code:   `declare const a: 'a'; const value: string = <string>a;`,
				Output: []string{`declare const a: 'a'; const value: string = a;`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "contextuallyUnnecessary",
					Line:      1,
					Column:    45,
					EndLine:   1,
					EndColumn: 54,
				}},
			},
			{
				Code:   `declare function fn(x: any): void; declare const value: string | number; fn(value as number);`,
				Output: []string{`declare function fn(x: any): void; declare const value: string | number; fn(value);`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "contextuallyUnnecessary",
					Line:      1,
					Column:    77,
					EndLine:   1,
					EndColumn: 92,
				}},
			},
		},
	)
}

func TestNoUnnecessaryTypeAssertionIndexSignatures(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnnecessaryTypeAssertionRule,
		[]rule_tester.ValidTestCase{
			{Code: `
declare const config: { field?: boolean } | undefined;
export const value = (config as Record<string, unknown> | undefined)?.field;
`},
			{Code: `declare const value: { field?: boolean } | undefined; const result = (value as Record<string, unknown> | undefined)?.extra;`},
			{Code: `
declare const config: { field?: boolean } | null;
export const value = (<Record<string, unknown> | null>config)?.field;
`},
			{Code: `
declare const config: { field?: boolean } | null | undefined;
export const value = (config as Record<string, unknown> | null | undefined)?.field;
`},
			{Code: `declare const value: { field?: boolean }; const result = value as { field?: boolean; [key: string]: unknown };`},
			{Code: `declare const value: {}; const result = (value as { [key: number]: unknown })[0];`},
			{Code: `declare const key: unique symbol; declare const value: {}; const result = (value as { [key: symbol]: unknown })[key];`},
			{Code: "declare const value: {}; const result = (value as { [key: `data-${string}`]: unknown })['data-x'];"},
			{Code: `declare const value: {}; const result = value as { readonly [key: string]: unknown };`},
			{Code: `declare const value: Record<string, unknown> | undefined; const result = value as {} | undefined;`},
			{Code: `
declare const config: { field?: boolean } | undefined;
declare function identity<T>(value: T): T;
const result = identity(config as Record<string, unknown> | undefined);
`},
			{Code: `declare const value: { field?: boolean } | undefined; const result = value as unknown as Record<string, unknown> | undefined;`},
			{
				Code:     `declare const value: { field?: boolean } | undefined; const result = value as Record<string, unknown> | undefined;`,
				TSConfig: "tsconfig.exactOptionalPropertyTypes.json",
			},
			{
				Code:     `declare const value: { field?: boolean } | undefined; const result = value as Record<string, unknown> | undefined;`,
				TSConfig: "tsconfig.noUncheckedIndexedAccess.json",
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `
interface Config { field?: boolean; [key: string]: unknown }
declare const config: Config | undefined;
export const value = (config as Record<string, unknown> | undefined)?.field;
`,
				Output: []string{`
interface Config { field?: boolean; [key: string]: unknown }
declare const config: Config | undefined;
export const value = (config)?.field;
`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion"}},
			},
			{
				Code:   `declare const value: Record<string, unknown> | null; const result = <Record<string, unknown> | null>value;`,
				Output: []string{`declare const value: Record<string, unknown> | null; const result = value;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion"}},
			},
			{
				Code:   `declare const value: { field?: boolean } | undefined; const result = value as { field?: boolean } | undefined;`,
				Output: []string{`declare const value: { field?: boolean } | undefined; const result = value;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion"}},
			},
			{
				Code:   `declare const value: { [key: string]: unknown }; const result = value as { [key: string]: unknown };`,
				Output: []string{`declare const value: { [key: string]: unknown }; const result = value;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion"}},
			},
			{
				Code:     `declare const value: Record<string, unknown> | undefined; const result = value as Record<string, unknown> | undefined;`,
				Output:   []string{`declare const value: Record<string, unknown> | undefined; const result = value;`},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion"}},
				TSConfig: "tsconfig.exactOptionalPropertyTypes.json",
			},
			{
				Code:   `declare const value: { field?: boolean }; const result: Record<string, unknown> = value as Record<string, unknown>;`,
				Output: []string{`declare const value: { field?: boolean }; const result: Record<string, unknown> = value;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "contextuallyUnnecessary"}},
			},
			{
				Code:   `declare const value: { field?: boolean }; declare function consume(value: Record<string, unknown>): void; consume(value as Record<string, unknown>);`,
				Output: []string{`declare const value: { field?: boolean }; declare function consume(value: Record<string, unknown>): void; consume(value);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "contextuallyUnnecessary"}},
			},
		},
	)
}

func TestNoUnnecessaryTypeAssertionInferenceGuards(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnnecessaryTypeAssertionRule,
		[]rule_tester.ValidTestCase{
			{Code: `
namespace N { export type R = any }
type R = () => N.R;
declare const x: R;
declare function infer<T>(input: { value: T }): T;
let y = infer({ value: x as object });
y = {};
`},
			{Code: `
type R = () => R | any;
declare const x: R;
declare function infer<T>(input: { value: T }): T;
let y = infer({ value: x as object });
y = {};
`},
			{Code: `
type R = () => [R, any];
declare const x: R;
declare function infer<T>(input: { value: T }): T;
let y = infer({ value: x as object });
y = {};
`},
			{Code: `
type R = () => [any, R];
declare const x: R;
declare function infer<T>(input: { value: T }): T;
let y = infer({ value: x as object });
y = {};
`},
			{Code: `
type R = (value: any) => R;
declare const x: R;
declare function infer<T>(input: { value: T }): T;
let y = infer({ value: x as object });
y = {};
`},
			{Code: `
type Sig<T> = () => T;
type X = Sig<any>;
type A = Sig<X>;
type Root = A | X;
declare const x: Root;
declare function infer<T>(input: { value: T }): T;
let y = infer({ value: x as object });
y = {};
`},
			{Code: `
type Prev = [never, 0, 1, 2];
type R<N extends 0 | 1 | 2 | 3> = N extends 0 ? any : () => R<Prev[N]>;
declare const x: R<2>;
declare function infer<T>(input: { value: T }): T;
let y = infer({ value: x as object });
y = {};
`},
			{Code: `
type Prev = [never, 0, 1, 2];
type R<N extends 0 | 1 | 2 | 3> = N extends 0 ? string : () => R<Prev[N]>;
declare const x: R<2>;
declare function infer<T>(input: { value: T }): T;
let y = infer({ value: x as object });
y = {};
`},
			{Code: `
type A<T> = () => B<T & { a: true }>;
type B<T> = () => A<T & { b: true }>;
declare const recursive: A<{}>;
declare const value: object;
value as unknown as typeof recursive;
`},
			{Code: `
type R<T = {}> = (next: R<T & { next: true }>) => number;
declare const recursive: R;
declare const value: object;
value as unknown as typeof recursive;
`},
			{Code: `
type Test<T> = { [key: symbol]: never };
declare function inferred<T extends Test<never>[]>(input: { addons?: T }): {
  options: T[number] extends Test<infer U> ? U : unknown;
};
const test = inferred({ addons: [{} as Test<{ parameters: { potato: boolean } }>] });
test.options.parameters.potato;
`},
			{Code: `
type Pattern<T> = { [key: ` + "`data-${string}`" + `]: never };
declare function inferred<T extends Pattern<never>[]>(input: { addons?: T }): {
  options: T[number] extends Pattern<infer U> ? U : unknown;
};
const test = inferred({ addons: [{} as Pattern<{ parameters: { potato: boolean } }>] });
test.options.parameters.potato;
`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `
type Wrapper<T> = () => T;
declare const x: Wrapper<string> & Wrapper<number>;
declare function consume(value: object): void;
consume(x as object);
`,
				Output: []string{`
type Wrapper<T> = () => T;
declare const x: Wrapper<string> & Wrapper<number>;
declare function consume(value: object): void;
consume(x);
`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "contextuallyUnnecessary"}},
			},
			{
				Code: `
type Test<T> = { [key: string]: never };
declare function inferred<T extends Test<never>[]>(input: { addons?: T }): void;
inferred({ addons: [{} as Test<{ parameter: boolean }>] });
`,
				Output: []string{`
type Test<T> = { [key: string]: never };
declare function inferred<T extends Test<never>[]>(input: { addons?: T }): void;
inferred({ addons: [{}] });
`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "contextuallyUnnecessary"}},
			},
			{
				Code: `
type Test<T> = { [key: number]: never };
declare function inferred<T extends Test<never>[]>(input: { addons?: T }): void;
inferred({ addons: [{} as Test<{ parameter: boolean }>] });
`,
				Output: []string{`
type Test<T> = { [key: number]: never };
declare function inferred<T extends Test<never>[]>(input: { addons?: T }): void;
inferred({ addons: [{}] });
`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "contextuallyUnnecessary"}},
			},
		},
	)
}
