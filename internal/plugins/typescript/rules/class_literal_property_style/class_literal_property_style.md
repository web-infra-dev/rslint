# class-literal-property-style

## Rule Details

Enforce that literals on classes are exposed in a consistent style, either as readonly fields or as getter methods. When a class has a property that always returns a literal value, there are two ways to expose it: a `readonly` field or a `get` accessor. This rule enforces one style for consistency.

The rule supports two modes: `"fields"` (default) prefers `readonly` fields, and `"getters"` prefers getter methods.

Examples of **incorrect** code for this rule (with default `"fields"` option):

```typescript
class Foo {
  get name() {
    return 'foo';
  }
}

class Bar {
  get count() {
    return 42;
  }
}
```

Examples of **correct** code for this rule (with default `"fields"` option):

```typescript
class Foo {
  readonly name = 'foo';
}

class Bar {
  readonly count = 42;
}
```

## Original Documentation

- [typescript-eslint class-literal-property-style](https://typescript-eslint.io/rules/class-literal-property-style)
