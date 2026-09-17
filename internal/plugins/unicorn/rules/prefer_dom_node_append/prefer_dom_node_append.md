# prefer-dom-node-append

## Rule Details

Prefer `Element#append()` over `Node#appendChild()`.

An automatic fix is offered only for expression statements. Other uses are reported without a fix because `appendChild` returns the appended child, whereas `append` returns undefined. Computed names, optional calls, and values that cannot be DOM nodes are excluded.

Examples of **incorrect** code:

```javascript
element.appendChild(child);
```

Examples of **correct** code:

```javascript
element.append(child);
```

## Options

This rule has no options. It is enabled at `error` severity in `unicornPlugin.configs.recommended`.

## Original Documentation

- [eslint-plugin-unicorn: prefer-dom-node-append](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/docs/rules/prefer-dom-node-append.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/rules/prefer-dom-node-append.js)
