import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';

function expectOmitBooleanError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'omitBoolean',
    message: 'Value must be omitted for boolean attribute `foo`',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectOmitPropAndBooleanError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'omitPropAndBoolean',
    message: 'Value must be omitted for `false` attribute: `bak`',
    line,
    column,
    endLine,
    endColumn,
  };
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
        {
          messageId: 'omitBoolean',
          message: 'Value must be omitted for boolean attribute `disabled`',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 21,
        },
      ],
      output: `<Foo disabled />`,
    },
    {
      code: `<Foo disabled />`,
      options: ['always'],
      errors: [
        {
          messageId: 'setBoolean',
          message: 'Value must be set for boolean attribute `disabled`',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 14,
        },
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
        {
          messageId: 'omitBoolean',
          message: 'Value must be omitted for boolean attribute `bar`',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 27,
        },
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
        {
          messageId: 'setBoolean',
          message: 'Value must be set for boolean attribute `foo`',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 9,
        },
      ],
      output: `<App foo={true} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `<App foo bar baz />;`,
      options: ['never', { always: ['foo', 'bar'] }],
      errors: [
        {
          messageId: 'setBoolean',
          message: 'Value must be set for boolean attribute `foo`',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 9,
        },
        {
          messageId: 'setBoolean',
          message: 'Value must be set for boolean attribute `bar`',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 13,
        },
      ],
      output: `<App foo={true} bar={true} baz />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `<App foo={false} bak={false} />;`,
      options: ['never', { assumeUndefinedIsFalse: true }],
      errors: [
        {
          messageId: 'omitPropAndBoolean',
          message: 'Value must be omitted for `false` attribute: `foo`',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 17,
        },
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
        {
          messageId: 'omitPropAndBoolean',
          message: 'Value must be omitted for `false` attribute: `baz`',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 40,
        },
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
        {
          messageId: 'omitBoolean',
          message: 'Value must be omitted for boolean attribute `bar`',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 27,
        },
        {
          messageId: 'setBoolean',
          message: 'Value must be set for boolean attribute `baz`',
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 31,
        },
      ],
      output: `<App foo bar baz={true} />;`,
      languageOptions: upstreamModuleOptions,
    },
  ],
});
