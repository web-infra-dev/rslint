import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('max-expects', {} as never, {
  valid: [
    { code: `test('should pass')` },
    { code: `test('should pass', () => {})` },
    { code: `test.skip('should pass', () => {})` },
    {
      code: `
        test('should pass', function () {
          expect(true).toBeDefined();
        });
      `,
    },
    {
      code: `
        test('should pass', () => {
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
        });
      `,
    },
    {
      code: `
        test('should pass', async () => {
          expect.hasAssertions();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toEqual(expect.any(Boolean));
        });
      `,
    },
    {
      code: `
        describe('given decimal places', () => {
          it("test 1", fakeAsync(() => {
            expect(true).toBeTrue();
            expect(true).toBeTrue();
            expect(true).toBeTrue();
          }))

          it("test 2", fakeAsync(() => {
            expect(true).toBeTrue();
            expect(true).toBeTrue();
            expect(true).toBeTrue();
          }))
        })
      `,
      options: [{ max: 5 }],
    },
  ],
  invalid: [
    {
      code: `
        test('should not pass', function () {
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
        });
      `,
      errors: [
        {
          messageId: 'exceededMaxAssertion',
          message: 'Too many assertion calls (6) - maximum allowed is 5',
          line: 8,
          column: 11,
        },
      ],
    },
    {
      code: `
        test('should not pass', () => {
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
        });
      `,
      errors: [
        {
          messageId: 'exceededMaxAssertion',
          message: 'Too many assertion calls (6) - maximum allowed is 5',
          line: 8,
          column: 11,
        },
      ],
    },
    {
      code: `
        test('should not pass', () => {
          expect(true).toBeDefined();
          expect(true).toBeDefined();
        });
      `,
      options: [{ max: 1 }],
      errors: [
        {
          messageId: 'exceededMaxAssertion',
          message: 'Too many assertion calls (2) - maximum allowed is 1',
          line: 4,
          column: 11,
        },
      ],
    },
    {
      code: `
        describe('given decimal places', () => {
          it("test 1", fakeAsync(() => {
            expect(true).toBeTrue();
            expect(true).toBeTrue();
            expect(true).toBeTrue();
          }))

          it("test 2", fakeAsync(() => {
            expect(true).toBeTrue();
            expect(true).toBeTrue();
            expect(true).toBeTrue();
            expect(true).toBeTrue();
            expect(true).toBeTrue();
          }))
        })
      `,
      options: [{ max: 3 }],
      errors: [
        {
          messageId: 'exceededMaxAssertion',
          message: 'Too many assertion calls (4) - maximum allowed is 3',
          line: 13,
          column: 13,
        },
        {
          messageId: 'exceededMaxAssertion',
          message: 'Too many assertion calls (5) - maximum allowed is 3',
          line: 14,
          column: 13,
        },
      ],
    },
    {
      code: `
        test('chai chains still count as assertions', () => {
          expect(true).toBeDefined();
          expect('hey').to.be.a('string');
        });
      `,
      options: [{ max: 1 }],
      errors: [
        {
          messageId: 'exceededMaxAssertion',
          message: 'Too many assertion calls (2) - maximum allowed is 1',
          line: 4,
          column: 11,
        },
      ],
    },
    {
      code: `const options = { timeout: 1000 };
test('wrapped callback with non-literal options', options, fakeAsync(() => {
  expect(1).toBe(1);
  expect(2).toBe(2);
}));`,
      options: [{ max: 1 }],
      errors: [
        {
          messageId: 'exceededMaxAssertion',
          message: 'Too many assertion calls (2) - maximum allowed is 1',
          line: 4,
          column: 3,
        },
      ],
    },
    {
      code: `type TestCallback = () => void;
test('callback behind as expression', (() => {
  expect(1).toBe(1);
  expect(2).toBe(2);
}) as TestCallback);`,
      options: [{ max: 1 }],
      errors: [
        {
          messageId: 'exceededMaxAssertion',
          message: 'Too many assertion calls (2) - maximum allowed is 1',
          line: 4,
          column: 3,
        },
      ],
    },
  ],
});
