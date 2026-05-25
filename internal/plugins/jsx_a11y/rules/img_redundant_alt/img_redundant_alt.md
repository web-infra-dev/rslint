# img-redundant-alt

Enforce that `<img>` alt attributes do not contain the words `image`,
`picture`, or `photo`. Screen readers already announce `img` tags as images,
so describing them as such in the alt text is redundant.

## Rule Details

This rule reports an `<img>` element whose `alt` value (a static string
literal) contains any of the words `image`, `picture`, or `photo` as a
whitespace-delimited token. Matching is case-insensitive.

For non-ASCII alt values (text without any ASCII printable characters), the
rule falls back to substring matching against the lowercased word list. This
keeps the rule effective for scripts that don't use space-delimited words
(e.g. `<img alt="イメージ" />` with `words: ["イメージ"]`).

Examples of **incorrect** code for this rule:

```jsx
<img src="cat.jpg" alt="Photo of a cat" />
<img src="dog.jpg" alt="Picture of a dog" />
<img src="bird.jpg" alt="Image of a bird" />
```

Examples of **correct** code for this rule:

```jsx
<img src="cat.jpg" alt="A black cat sitting on a windowsill" />
<img src="dog.jpg" alt="A dog catching a frisbee" />
<img src="bird.jpg" alt="" aria-hidden />
<img src="logo.png" alt={imageAlt} />
```

## Rule Options

### `components`

Type: `string[]`. Default: `[]`.

Additional JSX element names to validate alongside `<img>`. Use this for
wrapper components that render an `<img>` internally.

Examples of **incorrect** code with `{ "components": ["Image"] }`:

```json
{ "jsx-a11y/img-redundant-alt": ["error", { "components": ["Image"] }] }
```

```jsx
<Image alt="Picture of a cat" />
```

Examples of **correct** code with `{ "components": ["Image"] }`:

```json
{ "jsx-a11y/img-redundant-alt": ["error", { "components": ["Image"] }] }
```

```jsx
<Image alt="A cat sitting on a windowsill" />
```

### `words`

Type: `string[]`. Default: `[]`.

Additional words to flag as redundant. Appended to the built-in list
(`image`, `photo`, `picture`). Useful for non-English projects.

Examples of **incorrect** code with `{ "words": ["Bild", "Foto"] }`:

```json
{ "jsx-a11y/img-redundant-alt": ["error", { "words": ["Bild", "Foto"] }] }
```

```jsx
<img alt="Foto of a cat" />
```

Examples of **correct** code with `{ "words": ["Bild", "Foto"] }`:

```json
{ "jsx-a11y/img-redundant-alt": ["error", { "words": ["Bild", "Foto"] }] }
```

```jsx
<img alt="A cat sitting on a windowsill" />
```

## Original Documentation

- [eslint-plugin-jsx-a11y/img-redundant-alt](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/main/docs/rules/img-redundant-alt.md)
