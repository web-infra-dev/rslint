# no-native

Require declaring or importing `Promise` before using it.

## Rule Details

This rule helps projects that use a Promise library such as Bluebird or target an environment without native promises. Each file must provide its own `Promise` binding instead of relying on a global implementation.

Examples of **incorrect** code for this rule:

```javascript
const x = Promise.resolve('bad');
new Promise((resolve) => resolve());
```

Examples of **correct** code for this rule:

```javascript
const Promise = require('bluebird');
const x = Promise.resolve('good');
```

```javascript
import Promise from 'bluebird';
const x = Promise.resolve('good');
```

Parameters and other local declarations named `Promise` also satisfy the rule. Accessing a property such as `window.Promise` does not reference the bare `Promise` binding and is allowed.

Configuring `Promise` in `languageOptions.globals`, adding a `/* global Promise */` comment, or selecting an ECMAScript version with native promises does not satisfy this rule.

## Options

This rule has no options.

## Original Documentation

- [eslint-plugin-promise: no-native](https://github.com/eslint-community/eslint-plugin-promise/blob/v7.3.0/docs/rules/no-native.md)
- [Source code](https://github.com/eslint-community/eslint-plugin-promise/blob/v7.3.0/rules/no-native.js)
