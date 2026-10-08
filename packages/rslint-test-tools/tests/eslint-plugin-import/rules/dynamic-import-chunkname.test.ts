// Upstream: https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/dynamic-import-chunkname.js
// Documentation: https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/dynamic-import-chunkname.md
// This wrapper checks diagnostic counts and messages. Go tests additionally
// assert complete ranges and suggestion outputs.
import { RuleTester } from '../rule-tester.js';

const ruleTester = new RuleTester();

ruleTester.run('dynamic-import-chunkname', null as never, {
  valid: [
    // ---- upstream valid: JavaScript ----
    {
      code: `dynamicImport(
        /* webpackChunkName: "someModule" */
        'test'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `dynamicImport(
        /* webpackChunkName: "Some_Other_Module" */
        "test"
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `dynamicImport(
        /* webpackChunkName: "SomeModule123" */
        "test"
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `dynamicImport(
        /* webpackChunkName: "someModule" */
        'someModule'
      )`,
      options: [
        {
          importFunctions: ['dynamicImport'],
          webpackChunknameFormat: '[a-zA-Z-_/.]+',
        },
      ],
    },
    {
      code: `dynamicImport(
        /* webpackChunkName: "[request]" */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `dynamicImport(
        /* webpackChunkName: "my-chunk-[request]-custom" */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `dynamicImport(
        /* webpackChunkName: '[index]' */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `dynamicImport(
        /* webpackChunkName: 'my-chunk.[index].with-index' */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import('test')`,
      options: [{ importFunctions: ['dynamicImport'], allowEmpty: true }],
    },
    {
      code: `import(
        /* webpackMode: "lazy" */
        'test'
      )`,
      options: [{ importFunctions: ['dynamicImport'], allowEmpty: true }],
    },
    {
      code: `import(
        /* webpackChunkName: "someModule" */
        'test'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
        /* webpackChunkName: "Some_Other_Module" */
        "test"
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
        /* webpackChunkName: "SomeModule123" */
        "test"
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
        /* webpackChunkName: "someModule", webpackPrefetch: true */
        'test'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
        /* webpackChunkName: "someModule", webpackPrefetch: true, */
        'test'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
        /* webpackPrefetch: true, webpackChunkName: "someModule" */
        'test'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
        /* webpackPrefetch: true, webpackChunkName: "someModule", */
        'test'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
        /* webpackPrefetch: true */
        /* webpackChunkName: "someModule" */
        'test'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
        /* webpackChunkName: "someModule" */
        /* webpackPrefetch: true */
        'test'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
        /* webpackChunkName: "someModule" */
        /* webpackPrefetch: 12 */
        'test'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
        /* webpackChunkName: "someModule" */
        /* webpackPrefetch: -30 */
        'test'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
        /* webpackChunkName: 'someModule' */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
        /* webpackChunkName: "someModule" */
        'someModule'
      )`,
      options: [
        {
          importFunctions: ['dynamicImport'],
          webpackChunknameFormat: '[a-zA-Z-_/.]+',
        },
      ],
    },
    {
      code: `import(
        /* webpackChunkName: "[request]" */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
        /* webpackChunkName: "my-chunk-[request]-custom" */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
        /* webpackChunkName: '[index]' */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
        /* webpackChunkName: 'my-chunk.[index].with-index' */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
        /* webpackChunkName: "someModule" */
        /* webpackInclude: /\\.json$/ */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
        /* webpackChunkName: "someModule", webpackInclude: /\\.json$/ */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
        /* webpackChunkName: "someModule" */
        /* webpackExclude: /\\.json$/ */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
        /* webpackChunkName: "someModule", webpackExclude: /\\.json$/ */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
        /* webpackChunkName: "someModule" */
        /* webpackPreload: true */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
        /* webpackChunkName: "someModule" */
        /* webpackPreload: 0 */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
        /* webpackChunkName: "someModule" */
        /* webpackPreload: -2 */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
        /* webpackChunkName: "someModule", webpackPreload: false */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
        /* webpackChunkName: "someModule" */
        /* webpackIgnore: false */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
        /* webpackChunkName: "someModule", webpackIgnore: true */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
        /* webpackChunkName: "someModule" */
        /* webpackMode: "lazy" */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
        /* webpackChunkName: 'someModule', webpackMode: 'lazy' */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
        /* webpackChunkName: "someModule" */
        /* webpackMode: "lazy-once" */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
        /* webpackMode: "eager" */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
        /* webpackChunkName: "someModule" */
        /* webpackMode: "weak" */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
        /* webpackChunkName: "someModule" */
        /* webpackExports: "default" */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
        /* webpackChunkName: "someModule", webpackExports: "named" */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
        /* webpackChunkName: "someModule" */
        /* webpackExports: ["default", "named"] */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
        /* webpackChunkName: 'someModule', webpackExports: ['default', 'named'] */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
        /* webpackChunkName: "someModule" */
        /* webpackInclude: /\\.json$/ */
        /* webpackExclude: /\\.json$/ */
        /* webpackPrefetch: true */
        /* webpackPreload: true */
        /* webpackIgnore: false */
        /* webpackMode: "lazy" */
        /* webpackExports: ["default", "named"] */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: 'for (let { foo, bar } of baz) {}',
    },
    {
      code: 'for (let [ foo, bar ] of baz) {}',
    },
    {
      code: 'const { x, y } = bar',
    },
    {
      code: 'const { x, y, ...z } = bar',
    },
    {
      code: 'let x; export { x }',
    },
    {
      code: 'let x; export { x as y }',
    },
    {
      code: 'export const x = null',
    },
    {
      code: 'export var x = null',
    },
    {
      code: 'export let x = null',
    },
    {
      code: 'export default x',
    },
    {
      code: 'export default class x {}',
    },
    {
      code: 'import json from "./data.json"',
    },
    {
      code: 'import foo from "./foobar.json";',
    },
    {
      code: 'import foo from "./foobar";',
    },
    {
      code: 'import { foo } from "./issue-370-commonjs-namespace/bar"',
    },
    {
      code: 'export * from "./issue-370-commonjs-namespace/bar"',
    },
    {
      code: 'import * as a from "./commonjs-namespace/a"; a.b',
    },
    {
      code: 'import { foo } from "./ignore.invalid.extension"',
    },
    // ---- upstream valid: TypeScript ----
    {
      code: `import('test')`,
      options: [{ importFunctions: ['dynamicImport'], allowEmpty: true }],
    },
    {
      code: `import(
            /* webpackMode: "lazy" */
            'test'
          )`,
      options: [{ importFunctions: ['dynamicImport'], allowEmpty: true }],
    },
    {
      code: `import(
            /* webpackChunkName: "someModule" */
            'test'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
            /* webpackChunkName: "Some_Other_Module" */
            "test"
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
            /* webpackChunkName: "SomeModule123" */
            "test"
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
            /* webpackChunkName: "someModule", webpackPrefetch: true */
            'test'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
            /* webpackChunkName: "someModule", webpackPrefetch: true, */
            'test'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
            /* webpackPrefetch: true, webpackChunkName: "someModule" */
            'test'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
            /* webpackPrefetch: true, webpackChunkName: "someModule", */
            'test'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
            /* webpackPrefetch: true */
            /* webpackChunkName: "someModule" */
            'test'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
            /* webpackChunkName: "someModule" */
            /* webpackPrefetch: true */
            'test'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
            /* webpackChunkName: "someModule" */
            /* webpackPrefetch: 11 */
            'test'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
            /* webpackChunkName: "someModule" */
            /* webpackPrefetch: -11 */
            'test'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
            /* webpackChunkName: "someModule" */
            'someModule'
          )`,
      options: [
        {
          importFunctions: ['dynamicImport'],
          webpackChunknameFormat: '[a-zA-Z-_/.]+',
        },
      ],
    },
    {
      code: `import(
            /* webpackChunkName: 'someModule' */
            'test'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
            /* webpackChunkName: "[request]" */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
            /* webpackChunkName: "my-chunk-[request]-custom" */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
            /* webpackChunkName: '[index]' */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
            /* webpackChunkName: 'my-chunk.[index].with-index' */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
            /* webpackChunkName: "someModule" */
            /* webpackInclude: /\\.json$/ */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
            /* webpackChunkName: "someModule", webpackInclude: /\\.json$/ */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
            /* webpackChunkName: "someModule" */
            /* webpackExclude: /\\.json$/ */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
            /* webpackChunkName: "someModule", webpackExclude: /\\.json$/ */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
            /* webpackChunkName: "someModule" */
            /* webpackPreload: true */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
            /* webpackChunkName: "someModule", webpackPreload: false */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
            /* webpackChunkName: "someModule" */
            /* webpackIgnore: false */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
            /* webpackChunkName: "someModule", webpackIgnore: true */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
            /* webpackChunkName: 'someModule', webpackMode: 'lazy' */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
            /* webpackChunkName: "someModule" */
            /* webpackMode: "lazy-once" */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
            /* webpackChunkName: "someModule" */
            /* webpackMode: "lazy" */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
            /* webpackChunkName: "someModule" */
            /* webpackMode: "weak" */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
            /* webpackChunkName: "someModule" */
            /* webpackExports: "default" */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
            /* webpackChunkName: "someModule", webpackExports: "named" */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
            /* webpackChunkName: "someModule" */
            /* webpackExports: ["default", "named"] */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
            /* webpackChunkName: 'someModule', webpackExports: ['default', 'named'] */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
            /* webpackChunkName: "someModule" */
            /* webpackInclude: /\\.json$/ */
            /* webpackExclude: /\\.json$/ */
            /* webpackPrefetch: true */
            /* webpackPreload: true */
            /* webpackIgnore: false */
            /* webpackMode: "lazy" */
            /* webpackExports: ["default", "named"] */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
    {
      code: `import(
            /* webpackMode: "eager" */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
    },
  ],

  invalid: [
    // ---- upstream invalid: JavaScript ----
    {
      code: `import(
        // webpackChunkName: "someModule"
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a /* foo */ style comment, not a // foo comment',
        },
      ],
    },
    {
      code: `import('test')`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a leading comment with the webpack chunkname',
        },
      ],
    },
    {
      code: `import(
        /* webpackChunkName: someModule */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
        /* webpackChunkName: "someModule' */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
        /* webpackChunkName: 'someModule" */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
        /* webpackChunkName "someModule" */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
        /* webpackChunkName:"someModule" */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
        /* webpackChunkName: true */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a leading comment in the form /*webpackChunkName: ["\']([0-9a-zA-Z-_/.]|\\[(request|index)\\])+["\'],? */',
        },
      ],
    },
    {
      code: `import(
        /* webpackChunkName: "my-module-[id]" */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a leading comment in the form /*webpackChunkName: ["\']([0-9a-zA-Z-_/.]|\\[(request|index)\\])+["\'],? */',
        },
      ],
    },
    {
      code: `import(
        /* webpackChunkName: ["request"] */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a leading comment in the form /*webpackChunkName: ["\']([0-9a-zA-Z-_/.]|\\[(request|index)\\])+["\'],? */',
        },
      ],
    },
    {
      code: `import(
        /*webpackChunkName: "someModule"*/
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a block comment padded with spaces - /* foo */',
        },
      ],
    },
    {
      code: `import(
        /* webpackChunkName  :  "someModule" */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
        /* webpackChunkName: "someModule" ; */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
        /* totally not webpackChunkName: "someModule" */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
        /* webpackPrefetch: true */
        /* webpackChunk: "someModule" */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
        /* webpackPrefetch: true, webpackChunk: "someModule" */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
        /* webpackChunkName: "someModule123" */
        'someModule'
      )`,
      options: [
        {
          importFunctions: ['dynamicImport'],
          webpackChunknameFormat: '[a-zA-Z-_/.]+',
        },
      ],
      errors: [
        {
          message:
            'dynamic imports require a leading comment in the form /*webpackChunkName: ["\'][a-zA-Z-_/.]+["\'],? */',
        },
      ],
    },
    {
      code: `import(
        /* webpackPrefetch: "module", webpackChunkName: "someModule" */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
        /* webpackPreload: "module", webpackChunkName: "someModule" */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
        /* webpackIgnore: "no", webpackChunkName: "someModule" */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
        /* webpackInclude: "someModule", webpackChunkName: "someModule" */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
        /* webpackInclude: true, webpackChunkName: "someModule" */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
        /* webpackExclude: "someModule", webpackChunkName: "someModule" */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
        /* webpackExclude: true, webpackChunkName: "someModule" */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
        /* webpackMode: "fast", webpackChunkName: "someModule" */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
        /* webpackMode: true, webpackChunkName: "someModule" */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
        /* webpackExports: true, webpackChunkName: "someModule" */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
        /* webpackExports: /default/, webpackChunkName: "someModule" */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `dynamicImport(
        /* webpackChunkName "someModule" */
        'someModule'
      )`,
      options: [
        { importFunctions: ['dynamicImport', 'definitelyNotStaticImport'] },
      ],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `definitelyNotStaticImport(
        /* webpackChunkName "someModule" */
        'someModule'
      )`,
      options: [
        { importFunctions: ['dynamicImport', 'definitelyNotStaticImport'] },
      ],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `dynamicImport(
        // webpackChunkName: "someModule"
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a /* foo */ style comment, not a // foo comment',
        },
      ],
    },
    {
      code: `dynamicImport('test')`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a leading comment with the webpack chunkname',
        },
      ],
    },
    {
      code: `dynamicImport(
        /* webpackChunkName: someModule */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `dynamicImport(
        /* webpackChunkName "someModule" */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `dynamicImport(
        /* webpackChunkName:"someModule" */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `dynamicImport(
        /* webpackChunkName: "someModule123" */
        'someModule'
      )`,
      options: [
        {
          importFunctions: ['dynamicImport'],
          webpackChunknameFormat: '[a-zA-Z-_/.]+',
        },
      ],
      errors: [
        {
          message:
            'dynamic imports require a leading comment in the form /*webpackChunkName: ["\'][a-zA-Z-_/.]+["\'],? */',
        },
      ],
    },
    {
      code: `import(
        /* webpackChunkName: "someModule" */
        /* webpackMode: "eager" */
        'someModule'
      )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports using eager mode do not need a webpackChunkName',
          suggestions: [
            {
              desc: 'Remove webpackChunkName',
              output: `import(
        
        /* webpackMode: "eager" */
        'someModule'
      )`,
            },
            {
              desc: 'Remove webpackMode',
              output: `import(
        /* webpackChunkName: "someModule" */
        
        'someModule'
      )`,
            },
          ],
        },
      ],
    },
    // ---- upstream invalid: TypeScript ----
    {
      code: `import(
            // webpackChunkName: "someModule"
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a /* foo */ style comment, not a // foo comment',
        },
      ],
    },
    {
      code: `import('test')`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a leading comment with the webpack chunkname',
        },
      ],
    },
    {
      code: `import(
            /* webpackChunkName: someModule */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
            /* webpackChunkName "someModule' */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
            /* webpackChunkName 'someModule" */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
            /* webpackChunkName "someModule" */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
            /* webpackChunkName:"someModule" */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
            /*webpackChunkName: "someModule"*/
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a block comment padded with spaces - /* foo */',
        },
      ],
    },
    {
      code: `import(
            /* webpackChunkName  :  "someModule" */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
            /* webpackChunkName: "someModule" ; */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
            /* totally not webpackChunkName: "someModule" */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
            /* webpackPrefetch: true */
            /* webpackChunk: "someModule" */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
            /* webpackPrefetch: true, webpackChunk: "someModule" */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
            /* webpackChunkName: true */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a leading comment in the form /*webpackChunkName: ["\']([0-9a-zA-Z-_/.]|\\[(request|index)\\])+["\'],? */',
        },
      ],
    },
    {
      code: `import(
            /* webpackChunkName: "my-module-[id]" */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a leading comment in the form /*webpackChunkName: ["\']([0-9a-zA-Z-_/.]|\\[(request|index)\\])+["\'],? */',
        },
      ],
    },
    {
      code: `import(
            /* webpackChunkName: ["request"] */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a leading comment in the form /*webpackChunkName: ["\']([0-9a-zA-Z-_/.]|\\[(request|index)\\])+["\'],? */',
        },
      ],
    },
    {
      code: `import(
            /* webpackChunkName: "someModule123" */
            'someModule'
          )`,
      options: [
        {
          importFunctions: ['dynamicImport'],
          webpackChunknameFormat: '[a-zA-Z-_/.]+',
        },
      ],
      errors: [
        {
          message:
            'dynamic imports require a leading comment in the form /*webpackChunkName: ["\'][a-zA-Z-_/.]+["\'],? */',
        },
      ],
    },
    {
      code: `import(
            /* webpackPrefetch: "module", webpackChunkName: "someModule" */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
            /* webpackPreload: "module", webpackChunkName: "someModule" */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
            /* webpackIgnore: "no", webpackChunkName: "someModule" */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
            /* webpackInclude: "someModule", webpackChunkName: "someModule" */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
            /* webpackInclude: true, webpackChunkName: "someModule" */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
            /* webpackExclude: "someModule", webpackChunkName: "someModule" */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
            /* webpackExclude: true, webpackChunkName: "someModule" */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
            /* webpackMode: "fast", webpackChunkName: "someModule" */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
            /* webpackMode: true, webpackChunkName: "someModule" */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
            /* webpackExports: true, webpackChunkName: "someModule" */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
            /* webpackExports: /default/, webpackChunkName: "someModule" */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
        },
      ],
    },
    {
      code: `import(
            /* webpackChunkName: "someModule", webpackMode: "eager" */
            'someModule'
          )`,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports using eager mode do not need a webpackChunkName',
          suggestions: [
            {
              desc: 'Remove webpackChunkName',
              output: `import(
            /* webpackMode: "eager" */
            'someModule'
          )`,
            },
            {
              desc: 'Remove webpackMode',
              output: `import(
            /* webpackChunkName: "someModule" */
            'someModule'
          )`,
            },
          ],
        },
      ],
    },
    {
      code: `
            import(
              /* webpackMode: "eager", webpackChunkName: "someModule" */
              'someModule'
            )
          `,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports using eager mode do not need a webpackChunkName',
          suggestions: [
            {
              desc: 'Remove webpackChunkName',
              output: `
            import(
              /* webpackMode: "eager" */
              'someModule'
            )
          `,
            },
            {
              desc: 'Remove webpackMode',
              output: `
            import(
              /* webpackChunkName: "someModule" */
              'someModule'
            )
          `,
            },
          ],
        },
      ],
    },
    {
      code: `
            import(
              /* webpackMode: "eager", webpackPrefetch: true, webpackChunkName: "someModule" */
              'someModule'
            )
          `,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports using eager mode do not need a webpackChunkName',
          suggestions: [
            {
              desc: 'Remove webpackChunkName',
              output: `
            import(
              /* webpackMode: "eager", webpackPrefetch: true */
              'someModule'
            )
          `,
            },
            {
              desc: 'Remove webpackMode',
              output: `
            import(
              /* webpackPrefetch: true, webpackChunkName: "someModule" */
              'someModule'
            )
          `,
            },
          ],
        },
      ],
    },
    {
      code: `
            import(
              /* webpackChunkName: "someModule" */
              /* webpackMode: "eager" */
              'someModule'
            )
          `,
      options: [{ importFunctions: ['dynamicImport'] }],
      errors: [
        {
          message:
            'dynamic imports using eager mode do not need a webpackChunkName',
          suggestions: [
            {
              desc: 'Remove webpackChunkName',
              output: `
            import(
              
              /* webpackMode: "eager" */
              'someModule'
            )
          `,
            },
            {
              desc: 'Remove webpackMode',
              output: `
            import(
              /* webpackChunkName: "someModule" */
              
              'someModule'
            )
          `,
            },
          ],
        },
      ],
    },
  ],
});
