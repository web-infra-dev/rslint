# no-irregular-whitespace

## Rule Details

Disallows irregular whitespace characters outside of strings, comments, regular expressions, and template literals. Irregular whitespace characters can cause issues with various parsers and can be difficult to debug.

The following characters are considered irregular whitespace:

- `\u000B` - Line Tabulation
- `\u000C` - Form Feed
- `\u0085` - Next Line
- `\u00A0` - No-Break Space
- `\u1680` - Ogham Space Mark
- `\u180E` - Mongolian Vowel Separator
- `\u2000` - En Quad through `\u200B` - Zero Width Space
- `\u202F` - Narrow No-Break Space
- `\u205F` - Medium Mathematical Space
- `\u3000` - Ideographic Space
- `\uFEFF` - Zero Width No-Break Space (BOM)
- `\u2028` - Line Separator
- `\u2029` - Paragraph Separator

Examples of **incorrect** code for this rule:

```javascript
var any = 'thing';
```

Examples of **correct** code for this rule:

```javascript
var any = 'thing';
```

Examples of **correct** code for this rule with `{ "skipStrings": true }` (default):

```json
{ "no-irregular-whitespace": ["error", { "skipStrings": true }] }
```

```javascript
var foo = ' ';
```

Examples of **correct** code for this rule with `{ "skipComments": true }`:

```json
{ "no-irregular-whitespace": ["error", { "skipComments": true }] }
```

```javascript
// Comment with irregular whitespace
/* Block comment with irregular whitespace */
```

Examples of **correct** code for this rule with `{ "skipTemplates": true }`:

```json
{ "no-irregular-whitespace": ["error", { "skipTemplates": true }] }
```

```javascript
var foo = ` `;
```

## Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `skipStrings` | `boolean` | `true` | Allow irregular whitespace in string literals |
| `skipComments` | `boolean` | `false` | Allow irregular whitespace in comments |
| `skipRegExps` | `boolean` | `false` | Allow irregular whitespace in regular expressions |
| `skipTemplates` | `boolean` | `false` | Allow irregular whitespace in template literals |
| `skipJSXText` | `boolean` | `false` | Allow irregular whitespace in JSX text |

## Original Documentation

[ESLint - no-irregular-whitespace](https://eslint.org/docs/latest/rules/no-irregular-whitespace)
