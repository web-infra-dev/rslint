import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-to-contain', {} as never, {
  valid: [
    { code: 'expect.hasAssertions' },
    { code: 'expect.hasAssertions()' },
    { code: 'expect.assertions(1)' },
    { code: 'expect().toBe(false);' },
    { code: 'expect(a).toContain(b);' },
    { code: "expect(a.name).toBe('b');" },
    { code: `expect(a).toBe(true);` },
    { code: `expect(a).toEqual(b)` },
    { code: `expect(a.test(c)).toEqual(b)` },
    { code: `expect(a.includes(b)).toEqual()` },
    { code: `expect(a.includes(b)).toEqual("test")` },
    { code: `expect(a.includes(b)).toBe("test")` },
    { code: `expect(a.includes()).toEqual()` },
    { code: `expect(a.includes()).toEqual(true)` },
    { code: `expect(a.includes(b,c)).toBe(true)` },
    { code: `expect([{a:1}]).toContain({a:1})` },
    { code: `expect([1].includes(1)).toEqual` },
    { code: `expect([1].includes).toEqual` },
    { code: `expect([1].includes).not` },
    { code: `expect(a.test(b)).resolves.toEqual(true)` },
    { code: `expect(a.test(b)).resolves.not.toEqual(true)` },
    { code: `expect(a).not.toContain(b)` },
    { code: `expect(a.includes(...[])).toBe(true)` },
    { code: `expect(a.includes(b)).toBe(...true)` },
    { code: `expect(a);` },
    {
      code: "(expect('Model must be bound to an array if the multiple property is true') as any).toHaveBeenTipped()",
    },
    { code: 'expect(a.includes(b)).toEqual(0 as boolean);' },
  ],
  invalid: [
    {
      code: 'expect(a.includes(b)).toEqual(true);',
      output: 'expect(a).toContain(b);',
      errors: [{ messageId: 'useToContain', column: 23, line: 1 }],
    },
    {
      code: 'expect(a.includes(b,),).toEqual(true,);',
      output: 'expect(a,).toContain(b,);',
      errors: [{ messageId: 'useToContain', column: 25, line: 1 }],
    },
    {
      code: "expect(a['includes'](b)).toEqual(true);",
      output: 'expect(a).toContain(b);',
      errors: [{ messageId: 'useToContain', column: 26, line: 1 }],
    },
    {
      code: "expect(a['includes'](b))['toEqual'](true);",
      output: 'expect(a).toContain(b);',
      errors: [{ messageId: 'useToContain', column: 26, line: 1 }],
    },
    {
      code: "expect(a['includes'](b)).toEqual(false);",
      output: 'expect(a).not.toContain(b);',
      errors: [{ messageId: 'useToContain', column: 26, line: 1 }],
    },
    {
      code: "expect(a['includes'](b)).not.toEqual(false);",
      output: 'expect(a).toContain(b);',
      errors: [{ messageId: 'useToContain', column: 30, line: 1 }],
    },
    {
      code: "expect(a['includes'](b))['not'].toEqual(false);",
      output: 'expect(a).toContain(b);',
      errors: [{ messageId: 'useToContain', column: 33, line: 1 }],
    },
    {
      code: "expect(a['includes'](b))['not']['toEqual'](false);",
      output: 'expect(a).toContain(b);',
      errors: [{ messageId: 'useToContain', column: 33, line: 1 }],
    },
    {
      code: 'expect(a.includes(b)).toEqual(false);',
      output: 'expect(a).not.toContain(b);',
      errors: [{ messageId: 'useToContain', column: 23, line: 1 }],
    },
    {
      code: 'expect(a.includes(b)).not.toEqual(false);',
      output: 'expect(a).toContain(b);',
      errors: [{ messageId: 'useToContain', column: 27, line: 1 }],
    },
    {
      code: 'expect(a.includes(b)).not.toEqual(true);',
      output: 'expect(a).not.toContain(b);',
      errors: [{ messageId: 'useToContain', column: 27, line: 1 }],
    },
    {
      code: 'expect(a.includes(b)).toBe(true);',
      output: 'expect(a).toContain(b);',
      errors: [{ messageId: 'useToContain', column: 23, line: 1 }],
    },
    {
      code: 'expect(a.includes(b)).toBe(false);',
      output: 'expect(a).not.toContain(b);',
      errors: [{ messageId: 'useToContain', column: 23, line: 1 }],
    },
    {
      code: 'expect(a.includes(b)).not.toBe(false);',
      output: 'expect(a).toContain(b);',
      errors: [{ messageId: 'useToContain', column: 27, line: 1 }],
    },
    {
      code: 'expect(a.includes(b)).not.toBe(true);',
      output: 'expect(a).not.toContain(b);',
      errors: [{ messageId: 'useToContain', column: 27, line: 1 }],
    },
    {
      code: 'expect(a.includes(b)).toStrictEqual(true);',
      output: 'expect(a).toContain(b);',
      errors: [{ messageId: 'useToContain', column: 23, line: 1 }],
    },
    {
      code: 'expect(a.includes(b)).toStrictEqual(false);',
      output: 'expect(a).not.toContain(b);',
      errors: [{ messageId: 'useToContain', column: 23, line: 1 }],
    },
    {
      code: 'expect(a.includes(b)).not.toStrictEqual(false);',
      output: 'expect(a).toContain(b);',
      errors: [{ messageId: 'useToContain', column: 27, line: 1 }],
    },
    {
      code: 'expect(a.includes(b)).not.toStrictEqual(true);',
      output: 'expect(a).not.toContain(b);',
      errors: [{ messageId: 'useToContain', column: 27, line: 1 }],
    },
    {
      code: 'expect(a.test(t).includes(b.test(p))).toEqual(true);',
      output: 'expect(a.test(t)).toContain(b.test(p));',
      errors: [{ messageId: 'useToContain', column: 39, line: 1 }],
    },
    {
      code: 'expect(a.test(t).includes(b.test(p))).toEqual(false);',
      output: 'expect(a.test(t)).not.toContain(b.test(p));',
      errors: [{ messageId: 'useToContain', column: 39, line: 1 }],
    },
    {
      code: 'expect(a.test(t).includes(b.test(p))).not.toEqual(true);',
      output: 'expect(a.test(t)).not.toContain(b.test(p));',
      errors: [{ messageId: 'useToContain', column: 43, line: 1 }],
    },
    {
      code: 'expect(a.test(t).includes(b.test(p))).not.toEqual(false);',
      output: 'expect(a.test(t)).toContain(b.test(p));',
      errors: [{ messageId: 'useToContain', column: 43, line: 1 }],
    },
    {
      code: 'expect([{a:1}].includes({a:1})).toBe(true);',
      output: 'expect([{a:1}]).toContain({a:1});',
      errors: [{ messageId: 'useToContain', column: 33, line: 1 }],
    },
    {
      code: 'expect([{a:1}].includes({a:1})).toBe(false);',
      output: 'expect([{a:1}]).not.toContain({a:1});',
      errors: [{ messageId: 'useToContain', column: 33, line: 1 }],
    },
    {
      code: 'expect([{a:1}].includes({a:1})).not.toBe(true);',
      output: 'expect([{a:1}]).not.toContain({a:1});',
      errors: [{ messageId: 'useToContain', column: 37, line: 1 }],
    },
    {
      code: 'expect([{a:1}].includes({a:1})).not.toBe(false);',
      output: 'expect([{a:1}]).toContain({a:1});',
      errors: [{ messageId: 'useToContain', column: 37, line: 1 }],
    },
    {
      code: 'expect([{a:1}].includes({a:1})).toStrictEqual(true);',
      output: 'expect([{a:1}]).toContain({a:1});',
      errors: [{ messageId: 'useToContain', column: 33, line: 1 }],
    },
    {
      code: 'expect([{a:1}].includes({a:1})).toStrictEqual(false);',
      output: 'expect([{a:1}]).not.toContain({a:1});',
      errors: [{ messageId: 'useToContain', column: 33, line: 1 }],
    },
    {
      code: 'expect([{a:1}].includes({a:1})).not.toStrictEqual(true);',
      output: 'expect([{a:1}]).not.toContain({a:1});',
      errors: [{ messageId: 'useToContain', column: 37, line: 1 }],
    },
    {
      code: 'expect([{a:1}].includes({a:1})).not.toStrictEqual(false);',
      output: 'expect([{a:1}]).toContain({a:1});',
      errors: [{ messageId: 'useToContain', column: 37, line: 1 }],
    },
    {
      code: `
        import { expect as pleaseExpect } from '@jest/globals';

        pleaseExpect([{a:1}].includes({a:1})).not.toStrictEqual(false);
      `,
      output: `
        import { expect as pleaseExpect } from '@jest/globals';

        pleaseExpect([{a:1}]).toContain({a:1});
      `,
      errors: [{ messageId: 'useToContain', column: 43, line: 3 }],
    },
    {
      code: 'expect(a.includes(b)).toEqual(false as boolean);',
      output: 'expect(a).not.toContain(b);',
      errors: [{ messageId: 'useToContain', column: 23, line: 1 }],
    },
    {
      code: 'expect(a.includes(b)).resolves.toBe(true);',
      output: 'expect(a).resolves.toContain(b);',
      errors: [{ messageId: 'useToContain', column: 32, line: 1 }],
    },
    {
      code: 'expect(a.includes(b)).resolves.not.toBe(true);',
      output: 'expect(a).resolves.not.toContain(b);',
      errors: [{ messageId: 'useToContain', column: 36, line: 1 }],
    },
  ],
});
