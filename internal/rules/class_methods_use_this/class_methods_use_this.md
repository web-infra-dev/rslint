# class-methods-use-this

## Rule Details

This rule requires class instance methods to use `this` or `super`. Methods
that do not depend on an instance can often be ordinary functions or static
methods instead.

Examples of **incorrect** code for this rule:

```javascript
class Formatter {
  format(value) {
    return String(value);
  }
}
```

Examples of **correct** code for this rule:

```javascript
class Formatter {
  format(value) {
    return this.prefix + value;
  }

  static createDefault() {
    return new Formatter();
  }
}
```

## Options

### `exceptMethods`

Use `exceptMethods` to exempt methods whose instance shape is required by an
external API.

```json
{ "class-methods-use-this": ["error", { "exceptMethods": ["render", "#reset"] }] }
```

```javascript
class Component {
  render() {
    return null;
  }

  #reset() {}
}
```

### `enforceForClassFields`

`enforceForClassFields` defaults to `true`. Set it to `false` to ignore arrow
functions and function expressions used as instance field initializers.

```json
{ "class-methods-use-this": ["error", { "enforceForClassFields": false }] }
```

```javascript
class Component {
  render = () => null;
}
```

### `ignoreOverrideMethods`

Use `ignoreOverrideMethods` to ignore TypeScript members marked with
`override`.

```json
{ "class-methods-use-this": ["error", { "ignoreOverrideMethods": true }] }
```

```typescript
class Derived extends Base {
  override render() {
    return null;
  }
}
```

### `ignoreClassesWithImplements`

Use `"all"` to ignore every member of a TypeScript class with an `implements`
clause, or `"public-fields"` to ignore only its public members.

```json
{ "class-methods-use-this": ["error", { "ignoreClassesWithImplements": "all" }] }
```

```typescript
interface Service {
  run(): void;
}

class Worker implements Service {
  run() {}
}
```

## Original Documentation

- [ESLint: class-methods-use-this](https://eslint.org/docs/latest/rules/class-methods-use-this)
- [Source code](https://github.com/eslint/eslint/blob/v10.9.0/lib/rules/class-methods-use-this.js)
