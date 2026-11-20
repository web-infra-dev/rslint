import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';

function expectNoSpaceBeforeError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'noSpaceBefore',
    message: "There should be no space before '='",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectNeedSpaceBeforeError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'needSpaceBefore',
    message: "A space is required before '='",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectNeedSpaceAfterError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'needSpaceAfter',
    message: "A space is required after '='",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectNoSpaceAfterError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'noSpaceAfter',
    message: "There should be no space after '='",
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('jsx-equals-spacing', {
  valid: [
    {
      code: `<Foo name="value" />`,
      options: ['never'],
    },
    {
      code: `<Foo name = "value" />`,
      options: ['always'],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 0 (7.37.5)',
      code: `<App />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 1 (7.37.5)',
      code: `<App foo />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 2 (7.37.5)',
      code: `<App foo="bar" />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `<App foo={e => bar(e)} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `<App {...props} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `<App />`,
      options: ['never'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `<App foo />`,
      options: ['never'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `<App foo="bar" />`,
      options: ['never'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `<App foo={e => bar(e)} />`,
      options: ['never'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `<App {...props} />`,
      options: ['never'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 10 (7.37.5)',
      code: `<App />`,
      options: ['always'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: `<App foo />`,
      options: ['always'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 12 (7.37.5)',
      code: `<App foo = "bar" />`,
      options: ['always'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 13 (7.37.5)',
      code: `<App foo = {e => bar(e)} />`,
      options: ['always'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 14 (7.37.5)',
      code: `<App {...props} />`,
      options: ['always'],
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    {
      code: `<Foo name ="value" />`,
      options: ['never'],
      errors: [expectNoSpaceBeforeError([1, 11, 1, 12])],
      output: `<Foo name="value" />`,
    },
    {
      code: `<Foo name="value" />`,
      options: ['always'],
      errors: [
        expectNeedSpaceBeforeError([1, 10, 1, 11]),
        expectNeedSpaceAfterError([1, 10, 1, 11]),
      ],
      output: `<Foo name = "value" />`,
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
      code: `<App foo = {bar} />`,
      errors: [
        expectNoSpaceBeforeError([1, 10, 1, 11]),
        expectNoSpaceAfterError([1, 10, 1, 11]),
      ],
      output: `<App foo={bar} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `<App foo = {bar} />`,
      options: ['never'],
      errors: [
        expectNoSpaceBeforeError([1, 10, 1, 11]),
        expectNoSpaceAfterError([1, 10, 1, 11]),
      ],
      output: `<App foo={bar} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `<App foo ={bar} />`,
      options: ['never'],
      errors: [expectNoSpaceBeforeError([1, 10, 1, 11])],
      output: `<App foo={bar} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `<App foo= {bar} />`,
      options: ['never'],
      errors: [expectNoSpaceAfterError([1, 9, 1, 10])],
      output: `<App foo={bar} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `<App foo= {bar} bar = {baz} />`,
      options: ['never'],
      errors: [
        expectNoSpaceAfterError([1, 9, 1, 10]),
        expectNoSpaceBeforeError([1, 21, 1, 22]),
        expectNoSpaceAfterError([1, 21, 1, 22]),
      ],
      output: `<App foo={bar} bar={baz} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `<App foo={bar} />`,
      options: ['always'],
      errors: [
        expectNeedSpaceBeforeError([1, 9, 1, 10]),
        expectNeedSpaceAfterError([1, 9, 1, 10]),
      ],
      output: `<App foo = {bar} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `<App foo ={bar} />`,
      options: ['always'],
      errors: [expectNeedSpaceAfterError([1, 10, 1, 11])],
      output: `<App foo = {bar} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `<App foo= {bar} />`,
      options: ['always'],
      errors: [expectNeedSpaceBeforeError([1, 9, 1, 10])],
      output: `<App foo = {bar} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `<App foo={bar} bar ={baz} />`,
      options: ['always'],
      errors: [
        expectNeedSpaceBeforeError([1, 9, 1, 10]),
        expectNeedSpaceAfterError([1, 9, 1, 10]),
        expectNeedSpaceAfterError([1, 20, 1, 21]),
      ],
      output: `<App foo = {bar} bar = {baz} />`,
      languageOptions: upstreamModuleOptions,
    },
  ],
});
