// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('no-unneeded-async-expect-function', {
  valid: [
    ...(
      [
        'expect.hasAssertions()',
        `
      it('pass', async () => {
        expect();
      })
   \u0020`,
        `
      it('pass', async () => {
        await expect(doSomethingAsync()).rejects.toThrow();
      })
   \u0020`,
        `
      it('pass', async () => {
        await expect(doSomethingAsync(1, 2)).resolves.toBe(1);
      })
   \u0020`,
        `
      it('pass', async () => {
        await expect(async () => {
          await doSomethingAsync();
          await doSomethingTwiceAsync(1, 2);
        }).rejects.toThrow();
      })
   \u0020`,
        `
        import { expect as pleaseExpect } from '@jest/globals';
        it('pass', async () => {
          await pleaseExpect(doSomethingAsync()).rejects.toThrow();
        })
     \u0020`,
        `
      it('pass', async () => {
        await expect(async () => {
          doSomethingAsync();
        }).rejects.toThrow();
      })
   \u0020`,
        `
      it('pass', async () => {
        await expect(async () => {
          const a = 1;
          await doSomethingAsync(a);
        }).rejects.toThrow();
      })
   \u0020`,
        `
      it('pass for non-async expect', async () => {
        await expect(() => {
          doSomethingSync(a);
        }).rejects.toThrow();
      })
   \u0020`,
        `
      it('pass for await in expect', async () => {
        await expect(await doSomethingAsync()).rejects.toThrow();
      })
   \u0020`,
        `
      it('pass for different matchers', async () => {
        await expect(await doSomething()).not.toThrow();
        await expect(await doSomething()).toHaveLength(2);
        await expect(await doSomething()).toHaveReturned();
        await expect(await doSomething()).not.toHaveBeenCalled();
        await expect(await doSomething()).not.toBeDefined();
        await expect(await doSomething()).toEqual(2);
      })
   \u0020`,
        `
      it('pass for using await within for-loop', async () => {
        const b = [async () => Promise.resolve(1), async () => Promise.reject(2)];
        await expect(async() => {
          for (const a of b) {
            await b();
          }
        }).rejects.toThrow();
      })
   \u0020`,
        `
      it('pass for using await within array', async () => {
        await expect(async() => [await Promise.reject(2)]).rejects.toThrow(2);
      })
   \u0020`,
        `
      it('does not unwrap awaited identifiers', async () => {
        const promise = doSomethingAsync();
        await expect(async () => {
          await promise;
        }).rejects.toThrow();
      })
   \u0020`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    ...(
      [
        'expect.hasAssertions()',
        `it('pass', async () => {
  expect();
})`,
        `it('pass', async () => {
  await expect(doSomethingAsync()).rejects.toThrow();
})`,
        `it('pass', async () => {
  await expect(doSomethingAsync(1, 2)).resolves.toBe(1);
})`,
        `it('pass', async () => {
  await expect(async () => {
    await doSomethingAsync();
    await doSomethingTwiceAsync(1, 2);
  }).rejects.toThrow();
})`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream no-unneeded-async-expect-function / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
    })),
    {
      name: 'upstream no-unneeded-async-expect-function / valid 5',
      code: `import { expect as pleaseExpect } from '@jest/globals';
it('pass', async () => {
  await pleaseExpect(doSomethingAsync()).rejects.toThrow();
})`,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2017,
      },
    },
    ...(
      [
        `it('pass', async () => {
  await expect(async () => {
    doSomethingAsync();
  }).rejects.toThrow();
})`,
        `it('pass', async () => {
  await expect(async () => {
    const a = 1;
    await doSomethingAsync(a);
  }).rejects.toThrow();
})`,
        `it('pass for non-async expect', async () => {
  await expect(() => {
    doSomethingSync(a);
  }).rejects.toThrow();
})`,
        `it('pass for await in expect', async () => {
  await expect(await doSomethingAsync()).rejects.toThrow();
})`,
        `it('pass for different matchers', async () => {
  await expect(await doSomething()).not.toThrow();
  await expect(await doSomething()).toHaveLength(2);
  await expect(await doSomething()).toHaveReturned();
  await expect(await doSomething()).not.toHaveBeenCalled();
  await expect(await doSomething()).not.toBeDefined();
  await expect(await doSomething()).toEqual(2);
})`,
        `it('pass for using await within for-loop', async () => {
  const b = [async () => Promise.resolve(1), async () => Promise.reject(2)];
  await expect(async() => {
    for (const a of b) {
      await b();
    }
  }).rejects.toThrow();
})`,
        `it('pass for using await within array', async () => {
  await expect(async() => [await Promise.reject(2)]).rejects.toThrow(2);
})`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream no-unneeded-async-expect-function / valid ${index + 6}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
    })),
  ],
  invalid: [
    ...(
      [
        [
          `
        it('should be fixed', async () => {
          await expect(async () => {
            await doSomethingAsync();
          }).rejects.toThrow();\u0020
        })
     \u0020`,
          3,
          5,
          `
        it('should be fixed', async () => {
          await expect(doSomethingAsync()).rejects.toThrow();\u0020
        })
     \u0020`,
        ],
        // Original upstream output: "\n        it('should be fixed', async () => {\n          await expect(doSomethingAsync()).rejects.toThrow(); \n        })\n    "
        [
          `
        it('should be fixed', async () => {
          await expect(async function () {
            await doSomethingAsync();
          }).rejects.toThrow();\u0020
        })
     \u0020`,
          3,
          5,
          `
        it('should be fixed', async () => {
          await expect(doSomethingAsync()).rejects.toThrow();\u0020
        })
     \u0020`,
        ],
        [
          `
        it('should be fixed for async arrow function', async () => {
          await expect(async () => {
            await doSomethingAsync(1, 2);
          }).rejects.toThrow();\u0020
        })
     \u0020`,
          3,
          5,
          `
        it('should be fixed for async arrow function', async () => {
          await expect(doSomethingAsync(1, 2)).rejects.toThrow();\u0020
        })
     \u0020`,
        ],
        [
          `
        it('should be fixed for async normal function', async () => {
          await expect(async function () {
            await doSomethingAsync(1, 2);
          }).rejects.toThrow();\u0020
        })
     \u0020`,
          3,
          5,
          `
        it('should be fixed for async normal function', async () => {
          await expect(doSomethingAsync(1, 2)).rejects.toThrow();\u0020
        })
     \u0020`,
        ],
        [
          `
        it('should be fixed for Promise.all', async () => {
          await expect(async function () {
            await Promise.all([doSomethingAsync(1, 2), doSomethingAsync()]);
          }).rejects.toThrow();\u0020
        })
     \u0020`,
          3,
          5,
          `
        it('should be fixed for Promise.all', async () => {
          await expect(Promise.all([doSomethingAsync(1, 2), doSomethingAsync()])).rejects.toThrow();\u0020
        })
     \u0020`,
        ],
        [
          `
        it('should be fixed for async ref to expect', async () => {
          const a = async () => { await doSomethingAsync() };
          await expect(async () => {
            await a();
          }).rejects.toThrow();
        })
     \u0020`,
          4,
          6,
          `
        it('should be fixed for async ref to expect', async () => {
          const a = async () => { await doSomethingAsync() };
          await expect(a()).rejects.toThrow();
        })
     \u0020`,
        ],
        [
          `
        it('should be fixed for resolves', async () => {
          await expect(async () => {
            await doSomethingAsync();
          }).resolves.toBe(1);
        })
     \u0020`,
          3,
          5,
          `
        it('should be fixed for resolves', async () => {
          await expect(doSomethingAsync()).resolves.toBe(1);
        })
     \u0020`,
        ],
      ] as const
    ).map(([code, line, endLine, output], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'noAsyncWrapperForExpectedPromise',
          message:
            'Avoid wrapping asynchronous expectations in an unnecessary async function.',
          line,
          column: 24,
          endLine,
          endColumn: 12,
        },
      ],
      output,
    })),
    ...(
      [
        [
          `it('should be fixed', async () => {
  await expect(async () => {
    await doSomethingAsync();
  }).rejects.toThrow();\u0020
})`,
          2,
          4,
          `it('should be fixed', async () => {
  await expect(doSomethingAsync()).rejects.toThrow();\u0020
})`,
        ],
        [
          `it('should be fixed', async () => {
  await expect(async function () {
    await doSomethingAsync();
  }).rejects.toThrow();\u0020
})`,
          2,
          4,
          `it('should be fixed', async () => {
  await expect(doSomethingAsync()).rejects.toThrow();\u0020
})`,
        ],
        [
          `it('should be fixed for async arrow function', async () => {
  await expect(async () => {
    await doSomethingAsync(1, 2);
  }).rejects.toThrow();\u0020
})`,
          2,
          4,
          `it('should be fixed for async arrow function', async () => {
  await expect(doSomethingAsync(1, 2)).rejects.toThrow();\u0020
})`,
        ],
        [
          `it('should be fixed for async normal function', async () => {
  await expect(async function () {
    await doSomethingAsync(1, 2);
  }).rejects.toThrow();\u0020
})`,
          2,
          4,
          `it('should be fixed for async normal function', async () => {
  await expect(doSomethingAsync(1, 2)).rejects.toThrow();\u0020
})`,
        ],
        [
          `it('should be fixed for Promise.all', async () => {
  await expect(async function () {
    await Promise.all([doSomethingAsync(1, 2), doSomethingAsync()]);
  }).rejects.toThrow();\u0020
})`,
          2,
          4,
          `it('should be fixed for Promise.all', async () => {
  await expect(Promise.all([doSomethingAsync(1, 2), doSomethingAsync()])).rejects.toThrow();\u0020
})`,
        ],
        [
          `it('should be fixed for async ref to expect', async () => {
  const a = async () => { await doSomethingAsync() };
  await expect(async () => {
    await a();
  }).rejects.toThrow();
})`,
          3,
          5,
          `it('should be fixed for async ref to expect', async () => {
  const a = async () => { await doSomethingAsync() };
  await expect(a()).rejects.toThrow();
})`,
        ],
      ] as const
    ).map(([code, line, endLine, output], index): TestCase => ({
      name: `upstream no-unneeded-async-expect-function / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'noAsyncWrapperForExpectedPromise',
          message:
            'Avoid wrapping asynchronous expectations in an unnecessary async function.',
          line,
          column: 16,
          endLine,
          endColumn: 4,
        },
      ],
      output,
    })),
  ],
});
