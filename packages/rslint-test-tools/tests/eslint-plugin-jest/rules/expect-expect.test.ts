import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('expect-expect', {} as never, {
  valid: [
    { code: 'it.todo("will test something eventually")' },
    { code: 'test.todo("will test something eventually")' },
    { code: "['x']();" },
    { code: 'it("should pass", () => expect(true).toBeDefined())' },
    { code: 'test("should pass", () => expect(true).toBeDefined())' },
    {
      code: 'it("should pass", () => somePromise().then(() => expect(true).toBeDefined()))',
    },
    {
      code: 'it("should pass", myTest); function myTest() { expect(true).toBeDefined() }',
    },
    {
      code: `
        test('should pass', () => {
          expect(true).toBeDefined();
          foo(true).toBe(true);
        });
      `,
      options: [{ assertFunctionNames: ['expect', 'foo'] }],
    },
    {
      code: 'it("should return undefined",() => expectSaga(mySaga).returns());',
      options: [{ assertFunctionNames: ['expectSaga'] }],
    },
    {
      code: "test('verifies expect method call', () => expect$(123));",
      options: [{ assertFunctionNames: ['expect\\$'] }],
    },
    {
      code: "test('verifies expect method call', () => new Foo().expect(123));",
      options: [{ assertFunctionNames: ['Foo.expect'] }],
    },
    {
      code: `
        test('verifies deep expect method call', () => {
          tester.foo().expect(123);
        });
      `,
      options: [{ assertFunctionNames: ['tester.foo.expect'] }],
    },
    {
      code: `
        test('verifies chained expect method call', () => {
          tester
            .foo()
            .bar()
            .expect(456);
        });
      `,
      options: [{ assertFunctionNames: ['tester.foo.bar.expect'] }],
    },
    {
      code: `
        test("verifies the function call", () => {
          td.verify(someFunctionCall())
        })
      `,
      options: [{ assertFunctionNames: ['td.verify'] }],
    },
    {
      code: 'it("should pass", () => expect(true).toBeDefined())',
      options: [
        {
          assertFunctionNames: undefined,
          additionalTestBlockFunctions: undefined,
        },
      ],
    },
    {
      code: `
        theoretically('the number {input} is correctly translated to string', theories, theory => {
          const output = NumberToLongString(theory.input);
          expect(output).toBe(theory.expected);
        })
      `,
      options: [{ additionalTestBlockFunctions: ['theoretically'] }],
    },

    // wildcards
    {
      code: "test('should pass *', () => expect404ToBeLoaded());",
      options: [{ assertFunctionNames: ['expect*'] }],
    },
    {
      code: "test('should pass *', () => expect.toHaveStatus404());",
      options: [{ assertFunctionNames: ['expect.**'] }],
    },
    {
      code: "test('should pass', () => tester.foo().expect(123));",
      options: [{ assertFunctionNames: ['tester.*.expect'] }],
    },
    {
      code: "test('should pass **', () => tester.foo().expect(123));",
      options: [{ assertFunctionNames: ['**'] }],
    },
    {
      code: "test('should pass *', () => tester.foo().expect(123));",
      options: [{ assertFunctionNames: ['*'] }],
    },
    {
      code: "test('should pass', () => tester.foo().expect(123));",
      options: [{ assertFunctionNames: ['tester.**'] }],
    },
    {
      code: "test('should pass', () => tester.foo().expect(123));",
      options: [{ assertFunctionNames: ['tester.*'] }],
    },
    {
      code: "test('should pass', () => tester.foo().bar().expectIt(456));",
      options: [{ assertFunctionNames: ['tester.**.expect*'] }],
    },
    {
      code: "test('should pass', () => request.get().foo().expect(456));",
      options: [{ assertFunctionNames: ['request.**.expect'] }],
    },
    {
      code: "test('should pass', () => request.get().foo().expect(456));",
      options: [{ assertFunctionNames: ['request.**.e*e*t'] }],
    },

    // alias
    {
      code: `
        import { test } from '@jest/globals';

        test('should pass', () => {
          expect(true).toBeDefined();
          foo(true).toBe(true);
        });
      `,
      options: [{ assertFunctionNames: ['expect', 'foo'] }],
    },
    {
      code: `
        import { test as checkThat } from '@jest/globals';

        checkThat('this passes', () => {
          expect(true).toBeDefined();
          foo(true).toBe(true);
        });
      `,
      options: [{ assertFunctionNames: ['expect', 'foo'] }],
    },
    {
      code: `
        const { test } = require('@jest/globals');

        test('verifies chained expect method call', () => {
          tester
            .foo()
            .bar()
            .expect(456);
        });
      `,
      options: [{ assertFunctionNames: ['tester.foo.bar.expect'] }],
    },
  ],
  invalid: [
    {
      code: 'it("should fail", () => {});',
      errors: [
        {
          messageId: 'noAssertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'it("should fail", myTest); function myTest() {}',
      errors: [
        {
          messageId: 'noAssertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'test("should fail", () => {});',
      errors: [
        {
          messageId: 'noAssertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 5,
        },
      ],
    },
    {
      code: 'test.skip("should fail", () => {});',
      errors: [
        {
          messageId: 'noAssertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: 'afterEach(() => {});',
      options: [{ additionalTestBlockFunctions: ['afterEach'] }],
      errors: [
        {
          messageId: 'noAssertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: `
        theoretically('the number {input} is correctly translated to string', theories, theory => {
          const output = NumberToLongString(theory.input);
        })
      `,
      options: [{ additionalTestBlockFunctions: ['theoretically'] }],
      errors: [
        {
          messageId: 'noAssertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'it("should fail", () => { somePromise.then(() => {}); });',
      errors: [
        {
          messageId: 'noAssertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'test("should fail", () => { foo(true).toBe(true); })',
      options: [{ assertFunctionNames: ['expect'] }],
      errors: [
        {
          messageId: 'noAssertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 5,
        },
      ],
    },
    {
      code: 'it("should also fail",() => expectSaga(mySaga).returns());',
      options: [{ assertFunctionNames: ['expect'] }],
      errors: [
        {
          messageId: 'noAssertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 3,
        },
      ],
    },

    // wildcards
    {
      code: "test('should fail', () => request.get().foo().expect(456));",
      options: [{ assertFunctionNames: ['request.*.expect'] }],
      errors: [
        {
          messageId: 'noAssertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 5,
        },
      ],
    },
    {
      code: "test('should fail', () => request.get().foo().bar().expect(456));",
      options: [{ assertFunctionNames: ['request.foo**.expect'] }],
      errors: [
        {
          messageId: 'noAssertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 5,
        },
      ],
    },
    {
      code: "test('should fail', () => tester.request(123));",
      options: [{ assertFunctionNames: ['request.*'] }],
      errors: [
        {
          messageId: 'noAssertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 5,
        },
      ],
    },
    {
      code: "test('should fail', () => request(123));",
      options: [{ assertFunctionNames: ['request.*'] }],
      errors: [
        {
          messageId: 'noAssertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 5,
        },
      ],
    },
    {
      code: "test('should fail', () => request(123));",
      options: [{ assertFunctionNames: ['request.**'] }],
      errors: [
        {
          messageId: 'noAssertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 5,
        },
      ],
    },

    // aliases
    {
      code: `
        import { test as checkThat } from '@jest/globals';

        checkThat('this passes', () => {
          // ...
        });
      `,
      options: [{ assertFunctionNames: ['expect', 'foo'] }],
      errors: [
        {
          messageId: 'noAssertions',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 10,
        },
      ],
    },
    {
      code: `
        import { test as checkThat } from '@jest/globals';

        checkThat.skip('this passes', () => {
          // ...
        });
      `,
      errors: [
        {
          messageId: 'noAssertions',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 15,
        },
      ],
    },
  ],
});
