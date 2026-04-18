# no-misleading-character-class

## Rule Details

Disallow characters whose visual rendering is made from multiple code points
(combining marks, surrogate pairs, regional indicators, emoji-modifier
sequences, joined ZWJ sequences) from appearing inside a regex character class
`[...]`. Such sequences cannot be matched as a single unit by the regex engine
and therefore produce surprising matches.

Examples of **incorrect** code for this rule:

```javascript
/^[Á]$/u;        // a + combining acute
/^[❇️]$/u;       // base + variation selector
/^[👶🏻]$/u;     // base emoji + skin tone modifier
/^[🇯🇵]$/u;     // two regional indicator symbols
/^[👨‍👩‍👦]$/u; // ZWJ-joined family sequence
/^[👍]$/;        // astral character without `u` / `v` flag
new RegExp("[🎵]");
```

Examples of **correct** code for this rule:

```javascript
/^[abc]$/;
/^[👍]$/u;
/^[\q{👶🏻}]$/v; // v-flag grouping preserves the sequence
new RegExp("^[]$");
/[\ud83d\udc4d]/;
/[\u00B7\u0300-\u036F]/u;
```

### Options

This rule accepts an options object:

```json
{ "no-misleading-character-class": ["error", { "allowEscape": true }] }
```

- `allowEscape` (boolean, default `false`) — when enabled, allows combining
  the troublesome characters inside a character class as long as the
  combining portion is written using a backslash escape sequence. This
  applies to regex literals and to `RegExp(...)` calls whose first argument
  is a string or no-substitution template literal.

Examples of **correct** code with `{ "allowEscape": true }`:

```javascript
/[\ud83d\udc4d]/;                // surrogate pair written with escapes
/[A\u0301]/;                     // combining acute written with escape
new RegExp("[\\uD83D\\uDC4D]");  // surrogate pair in string literal
```

## Differences from ESLint

- The scope of recognition for the `RegExp(...)` constructor is limited to
  calls whose first argument is directly a string literal, template literal
  (no-substitution) or regex literal, or where the callee is `RegExp` /
  `globalThis.RegExp` / `window.RegExp` / `self.RegExp` / `global.RegExp`.
  Patterns constructed through variables, spread arguments, or dynamic
  expressions are not evaluated.
- Suggestions to add the `u` flag use a simplified heuristic to decide
  whether the pattern remains valid under the flag. Patterns with identity
  escapes on letters (e.g. `/[👍]\a/`) are correctly detected as unfixable;
  more exotic cases may rarely produce a suggestion that ESLint would
  suppress.
- In rare edge cases where a character written as `\q{...}` or `\p{...}`
  escape appears outside its valid flag context (e.g. without the v flag),
  the scanner may still treat it as a breaker rather than a plain character.

## Original Documentation

- <https://eslint.org/docs/latest/rules/no-misleading-character-class>
