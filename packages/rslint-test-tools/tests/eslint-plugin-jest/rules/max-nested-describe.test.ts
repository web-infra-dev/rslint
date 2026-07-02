import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('max-nested-describe', {} as never, {
  valid: [
    {
      code: `
      describe('foo', function() {
        describe('bar', function () {
          describe('baz', function () {
            describe('qux', function () {
              describe('qux', function () {
                it('should get something', () => {
                  expect(getSomething()).toBe('Something');
                });
              })
            })
          })
        })
      });
    `,
    },
    {
      code: `
      describe('foo', function() {
        describe('bar', function () {
          describe('baz', function () {
            describe('qux', function () {
              describe('qux', function () {
                it('should get something', () => {
                  expect(getSomething()).toBe('Something');
                });
              });

              fdescribe('qux', () => {
                it('something', async () => {
                  expect('something').toBe('something');
                });
              });
            })
          })
        })
      });
    `,
    },
    {
      code: `
        describe('foo', () => {
          describe.only('bar', () => {
            describe.skip('baz', () => {
              it('something', async () => {
                expect('something').toBe('something');
              });
            });
          });
        });
    `,
      options: [{ max: 3 }],
    },
    {
      code: `
        it('something', async () => {
          expect('something').toBe('something');
        });
    `,
      options: [{ max: 0 }],
    },
    {
      code: `
      describe('foo', () => {
        describe.each(['hello', 'world'])("%s", (a) => {});
      });
    `,
    },
    {
      code: `
      describe('foo', () => {
        describe.each\`
        foo  | bar
        ${'1'} | ${'2'}
        \`('$foo $bar', ({ foo, bar }) => {});
      });
    `,
    },
  ],
  invalid: [
    {
      code: `
        describe('foo', function() {
          describe('bar', function () {
            describe('baz', function () {
              describe('qux', function () {
                describe('quxx', function () {
                  describe('over limit', function () {
                    it('should get something', () => {
                      expect(getSomething()).toBe('Something');
                    });
                  });
                });
              });
            });
          });
        });
      `,
      errors: [{ messageId: 'exceededMaxDepth', line: 6, column: 11 }],
    },
    {
      code: `
        describe('foo', () => {
          describe('bar', () => {
            describe('baz', () => {
              describe('baz1', () => {
                describe('baz2', () => {
                  describe('baz3', () => {
                    it('should get something', () => {
                      expect(getSomething()).toBe('Something');
                    });
                  });

                  describe('baz4', () => {
                    it('should get something', () => {
                      expect(getSomething()).toBe('Something');
                    });
                  });
                });
              });
            });

            describe('qux', function () {
              it('should get something', () => {
                expect(getSomething()).toBe('Something');
              });
            });
          })
        });
      `,
      errors: [
        { messageId: 'exceededMaxDepth', line: 6, column: 11 },
        { messageId: 'exceededMaxDepth', line: 12, column: 11 },
      ],
    },
    {
      code: `
        fdescribe('foo', () => {
          describe.only('bar', () => {
            describe.skip('baz', () => {
              it('should get something', () => {
                expect(getSomething()).toBe('Something');
              });
            });

            describe('baz', () => {
              it('should get something', () => {
                expect(getSomething()).toBe('Something');
              });
            });
          });
        });

        xdescribe('qux', () => {
          it('should get something', () => {
            expect(getSomething()).toBe('Something');
          });
        });
      `,
      options: [{ max: 2 }],
      errors: [
        { messageId: 'exceededMaxDepth', line: 3, column: 5 },
        { messageId: 'exceededMaxDepth', line: 9, column: 5 },
      ],
    },
    {
      code: `
        describe('qux', () => {
          it('should get something', () => {
            expect(getSomething()).toBe('Something');
          });
        });
      `,
      options: [{ max: 0 }],
      errors: [{ messageId: 'exceededMaxDepth', line: 1, column: 1 }],
    },
    {
      code: `
        describe('foo', () => {
          describe.each(['hello', 'world'])("%s", (a) => {});
        });
      `,
      options: [{ max: 1 }],
      errors: [{ messageId: 'exceededMaxDepth', line: 2, column: 3 }],
    },
    {
      code: `
        describe('foo', () => {
          describe.each\`
          foo  | bar
          ${'1'} | ${'2'}
          \`('$foo $bar', ({ foo, bar }) => {});
        });
      `,
      options: [{ max: 1 }],
      errors: [{ messageId: 'exceededMaxDepth', line: 2, column: 3 }],
    },
  ],
});
