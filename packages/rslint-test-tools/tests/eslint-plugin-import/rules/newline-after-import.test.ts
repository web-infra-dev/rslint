// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
new RuleTester(fixtureConfig()).run('newline-after-import', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: "import path from 'path';\nimport foo from 'foo';\n",
    },
    {
      name: 'run 0 valid 1',
      code: "import path from 'path';import foo from 'foo';\n",
    },
    {
      name: 'run 0 valid 2',
      code: "import path from 'path';import foo from 'foo';\n\nvar bar = 42;",
    },
    {
      name: 'run 0 valid 3',
      code: "import foo from 'foo';\n\nvar bar = 'bar';",
    },
    {
      name: 'run 0 valid 4',
      code: "import foo from 'foo';",
    },
    {
      name: 'run 0 valid 5',
      code: "import * as foo from 'foo';\n\nvar x = 1;",
    },
    {
      name: 'run 0 valid 6',
      code: "import 'side-effect';\n\nvar x = 1;",
    },
    {
      name: 'run 0 valid 7',
      code: "import type { Foo } from 'foo';\n\nvar x = 1;",
    },
    {
      name: 'run 0 valid 8',
      code: "import stub from './stub';\n\nexport { stub }",
    },
    {
      name: 'run 0 valid 9',
      code: "import foo from 'foo';\n\n\nvar bar = 'bar';",
      options: [
        {
          count: 2,
        },
      ],
    },
    {
      name: 'run 0 valid 10',
      code: "import foo from 'foo';\n\n\nvar bar = 'bar';",
      options: [
        {
          count: 2,
          exactCount: true,
        },
      ],
    },
    {
      name: 'run 0 valid 11',
      code: "import foo from 'foo';\n\nvar bar = 'bar';",
      options: [
        {
          count: 1,
          exactCount: true,
        },
      ],
    },
    {
      name: 'run 0 valid 12',
      code: "import foo from 'foo';\n\n\nvar bar = 'bar';",
      options: [
        {
          count: 1,
        },
      ],
    },
    {
      name: 'run 0 valid 13',
      code: "import foo from 'foo';\n\n\n\n\nvar bar = 'bar';",
      options: [
        {
          count: 4,
        },
      ],
    },
    {
      name: 'run 0 valid 14',
      code: "import foo from 'foo';\nimport { bar } from './bar-lib';",
    },
    {
      name: 'run 0 valid 15',
      code: "import foo from 'foo';\n\nvar a = 123;\n\nimport { bar } from './bar-lib';",
    },
    {
      name: 'run 0 valid 16',
      code: "import foo from 'foo';\n\n// Some random comment\nvar bar = 'bar';",
      options: [
        {
          count: 1,
          exactCount: true,
          considerComments: true,
        },
      ],
    },
    {
      name: 'run 0 valid 17',
      code: "import foo from 'foo';\n\n\n// Some random comment\nvar bar = 'bar';",
      options: [
        {
          count: 2,
          exactCount: true,
          considerComments: true,
        },
      ],
    },
    {
      name: 'run 0 valid 18',
      code: "import foo from 'foo';\n\n// Some random comment\nvar bar = 'bar';",
      options: [
        {
          count: 2,
          exactCount: true,
        },
      ],
    },
    {
      name: 'run 0 valid 19',
      code: "import foo from 'foo';\n// Some random comment\nvar bar = 'bar';",
      options: [
        {
          count: 1,
          exactCount: true,
        },
      ],
    },
    {
      name: 'run 0 valid 20',
      code: "import path from 'path';\nimport foo from 'foo';\n/**\n * some multiline comment\n**/\nvar bar = 42;",
    },
    {
      name: 'run 0 valid 21',
      code: "var path = require('path');\nvar foo = require('foo');\n",
    },
    {
      name: 'run 0 valid 22',
      code: "require('foo');",
    },
    {
      name: 'run 0 valid 23',
      code: "var foo = require('foo-module');\n\nvar foo = 'bar';",
    },
    {
      name: 'run 0 valid 24',
      code: "var foo = require('foo-module');\n\n\nvar foo = 'bar';",
      options: [
        {
          count: 2,
        },
      ],
    },
    {
      name: 'run 0 valid 25',
      code: "var foo = require('foo-module');\n\n\n\n\nvar foo = 'bar';",
      options: [
        {
          count: 4,
          exactCount: true,
        },
      ],
    },
    {
      name: 'run 0 valid 26',
      code: "require('foo-module');\n\nvar foo = 'bar';",
    },
    {
      name: 'run 0 valid 27',
      code: "var foo = require('foo-module');\n\nvar a = 123;\n\nvar bar = require('bar-lib');",
    },
    {
      name: 'run 0 valid 28',
      code: "var foo = require('foo-module');\n\n\n// Some random comment\nvar foo = 'bar';",
      options: [
        {
          count: 2,
          considerComments: true,
        },
      ],
    },
    {
      name: 'run 0 valid 29',
      code: "function x() { require('baz'); }",
    },
    {
      name: 'run 0 valid 30',
      code: "a(require('b'), require('c'), require('d'));",
    },
    {
      name: 'run 0 valid 31',
      code: "switch ('foo') { case 'bar': require('baz'); }",
    },
    {
      name: 'run 0 valid 32',
      code: "if (true) {\n  var foo = require('foo');\n  foo();\n}",
    },
    {
      name: 'run 0 valid 33',
      code: "const x = () => require('baz') && require('bar')",
    },
    {
      name: 'run 0 valid 34',
      code: "var x = { foo: require('foo') };\nvar y = 1;",
    },
    {
      name: 'run 0 valid 35',
      code: "import { ExecaReturnValue } from 'execa';\nimport execa = require('execa');",
    },
    {
      name: 'run 0 valid 36',
      code: "import execa = require('execa');\nimport { ExecaReturnValue } from 'execa';",
    },
    {
      name: 'run 0 valid 37',
      code: 'export import a = obj;\nf(a);',
    },
    {
      name: 'run 0 valid 38',
      code: "import foo from 'foo';\n\n@SomeDecorator(foo)\nclass Foo {}",
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: "import foo from 'foo';\nexport default function() {};",
      errors: [
        {
          messageId: 'newlineAfterImport',
          message:
            'Expected 1 empty line after import statement not followed by another import.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 1,
        },
      ],
      output: "import foo from 'foo';\n\nexport default function() {};",
    },
    {
      name: 'run 0 invalid 1',
      code: "import foo from 'foo';\n\nexport default function() {};",
      options: [
        {
          count: 2,
        },
      ],
      errors: [
        {
          messageId: 'newlineAfterImport',
          message:
            'Expected 2 empty lines after import statement not followed by another import.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 1,
        },
      ],
      output: "import foo from 'foo';\n\n\nexport default function() {};",
    },
    {
      name: 'run 0 invalid 2',
      code: "import path from 'path';\nimport foo from 'foo';\nvar bar = 42;",
      errors: [
        {
          messageId: 'newlineAfterImport',
          message:
            'Expected 1 empty line after import statement not followed by another import.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 1,
        },
      ],
      output:
        "import path from 'path';\nimport foo from 'foo';\n\nvar bar = 42;",
    },
    {
      name: 'run 0 invalid 3',
      code: "import path from 'path';import foo from 'foo';var bar = 42;",
      errors: [
        {
          messageId: 'newlineAfterImport',
          message:
            'Expected 1 empty line after import statement not followed by another import.',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 25,
        },
      ],
      output: "import path from 'path';import foo from 'foo';\n\nvar bar = 42;",
    },
    {
      name: 'run 0 invalid 4',
      code: "import foo from 'foo';\nvar a = 123;\n\nimport { bar } from './bar-lib';\nvar b=456;",
      errors: [
        {
          messageId: 'newlineAfterImport',
          message:
            'Expected 1 empty line after import statement not followed by another import.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 1,
        },
        {
          messageId: 'newlineAfterImport',
          message:
            'Expected 1 empty line after import statement not followed by another import.',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 1,
        },
      ],
      output:
        "import foo from 'foo';\n\nvar a = 123;\n\nimport { bar } from './bar-lib';\n\nvar b=456;",
    },
    {
      name: 'run 0 invalid 5',
      code: "var foo = require('foo-module');\nvar something = 123;",
      errors: [
        {
          messageId: 'newlineAfterRequire',
          message:
            'Expected 1 empty line after require statement not followed by another require.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 1,
        },
      ],
      output: "var foo = require('foo-module');\n\nvar something = 123;",
    },
    {
      name: 'run 0 invalid 6',
      code: "var path = require('path');\nvar foo = require('foo');\nvar bar = 42;",
      errors: [
        {
          messageId: 'newlineAfterRequire',
          message:
            'Expected 1 empty line after require statement not followed by another require.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 1,
        },
      ],
      output:
        "var path = require('path');\nvar foo = require('foo');\n\nvar bar = 42;",
    },
    {
      name: 'run 0 invalid 7',
      code: "var foo = require('foo-module');\nvar a = 123;\n\nvar bar = require('bar-lib');\nvar b=456;",
      errors: [
        {
          messageId: 'newlineAfterRequire',
          message:
            'Expected 1 empty line after require statement not followed by another require.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 1,
        },
        {
          messageId: 'newlineAfterRequire',
          message:
            'Expected 1 empty line after require statement not followed by another require.',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 1,
        },
      ],
      output:
        "var foo = require('foo-module');\n\nvar a = 123;\n\nvar bar = require('bar-lib');\n\nvar b=456;",
    },
    {
      name: 'run 0 invalid 8',
      code: "var assign = Object.assign || require('object-assign');\nvar foo = require('foo');\nvar bar = 42;",
      errors: [
        {
          messageId: 'newlineAfterRequire',
          message:
            'Expected 1 empty line after require statement not followed by another require.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 1,
        },
      ],
      output:
        "var assign = Object.assign || require('object-assign');\nvar foo = require('foo');\n\nvar bar = 42;",
    },
    {
      name: 'run 0 invalid 9',
      code: "require('a');\nfoo(require('b'), require('c'), require('d'));\nrequire('d');\nvar foo = 'bar';",
      errors: [
        {
          messageId: 'newlineAfterRequire',
          message:
            'Expected 1 empty line after require statement not followed by another require.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 1,
        },
      ],
      output:
        "require('a');\nfoo(require('b'), require('c'), require('d'));\nrequire('d');\n\nvar foo = 'bar';",
    },
    {
      name: 'run 0 invalid 10',
      code: "import foo from 'foo';\n\nexport default function() {};",
      options: [
        {
          count: 2,
          exactCount: true,
        },
      ],
      errors: [
        {
          messageId: 'newlineAfterImport',
          message:
            'Expected 2 empty lines after import statement not followed by another import.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 1,
        },
      ],
      output: "import foo from 'foo';\n\n\nexport default function() {};",
    },
    {
      name: 'run 0 invalid 11',
      code: "import foo from 'foo';\n\n\n\nexport default function() {};",
      options: [
        {
          count: 2,
          exactCount: true,
        },
      ],
      errors: [
        {
          messageId: 'newlineAfterImport',
          message:
            'Expected 2 empty lines after import statement not followed by another import.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 1,
        },
      ],
    },
    {
      name: 'run 0 invalid 12',
      code: "import foo from 'foo';export default function() {};",
      options: [
        {
          count: 1,
          exactCount: true,
        },
      ],
      errors: [
        {
          messageId: 'newlineAfterImport',
          message:
            'Expected 1 empty line after import statement not followed by another import.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 1,
        },
      ],
      output: "import foo from 'foo';\n\nexport default function() {};",
    },
    {
      name: 'run 0 invalid 13',
      code: "import path from 'path';\nimport foo from 'foo';\n// Some random single line comment\nvar bar = 42;",
      options: [
        {
          considerComments: true,
          count: 1,
        },
      ],
      errors: [
        {
          messageId: 'newlineAfterImport',
          message:
            'Expected 1 empty line after import statement not followed by another import.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 1,
        },
      ],
      output:
        "import path from 'path';\nimport foo from 'foo';\n\n// Some random single line comment\nvar bar = 42;",
    },
    {
      name: 'run 0 invalid 14',
      code: "var foo = require('foo-module');\n\n/**\n * Test comment\n */\nvar foo = 'bar';",
      options: [
        {
          considerComments: true,
          count: 2,
        },
      ],
      errors: [
        {
          messageId: 'newlineAfterRequire',
          message:
            'Expected 2 empty lines after require statement not followed by another require.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 1,
        },
      ],
      output:
        "var foo = require('foo-module');\n\n\n/**\n * Test comment\n */\nvar foo = 'bar';",
    },
    {
      name: 'run 0 invalid 15',
      code: "import foo from 'foo';\n// some random comment\nexport default function() {};",
      options: [
        {
          count: 2,
          exactCount: true,
          considerComments: true,
        },
      ],
      errors: [
        {
          messageId: 'newlineAfterImport',
          message:
            'Expected 2 empty lines after import statement not followed by another import.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 1,
        },
      ],
      output:
        "import foo from 'foo';\n\n\n// some random comment\nexport default function() {};",
    },
    {
      name: 'run 0 invalid 16',
      code: "import foo from 'foo';// some random comment\nexport default function() {};",
      options: [
        {
          count: 1,
          exactCount: true,
          considerComments: true,
        },
      ],
      errors: [
        {
          messageId: 'newlineAfterImport',
          message:
            'Expected 1 empty line after import statement not followed by another import.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 1,
        },
      ],
      output:
        "import foo from 'foo';\n\n// some random comment\nexport default function() {};",
    },
    {
      name: 'run 0 invalid 17',
      code: "import foo from 'foo';\n@SomeDecorator(foo)\nclass Foo {}",
      errors: [
        {
          messageId: 'newlineAfterImport',
          message:
            'Expected 1 empty line after import statement not followed by another import.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 1,
        },
      ],
      output: "import foo from 'foo';\n\n@SomeDecorator(foo)\nclass Foo {}",
    },
    {
      name: 'run 0 invalid 18',
      code: "import foo from 'foo';\nvar bar = require('bar');\nvar baz = 42;",
      errors: [
        {
          messageId: 'newlineAfterImport',
          message:
            'Expected 1 empty line after import statement not followed by another import.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 1,
        },
        {
          messageId: 'newlineAfterRequire',
          message:
            'Expected 1 empty line after require statement not followed by another require.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 1,
        },
      ],
      output:
        "import foo from 'foo';\n\nvar bar = require('bar');\n\nvar baz = 42;",
    },
  ],
});
