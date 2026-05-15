import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('valid-title', {} as never, {
  valid: [
    // Baselines (long titles; listed once)
    {
      code: `describe("the correct way to properly handle all the things", () => {});`,
    },
    { code: `test("that all is as it should be", () => {});` },

    // disallowedWords
    {
      code: 'it("correctly sets the value", () => {});',
      options: [
        { ignoreTypeOfDescribeName: false, disallowedWords: ['correct'] },
      ],
    },
    {
      code: 'it("correctly sets the value", () => {});',
      options: [{ disallowedWords: undefined }],
    },

    // mustMatch (valid)
    {
      code: 'it("correctly sets the value", () => {});',
      options: [{ mustMatch: {} }],
    },
    {
      code: 'it("correctly sets the value", () => {});',
      options: [{ mustMatch: / /u.source }],
    },
    {
      code: 'it("correctly sets the value", () => {});',
      options: [{ mustMatch: [/ /u.source] }],
    },
    {
      code: 'it("correctly sets the value #unit", () => {});',
      options: [{ mustMatch: `#(?:unit|integration|e2e)` }],
    },
    {
      code: 'it("correctly sets the value", () => {});',
      options: [{ mustMatch: `^[^#]+$|(?:#(?:unit|e2e))` }],
    },
    {
      code: 'it("correctly sets the value", () => {});',
      options: [{ mustMatch: { test: `#(?:unit|integration|e2e)` } }],
    },
    {
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
      `,
      options: [{ mustMatch: { test: `^[^#]+$|(?:#(?:unit|e2e))` } }],
    },

    // Title must be a string
    { code: 'it("is a string", () => {});' },
    { code: 'it("is" + " a " + " string", () => {});' },
    { code: 'it(1 + " + " + 1, () => {});' },
    { code: 'test("is a string", () => {});' },
    { code: 'xtest("is a string", () => {});' },
    { code: 'xtest(`${myFunc} is a string`, () => {});' },
    { code: 'describe("is a string", () => {});' },
    { code: 'describe.skip("is a string", () => {});' },
    { code: 'describe.skip(`${myFunc} is a string`, () => {});' },
    { code: 'fdescribe("is a string", () => {});' },
    {
      code: 'describe(String(/.+/), () => {});',
      options: [{ ignoreTypeOfDescribeName: true }],
    },
    {
      code: 'it(String(/.+/), () => {});',
      options: [{ ignoreTypeOfTestName: true }],
    },
    {
      code: 'const foo = "my-title"; it(foo, () => {});',
      options: [{ ignoreTypeOfTestName: true }],
    },
    {
      code: 'describe(myFunction, () => {});',
      options: [{ ignoreTypeOfDescribeName: true }],
    },
    {
      code: 'xdescribe(skipFunction, () => {});',
      options: [{ ignoreTypeOfDescribeName: true, disallowedWords: [] }],
    },

    // No empty title
    { code: 'describe()' },
    { code: 'someFn("", function () {})' },
    { code: 'describe("foo", function () {})' },
    {
      code: 'describe("foo", function () { it("bar", function () {}) })',
    },
    { code: 'test(`foo`, function () {})' },
    { code: 'test.concurrent(`foo`, function () {})' },
    { code: 'test(`${foo}`, function () {})' },
    { code: 'test.concurrent(`${foo}`, function () {})' },
    { code: "it('foo', function () {})" },
    { code: 'it.each([])()' },
    { code: "it.concurrent('foo', function () {})" },
    { code: "xdescribe('foo', function () {})" },
    { code: "xit('foo', function () {})" },
    { code: "xtest('foo', function () {})" },

    // No accidental space (incl. fdescribe/xdescribe duplicates folded here)
    { code: `it()` },
    { code: `it.concurrent()` },
    { code: `it.each()()` },
    { code: `describe("foo", function () {})` },
    { code: `fdescribe("foo", function () {})` },
    { code: `xdescribe("foo", function () {})` },
    { code: `it("foo", function () {})` },
    { code: `it.concurrent("foo", function () {})` },
    { code: `fit("foo", function () {})` },
    { code: `fit.concurrent("foo", function () {})` },
    { code: `xit("foo", function () {})` },
    { code: `test("foo", function () {})` },
    { code: `test.concurrent("foo", function () {})` },
    { code: `xtest("foo", function () {})` },
    { code: `xtest(\`foo\`, function () {})` },
    { code: `someFn("foo", function () {})` },
    {
      code: `
      describe('foo', () => {
        it('bar', () => {})
      })
    `,
    },
    {
      code: 'it(`GIVEN... \\n  `, () => {});',
      options: [{ ignoreSpaces: true }],
    },

    // Duplicate-prefix valid (extras not listed above)
    { code: 'xdescribe(`foo`, function () {})' },
    { code: "test('foo', function () {})" },
    { code: `test("foo test", function () {})` },
    { code: `xtest("foo test", function () {})` },
    { code: `it("foos it correctly", function () {})` },
    {
      code: `
      describe('foo', () => {
        it('describes things correctly', () => {})
      })
    `,
    },
    {
      code: `
      it.each\`
        num   | value
        \${1} | \${true}
      \`('trues', ({ value }) => {
        expect(value).toBe(true);
      });
    `,
    },

    // each / printf-valid titles (static blocks)
    {
      code: `
      test.each([
        [1, 1, 2],
        [1, 2, 3],
        [2, 1, 3],
      ])('.add(%i, %i) = %i', (a, b, expected) => {
        expect(a + b).toBe(expected);
      });
    `,
    },
    {
      code: `
      test.only.each([
        [1, 1, 2],
        [1, 2, 3],
        [2, 1, 3],
      ])('.add(%i, %i)', (a, b, expected) => {
        expect(a + b).toBe(expected);
      });
    `,
    },
    {
      code: `
      describe.each([
        [1, 1, 2],
        [1, 2, 3],
        [2, 1, 3],
      ])('.add(%i, %i)', (a, b, expected) => {
        it('is correct', () => {
          expect(a + b).toBe(expected);
        });
      });
    `,
    },
    {
      code: `
      describe.skip.each([
        [1, 1, 2],
        [1, 2, 3],
        [2, 1, 3],
      ])('.add(%i, %i)', (a, b, expected) => {
        it('is correct', () => {
          expect(a + b).toBe(expected);
        });
      });
    `,
    },
    {
      code: `
      test.each([
        {a: 1, b: 1, expected: 2},
        {a: 1, b: 2, expected: 3},
        {a: 2, b: 1, expected: 3},
      ])('.add($a, $b)', ({a, b, expected}) => {
        expect(a + b).toBe(expected);
      });
    `,
    },
    { code: `it("returns 100%", () => {})` },
    { code: `it("returns %100%", () => {})` },
    { code: `it("returns %100", () => {})` },
    { code: `it("returns a percent%", () => {})` },
    { code: `it("returns a %percent%", () => {})` },
    { code: `it("returns a %percent", () => {})` },
    { code: `it.each([])("%i %p", () => {})` },
    { code: `it.skip.each([])("%%%", () => {})` },
    { code: `it.only.each([])("%%%%", () => {})` },
    { code: `it.failing.each([])("%%%%%", () => {})` },
    { code: `it.each([])("x%%y", () => {})` },
    { code: `it.each([])("x %% y", () => {})` },
    { code: `it.each([])("%int", () => {})` },
    {
      code: `
      const x = '%x';

      it.each([])(\`has a value of \${x}\`, () => {});
    `,
    },
    {
      code: `
      const x = '%x';

      it.each([])(\`has a value of %\${x}\`, () => {});
    `,
    },
    {
      code: `
      test.each\`
        a    | b    | expected
        \${1} | \${1} | \${2}
        \${1} | \${2} | \${3}
        \${2} | \${1} | \${3}
      \`('returns $expected when $a is added to $b', ({a, b, expected}) => {
        expect(a + b).toBe(expected);
      });
    `,
    },
    {
      code: `
      test.each\`
        a    | b    | expected
        \${1} | \${1} | \${2}
        \${1} | \${2} | \${3}
        \${2} | \${1} | \${3}
      \`('returns %expected when %a is added to %b', ({a, b, expected}) => {
        expect(a + b).toBe(expected);
      });
    `,
    },

    // each printf specifiers (flattened)
    ...['p', 's', 'd', 'i', 'f', 'j', 'o', '#', '$'].flatMap((param) => [
      { code: `it.each([])("%${param}", () => {})` },
      { code: `it.each([])("%%${param}", () => {})` },
      { code: `it.each([])("%${param}%", () => {})` },
      { code: `it.each([])("%%${param}%", () => {})` },
      { code: `it.each([])("%${param}!", () => {})` },
      {
        code: `
        test.each([
          [1, 1, 2],
          [1, 2, 3],
          [2, 1, 3],
        ])(\`.add(%${param}, %${param})\`, (a, b, expected) => {
          expect(a + b).toBe(expected);
        });
      `,
      },
      {
        code: `
        test.each([
          [1, 1, 2],
          [1, 2, 3],
          [2, 1, 3],
        ])(\`.add(%${param}, 1)\`, (a, b, expected) => {
          expect(a + b).toBe(expected);
        });
      `,
      },
    ]),
  ],
  invalid: [
    // disallowedWords
    {
      code: `test("the correct way to properly handle all things", () => {});`,
      options: [{ disallowedWords: ['correct', 'properly', 'all'] }],
      errors: [{ messageId: 'disallowedWord' }],
    },
    {
      code: `describe("the correct way to do things", function () {})`,
      options: [{ disallowedWords: ['correct'] }],
      errors: [{ messageId: 'disallowedWord' }],
    },
    {
      code: `it("has ALL the things", () => {})`,
      options: [{ disallowedWords: ['all'] }],
      errors: [{ messageId: 'disallowedWord' }],
    },
    {
      code: `xdescribe("every single one of them", function () {})`,
      options: [{ disallowedWords: ['every'] }],
      errors: [{ messageId: 'disallowedWord' }],
    },
    {
      code: `describe('Very Descriptive Title Goes Here', function () {})`,
      options: [{ disallowedWords: ['descriptive'] }],
      errors: [{ messageId: 'disallowedWord' }],
    },
    {
      code: 'test(`that the value is set properly`, function () {})',
      options: [{ disallowedWords: ['properly'] }],
      errors: [{ messageId: 'disallowedWord' }],
    },

    // mustMatch / mustNotMatch
    {
      code: `
        describe('things to test', () => {
          describe('unit tests #unit', () => {
            it('is true', () => {
              expect(true).toBe(true);
            });
          });

          describe('e2e tests #e4e', () => {
            it('is another test #e2e #jest4life', () => {});
          });
        });
      `,
      options: [
        {
          mustNotMatch: /(?:#(?!unit|e2e))\w+/u.source,
          mustMatch: `^[^#]+$|(?:#(?:unit|e2e))`,
        },
      ],
      errors: [{ messageId: 'mustNotMatch' }, { messageId: 'mustNotMatch' }],
    },
    {
      code: `
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
      `,
      options: [
        {
          mustNotMatch: /(?:#(?!unit|e2e))\w+/u.source,
          mustMatch: `^[^#]+$|(?:#(?:unit|e2e))`,
        },
      ],
      errors: [{ messageId: 'mustNotMatch' }, { messageId: 'mustNotMatch' }],
    },
    {
      code: `
        describe('things to test', () => {
          describe('unit tests #unit', () => {
            it('is true', () => {
              expect(true).toBe(true);
            });
          });

          describe('e2e tests #e4e', () => {
            it('is another test #e2e #jest4life', () => {});
          });
        });
      `,
      options: [
        {
          mustNotMatch: [
            /(?:#(?!unit|e2e))\w+/u.source,
            'Please include "#unit" or "#e2e" in titles',
          ],
          mustMatch: [
            `^[^#]+$|(?:#(?:unit|e2e))`,
            'Please include "#unit" or "#e2e" in titles',
          ],
        },
      ],
      errors: [
        { messageId: 'mustNotMatchCustom' },
        { messageId: 'mustNotMatchCustom' },
      ],
    },
    {
      code: `
        describe('things to test', () => {
          describe('unit tests #unit', () => {
            it('is true', () => {
              expect(true).toBe(true);
            });
          });

          describe('e2e tests #e4e', () => {
            it('is another test #e2e #jest4life', () => {});
          });
        });
      `,
      options: [
        {
          mustNotMatch: { describe: [/(?:#(?!unit|e2e))\w+/u.source] },
          mustMatch: { describe: `^[^#]+$|(?:#(?:unit|e2e))` },
        },
      ],
      errors: [{ messageId: 'mustNotMatch' }],
    },
    {
      code: `
        describe('things to test', () => {
          describe('unit tests #unit', () => {
            it('is true', () => {
              expect(true).toBe(true);
            });
          });

          describe('e2e tests #e4e', () => {
            it('is another test #e2e #jest4life', () => {});
          });
        });
      `,
      options: [
        {
          mustNotMatch: {
            describe: [
              /(?:#(?!unit|e2e))\w+/u.source,
              'Please include "#unit" or "#e2e" in describe titles',
            ],
          },
          mustMatch: { describe: `^[^#]+$|(?:#(?:unit|e2e))` },
        },
      ],
      errors: [{ messageId: 'mustNotMatchCustom' }],
    },
    {
      code: `
        describe('things to test', () => {
          describe('unit tests #unit', () => {
            it('is true', () => {
              expect(true).toBe(true);
            });
          });

          describe('e2e tests #e4e', () => {
            it('is another test #e2e #jest4life', () => {});
          });
        });
      `,
      options: [
        {
          mustNotMatch: { describe: /(?:#(?!unit|e2e))\w+/u.source },
          mustMatch: { it: `^[^#]+$|(?:#(?:unit|e2e))` },
        },
      ],
      errors: [{ messageId: 'mustNotMatch' }],
    },
    {
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
      `,
      options: [
        {
          mustNotMatch: {
            describe: [
              /(?:#(?!unit|e2e))\w+/u.source,
              'Please include "#unit" or "#e2e" in describe titles',
            ],
          },
          mustMatch: {
            it: [
              `^[^#]+$|(?:#(?:unit|e2e))`,
              'Please include "#unit" or "#e2e" in it titles',
            ],
          },
        },
      ],
      errors: [
        { messageId: 'mustMatchCustom' },
        { messageId: 'mustNotMatchCustom' },
      ],
    },
    {
      code: `test("the correct way to properly handle all things", () => {});`,
      options: [{ mustMatch: `#(?:unit|integration|e2e)` }],
      errors: [{ messageId: 'mustMatch' }],
    },
    {
      code: `describe("the test", () => {});`,
      options: [{ mustMatch: { describe: `#(?:unit|integration|e2e)` } }],
      errors: [{ messageId: 'mustMatch' }],
    },
    {
      code: `xdescribe("the test", () => {});`,
      options: [{ mustMatch: { describe: `#(?:unit|integration|e2e)` } }],
      errors: [{ messageId: 'mustMatch' }],
    },
    {
      code: `describe.skip("the test", () => {});`,
      options: [{ mustMatch: { describe: `#(?:unit|integration|e2e)` } }],
      errors: [{ messageId: 'mustMatch' }],
    },

    // Title must be a string
    {
      code: `it.each([])(1, () => {});`,
      errors: [{ messageId: 'titleMustBeString' }],
    },
    {
      code: `it(String(/.+/), () => {});`,
      options: [{ ignoreTypeOfTestName: false }],
      errors: [{ messageId: 'titleMustBeString' }],
    },
    {
      code: `const foo = "my-title"; it(foo, () => {});`,
      options: [{ ignoreTypeOfTestName: false }],
      errors: [{ messageId: 'titleMustBeString' }],
    },
    {
      code: `it.skip.each([])(1, () => {});`,
      errors: [{ messageId: 'titleMustBeString' }],
    },
    {
      code: 'it.skip.each``(1, () => {});',
      errors: [{ messageId: 'titleMustBeString' }],
    },
    {
      code: `it(123, () => {});`,
      errors: [{ messageId: 'titleMustBeString' }],
    },
    {
      code: `it.concurrent(123, () => {});`,
      errors: [{ messageId: 'titleMustBeString' }],
    },
    {
      code: `it(1 + 2 + 3, () => {});`,
      errors: [{ messageId: 'titleMustBeString' }],
    },
    {
      code: `it.concurrent(1 + 2 + 3, () => {});`,
      errors: [{ messageId: 'titleMustBeString' }],
    },
    {
      code: `test.skip(123, () => {});`,
      options: [{ ignoreTypeOfDescribeName: true }],
      errors: [{ messageId: 'titleMustBeString' }],
    },
    {
      code: `describe(String(/.+/), () => {});`,
      errors: [{ messageId: 'titleMustBeString' }],
    },
    {
      code: `describe(myFunction, () => 1);`,
      options: [{ ignoreTypeOfDescribeName: false }],
      errors: [{ messageId: 'titleMustBeString' }],
    },
    {
      code: `describe(myFunction, () => {});`,
      errors: [{ messageId: 'titleMustBeString' }],
    },
    {
      code: `xdescribe(myFunction, () => {});`,
      errors: [{ messageId: 'titleMustBeString' }],
    },
    {
      code: `describe(6, function () {})`,
      errors: [{ messageId: 'titleMustBeString' }],
    },
    {
      code: `describe.skip(123, () => {});`,
      errors: [{ messageId: 'titleMustBeString' }],
    },

    // No empty title
    {
      code: `describe("", function () {})`,
      errors: [{ messageId: 'emptyTitle' }],
    },
    {
      code: `
        describe('foo', () => {
          it('', () => {});
        });
      `,
      errors: [{ messageId: 'emptyTitle' }],
    },
    {
      code: `it("", function () {})`,
      errors: [{ messageId: 'emptyTitle' }],
    },
    {
      code: `it.concurrent("", function () {})`,
      errors: [{ messageId: 'emptyTitle' }],
    },
    {
      code: `test("", function () {})`,
      errors: [{ messageId: 'emptyTitle' }],
    },
    {
      code: `test.concurrent("", function () {})`,
      errors: [{ messageId: 'emptyTitle' }],
    },
    {
      code: 'test(``, function () {})',
      errors: [{ messageId: 'emptyTitle' }],
    },
    {
      code: 'test.concurrent(``, function () {})',
      errors: [{ messageId: 'emptyTitle' }],
    },
    {
      code: `xdescribe('', () => {})`,
      errors: [{ messageId: 'emptyTitle' }],
    },
    {
      code: `xit('', () => {})`,
      errors: [{ messageId: 'emptyTitle' }],
    },
    {
      code: `xtest('', () => {})`,
      errors: [{ messageId: 'emptyTitle' }],
    },

    // Accidental space (with autofix outputs)
    {
      code: `describe(" foo", function () {})`,
      output: `describe("foo", function () {})`,
      errors: [{ messageId: 'accidentalSpace' }],
    },
    {
      code: `describe.each()(" foo", function () {})`,
      output: `describe.each()("foo", function () {})`,
      errors: [{ messageId: 'accidentalSpace' }],
    },
    {
      code: `describe.only.each()(" foo", function () {})`,
      output: `describe.only.each()("foo", function () {})`,
      errors: [{ messageId: 'accidentalSpace' }],
    },
    {
      code: `describe(" foo foe fum", function () {})`,
      output: `describe("foo foe fum", function () {})`,
      errors: [{ messageId: 'accidentalSpace' }],
    },
    {
      code: `describe("foo foe fum ", function () {})`,
      output: `describe("foo foe fum", function () {})`,
      errors: [{ messageId: 'accidentalSpace' }],
    },
    {
      code: `fdescribe(" foo", function () {})`,
      output: `fdescribe("foo", function () {})`,
      errors: [{ messageId: 'accidentalSpace' }],
    },
    {
      code: `fdescribe(' foo', function () {})`,
      output: `fdescribe('foo', function () {})`,
      errors: [{ messageId: 'accidentalSpace' }],
    },
    {
      code: `xdescribe(" foo", function () {})`,
      output: `xdescribe("foo", function () {})`,
      errors: [{ messageId: 'accidentalSpace' }],
    },
    {
      code: `it(" foo", function () {})`,
      output: `it("foo", function () {})`,
      errors: [{ messageId: 'accidentalSpace' }],
    },
    {
      code: `it.concurrent(" foo", function () {})`,
      output: `it.concurrent("foo", function () {})`,
      errors: [{ messageId: 'accidentalSpace' }],
    },
    {
      code: `fit(" foo", function () {})`,
      output: `fit("foo", function () {})`,
      errors: [{ messageId: 'accidentalSpace' }],
    },
    {
      code: `it.skip(" foo", function () {})`,
      output: `it.skip("foo", function () {})`,
      errors: [{ messageId: 'accidentalSpace' }],
    },
    {
      code: `fit("foo ", function () {})`,
      output: `fit("foo", function () {})`,
      errors: [{ messageId: 'accidentalSpace' }],
    },
    {
      code: `it.skip("foo ", function () {})`,
      output: `it.skip("foo", function () {})`,
      errors: [{ messageId: 'accidentalSpace' }],
    },
    {
      code: `
        import { test as testThat } from '@jest/globals';

        testThat('foo works ', () => {});
      `,
      output: `
        import { test as testThat } from '@jest/globals';

        testThat('foo works', () => {});
      `,
      errors: [{ messageId: 'accidentalSpace' }],
    },
    {
      code: `xit(" foo", function () {})`,
      output: `xit("foo", function () {})`,
      errors: [{ messageId: 'accidentalSpace' }],
    },
    {
      code: `test(" foo", function () {})`,
      output: `test("foo", function () {})`,
      errors: [{ messageId: 'accidentalSpace' }],
    },
    {
      code: `test.concurrent(" foo", function () {})`,
      output: `test.concurrent("foo", function () {})`,
      errors: [{ messageId: 'accidentalSpace' }],
    },
    {
      code: 'test(` foo`, function () {})',
      output: 'test(`foo`, function () {})',
      errors: [{ messageId: 'accidentalSpace' }],
    },
    {
      code: 'test.concurrent(` foo`, function () {})',
      output: 'test.concurrent(`foo`, function () {})',
      errors: [{ messageId: 'accidentalSpace' }],
    },
    {
      code: 'test(` foo bar bang`, function () {})',
      output: 'test(`foo bar bang`, function () {})',
      errors: [{ messageId: 'accidentalSpace' }],
    },
    {
      code: 'test.concurrent(` foo bar bang`, function () {})',
      output: 'test.concurrent(`foo bar bang`, function () {})',
      errors: [{ messageId: 'accidentalSpace' }],
    },
    {
      code: 'test(` foo bar bang  `, function () {})',
      output: 'test(`foo bar bang`, function () {})',
      errors: [{ messageId: 'accidentalSpace' }],
    },
    {
      code: 'test.concurrent(` foo bar bang  `, function () {})',
      output: 'test.concurrent(`foo bar bang`, function () {})',
      errors: [{ messageId: 'accidentalSpace' }],
    },
    {
      code: `xtest(" foo", function () {})`,
      output: `xtest("foo", function () {})`,
      errors: [{ messageId: 'accidentalSpace' }],
    },
    {
      code: `xtest(" foo  ", function () {})`,
      output: `xtest("foo", function () {})`,
      errors: [{ messageId: 'accidentalSpace' }],
    },
    {
      code: `
        describe(' foo', () => {
          it('bar', () => {})
        })
      `,
      output: `
        describe('foo', () => {
          it('bar', () => {})
        })
      `,
      errors: [{ messageId: 'accidentalSpace' }],
    },
    {
      code: `
        describe('foo', () => {
          it(' bar', () => {})
        })
      `,
      output: `
        describe('foo', () => {
          it('bar', () => {})
        })
      `,
      errors: [{ messageId: 'accidentalSpace' }],
    },

    // Duplicate-prefix (with autofix)
    {
      code: `describe("describe foo", function () {})`,
      output: `describe("foo", function () {})`,
      errors: [{ messageId: 'duplicatePrefix' }],
    },
    {
      code: `fdescribe("describe foo", function () {})`,
      output: `fdescribe("foo", function () {})`,
      errors: [{ messageId: 'duplicatePrefix' }],
    },
    {
      code: `xdescribe("describe foo", function () {})`,
      output: `xdescribe("foo", function () {})`,
      errors: [{ messageId: 'duplicatePrefix' }],
    },
    {
      code: `describe('describe foo', function () {})`,
      output: `describe('foo', function () {})`,
      errors: [{ messageId: 'duplicatePrefix' }],
    },
    {
      code: `test("test foo", function () {})`,
      output: `test("foo", function () {})`,
      errors: [{ messageId: 'duplicatePrefix' }],
    },
    {
      code: `xtest("test foo", function () {})`,
      output: `xtest("foo", function () {})`,
      errors: [{ messageId: 'duplicatePrefix' }],
    },
    {
      code: `it("it foo", function () {})`,
      output: `it("foo", function () {})`,
      errors: [{ messageId: 'duplicatePrefix' }],
    },
    {
      code: `fit("it foo", function () {})`,
      output: `fit("foo", function () {})`,
      errors: [{ messageId: 'duplicatePrefix' }],
    },
    {
      code: `xit("it foo", function () {})`,
      output: `xit("foo", function () {})`,
      errors: [{ messageId: 'duplicatePrefix' }],
    },
    {
      code: `it("it foos it correctly", function () {})`,
      output: `it("foos it correctly", function () {})`,
      errors: [{ messageId: 'duplicatePrefix' }],
    },
    {
      code: `
        describe('describe foo', () => {
          it('bar', () => {})
        })
      `,
      output: `
        describe('foo', () => {
          it('bar', () => {})
        })
      `,
      errors: [{ messageId: 'duplicatePrefix' }],
    },
    {
      code: `
        describe('describe foo', () => {
          it('describes things correctly', () => {})
        })
      `,
      output: `
        describe('foo', () => {
          it('describes things correctly', () => {})
        })
      `,
      errors: [{ messageId: 'duplicatePrefix' }],
    },
    {
      code: `
        describe('foo', () => {
          it('it bar', () => {})
        })
      `,
      output: `
        describe('foo', () => {
          it('bar', () => {})
        })
      `,
      errors: [{ messageId: 'duplicatePrefix' }],
    },

    // Invalid each printf specifiers
    {
      code: `it.each([])("%y", () => {})`,
      errors: [{ messageId: 'invalidEachSpecifier' }],
    },
    {
      code: `it.each([])("%s+%y", () => {})`,
      errors: [{ messageId: 'invalidEachSpecifier' }],
    },
    {
      code: `it.each([])("%y+%%y", () => {})`,
      errors: [{ messageId: 'invalidEachSpecifier' }],
    },
    {
      code: `it.each([])("%%y+%y", () => {})`,
      errors: [{ messageId: 'invalidEachSpecifier' }],
    },
    {
      code: `it.each([])("%%%x", () => {})`,
      errors: [{ messageId: 'invalidEachSpecifier' }],
    },
    {
      code: `
        test.each([
          [1, 1, 2],
          [1, 2, 3],
          [2, 1, 3],
        ])('.add(%x, %x)', (a, b, expected) => {
          expect(a + b).toBe(expected);
        });
      `,
      errors: [{ messageId: 'invalidEachSpecifier' }],
    },
    {
      code: `
        test.only.each([
          [1, 1, 2],
          [1, 2, 3],
          [2, 1, 3],
        ])('.add(%x, %y)', (a, b, expected) => {
          expect(a + b).toBe(expected);
        });
      `,
      errors: [{ messageId: 'invalidEachSpecifier' }],
    },
    {
      code: `
        describe.each([
          [1, 1, 2],
          [1, 2, 3],
          [2, 1, 3],
        ])('.add(%x, %y)', (a, b, expected) => {
          it('is correct', () => {
            expect(a + b).toBe(expected);
          });
        });
      `,
      errors: [{ messageId: 'invalidEachSpecifier' }],
    },
    {
      code: `
        test.each([
          [1, 1, 2],
          [1, 2, 3],
          [2, 1, 3],
        ])('.add(%i, %x)', (a, b, expected) => {
          expect(a + b).toBe(expected);
        });
      `,
      errors: [{ messageId: 'invalidEachSpecifier' }],
    },
    {
      code: `
        it.skip.each(entries)('.add(%i, %x)', (a, b, expected) => {
          expect(a + b).toBe(expected);
        });
      `,
      errors: [{ messageId: 'invalidEachSpecifier' }],
    },
    {
      code: `
        const entries = [
          [1, 1, 2],
          [1, 2, 3],
          [2, 1, 3],
        ];

        test.each(entries)('.add(%i, %x)', (a, b, expected) => {
          expect(a + b).toBe(expected);
        });
      `,
      errors: [{ messageId: 'invalidEachSpecifier' }],
    },

    ...['P', 'S', 'D', 'I', 'F', 'J', 'O'].map((letter) => ({
      code: `it.each([])("%${letter}", () => {})`,
      errors: [{ messageId: 'invalidEachSpecifier' }],
    })),
  ],
});
