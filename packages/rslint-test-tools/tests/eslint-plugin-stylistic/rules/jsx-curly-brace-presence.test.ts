import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

// The JS suite mirrors upstream @stylistic Layer 1 (registration + wire
// protocol + ESLint-compatible diagnostic shape). The full edge / branch / and
// @stylistic-vs-react quote-gate coverage lives in the Go suite
// (jsx_curly_brace_presence_upstream_test.go + _extras_test.go). Errors assert
// the user-facing message text — the rule-tester verifies diagnostic count,
// rule name, and message, but not messageId.
const UNNECESSARY = 'Curly braces are unnecessary here.';
const MISSING = 'Need to wrap this literal in a JSX expression.';

ruleTester.run('jsx-curly-brace-presence', null as never, {
  valid: [
    { code: `<App {...props}>foo</App>` },
    { code: `<>foo</>` },
    { code: `<App>{' '}</App>` },
    { code: `<App>{' '}</App>`, options: [{ children: 'never' }] },
    {
      code: '<App>{`Hello ${word} World`}</App>',
      options: [{ children: 'never' }],
    },
    { code: `<App>{[]}</App>` },
    { code: `<App>foo</App>` },
    { code: `<App prop='bar'>foo</App>` },
    { code: `<App prop={true}>foo</App>` },
    {
      code: `<MyComponent prop={'bar'}>foo</MyComponent>`,
      options: [{ props: 'always' }],
    },
    {
      code: `<MyComponent>{'foo'}</MyComponent>`,
      options: [{ children: 'always' }],
    },
    {
      code: `<MyComponent>{'foo'}</MyComponent>`,
      options: [{ children: 'ignore' }],
    },
    {
      code: `<MyComponent prop={'bar'}>{'foo'}</MyComponent>`,
      options: ['always'],
    },
    { code: `<MyComponent prop="bar" attr='foo' />`, options: ['never'] },
    { code: `<MyComponent>{"<Foo />"}</MyComponent>`, options: ['never'] },
    {
      code: `<MyComponent>{"Hello &middot; world"}</MyComponent>`,
      options: ['never'],
    },
    { code: `<App>{/* comment */}</App>` },
    { code: `<App>{/* comment */ 'foo'}</App>` },
    // propElementValues defaults to 'ignore': JSX element prop value passes.
    { code: `<App horror={<div />} />` },
  ],
  invalid: [
    {
      code: '<App prop={`foo`} />',
      options: [{ props: 'never' }],
      output: '<App prop="foo" />',
      errors: [{ message: UNNECESSARY }],
    },
    {
      code: '<App>{<myApp></myApp>}</App>',
      options: [{ children: 'never' }],
      output: '<App><myApp></myApp></App>',
      errors: [{ message: UNNECESSARY }],
    },
    {
      code: `<MyComponent>{'foo'}</MyComponent>`,
      output: '<MyComponent>foo</MyComponent>',
      errors: [{ message: UNNECESSARY }],
    },
    {
      code: `<MyComponent prop={'bar'}>foo</MyComponent>`,
      options: [{ props: 'never' }],
      output: `<MyComponent prop="bar">foo</MyComponent>`,
      errors: [{ message: UNNECESSARY }],
    },
    // Quote-bearing children literal still reports under 'never'.
    {
      code: `<App>{'foo "bar"'}</App>`,
      options: [{ children: 'never' }],
      output: `<App>foo "bar"</App>`,
      errors: [{ message: UNNECESSARY }],
    },
    {
      code: `<MyComponent prop='bar'>foo</MyComponent>`,
      options: [{ props: 'always' }],
      output: '<MyComponent prop={"bar"}>foo</MyComponent>',
      errors: [{ message: MISSING }],
    },
    {
      code: '<MyComponent>foo bar </MyComponent>',
      options: [{ children: 'always' }],
      output: `<MyComponent>{"foo bar "}</MyComponent>`,
      errors: [{ message: MISSING }],
    },
    {
      code: `<MyComponent prop={'bar'}>{'foo'}</MyComponent>`,
      options: ['never'],
      output: '<MyComponent prop="bar">foo</MyComponent>',
      errors: [{ message: UNNECESSARY }, { message: UNNECESSARY }],
    },
    {
      code: `<MyComponent prop='bar'>foo</MyComponent>`,
      options: ['always'],
      output: '<MyComponent prop={"bar"}>{"foo"}</MyComponent>',
      errors: [{ message: MISSING }, { message: MISSING }],
    },
  ],
});
