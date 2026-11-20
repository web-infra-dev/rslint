// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('require-hook', {
  valid: [
    ...(
      [
        'describe()',
        'describe("just a title")',
        `
      describe('a test', () =>
        test('something', () => {
          expect(true).toBe(true);
        }));
   \u0020`,
        `
      test('it', () => {
        //
      });
   \u0020`,
        `
      const { myFn } = require('../functions');

      test('myFn', () => {
        expect(myFn()).toBe(1);
      });
   \u0020`,
        `
        import { myFn } from '../functions';

        test('myFn', () => {
          expect(myFn()).toBe(1);
        });
     \u0020`,
        `
      class MockLogger {
        log() {}
      }

      test('myFn', () => {
        expect(myFn()).toBe(1);
      });
   \u0020`,
        `
      const { myFn } = require('../functions');

      describe('myFn', () => {
        it('returns one', () => {
          expect(myFn()).toBe(1);
        });
      });
   \u0020`,
        `
      describe('some tests', () => {
        it('is true', () => {
          expect(true).toBe(true);
        });
      });
   \u0020`,
        `
      describe('some tests', () => {
        it('is true', () => {
          expect(true).toBe(true);
        });

        describe('more tests', () => {
          it('is false', () => {
            expect(true).toBe(false);
          });
        });
      });
   \u0020`,
        `
      describe('some tests', () => {
        let consoleLogSpy;

        beforeEach(() => {
          consoleLogSpy = jest.spyOn(console, 'log');
        });

        it('prints a message', () => {
          printMessage('hello world');

          expect(consoleLogSpy).toHaveBeenCalledWith('hello world');
        });
      });
   \u0020`,
        `
      let consoleErrorSpy = null;

      beforeEach(() => {
        consoleErrorSpy = jest.spyOn(console, 'error');
      });
   \u0020`,
        `
      let consoleErrorSpy = undefined;

      beforeEach(() => {
        consoleErrorSpy = jest.spyOn(console, 'error');
      });
   \u0020`,
        `
      describe('some tests', () => {
        beforeEach(() => {
          setup();
        });
      });
   \u0020`,
        `
      beforeEach(() => {
        initializeCityDatabase();
      });

      afterEach(() => {
        clearCityDatabase();
      });

      test('city database has Vienna', () => {
        expect(isCity('Vienna')).toBeTruthy();
      });

      test('city database has San Juan', () => {
        expect(isCity('San Juan')).toBeTruthy();
      });
   \u0020`,
        `
      describe('cities', () => {
        beforeEach(() => {
          initializeCityDatabase();
        });

        test('city database has Vienna', () => {
          expect(isCity('Vienna')).toBeTruthy();
        });

        test('city database has San Juan', () => {
          expect(isCity('San Juan')).toBeTruthy();
        });

        afterEach(() => {
          clearCityDatabase();
        });
      });
   \u0020`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    {
      name: 'run 0 valid 16',
      code: `
        enableAutoDestroy(afterEach);

        describe('some tests', () => {
          it('is false', () => {
            expect(true).toBe(true);
          });
        });
     \u0020`,
      options: [
        {
          allowedFunctionCalls: ['enableAutoDestroy'],
        },
      ],
    },
    {
      name: 'run 0 valid 17',
      code: `
      import { myFn } from '../functions';

      // todo: https://github.com/DefinitelyTyped/DefinitelyTyped/pull/56545
      declare module 'eslint' {
        namespace ESLint {
          interface LintResult {
            fatalErrorCount: number;
          }
        }
      }

      test('myFn', () => {
        expect(myFn()).toBe(1);
      });
   \u0020`,
    },
    ...(
      [
        'describe()',
        'describe("just a title")',
        `describe('a test', () =>
  test('something', () => {
    expect(true).toBe(true);
  }));`,
        `test('it', () => {
  //
});`,
        `const { myFn } = require('../functions');

test('myFn', () => {
  expect(myFn()).toBe(1);
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream require-hook / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
    })),
    {
      name: 'upstream require-hook / valid 5',
      code: `import { myFn } from '../functions';

test('myFn', () => {
  expect(myFn()).toBe(1);
});`,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2017,
      },
    },
    ...(
      [
        `class MockLogger {
  log() {}
}

test('myFn', () => {
  expect(myFn()).toBe(1);
});`,
        `const { myFn } = require('../functions');

describe('myFn', () => {
  it('returns one', () => {
    expect(myFn()).toBe(1);
  });
});`,
        `describe('some tests', () => {
  it('is true', () => {
    expect(true).toBe(true);
  });
});`,
        `describe('some tests', () => {
  it('is true', () => {
    expect(true).toBe(true);
  });

  describe('more tests', () => {
    it('is false', () => {
      expect(true).toBe(false);
    });
  });
});`,
        `describe('some tests', () => {
  let consoleLogSpy;

  beforeEach(() => {
    consoleLogSpy = jest.spyOn(console, 'log');
  });

  it('prints a message', () => {
    printMessage('hello world');

    expect(consoleLogSpy).toHaveBeenCalledWith('hello world');
  });
});`,
        `let consoleErrorSpy = null;

beforeEach(() => {
  consoleErrorSpy = jest.spyOn(console, 'error');
});`,
        `let consoleErrorSpy = undefined;

beforeEach(() => {
  consoleErrorSpy = jest.spyOn(console, 'error');
});`,
        `describe('some tests', () => {
  beforeEach(() => {
    setup();
  });
});`,
        `beforeEach(() => {
  initializeCityDatabase();
});

afterEach(() => {
  clearCityDatabase();
});

test('city database has Vienna', () => {
  expect(isCity('Vienna')).toBeTruthy();
});

test('city database has San Juan', () => {
  expect(isCity('San Juan')).toBeTruthy();
});`,
        `describe('cities', () => {
  beforeEach(() => {
    initializeCityDatabase();
  });

  test('city database has Vienna', () => {
    expect(isCity('Vienna')).toBeTruthy();
  });

  test('city database has San Juan', () => {
    expect(isCity('San Juan')).toBeTruthy();
  });

  afterEach(() => {
    clearCityDatabase();
  });
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream require-hook / valid ${index + 6}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
    })),
    {
      name: 'upstream require-hook / valid 16',
      code: `enableAutoDestroy(afterEach);

describe('some tests', () => {
  it('is false', () => {
    expect(true).toBe(true);
  });
});`,
      options: [
        {
          allowedFunctionCalls: ['enableAutoDestroy'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
    },
    {
      name: 'upstream require-hook: typescript edition / valid 0',
      code: `import { myFn } from '../functions';

// todo: https://github.com/DefinitelyTyped/DefinitelyTyped/pull/56545
declare module 'eslint' {
  namespace ESLint {
    interface LintResult {
      fatalErrorCount: number;
    }
  }
}

test('myFn', () => {
  expect(myFn()).toBe(1);
});`,
    },
  ],
  invalid: [
    ...(
      [
        ['setup();', 1, 1],
        [
          `
        describe('some tests', () => {
          setup();
        });
     \u0020`,
          3,
          11,
        ],
      ] as const
    ).map(([code, line, column], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'useHook',
          message: 'This should be done within a hook',
          line,
          column,
          endLine: line,
          endColumn: column + 8,
        },
      ],
    })),
    ...(
      [
        [
          `
        let { setup } = require('./test-utils');

        describe('some tests', () => {
          setup();
        });
     \u0020`,
          2,
          9,
          49,
          5,
          11,
          19,
        ],
        [
          `
        describe('some tests', () => {
          setup();

          it('is true', () => {
            expect(true).toBe(true);
          });

          describe('more tests', () => {
            setup();

            it('is false', () => {
              expect(true).toBe(false);
            });
          });
        });
     \u0020`,
          3,
          11,
          19,
          10,
          13,
          21,
        ],
        [
          `
        let consoleErrorSpy = jest.spyOn(console, 'error');

        describe('when loading cities from the api', () => {
          let consoleWarnSpy = jest.spyOn(console, 'warn');
        });
     \u0020`,
          2,
          9,
          60,
          5,
          11,
          60,
        ],
      ] as const
    ).map(
      (
        [code, line1, column1, endColumn1, line2, column2, endColumn2],
        index,
      ): TestCase => ({
        name: `run 0 invalid ${index + 2}`,
        code,
        errors: [
          {
            messageId: 'useHook',
            message: 'This should be done within a hook',
            line: line1,
            column: column1,
            endLine: line1,
            endColumn: endColumn1,
          },
          {
            messageId: 'useHook',
            message: 'This should be done within a hook',
            line: line2,
            column: column2,
            endLine: line2,
            endColumn: endColumn2,
          },
        ],
      }),
    ),
    ...(
      [
        [
          `
        let consoleErrorSpy = null;

        describe('when loading cities from the api', () => {
          let consoleWarnSpy = jest.spyOn(console, 'warn');
        });
     \u0020`,
          5,
          11,
          60,
        ],
        ['let value = 1', 1, 1, 14],
        [
          "let consoleErrorSpy, consoleWarnSpy = jest.spyOn(console, 'error');",
          1,
          1,
          68,
        ],
        [
          "let consoleErrorSpy = jest.spyOn(console, 'error'), consoleWarnSpy;",
          1,
          1,
          68,
        ],
      ] as const
    ).map(([code, line, column, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index + 5}`,
      code,
      errors: [
        {
          messageId: 'useHook',
          message: 'This should be done within a hook',
          line,
          column,
          endLine: line,
          endColumn,
        },
      ],
    })),
    {
      name: 'run 0 invalid 9',
      code: `
        import { database, isCity } from '../database';
        import { loadCities } from '../api';

        jest.mock('../api');

        const initializeCityDatabase = () => {
          database.addCity('Vienna');
          database.addCity('San Juan');
          database.addCity('Wellington');
        };

        const clearCityDatabase = () => {
          database.clear();
        };

        initializeCityDatabase();

        test('that persists cities', () => {
          expect(database.cities.length).toHaveLength(3);
        });

        test('city database has Vienna', () => {
          expect(isCity('Vienna')).toBeTruthy();
        });

        test('city database has San Juan', () => {
          expect(isCity('San Juan')).toBeTruthy();
        });

        describe('when loading cities from the api', () => {
          let consoleWarnSpy = jest.spyOn(console, 'warn');

          loadCities.mockResolvedValue(['Wellington', 'London']);

          it('does not duplicate cities', async () => {
            await database.loadCities();

            expect(database.cities).toHaveLength(4);
          });

          it('logs any duplicates', async () => {
            await database.loadCities();

            expect(consoleWarnSpy).toHaveBeenCalledWith(
              'Ignored duplicate cities: Wellington',
            );
          });
        });

        clearCityDatabase();
     \u0020`,
      errors: [
        {
          messageId: 'useHook',
          message: 'This should be done within a hook',
          line: 17,
          column: 9,
          endLine: 17,
          endColumn: 34,
        },
        {
          messageId: 'useHook',
          message: 'This should be done within a hook',
          line: 32,
          column: 11,
          endLine: 32,
          endColumn: 60,
        },
        {
          messageId: 'useHook',
          message: 'This should be done within a hook',
          line: 34,
          column: 11,
          endLine: 34,
          endColumn: 66,
        },
        {
          messageId: 'useHook',
          message: 'This should be done within a hook',
          line: 51,
          column: 9,
          endLine: 51,
          endColumn: 29,
        },
      ],
    },
    {
      name: 'run 0 invalid 10',
      code: `
        enableAutoDestroy(afterEach);

        describe('some tests', () => {
          it('is false', () => {
            expect(true).toBe(true);
          });
        });
     \u0020`,
      options: [
        {
          allowedFunctionCalls: ['someOtherName'],
        },
      ],
      errors: [
        {
          messageId: 'useHook',
          message: 'This should be done within a hook',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 38,
        },
      ],
    },
    {
      name: 'run 0 invalid 11',
      code: `
        import { setup } from '../test-utils';

        // todo: https://github.com/DefinitelyTyped/DefinitelyTyped/pull/56545
        declare module 'eslint' {
          namespace ESLint {
            interface LintResult {
              fatalErrorCount: number;
            }
          }
        }

        describe('some tests', () => {
          setup();
        });
     \u0020`,
      errors: [
        {
          messageId: 'useHook',
          message: 'This should be done within a hook',
          line: 14,
          column: 11,
          endLine: 14,
          endColumn: 19,
        },
      ],
    },
    ...(
      [
        ['setup();', 1, 1],
        [
          `describe('some tests', () => {
  setup();
});`,
          2,
          3,
        ],
      ] as const
    ).map(([code, line, column], index): TestCase => ({
      name: `upstream require-hook / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'useHook',
          message: 'This should be done within a hook',
          line,
          column,
          endLine: line,
          endColumn: column + 8,
        },
      ],
    })),
    ...(
      [
        [
          `let { setup } = require('./test-utils');

describe('some tests', () => {
  setup();
});`,
          1,
          1,
          41,
          4,
          3,
          11,
        ],
        [
          `describe('some tests', () => {
  setup();

  it('is true', () => {
    expect(true).toBe(true);
  });

  describe('more tests', () => {
    setup();

    it('is false', () => {
      expect(true).toBe(false);
    });
  });
});`,
          2,
          3,
          11,
          9,
          5,
          13,
        ],
        [
          `let consoleErrorSpy = jest.spyOn(console, 'error');

describe('when loading cities from the api', () => {
  let consoleWarnSpy = jest.spyOn(console, 'warn');
});`,
          1,
          1,
          52,
          4,
          3,
          52,
        ],
      ] as const
    ).map(
      (
        [code, line1, column1, endColumn1, line2, column2, endColumn2],
        index,
      ): TestCase => ({
        name: `upstream require-hook / invalid ${index + 2}`,
        code,
        languageOptions: {
          ecmaVersion: 2017,
        },
        errors: [
          {
            messageId: 'useHook',
            message: 'This should be done within a hook',
            line: line1,
            column: column1,
            endLine: line1,
            endColumn: endColumn1,
          },
          {
            messageId: 'useHook',
            message: 'This should be done within a hook',
            line: line2,
            column: column2,
            endLine: line2,
            endColumn: endColumn2,
          },
        ],
      }),
    ),
    ...(
      [
        [
          `let consoleErrorSpy = null;

describe('when loading cities from the api', () => {
  let consoleWarnSpy = jest.spyOn(console, 'warn');
});`,
          4,
          3,
          52,
        ],
        ['let value = 1', 1, 1, 14],
        [
          "let consoleErrorSpy, consoleWarnSpy = jest.spyOn(console, 'error');",
          1,
          1,
          68,
        ],
        [
          "let consoleErrorSpy = jest.spyOn(console, 'error'), consoleWarnSpy;",
          1,
          1,
          68,
        ],
      ] as const
    ).map(([code, line, column, endColumn], index): TestCase => ({
      name: `upstream require-hook / invalid ${index + 5}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'useHook',
          message: 'This should be done within a hook',
          line,
          column,
          endLine: line,
          endColumn,
        },
      ],
    })),
    {
      name: 'upstream require-hook / invalid 9',
      code: `import { database, isCity } from '../database';
import { loadCities } from '../api';

jest.mock('../api');

const initializeCityDatabase = () => {
  database.addCity('Vienna');
  database.addCity('San Juan');
  database.addCity('Wellington');
};

const clearCityDatabase = () => {
  database.clear();
};

initializeCityDatabase();

test('that persists cities', () => {
  expect(database.cities.length).toHaveLength(3);
});

test('city database has Vienna', () => {
  expect(isCity('Vienna')).toBeTruthy();
});

test('city database has San Juan', () => {
  expect(isCity('San Juan')).toBeTruthy();
});

describe('when loading cities from the api', () => {
  let consoleWarnSpy = jest.spyOn(console, 'warn');

  loadCities.mockResolvedValue(['Wellington', 'London']);

  it('does not duplicate cities', async () => {
    await database.loadCities();

    expect(database.cities).toHaveLength(4);
  });

  it('logs any duplicates', async () => {
    await database.loadCities();

    expect(consoleWarnSpy).toHaveBeenCalledWith(
      'Ignored duplicate cities: Wellington',
    );
  });
});

clearCityDatabase();`,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'useHook',
          message: 'This should be done within a hook',
          line: 16,
          column: 1,
          endLine: 16,
          endColumn: 26,
        },
        {
          messageId: 'useHook',
          message: 'This should be done within a hook',
          line: 31,
          column: 3,
          endLine: 31,
          endColumn: 52,
        },
        {
          messageId: 'useHook',
          message: 'This should be done within a hook',
          line: 33,
          column: 3,
          endLine: 33,
          endColumn: 58,
        },
        {
          messageId: 'useHook',
          message: 'This should be done within a hook',
          line: 50,
          column: 1,
          endLine: 50,
          endColumn: 21,
        },
      ],
    },
    {
      name: 'upstream require-hook / invalid 10',
      code: `enableAutoDestroy(afterEach);

describe('some tests', () => {
  it('is false', () => {
    expect(true).toBe(true);
  });
});`,
      options: [
        {
          allowedFunctionCalls: ['someOtherName'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'useHook',
          message: 'This should be done within a hook',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      name: 'upstream require-hook: typescript edition / invalid 0',
      code: `import { setup } from '../test-utils';

// todo: https://github.com/DefinitelyTyped/DefinitelyTyped/pull/56545
declare module 'eslint' {
  namespace ESLint {
    interface LintResult {
      fatalErrorCount: number;
    }
  }
}

describe('some tests', () => {
  setup();
});`,
      errors: [
        {
          messageId: 'useHook',
          message: 'This should be done within a hook',
          line: 13,
          column: 3,
          endLine: 13,
          endColumn: 11,
        },
      ],
    },
  ],
});
