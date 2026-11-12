import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('no-interpolation-in-snapshots', {
  valid: [
    {
      code: 'expect(something).toMatchInlineSnapshot(`No interpolation`);',
    },
    {
      code: 'expect(fn).toThrowErrorMatchingInlineSnapshot("snap", `case ${id}`);',
    },
    {
      code: 'expect(value).toMatchInlineSnapshot({}, "snap", `case ${id}`);',
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
      errors: [
        {
          messageId: 'noInterpolation',
          message: 'Do not use string interpolation inside of snapshots',
          line: 1,
          column: 41,
          endLine: 1,
          endColumn: 58,
        },
      ],
    },
    {
      code: 'expect(something).toThrowErrorMatchingInlineSnapshot(`${interpolated}`);',
      errors: [
        {
          messageId: 'noInterpolation',
          message: 'Do not use string interpolation inside of snapshots',
          line: 1,
          column: 54,
          endLine: 1,
          endColumn: 71,
        },
      ],
    },
    {
      code: 'expect(something).toMatchInlineSnapshot({}, `${interpolated}`, `case ${id}`);',
      errors: [
        {
          messageId: 'noInterpolation',
          message: 'Do not use string interpolation inside of snapshots',
          line: 1,
          column: 45,
          endLine: 1,
          endColumn: 62,
        },
      ],
    },
    {
      code: 'test("case", ({ expect }) => expect(value).toMatchInlineSnapshot(`${interpolated}`));',
      errors: [
        {
          messageId: 'noInterpolation',
          message: 'Do not use string interpolation inside of snapshots',
          line: 1,
          column: 66,
          endLine: 1,
          endColumn: 83,
        },
      ],
    },
  ],
});
