import { test } from '../utils.js';

import { RuleTester } from '../rule-tester.js';

const ruleTester = new RuleTester();

ruleTester.run('first', null as never, {
  valid: [
    // Import statement variants
    test({
      code: "import { x } from './foo'; import { y } from './bar';\nexport { x, y }",
    }),
    test({ code: "import a from 'a';\nimport b from 'b';" }),
    test({ code: "import * as ns from 'foo';\nimport { x } from 'bar';" }),
    test({ code: "import 'foo';\nimport 'bar';" }),
    test({
      code: "import a from 'a';\nimport { b } from 'b';\nimport * as c from 'c';\nimport 'd';",
    }),
    test({
      code: "import type { Foo } from './foo';\nimport { bar } from './bar';",
    }),
    test({ code: "import y = require('bar');\nimport { x } from 'foo';" }),

    // Directive handling
    test({ code: "'use directive';\nimport { x } from 'foo';" }),
    test({ code: "'use strict';\n'use asm';\nimport { x } from 'foo';" }),

    // absolute-first option
    test({ code: "import { x } from 'foo'; import { y } from './bar'" }),
    test({ code: "import { x } from './foo'; import { y } from 'bar'" }),
    test({
      code: "import { x } from './foo'; import { y } from 'bar'",
      options: ['disable-absolute-first'],
    }),
    test({
      code: "import a from 'a';\nimport b from './b';",
      options: ['absolute-first'],
    }),
    test({
      code: "import a from '@scope/pkg';\nimport b from './b';",
      options: ['absolute-first'],
    }),

    // Edge cases
    test({ code: '' }),
    test({ code: 'var a = 1;' }),
  ],
  invalid: [
    // Basic misplaced import detection
    test({
      code: "import { x } from './foo';\nexport { x };\nimport { y } from './bar';",
      errors: 1,
    }),
    test({
      code: "import { x } from './foo';\nexport { x };\nimport { y } from './bar';\nimport { z } from './baz';",
      errors: 2,
    }),
    test({
      code: "var a = 1;\nimport { y } from './bar';",
      errors: 1,
    }),
    test({
      code: "if (true) { console.log(1) }import a from 'b'",
      errors: 1,
    }),

    // Directive after first import
    test({
      code: "import { x } from 'foo';\n'use directive';\nimport { y } from 'bar';",
      errors: 1,
    }),

    // Import statement variants (misplaced)
    test({ code: "var a = 1;\nimport x from './foo';", errors: 1 }),
    test({ code: "var a = 1;\nimport * as ns from './foo';", errors: 1 }),
    test({ code: "var a = 1;\nimport './foo';", errors: 1 }),
    test({ code: "var a = 1;\nimport type { Foo } from './foo';", errors: 1 }),
    test({ code: "var a = 1;\nimport x = require('./foo');", errors: 1 }),

    // Interleaved imports and code
    test({
      code: "import a from 'a';\nvar x = 1;\nvar y = 2;\nimport b from 'b';",
      errors: 1,
    }),
    test({
      code: "import a from 'a';\nfunction foo() {}\nimport b from 'b';",
      errors: 1,
    }),
    test({
      code: "import a from 'a';\nvar x = 1;\nimport b from 'b';\nvar y = 2;\nimport c from 'c';",
      errors: 2,
    }),
    test({
      code: "var x = 1;\nimport a from 'a';\nimport b from 'b';",
      errors: 2,
    }),

    // absolute-first option
    test({
      code: "import { x } from './foo'; import { y } from 'bar'",
      options: ['absolute-first'],
      errors: 1,
    }),
    test({
      code: "import a from './a';\nimport b from 'b';\nimport c from 'c';",
      options: ['absolute-first'],
      errors: 2,
    }),
    test({
      code: "import a from './a';\nimport b from '@scope/pkg';",
      options: ['absolute-first'],
      errors: 1,
    }),

    // Dynamic import() is NOT an import statement
    test({
      code: "const a = import('foo');\nimport { x } from 'bar';",
      errors: 1,
    }),
    // Re-export is NOT an import statement
    test({
      code: "import { x } from 'foo';\nexport { y } from 'bar';\nimport { z } from 'baz';",
      errors: 1,
    }),
  ],
});
