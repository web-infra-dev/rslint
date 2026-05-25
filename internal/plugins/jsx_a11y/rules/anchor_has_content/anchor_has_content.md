# anchor-has-content

## Rule Details

This rule enforces that anchors (`<a>` elements and any configured custom
anchor components) have accessible content — text or descendant elements
that screen readers can announce. An empty anchor that lacks both a `title`
and an `aria-label` is reported.

An anchor is considered to have accessible content when at least one of the
following holds:

- It contains non-empty text or a non-hidden child element.
- It contains a JSX expression whose value is anything other than the bare
  identifier `undefined`.
- It declares a `dangerouslySetInnerHTML` or `children` prop.
- It declares a `title` or `aria-label` attribute.

Examples of **incorrect** code for this rule:

```jsx
<a />
<a><Bar aria-hidden /></a>
<a>{undefined}</a>
```

Examples of **correct** code for this rule:

```jsx
<a>Anchor Content!</a>
<a><span>Anchor Content!</span></a>
<a dangerouslySetInnerHTML={{ __html: 'foo' }} />
<a children={children} />
<a title="Some content" />
<a aria-label="Some content" />
```

## Rule Options

The rule accepts an options object with the following properties:

- `components` — array of additional component names (besides the built-in
  `a`) that should be checked for accessible content.

Examples of **incorrect** code for this rule with `{ "components": ["Anchor"] }`:

```json
{ "anchor-has-content": ["error", { "components": ["Anchor"] }] }
```

```jsx
<Anchor />
```

Examples of **correct** code for this rule with `{ "components": ["Anchor"] }`:

```json
{ "anchor-has-content": ["error", { "components": ["Anchor"] }] }
```

```jsx
<Anchor>Anchor Content!</Anchor>
```

## Original Documentation

- [eslint-plugin-jsx-a11y/anchor-has-content](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-has-content.md)
