import { RuleTester } from '../rule-tester';

// Every upstream case and documentation example, pinned to eslint-plugin-n v18.3.0.
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/exports-style.js
// Exact messages, ranges and fixes were checked with ESLint 10.9.0.
new RuleTester({
  languageOptions: {
    sourceType: 'commonjs',
    globals: { exports: 'writable', module: 'readonly' },
  },
}).run(
  'exports-style',
  {},
  {
    valid: [
      {
        name: 'upstream valid 1',
        code: 'module.exports = {foo: 1}',
      },
      {
        name: 'upstream valid 2',
        code: 'module.exports = {foo: 1}',
        options: ['module.exports'],
      },
      {
        name: 'upstream valid 3',
        code: 'exports.foo = 1',
        options: ['exports'],
      },
      {
        name: 'upstream valid 4',
        code: 'exports = module.exports = {foo: 1}',
        options: [
          'module.exports',
          {
            allowBatchAssign: true,
          },
        ],
      },
      {
        name: 'upstream valid 5',
        code: 'module.exports = exports = {foo: 1}',
        options: [
          'module.exports',
          {
            allowBatchAssign: true,
          },
        ],
      },
      {
        name: 'upstream valid 6',
        code: 'exports = module.exports = {foo: 1}',
        options: [
          'exports',
          {
            allowBatchAssign: true,
          },
        ],
      },
      {
        name: 'upstream valid 7',
        code: 'module.exports = exports = {foo: 1}',
        options: [
          'exports',
          {
            allowBatchAssign: true,
          },
        ],
      },
      {
        name: 'upstream valid 8',
        code: 'exports = module.exports = {foo: 1}; exports.bar = 2',
        options: [
          'exports',
          {
            allowBatchAssign: true,
          },
        ],
      },
      {
        name: 'upstream valid 9',
        code: 'module.exports = exports = {foo: 1}; exports.bar = 2',
        options: [
          'exports',
          {
            allowBatchAssign: true,
          },
        ],
      },
      {
        name: 'upstream valid 10',
        code: 'module = {}; module.foo = 1',
        options: ['exports'],
      },
      {
        name: 'upstream valid 11',
        code: 'exports.foo = 1',
        options: ['module.exports'],
        languageOptions: {
          globals: {
            exports: 'off',
          },
        },
      },
      {
        name: 'upstream valid 12',
        code: 'module.exports = {foo: 1}',
        options: ['exports'],
        languageOptions: {
          globals: {
            module: 'off',
          },
        },
      },
      {
        name: 'documentation 3',
        code: '/*eslint node/exports-style: ["error", "module.exports"]*/\n\nmodule.exports = {\n    foo: 1,\n    bar: 2\n}\n\nmodule.exports.baz = 3',
        options: ['module.exports'],
      },
      {
        name: 'documentation 5',
        code: '/*eslint node/exports-style: ["error", "exports"]*/\n\nexports.foo = 1\nexports.bar = 2',
        options: ['exports'],
      },
      {
        name: 'documentation 6',
        code: '/*eslint node/exports-style: ["error", "exports", {"allowBatchAssign": true}]*/\n\n// Allow `module.exports` in the same assignment expression as `exports`.\nmodule.exports = exports = function foo() {\n    // do something.\n}\n\nexports.bar = 1',
        options: [
          'exports',
          {
            allowBatchAssign: true,
          },
        ],
      },
    ],
    invalid: [
      {
        name: 'upstream invalid 1',
        code: 'exports = {foo: 1}',
        errors: [
          {
            messageId: 'unexpectedExports',
            message:
              "Unexpected access to 'exports'. Use 'module.exports' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 10,
          },
        ],
      },
      {
        name: 'upstream invalid 2',
        code: 'exports.foo = 1',
        errors: [
          {
            messageId: 'unexpectedExports',
            message:
              "Unexpected access to 'exports'. Use 'module.exports' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 9,
          },
        ],
      },
      {
        name: 'upstream invalid 3',
        code: 'module.exports = exports = {foo: 1}',
        errors: [
          {
            messageId: 'unexpectedExports',
            message:
              "Unexpected access to 'exports'. Use 'module.exports' instead.",
            line: 1,
            column: 18,
            endLine: 1,
            endColumn: 27,
          },
        ],
      },
      {
        name: 'upstream invalid 4',
        code: 'exports = module.exports = {foo: 1}',
        errors: [
          {
            messageId: 'unexpectedExports',
            message:
              "Unexpected access to 'exports'. Use 'module.exports' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 10,
          },
        ],
      },
      {
        name: 'upstream invalid 5',
        code: 'exports = {foo: 1}',
        options: ['module.exports'],
        errors: [
          {
            messageId: 'unexpectedExports',
            message:
              "Unexpected access to 'exports'. Use 'module.exports' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 10,
          },
        ],
      },
      {
        name: 'upstream invalid 6',
        code: 'exports.foo = 1',
        options: ['module.exports'],
        errors: [
          {
            messageId: 'unexpectedExports',
            message:
              "Unexpected access to 'exports'. Use 'module.exports' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 9,
          },
        ],
      },
      {
        name: 'upstream invalid 7',
        code: 'module.exports = exports = {foo: 1}',
        options: ['module.exports'],
        errors: [
          {
            messageId: 'unexpectedExports',
            message:
              "Unexpected access to 'exports'. Use 'module.exports' instead.",
            line: 1,
            column: 18,
            endLine: 1,
            endColumn: 27,
          },
        ],
      },
      {
        name: 'upstream invalid 8',
        code: 'exports = module.exports = {foo: 1}',
        options: ['module.exports'],
        errors: [
          {
            messageId: 'unexpectedExports',
            message:
              "Unexpected access to 'exports'. Use 'module.exports' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 10,
          },
        ],
      },
      {
        name: 'upstream invalid 9',
        code: 'exports = {foo: 1}',
        options: ['exports'],
        errors: [
          {
            messageId: 'unexpectedAssignment',
            message:
              "Unexpected assignment to 'exports'. Don't modify 'exports' itself.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 10,
          },
        ],
      },
      {
        name: 'upstream invalid 10',
        code: 'module.exports = {foo: 1}',
        options: ['exports'],
        errors: [
          {
            messageId: 'unexpectedModuleExports',
            message:
              "Unexpected access to 'module.exports'. Use 'exports' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 17,
            fixes: [
              {
                text: 'exports.foo = 1;',
                startPos: 0,
                endPos: 25,
              },
            ],
          },
        ],
        output: 'exports.foo = 1;',
      },
      {
        name: 'upstream invalid 11',
        code: 'module.exports.foo = 1',
        options: ['exports'],
        errors: [
          {
            messageId: 'unexpectedModuleExports',
            message:
              "Unexpected access to 'module.exports'. Use 'exports' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 16,
            fixes: [
              {
                text: 'exports',
                startPos: 0,
                endPos: 14,
              },
            ],
          },
        ],
        output: 'exports.foo = 1',
      },
      {
        name: 'upstream invalid 12',
        code: 'module.exports = { a: 1 }',
        options: ['exports'],
        errors: [
          {
            messageId: 'unexpectedModuleExports',
            message:
              "Unexpected access to 'module.exports'. Use 'exports' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 17,
            fixes: [
              {
                text: 'exports.a = 1;',
                startPos: 0,
                endPos: 25,
              },
            ],
          },
        ],
        output: 'exports.a = 1;',
      },
      {
        name: 'upstream invalid 13',
        code: 'module.exports = { a: 1, b: 2 }',
        options: ['exports'],
        errors: [
          {
            messageId: 'unexpectedModuleExports',
            message:
              "Unexpected access to 'module.exports'. Use 'exports' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 17,
            fixes: [
              {
                text: 'exports.a = 1;\n\nexports.b = 2;',
                startPos: 0,
                endPos: 31,
              },
            ],
          },
        ],
        output: 'exports.a = 1;\n\nexports.b = 2;',
      },
      {
        name: 'upstream invalid 14',
        code: 'module.exports = { // before a\na: 1, // between a and b\nb: 2 // after b\n}',
        options: ['exports'],
        errors: [
          {
            messageId: 'unexpectedModuleExports',
            message:
              "Unexpected access to 'module.exports'. Use 'exports' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 17,
            fixes: [
              {
                text: '// before a\nexports.a = 1;\n\n// between a and b\nexports.b = 2;\n// after b',
                startPos: 0,
                endPos: 73,
              },
            ],
          },
        ],
        output:
          '// before a\nexports.a = 1;\n\n// between a and b\nexports.b = 2;\n// after b',
      },
      {
        name: 'upstream invalid 15',
        code: 'foo(module.exports = {foo: 1})',
        options: ['exports'],
        errors: [
          {
            messageId: 'unexpectedModuleExports',
            message:
              "Unexpected access to 'module.exports'. Use 'exports' instead.",
            line: 1,
            column: 5,
            endLine: 1,
            endColumn: 21,
          },
        ],
      },
      {
        name: 'upstream invalid 16',
        code: 'if(foo){ module.exports = { foo: 1};} else { module.exports = {foo: 2};}',
        options: ['exports'],
        errors: [
          {
            messageId: 'unexpectedModuleExports',
            message:
              "Unexpected access to 'module.exports'. Use 'exports' instead.",
            line: 1,
            column: 10,
            endLine: 1,
            endColumn: 26,
          },
          {
            messageId: 'unexpectedModuleExports',
            message:
              "Unexpected access to 'module.exports'. Use 'exports' instead.",
            line: 1,
            column: 46,
            endLine: 1,
            endColumn: 62,
          },
        ],
      },
      {
        name: 'upstream invalid 17',
        code: 'function bar() { module.exports = { foo: 1 }; }',
        options: ['exports'],
        errors: [
          {
            messageId: 'unexpectedModuleExports',
            message:
              "Unexpected access to 'module.exports'. Use 'exports' instead.",
            line: 1,
            column: 18,
            endLine: 1,
            endColumn: 34,
          },
        ],
      },
      {
        name: 'upstream invalid 18',
        code: 'module.exports = { get a() {} }',
        options: ['exports'],
        errors: [
          {
            messageId: 'unexpectedModuleExports',
            message:
              "Unexpected access to 'module.exports'. Use 'exports' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 17,
          },
        ],
      },
      {
        name: 'upstream invalid 19',
        code: 'module.exports = { set a(a) {} }',
        options: ['exports'],
        errors: [
          {
            messageId: 'unexpectedModuleExports',
            message:
              "Unexpected access to 'module.exports'. Use 'exports' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 17,
          },
        ],
      },
      {
        name: 'upstream invalid 20',
        code: 'module.exports = { a }',
        options: ['exports'],
        errors: [
          {
            messageId: 'unexpectedModuleExports',
            message:
              "Unexpected access to 'module.exports'. Use 'exports' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 17,
            fixes: [
              {
                text: 'exports.a = a;',
                startPos: 0,
                endPos: 22,
              },
            ],
          },
        ],
        output: 'exports.a = a;',
      },
      {
        name: 'upstream invalid 21',
        code: 'module.exports = { ...a }',
        options: ['exports'],
        errors: [
          {
            messageId: 'unexpectedModuleExports',
            message:
              "Unexpected access to 'module.exports'. Use 'exports' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 17,
          },
        ],
      },
      {
        name: 'upstream invalid 22',
        code: "module.exports = { ['a' + 'b']: 1 }",
        options: ['exports'],
        errors: [
          {
            messageId: 'unexpectedModuleExports',
            message:
              "Unexpected access to 'module.exports'. Use 'exports' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 17,
            fixes: [
              {
                text: "exports['a' + 'b'] = 1;",
                startPos: 0,
                endPos: 35,
              },
            ],
          },
        ],
        output: "exports['a' + 'b'] = 1;",
      },
      {
        name: 'upstream invalid 23',
        code: "module.exports = { 'foo': 1 }",
        options: ['exports'],
        errors: [
          {
            messageId: 'unexpectedModuleExports',
            message:
              "Unexpected access to 'module.exports'. Use 'exports' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 17,
            fixes: [
              {
                text: "exports['foo'] = 1;",
                startPos: 0,
                endPos: 29,
              },
            ],
          },
        ],
        output: "exports['foo'] = 1;",
      },
      {
        name: 'upstream invalid 24',
        code: 'module.exports = { foo(a) {} }',
        options: ['exports'],
        errors: [
          {
            messageId: 'unexpectedModuleExports',
            message:
              "Unexpected access to 'module.exports'. Use 'exports' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 17,
            fixes: [
              {
                text: 'exports.foo = function (a) {};',
                startPos: 0,
                endPos: 30,
              },
            ],
          },
        ],
        output: 'exports.foo = function (a) {};',
      },
      {
        name: 'upstream invalid 25',
        code: 'module.exports = { *foo(a) {} }',
        options: ['exports'],
        errors: [
          {
            messageId: 'unexpectedModuleExports',
            message:
              "Unexpected access to 'module.exports'. Use 'exports' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 17,
            fixes: [
              {
                text: 'exports.foo = function* (a) {};',
                startPos: 0,
                endPos: 31,
              },
            ],
          },
        ],
        output: 'exports.foo = function* (a) {};',
      },
      {
        name: 'upstream invalid 26',
        code: 'module.exports = { async foo(a) {} }',
        options: ['exports'],
        errors: [
          {
            messageId: 'unexpectedModuleExports',
            message:
              "Unexpected access to 'module.exports'. Use 'exports' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 17,
            fixes: [
              {
                text: 'exports.foo = async function (a) {};',
                startPos: 0,
                endPos: 36,
              },
            ],
          },
        ],
        output: 'exports.foo = async function (a) {};',
      },
      {
        name: 'upstream invalid 27',
        code: 'module.exports.foo()',
        options: ['exports'],
        errors: [
          {
            messageId: 'unexpectedModuleExports',
            message:
              "Unexpected access to 'module.exports'. Use 'exports' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 16,
            fixes: [
              {
                text: 'exports',
                startPos: 0,
                endPos: 14,
              },
            ],
          },
        ],
        output: 'exports.foo()',
      },
      {
        name: 'upstream invalid 28',
        code: "a = module.exports.foo + module.exports['bar']",
        options: ['exports'],
        errors: [
          {
            messageId: 'unexpectedModuleExports',
            message:
              "Unexpected access to 'module.exports'. Use 'exports' instead.",
            line: 1,
            column: 5,
            endLine: 1,
            endColumn: 20,
            fixes: [
              {
                text: 'exports',
                startPos: 4,
                endPos: 18,
              },
            ],
          },
          {
            messageId: 'unexpectedModuleExports',
            message:
              "Unexpected access to 'module.exports'. Use 'exports' instead.",
            line: 1,
            column: 26,
            endLine: 1,
            endColumn: 41,
            fixes: [
              {
                text: 'exports',
                startPos: 25,
                endPos: 39,
              },
            ],
          },
        ],
        output: "a = exports.foo + exports['bar']",
      },
      {
        name: 'upstream invalid 29',
        code: 'module.exports = exports = {foo: 1}',
        options: ['exports'],
        errors: [
          {
            messageId: 'unexpectedModuleExports',
            message:
              "Unexpected access to 'module.exports'. Use 'exports' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 17,
          },
          {
            messageId: 'unexpectedAssignment',
            message:
              "Unexpected assignment to 'exports'. Don't modify 'exports' itself.",
            line: 1,
            column: 18,
            endLine: 1,
            endColumn: 27,
          },
        ],
      },
      {
        name: 'upstream invalid 30',
        code: 'exports = module.exports = {foo: 1}',
        options: ['exports'],
        errors: [
          {
            messageId: 'unexpectedAssignment',
            message:
              "Unexpected assignment to 'exports'. Don't modify 'exports' itself.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 10,
          },
          {
            messageId: 'unexpectedModuleExports',
            message:
              "Unexpected access to 'module.exports'. Use 'exports' instead.",
            line: 1,
            column: 11,
            endLine: 1,
            endColumn: 27,
          },
        ],
      },
      {
        name: 'upstream invalid 31',
        code: 'module.exports = exports = {foo: 1}; exports = obj',
        options: [
          'exports',
          {
            allowBatchAssign: true,
          },
        ],
        errors: [
          {
            messageId: 'unexpectedAssignment',
            message:
              "Unexpected assignment to 'exports'. Don't modify 'exports' itself.",
            line: 1,
            column: 38,
            endLine: 1,
            endColumn: 47,
          },
        ],
      },
      {
        name: 'upstream invalid 32',
        code: 'exports = module.exports = {foo: 1}; exports = obj',
        options: [
          'exports',
          {
            allowBatchAssign: true,
          },
        ],
        errors: [
          {
            messageId: 'unexpectedAssignment',
            message:
              "Unexpected assignment to 'exports'. Don't modify 'exports' itself.",
            line: 1,
            column: 38,
            endLine: 1,
            endColumn: 47,
          },
        ],
      },
      {
        name: 'documentation 1',
        code: 'module.exports = {\n    foo: 1\n}\n\nexports.bar = 2',
        options: [],
        errors: [
          {
            messageId: 'unexpectedExports',
            message:
              "Unexpected access to 'exports'. Use 'module.exports' instead.",
            line: 5,
            column: 1,
            endLine: 5,
            endColumn: 9,
          },
        ],
      },
      {
        name: 'documentation 2',
        code: '/*eslint node/exports-style: ["error", "module.exports"]*/\n\nexports.foo = 1\nexports.bar = 2',
        options: ['module.exports'],
        errors: [
          {
            messageId: 'unexpectedExports',
            message:
              "Unexpected access to 'exports'. Use 'module.exports' instead.",
            line: 3,
            column: 1,
            endLine: 3,
            endColumn: 9,
          },
          {
            messageId: 'unexpectedExports',
            message:
              "Unexpected access to 'exports'. Use 'module.exports' instead.",
            line: 4,
            column: 1,
            endLine: 4,
            endColumn: 9,
          },
        ],
      },
      {
        name: 'documentation 4',
        code: '/*eslint node/exports-style: ["error", "exports"]*/\n\nmodule.exports = {\n    foo: 1,\n    bar: 2\n}\n\nmodule.exports.baz = 3',
        options: ['exports'],
        errors: [
          {
            messageId: 'unexpectedModuleExports',
            message:
              "Unexpected access to 'module.exports'. Use 'exports' instead.",
            line: 3,
            column: 1,
            endLine: 3,
            endColumn: 17,
            fixes: [
              {
                text: 'exports.foo = 1;\n\nexports.bar = 2;',
                startPos: 53,
                endPos: 96,
              },
            ],
          },
          {
            messageId: 'unexpectedModuleExports',
            message:
              "Unexpected access to 'module.exports'. Use 'exports' instead.",
            line: 8,
            column: 1,
            endLine: 8,
            endColumn: 16,
            fixes: [
              {
                text: 'exports',
                startPos: 98,
                endPos: 112,
              },
            ],
          },
        ],
        output:
          '/*eslint node/exports-style: ["error", "exports"]*/\n\nexports.foo = 1;\n\nexports.bar = 2;\n\nexports.baz = 3',
      },
    ],
  },
);
