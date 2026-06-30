import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-spy-on', {} as never, {
  valid: [
    { code: 'Date.now = () => 10' },
    { code: 'window.fetch = jest.fn' },
    { code: 'Date.now = fn()' },
    { code: 'obj.mock = jest.something()' },
    { code: 'const mock = jest.fn()' },
    { code: 'mock = jest.fn()' },
    { code: 'const mockObj = { mock: jest.fn() }' },
    { code: 'mockObj = { mock: jest.fn() }' },
    { code: 'window[`${name}`] = jest[`fn${expression}`]()' },
    { code: 'class A { #f; m() { this.#f = jest.fn(); } }' },
  ],
  invalid: [
    {
      code: 'obj.a = jest.fn(); const test = 10;',
      output: "jest.spyOn(obj, 'a').mockImplementation(); const test = 10;",
      errors: [
        {
          messageId: 'useJestSpyOn',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: "Date['now'] = jest['fn']()",
      output: "jest.spyOn(Date, 'now').mockImplementation()",
      errors: [
        {
          messageId: 'useJestSpyOn',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: 'window[`${name}`] = jest[`fn`]()',
      output: 'jest.spyOn(window, `${name}`).mockImplementation()',
      errors: [
        {
          messageId: 'useJestSpyOn',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: "obj['prop' + 1] = jest['fn']()",
      output: "jest.spyOn(obj, 'prop' + 1).mockImplementation()",
      errors: [
        {
          messageId: 'useJestSpyOn',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: 'obj.one.two = jest.fn(); const test = 10;',
      output:
        "jest.spyOn(obj.one, 'two').mockImplementation(); const test = 10;",
      errors: [
        {
          messageId: 'useJestSpyOn',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'obj.a = jest.fn(() => 10,)',
      output: "jest.spyOn(obj, 'a').mockImplementation(() => 10)",
      errors: [
        {
          messageId: 'useJestSpyOn',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: "obj.a.b = jest.fn(() => ({})).mockReturnValue('default').mockReturnValueOnce('first call'); test();",
      output:
        "jest.spyOn(obj.a, 'b').mockImplementation(() => ({})).mockReturnValue('default').mockReturnValueOnce('first call'); test();",
      errors: [
        {
          messageId: 'useJestSpyOn',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 91,
        },
      ],
    },
    {
      code: 'window.fetch = jest.fn(() => ({})).one.two().three().four',
      output:
        "jest.spyOn(window, 'fetch').mockImplementation(() => ({})).one.two().three().four",
      errors: [
        {
          messageId: 'useJestSpyOn',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 58,
        },
      ],
    },
    {
      // https://github.com/jest-community/eslint-plugin-jest/issues/1304
      code: 'foo[bar] = jest.fn().mockReturnValue(undefined)',
      output:
        'jest.spyOn(foo, bar).mockImplementation().mockReturnValue(undefined)',
      errors: [
        {
          messageId: 'useJestSpyOn',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },
    {
      // https://github.com/jest-community/eslint-plugin-jest/issues/1307
      code: `
        foo.bar = jest.fn().mockImplementation(baz => baz)
        foo.bar = jest.fn(a => b).mockImplementation(baz => baz)
      `,
      output: `
        jest.spyOn(foo, 'bar').mockImplementation(baz => baz)
        jest.spyOn(foo, 'bar').mockImplementation(baz => baz)
      `,
      errors: [
        {
          messageId: 'useJestSpyOn',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 59,
        },
        {
          messageId: 'useJestSpyOn',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 65,
        },
      ],
    },
    {
      code: 'foo.bar = (jest.fn()).mockImplementation(baz => baz)',
      output: "jest.spyOn(foo, 'bar').mockImplementation(baz => baz)",
      errors: [
        {
          messageId: 'useJestSpyOn',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
    {
      code: 'foo.bar = (jest.fn(a => b)).mockImplementation(baz => baz)',
      output: "jest.spyOn(foo, 'bar').mockImplementation(baz => baz)",
      errors: [
        {
          messageId: 'useJestSpyOn',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 59,
        },
      ],
    },
    {
      code: 'obj.a = (jest.fn().mockReturnValue(1))',
      output: "jest.spyOn(obj, 'a').mockImplementation().mockReturnValue(1)",
      errors: [
        {
          messageId: 'useJestSpyOn',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      code: 'obj.a = (jest.fn())',
      output: "jest.spyOn(obj, 'a').mockImplementation()",
      errors: [
        {
          messageId: 'useJestSpyOn',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: 'obj.a = ((jest.fn()))',
      output: "jest.spyOn(obj, 'a').mockImplementation()",
      errors: [
        {
          messageId: 'useJestSpyOn',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'obj.a = (jest.fn(() => 10))',
      output: "jest.spyOn(obj, 'a').mockImplementation(() => 10)",
      errors: [
        {
          messageId: 'useJestSpyOn',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: 'obj.a = (jest.fn()).one.two()',
      output: "jest.spyOn(obj, 'a').mockImplementation().one.two()",
      errors: [
        {
          messageId: 'useJestSpyOn',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
  ],
});
