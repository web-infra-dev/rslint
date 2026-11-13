// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
new RuleTester(fixtureConfig()).run('group-exports', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: 'export const test = true',
    },
    {
      name: 'run 0 valid 1',
      code: '\n      export default {}\n      export const test = true\n    ',
    },
    {
      name: 'run 0 valid 2',
      code: '\n      const first = true\n      const second = true\n      export {\n        first,\n        second\n      }\n    ',
    },
    {
      name: 'run 0 valid 3',
      code: '\n      export default {}\n      /* test */\n      export const test = true\n    ',
    },
    {
      name: 'run 0 valid 4',
      code: '\n      export default {}\n      // test\n      export const test = true\n    ',
    },
    {
      name: 'run 0 valid 5',
      code: '\n      export const test = true\n      /* test */\n      export default {}\n    ',
    },
    {
      name: 'run 0 valid 6',
      code: '\n      export const test = true\n      // test\n      export default {}\n    ',
    },
    {
      name: 'run 0 valid 7',
      code: "\n      export { default as module1 } from './module-1'\n      export { default as module2 } from './module-2'\n    ",
    },
    {
      name: 'run 0 valid 8',
      code: 'module.exports = {} ',
    },
    {
      name: 'run 0 valid 9',
      code: '\n      module.exports = { test: true,\n        another: false }\n    ',
    },
    {
      name: 'run 0 valid 10',
      code: 'exports.test = true',
    },
    {
      name: 'run 0 valid 11',
      code: '\n      module.exports = {}\n      const test = module.exports\n    ',
    },
    {
      name: 'run 0 valid 12',
      code: '\n      exports.test = true\n      const test = exports.test\n    ',
    },
    {
      name: 'run 0 valid 13',
      code: '\n      module.exports = {}\n      module.exports.too.deep = true\n    ',
    },
    {
      name: 'run 0 valid 14',
      code: '\n      module.exports.deep.first = true\n      module.exports.deep.second = true\n    ',
    },
    {
      name: 'run 0 valid 15',
      code: '\n      module.exports = {}\n      exports.too.deep = true\n    ',
    },
    {
      name: 'run 0 valid 16',
      code: '\n      export default {}\n      const test = true\n      export { test }\n    ',
    },
    {
      name: 'run 0 valid 17',
      code: '\n      const test = true\n      export { test }\n      const another = true\n      export default {}\n    ',
    },
    {
      name: 'run 0 valid 18',
      code: '\n      module.something.else = true\n      module.something.different = true\n    ',
    },
    {
      name: 'run 0 valid 19',
      code: '\n      module.exports.test = true\n      module.something.different = true\n    ',
    },
    {
      name: 'run 0 valid 20',
      code: '\n      exports.test = true\n      module.something.different = true\n    ',
    },
    {
      name: 'run 0 valid 21',
      code: "\n      unrelated = 'assignment'\n      module.exports.test = true\n    ",
    },
    {
      name: 'run 0 valid 22',
      code: '\n      type firstType = {\n        propType: string\n      };\n      const first = {};\n      export type { firstType };\n      export { first };\n    ',
    },
    {
      name: 'run 0 valid 23',
      code: '\n      type firstType = {\n        propType: string\n      };\n      type secondType = {\n        propType: string\n      };\n      export type { firstType, secondType };\n    ',
    },
    {
      name: 'run 0 valid 24',
      code: "\n      export type { type1A, type1B } from './module-1'\n      export { method1 } from './module-1'\n    ",
    },
    {
      name: 'documentation examples / run 1 valid 0',
      code: '// A single named export declaration -> ok\nexport const valid = true\n',
    },
    {
      name: 'documentation examples / run 1 valid 1',
      code: 'const first = true\nconst second = true\n\n// A single named export declaration -> ok\nexport {\n  first,\n  second,\n}\n',
    },
    {
      name: 'documentation examples / run 1 valid 2',
      code: "// Aggregating exports -> ok\nexport { default as module1 } from 'module-1'\nexport { default as module2 } from 'module-2'\n",
    },
    {
      name: 'documentation examples / run 1 valid 3',
      code: '// A single exports assignment -> ok\nmodule.exports = {\n  first: true,\n  second: true\n}\n',
    },
    {
      name: 'documentation examples / run 1 valid 4',
      code: 'const first = true\nconst second = true\n\n// A single exports assignment -> ok\nmodule.exports = {\n  first,\n  second,\n}\n',
    },
    {
      name: 'documentation examples / run 1 valid 5',
      code: 'function test() {}\ntest.property = true\ntest.another = true\n\n// A single exports assignment -> ok\nmodule.exports = test\n',
    },
    {
      name: 'documentation examples / run 1 valid 6',
      code: 'const first = true;\ntype firstType = boolean\n\n// A single named export declaration (type exports handled separately) -> ok\nexport {first}\nexport type {firstType}\n',
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: '\n        export const test = true\n        export const another = true\n      ',
      errors: [
        {
          messageId: '',
          message:
            'Multiple named export declarations; consolidate all named exports into a single export declaration',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 33,
        },
        {
          messageId: '',
          message:
            'Multiple named export declarations; consolidate all named exports into a single export declaration',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 36,
        },
      ],
    },
    {
      name: 'run 0 invalid 1',
      code: "\n        export { method1 } from './module-1'\n        export { method2 } from './module-1'\n      ",
      errors: [
        {
          messageId: '',
          message:
            'Multiple named export declarations; consolidate all named exports into a single export declaration',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 45,
        },
        {
          messageId: '',
          message:
            'Multiple named export declarations; consolidate all named exports into a single export declaration',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 45,
        },
      ],
    },
    {
      name: 'run 0 invalid 2',
      code: '\n        module.exports = {}\n        module.exports.test = true\n        module.exports.another = true\n      ',
      errors: [
        {
          messageId: '',
          message:
            'Multiple CommonJS exports; consolidate all exports into a single assignment to `module.exports`',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 28,
        },
        {
          messageId: '',
          message:
            'Multiple CommonJS exports; consolidate all exports into a single assignment to `module.exports`',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 35,
        },
        {
          messageId: '',
          message:
            'Multiple CommonJS exports; consolidate all exports into a single assignment to `module.exports`',
          line: 4,
          column: 9,
          endLine: 4,
          endColumn: 38,
        },
      ],
    },
    {
      name: 'run 0 invalid 3',
      code: '\n        module.exports = {}\n        module.exports.test = true\n      ',
      errors: [
        {
          messageId: '',
          message:
            'Multiple CommonJS exports; consolidate all exports into a single assignment to `module.exports`',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 28,
        },
        {
          messageId: '',
          message:
            'Multiple CommonJS exports; consolidate all exports into a single assignment to `module.exports`',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 35,
        },
      ],
    },
    {
      name: 'run 0 invalid 4',
      code: '\n        module.exports = { test: true }\n        module.exports.another = true\n      ',
      errors: [
        {
          messageId: '',
          message:
            'Multiple CommonJS exports; consolidate all exports into a single assignment to `module.exports`',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 40,
        },
        {
          messageId: '',
          message:
            'Multiple CommonJS exports; consolidate all exports into a single assignment to `module.exports`',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 38,
        },
      ],
    },
    {
      name: 'run 0 invalid 5',
      code: '\n        module.exports.test = true\n        module.exports.another = true\n      ',
      errors: [
        {
          messageId: '',
          message:
            'Multiple CommonJS exports; consolidate all exports into a single assignment to `module.exports`',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 35,
        },
        {
          messageId: '',
          message:
            'Multiple CommonJS exports; consolidate all exports into a single assignment to `module.exports`',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 38,
        },
      ],
    },
    {
      name: 'run 0 invalid 6',
      code: '\n        exports.test = true\n        module.exports.another = true\n      ',
      errors: [
        {
          messageId: '',
          message:
            'Multiple CommonJS exports; consolidate all exports into a single assignment to `module.exports`',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 28,
        },
        {
          messageId: '',
          message:
            'Multiple CommonJS exports; consolidate all exports into a single assignment to `module.exports`',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 38,
        },
      ],
    },
    {
      name: 'run 0 invalid 7',
      code: '\n        module.exports = () => {}\n        module.exports.attached = true\n      ',
      errors: [
        {
          messageId: '',
          message:
            'Multiple CommonJS exports; consolidate all exports into a single assignment to `module.exports`',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 34,
        },
        {
          messageId: '',
          message:
            'Multiple CommonJS exports; consolidate all exports into a single assignment to `module.exports`',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 39,
        },
      ],
    },
    {
      name: 'run 0 invalid 8',
      code: '\n        module.exports = function test() {}\n        module.exports.attached = true\n      ',
      errors: [
        {
          messageId: '',
          message:
            'Multiple CommonJS exports; consolidate all exports into a single assignment to `module.exports`',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 44,
        },
        {
          messageId: '',
          message:
            'Multiple CommonJS exports; consolidate all exports into a single assignment to `module.exports`',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 39,
        },
      ],
    },
    {
      name: 'run 0 invalid 9',
      code: '\n        module.exports = () => {}\n        exports.test = true\n        exports.another = true\n      ',
      errors: [
        {
          messageId: '',
          message:
            'Multiple CommonJS exports; consolidate all exports into a single assignment to `module.exports`',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 34,
        },
        {
          messageId: '',
          message:
            'Multiple CommonJS exports; consolidate all exports into a single assignment to `module.exports`',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 28,
        },
        {
          messageId: '',
          message:
            'Multiple CommonJS exports; consolidate all exports into a single assignment to `module.exports`',
          line: 4,
          column: 9,
          endLine: 4,
          endColumn: 31,
        },
      ],
    },
    {
      name: 'run 0 invalid 10',
      code: '\n        module.exports = "non-object"\n        module.exports.attached = true\n      ',
      errors: [
        {
          messageId: '',
          message:
            'Multiple CommonJS exports; consolidate all exports into a single assignment to `module.exports`',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 38,
        },
        {
          messageId: '',
          message:
            'Multiple CommonJS exports; consolidate all exports into a single assignment to `module.exports`',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 39,
        },
      ],
    },
    {
      name: 'run 0 invalid 11',
      code: '\n        module.exports = "non-object"\n        module.exports.attached = true\n        module.exports.another = true\n      ',
      errors: [
        {
          messageId: '',
          message:
            'Multiple CommonJS exports; consolidate all exports into a single assignment to `module.exports`',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 38,
        },
        {
          messageId: '',
          message:
            'Multiple CommonJS exports; consolidate all exports into a single assignment to `module.exports`',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 39,
        },
        {
          messageId: '',
          message:
            'Multiple CommonJS exports; consolidate all exports into a single assignment to `module.exports`',
          line: 4,
          column: 9,
          endLine: 4,
          endColumn: 38,
        },
      ],
    },
    {
      name: 'run 0 invalid 12',
      code: '\n        type firstType = {\n          propType: string\n        };\n        type secondType = {\n          propType: string\n        };\n        const first = {};\n        export type { firstType };\n        export type { secondType };\n        export { first };\n      ',
      errors: [
        {
          messageId: '',
          message:
            'Multiple named export declarations; consolidate all named exports into a single export declaration',
          line: 9,
          column: 9,
          endLine: 9,
          endColumn: 35,
        },
        {
          messageId: '',
          message:
            'Multiple named export declarations; consolidate all named exports into a single export declaration',
          line: 10,
          column: 9,
          endLine: 10,
          endColumn: 36,
        },
      ],
    },
    {
      name: 'run 0 invalid 13',
      code: "\n        export type { type1 } from './module-1'\n        export type { type2 } from './module-1'\n      ",
      errors: [
        {
          messageId: '',
          message:
            'Multiple named export declarations; consolidate all named exports into a single export declaration',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 48,
        },
        {
          messageId: '',
          message:
            'Multiple named export declarations; consolidate all named exports into a single export declaration',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 48,
        },
      ],
    },
    {
      name: 'documentation examples / run 1 invalid 0',
      code: '// Multiple named export statements -> not ok!\nexport const first = true\nexport const second = true\n',
      errors: [
        {
          messageId: '',
          message:
            'Multiple named export declarations; consolidate all named exports into a single export declaration',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 26,
        },
        {
          messageId: '',
          message:
            'Multiple named export declarations; consolidate all named exports into a single export declaration',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 27,
        },
      ],
    },
    {
      name: 'documentation examples / run 1 invalid 1',
      code: "// Aggregating exports from the same module -> not ok!\nexport { module1 } from 'module-1'\nexport { module2 } from 'module-1'\n",
      errors: [
        {
          messageId: '',
          message:
            'Multiple named export declarations; consolidate all named exports into a single export declaration',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 35,
        },
        {
          messageId: '',
          message:
            'Multiple named export declarations; consolidate all named exports into a single export declaration',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 35,
        },
      ],
    },
    {
      name: 'documentation examples / run 1 invalid 2',
      code: '// Multiple exports assignments -> not ok!\nexports.first = true\nexports.second = true\n',
      errors: [
        {
          messageId: '',
          message:
            'Multiple CommonJS exports; consolidate all exports into a single assignment to `module.exports`',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 21,
        },
        {
          messageId: '',
          message:
            'Multiple CommonJS exports; consolidate all exports into a single assignment to `module.exports`',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 22,
        },
      ],
    },
    {
      name: 'documentation examples / run 1 invalid 3',
      code: '// Multiple exports assignments -> not ok!\nmodule.exports = {}\nmodule.exports.first = true\n',
      errors: [
        {
          messageId: '',
          message:
            'Multiple CommonJS exports; consolidate all exports into a single assignment to `module.exports`',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 20,
        },
        {
          messageId: '',
          message:
            'Multiple CommonJS exports; consolidate all exports into a single assignment to `module.exports`',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 28,
        },
      ],
    },
    {
      name: 'documentation examples / run 1 invalid 4',
      code: '// Multiple exports assignments -> not ok!\nmodule.exports = () => {}\nmodule.exports.first = true\nmodule.exports.second = true\n',
      errors: [
        {
          messageId: '',
          message:
            'Multiple CommonJS exports; consolidate all exports into a single assignment to `module.exports`',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 26,
        },
        {
          messageId: '',
          message:
            'Multiple CommonJS exports; consolidate all exports into a single assignment to `module.exports`',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 28,
        },
        {
          messageId: '',
          message:
            'Multiple CommonJS exports; consolidate all exports into a single assignment to `module.exports`',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 29,
        },
      ],
    },
    {
      name: 'documentation examples / run 1 invalid 5',
      code: 'type firstType = boolean\ntype secondType = any\n\n// Multiple named type export statements -> not ok!\nexport type {firstType}\nexport type {secondType}\n',
      errors: [
        {
          messageId: '',
          message:
            'Multiple named export declarations; consolidate all named exports into a single export declaration',
          line: 5,
          column: 1,
          endLine: 5,
          endColumn: 24,
        },
        {
          messageId: '',
          message:
            'Multiple named export declarations; consolidate all named exports into a single export declaration',
          line: 6,
          column: 1,
          endLine: 6,
          endColumn: 25,
        },
      ],
    },
  ],
});
