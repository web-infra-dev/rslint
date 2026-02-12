# no-base-to-string

## Rule Details

Require `.toString()` and `.toLocaleString()` to only be called on objects which provide useful information when stringified. JavaScript calls `toString()` implicitly in string contexts such as template literals and string concatenation. The default `Object.prototype.toString()` returns `"[object Object]"`, which is rarely useful.

This rule uses type information to detect when a value will use the base `Object.prototype.toString()` and flags string concatenation, template literals, explicit `.toString()`/`.toLocaleString()` calls, `String()` calls, and `.join()` on arrays containing such types.

Examples of **incorrect** code for this rule:

```typescript
class MyClass {}
const obj = new MyClass();

`Value: ${obj}`;
'' + obj;
obj.toString();
[obj].join(',');
```

Examples of **correct** code for this rule:

```typescript
`Value: ${'str'}`;
'' + 42;

class MyClass {
  toString() {
    return 'MyClass';
  }
}
`Value: ${new MyClass()}`;
```

## Original Documentation

- [typescript-eslint no-base-to-string](https://typescript-eslint.io/rules/no-base-to-string)
