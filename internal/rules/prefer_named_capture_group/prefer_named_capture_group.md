# prefer-named-capture-group

## Rule Details

This rule enforces using named capture groups instead of numbered capture
groups in regular expressions. Named capture groups make a regular
expression's intent easier to read, and make the captured values easier to
retrieve (`match.groups.name` instead of a numbered index that shifts if the
pattern changes).

This rule checks regex literals as well as string patterns passed to the
`RegExp` constructor (including through `globalThis`/`window`/`self`/`global`)
when the pattern is statically determinable.

Examples of **incorrect** code for this rule:

```javascript
const foo = /(ba[rz])/;
const bar = new RegExp("(ba[rz])");
const baz = RegExp("(ba[rz])");

foo.exec("bar")[1]; // Retrieve the group result.
```

Examples of **correct** code for this rule:

```javascript
const foo = /(?<id>ba[rz])/;
const bar = new RegExp("(?<id>ba[rz])");
const baz = RegExp("(?<id>ba[rz])");
const xyz = /xyz(?:zy|abc)/;

foo.exec("bar").groups.id; // Retrieve the group result.
```

## Options

This rule has no configurable options.

## Differences from ESLint

- rslint doesn't follow a local alias assigned from the global `RegExp`
  constructor — `const R = RegExp; new R("(a)")` is not reported, while
  ESLint reports it.

## When Not To Use It

If you are targeting ECMAScript 2017 or older environments, you should
disable this rule, because named capture groups are only supported in
ECMAScript 2018 and newer environments.

## Original Documentation

- [ESLint: prefer-named-capture-group](https://eslint.org/docs/latest/rules/prefer-named-capture-group)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/prefer-named-capture-group.js)
