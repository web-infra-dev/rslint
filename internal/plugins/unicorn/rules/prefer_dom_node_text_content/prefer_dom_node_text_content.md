# prefer-dom-node-text-content

## Rule Details

Prefer `.textContent` over `.innerText` on DOM nodes. Reading `.innerText` depends
on rendered layout and can trigger reflow. Use `.innerText` when you need the
text as rendered. [MDN `Node.textContent`](https://developer.mozilla.org/en-US/docs/Web/API/Node/textContent#differences_from_innertext)

This rule checks dot access, including optional chains, and identifier keys in
object destructuring. It skips values known to have a non-DOM type. Its editor
suggestion preserves the local name in shorthand destructuring.

Examples of **incorrect** code

```javascript
const text = foo.innerText;
const { innerText } = foo;
foo.innerText = '🦄';
```

Examples of **correct** code

```javascript
const text = foo.textContent;
const { textContent } = foo;
foo.textContent = '🦄';
```

## Options

This rule has no options.

## Differences from upstream

With project type information, rslint can distinguish browser DOM types from
application types that share their names. For example, if `./library` exports
its own `Node` class, eslint-plugin-unicorn reports the access below and rslint
skips it.

```typescript
import { Node as Model } from './library';

new Model().innerText;
```

## Original Documentation

- [eslint-plugin-unicorn prefer-dom-node-text-content](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/prefer-dom-node-text-content.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/prefer-dom-node-text-content.js)
