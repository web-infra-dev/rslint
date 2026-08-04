package no_misused_new

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoMisusedNewRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoMisusedNewRule, []rule_tester.ValidTestCase{
		{
			Code: `
				declare abstract class C {
				  foo() {}
				  get new();
				  bar();
				}
			`,
		},
		{
			Code: `
				class C {
				  constructor();
				}
			`,
		},
		{
			Code: `
				const foo = class {
				  constructor();
				};
			`,
		},
		{
			Code: `
				const foo = class {
				  new(): X;
				};
			`,
		},
		{
			Code: `
				class C {
				  constructor() {}
				}
		 `,
		},
		{
			Code: `
				const foo = class {
				  new() {}
				};
			`,
		},
		{
			Code: `
				const foo = class {
				  constructor() {}
				};
			`,
		},
		{
			Code: `
				interface I {
				  new (): {};
				}
			`,
		},
		{
			Code: `type T = { new (): T };`,
		},
		{
			Code: `
				export default class {
				  constructor();
				}
			`,
		},
		{
			Code: `
				interface foo {
				  new <T>(): bar<T>;
				}
			`,
		},
		{
			Code: `
				interface foo {
				  new <T>(): 'x';
				}
			`,
		},
		{
			Code: `
				declare class C {
				  [methodName](): C;
				  new(): Namespace.C;
				  new(): (C);
				}
				interface I {
				  [methodName](): void;
				}
				type T = {
				  [methodName](): void;
				};
			`,
		},
		{
			Code: "class C { [`new`](): C {} }",
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
				interface I {
				  new (): I;
				  constructor(): void;
				}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorMessageInterface",
					Line:      3,
				},
				{
					MessageId: "errorMessageInterface",
					Line:      4,
				},
			},
		},
		{
			Code: `
				interface G {
				  new <T>(): G<T>;
				}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorMessageInterface",
				},
			},
		},
		{
			Code: `
				type T = {
				  constructor(): void;
				};
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorMessageInterface",
				},
			},
		},
		{
			Code: `
				class C {
				  new(): C;
				}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorMessageClass",
				},
			},
		},
		{
			Code: `
				declare abstract class C {
				  new(): C;
				}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorMessageClass",
				},
			},
		},
		{
			Code: `
				interface I {
				  constructor(): '';
				}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorMessageInterface",
				},
			},
		},
		{
			Code: `
				class C {
					['constructor']() {};
				}
			`,
		},
		{
			Code: `
				class C {
				  ['new'](): C;
				}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorMessageClass",
				},
			},
		},
		{
			Code: `
				declare abstract class C {
				  ['constructor']() {};
				}
			`,
		},
		{
			Code: `
				declare abstract class C {
				  ['new'](): C;
				}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorMessageClass",
				},
			},
		},
		{
			Code: `
				interface I {
					['new'](): I;
				}
			`,
		},
		{
			Code: `
				interface I {
				  ['constructor'](): '';
				}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorMessageInterface",
				},
			},
		},
		{
			Code: "declare class C { [`new`](): C; }",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorMessageClass",
				},
			},
		},
		{
			Code: `
				declare class C {
				  ["\u006e\u0065\u0077"](): C;
				  new<T>(): C<T>;
				}
				const Value = class Inner {
				  new(): Inner;
				};
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMessageClass"},
				{MessageId: "errorMessageClass"},
				{MessageId: "errorMessageClass"},
			},
		},
		{
			Code: "interface I { [`constructor`](): void; }\ntype T = { [`constructor`](): void; };",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMessageInterface"},
				{MessageId: "errorMessageInterface"},
			},
		},
	})
}
