import { RuleTester } from '../rule-tester.js';

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/group-exports.js
// This wrapper checks counts and messages; Go tests also assert exact ranges,
// empty message IDs, and the absence of fixes and suggestions.
const rule = null as never;

/* eslint-disable max-len */
const errors = {
  named:
    'Multiple named export declarations; consolidate all named exports into a single export declaration',
  commonjs:
    'Multiple CommonJS exports; consolidate all exports into a single assignment to `module.exports`',
};
/* eslint-enable max-len */
const ruleTester = new RuleTester();

ruleTester.run('group-exports', rule, {
  valid: [
    { code: 'export const test = true' },
    {
      code: `
      export default {}
      export const test = true
    `,
    },
    {
      code: `
      const first = true
      const second = true
      export {
        first,
        second
      }
    `,
    },
    {
      code: `
      export default {}
      /* test */
      export const test = true
    `,
    },
    {
      code: `
      export default {}
      // test
      export const test = true
    `,
    },
    {
      code: `
      export const test = true
      /* test */
      export default {}
    `,
    },
    {
      code: `
      export const test = true
      // test
      export default {}
    `,
    },
    {
      code: `
      export { default as module1 } from './module-1'
      export { default as module2 } from './module-2'
    `,
    },
    { code: 'module.exports = {} ' },
    {
      code: `
      module.exports = { test: true,
        another: false }
    `,
    },
    { code: 'exports.test = true' },

    {
      code: `
      module.exports = {}
      const test = module.exports
    `,
    },
    {
      code: `
      exports.test = true
      const test = exports.test
    `,
    },
    {
      code: `
      module.exports = {}
      module.exports.too.deep = true
    `,
    },
    {
      code: `
      module.exports.deep.first = true
      module.exports.deep.second = true
    `,
    },
    {
      code: `
      module.exports = {}
      exports.too.deep = true
    `,
    },
    {
      code: `
      export default {}
      const test = true
      export { test }
    `,
    },
    {
      code: `
      const test = true
      export { test }
      const another = true
      export default {}
    `,
    },
    {
      code: `
      module.something.else = true
      module.something.different = true
    `,
    },
    {
      code: `
      module.exports.test = true
      module.something.different = true
    `,
    },
    {
      code: `
      exports.test = true
      module.something.different = true
    `,
    },
    {
      code: `
      unrelated = 'assignment'
      module.exports.test = true
    `,
    },
    {
      code: `
      type firstType = {
        propType: string
      };
      const first = {};
      export type { firstType };
      export { first };
    `,
    },
    {
      code: `
      type firstType = {
        propType: string
      };
      type secondType = {
        propType: string
      };
      export type { firstType, secondType };
    `,
    },
    {
      code: `
      export type { type1A, type1B } from './module-1'
      export { method1 } from './module-1'
    `,
    },
  ],
  invalid: [
    {
      code: `
        export const test = true
        export const another = true
      `,
      errors: [errors.named, errors.named],
    },
    {
      code: `
        export { method1 } from './module-1'
        export { method2 } from './module-1'
      `,
      errors: [errors.named, errors.named],
    },
    {
      code: `
        module.exports = {}
        module.exports.test = true
        module.exports.another = true
      `,
      errors: [errors.commonjs, errors.commonjs, errors.commonjs],
    },
    {
      code: `
        module.exports = {}
        module.exports.test = true
      `,
      errors: [errors.commonjs, errors.commonjs],
    },
    {
      code: `
        module.exports = { test: true }
        module.exports.another = true
      `,
      errors: [errors.commonjs, errors.commonjs],
    },
    {
      code: `
        module.exports.test = true
        module.exports.another = true
      `,
      errors: [errors.commonjs, errors.commonjs],
    },
    {
      code: `
        exports.test = true
        module.exports.another = true
      `,
      errors: [errors.commonjs, errors.commonjs],
    },
    {
      code: `
        module.exports = () => {}
        module.exports.attached = true
      `,
      errors: [errors.commonjs, errors.commonjs],
    },
    {
      code: `
        module.exports = function test() {}
        module.exports.attached = true
      `,
      errors: [errors.commonjs, errors.commonjs],
    },
    {
      code: `
        module.exports = () => {}
        exports.test = true
        exports.another = true
      `,
      errors: [errors.commonjs, errors.commonjs, errors.commonjs],
    },
    {
      code: `
        module.exports = "non-object"
        module.exports.attached = true
      `,
      errors: [errors.commonjs, errors.commonjs],
    },
    {
      code: `
        module.exports = "non-object"
        module.exports.attached = true
        module.exports.another = true
      `,
      errors: [errors.commonjs, errors.commonjs, errors.commonjs],
    },
    {
      code: `
        type firstType = {
          propType: string
        };
        type secondType = {
          propType: string
        };
        const first = {};
        export type { firstType };
        export type { secondType };
        export { first };
      `,
      errors: [errors.named, errors.named],
    },
    {
      code: `
        export type { type1 } from './module-1'
        export type { type2 } from './module-1'
      `,
      errors: [errors.named, errors.named],
    },
  ],
});

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/group-exports.md
describe('documentation examples', () => {
  ruleTester.run('group-exports', rule, {
    valid: [
      {
        code: '// A single named export declaration -> ok\nexport const valid = true\n',
      },
      {
        code: 'const first = true\nconst second = true\n\n// A single named export declaration -> ok\nexport {\n  first,\n  second,\n}\n',
      },
      {
        code: "// Aggregating exports -> ok\nexport { default as module1 } from 'module-1'\nexport { default as module2 } from 'module-2'\n",
      },
      {
        code: '// A single exports assignment -> ok\nmodule.exports = {\n  first: true,\n  second: true\n}\n',
      },
      {
        code: 'const first = true\nconst second = true\n\n// A single exports assignment -> ok\nmodule.exports = {\n  first,\n  second,\n}\n',
      },
      {
        code: 'function test() {}\ntest.property = true\ntest.another = true\n\n// A single exports assignment -> ok\nmodule.exports = test\n',
      },
      {
        code: 'const first = true;\ntype firstType = boolean\n\n// A single named export declaration (type exports handled separately) -> ok\nexport {first}\nexport type {firstType}\n',
      },
    ],
    invalid: [
      {
        code: '// Multiple named export statements -> not ok!\nexport const first = true\nexport const second = true\n',
        errors: [errors.named, errors.named],
      },
      {
        code: "// Aggregating exports from the same module -> not ok!\nexport { module1 } from 'module-1'\nexport { module2 } from 'module-1'\n",
        errors: [errors.named, errors.named],
      },
      {
        code: '// Multiple exports assignments -> not ok!\nexports.first = true\nexports.second = true\n',
        errors: [errors.commonjs, errors.commonjs],
      },
      {
        code: '// Multiple exports assignments -> not ok!\nmodule.exports = {}\nmodule.exports.first = true\n',
        errors: [errors.commonjs, errors.commonjs],
      },
      {
        code: '// Multiple exports assignments -> not ok!\nmodule.exports = () => {}\nmodule.exports.first = true\nmodule.exports.second = true\n',
        errors: [errors.commonjs, errors.commonjs, errors.commonjs],
      },
      {
        code: 'type firstType = boolean\ntype secondType = any\n\n// Multiple named type export statements -> not ok!\nexport type {firstType}\nexport type {secondType}\n',
        errors: [errors.named, errors.named],
      },
    ],
  });
});
