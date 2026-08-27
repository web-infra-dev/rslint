import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-expect-type-of', {} as never, {
  valid: [
    { code: `expect("name").toBeTypeOf("string")` },
    { code: `expect("name").not.toBeTypeOf("string")` },
    { code: `expect(12).toBeTypeOf("number")` },
    { code: `expect(true).toBeTypeOf("boolean")` },
    { code: `expect({a: 1}).toBeTypeOf("object")` },
    { code: `expect(() => {}).toBeTypeOf("function")` },
    { code: `expect(sym).toBeTypeOf("symbol")` },
    { code: `expect(BigInt(123)).toBeTypeOf("bigint")` },
    { code: `expect(undefined).toBeTypeOf("undefined")` },
    { code: `expect(value).not.toBe(42)` },
    { code: `expect(value).not.toEqual(42)` },
  ],
  invalid: [
    {
      code: `expect(typeof 12).toBe("number")`,
      output: `expect(12).toBeTypeOf("number")`,
      errors: [
        {
          messageId: 'preferExpectTypeOf',
          message: `Use \`expect(12).toBeTypeOf("number")\` instead of \`expect(typeof 12).toBe("number")\``,
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: `expect(typeof "name").toBe("string")`,
      output: `expect("name").toBeTypeOf("string")`,
      errors: [
        {
          messageId: 'preferExpectTypeOf',
          message: `Use \`expect("name").toBeTypeOf("string")\` instead of \`expect(typeof "name").toBe("string")\``,
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: `expect(typeof true).toBe("boolean")`,
      output: `expect(true).toBeTypeOf("boolean")`,
      errors: [
        {
          messageId: 'preferExpectTypeOf',
          message: `Use \`expect(true).toBeTypeOf("boolean")\` instead of \`expect(typeof true).toBe("boolean")\``,
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      code: `expect(typeof variable).toBe("object")`,
      output: `expect(variable).toBeTypeOf("object")`,
      errors: [
        {
          messageId: 'preferExpectTypeOf',
          message: `Use \`expect(variable).toBeTypeOf("object")\` instead of \`expect(typeof variable).toBe("object")\``,
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      code: `expect(typeof fn).toBe("function")`,
      output: `expect(fn).toBeTypeOf("function")`,
      errors: [
        {
          messageId: 'preferExpectTypeOf',
          message: `Use \`expect(fn).toBeTypeOf("function")\` instead of \`expect(typeof fn).toBe("function")\``,
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: `expect(typeof sym).toBe("symbol")`,
      output: `expect(sym).toBeTypeOf("symbol")`,
      errors: [
        {
          messageId: 'preferExpectTypeOf',
          message: `Use \`expect(sym).toBeTypeOf("symbol")\` instead of \`expect(typeof sym).toBe("symbol")\``,
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: `expect(typeof big).toBe("bigint")`,
      output: `expect(big).toBeTypeOf("bigint")`,
      errors: [
        {
          messageId: 'preferExpectTypeOf',
          message: `Use \`expect(big).toBeTypeOf("bigint")\` instead of \`expect(typeof big).toBe("bigint")\``,
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: `expect(typeof value).toBe("undefined")`,
      output: `expect(value).toBeTypeOf("undefined")`,
      errors: [
        {
          messageId: 'preferExpectTypeOf',
          message: `Use \`expect(value).toBeTypeOf("undefined")\` instead of \`expect(typeof value).toBe("undefined")\``,
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      code: `expect(typeof value).toEqual("string")`,
      output: `expect(value).toBeTypeOf("string")`,
      errors: [
        {
          messageId: 'preferExpectTypeOf',
          message: `Use \`expect(value).toBeTypeOf("string")\` instead of \`expect(typeof value).toBe("string")\``,
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      code: `expect(typeof value).not.toBe("string")`,
      output: `expect(value).not.toBeTypeOf("string")`,
      errors: [
        {
          messageId: 'preferExpectTypeOf',
          message: `Use \`expect(value).toBeTypeOf("string")\` instead of \`expect(typeof value).toBe("string")\``,
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      code: `expect(typeof value).toBe("unknown")`,
      output: `expect(value).toBeTypeOf("unknown")`,
      errors: [
        {
          messageId: 'preferExpectTypeOf',
          message: `Use \`expect(value).toBeTypeOf("unknown")\` instead of \`expect(typeof value).toBe("unknown")\``,
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: `expect(typeof value).toBe(typeName)`,
      output: `expect(value).toBeTypeOf(typeName)`,
      errors: [
        {
          messageId: 'preferExpectTypeOf',
          message: `Use \`expect(value).toBeTypeOf(typeName)\` instead of \`expect(typeof value).toBe(typeName)\``,
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
  ],
});
