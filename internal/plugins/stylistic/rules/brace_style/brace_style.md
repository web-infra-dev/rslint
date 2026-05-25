# brace-style

Enforce consistent brace style for blocks.

## Rule Details

This rule enforces a consistent brace style across block-bearing constructs — function bodies, control-flow blocks (`if`, `else`, `for`, `while`, `do`, `try`, `catch`, `finally`), class bodies, `switch` bodies, class static blocks, and TypeScript `namespace` / `module` bodies.

The three supported styles place the opening curly differently relative to the controlling statement, and treat the closing curly's relationship to a following keyword (`else`, `catch`, `finally`) differently.

## Options

This rule has a string option:

- `"1tbs"` (default) — "one true brace style". Opening brace on the same line as the controlling statement; closing brace on the same line as the following keyword (`else`, `catch`, `finally`).
- `"stroustrup"` — like `"1tbs"`, but the closing brace must be on its own line before `else`/`catch`/`finally`.
- `"allman"` — opening brace on a new line by itself; closing brace also on its own line.

This rule has an object option:

- `"allowSingleLine": true` (default `false`) — allows the opening and closing braces for a block to be on the **same** line.

### 1tbs

Examples of **incorrect** code for this rule with the default `"1tbs"` option:

```javascript
function foo()
{
  return true;
}

if (foo)
{
  bar();
}

try
{
  somethingRisky();
} catch(e)
{
  handleError();
}

if (foo) {
  bar();
}
else {
  baz();
}

class C
{
}
```

Examples of **correct** code for this rule with the default `"1tbs"` option:

```javascript
function foo() {
  return true;
}

if (foo) {
  bar();
}

try {
  somethingRisky();
} catch (e) {
  handleError();
}

if (foo) {
  bar();
} else {
  baz();
}

class C {
}
```

### stroustrup

Examples of **incorrect** code for this rule with the `"stroustrup"` option:

```json
{ "@stylistic/brace-style": ["error", "stroustrup"] }
```

```javascript
if (foo) {
  bar();
} else {
  baz();
}

try {
  somethingRisky();
} catch (e) {
  handleError();
}
```

Examples of **correct** code for this rule with the `"stroustrup"` option:

```json
{ "@stylistic/brace-style": ["error", "stroustrup"] }
```

```javascript
function foo() {
  return true;
}

if (foo) {
  bar();
}
else {
  baz();
}

try {
  somethingRisky();
}
catch (e) {
  handleError();
}
```

### allman

Examples of **incorrect** code for this rule with the `"allman"` option:

```json
{ "@stylistic/brace-style": ["error", "allman"] }
```

```javascript
function foo() {
  return true;
}

if (foo) {
  bar();
} else {
  baz();
}
```

Examples of **correct** code for this rule with the `"allman"` option:

```json
{ "@stylistic/brace-style": ["error", "allman"] }
```

```javascript
function foo()
{
  return true;
}

if (foo)
{
  bar();
}
else
{
  baz();
}

try
{
  somethingRisky();
}
catch (e)
{
  handleError();
}
```

### allowSingleLine

Examples of **correct** code for this rule with the `{ "allowSingleLine": true }` option:

```json
{ "@stylistic/brace-style": ["error", "1tbs", { "allowSingleLine": true }] }
```

```javascript
function nop() { return; }

if (foo) { bar(); }

if (foo) { bar(); } else { baz(); }

try { somethingRisky(); } catch (e) { handleError(); }
```

## Original Documentation

- [@stylistic/brace-style](https://eslint.style/rules/brace-style)
