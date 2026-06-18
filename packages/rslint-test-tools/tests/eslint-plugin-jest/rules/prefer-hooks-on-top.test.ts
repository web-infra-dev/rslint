import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-hooks-on-top', {} as never, {
  valid: [
    {
      code: `
      describe('foo', () => {
        beforeEach(() => {});
        someSetupFn();
        afterEach(() => {});

        test('bar', () => {
          someFn();
        });
      });
    `,
    },
    {
      code: `
      describe('foo', () => {
        someSetupFn();
        beforeEach(() => {});
        afterEach(() => {});

        test('bar', () => {
          someFn();
        });
      });
    `,
    },
    {
      code: `
      describe.skip('foo', () => {
        beforeEach(() => {});
        beforeAll(() => {});

        test('bar', () => {
          someFn();
        });
      });

      describe('foo', () => {
        beforeEach(() => {});

        test('bar', () => {
          someFn();
        });
      });
    `,
    },
    {
      code: `
      describe('foo', () => {
        beforeEach(() => {});
        test('bar', () => {
          someFn();
        });

        describe('inner_foo', () => {
          beforeEach(() => {});
          test('inner bar', () => {
            someFn();
          });
        });
      });
    `,
    },
  ],
  invalid: [
    {
      code: `
        describe('foo', () => {
          beforeEach(() => {});
          test('bar', () => {
            someFn();
          });

          beforeAll(() => {});
          test('bar', () => {
            someFn();
          });
        });
      `,
      errors: [
        {
          messageId: 'noHookOnTop',
          column: 3,
          line: 7,
        },
      ],
    },
    {
      code: `
        describe('foo', () => {
          beforeEach(() => {});
          test.each\`\`('bar', () => {
            someFn();
          });

          beforeAll(() => {});
          test.only('bar', () => {
            someFn();
          });
        });
      `,
      errors: [
        {
          messageId: 'noHookOnTop',
          column: 3,
          line: 7,
        },
      ],
    },
    {
      code: `
        describe('foo', () => {
          beforeEach(() => {});
          test.only.each\`\`('bar', () => {
            someFn();
          });

          beforeAll(() => {});
          test.only('bar', () => {
            someFn();
          });
        });
      `,
      errors: [
        {
          messageId: 'noHookOnTop',
          column: 3,
          line: 7,
        },
      ],
    },
    {
      code: `
        describe.skip('foo', () => {
          beforeEach(() => {});
          test('bar', () => {
            someFn();
          });

          beforeAll(() => {});
          test('bar', () => {
            someFn();
          });
        });
        describe('foo', () => {
          beforeEach(() => {});
          beforeEach(() => {});
          beforeAll(() => {});

          test('bar', () => {
            someFn();
          });
        });

        describe('foo', () => {
          test('bar', () => {
            someFn();
          });

          beforeEach(() => {});
          beforeEach(() => {});
          beforeAll(() => {});
        });
      `,
      errors: [
        {
          messageId: 'noHookOnTop',
          column: 3,
          line: 7,
        },
        {
          messageId: 'noHookOnTop',
          column: 3,
          line: 27,
        },
        {
          messageId: 'noHookOnTop',
          column: 3,
          line: 28,
        },
        {
          messageId: 'noHookOnTop',
          column: 3,
          line: 29,
        },
      ],
    },
    {
      code: `
        describe('foo', () => {
          beforeAll(() => {});
          test('bar', () => {
            someFn();
          });

          describe('inner_foo', () => {
            beforeEach(() => {});
            test('inner bar', () => {
              someFn();
            });

            test('inner bar', () => {
              someFn();
            });

            beforeAll(() => {});
            afterAll(() => {});
            test('inner bar', () => {
              someFn();
            });
          });
        });
      `,
      errors: [
        {
          messageId: 'noHookOnTop',
          column: 5,
          line: 17,
        },
        {
          messageId: 'noHookOnTop',
          column: 5,
          line: 18,
        },
      ],
    },
  ],
});
