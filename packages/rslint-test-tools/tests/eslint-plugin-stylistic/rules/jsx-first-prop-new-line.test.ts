import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const PROP_ON_NEW_LINE = 'Property should be placed on a new line';
const PROP_ON_SAME_LINE =
  'Property should be placed on the same line as the component declaration';

// Mirrors upstream packages/eslint-plugin/rules/jsx-first-prop-new-line/
// jsx-first-prop-new-line.test.ts (Layer 1). This JS suite verifies binary
// registration + wire protocol + ESLint-compatible diagnostics; edge-shape and
// branch lock-in cases live in the Go _extras_test.go.
ruleTester.run('jsx-first-prop-new-line', null as never, {
  valid: [
    // never
    { code: `<Foo />`, options: ['never'] },
    { code: `<Foo prop="bar" />`, options: ['never'] },
    { code: `<Foo {...this.props} />`, options: ['never'] },
    { code: `<Foo a a a />`, options: ['never'] },
    {
      code: `
        <Foo a
          b
        />
      `,
      options: ['never'],
    },
    // multiline
    { code: `<Foo />`, options: ['multiline'] },
    { code: `<Foo prop="one" />`, options: ['multiline'] },
    { code: `<Foo {...this.props} />`, options: ['multiline'] },
    { code: `<Foo a a a />`, options: ['multiline'] },
    {
      code: `
        <Foo
          propOne="one"
          propTwo="two"
        />
      `,
      options: ['multiline'],
    },
    {
      code: `
        <Foo
          {...this.props}
          propTwo="two"
        />
      `,
      options: ['multiline'],
    },
    // multiline-multiprop (default)
    {
      code: `
        <Foo bar />
      `,
      options: ['multiline-multiprop'],
    },
    {
      code: `
        <Foo bar baz />
      `,
      options: ['multiline-multiprop'],
    },
    {
      code: `
        <Foo prop={{
        }} />
      `,
      options: ['multiline-multiprop'],
    },
    {
      code: `
        <Foo
          foo={{
          }}
          bar
        />
      `,
      options: ['multiline-multiprop'],
    },
    // always
    { code: `<Foo />`, options: ['always'] },
    {
      code: `
        <Foo
          propOne="one"
          propTwo="two"
        />
      `,
      options: ['always'],
    },
    {
      code: `
        <Foo
          {...this.props}
          propTwo="two"
        />
      `,
      options: ['always'],
    },
    // multiprop
    {
      code: `
        <Foo />
      `,
      options: ['multiprop'],
    },
    {
      code: `
        <Foo bar />
      `,
      options: ['multiprop'],
    },
    {
      code: `
        <Foo {...this.props} />
      `,
      options: ['multiprop'],
    },
  ],

  invalid: [
    // always: single-line tag, first prop on same line
    {
      code: `
        <Foo propOne="one" propTwo="two" />
      `,
      output: `
        <Foo
propOne="one" propTwo="two" />
      `,
      options: ['always'],
      errors: [{ messageId: 'propOnNewLine', message: PROP_ON_NEW_LINE }],
    },
    // always: multiline tag, first prop on same line
    {
      code: `
        <Foo propOne="one"
          propTwo="two"
        />
      `,
      output: `
        <Foo
propOne="one"
          propTwo="two"
        />
      `,
      options: ['always'],
      errors: [{ messageId: 'propOnNewLine', message: PROP_ON_NEW_LINE }],
    },
    // never: first prop on new line
    {
      code: `
        <Foo
          propOne="one"
          propTwo="two"
        />
      `,
      output: `
        <Foo propOne="one"
          propTwo="two"
        />
      `,
      options: ['never'],
      errors: [{ messageId: 'propOnSameLine', message: PROP_ON_SAME_LINE }],
    },
    // multiline: prop on same line of a multiline tag
    {
      code: `
        <Foo prop={{
        }} />
      `,
      output: `
        <Foo
prop={{
        }} />
      `,
      options: ['multiline'],
      errors: [{ messageId: 'propOnNewLine', message: PROP_ON_NEW_LINE }],
    },
    // multiline-multiprop: multiline, multiple props, first on same line
    {
      code: `
        <Foo bar={{
        }} baz />
      `,
      output: `
        <Foo
bar={{
        }} baz />
      `,
      options: ['multiline-multiprop'],
      errors: [{ messageId: 'propOnNewLine', message: PROP_ON_NEW_LINE }],
    },
    // multiprop: multiple props on same line
    {
      code: `
      <Foo propOne="one" propTwo="two" />
      `,
      output: `
      <Foo
propOne="one" propTwo="two" />
      `,
      options: ['multiprop'],
      errors: [{ messageId: 'propOnNewLine', message: PROP_ON_NEW_LINE }],
    },
    // multiprop: single prop on new line in a multiline tag -> same line
    {
      code: `
      <Foo
bar />
      `,
      output: `
      <Foo bar />
      `,
      options: ['multiprop'],
      errors: [{ messageId: 'propOnSameLine', message: PROP_ON_SAME_LINE }],
    },
    // multiprop: single spread prop on new line -> same line
    {
      code: `
      <Foo
{...this.props} />
      `,
      output: `
      <Foo {...this.props} />
      `,
      options: ['multiprop'],
      errors: [{ messageId: 'propOnSameLine', message: PROP_ON_SAME_LINE }],
    },
    // multiline: TypeScript generic component (typeArguments fix anchor)
    {
      code: `
        <DataTable<Items> fullscreen keyField="id" items={items}
          activeSortableColumn={sorting}
          onSortClick={handleSortedClick}
          rowActions={[
          ]}
        />
      `,
      output: `
        <DataTable<Items>
fullscreen keyField="id" items={items}
          activeSortableColumn={sorting}
          onSortClick={handleSortedClick}
          rowActions={[
          ]}
        />
      `,
      options: ['multiline'],
      errors: [{ messageId: 'propOnNewLine', message: PROP_ON_NEW_LINE }],
    },
  ],
});
