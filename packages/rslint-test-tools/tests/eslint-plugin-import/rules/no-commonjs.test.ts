// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
new RuleTester(fixtureConfig()).run('no-commonjs', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: 'import "x";',
    },
    {
      name: 'run 0 valid 1',
      code: 'import x from "x"',
    },
    {
      name: 'run 0 valid 2',
      code: 'import { x } from "x"',
    },
    {
      name: 'run 0 valid 3',
      code: 'export default "x"',
    },
    {
      name: 'run 0 valid 4',
      code: 'export function house() {}',
    },
    {
      name: 'run 0 valid 5',
      code: "\n        function someFunc() {\n          const exports = someComputation();\n          expect(exports.someProp).toEqual({ a: 'value' });\n        }\n      ",
    },
    {
      name: 'run 0 valid 6',
      code: 'function a() { var x = require("y"); }',
    },
    {
      name: 'run 0 valid 7',
      code: 'var a = c && require("b")',
    },
    {
      name: 'run 0 valid 8',
      code: 'require.resolve("help")',
    },
    {
      name: 'run 0 valid 9',
      code: 'require.ensure([])',
    },
    {
      name: 'run 0 valid 10',
      code: 'require([], function(a, b, c) {})',
    },
    {
      name: 'run 0 valid 11',
      code: "var bar = require('./bar', true);",
    },
    {
      name: 'run 0 valid 12',
      code: "var bar = proxyquire('./bar');",
    },
    {
      name: 'run 0 valid 13',
      code: "var bar = require('./ba' + 'r');",
    },
    {
      name: 'run 0 valid 14',
      code: 'var bar = require(`x${1}`);',
    },
    {
      name: 'run 0 valid 15',
      code: 'var zero = require(0);',
    },
    {
      name: 'run 0 valid 16',
      code: 'require("x")',
      options: [
        {
          allowRequire: true,
        },
      ],
    },
    {
      name: 'run 0 valid 17',
      code: 'require(rootRequire("x"))',
      options: [
        {
          allowRequire: true,
        },
      ],
    },
    {
      name: 'run 0 valid 18',
      code: 'require(String("x"))',
      options: [
        {
          allowRequire: true,
        },
      ],
    },
    {
      name: 'run 0 valid 19',
      code: 'require(["x", "y", "z"].join("/"))',
      options: [
        {
          allowRequire: true,
        },
      ],
    },
    {
      name: 'run 0 valid 20',
      code: 'rootRequire("x")',
      options: [
        {
          allowRequire: true,
        },
      ],
    },
    {
      name: 'run 0 valid 21',
      code: 'rootRequire("x")',
      options: [
        {
          allowRequire: false,
        },
      ],
    },
    {
      name: 'run 0 valid 22',
      code: 'module.exports = function () {}',
      options: ['allow-primitive-modules'],
    },
    {
      name: 'run 0 valid 23',
      code: 'module.exports = function () {}',
      options: [
        {
          allowPrimitiveModules: true,
        },
      ],
    },
    {
      name: 'run 0 valid 24',
      code: 'module.exports = "foo"',
      options: ['allow-primitive-modules'],
    },
    {
      name: 'run 0 valid 25',
      code: 'module.exports = "foo"',
      options: [
        {
          allowPrimitiveModules: true,
        },
      ],
    },
    {
      name: 'run 0 valid 26',
      code: 'if (typeof window !== "undefined") require("x")',
      options: [
        {
          allowRequire: true,
        },
      ],
    },
    {
      name: 'run 0 valid 27',
      code: 'if (typeof window !== "undefined") require("x")',
      options: [
        {
          allowRequire: false,
        },
      ],
    },
    {
      name: 'run 0 valid 28',
      code: 'if (typeof window !== "undefined") { require("x") }',
      options: [
        {
          allowRequire: true,
        },
      ],
    },
    {
      name: 'run 0 valid 29',
      code: 'if (typeof window !== "undefined") { require("x") }',
      options: [
        {
          allowRequire: false,
        },
      ],
    },
    {
      name: 'run 0 valid 30',
      code: 'try { require("x") } catch (error) {}',
    },
    {
      name: 'run 1 valid 0',
      code: "/*eslint no-commonjs: [2, { allowRequire: true }]*/\nvar mod = require('./mod');",
      options: [
        {
          allowRequire: true,
        },
      ],
    },
    {
      name: 'run 1 valid 1',
      code: 'var a = b && require("c")\n\nif (typeof window !== "undefined") {\n  require(\'that-ugly-thing\');\n}\n\nvar fs = null;\ntry {\n  fs = require("fs")\n} catch (error) {}',
    },
    {
      name: 'run 1 valid 2',
      code: '/*eslint no-commonjs: [2, { allowPrimitiveModules: true }]*/\n\nmodule.exports = "foo"\nmodule.exports = function rule(context) { return { /* ... */ } }',
      options: [
        {
          allowPrimitiveModules: true,
        },
      ],
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: 'var x = require("x")',
      errors: [
        {
          messageId: '',
          message: 'Expected "import" instead of "require()"',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      name: 'run 0 invalid 1',
      code: 'x = require("x")',
      errors: [
        {
          messageId: '',
          message: 'Expected "import" instead of "require()"',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      name: 'run 0 invalid 2',
      code: 'require("x")',
      errors: [
        {
          messageId: '',
          message: 'Expected "import" instead of "require()"',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      name: 'run 0 invalid 3',
      code: 'require(`x`)',
      errors: [
        {
          messageId: '',
          message: 'Expected "import" instead of "require()"',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      name: 'run 0 invalid 4',
      code: 'if (typeof window !== "undefined") require("x")',
      options: [
        {
          allowConditionalRequire: false,
        },
      ],
      errors: [
        {
          messageId: '',
          message: 'Expected "import" instead of "require()"',
          line: 1,
          column: 36,
          endLine: 1,
          endColumn: 43,
        },
      ],
    },
    {
      name: 'run 0 invalid 5',
      code: 'if (typeof window !== "undefined") { require("x") }',
      options: [
        {
          allowConditionalRequire: false,
        },
      ],
      errors: [
        {
          messageId: '',
          message: 'Expected "import" instead of "require()"',
          line: 1,
          column: 38,
          endLine: 1,
          endColumn: 45,
        },
      ],
    },
    {
      name: 'run 0 invalid 6',
      code: 'try { require("x") } catch (error) {}',
      options: [
        {
          allowConditionalRequire: false,
        },
      ],
      errors: [
        {
          messageId: '',
          message: 'Expected "import" instead of "require()"',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      name: 'run 0 invalid 7',
      code: 'exports.face = "palm"',
      errors: [
        {
          messageId: '',
          message: 'Expected "export" or "export default"',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      name: 'run 0 invalid 8',
      code: 'module.exports.face = "palm"',
      errors: [
        {
          messageId: '',
          message: 'Expected "export" or "export default"',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      name: 'run 0 invalid 9',
      code: 'module.exports = face',
      errors: [
        {
          messageId: '',
          message: 'Expected "export" or "export default"',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      name: 'run 0 invalid 10',
      code: 'exports = module.exports = {}',
      errors: [
        {
          messageId: '',
          message: 'Expected "export" or "export default"',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      name: 'run 0 invalid 11',
      code: 'var x = module.exports = {}',
      errors: [
        {
          messageId: '',
          message: 'Expected "export" or "export default"',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      name: 'run 0 invalid 12',
      code: 'module.exports = {}',
      options: ['allow-primitive-modules'],
      errors: [
        {
          messageId: '',
          message: 'Expected "export" or "export default"',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      name: 'run 0 invalid 13',
      code: 'var x = module.exports',
      options: ['allow-primitive-modules'],
      errors: [
        {
          messageId: '',
          message: 'Expected "export" or "export default"',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      name: 'run 1 invalid 0',
      code: "var mod = require('./mod')\n  , common = require('./common')\n  , fs = require('fs')\n  , whateverModule = require('./not-found')\n\nmodule.exports = { a: \"b\" }\nexports.c = \"d\"",
      errors: [
        {
          messageId: '',
          message: 'Expected "import" instead of "require()"',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 18,
        },
        {
          messageId: '',
          message: 'Expected "import" instead of "require()"',
          line: 2,
          column: 14,
          endLine: 2,
          endColumn: 21,
        },
        {
          messageId: '',
          message: 'Expected "import" instead of "require()"',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 17,
        },
        {
          messageId: '',
          message: 'Expected "import" instead of "require()"',
          line: 4,
          column: 22,
          endLine: 4,
          endColumn: 29,
        },
        {
          messageId: '',
          message: 'Expected "export" or "export default"',
          line: 6,
          column: 1,
          endLine: 6,
          endColumn: 15,
        },
        {
          messageId: '',
          message: 'Expected "export" or "export default"',
          line: 7,
          column: 1,
          endLine: 7,
          endColumn: 10,
        },
      ],
    },
    {
      name: 'run 1 invalid 1',
      code: 'var a = b && require("c")\n\nif (typeof window !== "undefined") {\n  require(\'that-ugly-thing\');\n}\n\nvar fs = null;\ntry {\n  fs = require("fs")\n} catch (error) {}',
      options: [
        {
          allowConditionalRequire: false,
        },
      ],
      errors: [
        {
          messageId: '',
          message: 'Expected "import" instead of "require()"',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 21,
        },
        {
          messageId: '',
          message: 'Expected "import" instead of "require()"',
          line: 4,
          column: 3,
          endLine: 4,
          endColumn: 10,
        },
        {
          messageId: '',
          message: 'Expected "import" instead of "require()"',
          line: 9,
          column: 8,
          endLine: 9,
          endColumn: 15,
        },
      ],
    },
    {
      name: 'run 1 invalid 2',
      code: '/*eslint no-commonjs: [2, { allowPrimitiveModules: true }]*/\n\nmodule.exports = { x: "y" }\nexports.z = function boop() { /* ... */ }',
      options: [
        {
          allowPrimitiveModules: true,
        },
      ],
      errors: [
        {
          messageId: '',
          message: 'Expected "export" or "export default"',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 15,
        },
        {
          messageId: '',
          message: 'Expected "export" or "export default"',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 10,
        },
      ],
    },
  ],
});
