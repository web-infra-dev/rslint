import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-interpolation-in-snapshots', {} as never, {
  valid: [
    {
      code: 'expect(something).toMatchInlineSnapshot(`No interpolation`);',
    },
    {
      // toMatchFileSnapshot takes a path, so interpolation is intended
      code: 'test("case", async () => { await expect(data).toMatchFileSnapshot(`./__snapshots__/${name}.json`); });',
    },
    {
      code: 'import { expect } from "vitest";\nexpect(something).toMatchInlineSnapshot(`${interpolated}`);',
    },
  ],
  invalid: [
    {
      code: 'expect(something).toMatchInlineSnapshot(`${interpolated}`);',
      errors: [{ messageId: 'noInterpolation', line: 1, column: 41 }],
    },
    {
      code: 'expect(something).toThrowErrorMatchingInlineSnapshot(`${interpolated}`);',
      errors: [{ messageId: 'noInterpolation', line: 1, column: 54 }],
    },
    {
      code: 'test("case", ({ expect }) => expect(value).toMatchInlineSnapshot(`${interpolated}`));',
      errors: [{ messageId: 'noInterpolation', line: 1, column: 66 }],
    },
  ],
});
