# class-methods-use-this

## Rule Details

Enforce that class methods utilize `this`.

If a class method does not use `this`, it can sometimes be made into a static function. If you do convert the method into a static function, instances of the class that call that particular method will have to be converted to a static call (i.e., `MyClass.callStaticMethod()`).

This rule extends the base ESLint `class-methods-use-this` rule with two TypeScript-specific options: `ignoreOverrideMethods` (skip members marked with the `override` modifier) and `ignoreClassesThatImplementAnInterface` (skip members of classes that `implements` an interface).

Examples of **incorrect** code for this rule:

```javascript
class A {
  foo() {
    console.log('Hello World');
  }
}
```

Examples of **correct** code for this rule:

```javascript
class A {
  foo() {
    this.bar = 'Hello World';
  }
}

class A {
  constructor() {
    // OK. constructor is exempt.
  }
}

class A {
  static foo() {
    // OK. static methods are exempt.
  }
}
```

## Options

### `enforceForClassFields`

**Type:** `boolean` — **Default:** `true`

Enforces that functions used as instance field initializers utilize `this`.

Examples of **incorrect** code with `{ "enforceForClassFields": true }` (default):

```json
{ "@typescript-eslint/class-methods-use-this": ["error", { "enforceForClassFields": true }] }
```

```javascript
class A {
  foo = () => {};
}
```

Examples of **correct** code with `{ "enforceForClassFields": false }`:

```json
{ "@typescript-eslint/class-methods-use-this": ["error", { "enforceForClassFields": false }] }
```

```javascript
class A {
  foo = () => {};
}
```

### `exceptMethods`

**Type:** `string[]` — **Default:** `[]`

Allows specified method names to be ignored by this rule. Private class members can be referenced via their `#`-prefixed name (`#foo`).

Examples of **correct** code with `{ "exceptMethods": ["foo", "#bar"] }`:

```json
{ "@typescript-eslint/class-methods-use-this": ["error", { "exceptMethods": ["foo", "#bar"] }] }
```

```javascript
class A {
  foo() {}
  #bar() {}
}
```

### `ignoreOverrideMethods`

**Type:** `boolean` — **Default:** `false`

Whether to ignore class members marked with the `override` modifier.

Examples of **correct** code with `{ "ignoreOverrideMethods": true }`:

```json
{ "@typescript-eslint/class-methods-use-this": ["error", { "ignoreOverrideMethods": true }] }
```

```typescript
class Base {
  method() {}
}

class Derived extends Base {
  override method() {}
}
```

### `ignoreClassesThatImplementAnInterface`

**Type:** `boolean | 'public-fields'` — **Default:** `false`

Whether to ignore class members that are defined within a class that `implements` an interface.

- `true` — ignore every member of any class that implements an interface.
- `'public-fields'` — only ignore public members (those without a `private` or `protected` modifier).

Examples of **correct** code with `{ "ignoreClassesThatImplementAnInterface": true }`:

```json
{ "@typescript-eslint/class-methods-use-this": ["error", { "ignoreClassesThatImplementAnInterface": true }] }
```

```typescript
class Foo implements Bar {
  method() {}
  property = () => {};
}
```

Examples of **correct** code with `{ "ignoreClassesThatImplementAnInterface": "public-fields" }`:

```json
{ "@typescript-eslint/class-methods-use-this": ["error", { "ignoreClassesThatImplementAnInterface": "public-fields" }] }
```

```typescript
class Foo implements Bar {
  method() {}
}
```

Examples of **incorrect** code with `{ "ignoreClassesThatImplementAnInterface": "public-fields" }` (`private`/`protected` members are still checked):

```typescript
class Foo implements Bar {
  private method() {}
  protected property = () => {};
}
```

## Original Documentation

- [typescript-eslint class-methods-use-this](https://typescript-eslint.io/rules/class-methods-use-this)
- [ESLint class-methods-use-this](https://eslint.org/docs/latest/rules/class-methods-use-this)
