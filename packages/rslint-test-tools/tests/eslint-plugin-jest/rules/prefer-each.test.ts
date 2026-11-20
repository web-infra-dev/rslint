// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('prefer-each', {
  valid: [
    ...(
      [
        'it("is true", () => { expect(true).toBe(false) });',
        `
      it.each(getNumbers())("only returns numbers that are greater than seven", number => {
        expect(number).toBeGreaterThan(7);
      });
   \u0020`,
        `
      it("returns numbers that are greater than five", function () {
        for (const number of getNumbers()) {
          expect(number).toBeGreaterThan(5);
        }
      });
   \u0020`,
        `
      it("returns things that are less than ten", function () {
        for (const thing in things) {
          expect(thing).toBeLessThan(10);
        }
      });
   \u0020`,
        `
      it("only returns numbers that are greater than seven", function () {
        const numbers = getNumbers();

        for (let i = 0; i < numbers.length; i++) {
          expect(numbers[i]).toBeGreaterThan(7);
        }
      });
   \u0020`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    ...(
      [
        'it("is true", () => { expect(true).toBe(false) });',
        `it.each(getNumbers())("only returns numbers that are greater than seven", number => {
  expect(number).toBeGreaterThan(7);
});`,
        `it("returns numbers that are greater than five", function () {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(5);
  }
});`,
        `it("returns things that are less than ten", function () {
  for (const thing in things) {
    expect(thing).toBeLessThan(10);
  }
});`,
        `it("only returns numbers that are greater than seven", function () {
  const numbers = getNumbers();

  for (let i = 0; i < numbers.length; i++) {
    expect(numbers[i]).toBeGreaterThan(7);
  }
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream prefer-each / valid ${index}`,
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
        for (const [input, expected] of data) {
          it(\`results in \${expected}\`, () => {
            expect(fn(input)).toBe(expected)
          });
        }
     \u0020`,
          'prefer using `it.each` rather than a manual loop',
          6,
        ],
        [
          `
        for (const [input, expected] of data) {
          describe(\`when the input is \${input}\`, () => {
            it(\`results in \${expected}\`, () => {
              expect(fn(input)).toBe(expected)
            });
          });
        }
     \u0020`,
          'prefer using `describe.each` rather than a manual loop',
          8,
        ],
      ] as const
    ).map(([code, message, endLine], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'preferEach',
          message,
          line: 2,
          column: 9,
          endLine,
          endColumn: 10,
        },
      ],
    })),
    {
      name: 'run 0 invalid 2',
      code: `
        for (const [input, expected] of data) {
          describe(\`when the input is \${input}\`, () => {
            it(\`results in \${expected}\`, () => {
              expect(fn(input)).toBe(expected)
            });
          });
        }

        for (const [input, expected] of data) {
          it.skip(\`results in \${expected}\`, () => {
            expect(fn(input)).toBe(expected)
          });
        }
     \u0020`,
      errors: [
        {
          messageId: 'preferEach',
          message: 'prefer using `describe.each` rather than a manual loop',
          line: 2,
          column: 9,
          endLine: 8,
          endColumn: 10,
        },
        {
          messageId: 'preferEach',
          message: 'prefer using `it.each` rather than a manual loop',
          line: 10,
          column: 9,
          endLine: 14,
          endColumn: 10,
        },
      ],
    },
    ...(
      [
        [
          `
        for (const [input, expected] of data) {
          it.skip(\`results in \${expected}\`, () => {
            expect(fn(input)).toBe(expected)
          });
        }
     \u0020`,
          'prefer using `it.each` rather than a manual loop',
          2,
          6,
        ],
        [
          `
        it('is true', () => {
          expect(true).toBe(false);
        });

        for (const [input, expected] of data) {
          it.skip(\`results in \${expected}\`, () => {
            expect(fn(input)).toBe(expected)
          });
        }
     \u0020`,
          'prefer using `it.each` rather than a manual loop',
          6,
          10,
        ],
        [
          `
        for (const [input, expected] of data) {
          it.skip(\`results in \${expected}\`, () => {
            expect(fn(input)).toBe(expected)
          });
        }

        it('is true', () => {
          expect(true).toBe(false);
        });
     \u0020`,
          'prefer using `it.each` rather than a manual loop',
          2,
          6,
        ],
        [
          `
        it('is true', () => {
          expect(true).toBe(false);
        });

        for (const [input, expected] of data) {
          it.skip(\`results in \${expected}\`, () => {
            expect(fn(input)).toBe(expected)
          });
        }

        it('is true', () => {
          expect(true).toBe(false);
        });
     \u0020`,
          'prefer using `it.each` rather than a manual loop',
          6,
          10,
        ],
        [
          `
        for (const [input, expected] of data) {
          it(\`results in \${expected}\`, () => {
            expect(fn(input)).toBe(expected)
          });

          it(\`results in \${expected}\`, () => {
            expect(fn(input)).toBe(expected)
          });
        }
     \u0020`,
          'prefer using `describe.each` rather than a manual loop',
          2,
          10,
        ],
      ] as const
    ).map(([code, message, line, endLine], index): TestCase => ({
      name: `run 0 invalid ${index + 3}`,
      code,
      errors: [
        {
          messageId: 'preferEach',
          message,
          line,
          column: 9,
          endLine,
          endColumn: 10,
        },
      ],
    })),
    ...(
      [
        `
        for (const [input, expected] of data) {
          it(\`results in \${expected}\`, () => {
            expect(fn(input)).toBe(expected)
          });
        }

        for (const [input, expected] of data) {
          it(\`results in \${expected}\`, () => {
            expect(fn(input)).toBe(expected)
          });
        }
     \u0020`,
        `
        for (const [input, expected] of data) {
          it(\`results in \${expected}\`, () => {
            expect(fn(input)).toBe(expected)
          });
        }

        for (const [input, expected] of data) {
          test(\`results in \${expected}\`, () => {
            expect(fn(input)).toBe(expected)
          });
        }
     \u0020`,
      ] as const
    ).map((code, index): TestCase => ({
      name: `run 0 invalid ${index + 8}`,
      code,
      errors: [
        {
          messageId: 'preferEach',
          message: 'prefer using `it.each` rather than a manual loop',
          line: 2,
          column: 9,
          endLine: 6,
          endColumn: 10,
        },
        {
          messageId: 'preferEach',
          message: 'prefer using `it.each` rather than a manual loop',
          line: 8,
          column: 9,
          endLine: 12,
          endColumn: 10,
        },
      ],
    })),
    ...(
      [
        [
          `
        for (const [input, expected] of data) {
          beforeEach(() => setupSomething(input));

          test(\`results in \${expected}\`, () => {
            expect(doSomething()).toBe(expected)
          });
        }
     \u0020`,
          'prefer using `describe.each` rather than a manual loop',
          8,
        ],
        [
          `
        for (const [input, expected] of data) {
          it("only returns numbers that are greater than seven", function () {
            const numbers = getNumbers(input);

            for (let i = 0; i < numbers.length; i++) {
              expect(numbers[i]).toBeGreaterThan(7);
            }
          });
        }
     \u0020`,
          'prefer using `it.each` rather than a manual loop',
          10,
        ],
        [
          `
        for (const [input, expected] of data) {
          beforeEach(() => setupSomething(input));

          it("only returns numbers that are greater than seven", function () {
            const numbers = getNumbers();

            for (let i = 0; i < numbers.length; i++) {
              expect(numbers[i]).toBeGreaterThan(7);
            }
          });
        }
     \u0020`,
          'prefer using `describe.each` rather than a manual loop',
          12,
        ],
      ] as const
    ).map(([code, message, endLine], index): TestCase => ({
      name: `run 0 invalid ${index + 10}`,
      code,
      errors: [
        {
          messageId: 'preferEach',
          message,
          line: 2,
          column: 9,
          endLine,
          endColumn: 10,
        },
      ],
    })),
    ...(
      [
        [
          `for (const [input, expected] of data) {
  it(\`results in \${expected}\`, () => {
    expect(fn(input)).toBe(expected)
  });
}`,
          'prefer using `it.each` rather than a manual loop',
          5,
        ],
        [
          `for (const [input, expected] of data) {
  describe(\`when the input is \${input}\`, () => {
    it(\`results in \${expected}\`, () => {
      expect(fn(input)).toBe(expected)
    });
  });
}`,
          'prefer using `describe.each` rather than a manual loop',
          7,
        ],
      ] as const
    ).map(([code, message, endLine], index): TestCase => ({
      name: `upstream prefer-each / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'preferEach',
          message,
          line: 1,
          column: 1,
          endLine,
          endColumn: 2,
        },
      ],
    })),
    {
      name: 'upstream prefer-each / invalid 2',
      code: `for (const [input, expected] of data) {
  describe(\`when the input is \${input}\`, () => {
    it(\`results in \${expected}\`, () => {
      expect(fn(input)).toBe(expected)
    });
  });
}

for (const [input, expected] of data) {
  it.skip(\`results in \${expected}\`, () => {
    expect(fn(input)).toBe(expected)
  });
}`,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'preferEach',
          message: 'prefer using `describe.each` rather than a manual loop',
          line: 1,
          column: 1,
          endLine: 7,
          endColumn: 2,
        },
        {
          messageId: 'preferEach',
          message: 'prefer using `it.each` rather than a manual loop',
          line: 9,
          column: 1,
          endLine: 13,
          endColumn: 2,
        },
      ],
    },
    ...(
      [
        [
          `for (const [input, expected] of data) {
  it.skip(\`results in \${expected}\`, () => {
    expect(fn(input)).toBe(expected)
  });
}`,
          'prefer using `it.each` rather than a manual loop',
          1,
          5,
        ],
        [
          `it('is true', () => {
  expect(true).toBe(false);
});

for (const [input, expected] of data) {
  it.skip(\`results in \${expected}\`, () => {
    expect(fn(input)).toBe(expected)
  });
}`,
          'prefer using `it.each` rather than a manual loop',
          5,
          9,
        ],
        [
          `for (const [input, expected] of data) {
  it.skip(\`results in \${expected}\`, () => {
    expect(fn(input)).toBe(expected)
  });
}

it('is true', () => {
  expect(true).toBe(false);
});`,
          'prefer using `it.each` rather than a manual loop',
          1,
          5,
        ],
        [
          `it('is true', () => {
  expect(true).toBe(false);
});

for (const [input, expected] of data) {
  it.skip(\`results in \${expected}\`, () => {
    expect(fn(input)).toBe(expected)
  });
}

it('is true', () => {
  expect(true).toBe(false);
});`,
          'prefer using `it.each` rather than a manual loop',
          5,
          9,
        ],
        [
          `for (const [input, expected] of data) {
  it(\`results in \${expected}\`, () => {
    expect(fn(input)).toBe(expected)
  });

  it(\`results in \${expected}\`, () => {
    expect(fn(input)).toBe(expected)
  });
}`,
          'prefer using `describe.each` rather than a manual loop',
          1,
          9,
        ],
      ] as const
    ).map(([code, message, line, endLine], index): TestCase => ({
      name: `upstream prefer-each / invalid ${index + 3}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'preferEach',
          message,
          line,
          column: 1,
          endLine,
          endColumn: 2,
        },
      ],
    })),
    ...(
      [
        `for (const [input, expected] of data) {
  it(\`results in \${expected}\`, () => {
    expect(fn(input)).toBe(expected)
  });
}

for (const [input, expected] of data) {
  it(\`results in \${expected}\`, () => {
    expect(fn(input)).toBe(expected)
  });
}`,
        `for (const [input, expected] of data) {
  it(\`results in \${expected}\`, () => {
    expect(fn(input)).toBe(expected)
  });
}

for (const [input, expected] of data) {
  test(\`results in \${expected}\`, () => {
    expect(fn(input)).toBe(expected)
  });
}`,
      ] as const
    ).map((code, index): TestCase => ({
      name: `upstream prefer-each / invalid ${index + 8}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'preferEach',
          message: 'prefer using `it.each` rather than a manual loop',
          line: 1,
          column: 1,
          endLine: 5,
          endColumn: 2,
        },
        {
          messageId: 'preferEach',
          message: 'prefer using `it.each` rather than a manual loop',
          line: 7,
          column: 1,
          endLine: 11,
          endColumn: 2,
        },
      ],
    })),
    ...(
      [
        [
          `for (const [input, expected] of data) {
  beforeEach(() => setupSomething(input));

  test(\`results in \${expected}\`, () => {
    expect(doSomething()).toBe(expected)
  });
}`,
          'prefer using `describe.each` rather than a manual loop',
          7,
        ],
        [
          `for (const [input, expected] of data) {
  it("only returns numbers that are greater than seven", function () {
    const numbers = getNumbers(input);

    for (let i = 0; i < numbers.length; i++) {
      expect(numbers[i]).toBeGreaterThan(7);
    }
  });
}`,
          'prefer using `it.each` rather than a manual loop',
          9,
        ],
        [
          `for (const [input, expected] of data) {
  beforeEach(() => setupSomething(input));

  it("only returns numbers that are greater than seven", function () {
    const numbers = getNumbers();

    for (let i = 0; i < numbers.length; i++) {
      expect(numbers[i]).toBeGreaterThan(7);
    }
  });
}`,
          'prefer using `describe.each` rather than a manual loop',
          11,
        ],
      ] as const
    ).map(([code, message, endLine], index): TestCase => ({
      name: `upstream prefer-each / invalid ${index + 10}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'preferEach',
          message,
          line: 1,
          column: 1,
          endLine,
          endColumn: 2,
        },
      ],
    })),
  ],
});
