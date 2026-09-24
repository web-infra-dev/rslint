import { RuleTester } from '../rule-tester';

// Upstream: https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/enforce-node-protocol-usage.js
// Preserve all templates in both directions, including the always cases skipped
// by the upstream semver guard. The Go suite checks ranges, message IDs and fixes;
// this wrapper currently checks counts and exact message text only.
const invalidTemplates = [
  {
    code: 'import x from "fs";',
    output: 'import x from "node:fs";',
    moduleName: 'fs',
  },
  {
    code: 'export {promises} from "fs";',
    output: 'export {promises} from "node:fs";',
    moduleName: 'fs',
  },
  {
    code: "\n        async function foo() {\n          const x = await import('fs');\n        }\n      ",
    output:
      "\n        async function foo() {\n          const x = await import('node:fs');\n        }\n      ",
    moduleName: 'fs',
  },
  {
    code: 'import x from "fs/promises";',
    output: 'import x from "node:fs/promises";',
    moduleName: 'fs/promises',
  },
  {
    code: 'export {promises} from "fs/promises";',
    output: 'export {promises} from "node:fs/promises";',
    moduleName: 'fs/promises',
  },
  {
    code: "\n        async function foo() {\n          const x = await import('fs/promises');\n        }\n      ",
    output:
      "\n        async function foo() {\n          const x = await import('node:fs/promises');\n        }\n      ",
    moduleName: 'fs/promises',
  },
  {
    code: 'import x from "buffer";',
    output: 'import x from "node:buffer";',
    moduleName: 'buffer',
  },
  {
    code: 'export {promises} from "buffer";',
    output: 'export {promises} from "node:buffer";',
    moduleName: 'buffer',
  },
  {
    code: "\n        async function foo() {\n          const x = await import('buffer');\n        }\n      ",
    output:
      "\n        async function foo() {\n          const x = await import('node:buffer');\n        }\n      ",
    moduleName: 'buffer',
  },
  {
    code: 'import x from "child_process";',
    output: 'import x from "node:child_process";',
    moduleName: 'child_process',
  },
  {
    code: 'export {promises} from "child_process";',
    output: 'export {promises} from "node:child_process";',
    moduleName: 'child_process',
  },
  {
    code: "\n        async function foo() {\n          const x = await import('child_process');\n        }\n      ",
    output:
      "\n        async function foo() {\n          const x = await import('node:child_process');\n        }\n      ",
    moduleName: 'child_process',
  },
  {
    code: 'import x from "timers/promises";',
    output: 'import x from "node:timers/promises";',
    moduleName: 'timers/promises',
  },
  {
    code: 'export {promises} from "timers/promises";',
    output: 'export {promises} from "node:timers/promises";',
    moduleName: 'timers/promises',
  },
  {
    code: "\n        async function foo() {\n          const x = await import('timers/promises');\n        }\n      ",
    output:
      "\n        async function foo() {\n          const x = await import('node:timers/promises');\n        }\n      ",
    moduleName: 'timers/promises',
  },
  {
    code: 'import fs from "fs/promises";',
    output: 'import fs from "node:fs/promises";',
    moduleName: 'fs/promises',
    settings: {
      'import/node-version': '16.0.0',
    },
  },
  {
    code: 'export {default} from "fs/promises";',
    output: 'export {default} from "node:fs/promises";',
    moduleName: 'fs/promises',
    settings: {
      'import/node-version': '16.0.0',
    },
  },
  {
    code: "\n      async function foo() {\n        const fs = await import('fs/promises');\n      }\n    ",
    output:
      "\n      async function foo() {\n        const fs = await import('node:fs/promises');\n      }\n    ",
    moduleName: 'fs/promises',
    settings: {
      'import/node-version': '16.0.0',
    },
  },
  {
    code: 'import {promises} from "fs";',
    output: 'import {promises} from "node:fs";',
    moduleName: 'fs',
  },
  {
    code: 'export {default as promises} from "fs";',
    output: 'export {default as promises} from "node:fs";',
    moduleName: 'fs',
  },
  {
    code: '\n      async function foo() {\n        const fs = await import("fs/promises");\n      }\n    ',
    output:
      '\n      async function foo() {\n        const fs = await import("node:fs/promises");\n      }\n    ',
    moduleName: 'fs/promises',
    settings: {
      'import/node-version': '16.0.0',
    },
  },
  {
    code: 'import "buffer";',
    output: 'import "node:buffer";',
    moduleName: 'buffer',
  },
  {
    code: 'import "child_process";',
    output: 'import "node:child_process";',
    moduleName: 'child_process',
  },
  {
    code: 'import "timers/promises";',
    output: 'import "node:timers/promises";',
    moduleName: 'timers/promises',
    settings: {
      'import/node-version': '16.0.0',
    },
  },
  {
    code: 'const {promises} = require("fs")',
    output: 'const {promises} = require("node:fs")',
    moduleName: 'fs',
  },
  {
    code: 'const fs = require("fs/promises")',
    output: 'const fs = require("node:fs/promises")',
    moduleName: 'fs/promises',
    settings: {
      'import/node-version': '16.0.0',
    },
  },
];

const ruleTester = new RuleTester();
ruleTester.run('enforce-node-protocol-usage', {} as never, {
  valid: [
    {
      code: 'import unicorn from "unicorn";',
      options: ['always'],
    },
    {
      code: 'import fs from "./fs";',
      options: ['always'],
    },
    {
      code: 'import fs from "unknown-builtin-module";',
      options: ['always'],
    },
    {
      code: 'import fs from "node:fs";',
      options: ['always'],
    },
    {
      code: '\n        async function foo() {\n          const fs = await import(fs);\n        }\n      ',
      options: ['always'],
    },
    {
      code: '\n        async function foo() {\n          const fs = await import(0);\n        }\n      ',
      options: ['always'],
    },
    {
      code: '\n        async function foo() {\n          const fs = await import(`fs`);\n        }\n      ',
      options: ['always'],
    },
    {
      code: 'import "punycode/";',
      options: ['always'],
    },
    {
      code: 'const fs = require("node:fs");',
      options: ['always'],
    },
    {
      code: 'const fs = require("node:fs/promises");',
      options: ['always'],
      settings: {
        'import/node-version': '16.0.0',
      },
    },
    {
      code: 'const fs = require(fs);',
      options: ['always'],
    },
    {
      code: 'const fs = notRequire("fs");',
      options: ['always'],
    },
    {
      code: 'const fs = foo.require("fs");',
      options: ['always'],
    },
    {
      code: 'const fs = require.resolve("fs");',
      options: ['always'],
    },
    {
      code: 'const fs = require(`fs`);',
      options: ['always'],
    },
    {
      code: 'const fs = require?.("fs");',
      options: ['always'],
    },
    {
      code: 'const fs = require("fs", extra);',
      options: ['always'],
    },
    {
      code: 'const fs = require();',
      options: ['always'],
    },
    {
      code: 'const fs = require(...["fs"]);',
      options: ['always'],
    },
    {
      code: 'const fs = require("unicorn");',
      options: ['always'],
    },
    {
      code: 'import fs from "fs";',
      options: ['never'],
    },
    {
      code: 'const fs = require("fs");',
      options: ['never'],
    },
    {
      code: 'const fs = require("fs/promises");',
      options: ['never'],
      settings: {
        'import/node-version': '16.0.0',
      },
    },
    {
      code: 'import "punycode/";',
      options: ['never'],
    },
    {
      code: 'const fs = require("node:test");',
      options: ['never'],
      settings: {
        'import/node-version': '16.0.0',
      },
    },
    {
      code: '\n            export class Thing {\n              constructor(public readonly name: string) {\n                  // Do nothing.\n              }\n\n              public sayHello(): void {\n                  console.log(`Hello, ${this.name}!`);\n              }\n            }\n          ',
      options: ['always'],
    },
  ],
  invalid: ['always', 'never'].flatMap((mode) =>
    invalidTemplates.map(({ code, output, moduleName, settings }) => ({
      code: mode === 'always' ? code : output,
      output: mode === 'always' ? output : code,
      options: [mode],
      settings:
        mode === 'always' ? settings : { 'import/node-version': '16.0.0' },
      errors: [
        {
          message:
            mode === 'always'
              ? `Prefer \`node:${moduleName}\` over \`${moduleName}\`.`
              : `Prefer \`${moduleName}\` over \`node:${moduleName}\`.`,
        },
      ],
    })),
  ),
});
