import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

// Mirrors aria-query's `dom.keys()` in order. Used to drive procedural test
// generation that mirrors upstream's
// `__tests__/src/rules/aria-unsupported-elements-test.js`.
const allDomElements = [
  'a',
  'abbr',
  'acronym',
  'address',
  'applet',
  'area',
  'article',
  'aside',
  'audio',
  'b',
  'base',
  'bdi',
  'bdo',
  'big',
  'blink',
  'blockquote',
  'body',
  'br',
  'button',
  'canvas',
  'caption',
  'center',
  'cite',
  'code',
  'col',
  'colgroup',
  'content',
  'data',
  'datalist',
  'dd',
  'del',
  'details',
  'dfn',
  'dialog',
  'dir',
  'div',
  'dl',
  'dt',
  'em',
  'embed',
  'fieldset',
  'figcaption',
  'figure',
  'font',
  'footer',
  'form',
  'frame',
  'frameset',
  'h1',
  'h2',
  'h3',
  'h4',
  'h5',
  'h6',
  'head',
  'header',
  'hgroup',
  'hr',
  'html',
  'i',
  'iframe',
  'img',
  'input',
  'ins',
  'kbd',
  'keygen',
  'label',
  'legend',
  'li',
  'link',
  'main',
  'map',
  'mark',
  'marquee',
  'menu',
  'menuitem',
  'meta',
  'meter',
  'nav',
  'noembed',
  'noscript',
  'object',
  'ol',
  'optgroup',
  'option',
  'output',
  'p',
  'param',
  'picture',
  'pre',
  'progress',
  'q',
  'rp',
  'rt',
  'rtc',
  'ruby',
  's',
  'samp',
  'script',
  'section',
  'select',
  'small',
  'source',
  'spacer',
  'span',
  'strike',
  'strong',
  'style',
  'sub',
  'summary',
  'sup',
  'table',
  'tbody',
  'td',
  'textarea',
  'tfoot',
  'th',
  'thead',
  'time',
  'title',
  'tr',
  'track',
  'tt',
  'u',
  'ul',
  'var',
  'video',
  'wbr',
  'xmp',
];

const reservedSet = new Set([
  'base',
  'col',
  'colgroup',
  'head',
  'html',
  'link',
  'meta',
  'noembed',
  'noscript',
  'param',
  'picture',
  'script',
  'source',
  'style',
  'title',
  'track',
]);

const roleValidityTests = allDomElements.map((el) => ({
  code: `<${el} ${reservedSet.has(el) ? '' : 'role'} />`,
}));

const ariaValidityTests = allDomElements
  .map((el) => ({
    code: `<${el} ${reservedSet.has(el) ? '' : 'aria-hidden'} />`,
  }))
  .concat({ code: '<fake aria-hidden />' });

new RuleTester(config).run('aria-unsupported-elements', {
  valid: [
    ...roleValidityTests,
    ...ariaValidityTests,
    // Extra rslint lockdowns:
    { code: '<Custom aria-hidden />' },
    { code: '<base {...props} />' },
    { code: '<base data-foo="x" />' },
    { code: '<meta charset="UTF-8" />' },
    { code: '<div role="button" aria-pressed="true" />' },
    { code: '<base aria:hidden />' },
    {
      code: '<Title aria-hidden />',
      settings: { 'jsx-a11y': { components: { Title: 'div' } } },
    },
    // Member-expression / namespaced / custom-element tag names.
    { code: '<Foo.base aria-hidden />' },
    { code: '<lib.Title aria-hidden />' },
    { code: '<svg:title aria-hidden />' },
    { code: '<my-element aria-hidden />' },
    { code: '<image aria-hidden />' },
    // Comments between attributes.
    { code: '<base /* leading */ data-foo="x" /* trailing */ />' },
    // Spread of literal containing aria-* (upstream limitation).
    { code: "<base {...{'aria-hidden': true}} />" },
    // Polymorphic with non-literal / falsy / disallowed shapes.
    {
      code: '<Box as={dynamicTag} aria-hidden />',
      settings: { 'jsx-a11y': { polymorphicPropName: 'as' } },
    },
    {
      code: '<Box as="" aria-hidden />',
      settings: { 'jsx-a11y': { polymorphicPropName: 'as' } },
    },
    {
      code: '<Box as={false} aria-hidden />',
      settings: { 'jsx-a11y': { polymorphicPropName: 'as' } },
    },
    {
      code: '<Box as="meta" aria-hidden />',
      settings: {
        'jsx-a11y': {
          polymorphicPropName: 'as',
          polymorphicAllowList: ['OtherComponent'],
        },
      },
    },
    // Settings shape edge cases.
    { code: '<Foo aria-hidden />', settings: { 'jsx-a11y': {} } },
    {
      code: '<Foo aria-hidden />',
      settings: { 'jsx-a11y': { components: { Foo: 123 } } },
    },
    {
      code: '<Foo aria-hidden />',
      settings: { 'jsx-a11y': { components: { Bar: 'meta' } } },
    },
    // JSX inside hook callback.
    {
      code: 'function F() { useEffect(() => { renderer(<div role="button" />); }); }',
    },
    // Reserved element without aria/role.
    { code: '<base href="https://example.com" target="_blank" />' },
    { code: '<base key="x" />' },
  ],
  invalid: [
    {
      code: '<base role {...props} />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'role'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<col role {...props} />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'role'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<colgroup role {...props} />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'role'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<head role {...props} />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'role'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<html role {...props} />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'role'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<link role {...props} />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'role'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<meta role {...props} />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'role'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<noembed role {...props} />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'role'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: '<noscript role {...props} />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'role'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<param role {...props} />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'role'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<picture role {...props} />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'role'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: '<script role {...props} />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'role'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: '<source role {...props} />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'role'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: '<style role {...props} />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'role'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<title role {...props} />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'role'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<track role {...props} />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'role'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<Meta aria-hidden />',
      settings: {
        'jsx-a11y': {
          components: {
            Meta: 'meta',
          },
        },
      },
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'aria-hidden'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<base aria-hidden aria-role="none" {...props} />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'aria-hidden'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 49,
        },
      ],
    },
    {
      code: '<col aria-hidden aria-role="none" {...props} />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'aria-hidden'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },
    {
      code: '<colgroup aria-hidden aria-role="none" {...props} />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'aria-hidden'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
    {
      code: '<head aria-hidden aria-role="none" {...props} />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'aria-hidden'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 49,
        },
      ],
    },
    {
      code: '<html aria-hidden aria-role="none" {...props} />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'aria-hidden'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 49,
        },
      ],
    },
    {
      code: '<link aria-hidden aria-role="none" {...props} />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'aria-hidden'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 49,
        },
      ],
    },
    {
      code: '<meta aria-hidden aria-role="none" {...props} />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'aria-hidden'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 49,
        },
      ],
    },
    {
      code: '<noembed aria-hidden aria-role="none" {...props} />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'aria-hidden'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 52,
        },
      ],
    },
    {
      code: '<noscript aria-hidden aria-role="none" {...props} />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'aria-hidden'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
    {
      code: '<param aria-hidden aria-role="none" {...props} />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'aria-hidden'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
    {
      code: '<picture aria-hidden aria-role="none" {...props} />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'aria-hidden'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 52,
        },
      ],
    },
    {
      code: '<script aria-hidden aria-role="none" {...props} />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'aria-hidden'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },
    {
      code: '<source aria-hidden aria-role="none" {...props} />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'aria-hidden'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },
    {
      code: '<style aria-hidden aria-role="none" {...props} />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'aria-hidden'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
    {
      code: '<title aria-hidden aria-role="none" {...props} />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'aria-hidden'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
    {
      code: '<track aria-hidden aria-role="none" {...props} />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'aria-hidden'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
    {
      code: '<base ROLE />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'role'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: '<base ARIA-HIDDEN />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'aria-hidden'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<base role aria-hidden aria-label="x" />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'role'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
        },
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'aria-hidden'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
        },
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'aria-label'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      code: '<style aria-hidden></style>',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'aria-hidden'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<Box as="base" aria-hidden />',
      settings: {
        'jsx-a11y': {
          polymorphicPropName: 'as',
        },
      },
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'aria-hidden'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '<Foo as="meta" role="none" />',
      settings: {
        'jsx-a11y': {
          polymorphicPropName: 'as',
        },
      },
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'role'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '<Foo aria-hidden />',
      settings: {
        'jsx-a11y': {
          components: {
            Foo: 'meta',
          },
        },
      },
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'aria-hidden'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<div><base aria-hidden /></div>',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'aria-hidden'.",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<head role="x"><base aria-hidden /></head>',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'role'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
        },
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'aria-hidden'.",
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      code: '<base href="x" target="_blank" aria-hidden />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'aria-hidden'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 46,
        },
      ],
    },
    {
      code: '<base {...a} role {...b} aria-hidden {...c} />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'role'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 47,
        },
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'aria-hidden'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 47,
        },
      ],
    },
    {
      code: '<base role />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'role'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: '<>{cond && <title aria-hidden>x</title>}</>',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'aria-hidden'.",
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<base aria-label />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'aria-label'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<base aria-hidden>x</base>',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'aria-hidden'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<base aria-hidden/>',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'aria-hidden'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: 'function L({xs}) { return xs.map(x => <base aria-hidden key={x} />); }',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'aria-hidden'.",
          line: 1,
          column: 39,
          endLine: 1,
          endColumn: 67,
        },
      ],
    },
    {
      code: 'class C { render() { return <link aria-hidden />; } }',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'aria-hidden'.",
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 49,
        },
      ],
    },
    {
      code: 'function F() { return <div>{cond && <meta aria-hidden />}</div>; }',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'aria-hidden'.",
          line: 1,
          column: 37,
          endLine: 1,
          endColumn: 57,
        },
      ],
    },
    {
      code: '<base aria-hidden={true as boolean} />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'aria-hidden'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      code: '<base aria-label={"x" as const} />',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'aria-label'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: 'function f<T>(x: T) { return <base aria-hidden />; }',
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'aria-hidden'.",
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
    {
      code: '<Box as="meta" aria-hidden />',
      settings: {
        'jsx-a11y': {
          polymorphicPropName: 'as',
          polymorphicAllowList: ['Box'],
        },
      },
      errors: [
        {
          messageId: 'unsupported',
          message:
            "This element does not support ARIA roles, states and properties. Try removing the prop 'aria-hidden'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
  ],
});
