# max-classes-per-file

## Rule Details

This rule enforces that each file may contain only a particular number of classes and no more.

Examples of **incorrect** code for this rule:

```javascript
class Foo {}
class Bar {}
```

Examples of **correct** code for this rule:

```javascript
class Foo {}
```

## Options

This rule may be configured with either an object or a number.

If the option is an object, it may contain one or both of:

- `ignoreExpressions`: a boolean option (defaulted to `false`) to ignore class expressions.
- `max`: a numeric option (defaulted to 1) to specify the maximum number of classes.

Examples of **correct** code for this rule with `{ "max-classes-per-file": ["error", 2] }`:

```json
{ "max-classes-per-file": ["error", 2] }
```

```javascript
class Foo {}
class Bar {}
```

Examples of **correct** code for this rule with `{ "max-classes-per-file": ["error", { "ignoreExpressions": true }] }`:

```json
{ "max-classes-per-file": ["error", { "ignoreExpressions": true }] }
```

```javascript
class VisitorFactory {
  forDescriptor(descriptor) {
    return class {
      visit(node) {
        return `Visiting ${descriptor}.`;
      }
    };
  }
}
```

## Original Documentation

- [ESLint: max-classes-per-file](https://eslint.org/docs/latest/rules/max-classes-per-file)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/max-classes-per-file.js)
