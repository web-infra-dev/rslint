import { RuleTester } from '../rule-tester';

const errorMessage = (
  attr: string,
  role: string,
  tag: string,
  isImplicit: boolean,
) =>
  isImplicit
    ? `The attribute ${attr} is not supported by the role ${role}. This role is implicit on the element ${tag}.`
    : `The attribute ${attr} is not supported by the role ${role}.`;

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

new RuleTester().run('role-supports-aria-props', null as never, {
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
      errors: [{ message: errorMessage('aria-checked', 'link', 'a', true) }],
    },
    {
      code: '<area href="#" aria-checked />',
      errors: [{ message: errorMessage('aria-checked', 'link', 'area', true) }],
    },
    {
      code: '<link href="#" aria-checked />',
      errors: [{ message: errorMessage('aria-checked', 'link', 'link', true) }],
    },
    {
      code: '<img alt="foobar" aria-checked />',
      errors: [{ message: errorMessage('aria-checked', 'img', 'img', true) }],
    },
    {
      code: '<menu type="toolbar" aria-checked />',
      errors: [
        { message: errorMessage('aria-checked', 'toolbar', 'menu', true) },
      ],
    },
    {
      code: '<aside aria-checked />',
      errors: [
        {
          message: errorMessage('aria-checked', 'complementary', 'aside', true),
        },
      ],
    },
    {
      code: '<ul aria-expanded />',
      errors: [{ message: errorMessage('aria-expanded', 'list', 'ul', true) }],
    },
    {
      code: '<details aria-expanded />',
      errors: [
        { message: errorMessage('aria-expanded', 'group', 'details', true) },
      ],
    },
    {
      code: '<dialog aria-expanded />',
      errors: [
        { message: errorMessage('aria-expanded', 'dialog', 'dialog', true) },
      ],
    },
    {
      code: '<aside aria-expanded />',
      errors: [
        {
          message: errorMessage(
            'aria-expanded',
            'complementary',
            'aside',
            true,
          ),
        },
      ],
    },
    {
      code: '<article aria-expanded />',
      errors: [
        { message: errorMessage('aria-expanded', 'article', 'article', true) },
      ],
    },
    {
      code: '<body aria-expanded />',
      errors: [
        { message: errorMessage('aria-expanded', 'document', 'body', true) },
      ],
    },
    {
      code: '<li aria-expanded />',
      errors: [
        { message: errorMessage('aria-expanded', 'listitem', 'li', true) },
      ],
    },
    {
      code: '<nav aria-expanded />',
      errors: [
        { message: errorMessage('aria-expanded', 'navigation', 'nav', true) },
      ],
    },
    {
      code: '<ol aria-expanded />',
      errors: [{ message: errorMessage('aria-expanded', 'list', 'ol', true) }],
    },
    {
      code: '<output aria-expanded />',
      errors: [
        { message: errorMessage('aria-expanded', 'status', 'output', true) },
      ],
    },
    {
      code: '<section aria-expanded />',
      errors: [
        { message: errorMessage('aria-expanded', 'region', 'section', true) },
      ],
    },
    {
      code: '<tbody aria-expanded />',
      errors: [
        { message: errorMessage('aria-expanded', 'rowgroup', 'tbody', true) },
      ],
    },
    {
      code: '<tfoot aria-expanded />',
      errors: [
        { message: errorMessage('aria-expanded', 'rowgroup', 'tfoot', true) },
      ],
    },
    {
      code: '<thead aria-expanded />',
      errors: [
        { message: errorMessage('aria-expanded', 'rowgroup', 'thead', true) },
      ],
    },
    {
      code: '<input type="radio" aria-invalid />',
      errors: [
        { message: errorMessage('aria-invalid', 'radio', 'input', true) },
      ],
    },
    {
      code: '<input type="radio" aria-selected />',
      errors: [
        { message: errorMessage('aria-selected', 'radio', 'input', true) },
      ],
    },
    {
      code: '<input type="radio" aria-haspopup />',
      errors: [
        { message: errorMessage('aria-haspopup', 'radio', 'input', true) },
      ],
    },
    {
      code: '<input type="checkbox" aria-haspopup />',
      errors: [
        { message: errorMessage('aria-haspopup', 'checkbox', 'input', true) },
      ],
    },
    {
      code: '<input type="reset" aria-invalid />',
      errors: [
        { message: errorMessage('aria-invalid', 'button', 'input', true) },
      ],
    },
    {
      code: '<input type="submit" aria-invalid />',
      errors: [
        { message: errorMessage('aria-invalid', 'button', 'input', true) },
      ],
    },
    {
      code: '<input type="image" aria-invalid />',
      errors: [
        { message: errorMessage('aria-invalid', 'button', 'input', true) },
      ],
    },
    {
      code: '<input type="button" aria-invalid />',
      errors: [
        { message: errorMessage('aria-invalid', 'button', 'input', true) },
      ],
    },
    {
      code: '<menuitem type="command" aria-invalid />',
      errors: [
        {
          message: errorMessage('aria-invalid', 'menuitem', 'menuitem', true),
        },
      ],
    },
    {
      code: '<menuitem type="radio" aria-selected />',
      errors: [
        {
          message: errorMessage(
            'aria-selected',
            'menuitemradio',
            'menuitem',
            true,
          ),
        },
      ],
    },
    {
      code: '<menu type="toolbar" aria-haspopup />',
      errors: [
        { message: errorMessage('aria-haspopup', 'toolbar', 'menu', true) },
      ],
    },
    {
      code: '<menu type="toolbar" aria-invalid />',
      errors: [
        { message: errorMessage('aria-invalid', 'toolbar', 'menu', true) },
      ],
    },
    {
      code: '<menu type="toolbar" aria-expanded />',
      errors: [
        { message: errorMessage('aria-expanded', 'toolbar', 'menu', true) },
      ],
    },
    {
      code: '<link href="#" aria-invalid />',
      errors: [{ message: errorMessage('aria-invalid', 'link', 'link', true) }],
    },
    {
      code: '<area href="#" aria-invalid />',
      errors: [{ message: errorMessage('aria-invalid', 'link', 'area', true) }],
    },
    {
      code: '<a href="#" aria-invalid />',
      errors: [{ message: errorMessage('aria-invalid', 'link', 'a', true) }],
    },
    {
      code: '<Link href="#" aria-checked />',
      settings: componentsSettings,
      errors: [{ message: errorMessage('aria-checked', 'link', 'a', true) }],
    },

    // ---- Explicit role; isImplicit=false branch. ----
    {
      code: '<div role="link" aria-checked />',
      errors: [{ message: errorMessage('aria-checked', 'link', 'div', false) }],
    },
    // ---- Explicit role on custom element (no IsDOMElement gate). ----
    {
      code: '<Foo role="link" aria-checked />',
      errors: [{ message: errorMessage('aria-checked', 'link', 'Foo', false) }],
    },

    // ---- Multiple invalid props on the same element — one report per. ----
    {
      code: '<a href="#" aria-checked aria-pressed />',
      errors: [
        { message: errorMessage('aria-checked', 'link', 'a', true) },
        { message: errorMessage('aria-pressed', 'link', 'a', true) },
      ],
    },

    // ---- Non-nullish ARIA prop values still trigger reports. ----
    {
      code: '<a href="/" aria-checked={false} />',
      errors: [{ message: errorMessage('aria-checked', 'link', 'a', true) }],
    },
    {
      code: '<a href="/" aria-checked="" />',
      errors: [{ message: errorMessage('aria-checked', 'link', 'a', true) }],
    },
    {
      code: '<a href="/" aria-checked={0} />',
      errors: [{ message: errorMessage('aria-checked', 'link', 'a', true) }],
    },
    {
      code: '<a href="/" aria-checked={someVar} />',
      errors: [{ message: errorMessage('aria-checked', 'link', 'a', true) }],
    },
    {
      code: '<a href="/" aria-checked={fn()} />',
      errors: [{ message: errorMessage('aria-checked', 'link', 'a', true) }],
    },

    // ---- Non-null assertion on a nullish value — synthesizes a non-null
    //      string ("null!") so the attr proceeds to the membership check.
    {
      code: '<a href="/" aria-checked={null!} />',
      errors: [{ message: errorMessage('aria-checked', 'link', 'a', true) }],
    },

    // ---- Spread + literal containing role + outside aria-* report. ----
    {
      code: '<div {...{role: "link"}} aria-checked />',
      errors: [{ message: errorMessage('aria-checked', 'link', 'div', false) }],
    },

    // ---- Components map + polymorphic invalid cases. ----
    {
      code: '<Box as="a" href="/" aria-checked />',
      settings: componentsAndPolymorphicSettings,
      errors: [{ message: errorMessage('aria-checked', 'link', 'a', true) }],
    },
    {
      code: '<Anchor href="/" aria-checked />',
      settings: componentsAndPolymorphicSettings,
      errors: [{ message: errorMessage('aria-checked', 'link', 'a', true) }],
    },

    // ---- Nested element where only the inner reports. ----
    {
      code: `
<div>
  <a href="/" aria-checked />
</div>
      `,
      errors: [{ message: errorMessage('aria-checked', 'link', 'a', true) }],
    },

    // ---- select multiple / size>1 → listbox; aria-checked invalid. ----
    {
      code: '<select multiple aria-checked />',
      errors: [
        { message: errorMessage('aria-checked', 'listbox', 'select', true) },
      ],
    },
    {
      code: '<select size={5} aria-checked />',
      errors: [
        { message: errorMessage('aria-checked', 'listbox', 'select', true) },
      ],
    },

    // ---- presentation / none — restricted / empty supported sets. ----
    {
      code: '<div role="presentation" aria-checked />',
      errors: [
        {
          message: errorMessage('aria-checked', 'presentation', 'div', false),
        },
      ],
    },
    {
      code: '<div role="none" aria-busy />',
      errors: [{ message: errorMessage('aria-busy', 'none', 'div', false) }],
    },

    // ---- bare <button> / <img> / <details> / <datalist> implicit roles. ----
    {
      code: '<button aria-checked />',
      errors: [
        { message: errorMessage('aria-checked', 'button', 'button', true) },
      ],
    },
    {
      code: '<img aria-checked />',
      errors: [{ message: errorMessage('aria-checked', 'img', 'img', true) }],
    },
    {
      code: '<details aria-checked />',
      errors: [
        { message: errorMessage('aria-checked', 'group', 'details', true) },
      ],
    },
    {
      code: '<datalist aria-checked />',
      errors: [
        { message: errorMessage('aria-checked', 'listbox', 'datalist', true) },
      ],
    },
  ],
});
