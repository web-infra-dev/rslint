# dot-notation

## Rule Details

Enforce dot notation whenever possible. Dot notation (`obj.foo`) is generally preferred over bracket notation (`obj["foo"]`) for readability, since bracket notation is really only needed when the property name isn't a valid identifier or is computed dynamically.

The rule provides autofixes to convert bracket notation to dot notation (and, with `allowKeywords: false`, dot notation to bracket notation for reserved words) when it's safe to do so.

Examples of **incorrect** code for this rule:

```javascript
const x = foo["bar"];
```

Examples of **correct** code for this rule:

```javascript
const x = foo.bar;
const y = foo["bar-baz"]; // not a valid identifier
const z = foo[getKey()]; // computed access
```

### Options

This rule accepts a single options object with two properties:

- `allowKeywords` (default `true`) — when `false`, reserved words (`class`, `default`, `new`, …) must use bracket notation instead of dot notation.
- `allowPattern` (default `""`) — a regular expression; bracket-accessed keys matching it are left alone even if they could be written with dot notation.

Examples of **incorrect** code for this rule with `{ "allowKeywords": false }`:

```json
{ "dot-notation": ["error", { "allowKeywords": false }] }
```

```javascript
const x = foo.class;
```

Examples of **correct** code for this rule with `{ "allowKeywords": false }`:

```json
{ "dot-notation": ["error", { "allowKeywords": false }] }
```

```javascript
const x = foo["class"];
```

Examples of **correct** code for this rule with `{ "allowPattern": "^[a-z]+(_[a-z]+)+$" }`:

```json
{ "dot-notation": ["error", { "allowPattern": "^[a-z]+(_[a-z]+)+$" }] }
```

```javascript
const x = foo["snake_case"];
```

## Original Documentation

- [ESLint dot-notation](https://eslint.org/docs/latest/rules/dot-notation)
