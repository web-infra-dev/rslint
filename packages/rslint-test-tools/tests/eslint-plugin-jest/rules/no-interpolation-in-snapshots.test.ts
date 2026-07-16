import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-interpolation-in-snapshots', {} as never, {
  valid: [
    { code: 'expect("something").toEqual("else");' },
    { code: 'expect(something).toMatchInlineSnapshot();' },
    { code: 'expect(something).toMatchInlineSnapshot(`No interpolation`);' },
    {
      code: 'expect(something).toMatchInlineSnapshot({}, `No interpolation`);',
    },
    { code: 'expect(something);' },
    { code: 'expect(something).not;' },
    { code: 'expect.toHaveAssertions();' },
    { code: 'myObjectWants.toMatchInlineSnapshot({}, `${interpolated}`);' },
    {
      code: 'myObjectWants.toMatchInlineSnapshot({}, `${interpolated1} ${interpolated2}`);',
    },
    { code: 'toMatchInlineSnapshot({}, `${interpolated}`);' },
    {
      code: 'toMatchInlineSnapshot({}, `${interpolated1} ${interpolated2}`);',
    },
    { code: 'expect(something).toThrowErrorMatchingInlineSnapshot();' },
    {
      code: 'expect(something).toThrowErrorMatchingInlineSnapshot(`No interpolation`);',
    },
  ],
  invalid: [
    {
      code: 'expect(something).toMatchInlineSnapshot(`${interpolated}`);',
      errors: [
        {
          endColumn: 58,
          column: 41,
          messageId: 'noInterpolation',
        },
      ],
    },
    {
      code: 'expect(something).not.toMatchInlineSnapshot(`${interpolated}`);',
      errors: [
        {
          endColumn: 62,
          column: 45,
          messageId: 'noInterpolation',
        },
      ],
    },
    {
      code: 'expect(something).toMatchInlineSnapshot({}, `${interpolated}`);',
      errors: [
        {
          endColumn: 62,
          column: 45,
          messageId: 'noInterpolation',
        },
      ],
    },
    {
      code: 'expect(something).not.toMatchInlineSnapshot({}, `${interpolated}`);',
      errors: [
        {
          endColumn: 66,
          column: 49,
          messageId: 'noInterpolation',
        },
      ],
    },
    {
      code: 'expect(something).toThrowErrorMatchingInlineSnapshot(`${interpolated}`);',
      errors: [
        {
          endColumn: 71,
          column: 54,
          messageId: 'noInterpolation',
        },
      ],
    },
    {
      code: 'expect(something).not.toThrowErrorMatchingInlineSnapshot(`${interpolated}`);',
      errors: [
        {
          endColumn: 75,
          column: 58,
          messageId: 'noInterpolation',
        },
      ],
    },
  ],
});
