package prefer_hooks_in_order_test

import (
	"fmt"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/prefer_hooks_in_order"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestPreferHooksInOrderUpstream migrates the full valid/invalid suite from
// upstream eslint-plugin-jest's prefer-hooks-in-order coverage as preserved in
// rslint's Jest port. Position assertions cover line/column for every invalid
// case. Rstest-specific lock-in cases live in
// prefer_hooks_in_order_extras_test.go.
func TestPreferHooksInOrderUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_hooks_in_order.PreferHooksInOrderRule,
		[]rule_tester.ValidTestCase{
			// ---- upstream valid ----
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
			// ---- upstream invalid ----
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
					reorderHooksError("beforeAll", "afterAll", 6, 11),
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
					reorderHooksErrorWithRange("beforeAll", "afterAll", 5, 9, 7, 11),
				},
			},
			{
				Code: `
        afterAll(() => {});
        beforeAll(() => {});
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					reorderHooksErrorWithRange("beforeAll", "afterAll", 3, 9, 3, 28),
				},
			},
			{
				Code: `
        afterEach(() => {});
        beforeEach(() => {});
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					reorderHooksError("beforeEach", "afterEach", 3, 9),
				},
			},
			{
				Code: `
        afterEach(() => {});
        beforeAll(() => {});
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					reorderHooksError("beforeAll", "afterEach", 3, 9),
				},
			},
			{
				Code: `
        beforeEach(() => {});
        beforeAll(() => {});
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					reorderHooksError("beforeAll", "beforeEach", 3, 9),
				},
			},
			{
				Code: `
        afterAll(() => {});
        afterEach(() => {});
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					reorderHooksError("afterEach", "afterAll", 3, 9),
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
					reorderHooksError("afterEach", "afterAll", 5, 9),
				},
			},
			{
				Code: `
        afterAll(() => {});
        afterAll(() => {});
        afterEach(() => {});
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					reorderHooksError("afterEach", "afterAll", 4, 9),
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
					reorderHooksError("afterEach", "afterAll", 4, 11),
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
					reorderHooksError("afterEach", "afterAll", 4, 11),
					reorderHooksError("beforeAll", "beforeEach", 9, 11),
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
					reorderHooksError("afterEach", "afterAll", 4, 11),
					reorderHooksError("beforeAll", "beforeEach", 9, 11),
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
					reorderHooksError("beforeAll", "beforeEach", 7, 13),
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
					reorderHooksError("beforeAll", "afterAll", 5, 11),
					reorderHooksError("beforeEach", "afterEach", 10, 13),
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
					reorderHooksError("afterEach", "afterAll", 23, 15),
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
					reorderHooksError("beforeAll", "beforeEach", 8, 13),
					reorderHooksError("afterEach", "afterAll", 19, 13),
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
					reorderHooksError("beforeAll", "beforeEach", 7, 11),
					reorderHooksError("beforeEach", "afterEach", 38, 13),
				},
			},
		},
	)
}

func reorderHooksError(currentHook, previousHook string, line, column int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "reorderHooks",
		Message: fmt.Sprintf(
			"`%s` hooks should be before any `%s` hooks",
			currentHook,
			previousHook,
		),
		Line:   line,
		Column: column,
	}
}

func reorderHooksErrorWithRange(
	currentHook,
	previousHook string,
	line,
	column,
	endLine,
	endColumn int,
) rule_tester.InvalidTestCaseError {
	err := reorderHooksError(currentHook, previousHook, line, column)
	err.EndLine = endLine
	err.EndColumn = endColumn
	return err
}
