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
const assertionSetup = `          switch(something) {
            case 'value':
              expect(something).toHaveBeenCalled();
            default:
              break;
          }
        })
     \u0020`;
const assertionSetup2 = `          try {

          } catch (err) {
            expect(err).toMatch('Error');
          }
        })
     \u0020`;

new RuleTester(fixtureConfig()).run('no-conditional-expect', {
  valid: [
    ...(
      [
        `
      it('foo', () => {
        expect(1).toBe(2);
      });
   \u0020`,
        `
      it('foo', () => {
        expect(!true).toBe(false);
      });
   \u0020`,
        `
      it('foo', () => {
        process.env.FAIL && setNum(1);

        expect(num).toBe(2);
      });
   \u0020`,
        `
      function getValue() {
        let num = 2;

        process.env.FAIL && setNum(1);

        return num;
      }

      it('foo', () => {
        expect(getValue()).toBe(2);
      });
   \u0020`,
        `
      it('foo', () => {
        process.env.FAIL || setNum(1);

        expect(num).toBe(2);
      });
   \u0020`,
        `
      function getValue() {
        let num = 2;

        process.env.FAIL || setNum(1);

        return num;
      }

      it('foo', () => {
        expect(getValue()).toBe(2);
      });
   \u0020`,
        `
      it('foo', () => {
        const num = process.env.FAIL ? 1 : 2;

        expect(num).toBe(2);
      });
   \u0020`,
        `
      function getValue() {
        return process.env.FAIL ? 1 : 2
      }

      it('foo', () => {
        expect(getValue()).toBe(2);
      });
   \u0020`,
        `
      it('foo', () => {
        let num;

        switch(process.env.FAIL) {
          case true:
            num = 1;
            break;
          case false:
            num = 2;
            break;
        }

        expect(num).toBe(2);
      });
   \u0020`,
        `
      function getValue() {
        switch(process.env.FAIL) {
          case true:
            return 1;
          case false:
            return 2;
        }
      }

      it('foo', () => {
        expect(getValue()).toBe(2);
      });
   \u0020`,
        `
      it('foo', () => {
        let num = 2;

        if(process.env.FAIL) {
          num = 1;
        }

        expect(num).toBe(2);
      });
   \u0020`,
        `
      function getValue() {
        if(process.env.FAIL) {
          return 1;
        }

        return 2;
      }

      it('foo', () => {
        expect(getValue()).toBe(2);
      });
   \u0020`,
        `
      it('foo', () => {
        try {
          // do something
        } catch {
          // ignore errors
        } finally {
          expect(something).toHaveBeenCalled();
        }
      });
   \u0020`,
        `
      it('foo', () => {
        try {
          // do something
        } catch {
          // ignore errors
        }

        expect(something).toHaveBeenCalled();
      });
   \u0020`,
        `
      function getValue() {
        try {
          // do something
        } catch {
          // ignore errors
        } finally {
          expect(something).toHaveBeenCalled();
        }
      }

      it('foo', getValue);
   \u0020`,
        `
      function getValue() {
        try {
          process.env.FAIL.toString();

          return 1;
        } catch {
          return 2;
        }
      }

      it('foo', () => {
        expect(getValue()).toBe(2);
      });
   \u0020`,
        `
      it('works', async () => {
        try {
          await Promise.resolve().then(() => {
            throw new Error('oh noes!');
          });
        } catch {
          // ignore errors
        } finally {
          expect(something).toHaveBeenCalled();
        }
      });
   \u0020`,
        `
      it('works', async () => {
        await doSomething().catch(error => error);

        expect(error).toBeInstanceOf(Error);
      });
   \u0020`,
        `
      it('works', async () => {
        try {
          await Promise.resolve().then(() => {
            throw new Error('oh noes!');
          });
        } catch {
          // ignore errors
        }

        expect(something).toHaveBeenCalled();
      });
   \u0020`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    ...(
      [
        `
      it('foo', () => {
        expect(1).toBe(2);
      });
   \u0020`,
        `
      it('foo', () => {
        expect(!true).toBe(false);
      });
   \u0020`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream common tests / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2019,
      },
    })),
    ...(
      [
        `
      it('foo', () => {
        process.env.FAIL && setNum(1);

        expect(num).toBe(2);
      });
   \u0020`,
        `
      function getValue() {
        let num = 2;

        process.env.FAIL && setNum(1);

        return num;
      }

      it('foo', () => {
        expect(getValue()).toBe(2);
      });
   \u0020`,
        `
      it('foo', () => {
        process.env.FAIL || setNum(1);

        expect(num).toBe(2);
      });
   \u0020`,
        `
      function getValue() {
        let num = 2;

        process.env.FAIL || setNum(1);

        return num;
      }

      it('foo', () => {
        expect(getValue()).toBe(2);
      });
   \u0020`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream logical conditions / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2019,
      },
    })),
    ...(
      [
        `
      it('foo', () => {
        const num = process.env.FAIL ? 1 : 2;

        expect(num).toBe(2);
      });
   \u0020`,
        `
      function getValue() {
        return process.env.FAIL ? 1 : 2
      }

      it('foo', () => {
        expect(getValue()).toBe(2);
      });
   \u0020`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream conditional conditions / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2019,
      },
    })),
    ...(
      [
        `
      it('foo', () => {
        let num;

        switch(process.env.FAIL) {
          case true:
            num = 1;
            break;
          case false:
            num = 2;
            break;
        }

        expect(num).toBe(2);
      });
   \u0020`,
        `
      function getValue() {
        switch(process.env.FAIL) {
          case true:
            return 1;
          case false:
            return 2;
        }
      }

      it('foo', () => {
        expect(getValue()).toBe(2);
      });
   \u0020`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream switch conditions / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2019,
      },
    })),
    ...(
      [
        `
      it('foo', () => {
        let num = 2;

        if(process.env.FAIL) {
          num = 1;
        }

        expect(num).toBe(2);
      });
   \u0020`,
        `
      function getValue() {
        if(process.env.FAIL) {
          return 1;
        }

        return 2;
      }

      it('foo', () => {
        expect(getValue()).toBe(2);
      });
   \u0020`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream if conditions / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2019,
      },
    })),
    ...(
      [
        `
      it('foo', () => {
        try {
          // do something
        } catch {
          // ignore errors
        } finally {
          expect(something).toHaveBeenCalled();
        }
      });
   \u0020`,
        `
      it('foo', () => {
        try {
          // do something
        } catch {
          // ignore errors
        }

        expect(something).toHaveBeenCalled();
      });
   \u0020`,
        `
      function getValue() {
        try {
          // do something
        } catch {
          // ignore errors
        } finally {
          expect(something).toHaveBeenCalled();
        }
      }

      it('foo', getValue);
   \u0020`,
        `
      function getValue() {
        try {
          process.env.FAIL.toString();

          return 1;
        } catch {
          return 2;
        }
      }

      it('foo', () => {
        expect(getValue()).toBe(2);
      });
   \u0020`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream catch conditions / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2019,
      },
    })),
    ...(
      [
        `
      it('works', async () => {
        try {
          await Promise.resolve().then(() => {
            throw new Error('oh noes!');
          });
        } catch {
          // ignore errors
        } finally {
          expect(something).toHaveBeenCalled();
        }
      });
   \u0020`,
        `
      it('works', async () => {
        await doSomething().catch(error => error);

        expect(error).toBeInstanceOf(Error);
      });
   \u0020`,
        `
      it('works', async () => {
        try {
          await Promise.resolve().then(() => {
            throw new Error('oh noes!');
          });
        } catch {
          // ignore errors
        }

        expect(something).toHaveBeenCalled();
      });
   \u0020`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream promises / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2019,
      },
    })),
  ],
  invalid: [
    ...(
      [
        [
          `
        it('foo', () => {
          something && expect(something).toHaveBeenCalled();
        })
     \u0020`,
          3,
          24,
          60,
        ],
        [
          `
        it('foo', () => {
          a || b && expect(something).toHaveBeenCalled();
        })
     \u0020`,
          3,
          21,
          57,
        ],
        [
          `
        it('foo', () => {
          (a || b) && expect(something).toHaveBeenCalled();
        })
     \u0020`,
          3,
          23,
          59,
        ],
        [
          `
        it('foo', () => {
          a || (b && expect(something).toHaveBeenCalled());
        })
     \u0020`,
          3,
          22,
          58,
        ],
        [
          `
        it('foo', () => {
          a && b && expect(something).toHaveBeenCalled();
        })
     \u0020`,
          3,
          21,
          57,
        ],
        [
          `
        it('foo', () => {
          a && b || expect(something).toHaveBeenCalled();
        })
     \u0020`,
          3,
          21,
          57,
        ],
        [
          `
        it('foo', () => {
          (a && b) || expect(something).toHaveBeenCalled();
        })
     \u0020`,
          3,
          23,
          59,
        ],
        [
          `
        function getValue() {
          something && expect(something).toHaveBeenCalled();
        }

        it('foo', getValue);
     \u0020`,
          3,
          24,
          60,
        ],
        [
          `
        it('foo', () => {
          something || expect(something).toHaveBeenCalled();
        })
     \u0020`,
          3,
          24,
          60,
        ],
        [
          `
        it.each\`\`('foo', () => {
          something || expect(something).toHaveBeenCalled();
        })
     \u0020`,
          3,
          24,
          60,
        ],
        [
          `
        it.each()('foo', () => {
          something || expect(something).toHaveBeenCalled();
        })
     \u0020`,
          3,
          24,
          60,
        ],
        [
          `
        function getValue() {
          something || expect(something).toHaveBeenCalled();
        }

        it('foo', getValue);
     \u0020`,
          3,
          24,
          60,
        ],
        [
          `
        it('foo', () => {
          something ? expect(something).toHaveBeenCalled() : noop();
        })
     \u0020`,
          3,
          23,
          59,
        ],
        [
          `
        function getValue() {
          something ? expect(something).toHaveBeenCalled() : noop();
        }

        it('foo', getValue);
     \u0020`,
          3,
          23,
          59,
        ],
        [
          `
        it('foo', () => {
          something ? noop() : expect(something).toHaveBeenCalled();
        })
     \u0020`,
          3,
          32,
          68,
        ],
        [
          `
        it.each\`\`('foo', () => {
          something ? noop() : expect(something).toHaveBeenCalled();
        })
     \u0020`,
          3,
          32,
          68,
        ],
        [
          `
        it.each()('foo', () => {
          something ? noop() : expect(something).toHaveBeenCalled();
        })
     \u0020`,
          3,
          32,
          68,
        ],
        [
          `
        function getValue() {
          something ? noop() : expect(something).toHaveBeenCalled();
        }

        it('foo', getValue);
     \u0020`,
          3,
          32,
          68,
        ],
        [
          `
        it('foo', () => {
          switch(something) {
            case 'value':
              break;
            default:
              expect(something).toHaveBeenCalled();
          }
        })
     \u0020`,
          7,
          15,
          51,
        ],
        [
          `
        it('foo', () => {
` + assertionSetup,
          5,
          15,
          51,
        ],
        [
          `
        it.each\`\`('foo', () => {
` + assertionSetup,
          5,
          15,
          51,
        ],
        [
          `
        it.each()('foo', () => {
` + assertionSetup,
          5,
          15,
          51,
        ],
        [
          `
        function getValue() {
          switch(something) {
            case 'value':
              break;
            default:
              expect(something).toHaveBeenCalled();
          }
        }

        it('foo', getValue);
     \u0020`,
          7,
          15,
          51,
        ],
        [
          `
        it('foo', () => {
          if(doSomething) {
            expect(something).toHaveBeenCalled();
          }
        })
     \u0020`,
          4,
          13,
          49,
        ],
        [
          `
        it('foo', () => {
          if(!doSomething) {
            // do nothing
          } else {
            expect(something).toHaveBeenCalled();
          }
        })
     \u0020`,
          6,
          13,
          49,
        ],
        [
          `
        it.each\`\`('foo', () => {
          if(!doSomething) {
            // do nothing
          } else {
            expect(something).toHaveBeenCalled();
          }
        })
     \u0020`,
          6,
          13,
          49,
        ],
        [
          `
        it.each()('foo', () => {
          if(!doSomething) {
            // do nothing
          } else {
            expect(something).toHaveBeenCalled();
          }
        })
     \u0020`,
          6,
          13,
          49,
        ],
        [
          `
        function getValue() {
          if(doSomething) {
            expect(something).toHaveBeenCalled();
          }
        }

        it('foo', getValue);
     \u0020`,
          4,
          13,
          49,
        ],
        [
          `
        function getValue() {
          if(!doSomething) {
            // do nothing
          } else {
            expect(something).toHaveBeenCalled();
          }
        }

        it('foo', getValue);
     \u0020`,
          6,
          13,
          49,
        ],
        [
          `
        it('foo', () => {
` + assertionSetup2,
          6,
          13,
          41,
        ],
        [
          `
        it.each\`\`('foo', () => {
` + assertionSetup2,
          6,
          13,
          41,
        ],
        [
          `
        it.each()('foo', () => {
` + assertionSetup2,
          6,
          13,
          41,
        ],
        [
          `
        it.skip.each\`\`('foo', () => {
` + assertionSetup2,
          6,
          13,
          41,
        ],
        [
          `
        it.skip.each()('foo', () => {
` + assertionSetup2,
          6,
          13,
          41,
        ],
        [
          `
        function getValue() {
          try {
            // do something
          } catch {
            expect(something).toHaveBeenCalled();
          }
        }

        it('foo', getValue);
     \u0020`,
          6,
          13,
          49,
        ],
        [
          `
        it('works', async () => {
          await Promise.resolve()
            .then(() => { throw new Error('oh noes!'); })
            .catch(error => expect(error).toBeInstanceOf(Error));
        });
     \u0020`,
          5,
          29,
          64,
        ],
      ] as const
    ).map(([code, line, column, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'conditionalExpect',
          message: 'Avoid calling `expect` conditionally',
          line,
          column,
          endLine: line,
          endColumn,
        },
      ],
    })),
    ...(
      [
        [
          `
        it('works', async () => {
          await Promise.resolve()
            .then(() => { throw new Error('oh noes!'); })
            .catch(error => expect(error).toBeInstanceOf(Error))
            .then(() => { throw new Error('oh noes!'); })
            .catch(error => expect(error).toBeInstanceOf(Error))
            .then(() => { throw new Error('oh noes!'); })
            .catch(error => expect(error).toBeInstanceOf(Error));
        });
     \u0020`,
          5,
          7,
          9,
        ],
        [
          `
        it('works', async () => {
          await Promise.resolve()
            .catch(error => expect(error).toBeInstanceOf(Error))
            .catch(error => expect(error).toBeInstanceOf(Error))
            .catch(error => expect(error).toBeInstanceOf(Error));
        });
     \u0020`,
          4,
          5,
          6,
        ],
      ] as const
    ).map(([code, line1, line2, line3], index): TestCase => ({
      name: `run 0 invalid ${index + 36}`,
      code,
      errors: [
        {
          messageId: 'conditionalExpect',
          message: 'Avoid calling `expect` conditionally',
          line: line1,
          column: 29,
          endLine: line1,
          endColumn: 64,
        },
        {
          messageId: 'conditionalExpect',
          message: 'Avoid calling `expect` conditionally',
          line: line2,
          column: 29,
          endLine: line2,
          endColumn: 64,
        },
        {
          messageId: 'conditionalExpect',
          message: 'Avoid calling `expect` conditionally',
          line: line3,
          column: 29,
          endLine: line3,
          endColumn: 64,
        },
      ],
    })),
    ...(
      [
        [
          `
        it('works', async () => {
          await Promise.resolve()
            .catch(error => expect(error).toBeInstanceOf(Error))
            .then(() => { throw new Error('oh noes!'); })
            .then(() => { throw new Error('oh noes!'); })
            .then(() => { throw new Error('oh noes!'); });
        });
     \u0020`,
          4,
          29,
        ],
        [
          `
        it('works', async () => {
          await somePromise
            .then(() => { throw new Error('oh noes!'); })
            .catch(error => expect(error).toBeInstanceOf(Error));
        });
     \u0020`,
          5,
          29,
        ],
        [
          `
        it('works', async () => {
          await somePromise.catch(error => expect(error).toBeInstanceOf(Error));
        });
     \u0020`,
          3,
          44,
        ],
      ] as const
    ).map(([code, line, column], index): TestCase => ({
      name: `run 0 invalid ${index + 38}`,
      code,
      errors: [
        {
          messageId: 'conditionalExpect',
          message: 'Avoid calling `expect` conditionally',
          line,
          column,
          endLine: line,
          endColumn: column + 35,
        },
      ],
    })),
    ...(
      [
        [
          `
        it('foo', () => {
          something && expect(something).toHaveBeenCalled();
        })
     \u0020`,
          24,
        ],
        [
          `
        it('foo', () => {
          a || b && expect(something).toHaveBeenCalled();
        })
     \u0020`,
          21,
        ],
        [
          `
        it('foo', () => {
          (a || b) && expect(something).toHaveBeenCalled();
        })
     \u0020`,
          23,
        ],
        [
          `
        it('foo', () => {
          a || (b && expect(something).toHaveBeenCalled());
        })
     \u0020`,
          22,
        ],
        [
          `
        it('foo', () => {
          a && b && expect(something).toHaveBeenCalled();
        })
     \u0020`,
          21,
        ],
        [
          `
        it('foo', () => {
          a && b || expect(something).toHaveBeenCalled();
        })
     \u0020`,
          21,
        ],
        [
          `
        it('foo', () => {
          (a && b) || expect(something).toHaveBeenCalled();
        })
     \u0020`,
          23,
        ],
        [
          `
        function getValue() {
          something && expect(something).toHaveBeenCalled();
        }

        it('foo', getValue);
     \u0020`,
          24,
        ],
        [
          `
        it('foo', () => {
          something || expect(something).toHaveBeenCalled();
        })
     \u0020`,
          24,
        ],
        [
          `
        it.each\`\`('foo', () => {
          something || expect(something).toHaveBeenCalled();
        })
     \u0020`,
          24,
        ],
        [
          `
        it.each()('foo', () => {
          something || expect(something).toHaveBeenCalled();
        })
     \u0020`,
          24,
        ],
        [
          `
        function getValue() {
          something || expect(something).toHaveBeenCalled();
        }

        it('foo', getValue);
     \u0020`,
          24,
        ],
      ] as const
    ).map(([code, column], index): TestCase => ({
      name: `upstream logical conditions / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2019,
      },
      errors: [
        {
          messageId: 'conditionalExpect',
          message: 'Avoid calling `expect` conditionally',
          line: 3,
          column,
          endLine: 3,
          endColumn: column + 36,
        },
      ],
    })),
    ...(
      [
        [
          `
        it('foo', () => {
          something ? expect(something).toHaveBeenCalled() : noop();
        })
     \u0020`,
          23,
        ],
        [
          `
        function getValue() {
          something ? expect(something).toHaveBeenCalled() : noop();
        }

        it('foo', getValue);
     \u0020`,
          23,
        ],
        [
          `
        it('foo', () => {
          something ? noop() : expect(something).toHaveBeenCalled();
        })
     \u0020`,
          32,
        ],
        [
          `
        it.each\`\`('foo', () => {
          something ? noop() : expect(something).toHaveBeenCalled();
        })
     \u0020`,
          32,
        ],
        [
          `
        it.each()('foo', () => {
          something ? noop() : expect(something).toHaveBeenCalled();
        })
     \u0020`,
          32,
        ],
        [
          `
        function getValue() {
          something ? noop() : expect(something).toHaveBeenCalled();
        }

        it('foo', getValue);
     \u0020`,
          32,
        ],
      ] as const
    ).map(([code, column], index): TestCase => ({
      name: `upstream conditional conditions / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2019,
      },
      errors: [
        {
          messageId: 'conditionalExpect',
          message: 'Avoid calling `expect` conditionally',
          line: 3,
          column,
          endLine: 3,
          endColumn: column + 36,
        },
      ],
    })),
    ...(
      [
        [
          `
        it('foo', () => {
          switch(something) {
            case 'value':
              break;
            default:
              expect(something).toHaveBeenCalled();
          }
        })
     \u0020`,
          7,
        ],
        [
          `
        it('foo', () => {
` + assertionSetup,
          5,
        ],
        [
          `
        it.each\`\`('foo', () => {
` + assertionSetup,
          5,
        ],
        [
          `
        it.each()('foo', () => {
` + assertionSetup,
          5,
        ],
        [
          `
        function getValue() {
          switch(something) {
            case 'value':
              break;
            default:
              expect(something).toHaveBeenCalled();
          }
        }

        it('foo', getValue);
     \u0020`,
          7,
        ],
      ] as const
    ).map(([code, line], index): TestCase => ({
      name: `upstream switch conditions / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2019,
      },
      errors: [
        {
          messageId: 'conditionalExpect',
          message: 'Avoid calling `expect` conditionally',
          line,
          column: 15,
          endLine: line,
          endColumn: 51,
        },
      ],
    })),
    ...(
      [
        [
          `
        it('foo', () => {
          if(doSomething) {
            expect(something).toHaveBeenCalled();
          }
        })
     \u0020`,
          4,
        ],
        [
          `
        it('foo', () => {
          if(!doSomething) {
            // do nothing
          } else {
            expect(something).toHaveBeenCalled();
          }
        })
     \u0020`,
          6,
        ],
        [
          `
        it.each\`\`('foo', () => {
          if(!doSomething) {
            // do nothing
          } else {
            expect(something).toHaveBeenCalled();
          }
        })
     \u0020`,
          6,
        ],
        [
          `
        it.each()('foo', () => {
          if(!doSomething) {
            // do nothing
          } else {
            expect(something).toHaveBeenCalled();
          }
        })
     \u0020`,
          6,
        ],
        [
          `
        function getValue() {
          if(doSomething) {
            expect(something).toHaveBeenCalled();
          }
        }

        it('foo', getValue);
     \u0020`,
          4,
        ],
        [
          `
        function getValue() {
          if(!doSomething) {
            // do nothing
          } else {
            expect(something).toHaveBeenCalled();
          }
        }

        it('foo', getValue);
     \u0020`,
          6,
        ],
      ] as const
    ).map(([code, line], index): TestCase => ({
      name: `upstream if conditions / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2019,
      },
      errors: [
        {
          messageId: 'conditionalExpect',
          message: 'Avoid calling `expect` conditionally',
          line,
          column: 13,
          endLine: line,
          endColumn: 49,
        },
      ],
    })),
    ...(
      [
        [
          `
        it('foo', () => {
` + assertionSetup2,
          41,
        ],
        [
          `
        it.each\`\`('foo', () => {
` + assertionSetup2,
          41,
        ],
        [
          `
        it.each()('foo', () => {
` + assertionSetup2,
          41,
        ],
        [
          `
        it.skip.each\`\`('foo', () => {
` + assertionSetup2,
          41,
        ],
        [
          `
        it.skip.each()('foo', () => {
` + assertionSetup2,
          41,
        ],
        [
          `
        function getValue() {
          try {
            // do something
          } catch {
            expect(something).toHaveBeenCalled();
          }
        }

        it('foo', getValue);
     \u0020`,
          49,
        ],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `upstream catch conditions / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2019,
      },
      errors: [
        {
          messageId: 'conditionalExpect',
          message: 'Avoid calling `expect` conditionally',
          line: 6,
          column: 13,
          endLine: 6,
          endColumn,
        },
      ],
    })),
    {
      name: 'upstream promises / invalid 0',
      code: `it('works', async () => {
  await Promise.resolve()
    .then(() => { throw new Error('oh noes!'); })
    .catch(error => expect(error).toBeInstanceOf(Error));
});`,
      languageOptions: {
        ecmaVersion: 2019,
      },
      errors: [
        {
          messageId: 'conditionalExpect',
          message: 'Avoid calling `expect` conditionally',
          line: 4,
          column: 21,
          endLine: 4,
          endColumn: 56,
        },
      ],
    },
    ...(
      [
        [
          `it('works', async () => {
  await Promise.resolve()
    .then(() => { throw new Error('oh noes!'); })
    .catch(error => expect(error).toBeInstanceOf(Error))
    .then(() => { throw new Error('oh noes!'); })
    .catch(error => expect(error).toBeInstanceOf(Error))
    .then(() => { throw new Error('oh noes!'); })
    .catch(error => expect(error).toBeInstanceOf(Error));
});`,
          4,
          6,
          8,
        ],
        [
          `it('works', async () => {
  await Promise.resolve()
    .catch(error => expect(error).toBeInstanceOf(Error))
    .catch(error => expect(error).toBeInstanceOf(Error))
    .catch(error => expect(error).toBeInstanceOf(Error));
});`,
          3,
          4,
          5,
        ],
      ] as const
    ).map(([code, line1, line2, line3], index): TestCase => ({
      name: `upstream promises / invalid ${index + 1}`,
      code,
      languageOptions: {
        ecmaVersion: 2019,
      },
      errors: [
        {
          messageId: 'conditionalExpect',
          message: 'Avoid calling `expect` conditionally',
          line: line1,
          column: 21,
          endLine: line1,
          endColumn: 56,
        },
        {
          messageId: 'conditionalExpect',
          message: 'Avoid calling `expect` conditionally',
          line: line2,
          column: 21,
          endLine: line2,
          endColumn: 56,
        },
        {
          messageId: 'conditionalExpect',
          message: 'Avoid calling `expect` conditionally',
          line: line3,
          column: 21,
          endLine: line3,
          endColumn: 56,
        },
      ],
    })),
    ...(
      [
        [
          `it('works', async () => {
  await Promise.resolve()
    .catch(error => expect(error).toBeInstanceOf(Error))
    .then(() => { throw new Error('oh noes!'); })
    .then(() => { throw new Error('oh noes!'); })
    .then(() => { throw new Error('oh noes!'); });
});`,
          3,
          21,
        ],
        [
          `it('works', async () => {
  await somePromise
    .then(() => { throw new Error('oh noes!'); })
    .catch(error => expect(error).toBeInstanceOf(Error));
});`,
          4,
          21,
        ],
        [
          `it('works', async () => {
  await somePromise.catch(error => expect(error).toBeInstanceOf(Error));
});`,
          2,
          36,
        ],
      ] as const
    ).map(([code, line, column], index): TestCase => ({
      name: `upstream promises / invalid ${index + 3}`,
      code,
      languageOptions: {
        ecmaVersion: 2019,
      },
      errors: [
        {
          messageId: 'conditionalExpect',
          message: 'Avoid calling `expect` conditionally',
          line,
          column,
          endLine: line,
          endColumn: column + 35,
        },
      ],
    })),
  ],
});
