# consistent-date-clone

## Rule Details

Prefer passing an existing `Date` directly to the `Date` constructor when
cloning it. Since ECMAScript 2015, `new Date(date)` copies the original date's
time value without requiring an intermediate number.

Examples of **incorrect** code for this rule:

```javascript
const copy = new Date(date.getTime());
```

Examples of **correct** code for this rule:

```javascript
const copy = new Date(date);
```

The rule only reports the direct `new Date(date.getTime())` shape. Calls with
additional constructor arguments or component-wise date construction are left
unchanged:

```javascript
new Date(date.getTime(), extraArgument);

new Date(
  date.getFullYear(),
  date.getMonth(),
  date.getDate(),
  date.getHours(),
  date.getMinutes(),
  date.getSeconds(),
);
```

## Original Documentation

- [eslint-plugin-unicorn: consistent-date-clone](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/docs/rules/consistent-date-clone.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/rules/consistent-date-clone.js)
