# no-dupe-class-members

## Rule Details

Disallow duplicate names in class members. If there are declarations of the same name in class members, the last declaration silently overwrites the earlier ones, which can cause unexpected behavior.

Examples of **incorrect** code for this rule:

```javascript
class A {
  bar() {}
  bar() {}
}

class B {
  bar() {}
  get bar() {}
}

class C {
  bar;
  bar;
}

class D {
  bar;
  bar() {}
}

class E {
  static bar() {}
  static bar() {}
}
```

Examples of **correct** code for this rule:

```javascript
class A {
  bar() {}
  qux() {}
}

class B {
  get bar() {}
  set bar(value) {}
}

class C {
  bar;
  qux;
}

class D {
  bar;
  qux() {}
}

class E {
  static bar() {}
  bar() {}
}
```

TypeScript method overload signatures are allowed:

```typescript
class A {
  foo(value: string): void;
  foo(value: number): void;
  foo(value: string | number) {}
}
```

## Original Documentation

https://eslint.org/docs/latest/rules/no-dupe-class-members
