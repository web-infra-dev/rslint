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
    { code: declaration + 'expect(service.method).toMatchSnapshot();' },
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
