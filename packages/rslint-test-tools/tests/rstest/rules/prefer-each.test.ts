import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-each', {} as never, {
  valid: [
    { code: 'it("is true", () => { expect(true).toBe(false) });' },
    {
      code: `
      it.each(getNumbers())("only returns numbers that are greater than seven", number => {
        expect(number).toBeGreaterThan(7);
      });
    `,
    },
    {
      code: `
      it("returns numbers that are greater than five", function () {
        for (const number of getNumbers()) {
          expect(number).toBeGreaterThan(5);
        }
      });
    `,
    },
    {
      code: `
      it("returns things that are less than ten", function () {
        for (const thing in things) {
          expect(thing).toBeLessThan(10);
        }
      });
    `,
    },
    {
      code: `
      it("only returns numbers that are greater than seven", function () {
        const numbers = getNumbers();

        for (let i = 0; i < numbers.length; i++) {
          expect(numbers[i]).toBeGreaterThan(7);
        }
      });
    `,
    },
  ],
  invalid: [
    {
      code: `
        for (const [input, expected] of data) {
          it(\`results in $\{expected}\`, () => {
            expect(fn(input)).toBe(expected)
          });
        }
      `,
      errors: [
        {
          messageId: 'preferEach',
          message: 'prefer using `it.each` rather than a manual loop',
          line: 2,
          column: 9,
        },
      ],
    },
    {
      code: `
        for (const [input, expected] of data) {
          describe(\`when the input is $\{input}\`, () => {
            it(\`results in $\{expected}\`, () => {
              expect(fn(input)).toBe(expected)
            });
          });
        }
      `,
      errors: [
        {
          messageId: 'preferEach',
          message: 'prefer using `describe.each` rather than a manual loop',
          line: 2,
          column: 9,
        },
      ],
    },
    {
      code: `
        for (const [input, expected] of data) {
          describe(\`when the input is $\{input}\`, () => {
            it(\`results in $\{expected}\`, () => {
              expect(fn(input)).toBe(expected)
            });
          });
        }

        for (const [input, expected] of data) {
          it.skip(\`results in $\{expected}\`, () => {
            expect(fn(input)).toBe(expected)
          });
        }
      `,
      errors: [
        {
          messageId: 'preferEach',
          message: 'prefer using `describe.each` rather than a manual loop',
          line: 2,
          column: 9,
        },
        {
          messageId: 'preferEach',
          message: 'prefer using `it.each` rather than a manual loop',
          line: 9,
          column: 9,
        },
      ],
    },
    {
      code: `
        for (const [input, expected] of data) {
          it.skip(\`results in $\{expected}\`, () => {
            expect(fn(input)).toBe(expected)
          });
        }
      `,
      errors: [
        {
          messageId: 'preferEach',
          message: 'prefer using `it.each` rather than a manual loop',
          line: 2,
          column: 9,
        },
      ],
    },
    {
      code: `
        it('is true', () => {
          expect(true).toBe(false);
        });

        for (const [input, expected] of data) {
          it.skip(\`results in $\{expected}\`, () => {
            expect(fn(input)).toBe(expected)
          });
        }
      `,
      errors: [
        {
          messageId: 'preferEach',
          message: 'prefer using `it.each` rather than a manual loop',
          line: 6,
          column: 9,
        },
      ],
    },
    {
      code: `
        for (const [input, expected] of data) {
          it.skip(\`results in $\{expected}\`, () => {
            expect(fn(input)).toBe(expected)
          });
        }

        it('is true', () => {
          expect(true).toBe(false);
        });
      `,
      errors: [
        {
          messageId: 'preferEach',
          message: 'prefer using `it.each` rather than a manual loop',
          line: 2,
          column: 9,
        },
      ],
    },
    {
      code: `
        it('is true', () => {
          expect(true).toBe(false);
        });

        for (const [input, expected] of data) {
          it.skip(\`results in $\{expected}\`, () => {
            expect(fn(input)).toBe(expected)
          });
        }

        it('is true', () => {
          expect(true).toBe(false);
        });
      `,
      errors: [
        {
          messageId: 'preferEach',
          message: 'prefer using `it.each` rather than a manual loop',
          line: 6,
          column: 9,
        },
      ],
    },
    {
      code: `
        for (const [input, expected] of data) {
          it(\`results in $\{expected}\`, () => {
            expect(fn(input)).toBe(expected)
          });

          it(\`results in $\{expected}\`, () => {
            expect(fn(input)).toBe(expected)
          });
        }
      `,
      errors: [
        {
          messageId: 'preferEach',
          message: 'prefer using `describe.each` rather than a manual loop',
          line: 2,
          column: 9,
        },
      ],
    },
    {
      code: `
        for (const [input, expected] of data) {
          it(\`results in $\{expected}\`, () => {
            expect(fn(input)).toBe(expected)
          });
        }

        for (const [input, expected] of data) {
          it(\`results in $\{expected}\`, () => {
            expect(fn(input)).toBe(expected)
          });
        }
      `,
      errors: [
        {
          messageId: 'preferEach',
          message: 'prefer using `it.each` rather than a manual loop',
          line: 2,
          column: 9,
        },
        {
          messageId: 'preferEach',
          message: 'prefer using `it.each` rather than a manual loop',
          line: 8,
          column: 9,
        },
      ],
    },
    {
      code: `
        for (const [input, expected] of data) {
          it(\`results in $\{expected}\`, () => {
            expect(fn(input)).toBe(expected)
          });
        }

        for (const [input, expected] of data) {
          test(\`results in $\{expected}\`, () => {
            expect(fn(input)).toBe(expected)
          });
        }
      `,
      errors: [
        {
          messageId: 'preferEach',
          message: 'prefer using `it.each` rather than a manual loop',
          line: 2,
          column: 9,
        },
        {
          messageId: 'preferEach',
          message: 'prefer using `test.each` rather than a manual loop',
          line: 8,
          column: 9,
        },
      ],
    },
    {
      code: `
        for (const [input, expected] of data) {
          beforeEach(() => setupSomething(input));

          test(\`results in $\{expected}\`, () => {
            expect(doSomething()).toBe(expected)
          });
        }
      `,
      errors: [
        {
          messageId: 'preferEach',
          message: 'prefer using `describe.each` rather than a manual loop',
          line: 2,
          column: 9,
        },
      ],
    },
    {
      code: `
        for (const [input, expected] of data) {
          it("only returns numbers that are greater than seven", function () {
            const numbers = getNumbers(input);

            for (let i = 0; i < numbers.length; i++) {
              expect(numbers[i]).toBeGreaterThan(7);
            }
          });
        }
      `,
      errors: [
        {
          messageId: 'preferEach',
          message: 'prefer using `it.each` rather than a manual loop',
          line: 2,
          column: 9,
        },
      ],
    },
    {
      code: `
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
      errors: [
        {
          messageId: 'preferEach',
          message: 'prefer using `describe.each` rather than a manual loop',
          line: 2,
          column: 9,
        },
      ],
    },
  ],
});
