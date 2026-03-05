# no-extraneous-class

## Rule Details

Disallows classes used as namespaces or that serve no purpose beyond wrapping static members, constructors, or being empty. In JavaScript and TypeScript, classes that contain only static members, only a constructor, or no members at all can typically be replaced with standalone functions, plain objects, or modules. This rule reports on classes that do not benefit from the class structure.

Examples of **incorrect** code for this rule:

```typescript
class Empty {}

class OnlyConstructor {
  constructor() {
    doSomething();
  }
}

class StaticOnly {
  static utility() {}
  static helper() {}
}
```

Examples of **correct** code for this rule:

```typescript
class MyClass {
  value: string;
  constructor(value: string) {
    this.value = value;
  }
  greet() {
    return this.value;
  }
}

// Use module-level functions instead
export function utility() {}
export function helper() {}
```

## Original Documentation

- [typescript-eslint no-extraneous-class](https://typescript-eslint.io/rules/no-extraneous-class)
