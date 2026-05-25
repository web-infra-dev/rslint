# naming-convention

## Rule Details

Enforces naming conventions for everything across a codebase.

This rule allows you to enforce conventions for any identifier, using granular selectors to create a fine-grained style guide. It supports a wide variety of selectors, modifiers, formats, and custom patterns. Each selector can be configured independently, and more specific selectors take precedence over less specific ones.

Examples of **incorrect** code for this rule (with default config):

```typescript
const my_variable = 1;
function my_function() {}
class my_class {}
interface my_interface {}
type my_type = string;
enum my_enum {
  my_member,
}
```

Examples of **correct** code for this rule (with default config):

```typescript
const myVariable = 1;
function myFunction() {}
class MyClass {}
interface MyInterface {}
type MyType = string;
enum MyEnum {
  MyMember,
}
```

## Options

This rule accepts an array of objects, where each object describes a naming convention to enforce. Each object can have the following properties:

### `selector`

**(Required)** The selector(s) to apply the convention to. Can be a string or an array of strings.

**Individual selectors:** `variable`, `function`, `parameter`, `property`, `parameterProperty`, `accessor`, `enumMember`, `classMethod`, `objectLiteralMethod`, `typeMethod`, `classProperty`, `objectLiteralProperty`, `typeProperty`, `class`, `interface`, `typeAlias`, `enum`, `typeParameter`, `import`

**Group selectors:** `default` (matches all), `variableLike` (variable, function, parameter), `memberLike` (property, parameterProperty, enumMember, classMethod, objectLiteralMethod, typeMethod, classProperty, objectLiteralProperty, typeProperty, accessor), `typeLike` (class, interface, typeAlias, enum), `method` (classMethod, objectLiteralMethod, typeMethod), `objectLiteralMember` (objectLiteralProperty, objectLiteralMethod)

### `format`

The format(s) that the identifier must match. Set to `null` to skip format checking (useful for names that require quotes). Can be an array to allow multiple formats.

**Allowed values:** `camelCase`, `strictCamelCase`, `PascalCase`, `StrictPascalCase`, `snake_case`, `UPPER_CASE`

```json
{
  "@typescript-eslint/naming-convention": [
    "warn",
    { "selector": "variable", "format": ["camelCase", "UPPER_CASE"] }
  ]
}
```

### `leadingUnderscore` / `trailingUnderscore`

Controls whether leading/trailing underscores are allowed, required, or forbidden.

**Allowed values:** `forbid`, `require`, `requireDouble`, `allow`, `allowDouble`, `allowSingleOrDouble`

```json
{
  "@typescript-eslint/naming-convention": [
    "warn",
    {
      "selector": "variable",
      "format": ["camelCase"],
      "leadingUnderscore": "allow"
    }
  ]
}
```

### `prefix` / `suffix`

Requires identifiers to start/end with one of the given strings. The prefix/suffix is stripped before format checking.

```json
{
  "@typescript-eslint/naming-convention": [
    "warn",
    { "selector": "interface", "format": ["PascalCase"], "prefix": ["I"] }
  ]
}
```

### `custom`

A custom regex pattern that the identifier must match (or not match). Requires a `regex` string and a `match` boolean.

```json
{
  "@typescript-eslint/naming-convention": [
    "warn",
    {
      "selector": "variable",
      "format": ["camelCase"],
      "custom": { "regex": "^I[A-Z]", "match": false }
    }
  ]
}
```

### `filter`

A regex filter to limit which identifiers are checked by this selector. Identifiers matching the filter are skipped (`match: false`) or exclusively checked (`match: true`).

```json
{
  "@typescript-eslint/naming-convention": [
    "warn",
    {
      "selector": "property",
      "format": null,
      "filter": {
        "regex": "^(Property-Name-One|Property-Name-Two)$",
        "match": true
      }
    }
  ]
}
```

### `modifiers`

Limits the selector to only match identifiers with the specified modifiers. All specified modifiers must be present for the selector to match.

**Allowed values:** `const`, `readonly`, `static`, `public`, `protected`, `private`, `#private`, `abstract`, `destructured`, `global`, `exported`, `unused`, `requiresQuotes`, `override`, `async`, `default`, `namespace`

```json
{
  "@typescript-eslint/naming-convention": [
    "warn",
    { "selector": "variable", "modifiers": ["const"], "format": ["UPPER_CASE"] }
  ]
}
```

### `types`

Limits the selector to only match identifiers whose type matches. Only available for: `variable`, `parameter`, `classProperty`, `objectLiteralProperty`, `typeProperty`, `accessor`, `property`, `parameterProperty`.

**Allowed values:** `boolean`, `string`, `number`, `function`, `array`

```json
{
  "@typescript-eslint/naming-convention": [
    "warn",
    {
      "selector": "variable",
      "types": ["boolean"],
      "format": ["PascalCase"],
      "prefix": ["is", "has"]
    }
  ]
}
```

## Default Configuration

When no options are provided, the rule uses the following defaults:

```json
[
  {
    "selector": "default",
    "format": ["camelCase"],
    "leadingUnderscore": "allow",
    "trailingUnderscore": "allow"
  },
  { "selector": "import", "format": ["camelCase", "PascalCase"] },
  {
    "selector": "variable",
    "format": ["camelCase", "UPPER_CASE"],
    "leadingUnderscore": "allow",
    "trailingUnderscore": "allow"
  },
  { "selector": "typeLike", "format": ["PascalCase"] }
]
```

## Original Documentation

- [typescript-eslint naming-convention](https://typescript-eslint.io/rules/naming-convention)
