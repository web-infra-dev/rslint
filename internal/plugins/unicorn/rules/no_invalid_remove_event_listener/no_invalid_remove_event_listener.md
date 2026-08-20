# no-invalid-remove-event-listener

## Rule Details

Prevent calling `EventTarget#removeEventListener()` with an inline function or
the result of an inline `.bind()` call. `removeEventListener()` must receive the
same function reference that was passed to `addEventListener()`.

Examples of **incorrect** code for this rule:

```javascript
window.removeEventListener("click", listener.bind(window));
window.removeEventListener("click", () => {});
window.removeEventListener("click", function () {});
```

Examples of **correct** code for this rule:

```javascript
const listener = () => {};
window.addEventListener("click", listener);
// ...
window.removeEventListener("click", listener);
```

## Original Documentation

- [eslint-plugin-unicorn: no-invalid-remove-event-listener](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/docs/rules/no-invalid-remove-event-listener.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/rules/no-invalid-remove-event-listener.js)
