# prefer-includes

## Rule Details

Disallow `indexOf(...) !== -1` / `indexOf(...) === -1` style checks when
`includes(...)` expresses intent more clearly.

Examples of **incorrect** code for this rule:

```ts
if (arr.indexOf(value) !== -1) {
  doSomething();
}
```

Examples of **correct** code for this rule:

```ts
if (arr.includes(value)) {
  doSomething();
}
```

## Original Documentation

- [typescript-eslint: prefer-includes](https://typescript-eslint.io/rules/prefer-includes)
- [Source code](https://github.com/typescript-eslint/typescript-eslint/blob/v8.67.0/packages/eslint-plugin/src/rules/prefer-includes.ts)
