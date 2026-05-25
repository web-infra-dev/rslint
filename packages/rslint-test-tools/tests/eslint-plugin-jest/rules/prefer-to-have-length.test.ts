import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-to-have-length', {} as never, {
  valid: [
    { code: 'expect.hasAssertions' },
    { code: 'expect.hasAssertions()' },
    { code: 'expect(files).toHaveLength(1);' },
    { code: "expect(files.name).toBe('file');" },
    { code: "expect(files[`name`]).toBe('file');" },
    { code: 'expect(users[0]?.permissions?.length).toBe(1);' },
    { code: 'expect(result).toBe(true);' },
    { code: `expect(user.getUserName(5)).resolves.toEqual('Paul')` },
    { code: `expect(user.getUserName(5)).rejects.toEqual('Paul')` },
    { code: 'expect(a);' },
  ],

  invalid: [
    {
      code: 'expect(files["length"]).toBe(1);',
      output: 'expect(files).toHaveLength(1);',
      errors: [{ messageId: 'useToHaveLength', column: 25, line: 1 }],
    },
    {
      code: 'expect(files["length"]).toBe(1,);',
      output: 'expect(files).toHaveLength(1,);',
      errors: [{ messageId: 'useToHaveLength', column: 25, line: 1 }],
    },
    {
      code: 'expect(files["length"])["not"].toBe(1);',
      output: 'expect(files)["not"].toHaveLength(1);',
      errors: [{ messageId: 'useToHaveLength', column: 32, line: 1 }],
    },
    {
      code: 'expect(files["length"])["toBe"](1);',
      output: 'expect(files).toHaveLength(1);',
      errors: [{ messageId: 'useToHaveLength', column: 25, line: 1 }],
    },
    {
      code: 'expect(files["length"]).not["toBe"](1);',
      output: 'expect(files).not.toHaveLength(1);',
      errors: [{ messageId: 'useToHaveLength', column: 29, line: 1 }],
    },
    {
      code: 'expect(files["length"])["not"]["toBe"](1);',
      output: 'expect(files)["not"].toHaveLength(1);',
      errors: [{ messageId: 'useToHaveLength', column: 32, line: 1 }],
    },
    {
      code: 'expect(files.length).toBe(1);',
      output: 'expect(files).toHaveLength(1);',
      errors: [{ messageId: 'useToHaveLength', column: 22, line: 1 }],
    },
    {
      code: 'expect(files.length).toEqual(1);',
      output: 'expect(files).toHaveLength(1);',
      errors: [{ messageId: 'useToHaveLength', column: 22, line: 1 }],
    },
    {
      code: 'expect(files.length).toStrictEqual(1);',
      output: 'expect(files).toHaveLength(1);',
      errors: [{ messageId: 'useToHaveLength', column: 22, line: 1 }],
    },
    {
      code: 'expect(files.length).not.toStrictEqual(1);',
      output: 'expect(files).not.toHaveLength(1);',
      errors: [{ messageId: 'useToHaveLength', column: 26, line: 1 }],
    },
  ],
});
