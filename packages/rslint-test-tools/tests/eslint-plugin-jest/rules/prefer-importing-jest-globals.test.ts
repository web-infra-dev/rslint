import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-importing-jest-globals', {} as never, {
  valid: [
    {
      code: `
        // with import
        import { test, expect } from '@jest/globals';
        test('should pass', () => {
            expect(true).toBeDefined();
        });
      `,
    },
    {
      code: `
        // with import
        import { 'test' as test, expect } from '@jest/globals';
        test('should pass', () => {
            expect(true).toBeDefined();
        });
      `,
    },
    {
      code: `
        test('should pass', () => {
            expect(true).toBeDefined();
        });
      `,
      options: [{ types: ['jest'] }],
    },
    {
      code: `
        const { it } = require('@jest/globals');
        it('should pass', () => {
            expect(true).toBeDefined();
        });
      `,
      options: [{ types: ['test'] }],
    },
    {
      code: `
        // with require
        const { test, expect } = require('@jest/globals');
        test('should pass', () => {
            expect(true).toBeDefined();
        });
      `,
    },
    {
      code: `
        const { test, expect } = require(\`@jest/globals\`);
        test('should pass', () => {
            expect(true).toBeDefined();
        });
      `,
    },
    {
      code: `
        import { it as itChecks } from '@jest/globals';
        itChecks("foo");
      `,
    },
    {
      code: `
        import { 'it' as itChecks } from '@jest/globals';
        itChecks("foo");
      `,
    },
    {
      code: `
        const { test } = require('@jest/globals');
        test("foo");
      `,
    },
    {
      code: `
        const { test } = require('my-test-library');
        test("foo");
      `,
    },
  ],
  invalid: [
    {
      code: `
        import describe from '@jest/globals';
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
      output: `
        import { describe, expect, test } from '@jest/globals';
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
      errors: [
        {
          endColumn: 7,
          column: 3,
          line: 3,
          messageId: 'preferImportingJestGlobal',
        },
      ],
    },
    {
      code: `
        import { describe as context } from '@jest/globals';
        context("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
      output: `
        import { describe as context, expect, test } from '@jest/globals';
        context("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
      errors: [
        {
          endColumn: 7,
          column: 3,
          line: 3,
          messageId: 'preferImportingJestGlobal',
        },
      ],
    },
    {
      code: `
        import { describe as context } from '@jest/globals';
        describe("something", () => {
          context("suite", () => {
            test("foo");
            expect(true).toBeDefined();
          })
        })
      `,
      output: `
        import { describe, describe as context, expect, test } from '@jest/globals';
        describe("something", () => {
          context("suite", () => {
            test("foo");
            expect(true).toBeDefined();
          })
        })
      `,
      errors: [
        {
          endColumn: 9,
          column: 1,
          line: 2,
          messageId: 'preferImportingJestGlobal',
        },
      ],
    },
    {
      code: `
        import { 'describe' as describe } from '@jest/globals';
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
      output: `
        import { 'describe' as describe, expect, test } from '@jest/globals';
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
      errors: [
        {
          endColumn: 7,
          column: 3,
          line: 3,
          messageId: 'preferImportingJestGlobal',
        },
      ],
    },
    {
      code: `
        import { 'describe' as context } from '@jest/globals';
        context("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
      output: `
        import { 'describe' as context, expect, test } from '@jest/globals';
        context("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
      errors: [
        {
          endColumn: 7,
          column: 3,
          line: 3,
          messageId: 'preferImportingJestGlobal',
        },
      ],
    },
    {
      // rslint picks require() when the file has no import/export (no sourceType).
      code: `
        jest.useFakeTimers();
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
      output: `
        const { jest } = require('@jest/globals');
        jest.useFakeTimers();
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
      options: [{ types: ['jest'] }],
      errors: [
        {
          endColumn: 5,
          column: 1,
          line: 1,
          messageId: 'preferImportingJestGlobal',
        },
      ],
    },
    {
      code: `
        import React from 'react';
        import { yourFunction } from './yourFile';
        import something from "something";
        import { test } from '@jest/globals';
        import { xit } from '@jest/globals';
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
      output: `
        import React from 'react';
        import { yourFunction } from './yourFile';
        import something from "something";
        import { describe, expect, test } from '@jest/globals';
        import { xit } from '@jest/globals';
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
      errors: [
        {
          endColumn: 9,
          column: 1,
          line: 6,
          messageId: 'preferImportingJestGlobal',
        },
      ],
    },
    {
      code: `
        console.log('hello');
        import * as fs from 'fs';
        const { test, 'describe': describe } = require('@jest/globals');
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
      output: `
        console.log('hello');
        import * as fs from 'fs';
        import { describe, expect, test } from '@jest/globals';
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
      errors: [
        {
          endColumn: 9,
          column: 3,
          line: 6,
          messageId: 'preferImportingJestGlobal',
        },
      ],
    },
    {
      code: `
        console.log('hello');
        import jestGlobals from '@jest/globals';
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
      output: `
        console.log('hello');
        import { describe, expect, jestGlobals, test } from '@jest/globals';
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
      errors: [
        {
          endColumn: 9,
          column: 1,
          line: 3,
          messageId: 'preferImportingJestGlobal',
        },
      ],
    },
    {
      code: `
        import { pending } from 'actions';
        describe('foo', () => {
          test.each(['hello', 'world'])("%s", (a) => {});
        });
      `,
      output: `
        import { describe, test } from '@jest/globals';
        import { pending } from 'actions';
        describe('foo', () => {
          test.each(['hello', 'world'])("%s", (a) => {});
        });
      `,
      errors: [
        {
          endColumn: 9,
          column: 1,
          line: 2,
          messageId: 'preferImportingJestGlobal',
        },
      ],
    },
    {
      code: `
        const {describe} = require('@jest/globals');
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
      output: `
        const { describe, expect, test } = require('@jest/globals');
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
      errors: [
        {
          endColumn: 7,
          column: 3,
          line: 3,
          messageId: 'preferImportingJestGlobal',
        },
      ],
    },
    {
      code: `
        const {describe: context} = require('@jest/globals');
        context("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
      output: `
        const { describe: context, expect, test } = require('@jest/globals');
        context("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
      errors: [
        {
          endColumn: 7,
          column: 3,
          line: 3,
          messageId: 'preferImportingJestGlobal',
        },
      ],
    },
    {
      code: `
        const {describe: context} = require('@jest/globals');
        describe("something", () => {
          context("suite", () => {
            test("foo");
            expect(true).toBeDefined();
          })
        })
      `,
      output: `
        const { describe, describe: context, expect, test } = require('@jest/globals');
        describe("something", () => {
          context("suite", () => {
            test("foo");
            expect(true).toBeDefined();
          })
        })
      `,
      errors: [
        {
          endColumn: 9,
          column: 1,
          line: 2,
          messageId: 'preferImportingJestGlobal',
        },
      ],
    },
    {
      code: `
        const {describe: []} = require('@jest/globals');
        describe("something", () => {
          context("suite", () => {
            test("foo");
            expect(true).toBeDefined();
          })
        })
      `,
      output: `
        const { describe, expect, test } = require('@jest/globals');
        describe("something", () => {
          context("suite", () => {
            test("foo");
            expect(true).toBeDefined();
          })
        })
      `,
      errors: [
        {
          endColumn: 9,
          column: 1,
          line: 2,
          messageId: 'preferImportingJestGlobal',
        },
      ],
    },
    {
      code: `
        const {describe} = require(\`@jest/globals\`);
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
      // todo: we should really maintain the template literals
      output: `
        const { describe, expect, test } = require('@jest/globals');
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
      errors: [
        {
          endColumn: 7,
          column: 3,
          line: 3,
          messageId: 'preferImportingJestGlobal',
        },
      ],
    },
    {
      code: `
        const source = 'globals';
        const {describe} = require(\`@jest/\${source}\`);
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
      output: `
        const { expect, test } = require('@jest/globals');
        const source = 'globals';
        const {describe} = require(\`@jest/\${source}\`);
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
      errors: [
        {
          endColumn: 7,
          column: 3,
          line: 4,
          messageId: 'preferImportingJestGlobal',
        },
      ],
    },
    {
      code: `
        const { [() => {}]: it } = require('@jest/globals');
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
      output: `
        const { describe, expect, test } = require('@jest/globals');
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
      errors: [
        {
          endColumn: 9,
          column: 1,
          line: 2,
          messageId: 'preferImportingJestGlobal',
        },
      ],
    },
    {
      code: `
        console.log('hello');
        const fs = require('fs');
        const { test, 'describe': describe } = require('@jest/globals');
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
      output: `
        console.log('hello');
        const fs = require('fs');
        const { describe, expect, test } = require('@jest/globals');
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
      errors: [
        {
          endColumn: 9,
          column: 3,
          line: 6,
          messageId: 'preferImportingJestGlobal',
        },
      ],
    },
    {
      code: `
        console.log('hello');
        const jestGlobals = require('@jest/globals');
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
      output: `
        console.log('hello');
        const { describe, expect, test } = require('@jest/globals');
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
      errors: [
        {
          endColumn: 9,
          column: 1,
          line: 3,
          messageId: 'preferImportingJestGlobal',
        },
      ],
    },
    {
      code: `
        const { pending } = require('actions');
        describe('foo', () => {
          test.each(['hello', 'world'])("%s", (a) => {});
        });
      `,
      output: `
        const { describe, test } = require('@jest/globals');
        const { pending } = require('actions');
        describe('foo', () => {
          test.each(['hello', 'world'])("%s", (a) => {});
        });
      `,
      errors: [
        {
          endColumn: 9,
          column: 1,
          line: 2,
          messageId: 'preferImportingJestGlobal',
        },
      ],
    },
    {
      code: `
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
      output: `
        const { describe, expect, test } = require('@jest/globals');
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
      errors: [
        {
          endColumn: 9,
          column: 1,
          line: 1,
          messageId: 'preferImportingJestGlobal',
        },
      ],
    },
    {
      // Shebang must be at byte 0.
      code: '#!/usr/bin/env node\ndescribe("suite", () => {\n  test("foo");\n  expect(true).toBeDefined();\n})\n',
      output:
        '#!/usr/bin/env node\nconst { describe, expect, test } = require(\'@jest/globals\');\ndescribe("suite", () => {\n  test("foo");\n  expect(true).toBeDefined();\n})\n',
      errors: [
        {
          endColumn: 9,
          column: 1,
          line: 2,
          messageId: 'preferImportingJestGlobal',
        },
      ],
    },
    {
      code: `
        // with comment above
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
      output: `
        // with comment above
        const { describe, expect, test } = require('@jest/globals');
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
      errors: [
        {
          endColumn: 9,
          column: 1,
          line: 2,
          messageId: 'preferImportingJestGlobal',
        },
      ],
    },
    {
      code: `
        'use strict';
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
      output: `
        'use strict';
        const { describe, expect, test } = require('@jest/globals');
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
      errors: [
        {
          endColumn: 9,
          column: 1,
          line: 2,
          messageId: 'preferImportingJestGlobal',
        },
      ],
    },
    {
      code: `
        \`use strict\`;
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
      output: `
        \`use strict\`;
        const { describe, expect, test } = require('@jest/globals');
        describe("suite", () => {
          test("foo");
          expect(true).toBeDefined();
        })
      `,
      errors: [
        {
          endColumn: 9,
          column: 1,
          line: 2,
          messageId: 'preferImportingJestGlobal',
        },
      ],
    },
    {
      code: `
        console.log('hello');
        const onClick = jest.fn();
        describe("suite", () => {
          test("foo");
          expect(onClick).toHaveBeenCalled();
        })
      `,
      output: `
        const { describe, expect, jest, test } = require('@jest/globals');
        console.log('hello');
        const onClick = jest.fn();
        describe("suite", () => {
          test("foo");
          expect(onClick).toHaveBeenCalled();
        })
      `,
      errors: [
        {
          endColumn: 21,
          column: 17,
          line: 2,
          messageId: 'preferImportingJestGlobal',
        },
      ],
    },
  ],
});
