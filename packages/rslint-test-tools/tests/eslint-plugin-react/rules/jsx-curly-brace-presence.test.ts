import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-curly-brace-presence', {} as never, {
  valid: [
    // ---- Defaults ----
    { code: `<App {...props}>foo</App>` },
    { code: `<>foo</>` },
    { code: `<App {...props}>foo</App>`, options: [{ props: 'never' }] },

    // ---- Whitespace expressions are always allowed ----
    { code: `<App>{' '}</App>` },
    { code: `<App>{'     '}</App>` },
    { code: `<App>{' '}</App>`, options: [{ children: 'never' }] },
    { code: `<App>{' '}</App>`, options: [{ children: 'always' }] },

    // ---- Templates with substitutions stay wrapped ----
    {
      code: '<App>{`Hello ${word} World`}</App>',
      options: [{ children: 'never' }],
    },
    {
      code: '<App prop={`foo ${word} bar`} />',
      options: [{ props: 'never' }],
    },
    { code: '<App label={`${label}`} />', options: ['never'] },

    // ---- always-children allows braces around JSX child elements ----
    {
      code: `<App>{<myApp></myApp>}</App>`,
      options: [{ children: 'always' }],
    },

    // ---- never-children: identifiers, arrays, booleans pass through ----
    { code: `<App>{[]}</App>` },
    { code: `<App>foo</App>` },
    { code: `<App prop={true}>foo</App>` },
    { code: `<App prop>foo</App>` },

    // ---- Per-option matrix ----
    {
      code: `<MyComponent prop='bar'>foo</MyComponent>`,
      options: [{ props: 'never' }],
    },
    {
      code: `<MyComponent>foo</MyComponent>`,
      options: [{ children: 'never' }],
    },
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
      code: `<MyComponent prop={'bar'}>foo</MyComponent>`,
      options: [{ props: 'ignore' }],
    },
    {
      code: `<MyComponent prop={'bar'}>{'foo'}</MyComponent>`,
      options: ['always'],
    },
    {
      code: `<MyComponent prop="bar" attr='foo' />`,
      options: ['never'],
    },

    // ---- never-children: literals containing JSX-disallowed chars stay wrapped ----
    {
      code: `<MyComponent>{"div { margin-top: 0; }"}</MyComponent>`,
      options: ['never'],
    },
    {
      code: `<MyComponent>{"<Foo />"}</MyComponent>`,
      options: ['never'],
    },

    // ---- HTML entities preserved ----
    {
      code: `<MyComponent prop={"Hello &middot; world"}>bar</MyComponent>`,
      options: ['never'],
    },
    {
      code: `<MyComponent>{"Hello &middot; world"}</MyComponent>`,
      options: ['never'],
    },

    // ---- Trailing whitespace / leading whitespace strings ----
    {
      code: `<MyComponent>{"space after "}</MyComponent>`,
      options: ['never'],
    },
    {
      code: `<MyComponent>{" space before"}</MyComponent>`,
      options: ['never'],
    },

    // ---- propElementValues default ('ignore'): JSX prop values pass ----
    { code: `<MyComponent p={<Foo>Bar</Foo>} />` },
    { code: `<App horror={<div />} />` },
    {
      code: `<App horror={<div />} />`,
      options: [{ propElementValues: 'ignore' }],
    },

    // ---- Comments inside JSXExpression: braces never removed ----
    { code: `<App>{/* comment */}</App>` },
    { code: `<App>{/* comment */ <Foo />}</App>` },
    { code: `<App>{/* comment */ 'foo'}</App>` },
    { code: `<App prop={/* comment */ 'foo'} />` },

    // ---- script-like child: never-children leaves the template alone ----
    {
      code: '<script>{`window.foo = "bar"`}</script>',
    },
  ],
  invalid: [
    // ---- Unnecessary curly: template literal in props ----
    {
      code: '<App prop={`foo`} />',
      options: [{ props: 'never' }],
      output: '<App prop="foo" />',
      errors: [{ messageId: 'unnecessaryCurly' }],
    },
    // ---- Unnecessary curly: JSX element child wrapped in `{}` ----
    {
      code: `<App>{<myApp></myApp>}</App>`,
      options: [{ children: 'never' }],
      output: `<App><myApp></myApp></App>`,
      errors: [{ messageId: 'unnecessaryCurly' }],
    },
    {
      code: `<App>{<myApp></myApp>}</App>`,
      output: `<App><myApp></myApp></App>`,
      errors: [{ messageId: 'unnecessaryCurly' }],
    },
    // ---- Children template literal collapsed ----
    {
      code: '<App>{`foo`}</App>',
      options: [{ children: 'never' }],
      output: '<App>foo</App>',
      errors: [{ messageId: 'unnecessaryCurly' }],
    },
    {
      code: '<>{`foo`}</>',
      options: [{ children: 'never' }],
      output: '<>foo</>',
      errors: [{ messageId: 'unnecessaryCurly' }],
    },
    {
      code: `<MyComponent>{'foo'}</MyComponent>`,
      output: `<MyComponent>foo</MyComponent>`,
      errors: [{ messageId: 'unnecessaryCurly' }],
    },
    {
      code: `<MyComponent prop={'bar'}>foo</MyComponent>`,
      output: `<MyComponent prop="bar">foo</MyComponent>`,
      errors: [{ messageId: 'unnecessaryCurly' }],
    },

    // ---- Missing curly: props=always ----
    {
      code: `<MyComponent prop='bar'>foo</MyComponent>`,
      options: [{ props: 'always' }],
      output: `<MyComponent prop={"bar"}>foo</MyComponent>`,
      errors: [{ messageId: 'missingCurly' }],
    },
    {
      code: `<MyComponent prop="foo 'bar'">foo</MyComponent>`,
      options: [{ props: 'always' }],
      output: `<MyComponent prop={"foo 'bar'"}>foo</MyComponent>`,
      errors: [{ messageId: 'missingCurly' }],
    },
    {
      code: `<MyComponent prop='foo "bar"'>foo</MyComponent>`,
      options: [{ props: 'always' }],
      output: `<MyComponent prop={"foo \\"bar\\""}>foo</MyComponent>`,
      errors: [{ messageId: 'missingCurly' }],
    },

    // ---- Missing curly: children=always ----
    {
      code: `<MyComponent>foo bar </MyComponent>`,
      options: [{ children: 'always' }],
      output: `<MyComponent>{"foo bar "}</MyComponent>`,
      errors: [{ messageId: 'missingCurly' }],
    },
    {
      code: `<MyComponent>foo bar 'foo'</MyComponent>`,
      options: [{ children: 'always' }],
      output: `<MyComponent>{"foo bar 'foo'"}</MyComponent>`,
      errors: [{ messageId: 'missingCurly' }],
    },
    {
      code: '<MyComponent>foo bar "foo"</MyComponent>',
      options: [{ children: 'always' }],
      output: `<MyComponent>{"foo bar \\"foo\\""}</MyComponent>`,
      errors: [{ messageId: 'missingCurly' }],
    },
    {
      code: '<MyComponent>foo bar <App/></MyComponent>',
      options: [{ children: 'always' }],
      output: `<MyComponent>{"foo bar "}<App/></MyComponent>`,
      errors: [{ messageId: 'missingCurly' }],
    },

    // ---- Two-prop attribute fixes ----
    {
      code: `<App prop={'foo'} attr={" foo "} />`,
      options: [{ props: 'never' }],
      output: `<App prop="foo" attr=" foo " />`,
      errors: [
        { messageId: 'unnecessaryCurly' },
        { messageId: 'unnecessaryCurly' },
      ],
    },
    {
      code: `<App prop='foo' attr="bar" />`,
      options: [{ props: 'always' }],
      output: `<App prop={"foo"} attr={"bar"} />`,
      errors: [{ messageId: 'missingCurly' }, { messageId: 'missingCurly' }],
    },

    // ---- HTML entities + always-children: whole text wraps ----
    {
      code: `<App>foo &middot; bar</App>`,
      options: [{ children: 'always' }],
      output: `<App>{"foo &middot; bar"}</App>`,
      errors: [{ messageId: 'missingCurly' }],
    },
    {
      code: `<App prop='foo &middot; bar' />`,
      options: [{ props: 'always' }],
      output: `<App prop={"foo &middot; bar"} />`,
      errors: [{ messageId: 'missingCurly' }],
    },

    // ---- Quote payload, never-children unwraps ----
    {
      code: `<App>{'foo "bar"'}</App>`,
      options: [{ children: 'never' }],
      output: `<App>foo "bar"</App>`,
      errors: [{ messageId: 'unnecessaryCurly' }],
    },
    {
      code: `<App>{"foo 'bar'"}</App>`,
      options: [{ children: 'never' }],
      output: `<App>foo 'bar'</App>`,
      errors: [{ messageId: 'unnecessaryCurly' }],
    },

    // ---- propElementValues=never collapses braces around JSX prop value ----
    {
      code: `<App horror={<div />} />`,
      options: [
        { props: 'never', children: 'never', propElementValues: 'never' },
      ],
      output: `<App horror=<div /> />`,
      errors: [{ messageId: 'unnecessaryCurly' }],
    },

    // ---- Quote-only payload ----
    {
      code: `<Foo bar={"'"} />`,
      options: [
        { props: 'never', children: 'never', propElementValues: 'never' },
      ],
      output: `<Foo bar="'" />`,
      errors: [{ messageId: 'unnecessaryCurly' }],
    },
  ],
});
