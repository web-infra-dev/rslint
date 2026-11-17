// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
new RuleTester(fixtureConfig()).run('max-dependencies', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: 'import "./foo.js"',
    },
    {
      name: 'run 0 valid 1',
      code: 'import "./foo.js"; import "./bar.js";',
      options: [
        {
          max: 2,
        },
      ],
    },
    {
      name: 'run 0 valid 2',
      code: 'import "./foo.js"; import "./bar.js"; const a = require("./foo.js"); const b = require("./bar.js");',
      options: [
        {
          max: 2,
        },
      ],
    },
    {
      name: 'run 0 valid 3',
      code: 'import {x, y, z} from "./foo"',
    },
    {
      name: 'run 0 valid 4',
      code: "import type { x } from './foo'; import { y } from './bar';",
      options: [
        {
          max: 1,
          ignoreTypeImports: true,
        },
      ],
    },
    {
      name: 'run 0 valid 5',
      code: "import a from './a'; // 1\nconst anotherA = require('./a'); // still 1\nimport {x, y, z} from './foo'; // 2",
      options: [
        {
          max: 2,
        },
      ],
    },
    {
      name: 'run 0 valid 6',
      code: "import a from './a';\nimport b from './b';\nimport type c from './c'; // Doesn't count against max",
      options: [
        {
          max: 2,
          ignoreTypeImports: true,
        },
      ],
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: "import { x } from './foo'; import { y } from './foo'; import {z} from './bar';",
      options: [
        {
          max: 1,
        },
      ],
      errors: [
        {
          messageId: '',
          message: 'Maximum number of dependencies (1) exceeded.',
          line: 1,
          column: 71,
          endLine: 1,
          endColumn: 78,
        },
      ],
    },
    {
      name: 'run 0 invalid 1',
      code: "import { x } from './foo'; import { y } from './bar'; import { z } from './baz';",
      options: [
        {
          max: 2,
        },
      ],
      errors: [
        {
          messageId: '',
          message: 'Maximum number of dependencies (2) exceeded.',
          line: 1,
          column: 73,
          endLine: 1,
          endColumn: 80,
        },
      ],
    },
    {
      name: 'run 0 invalid 2',
      code: "import { x } from './foo'; require(\"./bar\"); import { z } from './baz';",
      options: [
        {
          max: 2,
        },
      ],
      errors: [
        {
          messageId: '',
          message: 'Maximum number of dependencies (2) exceeded.',
          line: 1,
          column: 64,
          endLine: 1,
          endColumn: 71,
        },
      ],
    },
    {
      name: 'run 0 invalid 3',
      code: 'import { x } from \'./foo\'; import { z } from \'./foo\'; require("./bar"); const path = require("path");',
      options: [
        {
          max: 2,
        },
      ],
      errors: [
        {
          messageId: '',
          message: 'Maximum number of dependencies (2) exceeded.',
          line: 1,
          column: 94,
          endLine: 1,
          endColumn: 100,
        },
      ],
    },
    {
      name: 'run 0 invalid 4',
      code: "import type { x } from './foo'; import type { y } from './bar'",
      options: [
        {
          max: 1,
        },
      ],
      errors: [
        {
          messageId: '',
          message: 'Maximum number of dependencies (1) exceeded.',
          line: 1,
          column: 56,
          endLine: 1,
          endColumn: 63,
        },
      ],
    },
    {
      name: 'run 0 invalid 5',
      code: "import type { x } from './foo'; import type { y } from './bar'; import type { z } from './baz'",
      options: [
        {
          max: 2,
          ignoreTypeImports: false,
        },
      ],
      errors: [
        {
          messageId: '',
          message: 'Maximum number of dependencies (2) exceeded.',
          line: 1,
          column: 88,
          endLine: 1,
          endColumn: 95,
        },
      ],
    },
    {
      name: 'run 0 invalid 6',
      code: "import type { x } from './foo'; import type { y } from './bar'",
      options: [
        {
          max: 1,
        },
      ],
      errors: [
        {
          messageId: '',
          message: 'Maximum number of dependencies (1) exceeded.',
          line: 1,
          column: 56,
          endLine: 1,
          endColumn: 63,
        },
      ],
    },
    {
      name: 'run 0 invalid 7',
      code: "import type { x } from './foo'; import type { y } from './bar'; import type { z } from './baz'",
      options: [
        {
          max: 2,
          ignoreTypeImports: false,
        },
      ],
      errors: [
        {
          messageId: '',
          message: 'Maximum number of dependencies (2) exceeded.',
          line: 1,
          column: 88,
          endLine: 1,
          endColumn: 95,
        },
      ],
    },
    {
      name: 'run 0 invalid 8',
      code: "import a from './a'; // 1\nconst b = require('./b'); // 2\nimport c from './c'; // 3 - exceeds max!",
      options: [
        {
          max: 2,
        },
      ],
      errors: [
        {
          messageId: '',
          message: 'Maximum number of dependencies (2) exceeded.',
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 20,
        },
      ],
    },
    {
      name: 'run 0 invalid 9',
      code: "import a from './a';\nimport b from './b';\nimport c from './c';",
      options: [
        {
          max: 2,
          ignoreTypeImports: true,
        },
      ],
      errors: [
        {
          messageId: '',
          message: 'Maximum number of dependencies (2) exceeded.',
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 20,
        },
      ],
    },
  ],
});
