# rstest/no-untyped-mock-factory

## Rule Details

Require a type argument on module mock calls, or an explicit return type on an inline factory. This lets TypeScript check the mock against the intended module shape when exports change. Enable this rule for TypeScript files. The rule checks for a written annotation, so even `any` satisfies it; it does not verify that the annotation describes the actual module.

The rule checks `rs.mock`, `rs.doMock`, `rs.mockRequire` and `rs.doMockRequire`, and the same methods on `rstest`. Calls to `mock` and `doMock` with `import('./module')` already infer the module type and are exempt. Calls without a factory and options objects such as `{ spy: true }` or `{ mock: true }` are exempt too. See [Mock modules](https://rstest.rs/api/runtime-api/rstest/mock-modules) for the module mocking API.

Calls must stand alone as statements and have exactly two arguments without spreads. Parentheses and TypeScript expression wrappers are accepted. Rstest transforms calls spelled `rs` or `rstest`, including imports from `@rstest/core` or `rstack/test`, CommonJS bindings, globals and local declarations with those names. Renamed receivers, namespace members, `import.meta.rstest.rs`, computed methods and optional chains are not checked because Rstest does not transform those mock calls.

Inline functions are always checked. For `doMock` and `doMockRequire`, an unchanged local function variable is also checked when its initializer necessarily runs before the call; type information can identify other callable arguments at that point. A declaration in a conditional branch, a declaration after the call, a reassigned binding, or another value that cannot be distinguished from an options object is left alone.

Because `mock` and `mockRequire` are hoisted, their named factories must be available in the same hoisted phase. The rule recognizes top-level function declarations and top-level bindings initialized by `rs.hoisted` or `rstest.hoisted`. Ordinary variables, imports, parameters and nested function declarations are left alone because the lifted call cannot read their runtime values. A return annotation on a separately declared factory does not exempt a reported call; write a type argument on the mock call in that case.

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
