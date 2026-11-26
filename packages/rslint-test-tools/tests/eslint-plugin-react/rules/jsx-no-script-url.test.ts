import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';

function expectNoScriptURLError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'noScriptURL',
    message:
      'A future version of React will block javascript: URLs as a security precaution. Use event handlers instead if you can. If you need to generate unsafe HTML, try using dangerouslySetInnerHTML instead.',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('jsx-no-script-url', {
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

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `<Foo href="javascript:"></Foo>`,
      settings: {
        linkComponents: [{ name: 'Foo', linkAttribute: ['to', 'href'] }],
      },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 10 (7.37.5)',
      code: `<Foo href="javascript:"></Foo>`,
      options: [[], { includeFromSettings: false }],
      settings: {
        linkComponents: [{ name: 'Foo', linkAttribute: ['to', 'href'] }],
      },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: `<Foo other="javascript:"></Foo>`,
      options: [[], { includeFromSettings: true }],
      settings: {
        linkComponents: [{ name: 'Foo', linkAttribute: ['to', 'href'] }],
      },
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    // ---- Upstream invalid cases ----
    {
      code: `<a href="javascript:"></a>`,
      errors: [expectNoScriptURLError([1, 4, 1, 22])],
    },
    {
      code: `<a href="javascript:void(0)"></a>`,
      errors: [expectNoScriptURLError([1, 4, 1, 29])],
    },
    {
      code: '<a href="j\n\n\na\rv\tascript:"></a>',
      errors: [expectNoScriptURLError([1, 4, 5, 12])],
    },
    // With component passed by options
    {
      code: `<Foo to="javascript:"></Foo>`,
      errors: [expectNoScriptURLError([1, 6, 1, 22])],
      options: [[{ name: 'Foo', props: ['to', 'href'] }]],
    },
    {
      code: `<Foo href="javascript:"></Foo>`,
      errors: [expectNoScriptURLError([1, 6, 1, 24])],
      options: [[{ name: 'Foo', props: ['to', 'href'] }]],
    },
    {
      code: `<a href="javascript:void(0)"></a>`,
      errors: [expectNoScriptURLError([1, 4, 1, 29])],
      options: [[{ name: 'Foo', props: ['to', 'href'] }]],
    },

    // ---- Additional edge cases ----
    // Case insensitive
    {
      code: `<a href="JAVASCRIPT:void(0)"></a>`,
      errors: [expectNoScriptURLError([1, 4, 1, 29])],
    },
    {
      code: `<a href="JavaScript:void(0)"></a>`,
      errors: [expectNoScriptURLError([1, 4, 1, 29])],
    },
    // Spaces before javascript:
    {
      code: `<a href="  javascript:"></a>`,
      errors: [expectNoScriptURLError([1, 4, 1, 24])],
    },
    // Namespaced tag — local part "a" matches default config
    {
      code: `<ns:a href="javascript:"></ns:a>`,
      errors: [expectNoScriptURLError([1, 7, 1, 25])],
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
      errors: [expectNoScriptURLError([1, 7, 1, 23])],
    },
    // Multiline
    {
      code: `<a\n\thref="javascript:void(0)">\n</a>`,
      errors: [expectNoScriptURLError([2, 2, 2, 27])],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `<Foo to="javascript:"></Foo>`,
      options: [
        [{ name: 'Bar', props: ['to', 'href'] }],
        { includeFromSettings: true },
      ],
      settings: { linkComponents: [{ name: 'Foo', linkAttribute: 'to' }] },
      errors: [expectNoScriptURLError([1, 6, 1, 22])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `<Foo href="javascript:"></Foo>`,
      options: [{ includeFromSettings: true }],
      settings: {
        linkComponents: [{ name: 'Foo', linkAttribute: ['to', 'href'] }],
      },
      errors: [expectNoScriptURLError([1, 6, 1, 24])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `
      <div>
        <Foo href="javascript:"></Foo>
        <Bar link="javascript:"></Bar>
      </div>
    `,
      options: [
        [{ name: 'Bar', props: ['link'] }],
        { includeFromSettings: true },
      ],
      settings: {
        linkComponents: [{ name: 'Foo', linkAttribute: ['to', 'href'] }],
      },
      errors: [
        expectNoScriptURLError([3, 14, 3, 32]),
        expectNoScriptURLError([4, 14, 4, 32]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 9 (7.37.5)',
      code: `
      <div>
        <Foo href="javascript:"></Foo>
        <Bar link="javascript:"></Bar>
      </div>
    `,
      options: [[{ name: 'Bar', props: ['link'] }]],
      settings: {
        linkComponents: [{ name: 'Foo', linkAttribute: ['to', 'href'] }],
      },
      errors: [expectNoScriptURLError([4, 14, 4, 32])],
      languageOptions: upstreamModuleOptions,
    },
  ],
});
