// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

// Shared source templates; reporting ranges remain explicit in each matrix.
const testSetup = `  describe('some condition', () => {
    test('foo', () => {
      const xyz = 987;

      // Comment
      expect(xyz).toEqual(987);
      expect(1)
        .toEqual(1);
      expect(true).toEqual(true);
    });
  });
});

test('awaited expect', async () => {
  const abc = 123;
  const hasAPromise = () => Promise.resolve('foo');

  await expect(hasAPromise()).resolves.toEqual('foo');
  expect(abc).toEqual(123);

  const efg = 456;

  expect(123).toEqual(abc);
  await expect(hasAPromise()).resolves.toEqual('foo');

  const hij = 789;

  await expect(hasAPromise()).resolves.toEqual('foo');
  await expect(hasAPromise()).resolves.toEqual('foo');

  const somethingElseAsync = () => Promise.resolve('bar');
  await somethingElseAsync();

  await expect(hasAPromise()).resolves.toEqual('foo');
});
`;
const testSetup2 = `
foo();
bar();

const someText = 'abc';
const someObject = {
  one: 1,
  two: 2,
};

test('thing one', () => {
  let abc = 123;

  expect(abc).toEqual(123);
  expect(123).toEqual(abc); // Line comment

  abc = 456;

  expect(abc).toEqual(456);
});

test('thing one', () => {
  const abc = 123;

  expect(abc).toEqual(123);

  const xyz = 987;

  expect(123).toEqual(abc); // Line comment
});

describe('someText', () => {
`;

new RuleTester(fixtureConfig()).run('padding-around-expect-groups', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: `const value = load();

expect(value).toBeDefined();`,
    },
    {
      name: 'upstream padding-around-expect-groups / valid 0',
      code: testSetup2 + testSetup,
      languageOptions: {
        ecmaVersion: 2017,
      },
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: `const value = load();
expect(value).toBeDefined();`,
      errors: [
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 29,
        },
      ],
      output: `const value = load();

expect(value).toBeDefined();`,
    },
    {
      name: 'upstream padding-around-expect-groups / invalid 0',
      code: `
foo();
bar();

const someText = 'abc';
const someObject = {
  one: 1,
  two: 2,
};

test('thing one', () => {
  let abc = 123;
  expect(abc).toEqual(123);
  expect(123).toEqual(abc); // Line comment
  abc = 456;
  expect(abc).toEqual(456);
});

test('thing one', () => {
  const abc = 123;
  expect(abc).toEqual(123);

  const xyz = 987;
  expect(123).toEqual(abc); // Line comment
});

describe('someText', () => {
  describe('some condition', () => {
    test('foo', () => {
      const xyz = 987;
      // Comment
      expect(xyz).toEqual(987);
      expect(1)
        .toEqual(1);
      expect(true).toEqual(true);
    });
  });
});

test('awaited expect', async () => {
  const abc = 123;
  const hasAPromise = () => Promise.resolve('foo');
  await expect(hasAPromise()).resolves.toEqual('foo');
  expect(abc).toEqual(123);

  const efg = 456;
  expect(123).toEqual(abc);
  await expect(hasAPromise()).resolves.toEqual('foo');

  const hij = 789;
  await expect(hasAPromise()).resolves.toEqual('foo');
  await expect(hasAPromise()).resolves.toEqual('foo');

  const somethingElseAsync = () => Promise.resolve('bar');
  await somethingElseAsync();
  await expect(hasAPromise()).resolves.toEqual('foo');
});
`,
      filename: 'src/component.test.jsx',
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 13,
          column: 3,
          endLine: 13,
          endColumn: 28,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 15,
          column: 3,
          endLine: 15,
          endColumn: 13,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 16,
          column: 3,
          endLine: 16,
          endColumn: 28,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 21,
          column: 3,
          endLine: 21,
          endColumn: 28,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 24,
          column: 3,
          endLine: 24,
          endColumn: 28,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 32,
          column: 7,
          endLine: 32,
          endColumn: 32,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 43,
          column: 3,
          endLine: 43,
          endColumn: 55,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 47,
          column: 3,
          endLine: 47,
          endColumn: 28,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 51,
          column: 3,
          endLine: 51,
          endColumn: 55,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 56,
          column: 3,
          endLine: 56,
          endColumn: 55,
        },
      ],
      output: testSetup2 + testSetup,
    },
    {
      name: 'upstream padding-around-expect-groups / invalid 1',
      code: `
foo();
bar();

const someText = 'abc';
const someObject = {
  one: 1,
  two: 2,
};

test('thing one', () => {
  let abc = 123;
  expect(abc).toEqual(123);
  expect(123).toEqual(abc); // Line comment
  abc = 456;
  expect(abc).toEqual(456);
});

test('thing one', () => {
  const abc = 123;
  expect(abc).toEqual(123);

  const xyz = 987;
  expect(123).toEqual(abc); // Line comment
});

describe('someText', () => {
  describe('some condition', () => {
    test('foo', () => {
      const xyz = 987;
      // Comment
      expect(xyz).toEqual(987);
      expect(1)
        .toEqual(1);
      expect(true).toEqual(true);
    });
  });
});

test('awaited expect', async () => {
  const abc = 123;
  const hasAPromise = () => Promise.resolve('foo');
  await expect(hasAPromise()).resolves.toEqual('foo');
  expect(abc).toEqual(123);

  const efg = 456;
  expect(123).toEqual(abc);
  await expect(hasAPromise()).resolves.toEqual('foo');

  const hij = 789;
  await expect(hasAPromise()).resolves.toEqual('foo');
  await expect(hasAPromise()).resolves.toEqual('foo');

  const somethingElseAsync = () => Promise.resolve('bar');
  await somethingElseAsync();
  await expect(hasAPromise()).resolves.toEqual('foo');
});
`,
      filename: 'src/component.test.js',
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 13,
          column: 3,
          endLine: 13,
          endColumn: 28,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 15,
          column: 3,
          endLine: 15,
          endColumn: 13,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 16,
          column: 3,
          endLine: 16,
          endColumn: 28,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 21,
          column: 3,
          endLine: 21,
          endColumn: 28,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 24,
          column: 3,
          endLine: 24,
          endColumn: 28,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 32,
          column: 7,
          endLine: 32,
          endColumn: 32,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 43,
          column: 3,
          endLine: 43,
          endColumn: 55,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 47,
          column: 3,
          endLine: 47,
          endColumn: 28,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 51,
          column: 3,
          endLine: 51,
          endColumn: 55,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 56,
          column: 3,
          endLine: 56,
          endColumn: 55,
        },
      ],
      output: testSetup2 + testSetup,
    },
  ],
});
