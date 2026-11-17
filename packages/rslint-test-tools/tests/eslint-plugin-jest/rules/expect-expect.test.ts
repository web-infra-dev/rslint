// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('expect-expect', {
  valid: [
    ...(
      [
        'it.todo("will test something eventually")',
        'test.todo("will test something eventually")',
        "['x']();",
        'it("should pass", () => expect(true).toBeDefined())',
        'test("should pass", () => expect(true).toBeDefined())',
        'it("should pass", () => somePromise().then(() => expect(true).toBeDefined()))',
        'it("should pass", myTest); function myTest() { expect(true).toBeDefined() }',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    {
      name: 'run 0 valid 7',
      code: `
        test('should pass', () => {
          expect(true).toBeDefined();
          foo(true).toBe(true);
        });
     \u0020`,
      options: [
        {
          assertFunctionNames: ['expect', 'foo'],
        },
      ],
    },
    {
      name: 'run 0 valid 8',
      code: 'it("should return undefined",() => expectSaga(mySaga).returns());',
      options: [
        {
          assertFunctionNames: ['expectSaga'],
        },
      ],
    },
    {
      name: 'run 0 valid 9',
      code: "test('verifies expect method call', () => expect$(123));",
      options: [
        {
          assertFunctionNames: ['expect\\$'],
        },
      ],
    },
    {
      name: 'run 0 valid 10',
      code: "test('verifies expect method call', () => new Foo().expect(123));",
      options: [
        {
          assertFunctionNames: ['Foo.expect'],
        },
      ],
    },
    {
      name: 'run 0 valid 11',
      code: `
        test('verifies deep expect method call', () => {
          tester.foo().expect(123);
        });
     \u0020`,
      options: [
        {
          assertFunctionNames: ['tester.foo.expect'],
        },
      ],
    },
    {
      name: 'run 0 valid 12',
      code: `
        test('verifies chained expect method call', () => {
          tester
            .foo()
            .bar()
            .expect(456);
        });
     \u0020`,
      options: [
        {
          assertFunctionNames: ['tester.foo.bar.expect'],
        },
      ],
    },
    {
      name: 'run 0 valid 13',
      code: `
        test("verifies the function call", () => {
          td.verify(someFunctionCall())
        })
     \u0020`,
      options: [
        {
          assertFunctionNames: ['td.verify'],
        },
      ],
    },
    {
      name: 'run 0 valid 14',
      code: 'it("should pass", () => expect(true).toBeDefined())',
      options: [
        {
          assertFunctionNames: undefined,
          additionalTestBlockFunctions: undefined,
        },
      ],
    },
    {
      name: 'run 0 valid 15',
      code: `
        theoretically('the number {input} is correctly translated to string', theories, theory => {
          const output = NumberToLongString(theory.input);
          expect(output).toBe(theory.expected);
        })
     \u0020`,
      options: [
        {
          additionalTestBlockFunctions: ['theoretically'],
        },
      ],
    },
    {
      name: 'run 0 valid 16',
      code: "test('should pass *', () => expect404ToBeLoaded());",
      options: [
        {
          assertFunctionNames: ['expect*'],
        },
      ],
    },
    {
      name: 'run 0 valid 17',
      code: "test('should pass *', () => expect.toHaveStatus404());",
      options: [
        {
          assertFunctionNames: ['expect.**'],
        },
      ],
    },
    {
      name: 'run 0 valid 18',
      code: "test('should pass', () => tester.foo().expect(123));",
      options: [
        {
          assertFunctionNames: ['tester.*.expect'],
        },
      ],
    },
    {
      name: 'run 0 valid 19',
      code: "test('should pass **', () => tester.foo().expect(123));",
      options: [
        {
          assertFunctionNames: ['**'],
        },
      ],
    },
    {
      name: 'run 0 valid 20',
      code: "test('should pass *', () => tester.foo().expect(123));",
      options: [
        {
          assertFunctionNames: ['*'],
        },
      ],
    },
    {
      name: 'run 0 valid 21',
      code: "test('should pass', () => tester.foo().expect(123));",
      options: [
        {
          assertFunctionNames: ['tester.**'],
        },
      ],
    },
    {
      name: 'run 0 valid 22',
      code: "test('should pass', () => tester.foo().expect(123));",
      options: [
        {
          assertFunctionNames: ['tester.*'],
        },
      ],
    },
    {
      name: 'run 0 valid 23',
      code: "test('should pass', () => tester.foo().bar().expectIt(456));",
      options: [
        {
          assertFunctionNames: ['tester.**.expect*'],
        },
      ],
    },
    {
      name: 'run 0 valid 24',
      code: "test('should pass', () => request.get().foo().expect(456));",
      options: [
        {
          assertFunctionNames: ['request.**.expect'],
        },
      ],
    },
    {
      name: 'run 0 valid 25',
      code: "test('should pass', () => request.get().foo().expect(456));",
      options: [
        {
          assertFunctionNames: ['request.**.e*e*t'],
        },
      ],
    },
    {
      name: 'run 0 valid 26',
      code: "test('should pass', () => foo[bar].expect(1));",
      options: [
        {
          assertFunctionNames: ['foo.bar.expect'],
        },
      ],
    },
    ...(
      [
        `
        import { test } from '@jest/globals';

        test('should pass', () => {
          expect(true).toBeDefined();
          foo(true).toBe(true);
        });
     \u0020`,
        `
        import { test as checkThat } from '@jest/globals';

        checkThat('this passes', () => {
          expect(true).toBeDefined();
          foo(true).toBe(true);
        });
     \u0020`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index + 27}`,
      code,
      options: [
        {
          assertFunctionNames: ['expect', 'foo'],
        },
      ],
    })),
    {
      name: 'run 0 valid 29',
      code: `
        const { test } = require('@jest/globals');

        test('verifies chained expect method call', () => {
          tester
            .foo()
            .bar()
            .expect(456);
        });
     \u0020`,
      options: [
        {
          assertFunctionNames: ['tester.foo.bar.expect'],
        },
      ],
    },
    ...(
      [
        'it.todo("will test something eventually")',
        'test.todo("will test something eventually")',
        "['x']();",
        'it("should pass", () => expect(true).toBeDefined())',
        'test("should pass", () => expect(true).toBeDefined())',
        'it("should pass", () => somePromise().then(() => expect(true).toBeDefined()))',
        'it("should pass", myTest); function myTest() { expect(true).toBeDefined() }',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream expect-expect / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
    })),
    {
      name: 'upstream expect-expect / valid 7',
      code: `test('should pass', () => {
  expect(true).toBeDefined();
  foo(true).toBe(true);
});`,
      options: [
        {
          assertFunctionNames: ['expect', 'foo'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
    {
      name: 'upstream expect-expect / valid 8',
      code: 'it("should return undefined",() => expectSaga(mySaga).returns());',
      options: [
        {
          assertFunctionNames: ['expectSaga'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
    {
      name: 'upstream expect-expect / valid 9',
      code: "test('verifies expect method call', () => expect$(123));",
      options: [
        {
          assertFunctionNames: ['expect\\$'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
    {
      name: 'upstream expect-expect / valid 10',
      code: "test('verifies expect method call', () => new Foo().expect(123));",
      options: [
        {
          assertFunctionNames: ['Foo.expect'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
    {
      name: 'upstream expect-expect / valid 11',
      code: `test('verifies deep expect method call', () => {
  tester.foo().expect(123);
});`,
      options: [
        {
          assertFunctionNames: ['tester.foo.expect'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
    {
      name: 'upstream expect-expect / valid 12',
      code: `test('verifies chained expect method call', () => {
  tester
    .foo()
    .bar()
    .expect(456);
});`,
      options: [
        {
          assertFunctionNames: ['tester.foo.bar.expect'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
    {
      name: 'upstream expect-expect / valid 13',
      code: `test("verifies the function call", () => {
  td.verify(someFunctionCall())
})`,
      options: [
        {
          assertFunctionNames: ['td.verify'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
    {
      name: 'upstream expect-expect / valid 14',
      code: 'it("should pass", () => expect(true).toBeDefined())',
      options: [
        {
          assertFunctionNames: undefined,
          additionalTestBlockFunctions: undefined,
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
    {
      name: 'upstream expect-expect / valid 15',
      code: `theoretically('the number {input} is correctly translated to string', theories, theory => {
  const output = NumberToLongString(theory.input);
  expect(output).toBe(theory.expected);
})`,
      options: [
        {
          additionalTestBlockFunctions: ['theoretically'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
    {
      name: 'upstream wildcards / valid 0',
      code: "test('should pass *', () => expect404ToBeLoaded());",
      options: [
        {
          assertFunctionNames: ['expect*'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
    {
      name: 'upstream wildcards / valid 1',
      code: "test('should pass *', () => expect.toHaveStatus404());",
      options: [
        {
          assertFunctionNames: ['expect.**'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
    {
      name: 'upstream wildcards / valid 2',
      code: "test('should pass', () => tester.foo().expect(123));",
      options: [
        {
          assertFunctionNames: ['tester.*.expect'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
    {
      name: 'upstream wildcards / valid 3',
      code: "test('should pass **', () => tester.foo().expect(123));",
      options: [
        {
          assertFunctionNames: ['**'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
    {
      name: 'upstream wildcards / valid 4',
      code: "test('should pass *', () => tester.foo().expect(123));",
      options: [
        {
          assertFunctionNames: ['*'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
    {
      name: 'upstream wildcards / valid 5',
      code: "test('should pass', () => tester.foo().expect(123));",
      options: [
        {
          assertFunctionNames: ['tester.**'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
    {
      name: 'upstream wildcards / valid 6',
      code: "test('should pass', () => tester.foo().expect(123));",
      options: [
        {
          assertFunctionNames: ['tester.*'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
    {
      name: 'upstream wildcards / valid 7',
      code: "test('should pass', () => tester.foo().bar().expectIt(456));",
      options: [
        {
          assertFunctionNames: ['tester.**.expect*'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
    {
      name: 'upstream wildcards / valid 8',
      code: "test('should pass', () => request.get().foo().expect(456));",
      options: [
        {
          assertFunctionNames: ['request.**.expect'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
    {
      name: 'upstream wildcards / valid 9',
      code: "test('should pass', () => request.get().foo().expect(456));",
      options: [
        {
          assertFunctionNames: ['request.**.e*e*t'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
    ...(
      [
        `import { test } from '@jest/globals';

test('should pass', () => {
  expect(true).toBeDefined();
  foo(true).toBe(true);
});`,
        `import { test as checkThat } from '@jest/globals';

checkThat('this passes', () => {
  expect(true).toBeDefined();
  foo(true).toBe(true);
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream expect-expect (aliases) / valid ${index}`,
      code,
      options: [
        {
          assertFunctionNames: ['expect', 'foo'],
        },
      ],
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2015,
      },
    })),
    {
      name: 'upstream expect-expect (aliases) / valid 2',
      code: `const { test } = require('@jest/globals');

test('verifies chained expect method call', () => {
  tester
    .foo()
    .bar()
    .expect(456);
});`,
      options: [
        {
          assertFunctionNames: ['tester.foo.bar.expect'],
        },
      ],
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2015,
      },
    },
  ],
  invalid: [
    ...(
      [
        ['it("should fail", () => {});', 3],
        ['it("should fail", myTest); function myTest() {}', 3],
        ['test("should fail", () => {});', 5],
        ['test.skip("should fail", () => {});', 10],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'noAssertions',
          message: 'Test has no assertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    {
      name: 'run 0 invalid 4',
      code: 'afterEach(() => {});',
      options: [
        {
          additionalTestBlockFunctions: ['afterEach'],
        },
      ],
      errors: [
        {
          messageId: 'noAssertions',
          message: 'Test has no assertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      name: 'run 0 invalid 5',
      code: `
        theoretically('the number {input} is correctly translated to string', theories, theory => {
          const output = NumberToLongString(theory.input);
        })
     \u0020`,
      options: [
        {
          additionalTestBlockFunctions: ['theoretically'],
        },
      ],
      errors: [
        {
          messageId: 'noAssertions',
          message: 'Test has no assertions',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 22,
        },
      ],
    },
    {
      name: 'run 0 invalid 6',
      code: 'it("should fail", () => { somePromise.then(() => {}); });',
      errors: [
        {
          messageId: 'noAssertions',
          message: 'Test has no assertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 3,
        },
      ],
    },
    ...(
      [
        ['test("should fail", () => { foo(true).toBe(true); })', 5],
        ['it("should also fail",() => expectSaga(mySaga).returns());', 3],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index + 7}`,
      code,
      options: [
        {
          assertFunctionNames: ['expect'],
        },
      ],
      errors: [
        {
          messageId: 'noAssertions',
          message: 'Test has no assertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    {
      name: 'run 0 invalid 9',
      code: "test('should fail', () => request.get().foo().expect(456));",
      options: [
        {
          assertFunctionNames: ['request.*.expect'],
        },
      ],
      errors: [
        {
          messageId: 'noAssertions',
          message: 'Test has no assertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 5,
        },
      ],
    },
    {
      name: 'run 0 invalid 10',
      code: "test('should fail', () => request.get().foo().bar().expect(456));",
      options: [
        {
          assertFunctionNames: ['request.foo**.expect'],
        },
      ],
      errors: [
        {
          messageId: 'noAssertions',
          message: 'Test has no assertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 5,
        },
      ],
    },
    ...(
      [
        "test('should fail', () => tester.request(123));",
        "test('should fail', () => request(123));",
      ] as const
    ).map((code, index): TestCase => ({
      name: `run 0 invalid ${index + 11}`,
      code,
      options: [
        {
          assertFunctionNames: ['request.*'],
        },
      ],
      errors: [
        {
          messageId: 'noAssertions',
          message: 'Test has no assertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 5,
        },
      ],
    })),
    {
      name: 'run 0 invalid 13',
      code: "test('should fail', () => request(123));",
      options: [
        {
          assertFunctionNames: ['request.**'],
        },
      ],
      errors: [
        {
          messageId: 'noAssertions',
          message: 'Test has no assertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 5,
        },
      ],
    },
    {
      name: 'run 0 invalid 14',
      code: "test('should fail for identifier bracket access mismatch', () => foo[bar].expect(1));",
      options: [
        {
          assertFunctionNames: ['foo.expect'],
        },
      ],
      errors: [
        {
          messageId: 'noAssertions',
          message: 'Test has no assertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 5,
        },
      ],
    },
    {
      name: 'run 0 invalid 15',
      code: "test('should fail for numeric bracket access', () => foo[1].expect(1));",
      options: [
        {
          assertFunctionNames: ['foo.1.expect'],
        },
      ],
      errors: [
        {
          messageId: 'noAssertions',
          message: 'Test has no assertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 5,
        },
      ],
    },
    {
      name: 'run 0 invalid 16',
      code: "test('should fail for dynamic bracket access', () => foo[a + b].expect(1));",
      options: [
        {
          assertFunctionNames: ['foo.expect'],
        },
      ],
      errors: [
        {
          messageId: 'noAssertions',
          message: 'Test has no assertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 5,
        },
      ],
    },
    {
      name: 'run 0 invalid 17',
      code: `
        import { test as checkThat } from '@jest/globals';

        checkThat('this passes', () => {
          // ...
        });
     \u0020`,
      options: [
        {
          assertFunctionNames: ['expect', 'foo'],
        },
      ],
      errors: [
        {
          messageId: 'noAssertions',
          message: 'Test has no assertions',
          line: 4,
          column: 9,
          endLine: 4,
          endColumn: 18,
        },
      ],
    },
    {
      name: 'run 0 invalid 18',
      code: `
        import { test as checkThat } from '@jest/globals';

        checkThat.skip('this passes', () => {
          // ...
        });
     \u0020`,
      errors: [
        {
          messageId: 'noAssertions',
          message: 'Test has no assertions',
          line: 4,
          column: 9,
          endLine: 4,
          endColumn: 23,
        },
      ],
    },
    ...(
      [
        ['it("should fail", () => {});', 3],
        ['it("should fail", myTest); function myTest() {}', 3],
        ['test("should fail", () => {});', 5],
        ['test.skip("should fail", () => {});', 10],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `upstream expect-expect / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'noAssertions',
          message: 'Test has no assertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    {
      name: 'upstream expect-expect / invalid 4',
      code: 'afterEach(() => {});',
      options: [
        {
          additionalTestBlockFunctions: ['afterEach'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'noAssertions',
          message: 'Test has no assertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      name: 'upstream expect-expect / invalid 5',
      code: `theoretically('the number {input} is correctly translated to string', theories, theory => {
  const output = NumberToLongString(theory.input);
})`,
      options: [
        {
          additionalTestBlockFunctions: ['theoretically'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'noAssertions',
          message: 'Test has no assertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      name: 'upstream expect-expect / invalid 6',
      code: 'it("should fail", () => { somePromise.then(() => {}); });',
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'noAssertions',
          message: 'Test has no assertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 3,
        },
      ],
    },
    ...(
      [
        ['test("should fail", () => { foo(true).toBe(true); })', 5],
        ['it("should also fail",() => expectSaga(mySaga).returns());', 3],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `upstream expect-expect / invalid ${index + 7}`,
      code,
      options: [
        {
          assertFunctionNames: ['expect'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'noAssertions',
          message: 'Test has no assertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    {
      name: 'upstream wildcards / invalid 0',
      code: "test('should fail', () => request.get().foo().expect(456));",
      options: [
        {
          assertFunctionNames: ['request.*.expect'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'noAssertions',
          message: 'Test has no assertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 5,
        },
      ],
    },
    {
      name: 'upstream wildcards / invalid 1',
      code: "test('should fail', () => request.get().foo().bar().expect(456));",
      options: [
        {
          assertFunctionNames: ['request.foo**.expect'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'noAssertions',
          message: 'Test has no assertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 5,
        },
      ],
    },
    ...(
      [
        "test('should fail', () => tester.request(123));",
        "test('should fail', () => request(123));",
      ] as const
    ).map((code, index): TestCase => ({
      name: `upstream wildcards / invalid ${index + 2}`,
      code,
      options: [
        {
          assertFunctionNames: ['request.*'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'noAssertions',
          message: 'Test has no assertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 5,
        },
      ],
    })),
    {
      name: 'upstream wildcards / invalid 4',
      code: "test('should fail', () => request(123));",
      options: [
        {
          assertFunctionNames: ['request.**'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'noAssertions',
          message: 'Test has no assertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 5,
        },
      ],
    },
    {
      name: 'upstream expect-expect (aliases) / invalid 0',
      code: `import { test as checkThat } from '@jest/globals';

checkThat('this passes', () => {
  // ...
});`,
      options: [
        {
          assertFunctionNames: ['expect', 'foo'],
        },
      ],
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'noAssertions',
          message: 'Test has no assertions',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 10,
        },
      ],
    },
    {
      name: 'upstream expect-expect (aliases) / invalid 1',
      code: `import { test as checkThat } from '@jest/globals';

checkThat.skip('this passes', () => {
  // ...
});`,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'noAssertions',
          message: 'Test has no assertions',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 15,
        },
      ],
    },
  ],
});
