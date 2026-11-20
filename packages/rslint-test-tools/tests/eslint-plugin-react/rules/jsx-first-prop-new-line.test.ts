import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';
import { diagnostic } from '../expectations';

function expectPropOnNewLineError(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'propOnNewLine',
    'Property should be placed on a new line',
    range,
  );
}

function expectPropOnSameLineError(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'propOnSameLine',
    'Property should be placed on the same line as the component declaration',
    range,
  );
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('jsx-first-prop-new-line', {
  valid: [
    {
      code: `<Foo bar="baz" />`,
      options: ['never'],
    },
    {
      code: `<Foo\n  bar="baz"\n/>`,
      options: ['always'],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 0 (7.37.5)',
      code: `<Foo />`,
      options: ['never'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 1 (7.37.5)',
      code: `<Foo prop="bar" />`,
      options: ['never'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 2 (7.37.5)',
      code: `<Foo {...this.props} />`,
      options: ['never'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `<Foo a a a />`,
      options: ['never'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `
        <Foo a
          b
        />
      `,
      options: ['never'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `<Foo />`,
      options: ['multiline'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `<Foo prop="one" />`,
      options: ['multiline'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `<Foo {...this.props} />`,
      options: ['multiline'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `<Foo a a a />`,
      options: ['multiline'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `
        <Foo
          propOne="one"
          propTwo="two"
        />
      `,
      options: ['multiline'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 10 (7.37.5)',
      code: `
        <Foo
          {...this.props}
          propTwo="two"
        />
      `,
      options: ['multiline'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: `
        <Foo bar />
      `,
      options: ['multiline-multiprop'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 12 (7.37.5)',
      code: `
        <Foo bar baz />
      `,
      options: ['multiline-multiprop'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 13 (7.37.5)',
      code: `
        <Foo prop={{
        }} />
      `,
      options: ['multiline-multiprop'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 14 (7.37.5)',
      code: `
        <Foo
          foo={{
          }}
          bar
        />
      `,
      options: ['multiline-multiprop'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 15 (7.37.5)',
      code: `<Foo />`,
      options: ['always'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 16 (7.37.5)',
      code: `
        <Foo
          propOne="one"
          propTwo="two"
        />
      `,
      options: ['always'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 17 (7.37.5)',
      code: `
        <Foo
          {...this.props}
          propTwo="two"
        />
      `,
      options: ['always'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 18 (7.37.5)',
      code: `
        <Foo />
      `,
      options: ['multiprop'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 19 (7.37.5)',
      code: `
        <Foo bar />
      `,
      options: ['multiprop'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 20 (7.37.5)',
      code: `
        <Foo {...this.props} />
      `,
      options: ['multiprop'],
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    {
      code: `<Foo bar="baz"\n  qux="quux"\n/>`,
      options: ['always'],
      errors: [expectPropOnNewLineError([1, 6, 1, 15])],
      output: `<Foo
bar="baz"
  qux="quux"
/>`,
    },
    {
      code: `<Foo\n  bar="baz"\n/>`,
      options: ['never'],
      errors: [expectPropOnSameLineError([2, 3, 2, 12])],
      output: `<Foo bar="baz"
/>`,
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
      code: `
        <Foo propOne="one" propTwo="two" />
      `,
      options: ['always'],
      errors: [expectPropOnNewLineError([2, 14, 2, 27])],
      output: `
        <Foo
propOne="one" propTwo="two" />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `
        <Foo propOne="one"
          propTwo="two"
        />
      `,
      options: ['always'],
      errors: [expectPropOnNewLineError([2, 14, 2, 27])],
      output: `
        <Foo
propOne="one"
          propTwo="two"
        />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `
        <Foo
          propOne="one"
          propTwo="two"
        />
      `,
      options: ['never'],
      errors: [expectPropOnSameLineError([3, 11, 3, 24])],
      output: `
        <Foo propOne="one"
          propTwo="two"
        />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `
        <Foo prop={{
        }} />
      `,
      options: ['multiline'],
      errors: [expectPropOnNewLineError([2, 14, 3, 11])],
      output: `
        <Foo
prop={{
        }} />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `
        <Foo bar={{
        }} baz />
      `,
      options: ['multiline-multiprop'],
      errors: [expectPropOnNewLineError([2, 14, 3, 11])],
      output: `
        <Foo
bar={{
        }} baz />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `
      <Foo propOne="one" propTwo="two" />
      `,
      options: ['multiprop'],
      errors: [expectPropOnNewLineError([2, 12, 2, 25])],
      output: `
      <Foo
propOne="one" propTwo="two" />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `
      <Foo
bar />
      `,
      options: ['multiprop'],
      errors: [expectPropOnSameLineError([3, 1, 3, 4])],
      output: `
      <Foo bar />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `
      <Foo
{...this.props} />
      `,
      options: ['multiprop'],
      errors: [expectPropOnSameLineError([3, 1, 3, 16])],
      output: `
      <Foo {...this.props} />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `
        <DataTable<Items> fullscreen keyField="id" items={items}
          activeSortableColumn={sorting}
          onSortClick={handleSortedClick}
          rowActions={[
          ]}
        />
      `,
      options: ['multiline'],
      errors: [expectPropOnNewLineError([2, 27, 2, 37])],
      output: `
        <DataTable<Items>
fullscreen keyField="id" items={items}
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
