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
      code: `expect(x).toHaveBeenCalledOnce(); if (c) { x.mockClear(); } expect(x).toHaveBeenCalledWith('hoge')`,
    },
    {
      code: `expect(obj.fn).toHaveBeenCalledOnce(); obj.fn.mockRestore(); expect(obj.fn).toHaveBeenCalledWith('hoge')`,
    },
    {
      code: `expect.soft(x).toHaveBeenCalledOnce(); expect(x).toHaveBeenCalledWith('hoge')`,
    },
    {
      code: `import { expect } from 'vitest'; expect(x).toHaveBeenCalledOnce(); expect(x).toHaveBeenCalledWith('a')`,
    },
    { code: `expect(x).to.have.been.calledOnceWith('a')` },
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
  ],
});
