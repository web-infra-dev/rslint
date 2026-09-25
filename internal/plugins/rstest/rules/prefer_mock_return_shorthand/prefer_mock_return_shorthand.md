# prefer-mock-return-shorthand

## Rule Details

This rule requires [`mockReturnValue()`](https://rstest.rs/api/runtime-api/rstest/mock-instance#mockreturnvalue) when `mockImplementation()` is given a callback that does nothing but return a value, and `mockReturnValueOnce()` in the same situation for `mockImplementationOnce()`. Writing the value directly says what the mock produces without a function wrapped around it.

The rule matches on the method name alone, so it applies to any receiver: a mock from `rs.fn()`, a spy from `rs.spyOn()`, a value passed through `rs.mocked()`, a mock held in a variable, and the same calls reached through the `rstest` namespace, `import.meta.rstest` or a renamed import. Dot, string-literal and template-literal accessors are all recognized, and so is optional chaining. A bracketed key that is a variable rather than a literal is not, because the call site does not name a method. The rule does not require type information.

A callback is left alone unless replacing it with its value is equivalent on every call. That excludes a callback taking parameters, an `async` callback, a generator, a callback declaring type parameters, and a block body whose first statement is not a `return`. It also excludes a returned expression that writes to something — an assignment, an increment or a `delete` — since the write would move from every call to the single moment the mock is configured, and a returned expression that reads a binding declared with `let` or `var`, since the shorthand would freeze whichever value that binding held at that moment. A read inside a function the expression hands back counts too, because the returned closure observes the binding later.

A `function` callback that reads `this`, `arguments` or `new.target` is left alone as well. All three are bound by the call, and a mock can be called with `new`, so lifting them out of the callback makes them name whatever encloses the mock's configuration instead. An arrow callback has none of its own, so they already name the enclosing scope and the rule treats it normally.

A callback returning `Promise.reject()` is also left alone. The shorthand would build the rejected promise when the mock is configured rather than when it is called, so a mock that ends up never being called leaves a rejection nobody handles, which Rstest reports as an error even when every test passes. Use [`mockRejectedValue()`](https://rstest.rs/api/runtime-api/rstest/mock-instance#mockrejectedvalue), which builds the promise per call.

## Incorrect

```ts
const readConfig = rs.fn().mockImplementation(() => ({ retries: 3 }));

rs.spyOn(clock, 'now').mockImplementationOnce(() => 1_700_000_000);

rs.mocked(featureFlags.isEnabled).mockImplementation(function () {
  return true;
});
```

## Correct

```ts
const readConfig = rs.fn().mockReturnValue({ retries: 3 });

rs.spyOn(clock, 'now').mockReturnValueOnce(1_700_000_000);

rs.mocked(featureFlags.isEnabled).mockReturnValue(true);

// Still an implementation: it reads the arguments.
rs.fn().mockImplementation((a, b) => a + b);

// Still an implementation: `attempt` changes between calls.
let attempt = 0;
rs.spyOn(retry, 'count').mockImplementation(() => attempt);
```

## Autofix

The autofix renames the method and replaces the callback with the expression it returned. It keeps the accessor's dot, quote or template-literal style, any optional chaining, the call's type arguments, and any further arguments the call was given.

An object or array literal returned by the callback is a new value on every call, and becomes one value shared by every call. That is the point of the shorthand, but it means a caller that mutates what the mock returned now affects later calls.

Redundant parentheses around the returned expression are dropped, except around a comma expression, where they decide which operand the mock returns. The fix is withheld when the accessor cannot be rewritten — a private name, for instance.
