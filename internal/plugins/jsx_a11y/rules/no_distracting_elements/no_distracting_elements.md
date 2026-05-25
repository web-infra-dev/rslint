# no-distracting-elements

## Rule Details

Enforces that no distracting elements are used. Elements that can be visually
distracting can cause accessibility issues for visually impaired users — they
are also typically deprecated in the HTML specification. By default, this
rule reports `<marquee>` and `<blink>`.

The check is case-sensitive: only the lowercase intrinsic tags `marquee` and
`blink` are flagged. React components named `Marquee` or `Blink` are
considered safe unless mapped via the `jsx-a11y` `components` setting.

Examples of **incorrect** code for this rule:

```jsx
<marquee />
<blink />
<marquee>scrolling text</marquee>
```

Examples of **correct** code for this rule:

```jsx
<div />
<Marquee />
<div marquee />
```

## Rule Options

```json
{
  "jsx-a11y/no-distracting-elements": [
    "error",
    {
      "elements": ["marquee", "blink"]
    }
  ]
}
```

### `elements`

Type: `string[]`. Default: `["marquee", "blink"]`.

The list of intrinsic tag names to flag. Provide a subset to disable the
check for specific tags, or `[]` to disable the rule entirely.

Examples of **incorrect** code with `{ "elements": ["blink"] }`:

```json
{ "jsx-a11y/no-distracting-elements": ["error", { "elements": ["blink"] }] }
```

```jsx
<blink />
```

Examples of **correct** code with `{ "elements": ["blink"] }`:

```json
{ "jsx-a11y/no-distracting-elements": ["error", { "elements": ["blink"] }] }
```

```jsx
<marquee />
```

## Settings

The rule reads `settings['jsx-a11y']` to resolve the effective element type
for a JSX tag. Resolution is two-step: the polymorphic prop runs first,
then the components map looks up the (possibly already-replaced) name.

- `polymorphicPropName` — name of a polymorphic prop (e.g. `"as"`) that
  remaps the element type. With `polymorphicPropName: "as"`,
  `<Box as="marquee" />` is flagged as `<marquee />`. Reverse direction is
  supported too — `<marquee as="div" />` resolves to `"div"` and is **not**
  flagged.
- `polymorphicAllowList` — `string[]` restricting which raw component names
  may be remapped via the polymorphic prop. When omitted, every component
  may be remapped.
- `components` — a `{ ComponentName: "html-element" }` map. With
  `{ "Blink": "blink" }`, `<Blink />` is flagged as `<blink />`. The map can
  also remap an intrinsic tag away from the flagged set — e.g.
  `{ "marquee": "div" }` causes `<marquee />` to be treated as `<div />` and
  the rule will not report it. When combined with `polymorphicPropName`,
  the lookup runs against the post-polymorphic name — e.g.
  `<Foo as="Bar" />` with `{ "components": { "Bar": "marquee" } }` is
  flagged as `<marquee />`.

These mirror the upstream `eslint-plugin-jsx-a11y` settings exactly.

## Resources

- [WCAG 2.2.2 — Pause, Stop, Hide](https://www.w3.org/WAI/WCAG21/Understanding/pause-stop-hide)
- [Deque University — `<marquee>`](https://dequeuniversity.com/rules/axe/3.2/marquee)
- [Deque University — `<blink>`](https://dequeuniversity.com/rules/axe/3.2/blink)

## Original Documentation

- [eslint-plugin-jsx-a11y/no-distracting-elements](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/no-distracting-elements.md)
