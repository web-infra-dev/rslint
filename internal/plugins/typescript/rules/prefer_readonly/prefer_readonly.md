# prefer-readonly

## Rule Details

Require private members to be marked as `readonly` if they're never modified outside of the constructor.

Member variables with the `private` modifier or `#` private fields are only accessible within their declaring class. If that member is never reassigned after initialization (either at declaration or in the constructor), it should be marked as `readonly` to communicate intent and prevent accidental mutation.

Examples of **incorrect** code for this rule:

```typescript
class Foo {
  private neverModified = 'unchanged';
}

class Bar {
  #neverModified = 'unchanged';
}

class Baz {
  private neverModified = 'unchanged';

  public constructor() {
    this.neverModified = 'reassigned in constructor only';
  }
}
```

Examples of **correct** code for this rule:

```typescript
class Foo {
  private readonly neverModified = 'unchanged';
}

class Bar {
  readonly #neverModified = 'unchanged';
}

class Baz {
  private modifiedLater = 'unchanged';

  public mutate() {
    this.modifiedLater = 'changed outside constructor';
  }
}
```

## Options

### `onlyInlineLambdas`

When set to `true`, only checks members that are assigned an arrow function expression. This can be useful when a project wants to enforce `readonly` only for function-like members.

```json
{
  "@typescript-eslint/prefer-readonly": ["warn", { "onlyInlineLambdas": true }]
}
```

## Original Documentation

- [typescript-eslint prefer-readonly](https://typescript-eslint.io/rules/prefer-readonly)
