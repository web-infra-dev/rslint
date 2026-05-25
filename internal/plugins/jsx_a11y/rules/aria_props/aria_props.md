# aria-props

Enforce that every `aria-*` attribute on a JSX element is a recognized ARIA
state or property as defined by [aria-query](https://github.com/A11yance/aria-query).
Catches typos (`aria-labeledby` for `aria-labelledby`) and made-up names
(`aria-onclick`, `aria-tabindex`).

## Rule Details

The rule fires on a JSX attribute whose name literally begins with `aria-`
(lowercase, case-sensitive — `ARIA-HIDDEN` and `Aria-Hidden` are ignored)
but is not present in `aria-query`'s `ariaPropsMap`.

When the offending name is within Damerau-Levenshtein distance 2 of one or
more canonical attributes (after upper-casing both sides), the diagnostic
appends up to two suggestions in the form
`Did you mean to use <suggestion1>,<suggestion2>?`.

`JSXSpreadAttribute` is not visited — the spread payload is not inspected,
matching upstream's listener.

Examples of **incorrect** code for this rule:

```jsx
<div aria-labeledby="foo" />
<div aria-skldjfaria-klajsd="foo" />
<div aria-="foo" />
<div aria-tabindex="0" />
<div aria-onclick={handler} />
```

Examples of **correct** code for this rule:

```jsx
<div />
<div aria-labelledby="foo" />
<div abcARIAdef="true" />
<div fooaria-hidden="true" />
<Bar baz />
<input type="text" aria-errormessage="foo" />
```

## Resources

- [WAI-ARIA 1.2 — States and Properties](https://www.w3.org/TR/wai-aria-1.2/#state_prop_def)
- [MDN — ARIA attributes](https://developer.mozilla.org/en-US/docs/Web/Accessibility/ARIA/Attributes)
- [aria-query — `ariaPropsMap`](https://github.com/A11yance/aria-query/blob/HEAD/src/ariaPropsMap.js)

## Original Documentation

- [eslint-plugin-jsx-a11y/aria-props](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/aria-props.md)
