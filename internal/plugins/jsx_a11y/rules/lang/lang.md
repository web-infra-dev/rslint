# lang

## Rule Details

This rule enforces that every `<html>` element's `lang` prop, when set to a
literal value, is a well-formed and registered
[BCP-47](https://datatracker.ietf.org/doc/html/rfc5646) language tag.

The `lang` attribute on the root `<html>` element tells assistive technology
which natural language to use when announcing the page; an invalid value
makes the assistive technology fall back to the user's default voice and
violates
[WCAG 3.1.1](https://www.w3.org/WAI/WCAG21/Understanding/language-of-page).

The element name is resolved through the standard jsx-a11y settings —
`components` and `polymorphicPropName` — so a custom component mapped to
`html` (e.g. `<HTMLRoot />` with
`settings: { 'jsx-a11y': { components: { HTMLRoot: 'html' } } }`) is checked
the same way as a literal `<html>` tag.

Examples of **incorrect** code for this rule:

```jsx
<html lang="foo" />
<html lang="zz-LL" />
<html lang={undefined} />
<html lang="en_US" />
```

Examples of **correct** code for this rule:

```jsx
<html lang="en" />
<html lang="en-US" />
<html lang="zh-Hans" />
<html lang="zh-Hant-HK" />
<html lang={locale} />
```

## Differences from ESLint

- A non-string literal `lang` value — `<html lang />`, `<html lang={true} />`,
  `<html lang="true" />`, `<html lang={1} />`, `<html lang={["en"]} />` — is
  reported as an invalid value. ESLint's `tags.check(value)` throws a
  `TypeError` on non-string input, which aborts the lint run on the affected
  file; rslint emits a normal diagnostic so the developer sees the problem
  without losing the rest of the file's reports.

## Accessibility guidelines

- [WCAG 3.1.1](https://www.w3.org/WAI/WCAG21/Understanding/language-of-page)

### Resources

- [BCP-47 (RFC 5646)](https://datatracker.ietf.org/doc/html/rfc5646)
- [IANA Language Subtag Registry](https://www.iana.org/assignments/language-subtag-registry)
- [axe-core, html-lang-valid](https://dequeuniversity.com/rules/axe/3.2/html-lang-valid)

## Original Documentation

- [eslint-plugin-jsx-a11y/lang](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/lang.md)
