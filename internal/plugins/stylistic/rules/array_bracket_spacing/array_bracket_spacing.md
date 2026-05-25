# array-bracket-spacing

Enforce consistent spacing inside array brackets.

## Rule Details

This rule applies to both array literals (`[1, 2]`) and array destructuring patterns (`var [a, b] = arr`).

Multi-line arrays are always allowed regardless of the option — the rule only fires when the opening and closing bracket sit on the same line as their adjacent token.

## Options

This rule has a string option:

- `"never"` (default) disallows spaces inside array brackets.
- `"always"` requires one or more spaces or newlines inside array brackets.

This rule has an object option with three boolean exception keys. Each exception flips the spacing requirement for the matching shape — setting it to the opposite of the primary mode (e.g. `"singleValue": true` with `"never"`, or `"singleValue": false` with `"always"`) turns the exception on.

Empty array literals (`[]`) do not require spaces under the `"always"` option.

### never

Examples of **incorrect** code for this rule with the default `"never"` option:

```javascript
var arr = [ 'foo', 'bar' ];
var arr = ['foo', 'bar' ];
var arr = [ ['foo'], 'bar'];
var [ x, y ] = z;
var arr = [ ];
```

Examples of **correct** code for this rule with the default `"never"` option:

```javascript
var arr = [];
var arr = ['foo', 'bar', 'baz'];
var arr = [['foo'], 'bar', 'baz'];
var [x, y] = z;
var [x, ...y] = z;
var arr = [
  'foo',
  'bar',
];
```

### always

Examples of **incorrect** code for this rule with the `"always"` option:

```json
{ "@stylistic/array-bracket-spacing": ["error", "always"] }
```

```javascript
var arr = ['foo', 'bar'];
var arr = ['foo', 'bar' ];
var arr = [ ['foo'], 'bar'];
var [x, y] = z;
```

Examples of **correct** code for this rule with the `"always"` option:

```json
{ "@stylistic/array-bracket-spacing": ["error", "always"] }
```

```javascript
var arr = [];
var arr = [ 'foo', 'bar', 'baz' ];
var arr = [ [ 'foo' ], 'bar', 'baz' ];
var [ x, y ] = z;
var [ x, ...y ] = z;
```

### singleValue

Examples of **correct** code for this rule with `"never", { "singleValue": true }`:

```json
{ "@stylistic/array-bracket-spacing": ["error", "never", { "singleValue": true }] }
```

```javascript
var foo = [ 'bar' ];
var foo = [ 2 ];
var foo = [ {'foo': 'bar'} ];
```

Examples of **correct** code for this rule with `"always", { "singleValue": false }`:

```json
{ "@stylistic/array-bracket-spacing": ["error", "always", { "singleValue": false }] }
```

```javascript
var foo = ['bar'];
var foo = [2];
var foo = [{ 'foo': 'bar' }];
```

### objectsInArrays

Examples of **correct** code for this rule with `"never", { "objectsInArrays": true }`:

```json
{ "@stylistic/array-bracket-spacing": ["error", "never", { "objectsInArrays": true }] }
```

```javascript
var arr = [ {'foo': 'bar'} ];
var arr = [ {'foo': 'bar'}, 1, 5];
var arr = [1, 5, {'foo': 'bar'} ];
```

Examples of **correct** code for this rule with `"always", { "objectsInArrays": false }`:

```json
{ "@stylistic/array-bracket-spacing": ["error", "always", { "objectsInArrays": false }] }
```

```javascript
var arr = [{ 'foo': 'bar' }];
var arr = [{ 'foo': 'bar' }, 1, 5 ];
var arr = [ 1, 5, { 'foo': 'bar' }];
```

### arraysInArrays

Examples of **correct** code for this rule with `"never", { "arraysInArrays": true }`:

```json
{ "@stylistic/array-bracket-spacing": ["error", "never", { "arraysInArrays": true }] }
```

```javascript
var arr = [ [1, 2] ];
var arr = [ [1, 2], 2, 3, 4];
var arr = [1, 2, 3, [4] ];
```

Examples of **correct** code for this rule with `"always", { "arraysInArrays": false }`:

```json
{ "@stylistic/array-bracket-spacing": ["error", "always", { "arraysInArrays": false }] }
```

```javascript
var arr = [[ 1, 2 ]];
var arr = [[ 1, 2 ], 2, 3, 4 ];
var arr = [ 1, 2, 3, [ 4 ]];
```

## Original Documentation

- [@stylistic/array-bracket-spacing](https://eslint.style/rules/array-bracket-spacing)
