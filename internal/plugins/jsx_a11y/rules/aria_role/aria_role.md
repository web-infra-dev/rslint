# aria-role

## Rule Details

Elements with ARIA roles must use a valid, non-abstract ARIA role. A reference to role definitions can be found at [WAI-ARIA](https://www.w3.org/TR/wai-aria-1.2/#role_definitions).

The rule reads each JSX attribute named `role` (case-insensitive) on any JSX element, splits the literal value by single spaces, and reports if any space-delimited token is not a recognized non-abstract ARIA role (per [`aria-query`](https://github.com/A11yance/aria-query)'s `rolesMap`) and is not on the user-provided `allowedInvalidRoles` allow-list.

Examples of **incorrect** code for this rule:

```jsx
<div role="datepicker" />
<div role="range" />
<div role="" />
<div role={null} />
<div role />
<div role="tabpanel row foobar" />
```

Examples of **correct** code for this rule:

```jsx
<div />
<div role={role} />
<div role={role || "button"} />
<div role="tabpanel row" />
<div role="switch" />
<div role="doc-abstract" />
<div role="graphics-document document" />
```

## Rule Options

```json
{
  "jsx-a11y/aria-role": [
    "error",
    {
      "allowedInvalidRoles": ["text"],
      "ignoreNonDOM": true
    }
  ]
}
```

### `allowedInvalidRoles`

An optional array of role names that should be treated as valid in addition to the ARIA spec — useful when an application defines its own role conventions or uses non-standard roles for [text-splitting](https://axesslab.com/text-splitting).

Examples of **correct** code with `{ "allowedInvalidRoles": ["invalid-role", "other-invalid-role"] }`:

```json
{ "jsx-a11y/aria-role": ["error", { "allowedInvalidRoles": ["invalid-role", "other-invalid-role"] }] }
```

```jsx
<img role="invalid-role" />
<img role="invalid-role tabpanel" />
<img role="invalid-role other-invalid-role" />
```

### `ignoreNonDOM`

When set to `true`, the rule skips elements whose resolved type is not a standard HTML element (per `aria-query`'s `dom` map). Useful for designs that wrap accessibility primitives in custom components and want validation only on real DOM elements.

Examples of **correct** code with `{ "ignoreNonDOM": true }`:

```json
{ "jsx-a11y/aria-role": ["error", { "ignoreNonDOM": true }] }
```

```jsx
<Foo role="bar" />
<fakeDOM role="bar" />
```

Examples of **incorrect** code with `{ "ignoreNonDOM": true }` (a real DOM element is still checked):

```json
{ "jsx-a11y/aria-role": ["error", { "ignoreNonDOM": true }] }
```

```jsx
<img role="invalid-role" />
```

## Accessibility guidelines

- [WCAG 4.1.2](https://www.w3.org/WAI/WCAG21/Understanding/name-role-value)

### Resources

- [Chrome Audit Rules, AX_ARIA_01](https://github.com/GoogleChrome/accessibility-developer-tools/wiki/Audit-Rules#ax_aria_01)
- [DPUB-ARIA roles](https://www.w3.org/TR/dpub-aria-1.0/)
- [MDN: Using ARIA: Roles, states, and properties](https://developer.mozilla.org/en-US/docs/Web/Accessibility/ARIA/ARIA_Techniques)

## Original Documentation

- [eslint-plugin-jsx-a11y/aria-role](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/aria-role.md)
