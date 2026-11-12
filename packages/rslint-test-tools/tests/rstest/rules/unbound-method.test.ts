import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);
const filename = 'files/unbound-method.ts';
const declaration = `export {};
class Service { method() {} }
const service = new Service();
`;

ruleTester.run('unbound-method', {
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
      code: 'export {};\nclass Service { method() {} }\nconst service = new Service();\nexpect(service.method).toMatchSnapshot();',
      filename: 'files/unbound-method.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.\nIf your function does not access `this`, you can annotate it with `this: void`, or consider using an arrow function instead.',
          line: 4,
          column: 8,
          endLine: 4,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'export {};\nclass Service { method() {} }\nconst service = new Service();\nexpect(service.method).toThrow();',
      filename: 'files/unbound-method.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.\nIf your function does not access `this`, you can annotate it with `this: void`, or consider using an arrow function instead.',
          line: 4,
          column: 8,
          endLine: 4,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'export {};\nclass Service { method() {} }\nconst service = new Service();\nexpect(service.method).not.toThrow();',
      filename: 'files/unbound-method.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.\nIf your function does not access `this`, you can annotate it with `this: void`, or consider using an arrow function instead.',
          line: 4,
          column: 8,
          endLine: 4,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'export {};\nclass Service { method() {} }\nconst service = new Service();\nexpect(service.method).toThrowError();',
      filename: 'files/unbound-method.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.\nIf your function does not access `this`, you can annotate it with `this: void`, or consider using an arrow function instead.',
          line: 4,
          column: 8,
          endLine: 4,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'export {};\nclass Service { method() {} }\nconst service = new Service();\nexpect(service.method).not.toThrowError();',
      filename: 'files/unbound-method.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.\nIf your function does not access `this`, you can annotate it with `this: void`, or consider using an arrow function instead.',
          line: 4,
          column: 8,
          endLine: 4,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'export {};\nclass Service { method() {} }\nconst service = new Service();\nexpect(service.method).toThrowErrorMatchingSnapshot();',
      filename: 'files/unbound-method.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.\nIf your function does not access `this`, you can annotate it with `this: void`, or consider using an arrow function instead.',
          line: 4,
          column: 8,
          endLine: 4,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'export {};\nclass Service { method() {} }\nconst service = new Service();\nexpect(service.method).not.toThrowErrorMatchingSnapshot();',
      filename: 'files/unbound-method.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.\nIf your function does not access `this`, you can annotate it with `this: void`, or consider using an arrow function instead.',
          line: 4,
          column: 8,
          endLine: 4,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'export {};\nclass Service { method() {} }\nconst service = new Service();\nexpect(service.method).toThrowErrorMatchingInlineSnapshot();',
      filename: 'files/unbound-method.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.\nIf your function does not access `this`, you can annotate it with `this: void`, or consider using an arrow function instead.',
          line: 4,
          column: 8,
          endLine: 4,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'export {};\nclass Service { method() {} }\nconst service = new Service();\nexpect(service.method).not.toThrowErrorMatchingInlineSnapshot();',
      filename: 'files/unbound-method.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.\nIf your function does not access `this`, you can annotate it with `this: void`, or consider using an arrow function instead.',
          line: 4,
          column: 8,
          endLine: 4,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'export {};\nclass Service { method() {} }\nconst service = new Service();\nexpect(service.method);',
      filename: 'files/unbound-method.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.\nIf your function does not access `this`, you can annotate it with `this: void`, or consider using an arrow function instead.',
          line: 4,
          column: 8,
          endLine: 4,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'export {};\nclass Service { method() {} }\nconst service = new Service();\nexpect.extend({ toBe(received) { received(); return { pass: true, message: () => "" }; } }); expect(service.method).toBe(1);',
      filename: 'files/unbound-method.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.\nIf your function does not access `this`, you can annotate it with `this: void`, or consider using an arrow function instead.',
          line: 4,
          column: 101,
          endLine: 4,
          endColumn: 115,
        },
      ],
    },
    {
      code: 'export {};\nclass Service { method() {} }\nconst service = new Service();\nexpect.addEqualityTesters([(received) => { if (typeof received === "function") received(); return true; }]); expect(service.method).toEqual(1);',
      filename: 'files/unbound-method.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.\nIf your function does not access `this`, you can annotate it with `this: void`, or consider using an arrow function instead.',
          line: 4,
          column: 117,
          endLine: 4,
          endColumn: 131,
        },
      ],
    },
    {
      code: 'export {};\nclass Service { method() {} }\nconst service = new Service();\nexpect(service.method).toEqual(expect.toSatisfy(received => { received(); return true; }));',
      filename: 'files/unbound-method.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.\nIf your function does not access `this`, you can annotate it with `this: void`, or consider using an arrow function instead.',
          line: 4,
          column: 8,
          endLine: 4,
          endColumn: 22,
        },
      ],
    },
    {
      code: "export {};\nclass Service { method() {} }\nconst service = new Service();\nexpect(service.method).toEqual(expect.schemaMatching({ '~standard': { version: 1, vendor: 'test', validate(received) { received(); return { value: received }; } } }));",
      filename: 'files/unbound-method.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.\nIf your function does not access `this`, you can annotate it with `this: void`, or consider using an arrow function instead.',
          line: 4,
          column: 8,
          endLine: 4,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'export {};\nclass Service { method() {} }\nconst service = new Service();\nconst matcher = { asymmetricMatch(received) { received(); return true; } }; expect(service.method).toEqual(expect.toBeOneOf([matcher]));',
      filename: 'files/unbound-method.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.\nIf your function does not access `this`, you can annotate it with `this: void`, or consider using an arrow function instead.',
          line: 4,
          column: 84,
          endLine: 4,
          endColumn: 98,
        },
      ],
    },
    {
      code: 'export {};\nclass Service { method() {} }\nconst service = new Service();\nconst matchers = [{ asymmetricMatch(received) { received(); return true; } }]; expect(service.method).toBeOneOf(matchers);',
      filename: 'files/unbound-method.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.\nIf your function does not access `this`, you can annotate it with `this: void`, or consider using an arrow function instead.',
          line: 4,
          column: 87,
          endLine: 4,
          endColumn: 101,
        },
      ],
    },
    {
      code: 'export {};\nclass Service { method() {} }\nconst service = new Service();\nconst method = service.method; rs.mocked(method);',
      filename: 'files/unbound-method.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.\nIf your function does not access `this`, you can annotate it with `this: void`, or consider using an arrow function instead.',
          line: 4,
          column: 16,
          endLine: 4,
          endColumn: 30,
        },
      ],
    },
  ],
});
