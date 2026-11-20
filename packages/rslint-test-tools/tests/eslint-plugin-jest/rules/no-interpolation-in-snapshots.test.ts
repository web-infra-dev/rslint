// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('no-interpolation-in-snapshots', {
  valid: [
    ...(
      [
        'expect("something").toEqual("else");',
        'expect(something).toMatchInlineSnapshot();',
        'expect(something).toMatchInlineSnapshot(`No interpolation`);',
        'expect(something).toMatchInlineSnapshot({}, `No interpolation`);',
        'expect(something);',
        'expect(something).not;',
        'expect.toHaveAssertions();',
        'myObjectWants.toMatchInlineSnapshot({}, `${interpolated}`);',
        'myObjectWants.toMatchInlineSnapshot({}, `${interpolated1} ${interpolated2}`);',
        'toMatchInlineSnapshot({}, `${interpolated}`);',
        'toMatchInlineSnapshot({}, `${interpolated1} ${interpolated2}`);',
        'expect(something).toThrowErrorMatchingInlineSnapshot();',
        'expect(something).toThrowErrorMatchingInlineSnapshot(`No interpolation`);',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    ...(
      [
        'expect("something").toEqual("else");',
        'expect(something).toMatchInlineSnapshot();',
        'expect(something).toMatchInlineSnapshot(`No interpolation`);',
        'expect(something).toMatchInlineSnapshot({}, `No interpolation`);',
        'expect(something);',
        'expect(something).not;',
        'expect.toHaveAssertions();',
        'myObjectWants.toMatchInlineSnapshot({}, `${interpolated}`);',
        'myObjectWants.toMatchInlineSnapshot({}, `${interpolated1} ${interpolated2}`);',
        'toMatchInlineSnapshot({}, `${interpolated}`);',
        'toMatchInlineSnapshot({}, `${interpolated1} ${interpolated2}`);',
        'expect(something).toThrowErrorMatchingInlineSnapshot();',
        'expect(something).toThrowErrorMatchingInlineSnapshot(`No interpolation`);',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream no-interpolation-in-snapshots / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
    })),
  ],
  invalid: [
    ...(
      [
        ['expect(something).toMatchInlineSnapshot(`${interpolated}`);', 41],
        ['expect(something).not.toMatchInlineSnapshot(`${interpolated}`);', 45],
        ['expect(something).toMatchInlineSnapshot({}, `${interpolated}`);', 45],
        [
          'expect(something).not.toMatchInlineSnapshot({}, `${interpolated}`);',
          49,
        ],
        [
          'expect(something).toThrowErrorMatchingInlineSnapshot(`${interpolated}`);',
          54,
        ],
        [
          'expect(something).not.toThrowErrorMatchingInlineSnapshot(`${interpolated}`);',
          58,
        ],
      ] as const
    ).map(([code, column], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'noInterpolation',
          message: 'Do not use string interpolation inside of snapshots',
          line: 1,
          column,
          endLine: 1,
          endColumn: column + 17,
        },
      ],
    })),
    ...(
      [
        ['expect(something).toMatchInlineSnapshot(`${interpolated}`);', 41],
        ['expect(something).not.toMatchInlineSnapshot(`${interpolated}`);', 45],
        ['expect(something).toMatchInlineSnapshot({}, `${interpolated}`);', 45],
        [
          'expect(something).not.toMatchInlineSnapshot({}, `${interpolated}`);',
          49,
        ],
        [
          'expect(something).toThrowErrorMatchingInlineSnapshot(`${interpolated}`);',
          54,
        ],
        [
          'expect(something).not.toThrowErrorMatchingInlineSnapshot(`${interpolated}`);',
          58,
        ],
      ] as const
    ).map(([code, column], index): TestCase => ({
      name: `upstream no-interpolation-in-snapshots / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'noInterpolation',
          message: 'Do not use string interpolation inside of snapshots',
          line: 1,
          column,
          endLine: 1,
          endColumn: column + 17,
        },
      ],
    })),
  ],
});
