# block-spacing

Disallow or enforce spaces inside of blocks after opening block and before closing block.

## Rule Details

This rule enforces consistent spacing inside an open block token and the next token on the same line. This rule also enforces consistent spacing inside a close block token and previous token on the same line.

When the opening or closing brace and its neighboring token are on different lines, the rule does not fire — multi-line blocks are always allowed regardless of the option.

The rule applies to every `{ ... }` block:

- function / arrow / method / getter / setter / constructor bodies
- `if` / `else` / `for` / `while` / `do-while` / `try` / `catch` / `finally` bodies
- `switch` case blocks
- class `static` blocks

## Options

This rule has a string option:

- `"always"` (default) requires one or more spaces or newlines inside braces.
- `"never"` disallows spaces inside braces.

### always

Examples of **incorrect** code for this rule with the default `"always"` option:

```javascript
function foo() {return true;}
if (foo) { bar = 0;}
function baz() {let i = 0;
    return i;
}
class C {
    static {this.bar = 0;}
}
```

Examples of **correct** code for this rule with the default `"always"` option:

```javascript
function foo() { return true; }
if (foo) { bar = 0; }
class C {
    static { this.bar = 0; }
}
```

### never

Examples of **incorrect** code for this rule with the `"never"` option:

```json
{ "@stylistic/block-spacing": ["error", "never"] }
```

```javascript
function foo() { return true; }
if (foo) { bar = 0;}
class C {
    static { this.bar = 0; }
}
```

Examples of **correct** code for this rule with the `"never"` option:

```json
{ "@stylistic/block-spacing": ["error", "never"] }
```

```javascript
function foo() {return true;}
if (foo) {bar = 0;}
class C {
    static {this.bar = 0;}
}
```

With `"never"`, a `{` immediately followed by a `//` line comment is intentionally exempt — the closing `}` is still checked normally.

## Original Documentation

- [@stylistic/block-spacing](https://eslint.style/rules/block-spacing)
