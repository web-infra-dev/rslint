// Upstream: https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/no-nodejs-modules.js
// Documentation: https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-nodejs-modules.md
import { RuleTester } from '../rule-tester.js';

const ruleTester = new RuleTester();
ruleTester.run('no-nodejs-modules', null as never, {
  valid: [
    // Upstream (including all node: cases)
    {
      code: 'import _ from "lodash"',
    },
    {
      code: 'import find from "lodash.find"',
    },
    {
      code: 'import foo from "./foo"',
    },
    {
      code: 'import foo from "../foo"',
    },
    {
      code: 'import foo from "foo"',
    },
    {
      code: 'import foo from "./"',
    },
    {
      code: 'import foo from "@scope/foo"',
    },
    {
      code: 'var _ = require("lodash")',
    },
    {
      code: 'var find = require("lodash.find")',
    },
    {
      code: 'var foo = require("./foo")',
    },
    {
      code: 'var foo = require("../foo")',
    },
    {
      code: 'var foo = require("foo")',
    },
    {
      code: 'var foo = require("./")',
    },
    {
      code: 'var foo = require("@scope/foo")',
    },
    {
      code: 'import events from "events"',
      options: [
        {
          allow: ['events'],
        },
      ],
    },
    {
      code: 'import path from "path"',
      options: [
        {
          allow: ['path'],
        },
      ],
    },
    {
      code: 'var events = require("events")',
      options: [
        {
          allow: ['events'],
        },
      ],
    },
    {
      code: 'var path = require("path")',
      options: [
        {
          allow: ['path'],
        },
      ],
    },
    {
      code: 'import path from "path";import events from "events"',
      options: [
        {
          allow: ['path', 'events'],
        },
      ],
    },
    {
      code: 'import events from "node:events"',
      options: [
        {
          allow: ['node:events'],
        },
      ],
    },
    {
      code: 'var events = require("node:events")',
      options: [
        {
          allow: ['node:events'],
        },
      ],
    },
    {
      code: 'import path from "node:path"',
      options: [
        {
          allow: ['node:path'],
        },
      ],
    },
    {
      code: 'var path = require("node:path")',
      options: [
        {
          allow: ['node:path'],
        },
      ],
    },
    {
      code: 'import path from "node:path";import events from "node:events"',
      options: [
        {
          allow: ['node:path', 'node:events'],
        },
      ],
    },
    // Upstream documentation examples
    {
      code: "import _ from 'lodash';",
    },
    {
      code: "import foo from 'foo';",
    },
    {
      code: "import foo from './foo';",
    },
    {
      code: "var _ = require('lodash');",
    },
    {
      code: "var foo = require('foo');",
    },
    {
      code: "var foo = require('./foo');",
    },
    {
      code: "import path from 'path';",
      options: [
        {
          allow: ['path'],
        },
      ],
    },
  ],
  invalid: [
    // Upstream (including all node: cases)
    {
      code: 'import path from "path"',
      errors: ['Do not import Node.js builtin module "path"'],
    },
    {
      code: 'import fs from "fs"',
      errors: ['Do not import Node.js builtin module "fs"'],
    },
    {
      code: 'var path = require("path")',
      errors: ['Do not import Node.js builtin module "path"'],
    },
    {
      code: 'var fs = require("fs")',
      errors: ['Do not import Node.js builtin module "fs"'],
    },
    {
      code: 'import fs from "fs"',
      options: [
        {
          allow: ['path'],
        },
      ],
      errors: ['Do not import Node.js builtin module "fs"'],
    },
    {
      code: 'import path from "node:path"',
      errors: ['Do not import Node.js builtin module "node:path"'],
    },
    {
      code: 'var path = require("node:path")',
      errors: ['Do not import Node.js builtin module "node:path"'],
    },
    {
      code: 'import fs from "node:fs"',
      errors: ['Do not import Node.js builtin module "node:fs"'],
    },
    {
      code: 'var fs = require("node:fs")',
      errors: ['Do not import Node.js builtin module "node:fs"'],
    },
    {
      code: 'import fs from "node:fs"',
      options: [
        {
          allow: ['node:path'],
        },
      ],
      errors: ['Do not import Node.js builtin module "node:fs"'],
    },
    // Upstream documentation examples
    {
      code: "import fs from 'fs';",
      errors: ['Do not import Node.js builtin module "fs"'],
    },
    {
      code: "import path from 'path';",
      errors: ['Do not import Node.js builtin module "path"'],
    },
    {
      code: "var fs = require('fs');",
      errors: ['Do not import Node.js builtin module "fs"'],
    },
    {
      code: "var path = require('path');",
      errors: ['Do not import Node.js builtin module "path"'],
    },
  ],
});
