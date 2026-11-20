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
const numbersLoopBody = `
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});`;
const singleAssertionTest = `it("is a number that is greater than four", () => {
  expect(number).toBeGreaterThan(4);
});`;
const helperAssertionBody = `
  const expectNumbersToBeGreaterThan = (numbers, value) => {
    for (let number of numbers) {
      expect(number).toBeGreaterThan(value);
    }
  };

  expectNumbersToBeGreaterThan(getNumbers(), 2);
});`;
const numbersAboveFiveTest = `
it("returns numbers that are greater than five", () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(5);
  }
});`;
const suiteWithAssertionHook = `describe('some tests', () => {
  beforeEach(() => { expect.hasAssertions(); });
  it("it1", function() {
    someFunctionToDo();
    someFunctionToDo2();
  });
});`;
const requestCallbackBody = `
  client.get('/user', response => {
    expect(response.status).toBe(200);
  });
});`;
const numberAboveFiveTest = `
it("returns numbers that are greater than five", () => {
  expect(number).toBeGreaterThan(5);
});`;
const requestCallbackWithAssertionCount = `  const expectOkResponse = response => {
    expect.assertions(1);

    expect(response.status).toBe(200);
  };

  client.get('/user', expectOkResponse);
});`;

new RuleTester(fixtureConfig()).run('prefer-expect-assertions', {
  valid: [
    ...(
      [
        'test("nonsense", [])',
        'test("it1", () => {expect.assertions(0);})',
        'test("it1", function() {expect.assertions(0);})',
        'test("it1", function() {expect.hasAssertions();})',
        'it("it1", function() {expect.assertions(0);})',
        `it("it1", function() {
  expect.assertions(1);
  expect(someValue).toBe(true)
});`,
        'test("it1")',
        'itHappensToStartWithIt("foo", function() {})',
        'testSomething("bar", function() {})',
        'it(async () => {expect.assertions(0);})',
        `it("returns numbers that are greater than four", function() {
  expect.assertions(2);

  for(let thing in things) {
    expect(number).toBeGreaterThan(4);
  }
});`,
        `it("returns numbers that are greater than four", function() {
  expect.hasAssertions();

  for (let i = 0; i < things.length; i++) {
    expect(number).toBeGreaterThan(4);
  }
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    ...(
      [
        `it("it1", async () => {
  expect.assertions(1);
  expect(someValue).toBe(true)
});`,
        `it("it1", function() {
  expect(someValue).toBe(true)
});`,
        'it("it1", () => {})',
        `it("returns numbers that are greater than four", async () => {
  expect.assertions(2);

  for(let thing in things) {
    expect(number).toBeGreaterThan(4);
  }
});`,
        `it("returns numbers that are greater than four", () => {
  for(let thing in things) {
    expect(number).toBeGreaterThan(4);
  }
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index + 12}`,
      code,
      options: [
        {
          onlyFunctionsWithAsyncKeyword: true,
        },
      ],
    })),
    ...(
      [
        `import { expect as pleaseExpect } from '@jest/globals';

it("returns numbers that are greater than four", function() {
  pleaseExpect.assertions(2);

  for(let thing in things) {
    pleaseExpect(number).toBeGreaterThan(4);
  }
});`,
        `beforeEach(() => expect.hasAssertions());

it('responds ok', function () {` +
          requestCallbackBody +
          `

` +
          singleAssertionTest,
        `afterEach(() => {
  expect.hasAssertions();
});

it('responds ok', function () {` +
          requestCallbackBody +
          `

` +
          singleAssertionTest,
        `afterEach(() => {
  expect.hasAssertions();
});

it('responds ok', function () {` +
          requestCallbackBody +
          `

it("is a number that is greater than four", () => {
  expect.hasAssertions();

  expect(number).toBeGreaterThan(4);
});`,
        `beforeEach(() => { expect.hasAssertions(); });

describe('my tests', () => {
  it('responds ok', function () {
    client.get('/user', response => {
      expect(response.status).toBe(200);
    });
  });

  it("is a number that is greater than four", () => {
    expect.hasAssertions();

    expect(number).toBeGreaterThan(4);
  });
});`,
        `describe('my tests', () => {
  beforeEach(() => { expect.hasAssertions(); });

  describe('left', () => {
    describe('inner', () => {
      it('responds ok', function () {
        client.get('/user', response => {
          expect(response.status).toBe(200);
        });
      });
    });
  });

  describe('right', () => {
    it("is a number that is greater than four", () => {
      expect(number).toBeGreaterThan(4);
    });
  });
});`,
        `describe('my tests', () => {
  beforeEach(() => { expect.hasAssertions(); });

  describe('left', () => {
    it('responds ok', function () {
      client.get('/user', response => {
        expect(response.status).toBe(200);
      });
    });
  });

  describe('right', () => {
    it("is a number that is greater than four", () => {
      expect(number).toBeGreaterThan(4);
    });
  });
});`,
        `describe('my tests', () => {
  beforeEach(() => { expect.hasAssertions(); });

  describe('left', () => {
    beforeEach(() => { expect.hasAssertions(); });

    it('responds ok', function () {
      client.get('/user', response => {
        expect(response.status).toBe(200);
      });
    });
  });

  describe('right', () => {
    it("is a number that is greater than four", () => {
      expect(number).toBeGreaterThan(4);
    });
  });
});`,
        `describe('my tests', () => {
  beforeEach(() => { expect.hasAssertions(); });

  describe('left', () => {
    afterEach(() => { expect.hasAssertions(); });

    it('responds ok', function () {
      client.get('/user', response => {
        expect(response.status).toBe(200);
      });
    });
  });

  describe('right', () => {
    it("is a number that is greater than four", () => {
      expect(number).toBeGreaterThan(4);
    });
  });
});`,
        `describe('my tests', () => {
  beforeEach(() => { expect.hasAssertions(); });

  it('responds ok', function () {
    client.get('/user', response => {
      expect(response.status).toBe(200);
    });
  });

  it("is a number that is greater than four", () => {
    expect.hasAssertions();

    expect(number).toBeGreaterThan(4);
  });
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index + 17}`,
      code,
    })),
    ...(
      [
        `const expectNumbersToBeGreaterThan = (numbers, value) => {
  for (let number of numbers) {
    expect(number).toBeGreaterThan(value);
  }
};

it('returns numbers that are greater than two', function () {
  expectNumbersToBeGreaterThan(getNumbers(), 2);
});`,
        `it("returns numbers that are greater than five", function () {
  expect.assertions(2);

  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(5);
  }
});`,
        `it("returns things that are less than ten", function () {
  expect.hasAssertions();

  for (const thing in things) {
    expect(thing).toBeLessThan(10);
  }
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index + 27}`,
      code,
      options: [
        {
          onlyFunctionsWithExpectInLoop: true,
        },
      ],
    })),
    ...(
      [
        `const expectNumbersToBeGreaterThan = (numbers, value) => {
  numbers.forEach(number => {
    expect(number).toBeGreaterThan(value);
  });
};

it('returns numbers that are greater than two', function () {
  expectNumbersToBeGreaterThan(getNumbers(), 2);
});`,
        `it('returns numbers that are greater than two', function () {
  expect.assertions(2);
` + helperAssertionBody,
        `beforeEach(() => expect.hasAssertions());

it('returns numbers that are greater than two', function () {` +
          helperAssertionBody,
        `it("returns numbers that are greater than five", function () {
  expect.assertions(2);

  getNumbers().forEach(number => {
    expect(number).toBeGreaterThan(5);
  });
});`,
        `it("returns things that are less than ten", function () {
  expect.hasAssertions();

  things.forEach(thing => {
    expect(thing).toBeLessThan(10);
  });
});`,
        `it('sends the data as a string', () => {
  expect.hasAssertions();

  const stream = openStream();

  stream.on('data', data => {
    expect(data).toBe(expect.any(String));
  });
});`,
        `it('responds ok', function () {
  expect.assertions(1);
` + requestCallbackBody,
        `it.each([1, 2, 3])("returns ok", id => {
  expect.assertions(3);

  client.get(\`/users/\${id}\`, response => {
    expect(response.status).toBe(200);
  });
});

` + singleAssertionTest,
        `it('is a test', () => {
  expect(expected).toBe(actual);
});

describe('my test', () => {
  it('is another test', () => {
    expect(expected).toBe(actual);
  });
});`,
        `it('responds ok', function () {
  expect.assertions(1);
` +
          requestCallbackBody +
          `

describe('my test', () => {
  beforeEach(() => expect.hasAssertions());

  it('responds ok', function () {
    client.get('/user', response => {
      expect(response.status).toBe(200);
    });
  });
});`,
        `it('responds ok', function () {
  expect.assertions(1);
` +
          requestCallbackBody +
          `

describe('my test', () => {
  afterEach(() => expect.hasAssertions());

  it('responds ok', function () {
    client.get('/user', response => {
      expect(response.status).toBe(200);
    });
  });
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index + 30}`,
      code,
      options: [
        {
          onlyFunctionsWithExpectInCallback: true,
        },
      ],
    })),
    ...(
      [
        `it('only returns numbers that are greater than zero', async () => {
  expect.hasAssertions();

  for (const number of await getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});`,
        `it('only returns numbers that are greater than zero', async () => {
  expect.assertions(2);

  for (const number of await getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index + 41}`,
      code,
      options: [
        {
          onlyFunctionsWithAsyncKeyword: true,
          onlyFunctionsWithExpectInLoop: true,
        },
      ],
    })),
    ...(
      [
        'test.each()("is fine", () => { expect.assertions(0); })',
        'test.each``("is fine", () => { expect.assertions(0); })',
        'test.each()("is fine", () => { expect.hasAssertions(); })',
        'test.each``("is fine", () => { expect.hasAssertions(); })',
        'it.each()("is fine", () => { expect.assertions(0); })',
        'it.each``("is fine", () => { expect.assertions(0); })',
        'it.each()("is fine", () => { expect.hasAssertions(); })',
        'it.each``("is fine", () => { expect.hasAssertions(); })',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index + 43}`,
      code,
    })),
    ...(
      [
        'test.each()("is fine", () => {})',
        'test.each``("is fine", () => {})',
        'it.each()("is fine", () => {})',
        'it.each``("is fine", () => {})',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index + 51}`,
      code,
      options: [
        {
          onlyFunctionsWithAsyncKeyword: true,
        },
      ],
    })),
    ...(
      [
        `describe.each(['hello'])('%s', () => {
  it('is fine', () => {
    expect.assertions(0);
  });
});`,
        `describe.each\`\`('%s', () => {
  it('is fine', () => {
    expect.assertions(0);
  });
});`,
        `describe.each(['hello'])('%s', () => {
  it('is fine', () => {
    expect.hasAssertions();
  });
});`,
        `describe.each\`\`('%s', () => {
  it('is fine', () => {
    expect.hasAssertions();
  });
});`,
        `describe.each(['hello'])('%s', () => {
  it.each()('is fine', () => {
    expect.assertions(0);
  });
});`,
        `describe.each\`\`('%s', () => {
  it.each()('is fine', () => {
    expect.assertions(0);
  });
});`,
        `describe.each(['hello'])('%s', () => {
  it.each()('is fine', () => {
    expect.hasAssertions();
  });
});`,
        `describe.each\`\`('%s', () => {
  it.each()('is fine', () => {
    expect.hasAssertions();
  });
});`,
        `describe('some tests', () => {
  it("it1", function() {
    someFunctionToDo();
    someFunctionToDo2();
  });

  beforeEach(() => { expect.hasAssertions(); });

  it("it1", function() {
    someFunctionToDo();
    someFunctionToDo2();
  });
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index + 55}`,
      code,
    })),
    ...(
      [
        'test("nonsense", [])',
        'test("it1", () => {expect.assertions(0);})',
        'test("it1", function() {expect.assertions(0);})',
        'test("it1", function() {expect.hasAssertions();})',
        'it("it1", function() {expect.assertions(0);})',
        `it("it1", function() {
  expect.assertions(1);
  expect(someValue).toBe(true)
});`,
        'test("it1")',
        'itHappensToStartWithIt("foo", function() {})',
        'testSomething("bar", function() {})',
        'it(async () => {expect.assertions(0);})',
        `it("returns numbers that are greater than four", function() {
  expect.assertions(2);

  for(let thing in things) {
    expect(number).toBeGreaterThan(4);
  }
});`,
        `it("returns numbers that are greater than four", function() {
  expect.hasAssertions();

  for (let i = 0; i < things.length; i++) {
    expect(number).toBeGreaterThan(4);
  }
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream prefer-expect-assertions / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
    })),
    ...(
      [
        `it("it1", async () => {
  expect.assertions(1);
  expect(someValue).toBe(true)
});`,
        `it("it1", function() {
  expect(someValue).toBe(true)
});`,
        'it("it1", () => {})',
        `it("returns numbers that are greater than four", async () => {
  expect.assertions(2);

  for(let thing in things) {
    expect(number).toBeGreaterThan(4);
  }
});`,
        `it("returns numbers that are greater than four", () => {
  for(let thing in things) {
    expect(number).toBeGreaterThan(4);
  }
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream prefer-expect-assertions / valid ${index + 12}`,
      code,
      options: [
        {
          onlyFunctionsWithAsyncKeyword: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
    })),
    {
      name: 'upstream prefer-expect-assertions / valid 17',
      code: `import { expect as pleaseExpect } from '@jest/globals';

it("returns numbers that are greater than four", function() {
  pleaseExpect.assertions(2);

  for(let thing in things) {
    pleaseExpect(number).toBeGreaterThan(4);
  }
});`,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2017,
      },
    },
    ...(
      [
        `beforeEach(() => expect.hasAssertions());

it('responds ok', function () {` +
          requestCallbackBody +
          `

` +
          singleAssertionTest,
        `afterEach(() => {
  expect.hasAssertions();
});

it('responds ok', function () {` +
          requestCallbackBody +
          `

` +
          singleAssertionTest,
        `afterEach(() => {
  expect.hasAssertions();
});

it('responds ok', function () {` +
          requestCallbackBody +
          `

it("is a number that is greater than four", () => {
  expect.hasAssertions();

  expect(number).toBeGreaterThan(4);
});`,
        `beforeEach(() => { expect.hasAssertions(); });

describe('my tests', () => {
  it('responds ok', function () {
    client.get('/user', response => {
      expect(response.status).toBe(200);
    });
  });

  it("is a number that is greater than four", () => {
    expect.hasAssertions();

    expect(number).toBeGreaterThan(4);
  });
});`,
        `describe('my tests', () => {
  beforeEach(() => { expect.hasAssertions(); });

  describe('left', () => {
    describe('inner', () => {
      it('responds ok', function () {
        client.get('/user', response => {
          expect(response.status).toBe(200);
        });
      });
    });
  });

  describe('right', () => {
    it("is a number that is greater than four", () => {
      expect(number).toBeGreaterThan(4);
    });
  });
});`,
        `describe('my tests', () => {
  beforeEach(() => { expect.hasAssertions(); });

  describe('left', () => {
    it('responds ok', function () {
      client.get('/user', response => {
        expect(response.status).toBe(200);
      });
    });
  });

  describe('right', () => {
    it("is a number that is greater than four", () => {
      expect(number).toBeGreaterThan(4);
    });
  });
});`,
        `describe('my tests', () => {
  beforeEach(() => { expect.hasAssertions(); });

  describe('left', () => {
    beforeEach(() => { expect.hasAssertions(); });

    it('responds ok', function () {
      client.get('/user', response => {
        expect(response.status).toBe(200);
      });
    });
  });

  describe('right', () => {
    it("is a number that is greater than four", () => {
      expect(number).toBeGreaterThan(4);
    });
  });
});`,
        `describe('my tests', () => {
  beforeEach(() => { expect.hasAssertions(); });

  describe('left', () => {
    afterEach(() => { expect.hasAssertions(); });

    it('responds ok', function () {
      client.get('/user', response => {
        expect(response.status).toBe(200);
      });
    });
  });

  describe('right', () => {
    it("is a number that is greater than four", () => {
      expect(number).toBeGreaterThan(4);
    });
  });
});`,
        `describe('my tests', () => {
  beforeEach(() => { expect.hasAssertions(); });

  it('responds ok', function () {
    client.get('/user', response => {
      expect(response.status).toBe(200);
    });
  });

  it("is a number that is greater than four", () => {
    expect.hasAssertions();

    expect(number).toBeGreaterThan(4);
  });
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream prefer-expect-assertions / valid ${index + 18}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
    })),
    // Upstream 29.16.0 checks source order and reports the first test. Jest
    // collects the whole suite before execution, so this later beforeEach hook
    // sets hasAssertions for both tests. Rslint intentionally accepts both.
    {
      name: 'upstream prefer-expect-assertions / invalid 8',
      code: `describe('some tests', () => {
  it("it1", function() {
    someFunctionToDo();
    someFunctionToDo2();
  });

  beforeEach(() => { expect.hasAssertions(); });

  it("it1", function() {
    someFunctionToDo();
    someFunctionToDo2();
  });
});`,
      languageOptions: {
        ecmaVersion: 2017,
      },
    },
    ...(
      [
        `const expectNumbersToBeGreaterThan = (numbers, value) => {
  for (let number of numbers) {
    expect(number).toBeGreaterThan(value);
  }
};

it('returns numbers that are greater than two', function () {
  expectNumbersToBeGreaterThan(getNumbers(), 2);
});`,
        `it("returns numbers that are greater than five", function () {
  expect.assertions(2);

  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(5);
  }
});`,
        `it("returns things that are less than ten", function () {
  expect.hasAssertions();

  for (const thing in things) {
    expect(thing).toBeLessThan(10);
  }
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream prefer-expect-assertions (loops) / valid ${index}`,
      code,
      options: [
        {
          onlyFunctionsWithExpectInLoop: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
    })),
    ...(
      [
        `const expectNumbersToBeGreaterThan = (numbers, value) => {
  numbers.forEach(number => {
    expect(number).toBeGreaterThan(value);
  });
};

it('returns numbers that are greater than two', function () {
  expectNumbersToBeGreaterThan(getNumbers(), 2);
});`,
        `it('returns numbers that are greater than two', function () {
  expect.assertions(2);
` + helperAssertionBody,
        `beforeEach(() => expect.hasAssertions());

it('returns numbers that are greater than two', function () {` +
          helperAssertionBody,
        `it("returns numbers that are greater than five", function () {
  expect.assertions(2);

  getNumbers().forEach(number => {
    expect(number).toBeGreaterThan(5);
  });
});`,
        `it("returns things that are less than ten", function () {
  expect.hasAssertions();

  things.forEach(thing => {
    expect(thing).toBeLessThan(10);
  });
});`,
        `it('sends the data as a string', () => {
  expect.hasAssertions();

  const stream = openStream();

  stream.on('data', data => {
    expect(data).toBe(expect.any(String));
  });
});`,
        `it('responds ok', function () {
  expect.assertions(1);
` + requestCallbackBody,
        `it.each([1, 2, 3])("returns ok", id => {
  expect.assertions(3);

  client.get(\`/users/\${id}\`, response => {
    expect(response.status).toBe(200);
  });
});

` + singleAssertionTest,
        `it('is a test', () => {
  expect(expected).toBe(actual);
});

describe('my test', () => {
  it('is another test', () => {
    expect(expected).toBe(actual);
  });
});`,
        `it('responds ok', function () {
  expect.assertions(1);
` +
          requestCallbackBody +
          `

describe('my test', () => {
  beforeEach(() => expect.hasAssertions());

  it('responds ok', function () {
    client.get('/user', response => {
      expect(response.status).toBe(200);
    });
  });
});`,
        `it('responds ok', function () {
  expect.assertions(1);
` +
          requestCallbackBody +
          `

describe('my test', () => {
  afterEach(() => expect.hasAssertions());

  it('responds ok', function () {
    client.get('/user', response => {
      expect(response.status).toBe(200);
    });
  });
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream prefer-expect-assertions (callbacks) / valid ${index}`,
      code,
      options: [
        {
          onlyFunctionsWithExpectInCallback: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
    })),
    ...(
      [
        `it('only returns numbers that are greater than zero', async () => {
  expect.hasAssertions();

  for (const number of await getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});`,
        `it('only returns numbers that are greater than zero', async () => {
  expect.assertions(2);

  for (const number of await getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream prefer-expect-assertions (mixed) / valid ${index}`,
      code,
      options: [
        {
          onlyFunctionsWithAsyncKeyword: true,
          onlyFunctionsWithExpectInLoop: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
    })),
    ...(
      [
        'test.each()("is fine", () => { expect.assertions(0); })',
        'test.each``("is fine", () => { expect.assertions(0); })',
        'test.each()("is fine", () => { expect.hasAssertions(); })',
        'test.each``("is fine", () => { expect.hasAssertions(); })',
        'it.each()("is fine", () => { expect.assertions(0); })',
        'it.each``("is fine", () => { expect.assertions(0); })',
        'it.each()("is fine", () => { expect.hasAssertions(); })',
        'it.each``("is fine", () => { expect.hasAssertions(); })',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream .each support / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
    })),
    ...(
      [
        'test.each()("is fine", () => {})',
        'test.each``("is fine", () => {})',
        'it.each()("is fine", () => {})',
        'it.each``("is fine", () => {})',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream .each support / valid ${index + 8}`,
      code,
      options: [
        {
          onlyFunctionsWithAsyncKeyword: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
    })),
    ...(
      [
        `describe.each(['hello'])('%s', () => {
  it('is fine', () => {
    expect.assertions(0);
  });
});`,
        `describe.each\`\`('%s', () => {
  it('is fine', () => {
    expect.assertions(0);
  });
});`,
        `describe.each(['hello'])('%s', () => {
  it('is fine', () => {
    expect.hasAssertions();
  });
});`,
        `describe.each\`\`('%s', () => {
  it('is fine', () => {
    expect.hasAssertions();
  });
});`,
        `describe.each(['hello'])('%s', () => {
  it.each()('is fine', () => {
    expect.assertions(0);
  });
});`,
        `describe.each\`\`('%s', () => {
  it.each()('is fine', () => {
    expect.assertions(0);
  });
});`,
        `describe.each(['hello'])('%s', () => {
  it.each()('is fine', () => {
    expect.hasAssertions();
  });
});`,
        `describe.each\`\`('%s', () => {
  it.each()('is fine', () => {
    expect.hasAssertions();
  });
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream .each support / valid ${index + 12}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
    })),
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: `beforeEach(() => {
  setTimeout(() => expect.hasAssertions(), 5000);
});

it('only returns numbers that are greater than six', () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(6);
  }
});`,
      errors: [
        {
          messageId: 'haveExpectAssertions',
          message:
            'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
          line: 5,
          column: 1,
          endLine: 9,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              desc: 'Add `expect.hasAssertions()`',
              output: `beforeEach(() => {
  setTimeout(() => expect.hasAssertions(), 5000);
});

it('only returns numbers that are greater than six', () => {expect.hasAssertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(6);
  }
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              desc: 'Add `expect.assertions(<number of assertions>)`',
              output: `beforeEach(() => {
  setTimeout(() => expect.hasAssertions(), 5000);
});

it('only returns numbers that are greater than six', () => {expect.assertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(6);
  }
});`,
            },
          ],
        },
      ],
    },
    ...(
      [
        ['it("it1", () => foo())', 23],
        ["it('resolves', () => expect(staged()).toBe(true));", 50],
        ["it('resolves', async () => expect(await staged()).toBe(true));", 62],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index + 1}`,
      code,
      errors: [
        {
          messageId: 'haveExpectAssertions',
          message:
            'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    ...(
      [
        [
          'it("it1", () => {})',
          1,
          1,
          1,
          20,
          'it("it1", () => {expect.hasAssertions();})',
          'it("it1", () => {expect.assertions();})',
        ],
        [
          'it("it1", () => { foo()})',
          1,
          1,
          1,
          26,
          'it("it1", () => {expect.hasAssertions(); foo()})',
          'it("it1", () => {expect.assertions(); foo()})',
        ],
        [
          `it("it1", function() {
  someFunctionToDo();
  someFunctionToDo2();
});`,
          1,
          1,
          4,
          3,
          `it("it1", function() {expect.hasAssertions();
  someFunctionToDo();
  someFunctionToDo2();
});`,
          `it("it1", function() {expect.assertions();
  someFunctionToDo();
  someFunctionToDo2();
});`,
        ],
        [
          `it("it1", function() {
  someFunctionToDo();
  someFunctionToDo2();
});

` + suiteWithAssertionHook,
          1,
          1,
          4,
          3,
          `it("it1", function() {expect.hasAssertions();
  someFunctionToDo();
  someFunctionToDo2();
});

` + suiteWithAssertionHook,
          `it("it1", function() {expect.assertions();
  someFunctionToDo();
  someFunctionToDo2();
});

` + suiteWithAssertionHook,
        ],
        [
          `it("it1", function() {
  someFunctionToDo();
  someFunctionToDo2();
});

describe('some tests', () => {
  afterEach(() => { expect.hasAssertions(); });
  it("it1", function() {
    someFunctionToDo();
    someFunctionToDo2();
  });
});`,
          1,
          1,
          4,
          3,
          `it("it1", function() {expect.hasAssertions();
  someFunctionToDo();
  someFunctionToDo2();
});

describe('some tests', () => {
  afterEach(() => { expect.hasAssertions(); });
  it("it1", function() {
    someFunctionToDo();
    someFunctionToDo2();
  });
});`,
          `it("it1", function() {expect.assertions();
  someFunctionToDo();
  someFunctionToDo2();
});

describe('some tests', () => {
  afterEach(() => { expect.hasAssertions(); });
  it("it1", function() {
    someFunctionToDo();
    someFunctionToDo2();
  });
});`,
        ],
        [
          suiteWithAssertionHook +
            `

it("it1", function() {
  someFunctionToDo();
  someFunctionToDo2();
});`,
          9,
          1,
          12,
          3,
          suiteWithAssertionHook +
            `

it("it1", function() {expect.hasAssertions();
  someFunctionToDo();
  someFunctionToDo2();
});`,
          suiteWithAssertionHook +
            `

it("it1", function() {expect.assertions();
  someFunctionToDo();
  someFunctionToDo2();
});`,
        ],
        [
          suiteWithAssertionHook +
            `

describe('more tests', () => {
  it("it1", function() {
    someFunctionToDo();
    someFunctionToDo2();
  });
});`,
          10,
          3,
          13,
          5,
          suiteWithAssertionHook +
            `

describe('more tests', () => {
  it("it1", function() {expect.hasAssertions();
    someFunctionToDo();
    someFunctionToDo2();
  });
});`,
          suiteWithAssertionHook +
            `

describe('more tests', () => {
  it("it1", function() {expect.assertions();
    someFunctionToDo();
    someFunctionToDo2();
  });
});`,
        ],
        [
          'it("it1", function() {var a = 2;})',
          1,
          1,
          1,
          35,
          'it("it1", function() {expect.hasAssertions();var a = 2;})',
          'it("it1", function() {expect.assertions();var a = 2;})',
        ],
      ] as const
    ).map(
      (
        [
          code,
          line,
          column,
          endLine,
          endColumn,
          suggestionOutput1,
          suggestionOutput2,
        ],
        index,
      ): TestCase => ({
        name: `run 0 invalid ${index + 4}`,
        code,
        errors: [
          {
            messageId: 'haveExpectAssertions',
            message:
              'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
            line,
            column,
            endLine,
            endColumn,
            suggestions: [
              {
                messageId: 'suggestAddingHasAssertions',
                desc: 'Add `expect.hasAssertions()`',
                output: suggestionOutput1,
              },
              {
                messageId: 'suggestAddingAssertions',
                desc: 'Add `expect.assertions(<number of assertions>)`',
                output: suggestionOutput2,
              },
            ],
          },
        ],
      }),
    ),
    {
      name: 'run 0 invalid 12',
      code: 'it("it1", function() {expect.assertions();})',
      errors: [
        {
          messageId: 'assertionsRequiresOneArgument',
          message:
            '`expect.assertions` expects a single argument of type number',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    ...(
      [
        'it("it1", function() {expect.assertions(1,2);})',
        'it("it1", function() {expect.assertions(1,2,);})',
      ] as const
    ).map((code, index): TestCase => ({
      name: `run 0 invalid ${index + 13}`,
      code,
      errors: [
        {
          messageId: 'assertionsRequiresOneArgument',
          message:
            '`expect.assertions` expects a single argument of type number',
          line: 1,
          column: 43,
          endLine: 1,
          endColumn: 44,
          suggestions: [
            {
              messageId: 'suggestRemovingExtraArguments',
              desc: 'Remove extra arguments',
              output: 'it("it1", function() {expect.assertions(1,);})',
            },
          ],
        },
      ],
    })),
    {
      name: 'run 0 invalid 15',
      code: 'it("it1", function() {expect.assertions("1");})',
      errors: [
        {
          messageId: 'assertionsRequiresNumberArgument',
          message: 'This argument should be a number',
          line: 1,
          column: 41,
          endLine: 1,
          endColumn: 44,
        },
      ],
    },
    ...(
      [
        [
          'beforeEach(() => { expect.hasAssertions("1") })',
          1,
          27,
          'beforeEach(() => { expect.hasAssertions() })',
        ],
        [
          'beforeEach(() => expect.hasAssertions("1"))',
          1,
          25,
          'beforeEach(() => expect.hasAssertions())',
        ],
        [
          'afterEach(() => { expect.hasAssertions("1") })',
          1,
          26,
          'afterEach(() => { expect.hasAssertions() })',
        ],
        [
          'afterEach(() => expect.hasAssertions("1"))',
          1,
          24,
          'afterEach(() => expect.hasAssertions())',
        ],
        [
          'it("it1", function() {expect.hasAssertions("1");})',
          1,
          30,
          'it("it1", function() {expect.hasAssertions();})',
        ],
        [
          'it("it1", function() {expect.hasAssertions("1",);})',
          1,
          30,
          'it("it1", function() {expect.hasAssertions();})',
        ],
        [
          'it("it1", function() {expect.hasAssertions("1", "2");})',
          1,
          30,
          'it("it1", function() {expect.hasAssertions();})',
        ],
        [
          `it("it1", function() {
  expect.hasAssertions(() => {
    someFunctionToDo();
    someFunctionToDo2();
  });
});`,
          2,
          10,
          `it("it1", function() {
  expect.hasAssertions();
});`,
        ],
      ] as const
    ).map(([code, line, column, suggestionOutput], index): TestCase => ({
      name: `run 0 invalid ${index + 16}`,
      code,
      errors: [
        {
          messageId: 'hasAssertionsTakesNoArguments',
          message: '`expect.hasAssertions` expects no arguments',
          line,
          column,
          endLine: line,
          endColumn: column + 13,
          suggestions: [
            {
              messageId: 'suggestRemovingExtraArguments',
              desc: 'Remove extra arguments',
              output: suggestionOutput,
            },
          ],
        },
      ],
    })),
    ...(
      [
        [
          `it("it1", async function() {
  expect(someValue).toBe(true);
});`,
          1,
          3,
          `it("it1", async function() {expect.hasAssertions();
  expect(someValue).toBe(true);
});`,
          `it("it1", async function() {expect.assertions();
  expect(someValue).toBe(true);
});`,
        ],
        [
          `it("returns numbers that are greater than four", async () => {
  for(let thing in things) {
    expect(number).toBeGreaterThan(4);
  }
});`,
          1,
          5,
          `it("returns numbers that are greater than four", async () => {expect.hasAssertions();
  for(let thing in things) {
    expect(number).toBeGreaterThan(4);
  }
});`,
          `it("returns numbers that are greater than four", async () => {expect.assertions();
  for(let thing in things) {
    expect(number).toBeGreaterThan(4);
  }
});`,
        ],
        [
          'it("returns numbers that are greater than four", async () => {' +
            numbersLoopBody,
          1,
          5,
          'it("returns numbers that are greater than four", async () => {expect.hasAssertions();' +
            numbersLoopBody,
          'it("returns numbers that are greater than four", async () => {expect.assertions();' +
            numbersLoopBody,
        ],
        [
          'it("returns numbers that are greater than four", async () => {' +
            numbersLoopBody +
            `
` +
            numbersAboveFiveTest,
          1,
          5,
          'it("returns numbers that are greater than four", async () => {expect.hasAssertions();' +
            numbersLoopBody +
            `
` +
            numbersAboveFiveTest,
          'it("returns numbers that are greater than four", async () => {expect.assertions();' +
            numbersLoopBody +
            `
` +
            numbersAboveFiveTest,
        ],
        [
          `beforeAll(() => { expect.hasAssertions(); });

it("returns numbers that are greater than four", async () => {` +
            numbersLoopBody +
            `
` +
            numbersAboveFiveTest,
          3,
          7,
          `beforeAll(() => { expect.hasAssertions(); });

it("returns numbers that are greater than four", async () => {expect.hasAssertions();` +
            numbersLoopBody +
            `
` +
            numbersAboveFiveTest,
          `beforeAll(() => { expect.hasAssertions(); });

it("returns numbers that are greater than four", async () => {expect.assertions();` +
            numbersLoopBody +
            `
` +
            numbersAboveFiveTest,
        ],
        [
          `afterAll(() => { expect.hasAssertions(); });

it("returns numbers that are greater than four", async () => {` +
            numbersLoopBody +
            `
` +
            numbersAboveFiveTest,
          3,
          7,
          `afterAll(() => { expect.hasAssertions(); });

it("returns numbers that are greater than four", async () => {expect.hasAssertions();` +
            numbersLoopBody +
            `
` +
            numbersAboveFiveTest,
          `afterAll(() => { expect.hasAssertions(); });

it("returns numbers that are greater than four", async () => {expect.assertions();` +
            numbersLoopBody +
            `
` +
            numbersAboveFiveTest,
        ],
      ] as const
    ).map(
      (
        [code, line, endLine, suggestionOutput1, suggestionOutput2],
        index,
      ): TestCase => ({
        name: `run 0 invalid ${index + 24}`,
        code,
        options: [
          {
            onlyFunctionsWithAsyncKeyword: true,
          },
        ],
        errors: [
          {
            messageId: 'haveExpectAssertions',
            message:
              'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
            line,
            column: 1,
            endLine,
            endColumn: 3,
            suggestions: [
              {
                messageId: 'suggestAddingHasAssertions',
                desc: 'Add `expect.hasAssertions()`',
                output: suggestionOutput1,
              },
              {
                messageId: 'suggestAddingAssertions',
                desc: 'Add `expect.assertions(<number of assertions>)`',
                output: suggestionOutput2,
              },
            ],
          },
        ],
      }),
    ),
    ...(
      [
        [
          `it('only returns numbers that are greater than six', () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(6);
  }
});`,
          1,
          1,
          5,
          `it('only returns numbers that are greater than six', () => {expect.hasAssertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(6);
  }
});`,
          `it('only returns numbers that are greater than six', () => {expect.assertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(6);
  }
});`,
        ],
        [
          "it('returns numbers that are greater than two', function () {" +
            helperAssertionBody,
          1,
          1,
          9,
          "it('returns numbers that are greater than two', function () {expect.hasAssertions();" +
            helperAssertionBody,
          "it('returns numbers that are greater than two', function () {expect.assertions();" +
            helperAssertionBody,
        ],
        [
          `it("only returns numbers that are greater than seven", function () {
  const numbers = getNumbers();

  for (let i = 0; i < numbers.length; i++) {
    expect(numbers[i]).toBeGreaterThan(7);
  }
});`,
          1,
          1,
          7,
          `it("only returns numbers that are greater than seven", function () {expect.hasAssertions();
  const numbers = getNumbers();

  for (let i = 0; i < numbers.length; i++) {
    expect(numbers[i]).toBeGreaterThan(7);
  }
});`,
          `it("only returns numbers that are greater than seven", function () {expect.assertions();
  const numbers = getNumbers();

  for (let i = 0; i < numbers.length; i++) {
    expect(numbers[i]).toBeGreaterThan(7);
  }
});`,
        ],
        [
          `it('has the number two', () => {
  expect(number).toBe(2);
});

it('only returns numbers that are less than twenty', () => {
  for (const number of getNumbers()) {
    expect(number).toBeLessThan(20);
  }
});`,
          5,
          1,
          9,
          `it('has the number two', () => {
  expect(number).toBe(2);
});

it('only returns numbers that are less than twenty', () => {expect.hasAssertions();
  for (const number of getNumbers()) {
    expect(number).toBeLessThan(20);
  }
});`,
          `it('has the number two', () => {
  expect(number).toBe(2);
});

it('only returns numbers that are less than twenty', () => {expect.assertions();
  for (const number of getNumbers()) {
    expect(number).toBeLessThan(20);
  }
});`,
        ],
        [
          `it("is wrong");

it("is a test", () => {` + numbersLoopBody,
          3,
          1,
          7,
          `it("is wrong");

it("is a test", () => {expect.hasAssertions();` + numbersLoopBody,
          `it("is wrong");

it("is a test", () => {expect.assertions();` + numbersLoopBody,
        ],
        [
          singleAssertionTest +
            `

it("returns numbers that are greater than four", () => {` +
            numbersLoopBody +
            `
` +
            numberAboveFiveTest,
          5,
          1,
          9,
          singleAssertionTest +
            `

it("returns numbers that are greater than four", () => {expect.hasAssertions();` +
            numbersLoopBody +
            `
` +
            numberAboveFiveTest,
          singleAssertionTest +
            `

it("returns numbers that are greater than four", () => {expect.assertions();` +
            numbersLoopBody +
            `
` +
            numberAboveFiveTest,
        ],
        [
          `describe('my tests', () => {
  beforeEach(expect.hasAssertions);
  it("is a number that is greater than four", () => {
    expect(number).toBeGreaterThan(4);
  });
});

describe('more tests', () => {
  it("returns numbers that are greater than four", () => {
    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(4);
    }
  });
});
` + numberAboveFiveTest,
          9,
          3,
          13,
          `describe('my tests', () => {
  beforeEach(expect.hasAssertions);
  it("is a number that is greater than four", () => {
    expect(number).toBeGreaterThan(4);
  });
});

describe('more tests', () => {
  it("returns numbers that are greater than four", () => {expect.hasAssertions();
    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(4);
    }
  });
});
` + numberAboveFiveTest,
          `describe('my tests', () => {
  beforeEach(expect.hasAssertions);
  it("is a number that is greater than four", () => {
    expect(number).toBeGreaterThan(4);
  });
});

describe('more tests', () => {
  it("returns numbers that are greater than four", () => {expect.assertions();
    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(4);
    }
  });
});
` + numberAboveFiveTest,
        ],
        [
          'it.each([1, 2, 3])("returns numbers that are greater than four", () => {' +
            numbersLoopBody +
            `

` +
            singleAssertionTest,
          1,
          1,
          5,
          'it.each([1, 2, 3])("returns numbers that are greater than four", () => {expect.hasAssertions();' +
            numbersLoopBody +
            `

` +
            singleAssertionTest,
          'it.each([1, 2, 3])("returns numbers that are greater than four", () => {expect.assertions();' +
            numbersLoopBody +
            `

` +
            singleAssertionTest,
        ],
        [
          'it("returns numbers that are greater than four", () => {' +
            numbersLoopBody +
            `

` +
            singleAssertionTest,
          1,
          1,
          5,
          'it("returns numbers that are greater than four", () => {expect.hasAssertions();' +
            numbersLoopBody +
            `

` +
            singleAssertionTest,
          'it("returns numbers that are greater than four", () => {expect.assertions();' +
            numbersLoopBody +
            `

` +
            singleAssertionTest,
        ],
        [
          'it("returns numbers that are greater than four", () => {' +
            numbersLoopBody +
            `

it("is a number that is greater than four", () => {
  expect.hasAssertions();

  expect(number).toBeGreaterThan(4);
});`,
          1,
          1,
          5,
          'it("returns numbers that are greater than four", () => {expect.hasAssertions();' +
            numbersLoopBody +
            `

it("is a number that is greater than four", () => {
  expect.hasAssertions();

  expect(number).toBeGreaterThan(4);
});`,
          'it("returns numbers that are greater than four", () => {expect.assertions();' +
            numbersLoopBody +
            `

it("is a number that is greater than four", () => {
  expect.hasAssertions();

  expect(number).toBeGreaterThan(4);
});`,
        ],
        [
          `it("it1", () => {
  expect.hasAssertions();

  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});

it("it1", () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});`,
          9,
          1,
          13,
          `it("it1", () => {
  expect.hasAssertions();

  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});

it("it1", () => {expect.hasAssertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});`,
          `it("it1", () => {
  expect.hasAssertions();

  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});

it("it1", () => {expect.assertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});`,
        ],
      ] as const
    ).map(
      (
        [code, line, column, endLine, suggestionOutput1, suggestionOutput2],
        index,
      ): TestCase => ({
        name: `run 0 invalid ${index + 30}`,
        code,
        options: [
          {
            onlyFunctionsWithExpectInLoop: true,
          },
        ],
        errors: [
          {
            messageId: 'haveExpectAssertions',
            message:
              'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
            line,
            column,
            endLine,
            endColumn: column + 2,
            suggestions: [
              {
                messageId: 'suggestAddingHasAssertions',
                desc: 'Add `expect.hasAssertions()`',
                output: suggestionOutput1,
              },
              {
                messageId: 'suggestAddingAssertions',
                desc: 'Add `expect.assertions(<number of assertions>)`',
                output: suggestionOutput2,
              },
            ],
          },
        ],
      }),
    ),
    {
      name: 'run 0 invalid 41',
      code:
        `it("returns numbers that are greater than four", async () => {
  for (const number of await getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});
` + numbersAboveFiveTest,
      options: [
        {
          onlyFunctionsWithExpectInLoop: true,
        },
      ],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          message:
            'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
          line: 1,
          column: 1,
          endLine: 5,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              desc: 'Add `expect.hasAssertions()`',
              output:
                `it("returns numbers that are greater than four", async () => {expect.hasAssertions();
  for (const number of await getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});
` + numbersAboveFiveTest,
            },
            {
              messageId: 'suggestAddingAssertions',
              desc: 'Add `expect.assertions(<number of assertions>)`',
              output:
                `it("returns numbers that are greater than four", async () => {expect.assertions();
  for (const number of await getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});
` + numbersAboveFiveTest,
            },
          ],
        },
        {
          messageId: 'haveExpectAssertions',
          message:
            'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
          line: 7,
          column: 1,
          endLine: 11,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              desc: 'Add `expect.hasAssertions()`',
              output: `it("returns numbers that are greater than four", async () => {
  for (const number of await getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});

it("returns numbers that are greater than five", () => {expect.hasAssertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(5);
  }
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              desc: 'Add `expect.assertions(<number of assertions>)`',
              output: `it("returns numbers that are greater than four", async () => {
  for (const number of await getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});

it("returns numbers that are greater than five", () => {expect.assertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(5);
  }
});`,
            },
          ],
        },
      ],
    },
    ...(
      [
        [
          `it("it1", async () => {
  expect.hasAssertions();
` +
            numbersLoopBody +
            `

it("it1", () => {` +
            numbersLoopBody,
          9,
          1,
          13,
          `it("it1", async () => {
  expect.hasAssertions();
` +
            numbersLoopBody +
            `

it("it1", () => {expect.hasAssertions();` +
            numbersLoopBody,
          `it("it1", async () => {
  expect.hasAssertions();
` +
            numbersLoopBody +
            `

it("it1", () => {expect.assertions();` +
            numbersLoopBody,
        ],
        [
          `describe('my tests', () => {
  beforeEach(() => { expect.hasAssertions(); });

  it("it1", async () => {
    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(4);
    }
  });
});

it("it1", () => {` + numbersLoopBody,
          11,
          1,
          15,
          `describe('my tests', () => {
  beforeEach(() => { expect.hasAssertions(); });

  it("it1", async () => {
    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(4);
    }
  });
});

it("it1", () => {expect.hasAssertions();` + numbersLoopBody,
          `describe('my tests', () => {
  beforeEach(() => { expect.hasAssertions(); });

  it("it1", async () => {
    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(4);
    }
  });
});

it("it1", () => {expect.assertions();` + numbersLoopBody,
        ],
        [
          `describe('my tests', () => {
  afterEach(() => { expect.hasAssertions(); });

  it("it1", async () => {
    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(4);
    }
  });
});

it("it1", () => {` + numbersLoopBody,
          11,
          1,
          15,
          `describe('my tests', () => {
  afterEach(() => { expect.hasAssertions(); });

  it("it1", async () => {
    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(4);
    }
  });
});

it("it1", () => {expect.hasAssertions();` + numbersLoopBody,
          `describe('my tests', () => {
  afterEach(() => { expect.hasAssertions(); });

  it("it1", async () => {
    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(4);
    }
  });
});

it("it1", () => {expect.assertions();` + numbersLoopBody,
        ],
        [
          `it.skip.each\`\`("it1", async () => {
  expect.hasAssertions();
` +
            numbersLoopBody +
            `

it("it1", () => {` +
            numbersLoopBody,
          9,
          1,
          13,
          `it.skip.each\`\`("it1", async () => {
  expect.hasAssertions();
` +
            numbersLoopBody +
            `

it("it1", () => {expect.hasAssertions();` +
            numbersLoopBody,
          `it.skip.each\`\`("it1", async () => {
  expect.hasAssertions();
` +
            numbersLoopBody +
            `

it("it1", () => {expect.assertions();` +
            numbersLoopBody,
        ],
        [
          'it("it1", async () => {' +
            numbersLoopBody +
            `

it("it1", () => {
  expect.hasAssertions();
` +
            numbersLoopBody,
          1,
          1,
          5,
          'it("it1", async () => {expect.hasAssertions();' +
            numbersLoopBody +
            `

it("it1", () => {
  expect.hasAssertions();
` +
            numbersLoopBody,
          'it("it1", async () => {expect.assertions();' +
            numbersLoopBody +
            `

it("it1", () => {
  expect.hasAssertions();
` +
            numbersLoopBody,
        ],
        [
          `describe('my tests', () => {
  it("it1", async () => {
    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(4);
    }
  });
});

it("it1", () => {
  expect.hasAssertions();
` + numbersLoopBody,
          2,
          3,
          6,
          `describe('my tests', () => {
  it("it1", async () => {expect.hasAssertions();
    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(4);
    }
  });
});

it("it1", () => {
  expect.hasAssertions();
` + numbersLoopBody,
          `describe('my tests', () => {
  it("it1", async () => {expect.assertions();
    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(4);
    }
  });
});

it("it1", () => {
  expect.hasAssertions();
` + numbersLoopBody,
        ],
      ] as const
    ).map(
      (
        [code, line, column, endLine, suggestionOutput1, suggestionOutput2],
        index,
      ): TestCase => ({
        name: `run 0 invalid ${index + 42}`,
        code,
        options: [
          {
            onlyFunctionsWithExpectInLoop: true,
          },
        ],
        errors: [
          {
            messageId: 'haveExpectAssertions',
            message:
              'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
            line,
            column,
            endLine,
            endColumn: column + 2,
            suggestions: [
              {
                messageId: 'suggestAddingHasAssertions',
                desc: 'Add `expect.hasAssertions()`',
                output: suggestionOutput1,
              },
              {
                messageId: 'suggestAddingAssertions',
                desc: 'Add `expect.assertions(<number of assertions>)`',
                output: suggestionOutput2,
              },
            ],
          },
        ],
      }),
    ),
    ...(
      [
        [
          `it('sends the data as a string', () => {
  const stream = openStream();

  stream.on('data', data => {
    expect(data).toBe(expect.any(String));
  });
});`,
          1,
          7,
          `it('sends the data as a string', () => {expect.hasAssertions();
  const stream = openStream();

  stream.on('data', data => {
    expect(data).toBe(expect.any(String));
  });
});`,
          `it('sends the data as a string', () => {expect.assertions();
  const stream = openStream();

  stream.on('data', data => {
    expect(data).toBe(expect.any(String));
  });
});`,
        ],
        [
          "it('responds ok', function () {" + requestCallbackBody,
          1,
          5,
          "it('responds ok', function () {expect.hasAssertions();" +
            requestCallbackBody,
          "it('responds ok', function () {expect.assertions();" +
            requestCallbackBody,
        ],
        [
          `it('responds ok', function () {
  client.get('/user', response => {
    expect.assertions(1);

    expect(response.status).toBe(200);
  });
});`,
          1,
          7,
          `it('responds ok', function () {expect.hasAssertions();
  client.get('/user', response => {
    expect.assertions(1);

    expect(response.status).toBe(200);
  });
});`,
          `it('responds ok', function () {expect.assertions();
  client.get('/user', response => {
    expect.assertions(1);

    expect(response.status).toBe(200);
  });
});`,
        ],
        [
          `it('responds ok', function () {
` + requestCallbackWithAssertionCount,
          1,
          9,
          `it('responds ok', function () {expect.hasAssertions();
` + requestCallbackWithAssertionCount,
          `it('responds ok', function () {expect.assertions();
` + requestCallbackWithAssertionCount,
        ],
        [
          `it('returns numbers that are greater than two', function () {
  const expectNumberToBeGreaterThan = (number, value) => {
    expect(number).toBeGreaterThan(value);
  };

  expectNumberToBeGreaterThan(1, 2);
});`,
          1,
          7,
          `it('returns numbers that are greater than two', function () {expect.hasAssertions();
  const expectNumberToBeGreaterThan = (number, value) => {
    expect(number).toBeGreaterThan(value);
  };

  expectNumberToBeGreaterThan(1, 2);
});`,
          `it('returns numbers that are greater than two', function () {expect.assertions();
  const expectNumberToBeGreaterThan = (number, value) => {
    expect(number).toBeGreaterThan(value);
  };

  expectNumberToBeGreaterThan(1, 2);
});`,
        ],
        [
          "it('returns numbers that are greater than two', function () {" +
            helperAssertionBody,
          1,
          9,
          "it('returns numbers that are greater than two', function () {expect.hasAssertions();" +
            helperAssertionBody,
          "it('returns numbers that are greater than two', function () {expect.assertions();" +
            helperAssertionBody,
        ],
        [
          `it('only returns numbers that are greater than six', () => {
  getNumbers().forEach(number => {
    expect(number).toBeGreaterThan(6);
  });
});`,
          1,
          5,
          `it('only returns numbers that are greater than six', () => {expect.hasAssertions();
  getNumbers().forEach(number => {
    expect(number).toBeGreaterThan(6);
  });
});`,
          `it('only returns numbers that are greater than six', () => {expect.assertions();
  getNumbers().forEach(number => {
    expect(number).toBeGreaterThan(6);
  });
});`,
        ],
        [
          `it("is wrong");

it('responds ok', function () {
` + requestCallbackWithAssertionCount,
          3,
          11,
          `it("is wrong");

it('responds ok', function () {expect.hasAssertions();
` + requestCallbackWithAssertionCount,
          `it("is wrong");

it('responds ok', function () {expect.assertions();
` + requestCallbackWithAssertionCount,
        ],
        [
          singleAssertionTest +
            `

it('responds ok', function () {
  const expectOkResponse = response => {
    expect(response.status).toBe(200);
  };

  client.get('/user', expectOkResponse);
});
` +
            numberAboveFiveTest,
          5,
          11,
          singleAssertionTest +
            `

it('responds ok', function () {expect.hasAssertions();
  const expectOkResponse = response => {
    expect(response.status).toBe(200);
  };

  client.get('/user', expectOkResponse);
});
` +
            numberAboveFiveTest,
          singleAssertionTest +
            `

it('responds ok', function () {expect.assertions();
  const expectOkResponse = response => {
    expect(response.status).toBe(200);
  };

  client.get('/user', expectOkResponse);
});
` +
            numberAboveFiveTest,
        ],
        [
          singleAssertionTest +
            `

it("returns numbers that are greater than four", () => {
  getNumbers().map(number => {
    expect(number).toBeGreaterThan(0);
  });
});
` +
            numberAboveFiveTest,
          5,
          9,
          singleAssertionTest +
            `

it("returns numbers that are greater than four", () => {expect.hasAssertions();
  getNumbers().map(number => {
    expect(number).toBeGreaterThan(0);
  });
});
` +
            numberAboveFiveTest,
          singleAssertionTest +
            `

it("returns numbers that are greater than four", () => {expect.assertions();
  getNumbers().map(number => {
    expect(number).toBeGreaterThan(0);
  });
});
` +
            numberAboveFiveTest,
        ],
        [
          `it.each([1, 2, 3])("returns ok", id => {
  client.get(\`/users/\${id}\`, response => {
    expect(response.status).toBe(200);
  });
});

` + singleAssertionTest,
          1,
          5,
          `it.each([1, 2, 3])("returns ok", id => {expect.hasAssertions();
  client.get(\`/users/\${id}\`, response => {
    expect(response.status).toBe(200);
  });
});

` + singleAssertionTest,
          `it.each([1, 2, 3])("returns ok", id => {expect.assertions();
  client.get(\`/users/\${id}\`, response => {
    expect(response.status).toBe(200);
  });
});

` + singleAssertionTest,
        ],
        [
          "it('responds ok', function () {" +
            requestCallbackBody +
            `

` +
            singleAssertionTest,
          1,
          5,
          "it('responds ok', function () {expect.hasAssertions();" +
            requestCallbackBody +
            `

` +
            singleAssertionTest,
          "it('responds ok', function () {expect.assertions();" +
            requestCallbackBody +
            `

` +
            singleAssertionTest,
        ],
        [
          "it('responds ok', function () {" +
            requestCallbackBody +
            `

it("is a number that is greater than four", () => {
  expect.hasAssertions();

  expect(number).toBeGreaterThan(4);
});`,
          1,
          5,
          "it('responds ok', function () {expect.hasAssertions();" +
            requestCallbackBody +
            `

it("is a number that is greater than four", () => {
  expect.hasAssertions();

  expect(number).toBeGreaterThan(4);
});`,
          "it('responds ok', function () {expect.assertions();" +
            requestCallbackBody +
            `

it("is a number that is greater than four", () => {
  expect.hasAssertions();

  expect(number).toBeGreaterThan(4);
});`,
        ],
        [
          `it("it1", () => {
  expect.hasAssertions();

  getNumbers().forEach(number => {
    expect(number).toBeGreaterThan(0);
  });
});

it("it1", () => {
  getNumbers().forEach(number => {
    expect(number).toBeGreaterThan(0);
  });
});`,
          9,
          13,
          `it("it1", () => {
  expect.hasAssertions();

  getNumbers().forEach(number => {
    expect(number).toBeGreaterThan(0);
  });
});

it("it1", () => {expect.hasAssertions();
  getNumbers().forEach(number => {
    expect(number).toBeGreaterThan(0);
  });
});`,
          `it("it1", () => {
  expect.hasAssertions();

  getNumbers().forEach(number => {
    expect(number).toBeGreaterThan(0);
  });
});

it("it1", () => {expect.assertions();
  getNumbers().forEach(number => {
    expect(number).toBeGreaterThan(0);
  });
});`,
        ],
        [
          `it('responds ok', function () {
  expect.hasAssertions();
` +
            requestCallbackBody +
            `

it('responds not found', function () {
  client.get('/user', response => {
    expect(response.status).toBe(404);
  });
});`,
          9,
          13,
          `it('responds ok', function () {
  expect.hasAssertions();
` +
            requestCallbackBody +
            `

it('responds not found', function () {expect.hasAssertions();
  client.get('/user', response => {
    expect(response.status).toBe(404);
  });
});`,
          `it('responds ok', function () {
  expect.hasAssertions();
` +
            requestCallbackBody +
            `

it('responds not found', function () {expect.assertions();
  client.get('/user', response => {
    expect(response.status).toBe(404);
  });
});`,
        ],
        [
          `it.skip.each\`\`("it1", async () => {
  expect.hasAssertions();
` +
            requestCallbackBody +
            `

it("responds ok", () => {` +
            requestCallbackBody,
          9,
          13,
          `it.skip.each\`\`("it1", async () => {
  expect.hasAssertions();
` +
            requestCallbackBody +
            `

it("responds ok", () => {expect.hasAssertions();` +
            requestCallbackBody,
          `it.skip.each\`\`("it1", async () => {
  expect.hasAssertions();
` +
            requestCallbackBody +
            `

it("responds ok", () => {expect.assertions();` +
            requestCallbackBody,
        ],
      ] as const
    ).map(
      (
        [code, line, endLine, suggestionOutput1, suggestionOutput2],
        index,
      ): TestCase => ({
        name: `run 0 invalid ${index + 48}`,
        code,
        options: [
          {
            onlyFunctionsWithExpectInCallback: true,
          },
        ],
        errors: [
          {
            messageId: 'haveExpectAssertions',
            message:
              'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
            line,
            column: 1,
            endLine,
            endColumn: 3,
            suggestions: [
              {
                messageId: 'suggestAddingHasAssertions',
                desc: 'Add `expect.hasAssertions()`',
                output: suggestionOutput1,
              },
              {
                messageId: 'suggestAddingAssertions',
                desc: 'Add `expect.assertions(<number of assertions>)`',
                output: suggestionOutput2,
              },
            ],
          },
        ],
      }),
    ),
    {
      name: 'run 0 invalid 64',
      code: `it("returns numbers that are greater than four", function(expect) {
  expect.assertions(2);

  for(let thing in things) {
    expect(number).toBeGreaterThan(4);
  }
});`,
      errors: [
        {
          messageId: 'haveExpectAssertions',
          message:
            'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
          line: 1,
          column: 1,
          endLine: 7,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              desc: 'Add `expect.hasAssertions()`',
              output: `it("returns numbers that are greater than four", function(expect) {expect.hasAssertions();
  expect.assertions(2);

  for(let thing in things) {
    expect(number).toBeGreaterThan(4);
  }
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              desc: 'Add `expect.assertions(<number of assertions>)`',
              output: `it("returns numbers that are greater than four", function(expect) {expect.assertions();
  expect.assertions(2);

  for(let thing in things) {
    expect(number).toBeGreaterThan(4);
  }
});`,
            },
          ],
        },
      ],
    },
    ...(
      [
        [
          `it('only returns numbers that are greater than zero', () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});

it("is zero", () => {
  expect.hasAssertions();

  expect(0).toBe(0);
});`,
          1,
          5,
          `it('only returns numbers that are greater than zero', () => {expect.hasAssertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});

it("is zero", () => {
  expect.hasAssertions();

  expect(0).toBe(0);
});`,
          `it('only returns numbers that are greater than zero', () => {expect.assertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});

it("is zero", () => {
  expect.hasAssertions();

  expect(0).toBe(0);
});`,
        ],
        [
          `it('only returns numbers that are greater than zero', () => {
  expect.hasAssertions();

  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});

it('only returns numbers that are less than 100', () => {
  for (const number of getNumbers()) {
    expect(number).toBeLessThan(0);
  }
});`,
          9,
          13,
          `it('only returns numbers that are greater than zero', () => {
  expect.hasAssertions();

  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});

it('only returns numbers that are less than 100', () => {expect.hasAssertions();
  for (const number of getNumbers()) {
    expect(number).toBeLessThan(0);
  }
});`,
          `it('only returns numbers that are greater than zero', () => {
  expect.hasAssertions();

  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});

it('only returns numbers that are less than 100', () => {expect.assertions();
  for (const number of getNumbers()) {
    expect(number).toBeLessThan(0);
  }
});`,
        ],
      ] as const
    ).map(
      (
        [code, line, endLine, suggestionOutput1, suggestionOutput2],
        index,
      ): TestCase => ({
        name: `run 0 invalid ${index + 65}`,
        code,
        options: [
          {
            onlyFunctionsWithExpectInLoop: true,
          },
        ],
        errors: [
          {
            messageId: 'haveExpectAssertions',
            message:
              'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
            line,
            column: 1,
            endLine,
            endColumn: 3,
            suggestions: [
              {
                messageId: 'suggestAddingHasAssertions',
                desc: 'Add `expect.hasAssertions()`',
                output: suggestionOutput1,
              },
              {
                messageId: 'suggestAddingAssertions',
                desc: 'Add `expect.assertions(<number of assertions>)`',
                output: suggestionOutput2,
              },
            ],
          },
        ],
      }),
    ),
    ...(
      [
        [
          `it("to be true", async function() {
  expect(someValue).toBe(true);
});`,
          3,
          `it("to be true", async function() {expect.hasAssertions();
  expect(someValue).toBe(true);
});`,
          `it("to be true", async function() {expect.assertions();
  expect(someValue).toBe(true);
});`,
        ],
        [
          `it('only returns numbers that are greater than zero', async () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});`,
          5,
          `it('only returns numbers that are greater than zero', async () => {expect.hasAssertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});`,
          `it('only returns numbers that are greater than zero', async () => {expect.assertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});`,
        ],
      ] as const
    ).map(
      (
        [code, endLine, suggestionOutput1, suggestionOutput2],
        index,
      ): TestCase => ({
        name: `run 0 invalid ${index + 67}`,
        code,
        options: [
          {
            onlyFunctionsWithAsyncKeyword: true,
            onlyFunctionsWithExpectInLoop: true,
          },
        ],
        errors: [
          {
            messageId: 'haveExpectAssertions',
            message:
              'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
            line: 1,
            column: 1,
            endLine,
            endColumn: 3,
            suggestions: [
              {
                messageId: 'suggestAddingHasAssertions',
                desc: 'Add `expect.hasAssertions()`',
                output: suggestionOutput1,
              },
              {
                messageId: 'suggestAddingAssertions',
                desc: 'Add `expect.assertions(<number of assertions>)`',
                output: suggestionOutput2,
              },
            ],
          },
        ],
      }),
    ),
    ...(
      [
        [
          `test.each()("is not fine", () => {
  expect(someValue).toBe(true);
});`,
          1,
          1,
          3,
          `test.each()("is not fine", () => {expect.hasAssertions();
  expect(someValue).toBe(true);
});`,
          `test.each()("is not fine", () => {expect.assertions();
  expect(someValue).toBe(true);
});`,
        ],
        [
          `describe.each()('something', () => {
  it("is not fine", () => {
    expect(someValue).toBe(true);
  });
});`,
          2,
          3,
          4,
          `describe.each()('something', () => {
  it("is not fine", () => {expect.hasAssertions();
    expect(someValue).toBe(true);
  });
});`,
          `describe.each()('something', () => {
  it("is not fine", () => {expect.assertions();
    expect(someValue).toBe(true);
  });
});`,
        ],
        [
          `describe.each()('something', () => {
  test.each()("is not fine", () => {
    expect(someValue).toBe(true);
  });
});`,
          2,
          3,
          4,
          `describe.each()('something', () => {
  test.each()("is not fine", () => {expect.hasAssertions();
    expect(someValue).toBe(true);
  });
});`,
          `describe.each()('something', () => {
  test.each()("is not fine", () => {expect.assertions();
    expect(someValue).toBe(true);
  });
});`,
        ],
      ] as const
    ).map(
      (
        [code, line, column, endLine, suggestionOutput1, suggestionOutput2],
        index,
      ): TestCase => ({
        name: `run 0 invalid ${index + 69}`,
        code,
        errors: [
          {
            messageId: 'haveExpectAssertions',
            message:
              'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
            line,
            column,
            endLine,
            endColumn: column + 2,
            suggestions: [
              {
                messageId: 'suggestAddingHasAssertions',
                desc: 'Add `expect.hasAssertions()`',
                output: suggestionOutput1,
              },
              {
                messageId: 'suggestAddingAssertions',
                desc: 'Add `expect.assertions(<number of assertions>)`',
                output: suggestionOutput2,
              },
            ],
          },
        ],
      }),
    ),
    ...(
      [
        [
          `test.each()("is not fine", async () => {
  expect(someValue).toBe(true);
});`,
          1,
          1,
          3,
          `test.each()("is not fine", async () => {expect.hasAssertions();
  expect(someValue).toBe(true);
});`,
          `test.each()("is not fine", async () => {expect.assertions();
  expect(someValue).toBe(true);
});`,
        ],
        [
          `it.each()("is not fine", async () => {
  expect(someValue).toBe(true);
});`,
          1,
          1,
          3,
          `it.each()("is not fine", async () => {expect.hasAssertions();
  expect(someValue).toBe(true);
});`,
          `it.each()("is not fine", async () => {expect.assertions();
  expect(someValue).toBe(true);
});`,
        ],
        [
          `describe.each()('something', () => {
  test.each()("is not fine", async () => {
    expect(someValue).toBe(true);
  });
});`,
          2,
          3,
          4,
          `describe.each()('something', () => {
  test.each()("is not fine", async () => {expect.hasAssertions();
    expect(someValue).toBe(true);
  });
});`,
          `describe.each()('something', () => {
  test.each()("is not fine", async () => {expect.assertions();
    expect(someValue).toBe(true);
  });
});`,
        ],
      ] as const
    ).map(
      (
        [code, line, column, endLine, suggestionOutput1, suggestionOutput2],
        index,
      ): TestCase => ({
        name: `run 0 invalid ${index + 72}`,
        code,
        options: [
          {
            onlyFunctionsWithAsyncKeyword: true,
          },
        ],
        errors: [
          {
            messageId: 'haveExpectAssertions',
            message:
              'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
            line,
            column,
            endLine,
            endColumn: column + 2,
            suggestions: [
              {
                messageId: 'suggestAddingHasAssertions',
                desc: 'Add `expect.hasAssertions()`',
                output: suggestionOutput1,
              },
              {
                messageId: 'suggestAddingAssertions',
                desc: 'Add `expect.assertions(<number of assertions>)`',
                output: suggestionOutput2,
              },
            ],
          },
        ],
      }),
    ),
    // Upstream 29.16.0 counts the nested hasAssertions call as hook coverage.
    // The timer may run after the test has finished, leaving an empty loop with
    // no assertion guard. Rslint therefore reports the missing guard.
    {
      name: 'upstream prefer-expect-assertions / valid 27',
      code: `beforeEach(() => {
  setTimeout(() => expect.hasAssertions(), 5000);
});

it('only returns numbers that are greater than six', () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(6);
  }
});`,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'haveExpectAssertions',
          message:
            'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
          line: 5,
          column: 1,
          endLine: 9,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              desc: 'Add `expect.hasAssertions()`',
              output: `beforeEach(() => {
  setTimeout(() => expect.hasAssertions(), 5000);
});

it('only returns numbers that are greater than six', () => {expect.hasAssertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(6);
  }
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              desc: 'Add `expect.assertions(<number of assertions>)`',
              output: `beforeEach(() => {
  setTimeout(() => expect.hasAssertions(), 5000);
});

it('only returns numbers that are greater than six', () => {expect.assertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(6);
  }
});`,
            },
          ],
        },
      ],
    },
    ...(
      [
        ['it("it1", () => foo())', 23],
        ["it('resolves', () => expect(staged()).toBe(true));", 50],
        ["it('resolves', async () => expect(await staged()).toBe(true));", 62],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `upstream prefer-expect-assertions / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'haveExpectAssertions',
          message:
            'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    ...(
      [
        [
          'it("it1", () => {})',
          'upstream prefer-expect-assertions / invalid 3',
          1,
          1,
          1,
          20,
          'it("it1", () => {expect.hasAssertions();})',
          'it("it1", () => {expect.assertions();})',
        ],
        [
          'it("it1", () => { foo()})',
          'upstream prefer-expect-assertions / invalid 4',
          1,
          1,
          1,
          26,
          'it("it1", () => {expect.hasAssertions(); foo()})',
          'it("it1", () => {expect.assertions(); foo()})',
        ],
        [
          `it("it1", function() {
  someFunctionToDo();
  someFunctionToDo2();
});`,
          'upstream prefer-expect-assertions / invalid 5',
          1,
          1,
          4,
          3,
          `it("it1", function() {expect.hasAssertions();
  someFunctionToDo();
  someFunctionToDo2();
});`,
          `it("it1", function() {expect.assertions();
  someFunctionToDo();
  someFunctionToDo2();
});`,
        ],
        [
          `it("it1", function() {
  someFunctionToDo();
  someFunctionToDo2();
});

` + suiteWithAssertionHook,
          'upstream prefer-expect-assertions / invalid 6',
          1,
          1,
          4,
          3,
          `it("it1", function() {expect.hasAssertions();
  someFunctionToDo();
  someFunctionToDo2();
});

` + suiteWithAssertionHook,
          `it("it1", function() {expect.assertions();
  someFunctionToDo();
  someFunctionToDo2();
});

` + suiteWithAssertionHook,
        ],
        [
          `it("it1", function() {
  someFunctionToDo();
  someFunctionToDo2();
});

describe('some tests', () => {
  afterEach(() => { expect.hasAssertions(); });
  it("it1", function() {
    someFunctionToDo();
    someFunctionToDo2();
  });
});`,
          'upstream prefer-expect-assertions / invalid 7',
          1,
          1,
          4,
          3,
          `it("it1", function() {expect.hasAssertions();
  someFunctionToDo();
  someFunctionToDo2();
});

describe('some tests', () => {
  afterEach(() => { expect.hasAssertions(); });
  it("it1", function() {
    someFunctionToDo();
    someFunctionToDo2();
  });
});`,
          `it("it1", function() {expect.assertions();
  someFunctionToDo();
  someFunctionToDo2();
});

describe('some tests', () => {
  afterEach(() => { expect.hasAssertions(); });
  it("it1", function() {
    someFunctionToDo();
    someFunctionToDo2();
  });
});`,
        ],
        [
          suiteWithAssertionHook +
            `

it("it1", function() {
  someFunctionToDo();
  someFunctionToDo2();
});`,
          'upstream prefer-expect-assertions / invalid 9',
          9,
          1,
          12,
          3,
          suiteWithAssertionHook +
            `

it("it1", function() {expect.hasAssertions();
  someFunctionToDo();
  someFunctionToDo2();
});`,
          suiteWithAssertionHook +
            `

it("it1", function() {expect.assertions();
  someFunctionToDo();
  someFunctionToDo2();
});`,
        ],
        [
          suiteWithAssertionHook +
            `

describe('more tests', () => {
  it("it1", function() {
    someFunctionToDo();
    someFunctionToDo2();
  });
});`,
          'upstream prefer-expect-assertions / invalid 10',
          10,
          3,
          13,
          5,
          suiteWithAssertionHook +
            `

describe('more tests', () => {
  it("it1", function() {expect.hasAssertions();
    someFunctionToDo();
    someFunctionToDo2();
  });
});`,
          suiteWithAssertionHook +
            `

describe('more tests', () => {
  it("it1", function() {expect.assertions();
    someFunctionToDo();
    someFunctionToDo2();
  });
});`,
        ],
        [
          'it("it1", function() {var a = 2;})',
          'upstream prefer-expect-assertions / invalid 11',
          1,
          1,
          1,
          35,
          'it("it1", function() {expect.hasAssertions();var a = 2;})',
          'it("it1", function() {expect.assertions();var a = 2;})',
        ],
      ] as const
    ).map(
      ([
        code,
        name,
        line,
        column,
        endLine,
        endColumn,
        suggestionOutput1,
        suggestionOutput2,
      ]): TestCase => ({
        name,
        code,
        languageOptions: {
          ecmaVersion: 2017,
        },
        errors: [
          {
            messageId: 'haveExpectAssertions',
            message:
              'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
            line,
            column,
            endLine,
            endColumn,
            suggestions: [
              {
                messageId: 'suggestAddingHasAssertions',
                desc: 'Add `expect.hasAssertions()`',
                output: suggestionOutput1,
              },
              {
                messageId: 'suggestAddingAssertions',
                desc: 'Add `expect.assertions(<number of assertions>)`',
                output: suggestionOutput2,
              },
            ],
          },
        ],
      }),
    ),
    {
      name: 'upstream prefer-expect-assertions / invalid 12',
      code: 'it("it1", function() {expect.assertions();})',
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'assertionsRequiresOneArgument',
          message:
            '`expect.assertions` expects a single argument of type number',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    ...(
      [
        'it("it1", function() {expect.assertions(1,2);})',
        'it("it1", function() {expect.assertions(1,2,);})',
      ] as const
    ).map((code, index): TestCase => ({
      name: `upstream prefer-expect-assertions / invalid ${index + 13}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'assertionsRequiresOneArgument',
          message:
            '`expect.assertions` expects a single argument of type number',
          line: 1,
          column: 43,
          endLine: 1,
          endColumn: 44,
          suggestions: [
            {
              messageId: 'suggestRemovingExtraArguments',
              desc: 'Remove extra arguments',
              output: 'it("it1", function() {expect.assertions(1,);})',
            },
          ],
        },
      ],
    })),
    {
      name: 'upstream prefer-expect-assertions / invalid 15',
      code: 'it("it1", function() {expect.assertions("1");})',
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'assertionsRequiresNumberArgument',
          message: 'This argument should be a number',
          line: 1,
          column: 41,
          endLine: 1,
          endColumn: 44,
        },
      ],
    },
    ...(
      [
        [
          'beforeEach(() => { expect.hasAssertions("1") })',
          1,
          27,
          'beforeEach(() => { expect.hasAssertions() })',
        ],
        [
          'beforeEach(() => expect.hasAssertions("1"))',
          1,
          25,
          'beforeEach(() => expect.hasAssertions())',
        ],
        [
          'afterEach(() => { expect.hasAssertions("1") })',
          1,
          26,
          'afterEach(() => { expect.hasAssertions() })',
        ],
        [
          'afterEach(() => expect.hasAssertions("1"))',
          1,
          24,
          'afterEach(() => expect.hasAssertions())',
        ],
        [
          'it("it1", function() {expect.hasAssertions("1");})',
          1,
          30,
          'it("it1", function() {expect.hasAssertions();})',
        ],
        [
          'it("it1", function() {expect.hasAssertions("1",);})',
          1,
          30,
          'it("it1", function() {expect.hasAssertions();})',
        ],
        [
          'it("it1", function() {expect.hasAssertions("1", "2");})',
          1,
          30,
          'it("it1", function() {expect.hasAssertions();})',
        ],
        [
          `it("it1", function() {
  expect.hasAssertions(() => {
    someFunctionToDo();
    someFunctionToDo2();
  });
});`,
          2,
          10,
          `it("it1", function() {
  expect.hasAssertions();
});`,
        ],
      ] as const
    ).map(([code, line, column, suggestionOutput], index): TestCase => ({
      name: `upstream prefer-expect-assertions / invalid ${index + 16}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'hasAssertionsTakesNoArguments',
          message: '`expect.hasAssertions` expects no arguments',
          line,
          column,
          endLine: line,
          endColumn: column + 13,
          suggestions: [
            {
              messageId: 'suggestRemovingExtraArguments',
              desc: 'Remove extra arguments',
              output: suggestionOutput,
            },
          ],
        },
      ],
    })),
    ...(
      [
        [
          `it("it1", async function() {
  expect(someValue).toBe(true);
});`,
          1,
          3,
          `it("it1", async function() {expect.hasAssertions();
  expect(someValue).toBe(true);
});`,
          `it("it1", async function() {expect.assertions();
  expect(someValue).toBe(true);
});`,
        ],
        [
          `it("returns numbers that are greater than four", async () => {
  for(let thing in things) {
    expect(number).toBeGreaterThan(4);
  }
});`,
          1,
          5,
          `it("returns numbers that are greater than four", async () => {expect.hasAssertions();
  for(let thing in things) {
    expect(number).toBeGreaterThan(4);
  }
});`,
          `it("returns numbers that are greater than four", async () => {expect.assertions();
  for(let thing in things) {
    expect(number).toBeGreaterThan(4);
  }
});`,
        ],
        [
          'it("returns numbers that are greater than four", async () => {' +
            numbersLoopBody,
          1,
          5,
          'it("returns numbers that are greater than four", async () => {expect.hasAssertions();' +
            numbersLoopBody,
          'it("returns numbers that are greater than four", async () => {expect.assertions();' +
            numbersLoopBody,
        ],
        [
          'it("returns numbers that are greater than four", async () => {' +
            numbersLoopBody +
            `
` +
            numbersAboveFiveTest,
          1,
          5,
          'it("returns numbers that are greater than four", async () => {expect.hasAssertions();' +
            numbersLoopBody +
            `
` +
            numbersAboveFiveTest,
          'it("returns numbers that are greater than four", async () => {expect.assertions();' +
            numbersLoopBody +
            `
` +
            numbersAboveFiveTest,
        ],
        [
          `beforeAll(() => { expect.hasAssertions(); });

it("returns numbers that are greater than four", async () => {` +
            numbersLoopBody +
            `
` +
            numbersAboveFiveTest,
          3,
          7,
          `beforeAll(() => { expect.hasAssertions(); });

it("returns numbers that are greater than four", async () => {expect.hasAssertions();` +
            numbersLoopBody +
            `
` +
            numbersAboveFiveTest,
          `beforeAll(() => { expect.hasAssertions(); });

it("returns numbers that are greater than four", async () => {expect.assertions();` +
            numbersLoopBody +
            `
` +
            numbersAboveFiveTest,
        ],
        [
          `afterAll(() => { expect.hasAssertions(); });

it("returns numbers that are greater than four", async () => {` +
            numbersLoopBody +
            `
` +
            numbersAboveFiveTest,
          3,
          7,
          `afterAll(() => { expect.hasAssertions(); });

it("returns numbers that are greater than four", async () => {expect.hasAssertions();` +
            numbersLoopBody +
            `
` +
            numbersAboveFiveTest,
          `afterAll(() => { expect.hasAssertions(); });

it("returns numbers that are greater than four", async () => {expect.assertions();` +
            numbersLoopBody +
            `
` +
            numbersAboveFiveTest,
        ],
      ] as const
    ).map(
      (
        [code, line, endLine, suggestionOutput1, suggestionOutput2],
        index,
      ): TestCase => ({
        name: `upstream prefer-expect-assertions / invalid ${index + 24}`,
        code,
        options: [
          {
            onlyFunctionsWithAsyncKeyword: true,
          },
        ],
        languageOptions: {
          ecmaVersion: 2017,
        },
        errors: [
          {
            messageId: 'haveExpectAssertions',
            message:
              'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
            line,
            column: 1,
            endLine,
            endColumn: 3,
            suggestions: [
              {
                messageId: 'suggestAddingHasAssertions',
                desc: 'Add `expect.hasAssertions()`',
                output: suggestionOutput1,
              },
              {
                messageId: 'suggestAddingAssertions',
                desc: 'Add `expect.assertions(<number of assertions>)`',
                output: suggestionOutput2,
              },
            ],
          },
        ],
      }),
    ),
    ...(
      [
        [
          `it('only returns numbers that are greater than six', () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(6);
  }
});`,
          1,
          1,
          5,
          `it('only returns numbers that are greater than six', () => {expect.hasAssertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(6);
  }
});`,
          `it('only returns numbers that are greater than six', () => {expect.assertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(6);
  }
});`,
        ],
        [
          "it('returns numbers that are greater than two', function () {" +
            helperAssertionBody,
          1,
          1,
          9,
          "it('returns numbers that are greater than two', function () {expect.hasAssertions();" +
            helperAssertionBody,
          "it('returns numbers that are greater than two', function () {expect.assertions();" +
            helperAssertionBody,
        ],
        [
          `it("only returns numbers that are greater than seven", function () {
  const numbers = getNumbers();

  for (let i = 0; i < numbers.length; i++) {
    expect(numbers[i]).toBeGreaterThan(7);
  }
});`,
          1,
          1,
          7,
          `it("only returns numbers that are greater than seven", function () {expect.hasAssertions();
  const numbers = getNumbers();

  for (let i = 0; i < numbers.length; i++) {
    expect(numbers[i]).toBeGreaterThan(7);
  }
});`,
          `it("only returns numbers that are greater than seven", function () {expect.assertions();
  const numbers = getNumbers();

  for (let i = 0; i < numbers.length; i++) {
    expect(numbers[i]).toBeGreaterThan(7);
  }
});`,
        ],
        [
          `it('has the number two', () => {
  expect(number).toBe(2);
});

it('only returns numbers that are less than twenty', () => {
  for (const number of getNumbers()) {
    expect(number).toBeLessThan(20);
  }
});`,
          5,
          1,
          9,
          `it('has the number two', () => {
  expect(number).toBe(2);
});

it('only returns numbers that are less than twenty', () => {expect.hasAssertions();
  for (const number of getNumbers()) {
    expect(number).toBeLessThan(20);
  }
});`,
          `it('has the number two', () => {
  expect(number).toBe(2);
});

it('only returns numbers that are less than twenty', () => {expect.assertions();
  for (const number of getNumbers()) {
    expect(number).toBeLessThan(20);
  }
});`,
        ],
        [
          `it("is wrong");

it("is a test", () => {` + numbersLoopBody,
          3,
          1,
          7,
          `it("is wrong");

it("is a test", () => {expect.hasAssertions();` + numbersLoopBody,
          `it("is wrong");

it("is a test", () => {expect.assertions();` + numbersLoopBody,
        ],
        [
          singleAssertionTest +
            `

it("returns numbers that are greater than four", () => {` +
            numbersLoopBody +
            `
` +
            numberAboveFiveTest,
          5,
          1,
          9,
          singleAssertionTest +
            `

it("returns numbers that are greater than four", () => {expect.hasAssertions();` +
            numbersLoopBody +
            `
` +
            numberAboveFiveTest,
          singleAssertionTest +
            `

it("returns numbers that are greater than four", () => {expect.assertions();` +
            numbersLoopBody +
            `
` +
            numberAboveFiveTest,
        ],
        [
          `describe('my tests', () => {
  beforeEach(expect.hasAssertions);
  it("is a number that is greater than four", () => {
    expect(number).toBeGreaterThan(4);
  });
});

describe('more tests', () => {
  it("returns numbers that are greater than four", () => {
    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(4);
    }
  });
});
` + numberAboveFiveTest,
          9,
          3,
          13,
          `describe('my tests', () => {
  beforeEach(expect.hasAssertions);
  it("is a number that is greater than four", () => {
    expect(number).toBeGreaterThan(4);
  });
});

describe('more tests', () => {
  it("returns numbers that are greater than four", () => {expect.hasAssertions();
    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(4);
    }
  });
});
` + numberAboveFiveTest,
          `describe('my tests', () => {
  beforeEach(expect.hasAssertions);
  it("is a number that is greater than four", () => {
    expect(number).toBeGreaterThan(4);
  });
});

describe('more tests', () => {
  it("returns numbers that are greater than four", () => {expect.assertions();
    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(4);
    }
  });
});
` + numberAboveFiveTest,
        ],
        [
          'it.each([1, 2, 3])("returns numbers that are greater than four", () => {' +
            numbersLoopBody +
            `

` +
            singleAssertionTest,
          1,
          1,
          5,
          'it.each([1, 2, 3])("returns numbers that are greater than four", () => {expect.hasAssertions();' +
            numbersLoopBody +
            `

` +
            singleAssertionTest,
          'it.each([1, 2, 3])("returns numbers that are greater than four", () => {expect.assertions();' +
            numbersLoopBody +
            `

` +
            singleAssertionTest,
        ],
        [
          'it("returns numbers that are greater than four", () => {' +
            numbersLoopBody +
            `

` +
            singleAssertionTest,
          1,
          1,
          5,
          'it("returns numbers that are greater than four", () => {expect.hasAssertions();' +
            numbersLoopBody +
            `

` +
            singleAssertionTest,
          'it("returns numbers that are greater than four", () => {expect.assertions();' +
            numbersLoopBody +
            `

` +
            singleAssertionTest,
        ],
        [
          'it("returns numbers that are greater than four", () => {' +
            numbersLoopBody +
            `

it("is a number that is greater than four", () => {
  expect.hasAssertions();

  expect(number).toBeGreaterThan(4);
});`,
          1,
          1,
          5,
          'it("returns numbers that are greater than four", () => {expect.hasAssertions();' +
            numbersLoopBody +
            `

it("is a number that is greater than four", () => {
  expect.hasAssertions();

  expect(number).toBeGreaterThan(4);
});`,
          'it("returns numbers that are greater than four", () => {expect.assertions();' +
            numbersLoopBody +
            `

it("is a number that is greater than four", () => {
  expect.hasAssertions();

  expect(number).toBeGreaterThan(4);
});`,
        ],
        [
          `it("it1", () => {
  expect.hasAssertions();

  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});

it("it1", () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});`,
          9,
          1,
          13,
          `it("it1", () => {
  expect.hasAssertions();

  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});

it("it1", () => {expect.hasAssertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});`,
          `it("it1", () => {
  expect.hasAssertions();

  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});

it("it1", () => {expect.assertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});`,
        ],
      ] as const
    ).map(
      (
        [code, line, column, endLine, suggestionOutput1, suggestionOutput2],
        index,
      ): TestCase => ({
        name: `upstream prefer-expect-assertions (loops) / invalid ${index}`,
        code,
        options: [
          {
            onlyFunctionsWithExpectInLoop: true,
          },
        ],
        languageOptions: {
          ecmaVersion: 2017,
        },
        errors: [
          {
            messageId: 'haveExpectAssertions',
            message:
              'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
            line,
            column,
            endLine,
            endColumn: column + 2,
            suggestions: [
              {
                messageId: 'suggestAddingHasAssertions',
                desc: 'Add `expect.hasAssertions()`',
                output: suggestionOutput1,
              },
              {
                messageId: 'suggestAddingAssertions',
                desc: 'Add `expect.assertions(<number of assertions>)`',
                output: suggestionOutput2,
              },
            ],
          },
        ],
      }),
    ),
    {
      name: 'upstream prefer-expect-assertions (loops) / invalid 11',
      code:
        `it("returns numbers that are greater than four", async () => {
  for (const number of await getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});
` + numbersAboveFiveTest,
      options: [
        {
          onlyFunctionsWithExpectInLoop: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'haveExpectAssertions',
          message:
            'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
          line: 1,
          column: 1,
          endLine: 5,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              desc: 'Add `expect.hasAssertions()`',
              output:
                `it("returns numbers that are greater than four", async () => {expect.hasAssertions();
  for (const number of await getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});
` + numbersAboveFiveTest,
            },
            {
              messageId: 'suggestAddingAssertions',
              desc: 'Add `expect.assertions(<number of assertions>)`',
              output:
                `it("returns numbers that are greater than four", async () => {expect.assertions();
  for (const number of await getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});
` + numbersAboveFiveTest,
            },
          ],
        },
        {
          messageId: 'haveExpectAssertions',
          message:
            'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
          line: 7,
          column: 1,
          endLine: 11,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              desc: 'Add `expect.hasAssertions()`',
              output: `it("returns numbers that are greater than four", async () => {
  for (const number of await getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});

it("returns numbers that are greater than five", () => {expect.hasAssertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(5);
  }
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              desc: 'Add `expect.assertions(<number of assertions>)`',
              output: `it("returns numbers that are greater than four", async () => {
  for (const number of await getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});

it("returns numbers that are greater than five", () => {expect.assertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(5);
  }
});`,
            },
          ],
        },
      ],
    },
    ...(
      [
        [
          `it("it1", async () => {
  expect.hasAssertions();
` +
            numbersLoopBody +
            `

it("it1", () => {` +
            numbersLoopBody,
          9,
          1,
          13,
          `it("it1", async () => {
  expect.hasAssertions();
` +
            numbersLoopBody +
            `

it("it1", () => {expect.hasAssertions();` +
            numbersLoopBody,
          `it("it1", async () => {
  expect.hasAssertions();
` +
            numbersLoopBody +
            `

it("it1", () => {expect.assertions();` +
            numbersLoopBody,
        ],
        [
          `describe('my tests', () => {
  beforeEach(() => { expect.hasAssertions(); });

  it("it1", async () => {
    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(4);
    }
  });
});

it("it1", () => {` + numbersLoopBody,
          11,
          1,
          15,
          `describe('my tests', () => {
  beforeEach(() => { expect.hasAssertions(); });

  it("it1", async () => {
    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(4);
    }
  });
});

it("it1", () => {expect.hasAssertions();` + numbersLoopBody,
          `describe('my tests', () => {
  beforeEach(() => { expect.hasAssertions(); });

  it("it1", async () => {
    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(4);
    }
  });
});

it("it1", () => {expect.assertions();` + numbersLoopBody,
        ],
        [
          `describe('my tests', () => {
  afterEach(() => { expect.hasAssertions(); });

  it("it1", async () => {
    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(4);
    }
  });
});

it("it1", () => {` + numbersLoopBody,
          11,
          1,
          15,
          `describe('my tests', () => {
  afterEach(() => { expect.hasAssertions(); });

  it("it1", async () => {
    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(4);
    }
  });
});

it("it1", () => {expect.hasAssertions();` + numbersLoopBody,
          `describe('my tests', () => {
  afterEach(() => { expect.hasAssertions(); });

  it("it1", async () => {
    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(4);
    }
  });
});

it("it1", () => {expect.assertions();` + numbersLoopBody,
        ],
        [
          `it.skip.each\`\`("it1", async () => {
  expect.hasAssertions();
` +
            numbersLoopBody +
            `

it("it1", () => {` +
            numbersLoopBody,
          9,
          1,
          13,
          `it.skip.each\`\`("it1", async () => {
  expect.hasAssertions();
` +
            numbersLoopBody +
            `

it("it1", () => {expect.hasAssertions();` +
            numbersLoopBody,
          `it.skip.each\`\`("it1", async () => {
  expect.hasAssertions();
` +
            numbersLoopBody +
            `

it("it1", () => {expect.assertions();` +
            numbersLoopBody,
        ],
        [
          'it("it1", async () => {' +
            numbersLoopBody +
            `

it("it1", () => {
  expect.hasAssertions();
` +
            numbersLoopBody,
          1,
          1,
          5,
          'it("it1", async () => {expect.hasAssertions();' +
            numbersLoopBody +
            `

it("it1", () => {
  expect.hasAssertions();
` +
            numbersLoopBody,
          'it("it1", async () => {expect.assertions();' +
            numbersLoopBody +
            `

it("it1", () => {
  expect.hasAssertions();
` +
            numbersLoopBody,
        ],
        [
          `describe('my tests', () => {
  it("it1", async () => {
    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(4);
    }
  });
});

it("it1", () => {
  expect.hasAssertions();
` + numbersLoopBody,
          2,
          3,
          6,
          `describe('my tests', () => {
  it("it1", async () => {expect.hasAssertions();
    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(4);
    }
  });
});

it("it1", () => {
  expect.hasAssertions();
` + numbersLoopBody,
          `describe('my tests', () => {
  it("it1", async () => {expect.assertions();
    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(4);
    }
  });
});

it("it1", () => {
  expect.hasAssertions();
` + numbersLoopBody,
        ],
      ] as const
    ).map(
      (
        [code, line, column, endLine, suggestionOutput1, suggestionOutput2],
        index,
      ): TestCase => ({
        name: `upstream prefer-expect-assertions (loops) / invalid ${index + 12}`,
        code,
        options: [
          {
            onlyFunctionsWithExpectInLoop: true,
          },
        ],
        languageOptions: {
          ecmaVersion: 2017,
        },
        errors: [
          {
            messageId: 'haveExpectAssertions',
            message:
              'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
            line,
            column,
            endLine,
            endColumn: column + 2,
            suggestions: [
              {
                messageId: 'suggestAddingHasAssertions',
                desc: 'Add `expect.hasAssertions()`',
                output: suggestionOutput1,
              },
              {
                messageId: 'suggestAddingAssertions',
                desc: 'Add `expect.assertions(<number of assertions>)`',
                output: suggestionOutput2,
              },
            ],
          },
        ],
      }),
    ),
    ...(
      [
        [
          `it('sends the data as a string', () => {
  const stream = openStream();

  stream.on('data', data => {
    expect(data).toBe(expect.any(String));
  });
});`,
          1,
          7,
          `it('sends the data as a string', () => {expect.hasAssertions();
  const stream = openStream();

  stream.on('data', data => {
    expect(data).toBe(expect.any(String));
  });
});`,
          `it('sends the data as a string', () => {expect.assertions();
  const stream = openStream();

  stream.on('data', data => {
    expect(data).toBe(expect.any(String));
  });
});`,
        ],
        [
          "it('responds ok', function () {" + requestCallbackBody,
          1,
          5,
          "it('responds ok', function () {expect.hasAssertions();" +
            requestCallbackBody,
          "it('responds ok', function () {expect.assertions();" +
            requestCallbackBody,
        ],
        [
          `it('responds ok', function () {
  client.get('/user', response => {
    expect.assertions(1);

    expect(response.status).toBe(200);
  });
});`,
          1,
          7,
          `it('responds ok', function () {expect.hasAssertions();
  client.get('/user', response => {
    expect.assertions(1);

    expect(response.status).toBe(200);
  });
});`,
          `it('responds ok', function () {expect.assertions();
  client.get('/user', response => {
    expect.assertions(1);

    expect(response.status).toBe(200);
  });
});`,
        ],
        [
          `it('responds ok', function () {
` + requestCallbackWithAssertionCount,
          1,
          9,
          `it('responds ok', function () {expect.hasAssertions();
` + requestCallbackWithAssertionCount,
          `it('responds ok', function () {expect.assertions();
` + requestCallbackWithAssertionCount,
        ],
        [
          `it('returns numbers that are greater than two', function () {
  const expectNumberToBeGreaterThan = (number, value) => {
    expect(number).toBeGreaterThan(value);
  };

  expectNumberToBeGreaterThan(1, 2);
});`,
          1,
          7,
          `it('returns numbers that are greater than two', function () {expect.hasAssertions();
  const expectNumberToBeGreaterThan = (number, value) => {
    expect(number).toBeGreaterThan(value);
  };

  expectNumberToBeGreaterThan(1, 2);
});`,
          `it('returns numbers that are greater than two', function () {expect.assertions();
  const expectNumberToBeGreaterThan = (number, value) => {
    expect(number).toBeGreaterThan(value);
  };

  expectNumberToBeGreaterThan(1, 2);
});`,
        ],
        [
          "it('returns numbers that are greater than two', function () {" +
            helperAssertionBody,
          1,
          9,
          "it('returns numbers that are greater than two', function () {expect.hasAssertions();" +
            helperAssertionBody,
          "it('returns numbers that are greater than two', function () {expect.assertions();" +
            helperAssertionBody,
        ],
        [
          `it('only returns numbers that are greater than six', () => {
  getNumbers().forEach(number => {
    expect(number).toBeGreaterThan(6);
  });
});`,
          1,
          5,
          `it('only returns numbers that are greater than six', () => {expect.hasAssertions();
  getNumbers().forEach(number => {
    expect(number).toBeGreaterThan(6);
  });
});`,
          `it('only returns numbers that are greater than six', () => {expect.assertions();
  getNumbers().forEach(number => {
    expect(number).toBeGreaterThan(6);
  });
});`,
        ],
        [
          `it("is wrong");

it('responds ok', function () {
` + requestCallbackWithAssertionCount,
          3,
          11,
          `it("is wrong");

it('responds ok', function () {expect.hasAssertions();
` + requestCallbackWithAssertionCount,
          `it("is wrong");

it('responds ok', function () {expect.assertions();
` + requestCallbackWithAssertionCount,
        ],
        [
          singleAssertionTest +
            `

it('responds ok', function () {
  const expectOkResponse = response => {
    expect(response.status).toBe(200);
  };

  client.get('/user', expectOkResponse);
});
` +
            numberAboveFiveTest,
          5,
          11,
          singleAssertionTest +
            `

it('responds ok', function () {expect.hasAssertions();
  const expectOkResponse = response => {
    expect(response.status).toBe(200);
  };

  client.get('/user', expectOkResponse);
});
` +
            numberAboveFiveTest,
          singleAssertionTest +
            `

it('responds ok', function () {expect.assertions();
  const expectOkResponse = response => {
    expect(response.status).toBe(200);
  };

  client.get('/user', expectOkResponse);
});
` +
            numberAboveFiveTest,
        ],
        [
          singleAssertionTest +
            `

it("returns numbers that are greater than four", () => {
  getNumbers().map(number => {
    expect(number).toBeGreaterThan(0);
  });
});
` +
            numberAboveFiveTest,
          5,
          9,
          singleAssertionTest +
            `

it("returns numbers that are greater than four", () => {expect.hasAssertions();
  getNumbers().map(number => {
    expect(number).toBeGreaterThan(0);
  });
});
` +
            numberAboveFiveTest,
          singleAssertionTest +
            `

it("returns numbers that are greater than four", () => {expect.assertions();
  getNumbers().map(number => {
    expect(number).toBeGreaterThan(0);
  });
});
` +
            numberAboveFiveTest,
        ],
        [
          `it.each([1, 2, 3])("returns ok", id => {
  client.get(\`/users/\${id}\`, response => {
    expect(response.status).toBe(200);
  });
});

` + singleAssertionTest,
          1,
          5,
          `it.each([1, 2, 3])("returns ok", id => {expect.hasAssertions();
  client.get(\`/users/\${id}\`, response => {
    expect(response.status).toBe(200);
  });
});

` + singleAssertionTest,
          `it.each([1, 2, 3])("returns ok", id => {expect.assertions();
  client.get(\`/users/\${id}\`, response => {
    expect(response.status).toBe(200);
  });
});

` + singleAssertionTest,
        ],
        [
          "it('responds ok', function () {" +
            requestCallbackBody +
            `

` +
            singleAssertionTest,
          1,
          5,
          "it('responds ok', function () {expect.hasAssertions();" +
            requestCallbackBody +
            `

` +
            singleAssertionTest,
          "it('responds ok', function () {expect.assertions();" +
            requestCallbackBody +
            `

` +
            singleAssertionTest,
        ],
        [
          "it('responds ok', function () {" +
            requestCallbackBody +
            `

it("is a number that is greater than four", () => {
  expect.hasAssertions();

  expect(number).toBeGreaterThan(4);
});`,
          1,
          5,
          "it('responds ok', function () {expect.hasAssertions();" +
            requestCallbackBody +
            `

it("is a number that is greater than four", () => {
  expect.hasAssertions();

  expect(number).toBeGreaterThan(4);
});`,
          "it('responds ok', function () {expect.assertions();" +
            requestCallbackBody +
            `

it("is a number that is greater than four", () => {
  expect.hasAssertions();

  expect(number).toBeGreaterThan(4);
});`,
        ],
        [
          `it("it1", () => {
  expect.hasAssertions();

  getNumbers().forEach(number => {
    expect(number).toBeGreaterThan(0);
  });
});

it("it1", () => {
  getNumbers().forEach(number => {
    expect(number).toBeGreaterThan(0);
  });
});`,
          9,
          13,
          `it("it1", () => {
  expect.hasAssertions();

  getNumbers().forEach(number => {
    expect(number).toBeGreaterThan(0);
  });
});

it("it1", () => {expect.hasAssertions();
  getNumbers().forEach(number => {
    expect(number).toBeGreaterThan(0);
  });
});`,
          `it("it1", () => {
  expect.hasAssertions();

  getNumbers().forEach(number => {
    expect(number).toBeGreaterThan(0);
  });
});

it("it1", () => {expect.assertions();
  getNumbers().forEach(number => {
    expect(number).toBeGreaterThan(0);
  });
});`,
        ],
        [
          `it('responds ok', function () {
  expect.hasAssertions();
` +
            requestCallbackBody +
            `

it('responds not found', function () {
  client.get('/user', response => {
    expect(response.status).toBe(404);
  });
});`,
          9,
          13,
          `it('responds ok', function () {
  expect.hasAssertions();
` +
            requestCallbackBody +
            `

it('responds not found', function () {expect.hasAssertions();
  client.get('/user', response => {
    expect(response.status).toBe(404);
  });
});`,
          `it('responds ok', function () {
  expect.hasAssertions();
` +
            requestCallbackBody +
            `

it('responds not found', function () {expect.assertions();
  client.get('/user', response => {
    expect(response.status).toBe(404);
  });
});`,
        ],
        [
          `it.skip.each\`\`("it1", async () => {
  expect.hasAssertions();
` +
            requestCallbackBody +
            `

it("responds ok", () => {` +
            requestCallbackBody,
          9,
          13,
          `it.skip.each\`\`("it1", async () => {
  expect.hasAssertions();
` +
            requestCallbackBody +
            `

it("responds ok", () => {expect.hasAssertions();` +
            requestCallbackBody,
          `it.skip.each\`\`("it1", async () => {
  expect.hasAssertions();
` +
            requestCallbackBody +
            `

it("responds ok", () => {expect.assertions();` +
            requestCallbackBody,
        ],
      ] as const
    ).map(
      (
        [code, line, endLine, suggestionOutput1, suggestionOutput2],
        index,
      ): TestCase => ({
        name: `upstream prefer-expect-assertions (callbacks) / invalid ${index}`,
        code,
        options: [
          {
            onlyFunctionsWithExpectInCallback: true,
          },
        ],
        languageOptions: {
          ecmaVersion: 2017,
        },
        errors: [
          {
            messageId: 'haveExpectAssertions',
            message:
              'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
            line,
            column: 1,
            endLine,
            endColumn: 3,
            suggestions: [
              {
                messageId: 'suggestAddingHasAssertions',
                desc: 'Add `expect.hasAssertions()`',
                output: suggestionOutput1,
              },
              {
                messageId: 'suggestAddingAssertions',
                desc: 'Add `expect.assertions(<number of assertions>)`',
                output: suggestionOutput2,
              },
            ],
          },
        ],
      }),
    ),
    {
      name: 'upstream prefer-expect-assertions (callbacks) / invalid 16',
      code: `it("returns numbers that are greater than four", function(expect) {
  expect.assertions(2);

  for(let thing in things) {
    expect(number).toBeGreaterThan(4);
  }
});`,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'haveExpectAssertions',
          message:
            'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
          line: 1,
          column: 1,
          endLine: 7,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              desc: 'Add `expect.hasAssertions()`',
              output: `it("returns numbers that are greater than four", function(expect) {expect.hasAssertions();
  expect.assertions(2);

  for(let thing in things) {
    expect(number).toBeGreaterThan(4);
  }
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              desc: 'Add `expect.assertions(<number of assertions>)`',
              output: `it("returns numbers that are greater than four", function(expect) {expect.assertions();
  expect.assertions(2);

  for(let thing in things) {
    expect(number).toBeGreaterThan(4);
  }
});`,
            },
          ],
        },
      ],
    },
    ...(
      [
        [
          `it('only returns numbers that are greater than zero', () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});

it("is zero", () => {
  expect.hasAssertions();

  expect(0).toBe(0);
});`,
          1,
          5,
          `it('only returns numbers that are greater than zero', () => {expect.hasAssertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});

it("is zero", () => {
  expect.hasAssertions();

  expect(0).toBe(0);
});`,
          `it('only returns numbers that are greater than zero', () => {expect.assertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});

it("is zero", () => {
  expect.hasAssertions();

  expect(0).toBe(0);
});`,
        ],
        [
          `it('only returns numbers that are greater than zero', () => {
  expect.hasAssertions();

  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});

it('only returns numbers that are less than 100', () => {
  for (const number of getNumbers()) {
    expect(number).toBeLessThan(0);
  }
});`,
          9,
          13,
          `it('only returns numbers that are greater than zero', () => {
  expect.hasAssertions();

  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});

it('only returns numbers that are less than 100', () => {expect.hasAssertions();
  for (const number of getNumbers()) {
    expect(number).toBeLessThan(0);
  }
});`,
          `it('only returns numbers that are greater than zero', () => {
  expect.hasAssertions();

  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});

it('only returns numbers that are less than 100', () => {expect.assertions();
  for (const number of getNumbers()) {
    expect(number).toBeLessThan(0);
  }
});`,
        ],
      ] as const
    ).map(
      (
        [code, line, endLine, suggestionOutput1, suggestionOutput2],
        index,
      ): TestCase => ({
        name: `upstream prefer-expect-assertions (mixed) / invalid ${index}`,
        code,
        options: [
          {
            onlyFunctionsWithExpectInLoop: true,
          },
        ],
        languageOptions: {
          ecmaVersion: 2017,
        },
        errors: [
          {
            messageId: 'haveExpectAssertions',
            message:
              'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
            line,
            column: 1,
            endLine,
            endColumn: 3,
            suggestions: [
              {
                messageId: 'suggestAddingHasAssertions',
                desc: 'Add `expect.hasAssertions()`',
                output: suggestionOutput1,
              },
              {
                messageId: 'suggestAddingAssertions',
                desc: 'Add `expect.assertions(<number of assertions>)`',
                output: suggestionOutput2,
              },
            ],
          },
        ],
      }),
    ),
    ...(
      [
        [
          `it("to be true", async function() {
  expect(someValue).toBe(true);
});`,
          3,
          `it("to be true", async function() {expect.hasAssertions();
  expect(someValue).toBe(true);
});`,
          `it("to be true", async function() {expect.assertions();
  expect(someValue).toBe(true);
});`,
        ],
        [
          `it('only returns numbers that are greater than zero', async () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});`,
          5,
          `it('only returns numbers that are greater than zero', async () => {expect.hasAssertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});`,
          `it('only returns numbers that are greater than zero', async () => {expect.assertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});`,
        ],
      ] as const
    ).map(
      (
        [code, endLine, suggestionOutput1, suggestionOutput2],
        index,
      ): TestCase => ({
        name: `upstream prefer-expect-assertions (mixed) / invalid ${index + 2}`,
        code,
        options: [
          {
            onlyFunctionsWithAsyncKeyword: true,
            onlyFunctionsWithExpectInLoop: true,
          },
        ],
        languageOptions: {
          ecmaVersion: 2017,
        },
        errors: [
          {
            messageId: 'haveExpectAssertions',
            message:
              'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
            line: 1,
            column: 1,
            endLine,
            endColumn: 3,
            suggestions: [
              {
                messageId: 'suggestAddingHasAssertions',
                desc: 'Add `expect.hasAssertions()`',
                output: suggestionOutput1,
              },
              {
                messageId: 'suggestAddingAssertions',
                desc: 'Add `expect.assertions(<number of assertions>)`',
                output: suggestionOutput2,
              },
            ],
          },
        ],
      }),
    ),
    ...(
      [
        [
          `test.each()("is not fine", () => {
  expect(someValue).toBe(true);
});`,
          1,
          1,
          3,
          `test.each()("is not fine", () => {expect.hasAssertions();
  expect(someValue).toBe(true);
});`,
          `test.each()("is not fine", () => {expect.assertions();
  expect(someValue).toBe(true);
});`,
        ],
        [
          `describe.each()('something', () => {
  it("is not fine", () => {
    expect(someValue).toBe(true);
  });
});`,
          2,
          3,
          4,
          `describe.each()('something', () => {
  it("is not fine", () => {expect.hasAssertions();
    expect(someValue).toBe(true);
  });
});`,
          `describe.each()('something', () => {
  it("is not fine", () => {expect.assertions();
    expect(someValue).toBe(true);
  });
});`,
        ],
        [
          `describe.each()('something', () => {
  test.each()("is not fine", () => {
    expect(someValue).toBe(true);
  });
});`,
          2,
          3,
          4,
          `describe.each()('something', () => {
  test.each()("is not fine", () => {expect.hasAssertions();
    expect(someValue).toBe(true);
  });
});`,
          `describe.each()('something', () => {
  test.each()("is not fine", () => {expect.assertions();
    expect(someValue).toBe(true);
  });
});`,
        ],
      ] as const
    ).map(
      (
        [code, line, column, endLine, suggestionOutput1, suggestionOutput2],
        index,
      ): TestCase => ({
        name: `upstream .each support / invalid ${index}`,
        code,
        languageOptions: {
          ecmaVersion: 2017,
        },
        errors: [
          {
            messageId: 'haveExpectAssertions',
            message:
              'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
            line,
            column,
            endLine,
            endColumn: column + 2,
            suggestions: [
              {
                messageId: 'suggestAddingHasAssertions',
                desc: 'Add `expect.hasAssertions()`',
                output: suggestionOutput1,
              },
              {
                messageId: 'suggestAddingAssertions',
                desc: 'Add `expect.assertions(<number of assertions>)`',
                output: suggestionOutput2,
              },
            ],
          },
        ],
      }),
    ),
    ...(
      [
        [
          `test.each()("is not fine", async () => {
  expect(someValue).toBe(true);
});`,
          1,
          1,
          3,
          `test.each()("is not fine", async () => {expect.hasAssertions();
  expect(someValue).toBe(true);
});`,
          `test.each()("is not fine", async () => {expect.assertions();
  expect(someValue).toBe(true);
});`,
        ],
        [
          `it.each()("is not fine", async () => {
  expect(someValue).toBe(true);
});`,
          1,
          1,
          3,
          `it.each()("is not fine", async () => {expect.hasAssertions();
  expect(someValue).toBe(true);
});`,
          `it.each()("is not fine", async () => {expect.assertions();
  expect(someValue).toBe(true);
});`,
        ],
        [
          `describe.each()('something', () => {
  test.each()("is not fine", async () => {
    expect(someValue).toBe(true);
  });
});`,
          2,
          3,
          4,
          `describe.each()('something', () => {
  test.each()("is not fine", async () => {expect.hasAssertions();
    expect(someValue).toBe(true);
  });
});`,
          `describe.each()('something', () => {
  test.each()("is not fine", async () => {expect.assertions();
    expect(someValue).toBe(true);
  });
});`,
        ],
      ] as const
    ).map(
      (
        [code, line, column, endLine, suggestionOutput1, suggestionOutput2],
        index,
      ): TestCase => ({
        name: `upstream .each support / invalid ${index + 3}`,
        code,
        options: [
          {
            onlyFunctionsWithAsyncKeyword: true,
          },
        ],
        languageOptions: {
          ecmaVersion: 2017,
        },
        errors: [
          {
            messageId: 'haveExpectAssertions',
            message:
              'Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression',
            line,
            column,
            endLine,
            endColumn: column + 2,
            suggestions: [
              {
                messageId: 'suggestAddingHasAssertions',
                desc: 'Add `expect.hasAssertions()`',
                output: suggestionOutput1,
              },
              {
                messageId: 'suggestAddingAssertions',
                desc: 'Add `expect.assertions(<number of assertions>)`',
                output: suggestionOutput2,
              },
            ],
          },
        ],
      }),
    ),
  ],
});
