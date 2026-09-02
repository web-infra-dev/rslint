# prefer-blob-reading-methods

## Rule Details

Prefer the promise-based `Blob#arrayBuffer()` and `Blob#text()` methods over
the corresponding callback-based `FileReader` methods.

Examples of **incorrect** code for this rule:

```javascript
fileReader.readAsArrayBuffer(blob);
fileReader.readAsText(blob);
```

Examples of **correct** code for this rule:

```javascript
const arrayBuffer = await blob.arrayBuffer();
const text = await blob.text();
```

Calls that pass an encoding to `FileReader#readAsText()` are allowed because
`Blob#text()` always decodes as UTF-8. Other `FileReader` methods are also left
unchanged:

```javascript
fileReader.readAsText(blob, "ascii");
fileReader.readAsDataURL(blob);
```

## Original Documentation

- [eslint-plugin-unicorn: prefer-blob-reading-methods](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/docs/rules/prefer-blob-reading-methods.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/rules/prefer-blob-reading-methods.js)
