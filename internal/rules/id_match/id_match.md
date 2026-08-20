# id-match

## Rule Details

This rule requires every identifier in the source to match a regular expression, so a project can enforce one naming convention — `camelCase`, `snake_case`, a required prefix — across variables, functions, classes, parameters and, optionally, properties and class fields.

The pattern is the rule's first option and is read as a JavaScript regular expression with the `u` flag. It is matched anywhere in the name unless it is anchored, so `^[a-z]+$` requires the whole name to be lowercase while `[a-z]+` only requires the name to contain a lowercase run.

TypeScript names count as identifiers too: type aliases, interfaces, type parameters, enums and their members, namespaces, and the names a type annotation refers to are all held to the pattern.

Names the file does not declare are left alone, because the author has no say in what they are called: a reference to a global such as `Object` or `Array`, a type the TypeScript standard library declares such as `Record` or `Partial`, the key of an import attribute (`with { type: "json" }`), and `import.meta` / `new.target`. A type the file declares or imports is the author's, and is checked.

Examples of **incorrect** code for this rule:

```json
{ "id-match": ["error", "^[a-z]+$"] }
```

```javascript
var first_name = 'Ada';

function no_under() {}

class My_Class {}
```

Examples of **correct** code for this rule:

```json
{ "id-match": ["error", "^[a-z]+$"] }
```

```javascript
var name = 'Ada';

function greet() {}

no_under();

var myDate = new Date();
```

## Options

This rule has a string as its first option (the pattern, default `"^.+$"`) and an object as its second.

### properties

`properties` (default `false`) also checks property names: the keys of object literals, and the property of a member access when that member access is the assigned-to side of an assignment.

Examples of **incorrect** code with `{ "properties": true }`:

```json
{ "id-match": ["error", "^[^_]+$", { "properties": true }] }
```

```javascript
var obj = { no_under: 1 };

obj.no_under = 2;
```

Examples of **correct** code with `{ "properties": true }`:

```json
{ "id-match": ["error", "^[^_]+$", { "properties": true }] }
```

```javascript
var obj = { valid: 1 };

var value = other.no_under;

if (other.no_under) {
}
```

### classFields

`classFields` (default `false`) also checks class field names, including private ones. A class method is checked whatever this option says.

Examples of **incorrect** code with `{ "classFields": true }`:

```json
{ "id-match": ["error", "^[^_]+$", { "classFields": true }] }
```

```javascript
class Foo {
  _bar = 1;
  #_baz = 2;
}
```

Examples of **correct** code with `{ "classFields": true }`:

```json
{ "id-match": ["error", "^[^_]+$", { "classFields": true }] }
```

```javascript
class Foo {
  bar = 1;
  #baz = 2;
}
```

### onlyDeclarations

`onlyDeclarations` (default `false`) narrows the check to variable declarations, function declarations, and the parameters of a function declaration. A name that is only read is left alone.

Examples of **incorrect** code with `{ "onlyDeclarations": true }`:

```json
{ "id-match": ["error", "^[a-z]+$", { "onlyDeclarations": true }] }
```

```javascript
var __foo = 'Ada';
```

Examples of **correct** code with `{ "onlyDeclarations": true }`:

```json
{ "id-match": ["error", "^[a-z]+$", { "onlyDeclarations": true }] }
```

```javascript
__foo = 'Ada';
```

### ignoreDestructuring

`ignoreDestructuring` (default `false`) leaves a destructured name alone when it repeats a property of the source object. A name the destructuring introduces — a rename, or a rest element — is still checked.

Examples of **incorrect** code with `{ "ignoreDestructuring": true }`:

```json
{ "id-match": ["error", "^[^_]+$", { "ignoreDestructuring": true }] }
```

```javascript
var { category_id: category_alias } = query;

var { categoryId, ...other_props } = query;
```

Examples of **correct** code with `{ "ignoreDestructuring": true }`:

```json
{ "id-match": ["error", "^[^_]+$", { "ignoreDestructuring": true }] }
```

```javascript
var { category_id } = query;

var { category_id = 1 } = query;
```

## Differences from ESLint

- A parameter carrying a type annotation is reported over the parameter name alone. ESLint with a TypeScript parser reports over the name and its annotation together, so on `function f(a_1: string) {}` the report ends at column 15 here and at column 26 there.

## Original Documentation

- [ESLint: id-match](https://eslint.org/docs/latest/rules/id-match)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/id-match.js)
