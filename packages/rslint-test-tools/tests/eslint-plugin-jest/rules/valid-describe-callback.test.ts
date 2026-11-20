// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('valid-describe-callback', {
  valid: [
    ...(
      [
        "describe.each([1, 2, 3])('%s', (a, b) => {});",
        "describe('foo', function() {})",
        "describe('foo', () => {})",
        'describe(`foo`, () => {})',
        "xdescribe('foo', () => {})",
        "fdescribe('foo', () => {})",
        "describe.only('foo', () => {})",
        "describe.skip('foo', () => {})",
        `describe('foo', () => {
        it('bar', () => {
          return Promise.resolve(42).then(value => {
            expect(value).toBe(42)
          })
        })
      })`,
        `describe('foo', () => {
        it('bar', async () => {
          expect(await Promise.resolve(42)).toBe(42)
        })
      })`,
        'if (hasOwnProperty(obj, key)) {}',
        `describe.each\`
          foo  | foe
          \${'1'} | \${'2'}
        \`('$something', ({ foo, foe }) => {});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    ...(
      [
        'describe.each([1, 2, 3])("%s", (a, b) => {});',
        'describe("foo", function() {})',
        'describe("foo", () => {})',
        'describe(`foo`, () => {})',
        'xdescribe("foo", () => {})',
        'fdescribe("foo", () => {})',
        'describe.only("foo", () => {})',
        'describe.skip("foo", () => {})',
        `describe('foo', () => {
  it('bar', () => {
    return Promise.resolve(42).then(value => {
      expect(value).toBe(42)
    })
  })
})`,
        `describe('foo', () => {
  it('bar', async () => {
    expect(await Promise.resolve(42)).toBe(42)
  })
})`,
        `if (hasOwnProperty(obj, key)) {
}`,
        `describe.each\`
  foo  | foe
  1 | 2
\`('$something', ({ foo, foe }) => {});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream valid-describe-callback / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
    })),
  ],
  invalid: [
    ...(
      [
        ['describe.each()()', 1, 18],
        ["describe['each']()()", 1, 21],
        ['describe.each(() => {})()', 1, 26],
        ["describe.each(() => {})('foo')", 25, 30],
        ['describe.each()(() => {})', 17, 25],
        ["describe['each']()(() => {})", 20, 28],
        ["describe.each('foo')(() => {})", 22, 30],
        ["describe.only.each('foo')(() => {})", 27, 35],
        ['describe(() => {})', 10, 18],
        ["describe('foo')", 10, 15],
      ] as const
    ).map(([code, column, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'nameAndCallback',
          message: 'Describe requires name and callback arguments',
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    {
      name: 'run 0 invalid 10',
      code: "describe('foo', 'foo2')",
      errors: [
        {
          messageId: 'secondArgumentMustBeFunction',
          message: 'Second argument must be function',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      name: 'run 0 invalid 11',
      code: 'describe()',
      errors: [
        {
          messageId: 'nameAndCallback',
          message: 'Describe requires name and callback arguments',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    ...(
      [
        ["describe('foo', async () => {})", 1, 17, 1, 31],
        ["describe('foo', async function () {})", 1, 17, 1, 37],
        ["xdescribe('foo', async function () {})", 1, 18, 1, 38],
        ["fdescribe('foo', async function () {})", 1, 18, 1, 38],
        [
          `
        import { fdescribe } from '@jest/globals';
        fdescribe('foo', async function () {})
     \u0020`,
          3,
          26,
          3,
          46,
        ],
        ["describe.only('foo', async function () {})", 1, 22, 1, 42],
        ["describe.skip('foo', async function () {})", 1, 22, 1, 42],
        [
          `describe('sample case', () => {
        it('works', () => {
          expect(true).toEqual(true);
        });
        describe('async', async () => {
          await new Promise(setImmediate);
          it('breaks', () => {
            throw new Error('Fail');
          });
        });
      });`,
          5,
          27,
          10,
          10,
        ],
      ] as const
    ).map(([code, line, column, endLine, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index + 12}`,
      code,
      errors: [
        {
          messageId: 'noAsyncDescribeCallback',
          message: 'No async describe callback',
          line,
          column,
          endLine,
          endColumn,
        },
      ],
    })),
    {
      name: 'run 0 invalid 20',
      code: `describe('foo', function () {
        return Promise.resolve().then(() => {
          it('breaks', () => {
            throw new Error('Fail')
          })
        })
      })`,
      errors: [
        {
          messageId: 'unexpectedReturnInDescribe',
          message: 'Unexpected return statement in describe callback',
          line: 2,
          column: 9,
          endLine: 6,
          endColumn: 11,
        },
      ],
    },
    {
      name: 'run 0 invalid 21',
      code: `describe('foo', () => {
        return Promise.resolve().then(() => {
          it('breaks', () => {
            throw new Error('Fail')
          })
        })
        describe('nested', () => {
          return Promise.resolve().then(() => {
            it('breaks', () => {
              throw new Error('Fail')
            })
          })
        })
      })`,
      errors: [
        {
          messageId: 'unexpectedReturnInDescribe',
          message: 'Unexpected return statement in describe callback',
          line: 2,
          column: 9,
          endLine: 6,
          endColumn: 11,
        },
        {
          messageId: 'unexpectedReturnInDescribe',
          message: 'Unexpected return statement in describe callback',
          line: 8,
          column: 11,
          endLine: 12,
          endColumn: 13,
        },
      ],
    },
    {
      name: 'run 0 invalid 22',
      code: `describe('foo', async () => {
        await something()
        it('does something')
        describe('nested', () => {
          return Promise.resolve().then(() => {
            it('breaks', () => {
              throw new Error('Fail')
            })
          })
        })
      })`,
      errors: [
        {
          messageId: 'noAsyncDescribeCallback',
          message: 'No async describe callback',
          line: 1,
          column: 17,
          endLine: 11,
          endColumn: 8,
        },
        {
          messageId: 'unexpectedReturnInDescribe',
          message: 'Unexpected return statement in describe callback',
          line: 5,
          column: 11,
          endLine: 9,
          endColumn: 13,
        },
      ],
    },
    {
      name: 'run 0 invalid 23',
      code: "describe('foo', () => test('bar', () => {}))",
      errors: [
        {
          messageId: 'unexpectedReturnInDescribe',
          message: 'Unexpected return statement in describe callback',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 44,
        },
      ],
    },
    ...(
      [
        ["describe('foo', done => {})", 17, 21],
        ["describe('foo', function (done) {})", 27, 31],
        ["describe('foo', function (one, two, three) {})", 27, 30],
      ] as const
    ).map(([code, column, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index + 24}`,
      code,
      errors: [
        {
          messageId: 'unexpectedDescribeArgument',
          message: 'Unexpected argument(s) in describe callback',
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    {
      name: 'run 0 invalid 27',
      code: "describe('foo', async function (done) {})",
      errors: [
        {
          messageId: 'noAsyncDescribeCallback',
          message: 'No async describe callback',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 41,
        },
        {
          messageId: 'unexpectedDescribeArgument',
          message: 'Unexpected argument(s) in describe callback',
          line: 1,
          column: 33,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    ...(
      [
        ['describe.each()()', 1, 18],
        ['describe["each"]()()', 1, 21],
        ['describe.each(() => {})()', 1, 26],
        ['describe.each(() => {})("foo")', 25, 30],
        ['describe.each()(() => {})', 17, 25],
        ['describe["each"]()(() => {})', 20, 28],
        ['describe.each("foo")(() => {})', 22, 30],
        ['describe.only.each("foo")(() => {})', 27, 35],
        ['describe(() => {})', 10, 18],
        ['describe("foo")', 10, 15],
      ] as const
    ).map(([code, column, endColumn], index): TestCase => ({
      name: `upstream valid-describe-callback / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'nameAndCallback',
          message: 'Describe requires name and callback arguments',
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    {
      name: 'upstream valid-describe-callback / invalid 10',
      code: 'describe("foo", "foo2")',
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'secondArgumentMustBeFunction',
          message: 'Second argument must be function',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      name: 'upstream valid-describe-callback / invalid 11',
      code: 'describe()',
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'nameAndCallback',
          message: 'Describe requires name and callback arguments',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    ...(
      [
        ['describe("foo", async () => {})', 17, 31],
        ['describe("foo", async function () {})', 17, 37],
        ['xdescribe("foo", async function () {})', 18, 38],
        ['fdescribe("foo", async function () {})', 18, 38],
      ] as const
    ).map(([code, column, endColumn], index): TestCase => ({
      name: `upstream valid-describe-callback / invalid ${index + 12}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'noAsyncDescribeCallback',
          message: 'No async describe callback',
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    {
      name: 'upstream valid-describe-callback / invalid 16',
      code: `import { fdescribe } from '@jest/globals';

fdescribe("foo", async function () {})`,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'noAsyncDescribeCallback',
          message: 'No async describe callback',
          line: 3,
          column: 18,
          endLine: 3,
          endColumn: 38,
        },
      ],
    },
    ...(
      [
        ['describe.only("foo", async function () {})', 1, 22, 1, 42],
        ['describe.skip("foo", async function () {})', 1, 22, 1, 42],
        [
          `describe('sample case', () => {
  it('works', () => {
    expect(true).toEqual(true);
  });
  describe('async', async () => {
    await new Promise(setImmediate);
    it('breaks', () => {
      throw new Error('Fail');
    });
  });
});`,
          5,
          21,
          10,
          4,
        ],
      ] as const
    ).map(([code, line, column, endLine, endColumn], index): TestCase => ({
      name: `upstream valid-describe-callback / invalid ${index + 17}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'noAsyncDescribeCallback',
          message: 'No async describe callback',
          line,
          column,
          endLine,
          endColumn,
        },
      ],
    })),
    {
      name: 'upstream valid-describe-callback / invalid 20',
      code: `describe('foo', function () {
  return Promise.resolve().then(() => {
    it('breaks', () => {
      throw new Error('Fail')
    })
  })
})`,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'unexpectedReturnInDescribe',
          message: 'Unexpected return statement in describe callback',
          line: 2,
          column: 3,
          endLine: 6,
          endColumn: 5,
        },
      ],
    },
    {
      name: 'upstream valid-describe-callback / invalid 21',
      code: `describe('foo', () => {
  return Promise.resolve().then(() => {
    it('breaks', () => {
      throw new Error('Fail')
    })
  })
  describe('nested', () => {
    return Promise.resolve().then(() => {
      it('breaks', () => {
        throw new Error('Fail')
      })
    })
  })
})`,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'unexpectedReturnInDescribe',
          message: 'Unexpected return statement in describe callback',
          line: 2,
          column: 3,
          endLine: 6,
          endColumn: 5,
        },
        {
          messageId: 'unexpectedReturnInDescribe',
          message: 'Unexpected return statement in describe callback',
          line: 8,
          column: 5,
          endLine: 12,
          endColumn: 7,
        },
      ],
    },
    {
      name: 'upstream valid-describe-callback / invalid 22',
      code: `describe('foo', async () => {
  await something()
  it('does something')
  describe('nested', () => {
    return Promise.resolve().then(() => {
      it('breaks', () => {
        throw new Error('Fail')
      })
    })
  })
})`,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'noAsyncDescribeCallback',
          message: 'No async describe callback',
          line: 1,
          column: 17,
          endLine: 11,
          endColumn: 2,
        },
        {
          messageId: 'unexpectedReturnInDescribe',
          message: 'Unexpected return statement in describe callback',
          line: 5,
          column: 5,
          endLine: 9,
          endColumn: 7,
        },
      ],
    },
    {
      name: 'upstream valid-describe-callback / invalid 23',
      code: `describe('foo', () =>
  test('bar', () => {})
)`,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'unexpectedReturnInDescribe',
          message: 'Unexpected return statement in describe callback',
          line: 2,
          column: 3,
          endLine: 2,
          endColumn: 24,
        },
      ],
    },
    ...(
      [
        ['describe("foo", done => {})', 17, 21],
        ['describe("foo", function (done) {})', 27, 31],
        ['describe("foo", function (one, two, three) {})', 27, 30],
      ] as const
    ).map(([code, column, endColumn], index): TestCase => ({
      name: `upstream valid-describe-callback / invalid ${index + 24}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'unexpectedDescribeArgument',
          message: 'Unexpected argument(s) in describe callback',
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    {
      name: 'upstream valid-describe-callback / invalid 27',
      code: 'describe("foo", async function (done) {})',
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'noAsyncDescribeCallback',
          message: 'No async describe callback',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 41,
        },
        {
          messageId: 'unexpectedDescribeArgument',
          message: 'Unexpected argument(s) in describe callback',
          line: 1,
          column: 33,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
  ],
});
