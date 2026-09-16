import path from 'node:path';
import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester({
  languageOptions: {
    sourceType: 'module',
    globals: {
      __dirname: 'readonly',
      __filename: 'readonly',
      require: 'readonly',
    },
  },
});

// All cases and documentation examples from eslint-plugin-n v18.3.0:
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-path-concat.js
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-path-concat.md
ruleTester.run(
  'no-path-concat',
  {},
  {
    valid: [
      { name: 'upstream 1', code: 'var fullPath = dirname + "foo.js";' },
      { name: 'upstream 2', code: 'var fullPath = __dirname == "foo.js";' },
      { name: 'upstream 3', code: 'if (fullPath === __dirname) {}' },
      { name: 'upstream 4', code: 'if (__dirname === fullPath) {}' },
      { name: 'upstream 5', code: 'var fullPath = "/foo.js" + __filename;' },
      { name: 'upstream 6', code: 'var fullPath = "/foo.js" + __dirname;' },
      { name: 'upstream 7', code: 'var fullPath = __filename + ".map";' },
      { name: 'upstream 8', code: 'var fullPath = `${__filename}.map`;' },
      {
        name: 'upstream 9',
        code: 'var fullPath = __filename + (test ? ".js" : ".ts");',
      },
      {
        name: 'upstream 10',
        code: 'var fullPath = __filename + (ext || ".js");',
      },
      {
        name: 'upstream 11',
        code: 'var fullPath = import.meta.dirname + ".map";',
      },
      {
        name: 'upstream 12',
        code: 'var fullPath = import.meta.filename + ".map";',
      },
      { name: 'upstream 13', code: 'var fullUrl = import.meta.url + ".map";' },
      {
        name: 'documentation 1',
        code: 'var fullPath = path.join(__dirname, "foo.js");',
      },
      {
        name: 'documentation 2',
        code: 'var fullPath = path.resolve(__dirname, "foo.js");',
      },
      {
        name: 'documentation 3',
        code: 'const url = new URL("./foo.js", import.meta.url)',
      },
      {
        name: 'documentation 4',
        code: 'const fullPath1 = path.join(__dirname, "foo.js");\nconst fullPath2 = path.join(__filename, "foo.js");\nconst fullPath3 = __dirname + ".js";\nconst fullPath4 = __filename + ".map";\nconst fullPath5 = `${__dirname}_foo.js`;\nconst fullPath6 = `${__filename}.test.js`;\nconst fullPath7 = path.join(import.meta.dirname, "foo.js");\nconst fullPath8 = path.join(import.meta.filename, "foo.js");\nconst fullUrl = new URL("./foo.js", import.meta.url);',
      },
    ],
    invalid: [
      {
        name: 'upstream 1',
        code: 'var fullPath = __dirname + "/foo.js";',
        errors: [
          {
            messageId: 'usePathFunctions',
            message:
              'Use path.join() or path.resolve() instead of string concatenation.',
            line: 1,
            column: 16,
            endLine: 1,
            endColumn: 37,
          },
        ],
      },
      {
        name: 'upstream 2',
        code: 'var fullPath = __filename + "/foo.js";',
        errors: [
          {
            messageId: 'usePathFunctions',
            message:
              'Use path.join() or path.resolve() instead of string concatenation.',
            line: 1,
            column: 16,
            endLine: 1,
            endColumn: 38,
          },
        ],
      },
      {
        name: 'upstream 3',
        code: 'var fullPath = `${__dirname}/foo.js`;',
        errors: [
          {
            messageId: 'usePathFunctions',
            message:
              'Use path.join() or path.resolve() instead of string concatenation.',
            line: 1,
            column: 16,
            endLine: 1,
            endColumn: 37,
          },
        ],
      },
      {
        name: 'upstream 4',
        code: 'var fullPath = `${__filename}/foo.js`;',
        errors: [
          {
            messageId: 'usePathFunctions',
            message:
              'Use path.join() or path.resolve() instead of string concatenation.',
            line: 1,
            column: 16,
            endLine: 1,
            endColumn: 38,
          },
        ],
      },
      {
        name: 'upstream 5',
        code: 'var path = require("path"); var fullPath = `${__dirname}${path.sep}foo.js`;',
        errors: [
          {
            messageId: 'usePathFunctions',
            message:
              'Use path.join() or path.resolve() instead of string concatenation.',
            line: 1,
            column: 44,
            endLine: 1,
            endColumn: 75,
          },
        ],
      },
      {
        name: 'upstream 6',
        code: 'var path = require("path"); var fullPath = `${__filename}${path.sep}foo.js`;',
        errors: [
          {
            messageId: 'usePathFunctions',
            message:
              'Use path.join() or path.resolve() instead of string concatenation.',
            line: 1,
            column: 44,
            endLine: 1,
            endColumn: 76,
          },
        ],
      },
      {
        name: 'upstream 7',
        code: 'var path = require("path"); var fullPath = __dirname + path.sep + `foo.js`;',
        errors: [
          {
            messageId: 'usePathFunctions',
            message:
              'Use path.join() or path.resolve() instead of string concatenation.',
            line: 1,
            column: 44,
            endLine: 1,
            endColumn: 64,
          },
        ],
      },
      {
        name: 'upstream 8',
        code: 'var fullPath = __dirname + "/" + "foo.js";',
        errors: [
          {
            messageId: 'usePathFunctions',
            message:
              'Use path.join() or path.resolve() instead of string concatenation.',
            line: 1,
            column: 16,
            endLine: 1,
            endColumn: 31,
          },
        ],
      },
      {
        name: 'upstream 9',
        code: 'var fullPath = __dirname + ("/" + "foo.js");',
        errors: [
          {
            messageId: 'usePathFunctions',
            message:
              'Use path.join() or path.resolve() instead of string concatenation.',
            line: 1,
            column: 16,
            endLine: 1,
            endColumn: 44,
          },
        ],
      },
      {
        name: 'upstream 10',
        code: 'var fullPath = __dirname + (test ? "/foo.js" : "/bar.js");',
        errors: [
          {
            messageId: 'usePathFunctions',
            message:
              'Use path.join() or path.resolve() instead of string concatenation.',
            line: 1,
            column: 16,
            endLine: 1,
            endColumn: 58,
          },
        ],
      },
      {
        name: 'upstream 11',
        code: 'var fullPath = __dirname + (extraPath || "/default.js");',
        errors: [
          {
            messageId: 'usePathFunctions',
            message:
              'Use path.join() or path.resolve() instead of string concatenation.',
            line: 1,
            column: 16,
            endLine: 1,
            endColumn: 56,
          },
        ],
      },
      {
        name: 'upstream 12',
        code: 'var fullPath = __dirname + "\\' + path.sep + 'foo.js";',
        errors: [
          {
            messageId: 'usePathFunctions',
            message:
              'Use path.join() or path.resolve() instead of string concatenation.',
            line: 1,
            column: 16,
            endLine: 1,
            endColumn: 38,
          },
        ],
      },
      {
        name: 'upstream 13',
        code: 'var fullPath = __filename + "\\' + path.sep + 'foo.js";',
        errors: [
          {
            messageId: 'usePathFunctions',
            message:
              'Use path.join() or path.resolve() instead of string concatenation.',
            line: 1,
            column: 16,
            endLine: 1,
            endColumn: 39,
          },
        ],
      },
      {
        name: 'upstream 14',
        code: 'var fullPath = `${__dirname}\\' + path.sep + 'foo.js`;',
        errors: [
          {
            messageId: 'usePathFunctions',
            message:
              'Use path.join() or path.resolve() instead of string concatenation.',
            line: 1,
            column: 16,
            endLine: 1,
            endColumn: 38,
          },
        ],
      },
      {
        name: 'upstream 15',
        code: 'var fullPath = `${__filename}\\' + path.sep + 'foo.js`;',
        errors: [
          {
            messageId: 'usePathFunctions',
            message:
              'Use path.join() or path.resolve() instead of string concatenation.',
            line: 1,
            column: 16,
            endLine: 1,
            endColumn: 39,
          },
        ],
      },
      {
        name: 'upstream 16',
        code: 'var fullPath = import.meta.dirname + "/foo.js";',
        errors: [
          {
            messageId: 'usePathFunctions',
            message:
              'Use path.join() or path.resolve() instead of string concatenation.',
            line: 1,
            column: 16,
            endLine: 1,
            endColumn: 47,
          },
        ],
      },
      {
        name: 'upstream 17',
        code: 'var fullPath = import.meta.filename + "/foo.js";',
        errors: [
          {
            messageId: 'usePathFunctions',
            message:
              'Use path.join() or path.resolve() instead of string concatenation.',
            line: 1,
            column: 16,
            endLine: 1,
            endColumn: 48,
          },
        ],
      },
      {
        name: 'upstream 18',
        code: 'var fullUrl = import.meta.url + "/foo.js";',
        errors: [
          {
            messageId: 'useUrl',
            message: 'Use new URL() instead of string concatenation.',
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 42,
          },
        ],
      },
      {
        name: 'upstream 19',
        code: 'var fullUrl = import.meta["url"] + "/foo.js";',
        errors: [
          {
            messageId: 'useUrl',
            message: 'Use new URL() instead of string concatenation.',
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 45,
          },
        ],
      },
      {
        name: 'upstream 20',
        code: 'var fullUrl = `${import.meta[`url`]}/foo.js`;',
        errors: [
          {
            messageId: 'useUrl',
            message: 'Use new URL() instead of string concatenation.',
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 45,
          },
        ],
      },
      {
        name: 'documentation 1',
        code: 'var fullPath = __dirname + "/foo.js";',
        errors: [
          {
            messageId: 'usePathFunctions',
            message:
              'Use path.join() or path.resolve() instead of string concatenation.',
            line: 1,
            column: 16,
            endLine: 1,
            endColumn: 37,
          },
        ],
      },
      {
        name: 'documentation 2',
        code: 'const fullPath1 = __dirname + "/foo.js";\nconst fullPath2 = __filename + "/foo.js";\nconst fullPath3 = `${__dirname}/foo.js`;\nconst fullPath4 = `${__filename}/foo.js`;\nconst fullPath5 = import.meta.dirname + "/foo.js";\nconst fullPath6 = import.meta.filename + "/foo.js";\nconst fullUrl = import.meta.url + "/foo.js";',
        errors: [
          {
            messageId: 'usePathFunctions',
            message:
              'Use path.join() or path.resolve() instead of string concatenation.',
            line: 1,
            column: 19,
            endLine: 1,
            endColumn: 40,
          },
          {
            messageId: 'usePathFunctions',
            message:
              'Use path.join() or path.resolve() instead of string concatenation.',
            line: 2,
            column: 19,
            endLine: 2,
            endColumn: 41,
          },
          {
            messageId: 'usePathFunctions',
            message:
              'Use path.join() or path.resolve() instead of string concatenation.',
            line: 3,
            column: 19,
            endLine: 3,
            endColumn: 40,
          },
          {
            messageId: 'usePathFunctions',
            message:
              'Use path.join() or path.resolve() instead of string concatenation.',
            line: 4,
            column: 19,
            endLine: 4,
            endColumn: 41,
          },
          {
            messageId: 'usePathFunctions',
            message:
              'Use path.join() or path.resolve() instead of string concatenation.',
            line: 5,
            column: 19,
            endLine: 5,
            endColumn: 50,
          },
          {
            messageId: 'usePathFunctions',
            message:
              'Use path.join() or path.resolve() instead of string concatenation.',
            line: 6,
            column: 19,
            endLine: 6,
            endColumn: 51,
          },
          {
            messageId: 'useUrl',
            message: 'Use new URL() instead of string concatenation.',
            line: 7,
            column: 17,
            endLine: 7,
            endColumn: 44,
          },
        ],
      },
    ],
  },
);
