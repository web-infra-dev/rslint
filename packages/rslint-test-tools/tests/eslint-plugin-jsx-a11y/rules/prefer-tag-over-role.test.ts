import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

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

new RuleTester(config).run('prefer-tag-over-role', {
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
        {
          messageId: 'preferTagOverRole',
          message:
            'Use <input type="checkbox"> instead of the "checkbox" role to ensure accessibility across all devices.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<div role="button checkbox" />',
      errors: [
        {
          messageId: 'preferTagOverRole',
          message:
            'Use <input type="checkbox"> instead of the "checkbox" role to ensure accessibility across all devices.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<div role="heading" />',
      errors: [
        {
          messageId: 'preferTagOverRole',
          message:
            'Use <h1>, <h2>, <h3>, <h4>, <h5>, or <h6> instead of the "heading" role to ensure accessibility across all devices.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<div role="link" />',
      errors: [
        {
          messageId: 'preferTagOverRole',
          message:
            'Use <a href=...>, or <area href=...> instead of the "link" role to ensure accessibility across all devices.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<div role="rowgroup" />',
      errors: [
        {
          messageId: 'preferTagOverRole',
          message:
            'Use <tbody>, <tfoot>, or <thead> instead of the "rowgroup" role to ensure accessibility across all devices.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<span role="checkbox" />',
      errors: [
        {
          messageId: 'preferTagOverRole',
          message:
            'Use <input type="checkbox"> instead of the "checkbox" role to ensure accessibility across all devices.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<other role="checkbox" />',
      errors: [
        {
          messageId: 'preferTagOverRole',
          message:
            'Use <input type="checkbox"> instead of the "checkbox" role to ensure accessibility across all devices.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<div role="banner" />',
      errors: [
        {
          messageId: 'preferTagOverRole',
          message:
            'Use <header> instead of the "banner" role to ensure accessibility across all devices.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },

    // ============================================================
    // role inside `{…}` JsxExpression — literal extraction must apply.
    // ============================================================
    {
      code: '<div role={"checkbox"} />',
      errors: [
        {
          messageId: 'preferTagOverRole',
          message:
            'Use <input type="checkbox"> instead of the "checkbox" role to ensure accessibility across all devices.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<div role={`checkbox`} />',
      errors: [
        {
          messageId: 'preferTagOverRole',
          message:
            'Use <input type="checkbox"> instead of the "checkbox" role to ensure accessibility across all devices.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<div role={"check" + "box"} />',
      errors: [
        {
          messageId: 'preferTagOverRole',
          message:
            'Use <input type="checkbox"> instead of the "checkbox" role to ensure accessibility across all devices.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },

    // ============================================================
    // Case-insensitive role-ATTRIBUTE-name match.
    // ============================================================
    {
      code: '<div ROLE="checkbox" />',
      errors: [
        {
          messageId: 'preferTagOverRole',
          message:
            'Use <input type="checkbox"> instead of the "checkbox" role to ensure accessibility across all devices.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },

    // ============================================================
    // Literal-object spread.
    // ============================================================
    {
      code: '<div {...{role:"checkbox"}} />',
      errors: [
        {
          messageId: 'preferTagOverRole',
          message:
            'Use <input type="checkbox"> instead of the "checkbox" role to ensure accessibility across all devices.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },

    // ============================================================
    // Paired form `<X role="…">…</X>` — opening element is reported.
    // ============================================================
    {
      code: '<div role="checkbox"></div>',
      errors: [
        {
          messageId: 'preferTagOverRole',
          message:
            'Use <input type="checkbox"> instead of the "checkbox" role to ensure accessibility across all devices.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
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
        {
          messageId: 'preferTagOverRole',
          message:
            'Use <input type="checkbox"> instead of the "checkbox" role to ensure accessibility across all devices.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: '<NotAnImg role="img" />',
      settings: componentsSettings,
      errors: [
        {
          messageId: 'preferTagOverRole',
          message:
            'Use <img alt=...>, or <img alt=...> instead of the "img" role to ensure accessibility across all devices.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
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
          messageId: 'preferTagOverRole',
          message:
            'Use <a href=...>, or <area href=...> instead of the "link" role to ensure accessibility across all devices.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: '<svg:foo role="checkbox" />',
      errors: [
        {
          messageId: 'preferTagOverRole',
          message:
            'Use <input type="checkbox"> instead of the "checkbox" role to ensure accessibility across all devices.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },

    // ============================================================
    // Multi-token role — LAST token semantics.
    // ============================================================
    {
      code: '<div role="alert button checkbox" />',
      errors: [
        {
          messageId: 'preferTagOverRole',
          message:
            'Use <input type="checkbox"> instead of the "checkbox" role to ensure accessibility across all devices.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: '<div role=" checkbox" />',
      errors: [
        {
          messageId: 'preferTagOverRole',
          message:
            'Use <input type="checkbox"> instead of the "checkbox" role to ensure accessibility across all devices.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },

    // ============================================================
    // Element-vs-role mismatches across single-tag mappings.
    // ============================================================
    {
      code: '<div role="article" />',
      errors: [
        {
          messageId: 'preferTagOverRole',
          message:
            'Use <article> instead of the "article" role to ensure accessibility across all devices.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<div role="navigation" />',
      errors: [
        {
          messageId: 'preferTagOverRole',
          message:
            'Use <nav> instead of the "navigation" role to ensure accessibility across all devices.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<div role="separator" />',
      errors: [
        {
          messageId: 'preferTagOverRole',
          message:
            'Use <hr> instead of the "separator" role to ensure accessibility across all devices.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<div role="paragraph" />',
      errors: [
        {
          messageId: 'preferTagOverRole',
          message:
            'Use <p> instead of the "paragraph" role to ensure accessibility across all devices.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },

    // ============================================================
    // Roles where formatTag emits the value-less `...` fallback.
    // ============================================================
    {
      code: '<div role="textbox" />',
      errors: [
        {
          messageId: 'preferTagOverRole',
          message:
            'Use <input type=...>, <input list=...>, <input list=...>, <input list=...>, <input list=...>, or <textarea> instead of the "textbox" role to ensure accessibility across all devices.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<div role="searchbox" />',
      errors: [
        {
          messageId: 'preferTagOverRole',
          message:
            'Use <input list=...> instead of the "searchbox" role to ensure accessibility across all devices.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },

    // ============================================================
    // button-on-element-that's-not-button (unidirectional).
    // ============================================================
    {
      code: '<button role="link" />',
      errors: [
        {
          messageId: 'preferTagOverRole',
          message:
            'Use <a href=...>, or <area href=...> instead of the "link" role to ensure accessibility across all devices.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<a role="button" />',
      errors: [
        {
          messageId: 'preferTagOverRole',
          message:
            'Use <input type="button">, <input type="image">, <input type="reset">, <input type="submit">, or <button> instead of the "button" role to ensure accessibility across all devices.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
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
        {
          messageId: 'preferTagOverRole',
          message:
            'Use <input type="checkbox"> instead of the "checkbox" role to ensure accessibility across all devices.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: '<div role="button&#32;checkbox" />',
      errors: [
        {
          messageId: 'preferTagOverRole',
          message:
            'Use <input type="checkbox"> instead of the "checkbox" role to ensure accessibility across all devices.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 35,
        },
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
        {
          messageId: 'preferTagOverRole',
          message:
            'Use <input type="checkbox"> instead of the "checkbox" role to ensure accessibility across all devices.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<div role={"" || checkbox} />',
      errors: [
        {
          messageId: 'preferTagOverRole',
          message:
            'Use <input type="checkbox"> instead of the "checkbox" role to ensure accessibility across all devices.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },

    // ============================================================
    // Duplicate-role attribute — FIRST occurrence wins (matches
    // jsx-ast-utils' getProp).
    // ============================================================
    {
      code: '<div role="checkbox" role="banner" />',
      errors: [
        {
          messageId: 'preferTagOverRole',
          message:
            'Use <input type="checkbox"> instead of the "checkbox" role to ensure accessibility across all devices.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },

    // ============================================================
    // Spread + explicit role ordering.
    // ============================================================
    {
      code: '<div {...{role:"checkbox"}} role="link" />',
      errors: [
        {
          messageId: 'preferTagOverRole',
          message:
            'Use <input type="checkbox"> instead of the "checkbox" role to ensure accessibility across all devices.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 43,
        },
      ],
    },
    {
      code: '<div {...props} role="checkbox" />',
      errors: [
        {
          messageId: 'preferTagOverRole',
          message:
            'Use <input type="checkbox"> instead of the "checkbox" role to ensure accessibility across all devices.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },

    // ============================================================
    // Real-world component patterns.
    // ============================================================
    {
      code: 'function Card() { return <div role="checkbox" />; }',
      errors: [
        {
          messageId: 'preferTagOverRole',
          message:
            'Use <input type="checkbox"> instead of the "checkbox" role to ensure accessibility across all devices.',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 49,
        },
      ],
    },
    {
      code: 'const Card = () => <div role="heading" />;',
      errors: [
        {
          messageId: 'preferTagOverRole',
          message:
            'Use <h1>, <h2>, <h3>, <h4>, <h5>, or <h6> instead of the "heading" role to ensure accessibility across all devices.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      code: 'const Card = React.memo(() => <span role="checkbox" />);',
      errors: [
        {
          messageId: 'preferTagOverRole',
          message:
            'Use <input type="checkbox"> instead of the "checkbox" role to ensure accessibility across all devices.',
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 55,
        },
      ],
    },

    // ============================================================
    // Multi-role last-token semantics on a non-checkbox last token.
    // ============================================================
    {
      code: '<div role="presentation banner" />',
      errors: [
        {
          messageId: 'preferTagOverRole',
          message:
            'Use <header> instead of the "banner" role to ensure accessibility across all devices.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
  ],
});
