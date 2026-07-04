# prefer-array-flat-map

## Rule Details

Prefer `Array#flatMap()` over chaining `Array#map()` and `Array#flat()` when the flat depth is omitted or exactly `1`.

Examples of **incorrect** code for this rule:

```javascript
const foo = bar.map(element => unicorn(element)).flat();
const foo = bar.map(element => unicorn(element)).flat(1);
```

Examples of **correct** code for this rule:

```javascript
const foo = bar.flatMap(element => unicorn(element));
const foo = bar.map(element => unicorn(element)).flat(2);
const foo = React.Children.map(children, fn).flat();
```

## Original Documentation

- [eslint-plugin-unicorn: prefer-array-flat-map](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/main/docs/rules/prefer-array-flat-map.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/main/rules/prefer-array-flat-map.js)
