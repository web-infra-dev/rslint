import { RuleTester } from '../rule-tester.js';
import { testFixturePath } from '../utils.js';

// Upstream: eslint-plugin-import v2.32.0 tests/src/rules/no-named-as-default.js
// and docs/rules/no-named-as-default.md, including expanded SYNTAX_CASES.
// This wrapper checks messages; complete ranges and IDs are asserted in Go.
const filename = testFixturePath('no-named-as-default-rule/consumer.ts');

new RuleTester().run('no-named-as-default', null as never, {
  valid: [
    // The test Program rejects the malformed dependency before rules run.
    // Skipped: import "./malformed.js"
    { code: 'import bar, { foo } from "./bar";' },
    { code: 'import bar, { foo } from "./empty-folder";' },
    // Babel's export-default-from proposal is not supported by tsgo.
    // Skipped: export bar, { foo } from "./bar";
    // Skipped: export bar from "./bar";
    // Skipped: export default from "./bar";
    { code: 'import bar, { foo } from "./export-default-string-and-named"' },
    // #1594: the default and named exports are direct re-exports.
    { code: 'import something from "./no-named-as-default/re-exports.js";' },
    {
      code: 'import { something } from "./no-named-as-default/re-exports.js";',
    },
    {
      code: 'import myOwnNameForVariable from "./no-named-as-default/exports.js";',
    },
    { code: 'import { variable } from "./no-named-as-default/exports.js";' },
    {
      code: 'import variable from "./no-named-as-default/misleading-re-exports.js";',
    },
    {
      code: 'import foobar from "./no-named-as-default/no-default-export.js";',
    },
    // Same upstream cases for exports; only proposal syntax is skipped.
    // Skipped: export something from "./no-named-as-default/re-exports.js";
    {
      code: 'export { something } from "./no-named-as-default/re-exports.js";',
    },
    // Skipped: export myOwnNameForVariable from "./no-named-as-default/exports.js";
    { code: 'export { variable } from "./no-named-as-default/exports.js";' },
    // Skipped: export variable from "./no-named-as-default/misleading-re-exports.js";
    // Skipped: export foobar from "./no-named-as-default/no-default-export.js";
    // SYNTAX_CASES
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
    // Documentation examples (the module, then its consumers).
    { code: "export default 'foo'; export const bar = 'baz';" },
    { code: "import foo from './foo.js';" },
    // Skipped: export foo from './foo.js';
  ].map((test) => ({ ...test, filename })),
  invalid: [
    {
      code: 'import foo from "./bar";',
      errors: [
        {
          message:
            "Using exported name 'foo' as identifier for default import.",
        },
      ],
    },
    {
      code: 'import foo, { foo as bar } from "./bar";',
      errors: [
        {
          message:
            "Using exported name 'foo' as identifier for default import.",
        },
      ],
    },
    // Babel's export-default-from proposal is not supported by tsgo.
    // Skipped: export foo from "./bar";
    // Skipped: export foo, { foo as bar } from "./bar";
    // Dependency parse errors are not converted into rule diagnostics.
    // Skipped: import foo from "./malformed.js"
    {
      code: 'import foo from "./export-default-string-and-named"',
      errors: [
        {
          message:
            "Using exported name 'foo' as identifier for default import.",
        },
      ],
    },
    {
      code: 'import foo, { foo as bar } from "./export-default-string-and-named"',
      errors: [
        {
          message:
            "Using exported name 'foo' as identifier for default import.",
        },
      ],
    },
    {
      code: 'import something from "./no-named-as-default/misleading-re-exports.js";',
      errors: [
        {
          message:
            "Using exported name 'something' as identifier for default import.",
        },
      ],
    },
    // Upstream still reports when the same value is exported locally.
    {
      code: 'import variable from "./no-named-as-default/exports.js";',
      errors: [
        {
          message:
            "Using exported name 'variable' as identifier for default import.",
        },
      ],
    },
    // Skipped: export variable from "./no-named-as-default/exports.js";
    // The docs' import example labels this an export; the rule says import.
    {
      code: "import bar from './foo.js';",
      errors: [
        {
          message:
            "Using exported name 'bar' as identifier for default import.",
        },
      ],
    },
    // Skipped: export bar from './foo.js';
  ].map((test) => ({ ...test, filename })),
});
