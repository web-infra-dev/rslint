# no-unsafe-declaration-merging

## Rule Details

Disallows unsafe declaration merging between a class and an interface. TypeScript allows a class and an interface that share the same name in the same scope to merge into a single declaration. Properties declared on the interface are added to the resulting type without forcing the class to actually initialize them, so accessing such a property on a class instance type-checks but throws `Cannot read properties of undefined` at runtime.

Examples of **incorrect** code for this rule:

```typescript
interface Foo {}
class Foo {}
```

```typescript
class Foo {}
interface Foo {}
```

```typescript
declare global {
  interface Foo {}
  class Foo {}
}
```

Examples of **correct** code for this rule:

```typescript
interface Foo {}
class Bar implements Foo {}
```

```typescript
namespace Foo {}
namespace Foo {}
```

```typescript
enum Foo {}
namespace Foo {}
```

```typescript
namespace Qux {}
function Qux() {}
```

```typescript
const Foo = class {};
```

```typescript
interface Foo {
  props: string;
}

function bar() {
  return class Foo {};
}
```

```typescript
declare global {
  interface Foo {}
}

class Foo {}
```

## Original Documentation

- [typescript-eslint no-unsafe-declaration-merging](https://typescript-eslint.io/rules/no-unsafe-declaration-merging)
