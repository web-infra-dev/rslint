package prefer_hooks_in_order_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_hooks_in_order"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferHooksInOrderRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_hooks_in_order.PreferHooksInOrderRule,
		[]rule_tester.ValidTestCase{
			{Code: `beforeAll(() => {})`},
			{Code: `beforeEach(() => {})`},
			{Code: `afterEach(() => {})`},
			{Code: `afterAll(() => {})`},
			{Code: `describe(() => {})`},
			{Code: `
      beforeAll(() => {});
      beforeEach(() => {});
      afterEach(() => {});
      afterAll(() => {});
    `},
			{Code: `
      describe('foo', () => {
        someSetupFn();
        beforeEach(() => {});
        afterEach(() => {});

        test('bar', () => {
          someFn();
        });
      });
    `},
			{Code: `
      beforeAll(() => {});
      afterAll(() => {});
    `},
			{Code: `
      beforeEach(() => {});
      afterEach(() => {});
    `},
			{Code: `
      beforeAll(() => {});
      afterEach(() => {});
    `},
			{Code: `
      beforeAll(() => {});
      beforeEach(() => {});
    `},
			{Code: `
      afterEach(() => {});
      afterAll(() => {});
    `},
			{Code: `
      beforeAll(() => {});
      beforeAll(() => {});
    `},
			{Code: `
      describe('my test', () => {
        afterEach(() => {});
        afterAll(() => {});
      });
    `},
			{Code: `
      describe('my test', () => {
        afterEach(() => {});
        afterAll(() => {});

        doSomething();

        beforeAll(() => {});
        beforeEach(() => {});
      });
    `},
			{Code: `
      describe('my test', () => {
        afterEach(() => {});
        afterAll(() => {});

        it('is a test', () => {});

        beforeAll(() => {});
        beforeEach(() => {});
      });
    `},
			{Code: `
      describe('my test', () => {
        afterAll(() => {});

        describe('when something is true', () => {
          beforeAll(() => {});
          beforeEach(() => {});
        });
      });
    `},
			{Code: `
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
    `},
			{Code: `
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
    `},
			{Code: `
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
    `},
			{Code: `
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
    `},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `
        const withDatabase = () => {
          afterAll(() => {
            removeMyDatabase();
          });
          beforeAll(() => {
            createMyDatabase();
          });
        };
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reorderHooks", Line: 6, Column: 11},
				},
			},
			{
				Code: `
        afterAll(() => {
          removeMyDatabase();
        });
        beforeAll(() => {
          createMyDatabase();
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reorderHooks", Line: 5, Column: 9},
				},
			},
			{
				Code: `
        afterAll(() => {});
        beforeAll(() => {});
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reorderHooks", Line: 3, Column: 9},
				},
			},
			{
				Code: `
        afterEach(() => {});
        beforeEach(() => {});
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reorderHooks", Line: 3, Column: 9},
				},
			},
			{
				Code: `
        afterEach(() => {});
        beforeAll(() => {});
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reorderHooks", Line: 3, Column: 9},
				},
			},
			{
				Code: `
        beforeEach(() => {});
        beforeAll(() => {});
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reorderHooks", Line: 3, Column: 9},
				},
			},
			{
				Code: `
        afterAll(() => {});
        afterEach(() => {});
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reorderHooks", Line: 3, Column: 9},
				},
			},
			{
				Code: `
        afterAll(() => {});
        // The afterEach should do this
        // This comment does not matter for the order
        afterEach(() => {});
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reorderHooks", Line: 5, Column: 9},
				},
			},
			{
				Code: `
        afterAll(() => {});
        afterAll(() => {});
        afterEach(() => {});
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reorderHooks", Line: 4, Column: 9},
				},
			},
			{
				Code: `
        describe('my test', () => {
          afterAll(() => {});
          afterEach(() => {});
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reorderHooks", Line: 4, Column: 11},
				},
			},
			{
				Code: `
        describe('my test', () => {
          afterAll(() => {});
          afterEach(() => {});

          doSomething();

          beforeEach(() => {});
          beforeAll(() => {});
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reorderHooks", Line: 4, Column: 11},
					{MessageId: "reorderHooks", Line: 9, Column: 11},
				},
			},
			{
				Code: `
        describe('my test', () => {
          afterAll(() => {});
          afterEach(() => {});

          it('is a test', () => {});

          beforeEach(() => {});
          beforeAll(() => {});
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reorderHooks", Line: 4, Column: 11},
					{MessageId: "reorderHooks", Line: 9, Column: 11},
				},
			},
			{
				Code: `
        describe('my test', () => {
          afterAll(() => {});

          describe('when something is true', () => {
            beforeEach(() => {});
            beforeAll(() => {});
          });
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reorderHooks", Line: 7, Column: 13},
				},
			},
			{
				Code: `
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
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reorderHooks", Line: 5, Column: 11},
					{MessageId: "reorderHooks", Line: 10, Column: 13},
				},
			},
			{
				Code: `
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
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reorderHooks", Line: 23, Column: 15},
				},
			},
			{
				Code: `
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
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reorderHooks", Line: 8, Column: 13},
					{MessageId: "reorderHooks", Line: 19, Column: 13},
				},
			},
			{
				Code: `
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
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reorderHooks", Line: 7, Column: 11},
					{MessageId: "reorderHooks", Line: 38, Column: 13},
				},
			},
		},
	)
}
