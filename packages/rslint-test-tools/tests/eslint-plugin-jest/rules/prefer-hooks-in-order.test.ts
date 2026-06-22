import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-hooks-in-order', {} as never, {
  valid: [
    { code: 'beforeAll(() => {})' },
    { code: 'beforeEach(() => {})' },
    { code: 'afterEach(() => {})' },
    { code: 'afterAll(() => {})' },
    { code: 'describe(() => {})' },
    {
      code: `
      beforeAll(() => {});
      beforeEach(() => {});
      afterEach(() => {});
      afterAll(() => {});
    `,
    },
    {
      code: `
      describe('foo', () => {
        someSetupFn();
        beforeEach(() => {});
        afterEach(() => {});

        test('bar', () => {
          someFn();
        });
      });
    `,
    },
    {
      code: `
      beforeAll(() => {});
      afterAll(() => {});
    `,
    },
    {
      code: `
      beforeEach(() => {});
      afterEach(() => {});
    `,
    },
    {
      code: `
      beforeAll(() => {});
      afterEach(() => {});
    `,
    },
    {
      code: `
      beforeAll(() => {});
      beforeEach(() => {});
    `,
    },
    {
      code: `
      afterEach(() => {});
      afterAll(() => {});
    `,
    },
    {
      code: `
      beforeAll(() => {});
      beforeAll(() => {});
    `,
    },
    {
      code: `
      describe('my test', () => {
        afterEach(() => {});
        afterAll(() => {});
      });
    `,
    },
    {
      code: `
      describe('my test', () => {
        afterEach(() => {});
        afterAll(() => {});

        doSomething();

        beforeAll(() => {});
        beforeEach(() => {});
      });
    `,
    },
    {
      code: `
      describe('my test', () => {
        afterEach(() => {});
        afterAll(() => {});

        it('is a test', () => {});

        beforeAll(() => {});
        beforeEach(() => {});
      });
    `,
    },
    {
      code: `
      describe('my test', () => {
        afterAll(() => {});

        describe('when something is true', () => {
          beforeAll(() => {});
          beforeEach(() => {});
        });
      });
    `,
    },
    {
      code: `
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
    `,
    },
    {
      code: `
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
    `,
    },
    {
      code: `
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
    `,
    },
    {
      code: `
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
    `,
    },
  ],
  invalid: [
    {
      code: `
        const withDatabase = () => {
          afterAll(() => {
            removeMyDatabase();
          });
          beforeAll(() => {
            createMyDatabase();
          });
        };
      `,
      errors: [
        {
          messageId: 'reorderHooks',
          data: { currentHook: 'beforeAll', previousHook: 'afterAll' },
          column: 3,
          line: 5,
        },
      ],
    },
    {
      code: `
        afterAll(() => {
          removeMyDatabase();
        });
        beforeAll(() => {
          createMyDatabase();
        });
      `,
      errors: [
        {
          messageId: 'reorderHooks',
          data: { currentHook: 'beforeAll', previousHook: 'afterAll' },
          column: 1,
          line: 4,
        },
      ],
    },
    {
      code: `
        afterAll(() => {});
        beforeAll(() => {});
      `,
      errors: [
        {
          messageId: 'reorderHooks',
          data: { currentHook: 'beforeAll', previousHook: 'afterAll' },
          column: 1,
          line: 2,
        },
      ],
    },
    {
      code: `
        afterEach(() => {});
        beforeEach(() => {});
      `,
      errors: [
        {
          messageId: 'reorderHooks',
          data: { currentHook: 'beforeEach', previousHook: 'afterEach' },
          column: 1,
          line: 2,
        },
      ],
    },
    {
      code: `
        afterEach(() => {});
        beforeAll(() => {});
      `,
      errors: [
        {
          messageId: 'reorderHooks',
          data: { currentHook: 'beforeAll', previousHook: 'afterEach' },
          column: 1,
          line: 2,
        },
      ],
    },
    {
      code: `
        beforeEach(() => {});
        beforeAll(() => {});
      `,
      errors: [
        {
          messageId: 'reorderHooks',
          data: { currentHook: 'beforeAll', previousHook: 'beforeEach' },
          column: 1,
          line: 2,
        },
      ],
    },
    {
      code: `
        afterAll(() => {});
        afterEach(() => {});
      `,
      errors: [
        {
          messageId: 'reorderHooks',
          data: { currentHook: 'afterEach', previousHook: 'afterAll' },
          column: 1,
          line: 2,
        },
      ],
    },
    {
      code: `
        afterAll(() => {});
        // The afterEach should do this
        // This comment does not matter for the order
        afterEach(() => {});
      `,
      errors: [
        {
          messageId: 'reorderHooks',
          data: { currentHook: 'afterEach', previousHook: 'afterAll' },
          column: 1,
          line: 4,
        },
      ],
    },
    {
      code: `
        afterAll(() => {});
        afterAll(() => {});
        afterEach(() => {});
      `,
      errors: [
        {
          messageId: 'reorderHooks',
          data: { currentHook: 'afterEach', previousHook: 'afterAll' },
          column: 1,
          line: 3,
        },
      ],
    },
    {
      code: `
        describe('my test', () => {
          afterAll(() => {});
          afterEach(() => {});
        });
      `,
      errors: [
        {
          messageId: 'reorderHooks',
          data: { currentHook: 'afterEach', previousHook: 'afterAll' },
          column: 3,
          line: 3,
        },
      ],
    },
    {
      code: `
        describe('my test', () => {
          afterAll(() => {});
          afterEach(() => {});

          doSomething();

          beforeEach(() => {});
          beforeAll(() => {});
        });
      `,
      errors: [
        {
          messageId: 'reorderHooks',
          data: { currentHook: 'afterEach', previousHook: 'afterAll' },
          column: 3,
          line: 3,
        },
        {
          messageId: 'reorderHooks',
          data: { currentHook: 'beforeAll', previousHook: 'beforeEach' },
          column: 3,
          line: 8,
        },
      ],
    },
    {
      code: `
        describe('my test', () => {
          afterAll(() => {});
          afterEach(() => {});

          it('is a test', () => {});

          beforeEach(() => {});
          beforeAll(() => {});
        });
      `,
      errors: [
        {
          messageId: 'reorderHooks',
          data: { currentHook: 'afterEach', previousHook: 'afterAll' },
          column: 3,
          line: 3,
        },
        {
          messageId: 'reorderHooks',
          data: { currentHook: 'beforeAll', previousHook: 'beforeEach' },
          column: 3,
          line: 8,
        },
      ],
    },
    {
      code: `
        describe('my test', () => {
          afterAll(() => {});

          describe('when something is true', () => {
            beforeEach(() => {});
            beforeAll(() => {});
          });
        });
      `,
      errors: [
        {
          messageId: 'reorderHooks',
          data: { currentHook: 'beforeAll', previousHook: 'beforeEach' },
          column: 5,
          line: 6,
        },
      ],
    },
    {
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
      `,
      errors: [
        {
          messageId: 'reorderHooks',
          data: { currentHook: 'beforeAll', previousHook: 'afterAll' },
          column: 3,
          line: 4,
        },
        {
          messageId: 'reorderHooks',
          data: { currentHook: 'beforeEach', previousHook: 'afterEach' },
          column: 5,
          line: 9,
        },
      ],
    },
    {
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
      `,
      errors: [
        {
          messageId: 'reorderHooks',
          data: { currentHook: 'afterEach', previousHook: 'afterAll' },
          column: 7,
          line: 22,
        },
      ],
    },
    {
      code: `
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
      `,
      errors: [
        {
          messageId: 'reorderHooks',
          data: { currentHook: 'beforeAll', previousHook: 'beforeEach' },
          column: 5,
          line: 7,
        },
        {
          messageId: 'reorderHooks',
          data: { currentHook: 'afterEach', previousHook: 'afterAll' },
          column: 5,
          line: 18,
        },
      ],
    },
    {
      code: `
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
      `,
      errors: [
        {
          messageId: 'reorderHooks',
          data: { currentHook: 'beforeAll', previousHook: 'beforeEach' },
          column: 3,
          line: 6,
        },
        {
          messageId: 'reorderHooks',
          data: { currentHook: 'beforeEach', previousHook: 'afterEach' },
          column: 5,
          line: 37,
        },
      ],
    },
  ],
});
