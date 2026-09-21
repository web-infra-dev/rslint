# prefer-mock-return-shorthand

## Rule Details

Prefer `mockReturnValue()` over `mockImplementation()` when the callback does
nothing but return a value, and `mockReturnValueOnce()` over
`mockImplementationOnce()` in the same situation. Writing the value directly says
what the mock produces without a function wrapped around it.

The rule matches on the method name alone, so it applies to any receiver: a mock
from `jest.fn()`, a spy from `jest.spyOn()`, or a mock held in a variable. Dot,
string-literal and template-literal accessors are all recognized, and so is
optional chaining. A bracketed key that is a variable rather than a literal is
not, because the call site does not name a method. It is fixable.

A callback is left alone unless replacing it with its value is equivalent on
every call. That excludes:

- a callback taking parameters, including a TypeScript `this` parameter;
- an `async` callback, a generator, and a callback declaring type parameters;
- a `function` callback reading `this`, `arguments` or `new.target`, because
  all three are bound by the call and a mock can be called with `new`. An
  arrow callback has none of its own and is treated normally;
- a block body whose first statement is not a `return`;
- a returned expression that writes to something — an assignment, an increment
  or a `delete` — because the write would move from every call to the single
  moment the mock is configured;
- a returned expression that reads a binding declared with `let` or `var`,
  because the shorthand would freeze whichever value that binding held at that
  moment. A read inside a function the expression hands back counts too;
- a returned `Promise.reject()`, because the shorthand would build the rejected
  promise when the mock is configured rather than when it is called, leaving an
  unhandled rejection if the mock is never called. `mockRejectedValue()` builds
  the promise per call.

Examples of **incorrect** code for this rule:

```typescript
jest.fn().mockImplementation(() => ({ retries: 3 }));
jest.spyOn(clock, 'now').mockImplementationOnce(() => 1700000000);
jest.mocked(featureFlags.isEnabled).mockImplementation(function () {
  return true;
});
```

Examples of **correct** code for this rule:

```typescript
jest.fn().mockReturnValue({ retries: 3 });
jest.spyOn(clock, 'now').mockReturnValueOnce(1700000000);
jest.fn().mockImplementation((a, b) => a + b);

let attempt = 0;
jest.spyOn(retry, 'count').mockImplementation(() => attempt);
```

The autofix renames the method and replaces the callback with the expression it
returned, keeping the accessor's dot, quote or template-literal style, any
optional chaining, the call's type arguments, and any further arguments the call
was given. Redundant parentheses around the returned expression are dropped,
except around a comma expression, where they decide which operand the mock
returns.

## Original Documentation

- [eslint-plugin-jest: prefer-mock-return-shorthand](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.1/docs/rules/prefer-mock-return-shorthand.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.1/src/rules/prefer-mock-return-shorthand.ts)
