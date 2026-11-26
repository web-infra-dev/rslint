import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';

function expectNewLineError1(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'newLine',
    message: 'Prop `bar` must be placed on a new line',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectNewLineError2(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'newLine',
    message: 'Prop `baz` must be placed on a new line',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectNewLineError3(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'newLine',
    message: 'Prop `spread` must be placed on a new line',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('jsx-max-props-per-line', {
  valid: [
    {
      code: `<Foo bar="baz" />`,
    },
    {
      code: `<Foo bar="baz" qux="quux" />`,
      options: [{ maximum: 2 }],
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
      code: `<App foo bar />`,
      options: [{ maximum: 2 }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `<App foo bar />`,
      options: [{ when: 'multiline' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `<App foo {...this.props} />`,
      options: [{ when: 'multiline' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `<App foo bar baz />`,
      options: [{ maximum: 2, when: 'multiline' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `<App {...this.props} bar />`,
      options: [{ maximum: 2 }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `
        <App
          foo
          bar
        />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `
        <App
          foo bar
          baz
        />
      `,
      options: [{ maximum: 2 }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `
        <App
          foo bar
          baz
        />
      `,
      options: [{ maximum: { multi: 2 } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 10 (7.37.5)',
      code: `
        <App
          bar
          baz
        />
      `,
      options: [{ maximum: { multi: 2, single: 1 } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: `<App foo baz bar />`,
      options: [{ maximum: { multi: 2, single: 3 } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 12 (7.37.5)',
      code: `<App {...this.props} bar />`,
      options: [{ maximum: { single: 2 } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 13 (7.37.5)',
      code: `
        <App
          foo bar
          baz bor
        />
      `,
      options: [{ maximum: { multi: 2, single: 1 } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 14 (7.37.5)',
      code: `<App foo baz bar />`,
      options: [{ maximum: { multi: 2 } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 15 (7.37.5)',
      code: `
        <App
          foo bar
          baz bor
        />
      `,
      options: [{ maximum: { single: 1 } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 16 (7.37.5)',
      code: `
        <App foo bar
          baz bor
        />
      `,
      options: [{ maximum: { single: 2, multi: 2 } }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 17 (7.37.5)',
      code: `
        <App foo bar
          baz bor
        />
      `,
      options: [{ maximum: 2 }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 18 (7.37.5)',
      code: `
        <App foo
          bar
        />
      `,
      options: [{ maximum: 1, when: 'multiline' }],
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    {
      code: `<Foo bar="baz" qux="quux" />`,
      errors: [
        {
          messageId: 'newLine',
          message: 'Prop `qux` must be placed on a new line',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 26,
        },
      ],
      output: `<Foo bar="baz"
qux="quux" />`,
    },
    {
      // ESLint reports one error per line at the first excess prop
      code: `<Foo bar="baz" qux="quux" abc="def" />`,
      options: [{ maximum: 1 }],
      errors: [
        {
          messageId: 'newLine',
          message: 'Prop `qux` must be placed on a new line',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 26,
        },
      ],
      output: `<Foo bar="baz"
qux="quux"
abc="def" />`,
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
      code: `
        <App foo bar baz />;
      `,
      errors: [expectNewLineError1([2, 18, 2, 21])],
      output: `
        <App foo
bar
baz />;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `
        <App foo bar baz />;
      `,
      options: [{ maximum: 2 }],
      errors: [expectNewLineError2([2, 22, 2, 25])],
      output: `
        <App foo bar
baz />;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `
        <App {...this.props} bar />;
      `,
      errors: [expectNewLineError1([2, 30, 2, 33])],
      output: `
        <App {...this.props}
bar />;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `
        <App bar {...this.props} />;
      `,
      errors: [expectNewLineError3([2, 18, 2, 33])],
      output: `
        <App bar
{...this.props} />;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `
        <App
          foo bar
          baz
        />
      `,
      errors: [expectNewLineError1([3, 15, 3, 18])],
      output: `
        <App
          foo
bar
          baz
        />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `
        <App
          foo {...this.props}
          baz
        />
      `,
      errors: [expectNewLineError3([3, 15, 3, 30])],
      output: `
        <App
          foo
{...this.props}
          baz
        />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `
        <App
          foo={{
          }} bar
        />
      `,
      errors: [expectNewLineError1([4, 14, 4, 17])],
      output: `
        <App
          foo={{
          }}
bar
        />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `
        <App foo={{
        }} bar />
      `,
      errors: [expectNewLineError1([3, 12, 3, 15])],
      output: `
        <App foo={{
        }}
bar />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `
        <App foo bar={{
        }} baz />
      `,
      options: [{ maximum: 2 }],
      errors: [expectNewLineError2([3, 12, 3, 15])],
      output: `
        <App foo bar={{
        }}
baz />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 9 (7.37.5)',
      code: `
        <App foo={{
        }} {...rest} />
      `,
      errors: [expectNewLineError3([3, 12, 3, 21])],
      output: `
        <App foo={{
        }}
{...rest} />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 10 (7.37.5)',
      code: `
        <App {
          ...this.props
        } bar />
      `,
      errors: [expectNewLineError1([4, 11, 4, 14])],
      output: `
        <App {
          ...this.props
        }
bar />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 11 (7.37.5)',
      code: `
        <App {
          ...this.props
        } {
          ...rest
        } />
      `,
      errors: [expectNewLineError3([4, 11, 6, 10])],
      output: `
        <App {
          ...this.props
        }
{
          ...rest
        } />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 12 (7.37.5)',
      code: `
        <App
          foo={{
          }} bar baz bor
        />
      `,
      options: [{ maximum: 2 }],
      errors: [expectNewLineError2([4, 18, 4, 21])],
      output: `
        <App
          foo={{
          }} bar
baz bor
        />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 13 (7.37.5)',
      code: `
        <App foo bar baz />
      `,
      options: [{ maximum: { single: 1, multi: 1 } }],
      errors: [expectNewLineError1([2, 18, 2, 21])],
      output: `
        <App foo
bar
baz />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 14 (7.37.5)',
      code: `
        <App
          foo bar baz
        />
      `,
      options: [{ maximum: { single: 1, multi: 1 } }],
      errors: [expectNewLineError1([3, 15, 3, 18])],
      output: `
        <App
          foo
bar
baz
        />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 15 (7.37.5)',
      code: `
        <App foo
          bar baz
        />
      `,
      options: [{ maximum: { single: 1, multi: 1 } }],
      errors: [expectNewLineError2([3, 15, 3, 18])],
      output: `
        <App foo
          bar
baz
        />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 16 (7.37.5)',
      code: `
        <App foo bar
          bar baz bor
        />
      `,
      options: [{ maximum: { single: 1, multi: 2 } }],
      errors: [
        {
          messageId: 'newLine',
          message: 'Prop `bor` must be placed on a new line',
          line: 3,
          column: 19,
          endLine: 3,
          endColumn: 22,
        },
      ],
      output: `
        <App foo bar
          bar baz
bor
        />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 17 (7.37.5)',
      code: `
        <App foo bar baz bor />
      `,
      options: [{ maximum: { single: 3, multi: 2 } }],
      errors: [
        {
          messageId: 'newLine',
          message: 'Prop `bor` must be placed on a new line',
          line: 2,
          column: 26,
          endLine: 2,
          endColumn: 29,
        },
      ],
      output: `
        <App foo bar
baz
bor />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 18 (7.37.5)',
      code: `
        <App
          foo={{
          }} bar baz bor
        />
      `,
      options: [{ maximum: { multi: 2 } }],
      errors: [expectNewLineError2([4, 18, 4, 21])],
      output: `
        <App
          foo={{
          }} bar
baz bor
        />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 19 (7.37.5)',
      code: `
        <App boz fuz
          foo={{
          }} bar baz bor
        />
      `,
      options: [{ maximum: { multi: 2, single: 1 } }],
      errors: [expectNewLineError2([4, 18, 4, 21])],
      output: `
        <App boz fuz
          foo={{
          }} bar
baz bor
        />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 20 (7.37.5)',
      code: `
        <DataTable<Items> fullscreen keyField="id" items={items}
          activeSortableColumn={sorting}
          onSortClick={handleSortedClick}
          rowActions={[
          ]}
        />
      `,
      options: [{ maximum: { multi: 1, single: 1 } }],
      errors: [
        {
          messageId: 'newLine',
          message: 'Prop `keyField` must be placed on a new line',
          line: 2,
          column: 38,
          endLine: 2,
          endColumn: 51,
        },
      ],
      output: `
        <DataTable<Items> fullscreen
keyField="id"
items={items}
          activeSortableColumn={sorting}
          onSortClick={handleSortedClick}
          rowActions={[
          ]}
        />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 21 (7.37.5)',
      code: `
        <DataTable<Items>
fullscreen keyField="id" items={items}
          activeSortableColumn={sorting}
          onSortClick={handleSortedClick}
          rowActions={[
          ]}
        />
      `,
      options: [{ maximum: { multi: 1, single: 1 } }],
      errors: [
        {
          messageId: 'newLine',
          message: 'Prop `keyField` must be placed on a new line',
          line: 3,
          column: 12,
          endLine: 3,
          endColumn: 25,
        },
      ],
      output: `
        <DataTable<Items>
fullscreen
keyField="id"
items={items}
          activeSortableColumn={sorting}
          onSortClick={handleSortedClick}
          rowActions={[
          ]}
        />
      `,
      languageOptions: upstreamModuleOptions,
    },
  ],
});
