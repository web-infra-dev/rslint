import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('prefer-called-exactly-once-with', {
  valid: [
    { code: `expect(fn).toHaveBeenCalledExactlyOnceWith()` },
    { code: `expect(x).toHaveBeenCalledOnce()` },
    { code: `expect(x).toHaveBeenCalledWith('hoge')` },
    {
      code: `expect(x).toHaveBeenCalledOnce(); expect(x).not.toHaveBeenCalledWith('hoge')`,
    },
    {
      code: `expect(x).resolves.toHaveBeenCalledOnce(); expect(x).toHaveBeenCalledWith('hoge')`,
    },
    {
      code: `expect(x).toHaveBeenCalledOnce(); x.mockClear(); expect(x).toHaveBeenCalledWith('hoge')`,
    },
    {
      code: `expect(obj.fn).toHaveBeenCalledOnce(); obj /* keep */ .fn.mockClear(); expect(obj.fn).toHaveBeenCalledWith('a')`,
    },
    {
      code: `expect(obj.fn).toHaveBeenCalledOnce(); (obj.fn).mockClear(); expect(obj.fn).toHaveBeenCalledWith('a')`,
    },
    {
      code: `expect(x).to.have.been.calledOnce.and.not.calledWith('a')`,
    },
    {
      code: `expect(x).toHaveBeenCalledOnce(); if (c) { x.mockClear(); } expect(x).toHaveBeenCalledWith('hoge')`,
    },
    {
      code: `expect(obj.fn).toHaveBeenCalledOnce(); obj.fn.mockRestore(); expect(obj.fn).toHaveBeenCalledWith('hoge')`,
    },
    {
      code: `expect(x).toHaveBeenCalledOnce(); rstest.clearAllMocks(); expect(x).toHaveBeenCalledWith('a')`,
    },
    {
      code: `expect(x).toHaveBeenCalledOnce(); rstest.resetAllMocks(); expect(x).toHaveBeenCalledWith('a')`,
    },
    {
      code: `expect(x).toHaveBeenCalledOnce(); restoreAllMocks(); expect(x).toHaveBeenCalledWith('a')`,
    },
    {
      code: `expect(x).toHaveBeenCalledOnce(); rstest.mocked(x).mockClear(); expect(x).toHaveBeenCalledWith('a')`,
    },
    {
      code: `expect(rstest.mocked(x)).toHaveBeenCalledOnce(); x.mockClear(); expect(rstest.mocked(x)).toHaveBeenCalledWith('a')`,
    },
    {
      code: `expect(getMock()).toHaveBeenCalledOnce(); getMock().mockClear(); expect(getMock()).toHaveBeenCalledWith('a')`,
    },
    {
      code: `expect(x).toHaveBeenCalledOnce(); x = createMock(); expect(x).toHaveBeenCalledWith('a')`,
    },
    {
      code: `expect(x).toHaveBeenCalledOnce(); x('b'); expect(x).toHaveBeenCalledWith('a')`,
    },
    {
      code: `expect(x).toHaveBeenCalledOnce(); doMoreWork(); expect(x).toHaveBeenCalledWith('a')`,
    },
    {
      code: `expect(x).toHaveBeenCalledOnce(); console.log(x); expect(x).toHaveBeenCalledWith('a')`,
    },
    {
      code: `expect(x).toHaveBeenCalledOnce(); [1, 2].forEach(cb); expect(x).toHaveBeenCalledWith('a')`,
    },
    {
      code: `expect(x).toHaveBeenCalledOnce(); const y = compute(); expect(x).toHaveBeenCalledWith('a')`,
    },
    {
      code: `expect(x).toHaveBeenCalledOnce(); expect(callX).toThrow(); expect(x).toHaveBeenCalledWith('a')`,
    },
    {
      code: `expect(x).toHaveBeenCalledOnce(); expect(callX).toThrowErrorMatchingSnapshot(); expect(x).toHaveBeenCalledWith('a')`,
    },
    {
      code: `expect(x).toHaveBeenCalledOnce(); expect(callX).toThrowErrorMatchingInlineSnapshot(); expect(x).toHaveBeenCalledWith('a')`,
    },
    {
      code: `expect(x).toHaveBeenCalledOnce(); expect(v).to.satisfy(pred); expect(x).toHaveBeenCalledWith('a')`,
    },
    {
      code: `expect(x).toHaveBeenCalledOnce(); expect(v).to.satisfies(pred); expect(x).toHaveBeenCalledWith('a')`,
    },
    {
      code: `expect(x).toHaveBeenCalledOnce(); expect(callX).to.change(obj, 'value'); expect(x).toHaveBeenCalledWith('a')`,
    },
    {
      code: `expect(x).toHaveBeenCalledOnce(); expect(callX).to.increase(obj, 'value'); expect(x).toHaveBeenCalledWith('a')`,
    },
    {
      code: `expect(x).toHaveBeenCalledOnce(); expect(callX).to.decrease(obj, 'value'); expect(x).toHaveBeenCalledWith('a')`,
    },
    {
      code: `expect(x).toHaveBeenCalledOnce(); eval('x("b")'); expect(x).toHaveBeenCalledWith('a')`,
    },
    {
      code: `expect(x).toHaveBeenCalledOnce(); globalThis.eval('x("b")'); expect(x).toHaveBeenCalledWith('a')`,
    },
    {
      code: `declare const x: { (a: string): void }; expect(x).toHaveBeenCalledOnce(); x.call(null, 'b'); expect(x).toHaveBeenCalledWith('a')`,
    },
    {
      code: `declare const obj: { fn: (a: string) => void }; expect(obj.fn).toHaveBeenCalledOnce(); obj.fn.apply(null, ['b']); expect(obj.fn).toHaveBeenCalledWith('a')`,
    },
    {
      code: `declare const cmp: ((a: number, b: number) => number) | undefined; expect(x).toHaveBeenCalledOnce(); [2, 1].sort(cmp); expect(x).toHaveBeenCalledWith('a')`,
    },
    {
      code: `declare const replacer: ((k: string, v: unknown) => unknown) | undefined; expect(x).toHaveBeenCalledOnce(); JSON.stringify({}, replacer); expect(x).toHaveBeenCalledWith('a')`,
    },
    {
      code: `declare const pred: ((v: number) => boolean) | undefined; expect(x).toHaveBeenCalledOnce(); [1, 2].every(pred!); expect(x).toHaveBeenCalledWith('a')`,
    },
    {
      code: `declare const t: object; expect(x).toHaveBeenCalledOnce(); [1].forEach(t as unknown as () => void); expect(x).toHaveBeenCalledWith('a')`,
    },
    {
      code: `async function f() { await expect.poll(getSpy).toHaveBeenCalledOnce(); await expect.poll(getSpy).toHaveBeenCalledWith('a'); }`,
    },
    {
      code: `async function f() { await expect.element(locator).toHaveBeenCalledOnce(); await expect.element(locator).toHaveBeenCalledWith('a'); }`,
    },
    {
      code: `expect(x).toHaveBeenCalledOnce(); expect(y).toEqual(expect.toSatisfy(pred)); expect(x).toHaveBeenCalledWith('a')`,
    },
    {
      code: `expect(x).toHaveBeenCalledOnce(); expect(y).toEqual(expect.schemaMatching(schema)); expect(x).toHaveBeenCalledWith('a')`,
    },
    {
      code: `expect(x).toHaveBeenCalledOnce(); expect(y).toEqual([...gen]); expect(x).toHaveBeenCalledWith('a')`,
    },
    {
      code: `expect(x).toHaveBeenCalledOnce(); const [a] = items; expect(x).toHaveBeenCalledWith('a')`,
    },
    {
      code: `expect(x).toHaveBeenCalledOnce(); const { b } = obj; expect(x).toHaveBeenCalledWith('a')`,
    },
    {
      code: `async function f() { expect(x).toHaveBeenCalledOnce(); await expect(p).resolves.toBe(1); expect(x).toHaveBeenCalledWith('a'); }`,
    },
    {
      code: `expect(x).toHaveBeenCalledOnce(); expect(y).toEqual(makeExpected()); expect(x).toHaveBeenCalledWith('a')`,
    },
    {
      code: `expect.soft(x).toHaveBeenCalledOnce(); expect(x).toHaveBeenCalledWith('hoge')`,
    },
    {
      code: `import { expect } from 'vitest'; expect(x).toHaveBeenCalledOnce(); expect(x).toHaveBeenCalledWith('a')`,
    },
    { code: `expect(x).to.have.been.calledOnceWith('a')` },
    {
      code: `async function f() { await expect(x).resolves.toHaveBeenCalledOnce(); expect(x).resolves.toHaveBeenCalledWith('a'); }`,
    },
  ],
  invalid: [
    {
      code: `expect(x).toHaveBeenCalledOnce(); expect(x).toHaveBeenCalledWith('hoge');`,
      output: ` expect(x).toHaveBeenCalledExactlyOnceWith('hoge');`,
      errors: [
        {
          messageId: 'preferCalledExactlyOnceWith',
          message:
            'Using `toHaveBeenCalledOnce` and `toHaveBeenCalledWith` on the same target; prefer `toHaveBeenCalledExactlyOnceWith` instead.',
          line: 1,
          column: 45,
          endLine: 1,
          endColumn: 65,
        },
      ],
    },
    {
      code: `expect(x).toHaveBeenCalledWith<[string, number]>('hoge', /* keep */ 123);
expect(x).toHaveBeenCalledOnce();`,
      output: `expect(x).toHaveBeenCalledExactlyOnceWith<[string, number]>('hoge', /* keep */ 123);
`,
      errors: [
        {
          messageId: 'preferCalledExactlyOnceWith',
          message:
            'Using `toHaveBeenCalledOnce` and `toHaveBeenCalledWith` on the same target; prefer `toHaveBeenCalledExactlyOnceWith` instead.',
          line: 2,
          column: 11,
          endLine: 2,
          endColumn: 31,
        },
      ],
    },
    {
      code: `test('x', ({ expect }) => {
  expect(x).toHaveBeenCalledOnce();
  expect(x).toHaveBeenCalledWith('a');
});`,
      output: `test('x', ({ expect }) => {
  expect(x).toHaveBeenCalledExactlyOnceWith('a');
});`,
      errors: [
        {
          messageId: 'preferCalledExactlyOnceWith',
          message:
            'Using `toHaveBeenCalledOnce` and `toHaveBeenCalledWith` on the same target; prefer `toHaveBeenCalledExactlyOnceWith` instead.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 33,
        },
      ],
    },
    {
      code: `import { expect as check } from '@rstest/core';
check(x).toHaveBeenCalledOnce();
check(x).toHaveBeenCalledWith('a');`,
      errors: [
        {
          messageId: 'preferCalledExactlyOnceWith',
          message:
            'Using `toHaveBeenCalledOnce` and `toHaveBeenCalledWith` on the same target; prefer `toHaveBeenCalledExactlyOnceWith` instead.',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 30,
        },
      ],
      output:
        "import { expect as check } from '@rstest/core';\ncheck(x).toHaveBeenCalledExactlyOnceWith('a');",
    },
    {
      code: `import.meta.rstest.expect(x).toHaveBeenCalledOnce();
import.meta.rstest.expect(x).toHaveBeenCalledWith('a');`,
      errors: [
        {
          messageId: 'preferCalledExactlyOnceWith',
          message:
            'Using `toHaveBeenCalledOnce` and `toHaveBeenCalledWith` on the same target; prefer `toHaveBeenCalledExactlyOnceWith` instead.',
          line: 2,
          column: 30,
          endLine: 2,
          endColumn: 50,
        },
      ],
      output:
        "import.meta.rstest.expect(x).toHaveBeenCalledExactlyOnceWith('a');",
    },
    {
      code: `expect(x).to.have.been.calledOnce;
expect(x).to.have.been.calledWith('a');`,
      output: `expect(x).to.have.been.calledOnceWith('a');`,
      errors: [
        {
          messageId: 'preferCalledExactlyOnceWith',
          message:
            'Using `calledOnce` and `calledWith` on the same target; prefer `calledOnceWith` instead.',
          line: 2,
          column: 24,
          endLine: 2,
          endColumn: 34,
        },
      ],
    },
    {
      code: `expect(x).to.have.been.calledOnce.and.calledWith('a');`,
      errors: [
        {
          messageId: 'preferCalledExactlyOnceWith',
          message:
            'Using `calledOnce` and `calledWith` on the same target; prefer `calledOnceWith` instead.',
          line: 1,
          column: 39,
          endLine: 1,
          endColumn: 49,
        },
      ],
    },
    {
      code: `expect(x).to.have.been.calledOnce.and.to.be.ok;
expect(x).to.have.been.calledWith('a');`,
      errors: [
        {
          messageId: 'preferCalledExactlyOnceWith',
          message:
            'Using `calledOnce` and `calledWith` on the same target; prefer `calledOnceWith` instead.',
          line: 2,
          column: 24,
          endLine: 2,
          endColumn: 34,
        },
      ],
    },
    {
      code: `expect(getMock()).toHaveBeenCalledOnce(); expect(getMock()).toHaveBeenCalledWith('a');`,
      errors: [
        {
          messageId: 'preferCalledExactlyOnceWith',
          message:
            'Using `toHaveBeenCalledOnce` and `toHaveBeenCalledWith` on the same target; prefer `toHaveBeenCalledExactlyOnceWith` instead.',
          line: 1,
          column: 61,
          endLine: 1,
          endColumn: 81,
        },
      ],
    },
    {
      code: `expect(mocks['a']).toHaveBeenCalledOnce(); expect(mocks['a']).toHaveBeenCalledWith('x');`,
      output: ` expect(mocks['a']).toHaveBeenCalledExactlyOnceWith('x');`,
      errors: [
        {
          messageId: 'preferCalledExactlyOnceWith',
          message:
            'Using `toHaveBeenCalledOnce` and `toHaveBeenCalledWith` on the same target; prefer `toHaveBeenCalledExactlyOnceWith` instead.',
          line: 1,
          column: 63,
          endLine: 1,
          endColumn: 83,
        },
      ],
    },
    {
      code: `expect(x).toHaveBeenCalledOnce();
console.log('checkpoint');
expect(x).toHaveBeenCalledWith('a');`,
      output:
        "console.log('checkpoint');\nexpect(x).toHaveBeenCalledExactlyOnceWith('a');",
      errors: [
        {
          messageId: 'preferCalledExactlyOnceWith',
          message:
            'Using `toHaveBeenCalledOnce` and `toHaveBeenCalledWith` on the same target; prefer `toHaveBeenCalledExactlyOnceWith` instead.',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 31,
        },
      ],
    },
    {
      code: `expect(x).toHaveBeenCalledOnce();
var hoge = 'foo';
expect(x).toHaveBeenCalledWith('a');`,
      output:
        "var hoge = 'foo';\nexpect(x).toHaveBeenCalledExactlyOnceWith('a');",
      errors: [
        {
          messageId: 'preferCalledExactlyOnceWith',
          message:
            'Using `toHaveBeenCalledOnce` and `toHaveBeenCalledWith` on the same target; prefer `toHaveBeenCalledExactlyOnceWith` instead.',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 31,
        },
      ],
    },
    {
      code: `expect(x).toHaveBeenCalledOnce();
expect(y).toEqual({ ...obj });
expect(x).toHaveBeenCalledWith('a');`,
      output:
        "expect(y).toEqual({ ...obj });\nexpect(x).toHaveBeenCalledExactlyOnceWith('a');",
      errors: [
        {
          messageId: 'preferCalledExactlyOnceWith',
          message:
            'Using `toHaveBeenCalledOnce` and `toHaveBeenCalledWith` on the same target; prefer `toHaveBeenCalledExactlyOnceWith` instead.',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 31,
        },
      ],
    },
    {
      code: `expect(x).toHaveBeenCalledOnce();
expect(y).toEqual({ id: 1 });
expect(x).toHaveBeenCalledWith('a');`,
      output:
        "expect(y).toEqual({ id: 1 });\nexpect(x).toHaveBeenCalledExactlyOnceWith('a');",
      errors: [
        {
          messageId: 'preferCalledExactlyOnceWith',
          message:
            'Using `toHaveBeenCalledOnce` and `toHaveBeenCalledWith` on the same target; prefer `toHaveBeenCalledExactlyOnceWith` instead.',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 31,
        },
      ],
    },
    {
      code: `expect(x).toHaveBeenCalledOnce();
expect(y).toHaveBeenCalledWith(expect.any(String));
expect(x).toHaveBeenCalledWith('a');`,
      output:
        "expect(y).toHaveBeenCalledWith(expect.any(String));\nexpect(x).toHaveBeenCalledExactlyOnceWith('a');",
      errors: [
        {
          messageId: 'preferCalledExactlyOnceWith',
          message:
            'Using `toHaveBeenCalledOnce` and `toHaveBeenCalledWith` on the same target; prefer `toHaveBeenCalledExactlyOnceWith` instead.',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 31,
        },
      ],
    },
    {
      code: `async function f() {
  await expect(x).resolves.toHaveBeenCalledOnce();
  await expect(x).resolves.toHaveBeenCalledWith('a');
}`,
      output: `async function f() {
  await expect(x).resolves.toHaveBeenCalledExactlyOnceWith('a');
}`,
      errors: [
        {
          messageId: 'preferCalledExactlyOnceWith',
          message:
            'Using `toHaveBeenCalledOnce` and `toHaveBeenCalledWith` on the same target; prefer `toHaveBeenCalledExactlyOnceWith` instead.',
          line: 3,
          column: 28,
          endLine: 3,
          endColumn: 48,
        },
      ],
    },
  ],
});
