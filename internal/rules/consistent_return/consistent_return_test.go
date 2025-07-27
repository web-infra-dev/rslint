package consistent_return

import (
	"testing"

	"github.com/typescript-eslint/rslint/internal/rule_tester"
	"github.com/typescript-eslint/rslint/internal/rules/fixtures"
)

func TestConsistentReturnRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ConsistentReturnRule, []rule_tester.ValidTestCase{
		// Base rule cases
		{Code: `
function foo() {
  return;
}
		`},
		{Code: `
const foo = (flag: boolean) => {
  if (flag) return true;
  return false;
};
		`},
		{Code: `
class A {
  foo() {
    if (a) return true;
    return false;
  }
}
		`},
		{
			Code: `
const foo = (flag: boolean) => {
  if (flag) return;
  else return undefined;
};
			`,
			Options: map[string]interface{}{
				"treatUndefinedAsUnspecified": true,
			},
		},
		// Void return type cases
		{Code: `
declare function bar(): void;
function foo(flag: boolean): void {
  if (flag) {
    return bar();
  }
  return;
}
		`},
		{Code: `
declare function bar(): void;
const foo = (flag: boolean): void => {
  if (flag) {
    return;
  }
  return bar();
};
		`},
		{Code: `
function foo(flag?: boolean): number | void {
  if (flag) {
    return 42;
  }
  return;
}
		`},
		{Code: `
function foo(): boolean;
function foo(flag: boolean): void;
function foo(flag?: boolean): boolean | void {
  if (flag) {
    return;
  }
  return true;
}
		`},
		{Code: `
class Foo {
  baz(): void {}
  bar(flag: boolean): void {
    if (flag) return baz();
    return;
  }
}
		`},
		{Code: `
declare function bar(): void;
function foo(flag: boolean): void {
  function fn(): string {
    return '1';
  }
  if (flag) {
    return bar();
  }
  return;
}
		`},
		{Code: `
class Foo {
  foo(flag: boolean): void {
    const bar = (): void => {
      if (flag) return;
      return this.foo();
    };
    if (flag) {
      return this.bar();
    }
    return;
  }
}
		`},
		// Async cases
		{Code: `
declare function bar(): void;
async function foo(flag?: boolean): Promise<void> {
  if (flag) {
    return bar();
  }
  return;
}
		`},
		{Code: `
declare function bar(): Promise<void>;
async function foo(flag?: boolean): Promise<ReturnType<typeof bar>> {
  if (flag) {
    return bar();
  }
  return;
}
		`},
		{Code: `
async function foo(flag?: boolean): Promise<Promise<void | undefined>> {
  if (flag) {
    return undefined;
  }
  return;
}
		`},
		{Code: `
type PromiseVoidNumber = Promise<void | number>;
async function foo(flag?: boolean): PromiseVoidNumber {
  if (flag) {
    return 42;
  }
  return;
}
		`},
		{Code: `
class Foo {
  baz(): void {}
  async bar(flag: boolean): Promise<void> {
    if (flag) return baz();
    return;
  }
}
		`},
		// treatUndefinedAsUnspecified option cases
		{
			Code: `
declare const undef: undefined;
function foo(flag: boolean) {
  if (flag) {
    return undef;
  }
  return 'foo';
}
			`,
			Options: map[string]interface{}{
				"treatUndefinedAsUnspecified": false,
			},
		},
		{
			Code: `
function foo(flag: boolean): undefined {
  if (flag) {
    return undefined;
  }
  return;
}
			`,
			Options: map[string]interface{}{
				"treatUndefinedAsUnspecified": true,
			},
		},
		{
			Code: `
declare const undef: undefined;
function foo(flag: boolean): undefined {
  if (flag) {
    return undef;
  }
  return;
}
			`,
			Options: map[string]interface{}{
				"treatUndefinedAsUnspecified": true,
			},
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
function foo(flag: boolean): any {
  if (flag) return true;
  else return;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingReturnValue",
					Line:      4,
					Column:    8,
					EndLine:   4,
					EndColumn: 15,
				},
			},
		},
		{
			Code: `
function bar(): undefined {}
function foo(flag: boolean): undefined {
  if (flag) return bar();
  return;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingReturnValue",
					Line:      5,
					Column:    3,
					EndLine:   5,
					EndColumn: 10,
				},
			},
		},
		{
			Code: `
declare function foo(): void;
function bar(flag: boolean): undefined {
  function baz(): undefined {
    if (flag) return;
    return undefined;
  }
  if (flag) return baz();
  return;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedReturnValue",
					Line:      6,
					Column:    12,
					EndLine:   6,
					EndColumn: 21,
				},
				{
					MessageId: "missingReturnValue",
					Line:      9,
					Column:    3,
					EndLine:   9,
					EndColumn: 10,
				},
			},
		},
		{
			Code: `
function foo(flag: boolean): Promise<void> {
  if (flag) return Promise.resolve(void 0);
  else return;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingReturnValue",
					Line:      4,
					Column:    8,
					EndLine:   4,
					EndColumn: 15,
				},
			},
		},
		{
			Code: `
async function foo(flag: boolean): Promise<string> {
  if (flag) return;
  else return 'value';
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedReturnValue",
					Line:      4,
					Column:    15,
					EndLine:   4,
					EndColumn: 22,
				},
			},
		},
		{
			Code: `
async function foo(flag: boolean): Promise<string | undefined> {
  if (flag) return 'value';
  else return;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingReturnValue",
					Line:      4,
					Column:    8,
					EndLine:   4,
					EndColumn: 15,
				},
			},
		},
		{
			Code: `
async function foo(flag: boolean) {
  if (flag) return;
  return 1;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedReturnValue",
					Line:      4,
					Column:    10,
					EndLine:   4,
					EndColumn: 11,
				},
			},
		},
		{
			Code: `
function foo(flag: boolean): Promise<string | undefined> {
  if (flag) return;
  else return 'value';
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedReturnValue",
					Line:      4,
					Column:    15,
					EndLine:   4,
					EndColumn: 22,
				},
			},
		},
		{
			Code: `
declare function bar(): Promise<void>;
function foo(flag?: boolean): Promise<void> {
  if (flag) {
    return bar();
  }
  return;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingReturnValue",
					Line:      7,
					Column:    3,
					EndLine:   7,
					EndColumn: 10,
				},
			},
		},
		// treatUndefinedAsUnspecified option cases
		{
			Code: `
function foo(flag: boolean): undefined | boolean {
  if (flag) {
    return undefined;
  }
  return true;
}
			`,
			Options: map[string]interface{}{
				"treatUndefinedAsUnspecified": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedReturnValue",
					Line:      6,
					Column:    10,
					EndLine:   6,
					EndColumn: 14,
				},
			},
		},
		{
			Code: `
declare const undefOrNum: undefined | number;
function foo(flag: boolean) {
  if (flag) {
    return;
  }
  return undefOrNum;
}
			`,
			Options: map[string]interface{}{
				"treatUndefinedAsUnspecified": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedReturnValue",
					Line:      7,
					Column:    10,
					EndLine:   7,
					EndColumn: 20,
				},
			},
		},
	})
}