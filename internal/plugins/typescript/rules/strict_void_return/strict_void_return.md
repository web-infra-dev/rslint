# strict-void-return

## Rule Details

Disallow passing a value-returning function in a position accepting a void
function. TypeScript permits a function returning a value to be used where a
`void`-returning function is expected — callbacks can return any value and it
is silently discarded — but it hides several common mistakes: forgotten
`await`s on promise-returning callbacks, generators or async functions misused
as fire-and-forget event handlers, and accidental dead values from arrow
shorthands. This rule reports any value-returning function used in a context
that expects a function whose return type is `void`. It checks function
arguments, JSX attribute values, array elements, assignments, variable
initializers, object properties, class members (against extended base classes
and implemented interfaces), and `return` statements.

Examples of **incorrect** code for this rule:

```typescript
const getNothing: () => void = () => 2137;

declare function takesCallback(cb: () => void): void;
takesCallback(async () => {
  const response = await fetch('https://api.example.com/');
});

takesCallback(function* () {
  yield 'Hello';
});

['Alice', 'Bob'].forEach(name => `Hello, ${name}!`);

class Foo {
  cb() {
    console.log('foo');
  }
}
class Bar extends Foo {
  cb() {
    return 'bar';
  }
}

interface Foo {
  cb(): void;
}
class Bar implements Foo {
  cb() {
    return 'cb';
  }
}
```

Examples of **correct** code for this rule:

```typescript
const getNothing: () => void = () => {};

declare function takesCallback(cb: () => void): void;
takesCallback(() => {
  void (async () => {
    const response = await fetch('https://api.example.com/');
  })();
});

takesCallback(() => {
  function* gen() {
    yield 'Hello';
  }
  for (const _ of gen());
});

['Alice', 'Bob'].forEach(name => console.log(`Hello, ${name}!`));

class Foo {
  cb() {
    console.log('foo');
  }
}
class Bar extends Foo {
  cb() {
    super.cb();
    console.log('bar');
  }
}

interface Foo {
  cb(): void;
}
class Bar implements Foo {
  cb() {
    console.log('cb');
  }
}
```

## Options

### `allowReturnAny`

**Type:** `boolean` — **Default:** `false`

When `false` (default), a function returning `any` is treated the same as any
other non-void return — for example, `fn(() => JSON.parse('{}'))` is reported.
When `true`, functions returning `any` are accepted in void positions. Useful
for codebases where untyped values flow through callbacks intentionally.

Examples of **correct** code with `{ "allowReturnAny": true }`:

```json
{ "@typescript-eslint/strict-void-return": ["error", { "allowReturnAny": true }] }
```

```typescript
declare function fn(cb: () => void): void;
fn(() => JSON.parse('{}'));
fn(() => {
  return someUntypedApi();
});
```

## Original Documentation

- [typescript-eslint strict-void-return](https://typescript-eslint.io/rules/strict-void-return)
