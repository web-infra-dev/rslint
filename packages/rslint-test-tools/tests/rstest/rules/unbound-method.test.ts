import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();
const filename = 'files/unbound-method.ts';
const declaration = `export {};
class Service { method() {} }
const service = new Service();
`;

ruleTester.run('unbound-method', {} as never, {
  valid: [
    { code: declaration + 'expect(service.method).toHaveBeenCalledTimes(1);' },
    { code: declaration + 'expect(service.method).not.toHaveBeenCalled();' },
    { code: declaration + 'expect(service.method).toStrictEqual(other);' },
    { code: declaration + 'expect(service.method).resolves.toBeDefined();' },
    {
      code:
        declaration +
        'const matcher = { asymmetricMatch(received) { received(); return true; } }; expect(service.method).not.toEqual({ nested: matcher });',
    },
    {
      code: declaration + "expect(service.method).to.have.property('name');",
    },
    {
      code:
        declaration + 'rs.mocked(service.method).mockImplementation(() => {});',
    },
    { code: declaration + 'rstest.mocked(service.method).mock.calls[0];' },
    {
      code:
        declaration +
        'expect(() => expect(service.method).toHaveBeenCalled()).not.toThrow();',
    },
    { code: 'expect(console.log).toThrow();' },
  ].map((testCase) => ({ ...testCase, filename })),
  invalid: [
    {
      code: declaration + 'expect(service.method).toMatchSnapshot();',
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
      code:
        declaration +
        'expect.extend({ toBe(received) { received(); return { pass: true, message: () => "" }; } }); expect(service.method).toBe(1);',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          line: 4,
          column: 101,
          endLine: 4,
          endColumn: 115,
        },
      ],
    },
    {
      code:
        declaration +
        'expect.addEqualityTesters([(received) => { if (typeof received === "function") received(); return true; }]); expect(service.method).toEqual(1);',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          line: 4,
          column: 116,
          endLine: 4,
          endColumn: 130,
        },
      ],
    },
    {
      code:
        declaration +
        'expect(service.method).toEqual(expect.toSatisfy(received => { received(); return true; }));',
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
      code:
        declaration +
        "expect(service.method).toEqual(expect.schemaMatching({ '~standard': { version: 1, vendor: 'test', validate(received) { received(); return { value: received }; } } }));",
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
      code:
        declaration +
        'const matcher = { asymmetricMatch(received) { received(); return true; } }; expect(service.method).toEqual(expect.toBeOneOf([matcher]));',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          line: 4,
          column: 90,
          endLine: 4,
          endColumn: 104,
        },
      ],
    },
    {
      code:
        declaration +
        'const matchers = [{ asymmetricMatch(received) { received(); return true; } }]; expect(service.method).toBeOneOf(matchers);',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          line: 4,
          column: 88,
          endLine: 4,
          endColumn: 102,
        },
      ],
    },
    {
      code: declaration + 'const method = service.method; rs.mocked(method);',
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
  ].map((testCase) => ({ ...testCase, filename })),
});
