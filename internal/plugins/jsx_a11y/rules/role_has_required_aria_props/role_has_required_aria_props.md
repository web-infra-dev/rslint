# role-has-required-aria-props

## Rule Details

Elements with ARIA roles must have every required `aria-*` attribute for that role defined. The required-attributes table is sourced from [`aria-query`](https://github.com/A11yance/aria-query)'s `rolesMap`.

The rule reads each JSX attribute named `role` (case-insensitive). When the attribute's value is a literal string, it lowercases the value and splits on single ASCII spaces. For each space-delimited token that is a recognized non-abstract ARIA role with non-empty required attributes, it verifies that every required `aria-*` attribute is present on the same element. Tokens that are not valid ARIA roles, and roles whose required-attribute set is empty (e.g. `button`, `row`), are ignored.

Elements that natively imply the role are skipped. For example, `<input type="checkbox">` already provides the `checkbox` role, so `<input type="checkbox" role="switch" />` is not flagged for missing `aria-checked`.

Examples of **incorrect** code for this rule:

```jsx
<div role="checkbox" />
<div role="combobox" />
<div role="heading" />
<div role="option" />
<div role="scrollbar" aria-valuemax aria-valuemin />
<span role="checkbox" aria-labelledby="foo" tabindex="0"></span>
```

Examples of **correct** code for this rule:

```jsx
<div />
<div role={role} />
<div role="row" />
<div role="heading" aria-level={2} />
<span role="checkbox" aria-checked="false" aria-labelledby="foo" tabindex="0"></span>
<input type="checkbox" role="switch" />
```

## Rule Options

This rule takes no options.

## Accessibility guidelines

- [WCAG 4.1.2](https://www.w3.org/WAI/WCAG21/Understanding/name-role-value)

### Resources

- [ARIA Spec, Roles](https://www.w3.org/TR/wai-aria/#roles)
- [Chrome Audit Rules, AX_ARIA_03](https://github.com/GoogleChrome/accessibility-developer-tools/wiki/Audit-Rules#ax_aria_03)

## Original Documentation

- [eslint-plugin-jsx-a11y/role-has-required-aria-props](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/role-has-required-aria-props.md)
