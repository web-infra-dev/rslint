# no-empty-function

## Rule Details

Disallow empty functions. Empty functions can hide incomplete refactors; add a
comment to intentionally empty function bodies, or configure an allowed function
kind.

Examples of **incorrect** code for this rule:

```javascript
function noop() {}

const handler = () => {};

class Service {
  connect() {}
}
```

Examples of **correct** code for this rule:

```javascript
function noop() {
  // intentionally empty
}

const handler = () => value;

class Service {
  connect() {
    start();
  }
}
```

TypeScript parameter-property constructors are also allowed:

```typescript
class Store {
  constructor(private readonly id: string) {}
}
```

## Options

- `allow`: An array of empty function kinds to permit. Default: `[]`.

Allowed values:

- `functions`
- `arrowFunctions`
- `generatorFunctions`
- `methods`
- `generatorMethods`
- `getters`
- `setters`
- `constructors`
- `asyncFunctions`
- `asyncMethods`
- `privateConstructors`
- `protectedConstructors`
- `decoratedFunctions`
- `overrideMethods`

Examples of **correct** code for this rule with `{ "allow": ["constructors"] }`:

```json
{ "no-empty-function": ["error", { "allow": ["constructors"] }] }
```

```javascript
class Service {
  constructor() {}
}
```

Examples of **correct** code for this rule with `{ "allow": ["decoratedFunctions"] }`:

```json
{ "no-empty-function": ["error", { "allow": ["decoratedFunctions"] }] }
```

```typescript
class Service {
  @bound
  handle() {}
}
```

## Original Documentation

- [ESLint: no-empty-function](https://eslint.org/docs/latest/rules/no-empty-function)
