import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-to-be', {} as never, {
  valid: [
    { code: 'expect(null).toBeNull();' },
    { code: 'expect(null).not.toBeNull();' },
    { code: 'expect(null).toBe(1);' },
    { code: 'expect(null).toBe(-1);' },
    { code: 'expect(null).toBe(...1);' },
    { code: 'expect(obj).toStrictEqual([ x, 1 ]);' },
    { code: 'expect(obj).toStrictEqual({ x: 1 });' },
    { code: 'expect(obj).not.toStrictEqual({ x: 1 });' },
    { code: 'expect(value).toMatchSnapshot();' },
    { code: "expect(catchError()).toStrictEqual({ message: 'oh noes!' })" },
    { code: 'expect("something");' },
    { code: 'expect(token).toStrictEqual(/[abc]+/g);' },
    { code: "expect(token).toStrictEqual(new RegExp('[abc]+', 'g'));" },
    { code: 'expect(value).toEqual(dedent`my string`);' },

    // null
    { code: 'expect(null).not.toEqual();' },
    { code: 'expect(null).toBe();' },
    { code: 'expect(null).toMatchSnapshot();' },
    { code: 'expect("a string").toMatchSnapshot(null);' },
    { code: 'expect("a string").not.toMatchSnapshot();' },
    { code: 'expect(null).toBe' },

    // undefined
    { code: 'expect(undefined).toBeUndefined();' },
    { code: 'expect(true).toBeDefined();' },

    // NaN
    { code: 'expect(NaN).toBeNaN();' },
    { code: 'expect(true).not.toBeNaN();' },
    { code: 'expect({}).toEqual({});' },
    { code: 'expect(something).toBe()' },
    { code: 'expect(something).toBe(somethingElse)' },
    { code: 'expect(something).toEqual(somethingElse)' },
    { code: 'expect(something).not.toBe(somethingElse)' },
    { code: 'expect(something).not.toEqual(somethingElse)' },
    { code: 'expect(undefined).toBe' },

    // typescript edition
    {
      code: "(expect('Model must be bound to an array if the multiple property is true') as any).toHaveBeenTipped()",
    },
  ],
  invalid: [
    {
      code: 'expect(value).toEqual("my string");',
      output: 'expect(value).toBe("my string");',
      errors: [{ messageId: 'useToBe', column: 15, line: 1 }],
    },
    {
      code: 'expect(value).toStrictEqual("my string");',
      output: 'expect(value).toBe("my string");',
      errors: [{ messageId: 'useToBe', column: 15, line: 1 }],
    },
    {
      code: 'expect(value).toStrictEqual(1);',
      output: 'expect(value).toBe(1);',
      errors: [{ messageId: 'useToBe', column: 15, line: 1 }],
    },
    {
      code: 'expect(value).toStrictEqual(1,);',
      output: 'expect(value).toBe(1,);',
      errors: [{ messageId: 'useToBe', column: 15, line: 1 }],
    },
    {
      code: 'expect(value).toStrictEqual(-1);',
      output: 'expect(value).toBe(-1);',
      errors: [{ messageId: 'useToBe', column: 15, line: 1 }],
    },
    {
      code: 'expect(value).toEqual(`my string`);',
      output: 'expect(value).toBe(`my string`);',
      errors: [{ messageId: 'useToBe', column: 15, line: 1 }],
    },
    {
      code: 'expect(value)["toEqual"](`my string`);',
      output: "expect(value)['toBe'](`my string`);",
      errors: [{ messageId: 'useToBe', column: 15, line: 1 }],
    },
    {
      code: 'expect(value).toStrictEqual(`my ${string}`);',
      output: 'expect(value).toBe(`my ${string}`);',
      errors: [{ messageId: 'useToBe', column: 15, line: 1 }],
    },
    {
      code: 'expect(loadMessage()).resolves.toStrictEqual("hello world");',
      output: 'expect(loadMessage()).resolves.toBe("hello world");',
      errors: [{ messageId: 'useToBe', column: 32, line: 1 }],
    },
    {
      code: 'expect(loadMessage()).resolves["toStrictEqual"]("hello world");',
      output: 'expect(loadMessage()).resolves[\'toBe\']("hello world");',
      errors: [{ messageId: 'useToBe', column: 32, line: 1 }],
    },
    {
      code: 'expect(loadMessage())["resolves"].toStrictEqual("hello world");',
      output: 'expect(loadMessage())["resolves"].toBe("hello world");',
      errors: [{ messageId: 'useToBe', column: 35, line: 1 }],
    },
    {
      code: 'expect(loadMessage()).resolves.toStrictEqual(false);',
      output: 'expect(loadMessage()).resolves.toBe(false);',
      errors: [{ messageId: 'useToBe', column: 32, line: 1 }],
    },

    // null
    {
      code: 'expect(null).toBe(null);',
      output: 'expect(null).toBeNull();',
      errors: [{ messageId: 'useToBeNull', column: 14, line: 1 }],
    },
    {
      code: 'expect(null).toEqual(null);',
      output: 'expect(null).toBeNull();',
      errors: [{ messageId: 'useToBeNull', column: 14, line: 1 }],
    },
    {
      code: 'expect(null).toEqual(null,);',
      output: 'expect(null).toBeNull();',
      errors: [{ messageId: 'useToBeNull', column: 14, line: 1 }],
    },
    {
      code: 'expect(null).toStrictEqual(null);',
      output: 'expect(null).toBeNull();',
      errors: [{ messageId: 'useToBeNull', column: 14, line: 1 }],
    },
    {
      code: 'expect("a string").not.toBe(null);',
      output: 'expect("a string").not.toBeNull();',
      errors: [{ messageId: 'useToBeNull', column: 24, line: 1 }],
    },
    {
      code: 'expect("a string").not["toBe"](null);',
      output: 'expect("a string").not[\'toBeNull\']();',
      errors: [{ messageId: 'useToBeNull', column: 24, line: 1 }],
    },
    {
      code: 'expect("a string")["not"]["toBe"](null);',
      output: 'expect("a string")["not"][\'toBeNull\']();',
      errors: [{ messageId: 'useToBeNull', column: 27, line: 1 }],
    },
    {
      code: 'expect("a string").not.toEqual(null);',
      output: 'expect("a string").not.toBeNull();',
      errors: [{ messageId: 'useToBeNull', column: 24, line: 1 }],
    },
    {
      code: 'expect("a string").not.toStrictEqual(null);',
      output: 'expect("a string").not.toBeNull();',
      errors: [{ messageId: 'useToBeNull', column: 24, line: 1 }],
    },

    // undefined
    {
      code: 'expect(undefined).toBe(undefined);',
      output: 'expect(undefined).toBeUndefined();',
      errors: [{ messageId: 'useToBeUndefined', column: 19, line: 1 }],
    },
    {
      code: 'expect(undefined).toEqual(undefined);',
      output: 'expect(undefined).toBeUndefined();',
      errors: [{ messageId: 'useToBeUndefined', column: 19, line: 1 }],
    },
    {
      code: 'expect(undefined).toStrictEqual(undefined);',
      output: 'expect(undefined).toBeUndefined();',
      errors: [{ messageId: 'useToBeUndefined', column: 19, line: 1 }],
    },
    {
      code: 'expect("a string").not.toBe(undefined);',
      output: 'expect("a string").toBeDefined();',
      errors: [{ messageId: 'useToBeDefined', column: 24, line: 1 }],
    },
    {
      code: 'expect("a string").rejects.not.toBe(undefined);',
      output: 'expect("a string").rejects.toBeDefined();',
      errors: [{ messageId: 'useToBeDefined', column: 32, line: 1 }],
    },
    {
      code: 'expect("a string").rejects.not["toBe"](undefined);',
      output: 'expect("a string").rejects[\'toBeDefined\']();',
      errors: [{ messageId: 'useToBeDefined', column: 32, line: 1 }],
    },
    {
      code: 'expect("a string").not.toEqual(undefined);',
      output: 'expect("a string").toBeDefined();',
      errors: [{ messageId: 'useToBeDefined', column: 24, line: 1 }],
    },
    {
      code: 'expect("a string").not.toStrictEqual(undefined);',
      output: 'expect("a string").toBeDefined();',
      errors: [{ messageId: 'useToBeDefined', column: 24, line: 1 }],
    },

    // NaN
    {
      code: 'expect(NaN).toBe(NaN);',
      output: 'expect(NaN).toBeNaN();',
      errors: [{ messageId: 'useToBeNaN', column: 13, line: 1 }],
    },
    {
      code: 'expect(NaN).toEqual(NaN);',
      output: 'expect(NaN).toBeNaN();',
      errors: [{ messageId: 'useToBeNaN', column: 13, line: 1 }],
    },
    {
      code: 'expect(NaN).toStrictEqual(NaN);',
      output: 'expect(NaN).toBeNaN();',
      errors: [{ messageId: 'useToBeNaN', column: 13, line: 1 }],
    },
    {
      code: 'expect("a string").not.toBe(NaN);',
      output: 'expect("a string").not.toBeNaN();',
      errors: [{ messageId: 'useToBeNaN', column: 24, line: 1 }],
    },
    {
      code: 'expect("a string").rejects.not.toBe(NaN);',
      output: 'expect("a string").rejects.not.toBeNaN();',
      errors: [{ messageId: 'useToBeNaN', column: 32, line: 1 }],
    },
    {
      code: 'expect("a string")["rejects"].not.toBe(NaN);',
      output: 'expect("a string")["rejects"].not.toBeNaN();',
      errors: [{ messageId: 'useToBeNaN', column: 35, line: 1 }],
    },
    {
      code: 'expect("a string").not.toEqual(NaN);',
      output: 'expect("a string").not.toBeNaN();',
      errors: [{ messageId: 'useToBeNaN', column: 24, line: 1 }],
    },
    {
      code: 'expect("a string").not.toStrictEqual(NaN);',
      output: 'expect("a string").not.toBeNaN();',
      errors: [{ messageId: 'useToBeNaN', column: 24, line: 1 }],
    },

    // undefined vs defined
    {
      code: 'expect(undefined).not.toBeDefined();',
      output: 'expect(undefined).toBeUndefined();',
      errors: [{ messageId: 'useToBeUndefined', column: 23, line: 1 }],
    },
    {
      code: 'expect(undefined).resolves.not.toBeDefined();',
      output: 'expect(undefined).resolves.toBeUndefined();',
      errors: [{ messageId: 'useToBeUndefined', column: 32, line: 1 }],
    },
    {
      code: 'expect(undefined).resolves.toBe(undefined);',
      output: 'expect(undefined).resolves.toBeUndefined();',
      errors: [{ messageId: 'useToBeUndefined', column: 28, line: 1 }],
    },
    {
      code: 'expect("a string").not.toBeUndefined();',
      output: 'expect("a string").toBeDefined();',
      errors: [{ messageId: 'useToBeDefined', column: 24, line: 1 }],
    },
    {
      code: 'expect("a string").rejects.not.toBeUndefined();',
      output: 'expect("a string").rejects.toBeDefined();',
      errors: [{ messageId: 'useToBeDefined', column: 32, line: 1 }],
    },

    // typescript edition
    {
      code: 'expect(null).toEqual(1 as unknown as string as unknown as any);',
      output: 'expect(null).toBe(1 as unknown as string as unknown as any);',
      errors: [{ messageId: 'useToBe', column: 14, line: 1 }],
    },
    {
      code: 'expect(null).toEqual(-1 as unknown as string as unknown as any);',
      output: 'expect(null).toBe(-1 as unknown as string as unknown as any);',
      errors: [{ messageId: 'useToBe', column: 14, line: 1 }],
    },
    {
      code: 'expect("a string").not.toStrictEqual("string" as number);',
      output: 'expect("a string").not.toBe("string" as number);',
      errors: [{ messageId: 'useToBe', column: 24, line: 1 }],
    },
    {
      code: 'expect(null).toBe(null as unknown as string as unknown as any);',
      output: 'expect(null).toBeNull();',
      errors: [{ messageId: 'useToBeNull', column: 14, line: 1 }],
    },
    {
      code: 'expect("a string").not.toEqual(null as number);',
      output: 'expect("a string").not.toBeNull();',
      errors: [{ messageId: 'useToBeNull', column: 24, line: 1 }],
    },
    {
      code: 'expect(undefined).toBe(undefined as unknown as string as any);',
      output: 'expect(undefined).toBeUndefined();',
      errors: [{ messageId: 'useToBeUndefined', column: 19, line: 1 }],
    },
    {
      code: 'expect("a string").toEqual(undefined as number);',
      output: 'expect("a string").toBeUndefined();',
      errors: [{ messageId: 'useToBeUndefined', column: 20, line: 1 }],
    },
  ],
});
