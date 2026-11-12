import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';
import { singleLineError } from '../matrix';

const componentsSettings = {
  'jsx-a11y': {
    components: { Link: 'a' },
  },
};

const componentsAndPolymorphicSettings = {
  'jsx-a11y': {
    polymorphicPropName: 'as',
    components: { Box: 'div', Anchor: 'a' },
  },
};

const polymorphicAllowListSettings = {
  'jsx-a11y': {
    polymorphicPropName: 'as',
    polymorphicAllowList: ['Box'],
  },
};

new RuleTester(config).run('role-supports-aria-props', {
  valid: [
    // ---- Generic / non-DOM cases. ----
    { code: '<Foo bar />' },
    { code: '<div />' },
    { code: '<div id="main" />' },
    { code: '<div role />' },
    { code: '<div role="presentation" {...props} />' },
    { code: '<Foo.Bar baz={true} />' },
    { code: '<Link href="#" aria-checked />' },

    // ---- A: implicit role link. ----
    { code: '<a href="#" aria-expanded />' },
    { code: '<a href="#" aria-atomic />' },
    { code: '<a href="#" aria-busy />' },
    { code: '<a href="#" aria-controls />' },
    { code: '<a href="#" aria-current />' },
    { code: '<a href="#" aria-describedby />' },
    { code: '<a href="#" aria-disabled />' },
    { code: '<a href="#" aria-dropeffect />' },
    { code: '<a href="#" aria-flowto />' },
    { code: '<a href="#" aria-haspopup />' },
    { code: '<a href="#" aria-grabbed />' },
    { code: '<a href="#" aria-hidden />' },
    { code: '<a href="#" aria-label />' },
    { code: '<a href="#" aria-labelledby />' },
    { code: '<a href="#" aria-live />' },
    { code: '<a href="#" aria-owns />' },
    { code: '<a href="#" aria-relevant />' },
    { code: '<a aria-checked />' },

    // ---- AREA: implicit role link (with href). ----
    { code: '<area href="#" aria-expanded />' },
    { code: '<area href="#" aria-atomic />' },
    { code: '<area href="#" aria-busy />' },
    { code: '<area href="#" aria-controls />' },
    { code: '<area href="#" aria-describedby />' },
    { code: '<area href="#" aria-disabled />' },
    { code: '<area href="#" aria-dropeffect />' },
    { code: '<area href="#" aria-flowto />' },
    { code: '<area href="#" aria-grabbed />' },
    { code: '<area href="#" aria-haspopup />' },
    { code: '<area href="#" aria-hidden />' },
    { code: '<area href="#" aria-label />' },
    { code: '<area href="#" aria-labelledby />' },
    { code: '<area href="#" aria-live />' },
    { code: '<area href="#" aria-owns />' },
    { code: '<area href="#" aria-relevant />' },
    { code: '<area aria-checked />' },

    // ---- LINK: implicit role link (with href). ----
    { code: '<link href="#" aria-expanded />' },
    { code: '<link href="#" aria-atomic />' },
    { code: '<link href="#" aria-busy />' },
    { code: '<link href="#" aria-controls />' },
    { code: '<link href="#" aria-describedby />' },
    { code: '<link href="#" aria-disabled />' },
    { code: '<link href="#" aria-dropeffect />' },
    { code: '<link href="#" aria-flowto />' },
    { code: '<link href="#" aria-grabbed />' },
    { code: '<link href="#" aria-hidden />' },
    { code: '<link href="#" aria-haspopup />' },
    { code: '<link href="#" aria-label />' },
    { code: '<link href="#" aria-labelledby />' },
    { code: '<link href="#" aria-live />' },
    { code: '<link href="#" aria-owns />' },
    { code: '<link href="#" aria-relevant />' },
    { code: '<link aria-checked />' },

    // ---- IMG. ----
    { code: '<img alt="" aria-checked />' },
    { code: '<img alt="foobar" aria-busy />' },

    // ---- MENU type=toolbar → toolbar. ----
    { code: '<menu type="toolbar" aria-activedescendant />' },
    { code: '<menu type="toolbar" aria-atomic />' },
    { code: '<menu type="toolbar" aria-busy />' },
    { code: '<menu type="toolbar" aria-controls />' },
    { code: '<menu type="toolbar" aria-describedby />' },
    { code: '<menu type="toolbar" aria-disabled />' },
    { code: '<menu type="toolbar" aria-dropeffect />' },
    { code: '<menu type="toolbar" aria-flowto />' },
    { code: '<menu type="toolbar" aria-grabbed />' },
    { code: '<menu type="toolbar" aria-hidden />' },
    { code: '<menu type="toolbar" aria-label />' },
    { code: '<menu type="toolbar" aria-labelledby />' },
    { code: '<menu type="toolbar" aria-live />' },
    { code: '<menu type="toolbar" aria-owns />' },
    { code: '<menu type="toolbar" aria-relevant />' },
    { code: '<menu aria-checked />' },

    // ---- MENUITEM type=command → menuitem. ----
    { code: '<menuitem type="command" aria-atomic />' },
    { code: '<menuitem type="command" aria-busy />' },
    { code: '<menuitem type="command" aria-controls />' },
    { code: '<menuitem type="command" aria-describedby />' },
    { code: '<menuitem type="command" aria-disabled />' },
    { code: '<menuitem type="command" aria-dropeffect />' },
    { code: '<menuitem type="command" aria-flowto />' },
    { code: '<menuitem type="command" aria-grabbed />' },
    { code: '<menuitem type="command" aria-haspopup />' },
    { code: '<menuitem type="command" aria-hidden />' },
    { code: '<menuitem type="command" aria-label />' },
    { code: '<menuitem type="command" aria-labelledby />' },
    { code: '<menuitem type="command" aria-live />' },
    { code: '<menuitem type="command" aria-owns />' },
    { code: '<menuitem type="command" aria-relevant />' },

    // ---- MENUITEM type=checkbox → menuitemcheckbox. ----
    { code: '<menuitem type="checkbox" aria-checked />' },
    { code: '<menuitem type="checkbox" aria-atomic />' },
    { code: '<menuitem type="checkbox" aria-busy />' },
    { code: '<menuitem type="checkbox" aria-controls />' },
    { code: '<menuitem type="checkbox" aria-describedby />' },
    { code: '<menuitem type="checkbox" aria-disabled />' },
    { code: '<menuitem type="checkbox" aria-dropeffect />' },
    { code: '<menuitem type="checkbox" aria-flowto />' },
    { code: '<menuitem type="checkbox" aria-grabbed />' },
    { code: '<menuitem type="checkbox" aria-haspopup />' },
    { code: '<menuitem type="checkbox" aria-hidden />' },
    { code: '<menuitem type="checkbox" aria-invalid />' },
    { code: '<menuitem type="checkbox" aria-label />' },
    { code: '<menuitem type="checkbox" aria-labelledby />' },
    { code: '<menuitem type="checkbox" aria-live />' },
    { code: '<menuitem type="checkbox" aria-owns />' },
    { code: '<menuitem type="checkbox" aria-relevant />' },

    // ---- MENUITEM type=radio → menuitemradio. ----
    { code: '<menuitem type="radio" aria-checked />' },
    { code: '<menuitem type="radio" aria-atomic />' },
    { code: '<menuitem type="radio" aria-busy />' },
    { code: '<menuitem type="radio" aria-controls />' },
    { code: '<menuitem type="radio" aria-describedby />' },
    { code: '<menuitem type="radio" aria-disabled />' },
    { code: '<menuitem type="radio" aria-dropeffect />' },
    { code: '<menuitem type="radio" aria-flowto />' },
    { code: '<menuitem type="radio" aria-grabbed />' },
    { code: '<menuitem type="radio" aria-haspopup />' },
    { code: '<menuitem type="radio" aria-hidden />' },
    { code: '<menuitem type="radio" aria-invalid />' },
    { code: '<menuitem type="radio" aria-label />' },
    { code: '<menuitem type="radio" aria-labelledby />' },
    { code: '<menuitem type="radio" aria-live />' },
    { code: '<menuitem type="radio" aria-owns />' },
    { code: '<menuitem type="radio" aria-relevant />' },
    { code: '<menuitem type="radio" aria-posinset />' },
    { code: '<menuitem type="radio" aria-setsize />' },
    { code: '<menuitem aria-checked />' },
    { code: '<menuitem type="foo" aria-checked />' },

    // ---- INPUT type=button/image/reset/submit → button. ----
    { code: '<input type="button" aria-expanded />' },
    { code: '<input type="button" aria-pressed />' },
    { code: '<input type="image" aria-pressed />' },
    { code: '<input type="reset" aria-pressed />' },
    { code: '<input type="submit" aria-pressed />' },

    // ---- INPUT type=checkbox → checkbox. ----
    { code: '<input type="checkbox" aria-checked />' },
    { code: '<input type="checkbox" aria-disabled />' },

    // ---- INPUT type=radio → radio. ----
    { code: '<input type="radio" aria-checked />' },
    { code: '<input type="radio" aria-posinset />' },
    { code: '<input type="radio" aria-setsize />' },

    // ---- INPUT type=range → slider. ----
    { code: '<input type="range" aria-valuemax />' },
    { code: '<input type="range" aria-valuemin />' },
    { code: '<input type="range" aria-valuenow />' },
    { code: '<input type="range" aria-orientation />' },
    { code: '<input type="range" aria-valuetext />' },

    // ---- INPUT default → textbox. ----
    { code: '<input type="email" aria-disabled />' },
    { code: '<input type="password" aria-disabled />' },
    { code: '<input type="search" aria-disabled />' },
    { code: '<input type="tel" aria-disabled />' },
    { code: '<input type="url" aria-disabled />' },
    { code: '<input aria-disabled />' },

    // ---- Allow null/undefined values regardless of role. ----
    { code: '<h2 role="presentation" aria-level={null} />' },
    { code: '<h2 role="presentation" aria-level={undefined} />' },

    // ---- Other implicit roles. ----
    { code: '<button aria-pressed />' },
    { code: '<form aria-hidden />' },
    { code: '<h1 aria-hidden />' },
    { code: '<h2 aria-hidden />' },
    { code: '<h3 aria-hidden />' },
    { code: '<h4 aria-hidden />' },
    { code: '<h5 aria-hidden />' },
    { code: '<h6 aria-hidden />' },
    { code: '<hr aria-hidden />' },
    { code: '<li aria-current />' },
    { code: '<meter aria-atomic />' },
    { code: '<option aria-atomic />' },
    { code: '<progress aria-atomic />' },
    { code: '<textarea aria-hidden />' },
    { code: '<select aria-expanded />' },
    { code: '<datalist aria-expanded />' },
    { code: '<div role="heading" aria-level />' },
    { code: '<div role="heading" aria-level="1" />' },

    // ---- Identifier / non-literal role values short-circuit. ----
    { code: '<div role={x} aria-checked />' },
    { code: '<div role={fn()} aria-checked />' },
    { code: '<div role={a || b} aria-checked />' },

    // ---- Case-sensitive membership: `BUTTON` !== `button`. ----
    { code: '<div role="BUTTON" aria-checked />' },

    // ---- Unknown role names skip. ----
    { code: '<div role="not-a-real-role" aria-checked />' },

    // ---- Spread is opaque. ----
    { code: '<a href="#" {...{"aria-checked": true}} />' },

    // ---- Real-world component patterns. ----
    { code: '<Foo.Bar role="link" aria-expanded />' },
    { code: '<svg:circle role="img" aria-label="x" />' },
    {
      code: `
const ForwardedAnchor = React.forwardRef((props, ref) => (
  <a ref={ref} href={props.to} aria-expanded={props.expanded} />
));
      `,
    },
    {
      code: `
<>
  <a href="/a" aria-expanded />
  <a href="/b" aria-busy />
</>
      `,
    },
    { code: 'cond ? <a href="/" aria-expanded /> : <button aria-pressed />' },

    // ---- TypeScript wrappers. ----
    { code: '<div role={"button" as string} aria-checked />' },
    { code: '<div role={someRole!} aria-checked />' },
    { code: '<div role={("button")} aria-pressed />' },
    { code: '<a href="/" aria-checked={null as any} />' },

    // ---- ARIA prop value coverage — nullish forms always skip. ----
    { code: '<a href="/" aria-checked={null} />' },
    { code: '<a href="/" aria-checked={undefined} />' },
    { code: '<a href="/" aria-checked={(null)} />' },

    // ---- Spread + literal interactions. ----
    { code: '<div {...{role: "button"}} aria-pressed />' },
    { code: '<div {...rest} aria-checked />' },
    { code: '<div {...rest} role="button" aria-pressed />' },

    // ---- Components map + polymorphic combinations. ----
    {
      code: '<Box as="a" href="/" aria-expanded />',
      settings: componentsAndPolymorphicSettings,
    },
    {
      code: '<Box aria-checked />',
      settings: componentsAndPolymorphicSettings,
    },
    {
      code: '<Container as="a" href="/" aria-checked />',
      settings: polymorphicAllowListSettings,
    },

    // ---- Multi-line / formatted JSX. ----
    {
      code: `
<a
  href="/"
  aria-expanded
  aria-busy
  aria-controls="x"
/>
      `,
    },

    // ---- Nested elements — each is independently classified. ----
    {
      code: `
<div aria-controls="panel">
  <a href="/" aria-expanded />
</div>
      `,
    },

    // ---- Implicit-role table boundary — href present in any form. ----
    { code: '<a href aria-expanded />' },
    { code: '<a href={someVar} aria-expanded />' },
    { code: '<a href={null} aria-expanded />' },

    // ---- select size > 1 / multiple → listbox. ----
    { code: '<select size={2} aria-multiselectable />' },
    { code: '<select multiple aria-multiselectable />' },
  ],
  invalid: [
    // ---- Implicit basic checks. ----
    {
      code: '<a href="#" aria-checked />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-checked is not supported by the role link. This role is implicit on the element a.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: '<area href="#" aria-checked />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-checked is not supported by the role link. This role is implicit on the element area.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<link href="#" aria-checked />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-checked is not supported by the role link. This role is implicit on the element link.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<img alt="foobar" aria-checked />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-checked is not supported by the role img. This role is implicit on the element img.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: '<menu type="toolbar" aria-checked />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-checked is not supported by the role toolbar. This role is implicit on the element menu.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: '<aside aria-checked />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-checked is not supported by the role complementary. This role is implicit on the element aside.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<ul aria-expanded />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-expanded is not supported by the role list. This role is implicit on the element ul.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<details aria-expanded />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-expanded is not supported by the role group. This role is implicit on the element details.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<dialog aria-expanded />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-expanded is not supported by the role dialog. This role is implicit on the element dialog.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<aside aria-expanded />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-expanded is not supported by the role complementary. This role is implicit on the element aside.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<article aria-expanded />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-expanded is not supported by the role article. This role is implicit on the element article.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<body aria-expanded />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-expanded is not supported by the role document. This role is implicit on the element body.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<li aria-expanded />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-expanded is not supported by the role listitem. This role is implicit on the element li.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<nav aria-expanded />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-expanded is not supported by the role navigation. This role is implicit on the element nav.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<ol aria-expanded />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-expanded is not supported by the role list. This role is implicit on the element ol.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<output aria-expanded />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-expanded is not supported by the role status. This role is implicit on the element output.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<section aria-expanded />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-expanded is not supported by the role region. This role is implicit on the element section.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<tbody aria-expanded />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-expanded is not supported by the role rowgroup. This role is implicit on the element tbody.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<tfoot aria-expanded />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-expanded is not supported by the role rowgroup. This role is implicit on the element tfoot.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<thead aria-expanded />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-expanded is not supported by the role rowgroup. This role is implicit on the element thead.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<input type="radio" aria-invalid />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-invalid is not supported by the role radio. This role is implicit on the element input.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      code: '<input type="radio" aria-selected />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-selected is not supported by the role radio. This role is implicit on the element input.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: '<input type="radio" aria-haspopup />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-haspopup is not supported by the role radio. This role is implicit on the element input.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: '<input type="checkbox" aria-haspopup />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-haspopup is not supported by the role checkbox. This role is implicit on the element input.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      code: '<input type="reset" aria-invalid />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-invalid is not supported by the role button. This role is implicit on the element input.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      code: '<input type="submit" aria-invalid />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-invalid is not supported by the role button. This role is implicit on the element input.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: '<input type="image" aria-invalid />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-invalid is not supported by the role button. This role is implicit on the element input.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      code: '<input type="button" aria-invalid />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-invalid is not supported by the role button. This role is implicit on the element input.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: '<menuitem type="command" aria-invalid />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-invalid is not supported by the role menuitem. This role is implicit on the element menuitem.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      code: '<menuitem type="radio" aria-selected />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-selected is not supported by the role menuitemradio. This role is implicit on the element menuitem.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      code: '<menu type="toolbar" aria-haspopup />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-haspopup is not supported by the role toolbar. This role is implicit on the element menu.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      code: '<menu type="toolbar" aria-invalid />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-invalid is not supported by the role toolbar. This role is implicit on the element menu.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: '<menu type="toolbar" aria-expanded />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-expanded is not supported by the role toolbar. This role is implicit on the element menu.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      code: '<link href="#" aria-invalid />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-invalid is not supported by the role link. This role is implicit on the element link.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<area href="#" aria-invalid />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-invalid is not supported by the role link. This role is implicit on the element area.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<a href="#" aria-invalid />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-invalid is not supported by the role link. This role is implicit on the element a.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: '<Link href="#" aria-checked />',
      settings: componentsSettings,
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-checked is not supported by the role link. This role is implicit on the element a.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },

    // ---- Explicit role; isImplicit=false branch. ----
    {
      code: '<div role="link" aria-checked />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-checked is not supported by the role link.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    // ---- Explicit role on custom element (no IsDOMElement gate). ----
    {
      code: '<Foo role="link" aria-checked />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-checked is not supported by the role link.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },

    // ---- Multiple invalid props on the same element — one report per. ----
    {
      code: '<a href="#" aria-checked aria-pressed />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-checked is not supported by the role link. This role is implicit on the element a.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
        },
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-pressed is not supported by the role link. This role is implicit on the element a.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },

    // ---- Non-nullish ARIA prop values still trigger reports. ----
    {
      code: '<a href="/" aria-checked={false} />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-checked is not supported by the role link. This role is implicit on the element a.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      code: '<a href="/" aria-checked="" />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-checked is not supported by the role link. This role is implicit on the element a.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<a href="/" aria-checked={0} />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-checked is not supported by the role link. This role is implicit on the element a.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: '<a href="/" aria-checked={someVar} />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-checked is not supported by the role link. This role is implicit on the element a.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      code: '<a href="/" aria-checked={fn()} />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-checked is not supported by the role link. This role is implicit on the element a.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },

    // ---- Non-null assertion on a nullish value — synthesizes a non-null
    //      string ("null!") so the attr proceeds to the membership check.
    {
      code: '<a href="/" aria-checked={null!} />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-checked is not supported by the role link. This role is implicit on the element a.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },

    // ---- Spread + literal containing role + outside aria-* report. ----
    {
      code: '<div {...{role: "link"}} aria-checked />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-checked is not supported by the role link.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },

    // ---- Components map + polymorphic invalid cases. ----
    {
      code: '<Box as="a" href="/" aria-checked />',
      settings: componentsAndPolymorphicSettings,
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-checked is not supported by the role link. This role is implicit on the element a.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: '<Anchor href="/" aria-checked />',
      settings: componentsAndPolymorphicSettings,
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-checked is not supported by the role link. This role is implicit on the element a.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },

    // ---- Nested element where only the inner reports. ----
    {
      code: `
<div>
  <a href="/" aria-checked />
</div>
      `,
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-checked is not supported by the role link. This role is implicit on the element a.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 30,
        },
      ],
    },

    // ---- select multiple / size>1 → listbox; aria-checked invalid. ----
    {
      code: '<select multiple aria-checked />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-checked is not supported by the role listbox. This role is implicit on the element select.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: '<select size={5} aria-checked />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-checked is not supported by the role listbox. This role is implicit on the element select.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },

    // ---- presentation / none — restricted / empty supported sets. ----
    {
      code: '<div role="presentation" aria-checked />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-checked is not supported by the role presentation.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      code: '<div role="none" aria-busy />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message: 'The attribute aria-busy is not supported by the role none.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },

    // ---- bare <button> / <img> / <details> / <datalist> implicit roles. ----
    {
      code: '<button aria-checked />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-checked is not supported by the role button. This role is implicit on the element button.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<img aria-checked />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-checked is not supported by the role img. This role is implicit on the element img.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<details aria-checked />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-checked is not supported by the role group. This role is implicit on the element details.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<datalist aria-checked />',
      errors: [
        {
          messageId: 'role-supports-aria-props',
          message:
            'The attribute aria-checked is not supported by the role listbox. This role is implicit on the element datalist.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
  ],
});

// Additional upstream v6.10.2 inputs. Matrix rows freeze the expected
// classification and configuration; ranges follow the authored source templates.
// No rule implementation or lint output is consulted when generating cases.

new RuleTester(config).run('role-supports-aria-props', {
  valid: [
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 132',
      code: '<input type="button" aria-atomic />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 133',
      code: '<input type="button" aria-busy />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 134',
      code: '<input type="button" aria-controls />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 135',
      code: '<input type="button" aria-describedby />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 136',
      code: '<input type="button" aria-disabled />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 137',
      code: '<input type="button" aria-dropeffect />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 138',
      code: '<input type="button" aria-flowto />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 139',
      code: '<input type="button" aria-grabbed />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 140',
      code: '<input type="button" aria-haspopup />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 141',
      code: '<input type="button" aria-hidden />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 142',
      code: '<input type="button" aria-label />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 143',
      code: '<input type="button" aria-labelledby />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 144',
      code: '<input type="button" aria-live />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 145',
      code: '<input type="button" aria-owns />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 146',
      code: '<input type="button" aria-relevant />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 147',
      code: '<input type="image" aria-expanded />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 149',
      code: '<input type="image" aria-atomic />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 150',
      code: '<input type="image" aria-busy />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 151',
      code: '<input type="image" aria-controls />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 152',
      code: '<input type="image" aria-describedby />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 153',
      code: '<input type="image" aria-disabled />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 154',
      code: '<input type="image" aria-dropeffect />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 155',
      code: '<input type="image" aria-flowto />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 156',
      code: '<input type="image" aria-grabbed />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 157',
      code: '<input type="image" aria-haspopup />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 158',
      code: '<input type="image" aria-hidden />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 159',
      code: '<input type="image" aria-label />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 160',
      code: '<input type="image" aria-labelledby />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 161',
      code: '<input type="image" aria-live />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 162',
      code: '<input type="image" aria-owns />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 163',
      code: '<input type="image" aria-relevant />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 164',
      code: '<input type="reset" aria-expanded />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 166',
      code: '<input type="reset" aria-atomic />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 167',
      code: '<input type="reset" aria-busy />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 168',
      code: '<input type="reset" aria-controls />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 169',
      code: '<input type="reset" aria-describedby />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 170',
      code: '<input type="reset" aria-disabled />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 171',
      code: '<input type="reset" aria-dropeffect />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 172',
      code: '<input type="reset" aria-flowto />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 173',
      code: '<input type="reset" aria-grabbed />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 174',
      code: '<input type="reset" aria-haspopup />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 175',
      code: '<input type="reset" aria-hidden />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 176',
      code: '<input type="reset" aria-label />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 177',
      code: '<input type="reset" aria-labelledby />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 178',
      code: '<input type="reset" aria-live />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 179',
      code: '<input type="reset" aria-owns />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 180',
      code: '<input type="reset" aria-relevant />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 181',
      code: '<input type="submit" aria-expanded />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 183',
      code: '<input type="submit" aria-atomic />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 184',
      code: '<input type="submit" aria-busy />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 185',
      code: '<input type="submit" aria-controls />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 186',
      code: '<input type="submit" aria-describedby />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 187',
      code: '<input type="submit" aria-disabled />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 188',
      code: '<input type="submit" aria-dropeffect />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 189',
      code: '<input type="submit" aria-flowto />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 190',
      code: '<input type="submit" aria-grabbed />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 191',
      code: '<input type="submit" aria-haspopup />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 192',
      code: '<input type="submit" aria-hidden />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 193',
      code: '<input type="submit" aria-label />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 194',
      code: '<input type="submit" aria-labelledby />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 195',
      code: '<input type="submit" aria-live />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 196',
      code: '<input type="submit" aria-owns />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 197',
      code: '<input type="submit" aria-relevant />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 198',
      code: '<input type="checkbox" aria-atomic />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 199',
      code: '<input type="checkbox" aria-busy />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 201',
      code: '<input type="checkbox" aria-controls />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 202',
      code: '<input type="checkbox" aria-describedby />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 204',
      code: '<input type="checkbox" aria-dropeffect />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 205',
      code: '<input type="checkbox" aria-flowto />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 206',
      code: '<input type="checkbox" aria-grabbed />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 207',
      code: '<input type="checkbox" aria-hidden />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 208',
      code: '<input type="checkbox" aria-invalid />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 209',
      code: '<input type="checkbox" aria-label />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 210',
      code: '<input type="checkbox" aria-labelledby />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 211',
      code: '<input type="checkbox" aria-live />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 212',
      code: '<input type="checkbox" aria-owns />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 213',
      code: '<input type="checkbox" aria-relevant />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 214',
      code: '<input type="radio" aria-atomic />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 215',
      code: '<input type="radio" aria-busy />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 217',
      code: '<input type="radio" aria-controls />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 218',
      code: '<input type="radio" aria-describedby />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 219',
      code: '<input type="radio" aria-disabled />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 220',
      code: '<input type="radio" aria-dropeffect />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 221',
      code: '<input type="radio" aria-flowto />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 222',
      code: '<input type="radio" aria-grabbed />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 223',
      code: '<input type="radio" aria-hidden />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 224',
      code: '<input type="radio" aria-label />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 225',
      code: '<input type="radio" aria-labelledby />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 226',
      code: '<input type="radio" aria-live />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 227',
      code: '<input type="radio" aria-owns />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 228',
      code: '<input type="radio" aria-relevant />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 235',
      code: '<input type="range" aria-atomic />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 236',
      code: '<input type="range" aria-busy />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 237',
      code: '<input type="range" aria-controls />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 238',
      code: '<input type="range" aria-describedby />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 239',
      code: '<input type="range" aria-disabled />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 240',
      code: '<input type="range" aria-dropeffect />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 241',
      code: '<input type="range" aria-flowto />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 242',
      code: '<input type="range" aria-grabbed />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 243',
      code: '<input type="range" aria-haspopup />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 244',
      code: '<input type="range" aria-hidden />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 245',
      code: '<input type="range" aria-invalid />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 246',
      code: '<input type="range" aria-label />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 247',
      code: '<input type="range" aria-labelledby />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 248',
      code: '<input type="range" aria-live />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 249',
      code: '<input type="range" aria-owns />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 250',
      code: '<input type="range" aria-relevant />',
    },
    {
      name: 'upstream v6.10.2 role-supports-aria-props valid 278',
      code: '\n        const HelloThere = () => (\n            <Hello\n                role="searchbox"\n                frag={\n                    <>\n                        <div>Hello</div>\n                        <div>There</div>\n                    </>\n                }\n            />\n        );\n\n        const Hello = (props) => <div>{props.frag}</div>;\n      ',
    },
    ...(
      [
        [
          'alert',
          279,
          'aria-atomic aria-live aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-owns aria-relevant aria-roledescription',
        ],
        [
          'alertdialog',
          296,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription aria-modal',
        ],
        [
          'application',
          314,
          'aria-activedescendant aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'article',
          337,
          'aria-posinset aria-setsize aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'banner',
          356,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'blockquote',
          373,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'button',
          390,
          'aria-disabled aria-expanded aria-haspopup aria-pressed aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'caption',
          411,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'cell',
          428,
          'aria-colindex aria-colspan aria-rowindex aria-rowspan aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'checkbox',
          449,
          'aria-checked aria-errormessage aria-expanded aria-invalid aria-readonly aria-required aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription aria-disabled',
        ],
        [
          'code',
          473,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'columnheader',
          490,
          'aria-sort aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription aria-colindex aria-colspan aria-rowindex aria-rowspan aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-readonly aria-required aria-selected',
        ],
        [
          'combobox',
          520,
          'aria-activedescendant aria-autocomplete aria-errormessage aria-invalid aria-readonly aria-required aria-expanded aria-haspopup aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription aria-disabled',
        ],
        [
          'complementary',
          546,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'contentinfo',
          563,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'definition',
          580,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'deletion',
          597,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'dialog',
          614,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription aria-modal',
        ],
        [
          'directory',
          632,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'document',
          649,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'emphasis',
          666,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'feed',
          683,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'figure',
          700,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'form',
          717,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'generic',
          734,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'grid',
          751,
          'aria-multiselectable aria-readonly aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription aria-activedescendant aria-disabled aria-colcount aria-rowcount',
        ],
        [
          'gridcell',
          774,
          'aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-readonly aria-required aria-selected aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription aria-colindex aria-colspan aria-rowindex aria-rowspan',
        ],
        [
          'group',
          803,
          'aria-activedescendant aria-disabled aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'heading',
          823,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'img',
          840,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'insertion',
          857,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'link',
          874,
          'aria-disabled aria-expanded aria-haspopup aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'list',
          894,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'listbox',
          911,
          'aria-errormessage aria-expanded aria-invalid aria-multiselectable aria-readonly aria-required aria-orientation aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription aria-activedescendant aria-disabled',
        ],
        [
          'listitem',
          937,
          'aria-level aria-posinset aria-setsize aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'log',
          957,
          'aria-live aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-owns aria-relevant aria-roledescription',
        ],
        [
          'main',
          974,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'mark',
          991,
          'aria-braillelabel aria-brailleroledescription aria-description aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'marquee',
          1011,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'math',
          1028,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'menu',
          1045,
          'aria-orientation aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription aria-activedescendant aria-disabled',
        ],
        [
          'menubar',
          1065,
          'aria-orientation aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription aria-activedescendant aria-disabled',
        ],
        [
          'menuitem',
          1085,
          'aria-disabled aria-expanded aria-haspopup aria-posinset aria-setsize aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'menuitemcheckbox',
          1107,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription aria-disabled aria-checked aria-errormessage aria-expanded aria-invalid aria-readonly aria-required aria-haspopup aria-posinset aria-setsize',
        ],
        [
          'menuitemradio',
          1134,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription aria-disabled aria-checked aria-errormessage aria-expanded aria-invalid aria-readonly aria-required aria-haspopup aria-posinset aria-setsize',
        ],
        [
          'meter',
          1161,
          'aria-valuetext aria-valuemax aria-valuemin aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription aria-valuenow',
        ],
        [
          'navigation',
          1182,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'note',
          1199,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'option',
          1216,
          'aria-checked aria-posinset aria-setsize aria-selected aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription aria-disabled',
        ],
        [
          'paragraph',
          1238,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'presentation',
          1255,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'progressbar',
          1272,
          'aria-valuetext aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription aria-valuemax aria-valuemin aria-valuenow',
        ],
        [
          'radio',
          1293,
          'aria-checked aria-posinset aria-setsize aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription aria-disabled',
        ],
        [
          'radiogroup',
          1314,
          'aria-errormessage aria-invalid aria-readonly aria-required aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription aria-activedescendant aria-disabled aria-orientation',
        ],
        [
          'region',
          1338,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'row',
          1355,
          'aria-colindex aria-expanded aria-level aria-posinset aria-rowindex aria-selected aria-setsize aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription aria-activedescendant aria-disabled',
        ],
        [
          'rowgroup',
          1381,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'rowheader',
          1398,
          'aria-sort aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription aria-colindex aria-colspan aria-rowindex aria-rowspan aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-readonly aria-required aria-selected',
        ],
        [
          'scrollbar',
          1428,
          'aria-disabled aria-valuetext aria-orientation aria-valuemax aria-valuemin aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription aria-valuenow',
        ],
        [
          'search',
          1451,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'searchbox',
          1468,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription aria-disabled aria-activedescendant aria-autocomplete aria-errormessage aria-haspopup aria-invalid aria-multiline aria-placeholder aria-readonly aria-required',
        ],
        [
          'separator',
          1495,
          'aria-disabled aria-orientation aria-valuemax aria-valuemin aria-valuenow aria-valuetext aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'slider',
          1518,
          'aria-errormessage aria-haspopup aria-invalid aria-readonly aria-valuetext aria-orientation aria-valuemax aria-valuemin aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription aria-disabled aria-valuenow',
        ],
        [
          'spinbutton',
          1545,
          'aria-errormessage aria-invalid aria-readonly aria-required aria-valuetext aria-valuenow aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription aria-activedescendant aria-disabled aria-valuemax aria-valuemin',
        ],
        [
          'status',
          1572,
          'aria-atomic aria-live aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-owns aria-relevant aria-roledescription',
        ],
        [
          'strong',
          1589,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'subscript',
          1606,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'superscript',
          1623,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'switch',
          1640,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription aria-disabled aria-checked aria-errormessage aria-expanded aria-invalid aria-readonly aria-required',
        ],
        [
          'tab',
          1664,
          'aria-disabled aria-expanded aria-haspopup aria-posinset aria-setsize aria-selected aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'table',
          1687,
          'aria-colcount aria-rowcount aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'tablist',
          1706,
          'aria-level aria-multiselectable aria-orientation aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription aria-activedescendant aria-disabled',
        ],
        [
          'tabpanel',
          1728,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'term',
          1745,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'textbox',
          1762,
          'aria-activedescendant aria-autocomplete aria-errormessage aria-haspopup aria-invalid aria-multiline aria-placeholder aria-readonly aria-required aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription aria-disabled',
        ],
        [
          'time',
          1789,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'timer',
          1806,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'toolbar',
          1823,
          'aria-orientation aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription aria-activedescendant aria-disabled',
        ],
        [
          'tooltip',
          1843,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'tree',
          1860,
          'aria-errormessage aria-invalid aria-multiselectable aria-required aria-orientation aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription aria-activedescendant aria-disabled',
        ],
        [
          'treegrid',
          1884,
          'aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription aria-activedescendant aria-disabled aria-multiselectable aria-readonly aria-colcount aria-rowcount aria-orientation aria-errormessage aria-invalid aria-required',
        ],
        [
          'treeitem',
          1911,
          'aria-expanded aria-haspopup aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription aria-level aria-posinset aria-setsize aria-disabled aria-checked aria-selected',
        ],
        [
          'doc-abstract',
          1936,
          'aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'doc-acknowledgments',
          1958,
          'aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'doc-afterword',
          1980,
          'aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'doc-appendix',
          2002,
          'aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'doc-backlink',
          2024,
          'aria-errormessage aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription aria-disabled aria-expanded aria-haspopup',
        ],
        [
          'doc-biblioentry',
          2046,
          'aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription aria-level aria-posinset aria-setsize',
        ],
        [
          'doc-bibliography',
          2071,
          'aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'doc-biblioref',
          2093,
          'aria-errormessage aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription aria-disabled aria-expanded aria-haspopup',
        ],
        [
          'doc-chapter',
          2115,
          'aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'doc-colophon',
          2137,
          'aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'doc-conclusion',
          2159,
          'aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'doc-cover',
          2181,
          'aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'doc-credit',
          2203,
          'aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'doc-credits',
          2225,
          'aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'doc-dedication',
          2247,
          'aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'doc-endnote',
          2269,
          'aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription aria-level aria-posinset aria-setsize',
        ],
        [
          'doc-endnotes',
          2294,
          'aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'doc-epigraph',
          2316,
          'aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'doc-epilogue',
          2338,
          'aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'doc-errata',
          2360,
          'aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'doc-example',
          2382,
          'aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'doc-footnote',
          2404,
          'aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'doc-foreword',
          2426,
          'aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'doc-glossary',
          2448,
          'aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'doc-glossref',
          2470,
          'aria-errormessage aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription aria-disabled aria-expanded aria-haspopup',
        ],
        [
          'doc-index',
          2492,
          'aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'doc-introduction',
          2514,
          'aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'doc-noteref',
          2536,
          'aria-errormessage aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription aria-disabled aria-expanded aria-haspopup',
        ],
        [
          'doc-notice',
          2558,
          'aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'doc-pagebreak',
          2580,
          'aria-errormessage aria-expanded aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription aria-disabled aria-orientation aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-pagefooter',
          2607,
          'aria-braillelabel aria-brailleroledescription aria-description aria-disabled aria-errormessage aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'doc-pageheader',
          2631,
          'aria-braillelabel aria-brailleroledescription aria-description aria-disabled aria-errormessage aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'doc-pagelist',
          2655,
          'aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'doc-part',
          2677,
          'aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'doc-preface',
          2699,
          'aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'doc-prologue',
          2721,
          'aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'doc-qna',
          2743,
          'aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'doc-subtitle',
          2765,
          'aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'doc-tip',
          2787,
          'aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'doc-toc',
          2809,
          'aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'graphics-document',
          2831,
          'aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
        [
          'graphics-object',
          2853,
          'aria-errormessage aria-expanded aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription aria-activedescendant aria-disabled',
        ],
        [
          'graphics-symbol',
          2876,
          'aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-atomic aria-busy aria-controls aria-current aria-describedby aria-details aria-dropeffect aria-flowto aria-grabbed aria-hidden aria-keyshortcuts aria-label aria-labelledby aria-live aria-owns aria-relevant aria-roledescription',
        ],
      ] as const
    ).flatMap(([role, firstCase, values]) =>
      values.split(' ').map((prop, index) => {
        const openingElement = `<div role="${role}" ${prop} />`;
        const code = openingElement;
        return {
          name: `upstream v6.10.2 role-supports-aria-props valid ${firstCase + index}`,
          code,
        };
      }),
    ),
  ],
  invalid: [
    ...(
      [
        [
          'alert',
          37,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'alertdialog',
          71,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'application',
          104,
          'aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'article',
          132,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'banner',
          164,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'blockquote',
          198,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'button',
          232,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-errormessage aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'caption',
          262,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'cell',
          296,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'checkbox',
          326,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-colcount aria-colindex aria-colspan aria-description aria-haspopup aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'code',
          353,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'columnheader',
          387,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-rowcount aria-setsize aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'combobox',
          408,
          'aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'complementary',
          433,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'contentinfo',
          467,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'definition',
          501,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'deletion',
          535,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'dialog',
          569,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'directory',
          602,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'document',
          636,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'emphasis',
          670,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'feed',
          704,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'figure',
          738,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'form',
          772,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'generic',
          806,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'grid',
          840,
          'aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colindex aria-colspan aria-description aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-orientation aria-placeholder aria-posinset aria-pressed aria-required aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'gridcell',
          868,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-rowcount aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'group',
          890,
          'aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'heading',
          922,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'img',
          955,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'insertion',
          989,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'link',
          1023,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription',
        ],
        [
          'link',
          1028,
          'aria-colcount aria-colindex aria-colspan aria-description aria-errormessage aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'list',
          1054,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'listbox',
          1088,
          'aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-haspopup aria-level aria-modal aria-multiline aria-placeholder aria-posinset aria-pressed aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'listitem',
          1113,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'log',
          1144,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'main',
          1178,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'mark',
          1212,
          'aria-activedescendant aria-autocomplete aria-checked aria-colcount aria-colindex aria-colspan aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'marquee',
          1243,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'math',
          1277,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'menu',
          1311,
          'aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'menubar',
          1342,
          'aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'menuitem',
          1373,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-errormessage aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'menuitemcheckbox',
          1402,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-pressed aria-rowcount aria-rowindex aria-rowspan aria-selected aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'menuitemradio',
          1426,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-pressed aria-rowcount aria-rowindex aria-rowspan aria-selected aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'meter',
          1450,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort',
        ],
        [
          'navigation',
          1480,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'none',
          1514,
          'aria-activedescendant aria-atomic aria-autocomplete aria-braillelabel aria-brailleroledescription',
        ],
        [
          'none',
          1520,
          'aria-checked aria-colcount aria-colindex aria-colspan aria-controls aria-current aria-describedby aria-description aria-details aria-disabled aria-dropeffect aria-errormessage aria-expanded aria-flowto aria-grabbed aria-haspopup aria-hidden aria-invalid aria-keyshortcuts aria-label aria-labelledby aria-level aria-live aria-modal aria-multiline aria-multiselectable aria-orientation aria-owns aria-placeholder aria-posinset aria-pressed aria-readonly aria-relevant aria-required aria-roledescription aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'note',
          1565,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'option',
          1599,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-colcount aria-colindex aria-colspan aria-description aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'paragraph',
          1628,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'presentation',
          1662,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription',
        ],
        [
          'presentation',
          1667,
          'aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'progressbar',
          1696,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort',
        ],
        [
          'radio',
          1726,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-colcount aria-colindex aria-colspan aria-description aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'radiogroup',
          1756,
          'aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-expanded aria-haspopup aria-level aria-modal aria-multiline aria-multiselectable aria-placeholder aria-posinset aria-pressed aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'region',
          1783,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'row',
          1817,
          'aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colspan aria-description aria-errormessage aria-haspopup aria-invalid aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-pressed aria-readonly aria-required aria-rowcount aria-rowspan aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'rowgroup',
          1842,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'rowheader',
          1876,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-rowcount aria-setsize aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'scrollbar',
          1897,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort',
        ],
        [
          'search',
          1925,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'searchbox',
          1959,
          'aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-expanded aria-level aria-modal aria-multiselectable aria-orientation aria-posinset aria-pressed aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'separator',
          1983,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort',
        ],
        [
          'slider',
          2011,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-expanded aria-level aria-modal aria-multiline aria-multiselectable aria-placeholder aria-posinset aria-pressed aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort',
        ],
        [
          'spinbutton',
          2035,
          'aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-expanded aria-haspopup aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort',
        ],
        [
          'status',
          2059,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'strong',
          2093,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'subscript',
          2127,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'superscript',
          2161,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'switch',
          2195,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-colcount aria-colindex aria-colspan aria-description aria-haspopup aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'tab',
          2222,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-errormessage aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'table',
          2250,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'tablist',
          2282,
          'aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-errormessage aria-expanded aria-haspopup aria-invalid aria-modal aria-multiline aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'tabpanel',
          2311,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'term',
          2345,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'textbox',
          2379,
          'aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-expanded aria-level aria-modal aria-multiselectable aria-orientation aria-posinset aria-pressed aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'time',
          2403,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'timer',
          2437,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'toolbar',
          2471,
          'aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'tooltip',
          2502,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-disabled aria-errormessage aria-expanded aria-haspopup aria-invalid aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'tree',
          2536,
          'aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-expanded aria-haspopup aria-level aria-modal aria-multiline aria-placeholder aria-posinset aria-pressed aria-readonly aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'treegrid',
          2563,
          'aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colindex aria-colspan aria-description aria-expanded aria-haspopup aria-level aria-modal aria-multiline aria-placeholder aria-posinset aria-pressed aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'treeitem',
          2587,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-colcount aria-colindex aria-colspan aria-description aria-errormessage aria-invalid aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-abstract',
          2613,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-acknowledgments',
          2642,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-afterword',
          2671,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-appendix',
          2700,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-backlink',
          2729,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-biblioentry',
          2758,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-bibliography',
          2784,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-biblioref',
          2813,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-chapter',
          2842,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-colophon',
          2871,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-conclusion',
          2900,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-cover',
          2929,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-credit',
          2958,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-credits',
          2987,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-dedication',
          3016,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-endnote',
          3045,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-endnotes',
          3071,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-epigraph',
          3100,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-epilogue',
          3129,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-errata',
          3158,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-example',
          3187,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-footnote',
          3216,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-foreword',
          3245,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-glossary',
          3274,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-glossref',
          3303,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-index',
          3332,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-introduction',
          3361,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-noteref',
          3390,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-notice',
          3419,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-pagebreak',
          3448,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort',
        ],
        [
          'doc-pagefooter',
          3472,
          'aria-activedescendant aria-autocomplete aria-checked aria-colcount aria-colindex aria-colspan aria-expanded aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-pageheader',
          3499,
          'aria-activedescendant aria-autocomplete aria-checked aria-colcount aria-colindex aria-colspan aria-expanded aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-pagelist',
          3526,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-part',
          3555,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-preface',
          3584,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-prologue',
          3613,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-pullquote',
          3642,
          'aria-activedescendant aria-atomic aria-autocomplete aria-braillelabel aria-brailleroledescription aria-busy aria-checked aria-colcount aria-colindex aria-colspan aria-controls aria-current aria-describedby aria-description aria-details aria-disabled aria-dropeffect aria-errormessage aria-expanded aria-flowto aria-grabbed aria-haspopup aria-hidden aria-invalid aria-keyshortcuts aria-label aria-labelledby aria-level aria-live aria-modal aria-multiline aria-multiselectable aria-orientation aria-owns aria-placeholder aria-posinset aria-pressed aria-readonly aria-relevant aria-required aria-roledescription aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-qna',
          3693,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-subtitle',
          3722,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-tip',
          3751,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'doc-toc',
          3780,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'graphics-document',
          3809,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'graphics-object',
          3838,
          'aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
        [
          'graphics-symbol',
          3866,
          'aria-activedescendant aria-autocomplete aria-braillelabel aria-brailleroledescription aria-checked aria-colcount aria-colindex aria-colspan aria-description aria-level aria-modal aria-multiline aria-multiselectable aria-orientation aria-placeholder aria-posinset aria-pressed aria-readonly aria-required aria-rowcount aria-rowindex aria-rowspan aria-selected aria-setsize aria-sort aria-valuemax aria-valuemin aria-valuenow aria-valuetext',
        ],
      ] as const
    ).flatMap(([role, firstCase, values]) =>
      values.split(' ').map((prop, index) => {
        const openingElement = `<div role="${role}" ${prop} />`;
        const code = openingElement;
        return {
          name: `upstream v6.10.2 role-supports-aria-props invalid ${firstCase + index}`,
          code,
          errors: [
            singleLineError(
              'role-supports-aria-props',
              `The attribute ${prop} is not supported by the role ${role}.`,
              openingElement,
              1,
            ),
          ],
        };
      }),
    ),
  ],
});
