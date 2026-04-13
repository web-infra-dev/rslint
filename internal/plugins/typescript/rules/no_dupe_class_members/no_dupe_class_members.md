# no-dupe-class-members

## Rule Details

Disallow duplicate class members.

If there are declarations of the same name in class members, the last declaration overwrites other declarations silently. It can cause unexpected behaviors.

This rule extends the base ESLint `no-dupe-class-members` rule to support TypeScript method overload signatures, which should not be flagged as duplicates.

Examples of **incorrect** code for this rule:

```typescript
class A {
  foo() {}
  foo() {}
}

class B {
  foo;
  foo() {}
}

class C {
  static bar() {}
  static bar() {}
}
```

Examples of **correct** code for this rule:

```typescript
class A {
  foo() {}
  bar() {}
}

class B {
  get foo() {}
  set foo(value) {}
}

class C {
  static foo() {}
  foo() {}
}

// TypeScript method overloads are allowed
class D {
  foo(a: string): string;
  foo(a: number): number;
  foo(a: any): any {}
}
```

## Original Documentation

- [ESLint `no-dupe-class-members`](https://eslint.org/docs/latest/rules/no-dupe-class-members)
- [TypeScript-ESLint `no-dupe-class-members`](https://typescript-eslint.io/rules/no-dupe-class-members/)
