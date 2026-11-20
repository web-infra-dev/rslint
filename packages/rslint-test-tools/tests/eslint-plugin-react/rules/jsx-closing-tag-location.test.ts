import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamES2015Options } from '../config';

function expectMatchIndentError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'matchIndent',
    message: 'Expected closing tag to match indentation of opening.',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectOnOwnLineError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'onOwnLine',
    message:
      'Closing tag of a multiline JSX expression must be on its own line.',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('jsx-closing-tag-location', {
  valid: [
    {
      code: `<Foo>bar</Foo>`,
    },
    {
      code: `<Foo>\n  bar\n</Foo>`,
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 0 (7.37.5)',
      code: `
        const foo = () => {
          return <App>
       bar</App>
        }
      `,
      options: [{ location: 'line-aligned' }],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 1 (7.37.5)',
      code: `
        const foo = () => {
          return <App>
              bar</App>
        }
      `,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 2 (7.37.5)',
      code: `
        const foo = () => {
          return <App>
              bar
          </App>
        }
      `,
      options: ['line-aligned'],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `
        const foo = <App>
              bar
        </App>
      `,
      options: ['line-aligned'],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `
        const x = <App>
              foo
                  </App>
      `,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `
        const foo =
          <App>
              bar
          </App>
      `,
      options: ['line-aligned'],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `
        const foo =
          <App>
              bar
          </App>
      `,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `
        <App>
          foo
        </App>
      `,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `
        <App>
          foo
        </App>
      `,
      options: ['line-aligned'],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `
        <App>foo</App>
      `,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 10 (7.37.5)',
      code: `
        <>
          foo
        </>
      `,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: `
        <>foo</>
      `,
      languageOptions: upstreamES2015Options,
    },
  ],
  invalid: [
    {
      code: `<Foo>\n  bar\n  </Foo>`,
      errors: [expectMatchIndentError([3, 3, 3, 9])],
      output: `<Foo>
  bar
</Foo>`,
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
      code: `
        <App>
          foo
          </App>
      `,
      errors: [expectMatchIndentError([4, 11, 4, 17])],
      output: `
        <App>
          foo
        </App>
      `,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `
        <App>
          foo</App>
      `,
      errors: [expectOnOwnLineError([3, 14, 3, 20])],
      output: `
        <App>
          foo
        </App>
      `,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `
        <>
          foo
          </>
      `,
      errors: [expectMatchIndentError([4, 11, 4, 14])],
      output: `
        <>
          foo
        </>
      `,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `
        <>
          foo</>
      `,
      errors: [expectOnOwnLineError([3, 14, 3, 17])],
      output: `
        <>
          foo
        </>
      `,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `
        const x = () => {
          return <App>
              foo</App>
        }
      `,
      options: ['line-aligned'],
      errors: [expectOnOwnLineError([4, 18, 4, 24])],
      output: `
        const x = () => {
          return <App>
              foo
          </App>
        }
      `,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `
        const x = <App>
              foo
                  </App>
      `,
      options: ['line-aligned'],
      errors: [
        {
          messageId: 'alignWithOpening',
          message:
            'Expected closing tag to be aligned with the line containing the opening tag',
          line: 4,
          column: 19,
          endLine: 4,
          endColumn: 25,
        },
      ],
      output: `
        const x = <App>
              foo
        </App>
      `,
      languageOptions: upstreamES2015Options,
    },
  ],
});
