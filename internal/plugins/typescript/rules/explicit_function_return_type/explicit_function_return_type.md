# explicit-function-return-type

## Rule Details

Require explicit return types on functions and class methods.

Examples of **incorrect** code for this rule:

```typescript
function foo() {
  return;
}

const bar = () => 'value';

class Test {
  method() {
    return 1;
  }
}
```

Examples of **correct** code for this rule:

```typescript
function foo(): void {
  return;
}

const bar = (): string => 'value';

class Test {
  method(): number {
    return 1;
  }
}
```

## Options

This rule accepts an options object with the following fields (defaults shown):

- `allowConciseArrowFunctionExpressionsStartingWithVoid`: `false`
  - Allows `() => void expr` style arrows without explicit return types.
- `allowDirectConstAssertionInArrowFunctions`: `true`
  - Allows direct `as const` assertions in arrow bodies without explicit return types.
- `allowedNames`: `[]`
  - Function names that are exempt from this rule.
- `allowExpressions`: `false`
  - Allows function expressions in expression positions without explicit return types.
- `allowFunctionsWithoutTypeParameters`: `false`
  - Skips functions that have no type parameters.
- `allowHigherOrderFunctions`: `true`
  - Allows functions that immediately return another function expression.
- `allowIIFEs`: `false`
  - Allows immediately-invoked function expressions.
- `allowTypedFunctionExpressions`: `true`
  - Allows function expressions that are already typed by context (e.g. assignments, parameters).

## Original Documentation

https://typescript-eslint.io/rules/explicit-function-return-type
