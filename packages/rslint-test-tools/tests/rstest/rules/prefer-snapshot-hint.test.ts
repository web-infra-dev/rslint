import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';
const ruleTester = new RuleTester(config);
ruleTester.run('prefer-snapshot-hint', {
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
      errors: [
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 2,
          column: 20,
          endLine: 2,
          endColumn: 35,
        },
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 3,
          column: 42,
          endLine: 3,
          endColumn: 57,
        },
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 6,
          column: 19,
          endLine: 6,
          endColumn: 34,
        },
      ],
    },
    {
      code: "const outer = () => {\n  expect('before').toMatchSnapshot();\n  { const callback = () => expect('one').toMatchSnapshot(); test('one', callback); }\n  { const callback = () => expect('two').toMatchSnapshot(); test('two', callback); }\n  expect('after').toMatchSnapshot();\n};",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 2,
          column: 20,
          endLine: 2,
          endColumn: 35,
        },
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 5,
          column: 19,
          endLine: 5,
          endColumn: 34,
        },
      ],
    },
    {
      code: "it('is true', () => {\n      expect(1).toMatchSnapshot();\n      expect(2).toMatchSnapshot();\n       });\n     ",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 2,
          column: 17,
          endLine: 2,
          endColumn: 32,
        },
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 3,
          column: 17,
          endLine: 3,
          endColumn: 32,
        },
      ],
    },
    {
      code: "it('is true', () => {\n        expect(1).toMatchSnapshot();\n        expect(2).toThrowErrorMatchingSnapshot();\n      });\n       ",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 2,
          column: 19,
          endLine: 2,
          endColumn: 34,
        },
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 3,
          column: 19,
          endLine: 3,
          endColumn: 47,
        },
      ],
    },
    {
      code: "it('is true', () => {\n        expect(1).toThrowErrorMatchingSnapshot();\n        expect(2).toMatchSnapshot();\n      });\n       ",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 2,
          column: 19,
          endLine: 2,
          endColumn: 47,
        },
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 3,
          column: 19,
          endLine: 3,
          endColumn: 34,
        },
      ],
    },
    {
      code: "it('is true', () => {\n        expect(1).toMatchSnapshot({});\n        expect(2).toMatchSnapshot({});\n      });\n       ",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 2,
          column: 19,
          endLine: 2,
          endColumn: 34,
        },
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 3,
          column: 19,
          endLine: 3,
          endColumn: 34,
        },
      ],
    },
    {
      code: "it('is true', () => {\n       expect(1).toMatchSnapshot({});\n       {\n      expect(2).toMatchSnapshot({});\n       }\n     });\n      ",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 2,
          column: 18,
          endLine: 2,
          endColumn: 33,
        },
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 4,
          column: 17,
          endLine: 4,
          endColumn: 32,
        },
      ],
    },
    {
      code: "it('is true', () => {\n       { expect(1).toMatchSnapshot(); }\n       { expect(2).toMatchSnapshot(); }\n     });\n      ",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 2,
          column: 20,
          endLine: 2,
          endColumn: 35,
        },
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 3,
          column: 20,
          endLine: 3,
          endColumn: 35,
        },
      ],
    },
    {
      code: "it('is true', () => {\n        expect(1).toMatchSnapshot();\n        expect(2).toMatchSnapshot(undefined, 'my second snapshot');\n      });\n       ",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 2,
          column: 19,
          endLine: 2,
          endColumn: 34,
        },
      ],
    },
    {
      code: "it('is true', () => {\n        expect(1).toMatchSnapshot({});\n        expect(2).toMatchSnapshot(undefined, 'my second snapshot');\n      });",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 2,
          column: 19,
          endLine: 2,
          endColumn: 34,
        },
      ],
    },
    {
      code: "it('is true', () => {\n        expect(1).toMatchSnapshot({}, 'my first snapshot');\n        expect(2).toMatchSnapshot(undefined);\n      });\n       ",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 3,
          column: 19,
          endLine: 3,
          endColumn: 34,
        },
      ],
    },
    {
      code: "it('is true', () => {\n        expect(1).toMatchSnapshot({}, 'my first snapshot');\n        expect(2).toMatchSnapshot(undefined);\n        expect(2).toMatchSnapshot();\n      });\n       ",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 3,
          column: 19,
          endLine: 3,
          endColumn: 34,
        },
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 4,
          column: 19,
          endLine: 4,
          endColumn: 34,
        },
      ],
    },
    {
      code: "it('is true', () => {\n        expect(2).toMatchSnapshot();\n        expect(1).toMatchSnapshot({}, 'my second snapshot');\n        expect(2).toMatchSnapshot();\n      });\n       ",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 2,
          column: 19,
          endLine: 2,
          endColumn: 34,
        },
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 4,
          column: 19,
          endLine: 4,
          endColumn: 34,
        },
      ],
    },
    {
      code: "it('is true', () => {\n        expect(2).toMatchSnapshot(undefined);\n        expect(2).toMatchSnapshot();\n        expect(1).toMatchSnapshot(null, 'my third snapshot');\n      });\n       ",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 2,
          column: 19,
          endLine: 2,
          endColumn: 34,
        },
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 3,
          column: 19,
          endLine: 3,
          endColumn: 34,
        },
      ],
    },
    {
      code: "describe('my tests', () => {\n        it('is true', () => {\n       expect(1).toMatchSnapshot();\n        });\n \n        it('is false', () => {\n       expect(2).toMatchSnapshot();\n       expect(2).toMatchSnapshot();\n        });\n      });\n       ",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 7,
          column: 18,
          endLine: 7,
          endColumn: 33,
        },
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 8,
          column: 18,
          endLine: 8,
          endColumn: 33,
        },
      ],
    },
    {
      code: "describe('my tests', () => {\n        it('is true', () => {\n       expect(1).toMatchSnapshot();\n        });\n \n        it('is false', () => {\n       expect(2).toMatchSnapshot();\n       expect(2).toMatchSnapshot('hello world');\n        });\n      });\n       ",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 7,
          column: 18,
          endLine: 7,
          endColumn: 33,
        },
      ],
    },
    {
      code: "describe('my tests', () => {\n        describe('more tests', () => {\n       it('is true', () => {\n         expect(1).toMatchSnapshot();\n       });\n        });\n \n        it('is false', () => {\n       expect(2).toMatchSnapshot();\n       expect(2).toMatchSnapshot('hello world');\n        });\n      });\n       ",
      options: ['multi'],
      errors: [
        {
          messageId: 'missingHint',
          message: 'You should provide a hint for this snapshot',
          line: 9,
          column: 18,
          endLine: 9,
          endColumn: 33,
        },
      ],
    },
  ],
});
