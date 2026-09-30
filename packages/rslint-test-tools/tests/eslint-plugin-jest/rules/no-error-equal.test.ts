import {
  RuleTester,
  type InvalidTestCase,
  type ValidTestCase,
} from '../rule-tester';

// Upstream runs every case against a .ts fixture; the default .tsx file would
// parse `<T = Error>(v) => ...` as JSX.
const filename = 'src/virtual.ts';
const message = 'Avoid using equality matchers to check errors';

const withFilename = <T extends ValidTestCase>(cases: (T | string)[]): T[] =>
  cases.map((item) =>
    typeof item === 'string'
      ? ({ code: item, filename } as T)
      : { ...item, filename },
  );

const ruleTester = new RuleTester();

ruleTester.run('no-error-equal', {} as never, {
  valid: withFilename<ValidTestCase>([
    'expect',
    'expect.hasAssertions',
    'expect.hasAssertions()',
    'expect(a).toBe(b)',
    'expect(a).toThrow(b)',
    'expect(a).toThrowError(b)',
    `class MyError {}

expect(new MyError()).toBeInstanceOf(Error);

expect(new MyError()).toBe(new Error());
expect(new MyError()).toEqual(new Error());
expect(new MyError()).toStrictEqual(new Error());`,
    `class MyError {}

const myError = new MyError();

expect(myError).toBeInstanceOf(Error);

expect(myError).toBe(new Error());
expect(myError).toEqual(new Error());
expect(myError).toStrictEqual(new Error());`,
    `interface WithCode { code: string };

class MyError implements WithCode {}

const myError = new MyError();

expect(myError).toBeInstanceOf(Error);

expect(myError).toBe(new Error());
expect(myError).toEqual(new Error());
expect(myError).toStrictEqual(new Error());`,
    `function buildError(): Error | { code: string } | null {
  return new Error('oh noes');
}

const x: Array<number> = [1, 2, 3];

expect(buildError()).toEqual(null);
expect(x).toEqual(null);
expect(buildError()).toEqual(new Error());`,
    `type Mx = Array<number>;

const x: Mx = [1, 2, 3];

expect(buildError()).toEqual(null);
expect(x).toEqual(null);
expect(buildError()).toEqual(new Error());`,
    '<T = Error>(v) => expect(v as T).toEqual(new Error())',
    'expect(new AggregateError()).toBe(0)',
  ]),
  invalid: withFilename<InvalidTestCase>([
    {
      code: 'expect(new Error("hello world")).toEqual(new Error("hello world"))',
      errors: [{ messageId: 'equalError', message, line: 1 }],
    },
    {
      code: 'expect(Error("hello world")).toEqual(new Error("hello world"))',
      errors: [{ messageId: 'equalError', message, line: 1 }],
    },
    {
      code: 'expect(new Error("hello world")).toStrictEqual(new Error("hello world"))',
      errors: [{ messageId: 'equalError', message, line: 1 }],
    },
    {
      code: `class MyError extends Error {}

expect(new MyError()).toBeInstanceOf(Error);

expect(new MyError()).toBe(new Error());
expect(new MyError()).toEqual(new Error());
expect(new MyError()).toStrictEqual(new Error());`,
      errors: [
        { messageId: 'equalError', message, line: 6 },
        { messageId: 'equalError', message, line: 7 },
      ],
    },
    {
      code: `class MyError extends Error {}

const myError = new MyError();

expect(myError).toBeInstanceOf(Error);

expect(myError).toBe(new Error());
expect(myError).toEqual(new Error());
expect(myError).toStrictEqual(new Error());`,
      errors: [
        { messageId: 'equalError', message, line: 8 },
        { messageId: 'equalError', message, line: 9 },
      ],
    },
    {
      code: `expect(new AggregateError([], 'hello world')).toBe(new Error());
expect(new AggregateError([], 'hello world')).toEqual(new Error());
expect(new AggregateError([], 'hello world')).toStrictEqual(new Error());`,
      errors: [
        { messageId: 'equalError', message, line: 2 },
        { messageId: 'equalError', message, line: 3 },
      ],
    },
    {
      code: `const x = 'hello world';

expect(x as Error).toEqual('hello world');
expect(x as Error & {}).not.toStrictEqual('hello world');`,
      errors: [
        { messageId: 'equalError', message, line: 3 },
        { messageId: 'equalError', message, line: 4 },
      ],
    },
    {
      code: `declare function buildError<T>(): T;

expect(buildError<Error>()).toEqual('hello world');`,
      errors: [{ messageId: 'equalError', message, line: 3 }],
    },
    {
      code: `declare function addCode<T extends Error>(err: T, code: string): T & { code: string };

expect(addCode(new Error('hello world'), 'MODULE_NOT_FOUND')).toEqual('hello world');
expect(addCode(new AggregateError([], 'hello world'), 'MODULE_NOT_FOUND')).toEqual('hello world');`,
      errors: [
        { messageId: 'equalError', message, line: 3 },
        { messageId: 'equalError', message, line: 4 },
      ],
    },
    {
      code: '<T extends Error>(v: any) => expect(v as T).toStrictEqual(new Error())',
      errors: [{ messageId: 'equalError', message, line: 1 }],
    },
    {
      code: `function buildError(msg: string) {
  return new Error(msg);
}

expect(buildError('hello world')).toEqual('hello world');
expect(buildError('hello world')).toEqual(new Error('hello world'));`,
      errors: [
        { messageId: 'equalError', message, line: 5 },
        { messageId: 'equalError', message, line: 6 },
      ],
    },
    {
      code: `function buildError(msg: string): Error {
  return new Error(msg);
}

expect(buildError('hello world')).toEqual('hello world');
expect(buildError('hello world')).toEqual(new Error('hello world'));`,
      errors: [
        { messageId: 'equalError', message, line: 5 },
        { messageId: 'equalError', message, line: 6 },
      ],
    },
    {
      code: `class MyGenericError extends Error {}
class MySpecificError extends MyGenericError {}

expect(new MySpecificError('hello world')).toEqual('hello world');
expect(MySpecificError('hello world')).toStrictEqual(new Error('hello world'));`,
      errors: [{ messageId: 'equalError', message, line: 4 }],
    },
    {
      code: `interface Mx { code: string }

class MyError extends Error implements Mx {}

expect(new MyError('hello world')).toEqual('hello world');
expect(new MyError('hello world')).not.toStrictEqual(new Error('hello world'));`,
      errors: [
        { messageId: 'equalError', message, line: 5 },
        { messageId: 'equalError', message, line: 6 },
      ],
    },
    {
      code: `it('works', async () => {
  const err = await Promise.resolve(new Error('oh noes'));

  expect(err).toEqual(new Error('oh noes'));
});`,
      errors: [{ messageId: 'equalError', message, line: 4 }],
    },
  ]),
});
