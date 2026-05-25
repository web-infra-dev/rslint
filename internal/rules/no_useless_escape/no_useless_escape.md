# no-useless-escape

Disallow unnecessary escape characters.

Escaping non-special characters in strings, template literals, and regular expressions does not change behavior. Removing the redundant `\` keeps the code simpler and avoids confusion.

```js
let foo = "hol\a"; // > foo = "hola"
let bar = `${foo}\!`; // > bar = "hola!"
let baz = /\:/; // same as /:/
```

## Rule Details

This rule flags escapes that can be safely removed without changing behavior.

Examples of **incorrect** code for this rule:

```javascript
"\'";
'\"';
"\#";
"\e";
`\"`;
`\"${foo}\"`;
`\#{foo}`;
/\!/;
/\@/;
/[\[]/;
/[a-z\-]/;
```

Examples of **correct** code for this rule:

```javascript
"\"";
'\'';
"\x12";
"©";
"\371";
"xsℑ";
`\``;
`\${${foo}}`;
`$\{${foo}}`;
/\\/g;
/\t/g;
/\w\$\*\^\./;
/[[]/;
/[\]]/;
/[a-z-]/;
```

## Options

This rule has an object option:

- `allowRegexCharacters` — array of characters whose `\X` form is always allowed inside regular expressions, even when the `\` would otherwise be flagged. Useful for characters like `-` where the explicit escape can prevent the pattern from drifting into a range as the class grows.

### allowRegexCharacters

Examples of **incorrect** code for the `{ "allowRegexCharacters": ["-"] }` option:

```json
{ "no-useless-escape": ["error", { "allowRegexCharacters": ["-"] }] }
```

```javascript
/\!/;
/\@/;
/[a-z\^]/;
```

Examples of **correct** code for the `{ "allowRegexCharacters": ["-"] }` option:

```json
{ "no-useless-escape": ["error", { "allowRegexCharacters": ["-"] }] }
```

```javascript
/[0\-]/;
/[\-9]/;
/a\-b/;
```

## When Not To Use It

If you do not want to be notified about unnecessary escapes, you can safely disable this rule.

## Original Documentation

- [no-useless-escape](https://eslint.org/docs/latest/rules/no-useless-escape)
