# no-alias-methods

## Rule Details

Requires canonical matcher names so assertions use one consistent vocabulary. For example, it replaces `toBeCalled` with `toHaveBeenCalled` and `toThrowError` with `toThrow`.

The rule checks normal dot access and static bracket access. Computed matcher names are ignored because their value is only known at runtime.

## Incorrect

```ts
expect(handler).toBeCalled();
expect(handler).lastCalledWith('ready');
```

## Correct

```ts
expect(handler).toHaveBeenCalled();
expect(handler).toHaveBeenLastCalledWith('ready');
```

## Autofix

Replaces static dot and bracket matcher aliases with canonical names while preserving the original access style.
