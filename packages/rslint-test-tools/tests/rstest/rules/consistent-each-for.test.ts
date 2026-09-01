import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const methodError = (
  functionName: string,
  preferred: string,
  actual: string,
  line: number,
  column: number,
) => ({
  messageId: 'consistentMethod',
  message: `Prefer using \`${functionName}.${preferred}\` over \`${functionName}.${actual}\``,
  line,
  column,
});

ruleTester.run('consistent-each-for', {} as never, {
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
      errors: [methodError('test', 'each', 'for', 1, 6)],
    },
    {
      code: 'test.skip.for([1, 2, 3])("test", ([n]) => { expect(n).toBeDefined() })',
      options: [{ test: 'each' }],
      errors: [methodError('test', 'each', 'for', 1, 11)],
    },
    {
      code: 'test.only.for([1, 2, 3])("test", ([n]) => { expect(n).toBeDefined() })',
      options: [{ test: 'each' }],
      errors: [methodError('test', 'each', 'for', 1, 11)],
    },
    {
      code: 'test.each([1, 2, 3])("test", (n) => { expect(n).toBeDefined() })',
      options: [{ test: 'for' }],
      errors: [methodError('test', 'for', 'each', 1, 6)],
    },
    {
      code: 'test.skip.each([1, 2, 3])("test", (n) => { expect(n).toBeDefined() })',
      options: [{ test: 'for' }],
      errors: [methodError('test', 'for', 'each', 1, 11)],
    },
    {
      code: 'it.for([1, 2, 3])("test", ([n]) => { expect(n).toBeDefined() })',
      options: [{ it: 'each' }],
      errors: [methodError('it', 'each', 'for', 1, 4)],
    },
    {
      code: 'it.each([1, 2, 3])("test", (n) => { expect(n).toBeDefined() })',
      options: [{ it: 'for' }],
      errors: [methodError('it', 'for', 'each', 1, 4)],
    },
    {
      code: 'describe.for([1, 2, 3])("suite", ([n]) => { test("test", () => {}) })',
      options: [{ describe: 'each' }],
      errors: [methodError('describe', 'each', 'for', 1, 10)],
    },
    {
      code: 'describe.each([1, 2, 3])("suite", (n) => { test("test", () => {}) })',
      options: [{ describe: 'for' }],
      errors: [methodError('describe', 'for', 'each', 1, 10)],
    },
    {
      code: `test.for([1, 2])("test1", ([n]) => {})
test.for([3, 4])("test2", ([n]) => {})`,
      options: [{ test: 'each' }],
      errors: [
        methodError('test', 'each', 'for', 1, 6),
        methodError('test', 'each', 'for', 2, 6),
      ],
    },
  ],
});
