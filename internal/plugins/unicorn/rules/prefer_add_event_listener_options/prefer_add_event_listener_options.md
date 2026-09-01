# prefer-add-event-listener-options

## Rule Details

Prefer the options-object form of `.addEventListener()` over the legacy boolean
`useCapture` argument.

The object form makes the capture behavior explicit and leaves room for other
listener options such as `passive`, `once`, and `signal`.

Examples of **incorrect** code for this rule:

```javascript
element.addEventListener('click', handleClick, true);
element.addEventListener('scroll', handleScroll, false);
```

Examples of **correct** code for this rule:

```javascript
element.addEventListener('click', handleClick, { capture: true });
element.addEventListener('scroll', handleScroll, { capture: false });
element.addEventListener('scroll', handleScroll, {
  passive: true,
  capture: true,
});
```

## Original Documentation

- [eslint-plugin-unicorn: prefer-add-event-listener-options](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/docs/rules/prefer-add-event-listener-options.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/rules/prefer-add-event-listener-options.js)
