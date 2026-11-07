# prefer-regexp-test

## Rule details

Prefer `RegExp#test()` over `String#match()`, `String#search()`, and
`RegExp#exec()` when only checking whether a pattern matches.

Examples of **incorrect** code for this rule:

```javascript
if (string.match(/unicorn/)) {}
if (/unicorn/.exec(string)) {}
if (string.match(/unicorn/).length > 0) {}
if (string.search(/unicorn/) !== -1) {}
```

Examples of **correct** code for this rule:

```javascript
if (/unicorn/.test(string)) {}
```

The rule recognizes boolean contexts and comparisons of `search()` results with
`-1` or `0`. Calls with known incompatible receivers or arguments are ignored.

Static regular expressions without the `g` flag can be fixed automatically.
For `search()`, the `y` flag also prevents automatic fixes. Unknown expressions
and expressions with these flags receive suggestions. Positive `.length` checks
receive suggestions because replacing them can change how a failed match is
handled. Comments that would be removed prevent the fix or suggestion.

## Options

This rule has no options.

## Differences from upstream

Expressions inside Vue and Svelte templates are not checked. Expressions in
JavaScript, JSX, and TypeScript files are checked.

String intersections that resolve to `string`, such as `string & unknown`, are
treated as strings. The following call is left unchanged instead of receiving
a suggestion to use `.test()`:

```typescript
function f(pattern: string & unknown) {
  if (text.match(pattern)) {}
}
```

## Original documentation

- [eslint-plugin-unicorn: prefer-regexp-test](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/prefer-regexp-test.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/prefer-regexp-test.js)
