# prefer-query-selector

## Rule Details

Prefer `.querySelector()` and `.querySelectorAll()` over older DOM query methods.

This rule reports `.getElementById()`, `.getElementsByClassName()`,
`.getElementsByTagName()`, and `.getElementsByName()` calls to keep DOM queries
consistent. It automatically converts safe literal arguments to CSS selectors.
Queries on objects other than `document`, `window.document`, or
`globalThis.document` receive a `:scope` prefix. A following `[0]`, `.at(0)`, or
`.item(0)` becomes a single-element query when it is safe to remove the access.

Examples of **incorrect** code:

```javascript
document.getElementById('foo');
document.getElementsByClassName('foo bar');
document.getElementsByTagName('main');
document.getElementsByName('email');
document.getElementsByTagName('form')[0];
element.getElementsByClassName('foo');
```

Examples of **correct** code:

```javascript
document.querySelector('#foo');
document.querySelectorAll('.foo.bar');
document.querySelectorAll('main');
document.querySelectorAll('[name="email"]');
document.querySelector('form');
element.querySelectorAll(':scope .foo');
```

Values requiring CSS escaping, dynamic arguments, and empty strings are reported
without an automatic fix. Compose and escape dynamic selectors explicitly:

```javascript
const singleClassName = 'foo';

// Incorrect
document.getElementsByClassName(singleClassName);

// Correct
document.querySelectorAll(`.${CSS.escape(singleClassName)}`);
```

## Options

### `allowWithVariables`

Type: `boolean`\
Default: `false`

When `true`, allows non-literal arguments to `.getElementById()` and
`.getElementsByClassName()`. String literals, binary expressions, and template
literals containing non-whitespace text are still reported. Calls to
`.getElementsByTagName()` and `.getElementsByName()` are always checked.

```javascript
// With { allowWithVariables: true }
document.getElementById(someId);
document.getElementsByClassName(someClass);

// Still reported
document.getElementById('foo');
document.getElementsByClassName('foo');
document.getElementsByTagName(someTag);
document.getElementsByName(someName);
```

## Original Documentation

- [eslint-plugin-unicorn: prefer-query-selector](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/prefer-query-selector.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/prefer-query-selector.js)
