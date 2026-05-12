# aria-proptypes

Enforce that the literal value of every recognized `aria-*` state / property
matches the type declared in [aria-query](https://github.com/A11yance/aria-query)'s
`ariaPropsMap`. For example, `aria-hidden` is a boolean (`true` / `false` /
`"true"` / `"false"`) — `aria-hidden="yes"` reports.

## Rule Details

The rule fires on a JSX attribute whose name (after lowercasing) starts with
`aria-` and is one of the canonical ARIA states or properties. The value is
extracted via jsx-ast-utils' `getLiteralPropValue` semantics, so only static
literal values are inspected — expressions like `<div aria-hidden={someVar} />`,
`<div aria-label={fn()} />`, `<div aria-checked={cond ? "true" : "false"} />`
are not classified by this rule (the value isn't statically a literal).

Values whose static result is `null` or `undefined` are also skipped — those
are treated as "no value" rather than an invalid value.

Type-specific rules:

- **boolean**: must be `true` / `false` or one of the strings `"true"` /
  `"false"` (case-insensitive).
- **string** / **id**: must be a string. Empty string is valid.
- **integer** / **number**: must coerce to a non-NaN number under JS
  `Number(value)` semantics, and must not be a boolean.
- **tristate**: boolean, OR the literal string `"mixed"`.
- **token**: must be one of the permitted values (case-insensitive when
  string; strict-equal when boolean). Permitted lists are heterogeneous —
  `aria-haspopup` accepts both booleans and strings.
- **tokenlist**: must be a space-separated string where every token is in
  the permitted list (case-insensitive).
- **idlist**: must be a string.

Examples of **incorrect** code for this rule:

```jsx
<div aria-hidden="yes" />
<div aria-label />
<div aria-checked={1234} />
<div aria-level="yes" />
<div aria-sort="descnding" />
<div aria-sort="ascending descending" />
<div aria-relevant="additions removalss" />
```

Examples of **correct** code for this rule:

```jsx
<div aria-hidden={true} />
<div aria-hidden="false" />
<div aria-hidden />
<div aria-label="Close" />
<div aria-checked="mixed" />
<div aria-level={123} />
<div aria-level="3" />
<div aria-sort="ascending" />
<div aria-sort="ASCENDING" />
<div aria-relevant="additions removals text" />
<div aria-invalid={true} />
<div aria-invalid="grammar" />
```

Non-literal values are not flagged — the rule trusts the developer when
the value isn't statically determinable:

```jsx
<div aria-hidden={someVar} />
<div aria-level={getLevel()} />
<div aria-sort={cond ? "ascending" : "descending"} />
```

`null` and `undefined` are silently ignored:

```jsx
<div aria-hidden={null} />
<div aria-hidden={undefined} />
```

## Rule Options

This rule takes no options.

## Resources

- [WAI-ARIA 1.2 — States and Properties](https://www.w3.org/TR/wai-aria-1.2/#state_prop_def)
- [WCAG 2.2 — 4.1.2 Name, Role, Value](https://www.w3.org/WAI/WCAG22/Understanding/name-role-value.html)
- [MDN — ARIA attributes](https://developer.mozilla.org/en-US/docs/Web/Accessibility/ARIA/Reference/Attributes)
- [aria-query — `ariaPropsMap`](https://github.com/A11yance/aria-query/blob/HEAD/src/ariaPropsMap.js)

## Original Documentation

- [eslint-plugin-jsx-a11y/aria-proptypes](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/aria-proptypes.md)
