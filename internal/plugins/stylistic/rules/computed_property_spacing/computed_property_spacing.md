# computed-property-spacing

Enforce consistent spacing inside computed property brackets.

## Rule Details

This rule covers every place a computed name expression appears inside square brackets:

- Member access `obj[foo]` and optional member access `obj?.[foo]`
- Object literal keys `{ [foo]: 1 }`
- Object destructuring patterns `const { [foo]: x } = obj`
- Class member names (methods, accessors, fields, `accessor` properties)
- TypeScript indexed access types `type T = A[B]`

Multi-line forms are exempt — the rule only fires when the opening `[` and its first inner token (or the closing `]` and its last inner token) sit on the same source line.

## Options

This rule has a string option:

- `"never"` (default) disallows spaces inside computed property brackets.
- `"always"` requires one space inside computed property brackets.

This rule has an object option:

- `"enforceForClassMembers": true` (default) — also enforce the spacing rule on class member names. Set to `false` to leave class members unchecked.

### never

Examples of **incorrect** code for this rule with the default `"never"` option:

```javascript
obj[foo ];
obj[ 'foo'];
const x = { [ b ]: a };
const { [ a ]: someProp } = obj;
class A { [ a ]() {} }
type Foo = A[ B ];
```

Examples of **correct** code for this rule with the default `"never"` option:

```javascript
obj[foo];
obj['foo'];
obj?.[foo];
const x = { [b]: a };
const { [a]: someProp } = obj;
class A { [a]() {} }
type Foo = A[B];
```

### always

Examples of **incorrect** code for this rule with the `"always"` option:

```json
{ "@stylistic/computed-property-spacing": ["error", "always"] }
```

```javascript
obj[foo];
const x = { [b]: a };
const { [a]: someProp } = obj;
class A { [a]() {} }
type Foo = A[B];
```

Examples of **correct** code for this rule with the `"always"` option:

```json
{ "@stylistic/computed-property-spacing": ["error", "always"] }
```

```javascript
obj[ foo ];
obj?.[ foo ];
const x = { [ b ]: a };
const { [ a ]: someProp } = obj;
class A { [ a ]() {} }
type Foo = A[ B ];
```

### enforceForClassMembers

When set to `false`, class member names are exempt regardless of the `"never"` / `"always"` mode.

Examples of **correct** code for this rule with `"never", { "enforceForClassMembers": false }`:

```json
{ "@stylistic/computed-property-spacing": ["error", "never", { "enforceForClassMembers": false }] }
```

```javascript
class A {
  [ a ]() {}
  get [ b ]() {}
  static [ c ]() {}
  [ d ] = 1;
  accessor [ e ] = 1;
}
```

## Original Documentation

- [@stylistic/computed-property-spacing](https://eslint.style/rules/computed-property-spacing)
