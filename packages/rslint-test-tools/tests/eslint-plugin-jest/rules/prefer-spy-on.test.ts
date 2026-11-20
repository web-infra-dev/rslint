// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('prefer-spy-on', {
  valid: [
    ...(
      [
        'Date.now = () => 10',
        'window.fetch = jest.fn',
        'Date.now = fn()',
        'obj.mock = jest.something()',
        'const mock = jest.fn()',
        'mock = jest.fn()',
        'const mockObj = { mock: jest.fn() }',
        'mockObj = { mock: jest.fn() }',
        'window[`${name}`] = jest[`fn${expression}`]()',
        'class A { #f; m() { this.#f = jest.fn(); } }',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    ...(
      [
        'Date.now = () => 10',
        'window.fetch = jest.fn',
        'Date.now = fn()',
        'obj.mock = jest.something()',
        'const mock = jest.fn()',
        'mock = jest.fn()',
        'const mockObj = { mock: jest.fn() }',
        'mockObj = { mock: jest.fn() }',
        'window[`${name}`] = jest[`fn${expression}`]()',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream prefer-spy-on / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
    })),
  ],
  invalid: [
    ...(
      [
        [
          'obj.a = jest.fn(); const test = 10;',
          18,
          "jest.spyOn(obj, 'a').mockImplementation(); const test = 10;",
        ],
        [
          "Date['now'] = jest['fn']()",
          27,
          "jest.spyOn(Date, 'now').mockImplementation()",
        ],
        [
          'window[`${name}`] = jest[`fn`]()',
          33,
          'jest.spyOn(window, `${name}`).mockImplementation()',
        ],
        [
          "obj['prop' + 1] = jest['fn']()",
          31,
          "jest.spyOn(obj, 'prop' + 1).mockImplementation()",
        ],
        [
          'obj.one.two = jest.fn(); const test = 10;',
          24,
          "jest.spyOn(obj.one, 'two').mockImplementation(); const test = 10;",
        ],
        [
          'obj.a = jest.fn(() => 10,)',
          27,
          "jest.spyOn(obj, 'a').mockImplementation(() => 10)",
        ],
        [
          "obj.a.b = jest.fn(() => ({})).mockReturnValue('default').mockReturnValueOnce('first call'); test();",
          91,
          "jest.spyOn(obj.a, 'b').mockImplementation(() => ({})).mockReturnValue('default').mockReturnValueOnce('first call'); test();",
        ],
        [
          'window.fetch = jest.fn(() => ({})).one.two().three().four',
          58,
          "jest.spyOn(window, 'fetch').mockImplementation(() => ({})).one.two().three().four",
        ],
        [
          'foo[bar] = jest.fn().mockReturnValue(undefined)',
          48,
          'jest.spyOn(foo, bar).mockImplementation().mockReturnValue(undefined)',
        ],
      ] as const
    ).map(([code, endColumn, output], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'useJestSpyOn',
          message: 'Use jest.spyOn() instead',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    {
      name: 'run 0 invalid 9',
      code: `
        foo.bar = jest.fn().mockImplementation(baz => baz)
        foo.bar = jest.fn(a => b).mockImplementation(baz => baz)
     \u0020`,
      errors: [
        {
          messageId: 'useJestSpyOn',
          message: 'Use jest.spyOn() instead',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 59,
        },
        {
          messageId: 'useJestSpyOn',
          message: 'Use jest.spyOn() instead',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 65,
        },
      ],
      output: `
        jest.spyOn(foo, 'bar').mockImplementation(baz => baz)
        jest.spyOn(foo, 'bar').mockImplementation(baz => baz)
     \u0020`,
    },
    ...(
      [
        [
          'foo.bar = (jest.fn()).mockImplementation(baz => baz)',
          53,
          "jest.spyOn(foo, 'bar').mockImplementation(baz => baz)",
        ],
        [
          'foo.bar = (jest.fn(a => b)).mockImplementation(baz => baz)',
          59,
          "jest.spyOn(foo, 'bar').mockImplementation(baz => baz)",
        ],
        [
          'obj.a = (jest.fn().mockReturnValue(1))',
          39,
          "jest.spyOn(obj, 'a').mockImplementation().mockReturnValue(1)",
        ],
        [
          'obj.a = (jest.fn())',
          20,
          "jest.spyOn(obj, 'a').mockImplementation()",
        ],
        [
          'obj.a = ((jest.fn()))',
          22,
          "jest.spyOn(obj, 'a').mockImplementation()",
        ],
        [
          'obj.a = (jest.fn(() => 10))',
          28,
          "jest.spyOn(obj, 'a').mockImplementation(() => 10)",
        ],
        [
          'obj.a = (jest.fn()).one.two()',
          30,
          "jest.spyOn(obj, 'a').mockImplementation().one.two()",
        ],
      ] as const
    ).map(([code, endColumn, output], index): TestCase => ({
      name: `run 0 invalid ${index + 10}`,
      code,
      errors: [
        {
          messageId: 'useJestSpyOn',
          message: 'Use jest.spyOn() instead',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    ...(
      [
        [
          'obj.a = jest.fn(); const test = 10;',
          18,
          "jest.spyOn(obj, 'a').mockImplementation(); const test = 10;",
        ],
        [
          "Date['now'] = jest['fn']()",
          27,
          "jest.spyOn(Date, 'now').mockImplementation()",
        ],
        [
          'window[`${name}`] = jest[`fn`]()',
          33,
          'jest.spyOn(window, `${name}`).mockImplementation()',
        ],
        [
          "obj['prop' + 1] = jest['fn']()",
          31,
          "jest.spyOn(obj, 'prop' + 1).mockImplementation()",
        ],
        [
          'obj.one.two = jest.fn(); const test = 10;',
          24,
          "jest.spyOn(obj.one, 'two').mockImplementation(); const test = 10;",
        ],
      ] as const
    ).map(([code, endColumn, output], index): TestCase => ({
      name: `upstream prefer-spy-on / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'useJestSpyOn',
          message: 'Use jest.spyOn() instead',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    {
      name: 'upstream prefer-spy-on / invalid 5',
      code: 'obj.a = jest.fn(() => 10,)',
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'useJestSpyOn',
          message: 'Use jest.spyOn() instead',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
      output: "jest.spyOn(obj, 'a').mockImplementation(() => 10)",
    },
    ...(
      [
        [
          "obj.a.b = jest.fn(() => ({})).mockReturnValue('default').mockReturnValueOnce('first call'); test();",
          91,
          "jest.spyOn(obj.a, 'b').mockImplementation(() => ({})).mockReturnValue('default').mockReturnValueOnce('first call'); test();",
        ],
        [
          'window.fetch = jest.fn(() => ({})).one.two().three().four',
          58,
          "jest.spyOn(window, 'fetch').mockImplementation(() => ({})).one.two().three().four",
        ],
        [
          'foo[bar] = jest.fn().mockReturnValue(undefined)',
          48,
          'jest.spyOn(foo, bar).mockImplementation().mockReturnValue(undefined)',
        ],
      ] as const
    ).map(([code, endColumn, output], index): TestCase => ({
      name: `upstream prefer-spy-on / invalid ${index + 6}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'useJestSpyOn',
          message: 'Use jest.spyOn() instead',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    {
      name: 'upstream prefer-spy-on / invalid 9',
      code: `
        foo.bar = jest.fn().mockImplementation(baz => baz)
        foo.bar = jest.fn(a => b).mockImplementation(baz => baz)
     \u0020`,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'useJestSpyOn',
          message: 'Use jest.spyOn() instead',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 59,
        },
        {
          messageId: 'useJestSpyOn',
          message: 'Use jest.spyOn() instead',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 65,
        },
      ],
      output: `
        jest.spyOn(foo, 'bar').mockImplementation(baz => baz)
        jest.spyOn(foo, 'bar').mockImplementation(baz => baz)
     \u0020`,
    },
  ],
});
