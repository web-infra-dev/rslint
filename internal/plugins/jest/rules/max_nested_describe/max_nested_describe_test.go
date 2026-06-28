package max_nested_describe_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/max_nested_describe"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestMaxNestedDescribeRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&max_nested_describe.MaxNestedDescribeRule,
		[]rule_tester.ValidTestCase{
			{Code: `
      describe('foo', function() {
        describe('bar', function () {
          describe('baz', function () {
            describe('qux', function () {
              describe('qux', function () {
                it('should get something', () => {
                  expect(getSomething()).toBe('Something');
                });
              })
            })
          })
        })
      });
    `},
			{Code: `
      describe('foo', function() {
        describe('bar', function () {
          describe('baz', function () {
            describe('qux', function () {
              describe('qux', function () {
                it('should get something', () => {
                  expect(getSomething()).toBe('Something');
                });
              });

              fdescribe('qux', () => {
                it('something', async () => {
                  expect('something').toBe('something');
                });
              });
            })
          })
        })
      });
    `},
			{
				Code: `
        describe('foo', () => {
          describe.only('bar', () => {
            describe.skip('baz', () => {
              it('something', async () => {
                expect('something').toBe('something');
              });
            });
          });
        });
    `,
				Options: []interface{}{
					map[string]interface{}{"max": 3},
				},
			},
			{
				Code: `
        it('something', async () => {
          expect('something').toBe('something');
        });
    `,
				Options: []interface{}{
					map[string]interface{}{"max": 0},
				},
			},
			{
				Code: `
        describe('foo', () => {
          describe('bar', () => {
            describe('baz', () => {
              describe('qux', () => {
                describe('quux', () => {
                  it('something', async () => {
                    expect('something').toBe('something');
                  });
                });
              });
            });
          });
        });
    `,
				Options: []interface{}{
					map[string]interface{}{"max": -1},
				},
			},
			{
				Code: `
        describe('foo', () => {
          describe('bar', () => {
            describe('baz', () => {
              describe('qux', () => {
                describe('quux', () => {
                  it('something', async () => {
                    expect('something').toBe('something');
                  });
                });
              });
            });
          });
        });
    `,
				Options: []interface{}{
					map[string]interface{}{"max": 1.5},
				},
			},
			{Code: `
      describe('foo', () => {
        describe.each(['hello', 'world'])("%s", (a) => {});
      });
    `},
			{Code: "describe('foo', () => {\n  describe.each`\n  foo  | bar\n  ${'1'} | ${'2'}\n  `('$foo $bar', ({ foo, bar }) => {});\n});"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `
        describe('qux', () => {
          it('should get something', () => {
            expect(getSomething()).toBe('Something');
          });
        });
      `,
				Options: []interface{}{
					map[string]interface{}{"max": 0},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceededMaxDepth", Line: 2, Column: 9},
				},
			},
			{
				Code: `
        describe('foo', function() {
          describe('bar', function () {
            describe('baz', function () {
              describe('qux', function () {
                describe('quxx', function () {
                  describe('over limit', function () {
                    it('should get something', () => {
                      expect(getSomething()).toBe('Something');
                    });
                  });
                });
              });
            });
          });
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceededMaxDepth", Line: 7, Column: 19},
				},
			},
			{
				Code: `
        describe('foo', () => {
          describe('bar', () => {
            describe('baz', () => {
              describe('baz1', () => {
                describe('baz2', () => {
                  describe('baz3', () => {
                    it('should get something', () => {
                      expect(getSomething()).toBe('Something');
                    });
                  });

                  describe('baz4', () => {
                    it('should get something', () => {
                      expect(getSomething()).toBe('Something');
                    });
                  });
                });
              });
            });

            describe('qux', function () {
              it('should get something', () => {
                expect(getSomething()).toBe('Something');
              });
            });
          })
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceededMaxDepth", Line: 7, Column: 19},
					{MessageId: "exceededMaxDepth", Line: 13, Column: 19},
				},
			},
			{
				Code: `
        fdescribe('foo', () => {
          describe.only('bar', () => {
            describe.skip('baz', () => {
              it('should get something', () => {
                expect(getSomething()).toBe('Something');
              });
            });

            describe('baz', () => {
              it('should get something', () => {
                expect(getSomething()).toBe('Something');
              });
            });
          });
        });

        xdescribe('qux', () => {
          it('should get something', () => {
            expect(getSomething()).toBe('Something');
          });
        });
      `,
				Options: []interface{}{
					map[string]interface{}{"max": 2},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceededMaxDepth", Line: 4, Column: 13},
					{MessageId: "exceededMaxDepth", Line: 10, Column: 13},
				},
			},
			{
				Code: `
        describe('foo', () => {
          describe.each(['hello', 'world'])("%s", (a) => {});
        });
      `,
				Options: []interface{}{
					map[string]interface{}{"max": 1},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceededMaxDepth", Line: 3, Column: 11},
				},
			},
			{
				Code: "describe('foo', () => {\n  describe.each`\n  foo  | bar\n  ${'1'} | ${'2'}\n  `('$foo $bar', ({ foo, bar }) => {});\n});",
				Options: []interface{}{
					map[string]interface{}{"max": 1},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceededMaxDepth", Line: 2, Column: 3},
				},
			},
		},
	)
}
