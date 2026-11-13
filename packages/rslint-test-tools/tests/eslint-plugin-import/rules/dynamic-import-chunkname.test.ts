// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
new RuleTester(fixtureConfig()).run('dynamic-import-chunkname', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: 'dynamicImport(\n        /* webpackChunkName: "someModule" */\n        \'test\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 1',
      code: 'dynamicImport(\n        /* webpackChunkName: "Some_Other_Module" */\n        "test"\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 2',
      code: 'dynamicImport(\n        /* webpackChunkName: "SomeModule123" */\n        "test"\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 3',
      code: 'dynamicImport(\n        /* webpackChunkName: "someModule" */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
          webpackChunknameFormat: '[a-zA-Z-_/.]+',
        },
      ],
    },
    {
      name: 'run 0 valid 4',
      code: 'dynamicImport(\n        /* webpackChunkName: "[request]" */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 5',
      code: 'dynamicImport(\n        /* webpackChunkName: "my-chunk-[request]-custom" */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 6',
      code: "dynamicImport(\n        /* webpackChunkName: '[index]' */\n        'someModule'\n      )",
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 7',
      code: "dynamicImport(\n        /* webpackChunkName: 'my-chunk.[index].with-index' */\n        'someModule'\n      )",
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 8',
      code: "import('test')",
      options: [
        {
          importFunctions: ['dynamicImport'],
          allowEmpty: true,
        },
      ],
    },
    {
      name: 'run 0 valid 9',
      code: 'import(\n        /* webpackMode: "lazy" */\n        \'test\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
          allowEmpty: true,
        },
      ],
    },
    {
      name: 'run 0 valid 10',
      code: 'import(\n        /* webpackChunkName: "someModule" */\n        \'test\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 11',
      code: 'import(\n        /* webpackChunkName: "Some_Other_Module" */\n        "test"\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 12',
      code: 'import(\n        /* webpackChunkName: "SomeModule123" */\n        "test"\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 13',
      code: 'import(\n        /* webpackChunkName: "someModule", webpackPrefetch: true */\n        \'test\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 14',
      code: 'import(\n        /* webpackChunkName: "someModule", webpackPrefetch: true, */\n        \'test\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 15',
      code: 'import(\n        /* webpackPrefetch: true, webpackChunkName: "someModule" */\n        \'test\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 16',
      code: 'import(\n        /* webpackPrefetch: true, webpackChunkName: "someModule", */\n        \'test\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 17',
      code: 'import(\n        /* webpackPrefetch: true */\n        /* webpackChunkName: "someModule" */\n        \'test\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 18',
      code: 'import(\n        /* webpackChunkName: "someModule" */\n        /* webpackPrefetch: true */\n        \'test\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 19',
      code: 'import(\n        /* webpackChunkName: "someModule" */\n        /* webpackPrefetch: 12 */\n        \'test\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 20',
      code: 'import(\n        /* webpackChunkName: "someModule" */\n        /* webpackPrefetch: -30 */\n        \'test\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 21',
      code: "import(\n        /* webpackChunkName: 'someModule' */\n        'someModule'\n      )",
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 22',
      code: 'import(\n        /* webpackChunkName: "someModule" */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
          webpackChunknameFormat: '[a-zA-Z-_/.]+',
        },
      ],
    },
    {
      name: 'run 0 valid 23',
      code: 'import(\n        /* webpackChunkName: "[request]" */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 24',
      code: 'import(\n        /* webpackChunkName: "my-chunk-[request]-custom" */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 25',
      code: "import(\n        /* webpackChunkName: '[index]' */\n        'someModule'\n      )",
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 26',
      code: "import(\n        /* webpackChunkName: 'my-chunk.[index].with-index' */\n        'someModule'\n      )",
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 27',
      code: 'import(\n        /* webpackChunkName: "someModule" */\n        /* webpackInclude: /\\.json$/ */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 28',
      code: 'import(\n        /* webpackChunkName: "someModule", webpackInclude: /\\.json$/ */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 29',
      code: 'import(\n        /* webpackChunkName: "someModule" */\n        /* webpackExclude: /\\.json$/ */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 30',
      code: 'import(\n        /* webpackChunkName: "someModule", webpackExclude: /\\.json$/ */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 31',
      code: 'import(\n        /* webpackChunkName: "someModule" */\n        /* webpackPreload: true */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 32',
      code: 'import(\n        /* webpackChunkName: "someModule" */\n        /* webpackPreload: 0 */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 33',
      code: 'import(\n        /* webpackChunkName: "someModule" */\n        /* webpackPreload: -2 */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 34',
      code: 'import(\n        /* webpackChunkName: "someModule", webpackPreload: false */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 35',
      code: 'import(\n        /* webpackChunkName: "someModule" */\n        /* webpackIgnore: false */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 36',
      code: 'import(\n        /* webpackChunkName: "someModule", webpackIgnore: true */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 37',
      code: 'import(\n        /* webpackChunkName: "someModule" */\n        /* webpackMode: "lazy" */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 38',
      code: "import(\n        /* webpackChunkName: 'someModule', webpackMode: 'lazy' */\n        'someModule'\n      )",
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 39',
      code: 'import(\n        /* webpackChunkName: "someModule" */\n        /* webpackMode: "lazy-once" */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 40',
      code: 'import(\n        /* webpackMode: "eager" */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 41',
      code: 'import(\n        /* webpackChunkName: "someModule" */\n        /* webpackMode: "weak" */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 42',
      code: 'import(\n        /* webpackChunkName: "someModule" */\n        /* webpackExports: "default" */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 43',
      code: 'import(\n        /* webpackChunkName: "someModule", webpackExports: "named" */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 44',
      code: 'import(\n        /* webpackChunkName: "someModule" */\n        /* webpackExports: ["default", "named"] */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 45',
      code: "import(\n        /* webpackChunkName: 'someModule', webpackExports: ['default', 'named'] */\n        'someModule'\n      )",
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 46',
      code: 'import(\n        /* webpackChunkName: "someModule" */\n        /* webpackInclude: /\\.json$/ */\n        /* webpackExclude: /\\.json$/ */\n        /* webpackPrefetch: true */\n        /* webpackPreload: true */\n        /* webpackIgnore: false */\n        /* webpackMode: "lazy" */\n        /* webpackExports: ["default", "named"] */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 47',
      code: 'for (let { foo, bar } of baz) {}',
    },
    {
      name: 'run 0 valid 48',
      code: 'for (let [ foo, bar ] of baz) {}',
    },
    {
      name: 'run 0 valid 49',
      code: 'const { x, y } = bar',
    },
    {
      name: 'run 0 valid 50',
      code: 'const { x, y, ...z } = bar',
    },
    {
      name: 'run 0 valid 51',
      code: 'let x; export { x }',
    },
    {
      name: 'run 0 valid 52',
      code: 'let x; export { x as y }',
    },
    {
      name: 'run 0 valid 53',
      code: 'export const x = null',
    },
    {
      name: 'run 0 valid 54',
      code: 'export var x = null',
    },
    {
      name: 'run 0 valid 55',
      code: 'export let x = null',
    },
    {
      name: 'run 0 valid 56',
      code: 'export default x',
    },
    {
      name: 'run 0 valid 57',
      code: 'export default class x {}',
    },
    {
      name: 'run 0 valid 58',
      code: 'import json from "./data.json"',
    },
    {
      name: 'run 0 valid 59',
      code: 'import foo from "./foobar.json";',
    },
    {
      name: 'run 0 valid 60',
      code: 'import foo from "./foobar";',
    },
    {
      name: 'run 0 valid 61',
      code: 'import { foo } from "./issue-370-commonjs-namespace/bar"',
    },
    {
      name: 'run 0 valid 62',
      code: 'export * from "./issue-370-commonjs-namespace/bar"',
    },
    {
      name: 'run 0 valid 63',
      code: 'import * as a from "./commonjs-namespace/a"; a.b',
    },
    {
      name: 'run 0 valid 64',
      code: 'import { foo } from "./ignore.invalid.extension"',
    },
    {
      name: 'run 0 valid 65',
      code: "import('test')",
      options: [
        {
          importFunctions: ['dynamicImport'],
          allowEmpty: true,
        },
      ],
    },
    {
      name: 'run 0 valid 66',
      code: 'import(\n            /* webpackMode: "lazy" */\n            \'test\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
          allowEmpty: true,
        },
      ],
    },
    {
      name: 'run 0 valid 67',
      code: 'import(\n            /* webpackChunkName: "someModule" */\n            \'test\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 68',
      code: 'import(\n            /* webpackChunkName: "Some_Other_Module" */\n            "test"\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 69',
      code: 'import(\n            /* webpackChunkName: "SomeModule123" */\n            "test"\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 70',
      code: 'import(\n            /* webpackChunkName: "someModule", webpackPrefetch: true */\n            \'test\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 71',
      code: 'import(\n            /* webpackChunkName: "someModule", webpackPrefetch: true, */\n            \'test\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 72',
      code: 'import(\n            /* webpackPrefetch: true, webpackChunkName: "someModule" */\n            \'test\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 73',
      code: 'import(\n            /* webpackPrefetch: true, webpackChunkName: "someModule", */\n            \'test\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 74',
      code: 'import(\n            /* webpackPrefetch: true */\n            /* webpackChunkName: "someModule" */\n            \'test\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 75',
      code: 'import(\n            /* webpackChunkName: "someModule" */\n            /* webpackPrefetch: true */\n            \'test\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 76',
      code: 'import(\n            /* webpackChunkName: "someModule" */\n            /* webpackPrefetch: 11 */\n            \'test\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 77',
      code: 'import(\n            /* webpackChunkName: "someModule" */\n            /* webpackPrefetch: -11 */\n            \'test\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 78',
      code: 'import(\n            /* webpackChunkName: "someModule" */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
          webpackChunknameFormat: '[a-zA-Z-_/.]+',
        },
      ],
    },
    {
      name: 'run 0 valid 79',
      code: "import(\n            /* webpackChunkName: 'someModule' */\n            'test'\n          )",
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 80',
      code: 'import(\n            /* webpackChunkName: "[request]" */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 81',
      code: 'import(\n            /* webpackChunkName: "my-chunk-[request]-custom" */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 82',
      code: "import(\n            /* webpackChunkName: '[index]' */\n            'someModule'\n          )",
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 83',
      code: "import(\n            /* webpackChunkName: 'my-chunk.[index].with-index' */\n            'someModule'\n          )",
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 84',
      code: 'import(\n            /* webpackChunkName: "someModule" */\n            /* webpackInclude: /\\.json$/ */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 85',
      code: 'import(\n            /* webpackChunkName: "someModule", webpackInclude: /\\.json$/ */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 86',
      code: 'import(\n            /* webpackChunkName: "someModule" */\n            /* webpackExclude: /\\.json$/ */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 87',
      code: 'import(\n            /* webpackChunkName: "someModule", webpackExclude: /\\.json$/ */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 88',
      code: 'import(\n            /* webpackChunkName: "someModule" */\n            /* webpackPreload: true */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 89',
      code: 'import(\n            /* webpackChunkName: "someModule", webpackPreload: false */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 90',
      code: 'import(\n            /* webpackChunkName: "someModule" */\n            /* webpackIgnore: false */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 91',
      code: 'import(\n            /* webpackChunkName: "someModule", webpackIgnore: true */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 92',
      code: "import(\n            /* webpackChunkName: 'someModule', webpackMode: 'lazy' */\n            'someModule'\n          )",
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 93',
      code: 'import(\n            /* webpackChunkName: "someModule" */\n            /* webpackMode: "lazy-once" */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 94',
      code: 'import(\n            /* webpackChunkName: "someModule" */\n            /* webpackMode: "lazy" */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 95',
      code: 'import(\n            /* webpackChunkName: "someModule" */\n            /* webpackMode: "weak" */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 96',
      code: 'import(\n            /* webpackChunkName: "someModule" */\n            /* webpackExports: "default" */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 97',
      code: 'import(\n            /* webpackChunkName: "someModule", webpackExports: "named" */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 98',
      code: 'import(\n            /* webpackChunkName: "someModule" */\n            /* webpackExports: ["default", "named"] */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 99',
      code: "import(\n            /* webpackChunkName: 'someModule', webpackExports: ['default', 'named'] */\n            'someModule'\n          )",
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 100',
      code: 'import(\n            /* webpackChunkName: "someModule" */\n            /* webpackInclude: /\\.json$/ */\n            /* webpackExclude: /\\.json$/ */\n            /* webpackPrefetch: true */\n            /* webpackPreload: true */\n            /* webpackIgnore: false */\n            /* webpackMode: "lazy" */\n            /* webpackExports: ["default", "named"] */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
    {
      name: 'run 0 valid 101',
      code: 'import(\n            /* webpackMode: "eager" */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: 'import(\n        // webpackChunkName: "someModule"\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a /* foo */ style comment, not a // foo comment',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 8,
        },
      ],
    },
    {
      name: 'run 0 invalid 1',
      code: "import('test')",
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a leading comment with the webpack chunkname',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      name: 'run 0 invalid 2',
      code: "import(\n        /* webpackChunkName: someModule */\n        'someModule'\n      )",
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 8,
        },
      ],
    },
    {
      name: 'run 0 invalid 3',
      code: "import(\n        /* webpackChunkName: \"someModule' */\n        'someModule'\n      )",
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 8,
        },
      ],
    },
    {
      name: 'run 0 invalid 4',
      code: "import(\n        /* webpackChunkName: 'someModule\" */\n        'someModule'\n      )",
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 8,
        },
      ],
    },
    {
      name: 'run 0 invalid 5',
      code: 'import(\n        /* webpackChunkName "someModule" */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 8,
        },
      ],
    },
    {
      name: 'run 0 invalid 6',
      code: 'import(\n        /* webpackChunkName:"someModule" */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 8,
        },
      ],
    },
    {
      name: 'run 0 invalid 7',
      code: "import(\n        /* webpackChunkName: true */\n        'someModule'\n      )",
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a leading comment in the form /*webpackChunkName: ["\']([0-9a-zA-Z-_/.]|\\[(request|index)\\])+["\'],? */',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 8,
        },
      ],
    },
    {
      name: 'run 0 invalid 8',
      code: 'import(\n        /* webpackChunkName: "my-module-[id]" */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a leading comment in the form /*webpackChunkName: ["\']([0-9a-zA-Z-_/.]|\\[(request|index)\\])+["\'],? */',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 8,
        },
      ],
    },
    {
      name: 'run 0 invalid 9',
      code: 'import(\n        /* webpackChunkName: ["request"] */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a leading comment in the form /*webpackChunkName: ["\']([0-9a-zA-Z-_/.]|\\[(request|index)\\])+["\'],? */',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 8,
        },
      ],
    },
    {
      name: 'run 0 invalid 10',
      code: 'import(\n        /*webpackChunkName: "someModule"*/\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a block comment padded with spaces - /* foo */',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 8,
        },
      ],
    },
    {
      name: 'run 0 invalid 11',
      code: 'import(\n        /* webpackChunkName  :  "someModule" */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 8,
        },
      ],
    },
    {
      name: 'run 0 invalid 12',
      code: 'import(\n        /* webpackChunkName: "someModule" ; */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 8,
        },
      ],
    },
    {
      name: 'run 0 invalid 13',
      code: 'import(\n        /* totally not webpackChunkName: "someModule" */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 8,
        },
      ],
    },
    {
      name: 'run 0 invalid 14',
      code: 'import(\n        /* webpackPrefetch: true */\n        /* webpackChunk: "someModule" */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 5,
          endColumn: 8,
        },
      ],
    },
    {
      name: 'run 0 invalid 15',
      code: 'import(\n        /* webpackPrefetch: true, webpackChunk: "someModule" */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 8,
        },
      ],
    },
    {
      name: 'run 0 invalid 16',
      code: 'import(\n        /* webpackChunkName: "someModule123" */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
          webpackChunknameFormat: '[a-zA-Z-_/.]+',
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a leading comment in the form /*webpackChunkName: ["\'][a-zA-Z-_/.]+["\'],? */',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 8,
        },
      ],
    },
    {
      name: 'run 0 invalid 17',
      code: 'import(\n        /* webpackPrefetch: "module", webpackChunkName: "someModule" */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 8,
        },
      ],
    },
    {
      name: 'run 0 invalid 18',
      code: 'import(\n        /* webpackPreload: "module", webpackChunkName: "someModule" */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 8,
        },
      ],
    },
    {
      name: 'run 0 invalid 19',
      code: 'import(\n        /* webpackIgnore: "no", webpackChunkName: "someModule" */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 8,
        },
      ],
    },
    {
      name: 'run 0 invalid 20',
      code: 'import(\n        /* webpackInclude: "someModule", webpackChunkName: "someModule" */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 8,
        },
      ],
    },
    {
      name: 'run 0 invalid 21',
      code: 'import(\n        /* webpackInclude: true, webpackChunkName: "someModule" */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 8,
        },
      ],
    },
    {
      name: 'run 0 invalid 22',
      code: 'import(\n        /* webpackExclude: "someModule", webpackChunkName: "someModule" */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 8,
        },
      ],
    },
    {
      name: 'run 0 invalid 23',
      code: 'import(\n        /* webpackExclude: true, webpackChunkName: "someModule" */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 8,
        },
      ],
    },
    {
      name: 'run 0 invalid 24',
      code: 'import(\n        /* webpackMode: "fast", webpackChunkName: "someModule" */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 8,
        },
      ],
    },
    {
      name: 'run 0 invalid 25',
      code: 'import(\n        /* webpackMode: true, webpackChunkName: "someModule" */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 8,
        },
      ],
    },
    {
      name: 'run 0 invalid 26',
      code: 'import(\n        /* webpackExports: true, webpackChunkName: "someModule" */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 8,
        },
      ],
    },
    {
      name: 'run 0 invalid 27',
      code: 'import(\n        /* webpackExports: /default/, webpackChunkName: "someModule" */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 8,
        },
      ],
    },
    {
      name: 'run 0 invalid 28',
      code: 'dynamicImport(\n        /* webpackChunkName "someModule" */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport', 'definitelyNotStaticImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 8,
        },
      ],
    },
    {
      name: 'run 0 invalid 29',
      code: 'definitelyNotStaticImport(\n        /* webpackChunkName "someModule" */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport', 'definitelyNotStaticImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 8,
        },
      ],
    },
    {
      name: 'run 0 invalid 30',
      code: 'dynamicImport(\n        // webpackChunkName: "someModule"\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a /* foo */ style comment, not a // foo comment',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 8,
        },
      ],
    },
    {
      name: 'run 0 invalid 31',
      code: "dynamicImport('test')",
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a leading comment with the webpack chunkname',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      name: 'run 0 invalid 32',
      code: "dynamicImport(\n        /* webpackChunkName: someModule */\n        'someModule'\n      )",
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 8,
        },
      ],
    },
    {
      name: 'run 0 invalid 33',
      code: 'dynamicImport(\n        /* webpackChunkName "someModule" */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 8,
        },
      ],
    },
    {
      name: 'run 0 invalid 34',
      code: 'dynamicImport(\n        /* webpackChunkName:"someModule" */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 8,
        },
      ],
    },
    {
      name: 'run 0 invalid 35',
      code: 'dynamicImport(\n        /* webpackChunkName: "someModule123" */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
          webpackChunknameFormat: '[a-zA-Z-_/.]+',
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a leading comment in the form /*webpackChunkName: ["\'][a-zA-Z-_/.]+["\'],? */',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 8,
        },
      ],
    },
    {
      name: 'run 0 invalid 36',
      code: 'import(\n        /* webpackChunkName: "someModule" */\n        /* webpackMode: "eager" */\n        \'someModule\'\n      )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports using eager mode do not need a webpackChunkName',
          line: 1,
          column: 1,
          endLine: 5,
          endColumn: 8,
          suggestions: [
            {
              messageId: '',
              desc: 'Remove webpackChunkName',
              output:
                'import(\n        \n        /* webpackMode: "eager" */\n        \'someModule\'\n      )',
            },
            {
              messageId: '',
              desc: 'Remove webpackMode',
              output:
                'import(\n        /* webpackChunkName: "someModule" */\n        \n        \'someModule\'\n      )',
            },
          ],
        },
      ],
    },
    {
      name: 'run 0 invalid 37',
      code: 'import(\n            // webpackChunkName: "someModule"\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a /* foo */ style comment, not a // foo comment',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      name: 'run 0 invalid 38',
      code: "import('test')",
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a leading comment with the webpack chunkname',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      name: 'run 0 invalid 39',
      code: "import(\n            /* webpackChunkName: someModule */\n            'someModule'\n          )",
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      name: 'run 0 invalid 40',
      code: "import(\n            /* webpackChunkName \"someModule' */\n            'someModule'\n          )",
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      name: 'run 0 invalid 41',
      code: "import(\n            /* webpackChunkName 'someModule\" */\n            'someModule'\n          )",
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      name: 'run 0 invalid 42',
      code: 'import(\n            /* webpackChunkName "someModule" */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      name: 'run 0 invalid 43',
      code: 'import(\n            /* webpackChunkName:"someModule" */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      name: 'run 0 invalid 44',
      code: 'import(\n            /*webpackChunkName: "someModule"*/\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a block comment padded with spaces - /* foo */',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      name: 'run 0 invalid 45',
      code: 'import(\n            /* webpackChunkName  :  "someModule" */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      name: 'run 0 invalid 46',
      code: 'import(\n            /* webpackChunkName: "someModule" ; */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      name: 'run 0 invalid 47',
      code: 'import(\n            /* totally not webpackChunkName: "someModule" */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      name: 'run 0 invalid 48',
      code: 'import(\n            /* webpackPrefetch: true */\n            /* webpackChunk: "someModule" */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 5,
          endColumn: 12,
        },
      ],
    },
    {
      name: 'run 0 invalid 49',
      code: 'import(\n            /* webpackPrefetch: true, webpackChunk: "someModule" */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      name: 'run 0 invalid 50',
      code: "import(\n            /* webpackChunkName: true */\n            'someModule'\n          )",
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a leading comment in the form /*webpackChunkName: ["\']([0-9a-zA-Z-_/.]|\\[(request|index)\\])+["\'],? */',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      name: 'run 0 invalid 51',
      code: 'import(\n            /* webpackChunkName: "my-module-[id]" */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a leading comment in the form /*webpackChunkName: ["\']([0-9a-zA-Z-_/.]|\\[(request|index)\\])+["\'],? */',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      name: 'run 0 invalid 52',
      code: 'import(\n            /* webpackChunkName: ["request"] */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a leading comment in the form /*webpackChunkName: ["\']([0-9a-zA-Z-_/.]|\\[(request|index)\\])+["\'],? */',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      name: 'run 0 invalid 53',
      code: 'import(\n            /* webpackChunkName: "someModule123" */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
          webpackChunknameFormat: '[a-zA-Z-_/.]+',
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a leading comment in the form /*webpackChunkName: ["\'][a-zA-Z-_/.]+["\'],? */',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      name: 'run 0 invalid 54',
      code: 'import(\n            /* webpackPrefetch: "module", webpackChunkName: "someModule" */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      name: 'run 0 invalid 55',
      code: 'import(\n            /* webpackPreload: "module", webpackChunkName: "someModule" */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      name: 'run 0 invalid 56',
      code: 'import(\n            /* webpackIgnore: "no", webpackChunkName: "someModule" */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      name: 'run 0 invalid 57',
      code: 'import(\n            /* webpackInclude: "someModule", webpackChunkName: "someModule" */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      name: 'run 0 invalid 58',
      code: 'import(\n            /* webpackInclude: true, webpackChunkName: "someModule" */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      name: 'run 0 invalid 59',
      code: 'import(\n            /* webpackExclude: "someModule", webpackChunkName: "someModule" */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      name: 'run 0 invalid 60',
      code: 'import(\n            /* webpackExclude: true, webpackChunkName: "someModule" */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      name: 'run 0 invalid 61',
      code: 'import(\n            /* webpackMode: "fast", webpackChunkName: "someModule" */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      name: 'run 0 invalid 62',
      code: 'import(\n            /* webpackMode: true, webpackChunkName: "someModule" */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      name: 'run 0 invalid 63',
      code: 'import(\n            /* webpackExports: true, webpackChunkName: "someModule" */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      name: 'run 0 invalid 64',
      code: 'import(\n            /* webpackExports: /default/, webpackChunkName: "someModule" */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports require a "webpack" comment with valid syntax',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      name: 'run 0 invalid 65',
      code: 'import(\n            /* webpackChunkName: "someModule", webpackMode: "eager" */\n            \'someModule\'\n          )',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports using eager mode do not need a webpackChunkName',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 12,
          suggestions: [
            {
              messageId: '',
              desc: 'Remove webpackChunkName',
              output:
                'import(\n            /* webpackMode: "eager" */\n            \'someModule\'\n          )',
            },
            {
              messageId: '',
              desc: 'Remove webpackMode',
              output:
                'import(\n            /* webpackChunkName: "someModule" */\n            \'someModule\'\n          )',
            },
          ],
        },
      ],
    },
    {
      name: 'run 0 invalid 66',
      code: '\n            import(\n              /* webpackMode: "eager", webpackChunkName: "someModule" */\n              \'someModule\'\n            )\n          ',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports using eager mode do not need a webpackChunkName',
          line: 2,
          column: 13,
          endLine: 5,
          endColumn: 14,
          suggestions: [
            {
              messageId: '',
              desc: 'Remove webpackChunkName',
              output:
                '\n            import(\n              /* webpackMode: "eager" */\n              \'someModule\'\n            )\n          ',
            },
            {
              messageId: '',
              desc: 'Remove webpackMode',
              output:
                '\n            import(\n              /* webpackChunkName: "someModule" */\n              \'someModule\'\n            )\n          ',
            },
          ],
        },
      ],
    },
    {
      name: 'run 0 invalid 67',
      code: '\n            import(\n              /* webpackMode: "eager", webpackPrefetch: true, webpackChunkName: "someModule" */\n              \'someModule\'\n            )\n          ',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports using eager mode do not need a webpackChunkName',
          line: 2,
          column: 13,
          endLine: 5,
          endColumn: 14,
          suggestions: [
            {
              messageId: '',
              desc: 'Remove webpackChunkName',
              output:
                '\n            import(\n              /* webpackMode: "eager", webpackPrefetch: true */\n              \'someModule\'\n            )\n          ',
            },
            {
              messageId: '',
              desc: 'Remove webpackMode',
              output:
                '\n            import(\n              /* webpackPrefetch: true, webpackChunkName: "someModule" */\n              \'someModule\'\n            )\n          ',
            },
          ],
        },
      ],
    },
    {
      name: 'run 0 invalid 68',
      code: '\n            import(\n              /* webpackChunkName: "someModule" */\n              /* webpackMode: "eager" */\n              \'someModule\'\n            )\n          ',
      options: [
        {
          importFunctions: ['dynamicImport'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'dynamic imports using eager mode do not need a webpackChunkName',
          line: 2,
          column: 13,
          endLine: 6,
          endColumn: 14,
          suggestions: [
            {
              messageId: '',
              desc: 'Remove webpackChunkName',
              output:
                '\n            import(\n              \n              /* webpackMode: "eager" */\n              \'someModule\'\n            )\n          ',
            },
            {
              messageId: '',
              desc: 'Remove webpackMode',
              output:
                '\n            import(\n              /* webpackChunkName: "someModule" */\n              \n              \'someModule\'\n            )\n          ',
            },
          ],
        },
      ],
    },
  ],
});
