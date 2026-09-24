# rstest/no-untyped-mock-factory

## Rule Details

Require a type argument on module mock calls, or an explicit return type on an inline factory. This lets TypeScript check the mock against the intended module shape when exports change. Enable this rule for TypeScript files. The rule checks for a written annotation, so even `any` satisfies it; it does not verify that the annotation describes the actual module.

The rule checks `rs.mock`, `rs.doMock`, `rs.mockRequire` and `rs.doMockRequire`, and the same methods on `rstest`. Calls to `mock` and `doMock` with `import('./module')` already infer the module type and are exempt. Calls without a factory and options objects such as `{ spy: true }` or `{ mock: true }` are exempt too. See [Mock modules](https://rstest.rs/api/runtime-api/rstest/mock-modules) for the module mocking API.

Calls must stand alone as statements and have exactly two arguments without spreads. Parentheses and TypeScript expression wrappers are accepted. Rstest transforms calls spelled `rs` or `rstest`, including imports from `@rstest/core` or `rstack/test`, CommonJS bindings, globals and local declarations with those names. Renamed receivers, namespace members, `import.meta.rstest.rs`, computed methods and optional chains are not checked because Rstest does not transform those mock calls.

Inline functions and directly declared, unchanged local functions can be checked without type information. With type information, other arguments known to be callable are checked too. An argument that cannot be distinguished from an options object is left alone. A return annotation on a separately declared function does not exempt the call; write a type argument on the mock call in that case.

## Incorrect

```ts
import { rs } from '@rstest/core';

rs.mock('./user-service', () => ({
  fetchUser: rs.fn().mockResolvedValue({ id: 1, name: 'Ada' }),
}));
```

## Correct

```ts
import { rs } from '@rstest/core';

rs.mock<typeof import('./user-service')>('./user-service', () => ({
  fetchUser: rs.fn().mockResolvedValue({ id: 1, name: 'Ada' }),
}));
```

You can also use `rs.mock(import('./user-service'), factory)` to infer the module type, or give an inline factory an explicit return type.

## Autofix

Adds `<typeof import('./module')>` using the first argument's quoted module name. Quotes, escapes, comments and parentheses are preserved. If the module name is not a quoted string, the rule reports the call without a fix.
