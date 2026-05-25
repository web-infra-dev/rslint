# iframe-has-title

## Rule Details

This rule enforces that every `<iframe>` element carries a non-empty `title`
attribute so screen-reader users can identify the embedded frame's purpose.
An iframe without a usable title violates
[WCAG 2.4.1](https://www.w3.org/WAI/WCAG21/Understanding/bypass-blocks) and
[WCAG 4.1.2](https://www.w3.org/WAI/WCAG21/Understanding/name-role-value).

The element name is resolved through the standard jsx-a11y settings —
`components` and `polymorphicPropName` — so a custom component mapped to
`iframe` (e.g. `<FooComponent />` with
`settings: { 'jsx-a11y': { components: { FooComponent: 'iframe' } } }`)
is checked the same way as a literal `<iframe>` tag.

The `title` value is required to satisfy `value && typeof value === 'string'`.
Truthy non-string values (numbers, booleans, plain objects, arrays, RegExps,
functions) all fail this gate. Notably:

- `<iframe title />` is reported (boolean attribute form coerces to `true`).
- `<iframe title={42} />` and `<iframe title={true} />` are reported.
- Empty-string forms — `<iframe title="" />`, `<iframe title={""} />`, and
  the empty template literal — are all reported.
- `<iframe title={undefined} />` is reported.

Examples of **incorrect** code for this rule:

```jsx
<iframe />
<iframe {...props} />
<iframe title={undefined} />
<iframe title="" />
<iframe title={false} />
<iframe title={true} />
<iframe title={42} />
<iframe title={``} />
```

Examples of **correct** code for this rule:

```jsx
<iframe title="Unique title" />
<iframe title={titleText} />
<iframe title={`Frame for ${name}`} />
```

## Accessibility guidelines

- [WCAG 2.4.1](https://www.w3.org/WAI/WCAG21/Understanding/bypass-blocks)
- [WCAG 4.1.2](https://www.w3.org/WAI/WCAG21/Understanding/name-role-value)

### Resources

- [axe-core, frame-title](https://dequeuniversity.com/rules/axe/3.2/frame-title)
- [H64: Using the title attribute of the frame and iframe elements](https://www.w3.org/TR/WCAG20-TECHS/H64.html)

## Original Documentation

- [eslint-plugin-jsx-a11y/iframe-has-title](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/iframe-has-title.md)
