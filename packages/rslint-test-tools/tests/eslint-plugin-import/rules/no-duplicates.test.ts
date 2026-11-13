// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
new RuleTester(fixtureConfig()).run('no-duplicates', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: "import { x } from './foo'; import { y } from './bar'",
    },
    {
      name: 'run 0 valid 1',
      code: 'import foo from "234artaf"; import { shoop } from "234q25ad"',
    },
    {
      name: 'run 0 valid 2',
      code: "import * as ns from './foo'; import {y} from './foo'",
    },
    {
      name: 'run 0 valid 3',
      code: "import {y} from './foo'; import * as ns from './foo'",
    },
    {
      name: 'run 0 valid 4',
      code: "import type { x } from './foo'; import y from './foo'",
    },
    {
      name: 'run 0 valid 5',
      code: "import type x from './foo'; import type y from './bar'",
    },
    {
      name: 'run 0 valid 6',
      code: "import type {x} from './foo'; import type {y} from './bar'",
    },
    {
      name: 'run 0 valid 7',
      code: "import type x from './foo'; import type {y} from './foo'",
    },
    {
      name: 'run 0 valid 8',
      code: "import x from './bar?optionX'; import y from './bar?optionY';",
      options: [
        {
          considerQueryString: true,
        },
      ],
    },
    {
      name: 'run 0 valid 9',
      code: "import { type x } from './foo'; import y from './foo'",
    },
    {
      name: 'run 0 valid 10',
      code: "import { type x } from './foo'; import { y } from './foo'",
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: "import { x } from './foo'; import { y } from './foo'",
      errors: [
        {
          messageId: 'noDuplicates',
          message: "'./foo' imported multiple times.",
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 26,
        },
        {
          messageId: 'noDuplicates',
          message: "'./foo' imported multiple times.",
          line: 1,
          column: 46,
          endLine: 1,
          endColumn: 53,
        },
      ],
      output: "import { x , y } from './foo'; ",
    },
    {
      name: 'run 0 invalid 1',
      code: "import {x} from './foo'; import {y} from './foo'; import { z } from './foo'",
      errors: [
        {
          messageId: 'noDuplicates',
          message: "'./foo' imported multiple times.",
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 24,
        },
        {
          messageId: 'noDuplicates',
          message: "'./foo' imported multiple times.",
          line: 1,
          column: 42,
          endLine: 1,
          endColumn: 49,
        },
        {
          messageId: 'noDuplicates',
          message: "'./foo' imported multiple times.",
          line: 1,
          column: 69,
          endLine: 1,
          endColumn: 76,
        },
      ],
      output: "import {x,y, z } from './foo';  ",
    },
    {
      name: 'run 0 invalid 2',
      code: "import './foo'; import {x} from './foo'",
      errors: [
        {
          messageId: 'noDuplicates',
          message: "'./foo' imported multiple times.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 15,
        },
        {
          messageId: 'noDuplicates',
          message: "'./foo' imported multiple times.",
          line: 1,
          column: 33,
          endLine: 1,
          endColumn: 40,
        },
      ],
      output: "import {x} from './foo'; ",
    },
    {
      name: 'run 0 invalid 3',
      code: "import def from './foo'; import {x} from './foo'",
      errors: [
        {
          messageId: 'noDuplicates',
          message: "'./foo' imported multiple times.",
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 24,
        },
        {
          messageId: 'noDuplicates',
          message: "'./foo' imported multiple times.",
          line: 1,
          column: 42,
          endLine: 1,
          endColumn: 49,
        },
      ],
      output: "import def, {x} from './foo'; ",
    },
    {
      name: 'run 0 invalid 4',
      code: "import './foo'; import './foo'",
      errors: [
        {
          messageId: 'noDuplicates',
          message: "'./foo' imported multiple times.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 15,
        },
        {
          messageId: 'noDuplicates',
          message: "'./foo' imported multiple times.",
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 31,
        },
      ],
      output: "import './foo'; ",
    },
    {
      name: 'run 0 invalid 5',
      code: "import foo from 'non-existent'; import bar from 'non-existent';",
      errors: [
        {
          messageId: 'noDuplicates',
          message: "'non-existent' imported multiple times.",
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 31,
        },
        {
          messageId: 'noDuplicates',
          message: "'non-existent' imported multiple times.",
          line: 1,
          column: 49,
          endLine: 1,
          endColumn: 63,
        },
      ],
    },
    {
      name: 'run 0 invalid 6',
      code: "import type {x} from './foo'; import type {y} from './foo'",
      errors: [
        {
          messageId: 'noDuplicates',
          message: "'./foo' imported multiple times.",
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 29,
        },
        {
          messageId: 'noDuplicates',
          message: "'./foo' imported multiple times.",
          line: 1,
          column: 52,
          endLine: 1,
          endColumn: 59,
        },
      ],
      output: "import type {x,y} from './foo'; ",
    },
    {
      name: 'run 0 invalid 7',
      code: "import {AValue} from './foo'; import type {AType} from './foo'",
      options: [
        {
          'prefer-inline': true,
        },
      ],
      errors: [
        {
          messageId: 'noDuplicates',
          message: "'./foo' imported multiple times.",
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 29,
        },
        {
          messageId: 'noDuplicates',
          message: "'./foo' imported multiple times.",
          line: 1,
          column: 56,
          endLine: 1,
          endColumn: 63,
        },
      ],
      output: "import {AValue,type AType} from './foo'; ",
    },
    {
      name: 'run 0 invalid 8',
      code: "import {type x} from './foo'; import {type y} from './foo'",
      options: [
        {
          'prefer-inline': true,
        },
      ],
      errors: [
        {
          messageId: 'noDuplicates',
          message: "'./foo' imported multiple times.",
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 29,
        },
        {
          messageId: 'noDuplicates',
          message: "'./foo' imported multiple times.",
          line: 1,
          column: 52,
          endLine: 1,
          endColumn: 59,
        },
      ],
      output: "import {type x,type y} from './foo'; ",
    },
    {
      name: 'run 0 invalid 9',
      code: "import * as ns1 from './foo'; import * as ns2 from './foo'",
      errors: [
        {
          messageId: 'noDuplicates',
          message: "'./foo' imported multiple times.",
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 29,
        },
        {
          messageId: 'noDuplicates',
          message: "'./foo' imported multiple times.",
          line: 1,
          column: 52,
          endLine: 1,
          endColumn: 59,
        },
      ],
    },
    {
      name: 'run 0 invalid 10',
      code: "import x from './bar?optionX'; import y from './bar?optionY';",
      errors: [
        {
          messageId: 'noDuplicates',
          message: "'./bar' imported multiple times.",
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 30,
        },
        {
          messageId: 'noDuplicates',
          message: "'./bar' imported multiple times.",
          line: 1,
          column: 46,
          endLine: 1,
          endColumn: 61,
        },
      ],
    },
    {
      name: 'run 0 invalid 11',
      code: "import { Foo } from './foo';\nimport { Bar } from './foo';\nexport const value = {}",
      errors: [
        {
          messageId: 'noDuplicates',
          message: "'./foo' imported multiple times.",
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 28,
        },
        {
          messageId: 'noDuplicates',
          message: "'./foo' imported multiple times.",
          line: 2,
          column: 21,
          endLine: 2,
          endColumn: 28,
        },
      ],
      output: "import { Foo , Bar } from './foo';\nexport const value = {}",
    },
  ],
});
