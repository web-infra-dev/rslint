import { RuleTester } from '../rule-tester';

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

const errorMessage = (invalidProp: string) =>
  `This element does not support ARIA roles, states and properties. Try removing the prop '${invalidProp}'.`;

const roleValidityTests = allDomElements.map((el) => ({
  code: `<${el} ${reservedSet.has(el) ? '' : 'role'} />`,
}));

const ariaValidityTests = allDomElements
  .map((el) => ({
    code: `<${el} ${reservedSet.has(el) ? '' : 'aria-hidden'} />`,
  }))
  .concat({ code: '<fake aria-hidden />' });

const invalidRoleValidityTests = allDomElements
  .filter((el) => reservedSet.has(el))
  .map((el) => ({
    code: `<${el} role {...props} />`,
    errors: [{ message: errorMessage('role') }],
  }))
  .concat({
    code: '<Meta aria-hidden />',
    errors: [{ message: errorMessage('aria-hidden') }],
    settings: { 'jsx-a11y': { components: { Meta: 'meta' } } },
  } as never);

const invalidAriaValidityTests = allDomElements
  .filter((el) => reservedSet.has(el))
  .map((el) => ({
    code: `<${el} aria-hidden aria-role="none" {...props} />`,
    errors: [{ message: errorMessage('aria-hidden') }],
  }));

new RuleTester().run('aria-unsupported-elements', null as never, {
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
    ...invalidRoleValidityTests,
    ...invalidAriaValidityTests,
    // Case-insensitive matching.
    {
      code: '<base ROLE />',
      errors: [{ message: errorMessage('role') }],
    },
    {
      code: '<base ARIA-HIDDEN />',
      errors: [{ message: errorMessage('aria-hidden') }],
    },
    // Multiple invalid attributes on the same element.
    {
      code: '<base role aria-hidden aria-label="x" />',
      errors: [
        { message: errorMessage('role') },
        { message: errorMessage('aria-hidden') },
        { message: errorMessage('aria-label') },
      ],
    },
    // Paired opening/closing form.
    {
      code: '<style aria-hidden></style>',
      errors: [{ message: errorMessage('aria-hidden') }],
    },
    // polymorphicPropName resolves to a reserved element.
    {
      code: '<Box as="base" aria-hidden />',
      settings: { 'jsx-a11y': { polymorphicPropName: 'as' } },
      errors: [{ message: errorMessage('aria-hidden') }],
    },
    {
      code: '<Foo as="meta" role="none" />',
      settings: { 'jsx-a11y': { polymorphicPropName: 'as' } },
      errors: [{ message: errorMessage('role') }],
    },
    // Components map points at a reserved element.
    {
      code: '<Foo aria-hidden />',
      settings: { 'jsx-a11y': { components: { Foo: 'meta' } } },
      errors: [{ message: errorMessage('aria-hidden') }],
    },
    // Reserved nested in non-reserved.
    {
      code: '<div><base aria-hidden /></div>',
      errors: [{ message: errorMessage('aria-hidden') }],
    },
    // Reserved nested in reserved — both fire independently.
    {
      code: '<head role="x"><base aria-hidden /></head>',
      errors: [
        { message: errorMessage('role') },
        { message: errorMessage('aria-hidden') },
      ],
    },
    // Mixed valid + invalid attributes on a reserved element.
    {
      code: '<base href="x" target="_blank" aria-hidden />',
      errors: [{ message: errorMessage('aria-hidden') }],
    },
    // Spread interleaved with named attributes.
    {
      code: '<base {...a} role {...b} aria-hidden {...c} />',
      errors: [
        { message: errorMessage('role') },
        { message: errorMessage('aria-hidden') },
      ],
    },
    // Boolean form of `role`.
    {
      code: '<base role />',
      errors: [{ message: errorMessage('role') }],
    },
    // Inside a JSX fragment / conditional expression.
    {
      code: '<>{cond && <title aria-hidden>x</title>}</>',
      errors: [{ message: errorMessage('aria-hidden') }],
    },
    // Boolean form of an aria-* attribute.
    {
      code: '<base aria-label />',
      errors: [{ message: errorMessage('aria-label') }],
    },
    // Reserved tag with children.
    {
      code: '<base aria-hidden>x</base>',
      errors: [{ message: errorMessage('aria-hidden') }],
    },
    // Self-closing without space.
    {
      code: '<base aria-hidden/>',
      errors: [{ message: errorMessage('aria-hidden') }],
    },
    // Array.map iteration — JSX inside expression body.
    {
      code: 'function L({xs}) { return xs.map(x => <base aria-hidden key={x} />); }',
      errors: [{ message: errorMessage('aria-hidden') }],
    },
    // Class component render method.
    {
      code: 'class C { render() { return <link aria-hidden />; } }',
      errors: [{ message: errorMessage('aria-hidden') }],
    },
    // Conditional render via && inside parent JSX.
    {
      code: 'function F() { return <div>{cond && <meta aria-hidden />}</div>; }',
      errors: [{ message: errorMessage('aria-hidden') }],
    },
    // TS type assertion on attribute value.
    {
      code: '<base aria-hidden={true as boolean} />',
      errors: [{ message: errorMessage('aria-hidden') }],
    },
    // `as const` assertion.
    {
      code: '<base aria-label={"x" as const} />',
      errors: [{ message: errorMessage('aria-label') }],
    },
    // JSX inside generic function.
    {
      code: 'function f<T>(x: T) { return <base aria-hidden />; }',
      errors: [{ message: errorMessage('aria-hidden') }],
    },
    // polymorphicAllowList allows the substitution.
    {
      code: '<Box as="meta" aria-hidden />',
      settings: {
        'jsx-a11y': {
          polymorphicPropName: 'as',
          polymorphicAllowList: ['Box'],
        },
      },
      errors: [{ message: errorMessage('aria-hidden') }],
    },
  ],
});
