// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

// Shared source templates; reporting ranges remain explicit in each matrix.
const assertionSetup = `  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
});`;
const assertionSetup2 = `          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
        });
     \u0020`;
const assertionSetup3 = `
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
      });
   \u0020`;

new RuleTester(fixtureConfig()).run('max-expects', {
  valid: [
    ...(
      [
        "test('should pass')",
        "test('should pass', () => {})",
        "test.skip('should pass', () => {})",
        `
      test('should pass', function () {
        expect(true).toBeDefined();
      });
   \u0020`,
        `
      test('should pass', () => {
        expect(true).toBeDefined();
        expect(true).toBeDefined();` + assertionSetup3,
        `
      test('should pass', () => {
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        // expect(true).toBeDefined();
      });
   \u0020`,
        `
      it('should pass', () => {
        expect(true).toBeDefined();
        expect(true).toBeDefined();` + assertionSetup3,
        `
      test('should pass', async () => {
        expect(true).toBeDefined();
        expect(true).toBeDefined();` + assertionSetup3,
        `
      test('should pass', async () => {
        expect.hasAssertions();

        expect(true).toBeDefined();
        expect(true).toBeDefined();` + assertionSetup3,
        `
      test('should pass', async () => {
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toEqual(expect.any(Boolean));
      });
   \u0020`,
        `
      test('should pass', async () => {
        expect.hasAssertions();

        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toEqual(expect.any(Boolean));
      });
   \u0020`,
        `
      describe('test', () => {
        test('should pass', () => {
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
        });
      });
   \u0020`,
        `
      test.each(['should', 'pass'])('case', () => {
        expect(true).toBeDefined();
        expect(true).toBeDefined();` + assertionSetup3,
        `
      test('should pass', () => {` +
          assertionSetup3 +
          "  test('should pass', () => {" +
          assertionSetup3,
        `
      test('should not pass', () => {
        const checkValue = (value) => {
          expect(value).toBeDefined();
          expect(value).toBeDefined();
        };

        checkValue(true);
      });
      test('should not pass', () => {` + assertionSetup3,
        `
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
        expect(true).toBeDefined();` + assertionSetup3,
        `
      function myHelper() {
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
      };

      test('should pass', () => {` + assertionSetup3,
        `
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
   \u0020`,
        `
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
   \u0020`,
        `
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
   \u0020`,
        `
      test('ignores expects inside object methods', () => {
        const helper = {
          check() {
            expect(true).toBeDefined();
            expect(true).toBeDefined();
            expect(true).toBeDefined();
          },
        };
` + assertionSetup3,
        `
      test('ignores expects inside class members', () => {
        class Helper {
          constructor() {
            expect(true).toBeDefined();
            expect(true).toBeDefined();
          }

          get value() {
            expect(true).toBeDefined();
            return true;
          }

          set value(next) {
            expect(next).toBeDefined();
          }

          check() {
            expect(true).toBeDefined();
            expect(true).toBeDefined();
          }
        }
` + assertionSetup3,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    {
      name: 'run 0 valid 22',
      code:
        `
        test('should pass', () => {
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
` + assertionSetup2,
      options: [
        {
          max: 10,
        },
      ],
    },
    {
      name: 'run 0 valid 23',
      code: `
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
     \u0020`,
      options: [
        {
          max: 5,
        },
      ],
    },
    ...(
      [
        "test('should pass')",
        "test('should pass', () => {})",
        "test.skip('should pass', () => {})",
        `test('should pass', function () {
  expect(true).toBeDefined();
});`,
        `test('should pass', () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
` + assertionSetup,
        `test('should pass', () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  // expect(true).toBeDefined();
});`,
        `it('should pass', () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
` + assertionSetup,
        `test('should pass', async () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
` + assertionSetup,
        `test('should pass', async () => {
  expect.hasAssertions();

  expect(true).toBeDefined();
  expect(true).toBeDefined();
` + assertionSetup,
        `test('should pass', async () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toEqual(expect.any(Boolean));
});`,
        `test('should pass', async () => {
  expect.hasAssertions();

  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toEqual(expect.any(Boolean));
});`,
        `describe('test', () => {
  test('should pass', () => {
    expect(true).toBeDefined();
    expect(true).toBeDefined();
    expect(true).toBeDefined();
    expect(true).toBeDefined();
    expect(true).toBeDefined();
  });
});`,
        `test.each(['should', 'pass'], () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
` + assertionSetup,
        `test('should pass', () => {
` +
          assertionSetup +
          `
test('should pass', () => {
` +
          assertionSetup,
        `test('should not pass', () => {
  const checkValue = (value) => {
    expect(value).toBeDefined();
    expect(value).toBeDefined();
  };

  checkValue(true);
});
test('should not pass', () => {
` + assertionSetup,
        `test('should not pass', done => {
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
` + assertionSetup,
        `function myHelper() {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
};

test('should pass', () => {
` + assertionSetup,
        `function myHelper1() {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
};

test('should pass', () => {
` +
          assertionSetup +
          `

function myHelper2() {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
};`,
        `test('should pass', () => {
` +
          assertionSetup +
          `

function myHelper() {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
};`,
        `const myHelper1 = () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
};

test('should pass', function() {
` +
          assertionSetup +
          `

const myHelper2 = function() {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
};`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream max-expects / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
    })),
    {
      name: 'upstream max-expects / valid 20',
      code:
        `test('should pass', () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
` + assertionSetup,
      options: [
        {
          max: 10,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
    },
    {
      name: 'upstream max-expects / valid 21',
      code: `describe('given decimal places', () => {
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
})`,
      options: [
        {
          max: 5,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
    },
    // Upstream 29.16.0 reports this, but each(table, callback) does not register
    // a test: the returned function is never called. Rslint counts assertions in
    // registered callbacks; each(table)(title, callback) still reports the excess.
    {
      name: 'upstream max-expects / invalid 12',
      code:
        `test.each(['should', 'not', 'pass'], () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
` + assertionSetup,
      languageOptions: {
        ecmaVersion: 2017,
      },
    },
  ],
  invalid: [
    ...(
      [
        `
        test('should not pass', function () {
` + assertionSetup2,
        `
        test('should not pass', () => {
` + assertionSetup2,
        `
        it('should not pass', () => {
` + assertionSetup2,
        `
        it('should not pass', async () => {
` + assertionSetup2,
      ] as const
    ).map((code, index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'exceededMaxAssertion',
          message: 'Too many assertion calls (6) - maximum allowed is 5',
          line: 8,
          column: 11,
          endLine: 8,
          endColumn: 37,
        },
      ],
    })),
    {
      name: 'run 0 invalid 4',
      code:
        `
        test('should not pass', () => {
` +
        assertionSetup2 +
        `  test('should not pass', () => {
` +
        assertionSetup2,
      errors: [
        {
          messageId: 'exceededMaxAssertion',
          message: 'Too many assertion calls (6) - maximum allowed is 5',
          line: 8,
          column: 11,
          endLine: 8,
          endColumn: 37,
        },
        {
          messageId: 'exceededMaxAssertion',
          message: 'Too many assertion calls (6) - maximum allowed is 5',
          line: 16,
          column: 11,
          endLine: 16,
          endColumn: 37,
        },
      ],
    },
    {
      name: 'run 0 invalid 5',
      code: `
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
     \u0020`,
      options: [
        {
          max: 1,
        },
      ],
      errors: [
        {
          messageId: 'exceededMaxAssertion',
          message: 'Too many assertion calls (2) - maximum allowed is 1',
          line: 5,
          column: 13,
          endLine: 5,
          endColumn: 40,
        },
        {
          messageId: 'exceededMaxAssertion',
          message: 'Too many assertion calls (2) - maximum allowed is 1',
          line: 12,
          column: 11,
          endLine: 12,
          endColumn: 37,
        },
        {
          messageId: 'exceededMaxAssertion',
          message: 'Too many assertion calls (3) - maximum allowed is 1',
          line: 13,
          column: 11,
          endLine: 13,
          endColumn: 37,
        },
      ],
    },
    ...(
      [
        [
          `
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
     \u0020`,
          13,
        ],
        [
          `
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
     \u0020`,
          14,
        ],
      ] as const
    ).map(([code, line], index): TestCase => ({
      name: `run 0 invalid ${index + 6}`,
      code,
      options: [
        {
          max: 2,
        },
      ],
      errors: [
        {
          messageId: 'exceededMaxAssertion',
          message: 'Too many assertion calls (3) - maximum allowed is 2',
          line,
          column: 11,
          endLine: line,
          endColumn: 37,
        },
      ],
    })),
    {
      name: 'run 0 invalid 8',
      code: `
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
     \u0020`,
      options: [
        {
          max: 1,
        },
      ],
      errors: [
        {
          messageId: 'exceededMaxAssertion',
          message: 'Too many assertion calls (2) - maximum allowed is 1',
          line: 5,
          column: 13,
          endLine: 5,
          endColumn: 40,
        },
        {
          messageId: 'exceededMaxAssertion',
          message: 'Too many assertion calls (2) - maximum allowed is 1',
          line: 12,
          column: 11,
          endLine: 12,
          endColumn: 37,
        },
        {
          messageId: 'exceededMaxAssertion',
          message: 'Too many assertion calls (3) - maximum allowed is 1',
          line: 13,
          column: 11,
          endLine: 13,
          endColumn: 37,
        },
      ],
    },
    {
      name: 'run 0 invalid 9',
      code: `
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
     \u0020`,
      options: [
        {
          max: 2,
        },
      ],
      errors: [
        {
          messageId: 'exceededMaxAssertion',
          message: 'Too many assertion calls (3) - maximum allowed is 2',
          line: 13,
          column: 11,
          endLine: 13,
          endColumn: 37,
        },
      ],
    },
    {
      name: 'run 0 invalid 10',
      code: `
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
     \u0020`,
      options: [
        {
          max: 3,
        },
      ],
      errors: [
        {
          messageId: 'exceededMaxAssertion',
          message: 'Too many assertion calls (4) - maximum allowed is 3',
          line: 13,
          column: 13,
          endLine: 13,
          endColumn: 36,
        },
        {
          messageId: 'exceededMaxAssertion',
          message: 'Too many assertion calls (5) - maximum allowed is 3',
          line: 14,
          column: 13,
          endLine: 14,
          endColumn: 36,
        },
      ],
    },
    ...(
      [
        [
          `
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
     \u0020`,
          9,
          13,
        ],
        [
          `
        test.each(['should', 'not', 'pass'])('case', () => {
` + assertionSetup2,
          8,
          11,
        ],
      ] as const
    ).map(([code, line, column], index): TestCase => ({
      name: `run 0 invalid ${index + 11}`,
      code,
      errors: [
        {
          messageId: 'exceededMaxAssertion',
          message: 'Too many assertion calls (6) - maximum allowed is 5',
          line,
          column,
          endLine: line,
          endColumn: column + 26,
        },
      ],
    })),
    {
      name: 'run 0 invalid 13',
      code: `
        test('should not pass', () => {
          expect(true).toBeDefined();
          expect(true).toBeDefined();
        });
     \u0020`,
      options: [
        {
          max: 1,
        },
      ],
      errors: [
        {
          messageId: 'exceededMaxAssertion',
          message: 'Too many assertion calls (2) - maximum allowed is 1',
          line: 4,
          column: 11,
          endLine: 4,
          endColumn: 37,
        },
      ],
    },
    {
      name: 'run 0 invalid 14',
      code: `
        test('x', () => {
          expect(1).toBe(1);
          expect(1).toBe(1);
          expect(1).toBe(1);
          run((() => {
            expect(1).toBe(1);
            expect(1).toBe(1);
            expect(1).toBe(1);
          }));
        });
     \u0020`,
      errors: [
        {
          messageId: 'exceededMaxAssertion',
          message: 'Too many assertion calls (6) - maximum allowed is 5',
          line: 9,
          column: 13,
          endLine: 9,
          endColumn: 30,
        },
      ],
    },
    ...(
      [
        `test('should not pass', function () {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
` + assertionSetup,
        `test('should not pass', () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
` + assertionSetup,
        `it('should not pass', () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
` + assertionSetup,
        `it('should not pass', async () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
` + assertionSetup,
      ] as const
    ).map((code, index): TestCase => ({
      name: `upstream max-expects / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'exceededMaxAssertion',
          message: 'Too many assertion calls (6) - maximum allowed is 5',
          line: 7,
          column: 3,
          endLine: 7,
          endColumn: 29,
        },
      ],
    })),
    {
      name: 'upstream max-expects / invalid 4',
      code:
        `test('should not pass', () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
` +
        assertionSetup +
        `
test('should not pass', () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
` +
        assertionSetup,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'exceededMaxAssertion',
          message: 'Too many assertion calls (6) - maximum allowed is 5',
          line: 7,
          column: 3,
          endLine: 7,
          endColumn: 29,
        },
        {
          messageId: 'exceededMaxAssertion',
          message: 'Too many assertion calls (6) - maximum allowed is 5',
          line: 15,
          column: 3,
          endLine: 15,
          endColumn: 29,
        },
      ],
    },
    {
      name: 'upstream max-expects / invalid 5',
      code:
        `test('should not pass', () => {
  const checkValue = (value) => {
    expect(value).toBeDefined();
    expect(value).toBeDefined();
  };

  checkValue(true);
});
test('should not pass', () => {
` + assertionSetup,
      options: [
        {
          max: 1,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'exceededMaxAssertion',
          message: 'Too many assertion calls (2) - maximum allowed is 1',
          line: 4,
          column: 5,
          endLine: 4,
          endColumn: 32,
        },
        {
          messageId: 'exceededMaxAssertion',
          message: 'Too many assertion calls (2) - maximum allowed is 1',
          line: 11,
          column: 3,
          endLine: 11,
          endColumn: 29,
        },
        {
          messageId: 'exceededMaxAssertion',
          message: 'Too many assertion calls (3) - maximum allowed is 1',
          line: 12,
          column: 3,
          endLine: 12,
          endColumn: 29,
        },
      ],
    },
    ...(
      [
        [
          `test('should not pass', () => {
  const checkValue = (value) => {
    expect(value).toBeDefined();
    expect(value).toBeDefined();
  };

  checkValue(true);
});
test('should not pass', () => {
` + assertionSetup,
          12,
        ],
        [
          `test('should not pass', () => {
  const checkValue = (value) => {
    expect(value).toBeDefined();
    expect(value).toBeDefined();
  };

  expect(value).toBeDefined();
  checkValue(true);
});
test('should not pass', () => {
` + assertionSetup,
          13,
        ],
      ] as const
    ).map(([code, line], index): TestCase => ({
      name: `upstream max-expects / invalid ${index + 6}`,
      code,
      options: [
        {
          max: 2,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'exceededMaxAssertion',
          message: 'Too many assertion calls (3) - maximum allowed is 2',
          line,
          column: 3,
          endLine: line,
          endColumn: 29,
        },
      ],
    })),
    {
      name: 'upstream max-expects / invalid 8',
      code:
        `test('should not pass', done => {
  emitter.on('event', value => {
    expect(value).toBeDefined();
    expect(value).toBeDefined();

    done();
  });
});
test('should not pass', () => {
` + assertionSetup,
      options: [
        {
          max: 1,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'exceededMaxAssertion',
          message: 'Too many assertion calls (2) - maximum allowed is 1',
          line: 4,
          column: 5,
          endLine: 4,
          endColumn: 32,
        },
        {
          messageId: 'exceededMaxAssertion',
          message: 'Too many assertion calls (2) - maximum allowed is 1',
          line: 11,
          column: 3,
          endLine: 11,
          endColumn: 29,
        },
        {
          messageId: 'exceededMaxAssertion',
          message: 'Too many assertion calls (3) - maximum allowed is 1',
          line: 12,
          column: 3,
          endLine: 12,
          endColumn: 29,
        },
      ],
    },
    {
      name: 'upstream max-expects / invalid 9',
      code:
        `test('should not pass', done => {
  emitter.on('event', value => {
    expect(value).toBeDefined();
    expect(value).toBeDefined();

    done();
  });
});
test('should not pass', () => {
` + assertionSetup,
      options: [
        {
          max: 2,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'exceededMaxAssertion',
          message: 'Too many assertion calls (3) - maximum allowed is 2',
          line: 12,
          column: 3,
          endLine: 12,
          endColumn: 29,
        },
      ],
    },
    {
      name: 'upstream max-expects / invalid 10',
      code: `describe('given decimal places', () => {
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
})`,
      options: [
        {
          max: 3,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'exceededMaxAssertion',
          message: 'Too many assertion calls (4) - maximum allowed is 3',
          line: 12,
          column: 5,
          endLine: 12,
          endColumn: 28,
        },
        {
          messageId: 'exceededMaxAssertion',
          message: 'Too many assertion calls (5) - maximum allowed is 3',
          line: 13,
          column: 5,
          endLine: 13,
          endColumn: 28,
        },
      ],
    },
    {
      name: 'upstream max-expects / invalid 11',
      code: `describe('test', () => {
  test('should not pass', () => {
    expect(true).toBeDefined();
    expect(true).toBeDefined();
    expect(true).toBeDefined();
    expect(true).toBeDefined();
    expect(true).toBeDefined();
    expect(true).toBeDefined();
  });
});`,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'exceededMaxAssertion',
          message: 'Too many assertion calls (6) - maximum allowed is 5',
          line: 8,
          column: 5,
          endLine: 8,
          endColumn: 31,
        },
      ],
    },
    {
      name: 'upstream max-expects / invalid 13',
      code: `test('should not pass', () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
});`,
      options: [
        {
          max: 1,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'exceededMaxAssertion',
          message: 'Too many assertion calls (2) - maximum allowed is 1',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 29,
        },
      ],
    },
  ],
});
