import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-conditional-in-test', {} as never, {
  valid: [
    { code: 'const x = y ? 1 : 0' },
    { code: 'const x = foo && bar' },
    { code: 'const x = foo || bar' },
    { code: 'const x = foo ?? bar' },
    {
      code: `
      const foo = function (bar) {
        return foo ? bar : null;
      };

      it('foo', () => {
        foo();
      });
    `,
    },
    {
      code: `
      const foo = function (bar) {
        return foo ? bar : null;
      };

      it.each()('foo', function () {
        foo();
      });
    `,
    },
    {
      code: `
      fit.concurrent('foo', () => {
        switch('bar') {}
      })
    `,
    },
    { code: `it('foo', () => {})` },
    {
      code: `
      switch (true) {
        case true: {}
      }
    `,
    },
    {
      code: `
      it('foo', () => {});
      function myTest() {
        switch ('bar') {
        }
      }
    `,
    },
    {
      code: `
      foo('bar', () => {
        switch(baz) {}
      })
    `,
    },
    {
      code: `
      describe('foo', () => {
        switch('bar') {}
      })
    `,
    },
    {
      code: `
      describe.skip('foo', () => {
        switch('bar') {}
      })
    `,
    },
    {
      code: `
      describe.skip.each()('foo', () => {
        switch('bar') {}
      })
    `,
    },
    {
      code: `
      xdescribe('foo', () => {
        switch('bar') {}
      })
    `,
    },
    {
      code: `
      fdescribe('foo', () => {
        switch('bar') {}
      })
    `,
    },
    {
      code: `
      describe('foo', () => {
        switch('bar') {}
      })
      switch('bar') {}
    `,
    },
    {
      code: `
      describe('foo', () => {
        afterEach(() => {
          switch('bar') {}
        });
      });
    `,
    },
    {
      code: `
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
    `,
    },
    {
      code: `
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
    `,
    },
    { code: 'if (foo) {}' },
    { code: "it('foo', () => {})" },
    { code: 'it("foo", function () {})' },
    { code: "it('foo', () => {}); function myTest() { if ('bar') {} }" },
    {
      code: `
      foo('bar', () => {
        if (baz) {}
      })
    `,
    },
    {
      code: `
      describe('foo', () => {
        if ('bar') {}
      })
    `,
    },
    {
      code: `
      describe.skip('foo', () => {
        if ('bar') {}
      })
    `,
    },
    {
      code: `
      xdescribe('foo', () => {
        if ('bar') {}
      })
    `,
    },
    {
      code: `
      fdescribe('foo', () => {
        if ('bar') {}
      })
    `,
    },
    {
      code: `
      describe('foo', () => {
        if ('bar') {}
      })
      if ('baz') {}
    `,
    },
    {
      code: `
      describe('foo', () => {
        afterEach(() => {
          if ('bar') {}
        });
      })
    `,
    },
    {
      code: `
      describe.each\`\`('foo', () => {
        afterEach(() => {
          if ('bar') {}
        });
      })
    `,
    },
    {
      code: `
      describe('foo', () => {
        beforeEach(() => {
          if ('bar') {}
        });
      })
    `,
    },
    { code: 'const foo = bar ? foo : baz;' },
    {
      code: `
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
    `,
    },
    {
      code: `
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
    `,
    },
    {
      code: `
      fit.concurrent('foo', () => {
        if ('bar') {}
      })
    `,
    },
    { code: 'const x = obj?.foo' },
    {
      code: `
      it('foo', () => {
        const value = obj?.bar;
      })
    `,
    },
    {
      code: `
      it('foo', () => {
        obj?.foo?.bar;
      })
    `,
    },
    {
      code: `
      it('foo', () => {
        obj?.foo();
      })
    `,
    },
    {
      code: `
      it('foo', () => {
        obj?.[key];
      })
    `,
    },
    {
      code: `
      test('foo', () => {
        obj?.bar;
      })
    `,
    },
    {
      code: `
      it('is valid', () => {
        const values = something.map(thing => thing?.foo);

        expect(values).toStrictEqual(['foo']);
      });
    `,
    },
    {
      code: 'const x = obj?.foo',
      options: [{ allowOptionalChaining: false }],
    },
    {
      code: `
        const foo = obj?.bar;

        it('foo', () => {
          expect(foo).toBe(undefined);
        });
      `,
      options: [{ allowOptionalChaining: false }],
    },
    {
      code: `
        describe('foo', () => {
          const val = obj?.bar;
        })
      `,
      options: [{ allowOptionalChaining: false }],
    },
    {
      code: `
        describe('foo', () => {
          beforeEach(() => {
            const val = obj?.bar;
          });
        })
      `,
      options: [{ allowOptionalChaining: false }],
    },
    {
      code: `
        describe('foo', () => {
          afterEach(() => {
            const val = obj?.bar;
          });
        })
      `,
      options: [{ allowOptionalChaining: false }],
    },
    {
      code: `
        const values = something.map(thing => thing?.foo);

        it('valid', () => {
          expect(values).toStrictEqual(['foo']);
        });
      `,
      options: [{ allowOptionalChaining: false }],
    },
    {
      code: `
        describe('valid', () => {
          const values = something.map(thing => thing?.foo);
          it('still valid', () => {
            expect(values).toStrictEqual(['foo']);
          });
        });
      `,
      options: [{ allowOptionalChaining: false }],
    },
  ],
  invalid: [
    {
      code: `
        it('foo', () => {
          expect(bar ? foo : baz).toBe(boo);
        })
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 10,
          line: 2,
        },
      ],
    },
    {
      code: `
        it('foo', () => {
          foo && expect(foo).toBe(true);
        })
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 3,
          line: 2,
        },
      ],
    },
    {
      code: `
        it('foo', () => {
          const value = foo || bar;
        })
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 17,
          line: 2,
        },
      ],
    },
    {
      code: `
        it('foo', () => {
          const value = foo ?? bar;
        })
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 17,
          line: 2,
        },
      ],
    },
    {
      code: `
        it('foo', function () {
          const foo = function (bar) {
            return foo ? bar : null;
          };
        });
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 12,
          line: 3,
        },
      ],
    },
    {
      code: `
        it('foo', () => {
          const foo = bar ? foo : baz;
        })
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 15,
          line: 2,
        },
      ],
    },
    {
      code: `
        it('foo', () => {
          const foo = bar ? foo : baz;
        })
        const foo = bar ? foo : baz;
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 15,
          line: 2,
        },
      ],
    },
    {
      code: `
        it('foo', () => {
          const foo = bar ? foo : baz;
          const anotherFoo = anotherBar ? anotherFoo : anotherBaz;
        })
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 15,
          line: 2,
        },
        {
          messageId: 'conditionalInTest',
          column: 22,
          line: 3,
        },
      ],
    },
    {
      code: `
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
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 5,
          line: 3,
        },
      ],
    },
    {
      code: `
        it('foo', () => {
          switch (true) {
            case true: {}
          }
        })
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 3,
          line: 2,
        },
      ],
    },
    {
      code: `
        it('foo', () => {
          switch('bar') {}
        })
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 3,
          line: 2,
        },
      ],
    },
    {
      code: `
        it.skip('foo', () => {
          switch('bar') {}
        })
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 3,
          line: 2,
        },
      ],
    },
    {
      code: `
        it.only('foo', () => {
          switch('bar') {}
        })
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 3,
          line: 2,
        },
      ],
    },
    {
      code: `
        xit('foo', () => {
          switch('bar') {}
        })
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 3,
          line: 2,
        },
      ],
    },
    {
      code: `
        fit('foo', () => {
          switch('bar') {}
        })
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 3,
          line: 2,
        },
      ],
    },
    {
      code: `
        test('foo', () => {
          switch('bar') {}
        })
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 3,
          line: 2,
        },
      ],
    },
    {
      code: `
        test.skip('foo', () => {
          switch('bar') {}
        })
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 3,
          line: 2,
        },
      ],
    },
    {
      code: `
        test.only('foo', () => {
          switch('bar') {}
        })
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 3,
          line: 2,
        },
      ],
    },
    {
      code: `
        xtest('foo', () => {
          switch('bar') {}
        })
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 3,
          line: 2,
        },
      ],
    },
    {
      code: `
        xtest('foo', function () {
          switch('bar') {}
        })
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 3,
          line: 2,
        },
      ],
    },
    {
      code: `
        describe('foo', () => {
          it('bar', () => {

            switch('bar') {}
          })
        })
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 5,
          line: 4,
        },
      ],
    },
    {
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
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 5,
          line: 3,
        },
        {
          messageId: 'conditionalInTest',
          column: 5,
          line: 6,
        },
        {
          messageId: 'conditionalInTest',
          column: 5,
          line: 7,
        },
      ],
    },
    {
      code: `
        it('foo', () => {
          callExpression()
          switch ('bar') {}
        })
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 3,
          line: 3,
        },
      ],
    },
    {
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
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 9,
          line: 5,
        },
        {
          messageId: 'conditionalInTest',
          column: 7,
          line: 13,
        },
      ],
    },
    {
      code: `
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
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 5,
          line: 3,
        },
      ],
    },
    {
      code: `
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
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 5,
          line: 3,
        },
      ],
    },
    {
      code: `
        it('foo', () => {
          if ('bar') {}
        })
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 3,
          line: 2,
        },
      ],
    },
    {
      code: `
        it.skip('foo', () => {
          if ('bar') {}
        })
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 3,
          line: 2,
        },
      ],
    },
    {
      code: `
        it.skip('foo', function () {
          if ('bar') {}
        })
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 3,
          line: 2,
        },
      ],
    },
    {
      code: `
        it.only('foo', () => {
          if ('bar') {}
        })
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 3,
          line: 2,
        },
      ],
    },
    {
      code: `
        xit('foo', () => {
          if ('bar') {}
        })
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 3,
          line: 2,
        },
      ],
    },
    {
      code: `
        fit('foo', () => {
          if ('bar') {}
        })
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 3,
          line: 2,
        },
      ],
    },
    {
      code: `
        test('foo', () => {
          if ('bar') {}
        })
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 3,
          line: 2,
        },
      ],
    },
    {
      code: `
        test.skip('foo', () => {
          if ('bar') {}
        })
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 3,
          line: 2,
        },
      ],
    },
    {
      code: `
        test.only('foo', () => {
          if ('bar') {}
        })
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 3,
          line: 2,
        },
      ],
    },
    {
      code: `
        xtest('foo', () => {
          if ('bar') {}
        })
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 3,
          line: 2,
        },
      ],
    },
    {
      code: `
        describe('foo', () => {
          it('bar', () => {
            if ('bar') {}
          })
        })
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 5,
          line: 3,
        },
      ],
    },
    {
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
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 5,
          line: 3,
        },
        {
          messageId: 'conditionalInTest',
          column: 5,
          line: 6,
        },
        {
          messageId: 'conditionalInTest',
          column: 5,
          line: 7,
        },
      ],
    },
    {
      code: `
        it('foo', () => {
          callExpression()
          if ('bar') {}
        })
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 3,
          line: 3,
        },
      ],
    },
    {
      code: `
        it.each\`\`('foo', () => {
          callExpression()
          if ('bar') {}
        })
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 3,
          line: 3,
        },
      ],
    },
    {
      code: `
        it.each()('foo', () => {
          callExpression()
          if ('bar') {}
        })
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 3,
          line: 3,
        },
      ],
    },
    {
      code: `
        it.only.each\`\`('foo', () => {
          callExpression()
          if ('bar') {}
        })
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 3,
          line: 3,
        },
      ],
    },
    {
      code: `
        it.only.each()('foo', () => {
          callExpression()
          if ('bar') {}
        })
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 3,
          line: 3,
        },
      ],
    },
    {
      code: `
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
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 9,
          line: 5,
        },
        {
          messageId: 'conditionalInTest',
          column: 7,
          line: 12,
        },
      ],
    },
    {
      code: `
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
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 3,
          line: 2,
        },
        {
          messageId: 'conditionalInTest',
          column: 3,
          line: 9,
        },
      ],
    },
    {
      code: `
        it('foo', () => {
          const value = obj?.bar;
        })
      `,
      options: [{ allowOptionalChaining: false }],
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 17,
          line: 2,
        },
      ],
    },
    {
      code: `
        it('foo', () => {
          obj?.foo?.bar;
        })
      `,
      options: [{ allowOptionalChaining: false }],
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 3,
          line: 2,
        },
      ],
    },
    {
      code: `
        it('foo', () => {
          obj?.foo();
        })
      `,
      options: [{ allowOptionalChaining: false }],
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 3,
          line: 2,
        },
      ],
    },
    {
      code: `
        it('foo', () => {
          obj?.[key];
        })
      `,
      options: [{ allowOptionalChaining: false }],
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 3,
          line: 2,
        },
      ],
    },
    {
      code: `
        it.skip('foo', () => {
          obj?.bar;
        })
      `,
      options: [{ allowOptionalChaining: false }],
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 3,
          line: 2,
        },
      ],
    },
    {
      code: `
        it.only('foo', () => {
          obj?.bar;
        })
      `,
      options: [{ allowOptionalChaining: false }],
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 3,
          line: 2,
        },
      ],
    },
    {
      code: `
        test('foo', () => {
          obj?.bar;
        })
      `,
      options: [{ allowOptionalChaining: false }],
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 3,
          line: 2,
        },
      ],
    },
    {
      code: `
        xtest('foo', () => {
          obj?.bar;
        })
      `,
      options: [{ allowOptionalChaining: false }],
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 3,
          line: 2,
        },
      ],
    },
    {
      code: `
        fit('foo', () => {
          obj?.bar;
        })
      `,
      options: [{ allowOptionalChaining: false }],
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 3,
          line: 2,
        },
      ],
    },
    {
      code: `
        xit('foo', () => {
          obj?.bar;
        })
      `,
      options: [{ allowOptionalChaining: false }],
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 3,
          line: 2,
        },
      ],
    },
    {
      code: `
        describe('foo', () => {
          it('bar', () => {
            obj?.bar;
          })
        })
      `,
      options: [{ allowOptionalChaining: false }],
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 5,
          line: 3,
        },
      ],
    },
    {
      code: `
        it('is invalid', () => {
          const values = something.map(thing => thing?.foo);

          expect(values).toStrictEqual(['foo']);
        });
      `,
      options: [{ allowOptionalChaining: false }],
      errors: [
        {
          messageId: 'conditionalInTest',
          column: 41,
          line: 2,
        },
      ],
    },
  ],
});
