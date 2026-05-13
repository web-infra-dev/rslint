# html-has-lang

## Rule Details

This rule enforces that every `<html>` element carries a truthy `lang` prop.
Screen readers and assistive technology rely on the document language to
pronounce content correctly; an `<html>` element without a usable `lang`
value violates [WCAG 3.1.1](https://www.w3.org/WAI/WCAG21/Understanding/language-of-page).

The element name is resolved through the standard jsx-a11y settings —
`components` and `polymorphicPropName` — so a custom component mapped to
`html` (e.g. `<HTMLTop />` with `settings: { 'jsx-a11y': { components: { HTMLTop: 'html' } } }`)
is checked the same way as a literal `<html>` tag.

The `lang` prop is considered present when its value extracts to a JS-truthy
value via `getPropValue`. Notably, `<html lang />` is treated as truthy
(boolean `true`); `<html lang={undefined} />` is falsy and reported.

This rule is largely superseded by the [`lang` rule](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/lang.md),
which additionally validates that the value is a recognized BCP-47 language tag.

Examples of **incorrect** code for this rule:

```jsx
<html />
<html {...props} />
<html lang={undefined} />
```

Examples of **correct** code for this rule:

```jsx
<html lang="en" />
<html lang="en-US" />
<html lang={language} />
<html lang />
```

## Accessibility guidelines

- [WCAG 3.1.1](https://www.w3.org/WAI/WCAG21/Understanding/language-of-page)

### Resources

- [axe-core, html-has-lang](https://dequeuniversity.com/rules/axe/3.2/html-has-lang)
- [axe-core, html-lang-valid](https://dequeuniversity.com/rules/axe/3.2/html-lang-valid)

## Original Documentation

- [eslint-plugin-jsx-a11y/html-has-lang](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/html-has-lang.md)
