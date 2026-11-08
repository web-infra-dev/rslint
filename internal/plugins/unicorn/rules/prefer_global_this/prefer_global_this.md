# prefer-global-this

## Rule Details

Prefer `globalThis` over `window`, `self`, and `global` to keep code portable
between browsers, Node.js, and other JavaScript runtimes.

This rule reports references to those global objects, including references not
declared in `languageOptions.globals`. Local bindings are respected.

Examples of **incorrect** code:

```javascript
window;
window.navigator;
window.location;
global.foo;
self.postMessage('Hello');
const { foo } = window;
window.addEventListener('click', handler);
```

Examples of **correct** code:

```javascript
globalThis;
globalThis.navigator;
globalThis.location;
globalThis.foo;
globalThis.postMessage('Hello');
const { foo } = globalThis;
globalThis.addEventListener('click', handler);
```

Window-specific APIs and events remain allowed, as do computed accesses on
`window`, runtime `typeof` checks, and TypeScript `typeof` type queries:

```typescript
window.innerWidth;
window.open('https://example.com');
window.addEventListener('resize', handler);
window.addEventListener(`load`, handler);
window[foo];
'open' in window;
typeof window === 'undefined';
type WindowType = typeof window;
```

## Fixes

The rule automatically replaces read references. Shorthand object properties
retain their keys: `{ window }` becomes `{ window: globalThis }`. Writes such as
`window = value` are reported without a fix because replacing the assignment
target could overwrite `globalThis`.

## Options

This rule has no options.

## Differences from upstream

Unlike upstream, the following references are reported without an automatic fix:

- A local value named `globalThis` would change what the replacement refers to.
  For example, `function f(globalThis) { return window; }` must keep returning the
  global object, rather than the parameter. Rename the local binding to enable
  the fix.
- A paired JSX tag in a JavaScript file, such as
  `<window.Widget></window.Widget>`, must keep its opening and closing names in
  sync. Change both names manually. Self-closing tags and TypeScript TSX tags
  retain automatic fixes.

## Original Documentation

- [eslint-plugin-unicorn: prefer-global-this](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/prefer-global-this.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/prefer-global-this.js)
