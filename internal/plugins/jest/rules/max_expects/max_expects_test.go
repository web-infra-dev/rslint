package max_expects_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/max_expects"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestMaxExpectsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&max_expects.MaxExpectsRule,
		[]rule_tester.ValidTestCase{
			{Code: `test('should pass')`},
			{Code: `test('should pass', () => {})`},
			{Code: `test.skip('should pass', () => {})`},
			{Code: `
      test('should pass', function () {
        expect(true).toBeDefined();
      });
    `},
			{Code: `
      test('should pass', () => {
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
      });
    `},
			{Code: `
      test('should pass', () => {
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        // expect(true).toBeDefined();
      });
    `},
			{Code: `
      it('should pass', () => {
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
      });
    `},
			{Code: `
      test('should pass', async () => {
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
      });
    `},
			{Code: `
      test('should pass', async () => {
        expect.hasAssertions();

        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
      });
    `},
			{Code: `
      test('should pass', async () => {
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toEqual(expect.any(Boolean));
      });
    `},
			{Code: `
      test('should pass', async () => {
        expect.hasAssertions();

        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toEqual(expect.any(Boolean));
      });
    `},
			{Code: `
      describe('test', () => {
        test('should pass', () => {
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
        });
      });
    `},
			{Code: `
      test.each(['should', 'pass'], () => {
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
      });
    `},
			{Code: `
      test('should pass', () => {
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
      });
      test('should pass', () => {
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
      });
    `},
			{Code: `
      test('should not pass', () => {
        const checkValue = (value) => {
          expect(value).toBeDefined();
          expect(value).toBeDefined();
        };

        checkValue(true);
      });
      test('should not pass', () => {
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
      });
    `},
			{Code: `
      test('should not pass', done => {
        emitter.on('event', value => {
          expect(value).toBeDefined();
          expect(value).toBeDefined();
          expect(value).toBeDefined();
          expect(value).toBeDefined();

          done();
        });
      });
      test('should not pass', () => {
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
      });
    `},
			{Code: `
      function myHelper() {
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
      };

      test('should pass', () => {
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
      });
    `},
			{Code: `
      function myHelper1() {
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
      };

      test('should pass', () => {
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
      });

      function myHelper2() {
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
      };
    `},
			{Code: `
      test('should pass', () => {
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
      });

      function myHelper() {
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
      };
    `},
			{Code: `
      const myHelper1 = () => {
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
      };

      test('should pass', function() {
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
      });

      const myHelper2 = function() {
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
      };
    `},
			{
				Code: `
        test('should pass', () => {
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
        });
      `,
				Options: []interface{}{
					map[string]interface{}{"max": 10},
				},
			},
			{
				Code: `
        describe('given decimal places', () => {
          it("test 1", fakeAsync(() => {
            expect(true).toBeTrue();
            expect(true).toBeTrue();
            expect(true).toBeTrue();
          }))

          it("test 2", fakeAsync(() => {
            expect(true).toBeTrue();
            expect(true).toBeTrue();
            expect(true).toBeTrue();
          }))
        })
      `,
				Options: []interface{}{map[string]interface{}{"max": 5}},
			},
			{
				Code: `
        test('schema-invalid max zero falls back to default', () => {
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
        });
      `,
				Options: []interface{}{map[string]interface{}{"max": 0}},
			},
			{
				Code: `
        test('schema-invalid negative max falls back to default', () => {
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
        });
      `,
				Options: []interface{}{map[string]interface{}{"max": -1}},
			},
			{
				Code: `
        test('schema-invalid fractional max falls back to default', () => {
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
        });
      `,
				Options: []interface{}{map[string]interface{}{"max": 1.5}},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `
        test('should not pass', function () {
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceededMaxAssertion", Line: 8, Column: 11},
				},
			},
			{
				Code: `
        test('should not pass', () => {
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceededMaxAssertion", Line: 8, Column: 11},
				},
			},
			{
				Code: `
        it('should not pass', () => {
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceededMaxAssertion", Line: 8, Column: 11},
				},
			},
			{
				Code: `
        it('should not pass', async () => {
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceededMaxAssertion", Line: 8, Column: 11},
				},
			},
			{
				Code: `
        test('should not pass', () => {
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
        });
        test('should not pass', () => {
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceededMaxAssertion", Line: 8, Column: 11},
					{MessageId: "exceededMaxAssertion", Line: 16, Column: 11},
				},
			},
			{
				Code: `
        test('should not pass', () => {
          const checkValue = (value) => {
            expect(value).toBeDefined();
            expect(value).toBeDefined();
          };

          checkValue(true);
        });
        test('should not pass', () => {
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
        });
      `,
				Options: []interface{}{map[string]interface{}{"max": 1}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceededMaxAssertion", Line: 5, Column: 13},
					{MessageId: "exceededMaxAssertion", Line: 12, Column: 11},
					{MessageId: "exceededMaxAssertion", Line: 13, Column: 11},
				},
			},
			{
				Code: `
        test('should not pass', () => {
          const checkValue = (value) => {
            expect(value).toBeDefined();
            expect(value).toBeDefined();
          };

          checkValue(true);
        });
        test('should not pass', () => {
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
        });
      `,
				Options: []interface{}{map[string]interface{}{"max": 2}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceededMaxAssertion", Line: 13, Column: 11},
				},
			},
			{
				Code: `
        test('should not pass', () => {
          const checkValue = (value) => {
            expect(value).toBeDefined();
            expect(value).toBeDefined();
          };

          expect(value).toBeDefined();
          checkValue(true);
        });
        test('should not pass', () => {
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
        });
      `,
				Options: []interface{}{map[string]interface{}{"max": 2}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceededMaxAssertion", Line: 14, Column: 11},
				},
			},
			{
				Code: `
        test('should not pass', done => {
          emitter.on('event', value => {
            expect(value).toBeDefined();
            expect(value).toBeDefined();

            done();
          });
        });
        test('should not pass', () => {
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
        });
      `,
				Options: []interface{}{map[string]interface{}{"max": 1}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceededMaxAssertion", Line: 5, Column: 13},
					{MessageId: "exceededMaxAssertion", Line: 12, Column: 11},
					{MessageId: "exceededMaxAssertion", Line: 13, Column: 11},
				},
			},
			{
				Code: `
        test('should not pass', done => {
          emitter.on('event', value => {
            expect(value).toBeDefined();
            expect(value).toBeDefined();

            done();
          });
        });
        test('should not pass', () => {
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
        });
      `,
				Options: []interface{}{map[string]interface{}{"max": 2}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceededMaxAssertion", Line: 13, Column: 11},
				},
			},
			{
				Code: `
        describe('given decimal places', () => {
          it("test 1", fakeAsync(() => {
            expect(true).toBeTrue();
            expect(true).toBeTrue();
            expect(true).toBeTrue();
          }))

          it("test 2", fakeAsync(() => {
            expect(true).toBeTrue();
            expect(true).toBeTrue();
            expect(true).toBeTrue();
            expect(true).toBeTrue();
            expect(true).toBeTrue();
          }))
        })
      `,
				Options: []interface{}{map[string]interface{}{"max": 3}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceededMaxAssertion", Line: 13, Column: 13},
					{MessageId: "exceededMaxAssertion", Line: 14, Column: 13},
				},
			},
			{
				Code: `
        describe('test', () => {
          test('should not pass', () => {
            expect(true).toBeDefined();
            expect(true).toBeDefined();
            expect(true).toBeDefined();
            expect(true).toBeDefined();
            expect(true).toBeDefined();
            expect(true).toBeDefined();
          });
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceededMaxAssertion", Line: 9, Column: 13},
				},
			},
			{
				Code: `
        test.each(['should', 'not', 'pass'], () => {
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceededMaxAssertion", Line: 8, Column: 11},
				},
			},
			{
				Code: `
        test('should not pass', () => {
          expect(true).toBeDefined();
          expect(true).toBeDefined();
        });
      `,
				Options: []interface{}{
					map[string]interface{}{"max": 1},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceededMaxAssertion", Line: 4, Column: 11},
				},
			},
		},
	)
}
