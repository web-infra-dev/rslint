// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('no-done-callback', {
  valid: [
    ...(
      [
        'test("something", () => {})',
        'test("something", async () => {})',
        'test("something", function() {})',
        'test.each``("something", ({ a, b }) => {})',
        'test.each()("something", ({ a, b }) => {})',
        'it.each()("something", ({ a, b }) => {})',
        'it.each([])("something", (a, b) => {})',
        'it.each``("something", ({ a, b }) => {})',
        'it.each([])("something", (a, b) => { a(); b(); })',
        'it.each``("something", ({ a, b }) => { a(); b(); })',
        'test("something", async function () {})',
        'test("something", someArg)',
        'beforeEach(() => {})',
        'beforeAll(async () => {})',
        'afterAll(() => {})',
        'afterAll(async function () {})',
        'afterAll(async function () {}, 5)',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    ...(
      [
        'test("something", () => {})',
        'test("something", async () => {})',
        'test("something", function() {})',
        'test.each``("something", ({ a, b }) => {})',
        'test.each()("something", ({ a, b }) => {})',
        'it.each()("something", ({ a, b }) => {})',
        'it.each([])("something", (a, b) => {})',
        'it.each``("something", ({ a, b }) => {})',
        'it.each([])("something", (a, b) => { a(); b(); })',
        'it.each``("something", ({ a, b }) => { a(); b(); })',
        'test("something", async function () {})',
        'test("something", someArg)',
        'beforeEach(() => {})',
        'beforeAll(async () => {})',
        'afterAll(() => {})',
        'afterAll(async function () {})',
        'afterAll(async function () {}, 5)',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream no-done-callback / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
    })),
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: 'test("something", (...args) => {args[0]();})',
      errors: [
        {
          messageId: 'noDoneCallback',
          message: 'Return a Promise instead of relying on callback parameter',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    ...(
      [
        ['test("something", done => {done();})', 19],
        ['test("something", (done,) => {done();})', 20],
      ] as const
    ).map(([code, column], index): TestCase => ({
      name: `run 0 invalid ${index + 1}`,
      code,
      errors: [
        {
          messageId: 'noDoneCallback',
          message: 'Return a Promise instead of relying on callback parameter',
          line: 1,
          column,
          endLine: 1,
          endColumn: column + 4,
          suggestions: [
            {
              messageId: 'suggestWrappingInPromise',
              desc: 'Wrap in `new Promise(done => ...`',
              output:
                'test("something", () => {return new Promise(done => {done();})})',
            },
          ],
        },
      ],
    })),
    {
      name: 'run 0 invalid 3',
      code: 'test("something", finished => {finished();})',
      errors: [
        {
          messageId: 'noDoneCallback',
          message: 'Return a Promise instead of relying on callback parameter',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 27,
          suggestions: [
            {
              messageId: 'suggestWrappingInPromise',
              desc: 'Wrap in `new Promise(finished => ...`',
              output:
                'test("something", () => {return new Promise(finished => {finished();})})',
            },
          ],
        },
      ],
    },
    ...(
      [
        [
          'test("something", (done) => {done();})',
          20,
          'test("something", () => {return new Promise(done => {done();})})',
        ],
        [
          'test("something", done => done())',
          19,
          'test("something", () => new Promise(done => done()))',
        ],
        [
          'test("something", (done) => done())',
          20,
          'test("something", () => new Promise(done => done()))',
        ],
        [
          'test("something", function(done) {done();})',
          28,
          'test("something", function() {return new Promise(done => {done();})})',
        ],
        [
          'test("something", function (done) {done();})',
          29,
          'test("something", function () {return new Promise(done => {done();})})',
        ],
      ] as const
    ).map(([code, column, suggestionOutput], index): TestCase => ({
      name: `run 0 invalid ${index + 4}`,
      code,
      errors: [
        {
          messageId: 'noDoneCallback',
          message: 'Return a Promise instead of relying on callback parameter',
          line: 1,
          column,
          endLine: 1,
          endColumn: column + 4,
          suggestions: [
            {
              messageId: 'suggestWrappingInPromise',
              desc: 'Wrap in `new Promise(done => ...`',
              output: suggestionOutput,
            },
          ],
        },
      ],
    })),
    ...(
      [
        ['test("something", async done => {done();})', 1, 25],
        ['test("something", async done => done())', 1, 25],
        ['test("something", async function (done) {done();})', 1, 35],
        [
          `
        test('my test', async (done) => {
          await myAsyncTask();
          expect(true).toBe(false);
          done();
        });
     \u0020`,
          2,
          32,
        ],
      ] as const
    ).map(([code, line, column], index): TestCase => ({
      name: `run 0 invalid ${index + 9}`,
      code,
      errors: [
        {
          messageId: 'useAwaitInsteadOfCallback',
          message: 'Use await instead of callback in async functions',
          line,
          column,
          endLine: line,
          endColumn: column + 4,
        },
      ],
    })),
    {
      name: 'run 0 invalid 13',
      code: `
        test('something', (done) => {
          done();
        });
     \u0020`,
      errors: [
        {
          messageId: 'noDoneCallback',
          message: 'Return a Promise instead of relying on callback parameter',
          line: 2,
          column: 28,
          endLine: 2,
          endColumn: 32,
          suggestions: [
            {
              messageId: 'suggestWrappingInPromise',
              desc: 'Wrap in `new Promise(done => ...`',
              output: `
        test('something', () => {return new Promise(done => {
          done();
        })});
     \u0020`,
            },
          ],
        },
      ],
    },
    {
      name: 'run 0 invalid 14',
      code: 'afterEach((...args) => {args[0]();})',
      errors: [
        {
          messageId: 'noDoneCallback',
          message: 'Return a Promise instead of relying on callback parameter',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      name: 'run 0 invalid 15',
      code: 'beforeAll(done => {done();})',
      errors: [
        {
          messageId: 'noDoneCallback',
          message: 'Return a Promise instead of relying on callback parameter',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 15,
          suggestions: [
            {
              messageId: 'suggestWrappingInPromise',
              desc: 'Wrap in `new Promise(done => ...`',
              output:
                'beforeAll(() => {return new Promise(done => {done();})})',
            },
          ],
        },
      ],
    },
    {
      name: 'run 0 invalid 16',
      code: 'beforeAll(finished => {finished();})',
      errors: [
        {
          messageId: 'noDoneCallback',
          message: 'Return a Promise instead of relying on callback parameter',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'suggestWrappingInPromise',
              desc: 'Wrap in `new Promise(finished => ...`',
              output:
                'beforeAll(() => {return new Promise(finished => {finished();})})',
            },
          ],
        },
      ],
    },
    ...(
      [
        [
          'beforeEach((done) => {done();})',
          13,
          'beforeEach(() => {return new Promise(done => {done();})})',
        ],
        [
          'afterAll(done => done())',
          10,
          'afterAll(() => new Promise(done => done()))',
        ],
        [
          'afterEach((done) => done())',
          12,
          'afterEach(() => new Promise(done => done()))',
        ],
        [
          'beforeAll(function(done) {done();})',
          20,
          'beforeAll(function() {return new Promise(done => {done();})})',
        ],
        [
          'afterEach(function (done) {done();})',
          21,
          'afterEach(function () {return new Promise(done => {done();})})',
        ],
      ] as const
    ).map(([code, column, suggestionOutput], index): TestCase => ({
      name: `run 0 invalid ${index + 17}`,
      code,
      errors: [
        {
          messageId: 'noDoneCallback',
          message: 'Return a Promise instead of relying on callback parameter',
          line: 1,
          column,
          endLine: 1,
          endColumn: column + 4,
          suggestions: [
            {
              messageId: 'suggestWrappingInPromise',
              desc: 'Wrap in `new Promise(done => ...`',
              output: suggestionOutput,
            },
          ],
        },
      ],
    })),
    ...(
      [
        ['beforeAll(async done => {done();})', 1, 17],
        ['beforeAll(async done => done())', 1, 17],
        ['beforeAll(async function (done) {done();})', 1, 27],
        [
          `
        afterAll(async (done) => {
          await myAsyncTask();
          done();
        });
     \u0020`,
          2,
          25,
        ],
      ] as const
    ).map(([code, line, column], index): TestCase => ({
      name: `run 0 invalid ${index + 22}`,
      code,
      errors: [
        {
          messageId: 'useAwaitInsteadOfCallback',
          message: 'Use await instead of callback in async functions',
          line,
          column,
          endLine: line,
          endColumn: column + 4,
        },
      ],
    })),
    ...(
      [
        [
          `
        beforeEach((done) => {
          done();
        });
     \u0020`,
          2,
          21,
          `
        beforeEach(() => {return new Promise(done => {
          done();
        })});
     \u0020`,
        ],
        [
          `
        import { beforeEach } from '@jest/globals';

        beforeEach((done) => {
          done();
        });
     \u0020`,
          4,
          21,
          `
        import { beforeEach } from '@jest/globals';

        beforeEach(() => {return new Promise(done => {
          done();
        })});
     \u0020`,
        ],
        [
          `
        import { beforeEach as atTheStartOfEachTest } from '@jest/globals';

        atTheStartOfEachTest((done) => {
          done();
        });
     \u0020`,
          4,
          31,
          `
        import { beforeEach as atTheStartOfEachTest } from '@jest/globals';

        atTheStartOfEachTest(() => {return new Promise(done => {
          done();
        })});
     \u0020`,
        ],
        [
          'test.each``("something", ({ a, b }, done) => { done(); })',
          1,
          37,
          'test.each``("something", () => {return new Promise(done => { done(); })})',
        ],
        [
          'it.each``("something", ({ a, b }, done) => { done(); })',
          1,
          35,
          'it.each``("something", () => {return new Promise(done => { done(); })})',
        ],
      ] as const
    ).map(([code, line, column, suggestionOutput], index): TestCase => ({
      name: `run 0 invalid ${index + 26}`,
      code,
      errors: [
        {
          messageId: 'noDoneCallback',
          message: 'Return a Promise instead of relying on callback parameter',
          line,
          column,
          endLine: line,
          endColumn: column + 4,
          suggestions: [
            {
              messageId: 'suggestWrappingInPromise',
              desc: 'Wrap in `new Promise(done => ...`',
              output: suggestionOutput,
            },
          ],
        },
      ],
    })),
    {
      name: 'upstream no-done-callback / invalid 0',
      code: 'test("something", (...args) => {args[0]();})',
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'noDoneCallback',
          message: 'Return a Promise instead of relying on callback parameter',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    ...(
      [
        ['test("something", done => {done();})', 19],
        ['test("something", (done,) => {done();})', 20],
      ] as const
    ).map(([code, column], index): TestCase => ({
      name: `upstream no-done-callback / invalid ${index + 1}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'noDoneCallback',
          message: 'Return a Promise instead of relying on callback parameter',
          line: 1,
          column,
          endLine: 1,
          endColumn: column + 4,
          suggestions: [
            {
              messageId: 'suggestWrappingInPromise',
              desc: 'Wrap in `new Promise(done => ...`',
              output:
                'test("something", () => {return new Promise(done => {done();})})',
            },
          ],
        },
      ],
    })),
    {
      name: 'upstream no-done-callback / invalid 3',
      code: 'test("something", finished => {finished();})',
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'noDoneCallback',
          message: 'Return a Promise instead of relying on callback parameter',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 27,
          suggestions: [
            {
              messageId: 'suggestWrappingInPromise',
              desc: 'Wrap in `new Promise(finished => ...`',
              output:
                'test("something", () => {return new Promise(finished => {finished();})})',
            },
          ],
        },
      ],
    },
    ...(
      [
        [
          'test("something", (done) => {done();})',
          20,
          'test("something", () => {return new Promise(done => {done();})})',
        ],
        [
          'test("something", done => done())',
          19,
          'test("something", () => new Promise(done => done()))',
        ],
        [
          'test("something", (done) => done())',
          20,
          'test("something", () => new Promise(done => done()))',
        ],
        [
          'test("something", function(done) {done();})',
          28,
          'test("something", function() {return new Promise(done => {done();})})',
        ],
        [
          'test("something", function (done) {done();})',
          29,
          'test("something", function () {return new Promise(done => {done();})})',
        ],
      ] as const
    ).map(([code, column, suggestionOutput], index): TestCase => ({
      name: `upstream no-done-callback / invalid ${index + 4}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'noDoneCallback',
          message: 'Return a Promise instead of relying on callback parameter',
          line: 1,
          column,
          endLine: 1,
          endColumn: column + 4,
          suggestions: [
            {
              messageId: 'suggestWrappingInPromise',
              desc: 'Wrap in `new Promise(done => ...`',
              output: suggestionOutput,
            },
          ],
        },
      ],
    })),
    ...(
      [
        ['test("something", async done => {done();})', 25],
        ['test("something", async done => done())', 25],
        ['test("something", async function (done) {done();})', 35],
        [
          `test('my test', async (done) => {
  await myAsyncTask();
  expect(true).toBe(false);
  done();
});`,
          24,
        ],
      ] as const
    ).map(([code, column], index): TestCase => ({
      name: `upstream no-done-callback / invalid ${index + 9}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'useAwaitInsteadOfCallback',
          message: 'Use await instead of callback in async functions',
          line: 1,
          column,
          endLine: 1,
          endColumn: column + 4,
        },
      ],
    })),
    {
      name: 'upstream no-done-callback / invalid 13',
      code: `test('something', (done) => {
  done();
});`,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'noDoneCallback',
          message: 'Return a Promise instead of relying on callback parameter',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 24,
          suggestions: [
            {
              messageId: 'suggestWrappingInPromise',
              desc: 'Wrap in `new Promise(done => ...`',
              output: `test('something', () => {return new Promise(done => {
  done();
})});`,
            },
          ],
        },
      ],
    },
    {
      name: 'upstream no-done-callback / invalid 14',
      code: 'afterEach((...args) => {args[0]();})',
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'noDoneCallback',
          message: 'Return a Promise instead of relying on callback parameter',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      name: 'upstream no-done-callback / invalid 15',
      code: 'beforeAll(done => {done();})',
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'noDoneCallback',
          message: 'Return a Promise instead of relying on callback parameter',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 15,
          suggestions: [
            {
              messageId: 'suggestWrappingInPromise',
              desc: 'Wrap in `new Promise(done => ...`',
              output:
                'beforeAll(() => {return new Promise(done => {done();})})',
            },
          ],
        },
      ],
    },
    {
      name: 'upstream no-done-callback / invalid 16',
      code: 'beforeAll(finished => {finished();})',
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'noDoneCallback',
          message: 'Return a Promise instead of relying on callback parameter',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'suggestWrappingInPromise',
              desc: 'Wrap in `new Promise(finished => ...`',
              output:
                'beforeAll(() => {return new Promise(finished => {finished();})})',
            },
          ],
        },
      ],
    },
    ...(
      [
        [
          'beforeEach((done) => {done();})',
          13,
          'beforeEach(() => {return new Promise(done => {done();})})',
        ],
        [
          'afterAll(done => done())',
          10,
          'afterAll(() => new Promise(done => done()))',
        ],
        [
          'afterEach((done) => done())',
          12,
          'afterEach(() => new Promise(done => done()))',
        ],
        [
          'beforeAll(function(done) {done();})',
          20,
          'beforeAll(function() {return new Promise(done => {done();})})',
        ],
        [
          'afterEach(function (done) {done();})',
          21,
          'afterEach(function () {return new Promise(done => {done();})})',
        ],
      ] as const
    ).map(([code, column, suggestionOutput], index): TestCase => ({
      name: `upstream no-done-callback / invalid ${index + 17}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'noDoneCallback',
          message: 'Return a Promise instead of relying on callback parameter',
          line: 1,
          column,
          endLine: 1,
          endColumn: column + 4,
          suggestions: [
            {
              messageId: 'suggestWrappingInPromise',
              desc: 'Wrap in `new Promise(done => ...`',
              output: suggestionOutput,
            },
          ],
        },
      ],
    })),
    ...(
      [
        ['beforeAll(async done => {done();})', 17],
        ['beforeAll(async done => done())', 17],
        ['beforeAll(async function (done) {done();})', 27],
        [
          `afterAll(async (done) => {
  await myAsyncTask();
  done();
});`,
          17,
        ],
      ] as const
    ).map(([code, column], index): TestCase => ({
      name: `upstream no-done-callback / invalid ${index + 22}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'useAwaitInsteadOfCallback',
          message: 'Use await instead of callback in async functions',
          line: 1,
          column,
          endLine: 1,
          endColumn: column + 4,
        },
      ],
    })),
    {
      name: 'upstream no-done-callback / invalid 26',
      code: `beforeEach((done) => {
  done();
});`,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'noDoneCallback',
          message: 'Return a Promise instead of relying on callback parameter',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 17,
          suggestions: [
            {
              messageId: 'suggestWrappingInPromise',
              desc: 'Wrap in `new Promise(done => ...`',
              output: `beforeEach(() => {return new Promise(done => {
  done();
})});`,
            },
          ],
        },
      ],
    },
    ...(
      [
        [
          `import { beforeEach } from '@jest/globals';

beforeEach((done) => {
  done();
});`,
          13,
          `import { beforeEach } from '@jest/globals';

beforeEach(() => {return new Promise(done => {
  done();
})});`,
        ],
        [
          `import { beforeEach as atTheStartOfEachTest } from '@jest/globals';

atTheStartOfEachTest((done) => {
  done();
});`,
          23,
          `import { beforeEach as atTheStartOfEachTest } from '@jest/globals';

atTheStartOfEachTest(() => {return new Promise(done => {
  done();
})});`,
        ],
      ] as const
    ).map(([code, column, suggestionOutput], index): TestCase => ({
      name: `upstream no-done-callback / invalid ${index + 27}`,
      code,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'noDoneCallback',
          message: 'Return a Promise instead of relying on callback parameter',
          line: 3,
          column,
          endLine: 3,
          endColumn: column + 4,
          suggestions: [
            {
              messageId: 'suggestWrappingInPromise',
              desc: 'Wrap in `new Promise(done => ...`',
              output: suggestionOutput,
            },
          ],
        },
      ],
    })),
    ...(
      [
        [
          'test.each``("something", ({ a, b }, done) => { done(); })',
          37,
          'test.each``("something", () => {return new Promise(done => { done(); })})',
        ],
        [
          'it.each``("something", ({ a, b }, done) => { done(); })',
          35,
          'it.each``("something", () => {return new Promise(done => { done(); })})',
        ],
      ] as const
    ).map(([code, column, suggestionOutput], index): TestCase => ({
      name: `upstream no-done-callback / invalid ${index + 29}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'noDoneCallback',
          message: 'Return a Promise instead of relying on callback parameter',
          line: 1,
          column,
          endLine: 1,
          endColumn: column + 4,
          suggestions: [
            {
              messageId: 'suggestWrappingInPromise',
              desc: 'Wrap in `new Promise(done => ...`',
              output: suggestionOutput,
            },
          ],
        },
      ],
    })),
  ],
});
