import { RuleTester } from '../rule-tester.js';

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/max-dependencies.js
// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/max-dependencies.md
// Flow and TypeScript parser variants use the suite's TypeScript parser.
const ruleTester = new RuleTester();

ruleTester.run('max-dependencies', null as never, {
  valid: [
    // max-dependencies.
    {
      code: 'import "./foo.js"',
    },
    {
      code: 'import "./foo.js"; import "./bar.js";',
      options: [
        {
          max: 2,
        },
      ],
    },
    {
      code: 'import "./foo.js"; import "./bar.js"; const a = require("./foo.js"); const b = require("./bar.js");',
      options: [
        {
          max: 2,
        },
      ],
    },
    {
      code: 'import {x, y, z} from "./foo"',
    },
    // max-dependencies (typescript).
    {
      code: "import type { x } from './foo'; import { y } from './bar';",
      options: [
        {
          max: 1,
          ignoreTypeImports: true,
        },
      ],
    },
    // Documentation examples.
    {
      code: "import a from './a'; // 1\nconst anotherA = require('./a'); // still 1\nimport {x, y, z} from './foo'; // 2",
      options: [
        {
          max: 2,
        },
      ],
    },
    {
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
    // max-dependencies.
    {
      code: "import { x } from './foo'; import { y } from './foo'; import {z} from './bar';",
      options: [
        {
          max: 1,
        },
      ],
      errors: ['Maximum number of dependencies (1) exceeded.'],
    },
    {
      code: "import { x } from './foo'; import { y } from './bar'; import { z } from './baz';",
      options: [
        {
          max: 2,
        },
      ],
      errors: ['Maximum number of dependencies (2) exceeded.'],
    },
    {
      code: "import { x } from './foo'; require(\"./bar\"); import { z } from './baz';",
      options: [
        {
          max: 2,
        },
      ],
      errors: ['Maximum number of dependencies (2) exceeded.'],
    },
    {
      code: 'import { x } from \'./foo\'; import { z } from \'./foo\'; require("./bar"); const path = require("path");',
      options: [
        {
          max: 2,
        },
      ],
      errors: ['Maximum number of dependencies (2) exceeded.'],
    },
    {
      code: "import type { x } from './foo'; import type { y } from './bar'",
      options: [
        {
          max: 1,
        },
      ],
      errors: ['Maximum number of dependencies (1) exceeded.'],
    },
    {
      code: "import type { x } from './foo'; import type { y } from './bar'; import type { z } from './baz'",
      options: [
        {
          max: 2,
          ignoreTypeImports: false,
        },
      ],
      errors: ['Maximum number of dependencies (2) exceeded.'],
    },
    // max-dependencies (typescript).
    {
      code: "import type { x } from './foo'; import type { y } from './bar'",
      options: [
        {
          max: 1,
        },
      ],
      errors: ['Maximum number of dependencies (1) exceeded.'],
    },
    {
      code: "import type { x } from './foo'; import type { y } from './bar'; import type { z } from './baz'",
      options: [
        {
          max: 2,
          ignoreTypeImports: false,
        },
      ],
      errors: ['Maximum number of dependencies (2) exceeded.'],
    },
    // Documentation examples.
    {
      code: "import a from './a'; // 1\nconst b = require('./b'); // 2\nimport c from './c'; // 3 - exceeds max!",
      options: [
        {
          max: 2,
        },
      ],
      errors: ['Maximum number of dependencies (2) exceeded.'],
    },
    {
      code: "import a from './a';\nimport b from './b';\nimport c from './c';",
      options: [
        {
          max: 2,
          ignoreTypeImports: true,
        },
      ],
      errors: ['Maximum number of dependencies (2) exceeded.'],
    },
  ],
});
