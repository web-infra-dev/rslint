import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester({
  languageOptions: { sourceType: 'commonjs' },
});

// Every case from https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-mixed-requires.js
ruleTester.run(
  'no-mixed-requires',
  {},
  {
    valid: [
      {
        name: 'upstream 0: var a, b = 42, c = doStuff()',
        code: 'var a, b = 42, c = doStuff()',
        options: [false],
      },
      {
        name: "upstream 1: var a = require(42), b = require(), c = require('y'), d = require(doStuff())",
        code: "var a = require(42), b = require(), c = require('y'), d = require(doStuff())",
        options: [false],
      },
      {
        name: "upstream 2: var fs = require('fs'), foo = require('foo')",
        code: "var fs = require('fs'), foo = require('foo')",
        options: [false],
      },
      {
        name: "upstream 3: var exec = require('child_process').exec, foo = require('foo')",
        code: "var exec = require('child_process').exec, foo = require('foo')",
        options: [false],
      },
      {
        name: "upstream 4: var fs = require('fs'), foo = require('./foo')",
        code: "var fs = require('fs'), foo = require('./foo')",
        options: [false],
      },
      {
        name: "upstream 5: var foo = require('foo'), foo2 = require('./foo')",
        code: "var foo = require('foo'), foo2 = require('./foo')",
        options: [false],
      },
      {
        name: "upstream 6: var emitter = require('events').EventEmitter, fs = require('fs')",
        code: "var emitter = require('events').EventEmitter, fs = require('fs')",
        options: [false],
      },
      {
        name: 'upstream 7: var foo = require(42), bar = require(getName())',
        code: 'var foo = require(42), bar = require(getName())',
        options: [false],
      },
      {
        name: 'upstream 8: var foo = require(42), bar = require(getName())',
        code: 'var foo = require(42), bar = require(getName())',
        options: [true],
      },
      {
        name: "upstream 9: var fs = require('fs'), foo = require('./foo')",
        code: "var fs = require('fs'), foo = require('./foo')",
        options: [
          {
            grouping: false,
          },
        ],
      },
      {
        name: "upstream 10: var foo = require('foo'), bar = require(getName())",
        code: "var foo = require('foo'), bar = require(getName())",
        options: [false],
      },
      {
        name: 'upstream 11: var a;',
        code: 'var a;',
        options: [true],
      },
      {
        name: "upstream 12: var async = require('async'), debug = require('diagnostics')('my-module')",
        code: "var async = require('async'), debug = require('diagnostics')('my-module')",
        options: [
          {
            allowCall: true,
          },
        ],
      },
    ],
    invalid: [
      {
        name: "upstream 0: var fs = require('fs'), foo = 42",
        code: "var fs = require('fs'), foo = 42",
        options: [false],
        errors: [
          {
            messageId: 'noMixRequire',
            message: "Do not mix 'require' and other declarations.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 33,
          },
        ],
      },
      {
        name: "upstream 1: var fs = require('fs'), foo",
        code: "var fs = require('fs'), foo",
        options: [false],
        errors: [
          {
            messageId: 'noMixRequire',
            message: "Do not mix 'require' and other declarations.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 28,
          },
        ],
      },
      {
        name: "upstream 2: var a = require(42), b = require(), c = require('y'), d = require(doStuff())",
        code: "var a = require(42), b = require(), c = require('y'), d = require(doStuff())",
        options: [true],
        errors: [
          {
            messageId: 'noMixCoreModuleFileComputed',
            message: 'Do not mix core, module, file and computed requires.',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 77,
          },
        ],
      },
      {
        name: "upstream 3: var fs = require('fs'), foo = require('foo')",
        code: "var fs = require('fs'), foo = require('foo')",
        options: [true],
        errors: [
          {
            messageId: 'noMixCoreModuleFileComputed',
            message: 'Do not mix core, module, file and computed requires.',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 45,
          },
        ],
      },
      {
        name: "upstream 4: var fs = require('fs'), foo = require('foo')",
        code: "var fs = require('fs'), foo = require('foo')",
        options: [
          {
            grouping: true,
          },
        ],
        errors: [
          {
            messageId: 'noMixCoreModuleFileComputed',
            message: 'Do not mix core, module, file and computed requires.',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 45,
          },
        ],
      },
      {
        name: "upstream 5: var exec = require('child_process').exec, foo = require('foo')",
        code: "var exec = require('child_process').exec, foo = require('foo')",
        options: [true],
        errors: [
          {
            messageId: 'noMixCoreModuleFileComputed',
            message: 'Do not mix core, module, file and computed requires.',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 63,
          },
        ],
      },
      {
        name: "upstream 6: var fs = require('fs'), foo = require('./foo')",
        code: "var fs = require('fs'), foo = require('./foo')",
        options: [true],
        errors: [
          {
            messageId: 'noMixCoreModuleFileComputed',
            message: 'Do not mix core, module, file and computed requires.',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 47,
          },
        ],
      },
      {
        name: "upstream 7: var foo = require('foo'), foo2 = require('./foo')",
        code: "var foo = require('foo'), foo2 = require('./foo')",
        options: [true],
        errors: [
          {
            messageId: 'noMixCoreModuleFileComputed',
            message: 'Do not mix core, module, file and computed requires.',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 50,
          },
        ],
      },
      {
        name: "upstream 8: var foo = require('foo'), bar = require(getName())",
        code: "var foo = require('foo'), bar = require(getName())",
        options: [true],
        errors: [
          {
            messageId: 'noMixCoreModuleFileComputed',
            message: 'Do not mix core, module, file and computed requires.',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 51,
          },
        ],
      },
      {
        name: "upstream 9: var async = require('async'), debug = require('diagnostics').someFun('my-module')",
        code: "var async = require('async'), debug = require('diagnostics').someFun('my-module')",
        options: [
          {
            allowCall: true,
          },
        ],
        errors: [
          {
            messageId: 'noMixRequire',
            message: "Do not mix 'require' and other declarations.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 82,
          },
        ],
      },
    ],
  },
);

// Every case from https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-mixed-requires.md
ruleTester.run(
  'no-mixed-requires',
  {},
  {
    valid: [
      {
        name: 'documentation 0: // only require declarations (grouping off)',
        code: "// only require declarations (grouping off)\nvar eventEmitter = require('events').EventEmitter,\n    myUtils = require('./utils'),\n    util = require('util'),\n    bar = require(getBarModuleName());\n\n// only non-require declarations\nvar foo = 42,\n    bar = 'baz';\n\n// always valid regardless of grouping because all declarations are of the same type\nvar foo = require('foo' + VERSION),\n    bar = require(getBarModuleName()),\n    baz = require();",
      },
      {
        name: "documentation 1: var async = require('async'),",
        code: "var async = require('async'),\n    debug = require('diagnostics')('my-module'),\n    eslint = require('eslint');",
        options: [
          {
            allowCall: true,
          },
        ],
      },
    ],
    invalid: [
      {
        name: 'documentation 0: var fs = require(\'fs\'),        // "core"     \\',
        code: `var fs = require('fs'),        // "core"     \\
    async = require('async'),  // "module"   |- these are "require declaration"s
    foo = require('./foo'),    // "file"     |
    bar = require(getName()),  // "computed" /
    baz = 42,                  // "other"
    bam;                       // "uninitialized"`,
        errors: [
          {
            messageId: 'noMixRequire',
            message: "Do not mix 'require' and other declarations.",
            line: 1,
            column: 1,
            endLine: 6,
            endColumn: 9,
          },
        ],
      },
      {
        name: "documentation 1: var fs = require('fs'),",
        code: "var fs = require('fs'),\n    i = 0;\n\nvar async = require('async'),\n    debug = require('diagnostics').someFunction('my-module'),\n    eslint = require('eslint');",
        errors: [
          {
            messageId: 'noMixRequire',
            message: "Do not mix 'require' and other declarations.",
            line: 1,
            column: 1,
            endLine: 2,
            endColumn: 11,
          },
          {
            messageId: 'noMixRequire',
            message: "Do not mix 'require' and other declarations.",
            line: 4,
            column: 1,
            endLine: 6,
            endColumn: 32,
          },
        ],
      },
      {
        name: 'documentation 2: // invalid because of mixed types "core" and "module"',
        code: `// invalid because of mixed types "core" and "module"
var fs = require('fs'),
    async = require('async');

// invalid because of mixed types "file" and "unknown"
var foo = require('foo'),
    bar = require(getBarModuleName());`,
        options: [
          {
            grouping: true,
          },
        ],
        errors: [
          {
            messageId: 'noMixCoreModuleFileComputed',
            message: 'Do not mix core, module, file and computed requires.',
            line: 2,
            column: 1,
            endLine: 3,
            endColumn: 30,
          },
          {
            messageId: 'noMixCoreModuleFileComputed',
            message: 'Do not mix core, module, file and computed requires.',
            line: 6,
            column: 1,
            endLine: 7,
            endColumn: 39,
          },
        ],
      },
      {
        name: "documentation 3: var async = require('async'),",
        code: "var async = require('async'),\n    debug = require('diagnostics').someFunction('my-module'), /* allowCall doesn't allow calling any function */\n    eslint = require('eslint');",
        options: [
          {
            allowCall: true,
          },
        ],
        errors: [
          {
            messageId: 'noMixRequire',
            message: "Do not mix 'require' and other declarations.",
            line: 1,
            column: 1,
            endLine: 3,
            endColumn: 32,
          },
        ],
      },
    ],
  },
);
