package prefer_each_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/prefer_each"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

type preferEachUpstreamError struct {
	message string
	line    int
	endLine int
}

func upstreamLoopInvalid(code string, errs ...preferEachUpstreamError) rule_tester.InvalidTestCase {
	errors := make([]rule_tester.InvalidTestCaseError, 0, len(errs))
	for _, err := range errs {
		errors = append(errors, rule_tester.InvalidTestCaseError{
			MessageId: "preferEach",
			Message:   err.message,
			Line:      err.line,
			Column:    9,
			EndLine:   err.endLine,
			EndColumn: 10,
		})
	}
	return rule_tester.InvalidTestCase{
		Code:   code,
		Errors: errors,
	}
}

func TestPreferEachUpstream(t *testing.T) {
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
		},
		[]rule_tester.InvalidTestCase{
			upstreamLoopInvalid(`
        for (const [input, expected] of data) {
          it(`+"`results in ${expected}`"+`, () => {
            expect(fn(input)).toBe(expected)
          });
        }
      `,
				preferEachUpstreamError{
					message: "prefer using `it.each` rather than a manual loop",
					line:    2,
					endLine: 6,
				},
			),
			upstreamLoopInvalid(`
        for (const [input, expected] of data) {
          describe(`+"`when the input is ${input}`"+`, () => {
            it(`+"`results in ${expected}`"+`, () => {
              expect(fn(input)).toBe(expected)
            });
          });
        }
      `,
				preferEachUpstreamError{
					message: "prefer using `describe.each` rather than a manual loop",
					line:    2,
					endLine: 8,
				},
			),
			upstreamLoopInvalid(`
        for (const [input, expected] of data) {
          describe(`+"`when the input is ${input}`"+`, () => {
            it(`+"`results in ${expected}`"+`, () => {
              expect(fn(input)).toBe(expected)
            });
          });
        }

        for (const [input, expected] of data) {
          it.skip(`+"`results in ${expected}`"+`, () => {
            expect(fn(input)).toBe(expected)
          });
        }
      `,
				preferEachUpstreamError{
					message: "prefer using `describe.each` rather than a manual loop",
					line:    2,
					endLine: 8,
				},
				preferEachUpstreamError{
					message: "prefer using `it.each` rather than a manual loop",
					line:    10,
					endLine: 14,
				},
			),
			upstreamLoopInvalid(`
        for (const [input, expected] of data) {
          it.skip(`+"`results in ${expected}`"+`, () => {
            expect(fn(input)).toBe(expected)
          });
        }
      `,
				preferEachUpstreamError{
					message: "prefer using `it.each` rather than a manual loop",
					line:    2,
					endLine: 6,
				},
			),
			upstreamLoopInvalid(`
        it('is true', () => {
          expect(true).toBe(false);
        });

        for (const [input, expected] of data) {
          it.skip(`+"`results in ${expected}`"+`, () => {
            expect(fn(input)).toBe(expected)
          });
        }
      `,
				preferEachUpstreamError{
					message: "prefer using `it.each` rather than a manual loop",
					line:    6,
					endLine: 10,
				},
			),
			upstreamLoopInvalid(`
        for (const [input, expected] of data) {
          it.skip(`+"`results in ${expected}`"+`, () => {
            expect(fn(input)).toBe(expected)
          });
        }

        it('is true', () => {
          expect(true).toBe(false);
        });
      `,
				preferEachUpstreamError{
					message: "prefer using `it.each` rather than a manual loop",
					line:    2,
					endLine: 6,
				},
			),
			upstreamLoopInvalid(`
        it('is true', () => {
          expect(true).toBe(false);
        });

        for (const [input, expected] of data) {
          it.skip(`+"`results in ${expected}`"+`, () => {
            expect(fn(input)).toBe(expected)
          });
        }

        it('is true', () => {
          expect(true).toBe(false);
        });
      `,
				preferEachUpstreamError{
					message: "prefer using `it.each` rather than a manual loop",
					line:    6,
					endLine: 10,
				},
			),
			upstreamLoopInvalid(`
        for (const [input, expected] of data) {
          it(`+"`results in ${expected}`"+`, () => {
            expect(fn(input)).toBe(expected)
          });

          it(`+"`results in ${expected}`"+`, () => {
            expect(fn(input)).toBe(expected)
          });
        }
      `,
				preferEachUpstreamError{
					message: "prefer using `describe.each` rather than a manual loop",
					line:    2,
					endLine: 10,
				},
			),
			upstreamLoopInvalid(`
        for (const [input, expected] of data) {
          it(`+"`results in ${expected}`"+`, () => {
            expect(fn(input)).toBe(expected)
          });
        }

        for (const [input, expected] of data) {
          it(`+"`results in ${expected}`"+`, () => {
            expect(fn(input)).toBe(expected)
          });
        }
      `,
				preferEachUpstreamError{
					message: "prefer using `it.each` rather than a manual loop",
					line:    2,
					endLine: 6,
				},
				preferEachUpstreamError{
					message: "prefer using `it.each` rather than a manual loop",
					line:    8,
					endLine: 12,
				},
			),
			upstreamLoopInvalid(
				// Rstest adaptation: a single `test(...)` loop should suggest
				// `test.each`, not `it.each`.
				`
        for (const [input, expected] of data) {
          it(`+"`results in ${expected}`"+`, () => {
            expect(fn(input)).toBe(expected)
          });
        }

        for (const [input, expected] of data) {
          test(`+"`results in ${expected}`"+`, () => {
            expect(fn(input)).toBe(expected)
          });
        }
      `,
				preferEachUpstreamError{
					message: "prefer using `it.each` rather than a manual loop",
					line:    2,
					endLine: 6,
				},
				preferEachUpstreamError{
					message: "prefer using `test.each` rather than a manual loop",
					line:    8,
					endLine: 12,
				},
			),
			upstreamLoopInvalid(`
        for (const [input, expected] of data) {
          beforeEach(() => setupSomething(input));

          test(`+"`results in ${expected}`"+`, () => {
            expect(doSomething()).toBe(expected)
          });
        }
      `,
				preferEachUpstreamError{
					message: "prefer using `describe.each` rather than a manual loop",
					line:    2,
					endLine: 8,
				},
			),
			upstreamLoopInvalid(`
        for (const [input, expected] of data) {
          it("only returns numbers that are greater than seven", function () {
            const numbers = getNumbers(input);

            for (let i = 0; i < numbers.length; i++) {
              expect(numbers[i]).toBeGreaterThan(7);
            }
          });
        }
      `,
				preferEachUpstreamError{
					message: "prefer using `it.each` rather than a manual loop",
					line:    2,
					endLine: 10,
				},
			),
			upstreamLoopInvalid(`
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
				preferEachUpstreamError{
					message: "prefer using `describe.each` rather than a manual loop",
					line:    2,
					endLine: 12,
				},
			),
		},
	)
}
