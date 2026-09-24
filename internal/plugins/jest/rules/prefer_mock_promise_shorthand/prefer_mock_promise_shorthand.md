# prefer-mock-promise-shorthand

## Rule Details

Prefer `mockResolvedValue()` and `mockRejectedValue()` when a mock is
configured to return `Promise.resolve()` or `Promise.reject()`, either through
`mockReturnValue()` or through a `mockImplementation()` callback that does
nothing but return the promise. The `Once` variants map to
`mockResolvedValueOnce()` and `mockRejectedValueOnce()`. The shorthand builds a
new promise on every call; a rejected promise passed to `mockReturnValue()` is
built when the mock is configured, and becomes an unhandled rejection if the
mock is never called.

The rule matches on the method name alone, so it applies to any receiver: a mock
from `jest.fn()`, a spy from `jest.spyOn()`, or a mock held in a variable. Dot,
string-literal and template-literal accessors are all recognized, and so is
optional chaining. The promise has to come from the global `Promise`, in any of
the same accessor spellings, and may be wrapped in a type assertion; a
`Promise` declared in the file is ignored. It is fixable.

A `mockImplementation()` callback evaluates the promise's value on every call,
while the shorthand evaluates it once, when the mock is configured, so the
callback is left alone unless that move is equivalent. That excludes:

- a callback taking parameters, including a TypeScript `this` parameter;
- a generator, and a callback declaring type parameters;
- a `function` callback whose value reads `this`, `arguments` or `new.target`,
  because all three are bound by the call. An arrow callback has none of its own
  and is treated normally;
- a block body whose first statement is not a `return`;
- a value that writes to something — an assignment, an increment or a
  `delete` — or reads a binding declared with `let` or `var`, including a read
  inside a function the value hands back;
- a value that spreads its arguments, `Promise.resolve(...values)`, because the
  spread iterates its operand on every call;
- a value containing an `await`, which belongs to the `async` callback it is
  written in. An `async` callback is otherwise treated like any other.

None of these exclusions apply to `mockReturnValue()`, whose argument is already
evaluated once.

Examples of **incorrect** code for this rule:

```typescript
jest.fn().mockImplementation(() => Promise.resolve(123));
jest
  .spyOn(fs.promises, 'readFile')
  .mockReturnValue(Promise.reject(new Error('oh noes!')));

myFunction
  .mockReturnValueOnce(Promise.resolve(42))
  .mockImplementationOnce(() => Promise.resolve(42))
  .mockReturnValue(Promise.reject(new Error('too many calls!')));
```

Examples of **correct** code for this rule:

```typescript
jest.fn().mockResolvedValue(123);
jest.spyOn(fs.promises, 'readFile').mockRejectedValue(new Error('oh noes!'));

myFunction
  .mockResolvedValueOnce(42)
  .mockResolvedValueOnce(42)
  .mockRejectedValue(new Error('too many calls!'));

let attempt = 0;
jest.spyOn(retry, 'run').mockImplementation(() => Promise.resolve(attempt));
```

The autofix renames the method and replaces the promise, or the callback
returning it, with the value the promise was built from, or with `undefined`
when it was built without one. It keeps the accessor's dot, quote or
template-literal style, any optional chaining, and any further arguments the
call was given. Redundant parentheses around the value are dropped, except
around a comma expression, where they decide which operand the promise settles
with. No fix is offered when the promise is given more than one argument, when
the mock call has type arguments, when a type assertion wraps a resolved
promise, or when the rewrite would delete a comment or a declaration the value
relies on.

## Original Documentation

- [eslint-plugin-jest: prefer-mock-promise-shorthand](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/docs/rules/prefer-mock-promise-shorthand.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/src/rules/prefer-mock-promise-shorthand.ts)
