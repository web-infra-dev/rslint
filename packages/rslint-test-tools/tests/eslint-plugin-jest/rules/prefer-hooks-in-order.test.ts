// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('prefer-hooks-in-order', {
  valid: [
    ...(
      [
        'beforeAll(() => {})',
        'beforeEach(() => {})',
        'afterEach(() => {})',
        'afterAll(() => {})',
        'describe(() => {})',
        `
      beforeAll(() => {});
      beforeEach(() => {});
      afterEach(() => {});
      afterAll(() => {});
   \u0020`,
        `
      describe('foo', () => {
        someSetupFn();
        beforeEach(() => {});
        afterEach(() => {});

        test('bar', () => {
          someFn();
        });
      });
   \u0020`,
        `
      beforeAll(() => {});
      afterAll(() => {});
   \u0020`,
        `
      beforeEach(() => {});
      afterEach(() => {});
   \u0020`,
        `
      beforeAll(() => {});
      afterEach(() => {});
   \u0020`,
        `
      beforeAll(() => {});
      beforeEach(() => {});
   \u0020`,
        `
      afterEach(() => {});
      afterAll(() => {});
   \u0020`,
        `
      beforeAll(() => {});
      beforeAll(() => {});
   \u0020`,
        `
      describe('my test', () => {
        afterEach(() => {});
        afterAll(() => {});
      });
   \u0020`,
        `
      describe('my test', () => {
        afterEach(() => {});
        afterAll(() => {});

        doSomething();

        beforeAll(() => {});
        beforeEach(() => {});
      });
   \u0020`,
        `
      describe('my test', () => {
        afterEach(() => {});
        afterAll(() => {});

        it('is a test', () => {});

        beforeAll(() => {});
        beforeEach(() => {});
      });
   \u0020`,
        `
      describe('my test', () => {
        afterAll(() => {});

        describe('when something is true', () => {
          beforeAll(() => {});
          beforeEach(() => {});
        });
      });
   \u0020`,
        `
      describe('my test', () => {
        afterAll(() => {});

        describe('when something is true', () => {
          beforeAll(() => {});
          beforeEach(() => {});

          it('does something', () => {});

          beforeAll(() => {});
          beforeEach(() => {});
        });

        beforeAll(() => {});
        beforeEach(() => {});
      });

      describe('my test', () => {
        beforeAll(() => {});
        beforeEach(() => {});
        afterAll(() => {});

        describe('when something is true', () => {
          it('does something', () => {});

          beforeAll(() => {});
          beforeEach(() => {});
        });

        beforeAll(() => {});
        beforeEach(() => {});
      });
   \u0020`,
        `
      const withDatabase = () => {
        beforeAll(() => {
          createMyDatabase();
        });
        afterAll(() => {
          removeMyDatabase();
        });
      };

      describe('my test', () => {
        withDatabase();

        afterAll(() => {});

        describe('when something is true', () => {
          beforeAll(() => {});
          beforeEach(() => {});

          it('does something', () => {});

          beforeAll(() => {});
          beforeEach(() => {});
        });

        beforeAll(() => {});
        beforeEach(() => {});
      });

      describe('my test', () => {
        beforeAll(() => {});
        beforeEach(() => {});
        afterAll(() => {});

        withDatabase();

        describe('when something is true', () => {
          it('does something', () => {});

          beforeAll(() => {});
          beforeEach(() => {});
        });

        beforeAll(() => {});
        beforeEach(() => {});
      });
   \u0020`,
        `
      describe('foo', () => {
        beforeAll(() => {
          createMyDatabase();
        });

        beforeEach(() => {
          seedMyDatabase();
        });

        it('accepts this input', () => {
          // ...
        });

        it('returns that value', () => {
          // ...
        });

        describe('when the database has specific values', () => {
          const specificValue = '...';

          beforeEach(() => {
            seedMyDatabase(specificValue);
          });

          it('accepts that input', () => {
            // ...
          });

          it('throws an error', () => {
            // ...
          });

          beforeEach(() => {
            mockLogger();
          });

          afterEach(() => {
            clearLogger();
          });

          it('logs a message', () => {
            // ...
          });
        });

        afterAll(() => {
          removeMyDatabase();
        });
      });
   \u0020`,
        `
      describe('A file with a lot of test', () => {
        beforeAll(() => {
          setupTheDatabase();
          createMocks();
        });

        beforeAll(() => {
          doEvenMore();
        });

        beforeEach(() => {
          cleanTheDatabase();
          resetSomeThings();
        });

        afterEach(() => {
          cleanTheDatabase();
          resetSomeThings();
        });

        afterAll(() => {
          closeTheDatabase();
          stop();
        });

        it('does something', () => {
          const thing = getThing();
          expect(thing).toBe('something');
        });

        it('throws', () => {
          // Do something that throws
        });

        describe('Also have tests in here', () => {
          afterAll(() => {});
          it('tests something', () => {});
          it('tests something else', () => {});
          beforeAll(()=>{});
        });
      });
   \u0020`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    ...(
      [
        'beforeAll(() => {})',
        'beforeEach(() => {})',
        'afterEach(() => {})',
        'afterAll(() => {})',
        'describe(() => {})',
        `beforeAll(() => {});
beforeEach(() => {});
afterEach(() => {});
afterAll(() => {});`,
        `describe('foo', () => {
  someSetupFn();
  beforeEach(() => {});
  afterEach(() => {});

  test('bar', () => {
    someFn();
  });
});`,
        `beforeAll(() => {});
afterAll(() => {});`,
        `beforeEach(() => {});
afterEach(() => {});`,
        `beforeAll(() => {});
afterEach(() => {});`,
        `beforeAll(() => {});
beforeEach(() => {});`,
        `afterEach(() => {});
afterAll(() => {});`,
        `beforeAll(() => {});
beforeAll(() => {});`,
        `describe('my test', () => {
  afterEach(() => {});
  afterAll(() => {});
});`,
        `describe('my test', () => {
  afterEach(() => {});
  afterAll(() => {});

  doSomething();

  beforeAll(() => {});
  beforeEach(() => {});
});`,
        `describe('my test', () => {
  afterEach(() => {});
  afterAll(() => {});

  it('is a test', () => {});

  beforeAll(() => {});
  beforeEach(() => {});
});`,
        `describe('my test', () => {
  afterAll(() => {});

  describe('when something is true', () => {
    beforeAll(() => {});
    beforeEach(() => {});
  });
});`,
        `describe('my test', () => {
  afterAll(() => {});

  describe('when something is true', () => {
    beforeAll(() => {});
    beforeEach(() => {});

    it('does something', () => {});

    beforeAll(() => {});
    beforeEach(() => {});
  });

  beforeAll(() => {});
  beforeEach(() => {});
});

describe('my test', () => {
  beforeAll(() => {});
  beforeEach(() => {});
  afterAll(() => {});

  describe('when something is true', () => {
    it('does something', () => {});

    beforeAll(() => {});
    beforeEach(() => {});
  });

  beforeAll(() => {});
  beforeEach(() => {});
});`,
        `const withDatabase = () => {
  beforeAll(() => {
    createMyDatabase();
  });
  afterAll(() => {
    removeMyDatabase();
  });
};

describe('my test', () => {
  withDatabase();

  afterAll(() => {});

  describe('when something is true', () => {
    beforeAll(() => {});
    beforeEach(() => {});

    it('does something', () => {});

    beforeAll(() => {});
    beforeEach(() => {});
  });

  beforeAll(() => {});
  beforeEach(() => {});
});

describe('my test', () => {
  beforeAll(() => {});
  beforeEach(() => {});
  afterAll(() => {});

  withDatabase();

  describe('when something is true', () => {
    it('does something', () => {});

    beforeAll(() => {});
    beforeEach(() => {});
  });

  beforeAll(() => {});
  beforeEach(() => {});
});`,
        `describe('foo', () => {
  beforeAll(() => {
    createMyDatabase();
  });

  beforeEach(() => {
    seedMyDatabase();
  });

  it('accepts this input', () => {
    // ...
  });

  it('returns that value', () => {
    // ...
  });

  describe('when the database has specific values', () => {
    const specificValue = '...';

    beforeEach(() => {
      seedMyDatabase(specificValue);
    });

    it('accepts that input', () => {
      // ...
    });

    it('throws an error', () => {
      // ...
    });

    beforeEach(() => {
      mockLogger();
    });

    afterEach(() => {
      clearLogger();
    });

    it('logs a message', () => {
      // ...
    });
  });

  afterAll(() => {
    removeMyDatabase();
  });
});`,
        `describe('A file with a lot of test', () => {
  beforeAll(() => {
    setupTheDatabase();
    createMocks();
  });

  beforeAll(() => {
    doEvenMore();
  });

  beforeEach(() => {
    cleanTheDatabase();
    resetSomeThings();
  });

  afterEach(() => {
    cleanTheDatabase();
    resetSomeThings();
  });

  afterAll(() => {
    closeTheDatabase();
    stop();
  });

  it('does something', () => {
    const thing = getThing();
    expect(thing).toBe('something');
  });

  it('throws', () => {
    // Do something that throws
  });

  describe('Also have tests in here', () => {
    afterAll(() => {});
    it('tests something', () => {});
    it('tests something else', () => {});
    beforeAll(()=>{});
  });
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream prefer-hooks-in-order / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
    })),
  ],
  invalid: [
    ...(
      [
        [
          `
        const withDatabase = () => {
          afterAll(() => {
            removeMyDatabase();
          });
          beforeAll(() => {
            createMyDatabase();
          });
        };
     \u0020`,
          '`beforeAll` hooks should be before any `afterAll` hooks',
          6,
          11,
          8,
          13,
        ],
        [
          `
        afterAll(() => {
          removeMyDatabase();
        });
        beforeAll(() => {
          createMyDatabase();
        });
     \u0020`,
          '`beforeAll` hooks should be before any `afterAll` hooks',
          5,
          9,
          7,
          11,
        ],
        [
          `
        afterAll(() => {});
        beforeAll(() => {});
     \u0020`,
          '`beforeAll` hooks should be before any `afterAll` hooks',
          3,
          9,
          3,
          28,
        ],
        [
          `
        afterEach(() => {});
        beforeEach(() => {});
     \u0020`,
          '`beforeEach` hooks should be before any `afterEach` hooks',
          3,
          9,
          3,
          29,
        ],
        [
          `
        afterEach(() => {});
        beforeAll(() => {});
     \u0020`,
          '`beforeAll` hooks should be before any `afterEach` hooks',
          3,
          9,
          3,
          28,
        ],
        [
          `
        beforeEach(() => {});
        beforeAll(() => {});
     \u0020`,
          '`beforeAll` hooks should be before any `beforeEach` hooks',
          3,
          9,
          3,
          28,
        ],
        [
          `
        afterAll(() => {});
        afterEach(() => {});
     \u0020`,
          '`afterEach` hooks should be before any `afterAll` hooks',
          3,
          9,
          3,
          28,
        ],
        [
          `
        afterAll(() => {});
        // The afterEach should do this
        // This comment does not matter for the order
        afterEach(() => {});
     \u0020`,
          '`afterEach` hooks should be before any `afterAll` hooks',
          5,
          9,
          5,
          28,
        ],
        [
          `
        afterAll(() => {});
        afterAll(() => {});
        afterEach(() => {});
     \u0020`,
          '`afterEach` hooks should be before any `afterAll` hooks',
          4,
          9,
          4,
          28,
        ],
        [
          `
        describe('my test', () => {
          afterAll(() => {});
          afterEach(() => {});
        });
     \u0020`,
          '`afterEach` hooks should be before any `afterAll` hooks',
          4,
          11,
          4,
          30,
        ],
      ] as const
    ).map(
      ([code, message, line, column, endLine, endColumn], index): TestCase => ({
        name: `run 0 invalid ${index}`,
        code,
        errors: [
          {
            messageId: 'reorderHooks',
            message,
            line,
            column,
            endLine,
            endColumn,
          },
        ],
      }),
    ),
    ...(
      [
        `
        describe('my test', () => {
          afterAll(() => {});
          afterEach(() => {});

          doSomething();

          beforeEach(() => {});
          beforeAll(() => {});
        });
     \u0020`,
        `
        describe('my test', () => {
          afterAll(() => {});
          afterEach(() => {});

          it('is a test', () => {});

          beforeEach(() => {});
          beforeAll(() => {});
        });
     \u0020`,
      ] as const
    ).map((code, index): TestCase => ({
      name: `run 0 invalid ${index + 10}`,
      code,
      errors: [
        {
          messageId: 'reorderHooks',
          message: '`afterEach` hooks should be before any `afterAll` hooks',
          line: 4,
          column: 11,
          endLine: 4,
          endColumn: 30,
        },
        {
          messageId: 'reorderHooks',
          message: '`beforeAll` hooks should be before any `beforeEach` hooks',
          line: 9,
          column: 11,
          endLine: 9,
          endColumn: 30,
        },
      ],
    })),
    {
      name: 'run 0 invalid 12',
      code: `
        describe('my test', () => {
          afterAll(() => {});

          describe('when something is true', () => {
            beforeEach(() => {});
            beforeAll(() => {});
          });
        });
     \u0020`,
      errors: [
        {
          messageId: 'reorderHooks',
          message: '`beforeAll` hooks should be before any `beforeEach` hooks',
          line: 7,
          column: 13,
          endLine: 7,
          endColumn: 32,
        },
      ],
    },
    {
      name: 'run 0 invalid 13',
      code: `
        describe('my test', () => {
          beforeAll(() => {});
          afterAll(() => {});
          beforeAll(() => {});

          describe('when something is true', () => {
            beforeAll(() => {});
            afterEach(() => {});
            beforeEach(() => {});
            afterEach(() => {});
          });
        });
     \u0020`,
      errors: [
        {
          messageId: 'reorderHooks',
          message: '`beforeAll` hooks should be before any `afterAll` hooks',
          line: 5,
          column: 11,
          endLine: 5,
          endColumn: 30,
        },
        {
          messageId: 'reorderHooks',
          message: '`beforeEach` hooks should be before any `afterEach` hooks',
          line: 10,
          column: 13,
          endLine: 10,
          endColumn: 33,
        },
      ],
    },
    {
      name: 'run 0 invalid 14',
      code: `
        describe('my test', () => {
          beforeAll(() => {});
          beforeAll(() => {});
          afterAll(() => {});

          it('foo nested', () => {
            // this is a test
          });

          describe('when something is true', () => {
            beforeAll(() => {});
            afterEach(() => {});

            it('foo nested', () => {
              // this is a test
            });

            describe('deeply nested', () => {
              afterAll(() => {});
              afterAll(() => {});
              // This comment does nothing
              afterEach(() => {});

              it('foo nested', () => {
                // this is a test
              });
            })
            beforeEach(() => {});
            afterEach(() => {});
          });
        });
     \u0020`,
      errors: [
        {
          messageId: 'reorderHooks',
          message: '`afterEach` hooks should be before any `afterAll` hooks',
          line: 23,
          column: 15,
          endLine: 23,
          endColumn: 34,
        },
      ],
    },
    ...(
      [
        [
          `
        describe('my test', () => {
          const setupDatabase = () => {
            beforeEach(() => {
              initDatabase();
              fillWithData();
            });
            beforeAll(() => {
              setupMocks();
            });
          };

          it('foo', () => {
            // this is a test
          });

          describe('my nested test', () => {
            afterAll(() => {});
            afterEach(() => {});

            it('foo nested', () => {
              // this is a test
            });
          });
        });
     \u0020`,
          8,
          13,
          10,
          '`afterEach` hooks should be before any `afterAll` hooks',
          19,
          19,
          32,
        ],
        [
          `
        describe('foo', () => {
          beforeEach(() => {
            seedMyDatabase();
          });

          beforeAll(() => {
            createMyDatabase();
          });

          it('accepts this input', () => {
            // ...
          });

          it('returns that value', () => {
            // ...
          });

          describe('when the database has specific values', () => {
            const specificValue = '...';

            beforeEach(() => {
              seedMyDatabase(specificValue);
            });

            it('accepts that input', () => {
              // ...
            });

            it('throws an error', () => {
              // ...
            });

            afterEach(() => {
              clearLogger();
            });

            beforeEach(() => {
              mockLogger();
            });

            it('logs a message', () => {
              // ...
            });
          });

          afterAll(() => {
            removeMyDatabase();
          });
        });
     \u0020`,
          7,
          11,
          9,
          '`beforeEach` hooks should be before any `afterEach` hooks',
          38,
          40,
          15,
        ],
      ] as const
    ).map(
      (
        [code, line1, column1, endLine1, message2, line2, endLine2, endColumn2],
        index,
      ): TestCase => ({
        name: `run 0 invalid ${index + 15}`,
        code,
        errors: [
          {
            messageId: 'reorderHooks',
            message:
              '`beforeAll` hooks should be before any `beforeEach` hooks',
            line: line1,
            column: column1,
            endLine: endLine1,
            endColumn: column1 + 2,
          },
          {
            messageId: 'reorderHooks',
            message: message2,
            line: line2,
            column: 13,
            endLine: endLine2,
            endColumn: endColumn2,
          },
        ],
      }),
    ),
    ...(
      [
        [
          `const withDatabase = () => {
  afterAll(() => {
    removeMyDatabase();
  });
  beforeAll(() => {
    createMyDatabase();
  });
};`,
          '`beforeAll` hooks should be before any `afterAll` hooks',
          5,
          3,
          7,
          5,
        ],
        [
          `afterAll(() => {
  removeMyDatabase();
});
beforeAll(() => {
  createMyDatabase();
});`,
          '`beforeAll` hooks should be before any `afterAll` hooks',
          4,
          1,
          6,
          3,
        ],
        [
          `afterAll(() => {});
beforeAll(() => {});`,
          '`beforeAll` hooks should be before any `afterAll` hooks',
          2,
          1,
          2,
          20,
        ],
        [
          `afterEach(() => {});
beforeEach(() => {});`,
          '`beforeEach` hooks should be before any `afterEach` hooks',
          2,
          1,
          2,
          21,
        ],
        [
          `afterEach(() => {});
beforeAll(() => {});`,
          '`beforeAll` hooks should be before any `afterEach` hooks',
          2,
          1,
          2,
          20,
        ],
        [
          `beforeEach(() => {});
beforeAll(() => {});`,
          '`beforeAll` hooks should be before any `beforeEach` hooks',
          2,
          1,
          2,
          20,
        ],
        [
          `afterAll(() => {});
afterEach(() => {});`,
          '`afterEach` hooks should be before any `afterAll` hooks',
          2,
          1,
          2,
          20,
        ],
        [
          `afterAll(() => {});
// The afterEach should do this
// This comment does not matter for the order
afterEach(() => {});`,
          '`afterEach` hooks should be before any `afterAll` hooks',
          4,
          1,
          4,
          20,
        ],
        [
          `afterAll(() => {});
afterAll(() => {});
afterEach(() => {});`,
          '`afterEach` hooks should be before any `afterAll` hooks',
          3,
          1,
          3,
          20,
        ],
        [
          `describe('my test', () => {
  afterAll(() => {});
  afterEach(() => {});
});`,
          '`afterEach` hooks should be before any `afterAll` hooks',
          3,
          3,
          3,
          22,
        ],
      ] as const
    ).map(
      ([code, message, line, column, endLine, endColumn], index): TestCase => ({
        name: `upstream prefer-hooks-in-order / invalid ${index}`,
        code,
        languageOptions: {
          ecmaVersion: 2015,
        },
        errors: [
          {
            messageId: 'reorderHooks',
            message,
            line,
            column,
            endLine,
            endColumn,
          },
        ],
      }),
    ),
    ...(
      [
        `describe('my test', () => {
  afterAll(() => {});
  afterEach(() => {});

  doSomething();

  beforeEach(() => {});
  beforeAll(() => {});
});`,
        `describe('my test', () => {
  afterAll(() => {});
  afterEach(() => {});

  it('is a test', () => {});

  beforeEach(() => {});
  beforeAll(() => {});
});`,
      ] as const
    ).map((code, index): TestCase => ({
      name: `upstream prefer-hooks-in-order / invalid ${index + 10}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'reorderHooks',
          message: '`afterEach` hooks should be before any `afterAll` hooks',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 22,
        },
        {
          messageId: 'reorderHooks',
          message: '`beforeAll` hooks should be before any `beforeEach` hooks',
          line: 8,
          column: 3,
          endLine: 8,
          endColumn: 22,
        },
      ],
    })),
    {
      name: 'upstream prefer-hooks-in-order / invalid 12',
      code: `describe('my test', () => {
  afterAll(() => {});

  describe('when something is true', () => {
    beforeEach(() => {});
    beforeAll(() => {});
  });
});`,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'reorderHooks',
          message: '`beforeAll` hooks should be before any `beforeEach` hooks',
          line: 6,
          column: 5,
          endLine: 6,
          endColumn: 24,
        },
      ],
    },
    {
      name: 'upstream prefer-hooks-in-order / invalid 13',
      code: `describe('my test', () => {
  beforeAll(() => {});
  afterAll(() => {});
  beforeAll(() => {});

  describe('when something is true', () => {
    beforeAll(() => {});
    afterEach(() => {});
    beforeEach(() => {});
    afterEach(() => {});
  });
});`,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'reorderHooks',
          message: '`beforeAll` hooks should be before any `afterAll` hooks',
          line: 4,
          column: 3,
          endLine: 4,
          endColumn: 22,
        },
        {
          messageId: 'reorderHooks',
          message: '`beforeEach` hooks should be before any `afterEach` hooks',
          line: 9,
          column: 5,
          endLine: 9,
          endColumn: 25,
        },
      ],
    },
    {
      name: 'upstream prefer-hooks-in-order / invalid 14',
      code: `describe('my test', () => {
  beforeAll(() => {});
  beforeAll(() => {});
  afterAll(() => {});

  it('foo nested', () => {
    // this is a test
  });

  describe('when something is true', () => {
    beforeAll(() => {});
    afterEach(() => {});

    it('foo nested', () => {
      // this is a test
    });

    describe('deeply nested', () => {
      afterAll(() => {});
      afterAll(() => {});
      // This comment does nothing
      afterEach(() => {});

      it('foo nested', () => {
        // this is a test
      });
    })
    beforeEach(() => {});
    afterEach(() => {});
  });
});`,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'reorderHooks',
          message: '`afterEach` hooks should be before any `afterAll` hooks',
          line: 22,
          column: 7,
          endLine: 22,
          endColumn: 26,
        },
      ],
    },
    ...(
      [
        [
          `describe('my test', () => {
  const setupDatabase = () => {
    beforeEach(() => {
      initDatabase();
      fillWithData();
    });
    beforeAll(() => {
      setupMocks();
    });
  };

  it('foo', () => {
    // this is a test
  });

  describe('my nested test', () => {
    afterAll(() => {});
    afterEach(() => {});

    it('foo nested', () => {
      // this is a test
    });
  });
});`,
          7,
          5,
          9,
          '`afterEach` hooks should be before any `afterAll` hooks',
          18,
          18,
          24,
        ],
        [
          `describe('foo', () => {
  beforeEach(() => {
    seedMyDatabase();
  });

  beforeAll(() => {
    createMyDatabase();
  });

  it('accepts this input', () => {
    // ...
  });

  it('returns that value', () => {
    // ...
  });

  describe('when the database has specific values', () => {
    const specificValue = '...';

    beforeEach(() => {
      seedMyDatabase(specificValue);
    });

    it('accepts that input', () => {
      // ...
    });

    it('throws an error', () => {
      // ...
    });

    afterEach(() => {
      clearLogger();
    });

    beforeEach(() => {
      mockLogger();
    });

    it('logs a message', () => {
      // ...
    });
  });

  afterAll(() => {
    removeMyDatabase();
  });
});`,
          6,
          3,
          8,
          '`beforeEach` hooks should be before any `afterEach` hooks',
          37,
          39,
          7,
        ],
      ] as const
    ).map(
      (
        [code, line1, column1, endLine1, message2, line2, endLine2, endColumn2],
        index,
      ): TestCase => ({
        name: `upstream prefer-hooks-in-order / invalid ${index + 15}`,
        code,
        languageOptions: {
          ecmaVersion: 2015,
        },
        errors: [
          {
            messageId: 'reorderHooks',
            message:
              '`beforeAll` hooks should be before any `beforeEach` hooks',
            line: line1,
            column: column1,
            endLine: endLine1,
            endColumn: column1 + 2,
          },
          {
            messageId: 'reorderHooks',
            message: message2,
            line: line2,
            column: 5,
            endLine: endLine2,
            endColumn: endColumn2,
          },
        ],
      }),
    ),
  ],
});
