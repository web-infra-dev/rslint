import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('no-conditional-in-test', {
  valid: [
    { code: 'const x = y ? 1 : 0' },
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
        switch (true) {
          case true: {}
        }
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
        describe('foo', () => {
          beforeEach(() => {
            if ('bar') {}
          });
        })
      `,
    },
    { code: 'if (foo) {}' },
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
        it('foo', () => {
          obj?.foo?.bar;
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
        const values = something.map(thing => thing?.foo);

        it('valid', () => {
          expect(values).toStrictEqual(['foo']);
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
          message: 'Avoid having conditionals in tests',
          line: 3,
          column: 18,
          endLine: 3,
          endColumn: 33,
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
          message: 'Avoid having conditionals in tests',
          line: 4,
          column: 20,
          endLine: 4,
          endColumn: 36,
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
          message: 'Avoid having conditionals in tests',
          line: 3,
          column: 11,
          endLine: 5,
          endColumn: 12,
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
          message: 'Avoid having conditionals in tests',
          line: 4,
          column: 13,
          endLine: 4,
          endColumn: 29,
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
          message: 'Avoid having conditionals in tests',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 24,
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
          message: 'Avoid having conditionals in tests',
          line: 4,
          column: 13,
          endLine: 4,
          endColumn: 26,
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
      `,
      errors: [
        {
          messageId: 'conditionalInTest',
          message: 'Avoid having conditionals in tests',
          line: 3,
          column: 11,
          endLine: 5,
          endColumn: 12,
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
          message: 'Avoid having conditionals in tests',
          line: 3,
          column: 25,
          endLine: 3,
          endColumn: 33,
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
          message: 'Avoid having conditionals in tests',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 24,
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
          message: 'Avoid having conditionals in tests',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 21,
        },
      ],
    },
  ],
});
