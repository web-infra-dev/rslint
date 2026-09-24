import { RuleTester } from '../rule-tester.js';

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/no-named-default.js
const ruleTester = new RuleTester();

ruleTester.run('no-named-default', null as never, {
  valid: [
    { code: 'import bar from "./bar";' },
    { code: 'import bar, { foo } from "./bar";' },
    // Upstream's Flow type case also parses as TypeScript.
    { code: 'import { type default as Foo } from "./bar";' },
    // Skipped: tsgo does not parse Flow's typeof import specifiers.
    // { code: 'import { typeof default as Foo } from "./bar";' },

    // SYNTAX_CASES from tests/src/utils.js at the same tag.
    { code: 'for (let { foo, bar } of baz) {}' },
    { code: 'for (let [ foo, bar ] of baz) {}' },
    { code: 'const { x, y } = bar' },
    { code: 'const { x, y, ...z } = bar' },
    { code: 'let x; export { x }' },
    { code: 'let x; export { x as y }' },
    { code: 'export const x = null' },
    { code: 'export var x = null' },
    { code: 'export let x = null' },
    { code: 'export default x' },
    { code: 'export default class x {}' },
    {
      code: 'import json from "./data.json"',
      settings: { 'import/extensions': ['.js'] },
    },
    {
      code: 'import foo from "./foobar.json";',
      settings: { 'import/extensions': ['.js'] },
    },
    {
      code: 'import foo from "./foobar";',
      settings: { 'import/extensions': ['.js'] },
    },
    {
      code: 'import { foo } from "./issue-370-commonjs-namespace/bar"',
      settings: { 'import/ignore': ['foo'] },
    },
    {
      code: 'export * from "./issue-370-commonjs-namespace/bar"',
      settings: { 'import/ignore': ['foo'] },
    },
    { code: 'import * as a from "./commonjs-namespace/a"; a.b' },
    { code: 'import { foo } from "./ignore.invalid.extension"' },

    // https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-named-default.md
    { code: "// foo.js\nexport default 'foo';\nexport const bar = 'baz';" },
    { code: "import foo from './foo.js';" },
    { code: "import foo, { bar } from './foo.js';" },
  ],
  invalid: [
    {
      code: 'import { default as bar } from "./bar";',
      errors: ["Use default import syntax to import 'bar'."],
    },
    {
      code: 'import { foo, default as bar } from "./bar";',
      errors: ["Use default import syntax to import 'bar'."],
    },
    {
      code: 'import { "default" as bar } from "./bar";',
      errors: ["Use default import syntax to import 'bar'."],
    },
    // Documentation examples; use the source's message, not the stale doc comment.
    {
      code: "import { default as foo } from './foo.js';",
      errors: ["Use default import syntax to import 'foo'."],
    },
    {
      code: "import { default as foo, bar } from './foo.js';",
      errors: ["Use default import syntax to import 'foo'."],
    },
  ],
});
