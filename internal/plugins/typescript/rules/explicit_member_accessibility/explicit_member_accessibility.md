# explicit-member-accessibility

## Rule Details

Require explicit accessibility modifiers on class members, with configurable
overrides for specific member types. This mirrors
`@typescript-eslint/explicit-member-accessibility`.

Examples of **incorrect** code for this rule:

```ts
class Example {
  value: number;
  getX() {
    return this.value;
  }
}
```

Examples of **correct** code for this rule:

```ts
class Example {
  public value: number;
  public getX() {
    return this.value;
  }
}
```

## Original Documentation

https://typescript-eslint.io/rules/explicit-member-accessibility
