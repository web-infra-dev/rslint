import { RuleTester } from '../rule-tester';
const ruleTester = new RuleTester();
ruleTester.run('prefer-snapshot-hint', {} as never, {
  valid: [
    {
      code: 'expect(something).toStrictEqual(somethingElse);',
      options: ['always'],
    },
    {
      code: "a().toEqual('b')",
      options: ['always'],
    },
    {
      code: 'expect(a);',
      options: ['always'],
    },
    {
      code: 'expect(1).toMatchSnapshot({}, "my snapshot");',
      options: ['always'],
    },
    {
      code: 'expect(1).toMatchSnapshot("my snapshot");',
      options: ['always'],
    },
    {
      code: 'expect(1).toMatchSnapshot(`my snapshot`);',
      options: ['always'],
    },
    {
      code: 'const x = {};\nexpect(1).toMatchSnapshot(x, "my snapshot");',
      options: ['always'],
    },
    {
      code: 'expect(1).toThrowErrorMatchingSnapshot("my snapshot");',
      options: ['always'],
    },
    {
      code: 'expect(1).toMatchInlineSnapshot();',
      options: ['always'],
    },
    {
      code: 'expect(1).toThrowErrorMatchingInlineSnapshot();',
      options: ['always'],
    },
  ],
  invalid: [
    {
      code: 'expect(1).toMatchSnapshot();',
      options: ['always'],
      errors: [
        {
          messageId: 'missingHint',
          column: 11,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(1).toMatchSnapshot({});',
      options: ['always'],
      errors: [
        {
          messageId: 'missingHint',
          column: 11,
          line: 1,
        },
      ],
    },
    {
      code: 'const x = "we can\'t know if this is a string or not";\nexpect(1).toMatchSnapshot(x);',
      options: ['always'],
      errors: [
        {
          messageId: 'missingHint',
          column: 11,
          line: 2,
        },
      ],
    },
    {
      code: 'expect(1).toThrowErrorMatchingSnapshot();',
      options: ['always'],
      errors: [
        {
          messageId: 'missingHint',
          column: 11,
          line: 1,
        },
      ],
    },
    {
      code: "it('is true', () => {\n  expect(1).toMatchSnapshot();\n});",
      options: ['always'],
      errors: [
        {
          messageId: 'missingHint',
          column: 13,
          line: 2,
        },
      ],
    },
    {
      code: "it('is true', () => {\n  expect(1).toMatchSnapshot();\n  expect(2).toMatchSnapshot();\n});",
      options: ['always'],
      errors: [
        {
          messageId: 'missingHint',
          column: 13,
          line: 2,
        },
        {
          messageId: 'missingHint',
          column: 13,
          line: 3,
        },
      ],
    },
    {
      code: 'it(\'is true\', () => {\n  expect(1).toMatchSnapshot();\n  expect(2).toThrowErrorMatchingSnapshot("my error");\n});',
      options: ['always'],
      errors: [
        {
          messageId: 'missingHint',
          column: 13,
          line: 2,
        },
      ],
    },
    {
      code: 'const expectSnapshot = value => {\n  expect(value).toMatchSnapshot();\n};',
      options: ['always'],
      errors: [
        {
          messageId: 'missingHint',
          column: 17,
          line: 2,
        },
      ],
    },
    {
      code: 'const expectSnapshot = value => {\n  expect(value).toThrowErrorMatchingSnapshot();\n};',
      options: ['always'],
      errors: [
        {
          messageId: 'missingHint',
          column: 17,
          line: 2,
        },
      ],
    },
    {
      code: "it('is true', () => {\n  { expect(1).toMatchSnapshot(); }\n});",
      options: ['always'],
      errors: [
        {
          messageId: 'missingHint',
          column: 15,
          line: 2,
        },
      ],
    },
    {
      code: 'const x = "snapshot";\nexpect(1).toMatchSnapshot(`my ${x}`);',
      options: ['always'],
      errors: [
        {
          messageId: 'missingHint',
          column: 11,
          line: 2,
        },
      ],
    },
  ],
});
ruleTester.run('prefer-snapshot-hint', {} as never, {
  valid: [
    {
      code: 'expect(something).toStrictEqual(somethingElse);',
      options: ['multi'],
    },
    {
      code: "a().toEqual('b')",
      options: ['multi'],
    },
    {
      code: 'expect(a);',
      options: ['multi'],
    },
    {
      code: 'expect(1).toMatchSnapshot({}, "my snapshot");',
      options: ['multi'],
    },
    {
      code: 'expect(1).toThrowErrorMatchingSnapshot("my snapshot");',
      options: ['multi'],
    },
    {
      code: 'expect(1).toMatchSnapshot({});',
      options: ['multi'],
    },
    {
      code: 'expect(1).toThrowErrorMatchingSnapshot();',
      options: ['multi'],
    },
    {
      code: "it('is true', () => {\n  expect(1).toMatchSnapshot();\n});",
      options: ['multi'],
    },
    {
      code: "it('is true', () => {\n  expect(1).toMatchSnapshot(undefined, 'my first snapshot');\n});",
      options: ['multi'],
    },
    {
      code: "describe('my tests', () => {\n  it('is true', () => {\n    expect(1).toMatchSnapshot('this is a hint, all by itself');\n  });\n\n  it('is false', () => {\n    expect(2).toMatchSnapshot('this is a hint');\n    expect(2).toMatchSnapshot('and so is this');\n  });\n});",
      options: ['multi'],
    },
    {
      code: "it('is true', () => {\n  expect(1).toMatchSnapshot();\n});\n\nit('is false', () => {\n  expect(2).toMatchSnapshot('this is a hint');\n  expect(2).toMatchSnapshot('and so is this');\n});",
      options: ['multi'],
    },
    {
      code: "it('is true', () => {\n  expect(1).toMatchSnapshot();\n});\n\nit('is false', () => {\n  expect(2).toThrowErrorMatchingSnapshot();\n});",
      options: ['multi'],
    },
    {
      code: "it('is true', () => {\n  expect(1).toStrictEqual(1);\n  expect(1).toStrictEqual(2);\n  expect(1).toMatchSnapshot();\n});\n\nit('is false', () => {\n  expect(1).toStrictEqual(1);\n  expect(1).toStrictEqual(2);\n  expect(2).toThrowErrorMatchingSnapshot();\n});",
      options: ['multi'],
    },
    {
      code: "it('is true', () => {\n  expect(1).toMatchInlineSnapshot();\n});\n\nit('is false', () => {\n  expect(1).toMatchInlineSnapshot();\n  expect(1).toMatchInlineSnapshot();\n  expect(1).toThrowErrorMatchingInlineSnapshot();\n});",
      options: ['multi'],
    },
    {
      code: "it('is true', () => {\n  expect(1).toMatchSnapshot();\n});\n\nit('is false', () => {\n  expect(1).toMatchSnapshot();\n});",
      options: ['multi'],
    },
    {
      code: "import { it as itIs } from '@jest/globals';\n\nit('is true', () => {\n  expect(1).toMatchSnapshot();\n});\n\nitIs('false', () => {\n  expect(1).toMatchSnapshot();\n});",
      options: ['multi'],
    },
    {
      code: "const myReusableTestBody = (value, snapshotHint) => {\n  const innerFn = anotherValue => {\n    expect(anotherValue).toMatchSnapshot();\n\n    expect(value).toBe(1);\n  };\n\n  expect(value).toBe(1);\n};\n\nit('my test', () => {\n  expect(1).toMatchSnapshot();\n});",
      options: ['multi'],
    },
    {
      code: "const myReusableTestBody = (value, snapshotHint) => {\n  const innerFn = anotherValue => {\n    expect(value).toBe(1);\n  };\n\n  expect(value).toBe(1);\n  expect(anotherValue).toMatchSnapshot();\n};\n\nit('my test', () => {\n  expect(1).toMatchSnapshot();\n});",
      options: ['multi'],
    },
    {
      code: 'const myReusableTestBody = (value, snapshotHint) => {\n  const innerFn = anotherValue => {\n    expect(anotherValue).toMatchSnapshot();\n\n    expect(value).toBe(1);\n  };\n\n  expect(value).toBe(1);\n};\n\nexpect(1).toMatchSnapshot();',
      options: ['multi'],
    },
  ],
  invalid: [
    {
      code: "it('is true', () => {\n  expect(1).toMatchSnapshot();\n  expect(2).toMatchSnapshot();\n});",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          column: 13,
          line: 2,
        },
        {
          messageId: 'missingHint',
          column: 13,
          line: 3,
        },
      ],
    },
    {
      code: "it('is true', () => {\n  expect(1).toMatchSnapshot();\n  expect(2).toThrowErrorMatchingSnapshot();\n});",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          column: 13,
          line: 2,
        },
        {
          messageId: 'missingHint',
          column: 13,
          line: 3,
        },
      ],
    },
    {
      code: "it('is true', () => {\n  expect(1).toThrowErrorMatchingSnapshot();\n  expect(2).toMatchSnapshot();\n});",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          column: 13,
          line: 2,
        },
        {
          messageId: 'missingHint',
          column: 13,
          line: 3,
        },
      ],
    },
    {
      code: "it('is true', () => {\n  expect(1).toMatchSnapshot({});\n  expect(2).toMatchSnapshot({});\n});",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          column: 13,
          line: 2,
        },
        {
          messageId: 'missingHint',
          column: 13,
          line: 3,
        },
      ],
    },
    {
      code: "it('is true', () => {\n  expect(1).toMatchSnapshot({});\n  {\n    expect(2).toMatchSnapshot({});\n  }\n});",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          column: 13,
          line: 2,
        },
        {
          messageId: 'missingHint',
          column: 15,
          line: 4,
        },
      ],
    },
    {
      code: "it('is true', () => {\n  { expect(1).toMatchSnapshot(); }\n  { expect(2).toMatchSnapshot(); }\n});",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          column: 15,
          line: 2,
        },
        {
          messageId: 'missingHint',
          column: 15,
          line: 3,
        },
      ],
    },
    {
      code: "it('is true', () => {\n  expect(1).toMatchSnapshot();\n  expect(2).toMatchSnapshot(undefined, 'my second snapshot');\n});",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          column: 13,
          line: 2,
        },
      ],
    },
    {
      code: "it('is true', () => {\n  expect(1).toMatchSnapshot({});\n  expect(2).toMatchSnapshot(undefined, 'my second snapshot');\n});",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          column: 13,
          line: 2,
        },
      ],
    },
    {
      code: "it('is true', () => {\n  expect(1).toMatchSnapshot({}, 'my first snapshot');\n  expect(2).toMatchSnapshot(undefined);\n});",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          column: 13,
          line: 3,
        },
      ],
    },
    {
      code: "it('is true', () => {\n  expect(1).toMatchSnapshot({}, 'my first snapshot');\n  expect(2).toMatchSnapshot(undefined);\n  expect(2).toMatchSnapshot();\n});",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          column: 13,
          line: 3,
        },
        {
          messageId: 'missingHint',
          column: 13,
          line: 4,
        },
      ],
    },
    {
      code: "it('is true', () => {\n  expect(2).toMatchSnapshot();\n  expect(1).toMatchSnapshot({}, 'my second snapshot');\n  expect(2).toMatchSnapshot();\n});",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          column: 13,
          line: 2,
        },
        {
          messageId: 'missingHint',
          column: 13,
          line: 4,
        },
      ],
    },
    {
      code: "it('is true', () => {\n  expect(2).toMatchSnapshot(undefined);\n  expect(2).toMatchSnapshot();\n  expect(1).toMatchSnapshot(null, 'my third snapshot');\n});",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          column: 13,
          line: 2,
        },
        {
          messageId: 'missingHint',
          column: 13,
          line: 3,
        },
      ],
    },
    {
      code: "describe('my tests', () => {\n  it('is true', () => {\n    expect(1).toMatchSnapshot();\n  });\n\n  it('is false', () => {\n    expect(2).toMatchSnapshot();\n    expect(2).toMatchSnapshot();\n  });\n});",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          column: 15,
          line: 7,
        },
        {
          messageId: 'missingHint',
          column: 15,
          line: 8,
        },
      ],
    },
    {
      code: "describe('my tests', () => {\n  it('is true', () => {\n    expect(1).toMatchSnapshot();\n  });\n\n  it('is false', () => {\n    expect(2).toMatchSnapshot();\n    expect(2).toMatchSnapshot('hello world');\n  });\n});",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          column: 15,
          line: 7,
        },
      ],
    },
    {
      code: "describe('my tests', () => {\n  describe('more tests', () => {\n    it('is true', () => {\n      expect(1).toMatchSnapshot();\n    });\n  });\n\n  it('is false', () => {\n    expect(2).toMatchSnapshot();\n    expect(2).toMatchSnapshot('hello world');\n  });\n});",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          column: 15,
          line: 9,
        },
      ],
    },
    {
      code: "describe('my tests', () => {\n  it('is true', () => {\n    expect(1).toMatchSnapshot();\n  });\n\n  describe('more tests', () => {\n    it('is false', () => {\n      expect(2).toMatchSnapshot();\n      expect(2).toMatchSnapshot('hello world');\n    });\n  });\n});",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          column: 17,
          line: 8,
        },
      ],
    },
    {
      code: "import { describe as context, it as itIs } from '@jest/globals';\n\ndescribe('my tests', () => {\n  it('is true', () => {\n    expect(1).toMatchSnapshot();\n  });\n\n  context('more tests', () => {\n    itIs('false', () => {\n      expect(2).toMatchSnapshot();\n      expect(2).toMatchSnapshot('hello world');\n    });\n  });\n});",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          column: 17,
          line: 10,
        },
      ],
    },
    {
      code: 'const myReusableTestBody = (value, snapshotHint) => {\n  expect(value).toMatchSnapshot();\n\n  const innerFn = anotherValue => {\n    expect(anotherValue).toMatchSnapshot();\n  };\n\n  expect(value).toBe(1);\n  expect(value + 1).toMatchSnapshot(null);\n  expect(value + 2).toThrowErrorMatchingSnapshot(snapshotHint);\n};',
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          column: 17,
          line: 2,
        },
        {
          messageId: 'missingHint',
          column: 26,
          line: 5,
        },
        {
          messageId: 'missingHint',
          column: 21,
          line: 9,
        },
      ],
    },
    {
      code: 'const myReusableTestBody = (value, snapshotHint) => {\n  expect(value).toMatchSnapshot();\n\n  const innerFn = anotherValue => {\n    expect(anotherValue).toMatchSnapshot();\n\n    expect(value).toBe(1);\n    expect(value + 1).toMatchSnapshot(null);\n    expect(value + 2).toMatchSnapshot(null, snapshotHint);\n  };\n};',
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          column: 17,
          line: 2,
        },
        {
          messageId: 'missingHint',
          column: 26,
          line: 5,
        },
        {
          messageId: 'missingHint',
          column: 23,
          line: 8,
        },
      ],
    },
    {
      code: 'const myReusableTestBody = (value, snapshotHint) => {\n  const innerFn = anotherValue => {\n    expect(anotherValue).toMatchSnapshot();\n\n    expect(value).toBe(1);\n    expect(value + 1).toMatchSnapshot(null);\n    expect(value + 2).toMatchSnapshot(null, snapshotHint);\n  };\n\n  expect(value).toThrowErrorMatchingSnapshot();\n};',
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          column: 26,
          line: 3,
        },
        {
          messageId: 'missingHint',
          column: 23,
          line: 6,
        },
        {
          messageId: 'missingHint',
          column: 17,
          line: 10,
        },
      ],
    },
    {
      code: "const myReusableTestBody = (value, snapshotHint) => {\n  const innerFn = anotherValue => {\n    expect(anotherValue).toMatchSnapshot();\n\n    expect(value).toBe(1);\n  };\n\n  expect(value).toMatchSnapshot();\n};\n\nit('my test', () => {\n  expect(1).toMatchSnapshot();\n});",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          column: 26,
          line: 3,
        },
        {
          messageId: 'missingHint',
          column: 17,
          line: 8,
        },
      ],
    },
    {
      code: 'const myReusableTestBody = value => {\n  expect(value).toMatchSnapshot();\n};\n\nexpect(1).toMatchSnapshot();\nexpect(1).toThrowErrorMatchingSnapshot();',
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          column: 11,
          line: 5,
        },
        {
          messageId: 'missingHint',
          column: 11,
          line: 6,
        },
      ],
    },
  ],
});
