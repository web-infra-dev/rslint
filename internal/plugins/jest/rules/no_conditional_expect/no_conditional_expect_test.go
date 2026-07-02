package no_conditional_expect_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_conditional_expect"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoConditionalExpectRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_conditional_expect.NoConditionalExpectRule,
		[]rule_tester.ValidTestCase{
			{Code: `
      it('foo', () => {
        expect(1).toBe(2);
      });
    `},
			{Code: `
      it('foo', () => {
        expect(!true).toBe(false);
      });
    `},
			{Code: `
      it('foo', () => {
        process.env.FAIL && setNum(1);

        expect(num).toBe(2);
      });
    `},
			{Code: `
      function getValue() {
        let num = 2;

        process.env.FAIL && setNum(1);

        return num;
      }

      it('foo', () => {
        expect(getValue()).toBe(2);
      });
    `},
			{Code: `
      it('foo', () => {
        process.env.FAIL || setNum(1);

        expect(num).toBe(2);
      });
    `},
			{Code: `
      function getValue() {
        let num = 2;

        process.env.FAIL || setNum(1);

        return num;
      }

      it('foo', () => {
        expect(getValue()).toBe(2);
      });
    `},
			{Code: `
      it('foo', () => {
        const num = process.env.FAIL ? 1 : 2;

        expect(num).toBe(2);
      });
    `},
			{Code: `
      function getValue() {
        return process.env.FAIL ? 1 : 2
      }

      it('foo', () => {
        expect(getValue()).toBe(2);
      });
    `},
			{Code: `
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
    `},
			{Code: `
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
    `},
			{Code: `
      it('foo', () => {
        let num = 2;

        if(process.env.FAIL) {
          num = 1;
        }

        expect(num).toBe(2);
      });
    `},
			{Code: `
      function getValue() {
        if(process.env.FAIL) {
          return 1;
        }

        return 2;
      }

      it('foo', () => {
        expect(getValue()).toBe(2);
      });
    `},
			{Code: `
      it('foo', () => {
        try {
          // do something
        } catch {
          // ignore errors
        } finally {
          expect(something).toHaveBeenCalled();
        }
      });
    `},
			{Code: `
      it('foo', () => {
        try {
          // do something
        } catch {
          // ignore errors
        }

        expect(something).toHaveBeenCalled();
      });
    `},
			{Code: `
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
    `},
			{Code: `
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
    `},
			{Code: `
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
    `},
			{Code: `
      it('works', async () => {
        await doSomething().catch(error => error);

        expect(error).toBeInstanceOf(Error);
      });
    `},
			{Code: `
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
    `},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `
        it('foo', () => {
          something && expect(something).toHaveBeenCalled();
        })
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `
        it('foo', () => {
          a || b && expect(something).toHaveBeenCalled();
        })
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `
        it('foo', () => {
          (a || b) && expect(something).toHaveBeenCalled();
        })
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `
        it('foo', () => {
          a || (b && expect(something).toHaveBeenCalled());
        })
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `
        it('foo', () => {
          a && b && expect(something).toHaveBeenCalled();
        })
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `
        it('foo', () => {
          a && b || expect(something).toHaveBeenCalled();
        })
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `
        it('foo', () => {
          (a && b) || expect(something).toHaveBeenCalled();
        })
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `
        function getValue() {
          something && expect(something).toHaveBeenCalled();
        }

        it('foo', getValue);
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `
        const getValue = function() {
          something && expect(something).toHaveBeenCalled();
        }

        it('foo', getValue);
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `
        const getValue = () => {
          something || expect(something).toHaveBeenCalled();
        }

        it('foo', getValue);
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `
        it('foo', () => {
          something || expect(something).toHaveBeenCalled();
        })
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: "\n        it.each``('foo', () => {\n          something || expect(something).toHaveBeenCalled();\n        })\n      ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `
        it.each()('foo', () => {
          something || expect(something).toHaveBeenCalled();
        })
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `
        function getValue() {
          something || expect(something).toHaveBeenCalled();
        }

        it('foo', getValue);
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `
        it('foo', () => {
          something ? expect(something).toHaveBeenCalled() : noop();
        })
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `
        function getValue() {
          something ? expect(something).toHaveBeenCalled() : noop();
        }

        it('foo', getValue);
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `
        it('foo', () => {
          something ? noop() : expect(something).toHaveBeenCalled();
        })
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: "\n        it.each``('foo', () => {\n          something ? noop() : expect(something).toHaveBeenCalled();\n        })\n      ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `
        it.each()('foo', () => {
          something ? noop() : expect(something).toHaveBeenCalled();
        })
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `
        function getValue() {
          something ? noop() : expect(something).toHaveBeenCalled();
        }

        it('foo', getValue);
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `
        it('foo', () => {
          switch(something) {
            case 'value':
              break;
            default:
              expect(something).toHaveBeenCalled();
          }
        })
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `
        it('foo', () => {
          switch(something) {
            case 'value':
              expect(something).toHaveBeenCalled();
            default:
              break;
          }
        })
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: "\n        it.each``('foo', () => {\n          switch(something) {\n            case 'value':\n              expect(something).toHaveBeenCalled();\n            default:\n              break;\n          }\n        })\n      ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `
        it.each()('foo', () => {
          switch(something) {
            case 'value':
              expect(something).toHaveBeenCalled();
            default:
              break;
          }
        })
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `
        function getValue() {
          switch(something) {
            case 'value':
              break;
            default:
              expect(something).toHaveBeenCalled();
          }
        }

        it('foo', getValue);
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `
        it('foo', () => {
          if(doSomething) {
            expect(something).toHaveBeenCalled();
          }
        })
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `
        it('foo', () => {
          if(!doSomething) {
            // do nothing
          } else {
            expect(something).toHaveBeenCalled();
          }
        })
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: "\n        it.each``('foo', () => {\n          if(!doSomething) {\n            // do nothing\n          } else {\n            expect(something).toHaveBeenCalled();\n          }\n        })\n      ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `
        it.each()('foo', () => {
          if(!doSomething) {
            // do nothing
          } else {
            expect(something).toHaveBeenCalled();
          }
        })
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `
        function getValue() {
          if(doSomething) {
            expect(something).toHaveBeenCalled();
          }
        }

        it('foo', getValue);
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `
        function getValue() {
          if(!doSomething) {
            // do nothing
          } else {
            expect(something).toHaveBeenCalled();
          }
        }

        it('foo', getValue);
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `
        it('foo', () => {
          try {

          } catch (err) {
            expect(err).toMatch('Error');
          }
        })
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: "\n        it.each``('foo', () => {\n          try {\n\n          } catch (err) {\n            expect(err).toMatch('Error');\n          }\n        })\n      ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `
        it.each()('foo', () => {
          try {

          } catch (err) {
            expect(err).toMatch('Error');
          }
        })
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: "\n        it.skip.each``('foo', () => {\n          try {\n\n          } catch (err) {\n            expect(err).toMatch('Error');\n          }\n        })\n      ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `
        it.skip.each()('foo', () => {
          try {

          } catch (err) {
            expect(err).toMatch('Error');
          }
        })
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `
        function getValue() {
          try {
            // do something
          } catch {
            expect(something).toHaveBeenCalled();
          }
        }

        it('foo', getValue);
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `
        it('works', async () => {
          await Promise.resolve()
            .then(() => { throw new Error('oh noes!'); })
            .catch(error => expect(error).toBeInstanceOf(Error));
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `
        it('works', async () => {
          await Promise.resolve()
            .then(() => { throw new Error('oh noes!'); })
            .catch(error => expect(error).toBeInstanceOf(Error))
            .then(() => { throw new Error('oh noes!'); })
            .catch(error => expect(error).toBeInstanceOf(Error))
            .then(() => { throw new Error('oh noes!'); })
            .catch(error => expect(error).toBeInstanceOf(Error));
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `
        it('works', async () => {
          await Promise.resolve()
            .catch(error => expect(error).toBeInstanceOf(Error))
            .catch(error => expect(error).toBeInstanceOf(Error))
            .catch(error => expect(error).toBeInstanceOf(Error));
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `
        it('works', async () => {
          await Promise.resolve()
            .catch(error => expect(error).toBeInstanceOf(Error))
            .then(() => { throw new Error('oh noes!'); })
            .then(() => { throw new Error('oh noes!'); })
            .then(() => { throw new Error('oh noes!'); });
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `
        it('works', async () => {
          await somePromise
            .then(() => { throw new Error('oh noes!'); })
            .catch(error => expect(error).toBeInstanceOf(Error));
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `
        it('works', async () => {
          await somePromise.catch(error => expect(error).toBeInstanceOf(Error));
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
		},
	)
}
