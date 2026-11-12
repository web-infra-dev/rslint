import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('consistent-each-for', {
  valid: [
    {
      code: 'test.each([1, 2, 3])("test", (n) => { expect(n).toBeDefined() })',
    },
    {
      code: 'test.for([1, 2, 3])("test", ([n]) => { expect(n).toBeDefined() })',
    },
    {
      code: 'describe.each([1, 2, 3])("suite", (n) => { test("test", () => {}) })',
    },
    {
      code: 'test.each([1, 2, 3])("test", (n) => { expect(n).toBeDefined() })',
      options: [{ test: 'each' }],
    },
    {
      code: 'test.skip.each([1, 2, 3])("test", (n) => { expect(n).toBeDefined() })',
      options: [{ test: 'each' }],
    },
    {
      code: 'test.only.each([1, 2, 3])("test", (n) => { expect(n).toBeDefined() })',
      options: [{ test: 'each' }],
    },
    {
      code: 'test.concurrent.each([1, 2, 3])("test", (n) => { expect(n).toBeDefined() })',
      options: [{ test: 'each' }],
    },
    {
      code: 'test.for([1, 2, 3])("test", ([n]) => { expect(n).toBeDefined() })',
      options: [{ test: 'for' }],
    },
    {
      code: 'test.skip.for([1, 2, 3])("test", ([n]) => { expect(n).toBeDefined() })',
      options: [{ test: 'for' }],
    },
    {
      code: 'test.only.for([1, 2, 3])("test", ([n]) => { expect(n).toBeDefined() })',
      options: [{ test: 'for' }],
    },
    {
      code: 'it.each([1, 2, 3])("test", (n) => { expect(n).toBeDefined() })',
      options: [{ it: 'each' }],
    },
    {
      code: 'it.for([1, 2, 3])("test", ([n]) => { expect(n).toBeDefined() })',
      options: [{ it: 'for' }],
    },
    {
      code: 'describe.each([1, 2, 3])("suite", (n) => { test("test", () => {}) })',
      options: [{ describe: 'each' }],
    },
    {
      code: 'describe.skip.each([1, 2, 3])("suite", (n) => { test("test", () => {}) })',
      options: [{ describe: 'each' }],
    },
    {
      code: 'describe.for([1, 2, 3])("suite", ([n]) => { test("test", () => {}) })',
      options: [{ describe: 'for' }],
    },
    {
      code: `test.each([1, 2, 3])("test", (n) => { expect(n).toBeDefined() })
describe.for([1, 2, 3])("suite", ([n]) => { test("test", () => {}) })`,
      options: [{ test: 'each', describe: 'for' }],
    },
  ],
  invalid: [
    {
      code: 'test.for([1, 2, 3])("test", ([n]) => { expect(n).toBeDefined() })',
      options: [{ test: 'each' }],
      errors: [
        {
          messageId: 'consistentMethod',
          message: 'Prefer using `test.each` over `test.for`',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
    {
      code: 'test.skip.for([1, 2, 3])("test", ([n]) => { expect(n).toBeDefined() })',
      options: [{ test: 'each' }],
      errors: [
        {
          messageId: 'consistentMethod',
          message: 'Prefer using `test.each` over `test.for`',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'test.only.for([1, 2, 3])("test", ([n]) => { expect(n).toBeDefined() })',
      options: [{ test: 'each' }],
      errors: [
        {
          messageId: 'consistentMethod',
          message: 'Prefer using `test.each` over `test.for`',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'test.each([1, 2, 3])("test", (n) => { expect(n).toBeDefined() })',
      options: [{ test: 'for' }],
      errors: [
        {
          messageId: 'consistentMethod',
          message: 'Prefer using `test.for` over `test.each`',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: 'test.skip.each([1, 2, 3])("test", (n) => { expect(n).toBeDefined() })',
      options: [{ test: 'for' }],
      errors: [
        {
          messageId: 'consistentMethod',
          message: 'Prefer using `test.for` over `test.each`',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: 'it.for([1, 2, 3])("test", ([n]) => { expect(n).toBeDefined() })',
      options: [{ it: 'each' }],
      errors: [
        {
          messageId: 'consistentMethod',
          message: 'Prefer using `it.each` over `it.for`',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 7,
        },
      ],
    },
    {
      code: 'it.each([1, 2, 3])("test", (n) => { expect(n).toBeDefined() })',
      options: [{ it: 'for' }],
      errors: [
        {
          messageId: 'consistentMethod',
          message: 'Prefer using `it.for` over `it.each`',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: 'describe.for([1, 2, 3])("suite", ([n]) => { test("test", () => {}) })',
      options: [{ describe: 'each' }],
      errors: [
        {
          messageId: 'consistentMethod',
          message: 'Prefer using `describe.each` over `describe.for`',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: 'describe.each([1, 2, 3])("suite", (n) => { test("test", () => {}) })',
      options: [{ describe: 'for' }],
      errors: [
        {
          messageId: 'consistentMethod',
          message: 'Prefer using `describe.for` over `describe.each`',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: `test.for([1, 2])("test1", ([n]) => {})
test.for([3, 4])("test2", ([n]) => {})`,
      options: [{ test: 'each' }],
      errors: [
        {
          messageId: 'consistentMethod',
          message: 'Prefer using `test.each` over `test.for`',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 9,
        },
        {
          messageId: 'consistentMethod',
          message: 'Prefer using `test.each` over `test.for`',
          line: 2,
          column: 6,
          endLine: 2,
          endColumn: 9,
        },
      ],
    },
  ],
});
