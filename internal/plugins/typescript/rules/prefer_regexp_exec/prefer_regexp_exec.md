# prefer-regexp-exec

## Rule Details

Prefer `RegExp#exec` over `String#match` when a non-global regex match is used.

Examples of **incorrect** code for this rule:

```ts
const value = 'foo';
value.match(/foo/);
value.match('foo');
```

Examples of **correct** code:

```ts
const value = 'foo';
/foo/.exec(value);
value.match(/foo/g);
```

## Original Documentation

- https://typescript-eslint.io/rules/prefer-regexp-exec
