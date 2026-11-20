import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';

function expectNoSpaceAfterError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'noSpaceAfter',
    message: "There should be no space after '{'",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectNoSpaceBeforeError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'noSpaceBefore',
    message: "There should be no space before '}'",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectSpaceNeededAfterError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'spaceNeededAfter',
    message: "A space is required after '{'",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectSpaceNeededBeforeError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'spaceNeededBefore',
    message: "A space is required before '}'",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectNoNewlineAfterError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'noNewlineAfter',
    message: "There should be no newline after '{'",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectNoNewlineBeforeError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'noNewlineBefore',
    message: "There should be no newline before '}'",
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('jsx-curly-spacing', {
  valid: [
    { code: `<App foo={bar} />;` },
    { code: `<App foo={bar} />;`, options: [{ when: 'never' }] },
    { code: `<App foo={ bar } />;`, options: [{ when: 'always' }] },
    { code: `<App foo={ bar } />;`, options: ['always'] },
    { code: `<App foo={bar} />;`, options: ['never'] },
    { code: `<App {...bar} />;` },
    { code: `<App { ...bar } />;`, options: ['always'] },
    {
      code: `<App>{bar}</App>;`,
      options: [{ children: { when: 'never' } }],
    },
    {
      code: `<App>{ bar }</App>;`,
      options: [{ children: { when: 'always' } }],
    },
    {
      code: `<App foo={{ a: 1 }} />;`,
      options: [{ when: 'never' }],
    },
    {
      code: `<App foo={ {a: 1} } />;`,
      options: ['never', { spacing: { objectLiterals: 'always' } }],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 1 (7.37.5)',
      code: `<App foo={bar}>{bar}</App>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 2 (7.37.5)',
      code: `<App foo={bar}>{ bar }</App>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `
        <App foo={
        bar
        }>
        {bar}
        </App>;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `<App foo={{ bar: true, baz: true }}>{{ bar: true, baz: true }}</App>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `<App foo={{ bar: true, baz: true }}>{ { bar: true, baz: true } }</App>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `
        <App foo={
        { bar: true, baz: true }
        } />;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `
        <App foo={
        { bar: true, baz: true }
        }>
        {{ bar: true, baz: true }}
        </App>;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `<App>{ foo /* comment 1 */ }</App>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `<App>{ /* comment 1 */ foo }</App>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 10 (7.37.5)',
      code: `<App foo={bar} />;`,
      options: [{ attributes: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: `
        <App foo={
        bar
        } />;
      `,
      options: [{ attributes: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 12 (7.37.5)',
      code: `<App foo={{ bar: true, baz: true }} />;`,
      options: [{ attributes: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 13 (7.37.5)',
      code: `
        <App foo={
        { bar: true, baz: true }
        } />;
      `,
      options: [{ attributes: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 14 (7.37.5)',
      code: `<App foo={bar} />;`,
      options: [{ attributes: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 15 (7.37.5)',
      code: `
        <App foo={
        bar
        } />;
      `,
      options: [{ attributes: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 16 (7.37.5)',
      code: `<App foo={{ bar: true, baz: true }} />;`,
      options: [{ attributes: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 17 (7.37.5)',
      code: `
        <App foo={
        { bar: true, baz: true }
        } />;
      `,
      options: [{ attributes: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 18 (7.37.5)',
      code: `<App foo={ bar } />;`,
      options: [{ attributes: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 19 (7.37.5)',
      code: `<App foo={ { bar: true, baz: true } } />;`,
      options: [{ attributes: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 20 (7.37.5)',
      code: `<App>{bar}</App>;`,
      options: [{ children: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 21 (7.37.5)',
      code: `
        <App>{
        bar
        }</App>;
      `,
      options: [{ children: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 22 (7.37.5)',
      code: `<App>{{ bar: true, baz: true }}</App>;`,
      options: [{ children: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 23 (7.37.5)',
      code: `
        <App>{
        { bar: true, baz: true }
        }</App>;
      `,
      options: [{ children: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 24 (7.37.5)',
      code: `<App>{bar}</App>;`,
      options: [{ children: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 25 (7.37.5)',
      code: `
        <App>{
        bar
        }</App>;
      `,
      options: [{ children: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 26 (7.37.5)',
      code: `<App>{{ bar: true, baz: true }}</App>;`,
      options: [{ children: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 27 (7.37.5)',
      code: `
        <App>{
        { bar: true, baz: true }
        }</App>;
      `,
      options: [{ children: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 28 (7.37.5)',
      code: `<App>{ bar }</App>;`,
      options: [{ children: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 29 (7.37.5)',
      code: `<App>{ { bar: true, baz: true } }</App>;`,
      options: [{ children: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 31 (7.37.5)',
      code: `<App foo={bar} />;`,
      options: [{ when: 'never', allowMultiline: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 32 (7.37.5)',
      code: `<App foo={bar} />;`,
      options: [{ when: 'never', allowMultiline: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 33 (7.37.5)',
      code: `
        <App foo={
        bar
        } />;
      `,
      options: [{ when: 'never', allowMultiline: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 34 (7.37.5)',
      code: `<App foo={{ bar: true, baz: true }} />;`,
      options: [{ when: 'never', spacing: { objectLiterals: 'never' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 35 (7.37.5)',
      code: `
        <App foo={
        { bar: true, baz: true }
        } />;
      `,
      options: [{ when: 'never', spacing: { objectLiterals: 'never' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 37 (7.37.5)',
      code: `<App foo={ bar } />;`,
      options: [{ when: 'always', allowMultiline: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 38 (7.37.5)',
      code: `<App foo={ bar } />;`,
      options: [{ when: 'always', allowMultiline: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 39 (7.37.5)',
      code: `
        <App foo={
        bar
        } />;
      `,
      options: [{ when: 'always', allowMultiline: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 40 (7.37.5)',
      code: `<App foo={{ bar: true, baz: true }} />;`,
      options: [{ when: 'always', spacing: { objectLiterals: 'never' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 41 (7.37.5)',
      code: `
        <App foo={
        { bar: true, baz: true }
        } />;
      `,
      options: [{ when: 'always', spacing: { objectLiterals: 'never' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 42 (7.37.5)',
      code: `<App foo={bar} />;`,
      options: [{ attributes: { when: 'never' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 43 (7.37.5)',
      code: `<App foo={ bar } />;`,
      options: [{ attributes: { when: 'always' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 44 (7.37.5)',
      code: `<App foo={ bar } />;`,
      options: [{ attributes: { when: 'always', allowMultiline: false } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 45 (7.37.5)',
      code: `<App foo={{ bar:baz }} />;`,
      options: [{ attributes: { when: 'never' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 46 (7.37.5)',
      code: `
        <App foo={
        { bar: true, baz: true }
        } />;
      `,
      options: [{ attributes: { when: 'never' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 47 (7.37.5)',
      code: `<App foo={ {bar:baz} } />;`,
      options: [{ attributes: { when: 'always' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 48 (7.37.5)',
      code: `
        <App foo={
        { bar: true, baz: true }
        } />;
      `,
      options: [{ attributes: { when: 'always' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 49 (7.37.5)',
      code: `
        <App foo={
        bar
        } />;
      `,
      options: [{ attributes: { when: 'always' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 50 (7.37.5)',
      code: `
        <App foo={
        bar
        } />;
      `,
      options: [{ attributes: { when: 'always', allowMultiline: true } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 51 (7.37.5)',
      code: `
        <App foo={
        bar
        } />;
      `,
      options: [{ attributes: { when: 'never' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 52 (7.37.5)',
      code: `
        <App foo={
        bar
        } />;
      `,
      options: [{ attributes: { when: 'never', allowMultiline: true } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 53 (7.37.5)',
      code: `<App foo={bar/* comment 2 */} />;`,
      options: [{ attributes: { when: 'never' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 54 (7.37.5)',
      code: `<App foo={ bar } />;`,
      options: [{ attributes: { when: 'always', spacing: {} } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 55 (7.37.5)',
      code: `
        <App foo={
        bar
        } />;
      `,
      options: [{ attributes: { when: 'always', spacing: {} } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 56 (7.37.5)',
      code: `<App foo={{ bar: true, baz: true }} />;`,
      options: [
        {
          attributes: { when: 'always', spacing: { objectLiterals: 'never' } },
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 57 (7.37.5)',
      code: `
        <App foo={
        { bar: true, baz: true }
        } />;
      `,
      options: [
        {
          attributes: { when: 'always', spacing: { objectLiterals: 'never' } },
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 60 (7.37.5)',
      code: `<App>{ bar }</App>;`,
      options: [{ children: { when: 'always', allowMultiline: false } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 61 (7.37.5)',
      code: `<App>{{ bar:baz }}</App>;`,
      options: [{ children: { when: 'never' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 62 (7.37.5)',
      code: `
        <App>{
        { bar: true, baz: true }
        }</App>;
      `,
      options: [{ children: { when: 'never' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 63 (7.37.5)',
      code: `<App>{ {bar:baz} }</App>;`,
      options: [{ children: { when: 'always' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 64 (7.37.5)',
      code: `
        <App>{
        { bar: true, baz: true }
        }</App>;
      `,
      options: [{ children: { when: 'always' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 65 (7.37.5)',
      code: `
        <App>{
        bar
        }</App>;
      `,
      options: [{ children: { when: 'always' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 66 (7.37.5)',
      code: `
        <App>{
        bar
        }</App>;
      `,
      options: [{ children: { when: 'always', allowMultiline: true } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 67 (7.37.5)',
      code: `
        <App>{
        bar
        }</App>;
      `,
      options: [{ children: { when: 'never' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 68 (7.37.5)',
      code: `
        <App>{
        bar
        }</App>;
      `,
      options: [{ children: { when: 'never', allowMultiline: true } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 69 (7.37.5)',
      code: `
        <App>{/* comment 3 */}</App>;
      `,
      options: [{ children: { when: 'never' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 70 (7.37.5)',
      code: `<App>{bar/* comment 4 */}</App>;`,
      options: [{ children: { when: 'never' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 71 (7.37.5)',
      code: `<App>{ bar }</App>;`,
      options: [{ children: { when: 'always', spacing: {} } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 72 (7.37.5)',
      code: `
        <App>{
        bar
        }</App>;
      `,
      options: [{ children: { when: 'always', spacing: {} } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 73 (7.37.5)',
      code: `<App>{{ bar: true, baz: true }}</App>;`,
      options: [
        { children: { when: 'always', spacing: { objectLiterals: 'never' } } },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 74 (7.37.5)',
      code: `
        <App>{
        { bar: true, baz: true }
        }</App>;
      `,
      options: [
        { children: { when: 'always', spacing: { objectLiterals: 'never' } } },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 76 (7.37.5)',
      code: `<App {...bar} />;`,
      options: [{ attributes: { when: 'never' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 77 (7.37.5)',
      code: `<App { ...bar } />;`,
      options: [{ attributes: { when: 'always' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 78 (7.37.5)',
      code: `<App { ...bar } />;`,
      options: [{ attributes: { when: 'always', allowMultiline: false } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 79 (7.37.5)',
      code: `
        <App {
        ...bar
        } />;
      `,
      options: [{ attributes: { when: 'always' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 80 (7.37.5)',
      code: `
        <App {
        ...bar
        } />;
      `,
      options: [{ attributes: { when: 'always', allowMultiline: true } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 81 (7.37.5)',
      code: `
        <App {
        ...bar
        } />;
      `,
      options: [{ attributes: { when: 'never' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 82 (7.37.5)',
      code: `
        <App {
        ...bar
        } />;
      `,
      options: [{ attributes: { when: 'never', allowMultiline: true } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 83 (7.37.5)',
      code: `<App {...bar/* comment 5 */} />;`,
      options: [{ attributes: { when: 'never' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 84 (7.37.5)',
      code: `<App foo={bar} {...baz} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 85 (7.37.5)',
      code: `<App foo={bar} {...baz} />;`,
      options: [{ attributes: { when: 'never' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 86 (7.37.5)',
      code: `<App foo={ bar } { ...baz } />;`,
      options: [{ attributes: { when: 'always' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 87 (7.37.5)',
      code: `<App foo={ bar } { ...baz } />;`,
      options: [{ attributes: { when: 'always', allowMultiline: false } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 88 (7.37.5)',
      code: `<App foo={{ bar:baz }} {...baz} />;`,
      options: [{ attributes: { when: 'never' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 89 (7.37.5)',
      code: `<App foo={ {bar:baz} } { ...baz } />;`,
      options: [{ attributes: { when: 'always' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 90 (7.37.5)',
      code: `
        <App foo={
        bar
        } {
        ...bar
        } />;
      `,
      options: [{ attributes: { when: 'always' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 91 (7.37.5)',
      code: `<App foo={bar/* comment 6 */} {...baz/* comment 7 */} />;`,
      options: [{ attributes: { when: 'never' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 92 (7.37.5)',
      code: `<App foo={3} bar={ {a: 2} } />`,
      options: [
        {
          attributes: { when: 'never', spacing: { objectLiterals: 'always' } },
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 93 (7.37.5)',
      code: `<App>{bar/* comment 8 */}</App>;`,
      options: [{ children: { when: 'never' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 94 (7.37.5)',
      code: `<App>{bar} {baz}</App>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 95 (7.37.5)',
      code: `<App>{bar} {baz}</App>;`,
      options: [{ children: { when: 'never' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 96 (7.37.5)',
      code: `<App>{ bar } { baz }</App>;`,
      options: [{ children: { when: 'always' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 97 (7.37.5)',
      code: `<App>{ bar } { baz }</App>;`,
      options: [{ children: { when: 'always', allowMultiline: false } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 98 (7.37.5)',
      code: `<App>{{ bar:baz }} {baz}</App>;`,
      options: [{ children: { when: 'never' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 99 (7.37.5)',
      code: `<App>{ {bar:baz} } { baz }</App>;`,
      options: [{ children: { when: 'always' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 100 (7.37.5)',
      code: `
        <App>{
        bar
        } {
        bar
        }</App>;
      `,
      options: [{ children: { when: 'always' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 101 (7.37.5)',
      code: `<App>{bar/* comment 9 */} {baz/* comment 10 */}</App>;`,
      options: [{ children: { when: 'never' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 102 (7.37.5)',
      code: `<App>{3} { {a: 2} }</App>`,
      options: [
        { children: { when: 'never', spacing: { objectLiterals: 'always' } } },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 103 (7.37.5)',
      code: `<App foo={ bar }>{bar}</App>`,
      options: [{ attributes: { when: 'always' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 104 (7.37.5)',
      code: `
        <App foo={ 42 } { ...bar } baz={{ 4: 2 }}>
        {foo} {{ bar: baz }}
        </App>
      `,
      options: [
        {
          when: 'never',
          attributes: { when: 'always', spacing: { objectLiterals: 'never' } },
          children: true,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 105 (7.37.5)',
      code: `
        <App foo={42} {...bar} baz={ { 4: 2 } }>
        {foo} { { bar: baz } }
        </App>
      `,
      options: [
        {
          when: 'never',
          spacing: { objectLiterals: 'always' },
          attributes: true,
          children: { when: 'never' },
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 106 (7.37.5)',
      code: `
        <App foo={42} {...bar} baz={ { 4: 2 } }>
        {foo} { { bar: baz } }
        </App>
      `,
      options: [
        {
          spacing: { objectLiterals: 'always' },
          attributes: { when: 'never', spacing: { objectLiterals: 'always' } },
          children: { when: 'never' },
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 108 (7.37.5)',
      code: `
        <App foo={
        { bar: true, baz: true }
        } />;
      `,
      options: ['never', { spacing: { objectLiterals: 'never' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 110 (7.37.5)',
      code: `<App foo={ bar } />;`,
      options: ['always', { allowMultiline: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 111 (7.37.5)',
      code: `<App foo={{ bar:baz }} />;`,
      options: ['never'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 112 (7.37.5)',
      code: `
        <App foo={
        { bar: true, baz: true }
        } />;
      `,
      options: ['never'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 113 (7.37.5)',
      code: `<App foo={ {bar:baz} } />;`,
      options: ['always'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 114 (7.37.5)',
      code: `
        <App foo={
        { bar: true, baz: true }
        } />;
      `,
      options: ['always'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 115 (7.37.5)',
      code: `
        <App foo={
        bar
        } />;
      `,
      options: ['always'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 116 (7.37.5)',
      code: `
        <App foo={
        bar
        } />;
      `,
      options: ['never'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 117 (7.37.5)',
      code: `
        <App>{/* comment 11 */}</App>;
      `,
      options: ['never'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 118 (7.37.5)',
      code: `<App foo={bar/* comment 12 */} />;`,
      options: ['never'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 119 (7.37.5)',
      code: `<App foo={ bar } />;`,
      options: ['always', { spacing: {} }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 120 (7.37.5)',
      code: `
        <App foo={
        bar
        } />;
      `,
      options: ['always', { spacing: {} }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 121 (7.37.5)',
      code: `<App foo={{ bar: true, baz: true }} />;`,
      options: ['always', { spacing: { objectLiterals: 'never' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 122 (7.37.5)',
      code: `
        <App foo={
        bar
        } />;
      `,
      options: ['always', { allowMultiline: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 123 (7.37.5)',
      code: `
        <App foo={
        { bar: true, baz: true }
        } />;
      `,
      options: ['always', { spacing: { objectLiterals: 'never' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 124 (7.37.5)',
      code: `<App {...bar} />;`,
      options: ['never'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 126 (7.37.5)',
      code: `<App { ...bar } />;`,
      options: ['always', { allowMultiline: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 127 (7.37.5)',
      code: `
        <App {
        ...bar
        } />;
      `,
      options: ['always'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 128 (7.37.5)',
      code: `
        <App {
        ...bar
        } />;
      `,
      options: ['never'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 129 (7.37.5)',
      code: `<App {...bar/* comment 13 */} />;`,
      options: ['never'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 130 (7.37.5)',
      code: `<App foo={bar} {...baz} />;`,
      options: ['never'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 131 (7.37.5)',
      code: `<App foo={ bar } { ...baz } />;`,
      options: ['always'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 132 (7.37.5)',
      code: `<App foo={ bar } { ...baz } />;`,
      options: ['always', { allowMultiline: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 133 (7.37.5)',
      code: `<App foo={{ bar:baz }} {...baz} />;`,
      options: ['never'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 134 (7.37.5)',
      code: `<App foo={ {bar:baz} } { ...baz } />;`,
      options: ['always'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 135 (7.37.5)',
      code: `
        <App foo={
        bar
        } {
        ...bar
        }/>;
      `,
      options: ['always'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 136 (7.37.5)',
      code: `<App foo={bar/* comment 14 */} {...baz/* comment 15 */} />;`,
      options: ['never'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 137 (7.37.5)',
      code: `<App foo={3} bar={ {a: 2} } />`,
      options: ['never', { spacing: { objectLiterals: 'always' } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 138 (7.37.5)',
      code: `<App foo={ bar }>{bar}</App>`,
      options: ['always'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 139 (7.37.5)',
      code: `
        <App>{\`
        text
        \`}</App>
      `,
      options: [{ children: { when: 'never', allowMultiline: false } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 140 (7.37.5)',
      code: `<>{bar} {baz}</>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 141 (7.37.5)',
      code: `<div onLayout={() => { /* dummy callback to fix android bug with component measuring */ }} />`,
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    {
      code: `<App foo={ bar } />;`,
      options: ['never'],
      errors: [
        expectNoSpaceAfterError([1, 10, 1, 11]),
        expectNoSpaceBeforeError([1, 16, 1, 17]),
      ],
      output: `<App foo={bar} />;`,
    },
    {
      code: `<App foo={bar} />;`,
      options: ['always'],
      errors: [
        expectSpaceNeededAfterError([1, 10, 1, 11]),
        expectSpaceNeededBeforeError([1, 14, 1, 15]),
      ],
      output: `<App foo={ bar } />;`,
    },
    {
      code: `<App {...bar} />;`,
      options: ['always'],
      errors: [
        expectSpaceNeededAfterError([1, 6, 1, 7]),
        expectSpaceNeededBeforeError([1, 13, 1, 14]),
      ],
      output: `<App { ...bar } />;`,
    },
    {
      code: `<App>{ bar }</App>;`,
      options: [{ children: { when: 'never' } }],
      errors: [
        expectNoSpaceAfterError([1, 6, 1, 7]),
        expectNoSpaceBeforeError([1, 12, 1, 13]),
      ],
      output: `<App>{bar}</App>;`,
    },
    {
      code: `<App foo={{ a: 1 }} />;`,
      options: ['never', { spacing: { objectLiterals: 'always' } }],
      errors: [
        expectSpaceNeededAfterError([1, 10, 1, 11]),
        expectSpaceNeededBeforeError([1, 19, 1, 20]),
      ],
      output: `<App foo={ { a: 1 } } />;`,
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
      code: `<App foo={ bar }>{bar}</App>;`,
      errors: [
        expectNoSpaceAfterError([1, 10, 1, 11]),
        expectNoSpaceBeforeError([1, 16, 1, 17]),
      ],
      output: `<App foo={bar}>{bar}</App>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `<App foo={ bar }>{ bar }</App>;`,
      errors: [
        expectNoSpaceAfterError([1, 10, 1, 11]),
        expectNoSpaceBeforeError([1, 16, 1, 17]),
      ],
      output: `<App foo={bar}>{ bar }</App>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `<App foo={ { bar: true, baz: true } }>{{ bar: true, baz: true }}</App>;`,
      errors: [
        expectNoSpaceAfterError([1, 10, 1, 11]),
        expectNoSpaceBeforeError([1, 37, 1, 38]),
      ],
      output: `<App foo={{ bar: true, baz: true }}>{{ bar: true, baz: true }}</App>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `<App foo={ { bar: true, baz: true } }>{ { bar: true, baz: true } }</App>;`,
      errors: [
        expectNoSpaceAfterError([1, 10, 1, 11]),
        expectNoSpaceBeforeError([1, 37, 1, 38]),
      ],
      output: `<App foo={{ bar: true, baz: true }}>{ { bar: true, baz: true } }</App>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `<App foo={ bar } />;`,
      options: [{ attributes: true }],
      errors: [
        expectNoSpaceAfterError([1, 10, 1, 11]),
        expectNoSpaceBeforeError([1, 16, 1, 17]),
      ],
      output: `<App foo={bar} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `<App foo={ { bar: true, baz: true } } />;`,
      options: [{ attributes: true }],
      errors: [
        expectNoSpaceAfterError([1, 10, 1, 11]),
        expectNoSpaceBeforeError([1, 37, 1, 38]),
      ],
      output: `<App foo={{ bar: true, baz: true }} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `<App>{ bar }</App>;`,
      options: [{ children: true }],
      errors: [
        expectNoSpaceAfterError([1, 6, 1, 7]),
        expectNoSpaceBeforeError([1, 12, 1, 13]),
      ],
      output: `<App>{bar}</App>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `<>{ bar }</>;`,
      options: [{ children: true }],
      errors: [
        expectNoSpaceAfterError([1, 3, 1, 4]),
        expectNoSpaceBeforeError([1, 9, 1, 10]),
      ],
      output: `<>{bar}</>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `<App>{ { bar: true, baz: true } }</App>;`,
      options: [{ children: true }],
      errors: [
        expectNoSpaceAfterError([1, 6, 1, 7]),
        expectNoSpaceBeforeError([1, 33, 1, 34]),
      ],
      output: `<App>{{ bar: true, baz: true }}</App>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 9 (7.37.5)',
      code: `<App foo={ bar } />;`,
      options: [{ when: 'never' }],
      errors: [
        expectNoSpaceAfterError([1, 10, 1, 11]),
        expectNoSpaceBeforeError([1, 16, 1, 17]),
      ],
      output: `<App foo={bar} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 10 (7.37.5)',
      code: `
        <App foo={
        bar
        } />;
      `,
      options: [{ when: 'never', allowMultiline: false }],
      errors: [
        expectNoNewlineAfterError([2, 18, 2, 19]),
        expectNoNewlineBeforeError([4, 9, 4, 10]),
      ],
      output: `
        <App foo={bar} />;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 11 (7.37.5)',
      code: `<App foo={ { bar: true, baz: true } } />;`,
      options: [{ when: 'never', spacing: { objectLiterals: 'never' } }],
      errors: [
        expectNoSpaceAfterError([1, 10, 1, 11]),
        expectNoSpaceBeforeError([1, 37, 1, 38]),
      ],
      output: `<App foo={{ bar: true, baz: true }} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 12 (7.37.5)',
      code: `
        <App foo={
        { bar: true, baz: true }
        } />;
      `,
      options: [
        {
          when: 'never',
          allowMultiline: false,
          spacing: { objectLiterals: 'never' },
        },
      ],
      errors: [
        expectNoNewlineAfterError([2, 18, 2, 19]),
        expectNoNewlineBeforeError([4, 9, 4, 10]),
      ],
      output: `
        <App foo={{ bar: true, baz: true }} />;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 13 (7.37.5)',
      code: `<App foo={{ bar: true, baz: true }} />;`,
      options: [{ when: 'never', spacing: { objectLiterals: 'always' } }],
      errors: [
        expectSpaceNeededAfterError([1, 10, 1, 11]),
        expectSpaceNeededBeforeError([1, 35, 1, 36]),
      ],
      output: `<App foo={ { bar: true, baz: true } } />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 14 (7.37.5)',
      code: `
        <App foo={
        { bar: true, baz: true }
        } />;
      `,
      options: [
        {
          when: 'never',
          allowMultiline: false,
          spacing: { objectLiterals: 'always' },
        },
      ],
      errors: [
        expectNoNewlineAfterError([2, 18, 2, 19]),
        expectNoNewlineBeforeError([4, 9, 4, 10]),
      ],
      output: `
        <App foo={ { bar: true, baz: true } } />;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 15 (7.37.5)',
      code: `<App foo={bar} />;`,
      options: [{ when: 'always' }],
      errors: [
        expectSpaceNeededAfterError([1, 10, 1, 11]),
        expectSpaceNeededBeforeError([1, 14, 1, 15]),
      ],
      output: `<App foo={ bar } />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 16 (7.37.5)',
      code: `
        <App foo={
        bar
        } />;
      `,
      options: [{ when: 'always', allowMultiline: false }],
      errors: [
        expectNoNewlineAfterError([2, 18, 2, 19]),
        expectNoNewlineBeforeError([4, 9, 4, 10]),
      ],
      output: `
        <App foo={ bar } />;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 17 (7.37.5)',
      code: `<App foo={ { bar: true, baz: true } } />;`,
      options: [{ when: 'always', spacing: { objectLiterals: 'never' } }],
      errors: [
        expectNoSpaceAfterError([1, 10, 1, 11]),
        expectNoSpaceBeforeError([1, 37, 1, 38]),
      ],
      output: `<App foo={{ bar: true, baz: true }} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 18 (7.37.5)',
      code: `
        <App foo={
        { bar: true, baz: true }
        } />;
      `,
      options: [
        {
          when: 'always',
          allowMultiline: false,
          spacing: { objectLiterals: 'never' },
        },
      ],
      errors: [
        expectNoNewlineAfterError([2, 18, 2, 19]),
        expectNoNewlineBeforeError([4, 9, 4, 10]),
      ],
      output: `
        <App foo={{ bar: true, baz: true }} />;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 19 (7.37.5)',
      code: `<App foo={{ bar: true, baz: true }} />;`,
      options: [{ when: 'always', spacing: { objectLiterals: 'always' } }],
      errors: [
        expectSpaceNeededAfterError([1, 10, 1, 11]),
        expectSpaceNeededBeforeError([1, 35, 1, 36]),
      ],
      output: `<App foo={ { bar: true, baz: true } } />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 20 (7.37.5)',
      code: `
        <App foo={
        { bar: true, baz: true }
        } />;
      `,
      options: [
        {
          when: 'always',
          allowMultiline: false,
          spacing: { objectLiterals: 'always' },
        },
      ],
      errors: [
        expectNoNewlineAfterError([2, 18, 2, 19]),
        expectNoNewlineBeforeError([4, 9, 4, 10]),
      ],
      output: `
        <App foo={ { bar: true, baz: true } } />;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 21 (7.37.5)',
      code: `<App foo={ bar } />;`,
      options: [{ attributes: true, when: 'never' }],
      errors: [
        expectNoSpaceAfterError([1, 10, 1, 11]),
        expectNoSpaceBeforeError([1, 16, 1, 17]),
      ],
      output: `<App foo={bar} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 22 (7.37.5)',
      code: `
        <App foo={
        bar
        } />;
      `,
      options: [{ attributes: true, when: 'never', allowMultiline: false }],
      errors: [
        expectNoNewlineAfterError([2, 18, 2, 19]),
        expectNoNewlineBeforeError([4, 9, 4, 10]),
      ],
      output: `
        <App foo={bar} />;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 23 (7.37.5)',
      code: `<App foo={ { bar: true, baz: true } } />;`,
      options: [
        {
          attributes: true,
          when: 'never',
          spacing: { objectLiterals: 'never' },
        },
      ],
      errors: [
        expectNoSpaceAfterError([1, 10, 1, 11]),
        expectNoSpaceBeforeError([1, 37, 1, 38]),
      ],
      output: `<App foo={{ bar: true, baz: true }} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 24 (7.37.5)',
      code: `<App foo={{ bar: true, baz: true }} />;`,
      options: [
        {
          attributes: true,
          when: 'never',
          spacing: { objectLiterals: 'always' },
        },
      ],
      errors: [
        expectSpaceNeededAfterError([1, 10, 1, 11]),
        expectSpaceNeededBeforeError([1, 35, 1, 36]),
      ],
      output: `<App foo={ { bar: true, baz: true } } />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 25 (7.37.5)',
      code: `<App foo={bar} />;`,
      options: [{ attributes: true, when: 'always' }],
      errors: [
        expectSpaceNeededAfterError([1, 10, 1, 11]),
        expectSpaceNeededBeforeError([1, 14, 1, 15]),
      ],
      output: `<App foo={ bar } />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 26 (7.37.5)',
      code: `
        <App foo={
        bar
        } />;
      `,
      options: [{ attributes: true, when: 'always', allowMultiline: false }],
      errors: [
        expectNoNewlineAfterError([2, 18, 2, 19]),
        expectNoNewlineBeforeError([4, 9, 4, 10]),
      ],
      output: `
        <App foo={ bar } />;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 27 (7.37.5)',
      code: `<App foo={ { bar: true, baz: true } } />;`,
      options: [
        {
          attributes: true,
          when: 'always',
          spacing: { objectLiterals: 'never' },
        },
      ],
      errors: [
        expectNoSpaceAfterError([1, 10, 1, 11]),
        expectNoSpaceBeforeError([1, 37, 1, 38]),
      ],
      output: `<App foo={{ bar: true, baz: true }} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 28 (7.37.5)',
      code: `<App foo={{ bar: true, baz: true }} />;`,
      options: [
        {
          attributes: true,
          when: 'always',
          spacing: { objectLiterals: 'always' },
        },
      ],
      errors: [
        expectSpaceNeededAfterError([1, 10, 1, 11]),
        expectSpaceNeededBeforeError([1, 35, 1, 36]),
      ],
      output: `<App foo={ { bar: true, baz: true } } />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 29 (7.37.5)',
      code: `<App foo={ bar } />;`,
      options: [{ attributes: { when: 'never' } }],
      errors: [
        expectNoSpaceAfterError([1, 10, 1, 11]),
        expectNoSpaceBeforeError([1, 16, 1, 17]),
      ],
      output: `<App foo={bar} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 30 (7.37.5)',
      code: `<App foo={ bar } />;`,
      options: [{ attributes: { when: 'never', allowMultiline: false } }],
      errors: [
        expectNoSpaceAfterError([1, 10, 1, 11]),
        expectNoSpaceBeforeError([1, 16, 1, 17]),
      ],
      output: `<App foo={bar} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 31 (7.37.5)',
      code: `<App foo={bar} />;`,
      options: [{ attributes: { when: 'always' } }],
      errors: [
        expectSpaceNeededAfterError([1, 10, 1, 11]),
        expectSpaceNeededBeforeError([1, 14, 1, 15]),
      ],
      output: `<App foo={ bar } />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 32 (7.37.5)',
      code: `<App foo={bar} />;`,
      options: [{ attributes: { when: 'always', allowMultiline: false } }],
      errors: [
        expectSpaceNeededAfterError([1, 10, 1, 11]),
        expectSpaceNeededBeforeError([1, 14, 1, 15]),
      ],
      output: `<App foo={ bar } />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 33 (7.37.5)',
      code: `<App foo={ bar} />;`,
      options: [{ attributes: { when: 'always' } }],
      errors: [expectSpaceNeededBeforeError([1, 15, 1, 16])],
      output: `<App foo={ bar } />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 34 (7.37.5)',
      code: `<App foo={bar } />;`,
      options: [{ attributes: { when: 'always' } }],
      errors: [expectSpaceNeededAfterError([1, 10, 1, 11])],
      output: `<App foo={ bar } />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 35 (7.37.5)',
      code: `<App foo={ bar} />;`,
      options: [{ attributes: { when: 'never' } }],
      errors: [expectNoSpaceAfterError([1, 10, 1, 11])],
      output: `<App foo={bar} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 36 (7.37.5)',
      code: `<App foo={bar } />;`,
      options: [{ attributes: { when: 'never' } }],
      errors: [expectNoSpaceBeforeError([1, 15, 1, 16])],
      output: `<App foo={bar} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 37 (7.37.5)',
      code: `
        <App foo={
        bar
        } />;
      `,
      options: [{ attributes: { when: 'never', allowMultiline: false } }],
      errors: [
        expectNoNewlineAfterError([2, 18, 2, 19]),
        expectNoNewlineBeforeError([4, 9, 4, 10]),
      ],
      output: `
        <App foo={bar} />;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 38 (7.37.5)',
      code: `
        <App foo={
        bar
        } />;
      `,
      options: [{ attributes: { when: 'always', allowMultiline: false } }],
      errors: [
        expectNoNewlineAfterError([2, 18, 2, 19]),
        expectNoNewlineBeforeError([4, 9, 4, 10]),
      ],
      output: `
        <App foo={ bar } />;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 39 (7.37.5)',
      code: `<App foo={bar} />;`,
      options: [{ attributes: { when: 'always', spacing: {} } }],
      errors: [
        expectSpaceNeededAfterError([1, 10, 1, 11]),
        expectSpaceNeededBeforeError([1, 14, 1, 15]),
      ],
      output: `<App foo={ bar } />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 40 (7.37.5)',
      code: `<App foo={ bar} />;`,
      options: [{ attributes: { when: 'always', spacing: {} } }],
      errors: [expectSpaceNeededBeforeError([1, 15, 1, 16])],
      output: `<App foo={ bar } />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 41 (7.37.5)',
      code: `<App foo={bar } />;`,
      options: [{ attributes: { when: 'always', spacing: {} } }],
      errors: [expectSpaceNeededAfterError([1, 10, 1, 11])],
      output: `<App foo={ bar } />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 42 (7.37.5)',
      code: `<App foo={ {bar: true, baz: true} } />;`,
      options: [
        {
          attributes: { when: 'always', spacing: { objectLiterals: 'never' } },
        },
      ],
      errors: [
        expectNoSpaceAfterError([1, 10, 1, 11]),
        expectNoSpaceBeforeError([1, 35, 1, 36]),
      ],
      output: `<App foo={{bar: true, baz: true}} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 43 (7.37.5)',
      code: `<App>{ bar }</App>;`,
      options: [{ children: true, when: 'never' }],
      errors: [
        expectNoSpaceAfterError([1, 6, 1, 7]),
        expectNoSpaceBeforeError([1, 12, 1, 13]),
      ],
      output: `<App>{bar}</App>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 44 (7.37.5)',
      code: `
        <App>{
        bar
        }</App>;
      `,
      options: [{ children: true, when: 'never', allowMultiline: false }],
      errors: [
        expectNoNewlineAfterError([2, 14, 2, 15]),
        expectNoNewlineBeforeError([4, 9, 4, 10]),
      ],
      output: `
        <App>{bar}</App>;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 45 (7.37.5)',
      code: `<App>{ { bar: true, baz: true } }</App>;`,
      options: [
        { children: true, when: 'never', spacing: { objectLiterals: 'never' } },
      ],
      errors: [
        expectNoSpaceAfterError([1, 6, 1, 7]),
        expectNoSpaceBeforeError([1, 33, 1, 34]),
      ],
      output: `<App>{{ bar: true, baz: true }}</App>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 46 (7.37.5)',
      code: `<App>{{ bar: true, baz: true }}</App>;`,
      options: [
        {
          children: true,
          when: 'never',
          spacing: { objectLiterals: 'always' },
        },
      ],
      errors: [
        expectSpaceNeededAfterError([1, 6, 1, 7]),
        expectSpaceNeededBeforeError([1, 31, 1, 32]),
      ],
      output: `<App>{ { bar: true, baz: true } }</App>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 47 (7.37.5)',
      code: `<App>{bar}</App>;`,
      options: [{ children: true, when: 'always' }],
      errors: [
        expectSpaceNeededAfterError([1, 6, 1, 7]),
        expectSpaceNeededBeforeError([1, 10, 1, 11]),
      ],
      output: `<App>{ bar }</App>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 48 (7.37.5)',
      code: `
        <App>{
        bar
        }</App>;
      `,
      options: [{ children: true, when: 'always', allowMultiline: false }],
      errors: [
        expectNoNewlineAfterError([2, 14, 2, 15]),
        expectNoNewlineBeforeError([4, 9, 4, 10]),
      ],
      output: `
        <App>{ bar }</App>;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 49 (7.37.5)',
      code: `<App>{ { bar: true, baz: true } }</App>;`,
      options: [
        {
          children: true,
          when: 'always',
          spacing: { objectLiterals: 'never' },
        },
      ],
      errors: [
        expectNoSpaceAfterError([1, 6, 1, 7]),
        expectNoSpaceBeforeError([1, 33, 1, 34]),
      ],
      output: `<App>{{ bar: true, baz: true }}</App>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 50 (7.37.5)',
      code: `<App>{{ bar: true, baz: true }}</App>;`,
      options: [
        {
          children: true,
          when: 'always',
          spacing: { objectLiterals: 'always' },
        },
      ],
      errors: [
        expectSpaceNeededAfterError([1, 6, 1, 7]),
        expectSpaceNeededBeforeError([1, 31, 1, 32]),
      ],
      output: `<App>{ { bar: true, baz: true } }</App>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 52 (7.37.5)',
      code: `<App>{ bar }</App>;`,
      options: [{ children: { when: 'never', allowMultiline: false } }],
      errors: [
        expectNoSpaceAfterError([1, 6, 1, 7]),
        expectNoSpaceBeforeError([1, 12, 1, 13]),
      ],
      output: `<App>{bar}</App>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 53 (7.37.5)',
      code: `<App>{bar}</App>;`,
      options: [{ children: { when: 'always' } }],
      errors: [
        expectSpaceNeededAfterError([1, 6, 1, 7]),
        expectSpaceNeededBeforeError([1, 10, 1, 11]),
      ],
      output: `<App>{ bar }</App>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 54 (7.37.5)',
      code: `<App>{bar}</App>;`,
      options: [{ children: { when: 'always', allowMultiline: false } }],
      errors: [
        expectSpaceNeededAfterError([1, 6, 1, 7]),
        expectSpaceNeededBeforeError([1, 10, 1, 11]),
      ],
      output: `<App>{ bar }</App>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 55 (7.37.5)',
      code: `<App>{ bar}</App>;`,
      options: [{ children: { when: 'always' } }],
      errors: [expectSpaceNeededBeforeError([1, 11, 1, 12])],
      output: `<App>{ bar }</App>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 56 (7.37.5)',
      code: `<App>{bar }</App>;`,
      options: [{ children: { when: 'always' } }],
      errors: [expectSpaceNeededAfterError([1, 6, 1, 7])],
      output: `<App>{ bar }</App>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 57 (7.37.5)',
      code: `<App>{ bar}</App>;`,
      options: [{ children: { when: 'never' } }],
      errors: [expectNoSpaceAfterError([1, 6, 1, 7])],
      output: `<App>{bar}</App>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 58 (7.37.5)',
      code: `<App>{bar }</App>;`,
      options: [{ children: { when: 'never' } }],
      errors: [expectNoSpaceBeforeError([1, 11, 1, 12])],
      output: `<App>{bar}</App>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 59 (7.37.5)',
      code: `
        <App>{
        bar
        }</App>;
      `,
      options: [{ children: { when: 'never', allowMultiline: false } }],
      errors: [
        expectNoNewlineAfterError([2, 14, 2, 15]),
        expectNoNewlineBeforeError([4, 9, 4, 10]),
      ],
      output: `
        <App>{bar}</App>;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 60 (7.37.5)',
      code: `
        <App>{
        bar
        }</App>;
      `,
      options: [{ children: { when: 'always', allowMultiline: false } }],
      errors: [
        expectNoNewlineAfterError([2, 14, 2, 15]),
        expectNoNewlineBeforeError([4, 9, 4, 10]),
      ],
      output: `
        <App>{ bar }</App>;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 61 (7.37.5)',
      code: `<App>{bar}</App>;`,
      options: [{ children: { when: 'always', spacing: {} } }],
      errors: [
        expectSpaceNeededAfterError([1, 6, 1, 7]),
        expectSpaceNeededBeforeError([1, 10, 1, 11]),
      ],
      output: `<App>{ bar }</App>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 62 (7.37.5)',
      code: `<App>{ bar}</App>;`,
      options: [{ children: { when: 'always', spacing: {} } }],
      errors: [expectSpaceNeededBeforeError([1, 11, 1, 12])],
      output: `<App>{ bar }</App>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 63 (7.37.5)',
      code: `<App>{bar }</App>;`,
      options: [{ children: { when: 'always', spacing: {} } }],
      errors: [expectSpaceNeededAfterError([1, 6, 1, 7])],
      output: `<App>{ bar }</App>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 64 (7.37.5)',
      code: `<App>{ {bar: true, baz: true} }</App>;`,
      options: [
        { children: { when: 'always', spacing: { objectLiterals: 'never' } } },
      ],
      errors: [
        expectNoSpaceAfterError([1, 6, 1, 7]),
        expectNoSpaceBeforeError([1, 31, 1, 32]),
      ],
      output: `<App>{{bar: true, baz: true}}</App>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 65 (7.37.5)',
      code: `<App { ...bar } />;`,
      options: [{ attributes: { when: 'never' } }],
      errors: [
        expectNoSpaceAfterError([1, 6, 1, 7]),
        expectNoSpaceBeforeError([1, 15, 1, 16]),
      ],
      output: `<App {...bar} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 66 (7.37.5)',
      code: `<App { ...bar } />;`,
      options: [{ attributes: { when: 'never', allowMultiline: false } }],
      errors: [
        expectNoSpaceAfterError([1, 6, 1, 7]),
        expectNoSpaceBeforeError([1, 15, 1, 16]),
      ],
      output: `<App {...bar} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 67 (7.37.5)',
      code: `<App {...bar} />;`,
      options: [{ attributes: { when: 'always' } }],
      errors: [
        expectSpaceNeededAfterError([1, 6, 1, 7]),
        expectSpaceNeededBeforeError([1, 13, 1, 14]),
      ],
      output: `<App { ...bar } />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 68 (7.37.5)',
      code: `<App {...bar} />;`,
      options: [{ attributes: { when: 'always', allowMultiline: false } }],
      errors: [
        expectSpaceNeededAfterError([1, 6, 1, 7]),
        expectSpaceNeededBeforeError([1, 13, 1, 14]),
      ],
      output: `<App { ...bar } />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 69 (7.37.5)',
      code: `<App { ...bar} />;`,
      options: [{ attributes: { when: 'always' } }],
      errors: [expectSpaceNeededBeforeError([1, 14, 1, 15])],
      output: `<App { ...bar } />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 70 (7.37.5)',
      code: `<App {...bar } />;`,
      options: [{ attributes: { when: 'always' } }],
      errors: [expectSpaceNeededAfterError([1, 6, 1, 7])],
      output: `<App { ...bar } />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 71 (7.37.5)',
      code: `<App { ...bar} />;`,
      options: [{ attributes: { when: 'never' } }],
      errors: [expectNoSpaceAfterError([1, 6, 1, 7])],
      output: `<App {...bar} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 72 (7.37.5)',
      code: `<App {...bar } />;`,
      options: [{ attributes: { when: 'never' } }],
      errors: [expectNoSpaceBeforeError([1, 14, 1, 15])],
      output: `<App {...bar} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 73 (7.37.5)',
      code: `
        <App {
        ...bar
        } />;
      `,
      options: [{ attributes: { when: 'never', allowMultiline: false } }],
      errors: [
        expectNoNewlineAfterError([2, 14, 2, 15]),
        expectNoNewlineBeforeError([4, 9, 4, 10]),
      ],
      output: `
        <App {...bar} />;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 74 (7.37.5)',
      code: `
        <App {
        ...bar
        } />;
      `,
      options: [{ attributes: { when: 'always', allowMultiline: false } }],
      errors: [
        expectNoNewlineAfterError([2, 14, 2, 15]),
        expectNoNewlineBeforeError([4, 9, 4, 10]),
      ],
      output: `
        <App { ...bar } />;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 75 (7.37.5)',
      code: `<App foo={ bar } { ...baz } />;`,
      options: [{ attributes: { when: 'never' } }],
      errors: [
        expectNoSpaceAfterError([1, 10, 1, 11]),
        expectNoSpaceBeforeError([1, 16, 1, 17]),
        expectNoSpaceAfterError([1, 18, 1, 19]),
        expectNoSpaceBeforeError([1, 27, 1, 28]),
      ],
      output: `<App foo={bar} {...baz} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 76 (7.37.5)',
      code: `<App foo={ bar } { ...baz } />;`,
      options: [{ attributes: { when: 'never', allowMultiline: false } }],
      errors: [
        expectNoSpaceAfterError([1, 10, 1, 11]),
        expectNoSpaceBeforeError([1, 16, 1, 17]),
        expectNoSpaceAfterError([1, 18, 1, 19]),
        expectNoSpaceBeforeError([1, 27, 1, 28]),
      ],
      output: `<App foo={bar} {...baz} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 77 (7.37.5)',
      code: `<App foo={bar} {...baz} />;`,
      options: [{ attributes: { when: 'always' } }],
      errors: [
        expectSpaceNeededAfterError([1, 10, 1, 11]),
        expectSpaceNeededBeforeError([1, 14, 1, 15]),
        expectSpaceNeededAfterError([1, 16, 1, 17]),
        expectSpaceNeededBeforeError([1, 23, 1, 24]),
      ],
      output: `<App foo={ bar } { ...baz } />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 78 (7.37.5)',
      code: `<App foo={bar} {...baz} />;`,
      options: [{ attributes: { when: 'always', allowMultiline: false } }],
      errors: [
        expectSpaceNeededAfterError([1, 10, 1, 11]),
        expectSpaceNeededBeforeError([1, 14, 1, 15]),
        expectSpaceNeededAfterError([1, 16, 1, 17]),
        expectSpaceNeededBeforeError([1, 23, 1, 24]),
      ],
      output: `<App foo={ bar } { ...baz } />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 79 (7.37.5)',
      code: `<App foo={ bar} { ...baz} />;`,
      options: [{ attributes: { when: 'always' } }],
      errors: [
        expectSpaceNeededBeforeError([1, 15, 1, 16]),
        expectSpaceNeededBeforeError([1, 25, 1, 26]),
      ],
      output: `<App foo={ bar } { ...baz } />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 80 (7.37.5)',
      code: `<App foo={bar } {...baz } />;`,
      options: [{ attributes: { when: 'always' } }],
      errors: [
        expectSpaceNeededAfterError([1, 10, 1, 11]),
        expectSpaceNeededAfterError([1, 17, 1, 18]),
      ],
      output: `<App foo={ bar } { ...baz } />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 81 (7.37.5)',
      code: `<App foo={ bar} { ...baz} />;`,
      options: [{ attributes: { when: 'never' } }],
      errors: [
        expectNoSpaceAfterError([1, 10, 1, 11]),
        expectNoSpaceAfterError([1, 17, 1, 18]),
      ],
      output: `<App foo={bar} {...baz} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 82 (7.37.5)',
      code: `<App foo={bar } {...baz } />;`,
      options: [{ attributes: { when: 'never' } }],
      errors: [
        expectNoSpaceBeforeError([1, 15, 1, 16]),
        expectNoSpaceBeforeError([1, 25, 1, 26]),
      ],
      output: `<App foo={bar} {...baz} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 83 (7.37.5)',
      code: `
        <App foo={
        bar
        } {
        ...baz
        } />;
      `,
      options: [{ attributes: { when: 'never', allowMultiline: false } }],
      errors: [
        expectNoNewlineAfterError([2, 18, 2, 19]),
        expectNoNewlineBeforeError([4, 9, 4, 10]),
        expectNoNewlineAfterError([4, 11, 4, 12]),
        expectNoNewlineBeforeError([6, 9, 6, 10]),
      ],
      output: `
        <App foo={bar} {...baz} />;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 84 (7.37.5)',
      code: `
        <App foo={
        bar
        } {
        ...baz
        } />;
      `,
      options: [{ attributes: { when: 'always', allowMultiline: false } }],
      errors: [
        expectNoNewlineAfterError([2, 18, 2, 19]),
        expectNoNewlineBeforeError([4, 9, 4, 10]),
        expectNoNewlineAfterError([4, 11, 4, 12]),
        expectNoNewlineBeforeError([6, 9, 6, 10]),
      ],
      output: `
        <App foo={ bar } { ...baz } />;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 85 (7.37.5)',
      code: `<App foo={ 3 } bar={{a: 2}} />`,
      options: [
        {
          attributes: { when: 'never', spacing: { objectLiterals: 'always' } },
        },
      ],
      errors: [
        expectNoSpaceAfterError([1, 10, 1, 11]),
        expectNoSpaceBeforeError([1, 14, 1, 15]),
        expectSpaceNeededAfterError([1, 20, 1, 21]),
        expectSpaceNeededBeforeError([1, 27, 1, 28]),
      ],
      output: `<App foo={3} bar={ {a: 2} } />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 86 (7.37.5)',
      code: `<App foo={ foo /* comment 16 */ } />`,
      errors: [
        expectNoSpaceAfterError([1, 10, 1, 11]),
        expectNoSpaceBeforeError([1, 33, 1, 34]),
      ],
      output: `<App foo={foo /* comment 16 */} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 87 (7.37.5)',
      code: `<App foo={foo /* comment 17 */} />`,
      options: [{ attributes: { when: 'always' } }],
      errors: [
        expectSpaceNeededAfterError([1, 10, 1, 11]),
        expectSpaceNeededBeforeError([1, 31, 1, 32]),
      ],
      output: `<App foo={ foo /* comment 17 */ } />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 88 (7.37.5)',
      code: `<App foo={ /* comment 18 */ foo } />`,
      errors: [
        expectNoSpaceAfterError([1, 10, 1, 11]),
        expectNoSpaceBeforeError([1, 33, 1, 34]),
      ],
      output: `<App foo={/* comment 18 */ foo} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 89 (7.37.5)',
      code: `<App foo={/* comment 19 */ foo} />`,
      options: [{ attributes: { when: 'always' } }],
      errors: [
        expectSpaceNeededAfterError([1, 10, 1, 11]),
        expectSpaceNeededBeforeError([1, 31, 1, 32]),
      ],
      output: `<App foo={ /* comment 19 */ foo } />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 90 (7.37.5)',
      code: `<App>{ bar } { baz }</App>;`,
      options: [{ children: { when: 'never' } }],
      errors: [
        expectNoSpaceAfterError([1, 6, 1, 7]),
        expectNoSpaceBeforeError([1, 12, 1, 13]),
        expectNoSpaceAfterError([1, 14, 1, 15]),
        expectNoSpaceBeforeError([1, 20, 1, 21]),
      ],
      output: `<App>{bar} {baz}</App>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 91 (7.37.5)',
      code: `<App>{ bar } { baz }</App>;`,
      options: [{ children: { when: 'never', allowMultiline: false } }],
      errors: [
        expectNoSpaceAfterError([1, 6, 1, 7]),
        expectNoSpaceBeforeError([1, 12, 1, 13]),
        expectNoSpaceAfterError([1, 14, 1, 15]),
        expectNoSpaceBeforeError([1, 20, 1, 21]),
      ],
      output: `<App>{bar} {baz}</App>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 92 (7.37.5)',
      code: `<App>{bar} {baz}</App>;`,
      options: [{ children: { when: 'always' } }],
      errors: [
        expectSpaceNeededAfterError([1, 6, 1, 7]),
        expectSpaceNeededBeforeError([1, 10, 1, 11]),
        expectSpaceNeededAfterError([1, 12, 1, 13]),
        expectSpaceNeededBeforeError([1, 16, 1, 17]),
      ],
      output: `<App>{ bar } { baz }</App>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 93 (7.37.5)',
      code: `<App>{bar} {baz}</App>;`,
      options: [{ children: { when: 'always', allowMultiline: false } }],
      errors: [
        expectSpaceNeededAfterError([1, 6, 1, 7]),
        expectSpaceNeededBeforeError([1, 10, 1, 11]),
        expectSpaceNeededAfterError([1, 12, 1, 13]),
        expectSpaceNeededBeforeError([1, 16, 1, 17]),
      ],
      output: `<App>{ bar } { baz }</App>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 94 (7.37.5)',
      code: `<App>{ bar} { baz}</App>;`,
      options: [{ children: { when: 'always' } }],
      errors: [
        expectSpaceNeededBeforeError([1, 11, 1, 12]),
        expectSpaceNeededBeforeError([1, 18, 1, 19]),
      ],
      output: `<App>{ bar } { baz }</App>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 95 (7.37.5)',
      code: `<App>{bar } {baz }</App>;`,
      options: [{ children: { when: 'always' } }],
      errors: [
        expectSpaceNeededAfterError([1, 6, 1, 7]),
        expectSpaceNeededAfterError([1, 13, 1, 14]),
      ],
      output: `<App>{ bar } { baz }</App>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 96 (7.37.5)',
      code: `<App>{ bar} { baz}</App>;`,
      options: [{ children: { when: 'never' } }],
      errors: [
        expectNoSpaceAfterError([1, 6, 1, 7]),
        expectNoSpaceAfterError([1, 13, 1, 14]),
      ],
      output: `<App>{bar} {baz}</App>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 97 (7.37.5)',
      code: `<App>{bar } {baz }</App>;`,
      options: [{ children: { when: 'never' } }],
      errors: [
        expectNoSpaceBeforeError([1, 11, 1, 12]),
        expectNoSpaceBeforeError([1, 18, 1, 19]),
      ],
      output: `<App>{bar} {baz}</App>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 98 (7.37.5)',
      code: `
        <App>{
        bar
        } {
        baz
        }</App>;
      `,
      options: [{ children: { when: 'never', allowMultiline: false } }],
      errors: [
        expectNoNewlineAfterError([2, 14, 2, 15]),
        expectNoNewlineBeforeError([4, 9, 4, 10]),
        expectNoNewlineAfterError([4, 11, 4, 12]),
        expectNoNewlineBeforeError([6, 9, 6, 10]),
      ],
      output: `
        <App>{bar} {baz}</App>;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 99 (7.37.5)',
      code: `
        <App>{
        bar
        } {
        baz
        }</App>;
      `,
      options: [{ children: { when: 'always', allowMultiline: false } }],
      errors: [
        expectNoNewlineAfterError([2, 14, 2, 15]),
        expectNoNewlineBeforeError([4, 9, 4, 10]),
        expectNoNewlineAfterError([4, 11, 4, 12]),
        expectNoNewlineBeforeError([6, 9, 6, 10]),
      ],
      output: `
        <App>{ bar } { baz }</App>;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 100 (7.37.5)',
      code: `<App>{ 3 } bar={{a: 2}}</App>`,
      options: [
        { children: { when: 'never', spacing: { objectLiterals: 'always' } } },
      ],
      errors: [
        expectNoSpaceAfterError([1, 6, 1, 7]),
        expectNoSpaceBeforeError([1, 10, 1, 11]),
        expectSpaceNeededAfterError([1, 16, 1, 17]),
        expectSpaceNeededBeforeError([1, 23, 1, 24]),
      ],
      output: `<App>{3} bar={ {a: 2} }</App>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 101 (7.37.5)',
      code: `<App>{foo /* comment 20 */}</App>`,
      options: [{ children: { when: 'always' } }],
      errors: [
        expectSpaceNeededAfterError([1, 6, 1, 7]),
        expectSpaceNeededBeforeError([1, 27, 1, 28]),
      ],
      output: `<App>{ foo /* comment 20 */ }</App>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 102 (7.37.5)',
      code: `<App>{/* comment 21 */ foo}</App>`,
      options: [{ children: { when: 'always' } }],
      errors: [
        expectSpaceNeededAfterError([1, 6, 1, 7]),
        expectSpaceNeededBeforeError([1, 27, 1, 28]),
      ],
      output: `<App>{ /* comment 21 */ foo }</App>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 104 (7.37.5)',
      code: `<App foo={ bar } />;`,
      options: ['never', { allowMultiline: false }],
      errors: [
        expectNoSpaceAfterError([1, 10, 1, 11]),
        expectNoSpaceBeforeError([1, 16, 1, 17]),
      ],
      output: `<App foo={bar} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 105 (7.37.5)',
      code: `
        <App foo={
        { bar: true, baz: true }
        } />;
      `,
      options: [
        'never',
        { allowMultiline: false, spacing: { objectLiterals: 'never' } },
      ],
      errors: [
        expectNoNewlineAfterError([2, 18, 2, 19]),
        expectNoNewlineBeforeError([4, 9, 4, 10]),
      ],
      output: `
        <App foo={{ bar: true, baz: true }} />;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 106 (7.37.5)',
      code: `
        <App foo={
        { bar: true, baz: true }
        } />;
      `,
      options: [
        'never',
        { allowMultiline: false, spacing: { objectLiterals: 'always' } },
      ],
      errors: [
        expectNoNewlineAfterError([2, 18, 2, 19]),
        expectNoNewlineBeforeError([4, 9, 4, 10]),
      ],
      output: `
        <App foo={ { bar: true, baz: true } } />;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 107 (7.37.5)',
      code: `
        <App foo={
        { bar: true, baz: true }
        } />;
      `,
      options: [
        'always',
        { allowMultiline: false, spacing: { objectLiterals: 'never' } },
      ],
      errors: [
        expectNoNewlineAfterError([2, 18, 2, 19]),
        expectNoNewlineBeforeError([4, 9, 4, 10]),
      ],
      output: `
        <App foo={{ bar: true, baz: true }} />;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 108 (7.37.5)',
      code: `
        <App foo={
        { bar: true, baz: true }
        } />;
      `,
      options: [
        'always',
        { allowMultiline: false, spacing: { objectLiterals: 'always' } },
      ],
      errors: [
        expectNoNewlineAfterError([2, 18, 2, 19]),
        expectNoNewlineBeforeError([4, 9, 4, 10]),
      ],
      output: `
        <App foo={ { bar: true, baz: true } } />;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 110 (7.37.5)',
      code: `<App foo={bar} />;`,
      options: ['always', { allowMultiline: false }],
      errors: [
        expectSpaceNeededAfterError([1, 10, 1, 11]),
        expectSpaceNeededBeforeError([1, 14, 1, 15]),
      ],
      output: `<App foo={ bar } />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 111 (7.37.5)',
      code: `<App foo={ bar} />;`,
      options: ['always'],
      errors: [expectSpaceNeededBeforeError([1, 15, 1, 16])],
      output: `<App foo={ bar } />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 112 (7.37.5)',
      code: `<App foo={bar } />;`,
      options: ['always'],
      errors: [expectSpaceNeededAfterError([1, 10, 1, 11])],
      output: `<App foo={ bar } />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 113 (7.37.5)',
      code: `<App foo={ bar} />;`,
      options: ['never'],
      errors: [expectNoSpaceAfterError([1, 10, 1, 11])],
      output: `<App foo={bar} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 114 (7.37.5)',
      code: `<App foo={bar } />;`,
      options: ['never'],
      errors: [expectNoSpaceBeforeError([1, 15, 1, 16])],
      output: `<App foo={bar} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 115 (7.37.5)',
      code: `
        <App foo={
        bar
        } />;
      `,
      options: ['never', { allowMultiline: false }],
      errors: [
        expectNoNewlineAfterError([2, 18, 2, 19]),
        expectNoNewlineBeforeError([4, 9, 4, 10]),
      ],
      output: `
        <App foo={bar} />;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 116 (7.37.5)',
      code: `
        <App foo={
        bar
        } />;
      `,
      options: ['always', { allowMultiline: false }],
      errors: [
        expectNoNewlineAfterError([2, 18, 2, 19]),
        expectNoNewlineBeforeError([4, 9, 4, 10]),
      ],
      output: `
        <App foo={ bar } />;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 117 (7.37.5)',
      code: `<App foo={bar} />;`,
      options: ['always', { spacing: {} }],
      errors: [
        expectSpaceNeededAfterError([1, 10, 1, 11]),
        expectSpaceNeededBeforeError([1, 14, 1, 15]),
      ],
      output: `<App foo={ bar } />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 118 (7.37.5)',
      code: `<App foo={ bar} />;`,
      options: ['always', { spacing: {} }],
      errors: [expectSpaceNeededBeforeError([1, 15, 1, 16])],
      output: `<App foo={ bar } />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 119 (7.37.5)',
      code: `<App foo={bar } />;`,
      options: ['always', { spacing: {} }],
      errors: [expectSpaceNeededAfterError([1, 10, 1, 11])],
      output: `<App foo={ bar } />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 120 (7.37.5)',
      code: `<App foo={ {bar: true, baz: true} } />;`,
      options: ['always', { spacing: { objectLiterals: 'never' } }],
      errors: [
        expectNoSpaceAfterError([1, 10, 1, 11]),
        expectNoSpaceBeforeError([1, 35, 1, 36]),
      ],
      output: `<App foo={{bar: true, baz: true}} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 121 (7.37.5)',
      code: `<App { ...bar } />;`,
      options: ['never'],
      errors: [
        expectNoSpaceAfterError([1, 6, 1, 7]),
        expectNoSpaceBeforeError([1, 15, 1, 16]),
      ],
      output: `<App {...bar} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 122 (7.37.5)',
      code: `<App { ...bar } />;`,
      options: ['never', { allowMultiline: false }],
      errors: [
        expectNoSpaceAfterError([1, 6, 1, 7]),
        expectNoSpaceBeforeError([1, 15, 1, 16]),
      ],
      output: `<App {...bar} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 124 (7.37.5)',
      code: `<App {...bar} />;`,
      options: ['always', { allowMultiline: false }],
      errors: [
        expectSpaceNeededAfterError([1, 6, 1, 7]),
        expectSpaceNeededBeforeError([1, 13, 1, 14]),
      ],
      output: `<App { ...bar } />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 125 (7.37.5)',
      code: `<App { ...bar} />;`,
      options: ['always'],
      errors: [expectSpaceNeededBeforeError([1, 14, 1, 15])],
      output: `<App { ...bar } />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 126 (7.37.5)',
      code: `<App {...bar } />;`,
      options: ['always'],
      errors: [expectSpaceNeededAfterError([1, 6, 1, 7])],
      output: `<App { ...bar } />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 127 (7.37.5)',
      code: `<App { ...bar} />;`,
      options: ['never'],
      errors: [expectNoSpaceAfterError([1, 6, 1, 7])],
      output: `<App {...bar} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 128 (7.37.5)',
      code: `<App {...bar } />;`,
      options: ['never'],
      errors: [expectNoSpaceBeforeError([1, 14, 1, 15])],
      output: `<App {...bar} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 129 (7.37.5)',
      code: `
        <App {
        ...bar
        } />;
      `,
      options: ['never', { allowMultiline: false }],
      errors: [
        expectNoNewlineAfterError([2, 14, 2, 15]),
        expectNoNewlineBeforeError([4, 9, 4, 10]),
      ],
      output: `
        <App {...bar} />;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 130 (7.37.5)',
      code: `
        <App {
        ...bar
        } />;
      `,
      options: ['always', { allowMultiline: false }],
      errors: [
        expectNoNewlineAfterError([2, 14, 2, 15]),
        expectNoNewlineBeforeError([4, 9, 4, 10]),
      ],
      output: `
        <App { ...bar } />;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 131 (7.37.5)',
      code: `<App foo={ bar } { ...baz } />;`,
      options: ['never'],
      errors: [
        expectNoSpaceAfterError([1, 10, 1, 11]),
        expectNoSpaceBeforeError([1, 16, 1, 17]),
        expectNoSpaceAfterError([1, 18, 1, 19]),
        expectNoSpaceBeforeError([1, 27, 1, 28]),
      ],
      output: `<App foo={bar} {...baz} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 132 (7.37.5)',
      code: `<App foo={ bar } { ...baz } />;`,
      options: ['never', { allowMultiline: false }],
      errors: [
        expectNoSpaceAfterError([1, 10, 1, 11]),
        expectNoSpaceBeforeError([1, 16, 1, 17]),
        expectNoSpaceAfterError([1, 18, 1, 19]),
        expectNoSpaceBeforeError([1, 27, 1, 28]),
      ],
      output: `<App foo={bar} {...baz} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 133 (7.37.5)',
      code: `<App foo={bar} {...baz} />;`,
      options: ['always'],
      errors: [
        expectSpaceNeededAfterError([1, 10, 1, 11]),
        expectSpaceNeededBeforeError([1, 14, 1, 15]),
        expectSpaceNeededAfterError([1, 16, 1, 17]),
        expectSpaceNeededBeforeError([1, 23, 1, 24]),
      ],
      output: `<App foo={ bar } { ...baz } />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 134 (7.37.5)',
      code: `<App foo={bar} {...baz} />;`,
      options: ['always', { allowMultiline: false }],
      errors: [
        expectSpaceNeededAfterError([1, 10, 1, 11]),
        expectSpaceNeededBeforeError([1, 14, 1, 15]),
        expectSpaceNeededAfterError([1, 16, 1, 17]),
        expectSpaceNeededBeforeError([1, 23, 1, 24]),
      ],
      output: `<App foo={ bar } { ...baz } />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 135 (7.37.5)',
      code: `<App foo={ bar} { ...baz} />;`,
      options: ['always'],
      errors: [
        expectSpaceNeededBeforeError([1, 15, 1, 16]),
        expectSpaceNeededBeforeError([1, 25, 1, 26]),
      ],
      output: `<App foo={ bar } { ...baz } />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 136 (7.37.5)',
      code: `<App foo={bar } {...baz } />;`,
      options: ['always'],
      errors: [
        expectSpaceNeededAfterError([1, 10, 1, 11]),
        expectSpaceNeededAfterError([1, 17, 1, 18]),
      ],
      output: `<App foo={ bar } { ...baz } />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 137 (7.37.5)',
      code: `<App foo={ bar} { ...baz} />;`,
      options: ['never'],
      errors: [
        expectNoSpaceAfterError([1, 10, 1, 11]),
        expectNoSpaceAfterError([1, 17, 1, 18]),
      ],
      output: `<App foo={bar} {...baz} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 138 (7.37.5)',
      code: `<App foo={bar } {...baz } />;`,
      options: ['never'],
      errors: [
        expectNoSpaceBeforeError([1, 15, 1, 16]),
        expectNoSpaceBeforeError([1, 25, 1, 26]),
      ],
      output: `<App foo={bar} {...baz} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 139 (7.37.5)',
      code: `
        <App foo={
        bar
        } {
        ...baz
        } />;
      `,
      options: ['never', { allowMultiline: false }],
      errors: [
        expectNoNewlineAfterError([2, 18, 2, 19]),
        expectNoNewlineBeforeError([4, 9, 4, 10]),
        expectNoNewlineAfterError([4, 11, 4, 12]),
        expectNoNewlineBeforeError([6, 9, 6, 10]),
      ],
      output: `
        <App foo={bar} {...baz} />;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 140 (7.37.5)',
      code: `
        <App foo={
        bar
        } {
        ...baz
        } />;
      `,
      options: ['always', { allowMultiline: false }],
      errors: [
        expectNoNewlineAfterError([2, 18, 2, 19]),
        expectNoNewlineBeforeError([4, 9, 4, 10]),
        expectNoNewlineAfterError([4, 11, 4, 12]),
        expectNoNewlineBeforeError([6, 9, 6, 10]),
      ],
      output: `
        <App foo={ bar } { ...baz } />;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 141 (7.37.5)',
      code: `<App foo={ 3 } bar={{a: 2}} />`,
      options: ['never', { spacing: { objectLiterals: 'always' } }],
      errors: [
        expectNoSpaceAfterError([1, 10, 1, 11]),
        expectNoSpaceBeforeError([1, 14, 1, 15]),
        expectSpaceNeededAfterError([1, 20, 1, 21]),
        expectSpaceNeededBeforeError([1, 27, 1, 28]),
      ],
      output: `<App foo={3} bar={ {a: 2} } />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 142 (7.37.5)',
      code: `<App foo={foo /* comment 22 */} />`,
      options: ['always'],
      errors: [
        expectSpaceNeededAfterError([1, 10, 1, 11]),
        expectSpaceNeededBeforeError([1, 31, 1, 32]),
      ],
      output: `<App foo={ foo /* comment 22 */ } />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 143 (7.37.5)',
      code: `<App foo={/* comment 23 */ foo} />`,
      options: ['always'],
      errors: [
        expectSpaceNeededAfterError([1, 10, 1, 11]),
        expectSpaceNeededBeforeError([1, 31, 1, 32]),
      ],
      output: `<App foo={ /* comment 23 */ foo } />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 144 (7.37.5)',
      code: `<App>{/*comment24*/ }</App>`,
      options: [{ children: { when: 'never' } }],
      errors: [expectNoSpaceBeforeError([1, 21, 1, 22])],
      output: `<App>{/*comment24*/}</App>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 145 (7.37.5)',
      code: `<App>{ /*comment25*/}</App>`,
      options: [{ children: { when: 'never' } }],
      errors: [expectNoSpaceAfterError([1, 6, 1, 7])],
      output: `<App>{/*comment25*/}</App>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 146 (7.37.5)',
      code: `<App>{/*comment26*/}</App>`,
      options: [{ children: { when: 'always' } }],
      errors: [
        expectSpaceNeededAfterError([1, 6, 1, 7]),
        expectSpaceNeededBeforeError([1, 20, 1, 21]),
      ],
      output: `<App>{ /*comment26*/ }</App>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 147 (7.37.5)',
      code: `
        <App>
        { /* comment 27 */ }
        </App>;
      `,
      options: [{ when: 'never', children: true }],
      errors: [
        expectNoSpaceAfterError([3, 9, 3, 10]),
        expectNoSpaceBeforeError([3, 28, 3, 29]),
      ],
      output: `
        <App>
        {/* comment 27 */}
        </App>;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 148 (7.37.5)',
      code: `
        <App>
        {/* comment 28 */}
        </App>;
      `,
      options: [{ when: 'always', children: true }],
      errors: [
        expectSpaceNeededAfterError([3, 9, 3, 10]),
        expectSpaceNeededBeforeError([3, 26, 3, 27]),
      ],
      output: `
        <App>
        { /* comment 28 */ }
        </App>;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 149 (7.37.5)',
      code: `
        <App>
        {/*comment29*/
        }
        </App>
      `,
      options: [{ children: { when: 'never', allowMultiline: false } }],
      errors: [expectNoNewlineBeforeError([4, 9, 4, 10])],
      output: `
        <App>
        {/*comment29*/}
        </App>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 150 (7.37.5)',
      code: `
        <App>
        {
        /*comment30*/}
        </App>
      `,
      options: [{ children: { when: 'never', allowMultiline: false } }],
      errors: [expectNoNewlineAfterError([3, 9, 3, 10])],
      output: `
        <App>
        {/*comment30*/}
        </App>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 151 (7.37.5)',
      code: `
        <App>{ /* comment 31 */
        bar
        } {
        baz
        /* comment 32 */ }</App>;
      `,
      options: [{ when: 'never', children: true }],
      errors: [
        expectNoSpaceAfterError([2, 14, 2, 15]),
        expectNoSpaceBeforeError([6, 26, 6, 27]),
      ],
      output: `
        <App>{/* comment 31 */
        bar
        } {
        baz
        /* comment 32 */}</App>;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 152 (7.37.5)',
      code: `
        <App>{/* comment 33 */
        bar
        } {
        baz
        /* comment 33 */}</App>;
      `,
      options: [{ when: 'always', children: true }],
      errors: [
        expectSpaceNeededAfterError([2, 14, 2, 15]),
        expectSpaceNeededBeforeError([6, 25, 6, 26]),
      ],
      output: `
        <App>{ /* comment 33 */
        bar
        } {
        baz
        /* comment 33 */ }</App>;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 153 (7.37.5)',
      code: `
        <div className={ this.state.renderInfo ? "infoPanel col-xs-12" : "unToggled col-xs-12" } />
      `,
      options: ['never', { allowMultiline: true }],
      errors: [
        expectNoSpaceAfterError([2, 24, 2, 25]),
        expectNoSpaceBeforeError([2, 96, 2, 97]),
      ],
      output: `
        <div className={this.state.renderInfo ? "infoPanel col-xs-12" : "unToggled col-xs-12"} />
      `,
      languageOptions: upstreamModuleOptions,
    },
  ],
});
