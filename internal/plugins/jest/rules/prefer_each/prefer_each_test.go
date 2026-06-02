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
			{Code: `
        for (const suite of suites) {
          it(` + "`runs ${suite.name}`" + `, () => {
            expect(runSuite(suite)).toBe(true)
          });

          for (const item of suite.items) {
            setupItem(item);
          }
        }
      `},
		},
		[]rule_tester.InvalidTestCase{
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
