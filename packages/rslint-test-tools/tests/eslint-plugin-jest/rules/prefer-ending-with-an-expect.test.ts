// Upstream reference: eslint-plugin-jest@29.16.6; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('prefer-ending-with-an-expect', {
  valid: [
    ...(
      [
        'it.todo("will test something eventually")',
        'test.todo("will test something eventually")',
        "['x']();",
        'it("is weird", "because this should be a function")',
        'it("is weird", "because this should be a function", () => {})',
        'it("should pass", () => expect(true).toBeDefined())',
        'test("should pass", () => expect(true).toBeDefined())',
        'it("should pass", myTest); function myTest() { expect(true).toBeDefined() }',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    {
      name: 'run 0 valid 8',
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
      name: 'run 0 valid 9',
      code: 'it("should return undefined",() => expectSaga(mySaga).returns());',
      options: [
        {
          assertFunctionNames: ['expectSaga'],
        },
      ],
    },
    {
      name: 'run 0 valid 10',
      code: "test('verifies expect method call', () => expect$(123));",
      options: [
        {
          assertFunctionNames: ['expect\\$'],
        },
      ],
    },
    {
      name: 'run 0 valid 11',
      code: "test('verifies expect method call', () => new Foo().expect(123));",
      options: [
        {
          assertFunctionNames: ['Foo.expect'],
        },
      ],
    },
    {
      name: 'run 0 valid 12',
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
      name: 'run 0 valid 13',
      code: `
        test('verifies chained expect method call', () => {
          doSomething();

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
      name: 'run 0 valid 14',
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
      name: 'run 0 valid 15',
      code: 'it("should pass", async () => expect(true).toBeDefined())',
    },
    {
      name: 'run 0 valid 16',
      code: 'it("should pass", () => expect(true).toBeDefined())',
      options: [
        {
          assertFunctionNames: undefined,
          additionalTestBlockFunctions: undefined,
        },
      ],
    },
    ...(
      [
        'it("should pass", () => { expect(true).toBeDefined() })',
        'it("should pass", function () { expect(true).toBeDefined() })',
        `
      it('is a complete test', () => {
        const container = render(Greeter);

        expect(container).toBeDefined();

        container.setProp('name', 'Bob');

        expect(container.toHTML()).toContain('Hello Bob!');
      });
   \u0020`,
        `
        it('is a complete test', async () => {
          const container = render(Greeter);

          expect(container).toBeDefined();

          container.setProp('name', 'Bob');

          await expect(container.toHTML()).resolve.toContain('Hello Bob!');
        });
     \u0020`,
        `
        it('is a complete test', async function () {
          const container = render(Greeter);

          expect(container).toBeDefined();

          container.setProp('name', 'Bob');

          await expect(container.toHTML()).resolve.toContain('Hello Bob!');
        });
     \u0020`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index + 17}`,
      code,
    })),
    {
      name: 'run 0 valid 22',
      code: `
        describe('GET /user', function () {
          it('responds with json', function (done) {
            doSomething();
            request(app).get('/user').expect('Content-Type', /json/).expect(200, done);
          });
        });
     \u0020`,
      options: [
        {
          assertFunctionNames: ['expect', 'request.**.expect'],
        },
      ],
    },
    {
      name: 'run 0 valid 23',
      code: `
        each([
          [2, 3],
          [1, 3],
        ]).test(
          'the selection can change from %d to %d',
          (firstSelection, secondSelection) => {
            const container = render(MySelect, {
              props: { options: [1, 2, 3], selected: firstSelection },
            });

            expect(container).toBeDefined();
            expect(container.toHTML()).toContain(
              \`<option value="\${firstSelection}" selected>\`
            );

            container.setProp('selected', secondSelection);

            expect(container.toHTML()).not.toContain(
              \`<option value="\${firstSelection}" selected>\`
            );
            expect(container.toHTML()).toContain(
              \`<option value="\${secondSelection}" selected>\`
            );
          }
        );
     \u0020`,
      options: [
        {
          additionalTestBlockFunctions: ['each.test'],
        },
      ],
    },
    {
      name: 'run 0 valid 24',
      code: "test('should pass *', () => expect404ToBeLoaded());",
      options: [
        {
          assertFunctionNames: ['expect*'],
        },
      ],
    },
    {
      name: 'run 0 valid 25',
      code: "test('should pass *', () => expect.toHaveStatus404());",
      options: [
        {
          assertFunctionNames: ['expect.**'],
        },
      ],
    },
    {
      name: 'run 0 valid 26',
      code: "test('should pass', () => tester.foo().expect(123));",
      options: [
        {
          assertFunctionNames: ['tester.*.expect'],
        },
      ],
    },
    {
      name: 'run 0 valid 27',
      code: "test('should pass **', () => tester.foo().expect(123));",
      options: [
        {
          assertFunctionNames: ['**'],
        },
      ],
    },
    {
      name: 'run 0 valid 28',
      code: "test('should pass *', () => tester.foo().expect(123));",
      options: [
        {
          assertFunctionNames: ['*'],
        },
      ],
    },
    {
      name: 'run 0 valid 29',
      code: "test('should pass', () => tester.foo().expect(123));",
      options: [
        {
          assertFunctionNames: ['tester.**'],
        },
      ],
    },
    {
      name: 'run 0 valid 30',
      code: "test('should pass', () => tester.foo().expect(123));",
      options: [
        {
          assertFunctionNames: ['tester.*'],
        },
      ],
    },
    {
      name: 'run 0 valid 31',
      code: "test('should pass', () => tester.foo().bar().expectIt(456));",
      options: [
        {
          assertFunctionNames: ['tester.**.expect*'],
        },
      ],
    },
    {
      name: 'run 0 valid 32',
      code: "test('should pass', () => request.get().foo().expect(456));",
      options: [
        {
          assertFunctionNames: ['request.**.expect'],
        },
      ],
    },
    {
      name: 'run 0 valid 33',
      code: "test('should pass', () => request.get().foo().expect(456));",
      options: [
        {
          assertFunctionNames: ['request.**.e*e*t'],
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
      name: `run 0 valid ${index + 34}`,
      code,
      options: [
        {
          assertFunctionNames: ['expect', 'foo'],
        },
      ],
    })),
    {
      name: 'run 0 valid 36',
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
        'it("is weird", "because this should be a function")',
        'it("is weird", "because this should be a function", () => {})',
        'it("should pass", () => expect(true).toBeDefined())',
        'test("should pass", () => expect(true).toBeDefined())',
        'it("should pass", myTest); function myTest() { expect(true).toBeDefined() }',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream prefer-ending-with-an-expect / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
    })),
    {
      name: 'upstream prefer-ending-with-an-expect / valid 8',
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
      name: 'upstream prefer-ending-with-an-expect / valid 9',
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
      name: 'upstream prefer-ending-with-an-expect / valid 10',
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
      name: 'upstream prefer-ending-with-an-expect / valid 11',
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
      name: 'upstream prefer-ending-with-an-expect / valid 12',
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
      name: 'upstream prefer-ending-with-an-expect / valid 13',
      code: `test('verifies chained expect method call', () => {
  doSomething();

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
      name: 'upstream prefer-ending-with-an-expect / valid 14',
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
      name: 'upstream prefer-ending-with-an-expect / valid 15',
      code: 'it("should pass", async () => expect(true).toBeDefined())',
      languageOptions: {
        ecmaVersion: 2017,
      },
    },
    {
      name: 'upstream prefer-ending-with-an-expect / valid 16',
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
    ...(
      [
        'it("should pass", () => { expect(true).toBeDefined() })',
        'it("should pass", function () { expect(true).toBeDefined() })',
        `it('is a complete test', () => {
  const container = render(Greeter);

  expect(container).toBeDefined();

  container.setProp('name', 'Bob');

  expect(container.toHTML()).toContain('Hello Bob!');
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream prefer-ending-with-an-expect / valid ${index + 17}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
    })),
    ...(
      [
        `it('is a complete test', async () => {
  const container = render(Greeter);

  expect(container).toBeDefined();

  container.setProp('name', 'Bob');

  await expect(container.toHTML()).resolve.toContain('Hello Bob!');
});`,
        `it('is a complete test', async function () {
  const container = render(Greeter);

  expect(container).toBeDefined();

  container.setProp('name', 'Bob');

  await expect(container.toHTML()).resolve.toContain('Hello Bob!');
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream prefer-ending-with-an-expect / valid ${index + 20}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
    })),
    {
      name: 'upstream prefer-ending-with-an-expect / valid 22',
      code: `describe('GET /user', function () {
  it('responds with json', function (done) {
    doSomething();
    request(app).get('/user').expect('Content-Type', /json/).expect(200, done);
  });
});`,
      options: [
        {
          assertFunctionNames: ['expect', 'request.**.expect'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
    {
      name: 'upstream prefer-ending-with-an-expect / valid 23',
      code: `each([
  [2, 3],
  [1, 3],
]).test(
  'the selection can change from %d to %d',
  (firstSelection, secondSelection) => {
    const container = render(MySelect, {
      props: { options: [1, 2, 3], selected: firstSelection },
    });

    expect(container).toBeDefined();
    expect(container.toHTML()).toContain(
      \`<option value="\${firstSelection}" selected>\`
    );

    container.setProp('selected', secondSelection);

    expect(container.toHTML()).not.toContain(
      \`<option value="\${firstSelection}" selected>\`
    );
    expect(container.toHTML()).toContain(
      \`<option value="\${secondSelection}" selected>\`
    );
  }
);`,
      options: [
        {
          additionalTestBlockFunctions: ['each.test'],
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
      name: `upstream aliases / valid ${index}`,
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
      name: 'upstream aliases / valid 2',
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
        ['test("should fail", () => {});', 5],
        ['test.skip("should fail", () => {});', 10],
        ['it("should fail", () => { somePromise.then(() => {}); });', 3],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'mustEndWithExpect',
          message: 'Tests should end with an assertion',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    ...(
      [
        ['test("should fail", () => { foo(true).toBe(true); })', 5],
        ['it("should also fail",() => expectSaga(mySaga).returns());', 3],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index + 4}`,
      code,
      options: [
        {
          assertFunctionNames: ['expect'],
        },
      ],
      errors: [
        {
          messageId: 'mustEndWithExpect',
          message: 'Tests should end with an assertion',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    ...(
      [
        'it("should pass", () => somePromise().then(() => expect(true).toBeDefined()))',
        'it("should pass", () => render(Greeter))',
        'it("should pass", () => { render(Greeter) })',
        'it("should pass", function () { render(Greeter) })',
        'it("should not pass", () => class {})',
        'it("should not pass", () => ([]))',
        'it("should not pass", () => { const x = []; })',
        'it("should not pass", function () { class Mx {} })',
        `it('is a complete test', () => {
  const container = render(Greeter);

  expect(container).toBeDefined();

  container.setProp('name', 'Bob');
});`,
        `it('is a complete test', async () => {
  const container = render(Greeter);

  await expect(container).toBeDefined();

  await container.setProp('name', 'Bob');
});`,
      ] as const
    ).map((code, index): TestCase => ({
      name: `run 0 invalid ${index + 6}`,
      code,
      errors: [
        {
          messageId: 'mustEndWithExpect',
          message: 'Tests should end with an assertion',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 3,
        },
      ],
    })),
    {
      name: 'run 0 invalid 16',
      code: "test('should fail', () => request.get().foo().expect(456));",
      options: [
        {
          assertFunctionNames: ['request.*.expect'],
        },
      ],
      errors: [
        {
          messageId: 'mustEndWithExpect',
          message: 'Tests should end with an assertion',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 5,
        },
      ],
    },
    {
      name: 'run 0 invalid 17',
      code: "test('should fail', () => request.get().foo().bar().expect(456));",
      options: [
        {
          assertFunctionNames: ['request.foo**.expect'],
        },
      ],
      errors: [
        {
          messageId: 'mustEndWithExpect',
          message: 'Tests should end with an assertion',
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
      name: `run 0 invalid ${index + 18}`,
      code,
      options: [
        {
          assertFunctionNames: ['request.*'],
        },
      ],
      errors: [
        {
          messageId: 'mustEndWithExpect',
          message: 'Tests should end with an assertion',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 5,
        },
      ],
    })),
    {
      name: 'run 0 invalid 20',
      code: "test('should fail', () => request(123));",
      options: [
        {
          assertFunctionNames: ['request.**'],
        },
      ],
      errors: [
        {
          messageId: 'mustEndWithExpect',
          message: 'Tests should end with an assertion',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 5,
        },
      ],
    },
    {
      name: 'run 0 invalid 21',
      code: `import { test as checkThat } from '@jest/globals';

checkThat('this passes', () => {
  // ...
});`,
      options: [
        {
          assertFunctionNames: ['expect', 'foo'],
        },
      ],
      errors: [
        {
          messageId: 'mustEndWithExpect',
          message: 'Tests should end with an assertion',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 10,
        },
      ],
    },
    {
      name: 'run 0 invalid 22',
      code: `import { test as checkThat } from '@jest/globals';

checkThat.skip('this passes', () => {
  // ...
});`,
      errors: [
        {
          messageId: 'mustEndWithExpect',
          message: 'Tests should end with an assertion',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 15,
        },
      ],
    },
    ...(
      [
        ['it("should fail", () => {});', 3],
        ['test("should fail", () => {});', 5],
        ['test.skip("should fail", () => {});', 10],
        ['it("should fail", () => { somePromise.then(() => {}); });', 3],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `upstream prefer-ending-with-an-expect / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'mustEndWithExpect',
          message: 'Tests should end with an assertion',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    ...(
      [
        ['test("should fail", () => { foo(true).toBe(true); })', 5],
        ['it("should also fail",() => expectSaga(mySaga).returns());', 3],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `upstream prefer-ending-with-an-expect / invalid ${index + 4}`,
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
          messageId: 'mustEndWithExpect',
          message: 'Tests should end with an assertion',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    ...(
      [
        'it("should pass", () => somePromise().then(() => expect(true).toBeDefined()))',
        'it("should pass", () => render(Greeter))',
        'it("should pass", () => { render(Greeter) })',
        'it("should pass", function () { render(Greeter) })',
        'it("should not pass", () => class {})',
        'it("should not pass", () => ([]))',
        'it("should not pass", () => { const x = []; })',
        'it("should not pass", function () { class Mx {} })',
        `it('is a complete test', () => {
  const container = render(Greeter);

  expect(container).toBeDefined();

  container.setProp('name', 'Bob');
});`,
      ] as const
    ).map((code, index): TestCase => ({
      name: `upstream prefer-ending-with-an-expect / invalid ${index + 6}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'mustEndWithExpect',
          message: 'Tests should end with an assertion',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 3,
        },
      ],
    })),
    {
      name: 'upstream prefer-ending-with-an-expect / invalid 15',
      code: `it('is a complete test', async () => {
  const container = render(Greeter);

  await expect(container).toBeDefined();

  await container.setProp('name', 'Bob');
});`,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'mustEndWithExpect',
          message: 'Tests should end with an assertion',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 3,
        },
      ],
    },
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
          messageId: 'mustEndWithExpect',
          message: 'Tests should end with an assertion',
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
          messageId: 'mustEndWithExpect',
          message: 'Tests should end with an assertion',
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
          messageId: 'mustEndWithExpect',
          message: 'Tests should end with an assertion',
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
          messageId: 'mustEndWithExpect',
          message: 'Tests should end with an assertion',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 5,
        },
      ],
    },
    {
      name: 'upstream aliases / invalid 0',
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
          messageId: 'mustEndWithExpect',
          message: 'Tests should end with an assertion',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 10,
        },
      ],
    },
    {
      name: 'upstream aliases / invalid 1',
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
          messageId: 'mustEndWithExpect',
          message: 'Tests should end with an assertion',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 15,
        },
      ],
    },
  ],
});
