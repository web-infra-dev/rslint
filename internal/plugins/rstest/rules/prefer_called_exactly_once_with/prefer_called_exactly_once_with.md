# prefer-called-exactly-once-with

## Rule Details

Prefers one exact mock assertion over separate call-count and argument assertions. The combined matcher states the complete expectation in one place.

Chai-style `calledOnce` and `calledWith` assertions are supported too.

## Incorrect

```ts
expect(handler).toHaveBeenCalledOnce();
expect(handler).toHaveBeenCalledWith('ready');
```

## Correct

```ts
expect(handler).toHaveBeenCalledExactlyOnceWith('ready');
```

## Autofix

Merges the pair into the combined matcher, keeping the argument-side assertion and deleting the call-count one. The pair is reported without a fix when the subject cannot be evaluated once instead of twice, as in `expect(getMock())`, or when one of the two chains carries further assertions that deleting the statement would drop.
