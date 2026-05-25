# explicit-member-accessibility

## Rule Details

This rule reports class members and parameter properties whose accessibility
declarations don't match the configured policy. With the default `explicit`
mode, every class member and parameter property must be declared `public`,
`private`, or `protected`. Switching the policy to `no-public` flips the rule:
any redundant `public` modifier is reported (and removed by the autofix). Set
the policy to `off` to disable the check for that member kind.

Examples of **incorrect** code for this rule:

```typescript
class Animal {
  constructor(name: string) {}
  getName(): string {
    return this.name;
  }
  get legs(): number {
    return 4;
  }
}
```

Examples of **correct** code for this rule:

```typescript
class Animal {
  public constructor(public readonly name: string) {}
  public getName(): string {
    return this.name;
  }
  public get legs(): number {
    return 4;
  }
}
```

## Options

This rule accepts an options object with the following properties:

- `accessibility`: top-level policy applied to every member kind unless
  overridden. One of `'explicit'` (default), `'no-public'`, `'off'`.
- `ignoredMethodNames`: list of method names to skip entirely. Method name
  matching uses the same name normalization as the diagnostic message
  (identifier text, `#name` for private fields, the literal value for string
  / numeric literal keys).
- `overrides`: per-kind overrides. Each entry overrides `accessibility` for
  that member kind:
  - `accessors` — getters and setters.
  - `constructors` — constructors.
  - `methods` — regular methods (not getters/setters/constructors).
  - `parameterProperties` — `public`/`private`/`protected`/`readonly`
    parameters of a constructor.
  - `properties` — class fields, including auto-accessor (`accessor x`) and
    abstract properties.

### `accessibility: 'no-public'`

```json
{ "@typescript-eslint/explicit-member-accessibility": ["error", { "accessibility": "no-public" }] }
```

Examples of **incorrect** code with this option:

```typescript
class Animal {
  public name: string;
  public getName(): string {
    return this.name;
  }
}
```

Examples of **correct** code with this option:

```typescript
class Animal {
  name: string;
  getName(): string {
    return this.name;
  }
}
```

### `overrides`

Examples of **correct** code with mixed overrides:

```json
{
  "@typescript-eslint/explicit-member-accessibility": [
    "error",
    {
      "accessibility": "explicit",
      "overrides": { "constructors": "no-public", "accessors": "off" }
    }
  ]
}
```

```typescript
class Animal {
  constructor(private readonly name: string) {}
  public bark(): void {}
  get legs(): number {
    return 4;
  }
}
```

### `ignoredMethodNames`

Examples of **correct** code with `{ "ignoredMethodNames": ["getX"] }`:

```json
{ "@typescript-eslint/explicit-member-accessibility": ["error", { "ignoredMethodNames": ["getX"] }] }
```

```typescript
class Test {
  getX() {
    return 1;
  }
}
```

## Original Documentation

- [typescript-eslint explicit-member-accessibility](https://typescript-eslint.io/rules/explicit-member-accessibility)
