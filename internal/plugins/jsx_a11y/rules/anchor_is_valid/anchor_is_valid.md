# anchor-is-valid

## Rule Details

This rule enforces that `<a>` elements (and any configured custom anchor
components) function as proper hyperlinks. An anchor without a usable
`href`, with a non-navigable `href` such as `""`, `"#"`, or `javascript:…`,
or used purely for click handling, is reported.

The rule runs three independent aspects, all enabled by default:

- **noHref** — reports when no usable href is provided.
- **invalidHref** — reports when the `href` value is empty, the literal
  `"#"`, or matches `^\W*?javascript:` (e.g. `"javascript:void(0)"`).
- **preferButton** — when an `onClick` is present, reports that the element
  should be a `<button>` instead of an anchor.

A spread attribute (`{...props}`, including a literal-object spread
without a usable href) suppresses the `noHref` and `preferButton` reports
when no explicit href is found, because the spread may carry one at
runtime.

When more than one aspect would trigger on the same element,
`preferButton` takes precedence over `invalidHref` (and `preferButton`
takes precedence over `noHref` in the no-href branch).

Examples of **incorrect** code for this rule:

```jsx
<a />
<a href={undefined} />
<a href={null} />
<a href="" />
<a href="#" />
<a href={"#"} />
<a href="javascript:void(0)" />
<a href={"javascript:void(0)"} />
<a onClick={() => void 0} />
<a href="#" onClick={() => void 0} />
<a href="javascript:void(0)" onClick={() => void 0} />
```

Examples of **correct** code for this rule:

```jsx
<a href="https://example.com" />
<a href="/foo" />
<a href="#section" />
<a href={url} />
<a href="foo" onClick={() => void 0} />
<a {...props} />
```

## Rule Options

The rule accepts an options object with the following properties:

- `components` — array of additional component names (besides the built-in
  `a`) that should be checked.
- `specialLink` — array of additional prop names (besides the built-in
  `href`) that should be treated as href-like for the purposes of this
  rule.
- `aspects` — array of which sub-checks to run; one or more of `noHref`,
  `invalidHref`, `preferButton`. When omitted, all three are active.

Examples of **incorrect** code with `{ "components": ["Anchor", "Link"] }`:

```json
{ "anchor-is-valid": ["error", { "components": ["Anchor", "Link"] }] }
```

```jsx
<Link />
<Anchor href="#" />
```

Examples of **incorrect** code with `{ "specialLink": ["hrefLeft"] }`:

```json
{ "anchor-is-valid": ["error", { "specialLink": ["hrefLeft"] }] }
```

```jsx
<a hrefLeft="" />
<a hrefLeft="javascript:void(0)" />
```

Examples of **incorrect** code with `{ "aspects": ["noHref"] }`:

```json
{ "anchor-is-valid": ["error", { "aspects": ["noHref"] }] }
```

```jsx
<a />
<a onClick={() => void 0} />
```

With `aspects` set to `["noHref"]`, an anchor that carries an `onClick`
but no `href` is still reported as `noHref` — `preferButton` is off, so
the `onClick` does not redirect the report.

## Original Documentation

- [eslint-plugin-jsx-a11y/anchor-is-valid](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md)
