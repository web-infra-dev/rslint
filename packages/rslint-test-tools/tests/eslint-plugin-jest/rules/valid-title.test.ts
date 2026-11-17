// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

// Shared source templates; reporting ranges remain explicit in each matrix.
const testSetup = `        describe('things to test', () => {
          describe('unit tests #unit', () => {
            it('is true', () => {
              expect(true).toBe(true);
            });
          });

          describe('e2e tests #e4e', () => {
            it('is another test #e2e #jest4life', () => {});
          });
        });
     \u0020`;
const sharedSetup = `
        test.each([
          [1, 1, 2],
          [1, 2, 3],
          [2, 1, 3],
`;
const testSetup2 = `describe('things to test', () => {
  describe('unit tests #unit', () => {
    it('is true', () => {
      expect(true).toBe(true);
    });
  });

  describe('e2e tests #e4e', () => {
    it('is another test #e2e #jest4life', () => {});
  });
`;

new RuleTester(fixtureConfig()).run('valid-title', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: 'describe("the correct way to properly handle all the things", () => {});',
    },
    {
      name: 'run 0 valid 1',
      code: 'test("that all is as it should be", () => {});',
    },
    {
      name: 'run 0 valid 2',
      code: 'it("correctly sets the value", () => {});',
      options: [
        {
          ignoreTypeOfDescribeName: false,
          disallowedWords: ['correct'],
        },
      ],
    },
    {
      name: 'run 0 valid 3',
      code: 'it("correctly sets the value", () => {});',
      options: [
        {
          disallowedWords: undefined,
        },
      ],
    },
    {
      name: 'run 0 valid 4',
      code: 'it("correctly sets the value", () => {});',
      options: [
        {
          mustMatch: {},
        },
      ],
    },
    {
      name: 'run 0 valid 5',
      code: 'it("correctly sets the value", () => {});',
      options: [
        {
          mustMatch: ' ',
        },
      ],
    },
    {
      name: 'run 0 valid 6',
      code: 'it("correctly sets the value", () => {});',
      options: [
        {
          mustMatch: [' '],
        },
      ],
    },
    {
      name: 'run 0 valid 7',
      code: 'it("correctly sets the value #unit", () => {});',
      options: [
        {
          mustMatch: '#(?:unit|integration|e2e)',
        },
      ],
    },
    {
      name: 'run 0 valid 8',
      code: 'it("correctly sets the value", () => {});',
      options: [
        {
          mustMatch: '^[^#]+$|(?:#(?:unit|e2e))',
        },
      ],
    },
    {
      name: 'run 0 valid 9',
      code: 'it("correctly sets the value", () => {});',
      options: [
        {
          mustMatch: {
            test: '#(?:unit|integration|e2e)',
          },
        },
      ],
    },
    {
      name: 'run 0 valid 10',
      code: `
        describe('things to test', () => {
          describe('unit tests #unit', () => {
            it('is true', () => {
              expect(true).toBe(true);
            });
          });

          describe('e2e tests #e2e', () => {
            it('is another test #jest4life', () => {});
          });
        });
     \u0020`,
      options: [
        {
          mustMatch: {
            test: '^[^#]+$|(?:#(?:unit|e2e))',
          },
        },
      ],
    },
    ...(
      [
        'it("is a string", () => {});',
        'it("is" + " a " + " string", () => {});',
        'it(1 + " + " + 1, () => {});',
        'test("is a string", () => {});',
        'xtest("is a string", () => {});',
        'xtest(`${myFunc} is a string`, () => {});',
        'describe("is a string", () => {});',
        'describe.skip("is a string", () => {});',
        'describe.skip(`${myFunc} is a string`, () => {});',
        'fdescribe("is a string", () => {});',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index + 11}`,
      code,
    })),
    {
      name: 'run 0 valid 21',
      code: 'describe(String(/.+/), () => {});',
      options: [
        {
          ignoreTypeOfDescribeName: true,
        },
      ],
    },
    ...(
      [
        'it(String(/.+/), () => {});',
        'const foo = "my-title"; it(foo, () => {});',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index + 22}`,
      code,
      options: [
        {
          ignoreTypeOfTestName: true,
        },
      ],
    })),
    {
      name: 'run 0 valid 24',
      code: 'describe(myFunction, () => {});',
      options: [
        {
          ignoreTypeOfDescribeName: true,
        },
      ],
    },
    {
      name: 'run 0 valid 25',
      code: 'xdescribe(skipFunction, () => {});',
      options: [
        {
          ignoreTypeOfDescribeName: true,
          disallowedWords: [],
        },
      ],
    },
    ...(
      [
        'describe()',
        'someFn("", function () {})',
        'describe("foo", function () {})',
        'describe("foo", function () { it("bar", function () {}) })',
        'test(`foo`, function () {})',
        'test.concurrent(`foo`, function () {})',
        'test(`\\nfoo\\n`, function () {})',
        'test(`${foo}`, function () {})',
        'test.concurrent(`${foo}`, function () {})',
        "it('foo', function () {})",
        'it.each([])()',
        "it.concurrent('foo', function () {})",
        "xdescribe('foo', function () {})",
        "xit('foo', function () {})",
        "xtest('foo', function () {})",
        'it()',
        'it.concurrent()',
        'it.each()()',
        'describe("foo", function () {})',
        'fdescribe("foo", function () {})',
        'xdescribe("foo", function () {})',
        'it("foo", function () {})',
        'it.concurrent("foo", function () {})',
        'fit("foo", function () {})',
        'fit.concurrent("foo", function () {})',
        'xit("foo", function () {})',
        'test("foo", function () {})',
        'test.concurrent("foo", function () {})',
        'xtest("foo", function () {})',
        'xtest(`foo`, function () {})',
        'someFn("foo", function () {})',
        `
      describe('foo', () => {
        it('bar', () => {})
      })
   \u0020`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index + 26}`,
      code,
    })),
    {
      name: 'run 0 valid 58',
      code: 'it(`GIVEN... \\n  `, () => {});',
      options: [
        {
          ignoreSpaces: true,
        },
      ],
    },
    ...(
      [
        'xdescribe(`foo`, function () {})',
        "test('foo', function () {})",
        'test("foo test", function () {})',
        'xtest("foo test", function () {})',
        'it("foos it correctly", function () {})',
        `
      describe('foo', () => {
        it('describes things correctly', () => {})
      })
   \u0020`,
        `
      it.each\`
        num   | value
        \${1} | \${true}
      \`('trues', ({ value }) => {
        expect(value).toBe(true);
      });
   \u0020`,
        `
      test.each([
        [1, 1, 2],
        [1, 2, 3],
        [2, 1, 3],
      ])('.add(%i, %i) = %i', (a, b, expected) => {
        expect(a + b).toBe(expected);
      });
   \u0020`,
        `
      test.only.each([
        [1, 1, 2],
        [1, 2, 3],
        [2, 1, 3],
      ])('.add(%i, %i)', (a, b, expected) => {
        expect(a + b).toBe(expected);
      });
   \u0020`,
        `
      describe.each([
        [1, 1, 2],
        [1, 2, 3],
        [2, 1, 3],
      ])('.add(%i, %i)', (a, b, expected) => {
        it('is correct', () => {
          expect(a + b).toBe(expected);
        });
      });
   \u0020`,
        `
      describe.skip.each([
        [1, 1, 2],
        [1, 2, 3],
        [2, 1, 3],
      ])('.add(%i, %i)', (a, b, expected) => {
        it('is correct', () => {
          expect(a + b).toBe(expected);
        });
      });
   \u0020`,
        `
      test.each([
        {a: 1, b: 1, expected: 2},
        {a: 1, b: 2, expected: 3},
        {a: 2, b: 1, expected: 3},
      ])('.add($a, $b)', ({a, b, expected}) => {
        expect(a + b).toBe(expected);
      });
   \u0020`,
        'it("returns 100%", () => {})',
        'it("returns %100%", () => {})',
        'it("returns %100", () => {})',
        'it("returns a percent%", () => {})',
        'it("returns a %percent%", () => {})',
        'it("returns a %percent", () => {})',
        'it.each([])("%i %p", () => {})',
        'it.skip.each([])("%%%", () => {})',
        'it.only.each([])("%%%%", () => {})',
        'it.failing.each([])("%%%%%", () => {})',
        'it.each([])("x%%y", () => {})',
        'it.each([])("x %% y", () => {})',
        'it.each([])("%int", () => {})',
        `
      const x = '%x';

      it.each([])(\`has a value of \${x}\`, () => {});
   \u0020`,
        `
      const x = '%x';

      it.each([])(\`has a value of %\${x}\`, () => {});
   \u0020`,
        `
      test.each\`
        a    | b    | expected
        \${1} | \${1} | \${2}
        \${1} | \${2} | \${3}
        \${2} | \${1} | \${3}
      \`('returns $expected when $a is added to $b', ({a, b, expected}) => {
        expect(a + b).toBe(expected);
      });
   \u0020`,
        `
      test.each\`
        a    | b    | expected
        \${1} | \${1} | \${2}
        \${1} | \${2} | \${3}
        \${2} | \${1} | \${3}
      \`('returns %expected when %a is added to %b', ({a, b, expected}) => {
        expect(a + b).toBe(expected);
      });
   \u0020`,
        ...['p', 's', 'd', 'i', 'f', 'j', 'o', '#', '$'].flatMap(
          (specifier) => [
            `it.each([])("%${specifier}", () => {})`,
            `it.each([])("%%${specifier}", () => {})`,
            `it.each([])("%${specifier}%", () => {})`,
            `it.each([])("%%${specifier}%", () => {})`,
            `it.each([])("%${specifier}!", () => {})`,
            `
        test.each([
          [1, 1, 2],
          [1, 2, 3],
          [2, 1, 3],
        ])(\`.add(%${specifier}, %${specifier})\`, (a, b, expected) => {
          expect(a + b).toBe(expected);
        });
\u0020\u0020\u0020\u0020\u0020\u0020`,
            `
        test.each([
          [1, 1, 2],
          [1, 2, 3],
          [2, 1, 3],
        ])(\`.add(%${specifier}, 1)\`, (a, b, expected) => {
          expect(a + b).toBe(expected);
        });
\u0020\u0020\u0020\u0020\u0020\u0020`,
          ],
        ),
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index + 59}`,
      code,
    })),
    ...(
      [
        'describe("the correct way to properly handle all the things", () => {});',
        'test("that all is as it should be", () => {});',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream disallowedWords option / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
    })),
    {
      name: 'upstream disallowedWords option / valid 2',
      code: 'it("correctly sets the value", () => {});',
      options: [
        {
          ignoreTypeOfDescribeName: false,
          disallowedWords: ['correct'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
    },
    {
      name: 'upstream disallowedWords option / valid 3',
      code: 'it("correctly sets the value", () => {});',
      options: [
        {
          disallowedWords: undefined,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
    },
    {
      name: 'upstream mustMatch & mustNotMatch options / valid 2',
      code: 'it("correctly sets the value", () => {});',
      options: [
        {
          mustMatch: {},
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
    },
    {
      name: 'upstream mustMatch & mustNotMatch options / valid 3',
      code: 'it("correctly sets the value", () => {});',
      options: [
        {
          mustMatch: ' ',
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
    },
    {
      name: 'upstream mustMatch & mustNotMatch options / valid 4',
      code: 'it("correctly sets the value", () => {});',
      options: [
        {
          mustMatch: [' '],
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
    },
    {
      name: 'upstream mustMatch & mustNotMatch options / valid 5',
      code: 'it("correctly sets the value #unit", () => {});',
      options: [
        {
          mustMatch: '#(?:unit|integration|e2e)',
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
    },
    {
      name: 'upstream mustMatch & mustNotMatch options / valid 6',
      code: 'it("correctly sets the value", () => {});',
      options: [
        {
          mustMatch: '^[^#]+$|(?:#(?:unit|e2e))',
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
    },
    {
      name: 'upstream mustMatch & mustNotMatch options / valid 7',
      code: 'it("correctly sets the value", () => {});',
      options: [
        {
          mustMatch: {
            test: '#(?:unit|integration|e2e)',
          },
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
    },
    {
      name: 'upstream mustMatch & mustNotMatch options / valid 8',
      code: `describe('things to test', () => {
  describe('unit tests #unit', () => {
    it('is true', () => {
      expect(true).toBe(true);
    });
  });

  describe('e2e tests #e2e', () => {
    it('is another test #jest4life', () => {});
  });
});`,
      options: [
        {
          mustMatch: {
            test: '^[^#]+$|(?:#(?:unit|e2e))',
          },
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
    },
    ...(
      [
        'it("is a string", () => {});',
        'it("is" + " a " + " string", () => {});',
        'it(1 + " + " + 1, () => {});',
        'test("is a string", () => {});',
        'xtest("is a string", () => {});',
        'xtest(`${myFunc} is a string`, () => {});',
        'describe("is a string", () => {});',
        'describe.skip("is a string", () => {});',
        'describe.skip(`${myFunc} is a string`, () => {});',
        'fdescribe("is a string", () => {});',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream title-must-be-string / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
    })),
    {
      name: 'upstream title-must-be-string / valid 10',
      code: 'describe(String(/.+/), () => {});',
      options: [
        {
          ignoreTypeOfDescribeName: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
    },
    ...(
      [
        'it(String(/.+/), () => {});',
        'const foo = "my-title"; it(foo, () => {});',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream title-must-be-string / valid ${index + 11}`,
      code,
      options: [
        {
          ignoreTypeOfTestName: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
    })),
    {
      name: 'upstream title-must-be-string / valid 13',
      code: 'describe(myFunction, () => {});',
      options: [
        {
          ignoreTypeOfDescribeName: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
    },
    {
      name: 'upstream title-must-be-string / valid 14',
      code: 'xdescribe(skipFunction, () => {});',
      options: [
        {
          ignoreTypeOfDescribeName: true,
          disallowedWords: [],
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
    },
    ...(
      [
        'describe()',
        'someFn("", function () {})',
        'describe("foo", function () {})',
        'describe("foo", function () { it("bar", function () {}) })',
        'test("foo", function () {})',
        'test.concurrent("foo", function () {})',
        'test(`foo`, function () {})',
        'test.concurrent(`foo`, function () {})',
        'test(`${foo}`, function () {})',
        'test.concurrent(`${foo}`, function () {})',
        "it('foo', function () {})",
        'it.each([])()',
        "it.concurrent('foo', function () {})",
        "xdescribe('foo', function () {})",
        "xit('foo', function () {})",
        "xtest('foo', function () {})",
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream no-empty-title / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
    })),
    ...(
      [
        ['it()', 'upstream no-accidental-space / valid 0'],
        ['it.concurrent()', 'upstream no-accidental-space / valid 1'],
        ['it.each()()', 'upstream no-accidental-space / valid 3'],
        [
          'fdescribe("foo", function () {})',
          'upstream no-accidental-space / valid 5',
        ],
        [
          'xdescribe("foo", function () {})',
          'upstream no-accidental-space / valid 6',
        ],
        ['it("foo", function () {})', 'upstream no-accidental-space / valid 7'],
        [
          'it.concurrent("foo", function () {})',
          'upstream no-accidental-space / valid 8',
        ],
        [
          'fit("foo", function () {})',
          'upstream no-accidental-space / valid 9',
        ],
        [
          'fit.concurrent("foo", function () {})',
          'upstream no-accidental-space / valid 10',
        ],
        [
          'xit("foo", function () {})',
          'upstream no-accidental-space / valid 11',
        ],
        [
          'xtest("foo", function () {})',
          'upstream no-accidental-space / valid 14',
        ],
        [
          'xtest(`foo`, function () {})',
          'upstream no-accidental-space / valid 15',
        ],
        [
          'someFn("foo", function () {})',
          'upstream no-accidental-space / valid 16',
        ],
        [
          `describe('foo', () => {
  it('bar', () => {})
})`,
          'upstream no-accidental-space / valid 17',
        ],
      ] as const
    ).map(([code, name]): ValidTestCase => ({
      name,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
    })),
    {
      name: 'upstream no-accidental-space / valid 18',
      code: `it(\`GIVEN...\u0020
  \`, () => {});`,
      options: [
        {
          ignoreSpaces: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
    },
    {
      name: 'upstream no-duplicate-prefix ft describe / valid 3',
      code: 'xdescribe(`foo`, function () {})',
      languageOptions: {
        ecmaVersion: 2017,
      },
    },
    ...(
      [
        [
          "test('foo', function () {})",
          'upstream no-duplicate-prefix ft test / valid 1',
        ],
        [
          'test("foo test", function () {})',
          'upstream no-duplicate-prefix ft test / valid 4',
        ],
        [
          'xtest("foo test", function () {})',
          'upstream no-duplicate-prefix ft test / valid 5',
        ],
      ] as const
    ).map(([code, name]): ValidTestCase => ({
      name,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
    })),
    ...(
      [
        'xit(`foo`, function () {})',
        'it("foos it correctly", function () {})',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream no-duplicate-prefix ft it / valid ${index + 3}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
    })),
    ...(
      [
        `describe('foo', () => {
  it('describes things correctly', () => {})
})`,
        `it.each\`
  num   | value
  \${1} | \${true}
\`('trues', ({ value }) => {
  expect(value).toBe(true);
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream no-duplicate-prefix ft nested / valid ${index + 1}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
    })),
    ...(
      [
        [
          `test.each([
  [1, 1, 2],
  [1, 2, 3],
  [2, 1, 3],
])('.add(%i, %i) = %i', (a, b, expected) => {
  expect(a + b).toBe(expected);
});`,
          'upstream valid-title (each printf checking) / valid 0',
        ],
        [
          `test.only.each([
  [1, 1, 2],
  [1, 2, 3],
  [2, 1, 3],
])('.add(%i, %i)', (a, b, expected) => {
  expect(a + b).toBe(expected);
});`,
          'upstream valid-title (each printf checking) / valid 1',
        ],
        [
          `describe.each([
  [1, 1, 2],
  [1, 2, 3],
  [2, 1, 3],
])('.add(%i, %i)', (a, b, expected) => {
  it('is correct', () => {
    expect(a + b).toBe(expected);
  });
});`,
          'upstream valid-title (each printf checking) / valid 2',
        ],
        [
          `describe.skip.each([
  [1, 1, 2],
  [1, 2, 3],
  [2, 1, 3],
])('.add(%i, %i)', (a, b, expected) => {
  it('is correct', () => {
    expect(a + b).toBe(expected);
  });
});`,
          'upstream valid-title (each printf checking) / valid 3',
        ],
        [
          'it.each([])("%p", () => {})',
          'upstream valid-title (each printf checking) / valid 4',
        ],
        [
          'it.each([])("%%p", () => {})',
          'upstream valid-title (each printf checking) / valid 5',
        ],
        [
          'it.each([])("%p%", () => {})',
          'upstream valid-title (each printf checking) / valid 6',
        ],
        [
          'it.each([])("%%p%", () => {})',
          'upstream valid-title (each printf checking) / valid 7',
        ],
        [
          'it.each([])("%p!", () => {})',
          'upstream valid-title (each printf checking) / valid 8',
        ],
        [
          `test.each([
  [1, 1, 2],
  [1, 2, 3],
  [2, 1, 3],
])('.add(%p, %p)', (a, b, expected) => {
  expect(a + b).toBe(expected);
});`,
          'upstream valid-title (each printf checking) / valid 9',
        ],
        [
          `test.each([
  [1, 1, 2],
  [1, 2, 3],
  [2, 1, 3],
])('.add(%p, 1)', (a, b, expected) => {
  expect(a + b).toBe(expected);
});`,
          'upstream valid-title (each printf checking) / valid 10',
        ],
        [
          'it.each([])("%s", () => {})',
          'upstream valid-title (each printf checking) / valid 11',
        ],
        [
          'it.each([])("%%s", () => {})',
          'upstream valid-title (each printf checking) / valid 12',
        ],
        [
          'it.each([])("%s%", () => {})',
          'upstream valid-title (each printf checking) / valid 13',
        ],
        [
          'it.each([])("%%s%", () => {})',
          'upstream valid-title (each printf checking) / valid 14',
        ],
        [
          'it.each([])("%s!", () => {})',
          'upstream valid-title (each printf checking) / valid 15',
        ],
        [
          `test.each([
  [1, 1, 2],
  [1, 2, 3],
  [2, 1, 3],
])('.add(%s, %s)', (a, b, expected) => {
  expect(a + b).toBe(expected);
});`,
          'upstream valid-title (each printf checking) / valid 16',
        ],
        [
          `test.each([
  [1, 1, 2],
  [1, 2, 3],
  [2, 1, 3],
])('.add(%s, 1)', (a, b, expected) => {
  expect(a + b).toBe(expected);
});`,
          'upstream valid-title (each printf checking) / valid 17',
        ],
        [
          'it.each([])("%d", () => {})',
          'upstream valid-title (each printf checking) / valid 18',
        ],
        [
          'it.each([])("%%d", () => {})',
          'upstream valid-title (each printf checking) / valid 19',
        ],
        [
          'it.each([])("%d%", () => {})',
          'upstream valid-title (each printf checking) / valid 20',
        ],
        [
          'it.each([])("%%d%", () => {})',
          'upstream valid-title (each printf checking) / valid 21',
        ],
        [
          'it.each([])("%d!", () => {})',
          'upstream valid-title (each printf checking) / valid 22',
        ],
        [
          `test.each([
  [1, 1, 2],
  [1, 2, 3],
  [2, 1, 3],
])('.add(%d, %d)', (a, b, expected) => {
  expect(a + b).toBe(expected);
});`,
          'upstream valid-title (each printf checking) / valid 23',
        ],
        [
          `test.each([
  [1, 1, 2],
  [1, 2, 3],
  [2, 1, 3],
])('.add(%d, 1)', (a, b, expected) => {
  expect(a + b).toBe(expected);
});`,
          'upstream valid-title (each printf checking) / valid 24',
        ],
        [
          'it.each([])("%i", () => {})',
          'upstream valid-title (each printf checking) / valid 25',
        ],
        [
          'it.each([])("%%i", () => {})',
          'upstream valid-title (each printf checking) / valid 26',
        ],
        [
          'it.each([])("%i%", () => {})',
          'upstream valid-title (each printf checking) / valid 27',
        ],
        [
          'it.each([])("%%i%", () => {})',
          'upstream valid-title (each printf checking) / valid 28',
        ],
        [
          'it.each([])("%i!", () => {})',
          'upstream valid-title (each printf checking) / valid 29',
        ],
        [
          `test.each([
  [1, 1, 2],
  [1, 2, 3],
  [2, 1, 3],
])('.add(%i, %i)', (a, b, expected) => {
  expect(a + b).toBe(expected);
});`,
          'upstream valid-title (each printf checking) / valid 30',
        ],
        [
          `test.each([
  [1, 1, 2],
  [1, 2, 3],
  [2, 1, 3],
])('.add(%i, 1)', (a, b, expected) => {
  expect(a + b).toBe(expected);
});`,
          'upstream valid-title (each printf checking) / valid 31',
        ],
        [
          'it.each([])("%f", () => {})',
          'upstream valid-title (each printf checking) / valid 32',
        ],
        [
          'it.each([])("%%f", () => {})',
          'upstream valid-title (each printf checking) / valid 33',
        ],
        [
          'it.each([])("%f%", () => {})',
          'upstream valid-title (each printf checking) / valid 34',
        ],
        [
          'it.each([])("%%f%", () => {})',
          'upstream valid-title (each printf checking) / valid 35',
        ],
        [
          'it.each([])("%f!", () => {})',
          'upstream valid-title (each printf checking) / valid 36',
        ],
        [
          `test.each([
  [1, 1, 2],
  [1, 2, 3],
  [2, 1, 3],
])('.add(%f, %f)', (a, b, expected) => {
  expect(a + b).toBe(expected);
});`,
          'upstream valid-title (each printf checking) / valid 37',
        ],
        [
          `test.each([
  [1, 1, 2],
  [1, 2, 3],
  [2, 1, 3],
])('.add(%f, 1)', (a, b, expected) => {
  expect(a + b).toBe(expected);
});`,
          'upstream valid-title (each printf checking) / valid 38',
        ],
        [
          'it.each([])("%j", () => {})',
          'upstream valid-title (each printf checking) / valid 39',
        ],
        [
          'it.each([])("%%j", () => {})',
          'upstream valid-title (each printf checking) / valid 40',
        ],
        [
          'it.each([])("%j%", () => {})',
          'upstream valid-title (each printf checking) / valid 41',
        ],
        [
          'it.each([])("%%j%", () => {})',
          'upstream valid-title (each printf checking) / valid 42',
        ],
        [
          'it.each([])("%j!", () => {})',
          'upstream valid-title (each printf checking) / valid 43',
        ],
        [
          `test.each([
  [1, 1, 2],
  [1, 2, 3],
  [2, 1, 3],
])('.add(%j, %j)', (a, b, expected) => {
  expect(a + b).toBe(expected);
});`,
          'upstream valid-title (each printf checking) / valid 44',
        ],
        [
          `test.each([
  [1, 1, 2],
  [1, 2, 3],
  [2, 1, 3],
])('.add(%j, 1)', (a, b, expected) => {
  expect(a + b).toBe(expected);
});`,
          'upstream valid-title (each printf checking) / valid 45',
        ],
        [
          'it.each([])("%o", () => {})',
          'upstream valid-title (each printf checking) / valid 46',
        ],
        [
          'it.each([])("%%o", () => {})',
          'upstream valid-title (each printf checking) / valid 47',
        ],
        [
          'it.each([])("%o%", () => {})',
          'upstream valid-title (each printf checking) / valid 48',
        ],
        [
          'it.each([])("%%o%", () => {})',
          'upstream valid-title (each printf checking) / valid 49',
        ],
        [
          'it.each([])("%o!", () => {})',
          'upstream valid-title (each printf checking) / valid 50',
        ],
        [
          `test.each([
  [1, 1, 2],
  [1, 2, 3],
  [2, 1, 3],
])('.add(%o, %o)', (a, b, expected) => {
  expect(a + b).toBe(expected);
});`,
          'upstream valid-title (each printf checking) / valid 51',
        ],
        [
          `test.each([
  [1, 1, 2],
  [1, 2, 3],
  [2, 1, 3],
])('.add(%o, 1)', (a, b, expected) => {
  expect(a + b).toBe(expected);
});`,
          'upstream valid-title (each printf checking) / valid 52',
        ],
        [
          'it.each([])("%#", () => {})',
          'upstream valid-title (each printf checking) / valid 53',
        ],
        [
          'it.each([])("%%#", () => {})',
          'upstream valid-title (each printf checking) / valid 54',
        ],
        [
          'it.each([])("%#%", () => {})',
          'upstream valid-title (each printf checking) / valid 55',
        ],
        [
          'it.each([])("%%#%", () => {})',
          'upstream valid-title (each printf checking) / valid 56',
        ],
        [
          'it.each([])("%#!", () => {})',
          'upstream valid-title (each printf checking) / valid 57',
        ],
        [
          `test.each([
  [1, 1, 2],
  [1, 2, 3],
  [2, 1, 3],
])('.add(%#, %#)', (a, b, expected) => {
  expect(a + b).toBe(expected);
});`,
          'upstream valid-title (each printf checking) / valid 58',
        ],
        [
          `test.each([
  [1, 1, 2],
  [1, 2, 3],
  [2, 1, 3],
])('.add(%#, 1)', (a, b, expected) => {
  expect(a + b).toBe(expected);
});`,
          'upstream valid-title (each printf checking) / valid 59',
        ],
        [
          'it.each([])("%$", () => {})',
          'upstream valid-title (each printf checking) / valid 60',
        ],
        [
          'it.each([])("%%$", () => {})',
          'upstream valid-title (each printf checking) / valid 61',
        ],
        [
          'it.each([])("%$%", () => {})',
          'upstream valid-title (each printf checking) / valid 62',
        ],
        [
          'it.each([])("%%$%", () => {})',
          'upstream valid-title (each printf checking) / valid 63',
        ],
        [
          'it.each([])("%$!", () => {})',
          'upstream valid-title (each printf checking) / valid 64',
        ],
        [
          `test.each([
  [1, 1, 2],
  [1, 2, 3],
  [2, 1, 3],
])('.add(%$, %$)', (a, b, expected) => {
  expect(a + b).toBe(expected);
});`,
          'upstream valid-title (each printf checking) / valid 65',
        ],
        [
          `test.each([
  [1, 1, 2],
  [1, 2, 3],
  [2, 1, 3],
])('.add(%$, 1)', (a, b, expected) => {
  expect(a + b).toBe(expected);
});`,
          'upstream valid-title (each printf checking) / valid 66',
        ],
        [
          `test.each([
  {a: 1, b: 1, expected: 2},
  {a: 1, b: 2, expected: 3},
  {a: 2, b: 1, expected: 3},
])('.add($a, $b)', ({a, b, expected}) => {
  expect(a + b).toBe(expected);
});`,
          'upstream valid-title (each printf checking) / valid 67',
        ],
        [
          'it("returns 100%", () => {})',
          'upstream valid-title (each printf checking) / valid 68',
        ],
        [
          'it("returns %100%", () => {})',
          'upstream valid-title (each printf checking) / valid 69',
        ],
        [
          'it("returns %100", () => {})',
          'upstream valid-title (each printf checking) / valid 70',
        ],
        [
          'it("returns a percent%", () => {})',
          'upstream valid-title (each printf checking) / valid 71',
        ],
        [
          'it("returns a %percent%", () => {})',
          'upstream valid-title (each printf checking) / valid 72',
        ],
        [
          'it("returns a %percent", () => {})',
          'upstream valid-title (each printf checking) / valid 73',
        ],
        [
          'it.each([])("%i %p", () => {})',
          'upstream valid-title (each printf checking) / valid 74',
        ],
        [
          'it.skip.each([])("%%%", () => {})',
          'upstream valid-title (each printf checking) / valid 75',
        ],
        [
          'it.only.each([])("%%%%", () => {})',
          'upstream valid-title (each printf checking) / valid 76',
        ],
        [
          'it.failing.each([])("%%%%%", () => {})',
          'upstream valid-title (each printf checking) / valid 77',
        ],
        [
          'it.each([])("x%%y", () => {})',
          'upstream valid-title (each printf checking) / valid 78',
        ],
        [
          'it.each([])("x %% y", () => {})',
          'upstream valid-title (each printf checking) / valid 79',
        ],
        [
          'it.each([])("%int", () => {})',
          'upstream valid-title (each printf checking) / valid 80',
        ],
        [
          `const x = '%x';

it.each([])(\`has a value of \${x}\`, () => {});`,
          'upstream valid-title (each printf checking) / valid 81',
        ],
        [
          `const x = '%x';

it.each([])(\`has a value of %\${x}\`, () => {});`,
          'upstream valid-title (each printf checking) / valid 82',
        ],
        [
          `test.each\`
  a    | b    | expected
  \${1} | \${1} | \${2}
  \${1} | \${2} | \${3}
  \${2} | \${1} | \${3}
\`('returns $expected when $a is added to $b', ({a, b, expected}) => {
  expect(a + b).toBe(expected);
});`,
          'upstream valid-title (each printf checking) / valid 84',
        ],
        [
          `test.each\`
  a    | b    | expected
  \${1} | \${1} | \${2}
  \${1} | \${2} | \${3}
  \${2} | \${1} | \${3}
\`('returns %expected when %a is added to %b', ({a, b, expected}) => {
  expect(a + b).toBe(expected);
});`,
          'upstream valid-title (each printf checking) / valid 85',
        ],
      ] as const
    ).map(([code, name]): ValidTestCase => ({
      name,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
    })),
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: 'test("the correct way to properly handle all things", () => {});',
      options: [
        {
          disallowedWords: ['correct', 'properly', 'all'],
        },
      ],
      errors: [
        {
          messageId: 'disallowedWord',
          message: '"correct" is not allowed in test titles',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
    {
      name: 'run 0 invalid 1',
      code: 'describe("the correct way to do things", function () {})',
      options: [
        {
          disallowedWords: ['correct'],
        },
      ],
      errors: [
        {
          messageId: 'disallowedWord',
          message: '"correct" is not allowed in test titles',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      name: 'run 0 invalid 2',
      code: 'it("has ALL the things", () => {})',
      options: [
        {
          disallowedWords: ['all'],
        },
      ],
      errors: [
        {
          messageId: 'disallowedWord',
          message: '"ALL" is not allowed in test titles',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      name: 'run 0 invalid 3',
      code: 'xdescribe("every single one of them", function () {})',
      options: [
        {
          disallowedWords: ['every'],
        },
      ],
      errors: [
        {
          messageId: 'disallowedWord',
          message: '"every" is not allowed in test titles',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      name: 'run 0 invalid 4',
      code: "describe('Very Descriptive Title Goes Here', function () {})",
      options: [
        {
          disallowedWords: ['descriptive'],
        },
      ],
      errors: [
        {
          messageId: 'disallowedWord',
          message: '"Descriptive" is not allowed in test titles',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 44,
        },
      ],
    },
    {
      name: 'run 0 invalid 5',
      code: 'test(`that the value is set properly`, function () {})',
      options: [
        {
          disallowedWords: ['properly'],
        },
      ],
      errors: [
        {
          messageId: 'disallowedWord',
          message: '"properly" is not allowed in test titles',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      name: 'run 0 invalid 6',
      code: 'it("it foo", () => {});',
      options: [
        {
          disallowedWords: ['ok', '(unterminated'],
          mustMatch: '^should ',
        },
      ],
      errors: [
        {
          messageId: 'invalidPattern',
          message:
            'Invalid regular expression in `disallowedWords` option: `\\b(ok|(unterminated)\\b`: error parsing regexp: missing closing ) in `\\b(ok|(unterminated)\\b`',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 1,
        },
      ],
    },
    ...(
      [
        [
          `
` + testSetup,
          9,
          20,
          10,
          16,
        ],
        [
          `
        import { describe, describe as context, it as thisTest } from '@jest/globals';

        describe('things to test', () => {
          context('unit tests #unit', () => {
            thisTest('is true', () => {
              expect(true).toBe(true);
            });
          });

          context('e2e tests #e4e', () => {
            thisTest('is another test #e2e #jest4life', () => {});
          });
        });
     \u0020`,
          11,
          19,
          12,
          22,
        ],
      ] as const
    ).map(([code, line1, column1, line2, column2], index): TestCase => ({
      name: `run 0 invalid ${index + 7}`,
      code,
      options: [
        {
          mustNotMatch: '(?:#(?!unit|e2e))\\w+',
          mustMatch: '^[^#]+$|(?:#(?:unit|e2e))',
        },
      ],
      errors: [
        {
          messageId: 'mustNotMatch',
          message: 'describe should not match /(?:#(?!unit|e2e))\\w+/u',
          line: line1,
          column: column1,
          endLine: line1,
          endColumn: column1 + 16,
        },
        {
          messageId: 'mustNotMatch',
          message: 'it should not match /(?:#(?!unit|e2e))\\w+/u',
          line: line2,
          column: column2,
          endLine: line2,
          endColumn: column2 + 33,
        },
      ],
    })),
    {
      name: 'run 0 invalid 9',
      code:
        `
` + testSetup,
      options: [
        {
          mustNotMatch: [
            '(?:#(?!unit|e2e))\\w+',
            'Please include "#unit" or "#e2e" in titles',
          ],
          mustMatch: [
            '^[^#]+$|(?:#(?:unit|e2e))',
            'Please include "#unit" or "#e2e" in titles',
          ],
        },
      ],
      errors: [
        {
          messageId: 'mustNotMatchCustom',
          message: 'Please include "#unit" or "#e2e" in titles',
          line: 9,
          column: 20,
          endLine: 9,
          endColumn: 36,
        },
        {
          messageId: 'mustNotMatchCustom',
          message: 'Please include "#unit" or "#e2e" in titles',
          line: 10,
          column: 16,
          endLine: 10,
          endColumn: 49,
        },
      ],
    },
    {
      name: 'run 0 invalid 10',
      code:
        `
` + testSetup,
      options: [
        {
          mustNotMatch: {
            describe: ['(?:#(?!unit|e2e))\\w+'],
          },
          mustMatch: {
            describe: '^[^#]+$|(?:#(?:unit|e2e))',
          },
        },
      ],
      errors: [
        {
          messageId: 'mustNotMatch',
          message: 'describe should not match /(?:#(?!unit|e2e))\\w+/u',
          line: 9,
          column: 20,
          endLine: 9,
          endColumn: 36,
        },
      ],
    },
    {
      name: 'run 0 invalid 11',
      code:
        `
` + testSetup,
      options: [
        {
          mustNotMatch: {
            describe: [
              '(?:#(?!unit|e2e))\\w+',
              'Please include "#unit" or "#e2e" in describe titles',
            ],
          },
          mustMatch: {
            describe: '^[^#]+$|(?:#(?:unit|e2e))',
          },
        },
      ],
      errors: [
        {
          messageId: 'mustNotMatchCustom',
          message: 'Please include "#unit" or "#e2e" in describe titles',
          line: 9,
          column: 20,
          endLine: 9,
          endColumn: 36,
        },
      ],
    },
    {
      name: 'run 0 invalid 12',
      code:
        `
` + testSetup,
      options: [
        {
          mustNotMatch: {
            describe: '(?:#(?!unit|e2e))\\w+',
          },
          mustMatch: {
            it: '^[^#]+$|(?:#(?:unit|e2e))',
          },
        },
      ],
      errors: [
        {
          messageId: 'mustNotMatch',
          message: 'describe should not match /(?:#(?!unit|e2e))\\w+/u',
          line: 9,
          column: 20,
          endLine: 9,
          endColumn: 36,
        },
      ],
    },
    {
      name: 'run 0 invalid 13',
      code: `
        describe('things to test', () => {
          describe('unit tests #unit', () => {
            it('is true #jest4life', () => {
              expect(true).toBe(true);
            });
          });

          describe('e2e tests #e4e', () => {
            it('is another test #e2e #jest4life', () => {});
          });
        });
     \u0020`,
      options: [
        {
          mustNotMatch: {
            describe: [
              '(?:#(?!unit|e2e))\\w+',
              'Please include "#unit" or "#e2e" in describe titles',
            ],
          },
          mustMatch: {
            it: [
              '^[^#]+$|(?:#(?:unit|e2e))',
              'Please include "#unit" or "#e2e" in it titles',
            ],
          },
        },
      ],
      errors: [
        {
          messageId: 'mustMatchCustom',
          message: 'Please include "#unit" or "#e2e" in it titles',
          line: 4,
          column: 16,
          endLine: 4,
          endColumn: 36,
        },
        {
          messageId: 'mustNotMatchCustom',
          message: 'Please include "#unit" or "#e2e" in describe titles',
          line: 9,
          column: 20,
          endLine: 9,
          endColumn: 36,
        },
      ],
    },
    {
      name: 'run 0 invalid 14',
      code: 'test("the correct way to properly handle all things", () => {});',
      options: [
        {
          mustMatch: '#(?:unit|integration|e2e)',
        },
      ],
      errors: [
        {
          messageId: 'mustMatch',
          message: 'test should match /#(?:unit|integration|e2e)/u',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
    ...(
      [
        ['describe("the test", () => {});', 10],
        ['xdescribe("the test", () => {});', 11],
        ['describe.skip("the test", () => {});', 15],
      ] as const
    ).map(([code, column], index): TestCase => ({
      name: `run 0 invalid ${index + 15}`,
      code,
      options: [
        {
          mustMatch: {
            describe: '#(?:unit|integration|e2e)',
          },
        },
      ],
      errors: [
        {
          messageId: 'mustMatch',
          message: 'describe should match /#(?:unit|integration|e2e)/u',
          line: 1,
          column,
          endLine: 1,
          endColumn: column + 10,
        },
      ],
    })),
    {
      name: 'run 0 invalid 18',
      code: 'it.each([])(1, () => {});',
      errors: [
        {
          messageId: 'titleMustBeString',
          message: 'Title must be a string',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    ...(
      [
        ['it(String(/.+/), () => {});', 4, 16],
        ['const foo = "my-title"; it(foo, () => {});', 28, 31],
      ] as const
    ).map(([code, column, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index + 19}`,
      code,
      options: [
        {
          ignoreTypeOfTestName: false,
        },
      ],
      errors: [
        {
          messageId: 'titleMustBeString',
          message: 'Title must be a string',
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    ...(
      [
        ['it.skip.each([])(1, () => {});', 18, 19],
        ['it.skip.each``(1, () => {});', 16, 17],
        ['it(123, () => {});', 4, 7],
        ['it.concurrent(123, () => {});', 15, 18],
        ['it(1 + 2 + 3, () => {});', 4, 13],
        ["it('a' || 'b', () => {});", 4, 14],
        ["it(cond && 'b', () => {});", 4, 15],
        ["it(x = 'foo', () => {});", 4, 13],
        ['it(foo + `suffix${x}`, () => {});', 4, 22],
        ['it(`prefix${x}` + foo, () => {});', 4, 22],
        ['it.concurrent(1 + 2 + 3, () => {});', 15, 24],
      ] as const
    ).map(([code, column, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index + 21}`,
      code,
      errors: [
        {
          messageId: 'titleMustBeString',
          message: 'Title must be a string',
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    {
      name: 'run 0 invalid 32',
      code: 'test.skip(123, () => {});',
      options: [
        {
          ignoreTypeOfDescribeName: true,
        },
      ],
      errors: [
        {
          messageId: 'titleMustBeString',
          message: 'Title must be a string',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      name: 'run 0 invalid 33',
      code: 'describe(String(/.+/), () => {});',
      errors: [
        {
          messageId: 'titleMustBeString',
          message: 'Title must be a string',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      name: 'run 0 invalid 34',
      code: 'describe(myFunction, () => 1);',
      options: [
        {
          ignoreTypeOfDescribeName: false,
        },
      ],
      errors: [
        {
          messageId: 'titleMustBeString',
          message: 'Title must be a string',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    ...(
      [
        ['describe(myFunction, () => {});', 10, 20],
        ['xdescribe(myFunction, () => {});', 11, 21],
        ['describe(6, function () {})', 10, 11],
        ['describe.skip(123, () => {});', 15, 18],
      ] as const
    ).map(([code, column, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index + 35}`,
      code,
      errors: [
        {
          messageId: 'titleMustBeString',
          message: 'Title must be a string',
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    ...(
      [
        [
          'describe("", function () {})',
          'describe should not have an empty title',
          1,
          1,
          29,
        ],
        [
          `
        describe('foo', () => {
          it('', () => {});
        });
     \u0020`,
          'test should not have an empty title',
          3,
          11,
          27,
        ],
        [
          'it("", function () {})',
          'test should not have an empty title',
          1,
          1,
          23,
        ],
        [
          'it.concurrent("", function () {})',
          'test should not have an empty title',
          1,
          1,
          34,
        ],
        [
          'test("", function () {})',
          'test should not have an empty title',
          1,
          1,
          25,
        ],
        [
          'test.concurrent("", function () {})',
          'test should not have an empty title',
          1,
          1,
          36,
        ],
        [
          'test(``, function () {})',
          'test should not have an empty title',
          1,
          1,
          25,
        ],
        [
          'test.concurrent(``, function () {})',
          'test should not have an empty title',
          1,
          1,
          36,
        ],
        [
          "xdescribe('', () => {})",
          'describe should not have an empty title',
          1,
          1,
          24,
        ],
        ["xit('', () => {})", 'test should not have an empty title', 1, 1, 18],
        [
          "xtest('', () => {})",
          'test should not have an empty title',
          1,
          1,
          20,
        ],
      ] as const
    ).map(([code, message, line, column, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index + 39}`,
      code,
      errors: [
        {
          messageId: 'emptyTitle',
          message,
          line,
          column,
          endLine: line,
          endColumn,
        },
      ],
    })),
    ...(
      [
        [
          'describe(" foo", function () {})',
          1,
          10,
          16,
          'describe("foo", function () {})',
        ],
        [
          'describe.each()(" foo", function () {})',
          1,
          17,
          23,
          'describe.each()("foo", function () {})',
        ],
        [
          'describe.only.each()(" foo", function () {})',
          1,
          22,
          28,
          'describe.only.each()("foo", function () {})',
        ],
        [
          'describe(" foo foe fum", function () {})',
          1,
          10,
          24,
          'describe("foo foe fum", function () {})',
        ],
        [
          'describe("foo foe fum ", function () {})',
          1,
          10,
          24,
          'describe("foo foe fum", function () {})',
        ],
        [
          'fdescribe(" foo", function () {})',
          1,
          11,
          17,
          'fdescribe("foo", function () {})',
        ],
        [
          "fdescribe(' foo', function () {})",
          1,
          11,
          17,
          "fdescribe('foo', function () {})",
        ],
        [
          'xdescribe(" foo", function () {})',
          1,
          11,
          17,
          'xdescribe("foo", function () {})',
        ],
        ['it(" foo", function () {})', 1, 4, 10, 'it("foo", function () {})'],
        [
          'it.concurrent(" foo", function () {})',
          1,
          15,
          21,
          'it.concurrent("foo", function () {})',
        ],
        ['fit(" foo", function () {})', 1, 5, 11, 'fit("foo", function () {})'],
        [
          'it.skip(" foo", function () {})',
          1,
          9,
          15,
          'it.skip("foo", function () {})',
        ],
        ['fit("foo ", function () {})', 1, 5, 11, 'fit("foo", function () {})'],
        [
          'it.skip("foo ", function () {})',
          1,
          9,
          15,
          'it.skip("foo", function () {})',
        ],
        [
          `
        import { test as testThat } from '@jest/globals';

        testThat('foo works ', () => {});
     \u0020`,
          4,
          18,
          30,
          `
        import { test as testThat } from '@jest/globals';

        testThat('foo works', () => {});
     \u0020`,
        ],
        ['xit(" foo", function () {})', 1, 5, 11, 'xit("foo", function () {})'],
        [
          'test(" foo", function () {})',
          1,
          6,
          12,
          'test("foo", function () {})',
        ],
        [
          'test.concurrent(" foo", function () {})',
          1,
          17,
          23,
          'test.concurrent("foo", function () {})',
        ],
        [
          'test(` foo`, function () {})',
          1,
          6,
          12,
          'test(`foo`, function () {})',
        ],
        [
          'test.concurrent(` foo`, function () {})',
          1,
          17,
          23,
          'test.concurrent(`foo`, function () {})',
        ],
        [
          'test(` foo bar bang`, function () {})',
          1,
          6,
          21,
          'test(`foo bar bang`, function () {})',
        ],
        [
          'test.concurrent(` foo bar bang`, function () {})',
          1,
          17,
          32,
          'test.concurrent(`foo bar bang`, function () {})',
        ],
        [
          'test(` foo bar bang  `, function () {})',
          1,
          6,
          23,
          'test(`foo bar bang`, function () {})',
        ],
        [
          'test.concurrent(` foo bar bang  `, function () {})',
          1,
          17,
          34,
          'test.concurrent(`foo bar bang`, function () {})',
        ],
        [
          'xtest(" foo", function () {})',
          1,
          7,
          13,
          'xtest("foo", function () {})',
        ],
        [
          'xtest(" foo  ", function () {})',
          1,
          7,
          15,
          'xtest("foo", function () {})',
        ],
        [
          `
        describe(' foo', () => {
          it('bar', () => {})
        })
     \u0020`,
          2,
          18,
          24,
          `
        describe('foo', () => {
          it('bar', () => {})
        })
     \u0020`,
        ],
        [
          `
        describe('foo', () => {
          it(' bar', () => {})
        })
     \u0020`,
          3,
          14,
          20,
          `
        describe('foo', () => {
          it('bar', () => {})
        })
     \u0020`,
        ],
      ] as const
    ).map(([code, line, column, endColumn, output], index): TestCase => ({
      name: `run 0 invalid ${index + 50}`,
      code,
      errors: [
        {
          messageId: 'accidentalSpace',
          message: 'should not have leading or trailing spaces',
          line,
          column,
          endLine: line,
          endColumn,
        },
      ],
      output,
    })),
    ...(
      [
        [
          'describe("describe foo", function () {})',
          1,
          10,
          24,
          'describe("foo", function () {})',
        ],
        [
          'fdescribe("describe foo", function () {})',
          1,
          11,
          25,
          'fdescribe("foo", function () {})',
        ],
        [
          'xdescribe("describe foo", function () {})',
          1,
          11,
          25,
          'xdescribe("foo", function () {})',
        ],
        [
          "describe('describe foo', function () {})",
          1,
          10,
          24,
          "describe('foo', function () {})",
        ],
        [
          'test("test foo", function () {})',
          1,
          6,
          16,
          'test("foo", function () {})',
        ],
        [
          'xtest("test foo", function () {})',
          1,
          7,
          17,
          'xtest("foo", function () {})',
        ],
        ['it("it foo", function () {})', 1, 4, 12, 'it("foo", function () {})'],
        [
          'fit("it foo", function () {})',
          1,
          5,
          13,
          'fit("foo", function () {})',
        ],
        [
          'xit("it foo", function () {})',
          1,
          5,
          13,
          'xit("foo", function () {})',
        ],
        [
          'it("it foos it correctly", function () {})',
          1,
          4,
          26,
          'it("foos it correctly", function () {})',
        ],
        [
          `
        describe('describe foo', () => {
          it('bar', () => {})
        })
     \u0020`,
          2,
          18,
          32,
          `
        describe('foo', () => {
          it('bar', () => {})
        })
     \u0020`,
        ],
        [
          `
        describe('describe foo', () => {
          it('describes things correctly', () => {})
        })
     \u0020`,
          2,
          18,
          32,
          `
        describe('foo', () => {
          it('describes things correctly', () => {})
        })
     \u0020`,
        ],
        [
          `
        describe('foo', () => {
          it('it bar', () => {})
        })
     \u0020`,
          3,
          14,
          22,
          `
        describe('foo', () => {
          it('bar', () => {})
        })
     \u0020`,
        ],
      ] as const
    ).map(([code, line, column, endColumn, output], index): TestCase => ({
      name: `run 0 invalid ${index + 78}`,
      code,
      errors: [
        {
          messageId: 'duplicatePrefix',
          message: 'should not have duplicate prefix',
          line,
          column,
          endLine: line,
          endColumn,
        },
      ],
      output,
    })),
    ...(
      [
        [
          'it.each([])("%y", () => {})',
          '"%y" is not a valid format specifier',
          1,
          13,
          17,
        ],
        [
          'it.each([])("%s+%y", () => {})',
          '"%y" is not a valid format specifier',
          1,
          13,
          20,
        ],
        [
          'it.each([])("%y+%%y", () => {})',
          '"%y" is not a valid format specifier',
          1,
          13,
          21,
        ],
        [
          'it.each([])("%%y+%y", () => {})',
          '"%y" is not a valid format specifier',
          1,
          13,
          21,
        ],
        [
          'it.each([])("%%%x", () => {})',
          '"%x" is not a valid format specifier',
          1,
          13,
          19,
        ],
        [
          sharedSetup +
            `        ])('.add(%x, %x)', (a, b, expected) => {
          expect(a + b).toBe(expected);
        });
     \u0020`,
          '"%x" is not a valid format specifier',
          6,
          12,
          26,
        ],
        [
          `
        test.only.each([
          [1, 1, 2],
          [1, 2, 3],
          [2, 1, 3],
        ])('.add(%x, %y)', (a, b, expected) => {
          expect(a + b).toBe(expected);
        });
     \u0020`,
          '"%x" is not a valid format specifier',
          6,
          12,
          26,
        ],
        [
          `
        describe.each([
          [1, 1, 2],
          [1, 2, 3],
          [2, 1, 3],
        ])('.add(%x, %y)', (a, b, expected) => {
          it('is correct', () => {
            expect(a + b).toBe(expected);
          });
        });
     \u0020`,
          '"%x" is not a valid format specifier',
          6,
          12,
          26,
        ],
        [
          sharedSetup +
            `        ])('.add(%i, %x)', (a, b, expected) => {
          expect(a + b).toBe(expected);
        });
     \u0020`,
          '"%x" is not a valid format specifier',
          6,
          12,
          26,
        ],
        [
          `
        it.skip.each(entries)('.add(%i, %x)', (a, b, expected) => {
          expect(a + b).toBe(expected);
        });
     \u0020`,
          '"%x" is not a valid format specifier',
          2,
          31,
          45,
        ],
        [
          `
        const entries = [
          [1, 1, 2],
          [1, 2, 3],
          [2, 1, 3],
        ];

        test.each(entries)('.add(%i, %x)', (a, b, expected) => {
          expect(a + b).toBe(expected);
        });
     \u0020`,
          '"%x" is not a valid format specifier',
          8,
          28,
          42,
        ],
        ...['P', 'S', 'D', 'I', 'F', 'J', 'O'].flatMap((specifier) => [
          [
            `it.each([])("%${specifier}", () => {})`,
            `"%${specifier}" is not a valid format specifier`,
            1,
            13,
            17,
          ] as const,
        ]),
      ] as const
    ).map(([code, message, line, column, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index + 91}`,
      code,
      errors: [
        {
          messageId: 'invalidEachSpecifier',
          message,
          line,
          column,
          endLine: line,
          endColumn,
        },
      ],
    })),
    {
      name: 'upstream disallowedWords option / invalid 0',
      code: 'test("the correct way to properly handle all things", () => {});',
      options: [
        {
          disallowedWords: ['correct', 'properly', 'all'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'disallowedWord',
          message: '"correct" is not allowed in test titles',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
    {
      name: 'upstream disallowedWords option / invalid 1',
      code: 'describe("the correct way to do things", function () {})',
      options: [
        {
          disallowedWords: ['correct'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'disallowedWord',
          message: '"correct" is not allowed in test titles',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      name: 'upstream disallowedWords option / invalid 2',
      code: 'it("has ALL the things", () => {})',
      options: [
        {
          disallowedWords: ['all'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'disallowedWord',
          message: '"ALL" is not allowed in test titles',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      name: 'upstream disallowedWords option / invalid 3',
      code: 'xdescribe("every single one of them", function () {})',
      options: [
        {
          disallowedWords: ['every'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'disallowedWord',
          message: '"every" is not allowed in test titles',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      name: 'upstream disallowedWords option / invalid 4',
      code: "describe('Very Descriptive Title Goes Here', function () {})",
      options: [
        {
          disallowedWords: ['descriptive'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'disallowedWord',
          message: '"Descriptive" is not allowed in test titles',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 44,
        },
      ],
    },
    {
      name: 'upstream disallowedWords option / invalid 5',
      code: 'test(`that the value is set properly`, function () {})',
      options: [
        {
          disallowedWords: ['properly'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'disallowedWord',
          message: '"properly" is not allowed in test titles',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      name: 'upstream mustMatch & mustNotMatch options / invalid 0',
      code: testSetup2 + '});',
      options: [
        {
          mustNotMatch: '(?:#(?!unit|e2e))\\w+',
          mustMatch: '^[^#]+$|(?:#(?:unit|e2e))',
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'mustNotMatch',
          message: 'describe should not match /(?:#(?!unit|e2e))\\w+/u',
          line: 8,
          column: 12,
          endLine: 8,
          endColumn: 28,
        },
        {
          messageId: 'mustNotMatch',
          message: 'it should not match /(?:#(?!unit|e2e))\\w+/u',
          line: 9,
          column: 8,
          endLine: 9,
          endColumn: 41,
        },
      ],
    },
    {
      name: 'upstream mustMatch & mustNotMatch options / invalid 1',
      code: `import { describe, describe as context, it as thisTest } from '@jest/globals';

describe('things to test', () => {
  context('unit tests #unit', () => {
    thisTest('is true', () => {
      expect(true).toBe(true);
    });
  });

  context('e2e tests #e4e', () => {
    thisTest('is another test #e2e #jest4life', () => {});
  });
});`,
      options: [
        {
          mustNotMatch: '(?:#(?!unit|e2e))\\w+',
          mustMatch: '^[^#]+$|(?:#(?:unit|e2e))',
        },
      ],
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'mustNotMatch',
          message: 'describe should not match /(?:#(?!unit|e2e))\\w+/u',
          line: 10,
          column: 11,
          endLine: 10,
          endColumn: 27,
        },
        {
          messageId: 'mustNotMatch',
          message: 'it should not match /(?:#(?!unit|e2e))\\w+/u',
          line: 11,
          column: 14,
          endLine: 11,
          endColumn: 47,
        },
      ],
    },
    {
      name: 'upstream mustMatch & mustNotMatch options / invalid 2',
      code: testSetup2 + '});',
      options: [
        {
          mustNotMatch: [
            '(?:#(?!unit|e2e))\\w+',
            'Please include "#unit" or "#e2e" in titles',
          ],
          mustMatch: [
            '^[^#]+$|(?:#(?:unit|e2e))',
            'Please include "#unit" or "#e2e" in titles',
          ],
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'mustNotMatchCustom',
          message: 'Please include "#unit" or "#e2e" in titles',
          line: 8,
          column: 12,
          endLine: 8,
          endColumn: 28,
        },
        {
          messageId: 'mustNotMatchCustom',
          message: 'Please include "#unit" or "#e2e" in titles',
          line: 9,
          column: 8,
          endLine: 9,
          endColumn: 41,
        },
      ],
    },
    {
      name: 'upstream mustMatch & mustNotMatch options / invalid 3',
      code: testSetup2 + '});',
      options: [
        {
          mustNotMatch: {
            describe: ['(?:#(?!unit|e2e))\\w+'],
          },
          mustMatch: {
            describe: '^[^#]+$|(?:#(?:unit|e2e))',
          },
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'mustNotMatch',
          message: 'describe should not match /(?:#(?!unit|e2e))\\w+/u',
          line: 8,
          column: 12,
          endLine: 8,
          endColumn: 28,
        },
      ],
    },
    {
      name: 'upstream mustMatch & mustNotMatch options / invalid 4',
      code: testSetup2 + '});',
      options: [
        {
          mustNotMatch: {
            describe: [
              '(?:#(?!unit|e2e))\\w+',
              'Please include "#unit" or "#e2e" in describe titles',
            ],
          },
          mustMatch: {
            describe: '^[^#]+$|(?:#(?:unit|e2e))',
          },
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'mustNotMatchCustom',
          message: 'Please include "#unit" or "#e2e" in describe titles',
          line: 8,
          column: 12,
          endLine: 8,
          endColumn: 28,
        },
      ],
    },
    {
      name: 'upstream mustMatch & mustNotMatch options / invalid 5',
      code: testSetup2 + '});',
      options: [
        {
          mustNotMatch: {
            describe: '(?:#(?!unit|e2e))\\w+',
          },
          mustMatch: {
            it: '^[^#]+$|(?:#(?:unit|e2e))',
          },
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'mustNotMatch',
          message: 'describe should not match /(?:#(?!unit|e2e))\\w+/u',
          line: 8,
          column: 12,
          endLine: 8,
          endColumn: 28,
        },
      ],
    },
    {
      name: 'upstream mustMatch & mustNotMatch options / invalid 6',
      code: `describe('things to test', () => {
  describe('unit tests #unit', () => {
    it('is true #jest4life', () => {
      expect(true).toBe(true);
    });
  });

  describe('e2e tests #e4e', () => {
    it('is another test #e2e #jest4life', () => {});
  });
});`,
      options: [
        {
          mustNotMatch: {
            describe: [
              '(?:#(?!unit|e2e))\\w+',
              'Please include "#unit" or "#e2e" in describe titles',
            ],
          },
          mustMatch: {
            it: [
              '^[^#]+$|(?:#(?:unit|e2e))',
              'Please include "#unit" or "#e2e" in it titles',
            ],
          },
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'mustMatchCustom',
          message: 'Please include "#unit" or "#e2e" in it titles',
          line: 3,
          column: 8,
          endLine: 3,
          endColumn: 28,
        },
        {
          messageId: 'mustNotMatchCustom',
          message: 'Please include "#unit" or "#e2e" in describe titles',
          line: 8,
          column: 12,
          endLine: 8,
          endColumn: 28,
        },
      ],
    },
    {
      name: 'upstream mustMatch & mustNotMatch options / invalid 7',
      code: 'test("the correct way to properly handle all things", () => {});',
      options: [
        {
          mustMatch: '#(?:unit|integration|e2e)',
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'mustMatch',
          message: 'test should match /#(?:unit|integration|e2e)/u',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
    ...(
      [
        ['describe("the test", () => {});', 10],
        ['xdescribe("the test", () => {});', 11],
        ['describe.skip("the test", () => {});', 15],
      ] as const
    ).map(([code, column], index): TestCase => ({
      name: `upstream mustMatch & mustNotMatch options / invalid ${index + 8}`,
      code,
      options: [
        {
          mustMatch: {
            describe: '#(?:unit|integration|e2e)',
          },
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'mustMatch',
          message: 'describe should match /#(?:unit|integration|e2e)/u',
          line: 1,
          column,
          endLine: 1,
          endColumn: column + 10,
        },
      ],
    })),
    {
      name: 'upstream title-must-be-string / invalid 0',
      code: 'it.each([])(1, () => {});',
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'titleMustBeString',
          message: 'Title must be a string',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    ...(
      [
        ['it(String(/.+/), () => {});', 4, 16],
        ['const foo = "my-title"; it(foo, () => {});', 28, 31],
      ] as const
    ).map(([code, column, endColumn], index): TestCase => ({
      name: `upstream title-must-be-string / invalid ${index + 1}`,
      code,
      options: [
        {
          ignoreTypeOfTestName: false,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'titleMustBeString',
          message: 'Title must be a string',
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    ...(
      [
        ['it.skip.each([])(1, () => {});', 18, 19],
        ['it.skip.each``(1, () => {});', 16, 17],
        ['it(123, () => {});', 4, 7],
        ['it.concurrent(123, () => {});', 15, 18],
        ['it(1 + 2 + 3, () => {});', 4, 13],
        ['it.concurrent(1 + 2 + 3, () => {});', 15, 24],
      ] as const
    ).map(([code, column, endColumn], index): TestCase => ({
      name: `upstream title-must-be-string / invalid ${index + 3}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'titleMustBeString',
          message: 'Title must be a string',
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    {
      name: 'upstream title-must-be-string / invalid 9',
      code: 'test.skip(123, () => {});',
      options: [
        {
          ignoreTypeOfDescribeName: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'titleMustBeString',
          message: 'Title must be a string',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      name: 'upstream title-must-be-string / invalid 10',
      code: 'describe(String(/.+/), () => {});',
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'titleMustBeString',
          message: 'Title must be a string',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      name: 'upstream title-must-be-string / invalid 11',
      code: 'describe(myFunction, () => 1);',
      options: [
        {
          ignoreTypeOfDescribeName: false,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'titleMustBeString',
          message: 'Title must be a string',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    ...(
      [
        ['describe(myFunction, () => {});', 10, 20],
        ['xdescribe(myFunction, () => {});', 11, 21],
        ['describe(6, function () {})', 10, 11],
        ['describe.skip(123, () => {});', 15, 18],
      ] as const
    ).map(([code, column, endColumn], index): TestCase => ({
      name: `upstream title-must-be-string / invalid ${index + 12}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'titleMustBeString',
          message: 'Title must be a string',
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    ...(
      [
        [
          'describe("", function () {})',
          'describe should not have an empty title',
          1,
          1,
          29,
        ],
        [
          `describe('foo', () => {
  it('', () => {});
});`,
          'test should not have an empty title',
          2,
          3,
          19,
        ],
        [
          'it("", function () {})',
          'test should not have an empty title',
          1,
          1,
          23,
        ],
        [
          'it.concurrent("", function () {})',
          'test should not have an empty title',
          1,
          1,
          34,
        ],
        [
          'test("", function () {})',
          'test should not have an empty title',
          1,
          1,
          25,
        ],
        [
          'test.concurrent("", function () {})',
          'test should not have an empty title',
          1,
          1,
          36,
        ],
        [
          'test(``, function () {})',
          'test should not have an empty title',
          1,
          1,
          25,
        ],
        [
          'test.concurrent(``, function () {})',
          'test should not have an empty title',
          1,
          1,
          36,
        ],
        [
          "xdescribe('', () => {})",
          'describe should not have an empty title',
          1,
          1,
          24,
        ],
        ["xit('', () => {})", 'test should not have an empty title', 1, 1, 18],
        [
          "xtest('', () => {})",
          'test should not have an empty title',
          1,
          1,
          20,
        ],
      ] as const
    ).map(([code, message, line, column, endColumn], index): TestCase => ({
      name: `upstream no-empty-title / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'emptyTitle',
          message,
          line,
          column,
          endLine: line,
          endColumn,
        },
      ],
    })),
    ...(
      [
        [
          'describe(" foo", function () {})',
          10,
          16,
          'describe("foo", function () {})',
        ],
        [
          'describe.each()(" foo", function () {})',
          17,
          23,
          'describe.each()("foo", function () {})',
        ],
        [
          'describe.only.each()(" foo", function () {})',
          22,
          28,
          'describe.only.each()("foo", function () {})',
        ],
        [
          'describe(" foo foe fum", function () {})',
          10,
          24,
          'describe("foo foe fum", function () {})',
        ],
        [
          'describe("foo foe fum ", function () {})',
          10,
          24,
          'describe("foo foe fum", function () {})',
        ],
        [
          'fdescribe(" foo", function () {})',
          11,
          17,
          'fdescribe("foo", function () {})',
        ],
        [
          "fdescribe(' foo', function () {})",
          11,
          17,
          "fdescribe('foo', function () {})",
        ],
        [
          'xdescribe(" foo", function () {})',
          11,
          17,
          'xdescribe("foo", function () {})',
        ],
        ['it(" foo", function () {})', 4, 10, 'it("foo", function () {})'],
        [
          'it.concurrent(" foo", function () {})',
          15,
          21,
          'it.concurrent("foo", function () {})',
        ],
        ['fit(" foo", function () {})', 5, 11, 'fit("foo", function () {})'],
        [
          'it.skip(" foo", function () {})',
          9,
          15,
          'it.skip("foo", function () {})',
        ],
        ['fit("foo ", function () {})', 5, 11, 'fit("foo", function () {})'],
        [
          'it.skip("foo ", function () {})',
          9,
          15,
          'it.skip("foo", function () {})',
        ],
      ] as const
    ).map(([code, column, endColumn, output], index): TestCase => ({
      name: `upstream no-accidental-space / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'accidentalSpace',
          message: 'should not have leading or trailing spaces',
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    {
      name: 'upstream no-accidental-space / invalid 14',
      code: `import { test as testThat } from '@jest/globals';

testThat('foo works ', () => {});`,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'accidentalSpace',
          message: 'should not have leading or trailing spaces',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 22,
        },
      ],
      output: `import { test as testThat } from '@jest/globals';

testThat('foo works', () => {});`,
    },
    ...(
      [
        ['xit(" foo", function () {})', 1, 5, 11, 'xit("foo", function () {})'],
        [
          'test(" foo", function () {})',
          1,
          6,
          12,
          'test("foo", function () {})',
        ],
        [
          'test.concurrent(" foo", function () {})',
          1,
          17,
          23,
          'test.concurrent("foo", function () {})',
        ],
        [
          'test(` foo`, function () {})',
          1,
          6,
          12,
          'test(`foo`, function () {})',
        ],
        [
          'test.concurrent(` foo`, function () {})',
          1,
          17,
          23,
          'test.concurrent(`foo`, function () {})',
        ],
        [
          'test(` foo bar bang`, function () {})',
          1,
          6,
          21,
          'test(`foo bar bang`, function () {})',
        ],
        [
          'test.concurrent(` foo bar bang`, function () {})',
          1,
          17,
          32,
          'test.concurrent(`foo bar bang`, function () {})',
        ],
        [
          'test(` foo bar bang  `, function () {})',
          1,
          6,
          23,
          'test(`foo bar bang`, function () {})',
        ],
        [
          'test.concurrent(` foo bar bang  `, function () {})',
          1,
          17,
          34,
          'test.concurrent(`foo bar bang`, function () {})',
        ],
        [
          'xtest(" foo", function () {})',
          1,
          7,
          13,
          'xtest("foo", function () {})',
        ],
        [
          'xtest(" foo  ", function () {})',
          1,
          7,
          15,
          'xtest("foo", function () {})',
        ],
        [
          `describe(' foo', () => {
  it('bar', () => {})
})`,
          1,
          10,
          16,
          `describe('foo', () => {
  it('bar', () => {})
})`,
        ],
        [
          `describe('foo', () => {
  it(' bar', () => {})
})`,
          2,
          6,
          12,
          `describe('foo', () => {
  it('bar', () => {})
})`,
        ],
      ] as const
    ).map(([code, line, column, endColumn, output], index): TestCase => ({
      name: `upstream no-accidental-space / invalid ${index + 15}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'accidentalSpace',
          message: 'should not have leading or trailing spaces',
          line,
          column,
          endLine: line,
          endColumn,
        },
      ],
      output,
    })),
    ...(
      [
        [
          'describe("describe foo", function () {})',
          10,
          'describe("foo", function () {})',
        ],
        [
          'fdescribe("describe foo", function () {})',
          11,
          'fdescribe("foo", function () {})',
        ],
        [
          'xdescribe("describe foo", function () {})',
          11,
          'xdescribe("foo", function () {})',
        ],
        [
          "describe('describe foo', function () {})",
          10,
          "describe('foo', function () {})",
        ],
        [
          'fdescribe(`describe foo`, function () {})',
          11,
          'fdescribe(`foo`, function () {})',
        ],
      ] as const
    ).map(([code, column, output], index): TestCase => ({
      name: `upstream no-duplicate-prefix ft describe / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'duplicatePrefix',
          message: 'should not have duplicate prefix',
          line: 1,
          column,
          endLine: 1,
          endColumn: column + 14,
        },
      ],
      output,
    })),
    ...(
      [
        [
          'test("test foo", function () {})',
          6,
          16,
          'test("foo", function () {})',
        ],
        [
          'xtest("test foo", function () {})',
          7,
          17,
          'xtest("foo", function () {})',
        ],
        [
          'test(`test foo`, function () {})',
          6,
          16,
          'test(`foo`, function () {})',
        ],
        [
          'test(`test foo test`, function () {})',
          6,
          21,
          'test(`foo test`, function () {})',
        ],
      ] as const
    ).map(([code, column, endColumn, output], index): TestCase => ({
      name: `upstream no-duplicate-prefix ft test / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'duplicatePrefix',
          message: 'should not have duplicate prefix',
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
        ['it("it foo", function () {})', 4, 12, 'it("foo", function () {})'],
        ['fit("it foo", function () {})', 5, 13, 'fit("foo", function () {})'],
        ['xit("it foo", function () {})', 5, 13, 'xit("foo", function () {})'],
        [
          'it("it foos it correctly", function () {})',
          4,
          26,
          'it("foos it correctly", function () {})',
        ],
      ] as const
    ).map(([code, column, endColumn, output], index): TestCase => ({
      name: `upstream no-duplicate-prefix ft it / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'duplicatePrefix',
          message: 'should not have duplicate prefix',
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
          `describe('describe foo', () => {
  it('bar', () => {})
})`,
          1,
          10,
          24,
          `describe('foo', () => {
  it('bar', () => {})
})`,
        ],
        [
          `describe('describe foo', () => {
  it('describes things correctly', () => {})
})`,
          1,
          10,
          24,
          `describe('foo', () => {
  it('describes things correctly', () => {})
})`,
        ],
        [
          `describe('foo', () => {
  it('it bar', () => {})
})`,
          2,
          6,
          14,
          `describe('foo', () => {
  it('bar', () => {})
})`,
        ],
      ] as const
    ).map(([code, line, column, endColumn, output], index): TestCase => ({
      name: `upstream no-duplicate-prefix ft nested / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'duplicatePrefix',
          message: 'should not have duplicate prefix',
          line,
          column,
          endLine: line,
          endColumn,
        },
      ],
      output,
    })),
    ...(
      [
        [
          'it.each([])("%y", () => {})',
          '"%y" is not a valid format specifier',
          1,
          13,
          17,
        ],
        ...['P', 'S', 'D', 'I', 'F', 'J', 'O'].flatMap((specifier) => [
          [
            `it.each([])("%${specifier}", () => {})`,
            `"%${specifier}" is not a valid format specifier`,
            1,
            13,
            17,
          ] as const,
        ]),
        [
          'it.each([])("%s+%y", () => {})',
          '"%y" is not a valid format specifier',
          1,
          13,
          20,
        ],
        [
          'it.each([])("%y+%%y", () => {})',
          '"%y" is not a valid format specifier',
          1,
          13,
          21,
        ],
        [
          'it.each([])("%%y+%y", () => {})',
          '"%y" is not a valid format specifier',
          1,
          13,
          21,
        ],
        [
          'it.each([])("%%%x", () => {})',
          '"%x" is not a valid format specifier',
          1,
          13,
          19,
        ],
        [
          `test.each([
  [1, 1, 2],
  [1, 2, 3],
  [2, 1, 3],
])('.add(%x, %x)', (a, b, expected) => {
  expect(a + b).toBe(expected);
});`,
          '"%x" is not a valid format specifier',
          5,
          4,
          18,
        ],
        [
          `test.only.each([
  [1, 1, 2],
  [1, 2, 3],
  [2, 1, 3],
])('.add(%x, %y)', (a, b, expected) => {
  expect(a + b).toBe(expected);
});`,
          '"%x" is not a valid format specifier',
          5,
          4,
          18,
        ],
        [
          `describe.each([
  [1, 1, 2],
  [1, 2, 3],
  [2, 1, 3],
])('.add(%x, %y)', (a, b, expected) => {
  it('is correct', () => {
    expect(a + b).toBe(expected);
  });
});`,
          '"%x" is not a valid format specifier',
          5,
          4,
          18,
        ],
        [
          `test.each([
  [1, 1, 2],
  [1, 2, 3],
  [2, 1, 3],
])('.add(%i, %x)', (a, b, expected) => {
  expect(a + b).toBe(expected);
});`,
          '"%x" is not a valid format specifier',
          5,
          4,
          18,
        ],
        [
          `it.skip.each(entries)('.add(%i, %x)', (a, b, expected) => {
  expect(a + b).toBe(expected);
});`,
          '"%x" is not a valid format specifier',
          1,
          23,
          37,
        ],
        [
          `const entries = [
  [1, 1, 2],
  [1, 2, 3],
  [2, 1, 3],
];\u0020

test.each(entries)('.add(%i, %x)', (a, b, expected) => {
  expect(a + b).toBe(expected);
});`,
          '"%x" is not a valid format specifier',
          7,
          20,
          34,
        ],
      ] as const
    ).map(([code, message, line, column, endColumn], index): TestCase => ({
      name: `upstream valid-title (each printf checking) / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'invalidEachSpecifier',
          message,
          line,
          column,
          endLine: line,
          endColumn,
        },
      ],
    })),
  ],
});
