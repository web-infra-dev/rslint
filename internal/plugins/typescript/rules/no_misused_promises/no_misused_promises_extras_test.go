package no_misused_promises

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoMisusedPromisesArgumentTypes(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoMisusedPromisesRule,
		[]rule_tester.ValidTestCase{
			{Code: `declare function invoke(): void; invoke();`},
			{Code: `declare function invoke(x: string): void; invoke("value");`},
			{Code: `
declare function invoke(callback: () => void, optional?: () => void): void;
declare function invoke(callback: () => Promise<void>): void;
invoke(async () => {});
`},
			{Code: `
declare const Factory: {
  new (callback: () => void, optional?: () => void): object;
  new (callback: () => Promise<void>): object;
};
new Factory(async () => {});
`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `
export {};
declare global { interface String { (): Promise<void>; } }
declare function invoke(callback: () => void): void;
invoke("value");
`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "voidReturnArgument"}},
			},
			{
				Code: `
declare function invoke<T>(value: T, ...callbacks: [(value: T) => void, ((value: T) => void)?]): void;
invoke(Promise.resolve(1), value => value.then(() => {}));
`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "voidReturnArgument"}},
			},
			{
				Code: `
declare function invoke(value: boolean, callback: () => void): void;
declare const args: [boolean, () => Promise<void>];
invoke(...args);
invoke(true, ...[async () => {}]);
`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "voidReturnArgument"}},
			},
		})
}

func TestNoMisusedPromisesJSXAttributeTypes(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoMisusedPromisesRule,
		[]rule_tester.ValidTestCase{
			{
				Code: `
declare function View(props: { text: string; enabled: boolean; metadata: object }): null;
<View text={"value"} enabled={(true)} metadata={{ callback: async () => {} }} />;
`,
				Tsx: true,
			},
			{
				Code: `
declare function View(props: { callback: () => void }): null;
<View callback={async () => {}} />;
`,
				Tsx:     true,
				Options: map[string]any{"checksVoidReturn": map[string]any{"attributes": false}},
			},
			{
				Code: `
declare function View(props: { callback: (() => void) | (() => Promise<void>) }): null;
<View callback={async () => {}} />;
`,
				Tsx: true,
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `
export {};
declare global { interface String { (): PromiseLike<void>; } }
declare function View(props: { callback: () => void }): null;
<View callback={("value")} />;
`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "voidReturnAttribute"}},
			},
			{
				Code: `
declare function View<T>(props: { value: T; callback: (value: T) => void }): null;
<View value={Promise.resolve(1)} callback={value => value.then(() => {})} />;
`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "voidReturnAttribute"}},
			},
			{
				Code: `
interface CustomThenable { then(resolve: (value: number) => void): void; }
declare const callback: () => CustomThenable;
declare function View(props: { callback: (() => void) & { description?: string } }): null;
<View callback={callback} />;
`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "voidReturnAttribute"}},
			},
		})
}
