import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('prefer-expect-assertions', {
  valid: [
    {
      code: `test("it1", () => {expect.assertions(0);})`,
    },
    {
      code: `test("it1", function() {expect.assertions(0);})`,
    },
    {
      code: `test("it1", function() {expect.hasAssertions();})`,
    },
    {
      code: `it("it1", function() {expect.assertions(0);})`,
    },
    {
      code: `test("it1")`,
    },
    {
      code: `itHappensToStartWithIt("foo", function() {})`,
    },
    {
      code: `testSomething("bar", function() {})`,
    },
    {
      code: `it(async () => {expect.assertions(0);})`,
    },
    {
      code: `test("example-fail", async ({ expect }) => {
    expect.assertions(1);
    await expect(Promise.resolve(null)).resolves.toBeNull();
  });
    `,
    },
    {
      code: `it("it1", () => {
    expect.assertions(0);
    const foo = { bar({ baz }) { baz(); } };
  });
    `,
    },
    {
      code: `
   const expectNumbersToBeGreaterThan = (numbers, value) => {
    for (let number of numbers) {
    expect(number).toBeGreaterThan(value);
   }
   };

   it('returns numbers that are greater than two', function () {
    expectNumbersToBeGreaterThan(getNumbers(), 2);
   });
   `,
      options: [{ onlyFunctionsWithExpectInLoop: true }],
    },
    {
      code: `
   it("returns numbers that are greater than five", function () {
    expect.assertions(2);
    for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(5);
   }
   });
   `,
      options: [{ onlyFunctionsWithExpectInLoop: true }],
    },
    {
      code: `it("returns things that are less than ten", function () {
    expect.hasAssertions();
    for (const thing in things) {
     expect(thing).toBeLessThan(10);
    }
   });`,
      options: [{ onlyFunctionsWithExpectInLoop: true }],
    },
  ],
  invalid: [
    {
      code: `it("it1", () => foo())`,
      errors: [
        {
          messageId: 'haveExpectAssertions',
          message:
            'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: `
it('my test description', ({ expect }) => {
  const a = 1;
  const b = 2;

  expect(sum(a, b)).toBe(a + b);
})
`,
      errors: [
        {
          messageId: 'haveExpectAssertions',
          message:
            'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
          line: 2,
          column: 1,
          endLine: 7,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              desc: 'Add `expect.hasAssertions()`',
              output:
                "\nit('my test description', ({ expect }) => {expect.hasAssertions();\n  const a = 1;\n  const b = 2;\n\n  expect(sum(a, b)).toBe(a + b);\n})\n",
            },
            {
              messageId: 'suggestAddingAssertions',
              desc: 'Add `expect.assertions(<number of assertions>)`',
              output:
                "\nit('my test description', ({ expect }) => {expect.assertions();\n  const a = 1;\n  const b = 2;\n\n  expect(sum(a, b)).toBe(a + b);\n})\n",
            },
          ],
        },
      ],
    },
    {
      code: `
it('my test description', (context) => {
  const a = 1;
  const b = 2;

  context.expect(sum(a, b)).toBe(a + b);
})
`,
      errors: [
        {
          messageId: 'haveExpectAssertions',
          message:
            'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
          line: 2,
          column: 1,
          endLine: 7,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              desc: 'Add `expect.hasAssertions()`',
              output:
                "\nit('my test description', (context) => {context.expect.hasAssertions();\n  const a = 1;\n  const b = 2;\n\n  context.expect(sum(a, b)).toBe(a + b);\n})\n",
            },
            {
              messageId: 'suggestAddingAssertions',
              desc: 'Add `expect.assertions(<number of assertions>)`',
              output:
                "\nit('my test description', (context) => {context.expect.assertions();\n  const a = 1;\n  const b = 2;\n\n  context.expect(sum(a, b)).toBe(a + b);\n})\n",
            },
          ],
        },
      ],
    },
    {
      code: `it('resolves', () => expect(staged()).toBe(true));`,
      errors: [
        {
          messageId: 'haveExpectAssertions',
          message:
            'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
    {
      code: `it('resolves', async () => expect(await staged()).toBe(true));`,
      errors: [
        {
          messageId: 'haveExpectAssertions',
          message:
            'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 62,
        },
      ],
    },
    {
      code: `it('checks value', () => {
  expect(value).to.be.a("string")
})`,
      errors: [
        {
          messageId: 'haveExpectAssertions',
          message:
            'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
          line: 1,
          column: 1,
          endLine: 3,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              desc: 'Add `expect.hasAssertions()`',
              output:
                'it(\'checks value\', () => {expect.hasAssertions();\n  expect(value).to.be.a("string")\n})',
            },
            {
              messageId: 'suggestAddingAssertions',
              desc: 'Add `expect.assertions(<number of assertions>)`',
              output:
                'it(\'checks value\', () => {expect.assertions();\n  expect(value).to.be.a("string")\n})',
            },
          ],
        },
      ],
    },
    {
      code: `it("it1", () => {})`,
      errors: [
        {
          messageId: 'haveExpectAssertions',
          message:
            'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              desc: 'Add `expect.hasAssertions()`',
              output: 'it("it1", () => {expect.hasAssertions();})',
            },
            {
              messageId: 'suggestAddingAssertions',
              desc: 'Add `expect.assertions(<number of assertions>)`',
              output: 'it("it1", () => {expect.assertions();})',
            },
          ],
        },
      ],
    },
    {
      code: `it("it1", () => { foo()})`,
      errors: [
        {
          messageId: 'haveExpectAssertions',
          message:
            'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              desc: 'Add `expect.hasAssertions()`',
              output: 'it("it1", () => {expect.hasAssertions(); foo()})',
            },
            {
              messageId: 'suggestAddingAssertions',
              desc: 'Add `expect.assertions(<number of assertions>)`',
              output: 'it("it1", () => {expect.assertions(); foo()})',
            },
          ],
        },
      ],
    },
    {
      code: `it("it1", function() {var a = 2;})`,
      errors: [
        {
          messageId: 'haveExpectAssertions',
          message:
            'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 35,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              desc: 'Add `expect.hasAssertions()`',
              output:
                'it("it1", function() {expect.hasAssertions();var a = 2;})',
            },
            {
              messageId: 'suggestAddingAssertions',
              desc: 'Add `expect.assertions(<number of assertions>)`',
              output: 'it("it1", function() {expect.assertions();var a = 2;})',
            },
          ],
        },
      ],
    },
    {
      code: `it("it1", function() {expect.assertions();})`,
      errors: [
        {
          messageId: 'assertionsRequiresOneArgument',
          message:
            '`expect.assertions` expects a single argument of type number',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      code: `it("it1", function() {expect.assertions(1,2);})`,
      errors: [
        {
          messageId: 'assertionsRequiresOneArgument',
          message:
            '`expect.assertions` expects a single argument of type number',
          line: 1,
          column: 43,
          endLine: 1,
          endColumn: 44,
          suggestions: [
            {
              messageId: 'suggestRemovingExtraArguments',
              desc: 'Remove extra arguments',
              output: 'it("it1", function() {expect.assertions(1,);})',
            },
          ],
        },
      ],
    },
    {
      code: `it("it1", function() {expect.assertions(1,2,);})`,
      errors: [
        {
          messageId: 'assertionsRequiresOneArgument',
          message:
            '`expect.assertions` expects a single argument of type number',
          line: 1,
          column: 43,
          endLine: 1,
          endColumn: 44,
          suggestions: [
            {
              messageId: 'suggestRemovingExtraArguments',
              desc: 'Remove extra arguments',
              output: 'it("it1", function() {expect.assertions(1,);})',
            },
          ],
        },
      ],
    },
    {
      code: `it("it1", function() {expect.assertions("1");})`,
      errors: [
        {
          messageId: 'assertionsRequiresNumberArgument',
          message: 'This argument should be a number',
          line: 1,
          column: 41,
          endLine: 1,
          endColumn: 44,
        },
      ],
    },
    {
      code: `it("it1", function() {expect.hasAssertions("1");})`,
      errors: [
        {
          messageId: 'hasAssertionsTakesNoArguments',
          message: '`expect.hasAssertions` expects no arguments',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 43,
          suggestions: [
            {
              messageId: 'suggestRemovingExtraArguments',
              desc: 'Remove extra arguments',
              output: 'it("it1", function() {expect.hasAssertions();})',
            },
          ],
        },
      ],
    },
    {
      code: `it("it1", function() {expect.hasAssertions("1",);})`,
      errors: [
        {
          messageId: 'hasAssertionsTakesNoArguments',
          message: '`expect.hasAssertions` expects no arguments',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 43,
          suggestions: [
            {
              messageId: 'suggestRemovingExtraArguments',
              desc: 'Remove extra arguments',
              output: 'it("it1", function() {expect.hasAssertions();})',
            },
          ],
        },
      ],
    },
    {
      code: `it("it1", function() {expect.hasAssertions("1", "2");})`,
      errors: [
        {
          messageId: 'hasAssertionsTakesNoArguments',
          message: '`expect.hasAssertions` expects no arguments',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 43,
          suggestions: [
            {
              messageId: 'suggestRemovingExtraArguments',
              desc: 'Remove extra arguments',
              output: 'it("it1", function() {expect.hasAssertions();})',
            },
          ],
        },
      ],
    },
    {
      code: `it("it1", () => {
    expect.hasAssertions();

    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(0);
    }
     });

     it("it1", () => {
    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(0);
    }
     });`,
      options: [{ onlyFunctionsWithExpectInLoop: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          message:
            'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
          line: 9,
          column: 6,
          endLine: 13,
          endColumn: 8,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              desc: 'Add `expect.hasAssertions()`',
              output:
                'it("it1", () => {\n    expect.hasAssertions();\n\n    for (const number of getNumbers()) {\n      expect(number).toBeGreaterThan(0);\n    }\n     });\n\n     it("it1", () => {expect.hasAssertions();\n    for (const number of getNumbers()) {\n      expect(number).toBeGreaterThan(0);\n    }\n     });',
            },
            {
              messageId: 'suggestAddingAssertions',
              desc: 'Add `expect.assertions(<number of assertions>)`',
              output:
                'it("it1", () => {\n    expect.hasAssertions();\n\n    for (const number of getNumbers()) {\n      expect(number).toBeGreaterThan(0);\n    }\n     });\n\n     it("it1", () => {expect.assertions();\n    for (const number of getNumbers()) {\n      expect(number).toBeGreaterThan(0);\n    }\n     });',
            },
          ],
        },
      ],
    },
    {
      code: `it("returns numbers that are greater than four", async () => {
     for (const number of await getNumbers()) {
    expect(number).toBeGreaterThan(4);
     }
   });

   it("returns numbers that are greater than five", () => {
     for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(5);
     }
   });
    `,
      options: [{ onlyFunctionsWithExpectInLoop: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          message:
            'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
          line: 1,
          column: 1,
          endLine: 5,
          endColumn: 6,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              desc: 'Add `expect.hasAssertions()`',
              output:
                'it("returns numbers that are greater than four", async () => {expect.hasAssertions();\n     for (const number of await getNumbers()) {\n    expect(number).toBeGreaterThan(4);\n     }\n   });\n\n   it("returns numbers that are greater than five", () => {\n     for (const number of getNumbers()) {\n    expect(number).toBeGreaterThan(5);\n     }\n   });\n    ',
            },
            {
              messageId: 'suggestAddingAssertions',
              desc: 'Add `expect.assertions(<number of assertions>)`',
              output:
                'it("returns numbers that are greater than four", async () => {expect.assertions();\n     for (const number of await getNumbers()) {\n    expect(number).toBeGreaterThan(4);\n     }\n   });\n\n   it("returns numbers that are greater than five", () => {\n     for (const number of getNumbers()) {\n    expect(number).toBeGreaterThan(5);\n     }\n   });\n    ',
            },
          ],
        },
        {
          messageId: 'haveExpectAssertions',
          message:
            'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
          line: 7,
          column: 4,
          endLine: 11,
          endColumn: 6,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              desc: 'Add `expect.hasAssertions()`',
              output:
                'it("returns numbers that are greater than four", async () => {\n     for (const number of await getNumbers()) {\n    expect(number).toBeGreaterThan(4);\n     }\n   });\n\n   it("returns numbers that are greater than five", () => {expect.hasAssertions();\n     for (const number of getNumbers()) {\n    expect(number).toBeGreaterThan(5);\n     }\n   });\n    ',
            },
            {
              messageId: 'suggestAddingAssertions',
              desc: 'Add `expect.assertions(<number of assertions>)`',
              output:
                'it("returns numbers that are greater than four", async () => {\n     for (const number of await getNumbers()) {\n    expect(number).toBeGreaterThan(4);\n     }\n   });\n\n   it("returns numbers that are greater than five", () => {expect.assertions();\n     for (const number of getNumbers()) {\n    expect(number).toBeGreaterThan(5);\n     }\n   });\n    ',
            },
          ],
        },
      ],
    },
    {
      code: `it("it1", () => {
    const foo = { bar({ baz }) { baz(); } };
  });
    `,
      errors: [
        {
          messageId: 'haveExpectAssertions',
          message:
            'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
          line: 1,
          column: 1,
          endLine: 3,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              desc: 'Add `expect.hasAssertions()`',
              output:
                'it("it1", () => {expect.hasAssertions();\n    const foo = { bar({ baz }) { baz(); } };\n  });\n    ',
            },
            {
              messageId: 'suggestAddingAssertions',
              desc: 'Add `expect.assertions(<number of assertions>)`',
              output:
                'it("it1", () => {expect.assertions();\n    const foo = { bar({ baz }) { baz(); } };\n  });\n    ',
            },
          ],
        },
      ],
    },
    {
      code: `it("it1", function() {expect.hasAssertions();})`,
      options: [{ disallowHasAssertions: true }],
      errors: [
        {
          messageId: 'preferAssertionsOverHasAssertions',
          message:
            'Prefer `expect.assertions(<number of assertions>)` over `expect.hasAssertions()`',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 43,
          suggestions: [
            {
              messageId: 'suggestReplacingWithAssertions',
              desc: 'Replace with `expect.assertions(<number of assertions>)`',
              output: 'it("it1", function() {expect.assertions();})',
            },
          ],
        },
      ],
    },
    {
      code: `it('my test description', ({ expect }) => {
  expect.hasAssertions();
  const a = 1;
  expect(a).toBe(1);
})`,
      options: [{ disallowHasAssertions: true }],
      errors: [
        {
          messageId: 'preferAssertionsOverHasAssertions',
          message:
            'Prefer `expect.assertions(<number of assertions>)` over `expect.hasAssertions()`',
          line: 2,
          column: 10,
          endLine: 2,
          endColumn: 23,
          suggestions: [
            {
              messageId: 'suggestReplacingWithAssertions',
              desc: 'Replace with `expect.assertions(<number of assertions>)`',
              output:
                "it('my test description', ({ expect }) => {\n  expect.assertions();\n  const a = 1;\n  expect(a).toBe(1);\n})",
            },
          ],
        },
      ],
    },
    {
      code: `it("it1", () => {})`,
      options: [{ disallowHasAssertions: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          message:
            'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
          suggestions: [
            {
              messageId: 'suggestAddingAssertions',
              desc: 'Add `expect.assertions(<number of assertions>)`',
              output: 'it("it1", () => {expect.assertions();})',
            },
          ],
        },
      ],
    },
  ],
});
