import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();
const declaration = `export {};
class Service { method() {} }
const service = new Service();
`;

ruleTester.run('unbound-method', {} as never, {
  valid: [
    declaration + 'expect(service.method).toHaveBeenCalledTimes(1);',
    declaration + 'expect(service.method).not.toHaveBeenCalled();',
    declaration + 'expect(service.method).toStrictEqual(somethingElse);',
    declaration + 'expect(value).toBe(service.method);',
    declaration + 'expect(service.method).rejects.toBe(1);',
    declaration + 'jest.mocked(service.method).mockImplementation(() => {});',
    declaration + 'const calls = jest.mocked(service.method).mock.calls;',
    declaration + "jest['mocked'](service.method);",
    declaration +
      "import { jest as j } from '@jest/globals';\nj.mocked(service.method);",
    declaration +
      'expect(() => { expect(service.method).toHaveBeenCalled(); }).not.toThrow();',
    {
      code: 'export {};\nclass Static { static method() {} }\nexpect(Static.method).toThrow();',
      options: [{ ignoreStatic: true }],
    },
  ],
  invalid: [
    ...[
      'toThrow',
      'toThrowError',
      'toThrowErrorMatchingSnapshot',
      'toThrowErrorMatchingInlineSnapshot',
    ].flatMap((matcher) =>
      ['', 'not.'].map((modifier) => ({
        code: declaration + `expect(service.method).${modifier}${matcher}();`,
        errors: [
          {
            messageId: 'unboundWithoutThisAnnotation',
            line: 4,
            column: 8,
            endLine: 4,
            endColumn: 22,
          },
        ],
      })),
    ),
    {
      code: declaration + 'expect(service.method);',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          line: 4,
          column: 8,
          endLine: 4,
          endColumn: 22,
        },
      ],
    },
    {
      code: declaration + 'expect(service?.method).toBe(1);',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          line: 4,
          column: 8,
          endLine: 4,
          endColumn: 23,
        },
      ],
    },
    {
      code:
        declaration + 'const method = service.method;\njest.mocked(method);',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          line: 4,
          column: 16,
          endLine: 4,
          endColumn: 30,
        },
      ],
    },
    {
      code: declaration + 'Promise.resolve().then(service.method);',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          line: 4,
          column: 24,
          endLine: 4,
          endColumn: 38,
        },
      ],
    },
    {
      code:
        declaration +
        'expect(() => { Promise.resolve().then(service.method); }).not.toThrow();',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          line: 4,
          column: 39,
          endLine: 4,
          endColumn: 53,
        },
      ],
    },
    {
      code: declaration + 'const { method } = service;',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          line: 4,
          column: 9,
          endLine: 4,
          endColumn: 15,
        },
      ],
    },
    {
      code: 'export {};\nclass Static { static method() {} }\nexpect(Static.method).toThrow();',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          line: 3,
          column: 8,
          endLine: 3,
          endColumn: 21,
        },
      ],
    },
  ],
});
