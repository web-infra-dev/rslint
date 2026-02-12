# related-getter-setter-pairs

## Rule Details

Enforce that `get()` types are assignable to their equivalent `set()` types. A getter and setter for the same property should have compatible types. The getter's return type should be assignable to the setter's parameter type; otherwise it creates a confusing API where writing a value and then reading it back produces a different type.

Examples of **incorrect** code for this rule:

```typescript
interface Foo {
  get value(): string;
  set value(newValue: number);
}

class Bar {
  get prop(): string {
    return this._prop;
  }
  set prop(newValue: number) {
    this._prop = String(newValue);
  }
}
```

Examples of **correct** code for this rule:

```typescript
interface Foo {
  get value(): string;
  set value(newValue: string);
}

class Bar {
  get prop(): string {
    return this._prop;
  }
  set prop(newValue: string) {
    this._prop = newValue;
  }
}
```

## Original Documentation

- [typescript-eslint related-getter-setter-pairs](https://typescript-eslint.io/rules/related-getter-setter-pairs)
