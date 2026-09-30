// Upstream tests and documentation from eslint-plugin-unicorn v76.0.0 (MIT).
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v76.0.0/test/no-anonymous-default-export.js
import path from 'node:path';
import { lint } from '@rslint/core/internal';

// Cases without an upstream filename use the same physical case.js in both runners.
const cases = {
  valid: [
    {
      code: 'export default function named() {}',
      filename: '/path/to/case.js',
      origin: 'Upstream snapshot cases',
    },
    {
      code: 'export default class named {}',
      filename: '/path/to/case.js',
      origin: 'Upstream snapshot cases',
    },
    {
      code: 'export default []',
      filename: '/path/to/case.js',
      origin: 'Upstream snapshot cases',
    },
    {
      code: 'export default {}',
      filename: '/path/to/case.js',
      origin: 'Upstream snapshot cases',
    },
    {
      code: 'export default 1',
      filename: '/path/to/case.js',
      origin: 'Upstream snapshot cases',
    },
    {
      code: 'export default false',
      filename: '/path/to/case.js',
      origin: 'Upstream snapshot cases',
    },
    {
      code: 'export default 0n',
      filename: '/path/to/case.js',
      origin: 'Upstream snapshot cases',
    },
    {
      code: 'notExports = class {}',
      filename: '/path/to/case.js',
      origin: 'Upstream snapshot cases',
    },
    {
      code: 'notModule.exports = class {}',
      filename: '/path/to/case.js',
      origin: 'Upstream snapshot cases',
    },
    {
      code: 'module.notExports = class {}',
      filename: '/path/to/case.js',
      origin: 'Upstream snapshot cases',
    },
    {
      code: 'module.exports.foo = class {}',
      filename: '/path/to/case.js',
      origin: 'Upstream snapshot cases',
    },
    {
      code: 'alert(exports = class {})',
      filename: '/path/to/case.js',
      origin: 'Upstream snapshot cases',
    },
    {
      code: 'foo = module.exports = class {}',
      filename: '/path/to/case.js',
      origin: 'Upstream snapshot cases',
    },
    {
      code: 'export default class Foo {}',
      filename: '/path/to/foo.js',
      origin: 'Documentation examples',
    },
    {
      code: 'export default function foo() {}',
      filename: '/path/to/foo.js',
      origin: 'Documentation examples',
    },
    {
      code: 'const foo = () => {};\nexport default foo;',
      filename: '/path/to/foo.js',
      origin: 'Documentation examples',
    },
    {
      code: 'module.exports = class Foo {};',
      filename: '/path/to/foo.js',
      origin: 'Documentation examples',
    },
    {
      code: 'module.exports = function foo() {};',
      filename: '/path/to/foo.js',
      origin: 'Documentation examples',
    },
    {
      code: 'const foo = () => {};\nmodule.exports = foo;',
      filename: '/path/to/foo.js',
      origin: 'Documentation examples',
    },
  ],
  invalid: [
    {
      code: 'export default function () {}',
      filename: '/path/to/case.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The function should be named.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 25,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `case_`.',
              output: 'export default function case_ () {}',
            },
          ],
        },
      ],
    },
    {
      code: 'export default class {}',
      filename: '/path/to/case.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The class should be named.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `Case`.',
              output: 'export default class Case {}',
            },
          ],
        },
      ],
    },
    {
      code: 'export default () => {}',
      filename: '/path/to/case.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The arrow function should be named.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `case_`.',
              output: 'const case_ = () => {};\nexport default case_;',
            },
          ],
        },
      ],
    },
    {
      code: 'export default function * () {}',
      filename: '/path/to/case.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The generator function should be named.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 27,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `case_`.',
              output: 'export default function * case_ () {}',
            },
          ],
        },
      ],
    },
    {
      code: 'export default async function () {}',
      filename: '/path/to/case.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The async function should be named.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 31,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `case_`.',
              output: 'export default async function case_ () {}',
            },
          ],
        },
      ],
    },
    {
      code: 'export default async function * () {}',
      filename: '/path/to/case.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The async generator function should be named.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 33,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `case_`.',
              output: 'export default async function * case_ () {}',
            },
          ],
        },
      ],
    },
    {
      code: 'export default async () => {}',
      filename: '/path/to/case.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The async arrow function should be named.',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 27,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `case_`.',
              output: 'const case_ = async () => {};\nexport default case_;',
            },
          ],
        },
      ],
    },
    {
      code: 'export default class {}',
      filename: '/path/to/foo.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The class should be named.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `Foo`.',
              output: 'export default class Foo {}',
            },
          ],
        },
      ],
    },
    {
      code: 'export default class extends class {} {}',
      filename: '/path/to/foo.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The class should be named.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 38,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `Foo`.',
              output: 'export default class Foo extends class {} {}',
            },
          ],
        },
      ],
    },
    {
      code: 'export default class{}',
      filename: '/path/to/foo.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The class should be named.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `Foo`.',
              output: 'export default class Foo{}',
            },
          ],
        },
      ],
    },
    {
      code: 'export default class {}',
      filename: '/path/to/foo-bar.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The class should be named.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `FooBar`.',
              output: 'export default class FooBar {}',
            },
          ],
        },
      ],
    },
    {
      code: 'export default class {}',
      filename: '/path/to/foo_bar.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The class should be named.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `FooBar`.',
              output: 'export default class FooBar {}',
            },
          ],
        },
      ],
    },
    {
      code: 'export default class {}',
      filename: '/path/to/foo+bar.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The class should be named.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `FooBar`.',
              output: 'export default class FooBar {}',
            },
          ],
        },
      ],
    },
    {
      code: 'export default class {}',
      filename: '/path/to/foo+bar123.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The class should be named.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `FooBar123`.',
              output: 'export default class FooBar123 {}',
            },
          ],
        },
      ],
    },
    {
      code: 'export default class {}',
      filename: '/path/to/foo*.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The class should be named.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `Foo`.',
              output: 'export default class Foo {}',
            },
          ],
        },
      ],
    },
    {
      code: 'export default class {}',
      filename: '/path/to/[foo].js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The class should be named.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `Foo`.',
              output: 'export default class Foo {}',
            },
          ],
        },
      ],
    },
    {
      code: 'export default class {}',
      filename: '/path/to/class.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The class should be named.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `Class`.',
              output: 'export default class Class {}',
            },
          ],
        },
      ],
    },
    {
      code: 'export default class {}',
      filename: '/path/to/foo.helper.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The class should be named.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `Foo`.',
              output: 'export default class Foo {}',
            },
          ],
        },
      ],
    },
    {
      code: 'export default class {}',
      filename: '/path/to/foo.bar.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The class should be named.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `Foo`.',
              output: 'export default class Foo {}',
            },
          ],
        },
      ],
    },
    {
      code: 'export default class {}',
      filename: '/path/to/foo.test.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The class should be named.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `Foo`.',
              output: 'export default class Foo {}',
            },
          ],
        },
      ],
    },
    {
      code: 'export default class {}',
      filename: '/path/to/.foo.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The class should be named.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 21,
          suggestions: [],
        },
      ],
    },
    {
      code: 'let Foo, Foo_, foo, foo_\nexport default class {}',
      filename: '/path/to/foo.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The class should be named.',
          line: 2,
          column: 16,
          endLine: 2,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `Foo__`.',
              output: 'let Foo, Foo_, foo, foo_\nexport default class Foo__ {}',
            },
          ],
        },
      ],
    },
    {
      code: 'let Foo, Foo_, foo, foo_\nexport default (class{})',
      filename: '/path/to/foo.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The class should be named.',
          line: 2,
          column: 17,
          endLine: 2,
          endColumn: 22,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `Foo__`.',
              output:
                'let Foo, Foo_, foo, foo_\nexport default (class Foo__{})',
            },
          ],
        },
      ],
    },
    {
      code: 'export default (class extends class {} {})',
      filename: '/path/to/foo.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The class should be named.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 39,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `Foo`.',
              output: 'export default (class Foo extends class {} {})',
            },
          ],
        },
      ],
    },
    {
      code: 'let Exports, Exports_, exports, exports_\nexports = class {}',
      filename: '/path/to/exports.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The class should be named.',
          line: 2,
          column: 11,
          endLine: 2,
          endColumn: 16,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `Exports__`.',
              output:
                'let Exports, Exports_, exports, exports_\nexports = class Exports__ {}',
            },
          ],
        },
      ],
    },
    {
      code: 'module.exports = class {}',
      filename: '/path/to/module.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The class should be named.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 23,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `Module`.',
              output: 'module.exports = class Module {}',
            },
          ],
        },
      ],
    },
    {
      code: 'module.exports = () => {}',
      filename: '/path/to/module.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The arrow function should be named.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 23,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `module_`.',
              output: 'const module_ = () => {};\nmodule.exports = module_;',
            },
          ],
        },
      ],
    },
    {
      code: 'exports = () => {}',
      filename: '/path/to/module.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The arrow function should be named.',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 16,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `module_`.',
              output: 'const module_ = () => {};\nexports = module_;',
            },
          ],
        },
      ],
    },
    {
      code: 'export default function () {}',
      filename: '/path/to/foo.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The function should be named.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 25,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `foo`.',
              output: 'export default function foo () {}',
            },
          ],
        },
      ],
    },
    {
      code: 'export default function* () {}',
      filename: '/path/to/foo.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The generator function should be named.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 26,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `foo`.',
              output: 'export default function* foo () {}',
            },
          ],
        },
      ],
    },
    {
      code: 'export default async function* () {}',
      filename: '/path/to/foo.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The async generator function should be named.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 32,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `foo`.',
              output: 'export default async function* foo () {}',
            },
          ],
        },
      ],
    },
    {
      code: 'export default async function*() {}',
      filename: '/path/to/foo.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The async generator function should be named.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 31,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `foo`.',
              output: 'export default async function* foo () {}',
            },
          ],
        },
      ],
    },
    {
      code: 'export default async function *() {}',
      filename: '/path/to/foo.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The async generator function should be named.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 32,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `foo`.',
              output: 'export default async function * foo () {}',
            },
          ],
        },
      ],
    },
    {
      code: 'export default async function   *   () {}',
      filename: '/path/to/foo.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The async generator function should be named.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 37,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `foo`.',
              output: 'export default async function   *   foo () {}',
            },
          ],
        },
      ],
    },
    {
      code: 'export default async function * /* comment */ () {}',
      filename: '/path/to/foo.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The async generator function should be named.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 47,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `foo`.',
              output: 'export default async function * /* comment */ foo () {}',
            },
          ],
        },
      ],
    },
    {
      code: 'export default async function * // comment\n() {}',
      filename: '/path/to/foo.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The async generator function should be named.',
          line: 1,
          column: 16,
          endLine: 2,
          endColumn: 1,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `foo`.',
              output: 'export default async function * // comment\n foo () {}',
            },
          ],
        },
      ],
    },
    {
      code: 'let Foo, Foo_, foo, foo_\nexport default async function * () {}',
      filename: '/path/to/foo.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The async generator function should be named.',
          line: 2,
          column: 16,
          endLine: 2,
          endColumn: 33,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `foo__`.',
              output:
                'let Foo, Foo_, foo, foo_\nexport default async function * foo__ () {}',
            },
          ],
        },
      ],
    },
    {
      code: 'let Foo, Foo_, foo, foo_\nexport default (async function * () {})',
      filename: '/path/to/foo.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The async generator function should be named.',
          line: 2,
          column: 17,
          endLine: 2,
          endColumn: 34,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `foo__`.',
              output:
                'let Foo, Foo_, foo, foo_\nexport default (async function * foo__ () {})',
            },
          ],
        },
      ],
    },
    {
      code: 'let Exports, Exports_, exports, exports_\nexports = function() {}',
      filename: '/path/to/exports.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The function should be named.',
          line: 2,
          column: 11,
          endLine: 2,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `exports__`.',
              output:
                'let Exports, Exports_, exports, exports_\nexports = function exports__ () {}',
            },
          ],
        },
      ],
    },
    {
      code: 'module.exports = function() {}',
      filename: '/path/to/module.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The function should be named.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 26,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `module_`.',
              output: 'module.exports = function module_ () {}',
            },
          ],
        },
      ],
    },
    {
      code: 'export default () => {}',
      filename: '/path/to/foo.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The arrow function should be named.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `foo`.',
              output: 'const foo = () => {};\nexport default foo;',
            },
          ],
        },
      ],
    },
    {
      code: 'export default async () => {}',
      filename: '/path/to/foo.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The async arrow function should be named.',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 27,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `foo`.',
              output: 'const foo = async () => {};\nexport default foo;',
            },
          ],
        },
      ],
    },
    {
      code: 'export default () => {};',
      filename: '/path/to/foo.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The arrow function should be named.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `foo`.',
              output: 'const foo = () => {};\nexport default foo;',
            },
          ],
        },
      ],
    },
    {
      code: 'export default() => {}',
      filename: '/path/to/foo.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The arrow function should be named.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 20,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `foo`.',
              output: 'const foo = () => {};\nexport default foo;',
            },
          ],
        },
      ],
    },
    {
      code: 'export default foo => {}',
      filename: '/path/to/foo.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The arrow function should be named.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 22,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `foo_`.',
              output: 'const foo_ = foo => {};\nexport default foo_;',
            },
          ],
        },
      ],
    },
    {
      code: 'export default (( () => {} ))',
      filename: '/path/to/foo.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The arrow function should be named.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 24,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `foo`.',
              output: 'const foo = (( () => {} ));\nexport default foo;',
            },
          ],
        },
      ],
    },
    {
      code: '/* comment 1 */ export /* comment 2 */ default /* comment 3 */  () => {}',
      filename: '/path/to/foo.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The arrow function should be named.',
          line: 1,
          column: 68,
          endLine: 1,
          endColumn: 70,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `foo`.',
              output:
                '/* comment 1 */ const foo = () => {};\nexport /* comment 2 */ default /* comment 3 */  foo;',
            },
          ],
        },
      ],
    },
    {
      code: '// comment 1\nexport\n// comment 2\ndefault\n// comment 3\n() => {}',
      filename: '/path/to/foo.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The arrow function should be named.',
          line: 6,
          column: 4,
          endLine: 6,
          endColumn: 6,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `foo`.',
              output:
                '// comment 1\nconst foo = () => {};\nexport\n// comment 2\ndefault\n// comment 3\nfoo;',
            },
          ],
        },
      ],
    },
    {
      code: 'let Foo, Foo_, foo, foo_\nexport default async () => {}',
      filename: '/path/to/foo.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The async arrow function should be named.',
          line: 2,
          column: 25,
          endLine: 2,
          endColumn: 27,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `foo__`.',
              output:
                'let Foo, Foo_, foo, foo_\nconst foo__ = async () => {};\nexport default foo__;',
            },
          ],
        },
      ],
    },
    {
      code: 'let Exports, Exports_, exports, exports_\nexports = (( () => {} ))',
      filename: '/path/to/exports.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The arrow function should be named.',
          line: 2,
          column: 17,
          endLine: 2,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `exports__`.',
              output:
                'let Exports, Exports_, exports, exports_\nconst exports__ = (( () => {} ));\nexports = exports__;',
            },
          ],
        },
      ],
    },
    {
      code: '// comment 1\nmodule\n// comment 2\n.exports\n// comment 3\n=\n// comment 4\n() => {};',
      filename: '/path/to/module.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The arrow function should be named.',
          line: 8,
          column: 4,
          endLine: 8,
          endColumn: 6,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `module_`.',
              output:
                '// comment 1\nconst module_ = () => {};\nmodule\n// comment 2\n.exports\n// comment 3\n=\n// comment 4\nmodule_;',
            },
          ],
        },
      ],
    },
    {
      code: '(( exports = (( () => {} )) ))',
      filename: '/path/to/foo.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The arrow function should be named.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 22,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `foo`.',
              output: 'const foo = (( () => {} ));\n(( exports = foo ));',
            },
          ],
        },
      ],
    },
    {
      code: '(( module.exports = (( () => {} )) ))',
      filename: '/path/to/foo.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The arrow function should be named.',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 29,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `foo`.',
              output:
                'const foo = (( () => {} ));\n(( module.exports = foo ));',
            },
          ],
        },
      ],
    },
    {
      code: '(( exports = (( () => {} )) ));',
      filename: '/path/to/foo.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The arrow function should be named.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 22,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `foo`.',
              output: 'const foo = (( () => {} ));\n(( exports = foo ));',
            },
          ],
        },
      ],
    },
    {
      code: '(( module.exports = (( () => {} )) ));',
      filename: '/path/to/foo.js',
      origin: 'Upstream snapshot cases',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The arrow function should be named.',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 29,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `foo`.',
              output:
                'const foo = (( () => {} ));\n(( module.exports = foo ));',
            },
          ],
        },
      ],
    },
    {
      code: 'export default () => {\r\n\tbar();\r\n};\r\n',
      filename: '/path/to/foo.js',
      origin: 'Upstream line endings',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The arrow function should be named.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `foo`.',
              output:
                'const foo = () => {\r\n\tbar();\r\n};\r\nexport default foo;\r\n',
            },
          ],
        },
      ],
    },
    {
      code: 'export default class {}',
      filename: '/path/to/foo.js',
      origin: 'Documentation examples',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The class should be named.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `Foo`.',
              output: 'export default class Foo {}',
            },
          ],
        },
      ],
    },
    {
      code: 'export default function () {}',
      filename: '/path/to/foo.js',
      origin: 'Documentation examples',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The function should be named.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 25,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `foo`.',
              output: 'export default function foo () {}',
            },
          ],
        },
      ],
    },
    {
      code: 'export default () => {};',
      filename: '/path/to/foo.js',
      origin: 'Documentation examples',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The arrow function should be named.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `foo`.',
              output: 'const foo = () => {};\nexport default foo;',
            },
          ],
        },
      ],
    },
    {
      code: 'module.exports = class {};',
      filename: '/path/to/foo.js',
      origin: 'Documentation examples',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The class should be named.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 23,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `Foo`.',
              output: 'module.exports = class Foo {};',
            },
          ],
        },
      ],
    },
    {
      code: 'module.exports = function () {};',
      filename: '/path/to/foo.js',
      origin: 'Documentation examples',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The function should be named.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 27,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `foo`.',
              output: 'module.exports = function foo () {};',
            },
          ],
        },
      ],
    },
    {
      code: 'module.exports = () => {};',
      filename: '/path/to/foo.js',
      origin: 'Documentation examples',
      errors: [
        {
          messageId: 'no-anonymous-default-export/error',
          message: 'The arrow function should be named.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 23,
          suggestions: [
            {
              messageId: 'no-anonymous-default-export/suggestion',
              message: 'Name it as `foo`.',
              output: 'const foo = () => {};\nmodule.exports = foo;',
            },
          ],
        },
      ],
    },
  ],
};

const directory = path.resolve(import.meta.dirname, '..');
const ruleName = 'unicorn/no-anonymous-default-export';
const run = (code: string, filename: string, fix = false) =>
  lint({
    config: [{ plugins: ['unicorn'], rules: { [ruleName]: 'error' } }],
    configDirectory: directory,
    workingDirectory: directory,
    fileContents: { [filename]: code },
    fix,
  });

describe(ruleName, () => {
  test('upstream valid cases and documentation', async () => {
    for (const item of cases.valid) {
      const result = await run(item.code, item.filename);
      expect(result.fileCount).toBe(1);
      expect(result.ruleCount).toBe(1);
      expect(result.diagnostics).toEqual([]);
    }
  });
  test('upstream diagnostics and suggestions', async () => {
    for (const item of cases.invalid) {
      const result = await run(item.code, item.filename);
      expect(result.fileCount).toBe(1);
      expect(result.ruleCount).toBe(1);
      expect(result.diagnostics).toHaveLength(item.errors.length);
      for (const [index, expected] of item.errors.entries()) {
        const diagnostic = result.diagnostics[index];
        expect(diagnostic.ruleName).toBe(ruleName);
        expect(diagnostic.messageId).toBe(expected.messageId);
        expect(diagnostic.message).toBe(expected.message);
        expect(diagnostic.range).toEqual({
          start: { line: expected.line, column: expected.column },
          end: { line: expected.endLine, column: expected.endColumn },
        });
        expect(diagnostic.fixes ?? []).toEqual([]);
        const suggestions = diagnostic.suggestions ?? [];
        expect(suggestions).toHaveLength(expected.suggestions.length);
        for (const [
          suggestionIndex,
          wanted,
        ] of expected.suggestions.entries()) {
          const suggestion = suggestions[suggestionIndex];
          expect(suggestion.messageId).toBe(wanted.messageId);
          expect(suggestion.message).toBe(wanted.message);
          let output = '';
          let offset = 0;
          for (const edit of [...(suggestion.fixes ?? [])].sort(
            (a, b) => a.startPos - b.startPos || a.endPos - b.endPos,
          )) {
            output += item.code.slice(offset, edit.startPos) + edit.text;
            offset = edit.endPos;
          }
          output += item.code.slice(offset);
          expect(output).toBe(wanted.output);
          expect((await run(output, item.filename)).diagnostics).toEqual([]);
        }
      }
    }
  });
  test('suggestions do not become automatic fixes', async () => {
    const result = await run(
      'export default () => {};',
      '/path/to/foo.js',
      true,
    );
    expect(result.output ?? {}).toEqual({});
    expect(result.fixableErrorCount).toBe(0);
  });
});
