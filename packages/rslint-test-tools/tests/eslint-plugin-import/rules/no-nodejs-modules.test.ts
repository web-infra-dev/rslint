// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
new RuleTester(fixtureConfig()).run('no-nodejs-modules', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: 'import _ from "lodash"',
    },
    {
      name: 'run 0 valid 1',
      code: 'import find from "lodash.find"',
    },
    {
      name: 'run 0 valid 2',
      code: 'import foo from "./foo"',
    },
    {
      name: 'run 0 valid 3',
      code: 'import foo from "../foo"',
    },
    {
      name: 'run 0 valid 4',
      code: 'import foo from "foo"',
    },
    {
      name: 'run 0 valid 5',
      code: 'import foo from "./"',
    },
    {
      name: 'run 0 valid 6',
      code: 'import foo from "@scope/foo"',
    },
    {
      name: 'run 0 valid 7',
      code: 'var _ = require("lodash")',
    },
    {
      name: 'run 0 valid 8',
      code: 'var find = require("lodash.find")',
    },
    {
      name: 'run 0 valid 9',
      code: 'var foo = require("./foo")',
    },
    {
      name: 'run 0 valid 10',
      code: 'var foo = require("../foo")',
    },
    {
      name: 'run 0 valid 11',
      code: 'var foo = require("foo")',
    },
    {
      name: 'run 0 valid 12',
      code: 'var foo = require("./")',
    },
    {
      name: 'run 0 valid 13',
      code: 'var foo = require("@scope/foo")',
    },
    {
      name: 'run 0 valid 14',
      code: 'import events from "events"',
      options: [
        {
          allow: ['events'],
        },
      ],
    },
    {
      name: 'run 0 valid 15',
      code: 'import path from "path"',
      options: [
        {
          allow: ['path'],
        },
      ],
    },
    {
      name: 'run 0 valid 16',
      code: 'var events = require("events")',
      options: [
        {
          allow: ['events'],
        },
      ],
    },
    {
      name: 'run 0 valid 17',
      code: 'var path = require("path")',
      options: [
        {
          allow: ['path'],
        },
      ],
    },
    {
      name: 'run 0 valid 18',
      code: 'import path from "path";import events from "events"',
      options: [
        {
          allow: ['path', 'events'],
        },
      ],
    },
    {
      name: 'run 0 valid 19',
      code: 'import events from "node:events"',
      options: [
        {
          allow: ['node:events'],
        },
      ],
    },
    {
      name: 'run 0 valid 20',
      code: 'var events = require("node:events")',
      options: [
        {
          allow: ['node:events'],
        },
      ],
    },
    {
      name: 'run 0 valid 21',
      code: 'import path from "node:path"',
      options: [
        {
          allow: ['node:path'],
        },
      ],
    },
    {
      name: 'run 0 valid 22',
      code: 'var path = require("node:path")',
      options: [
        {
          allow: ['node:path'],
        },
      ],
    },
    {
      name: 'run 0 valid 23',
      code: 'import path from "node:path";import events from "node:events"',
      options: [
        {
          allow: ['node:path', 'node:events'],
        },
      ],
    },
    {
      name: 'run 0 valid 24',
      code: "import _ from 'lodash';",
    },
    {
      name: 'run 0 valid 25',
      code: "import foo from 'foo';",
    },
    {
      name: 'run 0 valid 26',
      code: "import foo from './foo';",
    },
    {
      name: 'run 0 valid 27',
      code: "var _ = require('lodash');",
    },
    {
      name: 'run 0 valid 28',
      code: "var foo = require('foo');",
    },
    {
      name: 'run 0 valid 29',
      code: "var foo = require('./foo');",
    },
    {
      name: 'run 0 valid 30',
      code: "import path from 'path';",
      options: [
        {
          allow: ['path'],
        },
      ],
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: 'import path from "path"',
      errors: [
        {
          messageId: '',
          message: 'Do not import Node.js builtin module "path"',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      name: 'run 0 invalid 1',
      code: 'import fs from "fs"',
      errors: [
        {
          messageId: '',
          message: 'Do not import Node.js builtin module "fs"',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      name: 'run 0 invalid 2',
      code: 'var path = require("path")',
      errors: [
        {
          messageId: '',
          message: 'Do not import Node.js builtin module "path"',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      name: 'run 0 invalid 3',
      code: 'var fs = require("fs")',
      errors: [
        {
          messageId: '',
          message: 'Do not import Node.js builtin module "fs"',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      name: 'run 0 invalid 4',
      code: 'import fs from "fs"',
      options: [
        {
          allow: ['path'],
        },
      ],
      errors: [
        {
          messageId: '',
          message: 'Do not import Node.js builtin module "fs"',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      name: 'run 0 invalid 5',
      code: 'import path from "node:path"',
      errors: [
        {
          messageId: '',
          message: 'Do not import Node.js builtin module "node:path"',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      name: 'run 0 invalid 6',
      code: 'var path = require("node:path")',
      errors: [
        {
          messageId: '',
          message: 'Do not import Node.js builtin module "node:path"',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      name: 'run 0 invalid 7',
      code: 'import fs from "node:fs"',
      errors: [
        {
          messageId: '',
          message: 'Do not import Node.js builtin module "node:fs"',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      name: 'run 0 invalid 8',
      code: 'var fs = require("node:fs")',
      errors: [
        {
          messageId: '',
          message: 'Do not import Node.js builtin module "node:fs"',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      name: 'run 0 invalid 9',
      code: 'import fs from "node:fs"',
      options: [
        {
          allow: ['node:path'],
        },
      ],
      errors: [
        {
          messageId: '',
          message: 'Do not import Node.js builtin module "node:fs"',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      name: 'run 0 invalid 10',
      code: "import fs from 'fs';",
      errors: [
        {
          messageId: '',
          message: 'Do not import Node.js builtin module "fs"',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      name: 'run 0 invalid 11',
      code: "import path from 'path';",
      errors: [
        {
          messageId: '',
          message: 'Do not import Node.js builtin module "path"',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      name: 'run 0 invalid 12',
      code: "var fs = require('fs');",
      errors: [
        {
          messageId: '',
          message: 'Do not import Node.js builtin module "fs"',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      name: 'run 0 invalid 13',
      code: "var path = require('path');",
      errors: [
        {
          messageId: '',
          message: 'Do not import Node.js builtin module "path"',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
  ],
});
