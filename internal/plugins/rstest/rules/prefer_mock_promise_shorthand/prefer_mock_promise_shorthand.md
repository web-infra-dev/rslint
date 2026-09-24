# prefer-mock-promise-shorthand

## Rule Details

This rule requires [`mockResolvedValue()`](https://rstest.rs/api/runtime-api/rstest/mock-instance#mockresolvedvalue) and [`mockRejectedValue()`](https://rstest.rs/api/runtime-api/rstest/mock-instance#mockrejectedvalue) when a mock is configured to return `Promise.resolve()` or `Promise.reject()`, either through `mockReturnValue()` or through a `mockImplementation()` callback that does nothing but return the promise. The `Once` variants map to `mockResolvedValueOnce()` and `mockRejectedValueOnce()`. The shorthand builds a new promise on every call. A rejected promise passed to `mockReturnValue()` is built when the mock is configured, so a mock that ends up never being called leaves a rejection nobody handles, which Rstest reports as an error even when every test passes.

The rule matches on the method name alone, so it applies to any receiver: a mock from `rs.fn()`, a spy from `rs.spyOn()`, a value passed through `rs.mocked()`, a mock held in a variable, and the same calls reached through the `rstest` namespace, `import.meta.rstest` or a renamed import. Dot, string-literal and template-literal accessors are all recognized, and so is optional chaining. The promise has to come from the global `Promise`, in any of the same accessor spellings, and may be wrapped in a type assertion. A `Promise` declared in the file is some other object and is ignored, as are `Promise.all()` and a promise with further calls chained onto it. The rule does not require type information; when it is available, the autofix uses it as described below.

A `mockImplementation()` callback evaluates the promise's value on every call, while the shorthand evaluates it once, when the mock is configured. The callback is therefore left alone unless that move is equivalent. That excludes a callback taking parameters, a generator, a callback declaring type parameters, and a block body whose first statement is not a `return`. It also excludes a value that writes to something, reads a binding declared with `let` or `var`, spreads its arguments, or contains an `await`, and a `function` callback whose value reads `this`, `arguments` or `new.target`. An `async` callback is otherwise treated like any other, since an async function returning a promise settles the way that promise does. None of these exclusions apply to `mockReturnValue()`, whose argument is already evaluated once.

## Incorrect

```ts
rs.spyOn(api, 'fetchUser').mockReturnValue(Promise.resolve({ id: 1 }));

rs.spyOn(fs.promises, 'readFile').mockImplementation(() =>
  Promise.reject(new Error('file not found')),
);

const save = rs.fn().mockImplementationOnce(async () => Promise.resolve(true));
```

## Correct

```ts
rs.spyOn(api, 'fetchUser').mockResolvedValue({ id: 1 });

rs.spyOn(fs.promises, 'readFile').mockRejectedValue(new Error('file not found'));

const save = rs.fn().mockResolvedValueOnce(true);

// Still an implementation: `attempt` changes between calls.
let attempt = 0;
rs.spyOn(retry, 'run').mockImplementation(() => Promise.resolve(attempt));
```

## Autofix

The autofix renames the method and replaces the promise, or the callback returning it, with the value the promise was built from, or with `undefined` when it was built without one, written `void 0` where a local binding named `undefined` would be read instead. It keeps the accessor's dot, quote or template-literal style, any optional chaining, and any further arguments the call was given. Redundant parentheses around the value are dropped, except around a comma expression, where they decide which operand the promise settles with.

The rule reports without fixing when the promise is given more than one argument, when the mock call has type arguments, which describe the promise rather than its value, and when a type assertion wraps a resolved promise, since dropping it changes the type the value is checked against. With type information, it also reports without fixing when the value passed to `Promise.resolve()` may itself be a promise: `mockResolvedValue()` settles with it all the same, but its parameter is typed as the settled value, so the rewritten call would not type-check. It also reports without fixing when the rewrite would delete a comment, a function declared after the `return`, or the name of a named function expression that the value refers to.
