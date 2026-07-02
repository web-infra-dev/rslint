# no-conditional-expect

## Rule Details

Disallow calling `expect` conditionally. Jest only marks a test as failed when it throws; if an assertion runs inside a branch that is skipped, the test can pass without exercising the assertion at all. Conditionals also make tests harder to read and reason about. While `expect.assertions` and `expect.hasAssertions` can help catch silent skips, combining them with conditionals usually adds even more complexity.

This rule reports `expect` calls that sit inside conditional control flow, including:

- `if` / `else` branches
- `switch` cases
- `try` / `catch` handlers (including empty `catch` blocks that contain `expect`)
- Short-circuit expressions (`&&`, `||`) and ternary expressions where `expect` is on a branch that may not run
- Promise `.catch()` callbacks whose parameter is treated as an error handler (for example `.catch(error => expect(error)...)`)

The same checks apply when the conditional `expect` lives in a helper function passed as the test callback. Conditionals that run **before** an unconditional `expect`, or that only affect the value passed **into** `expect`, are allowed.

Examples of **incorrect** code for this rule:

```javascript
it('foo', () => {
  doTest && expect(1).toBe(2);
});

it('bar', () => {
  if (!skipTest) {
    expect(1).toEqual(2);
  }
});

it('baz', () => {
  something ? expect(something).toHaveBeenCalled() : noop();
});

it('qux', () => {
  switch (something) {
    case 'value':
      expect(something).toHaveBeenCalled();
      break;
    default:
      break;
  }
});

it('handles errors', () => {
  try {
    processRequest(request);
  } catch (err) {
    expect(err).toMatchObject({ code: 'MODULE_NOT_FOUND' });
  }
});

it('throws an error', async () => {
  await foo().catch(error => expect(error).toBeInstanceOf(Error));
});
```

Examples of **correct** code for this rule:

```javascript
it('foo', () => {
  expect(!value).toBe(false);
});

it('foo', () => {
  process.env.FAIL && setNum(1);

  expect(num).toBe(2);
});

function getValue() {
  if (process.env.FAIL) {
    return 1;
  }

  return 2;
}

it('foo', () => {
  expect(getValue()).toBe(2);
});

it('validates the request', () => {
  try {
    processRequest(request);
  } catch {
    // ignore errors
  } finally {
    expect(validRequest).toHaveBeenCalledWith(request);
  }
});

it('throws an error', async () => {
  await expect(foo).rejects.toThrow(Error);
});
```

### Testing thrown errors without violating this rule

A common pattern is asserting properties on a caught error when `toThrow` only checks the message. A `try` / `catch` with `expect` in the `catch` block looks fine but still passes if nothing is thrown:

```javascript
it('includes the status code in the error', async () => {
  try {
    await makeRequest(url);
  } catch (error) {
    expect(error).toHaveProperty('statusCode', 404);
  }
});
```

Prefer a small wrapper that always returns a value (or throws a sentinel when no error occurred), then assert unconditionally:

```javascript
class NoErrorThrownError extends Error {}

const getError = async call => {
  try {
    await call();
    throw new NoErrorThrownError();
  } catch (error) {
    return error;
  }
};

it('includes the status code in the error', async () => {
  const error = await getError(() => makeRequest(url));

  expect(error).not.toBeInstanceOf(NoErrorThrownError);
  expect(error).toHaveProperty('statusCode', 404);
});
```

## Original Documentation

- [jest/no-conditional-expect](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/no-conditional-expect.md)
