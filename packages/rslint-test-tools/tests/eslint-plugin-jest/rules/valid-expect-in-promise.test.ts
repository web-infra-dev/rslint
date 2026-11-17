// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

// Shared source templates; reporting ranges remain explicit in each matrix.
const testSetup = `test('promise test', async function () {
  let somePromise = getPromise().then((data) => {
    expect(data).toEqual('foo');
`;
const testSetup2 = `it('is valid', async () => {
  const promise = loadNumber().then(number => {
    expect(typeof number).toBe('number');

    return number + 1;
  });

`;
const testSetup3 = `test('later return', async () => {
  const promise = something().then(value => {
    expect(value).toBe('red');
  });

`;

new RuleTester(fixtureConfig()).run('valid-expect-in-promise', {
  valid: [
    ...(
      [
        `test('case', async () => {
        await load().then(value => expect(value).toBe(1));
      });`,
        `test('case', async () => {
        let pending;
        pending ||= load().then(value => expect(value).toBe(1));
        await pending;
      });`,
        `test('case', async () => {
        let pending = fallback || load().then(value => expect(value).toBe(1));
        await pending;
      });`,
        `test('case', async () => {
        let pending = condition ? load().then(value => expect(value).toBe(1)) : other;
        await pending;
      });`,
        `test('case', async () => {
        const pending = [load().then(value => expect(value).toBe(1))];
        await Promise.all(pending);
      });`,
        `test('case', async () => {
        const pending = [load().then(value => expect(value).toBe(1))];
        for await (const settled of pending) {
        }
      });`,
        `test('case', async () => {
        let pending = load().then(value => expect(value).toBe(1));
        pending ??= fallback;
        await pending;
      });`,
        `test('case', () =>
        load().then(value => expect(value).toBe(1))
      );`,
        `test('case', () => {
        try {
          return load().then(value => expect(value).toBe(1));
        } catch (error) {}
      });`,
        `test('case', async () => {
        try {
          return load().then(value => expect(value).toBe(1));
        } catch (error) {}
      });`,
        `test('case', async () => {
        const pending = load().then(value => expect(value).toBe(1));
        unrelated = 1;
        await pending;
      });`,
        `test('case', () => {
        const pending = load().then(value => expect(value).toBe(1));
        return Promise.all([pending]);
      });`,
        `test('case', () => {
        const [pending] = [load().then(value => expect(value).toBe(1))];
        return pending;
      });`,
        `test('case', () => {
        const pending = load().then(value => expect(value).toBe(1));
        expect(pending).resolves.toBeUndefined();
      });`,
        `test('done', done => {
        load().then(value => {
          expect(value).toBe(1);
          done();
        });
      });`,
        `function callback() {
        const pending = load().then(value => expect(value).toBe(1));
        return pending;
      }
      test('case', callback);`,
        `function handler() {
        queue.then(handler);
      }
      test('case', () => promise.then(handler));`,
        `function handler(value) {
        expect(value).toBe(1);
        return queue.then(handler);
      }
      test('case', () => promise.then(handler));`,
        `test('case', async () => {
        const pending = load().then(value => expect(value).toBe(1));
        try {
          await pending;
        } catch (error) {
          throw error;
        }
      });`,
        `test('case', async () => {
        const pending = load().then(value => expect(value).toBe(1));
        if (!ready) throw new Error('no');
        await pending;
      });`,
        `test('case', async () => {
        const pending = load().then(value => expect(value).toBe(1));
        try {
          setup();
          await pending;
        } catch (error) {
          throw error;
        }
      });`,
        `try {
        test('case', async () => {
          const pending = load().then(value => expect(value).toBe(1));
          await pending;
        });
      } catch (error) {}`,
        `try {
        test('case', () => {
          const pending = load().then(value => expect(value).toBe(1));
          return Promise.resolve(pending);
        });
      } catch (error) {}`,
        `try {
        test('case', () => {
          const pending = load().then(value => expect(value).toBe(1));
          return Promise.all([pending]);
        });
      } catch (error) {}`,
        'Promise.resolve().then(() => expect(1).toBe(2));',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    ...(
      [
        [
          "test('something', () => Promise.resolve().then(() => expect(1).toBe(2)));",
          'upstream valid-expect-in-promise / valid 0',
        ],
        [
          'Promise.resolve().then(() => expect(1).toBe(2))',
          'upstream valid-expect-in-promise / valid 1',
        ],
        [
          'const x = Promise.resolve().then(() => expect(1).toBe(2))',
          'upstream valid-expect-in-promise / valid 2',
        ],
        ['it.todo("something")', 'upstream valid-expect-in-promise / valid 3'],
        [
          `it('is valid', () => {
  const promise = loadNumber().then(number => {
    expect(typeof number).toBe('number');

    return number + 1;
  });

  expect(promise).resolves.toBe(1);
});`,
          'upstream valid-expect-in-promise / valid 4',
        ],
        [
          `it('is valid', () => {
  const promise = loadNumber().then(number => {
    expect(typeof number).toBe('number');

    return number + 1;
  });

  expect(promise).resolves.not.toBe(2);
});`,
          'upstream valid-expect-in-promise / valid 5',
        ],
        [
          `it('is valid', () => {
  const promise = loadNumber().then(number => {
    expect(typeof number).toBe('number');

    return number + 1;
  });

  expect(promise).rejects.toBe(1);
});`,
          'upstream valid-expect-in-promise / valid 6',
        ],
        [
          `it('is valid', () => {
  const promise = loadNumber().then(number => {
    expect(typeof number).toBe('number');

    return number + 1;
  });

  expect(promise).rejects.not.toBe(2);
});`,
          'upstream valid-expect-in-promise / valid 7',
        ],
        [
          testSetup2 +
            `  expect(await promise).toBeGreaterThan(1);
});`,
          'upstream valid-expect-in-promise / valid 8',
        ],
        [
          testSetup2 +
            `  expect(await promise).resolves.toBeGreaterThan(1);
});`,
          'upstream valid-expect-in-promise / valid 9',
        ],
        [
          testSetup2 +
            `  expect(1).toBeGreaterThan(await promise);
});`,
          'upstream valid-expect-in-promise / valid 10',
        ],
        [
          testSetup2 +
            `  expect.this.that.is(await promise);
});`,
          'upstream valid-expect-in-promise / valid 11',
        ],
        [
          `it('is valid', async () => {
  expect(await loadNumber().then(number => {
    expect(typeof number).toBe('number');

    return number + 1;
  })).toBeGreaterThan(1);
});`,
          'upstream valid-expect-in-promise / valid 12',
        ],
        [
          testSetup2 +
            `  expect([await promise]).toHaveLength(1);
});`,
          'upstream valid-expect-in-promise / valid 13',
        ],
        [
          testSetup2 +
            `  expect([,,await promise,,]).toHaveLength(1);
});`,
          'upstream valid-expect-in-promise / valid 14',
        ],
        [
          testSetup2 +
            `  expect([[await promise]]).toHaveLength(1);
});`,
          'upstream valid-expect-in-promise / valid 15',
        ],
        [
          testSetup2 +
            `  logValue(await promise);
});`,
          'upstream valid-expect-in-promise / valid 16',
        ],
        [
          `it('is valid', async () => {
  const promise = loadNumber().then(number => {
    expect(typeof number).toBe('number');

    return 1;
  });

  expect.assertions(await promise);
});`,
          'upstream valid-expect-in-promise / valid 17',
        ],
        [
          `it('is valid', async () => {
  await loadNumber().then(number => {
    expect(typeof number).toBe('number');
  });
});`,
          'upstream valid-expect-in-promise / valid 18',
        ],
        [
          `it('it1', () => new Promise((done) => {
  test()
    .then(() => {
      expect(someThing).toEqual(true);
      done();
    });
}));`,
          'upstream valid-expect-in-promise / valid 19',
        ],
        [
          `it('it1', () => {
  return new Promise(done => {
    test().then(() => {
      expect(someThing).toEqual(true);
      done();
    });
  });
});`,
          'upstream valid-expect-in-promise / valid 20',
        ],
        [
          `it('passes', () => {
  Promise.resolve().then(() => {
    grabber.grabSomething();
  });
});`,
          'upstream valid-expect-in-promise / valid 21',
        ],
        [
          `it('passes', async () => {
  const grabbing = Promise.resolve().then(() => {
    grabber.grabSomething();
  });

  await grabbing;

  expect(grabber.grabbedItems).toHaveLength(1);
});`,
          'upstream valid-expect-in-promise / valid 22',
        ],
        [
          `const myFn = () => {
  Promise.resolve().then(() => {
    expect(true).toBe(false);
  });
};`,
          'upstream valid-expect-in-promise / valid 23',
        ],
        [
          `const myFn = () => {
  Promise.resolve().then(() => {
    subject.invokeMethod();
  });
};`,
          'upstream valid-expect-in-promise / valid 24',
        ],
        [
          `const myFn = () => {
  Promise.resolve().then(() => {
    expect(true).toBe(false);
  });
};

it('it1', () => {
  return somePromise.then(() => {
    expect(someThing).toEqual(true);
  });
});`,
          'upstream valid-expect-in-promise / valid 25',
        ],
        [
          `it('it1', () => new Promise((done) => {
  test()
    .finally(() => {
      expect(someThing).toEqual(true);
      done();
    });
}));`,
          'upstream valid-expect-in-promise / valid 26',
        ],
        [
          `it('it1', () => {
  return somePromise.then(() => {
    expect(someThing).toEqual(true);
  });
});`,
          'upstream valid-expect-in-promise / valid 27',
        ],
        [
          `it('it1', () => {
  return somePromise.finally(() => {
    expect(someThing).toEqual(true);
  });
});`,
          'upstream valid-expect-in-promise / valid 28',
        ],
        [
          `it('it1', function() {
  return somePromise.catch(function() {
    expect(someThing).toEqual(true);
  });
});`,
          'upstream valid-expect-in-promise / valid 29',
        ],
        [
          `xtest('it1', function() {
  return somePromise.catch(function() {
    expect(someThing).toEqual(true);
  });
});`,
          'upstream valid-expect-in-promise / valid 30',
        ],
        [
          `it('it1', function() {
  return somePromise.then(function() {
    doSomeThingButNotExpect();
  });
});`,
          'upstream valid-expect-in-promise / valid 31',
        ],
        [
          `it('it1', function() {
  return getSomeThing().getPromise().then(function() {
    expect(someThing).toEqual(true);
  });
});`,
          'upstream valid-expect-in-promise / valid 32',
        ],
        [
          `it('it1', function() {
  return Promise.resolve().then(function() {
    expect(someThing).toEqual(true);
  });
});`,
          'upstream valid-expect-in-promise / valid 33',
        ],
        [
          `it('it1', function () {
  return Promise.resolve().then(function () {
    /*fulfillment*/
    expect(someThing).toEqual(true);
  }, function () {
    /*rejection*/
    expect(someThing).toEqual(true);
  });
});`,
          'upstream valid-expect-in-promise / valid 34',
        ],
        [
          `it('it1', function () {
  Promise.resolve().then(/*fulfillment*/ function () {
  }, undefined, /*rejection*/ function () {
    expect(someThing).toEqual(true)
  })
});`,
          'upstream valid-expect-in-promise / valid 35',
        ],
        [
          `it('it1', function () {
  return Promise.resolve().then(function () {
    /*fulfillment*/
  }, function () {
    /*rejection*/
    expect(someThing).toEqual(true);
  });
});`,
          'upstream valid-expect-in-promise / valid 36',
        ],
        [
          `it('it1', function () {
  return somePromise.then()
});`,
          'upstream valid-expect-in-promise / valid 37',
        ],
        [
          `it('it1', async () => {
  await Promise.resolve().then(function () {
    expect(someThing).toEqual(true)
  });
});`,
          'upstream valid-expect-in-promise / valid 38',
        ],
        [
          `it('it1', async () => {
  await somePromise.then(() => {
    expect(someThing).toEqual(true)
  });
});`,
          'upstream valid-expect-in-promise / valid 39',
        ],
        [
          `it('it1', async () => {
  await getSomeThing().getPromise().then(function () {
    expect(someThing).toEqual(true)
  });
});`,
          'upstream valid-expect-in-promise / valid 40',
        ],
        [
          `it('it1', () => {
  return somePromise.then(() => {
    expect(someThing).toEqual(true);
  })
  .then(() => {
    expect(someThing).toEqual(true);
  })
});`,
          'upstream valid-expect-in-promise / valid 41',
        ],
        [
          `it('it1', () => {
  return somePromise.then(() => {
    return value;
  })
  .then(value => {
    expect(someThing).toEqual(value);
  })
});`,
          'upstream valid-expect-in-promise / valid 42',
        ],
        [
          `it('it1', () => {
  return somePromise.then(() => {
    expect(someThing).toEqual(true);
  })
  .then(() => {
    console.log('this is silly');
  })
});`,
          'upstream valid-expect-in-promise / valid 43',
        ],
        [
          `it('it1', () => {
  return somePromise.then(() => {
    expect(someThing).toEqual(true);
  })
  .catch(() => {
    expect(someThing).toEqual(false);
  })
});`,
          'upstream valid-expect-in-promise / valid 44',
        ],
        [
          `test('later return', () => {
  const promise = something().then(value => {
    expect(value).toBe('red');
  });

  return promise;
});`,
          'upstream valid-expect-in-promise / valid 45',
        ],
        [
          testSetup3 +
            `  await promise;
});`,
          'upstream valid-expect-in-promise / valid 46',
        ],
        [
          `test.only('later return', () => {
  const promise = something().then(value => {
    expect(value).toBe('red');
  });

  return promise;
});`,
          'upstream valid-expect-in-promise / valid 47',
        ],
        [
          `test('that we bailout if destructuring is used', async () => {
  const [promise] = await something().then(value => {
    expect(value).toBe('red');
  });
});`,
          'upstream valid-expect-in-promise / valid 49',
        ],
        [
          `it('shorthand arrow', () =>
  something().then(value => {
    expect(() => {
      value();
    }).toThrow();
  })
);`,
          'upstream valid-expect-in-promise / valid 53',
        ],
        [
          `it('crawls for files based on patterns', () => {
  const promise = nodeCrawl({}).then(data => {
    expect(childProcess.spawn).lastCalledWith('find');
  });
  return promise;
});`,
          'upstream valid-expect-in-promise / valid 54',
        ],
        [
          `it('is a test', async () => {
  const value = await somePromise().then(response => {
    expect(response).toHaveProperty('data');

    return response.data;
  });

  expect(value).toBe('hello world');
});`,
          'upstream valid-expect-in-promise / valid 55',
        ],
        [
          `it('is a test', async () => {
  return await somePromise().then(response => {
    expect(response).toHaveProperty('data');

    return response.data;
  });
});`,
          'upstream valid-expect-in-promise / valid 56',
        ],
        [
          `it('is a test', async () => {
  return somePromise().then(response => {
    expect(response).toHaveProperty('data');

    return response.data;
  });
});`,
          'upstream valid-expect-in-promise / valid 57',
        ],
        [
          `it('is a test', async () => {
  await somePromise().then(response => {
    expect(response).toHaveProperty('data');

    return response.data;
  });
});`,
          'upstream valid-expect-in-promise / valid 58',
        ],
        [
          `it(
  'test function',
  () => {
    return Builder
      .getPromiseBuilder()
      .get().build()
      .then((data) => {
        expect(data).toEqual('Hi');
      });
  }
);`,
          'upstream valid-expect-in-promise / valid 59',
        ],
        [
          `notATestFunction(
  'not a test function',
  () => {
    Builder
      .getPromiseBuilder()
      .get()
      .build()
      .then((data) => {
        expect(data).toEqual('Hi');
      });
  }
);`,
          'upstream valid-expect-in-promise / valid 60',
        ],
        [
          `it('is valid', async () => {
  const promiseOne = loadNumber().then(number => {
    expect(typeof number).toBe('number');
  });
  const promiseTwo = loadNumber().then(number => {
    expect(typeof number).toBe('number');
  });

  await promiseTwo;
  await promiseOne;
});`,
          'upstream valid-expect-in-promise / valid 61',
        ],
        [
          `it("it1", () => somePromise.then(() => {
  expect(someThing).toEqual(true)
}))`,
          'upstream valid-expect-in-promise / valid 62',
        ],
        [
          'it("it1", () => somePromise.then(() => expect(someThing).toEqual(true)))',
          'upstream valid-expect-in-promise / valid 63',
        ],
        [
          `it('promise test with done', (done) => {
  const promise = getPromise();
  promise.then(() => expect(someThing).toEqual(true));
});`,
          'upstream valid-expect-in-promise / valid 64',
        ],
        [
          `it('name of done param does not matter', (nameDoesNotMatter) => {
  const promise = getPromise();
  promise.then(() => expect(someThing).toEqual(true));
});`,
          'upstream valid-expect-in-promise / valid 65',
        ],
        [
          `it.each([])('name of done param does not matter', (nameDoesNotMatter) => {
  const promise = getPromise();
  promise.then(() => expect(someThing).toEqual(true));
});`,
          'upstream valid-expect-in-promise / valid 66',
        ],
        [
          `it.each\`\`('name of done param does not matter', ({}, nameDoesNotMatter) => {
  const promise = getPromise();
  promise.then(() => expect(someThing).toEqual(true));
});`,
          'upstream valid-expect-in-promise / valid 67',
        ],
        [
          `test('valid-expect-in-promise', async () => {
  const text = await fetch('url')
      .then(res => res.text())
      .then(text => text);

  expect(text).toBe('text');
});`,
          'upstream valid-expect-in-promise / valid 68',
        ],
        [
          testSetup +
            `  }), x = 1;

  await somePromise;
});`,
          'upstream valid-expect-in-promise / valid 69',
        ],
        [
          `test('promise test', async function () {
  let x = 1, somePromise = getPromise().then((data) => {
    expect(data).toEqual('foo');
  });

  await somePromise;
});`,
          'upstream valid-expect-in-promise / valid 70',
        ],
        [
          testSetup +
            `  });

  await somePromise;

  somePromise = getPromise().then((data) => {
    expect(data).toEqual('foo');
  });

  await somePromise;
});`,
          'upstream valid-expect-in-promise / valid 71',
        ],
        [
          testSetup +
            `  });

  await somePromise;

  somePromise = getPromise().then((data) => {
    expect(data).toEqual('foo');
  });

  return somePromise;
});`,
          'upstream valid-expect-in-promise / valid 72',
        ],
        [
          testSetup +
            `  });

  {}

  await somePromise;
});`,
          'upstream valid-expect-in-promise / valid 73',
        ],
        [
          `test('promise test', async function () {
  const somePromise = getPromise().then((data) => {
    expect(data).toEqual('foo');
  });

  {
    await somePromise;
  }
});`,
          'upstream valid-expect-in-promise / valid 74',
        ],
        [
          testSetup +
            `  });

  {
    await somePromise;

    somePromise = getPromise().then((data) => {
      expect(data).toEqual('foo');
    });

    await somePromise;
  }
});`,
          'upstream valid-expect-in-promise / valid 75',
        ],
        [
          testSetup +
            `  });

  await somePromise;

  {
    somePromise = getPromise().then((data) => {
      expect(data).toEqual('foo');
    });

    await somePromise;
  }
});`,
          'upstream valid-expect-in-promise / valid 76',
        ],
        [
          testSetup +
            `  });

  somePromise = somePromise.then((data) => {
    expect(data).toEqual('foo');
  });

  await somePromise;
});`,
          'upstream valid-expect-in-promise / valid 77',
        ],
        [
          testSetup +
            `  });

  somePromise = somePromise
    .then((data) => data)
    .then((data) => data)
    .then((data) => {
      expect(data).toEqual('foo');
    });

  await somePromise;
});`,
          'upstream valid-expect-in-promise / valid 78',
        ],
        [
          testSetup +
            `  });

  somePromise = somePromise
    .then((data) => data)
    .then((data) => data)

  await somePromise;
});`,
          'upstream valid-expect-in-promise / valid 79',
        ],
        [
          testSetup +
            `  });

  await somePromise;

  {
    somePromise = getPromise().then((data) => {
      expect(data).toEqual('foo');
    });

    {
      await somePromise;
    }
  }
});`,
          'upstream valid-expect-in-promise / valid 80',
        ],
        [
          `test('promise test', async function () {
  const somePromise = getPromise().then((data) => {
    expect(data).toEqual('foo');
  });

  await Promise.all([somePromise]);
});`,
          'upstream valid-expect-in-promise / valid 81',
        ],
        [
          `test('promise test', async function () {
  const somePromise = getPromise().then((data) => {
    expect(data).toEqual('foo');
  });

  return Promise.all([somePromise]);
});`,
          'upstream valid-expect-in-promise / valid 82',
        ],
        [
          `test('promise test', async function () {
  const somePromise = getPromise().then((data) => {
    expect(data).toEqual('foo');
  });

  return Promise.resolve(somePromise);
});`,
          'upstream valid-expect-in-promise / valid 83',
        ],
        [
          `test('promise test', async function () {
  const somePromise = getPromise().then((data) => {
    expect(data).toEqual('foo');
  });

  await Promise.resolve(somePromise);
});`,
          'upstream valid-expect-in-promise / valid 85',
        ],
        [
          `test('later return', async () => {
  const onePromise = something().then(value => {
    console.log(value);
  });
  const twoPromise = something().then(value => {
    expect(value).toBe('red');
  });

  return Promise.all([onePromise, twoPromise]);
});`,
          'upstream valid-expect-in-promise / valid 87',
        ],
      ] as const
    ).map(([code, name]): ValidTestCase => ({
      name,
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
          `test('case', () => {
        list.push(load().then(value => expect(value).toBe(1)));
      });`,
          9,
          64,
        ],
        [
          `test('case', async () => {
        const pending = [load().then(value => expect(value).toBe(1))];
        await pending;
      });`,
          26,
          69,
        ],
        [
          `test('case', async () => {
        let pending = load().then(value => expect(value).toBe(1));
        pending = pending && fallback;
        await pending;
      });`,
          13,
          66,
        ],
      ] as const
    ).map(([code, column, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'expectInFloatingPromise',
          message:
            'This promise should either be returned or awaited to ensure the expects in its chain are called',
          line: 2,
          column,
          endLine: 2,
          endColumn,
        },
      ],
    })),
    {
      name: 'run 0 invalid 3',
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
        {
          messageId: 'expectInFloatingPromise',
          message:
            'This promise should either be returned or awaited to ensure the expects in its chain are called',
          line: 8,
          column: 9,
          endLine: 8,
          endColumn: 30,
        },
        {
          messageId: 'expectInFloatingPromise',
          message:
            'This promise should either be returned or awaited to ensure the expects in its chain are called',
          line: 11,
          column: 9,
          endLine: 11,
          endColumn: 30,
        },
      ],
    },
    ...(
      [
        [
          `test('case', () => {
        load().then(value => expect(value).toBe(1));
      });`,
          2,
          9,
          2,
          53,
        ],
        [
          `test('case', () => {
        const pending = load().then(value => expect(value).toBe(1));
        return Promise.reject(pending);
      });`,
          2,
          15,
          2,
          68,
        ],
        [
          `test('case', () => {
        const pending = load().then(value => expect(value).toBe(1));
        return Promise.allSettled([pending]);
      });`,
          2,
          15,
          2,
          68,
        ],
        [
          `test('case', async () => {
        const pending = load().then(value => expect(value).toBe(1));
        if (condition) await pending;
      });`,
          2,
          15,
          2,
          68,
        ],
        [
          `test('case', async () => {
        const pending = load().then(value => expect(value).toBe(1));
        try {
          await pending;
        } catch {}
      });`,
          2,
          15,
          2,
          68,
        ],
        [
          `test('case', async () => {
        const pending = load().then(value => expect(value).toBe(1));
        try {
          throw new Error('caught');
        } catch {}
      });`,
          2,
          15,
          2,
          68,
        ],
        [
          `test('case', async () => {
        const pending = load().then(value => expect(value).toBe(1));
        try {
          throw new Error('suppressed');
        } finally {
          return;
        }
      });`,
          2,
          15,
          2,
          68,
        ],
        [
          `function callback() {
        load().then(value => expect(value).toBe(1));
      }
      test('case', callback);`,
          2,
          9,
          2,
          53,
        ],
        [
          `test('case', async () => {
        await first().then(() => {
          second().then(value => expect(value).toBe(1));
        });
      });`,
          3,
          11,
          3,
          57,
        ],
        [
          `test('case', () => {
        load()
          .then(value => expect(value).toBe(1))
          .then(value => expect(value).toBe(2));
      });`,
          2,
          9,
          4,
          49,
        ],
        [
          `test('case', () => {
        const [pending] = load().then(value => expect(value).toBe(1));
      });`,
          2,
          15,
          2,
          70,
        ],
        [
          `test('case', () => {
        const [pending] = [load().then(value => expect(value).toBe(1))];
      });`,
          2,
          28,
          2,
          71,
        ],
        [
          `test('case', () => {
        const { pending } = {
          pending: load().then(value => expect(value).toBe(1)),
        };
      });`,
          3,
          20,
          3,
          63,
        ],
        [
          `test('case', () => {
        consume({
          pending: load().then(value => expect(value).toBe(1)),
        });
      });`,
          3,
          20,
          3,
          63,
        ],
      ] as const
    ).map(([code, line, column, endLine, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index + 4}`,
      code,
      errors: [
        {
          messageId: 'expectInFloatingPromise',
          message:
            'This promise should either be returned or awaited to ensure the expects in its chain are called',
          line,
          column,
          endLine,
          endColumn,
        },
      ],
    })),
    {
      name: 'run 0 invalid 18',
      code: `function handler(value) {
        expect(value).toBe(1);
      }
      test('one', () => { first().then(handler); });
      test('two', () => { second().then(handler); });`,
      errors: [
        {
          messageId: 'expectInFloatingPromise',
          message:
            'This promise should either be returned or awaited to ensure the expects in its chain are called',
          line: 4,
          column: 27,
          endLine: 4,
          endColumn: 49,
        },
        {
          messageId: 'expectInFloatingPromise',
          message:
            'This promise should either be returned or awaited to ensure the expects in its chain are called',
          line: 5,
          column: 27,
          endLine: 5,
          endColumn: 50,
        },
      ],
    },
    ...(
      [
        [
          `test('case', async () => {
        let pending;
        pending ||= load().then(value => expect(value).toBe(1));
      });`,
          3,
          9,
          64,
        ],
        [
          `test('case', async () => {
        let pending = load().then(value => expect(value).toBe(1));
        pending &&= fallback;
        await pending;
      });`,
          2,
          13,
          66,
        ],
      ] as const
    ).map(([code, line, column, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index + 19}`,
      code,
      errors: [
        {
          messageId: 'expectInFloatingPromise',
          message:
            'This promise should either be returned or awaited to ensure the expects in its chain are called',
          line,
          column,
          endLine: line,
          endColumn,
        },
      ],
    })),
    // Upstream 29.16.0 classifies this as valid; current native behavior is invalid.
    ...(
      [
        // Upstream 29.16.0 classifies this as valid; current native behavior is invalid.
        [
          `test('that we bailout if destructuring is used', () => {
  const [promise] = something().then(value => {
    expect(value).toBe('red');
  });
});`,
          'upstream valid-expect-in-promise / valid 48',
          2,
          9,
          4,
          5,
        ],
        // Upstream 29.16.0 classifies this as valid; current native behavior is invalid.
        [
          `test('that we bailout if destructuring is used', () => {
  const [promise] = [
    something().then(value => {
      expect(value).toBe('red');
    })
  ];
});`,
          'upstream valid-expect-in-promise / valid 50',
          3,
          5,
          5,
          7,
        ],
        // Upstream 29.16.0 classifies this as valid; current native behavior is invalid.
        [
          `test('that we bailout if destructuring is used', () => {
  const {promise} = {
    promise: something().then(value => {
      expect(value).toBe('red');
    })
  };
});`,
          'upstream valid-expect-in-promise / valid 51',
          3,
          14,
          5,
          7,
        ],
        // Upstream 29.16.0 classifies this as valid; current native behavior is invalid.
        [
          `test('that we bailout in complex cases', () => {
  promiseSomething({
    timeout: 500,
    promise: something().then(value => {
      expect(value).toBe('red');
    })
  });
});`,
          'upstream valid-expect-in-promise / valid 52',
          4,
          14,
          6,
          7,
        ],
        // Upstream 29.16.0 classifies this as valid; current native behavior is invalid.
        [
          `test('promise test', async function () {
  const somePromise = getPromise().then((data) => {
    expect(data).toEqual('foo');
  });

  return Promise.reject(somePromise);
});`,
          'upstream valid-expect-in-promise / valid 84',
          2,
          9,
          4,
          5,
        ],
        // Upstream 29.16.0 classifies this as valid; current native behavior is invalid.
        [
          `test('promise test', async function () {
  const somePromise = getPromise().then((data) => {
    expect(data).toEqual('foo');
  });

  await Promise.reject(somePromise);
});`,
          'upstream valid-expect-in-promise / valid 86',
          2,
          9,
          4,
          5,
        ],
        // Upstream 29.16.0 classifies this as valid; current native behavior is invalid.
        [
          `test('later return', async () => {
  const onePromise = something().then(value => {
    console.log(value);
  });
  const twoPromise = something().then(value => {
    expect(value).toBe('red');
  });

  return Promise.allSettled([onePromise, twoPromise]);
});`,
          'upstream valid-expect-in-promise / valid 88',
          5,
          9,
          7,
          5,
        ],
      ] as const
    ).map(([code, name, line, column, endLine, endColumn]): TestCase => ({
      name,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'expectInFloatingPromise',
          message:
            'This promise should either be returned or awaited to ensure the expects in its chain are called',
          line,
          column,
          endLine,
          endColumn,
        },
      ],
    })),
    ...(
      [
        [
          `const myFn = () => {
  Promise.resolve().then(() => {
    expect(true).toBe(false);
  });
};

it('it1', () => {
  somePromise.then(() => {
    expect(someThing).toEqual(true);
  });
});`,
          8,
          3,
          10,
          6,
        ],
        [
          `it('it1', () => {
  somePromise.then(() => {
    expect(someThing).toEqual(true);
  });
});`,
          2,
          3,
          4,
          6,
        ],
        [
          `it('it1', () => {
  somePromise.finally(() => {
    expect(someThing).toEqual(true);
  });
});`,
          2,
          3,
          4,
          6,
        ],
        [
          `
       it('it1', () => {
         somePromise['then'](() => {
           expect(someThing).toEqual(true);
         });
       });
     \u0020`,
          3,
          10,
          5,
          13,
        ],
        [
          `it('it1', function() {
  getSomeThing().getPromise().then(function() {
    expect(someThing).toEqual(true);
  });
});`,
          2,
          3,
          4,
          6,
        ],
        [
          `it('it1', function() {
  Promise.resolve().then(function() {
    expect(someThing).toEqual(true);
  });
});`,
          2,
          3,
          4,
          6,
        ],
        [
          `it('it1', function() {
  somePromise.catch(function() {
    expect(someThing).toEqual(true)
  })
})`,
          2,
          3,
          4,
          5,
        ],
        [
          `xtest('it1', function() {
  somePromise.catch(function() {
    expect(someThing).toEqual(true)
  })
})`,
          2,
          3,
          4,
          5,
        ],
        [
          `it('it1', function() {
  somePromise.then(function() {
    expect(someThing).toEqual(true)
  })
})`,
          2,
          3,
          4,
          5,
        ],
        [
          `it('it1', function () {
  Promise.resolve().then(/*fulfillment*/ function () {
    expect(someThing).toEqual(true);
  }, /*rejection*/ function () {
    expect(someThing).toEqual(true);
  })
})`,
          2,
          3,
          6,
          5,
        ],
        [
          `it('it1', function () {
  Promise.resolve().then(/*fulfillment*/ function () {
  }, /*rejection*/ function () {
    expect(someThing).toEqual(true)
  })
});`,
          2,
          3,
          5,
          5,
        ],
        [
          `it('test function', () => {
  Builder.getPromiseBuilder()
    .get()
    .build()
    .then(data => expect(data).toEqual('Hi'));
});`,
          2,
          3,
          5,
          47,
        ],
        [
          `
        it('test function', async () => {
          Builder.getPromiseBuilder()
            .get()
            .build()
            .then(data => expect(data).toEqual('Hi'));
        });
     \u0020`,
          3,
          11,
          6,
          55,
        ],
        [
          `it('it1', () => {
  somePromise.then(() => {
    doSomeOperation();
    expect(someThing).toEqual(true);
  })
});`,
          2,
          3,
          5,
          5,
        ],
        [
          `it('is a test', () => {
  somePromise
    .then(() => {})
    .then(() => expect(someThing).toEqual(value))
});`,
          2,
          3,
          4,
          50,
        ],
        [
          `it('is a test', () => {
  somePromise
    .then(() => expect(someThing).toEqual(value))
    .then(() => {})
});`,
          2,
          3,
          4,
          20,
        ],
        [
          `it('is a test', () => {
  somePromise.then(() => {
    return value;
  })
  .then(value => {
    expect(someThing).toEqual(value);
  })
});`,
          2,
          3,
          7,
          5,
        ],
        [
          `it('is a test', () => {
  somePromise.then(() => {
    expect(someThing).toEqual(true);
  })
  .then(() => {
    console.log('this is silly');
  })
});`,
          2,
          3,
          7,
          5,
        ],
        [
          `it('is a test', () => {
  somePromise.then(() => {
    // return value;
  })
  .then(value => {
    expect(someThing).toEqual(value);
  })
});`,
          2,
          3,
          7,
          5,
        ],
        [
          `it('is a test', () => {
  somePromise.then(() => {
    return value;
  })
  .then(value => {
    expect(someThing).toEqual(value);
  })

  return anotherPromise.then(() => expect(x).toBe(y));
});`,
          2,
          3,
          7,
          5,
        ],
        [
          `it('is a test', () => {
  somePromise
    .then(() => 1)
    .then(x => x + 1)
    .catch(() => -1)
    .then(v => expect(v).toBe(2));

  return anotherPromise.then(() => expect(x).toBe(y));
});`,
          2,
          3,
          6,
          35,
        ],
        [
          `it('is a test', () => {
  somePromise
    .then(() => 1)
    .then(v => expect(v).toBe(2))
    .then(x => x + 1)
    .catch(() => -1);

  return anotherPromise.then(() => expect(x).toBe(y));
});`,
          2,
          3,
          6,
          22,
        ],
        [
          `it('it1', () => {
  somePromise.finally(() => {
    doSomeOperation();
    expect(someThing).toEqual(true);
  })
});`,
          2,
          3,
          5,
          5,
        ],
        [
          `test('invalid return', () => {
  const promise = something().then(value => {
    const foo = "foo";
    return expect(value).toBe('red');
  });
});`,
          2,
          9,
          5,
          5,
        ],
        [
          `fit('it1', () => {
  somePromise.then(() => {
    doSomeOperation();
    expect(someThing).toEqual(true);
  })
});`,
          2,
          3,
          5,
          5,
        ],
        [
          `it.skip('it1', () => {
  somePromise.then(() => {
    doSomeOperation();
    expect(someThing).toEqual(true);
  })
});`,
          2,
          3,
          5,
          5,
        ],
        [
          testSetup3 +
            `  promise;
});`,
          2,
          9,
          4,
          5,
        ],
        [
          testSetup3 +
            `  return;

  await promise;
});`,
          2,
          9,
          4,
          5,
        ],
        [
          testSetup3 +
            `  return 1;

  await promise;
});`,
          2,
          9,
          4,
          5,
        ],
        [
          testSetup3 +
            `  return [];

  await promise;
});`,
          2,
          9,
          4,
          5,
        ],
        [
          testSetup3 +
            `  return Promise.all([anotherPromise]);

  await promise;
});`,
          2,
          9,
          4,
          5,
        ],
        [
          testSetup3 +
            `  return {};

  await promise;
});`,
          2,
          9,
          4,
          5,
        ],
        [
          testSetup3 +
            `  return Promise.all([]);

  await promise;
});`,
          2,
          9,
          4,
          5,
        ],
        [
          testSetup3 +
            `  await 1;
});`,
          2,
          9,
          4,
          5,
        ],
        [
          testSetup3 +
            `  await [];
});`,
          2,
          9,
          4,
          5,
        ],
        [
          testSetup3 +
            `  await Promise.all([anotherPromise]);
});`,
          2,
          9,
          4,
          5,
        ],
        [
          testSetup3 +
            `  await {};
});`,
          2,
          9,
          4,
          5,
        ],
        [
          testSetup3 +
            `  await Promise.all([]);
});`,
          2,
          9,
          4,
          5,
        ],
        [
          `test('later return', async () => {
  const promise = something().then(value => {
    expect(value).toBe('red');
  }), x = 1;
});`,
          2,
          9,
          4,
          5,
        ],
        [
          `test('later return', async () => {
  const x = 1, promise = something().then(value => {
    expect(value).toBe('red');
  });
});`,
          2,
          16,
          4,
          5,
        ],
      ] as const
    ).map(([code, line, column, endLine, endColumn], index): TestCase => ({
      name: `upstream valid-expect-in-promise / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'expectInFloatingPromise',
          message:
            'This promise should either be returned or awaited to ensure the expects in its chain are called',
          line,
          column,
          endLine,
          endColumn,
        },
      ],
    })),
    {
      name: 'upstream valid-expect-in-promise / invalid 40',
      code: `import { test } from '@jest/globals';

test('later return', async () => {
  const x = 1, promise = something().then(value => {
    expect(value).toBe('red');
  });
});`,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'expectInFloatingPromise',
          message:
            'This promise should either be returned or awaited to ensure the expects in its chain are called',
          line: 4,
          column: 16,
          endLine: 6,
          endColumn: 5,
        },
      ],
    },
    ...(
      [
        [
          `it('promise test', () => {
  const somePromise = getThatPromise();
  somePromise.then((data) => {
    expect(data).toEqual('foo');
  });
  expect(somePromise).toBeDefined();
  return somePromise;
});`,
          3,
          3,
          5,
          6,
        ],
        [
          `test('promise test', function () {
  let somePromise = getThatPromise();
  somePromise.then((data) => {
    expect(data).toEqual('foo');
  });
  expect(somePromise).toBeDefined();
  return somePromise;
});`,
          3,
          3,
          5,
          6,
        ],
        [
          testSetup +
            `  });

  somePromise = null;

  await somePromise;
});`,
          2,
          7,
          4,
          5,
        ],
        [
          testSetup +
            `  });

  somePromise = getPromise().then((data) => {
    expect(data).toEqual('foo');
  });

  await somePromise;
});`,
          2,
          7,
          4,
          5,
        ],
        [
          testSetup +
            `  });

  ({ somePromise } = {})
});`,
          2,
          7,
          4,
          5,
        ],
        [
          testSetup +
            `  });

  {
    somePromise = getPromise().then((data) => {
      expect(data).toEqual('foo');
    });

    await somePromise;
  }
});`,
          2,
          7,
          4,
          5,
        ],
        [
          `test('that we error on this destructuring', async () => {
  [promise] = something().then(value => {
    expect(value).toBe('red');
  });
});`,
          2,
          3,
          4,
          5,
        ],
        [
          `test('that we error on this', () => {
  const promise = something().then(value => {
    expect(value).toBe('red');
  });

  log(promise);
});`,
          2,
          9,
          4,
          5,
        ],
        [
          testSetup2 +
            `  expect(promise).toBeInstanceOf(Promise);
});`,
          2,
          9,
          6,
          5,
        ],
        [
          testSetup2 +
            `  expect(anotherPromise).resolves.toBe(1);
});`,
          2,
          9,
          6,
          5,
        ],
      ] as const
    ).map(([code, line, column, endLine, endColumn], index): TestCase => ({
      name: `upstream valid-expect-in-promise / invalid ${index + 41}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'expectInFloatingPromise',
          message:
            'This promise should either be returned or awaited to ensure the expects in its chain are called',
          line,
          column,
          endLine,
          endColumn,
        },
      ],
    })),
    {
      name: 'upstream valid-expect-in-promise / invalid 51',
      code: `import { it as promiseThatThis } from '@jest/globals';

promiseThatThis('is valid', async () => {
  const promise = loadNumber().then(number => {
    expect(typeof number).toBe('number');

    return number + 1;
  });

  expect(anotherPromise).resolves.toBe(1);
});`,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'expectInFloatingPromise',
          message:
            'This promise should either be returned or awaited to ensure the expects in its chain are called',
          line: 4,
          column: 9,
          endLine: 8,
          endColumn: 5,
        },
      ],
    },
    {
      name: 'upstream valid-expect-in-promise / invalid 52',
      code: `promiseThatThis('is valid', async () => {
  const promise = loadNumber().then(number => {
    expect(typeof number).toBe('number');

    return number + 1;
  });

  expect(anotherPromise).resolves.toBe(1);
});`,
      settings: {
        jest: {
          globalAliases: {
            xit: ['promiseThatThis'],
          },
        },
      },
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'expectInFloatingPromise',
          message:
            'This promise should either be returned or awaited to ensure the expects in its chain are called',
          line: 2,
          column: 9,
          endLine: 6,
          endColumn: 5,
        },
      ],
    },
  ],
});
