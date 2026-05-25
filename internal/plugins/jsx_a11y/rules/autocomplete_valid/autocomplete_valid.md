# autocomplete-valid

The `autocomplete` attribute on form controls follows a strict grammar from
the [WHATWG HTML autofill detail tokens](https://html.spec.whatwg.org/multipage/form-control-infrastructure.html#autofill-detail-tokens).
Values that don't match the grammar give browsers no useful hint and break
autofill for users who rely on it — this rule reports such values.

## Rule Details

This rule applies to `<input>` elements (plus any custom components opted
in via the `inputComponents` option or the `jsx-a11y` `components` /
`polymorphicPropName` settings). It accepts:

- the empty string, `on`, `off`, and a few state-like aliases (`none`,
  `null`, `disabled`, `enabled`, `undefined`, `true`, `false`, `xon`,
  `xoff`);
- a single autofill field-name token (e.g. `name`, `email`, `street-address`,
  `current-password`);
- a token sequence of the form
  `[section-* ]? [billing|shipping]? [home|work|mobile|fax|pager]? <field-name> [webauthn]?`
  where `<field-name>` must be drawn from the qualified list when a
  contact qualifier (`home`, `work`, …) is present.

Comparison is case-insensitive and surrounding whitespace is trimmed.
Dynamic values (`autocomplete={someVar}`) are not checked.

Examples of **incorrect** code for this rule:

```jsx
<input type="text" autocomplete="foo" />
<input type="text" autocomplete="name invalid" />
<input type="text" autocomplete="invalid name" />
<input type="text" autocomplete="home url" />
```

Examples of **correct** code for this rule:

```jsx
<input type="text" autocomplete="name" />
<input type="text" autocomplete="" />
<input type="text" autocomplete="off" />
<input type="text" autocomplete="on" />
<input type="text" autocomplete="billing family-name" />
<input type="text" autocomplete="section-blue shipping street-address" />
<input type="text" autocomplete="section-somewhere shipping work email" />
<input type="text" autocomplete />
<input type="text" autocomplete={dynamicValue} />
<Foo autocomplete="bar" />
```

The rule does not run on inputs whose `type` is `submit`, `reset`,
`button`, or `hidden` — those controls don't carry autofill data:

```jsx
<input type="hidden" autocomplete="foo" />
<input type="submit" autocomplete="foo" />
```

## Rule Options

```json
{
  "jsx-a11y/autocomplete-valid": [
    "error",
    {
      "inputComponents": ["MyInput"]
    }
  ]
}
```

### `inputComponents`

Type: `string[]`. Default: `[]`.

Component names that should be validated as if they were `<input>` elements.
Components not listed are ignored.

Examples of **incorrect** code with `{ "inputComponents": ["MyInput"] }`:

```json
{ "jsx-a11y/autocomplete-valid": ["error", { "inputComponents": ["MyInput"] }] }
```

```jsx
<MyInput autocomplete="foo" />
```

Examples of **correct** code with `{ "inputComponents": ["MyInput"] }`:

```json
{ "jsx-a11y/autocomplete-valid": ["error", { "inputComponents": ["MyInput"] }] }
```

```jsx
<MyInput autocomplete="name" />
<MyInput autocomplete={dynamicValue} />
```

## Settings

The rule reads `settings['jsx-a11y']` to resolve the effective element
type for a JSX tag:

- `polymorphicPropName` — name of a polymorphic prop (e.g. `"as"`) that
  remaps the element type. With `polymorphicPropName: "as"`,
  `<Box as="input" />` is validated as `<input />`.
- `polymorphicAllowList` — `string[]` restricting which raw component
  names may be remapped via the polymorphic prop. When omitted, every
  component may be remapped.
- `components` — a `{ ComponentName: "html-element" }` map. With
  `{ "Input": "input" }`, `<Input />` is validated as `<input />`.

These mirror the upstream `eslint-plugin-jsx-a11y` settings exactly.

## Original Documentation

- [eslint-plugin-jsx-a11y/autocomplete-valid](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/autocomplete-valid.md)
