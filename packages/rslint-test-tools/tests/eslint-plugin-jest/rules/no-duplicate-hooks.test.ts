import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-duplicate-hooks', {} as never, {
  valid: [
    {
      code: `
      describe("foo", () => {
        beforeEach(() => {})
        test("bar", () => {
          someFn();
        })
      })
    `,
    },
    {
      code: `
      beforeEach(() => {})
      test("bar", () => {
        someFn();
      })
    `,
    },
    {
      code: `
      describe("foo", () => {
        beforeAll(() => {}),
        beforeEach(() => {})
        afterEach(() => {})
        afterAll(() => {})

        test("bar", () => {
          someFn();
        })
      })
    `,
    },
    // multiple describe blocks
    {
      code: `
      describe.skip("foo", () => {
        beforeEach(() => {}),
        beforeAll(() => {}),
        test("bar", () => {
          someFn();
        })
      })
      describe("foo", () => {
        beforeEach(() => {}),
        beforeAll(() => {}),
        test("bar", () => {
          someFn();
        })
      })
    `,
    },
    // nested describe blocks
    {
      code: `
      describe("foo", () => {
        beforeEach(() => {}),
        test("bar", () => {
          someFn();
        })
        describe("inner_foo", () => {
          beforeEach(() => {})
          test("inner bar", () => {
            someFn();
          })
        })
      })
    `,
    },
    // describe.each blocks
    {
      code: `
      describe.each(['hello'])('%s', () => {
        beforeEach(() => {});

        it('is fine', () => {});
      });
    `,
    },
    {
      code: `
      describe('something', () => {
        describe.each(['hello'])('%s', () => {
          beforeEach(() => {});

          it('is fine', () => {});
        });

        describe.each(['world'])('%s', () => {
          beforeEach(() => {});

          it('is fine', () => {});
        });
      });
    `,
    },
    {
      code: `
      describe.each\`\`('%s', () => {
        beforeEach(() => {});

        it('is fine', () => {});
      });
    `,
    },
    {
      code: `
      describe('something', () => {
        describe.each\`\`('%s', () => {
          beforeEach(() => {});

          it('is fine', () => {});
        });

        describe.each\`\`('%s', () => {
          beforeEach(() => {});

          it('is fine', () => {});
        });
      });
    `,
    },
  ],
  invalid: [
    {
      code: `
        describe("foo", () => {
          beforeEach(() => {}),
          beforeEach(() => {}),
          test("bar", () => {
            someFn();
          })
        })
      `,
      errors: [
        {
          messageId: 'noDuplicateHook',
          data: { hook: 'beforeEach' },
          column: 3,
          line: 3,
        },
      ],
    },
    {
      code: `
        describe.skip("foo", () => {
          beforeEach(() => {}),
          beforeAll(() => {}),
          beforeAll(() => {}),
          test("bar", () => {
            someFn();
          })
        })
      `,
      errors: [
        {
          messageId: 'noDuplicateHook',
          data: { hook: 'beforeAll' },
          column: 3,
          line: 4,
        },
      ],
    },
    {
      code: `
        describe.skip("foo", () => {
          afterEach(() => {}),
          afterEach(() => {}),
          test("bar", () => {
            someFn();
          })
        })
      `,
      errors: [
        {
          messageId: 'noDuplicateHook',
          data: { hook: 'afterEach' },
          column: 3,
          line: 3,
        },
      ],
    },
    {
      code: `
        import { afterEach } from '@jest/globals';

        describe.skip("foo", () => {
          afterEach(() => {}),
          afterEach(() => {}),
          test("bar", () => {
            someFn();
          })
        })
      `,
      errors: [
        {
          messageId: 'noDuplicateHook',
          data: { hook: 'afterEach' },
          column: 3,
          line: 5,
        },
      ],
    },
    {
      code: `
        import { afterEach, afterEach as somethingElse } from '@jest/globals';

        describe.skip("foo", () => {
          afterEach(() => {}),
          somethingElse(() => {}),
          test("bar", () => {
            someFn();
          })
        })
      `,
      errors: [
        {
          messageId: 'noDuplicateHook',
          data: { hook: 'afterEach' },
          column: 3,
          line: 5,
        },
      ],
    },
    {
      code: `
        describe.skip("foo", () => {
          afterAll(() => {}),
          afterAll(() => {}),
          test("bar", () => {
            someFn();
          })
        })
      `,
      errors: [
        {
          messageId: 'noDuplicateHook',
          data: { hook: 'afterAll' },
          column: 3,
          line: 3,
        },
      ],
    },
    {
      code: `
        afterAll(() => {}),
        afterAll(() => {}),
        test("bar", () => {
          someFn();
        })
      `,
      errors: [
        {
          messageId: 'noDuplicateHook',
          data: { hook: 'afterAll' },
          column: 1,
          line: 2,
        },
      ],
    },
    {
      code: `
        describe("foo", () => {
          beforeEach(() => {}),
          beforeEach(() => {}),
          beforeEach(() => {}),
          test("bar", () => {
            someFn();
          })
        })
      `,
      errors: [
        {
          messageId: 'noDuplicateHook',
          data: { hook: 'beforeEach' },
          column: 3,
          line: 3,
        },
        {
          messageId: 'noDuplicateHook',
          data: { hook: 'beforeEach' },
          column: 3,
          line: 4,
        },
      ],
    },
    {
      code: `
        describe.skip("foo", () => {
          afterAll(() => {}),
          afterAll(() => {}),
          beforeAll(() => {}),
          beforeAll(() => {}),
          test("bar", () => {
            someFn();
          })
        })
      `,
      errors: [
        {
          messageId: 'noDuplicateHook',
          data: { hook: 'afterAll' },
          column: 3,
          line: 3,
        },
        {
          messageId: 'noDuplicateHook',
          data: { hook: 'beforeAll' },
          column: 3,
          line: 5,
        },
      ],
    },
    // multiple describe blocks
    {
      code: `
        describe.skip("foo", () => {
          beforeEach(() => {}),
          beforeAll(() => {}),
          test("bar", () => {
            someFn();
          })
        })
        describe("foo", () => {
          beforeEach(() => {}),
          beforeEach(() => {}),
          beforeAll(() => {}),
          test("bar", () => {
            someFn();
          })
        })
      `,
      errors: [
        {
          messageId: 'noDuplicateHook',
          data: { hook: 'beforeEach' },
          column: 3,
          line: 10,
        },
      ],
    },
    // nested describe blocks
    {
      code: `
        describe("foo", () => {
          beforeAll(() => {}),
          test("bar", () => {
            someFn();
          })
          describe("inner_foo", () => {
            beforeEach(() => {})
            beforeEach(() => {})
            test("inner bar", () => {
              someFn();
            })
          })
        })
      `,
      errors: [
        {
          messageId: 'noDuplicateHook',
          data: { hook: 'beforeEach' },
          column: 5,
          line: 8,
        },
      ],
    },
    // describe.each blocks
    {
      code: `
      describe.each(['hello'])('%s', () => {
        beforeEach(() => {});
        beforeEach(() => {});

        it('is not fine', () => {});
      });
    `,
      errors: [
        {
          messageId: 'noDuplicateHook',
          data: { hook: 'beforeEach' },
          column: 3,
          line: 3,
        },
      ],
    },
    {
      code: `
      describe('something', () => {
        describe.each(['hello'])('%s', () => {
          beforeEach(() => {});

          it('is fine', () => {});
        });

        describe.each(['world'])('%s', () => {
          beforeEach(() => {});
          beforeEach(() => {});

          it('is not fine', () => {});
        });
      });
    `,
      errors: [
        {
          messageId: 'noDuplicateHook',
          data: { hook: 'beforeEach' },
          column: 5,
          line: 10,
        },
      ],
    },
    {
      code: `
      describe('something', () => {
        describe.each(['hello'])('%s', () => {
          beforeEach(() => {});

          it('is fine', () => {});
        });

        describe.each(['world'])('%s', () => {
          describe('some more', () => {
            beforeEach(() => {});
            beforeEach(() => {});

            it('is not fine', () => {});
          });
        });
      });
    `,
      errors: [
        {
          messageId: 'noDuplicateHook',
          data: { hook: 'beforeEach' },
          column: 7,
          line: 11,
        },
      ],
    },
    {
      code: `
      describe.each\`\`('%s', () => {
        beforeEach(() => {});
        beforeEach(() => {});

        it('is fine', () => {});
      });
    `,
      errors: [
        {
          messageId: 'noDuplicateHook',
          data: { hook: 'beforeEach' },
          column: 3,
          line: 3,
        },
      ],
    },
    {
      code: `
      describe('something', () => {
        describe.each\`\`('%s', () => {
          beforeEach(() => {});

          it('is fine', () => {});
        });

        describe.each\`\`('%s', () => {
          beforeEach(() => {});
          beforeEach(() => {});

          it('is not fine', () => {});
        });
      });
    `,
      errors: [
        {
          messageId: 'noDuplicateHook',
          data: { hook: 'beforeEach' },
          column: 5,
          line: 10,
        },
      ],
    },
  ],
});
