import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('valid-expect-in-promise', {} as never, {
  valid: [
    {
      code: `test('case', async () => {
        await load().then(value => expect(value).toBe(1));
      });`,
    },
    {
      code: `test('case', async () => {
        let pending;
        pending ||= load().then(value => expect(value).toBe(1));
        await pending;
      });`,
    },
    {
      code: `test('case', async () => {
        let pending = fallback || load().then(value => expect(value).toBe(1));
        await pending;
      });`,
    },
    {
      code: `test('case', async () => {
        let pending = condition ? load().then(value => expect(value).toBe(1)) : other;
        await pending;
      });`,
    },
    {
      code: `test('case', async () => {
        const pending = [load().then(value => expect(value).toBe(1))];
        await Promise.all(pending);
      });`,
    },
    {
      code: `test('case', async () => {
        const pending = [load().then(value => expect(value).toBe(1))];
        for await (const settled of pending) {
        }
      });`,
    },
    {
      code: `test('case', async () => {
        let pending = load().then(value => expect(value).toBe(1));
        pending ??= fallback;
        await pending;
      });`,
    },
    {
      code: `test('case', () =>
        load().then(value => expect(value).toBe(1))
      );`,
    },
    {
      code: `test('case', () => {
        try {
          return load().then(value => expect(value).toBe(1));
        } catch (error) {}
      });`,
    },
    {
      code: `test('case', async () => {
        try {
          return load().then(value => expect(value).toBe(1));
        } catch (error) {}
      });`,
    },
    {
      code: `test('case', async () => {
        const pending = load().then(value => expect(value).toBe(1));
        unrelated = 1;
        await pending;
      });`,
    },
    {
      code: `test('case', () => {
        const pending = load().then(value => expect(value).toBe(1));
        return Promise.all([pending]);
      });`,
    },
    {
      code: `test('case', () => {
        const [pending] = [load().then(value => expect(value).toBe(1))];
        return pending;
      });`,
    },
    {
      code: `test('case', () => {
        const pending = load().then(value => expect(value).toBe(1));
        expect(pending).resolves.toBeUndefined();
      });`,
    },
    {
      code: `test('done', done => {
        load().then(value => {
          expect(value).toBe(1);
          done();
        });
      });`,
    },
    {
      code: `function callback() {
        const pending = load().then(value => expect(value).toBe(1));
        return pending;
      }
      test('case', callback);`,
    },
    {
      code: `function handler() {
        queue.then(handler);
      }
      test('case', () => promise.then(handler));`,
    },
    {
      code: `function handler(value) {
        expect(value).toBe(1);
        return queue.then(handler);
      }
      test('case', () => promise.then(handler));`,
    },
    {
      code: `test('case', async () => {
        const pending = load().then(value => expect(value).toBe(1));
        try {
          await pending;
        } catch (error) {
          throw error;
        }
      });`,
    },
    {
      code: `test('case', async () => {
        const pending = load().then(value => expect(value).toBe(1));
        if (!ready) throw new Error('no');
        await pending;
      });`,
    },
    {
      code: `test('case', async () => {
        const pending = load().then(value => expect(value).toBe(1));
        try {
          setup();
          await pending;
        } catch (error) {
          throw error;
        }
      });`,
    },
    {
      code: `try {
        test('case', async () => {
          const pending = load().then(value => expect(value).toBe(1));
          await pending;
        });
      } catch (error) {}`,
    },
    {
      code: `try {
        test('case', () => {
          const pending = load().then(value => expect(value).toBe(1));
          return Promise.resolve(pending);
        });
      } catch (error) {}`,
    },
    {
      code: `try {
        test('case', () => {
          const pending = load().then(value => expect(value).toBe(1));
          return Promise.all([pending]);
        });
      } catch (error) {}`,
    },
    {
      code: `Promise.resolve().then(() => expect(1).toBe(2));`,
    },
  ],
  invalid: [
    {
      code: `test('case', () => {
        list.push(load().then(value => expect(value).toBe(1)));
      });`,
      errors: [{ messageId: 'expectInFloatingPromise' }],
    },
    {
      code: `test('case', async () => {
        const pending = [load().then(value => expect(value).toBe(1))];
        await pending;
      });`,
      errors: [{ messageId: 'expectInFloatingPromise' }],
    },
    {
      code: `test('case', async () => {
        let pending = load().then(value => expect(value).toBe(1));
        pending = pending && fallback;
        await pending;
      });`,
      errors: [{ messageId: 'expectInFloatingPromise' }],
    },
    {
      code: `function leaf(value) {
        expect(value).toBe(1);
      }
      function shared() {
        return queue.then(leaf);
      }
      test('one', () => {
        promise.then(shared);
      });
      test('two', () => {
        promise.then(shared);
      });`,
      errors: [
        { messageId: 'expectInFloatingPromise' },
        { messageId: 'expectInFloatingPromise' },
      ],
    },
    {
      code: `test('case', () => {
        load().then(value => expect(value).toBe(1));
      });`,
      errors: [{ messageId: 'expectInFloatingPromise', line: 2, column: 9 }],
    },
    {
      code: `test('case', () => {
        const pending = load().then(value => expect(value).toBe(1));
        return Promise.reject(pending);
      });`,
      errors: [{ messageId: 'expectInFloatingPromise', line: 2, column: 15 }],
    },
    {
      code: `test('case', () => {
        const pending = load().then(value => expect(value).toBe(1));
        return Promise.allSettled([pending]);
      });`,
      errors: [{ messageId: 'expectInFloatingPromise', line: 2, column: 15 }],
    },
    {
      code: `test('case', async () => {
        const pending = load().then(value => expect(value).toBe(1));
        if (condition) await pending;
      });`,
      errors: [{ messageId: 'expectInFloatingPromise' }],
    },
    {
      code: `test('case', async () => {
        const pending = load().then(value => expect(value).toBe(1));
        try {
          await pending;
        } catch {}
      });`,
      errors: [{ messageId: 'expectInFloatingPromise' }],
    },
    {
      code: `test('case', async () => {
        const pending = load().then(value => expect(value).toBe(1));
        try {
          throw new Error('caught');
        } catch {}
      });`,
      errors: [{ messageId: 'expectInFloatingPromise' }],
    },
    {
      code: `test('case', async () => {
        const pending = load().then(value => expect(value).toBe(1));
        try {
          throw new Error('suppressed');
        } finally {
          return;
        }
      });`,
      errors: [{ messageId: 'expectInFloatingPromise' }],
    },
    {
      code: `function callback() {
        load().then(value => expect(value).toBe(1));
      }
      test('case', callback);`,
      errors: [{ messageId: 'expectInFloatingPromise' }],
    },
    {
      code: `test('case', async () => {
        await first().then(() => {
          second().then(value => expect(value).toBe(1));
        });
      });`,
      errors: [{ messageId: 'expectInFloatingPromise' }],
    },
    {
      code: `test('case', () => {
        load()
          .then(value => expect(value).toBe(1))
          .then(value => expect(value).toBe(2));
      });`,
      errors: 1,
    },
    {
      code: `test('case', () => {
        const [pending] = load().then(value => expect(value).toBe(1));
      });`,
      errors: [{ messageId: 'expectInFloatingPromise' }],
    },
    {
      code: `test('case', () => {
        const [pending] = [load().then(value => expect(value).toBe(1))];
      });`,
      errors: [{ messageId: 'expectInFloatingPromise' }],
    },
    {
      code: `test('case', () => {
        const { pending } = {
          pending: load().then(value => expect(value).toBe(1)),
        };
      });`,
      errors: [{ messageId: 'expectInFloatingPromise' }],
    },
    {
      code: `test('case', () => {
        consume({
          pending: load().then(value => expect(value).toBe(1)),
        });
      });`,
      errors: [{ messageId: 'expectInFloatingPromise' }],
    },
    {
      code: `function handler(value) {
        expect(value).toBe(1);
      }
      test('one', () => { first().then(handler); });
      test('two', () => { second().then(handler); });`,
      errors: 2,
    },
    {
      code: `test('case', async () => {
        let pending;
        pending ||= load().then(value => expect(value).toBe(1));
      });`,
      errors: [{ messageId: 'expectInFloatingPromise' }],
    },
    {
      code: `test('case', async () => {
        let pending = load().then(value => expect(value).toBe(1));
        pending &&= fallback;
        await pending;
      });`,
      errors: [{ messageId: 'expectInFloatingPromise' }],
    },
  ],
});
