import { RuleTester } from '../rule-tester';

const errorMessage = (role: string, tag: string): string =>
  `Use ${tag} instead of the "${role}" role to ensure accessibility across all devices.`;

const polymorphicSettings = {
  'jsx-a11y': {
    polymorphicPropName: 'as',
  },
};

const componentsSettings = {
  'jsx-a11y': {
    components: {
      MyHeader: 'header',
      NotAnImg: 'div',
    },
  },
};

new RuleTester().run('prefer-tag-over-role', null as never, {
  valid: [
    // ============================================================
    // Upstream valid set (mirrors __tests__/.../prefer-tag-over-role-test.js).
    // ============================================================
    { code: '<div />;' },
    { code: '<div role="unknown" />;' },
    { code: '<div role="also unknown" />;' },
    { code: '<other />' },
    { code: '<img role="img" />' },
    { code: '<input role="checkbox" />' },

    // ============================================================
    // Non-literal role values short-circuit before any roleElements lookup.
    // ============================================================
    { code: '<div role={someRole} />' },
    { code: '<div role={r || "checkbox"} />' },
    { code: '<div role={getRole()} />' },
    { code: '<div role={config.role} />' },
    { code: '<div role />' },
    { code: '<div role="" />' },
    { code: '<div role={""} />' },
    { code: '<div role={true} />' },
    { code: '<div role={null} />' },
    { code: '<div role={42} />' },

    // ============================================================
    // Element already IS one of the role's semantic tags → upstream's
    // `.some()` short-circuits → skip.
    // ============================================================
    { code: '<h1 role="heading" />' },
    { code: '<h6 role="heading" />' },
    { code: '<tbody role="rowgroup" />' },
    { code: '<thead role="rowgroup" />' },
    { code: '<tfoot role="rowgroup" />' },
    { code: '<a role="link" />' }, // href not required for the name match
    { code: '<area role="link" />' },
    { code: '<header role="banner" />' },
    { code: '<input role="textbox" />' },
    { code: '<textarea role="textbox" />' },
    { code: '<select role="combobox" />' },
    { code: '<input role="combobox" />' },
    { code: '<select role="listbox" />' },
    { code: '<datalist role="listbox" />' },

    // ============================================================
    // Case-sensitive roleElements lookup — `CHECKBOX` is not a key.
    // ============================================================
    { code: '<div role="CHECKBOX" />' },
    { code: '<div role="Heading" />' },

    // ============================================================
    // polymorphicPropName: <Box as="input" role="checkbox" /> resolves
    // to elementType "input" → input is in checkbox.tagNames → valid.
    // ============================================================
    {
      code: '<Box as="input" role="checkbox" />',
      settings: polymorphicSettings,
    },

    // ============================================================
    // components map: <MyHeader role="banner" /> resolves to header.
    // ============================================================
    {
      code: '<MyHeader role="banner" />',
      settings: componentsSettings,
    },

    // ============================================================
    // Unknown / abstract roles: not in roleElementMap → skip.
    // ============================================================
    { code: '<div role="not-a-real-role" />' },
    { code: '<div role="checkbox unknown" />' }, // last token unknown
    { code: '<div role="command" />' }, // abstract
    { code: '<div role="widget" />' }, // abstract
    { code: '<div role="composite" />' }, // abstract

    // ============================================================
    // Trailing / whitespace-only role string.
    // ============================================================
    { code: '<div role="checkbox " />' }, // trailing space → "" → skip
    { code: '<div role=" " />' },
  ],
  invalid: [
    // ============================================================
    // Upstream invalid set.
    // ============================================================
    {
      code: '<div role="checkbox" />',
      errors: [
        { message: errorMessage('checkbox', '<input type="checkbox">') },
      ],
    },
    {
      code: '<div role="button checkbox" />',
      errors: [
        { message: errorMessage('checkbox', '<input type="checkbox">') },
      ],
    },
    {
      code: '<div role="heading" />',
      errors: [
        {
          message: errorMessage(
            'heading',
            '<h1>, <h2>, <h3>, <h4>, <h5>, or <h6>',
          ),
        },
      ],
    },
    {
      code: '<div role="link" />',
      errors: [
        {
          message: errorMessage('link', '<a href=...>, or <area href=...>'),
        },
      ],
    },
    {
      code: '<div role="rowgroup" />',
      errors: [
        {
          message: errorMessage('rowgroup', '<tbody>, <tfoot>, or <thead>'),
        },
      ],
    },
    {
      code: '<span role="checkbox" />',
      errors: [
        { message: errorMessage('checkbox', '<input type="checkbox">') },
      ],
    },
    {
      code: '<other role="checkbox" />',
      errors: [
        { message: errorMessage('checkbox', '<input type="checkbox">') },
      ],
    },
    {
      code: '<div role="banner" />',
      errors: [{ message: errorMessage('banner', '<header>') }],
    },

    // ============================================================
    // role inside `{…}` JsxExpression — literal extraction must apply.
    // ============================================================
    {
      code: '<div role={"checkbox"} />',
      errors: [
        { message: errorMessage('checkbox', '<input type="checkbox">') },
      ],
    },
    {
      code: '<div role={`checkbox`} />',
      errors: [
        { message: errorMessage('checkbox', '<input type="checkbox">') },
      ],
    },
    {
      code: '<div role={"check" + "box"} />',
      errors: [
        { message: errorMessage('checkbox', '<input type="checkbox">') },
      ],
    },

    // ============================================================
    // Case-insensitive role-ATTRIBUTE-name match.
    // ============================================================
    {
      code: '<div ROLE="checkbox" />',
      errors: [
        { message: errorMessage('checkbox', '<input type="checkbox">') },
      ],
    },

    // ============================================================
    // Literal-object spread.
    // ============================================================
    {
      code: '<div {...{role:"checkbox"}} />',
      errors: [
        { message: errorMessage('checkbox', '<input type="checkbox">') },
      ],
    },

    // ============================================================
    // Paired form `<X role="…">…</X>` — opening element is reported.
    // ============================================================
    {
      code: '<div role="checkbox"></div>',
      errors: [
        { message: errorMessage('checkbox', '<input type="checkbox">') },
      ],
    },

    // ============================================================
    // polymorphicPropName + components map: element-name resolution
    // happens BEFORE tagNames check.
    // ============================================================
    {
      code: '<Box as="div" role="checkbox" />',
      settings: polymorphicSettings,
      errors: [
        { message: errorMessage('checkbox', '<input type="checkbox">') },
      ],
    },
    {
      code: '<NotAnImg role="img" />',
      settings: componentsSettings,
      errors: [
        {
          message: errorMessage('img', '<img alt=...>, or <img alt=...>'),
        },
      ],
    },

    // ============================================================
    // PascalCase custom element / namespaced JSX — neither in tagNames.
    // ============================================================
    {
      code: '<MyComponent role="link" />',
      errors: [
        {
          message: errorMessage('link', '<a href=...>, or <area href=...>'),
        },
      ],
    },
    {
      code: '<svg:foo role="checkbox" />',
      errors: [
        { message: errorMessage('checkbox', '<input type="checkbox">') },
      ],
    },

    // ============================================================
    // Multi-token role — LAST token semantics.
    // ============================================================
    {
      code: '<div role="alert button checkbox" />',
      errors: [
        { message: errorMessage('checkbox', '<input type="checkbox">') },
      ],
    },
    {
      code: '<div role=" checkbox" />',
      errors: [
        { message: errorMessage('checkbox', '<input type="checkbox">') },
      ],
    },

    // ============================================================
    // Element-vs-role mismatches across single-tag mappings.
    // ============================================================
    {
      code: '<div role="article" />',
      errors: [{ message: errorMessage('article', '<article>') }],
    },
    {
      code: '<div role="navigation" />',
      errors: [{ message: errorMessage('navigation', '<nav>') }],
    },
    {
      code: '<div role="separator" />',
      errors: [{ message: errorMessage('separator', '<hr>') }],
    },
    {
      code: '<div role="paragraph" />',
      errors: [{ message: errorMessage('paragraph', '<p>') }],
    },

    // ============================================================
    // Roles where formatTag emits the value-less `...` fallback.
    // ============================================================
    {
      code: '<div role="textbox" />',
      errors: [
        {
          message: errorMessage(
            'textbox',
            '<input type=...>, <input list=...>, <input list=...>, <input list=...>, <input list=...>, or <textarea>',
          ),
        },
      ],
    },
    {
      code: '<div role="searchbox" />',
      errors: [{ message: errorMessage('searchbox', '<input list=...>') }],
    },

    // ============================================================
    // button-on-element-that's-not-button (unidirectional).
    // ============================================================
    {
      code: '<button role="link" />',
      errors: [
        {
          message: errorMessage('link', '<a href=...>, or <area href=...>'),
        },
      ],
    },
    {
      code: '<a role="button" />',
      errors: [
        {
          message: errorMessage(
            'button',
            '<input type="button">, <input type="image">, <input type="reset">, <input type="submit">, or <button>',
          ),
        },
      ],
    },

    // ============================================================
    // HTML entity decoding on direct attributes — extractRoleString
    // patches the decode so &#99; → "c", entire value "checkbox".
    // ============================================================
    {
      code: '<div role="&#99;heckbox" />',
      errors: [
        { message: errorMessage('checkbox', '<input type="checkbox">') },
      ],
    },
    {
      code: '<div role="button&#32;checkbox" />',
      errors: [
        { message: errorMessage('checkbox', '<input type="checkbox">') },
      ],
    },

    // ============================================================
    // Identifier whose NAME happens to be a role string — upstream's
    // getPropValue returns the bare name. Real codebase pattern:
    //   const checkbox = 'checkbox'; <div role={checkbox} />
    // ============================================================
    {
      code: '<div role={checkbox} />',
      errors: [
        { message: errorMessage('checkbox', '<input type="checkbox">') },
      ],
    },
    {
      code: '<div role={"" || checkbox} />',
      errors: [
        { message: errorMessage('checkbox', '<input type="checkbox">') },
      ],
    },

    // ============================================================
    // Duplicate-role attribute — FIRST occurrence wins (matches
    // jsx-ast-utils' getProp).
    // ============================================================
    {
      code: '<div role="checkbox" role="banner" />',
      errors: [
        { message: errorMessage('checkbox', '<input type="checkbox">') },
      ],
    },

    // ============================================================
    // Spread + explicit role ordering.
    // ============================================================
    {
      code: '<div {...{role:"checkbox"}} role="link" />',
      errors: [
        { message: errorMessage('checkbox', '<input type="checkbox">') },
      ],
    },
    {
      code: '<div {...props} role="checkbox" />',
      errors: [
        { message: errorMessage('checkbox', '<input type="checkbox">') },
      ],
    },

    // ============================================================
    // Real-world component patterns.
    // ============================================================
    {
      code: 'function Card() { return <div role="checkbox" />; }',
      errors: [
        { message: errorMessage('checkbox', '<input type="checkbox">') },
      ],
    },
    {
      code: 'const Card = () => <div role="heading" />;',
      errors: [
        {
          message: errorMessage(
            'heading',
            '<h1>, <h2>, <h3>, <h4>, <h5>, or <h6>',
          ),
        },
      ],
    },
    {
      code: 'const Card = React.memo(() => <span role="checkbox" />);',
      errors: [
        { message: errorMessage('checkbox', '<input type="checkbox">') },
      ],
    },

    // ============================================================
    // Multi-role last-token semantics on a non-checkbox last token.
    // ============================================================
    {
      code: '<div role="presentation banner" />',
      errors: [{ message: errorMessage('banner', '<header>') }],
    },
  ],
});
