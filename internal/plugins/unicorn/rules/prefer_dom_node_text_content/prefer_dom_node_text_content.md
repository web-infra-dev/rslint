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

This rule has no options. Enable it explicitly as `unicorn/prefer-dom-node-text-content` in your rules configuration.

## Differences from upstream

In JavaScript linted without project type information, rslint reports
`make().innerText` with the annotation below. eslint-plugin-unicorn uses the
callback's documented return type and skips the access.

```javascript
/** @param {function(): {innerText: string}} make */
function read(make) {
  return make().innerText;
}
```

Use `@param {() => {innerText: string}} make` to let rslint recognize the
callback's return type in this case.

## Original Documentation

- [eslint-plugin-unicorn prefer-dom-node-text-content](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/prefer-dom-node-text-content.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/prefer-dom-node-text-content.js)
