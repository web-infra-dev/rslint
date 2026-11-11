import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('prefer-to-be', {
  valid: [
    { code: `expect(null).toBeNull();` },
    { code: `expect(null).not.toBeNull();` },
    { code: `expect(null).toBe(-1);` },
    { code: `expect(null).toBe(1);` },
    { code: `expect(obj).toStrictEqual([ x, 1 ]);` },
    { code: `expect(obj).toStrictEqual({ x: 1 });` },
    { code: `expect(obj).not.toStrictEqual({ x: 1 });` },
    { code: `expect(value).toMatchSnapshot();` },
    { code: `expect(catchError()).toStrictEqual({ message: 'oh noes!' })` },
    { code: `expect("something");` },
    { code: `expect("hey").to.be.a("string");` },
    { code: `expect(token).toStrictEqual(/[abc]+/g);` },
    { code: `expect(token).toStrictEqual(new RegExp('[abc]+', 'g'));` },
    { code: `expect(0.1 + 0.2).toEqual(0.3);` },
    { code: `expect(NaN).toBeNaN();` },
    { code: `expect(true).not.toBeNaN();` },
    { code: `expect({}).toEqual({});` },
    { code: `expect(something).toBe()` },
    { code: `expect(something).toBe(somethingElse)` },
    { code: `expect(something).toEqual(somethingElse)` },
    { code: `expect(something).not.toBe(somethingElse)` },
    { code: `expect(something).not.toEqual(somethingElse)` },
    { code: `expect(undefined).toBe` },
    { code: `expect("something");` },
  ],
  invalid: [
    {
      code: `expect(value).toEqual("my string");`,
      output: `expect(value).toBe("my string");`,
      errors: [
        {
          messageId: 'useToBe',
          message: 'Use `toBe` instead',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: `expect("a string").not.toEqual(null);`,
      output: `expect("a string").not.toBeNull();`,
      errors: [
        {
          messageId: 'useToBeNull',
          message: 'Use `toBeNull()` instead',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: `expect("a string").not.toStrictEqual(null);`,
      output: `expect("a string").not.toBeNull();`,
      errors: [
        {
          messageId: 'useToBeNull',
          message: 'Use `toBeNull()` instead',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: `expect(NaN).toBe(NaN);`,
      output: `expect(NaN).toBeNaN();`,
      errors: [
        {
          messageId: 'useToBeNaN',
          message: 'Use `toBeNaN()` instead',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: `expect("a string").not.toBe(NaN);`,
      output: `expect("a string").not.toBeNaN();`,
      errors: [
        {
          messageId: 'useToBeNaN',
          message: 'Use `toBeNaN()` instead',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: `expect("a string").not.toStrictEqual(NaN);`,
      output: `expect("a string").not.toBeNaN();`,
      errors: [
        {
          messageId: 'useToBeNaN',
          message: 'Use `toBeNaN()` instead',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: `expect(null).toBe(null);`,
      output: `expect(null).toBeNull();`,
      errors: [
        {
          messageId: 'useToBeNull',
          message: 'Use `toBeNull()` instead',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: `expect(null).toEqual(null);`,
      output: `expect(null).toBeNull();`,
      errors: [
        {
          messageId: 'useToBeNull',
          message: 'Use `toBeNull()` instead',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: `expect("a string").not.toEqual(null as number);`,
      output: `expect("a string").not.toBeNull();`,
      errors: [
        {
          messageId: 'useToBeNull',
          message: 'Use `toBeNull()` instead',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: `expect(undefined).toBe(undefined as unknown as string as any);`,
      output: `expect(undefined).toBeUndefined();`,
      errors: [
        {
          messageId: 'useToBeUndefined',
          message: 'Use `toBeUndefined()` instead',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: `expect("a string").toEqual(undefined as number);`,
      output: `expect("a string").toBeUndefined();`,
      errors: [
        {
          messageId: 'useToBeUndefined',
          message: 'Use `toBeUndefined()` instead',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
  ],
});
