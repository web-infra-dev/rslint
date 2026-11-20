import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';
import { diagnostic } from '../expectations';

function expectOmitBooleanError(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'omitBoolean',
    'Value must be omitted for boolean attribute `foo`',
    range,
  );
}

function expectOmitPropAndBooleanError(
  range: Parameters<typeof diagnostic>[2],
) {
  return diagnostic(
    'omitPropAndBoolean',
    'Value must be omitted for `false` attribute: `bak`',
    range,
  );
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('jsx-boolean-value', {
  valid: [
    {
      code: `<Foo disabled />`,
    },
    {
      code: `<Foo disabled />`,
      options: ['never'],
    },
    {
      code: `<Foo disabled={true} />`,
      options: ['always'],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 0 (7.37.5)',
      code: `<App foo />;`,
      options: ['never'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 1 (7.37.5)',
      code: `<App foo bar={true} />;`,
      options: ['always', { never: ['foo'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 2 (7.37.5)',
      code: `<App foo />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `<App foo={true} />;`,
      options: ['always'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `<App foo={true} bar />;`,
      options: ['never', { always: ['foo'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `<App />;`,
      options: ['never', { assumeUndefinedIsFalse: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `<App foo={false} />;`,
      options: ['never', { assumeUndefinedIsFalse: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `<App foo={false} />;`,
      options: ['never', { assumeUndefinedIsFalse: true, always: ['foo'] }],
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    {
      code: `<Foo disabled={true} />`,
      options: ['never'],
      errors: [
        diagnostic(
          'omitBoolean',
          'Value must be omitted for boolean attribute `disabled`',
          [1, 6, 1, 21],
        ),
      ],
      output: `<Foo disabled />`,
    },
    {
      code: `<Foo disabled />`,
      options: ['always'],
      errors: [
        diagnostic(
          'setBoolean',
          'Value must be set for boolean attribute `disabled`',
          [1, 6, 1, 14],
        ),
      ],
      output: `<Foo disabled={true} />`,
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
      code: `<App foo={true} />;`,
      options: ['never'],
      errors: [expectOmitBooleanError([1, 6, 1, 16])],
      output: `<App foo />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `<App foo={true} bar={true} baz={true} />;`,
      options: ['always', { never: ['foo', 'bar'] }],
      errors: [
        expectOmitBooleanError([1, 6, 1, 16]),
        diagnostic(
          'omitBoolean',
          'Value must be omitted for boolean attribute `bar`',
          [1, 17, 1, 27],
        ),
      ],
      output: `<App foo bar baz={true} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `<App foo={true} />;`,
      errors: [expectOmitBooleanError([1, 6, 1, 16])],
      output: `<App foo />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `<App foo = {true} />;`,
      errors: [expectOmitBooleanError([1, 6, 1, 18])],
      output: `<App foo />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `<App foo />;`,
      options: ['always'],
      errors: [
        diagnostic(
          'setBoolean',
          'Value must be set for boolean attribute `foo`',
          [1, 6, 1, 9],
        ),
      ],
      output: `<App foo={true} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `<App foo bar baz />;`,
      options: ['never', { always: ['foo', 'bar'] }],
      errors: [
        diagnostic(
          'setBoolean',
          'Value must be set for boolean attribute `foo`',
          [1, 6, 1, 9],
        ),
        diagnostic(
          'setBoolean',
          'Value must be set for boolean attribute `bar`',
          [1, 10, 1, 13],
        ),
      ],
      output: `<App foo={true} bar={true} baz />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `<App foo={false} bak={false} />;`,
      options: ['never', { assumeUndefinedIsFalse: true }],
      errors: [
        diagnostic(
          'omitPropAndBoolean',
          'Value must be omitted for `false` attribute: `foo`',
          [1, 6, 1, 17],
        ),
        expectOmitPropAndBooleanError([1, 18, 1, 29]),
      ],
      output: `<App />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `<App foo={true} bak={false} />;`,
      options: ['never', { assumeUndefinedIsFalse: true }],
      errors: [
        expectOmitBooleanError([1, 6, 1, 16]),
        expectOmitPropAndBooleanError([1, 17, 1, 28]),
      ],
      output: `<App foo />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `<App foo={true} bar={false} baz={false} bak={false} />;`,
      options: [
        'always',
        { assumeUndefinedIsFalse: true, never: ['baz', 'bak'] },
      ],
      errors: [
        diagnostic(
          'omitPropAndBoolean',
          'Value must be omitted for `false` attribute: `baz`',
          [1, 29, 1, 40],
        ),
        expectOmitPropAndBooleanError([1, 41, 1, 52]),
      ],
      output: `<App foo={true} bar={false} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 9 (7.37.5)',
      code: `<App foo={true} bar={true} baz />;`,
      options: ['always', { never: ['foo', 'bar'] }],
      errors: [
        expectOmitBooleanError([1, 6, 1, 16]),
        diagnostic(
          'omitBoolean',
          'Value must be omitted for boolean attribute `bar`',
          [1, 17, 1, 27],
        ),
        diagnostic(
          'setBoolean',
          'Value must be set for boolean attribute `baz`',
          [1, 28, 1, 31],
        ),
      ],
      output: `<App foo bar baz={true} />;`,
      languageOptions: upstreamModuleOptions,
    },
  ],
});
