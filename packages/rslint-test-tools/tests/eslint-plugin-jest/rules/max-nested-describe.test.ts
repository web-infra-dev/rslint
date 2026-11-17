// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('max-nested-describe', {
  valid: [
    {
      name: 'run 0 valid 0',
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
   \u0020`,
    },
    {
      name: 'run 0 valid 1',
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
   \u0020`,
    },
    {
      name: 'run 0 valid 2',
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
   \u0020`,
      options: [
        {
          max: 3,
        },
      ],
    },
    {
      name: 'run 0 valid 3',
      code: `
        it('something', async () => {
          expect('something').toBe('something');
        });
   \u0020`,
      options: [
        {
          max: 0,
        },
      ],
    },
    {
      name: 'run 0 valid 4',
      code: `
      describe('foo', () => {
        describe.each(['hello', 'world'])("%s", (a) => {});
      });
   \u0020`,
    },
    {
      name: 'run 0 valid 5',
      code: `
      describe('foo', () => {
        describe.each\`
        foo  | bar
        1 | 2
        \`('$foo $bar', ({ foo, bar }) => {});
      });
   \u0020`,
    },
    ...(
      [
        `describe('foo', function() {
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
});`,
        `describe('foo', function() {
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
});`,
        `describe('foo', () => {
  describe('bar', () => {
    it('hello', async () => {
      expect('hello').toBe('hello');
    });
  });
});

xdescribe('foo', function() {
  describe('bar', function() {
    it('something', async () => {
      expect('something').toBe('something');
    });
  });
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream max-nested-describe / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
    })),
    {
      name: 'upstream max-nested-describe / valid 3',
      code: `describe('foo', () => {
  describe.only('bar', () => {
    describe.skip('baz', () => {
      it('something', async () => {
        expect('something').toBe('something');
      });
    });
  });
});`,
      options: [
        {
          max: 3,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
    },
    {
      name: 'upstream max-nested-describe / valid 4',
      code: `it('something', async () => {
  expect('something').toBe('something');
});`,
      options: [
        {
          max: 0,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
    },
    ...(
      [
        `describe('foo', () => {
  describe.each(['hello', 'world'])("%s", (a) => {});
});`,
        `describe('foo', () => {
  describe.each\`
  foo  | bar
  1 | 2
  \`('$foo $bar', ({ foo, bar }) => {});
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream max-nested-describe / valid ${index + 5}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
    })),
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
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
     \u0020`,
      errors: [
        {
          messageId: 'exceededMaxDepth',
          message: 'Too many nested describe calls (6) - maximum allowed is 5',
          line: 7,
          column: 19,
          endLine: 11,
          endColumn: 21,
        },
      ],
    },
    {
      name: 'run 0 invalid 1',
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
     \u0020`,
      errors: [
        {
          messageId: 'exceededMaxDepth',
          message: 'Too many nested describe calls (6) - maximum allowed is 5',
          line: 7,
          column: 19,
          endLine: 11,
          endColumn: 21,
        },
        {
          messageId: 'exceededMaxDepth',
          message: 'Too many nested describe calls (6) - maximum allowed is 5',
          line: 13,
          column: 19,
          endLine: 17,
          endColumn: 21,
        },
      ],
    },
    {
      name: 'run 0 invalid 2',
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
     \u0020`,
      options: [
        {
          max: 2,
        },
      ],
      errors: [
        {
          messageId: 'exceededMaxDepth',
          message: 'Too many nested describe calls (3) - maximum allowed is 2',
          line: 4,
          column: 13,
          endLine: 8,
          endColumn: 15,
        },
        {
          messageId: 'exceededMaxDepth',
          message: 'Too many nested describe calls (3) - maximum allowed is 2',
          line: 10,
          column: 13,
          endLine: 14,
          endColumn: 15,
        },
      ],
    },
    {
      name: 'run 0 invalid 3',
      code: `
        describe('qux', () => {
          it('should get something', () => {
            expect(getSomething()).toBe('Something');
          });
        });
     \u0020`,
      options: [
        {
          max: 0,
        },
      ],
      errors: [
        {
          messageId: 'exceededMaxDepth',
          message: 'Too many nested describe calls (1) - maximum allowed is 0',
          line: 2,
          column: 9,
          endLine: 6,
          endColumn: 11,
        },
      ],
    },
    ...(
      [
        [
          `
        describe('foo', () => {
          describe.each(['hello', 'world'])("%s", (a) => {});
        });
     \u0020`,
          3,
          61,
        ],
        [
          `
        describe('foo', () => {
          describe.each\`
          foo  | bar
          1 | 2
          \`('$foo $bar', ({ foo, bar }) => {});
        });
     \u0020`,
          6,
          47,
        ],
      ] as const
    ).map(([code, endLine, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index + 4}`,
      code,
      options: [
        {
          max: 1,
        },
      ],
      errors: [
        {
          messageId: 'exceededMaxDepth',
          message: 'Too many nested describe calls (2) - maximum allowed is 1',
          line: 3,
          column: 11,
          endLine,
          endColumn,
        },
      ],
    })),
    {
      name: 'upstream max-nested-describe / invalid 0',
      code: `describe('foo', function() {
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
});`,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'exceededMaxDepth',
          message: 'Too many nested describe calls (6) - maximum allowed is 5',
          line: 6,
          column: 11,
          endLine: 10,
          endColumn: 13,
        },
      ],
    },
    {
      name: 'upstream max-nested-describe / invalid 1',
      code: `describe('foo', () => {
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
});`,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'exceededMaxDepth',
          message: 'Too many nested describe calls (6) - maximum allowed is 5',
          line: 6,
          column: 11,
          endLine: 10,
          endColumn: 13,
        },
        {
          messageId: 'exceededMaxDepth',
          message: 'Too many nested describe calls (6) - maximum allowed is 5',
          line: 12,
          column: 11,
          endLine: 16,
          endColumn: 13,
        },
      ],
    },
    {
      name: 'upstream max-nested-describe / invalid 2',
      code: `fdescribe('foo', () => {
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
});`,
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
          messageId: 'exceededMaxDepth',
          message: 'Too many nested describe calls (3) - maximum allowed is 2',
          line: 3,
          column: 5,
          endLine: 7,
          endColumn: 7,
        },
        {
          messageId: 'exceededMaxDepth',
          message: 'Too many nested describe calls (3) - maximum allowed is 2',
          line: 9,
          column: 5,
          endLine: 13,
          endColumn: 7,
        },
      ],
    },
    {
      name: 'upstream max-nested-describe / invalid 3',
      code: `describe('qux', () => {
  it('should get something', () => {
    expect(getSomething()).toBe('Something');
  });
});`,
      options: [
        {
          max: 0,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'exceededMaxDepth',
          message: 'Too many nested describe calls (1) - maximum allowed is 0',
          line: 1,
          column: 1,
          endLine: 5,
          endColumn: 3,
        },
      ],
    },
    ...(
      [
        [
          `describe('foo', () => {
  describe.each(['hello', 'world'])("%s", (a) => {});
});`,
          2,
          53,
        ],
        [
          `describe('foo', () => {
  describe.each\`
  foo  | bar
  1 | 2
  \`('$foo $bar', ({ foo, bar }) => {});
});`,
          5,
          39,
        ],
      ] as const
    ).map(([code, endLine, endColumn], index): TestCase => ({
      name: `upstream max-nested-describe / invalid ${index + 4}`,
      code,
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
          messageId: 'exceededMaxDepth',
          message: 'Too many nested describe calls (2) - maximum allowed is 1',
          line: 2,
          column: 3,
          endLine,
          endColumn,
        },
      ],
    })),
  ],
});
