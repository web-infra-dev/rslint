import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-called-exactly-once-with', {} as never, {
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
      errors: [{ messageId: 'preferCalledExactlyOnceWith' }],
    },
    {
      code: `expect(x).toHaveBeenCalledWith<[string, number]>('hoge', /* keep */ 123);
expect(x).toHaveBeenCalledOnce();`,
      output: `expect(x).toHaveBeenCalledExactlyOnceWith<[string, number]>('hoge', /* keep */ 123);
`,
      errors: [{ messageId: 'preferCalledExactlyOnceWith' }],
    },
    {
      code: `test('x', ({ expect }) => {
  expect(x).toHaveBeenCalledOnce();
  expect(x).toHaveBeenCalledWith('a');
});`,
      output: `test('x', ({ expect }) => {
  expect(x).toHaveBeenCalledExactlyOnceWith('a');
});`,
      errors: [{ messageId: 'preferCalledExactlyOnceWith' }],
    },
    {
      code: `import { expect as check } from '@rstest/core';
check(x).toHaveBeenCalledOnce();
check(x).toHaveBeenCalledWith('a');`,
      errors: [{ messageId: 'preferCalledExactlyOnceWith' }],
    },
    {
      code: `import.meta.rstest.expect(x).toHaveBeenCalledOnce();
import.meta.rstest.expect(x).toHaveBeenCalledWith('a');`,
      errors: [{ messageId: 'preferCalledExactlyOnceWith' }],
    },
    {
      code: `expect(x).to.have.been.calledOnce;
expect(x).to.have.been.calledWith('a');`,
      output: `expect(x).to.have.been.calledOnceWith('a');`,
      errors: [{ messageId: 'preferCalledExactlyOnceWith' }],
    },
    {
      code: `expect(x).to.have.been.calledOnce.and.calledWith('a');`,
      errors: [{ messageId: 'preferCalledExactlyOnceWith' }],
    },
    {
      code: `expect(x).to.have.been.calledOnce.and.to.be.ok;
expect(x).to.have.been.calledWith('a');`,
      errors: [{ messageId: 'preferCalledExactlyOnceWith' }],
    },
    {
      code: `expect(getMock()).toHaveBeenCalledOnce(); expect(getMock()).toHaveBeenCalledWith('a');`,
      output: null,
      errors: [{ messageId: 'preferCalledExactlyOnceWith' }],
    },
    {
      code: `expect(mocks['a']).toHaveBeenCalledOnce(); expect(mocks['a']).toHaveBeenCalledWith('x');`,
      output: ` expect(mocks['a']).toHaveBeenCalledExactlyOnceWith('x');`,
      errors: [{ messageId: 'preferCalledExactlyOnceWith' }],
    },
    {
      code: `expect(x).toHaveBeenCalledOnce();
console.log('checkpoint');
expect(x).toHaveBeenCalledWith('a');`,
      output: `console.log('checkpoint');
expect(x).toHaveBeenCalledExactlyOnceWith('a');
`,
      errors: [{ messageId: 'preferCalledExactlyOnceWith' }],
    },
    {
      code: `expect(x).toHaveBeenCalledOnce();
var hoge = 'foo';
expect(x).toHaveBeenCalledWith('a');`,
      output: `var hoge = 'foo';
expect(x).toHaveBeenCalledExactlyOnceWith('a');
`,
      errors: [{ messageId: 'preferCalledExactlyOnceWith' }],
    },
    {
      code: `expect(x).toHaveBeenCalledOnce();
expect(y).toEqual({ id: 1 });
expect(x).toHaveBeenCalledWith('a');`,
      output: `expect(y).toEqual({ id: 1 });
expect(x).toHaveBeenCalledExactlyOnceWith('a');
`,
      errors: [{ messageId: 'preferCalledExactlyOnceWith' }],
    },
    {
      code: `expect(x).toHaveBeenCalledOnce();
expect(y).toHaveBeenCalledWith(expect.any(String));
expect(x).toHaveBeenCalledWith('a');`,
      output: `expect(y).toHaveBeenCalledWith(expect.any(String));
expect(x).toHaveBeenCalledExactlyOnceWith('a');
`,
      errors: [{ messageId: 'preferCalledExactlyOnceWith' }],
    },
    {
      code: `async function f() {
  await expect(x).resolves.toHaveBeenCalledOnce();
  await expect(x).resolves.toHaveBeenCalledWith('a');
}`,
      output: `async function f() {
  await expect(x).resolves.toHaveBeenCalledExactlyOnceWith('a');
}`,
      errors: [{ messageId: 'preferCalledExactlyOnceWith' }],
    },
  ],
});
