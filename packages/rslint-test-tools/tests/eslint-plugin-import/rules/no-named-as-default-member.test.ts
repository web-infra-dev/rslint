import { RuleTester } from '../rule-tester.js';
import { testFixturePath } from '../utils.js';

// Every upstream test and documentation example from eslint-plugin-import v2.32.0,
// including SYNTAX_CASES. Complete diagnostic ranges and IDs are asserted in Go;
// this wrapper checks messages and counts. No upstream cases are skipped.
const filename = testFixturePath('no-named-as-default-member-rule/consumer.ts');

new RuleTester().run('no-named-as-default-member', null as never, {
  valid: [
    // upstream
    { code: 'import bar, {foo} from "./bar";' },
    { code: 'import bar from "./bar"; const baz = bar.baz' },
    { code: 'import {foo} from "./bar"; const baz = foo.baz;' },
    { code: 'import * as named from "./named-exports"; const a = named.a' },
    {
      code: 'import foo from "./default-export-default-property"; const a = foo.default',
    },
    // Arbitrary module namespace identifier names
    { code: 'import bar, { foo } from "./export-default-string-and-named"' },
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
    // Documentation module and valid consumer
    { code: "export default 'foo'; export const bar = 'baz';" },
    { code: "import foo, {bar} from './foo.js';" },
  ].map((test) => ({ ...test, filename })),
  invalid: [
    // upstream
    {
      code: 'import bar from "./bar"; const foo = bar.foo;',
      errors: [
        "Caution: `bar` also has a named export `foo`. Check if you meant to write `import {foo} from './bar'` instead.",
      ],
    },
    {
      code: 'import bar from "./bar"; bar.foo();',
      errors: [
        "Caution: `bar` also has a named export `foo`. Check if you meant to write `import {foo} from './bar'` instead.",
      ],
    },
    {
      code: 'import bar from "./bar"; const {foo} = bar;',
      errors: [
        "Caution: `bar` also has a named export `foo`. Check if you meant to write `import {foo} from './bar'` instead.",
      ],
    },
    {
      code: 'import bar from "./bar"; const {foo: foo2, baz} = bar;',
      errors: [
        "Caution: `bar` also has a named export `foo`. Check if you meant to write `import {foo} from './bar'` instead.",
      ],
    },
    // Arbitrary module namespace identifier names
    {
      code: 'import bar from "./export-default-string-and-named"; const foo = bar.foo;',
      errors: [
        "Caution: `bar` also has a named export `foo`. Check if you meant to write `import {foo} from './export-default-string-and-named'` instead.",
      ],
    },
    // Documentation member access
    {
      code: "import foo from './foo.js';\nconst bar = foo.bar;",
      errors: [
        "Caution: `foo` also has a named export `bar`. Check if you meant to write `import {bar} from './foo.js'` instead.",
      ],
    },
    // Documentation destructuring
    {
      code: "import foo from './foo.js';\nconst {bar} = foo;",
      errors: [
        "Caution: `foo` also has a named export `bar`. Check if you meant to write `import {bar} from './foo.js'` instead.",
      ],
    },
  ].map((test) => ({ ...test, filename })),
});
