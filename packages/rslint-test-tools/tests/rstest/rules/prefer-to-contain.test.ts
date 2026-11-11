import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('prefer-to-contain', {
  valid: [
    { code: 'expect.hasAssertions' },
    { code: 'expect.hasAssertions()' },
    { code: 'expect.assertions(1)' },
    { code: 'expect().toBe(false);' },
    { code: 'expect(a).toContain(b);' },
    { code: "expect(a.name).toBe('b');" },
    { code: 'expect(a).toBe(true);' },
    { code: 'expect(a).toEqual(b)' },
    { code: 'expect(a.test(c)).toEqual(b)' },
    { code: 'expect(a.includes(b)).toEqual()' },
    { code: 'expect(a.includes(b)).toEqual("test")' },
    { code: 'expect(a.includes(b)).toBe("test")' },
    { code: 'expect(a.includes()).toEqual()' },
    { code: 'expect(a.includes()).toEqual(true)' },
    { code: 'expect(a.includes(b,c)).toBe(true)' },
    { code: 'expect([{a:1}]).toContain({a:1})' },
    { code: 'expect([1].includes(1)).toEqual' },
    { code: 'expect([1].includes).toEqual' },
    { code: 'expect([1].includes).not' },
    { code: 'expect(a.test(b)).resolves.toEqual(true)' },
    { code: 'expect(a.test(b)).resolves.not.toEqual(true)' },
    { code: 'expect(a).not.toContain(b)' },
    { code: 'expect(a.includes(...[])).toBe(true)' },
    { code: 'expect(a.includes(b)).toBe(...true)' },
    { code: 'expect(a);' },
    { code: 'expect(a).to.be.a("string");' },
    {
      code: 'expect([{ a: 1 }].includes({ a: 1 })).deep.toBe(false);',
    },
    { code: 'expect(a.includes(b)).toBe(true).toBe(true);' },
    { code: 'expect(a.includes(b)).toBe(true).to.be.ok;' },
    { code: 'expect(a.includes(b)).toBe(true).then(done);' },
  ],
  invalid: [
    {
      code: 'expect(a.includes(b)).toEqual(true);',
      output: 'expect(a).toContain(b);',
      errors: [
        {
          messageId: 'useToContain',
          message: 'Use toContain() instead',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: 'expect(a.includes(b,),).toEqual(true,);',
      output: 'expect(a,).toContain(b,);',
      errors: [
        {
          messageId: 'useToContain',
          message: 'Use toContain() instead',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: "expect(a['includes'](b)).toEqual(true);",
      output: 'expect(a).toContain(b);',
      errors: [
        {
          messageId: 'useToContain',
          message: 'Use toContain() instead',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: "expect(a['includes'](b))['toEqual'](true);",
      output: "expect(a)['toContain'](b);",
      errors: [
        {
          messageId: 'useToContain',
          message: 'Use toContain() instead',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: "expect(a['includes'](b)).toEqual(false);",
      output: 'expect(a).not.toContain(b);',
      errors: [
        {
          messageId: 'useToContain',
          message: 'Use toContain() instead',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: "expect(a['includes'](b)).not.toEqual(false);",
      output: 'expect(a).toContain(b);',
      errors: [
        {
          messageId: 'useToContain',
          message: 'Use toContain() instead',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: "expect(a['includes'](b))['not']['toEqual'](false);",
      output: "expect(a)['toContain'](b);",
      errors: [
        {
          messageId: 'useToContain',
          message: 'Use toContain() instead',
          line: 1,
          column: 33,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      code: 'expect(a.includes(b)).toEqual(false);',
      output: 'expect(a).not.toContain(b);',
      errors: [
        {
          messageId: 'useToContain',
          message: 'Use toContain() instead',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: 'expect(a.includes(b)).not.toEqual(false);',
      output: 'expect(a).toContain(b);',
      errors: [
        {
          messageId: 'useToContain',
          message: 'Use toContain() instead',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: 'expect(a.includes(b)).not.toEqual(true);',
      output: 'expect(a).not.toContain(b);',
      errors: [
        {
          messageId: 'useToContain',
          message: 'Use toContain() instead',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: 'expect(a.includes(b)).toBe(true);',
      output: 'expect(a).toContain(b);',
      errors: [
        {
          messageId: 'useToContain',
          message: 'Use toContain() instead',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: 'expect(a.includes(b)).toBe(false);',
      output: 'expect(a).not.toContain(b);',
      errors: [
        {
          messageId: 'useToContain',
          message: 'Use toContain() instead',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: 'expect(a.includes(b)).not.toBe(false);',
      output: 'expect(a).toContain(b);',
      errors: [
        {
          messageId: 'useToContain',
          message: 'Use toContain() instead',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: 'expect(a.includes(b)).not.toBe(true);',
      output: 'expect(a).not.toContain(b);',
      errors: [
        {
          messageId: 'useToContain',
          message: 'Use toContain() instead',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: 'expect<boolean>(a.includes(b)).toBe(true);',
      output: 'expect(a).toContain(b);',
      errors: [
        {
          messageId: 'useToContain',
          message: 'Use toContain() instead',
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      code: 'expect.soft<boolean>(a.includes(b)).toBe<boolean>(true);',
      output: 'expect.soft(a).toContain(b);',
      errors: [
        {
          messageId: 'useToContain',
          message: 'Use toContain() instead',
          line: 1,
          column: 37,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      code: 'expect<boolean>(a.includes(b)).ordered.toBe(true);',
      output: 'expect(a).ordered.toContain(b);',
      errors: [
        {
          messageId: 'useToContain',
          message: 'Use toContain() instead',
          line: 1,
          column: 40,
          endLine: 1,
          endColumn: 44,
        },
      ],
    },
    {
      code: 'expect(a.includes(b)).toStrictEqual(true);',
      output: 'expect(a).toContain(b);',
      errors: [
        {
          messageId: 'useToContain',
          message: 'Use toContain() instead',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      code: 'expect(a.includes(b)).toStrictEqual(false);',
      output: 'expect(a).not.toContain(b);',
      errors: [
        {
          messageId: 'useToContain',
          message: 'Use toContain() instead',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      code: 'expect(a.includes(b)).not.toStrictEqual(false);',
      output: 'expect(a).toContain(b);',
      errors: [
        {
          messageId: 'useToContain',
          message: 'Use toContain() instead',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      code: 'expect(a.includes(b)).not.toStrictEqual(true);',
      output: 'expect(a).not.toContain(b);',
      errors: [
        {
          messageId: 'useToContain',
          message: 'Use toContain() instead',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      code: 'expect(a.test(t).includes(b.test(p))).toEqual(true);',
      errors: [
        {
          messageId: 'useToContain',
          message: 'Use toContain() instead',
          line: 1,
          column: 39,
          endLine: 1,
          endColumn: 46,
        },
      ],
    },
    {
      code: 'expect(a.test(t).includes(b.test(p))).not.toEqual(false);',
      errors: [
        {
          messageId: 'useToContain',
          message: 'Use toContain() instead',
          line: 1,
          column: 43,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
    {
      code: 'expect([{a:1}].includes({a:1})).toBe(true);',
      output: 'expect([{a:1}]).toContain({a:1});',
      errors: [
        {
          messageId: 'useToContain',
          message: 'Use toContain() instead',
          line: 1,
          column: 33,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: 'expect([{a:1}].includes({a:1})).not.toStrictEqual(false);',
      output: 'expect([{a:1}]).toContain({a:1});',
      errors: [
        {
          messageId: 'useToContain',
          message: 'Use toContain() instead',
          line: 1,
          column: 37,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
  ],
});
