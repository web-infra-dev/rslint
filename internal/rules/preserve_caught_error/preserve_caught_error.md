# preserve-caught-error

## Rule Details

When a `catch` block reacts to a failure by throwing a new, more descriptive error, the original error should travel along with it as the new error's `cause`. This rule requires every error constructed and thrown inside a `catch` block to receive the caught error as its `cause`, so the underlying failure stays visible in stack traces and error reports.

The rule checks the global error constructors — `Error`, `EvalError`, `RangeError`, `ReferenceError`, `SyntaxError`, `TypeError`, `URIError`, and `AggregateError` — and reports both a missing `cause` and a `cause` set to anything other than the caught error. Each report comes with a suggestion that attaches the caught error.

Examples of **incorrect** code for this rule:

```javascript
try {
  doSomething();
} catch (error) {
  throw new Error('Failed to perform error prone operations');
}

try {
  doSomething();
} catch (error) {
  throw new Error('Failed to perform error prone operations', {
    cause: error.message,
  });
}

try {
  doSomething();
} catch (error) {
  if (whatever) {
    const error = anotherError;
    throw new Error('Something went wrong', { cause: error });
  }
}

try {
  doSomething();
} catch ({ message }) {
  throw new Error(message);
}
```

Examples of **correct** code for this rule:

```javascript
try {
  doSomething();
} catch (error) {
  throw new Error('Failed to perform error prone operations', { cause: error });
}

try {
  doSomething();
} catch (error) {
  throw new Error('Failed to perform error prone operations', {
    cause: error,
    retryable: true,
  });
}

try {
  doSomething();
} catch (error) {
  console.error(error);
}

try {
  doSomething();
} catch (error) {
  foo = {
    bar() {
      throw new Error('Unrelated to the caught error');
    },
  };
}
```

## Options

### `requireCatchParameter`

By default a `catch` block may omit the parameter entirely, which discards the caught error before the rule can ask for it. Set `requireCatchParameter` to `true` to require the parameter so the caught error stays available.

Examples of **incorrect** code for this rule with `{ "requireCatchParameter": true }`:

```json
{ "preserve-caught-error": ["error", { "requireCatchParameter": true }] }
```

```javascript
try {
  doSomething();
} catch {
  throw new Error('Something went wrong');
}
```

Examples of **correct** code for this rule with `{ "requireCatchParameter": true }`:

```json
{ "preserve-caught-error": ["error", { "requireCatchParameter": true }] }
```

```javascript
try {
  doSomething();
} catch (error) {
  throw new Error('Something went wrong', { cause: error });
}
```

### `errorClassNames`

Custom error classes are checked when their name is listed in `errorClassNames`. A bare string means the class takes its error options as the second argument, matching the built-in error signature. To describe a different signature, use an object with `name` and the 1-based `argumentPosition` of the options argument.

The name is matched against the constructor identifier, including the property name of a namespaced constructor such as `new errors.AppError()`.

Examples of **incorrect** code for this rule with `{ "errorClassNames": ["AppError"] }`:

```json
{ "preserve-caught-error": ["error", { "errorClassNames": ["AppError"] }] }
```

```javascript
class AppError extends Error {}

try {
  doSomething();
} catch (error) {
  throw new AppError('Something went wrong');
}
```

Examples of **correct** code for this rule with `{ "errorClassNames": [{ "name": "AppError", "argumentPosition": 3 }] }`:

```json
{
  "preserve-caught-error": [
    "error",
    { "errorClassNames": [{ "name": "AppError", "argumentPosition": 3 }] }
  ]
}
```

```javascript
class AppError extends Error {}

try {
  doSomething();
} catch (error) {
  throw new AppError('Something went wrong', context, { cause: error });
}
```

## Differences from ESLint

- For a constructor written without an argument list, the suggestion adds the arguments after the whole expression — `new AppError<T>` becomes `new AppError<T>({ cause: error })` and `new (AppError)` becomes `new (AppError)({ cause: error })`. ESLint puts them directly after the callee name, which lands inside the type arguments or the parentheses.

## Original Documentation

[https://eslint.org/docs/latest/rules/preserve-caught-error](https://eslint.org/docs/latest/rules/preserve-caught-error)
