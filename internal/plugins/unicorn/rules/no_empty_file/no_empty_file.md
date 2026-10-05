# no-empty-file

## Rule Details

Disallow empty files, including files containing only whitespace, comments,
directive prologues, empty statements, empty blocks, or a hashbang.

Examples of **incorrect** code for this rule:

```javascript
// Nothing implemented yet.
```

```javascript
'use strict';
{}
```

Examples of **correct** code for this rule:

```javascript
export {};
```

```typescript
/// <reference types="node" />
```

Triple-slash comments are allowed, including TypeScript reference directives.
Parenthesized strings and strings inside ordinary blocks count as content.

## Options

### allowComments

Type: `boolean`. Default: `false`.

Set `allowComments: true` to allow files containing only regular comments.
Disable and enable directives do not count as regular comments. A hashbang,
directive prologue, empty statement, or empty block is still reported even if
accompanied by comments.

```javascript
{
  rules: {
    'unicorn/no-empty-file': ['error', { allowComments: true }],
  },
}
```

This rule has no automatic fix or suggestions.

## Differences from upstream

The `rslint-disable` and `rslint-enable` aliases also count as lint directives.
For example, a file containing only `/* rslint-enable */` is reported even with
`allowComments: true`; upstream treats that comment as ordinary text and allows it.

## Original Documentation

- [eslint-plugin-unicorn: no-empty-file](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/docs/rules/no-empty-file.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/rules/no-empty-file.js)
