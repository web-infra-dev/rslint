import { RuleTester } from '../rule-tester';
const ruleTester = new RuleTester();
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
      code: "\n       it('is true', () => {\n      expect(1).toMatchSnapshot();\n       });\n     ",
      options: ['multi'],
    },
    {
      code: "\n       it('is true', () => {\n      expect(1).toMatchSnapshot(undefined, 'my first snapshot');\n       });\n     ",
      options: ['multi'],
    },
    {
      code: "\n       describe('my tests', () => {\n      it('is true', () => {\n        expect(1).toMatchSnapshot('this is a hint, all by itself');\n      });\n     \n      it('is false', () => {\n        expect(2).toMatchSnapshot('this is a hint');\n        expect(2).toMatchSnapshot('and so is this');\n      });\n       });\n     ",
      options: ['multi'],
    },
    {
      code: "\n       it('is true', () => {\n      expect(1).toMatchSnapshot();\n       });\n     \n       it('is false', () => {\n      expect(2).toMatchSnapshot('this is a hint');\n      expect(2).toMatchSnapshot('and so is this');\n       });\n     ",
      options: ['multi'],
    },
    {
      code: "\n       it('is true', () => {\n      expect(1).toMatchSnapshot();\n       });\n     \n       it('is false', () => {\n      expect(2).toThrowErrorMatchingSnapshot();\n       });\n     ",
      options: ['multi'],
    },
    {
      code: "\n       it('is true', () => {\n      expect(1).toStrictEqual(1);\n      expect(1).toStrictEqual(2);\n      expect(1).toMatchSnapshot();\n       });\n     \n       it('is false', () => {\n      expect(1).toStrictEqual(1);\n      expect(1).toStrictEqual(2);\n      expect(2).toThrowErrorMatchingSnapshot();\n       });\n     ",
      options: ['multi'],
    },
    {
      code: "\n       it('is true', () => {\n      expect(1).toMatchInlineSnapshot();\n       });\n     \n       it('is false', () => {\n      expect(1).toMatchInlineSnapshot();\n      expect(1).toMatchInlineSnapshot();\n      expect(1).toThrowErrorMatchingInlineSnapshot();\n       });\n     ",
      options: ['multi'],
    },
    {
      code: "\n       it('is true', () => {\n      expect(1).toMatchSnapshot();\n       });\n     \n       it('is false', () => {\n      expect(1).toMatchSnapshot();\n       });\n     ",
      options: ['multi'],
    },
    {
      code: "\n       const myReusableTestBody = (value, snapshotHint) => {\n      const innerFn = anotherValue => {\n        expect(anotherValue).toMatchSnapshot();\n     \n        expect(value).toBe(1);\n      };\n     \n      expect(value).toBe(1);\n       };\n     \n       it('my test', () => {\n      expect(1).toMatchSnapshot();\n       });\n     ",
      options: ['multi'],
    },
    {
      code: "\n       const myReusableTestBody = (value, snapshotHint) => {\n      const innerFn = anotherValue => {\n        expect(value).toBe(1);\n      };\n     \n      expect(value).toBe(1);\n      expect(anotherValue).toMatchSnapshot();\n       };\n     \n       it('my test', () => {\n      expect(1).toMatchSnapshot();\n       });\n     ",
      options: ['multi'],
    },
    {
      code: '\n       const myReusableTestBody = (value, snapshotHint) => {\n      const innerFn = anotherValue => {\n        expect(anotherValue).toMatchSnapshot();\n     \n        expect(value).toBe(1);\n      };\n     \n      expect(value).toBe(1);\n       };\n     \n       expect(1).toMatchSnapshot();\n     ',
      options: ['multi'],
    },
    {
      code: "const register = () => {\n  const first = () => expect('first').toMatchSnapshot();\n  test('first', first);\n  const second = () => expect('second').toMatchSnapshot();\n  test('second', second);\n};\ndescribe('suite', register);",
      options: ['multi'],
    },
    {
      code: "expect('x').toBe('x').and.not.toMatchSnapshot();",
      options: ['always'],
    },
    {
      code: "test('case', callback);\nfunction callback(): void;\nfunction callback() { expect('test').toMatchSnapshot(); }\nfunction helper() { expect('helper').toMatchSnapshot(); }",
      options: ['multi'],
    },
  ],
  invalid: [
    {
      code: "const outer = () => {\n  expect('before').toMatchSnapshot();\n  let callback = () => { expect('stale').toMatchSnapshot(); };\n  callback = () => {};\n  test('case', callback);\n  expect('after').toMatchSnapshot();\n};",
      options: ['multi'],
      errors: 3,
    },
    {
      code: "const outer = () => {\n  expect('before').toMatchSnapshot();\n  { const callback = () => expect('one').toMatchSnapshot(); test('one', callback); }\n  { const callback = () => expect('two').toMatchSnapshot(); test('two', callback); }\n  expect('after').toMatchSnapshot();\n};",
      options: ['multi'],
      errors: 2,
    },
    {
      code: "it('is true', () => {\n      expect(1).toMatchSnapshot();\n      expect(2).toMatchSnapshot();\n       });\n     ",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          column: 17,
          line: 2,
        },
        {
          messageId: 'missingHint',
          column: 17,
          line: 3,
        },
      ],
    },
    {
      code: "it('is true', () => {\n        expect(1).toMatchSnapshot();\n        expect(2).toThrowErrorMatchingSnapshot();\n      });\n       ",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          column: 19,
          line: 2,
        },
        {
          messageId: 'missingHint',
          column: 19,
          line: 3,
        },
      ],
    },
    {
      code: "it('is true', () => {\n        expect(1).toThrowErrorMatchingSnapshot();\n        expect(2).toMatchSnapshot();\n      });\n       ",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          column: 19,
          line: 2,
        },
        {
          messageId: 'missingHint',
          column: 19,
          line: 3,
        },
      ],
    },
    {
      code: "it('is true', () => {\n        expect(1).toMatchSnapshot({});\n        expect(2).toMatchSnapshot({});\n      });\n       ",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          column: 19,
          line: 2,
        },
        {
          messageId: 'missingHint',
          column: 19,
          line: 3,
        },
      ],
    },
    {
      code: "it('is true', () => {\n       expect(1).toMatchSnapshot({});\n       {\n      expect(2).toMatchSnapshot({});\n       }\n     });\n      ",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          column: 18,
          line: 2,
        },
        {
          messageId: 'missingHint',
          column: 17,
          line: 4,
        },
      ],
    },
    {
      code: "it('is true', () => {\n       { expect(1).toMatchSnapshot(); }\n       { expect(2).toMatchSnapshot(); }\n     });\n      ",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          column: 20,
          line: 2,
        },
        {
          messageId: 'missingHint',
          column: 20,
          line: 3,
        },
      ],
    },
    {
      code: "it('is true', () => {\n        expect(1).toMatchSnapshot();\n        expect(2).toMatchSnapshot(undefined, 'my second snapshot');\n      });\n       ",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          column: 19,
          line: 2,
        },
      ],
    },
    {
      code: "it('is true', () => {\n        expect(1).toMatchSnapshot({});\n        expect(2).toMatchSnapshot(undefined, 'my second snapshot');\n      });",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          column: 19,
          line: 2,
        },
      ],
    },
    {
      code: "it('is true', () => {\n        expect(1).toMatchSnapshot({}, 'my first snapshot');\n        expect(2).toMatchSnapshot(undefined);\n      });\n       ",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          column: 19,
          line: 3,
        },
      ],
    },
    {
      code: "it('is true', () => {\n        expect(1).toMatchSnapshot({}, 'my first snapshot');\n        expect(2).toMatchSnapshot(undefined);\n        expect(2).toMatchSnapshot();\n      });\n       ",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          column: 19,
          line: 3,
        },
        {
          messageId: 'missingHint',
          column: 19,
          line: 4,
        },
      ],
    },
    {
      code: "it('is true', () => {\n        expect(2).toMatchSnapshot();\n        expect(1).toMatchSnapshot({}, 'my second snapshot');\n        expect(2).toMatchSnapshot();\n      });\n       ",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          column: 19,
          line: 2,
        },
        {
          messageId: 'missingHint',
          column: 19,
          line: 4,
        },
      ],
    },
    {
      code: "it('is true', () => {\n        expect(2).toMatchSnapshot(undefined);\n        expect(2).toMatchSnapshot();\n        expect(1).toMatchSnapshot(null, 'my third snapshot');\n      });\n       ",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          column: 19,
          line: 2,
        },
        {
          messageId: 'missingHint',
          column: 19,
          line: 3,
        },
      ],
    },
    {
      code: "describe('my tests', () => {\n        it('is true', () => {\n       expect(1).toMatchSnapshot();\n        });\n \n        it('is false', () => {\n       expect(2).toMatchSnapshot();\n       expect(2).toMatchSnapshot();\n        });\n      });\n       ",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          column: 18,
          line: 7,
        },
        {
          messageId: 'missingHint',
          column: 18,
          line: 8,
        },
      ],
    },
    {
      code: "describe('my tests', () => {\n        it('is true', () => {\n       expect(1).toMatchSnapshot();\n        });\n \n        it('is false', () => {\n       expect(2).toMatchSnapshot();\n       expect(2).toMatchSnapshot('hello world');\n        });\n      });\n       ",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          column: 18,
          line: 7,
        },
      ],
    },
    {
      code: "describe('my tests', () => {\n        describe('more tests', () => {\n       it('is true', () => {\n         expect(1).toMatchSnapshot();\n       });\n        });\n \n        it('is false', () => {\n       expect(2).toMatchSnapshot();\n       expect(2).toMatchSnapshot('hello world');\n        });\n      });\n       ",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          column: 18,
          line: 9,
        },
      ],
    },
  ],
});
