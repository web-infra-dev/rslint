package only_throw_error

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
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
			// Skip: 'errors' module doesn't exist in test fixtures, so createError()
			// resolves to 'any' and can't match the package specifier.
			Skip: true,
			Options: OnlyThrowErrorOptions{
				Allow:                []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromPackage, Name: []string{"ErrorLike"}, Package: "errors"}},
				AllowThrowingAny:     utils.Ref(false),
				AllowThrowingUnknown: utils.Ref(false),
			},
		},
		// allowRethrowing: valid cases
		{
			Code: `
try {
} catch (e) {
  throw e;
}
      `,
			Options: OnlyThrowErrorOptions{
				AllowRethrowing:      utils.Ref(true),
				AllowThrowingAny:     utils.Ref(false),
				AllowThrowingUnknown: utils.Ref(false),
			},
		},
		{
			Code: `
try {
} catch (eOuter) {
  try {
    if (Math.random() > 0.5) {
      throw eOuter;
    }
  } catch (eInner) {
    if (Math.random() > 0.5) {
      throw eOuter;
    } else {
      throw eInner;
    }
  }
}
      `,
			Options: OnlyThrowErrorOptions{
				AllowRethrowing:      utils.Ref(true),
				AllowThrowingAny:     utils.Ref(false),
				AllowThrowingUnknown: utils.Ref(false),
			},
		},
		{
			Code: `
Promise.reject('foo').catch(e => {
  throw e;
});
      `,
			Options: OnlyThrowErrorOptions{
				AllowRethrowing:      utils.Ref(true),
				AllowThrowingAny:     utils.Ref(false),
				AllowThrowingUnknown: utils.Ref(false),
			},
		},
		// allow string shorthand with generic union
		{
			Code: `
function func<T1, T2>() {
  let err: Promise<T1> | Promise<T2>;
  throw err;
}
      `,
			Options: OnlyThrowErrorOptions{
				Allow: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromString, Name: []string{"Promise"}}},
			},
		},
		// throw await resolving to Error with allowThrowingAny: false
		{
			Code: `
async function foo() {
  throw await Promise.resolve(new Error('error'));
}
      `,
			Options: OnlyThrowErrorOptions{
				AllowThrowingAny: utils.Ref(false),
			},
		},
		// generator with typed return
		{
			Code: `
function* foo(): Generator<number, void, Error> {
  throw yield 303;
}
      `,
			Options: OnlyThrowErrorOptions{
				AllowThrowingAny: utils.Ref(false),
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
		// allowRethrowing: invalid cases
		{
			Code: `
let x = 1;
Promise.reject('foo').catch(e => {
  throw x;
});
      `,
			Options: OnlyThrowErrorOptions{
				AllowRethrowing:      utils.Ref(true),
				AllowThrowingAny:     utils.Ref(false),
				AllowThrowingUnknown: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
		{
			Code: `
Promise.reject('foo').catch((...e) => {
  throw e;
});
      `,
			Options: OnlyThrowErrorOptions{
				AllowRethrowing:      utils.Ref(true),
				AllowThrowingAny:     utils.Ref(false),
				AllowThrowingUnknown: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
		{
			Code: `
declare const x: any[];
Promise.reject('foo').catch(...x, e => {
  throw e;
});
      `,
			Options: OnlyThrowErrorOptions{
				AllowRethrowing:      utils.Ref(true),
				AllowThrowingAny:     utils.Ref(false),
				AllowThrowingUnknown: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
		{
			Code: `
declare const x: any[];
Promise.reject('foo').then(...x, e => {
  throw e;
});
      `,
			Options: OnlyThrowErrorOptions{
				AllowRethrowing:      utils.Ref(true),
				AllowThrowingAny:     utils.Ref(false),
				AllowThrowingUnknown: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
		{
			Code: `
declare const onFulfilled: any;
declare const x: any[];
Promise.reject('foo').then(onFulfilled, ...x, e => {
  throw e;
});
      `,
			Options: OnlyThrowErrorOptions{
				AllowRethrowing:      utils.Ref(true),
				AllowThrowingAny:     utils.Ref(false),
				AllowThrowingUnknown: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
		{
			Code: `
Promise.reject('foo').then((...e) => {
  throw e;
});
      `,
			Options: OnlyThrowErrorOptions{
				AllowRethrowing:      utils.Ref(true),
				AllowThrowingAny:     utils.Ref(false),
				AllowThrowingUnknown: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
		{
			Code: `
Promise.reject('foo').then(e => {
  throw globalThis;
});
      `,
			Options: OnlyThrowErrorOptions{
				AllowRethrowing:      utils.Ref(true),
				AllowThrowingAny:     utils.Ref(false),
				AllowThrowingUnknown: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
		// allow string shorthand with generic union including void
		{
			Code: `
function func<T1, T2>() {
  let err: Promise<T1> | Promise<T2> | void;
  throw err;
}
      `,
			Options: OnlyThrowErrorOptions{
				Allow: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromString, Name: []string{"Promise"}}},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
		// throw await bar with allowThrowingAny: false
		{
			Code: `
async function foo() {
  throw await bar;
}
      `,
			Options: OnlyThrowErrorOptions{
				AllowThrowingAny: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
		// throw await resolving to number with allowThrowingAny: false
		{
			Code: `
async function foo() {
  throw await Promise.resolve<number>(303);
}
      `,
			Options: OnlyThrowErrorOptions{
				AllowThrowingAny: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "object",
				},
			},
		},
	})
}
