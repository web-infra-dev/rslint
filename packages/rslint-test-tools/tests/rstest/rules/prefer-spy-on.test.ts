import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-spy-on', {} as never, {
  valid: [
    { code: 'Date.now = () => 10' },
    { code: 'window.fetch = rs.fn' },
    { code: 'Date.now = fn()' },
    { code: 'obj.mock = rs.something()' },
    { code: 'const mock = rs.fn()' },
    { code: 'mock = rs.fn()' },
    { code: 'const mockObj = { mock: rs.fn() }' },
    { code: 'mockObj = { mock: rs.fn() }' },
    { code: 'window[`${name}`] = rs[`fn${expression}`]()' },
  ],
  invalid: [
    {
      code: 'obj.a = rs.fn(); const test = 10;',
      output: `rs.spyOn(obj, 'a').mockImplementation(() => undefined); const test = 10;`,
      errors: [
        {
          messageId: 'useRsSpyOn',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: `Date['now'] = rs['fn']()`,
      output: `rs.spyOn(Date, 'now').mockImplementation(() => undefined)`,
      errors: [
        {
          messageId: 'useRsSpyOn',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: 'window[`${name}`] = rs[`fn`]()',
      output: 'rs.spyOn(window, `${name}`).mockImplementation(() => undefined)',
      errors: [
        {
          messageId: 'useRsSpyOn',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: `obj['prop' + 1] = rs['fn']()`,
      output: `rs.spyOn(obj, 'prop' + 1).mockImplementation(() => undefined)`,
      errors: [
        {
          messageId: 'useRsSpyOn',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: 'obj.one.two = rs.fn(); const test = 10;',
      output: `rs.spyOn(obj.one, 'two').mockImplementation(() => undefined); const test = 10;`,
      errors: [
        {
          messageId: 'useRsSpyOn',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'obj.a = rs.fn(() => 10,)',
      output: `rs.spyOn(obj, 'a').mockImplementation(() => 10)`,
      errors: [
        {
          messageId: 'useRsSpyOn',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: `obj.a.b = rs.fn(() => ({})).mockReturnValue('default').mockReturnValueOnce('first call'); test();`,
      output: `rs.spyOn(obj.a, 'b').mockImplementation(() => ({})).mockReturnValue('default').mockReturnValueOnce('first call'); test();`,
      errors: [
        {
          messageId: 'useRsSpyOn',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 89,
        },
      ],
    },
    {
      code: 'window.fetch = rs.fn(() => ({})).one.two().three().four',
      output: `rs.spyOn(window, 'fetch').mockImplementation(() => ({})).one.two().three().four`,
      errors: [
        {
          messageId: 'useRsSpyOn',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 56,
        },
      ],
    },
    {
      code: 'foo[bar] = rs.fn().mockReturnValue(undefined)',
      output:
        'rs.spyOn(foo, bar).mockImplementation(() => undefined).mockReturnValue(undefined)',
      errors: [
        {
          messageId: 'useRsSpyOn',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 46,
        },
      ],
    },
    {
      code: `
        foo.bar = rs.fn().mockImplementation(baz => baz)
        foo.bar = rs.fn(a => b).mockImplementation(baz => baz)
      `,
      output: `
        rs.spyOn(foo, 'bar').mockImplementation(baz => baz)
        rs.spyOn(foo, 'bar').mockImplementation(baz => baz)
      `,
      errors: [
        {
          messageId: 'useRsSpyOn',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 57,
        },
        {
          messageId: 'useRsSpyOn',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 63,
        },
      ],
    },
  ],
});
