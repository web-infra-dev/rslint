// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
new RuleTester(fixtureConfig()).run('first', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: "import { x } from './foo'; import { y } from './bar';\nexport { x, y }",
    },
    {
      name: 'run 0 valid 1',
      code: "import a from 'a';\nimport b from 'b';",
    },
    {
      name: 'run 0 valid 2',
      code: "import * as ns from 'foo';\nimport { x } from 'bar';",
    },
    {
      name: 'run 0 valid 3',
      code: "import 'foo';\nimport 'bar';",
    },
    {
      name: 'run 0 valid 4',
      code: "import a from 'a';\nimport { b } from 'b';\nimport * as c from 'c';\nimport 'd';",
    },
    {
      name: 'run 0 valid 5',
      code: "import type { Foo } from './foo';\nimport { bar } from './bar';",
    },
    {
      name: 'run 0 valid 6',
      code: "import y = require('bar');\nimport { x } from 'foo';",
    },
    {
      name: 'run 0 valid 7',
      code: "'use directive';\nimport { x } from 'foo';",
    },
    {
      name: 'run 0 valid 8',
      code: "'use strict';\n'use asm';\nimport { x } from 'foo';",
    },
    {
      name: 'run 0 valid 9',
      code: "import { x } from 'foo'; import { y } from './bar'",
    },
    {
      name: 'run 0 valid 10',
      code: "import { x } from './foo'; import { y } from 'bar'",
    },
    {
      name: 'run 0 valid 11',
      code: "import { x } from './foo'; import { y } from 'bar'",
      options: ['disable-absolute-first'],
    },
    {
      name: 'run 0 valid 12',
      code: "import a from 'a';\nimport b from './b';",
      options: ['absolute-first'],
    },
    {
      name: 'run 0 valid 13',
      code: "import a from '@scope/pkg';\nimport b from './b';",
      options: ['absolute-first'],
    },
    {
      name: 'run 0 valid 14',
      code: '',
    },
    {
      name: 'run 0 valid 15',
      code: 'var a = 1;',
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: "import { x } from './foo';\nexport { x };\nimport { y } from './bar';",
      errors: [
        {
          messageId: 'first',
          message: 'Import in body of module; reorder to top.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 27,
        },
      ],
      output:
        "import { x } from './foo';\nimport { y } from './bar';\nexport { x };",
    },
    {
      name: 'run 0 invalid 1',
      code: "import { x } from './foo';\nexport { x };\nimport { y } from './bar';\nimport { z } from './baz';",
      errors: [
        {
          messageId: 'first',
          message: 'Import in body of module; reorder to top.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 27,
        },
        {
          messageId: 'first',
          message: 'Import in body of module; reorder to top.',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 27,
        },
      ],
      output:
        "import { x } from './foo';\nimport { y } from './bar';\nimport { z } from './baz';\nexport { x };",
    },
    {
      name: 'run 0 invalid 2',
      code: "var a = 1;\nimport { y } from './bar';",
      errors: [
        {
          messageId: 'first',
          message: 'Import in body of module; reorder to top.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 27,
        },
      ],
      output: "import { y } from './bar';\nvar a = 1;",
    },
    {
      name: 'run 0 invalid 3',
      code: "if (true) { console.log(1) }import a from 'b'",
      errors: [
        {
          messageId: 'first',
          message: 'Import in body of module; reorder to top.',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 46,
        },
      ],
      output: "import a from 'b'\nif (true) { console.log(1) }",
    },
    {
      name: 'run 0 invalid 4',
      code: "import { x } from 'foo';\n'use directive';\nimport { y } from 'bar';",
      errors: [
        {
          messageId: 'first',
          message: 'Import in body of module; reorder to top.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 25,
        },
      ],
      output:
        "import { x } from 'foo';\nimport { y } from 'bar';\n'use directive';",
    },
    {
      name: 'run 0 invalid 5',
      code: "var a = 1;\nimport x from './foo';",
      errors: [
        {
          messageId: 'first',
          message: 'Import in body of module; reorder to top.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 23,
        },
      ],
      output: "import x from './foo';\nvar a = 1;",
    },
    {
      name: 'run 0 invalid 6',
      code: "var a = 1;\nimport * as ns from './foo';",
      errors: [
        {
          messageId: 'first',
          message: 'Import in body of module; reorder to top.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 29,
        },
      ],
      output: "import * as ns from './foo';\nvar a = 1;",
    },
    {
      name: 'run 0 invalid 7',
      code: "var a = 1;\nimport './foo';",
      errors: [
        {
          messageId: 'first',
          message: 'Import in body of module; reorder to top.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 16,
        },
      ],
      output: "import './foo';\nvar a = 1;",
    },
    {
      name: 'run 0 invalid 8',
      code: "var a = 1;\nimport type { Foo } from './foo';",
      errors: [
        {
          messageId: 'first',
          message: 'Import in body of module; reorder to top.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 34,
        },
      ],
      output: "import type { Foo } from './foo';\nvar a = 1;",
    },
    {
      name: 'run 0 invalid 9',
      code: "var a = 1;\nimport x = require('./foo');",
      errors: [
        {
          messageId: 'first',
          message: 'Import in body of module; reorder to top.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 29,
        },
      ],
      output: "import x = require('./foo');\nvar a = 1;",
    },
    {
      name: 'run 0 invalid 10',
      code: "import a from 'a';\nvar x = 1;\nvar y = 2;\nimport b from 'b';",
      errors: [
        {
          messageId: 'first',
          message: 'Import in body of module; reorder to top.',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 19,
        },
      ],
      output: "import a from 'a';\nimport b from 'b';\nvar x = 1;\nvar y = 2;",
    },
    {
      name: 'run 0 invalid 11',
      code: "import a from 'a';\nfunction foo() {}\nimport b from 'b';",
      errors: [
        {
          messageId: 'first',
          message: 'Import in body of module; reorder to top.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 19,
        },
      ],
      output: "import a from 'a';\nimport b from 'b';\nfunction foo() {}",
    },
    {
      name: 'run 0 invalid 12',
      code: "import a from 'a';\nvar x = 1;\nimport b from 'b';\nvar y = 2;\nimport c from 'c';",
      errors: [
        {
          messageId: 'first',
          message: 'Import in body of module; reorder to top.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 19,
        },
        {
          messageId: 'first',
          message: 'Import in body of module; reorder to top.',
          line: 5,
          column: 1,
          endLine: 5,
          endColumn: 19,
        },
      ],
      output:
        "import a from 'a';\nimport b from 'b';\nimport c from 'c';\nvar x = 1;\nvar y = 2;",
    },
    {
      name: 'run 0 invalid 13',
      code: "var x = 1;\nimport a from 'a';\nimport b from 'b';",
      errors: [
        {
          messageId: 'first',
          message: 'Import in body of module; reorder to top.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 19,
        },
        {
          messageId: 'first',
          message: 'Import in body of module; reorder to top.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 19,
        },
      ],
      output: "import a from 'a';\nimport b from 'b';\nvar x = 1;",
    },
    {
      name: 'run 0 invalid 14',
      code: "import { x } from './foo'; import { y } from 'bar'",
      options: ['absolute-first'],
      errors: [
        {
          messageId: 'absolute',
          message: 'Absolute imports should come before relative imports.',
          line: 1,
          column: 46,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },
    {
      name: 'run 0 invalid 15',
      code: "import a from './a';\nimport b from 'b';\nimport c from 'c';",
      options: ['absolute-first'],
      errors: [
        {
          messageId: 'absolute',
          message: 'Absolute imports should come before relative imports.',
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 18,
        },
        {
          messageId: 'absolute',
          message: 'Absolute imports should come before relative imports.',
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 18,
        },
      ],
    },
    {
      name: 'run 0 invalid 16',
      code: "import a from './a';\nimport b from '@scope/pkg';",
      options: ['absolute-first'],
      errors: [
        {
          messageId: 'absolute',
          message: 'Absolute imports should come before relative imports.',
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 27,
        },
      ],
    },
    {
      name: 'run 0 invalid 17',
      code: "const a = import('foo');\nimport { x } from 'bar';",
      errors: [
        {
          messageId: 'first',
          message: 'Import in body of module; reorder to top.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 25,
        },
      ],
      output: "import { x } from 'bar';\nconst a = import('foo');",
    },
    {
      name: 'run 0 invalid 18',
      code: "import { x } from 'foo';\nexport { y } from 'bar';\nimport { z } from 'baz';",
      errors: [
        {
          messageId: 'first',
          message: 'Import in body of module; reorder to top.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 25,
        },
      ],
      output:
        "import { x } from 'foo';\nimport { z } from 'baz';\nexport { y } from 'bar';",
    },
  ],
});
