// Upstream tests and documentation: eslint-plugin-unicorn v79.0.0.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-module.js
import assert from 'node:assert/strict';
import path from 'node:path';
import { lint } from '@rslint/core/internal';
import type { RslintConfigEntry } from '@rslint/core';

interface Case {
  code: string;
  filename: string;
  languageOptions: RslintConfigEntry['languageOptions'];
  errors: {
    messageId: string;
    message: string;
    line: number;
    column: number;
    endLine: number;
    endColumn: number;
    suggestions: { messageId: string; desc: string; output: string }[];
  }[];
  output: string | null;
}

const globals = {
  exports: 'readonly',
  require: 'readonly',
  module: 'readonly',
  __filename: 'readonly',
  __dirname: 'readonly',
} as const;
const groups: { name: string; valid: Case[]; invalid: Case[] }[] = [
  {
    name: 'Strict mode',
    valid: [
      {
        code: 'const foo = "use strict";',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [],
        output: null,
      },
      {
        code: '("use strict")',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [],
        output: null,
      },
      {
        code: '"use strong";',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [],
        output: null,
      },
      {
        code: 'eval("\'use strict\'; var x = 42; x;");',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [],
        output: null,
      },
      {
        code: 'new Function("\'use strict\'; var x = 42; return x;");',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [],
        output: null,
      },
    ],
    invalid: [
      {
        code: "'use strict';\n\nconsole.log(1);",
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/use-strict-directive',
            message: 'Do not use "use strict" directive.',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 14,
            suggestions: [
              {
                messageId: 'suggestion/use-strict-directive',
                desc: 'Remove "use strict" directive.',
                output: 'console.log(1);',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: '"use strict";\n\nconsole.log(1);',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/use-strict-directive',
            message: 'Do not use "use strict" directive.',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 14,
            suggestions: [
              {
                messageId: 'suggestion/use-strict-directive',
                desc: 'Remove "use strict" directive.',
                output: 'console.log(1);',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'function foo () {\n\t"use strict";\n\tconsole.log(1);\n}',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/use-strict-directive',
            message: 'Do not use "use strict" directive.',
            line: 2,
            column: 2,
            endLine: 2,
            endColumn: 15,
            suggestions: [
              {
                messageId: 'suggestion/use-strict-directive',
                desc: 'Remove "use strict" directive.',
                output: 'function foo () {\n\tconsole.log(1);\n}',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: "'use strict';\n\nconsole.log(1);",
        filename: 'example.mjs',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/use-strict-directive',
            message: 'Do not use "use strict" directive.',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 14,
            suggestions: [],
          },
        ],
        output: 'console.log(1);',
      },
    ],
  },
  {
    name: 'Global return',
    valid: [
      {
        code: 'function a() {\n\treturn;\n}',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [],
        output: null,
      },
    ],
    invalid: [
      {
        code: 'if (foo) {\n\treturn;\n}',
        filename: 'upstream.js',
        languageOptions: {
          sourceType: 'script',
          parserOptions: {
            ecmaFeatures: {
              globalReturn: true,
            },
          },
        },
        errors: [
          {
            messageId: 'error/global-return',
            message: '"return" should be used inside a function.',
            line: 2,
            column: 2,
            endLine: 2,
            endColumn: 8,
            suggestions: [],
          },
        ],
        output: null,
      },
    ],
  },
  {
    name: '__dirname and __filename',
    valid: [
      {
        code: 'const __filename = 1;\nconst foo = __filename;',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [],
        output: null,
      },
      {
        code: 'const __dirname = 1;\nconst foo = __dirname;',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [],
        output: null,
      },
      {
        code: 'import {__filename as filename} from "foo.mjs"',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [],
        output: null,
      },
      {
        code: 'const foo = 1;export {foo as __dirname}',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [],
        output: null,
      },
    ],
    invalid: [
      {
        code: 'const dirname = __dirname;',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "__dirname".',
            line: 1,
            column: 17,
            endLine: 1,
            endColumn: 26,
            suggestions: [
              {
                messageId: 'suggestion/import-meta-dirname',
                desc: 'Replace `__dirname` with `import.meta.dirname`.',
                output: 'const dirname = import.meta.dirname;',
              },
              {
                messageId: 'suggestion/import-meta-url-to-dirname',
                desc: 'Replace `__dirname` with `…(import.meta.url)`.',
                output:
                  'const dirname = path.dirname(url.fileURLToPath(import.meta.url));',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const dirname = __filename;',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "__filename".',
            line: 1,
            column: 17,
            endLine: 1,
            endColumn: 27,
            suggestions: [
              {
                messageId: 'suggestion/import-meta-filename',
                desc: 'Replace `__filename` with `import.meta.filename`.',
                output: 'const dirname = import.meta.filename;',
              },
              {
                messageId: 'suggestion/import-meta-url-to-filename',
                desc: 'Replace `__filename` with `…(import.meta.url)`.',
                output: 'const dirname = url.fileURLToPath(import.meta.url);',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const foo = { __dirname};',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "__dirname".',
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 24,
            suggestions: [
              {
                messageId: 'suggestion/import-meta-dirname',
                desc: 'Replace `__dirname` with `import.meta.dirname`.',
                output: 'const foo = { __dirname: import.meta.dirname};',
              },
              {
                messageId: 'suggestion/import-meta-url-to-dirname',
                desc: 'Replace `__dirname` with `…(import.meta.url)`.',
                output:
                  'const foo = { __dirname: path.dirname(url.fileURLToPath(import.meta.url))};',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const foo = {__filename, };',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "__filename".',
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 24,
            suggestions: [
              {
                messageId: 'suggestion/import-meta-filename',
                desc: 'Replace `__filename` with `import.meta.filename`.',
                output: 'const foo = {__filename: import.meta.filename, };',
              },
              {
                messageId: 'suggestion/import-meta-url-to-filename',
                desc: 'Replace `__filename` with `…(import.meta.url)`.',
                output:
                  'const foo = {__filename: url.fileURLToPath(import.meta.url), };',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'if (__dirname.startsWith("/project/src/")) {}',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "__dirname".',
            line: 1,
            column: 5,
            endLine: 1,
            endColumn: 14,
            suggestions: [
              {
                messageId: 'suggestion/import-meta-dirname',
                desc: 'Replace `__dirname` with `import.meta.dirname`.',
                output:
                  'if (import.meta.dirname.startsWith("/project/src/")) {}',
              },
              {
                messageId: 'suggestion/import-meta-url-to-dirname',
                desc: 'Replace `__dirname` with `…(import.meta.url)`.',
                output:
                  'if (path.dirname(url.fileURLToPath(import.meta.url)).startsWith("/project/src/")) {}',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'if (__filename.endsWith(".js")) {}',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "__filename".',
            line: 1,
            column: 5,
            endLine: 1,
            endColumn: 15,
            suggestions: [
              {
                messageId: 'suggestion/import-meta-filename',
                desc: 'Replace `__filename` with `import.meta.filename`.',
                output: 'if (import.meta.filename.endsWith(".js")) {}',
              },
              {
                messageId: 'suggestion/import-meta-url-to-filename',
                desc: 'Replace `__filename` with `…(import.meta.url)`.',
                output:
                  'if (url.fileURLToPath(import.meta.url).endsWith(".js")) {}',
              },
            ],
          },
        ],
        output: null,
      },
    ],
  },
  {
    name: 'require',
    valid: [
      {
        code: 'import {createRequire} from \'module\';\nconst require = createRequire(import.meta.url);\nconst foo = require("foo");',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [],
        output: null,
      },
      {
        code: 'const require = function require(require, id) {\n\treturn require(id);\n}',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [],
        output: null,
      },
      {
        code: 'const require = class A {\n\trequire(require) {\n\t\treturn require(id);\n\t}\n}',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [],
        output: null,
      },
    ],
    invalid: [
      {
        code: 'require("foo");',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 8,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import "foo";',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: "require('foo');",
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 8,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: "import 'foo';",
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'require( (("foo")) );',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 8,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import  "foo" ;',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: '((require))("foo");',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 3,
            endLine: 1,
            endColumn: 10,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import "foo";',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: '(( require("foo") ));',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 4,
            endLine: 1,
            endColumn: 11,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: ' import "foo" ;',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const foo=require("foo");',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 11,
            endLine: 1,
            endColumn: 18,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import foo from "foo";',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const foo = require.resolve("foo");',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 20,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'const foo\n\t=\n\trequire("foo");',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 3,
            column: 2,
            endLine: 3,
            endColumn: 9,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import foo from "foo";',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const foo = require("foo");',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 20,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import foo from "foo";',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const foo = require( (("foo")) );',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 20,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import foo from  "foo" ;',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const foo = ((require))("foo");',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 22,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import foo from "foo";',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const foo = (( require("foo") ));',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 16,
            endLine: 1,
            endColumn: 23,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import foo from  "foo" ;',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const {foo}=require("foo");',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 20,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import {foo} from "foo";',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const {foo}\n\t=\n\trequire("foo");',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 3,
            column: 2,
            endLine: 3,
            endColumn: 9,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import {foo} from "foo";',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const {foo} = require("foo");',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 22,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import {foo} from "foo";',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const {foo} = require( (("foo")) );',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 22,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import {foo} from  "foo" ;',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const {foo} = ((require))("foo");',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 17,
            endLine: 1,
            endColumn: 24,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import {foo} from "foo";',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const {foo} = (( require("foo") ));',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 18,
            endLine: 1,
            endColumn: 25,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import {foo} from  "foo" ;',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const {foo: foo}=require("foo");',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 18,
            endLine: 1,
            endColumn: 25,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import {foo as foo} from "foo";',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const {foo: foo}\n\t=\n\trequire("foo");',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 3,
            column: 2,
            endLine: 3,
            endColumn: 9,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import {foo as foo} from "foo";',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const {foo: foo} = require("foo");',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 20,
            endLine: 1,
            endColumn: 27,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import {foo as foo} from "foo";',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const {foo: foo} = require( (("foo")) );',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 20,
            endLine: 1,
            endColumn: 27,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import {foo as foo} from  "foo" ;',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const {foo: foo} = ((require))("foo");',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 22,
            endLine: 1,
            endColumn: 29,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import {foo as foo} from "foo";',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const {foo: foo} = (( require("foo") ));',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 23,
            endLine: 1,
            endColumn: 30,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import {foo as foo} from  "foo" ;',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const {foo:bar}=require("foo");',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 17,
            endLine: 1,
            endColumn: 24,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import {foo as bar} from "foo";',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const {foo:bar}\n\t=\n\trequire("foo");',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 3,
            column: 2,
            endLine: 3,
            endColumn: 9,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import {foo as bar} from "foo";',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const {foo:bar} = require("foo");',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 19,
            endLine: 1,
            endColumn: 26,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import {foo as bar} from "foo";',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const {foo:bar} = require( (("foo")) );',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 19,
            endLine: 1,
            endColumn: 26,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import {foo as bar} from  "foo" ;',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const {foo:bar} = ((require))("foo");',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 21,
            endLine: 1,
            endColumn: 28,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import {foo as bar} from "foo";',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const {foo:bar} = (( require("foo") ));',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 22,
            endLine: 1,
            endColumn: 29,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import {foo as bar} from  "foo" ;',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const {a   :foo, b:   bar, default   :   baz}=require("foo");',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 47,
            endLine: 1,
            endColumn: 54,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output:
                  'import {a as foo, b as bar, default as baz} from "foo";',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const {\n\ta   :foo,\n\tb:   bar,\n\tdefault   :   baz,\n}\n\t=\n\trequire("foo");',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 7,
            column: 2,
            endLine: 7,
            endColumn: 9,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output:
                  'import {\n\ta as foo,\n\tb as bar,\n\tdefault as baz,\n} from "foo";',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const {a   :foo, b:   bar, default   :   baz} = require("foo");',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 49,
            endLine: 1,
            endColumn: 56,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output:
                  'import {a as foo, b as bar, default as baz} from "foo";',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const {a   :foo, b:   bar, default   :   baz} = require( (("foo")) );',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 49,
            endLine: 1,
            endColumn: 56,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output:
                  'import {a as foo, b as bar, default as baz} from  "foo" ;',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const {a   :foo, b:   bar, default   :   baz} = ((require))("foo");',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 51,
            endLine: 1,
            endColumn: 58,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output:
                  'import {a as foo, b as bar, default as baz} from "foo";',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const {a   :foo, b:   bar, default   :   baz} = (( require("foo") ));',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 52,
            endLine: 1,
            endColumn: 59,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output:
                  'import {a as foo, b as bar, default as baz} from  "foo" ;',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const {} = require("foo");',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 12,
            endLine: 1,
            endColumn: 19,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import {} from "foo";',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const{   }=require("foo");',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 12,
            endLine: 1,
            endColumn: 19,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import{   } from "foo";',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const r = require;\nconst foo = r("foo");',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 11,
            endLine: 1,
            endColumn: 18,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'new require("foo")',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 5,
            endLine: 1,
            endColumn: 12,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'require("foo", extraArgument)',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 8,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'const a = require()',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 11,
            endLine: 1,
            endColumn: 18,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'require(..."foo")',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 8,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'require("../" + "file.js")',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 8,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'require(file)',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 8,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'a = require("foo")',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 5,
            endLine: 1,
            endColumn: 12,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'function a(a = require("foo")) {}',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 16,
            endLine: 1,
            endColumn: 23,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'let foo = require("foo");',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 11,
            endLine: 1,
            endColumn: 18,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'const foo = require("foo"), bar = require("bar");',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 20,
            suggestions: [],
          },
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 35,
            endLine: 1,
            endColumn: 42,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'const {[foo]: foo} = require("foo");',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 22,
            endLine: 1,
            endColumn: 29,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'const {["foo"]: foo} = require("foo");',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 24,
            endLine: 1,
            endColumn: 31,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'if (foo) require("foo");',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 10,
            endLine: 1,
            endColumn: 17,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'const foo = require`foo`;',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 20,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'function loadModule() {\n\treturn require("foo");\n}',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 2,
            column: 9,
            endLine: 2,
            endColumn: 16,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'function loadModule() {\n\tconst foo = require("foo");\n\treturn foo;\n}',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 2,
            column: 14,
            endLine: 2,
            endColumn: 21,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'const foo = require("foo"), bar = 1;',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 20,
            suggestions: [],
          },
        ],
        output: null,
      },
    ],
  },
  {
    name: 'exports and module',
    valid: [
      {
        code: 'function exports(exports) {}',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [],
        output: null,
      },
      {
        code: 'function module(module) {}',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [],
        output: null,
      },
      {
        code: 'const exports = foo;\nexports.bar = bar;',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [],
        output: null,
      },
      {
        code: 'const exports = 1;',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [],
        output: null,
      },
      {
        code: 'const module = foo;\nmodule.exports = bar;\nmodule.exports.bar = bar;',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [],
        output: null,
      },
      {
        code: 'const module = 1;',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [],
        output: null,
      },
    ],
    invalid: [
      {
        code: 'exports = foo;',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "exports".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 8,
            suggestions: [
              {
                messageId: 'suggestion/export',
                desc: 'Switch to `export`.',
                output: 'export default foo;',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'module.exports = foo;',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "module".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 7,
            suggestions: [
              {
                messageId: 'suggestion/export',
                desc: 'Switch to `export`.',
                output: 'export default foo;',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: '(( ((exports)) = ((foo)) ));',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "exports".',
            line: 1,
            column: 6,
            endLine: 1,
            endColumn: 13,
            suggestions: [
              {
                messageId: 'suggestion/export',
                desc: 'Switch to `export`.',
                output: ' export default  ((foo)) ;',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: '(( ((module.exports)) = ((foo)) ));',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "module".',
            line: 1,
            column: 6,
            endLine: 1,
            endColumn: 12,
            suggestions: [
              {
                messageId: 'suggestion/export',
                desc: 'Switch to `export`.',
                output: ' export default  ((foo)) ;',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const foo = 1;exports.foo = foo;',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "exports".',
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 22,
            suggestions: [
              {
                messageId: 'suggestion/export',
                desc: 'Switch to `export`.',
                output: 'const foo = 1;export {foo as foo};',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const foo = 1;module.exports.foo = foo;',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "module".',
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 21,
            suggestions: [
              {
                messageId: 'suggestion/export',
                desc: 'Switch to `export`.',
                output: 'const foo = 1;export {foo as foo};',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'exports["foo"] = foo;',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "exports".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 8,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'module.exports["foo"] = foo;',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "module".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 7,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'const foo = exports;',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "exports".',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 20,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'const foo = exports.foo;',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "exports".',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 20,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'const foo = module.exports;',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "module".',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 19,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'const foo = module.exports.foo;',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "module".',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 19,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'module["exports"] = foo;',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "module".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 7,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'module[exports] = foo;',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "module".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 7,
            suggestions: [],
          },
          {
            messageId: 'error/identifier',
            message: 'Do not use "exports".',
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 15,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'module.exports.foo.bar = foo;',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "module".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 7,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'const foo = 1;exports.default = foo;',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "exports".',
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 22,
            suggestions: [
              {
                messageId: 'suggestion/export',
                desc: 'Switch to `export`.',
                output: 'const foo = 1;export {foo as default};',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const foo = 1;module.exports.default = foo;',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "module".',
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 21,
            suggestions: [
              {
                messageId: 'suggestion/export',
                desc: 'Switch to `export`.',
                output: 'const foo = 1;export {foo as default};',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'exports.foo.bar = foo;',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "exports".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 8,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'exports = 1;',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "exports".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 8,
            suggestions: [
              {
                messageId: 'suggestion/export',
                desc: 'Switch to `export`.',
                output: 'export default 1;',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'exports.foo = [];',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "exports".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 8,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'module.exports = function() {};',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "module".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 7,
            suggestions: [
              {
                messageId: 'suggestion/export',
                desc: 'Switch to `export`.',
                output: 'export default function() {};',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'module.exports.foo = foo || bar;',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "module".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 7,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'exports += foo;',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "exports".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 8,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'const foo = module.children',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "module".',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 19,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'const parentModule = module.parent',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "module".',
            line: 1,
            column: 22,
            endLine: 1,
            endColumn: 28,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'function foo() {\n\texports.foo = foo;\n\tmodule.exports.foo = foo;\n}',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "exports".',
            line: 2,
            column: 2,
            endLine: 2,
            endColumn: 9,
            suggestions: [],
          },
          {
            messageId: 'error/identifier',
            message: 'Do not use "module".',
            line: 3,
            column: 2,
            endLine: 3,
            endColumn: 8,
            suggestions: [],
          },
        ],
        output: null,
      },
    ],
  },
  {
    name: 'TypeScript',
    valid: [
      {
        code: 'type module = number[];',
        filename: 'upstream.ts',
        languageOptions: {},
        errors: [],
        output: null,
      },
      {
        code: 'type ModuleRegistry = { [module: string]: string };',
        filename: 'upstream.ts',
        languageOptions: {},
        errors: [],
        output: null,
      },
      {
        code: 'const module = 1; type ModuleRegistry = { [module: string]: string };',
        filename: 'upstream.ts',
        languageOptions: {},
        errors: [],
        output: null,
      },
      {
        code: 'type module = number[]; type ModuleRegistry = { [module: string]: string };',
        filename: 'upstream.ts',
        languageOptions: {},
        errors: [],
        output: null,
      },
      {
        code: 'type Data = { [module in keyof string]: number; };',
        filename: 'upstream.ts',
        languageOptions: {},
        errors: [],
        output: null,
      },
      {
        code: 'type ModuleRegistry = { [exports: string]: string };',
        filename: 'upstream.ts',
        languageOptions: {},
        errors: [],
        output: null,
      },
    ],
    invalid: [
      {
        code: 'const foo: Foo = require("foo");',
        filename: 'upstream.ts',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 18,
            endLine: 1,
            endColumn: 25,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'const {foo}: Foo = require("foo");',
        filename: 'upstream.ts',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 20,
            endLine: 1,
            endColumn: 27,
            suggestions: [],
          },
        ],
        output: null,
      },
    ],
  },
  {
    name: 'JSON modules',
    valid: [],
    invalid: [
      {
        code: 'const data = require("./a.json");',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 21,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import data from "./a.json" with {type: "json"};',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'require("./a.json");',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 8,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import "./a.json" with {type: "json"};',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: "const data = (require(('./a.json')));",
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 22,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: "import data from './a.json' with {type: 'json'};",
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const {foo} = require("./a.json");',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 22,
            suggestions: [],
          },
        ],
        output: null,
      },
    ],
  },
  {
    name: '.cjs files',
    valid: [
      {
        code: '__dirname',
        filename: 'foo.cjs',
        languageOptions: {},
        errors: [],
        output: null,
      },
      {
        code: '__dirname',
        filename: 'foo.cjS',
        languageOptions: {},
        errors: [],
        output: null,
      },
    ],
    invalid: [
      {
        code: '__filename',
        filename: 'foo.mjs',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "__filename".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 11,
            suggestions: [
              {
                messageId: 'suggestion/import-meta-filename',
                desc: 'Replace `__filename` with `import.meta.filename`.',
                output: 'import.meta.filename',
              },
              {
                messageId: 'suggestion/import-meta-url-to-filename',
                desc: 'Replace `__filename` with `…(import.meta.url)`.',
                output: 'url.fileURLToPath(import.meta.url)',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'require("lodash")',
        filename: 'foo.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 8,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import "lodash"',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'require("lodash")',
        filename: 'foo.cjs/foo.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 8,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import "lodash"',
              },
            ],
          },
        ],
        output: null,
      },
    ],
  },
  {
    name: 'Comments in named exports',
    valid: [],
    invalid: [
      {
        code: 'const foo = 1;\nexports /* note */ .foo = foo;',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "exports".',
            line: 2,
            column: 1,
            endLine: 2,
            endColumn: 8,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'const foo = 1;\nexports./* c */foo = foo;',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "exports".',
            line: 2,
            column: 1,
            endLine: 2,
            endColumn: 8,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'const foo = 1;\nexports.foo /* c */ = foo;',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "exports".',
            line: 2,
            column: 1,
            endLine: 2,
            endColumn: 8,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'const foo = 1;\nmodule.exports /* c */ .foo = foo;',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "module".',
            line: 2,
            column: 1,
            endLine: 2,
            endColumn: 7,
            suggestions: [],
          },
        ],
        output: null,
      },
    ],
  },
  {
    name: 'Constructor callees',
    valid: [],
    invalid: [
      {
        code: 'new __filename.constructor("text");',
        filename: 'upstream.js',
        languageOptions: {
          globals: {
            __filename: 'readonly',
          },
        },
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "__filename".',
            line: 1,
            column: 5,
            endLine: 1,
            endColumn: 15,
            suggestions: [
              {
                messageId: 'suggestion/import-meta-filename',
                desc: 'Replace `__filename` with `import.meta.filename`.',
                output: 'new import.meta.filename.constructor("text");',
              },
              {
                messageId: 'suggestion/import-meta-url-to-filename',
                desc: 'Replace `__filename` with `…(import.meta.url)`.',
                output:
                  'new (url.fileURLToPath(import.meta.url)).constructor("text");',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'new (__filename).constructor("text");',
        filename: 'upstream.js',
        languageOptions: {
          globals: {
            __filename: 'readonly',
          },
        },
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "__filename".',
            line: 1,
            column: 6,
            endLine: 1,
            endColumn: 16,
            suggestions: [
              {
                messageId: 'suggestion/import-meta-filename',
                desc: 'Replace `__filename` with `import.meta.filename`.',
                output: 'new (import.meta.filename).constructor("text");',
              },
              {
                messageId: 'suggestion/import-meta-url-to-filename',
                desc: 'Replace `__filename` with `…(import.meta.url)`.',
                output:
                  'new (url.fileURLToPath(import.meta.url)).constructor("text");',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'new __filename.constructor("text");',
        filename: 'upstream.ts',
        languageOptions: {
          globals: {
            __filename: 'readonly',
          },
        },
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "__filename".',
            line: 1,
            column: 5,
            endLine: 1,
            endColumn: 15,
            suggestions: [
              {
                messageId: 'suggestion/import-meta-filename',
                desc: 'Replace `__filename` with `import.meta.filename`.',
                output: 'new (import.meta.filename).constructor("text");',
              },
              {
                messageId: 'suggestion/import-meta-url-to-filename',
                desc: 'Replace `__filename` with `…(import.meta.url)`.',
                output:
                  'new (url.fileURLToPath(import.meta.url)).constructor("text");',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'new __filename!.constructor("text");',
        filename: 'upstream.ts',
        languageOptions: {
          globals: {
            __filename: 'readonly',
          },
        },
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "__filename".',
            line: 1,
            column: 5,
            endLine: 1,
            endColumn: 15,
            suggestions: [
              {
                messageId: 'suggestion/import-meta-filename',
                desc: 'Replace `__filename` with `import.meta.filename`.',
                output: 'new (import.meta.filename)!.constructor("text");',
              },
              {
                messageId: 'suggestion/import-meta-url-to-filename',
                desc: 'Replace `__filename` with `…(import.meta.url)`.',
                output:
                  'new (url.fileURLToPath(import.meta.url))!.constructor("text");',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'new __dirname.constructor("text");',
        filename: 'upstream.js',
        languageOptions: {
          globals: {
            __dirname: 'readonly',
          },
        },
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "__dirname".',
            line: 1,
            column: 5,
            endLine: 1,
            endColumn: 14,
            suggestions: [
              {
                messageId: 'suggestion/import-meta-dirname',
                desc: 'Replace `__dirname` with `import.meta.dirname`.',
                output: 'new import.meta.dirname.constructor("text");',
              },
              {
                messageId: 'suggestion/import-meta-url-to-dirname',
                desc: 'Replace `__dirname` with `…(import.meta.url)`.',
                output:
                  'new (path.dirname(url.fileURLToPath(import.meta.url))).constructor("text");',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'new (__dirname).constructor("text");',
        filename: 'upstream.js',
        languageOptions: {
          globals: {
            __dirname: 'readonly',
          },
        },
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "__dirname".',
            line: 1,
            column: 6,
            endLine: 1,
            endColumn: 15,
            suggestions: [
              {
                messageId: 'suggestion/import-meta-dirname',
                desc: 'Replace `__dirname` with `import.meta.dirname`.',
                output: 'new (import.meta.dirname).constructor("text");',
              },
              {
                messageId: 'suggestion/import-meta-url-to-dirname',
                desc: 'Replace `__dirname` with `…(import.meta.url)`.',
                output:
                  'new (path.dirname(url.fileURLToPath(import.meta.url))).constructor("text");',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'new __dirname.constructor("text");',
        filename: 'upstream.ts',
        languageOptions: {
          globals: {
            __dirname: 'readonly',
          },
        },
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "__dirname".',
            line: 1,
            column: 5,
            endLine: 1,
            endColumn: 14,
            suggestions: [
              {
                messageId: 'suggestion/import-meta-dirname',
                desc: 'Replace `__dirname` with `import.meta.dirname`.',
                output: 'new (import.meta.dirname).constructor("text");',
              },
              {
                messageId: 'suggestion/import-meta-url-to-dirname',
                desc: 'Replace `__dirname` with `…(import.meta.url)`.',
                output:
                  'new (path.dirname(url.fileURLToPath(import.meta.url))).constructor("text");',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'new __dirname!.constructor("text");',
        filename: 'upstream.ts',
        languageOptions: {
          globals: {
            __dirname: 'readonly',
          },
        },
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "__dirname".',
            line: 1,
            column: 5,
            endLine: 1,
            endColumn: 14,
            suggestions: [
              {
                messageId: 'suggestion/import-meta-dirname',
                desc: 'Replace `__dirname` with `import.meta.dirname`.',
                output: 'new (import.meta.dirname)!.constructor("text");',
              },
              {
                messageId: 'suggestion/import-meta-url-to-dirname',
                desc: 'Replace `__dirname` with `…(import.meta.url)`.',
                output:
                  'new (path.dirname(url.fileURLToPath(import.meta.url)))!.constructor("text");',
              },
            ],
          },
        ],
        output: null,
      },
    ],
  },
  {
    name: 'Static template module sources',
    valid: [],
    invalid: [
      {
        code: 'const file = require(`file`);',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 21,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: "import file from 'file';",
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'require(`file`);',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 8,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: "import 'file';",
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const file = require(`./file.json`);',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 21,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: "import file from './file.json' with {type: 'json'};",
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const {foo: bar} = require(`file`);',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 20,
            endLine: 1,
            endColumn: 27,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: "import {foo as bar} from 'file';",
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: "const file = require(`./it's.js`);",
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 21,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: "import file from './it\\'s.js';",
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const file = require(`./line\nfile.js`);',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 21,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: "import file from './line\\nfile.js';",
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const file = require(`./back\\`tick.js`);',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 21,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: "import file from './back`tick.js';",
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const file = require(`./fi\\u006ce`);',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 21,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: "import file from './file';",
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const file = require(/* keep */ (`file`));',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 21,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: "import file from /* keep */ 'file';",
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const file = require(`${name}`);',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 21,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'const file = require(tag`file`);',
        filename: 'upstream.js',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 21,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'const file: Foo = require(`file`);',
        filename: 'upstream.ts',
        languageOptions: {},
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 19,
            endLine: 1,
            endColumn: 26,
            suggestions: [],
          },
        ],
        output: null,
      },
    ],
  },
  {
    name: 'Documentation examples',
    valid: [
      {
        code: 'const __dirname = import.meta.dirname;\nconst __filename = import.meta.filename;',
        filename: 'upstream.js',
        languageOptions: {},
        output: null,
        errors: [],
      },
      {
        code: "import {fileURLToPath} from 'node:url';\nimport path from 'node:path';\nconst __filename = fileURLToPath(import.meta.url);\nconst __dirname = path.dirname(fileURLToPath(import.meta.url));",
        filename: 'upstream.js',
        languageOptions: {},
        output: null,
        errors: [],
      },
      {
        code: "import {fileURLToPath} from 'node:url';\nconst foo = fileURLToPath(new URL('foo.js', import.meta.url));",
        filename: 'upstream.js',
        languageOptions: {},
        output: null,
        errors: [],
      },
      {
        code: "const foo = new URL('foo.js', import.meta.url);",
        filename: 'upstream.js',
        languageOptions: {},
        output: null,
        errors: [],
      },
      {
        code: 'function run() {\n\tif (foo) {\n\t\treturn;\n\t}\n}\nrun();',
        filename: 'upstream.js',
        languageOptions: {},
        output: null,
        errors: [],
      },
      {
        code: "const file = path.join(import.meta.dirname, 'foo.js');",
        filename: 'upstream.js',
        languageOptions: {},
        output: null,
        errors: [],
      },
      {
        code: "const file = path.join(path.dirname(url.fileURLToPath(import.meta.url)), 'foo.js');",
        filename: 'upstream.js',
        languageOptions: {},
        output: null,
        errors: [],
      },
      {
        code: "const content = fs.readFileSync(import.meta.filename, 'utf8');",
        filename: 'upstream.js',
        languageOptions: {},
        output: null,
        errors: [],
      },
      {
        code: "import {fromPairs} from 'lodash-es';",
        filename: 'upstream.js',
        languageOptions: {},
        output: null,
        errors: [],
      },
      {
        code: 'export default foo;',
        filename: 'upstream.js',
        languageOptions: {},
        output: null,
        errors: [],
      },
      {
        code: 'const foo = 1;export {foo};',
        filename: 'upstream.js',
        languageOptions: {},
        output: null,
        errors: [],
      },
    ],
    invalid: [
      {
        code: "'use strict';",
        filename: 'upstream.js',
        languageOptions: {},
        output: null,
        errors: [
          {
            messageId: 'error/use-strict-directive',
            message: 'Do not use "use strict" directive.',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 14,
            suggestions: [
              {
                messageId: 'suggestion/use-strict-directive',
                desc: 'Remove "use strict" directive.',
                output: '',
              },
            ],
          },
        ],
      },
      {
        code: 'if (foo) {\n\treturn;\n}',
        filename: 'upstream.js',
        languageOptions: {
          sourceType: 'script',
          parserOptions: {
            ecmaFeatures: {
              globalReturn: true,
            },
          },
        },
        output: null,
        errors: [
          {
            messageId: 'error/global-return',
            message: '"return" should be used inside a function.',
            line: 2,
            column: 2,
            endLine: 2,
            endColumn: 8,
            suggestions: [],
          },
        ],
      },
      {
        code: "const file = path.join(__dirname, 'foo.js');",
        filename: 'upstream.js',
        languageOptions: {},
        output: null,
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "__dirname".',
            line: 1,
            column: 24,
            endLine: 1,
            endColumn: 33,
            suggestions: [
              {
                messageId: 'suggestion/import-meta-dirname',
                desc: 'Replace `__dirname` with `import.meta.dirname`.',
                output:
                  "const file = path.join(import.meta.dirname, 'foo.js');",
              },
              {
                messageId: 'suggestion/import-meta-url-to-dirname',
                desc: 'Replace `__dirname` with `…(import.meta.url)`.',
                output:
                  "const file = path.join(path.dirname(url.fileURLToPath(import.meta.url)), 'foo.js');",
              },
            ],
          },
        ],
      },
      {
        code: "const content = fs.readFileSync(__filename, 'utf8');",
        filename: 'upstream.js',
        languageOptions: {},
        output: null,
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "__filename".',
            line: 1,
            column: 33,
            endLine: 1,
            endColumn: 43,
            suggestions: [
              {
                messageId: 'suggestion/import-meta-filename',
                desc: 'Replace `__filename` with `import.meta.filename`.',
                output:
                  "const content = fs.readFileSync(import.meta.filename, 'utf8');",
              },
              {
                messageId: 'suggestion/import-meta-url-to-filename',
                desc: 'Replace `__filename` with `…(import.meta.url)`.',
                output:
                  "const content = fs.readFileSync(url.fileURLToPath(import.meta.url), 'utf8');",
              },
            ],
          },
        ],
      },
      {
        code: "const {fromPairs} = require('lodash');",
        filename: 'upstream.js',
        languageOptions: {},
        output: null,
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 21,
            endLine: 1,
            endColumn: 28,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: "import {fromPairs} from 'lodash';",
              },
            ],
          },
        ],
      },
      {
        code: 'module.exports = foo;',
        filename: 'upstream.js',
        languageOptions: {},
        output: null,
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "module".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 7,
            suggestions: [
              {
                messageId: 'suggestion/export',
                desc: 'Switch to `export`.',
                output: 'export default foo;',
              },
            ],
          },
        ],
      },
      {
        code: 'exports.foo = foo;',
        filename: 'upstream.js',
        languageOptions: {},
        output: null,
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "exports".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 8,
            suggestions: [
              {
                messageId: 'suggestion/export',
                desc: 'Switch to `export`.',
                output: 'export {foo as foo};',
              },
            ],
          },
        ],
      },
    ],
  },
];

for (const group of groups) {
  describe(group.name, () => {
    test('valid', async () => {
      for (const item of group.valid) {
        // The compiler's file selection does not accept this extension. The
        // case remains here and Go extras check the rule's suffix handling.
        if (item.filename.endsWith('.cjS')) continue;
        const result = await run(item);
        assert.deepEqual(result.diagnostics, [], item.code);
      }
    });
    test('diagnostics and edits', async () => {
      for (const item of group.invalid) {
        const result = await run(item);
        assert.equal(result.diagnostics.length, item.errors.length, item.code);
        for (const [index, diagnostic] of result.diagnostics.entries()) {
          const expected = item.errors[index];
          assert.equal(diagnostic.ruleName, 'unicorn/prefer-module', item.code);
          assert.equal(diagnostic.messageId, expected.messageId, item.code);
          assert.equal(diagnostic.message, expected.message, item.code);
          assert.equal(diagnostic.severity, 'error', item.code);
          assert.deepEqual(
            diagnostic.range,
            {
              start: { line: expected.line, column: expected.column },
              end: { line: expected.endLine, column: expected.endColumn },
            },
            item.code,
          );
          assert.deepEqual(
            (diagnostic.suggestions ?? []).map((s) => ({
              messageId: s.messageId,
              desc: s.message,
              output: apply(item.code, s.fixes ?? []),
            })),
            expected.suggestions,
            item.code,
          );
        }
        const fixes = result.diagnostics.flatMap((d) => d.fixes ?? []);
        assert.equal(
          fixes.length ? apply(item.code, fixes) : null,
          item.output,
          item.code,
        );
      }
    });
  });
}

it.skip('mixed-case foo.cjS is excluded by compiler file selection', () => {});

async function run(item: Case) {
  const filename = path.resolve(import.meta.dirname, item.filename);
  const result = await lint({
    config: [
      {
        plugins: ['unicorn'],
        languageOptions: {
          ...item.languageOptions,
          globals: { ...globals, ...item.languageOptions?.globals },
          parserOptions: { project: false, projectService: false },
        },
        rules: { 'unicorn/prefer-module': 'error' },
      },
    ],
    configDirectory: import.meta.dirname,
    workingDirectory: process.cwd(),
    fileContents: { [filename]: item.code },
  });
  assert.equal(result.fileCount, 1, item.code);
  assert.equal(result.ruleCount, 1, item.code);
  return result;
}

function apply(
  code: string,
  fixes: { startPos: number; endPos: number; text: string }[],
) {
  for (const fix of [...fixes].sort((a, b) => b.startPos - a.startPos)) {
    code = code.slice(0, fix.startPos) + fix.text + code.slice(fix.endPos);
  }
  return code;
}
