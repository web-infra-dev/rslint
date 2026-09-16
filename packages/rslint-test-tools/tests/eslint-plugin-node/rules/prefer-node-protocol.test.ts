// All cases in upstream order. Ranges and fixes were checked against the pinned reference.
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/prefer-node-protocol.js
import { RuleTester } from '../rule-tester';

function error(
  moduleName: string,
  line: number,
  column: number,
  endLine: number,
  endColumn: number,
) {
  return {
    messageId: 'preferNodeProtocol',
    message: `Prefer \`node:${moduleName}\` over \`${moduleName}\`.`,
    line,
    column,
    endLine,
    endColumn,
  };
}

new RuleTester({ languageOptions: { sourceType: 'module' } }).run(
  'prefer-node-protocol',
  undefined,
  {
    valid: [
      // Imports
      {
        name: 'upstream valid 1',
        code: 'import nodePlugin from "eslint-plugin-n";',
      },
      { name: 'upstream valid 2', code: 'import fs from "./fs";' },
      {
        name: 'upstream valid 3',
        code: 'import fs from "unknown-builtin-module";',
      },
      { name: 'upstream valid 4', code: 'import fs from "node:fs";' },
      {
        name: 'upstream valid 5',
        code: '\n            async function foo() {\n\t\t\t    const fs = await import(fs);\n\t\t    }\n        ',
      },
      {
        name: 'upstream valid 6',
        code: '\n            async function foo() {\n                const fs = await import(0);\n            }\n        ',
      },
      {
        name: 'upstream valid 7',
        code: '\n            async function foo() {\n                const fs = await import(`fs`);\n            }\n        ',
      },
      { name: 'upstream valid 8', code: 'import "punycode";' },
      { name: 'upstream valid 9', code: 'import "punycode/";' },
      { name: 'upstream valid 10', code: 'import "bun";' },
      { name: 'upstream valid 11', code: 'import "bun:jsc";' },
      { name: 'upstream valid 12', code: 'import "bun:sqlite";' },
      { name: 'upstream valid 13', code: 'export {promises} from "node:fs";' },
      // require
      { name: 'upstream valid 14', code: 'const fs = require("node:fs");' },
      {
        name: 'upstream valid 15',
        code: 'const fs = require("node:fs/promises");',
      },
      { name: 'upstream valid 16', code: 'const fs = require(fs);' },
      { name: 'upstream valid 17', code: 'const fs = notRequire("fs");' },
      { name: 'upstream valid 18', code: 'const fs = foo.require("fs");' },
      { name: 'upstream valid 19', code: 'const fs = require.resolve("fs");' },
      { name: 'upstream valid 20', code: 'const fs = require(`fs`);' },
      { name: 'upstream valid 21', code: 'const fs = require?.("fs");' },
      { name: 'upstream valid 22', code: 'const fs = require("fs", extra);' },
      { name: 'upstream valid 23', code: 'const fs = require();' },
      { name: 'upstream valid 24', code: 'const fs = require(...["fs"]);' },
      {
        name: 'upstream valid 25',
        code: 'const fs = require("eslint-plugin-n");',
      },
      // Unsupported Node versions
      {
        name: 'upstream valid 26',
        code: 'import fs from "fs";',
        options: [{ version: '>=10.13.0' }],
      },
      {
        name: 'upstream valid 27',
        code: 'import fs from "fs";',
        options: [{ version: '12.19.1' }],
      },
      {
        name: 'upstream valid 28',
        code: 'import fs from "fs";',
        options: [{ version: '13.14.0' }],
      },
      {
        name: 'upstream valid 29',
        code: 'import fs from "fs";',
        options: [{ version: '14.13.0' }],
      },
      {
        name: 'upstream valid 30',
        code: 'const fs = require("fs");',
        options: [{ version: '14.17.6' }],
      },
      {
        name: 'upstream valid 31',
        code: 'const fs = require("fs");',
        options: [{ version: '15.14.0' }],
      },
      // process.getBuiltinModule
      {
        name: 'upstream valid 32',
        code: 'const fs = process.getBuiltinModule("node:fs");',
      },
      {
        name: 'upstream valid 33',
        code: 'const fs = globalThis.process.getBuiltinModule("node:fs");',
      },
      {
        name: 'upstream valid 34',
        code: 'const fs = process.getBuiltinModule("node:fs/promises");',
      },
      {
        name: 'upstream valid 35',
        code: 'const fs = process.getNotBuiltinModule("node:fs", extra);',
      },
      {
        name: 'upstream valid 36',
        code: 'const fs = process.getBuiltinModule(fs);',
      },
      {
        name: 'upstream valid 37',
        code: 'const fs = process.getNotBuiltinModule("fs");',
      },
      {
        name: 'upstream valid 38',
        code: 'const fs = process.foo.getNotBuiltinModule("fs");',
      },
      {
        name: 'upstream valid 39',
        code: 'const fs = foo.process.getNotBuiltinModule("fs");',
      },
      {
        name: 'upstream valid 40',
        code: 'const fs = process.getNotBuiltinModule.foo("fs");',
      },
      {
        name: 'upstream valid 41',
        code: 'const fs = process.getNotBuiltinModule(`fs`);',
      },
      {
        name: 'upstream valid 42',
        code: 'const fs = process.getNotBuiltinModule();',
      },
      {
        name: 'upstream valid 43',
        code: 'const fs = process.getNotBuiltinModule(...["fs"]);',
      },
      {
        name: 'upstream valid 44',
        code: 'const fs = process.getNotBuiltinModule("eslint-plugin-n");',
      },
      {
        name: 'upstream valid 45',
        code: 'const fs = process.getBuiltinModule("node:fs");',
        options: [{ version: '12.19.1' }],
      },
    ],
    invalid: [
      // Imports
      {
        name: 'upstream invalid 1',
        code: 'import fs from "fs";',
        output: 'import fs from "node:fs";',
        errors: [error('fs', 1, 16, 1, 20)],
      },
      {
        name: 'upstream invalid 2',
        code: 'export {promises} from "fs";',
        output: 'export {promises} from "node:fs";',
        errors: [error('fs', 1, 24, 1, 28)],
      },
      {
        name: 'upstream invalid 3',
        code: "\n                async function foo() {\n                    const fs = await import('fs');\n                }\n            ",
        output:
          "\n                async function foo() {\n                    const fs = await import('node:fs');\n                }\n            ",
        errors: [error('fs', 3, 45, 3, 49)],
      },
      {
        name: 'upstream invalid 4',
        code: 'import fs from "fs/promises";',
        output: 'import fs from "node:fs/promises";',
        errors: [error('fs/promises', 1, 16, 1, 29)],
      },
      {
        name: 'upstream invalid 5',
        code: 'export {default} from "fs/promises";',
        output: 'export {default} from "node:fs/promises";',
        errors: [error('fs/promises', 1, 23, 1, 36)],
      },
      {
        name: 'upstream invalid 6',
        code: "\n                async function foo() {\n                    const fs = await import('fs/promises');\n                }\n            ",
        output:
          "\n                async function foo() {\n                    const fs = await import('node:fs/promises');\n                }\n            ",
        errors: [error('fs/promises', 3, 45, 3, 58)],
      },
      {
        name: 'upstream invalid 7',
        code: 'import {promises} from "fs";',
        output: 'import {promises} from "node:fs";',
        errors: [error('fs', 1, 24, 1, 28)],
      },
      {
        name: 'upstream invalid 8',
        code: 'export {default as promises} from "fs";',
        output: 'export {default as promises} from "node:fs";',
        errors: [error('fs', 1, 35, 1, 39)],
      },
      {
        name: 'upstream invalid 9',
        code: "import {promises} from 'fs';",
        output: "import {promises} from 'node:fs';",
        errors: [error('fs', 1, 24, 1, 28)],
      },
      {
        name: 'upstream invalid 10',
        code: '\n                async function foo() {\n                    const fs = await import("fs/promises");\n                }\n            ',
        output:
          '\n                async function foo() {\n                    const fs = await import("node:fs/promises");\n                }\n            ',
        errors: [error('fs/promises', 3, 45, 3, 58)],
      },
      {
        name: 'upstream invalid 11',
        code: '\n                async function foo() {\n                    const fs = await import(/* escaped */"\\u{66}s/promises");\n                }\n            ',
        output:
          '\n                async function foo() {\n                    const fs = await import(/* escaped */"node:\\u{66}s/promises");\n                }\n            ',
        errors: [error('fs/promises', 3, 58, 3, 76)],
      },
      {
        name: 'upstream invalid 12',
        code: 'import "buffer";',
        output: 'import "node:buffer";',
        errors: [error('buffer', 1, 8, 1, 16)],
      },
      {
        name: 'upstream invalid 13',
        code: 'import "child_process";',
        output: 'import "node:child_process";',
        errors: [error('child_process', 1, 8, 1, 23)],
      },
      {
        name: 'upstream invalid 14',
        code: 'import "timers/promises";',
        output: 'import "node:timers/promises";',
        errors: [error('timers/promises', 1, 8, 1, 25)],
      },
      // require
      {
        name: 'upstream invalid 15',
        code: 'const {promises} = require("fs")',
        output: 'const {promises} = require("node:fs")',
        errors: [error('fs', 1, 28, 1, 32)],
      },
      {
        name: 'upstream invalid 16',
        code: "const fs = require('fs/promises')",
        output: "const fs = require('node:fs/promises')",
        errors: [error('fs/promises', 1, 20, 1, 33)],
      },
      {
        name: 'upstream invalid 17',
        code: "\n                const express = require('express');\n                const fs = require('fs/promises');\n            ",
        output:
          "\n                const express = require('express');\n                const fs = require('node:fs/promises');\n            ",
        errors: [error('fs/promises', 3, 36, 3, 49)],
      },
      // Supported Node versions
      {
        name: 'upstream invalid 18',
        code: 'import fs from "fs";',
        options: [{ version: '12.20.0' }],
        output: 'import fs from "node:fs";',
        errors: [error('fs', 1, 16, 1, 20)],
      },
      {
        name: 'upstream invalid 19',
        code: 'import fs from "fs";',
        options: [{ version: '14.13.1' }],
        output: 'import fs from "node:fs";',
        errors: [error('fs', 1, 16, 1, 20)],
      },
      {
        name: 'upstream invalid 20',
        code: 'const fs = require("fs");',
        options: [{ version: '14.18.0' }],
        output: 'const fs = require("node:fs");',
        errors: [error('fs', 1, 20, 1, 24)],
      },
      {
        name: 'upstream invalid 21',
        code: 'const fs = require("fs");',
        options: [{ version: '16.0.0' }],
        output: 'const fs = require("node:fs");',
        errors: [error('fs', 1, 20, 1, 24)],
      },
      {
        name: 'upstream invalid 22',
        code: '\n                const fs = require("fs");\n                import buffer from \'buffer\'\n            ',
        options: [{ version: '12.20.0' }],
        output:
          '\n                const fs = require("fs");\n                import buffer from \'node:buffer\'\n            ',
        errors: [error('buffer', 3, 36, 3, 44)],
      },
      // Regression: upstream issue #431
      {
        name: 'upstream invalid 23',
        code: 'import https from "https";',
        output: 'import https from "node:https";',
        errors: [error('https', 1, 19, 1, 26)],
      },
      // process.getBuiltinModule
      {
        name: 'upstream invalid 24',
        code: 'const {promises} = process.getBuiltinModule("fs")',
        output: 'const {promises} = process.getBuiltinModule("node:fs")',
        errors: [error('fs', 1, 45, 1, 49)],
      },
      {
        name: 'upstream invalid 25',
        code: 'const {promises} = globalThis.process.getBuiltinModule("fs")',
        output:
          'const {promises} = globalThis.process.getBuiltinModule("node:fs")',
        errors: [error('fs', 1, 56, 1, 60)],
      },
      {
        name: 'upstream invalid 26',
        code: 'const {promises} = process.getBuiltinModule("fs", extra)',
        output: 'const {promises} = process.getBuiltinModule("node:fs", extra)',
        errors: [error('fs', 1, 45, 1, 49)],
      },
      {
        name: 'upstream invalid 27',
        code: "const fs = process.getBuiltinModule('fs/promises')",
        output: "const fs = process.getBuiltinModule('node:fs/promises')",
        errors: [error('fs/promises', 1, 37, 1, 50)],
      },
      {
        name: 'upstream invalid 28',
        code: "\n                const express = process.getBuiltinModule('express');\n                const fs = process.getBuiltinModule('fs/promises');\n            ",
        output:
          "\n                const express = process.getBuiltinModule('express');\n                const fs = process.getBuiltinModule('node:fs/promises');\n            ",
        errors: [error('fs/promises', 3, 53, 3, 66)],
      },
      {
        name: 'upstream invalid 29',
        code: 'const {promises} = process.getBuiltinModule("fs")',
        options: [{ version: '12.19.1' }],
        output: 'const {promises} = process.getBuiltinModule("node:fs")',
        errors: [error('fs', 1, 45, 1, 49)],
      },
    ],
  },
);
