// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
new RuleTester(fixtureConfig()).run('enforce-node-protocol-usage', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: 'import unicorn from "unicorn";',
      options: ['always'],
    },
    {
      name: 'run 0 valid 1',
      code: 'import fs from "./fs";',
      options: ['always'],
    },
    {
      name: 'run 0 valid 2',
      code: 'import fs from "unknown-builtin-module";',
      options: ['always'],
    },
    {
      name: 'run 0 valid 3',
      code: 'import fs from "node:fs";',
      options: ['always'],
    },
    {
      name: 'run 0 valid 4',
      code: '\n        async function foo() {\n          const fs = await import(fs);\n        }\n      ',
      options: ['always'],
    },
    {
      name: 'run 0 valid 5',
      code: '\n        async function foo() {\n          const fs = await import(0);\n        }\n      ',
      options: ['always'],
    },
    {
      name: 'run 0 valid 6',
      code: '\n        async function foo() {\n          const fs = await import(`fs`);\n        }\n      ',
      options: ['always'],
    },
    {
      name: 'run 0 valid 7',
      code: 'import "punycode/";',
      options: ['always'],
    },
    {
      name: 'run 0 valid 8',
      code: 'const fs = require("node:fs");',
      options: ['always'],
    },
    {
      name: 'run 0 valid 9',
      code: 'const fs = require("node:fs/promises");',
      options: ['always'],
      settings: {
        'import/node-version': '16.0.0',
      },
    },
    {
      name: 'run 0 valid 10',
      code: 'const fs = require(fs);',
      options: ['always'],
    },
    {
      name: 'run 0 valid 11',
      code: 'const fs = notRequire("fs");',
      options: ['always'],
    },
    {
      name: 'run 0 valid 12',
      code: 'const fs = foo.require("fs");',
      options: ['always'],
    },
    {
      name: 'run 0 valid 13',
      code: 'const fs = require.resolve("fs");',
      options: ['always'],
    },
    {
      name: 'run 0 valid 14',
      code: 'const fs = require(`fs`);',
      options: ['always'],
    },
    {
      name: 'run 0 valid 15',
      code: 'const fs = require?.("fs");',
      options: ['always'],
    },
    {
      name: 'run 0 valid 16',
      code: 'const fs = require("fs", extra);',
      options: ['always'],
    },
    {
      name: 'run 0 valid 17',
      code: 'const fs = require();',
      options: ['always'],
    },
    {
      name: 'run 0 valid 18',
      code: 'const fs = require(...["fs"]);',
      options: ['always'],
    },
    {
      name: 'run 0 valid 19',
      code: 'const fs = require("unicorn");',
      options: ['always'],
    },
    {
      name: 'run 0 valid 20',
      code: 'import fs from "fs";',
      options: ['never'],
    },
    {
      name: 'run 0 valid 21',
      code: 'const fs = require("fs");',
      options: ['never'],
    },
    {
      name: 'run 0 valid 22',
      code: 'const fs = require("fs/promises");',
      options: ['never'],
      settings: {
        'import/node-version': '16.0.0',
      },
    },
    {
      name: 'run 0 valid 23',
      code: 'import "punycode/";',
      options: ['never'],
    },
    {
      name: 'run 0 valid 24',
      code: 'const fs = require("node:test");',
      options: ['never'],
      settings: {
        'import/node-version': '16.0.0',
      },
    },
    {
      name: 'run 0 valid 25',
      code: '\n            export class Thing {\n              constructor(public readonly name: string) {\n                  // Do nothing.\n              }\n\n              public sayHello(): void {\n                  console.log(`Hello, ${this.name}!`);\n              }\n            }\n          ',
      options: ['always'],
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: 'import x from "fs";',
      options: ['always'],
      errors: [
        {
          messageId: '',
          message: 'Prefer `node:fs` over `fs`.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: 'import x from "node:fs";',
    },
    {
      name: 'run 0 invalid 1',
      code: 'export {promises} from "fs";',
      options: ['always'],
      errors: [
        {
          messageId: '',
          message: 'Prefer `node:fs` over `fs`.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 28,
        },
      ],
      output: 'export {promises} from "node:fs";',
    },
    {
      name: 'run 0 invalid 2',
      code: "\n        async function foo() {\n          const x = await import('fs');\n        }\n      ",
      options: ['always'],
      errors: [
        {
          messageId: '',
          message: 'Prefer `node:fs` over `fs`.',
          line: 3,
          column: 34,
          endLine: 3,
          endColumn: 38,
        },
      ],
      output:
        "\n        async function foo() {\n          const x = await import('node:fs');\n        }\n      ",
    },
    {
      name: 'run 0 invalid 3',
      code: 'import x from "fs/promises";',
      options: ['always'],
      errors: [
        {
          messageId: '',
          message: 'Prefer `node:fs/promises` over `fs/promises`.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 28,
        },
      ],
      output: 'import x from "node:fs/promises";',
    },
    {
      name: 'run 0 invalid 4',
      code: 'export {promises} from "fs/promises";',
      options: ['always'],
      errors: [
        {
          messageId: '',
          message: 'Prefer `node:fs/promises` over `fs/promises`.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 37,
        },
      ],
      output: 'export {promises} from "node:fs/promises";',
    },
    {
      name: 'run 0 invalid 5',
      code: "\n        async function foo() {\n          const x = await import('fs/promises');\n        }\n      ",
      options: ['always'],
      errors: [
        {
          messageId: '',
          message: 'Prefer `node:fs/promises` over `fs/promises`.',
          line: 3,
          column: 34,
          endLine: 3,
          endColumn: 47,
        },
      ],
      output:
        "\n        async function foo() {\n          const x = await import('node:fs/promises');\n        }\n      ",
    },
    {
      name: 'run 0 invalid 6',
      code: 'import x from "buffer";',
      options: ['always'],
      errors: [
        {
          messageId: '',
          message: 'Prefer `node:buffer` over `buffer`.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 23,
        },
      ],
      output: 'import x from "node:buffer";',
    },
    {
      name: 'run 0 invalid 7',
      code: 'export {promises} from "buffer";',
      options: ['always'],
      errors: [
        {
          messageId: '',
          message: 'Prefer `node:buffer` over `buffer`.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 32,
        },
      ],
      output: 'export {promises} from "node:buffer";',
    },
    {
      name: 'run 0 invalid 8',
      code: "\n        async function foo() {\n          const x = await import('buffer');\n        }\n      ",
      options: ['always'],
      errors: [
        {
          messageId: '',
          message: 'Prefer `node:buffer` over `buffer`.',
          line: 3,
          column: 34,
          endLine: 3,
          endColumn: 42,
        },
      ],
      output:
        "\n        async function foo() {\n          const x = await import('node:buffer');\n        }\n      ",
    },
    {
      name: 'run 0 invalid 9',
      code: 'import x from "child_process";',
      options: ['always'],
      errors: [
        {
          messageId: '',
          message: 'Prefer `node:child_process` over `child_process`.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 30,
        },
      ],
      output: 'import x from "node:child_process";',
    },
    {
      name: 'run 0 invalid 10',
      code: 'export {promises} from "child_process";',
      options: ['always'],
      errors: [
        {
          messageId: '',
          message: 'Prefer `node:child_process` over `child_process`.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 39,
        },
      ],
      output: 'export {promises} from "node:child_process";',
    },
    {
      name: 'run 0 invalid 11',
      code: "\n        async function foo() {\n          const x = await import('child_process');\n        }\n      ",
      options: ['always'],
      errors: [
        {
          messageId: '',
          message: 'Prefer `node:child_process` over `child_process`.',
          line: 3,
          column: 34,
          endLine: 3,
          endColumn: 49,
        },
      ],
      output:
        "\n        async function foo() {\n          const x = await import('node:child_process');\n        }\n      ",
    },
    {
      name: 'run 0 invalid 12',
      code: 'import x from "timers/promises";',
      options: ['always'],
      errors: [
        {
          messageId: '',
          message: 'Prefer `node:timers/promises` over `timers/promises`.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 32,
        },
      ],
      output: 'import x from "node:timers/promises";',
    },
    {
      name: 'run 0 invalid 13',
      code: 'export {promises} from "timers/promises";',
      options: ['always'],
      errors: [
        {
          messageId: '',
          message: 'Prefer `node:timers/promises` over `timers/promises`.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 41,
        },
      ],
      output: 'export {promises} from "node:timers/promises";',
    },
    {
      name: 'run 0 invalid 14',
      code: "\n        async function foo() {\n          const x = await import('timers/promises');\n        }\n      ",
      options: ['always'],
      errors: [
        {
          messageId: '',
          message: 'Prefer `node:timers/promises` over `timers/promises`.',
          line: 3,
          column: 34,
          endLine: 3,
          endColumn: 51,
        },
      ],
      output:
        "\n        async function foo() {\n          const x = await import('node:timers/promises');\n        }\n      ",
    },
    {
      name: 'run 0 invalid 15',
      code: 'import fs from "fs/promises";',
      options: ['always'],
      settings: {
        'import/node-version': '16.0.0',
      },
      errors: [
        {
          messageId: '',
          message: 'Prefer `node:fs/promises` over `fs/promises`.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 29,
        },
      ],
      output: 'import fs from "node:fs/promises";',
    },
    {
      name: 'run 0 invalid 16',
      code: 'export {default} from "fs/promises";',
      options: ['always'],
      settings: {
        'import/node-version': '16.0.0',
      },
      errors: [
        {
          messageId: '',
          message: 'Prefer `node:fs/promises` over `fs/promises`.',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 36,
        },
      ],
      output: 'export {default} from "node:fs/promises";',
    },
    {
      name: 'run 0 invalid 17',
      code: "\n      async function foo() {\n        const fs = await import('fs/promises');\n      }\n    ",
      options: ['always'],
      settings: {
        'import/node-version': '16.0.0',
      },
      errors: [
        {
          messageId: '',
          message: 'Prefer `node:fs/promises` over `fs/promises`.',
          line: 3,
          column: 33,
          endLine: 3,
          endColumn: 46,
        },
      ],
      output:
        "\n      async function foo() {\n        const fs = await import('node:fs/promises');\n      }\n    ",
    },
    {
      name: 'run 0 invalid 18',
      code: 'import {promises} from "fs";',
      options: ['always'],
      errors: [
        {
          messageId: '',
          message: 'Prefer `node:fs` over `fs`.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 28,
        },
      ],
      output: 'import {promises} from "node:fs";',
    },
    {
      name: 'run 0 invalid 19',
      code: 'export {default as promises} from "fs";',
      options: ['always'],
      errors: [
        {
          messageId: '',
          message: 'Prefer `node:fs` over `fs`.',
          line: 1,
          column: 35,
          endLine: 1,
          endColumn: 39,
        },
      ],
      output: 'export {default as promises} from "node:fs";',
    },
    {
      name: 'run 0 invalid 20',
      code: '\n      async function foo() {\n        const fs = await import("fs/promises");\n      }\n    ',
      options: ['always'],
      settings: {
        'import/node-version': '16.0.0',
      },
      errors: [
        {
          messageId: '',
          message: 'Prefer `node:fs/promises` over `fs/promises`.',
          line: 3,
          column: 33,
          endLine: 3,
          endColumn: 46,
        },
      ],
      output:
        '\n      async function foo() {\n        const fs = await import("node:fs/promises");\n      }\n    ',
    },
    {
      name: 'run 0 invalid 21',
      code: 'import "buffer";',
      options: ['always'],
      errors: [
        {
          messageId: '',
          message: 'Prefer `node:buffer` over `buffer`.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 16,
        },
      ],
      output: 'import "node:buffer";',
    },
    {
      name: 'run 0 invalid 22',
      code: 'import "child_process";',
      options: ['always'],
      errors: [
        {
          messageId: '',
          message: 'Prefer `node:child_process` over `child_process`.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 23,
        },
      ],
      output: 'import "node:child_process";',
    },
    {
      name: 'run 0 invalid 23',
      code: 'import "timers/promises";',
      options: ['always'],
      settings: {
        'import/node-version': '16.0.0',
      },
      errors: [
        {
          messageId: '',
          message: 'Prefer `node:timers/promises` over `timers/promises`.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 25,
        },
      ],
      output: 'import "node:timers/promises";',
    },
    {
      name: 'run 0 invalid 24',
      code: 'const {promises} = require("fs")',
      options: ['always'],
      errors: [
        {
          messageId: '',
          message: 'Prefer `node:fs` over `fs`.',
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 32,
        },
      ],
      output: 'const {promises} = require("node:fs")',
    },
    {
      name: 'run 0 invalid 25',
      code: 'const fs = require("fs/promises")',
      options: ['always'],
      settings: {
        'import/node-version': '16.0.0',
      },
      errors: [
        {
          messageId: '',
          message: 'Prefer `node:fs/promises` over `fs/promises`.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 33,
        },
      ],
      output: 'const fs = require("node:fs/promises")',
    },
    {
      name: 'run 0 invalid 26',
      code: 'import x from "node:fs";',
      options: ['never'],
      settings: {
        'import/node-version': '16.0.0',
      },
      errors: [
        {
          messageId: '',
          message: 'Prefer `fs` over `node:fs`.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 24,
        },
      ],
      output: 'import x from "fs";',
    },
    {
      name: 'run 0 invalid 27',
      code: 'export {promises} from "node:fs";',
      options: ['never'],
      settings: {
        'import/node-version': '16.0.0',
      },
      errors: [
        {
          messageId: '',
          message: 'Prefer `fs` over `node:fs`.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 33,
        },
      ],
      output: 'export {promises} from "fs";',
    },
    {
      name: 'run 0 invalid 28',
      code: "\n        async function foo() {\n          const x = await import('node:fs');\n        }\n      ",
      options: ['never'],
      settings: {
        'import/node-version': '16.0.0',
      },
      errors: [
        {
          messageId: '',
          message: 'Prefer `fs` over `node:fs`.',
          line: 3,
          column: 34,
          endLine: 3,
          endColumn: 43,
        },
      ],
      output:
        "\n        async function foo() {\n          const x = await import('fs');\n        }\n      ",
    },
    {
      name: 'run 0 invalid 29',
      code: 'import x from "node:fs/promises";',
      options: ['never'],
      settings: {
        'import/node-version': '16.0.0',
      },
      errors: [
        {
          messageId: '',
          message: 'Prefer `fs/promises` over `node:fs/promises`.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 33,
        },
      ],
      output: 'import x from "fs/promises";',
    },
    {
      name: 'run 0 invalid 30',
      code: 'export {promises} from "node:fs/promises";',
      options: ['never'],
      settings: {
        'import/node-version': '16.0.0',
      },
      errors: [
        {
          messageId: '',
          message: 'Prefer `fs/promises` over `node:fs/promises`.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 42,
        },
      ],
      output: 'export {promises} from "fs/promises";',
    },
    {
      name: 'run 0 invalid 31',
      code: "\n        async function foo() {\n          const x = await import('node:fs/promises');\n        }\n      ",
      options: ['never'],
      settings: {
        'import/node-version': '16.0.0',
      },
      errors: [
        {
          messageId: '',
          message: 'Prefer `fs/promises` over `node:fs/promises`.',
          line: 3,
          column: 34,
          endLine: 3,
          endColumn: 52,
        },
      ],
      output:
        "\n        async function foo() {\n          const x = await import('fs/promises');\n        }\n      ",
    },
    {
      name: 'run 0 invalid 32',
      code: 'import x from "node:buffer";',
      options: ['never'],
      settings: {
        'import/node-version': '16.0.0',
      },
      errors: [
        {
          messageId: '',
          message: 'Prefer `buffer` over `node:buffer`.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 28,
        },
      ],
      output: 'import x from "buffer";',
    },
    {
      name: 'run 0 invalid 33',
      code: 'export {promises} from "node:buffer";',
      options: ['never'],
      settings: {
        'import/node-version': '16.0.0',
      },
      errors: [
        {
          messageId: '',
          message: 'Prefer `buffer` over `node:buffer`.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 37,
        },
      ],
      output: 'export {promises} from "buffer";',
    },
    {
      name: 'run 0 invalid 34',
      code: "\n        async function foo() {\n          const x = await import('node:buffer');\n        }\n      ",
      options: ['never'],
      settings: {
        'import/node-version': '16.0.0',
      },
      errors: [
        {
          messageId: '',
          message: 'Prefer `buffer` over `node:buffer`.',
          line: 3,
          column: 34,
          endLine: 3,
          endColumn: 47,
        },
      ],
      output:
        "\n        async function foo() {\n          const x = await import('buffer');\n        }\n      ",
    },
    {
      name: 'run 0 invalid 35',
      code: 'import x from "node:child_process";',
      options: ['never'],
      settings: {
        'import/node-version': '16.0.0',
      },
      errors: [
        {
          messageId: '',
          message: 'Prefer `child_process` over `node:child_process`.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 35,
        },
      ],
      output: 'import x from "child_process";',
    },
    {
      name: 'run 0 invalid 36',
      code: 'export {promises} from "node:child_process";',
      options: ['never'],
      settings: {
        'import/node-version': '16.0.0',
      },
      errors: [
        {
          messageId: '',
          message: 'Prefer `child_process` over `node:child_process`.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 44,
        },
      ],
      output: 'export {promises} from "child_process";',
    },
    {
      name: 'run 0 invalid 37',
      code: "\n        async function foo() {\n          const x = await import('node:child_process');\n        }\n      ",
      options: ['never'],
      settings: {
        'import/node-version': '16.0.0',
      },
      errors: [
        {
          messageId: '',
          message: 'Prefer `child_process` over `node:child_process`.',
          line: 3,
          column: 34,
          endLine: 3,
          endColumn: 54,
        },
      ],
      output:
        "\n        async function foo() {\n          const x = await import('child_process');\n        }\n      ",
    },
    {
      name: 'run 0 invalid 38',
      code: 'import x from "node:timers/promises";',
      options: ['never'],
      settings: {
        'import/node-version': '16.0.0',
      },
      errors: [
        {
          messageId: '',
          message: 'Prefer `timers/promises` over `node:timers/promises`.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 37,
        },
      ],
      output: 'import x from "timers/promises";',
    },
    {
      name: 'run 0 invalid 39',
      code: 'export {promises} from "node:timers/promises";',
      options: ['never'],
      settings: {
        'import/node-version': '16.0.0',
      },
      errors: [
        {
          messageId: '',
          message: 'Prefer `timers/promises` over `node:timers/promises`.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 46,
        },
      ],
      output: 'export {promises} from "timers/promises";',
    },
    {
      name: 'run 0 invalid 40',
      code: "\n        async function foo() {\n          const x = await import('node:timers/promises');\n        }\n      ",
      options: ['never'],
      settings: {
        'import/node-version': '16.0.0',
      },
      errors: [
        {
          messageId: '',
          message: 'Prefer `timers/promises` over `node:timers/promises`.',
          line: 3,
          column: 34,
          endLine: 3,
          endColumn: 56,
        },
      ],
      output:
        "\n        async function foo() {\n          const x = await import('timers/promises');\n        }\n      ",
    },
    {
      name: 'run 0 invalid 41',
      code: 'import fs from "node:fs/promises";',
      options: ['never'],
      settings: {
        'import/node-version': '16.0.0',
      },
      errors: [
        {
          messageId: '',
          message: 'Prefer `fs/promises` over `node:fs/promises`.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 34,
        },
      ],
      output: 'import fs from "fs/promises";',
    },
    {
      name: 'run 0 invalid 42',
      code: 'export {default} from "node:fs/promises";',
      options: ['never'],
      settings: {
        'import/node-version': '16.0.0',
      },
      errors: [
        {
          messageId: '',
          message: 'Prefer `fs/promises` over `node:fs/promises`.',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 41,
        },
      ],
      output: 'export {default} from "fs/promises";',
    },
    {
      name: 'run 0 invalid 43',
      code: "\n      async function foo() {\n        const fs = await import('node:fs/promises');\n      }\n    ",
      options: ['never'],
      settings: {
        'import/node-version': '16.0.0',
      },
      errors: [
        {
          messageId: '',
          message: 'Prefer `fs/promises` over `node:fs/promises`.',
          line: 3,
          column: 33,
          endLine: 3,
          endColumn: 51,
        },
      ],
      output:
        "\n      async function foo() {\n        const fs = await import('fs/promises');\n      }\n    ",
    },
    {
      name: 'run 0 invalid 44',
      code: 'import {promises} from "node:fs";',
      options: ['never'],
      settings: {
        'import/node-version': '16.0.0',
      },
      errors: [
        {
          messageId: '',
          message: 'Prefer `fs` over `node:fs`.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 33,
        },
      ],
      output: 'import {promises} from "fs";',
    },
    {
      name: 'run 0 invalid 45',
      code: 'export {default as promises} from "node:fs";',
      options: ['never'],
      settings: {
        'import/node-version': '16.0.0',
      },
      errors: [
        {
          messageId: '',
          message: 'Prefer `fs` over `node:fs`.',
          line: 1,
          column: 35,
          endLine: 1,
          endColumn: 44,
        },
      ],
      output: 'export {default as promises} from "fs";',
    },
    {
      name: 'run 0 invalid 46',
      code: '\n      async function foo() {\n        const fs = await import("node:fs/promises");\n      }\n    ',
      options: ['never'],
      settings: {
        'import/node-version': '16.0.0',
      },
      errors: [
        {
          messageId: '',
          message: 'Prefer `fs/promises` over `node:fs/promises`.',
          line: 3,
          column: 33,
          endLine: 3,
          endColumn: 51,
        },
      ],
      output:
        '\n      async function foo() {\n        const fs = await import("fs/promises");\n      }\n    ',
    },
    {
      name: 'run 0 invalid 47',
      code: 'import "node:buffer";',
      options: ['never'],
      settings: {
        'import/node-version': '16.0.0',
      },
      errors: [
        {
          messageId: '',
          message: 'Prefer `buffer` over `node:buffer`.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 21,
        },
      ],
      output: 'import "buffer";',
    },
    {
      name: 'run 0 invalid 48',
      code: 'import "node:child_process";',
      options: ['never'],
      settings: {
        'import/node-version': '16.0.0',
      },
      errors: [
        {
          messageId: '',
          message: 'Prefer `child_process` over `node:child_process`.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 28,
        },
      ],
      output: 'import "child_process";',
    },
    {
      name: 'run 0 invalid 49',
      code: 'import "node:timers/promises";',
      options: ['never'],
      settings: {
        'import/node-version': '16.0.0',
      },
      errors: [
        {
          messageId: '',
          message: 'Prefer `timers/promises` over `node:timers/promises`.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 30,
        },
      ],
      output: 'import "timers/promises";',
    },
    {
      name: 'run 0 invalid 50',
      code: 'const {promises} = require("node:fs")',
      options: ['never'],
      settings: {
        'import/node-version': '16.0.0',
      },
      errors: [
        {
          messageId: '',
          message: 'Prefer `fs` over `node:fs`.',
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 37,
        },
      ],
      output: 'const {promises} = require("fs")',
    },
    {
      name: 'run 0 invalid 51',
      code: 'const fs = require("node:fs/promises")',
      options: ['never'],
      settings: {
        'import/node-version': '16.0.0',
      },
      errors: [
        {
          messageId: '',
          message: 'Prefer `fs/promises` over `node:fs/promises`.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 38,
        },
      ],
      output: 'const fs = require("fs/promises")',
    },
  ],
});
