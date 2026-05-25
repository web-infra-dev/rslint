# dot-location

Enforce consistent newline placement before or after dots in member expressions.

## Rule Details

JavaScript permits the dot in a member expression to sit on either the line of the object or the line of the property. Inconsistent placement reduces readability.

This rule applies to property access (`a.b`, `a?.b`), TypeScript type qualifiers (`A.B`), import-type qualifiers (`import('m').A`), `import.meta` / `new.target`, and JSX member tag names (`<Form.Input />`). Computed member access (`obj['prop']`) and indexed type access (`Foo[K]`) are not checked.

Examples of **incorrect** code for this rule with the default `"object"` option:

```javascript
var foo = object
.property;

type Foo = Obj
  .Prop;

type Bar = import('Obj')
  .Prop;
```

Examples of **correct** code for this rule with the default `"object"` option:

```javascript
var foo = object.
property;

var baz = object.property;

type Foo = Obj.
  Prop;

type Bar = import('Obj').
  Prop;
```

## Options

This rule has a string option:

- `"object"` (default) requires the dot to be on the same line as the object portion.
- `"property"` requires the dot to be on the same line as the property portion.

### object

Examples of **incorrect** code for this rule with the `"object"` option:

```json
{ "@stylistic/dot-location": ["error", "object"] }
```

```javascript
var foo = object
.property;

<Form
  .Input />
```

Examples of **correct** code for this rule with the `"object"` option:

```json
{ "@stylistic/dot-location": ["error", "object"] }
```

```javascript
var foo = object.
property;

<Form.
  Input />
```

### property

Examples of **incorrect** code for this rule with the `"property"` option:

```json
{ "@stylistic/dot-location": ["error", "property"] }
```

```javascript
var foo = object.
property;

<Form.
  Input />
```

Examples of **correct** code for this rule with the `"property"` option:

```json
{ "@stylistic/dot-location": ["error", "property"] }
```

```javascript
var foo = object
.property;

<Form
  .Input />
```

## Original Documentation

- [@stylistic/dot-location](https://eslint.style/rules/dot-location)
