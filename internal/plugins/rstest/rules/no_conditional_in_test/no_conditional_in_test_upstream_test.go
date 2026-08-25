// TestNoConditionalInTestUpstream migrates the full valid/invalid suite from
// eslint-plugin-jest v29.16.1 src/rules/__tests__/no-conditional-in-test.test.ts
// 1:1. rslint-specific lock-ins live in no_conditional_in_test_extras_test.go.
package no_conditional_in_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	no_conditional_in "github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_conditional_in_test"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoConditionalInTestUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// ---- conditional expressions ----
		{Code: `const x = y ? 1 : 0`},
		{Code: `
      const foo = function (bar) {
        return foo ? bar : null;
      };

      it('foo', () => {
        foo();
      });
    `},
		{Code: `
      const foo = function (bar) {
        return foo ? bar : null;
      };

      it.each()('foo', function () {
        foo();
      });
    `},
		{
			Code: `
      fit.concurrent('foo', () => {
        switch('bar') {}
      })
    `,
			Skip: true, // SKIP: rstest does not support Jest's legacy fit alias.
		},

		// ---- switch statements ----
		{Code: `it('foo', () => {})`},
		{Code: `
      switch (true) {
        case true: {}
      }
    `},
		{Code: `
      it('foo', () => {});
      function myTest() {
        switch ('bar') {
        }
      }
    `},
		{Code: `
      foo('bar', () => {
        switch(baz) {}
      })
    `},
		{Code: `
      describe('foo', () => {
        switch('bar') {}
      })
    `},
		{Code: `
      describe.skip('foo', () => {
        switch('bar') {}
      })
    `},
		{Code: `
      describe.skip.each()('foo', () => {
        switch('bar') {}
      })
    `},
		{
			Code: `
      xdescribe('foo', () => {
        switch('bar') {}
      })
    `,
			Skip: true, // SKIP: rstest does not support Jest's legacy xdescribe alias.
		},
		{
			Code: `
      fdescribe('foo', () => {
        switch('bar') {}
      })
    `,
			Skip: true, // SKIP: rstest does not support Jest's legacy fdescribe alias.
		},
		{Code: `
      describe('foo', () => {
        switch('bar') {}
      })
      switch('bar') {}
    `},
		{Code: `
      describe('foo', () => {
        afterEach(() => {
          switch('bar') {}
        });
      });
    `},
		{Code: `
      const values = something.map(thing => {
        switch (thing.isFoo) {
          case true:
            return thing.foo;
          default:
            return thing.bar;
        }
      });

      it('valid', () => {
        expect(values).toStrictEqual(['foo']);
      });
    `},
		{Code: `
      describe('valid', () => {
        const values = something.map(thing => {
          switch (thing.isFoo) {
            case true:
              return thing.foo;
            default:
              return thing.bar;
          }
        });
        it('still valid', () => {
          expect(values).toStrictEqual(['foo']);
        });
      });
    `},

		// ---- if statements ----
		{Code: `if (foo) {}`},
		{Code: `it('foo', () => {})`},
		{Code: `it("foo", function () {})`},
		{Code: `it('foo', () => {}); function myTest() { if ('bar') {} }`},
		{Code: `
      foo('bar', () => {
        if (baz) {}
      })
    `},
		{Code: `
      describe('foo', () => {
        if ('bar') {}
      })
    `},
		{Code: `
      describe.skip('foo', () => {
        if ('bar') {}
      })
    `},
		{
			Code: `
      xdescribe('foo', () => {
        if ('bar') {}
      })
    `,
			Skip: true, // SKIP: rstest does not support Jest's legacy xdescribe alias.
		},
		{
			Code: `
      fdescribe('foo', () => {
        if ('bar') {}
      })
    `,
			Skip: true, // SKIP: rstest does not support Jest's legacy fdescribe alias.
		},
		{Code: `
      describe('foo', () => {
        if ('bar') {}
      })
      if ('baz') {}
    `},
		{Code: `
      describe('foo', () => {
        afterEach(() => {
          if ('bar') {}
        });
      })
    `},
		{Code: "\n      describe.each``('foo', () => {\n        afterEach(() => {\n          if ('bar') {}\n        });\n      })\n    "},
		{Code: `
      describe('foo', () => {
        beforeEach(() => {
          if ('bar') {}
        });
      })
    `},
		{Code: `const foo = bar ? foo : baz;`},
		{Code: `
      const values = something.map((thing) => {
        if (thing.isFoo) {
          return thing.foo
        } else {
          return thing.bar;
        }
      });

      describe('valid', () => {
        it('still valid', () => {
          expect(values).toStrictEqual(['foo']);
        });
      });
    `},
		{Code: `
      describe('valid', () => {
        const values = something.map((thing) => {
          if (thing.isFoo) {
            return thing.foo
          } else {
            return thing.bar;
          }
        });

        describe('still valid', () => {
          it('really still valid', () => {
            expect(values).toStrictEqual(['foo']);
          });
        });
      });
    `},
		{
			Code: `
      fit.concurrent('foo', () => {
        if ('bar') {}
      })
    `,
			Skip: true, // SKIP: rstest does not support Jest's legacy fit alias.
		},

		// ---- optional chaining ----
		{Code: `const x = obj?.foo`},
		{Code: `
      it('foo', () => {
        const value = obj?.bar;
      })
    `},
		{Code: `
      it('foo', () => {
        obj?.foo?.bar;
      })
    `},
		{Code: `
      it('foo', () => {
        obj?.foo();
      })
    `},
		{Code: `
      it('foo', () => {
        obj?.[key];
      })
    `},
		{Code: `
      test('foo', () => {
        obj?.bar;
      })
    `},
		{Code: `
      it('is valid', () => {
        const values = something.map(thing => thing?.foo);

        expect(values).toStrictEqual(['foo']);
      });
    `},

		// ---- optional chaining with allowOptionalChaining=false ----
		{
			Code:    `const x = obj?.foo`,
			Options: map[string]any{"allowOptionalChaining": false},
		},
		{
			Code: `
        const foo = obj?.bar;

        it('foo', () => {
          expect(foo).toBe(undefined);
        });
      `,
			Options: map[string]any{"allowOptionalChaining": false},
		},
		{
			Code: `
        describe('foo', () => {
          const val = obj?.bar;
        })
      `,
			Options: map[string]any{"allowOptionalChaining": false},
		},
		{
			Code: `
        describe('foo', () => {
          beforeEach(() => {
            const val = obj?.bar;
          });
        })
      `,
			Options: map[string]any{"allowOptionalChaining": false},
		},
		{
			Code: `
        describe('foo', () => {
          afterEach(() => {
            const val = obj?.bar;
          });
        })
      `,
			Options: map[string]any{"allowOptionalChaining": false},
		},
		{
			Code: `
        const values = something.map(thing => thing?.foo);

        it('valid', () => {
          expect(values).toStrictEqual(['foo']);
        });
      `,
			Options: map[string]any{"allowOptionalChaining": false},
		},
		{
			Code: `
        describe('valid', () => {
          const values = something.map(thing => thing?.foo);
          it('still valid', () => {
            expect(values).toStrictEqual(['foo']);
          });
        });
      `,
			Options: map[string]any{"allowOptionalChaining": false},
		},
	}

	invalid := []rule_tester.InvalidTestCase{
		// ---- conditional expressions ----
		{
			Code: `
        it('foo', () => {
          expect(bar ? foo : baz).toBe(boo);
        })
      `,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionalInTest",
				Line:      2,
				Column:    10,
				EndLine:   2,
				EndColumn: 25,
			}},
		},
		{
			Code: `
        it('foo', function () {
          const foo = function (bar) {
            return foo ? bar : null;
          };
        });
      `,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionalInTest",
				Line:      3,
				Column:    12,
				EndLine:   3,
				EndColumn: 28,
			}},
		},
		{
			Code: `
        it('foo', () => {
          const foo = bar ? foo : baz;
        })
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 2, Column: 15}},
		},
		{
			Code: `
        it('foo', () => {
          const foo = bar ? foo : baz;
        })
        const foo = bar ? foo : baz;
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 2, Column: 15}},
		},
		{
			Code: `
        it('foo', () => {
          const foo = bar ? foo : baz;
          const anotherFoo = anotherBar ? anotherFoo : anotherBaz;
        })
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "conditionalInTest", Line: 2, Column: 15},
				{MessageId: "conditionalInTest", Line: 3, Column: 22},
			},
		},

		// ---- switch statements ----
		{
			Code: `
        it('is invalid', () => {
          const values = something.map(thing => {
            switch (thing.isFoo) {
              case true:
                return thing.foo;
              default:
                return thing.bar;
            }
          });

          expect(values).toStrictEqual(['foo']);
        });
      `,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionalInTest",
				Line:      3,
				Column:    5,
				EndLine:   8,
				EndColumn: 6,
			}},
		},
		{
			Code: `
        it('foo', () => {
          switch (true) {
            case true: {}
          }
        })
      `,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionalInTest",
				Line:      2,
				Column:    3,
				EndLine:   4,
				EndColumn: 4,
			}},
		},
		{
			Code: `
        it('foo', () => {
          switch('bar') {}
        })
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 2, Column: 3}},
		},
		{
			Code: `
        it.skip('foo', () => {
          switch('bar') {}
        })
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 2, Column: 3}},
		},
		{
			Code: `
        it.only('foo', () => {
          switch('bar') {}
        })
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 2, Column: 3}},
		},
		{
			Code: `
        xit('foo', () => {
          switch('bar') {}
        })
      `,
			Skip:   true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 2, Column: 3}},
		},
		{
			Code: `
        fit('foo', () => {
          switch('bar') {}
        })
      `,
			Skip:   true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 2, Column: 3}},
		},
		{
			Code: `
        test('foo', () => {
          switch('bar') {}
        })
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 2, Column: 3}},
		},
		{
			Code: `
        test.skip('foo', () => {
          switch('bar') {}
        })
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 2, Column: 3}},
		},
		{
			Code: `
        test.only('foo', () => {
          switch('bar') {}
        })
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 2, Column: 3}},
		},
		{
			Code: `
        xtest('foo', () => {
          switch('bar') {}
        })
      `,
			Skip:   true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 2, Column: 3}},
		},
		{
			Code: `
        xtest('foo', function () {
          switch('bar') {}
        })
      `,
			Skip:   true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 2, Column: 3}},
		},
		{
			Code: `
        describe('foo', () => {
          it('bar', () => {

            switch('bar') {}
          })
        })
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 4, Column: 5}},
		},
		{
			Code: `
        describe('foo', () => {
          it('bar', () => {
            switch('bar') {}
          })
          it('baz', () => {
            switch('qux') {}
            switch('quux') {}
          })
        })
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "conditionalInTest", Line: 3, Column: 5},
				{MessageId: "conditionalInTest", Line: 6, Column: 5},
				{MessageId: "conditionalInTest", Line: 7, Column: 5},
			},
		},
		{
			Code: `
        it('foo', () => {
          callExpression()
          switch ('bar') {}
        })
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 3, Column: 3}},
		},
		{
			Code: `
        describe('valid', () => {
          describe('still valid', () => {
            it('is not valid', () => {
              const values = something.map((thing) => {
                switch (thing.isFoo) {
                  case true:
                    return thing.foo;
                  default:
                    return thing.bar;
                }
              });

              switch('invalid') {
                case true:
                  expect(values).toStrictEqual(['foo']);
              }
            });
          });
        });
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "conditionalInTest", Line: 5, Column: 9},
				{MessageId: "conditionalInTest", Line: 13, Column: 7},
			},
		},

		// ---- if statements ----
		{
			Code: `
        it('foo', () => {
          const foo = function(bar) {
            if (bar) {
              return 1;
            } else {
              return 2;
            }
          };
        });
      `,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionalInTest",
				Line:      3,
				Column:    5,
				EndLine:   7,
				EndColumn: 6,
			}},
		},
		{
			Code: `
        it('foo', () => {
          function foo(bar) {
            if (bar) {
              return 1;
            } else {
              return 2;
            }
          };
        });
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 3, Column: 5}},
		},
		{
			Code: `
        it('foo', () => {
          if ('bar') {}
        })
      `,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionalInTest",
				Line:      2,
				Column:    3,
				EndLine:   2,
				EndColumn: 16,
			}},
		},
		{
			Code: `
        it.skip('foo', () => {
          if ('bar') {}
        })
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 2, Column: 3}},
		},
		{
			Code: `
        it.skip('foo', function () {
          if ('bar') {}
        })
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 2, Column: 3}},
		},
		{
			Code: `
        it.only('foo', () => {
          if ('bar') {}
        })
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 2, Column: 3}},
		},
		{
			Code: `
        xit('foo', () => {
          if ('bar') {}
        })
      `,
			Skip:   true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 2, Column: 3}},
		},
		{
			Code: `
        fit('foo', () => {
          if ('bar') {}
        })
      `,
			Skip:   true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 2, Column: 3}},
		},
		{
			Code: `
        test('foo', () => {
          if ('bar') {}
        })
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 2, Column: 3}},
		},
		{
			Code: `
        test.skip('foo', () => {
          if ('bar') {}
        })
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 2, Column: 3}},
		},
		{
			Code: `
        test.only('foo', () => {
          if ('bar') {}
        })
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 2, Column: 3}},
		},
		{
			Code: `
        xtest('foo', () => {
          if ('bar') {}
        })
      `,
			Skip:   true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 2, Column: 3}},
		},
		{
			Code: `
        describe('foo', () => {
          it('bar', () => {
            if ('bar') {}
          })
        })
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 3, Column: 5}},
		},
		{
			Code: `
        describe('foo', () => {
          it('bar', () => {
            if ('bar') {}
          })
          it('baz', () => {
            if ('qux') {}
            if ('quux') {}
          })
        })
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "conditionalInTest", Line: 3, Column: 5},
				{MessageId: "conditionalInTest", Line: 6, Column: 5},
				{MessageId: "conditionalInTest", Line: 7, Column: 5},
			},
		},
		{
			Code: `
        it('foo', () => {
          callExpression()
          if ('bar') {}
        })
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 3, Column: 3}},
		},
		{
			Code:   "\n        it.each``('foo', () => {\n          callExpression()\n          if ('bar') {}\n        })\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 3, Column: 3}},
		},
		{
			Code: `
        it.each()('foo', () => {
          callExpression()
          if ('bar') {}
        })
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 3, Column: 3}},
		},
		{
			Code:   "\n        it.only.each``('foo', () => {\n          callExpression()\n          if ('bar') {}\n        })\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 3, Column: 3}},
		},
		{
			Code: `
        it.only.each()('foo', () => {
          callExpression()
          if ('bar') {}
        })
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 3, Column: 3}},
		},
		{
			Code: `
        describe('valid', () => {
          describe('still valid', () => {
            it('is invalid', () => {
              const values = something.map((thing) => {
                if (thing.isFoo) {
                  return thing.foo
                } else {
                  return thing.bar;
                }
              });

              if ('invalid') {
                expect(values).toStrictEqual(['foo']);
              }
            });
          });
        });
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "conditionalInTest", Line: 5, Column: 9},
				{MessageId: "conditionalInTest", Line: 12, Column: 7},
			},
		},
		{
			Code: `
        test("shows error", () => {
          if (1 === 2) {
            expect(true).toBe(false);
          }
        });

        test("does not show error", () => {
          setTimeout(() => console.log("noop"));
          if (1 === 2) {
            expect(true).toBe(false);
          }
        });
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "conditionalInTest", Line: 2, Column: 3},
				{MessageId: "conditionalInTest", Line: 9, Column: 3},
			},
		},

		// ---- optional chaining with allowOptionalChaining=false ----
		{
			Code: `
        it('foo', () => {
          const value = obj?.bar;
        })
      `,
			Options: map[string]any{"allowOptionalChaining": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionalInTest",
				Line:      2,
				Column:    17,
				EndLine:   2,
				EndColumn: 25,
			}},
		},
		{
			Code: `
        it('foo', () => {
          obj?.foo?.bar;
        })
      `,
			Options: map[string]any{"allowOptionalChaining": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionalInTest",
				Line:      2,
				Column:    3,
				EndLine:   2,
				EndColumn: 16,
			}},
		},
		{
			Code: `
        it('foo', () => {
          obj?.foo();
        })
      `,
			Options: map[string]any{"allowOptionalChaining": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 2, Column: 3}},
		},
		{
			Code: `
        it('foo', () => {
          obj?.[key];
        })
      `,
			Options: map[string]any{"allowOptionalChaining": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 2, Column: 3}},
		},
		{
			Code: `
        it.skip('foo', () => {
          obj?.bar;
        })
      `,
			Options: map[string]any{"allowOptionalChaining": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 2, Column: 3}},
		},
		{
			Code: `
        it.only('foo', () => {
          obj?.bar;
        })
      `,
			Options: map[string]any{"allowOptionalChaining": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 2, Column: 3}},
		},
		{
			Code: `
        test('foo', () => {
          obj?.bar;
        })
      `,
			Options: map[string]any{"allowOptionalChaining": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 2, Column: 3}},
		},
		{
			Code: `
        xtest('foo', () => {
          obj?.bar;
        })
      `,
			Skip:    true,
			Options: map[string]any{"allowOptionalChaining": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 2, Column: 3}},
		},
		{
			Code: `
        fit('foo', () => {
          obj?.bar;
        })
      `,
			Skip:    true,
			Options: map[string]any{"allowOptionalChaining": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 2, Column: 3}},
		},
		{
			Code: `
        xit('foo', () => {
          obj?.bar;
        })
      `,
			Skip:    true,
			Options: map[string]any{"allowOptionalChaining": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 2, Column: 3}},
		},
		{
			Code: `
        describe('foo', () => {
          it('bar', () => {
            obj?.bar;
          })
        })
      `,
			Options: map[string]any{"allowOptionalChaining": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 3, Column: 5}},
		},
		{
			Code: `
        it('is invalid', () => {
          const values = something.map(thing => thing?.foo);

          expect(values).toStrictEqual(['foo']);
        });
      `,
			Options: map[string]any{"allowOptionalChaining": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 2, Column: 41}},
		},
	}

	for index := range valid {
		valid[index].Code = dedent(valid[index].Code)
	}
	for index := range invalid {
		invalid[index].Code = dedent(invalid[index].Code)
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_conditional_in.NoConditionalInTestRule,
		valid,
		invalid,
	)
}

func dedent(code string) string {
	lines := strings.Split(code, "\n")
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}

	indent := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		current := len(line) - len(strings.TrimLeft(line, " \t"))
		if indent == -1 || current < indent {
			indent = current
		}
	}
	if indent <= 0 {
		return strings.Join(lines, "\n")
	}
	for index := range lines {
		if len(lines[index]) >= indent {
			lines[index] = lines[index][indent:]
		}
	}
	return strings.Join(lines, "\n")
}
