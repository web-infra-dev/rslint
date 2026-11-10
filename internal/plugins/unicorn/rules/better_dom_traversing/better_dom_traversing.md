# better-dom-traversing

## Rule Details

Prefer `.firstChild` and `.firstElementChild` when reading the first child of a
DOM node. The rule also reports later `.children` indexes, repeated
`.parentElement` access, and chains of static `.querySelector()` calls. It
ignores writes to child collections and component `props.children` reads.

## Examples

### Incorrect

```javascript
element.childNodes[0];
element.children[0];
element.children[2];
element.parentElement.parentElement;
element.querySelector('a').querySelector('b');
```

### Correct

```javascript
element.firstChild;
element.firstElementChild;
element.querySelector(':scope selector');
element.closest('selector');
element.querySelector(':scope a b');
```

## Suggestions

The rule offers editor suggestions for first-child access and for chains of
static `querySelector()` calls with compatible selectors. First-child properties
return `null` when empty, while an indexed access returns `undefined`. A
combined selector can find matches outside the first element selected by the
original chain.

## Options

This rule has no options.

## Differences from upstream

In compact code, rslint adds spaces when a suggestion removes brackets or
parentheses next to a keyword. For example, the suggestion for
`element.children[0]in nodes` produces `element.firstElementChild in nodes`.
The space keeps the `in` expression valid.

## Original Documentation

- [eslint-plugin-unicorn: better-dom-traversing](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/better-dom-traversing.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/better-dom-traversing.js)
