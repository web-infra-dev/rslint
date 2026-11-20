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
        "test('something', () => Promise.resolve().then(() => expect(1).toBe(2)));",
        'Promise.resolve().then(() => expect(1).toBe(2))',
        'const x = Promise.resolve().then(() => expect(1).toBe(2))',
        'it.todo("something")',
        `it('is valid', () => {
  const promise = loadNumber().then(number => {
    expect(typeof number).toBe('number');

    return number + 1;
  });

  expect(promise).resolves.toBe(1);
});`,
        `it('is valid', () => {
  const promise = loadNumber().then(number => {
    expect(typeof number).toBe('number');

    return number + 1;
  });

  expect(promise).resolves.not.toBe(2);
});`,
        `it('is valid', () => {
  const promise = loadNumber().then(number => {
    expect(typeof number).toBe('number');

    return number + 1;
  });

  expect(promise).rejects.toBe(1);
});`,
        `it('is valid', () => {
  const promise = loadNumber().then(number => {
    expect(typeof number).toBe('number');

    return number + 1;
  });

  expect(promise).rejects.not.toBe(2);
});`,
        testSetup2 +
          `  expect(await promise).toBeGreaterThan(1);
});`,
        testSetup2 +
          `  expect(await promise).resolves.toBeGreaterThan(1);
});`,
        testSetup2 +
          `  expect(1).toBeGreaterThan(await promise);
});`,
        testSetup2 +
          `  expect.this.that.is(await promise);
});`,
        `it('is valid', async () => {
  expect(await loadNumber().then(number => {
    expect(typeof number).toBe('number');

    return number + 1;
  })).toBeGreaterThan(1);
});`,
        testSetup2 +
          `  expect([await promise]).toHaveLength(1);
});`,
        testSetup2 +
          `  expect([,,await promise,,]).toHaveLength(1);
});`,
        testSetup2 +
          `  expect([[await promise]]).toHaveLength(1);
});`,
        testSetup2 +
          `  logValue(await promise);
});`,
        `it('is valid', async () => {
  const promise = loadNumber().then(number => {
    expect(typeof number).toBe('number');

    return 1;
  });

  expect.assertions(await promise);
});`,
        `it('is valid', async () => {
  await loadNumber().then(number => {
    expect(typeof number).toBe('number');
  });
});`,
        `it('it1', () => new Promise((done) => {
  test()
    .then(() => {
      expect(someThing).toEqual(true);
      done();
    });
}));`,
        `it('it1', () => {
  return new Promise(done => {
    test().then(() => {
      expect(someThing).toEqual(true);
      done();
    });
  });
});`,
        `it('passes', () => {
  Promise.resolve().then(() => {
    grabber.grabSomething();
  });
});`,
        `it('passes', async () => {
  const grabbing = Promise.resolve().then(() => {
    grabber.grabSomething();
  });

  await grabbing;

  expect(grabber.grabbedItems).toHaveLength(1);
});`,
        `const myFn = () => {
  Promise.resolve().then(() => {
    expect(true).toBe(false);
  });
};`,
        `const myFn = () => {
  Promise.resolve().then(() => {
    subject.invokeMethod();
  });
};`,
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
        `it('it1', () => new Promise((done) => {
  test()
    .finally(() => {
      expect(someThing).toEqual(true);
      done();
    });
}));`,
        `it('it1', () => {
  return somePromise.then(() => {
    expect(someThing).toEqual(true);
  });
});`,
        `it('it1', () => {
  return somePromise.finally(() => {
    expect(someThing).toEqual(true);
  });
});`,
        `it('it1', function() {
  return somePromise.catch(function() {
    expect(someThing).toEqual(true);
  });
});`,
        `xtest('it1', function() {
  return somePromise.catch(function() {
    expect(someThing).toEqual(true);
  });
});`,
        `it('it1', function() {
  return somePromise.then(function() {
    doSomeThingButNotExpect();
  });
});`,
        `it('it1', function() {
  return getSomeThing().getPromise().then(function() {
    expect(someThing).toEqual(true);
  });
});`,
        `it('it1', function() {
  return Promise.resolve().then(function() {
    expect(someThing).toEqual(true);
  });
});`,
        `it('it1', function () {
  return Promise.resolve().then(function () {
    /*fulfillment*/
    expect(someThing).toEqual(true);
  }, function () {
    /*rejection*/
    expect(someThing).toEqual(true);
  });
});`,
        `it('it1', function () {
  Promise.resolve().then(/*fulfillment*/ function () {
  }, undefined, /*rejection*/ function () {
    expect(someThing).toEqual(true)
  })
});`,
        `it('it1', function () {
  return Promise.resolve().then(function () {
    /*fulfillment*/
  }, function () {
    /*rejection*/
    expect(someThing).toEqual(true);
  });
});`,
        `it('it1', function () {
  return somePromise.then()
});`,
        `it('it1', async () => {
  await Promise.resolve().then(function () {
    expect(someThing).toEqual(true)
  });
});`,
        `it('it1', async () => {
  await somePromise.then(() => {
    expect(someThing).toEqual(true)
  });
});`,
        `it('it1', async () => {
  await getSomeThing().getPromise().then(function () {
    expect(someThing).toEqual(true)
  });
});`,
        `it('it1', () => {
  return somePromise.then(() => {
    expect(someThing).toEqual(true);
  })
  .then(() => {
    expect(someThing).toEqual(true);
  })
});`,
        `it('it1', () => {
  return somePromise.then(() => {
    return value;
  })
  .then(value => {
    expect(someThing).toEqual(value);
  })
});`,
        `it('it1', () => {
  return somePromise.then(() => {
    expect(someThing).toEqual(true);
  })
  .then(() => {
    console.log('this is silly');
  })
});`,
        `it('it1', () => {
  return somePromise.then(() => {
    expect(someThing).toEqual(true);
  })
  .catch(() => {
    expect(someThing).toEqual(false);
  })
});`,
        `test('later return', () => {
  const promise = something().then(value => {
    expect(value).toBe('red');
  });

  return promise;
});`,
        testSetup3 +
          `  await promise;
});`,
        `test.only('later return', () => {
  const promise = something().then(value => {
    expect(value).toBe('red');
  });

  return promise;
});`,
        `test('that we bailout if destructuring is used', async () => {
  const [promise] = await something().then(value => {
    expect(value).toBe('red');
  });
});`,
        `it('shorthand arrow', () =>
  something().then(value => {
    expect(() => {
      value();
    }).toThrow();
  })
);`,
        `it('crawls for files based on patterns', () => {
  const promise = nodeCrawl({}).then(data => {
    expect(childProcess.spawn).lastCalledWith('find');
  });
  return promise;
});`,
        `it('is a test', async () => {
  const value = await somePromise().then(response => {
    expect(response).toHaveProperty('data');

    return response.data;
  });

  expect(value).toBe('hello world');
});`,
        `it('is a test', async () => {
  return await somePromise().then(response => {
    expect(response).toHaveProperty('data');

    return response.data;
  });
});`,
        `it('is a test', async () => {
  return somePromise().then(response => {
    expect(response).toHaveProperty('data');

    return response.data;
  });
});`,
        `it('is a test', async () => {
  await somePromise().then(response => {
    expect(response).toHaveProperty('data');

    return response.data;
  });
});`,
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
        `it("it1", () => somePromise.then(() => {
  expect(someThing).toEqual(true)
}))`,
        'it("it1", () => somePromise.then(() => expect(someThing).toEqual(true)))',
        `it('promise test with done', (done) => {
  const promise = getPromise();
  promise.then(() => expect(someThing).toEqual(true));
});`,
        `it('name of done param does not matter', (nameDoesNotMatter) => {
  const promise = getPromise();
  promise.then(() => expect(someThing).toEqual(true));
});`,
        `it.each([])('name of done param does not matter', (nameDoesNotMatter) => {
  const promise = getPromise();
  promise.then(() => expect(someThing).toEqual(true));
});`,
        `it.each\`\`('name of done param does not matter', ({}, nameDoesNotMatter) => {
  const promise = getPromise();
  promise.then(() => expect(someThing).toEqual(true));
});`,
        `test('valid-expect-in-promise', async () => {
  const text = await fetch('url')
      .then(res => res.text())
      .then(text => text);

  expect(text).toBe('text');
});`,
        testSetup +
          `  }), x = 1;

  await somePromise;
});`,
        `test('promise test', async function () {
  let x = 1, somePromise = getPromise().then((data) => {
    expect(data).toEqual('foo');
  });

  await somePromise;
});`,
        testSetup +
          `  });

  await somePromise;

  somePromise = getPromise().then((data) => {
    expect(data).toEqual('foo');
  });

  await somePromise;
});`,
        testSetup +
          `  });

  await somePromise;

  somePromise = getPromise().then((data) => {
    expect(data).toEqual('foo');
  });

  return somePromise;
});`,
        testSetup +
          `  });

  {}

  await somePromise;
});`,
        `test('promise test', async function () {
  const somePromise = getPromise().then((data) => {
    expect(data).toEqual('foo');
  });

  {
    await somePromise;
  }
});`,
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
        testSetup +
          `  });

  somePromise = somePromise.then((data) => {
    expect(data).toEqual('foo');
  });

  await somePromise;
});`,
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
        testSetup +
          `  });

  somePromise = somePromise
    .then((data) => data)
    .then((data) => data)

  await somePromise;
});`,
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
        `test('promise test', async function () {
  const somePromise = getPromise().then((data) => {
    expect(data).toEqual('foo');
  });

  await Promise.all([somePromise]);
});`,
        `test('promise test', async function () {
  const somePromise = getPromise().then((data) => {
    expect(data).toEqual('foo');
  });

  return Promise.all([somePromise]);
});`,
        `test('promise test', async function () {
  const somePromise = getPromise().then((data) => {
    expect(data).toEqual('foo');
  });

  return Promise.resolve(somePromise);
});`,
        `test('promise test', async function () {
  const somePromise = getPromise().then((data) => {
    expect(data).toEqual('foo');
  });

  await Promise.resolve(somePromise);
});`,
        `test('later return', async () => {
  const onePromise = something().then(value => {
    console.log(value);
  });
  const twoPromise = something().then(value => {
    expect(value).toBe('red');
  });

  return Promise.all([onePromise, twoPromise]);
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
  ],
  invalid: [
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
        [
          `import { test } from '@jest/globals';

test('later return', async () => {
  const x = 1, promise = something().then(value => {
    expect(value).toBe('red');
  });
});`,
          4,
          16,
          6,
          5,
        ],
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
        [
          `import { it as promiseThatThis } from '@jest/globals';

promiseThatThis('is valid', async () => {
  const promise = loadNumber().then(number => {
    expect(typeof number).toBe('number');

    return number + 1;
  });

  expect(anotherPromise).resolves.toBe(1);
});`,
          4,
          9,
          8,
          5,
        ],
      ] as const
    ).map(([code, line, column, endLine, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index}`,
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
      name: 'run 0 invalid 52',
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
