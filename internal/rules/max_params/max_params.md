# max-params

## Rule Details

This rule enforces a maximum number of parameters allowed in function
definitions.

Examples of **incorrect** code for this rule with the default `{ "max": 3 }`
option:

```javascript
function foo1(bar, baz, qux, qxx) {
  doSomething();
}

let foo2 = (bar, baz, qux, qxx) => {
  doSomething();
};
```

Examples of **correct** code for this rule with the default `{ "max": 3 }`
option:

```javascript
function foo1(bar, baz, qux) {
  doSomething();
}

let foo2 = (bar, baz, qux) => {
  doSomething();
};
```

## Options

This rule accepts a number (the maximum allowed) or an object with the
following properties:

- `max` (default `3`): the maximum number of parameters allowed.
- `countThis` (default `"except-void"`): TypeScript-only handling for `this`
  declarations. Use `"always"` to count `this`, `"never"` to ignore it, or
  `"except-void"` to ignore only `this: void`.
- `maximum`: deprecated alias for `max`. When both keys are present and
  `maximum` is truthy, `maximum` wins (matching ESLint's
  `option.maximum || option.max` coercion).
- `countVoidThis`: deprecated alias for `countThis`. `true` maps to
  `"always"` and `false` maps to `"except-void"`.

### `max`

Examples of **incorrect** code for this rule with `{ "max": 2 }`:

```json
{ "max-params": ["error", { "max": 2 }] }
```

```javascript
function foo(bar, baz, qux) {
  doSomething();
}
```

Examples of **correct** code for this rule with `{ "max": 2 }`:

```json
{ "max-params": ["error", { "max": 2 }] }
```

```javascript
function foo(bar, baz) {
  doSomething();
}
```

### `countThis`

Examples of **correct** TypeScript code for this rule with
`{ "max": 2, "countThis": "never" }`:

```json
{ "max-params": ["error", { "max": 2, "countThis": "never" }] }
```

```typescript
function hasThis(this: unknown[], first: string, second: string) {
  doSomething();
}
```

Examples of **incorrect** TypeScript code for this rule with
`{ "max": 2, "countThis": "always" }`:

```json
{ "max-params": ["error", { "max": 2, "countThis": "always" }] }
```

```typescript
function hasThis(this: void, first: string, second: string) {
  doSomething();
}
```

## Original Documentation

- [https://eslint.org/docs/latest/rules/max-params](https://eslint.org/docs/latest/rules/max-params)
