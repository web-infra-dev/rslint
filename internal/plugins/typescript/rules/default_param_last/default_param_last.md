# default-param-last

## Rule Details

Enforce default parameters to be last. This is the TypeScript-enhanced version of the ESLint `default-param-last` rule. It also handles TypeScript-specific optional parameters (those with a `?` modifier), enforcing that they come after required parameters.

Putting default and optional parameters last makes function calls clearer, since callers do not need to pass `undefined` to skip optional arguments.

Examples of **incorrect** code for this rule:

```typescript
function foo(a = 1, b: number) {}

function bar(a?: string, b: number) {}

class MyClass {
  method(a = 0, b: string) {}
}
```

Examples of **correct** code for this rule:

```typescript
function foo(a: number, b = 1) {}

function bar(a: number, b?: string) {}

function baz(a: number, b = 1, ...rest: number[]) {}
```

## Original Documentation

- [typescript-eslint default-param-last](https://typescript-eslint.io/rules/default-param-last)
