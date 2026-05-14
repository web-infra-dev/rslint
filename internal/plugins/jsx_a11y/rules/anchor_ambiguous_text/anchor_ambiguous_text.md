# anchor-ambiguous-text

## Rule Details

This rule reports anchor (`<a>`) elements whose accessible text is exactly
one of a configurable set of ambiguous phrases — `click here`, `here`,
`link`, `a link`, and `learn more` by default. Screen readers announce
anchors as links and rely on the link text for context, so vague text like
"click here" provides no information about the destination.

The accessible text of an anchor is computed as follows:

- If the element has an `aria-label`, its value is used in place of the
  children.
- If the element resolves to `<img>` (after the `jsx-a11y.components` and
  `jsx-a11y.polymorphicPropName` settings are applied) and has an `alt`
  attribute, the alt text is used.
- If the element is hidden from screen readers (`aria-hidden="true"`,
  `aria-hidden`, or `<input type="hidden">`), it contributes the empty
  string.
- Otherwise, child text and nested elements are joined with single spaces
  and recursed into.

The resulting text is compared against the ambiguous wordlist after
normalization: leading and trailing whitespace are trimmed, runs of
whitespace are collapsed to a single space, sentence-ending punctuation
(`,`, `.`, `?`, `¿`, `!`, `‽`, `¡`, `;`, `:`) is stripped, and the text is
lower-cased. Matching is exact against the normalized form.

Examples of **incorrect** code for this rule:

```jsx
<a>here</a>
<a>HERE</a>
<a>click here</a>
<a>learn more.</a>
<a>a link</a>
<a> a link </a>
<a><span>click</span> here</a>
<a><span aria-hidden="true">more text</span>learn more</a>
<a><img alt="click here" /></a>
<a aria-label="click here">something</a>
```

Examples of **correct** code for this rule:

```jsx
<a>read this tutorial</a>
<a aria-label="tutorial on using eslint-plugin-jsx-a11y">click here</a>
<a><img alt="documentation" /></a>
```

## Rule Options

The rule accepts an options object with the following property:

- `words` — array of phrases to treat as ambiguous. When supplied, this
  list **replaces** the defaults (it is not merged), so include every
  phrase you want flagged. Useful for adding terms in other languages or
  for disabling the rule entirely via an empty array.

Examples of **incorrect** code for this rule with
`{ "words": ["a disallowed word"] }`:

```json
{ "anchor-ambiguous-text": ["error", { "words": ["a disallowed word"] }] }
```

```jsx
<a>a disallowed word</a>
```

Examples of **correct** code for this rule with
`{ "words": ["disabling the defaults"] }` (the default wordlist is
replaced, so "click here" no longer triggers):

```json
{ "anchor-ambiguous-text": ["error", { "words": ["disabling the defaults"] }] }
```

```jsx
<a>click here</a>
```

## Accessibility guidelines

Ensure anchor tags describe the content of the link rather than simply
describing them as a link.

Compare:

```jsx
<p><a href="#">click here</a> to read a tutorial by Foo Bar</p>
```

with the more concise and accessible:

```jsx
<p>read <a href="#">a tutorial by Foo Bar</a></p>
```

### Resources

- [WebAIM, Hyperlinks](https://webaim.org/techniques/hypertext/)
- [Deque University, Link Checklist — "Avoid 'link' (or similar) in the link text"](https://dequeuniversity.com/checklists/web/links)

## Original Documentation

- [eslint-plugin-jsx-a11y/anchor-ambiguous-text](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-ambiguous-text.md)
