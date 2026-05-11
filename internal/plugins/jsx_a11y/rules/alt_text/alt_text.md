# alt-text

## Rule Details

Enforces that elements that require alternative text — `<img>`, `<object>`,
`<area>`, and `<input type="image">` — have meaningful information available
to screen readers.

The rule reports problems in these scenarios:

- An `<img>` has no `alt` prop, no `aria-label`, and no `aria-labelledby`,
  and is not marked with a presentational role.
- An `<img>` has an `alt` prop, but the value is `undefined`, `false`, the
  literal string `"false"`, or another expression that statically evaluates
  to falsy.
- An `<img>` declares `aria-label` or `aria-labelledby` but its value is
  empty or explicitly `undefined` (the rule still reports because that
  attribute can't actually serve as the text alternative).
- An `<img>` uses `role="presentation"` or `role="none"` instead of writing
  `alt=""` for a decorative image.
- An `<object>` has no inner accessible content, no `title`, no
  `aria-label`, and no `aria-labelledby`.
- An `<area>` or `<input type="image">` has no `alt`, no `aria-label`, and
  no `aria-labelledby`.

Examples of **incorrect** code for this rule:

```jsx
<img src="foo.png" />
<img alt />
<img alt={undefined} />
<img alt="false" />
<img role="presentation" />
<img aria-label="" />
<area />
<input type="image" />
<object />
<object aria-label="" />
<object title={undefined} />
```

Examples of **correct** code for this rule:

```jsx
<img alt="A descriptive image" />
<img alt="" />
<img alt={alt} />
<img aria-label="An image" />
<img aria-labelledby="caption-1" />
<area alt="A clickable region" />
<input type="image" alt="Submit" />
<object>Inner descriptive content</object>
<object aria-label="Embedded content" />
<object title="An object" />
```

## Rule Options

```json
{
  "jsx-a11y/alt-text": [
    "error",
    {
      "elements": ["img", "object", "area", "input[type=\"image\"]"],
      "img": ["Image"],
      "object": ["Object"],
      "area": ["Area"],
      "input[type=\"image\"]": ["InputImage"]
    }
  ]
}
```

### `elements`

Type: `string[]`. Default:
`["img", "object", "area", "input[type=\"image\"]"]`.

Whitelist of DOM elements the rule should validate. Set this to a subset to
disable the check for specific elements, or to `[]` to disable the rule
entirely.

### Per-element custom-component lists

Each element key (`img`, `object`, `area`, `input[type="image"]`) accepts an
array of component names. Listed components are validated alongside the
matching DOM element — for example, `"img": ["Thumbnail", "Image"]` makes
`<Thumbnail />` and `<Image />` follow the same `alt`-text rules as `<img>`.

Examples of **incorrect** code with `{ "img": ["Image"] }`:

```json
{ "jsx-a11y/alt-text": ["error", { "img": ["Image"] }] }
```

```jsx
<Image />
<Image src="xyz" />
<Image alt={undefined} />
```

Examples of **correct** code with `{ "img": ["Image"] }`:

```json
{ "jsx-a11y/alt-text": ["error", { "img": ["Image"] }] }
```

```jsx
<Image alt="" />
<Image alt="A descriptive image" />
<Image alt={alt} />
```

## Settings

The rule reads two `settings['jsx-a11y']` keys when resolving the effective
element type for a JSX tag:

- `polymorphicPropName` — name of a polymorphic prop (e.g. `"as"`) that
  remaps the element type. With `polymorphicPropName: "as"`,
  `<SomeComponent as="img" />` is treated as `<img />`.
- `polymorphicAllowList` — `string[]` restricting which raw component names
  may be remapped via the polymorphic prop. When omitted, every component
  may be remapped.
- `components` — a `{ ComponentName: "html-element" }` map. With
  `{ "Input": "input" }`, `<Input type="image" />` is treated as
  `<input type="image" />`.

These mirror the upstream `eslint-plugin-jsx-a11y` settings exactly.

## Original Documentation

- [eslint-plugin-jsx-a11y/alt-text](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/main/docs/rules/alt-text.md)
