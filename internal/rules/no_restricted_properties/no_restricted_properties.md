# no-restricted-properties

## Rule Details

Certain properties on objects may be disallowed in a codebase. This is useful for deprecating an API or restricting usage of a module's methods — for example, disallowing `describe.only` when using Mocha, or steering people toward `Object.assign` instead of `_.extend`.

This rule looks for accessing a given property key on a given object name, either when reading the property's value or invoking it as a function. It applies to both dot/bracket property access and destructuring.

The rule takes a list of objects, where the object name and property name are specified:

```json
{ "no-restricted-properties": ["error", { "object": "disallowedObjectName", "property": "disallowedPropertyName" }] }
```

Multiple object/property pairs can be disallowed, and each can specify an optional custom message:

```json
{
  "no-restricted-properties": [
    "error",
    { "object": "disallowedObjectName", "property": "disallowedPropertyName" },
    {
      "object": "disallowedObjectName",
      "property": "anotherDisallowedPropertyName",
      "message": "Please use allowedObjectName.allowedPropertyName."
    }
  ]
}
```

Examples of **incorrect** code for this rule:

```javascript
const example = disallowedObjectName.disallowedPropertyName;

disallowedObjectName.disallowedPropertyName();
```

Examples of **correct** code for this rule:

```javascript
const example = disallowedObjectName.somePropertyName;

allowedObjectName.disallowedPropertyName();
```

If the object name is omitted, the property is disallowed on every object:

```json
{ "no-restricted-properties": ["error", { "property": "__defineGetter__", "message": "Please use Object.defineProperty instead." }] }
```

If the property name is omitted, every property access on the given object is disallowed:

```json
{ "no-restricted-properties": ["error", { "object": "require", "message": "Please call require() directly." }] }
```

Examples of **incorrect** code for this rule with `{ "object": "require" }`:

```javascript
require.resolve("foo");
```

Examples of **correct** code for this rule with `{ "object": "require" }`:

```javascript
require("foo");
```

To restrict a property globally but allow specific objects to use it, add `allowObjects` (mutually exclusive with `object`):

```json
{ "no-restricted-properties": ["error", { "property": "push", "allowObjects": ["router", "history"], "message": "Prefer [...array, newValue]." }] }
```

Examples of **incorrect** code for this rule with `{ "property": "push", "allowObjects": ["router", "history"] }`:

```javascript
myArray.push(5);
```

Examples of **correct** code for this rule with `{ "property": "push", "allowObjects": ["router", "history"] }`:

```javascript
router.push("/home");
history.push("/about");
```

To restrict every property on an object except a chosen few, add `allowProperties` (mutually exclusive with `property`):

```json
{ "no-restricted-properties": ["error", { "object": "config", "allowProperties": ["settings", "version"] }] }
```

Examples of **incorrect** code for this rule with `{ "object": "config", "allowProperties": ["settings", "version"] }`:

```javascript
config.apiKey = "12345";
```

Examples of **correct** code for this rule with `{ "object": "config", "allowProperties": ["settings", "version"] }`:

```javascript
config.settings = { theme: "dark" };
config.version = "1.0.0";
```

## When Not To Use It

If you don't have any object/property combinations to restrict, you should not use this rule.

## Original Documentation

- [ESLint: no-restricted-properties](https://eslint.org/docs/latest/rules/no-restricted-properties)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/no-restricted-properties.js)
