// Upstream: https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/no-commonjs.js
import { RuleTester } from '../rule-tester.js';

const ruleTester = new RuleTester();

ruleTester.run('no-commonjs', null as never, {
  valid: [
    // Upstream valid: ES imports.
    {
      code: 'import "x";',
    },
    {
      code: 'import x from "x"',
    },
    {
      code: 'import { x } from "x"',
    },
    // Upstream valid: ES exports and a local exports binding.
    {
      code: 'export default "x"',
    },
    {
      code: 'export function house() {}',
    },
    {
      code: "\n        function someFunc() {\n          const exports = someComputation();\n          expect(exports.someProp).toEqual({ a: 'value' });\n        }\n      ",
    },
    // Upstream valid: allowed requires and option exceptions.
    {
      code: 'function a() { var x = require("y"); }',
    },
    {
      code: 'var a = c && require("b")',
    },
    {
      code: 'require.resolve("help")',
    },
    {
      code: 'require.ensure([])',
    },
    {
      code: 'require([], function(a, b, c) {})',
    },
    {
      code: "var bar = require('./bar', true);",
    },
    {
      code: "var bar = proxyquire('./bar');",
    },
    {
      code: "var bar = require('./ba' + 'r');",
    },
    {
      code: 'var bar = require(`x${1}`);',
    },
    {
      code: 'var zero = require(0);',
    },
    {
      code: 'require("x")',
      options: [
        {
          allowRequire: true,
        },
      ],
    },
    {
      code: 'require(rootRequire("x"))',
      options: [
        {
          allowRequire: true,
        },
      ],
    },
    {
      code: 'require(String("x"))',
      options: [
        {
          allowRequire: true,
        },
      ],
    },
    {
      code: 'require(["x", "y", "z"].join("/"))',
      options: [
        {
          allowRequire: true,
        },
      ],
    },
    {
      code: 'rootRequire("x")',
      options: [
        {
          allowRequire: true,
        },
      ],
    },
    {
      code: 'rootRequire("x")',
      options: [
        {
          allowRequire: false,
        },
      ],
    },
    {
      code: 'module.exports = function () {}',
      options: ['allow-primitive-modules'],
    },
    {
      code: 'module.exports = function () {}',
      options: [
        {
          allowPrimitiveModules: true,
        },
      ],
    },
    {
      code: 'module.exports = "foo"',
      options: ['allow-primitive-modules'],
    },
    {
      code: 'module.exports = "foo"',
      options: [
        {
          allowPrimitiveModules: true,
        },
      ],
    },
    {
      code: 'if (typeof window !== "undefined") require("x")',
      options: [
        {
          allowRequire: true,
        },
      ],
    },
    {
      code: 'if (typeof window !== "undefined") require("x")',
      options: [
        {
          allowRequire: false,
        },
      ],
    },
    {
      code: 'if (typeof window !== "undefined") { require("x") }',
      options: [
        {
          allowRequire: true,
        },
      ],
    },
    {
      code: 'if (typeof window !== "undefined") { require("x") }',
      options: [
        {
          allowRequire: false,
        },
      ],
    },
    {
      code: 'try { require("x") } catch (error) {}',
    },
  ],
  invalid: [
    // Upstream invalid: require calls (the ESLint >= 4 branch).
    {
      code: 'var x = require("x")',
      errors: [
        {
          message: 'Expected "import" instead of "require()"',
        },
      ],
      output: null,
    },
    {
      code: 'x = require("x")',
      errors: [
        {
          message: 'Expected "import" instead of "require()"',
        },
      ],
      output: null,
    },
    {
      code: 'require("x")',
      errors: [
        {
          message: 'Expected "import" instead of "require()"',
        },
      ],
      output: null,
    },
    {
      code: 'require(`x`)',
      errors: [
        {
          message: 'Expected "import" instead of "require()"',
        },
      ],
      output: null,
    },
    {
      code: 'if (typeof window !== "undefined") require("x")',
      options: [
        {
          allowConditionalRequire: false,
        },
      ],
      errors: [
        {
          message: 'Expected "import" instead of "require()"',
        },
      ],
      output: null,
    },
    {
      code: 'if (typeof window !== "undefined") { require("x") }',
      options: [
        {
          allowConditionalRequire: false,
        },
      ],
      errors: [
        {
          message: 'Expected "import" instead of "require()"',
        },
      ],
      output: null,
    },
    {
      code: 'try { require("x") } catch (error) {}',
      options: [
        {
          allowConditionalRequire: false,
        },
      ],
      errors: [
        {
          message: 'Expected "import" instead of "require()"',
        },
      ],
      output: null,
    },
    // Upstream invalid: CommonJS exports.
    {
      code: 'exports.face = "palm"',
      errors: [
        {
          message: 'Expected "export" or "export default"',
        },
      ],
      output: null,
    },
    {
      code: 'module.exports.face = "palm"',
      errors: [
        {
          message: 'Expected "export" or "export default"',
        },
      ],
      output: null,
    },
    {
      code: 'module.exports = face',
      errors: [
        {
          message: 'Expected "export" or "export default"',
        },
      ],
      output: null,
    },
    {
      code: 'exports = module.exports = {}',
      errors: [
        {
          message: 'Expected "export" or "export default"',
        },
      ],
      output: null,
    },
    {
      code: 'var x = module.exports = {}',
      errors: [
        {
          message: 'Expected "export" or "export default"',
        },
      ],
      output: null,
    },
    {
      code: 'module.exports = {}',
      options: ['allow-primitive-modules'],
      errors: [
        {
          message: 'Expected "export" or "export default"',
        },
      ],
      output: null,
    },
    {
      code: 'var x = module.exports',
      options: ['allow-primitive-modules'],
      errors: [
        {
          message: 'Expected "export" or "export default"',
        },
      ],
      output: null,
    },
  ],
});

// Examples from the pinned upstream documentation.
ruleTester.run('no-commonjs', null as never, {
  valid: [
    {
      code: "/*eslint no-commonjs: [2, { allowRequire: true }]*/\nvar mod = require('./mod');",
      options: [
        {
          allowRequire: true,
        },
      ],
    },
    {
      code: 'var a = b && require("c")\n\nif (typeof window !== "undefined") {\n  require(\'that-ugly-thing\');\n}\n\nvar fs = null;\ntry {\n  fs = require("fs")\n} catch (error) {}',
    },
    {
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
      code: "var mod = require('./mod')\n  , common = require('./common')\n  , fs = require('fs')\n  , whateverModule = require('./not-found')\n\nmodule.exports = { a: \"b\" }\nexports.c = \"d\"",
      errors: [
        {
          message: 'Expected "import" instead of "require()"',
        },
        {
          message: 'Expected "import" instead of "require()"',
        },
        {
          message: 'Expected "import" instead of "require()"',
        },
        {
          message: 'Expected "import" instead of "require()"',
        },
        {
          message: 'Expected "export" or "export default"',
        },
        {
          message: 'Expected "export" or "export default"',
        },
      ],
      output: null,
    },
    {
      code: 'var a = b && require("c")\n\nif (typeof window !== "undefined") {\n  require(\'that-ugly-thing\');\n}\n\nvar fs = null;\ntry {\n  fs = require("fs")\n} catch (error) {}',
      options: [
        {
          allowConditionalRequire: false,
        },
      ],
      errors: [
        {
          message: 'Expected "import" instead of "require()"',
        },
        {
          message: 'Expected "import" instead of "require()"',
        },
        {
          message: 'Expected "import" instead of "require()"',
        },
      ],
      output: null,
    },
    {
      code: '/*eslint no-commonjs: [2, { allowPrimitiveModules: true }]*/\n\nmodule.exports = { x: "y" }\nexports.z = function boop() { /* ... */ }',
      options: [
        {
          allowPrimitiveModules: true,
        },
      ],
      errors: [
        {
          message: 'Expected "export" or "export default"',
        },
        {
          message: 'Expected "export" or "export default"',
        },
      ],
      output: null,
    },
  ],
});
