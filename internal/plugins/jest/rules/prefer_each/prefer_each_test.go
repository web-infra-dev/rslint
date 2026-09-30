package prefer_each_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_each"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferEachRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_each.PreferEachRule,
		[]rule_tester.ValidTestCase{
			{Code: `it("is true", () => { expect(true).toBe(false) });`},
			{Code: `
      it.each(getNumbers())("only returns numbers that are greater than seven", number => {
        expect(number).toBeGreaterThan(7);
      });
    `},
			// while these cases could be done with .each, it's reasonable to have more
			// complex cases that would not look good in .each, so we consider this valid
			{Code: `
      it("returns numbers that are greater than five", function () {
        for (const number of getNumbers()) {
          expect(number).toBeGreaterThan(5);
        }
      });
    `},
			{Code: `
      it("returns things that are less than ten", function () {
        for (const thing in things) {
          expect(thing).toBeLessThan(10);
        }
      });
    `},
			{Code: `
      it("only returns numbers that are greater than seven", function () {
        const numbers = getNumbers();

        for (let i = 0; i < numbers.length; i++) {
          expect(numbers[i]).toBeGreaterThan(7);
        }
      });
    `},
			// A loop that registers nothing is business logic, also after a nested
			// registration in the same callback.
			{Code: `
        test("outer", () => {
          test("inner", () => {});
          for (const row of rows) {
            consume(row);
          }
        });
      `},
			// Registrations in for-in/of iterables and classic for control clauses
			// run once, not once per iteration.
			{Code: "for (const row of getRows(it('one', () => {}))) {}"},
			{Code: "for (let i = register(it('once', () => {})); i < 2; i++) {}"},
		},
		[]rule_tester.InvalidTestCase{
			// Each loop is judged from its own frame. The outer loop registers an
			// `it` on every iteration, so the business loop nested inside it does
			// not cancel the outer report (eslint-plugin-jest's flat list would).
			{
				Code: `
        for (const suite of suites) {
          it(` + "`runs ${suite.name}`" + `, () => {
            expect(runSuite(suite)).toBe(true)
          });

          for (const item of suite.items) {
            setupItem(item);
          }
        }
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferEach",
						Message:   "prefer using `it.each` rather than a manual loop",
						Line:      2,
						Column:    9,
						EndLine:   10,
						EndColumn: 10,
					},
				},
			},
			// The same loop with the business loop written first: reported by
			// eslint-plugin-jest too, so both orders now agree.
			{
				Code: `
        for (const suite of suites) {
          for (const item of suite.items) {
            setupItem(item);
          }

          it(` + "`runs ${suite.name}`" + `, () => {});
        }
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferEach",
						Message:   "prefer using `it.each` rather than a manual loop",
						Line:      2,
						Column:    9,
						EndLine:   8,
						EndColumn: 10,
					},
				},
			},
			// Hooks and describes registered by a loop inside a test callback:
			// eslint-plugin-jest skips every loop while its in-test flag is set.
			{
				Code: `
        test("outer", () => {
          for (const row of rows) {
            beforeEach(() => {});
          }
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferEach",
						Message:   "prefer using `describe.each` rather than a manual loop",
						Line:      3,
						Column:    11,
						EndLine:   5,
						EndColumn: 12,
					},
				},
			},
			{
				Code: `
        test("outer", () => {
          for (const row of rows) {
            describe(row.name, () => {});
          }
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferEach",
						Message:   "prefer using `describe.each` rather than a manual loop",
						Line:      3,
						Column:    11,
						EndLine:   5,
						EndColumn: 12,
					},
				},
			},
			// A registering loop inside a test callback is still a registering loop.
			{
				Code: `
        test("outer", () => {
          for (const row of rows) {
            test(row.name, () => {});
          }
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferEach",
						Message:   "prefer using `it.each` rather than a manual loop",
						Line:      3,
						Column:    11,
						EndLine:   5,
						EndColumn: 12,
					},
				},
			},
			// Both loops report: the inner registers `it`, the outer registers `describe`.
			{
				Code: `
        for (const suite of suites) {
          describe(suite.name, () => {
            for (const row of suite.rows) {
              it(row.name, () => {});
            }
          });
        }
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferEach",
						Message:   "prefer using `it.each` rather than a manual loop",
						Line:      4,
						Column:    13,
						EndLine:   6,
						EndColumn: 14,
					},
					{
						MessageId: "preferEach",
						Message:   "prefer using `describe.each` rather than a manual loop",
						Line:      2,
						Column:    9,
						EndLine:   8,
						EndColumn: 10,
					},
				},
			},
			{
				Code: `
        for (const [input, expected] of data) {
          it(` + "`results in ${expected}`" + `, () => {
            expect(fn(input)).toBe(expected)
          });
        }
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferEach",
						Message:   "prefer using `it.each` rather than a manual loop",
					},
				},
			},
			{
				Code: `
        for (const [input, expected] of data) {
          describe(` + "`when the input is ${input}`" + `, () => {
            it(` + "`results in ${expected}`" + `, () => {
              expect(fn(input)).toBe(expected)
            });
          });
        }
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferEach",
						Message:   "prefer using `describe.each` rather than a manual loop",
					},
				},
			},
			{
				Code: `
        test("outer", () => {
          test("inner", () => {
            expect(true).toBe(true);
          });

          for (const [input, expected] of data) {
            it(` + "`results in ${expected}`" + `, () => {
              expect(fn(input)).toBe(expected)
            });
          }
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferEach",
						Message:   "prefer using `it.each` rather than a manual loop",
					},
				},
			},
			{
				Code: `
        for (const [input, expected] of data) {
          describe(` + "`when the input is ${input}`" + `, () => {
            it(` + "`results in ${expected}`" + `, () => {
              expect(fn(input)).toBe(expected)
            });
          });
        }

        for (const [input, expected] of data) {
          it.skip(` + "`results in ${expected}`" + `, () => {
            expect(fn(input)).toBe(expected)
          });
        }
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferEach",
						Message:   "prefer using `describe.each` rather than a manual loop",
					},
					{
						MessageId: "preferEach",
						Message:   "prefer using `it.each` rather than a manual loop",
					},
				},
			},
			{
				Code: `
        for (const [input, expected] of data) {
          it.skip(` + "`results in ${expected}`" + `, () => {
            expect(fn(input)).toBe(expected)
          });
        }
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferEach",
						Message:   "prefer using `it.each` rather than a manual loop",
					},
				},
			},
			{
				Code: `
        it('is true', () => {
          expect(true).toBe(false);
        });

        for (const [input, expected] of data) {
          it.skip(` + "`results in ${expected}`" + `, () => {
            expect(fn(input)).toBe(expected)
          });
        }
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferEach",
						Message:   "prefer using `it.each` rather than a manual loop",
					},
				},
			},
			{
				Code: `
        for (const [input, expected] of data) {
          it.skip(` + "`results in ${expected}`" + `, () => {
            expect(fn(input)).toBe(expected)
          });
        }

        it('is true', () => {
          expect(true).toBe(false);
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferEach",
						Message:   "prefer using `it.each` rather than a manual loop",
					},
				},
			},
			{
				Code: `
        it('is true', () => {
          expect(true).toBe(false);
        });

        for (const [input, expected] of data) {
          it.skip(` + "`results in ${expected}`" + `, () => {
            expect(fn(input)).toBe(expected)
          });
        }

        it('is true', () => {
          expect(true).toBe(false);
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferEach",
						Message:   "prefer using `it.each` rather than a manual loop",
					},
				},
			},
			{
				Code: `
        for (const suite of suites) {
          for (const item of suite.items) {
            it(` + "`runs ${suite.name}/${item.name}`" + `, () => {
              expect(runItem(suite, item)).toBe(true)
            });
          }
        }
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferEach",
						Message:   "prefer using `it.each` rather than a manual loop",
					},
				},
			},
			{
				Code: `
        for (const [input, expected] of data) {
          it(` + "`results in ${expected}`" + `, () => {
            expect(fn(input)).toBe(expected)
          });

          it(` + "`results in ${expected}`" + `, () => {
            expect(fn(input)).toBe(expected)
          });
        }
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferEach",
						Message:   "prefer using `describe.each` rather than a manual loop",
					},
				},
			},
			{
				Code: `
        for (const [input, expected] of data) {
          it(` + "`results in ${expected}`" + `, () => {
            expect(fn(input)).toBe(expected)
          });
        }

        for (const [input, expected] of data) {
          it(` + "`results in ${expected}`" + `, () => {
            expect(fn(input)).toBe(expected)
          });
        }
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferEach",
						Message:   "prefer using `it.each` rather than a manual loop",
					},
					{
						MessageId: "preferEach",
						Message:   "prefer using `it.each` rather than a manual loop",
					},
				},
			},
			{
				Code: `
        for (const [input, expected] of data) {
          it(` + "`results in ${expected}`" + `, () => {
            expect(fn(input)).toBe(expected)
          });
        }

        for (const [input, expected] of data) {
          test(` + "`results in ${expected}`" + `, () => {
            expect(fn(input)).toBe(expected)
          });
        }
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferEach",
						Message:   "prefer using `it.each` rather than a manual loop",
					},
					{
						MessageId: "preferEach",
						Message:   "prefer using `it.each` rather than a manual loop",
					},
				},
			},
			{
				Code: `
        for (const [input, expected] of data) {
          beforeEach(() => setupSomething(input));

          test(` + "`results in ${expected}`" + `, () => {
            expect(doSomething()).toBe(expected)
          });
        }
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferEach",
						Message:   "prefer using `describe.each` rather than a manual loop",
					},
				},
			},
			{
				Code: `
        for (const [input, expected] of data) {
          it("only returns numbers that are greater than seven", function () {
            const numbers = getNumbers(input);

            for (let i = 0; i < numbers.length; i++) {
              expect(numbers[i]).toBeGreaterThan(7);
            }
          });
        }
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferEach",
						Message:   "prefer using `it.each` rather than a manual loop",
					},
				},
			},
			{
				Code: `
        for (const [input, expected] of data) {
          beforeEach(() => setupSomething(input));

          it("only returns numbers that are greater than seven", function () {
            const numbers = getNumbers();

            for (let i = 0; i < numbers.length; i++) {
              expect(numbers[i]).toBeGreaterThan(7);
            }
          });
        }
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferEach",
						Message:   "prefer using `describe.each` rather than a manual loop",
					},
				},
			},
		},
	)
}
