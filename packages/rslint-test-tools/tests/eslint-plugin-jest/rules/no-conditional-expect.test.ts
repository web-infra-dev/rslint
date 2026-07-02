import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-conditional-expect', {} as never, {
  valid: [
    {
      code: `
      it('foo', () => {
        expect(1).toBe(2);
      });
    `,
    },
    {
      code: `
      it('foo', () => {
        expect(!true).toBe(false);
      });
    `,
    },
    {
      code: `
      it('foo', () => {
        process.env.FAIL && setNum(1);

        expect(num).toBe(2);
      });
    `,
    },
    {
      code: `
      function getValue() {
        let num = 2;

        process.env.FAIL && setNum(1);

        return num;
      }

      it('foo', () => {
        expect(getValue()).toBe(2);
      });
    `,
    },
    {
      code: `
      it('foo', () => {
        process.env.FAIL || setNum(1);

        expect(num).toBe(2);
      });
    `,
    },
    {
      code: `
      function getValue() {
        let num = 2;

        process.env.FAIL || setNum(1);

        return num;
      }

      it('foo', () => {
        expect(getValue()).toBe(2);
      });
    `,
    },
    {
      code: `
      it('foo', () => {
        const num = process.env.FAIL ? 1 : 2;

        expect(num).toBe(2);
      });
    `,
    },
    {
      code: `
      function getValue() {
        return process.env.FAIL ? 1 : 2
      }

      it('foo', () => {
        expect(getValue()).toBe(2);
      });
    `,
    },
    {
      code: `
      it('foo', () => {
        let num;

        switch(process.env.FAIL) {
          case true:
            num = 1;
            break;
          case false:
            num = 2;
            break;
        }

        expect(num).toBe(2);
      });
    `,
    },
    {
      code: `
      function getValue() {
        switch(process.env.FAIL) {
          case true:
            return 1;
          case false:
            return 2;
        }
      }

      it('foo', () => {
        expect(getValue()).toBe(2);
      });
    `,
    },
    {
      code: `
      it('foo', () => {
        let num = 2;

        if(process.env.FAIL) {
          num = 1;
        }

        expect(num).toBe(2);
      });
    `,
    },
    {
      code: `
      function getValue() {
        if(process.env.FAIL) {
          return 1;
        }

        return 2;
      }

      it('foo', () => {
        expect(getValue()).toBe(2);
      });
    `,
    },
    {
      code: `
      it('foo', () => {
        try {
          // do something
        } catch {
          // ignore errors
        } finally {
          expect(something).toHaveBeenCalled();
        }
      });
    `,
    },
    {
      code: `
      it('foo', () => {
        try {
          // do something
        } catch {
          // ignore errors
        }

        expect(something).toHaveBeenCalled();
      });
    `,
    },
    {
      code: `
      function getValue() {
        try {
          // do something
        } catch {
          // ignore errors
        } finally {
          expect(something).toHaveBeenCalled();
        }
      }

      it('foo', getValue);
    `,
    },
    {
      code: `
      function getValue() {
        try {
          process.env.FAIL.toString();

          return 1;
        } catch {
          return 2;
        }
      }

      it('foo', () => {
        expect(getValue()).toBe(2);
      });
    `,
    },
    {
      code: `
      it('works', async () => {
        try {
          await Promise.resolve().then(() => {
            throw new Error('oh noes!');
          });
        } catch {
          // ignore errors
        } finally {
          expect(something).toHaveBeenCalled();
        }
      });
    `,
    },
    {
      code: `
      it('works', async () => {
        await doSomething().catch(error => error);

        expect(error).toBeInstanceOf(Error);
      });
    `,
    },
    {
      code: `
      it('works', async () => {
        try {
          await Promise.resolve().then(() => {
            throw new Error('oh noes!');
          });
        } catch {
          // ignore errors
        }

        expect(something).toHaveBeenCalled();
      });
    `,
    },
  ],
  invalid: [
    {
      code: `
        it('foo', () => {
          something && expect(something).toHaveBeenCalled();
        })
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        it('foo', () => {
          a || b && expect(something).toHaveBeenCalled();
        })
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        it('foo', () => {
          (a || b) && expect(something).toHaveBeenCalled();
        })
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        it('foo', () => {
          a || (b && expect(something).toHaveBeenCalled());
        })
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        it('foo', () => {
          a && b && expect(something).toHaveBeenCalled();
        })
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        it('foo', () => {
          a && b || expect(something).toHaveBeenCalled();
        })
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        it('foo', () => {
          (a && b) || expect(something).toHaveBeenCalled();
        })
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        function getValue() {
          something && expect(something).toHaveBeenCalled();
        }

        it('foo', getValue);
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        it('foo', () => {
          something || expect(something).toHaveBeenCalled();
        })
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        it.each\`\`('foo', () => {
          something || expect(something).toHaveBeenCalled();
        })
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        it.each()('foo', () => {
          something || expect(something).toHaveBeenCalled();
        })
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        function getValue() {
          something || expect(something).toHaveBeenCalled();
        }

        it('foo', getValue);
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        it('foo', () => {
          something ? expect(something).toHaveBeenCalled() : noop();
        })
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        function getValue() {
          something ? expect(something).toHaveBeenCalled() : noop();
        }

        it('foo', getValue);
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        it('foo', () => {
          something ? noop() : expect(something).toHaveBeenCalled();
        })
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        it.each\`\`('foo', () => {
          something ? noop() : expect(something).toHaveBeenCalled();
        })
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        it.each()('foo', () => {
          something ? noop() : expect(something).toHaveBeenCalled();
        })
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        function getValue() {
          something ? noop() : expect(something).toHaveBeenCalled();
        }

        it('foo', getValue);
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        it('foo', () => {
          switch(something) {
            case 'value':
              break;
            default:
              expect(something).toHaveBeenCalled();
          }
        })
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        it('foo', () => {
          switch(something) {
            case 'value':
              expect(something).toHaveBeenCalled();
            default:
              break;
          }
        })
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        it.each\`\`('foo', () => {
          switch(something) {
            case 'value':
              expect(something).toHaveBeenCalled();
            default:
              break;
          }
        })
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        it.each()('foo', () => {
          switch(something) {
            case 'value':
              expect(something).toHaveBeenCalled();
            default:
              break;
          }
        })
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        function getValue() {
          switch(something) {
            case 'value':
              break;
            default:
              expect(something).toHaveBeenCalled();
          }
        }

        it('foo', getValue);
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        it('foo', () => {
          if(doSomething) {
            expect(something).toHaveBeenCalled();
          }
        })
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        it('foo', () => {
          if(!doSomething) {
            // do nothing
          } else {
            expect(something).toHaveBeenCalled();
          }
        })
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        it.each\`\`('foo', () => {
          if(!doSomething) {
            // do nothing
          } else {
            expect(something).toHaveBeenCalled();
          }
        })
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        it.each()('foo', () => {
          if(!doSomething) {
            // do nothing
          } else {
            expect(something).toHaveBeenCalled();
          }
        })
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        function getValue() {
          if(doSomething) {
            expect(something).toHaveBeenCalled();
          }
        }

        it('foo', getValue);
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        function getValue() {
          if(!doSomething) {
            // do nothing
          } else {
            expect(something).toHaveBeenCalled();
          }
        }

        it('foo', getValue);
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        it('foo', () => {
          try {

          } catch (err) {
            expect(err).toMatch('Error');
          }
        })
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        it.each\`\`('foo', () => {
          try {

          } catch (err) {
            expect(err).toMatch('Error');
          }
        })
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        it.each()('foo', () => {
          try {

          } catch (err) {
            expect(err).toMatch('Error');
          }
        })
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        it.skip.each\`\`('foo', () => {
          try {

          } catch (err) {
            expect(err).toMatch('Error');
          }
        })
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        it.skip.each()('foo', () => {
          try {

          } catch (err) {
            expect(err).toMatch('Error');
          }
        })
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        function getValue() {
          try {
            // do something
          } catch {
            expect(something).toHaveBeenCalled();
          }
        }

        it('foo', getValue);
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        it('works', async () => {
          await Promise.resolve()
            .then(() => { throw new Error('oh noes!'); })
            .catch(error => expect(error).toBeInstanceOf(Error));
        });
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        it('works', async () => {
          await Promise.resolve()
            .then(() => { throw new Error('oh noes!'); })
            .catch(error => expect(error).toBeInstanceOf(Error))
            .then(() => { throw new Error('oh noes!'); })
            .catch(error => expect(error).toBeInstanceOf(Error))
            .then(() => { throw new Error('oh noes!'); })
            .catch(error => expect(error).toBeInstanceOf(Error));
        });
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        it('works', async () => {
          await Promise.resolve()
            .catch(error => expect(error).toBeInstanceOf(Error))
            .catch(error => expect(error).toBeInstanceOf(Error))
            .catch(error => expect(error).toBeInstanceOf(Error));
        });
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        it('works', async () => {
          await Promise.resolve()
            .catch(error => expect(error).toBeInstanceOf(Error))
            .then(() => { throw new Error('oh noes!'); })
            .then(() => { throw new Error('oh noes!'); })
            .then(() => { throw new Error('oh noes!'); });
        });
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        it('works', async () => {
          await somePromise
            .then(() => { throw new Error('oh noes!'); })
            .catch(error => expect(error).toBeInstanceOf(Error));
        });
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
    {
      code: `
        it('works', async () => {
          await somePromise.catch(error => expect(error).toBeInstanceOf(Error));
        });
      `,
      errors: [{ messageId: 'conditionalExpect' }],
    },
  ],
});
