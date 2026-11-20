// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

// Shared source templates; reporting ranges remain explicit in each matrix.
const testSetup = `        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
     \u0020`;
const testSetup2 = `const { describe, expect, test } = require('@jest/globals');
describe("suite", () => {
  test("foo");
  expect(true).toBeDefined();
})`;

new RuleTester(fixtureConfig()).run('prefer-importing-jest-globals', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: `
        // with import
        import { test, expect } from '@jest/globals';
        test('should pass', () => {
            expect(true).toBeDefined();
        });
     \u0020`,
    },
    {
      name: 'run 0 valid 1',
      code: `
        // with import
        import { 'test' as test, expect } from '@jest/globals';
        test('should pass', () => {
            expect(true).toBeDefined();
        });
     \u0020`,
    },
    {
      name: 'run 0 valid 2',
      code: `
        test('should pass', () => {
            expect(true).toBeDefined();
        });
     \u0020`,
      options: [
        {
          types: ['jest'],
        },
      ],
    },
    {
      name: 'run 0 valid 3',
      code: `
        const { it } = require('@jest/globals');
        it('should pass', () => {
            expect(true).toBeDefined();
        });
     \u0020`,
      options: [
        {
          types: ['test'],
        },
      ],
    },
    ...(
      [
        `
        // with require
        const { test, expect } = require('@jest/globals');
        test('should pass', () => {
            expect(true).toBeDefined();
        });
     \u0020`,
        `
        const { test, expect } = require(\`@jest/globals\`);
        test('should pass', () => {
            expect(true).toBeDefined();
        });
     \u0020`,
        `
        import { it as itChecks } from '@jest/globals';
        itChecks("foo");
     \u0020`,
        `
        import { 'it' as itChecks } from '@jest/globals';
        itChecks("foo");
     \u0020`,
        `
        const { test } = require('@jest/globals');
        test("foo");
     \u0020`,
        `
        const { test } = require('my-test-library');
        test("foo");
     \u0020`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index + 4}`,
      code,
    })),
    {
      name: 'upstream prefer-importing-jest-globals / valid 0',
      code: `// with import
import { test, expect } from '@jest/globals';
test('should pass', () => {
    expect(true).toBeDefined();
});`,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2015,
      },
    },
    {
      name: 'upstream prefer-importing-jest-globals / valid 1',
      code: `// with import
import { 'test' as test, expect } from '@jest/globals';
test('should pass', () => {
    expect(true).toBeDefined();
});`,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2022,
      },
    },
    {
      name: 'upstream prefer-importing-jest-globals / valid 2',
      code: `test('should pass', () => {
    expect(true).toBeDefined();
});`,
      options: [
        {
          types: ['jest'],
        },
      ],
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2015,
      },
    },
    {
      name: 'upstream prefer-importing-jest-globals / valid 3',
      code: `const { it } = require('@jest/globals');
it('should pass', () => {
    expect(true).toBeDefined();
});`,
      options: [
        {
          types: ['test'],
        },
      ],
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2015,
      },
    },
    ...(
      [
        `// with require
const { test, expect } = require('@jest/globals');
test('should pass', () => {
    expect(true).toBeDefined();
});`,
        `const { test, expect } = require(\`@jest/globals\`);
test('should pass', () => {
    expect(true).toBeDefined();
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream prefer-importing-jest-globals / valid ${index + 4}`,
      code,
      languageOptions: {
        sourceType: 'script',
        ecmaVersion: 2015,
      },
    })),
    {
      name: 'upstream prefer-importing-jest-globals / valid 6',
      code: `import { it as itChecks } from '@jest/globals';
itChecks("foo");`,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2015,
      },
    },
    {
      name: 'upstream prefer-importing-jest-globals / valid 7',
      code: `import { 'it' as itChecks } from '@jest/globals';
itChecks("foo");`,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2022,
      },
    },
    ...(
      [
        `const { test } = require('@jest/globals');
test("foo");`,
        `const { test } = require('my-test-library');
test("foo");`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream prefer-importing-jest-globals / valid ${index + 8}`,
      code,
      languageOptions: {
        sourceType: 'script',
        ecmaVersion: 2015,
      },
    })),
  ],
  invalid: [
    ...(
      [
        [
          `
        import describe from '@jest/globals';
` + testSetup,
          "Import the following Jest functions from '@jest/globals': test, expect",
          4,
          11,
          15,
          `
        import { describe, expect, test } from '@jest/globals';
` + testSetup,
        ],
        [
          `
        import { describe as context } from '@jest/globals';
        context("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
     \u0020`,
          "Import the following Jest functions from '@jest/globals': test, expect",
          4,
          11,
          15,
          `
        import { describe as context, expect, test } from '@jest/globals';
        context("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
     \u0020`,
        ],
        [
          `
        import { describe as context } from '@jest/globals';
        describe("something", () => {
          context("suite", () => {
            test("foo");
            expect(true).toBeDefined();
          })
        })
     \u0020`,
          "Import the following Jest functions from '@jest/globals': describe, test, expect",
          3,
          9,
          17,
          `
        import { describe, describe as context, expect, test } from '@jest/globals';
        describe("something", () => {
          context("suite", () => {
            test("foo");
            expect(true).toBeDefined();
          })
        })
     \u0020`,
        ],
        [
          `
        import { 'describe' as describe } from '@jest/globals';
` + testSetup,
          "Import the following Jest functions from '@jest/globals': test, expect",
          4,
          11,
          15,
          `
        import { 'describe' as describe, expect, test } from '@jest/globals';
` + testSetup,
        ],
        [
          `
        import { 'describe' as context } from '@jest/globals';
        context("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
     \u0020`,
          "Import the following Jest functions from '@jest/globals': test, expect",
          4,
          11,
          15,
          `
        import { 'describe' as context, expect, test } from '@jest/globals';
        context("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
     \u0020`,
        ],
      ] as const
    ).map(
      ([code, message, line, column, endColumn, output], index): TestCase => ({
        name: `run 0 invalid ${index}`,
        code,
        errors: [
          {
            messageId: 'preferImportingJestGlobal',
            message,
            line,
            column,
            endLine: line,
            endColumn,
          },
        ],
        output,
      }),
    ),
    // Original upstream output: "\n        const { jest } = require('@jest/globals');\n        jest.useFakeTimers();\n        describe(\"suite\", () => {\n          test(\"foo\");\n          expect(true).toBeDefined();\n        })\n      "
    {
      name: 'run 0 invalid 5',
      code:
        `
        jest.useFakeTimers();
` + testSetup,
      options: [
        {
          types: ['jest'],
        },
      ],
      errors: [
        {
          messageId: 'preferImportingJestGlobal',
          message:
            "Import the following Jest functions from '@jest/globals': jest",
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 13,
        },
      ],
      output:
        `
        import { jest } from '@jest/globals';
        jest.useFakeTimers();
` + testSetup,
    },
    ...(
      [
        [
          `
        import React from 'react';
        import { yourFunction } from './yourFile';
        import something from "something";
        import { test } from '@jest/globals';
        import { xit } from '@jest/globals';
` + testSetup,
          "Import the following Jest functions from '@jest/globals': describe, expect",
          7,
          9,
          17,
          `
        import React from 'react';
        import { yourFunction } from './yourFile';
        import something from "something";
        import { describe, expect, test } from '@jest/globals';
        import { xit } from '@jest/globals';
` + testSetup,
        ],
        [
          `
        console.log('hello');
        import * as fs from 'fs';
        const { test, 'describe': describe } = require('@jest/globals');
` + testSetup,
          "Import the following Jest functions from '@jest/globals': expect",
          7,
          11,
          17,
          `
        console.log('hello');
        import * as fs from 'fs';
        import { describe, expect, test } from '@jest/globals';
` + testSetup,
        ],
        [
          `
        console.log('hello');
        import jestGlobals from '@jest/globals';
` + testSetup,
          "Import the following Jest functions from '@jest/globals': describe, test, expect",
          4,
          9,
          17,
          `
        console.log('hello');
        import { describe, expect, jestGlobals, test } from '@jest/globals';
` + testSetup,
        ],
        [
          `
        import { pending } from 'actions';
        describe('foo', () => {
          test.each(['hello', 'world'])("%s", (a) => {});
        });
     \u0020`,
          "Import the following Jest functions from '@jest/globals': describe, test",
          3,
          9,
          17,
          `
        import { describe, test } from '@jest/globals';
        import { pending } from 'actions';
        describe('foo', () => {
          test.each(['hello', 'world'])("%s", (a) => {});
        });
     \u0020`,
        ],
        // Original upstream output: "\n        const { describe, expect, test } = require('@jest/globals');\n        describe(\"suite\", () => {\n          test(\"foo\");\n          expect(true).toBeDefined();\n        })\n      "
        [
          `
        const {describe} = require('@jest/globals');
` + testSetup,
          "Import the following Jest functions from '@jest/globals': test, expect",
          4,
          11,
          15,
          `
        import { describe, expect, test } from '@jest/globals';
` + testSetup,
        ],
        // Known existing defect: this module-mode autofix emits invalid ESM alias syntax.
        // Original upstream output: "\n        const { describe: context, expect, test } = require('@jest/globals');\n        context(\"suite\", () => {\n          test(\"foo\");\n          expect(true).toBeDefined();\n        })\n      "
        [
          `
        const {describe: context} = require('@jest/globals');
        context("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
     \u0020`,
          "Import the following Jest functions from '@jest/globals': test, expect",
          4,
          11,
          15,
          `
        import { describe: context, expect, test } from '@jest/globals';
        context("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
     \u0020`,
        ],
        // Known existing defect: this module-mode autofix emits invalid ESM alias syntax.
        // Original upstream output: "\n        const { describe, describe: context, expect, test } = require('@jest/globals');\n        describe(\"something\", () => {\n          context(\"suite\", () => {\n            test(\"foo\");\n            expect(true).toBeDefined();\n          })\n        })\n      "
        [
          `
        const {describe: context} = require('@jest/globals');
        describe("something", () => {
          context("suite", () => {
            test("foo");
            expect(true).toBeDefined();
          })
        })
     \u0020`,
          "Import the following Jest functions from '@jest/globals': describe, test, expect",
          3,
          9,
          17,
          `
        import { describe, describe: context, expect, test } from '@jest/globals';
        describe("something", () => {
          context("suite", () => {
            test("foo");
            expect(true).toBeDefined();
          })
        })
     \u0020`,
        ],
        // Original upstream output: "\n        const { describe, expect, test } = require('@jest/globals');\n        describe(\"something\", () => {\n          context(\"suite\", () => {\n            test(\"foo\");\n            expect(true).toBeDefined();\n          })\n        })\n      "
        [
          `
        const {describe: []} = require('@jest/globals');
        describe("something", () => {
          context("suite", () => {
            test("foo");
            expect(true).toBeDefined();
          })
        })
     \u0020`,
          "Import the following Jest functions from '@jest/globals': describe, test, expect",
          3,
          9,
          17,
          `
        import { describe, expect, test } from '@jest/globals';
        describe("something", () => {
          context("suite", () => {
            test("foo");
            expect(true).toBeDefined();
          })
        })
     \u0020`,
        ],
        // Original upstream output: "\n        const { describe, expect, test } = require('@jest/globals');\n        describe(\"suite\", () => {\n          test(\"foo\");\n          expect(true).toBeDefined();\n        })\n      "
        [
          `
        const {describe} = require(\`@jest/globals\`);
` + testSetup,
          "Import the following Jest functions from '@jest/globals': test, expect",
          4,
          11,
          15,
          `
        import { describe, expect, test } from '@jest/globals';
` + testSetup,
        ],
        // Original upstream output: "\n        const { expect, test } = require('@jest/globals');\n        const source = 'globals';\n        const {describe} = require(`@jest/${source}`);\n        describe(\"suite\", () => {\n          test(\"foo\");\n          expect(true).toBeDefined();\n        })\n      "
        [
          `
        const source = 'globals';
        const {describe} = require(\`@jest/\${source}\`);
` + testSetup,
          "Import the following Jest functions from '@jest/globals': test, expect",
          5,
          11,
          15,
          `
        import { expect, test } from '@jest/globals';
        const source = 'globals';
        const {describe} = require(\`@jest/\${source}\`);
` + testSetup,
        ],
        // Original upstream output: "\n        const { describe, expect, test } = require('@jest/globals');\n        describe(\"suite\", () => {\n          test(\"foo\");\n          expect(true).toBeDefined();\n        })\n      "
        [
          `
        const { [() => {}]: it } = require('@jest/globals');
` + testSetup,
          "Import the following Jest functions from '@jest/globals': describe, test, expect",
          3,
          9,
          17,
          `
        import { describe, expect, test } from '@jest/globals';
` + testSetup,
        ],
        // Original upstream output: "\n        console.log('hello');\n        const fs = require('fs');\n        const { describe, expect, test } = require('@jest/globals');\n        describe(\"suite\", () => {\n          test(\"foo\");\n          expect(true).toBeDefined();\n        })\n      "
        [
          `
        console.log('hello');
        const fs = require('fs');
        const { test, 'describe': describe } = require('@jest/globals');
` + testSetup,
          "Import the following Jest functions from '@jest/globals': expect",
          7,
          11,
          17,
          `
        console.log('hello');
        const fs = require('fs');
        import { describe, expect, test } from '@jest/globals';
` + testSetup,
        ],
        // Original upstream output: "\n        console.log('hello');\n        const { describe, expect, test } = require('@jest/globals');\n        describe(\"suite\", () => {\n          test(\"foo\");\n          expect(true).toBeDefined();\n        })\n      "
        [
          `
        console.log('hello');
        const jestGlobals = require('@jest/globals');
` + testSetup,
          "Import the following Jest functions from '@jest/globals': describe, test, expect",
          4,
          9,
          17,
          `
        console.log('hello');
        import { describe, expect, test } from '@jest/globals';
` + testSetup,
        ],
        // Original upstream output: "\n        const { describe, test } = require('@jest/globals');\n        const { pending } = require('actions');\n        describe('foo', () => {\n          test.each(['hello', 'world'])(\"%s\", (a) => {});\n        });\n      "
        [
          `
        const { pending } = require('actions');
        describe('foo', () => {
          test.each(['hello', 'world'])("%s", (a) => {});
        });
     \u0020`,
          "Import the following Jest functions from '@jest/globals': describe, test",
          3,
          9,
          17,
          `
        import { describe, test } from '@jest/globals';
        const { pending } = require('actions');
        describe('foo', () => {
          test.each(['hello', 'world'])("%s", (a) => {});
        });
     \u0020`,
        ],
        // Original upstream output: "\n        const { describe, expect, test } = require('@jest/globals');\n        describe(\"suite\", () => {\n          test(\"foo\");\n          expect(true).toBeDefined();\n        })\n      "
        [
          `
` + testSetup,
          "Import the following Jest functions from '@jest/globals': describe, test, expect",
          2,
          9,
          17,
          `
        import { describe, expect, test } from '@jest/globals';
` + testSetup,
        ],
        // Original upstream output: "#!/usr/bin/env node\nconst { describe, expect, test } = require('@jest/globals');\ndescribe(\"suite\", () => {\n  test(\"foo\");\n  expect(true).toBeDefined();\n})\n"
        [
          `#!/usr/bin/env node
describe("suite", () => {
  test("foo");
  expect(true).toBeDefined();
})
`,
          "Import the following Jest functions from '@jest/globals': describe, test, expect",
          2,
          1,
          9,
          `#!/usr/bin/env node
import { describe, expect, test } from '@jest/globals';
describe("suite", () => {
  test("foo");
  expect(true).toBeDefined();
})
`,
        ],
        // Original upstream output: "\n        // with comment above\n        const { describe, expect, test } = require('@jest/globals');\n        describe(\"suite\", () => {\n          test(\"foo\");\n          expect(true).toBeDefined();\n        })\n      "
        [
          `
        // with comment above
` + testSetup,
          "Import the following Jest functions from '@jest/globals': describe, test, expect",
          3,
          9,
          17,
          `
        // with comment above
        import { describe, expect, test } from '@jest/globals';
` + testSetup,
        ],
        // Original upstream output: "\n        'use strict';\n        const { describe, expect, test } = require('@jest/globals');\n        describe(\"suite\", () => {\n          test(\"foo\");\n          expect(true).toBeDefined();\n        })\n      "
        [
          `
        'use strict';
` + testSetup,
          "Import the following Jest functions from '@jest/globals': describe, test, expect",
          3,
          9,
          17,
          `
        'use strict';
        import { describe, expect, test } from '@jest/globals';
` + testSetup,
        ],
        // Original upstream output: "\n        `use strict`;\n        const { describe, expect, test } = require('@jest/globals');\n        describe(\"suite\", () => {\n          test(\"foo\");\n          expect(true).toBeDefined();\n        })\n      "
        [
          `
        \`use strict\`;
` + testSetup,
          "Import the following Jest functions from '@jest/globals': describe, test, expect",
          3,
          9,
          17,
          `
        \`use strict\`;
        import { describe, expect, test } from '@jest/globals';
` + testSetup,
        ],
        // Original upstream output: "\n        const { describe, expect, jest, test } = require('@jest/globals');\n        console.log('hello');\n        const onClick = jest.fn();\n        describe(\"suite\", () => {\n          test(\"foo\");\n          expect(onClick).toHaveBeenCalled();\n        })\n      "
        [
          `
        console.log('hello');
        const onClick = jest.fn();
        describe("suite", () => {
          test("foo");
          expect(onClick).toHaveBeenCalled();
        })
     \u0020`,
          "Import the following Jest functions from '@jest/globals': jest, describe, test, expect",
          3,
          25,
          29,
          `
        import { describe, expect, jest, test } from '@jest/globals';
        console.log('hello');
        const onClick = jest.fn();
        describe("suite", () => {
          test("foo");
          expect(onClick).toHaveBeenCalled();
        })
     \u0020`,
        ],
      ] as const
    ).map(
      ([code, message, line, column, endColumn, output], index): TestCase => ({
        name: `run 0 invalid ${index + 6}`,
        code,
        errors: [
          {
            messageId: 'preferImportingJestGlobal',
            message,
            line,
            column,
            endLine: line,
            endColumn,
          },
        ],
        output,
      }),
    ),
    ...(
      [
        [
          `import describe from '@jest/globals';
describe("suite", () => {
  test("foo");
  expect(true).toBeDefined();
})`,
          "Import the following Jest functions from '@jest/globals': test, expect",
          3,
          3,
          7,
          `import { describe, expect, test } from '@jest/globals';
describe("suite", () => {
  test("foo");
  expect(true).toBeDefined();
})`,
        ],
        [
          `import { describe as context } from '@jest/globals';
context("suite", () => {
  test("foo");
  expect(true).toBeDefined();
})`,
          "Import the following Jest functions from '@jest/globals': test, expect",
          3,
          3,
          7,
          `import { describe as context, expect, test } from '@jest/globals';
context("suite", () => {
  test("foo");
  expect(true).toBeDefined();
})`,
        ],
        [
          `import { describe as context } from '@jest/globals';
describe("something", () => {
  context("suite", () => {
    test("foo");
    expect(true).toBeDefined();
  })
})`,
          "Import the following Jest functions from '@jest/globals': describe, test, expect",
          2,
          1,
          9,
          `import { describe, describe as context, expect, test } from '@jest/globals';
describe("something", () => {
  context("suite", () => {
    test("foo");
    expect(true).toBeDefined();
  })
})`,
        ],
      ] as const
    ).map(
      ([code, message, line, column, endColumn, output], index): TestCase => ({
        name: `upstream prefer-importing-jest-globals / invalid ${index}`,
        code,
        languageOptions: {
          sourceType: 'module',
          ecmaVersion: 2015,
        },
        errors: [
          {
            messageId: 'preferImportingJestGlobal',
            message,
            line,
            column,
            endLine: line,
            endColumn,
          },
        ],
        output,
      }),
    ),
    ...(
      [
        [
          `import { 'describe' as describe } from '@jest/globals';
describe("suite", () => {
  test("foo");
  expect(true).toBeDefined();
})`,
          `import { 'describe' as describe, expect, test } from '@jest/globals';
describe("suite", () => {
  test("foo");
  expect(true).toBeDefined();
})`,
        ],
        [
          `import { 'describe' as context } from '@jest/globals';
context("suite", () => {
  test("foo");
  expect(true).toBeDefined();
})`,
          `import { 'describe' as context, expect, test } from '@jest/globals';
context("suite", () => {
  test("foo");
  expect(true).toBeDefined();
})`,
        ],
      ] as const
    ).map(([code, output], index): TestCase => ({
      name: `upstream prefer-importing-jest-globals / invalid ${index + 3}`,
      code,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2022,
      },
      errors: [
        {
          messageId: 'preferImportingJestGlobal',
          message:
            "Import the following Jest functions from '@jest/globals': test, expect",
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 7,
        },
      ],
      output,
    })),
    {
      name: 'upstream prefer-importing-jest-globals / invalid 5',
      code: `jest.useFakeTimers();
describe("suite", () => {
  test("foo");
  expect(true).toBeDefined();
})`,
      options: [
        {
          types: ['jest'],
        },
      ],
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'preferImportingJestGlobal',
          message:
            "Import the following Jest functions from '@jest/globals': jest",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 5,
        },
      ],
      output: `import { jest } from '@jest/globals';
jest.useFakeTimers();
describe("suite", () => {
  test("foo");
  expect(true).toBeDefined();
})`,
    },
    ...(
      [
        [
          `import React from 'react';
import { yourFunction } from './yourFile';
import something from "something";
import { test } from '@jest/globals';
import { xit } from '@jest/globals';
describe("suite", () => {
  test("foo");
  expect(true).toBeDefined();
})`,
          "Import the following Jest functions from '@jest/globals': describe, expect",
          6,
          1,
          `import React from 'react';
import { yourFunction } from './yourFile';
import something from "something";
import { describe, expect, test } from '@jest/globals';
import { xit } from '@jest/globals';
describe("suite", () => {
  test("foo");
  expect(true).toBeDefined();
})`,
        ],
        [
          `console.log('hello');
import * as fs from 'fs';
const { test, 'describe': describe } = require('@jest/globals');
describe("suite", () => {
  test("foo");
  expect(true).toBeDefined();
})`,
          "Import the following Jest functions from '@jest/globals': expect",
          6,
          3,
          `console.log('hello');
import * as fs from 'fs';
import { describe, expect, test } from '@jest/globals';
describe("suite", () => {
  test("foo");
  expect(true).toBeDefined();
})`,
        ],
        [
          `console.log('hello');
import jestGlobals from '@jest/globals';
describe("suite", () => {
  test("foo");
  expect(true).toBeDefined();
})`,
          "Import the following Jest functions from '@jest/globals': describe, test, expect",
          3,
          1,
          `console.log('hello');
import { describe, expect, jestGlobals, test } from '@jest/globals';
describe("suite", () => {
  test("foo");
  expect(true).toBeDefined();
})`,
        ],
        [
          `import { pending } from 'actions';
describe('foo', () => {
  test.each(['hello', 'world'])("%s", (a) => {});
});`,
          "Import the following Jest functions from '@jest/globals': describe, test",
          2,
          1,
          `import { describe, test } from '@jest/globals';
import { pending } from 'actions';
describe('foo', () => {
  test.each(['hello', 'world'])("%s", (a) => {});
});`,
        ],
      ] as const
    ).map(([code, message, line, column, output], index): TestCase => ({
      name: `upstream prefer-importing-jest-globals / invalid ${index + 6}`,
      code,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'preferImportingJestGlobal',
          message,
          line,
          column,
          endLine: line,
          endColumn: 9,
        },
      ],
      output,
    })),
    ...(
      [
        [
          `const {describe} = require('@jest/globals');
describe("suite", () => {
  test("foo");
  expect(true).toBeDefined();
})`,
          "Import the following Jest functions from '@jest/globals': test, expect",
          3,
          3,
          7,
          testSetup2,
        ],
        [
          `const {describe: context} = require('@jest/globals');
context("suite", () => {
  test("foo");
  expect(true).toBeDefined();
})`,
          "Import the following Jest functions from '@jest/globals': test, expect",
          3,
          3,
          7,
          `const { describe: context, expect, test } = require('@jest/globals');
context("suite", () => {
  test("foo");
  expect(true).toBeDefined();
})`,
        ],
        [
          `const {describe: context} = require('@jest/globals');
describe("something", () => {
  context("suite", () => {
    test("foo");
    expect(true).toBeDefined();
  })
})`,
          "Import the following Jest functions from '@jest/globals': describe, test, expect",
          2,
          1,
          9,
          `const { describe, describe: context, expect, test } = require('@jest/globals');
describe("something", () => {
  context("suite", () => {
    test("foo");
    expect(true).toBeDefined();
  })
})`,
        ],
        [
          `const {describe: []} = require('@jest/globals');
describe("something", () => {
  context("suite", () => {
    test("foo");
    expect(true).toBeDefined();
  })
})`,
          "Import the following Jest functions from '@jest/globals': describe, test, expect",
          2,
          1,
          9,
          `const { describe, expect, test } = require('@jest/globals');
describe("something", () => {
  context("suite", () => {
    test("foo");
    expect(true).toBeDefined();
  })
})`,
        ],
        [
          `const {describe} = require(\`@jest/globals\`);
describe("suite", () => {
  test("foo");
  expect(true).toBeDefined();
})`,
          "Import the following Jest functions from '@jest/globals': test, expect",
          3,
          3,
          7,
          testSetup2,
        ],
        [
          `const source = 'globals';
const {describe} = require(\`@jest/\${source}\`);
describe("suite", () => {
  test("foo");
  expect(true).toBeDefined();
})`,
          "Import the following Jest functions from '@jest/globals': test, expect",
          4,
          3,
          7,
          `const { expect, test } = require('@jest/globals');
const source = 'globals';
const {describe} = require(\`@jest/\${source}\`);
describe("suite", () => {
  test("foo");
  expect(true).toBeDefined();
})`,
        ],
        [
          `const { [() => {}]: it } = require('@jest/globals');
describe("suite", () => {
  test("foo");
  expect(true).toBeDefined();
})`,
          "Import the following Jest functions from '@jest/globals': describe, test, expect",
          2,
          1,
          9,
          testSetup2,
        ],
        [
          `console.log('hello');
const fs = require('fs');
const { test, 'describe': describe } = require('@jest/globals');
describe("suite", () => {
  test("foo");
  expect(true).toBeDefined();
})`,
          "Import the following Jest functions from '@jest/globals': expect",
          6,
          3,
          9,
          `console.log('hello');
const fs = require('fs');
` + testSetup2,
        ],
        [
          `console.log('hello');
const jestGlobals = require('@jest/globals');
describe("suite", () => {
  test("foo");
  expect(true).toBeDefined();
})`,
          "Import the following Jest functions from '@jest/globals': describe, test, expect",
          3,
          1,
          9,
          `console.log('hello');
` + testSetup2,
        ],
        [
          `const { pending } = require('actions');
describe('foo', () => {
  test.each(['hello', 'world'])("%s", (a) => {});
});`,
          "Import the following Jest functions from '@jest/globals': describe, test",
          2,
          1,
          9,
          `const { describe, test } = require('@jest/globals');
const { pending } = require('actions');
describe('foo', () => {
  test.each(['hello', 'world'])("%s", (a) => {});
});`,
        ],
        [
          `describe("suite", () => {
  test("foo");
  expect(true).toBeDefined();
})`,
          "Import the following Jest functions from '@jest/globals': describe, test, expect",
          1,
          1,
          9,
          testSetup2,
        ],
        [
          `#!/usr/bin/env node
describe("suite", () => {
  test("foo");
  expect(true).toBeDefined();
})`,
          "Import the following Jest functions from '@jest/globals': describe, test, expect",
          2,
          1,
          9,
          `#!/usr/bin/env node
` + testSetup2,
        ],
        [
          `// with comment above
describe("suite", () => {
  test("foo");
  expect(true).toBeDefined();
})`,
          "Import the following Jest functions from '@jest/globals': describe, test, expect",
          2,
          1,
          9,
          `// with comment above
` + testSetup2,
        ],
        [
          `'use strict';
describe("suite", () => {
  test("foo");
  expect(true).toBeDefined();
})`,
          "Import the following Jest functions from '@jest/globals': describe, test, expect",
          2,
          1,
          9,
          `'use strict';
` + testSetup2,
        ],
        [
          `\`use strict\`;
describe("suite", () => {
  test("foo");
  expect(true).toBeDefined();
})`,
          "Import the following Jest functions from '@jest/globals': describe, test, expect",
          2,
          1,
          9,
          `\`use strict\`;
` + testSetup2,
        ],
        [
          `console.log('hello');
const onClick = jest.fn();
describe("suite", () => {
  test("foo");
  expect(onClick).toHaveBeenCalled();
})`,
          "Import the following Jest functions from '@jest/globals': jest, describe, test, expect",
          2,
          17,
          21,
          `const { describe, expect, jest, test } = require('@jest/globals');
console.log('hello');
const onClick = jest.fn();
describe("suite", () => {
  test("foo");
  expect(onClick).toHaveBeenCalled();
})`,
        ],
      ] as const
    ).map(
      ([code, message, line, column, endColumn, output], index): TestCase => ({
        name: `upstream prefer-importing-jest-globals / invalid ${index + 10}`,
        code,
        languageOptions: {
          sourceType: 'script',
          ecmaVersion: 2015,
        },
        errors: [
          {
            messageId: 'preferImportingJestGlobal',
            message,
            line,
            column,
            endLine: line,
            endColumn,
          },
        ],
        output,
      }),
    ),
    {
      name: 'upstream prefer-importing-jest-globals / invalid 26',
      code: `console.log('hello');
const onClick = jest.fn();
describe("suite", () => {
  test("foo");
  expect(onClick).toHaveBeenCalled();
})`,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'preferImportingJestGlobal',
          message:
            "Import the following Jest functions from '@jest/globals': jest, describe, test, expect",
          line: 2,
          column: 17,
          endLine: 2,
          endColumn: 21,
        },
      ],
      output: `import { describe, expect, jest, test } from '@jest/globals';
console.log('hello');
const onClick = jest.fn();
describe("suite", () => {
  test("foo");
  expect(onClick).toHaveBeenCalled();
})`,
    },
    {
      name: 'upstream prefer-importing-jest-globals: typescript edition / invalid 0',
      code: `import describe from '@jest/globals';
describe("suite", () => {
  test("foo");
  expect(true).toBeDefined();
})`,
      languageOptions: {
        sourceType: 'module',
      },
      errors: [
        {
          messageId: 'preferImportingJestGlobal',
          message:
            "Import the following Jest functions from '@jest/globals': test, expect",
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 7,
        },
      ],
      output: `import { describe, expect, test } from '@jest/globals';
describe("suite", () => {
  test("foo");
  expect(true).toBeDefined();
})`,
    },
    {
      name: 'upstream prefer-importing-jest-globals: typescript edition / invalid 1',
      code: `const {describe} = require('@jest/globals');
describe("suite", () => {
  test("foo");
  expect(true).toBeDefined();
})`,
      languageOptions: {
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'preferImportingJestGlobal',
          message:
            "Import the following Jest functions from '@jest/globals': test, expect",
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 7,
        },
      ],
      output: testSetup2,
    },
  ],
});
