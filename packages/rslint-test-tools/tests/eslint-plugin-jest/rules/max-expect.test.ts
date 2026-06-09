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
      test('should pass', () => {
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        // expect(true).toBeDefined();
      });
    `,
    },
    {
      code: `
      it('should pass', () => {
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
        expect(true).toBeDefined();
      });
    `,
    },
    {
      code: `
      test('should pass', async () => {
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
      describe('test', () => {
        test('should pass', () => {
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
        });
      });
    `,
    },
    {
      code: `
      test.each(['should', 'pass'], () => {
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
      test('should pass', () => {
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
      });
      test('should pass', () => {
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
      });
    `,
    },
    {
      code: `
      test('should not pass', () => {
        const checkValue = (value) => {
          expect(value).toBeDefined();
          expect(value).toBeDefined();
        };

        checkValue(true);
      });
      test('should not pass', () => {
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
      });
    `,
    },
    {
      code: `
      test('should not pass', done => {
        emitter.on('event', value => {
          expect(value).toBeDefined();
          expect(value).toBeDefined();
          expect(value).toBeDefined();
          expect(value).toBeDefined();

          done();
        });
      });
      test('should not pass', () => {
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
      });
    `,
    },
    {
      code: `
      function myHelper() {
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
      };

      test('should pass', () => {
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
      });
    `,
    },
    {
      code: `
      function myHelper1() {
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
      };

      test('should pass', () => {
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
      });

      function myHelper2() {
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
      };
    `,
    },
    {
      code: `
      test('should pass', () => {
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
      });

      function myHelper() {
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
      };
    `,
    },
    {
      code: `
      const myHelper1 = () => {
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
      };

      test('should pass', function() {
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
      });

      const myHelper2 = function() {
        expect(true).toBeDefined();
        expect(true).toBeDefined();
        expect(true).toBeDefined();
      };
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
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
        });
      `,
      options: [
        {
          max: 10,
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
          line: 7,
          column: 3,
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
          line: 7,
          column: 3,
        },
      ],
    },
    {
      code: `
        it('should not pass', () => {
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
          line: 7,
          column: 3,
        },
      ],
    },
    {
      code: `
        it('should not pass', async () => {
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
          line: 7,
          column: 3,
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
          line: 7,
          column: 3,
        },
        {
          messageId: 'exceededMaxAssertion',
          line: 15,
          column: 3,
        },
      ],
    },
    {
      code: `
        test('should not pass', () => {
          const checkValue = (value) => {
            expect(value).toBeDefined();
            expect(value).toBeDefined();
          };

          checkValue(true);
        });
        test('should not pass', () => {
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
        });
      `,
      options: [{ max: 1 }],
      errors: [
        {
          messageId: 'exceededMaxAssertion',
          line: 4,
          column: 5,
        },
        {
          messageId: 'exceededMaxAssertion',
          line: 11,
          column: 3,
        },
        {
          messageId: 'exceededMaxAssertion',
          line: 12,
          column: 3,
        },
      ],
    },
    {
      code: `
        test('should not pass', () => {
          const checkValue = (value) => {
            expect(value).toBeDefined();
            expect(value).toBeDefined();
          };

          checkValue(true);
        });
        test('should not pass', () => {
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
        });
      `,
      options: [{ max: 2 }],
      errors: [
        {
          messageId: 'exceededMaxAssertion',
          line: 12,
          column: 3,
        },
      ],
    },
    {
      code: `
        test('should not pass', () => {
          const checkValue = (value) => {
            expect(value).toBeDefined();
            expect(value).toBeDefined();
          };

          expect(value).toBeDefined();
          checkValue(true);
        });
        test('should not pass', () => {
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
        });
      `,
      options: [{ max: 2 }],
      errors: [
        {
          messageId: 'exceededMaxAssertion',
          line: 13,
          column: 3,
        },
      ],
    },
    {
      code: `
        test('should not pass', done => {
          emitter.on('event', value => {
            expect(value).toBeDefined();
            expect(value).toBeDefined();

            done();
          });
        });
        test('should not pass', () => {
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
        });
      `,
      options: [{ max: 1 }],
      errors: [
        {
          messageId: 'exceededMaxAssertion',
          line: 4,
          column: 5,
        },
        {
          messageId: 'exceededMaxAssertion',
          line: 11,
          column: 3,
        },
        {
          messageId: 'exceededMaxAssertion',
          line: 12,
          column: 3,
        },
      ],
    },
    {
      code: `
        test('should not pass', done => {
          emitter.on('event', value => {
            expect(value).toBeDefined();
            expect(value).toBeDefined();

            done();
          });
        });
        test('should not pass', () => {
          expect(true).toBeDefined();
          expect(true).toBeDefined();
          expect(true).toBeDefined();
        });
      `,
      options: [{ max: 2 }],
      errors: [
        {
          messageId: 'exceededMaxAssertion',
          line: 12,
          column: 3,
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
          line: 12,
          column: 5,
        },
        {
          messageId: 'exceededMaxAssertion',
          line: 13,
          column: 5,
        },
      ],
    },
    {
      code: `
        describe('test', () => {
          test('should not pass', () => {
            expect(true).toBeDefined();
            expect(true).toBeDefined();
            expect(true).toBeDefined();
            expect(true).toBeDefined();
            expect(true).toBeDefined();
            expect(true).toBeDefined();
          });
        });
      `,
      errors: [
        {
          messageId: 'exceededMaxAssertion',
          line: 8,
          column: 5,
        },
      ],
    },
    {
      code: `
        test.each(['should', 'not', 'pass'], () => {
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
          line: 7,
          column: 3,
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
      options: [
        {
          max: 1,
        },
      ],
      errors: [
        {
          messageId: 'exceededMaxAssertion',
          line: 3,
          column: 3,
        },
      ],
    },
  ],
});
