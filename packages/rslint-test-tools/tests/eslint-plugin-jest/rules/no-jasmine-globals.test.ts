import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-jasmine-globals', {} as never, {
  valid: [
    { code: 'jest.spyOn()' },
    { code: 'jest.fn()' },
    { code: 'expect.extend()' },
    { code: 'expect.any()' },
    { code: 'it("foo", function () {})' },
    { code: 'test("foo", function () {})' },
    { code: 'foo()' },
    { code: `require('foo')('bar')` },
    { code: '(function(){})()' },
    { code: 'function callback(fail) { fail() }' },
    { code: 'var spyOn = require("actions"); spyOn("foo")' },
    { code: 'function callback(pending) { pending() }' },
    { code: 'function callback(jasmine) { jasmine.any() }' },
    { code: 'const jasmine = foo;' },
    { code: '(this as any).jasmine.foo()' },
    { code: '(this as any).jasmine.any()' },
    { code: 'jasmine[1].any()' },
  ],
  invalid: [
    {
      code: 'spyOn(some, "object")',
      output: null,
      errors: [
        {
          messageId: 'illegalGlobal',
          data: { global: 'spyOn', replacement: 'jest.spyOn' },
          column: 1,
          line: 1,
        },
      ],
    },
    {
      code: 'spyOnProperty(some, "object")',
      output: null,
      errors: [
        {
          messageId: 'illegalGlobal',
          data: { global: 'spyOnProperty', replacement: 'jest.spyOn' },
          column: 1,
          line: 1,
        },
      ],
    },
    {
      code: 'fail()',
      output: null,
      errors: [{ messageId: 'illegalFail', column: 1, line: 1 }],
    },
    {
      code: 'pending()',
      output: null,
      errors: [{ messageId: 'illegalPending', column: 1, line: 1 }],
    },
    {
      code: 'jasmine',
      output: null,
      errors: [{ messageId: 'illegalJasmine', column: 1, line: 1 }],
    },
    {
      code: 'const value = jasmine;',
      output: null,
      errors: [{ messageId: 'illegalJasmine', column: 15, line: 1 }],
    },
    {
      code: 'jasmine.DEFAULT_TIMEOUT_INTERVAL = 5000;',
      output: 'jest.setTimeout(5000);',
      errors: [{ messageId: 'illegalJasmine', column: 1, line: 1 }],
    },
    {
      code: 'jasmine.DEFAULT_TIMEOUT_INTERVAL = function() {}',
      output: null,
      errors: [{ messageId: 'illegalJasmine', column: 1, line: 1 }],
    },
    {
      code: 'jasmine.DEFAULT_TIMEOUT_INTERVAL += 1000;',
      output: null,
      errors: [{ messageId: 'illegalJasmine', column: 1, line: 1 }],
    },
    {
      code: 'jasmine["DEFAULT_TIMEOUT_INTERVAL"] = 5000;',
      output: 'jest.setTimeout(5000);',
      errors: [{ messageId: 'illegalJasmine', column: 1, line: 1 }],
    },
    {
      code: 'jasmine["DEFAULT_TIMEOUT_INTERVAL"] = function() {}',
      output: null,
      errors: [{ messageId: 'illegalJasmine', column: 1, line: 1 }],
    },
    {
      code: 'jasmine["DEFAULT_TIMEOUT_INTERVAL"] += 1000;',
      output: null,
      errors: [{ messageId: 'illegalJasmine', column: 1, line: 1 }],
    },
    {
      code: 'jasmine.addMatchers(matchers)',
      output: null,
      errors: [
        {
          messageId: 'illegalMethod',
          data: { method: 'jasmine.addMatchers', replacement: 'expect.extend' },
          column: 1,
          line: 1,
        },
      ],
    },
    {
      code: 'jasmine.createSpy()',
      output: null,
      errors: [
        {
          messageId: 'illegalMethod',
          data: { method: 'jasmine.createSpy', replacement: 'jest.fn' },
          column: 1,
          line: 1,
        },
      ],
    },
    {
      code: 'jasmine.any()',
      output: 'expect.any()',
      errors: [
        {
          messageId: 'illegalMethod',
          data: { method: 'jasmine.any', replacement: 'expect.any' },
          column: 1,
          line: 1,
        },
      ],
    },
    {
      code: 'jasmine.anything()',
      output: 'expect.anything()',
      errors: [
        {
          messageId: 'illegalMethod',
          data: { method: 'jasmine.anything', replacement: 'expect.anything' },
          column: 1,
          line: 1,
        },
      ],
    },
    {
      code: 'jasmine.arrayContaining()',
      output: 'expect.arrayContaining()',
      errors: [
        {
          messageId: 'illegalMethod',
          data: {
            method: 'jasmine.arrayContaining',
            replacement: 'expect.arrayContaining',
          },
          column: 1,
          line: 1,
        },
      ],
    },
    {
      code: 'jasmine.objectContaining()',
      output: 'expect.objectContaining()',
      errors: [
        {
          messageId: 'illegalMethod',
          data: {
            method: 'jasmine.objectContaining',
            replacement: 'expect.objectContaining',
          },
          column: 1,
          line: 1,
        },
      ],
    },
    {
      code: 'jasmine.stringMatching()',
      output: 'expect.stringMatching()',
      errors: [
        {
          messageId: 'illegalMethod',
          data: {
            method: 'jasmine.stringMatching',
            replacement: 'expect.stringMatching',
          },
          column: 1,
          line: 1,
        },
      ],
    },
    {
      code: 'jasmine.foo.any()',
      output: null,
      errors: [{ messageId: 'illegalJasmine', column: 1, line: 1 }],
    },
    {
      code: 'jasmine.foo.addMatchers()',
      output: null,
      errors: [{ messageId: 'illegalJasmine', column: 1, line: 1 }],
    },
    {
      code: 'jasmine.foo.createSpy()',
      output: null,
      errors: [{ messageId: 'illegalJasmine', column: 1, line: 1 }],
    },
    {
      code: 'console.log(jasmine.version)',
      output: null,
      errors: [{ messageId: 'illegalJasmine', column: 13, line: 1 }],
    },
    {
      code: 'jasmine.getEnv()',
      output: null,
      errors: [{ messageId: 'illegalJasmine', column: 1, line: 1 }],
    },
    {
      code: 'jasmine.empty()',
      output: null,
      errors: [{ messageId: 'illegalJasmine', column: 1, line: 1 }],
    },
    {
      code: 'jasmine.falsy()',
      output: null,
      errors: [{ messageId: 'illegalJasmine', column: 1, line: 1 }],
    },
    {
      code: 'jasmine.truthy()',
      output: null,
      errors: [{ messageId: 'illegalJasmine', column: 1, line: 1 }],
    },
    {
      code: 'jasmine.arrayWithExactContents()',
      output: null,
      errors: [{ messageId: 'illegalJasmine', column: 1, line: 1 }],
    },
    {
      code: 'jasmine.clock()',
      output: null,
      errors: [{ messageId: 'illegalJasmine', column: 1, line: 1 }],
    },
    {
      code: 'jasmine.MAX_PRETTY_PRINT_ARRAY_LENGTH = 42',
      output: null,
      errors: [{ messageId: 'illegalJasmine', column: 1, line: 1 }],
    },
    {
      code: 'jasmine["MAX_PRETTY_PRINT_ARRAY_LENGTH"] = 42',
      output: null,
      errors: [{ messageId: 'illegalJasmine', column: 1, line: 1 }],
    },
  ],
});
