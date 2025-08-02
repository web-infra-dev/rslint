package only_throw_error

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestOnlyThrowErrorRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &OnlyThrowErrorRule, []rule_tester.ValidTestCase{
		{Code: "throw new Error();"},
		{Code: "throw new Error('error');"},
		{Code: "throw Error('error');"},
		{Code: `
const e = new Error();
throw e;
    `},
		{Code: `
try {
  throw new Error();
} catch (e) {
  throw e;
}
    `},
		{Code: `
function foo() {
  return new Error();
}
throw foo();
    `},
		{Code: `
const foo = {
  bar: new Error(),
};
throw foo.bar;
    `},
		{Code: `
const foo = {
  bar: new Error(),
};

throw foo['bar'];
    `},
		{Code: `
const foo = {
  bar: new Error(),
};

const bar = 'bar';
throw foo[bar];
    `},
		{Code: `
class CustomError extends Error {}
throw new CustomError();
    `},
		{Code: `
class CustomError1 extends Error {}
class CustomError2 extends CustomError1 {}
throw new CustomError2();
    `},
		{Code: "throw (foo = new Error());"},
		{Code: "throw (1, 2, new Error());"},
		{Code: "throw 'literal' && new Error();"},
		{Code: "throw new Error() || 'literal';"},
		{Code: "throw foo ? new Error() : new Error();"},
		{Code: `
function* foo() {
  let index = 0;
  throw yield index++;
}
    `},
		{Code: `
async function foo() {
  throw await bar;
}
    `},
		{Code: `
import { Error } from './missing';
throw Error;
    `},
		{Code: `
class CustomError<T, C> extends Error {}
throw new CustomError<string, string>();
    `},
		{Code: `
class CustomError<T = {}> extends Error {}
throw new CustomError();
    `},
		{Code: `
class CustomError<T extends object> extends Error {}
throw new CustomError();
    `},
		{Code: `
function foo() {
  throw Object.assign(new Error('message'), { foo: 'bar' });
}
    `},
		{Code: `
const foo: Error | SyntaxError = bar();
function bar() {
  throw foo;
}
    `},
		{Code: `
declare const foo: Error | string;
throw foo as Error;
    `},
		{Code: "throw new Error() as Error;"},
		{Code: `
declare const nullishError: Error | undefined;
throw nullishError ?? new Error();
    `},
		{Code: `
declare const nullishError: Error | undefined;
throw nullishError || new Error();
    `},
		{Code: `
declare const nullishError: Error | undefined;
throw nullishError ? nullishError : new Error();
    `},
		{Code: `
function fun(value: any) {
  throw value;
}
    `},
		{Code: `
function fun(value: unknown) {
  throw value;
}
    `},
		{Code: `
function fun<T extends Error>(t: T): void {
  throw t;
}
    `},
		{
			Code: `
throw undefined;
      `,
			Options: OnlyThrowErrorOptions{
				Allow:                []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromLib, Name: []string{"undefined"}}},
				AllowThrowingAny:     utils.Ref(false),
				AllowThrowingUnknown: utils.Ref(false),
			},
		},
		{
			Code: `
class CustomError implements Error {}
throw new CustomError();
      `,
			Options: OnlyThrowErrorOptions{
				Allow:                []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromFile, Name: []string{"CustomError"}}},
				AllowThrowingAny:     utils.Ref(false),
				AllowThrowingUnknown: utils.Ref(false),
			},
		},
		{
			Code: `
throw new Map();
      `,
			Options: OnlyThrowErrorOptions{
				Allow:                []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromLib, Name: []string{"Map"}}},
				AllowThrowingAny:     utils.Ref(false),
				AllowThrowingUnknown: utils.Ref(false),
			},
		},
		{
			Code: `
        import { createError } from 'errors';
        throw createError();
      `,
			// TODO(port): type_matches_specifier doesn't support this yet
			Skip: true,
			Options: OnlyThrowErrorOptions{
				Allow:                []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromPackage, Name: []string{"ErrorLike"}, Package: "errors"}},
				AllowThrowingAny:     utils.Ref(false),
				AllowThrowingUnknown: utils.Ref(false),
			},
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: "throw undefined;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "undef",
				},
			},
		},
		{
			Code: "throw new String('');",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
		{
			Code: "throw 'error';",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
		{
			Code: "throw 0;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
		{
			Code: "throw false;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
		{
			Code: "throw null;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
		{
			Code: "throw {};",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
		{
			Code: "throw 'a' + 'b';",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
		{
			Code: `
const a = '';
throw a + 'b';
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
		{
			Code: "throw (foo = 'error');",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
		{
			Code: "throw (new Error(), 1, 2, 3);",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
		{
			Code: "throw 'literal' && 'not an Error';",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
		{
			Code: "throw 'literal' || new Error();",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
		{
			Code: "throw new Error() && 'literal';",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
		{
			Code: "throw 'literal' ?? new Error();",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
		{
			Code: "throw foo ? 'not an Error' : 'literal';",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
		{
			Code: "throw foo ? new Error() : 'literal';",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
		{
			Code: "throw foo ? 'literal' : new Error();",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
		{
			Code: "throw `${err}`;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
		{
			Code: `
const err = 'error';
throw err;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
		{
			Code: `
function foo(msg) {}
throw foo('error');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
		{
			Code: `
const foo = {
  msg: 'error',
};
throw foo.msg;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
		{
			Code: `
const foo = {
  msg: undefined,
};
throw foo.msg;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "undef",
				},
			},
		},
		{
			Code: `
class CustomError {}
throw new CustomError();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
		{
			Code: `
class Foo {}
class CustomError extends Foo {}
throw new CustomError();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
		{
			Code: `
const Error = null;
throw Error;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
		{
			Code: `
import { Error } from './class';
throw new Error();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
					Line:      3,
					Column:    7,
				},
			},
		},
		{
			Code: `
class CustomError<T extends object> extends Foo {}
throw new CustomError();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
					Line:      3,
					Column:    7,
				},
			},
		},
		{
			Code: `
function foo<T>() {
  const res: T;
  throw res;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
					Line:      4,
					Column:    9,
				},
			},
		},
		{
			Code: `
function foo<T>(fn: () => Promise<T>) {
  const promise = fn();
  const res = promise.then(() => {}).catch(() => {});
  throw res;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
					Line:      5,
					Column:    9,
				},
			},
		},
		{
			Code: `
function foo() {
  throw Object.assign({ foo: 'foo' }, { bar: 'bar' });
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
		{
			Code: `
const foo: Error | { bar: string } = bar();
function bar() {
  throw foo;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
		{
			Code: `
declare const foo: Error | string;
throw foo as string;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
		{
			Code: `
function fun(value: any) {
  throw value;
}
      `,
			Options: OnlyThrowErrorOptions{AllowThrowingAny: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
		{
			Code: `
function fun(value: unknown) {
  throw value;
}
      `,
			Options: OnlyThrowErrorOptions{AllowThrowingUnknown: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
		{
			Code: `
function fun<T extends number>(t: T): void {
  throw t;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
		{
			Code: `
class UnknownError implements Error {}
throw new UnknownError();
      `,
			Options: OnlyThrowErrorOptions{
				Allow:                []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromFile, Name: []string{"CustomError"}}},
				AllowThrowingAny:     utils.Ref(false),
				AllowThrowingUnknown: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
	})
}
