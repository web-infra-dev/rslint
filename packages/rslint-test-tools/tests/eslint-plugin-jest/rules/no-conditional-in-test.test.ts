// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('no-conditional-in-test', {
  valid: [
    ...(
      [
        'const x = y ? 1 : 0',
        'const x = foo && bar',
        'const x = foo || bar',
        'const x = foo ?? bar',
        `
      const foo = function (bar) {
        return foo ? bar : null;
      };

      it('foo', () => {
        foo();
      });
   \u0020`,
        `
      const foo = function (bar) {
        return foo ? bar : null;
      };

      it.each()('foo', function () {
        foo();
      });
   \u0020`,
        `
      fit.concurrent('foo', () => {
        switch('bar') {}
      })
   \u0020`,
        "it('foo', () => {})",
        `
      switch (true) {
        case true: {}
      }
   \u0020`,
        `
      it('foo', () => {});
      function myTest() {
        switch ('bar') {
        }
      }
   \u0020`,
        `
      foo('bar', () => {
        switch(baz) {}
      })
   \u0020`,
        `
      describe('foo', () => {
        switch('bar') {}
      })
   \u0020`,
        `
      describe.skip('foo', () => {
        switch('bar') {}
      })
   \u0020`,
        `
      describe.skip.each()('foo', () => {
        switch('bar') {}
      })
   \u0020`,
        `
      xdescribe('foo', () => {
        switch('bar') {}
      })
   \u0020`,
        `
      fdescribe('foo', () => {
        switch('bar') {}
      })
   \u0020`,
        `
      describe('foo', () => {
        switch('bar') {}
      })
      switch('bar') {}
   \u0020`,
        `
      describe('foo', () => {
        afterEach(() => {
          switch('bar') {}
        });
      });
   \u0020`,
        `
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
   \u0020`,
        `
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
   \u0020`,
        'if (foo) {}',
        "it('foo', () => {})",
        'it("foo", function () {})',
        "it('foo', () => {}); function myTest() { if ('bar') {} }",
        `
      foo('bar', () => {
        if (baz) {}
      })
   \u0020`,
        `
      describe('foo', () => {
        if ('bar') {}
      })
   \u0020`,
        `
      describe.skip('foo', () => {
        if ('bar') {}
      })
   \u0020`,
        `
      xdescribe('foo', () => {
        if ('bar') {}
      })
   \u0020`,
        `
      fdescribe('foo', () => {
        if ('bar') {}
      })
   \u0020`,
        `
      describe('foo', () => {
        if ('bar') {}
      })
      if ('baz') {}
   \u0020`,
        `
      describe('foo', () => {
        afterEach(() => {
          if ('bar') {}
        });
      })
   \u0020`,
        `
      describe.each\`\`('foo', () => {
        afterEach(() => {
          if ('bar') {}
        });
      })
   \u0020`,
        `
      describe('foo', () => {
        beforeEach(() => {
          if ('bar') {}
        });
      })
   \u0020`,
        'const foo = bar ? foo : baz;',
        `
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
   \u0020`,
        `
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
   \u0020`,
        `
      fit.concurrent('foo', () => {
        if ('bar') {}
      })
   \u0020`,
        'const x = obj?.foo',
        `
      it('foo', () => {
        const value = obj?.bar;
      })
   \u0020`,
        `
      it('foo', () => {
        obj?.foo?.bar;
      })
   \u0020`,
        `
      it('foo', () => {
        obj?.foo();
      })
   \u0020`,
        `
      it('foo', () => {
        obj?.[key];
      })
   \u0020`,
        `
      test('foo', () => {
        obj?.bar;
      })
   \u0020`,
        `
      it('is valid', () => {
        const values = something.map(thing => thing?.foo);

        expect(values).toStrictEqual(['foo']);
      });
   \u0020`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    ...(
      [
        'const x = obj?.foo',
        `
        const foo = obj?.bar;

        it('foo', () => {
          expect(foo).toBe(undefined);
        });
     \u0020`,
        `
        describe('foo', () => {
          const val = obj?.bar;
        })
     \u0020`,
        `
        describe('foo', () => {
          beforeEach(() => {
            const val = obj?.bar;
          });
        })
     \u0020`,
        `
        describe('foo', () => {
          afterEach(() => {
            const val = obj?.bar;
          });
        })
     \u0020`,
        `
        const values = something.map(thing => thing?.foo);

        it('valid', () => {
          expect(values).toStrictEqual(['foo']);
        });
     \u0020`,
        `
        describe('valid', () => {
          const values = something.map(thing => thing?.foo);
          it('still valid', () => {
            expect(values).toStrictEqual(['foo']);
          });
        });
     \u0020`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index + 44}`,
      code,
      options: [
        {
          allowOptionalChaining: false,
        },
      ],
    })),
    ...(
      [
        'const x = y ? 1 : 0',
        `const foo = function (bar) {
  return foo ? bar : null;
};

it('foo', () => {
  foo();
});`,
        `const foo = function (bar) {
  return foo ? bar : null;
};

it.each()('foo', function () {
  foo();
});`,
        `fit.concurrent('foo', () => {
  switch('bar') {}
})`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream conditional expressions / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2020,
      },
    })),
    ...(
      [
        "it('foo', () => {})",
        `switch (true) {
  case true: {}
}`,
        `it('foo', () => {});
function myTest() {
  switch ('bar') {
  }
}`,
        `foo('bar', () => {
  switch(baz) {}
})`,
        `describe('foo', () => {
  switch('bar') {}
})`,
        `describe.skip('foo', () => {
  switch('bar') {}
})`,
        `describe.skip.each()('foo', () => {
  switch('bar') {}
})`,
        `xdescribe('foo', () => {
  switch('bar') {}
})`,
        `fdescribe('foo', () => {
  switch('bar') {}
})`,
        `describe('foo', () => {
  switch('bar') {}
})
switch('bar') {}`,
        `describe('foo', () => {
  afterEach(() => {
    switch('bar') {}
  });
});`,
        `const values = something.map(thing => {
  switch (thing.isFoo) {
    case true:
      return thing.foo;
    default:
      return thing.bar;
  }
});

it('valid', () => {
  expect(values).toStrictEqual(['foo']);
});`,
        `describe('valid', () => {
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
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream switch statements / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2020,
      },
    })),
    ...(
      [
        ['if (foo) {}', 'upstream if statements / valid 0'],
        ['it("foo", function () {})', 'upstream if statements / valid 2'],
        [
          "it('foo', () => {}); function myTest() { if ('bar') {} }",
          'upstream if statements / valid 3',
        ],
        [
          `foo('bar', () => {
  if (baz) {}
})`,
          'upstream if statements / valid 4',
        ],
        [
          `describe('foo', () => {
  if ('bar') {}
})`,
          'upstream if statements / valid 5',
        ],
        [
          `describe.skip('foo', () => {
  if ('bar') {}
})`,
          'upstream if statements / valid 6',
        ],
        [
          `xdescribe('foo', () => {
  if ('bar') {}
})`,
          'upstream if statements / valid 7',
        ],
        [
          `fdescribe('foo', () => {
  if ('bar') {}
})`,
          'upstream if statements / valid 8',
        ],
        [
          `describe('foo', () => {
  if ('bar') {}
})
if ('baz') {}`,
          'upstream if statements / valid 9',
        ],
        [
          `describe('foo', () => {
  afterEach(() => {
    if ('bar') {}
  });
})`,
          'upstream if statements / valid 10',
        ],
        [
          `describe.each\`\`('foo', () => {
  afterEach(() => {
    if ('bar') {}
  });
})`,
          'upstream if statements / valid 11',
        ],
        [
          `describe('foo', () => {
  beforeEach(() => {
    if ('bar') {}
  });
})`,
          'upstream if statements / valid 12',
        ],
        ['const foo = bar ? foo : baz;', 'upstream if statements / valid 13'],
        [
          `const values = something.map((thing) => {
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
});`,
          'upstream if statements / valid 14',
        ],
        [
          `describe('valid', () => {
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
});`,
          'upstream if statements / valid 15',
        ],
        [
          `fit.concurrent('foo', () => {
  if ('bar') {}
})`,
          'upstream if statements / valid 16',
        ],
      ] as const
    ).map(([code, name]): ValidTestCase => ({
      name,
      code,
      languageOptions: {
        ecmaVersion: 2020,
      },
    })),
    ...(
      [
        'const x = obj?.foo',
        `it('foo', () => {
  const value = obj?.bar;
})`,
        `it('foo', () => {
  obj?.foo?.bar;
})`,
        `it('foo', () => {
  obj?.foo();
})`,
        `it('foo', () => {
  obj?.[key];
})`,
        `test('foo', () => {
  obj?.bar;
})`,
        `it('is valid', () => {
  const values = something.map(thing => thing?.foo);

  expect(values).toStrictEqual(['foo']);
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream optional chaining / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2020,
      },
    })),
    ...(
      [
        'const x = obj?.foo',
        `const foo = obj?.bar;

it('foo', () => {
  expect(foo).toBe(undefined);
});`,
        `describe('foo', () => {
  const val = obj?.bar;
})`,
        `describe('foo', () => {
  beforeEach(() => {
    const val = obj?.bar;
  });
})`,
        `describe('foo', () => {
  afterEach(() => {
    const val = obj?.bar;
  });
})`,
        `const values = something.map(thing => thing?.foo);

it('valid', () => {
  expect(values).toStrictEqual(['foo']);
});`,
        `describe('valid', () => {
  const values = something.map(thing => thing?.foo);
  it('still valid', () => {
    expect(values).toStrictEqual(['foo']);
  });
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream optional chaining with allowOptionalChaining=false / valid ${index}`,
      code,
      options: [
        {
          allowOptionalChaining: false,
        },
      ],
      languageOptions: {
        ecmaVersion: 2020,
      },
    })),
  ],
  invalid: [
    ...(
      [
        [
          `
        it('foo', () => {
          expect(bar ? foo : baz).toBe(boo);
        })
     \u0020`,
          3,
          18,
          33,
        ],
        [
          `
        it('foo', () => {
          foo && expect(foo).toBe(true);
        })
     \u0020`,
          3,
          11,
          40,
        ],
        [
          `
        it('foo', () => {
          const value = foo || bar;
        })
     \u0020`,
          3,
          25,
          35,
        ],
        [
          `
        it('foo', () => {
          const value = foo ?? bar;
        })
     \u0020`,
          3,
          25,
          35,
        ],
        [
          `
        it('foo', function () {
          const foo = function (bar) {
            return foo ? bar : null;
          };
        });
     \u0020`,
          4,
          20,
          36,
        ],
        [
          `
        it('foo', () => {
          const foo = bar ? foo : baz;
        })
     \u0020`,
          3,
          23,
          38,
        ],
        [
          `
        it('foo', () => {
          const foo = bar ? foo : baz;
        })
        const foo = bar ? foo : baz;
     \u0020`,
          3,
          23,
          38,
        ],
      ] as const
    ).map(([code, line, column, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'conditionalInTest',
          message: 'Avoid having conditionals in tests',
          line,
          column,
          endLine: line,
          endColumn,
        },
      ],
    })),
    {
      name: 'run 0 invalid 7',
      code: `
        it('foo', () => {
          const foo = bar ? foo : baz;
          const anotherFoo = anotherBar ? anotherFoo : anotherBaz;
        })
     \u0020`,
      errors: [
        {
          messageId: 'conditionalInTest',
          message: 'Avoid having conditionals in tests',
          line: 3,
          column: 23,
          endLine: 3,
          endColumn: 38,
        },
        {
          messageId: 'conditionalInTest',
          message: 'Avoid having conditionals in tests',
          line: 4,
          column: 30,
          endLine: 4,
          endColumn: 66,
        },
      ],
    },
    ...(
      [
        [
          `
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
     \u0020`,
          4,
          13,
          9,
          14,
        ],
        [
          `
        it('foo', () => {
          switch (true) {
            case true: {}
          }
        })
     \u0020`,
          3,
          11,
          5,
          12,
        ],
        [
          `
        it('foo', () => {
          switch('bar') {}
        })
     \u0020`,
          3,
          11,
          3,
          27,
        ],
        [
          `
        it.skip('foo', () => {
          switch('bar') {}
        })
     \u0020`,
          3,
          11,
          3,
          27,
        ],
        [
          `
        it.only('foo', () => {
          switch('bar') {}
        })
     \u0020`,
          3,
          11,
          3,
          27,
        ],
        [
          `
        xit('foo', () => {
          switch('bar') {}
        })
     \u0020`,
          3,
          11,
          3,
          27,
        ],
        [
          `
        fit('foo', () => {
          switch('bar') {}
        })
     \u0020`,
          3,
          11,
          3,
          27,
        ],
        [
          `
        test('foo', () => {
          switch('bar') {}
        })
     \u0020`,
          3,
          11,
          3,
          27,
        ],
        [
          `
        test.skip('foo', () => {
          switch('bar') {}
        })
     \u0020`,
          3,
          11,
          3,
          27,
        ],
        [
          `
        test.only('foo', () => {
          switch('bar') {}
        })
     \u0020`,
          3,
          11,
          3,
          27,
        ],
        [
          `
        xtest('foo', () => {
          switch('bar') {}
        })
     \u0020`,
          3,
          11,
          3,
          27,
        ],
        [
          `
        xtest('foo', function () {
          switch('bar') {}
        })
     \u0020`,
          3,
          11,
          3,
          27,
        ],
        [
          `
        describe('foo', () => {
          it('bar', () => {

            switch('bar') {}
          })
        })
     \u0020`,
          5,
          13,
          5,
          29,
        ],
      ] as const
    ).map(([code, line, column, endLine, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index + 8}`,
      code,
      errors: [
        {
          messageId: 'conditionalInTest',
          message: 'Avoid having conditionals in tests',
          line,
          column,
          endLine,
          endColumn,
        },
      ],
    })),
    {
      name: 'run 0 invalid 21',
      code: `
        describe('foo', () => {
          it('bar', () => {
            switch('bar') {}
          })
          it('baz', () => {
            switch('qux') {}
            switch('quux') {}
          })
        })
     \u0020`,
      errors: [
        {
          messageId: 'conditionalInTest',
          message: 'Avoid having conditionals in tests',
          line: 4,
          column: 13,
          endLine: 4,
          endColumn: 29,
        },
        {
          messageId: 'conditionalInTest',
          message: 'Avoid having conditionals in tests',
          line: 7,
          column: 13,
          endLine: 7,
          endColumn: 29,
        },
        {
          messageId: 'conditionalInTest',
          message: 'Avoid having conditionals in tests',
          line: 8,
          column: 13,
          endLine: 8,
          endColumn: 30,
        },
      ],
    },
    {
      name: 'run 0 invalid 22',
      code: `
        it('foo', () => {
          callExpression()
          switch ('bar') {}
        })
     \u0020`,
      errors: [
        {
          messageId: 'conditionalInTest',
          message: 'Avoid having conditionals in tests',
          line: 4,
          column: 11,
          endLine: 4,
          endColumn: 28,
        },
      ],
    },
    {
      name: 'run 0 invalid 23',
      code: `
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
     \u0020`,
      errors: [
        {
          messageId: 'conditionalInTest',
          message: 'Avoid having conditionals in tests',
          line: 6,
          column: 17,
          endLine: 11,
          endColumn: 18,
        },
        {
          messageId: 'conditionalInTest',
          message: 'Avoid having conditionals in tests',
          line: 14,
          column: 15,
          endLine: 17,
          endColumn: 16,
        },
      ],
    },
    ...(
      [
        [
          `
        it('foo', () => {
          const foo = function(bar) {
            if (bar) {
              return 1;
            } else {
              return 2;
            }
          };
        });
     \u0020`,
          4,
          13,
          8,
          14,
        ],
        [
          `
        it('foo', () => {
          function foo(bar) {
            if (bar) {
              return 1;
            } else {
              return 2;
            }
          };
        });
     \u0020`,
          4,
          13,
          8,
          14,
        ],
        [
          `
        it('foo', () => {
          if ('bar') {}
        })
     \u0020`,
          3,
          11,
          3,
          24,
        ],
        [
          `
        it.skip('foo', () => {
          if ('bar') {}
        })
     \u0020`,
          3,
          11,
          3,
          24,
        ],
        [
          `
        it.skip('foo', function () {
          if ('bar') {}
        })
     \u0020`,
          3,
          11,
          3,
          24,
        ],
        [
          `
        it.only('foo', () => {
          if ('bar') {}
        })
     \u0020`,
          3,
          11,
          3,
          24,
        ],
        [
          `
        xit('foo', () => {
          if ('bar') {}
        })
     \u0020`,
          3,
          11,
          3,
          24,
        ],
        [
          `
        fit('foo', () => {
          if ('bar') {}
        })
     \u0020`,
          3,
          11,
          3,
          24,
        ],
        [
          `
        test('foo', () => {
          if ('bar') {}
        })
     \u0020`,
          3,
          11,
          3,
          24,
        ],
        [
          `
        test.skip('foo', () => {
          if ('bar') {}
        })
     \u0020`,
          3,
          11,
          3,
          24,
        ],
        [
          `
        test.only('foo', () => {
          if ('bar') {}
        })
     \u0020`,
          3,
          11,
          3,
          24,
        ],
        [
          `
        xtest('foo', () => {
          if ('bar') {}
        })
     \u0020`,
          3,
          11,
          3,
          24,
        ],
        [
          `
        describe('foo', () => {
          it('bar', () => {
            if ('bar') {}
          })
        })
     \u0020`,
          4,
          13,
          4,
          26,
        ],
      ] as const
    ).map(([code, line, column, endLine, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index + 24}`,
      code,
      errors: [
        {
          messageId: 'conditionalInTest',
          message: 'Avoid having conditionals in tests',
          line,
          column,
          endLine,
          endColumn,
        },
      ],
    })),
    {
      name: 'run 0 invalid 37',
      code: `
        describe('foo', () => {
          it('bar', () => {
            if ('bar') {}
          })
          it('baz', () => {
            if ('qux') {}
            if ('quux') {}
          })
        })
     \u0020`,
      errors: [
        {
          messageId: 'conditionalInTest',
          message: 'Avoid having conditionals in tests',
          line: 4,
          column: 13,
          endLine: 4,
          endColumn: 26,
        },
        {
          messageId: 'conditionalInTest',
          message: 'Avoid having conditionals in tests',
          line: 7,
          column: 13,
          endLine: 7,
          endColumn: 26,
        },
        {
          messageId: 'conditionalInTest',
          message: 'Avoid having conditionals in tests',
          line: 8,
          column: 13,
          endLine: 8,
          endColumn: 27,
        },
      ],
    },
    ...(
      [
        `
        it('foo', () => {
          callExpression()
          if ('bar') {}
        })
     \u0020`,
        `
        it.each\`\`('foo', () => {
          callExpression()
          if ('bar') {}
        })
     \u0020`,
        `
        it.each()('foo', () => {
          callExpression()
          if ('bar') {}
        })
     \u0020`,
        `
        it.only.each\`\`('foo', () => {
          callExpression()
          if ('bar') {}
        })
     \u0020`,
        `
        it.only.each()('foo', () => {
          callExpression()
          if ('bar') {}
        })
     \u0020`,
      ] as const
    ).map((code, index): TestCase => ({
      name: `run 0 invalid ${index + 38}`,
      code,
      errors: [
        {
          messageId: 'conditionalInTest',
          message: 'Avoid having conditionals in tests',
          line: 4,
          column: 11,
          endLine: 4,
          endColumn: 24,
        },
      ],
    })),
    ...(
      [
        [
          `
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
     \u0020`,
          6,
          17,
          10,
          13,
          15,
          15,
        ],
        [
          `
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
     \u0020`,
          3,
          11,
          5,
          10,
          11,
          12,
        ],
      ] as const
    ).map(
      (
        [code, line1, column1, endLine1, line2, column2, endLine2],
        index,
      ): TestCase => ({
        name: `run 0 invalid ${index + 43}`,
        code,
        errors: [
          {
            messageId: 'conditionalInTest',
            message: 'Avoid having conditionals in tests',
            line: line1,
            column: column1,
            endLine: endLine1,
            endColumn: column1 + 1,
          },
          {
            messageId: 'conditionalInTest',
            message: 'Avoid having conditionals in tests',
            line: line2,
            column: column2,
            endLine: endLine2,
            endColumn: column2 + 1,
          },
        ],
      }),
    ),
    ...(
      [
        [
          `
        it('foo', () => {
          const value = obj?.bar;
        })
     \u0020`,
          3,
          25,
          33,
        ],
        [
          `
        it('foo', () => {
          obj?.foo?.bar;
        })
     \u0020`,
          3,
          11,
          24,
        ],
        [
          `
        it('foo', () => {
          obj?.foo();
        })
     \u0020`,
          3,
          11,
          21,
        ],
        [
          `
        it('foo', () => {
          obj?.[key];
        })
     \u0020`,
          3,
          11,
          21,
        ],
        [
          `
        it.skip('foo', () => {
          obj?.bar;
        })
     \u0020`,
          3,
          11,
          19,
        ],
        [
          `
        it.only('foo', () => {
          obj?.bar;
        })
     \u0020`,
          3,
          11,
          19,
        ],
        [
          `
        test('foo', () => {
          obj?.bar;
        })
     \u0020`,
          3,
          11,
          19,
        ],
        [
          `
        xtest('foo', () => {
          obj?.bar;
        })
     \u0020`,
          3,
          11,
          19,
        ],
        [
          `
        fit('foo', () => {
          obj?.bar;
        })
     \u0020`,
          3,
          11,
          19,
        ],
        [
          `
        xit('foo', () => {
          obj?.bar;
        })
     \u0020`,
          3,
          11,
          19,
        ],
        [
          `
        describe('foo', () => {
          it('bar', () => {
            obj?.bar;
          })
        })
     \u0020`,
          4,
          13,
          21,
        ],
        [
          `
        it('is invalid', () => {
          const values = something.map(thing => thing?.foo);

          expect(values).toStrictEqual(['foo']);
        });
     \u0020`,
          3,
          49,
          59,
        ],
      ] as const
    ).map(([code, line, column, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index + 45}`,
      code,
      options: [
        {
          allowOptionalChaining: false,
        },
      ],
      errors: [
        {
          messageId: 'conditionalInTest',
          message: 'Avoid having conditionals in tests',
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
          `it('foo', () => {
  expect(bar ? foo : baz).toBe(boo);
})`,
          2,
          10,
          25,
        ],
        [
          `it('foo', function () {
  const foo = function (bar) {
    return foo ? bar : null;
  };
});`,
          3,
          12,
          28,
        ],
        [
          `it('foo', () => {
  const foo = bar ? foo : baz;
})`,
          2,
          15,
          30,
        ],
        [
          `it('foo', () => {
  const foo = bar ? foo : baz;
})
const foo = bar ? foo : baz;`,
          2,
          15,
          30,
        ],
      ] as const
    ).map(([code, line, column, endColumn], index): TestCase => ({
      name: `upstream conditional expressions / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'conditionalInTest',
          message: 'Avoid having conditionals in tests',
          line,
          column,
          endLine: line,
          endColumn,
        },
      ],
    })),
    {
      name: 'upstream conditional expressions / invalid 4',
      code: `it('foo', () => {
  const foo = bar ? foo : baz;
  const anotherFoo = anotherBar ? anotherFoo : anotherBaz;
})`,
      languageOptions: {
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'conditionalInTest',
          message: 'Avoid having conditionals in tests',
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 30,
        },
        {
          messageId: 'conditionalInTest',
          message: 'Avoid having conditionals in tests',
          line: 3,
          column: 22,
          endLine: 3,
          endColumn: 58,
        },
      ],
    },
    ...(
      [
        [
          `it('is invalid', () => {
  const values = something.map(thing => {
    switch (thing.isFoo) {
      case true:
        return thing.foo;
      default:
        return thing.bar;
    }
  });

  expect(values).toStrictEqual(['foo']);
});`,
          3,
          5,
          8,
          6,
        ],
        [
          `it('foo', () => {
  switch (true) {
    case true: {}
  }
})`,
          2,
          3,
          4,
          4,
        ],
        [
          `it('foo', () => {
  switch('bar') {}
})`,
          2,
          3,
          2,
          19,
        ],
        [
          `it.skip('foo', () => {
  switch('bar') {}
})`,
          2,
          3,
          2,
          19,
        ],
        [
          `it.only('foo', () => {
  switch('bar') {}
})`,
          2,
          3,
          2,
          19,
        ],
        [
          `xit('foo', () => {
  switch('bar') {}
})`,
          2,
          3,
          2,
          19,
        ],
        [
          `fit('foo', () => {
  switch('bar') {}
})`,
          2,
          3,
          2,
          19,
        ],
        [
          `test('foo', () => {
  switch('bar') {}
})`,
          2,
          3,
          2,
          19,
        ],
        [
          `test.skip('foo', () => {
  switch('bar') {}
})`,
          2,
          3,
          2,
          19,
        ],
        [
          `test.only('foo', () => {
  switch('bar') {}
})`,
          2,
          3,
          2,
          19,
        ],
        [
          `xtest('foo', () => {
  switch('bar') {}
})`,
          2,
          3,
          2,
          19,
        ],
        [
          `xtest('foo', function () {
  switch('bar') {}
})`,
          2,
          3,
          2,
          19,
        ],
        [
          `describe('foo', () => {
  it('bar', () => {

    switch('bar') {}
  })
})`,
          4,
          5,
          4,
          21,
        ],
      ] as const
    ).map(([code, line, column, endLine, endColumn], index): TestCase => ({
      name: `upstream switch statements / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'conditionalInTest',
          message: 'Avoid having conditionals in tests',
          line,
          column,
          endLine,
          endColumn,
        },
      ],
    })),
    {
      name: 'upstream switch statements / invalid 13',
      code: `describe('foo', () => {
  it('bar', () => {
    switch('bar') {}
  })
  it('baz', () => {
    switch('qux') {}
    switch('quux') {}
  })
})`,
      languageOptions: {
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'conditionalInTest',
          message: 'Avoid having conditionals in tests',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 21,
        },
        {
          messageId: 'conditionalInTest',
          message: 'Avoid having conditionals in tests',
          line: 6,
          column: 5,
          endLine: 6,
          endColumn: 21,
        },
        {
          messageId: 'conditionalInTest',
          message: 'Avoid having conditionals in tests',
          line: 7,
          column: 5,
          endLine: 7,
          endColumn: 22,
        },
      ],
    },
    {
      name: 'upstream switch statements / invalid 14',
      code: `it('foo', () => {
  callExpression()
  switch ('bar') {}
})`,
      languageOptions: {
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'conditionalInTest',
          message: 'Avoid having conditionals in tests',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 20,
        },
      ],
    },
    {
      name: 'upstream switch statements / invalid 15',
      code: `describe('valid', () => {
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
});`,
      languageOptions: {
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'conditionalInTest',
          message: 'Avoid having conditionals in tests',
          line: 5,
          column: 9,
          endLine: 10,
          endColumn: 10,
        },
        {
          messageId: 'conditionalInTest',
          message: 'Avoid having conditionals in tests',
          line: 13,
          column: 7,
          endLine: 16,
          endColumn: 8,
        },
      ],
    },
    ...(
      [
        [
          `it('foo', () => {
  const foo = function(bar) {
    if (bar) {
      return 1;
    } else {
      return 2;
    }
  };
});`,
          3,
          5,
          7,
          6,
        ],
        [
          `it('foo', () => {
  function foo(bar) {
    if (bar) {
      return 1;
    } else {
      return 2;
    }
  };
});`,
          3,
          5,
          7,
          6,
        ],
        [
          `it('foo', () => {
  if ('bar') {}
})`,
          2,
          3,
          2,
          16,
        ],
        [
          `it.skip('foo', () => {
  if ('bar') {}
})`,
          2,
          3,
          2,
          16,
        ],
        [
          `it.skip('foo', function () {
  if ('bar') {}
})`,
          2,
          3,
          2,
          16,
        ],
        [
          `it.only('foo', () => {
  if ('bar') {}
})`,
          2,
          3,
          2,
          16,
        ],
        [
          `xit('foo', () => {
  if ('bar') {}
})`,
          2,
          3,
          2,
          16,
        ],
        [
          `fit('foo', () => {
  if ('bar') {}
})`,
          2,
          3,
          2,
          16,
        ],
        [
          `test('foo', () => {
  if ('bar') {}
})`,
          2,
          3,
          2,
          16,
        ],
        [
          `test.skip('foo', () => {
  if ('bar') {}
})`,
          2,
          3,
          2,
          16,
        ],
        [
          `test.only('foo', () => {
  if ('bar') {}
})`,
          2,
          3,
          2,
          16,
        ],
        [
          `xtest('foo', () => {
  if ('bar') {}
})`,
          2,
          3,
          2,
          16,
        ],
        [
          `describe('foo', () => {
  it('bar', () => {
    if ('bar') {}
  })
})`,
          3,
          5,
          3,
          18,
        ],
      ] as const
    ).map(([code, line, column, endLine, endColumn], index): TestCase => ({
      name: `upstream if statements / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'conditionalInTest',
          message: 'Avoid having conditionals in tests',
          line,
          column,
          endLine,
          endColumn,
        },
      ],
    })),
    {
      name: 'upstream if statements / invalid 13',
      code: `describe('foo', () => {
  it('bar', () => {
    if ('bar') {}
  })
  it('baz', () => {
    if ('qux') {}
    if ('quux') {}
  })
})`,
      languageOptions: {
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'conditionalInTest',
          message: 'Avoid having conditionals in tests',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 18,
        },
        {
          messageId: 'conditionalInTest',
          message: 'Avoid having conditionals in tests',
          line: 6,
          column: 5,
          endLine: 6,
          endColumn: 18,
        },
        {
          messageId: 'conditionalInTest',
          message: 'Avoid having conditionals in tests',
          line: 7,
          column: 5,
          endLine: 7,
          endColumn: 19,
        },
      ],
    },
    ...(
      [
        `it('foo', () => {
  callExpression()
  if ('bar') {}
})`,
        `it.each\`\`('foo', () => {
  callExpression()
  if ('bar') {}
})`,
        `it.each()('foo', () => {
  callExpression()
  if ('bar') {}
})`,
        `it.only.each\`\`('foo', () => {
  callExpression()
  if ('bar') {}
})`,
        `it.only.each()('foo', () => {
  callExpression()
  if ('bar') {}
})`,
      ] as const
    ).map((code, index): TestCase => ({
      name: `upstream if statements / invalid ${index + 14}`,
      code,
      languageOptions: {
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'conditionalInTest',
          message: 'Avoid having conditionals in tests',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 16,
        },
      ],
    })),
    ...(
      [
        [
          `describe('valid', () => {
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
});`,
          5,
          9,
          9,
          12,
          7,
          14,
        ],
        [
          `test("shows error", () => {
  if (1 === 2) {
    expect(true).toBe(false);
  }
});

test("does not show error", () => {
  setTimeout(() => console.log("noop"));
  if (1 === 2) {
    expect(true).toBe(false);
  }
});`,
          2,
          3,
          4,
          9,
          3,
          11,
        ],
      ] as const
    ).map(
      (
        [code, line1, column1, endLine1, line2, column2, endLine2],
        index,
      ): TestCase => ({
        name: `upstream if statements / invalid ${index + 19}`,
        code,
        languageOptions: {
          ecmaVersion: 2020,
        },
        errors: [
          {
            messageId: 'conditionalInTest',
            message: 'Avoid having conditionals in tests',
            line: line1,
            column: column1,
            endLine: endLine1,
            endColumn: column1 + 1,
          },
          {
            messageId: 'conditionalInTest',
            message: 'Avoid having conditionals in tests',
            line: line2,
            column: column2,
            endLine: endLine2,
            endColumn: column2 + 1,
          },
        ],
      }),
    ),
    ...(
      [
        [
          `it('foo', () => {
  const value = obj?.bar;
})`,
          2,
          17,
          25,
        ],
        [
          `it('foo', () => {
  obj?.foo?.bar;
})`,
          2,
          3,
          16,
        ],
        [
          `it('foo', () => {
  obj?.foo();
})`,
          2,
          3,
          13,
        ],
        [
          `it('foo', () => {
  obj?.[key];
})`,
          2,
          3,
          13,
        ],
        [
          `it.skip('foo', () => {
  obj?.bar;
})`,
          2,
          3,
          11,
        ],
        [
          `it.only('foo', () => {
  obj?.bar;
})`,
          2,
          3,
          11,
        ],
        [
          `test('foo', () => {
  obj?.bar;
})`,
          2,
          3,
          11,
        ],
        [
          `xtest('foo', () => {
  obj?.bar;
})`,
          2,
          3,
          11,
        ],
        [
          `fit('foo', () => {
  obj?.bar;
})`,
          2,
          3,
          11,
        ],
        [
          `xit('foo', () => {
  obj?.bar;
})`,
          2,
          3,
          11,
        ],
        [
          `describe('foo', () => {
  it('bar', () => {
    obj?.bar;
  })
})`,
          3,
          5,
          13,
        ],
        [
          `it('is invalid', () => {
  const values = something.map(thing => thing?.foo);

  expect(values).toStrictEqual(['foo']);
});`,
          2,
          41,
          51,
        ],
      ] as const
    ).map(([code, line, column, endColumn], index): TestCase => ({
      name: `upstream optional chaining with allowOptionalChaining=false / invalid ${index}`,
      code,
      options: [
        {
          allowOptionalChaining: false,
        },
      ],
      languageOptions: {
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'conditionalInTest',
          message: 'Avoid having conditionals in tests',
          line,
          column,
          endLine: line,
          endColumn,
        },
      ],
    })),
  ],
});
