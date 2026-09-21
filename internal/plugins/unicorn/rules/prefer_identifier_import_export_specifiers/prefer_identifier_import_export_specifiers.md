# prefer-identifier-import-export-specifiers

## Rule Details

Prefer identifier names over quoted strings in import/export specifiers and import attributes when the decoded string is a valid JavaScript identifier name.

Reserved words are valid module names and attribute keys. Strings containing spaces, punctuation, or invalid identifier characters remain quoted. Fixes preserve the spacing needed around neighboring keywords.

Examples of **incorrect** code:

```javascript
import {"name" as local} from "module";
```

Examples of **correct** code:

```javascript
import {name as local} from "module";
```

## Options

This rule has no options. It is enabled at `error` severity in `unicornPlugin.configs.recommended`.

## Original Documentation

- [eslint-plugin-unicorn: prefer-identifier-import-export-specifiers](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/docs/rules/prefer-identifier-import-export-specifiers.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/rules/prefer-identifier-import-export-specifiers.js)
