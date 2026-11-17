// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('valid-expect', {
  valid: [
    ...(
      [
        'expect.hasAssertions',
        'expect.hasAssertions()',
        'expect.resolves.toBe(2);',
        'expect("something").toEqual("else");',
        'expect(true).toBeDefined();',
        'expect([1, 2, 3]).toEqual([1, 2, 3]);',
        'expect(undefined).not.toBeDefined();',
        'test("valid-expect", () => { return expect(Promise.resolve(2)).resolves.toBeDefined(); });',
        'test("valid-expect", () => { return expect(Promise.reject(2)).rejects.toBeDefined(); });',
        'test("valid-expect", () => { return expect(Promise.resolve(2)).resolves.not.toBeDefined(); });',
        'test("valid-expect", () => { return expect(Promise.resolve(2)).rejects.not.toBeDefined(); });',
        'test("valid-expect", function () { return expect(Promise.resolve(2)).resolves.not.toBeDefined(); });',
        'test("valid-expect", function () { return expect(Promise.resolve(2)).rejects.not.toBeDefined(); });',
        'test("valid-expect", function () { return Promise.resolve(expect(Promise.resolve(2)).resolves.not.toBeDefined()); });',
        'test("valid-expect", function () { return Promise.resolve(expect(Promise.resolve(2)).rejects.not.toBeDefined()); });',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    {
      name: 'run 0 valid 15',
      code: 'test("valid-expect", () => expect(Promise.resolve(2)).resolves.toBeDefined());',
      options: [
        {
          alwaysAwait: true,
        },
      ],
    },
    ...(
      [
        'test("valid-expect", () => expect(Promise.resolve(2)).resolves.toBeDefined());',
        'test("valid-expect", () => expect(Promise.reject(2)).rejects.toBeDefined());',
        'test("valid-expect", () => expect(Promise.reject(2)).resolves.not.toBeDefined());',
        'test("valid-expect", () => expect(Promise.reject(2)).rejects.not.toBeDefined());',
        'test("valid-expect", async () => { await expect(Promise.reject(2)).resolves.not.toBeDefined(); });',
        'test("valid-expect", async () => { await expect(Promise.reject(2)).rejects.not.toBeDefined(); });',
        'test("valid-expect", async function () { await expect(Promise.reject(2)).resolves.not.toBeDefined(); });',
        'test("valid-expect", async function () { await expect(Promise.reject(2)).rejects.not.toBeDefined(); });',
        'test("valid-expect", async () => { await Promise.resolve(expect(Promise.reject(2)).rejects.not.toBeDefined()); });',
        'test("valid-expect", async () => { await Promise.reject(expect(Promise.reject(2)).rejects.not.toBeDefined()); });',
        'test("valid-expect", async () => { await Promise.all([expect(Promise.reject(2)).rejects.not.toBeDefined(), expect(Promise.reject(2)).rejects.not.toBeDefined()]); });',
        'test("valid-expect", async () => { await Promise.race([expect(Promise.reject(2)).rejects.not.toBeDefined(), expect(Promise.reject(2)).rejects.not.toBeDefined()]); });',
        'test("valid-expect", async () => { await Promise.allSettled([expect(Promise.reject(2)).rejects.not.toBeDefined(), expect(Promise.reject(2)).rejects.not.toBeDefined()]); });',
        'test("valid-expect", async () => { await Promise.any([expect(Promise.reject(2)).rejects.not.toBeDefined(), expect(Promise.reject(2)).rejects.not.toBeDefined()]); });',
        'test("valid-expect", async () => { return expect(Promise.reject(2)).resolves.not.toBeDefined().then(() => console.log("valid-case")); });',
        'test("valid-expect", async () => { return expect(Promise.reject(2)).resolves.not.toBeDefined().then(() => console.log("valid-case")).then(() => console.log("another valid case")); });',
        'test("valid-expect", async () => { return expect(Promise.reject(2)).resolves.not.toBeDefined().catch(() => console.log("valid-case")); });',
        'test("valid-expect", async () => { return expect(Promise.reject(2)).resolves.not.toBeDefined().then(() => console.log("valid-case")).catch(() => console.log("another valid case")); });',
        'test("valid-expect", async () => { return expect(Promise.reject(2)).resolves.not.toBeDefined().then(() => { expect(someMock).toHaveBeenCalledTimes(1); }); });',
        'test("valid-expect", async () => { await expect(Promise.reject(2)).resolves.not.toBeDefined().then(() => console.log("valid-case")); });',
        'test("valid-expect", async () => { await expect(Promise.reject(2)).resolves.not.toBeDefined().then(() => console.log("valid-case")).then(() => console.log("another valid case")); });',
        'test("valid-expect", async () => { await expect(Promise.reject(2)).resolves.not.toBeDefined().catch(() => console.log("valid-case")); });',
        'test("valid-expect", async () => { await expect(Promise.reject(2)).resolves.not.toBeDefined().then(() => console.log("valid-case")).catch(() => console.log("another valid case")); });',
        'test("valid-expect", async () => { await expect(Promise.reject(2)).resolves.not.toBeDefined().then(() => { expect(someMock).toHaveBeenCalledTimes(1); }); });',
        `
      test("valid-expect", () => {
        return expect(functionReturningAPromise()).resolves.toEqual(1).then(() => {
          return expect(Promise.resolve(2)).resolves.toBe(1);
        });
      });
   \u0020`,
        `
      test("valid-expect", () => {
        return expect(functionReturningAPromise()).resolves.toEqual(1).then(async () => {
          await expect(Promise.resolve(2)).resolves.toBe(1);
        });
      });
   \u0020`,
        `
      test("valid-expect", () => {
        return expect(functionReturningAPromise()).resolves.toEqual(1).then(() => expect(Promise.resolve(2)).resolves.toBe(1));
      });
   \u0020`,
        `
      expect.extend({
        toResolve(obj) {
          return this.isNot
            ? expect(obj).toBe(true)
            : expect(obj).resolves.not.toThrow();
        }
      });
   \u0020`,
        `
      expect.extend({
        toResolve(obj) {
          return this.isNot
            ? expect(obj).resolves.not.toThrow()
            : expect(obj).toBe(true);
        }
      });
   \u0020`,
        `
      expect.extend({
        toResolve(obj) {
          return this.isNot
            ? expect(obj).toBe(true)
            : anotherCondition
            ? expect(obj).resolves.not.toThrow()
            : expect(obj).toBe(false)
        }
      });
   \u0020`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index + 16}`,
      code,
    })),
    ...(['expect(1).toBe(2);', 'expect(1, "1 !== 2").toBe(2);'] as const).map(
      (code, index): ValidTestCase => ({
        name: `run 0 valid ${index + 46}`,
        code,
        options: [
          {
            maxArgs: 2,
          },
        ],
      }),
    ),
    {
      name: 'run 0 valid 48',
      code: 'expect(1, "1 !== 2").toBe(2);',
      options: [
        {
          maxArgs: 2,
          minArgs: 2,
        },
      ],
    },
    {
      name: 'run 0 valid 49',
      code: 'test("valid-expect", () => { expect(2).not.toBe(2); });',
      options: [
        {
          asyncMatchers: ['toRejectWith'],
        },
      ],
    },
    ...(
      [
        'test("valid-expect", () => { expect(Promise.reject(2)).toRejectWith(2); });',
        'test("valid-expect", async () => { await expect(Promise.resolve(2)).toResolve(); });',
        'test("valid-expect", async () => { expect(Promise.resolve(2)).toResolve(); });',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index + 50}`,
      code,
      options: [
        {
          asyncMatchers: ['toResolveWith'],
        },
      ],
    })),
    {
      name: 'run 0 valid 53',
      code: 'expect().pass();',
      options: [
        {
          minArgs: 0,
        },
      ],
    },
    ...(
      [
        'expect.hasAssertions',
        'expect.hasAssertions()',
        'expect("something").toEqual("else");',
        'expect(true).toBeDefined();',
        'expect([1, 2, 3]).toEqual([1, 2, 3]);',
        'expect(undefined).not.toBeDefined();',
        'test("valid-expect", () => { return expect(Promise.resolve(2)).resolves.toBeDefined(); });',
        'test("valid-expect", () => { return expect(Promise.reject(2)).rejects.toBeDefined(); });',
        'test("valid-expect", () => { return expect(Promise.resolve(2)).resolves.not.toBeDefined(); });',
        'test("valid-expect", () => { return expect(Promise.resolve(2)).rejects.not.toBeDefined(); });',
        'test("valid-expect", function () { return expect(Promise.resolve(2)).resolves.not.toBeDefined(); });',
        'test("valid-expect", function () { return expect(Promise.resolve(2)).rejects.not.toBeDefined(); });',
        'test("valid-expect", function () { return Promise.resolve(expect(Promise.resolve(2)).resolves.not.toBeDefined()); });',
        'test("valid-expect", function () { return Promise.resolve(expect(Promise.resolve(2)).rejects.not.toBeDefined()); });',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream valid-expect / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
    })),
    {
      name: 'upstream valid-expect / valid 14',
      code: 'test("valid-expect", () => expect(Promise.resolve(2)).resolves.toBeDefined());',
      options: [
        {
          alwaysAwait: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
    },
    ...(
      [
        'test("valid-expect", () => expect(Promise.resolve(2)).resolves.toBeDefined());',
        'test("valid-expect", () => expect(Promise.reject(2)).rejects.toBeDefined());',
        'test("valid-expect", () => expect(Promise.reject(2)).resolves.not.toBeDefined());',
        'test("valid-expect", () => expect(Promise.reject(2)).rejects.not.toBeDefined());',
        'test("valid-expect", async () => { await expect(Promise.reject(2)).resolves.not.toBeDefined(); });',
        'test("valid-expect", async () => { await expect(Promise.reject(2)).rejects.not.toBeDefined(); });',
        'test("valid-expect", async function () { await expect(Promise.reject(2)).resolves.not.toBeDefined(); });',
        'test("valid-expect", async function () { await expect(Promise.reject(2)).rejects.not.toBeDefined(); });',
        'test("valid-expect", async () => { await Promise.resolve(expect(Promise.reject(2)).rejects.not.toBeDefined()); });',
        'test("valid-expect", async () => { await Promise.reject(expect(Promise.reject(2)).rejects.not.toBeDefined()); });',
        'test("valid-expect", async () => { await Promise.all([expect(Promise.reject(2)).rejects.not.toBeDefined(), expect(Promise.reject(2)).rejects.not.toBeDefined()]); });',
        'test("valid-expect", async () => { await Promise.race([expect(Promise.reject(2)).rejects.not.toBeDefined(), expect(Promise.reject(2)).rejects.not.toBeDefined()]); });',
        'test("valid-expect", async () => { await Promise.allSettled([expect(Promise.reject(2)).rejects.not.toBeDefined(), expect(Promise.reject(2)).rejects.not.toBeDefined()]); });',
        'test("valid-expect", async () => { await Promise.any([expect(Promise.reject(2)).rejects.not.toBeDefined(), expect(Promise.reject(2)).rejects.not.toBeDefined()]); });',
        'test("valid-expect", async () => { return expect(Promise.reject(2)).resolves.not.toBeDefined().then(() => console.log("valid-case")); });',
        'test("valid-expect", async () => { return expect(Promise.reject(2)).resolves.not.toBeDefined().then(() => console.log("valid-case")).then(() => console.log("another valid case")); });',
        'test("valid-expect", async () => { return expect(Promise.reject(2)).resolves.not.toBeDefined().catch(() => console.log("valid-case")); });',
        'test("valid-expect", async () => { return expect(Promise.reject(2)).resolves.not.toBeDefined().then(() => console.log("valid-case")).catch(() => console.log("another valid case")); });',
        'test("valid-expect", async () => { return expect(Promise.reject(2)).resolves.not.toBeDefined().then(() => { expect(someMock).toHaveBeenCalledTimes(1); }); });',
        'test("valid-expect", async () => { await expect(Promise.reject(2)).resolves.not.toBeDefined().then(() => console.log("valid-case")); });',
        'test("valid-expect", async () => { await expect(Promise.reject(2)).resolves.not.toBeDefined().then(() => console.log("valid-case")).then(() => console.log("another valid case")); });',
        'test("valid-expect", async () => { await expect(Promise.reject(2)).resolves.not.toBeDefined().catch(() => console.log("valid-case")); });',
        'test("valid-expect", async () => { await expect(Promise.reject(2)).resolves.not.toBeDefined().then(() => console.log("valid-case")).catch(() => console.log("another valid case")); });',
        'test("valid-expect", async () => { await expect(Promise.reject(2)).resolves.not.toBeDefined().then(() => { expect(someMock).toHaveBeenCalledTimes(1); }); });',
        `test("valid-expect", () => {
  return expect(functionReturningAPromise()).resolves.toEqual(1).then(() => {
    return expect(Promise.resolve(2)).resolves.toBe(1);
  });
});`,
        `test("valid-expect", () => {
  return expect(functionReturningAPromise()).resolves.toEqual(1).then(async () => {
    await expect(Promise.resolve(2)).resolves.toBe(1);
  });
});`,
        `test("valid-expect", () => {
  return expect(functionReturningAPromise()).resolves.toEqual(1).then(() => expect(Promise.resolve(2)).resolves.toBe(1));
});`,
        `expect.extend({
  toResolve(obj) {
    return this.isNot
      ? expect(obj).toBe(true)
      : expect(obj).resolves.not.toThrow();
  }
});`,
        `expect.extend({
  toResolve(obj) {
    return this.isNot
      ? expect(obj).resolves.not.toThrow()
      : expect(obj).toBe(true);
  }
});`,
        `expect.extend({
  toResolve(obj) {
    return this.isNot
      ? expect(obj).toBe(true)
      : anotherCondition
      ? expect(obj).resolves.not.toThrow()
      : expect(obj).toBe(false)
  }
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream valid-expect / valid ${index + 15}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
    })),
    ...(['expect(1).toBe(2);', 'expect(1, "1 !== 2").toBe(2);'] as const).map(
      (code, index): ValidTestCase => ({
        name: `upstream valid-expect / valid ${index + 45}`,
        code,
        options: [
          {
            maxArgs: 2,
          },
        ],
        languageOptions: {
          ecmaVersion: 2017,
        },
      }),
    ),
    {
      name: 'upstream valid-expect / valid 47',
      code: 'expect(1, "1 !== 2").toBe(2);',
      options: [
        {
          maxArgs: 2,
          minArgs: 2,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
    },
    {
      name: 'upstream valid-expect / valid 48',
      code: 'test("valid-expect", () => { expect(2).not.toBe(2); });',
      options: [
        {
          asyncMatchers: ['toRejectWith'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
    },
    ...(
      [
        'test("valid-expect", () => { expect(Promise.reject(2)).toRejectWith(2); });',
        'test("valid-expect", async () => { await expect(Promise.resolve(2)).toResolve(); });',
        'test("valid-expect", async () => { expect(Promise.resolve(2)).toResolve(); });',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream valid-expect / valid ${index + 49}`,
      code,
      options: [
        {
          asyncMatchers: ['toResolveWith'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
    })),
    {
      name: 'upstream valid-expect / valid 52',
      code: 'expect().pass();',
      options: [
        {
          minArgs: 0,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: 'expect().toBe(2);',
      options: [
        {
          minArgs: undefined,
          maxArgs: undefined,
        },
      ],
      errors: [
        {
          messageId: 'notEnoughArgs',
          message: 'Expect requires at least 1 argument',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    ...(
      ['expect().toBe(true);', 'expect().toEqual("something");'] as const
    ).map((code, index): TestCase => ({
      name: `run 0 invalid ${index + 1}`,
      code,
      errors: [
        {
          messageId: 'notEnoughArgs',
          message: 'Expect requires at least 1 argument',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 8,
        },
      ],
    })),
    {
      name: 'run 0 invalid 3',
      code: 'expect("something", "else").toEqual("something");',
      errors: [
        {
          messageId: 'tooManyArgs',
          message: 'Expect takes at most 1 argument',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      name: 'run 0 invalid 4',
      code: 'expect("something", "else", "entirely").toEqual("something");',
      options: [
        {
          maxArgs: 2,
        },
      ],
      errors: [
        {
          messageId: 'tooManyArgs',
          message: 'Expect takes at most 2 arguments',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      name: 'run 0 invalid 5',
      code: 'expect("something", "else", "entirely").toEqual("something");',
      options: [
        {
          maxArgs: 2,
          minArgs: 2,
        },
      ],
      errors: [
        {
          messageId: 'tooManyArgs',
          message: 'Expect takes at most 2 arguments',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      name: 'run 0 invalid 6',
      code: 'expect("something", "else", "entirely").toEqual("something");',
      options: [
        {
          maxArgs: 2,
          minArgs: 1,
        },
      ],
      errors: [
        {
          messageId: 'tooManyArgs',
          message: 'Expect takes at most 2 arguments',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      name: 'run 0 invalid 7',
      code: 'expect("something").toEqual("something");',
      options: [
        {
          minArgs: 2,
        },
      ],
      errors: [
        {
          messageId: 'notEnoughArgs',
          message: 'Expect requires at least 2 arguments',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      name: 'run 0 invalid 8',
      code: 'expect("something", "else").toEqual("something");',
      options: [
        {
          maxArgs: 1,
          minArgs: 3,
        },
      ],
      errors: [
        {
          messageId: 'notEnoughArgs',
          message: 'Expect requires at least 3 arguments',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 8,
        },
        {
          messageId: 'tooManyArgs',
          message: 'Expect takes at most 1 argument',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    ...(
      [
        ['expect("something");', 20],
        ['expect();', 9],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index + 9}`,
      code,
      errors: [
        {
          messageId: 'matcherNotFound',
          message: 'Expect must have a corresponding matcher call',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    ...(
      [
        ['expect(true).toBeDefined;', 14],
        ['expect(true).not.toBeDefined;', 18],
        ['expect(true).nope.toBeDefined;', 19],
      ] as const
    ).map(([code, column], index): TestCase => ({
      name: `run 0 invalid ${index + 11}`,
      code,
      errors: [
        {
          messageId: 'matcherNotCalled',
          message: 'Matchers must be called to assert',
          line: 1,
          column,
          endLine: 1,
          endColumn: column + 11,
        },
      ],
    })),
    ...(
      [
        ['expect(true).nope.toBeDefined();', 32],
        ['expect(true).not.resolves.toBeDefined();', 40],
        ['expect(true).not.not.toBeDefined();', 35],
        ['expect(true).resolves.not.exactly.toBeDefined();', 48],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index + 14}`,
      code,
      errors: [
        {
          messageId: 'modifierUnknown',
          message: 'Expect has an unknown modifier',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    ...(
      [
        ['expect(true).resolves;', 22],
        ['expect(true).rejects;', 21],
        ['expect(true).not;', 17],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index + 18}`,
      code,
      errors: [
        {
          messageId: 'matcherNotFound',
          message: 'Expect must have a corresponding matcher call',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    ...(
      [
        ['expect(Promise.resolve(2)).resolves.toBeDefined();', 50],
        ['expect(Promise.resolve(2)).rejects.toBeDefined();', 49],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index + 21}`,
      code,
      errors: [
        {
          messageId: 'asyncMustBeAwaited',
          message: 'Async assertions must be awaited or returned',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    {
      name: 'run 0 invalid 23',
      code: 'expect(Promise.resolve(2)).resolves.toBeDefined();',
      options: [
        {
          alwaysAwait: true,
        },
      ],
      errors: [
        {
          messageId: 'asyncMustBeAwaited',
          message: 'Async assertions must be awaited',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
    ...(
      [
        [
          `
        expect.extend({
          toResolve(obj) {
            this.isNot
              ? expect(obj).toBe(true)
              : expect(obj).resolves.not.toThrow();
          }
        });
     \u0020`,
          6,
          17,
          51,
          `
        expect.extend({
          async toResolve(obj) {
            this.isNot
              ? expect(obj).toBe(true)
              : await expect(obj).resolves.not.toThrow();
          }
        });
     \u0020`,
        ],
        [
          `
        expect.extend({
          toResolve(obj) {
            this.isNot
              ? expect(obj).resolves.not.toThrow()
              : expect(obj).toBe(true);
          }
        });
     \u0020`,
          5,
          17,
          51,
          `
        expect.extend({
          async toResolve(obj) {
            this.isNot
              ? await expect(obj).resolves.not.toThrow()
              : expect(obj).toBe(true);
          }
        });
     \u0020`,
        ],
        [
          `
        expect.extend({
          toResolve(obj) {
            this.isNot
              ? expect(obj).toBe(true)
              : anotherCondition
              ? expect(obj).resolves.not.toThrow()
              : expect(obj).toBe(false)
          }
        });
     \u0020`,
          7,
          17,
          51,
          `
        expect.extend({
          async toResolve(obj) {
            this.isNot
              ? expect(obj).toBe(true)
              : anotherCondition
              ? await expect(obj).resolves.not.toThrow()
              : expect(obj).toBe(false)
          }
        });
     \u0020`,
        ],
        [
          'test("valid-expect", () => { expect(Promise.resolve(2)).resolves.toBeDefined(); });',
          1,
          30,
          79,
          'test("valid-expect", async () => { await expect(Promise.resolve(2)).resolves.toBeDefined(); });',
        ],
        [
          'test("valid-expect", () => { expect(Promise.resolve(2)).toResolve(); });',
          1,
          30,
          68,
          'test("valid-expect", async () => { await expect(Promise.resolve(2)).toResolve(); });',
        ],
      ] as const
    ).map(([code, line, column, endColumn, output], index): TestCase => ({
      name: `run 0 invalid ${index + 24}`,
      code,
      errors: [
        {
          messageId: 'asyncMustBeAwaited',
          message: 'Async assertions must be awaited or returned',
          line,
          column,
          endLine: line,
          endColumn,
        },
      ],
      output,
    })),
    {
      name: 'run 0 invalid 29',
      code: 'test("valid-expect", () => { expect(Promise.resolve(2)).toResolve(); });',
      options: [
        {
          asyncMatchers: undefined,
        },
      ],
      errors: [
        {
          messageId: 'asyncMustBeAwaited',
          message: 'Async assertions must be awaited or returned',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 68,
        },
      ],
      output:
        'test("valid-expect", async () => { await expect(Promise.resolve(2)).toResolve(); });',
    },
    ...(
      [
        [
          'test("valid-expect", () => { expect(Promise.resolve(2)).toReject(); });',
          30,
          67,
          'test("valid-expect", async () => { await expect(Promise.resolve(2)).toReject(); });',
        ],
        [
          'test("valid-expect", () => { expect(Promise.resolve(2)).not.toReject(); });',
          30,
          71,
          'test("valid-expect", async () => { await expect(Promise.resolve(2)).not.toReject(); });',
        ],
        [
          'test("valid-expect", () => { expect(Promise.resolve(2)).resolves.not.toBeDefined(); });',
          30,
          83,
          'test("valid-expect", async () => { await expect(Promise.resolve(2)).resolves.not.toBeDefined(); });',
        ],
        [
          'test("valid-expect", () => { expect(Promise.resolve(2)).rejects.toBeDefined(); });',
          30,
          78,
          'test("valid-expect", async () => { await expect(Promise.resolve(2)).rejects.toBeDefined(); });',
        ],
        [
          'test("valid-expect", () => { expect(Promise.resolve(2)).rejects.not.toBeDefined(); });',
          30,
          82,
          'test("valid-expect", async () => { await expect(Promise.resolve(2)).rejects.not.toBeDefined(); });',
        ],
        [
          'test("valid-expect", async () => { expect(Promise.resolve(2)).resolves.toBeDefined(); });',
          36,
          85,
          'test("valid-expect", async () => { await expect(Promise.resolve(2)).resolves.toBeDefined(); });',
        ],
        [
          'test("valid-expect", async () => { expect(Promise.resolve(2)).resolves.not.toBeDefined(); });',
          36,
          89,
          'test("valid-expect", async () => { await expect(Promise.resolve(2)).resolves.not.toBeDefined(); });',
        ],
      ] as const
    ).map(([code, column, endColumn, output], index): TestCase => ({
      name: `run 0 invalid ${index + 30}`,
      code,
      errors: [
        {
          messageId: 'asyncMustBeAwaited',
          message: 'Async assertions must be awaited or returned',
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    ...(
      [
        [
          'test("valid-expect", () => { expect(Promise.reject(2)).toRejectWith(2); });',
          'test("valid-expect", async () => { await expect(Promise.reject(2)).toRejectWith(2); });',
        ],
        [
          'test("valid-expect", () => { expect(Promise.reject(2)).rejects.toBe(2); });',
          'test("valid-expect", async () => { await expect(Promise.reject(2)).rejects.toBe(2); });',
        ],
      ] as const
    ).map(([code, output], index): TestCase => ({
      name: `run 0 invalid ${index + 37}`,
      code,
      options: [
        {
          asyncMatchers: ['toRejectWith'],
        },
      ],
      errors: [
        {
          messageId: 'asyncMustBeAwaited',
          message: 'Async assertions must be awaited or returned',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 71,
        },
      ],
      output,
    })),
    {
      name: 'run 0 invalid 39',
      code: `
        test("valid-expect", async () => {
          expect(Promise.resolve(2)).resolves.not.toBeDefined();
          expect(Promise.resolve(1)).rejects.toBeDefined();
        });
     \u0020`,
      errors: [
        {
          messageId: 'asyncMustBeAwaited',
          message: 'Async assertions must be awaited or returned',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 64,
        },
        {
          messageId: 'asyncMustBeAwaited',
          message: 'Async assertions must be awaited or returned',
          line: 4,
          column: 11,
          endLine: 4,
          endColumn: 59,
        },
      ],
      output: `
        test("valid-expect", async () => {
          await expect(Promise.resolve(2)).resolves.not.toBeDefined();
          await expect(Promise.resolve(1)).rejects.toBeDefined();
        });
     \u0020`,
    },
    {
      name: 'run 0 invalid 40',
      code: `
        test("valid-expect", async () => {
          await expect(Promise.resolve(2)).resolves.not.toBeDefined();
          expect(Promise.resolve(1)).rejects.toBeDefined();
        });
     \u0020`,
      errors: [
        {
          messageId: 'asyncMustBeAwaited',
          message: 'Async assertions must be awaited or returned',
          line: 4,
          column: 11,
          endLine: 4,
          endColumn: 59,
        },
      ],
      output: `
        test("valid-expect", async () => {
          await expect(Promise.resolve(2)).resolves.not.toBeDefined();
          await expect(Promise.resolve(1)).rejects.toBeDefined();
        });
     \u0020`,
    },
    {
      name: 'run 0 invalid 41',
      code: `
        test("valid-expect", async () => {
          expect(Promise.resolve(2)).resolves.not.toBeDefined();
          return expect(Promise.resolve(1)).rejects.toBeDefined();
        });
     \u0020`,
      options: [
        {
          alwaysAwait: true,
        },
      ],
      errors: [
        {
          messageId: 'asyncMustBeAwaited',
          message: 'Async assertions must be awaited',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 64,
        },
        {
          messageId: 'asyncMustBeAwaited',
          message: 'Async assertions must be awaited',
          line: 4,
          column: 18,
          endLine: 4,
          endColumn: 66,
        },
      ],
      output: `
        test("valid-expect", async () => {
          await expect(Promise.resolve(2)).resolves.not.toBeDefined();
          await expect(Promise.resolve(1)).rejects.toBeDefined();
        });
     \u0020`,
    },
    {
      name: 'run 0 invalid 42',
      code: `
        test("valid-expect", async () => {
          expect(Promise.resolve(2)).resolves.not.toBeDefined();
          return expect(Promise.resolve(1)).rejects.toBeDefined();
        });
     \u0020`,
      errors: [
        {
          messageId: 'asyncMustBeAwaited',
          message: 'Async assertions must be awaited or returned',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 64,
        },
      ],
      output: `
        test("valid-expect", async () => {
          await expect(Promise.resolve(2)).resolves.not.toBeDefined();
          return expect(Promise.resolve(1)).rejects.toBeDefined();
        });
     \u0020`,
    },
    ...(
      [
        [
          `
        test("valid-expect", async () => {
          await expect(Promise.resolve(2)).resolves.not.toBeDefined();
          return expect(Promise.resolve(1)).rejects.toBeDefined();
        });
     \u0020`,
          66,
          `
        test("valid-expect", async () => {
          await expect(Promise.resolve(2)).resolves.not.toBeDefined();
          await expect(Promise.resolve(1)).rejects.toBeDefined();
        });
     \u0020`,
        ],
        [
          `
        test("valid-expect", async () => {
          await expect(Promise.resolve(2)).toResolve();
          return expect(Promise.resolve(1)).toReject();
        });
     \u0020`,
          55,
          `
        test("valid-expect", async () => {
          await expect(Promise.resolve(2)).toResolve();
          await expect(Promise.resolve(1)).toReject();
        });
     \u0020`,
        ],
      ] as const
    ).map(([code, endColumn, output], index): TestCase => ({
      name: `run 0 invalid ${index + 43}`,
      code,
      options: [
        {
          alwaysAwait: true,
        },
      ],
      errors: [
        {
          messageId: 'asyncMustBeAwaited',
          message: 'Async assertions must be awaited',
          line: 4,
          column: 18,
          endLine: 4,
          endColumn,
        },
      ],
      output,
    })),
    ...(
      [
        [
          `
        test("valid-expect", () => {
          Promise.resolve(expect(Promise.resolve(2)).resolves.not.toBeDefined());
        });
     \u0020`,
          81,
          `
        test("valid-expect", async () => {
          await Promise.resolve(expect(Promise.resolve(2)).resolves.not.toBeDefined());
        });
     \u0020`,
        ],
        [
          `
        test("valid-expect", () => {
          Promise.reject(expect(Promise.resolve(2)).resolves.not.toBeDefined());
        });
     \u0020`,
          80,
          `
        test("valid-expect", async () => {
          await Promise.reject(expect(Promise.resolve(2)).resolves.not.toBeDefined());
        });
     \u0020`,
        ],
        [
          `
        test("valid-expect", async () => {
          Promise.reject(expect(Promise.resolve(2)).resolves.not.toBeDefined());
        });
     \u0020`,
          80,
          `
        test("valid-expect", async () => {
          await Promise.reject(expect(Promise.resolve(2)).resolves.not.toBeDefined());
        });
     \u0020`,
        ],
        [
          `
        test("valid-expect", () => {
          Promise.x(expect(Promise.resolve(2)).resolves.not.toBeDefined());
        });
     \u0020`,
          75,
          `
        test("valid-expect", async () => {
          await Promise.x(expect(Promise.resolve(2)).resolves.not.toBeDefined());
        });
     \u0020`,
        ],
      ] as const
    ).map(([code, endColumn, output], index): TestCase => ({
      name: `run 0 invalid ${index + 45}`,
      code,
      errors: [
        {
          messageId: 'promisesWithAsyncAssertionsMustBeAwaited',
          message:
            'Promises which return async assertions must be awaited or returned',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn,
        },
      ],
      output,
    })),
    {
      name: 'run 0 invalid 49',
      code: `
        test("valid-expect", () => {
          Promise.resolve(expect(Promise.resolve(2)).resolves.not.toBeDefined());
        });
     \u0020`,
      options: [
        {
          alwaysAwait: true,
        },
      ],
      errors: [
        {
          messageId: 'promisesWithAsyncAssertionsMustBeAwaited',
          message: 'Promises which return async assertions must be awaited',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 81,
        },
      ],
      output: `
        test("valid-expect", async () => {
          await Promise.resolve(expect(Promise.resolve(2)).resolves.not.toBeDefined());
        });
     \u0020`,
    },
    ...(
      [
        [
          `
        test("valid-expect", () => {
          Promise.all([
            expect(Promise.resolve(2)).resolves.not.toBeDefined(),
            expect(Promise.resolve(3)).resolves.not.toBeDefined(),
          ]);
        });
     \u0020`,
          6,
          13,
          `
        test("valid-expect", async () => {
          await Promise.all([
            expect(Promise.resolve(2)).resolves.not.toBeDefined(),
            expect(Promise.resolve(3)).resolves.not.toBeDefined(),
          ]);
        });
     \u0020`,
        ],
        [
          `
        test("valid-expect", () => {
          Promise.all(([expect(p).resolves.toBe(1), expect(q).resolves.toBe(2)]));
        });
     \u0020`,
          3,
          82,
          `
        test("valid-expect", async () => {
          await Promise.all(([expect(p).resolves.toBe(1), expect(q).resolves.toBe(2)]));
        });
     \u0020`,
        ],
        [
          `
        test("valid-expect", () => {
          Promise.x([
            expect(Promise.resolve(2)).resolves.not.toBeDefined(),
            expect(Promise.resolve(3)).resolves.not.toBeDefined(),
          ]);
        });
     \u0020`,
          6,
          13,
          `
        test("valid-expect", async () => {
          await Promise.x([
            expect(Promise.resolve(2)).resolves.not.toBeDefined(),
            expect(Promise.resolve(3)).resolves.not.toBeDefined(),
          ]);
        });
     \u0020`,
        ],
      ] as const
    ).map(([code, endLine, endColumn, output], index): TestCase => ({
      name: `run 0 invalid ${index + 50}`,
      code,
      errors: [
        {
          messageId: 'promisesWithAsyncAssertionsMustBeAwaited',
          message:
            'Promises which return async assertions must be awaited or returned',
          line: 3,
          column: 11,
          endLine,
          endColumn,
        },
      ],
      output,
    })),
    ...(
      [
        [
          `
        test("valid-expect", () => {
          const assertions = [
            expect(Promise.resolve(2)).resolves.not.toBeDefined(),
            expect(Promise.resolve(3)).resolves.not.toBeDefined(),
          ]
        });
     \u0020`,
          66,
          66,
          `
        test("valid-expect", async () => {
          const assertions = [
            await expect(Promise.resolve(2)).resolves.not.toBeDefined(),
            await expect(Promise.resolve(3)).resolves.not.toBeDefined(),
          ]
        });
     \u0020`,
        ],
        [
          `
        test("valid-expect", () => {
          const assertions = [
            expect(Promise.resolve(2)).toResolve(),
            expect(Promise.resolve(3)).toReject(),
          ]
        });
     \u0020`,
          51,
          50,
          `
        test("valid-expect", async () => {
          const assertions = [
            await expect(Promise.resolve(2)).toResolve(),
            await expect(Promise.resolve(3)).toReject(),
          ]
        });
     \u0020`,
        ],
        [
          `
        test("valid-expect", () => {
          const assertions = [
            expect(Promise.resolve(2)).not.toResolve(),
            expect(Promise.resolve(3)).resolves.toReject(),
          ]
        });
     \u0020`,
          55,
          59,
          `
        test("valid-expect", async () => {
          const assertions = [
            await expect(Promise.resolve(2)).not.toResolve(),
            await expect(Promise.resolve(3)).resolves.toReject(),
          ]
        });
     \u0020`,
        ],
      ] as const
    ).map(([code, endColumn1, endColumn2, output], index): TestCase => ({
      name: `run 0 invalid ${index + 53}`,
      code,
      errors: [
        {
          messageId: 'asyncMustBeAwaited',
          message: 'Async assertions must be awaited or returned',
          line: 4,
          column: 13,
          endLine: 4,
          endColumn: endColumn1,
        },
        {
          messageId: 'asyncMustBeAwaited',
          message: 'Async assertions must be awaited or returned',
          line: 5,
          column: 13,
          endLine: 5,
          endColumn: endColumn2,
        },
      ],
      output,
    })),
    {
      name: 'run 0 invalid 56',
      code: 'expect(Promise.resolve(2)).resolves.toBe;',
      errors: [
        {
          messageId: 'matcherNotCalled',
          message: 'Matchers must be called to assert',
          line: 1,
          column: 37,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    ...(
      [
        [
          `
        test("valid-expect", () => {
          return expect(functionReturningAPromise()).resolves.toEqual(1).then(() => {
            expect(Promise.resolve(2)).resolves.toBe(1);
          });
        });
     \u0020`,
          4,
          `
        test("valid-expect", () => {
          return expect(functionReturningAPromise()).resolves.toEqual(1).then(async () => {
            await expect(Promise.resolve(2)).resolves.toBe(1);
          });
        });
     \u0020`,
        ],
        [
          `
        test("valid-expect", () => {
          return expect(functionReturningAPromise()).resolves.toEqual(1).then(async () => {
            await expect(Promise.resolve(2)).resolves.toBe(1);
            expect(Promise.resolve(4)).resolves.toBe(4);
          });
        });
     \u0020`,
          5,
          `
        test("valid-expect", () => {
          return expect(functionReturningAPromise()).resolves.toEqual(1).then(async () => {
            await expect(Promise.resolve(2)).resolves.toBe(1);
            await expect(Promise.resolve(4)).resolves.toBe(4);
          });
        });
     \u0020`,
        ],
      ] as const
    ).map(([code, line, output], index): TestCase => ({
      name: `run 0 invalid ${index + 57}`,
      code,
      errors: [
        {
          messageId: 'asyncMustBeAwaited',
          message: 'Async assertions must be awaited or returned',
          line,
          column: 13,
          endLine: line,
          endColumn: 56,
        },
      ],
      output,
    })),
    {
      name: 'run 0 invalid 59',
      code: `
        test("valid-expect", async () => {
          await expect(Promise.resolve(1));
        });
     \u0020`,
      errors: [
        {
          messageId: 'matcherNotFound',
          message: 'Expect must have a corresponding matcher call',
          line: 3,
          column: 17,
          endLine: 3,
          endColumn: 43,
        },
      ],
    },
    {
      name: 'run 0 invalid 60',
      code: 'expect(true).assertions;',
      errors: [
        {
          messageId: 'matcherNotCalled',
          message: 'Matchers must be called to assert',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      name: 'run 0 invalid 61',
      code: `
        test("valid-expect", () => {
          Promise.all.x([
            expect(Promise.resolve(2)).resolves.not.toBeDefined(),
            expect(Promise.resolve(3)).resolves.not.toBeDefined(),
          ]);
        });
     \u0020`,
      errors: [
        {
          messageId: 'asyncMustBeAwaited',
          message: 'Async assertions must be awaited or returned',
          line: 4,
          column: 13,
          endLine: 4,
          endColumn: 66,
        },
        {
          messageId: 'asyncMustBeAwaited',
          message: 'Async assertions must be awaited or returned',
          line: 5,
          column: 13,
          endLine: 5,
          endColumn: 66,
        },
      ],
      output: `
        test("valid-expect", async () => {
          Promise.all.x([
            await expect(Promise.resolve(2)).resolves.not.toBeDefined(),
            await expect(Promise.resolve(3)).resolves.not.toBeDefined(),
          ]);
        });
     \u0020`,
    },
    {
      name: 'upstream valid-expect / invalid 0',
      code: 'expect().toBe(2);',
      options: [
        {
          minArgs: undefined,
          maxArgs: undefined,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'notEnoughArgs',
          message: 'Expect requires at least 1 argument',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    ...(
      ['expect().toBe(true);', 'expect().toEqual("something");'] as const
    ).map((code, index): TestCase => ({
      name: `upstream valid-expect / invalid ${index + 1}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'notEnoughArgs',
          message: 'Expect requires at least 1 argument',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 8,
        },
      ],
    })),
    {
      name: 'upstream valid-expect / invalid 3',
      code: 'expect("something", "else").toEqual("something");',
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'tooManyArgs',
          message: 'Expect takes at most 1 argument',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      name: 'upstream valid-expect / invalid 4',
      code: 'expect("something", "else", "entirely").toEqual("something");',
      options: [
        {
          maxArgs: 2,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'tooManyArgs',
          message: 'Expect takes at most 2 arguments',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      name: 'upstream valid-expect / invalid 5',
      code: 'expect("something", "else", "entirely").toEqual("something");',
      options: [
        {
          maxArgs: 2,
          minArgs: 2,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'tooManyArgs',
          message: 'Expect takes at most 2 arguments',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      name: 'upstream valid-expect / invalid 6',
      code: 'expect("something", "else", "entirely").toEqual("something");',
      options: [
        {
          maxArgs: 2,
          minArgs: 1,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'tooManyArgs',
          message: 'Expect takes at most 2 arguments',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      name: 'upstream valid-expect / invalid 7',
      code: 'expect("something").toEqual("something");',
      options: [
        {
          minArgs: 2,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'notEnoughArgs',
          message: 'Expect requires at least 2 arguments',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      name: 'upstream valid-expect / invalid 8',
      code: 'expect("something", "else").toEqual("something");',
      options: [
        {
          maxArgs: 1,
          minArgs: 3,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'notEnoughArgs',
          message: 'Expect requires at least 3 arguments',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 8,
        },
        {
          messageId: 'tooManyArgs',
          message: 'Expect takes at most 1 argument',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    ...(
      [
        ['expect("something");', 20],
        ['expect();', 9],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `upstream valid-expect / invalid ${index + 9}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'matcherNotFound',
          message: 'Expect must have a corresponding matcher call',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    ...(
      [
        ['expect(true).toBeDefined;', 14],
        ['expect(true).not.toBeDefined;', 18],
        ['expect(true).nope.toBeDefined;', 19],
      ] as const
    ).map(([code, column], index): TestCase => ({
      name: `upstream valid-expect / invalid ${index + 11}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'matcherNotCalled',
          message: 'Matchers must be called to assert',
          line: 1,
          column,
          endLine: 1,
          endColumn: column + 11,
        },
      ],
    })),
    ...(
      [
        ['expect(true).nope.toBeDefined();', 32],
        ['expect(true).not.resolves.toBeDefined();', 40],
        ['expect(true).not.not.toBeDefined();', 35],
        ['expect(true).resolves.not.exactly.toBeDefined();', 48],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `upstream valid-expect / invalid ${index + 14}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'modifierUnknown',
          message: 'Expect has an unknown modifier',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    ...(
      [
        ['expect(true).resolves;', 22],
        ['expect(true).rejects;', 21],
        ['expect(true).not;', 17],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `upstream valid-expect / invalid ${index + 18}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'matcherNotFound',
          message: 'Expect must have a corresponding matcher call',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    ...(
      [
        ['expect(Promise.resolve(2)).resolves.toBeDefined();', 50],
        ['expect(Promise.resolve(2)).rejects.toBeDefined();', 49],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `upstream valid-expect / invalid ${index + 21}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'asyncMustBeAwaited',
          message: 'Async assertions must be awaited or returned',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    {
      name: 'upstream valid-expect / invalid 23',
      code: 'expect(Promise.resolve(2)).resolves.toBeDefined();',
      options: [
        {
          alwaysAwait: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'asyncMustBeAwaited',
          message: 'Async assertions must be awaited',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
    ...(
      [
        [
          `expect.extend({
  toResolve(obj) {
    this.isNot
      ? expect(obj).toBe(true)
      : expect(obj).resolves.not.toThrow();
  }
});`,
          5,
          9,
          43,
          `expect.extend({
  async toResolve(obj) {
    this.isNot
      ? expect(obj).toBe(true)
      : await expect(obj).resolves.not.toThrow();
  }
});`,
        ],
        [
          `expect.extend({
  toResolve(obj) {
    this.isNot
      ? expect(obj).resolves.not.toThrow()
      : expect(obj).toBe(true);
  }
});`,
          4,
          9,
          43,
          `expect.extend({
  async toResolve(obj) {
    this.isNot
      ? await expect(obj).resolves.not.toThrow()
      : expect(obj).toBe(true);
  }
});`,
        ],
        [
          `expect.extend({
  toResolve(obj) {
    this.isNot
      ? expect(obj).toBe(true)
      : anotherCondition
      ? expect(obj).resolves.not.toThrow()
      : expect(obj).toBe(false)
  }
});`,
          6,
          9,
          43,
          `expect.extend({
  async toResolve(obj) {
    this.isNot
      ? expect(obj).toBe(true)
      : anotherCondition
      ? await expect(obj).resolves.not.toThrow()
      : expect(obj).toBe(false)
  }
});`,
        ],
        [
          'test("valid-expect", () => { expect(Promise.resolve(2)).resolves.toBeDefined(); });',
          1,
          30,
          79,
          'test("valid-expect", async () => { await expect(Promise.resolve(2)).resolves.toBeDefined(); });',
        ],
        [
          'test("valid-expect", () => { expect(Promise.resolve(2)).toResolve(); });',
          1,
          30,
          68,
          'test("valid-expect", async () => { await expect(Promise.resolve(2)).toResolve(); });',
        ],
      ] as const
    ).map(([code, line, column, endColumn, output], index): TestCase => ({
      name: `upstream valid-expect / invalid ${index + 24}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'asyncMustBeAwaited',
          message: 'Async assertions must be awaited or returned',
          line,
          column,
          endLine: line,
          endColumn,
        },
      ],
      output,
    })),
    {
      name: 'upstream valid-expect / invalid 29',
      code: 'test("valid-expect", () => { expect(Promise.resolve(2)).toResolve(); });',
      options: [
        {
          asyncMatchers: undefined,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'asyncMustBeAwaited',
          message: 'Async assertions must be awaited or returned',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 68,
        },
      ],
      output:
        'test("valid-expect", async () => { await expect(Promise.resolve(2)).toResolve(); });',
    },
    ...(
      [
        [
          'test("valid-expect", () => { expect(Promise.resolve(2)).toReject(); });',
          30,
          67,
          'test("valid-expect", async () => { await expect(Promise.resolve(2)).toReject(); });',
        ],
        [
          'test("valid-expect", () => { expect(Promise.resolve(2)).not.toReject(); });',
          30,
          71,
          'test("valid-expect", async () => { await expect(Promise.resolve(2)).not.toReject(); });',
        ],
        [
          'test("valid-expect", () => { expect(Promise.resolve(2)).resolves.not.toBeDefined(); });',
          30,
          83,
          'test("valid-expect", async () => { await expect(Promise.resolve(2)).resolves.not.toBeDefined(); });',
        ],
        [
          'test("valid-expect", () => { expect(Promise.resolve(2)).rejects.toBeDefined(); });',
          30,
          78,
          'test("valid-expect", async () => { await expect(Promise.resolve(2)).rejects.toBeDefined(); });',
        ],
        [
          'test("valid-expect", () => { expect(Promise.resolve(2)).rejects.not.toBeDefined(); });',
          30,
          82,
          'test("valid-expect", async () => { await expect(Promise.resolve(2)).rejects.not.toBeDefined(); });',
        ],
        [
          'test("valid-expect", async () => { expect(Promise.resolve(2)).resolves.toBeDefined(); });',
          36,
          85,
          'test("valid-expect", async () => { await expect(Promise.resolve(2)).resolves.toBeDefined(); });',
        ],
        [
          'test("valid-expect", async () => { expect(Promise.resolve(2)).resolves.not.toBeDefined(); });',
          36,
          89,
          'test("valid-expect", async () => { await expect(Promise.resolve(2)).resolves.not.toBeDefined(); });',
        ],
      ] as const
    ).map(([code, column, endColumn, output], index): TestCase => ({
      name: `upstream valid-expect / invalid ${index + 30}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'asyncMustBeAwaited',
          message: 'Async assertions must be awaited or returned',
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    ...(
      [
        [
          'test("valid-expect", () => { expect(Promise.reject(2)).toRejectWith(2); });',
          'test("valid-expect", async () => { await expect(Promise.reject(2)).toRejectWith(2); });',
        ],
        [
          'test("valid-expect", () => { expect(Promise.reject(2)).rejects.toBe(2); });',
          'test("valid-expect", async () => { await expect(Promise.reject(2)).rejects.toBe(2); });',
        ],
      ] as const
    ).map(([code, output], index): TestCase => ({
      name: `upstream valid-expect / invalid ${index + 37}`,
      code,
      options: [
        {
          asyncMatchers: ['toRejectWith'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'asyncMustBeAwaited',
          message: 'Async assertions must be awaited or returned',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 71,
        },
      ],
      output,
    })),
    {
      name: 'upstream valid-expect / invalid 39',
      code: `test("valid-expect", async () => {
  expect(Promise.resolve(2)).resolves.not.toBeDefined();
  expect(Promise.resolve(1)).rejects.toBeDefined();
});`,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'asyncMustBeAwaited',
          message: 'Async assertions must be awaited or returned',
          line: 2,
          column: 3,
          endLine: 2,
          endColumn: 56,
        },
        {
          messageId: 'asyncMustBeAwaited',
          message: 'Async assertions must be awaited or returned',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 51,
        },
      ],
      output: `test("valid-expect", async () => {
  await expect(Promise.resolve(2)).resolves.not.toBeDefined();
  await expect(Promise.resolve(1)).rejects.toBeDefined();
});`,
    },
    {
      name: 'upstream valid-expect / invalid 40',
      code: `test("valid-expect", async () => {
  await expect(Promise.resolve(2)).resolves.not.toBeDefined();
  expect(Promise.resolve(1)).rejects.toBeDefined();
});`,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'asyncMustBeAwaited',
          message: 'Async assertions must be awaited or returned',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 51,
        },
      ],
      output: `test("valid-expect", async () => {
  await expect(Promise.resolve(2)).resolves.not.toBeDefined();
  await expect(Promise.resolve(1)).rejects.toBeDefined();
});`,
    },
    {
      name: 'upstream valid-expect / invalid 41',
      code: `test("valid-expect", async () => {
  expect(Promise.resolve(2)).resolves.not.toBeDefined();
  return expect(Promise.resolve(1)).rejects.toBeDefined();
});`,
      options: [
        {
          alwaysAwait: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'asyncMustBeAwaited',
          message: 'Async assertions must be awaited',
          line: 2,
          column: 3,
          endLine: 2,
          endColumn: 56,
        },
        {
          messageId: 'asyncMustBeAwaited',
          message: 'Async assertions must be awaited',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 58,
        },
      ],
      output: `test("valid-expect", async () => {
  await expect(Promise.resolve(2)).resolves.not.toBeDefined();
  await expect(Promise.resolve(1)).rejects.toBeDefined();
});`,
    },
    {
      name: 'upstream valid-expect / invalid 42',
      code: `test("valid-expect", async () => {
  expect(Promise.resolve(2)).resolves.not.toBeDefined();
  return expect(Promise.resolve(1)).rejects.toBeDefined();
});`,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'asyncMustBeAwaited',
          message: 'Async assertions must be awaited or returned',
          line: 2,
          column: 3,
          endLine: 2,
          endColumn: 56,
        },
      ],
      output: `test("valid-expect", async () => {
  await expect(Promise.resolve(2)).resolves.not.toBeDefined();
  return expect(Promise.resolve(1)).rejects.toBeDefined();
});`,
    },
    ...(
      [
        [
          `test("valid-expect", async () => {
  await expect(Promise.resolve(2)).resolves.not.toBeDefined();
  return expect(Promise.resolve(1)).rejects.toBeDefined();
});`,
          58,
          `test("valid-expect", async () => {
  await expect(Promise.resolve(2)).resolves.not.toBeDefined();
  await expect(Promise.resolve(1)).rejects.toBeDefined();
});`,
        ],
        [
          `test("valid-expect", async () => {
  await expect(Promise.resolve(2)).toResolve();
  return expect(Promise.resolve(1)).toReject();
});`,
          47,
          `test("valid-expect", async () => {
  await expect(Promise.resolve(2)).toResolve();
  await expect(Promise.resolve(1)).toReject();
});`,
        ],
      ] as const
    ).map(([code, endColumn, output], index): TestCase => ({
      name: `upstream valid-expect / invalid ${index + 43}`,
      code,
      options: [
        {
          alwaysAwait: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'asyncMustBeAwaited',
          message: 'Async assertions must be awaited',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn,
        },
      ],
      output,
    })),
    ...(
      [
        [
          `test("valid-expect", () => {
  Promise.resolve(expect(Promise.resolve(2)).resolves.not.toBeDefined());
});`,
          73,
          `test("valid-expect", async () => {
  await Promise.resolve(expect(Promise.resolve(2)).resolves.not.toBeDefined());
});`,
        ],
        [
          `test("valid-expect", () => {
  Promise.reject(expect(Promise.resolve(2)).resolves.not.toBeDefined());
});`,
          72,
          `test("valid-expect", async () => {
  await Promise.reject(expect(Promise.resolve(2)).resolves.not.toBeDefined());
});`,
        ],
        [
          `test("valid-expect", async () => {
  Promise.reject(expect(Promise.resolve(2)).resolves.not.toBeDefined());
});`,
          72,
          `test("valid-expect", async () => {
  await Promise.reject(expect(Promise.resolve(2)).resolves.not.toBeDefined());
});`,
        ],
        [
          `test("valid-expect", () => {
  Promise.x(expect(Promise.resolve(2)).resolves.not.toBeDefined());
});`,
          67,
          `test("valid-expect", async () => {
  await Promise.x(expect(Promise.resolve(2)).resolves.not.toBeDefined());
});`,
        ],
      ] as const
    ).map(([code, endColumn, output], index): TestCase => ({
      name: `upstream valid-expect / invalid ${index + 45}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'promisesWithAsyncAssertionsMustBeAwaited',
          message:
            'Promises which return async assertions must be awaited or returned',
          line: 2,
          column: 3,
          endLine: 2,
          endColumn,
        },
      ],
      output,
    })),
    {
      name: 'upstream valid-expect / invalid 49',
      code: `test("valid-expect", () => {
  Promise.resolve(expect(Promise.resolve(2)).resolves.not.toBeDefined());
});`,
      options: [
        {
          alwaysAwait: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'promisesWithAsyncAssertionsMustBeAwaited',
          message: 'Promises which return async assertions must be awaited',
          line: 2,
          column: 3,
          endLine: 2,
          endColumn: 73,
        },
      ],
      output: `test("valid-expect", async () => {
  await Promise.resolve(expect(Promise.resolve(2)).resolves.not.toBeDefined());
});`,
    },
    ...(
      [
        [
          `test("valid-expect", () => {
  Promise.all([
    expect(Promise.resolve(2)).resolves.not.toBeDefined(),
    expect(Promise.resolve(3)).resolves.not.toBeDefined(),
  ]);
});`,
          `test("valid-expect", async () => {
  await Promise.all([
    expect(Promise.resolve(2)).resolves.not.toBeDefined(),
    expect(Promise.resolve(3)).resolves.not.toBeDefined(),
  ]);
});`,
        ],
        [
          `test("valid-expect", () => {
  Promise.x([
    expect(Promise.resolve(2)).resolves.not.toBeDefined(),
    expect(Promise.resolve(3)).resolves.not.toBeDefined(),
  ]);
});`,
          `test("valid-expect", async () => {
  await Promise.x([
    expect(Promise.resolve(2)).resolves.not.toBeDefined(),
    expect(Promise.resolve(3)).resolves.not.toBeDefined(),
  ]);
});`,
        ],
      ] as const
    ).map(([code, output], index): TestCase => ({
      name: `upstream valid-expect / invalid ${index + 50}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'promisesWithAsyncAssertionsMustBeAwaited',
          message:
            'Promises which return async assertions must be awaited or returned',
          line: 2,
          column: 3,
          endLine: 5,
          endColumn: 5,
        },
      ],
      output,
    })),
    ...(
      [
        [
          `test("valid-expect", () => {
  const assertions = [
    expect(Promise.resolve(2)).resolves.not.toBeDefined(),
    expect(Promise.resolve(3)).resolves.not.toBeDefined(),
  ]
});`,
          58,
          58,
          `test("valid-expect", async () => {
  const assertions = [
    await expect(Promise.resolve(2)).resolves.not.toBeDefined(),
    await expect(Promise.resolve(3)).resolves.not.toBeDefined(),
  ]
});`,
        ],
        [
          `test("valid-expect", () => {
  const assertions = [
    expect(Promise.resolve(2)).toResolve(),
    expect(Promise.resolve(3)).toReject(),
  ]
});`,
          43,
          42,
          `test("valid-expect", async () => {
  const assertions = [
    await expect(Promise.resolve(2)).toResolve(),
    await expect(Promise.resolve(3)).toReject(),
  ]
});`,
        ],
        [
          `test("valid-expect", () => {
  const assertions = [
    expect(Promise.resolve(2)).not.toResolve(),
    expect(Promise.resolve(3)).resolves.toReject(),
  ]
});`,
          47,
          51,
          `test("valid-expect", async () => {
  const assertions = [
    await expect(Promise.resolve(2)).not.toResolve(),
    await expect(Promise.resolve(3)).resolves.toReject(),
  ]
});`,
        ],
      ] as const
    ).map(([code, endColumn1, endColumn2, output], index): TestCase => ({
      name: `upstream valid-expect / invalid ${index + 52}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'asyncMustBeAwaited',
          message: 'Async assertions must be awaited or returned',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: endColumn1,
        },
        {
          messageId: 'asyncMustBeAwaited',
          message: 'Async assertions must be awaited or returned',
          line: 4,
          column: 5,
          endLine: 4,
          endColumn: endColumn2,
        },
      ],
      output,
    })),
    {
      name: 'upstream valid-expect / invalid 55',
      code: 'expect(Promise.resolve(2)).resolves.toBe;',
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'matcherNotCalled',
          message: 'Matchers must be called to assert',
          line: 1,
          column: 37,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    ...(
      [
        [
          `test("valid-expect", () => {
  return expect(functionReturningAPromise()).resolves.toEqual(1).then(() => {
    expect(Promise.resolve(2)).resolves.toBe(1);
  });
});`,
          3,
          `test("valid-expect", () => {
  return expect(functionReturningAPromise()).resolves.toEqual(1).then(async () => {
    await expect(Promise.resolve(2)).resolves.toBe(1);
  });
});`,
        ],
        [
          `test("valid-expect", () => {
  return expect(functionReturningAPromise()).resolves.toEqual(1).then(async () => {
    await expect(Promise.resolve(2)).resolves.toBe(1);
    expect(Promise.resolve(4)).resolves.toBe(4);
  });
});`,
          4,
          `test("valid-expect", () => {
  return expect(functionReturningAPromise()).resolves.toEqual(1).then(async () => {
    await expect(Promise.resolve(2)).resolves.toBe(1);
    await expect(Promise.resolve(4)).resolves.toBe(4);
  });
});`,
        ],
      ] as const
    ).map(([code, line, output], index): TestCase => ({
      name: `upstream valid-expect / invalid ${index + 56}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'asyncMustBeAwaited',
          message: 'Async assertions must be awaited or returned',
          line,
          column: 5,
          endLine: line,
          endColumn: 48,
        },
      ],
      output,
    })),
    {
      name: 'upstream valid-expect / invalid 58',
      code: `test("valid-expect", async () => {
  await expect(Promise.resolve(1));
});`,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'matcherNotFound',
          message: 'Expect must have a corresponding matcher call',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 35,
        },
      ],
    },
  ],
});
