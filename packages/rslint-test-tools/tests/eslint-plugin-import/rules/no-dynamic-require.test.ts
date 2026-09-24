// Upstream: https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/no-dynamic-require.js
// Documentation: https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-dynamic-require.md
import { RuleTester } from '../rule-tester.js';

const ruleTester = new RuleTester();
const requireMessage = 'Calls to require() should use string literals';
const importMessage = 'Calls to import() should use string literals';

ruleTester.run('no-dynamic-require', null as never, {
  valid: [
    // Upstream valid: imports and require calls.
    {
      code: 'import _ from "lodash"',
    },
    {
      code: 'require("foo")',
    },
    {
      code: 'require(`foo`)',
    },
    {
      code: 'require("./foo")',
    },
    {
      code: 'require("@scope/foo")',
    },
    {
      code: 'require()',
    },
    {
      code: 'require("./foo", "bar" + "okay")',
    },
    {
      code: 'var foo = require("foo")',
    },
    {
      code: 'var foo = require(`foo`)',
    },
    {
      code: 'var foo = require("./foo")',
    },
    {
      code: 'var foo = require("@scope/foo")',
    },
    // Upstream dynamic imports: Espree and Babel repeat these same cases.
    {
      code: 'import("foo")',
      options: [{ esmodule: true }],
    },
    {
      code: 'import(`foo`)',
      options: [{ esmodule: true }],
    },
    {
      code: 'import("./foo")',
      options: [{ esmodule: true }],
    },
    {
      code: 'import("@scope/foo")',
      options: [{ esmodule: true }],
    },
    {
      code: 'var foo = import("foo")',
      options: [{ esmodule: true }],
    },
    {
      code: 'var foo = import(`foo`)',
      options: [{ esmodule: true }],
    },
    {
      code: 'var foo = import("./foo")',
      options: [{ esmodule: true }],
    },
    {
      code: 'var foo = import("@scope/foo")',
      options: [{ esmodule: true }],
    },
    // Upstream dynamic imports are allowed when esmodule is omitted.
    {
      code: 'import("../" + name)',
    },
    {
      code: 'import(`../${name}`)',
    },
    // Examples from the pinned upstream documentation.
    {
      code: "require('../name');\nrequire(`../name`);",
    },
  ],
  invalid: [
    // Upstream invalid: require calls.
    {
      code: 'require("../" + name)',
      errors: [requireMessage],
    },
    {
      code: 'require(`../${name}`)',
      errors: [requireMessage],
    },
    {
      code: 'require(name)',
      errors: [requireMessage],
    },
    {
      code: 'require(name())',
      errors: [requireMessage],
    },
    {
      code: 'require(name + "foo", "bar")',
      options: [{ esmodule: true }],
      errors: [requireMessage],
    },
    // Upstream dynamic imports: both parser variants have the same semantics.
    {
      code: 'import("../" + name)',
      options: [{ esmodule: true }],
      errors: [importMessage],
    },
    {
      code: 'import(`../${name}`)',
      options: [{ esmodule: true }],
      errors: [importMessage],
    },
    {
      code: 'import(name)',
      options: [{ esmodule: true }],
      errors: [importMessage],
    },
    {
      code: 'import(name())',
      options: [{ esmodule: true }],
      errors: [importMessage],
    },
    // Upstream interpolated require arguments.
    {
      code: 'require(`foo${x}`)',
      errors: [requireMessage],
    },
    {
      code: 'var foo = require(`foo${x}`)',
      errors: [requireMessage],
    },
    // Examples from the pinned upstream documentation.
    {
      code: "require(name);\nrequire('../' + name);\nrequire(`../${name}`);\nrequire(name());",
      errors: [requireMessage, requireMessage, requireMessage, requireMessage],
    },
  ],
});
