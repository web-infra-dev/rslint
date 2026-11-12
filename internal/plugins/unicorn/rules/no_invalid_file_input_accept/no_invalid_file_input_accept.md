# no-invalid-file-input-accept

## Rule Details

An explicit `accept` attribute on a JSX file input must be a static,
comma-separated list of filename extensions, MIME types, or the wildcards
`audio/*`, `image/*`, and `video/*`.

The rule reports empty entries and invalid values. It can fix casing, spacing,
duplicate entries, missing extension dots, and common MIME type mistakes in
quoted attributes. An expression such as `accept={"IMAGE/PNG"}` also gets a
diagnostic when it needs normalization. Edit the expression to use the value
named in the diagnostic.

Examples of **incorrect** code for this rule

```jsx
<input type="file" accept="image/jpg" />;
<input type="file" accept="png" />;
<input type="file" accept="IMAGE/*" />;
<input type="file" accept="IMAGE/PNG,.PNG, image/png" />;
<input type="file" accept={allowedTypes} />;
```

Examples of **correct** code for this rule

```jsx
<input type="file" accept="image/jpeg" />;
<input type="file" accept=".png" />;
<input type="file" accept="image/*" />;
<input type="file" accept="image/png, .png" />;
```

## Options

This rule has no options.

## Differences from upstream

Rslint checks JSX in JavaScript and TypeScript files. Use ESLint Unicorn with
`html-eslint` to check standalone HTML files.

Autofixes preserve character references where needed. For example,
`accept="&quot;PNG"` becomes `accept=".&quot;png"`.

The rule checks the characters that numeric references represent. It skips an
input with `type="f&#x10069;le"` and reports `.😀png` for
`accept="&#x1F600;PNG"`. Edit the latter manually to preserve the emoji.
Normalizing `accept="&#xD800;PNG"` also requires a manual edit because the
value contains an unpaired surrogate.

## Original Documentation

- [eslint-plugin-unicorn: no-invalid-file-input-accept](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/no-invalid-file-input-accept.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/no-invalid-file-input-accept.js)
