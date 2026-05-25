import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-no-script-url', {} as never, {
  valid: [
    // ---- Upstream valid cases ----
    { code: `<a href="https://reactjs.org"></a>` },
    { code: `<a href="mailto:foo@bar.com"></a>` },
    { code: `<a href="#"></a>` },
    { code: `<a href=""></a>` },
    { code: `<a name="foo"></a>` },
    { code: `<a href={"javascript:"}></a>` },
    { code: `<Foo href="javascript:"></Foo>` },
    { code: `<a href />` },
    {
      code: `<Foo other="javascript:"></Foo>`,
      options: [[{ name: 'Foo', props: ['to', 'href'] }]],
    },

    // ---- Additional edge cases ----
    // Template literal — not flagged
    { code: '<a href={`javascript:`}></a>' },
    // Member expression tag — not matched
    { code: `<Foo.Bar href="javascript:"></Foo.Bar>` },
    // Spread — not a JsxAttribute
    { code: `const x: any = {href: "javascript:"}; <a {...x}></a>;` },
    // Non-javascript protocol
    { code: `<a href="jot:something"></a>` },
  ],
  invalid: [
    // ---- Upstream invalid cases ----
    {
      code: `<a href="javascript:"></a>`,
      errors: [{ message: /javascript: URLs/ }],
    },
    {
      code: `<a href="javascript:void(0)"></a>`,
      errors: [{ message: /javascript: URLs/ }],
    },
    {
      code: '<a href="j\n\n\na\rv\tascript:"></a>',
      errors: [{ message: /javascript: URLs/ }],
    },
    // With component passed by options
    {
      code: `<Foo to="javascript:"></Foo>`,
      errors: [{ message: /javascript: URLs/ }],
      options: [[{ name: 'Foo', props: ['to', 'href'] }]],
    },
    {
      code: `<Foo href="javascript:"></Foo>`,
      errors: [{ message: /javascript: URLs/ }],
      options: [[{ name: 'Foo', props: ['to', 'href'] }]],
    },
    {
      code: `<a href="javascript:void(0)"></a>`,
      errors: [{ message: /javascript: URLs/ }],
      options: [[{ name: 'Foo', props: ['to', 'href'] }]],
    },

    // ---- Additional edge cases ----
    // Case insensitive
    {
      code: `<a href="JAVASCRIPT:void(0)"></a>`,
      errors: [{ message: /javascript: URLs/ }],
    },
    {
      code: `<a href="JavaScript:void(0)"></a>`,
      errors: [{ message: /javascript: URLs/ }],
    },
    // Spaces before javascript:
    {
      code: `<a href="  javascript:"></a>`,
      errors: [{ message: /javascript: URLs/ }],
    },
    // Namespaced tag — local part "a" matches default config
    {
      code: `<ns:a href="javascript:"></ns:a>`,
      errors: [{ message: /javascript: URLs/ }],
    },
    // Multiple custom components
    {
      code: `<Link to="javascript:"></Link>`,
      options: [
        [
          { name: 'Link', props: ['to'] },
          { name: 'Button', props: ['href'] },
        ],
      ],
      errors: [{ message: /javascript: URLs/ }],
    },
    // Multiline
    {
      code: `<a\n\thref="javascript:void(0)">\n</a>`,
      errors: [{ message: /javascript: URLs/ }],
    },
  ],
});
