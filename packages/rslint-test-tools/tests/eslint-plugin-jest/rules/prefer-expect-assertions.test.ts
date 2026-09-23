import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-expect-assertions', {} as never, {
  valid: [
    {
      code: `test("nonsense", [])`,
    },
    {
      code: `test("it1", () => {expect.assertions(0);})`,
    },
    {
      code: `test("it1", function() {expect.assertions(0);})`,
    },
    {
      code: `test("it1", function() {expect.hasAssertions();})`,
    },
    {
      code: `it("it1", function() {expect.assertions(0);})`,
    },
    {
      code: `it("it1", function() {
  expect.assertions(1);
  expect(someValue).toBe(true)
});`,
    },
    {
      code: `test("it1")`,
    },
    {
      code: `itHappensToStartWithIt("foo", function() {})`,
    },
    {
      code: `testSomething("bar", function() {})`,
    },
    {
      code: `it(async () => {expect.assertions(0);})`,
    },
    {
      code: `it("returns numbers that are greater than four", function() {
  expect.assertions(2);

  for(let thing in things) {
    expect(number).toBeGreaterThan(4);
  }
});`,
    },
    {
      code: `it("returns numbers that are greater than four", function() {
  expect.hasAssertions();

  for (let i = 0; i < things.length; i++) {
    expect(number).toBeGreaterThan(4);
  }
});`,
    },
    {
      code: `it("it1", async () => {
  expect.assertions(1);
  expect(someValue).toBe(true)
});`,
      options: [{ onlyFunctionsWithAsyncKeyword: true }],
    },
    {
      code: `it("it1", function() {
  expect(someValue).toBe(true)
});`,
      options: [{ onlyFunctionsWithAsyncKeyword: true }],
    },
    {
      code: `it("it1", () => {})`,
      options: [{ onlyFunctionsWithAsyncKeyword: true }],
    },
    {
      code: `it("returns numbers that are greater than four", async () => {
  expect.assertions(2);

  for(let thing in things) {
    expect(number).toBeGreaterThan(4);
  }
});`,
      options: [{ onlyFunctionsWithAsyncKeyword: true }],
    },
    {
      code: `it("returns numbers that are greater than four", () => {
  for(let thing in things) {
    expect(number).toBeGreaterThan(4);
  }
});`,
      options: [{ onlyFunctionsWithAsyncKeyword: true }],
    },
    {
      code: `import { expect as pleaseExpect } from '@jest/globals';

it("returns numbers that are greater than four", function() {
  pleaseExpect.assertions(2);

  for(let thing in things) {
    pleaseExpect(number).toBeGreaterThan(4);
  }
});`,
    },
    {
      code: `beforeEach(() => expect.hasAssertions());

it('responds ok', function () {
  client.get('/user', response => {
    expect(response.status).toBe(200);
  });
});

it("is a number that is greater than four", () => {
  expect(number).toBeGreaterThan(4);
});`,
    },
    {
      code: `afterEach(() => {
  expect.hasAssertions();
});

it('responds ok', function () {
  client.get('/user', response => {
    expect(response.status).toBe(200);
  });
});

it("is a number that is greater than four", () => {
  expect(number).toBeGreaterThan(4);
});`,
    },
    {
      code: `afterEach(() => {
  expect.hasAssertions();
});

it('responds ok', function () {
  client.get('/user', response => {
    expect(response.status).toBe(200);
  });
});

it("is a number that is greater than four", () => {
  expect.hasAssertions();

  expect(number).toBeGreaterThan(4);
});`,
    },
    {
      code: `beforeEach(() => { expect.hasAssertions(); });

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
    },
    {
      code: `describe('my tests', () => {
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
    },
    {
      code: `describe('my tests', () => {
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
    },
    {
      code: `describe('my tests', () => {
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
    },
    {
      code: `describe('my tests', () => {
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
    },
    {
      code: `describe('my tests', () => {
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
    },
    {
      code: `const expectNumbersToBeGreaterThan = (numbers, value) => {
  for (let number of numbers) {
    expect(number).toBeGreaterThan(value);
  }
};

it('returns numbers that are greater than two', function () {
  expectNumbersToBeGreaterThan(getNumbers(), 2);
});`,
      options: [{ onlyFunctionsWithExpectInLoop: true }],
    },
    {
      code: `it("returns numbers that are greater than five", function () {
  expect.assertions(2);

  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(5);
  }
});`,
      options: [{ onlyFunctionsWithExpectInLoop: true }],
    },
    {
      code: `it("returns things that are less than ten", function () {
  expect.hasAssertions();

  for (const thing in things) {
    expect(thing).toBeLessThan(10);
  }
});`,
      options: [{ onlyFunctionsWithExpectInLoop: true }],
    },
    {
      code: `const expectNumbersToBeGreaterThan = (numbers, value) => {
  numbers.forEach(number => {
    expect(number).toBeGreaterThan(value);
  });
};

it('returns numbers that are greater than two', function () {
  expectNumbersToBeGreaterThan(getNumbers(), 2);
});`,
      options: [{ onlyFunctionsWithExpectInCallback: true }],
    },
    {
      code: `it('returns numbers that are greater than two', function () {
  expect.assertions(2);

  const expectNumbersToBeGreaterThan = (numbers, value) => {
    for (let number of numbers) {
      expect(number).toBeGreaterThan(value);
    }
  };

  expectNumbersToBeGreaterThan(getNumbers(), 2);
});`,
      options: [{ onlyFunctionsWithExpectInCallback: true }],
    },
    {
      code: `beforeEach(() => expect.hasAssertions());

it('returns numbers that are greater than two', function () {
  const expectNumbersToBeGreaterThan = (numbers, value) => {
    for (let number of numbers) {
      expect(number).toBeGreaterThan(value);
    }
  };

  expectNumbersToBeGreaterThan(getNumbers(), 2);
});`,
      options: [{ onlyFunctionsWithExpectInCallback: true }],
    },
    {
      code: `it("returns numbers that are greater than five", function () {
  expect.assertions(2);

  getNumbers().forEach(number => {
    expect(number).toBeGreaterThan(5);
  });
});`,
      options: [{ onlyFunctionsWithExpectInCallback: true }],
    },
    {
      code: `it("returns things that are less than ten", function () {
  expect.hasAssertions();

  things.forEach(thing => {
    expect(thing).toBeLessThan(10);
  });
});`,
      options: [{ onlyFunctionsWithExpectInCallback: true }],
    },
    {
      code: `it('sends the data as a string', () => {
  expect.hasAssertions();

  const stream = openStream();

  stream.on('data', data => {
    expect(data).toBe(expect.any(String));
  });
});`,
      options: [{ onlyFunctionsWithExpectInCallback: true }],
    },
    {
      code: `it('responds ok', function () {
  expect.assertions(1);

  client.get('/user', response => {
    expect(response.status).toBe(200);
  });
});`,
      options: [{ onlyFunctionsWithExpectInCallback: true }],
    },
    {
      code: `it.each([1, 2, 3])("returns ok", id => {
  expect.assertions(3);

  client.get(\`/users/\${id}\`, response => {
    expect(response.status).toBe(200);
  });
});

it("is a number that is greater than four", () => {
  expect(number).toBeGreaterThan(4);
});`,
      options: [{ onlyFunctionsWithExpectInCallback: true }],
    },
    {
      code: `it('is a test', () => {
  expect(expected).toBe(actual);
});

describe('my test', () => {
  it('is another test', () => {
    expect(expected).toBe(actual);
  });
});`,
      options: [{ onlyFunctionsWithExpectInCallback: true }],
    },
    {
      code: `it('responds ok', function () {
  expect.assertions(1);

  client.get('/user', response => {
    expect(response.status).toBe(200);
  });
});

describe('my test', () => {
  beforeEach(() => expect.hasAssertions());

  it('responds ok', function () {
    client.get('/user', response => {
      expect(response.status).toBe(200);
    });
  });
});`,
      options: [{ onlyFunctionsWithExpectInCallback: true }],
    },
    {
      code: `it('responds ok', function () {
  expect.assertions(1);

  client.get('/user', response => {
    expect(response.status).toBe(200);
  });
});

describe('my test', () => {
  afterEach(() => expect.hasAssertions());

  it('responds ok', function () {
    client.get('/user', response => {
      expect(response.status).toBe(200);
    });
  });
});`,
      options: [{ onlyFunctionsWithExpectInCallback: true }],
    },
    {
      code: `it('only returns numbers that are greater than zero', async () => {
  expect.hasAssertions();

  for (const number of await getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});`,
      options: [
        {
          onlyFunctionsWithAsyncKeyword: true,
          onlyFunctionsWithExpectInLoop: true,
        },
      ],
    },
    {
      code: `it('only returns numbers that are greater than zero', async () => {
  expect.assertions(2);

  for (const number of await getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});`,
      options: [
        {
          onlyFunctionsWithAsyncKeyword: true,
          onlyFunctionsWithExpectInLoop: true,
        },
      ],
    },
    {
      code: `test.each()("is fine", () => { expect.assertions(0); })`,
    },
    {
      code: `test.each\`\`("is fine", () => { expect.assertions(0); })`,
    },
    {
      code: `test.each()("is fine", () => { expect.hasAssertions(); })`,
    },
    {
      code: `test.each\`\`("is fine", () => { expect.hasAssertions(); })`,
    },
    {
      code: `it.each()("is fine", () => { expect.assertions(0); })`,
    },
    {
      code: `it.each\`\`("is fine", () => { expect.assertions(0); })`,
    },
    {
      code: `it.each()("is fine", () => { expect.hasAssertions(); })`,
    },
    {
      code: `it.each\`\`("is fine", () => { expect.hasAssertions(); })`,
    },
    {
      code: `test.each()("is fine", () => {})`,
      options: [{ onlyFunctionsWithAsyncKeyword: true }],
    },
    {
      code: `test.each\`\`("is fine", () => {})`,
      options: [{ onlyFunctionsWithAsyncKeyword: true }],
    },
    {
      code: `it.each()("is fine", () => {})`,
      options: [{ onlyFunctionsWithAsyncKeyword: true }],
    },
    {
      code: `it.each\`\`("is fine", () => {})`,
      options: [{ onlyFunctionsWithAsyncKeyword: true }],
    },
    {
      code: `describe.each(['hello'])('%s', () => {
  it('is fine', () => {
    expect.assertions(0);
  });
});`,
    },
    {
      code: `describe.each\`\`('%s', () => {
  it('is fine', () => {
    expect.assertions(0);
  });
});`,
    },
    {
      code: `describe.each(['hello'])('%s', () => {
  it('is fine', () => {
    expect.hasAssertions();
  });
});`,
    },
    {
      code: `describe.each\`\`('%s', () => {
  it('is fine', () => {
    expect.hasAssertions();
  });
});`,
    },
    {
      code: `describe.each(['hello'])('%s', () => {
  it.each()('is fine', () => {
    expect.assertions(0);
  });
});`,
    },
    {
      code: `describe.each\`\`('%s', () => {
  it.each()('is fine', () => {
    expect.assertions(0);
  });
});`,
    },
    {
      code: `describe.each(['hello'])('%s', () => {
  it.each()('is fine', () => {
    expect.hasAssertions();
  });
});`,
    },
    {
      code: `describe.each\`\`('%s', () => {
  it.each()('is fine', () => {
    expect.hasAssertions();
  });
});`,
    },
    // Upstream reports the first test because it tracks hooks in source
    // order; Jest collects every hook of a describe block before running it.
    {
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
    },
  ],
  invalid: [
    // Upstream lists this as valid with a todo noting it should not be: the
    // timer calls hasAssertions after Jest has checked the test.
    {
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
          line: 5,
          column: 1,
          endLine: 9,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
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
    {
      code: `it("it1", () => foo())`,
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: `it('resolves', () => expect(staged()).toBe(true));`,
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
    {
      code: `it('resolves', async () => expect(await staged()).toBe(true));`,
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 62,
        },
      ],
    },
    {
      code: `it("it1", () => {})`,
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it("it1", () => {expect.hasAssertions();})`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it("it1", () => {expect.assertions();})`,
            },
          ],
        },
      ],
    },
    {
      code: `it("it1", () => { foo()})`,
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it("it1", () => {expect.hasAssertions(); foo()})`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it("it1", () => {expect.assertions(); foo()})`,
            },
          ],
        },
      ],
    },
    {
      code: `it("it1", function() {
  someFunctionToDo();
  someFunctionToDo2();
});`,
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it("it1", function() {expect.hasAssertions();
  someFunctionToDo();
  someFunctionToDo2();
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it("it1", function() {expect.assertions();
  someFunctionToDo();
  someFunctionToDo2();
});`,
            },
          ],
        },
      ],
    },
    {
      code: `it("it1", function() {
  someFunctionToDo();
  someFunctionToDo2();
});

describe('some tests', () => {
  beforeEach(() => { expect.hasAssertions(); });
  it("it1", function() {
    someFunctionToDo();
    someFunctionToDo2();
  });
});`,
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it("it1", function() {expect.hasAssertions();
  someFunctionToDo();
  someFunctionToDo2();
});

describe('some tests', () => {
  beforeEach(() => { expect.hasAssertions(); });
  it("it1", function() {
    someFunctionToDo();
    someFunctionToDo2();
  });
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it("it1", function() {expect.assertions();
  someFunctionToDo();
  someFunctionToDo2();
});

describe('some tests', () => {
  beforeEach(() => { expect.hasAssertions(); });
  it("it1", function() {
    someFunctionToDo();
    someFunctionToDo2();
  });
});`,
            },
          ],
        },
      ],
    },
    {
      code: `it("it1", function() {
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
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it("it1", function() {expect.hasAssertions();
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
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it("it1", function() {expect.assertions();
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
            },
          ],
        },
      ],
    },
    {
      code: `describe('some tests', () => {
  beforeEach(() => { expect.hasAssertions(); });
  it("it1", function() {
    someFunctionToDo();
    someFunctionToDo2();
  });
});

it("it1", function() {
  someFunctionToDo();
  someFunctionToDo2();
});`,
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 9,
          column: 1,
          endLine: 12,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `describe('some tests', () => {
  beforeEach(() => { expect.hasAssertions(); });
  it("it1", function() {
    someFunctionToDo();
    someFunctionToDo2();
  });
});

it("it1", function() {expect.hasAssertions();
  someFunctionToDo();
  someFunctionToDo2();
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `describe('some tests', () => {
  beforeEach(() => { expect.hasAssertions(); });
  it("it1", function() {
    someFunctionToDo();
    someFunctionToDo2();
  });
});

it("it1", function() {expect.assertions();
  someFunctionToDo();
  someFunctionToDo2();
});`,
            },
          ],
        },
      ],
    },
    {
      code: `describe('some tests', () => {
  beforeEach(() => { expect.hasAssertions(); });
  it("it1", function() {
    someFunctionToDo();
    someFunctionToDo2();
  });
});

describe('more tests', () => {
  it("it1", function() {
    someFunctionToDo();
    someFunctionToDo2();
  });
});`,
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 10,
          column: 3,
          endLine: 13,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `describe('some tests', () => {
  beforeEach(() => { expect.hasAssertions(); });
  it("it1", function() {
    someFunctionToDo();
    someFunctionToDo2();
  });
});

describe('more tests', () => {
  it("it1", function() {expect.hasAssertions();
    someFunctionToDo();
    someFunctionToDo2();
  });
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `describe('some tests', () => {
  beforeEach(() => { expect.hasAssertions(); });
  it("it1", function() {
    someFunctionToDo();
    someFunctionToDo2();
  });
});

describe('more tests', () => {
  it("it1", function() {expect.assertions();
    someFunctionToDo();
    someFunctionToDo2();
  });
});`,
            },
          ],
        },
      ],
    },
    {
      code: `it("it1", function() {var a = 2;})`,
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 35,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it("it1", function() {expect.hasAssertions();var a = 2;})`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it("it1", function() {expect.assertions();var a = 2;})`,
            },
          ],
        },
      ],
    },
    {
      code: `it("it1", function() {expect.assertions();})`,
      errors: [
        {
          messageId: 'assertionsRequiresOneArgument',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      code: `it("it1", function() {expect.assertions(1,2);})`,
      errors: [
        {
          messageId: 'assertionsRequiresOneArgument',
          line: 1,
          column: 43,
          endLine: 1,
          endColumn: 44,
          suggestions: [
            {
              messageId: 'suggestRemovingExtraArguments',
              output: `it("it1", function() {expect.assertions(1,);})`,
            },
          ],
        },
      ],
    },
    {
      code: `it("it1", function() {expect.assertions(1,2,);})`,
      errors: [
        {
          messageId: 'assertionsRequiresOneArgument',
          line: 1,
          column: 43,
          endLine: 1,
          endColumn: 44,
          suggestions: [
            {
              messageId: 'suggestRemovingExtraArguments',
              output: `it("it1", function() {expect.assertions(1,);})`,
            },
          ],
        },
      ],
    },
    {
      code: `it("it1", function() {expect.assertions("1");})`,
      errors: [
        {
          messageId: 'assertionsRequiresNumberArgument',
          line: 1,
          column: 41,
          endLine: 1,
          endColumn: 44,
        },
      ],
    },
    {
      code: `beforeEach(() => { expect.hasAssertions("1") })`,
      errors: [
        {
          messageId: 'hasAssertionsTakesNoArguments',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 40,
          suggestions: [
            {
              messageId: 'suggestRemovingExtraArguments',
              output: `beforeEach(() => { expect.hasAssertions() })`,
            },
          ],
        },
      ],
    },
    {
      code: `beforeEach(() => expect.hasAssertions("1"))`,
      errors: [
        {
          messageId: 'hasAssertionsTakesNoArguments',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 38,
          suggestions: [
            {
              messageId: 'suggestRemovingExtraArguments',
              output: `beforeEach(() => expect.hasAssertions())`,
            },
          ],
        },
      ],
    },
    {
      code: `afterEach(() => { expect.hasAssertions("1") })`,
      errors: [
        {
          messageId: 'hasAssertionsTakesNoArguments',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 39,
          suggestions: [
            {
              messageId: 'suggestRemovingExtraArguments',
              output: `afterEach(() => { expect.hasAssertions() })`,
            },
          ],
        },
      ],
    },
    {
      code: `afterEach(() => expect.hasAssertions("1"))`,
      errors: [
        {
          messageId: 'hasAssertionsTakesNoArguments',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 37,
          suggestions: [
            {
              messageId: 'suggestRemovingExtraArguments',
              output: `afterEach(() => expect.hasAssertions())`,
            },
          ],
        },
      ],
    },
    {
      code: `it("it1", function() {expect.hasAssertions("1");})`,
      errors: [
        {
          messageId: 'hasAssertionsTakesNoArguments',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 43,
          suggestions: [
            {
              messageId: 'suggestRemovingExtraArguments',
              output: `it("it1", function() {expect.hasAssertions();})`,
            },
          ],
        },
      ],
    },
    {
      code: `it("it1", function() {expect.hasAssertions("1",);})`,
      errors: [
        {
          messageId: 'hasAssertionsTakesNoArguments',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 43,
          suggestions: [
            {
              messageId: 'suggestRemovingExtraArguments',
              output: `it("it1", function() {expect.hasAssertions();})`,
            },
          ],
        },
      ],
    },
    {
      code: `it("it1", function() {expect.hasAssertions("1", "2");})`,
      errors: [
        {
          messageId: 'hasAssertionsTakesNoArguments',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 43,
          suggestions: [
            {
              messageId: 'suggestRemovingExtraArguments',
              output: `it("it1", function() {expect.hasAssertions();})`,
            },
          ],
        },
      ],
    },
    {
      code: `it("it1", function() {
  expect.hasAssertions(() => {
    someFunctionToDo();
    someFunctionToDo2();
  });
});`,
      errors: [
        {
          messageId: 'hasAssertionsTakesNoArguments',
          line: 2,
          column: 10,
          endLine: 2,
          endColumn: 23,
          suggestions: [
            {
              messageId: 'suggestRemovingExtraArguments',
              output: `it("it1", function() {
  expect.hasAssertions();
});`,
            },
          ],
        },
      ],
    },
    {
      code: `it("it1", async function() {
  expect(someValue).toBe(true);
});`,
      options: [{ onlyFunctionsWithAsyncKeyword: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 1,
          column: 1,
          endLine: 3,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it("it1", async function() {expect.hasAssertions();
  expect(someValue).toBe(true);
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it("it1", async function() {expect.assertions();
  expect(someValue).toBe(true);
});`,
            },
          ],
        },
      ],
    },
    {
      code: `it("returns numbers that are greater than four", async () => {
  for(let thing in things) {
    expect(number).toBeGreaterThan(4);
  }
});`,
      options: [{ onlyFunctionsWithAsyncKeyword: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 1,
          column: 1,
          endLine: 5,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it("returns numbers that are greater than four", async () => {expect.hasAssertions();
  for(let thing in things) {
    expect(number).toBeGreaterThan(4);
  }
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it("returns numbers that are greater than four", async () => {expect.assertions();
  for(let thing in things) {
    expect(number).toBeGreaterThan(4);
  }
});`,
            },
          ],
        },
      ],
    },
    {
      code: `it("returns numbers that are greater than four", async () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});`,
      options: [{ onlyFunctionsWithAsyncKeyword: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 1,
          column: 1,
          endLine: 5,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it("returns numbers that are greater than four", async () => {expect.hasAssertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it("returns numbers that are greater than four", async () => {expect.assertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});`,
            },
          ],
        },
      ],
    },
    {
      code: `it("returns numbers that are greater than four", async () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});

it("returns numbers that are greater than five", () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(5);
  }
});`,
      options: [{ onlyFunctionsWithAsyncKeyword: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 1,
          column: 1,
          endLine: 5,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it("returns numbers that are greater than four", async () => {expect.hasAssertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});

it("returns numbers that are greater than five", () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(5);
  }
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it("returns numbers that are greater than four", async () => {expect.assertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});

it("returns numbers that are greater than five", () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(5);
  }
});`,
            },
          ],
        },
      ],
    },
    {
      code: `beforeAll(() => { expect.hasAssertions(); });

it("returns numbers that are greater than four", async () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});

it("returns numbers that are greater than five", () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(5);
  }
});`,
      options: [{ onlyFunctionsWithAsyncKeyword: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 3,
          column: 1,
          endLine: 7,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `beforeAll(() => { expect.hasAssertions(); });

it("returns numbers that are greater than four", async () => {expect.hasAssertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});

it("returns numbers that are greater than five", () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(5);
  }
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `beforeAll(() => { expect.hasAssertions(); });

it("returns numbers that are greater than four", async () => {expect.assertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});

it("returns numbers that are greater than five", () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(5);
  }
});`,
            },
          ],
        },
      ],
    },
    {
      code: `afterAll(() => { expect.hasAssertions(); });

it("returns numbers that are greater than four", async () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});

it("returns numbers that are greater than five", () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(5);
  }
});`,
      options: [{ onlyFunctionsWithAsyncKeyword: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 3,
          column: 1,
          endLine: 7,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `afterAll(() => { expect.hasAssertions(); });

it("returns numbers that are greater than four", async () => {expect.hasAssertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});

it("returns numbers that are greater than five", () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(5);
  }
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `afterAll(() => { expect.hasAssertions(); });

it("returns numbers that are greater than four", async () => {expect.assertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});

it("returns numbers that are greater than five", () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(5);
  }
});`,
            },
          ],
        },
      ],
    },
    {
      code: `it('only returns numbers that are greater than six', () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(6);
  }
});`,
      options: [{ onlyFunctionsWithExpectInLoop: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 1,
          column: 1,
          endLine: 5,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it('only returns numbers that are greater than six', () => {expect.hasAssertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(6);
  }
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it('only returns numbers that are greater than six', () => {expect.assertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(6);
  }
});`,
            },
          ],
        },
      ],
    },
    {
      code: `it('returns numbers that are greater than two', function () {
  const expectNumbersToBeGreaterThan = (numbers, value) => {
    for (let number of numbers) {
      expect(number).toBeGreaterThan(value);
    }
  };

  expectNumbersToBeGreaterThan(getNumbers(), 2);
});`,
      options: [{ onlyFunctionsWithExpectInLoop: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 1,
          column: 1,
          endLine: 9,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it('returns numbers that are greater than two', function () {expect.hasAssertions();
  const expectNumbersToBeGreaterThan = (numbers, value) => {
    for (let number of numbers) {
      expect(number).toBeGreaterThan(value);
    }
  };

  expectNumbersToBeGreaterThan(getNumbers(), 2);
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it('returns numbers that are greater than two', function () {expect.assertions();
  const expectNumbersToBeGreaterThan = (numbers, value) => {
    for (let number of numbers) {
      expect(number).toBeGreaterThan(value);
    }
  };

  expectNumbersToBeGreaterThan(getNumbers(), 2);
});`,
            },
          ],
        },
      ],
    },
    {
      code: `it("only returns numbers that are greater than seven", function () {
  const numbers = getNumbers();

  for (let i = 0; i < numbers.length; i++) {
    expect(numbers[i]).toBeGreaterThan(7);
  }
});`,
      options: [{ onlyFunctionsWithExpectInLoop: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 1,
          column: 1,
          endLine: 7,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it("only returns numbers that are greater than seven", function () {expect.hasAssertions();
  const numbers = getNumbers();

  for (let i = 0; i < numbers.length; i++) {
    expect(numbers[i]).toBeGreaterThan(7);
  }
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it("only returns numbers that are greater than seven", function () {expect.assertions();
  const numbers = getNumbers();

  for (let i = 0; i < numbers.length; i++) {
    expect(numbers[i]).toBeGreaterThan(7);
  }
});`,
            },
          ],
        },
      ],
    },
    {
      code: `it('has the number two', () => {
  expect(number).toBe(2);
});

it('only returns numbers that are less than twenty', () => {
  for (const number of getNumbers()) {
    expect(number).toBeLessThan(20);
  }
});`,
      options: [{ onlyFunctionsWithExpectInLoop: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 5,
          column: 1,
          endLine: 9,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it('has the number two', () => {
  expect(number).toBe(2);
});

it('only returns numbers that are less than twenty', () => {expect.hasAssertions();
  for (const number of getNumbers()) {
    expect(number).toBeLessThan(20);
  }
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it('has the number two', () => {
  expect(number).toBe(2);
});

it('only returns numbers that are less than twenty', () => {expect.assertions();
  for (const number of getNumbers()) {
    expect(number).toBeLessThan(20);
  }
});`,
            },
          ],
        },
      ],
    },
    {
      code: `it("is wrong");

it("is a test", () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});`,
      options: [{ onlyFunctionsWithExpectInLoop: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 3,
          column: 1,
          endLine: 7,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it("is wrong");

it("is a test", () => {expect.hasAssertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it("is wrong");

it("is a test", () => {expect.assertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});`,
            },
          ],
        },
      ],
    },
    {
      code: `it("is a number that is greater than four", () => {
  expect(number).toBeGreaterThan(4);
});

it("returns numbers that are greater than four", () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});

it("returns numbers that are greater than five", () => {
  expect(number).toBeGreaterThan(5);
});`,
      options: [{ onlyFunctionsWithExpectInLoop: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 5,
          column: 1,
          endLine: 9,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it("is a number that is greater than four", () => {
  expect(number).toBeGreaterThan(4);
});

it("returns numbers that are greater than four", () => {expect.hasAssertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});

it("returns numbers that are greater than five", () => {
  expect(number).toBeGreaterThan(5);
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it("is a number that is greater than four", () => {
  expect(number).toBeGreaterThan(4);
});

it("returns numbers that are greater than four", () => {expect.assertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});

it("returns numbers that are greater than five", () => {
  expect(number).toBeGreaterThan(5);
});`,
            },
          ],
        },
      ],
    },
    {
      code: `describe('my tests', () => {
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

it("returns numbers that are greater than five", () => {
  expect(number).toBeGreaterThan(5);
});`,
      options: [{ onlyFunctionsWithExpectInLoop: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 9,
          column: 3,
          endLine: 13,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `describe('my tests', () => {
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

it("returns numbers that are greater than five", () => {
  expect(number).toBeGreaterThan(5);
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `describe('my tests', () => {
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

it("returns numbers that are greater than five", () => {
  expect(number).toBeGreaterThan(5);
});`,
            },
          ],
        },
      ],
    },
    {
      code: `it.each([1, 2, 3])("returns numbers that are greater than four", () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});

it("is a number that is greater than four", () => {
  expect(number).toBeGreaterThan(4);
});`,
      options: [{ onlyFunctionsWithExpectInLoop: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 1,
          column: 1,
          endLine: 5,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it.each([1, 2, 3])("returns numbers that are greater than four", () => {expect.hasAssertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});

it("is a number that is greater than four", () => {
  expect(number).toBeGreaterThan(4);
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it.each([1, 2, 3])("returns numbers that are greater than four", () => {expect.assertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});

it("is a number that is greater than four", () => {
  expect(number).toBeGreaterThan(4);
});`,
            },
          ],
        },
      ],
    },
    {
      code: `it("returns numbers that are greater than four", () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});

it("is a number that is greater than four", () => {
  expect(number).toBeGreaterThan(4);
});`,
      options: [{ onlyFunctionsWithExpectInLoop: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 1,
          column: 1,
          endLine: 5,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it("returns numbers that are greater than four", () => {expect.hasAssertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});

it("is a number that is greater than four", () => {
  expect(number).toBeGreaterThan(4);
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it("returns numbers that are greater than four", () => {expect.assertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});

it("is a number that is greater than four", () => {
  expect(number).toBeGreaterThan(4);
});`,
            },
          ],
        },
      ],
    },
    {
      code: `it("returns numbers that are greater than four", () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});

it("is a number that is greater than four", () => {
  expect.hasAssertions();

  expect(number).toBeGreaterThan(4);
});`,
      options: [{ onlyFunctionsWithExpectInLoop: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 1,
          column: 1,
          endLine: 5,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it("returns numbers that are greater than four", () => {expect.hasAssertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});

it("is a number that is greater than four", () => {
  expect.hasAssertions();

  expect(number).toBeGreaterThan(4);
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it("returns numbers that are greater than four", () => {expect.assertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});

it("is a number that is greater than four", () => {
  expect.hasAssertions();

  expect(number).toBeGreaterThan(4);
});`,
            },
          ],
        },
      ],
    },
    {
      code: `it("it1", () => {
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
      options: [{ onlyFunctionsWithExpectInLoop: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 9,
          column: 1,
          endLine: 13,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it("it1", () => {
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
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it("it1", () => {
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
            },
          ],
        },
      ],
    },
    {
      code: `it("returns numbers that are greater than four", async () => {
  for (const number of await getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});

it("returns numbers that are greater than five", () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(5);
  }
});`,
      options: [{ onlyFunctionsWithExpectInLoop: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 1,
          column: 1,
          endLine: 5,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it("returns numbers that are greater than four", async () => {expect.hasAssertions();
  for (const number of await getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});

it("returns numbers that are greater than five", () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(5);
  }
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it("returns numbers that are greater than four", async () => {expect.assertions();
  for (const number of await getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});

it("returns numbers that are greater than five", () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(5);
  }
});`,
            },
          ],
        },
        {
          messageId: 'haveExpectAssertions',
          line: 7,
          column: 1,
          endLine: 11,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
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
    {
      code: `it("it1", async () => {
  expect.hasAssertions();

  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});

it("it1", () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});`,
      options: [{ onlyFunctionsWithExpectInLoop: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 9,
          column: 1,
          endLine: 13,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it("it1", async () => {
  expect.hasAssertions();

  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});

it("it1", () => {expect.hasAssertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it("it1", async () => {
  expect.hasAssertions();

  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});

it("it1", () => {expect.assertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});`,
            },
          ],
        },
      ],
    },
    {
      code: `describe('my tests', () => {
  beforeEach(() => { expect.hasAssertions(); });

  it("it1", async () => {
    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(4);
    }
  });
});

it("it1", () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});`,
      options: [{ onlyFunctionsWithExpectInLoop: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 11,
          column: 1,
          endLine: 15,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `describe('my tests', () => {
  beforeEach(() => { expect.hasAssertions(); });

  it("it1", async () => {
    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(4);
    }
  });
});

it("it1", () => {expect.hasAssertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `describe('my tests', () => {
  beforeEach(() => { expect.hasAssertions(); });

  it("it1", async () => {
    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(4);
    }
  });
});

it("it1", () => {expect.assertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});`,
            },
          ],
        },
      ],
    },
    {
      code: `describe('my tests', () => {
  afterEach(() => { expect.hasAssertions(); });

  it("it1", async () => {
    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(4);
    }
  });
});

it("it1", () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});`,
      options: [{ onlyFunctionsWithExpectInLoop: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 11,
          column: 1,
          endLine: 15,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `describe('my tests', () => {
  afterEach(() => { expect.hasAssertions(); });

  it("it1", async () => {
    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(4);
    }
  });
});

it("it1", () => {expect.hasAssertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `describe('my tests', () => {
  afterEach(() => { expect.hasAssertions(); });

  it("it1", async () => {
    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(4);
    }
  });
});

it("it1", () => {expect.assertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});`,
            },
          ],
        },
      ],
    },
    {
      code: `it.skip.each\`\`("it1", async () => {
  expect.hasAssertions();

  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});

it("it1", () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});`,
      options: [{ onlyFunctionsWithExpectInLoop: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 9,
          column: 1,
          endLine: 13,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it.skip.each\`\`("it1", async () => {
  expect.hasAssertions();

  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});

it("it1", () => {expect.hasAssertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it.skip.each\`\`("it1", async () => {
  expect.hasAssertions();

  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});

it("it1", () => {expect.assertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});`,
            },
          ],
        },
      ],
    },
    {
      code: `it("it1", async () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});

it("it1", () => {
  expect.hasAssertions();

  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});`,
      options: [{ onlyFunctionsWithExpectInLoop: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 1,
          column: 1,
          endLine: 5,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it("it1", async () => {expect.hasAssertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});

it("it1", () => {
  expect.hasAssertions();

  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it("it1", async () => {expect.assertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});

it("it1", () => {
  expect.hasAssertions();

  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});`,
            },
          ],
        },
      ],
    },
    {
      code: `describe('my tests', () => {
  it("it1", async () => {
    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(4);
    }
  });
});

it("it1", () => {
  expect.hasAssertions();

  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});`,
      options: [{ onlyFunctionsWithExpectInLoop: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 2,
          column: 3,
          endLine: 6,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `describe('my tests', () => {
  it("it1", async () => {expect.hasAssertions();
    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(4);
    }
  });
});

it("it1", () => {
  expect.hasAssertions();

  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `describe('my tests', () => {
  it("it1", async () => {expect.assertions();
    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(4);
    }
  });
});

it("it1", () => {
  expect.hasAssertions();

  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(4);
  }
});`,
            },
          ],
        },
      ],
    },
    {
      code: `it('sends the data as a string', () => {
  const stream = openStream();

  stream.on('data', data => {
    expect(data).toBe(expect.any(String));
  });
});`,
      options: [{ onlyFunctionsWithExpectInCallback: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 1,
          column: 1,
          endLine: 7,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it('sends the data as a string', () => {expect.hasAssertions();
  const stream = openStream();

  stream.on('data', data => {
    expect(data).toBe(expect.any(String));
  });
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it('sends the data as a string', () => {expect.assertions();
  const stream = openStream();

  stream.on('data', data => {
    expect(data).toBe(expect.any(String));
  });
});`,
            },
          ],
        },
      ],
    },
    {
      code: `it('responds ok', function () {
  client.get('/user', response => {
    expect(response.status).toBe(200);
  });
});`,
      options: [{ onlyFunctionsWithExpectInCallback: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 1,
          column: 1,
          endLine: 5,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it('responds ok', function () {expect.hasAssertions();
  client.get('/user', response => {
    expect(response.status).toBe(200);
  });
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it('responds ok', function () {expect.assertions();
  client.get('/user', response => {
    expect(response.status).toBe(200);
  });
});`,
            },
          ],
        },
      ],
    },
    {
      code: `it('responds ok', function () {
  client.get('/user', response => {
    expect.assertions(1);

    expect(response.status).toBe(200);
  });
});`,
      options: [{ onlyFunctionsWithExpectInCallback: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 1,
          column: 1,
          endLine: 7,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it('responds ok', function () {expect.hasAssertions();
  client.get('/user', response => {
    expect.assertions(1);

    expect(response.status).toBe(200);
  });
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it('responds ok', function () {expect.assertions();
  client.get('/user', response => {
    expect.assertions(1);

    expect(response.status).toBe(200);
  });
});`,
            },
          ],
        },
      ],
    },
    {
      code: `it('responds ok', function () {
  const expectOkResponse = response => {
    expect.assertions(1);

    expect(response.status).toBe(200);
  };

  client.get('/user', expectOkResponse);
});`,
      options: [{ onlyFunctionsWithExpectInCallback: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 1,
          column: 1,
          endLine: 9,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it('responds ok', function () {expect.hasAssertions();
  const expectOkResponse = response => {
    expect.assertions(1);

    expect(response.status).toBe(200);
  };

  client.get('/user', expectOkResponse);
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it('responds ok', function () {expect.assertions();
  const expectOkResponse = response => {
    expect.assertions(1);

    expect(response.status).toBe(200);
  };

  client.get('/user', expectOkResponse);
});`,
            },
          ],
        },
      ],
    },
    {
      code: `it('returns numbers that are greater than two', function () {
  const expectNumberToBeGreaterThan = (number, value) => {
    expect(number).toBeGreaterThan(value);
  };

  expectNumberToBeGreaterThan(1, 2);
});`,
      options: [{ onlyFunctionsWithExpectInCallback: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 1,
          column: 1,
          endLine: 7,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it('returns numbers that are greater than two', function () {expect.hasAssertions();
  const expectNumberToBeGreaterThan = (number, value) => {
    expect(number).toBeGreaterThan(value);
  };

  expectNumberToBeGreaterThan(1, 2);
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it('returns numbers that are greater than two', function () {expect.assertions();
  const expectNumberToBeGreaterThan = (number, value) => {
    expect(number).toBeGreaterThan(value);
  };

  expectNumberToBeGreaterThan(1, 2);
});`,
            },
          ],
        },
      ],
    },
    {
      code: `it('returns numbers that are greater than two', function () {
  const expectNumbersToBeGreaterThan = (numbers, value) => {
    for (let number of numbers) {
      expect(number).toBeGreaterThan(value);
    }
  };

  expectNumbersToBeGreaterThan(getNumbers(), 2);
});`,
      options: [{ onlyFunctionsWithExpectInCallback: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 1,
          column: 1,
          endLine: 9,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it('returns numbers that are greater than two', function () {expect.hasAssertions();
  const expectNumbersToBeGreaterThan = (numbers, value) => {
    for (let number of numbers) {
      expect(number).toBeGreaterThan(value);
    }
  };

  expectNumbersToBeGreaterThan(getNumbers(), 2);
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it('returns numbers that are greater than two', function () {expect.assertions();
  const expectNumbersToBeGreaterThan = (numbers, value) => {
    for (let number of numbers) {
      expect(number).toBeGreaterThan(value);
    }
  };

  expectNumbersToBeGreaterThan(getNumbers(), 2);
});`,
            },
          ],
        },
      ],
    },
    {
      code: `it('only returns numbers that are greater than six', () => {
  getNumbers().forEach(number => {
    expect(number).toBeGreaterThan(6);
  });
});`,
      options: [{ onlyFunctionsWithExpectInCallback: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 1,
          column: 1,
          endLine: 5,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it('only returns numbers that are greater than six', () => {expect.hasAssertions();
  getNumbers().forEach(number => {
    expect(number).toBeGreaterThan(6);
  });
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it('only returns numbers that are greater than six', () => {expect.assertions();
  getNumbers().forEach(number => {
    expect(number).toBeGreaterThan(6);
  });
});`,
            },
          ],
        },
      ],
    },
    {
      code: `it("is wrong");

it('responds ok', function () {
  const expectOkResponse = response => {
    expect.assertions(1);

    expect(response.status).toBe(200);
  };

  client.get('/user', expectOkResponse);
});`,
      options: [{ onlyFunctionsWithExpectInCallback: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 3,
          column: 1,
          endLine: 11,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it("is wrong");

it('responds ok', function () {expect.hasAssertions();
  const expectOkResponse = response => {
    expect.assertions(1);

    expect(response.status).toBe(200);
  };

  client.get('/user', expectOkResponse);
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it("is wrong");

it('responds ok', function () {expect.assertions();
  const expectOkResponse = response => {
    expect.assertions(1);

    expect(response.status).toBe(200);
  };

  client.get('/user', expectOkResponse);
});`,
            },
          ],
        },
      ],
    },
    {
      code: `it("is a number that is greater than four", () => {
  expect(number).toBeGreaterThan(4);
});

it('responds ok', function () {
  const expectOkResponse = response => {
    expect(response.status).toBe(200);
  };

  client.get('/user', expectOkResponse);
});

it("returns numbers that are greater than five", () => {
  expect(number).toBeGreaterThan(5);
});`,
      options: [{ onlyFunctionsWithExpectInCallback: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 5,
          column: 1,
          endLine: 11,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it("is a number that is greater than four", () => {
  expect(number).toBeGreaterThan(4);
});

it('responds ok', function () {expect.hasAssertions();
  const expectOkResponse = response => {
    expect(response.status).toBe(200);
  };

  client.get('/user', expectOkResponse);
});

it("returns numbers that are greater than five", () => {
  expect(number).toBeGreaterThan(5);
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it("is a number that is greater than four", () => {
  expect(number).toBeGreaterThan(4);
});

it('responds ok', function () {expect.assertions();
  const expectOkResponse = response => {
    expect(response.status).toBe(200);
  };

  client.get('/user', expectOkResponse);
});

it("returns numbers that are greater than five", () => {
  expect(number).toBeGreaterThan(5);
});`,
            },
          ],
        },
      ],
    },
    {
      code: `it("is a number that is greater than four", () => {
  expect(number).toBeGreaterThan(4);
});

it("returns numbers that are greater than four", () => {
  getNumbers().map(number => {
    expect(number).toBeGreaterThan(0);
  });
});

it("returns numbers that are greater than five", () => {
  expect(number).toBeGreaterThan(5);
});`,
      options: [{ onlyFunctionsWithExpectInCallback: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 5,
          column: 1,
          endLine: 9,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it("is a number that is greater than four", () => {
  expect(number).toBeGreaterThan(4);
});

it("returns numbers that are greater than four", () => {expect.hasAssertions();
  getNumbers().map(number => {
    expect(number).toBeGreaterThan(0);
  });
});

it("returns numbers that are greater than five", () => {
  expect(number).toBeGreaterThan(5);
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it("is a number that is greater than four", () => {
  expect(number).toBeGreaterThan(4);
});

it("returns numbers that are greater than four", () => {expect.assertions();
  getNumbers().map(number => {
    expect(number).toBeGreaterThan(0);
  });
});

it("returns numbers that are greater than five", () => {
  expect(number).toBeGreaterThan(5);
});`,
            },
          ],
        },
      ],
    },
    {
      code: `it.each([1, 2, 3])("returns ok", id => {
  client.get(\`/users/\${id}\`, response => {
    expect(response.status).toBe(200);
  });
});

it("is a number that is greater than four", () => {
  expect(number).toBeGreaterThan(4);
});`,
      options: [{ onlyFunctionsWithExpectInCallback: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 1,
          column: 1,
          endLine: 5,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it.each([1, 2, 3])("returns ok", id => {expect.hasAssertions();
  client.get(\`/users/\${id}\`, response => {
    expect(response.status).toBe(200);
  });
});

it("is a number that is greater than four", () => {
  expect(number).toBeGreaterThan(4);
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it.each([1, 2, 3])("returns ok", id => {expect.assertions();
  client.get(\`/users/\${id}\`, response => {
    expect(response.status).toBe(200);
  });
});

it("is a number that is greater than four", () => {
  expect(number).toBeGreaterThan(4);
});`,
            },
          ],
        },
      ],
    },
    {
      code: `it('responds ok', function () {
  client.get('/user', response => {
    expect(response.status).toBe(200);
  });
});

it("is a number that is greater than four", () => {
  expect(number).toBeGreaterThan(4);
});`,
      options: [{ onlyFunctionsWithExpectInCallback: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 1,
          column: 1,
          endLine: 5,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it('responds ok', function () {expect.hasAssertions();
  client.get('/user', response => {
    expect(response.status).toBe(200);
  });
});

it("is a number that is greater than four", () => {
  expect(number).toBeGreaterThan(4);
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it('responds ok', function () {expect.assertions();
  client.get('/user', response => {
    expect(response.status).toBe(200);
  });
});

it("is a number that is greater than four", () => {
  expect(number).toBeGreaterThan(4);
});`,
            },
          ],
        },
      ],
    },
    {
      code: `it('responds ok', function () {
  client.get('/user', response => {
    expect(response.status).toBe(200);
  });
});

it("is a number that is greater than four", () => {
  expect.hasAssertions();

  expect(number).toBeGreaterThan(4);
});`,
      options: [{ onlyFunctionsWithExpectInCallback: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 1,
          column: 1,
          endLine: 5,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it('responds ok', function () {expect.hasAssertions();
  client.get('/user', response => {
    expect(response.status).toBe(200);
  });
});

it("is a number that is greater than four", () => {
  expect.hasAssertions();

  expect(number).toBeGreaterThan(4);
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it('responds ok', function () {expect.assertions();
  client.get('/user', response => {
    expect(response.status).toBe(200);
  });
});

it("is a number that is greater than four", () => {
  expect.hasAssertions();

  expect(number).toBeGreaterThan(4);
});`,
            },
          ],
        },
      ],
    },
    {
      code: `it("it1", () => {
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
      options: [{ onlyFunctionsWithExpectInCallback: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 9,
          column: 1,
          endLine: 13,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it("it1", () => {
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
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it("it1", () => {
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
            },
          ],
        },
      ],
    },
    {
      code: `it('responds ok', function () {
  expect.hasAssertions();

  client.get('/user', response => {
    expect(response.status).toBe(200);
  });
});

it('responds not found', function () {
  client.get('/user', response => {
    expect(response.status).toBe(404);
  });
});`,
      options: [{ onlyFunctionsWithExpectInCallback: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 9,
          column: 1,
          endLine: 13,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it('responds ok', function () {
  expect.hasAssertions();

  client.get('/user', response => {
    expect(response.status).toBe(200);
  });
});

it('responds not found', function () {expect.hasAssertions();
  client.get('/user', response => {
    expect(response.status).toBe(404);
  });
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it('responds ok', function () {
  expect.hasAssertions();

  client.get('/user', response => {
    expect(response.status).toBe(200);
  });
});

it('responds not found', function () {expect.assertions();
  client.get('/user', response => {
    expect(response.status).toBe(404);
  });
});`,
            },
          ],
        },
      ],
    },
    {
      code: `it.skip.each\`\`("it1", async () => {
  expect.hasAssertions();

  client.get('/user', response => {
    expect(response.status).toBe(200);
  });
});

it("responds ok", () => {
  client.get('/user', response => {
    expect(response.status).toBe(200);
  });
});`,
      options: [{ onlyFunctionsWithExpectInCallback: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 9,
          column: 1,
          endLine: 13,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it.skip.each\`\`("it1", async () => {
  expect.hasAssertions();

  client.get('/user', response => {
    expect(response.status).toBe(200);
  });
});

it("responds ok", () => {expect.hasAssertions();
  client.get('/user', response => {
    expect(response.status).toBe(200);
  });
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it.skip.each\`\`("it1", async () => {
  expect.hasAssertions();

  client.get('/user', response => {
    expect(response.status).toBe(200);
  });
});

it("responds ok", () => {expect.assertions();
  client.get('/user', response => {
    expect(response.status).toBe(200);
  });
});`,
            },
          ],
        },
      ],
    },
    {
      code: `it("returns numbers that are greater than four", function(expect) {
  expect.assertions(2);

  for(let thing in things) {
    expect(number).toBeGreaterThan(4);
  }
});`,
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 1,
          column: 1,
          endLine: 7,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it("returns numbers that are greater than four", function(expect) {expect.hasAssertions();
  expect.assertions(2);

  for(let thing in things) {
    expect(number).toBeGreaterThan(4);
  }
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
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
    {
      code: `it('only returns numbers that are greater than zero', () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});

it("is zero", () => {
  expect.hasAssertions();

  expect(0).toBe(0);
});`,
      options: [{ onlyFunctionsWithExpectInLoop: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 1,
          column: 1,
          endLine: 5,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it('only returns numbers that are greater than zero', () => {expect.hasAssertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});

it("is zero", () => {
  expect.hasAssertions();

  expect(0).toBe(0);
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it('only returns numbers that are greater than zero', () => {expect.assertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});

it("is zero", () => {
  expect.hasAssertions();

  expect(0).toBe(0);
});`,
            },
          ],
        },
      ],
    },
    {
      code: `it('only returns numbers that are greater than zero', () => {
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
      options: [{ onlyFunctionsWithExpectInLoop: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 9,
          column: 1,
          endLine: 13,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it('only returns numbers that are greater than zero', () => {
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
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it('only returns numbers that are greater than zero', () => {
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
            },
          ],
        },
      ],
    },
    {
      code: `it("to be true", async function() {
  expect(someValue).toBe(true);
});`,
      options: [
        {
          onlyFunctionsWithAsyncKeyword: true,
          onlyFunctionsWithExpectInLoop: true,
        },
      ],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 1,
          column: 1,
          endLine: 3,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it("to be true", async function() {expect.hasAssertions();
  expect(someValue).toBe(true);
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it("to be true", async function() {expect.assertions();
  expect(someValue).toBe(true);
});`,
            },
          ],
        },
      ],
    },
    {
      code: `it('only returns numbers that are greater than zero', async () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});`,
      options: [
        {
          onlyFunctionsWithAsyncKeyword: true,
          onlyFunctionsWithExpectInLoop: true,
        },
      ],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 1,
          column: 1,
          endLine: 5,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it('only returns numbers that are greater than zero', async () => {expect.hasAssertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it('only returns numbers that are greater than zero', async () => {expect.assertions();
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(0);
  }
});`,
            },
          ],
        },
      ],
    },
    {
      code: `test.each()("is not fine", () => {
  expect(someValue).toBe(true);
});`,
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 1,
          column: 1,
          endLine: 3,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `test.each()("is not fine", () => {expect.hasAssertions();
  expect(someValue).toBe(true);
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `test.each()("is not fine", () => {expect.assertions();
  expect(someValue).toBe(true);
});`,
            },
          ],
        },
      ],
    },
    {
      code: `describe.each()('something', () => {
  it("is not fine", () => {
    expect(someValue).toBe(true);
  });
});`,
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 2,
          column: 3,
          endLine: 4,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `describe.each()('something', () => {
  it("is not fine", () => {expect.hasAssertions();
    expect(someValue).toBe(true);
  });
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `describe.each()('something', () => {
  it("is not fine", () => {expect.assertions();
    expect(someValue).toBe(true);
  });
});`,
            },
          ],
        },
      ],
    },
    {
      code: `describe.each()('something', () => {
  test.each()("is not fine", () => {
    expect(someValue).toBe(true);
  });
});`,
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 2,
          column: 3,
          endLine: 4,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `describe.each()('something', () => {
  test.each()("is not fine", () => {expect.hasAssertions();
    expect(someValue).toBe(true);
  });
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `describe.each()('something', () => {
  test.each()("is not fine", () => {expect.assertions();
    expect(someValue).toBe(true);
  });
});`,
            },
          ],
        },
      ],
    },
    {
      code: `test.each()("is not fine", async () => {
  expect(someValue).toBe(true);
});`,
      options: [{ onlyFunctionsWithAsyncKeyword: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 1,
          column: 1,
          endLine: 3,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `test.each()("is not fine", async () => {expect.hasAssertions();
  expect(someValue).toBe(true);
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `test.each()("is not fine", async () => {expect.assertions();
  expect(someValue).toBe(true);
});`,
            },
          ],
        },
      ],
    },
    {
      code: `it.each()("is not fine", async () => {
  expect(someValue).toBe(true);
});`,
      options: [{ onlyFunctionsWithAsyncKeyword: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 1,
          column: 1,
          endLine: 3,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `it.each()("is not fine", async () => {expect.hasAssertions();
  expect(someValue).toBe(true);
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `it.each()("is not fine", async () => {expect.assertions();
  expect(someValue).toBe(true);
});`,
            },
          ],
        },
      ],
    },
    {
      code: `describe.each()('something', () => {
  test.each()("is not fine", async () => {
    expect(someValue).toBe(true);
  });
});`,
      options: [{ onlyFunctionsWithAsyncKeyword: true }],
      errors: [
        {
          messageId: 'haveExpectAssertions',
          line: 2,
          column: 3,
          endLine: 4,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestAddingHasAssertions',
              output: `describe.each()('something', () => {
  test.each()("is not fine", async () => {expect.hasAssertions();
    expect(someValue).toBe(true);
  });
});`,
            },
            {
              messageId: 'suggestAddingAssertions',
              output: `describe.each()('something', () => {
  test.each()("is not fine", async () => {expect.assertions();
    expect(someValue).toBe(true);
  });
});`,
            },
          ],
        },
      ],
    },
  ],
});
