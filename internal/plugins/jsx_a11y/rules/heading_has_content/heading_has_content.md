# heading-has-content

## Rule Details

This rule enforces that heading elements (`h1`–`h6`, and any configured
custom heading components) have accessible content — text or descendant
elements that screen readers can announce. An empty heading with no
accessible content is reported.

A heading is considered to have accessible content when at least one of
the following holds:

- It contains non-empty text or a non-hidden child element.
- It contains a JSX expression whose value is anything other than the
  bare identifier `undefined`.
- It declares a `dangerouslySetInnerHTML` or `children` prop.

A heading that is itself hidden from screen readers via `aria-hidden`
is exempt — it is not announced.

Examples of **incorrect** code for this rule:

```jsx
<h1 />
<h1><Bar aria-hidden /></h1>
<h1>{undefined}</h1>
<h1><input type="hidden" /></h1>
```

Examples of **correct** code for this rule:

```jsx
<h1>Heading Content!</h1>
<h2><Bar /></h2>
<h3>{foo}</h3>
<h4 dangerouslySetInnerHTML={{ __html: 'foo' }} />
<h5 children={children} />
<h6 aria-hidden />
```

## Rule Options

The rule accepts an options object with the following properties:

- `components` — array of additional component names (besides the
  built-in `h1`–`h6`) that should be checked for accessible content.

Examples of **incorrect** code for this rule with `{ "components": ["Heading"] }`:

```json
{ "heading-has-content": ["error", { "components": ["Heading"] }] }
```

```jsx
<Heading />
```

Examples of **correct** code for this rule with `{ "components": ["Heading"] }`:

```json
{ "heading-has-content": ["error", { "components": ["Heading"] }] }
```

```jsx
<Heading>Heading Content!</Heading>
```

## Original Documentation

- [eslint-plugin-jsx-a11y/heading-has-content](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/heading-has-content.md)
