# parameter-properties

## Rule Details

Require or disallow parameter properties in class constructors.

TypeScript includes a shorthand for declaring and initializing class members from constructor parameters called parameter properties. This rule can be used to enforce consistent usage of this feature.

## Options

- `prefer` (`"class-property"` | `"parameter-property"`): Whether to prefer class properties or parameter properties. Default: `"class-property"`.
- `allow` (array of modifiers): Which parameter property modifiers to allow. Valid values: `"readonly"`, `"private"`, `"protected"`, `"public"`, `"private readonly"`, `"protected readonly"`, `"public readonly"`. Default: `[]`.

### `prefer: "class-property"` (default)

Examples of **incorrect** code:

```typescript
class Foo {
  constructor(readonly name: string) {}
}

class Bar {
  constructor(private age: number) {}
}
```

Examples of **correct** code:

```typescript
class Foo {
  constructor(name: string) {}
}
```

### `prefer: "parameter-property"`

Examples of **incorrect** code:

```typescript
class Foo {
  member: string;
  constructor(member: string) {
    this.member = member;
  }
}
```

Examples of **correct** code:

```typescript
class Foo {
  constructor(private member: string) {}
}
```

## Original Documentation

https://typescript-eslint.io/rules/parameter-properties
