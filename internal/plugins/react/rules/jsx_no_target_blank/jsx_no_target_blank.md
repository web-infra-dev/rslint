# jsx-no-target-blank

Disallow `target="_blank"` attribute without `rel="noreferrer"`.

## Rule Details

When creating a JSX element that has an `a` tag, it is often desired to have
the link open in a new tab using the `target="_blank"` attribute. Using this
attribute unaccompanied by `rel="noreferrer"`, however, is a severe security
vulnerability (see
[here](https://mathiasbynens.github.io/rel-noopener) for more details).

Examples of **incorrect** code for this rule:

```jsx
var Hello = <a target="_blank" href="http://example.com/">Foo</a>;
```

Examples of **correct** code for this rule:

```jsx
var Hello = <p target="_blank"></p>;
var Hello = <a target="_blank" rel="noreferrer" href="http://example.com"></a>;
var Hello = <a target="_blank" rel="noopener noreferrer" href="http://example.com"></a>;
var Hello = <a target="_blank" href="path/in/the/host"></a>;
```

## Rule Options

```json
{
  "react/jsx-no-target-blank": [
    "error",
    {
      "allowReferrer": false,
      "enforceDynamicLinks": "always",
      "warnOnSpreadAttributes": false,
      "links": true,
      "forms": false
    }
  ]
}
```

### `allowReferrer`

When `true` the rule permits `rel="noopener"` alone (the `Referer` header is
still sent). Defaults to `false`.

### `enforceDynamicLinks`

- `"always"` (default) — the rule also checks dynamic link targets
  (`href={value}`).
- `"never"` — dynamic link targets are exempt.

### `warnOnSpreadAttributes`

When `true`, spread attributes (`{...props}`) are treated as a potential
override of `target` / `href` / `rel`, so the rule may still report even when
the explicit attributes look safe. Defaults to `false`.

### `links` / `forms`

Toggle the checks for anchor-like link components (`links`, default `true`)
and `<form>`-like components (`forms`, default `false`).

### Custom components

The rule honors the top-level `settings.linkComponents` and
`settings.formComponents` entries, matching
[eslint-plugin-react's `linkComponents` configuration](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/README.md#configuration).
Example:

```json
{
  "settings": {
    "linkComponents": [
      "Hyperlink",
      { "name": "Link", "linkAttribute": "to" }
    ]
  }
}
```

## Differences from ESLint

None — this rule aims for 1:1 behavior with
[`react/jsx-no-target-blank`](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/jsx-no-target-blank.md).
A few upstream implementation details deserve a note:

- The form branch (`forms: true`) does not autofix, matching upstream.
- When the `rel` attribute uses an expression the rule cannot analyze
  (e.g. `rel={getRel()}`), the diagnostic is still reported but no fix is
  emitted.
- The secure-rel check for `<form>` always treats `allowReferrer` as `false`
  — forms with only `rel="noopener"` still report even when the rule-level
  `allowReferrer` option is on, matching upstream's
  `hasSecureRel(node)` call shape.

## Original Documentation

- [eslint-plugin-react docs](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/jsx-no-target-blank.md)
