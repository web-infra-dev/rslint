# naming-convention

## Rule Details

Enforces naming conventions for everything across a codebase.

This rule allows you to enforce conventions for any identifier, using granular selectors to create a fine-grained style guide.

Examples of **incorrect** code for this rule (with default config):

```typescript
const my_variable = 1;
function my_function() {}
class my_class {}
```

Examples of **correct** code for this rule (with default config):

```typescript
const myVariable = 1;
function myFunction() {}
class MyClass {}
```

## Options

This rule accepts an array of objects, where each object describes a naming convention to enforce. Each object can have the following properties:

- `selector` - The selector to apply the convention to
- `format` - The format(s) that the identifier must match
- `leadingUnderscore` - Whether leading underscores are allowed/required/forbidden
- `trailingUnderscore` - Whether trailing underscores are allowed/required/forbidden
- `prefix` - Required prefix(es)
- `suffix` - Required suffix(es)
- `custom` - Custom regex to match against
- `filter` - Regex filter to limit which identifiers are checked
- `modifiers` - Required modifiers to match
- `types` - Required types to match

## Original Documentation

https://typescript-eslint.io/rules/naming-convention
