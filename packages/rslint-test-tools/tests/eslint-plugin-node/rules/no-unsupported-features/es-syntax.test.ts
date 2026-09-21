// eslint-plugin-n v18.3.0 / eslint-plugin-es-x v7.8.0 upstream mirror.
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-unsupported-features/es-syntax.js
import { RuleTester } from '../../rule-tester';
const ruleTester = new RuleTester({
  languageOptions: { sourceType: 'commonjs' },
  fixtureFiles: {
    'package.json': '{}\n',
    'tsconfig.json':
      '{"compilerOptions":{"allowJs":true,"target":"ESNext","module":"ESNext"},"include":["**/*.js","**/*.ts","**/*.tsx"]}\n',
    'dev-engines-array-gte-7.5.0/package.json':
      '{\n    "name": "test",\n    "devEngines": {\n        "runtime": [\n            {\n                "name": "deno",\n                "version": ">=1.0.0"\n            },\n            {\n                "name": "node",\n                "version": ">=7.5.0"\n            }\n        ]\n    }\n}\n',
    'dev-engines-gte-7.5.0/package.json':
      '{\n    "name": "test",\n    "devEngines": {\n        "runtime": {\n            "name": "node",\n            "version": ">=7.5.0"\n        }\n    }\n}\n',
    'dev-engines-non-node/package.json':
      '{\n    "name": "test",\n    "engines": {\n        "node": ">=7.5.0"\n    },\n    "devEngines": {\n        "runtime": {\n            "name": "deno",\n            "version": ">=1.0.0"\n        }\n    }\n}\n',
    'dev-engines-non-node-only/package.json':
      '{\n    "name": "test",\n    "devEngines": {\n        "runtime": {\n            "name": "deno",\n            "version": ">=1.0.0"\n        }\n    }\n}\n',
    'engines-over-dev-engines/package.json':
      '{\n    "name": "test",\n    "engines": {\n        "node": ">=4.0.0"\n    },\n    "devEngines": {\n        "runtime": {\n            "name": "node",\n            "version": ">=7.5.0"\n        }\n    }\n}\n',
    'gte-0.12.8/package.json':
      '{\n    "name": "test",\n    "engines": {\n        "node": ">=0.12.8"\n    }\n}\n',
    'gte-4.0.0/package.json':
      '{\n    "name": "test",\n    "engines": {\n        "node": ">=4.0.0"\n    }\n}\n',
    'gte-4.4.0-lt-5.0.0/package.json':
      '{\n    "name": "test",\n    "engines": {\n        "node": ">=4.4.0 <5.0.0"\n    }\n}\n',
    'gte-7.10.0/package.json':
      '{\n    "name": "test",\n    "engines": {\n        "node": ">=7.10.0"\n    }\n}\n',
    'gte-7.5.0/package.json':
      '{\n    "name": "test",\n    "engines": {\n        "node": ">=7.5.0"\n    }\n}\n',
    'gte-7.6.0/package.json':
      '{\n    "name": "test",\n    "engines": {\n        "node": ">=7.6.0"\n    }\n}\n',
    'hat-4.1.2/package.json':
      '{\n    "name": "test",\n    "engines": {\n        "node": "^4.1.2"\n    }\n}\n',
    'invalid/package.json':
      '{\n    "name": "test",\n    "engines": {\n        "node": "foo-bar"\n    }\n}\n',
    'lt-6.0.0/package.json':
      '{\n    "name": "test",\n    "engines": {\n        "node": "<6.0.0"\n    }\n}\n',
    'nothing/package.json': '{\n    "name": "test"\n}\n',
    'star/package.json':
      '{\n    "name": "test",\n    "engines": {\n        "node": "*"\n    }\n}\n',
    'without-node/package.json':
      '{\n    "name": "test",\n    "engines": {\n        "vscode": "foo"\n    }\n}\n',
  },
});

// arrowFunctions
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        code: 'function f() {}',
        options: [
          {
            version: '3.9.9',
          },
        ],
        name: 'arrowFunctions valid 1',
      },
      {
        code: '!function f() {}',
        options: [
          {
            version: '3.9.9',
          },
        ],
        name: 'arrowFunctions valid 2',
      },
      {
        code: '(() => 1)',
        options: [
          {
            version: '4.0.0',
          },
        ],
        name: 'arrowFunctions valid 3',
      },
      {
        code: '(() => {})',
        options: [
          {
            version: '4.0.0',
          },
        ],
        name: 'arrowFunctions valid 4',
      },
      {
        code: '(() => 1)',
        options: [
          {
            version: '3.9.9',
            ignores: ['arrowFunctions'],
          },
        ],
        name: 'arrowFunctions valid 5',
      },
      {
        code: '(() => {})',
        options: [
          {
            version: '3.9.9',
            ignores: ['arrowFunctions'],
          },
        ],
        name: 'arrowFunctions valid 6',
      },
    ],
    invalid: [
      {
        code: '(() => 1)',
        options: [
          {
            version: '3.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'arrow-functions' is not supported until Node.js >=4.0.0. The configured version range is '3.9.9'.",
            line: 1,
            column: 2,
            endLine: 1,
            endColumn: 9,
          },
        ],
        name: 'arrowFunctions invalid 1',
      },
      {
        code: '(() => {})',
        options: [
          {
            version: '3.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'arrow-functions' is not supported until Node.js >=4.0.0. The configured version range is '3.9.9'.",
            line: 1,
            column: 2,
            endLine: 1,
            endColumn: 10,
          },
        ],
        name: 'arrowFunctions invalid 2',
      },
    ],
  },
);

// binaryNumericLiterals
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        code: '0x01',
        options: [
          {
            version: '3.9.9',
          },
        ],
        name: 'binaryNumericLiterals valid 1',
      },
      {
        code: '1',
        options: [
          {
            version: '3.9.9',
          },
        ],
        name: 'binaryNumericLiterals valid 2',
      },
      {
        code: '0b01',
        options: [
          {
            version: '4.0.0',
          },
        ],
        name: 'binaryNumericLiterals valid 3',
      },
      {
        code: '0b01',
        options: [
          {
            version: '3.9.9',
            ignores: ['binaryNumericLiterals'],
          },
        ],
        name: 'binaryNumericLiterals valid 4',
      },
    ],
    invalid: [
      {
        code: '0b01',
        options: [
          {
            version: '3.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'binary-numeric-literals' is not supported until Node.js >=4.0.0. The configured version range is '3.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 5,
          },
        ],
        name: 'binaryNumericLiterals invalid 1',
      },
    ],
  },
);

// blockScopedFunctions
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        code: "'use strict'; if (a) { function f() {} }",
        options: [
          {
            version: '4.0.0',
          },
        ],
        name: 'blockScopedFunctions valid 1',
      },
      {
        code: "'use strict'; function wrap() { if (a) { function f() {} } }",
        options: [
          {
            version: '4.0.0',
          },
        ],
        name: 'blockScopedFunctions valid 2',
      },
      {
        code: "function wrap() { 'use strict'; if (a) { function f() {} } }",
        options: [
          {
            version: '4.0.0',
          },
        ],
        name: 'blockScopedFunctions valid 3',
      },
      {
        code: 'if (a) { function f() {} }',
        options: [
          {
            version: '6.0.0',
          },
        ],
        name: 'blockScopedFunctions valid 4',
      },
      {
        code: "'use strict'; if (a) { function f() {} }",
        options: [
          {
            version: '3.9.9',
            ignores: ['blockScopedFunctions'],
          },
        ],
        name: 'blockScopedFunctions valid 5',
      },
      {
        code: 'if (a) { function f() {} }',
        options: [
          {
            version: '5.9.9',
            ignores: ['blockScopedFunctions'],
          },
        ],
        name: 'blockScopedFunctions valid 6',
      },
    ],
    invalid: [
      {
        code: "'use strict'; if (a) { function f() {} }",
        options: [
          {
            version: '3.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'block-scoped-functions' is not supported until Node.js >=4.0.0. The configured version range is '3.9.9'.",
            line: 1,
            column: 24,
            endLine: 1,
            endColumn: 39,
          },
        ],
        name: 'blockScopedFunctions invalid 1',
      },
      {
        code: 'if (a) { function f() {} }',
        options: [
          {
            version: '5.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'block-scoped-functions' is not supported until Node.js >=6.0.0. The configured version range is '5.9.9'.",
            line: 1,
            column: 10,
            endLine: 1,
            endColumn: 25,
          },
        ],
        name: 'blockScopedFunctions invalid 2',
      },
    ],
  },
);

// blockScopedVariables
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        code: 'var a = 0',
        options: [
          {
            version: '3.9.9',
          },
        ],
        name: 'blockScopedVariables valid 1',
      },
      {
        code: "'use strict'; let a = 0",
        options: [
          {
            version: '4.0.0',
          },
        ],
        name: 'blockScopedVariables valid 2',
      },
      {
        code: "'use strict'; const a = 0",
        options: [
          {
            version: '4.0.0',
          },
        ],
        name: 'blockScopedVariables valid 3',
      },
      {
        code: "'use strict'; function wrap() { const a = 0 }",
        options: [
          {
            version: '4.0.0',
          },
        ],
        name: 'blockScopedVariables valid 4',
      },
      {
        code: "function wrap() { 'use strict'; const a = 0 }",
        options: [
          {
            version: '4.0.0',
          },
        ],
        name: 'blockScopedVariables valid 5',
      },
      {
        code: 'let a = 0',
        options: [
          {
            version: '6.0.0',
          },
        ],
        name: 'blockScopedVariables valid 6',
      },
      {
        code: 'const a = 0',
        options: [
          {
            version: '6.0.0',
          },
        ],
        name: 'blockScopedVariables valid 7',
      },
      {
        code: "'use strict'; let a = 0",
        options: [
          {
            version: '3.9.9',
            ignores: ['blockScopedVariables'],
          },
        ],
        name: 'blockScopedVariables valid 8',
      },
      {
        code: "'use strict'; const a = 0",
        options: [
          {
            version: '3.9.9',
            ignores: ['blockScopedVariables'],
          },
        ],
        name: 'blockScopedVariables valid 9',
      },
      {
        code: 'let a = 0',
        options: [
          {
            version: '5.9.9',
            ignores: ['blockScopedVariables'],
          },
        ],
        name: 'blockScopedVariables valid 10',
      },
      {
        code: 'const a = 0',
        options: [
          {
            version: '5.9.9',
            ignores: ['blockScopedVariables'],
          },
        ],
        name: 'blockScopedVariables valid 11',
      },
    ],
    invalid: [
      {
        code: "'use strict'; let a = 0",
        options: [
          {
            version: '3.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'block-scoped-variables' is not supported until Node.js >=4.0.0. The configured version range is '3.9.9'.",
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 24,
          },
        ],
        name: 'blockScopedVariables invalid 1',
      },
      {
        code: "'use strict'; const a = 0",
        options: [
          {
            version: '3.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'block-scoped-variables' is not supported until Node.js >=4.0.0. The configured version range is '3.9.9'.",
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 26,
          },
        ],
        name: 'blockScopedVariables invalid 2',
      },
      {
        code: 'let a = 0',
        options: [
          {
            version: '5.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'block-scoped-variables' is not supported until Node.js >=6.0.0. The configured version range is '5.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 10,
          },
        ],
        name: 'blockScopedVariables invalid 3',
      },
      {
        code: 'const a = 0',
        options: [
          {
            version: '5.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'block-scoped-variables' is not supported until Node.js >=6.0.0. The configured version range is '5.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 12,
          },
        ],
        name: 'blockScopedVariables invalid 4',
      },
    ],
  },
);

// classes
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        code: "'use strict'; class A {}",
        options: [
          {
            version: '4.0.0',
          },
        ],
        name: 'classes valid 1',
      },
      {
        code: "'use strict'; function wrap() { class A {} }",
        options: [
          {
            version: '4.0.0',
          },
        ],
        name: 'classes valid 2',
      },
      {
        code: "function wrap() { 'use strict'; class A {} }",
        options: [
          {
            version: '4.0.0',
          },
        ],
        name: 'classes valid 3',
      },
      {
        code: 'class A {}',
        options: [
          {
            version: '6.0.0',
          },
        ],
        name: 'classes valid 4',
      },
      {
        code: "'use strict'; class A {}",
        options: [
          {
            version: '3.9.9',
            ignores: ['classes'],
          },
        ],
        name: 'classes valid 5',
      },
      {
        code: 'class A {}',
        options: [
          {
            version: '5.9.9',
            ignores: ['classes'],
          },
        ],
        name: 'classes valid 6',
      },
    ],
    invalid: [
      {
        code: "'use strict'; class A {}",
        options: [
          {
            version: '3.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'classes' is not supported until Node.js >=4.0.0. The configured version range is '3.9.9'.",
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 25,
          },
        ],
        name: 'classes invalid 1',
      },
      {
        code: 'class A {}',
        options: [
          {
            version: '5.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'classes' is not supported until Node.js >=6.0.0. The configured version range is '5.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 11,
          },
        ],
        name: 'classes invalid 2',
      },
    ],
  },
);

// computedProperties
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        code: "({ 0: 0, key: 1, 'key2': 2, key3, key4() {} })",
        options: [
          {
            version: '3.9.9',
            ignores: ['propertyShorthands'],
          },
        ],
        name: 'computedProperties valid 1',
      },
      {
        code: '({ [key]: 1 })',
        options: [
          {
            version: '4.0.0',
          },
        ],
        name: 'computedProperties valid 2',
      },
      {
        code: '({ [key]() {} })',
        options: [
          {
            version: '4.0.0',
          },
        ],
        name: 'computedProperties valid 3',
      },
      {
        code: 'class A { [key]() {} }',
        options: [
          {
            version: '4.0.0',
            ignores: ['classes'],
          },
        ],
        name: 'computedProperties valid 4',
      },
      {
        code: '(class { [key]() {} })',
        options: [
          {
            version: '4.0.0',
            ignores: ['classes'],
          },
        ],
        name: 'computedProperties valid 5',
      },
      {
        code: '({ [key]: 1 })',
        options: [
          {
            version: '3.9.9',
            ignores: ['computedProperties'],
          },
        ],
        name: 'computedProperties valid 6',
      },
      {
        code: '({ [key]() {} })',
        options: [
          {
            version: '3.9.9',
            ignores: ['propertyShorthands', 'computedProperties'],
          },
        ],
        name: 'computedProperties valid 7',
      },
      {
        code: 'class A { [key]() {} }',
        options: [
          {
            version: '3.9.9',
            ignores: ['classes', 'computedProperties'],
          },
        ],
        name: 'computedProperties valid 8',
      },
      {
        code: '(class { [key]() {} })',
        options: [
          {
            version: '3.9.9',
            ignores: ['classes', 'computedProperties'],
          },
        ],
        name: 'computedProperties valid 9',
      },
    ],
    invalid: [
      {
        code: '({ [key]: 1 })',
        options: [
          {
            version: '3.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'computed-properties' is not supported until Node.js >=4.0.0. The configured version range is '3.9.9'.",
            line: 1,
            column: 4,
            endLine: 1,
            endColumn: 12,
          },
        ],
        name: 'computedProperties invalid 1',
      },
      {
        code: '({ [key]() {} })',
        options: [
          {
            version: '3.9.9',
            ignores: ['propertyShorthands'],
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'computed-properties' is not supported until Node.js >=4.0.0. The configured version range is '3.9.9'.",
            line: 1,
            column: 4,
            endLine: 1,
            endColumn: 14,
          },
        ],
        name: 'computedProperties invalid 2',
      },
      {
        code: 'class A { [key]() {} }',
        options: [
          {
            version: '3.9.9',
            ignores: ['classes'],
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'computed-properties' is not supported until Node.js >=4.0.0. The configured version range is '3.9.9'.",
            line: 1,
            column: 11,
            endLine: 1,
            endColumn: 21,
          },
        ],
        name: 'computedProperties invalid 3',
      },
      {
        code: '(class { [key]() {} })',
        options: [
          {
            version: '3.9.9',
            ignores: ['classes'],
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'computed-properties' is not supported until Node.js >=4.0.0. The configured version range is '3.9.9'.",
            line: 1,
            column: 10,
            endLine: 1,
            endColumn: 20,
          },
        ],
        name: 'computedProperties invalid 4',
      },
    ],
  },
);

// defaultParameters
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        code: 'a = 0',
        options: [
          {
            version: '5.9.9',
          },
        ],
        name: 'defaultParameters valid 1',
      },
      {
        code: 'var a = 0',
        options: [
          {
            version: '5.9.9',
          },
        ],
        name: 'defaultParameters valid 2',
      },
      {
        code: 'var [a = 0] = []',
        options: [
          {
            version: '5.9.9',
            ignores: ['destructuring'],
          },
        ],
        name: 'defaultParameters valid 3',
      },
      {
        code: 'var {a = 0} = {}',
        options: [
          {
            version: '5.9.9',
            ignores: ['destructuring'],
          },
        ],
        name: 'defaultParameters valid 4',
      },
      {
        code: 'function f(a = 0) {}',
        options: [
          {
            version: '6.0.0',
          },
        ],
        name: 'defaultParameters valid 5',
      },
      {
        code: '(function(a = 0) {})',
        options: [
          {
            version: '6.0.0',
          },
        ],
        name: 'defaultParameters valid 6',
      },
      {
        code: '((a = 0) => a)',
        options: [
          {
            version: '6.0.0',
          },
        ],
        name: 'defaultParameters valid 7',
      },
      {
        code: '({ key(a = 0) {} })',
        options: [
          {
            version: '6.0.0',
          },
        ],
        name: 'defaultParameters valid 8',
      },
      {
        code: 'class A { key(a = 0) {} }',
        options: [
          {
            version: '6.0.0',
          },
        ],
        name: 'defaultParameters valid 9',
      },
      {
        code: '(class { key(a = 0) {} })',
        options: [
          {
            version: '6.0.0',
          },
        ],
        name: 'defaultParameters valid 10',
      },
      {
        code: 'function f(a = 0) {}',
        options: [
          {
            version: '5.9.9',
            ignores: ['defaultParameters'],
          },
        ],
        name: 'defaultParameters valid 11',
      },
      {
        code: '(function(a = 0) {})',
        options: [
          {
            version: '5.9.9',
            ignores: ['defaultParameters'],
          },
        ],
        name: 'defaultParameters valid 12',
      },
      {
        code: '((a = 0) => a)',
        options: [
          {
            version: '5.9.9',
            ignores: ['defaultParameters'],
          },
        ],
        name: 'defaultParameters valid 13',
      },
      {
        code: '({ key(a = 0) {} })',
        options: [
          {
            version: '5.9.9',
            ignores: ['defaultParameters'],
          },
        ],
        name: 'defaultParameters valid 14',
      },
      {
        code: 'class A { key(a = 0) {} }',
        options: [
          {
            version: '5.9.9',
            ignores: ['classes', 'defaultParameters'],
          },
        ],
        name: 'defaultParameters valid 15',
      },
      {
        code: '(class { key(a = 0) {} })',
        options: [
          {
            version: '5.9.9',
            ignores: ['classes', 'defaultParameters'],
          },
        ],
        name: 'defaultParameters valid 16',
      },
    ],
    invalid: [
      {
        code: 'function f(a = 0) {}',
        options: [
          {
            version: '5.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'default-parameters' is not supported until Node.js >=6.0.0. The configured version range is '5.9.9'.",
            line: 1,
            column: 12,
            endLine: 1,
            endColumn: 17,
          },
        ],
        name: 'defaultParameters invalid 1',
      },
      {
        code: '(function(a = 0) {})',
        options: [
          {
            version: '5.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'default-parameters' is not supported until Node.js >=6.0.0. The configured version range is '5.9.9'.",
            line: 1,
            column: 11,
            endLine: 1,
            endColumn: 16,
          },
        ],
        name: 'defaultParameters invalid 2',
      },
      {
        code: '((a = 0) => a)',
        options: [
          {
            version: '5.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'default-parameters' is not supported until Node.js >=6.0.0. The configured version range is '5.9.9'.",
            line: 1,
            column: 3,
            endLine: 1,
            endColumn: 8,
          },
        ],
        name: 'defaultParameters invalid 3',
      },
      {
        code: '({ key(a = 0) {} })',
        options: [
          {
            version: '5.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'default-parameters' is not supported until Node.js >=6.0.0. The configured version range is '5.9.9'.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 13,
          },
        ],
        name: 'defaultParameters invalid 4',
      },
      {
        code: 'class A { key(a = 0) {} }',
        options: [
          {
            version: '5.9.9',
            ignores: ['classes'],
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'default-parameters' is not supported until Node.js >=6.0.0. The configured version range is '5.9.9'.",
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 20,
          },
        ],
        name: 'defaultParameters invalid 5',
      },
      {
        code: '(class { key(a = 0) {} })',
        options: [
          {
            version: '5.9.9',
            ignores: ['classes'],
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'default-parameters' is not supported until Node.js >=6.0.0. The configured version range is '5.9.9'.",
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 19,
          },
        ],
        name: 'defaultParameters invalid 6',
      },
    ],
  },
);

// destructuring
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        code: 'function f(a = 0) {}',
        options: [
          {
            version: '5.9.9',
            ignores: ['defaultParameters'],
          },
        ],
        name: 'destructuring valid 1',
      },
      {
        code: '[...a]',
        options: [
          {
            version: '5.9.9',
          },
        ],
        name: 'destructuring valid 2',
      },
      {
        code: 'f(...a)',
        options: [
          {
            version: '5.9.9',
          },
        ],
        name: 'destructuring valid 3',
      },
      {
        code: 'new A(...a)',
        options: [
          {
            version: '5.9.9',
          },
        ],
        name: 'destructuring valid 4',
      },
      {
        code: 'var a = {}',
        options: [
          {
            version: '5.9.9',
          },
        ],
        name: 'destructuring valid 5',
      },
      {
        code: 'var {a} = {}',
        options: [
          {
            version: '6.0.0',
          },
        ],
        name: 'destructuring valid 6',
      },
      {
        code: 'var [a] = {}',
        options: [
          {
            version: '6.0.0',
          },
        ],
        name: 'destructuring valid 7',
      },
      {
        code: 'function f({a}) {}',
        options: [
          {
            version: '6.0.0',
          },
        ],
        name: 'destructuring valid 8',
      },
      {
        code: 'function f([a]) {}',
        options: [
          {
            version: '6.0.0',
          },
        ],
        name: 'destructuring valid 9',
      },
      {
        code: 'var {a} = {}',
        options: [
          {
            version: '5.9.9',
            ignores: ['destructuring'],
          },
        ],
        name: 'destructuring valid 10',
      },
      {
        code: 'var [a] = {}',
        options: [
          {
            version: '5.9.9',
            ignores: ['destructuring'],
          },
        ],
        name: 'destructuring valid 11',
      },
      {
        code: 'function f({a}) {}',
        options: [
          {
            version: '5.9.9',
            ignores: ['destructuring'],
          },
        ],
        name: 'destructuring valid 12',
      },
      {
        code: 'function f([a]) {}',
        options: [
          {
            version: '5.9.9',
            ignores: ['destructuring'],
          },
        ],
        name: 'destructuring valid 13',
      },
      {
        code: 'var {a: {b: [c = 0]}} = {}',
        options: [
          {
            version: '5.9.9',
            ignores: ['destructuring'],
          },
        ],
        name: 'destructuring valid 14',
      },
      {
        code: 'var [{a: [b = 0]}] = {}',
        options: [
          {
            version: '5.9.9',
            ignores: ['destructuring'],
          },
        ],
        name: 'destructuring valid 15',
      },
      {
        code: 'function f({a: {b: [c = 0]}}) {}',
        options: [
          {
            version: '5.9.9',
            ignores: ['destructuring'],
          },
        ],
        name: 'destructuring valid 16',
      },
      {
        code: 'function f([{a: [b = 0]}]) {}',
        options: [
          {
            version: '5.9.9',
            ignores: ['destructuring'],
          },
        ],
        name: 'destructuring valid 17',
      },
    ],
    invalid: [
      {
        code: 'var {a} = {}',
        options: [
          {
            version: '5.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'destructuring' is not supported until Node.js >=6.0.0. The configured version range is '5.9.9'.",
            line: 1,
            column: 5,
            endLine: 1,
            endColumn: 8,
          },
        ],
        name: 'destructuring invalid 1',
      },
      {
        code: 'var [a] = {}',
        options: [
          {
            version: '5.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'destructuring' is not supported until Node.js >=6.0.0. The configured version range is '5.9.9'.",
            line: 1,
            column: 5,
            endLine: 1,
            endColumn: 8,
          },
        ],
        name: 'destructuring invalid 2',
      },
      {
        code: 'function f({a}) {}',
        options: [
          {
            version: '5.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'destructuring' is not supported until Node.js >=6.0.0. The configured version range is '5.9.9'.",
            line: 1,
            column: 12,
            endLine: 1,
            endColumn: 15,
          },
        ],
        name: 'destructuring invalid 3',
      },
      {
        code: 'function f([a]) {}',
        options: [
          {
            version: '5.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'destructuring' is not supported until Node.js >=6.0.0. The configured version range is '5.9.9'.",
            line: 1,
            column: 12,
            endLine: 1,
            endColumn: 15,
          },
        ],
        name: 'destructuring invalid 4',
      },
      {
        code: 'var {a: {b: [c = 0]}} = {}',
        options: [
          {
            version: '5.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'destructuring' is not supported until Node.js >=6.0.0. The configured version range is '5.9.9'.",
            line: 1,
            column: 5,
            endLine: 1,
            endColumn: 22,
          },
        ],
        name: 'destructuring invalid 5',
      },
      {
        code: 'var [{a: [b = 0]}] = {}',
        options: [
          {
            version: '5.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'destructuring' is not supported until Node.js >=6.0.0. The configured version range is '5.9.9'.",
            line: 1,
            column: 5,
            endLine: 1,
            endColumn: 19,
          },
        ],
        name: 'destructuring invalid 6',
      },
      {
        code: 'function f({a: {b: [c = 0]}}) {}',
        options: [
          {
            version: '5.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'destructuring' is not supported until Node.js >=6.0.0. The configured version range is '5.9.9'.",
            line: 1,
            column: 12,
            endLine: 1,
            endColumn: 29,
          },
        ],
        name: 'destructuring invalid 7',
      },
      {
        code: 'function f([{a: [b = 0]}]) {}',
        options: [
          {
            version: '5.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'destructuring' is not supported until Node.js >=6.0.0. The configured version range is '5.9.9'.",
            line: 1,
            column: 12,
            endLine: 1,
            endColumn: 26,
          },
        ],
        name: 'destructuring invalid 8',
      },
    ],
  },
);

// forOfLoops
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        code: 'for (;;);',
        options: [
          {
            version: '0.11.9',
          },
        ],
        name: 'forOfLoops valid 1',
      },
      {
        code: 'for (a in b);',
        options: [
          {
            version: '0.11.9',
          },
        ],
        name: 'forOfLoops valid 2',
      },
      {
        code: 'for (var a in b);',
        options: [
          {
            version: '0.11.9',
          },
        ],
        name: 'forOfLoops valid 3',
      },
      {
        code: 'for (a of b);',
        options: [
          {
            version: '0.12.0',
          },
        ],
        name: 'forOfLoops valid 4',
      },
      {
        code: 'for (var a of b);',
        options: [
          {
            version: '0.12.0',
          },
        ],
        name: 'forOfLoops valid 5',
      },
      {
        code: 'for (a of b);',
        options: [
          {
            version: '0.11.9',
            ignores: ['forOfLoops'],
          },
        ],
        name: 'forOfLoops valid 6',
      },
      {
        code: 'for (var a of b);',
        options: [
          {
            version: '0.11.9',
            ignores: ['forOfLoops'],
          },
        ],
        name: 'forOfLoops valid 7',
      },
      {
        code: 'async function wrap() { for await (var a of b); }',
        options: [
          {
            version: '0.11.9',
            ignores: ['asyncFunctions', 'asyncIteration', 'forOfLoops'],
          },
        ],
        name: 'forOfLoops valid 8',
      },
    ],
    invalid: [
      {
        code: 'for (a of b);',
        options: [
          {
            version: '0.11.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'for-of-loops' is not supported until Node.js >=0.12.0. The configured version range is '0.11.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 14,
          },
        ],
        name: 'forOfLoops invalid 1',
      },
      {
        code: 'for (var a of b);',
        options: [
          {
            version: '0.11.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'for-of-loops' is not supported until Node.js >=0.12.0. The configured version range is '0.11.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 18,
          },
        ],
        name: 'forOfLoops invalid 2',
      },
      {
        code: 'async function wrap() { for await (var a of b); }',
        options: [
          {
            version: '0.11.9',
            ignores: ['asyncFunctions', 'asyncIteration'],
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'for-of-loops' is not supported until Node.js >=0.12.0. The configured version range is '0.11.9'.",
            line: 1,
            column: 25,
            endLine: 1,
            endColumn: 48,
          },
        ],
        name: 'forOfLoops invalid 3',
      },
    ],
  },
);

// generators
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        code: 'function f() {}',
        options: [
          {
            version: '3.9.9',
          },
        ],
        name: 'generators valid 1',
      },
      {
        code: 'async function f() {}',
        options: [
          {
            version: '3.9.9',
            ignores: ['asyncFunctions'],
          },
        ],
        name: 'generators valid 2',
      },
      {
        code: 'function* f() {}',
        options: [
          {
            version: '4.0.0',
          },
        ],
        name: 'generators valid 3',
      },
      {
        code: '(function*() {})',
        options: [
          {
            version: '4.0.0',
          },
        ],
        name: 'generators valid 4',
      },
      {
        code: '({ *f() {} })',
        options: [
          {
            version: '4.0.0',
            ignores: ['propertyShorthands'],
          },
        ],
        name: 'generators valid 5',
      },
      {
        code: 'class A { *f() {} }',
        options: [
          {
            version: '4.0.0',
            ignores: ['classes'],
          },
        ],
        name: 'generators valid 6',
      },
      {
        code: '(class { *f() {} })',
        options: [
          {
            version: '4.0.0',
            ignores: ['classes'],
          },
        ],
        name: 'generators valid 7',
      },
      {
        code: 'function* f() {}',
        options: [
          {
            version: '3.9.9',
            ignores: ['generators'],
          },
        ],
        name: 'generators valid 8',
      },
      {
        code: '(function*() {})',
        options: [
          {
            version: '3.9.9',
            ignores: ['generators'],
          },
        ],
        name: 'generators valid 9',
      },
      {
        code: '({ *f() {} })',
        options: [
          {
            version: '3.9.9',
            ignores: ['propertyShorthands', 'generators'],
          },
        ],
        name: 'generators valid 10',
      },
      {
        code: 'class A { *f() {} }',
        options: [
          {
            version: '3.9.9',
            ignores: ['classes', 'generators'],
          },
        ],
        name: 'generators valid 11',
      },
      {
        code: '(class { *f() {} })',
        options: [
          {
            version: '3.9.9',
            ignores: ['classes', 'generators'],
          },
        ],
        name: 'generators valid 12',
      },
    ],
    invalid: [
      {
        code: 'function* f() {}',
        options: [
          {
            version: '3.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'generators' is not supported until Node.js >=4.0.0. The configured version range is '3.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 17,
          },
        ],
        name: 'generators invalid 1',
      },
      {
        code: '(function*() {})',
        options: [
          {
            version: '3.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'generators' is not supported until Node.js >=4.0.0. The configured version range is '3.9.9'.",
            line: 1,
            column: 2,
            endLine: 1,
            endColumn: 16,
          },
        ],
        name: 'generators invalid 2',
      },
      {
        code: '({ *f() {} })',
        options: [
          {
            version: '3.9.9',
            ignores: ['propertyShorthands'],
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'generators' is not supported until Node.js >=4.0.0. The configured version range is '3.9.9'.",
            line: 1,
            column: 6,
            endLine: 1,
            endColumn: 11,
          },
        ],
        name: 'generators invalid 3',
      },
      {
        code: 'class A { *f() {} }',
        options: [
          {
            version: '3.9.9',
            ignores: ['classes'],
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'generators' is not supported until Node.js >=4.0.0. The configured version range is '3.9.9'.",
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 18,
          },
        ],
        name: 'generators invalid 4',
      },
      {
        code: '(class { *f() {} })',
        options: [
          {
            version: '3.9.9',
            ignores: ['classes'],
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'generators' is not supported until Node.js >=4.0.0. The configured version range is '3.9.9'.",
            line: 1,
            column: 12,
            endLine: 1,
            endColumn: 17,
          },
        ],
        name: 'generators invalid 5',
      },
    ],
  },
);

// modules
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        code: "require('a')",
        options: [
          {
            version: '0.0.0',
          },
        ],
        name: 'modules valid 1',
      },
      {
        code: 'module.exports = {}',
        options: [
          {
            version: '0.0.0',
          },
        ],
        name: 'modules valid 2',
      },
      {
        code: 'exports.a = {}',
        options: [
          {
            version: '0.0.0',
          },
        ],
        name: 'modules valid 3',
      },
      {
        code: "import a from 'a'",
        languageOptions: {
          sourceType: 'module',
        },
        options: [
          {
            version: '13.1.0',
            ignores: ['modules'],
          },
        ],
        name: 'modules valid 4',
      },
      {
        code: 'export default {}',
        languageOptions: {
          sourceType: 'module',
        },
        options: [
          {
            version: '13.1.0',
            ignores: ['modules'],
          },
        ],
        name: 'modules valid 5',
      },
      {
        code: 'export const a = {}',
        languageOptions: {
          sourceType: 'module',
        },
        options: [
          {
            version: '13.1.0',
            ignores: ['modules'],
          },
        ],
        name: 'modules valid 6',
      },
      {
        code: 'export {}',
        languageOptions: {
          sourceType: 'module',
        },
        options: [
          {
            version: '13.1.0',
            ignores: ['modules'],
          },
        ],
        name: 'modules valid 7',
      },
      {
        code: "import a from 'a'",
        languageOptions: {
          sourceType: 'module',
        },
        options: [
          {
            version: '10.0.0',
            ignores: ['modules'],
          },
        ],
        name: 'modules valid 8',
      },
      {
        code: 'export default {}',
        languageOptions: {
          sourceType: 'module',
        },
        options: [
          {
            version: '10.0.0',
            ignores: ['modules'],
          },
        ],
        name: 'modules valid 9',
      },
      {
        code: 'export const a = {}',
        languageOptions: {
          sourceType: 'module',
        },
        options: [
          {
            version: '10.0.0',
            ignores: ['modules'],
          },
        ],
        name: 'modules valid 10',
      },
      {
        code: 'export {}',
        languageOptions: {
          sourceType: 'module',
        },
        options: [
          {
            version: '10.0.0',
            ignores: ['modules'],
          },
        ],
        name: 'modules valid 11',
      },
    ],
    invalid: [
      {
        code: "import a from 'a'",
        languageOptions: {
          sourceType: 'module',
        },
        options: [
          {
            version: '10.0.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'modules' is not supported until Node.js ^12.17.0 || >=13.2.0. The configured version range is '10.0.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 18,
          },
        ],
        name: 'modules invalid 1',
      },
      {
        code: 'export default {}',
        languageOptions: {
          sourceType: 'module',
        },
        options: [
          {
            version: '10.0.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'modules' is not supported until Node.js ^12.17.0 || >=13.2.0. The configured version range is '10.0.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 18,
          },
        ],
        name: 'modules invalid 2',
      },
      {
        code: 'export const a = {}',
        languageOptions: {
          sourceType: 'module',
        },
        options: [
          {
            version: '10.0.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'modules' is not supported until Node.js ^12.17.0 || >=13.2.0. The configured version range is '10.0.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 20,
          },
        ],
        name: 'modules invalid 3',
      },
      {
        code: 'export {}',
        languageOptions: {
          sourceType: 'module',
        },
        options: [
          {
            version: '10.0.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'modules' is not supported until Node.js ^12.17.0 || >=13.2.0. The configured version range is '10.0.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 10,
          },
        ],
        name: 'modules invalid 4',
      },
    ],
  },
);

// newTarget, new.target
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        code: 'new target',
        options: [
          {
            version: '4.9.9',
          },
        ],
        name: 'newTarget, new.target valid 1',
      },
      {
        code: 'class A { constructor() { new.target } }',
        options: [
          {
            version: '5.0.0',
            ignores: ['classes'],
          },
        ],
        name: 'newTarget, new.target valid 2',
      },
      {
        code: 'function A() { new.target }',
        options: [
          {
            version: '5.0.0',
          },
        ],
        name: 'newTarget, new.target valid 3',
      },
      {
        code: 'class A { constructor() { new.target } }',
        options: [
          {
            version: '4.9.9',
            ignores: ['classes', 'newTarget'],
          },
        ],
        name: 'newTarget, new.target valid 4',
      },
      {
        code: 'function A() { new.target }',
        options: [
          {
            version: '4.9.9',
            ignores: ['newTarget'],
          },
        ],
        name: 'newTarget, new.target valid 5',
      },
      {
        code: 'class A { constructor() { new.target } }',
        options: [
          {
            version: '4.9.9',
            ignores: ['classes', 'new.target'],
          },
        ],
        name: 'newTarget, new.target valid 6',
      },
      {
        code: 'function A() { new.target }',
        options: [
          {
            version: '4.9.9',
            ignores: ['new.target'],
          },
        ],
        name: 'newTarget, new.target valid 7',
      },
    ],
    invalid: [
      {
        code: 'class A { constructor() { new.target } }',
        options: [
          {
            version: '4.9.9',
            ignores: ['classes'],
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'new-target' is not supported until Node.js >=5.0.0. The configured version range is '4.9.9'.",
            line: 1,
            column: 27,
            endLine: 1,
            endColumn: 37,
          },
        ],
        name: 'newTarget, new.target invalid 1',
      },
      {
        code: 'function A() { new.target }',
        options: [
          {
            version: '4.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'new-target' is not supported until Node.js >=5.0.0. The configured version range is '4.9.9'.",
            line: 1,
            column: 16,
            endLine: 1,
            endColumn: 26,
          },
        ],
        name: 'newTarget, new.target invalid 2',
      },
    ],
  },
);

// objectSuperProperties
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        code: 'class A { foo() { super.foo } }',
        options: [
          {
            version: '3.9.9',
            ignores: ['classes'],
          },
        ],
        name: 'objectSuperProperties valid 1',
      },
      {
        code: '(class { foo() { super.foo } })',
        options: [
          {
            version: '3.9.9',
            ignores: ['classes'],
          },
        ],
        name: 'objectSuperProperties valid 2',
      },
      {
        code: 'class A extends B { constructor() { super() } }',
        options: [
          {
            version: '3.9.9',
            ignores: ['classes'],
          },
        ],
        name: 'objectSuperProperties valid 3',
      },
      {
        code: '({ foo() { super.foo } })',
        options: [
          {
            version: '4.0.0',
            ignores: ['propertyShorthands'],
          },
        ],
        name: 'objectSuperProperties valid 4',
      },
      {
        code: '({ foo() { super.foo() } })',
        options: [
          {
            version: '4.0.0',
            ignores: ['propertyShorthands'],
          },
        ],
        name: 'objectSuperProperties valid 5',
      },
      {
        code: '({ foo() { super.foo } })',
        options: [
          {
            version: '3.9.9',
            ignores: ['propertyShorthands', 'objectSuperProperties'],
          },
        ],
        name: 'objectSuperProperties valid 6',
      },
      {
        code: '({ foo() { super.foo() } })',
        options: [
          {
            version: '3.9.9',
            ignores: ['propertyShorthands', 'objectSuperProperties'],
          },
        ],
        name: 'objectSuperProperties valid 7',
      },
    ],
    invalid: [
      {
        code: '({ foo() { super.foo } })',
        options: [
          {
            version: '3.9.9',
            ignores: ['propertyShorthands'],
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'object-super-properties' is not supported until Node.js >=4.0.0. The configured version range is '3.9.9'.",
            line: 1,
            column: 12,
            endLine: 1,
            endColumn: 17,
          },
        ],
        name: 'objectSuperProperties invalid 1',
      },
      {
        code: '({ foo() { super.foo() } })',
        options: [
          {
            version: '3.9.9',
            ignores: ['propertyShorthands'],
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'object-super-properties' is not supported until Node.js >=4.0.0. The configured version range is '3.9.9'.",
            line: 1,
            column: 12,
            endLine: 1,
            endColumn: 17,
          },
        ],
        name: 'objectSuperProperties invalid 2',
      },
    ],
  },
);

// octalNumericLiterals
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        code: '0755',
        skip: 'tsgo rejects legacy octal literals (TS1121) before linting',
        options: [
          {
            version: '3.9.9',
          },
        ],
        name: 'octalNumericLiterals valid 1',
      },
      {
        code: '0x755',
        options: [
          {
            version: '3.9.9',
          },
        ],
        name: 'octalNumericLiterals valid 2',
      },
      {
        code: '0X755',
        options: [
          {
            version: '3.9.9',
          },
        ],
        name: 'octalNumericLiterals valid 3',
      },
      {
        code: '0b01',
        options: [
          {
            version: '3.9.9',
            ignores: ['binaryNumericLiterals'],
          },
        ],
        name: 'octalNumericLiterals valid 4',
      },
      {
        code: '0B01',
        options: [
          {
            version: '3.9.9',
            ignores: ['binaryNumericLiterals'],
          },
        ],
        name: 'octalNumericLiterals valid 5',
      },
      {
        code: '0o755',
        options: [
          {
            version: '4.0.0',
          },
        ],
        name: 'octalNumericLiterals valid 6',
      },
      {
        code: '0O755',
        options: [
          {
            version: '4.0.0',
          },
        ],
        name: 'octalNumericLiterals valid 7',
      },
      {
        code: '0o755',
        options: [
          {
            version: '3.9.9',
            ignores: ['octalNumericLiterals'],
          },
        ],
        name: 'octalNumericLiterals valid 8',
      },
      {
        code: '0O755',
        options: [
          {
            version: '3.9.9',
            ignores: ['octalNumericLiterals'],
          },
        ],
        name: 'octalNumericLiterals valid 9',
      },
    ],
    invalid: [
      {
        code: '0o755',
        options: [
          {
            version: '3.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'octal-numeric-literals' is not supported until Node.js >=4.0.0. The configured version range is '3.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 6,
          },
        ],
        name: 'octalNumericLiterals invalid 1',
      },
      {
        code: '0O755',
        options: [
          {
            version: '3.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'octal-numeric-literals' is not supported until Node.js >=4.0.0. The configured version range is '3.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 6,
          },
        ],
        name: 'octalNumericLiterals invalid 2',
      },
    ],
  },
);

// propertyShorthands
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        code: '({ a: 1 })',
        options: [
          {
            version: '3.9.9',
          },
        ],
        name: 'propertyShorthands valid 1',
      },
      {
        code: '({ get: get })',
        options: [
          {
            version: '3.9.9',
          },
        ],
        name: 'propertyShorthands valid 2',
      },
      {
        code: '({ get a() {} })',
        options: [
          {
            version: '3.9.9',
          },
        ],
        name: 'propertyShorthands valid 3',
      },
      {
        code: '({ a })',
        options: [
          {
            version: '4.0.0',
          },
        ],
        name: 'propertyShorthands valid 4',
      },
      {
        code: '({ b() {} })',
        options: [
          {
            version: '4.0.0',
          },
        ],
        name: 'propertyShorthands valid 5',
      },
      {
        code: '({ get() {} })',
        options: [
          {
            version: '4.0.0',
          },
        ],
        name: 'propertyShorthands valid 6',
      },
      {
        code: '({ [c]() {} })',
        options: [
          {
            version: '4.0.0',
            ignores: ['computedProperties'],
          },
        ],
        name: 'propertyShorthands valid 7',
      },
      {
        code: '({ get })',
        options: [
          {
            version: '6.0.0',
          },
        ],
        name: 'propertyShorthands valid 8',
      },
      {
        code: '({ set })',
        options: [
          {
            version: '6.0.0',
          },
        ],
        name: 'propertyShorthands valid 9',
      },
      {
        code: '({ a })',
        options: [
          {
            version: '3.9.9',
            ignores: ['propertyShorthands'],
          },
        ],
        name: 'propertyShorthands valid 10',
      },
      {
        code: '({ b() {} })',
        options: [
          {
            version: '3.9.9',
            ignores: ['propertyShorthands'],
          },
        ],
        name: 'propertyShorthands valid 11',
      },
      {
        code: '({ [c]() {} })',
        options: [
          {
            version: '3.9.9',
            ignores: ['computedProperties', 'propertyShorthands'],
          },
        ],
        name: 'propertyShorthands valid 12',
      },
    ],
    invalid: [
      {
        code: '({ a })',
        options: [
          {
            version: '3.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'property-shorthands' is not supported until Node.js >=4.0.0. The configured version range is '3.9.9'.",
            line: 1,
            column: 4,
            endLine: 1,
            endColumn: 5,
          },
        ],
        name: 'propertyShorthands invalid 1',
      },
      {
        code: '({ b() {} })',
        options: [
          {
            version: '3.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'property-shorthands' is not supported until Node.js >=4.0.0. The configured version range is '3.9.9'.",
            line: 1,
            column: 4,
            endLine: 1,
            endColumn: 10,
          },
        ],
        name: 'propertyShorthands invalid 2',
      },
      {
        code: '({ [c]() {} })',
        options: [
          {
            version: '3.9.9',
            ignores: ['computedProperties'],
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'property-shorthands' is not supported until Node.js >=4.0.0. The configured version range is '3.9.9'.",
            line: 1,
            column: 4,
            endLine: 1,
            endColumn: 12,
          },
        ],
        name: 'propertyShorthands invalid 3',
      },
    ],
  },
);

// regexpU, regexpUFlag
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        code: '/foo/',
        options: [
          {
            version: '5.9.9',
          },
        ],
        name: 'regexpU, regexpUFlag valid 1',
      },
      {
        code: '/foo/gmi',
        options: [
          {
            version: '5.9.9',
          },
        ],
        name: 'regexpU, regexpUFlag valid 2',
      },
      {
        code: '/foo/y',
        options: [
          {
            version: '5.9.9',
            ignores: ['regexpYFlag'],
          },
        ],
        name: 'regexpU, regexpUFlag valid 3',
      },
      {
        code: '/foo/u',
        options: [
          {
            version: '6.0.0',
          },
        ],
        name: 'regexpU, regexpUFlag valid 4',
      },
      {
        code: '/foo/u',
        options: [
          {
            version: '5.9.9',
            ignores: ['regexpU'],
          },
        ],
        name: 'regexpU, regexpUFlag valid 5',
      },
      {
        code: '/foo/u',
        options: [
          {
            version: '5.9.9',
            ignores: ['regexpUFlag'],
          },
        ],
        name: 'regexpU, regexpUFlag valid 6',
      },
    ],
    invalid: [
      {
        code: '/foo/u',
        options: [
          {
            version: '5.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'regexp-u-flag' is not supported until Node.js >=6.0.0. The configured version range is '5.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 7,
          },
        ],
        name: 'regexpU, regexpUFlag invalid 1',
      },
    ],
  },
);

// regexpY, regexpYFlag
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        code: '/foo/',
        options: [
          {
            version: '5.9.9',
          },
        ],
        name: 'regexpY, regexpYFlag valid 1',
      },
      {
        code: '/foo/gmi',
        options: [
          {
            version: '5.9.9',
          },
        ],
        name: 'regexpY, regexpYFlag valid 2',
      },
      {
        code: '/foo/u',
        options: [
          {
            version: '5.9.9',
            ignores: ['regexpUFlag'],
          },
        ],
        name: 'regexpY, regexpYFlag valid 3',
      },
      {
        code: '/foo/y',
        options: [
          {
            version: '6.0.0',
          },
        ],
        name: 'regexpY, regexpYFlag valid 4',
      },
      {
        code: '/foo/y',
        options: [
          {
            version: '5.9.9',
            ignores: ['regexpY'],
          },
        ],
        name: 'regexpY, regexpYFlag valid 5',
      },
      {
        code: '/foo/y',
        options: [
          {
            version: '5.9.9',
            ignores: ['regexpYFlag'],
          },
        ],
        name: 'regexpY, regexpYFlag valid 6',
      },
    ],
    invalid: [
      {
        code: '/foo/y',
        options: [
          {
            version: '5.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'regexp-y-flag' is not supported until Node.js >=6.0.0. The configured version range is '5.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 7,
          },
        ],
        name: 'regexpY, regexpYFlag invalid 1',
      },
    ],
  },
);

// restParameters
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        code: 'var [...a] = b',
        options: [
          {
            version: '5.9.9',
            ignores: ['destructuring'],
          },
        ],
        name: 'restParameters valid 1',
      },
      {
        code: 'var {...a} = b',
        options: [
          {
            version: '5.9.9',
            ignores: ['destructuring', 'restSpreadProperties'],
          },
        ],
        name: 'restParameters valid 2',
      },
      {
        code: 'var a = [...b]',
        options: [
          {
            version: '5.9.9',
          },
        ],
        name: 'restParameters valid 3',
      },
      {
        code: 'var a = {...b}',
        options: [
          {
            version: '5.9.9',
            ignores: ['restSpreadProperties'],
          },
        ],
        name: 'restParameters valid 4',
      },
      {
        code: 'f(...a)',
        options: [
          {
            version: '5.9.9',
          },
        ],
        name: 'restParameters valid 5',
      },
      {
        code: 'function f(...a) {}',
        options: [
          {
            version: '6.0.0',
          },
        ],
        name: 'restParameters valid 6',
      },
      {
        code: 'function f(...[a, b = 0]) {}',
        options: [
          {
            version: '6.0.0',
            ignores: ['destructuring'],
          },
        ],
        name: 'restParameters valid 7',
      },
      {
        code: '(function(...a) {})',
        options: [
          {
            version: '6.0.0',
          },
        ],
        name: 'restParameters valid 8',
      },
      {
        code: '((...a) => {})',
        options: [
          {
            version: '6.0.0',
          },
        ],
        name: 'restParameters valid 9',
      },
      {
        code: '({ f(...a) {} })',
        options: [
          {
            version: '6.0.0',
          },
        ],
        name: 'restParameters valid 10',
      },
      {
        code: 'class A { f(...a) {} }',
        options: [
          {
            version: '6.0.0',
            ignores: ['classes'],
          },
        ],
        name: 'restParameters valid 11',
      },
      {
        code: '(class { f(...a) {} })',
        options: [
          {
            version: '6.0.0',
            ignores: ['classes'],
          },
        ],
        name: 'restParameters valid 12',
      },
      {
        code: 'function f(...a) {}',
        options: [
          {
            version: '5.9.9',
            ignores: ['restParameters'],
          },
        ],
        name: 'restParameters valid 13',
      },
      {
        code: 'function f(...[a, b = 0]) {}',
        options: [
          {
            version: '5.9.9',
            ignores: ['destructuring', 'restParameters'],
          },
        ],
        name: 'restParameters valid 14',
      },
      {
        code: '(function(...a) {})',
        options: [
          {
            version: '5.9.9',
            ignores: ['restParameters'],
          },
        ],
        name: 'restParameters valid 15',
      },
      {
        code: '((...a) => {})',
        options: [
          {
            version: '5.9.9',
            ignores: ['restParameters'],
          },
        ],
        name: 'restParameters valid 16',
      },
      {
        code: '({ f(...a) {} })',
        options: [
          {
            version: '5.9.9',
            ignores: ['restParameters'],
          },
        ],
        name: 'restParameters valid 17',
      },
      {
        code: 'class A { f(...a) {} }',
        options: [
          {
            version: '5.9.9',
            ignores: ['classes', 'restParameters'],
          },
        ],
        name: 'restParameters valid 18',
      },
      {
        code: '(class { f(...a) {} })',
        options: [
          {
            version: '5.9.9',
            ignores: ['classes', 'restParameters'],
          },
        ],
        name: 'restParameters valid 19',
      },
    ],
    invalid: [
      {
        code: 'function f(...a) {}',
        options: [
          {
            version: '5.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'rest-parameters' is not supported until Node.js >=6.0.0. The configured version range is '5.9.9'.",
            line: 1,
            column: 12,
            endLine: 1,
            endColumn: 16,
          },
        ],
        name: 'restParameters invalid 1',
      },
      {
        code: 'function f(...[a, b = 0]) {}',
        options: [
          {
            version: '5.9.9',
            ignores: ['destructuring'],
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'rest-parameters' is not supported until Node.js >=6.0.0. The configured version range is '5.9.9'.",
            line: 1,
            column: 12,
            endLine: 1,
            endColumn: 25,
          },
        ],
        name: 'restParameters invalid 2',
      },
      {
        code: '(function(...a) {})',
        options: [
          {
            version: '5.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'rest-parameters' is not supported until Node.js >=6.0.0. The configured version range is '5.9.9'.",
            line: 1,
            column: 11,
            endLine: 1,
            endColumn: 15,
          },
        ],
        name: 'restParameters invalid 3',
      },
      {
        code: '((...a) => {})',
        options: [
          {
            version: '5.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'rest-parameters' is not supported until Node.js >=6.0.0. The configured version range is '5.9.9'.",
            line: 1,
            column: 3,
            endLine: 1,
            endColumn: 7,
          },
        ],
        name: 'restParameters invalid 4',
      },
      {
        code: '({ f(...a) {} })',
        options: [
          {
            version: '5.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'rest-parameters' is not supported until Node.js >=6.0.0. The configured version range is '5.9.9'.",
            line: 1,
            column: 6,
            endLine: 1,
            endColumn: 10,
          },
        ],
        name: 'restParameters invalid 5',
      },
      {
        code: 'class A { f(...a) {} }',
        options: [
          {
            version: '5.9.9',
            ignores: ['classes'],
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'rest-parameters' is not supported until Node.js >=6.0.0. The configured version range is '5.9.9'.",
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 17,
          },
        ],
        name: 'restParameters invalid 6',
      },
      {
        code: '(class { f(...a) {} })',
        options: [
          {
            version: '5.9.9',
            ignores: ['classes'],
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'rest-parameters' is not supported until Node.js >=6.0.0. The configured version range is '5.9.9'.",
            line: 1,
            column: 12,
            endLine: 1,
            endColumn: 16,
          },
        ],
        name: 'restParameters invalid 7',
      },
    ],
  },
);

// spreadElements
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        code: 'var [...a] = b',
        options: [
          {
            version: '4.9.9',
            ignores: ['destructuring'],
          },
        ],
        name: 'spreadElements valid 1',
      },
      {
        code: 'var {...a} = b',
        options: [
          {
            version: '4.9.9',
            ignores: ['destructuring', 'restSpreadProperties'],
          },
        ],
        name: 'spreadElements valid 2',
      },
      {
        code: 'var a = {...b}',
        options: [
          {
            version: '4.9.9',
            ignores: ['restSpreadProperties'],
          },
        ],
        name: 'spreadElements valid 3',
      },
      {
        code: 'function f(...a) {}',
        options: [
          {
            version: '4.9.9',
            ignores: ['restParameters'],
          },
        ],
        name: 'spreadElements valid 4',
      },
      {
        code: '[...a]',
        options: [
          {
            version: '5.0.0',
          },
        ],
        name: 'spreadElements valid 5',
      },
      {
        code: '[...a, ...b]',
        options: [
          {
            version: '5.0.0',
          },
        ],
        name: 'spreadElements valid 6',
      },
      {
        code: 'f(...a)',
        options: [
          {
            version: '5.0.0',
          },
        ],
        name: 'spreadElements valid 7',
      },
      {
        code: 'new F(...a)',
        options: [
          {
            version: '5.0.0',
          },
        ],
        name: 'spreadElements valid 8',
      },
      {
        code: '[...a]',
        options: [
          {
            version: '4.9.9',
            ignores: ['spreadElements'],
          },
        ],
        name: 'spreadElements valid 9',
      },
      {
        code: '[...a, ...b]',
        options: [
          {
            version: '4.9.9',
            ignores: ['spreadElements'],
          },
        ],
        name: 'spreadElements valid 10',
      },
      {
        code: 'f(...a)',
        options: [
          {
            version: '4.9.9',
            ignores: ['spreadElements'],
          },
        ],
        name: 'spreadElements valid 11',
      },
      {
        code: 'new F(...a)',
        options: [
          {
            version: '4.9.9',
            ignores: ['spreadElements'],
          },
        ],
        name: 'spreadElements valid 12',
      },
    ],
    invalid: [
      {
        code: '[...a]',
        options: [
          {
            version: '4.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'spread-elements' is not supported until Node.js >=5.0.0. The configured version range is '4.9.9'.",
            line: 1,
            column: 2,
            endLine: 1,
            endColumn: 6,
          },
        ],
        name: 'spreadElements invalid 1',
      },
      {
        code: '[...a, ...b]',
        options: [
          {
            version: '4.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'spread-elements' is not supported until Node.js >=5.0.0. The configured version range is '4.9.9'.",
            line: 1,
            column: 2,
            endLine: 1,
            endColumn: 6,
          },
          {
            messageId: 'not-supported-till',
            message:
              "'spread-elements' is not supported until Node.js >=5.0.0. The configured version range is '4.9.9'.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 12,
          },
        ],
        name: 'spreadElements invalid 2',
      },
      {
        code: 'f(...a)',
        options: [
          {
            version: '4.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'spread-elements' is not supported until Node.js >=5.0.0. The configured version range is '4.9.9'.",
            line: 1,
            column: 3,
            endLine: 1,
            endColumn: 7,
          },
        ],
        name: 'spreadElements invalid 3',
      },
      {
        code: 'new F(...a)',
        options: [
          {
            version: '4.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'spread-elements' is not supported until Node.js >=5.0.0. The configured version range is '4.9.9'.",
            line: 1,
            column: 7,
            endLine: 1,
            endColumn: 11,
          },
        ],
        name: 'spreadElements invalid 4',
      },
    ],
  },
);

// templateLiterals
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        code: "'`foo`'",
        options: [
          {
            version: '3.9.9',
          },
        ],
        name: 'templateLiterals valid 1',
      },
      {
        code: '`foo`',
        options: [
          {
            version: '4.0.0',
          },
        ],
        name: 'templateLiterals valid 2',
      },
      {
        code: '`foo${a}bar`',
        options: [
          {
            version: '4.0.0',
          },
        ],
        name: 'templateLiterals valid 3',
      },
      {
        code: 'tag`foo`',
        options: [
          {
            version: '4.0.0',
          },
        ],
        name: 'templateLiterals valid 4',
      },
      {
        code: 'tag`foo${a}bar`',
        options: [
          {
            version: '4.0.0',
          },
        ],
        name: 'templateLiterals valid 5',
      },
      {
        code: '`foo`',
        options: [
          {
            version: '3.9.9',
            ignores: ['templateLiterals'],
          },
        ],
        name: 'templateLiterals valid 6',
      },
      {
        code: '`foo${a}bar`',
        options: [
          {
            version: '3.9.9',
            ignores: ['templateLiterals'],
          },
        ],
        name: 'templateLiterals valid 7',
      },
      {
        code: 'tag`foo`',
        options: [
          {
            version: '3.9.9',
            ignores: ['templateLiterals'],
          },
        ],
        name: 'templateLiterals valid 8',
      },
      {
        code: 'tag`foo${a}bar`',
        options: [
          {
            version: '3.9.9',
            ignores: ['templateLiterals'],
          },
        ],
        name: 'templateLiterals valid 9',
      },
    ],
    invalid: [
      {
        code: '`foo`',
        options: [
          {
            version: '3.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'template-literals' is not supported until Node.js >=4.0.0. The configured version range is '3.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 6,
          },
        ],
        name: 'templateLiterals invalid 1',
      },
      {
        code: '`foo${a}bar`',
        options: [
          {
            version: '3.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'template-literals' is not supported until Node.js >=4.0.0. The configured version range is '3.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 13,
          },
        ],
        name: 'templateLiterals invalid 2',
      },
      {
        code: 'tag`foo`',
        options: [
          {
            version: '3.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'template-literals' is not supported until Node.js >=4.0.0. The configured version range is '3.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 9,
          },
        ],
        name: 'templateLiterals invalid 3',
      },
      {
        code: 'tag`foo${a}bar`',
        options: [
          {
            version: '3.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'template-literals' is not supported until Node.js >=4.0.0. The configured version range is '3.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 16,
          },
        ],
        name: 'templateLiterals invalid 4',
      },
    ],
  },
);

// unicodeCodePointEscapes, unicodeCodepointEscapes
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        code: 'var a = "\\x61"',
        options: [
          {
            version: '3.9.9',
          },
        ],
        name: 'unicodeCodePointEscapes, unicodeCodepointEscapes valid 1',
      },
      {
        code: 'var a = "\\u0061"',
        options: [
          {
            version: '3.9.9',
          },
        ],
        name: 'unicodeCodePointEscapes, unicodeCodepointEscapes valid 2',
      },
      {
        code: 'var a = "\\\\u{61}"',
        options: [
          {
            version: '3.9.9',
          },
        ],
        name: 'unicodeCodePointEscapes, unicodeCodepointEscapes valid 3',
      },
      {
        code: 'var \\u{61} = 0',
        options: [
          {
            version: '4.0.0',
          },
        ],
        name: 'unicodeCodePointEscapes, unicodeCodepointEscapes valid 4',
      },
      {
        code: 'var a = "\\u{61}"',
        options: [
          {
            version: '4.0.0',
          },
        ],
        name: 'unicodeCodePointEscapes, unicodeCodepointEscapes valid 5',
      },
      {
        code: "var a = '\\u{61}'",
        options: [
          {
            version: '4.0.0',
          },
        ],
        name: 'unicodeCodePointEscapes, unicodeCodepointEscapes valid 6',
      },
      {
        code: 'var a = `\\u{61}`',
        options: [
          {
            version: '4.0.0',
          },
        ],
        name: 'unicodeCodePointEscapes, unicodeCodepointEscapes valid 7',
      },
      {
        code: 'var \\u{61} = 0',
        options: [
          {
            version: '3.9.9',
            ignores: ['unicodeCodePointEscapes'],
          },
        ],
        name: 'unicodeCodePointEscapes, unicodeCodepointEscapes valid 8',
      },
      {
        code: 'var a = "\\u{61}"',
        options: [
          {
            version: '3.9.9',
            ignores: ['unicodeCodePointEscapes'],
          },
        ],
        name: 'unicodeCodePointEscapes, unicodeCodepointEscapes valid 9',
      },
      {
        code: 'var a = "\\\\\\u{61}"',
        options: [
          {
            version: '3.9.9',
            ignores: ['unicodeCodePointEscapes'],
          },
        ],
        name: 'unicodeCodePointEscapes, unicodeCodepointEscapes valid 10',
      },
      {
        code: "var a = '\\u{61}'",
        options: [
          {
            version: '3.9.9',
            ignores: ['unicodeCodePointEscapes'],
          },
        ],
        name: 'unicodeCodePointEscapes, unicodeCodepointEscapes valid 11',
      },
      {
        code: 'var a = `\\u{61}`',
        options: [
          {
            version: '3.9.9',
            ignores: ['templateLiterals', 'unicodeCodePointEscapes'],
          },
        ],
        name: 'unicodeCodePointEscapes, unicodeCodepointEscapes valid 12',
      },
      {
        code: 'var \\u{61} = 0',
        options: [
          {
            version: '3.9.9',
            ignores: ['unicodeCodepointEscapes'],
          },
        ],
        name: 'unicodeCodePointEscapes, unicodeCodepointEscapes valid 13',
      },
      {
        code: 'var a = "\\u{61}"',
        options: [
          {
            version: '3.9.9',
            ignores: ['unicodeCodepointEscapes'],
          },
        ],
        name: 'unicodeCodePointEscapes, unicodeCodepointEscapes valid 14',
      },
      {
        code: 'var a = "\\\\\\u{61}"',
        options: [
          {
            version: '3.9.9',
            ignores: ['unicodeCodepointEscapes'],
          },
        ],
        name: 'unicodeCodePointEscapes, unicodeCodepointEscapes valid 15',
      },
      {
        code: "var a = '\\u{61}'",
        options: [
          {
            version: '3.9.9',
            ignores: ['unicodeCodepointEscapes'],
          },
        ],
        name: 'unicodeCodePointEscapes, unicodeCodepointEscapes valid 16',
      },
      {
        code: 'var a = `\\u{61}`',
        options: [
          {
            version: '3.9.9',
            ignores: ['templateLiterals', 'unicodeCodepointEscapes'],
          },
        ],
        name: 'unicodeCodePointEscapes, unicodeCodepointEscapes valid 17',
      },
    ],
    invalid: [
      {
        code: 'var \\u{61} = 0',
        options: [
          {
            version: '3.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'unicode-codepoint-escapes' is not supported until Node.js >=4.0.0. The configured version range is '3.9.9'.",
            line: 1,
            column: 5,
            endLine: 1,
            endColumn: 11,
          },
        ],
        name: 'unicodeCodePointEscapes, unicodeCodepointEscapes invalid 1',
      },
      {
        code: 'var a = "\\u{61}"',
        options: [
          {
            version: '3.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'unicode-codepoint-escapes' is not supported until Node.js >=4.0.0. The configured version range is '3.9.9'.",
            line: 1,
            column: 10,
            endLine: 1,
            endColumn: 16,
          },
        ],
        name: 'unicodeCodePointEscapes, unicodeCodepointEscapes invalid 2',
      },
      {
        code: 'var a = "\\\\\\u{61}"',
        options: [
          {
            version: '3.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'unicode-codepoint-escapes' is not supported until Node.js >=4.0.0. The configured version range is '3.9.9'.",
            line: 1,
            column: 12,
            endLine: 1,
            endColumn: 18,
          },
        ],
        name: 'unicodeCodePointEscapes, unicodeCodepointEscapes invalid 3',
      },
      {
        code: "var a = '\\u{61}'",
        options: [
          {
            version: '3.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'unicode-codepoint-escapes' is not supported until Node.js >=4.0.0. The configured version range is '3.9.9'.",
            line: 1,
            column: 10,
            endLine: 1,
            endColumn: 16,
          },
        ],
        name: 'unicodeCodePointEscapes, unicodeCodepointEscapes invalid 4',
      },
      {
        code: 'var a = `\\u{61}`',
        options: [
          {
            version: '3.9.9',
            ignores: ['templateLiterals'],
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'unicode-codepoint-escapes' is not supported until Node.js >=4.0.0. The configured version range is '3.9.9'.",
            line: 1,
            column: 10,
            endLine: 1,
            endColumn: 16,
          },
        ],
        name: 'unicodeCodePointEscapes, unicodeCodepointEscapes invalid 5',
      },
    ],
  },
);

// exponentialOperators
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        code: 'a ** b',
        options: [
          {
            version: '7.0.0',
          },
        ],
        name: 'exponentialOperators valid 1',
      },
      {
        code: 'a **= b',
        options: [
          {
            version: '7.0.0',
          },
        ],
        name: 'exponentialOperators valid 2',
      },
      {
        code: 'a * b',
        options: [
          {
            version: '6.9.9',
          },
        ],
        name: 'exponentialOperators valid 3',
      },
      {
        code: 'a *= b',
        options: [
          {
            version: '6.9.9',
          },
        ],
        name: 'exponentialOperators valid 4',
      },
      {
        code: 'a ** b',
        options: [
          {
            version: '6.9.9',
            ignores: ['exponentialOperators'],
          },
        ],
        name: 'exponentialOperators valid 5',
      },
      {
        code: 'a **= b',
        options: [
          {
            version: '6.9.9',
            ignores: ['exponentialOperators'],
          },
        ],
        name: 'exponentialOperators valid 6',
      },
    ],
    invalid: [
      {
        code: 'a ** b',
        options: [
          {
            version: '6.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'exponential-operators' is not supported until Node.js >=7.0.0. The configured version range is '6.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 7,
          },
        ],
        name: 'exponentialOperators invalid 1',
      },
      {
        code: 'a **= b',
        options: [
          {
            version: '6.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'exponential-operators' is not supported until Node.js >=7.0.0. The configured version range is '6.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 8,
          },
        ],
        name: 'exponentialOperators invalid 2',
      },
    ],
  },
);

// asyncFunctions
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        code: 'async function f() {}',
        options: [
          {
            version: '7.6.0',
          },
        ],
        name: 'asyncFunctions valid 1',
      },
      {
        code: 'async function f() { await 1 }',
        options: [
          {
            version: '7.6.0',
          },
        ],
        name: 'asyncFunctions valid 2',
      },
      {
        code: '(async function() { await 1 })',
        options: [
          {
            version: '7.6.0',
          },
        ],
        name: 'asyncFunctions valid 3',
      },
      {
        code: '(async() => { await 1 })',
        options: [
          {
            version: '7.6.0',
          },
        ],
        name: 'asyncFunctions valid 4',
      },
      {
        code: '({ async method() { await 1 } })',
        options: [
          {
            version: '7.6.0',
          },
        ],
        name: 'asyncFunctions valid 5',
      },
      {
        code: 'class A { async method() { await 1 } }',
        options: [
          {
            version: '7.6.0',
          },
        ],
        name: 'asyncFunctions valid 6',
      },
      {
        code: '(class { async method() { await 1 } })',
        options: [
          {
            version: '7.6.0',
          },
        ],
        name: 'asyncFunctions valid 7',
      },
      {
        code: 'async function f() {}',
        options: [
          {
            version: '7.5.9',
            ignores: ['asyncFunctions'],
          },
        ],
        name: 'asyncFunctions valid 8',
      },
      {
        code: 'async function f() { await 1 }',
        options: [
          {
            version: '7.5.9',
            ignores: ['asyncFunctions'],
          },
        ],
        name: 'asyncFunctions valid 9',
      },
      {
        code: '(async function() { await 1 })',
        options: [
          {
            version: '7.5.9',
            ignores: ['asyncFunctions'],
          },
        ],
        name: 'asyncFunctions valid 10',
      },
      {
        code: '(async() => { await 1 })',
        options: [
          {
            version: '7.5.9',
            ignores: ['asyncFunctions'],
          },
        ],
        name: 'asyncFunctions valid 11',
      },
      {
        code: '({ async method() { await 1 } })',
        options: [
          {
            version: '7.5.9',
            ignores: ['asyncFunctions'],
          },
        ],
        name: 'asyncFunctions valid 12',
      },
      {
        code: 'class A { async method() { await 1 } }',
        options: [
          {
            version: '7.5.9',
            ignores: ['asyncFunctions'],
          },
        ],
        name: 'asyncFunctions valid 13',
      },
      {
        code: '(class { async method() { await 1 } })',
        options: [
          {
            version: '7.5.9',
            ignores: ['asyncFunctions'],
          },
        ],
        name: 'asyncFunctions valid 14',
      },
    ],
    invalid: [
      {
        code: 'async function f() {}',
        options: [
          {
            version: '7.5.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'async-functions' is not supported until Node.js >=7.6.0. The configured version range is '7.5.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 22,
          },
        ],
        name: 'asyncFunctions invalid 1',
      },
      {
        code: 'async function f() { await 1 }',
        options: [
          {
            version: '7.5.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'async-functions' is not supported until Node.js >=7.6.0. The configured version range is '7.5.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 31,
          },
        ],
        name: 'asyncFunctions invalid 2',
      },
      {
        code: '(async function() { await 1 })',
        options: [
          {
            version: '7.5.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'async-functions' is not supported until Node.js >=7.6.0. The configured version range is '7.5.9'.",
            line: 1,
            column: 2,
            endLine: 1,
            endColumn: 30,
          },
        ],
        name: 'asyncFunctions invalid 3',
      },
      {
        code: '(async() => { await 1 })',
        options: [
          {
            version: '7.5.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'async-functions' is not supported until Node.js >=7.6.0. The configured version range is '7.5.9'.",
            line: 1,
            column: 2,
            endLine: 1,
            endColumn: 24,
          },
        ],
        name: 'asyncFunctions invalid 4',
      },
      {
        code: '({ async method() { await 1 } })',
        options: [
          {
            version: '7.5.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'async-functions' is not supported until Node.js >=7.6.0. The configured version range is '7.5.9'.",
            line: 1,
            column: 16,
            endLine: 1,
            endColumn: 30,
          },
        ],
        name: 'asyncFunctions invalid 5',
      },
      {
        code: 'class A { async method() { await 1 } }',
        options: [
          {
            version: '7.5.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'async-functions' is not supported until Node.js >=7.6.0. The configured version range is '7.5.9'.",
            line: 1,
            column: 23,
            endLine: 1,
            endColumn: 37,
          },
        ],
        name: 'asyncFunctions invalid 6',
      },
      {
        code: '(class { async method() { await 1 } })',
        options: [
          {
            version: '7.5.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'async-functions' is not supported until Node.js >=7.6.0. The configured version range is '7.5.9'.",
            line: 1,
            column: 22,
            endLine: 1,
            endColumn: 36,
          },
        ],
        name: 'asyncFunctions invalid 7',
      },
    ],
  },
);

// trailingCommasInFunctions, trailingFunctionCommas
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        code: 'function f(a,) {}',
        options: [
          {
            version: '8.0.0',
          },
        ],
        name: 'trailingCommasInFunctions, trailingFunctionCommas valid 1',
      },
      {
        code: '(function(a,) {})',
        options: [
          {
            version: '8.0.0',
          },
        ],
        name: 'trailingCommasInFunctions, trailingFunctionCommas valid 2',
      },
      {
        code: '((a,) => {})',
        options: [
          {
            version: '8.0.0',
          },
        ],
        name: 'trailingCommasInFunctions, trailingFunctionCommas valid 3',
      },
      {
        code: '({ method(a,) {} })',
        options: [
          {
            version: '8.0.0',
          },
        ],
        name: 'trailingCommasInFunctions, trailingFunctionCommas valid 4',
      },
      {
        code: 'class A { method(a,) {} }',
        options: [
          {
            version: '8.0.0',
          },
        ],
        name: 'trailingCommasInFunctions, trailingFunctionCommas valid 5',
      },
      {
        code: '(class { method(a,) {} })',
        options: [
          {
            version: '8.0.0',
          },
        ],
        name: 'trailingCommasInFunctions, trailingFunctionCommas valid 6',
      },
      {
        code: 'f(1,)',
        options: [
          {
            version: '8.0.0',
          },
        ],
        name: 'trailingCommasInFunctions, trailingFunctionCommas valid 7',
      },
      {
        code: 'new A(1,)',
        options: [
          {
            version: '8.0.0',
          },
        ],
        name: 'trailingCommasInFunctions, trailingFunctionCommas valid 8',
      },
      {
        code: 'function f(a,) {}',
        options: [
          {
            version: '7.9.9',
            ignores: ['trailingCommasInFunctions'],
          },
        ],
        name: 'trailingCommasInFunctions, trailingFunctionCommas valid 9',
      },
      {
        code: '(function(a,) {})',
        options: [
          {
            version: '7.9.9',
            ignores: ['trailingCommasInFunctions'],
          },
        ],
        name: 'trailingCommasInFunctions, trailingFunctionCommas valid 10',
      },
      {
        code: '((a,) => {})',
        options: [
          {
            version: '7.9.9',
            ignores: ['trailingCommasInFunctions'],
          },
        ],
        name: 'trailingCommasInFunctions, trailingFunctionCommas valid 11',
      },
      {
        code: '({ method(a,) {} })',
        options: [
          {
            version: '7.9.9',
            ignores: ['trailingCommasInFunctions'],
          },
        ],
        name: 'trailingCommasInFunctions, trailingFunctionCommas valid 12',
      },
      {
        code: 'class A { method(a,) {} }',
        options: [
          {
            version: '7.9.9',
            ignores: ['trailingCommasInFunctions'],
          },
        ],
        name: 'trailingCommasInFunctions, trailingFunctionCommas valid 13',
      },
      {
        code: '(class { method(a,) {} })',
        options: [
          {
            version: '7.9.9',
            ignores: ['trailingCommasInFunctions'],
          },
        ],
        name: 'trailingCommasInFunctions, trailingFunctionCommas valid 14',
      },
      {
        code: 'f(1,)',
        options: [
          {
            version: '7.9.9',
            ignores: ['trailingCommasInFunctions'],
          },
        ],
        name: 'trailingCommasInFunctions, trailingFunctionCommas valid 15',
      },
      {
        code: 'new A(1,)',
        options: [
          {
            version: '7.9.9',
            ignores: ['trailingCommasInFunctions'],
          },
        ],
        name: 'trailingCommasInFunctions, trailingFunctionCommas valid 16',
      },
      {
        code: 'function f(a,) {}',
        options: [
          {
            version: '7.9.9',
            ignores: ['trailingFunctionCommas'],
          },
        ],
        name: 'trailingCommasInFunctions, trailingFunctionCommas valid 17',
      },
      {
        code: '(function(a,) {})',
        options: [
          {
            version: '7.9.9',
            ignores: ['trailingFunctionCommas'],
          },
        ],
        name: 'trailingCommasInFunctions, trailingFunctionCommas valid 18',
      },
      {
        code: '((a,) => {})',
        options: [
          {
            version: '7.9.9',
            ignores: ['trailingFunctionCommas'],
          },
        ],
        name: 'trailingCommasInFunctions, trailingFunctionCommas valid 19',
      },
      {
        code: '({ method(a,) {} })',
        options: [
          {
            version: '7.9.9',
            ignores: ['trailingFunctionCommas'],
          },
        ],
        name: 'trailingCommasInFunctions, trailingFunctionCommas valid 20',
      },
      {
        code: 'class A { method(a,) {} }',
        options: [
          {
            version: '7.9.9',
            ignores: ['trailingFunctionCommas'],
          },
        ],
        name: 'trailingCommasInFunctions, trailingFunctionCommas valid 21',
      },
      {
        code: '(class { method(a,) {} })',
        options: [
          {
            version: '7.9.9',
            ignores: ['trailingFunctionCommas'],
          },
        ],
        name: 'trailingCommasInFunctions, trailingFunctionCommas valid 22',
      },
      {
        code: 'f(1,)',
        options: [
          {
            version: '7.9.9',
            ignores: ['trailingFunctionCommas'],
          },
        ],
        name: 'trailingCommasInFunctions, trailingFunctionCommas valid 23',
      },
      {
        code: 'new A(1,)',
        options: [
          {
            version: '7.9.9',
            ignores: ['trailingFunctionCommas'],
          },
        ],
        name: 'trailingCommasInFunctions, trailingFunctionCommas valid 24',
      },
    ],
    invalid: [
      {
        code: 'function f(a,) {}',
        options: [
          {
            version: '7.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'trailing-function-commas' is not supported until Node.js >=8.0.0. The configured version range is '7.9.9'.",
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 14,
          },
        ],
        name: 'trailingCommasInFunctions, trailingFunctionCommas invalid 1',
      },
      {
        code: '(function(a,) {})',
        options: [
          {
            version: '7.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'trailing-function-commas' is not supported until Node.js >=8.0.0. The configured version range is '7.9.9'.",
            line: 1,
            column: 12,
            endLine: 1,
            endColumn: 13,
          },
        ],
        name: 'trailingCommasInFunctions, trailingFunctionCommas invalid 2',
      },
      {
        code: '((a,) => {})',
        options: [
          {
            version: '7.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'trailing-function-commas' is not supported until Node.js >=8.0.0. The configured version range is '7.9.9'.",
            line: 1,
            column: 4,
            endLine: 1,
            endColumn: 5,
          },
        ],
        name: 'trailingCommasInFunctions, trailingFunctionCommas invalid 3',
      },
      {
        code: '({ method(a,) {} })',
        options: [
          {
            version: '7.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'trailing-function-commas' is not supported until Node.js >=8.0.0. The configured version range is '7.9.9'.",
            line: 1,
            column: 12,
            endLine: 1,
            endColumn: 13,
          },
        ],
        name: 'trailingCommasInFunctions, trailingFunctionCommas invalid 4',
      },
      {
        code: 'class A { method(a,) {} }',
        options: [
          {
            version: '7.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'trailing-function-commas' is not supported until Node.js >=8.0.0. The configured version range is '7.9.9'.",
            line: 1,
            column: 19,
            endLine: 1,
            endColumn: 20,
          },
        ],
        name: 'trailingCommasInFunctions, trailingFunctionCommas invalid 5',
      },
      {
        code: '(class { method(a,) {} })',
        options: [
          {
            version: '7.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'trailing-function-commas' is not supported until Node.js >=8.0.0. The configured version range is '7.9.9'.",
            line: 1,
            column: 18,
            endLine: 1,
            endColumn: 19,
          },
        ],
        name: 'trailingCommasInFunctions, trailingFunctionCommas invalid 6',
      },
      {
        code: 'f(1,)',
        options: [
          {
            version: '7.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'trailing-function-commas' is not supported until Node.js >=8.0.0. The configured version range is '7.9.9'.",
            line: 1,
            column: 4,
            endLine: 1,
            endColumn: 5,
          },
        ],
        name: 'trailingCommasInFunctions, trailingFunctionCommas invalid 7',
      },
      {
        code: 'new A(1,)',
        options: [
          {
            version: '7.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'trailing-function-commas' is not supported until Node.js >=8.0.0. The configured version range is '7.9.9'.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 9,
          },
        ],
        name: 'trailingCommasInFunctions, trailingFunctionCommas invalid 8',
      },
    ],
  },
);

// asyncIteration
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        code: 'async function f() { for await (const x of xs) {} }',
        options: [
          {
            version: '10.0.0',
          },
        ],
        name: 'asyncIteration valid 1',
      },
      {
        code: 'async function* f() { }',
        options: [
          {
            version: '10.0.0',
          },
        ],
        name: 'asyncIteration valid 2',
      },
      {
        code: '(async function* () { })',
        options: [
          {
            version: '10.0.0',
          },
        ],
        name: 'asyncIteration valid 3',
      },
      {
        code: '({ async* method() { } })',
        options: [
          {
            version: '10.0.0',
          },
        ],
        name: 'asyncIteration valid 4',
      },
      {
        code: 'class A { async* method() { } }',
        options: [
          {
            version: '10.0.0',
          },
        ],
        name: 'asyncIteration valid 5',
      },
      {
        code: '(class { async* method() { } })',
        options: [
          {
            version: '10.0.0',
          },
        ],
        name: 'asyncIteration valid 6',
      },
      {
        code: 'function f() { for (const x of xs) {} }',
        options: [
          {
            version: '9.9.9',
          },
        ],
        name: 'asyncIteration valid 7',
      },
      {
        code: 'async function f() { }',
        options: [
          {
            version: '9.9.9',
          },
        ],
        name: 'asyncIteration valid 8',
      },
      {
        code: 'function* f() { }',
        options: [
          {
            version: '9.9.9',
          },
        ],
        name: 'asyncIteration valid 9',
      },
      {
        code: 'async function f() { for await (const x of xs) {} }',
        options: [
          {
            version: '9.9.9',
            ignores: ['asyncIteration'],
          },
        ],
        name: 'asyncIteration valid 10',
      },
      {
        code: 'async function* f() { }',
        options: [
          {
            version: '9.9.9',
            ignores: ['asyncIteration'],
          },
        ],
        name: 'asyncIteration valid 11',
      },
      {
        code: '(async function* () { })',
        options: [
          {
            version: '9.9.9',
            ignores: ['asyncIteration'],
          },
        ],
        name: 'asyncIteration valid 12',
      },
      {
        code: '({ async* method() { } })',
        options: [
          {
            version: '9.9.9',
            ignores: ['asyncIteration'],
          },
        ],
        name: 'asyncIteration valid 13',
      },
      {
        code: 'class A { async* method() { } }',
        options: [
          {
            version: '9.9.9',
            ignores: ['asyncIteration'],
          },
        ],
        name: 'asyncIteration valid 14',
      },
      {
        code: '(class { async* method() { } })',
        options: [
          {
            version: '9.9.9',
            ignores: ['asyncIteration'],
          },
        ],
        name: 'asyncIteration valid 15',
      },
    ],
    invalid: [
      {
        code: 'async function f() { for await (const x of xs) {} }',
        options: [
          {
            version: '9.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'async-iteration' is not supported until Node.js >=10.0.0. The configured version range is '9.9.9'.",
            line: 1,
            column: 22,
            endLine: 1,
            endColumn: 50,
          },
        ],
        name: 'asyncIteration invalid 1',
      },
      {
        code: 'async function* f() { }',
        options: [
          {
            version: '9.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'async-iteration' is not supported until Node.js >=10.0.0. The configured version range is '9.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 24,
          },
        ],
        name: 'asyncIteration invalid 2',
      },
      {
        code: '(async function* () { })',
        options: [
          {
            version: '9.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'async-iteration' is not supported until Node.js >=10.0.0. The configured version range is '9.9.9'.",
            line: 1,
            column: 2,
            endLine: 1,
            endColumn: 24,
          },
        ],
        name: 'asyncIteration invalid 3',
      },
      {
        code: '({ async* method() { } })',
        options: [
          {
            version: '9.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'async-iteration' is not supported until Node.js >=10.0.0. The configured version range is '9.9.9'.",
            line: 1,
            column: 17,
            endLine: 1,
            endColumn: 23,
          },
        ],
        name: 'asyncIteration invalid 4',
      },
      {
        code: 'class A { async* method() { } }',
        options: [
          {
            version: '9.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'async-iteration' is not supported until Node.js >=10.0.0. The configured version range is '9.9.9'.",
            line: 1,
            column: 24,
            endLine: 1,
            endColumn: 30,
          },
        ],
        name: 'asyncIteration invalid 5',
      },
      {
        code: '(class { async* method() { } })',
        options: [
          {
            version: '9.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'async-iteration' is not supported until Node.js >=10.0.0. The configured version range is '9.9.9'.",
            line: 1,
            column: 23,
            endLine: 1,
            endColumn: 29,
          },
        ],
        name: 'asyncIteration invalid 6',
      },
    ],
  },
);

// malformedTemplateLiterals
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        code: 'tag`\\unicode`',
        options: [
          {
            version: '8.10.0',
          },
        ],
        name: 'malformedTemplateLiterals valid 1',
      },
      {
        code: 'tag`\\unicode`',
        options: [
          {
            version: '8.9.9',
            ignores: ['malformedTemplateLiterals'],
          },
        ],
        name: 'malformedTemplateLiterals valid 2',
      },
    ],
    invalid: [
      {
        code: 'tag`\\unicode`',
        options: [
          {
            version: '8.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'malformed-template-literals' is not supported until Node.js >=8.10.0. The configured version range is '8.9.9'.",
            line: 1,
            column: 4,
            endLine: 1,
            endColumn: 14,
          },
        ],
        name: 'malformedTemplateLiterals invalid 1',
      },
    ],
  },
);

// regexpLookbehind, regexpLookbehindAssertions
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        code: 'var a = /(?<=a)foo/',
        options: [
          {
            version: '8.10.0',
          },
        ],
        name: 'regexpLookbehind, regexpLookbehindAssertions valid 1',
      },
      {
        code: 'var a = /(?<!a)foo/',
        options: [
          {
            version: '8.10.0',
          },
        ],
        name: 'regexpLookbehind, regexpLookbehindAssertions valid 2',
      },
      {
        code: 'var a = new RegExp("/(?<=a)foo/")',
        options: [
          {
            version: '8.10.0',
          },
        ],
        name: 'regexpLookbehind, regexpLookbehindAssertions valid 3',
      },
      {
        code: 'var a = new RegExp(pattern)',
        options: [
          {
            version: '8.9.9',
          },
        ],
        name: 'regexpLookbehind, regexpLookbehindAssertions valid 4',
      },
      {
        code: 'var a = new RegExp("(?<=")',
        options: [
          {
            version: '8.9.9',
          },
        ],
        name: 'regexpLookbehind, regexpLookbehindAssertions valid 5',
      },
      {
        code: 'var a = /\\(?<=a\\)foo/',
        options: [
          {
            version: '8.9.9',
          },
        ],
        name: 'regexpLookbehind, regexpLookbehindAssertions valid 6',
      },
      {
        code: 'var a = /(?<=a)foo/',
        options: [
          {
            version: '8.9.9',
            ignores: ['regexpLookbehind'],
          },
        ],
        name: 'regexpLookbehind, regexpLookbehindAssertions valid 7',
      },
      {
        code: 'var a = /(?<!a)foo/',
        options: [
          {
            version: '8.9.9',
            ignores: ['regexpLookbehind'],
          },
        ],
        name: 'regexpLookbehind, regexpLookbehindAssertions valid 8',
      },
      {
        code: 'var a = new RegExp("/(?<=a)foo/")',
        options: [
          {
            version: '8.9.9',
            ignores: ['regexpLookbehind'],
          },
        ],
        name: 'regexpLookbehind, regexpLookbehindAssertions valid 9',
      },
      {
        code: 'var a = /(?<=a)foo/',
        options: [
          {
            version: '8.9.9',
            ignores: ['regexpLookbehindAssertions'],
          },
        ],
        name: 'regexpLookbehind, regexpLookbehindAssertions valid 10',
      },
      {
        code: 'var a = /(?<!a)foo/',
        options: [
          {
            version: '8.9.9',
            ignores: ['regexpLookbehindAssertions'],
          },
        ],
        name: 'regexpLookbehind, regexpLookbehindAssertions valid 11',
      },
      {
        code: 'var a = new RegExp("/(?<=a)foo/")',
        options: [
          {
            version: '8.9.9',
            ignores: ['regexpLookbehindAssertions'],
          },
        ],
        name: 'regexpLookbehind, regexpLookbehindAssertions valid 12',
      },
    ],
    invalid: [
      {
        code: 'var a = /(?<=a)foo/',
        options: [
          {
            version: '8.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'regexp-lookbehind-assertions' is not supported until Node.js >=8.10.0. The configured version range is '8.9.9'.",
            line: 1,
            column: 9,
            endLine: 1,
            endColumn: 20,
          },
        ],
        name: 'regexpLookbehind, regexpLookbehindAssertions invalid 1',
      },
      {
        code: 'var a = /(?<!a)foo/',
        options: [
          {
            version: '8.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'regexp-lookbehind-assertions' is not supported until Node.js >=8.10.0. The configured version range is '8.9.9'.",
            line: 1,
            column: 9,
            endLine: 1,
            endColumn: 20,
          },
        ],
        name: 'regexpLookbehind, regexpLookbehindAssertions invalid 2',
      },
      {
        code: 'var a = new RegExp("/(?<=a)foo/")',
        options: [
          {
            version: '8.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'regexp-lookbehind-assertions' is not supported until Node.js >=8.10.0. The configured version range is '8.9.9'.",
            line: 1,
            column: 9,
            endLine: 1,
            endColumn: 34,
          },
        ],
        name: 'regexpLookbehind, regexpLookbehindAssertions invalid 3',
      },
    ],
  },
);

// regexpNamedCaptureGroups
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        code: 'var a = /(?<key>a)foo/',
        options: [
          {
            version: '10.0.0',
          },
        ],
        name: 'regexpNamedCaptureGroups valid 1',
      },
      {
        code: 'var a = /(?<key>a)\\k<key>/',
        options: [
          {
            version: '10.0.0',
          },
        ],
        name: 'regexpNamedCaptureGroups valid 2',
      },
      {
        code: 'var a = new RegExp("(?<key>a)foo")',
        options: [
          {
            version: '10.0.0',
          },
        ],
        name: 'regexpNamedCaptureGroups valid 3',
      },
      {
        code: 'var a = new RegExp(pattern)',
        options: [
          {
            version: '8.9.9',
          },
        ],
        name: 'regexpNamedCaptureGroups valid 4',
      },
      {
        code: 'var a = new RegExp("(?<key")',
        options: [
          {
            version: '8.9.9',
          },
        ],
        name: 'regexpNamedCaptureGroups valid 5',
      },
      {
        code: 'var a = /\\(?<key>a\\)foo/',
        options: [
          {
            version: '8.9.9',
          },
        ],
        name: 'regexpNamedCaptureGroups valid 6',
      },
      {
        code: 'var a = /(?<key>a)foo/',
        options: [
          {
            version: '9.9.9',
            ignores: ['regexpNamedCaptureGroups'],
          },
        ],
        name: 'regexpNamedCaptureGroups valid 7',
      },
      {
        code: 'var a = /(?<key>a)\\k<key>/',
        options: [
          {
            version: '9.9.9',
            ignores: ['regexpNamedCaptureGroups'],
          },
        ],
        name: 'regexpNamedCaptureGroups valid 8',
      },
      {
        code: 'var a = new RegExp("(?<key>a)foo")',
        options: [
          {
            version: '9.9.9',
            ignores: ['regexpNamedCaptureGroups'],
          },
        ],
        name: 'regexpNamedCaptureGroups valid 9',
      },
    ],
    invalid: [
      {
        code: 'var a = /(?<key>a)foo/',
        options: [
          {
            version: '9.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'regexp-named-capture-groups' is not supported until Node.js >=10.0.0. The configured version range is '9.9.9'.",
            line: 1,
            column: 9,
            endLine: 1,
            endColumn: 23,
          },
        ],
        name: 'regexpNamedCaptureGroups invalid 1',
      },
      {
        code: 'var a = /(?<key>a)\\k<key>/',
        options: [
          {
            version: '9.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'regexp-named-capture-groups' is not supported until Node.js >=10.0.0. The configured version range is '9.9.9'.",
            line: 1,
            column: 9,
            endLine: 1,
            endColumn: 27,
          },
        ],
        name: 'regexpNamedCaptureGroups invalid 2',
      },
      {
        code: 'var a = new RegExp("(?<key>a)foo")',
        options: [
          {
            version: '9.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'regexp-named-capture-groups' is not supported until Node.js >=10.0.0. The configured version range is '9.9.9'.",
            line: 1,
            column: 9,
            endLine: 1,
            endColumn: 35,
          },
        ],
        name: 'regexpNamedCaptureGroups invalid 3',
      },
    ],
  },
);

// regexpS, regexpSFlag
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        code: 'var a = /foo/s',
        options: [
          {
            version: '8.10.0',
          },
        ],
        name: 'regexpS, regexpSFlag valid 1',
      },
      {
        code: 'var a = new RegExp("foo", "s")',
        options: [
          {
            version: '8.10.0',
          },
        ],
        name: 'regexpS, regexpSFlag valid 2',
      },
      {
        code: 'var a = new RegExp(a, b)',
        options: [
          {
            version: '8.9.9',
          },
        ],
        name: 'regexpS, regexpSFlag valid 3',
      },
      {
        code: 'var a = new RegExp("(aaaaa", b)',
        options: [
          {
            version: '8.9.9',
          },
        ],
        name: 'regexpS, regexpSFlag valid 4',
      },
      {
        code: 'var a = /foo/s',
        options: [
          {
            version: '8.9.9',
            ignores: ['regexpS'],
          },
        ],
        name: 'regexpS, regexpSFlag valid 5',
      },
      {
        code: 'var a = new RegExp("foo", "s")',
        options: [
          {
            version: '8.9.9',
            ignores: ['regexpS'],
          },
        ],
        name: 'regexpS, regexpSFlag valid 6',
      },
      {
        code: 'var a = /foo/s',
        options: [
          {
            version: '8.9.9',
            ignores: ['regexpSFlag'],
          },
        ],
        name: 'regexpS, regexpSFlag valid 7',
      },
      {
        code: 'var a = new RegExp("foo", "s")',
        options: [
          {
            version: '8.9.9',
            ignores: ['regexpSFlag'],
          },
        ],
        name: 'regexpS, regexpSFlag valid 8',
      },
    ],
    invalid: [
      {
        code: 'var a = /foo/s',
        options: [
          {
            version: '8.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'regexp-s-flag' is not supported until Node.js >=8.10.0. The configured version range is '8.9.9'.",
            line: 1,
            column: 9,
            endLine: 1,
            endColumn: 15,
          },
        ],
        name: 'regexpS, regexpSFlag invalid 1',
      },
      {
        code: 'var a = new RegExp("foo", "s")',
        options: [
          {
            version: '8.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'regexp-s-flag' is not supported until Node.js >=8.10.0. The configured version range is '8.9.9'.",
            line: 1,
            column: 9,
            endLine: 1,
            endColumn: 31,
          },
        ],
        name: 'regexpS, regexpSFlag invalid 2',
      },
    ],
  },
);

// regexpUnicodeProperties, regexpUnicodePropertyEscapes
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        code: 'var a = /\\p{Letter}/u',
        options: [
          {
            version: '10.0.0',
          },
        ],
        name: 'regexpUnicodeProperties, regexpUnicodePropertyEscapes valid 1',
      },
      {
        code: 'var a = /\\P{Letter}/u',
        options: [
          {
            version: '10.0.0',
          },
        ],
        name: 'regexpUnicodeProperties, regexpUnicodePropertyEscapes valid 2',
      },
      {
        code: 'var a = new RegExp("\\\\p{Letter}", "u")',
        options: [
          {
            version: '10.0.0',
          },
        ],
        name: 'regexpUnicodeProperties, regexpUnicodePropertyEscapes valid 3',
      },
      {
        code: 'var a = new RegExp("\\\\p{Letter}")',
        options: [
          {
            version: '9.9.9',
          },
        ],
        name: 'regexpUnicodeProperties, regexpUnicodePropertyEscapes valid 4',
      },
      {
        code: 'var a = new RegExp(pattern, "u")',
        options: [
          {
            version: '9.9.9',
          },
        ],
        name: 'regexpUnicodeProperties, regexpUnicodePropertyEscapes valid 5',
      },
      {
        code: 'var a = new RegExp("\\\\p{Letter")',
        options: [
          {
            version: '9.9.9',
          },
        ],
        name: 'regexpUnicodeProperties, regexpUnicodePropertyEscapes valid 6',
      },
      {
        code: 'var a = /\\p{Letter}/u',
        options: [
          {
            version: '9.9.9',
            ignores: ['regexpUnicodeProperties'],
          },
        ],
        name: 'regexpUnicodeProperties, regexpUnicodePropertyEscapes valid 7',
      },
      {
        code: 'var a = /\\P{Letter}/u',
        options: [
          {
            version: '9.9.9',
            ignores: ['regexpUnicodeProperties'],
          },
        ],
        name: 'regexpUnicodeProperties, regexpUnicodePropertyEscapes valid 8',
      },
      {
        code: 'var a = new RegExp("\\\\p{Letter}", "u")',
        options: [
          {
            version: '9.9.9',
            ignores: ['regexpUnicodeProperties'],
          },
        ],
        name: 'regexpUnicodeProperties, regexpUnicodePropertyEscapes valid 9',
      },
      {
        code: 'var a = /\\p{Letter}/u',
        options: [
          {
            version: '9.9.9',
            ignores: ['regexpUnicodePropertyEscapes'],
          },
        ],
        name: 'regexpUnicodeProperties, regexpUnicodePropertyEscapes valid 10',
      },
      {
        code: 'var a = /\\P{Letter}/u',
        options: [
          {
            version: '9.9.9',
            ignores: ['regexpUnicodePropertyEscapes'],
          },
        ],
        name: 'regexpUnicodeProperties, regexpUnicodePropertyEscapes valid 11',
      },
      {
        code: 'var a = new RegExp("\\\\p{Letter}", "u")',
        options: [
          {
            version: '9.9.9',
            ignores: ['regexpUnicodePropertyEscapes'],
          },
        ],
        name: 'regexpUnicodeProperties, regexpUnicodePropertyEscapes valid 12',
      },
    ],
    invalid: [
      {
        code: 'var a = /\\p{Letter}/u',
        options: [
          {
            version: '9.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'regexp-unicode-property-escapes' is not supported until Node.js >=10.0.0. The configured version range is '9.9.9'.",
            line: 1,
            column: 9,
            endLine: 1,
            endColumn: 22,
          },
        ],
        name: 'regexpUnicodeProperties, regexpUnicodePropertyEscapes invalid 1',
      },
      {
        code: 'var a = /\\P{Letter}/u',
        options: [
          {
            version: '9.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'regexp-unicode-property-escapes' is not supported until Node.js >=10.0.0. The configured version range is '9.9.9'.",
            line: 1,
            column: 9,
            endLine: 1,
            endColumn: 22,
          },
        ],
        name: 'regexpUnicodeProperties, regexpUnicodePropertyEscapes invalid 2',
      },
      {
        code: 'var a = new RegExp("\\\\p{Letter}", "u")',
        options: [
          {
            version: '9.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'regexp-unicode-property-escapes' is not supported until Node.js >=10.0.0. The configured version range is '9.9.9'.",
            line: 1,
            column: 9,
            endLine: 1,
            endColumn: 39,
          },
        ],
        name: 'regexpUnicodeProperties, regexpUnicodePropertyEscapes invalid 3',
      },
    ],
  },
);

// restSpreadProperties
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        code: '({ ...obj })',
        options: [
          {
            version: '8.3.0',
          },
        ],
        name: 'restSpreadProperties valid 1',
      },
      {
        code: '({ ...rest } = obj)',
        options: [
          {
            version: '8.3.0',
          },
        ],
        name: 'restSpreadProperties valid 2',
      },
      {
        code: '({ obj })',
        options: [
          {
            version: '8.2.9',
          },
        ],
        name: 'restSpreadProperties valid 3',
      },
      {
        code: '({ obj: 1 })',
        options: [
          {
            version: '8.2.9',
          },
        ],
        name: 'restSpreadProperties valid 4',
      },
      {
        code: '({ obj } = a)',
        options: [
          {
            version: '8.2.9',
          },
        ],
        name: 'restSpreadProperties valid 5',
      },
      {
        code: '({ obj: a } = b)',
        options: [
          {
            version: '8.2.9',
          },
        ],
        name: 'restSpreadProperties valid 6',
      },
      {
        code: '([...xs])',
        options: [
          {
            version: '8.2.9',
          },
        ],
        name: 'restSpreadProperties valid 7',
      },
      {
        code: '([a, ...xs] = ys)',
        options: [
          {
            version: '8.2.9',
          },
        ],
        name: 'restSpreadProperties valid 8',
      },
      {
        code: '({ ...obj })',
        options: [
          {
            version: '8.2.9',
            ignores: ['restSpreadProperties'],
          },
        ],
        name: 'restSpreadProperties valid 9',
      },
      {
        code: '({ ...rest } = obj)',
        options: [
          {
            version: '8.2.9',
            ignores: ['restSpreadProperties'],
          },
        ],
        name: 'restSpreadProperties valid 10',
      },
    ],
    invalid: [
      {
        code: '({ ...obj })',
        options: [
          {
            version: '8.2.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'rest-spread-properties' is not supported until Node.js >=8.3.0. The configured version range is '8.2.9'.",
            line: 1,
            column: 4,
            endLine: 1,
            endColumn: 10,
          },
        ],
        name: 'restSpreadProperties invalid 1',
      },
      {
        code: '({ ...rest } = obj)',
        options: [
          {
            version: '8.2.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'rest-spread-properties' is not supported until Node.js >=8.3.0. The configured version range is '8.2.9'.",
            line: 1,
            column: 4,
            endLine: 1,
            endColumn: 11,
          },
        ],
        name: 'restSpreadProperties invalid 2',
      },
    ],
  },
);

// jsonSuperset
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        code: "var s = 'foo'",
        options: [
          {
            version: '9.99.99',
          },
        ],
        name: 'jsonSuperset valid 1',
      },
      {
        code: "var s = '\\\u2028'",
        options: [
          {
            version: '9.99.99',
          },
        ],
        name: 'jsonSuperset valid 2',
      },
      {
        code: "var s = '\\\u2029'",
        options: [
          {
            version: '9.99.99',
          },
        ],
        name: 'jsonSuperset valid 3',
      },
      {
        code: "var s = '\u2028'",
        options: [
          {
            version: '10.0.0',
          },
        ],
        name: 'jsonSuperset valid 4',
      },
      {
        code: "var s = '\u2029'",
        options: [
          {
            version: '10.0.0',
          },
        ],
        name: 'jsonSuperset valid 5',
      },
      {
        code: "var s = '\u2028'",
        options: [
          {
            version: '9.99.99',
            ignores: ['jsonSuperset'],
          },
        ],
        name: 'jsonSuperset valid 6',
      },
      {
        code: "var s = '\u2029'",
        options: [
          {
            version: '9.99.99',
            ignores: ['jsonSuperset'],
          },
        ],
        name: 'jsonSuperset valid 7',
      },
    ],
    invalid: [
      {
        code: "var s = '\u2028'",
        options: [
          {
            version: '9.99.99',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'json-superset' is not supported until Node.js >=10.0.0. The configured version range is '9.99.99'.",
            line: 1,
            column: 10,
            endLine: 1,
            endColumn: 10,
          },
        ],
        name: 'jsonSuperset invalid 1',
      },
      {
        code: "var s = '\u2029'",
        options: [
          {
            version: '9.99.99',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'json-superset' is not supported until Node.js >=10.0.0. The configured version range is '9.99.99'.",
            line: 1,
            column: 10,
            endLine: 1,
            endColumn: 10,
          },
        ],
        name: 'jsonSuperset invalid 2',
      },
    ],
  },
);

// optionalCatchBinding
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        code: 'try {} catch {}',
        options: [
          {
            version: '10.0.0',
          },
        ],
        name: 'optionalCatchBinding valid 1',
      },
      {
        code: 'try {} catch (error) {}',
        options: [
          {
            version: '9.99.99',
          },
        ],
        name: 'optionalCatchBinding valid 2',
      },
      {
        code: 'try {} catch {}',
        options: [
          {
            version: '9.99.99',
            ignores: ['optionalCatchBinding'],
          },
        ],
        name: 'optionalCatchBinding valid 3',
      },
    ],
    invalid: [
      {
        code: 'try {} catch {}',
        options: [
          {
            version: '9.99.99',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'optional-catch-binding' is not supported until Node.js >=10.0.0. The configured version range is '9.99.99'.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 16,
          },
        ],
        name: 'optionalCatchBinding invalid 1',
      },
    ],
  },
);

// bigint
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        code: 'var n = 0n',
        options: [
          {
            version: '10.4.0',
          },
        ],
        name: 'bigint valid 1',
      },
      {
        code: 'var n = BigInt(0)',
        options: [
          {
            version: '10.4.0',
          },
        ],
        name: 'bigint valid 2',
      },
      {
        code: 'var n = new BigInt64Array()',
        options: [
          {
            version: '10.4.0',
          },
        ],
        name: 'bigint valid 3',
      },
      {
        code: 'var n = new BigUint64Array()',
        options: [
          {
            version: '10.4.0',
          },
        ],
        name: 'bigint valid 4',
      },
      {
        code: 'var n = { [0n]: 0 }',
        options: [
          {
            version: '10.4.0',
          },
        ],
        name: 'bigint valid 5',
      },
      {
        code: 'var n = class { [0n]() {} }',
        options: [
          {
            version: '10.4.0',
          },
        ],
        name: 'bigint valid 6',
      },
      {
        code: 'var n = 0n',
        options: [
          {
            version: '10.3.0',
            ignores: ['bigint'],
          },
        ],
        name: 'bigint valid 7',
      },
    ],
    invalid: [
      {
        code: 'var n = 0n',
        options: [
          {
            version: '10.3.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'bigint' is not supported until Node.js >=10.4.0. The configured version range is '10.3.0'.",
            line: 1,
            column: 9,
            endLine: 1,
            endColumn: 11,
          },
        ],
        name: 'bigint invalid 1',
      },
    ],
  },
);

// dynamicImport
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        code: 'obj.import(source)',
        options: [
          {
            version: '12.0.0',
          },
        ],
        name: 'dynamicImport valid 1',
      },
      {
        code: 'import(source)',
        options: [
          {
            version: '12.17.0',
          },
        ],
        name: 'dynamicImport valid 2',
      },
      {
        code: 'import(source)',
        options: [
          {
            version: '13.2.0',
          },
        ],
        name: 'dynamicImport valid 3',
      },
      {
        code: 'import(source)',
        options: [
          {
            version: '12.16.0',
            ignores: ['dynamicImport'],
          },
        ],
        name: 'dynamicImport valid 4',
      },
      {
        code: 'import(source)',
        options: [
          {
            version: '13.0.0',
            ignores: ['dynamicImport'],
          },
        ],
        name: 'dynamicImport valid 5',
      },
      {
        code: 'import(source)',
        options: [
          {
            version: '13.1.0',
            ignores: ['dynamicImport'],
          },
        ],
        name: 'dynamicImport valid 6',
      },
      {
        code: 'import(source)',
        options: [
          {
            version: '>=8.0.0',
            ignores: ['dynamicImport'],
          },
        ],
        name: 'dynamicImport valid 7',
      },
    ],
    invalid: [
      {
        code: 'import(source)',
        options: [
          {
            version: '12.16.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'dynamic-import' is not supported until Node.js ^12.17.0 || >=13.2.0. The configured version range is '12.16.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 15,
          },
        ],
        name: 'dynamicImport invalid 1',
      },
      {
        code: 'import(source)',
        options: [
          {
            version: '13.0.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'dynamic-import' is not supported until Node.js ^12.17.0 || >=13.2.0. The configured version range is '13.0.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 15,
          },
        ],
        name: 'dynamicImport invalid 2',
      },
      {
        code: 'import(source)',
        options: [
          {
            version: '13.1.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'dynamic-import' is not supported until Node.js ^12.17.0 || >=13.2.0. The configured version range is '13.1.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 15,
          },
        ],
        name: 'dynamicImport invalid 3',
      },
      {
        code: 'import(source)',
        options: [
          {
            version: '>=8.0.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'dynamic-import' is not supported until Node.js ^12.17.0 || >=13.2.0. The configured version range is '>=8.0.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 15,
          },
        ],
        name: 'dynamicImport invalid 4',
      },
    ],
  },
);

// optionalChaining
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        code: 'foo?.bar;',
        options: [
          {
            version: '14.0.0',
          },
        ],
        name: 'optionalChaining valid 1',
      },
      {
        code: 'foo?.bar',
        options: [
          {
            version: '13.0.0',
            ignores: ['optionalChaining'],
          },
        ],
        name: 'optionalChaining valid 2',
      },
    ],
    invalid: [
      {
        code: 'foo?.bar',
        options: [
          {
            version: '13.0.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'optional-chaining' is not supported until Node.js >=14.0.0. The configured version range is '13.0.0'.",
            line: 1,
            column: 4,
            endLine: 1,
            endColumn: 6,
          },
        ],
        name: 'optionalChaining invalid 1',
      },
    ],
  },
);

// nullishCoalescingOperators
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        code: 'foo ?? bar;',
        options: [
          {
            version: '14.0.0',
          },
        ],
        name: 'nullishCoalescingOperators valid 1',
      },
      {
        code: 'foo ?? bar;',
        settings: {
          node: {
            version: '14.0.0',
          },
        },
        name: 'nullishCoalescingOperators valid 2',
      },
      {
        code: 'foo ?? bar',
        options: [
          {
            version: '13.0.0',
            ignores: ['nullishCoalescingOperators'],
          },
        ],
        name: 'nullishCoalescingOperators valid 3',
      },
      {
        code: 'foo ?? bar',
        settings: {
          node: {
            version: '13.0.0',
          },
        },
        options: [
          {
            ignores: ['nullishCoalescingOperators'],
          },
        ],
        name: 'nullishCoalescingOperators valid 4',
      },
    ],
    invalid: [
      {
        code: 'foo ?? bar',
        options: [
          {
            version: '13.0.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'nullish-coalescing-operators' is not supported until Node.js >=14.0.0. The configured version range is '13.0.0'.",
            line: 1,
            column: 5,
            endLine: 1,
            endColumn: 7,
          },
        ],
        name: 'nullishCoalescingOperators invalid 1',
      },
      {
        code: 'foo ?? bar',
        settings: {
          node: {
            version: '13.0.0',
          },
        },
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'nullish-coalescing-operators' is not supported until Node.js >=14.0.0. The configured version range is '13.0.0'.",
            line: 1,
            column: 5,
            endLine: 1,
            endColumn: 7,
          },
        ],
        name: 'nullishCoalescingOperators invalid 2',
      },
    ],
  },
);

// logicalAssignmentOperators
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        code: 'a ||= b',
        options: [
          {
            version: '15.0.0',
          },
        ],
        name: 'logicalAssignmentOperators valid 1',
      },
      {
        code: 'a &&= b',
        options: [
          {
            version: '15.0.0',
          },
        ],
        name: 'logicalAssignmentOperators valid 2',
      },
      {
        code: 'a ??= b',
        options: [
          {
            version: '15.0.0',
          },
        ],
        name: 'logicalAssignmentOperators valid 3',
      },
      {
        code: 'a ||= b',
        options: [
          {
            version: '14.0.0',
            ignores: ['logicalAssignmentOperators'],
          },
        ],
        name: 'logicalAssignmentOperators valid 4',
      },
      {
        code: 'a &&= b',
        options: [
          {
            version: '14.0.0',
            ignores: ['logicalAssignmentOperators'],
          },
        ],
        name: 'logicalAssignmentOperators valid 5',
      },
      {
        code: 'a ??= b',
        options: [
          {
            version: '14.0.0',
            ignores: ['logicalAssignmentOperators'],
          },
        ],
        name: 'logicalAssignmentOperators valid 6',
      },
    ],
    invalid: [
      {
        code: 'a ||= b',
        options: [
          {
            version: '14.0.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'logical-assignment-operators' is not supported until Node.js >=15.0.0. The configured version range is '14.0.0'.",
            line: 1,
            column: 3,
            endLine: 1,
            endColumn: 6,
          },
        ],
        name: 'logicalAssignmentOperators invalid 1',
      },
      {
        code: 'a &&= b',
        options: [
          {
            version: '14.0.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'logical-assignment-operators' is not supported until Node.js >=15.0.0. The configured version range is '14.0.0'.",
            line: 1,
            column: 3,
            endLine: 1,
            endColumn: 6,
          },
        ],
        name: 'logicalAssignmentOperators invalid 2',
      },
      {
        code: 'a ??= b',
        options: [
          {
            version: '14.0.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'logical-assignment-operators' is not supported until Node.js >=15.0.0. The configured version range is '14.0.0'.",
            line: 1,
            column: 3,
            endLine: 1,
            endColumn: 6,
          },
        ],
        name: 'logicalAssignmentOperators invalid 3',
      },
    ],
  },
);

// numericSeparators
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        code: 'a = 123_456_789',
        options: [
          {
            version: '12.5.0',
          },
        ],
        name: 'numericSeparators valid 1',
      },
      {
        code: 'a = 123_456_789',
        options: [
          {
            version: '12.4.0',
            ignores: ['numericSeparators'],
          },
        ],
        name: 'numericSeparators valid 2',
      },
    ],
    invalid: [
      {
        code: 'a = 123_456_789',
        options: [
          {
            version: '12.4.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'numeric-separators' is not supported until Node.js >=12.5.0. The configured version range is '12.4.0'.",
            line: 1,
            column: 5,
            endLine: 1,
            endColumn: 16,
          },
        ],
        name: 'numericSeparators invalid 1',
      },
    ],
  },
);

// configuration
ruleTester.run(
  'no-unsupported-features/es-syntax',
  {},
  {
    valid: [
      {
        filename: 'gte-4.0.0/a.js',
        code: 'var a = () => 1',
        name: 'configuration valid 1',
      },
      {
        filename: 'gte-4.4.0-lt-5.0.0/a.js',
        code: 'var a = () => 1',
        name: 'configuration valid 2',
      },
      {
        filename: 'hat-4.1.2/a.js',
        code: 'var a = () => 1',
        name: 'configuration valid 3',
      },
      {
        code: "'\\\\u{0123}'",
        name: 'configuration valid 4',
      },
      {
        filename: 'gte-4.0.0/a.js',
        code: 'var a = async () => 1',
        options: [
          {
            ignores: ['asyncFunctions'],
          },
        ],
        name: 'configuration valid 5',
      },
      {
        filename: 'gte-7.6.0/a.js',
        code: 'var a = async () => 1',
        name: 'configuration valid 6',
      },
      {
        filename: 'gte-7.10.0/a.js',
        code: 'var a = async () => 1',
        name: 'configuration valid 7',
      },
      {
        filename: 'invalid/a.js',
        code: 'var a = () => 1',
        name: 'configuration valid 8',
      },
      {
        filename: 'nothing/a.js',
        code: 'var a = () => 1',
        name: 'configuration valid 9',
      },
      {
        code: 'var a = async () => 1',
        options: [
          {
            version: '7.10.0',
          },
        ],
        name: 'configuration valid 10',
      },
      {
        code: 'var a = async () => 1',
        settings: {
          node: {
            version: '7.10.0',
          },
        },
        name: 'configuration valid 11',
      },
      {
        filename: 'without-node/a.js',
        code: 'var a = () => 1',
        name: 'configuration valid 12',
      },
      {
        filename: 'dev-engines-non-node-only/a.js',
        code: 'var a = async () => 1',
        name: 'configuration valid 13',
      },
    ],
    invalid: [
      {
        filename: 'gte-0.12.8/a.js',
        code: 'var a = () => 1',
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'arrow-functions' is not supported until Node.js >=4.0.0. The configured version range is '>=0.12.8'.",
            line: 1,
            column: 9,
            endLine: 1,
            endColumn: 16,
          },
        ],
        name: 'configuration invalid 1',
      },
      {
        filename: 'invalid/a.js',
        code: 'var a = { ...obj }',
        options: [
          {
            version: '>=8.0.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'rest-spread-properties' is not supported until Node.js >=8.3.0. The configured version range is '>=8.0.0'.",
            line: 1,
            column: 11,
            endLine: 1,
            endColumn: 17,
          },
        ],
        name: 'configuration invalid 2',
      },
      {
        filename: 'lt-6.0.0/a.js',
        code: 'var a = () => 1',
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'arrow-functions' is not supported until Node.js >=4.0.0. The configured version range is '<6.0.0'.",
            line: 1,
            column: 9,
            endLine: 1,
            endColumn: 16,
          },
        ],
        name: 'configuration invalid 3',
      },
      {
        filename: 'nothing/a.js',
        code: 'var a = { ...obj }',
        options: [
          {
            version: '>=8.0.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'rest-spread-properties' is not supported until Node.js >=8.3.0. The configured version range is '>=8.0.0'.",
            line: 1,
            column: 11,
            endLine: 1,
            endColumn: 17,
          },
        ],
        name: 'configuration invalid 4',
      },
      {
        filename: 'gte-7.5.0/a.js',
        code: 'var a = async () => 1',
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'async-functions' is not supported until Node.js >=7.6.0. The configured version range is '>=7.5.0'.",
            line: 1,
            column: 9,
            endLine: 1,
            endColumn: 22,
          },
        ],
        name: 'configuration invalid 5',
      },
      {
        filename: 'star/a.js',
        code: '"use strict"; let a = 1',
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'block-scoped-variables' is not supported until Node.js >=4.0.0. The configured version range is '*'.",
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 24,
          },
        ],
        name: 'configuration invalid 6',
      },
      {
        filename: 'dev-engines-gte-7.5.0/a.js',
        code: 'var a = async () => 1',
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'async-functions' is not supported until Node.js >=7.6.0. The configured version range is '>=7.5.0'.",
            line: 1,
            column: 9,
            endLine: 1,
            endColumn: 22,
          },
        ],
        name: 'configuration invalid 7',
      },
      {
        filename: 'dev-engines-array-gte-7.5.0/a.js',
        code: 'var a = async () => 1',
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'async-functions' is not supported until Node.js >=7.6.0. The configured version range is '>=7.5.0'.",
            line: 1,
            column: 9,
            endLine: 1,
            endColumn: 22,
          },
        ],
        name: 'configuration invalid 8',
      },
      {
        filename: 'engines-over-dev-engines/a.js',
        code: 'var a = async () => 1',
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'async-functions' is not supported until Node.js >=7.6.0. The configured version range is '>=4.0.0'.",
            line: 1,
            column: 9,
            endLine: 1,
            endColumn: 22,
          },
        ],
        name: 'configuration invalid 9',
      },
      {
        filename: 'dev-engines-non-node/a.js',
        code: 'var a = async () => 1',
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'async-functions' is not supported until Node.js >=7.6.0. The configured version range is '>=7.5.0'.",
            line: 1,
            column: 9,
            endLine: 1,
            endColumn: 22,
          },
        ],
        name: 'configuration invalid 10',
      },
      {
        code: 'var a = async () => 1',
        options: [
          {
            version: '7.1.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'async-functions' is not supported until Node.js >=7.6.0. The configured version range is '7.1.0'.",
            line: 1,
            column: 9,
            endLine: 1,
            endColumn: 22,
          },
        ],
        name: 'configuration invalid 11',
      },
      {
        code: 'var a = async () => 1',
        settings: {
          node: {
            version: '7.1.0',
          },
        },
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "'async-functions' is not supported until Node.js >=7.6.0. The configured version range is '7.1.0'.",
            line: 1,
            column: 9,
            endLine: 1,
            endColumn: 22,
          },
        ],
        name: 'configuration invalid 12',
      },
    ],
  },
);
