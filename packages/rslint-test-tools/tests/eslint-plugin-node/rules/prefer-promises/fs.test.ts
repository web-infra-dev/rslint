// Every upstream test and documentation example from eslint-plugin-n v18.3.0.
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/prefer-promises/fs.js
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/prefer-promises/fs.md
import { RuleTester } from '../../rule-tester';

new RuleTester({
  languageOptions: {
    sourceType: 'module',
    globals: { require: 'readonly', process: 'readonly' },
  },
  fixtureFiles: { 'package.json': '{}' },
}).run(
  'prefer-promises/fs',
  {},
  {
    valid: [
      {
        name: 'Upstream valid 1',
        code: "const fs = require('fs'); fs.createReadStream()",
      },
      {
        name: 'Upstream valid 2',
        code: "const fs = require('fs'); fs.accessSync()",
      },
      {
        name: 'Upstream valid 3',
        code: "const fs = require('fs'); fs.promises.access()",
      },
      {
        name: 'Upstream valid 4',
        code: "const fs = require('node:fs'); fs.promises.access()",
      },
      {
        name: 'Upstream valid 5',
        code: "const {promises} = require('fs'); promises.access()",
      },
      {
        name: 'Upstream valid 6',
        code: "const {promises: fs} = require('fs'); fs.access()",
      },
      {
        name: 'Upstream valid 7',
        code: "const {promises: {access}} = require('fs'); access()",
      },
      {
        name: 'Upstream valid 8',
        code: "import fs from 'fs'; fs.promises.access()",
      },
      {
        name: 'Upstream valid 9',
        code: "import fs from 'node:fs'; fs.promises.access()",
      },
      {
        name: 'Upstream valid 10',
        code: "import * as fs from 'fs'; fs.promises.access()",
      },
      {
        name: 'Upstream valid 11',
        code: "import {promises} from 'fs'; promises.access()",
      },
      {
        name: 'Upstream valid 12',
        code: "import {promises as fs} from 'fs'; fs.access()",
      },
      {
        name: 'Upstream valid 13',
        code: "const fs = process.getBuiltinModule('fs'); fs.promises.access()",
      },
      {
        name: 'Upstream valid 14',
        code: "const fs = process.getBuiltinModule('node:fs'); fs.promises.access()",
      },
      {
        name: 'Upstream valid 15',
        code: "const {promises} = process.getBuiltinModule('fs'); promises.access()",
      },
      {
        name: 'Upstream valid 16',
        code: "const {promises: fs} = process.getBuiltinModule('fs'); fs.access()",
      },
      {
        name: 'Documentation example 3',
        code: 'const { promises: fs } = require("fs")\n\nasync function readData(filePath) {\n    const content = await fs.readFile(filePath, "utf8")\n    //...\n}',
      },
      {
        name: 'Documentation example 4',
        code: 'import { promises as fs } from "fs"\n\nasync function readData(filePath) {\n    const content = await fs.readFile(filePath, "utf8")\n    //...\n}',
      },
    ],
    invalid: [
      {
        name: 'Upstream invalid 1',
        code: "const fs = require('fs'); fs.access()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.access()' instead.",
            line: 1,
            column: 27,
            endLine: 1,
            endColumn: 38,
          },
        ],
      },
      {
        name: 'Upstream invalid 2',
        code: "const fs = require('node:fs'); fs.access()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.access()' instead.",
            line: 1,
            column: 32,
            endLine: 1,
            endColumn: 43,
          },
        ],
      },
      {
        name: 'Upstream invalid 3',
        code: "const {access} = require('fs'); access()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.access()' instead.",
            line: 1,
            column: 33,
            endLine: 1,
            endColumn: 41,
          },
        ],
      },
      {
        name: 'Upstream invalid 4',
        code: "import fs from 'fs'; fs.access()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.access()' instead.",
            line: 1,
            column: 22,
            endLine: 1,
            endColumn: 33,
          },
        ],
      },
      {
        name: 'Upstream invalid 5',
        code: "import fs from 'node:fs'; fs.access()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.access()' instead.",
            line: 1,
            column: 27,
            endLine: 1,
            endColumn: 38,
          },
        ],
      },
      {
        name: 'Upstream invalid 6',
        code: "import * as fs from 'fs'; fs.access()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.access()' instead.",
            line: 1,
            column: 27,
            endLine: 1,
            endColumn: 38,
          },
        ],
      },
      {
        name: 'Upstream invalid 7',
        code: "import {access} from 'fs'; access()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.access()' instead.",
            line: 1,
            column: 28,
            endLine: 1,
            endColumn: 36,
          },
        ],
      },
      {
        name: 'Upstream invalid 8',
        code: "const fs = process.getBuiltinModule('fs'); fs.access()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.access()' instead.",
            line: 1,
            column: 44,
            endLine: 1,
            endColumn: 55,
          },
        ],
      },
      {
        name: 'Upstream invalid 9',
        code: "const fs = process.getBuiltinModule('node:fs'); fs.access()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.access()' instead.",
            line: 1,
            column: 49,
            endLine: 1,
            endColumn: 60,
          },
        ],
      },
      {
        name: 'Upstream invalid 10',
        code: "const {access} = process.getBuiltinModule('fs'); access()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.access()' instead.",
            line: 1,
            column: 50,
            endLine: 1,
            endColumn: 58,
          },
        ],
      },
      {
        name: 'Upstream invalid 11',
        code: "const fs = require('fs'); fs.copyFile()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.copyFile()' instead.",
            line: 1,
            column: 27,
            endLine: 1,
            endColumn: 40,
          },
        ],
      },
      {
        name: 'Upstream invalid 12',
        code: "const fs = require('fs'); fs.open()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.open()' instead.",
            line: 1,
            column: 27,
            endLine: 1,
            endColumn: 36,
          },
        ],
      },
      {
        name: 'Upstream invalid 13',
        code: "const fs = require('fs'); fs.rename()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.rename()' instead.",
            line: 1,
            column: 27,
            endLine: 1,
            endColumn: 38,
          },
        ],
      },
      {
        name: 'Upstream invalid 14',
        code: "const fs = require('fs'); fs.truncate()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.truncate()' instead.",
            line: 1,
            column: 27,
            endLine: 1,
            endColumn: 40,
          },
        ],
      },
      {
        name: 'Upstream invalid 15',
        code: "const fs = require('fs'); fs.rmdir()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.rmdir()' instead.",
            line: 1,
            column: 27,
            endLine: 1,
            endColumn: 37,
          },
        ],
      },
      {
        name: 'Upstream invalid 16',
        code: "const fs = require('fs'); fs.mkdir()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.mkdir()' instead.",
            line: 1,
            column: 27,
            endLine: 1,
            endColumn: 37,
          },
        ],
      },
      {
        name: 'Upstream invalid 17',
        code: "const fs = require('fs'); fs.readdir()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.readdir()' instead.",
            line: 1,
            column: 27,
            endLine: 1,
            endColumn: 39,
          },
        ],
      },
      {
        name: 'Upstream invalid 18',
        code: "const fs = require('fs');fs.readlink()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.readlink()' instead.",
            line: 1,
            column: 26,
            endLine: 1,
            endColumn: 39,
          },
        ],
      },
      {
        name: 'Upstream invalid 19',
        code: "const fs = require('fs'); fs.symlink()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.symlink()' instead.",
            line: 1,
            column: 27,
            endLine: 1,
            endColumn: 39,
          },
        ],
      },
      {
        name: 'Upstream invalid 20',
        code: "const fs = require('fs'); fs.lstat()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.lstat()' instead.",
            line: 1,
            column: 27,
            endLine: 1,
            endColumn: 37,
          },
        ],
      },
      {
        name: 'Upstream invalid 21',
        code: "const fs = require('fs'); fs.stat()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.stat()' instead.",
            line: 1,
            column: 27,
            endLine: 1,
            endColumn: 36,
          },
        ],
      },
      {
        name: 'Upstream invalid 22',
        code: "const fs = require('fs'); fs.link()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.link()' instead.",
            line: 1,
            column: 27,
            endLine: 1,
            endColumn: 36,
          },
        ],
      },
      {
        name: 'Upstream invalid 23',
        code: "const fs = require('fs'); fs.unlink()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.unlink()' instead.",
            line: 1,
            column: 27,
            endLine: 1,
            endColumn: 38,
          },
        ],
      },
      {
        name: 'Upstream invalid 24',
        code: "const fs = require('fs'); fs.chmod()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.chmod()' instead.",
            line: 1,
            column: 27,
            endLine: 1,
            endColumn: 37,
          },
        ],
      },
      {
        name: 'Upstream invalid 25',
        code: "const fs = require('fs'); fs.lchmod()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.lchmod()' instead.",
            line: 1,
            column: 27,
            endLine: 1,
            endColumn: 38,
          },
        ],
      },
      {
        name: 'Upstream invalid 26',
        code: "const fs = require('fs'); fs.lchown()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.lchown()' instead.",
            line: 1,
            column: 27,
            endLine: 1,
            endColumn: 38,
          },
        ],
      },
      {
        name: 'Upstream invalid 27',
        code: "const fs = require('fs'); fs.chown()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.chown()' instead.",
            line: 1,
            column: 27,
            endLine: 1,
            endColumn: 37,
          },
        ],
      },
      {
        name: 'Upstream invalid 28',
        code: "const fs = require('fs'); fs.utimes()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.utimes()' instead.",
            line: 1,
            column: 27,
            endLine: 1,
            endColumn: 38,
          },
        ],
      },
      {
        name: 'Upstream invalid 29',
        code: "const fs = require('fs'); fs.realpath()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.realpath()' instead.",
            line: 1,
            column: 27,
            endLine: 1,
            endColumn: 40,
          },
        ],
      },
      {
        name: 'Upstream invalid 30',
        code: "const fs = require('fs'); fs.mkdtemp()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.mkdtemp()' instead.",
            line: 1,
            column: 27,
            endLine: 1,
            endColumn: 39,
          },
        ],
      },
      {
        name: 'Upstream invalid 31',
        code: "const fs = require('fs'); fs.writeFile()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.writeFile()' instead.",
            line: 1,
            column: 27,
            endLine: 1,
            endColumn: 41,
          },
        ],
      },
      {
        name: 'Upstream invalid 32',
        code: "const fs = require('fs'); fs.appendFile()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.appendFile()' instead.",
            line: 1,
            column: 27,
            endLine: 1,
            endColumn: 42,
          },
        ],
      },
      {
        name: 'Upstream invalid 33',
        code: "const fs = require('fs'); fs.readFile()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.readFile()' instead.",
            line: 1,
            column: 27,
            endLine: 1,
            endColumn: 40,
          },
        ],
      },
      {
        name: 'Upstream invalid 34',
        code: "const fs = require('fs'); fs.cp()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.cp()' instead.",
            line: 1,
            column: 27,
            endLine: 1,
            endColumn: 34,
          },
        ],
      },
      {
        name: 'Upstream invalid 35',
        code: "const fs = require('fs'); fs.glob()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.glob()' instead.",
            line: 1,
            column: 27,
            endLine: 1,
            endColumn: 36,
          },
        ],
      },
      {
        name: 'Upstream invalid 36',
        code: "const fs = require('fs'); fs.lutimes()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.lutimes()' instead.",
            line: 1,
            column: 27,
            endLine: 1,
            endColumn: 39,
          },
        ],
      },
      {
        name: 'Upstream invalid 37',
        code: "const fs = require('fs'); fs.opendir()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.opendir()' instead.",
            line: 1,
            column: 27,
            endLine: 1,
            endColumn: 39,
          },
        ],
      },
      {
        name: 'Upstream invalid 38',
        code: "const fs = require('fs'); fs.rm()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.rm()' instead.",
            line: 1,
            column: 27,
            endLine: 1,
            endColumn: 34,
          },
        ],
      },
      {
        name: 'Upstream invalid 39',
        code: "const fs = require('fs'); fs.statfs()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.statfs()' instead.",
            line: 1,
            column: 27,
            endLine: 1,
            endColumn: 38,
          },
        ],
      },
      {
        name: 'Documentation example 1',
        code: 'const fs = require("fs")\n\nfunction readData(filePath) {\n    fs.readFile(filePath, "utf8", (error, content) => {\n        //...\n    })\n}',
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.readFile()' instead.",
            line: 4,
            column: 5,
            endLine: 6,
            endColumn: 7,
          },
        ],
      },
      {
        name: 'Documentation example 2',
        code: 'import fs from "fs"\n\nfunction readData(filePath) {\n    fs.readFile(filePath, "utf8", (error, content) => {\n        //...\n    })\n}',
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'fs.promises.readFile()' instead.",
            line: 4,
            column: 5,
            endLine: 6,
            endColumn: 7,
          },
        ],
      },
    ],
  },
);
