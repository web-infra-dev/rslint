import { RuleTester } from '../rule-tester.js';

const ruleTester = new RuleTester();
const rule = null as never;

const IMPORT_ERROR_MESSAGE =
  'Expected 1 empty line after import statement not followed by another import.';
const IMPORT_ERROR_MESSAGE_MULTIPLE = (count: number) =>
  `Expected ${count} empty lines after import statement not followed by another import.`;
const REQUIRE_ERROR_MESSAGE =
  'Expected 1 empty line after require statement not followed by another require.';
const REQUIRE_ERROR_MESSAGE_MULTIPLE = (count: number) =>
  `Expected ${count} empty lines after require statement not followed by another require.`;

ruleTester.run('newline-after-import', rule, {
  valid: [
    // ===== Import declarations =====
    { code: `import path from 'path';\nimport foo from 'foo';\n` },
    { code: `import path from 'path';import foo from 'foo';\n` },
    { code: `import path from 'path';import foo from 'foo';\n\nvar bar = 42;` },
    { code: `import foo from 'foo';\n\nvar bar = 'bar';` },
    { code: `import foo from 'foo';` },
    // Various import forms
    { code: `import * as foo from 'foo';\n\nvar x = 1;` },
    { code: `import 'side-effect';\n\nvar x = 1;` },
    { code: `import type { Foo } from 'foo';\n\nvar x = 1;` },
    // Import followed by export block
    { code: `import stub from './stub';\n\nexport { stub }` },

    // ===== count option =====
    {
      code: `import foo from 'foo';\n\n\nvar bar = 'bar';`,
      options: [{ count: 2 }],
    },
    {
      code: `import foo from 'foo';\n\n\nvar bar = 'bar';`,
      options: [{ count: 2, exactCount: true }],
    },
    {
      code: `import foo from 'foo';\n\nvar bar = 'bar';`,
      options: [{ count: 1, exactCount: true }],
    },
    // More than required lines (no exactCount)
    {
      code: `import foo from 'foo';\n\n\nvar bar = 'bar';`,
      options: [{ count: 1 }],
    },
    {
      code: `import foo from 'foo';\n\n\n\n\nvar bar = 'bar';`,
      options: [{ count: 4 }],
    },
    // Multiple import groups
    { code: `import foo from 'foo';\nimport { bar } from './bar-lib';` },
    {
      code: `import foo from 'foo';\n\nvar a = 123;\n\nimport { bar } from './bar-lib';`,
    },

    // ===== considerComments =====
    {
      code: `import foo from 'foo';\n\n// Some random comment\nvar bar = 'bar';`,
      options: [{ count: 1, exactCount: true, considerComments: true }],
    },
    {
      code: `import foo from 'foo';\n\n\n// Some random comment\nvar bar = 'bar';`,
      options: [{ count: 2, exactCount: true, considerComments: true }],
    },
    // Comments without considerComments don't affect gap
    {
      code: `import foo from 'foo';\n\n// Some random comment\nvar bar = 'bar';`,
      options: [{ count: 2, exactCount: true }],
    },
    {
      code: `import foo from 'foo';\n// Some random comment\nvar bar = 'bar';`,
      options: [{ count: 1, exactCount: true }],
    },
    // Multiline block comment without considerComments
    {
      code: `import path from 'path';\nimport foo from 'foo';\n/**\n * some multiline comment\n**/\nvar bar = 42;`,
    },

    // ===== Require calls =====
    { code: `var path = require('path');\nvar foo = require('foo');\n` },
    { code: `require('foo');` },
    { code: `var foo = require('foo-module');\n\nvar foo = 'bar';` },
    {
      code: `var foo = require('foo-module');\n\n\nvar foo = 'bar';`,
      options: [{ count: 2 }],
    },
    {
      code: `var foo = require('foo-module');\n\n\n\n\nvar foo = 'bar';`,
      options: [{ count: 4, exactCount: true }],
    },
    { code: `require('foo-module');\n\nvar foo = 'bar';` },
    {
      code: `var foo = require('foo-module');\n\nvar a = 123;\n\nvar bar = require('bar-lib');`,
    },
    {
      code: `var foo = require('foo-module');\n\n\n// Some random comment\nvar foo = 'bar';`,
      options: [{ count: 2, considerComments: true }],
    },

    // ===== Require scope boundaries (NOT top-level) =====
    { code: `function x() { require('baz'); }` },
    { code: `a(require('b'), require('c'), require('d'));` },
    { code: `switch ('foo') { case 'bar': require('baz'); }` },
    { code: `if (true) {\n  var foo = require('foo');\n  foo();\n}` },
    { code: `const x = () => require('baz') && require('bar')` },
    { code: `var x = { foo: require('foo') };\nvar y = 1;` },

    // ===== TSImportEqualsDeclaration =====
    {
      code: `import { ExecaReturnValue } from 'execa';\nimport execa = require('execa');`,
    },
    {
      code: `import execa = require('execa');\nimport { ExecaReturnValue } from 'execa';`,
    },
    { code: `export import a = obj;\nf(a);` },

    // ===== Decorator class =====
    { code: `import foo from 'foo';\n\n@SomeDecorator(foo)\nclass Foo {}` },
  ],
  invalid: [
    // ===== Basic import errors =====
    {
      code: `import foo from 'foo';\nexport default function() {};`,
      errors: [{ message: IMPORT_ERROR_MESSAGE }],
    },
    {
      code: `import foo from 'foo';\n\nexport default function() {};`,
      options: [{ count: 2 }],
      errors: [{ message: IMPORT_ERROR_MESSAGE_MULTIPLE(2) }],
    },
    {
      code: `import path from 'path';\nimport foo from 'foo';\nvar bar = 42;`,
      errors: [{ message: IMPORT_ERROR_MESSAGE }],
    },
    {
      code: `import path from 'path';import foo from 'foo';var bar = 42;`,
      errors: [{ message: IMPORT_ERROR_MESSAGE }],
    },
    // Two import groups
    {
      code: `import foo from 'foo';\nvar a = 123;\n\nimport { bar } from './bar-lib';\nvar b=456;`,
      errors: [
        { message: IMPORT_ERROR_MESSAGE },
        { message: IMPORT_ERROR_MESSAGE },
      ],
    },

    // ===== Require errors =====
    {
      code: `var foo = require('foo-module');\nvar something = 123;`,
      errors: [{ message: REQUIRE_ERROR_MESSAGE }],
    },
    {
      code: `var path = require('path');\nvar foo = require('foo');\nvar bar = 42;`,
      errors: [{ message: REQUIRE_ERROR_MESSAGE }],
    },
    // Two require groups
    {
      code: `var foo = require('foo-module');\nvar a = 123;\n\nvar bar = require('bar-lib');\nvar b=456;`,
      errors: [
        { message: REQUIRE_ERROR_MESSAGE },
        { message: REQUIRE_ERROR_MESSAGE },
      ],
    },
    // Require in binary expression (top-level)
    {
      code: `var assign = Object.assign || require('object-assign');\nvar foo = require('foo');\nvar bar = 42;`,
      errors: [{ message: REQUIRE_ERROR_MESSAGE }],
    },
    // Mixed require with function args
    {
      code: `require('a');\nfoo(require('b'), require('c'), require('d'));\nrequire('d');\nvar foo = 'bar';`,
      errors: [{ message: REQUIRE_ERROR_MESSAGE }],
    },

    // ===== exactCount =====
    {
      code: `import foo from 'foo';\n\nexport default function() {};`,
      options: [{ count: 2, exactCount: true }],
      errors: [{ message: IMPORT_ERROR_MESSAGE_MULTIPLE(2) }],
    },
    {
      code: `import foo from 'foo';\n\n\n\nexport default function() {};`,
      options: [{ count: 2, exactCount: true }],
      errors: [{ message: IMPORT_ERROR_MESSAGE_MULTIPLE(2) }],
    },
    {
      code: `import foo from 'foo';export default function() {};`,
      options: [{ count: 1, exactCount: true }],
      errors: [{ message: IMPORT_ERROR_MESSAGE }],
    },

    // ===== considerComments =====
    {
      code: `import path from 'path';\nimport foo from 'foo';\n// Some random single line comment\nvar bar = 42;`,
      options: [{ considerComments: true, count: 1 }],
      errors: [{ message: IMPORT_ERROR_MESSAGE }],
    },
    {
      code: `var foo = require('foo-module');\n\n/**\n * Test comment\n */\nvar foo = 'bar';`,
      options: [{ considerComments: true, count: 2 }],
      errors: [{ message: REQUIRE_ERROR_MESSAGE_MULTIPLE(2) }],
    },
    // considerComments + exactCount
    {
      code: `import foo from 'foo';\n// some random comment\nexport default function() {};`,
      options: [{ count: 2, exactCount: true, considerComments: true }],
      errors: [{ message: IMPORT_ERROR_MESSAGE_MULTIPLE(2) }],
    },
    // Same-line trailing comment with considerComments
    {
      code: `import foo from 'foo';// some random comment\nexport default function() {};`,
      options: [{ count: 1, exactCount: true, considerComments: true }],
      errors: [{ message: IMPORT_ERROR_MESSAGE }],
    },

    // ===== Decorator class =====
    {
      code: `import foo from 'foo';\n@SomeDecorator(foo)\nclass Foo {}`,
      errors: [{ message: IMPORT_ERROR_MESSAGE }],
    },

    // ===== Mixed import + require =====
    {
      code: `import foo from 'foo';\nvar bar = require('bar');\nvar baz = 42;`,
      errors: [
        { message: IMPORT_ERROR_MESSAGE },
        { message: REQUIRE_ERROR_MESSAGE },
      ],
    },
  ],
});
